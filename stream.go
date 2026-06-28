package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	anthropicAPI = "https://api.anthropic.com/v1/messages"
	model        = "claude-sonnet-4-6"
)

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

	req, err := http.NewRequestWithContext(ctx, "POST", anthropicAPI, bytes.NewBuffer(reqBody))
	if err != nil {
		sendErr("Could not build request: " + err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return // client went away; expected, not an error worth reporting
		}
		sendErr("Could not reach Anthropic API: " + err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Anthropic API error %d: %s", resp.StatusCode, string(body))
		sendErr(fmt.Sprintf("API returned status %d — check your API key", resp.StatusCode))
		return
	}

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
