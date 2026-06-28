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
)

const (
	usersFile    = "data/users.json"
	sessionsFile = "data/sessions.json"
)

// ── DEV LOGIN BYPASS ──────────────────────────────────
// For testing only: when devAutoLogin is true, every request is treated as
// coming from devUser, so the login screen is skipped entirely. Flip this back
// to false (or comment out the block in getSessionUser) before real use.
const (
	devAutoLogin = true
	devUser      = "tester"
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
	usersMu.Lock()
	defer usersMu.Unlock()
	json.Unmarshal(data, &users)
}

func saveUsers() {
	usersMu.RLock()
	data, err := json.MarshalIndent(users, "", "  ")
	usersMu.RUnlock()
	if err != nil {
		log.Printf("saveUsers: %v", err)
		return
	}
	writeFileAtomic(usersFile, data, 0600)
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
	data, err := json.MarshalIndent(sessions, "", "  ")
	sessionsMu.RUnlock()
	if err != nil {
		log.Printf("saveSessions: %v", err)
		return
	}
	writeFileAtomic(sessionsFile, data, 0600)
}

func getSessionUser(r *http.Request) (string, bool) {
	// DEV LOGIN BYPASS — see the devAutoLogin const above. Comment out this
	// block to restore the normal cookie-based login.
	if devAutoLogin {
		return devUser, true
	}

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
