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
		ID           int    `json:"id"`
		Title        string `json:"title"`
		Summary      string `json:"summary"`
		Completed    bool   `json:"completed"`
		LessonCached bool   `json:"lessonCached"`
		// One flag per difficulty tier, e.g. {"beginner": true, "goat": false}.
		ChallengeCached map[string]bool `json:"challengeCached"`
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
			challenges := make(map[string]bool, len(difficultyOrder))
			for _, d := range difficultyOrder {
				challenges[d] = cacheHas(user, trackChallengeCacheKey(langID, t.ID, l.ID, d))
			}
			lessons[i] = LessonResp{
				ID:              l.ID,
				Title:           l.Title,
				Summary:         l.Summary,
				Completed:       done[fmt.Sprintf("track:%s:%d", t.ID, l.ID)],
				LessonCached:    cacheHas(user, fmt.Sprintf("%s:track:%s:%d", langID, t.ID, l.ID)),
				ChallengeCached: challenges,
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
		if cached, hit := cacheGet(user, cacheKey); hit {
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
		cacheStore(user, cacheKey, full)
	})
}

func handleTrackChallenge(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang       string `json:"lang"`
		TrackID    string `json:"track_id"`
		LessonID   int    `json:"lesson_id"`
		Difficulty string `json:"difficulty"` // beginner | intermediate | advanced | goat
		Force      bool   `json:"force"`      // true = bypass cache (Regenerate button)
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

	diff := normalizeDifficulty(req.Difficulty)
	spec := difficultySpec[diff]

	// Challenges are cached just like lessons (one entry per difficulty tier)
	// so they survive a server restart.
	cacheKey := trackChallengeCacheKey(req.Lang, req.TrackID, req.LessonID, diff)
	if !req.Force {
		if cached, hit := cacheGet(user, cacheKey); hit {
			streamCached(w, cached)
			return
		}
	}

	prevCtx := buildTrackContext(track, lesson.ID)

	prompt := fmt.Sprintf(`Create a coding challenge for Lesson %d: **%s** in the **%s** track (%s).

%s

This is the **%s** tier of four (Beginner → Intermediate → Advanced → GOAT).
Tier brief: %s

The challenge must:
- Focus specifically on **%s** concepts
- Build naturally on the previous lessons (reference and extend their patterns)

## Challenge: [Creative title]

**Difficulty**: %s

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
%s

## Hints
Exactly 3 progressive hints as a numbered list: a gentle nudge, a stronger
pointer, then a near-spoiler. This must be the LAST section — the UI keeps it
hidden until the student chooses to reveal it — so never reference the hints
from any other section.

Use the starter code above as a base, trimming or extending it to match the
tier brief.`,
		lesson.ID, lesson.Title, track.Title, lang.Name,
		prevCtx,
		spec.Label, spec.Brief,
		lesson.Title,
		spec.Label,
		lang.StarterTemplate,
	)

	streamFromAnthropic(r.Context(), w, lang.SystemPrompt, prompt, nil, func(full string) {
		cacheStore(user, cacheKey, full)
		// A fresh challenge starts with a clean hint record — hint use on the
		// old challenge shouldn't count against the new one.
		clearHintsUsed(user, cacheKey)
	})
}

func handleTrackEvaluate(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang       string `json:"lang"`
		TrackID    string `json:"track_id"`
		LessonID   int    `json:"lesson_id"`
		Difficulty string `json:"difficulty"` // tier being attempted
		Code       string `json:"code"`
		Challenge  string `json:"challenge"`
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

The challenge requirements are the specification. If the code does something because
the requirements explicitly ask for it — even if it goes against a general convention
or something the lesson teaches — do not flag it as an issue or style problem. A brief
"in real-world code you'd usually..." aside is fine, but the verdict and Issues section
must judge the code against the requirements as written.

Be encouraging. Note: code cannot be executed — evaluate on logic and conventions.`,
		lang.Name, track.Title, lesson.ID, lesson.Title,
		req.Challenge,
		lang.ID, req.Code,
		track.Title,
		lang.Name, lang.StyleNote,
	)
	prompt += lessonContextBlock(user, fmt.Sprintf("%s:track:%s:%d", req.Lang, req.TrackID, req.LessonID))
	prompt += hintUsageBlock(user, trackChallengeCacheKey(req.Lang, req.TrackID, req.LessonID, normalizeDifficulty(req.Difficulty)))

	streamFromAnthropic(r.Context(), w, lang.SystemPrompt, prompt, nil)
}

func handleTrackHint(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang       string `json:"lang"`
		TrackID    string `json:"track_id"`
		LessonID   int    `json:"lesson_id"`
		Difficulty string `json:"difficulty"` // tier being attempted
		Challenge  string `json:"challenge"`
		Code       string `json:"code"`
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
	prompt += lessonContextBlock(user, fmt.Sprintf("%s:track:%s:%d", req.Lang, req.TrackID, req.LessonID))

	// Record hint use only when the hint was actually delivered (onComplete
	// fires on a clean finish) — a failed request shouldn't count against the
	// student's no-hints run. The evaluation reads this flag.
	streamFromAnthropic(r.Context(), w, lang.SystemPrompt, prompt, nil, func(string) {
		markHintsUsed(user, trackChallengeCacheKey(req.Lang, req.TrackID, req.LessonID, normalizeDifficulty(req.Difficulty)))
	})
}

// handleTrackHintsViewed is the track equivalent of handleHintsViewed:
// revealing a track challenge's hidden Hints section counts as hint use.
func handleTrackHintsViewed(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang       string `json:"lang"`
		TrackID    string `json:"track_id"`
		LessonID   int    `json:"lesson_id"`
		Difficulty string `json:"difficulty"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}
	if _, _, ok := lookupTrackLesson(w, lang, req.TrackID, req.LessonID); !ok {
		return
	}
	markHintsUsed(user, trackChallengeCacheKey(req.Lang, req.TrackID, req.LessonID, normalizeDifficulty(req.Difficulty)))
	jsonOK(w, map[string]bool{"success": true})
}

func handleTrackChat(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang       string    `json:"lang"`
		TrackID    string    `json:"track_id"`
		LessonID   int       `json:"lesson_id"`
		Difficulty string    `json:"difficulty"` // tier the student is viewing
		Messages   []Message `json:"messages"`
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
		user,
		fmt.Sprintf("%s:track:%s:%d", req.Lang, req.TrackID, req.LessonID),
		trackChallengeCacheKey(req.Lang, req.TrackID, req.LessonID, normalizeDifficulty(req.Difficulty)),
	)
	system := fmt.Sprintf(`%s
The student is working through the **%s** track, Lesson %d: %s.
Answer their questions clearly and in the context of this specific lesson and track. When relevant, ground your answer in the lesson and challenge below.%s`,
		lang.SystemPrompt, track.Title, lesson.ID, lesson.Title, ctx,
	)

	streamFromAnthropic(r.Context(), w, system, "", req.Messages)
}
