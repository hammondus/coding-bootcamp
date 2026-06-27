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
	Tracks          []Track // advanced multi-lesson tracks; nil = no tracks for this language
}

// languageOrder controls the display order in the UI switcher.
var languageOrder = []string{"go", "zig"}

var languages = map[string]Language{
	"go":  goLanguage,
	"zig": zigLanguage,
}

// ── Go ────────────────────────────────────────────────

var goLanguage = Language{
	ID:          "go",
	Name:        "Go",
	Icon:        "🐹",
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
	Tracks: goTracks,
}

// ── Zig ───────────────────────────────────────────────

var zigLanguage = Language{
	ID:          "zig",
	Name:        "Zig",
	Icon:        "⚡",
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
