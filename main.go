package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	anthropicAPI = "https://api.anthropic.com/v1/messages"
	model        = "claude-sonnet-4-6"
	usersFile    = "data/users.json"
	sessionsFile = "data/sessions.json"
	progressFile = "data/progress.json"
	cacheFile    = "data/cache.json" // persists generated lesson content
)

// ── Message ───────────────────────────────────────────

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ── Users ─────────────────────────────────────────────

type UserRecord struct {
	PINHash string `json:"pin_hash"`
}

var (
	users   = map[string]UserRecord{}
	usersMu sync.RWMutex
)

func loadUsers() {
	data, err := os.ReadFile(usersFile)
	if err != nil {
		return
	}
	usersMu.Lock()
	defer usersMu.Unlock()
	json.Unmarshal(data, &users)
}

func saveUsers() {
	usersMu.RLock()
	data, _ := json.MarshalIndent(users, "", "  ")
	usersMu.RUnlock()
	os.WriteFile(usersFile, data, 0600)
}

// ── Sessions ──────────────────────────────────────────

var (
	sessions   = map[string]string{} // token → username
	sessionsMu sync.RWMutex
)

func loadSessions() {
	data, err := os.ReadFile(sessionsFile)
	if err != nil {
		return
	}
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	json.Unmarshal(data, &sessions)
}

func saveSessions() {
	sessionsMu.RLock()
	data, _ := json.MarshalIndent(sessions, "", "  ")
	sessionsMu.RUnlock()
	os.WriteFile(sessionsFile, data, 0600)
}

func getSessionUser(r *http.Request) (string, bool) {
	c, err := r.Cookie("session")
	if err != nil {
		return "", false
	}
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	user, ok := sessions[c.Value]
	return user, ok
}

// requireAuth wraps a handler that needs a username.
func requireAuth(next func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := getSessionUser(r)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		next(w, r, user)
	}
}

// ── Progress ──────────────────────────────────────────

// progress[username][langID][topicID] = completed
// Topic IDs are stored as strings because JSON keys are always strings.
var (
	progress   = map[string]map[string]map[string]bool{}
	progressMu sync.RWMutex
)

func loadProgress() {
	data, err := os.ReadFile(progressFile)
	if err != nil {
		return
	}
	progressMu.Lock()
	defer progressMu.Unlock()
	json.Unmarshal(data, &progress)
}

func saveProgress() {
	progressMu.RLock()
	data, _ := json.MarshalIndent(progress, "", "  ")
	progressMu.RUnlock()
	os.WriteFile(progressFile, data, 0600)
}

func getUserLangProgress(username, langID string) map[string]bool {
	progressMu.RLock()
	defer progressMu.RUnlock()
	if progress[username] == nil {
		return map[string]bool{}
	}
	src := progress[username][langID]
	out := make(map[string]bool, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

// ── Auth helpers ──────────────────────────────────────

func hashPIN(pin string) string {
	h := sha256.Sum256([]byte(pin))
	return hex.EncodeToString(h[:])
}

func newToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func setCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400 * 30, // 30 days
	})
}

func clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

func jsonErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}

func jsonOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// ── Auth handlers ─────────────────────────────────────

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Username string `json:"username"`
		PIN      string `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, 400, "invalid request")
		return
	}

	req.Username = strings.TrimSpace(strings.ToLower(req.Username))
	req.PIN = strings.TrimSpace(req.PIN)

	if len(req.Username) < 2 || len(req.Username) > 32 {
		jsonErr(w, 400, "username must be 2–32 characters")
		return
	}
	if len(req.PIN) < 4 || len(req.PIN) > 6 {
		jsonErr(w, 400, "PIN must be 4–6 digits")
		return
	}
	for _, c := range req.PIN {
		if c < '0' || c > '9' {
			jsonErr(w, 400, "PIN must contain digits only")
			return
		}
	}

	usersMu.Lock()
	if _, exists := users[req.Username]; exists {
		usersMu.Unlock()
		jsonErr(w, 409, "username already taken")
		return
	}
	users[req.Username] = UserRecord{PINHash: hashPIN(req.PIN)}
	usersMu.Unlock()
	saveUsers()

	token := newToken()
	sessionsMu.Lock()
	sessions[token] = req.Username
	sessionsMu.Unlock()
	saveSessions()

	setCookie(w, token)
	jsonOK(w, map[string]string{"username": req.Username})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Username string `json:"username"`
		PIN      string `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, 400, "invalid request")
		return
	}

	req.Username = strings.TrimSpace(strings.ToLower(req.Username))

	usersMu.RLock()
	user, exists := users[req.Username]
	usersMu.RUnlock()

	if !exists || user.PINHash != hashPIN(req.PIN) {
		jsonErr(w, 401, "invalid username or PIN")
		return
	}

	token := newToken()
	sessionsMu.Lock()
	sessions[token] = req.Username
	sessionsMu.Unlock()
	saveSessions()

	setCookie(w, token)
	jsonOK(w, map[string]string{"username": req.Username})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("session"); err == nil {
		sessionsMu.Lock()
		delete(sessions, c.Value)
		sessionsMu.Unlock()
		saveSessions()
	}
	clearCookie(w)
	jsonOK(w, map[string]bool{"ok": true})
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	user, ok := getSessionUser(r)
	if !ok {
		jsonErr(w, 401, "unauthorized")
		return
	}
	jsonOK(w, map[string]string{"username": user})
}

// ── Language handler ──────────────────────────────────

func handleLanguages(w http.ResponseWriter, r *http.Request) {
	type LangMeta struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Icon        string `json:"icon"`
		Cmd         string `json:"cmd"`
		AccentColor string `json:"accentColor"`
		AccentDark  string `json:"accentDark"`
		AccentGlow  string `json:"accentGlow"`
		CodeLabel   string `json:"codeLabel"`
	}
	metas := make([]LangMeta, 0, len(languageOrder))
	for _, id := range languageOrder {
		l := languages[id]
		metas = append(metas, LangMeta{
			ID:          l.ID,
			Name:        l.Name,
			Icon:        l.Icon,
			Cmd:         l.Cmd,
			AccentColor: l.AccentColor,
			AccentDark:  l.AccentDark,
			AccentGlow:  l.AccentGlow,
			CodeLabel:   l.CodeLabel,
		})
	}
	jsonOK(w, metas)
}

// ── Topics handler ────────────────────────────────────

func handleTopics(w http.ResponseWriter, r *http.Request, user string) {
	langID := r.URL.Query().Get("lang")
	lang, ok := languages[langID]
	if !ok {
		jsonErr(w, 400, "unknown language")
		return
	}

	done := getUserLangProgress(user, langID)

	type TopicResp struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Completed    bool   `json:"completed"`
		LessonCached bool   `json:"lessonCached"`
	}
	result := make([]TopicResp, len(lang.Topics))
	for i, t := range lang.Topics {
		ck := fmt.Sprintf("%s:lesson:%d", langID, t.ID)
		lessonCacheMu.RLock()
		_, cached := lessonCache[ck]
		lessonCacheMu.RUnlock()
		result[i] = TopicResp{
			ID:           t.ID,
			Name:         t.Name,
			Completed:    done[fmt.Sprintf("%d", t.ID)],
			LessonCached: cached,
		}
	}
	jsonOK(w, result)
}

// ── Progress handler ──────────────────────────────────

