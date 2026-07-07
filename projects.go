package main

// ── Go projects ───────────────────────────────────────
// Projects are capstones: the student builds a complete application from
// scratch across several milestones. To add a Go project, append a Project{}
// block here. To add projects for another language, create a new var (e.g.
// zigProjects) and assign it in that Language's definition in languages.go.
//
// Milestone summaries feed the progressive context for later milestones (the
// same pattern tracks use), so keep them to a concrete one-liner describing
// what that step delivers.

var goProjects = []Project{
	{
		ID:          "url-shortener",
		Title:       "URL Shortener",
		Icon:        "🔗",
		Description: "Build a complete URL shortening web service from scratch — an HTTP API with an in-memory store and file persistence.",
		Goal:        "a URL shortening web service: POST a long URL, get back a short code; visiting the short code redirects to the original URL. State persists across restarts.",
		Prereqs:     "all Go fundamentals, HTTP track lessons 5-6 (servers and routing), and Standard Library lessons 4-5 (encoding/json, sync)",
		Milestones: []ProjectMilestone{
			{1, "HTTP Server Skeleton", "A runnable net/http server with a health-check route and graceful structure to build on"},
			{2, "In-Memory Store", "A concurrency-safe store mapping short codes to long URLs, guarded by a mutex"},
			{3, "Shorten Endpoint", "A POST /shorten endpoint that accepts a URL, generates a unique short code, and returns it as JSON"},
			{4, "Redirect Endpoint", "A GET /{code} handler that looks up the code and issues an HTTP redirect to the original URL"},
			{5, "File Persistence", "Loading the store from a JSON file at startup and saving atomically on each write so links survive a restart"},
		},
	},
	{
		ID:          "markdown-to-html",
		Title:       "Markdown → HTML",
		Icon:        "📝",
		Description: "Build a Markdown-to-HTML converter from scratch — an incremental parser that turns a .md file into a styled HTML page. A natural stepping stone toward a static site generator.",
		Goal:        "a command-line tool that reads a Markdown file and writes an HTML file: headings, paragraphs, inline formatting, links, lists, and code blocks all converted correctly, using only the Go standard library.",
		Prereqs:     "all Go fundamentals and Standard Library lessons 1-2 (strings/strconv, io/bufio/os)",
		Milestones: []ProjectMilestone{
			{1, "CLI & File I/O", "A command-line program that takes a .md path, reads it, and writes a .html file alongside it"},
			{2, "Headings & Paragraphs", "Line-by-line block parsing that converts # headings and blank-line-separated paragraphs to HTML, with proper escaping"},
			{3, "Inline Formatting", "Converting **bold**, *italic*, `inline code`, and [links](url) inside text spans"},
			{4, "Lists & Code Blocks", "A block state machine handling ordered/unordered lists and fenced ``` code blocks"},
			{5, "Document Assembly", "Wrapping the output in a complete, styled HTML5 document — the foundation a static site generator would build on"},
		},
	},
	{
		ID:          "chat-server",
		Title:       "WebSocket Chat Server",
		Icon:        "💬",
		Description: "Build a real-time multi-user chat server from scratch — implementing the WebSocket protocol by hand over the standard library, with a concurrency-safe broadcast hub and a vanilla JS frontend.",
		Goal:        "a real-time chat server where multiple browser clients connect over WebSockets and see each other's messages instantly. The WebSocket handshake and framing are implemented from scratch using only the Go standard library (net/http hijacking, crypto/sha1, encoding/base64); the frontend is vanilla HTML/CSS/JS using the browser's native WebSocket API, written to work against this backend.",
		Prereqs:     "all Go fundamentals (especially Goroutines & Channels), HTTP track lessons 5-7, and Standard Library lesson 5 (sync); the SHA-1 + base64 handshake is taught inside milestone 2 itself",
		Milestones: []ProjectMilestone{
			{1, "HTTP Server & Frontend", "An net/http server that serves a vanilla HTML/CSS/JS chat page using the browser's native WebSocket API"},
			{2, "WebSocket Handshake", "Upgrading a GET request by hand: validating headers, computing Sec-WebSocket-Accept (SHA-1 + base64), and hijacking the connection"},
			{3, "Frame Decoding & Encoding", "Reading masked client frames and writing server frames per RFC 6455 — text messages and the close frame"},
			{4, "The Broadcast Hub", "A concurrency-safe hub of connected clients using goroutines and channels: register, unregister, and broadcast over a select loop"},
			{5, "Usernames, Join/Leave & Cleanup", "Names on join, broadcast join/leave notices, and graceful disconnect cleanup with no goroutine leaks"},
		},
	},
	{
		// The grand finale: the student rebuilds this very application. It is
		// deliberately full-stack — each milestone is a vertical slice, so the
		// Go backend and the vanilla JS frontend grow together rather than one
		// being finished before the other starts.
		ID:          "build-this-bootcamp",
		Title:       "Build This Bootcamp",
		Icon:        "🎓",
		Description: "The grand finale: rebuild this very bootcamp from scratch — a Go standard-library backend streaming AI-generated lessons over SSE to a vanilla HTML/CSS/JS frontend, with auth, sessions, per-user caching, progress tracking, and atomic JSON persistence. Built in vertical slices, backend and frontend together.",
		Goal:        "a complete AI-powered bootcamp web app — the very one you are using right now: a Go standard-library backend with cookie-session auth, mutex-guarded in-memory state, atomic JSON-file persistence, and a per-user lesson cache, streaming AI-generated lessons, challenge evaluations, hints, and chat from the Claude API over SSE to a vanilla HTML/CSS/JS single-page frontend with no build step. Every milestone is a vertical slice: the backend feature and the frontend that exercises it land together, so the app is runnable end-to-end at every step.",
	Prereqs:     "everything: all Go fundamentals plus the HTTP, Standard Library, and Testing tracks (JS/HTML/CSS fundamentals help for the frontend slices)",
		Milestones: []ProjectMilestone{
			{1, "Server Skeleton & App Shell", "A Go net/http server serving a static/ directory (port from an env var), plus the HTML/CSS shell of the single-page app: header, sidebar, and main content panel"},
			{2, "JSON API & Dynamic Boot", "jsonOK/jsonErr/decodePOST helpers, a /api/languages endpoint fed from an in-code config map, and app.js fetching it on load to render the language picker"},
			{3, "User Accounts", "An RWMutex-guarded in-memory user store with salted password hashing, a register endpoint, and the registration form wired to it"},
			{4, "Login, Sessions & Cookies", "Random session tokens in their own mutex-guarded map, set as an HttpOnly cookie on login, with the login screen, logout flow, and session-aware boot in the UI"},
			{5, "The requireAuth Middleware", "A handler wrapper that resolves the session cookie to a username and passes it to every protected endpoint, with the frontend returning to the login screen on a 401"},
			{6, "Atomic JSON Persistence", "A writeFileAtomic helper (marshal under a save mutex, write a .tmp file, rename into place) plus loaders that decode into a temp variable so a corrupt file can't half-populate live state — users and sessions now survive a restart"},
			{7, "Streaming from Claude", "A backend proxy that calls the Claude API with stream:true, parses the upstream SSE events, and forwards text deltas to the browser as its own SSE stream, cancelling the upstream call when the client disconnects"},
			{8, "Rendering the Stream", "Frontend code that reads the SSE response with fetch and renders markdown incrementally as deltas arrive, with loading and error states"},
			{9, "The Lesson Curriculum", "A per-language topic list endpoint and prompt builders that generate complete lessons, plus the sidebar topic navigation and lesson view that display them"},
			{10, "Per-User Lesson Cache", "A username → key → markdown cache guarded by its own RWMutex, checked before calling the API, persisted with a background save, and replayed instantly to the browser on a hit"},
			{11, "Progress Tracking", "A per-user progress store with a toggle-complete endpoint, snapshot-copied under RLock when building responses, with completion ticks appearing in the sidebar"},
			{12, "Challenges & Evaluation", "A challenge generator, a code submission editor in the UI, and an evaluate endpoint that streams back a structured pass/needs-work review"},
			{13, "Hints & Contextual Chat", "A single-nudge hint endpoint and a chat panel whose system prompt is grounded in the current lesson and the student's code, so answers stay on topic"},
			{14, "Hardening the Stream", "Dial/response timeouts and an overall per-call deadline, bounded backoff retries on 429/529/5xx before the first byte is streamed, forwarding upstream error events to the browser, and caching only after a clean message_stop"},
			{15, "Dev Mode & the Race Detector", "A -dev flag that bypasses login for local testing, then a full go vet and go run -race pass over every feature to prove the concurrency story holds"},
		},
	},
}

