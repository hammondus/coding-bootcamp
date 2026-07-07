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
// to append to a challenge, hint, or evaluation prompt, so the model can tie
// what it generates back to what the lesson actually taught. Returns "" if the
// lesson hasn't been generated yet, so hint/evaluation prompts still work
// without it (challenge generation refuses to run at all in that case — see
// handleChallenge).
func lessonContextBlock(user, lessonKey string) string {
	if lesson, ok := cacheGet(user, lessonKey); ok {
		return "\n\n--- THE LESSON THIS CHALLENGE IS BASED ON ---\n" + lesson
	}
	return ""
}

// hintUsageBlock returns evaluation-prompt context about whether the student
// used any hints on this challenge (💡 Hint button or revealing the hidden
// Hints section). A no-hints pass earns explicit praise; hint use is never
// mentioned, so there's no shame in asking.
func hintUsageBlock(user, challengeKey string) string {
	if hintsWereUsed(user, challengeKey) {
		return "\n\nContext: the student used one or more hints on this challenge. " +
			"That's a healthy way to learn — do not mention or penalize hint use in your feedback."
	}
	return "\n\nContext: the student has NOT used any hints on this challenge. " +
		"If the verdict is Pass, celebrate that explicitly in **What Works Well** — " +
		"e.g. \"🏆 Solved without a single hint — seriously impressive!\". " +
		"If the verdict is Needs Work, do not mention hints at all."
}

// ── Challenge difficulty tiers ────────────────────────
//
// Every lesson offers its challenge at four difficulty tiers. Each tier is
// generated and cached separately, so a student can warm up on Beginner and
// work up to GOAT on the same topic without losing the easier versions.

// difficultyOrder is the canonical tier order shown in the UI.
var difficultyOrder = []string{"beginner", "intermediate", "advanced", "goat"}

// difficultySpec maps a tier to its display label and the brief given to the
// model when generating a challenge at that tier.
var difficultySpec = map[string]struct {
	Label string
	Brief string
}{
	"beginner": {
		Label: "Beginner",
		Brief: "A gentle warm-up. One small, clearly scoped task using only the most basic form of the topic. Generous starter code. Completable in ~10 minutes.",
	},
	"intermediate": {
		Label: "Intermediate",
		Brief: "A practical task that combines the topic with everyday programming (conditionals, loops, simple data structures). Some starter code, but the student writes all the interesting parts. Completable in ~20 minutes.",
	},
	"advanced": {
		Label: "Advanced",
		Brief: "A demanding challenge that stretches the topic: edge cases, several interacting requirements, minimal hand-holding. Starter code is a bare skeleton. Completable in ~30–40 minutes.",
	},
	"goat": {
		Label: "GOAT 🐐",
		Brief: "Greatest Of All Time: a brutally hard challenge for bragging rights. Push the topic to its limits — tricky constraints, nasty edge cases, a correctness or performance twist — and give no starter code beyond an empty function or file skeleton. May take an hour or more. Get the difficulty from depth on covered material; the sequence rule states what may be assumed and how any reach beyond it must be flagged.",
	},
}

// normalizeDifficulty maps a client-supplied difficulty to a known tier,
// defaulting to beginner so older clients (which send no difficulty) work.
func normalizeDifficulty(d string) string {
	if _, ok := difficultySpec[d]; ok {
		return d
	}
	return "beginner"
}

// challengeCacheKey names a fundamentals challenge at a difficulty tier.
// Beginner keeps the pre-tier key ("go:challenge:1") so challenges generated
// before tiers existed still show up as the Beginner tier instead of being
// orphaned in the cache file.
func challengeCacheKey(langID string, topicID int, diff string) string {
	if diff == "beginner" {
		return fmt.Sprintf("%s:challenge:%d", langID, topicID)
	}
	return fmt.Sprintf("%s:challenge:%d:%s", langID, topicID, diff)
}

// trackChallengeCacheKey is the advanced-track equivalent of
// challengeCacheKey, with the same legacy-beginner rule.
func trackChallengeCacheKey(langID, trackID string, lessonID int, diff string) string {
	if diff == "beginner" {
		return fmt.Sprintf("%s:track:%s:challenge:%d", langID, trackID, lessonID)
	}
	return fmt.Sprintf("%s:track:%s:challenge:%d:%s", langID, trackID, lessonID, diff)
}

