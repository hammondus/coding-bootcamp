package main

import (
	"fmt"
	"net/http"
	"strings"
)

// ── Track helpers ─────────────────────────────────────

// buildTrackContext returns a markdown block summarising lessons already
// covered, so each lesson can build progressively on the previous ones.
func buildTrackContext(track Track, upToLessonID int) string {
	if upToLessonID <= 1 {
		return "This is the first lesson in this track — no prior context needed."
	}
	var sb strings.Builder
	sb.WriteString("**What previous lessons in this track covered:**\n")
	for _, l := range track.Lessons {
		if l.ID >= upToLessonID {
			break
		}
		fmt.Fprintf(&sb, "- Lesson %d — **%s**: %s\n", l.ID, l.Title, l.Summary)
	}
	return sb.String()
}

// findTrackLesson looks up a track and lesson by ID within a language.
func findTrackLesson(lang Language, trackID string, lessonID int) (Track, TrackLesson, bool) {
	for _, t := range lang.Tracks {
		if t.ID == trackID {
			for _, l := range t.Lessons {
				if l.ID == lessonID {
					return t, l, true
				}
			}
		}
	}
	return Track{}, TrackLesson{}, false
}

// ── Track handlers ────────────────────────────────────

func handleTracks(w http.ResponseWriter, r *http.Request, user string) {
	langID := r.URL.Query().Get("lang")
	lang, ok := languages[langID]
	if !ok {
		jsonErr(w, 400, "unknown language")
		return
	}

	// Snapshot progress under the lock (via getUserLangProgress) rather than
	// reading the shared map directly — reading it here while handleProgress
	// writes would be a data race and can panic the process.
	done := getUserLangProgress(user, langID)

	type LessonResp struct {
		ID              int    `json:"id"`
		Title           string `json:"title"`
		Summary         string `json:"summary"`
		Completed       bool   `json:"completed"`
		LessonCached    bool   `json:"lessonCached"`
		ChallengeCached bool   `json:"challengeCached"`
	}
	type TrackResp struct {
		ID          string       `json:"id"`
		Title       string       `json:"title"`
		Icon        string       `json:"icon"`
		Description string       `json:"description"`
		Lessons     []LessonResp `json:"lessons"`
	}

	result := make([]TrackResp, 0, len(lang.Tracks))
	for _, t := range lang.Tracks {
		lessons := make([]LessonResp, len(t.Lessons))
		for i, l := range t.Lessons {
			lessons[i] = LessonResp{
				ID:              l.ID,
				Title:           l.Title,
				Summary:         l.Summary,
				Completed:       done[fmt.Sprintf("track:%s:%d", t.ID, l.ID)],
				LessonCached:    cacheHas(fmt.Sprintf("%s:track:%s:%d", langID, t.ID, l.ID)),
				ChallengeCached: cacheHas(fmt.Sprintf("%s:track:%s:challenge:%d", langID, t.ID, l.ID)),
			}
		}
		result = append(result, TrackResp{
			ID:          t.ID,
			Title:       t.Title,
			Icon:        t.Icon,
			Description: t.Description,
			Lessons:     lessons,
		})
	}
	jsonOK(w, result)
}

func handleTrackLesson(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang     string `json:"lang"`
		TrackID  string `json:"track_id"`
		LessonID int    `json:"lesson_id"`
		Force    bool   `json:"force"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}
	track, lesson, ok := lookupTrackLesson(w, lang, req.TrackID, req.LessonID)
	if !ok {
		return
	}

	cacheKey := fmt.Sprintf("%s:track:%s:%d", req.Lang, req.TrackID, req.LessonID)
	if !req.Force {
		if cached, hit := cacheGet(cacheKey); hit {
			streamCached(w, cached)
			return
		}
	}

	prevCtx := buildTrackContext(track, lesson.ID)
	continuity := ""
	if lesson.ID > 1 {
		continuity = "\n### Building on previous lessons\nShow explicitly how this lesson's code extends or works alongside what was built before."
	}

	prompt := fmt.Sprintf(`You are teaching Lesson %d of %d in the **%s** track for %s.

**Track goal:** %s

%s

---

## Your task: teach **%s**

Structure your lesson as follows:

## Overview
What this lesson introduces and how it builds on what came before (2–3 sentences).

## Key Concepts
5–7 new concepts, each with a brief explanation.

## Code Examples

### Example 1: [Descriptive title]
A clear, focused example of the core new concept.

### Example 2: [Descriptive title]
A more complete or practical usage, building directly on Example 1.
%s

## Common Mistakes
2–3 mistakes specific to this lesson's topic.

## Summary
One sentence: what the student can now do that they couldn't before this lesson.`,
		lesson.ID, len(track.Lessons), track.Title, lang.Name,
		track.Description,
		prevCtx,
		lesson.Title,
		continuity,
	)

	streamFromAnthropic(r.Context(), w, lang.SystemPrompt, prompt, nil, func(full string) {
		cacheStore(cacheKey, full)
	})
}