func handleProgress(w http.ResponseWriter, r *http.Request, user string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Lang      string `json:"lang"`
		TopicID   int    `json:"topic_id"`  // fundamentals: non-zero
		TrackID   string `json:"track_id"`  // track: non-empty → key = "track:<id>:<lesson>"
		LessonID  int    `json:"lesson_id"` // track lesson number
		Completed bool   `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, 400, "invalid request")
		return
	}
	if _, ok := languages[req.Lang]; !ok {
		jsonErr(w, 400, "unknown language")
		return
	}

	var key string
	if req.TrackID != "" {
		key = fmt.Sprintf("track:%s:%d", req.TrackID, req.LessonID)
	} else {
		key = fmt.Sprintf("%d", req.TopicID)
	}

	progressMu.Lock()
	if progress[user] == nil {
		progress[user] = map[string]map[string]bool{}
	}
	if progress[user][req.Lang] == nil {
		progress[user][req.Lang] = map[string]bool{}
	}
	progress[user][req.Lang][key] = req.Completed
	progressMu.Unlock()

	saveProgress()
	jsonOK(w, map[string]bool{"success": true})
}

// ── Lesson cache ──────────────────────────────────────

// lessonCache["go:lesson:1"] = full markdown text
var (
	lessonCache   = map[string]string{}
	lessonCacheMu sync.RWMutex
)

func loadLessonCache() {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return
	}
	lessonCacheMu.Lock()
	defer lessonCacheMu.Unlock()
	json.Unmarshal(data, &lessonCache)
}

func saveLessonCache() {
	lessonCacheMu.RLock()
	data, _ := json.MarshalIndent(lessonCache, "", "  ")
	lessonCacheMu.RUnlock()
	os.WriteFile(cacheFile, data, 0644)
}

// streamCached sends already-generated content as a single SSE chunk.
func streamCached(w http.ResponseWriter, content string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}
	chunk, _ := json.Marshal(map[string]string{"text": content})
	fmt.Fprintf(w, "data: %s\n\n", chunk)
	flusher.Flush()
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

// streamFromAnthropic calls the Anthropic streaming API and forwards SSE chunks
// to the client. Pass an optional onComplete(fullText) callback — used by
// handleLesson to save the result to the lesson cache.
func streamFromAnthropic(w http.ResponseWriter, system, prompt string, messages []Message, onComplete ...func(string)) {
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

	req, _ := http.NewRequest("POST", anthropicAPI, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
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

	var buf strings.Builder // accumulate for optional onComplete callback
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		if event["type"] == "content_block_delta" {
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

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()

	// Fire the cache callback if one was provided
	if len(onComplete) > 0 && onComplete[0] != nil && buf.Len() > 0 {
		onComplete[0](buf.String())
	}
}

// ── Content handlers ──────────────────────────────────

func handleLesson(w http.ResponseWriter, r *http.Request, user string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Lang      string `json:"lang"`
		TopicID   int    `json:"topic_id"`
		TopicName string `json:"topic_name"`
		Force     bool   `json:"force"` // true = bypass cache (Regenerate button)
	}
	json.NewDecoder(r.Body).Decode(&req)

	lang, ok := languages[req.Lang]
	if !ok {
		http.Error(w, "unknown language", http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("%s:lesson:%d", req.Lang, req.TopicID)

	// Serve from cache unless the user explicitly asked to regenerate
	if !req.Force {
		lessonCacheMu.RLock()
		cached, hit := lessonCache[key]
		lessonCacheMu.RUnlock()
		if hit {
			log.Printf("cache hit: %s", key)
			streamCached(w, cached)
			return
		}
	}

	prompt := fmt.Sprintf(`Teach Topic %d of %d: **%s** in %s.

Use this structure:

## Overview
2–3 sentences introducing the concept and why it matters in %s.

## Key Concepts
5–7 bullet points, each with a brief explanation.

## Code Examples

### Example 1: [Descriptive title]
Show and explain a minimal, clear example.

### Example 2: [Descriptive title]
Show a more practical, realistic usage.

## Common Pitfalls
3 specific mistakes beginners make with this topic and how to avoid them.

## Summary
One crisp sentence: the essential takeaway.`,
		req.TopicID, len(lang.Topics), req.TopicName, lang.Name, lang.Name)

	// Stream from Anthropic and save the result to cache when done
	streamFromAnthropic(w, lang.SystemPrompt, prompt, nil, func(full string) {
		lessonCacheMu.Lock()
		lessonCache[key] = full
		lessonCacheMu.Unlock()
		go saveLessonCache()
		log.Printf("cached lesson: %s", key)
	})
}

func handleChallenge(w http.ResponseWriter, r *http.Request, user string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Lang      string `json:"lang"`
		TopicID   int    `json:"topic_id"`
		TopicName string `json:"topic_name"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	lang, ok := languages[req.Lang]
	if !ok {
		http.Error(w, "unknown language", http.StatusBadRequest)
		return
	}

	prompt := fmt.Sprintf(`Create a practical %s coding challenge for Topic %d: **%s**.

## Challenge: [Creative, specific title]

**Difficulty**: Beginner / Intermediate

**Task**
2–3 sentence description of what to implement. Be concrete and specific.

**Requirements**
- Specific, testable requirement
- Another requirement
- One more requirement

**Example**
`+"```"+`
Input:  ...
Output: ...
`+"```"+`

**Starter Code**
%s

Keep it focused purely on %s concepts. Make it engaging with a real-world flavor.`,
		lang.Name, req.TopicID, req.TopicName, lang.StarterTemplate, req.TopicName)

	streamFromAnthropic(w, lang.SystemPrompt, prompt, nil)
}