// ── JavaScript projects ───────────────────────────────

var javascriptProjects = []Project{
	{
		// A JavaScript-first game project that grows a Go backend. The ghost
		// design is deliberate: players never affect each other, so each
		// client owns its own simulation — no prediction or reconciliation,
		// just interpolation of other players' positions. Milestones 1–5 are
		// pure canvas game; 6–8 add the server. Recorded ghosts (7) land
		// before live ghosts (8) so ghost rendering and networking are never
		// debugged at the same time.
		ID:          "moon-patrol",
		Title:       "Moon Patrol Ghosts",
		Icon:        "🌙",
		Description: "Build a canvas remake of the 1982 arcade classic Moon Patrol — an auto-scrolling moon buggy that jumps craters and shoots UFOs, over parallax-scrolling mountains — then add ghost multiplayer: a small Go backend that keeps a leaderboard, replays recorded runs, and relays live players as translucent ghost buggies over WebSockets.",
		Goal:        "a browser remake of the arcade classic Moon Patrol with ghost multiplayer: a vanilla JavaScript canvas game (no build step, no libraries) where an auto-scrolling moon buggy jumps craters and shoots UFOs with a single fire key that shoots forward and upward at once, across a course with parallax-scrolling background layers — backed by a small Go standard-library server that serves the shared course data, keeps a persistent leaderboard, stores finished runs, and relays live positions over WebSockets so other players appear as translucent, non-interacting ghost buggies. Ghosts never affect your run, so each client owns its own simulation; the only networking trick needed is interpolating ghost positions between updates.",
		Prereqs:     "all JavaScript fundamentals and all four JavaScript tracks (Canvas & Animation, Game Development, Browser APIs, Real-Time & Networking); milestones 6-8 additionally assume Go fundamentals and the Go HTTP track, because the ghost backend is written in Go",
		Milestones: []ProjectMilestone{
			{1, "Canvas, Game Loop & Buggy", "A canvas page with a fixed-timestep requestAnimationFrame loop and the buggy driving on flat lunar ground with keyboard speed control"},
			{2, "Scrolling Terrain & Parallax", "An auto-scrolling course built from authored segment data — craters and rocks on the surface, mountain and hill layers scrolling at different speeds behind it"},
			{3, "Jumping, Collision & Checkpoints", "A gravity-based jump arc, collision with craters and rocks, and death/respawn at lettered checkpoints"},
			{4, "UFOs & Dual Fire", "One fire key shooting forward and upward at once, authored UFO attack waves, and UFO bombs that blast new craters into the terrain ahead"},
			{5, "HUD, Scoring & Game Over", "Score, lives, checkpoint progress and run time in a HUD, with a game-over and restart loop — the solo game is complete and playable"},
			{6, "Go Backend: Course & Leaderboard", "A small Go standard-library server that serves the static files and the course as JSON, and keeps a leaderboard persisted atomically to a JSON file"},
			{7, "Recorded Ghosts", "Runs recorded as timed position samples and posted to the server on finish, then replayed as translucent ghost buggies racing alongside the player"},
			{8, "Live Ghosts over WebSockets", "A Go broadcast hub relaying each connected player's position a few times a second, rendered as live ghosts with positions interpolated between updates"},
		},
	},
}

