package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
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
	model        = "claude-sonnet-4-6"
	maxRetries   = 3
)

// anthropicClient is a shared HTTP client configured for the Anthropic
// streaming API. No overall Timeout is set because a streaming response is
// unbounded in duration; ResponseHeaderTimeout guards against a stalled
// upstream that never sends the first byte.
var anthropicClient = &http.Client{
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
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

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Fprintf(w, "data: {\"error\":\"streaming not supported\"}\n\n")
		return
	}
	sendErr := func(msg string) {
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
		"max_tokens": 2048,
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
			if ctx.Err() != nil {
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

	var buf strings.Builder // accumulate for the optional onComplete callback
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
			break stream
		}
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		switch event["type"] {
		case "message_stop":
			// Anthropic's terminal event — the stream is done.
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

	if err := scanner.Err(); err != nil {
		log.Printf("stream read error: %v", err)
	}

	// If the client disconnected, don't bother writing the terminator or
	// caching what could be a partial response.
	if ctx.Err() != nil {
		return
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()

	clean := scanner.Err() == nil
	if clean && len(onComplete) > 0 && onComplete[0] != nil && buf.Len() > 0 {
		onComplete[0](buf.String())
	}
}
