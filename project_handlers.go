package main

import (
	"fmt"
	"net/http"
	"strings"
)

// ── Project helpers ───────────────────────────────────

// findProject looks up a project by ID within a language.
func findProject(lang Language, projectID string) (Project, bool) {
	for _, p := range lang.Projects {
		if p.ID == projectID {
			return p, true
		}
	}
	return Project{}, false
}

// findProjectMilestone looks up a project and milestone by ID within a language.
func findProjectMilestone(lang Language, projectID string, milestoneID int) (Project, ProjectMilestone, bool) {
	for _, p := range lang.Projects {
		if p.ID == projectID {
			for _, m := range p.Milestones {
				if m.ID == milestoneID {
					return p, m, true
				}
			}
		}
	}
	return Project{}, ProjectMilestone{}, false
}

// lookupProject resolves a project within a language, writing a 400 and
// returning ok=false when it's unknown.
func lookupProject(w http.ResponseWriter, lang Language, projectID string) (Project, bool) {
	p, ok := findProject(lang, projectID)
	if !ok {
		http.Error(w, "unknown project", http.StatusBadRequest)
	}
	return p, ok
}

// lookupProjectMilestone resolves a project + milestone within a language,
// writing a 400 and returning ok=false when either is unknown.
func lookupProjectMilestone(w http.ResponseWriter, lang Language, projectID string, milestoneID int) (Project, ProjectMilestone, bool) {
	p, m, ok := findProjectMilestone(lang, projectID, milestoneID)
	if !ok {
		http.Error(w, "unknown project or milestone", http.StatusBadRequest)
	}
	return p, m, ok
}

// projectRoadmap renders the full milestone list as markdown, so the brief and
// each milestone prompt can see how the whole project fits together.
func projectRoadmap(p Project) string {
	var sb strings.Builder
	for _, m := range p.Milestones {
		fmt.Fprintf(&sb, "%d. **%s** — %s\n", m.ID, m.Title, m.Summary)
	}
	return sb.String()
}

// buildProjectContext returns a markdown block summarising the milestones
// already built, so each milestone's guidance builds on the previous ones.
func buildProjectContext(p Project, upToMilestoneID int) string {
	if upToMilestoneID <= 1 {
		return "This is the first milestone — the project starts from an empty directory."
	}
	var sb strings.Builder
	sb.WriteString("**What earlier milestones already built:**\n")
	for _, m := range p.Milestones {
		if m.ID >= upToMilestoneID {
			break
		}
		fmt.Fprintf(&sb, "- Milestone %d — **%s**: %s\n", m.ID, m.Title, m.Summary)
	}
	return sb.String()
}

// projectBriefBlock pulls the cached project brief out of the cache so milestone,
// evaluate, hint, and chat prompts can ground their output in the agreed spec.
// Returns "" if the brief hasn't been generated yet, so prompts still work.
func projectBriefBlock(user, briefKey string) string {
	if brief, ok := cacheGet(user, briefKey); ok {
		return "\n\n--- THE PROJECT BRIEF (the agreed spec) ---\n" + brief
	}
	return ""
}

// ── Project handlers ──────────────────────────────────

func handleProjects(w http.ResponseWriter, r *http.Request, user string) {
	langID := r.URL.Query().Get("lang")
	lang, ok := languages[langID]
	if !ok {
		jsonErr(w, 400, "unknown language")
		return
	}

	// Snapshot progress under the lock (via getUserLangProgress) rather than
	// reading the shared map directly — see the same note in handleTracks.
	done := getUserLangProgress(user, langID)

	type MilestoneResp struct {
		ID             int    `json:"id"`
		Title          string `json:"title"`
		Summary        string `json:"summary"`
		Completed      bool   `json:"completed"`
		GuidanceCached bool   `json:"guidanceCached"`
	}
	type ProjectResp struct {
		ID          string          `json:"id"`
		Title       string          `json:"title"`
		Icon        string          `json:"icon"`
		Description string          `json:"description"`
		BriefCached bool            `json:"briefCached"`
		Milestones  []MilestoneResp `json:"milestones"`
	}

	result := make([]ProjectResp, 0, len(lang.Projects))
	for _, p := range lang.Projects {
		milestones := make([]MilestoneResp, len(p.Milestones))
		for i, m := range p.Milestones {
			milestones[i] = MilestoneResp{
				ID:             m.ID,
				Title:          m.Title,
				Summary:        m.Summary,
				Completed:      done[fmt.Sprintf("project:%s:%d", p.ID, m.ID)],
				GuidanceCached: cacheHas(user, fmt.Sprintf("%s:project:%s:milestone:%d", langID, p.ID, m.ID)),
			}
		}
		result = append(result, ProjectResp{
			ID:          p.ID,
			Title:       p.Title,
			Icon:        p.Icon,
			Description: p.Description,
			BriefCached: cacheHas(user, fmt.Sprintf("%s:project:%s:brief", langID, p.ID)),
			Milestones:  milestones,
		})
	}
	jsonOK(w, result)
}

