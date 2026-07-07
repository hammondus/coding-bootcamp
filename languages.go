package main

// Topic is a single lesson in a bootcamp curriculum. Skills is the contract
// between lessons and challenges: the lesson prompt must teach these skills,
// and challenge prompts may only require skills from this topic and earlier
// ones (see topicSkillsBlock in content.go). Keep each entry a concrete
// one-liner, like TrackLesson summaries.
type Topic struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Skills string `json:"-"`
}

// TrackLesson is one lesson within an advanced track.
type TrackLesson struct {
	ID      int
	Title   string
	Summary string // one-liner used to build progressive context for subsequent lessons
}

// Track is a multi-lesson deep-dive on a specific subject.
type Track struct {
	ID          string
	Title       string
	Icon        string
	Description string
	// Prereqs names what this track assumes beyond the language fundamentals
	// (other tracks or specific lessons). It is rendered into every track
	// prompt so generated content never requires skills taught elsewhere
	// without saying so. Empty = fundamentals only.
	Prereqs string
	Lessons []TrackLesson
}

// ProjectMilestone is one build step within a project. Unlike a TrackLesson
// (which is taught), a milestone is something the student builds: the generated
// guidance describes what to implement, and the student submits code for it.
type ProjectMilestone struct {
	ID      int
	Title   string
	Summary string // one-liner used to build progressive context for later milestones
}

// Project is a capstone: a complete application the student builds from scratch
// across several milestones, bringing together fundamentals and advanced track
// material. A project has one generated brief (the spec) plus per-milestone
// build guidance — see project_handlers.go.
type Project struct {
	ID          string
	Title       string
	Icon        string
	Description string
	Goal        string // one-line end goal, fed into the brief and milestone prompts
	// Prereqs names the fundamentals and tracks this capstone assumes. It is
	// rendered into the brief, milestone, evaluate, and chat prompts so the
	// student knows what to have covered and the guidance stays within it.
	Prereqs    string
	Milestones []ProjectMilestone
}

// Language defines a bootcamp language's complete configuration.
// To add a new language: create a var below and add it to the languages map.
type Language struct {
	ID              string
	Name            string
	Icon            string
	Cmd             string
	AccentColor     string
	AccentDark      string
	AccentGlow      string
	CodeLabel       string
	StyleNote       string
	StarterTemplate string
	SystemPrompt    string
	Topics          []Topic
	Tracks          []Track   // advanced multi-lesson tracks; nil = no tracks for this language
	Projects        []Project // capstone projects; nil = no projects for this language
}

// Category groups related languages together in the UI switcher.
type Category struct {
	ID    string
	Name  string
	Langs []string // language IDs, in display order within the category
}

// categories controls how languages are grouped and ordered in the UI switcher.
// To add a language: define its var below, add it to the languages map, and list
// its ID in the appropriate category here.
var categories = []Category{
	{ID: "backend", Name: "Backend", Langs: []string{"go", "zig"}},
	{ID: "frontend", Name: "Frontend", Langs: []string{"javascript", "html", "css"}},
	// "Development" holds craft courses that aren't tied to one programming
	// language: code management (git) and AI-assisted development (claude).
	{ID: "development", Name: "Development", Langs: []string{"git", "claude"}},
}

var languages = map[string]Language{
	"go":         goLanguage,
	"zig":        zigLanguage,
	"javascript": javascriptLanguage,
	"html":       htmlLanguage,
	"css":        cssLanguage,
	"claude":     claudeLanguage,
	"git":        gitLanguage,
}

// ── Go ────────────────────────────────────────────────

