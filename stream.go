package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	maxRetries = 3
	// streamTimeout caps one whole streaming call, retries included. The
	// client's ResponseHeaderTimeout only guards the wait for the *first* byte;
	// this guards against an upstream that stalls midway through the body.
	streamTimeout = 5 * time.Minute
)

// llmClient is a shared HTTP client configured for streaming LLM APIs. No
// overall Timeout is set because a streaming response is unbounded in
// duration; ResponseHeaderTimeout guards against a stalled upstream that
// never sends the first byte.
var llmClient = &http.Client{
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	},
}

// Message is one turn in a conversation sent to the LLM API. Both provider
// wire formats use the same role/content shape.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// streamLLM calls the user's currently selected model (see currentModel in
// settings.go) and forwards SSE text chunks to the browser. It speaks two
// wire formats: the Anthropic Messages API and OpenAI-compatible chat
// completions (DeepSeek) — see models.go for which provider uses which.
//
// ctx should be the request context (r.Context()) so that a client disconnect
// cancels the upstream call instead of streaming on and burning tokens.
//
// onComplete(fullText) is optional and used to populate the lesson cache. It
// only fires when the stream finishes cleanly (context not cancelled, no read
// error, upstream sent its terminal marker, reply not truncated or refused),
// so a disconnect mid-stream never caches a partial lesson.
func streamLLM(ctx context.Context, w http.ResponseWriter, user, system, prompt string, messages []Message, onComplete ...func(string)) {
	model := currentModel(user)
	prov := providers[model.Provider]
	apiKey := os.Getenv(prov.KeyEnv)

	// Overall deadline so a stalled upstream can never hang a request forever.
	// Cancelling also releases the upstream connection when we return early.
	ctx, cancel := context.WithTimeout(ctx, streamTimeout)
	defer cancel()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Fprintf(w, "data: {\"error\":\"streaming not supported\"}\n\n")
		return
	}
	sendErr := func(msg string) {
		// Log server-side too: otherwise failures like a missing API key are
		// only visible as an SSE event in the browser and leave the console
		// silent, which makes them hard to diagnose.
		log.Printf("stream error (%s): %s", model.ID, msg)
		data, _ := json.Marshal(map[string]string{"error": msg})
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	if apiKey == "" {
		sendErr(prov.KeyEnv + " is not set. Export it and restart the server.")
		return
	}

	msgs := messages
	if len(msgs) == 0 {
		msgs = []Message{{Role: "user", Content: prompt}}
	}

	// Build the provider-specific request. The two APIs want the same
	// information in slightly different shapes.
	headers := map[string]string{"Content-Type": "application/json"}
	var reqBody []byte
	switch prov.Style {
	case styleAnthropic:
		body := map[string]interface{}{
			"model":      model.ID,
			"max_tokens": prov.MaxTokens,
			"system":     system,
			"messages":   msgs,
			"stream":     true,
		}
		// Fable's safety classifiers can decline a request outright
		// (stop_reason "refusal") — even for harmless lesson content, false
		// positives happen. Opting into a server-side fallback lets Opus
		// answer within the same call instead of failing the stream.
		if model.ID == "claude-fable-5" {
			body["fallbacks"] = []map[string]string{{"model": "claude-opus-4-8"}}
			headers["anthropic-beta"] = "server-side-fallback-2026-06-01"
		}
		reqBody, _ = json.Marshal(body)
		headers["x-api-key"] = apiKey
		headers["anthropic-version"] = "2023-06-01"
	case styleOpenAI:
		// OpenAI-compatible APIs put the system prompt in the message list
		// rather than a separate field.
		oaMsgs := make([]Message, 0, len(msgs)+1)
		if system != "" {
			oaMsgs = append(oaMsgs, Message{Role: "system", Content: system})
		}
		oaMsgs = append(oaMsgs, msgs...)
		reqBody, _ = json.Marshal(map[string]interface{}{
			"model":      model.ID,
			"max_tokens": prov.MaxTokens,
			"messages":   oaMsgs,
			"stream":     true,
		})
		headers["Authorization"] = "Bearer " + apiKey
	}

	// Retry on rate-limit (429), overload (529), and transient server errors
	// (5xx) with exponential back-off. The request body must be rebuilt each
	// attempt because http.Request.Body is consumed on the first read.
	var (
		resp       *http.Response
		lastStatus int
	)
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(1<<uint(attempt-1)) * time.Second // 1 s, 2 s, 4 s
			log.Printf("%s API retry %d/%d after %v (last status %d)", prov.ID, attempt, maxRetries, delay, lastStatus)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", prov.URL, bytes.NewBuffer(reqBody))
		if err != nil {
			sendErr("Could not build request: " + err.Error())
			return
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err = llmClient.Do(req)
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return // client disconnected; nothing to report
			}
			sendErr("Could not reach " + prov.ID + " API: " + err.Error())
			return
		}

		if resp.StatusCode == http.StatusOK {
			break
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		lastStatus = resp.StatusCode
		log.Printf("%s API attempt %d/%d: status %d: %s", prov.ID, attempt+1, maxRetries+1, lastStatus, string(body))

		if lastStatus == 429 || lastStatus == 529 || lastStatus >= 500 {
			resp = nil
			continue // retryable
		}

		// Non-retryable (e.g. 401 bad key, 400 bad request).
		sendErr(fmt.Sprintf("API returned status %d — check your API key", lastStatus))
		return
	}
	if resp == nil {
		sendErr(fmt.Sprintf("API returned status %d after %d attempts — try again shortly", lastStatus, maxRetries+1))
		return
	}
	defer resp.Body.Close()

	var (
		buf        strings.Builder // accumulate for the optional onComplete callback
		sawEnd     bool            // upstream sent its terminal marker — the response is whole
		stopReason string          // "end_turn", "max_tokens", "refusal" / "stop", "length"
	)
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // tolerate long SSE lines
	forward := func(text string) {
		buf.WriteString(text)
		chunk, _ := json.Marshal(map[string]string{"text": text})
		fmt.Fprintf(w, "data: %s\n\n", chunk)
		flusher.Flush()
	}
