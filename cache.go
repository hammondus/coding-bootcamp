package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

const cacheFile = "data/cache.json" // persists generated lesson & challenge content

// lessonCache holds generated markdown per user so content (and the
// "regenerate" variations a student produces) is never shared between accounts.
// It is a two-level map: username → content-key → markdown, where the
// content-key names the kind, e.g.
//   "go:lesson:1"               — fundamentals lesson
//   "go:challenge:1"            — fundamentals challenge
//   "go:track:http:2"           — advanced track lesson
//   "go:track:http:challenge:2" — advanced track challenge
var (
	lessonCache   = map[string]map[string]string{}
	lessonCacheMu sync.RWMutex
)

func loadLessonCache() {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return
	}
	// Decode into a temp first so a parse failure (e.g. an older, pre-per-user
	// cache file) leaves the live cache as a clean empty map rather than a
	// half-populated one. Content just gets regenerated on demand.
	var loaded map[string]map[string]string
	if err := json.Unmarshal(data, &loaded); err != nil {
		log.Printf("loadLessonCache: ignoring incompatible cache file: %v", err)
		return
	}
	lessonCacheMu.Lock()
	defer lessonCacheMu.Unlock()
	lessonCache = loaded
}

func saveLessonCache() {
	writeFileAtomic(cacheFile, 0644, func() ([]byte, error) {
		lessonCacheMu.RLock()
		defer lessonCacheMu.RUnlock()
		return json.MarshalIndent(lessonCache, "", "  ")
	})
}

// cacheGet returns the user's cached content for a key, if present.
func cacheGet(user, key string) (string, bool) {
	lessonCacheMu.RLock()
	defer lessonCacheMu.RUnlock()
	u := lessonCache[user]
	if u == nil {
		return "", false
	}
	v, ok := u[key]
	return v, ok
}

// cacheHas reports whether a key is cached for the user (used to flag
// pre-generated lessons in list responses).
func cacheHas(user, key string) bool {
	lessonCacheMu.RLock()
	defer lessonCacheMu.RUnlock()
	u := lessonCache[user]
	if u == nil {
		return false
	}
	_, ok := u[key]
	return ok
}

// cacheStore saves content under a key for the user and persists the cache to
// disk.
func cacheStore(user, key, content string) {
	lessonCacheMu.Lock()
	if lessonCache[user] == nil {
		lessonCache[user] = map[string]string{}
	}
	lessonCache[user][key] = content
	lessonCacheMu.Unlock()
	go saveLessonCache()
	log.Printf("cached: %s/%s", user, key)
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
