package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func jsonErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}

func jsonOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// decodePOST enforces a POST method and decodes the JSON body into dst. On
// failure it writes a plain-text error (the content/track endpoints surface
// errors to the client as SSE/text) and returns false. Centralising this also
// makes body-decode errors consistent across handlers instead of silently
// ignored in some and checked in others.
func decodePOST(w http.ResponseWriter, r *http.Request, dst any) bool {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return false
	}
	return true
}

// noCache wraps a handler so the browser is told never to reuse a cached copy
// of the response. It's a local-development convenience: we edit the files in
// static/ constantly, and without this the browser happily serves a stale
// style.css / app.js until you remember to hard-refresh.
//
// "no-store" is the strongest option — the browser doesn't keep the file at
// all, so every request fetches the current version from disk. On a real
// public deployment you'd want caching turned ON instead (for speed), so this
// wrapper would be removed or only applied during development.
func noCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// lookupLang resolves a language ID, writing a 400 and returning ok=false when
// it's unknown.
func lookupLang(w http.ResponseWriter, id string) (Language, bool) {
	lang, ok := languages[id]
	if !ok {
		http.Error(w, "unknown language", http.StatusBadRequest)
	}
	return lang, ok
}

// lookupTopic resolves a fundamentals topic within a language, writing a 400
// and returning ok=false when it's unknown. Handlers use the returned Topic's
// Name and Skills rather than trusting the client-supplied topic name.
func lookupTopic(w http.ResponseWriter, lang Language, topicID int) (Topic, bool) {
	for _, t := range lang.Topics {
		if t.ID == topicID {
			return t, true
		}
	}
	http.Error(w, "unknown topic", http.StatusBadRequest)
	return Topic{}, false
}

// lookupTrackLesson resolves a track + lesson within a language, writing a 400
// and returning ok=false when either is unknown.
func lookupTrackLesson(w http.ResponseWriter, lang Language, trackID string, lessonID int) (Track, TrackLesson, bool) {
	track, lesson, ok := findTrackLesson(lang, trackID, lessonID)
	if !ok {
		http.Error(w, "unknown track or lesson", http.StatusBadRequest)
	}
	return track, lesson, ok
}
