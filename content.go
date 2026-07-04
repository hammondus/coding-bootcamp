package main

import (
	"fmt"
	"net/http"
	"strings"
)

// chatContextBlock pulls the already-generated lesson and challenge for a topic
// out of the cache and formats them as a reference block to append to the chat
// system prompt. Without this, chat only knows the topic *name* and can't speak
// to what the lesson actually taught or what the challenge asks. Either piece is
// simply omitted if it hasn't been generated yet (the student may open chat
// before loading the lesson or challenge), so this never blocks a question.
func chatContextBlock(user, lessonKey, challengeKey string) string {
	var sb strings.Builder
	if lesson, ok := cacheGet(user, lessonKey); ok {
		sb.WriteString("\n\n--- THE LESSON THE STUDENT IS STUDYING ---\n")
		sb.WriteString(lesson)
	}
	if challenge, ok := cacheGet(user, challengeKey); ok {
		sb.WriteString("\n\n--- THE CURRENT CHALLENGE ---\n")
		sb.WriteString(challenge)
	}
	return sb.String()
}

// lessonContextBlock returns the cached lesson for a topic as a reference block
// to append to a hint or evaluation prompt. Those prompts already include the
// challenge (sent from the UI) but never the lesson, so the model can't tie its
// feedback back to what was actually taught. Returns "" if the lesson hasn't
// been generated yet, so the prompt still works without it.
func lessonContextBlock(user, lessonKey string) string {
	if lesson, ok := cacheGet(user, lessonKey); ok {
		return "\n\n--- THE LESSON THIS CHALLENGE IS BASED ON ---\n" + lesson
	}
	return ""
}

// ── Language handler ──────────────────────────────────

func handleLanguages(w http.ResponseWriter, r *http.Request) {
	type LangMeta struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Icon        string `json:"icon"`
		Cmd         string `json:"cmd"`
		AccentColor string `json:"accentColor"`
		AccentDark  string `json:"accentDark"`
		AccentGlow  string `json:"accentGlow"`
		CodeLabel   string `json:"codeLabel"`
	}
	type CatMeta struct {
		ID        string     `json:"id"`
		Name      string     `json:"name"`
		Languages []LangMeta `json:"languages"`
	}
	result := make([]CatMeta, 0, len(categories))
	for _, c := range categories {
		cm := CatMeta{ID: c.ID, Name: c.Name, Languages: make([]LangMeta, 0, len(c.Langs))}
		for _, id := range c.Langs {
			l := languages[id]
			cm.Languages = append(cm.Languages, LangMeta{
				ID:          l.ID,
				Name:        l.Name,
				Icon:        l.Icon,
				Cmd:         l.Cmd,
				AccentColor: l.AccentColor,
				AccentDark:  l.AccentDark,
				AccentGlow:  l.AccentGlow,
				CodeLabel:   l.CodeLabel,
			})
		}
		result = append(result, cm)
	}
	jsonOK(w, result)
}

// ── Topics handler ────────────────────────────────────

func handleTopics(w http.ResponseWriter, r *http.Request, user string) {
	langID := r.URL.Query().Get("lang")
	lang, ok := languages[langID]
	if !ok {
		jsonErr(w, 400, "unknown language")
		return
	}

	done := getUserLangProgress(user, langID)

	type TopicResp struct {
		ID              int    `json:"id"`
		Name            string `json:"name"`
		Completed       bool   `json:"completed"`
		LessonCached    bool   `json:"lessonCached"`
		ChallengeCached bool   `json:"challengeCached"`
	}
	result := make([]TopicResp, len(lang.Topics))
	for i, t := range lang.Topics {
		result[i] = TopicResp{
			ID:              t.ID,
			Name:            t.Name,
			Completed:       done[fmt.Sprintf("%d", t.ID)],
			LessonCached:    cacheHas(user, fmt.Sprintf("%s:lesson:%d", langID, t.ID)),
			ChallengeCached: cacheHas(user, fmt.Sprintf("%s:challenge:%d", langID, t.ID)),
		}
	}
	jsonOK(w, result)
}

// ── Content handlers ──────────────────────────────────

func handleLesson(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang      string `json:"lang"`
		TopicID   int    `json:"topic_id"`
		TopicName string `json:"topic_name"`
		Force     bool   `json:"force"` // true = bypass cache (Regenerate button)
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}

	key := fmt.Sprintf("%s:lesson:%d", req.Lang, req.TopicID)

	// Serve from cache unless the user explicitly asked to regenerate.
	if !req.Force {
		if cached, hit := cacheGet(user, key); hit {
			streamCached(w, cached)
			return
		}
	}

	prompt := fmt.Sprintf(`Teach Topic %d of %d: **%s** in %s.

Use this structure:

## Overview
2–3 sentences introducing the concept and why it matters in %s.

## Key Concepts
5–7 bullet points, each with a brief explanation.

## Code Examples

### Example 1: [Descriptive title]
Show and explain a minimal, clear example.

### Example 2: [Descriptive title]
Show a more practical, realistic usage.

## Common Pitfalls
3 specific mistakes beginners make with this topic and how to avoid them.

## Summary
One crisp sentence: the essential takeaway.`,
		req.TopicID, len(lang.Topics), req.TopicName, lang.Name, lang.Name)

	streamFromAnthropic(r.Context(), w, lang.SystemPrompt, prompt, nil, func(full string) {
		cacheStore(user, key, full)
	})
}