// ── The teaching sequence ─────────────────────────────
//
// Topic order is a strict sequence: a lesson may assume earlier topics, and a
// challenge may only require skills its topic's lesson (or an earlier one)
// taught. These helpers render that sequence into prompts — they are the
// mechanism that stops a topic-1 challenge demanding fmt.Printf when no
// lesson has covered printing.

// topicSkillsBlock lists topics 1..uptoID with the skills each lesson covers,
// as a markdown block for prompts. Returns "" when nothing is in range (e.g.
// uptoID 0 for "before the first topic").
func topicSkillsBlock(lang Language, uptoID int) string {
	var sb strings.Builder
	for _, t := range lang.Topics {
		if t.ID > uptoID {
			break
		}
		fmt.Fprintf(&sb, "- Topic %d — **%s**: %s\n", t.ID, t.Name, t.Skills)
	}
	return sb.String()
}

// sequenceRule is the covered-skills rule injected into challenge prompts.
// Beginner→Advanced are strictly bound to what has been taught; GOAT may go
// beyond the syllabus but must say so up front, so the student knows the
// extra skills are homework, not something they missed.
func sequenceRule(diff string) string {
	if diff == "goat" {
		return `THE SEQUENCE RULE: lean on the covered skills listed above. You MAY reach
beyond them for the twist that makes this tier brutal — but if you do, the
challenge must open (immediately after the Difficulty line) with a short
"**Beyond the syllabus:**" line naming exactly which not-yet-covered skills it
uses, so the student knows to research them first.`
	}
	return `THE SEQUENCE RULE: the challenge — its requirements, examples, starter code,
stretch goals, and hints — must be solvable using ONLY the covered skills
listed above. Do not require, or write feedback-bait around, anything not yet
covered. If a tempting idea needs an uncovered skill, pick a different idea.`
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
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Completed    bool   `json:"completed"`
		LessonCached bool   `json:"lessonCached"`
		QuizCached   bool   `json:"quizCached"`
		// One flag per difficulty tier, e.g. {"beginner": true, "goat": false}.
		ChallengeCached map[string]bool `json:"challengeCached"`
		// One flag per tier: the student marked that tier's challenge complete.
		ChallengeDone map[string]bool `json:"challengeDone"`
	}
	result := make([]TopicResp, len(lang.Topics))
	for i, t := range lang.Topics {
		challenges := make(map[string]bool, len(difficultyOrder))
		challengesDone := make(map[string]bool, len(difficultyOrder))
		for _, d := range difficultyOrder {
			challenges[d] = cacheHas(user, challengeCacheKey(langID, t.ID, d))
			challengesDone[d] = done[fmt.Sprintf("%d:challenge:%s", t.ID, d)]
		}
		result[i] = TopicResp{
			ID:              t.ID,
			Name:            t.Name,
			Completed:       done[fmt.Sprintf("%d", t.ID)],
			LessonCached:    cacheHas(user, fmt.Sprintf("%s:lesson:%d", langID, t.ID)),
			QuizCached:      cacheHas(user, quizCacheKey(langID, t.ID)),
			ChallengeCached: challenges,
			ChallengeDone:   challengesDone,
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
	topic, ok := lookupTopic(w, lang, req.TopicID)
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

	// The teaching sequence: this lesson must deliver its topic's skills (the
	// contract later challenges rely on) and may assume only earlier topics.
	prior := topicSkillsBlock(lang, topic.ID-1)
	if prior == "" {
		prior = "Nothing — this is the very first topic. Assume no prior " + lang.Name + " knowledge at all."
	}

	prompt := fmt.Sprintf(`Teach Topic %d of %d: **%s** in %s.

**This lesson MUST cover these skills** — later challenges rely on them having been taught here:
%s

**What earlier lessons already covered** (safe to use without re-explaining):
%s

Do not build examples on any concept beyond this topic's skills and the earlier
topics above. If a small forward reference is truly unavoidable, explain it
inline in one sentence — never assume it.

This lesson is the student's main study text for the topic — be generous with
depth. Explain each idea fully rather than gesturing at it; a motivated
beginner should be able to attempt the challenges from this lesson alone.

Use this structure:

## Overview
2–3 sentences introducing the concept and why it matters in %s.

## Key Concepts
5–7 bullet points. Give each concept 2–3 sentences of real explanation, and
where a single line of code makes the idea concrete, include it inline.

## Code Examples

### Example 1: [Descriptive title]
A minimal, clear example of the core idea. After the code, walk through what
happens when it runs, step by step, in plain prose.

### Example 2: [Descriptive title]
A more practical, realistic usage. Explain what it adds beyond Example 1.

### Example 3: [Descriptive title]
An example that combines this topic with skills from the earlier topics above
(for the very first topic, show a variation or edge case instead). Explain how
the pieces work together.

## How It Works
A short section one level deeper: the mental model behind this topic — what
%s is actually doing and why the language designed it this way. Keep it
beginner-friendly; no internals beyond what helps them reason about their code.

## Common Pitfalls
3–4 specific mistakes beginners make with this topic. For each, show the
mistake and the fix — a short wrong-vs-right code snippet where it helps.

## Summary
One crisp sentence: the essential takeaway.`,
		topic.ID, len(lang.Topics), topic.Name, lang.Name,
		topic.Skills,
		prior,
		lang.Name,
		topic.Name)

	streamLLM(r.Context(), w, user, lang.SystemPrompt, prompt, nil, func(full string) {
		cacheStore(user, key, full)
	})
}

func handleChallenge(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang       string `json:"lang"`
		TopicID    int    `json:"topic_id"`
		TopicName  string `json:"topic_name"`
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

	topic, ok := lookupTopic(w, lang, req.TopicID)
	if !ok {
		return
	}

	diff := normalizeDifficulty(req.Difficulty)
	spec := difficultySpec[diff]

	// Challenges are cached just like lessons (one entry per difficulty tier)
	// so they survive a server restart.
	key := challengeCacheKey(req.Lang, req.TopicID, diff)
	if !req.Force {
		if cached, hit := cacheGet(user, key); hit {
			streamCached(w, cached)
			return
		}
	}

	// A challenge is generated FROM its lesson: the lesson text goes into the
	// prompt below so the challenge exercises what was actually taught. So the
	// lesson must be generated first. The UI checks this too and nudges the
	// student to the lesson tab; this is the authoritative guard.
	lessonKey := fmt.Sprintf("%s:lesson:%d", req.Lang, req.TopicID)
	if !cacheHas(user, lessonKey) {
		http.Error(w, "load the lesson first — the challenge is generated from it", http.StatusConflict)
		return
	}

	prompt := fmt.Sprintf(`Create a practical %s coding challenge for Topic %d: **%s**.

**What the course has covered so far** (this topic's lesson included):
%s

%s

This is the **%s** tier of four (Beginner → Intermediate → Advanced → GOAT).
Tier brief: %s

## Challenge: [Creative, specific title]

**Difficulty**: %s

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

**Stretch Goals**
2 optional extensions for students who finish early, as a short bulleted list.
Each must deepen this same topic — no new topics, and no code blocks (prose
only; the editor pre-fills from the last code block, which must stay the
starter code). State clearly that these are optional and not needed to pass.

**Starter Code**
%s

## Hints
Exactly 3 progressive hints as a numbered list: a gentle nudge, a stronger
pointer, then a near-spoiler. This must be the LAST section — the UI keeps it
hidden until the student chooses to reveal it — so never reference the hints
from any other section.

Use the starter code above as a base, trimming or extending it to match the
tier brief. Keep the challenge centred on %s, exercised through the covered
skills above. Make it engaging with a real-world flavor. The lesson the
student just studied is included below — keep the challenge consistent with
its terminology and examples, and don't require anything it didn't teach.`,
		lang.Name, topic.ID, topic.Name,
		topicSkillsBlock(lang, topic.ID),
		sequenceRule(diff),
		spec.Label, spec.Brief,
		spec.Label,
		lang.StarterTemplate, topic.Name)
	prompt += lessonContextBlock(user, lessonKey)

	streamLLM(r.Context(), w, user, lang.SystemPrompt, prompt, nil, func(full string) {
		cacheStore(user, key, full)
		// A fresh challenge starts with a clean hint record — hint use on the
		// old challenge shouldn't count against the new one.
		clearHintsUsed(user, key)
	})
}

func handleEvaluate(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang       string `json:"lang"`
		TopicID    int    `json:"topic_id"`
		TopicName  string `json:"topic_name"`
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
	topic, ok := lookupTopic(w, lang, req.TopicID)
	if !ok {
		return
	}
	challengeKey := challengeCacheKey(req.Lang, req.TopicID, normalizeDifficulty(req.Difficulty))

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

Judge within the course. The student has covered ONLY these skills so far:
%s
Never fail the submission or list an issue for not using techniques beyond them
(no demanding error handling, methods, or concurrency from a student who hasn't
met them yet). A one-line "later in the course you'll learn..." aside is fine.

Stretch goals are optional extras: never fail or penalize a submission for skipping
them. If the student attempted one, evaluate the attempt and celebrate it in
**What Works Well**.

Be encouraging and educational. Note: code cannot be executed — evaluate on logic and conventions.`,
		lang.Name, topic.ID, topic.Name,
		req.Challenge,
		lang.ID, req.Code,
		lang.Name,
		lang.Name, lang.StyleNote,
		topicSkillsBlock(lang, topic.ID))
	prompt += lessonContextBlock(user, fmt.Sprintf("%s:lesson:%d", req.Lang, req.TopicID))
	prompt += hintUsageBlock(user, challengeKey)

	// Save the submission and its feedback so the student can come back to
	// them later. onComplete only fires on a clean finish, so a truncated
	// evaluation is never saved.
	streamLLM(r.Context(), w, user, lang.SystemPrompt, prompt, nil, func(full string) {
		storeSolution(user, challengeKey, req.Code, full)
	})
}

func handleHint(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang       string `json:"lang"`
		TopicID    int    `json:"topic_id"`
		TopicName  string `json:"topic_name"`
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

	prompt := fmt.Sprintf(`Give exactly ONE helpful hint for this %s challenge on **%s**.

Challenge:
%s

Student's current code:
`+"```%s\n%s\n```"+`

Give ONE specific, encouraging nudge that moves them forward without revealing the answer. Maximum 3 sentences.
Stay within the skills the challenge itself uses — don't hint toward techniques the course hasn't covered yet.`,
		lang.Name, req.TopicName, req.Challenge, lang.ID, req.Code)
	prompt += lessonContextBlock(user, fmt.Sprintf("%s:lesson:%d", req.Lang, req.TopicID))

	// Record hint use only when the hint was actually delivered (onComplete
	// fires on a clean finish) — a failed request shouldn't count against the
	// student's no-hints run. The evaluation reads this flag.
	streamLLM(r.Context(), w, user, lang.SystemPrompt, prompt, nil, func(string) {
		markHintsUsed(user, challengeCacheKey(req.Lang, req.TopicID, normalizeDifficulty(req.Difficulty)))
	})
}

// handleHintsViewed records that the student revealed the hidden Hints
// section of a fundamentals challenge. Revealing counts the same as pressing
// the 💡 Hint button, so the evaluation can recognise a no-hints solve.
func handleHintsViewed(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang       string `json:"lang"`
		TopicID    int    `json:"topic_id"`
		Difficulty string `json:"difficulty"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	if _, ok := lookupLang(w, req.Lang); !ok {
		return
	}
	markHintsUsed(user, challengeCacheKey(req.Lang, req.TopicID, normalizeDifficulty(req.Difficulty)))
	jsonOK(w, map[string]bool{"success": true})
}

func handleChat(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang       string    `json:"lang"`
		TopicID    int       `json:"topic_id"`
		TopicName  string    `json:"topic_name"`
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

	ctx := chatContextBlock(
		user,
		fmt.Sprintf("%s:lesson:%d", req.Lang, req.TopicID),
		challengeCacheKey(req.Lang, req.TopicID, normalizeDifficulty(req.Difficulty)),
	)
	system := fmt.Sprintf(`%s
The student is studying Topic %d: %s. Answer their questions clearly, concisely, and encouragingly. When relevant, ground your answer in the lesson and challenge below.%s`,
		lang.SystemPrompt, req.TopicID, req.TopicName, ctx)

	// Save the conversation so it survives a reload. The history the client
	// sends already includes the newest question; add the answer on top.
	streamLLM(r.Context(), w, user, system, "", req.Messages, func(full string) {
		storeChat(user, chatStoreKey(req.Lang, req.TopicID),
			append(req.Messages, Message{Role: "assistant", Content: full}))
	})
}
