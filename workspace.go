package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

// A var, not a const: tests point this at a throwaway file (see TestMain in
// workspace_test.go) so their background saves can never touch real data.
var workspaceFile = "data/workspaces.json"

// A Solution is the student's last submitted answer to one challenge (at one
// difficulty tier), together with the evaluation feedback it received.
type Solution struct {
	Code     string `json:"code"`
	Feedback string `json:"feedback"`
}

// Saved student work, so a solution and the conversation around it survive a
// reload just like the generated challenge text does. Two maps, both guarded
// by workspaceMu:
//
//	solutions[username][solutionKey] = last submission + its feedback
//	chats[username][chatKey]         = full chat history for a selection
//
// Solutions reuse the challenge cache keys ("go:challenge:1:goat",
// "go:track:http:challenge:2"; for projects the milestone guidance key,
// "go:project:bootcamp:milestone:3"). Chats get their own keys (see
// chatStoreKey and friends) because chat belongs to the topic or lesson as a
// whole, not to one difficulty tier.
var (
	solutions   = map[string]map[string]Solution{}
	chats       = map[string]map[string][]Message{}
	workspaceMu sync.RWMutex
)

// workspaceFileData is the on-disk shape of workspaceFile.
type workspaceFileData struct {
	Solutions map[string]map[string]Solution  `json:"solutions"`
	Chats     map[string]map[string][]Message `json:"chats"`
}

// Chat history storage keys, one per mode.
func chatStoreKey(langID string, topicID int) string {
	return fmt.Sprintf("%s:chat:%d", langID, topicID)
}
func trackChatStoreKey(langID, trackID string, lessonID int) string {
	return fmt.Sprintf("%s:track:%s:chat:%d", langID, trackID, lessonID)
}
func projectChatStoreKey(langID, projectID string, milestoneID int) string {
	return fmt.Sprintf("%s:project:%s:chat:%d", langID, projectID, milestoneID)
}
func setupChatStoreKey(langID string) string {
	return fmt.Sprintf("%s:setup:chat", langID)
}

func loadWorkspaces() {
	data, err := os.ReadFile(workspaceFile)
	if err != nil {
		return
	}
	// Decode into a temp first so a corrupt file is logged and ignored rather
	// than half-populating the live maps (same pattern as loadLessonCache).
	var loaded workspaceFileData
	if err := json.Unmarshal(data, &loaded); err != nil {
		log.Printf("loadWorkspaces: ignoring unreadable %s: %v", workspaceFile, err)
		return
	}
	if loaded.Solutions == nil {
		loaded.Solutions = map[string]map[string]Solution{}
	}
	if loaded.Chats == nil {
		loaded.Chats = map[string]map[string][]Message{}
	}
	workspaceMu.Lock()
	defer workspaceMu.Unlock()
	solutions = loaded.Solutions
	chats = loaded.Chats
}

func saveWorkspaces() {
	writeFileAtomic(workspaceFile, 0644, func() ([]byte, error) {
		workspaceMu.RLock()
		defer workspaceMu.RUnlock()
		return json.MarshalIndent(workspaceFileData{Solutions: solutions, Chats: chats}, "", "  ")
	})
}

// storeSolution records the student's submission and the feedback it earned.
// Each new evaluation of the same challenge replaces the previous one.
func storeSolution(user, key, code, feedback string) {
	workspaceMu.Lock()
	if solutions[user] == nil {
		solutions[user] = map[string]Solution{}
	}
	solutions[user][key] = Solution{Code: code, Feedback: feedback}
	workspaceMu.Unlock()
	go saveWorkspaces()
}

// storeChat replaces the saved chat history for a selection. The slice is
// copied under the lock so the map never shares backing storage with a
// caller's slice (the same snapshot discipline as everywhere else).
func storeChat(user, key string, history []Message) {
	cp := make([]Message, len(history))
	copy(cp, history)
	workspaceMu.Lock()
	if chats[user] == nil {
		chats[user] = map[string][]Message{}
	}
	chats[user][key] = cp
	workspaceMu.Unlock()
	go saveWorkspaces()
}

// getWorkspace returns the saved solution (zero value if none) and a copy of
// the chat history for the given keys.
func getWorkspace(user, solutionKey, chatKey string) (Solution, []Message) {
	workspaceMu.RLock()
	defer workspaceMu.RUnlock()
	sol := solutions[user][solutionKey]
	src := chats[user][chatKey]
	out := make([]Message, len(src))
	copy(out, src)
	return sol, out
}

// ── Workspace handlers ────────────────────────────────
//
// One read endpoint per mode. There are no write endpoints: the server saves
// work as it flows through the existing evaluate and chat handlers, so the
// client can't get out of sync with what was actually evaluated or answered.

// workspaceResp is what the client restores from: the last submitted solution
// and its feedback for one challenge tier, plus the selection's chat history.
type workspaceResp struct {
	Code     string    `json:"code"`
	Feedback string    `json:"feedback"`
	Chat     []Message `json:"chat"`
}

func handleWorkspace(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang       string `json:"lang"`
		TopicID    int    `json:"topic_id"`
		Difficulty string `json:"difficulty"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	if _, ok := lookupLang(w, req.Lang); !ok {
		return
	}
	sol, chat := getWorkspace(user,
		challengeCacheKey(req.Lang, req.TopicID, normalizeDifficulty(req.Difficulty)),
		chatStoreKey(req.Lang, req.TopicID))
	jsonOK(w, workspaceResp{Code: sol.Code, Feedback: sol.Feedback, Chat: chat})
}

func handleTrackWorkspace(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang       string `json:"lang"`
		TrackID    string `json:"track_id"`
		LessonID   int    `json:"lesson_id"`
		Difficulty string `json:"difficulty"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}
	if _, _, ok := lookupTrackLesson(w, lang, req.TrackID, req.LessonID); !ok {
		return
	}
	sol, chat := getWorkspace(user,
		trackChallengeCacheKey(req.Lang, req.TrackID, req.LessonID, normalizeDifficulty(req.Difficulty)),
		trackChatStoreKey(req.Lang, req.TrackID, req.LessonID))
	jsonOK(w, workspaceResp{Code: sol.Code, Feedback: sol.Feedback, Chat: chat})
}

func handleProjectWorkspace(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang        string `json:"lang"`
		ProjectID   string `json:"project_id"`
		MilestoneID int    `json:"milestone_id"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}
	if _, _, ok := lookupProjectMilestone(w, lang, req.ProjectID, req.MilestoneID); !ok {
		return
	}
	// A milestone's guidance doubles as its challenge, so its cache key is the
	// solution key here (there are no difficulty tiers in project mode).
	sol, chat := getWorkspace(user,
		fmt.Sprintf("%s:project:%s:milestone:%d", req.Lang, req.ProjectID, req.MilestoneID),
		projectChatStoreKey(req.Lang, req.ProjectID, req.MilestoneID))
	jsonOK(w, workspaceResp{Code: sol.Code, Feedback: sol.Feedback, Chat: chat})
}