func handleEvaluate(w http.ResponseWriter, r *http.Request, user string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Lang      string `json:"lang"`
		TopicID   int    `json:"topic_id"`
		TopicName string `json:"topic_name"`
		Code      string `json:"code"`
		Challenge string `json:"challenge"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	lang, ok := languages[req.Lang]
	if !ok {
		http.Error(w, "unknown language", http.StatusBadRequest)
		return
	}

	prompt := fmt.Sprintf(`Evaluate this %s code submission for Topic %d: **%s**.

**The Challenge**:
%s

**Submitted Code**:
`+"```%s\n%s\n```"+`

## Verdict
✅ **Pass** OR ❌ **Needs Work** — state it clearly on the first line.

## What Works Well
Specific praise for correct approaches and %s idioms used well.

## Issues Found
Specific bugs, logical errors, or style issues. Write "None — looks good!" if clean.

## %s Style Note
%s

## Suggested Improvement (if needed)
A corrected or more idiomatic version. Skip if the code passed cleanly.

Be encouraging and educational. Note: code cannot be executed — evaluate on logic and conventions.`,
		lang.Name, req.TopicID, req.TopicName,
		req.Challenge,
		lang.ID, req.Code,
		lang.Name,
		lang.Name, lang.StyleNote)

	streamFromAnthropic(w, lang.SystemPrompt, prompt, nil)
}

func handleHint(w http.ResponseWriter, r *http.Request, user string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Lang      string `json:"lang"`
		TopicID   int    `json:"topic_id"`
		TopicName string `json:"topic_name"`
		Challenge string `json:"challenge"`
		Code      string `json:"code"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	lang, ok := languages[req.Lang]
	if !ok {
		http.Error(w, "unknown language", http.StatusBadRequest)
		return
	}

	prompt := fmt.Sprintf(`Give exactly ONE helpful hint for this %s challenge on **%s**.

Challenge:
%s

Student's current code:
`+"```%s\n%s\n```"+`

Give ONE specific, encouraging nudge that moves them forward without revealing the answer. Maximum 3 sentences.`,
		lang.Name, req.TopicName, req.Challenge, lang.ID, req.Code)

	streamFromAnthropic(w, lang.SystemPrompt, prompt, nil)
}

