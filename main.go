package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
	// -dev bypasses the login screen entirely and auto-authenticates every
	// request as the dev user. Local testing only — never enable in a shared
	// deployment. -dev-user picks which username that is.
	dev := flag.Bool("dev", false, "dev mode: bypass login, auto-authenticate as the dev user")
	devUserFlag := flag.String("dev-user", "dev", "username to auto-login as in dev mode")
	flag.Parse()
	if *dev {
		devAutoLogin = true
		devUser = *devUserFlag
		log.Printf("⚠️  DEV MODE — login bypassed, auto-logged in as %q", devUser)
	}

	os.MkdirAll("data", 0755)
	loadUsers()
	loadSessions()
	loadProgress()
	loadLessonCache()
	loadHintsUsed()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8181"
	}

	// Static assets and index.html are wrapped in noCache so the browser
	// always picks up the latest files during local development.
	http.Handle("/static/", noCache(http.StripPrefix("/static/", http.FileServer(http.Dir("static")))))
	http.Handle("/", noCache(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})))

	// Auth — no session required
	http.HandleFunc("/api/auth/register", handleRegister)
	http.HandleFunc("/api/auth/login", handleLogin)
	http.HandleFunc("/api/auth/logout", handleLogout)
	http.HandleFunc("/api/auth/me", handleMe)

	// Languages — no session required
	http.HandleFunc("/api/languages", handleLanguages)

	// Everything below requires a valid session
	http.HandleFunc("/api/topics", requireAuth(handleTopics))
	http.HandleFunc("/api/progress", requireAuth(handleProgress))
	http.HandleFunc("/api/lesson", requireAuth(handleLesson))
	http.HandleFunc("/api/challenge", requireAuth(handleChallenge))
	http.HandleFunc("/api/evaluate", requireAuth(handleEvaluate))
	http.HandleFunc("/api/hint", requireAuth(handleHint))
	http.HandleFunc("/api/hints-viewed", requireAuth(handleHintsViewed))
	http.HandleFunc("/api/chat", requireAuth(handleChat))

	// Advanced tracks
	http.HandleFunc("/api/tracks", requireAuth(handleTracks))
	http.HandleFunc("/api/track/lesson", requireAuth(handleTrackLesson))
	http.HandleFunc("/api/track/challenge", requireAuth(handleTrackChallenge))
	http.HandleFunc("/api/track/evaluate", requireAuth(handleTrackEvaluate))
	http.HandleFunc("/api/track/hint", requireAuth(handleTrackHint))
	http.HandleFunc("/api/track/hints-viewed", requireAuth(handleTrackHintsViewed))
	http.HandleFunc("/api/track/chat", requireAuth(handleTrackChat))

	// Capstone projects
	http.HandleFunc("/api/projects", requireAuth(handleProjects))
	http.HandleFunc("/api/project/brief", requireAuth(handleProjectBrief))
	http.HandleFunc("/api/project/milestone", requireAuth(handleProjectMilestone))
	http.HandleFunc("/api/project/evaluate", requireAuth(handleProjectEvaluate))
	http.HandleFunc("/api/project/hint", requireAuth(handleProjectHint))
	http.HandleFunc("/api/project/chat", requireAuth(handleProjectChat))

	log.Printf("🚀  Coding Bootcamp → http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