func handleProjectBrief(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang      string `json:"lang"`
		ProjectID string `json:"project_id"`
		Force     bool   `json:"force"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}
	project, ok := lookupProject(w, lang, req.ProjectID)
	if !ok {
		return
	}

	cacheKey := fmt.Sprintf("%s:project:%s:brief", req.Lang, req.ProjectID)
	if !req.Force {
		if cached, hit := cacheGet(user, cacheKey); hit {
			streamCached(w, cached)
			return
		}
	}

	prompt := fmt.Sprintf(`Write the project brief for a capstone %s project the student will build from scratch: **%s**.

**The goal:** %s

**The milestone roadmap (the student will build these in order):**
%s

This is a capstone — it should bring together fundamentals and the advanced track material. Write an engaging, motivating brief with this structure:

## %s
A 2–3 sentence introduction: what they're building and why it's a satisfying thing to build.

## What You'll Build
A concrete description of the finished application — its behaviour from the outside (endpoints, inputs, outputs).

## Requirements
The full functional requirements as a checklist. Be specific and testable.

## Suggested Architecture
The recommended file/package layout and the main types or functions, with one sentence on the role of each. Keep it idiomatic %s. Do NOT write the full implementation — this is a map, not the solution.

## How the Milestones Fit Together
A short paragraph tying the roadmap above into the architecture, so the student sees where each step is heading.

## Skills You'll Draw On
A short bullet list of the fundamentals and advanced concepts this project exercises.

Be encouraging. Frame it as "from scratch, bringing together everything you've learned."`,
		lang.Name, project.Title,
		project.Goal,
		projectRoadmap(project),
		project.Title,
		lang.Name,
	)

	streamFromAnthropic(r.Context(), w, lang.SystemPrompt, prompt, nil, func(full string) {
		cacheStore(user, cacheKey, full)
	})
}

func handleProjectMilestone(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang        string `json:"lang"`
		ProjectID   string `json:"project_id"`
		MilestoneID int    `json:"milestone_id"`
		Force       bool   `json:"force"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}
	project, milestone, ok := lookupProjectMilestone(w, lang, req.ProjectID, req.MilestoneID)
	if !ok {
		return
	}

	cacheKey := fmt.Sprintf("%s:project:%s:milestone:%d", req.Lang, req.ProjectID, req.MilestoneID)
	if !req.Force {
		if cached, hit := cacheGet(user, cacheKey); hit {
			streamCached(w, cached)
			return
		}
	}

	prevCtx := buildProjectContext(project, milestone.ID)

	prompt := fmt.Sprintf(`The student is building the **%s** project in %s and has reached Milestone %d of %d: **%s**.

%s

## Your task: guide them through **%s**

Give focused build guidance for THIS milestone only — not the whole project, and not a finished solution. Structure it as:

## Objective
1–2 sentences: what this milestone adds to the project and why it matters now.

## Requirements
A specific, testable checklist for this milestone.

## Approach
The key steps and which %s concepts or standard-library packages to reach for. Mention relevant patterns from the fundamentals or advanced tracks where they apply. You may show small illustrative snippets, but do NOT write the complete milestone solution — the student writes that.

## Done When
Clear acceptance criteria: how the student knows this milestone works before moving on.

Be encouraging and concrete.`,
		project.Title, lang.Name, milestone.ID, len(project.Milestones), milestone.Title,
		prevCtx,
		milestone.Title,
		lang.Name,
	)
	prompt += projectBriefBlock(user, fmt.Sprintf("%s:project:%s:brief", req.Lang, req.ProjectID))

	streamFromAnthropic(r.Context(), w, lang.SystemPrompt, prompt, nil, func(full string) {
		cacheStore(user, cacheKey, full)
	})
}