func handleChat(w http.ResponseWriter, r *http.Request, user string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Lang      string    `json:"lang"`
		TopicID   int       `json:"topic_id"`
		TopicName string    `json:"topic_name"`
		Messages  []Message `json:"messages"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	lang, ok := languages[req.Lang]
	if !ok {
		http.Error(w, "unknown language", http.StatusBadRequest)
		return
	}

	system := fmt.Sprintf(`%s
The student is studying Topic %d: %s. Answer their questions clearly, concisely, and encouragingly.`,
		lang.SystemPrompt, req.TopicID, req.TopicName)

	streamFromAnthropic(w, system, "", req.Messages)
}

// ── Track helpers ─────────────────────────────────────

// buildTrackContext returns a markdown block summarising lessons already covered.
func buildTrackContext(track Track, upToLessonID int) string {
	if upToLessonID <= 1 {
		return "This is the first lesson in this track — no prior context needed."
	}
	var sb strings.Builder
	sb.WriteString("**What previous lessons in this track covered:**\n")
	for _, l := range track.Lessons {
		if l.ID >= upToLessonID {
			break
		}
		fmt.Fprintf(&sb, "- Lesson %d — **%s**: %s\n", l.ID, l.Title, l.Summary)
	}
	return sb.String()
}

// findTrackLesson looks up a track and lesson by ID within a language.
func findTrackLesson(lang Language, trackID string, lessonID int) (Track, TrackLesson, bool) {
	for _, t := range lang.Tracks {
		if t.ID == trackID {
			for _, l := range t.Lessons {
				if l.ID == lessonID {
					return t, l, true
				}
			}
		}
	}
	return Track{}, TrackLesson{}, false
}

// ── Track handlers ────────────────────────────────────

func handleTracks(w http.ResponseWriter, r *http.Request, user string) {
	langID := r.URL.Query().Get("lang")
	lang, ok := languages[langID]
	if !ok {
		jsonErr(w, 400, "unknown language")
		return
	}

	progressMu.RLock()
	userProgress := progress[user][langID]
	progressMu.RUnlock()

	type LessonResp struct {
		ID           int    `json:"id"`
		Title        string `json:"title"`
		Summary      string `json:"summary"`
		Completed    bool   `json:"completed"`
		LessonCached bool   `json:"lessonCached"`
	}
	type TrackResp struct {
		ID          string       `json:"id"`
		Title       string       `json:"title"`
		Icon        string       `json:"icon"`
		Description string       `json:"description"`
		Lessons     []LessonResp `json:"lessons"`
	}

	result := make([]TrackResp, 0, len(lang.Tracks))
	for _, t := range lang.Tracks {
		lessons := make([]LessonResp, len(t.Lessons))
		for i, l := range t.Lessons {
			ck := fmt.Sprintf("%s:track:%s:%d", langID, t.ID, l.ID)
			lessonCacheMu.RLock()
			_, cached := lessonCache[ck]
			lessonCacheMu.RUnlock()
			pkey := fmt.Sprintf("track:%s:%d", t.ID, l.ID)
			lessons[i] = LessonResp{
				ID:           l.ID,
				Title:        l.Title,
				Summary:      l.Summary,
				Completed:    userProgress[pkey],
				LessonCached: cached,
			}
		}
		result = append(result, TrackResp{
			ID:          t.ID,
			Title:       t.Title,
			Icon:        t.Icon,
			Description: t.Description,
			Lessons:     lessons,
		})
	}
	jsonOK(w, result)
}

func handleTrackLesson(w http.ResponseWriter, r *http.Request, user string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Lang     string `json:"lang"`
		TrackID  string `json:"track_id"`
		LessonID int    `json:"lesson_id"`
		Force    bool   `json:"force"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	lang, ok := languages[req.Lang]
	if !ok {
		http.Error(w, "unknown language", http.StatusBadRequest)
		return
	}
	track, lesson, ok := findTrackLesson(lang, req.TrackID, req.LessonID)
	if !ok {
		http.Error(w, "unknown track or lesson", http.StatusBadRequest)
		return
	}

	cacheKey := fmt.Sprintf("%s:track:%s:%d", req.Lang, req.TrackID, req.LessonID)
	if !req.Force {
		lessonCacheMu.RLock()
		cached, hit := lessonCache[cacheKey]
		lessonCacheMu.RUnlock()
		if hit {
			log.Printf("cache hit: %s", cacheKey)
			streamCached(w, cached)
			return
		}
	}

	prevCtx := buildTrackContext(track, lesson.ID)
	continuity := ""
	if lesson.ID > 1 {
		continuity = "\n### Building on previous lessons\nShow explicitly how this lesson's code extends or works alongside what was built before."
	}

	prompt := fmt.Sprintf(`You are teaching Lesson %d of %d in the **%s** track for %s.

**Track goal:** %s

%s

---

## Your task: teach **%s**

Structure your lesson as follows:

## Overview
What this lesson introduces and how it builds on what came before (2–3 sentences).

## Key Concepts
5–7 new concepts, each with a brief explanation.

## Code Examples

### Example 1: [Descriptive title]
A clear, focused example of the core new concept.

### Example 2: [Descriptive title]
A more complete or practical usage, building directly on Example 1.
%s

## Common Mistakes
2–3 mistakes specific to this lesson's topic.

## Summary
One sentence: what the student can now do that they couldn't before this lesson.`,
		lesson.ID, len(track.Lessons), track.Title, lang.Name,
		track.Description,
		prevCtx,
		lesson.Title,
		continuity,
	)

	streamFromAnthropic(w, lang.SystemPrompt, prompt, nil, func(full string) {
		lessonCacheMu.Lock()
		lessonCache[cacheKey] = full
		lessonCacheMu.Unlock()
		go saveLessonCache()
		log.Printf("cached track lesson: %s", cacheKey)
	})
}

