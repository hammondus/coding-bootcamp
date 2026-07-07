package main

import (
	"fmt"
	"net/http"
	"strings"
)

// ── Quiz — the optional knowledge check between lesson and challenge ──
//
// After studying a lesson the student can take a short quiz: four multiple
// choice questions plus two free-response ones, generated FROM the cached
// lesson just like a challenge is. The quiz is entirely optional — nothing
// gates on it, and no progress is recorded — it exists so the student can
// find gaps before the challenge does.
//
// The quiz document deliberately contains NO answer key (it is shown to the
// student verbatim). Grading is a separate call: the browser sends the quiz
// text back with the student's answers and the model marks them in one
// streamed response.

// quizCacheKey names a fundamentals quiz, e.g. "go:quiz:3".
func quizCacheKey(langID string, topicID int) string {
	return fmt.Sprintf("%s:quiz:%d", langID, topicID)
}

// trackQuizCacheKey names an advanced-track quiz, e.g. "go:track:web:quiz:2".
func trackQuizCacheKey(langID, trackID string, lessonID int) string {
	return fmt.Sprintf("%s:track:%s:quiz:%d", langID, trackID, lessonID)
}

// quizFormatRules is the strict output format shared by both quiz prompts.
// The frontend parses these exact patterns into an interactive answer form
// (radio buttons for A–D, a textarea for free response), so the headings and
// option markers must not vary.
const quizFormatRules = `Format the quiz EXACTLY like this — the app parses it into an answer form:

## Quiz: [the topic name]
One friendly sentence saying this is a quick optional self-check before the
challenge.

### Question 1 (Multiple Choice)
The question. A short fenced code snippet is welcome when it helps (e.g.
"what does this print?").
- A) first option
- B) second option
- C) third option
- D) fourth option

Strict rules:
- Exactly 6 questions: Questions 1–4 are Multiple Choice, Questions 5–6 are
  Free Response (a short written answer: predict the output, spot the bug, or
  explain why something works).
- Every question heading must be exactly "### Question N (Multiple Choice)"
  or "### Question N (Free Response)" — nothing else on that line.
- Every Multiple Choice question has exactly four options, one per line,
  starting "- A) ", "- B) ", "- C) ", "- D) ", with exactly one correct.
- Free Response questions have no options.
- Do NOT include the answers, an answer key, or hints toward the answer
  anywhere — grading happens in a separate step.
- Each question should test a different skill from the lesson.`

// quizPrompt builds the generation prompt for either kind of quiz. subject
// names what's being tested ("Topic 3: Loops" / a track lesson), covered is
// the covered-skills block the questions must stay inside.
func quizPrompt(langName, subject, covered string) string {
	return fmt.Sprintf(`Create a short knowledge-check quiz on %s in %s.

The quiz tests ONLY what the lesson included below actually taught. Questions
must be answerable from the lesson alone — no outside knowledge, and nothing
from later topics.

**What the course has covered so far:**
%s

%s`, subject, langName, covered, quizFormatRules)
}

// quizGradePrompt builds the grading prompt: the quiz text plus the student's
// answers, marked in one streamed markdown response.
func quizGradePrompt(langName, subject, quiz string, answers []string) string {
	var sb strings.Builder
	for i, a := range answers {
		a = strings.TrimSpace(a)
		if a == "" {
			a = "(no answer)"
		}
		fmt.Fprintf(&sb, "**Question %d:** %s\n\n", i+1, a)
	}

	return fmt.Sprintf(`Grade this short %s quiz on %s. The quiz and the student's answers are below.

**The Quiz:**
%s

**The student's answers:**
%s
Respond in EXACTLY this format:

## Results
**Score: N / %d** — first line, counting each correct answer as 1 (a partly
right free response earns 0.5).

### Question 1: [✅ Correct | 🟡 Partly right | ❌ Not quite]
One or two sentences. For multiple choice: name the correct option and why.
For free response: what was right about their answer and what was missing.
For "(no answer)": mark ❌ and simply explain the correct answer — no scolding.
(Repeat for every question, in order.)

## What to Review
If anything was missed: 1–3 bullets pointing back to the lesson section worth
re-reading. If everything was correct: skip the bullets and celebrate instead.

Grading rules:
- For multiple choice, accept a bare letter ("B"), the option text, or both.
- For free response, judge the idea, not the wording — beginners phrasing
  things loosely is fine. Exact output questions do need the right output.
- This quiz is optional practice, not a gate. Be warm and encouraging; the
  goal is that the student knows exactly what to re-read before the challenge.`,
		langName, subject, quiz, sb.String(), len(answers))
}