var goLanguage = Language{
	ID:          "go",
	Name:        "Go",
	Icon:        "/static/icons/go.svg",
	Cmd:         "$ go run .",
	AccentColor: "#00ADD8",
	AccentDark:  "#007fa0",
	AccentGlow:  "rgba(0,173,216,0.15)",
	CodeLabel:   "GO",
	StyleNote:   "One targeted tip on camelCase vs PascalCase, idiomatic error handling, or Go conventions.",
	StarterTemplate: "```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n\t// Your solution here\n}\n```",
	SystemPrompt: `You are an expert Go programming instructor running a hands-on bootcamp.
You teach Go clearly, practically, and engagingly. You are patient and encouraging.
Format all responses in Markdown. Use triple-backtick go fences for all Go code examples.
Be concise but thorough. Include practical, real-world examples.
Style reminder: Go uses camelCase for unexported names and PascalCase for exported names. Flag naming issues if you see them.`,
	// Topic order is a strict teaching sequence: a challenge may only require
	// skills from its own topic and earlier ones. Closures & Defer must stay
	// before Goroutines & Channels (go func(){} is a closure; defer unlocks
	// mutexes), and Strings & Runes earns its own topic because string work
	// appears in most challenges.
	Topics: []Topic{
		{ID: 1, Name: "Program Structure",
			Skills: "package main, func main, imports, comments, go run; printing with fmt.Println, fmt.Printf verbs (%v %s %d), and fmt.Sprintf"},
		{ID: 2, Name: "Variables & Types",
			Skills: "var and := declarations, basic types (int, float64, string, bool), zero values, constants and iota, type conversions, string concatenation with +"},
		{ID: 3, Name: "Control Flow",
			Skills: "if / else if / else, comparison and logical operators, switch with cases and default, the if-with-statement form"},
		{ID: 4, Name: "Loops",
			Skills: "for in all forms (classic, while-style, infinite), break and continue, nested loops, looping over a counter"},
		{ID: 5, Name: "Functions",
			Skills: "parameters and return values, multiple returns, named results, variadic functions, functions as values"},
		{ID: 6, Name: "Arrays & Slices",
			Skills: "arrays vs slices, make, append, len and cap, slicing syntax, range over slices, copy"},
		{ID: 7, Name: "Strings & Runes",
			Skills: "immutability, bytes vs runes, ranging over a string, building strings efficiently, strings package basics (Contains, Split, Join, Fields, ToUpper), strconv.Atoi / Itoa"},
		{ID: 8, Name: "Maps",
			Skills: "make, insert / update / delete, the comma-ok lookup, ranging over maps, maps as counters and sets"},
		{ID: 9, Name: "Structs",
			Skills: "defining struct types, literals and field access, nested structs, anonymous structs, structs in slices and maps"},
		{ID: 10, Name: "Pointers",
			Skills: "& and *, pointers to structs, pass-by-value vs pointer semantics, when a function needs a pointer, new"},
		{ID: 11, Name: "Methods",
			Skills: "method declarations, value vs pointer receivers, method sets, constructor functions by convention (NewX)"},
		{ID: 12, Name: "Interfaces",
			Skills: "defining interfaces, implicit satisfaction, the empty interface / any, type assertions with comma-ok, fmt.Stringer"},
		{ID: 13, Name: "Error Handling",
			Skills: "the error interface, returning and checking errors, errors.New and fmt.Errorf with %w, sentinel errors, errors.Is / errors.As"},
		{ID: 14, Name: "Closures & Defer",
			Skills: "anonymous functions, closures capturing variables, defer semantics and LIFO ordering, defer for cleanup patterns"},
		{ID: 15, Name: "Goroutines & Channels",
			Skills: "go statements (including go func closures), channels: send / receive / close, buffered channels, range over channels, select, sync.WaitGroup"},
	},
	Tracks:   goTracks,
	Projects: goProjects,
}

// ── Zig ───────────────────────────────────────────────

