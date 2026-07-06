package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
)

const settingsFile = "data/settings.json"

// userModel[username] = the model ID the user picked in the web UI. One
// global choice per user drives lessons, challenges, hints, chat, and
// evaluation alike.
var (
	userModel   = map[string]string{}
	userModelMu sync.RWMutex
)

func loadUserModels() {
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		return
	}
	// Decode into a temp first so a corrupt file leaves the live map clean and
	// logs loudly instead of silently half-loading. (Same pattern as
	// loadLessonCache.)
	var loaded map[string]string
	if err := json.Unmarshal(data, &loaded); err != nil {
		log.Printf("loadUserModels: ignoring unreadable %s: %v", settingsFile, err)
		return
	}
	userModelMu.Lock()
	defer userModelMu.Unlock()
	userModel = loaded
}

func saveUserModels() {
	writeFileAtomic(settingsFile, 0644, func() ([]byte, error) {
		userModelMu.RLock()
		defer userModelMu.RUnlock()
		return json.MarshalIndent(userModel, "", "  ")
	})
}

// currentModel resolves which model to use for a user right now: their saved
// choice if it's still in the catalog and its provider still has a key,
// otherwise the default. Falling back (rather than erroring) means removing a
// provider's key file simply reverts affected users to the default model.
func currentModel(user string) Model {
	userModelMu.RLock()
	id := userModel[user]
	userModelMu.RUnlock()
	if m, ok := modelByID(id); ok && modelAvailable(m) {
		return m
	}
	def, _ := modelByID(defaultModelID)
	return def
}

// handleModels (GET /api/models) lists the models the user can pick from and
// which one is currently selected.
func handleModels(w http.ResponseWriter, r *http.Request, user string) {
	jsonOK(w, map[string]interface{}{
		"models":   availableModels(),
		"selected": currentModel(user).ID,
	})
}

// handleSetModel (POST /api/model) saves the user's model choice. It takes
// effect immediately: every generation request resolves the model through
// currentModel, so no restart is needed.
func handleSetModel(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Model string `json:"model"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	m, ok := modelByID(req.Model)
	if !ok || !modelAvailable(m) {
		jsonErr(w, 400, "unknown or unavailable model")
		return
	}
	userModelMu.Lock()
	userModel[user] = m.ID
	userModelMu.Unlock()
	go saveUserModels()
	log.Printf("model for %s → %s", user, m.ID)
	jsonOK(w, map[string]string{"selected": m.ID})
}
