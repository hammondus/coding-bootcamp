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
	anthropicAPI = "https://api.anthropic.com/v1/messages"
	// model        = "claude-sonnet-4-6"
	model      = "claude-opus-4-8"
	maxRetries = 3
	// maxTokens needs headroom: a full lesson (overview, 5–7 concepts, two code
	// examples, pitfalls) regularly runs past small caps, and a capped response
	// is a silently truncated lesson.
	maxTokens = 8192
	// streamTimeout caps one whole streaming call, retries included. The
	// client's ResponseHeaderTimeout only guards the wait for the *first* byte;
	// this guards against an upstream that stalls midway through the body.
	streamTimeout = 5 * time.Minute
)

// anthropicClient is a shared HTTP client configured for the Anthropic
// streaming API. No overall Timeout is set because a streaming response is
// unbounded in duration; ResponseHeaderTimeout guards against a stalled
// upstream that never sends the first byte.
var anthropicClient = &http.Client{
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	},
}

// Message is one turn in a conversation sent to the Anthropic API.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// streamFromAnthropic calls the Anthropic streaming API and forwards SSE chunks
// to the client.
//
// ctx should be the request context (r.Context()) so that a client disconnect
// cancels the upstream call instead of streaming on and burning tokens.
//
// onComplete(fullText) is optional and used to populate the lesson cache. It
// only fires when the stream finishes cleanly (context not cancelled, no read
// error), so a disconnect mid-stream never caches a truncated lesson.
func streamFromAnthropic(ctx context.Context, w http.ResponseWriter, system, prompt string, messages []Message, onComplete ...func(string)) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")

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
		log.Printf("stream error: %s", msg)
		data, _ := json.Marshal(map[string]string{"error": msg})
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	if apiKey == "" {
		sendErr("ANTHROPIC_API_KEY is not set. Export it and restart the server.")
		return
	}

	msgs := messages
	if len(msgs) == 0 {
		msgs = []Message{{Role: "user", Content: prompt}}
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model":      model,
		"max_tokens": maxTokens,
		"system":     system,
		"messages":   msgs,
		"stream":     true,
	})

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
			log.Printf("Anthropic API retry %d/%d after %v (last status %d)", attempt, maxRetries, delay, lastStatus)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", anthropicAPI, bytes.NewBuffer(reqBody))
		if err != nil {
			sendErr("Could not build request: " + err.Error())
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")

		resp, err = anthropicClient.Do(req)
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return // client disconnected; nothing to report
			}
			sendErr("Could not reach Anthropic API: " + err.Error())
			return
		}

		if resp.StatusCode == http.StatusOK {
			break
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		lastStatus = resp.StatusCode
		log.Printf("Anthropic API attempt %d/%d: status %d: %s", attempt+1, maxRetries+1, lastStatus, string(body))

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
		buf            strings.Builder // accumulate for the optional onComplete callback
		sawMessageStop bool            // upstream sent its terminal event — the response is whole
		stopReason     string          // from message_delta: "end_turn", "max_tokens", ...
	)
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // tolerate long SSE lines
stream:
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			sawMessageStop = true
			break stream
		}
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
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
			sawMessageStop = true
			break stream
		case "content_block_delta":
			delta, _ := event["delta"].(map[string]interface{})
			if delta == nil || delta["type"] != "text_delta" {
				continue
			}
			text, _ := delta["text"].(string)
			buf.WriteString(text)
			chunk, _ := json.Marshal(map[string]string{"text": text})
			fmt.Fprintf(w, "data: %s\n\n", chunk)
			flusher.Flush()
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
	// terminal event means the upstream dropped the connection mid-response.
	// Report it rather than passing truncated text off as a finished lesson.
	if !sawMessageStop {
		sendErr("Stream ended unexpectedly — try again.")
		return
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()

	// Only a verifiably complete response goes in the cache: message_stop was
	// seen (checked above) and the model wasn't cut off by the max_tokens cap.
	if stopReason == "max_tokens" {
		log.Printf("stream hit the max_tokens cap (%d) — truncated response not cached", maxTokens)
		return
	}
	if len(onComplete) > 0 && onComplete[0] != nil && buf.Len() > 0 {
		onComplete[0](buf.String())
	}
}