func handleTrackChallenge(w http.ResponseWriter, r *http.Request, user string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Lang     string `json:"lang"`
		TrackID  string `json:"track_id"`
		LessonID int    `json:"lesson_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	lang, ok := languages[req.Lang]
	if !ok {
		http.Error(w, "unknown language", http.StatusBadRequest)
		return
	}
	track, lesson, ok := findTrackLesson(lang, req.TrackID, req.LessonID)
	if !ok {
		http.Error(w, "unknown track or lesson", http.StatusBadRequest)
		return
	}

	prevCtx := buildTrackContext(track, lesson.ID)

	prompt := fmt.Sprintf(`Create a coding challenge for Lesson %d: **%s** in the **%s** track (%s).

%s

The challenge must:
- Focus specifically on **%s** concepts
- Build naturally on the previous lessons (reference and extend their patterns)
- Be completable in ~15–20 minutes

## Challenge: [Creative title]

**Difficulty**: Beginner / Intermediate

**Task**
2–3 sentences describing what to build. Be concrete and specific.

**Requirements**
- Specific, testable requirement
- Another requirement
- One more requirement

**Example**
`+"```"+`
Input:  ...
Output: ...
`+"```"+`

**Starter Code**
%s`,
		lesson.ID, lesson.Title, track.Title, lang.Name,
		prevCtx,
		lesson.Title,
		lang.StarterTemplate,
	)

	streamFromAnthropic(w, lang.SystemPrompt, prompt, nil)
}

func handleTrackEvaluate(w http.ResponseWriter, r *http.Request, user string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Lang      string `json:"lang"`
		TrackID   string `json:"track_id"`
		LessonID  int    `json:"lesson_id"`
		Code      string `json:"code"`
		Challenge string `json:"challenge"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	lang, ok := languages[req.Lang]
	if !ok {
		http.Error(w, "unknown language", http.StatusBadRequest)
		return
	}
	track, lesson, ok := findTrackLesson(lang, req.TrackID, req.LessonID)
	if !ok {
		http.Error(w, "unknown track or lesson", http.StatusBadRequest)
		return
	}

	prompt := fmt.Sprintf(`Evaluate this %s code submission for the **%s** track, Lesson %d: **%s**.

**The Challenge:**
%s

**Submitted Code:**
`+"```%s\n%s\n```"+`

## Verdict
✅ **Pass** OR ❌ **Needs Work** — state it clearly on the first line.

## What Works Well
Specific praise, especially for using patterns from the %s track correctly.

## Issues Found
Specific bugs, logical errors, or style issues. Write "None — looks good!" if clean.

## %s Style Note
%s

## Suggested Improvement (if needed)
A corrected or more idiomatic version. Skip if the code passed cleanly.

Be encouraging. Note: code cannot be executed — evaluate on logic and conventions.`,
		lang.Name, track.Title, lesson.ID, lesson.Title,
		req.Challenge,
		lang.ID, req.Code,
		track.Title,
		lang.Name, lang.StyleNote,
	)

	streamFromAnthropic(w, lang.SystemPrompt, prompt, nil)
}

