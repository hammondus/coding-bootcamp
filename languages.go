package main

// Topic is a single lesson in a bootcamp curriculum.
type Topic struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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
	Lessons     []TrackLesson
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
	Milestones  []ProjectMilestone
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
}

var languages = map[string]Language{
	"go":         goLanguage,
	"zig":        zigLanguage,
	"javascript": javascriptLanguage,
	"html":       htmlLanguage,
	"css":        cssLanguage,
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
	Topics: []Topic{
		{ID: 1, Name: "Program Structure"},
		{ID: 2, Name: "Variables & Types"},
		{ID: 3, Name: "Control Flow"},
		{ID: 4, Name: "Loops"},
		{ID: 5, Name: "Functions"},
		{ID: 6, Name: "Arrays & Slices"},
		{ID: 7, Name: "Maps"},
		{ID: 8, Name: "Structs"},
		{ID: 9, Name: "Pointers"},
		{ID: 10, Name: "Methods"},
		{ID: 11, Name: "Interfaces"},
		{ID: 12, Name: "Error Handling"},
		{ID: 13, Name: "Goroutines & Channels"},
		{ID: 14, Name: "Closures & Defer"},
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
	Topics: []Topic{
		{ID: 1, Name: "Program Structure"},
		{ID: 2, Name: "Variables & Types"},
		{ID: 3, Name: "Control Flow"},
		{ID: 4, Name: "Loops"},
		{ID: 5, Name: "Functions"},
		{ID: 6, Name: "Arrays & Slices"},
		{ID: 7, Name: "Pointers"},
		{ID: 8, Name: "Structs"},
		{ID: 9, Name: "Enums & Unions"},
		{ID: 10, Name: "Error Handling"},
		{ID: 11, Name: "Optionals"},
		{ID: 12, Name: "Allocators & Memory"},
		{ID: 13, Name: "Comptime"},
		{ID: 14, Name: "Build System & Testing"},
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
	Topics: []Topic{
		{ID: 1, Name: "Variables & Types"},
		{ID: 2, Name: "Operators & Expressions"},
		{ID: 3, Name: "Control Flow"},
		{ID: 4, Name: "Loops"},
		{ID: 5, Name: "Functions & Arrow Functions"},
		{ID: 6, Name: "Arrays"},
		{ID: 7, Name: "Objects"},
		{ID: 8, Name: "Strings & Template Literals"},
		{ID: 9, Name: "Scope & Closures"},
		{ID: 10, Name: "The DOM"},
		{ID: 11, Name: "Events"},
		{ID: 12, Name: "Promises & Async/Await"},
		{ID: 13, Name: "Fetch & APIs"},
		{ID: 14, Name: "Modules"},
	},
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
	Topics: []Topic{
		{ID: 1, Name: "Document Structure"},
		{ID: 2, Name: "Text & Headings"},
		{ID: 3, Name: "Links & Navigation"},
		{ID: 4, Name: "Images & Media"},
		{ID: 5, Name: "Lists"},
		{ID: 6, Name: "Tables"},
		{ID: 7, Name: "Forms & Inputs"},
		{ID: 8, Name: "Semantic Elements"},
		{ID: 9, Name: "Attributes & Metadata"},
		{ID: 10, Name: "Accessibility"},
		{ID: 11, Name: "Embedding Content"},
		{ID: 12, Name: "Forms Validation"},
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
	Topics: []Topic{
		{ID: 1, Name: "Selectors"},
		{ID: 2, Name: "Colors & Backgrounds"},
		{ID: 3, Name: "The Box Model"},
		{ID: 4, Name: "Typography"},
		{ID: 5, Name: "Display & Layout"},
		{ID: 6, Name: "Flexbox"},
		{ID: 7, Name: "Grid"},
		{ID: 8, Name: "Positioning"},
		{ID: 9, Name: "Pseudo-classes & Pseudo-elements"},
		{ID: 10, Name: "Transitions & Animations"},
		{ID: 11, Name: "Transforms"},
		{ID: 12, Name: "Custom Properties"},
		{ID: 13, Name: "Responsive Design"},
	},
}