var zigLanguage = Language{
	ID:          "zig",
	Name:        "Zig",
	Icon:        "/static/icons/zig.svg",
	Cmd:         "$ zig build run",
	AccentColor: "#F7A41D",
	AccentDark:  "#C47D0A",
	AccentGlow:  "rgba(247,164,29,0.15)",
	CodeLabel:   "ZIG",
	StyleNote:   "One targeted tip on snake_case vs PascalCase, explicit error handling, memory safety, or comptime usage.",
	StarterTemplate: "```zig\nconst std = @import(\"std\");\n\npub fn main() void {\n    // Your solution here\n}\n```",
	SystemPrompt: `You are an expert Zig programming instructor running a hands-on bootcamp.
You teach Zig clearly, practically, and engagingly. You are patient and encouraging.
Format all responses in Markdown. Use triple-backtick zig fences for all Zig code examples.
Be concise but thorough. Include practical, real-world examples.
Key Zig principles: no hidden control flow, no hidden allocations, explicit error handling with try/catch, comptime over runtime generics.
Style reminder: Zig uses snake_case for variables and function names, PascalCase for types and structs. Flag naming issues if you see them.`,
	// Optionals come before Error Handling deliberately: both use |v| payload
	// capture, and optionals are the simpler introduction to it. Printing with
	// .{} tuples is taught as a recipe in topic 1 and fully explained at Structs.
	Topics: []Topic{
		{ID: 1, Name: "Program Structure",
			Skills: "const std = @import(\"std\"), pub fn main, comments, zig run; printing with std.debug.print and its .{} argument tuple (taught as a recipe, fully explained in Structs)"},
		{ID: 2, Name: "Variables & Types",
			Skills: "const vs var, integer and float types (i32, u8, f64...), type inference, undefined, @as and integer casts"},
		{ID: 3, Name: "Control Flow",
			Skills: "if / else as statements and as expressions, switch with ranges and else, logical operators, comparison operators"},
		{ID: 4, Name: "Loops",
			Skills: "while with continue expressions, for over slices and index ranges, break and continue, labeled loops"},
		{ID: 5, Name: "Functions",
			Skills: "fn syntax, parameters and return types, pub, early returns, calling conventions of small pure functions"},
		{ID: 6, Name: "Arrays & Slices",
			Skills: "fixed arrays, slices, strings as []const u8, slicing syntax, string literals, iterating bytes, std.mem.eql for comparison"},
		{ID: 7, Name: "Pointers",
			Skills: "*T single-item pointers, &, dereference with .*, const pointers, pointers vs slices"},
		{ID: 8, Name: "Structs",
			Skills: "struct definition, field defaults, struct literals (and why .{} tuples worked all along), namespaced functions with self, returning structs"},
		{ID: 9, Name: "Enums & Unions",
			Skills: "enum declaration, exhaustive switch on enums, methods on enums, tagged unions, payload capture with |v|"},
		{ID: 10, Name: "Optionals",
			Skills: "?T, null, orelse, if (x) |v| capture, while (it.next()) |v| iteration, optional pointers"},
		{ID: 11, Name: "Error Handling",
			Skills: "error sets, error unions !T, try, catch with capture, errdefer, main returning anyerror!void"},
		{ID: 12, Name: "Allocators & Memory",
			Skills: "the allocator parameter convention, GeneralPurposeAllocator, alloc / free with defer, std.ArrayList and std.StringHashMap basics"},
		{ID: 13, Name: "Comptime",
			Skills: "comptime values and blocks, comptime function parameters, anytype, generic functions returning types"},
		{ID: 14, Name: "Build System & Testing",
			Skills: "zig init layout, build.zig at a glance, test blocks, std.testing.expect / expectEqual / expectError, zig test"},
	},
	Tracks: zigTracks,
}

// ── JavaScript ────────────────────────────────────────

