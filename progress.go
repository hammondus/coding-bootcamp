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
//
// A challenge completed at one difficulty tier gets its own key with a
// ":challenge:<tier>" suffix ("3:challenge:goat",
// "track:<trackID>:<lessonID>:challenge:goat"). Tier keys are separate from
// the plain topic/lesson key, so finishing a challenge and finishing the
// whole subject are tracked independently.
var (
	progress   = map[string]map[string]map[string]bool{}
	progressMu sync.RWMutex
)

func loadProgress() {
	data, err := os.ReadFile(progressFile)
	if err != nil {
		return
	}
	// Decode into a temp first so a corrupt file leaves the live map clean and
	// logs loudly instead of silently half-loading. (Same pattern as
	// loadLessonCache.)
	var loaded map[string]map[string]map[string]bool
	if err := json.Unmarshal(data, &loaded); err != nil {
		log.Printf("loadProgress: ignoring unreadable %s: %v", progressFile, err)
		return
	}
	progressMu.Lock()
	defer progressMu.Unlock()
	progress = loaded
}

func saveProgress() {
	writeFileAtomic(progressFile, 0644, func() ([]byte, error) {
		progressMu.RLock()
		defer progressMu.RUnlock()
		return json.MarshalIndent(progress, "", "  ")
	})
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
		Difficulty  string `json:"difficulty"`   // non-empty → mark one challenge tier, not the whole subject
		Completed   bool   `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, 400, "invalid request")
		return
	}
	lang, ok := languages[req.Lang]
	if !ok {
		jsonErr(w, 400, "unknown language")
		return
	}

	// Validate the ID against the curriculum before persisting, so a stale or
	// buggy client can't pollute progress.json with keys that don't exist.
	var key string
	switch {
	case req.ProjectID != "":
		if _, _, ok := findProjectMilestone(lang, req.ProjectID, req.MilestoneID); !ok {
			jsonErr(w, 400, "unknown project or milestone")
			return
		}
		key = fmt.Sprintf("project:%s:%d", req.ProjectID, req.MilestoneID)
	case req.TrackID != "":
		if _, _, ok := findTrackLesson(lang, req.TrackID, req.LessonID); !ok {
			jsonErr(w, 400, "unknown track or lesson")
			return
		}
		key = fmt.Sprintf("track:%s:%d", req.TrackID, req.LessonID)
	default:
		found := false
		for _, t := range lang.Topics {
			if t.ID == req.TopicID {
				found = true
				break
			}
		}
		if !found {
			jsonErr(w, 400, "unknown topic")
			return
		}
		key = fmt.Sprintf("%d", req.TopicID)
	}

	// A non-empty difficulty marks a single challenge tier complete rather
	// than the whole topic/lesson. Validate it like the IDs above, so a buggy
	// client can't invent tiers in progress.json.
	if req.Difficulty != "" {
		if req.ProjectID != "" {
			jsonErr(w, 400, "projects have no challenge tiers")
			return
		}
		if _, ok := difficultySpec[req.Difficulty]; !ok {
			jsonErr(w, 400, "unknown difficulty")
			return
		}
		key += ":challenge:" + req.Difficulty
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
