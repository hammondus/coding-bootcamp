package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	usersFile    = "data/users.json"
	sessionsFile = "data/sessions.json"
)

// ── DEV LOGIN BYPASS ──────────────────────────────────
// For local testing only. When devAutoLogin is true, every request is treated
// as coming from devUser, so the login screen is skipped entirely. These are
// set by the -dev / -dev-user startup flags in main.go and are off by default.
var (
	devAutoLogin = false
	devUser      = "dev"
)

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
	// Decode into a temp first so a corrupt file leaves the live map clean and
	// logs loudly, rather than silently starting with a half-populated map
	// that the next save would write back over the real data. (Same pattern
	// as loadLessonCache.)
	var loaded map[string]UserRecord
	if err := json.Unmarshal(data, &loaded); err != nil {
		log.Printf("loadUsers: ignoring unreadable %s: %v", usersFile, err)
		return
	}
	usersMu.Lock()
	defer usersMu.Unlock()
	users = loaded
}

func saveUsers() {
	writeFileAtomic(usersFile, 0600, func() ([]byte, error) {
		usersMu.RLock()
		defer usersMu.RUnlock()
		return json.MarshalIndent(users, "", "  ")
	})
}

// ── Sessions ──────────────────────────────────────────

// sessionTTL matches the session cookie's MaxAge, so a token the browser has
// already discarded doesn't live on forever in sessions.json.
const sessionTTL = 30 * 24 * time.Hour

// Session records who a token belongs to and when it was issued.
type Session struct {
	User    string    `json:"user"`
	Created time.Time `json:"created"`
}

var (
	sessions   = map[string]Session{} // token → session
	sessionsMu sync.RWMutex
)

func loadSessions() {
	data, err := os.ReadFile(sessionsFile)
	if err != nil {
		return
	}
	// Temp-decode for the same reason as loadUsers. An older sessions file
	// (the pre-timestamp token→username format) fails here and is dropped —
	// everyone just signs in again.
	var loaded map[string]Session
	if err := json.Unmarshal(data, &loaded); err != nil {
		log.Printf("loadSessions: ignoring unreadable %s: %v", sessionsFile, err)
		return
	}
	// Prune expired tokens so the file doesn't accumulate them forever.
	for token, s := range loaded {
		if time.Since(s.Created) > sessionTTL {
			delete(loaded, token)
		}
	}
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	sessions = loaded
}

func saveSessions() {
	writeFileAtomic(sessionsFile, 0600, func() ([]byte, error) {
		sessionsMu.RLock()
		defer sessionsMu.RUnlock()
		return json.MarshalIndent(sessions, "", "  ")
	})
}

func getSessionUser(r *http.Request) (string, bool) {
	// DEV LOGIN BYPASS — enabled by the -dev startup flag (see main.go).
	if devAutoLogin {
		return devUser, true
	}

	c, err := r.Cookie("session")
	if err != nil {
		return "", false
	}
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	s, ok := sessions[c.Value]
	if !ok || time.Since(s.Created) > sessionTTL {
		return "", false
	}
	return s.User, true
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
	sessions[token] = Session{User: req.Username, Created: time.Now()}
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
	sessions[token] = Session{User: req.Username, Created: time.Now()}
	sessionsMu.Unlock()
	saveSessions()

	setCookie(w, token)
	jsonOK(w, map[string]string{"username": req.Username})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
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