var javascriptLanguage = Language{
	ID:          "javascript",
	Name:        "JavaScript",
	Icon:        "/static/icons/javascript.svg",
	Cmd:         "$ node app.js",
	AccentColor: "#F7DF1E",
	AccentDark:  "#C7B100",
	AccentGlow:  "rgba(247,223,30,0.15)",
	CodeLabel:   "JS",
	StyleNote:   "One targeted tip on camelCase, const/let over var, strict equality (===), or modern ES idioms.",
	StarterTemplate: "```js\n// Your solution here\n```",
	SystemPrompt: `You are an expert JavaScript programming instructor running a hands-on bootcamp.
You teach JavaScript clearly, practically, and engagingly. You are patient and encouraging.
Format all responses in Markdown. Use triple-backtick js fences for all JavaScript code examples.
Be concise but thorough. Include practical, real-world examples.
Teach modern JavaScript (ES2015+): prefer const/let over var, arrow functions, template literals, destructuring, and async/await.
Style reminder: JavaScript uses camelCase for variables and functions, PascalCase for classes, and strict equality (===) over loose (==). Flag issues if you see them.`,
	// Strings sit at 3 (not after Objects) because template literals and string
	// methods appear in nearly every challenge; Arrays come after Functions so
	// callback methods (map, filter) have been prepared for.
	Topics: []Topic{
		{ID: 1, Name: "Variables & Types",
			Skills: "const and let (never var), primitives, typeof, null vs undefined; printing with console.log"},
		{ID: 2, Name: "Operators & Expressions",
			Skills: "arithmetic operators, strict equality === over ==, logical && || !, ternary, truthiness and falsy values"},
		{ID: 3, Name: "Strings & Template Literals",
			Skills: "template literals and ${}, length, slice / includes / split / trim / toUpperCase / padStart, immutability, comparing strings"},
		{ID: 4, Name: "Control Flow",
			Skills: "if / else if / else, switch, guard clauses and early returns"},
		{ID: 5, Name: "Loops",
			Skills: "for, while, for...of over strings, break and continue, nested loops"},
		{ID: 6, Name: "Functions & Arrow Functions",
			Skills: "declarations vs expressions, arrow syntax, default and rest parameters, return values, functions as arguments (callbacks)"},
		{ID: 7, Name: "Arrays",
			Skills: "literals, index and length, push / pop / shift / slice / indexOf, for...of, callback methods: forEach, map, filter, find, reduce, sort"},
		{ID: 8, Name: "Objects",
			Skills: "object literals, dot vs bracket access, methods and this basics, destructuring, spread, JSON.stringify / JSON.parse"},
		{ID: 9, Name: "Scope & Closures",
			Skills: "block vs function scope, closures capturing variables, counters and factories, common closure pitfalls in loops"},
		{ID: 10, Name: "The DOM",
			Skills: "querySelector / querySelectorAll, textContent vs innerHTML, createElement and append, classList, dataset, styles"},
		{ID: 11, Name: "Events",
			Skills: "addEventListener, the event object, bubbling and delegation, preventDefault, reading form inputs"},
		{ID: 12, Name: "Promises & Async/Await",
			Skills: "the event loop in brief, promises and .then / .catch, async functions and await, try/catch around await, Promise.all"},
		{ID: 13, Name: "Fetch & APIs",
			Skills: "fetch for GET and POST, response.ok and status, .json(), request headers and bodies, error handling around network calls"},
		{ID: 14, Name: "Modules",
			Skills: "export and import (named and default), script type=\"module\", splitting code across files"},
	},
	Tracks:   javascriptTracks,
	Projects: javascriptProjects,
}

// ── HTML ──────────────────────────────────────────────

var htmlLanguage = Language{
	ID:          "html",
	Name:        "HTML",
	Icon:        "/static/icons/html.svg",
	Cmd:         "$ open index.html",
	AccentColor: "#E34F26",
	AccentDark:  "#B23318",
	AccentGlow:  "rgba(227,79,38,0.15)",
	CodeLabel:   "HTML",
	StyleNote:   "One targeted tip on semantic elements, proper nesting, accessibility (alt text, labels), or valid document structure.",
	StarterTemplate: "```html\n<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n  <meta charset=\"UTF-8\">\n  <title>My Page</title>\n</head>\n<body>\n  <!-- Your solution here -->\n</body>\n</html>\n```",
	SystemPrompt: `You are an expert HTML instructor running a hands-on bootcamp.
You teach HTML clearly, practically, and engagingly. You are patient and encouraging.
Format all responses in Markdown. Use triple-backtick html fences for all HTML code examples.
Be concise but thorough. Include practical, real-world examples.
Emphasise semantic, accessible markup: the right element for the job, alt text on images, labels on form controls, and a valid document structure.
Style reminder: prefer semantic elements (header, nav, main, article, section, footer) over generic divs, and lowercase tag and attribute names. Flag issues if you see them.`,
	// The attribute mechanism (name="value", id, class) is pinned to topic 1
	// because attributes are used from topic 3 onward; topic 9 focuses on what
	// it was really about — the <head>.
	Topics: []Topic{
		{ID: 1, Name: "Document Structure",
			Skills: "doctype, html / head / body, charset and title, lang; how attributes work (name=\"value\"), id and class"},
		{ID: 2, Name: "Text & Headings",
			Skills: "h1-h6 hierarchy, p, strong and em, br and hr, blockquote, HTML entities (&amp;, &lt;)"},
		{ID: 3, Name: "Links & Navigation",
			Skills: "a and href (absolute, relative, #anchors, mailto), target and rel, title; grouping links into a simple nav"},
		{ID: 4, Name: "Images & Media",
			Skills: "img with src and meaningful alt, width / height, figure and figcaption, audio and video basics with controls"},
		{ID: 5, Name: "Lists",
			Skills: "ul / ol / li, nesting lists, dl / dt / dd, start and type attributes, lists inside nav"},
		{ID: 6, Name: "Tables",
			Skills: "table / tr / td / th, thead / tbody / tfoot, caption, colspan and rowspan, scope for header cells"},
		{ID: 7, Name: "Forms & Inputs",
			Skills: "form action and method, input types, label with for, name attributes, textarea, select and option, button types"},
		{ID: 8, Name: "Semantic Elements",
			Skills: "header, nav, main, article, section, aside, footer; choosing the right element, when a div is fine"},
		{ID: 9, Name: "Metadata & the <head>",
			Skills: "meta viewport and description, favicon links, linking CSS and JS (defer), Open Graph basics"},
		{ID: 10, Name: "Accessibility",
			Skills: "alt text quality, label association, heading order, landmark elements, keyboard focus, ARIA only when HTML can't"},
		{ID: 11, Name: "Embedding Content",
			Skills: "iframe and its attributes, embedding maps and video players, inline svg basics, details / summary"},
		{ID: 12, Name: "Forms Validation",
			Skills: "required, min / max / maxlength, pattern, type-driven validation (email, url, number), novalidate, validation UX"},
	},
}