// ── Fundamentals quiz handlers ────────────────────────

func handleQuiz(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang    string `json:"lang"`
		TopicID int    `json:"topic_id"`
		Force   bool   `json:"force"` // true = bypass cache (New Quiz button)
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

	key := quizCacheKey(req.Lang, req.TopicID)
	if !req.Force {
		if cached, hit := cacheGet(user, key); hit {
			streamCached(w, cached)
			return
		}
	}

	// Like a challenge, a quiz is generated FROM its lesson, so the lesson
	// must exist first. The UI checks this too; this is the authoritative guard.
	lessonKey := fmt.Sprintf("%s:lesson:%d", req.Lang, req.TopicID)
	if !cacheHas(user, lessonKey) {
		http.Error(w, "load the lesson first — the quiz is generated from it", http.StatusConflict)
		return
	}

	subject := fmt.Sprintf("Topic %d: **%s**", topic.ID, topic.Name)
	prompt := quizPrompt(lang.Name, subject, topicSkillsBlock(lang, topic.ID))
	prompt += lessonContextBlock(user, lessonKey)

	streamLLM(r.Context(), w, user, lang.SystemPrompt, prompt, nil, func(full string) {
		cacheStore(user, key, full)
	})
}

func handleQuizGrade(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang    string   `json:"lang"`
		TopicID int      `json:"topic_id"`
		Quiz    string   `json:"quiz"`
		Answers []string `json:"answers"` // one per question, "" = unanswered
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
	if req.Quiz == "" || len(req.Answers) == 0 {
		jsonErr(w, 400, "nothing to grade")
		return
	}

	subject := fmt.Sprintf("Topic %d: **%s**", topic.ID, topic.Name)
	prompt := quizGradePrompt(lang.Name, subject, req.Quiz, req.Answers)
	prompt += lessonContextBlock(user, fmt.Sprintf("%s:lesson:%d", req.Lang, req.TopicID))

	// Grading isn't cached — each attempt is fresh feedback on fresh answers.
	streamLLM(r.Context(), w, user, lang.SystemPrompt, prompt, nil, nil)
}

// ── Advanced-track quiz handlers ──────────────────────

func handleTrackQuiz(w http.ResponseWriter, r *http.Request, user string) {
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

	key := trackQuizCacheKey(req.Lang, req.TrackID, req.LessonID)
	if !req.Force {
		if cached, hit := cacheGet(user, key); hit {
			streamCached(w, cached)
			return
		}
	}

	lessonKey := fmt.Sprintf("%s:track:%s:%d", req.Lang, req.TrackID, req.LessonID)
	if !cacheHas(user, lessonKey) {
		http.Error(w, "load the lesson first — the quiz is generated from it", http.StatusConflict)
		return
	}

	subject := fmt.Sprintf("Lesson %d: **%s** in the **%s** track", lesson.ID, lesson.Title, track.Title)
	covered := trackAssumedBlock(lang, track) + "\n\n" + buildTrackContext(track, lesson.ID)
	prompt := quizPrompt(lang.Name, subject, covered)
	prompt += lessonContextBlock(user, lessonKey)

	streamLLM(r.Context(), w, user, lang.SystemPrompt, prompt, nil, func(full string) {
		cacheStore(user, key, full)
	})
}

func handleTrackQuizGrade(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang     string   `json:"lang"`
		TrackID  string   `json:"track_id"`
		LessonID int      `json:"lesson_id"`
		Quiz     string   `json:"quiz"`
		Answers  []string `json:"answers"`
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
	if req.Quiz == "" || len(req.Answers) == 0 {
		jsonErr(w, 400, "nothing to grade")
		return
	}

	subject := fmt.Sprintf("Lesson %d: **%s** in the **%s** track", lesson.ID, lesson.Title, track.Title)
	prompt := quizGradePrompt(lang.Name, subject, req.Quiz, req.Answers)
	prompt += lessonContextBlock(user, fmt.Sprintf("%s:track:%s:%d", req.Lang, req.TrackID, req.LessonID))

	streamLLM(r.Context(), w, user, lang.SystemPrompt, prompt, nil, nil)
}