func handleChallenge(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang      string `json:"lang"`
		TopicID   int    `json:"topic_id"`
		TopicName string `json:"topic_name"`
		Force     bool   `json:"force"` // true = bypass cache (Regenerate button)
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}

	// Challenges are cached just like lessons so they survive a server restart.
	key := fmt.Sprintf("%s:challenge:%d", req.Lang, req.TopicID)
	if !req.Force {
		if cached, hit := cacheGet(user, key); hit {
			streamCached(w, cached)
			return
		}
	}

	prompt := fmt.Sprintf(`Create a practical %s coding challenge for Topic %d: **%s**.

## Challenge: [Creative, specific title]

**Difficulty**: Beginner / Intermediate

**Task**
2–3 sentence description of what to implement. Be concrete and specific.

**Requirements**
- Specific, testable requirement
- Another requirement
- One more requirement

**Example**
`+"```"+`
Input:  ...
Output: ...
`+"```"+`

**Starter Code**
%s

Keep it focused purely on %s concepts. Make it engaging with a real-world flavor.`,
		lang.Name, req.TopicID, req.TopicName, lang.StarterTemplate, req.TopicName)

	streamFromAnthropic(r.Context(), w, lang.SystemPrompt, prompt, nil, func(full string) {
		cacheStore(user, key, full)
	})
}

func handleEvaluate(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang      string `json:"lang"`
		TopicID   int    `json:"topic_id"`
		TopicName string `json:"topic_name"`
		Code      string `json:"code"`
		Challenge string `json:"challenge"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}

	prompt := fmt.Sprintf(`Evaluate this %s code submission for Topic %d: **%s**.

**The Challenge**:
%s

**Submitted Code**:
`+"```%s\n%s\n```"+`

## Verdict
✅ **Pass** OR ❌ **Needs Work** — state it clearly on the first line.

## What Works Well
Specific praise for correct approaches and %s idioms used well.

## Issues Found
Specific bugs, logical errors, or style issues. Write "None — looks good!" if clean.

## %s Style Note
%s

## Suggested Improvement (if needed)
A corrected or more idiomatic version. Skip if the code passed cleanly.

The challenge requirements are the specification. If the code does something because
the requirements explicitly ask for it — even if it goes against a general convention
or something the lesson teaches — do not flag it as an issue or style problem. A brief
"in real-world code you'd usually..." aside is fine, but the verdict and Issues section
must judge the code against the requirements as written.

Be encouraging and educational. Note: code cannot be executed — evaluate on logic and conventions.`,
		lang.Name, req.TopicID, req.TopicName,
		req.Challenge,
		lang.ID, req.Code,
		lang.Name,
		lang.Name, lang.StyleNote)
	prompt += lessonContextBlock(user, fmt.Sprintf("%s:lesson:%d", req.Lang, req.TopicID))

	streamFromAnthropic(r.Context(), w, lang.SystemPrompt, prompt, nil)
}

func handleHint(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang      string `json:"lang"`
		TopicID   int    `json:"topic_id"`
		TopicName string `json:"topic_name"`
		Challenge string `json:"challenge"`
		Code      string `json:"code"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}

	prompt := fmt.Sprintf(`Give exactly ONE helpful hint for this %s challenge on **%s**.

Challenge:
%s

Student's current code:
`+"```%s\n%s\n```"+`

Give ONE specific, encouraging nudge that moves them forward without revealing the answer. Maximum 3 sentences.`,
		lang.Name, req.TopicName, req.Challenge, lang.ID, req.Code)
	prompt += lessonContextBlock(user, fmt.Sprintf("%s:lesson:%d", req.Lang, req.TopicID))

	streamFromAnthropic(r.Context(), w, lang.SystemPrompt, prompt, nil)
}

func handleChat(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang      string    `json:"lang"`
		TopicID   int       `json:"topic_id"`
		TopicName string    `json:"topic_name"`
		Messages  []Message `json:"messages"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}

	ctx := chatContextBlock(
		user,
		fmt.Sprintf("%s:lesson:%d", req.Lang, req.TopicID),
		fmt.Sprintf("%s:challenge:%d", req.Lang, req.TopicID),
	)
	system := fmt.Sprintf(`%s
The student is studying Topic %d: %s. Answer their questions clearly, concisely, and encouragingly. When relevant, ground your answer in the lesson and challenge below.%s`,
		lang.SystemPrompt, req.TopicID, req.TopicName, ctx)

	streamFromAnthropic(r.Context(), w, system, "", req.Messages)
}