// ── CSS ───────────────────────────────────────────────

var cssLanguage = Language{
	ID:          "css",
	Name:        "CSS",
	Icon:        "/static/icons/css.svg",
	Cmd:         "$ open index.html",
	AccentColor: "#1572B6",
	AccentDark:  "#0E4F7E",
	AccentGlow:  "rgba(21,114,182,0.15)",
	CodeLabel:   "CSS",
	StyleNote:   "One targeted tip on selector specificity, the box model, using rem/em over px, or choosing flexbox vs grid.",
	StarterTemplate: "```css\n/* Your solution here */\n```",
	SystemPrompt: `You are an expert CSS instructor running a hands-on bootcamp.
You teach CSS clearly, practically, and engagingly. You are patient and encouraging.
Format all responses in Markdown. Use triple-backtick css fences for all CSS code examples.
Be concise but thorough. Include practical, real-world examples.
Emphasise modern layout (flexbox and grid), the box model, specificity, and responsive design with relative units and media queries.
Style reminder: prefer class selectors over IDs for styling, keep specificity low, and use relative units (rem/em) where appropriate. Flag issues if you see them.`,
	// Transforms come before Transitions because the canonical transition
	// animates a transform. Units get an explicit home in The Box Model.
	// CSS challenges assume basic HTML; the prompts state that prerequisite.
	Topics: []Topic{
		{ID: 1, Name: "Selectors",
			Skills: "element / class / id selectors, grouping, descendant and child combinators, specificity basics, the cascade"},
		{ID: 2, Name: "Colors & Backgrounds",
			Skills: "color formats (named, hex, rgb, hsl), opacity and alpha, background-color / image / size / position, simple gradients"},
		{ID: 3, Name: "The Box Model",
			Skills: "content / padding / border / margin, box-sizing: border-box, width and height; units: px, %, rem, em and when to use each"},
		{ID: 4, Name: "Typography",
			Skills: "font-family stacks, font-size and line-height, font-weight, text-align / decoration / transform, letter-spacing, web fonts"},
		{ID: 5, Name: "Display & Layout",
			Skills: "display block / inline / inline-block / none, normal flow, margin collapsing, overflow, centering basics"},
		{ID: 6, Name: "Flexbox",
			Skills: "display: flex, direction, justify-content and align-items, gap, flex grow / shrink / basis, wrapping"},
		{ID: 7, Name: "Grid",
			Skills: "grid-template-columns / rows, fr and repeat, gap, placing items with line numbers, grid-template-areas"},
		{ID: 8, Name: "Positioning",
			Skills: "static / relative / absolute / fixed / sticky, offsets, z-index and stacking basics, positioning within a relative parent"},
		{ID: 9, Name: "Pseudo-classes & Pseudo-elements",
			Skills: ":hover / :focus / :active, :nth-child, :first-child / :last-child, ::before and ::after with content"},
		{ID: 10, Name: "Transforms",
			Skills: "translate / rotate / scale / skew, transform-origin, combining transforms, transforms don't affect layout"},
		{ID: 11, Name: "Transitions & Animations",
			Skills: "transition property / duration / timing-function / delay, transitioning transforms and colors, @keyframes, animation shorthand"},
		{ID: 12, Name: "Custom Properties",
			Skills: "--variables and var(), defining on :root, fallback values, scoping to components, theming patterns"},
		{ID: 13, Name: "Responsive Design",
			Skills: "media queries, mobile-first workflow, relative units in practice, responsive images, clamp() and fluid type"},
	},
}

