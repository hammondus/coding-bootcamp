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
		Milestones: []ProjectMilestone{
			{1, "HTTP Server & Frontend", "An net/http server that serves a vanilla HTML/CSS/JS chat page using the browser's native WebSocket API"},
			{2, "WebSocket Handshake", "Upgrading a GET request by hand: validating headers, computing Sec-WebSocket-Accept (SHA-1 + base64), and hijacking the connection"},
			{3, "Frame Decoding & Encoding", "Reading masked client frames and writing server frames per RFC 6455 — text messages and the close frame"},
			{4, "The Broadcast Hub", "A concurrency-safe hub of connected clients using goroutines and channels: register, unregister, and broadcast over a select loop"},
			{5, "Usernames, Join/Leave & Cleanup", "Names on join, broadcast join/leave notices, and graceful disconnect cleanup with no goroutine leaks"},
		},
	},
}