stream:
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			// OpenAI-style terminal marker (Anthropic's is the message_stop
			// event below).
			sawEnd = true
			break stream
		}
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		if prov.Style == styleOpenAI {
			// Chunk shape: {"choices":[{"delta":{"content":"…"},"finish_reason":…}]}
			// A mid-stream failure arrives as {"error":{…}} instead.
			if e, ok := event["error"].(map[string]interface{}); ok {
				msg, _ := e["message"].(string)
				if msg == "" {
					msg = "unknown upstream error"
				}
				sendErr(prov.ID + " API error: " + msg)
				return
			}
			choices, _ := event["choices"].([]interface{})
			if len(choices) == 0 {
				continue
			}
			choice, _ := choices[0].(map[string]interface{})
			if fr, ok := choice["finish_reason"].(string); ok && fr != "" {
				stopReason = fr // "stop", or "length" when cut off by max_tokens
			}
			// deepseek-reasoner streams its chain of thought as a separate
			// "reasoning_content" delta first; only the answer ("content") is
			// forwarded to the browser.
			delta, _ := choice["delta"].(map[string]interface{})
			if text, _ := delta["content"].(string); text != "" {
				forward(text)
			}
			continue
		}

		switch event["type"] {
		case "error":
			// The API can also fail mid-stream (e.g. overloaded_error) as an
			// SSE event rather than an HTTP status. Surface it to the browser
			// instead of letting the text silently stop.
			msg := "unknown upstream error"
			if e, ok := event["error"].(map[string]interface{}); ok {
				if m, ok := e["message"].(string); ok && m != "" {
					msg = m
				}
			}
			sendErr("Anthropic API error: " + msg)
			return
		case "message_delta":
			// Carries the stop_reason. "max_tokens" means the reply was cut
			// off by the cap, so it must not be cached as a complete lesson.
			if delta, ok := event["delta"].(map[string]interface{}); ok {
				if sr, ok := delta["stop_reason"].(string); ok && sr != "" {
					stopReason = sr
				}
			}
		case "message_stop":
			// Anthropic's terminal event — the stream is done.
			sawEnd = true
			break stream
		case "content_block_delta":
			delta, _ := event["delta"].(map[string]interface{})
			if delta == nil || delta["type"] != "text_delta" {
				continue
			}
			text, _ := delta["text"].(string)
			forward(text)
		}
	}

	// The stream can end because the context died: a plain cancel means the
	// client disconnected (nobody is listening — write nothing), while a
	// deadline means streamTimeout fired on a stalled upstream. Either way,
	// never cache what could be a partial response.
	if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			sendErr("Stream timed out — try again.")
		}
		return
	}

	if err := scanner.Err(); err != nil {
		log.Printf("stream read error: %v", err)
		sendErr("Stream interrupted — try again.")
		return
	}

	// A premature EOF is not a scanner error, so reaching here without the
	// terminal marker means the upstream dropped the connection mid-response.
	// Report it rather than passing truncated text off as a finished lesson.
	if !sawEnd {
		sendErr("Stream ended unexpectedly — try again.")
		return
	}

	// The model (or its safety layer) declined to answer. Whatever partial
	// text arrived must not be cached or passed off as a finished response.
	if stopReason == "refusal" {
		sendErr("The model declined this request — try regenerating, or switch models.")
		return
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()

	// Only a verifiably complete response goes in the cache: the terminal
	// marker was seen (checked above) and the model wasn't cut off by the
	// max_tokens cap ("length" is the OpenAI-style spelling of the same).
	if stopReason == "max_tokens" || stopReason == "length" {
		log.Printf("stream hit the %d-token cap — truncated response not cached", prov.MaxTokens)
		return
	}
	if len(onComplete) > 0 && onComplete[0] != nil && buf.Len() > 0 {
		onComplete[0](buf.String())
	}
}
