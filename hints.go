package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

const hintsFile = "data/hints.json"

// hintsUsed[username][challengeKey] = true once the student has used any hint
// for that challenge: pressing the 💡 Hint button or revealing the hidden
// Hints section both count. The challengeKey is the same key the challenge is
// cached under (see challengeCacheKey), so the record is per difficulty tier
// and is cleared when the challenge is regenerated. handleEvaluate reads this
// to recognise a no-hints solve in its feedback.
var (
	hintsUsed   = map[string]map[string]bool{}
	hintsUsedMu sync.RWMutex
)

func loadHintsUsed() {
	data, err := os.ReadFile(hintsFile)
	if err != nil {
		return
	}
	// Decode into a temp first so a corrupt file leaves the live map clean and
	// logs loudly instead of silently half-loading. (Same pattern as
	// loadLessonCache.)
	var loaded map[string]map[string]bool
	if err := json.Unmarshal(data, &loaded); err != nil {
		log.Printf("loadHintsUsed: ignoring unreadable %s: %v", hintsFile, err)
		return
	}
	hintsUsedMu.Lock()
	defer hintsUsedMu.Unlock()
	hintsUsed = loaded
}

func saveHintsUsed() {
	writeFileAtomic(hintsFile, 0644, func() ([]byte, error) {
		hintsUsedMu.RLock()
		defer hintsUsedMu.RUnlock()
		return json.MarshalIndent(hintsUsed, "", "  ")
	})
}

// markHintsUsed records that the user has used a hint on a challenge. Skips
// the disk write when the flag is already set (hint requests can repeat).
func markHintsUsed(user, key string) {
	hintsUsedMu.Lock()
	if hintsUsed[user][key] {
		hintsUsedMu.Unlock()
		return
	}
	if hintsUsed[user] == nil {
		hintsUsed[user] = map[string]bool{}
	}
	hintsUsed[user][key] = true
	hintsUsedMu.Unlock()
	go saveHintsUsed()
}

// clearHintsUsed resets the record for one challenge — called when the
// challenge is regenerated, since hint use on the old challenge shouldn't
// count against the new one.
func clearHintsUsed(user, key string) {
	hintsUsedMu.Lock()
	if !hintsUsed[user][key] {
		hintsUsedMu.Unlock()
		return
	}
	delete(hintsUsed[user], key)
	hintsUsedMu.Unlock()
	go saveHintsUsed()
}

// hintsWereUsed reports whether the user has used any hint on a challenge.
func hintsWereUsed(user, key string) bool {
	hintsUsedMu.RLock()
	defer hintsUsedMu.RUnlock()
	return hintsUsed[user][key]
}