func handleProjectEvaluate(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang        string `json:"lang"`
		ProjectID   string `json:"project_id"`
		MilestoneID int    `json:"milestone_id"`
		Code        string `json:"code"`
		Challenge   string `json:"challenge"` // the milestone guidance the student is working against
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}
	project, milestone, ok := lookupProjectMilestone(w, lang, req.ProjectID, req.MilestoneID)
	if !ok {
		return
	}

	prompt := fmt.Sprintf(`Evaluate this %s code submission for the **%s** project, Milestone %d: **%s**.

**The milestone guidance:**
%s

**Submitted Code:**
`+"```%s\n%s\n```"+`

## Verdict
✅ **Pass** OR ❌ **Needs Work** — state it clearly on the first line, judged against THIS milestone's requirements only.

## What Works Well
Specific praise, especially for sound architecture and idiomatic %s.

## Issues Found
Specific bugs, logical errors, or style issues. Write "None — looks good!" if clean.

## %s Style Note
%s

## Suggested Improvement (if needed)
A corrected or more idiomatic version of the relevant part. Skip if the code passed cleanly.

Be encouraging. Note: code cannot be executed — evaluate on logic and conventions.`,
		lang.Name, project.Title, milestone.ID, milestone.Title,
		req.Challenge,
		lang.ID, req.Code,
		lang.Name,
		lang.Name, lang.StyleNote,
	)
	prompt += projectBriefBlock(user, fmt.Sprintf("%s:project:%s:brief", req.Lang, req.ProjectID))

	streamFromAnthropic(r.Context(), w, lang.SystemPrompt, prompt, nil)
}

func handleProjectHint(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang        string `json:"lang"`
		ProjectID   string `json:"project_id"`
		MilestoneID int    `json:"milestone_id"`
		Challenge   string `json:"challenge"`
		Code        string `json:"code"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}
	project, milestone, ok := lookupProjectMilestone(w, lang, req.ProjectID, req.MilestoneID)
	if !ok {
		return
	}

	prompt := fmt.Sprintf(`Give ONE helpful hint for this milestone of the **%s** project (%s), Milestone %d: **%s**.

Milestone guidance:
%s

Student's current code:
`+"```%s\n%s\n```"+`

Give ONE specific, encouraging nudge that moves them forward without revealing the full solution. Maximum 3 sentences.`,
		project.Title, lang.Name, milestone.ID, milestone.Title,
		req.Challenge, lang.ID, req.Code,
	)
	prompt += projectBriefBlock(user, fmt.Sprintf("%s:project:%s:brief", req.Lang, req.ProjectID))

	streamFromAnthropic(r.Context(), w, lang.SystemPrompt, prompt, nil)
}

func handleProjectChat(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang        string    `json:"lang"`
		ProjectID   string    `json:"project_id"`
		MilestoneID int       `json:"milestone_id"`
		Messages    []Message `json:"messages"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	lang, ok := lookupLang(w, req.Lang)
	if !ok {
		return
	}
	project, milestone, ok := lookupProjectMilestone(w, lang, req.ProjectID, req.MilestoneID)
	if !ok {
		return
	}

	// Ground chat in the brief and the current milestone's guidance.
	ctx := projectBriefBlock(user, fmt.Sprintf("%s:project:%s:brief", req.Lang, req.ProjectID))
	if guidance, hit := cacheGet(user, fmt.Sprintf("%s:project:%s:milestone:%d", req.Lang, req.ProjectID, req.MilestoneID)); hit {
		ctx += "\n\n--- THE CURRENT MILESTONE GUIDANCE ---\n" + guidance
	}
	system := fmt.Sprintf(`%s
The student is building the **%s** project, currently on Milestone %d: %s.
Answer their questions clearly and in the context of this project and milestone. Help them think it through — guide, don't just hand over the whole solution. When relevant, ground your answer in the brief and milestone guidance below.%s`,
		lang.SystemPrompt, project.Title, milestone.ID, milestone.Title, ctx,
	)

	streamFromAnthropic(r.Context(), w, system, "", req.Messages)
}