// ── Claude & Claude Code ──────────────────────────────

var claudeLanguage = Language{
	ID:          "claude",
	Name:        "Claude",
	Icon:        "/static/icons/claude.svg",
	Cmd:         "$ claude",
	AccentColor: "#D97757",
	AccentDark:  "#B05730",
	AccentGlow:  "rgba(217,119,87,0.15)",
	CodeLabel:   "PROMPT",
	StyleNote:   "One targeted tip on prompt craft: specificity, providing context and done criteria, structuring with sections or examples, or verifying output before trusting it.",
	StarterTemplate: "```markdown\n<!-- Your prompt, CLAUDE.md, or configuration here -->\n```",
	SystemPrompt: `You are an expert instructor teaching AI-assisted software development with Claude and Claude Code, running a hands-on bootcamp.
You teach clearly, practically, and engagingly. You are patient and encouraging.
Format all responses in Markdown. Fence every example: markdown fences for prompts and CLAUDE.md files, json fences for settings and API payloads, bash fences for terminal commands.
Be concise but thorough. Include practical, real-world examples.
This course teaches a craft, not a programming language: student submissions are prompts, configuration files, and written workflows. Evaluate them on clarity, specificity, and whether they would actually steer Claude well — there is no compiler here, so your judgment is the feedback loop.
Claude Code evolves quickly. Teach durable concepts and workflows; where an exact flag, file path, or menu name may have changed since your knowledge was current, say so and point the student to /help and the official docs rather than guessing.`,
	// The sequence runs tool-free first (how models work, prompting) so topic 3
	// onward can assume prompt craft inside Claude Code. Context (4) precedes
	// task-writing (5) because scoping a task well requires knowing what Claude
	// can see. Reviewing output (6) lands before any automation topic — the
	// habit of verifying must exist before hooks and headless runs remove the
	// human from the loop. The API closes the course as the bridge to the
	// advanced tracks.
	Topics: []Topic{
		{ID: 1, Name: "How Claude Works",
			Skills: "what an LLM is in practice: tokens, the context window, training cutoff; why phrasing changes results; model tiers (Opus, Sonnet, Haiku) and the capability/speed/cost trade-off; what models can't do: no memory between conversations, confident-sounding mistakes (hallucination)"},
		{ID: 2, Name: "Prompting Fundamentals",
			Skills: "being specific and complete in one message, providing context the model can't guess, stating constraints and the desired output format, showing an example of what good looks like, iterating on a prompt instead of abandoning it"},
		{ID: 3, Name: "Claude Code First Steps",
			Skills: "starting a session with claude, giving a first task, how Claude Code reads files / edits / runs commands on its own, responding to permission prompts, /help and /clear, resuming a previous session"},
		{ID: 4, Name: "Context: What Claude Sees",
			Skills: "what is actually in context (the conversation, files it read, command output) and what is not, pointing Claude at the right files instead of hoping, when to /clear or start a fresh session, why very long sessions degrade and what to do about it"},
		{ID: 5, Name: "Effective Task Requests",
			Skills: "scoping a task to one reviewable change, stating goal + constraints + done criteria, giving the why behind a request, choosing small steps over big-bang asks, steering with follow-ups instead of restarting"},
		{ID: 6, Name: "Reviewing & Verifying Output",
			Skills: "reading a diff before accepting it, using tests and builds as the arbiter of done, spotting plausible-but-wrong code, asking Claude to explain or justify a change, pushing back and course-correcting when it goes off track"},
		{ID: 7, Name: "CLAUDE.md & Project Memory",
			Skills: "what CLAUDE.md is and when Claude reads it, writing rules that actually change behavior (concrete and imperative, not vague), project conventions vs one-off task instructions, keeping it short enough to be followed, checked-in project memory vs personal notes"},
		{ID: 8, Name: "Plan Mode & Working in Steps",
			Skills: "when to plan before letting Claude edit, entering and exiting plan mode, reviewing and revising a plan before approving it, todo lists for multi-step work, breaking a large feature into sessions"},
		{ID: 9, Name: "Slash Commands & Skills",
			Skills: "built-in slash commands, writing a custom command in .claude/commands as a markdown file, passing arguments, skills in .claude/skills with a SKILL.md, recognising when a prompt you keep retyping should become a command"},
		{ID: 10, Name: "Settings, Permissions & Hooks",
			Skills: "settings.json and its scopes (user, project, local), permission allow/deny rules to reduce prompting, what hooks are: shell commands run on lifecycle events, a simple hook as a guardrail (e.g. block edits to a protected file) or automation (e.g. format after edit)"},
		{ID: 11, Name: "Subagents & Parallel Work",
			Skills: "what a subagent is: a fresh context doing delegated work, when delegation helps (independent or parallel subtasks) and when it hurts (shared context, tiny tasks), defining a custom agent in .claude/agents with its own prompt and tools, the cost of re-establishing context"},
		{ID: 12, Name: "Headless Mode & Scripting",
			Skills: "claude -p for one-shot non-interactive runs, piping input in and capturing output, JSON output for machine consumption, using Claude Code inside shell scripts and CI jobs, why automation needs stricter permissions and verification than interactive use"},
		{ID: 13, Name: "The Claude API",
			Skills: "Claude Code vs the raw API and when each fits, API keys and the developer console, the Messages endpoint: model, max_tokens, system prompt, user/assistant roles, reading the response and stop reason, a first request with curl"},
	},
	Tracks:   claudeTracks,
	Projects: claudeProjects,
}