// ── Claude projects ───────────────────────────────────

var claudeProjects = []Project{
	{
		// Deliverables here are configuration and prompts, not program code —
		// each milestone is still submitted and evaluated as text, and the
		// student proves it works by running it against their own repository.
		ID:          "supercharged-repo",
		Title:       "Supercharge a Repository",
		Icon:        "⚡",
		Description: "Take a real repository from zero Claude setup to a complete, team-ready AI workflow: project memory, permissions, custom commands, guardrail hooks, a reviewer subagent, and a headless CI check.",
		Goal:        "a repository where Claude Code is a productive, safe team tool out of the box: a CLAUDE.md that changes behavior, permissions that stop the prompt-fatigue, commands for the repo's routine tasks, hooks that enforce its rules, a code-review subagent, and a CI job that runs Claude headlessly — all checked in, so a new teammate gets the whole setup on clone.",
		Prereqs:     "the Claude fundamentals course plus the Agentic Workflows and Extending Claude Code tracks; you also need a real repository of your own to apply each milestone to",
		Milestones: []ProjectMilestone{
			{1, "Audit & CLAUDE.md", "An audit of the repo's conventions, commands, and gotchas, distilled into a CLAUDE.md that measurably changes how Claude works in it"},
			{2, "Permissions & Settings", "A checked-in settings file with allow/deny rules tuned to the repo, so routine work runs without prompts and dangerous operations still ask"},
			{3, "Custom Commands", "Slash commands for the repo's three most repeated workflows (e.g. run-and-fix tests, draft a changelog entry, scaffold a component), each taking arguments"},
			{4, "Guardrail Hooks", "Hooks that enforce rules prompts can't guarantee: auto-format after edits, block edits to protected paths, and require a passing build before any commit"},
			{5, "Reviewer Subagent", "A code-review agent with a written rubric drawn from the repo's real conventions, restricted to read-only tools, invocable on demand"},
			{6, "Headless CI Check", "A CI job that runs the reviewer headlessly against each change and posts its findings as a report, with permissions locked down for unattended use"},
		},
	},
	{
		// The inverse of using Claude Code: building one. Milestones order the
		// risk — read-only tools come before file edits and command execution,
		// so the student adds the confirmation step before the agent can do
		// damage.
		ID:          "build-an-agent",
		Title:       "Build Your Own Agent",
		Icon:        "🦾",
		Description: "Build a miniature Claude Code from scratch in Go: a terminal agent that streams responses, holds a conversation, and uses tools to read files, edit them, and run commands — with a human confirmation step where it counts.",
		Goal:        "a command-line coding agent in Go, using only the standard library: it talks to the Messages API, streams replies to the terminal, keeps multi-turn context, and works in an agent loop with tools to list and read files, edit them, and run shell commands — destructive actions gated behind a y/N confirmation, with timeouts, retries, and clean cancellation throughout.",
		Prereqs:     "all Go fundamentals, the Go HTTP track lessons 1-4, and The Claude API in Go track (the agent loop and tool use are taught there; this project turns them into a real program)",
		Milestones: []ProjectMilestone{
			{1, "CLI & First Message", "A Go program that reads a prompt from its arguments, calls the Messages API with an API key from the environment, and prints the reply"},
			{2, "Streaming to the Terminal", "Replacing the blocking call with stream: true — parsing SSE events and printing text deltas as they arrive, finishing cleanly on message_stop"},
			{3, "The Conversation Loop", "A read-eval-print loop holding the messages array as state, so follow-up questions build on earlier turns, with Ctrl-C cancelling a stream without killing the session"},
			{4, "Read-Only Tools", "The agent loop: list_files and read_file tools defined with JSON schemas, executing tool_use requests and returning results until the model answers in text"},
			{5, "Edits & Commands, Confirmed", "edit_file and run_command tools gated behind a y/N terminal confirmation showing exactly what will change or run — the agent can now do real work, safely"},
			{6, "Hardening", "Per-call timeouts, bounded retry with backoff on 429/529/5xx, a token budget that stops runaway loops, and a transcript log of every tool call for review"},
		},
	},
}

