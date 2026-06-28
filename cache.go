package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

const cacheFile = "data/cache.json" // persists generated lesson content

// lessonCache["go:lesson:1"] or ["go:track:http:2"] = full markdown text
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
	data, err := json.MarshalIndent(lessonCache, "", "  ")
	lessonCacheMu.RUnlock()
	if err != nil {
		log.Printf("saveLessonCache: %v", err)
		return
	}
	writeFileAtomic(cacheFile, data, 0644)
}

// cacheGet returns cached content for a key, if present.
func cacheGet(key string) (string, bool) {
	lessonCacheMu.RLock()
	defer lessonCacheMu.RUnlock()
	v, ok := lessonCache[key]
	return v, ok
}

// cacheHas reports whether a key is cached (used to flag pre-generated lessons
// in list responses).
func cacheHas(key string) bool {
	lessonCacheMu.RLock()
	defer lessonCacheMu.RUnlock()
	_, ok := lessonCache[key]
	return ok
}

// cacheStore saves content under a key and persists the cache to disk.
func cacheStore(key, content string) {
	lessonCacheMu.Lock()
	lessonCache[key] = content
	lessonCacheMu.Unlock()
	go saveLessonCache()
	log.Printf("cached: %s", key)
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