func handleTrackChallenge(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang     string `json:"lang"`
		TrackID  string `json:"track_id"`
		LessonID int    `json:"lesson_id"`
		Force    bool   `json:"force"` // true = bypass cache (Regenerate button)
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}
	track, lesson, ok := lookupTrackLesson(w, lang, req.TrackID, req.LessonID)
	if !ok {
		return
	}

	// Challenges are cached just like lessons so they survive a server restart.
	cacheKey := fmt.Sprintf("%s:track:%s:challenge:%d", req.Lang, req.TrackID, req.LessonID)
	if !req.Force {
		if cached, hit := cacheGet(cacheKey); hit {
			streamCached(w, cached)
			return
		}
	}

	prevCtx := buildTrackContext(track, lesson.ID)

	prompt := fmt.Sprintf(`Create a coding challenge for Lesson %d: **%s** in the **%s** track (%s).

%s

The challenge must:
- Focus specifically on **%s** concepts
- Build naturally on the previous lessons (reference and extend their patterns)
- Be completable in ~15–20 minutes

## Challenge: [Creative title]

**Difficulty**: Beginner / Intermediate

**Task**
2–3 sentences describing what to build. Be concrete and specific.

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
%s`,
		lesson.ID, lesson.Title, track.Title, lang.Name,
		prevCtx,
		lesson.Title,
		lang.StarterTemplate,
	)

	streamFromAnthropic(r.Context(), w, lang.SystemPrompt, prompt, nil, func(full string) {
		cacheStore(cacheKey, full)
	})
}

func handleTrackEvaluate(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang      string `json:"lang"`
		TrackID   string `json:"track_id"`
		LessonID  int    `json:"lesson_id"`
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
	track, lesson, ok := lookupTrackLesson(w, lang, req.TrackID, req.LessonID)
	if !ok {
		return
	}

	prompt := fmt.Sprintf(`Evaluate this %s code submission for the **%s** track, Lesson %d: **%s**.

**The Challenge:**
%s

**Submitted Code:**
`+"```%s\n%s\n```"+`

## Verdict
✅ **Pass** OR ❌ **Needs Work** — state it clearly on the first line.

## What Works Well
Specific praise, especially for using patterns from the %s track correctly.

## Issues Found
Specific bugs, logical errors, or style issues. Write "None — looks good!" if clean.

## %s Style Note
%s

## Suggested Improvement (if needed)
A corrected or more idiomatic version. Skip if the code passed cleanly.

Be encouraging. Note: code cannot be executed — evaluate on logic and conventions.`,
		lang.Name, track.Title, lesson.ID, lesson.Title,
		req.Challenge,
		lang.ID, req.Code,
		track.Title,
		lang.Name, lang.StyleNote,
	)
	prompt += lessonContextBlock(fmt.Sprintf("%s:track:%s:%d", req.Lang, req.TrackID, req.LessonID))

	streamFromAnthropic(r.Context(), w, lang.SystemPrompt, prompt, nil)
}

func handleTrackHint(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang      string `json:"lang"`
		TrackID   string `json:"track_id"`
		LessonID  int    `json:"lesson_id"`
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
	track, lesson, ok := lookupTrackLesson(w, lang, req.TrackID, req.LessonID)
	if !ok {
		return
	}

	prompt := fmt.Sprintf(`Give ONE helpful hint for this %s challenge on **%s** (track: %s).

Challenge:
%s

Student's current code:
`+"```%s\n%s\n```"+`

Give ONE specific, encouraging nudge. Maximum 3 sentences. Don't reveal the answer.`,
		lang.Name, lesson.Title, track.Title,
		req.Challenge, lang.ID, req.Code,
	)
	prompt += lessonContextBlock(fmt.Sprintf("%s:track:%s:%d", req.Lang, req.TrackID, req.LessonID))

	streamFromAnthropic(r.Context(), w, lang.SystemPrompt, prompt, nil)
}

func handleTrackChat(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang     string    `json:"lang"`
		TrackID  string    `json:"track_id"`
		LessonID int       `json:"lesson_id"`
		Messages []Message `json:"messages"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}
	track, lesson, ok := lookupTrackLesson(w, lang, req.TrackID, req.LessonID)
	if !ok {
		return
	}

	ctx := chatContextBlock(
		fmt.Sprintf("%s:track:%s:%d", req.Lang, req.TrackID, req.LessonID),
		fmt.Sprintf("%s:track:%s:challenge:%d", req.Lang, req.TrackID, req.LessonID),
	)
	system := fmt.Sprintf(`%s
The student is working through the **%s** track, Lesson %d: %s.
Answer their questions clearly and in the context of this specific lesson and track. When relevant, ground your answer in the lesson and challenge below.%s`,
		lang.SystemPrompt, track.Title, lesson.ID, lesson.Title, ctx,
	)

	streamFromAnthropic(r.Context(), w, system, "", req.Messages)
}