// ── Git, GitHub & CI/CD ───────────────────────────────

var gitLanguage = Language{
	ID:          "git",
	Name:        "Git & CI/CD",
	Icon:        "/static/icons/git.svg",
	Cmd:         "$ git status",
	AccentColor: "#F05033",
	AccentDark:  "#BC3A26",
	AccentGlow:  "rgba(240,80,51,0.15)",
	CodeLabel:   "GIT",
	StyleNote:   "One targeted tip on commit hygiene: message quality (imperative subject line, why over what), keeping commits atomic, sensible branch names, or keeping workflow YAML minimal and readable.",
	StarterTemplate: "```bash\n# Your commands here — or replace this block with workflow YAML\n# or a written answer, whichever the challenge asks for\n```",
	SystemPrompt: `You are an expert instructor teaching code management with Git, GitHub, and CI/CD, running a hands-on bootcamp.
You teach clearly, practically, and engagingly. You are patient and encouraging.
Format all responses in Markdown. Fence every example: bash fences for terminal commands and their output, yaml fences for GitHub Actions workflows, text fences for file trees, diffs, and conflict markers.
Be concise but thorough. Include practical, real-world examples.
This course teaches a craft, not a programming language: student submissions are command sequences, commit messages, workflow YAML, and written explanations of what they would do and why. Evaluate them on correctness, safety (could this lose work or break a shared branch?), and whether the reasoning is sound — commands cannot be executed here, so your judgment is the feedback loop.
Assume the student works in a terminal on macOS or Linux with git installed and a free GitHub account. GitHub's UI and Actions evolve quickly; teach durable concepts and commands, and where a menu name or YAML detail may have changed since your knowledge was current, say so and point the student to the official docs rather than guessing.`,
	// The sequence is deliberately local-first: topics 1-6 need no GitHub
	// account (branches and even merge conflicts happen locally), 7-11 add
	// sharing and collaboration, and 12-15 automate it. Undoing Things sits
	// right after the first commits because fear of losing work is the biggest
	// beginner blocker — knowing the reflog safety net changes how bravely
	// everything after it is practiced. CI/CD waits until pull requests and
	// team workflow exist, because a status check has no meaning until there
	// is a merge for it to block.
	Topics: []Topic{
		{ID: 1, Name: "Version Control Concepts",
			Skills: "what version control solves (history, undo, collaboration, backup), a repository as a series of commits (snapshots, not diffs), the working tree vs the repository, distributed vs centralized — every clone is a full copy, where git ends and GitHub begins; installing git, one-time setup with git config --global user.name / user.email, getting help with git help"},
		{ID: 2, Name: "Your First Repository",
			Skills: "git init, the three areas: working tree → staging area (git add) → history (git commit), git status as the constant companion, writing a commit message with -m, viewing history with git log, ignoring generated files with .gitignore, git rm and git mv"},
		{ID: 3, Name: "Reading History & Diffs",
			Skills: "git log --oneline / -p / --stat / --graph and limiting it by count or path, git diff between working tree, staging area, and commits, git show for one commit, commit hashes and short hashes, HEAD and the ~ / ^ ancestry syntax, git blame for who-last-touched-this-line"},
		{ID: 4, Name: "Undoing Things",
			Skills: "discarding working-tree changes with git restore, unstaging with git restore --staged, fixing the last commit with git commit --amend, git revert as the safe public undo, git reset --soft / --mixed / --hard and when each is safe, the reflog as the safety net that makes almost anything recoverable"},
		{ID: 5, Name: "Branches & Merging",
			Skills: "a branch as a movable pointer to a commit, git branch and git switch (-c to create), why branches are cheap and enable fearless experiments, merging with git merge, fast-forward vs merge commits (--no-ff), listing branches and deleting merged ones"},
		{ID: 6, Name: "Merge Conflicts",
			Skills: "why conflicts happen (the same lines changed on both branches), reading conflict markers (<<<<<<< / ======= / >>>>>>>), resolving by editing then git add and completing the merge, bailing out with git merge --abort, keeping conflicts small with short-lived branches and frequent merges"},
		{ID: 7, Name: "Remotes & GitHub",
			Skills: "creating a repository on GitHub, git clone, remotes and origin, git remote -v, publishing a branch with git push -u origin, git fetch vs git pull, HTTPS vs SSH authentication and setting up an SSH key, the README as the repository's front page"},
		{ID: 8, Name: "Pull Requests & Code Review",
			Skills: "the feature-branch workflow: branch → push → open a pull request, writing a PR title and description that helps the reviewer, draft PRs, reviewing: line comments, approving, requesting changes, responding to review with follow-up commits, the three merge methods (merge commit / squash / rebase) and when each fits, linking issues with closes #N"},
		{ID: 9, Name: "Team Workflows",
			Skills: "trunk-based development vs long-lived release branches (GitFlow) and why small short-lived branches win, keeping a feature branch current: merging main in vs rebasing onto main, protected branches and required reviews, CODEOWNERS, the fork-and-PR model for open source, PR and issue templates in .github"},
		{ID: 10, Name: "Rewriting History",
			Skills: "interactive rebase with git rebase -i: reword, squash, fixup, drop, and reordering commits, cherry-pick to copy a commit between branches, force-pushing a rebased branch with --force-with-lease, the golden rule: never rewrite history others may have pulled, cleaning up a messy branch before opening the PR"},
		{ID: 11, Name: "Tags & Releases",
			Skills: "semantic versioning (major.minor.patch) and what each bump promises users, lightweight vs annotated tags (git tag -a), pushing tags, checking out a tag, GitHub Releases with release notes and attached files, maintaining a changelog humans can actually read"},
		{ID: 12, Name: "CI/CD Concepts",
			Skills: "what continuous integration is: build and test every change, fail fast; continuous delivery vs continuous deployment, the anatomy of a pipeline: trigger → jobs → artifacts, why automation beats checklists (repeatable, reviewable, blocking), status checks on PRs as the quality gate, where GitHub Actions sits among CI providers"},
		{ID: 13, Name: "First GitHub Actions Workflow",
			Skills: "workflow files live in .github/workflows, YAML anatomy: name, on (push / pull_request), jobs, runs-on, steps, uses for actions vs run for commands, actions/checkout and the setup-* actions, reading logs in the Actions tab and re-running failed jobs, a status badge in the README"},
		{ID: 14, Name: "Workflows in Practice",
			Skills: "triggers in depth: branch and path filters, workflow_dispatch for manual runs, schedule with cron; a job matrix across versions or operating systems, caching dependencies to speed up runs, encrypted secrets and why credentials never go in the repo, uploading artifacts, making a workflow a required status check so a red build blocks the merge"},
		{ID: 15, Name: "Deploying with Pipelines",
			Skills: "a deploy job that runs only after tests pass (needs), deploying on merge to main vs on a tag push, GitHub Environments with protection rules and manual approval, environment-scoped secrets, rollback by reverting the commit or redeploying the previous version, a first look at safer rollouts: blue-green and canary in brief"},
	},
	Tracks:   gitTracks,
	Projects: gitProjects,
}