// ── Git & CI/CD projects ──────────────────────────────

var gitProjects = []Project{
	{
		// Like the Claude capstone, deliverables here are not program code:
		// each milestone is submitted as the commands run, the YAML checked
		// in, and a short written account of what happened — and the student
		// proves it works against their own real repository on GitHub.
		ID:          "ship-it",
		Title:       "Ship It: A Complete Delivery Pipeline",
		Icon:        "🚢",
		Description: "Take a project of your own from a lone local folder to a GitHub repository with protected branches, a PR-based workflow, full CI, and automated versioned releases — the delivery setup a senior engineer would build.",
		Goal:        "a real repository on GitHub where nothing reaches main except through a reviewed, green-CI pull request, and every release is a tagged, automated, documented build — with the whole setup checked in and written up so a new teammate could operate it from day one",
		Prereqs:     "the entire Git & CI/CD fundamentals course; the GitHub Actions in Depth and Release Engineering tracks deepen several milestones but the fundamentals are enough to complete them all; you also need a small project of your own (any language, even a static site) to apply each milestone to",
		Milestones: []ProjectMilestone{
			{1, "Repository Foundations", "The project published to GitHub as a clean repository: a sensible .gitignore, a README that says what it is and how to run it, and a tidy commit history with well-written messages"},
			{2, "Branch Protection & PR Flow", "main protected so changes arrive only by pull request with a review and passing checks, a PR template that prompts for context, and a first feature merged through the full flow"},
			{3, "Continuous Integration", "A workflow that builds and tests the project on every pull request, made a required status check — from here on, a red build blocks the merge"},
			{4, "Faster, Broader CI", "The pipeline hardened: a matrix across versions or platforms where it makes sense, dependency caching, path filters so docs-only changes skip the build, and a status badge in the README"},
			{5, "Automated Releases", "A tag-triggered release pipeline: pushing v1.2.3 builds the artifacts, generates release notes from the history, and publishes a GitHub Release with no manual steps"},
			{6, "Deploy & Hand Over", "A deploy job gated behind a GitHub Environment with manual approval (a real deploy or a documented simulation), plus a CONTRIBUTING guide covering branching, reviews, releases, and rollback — the handover document for the next teammate"},
		},
	},
}
