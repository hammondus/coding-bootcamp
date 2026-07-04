package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// The Getting Started guides: one hand-written markdown file per language in
// assets/setup/ (go.md, zig.md, ...). Unlike lessons, these are NOT generated —
// they're full of URLs and install steps that must stay exactly right, so they
// live as plain files a human maintains. The 💬 Ask tab still works on this
// page (see handleSetupChat), so a student can ask follow-up questions like
// "how do I install Go on Windows?" and get a live answer.

// readSetupGuide loads a language's guide from disk on every request — same
// spirit as the noCache static file serving: edit the .md, refresh, see it.
// The language ID is validated against the languages map by every caller
// before it reaches a file path.
func readSetupGuide(langID string) (string, error) {
	data, err := os.ReadFile(filepath.Join("assets", "setup", langID+".md"))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// handleSetup returns the guide as one JSON blob — no streaming, no cache,
// no token cost. GET /api/setup?lang=go
func handleSetup(w http.ResponseWriter, r *http.Request, user string) {
	langID := r.URL.Query().Get("lang")
	if _, ok := languages[langID]; !ok {
		jsonErr(w, 400, "unknown language")
		return
	}
	guide, err := readSetupGuide(langID)
	if err != nil {
		jsonErr(w, 500, "no setup guide for this language yet")
		return
	}
	jsonOK(w, map[string]string{"markdown": guide})
}

// handleSetupChat is the Ask tab for the Getting Started page. It works like
// handleChat, but the reference block is the static guide instead of a
// generated lesson, and the scope is environment setup rather than a topic.
func handleSetupChat(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang     string    `json:"lang"`
		Messages []Message `json:"messages"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}

	// A missing guide file just means chat runs without the reference block —
	// setup questions are still answerable from the system prompt alone.
	guideBlock := ""
	if guide, err := readSetupGuide(req.Lang); err == nil {
		guideBlock = "\n\n--- THE GETTING STARTED GUIDE THE STUDENT IS READING ---\n" + guide
	}

	system := fmt.Sprintf(`%s
The student is on the Getting Started page — they may not have written any %s
yet. Their questions here are about getting set up: online playgrounds,
installing the tools, editor setup (VS Code, Zed), running a first program,
and where to find documentation. Give concrete, step-by-step answers. Setup
steps often differ by operating system — if the student hasn't said which one
they're on and it matters, ask. Recommend one good option rather than listing
every alternative; beginners want a path, not a survey.%s`,
		lang.SystemPrompt, lang.Name, guideBlock)

	// Save the conversation so it survives a reload, keyed separately from any
	// topic's chat (see setupChatStoreKey in workspace.go).
	streamFromAnthropic(r.Context(), w, system, "", req.Messages, func(full string) {
		storeChat(user, setupChatStoreKey(req.Lang),
			append(req.Messages, Message{Role: "assistant", Content: full}))
	})
}

// handleSetupWorkspace restores the saved Getting Started chat. There is no
// solution or feedback here — the page has no challenge — so only the chat
// field of the shared response shape is ever populated.
func handleSetupWorkspace(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang string `json:"lang"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	if _, ok := lookupLang(w, req.Lang); !ok {
		return
	}
	_, chat := getWorkspace(user, "", setupChatStoreKey(req.Lang))
	jsonOK(w, workspaceResp{Chat: chat})
}