func handleTrackHint(w http.ResponseWriter, r *http.Request, user string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Lang      string `json:"lang"`
		TrackID   string `json:"track_id"`
		LessonID  int    `json:"lesson_id"`
		Challenge string `json:"challenge"`
		Code      string `json:"code"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	lang, ok := languages[req.Lang]
	if !ok {
		http.Error(w, "unknown language", http.StatusBadRequest)
		return
	}
	track, lesson, ok := findTrackLesson(lang, req.TrackID, req.LessonID)
	if !ok {
		http.Error(w, "unknown track or lesson", http.StatusBadRequest)
		return
	}

	prompt := fmt.Sprintf(`Give ONE helpful hint for this %s challenge on **%s** (track: %s).

Challenge:
%s

Student's current code:
`+"```%s\n%s\n```"+`

Give ONE specific, encouraging nudge. Maximum 3 sentences. Don't reveal the answer.`,
		lang.Name, lesson.Title, track.Title,
		req.Challenge, lang.ID, req.Code,
	)

	streamFromAnthropic(w, lang.SystemPrompt, prompt, nil)
}

func handleTrackChat(w http.ResponseWriter, r *http.Request, user string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Lang     string    `json:"lang"`
		TrackID  string    `json:"track_id"`
		LessonID int       `json:"lesson_id"`
		Messages []Message `json:"messages"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	lang, ok := languages[req.Lang]
	if !ok {
		http.Error(w, "unknown language", http.StatusBadRequest)
		return
	}
	track, lesson, ok := findTrackLesson(lang, req.TrackID, req.LessonID)
	if !ok {
		http.Error(w, "unknown track or lesson", http.StatusBadRequest)
		return
	}

	system := fmt.Sprintf(`%s
The student is working through the **%s** track, Lesson %d: %s.
Answer their questions clearly and in the context of this specific lesson and track.`,
		lang.SystemPrompt, track.Title, lesson.ID, lesson.Title,
	)

	streamFromAnthropic(w, system, "", req.Messages)
}

// ── Main ──────────────────────────────────────────────

func main() {
	os.MkdirAll("data", 0755)
	loadUsers()
	loadSessions()
	loadProgress()
	loadLessonCache()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8181"
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	// Auth — no session required
	http.HandleFunc("/api/auth/register", handleRegister)
	http.HandleFunc("/api/auth/login", handleLogin)
	http.HandleFunc("/api/auth/logout", handleLogout)
	http.HandleFunc("/api/auth/me", handleMe)

	// Languages — no session required
	http.HandleFunc("/api/languages", handleLanguages)

	// Everything below requires a valid session
	http.HandleFunc("/api/topics", requireAuth(handleTopics))
	http.HandleFunc("/api/progress", requireAuth(handleProgress))
	http.HandleFunc("/api/lesson", requireAuth(handleLesson))
	http.HandleFunc("/api/challenge", requireAuth(handleChallenge))
	http.HandleFunc("/api/evaluate", requireAuth(handleEvaluate))
	http.HandleFunc("/api/hint", requireAuth(handleHint))
	http.HandleFunc("/api/chat", requireAuth(handleChat))

	// Advanced tracks
	http.HandleFunc("/api/tracks", requireAuth(handleTracks))
	http.HandleFunc("/api/track/lesson", requireAuth(handleTrackLesson))
	http.HandleFunc("/api/track/challenge", requireAuth(handleTrackChallenge))
	http.HandleFunc("/api/track/evaluate", requireAuth(handleTrackEvaluate))
	http.HandleFunc("/api/track/hint", requireAuth(handleTrackHint))
	http.HandleFunc("/api/track/chat", requireAuth(handleTrackChat))

	log.Printf("🚀  Coding Bootcamp → http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
