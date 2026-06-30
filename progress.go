package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

const progressFile = "data/progress.json"

// progress[username][langID][topicKey] = completed
// Keys are strings because JSON object keys are always strings. For
// fundamentals the key is the topic ID ("3"); for tracks it's
// "track:<trackID>:<lessonID>".
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
	data, err := json.MarshalIndent(progress, "", "  ")
	progressMu.RUnlock()
	if err != nil {
		log.Printf("saveProgress: %v", err)
		return
	}
	writeFileAtomic(progressFile, data, 0644)
}

// getUserLangProgress returns a copy of the user's completion map for a
// language. Copying under the read lock is what makes concurrent reads safe
// against writes in handleProgress — callers must use this rather than reading
// the shared map directly.
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

func handleProgress(w http.ResponseWriter, r *http.Request, user string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Lang        string `json:"lang"`
		TopicID     int    `json:"topic_id"`     // fundamentals: non-zero
		TrackID     string `json:"track_id"`     // track: non-empty → key = "track:<id>:<lesson>"
		LessonID    int    `json:"lesson_id"`    // track lesson number
		ProjectID   string `json:"project_id"`   // project: non-empty → key = "project:<id>:<milestone>"
		MilestoneID int    `json:"milestone_id"` // project milestone number
		Completed   bool   `json:"completed"`
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
	switch {
	case req.ProjectID != "":
		key = fmt.Sprintf("project:%s:%d", req.ProjectID, req.MilestoneID)
	case req.TrackID != "":
		key = fmt.Sprintf("track:%s:%d", req.TrackID, req.LessonID)
	default:
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
