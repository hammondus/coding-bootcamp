package main

// ── Go tracks ─────────────────────────────────────────
// To add a new Go track: append a Track{} block to goTracks.
// To add a track for a new language: create a new var here and assign it
// in the Language definition in languages.go.

// goTracks are listed with prerequisites first: Standard Library leads because
// HTTP (JSON bodies) and two other tracks build on it.
var goTracks = []Track{
	{
		ID:          "standard-library",
		Title:       "Standard Library",
		Icon:        "📚",
		Description: "Master the most useful packages in Go's rich standard library",
		Lessons: []TrackLesson{
			{1, "strings & strconv", "String manipulation, strings.Builder, and type conversion with strconv"},
			{2, "io, bufio & os", "Readers, writers, buffered I/O, and file operations"},
			{3, "time & context", "Time parsing and formatting, context cancellation, and deadlines"},
			{4, "encoding/json", "Marshaling, unmarshaling, struct tags, custom marshalers, and streaming"},
			{5, "sync & atomic", "Mutex, RWMutex, WaitGroup, Once, and low-level atomic operations"},
		},
	},
	{
		ID:          "http",
		Title:       "HTTP",
		Icon:        "🌐",
		Description: "Build robust HTTP clients and servers, from simple requests to production-ready APIs",
		Prereqs:     "the Standard Library track, lessons 1-4 (JSON request and response bodies appear from lesson 2 onward and encoding/json is taught there)",
		Lessons: []TrackLesson{
			{1, "Simple GET Requests", "Using http.Get to fetch data, reading the response body, and checking status codes"},
			{2, "POST Requests & Bodies", "Sending data with http.Post, constructing JSON bodies, and http.NewRequest"},
			{3, "Custom HTTP Client", "Configuring http.Client with timeouts, custom transports, and redirect policies"},
			{4, "Headers & Authentication", "Setting request headers, Bearer tokens, Basic auth, and API keys"},
			{5, "Simple HTTP Server", "http.HandleFunc, http.ListenAndServe, writing responses and reading request data"},
			{6, "Custom ServeMux & Routing", "http.NewServeMux, route patterns, path parameters, and query strings"},
			{7, "Middleware & Handler Chains", "Writing middleware, chaining handlers, logging, recovery, and auth middleware"},
			{8, "Project: Structured REST API", "A complete JSON REST API split across multiple files with proper separation of concerns"},
		},
	},
	{
		// htmx sits directly after the HTTP track because every lesson writes Go
		// handlers. Lesson order: html/template must come first (fragments are
		// the response format for everything after), targeting/swapping before
		// triggers (trigger examples are meaningless until the student controls
		// where responses land), and Search & Tables introduces no new
		// attributes — it deliberately consolidates lessons 3-5 in combination.
		ID:    "htmx",
		Title: "Web Apps with htmx",
		Icon:  "🔁",
		Description: "Build dynamic, server-rendered web apps in Go with htmx — no JavaScript framework, no build step. " +
			"Complete the Go fundamentals, the HTTP track, and the HTML fundamentals (Frontend section) first: this track builds on all three and won't re-teach those basics.",
		Prereqs: "the HTTP track, lessons 5-7 (every lesson here writes Go handlers, routing, and middleware), and the HTML fundamentals course from the Frontend section, especially Document Structure and Forms & Inputs (htmx enhances plain HTML forms and links, and these lessons never re-teach them). Basic CSS helps but is not required. html/template is taught in lesson 1 — assume no prior template experience",
		Lessons: []TrackLesson{
			{1, "Hypermedia & html/template", "The HTML-over-the-wire philosophy vs JSON APIs, rendering pages and reusable fragments with html/template, and how contextual auto-escaping keeps output safe"},
			{2, "First htmx Requests", "Adding htmx with one script tag, hx-get and hx-post on buttons and links, and Go handlers that return HTML fragments instead of JSON"},
			{3, "Targeting & Swapping", "hx-target with CSS selectors, hx-swap strategies (innerHTML, outerHTML, beforeend, delete), and structuring Go templates as swappable fragments"},
			{4, "Triggers & Polling", "hx-trigger beyond the defaults: event names, modifiers (once, changed, delay, throttle), load and revealed, and polling with every"},
			{5, "Forms & Validation", "Submitting forms with hx-post, reading fields with r.FormValue, validating on the server, and re-rendering the form fragment with inline error messages"},
			{6, "Search & Table Patterns", "Combining triggers, targets, and forms: active search with a debounced trigger, sortable and filterable tables, and infinite scroll with revealed"},
			{7, "Loading States & UX", "hx-indicator and the htmx-request class, disabling controls with hx-disabled-elt, confirming destructive actions with hx-confirm"},
			{8, "Out-of-Band Swaps & Headers", "Updating multiple page regions from one response with hx-swap-oob, detecting htmx with the HX-Request header, and driving the client with HX-Redirect and HX-Trigger response headers"},
			{9, "URLs, History & Boosting", "hx-push-url and back-button behavior, upgrading plain links and forms with hx-boost, and serving a full page or a fragment from the same Go handler"},
			{10, "Project: Server-Rendered Task Manager", "A complete Go + htmx CRUD app: task list with active search, inline editing, out-of-band counter updates, and working browser history"},
		},
	},
	{
		ID:          "project-structure",
		Title:       "Project Structure",
		Icon:        "📁",
		Description: "Idiomatic Go project organisation, from simple packages to layered architectures",
		Lessons: []TrackLesson{
			{1, "Go Modules & Workspaces", "go.mod, module paths, go get, go work, and the module cache"},
			{2, "Packages & Visibility", "Package design, exported vs unexported identifiers, and naming conventions"},
			{3, "Internal Packages", "The internal directory, restricting package access, and clean API boundaries"},
			{4, "Layered Architecture", "Separating concerns: handlers, services, repositories, models — and why it matters"},
		},
	},
	{
		ID:          "interfaces-deep",
		Title:       "Interfaces Deep Dive",
		Icon:        "🔷",
		Description: "Go beyond the basics — interface composition, key standard interfaces, and design patterns",
		Lessons: []TrackLesson{
			{1, "Interface Composition", "Embedding interfaces, building from small pieces, and interface segregation"},
			{2, "Type Assertions & Switches", "Runtime type checking, type switch patterns, and the comma-ok idiom"},
			{3, "io.Reader & io.Writer", "The most fundamental interfaces in Go and how to chain them with wrappers"},
			{4, "Designing with Interfaces", "Accept interfaces/return structs, dependency injection, and testable code"},
		},
	},
	{
		ID:          "generics",
		Title:       "Generics",
		Icon:        "⚙️",
		Description: "Write reusable, type-safe code with Go generics",
		Lessons: []TrackLesson{
			{1, "Type Parameters", "Generic function syntax, constraints, and the built-in any and comparable"},
			{2, "Generic Functions", "Map, Filter, Reduce, and other useful generic collection helpers"},
			{3, "Generic Types", "Generic structs, methods on generic types, and type inference"},
			{4, "Constraints & Patterns", "Custom type constraints, union types, and real-world patterns"},
		},
	},
	{
		ID:          "testing",
		Title:       "Testing",
		Icon:        "🧪",
		Description: "Build confidence in your code with Go's testing tools",
		Prereqs:     "for lesson 5 only, the HTTP track lessons 5-7 (httptest exercises HTTP handlers); lessons 1-4 need nothing beyond fundamentals",
		Lessons: []TrackLesson{
			{1, "Unit Tests", "testing.T, t.Fatal vs t.Error, test helpers, and the go test command"},
			{2, "Table-Driven Tests", "Test cases as data, t.Run subtests, and naming conventions"},
			{3, "Mocking with Interfaces", "Using interfaces to isolate dependencies and write test doubles"},
			{4, "Benchmarks", "testing.B, b.ReportAllocs, avoiding compiler optimisations, and reading results"},
			{5, "httptest & Integration Tests", "httptest.NewRecorder, httptest.NewServer, and testing HTTP handlers end-to-end"},
		},
	},
	{
		ID:          "documentation",
		Title:       "Documentation",
		Icon:        "📖",
		Description: "Write documentation that delights users and makes your code self-explanatory",
		Prereqs:     "the Testing track, lesson 1 (testable examples run under go test)",
		Lessons: []TrackLesson{
			{1, "godoc Conventions", "Doc comment format, package-level docs, deprecation notices, and cross-links"},
			{2, "Testable Examples", "Example functions, Output comments, and how they appear on pkg.go.dev"},
			{3, "go doc & pkg.go.dev", "Using go doc locally, publishing modules, and documentation best practices"},
		},
	},
	{
		ID:          "profiling",
		Title:       "Profiling & Optimisation",
		Icon:        "⚡",
		Description: "Find and fix performance bottlenecks using Go's built-in profiling tools",
		Prereqs:     "the Testing track, lessons 1-2 (profiling is benchmark-driven); lesson 4 leans on the Goroutines & Channels fundamentals topic",
		Lessons: []TrackLesson{
			{1, "pprof Basics", "CPU and heap profiling, go tool pprof, and reading flame graphs"},
			{2, "Benchmarks & Measurement", "Writing meaningful benchmarks, -benchmem, and avoiding false measurements"},
			{3, "Optimisation Patterns", "Reducing allocations, string efficiency, sync.Pool, and cache-friendly layouts"},
			{4, "Race Detection", "The -race flag, common data race patterns, and fixing them with sync primitives"},
		},
	},
}

// ── Zig tracks ────────────────────────────────────────

var zigTracks = []Track{
	{
		ID:          "http",
		Title:       "HTTP",
		Icon:        "🌐",
		Description: "Build HTTP clients and servers using Zig's standard library",
		Lessons: []TrackLesson{
			{1, "HTTP Client Basics", "Using std.http.Client to make GET requests, reading response bodies"},
			{2, "POST Requests & Headers", "Sending data, custom headers, and handling different content types"},
			{3, "Simple HTTP Server", "std.http.Server, accepting connections, and writing responses"},
			{4, "JSON over HTTP", "Combining std.json with HTTP to build a simple data API — std.json itself (parse and stringify) is taught in this lesson, no prior JSON lesson exists"},
			{5, "Project: REST Client & Server", "A complete HTTP application with structured client and server components"},
		},
	},
	{
		ID:          "build-system",
		Title:       "Build System",
		Icon:        "🔧",
		Description: "Master zig build — the powerful, programmable build system at the heart of Zig",
		Lessons: []TrackLesson{
			{1, "build.zig Fundamentals", "Build steps, executables, libraries, and the build graph"},
			{2, "Modules & Dependencies", "zig fetch, build.zig.zon, and consuming external packages"},
			{3, "Cross-Compilation", "Target triples, CPU features, and building for multiple platforms from one machine"},
			{4, "Build Options & Modes", "Debug vs ReleaseFast vs ReleaseSafe, custom options, and conditional compilation"},
		},
	},
	{
		ID:          "testing",
		Title:       "Testing",
		Icon:        "🧪",
		Description: "Test your Zig code thoroughly using the built-in testing framework",
		Lessons: []TrackLesson{
			{1, "Unit Tests with std.testing", "expect, expectEqual, expectError, and testing with allocators"},
			{2, "Test Organisation", "Test blocks in source files, test filters, and zig test options"},
			{3, "Fuzzing", "std.testing.fuzz, corpus-based testing, and finding edge cases automatically"},
		},
	},
	{
		ID:          "comptime-deep",
		Title:       "Comptime Deep Dive",
		Icon:        "🔮",
		Description: "Unlock the full power of Zig's compile-time metaprogramming",
		Lessons: []TrackLesson{
			{1, "Types as Values", "@TypeOf, @typeInfo, and working with types at compile time"},
			{2, "Comptime Functions", "Generic functions with anytype, comptime parameters, and specialisation"},
			{3, "Comptime Interfaces", "Duck typing at compile time and comptime polymorphism patterns"},
			{4, "Reflection Patterns", "@field, struct field iteration, and building data-driven code"},
		},
	},
	{
		ID:          "memory",
		Title:       "Memory Management",
		Icon:        "🧠",
		Description: "Take full control of memory: allocators, ownership, and safety patterns",
		Lessons: []TrackLesson{
			{1, "Allocator Interface Deep Dive", "The allocator contract, vtable dispatch, and why all allocators are interchangeable"},
			{2, "Arena & Fixed-Buffer Allocators", "Efficient short-lived allocations, stack allocation, and batch-free patterns"},
			{3, "Ownership & Safety Patterns", "defer free, errdefer, ownership rules, and preventing leaks"},
			{4, "Building a Custom Allocator", "Implementing the allocator interface and tracking allocations for debugging"},
		},
	},
}

// ── JavaScript tracks ─────────────────────────────────
// These four tracks stand alone, but together they ramp toward the Moon
// Patrol Ghosts capstone (projects.go): canvas drawing, then game mechanics,
// then the browser plumbing a game needs, then real-time networking for the
// ghost milestones.

var javascriptTracks = []Track{
	{
		ID:          "canvas",
		Title:       "Canvas & Animation",
		Icon:        "🎨",
		Description: "Draw and animate with the 2D canvas — the foundation of browser games and visualisations",
		Lessons: []TrackLesson{
			{1, "Canvas Setup & Coordinates", "The canvas element, getContext('2d'), the pixel coordinate system, and canvas size vs CSS display size"},
			{2, "Drawing Shapes & Paths", "fillRect and strokeRect, beginPath, lines, arcs and circles, fill and stroke styles"},
			{3, "Transforms & State", "translate, rotate, scale, save/restore, and drawing relative to a moving origin"},
			{4, "Text, Gradients & Images", "fillText for HUDs and labels, linear gradients for skies, and drawImage for sprites"},
			{5, "Animating with requestAnimationFrame", "The browser's frame callback, clearing and redrawing each frame, and time-based motion with delta time"},
		},
	},
	{
		ID:          "game-dev",
		Title:       "Game Development",
		Icon:        "🎮",
		Description: "The mechanics every 2D game is built from: the loop, input, physics, collision, and cameras",
		Prereqs:     "the Canvas & Animation track (lesson 1 builds directly on requestAnimationFrame and canvas drawing from Canvas lessons 1-5)",
		Lessons: []TrackLesson{
			{1, "The Game Loop", "requestAnimationFrame, delta time, and separating a fixed-timestep update from rendering"},
			{2, "Keyboard Input State", "Tracking held keys with keydown/keyup and a key-state object, and why event handlers alone aren't enough for games"},
			{3, "Velocity, Gravity & Jumping", "Position, velocity and acceleration, applying gravity, and tuning a jump arc that feels right"},
			{4, "Collision Detection", "Axis-aligned bounding boxes, circle overlap, and point-vs-terrain tests"},
			{5, "Scrolling & Parallax", "World vs screen coordinates, a side-scrolling camera, and background layers moving at different speeds"},
			{6, "Game States & Entities", "A state machine for title/playing/game-over, and managing arrays of entities that spawn and die"},
			{7, "Project: One-Screen Arcade Game", "A complete playable arcade game (Breakout or similar) bringing together the loop, input, physics, collision, and state"},
		},
	},
	{
		ID:          "browser-apis",
		Title:       "Browser APIs",
		Icon:        "🧰",
		Description: "The browser platform beyond the DOM: timing, storage, sound, and the page lifecycle",
		Lessons: []TrackLesson{
			{1, "Timers & Scheduling", "setTimeout and setInterval vs requestAnimationFrame, debouncing, and throttling"},
			{2, "localStorage", "Persisting settings and high scores as JSON, storage limits, and versioning stored data"},
			{3, "Sound & Audio", "Playing sound effects with the Audio element, overlapping playback, and a taste of the Web Audio API"},
			{4, "Page Visibility & Lifecycle", "Pausing loops when the tab is hidden, the visibilitychange event, and being a good citizen of the browser"},
		},
	},
	{
		ID:          "realtime",
		Title:       "Real-Time & Networking",
		Icon:        "📡",
		Description: "From request/response to live data: fetch in depth, streaming, and WebSockets",
		Lessons: []TrackLesson{
			{1, "Fetch in Depth", "Request options and methods, status and error handling, JSON round-trips, and AbortController"},
			{2, "Server-Sent Events", "The EventSource API and reading streamed responses — one-way real-time from server to browser"},
			{3, "WebSockets", "The WebSocket API: connecting, designing a JSON message protocol, and clean close and error handling"},
			{4, "Real-Time Patterns", "Heartbeats, reconnecting with backoff, and smoothing remote positions with interpolation"},
		},
	},
}

// ── Git tracks ────────────────────────────────────────
// Track order mirrors demand and dependency: Actions in Depth directly
// extends fundamentals topics 13-15 and is the skill teams ask for most;
// Advanced Git can be taken any time after the fundamentals; Release
// Engineering closes the loop from merged code to shipped software and
// leans on both.

var gitTracks = []Track{
	{
		ID:          "actions-in-depth",
		Title:       "GitHub Actions in Depth",
		Icon:        "⚙️",
		Description: "Go from using Actions to engineering them: expressions, reusable workflows, custom actions, security, and runners",
		Prereqs:     "the Git & CI/CD fundamentals course, especially First GitHub Actions Workflow, Workflows in Practice, and Deploying with Pipelines (topics 13-15) — this track deepens all three rather than re-teaching them",
		Lessons: []TrackLesson{
			{1, "Contexts & Expressions", "The ${{ }} expression syntax, the github / env / secrets / needs contexts, if: conditions on jobs and steps, and passing data between steps and jobs with outputs"},
			{2, "Reusable Workflows & Composite Actions", "DRY for pipelines: workflow_call with inputs and secrets, composite actions for repeated step sequences, and choosing between the two"},
			{3, "Writing a Custom Action", "An action of your own: action.yml metadata, inputs and outputs, implementing it as a JavaScript or Docker action, and versioning it with tags so others can pin it"},
			{4, "Actions Security", "Least-privilege GITHUB_TOKEN permissions, pinning third-party actions to a commit SHA, the pull_request_target trap, script injection through untrusted inputs, and OIDC to cloud providers instead of long-lived secrets"},
			{5, "Runners & Performance", "GitHub-hosted vs self-hosted runners, labels and runner groups, concurrency groups with cancel-in-progress, and cutting billable minutes with caching and path filters"},
			{6, "Debugging Workflows", "Reading raw logs, enabling step debug logging, re-running jobs with debug output, testing workflows locally, and instrumenting a flaky pipeline to find what is actually failing"},
		},
	},
	{
		ID:          "advanced-git",
		Title:       "Advanced Git",
		Icon:        "🌿",
		Description: "How git actually works — objects, refs, and the plumbing — and the power tools that knowledge unlocks",
		Prereqs:     "the Git & CI/CD fundamentals course, especially Undoing Things and Rewriting History (topics 4 and 10) — reset, rebase, and the reflog are assumed known and get explained one level deeper here",
		Lessons: []TrackLesson{
			{1, "The Object Model", "Blobs, trees, commits, and tags inside .git/objects, content addressing by SHA, why commits really are snapshots, and inspecting objects with git cat-file"},
			{2, "Refs, HEAD & the Reflog", "Branches as small files in .git/refs, symbolic refs, detached HEAD explained properly, and using the reflog to recover commits and branches that look lost"},
			{3, "Stash & Worktrees", "git stash beyond the basics: stashing partial changes, pop vs apply, and git worktree for working on two branches at once without a second clone"},
			{4, "Finding Things: bisect & pickaxe", "git bisect to binary-search history for the commit that broke it (including automating with bisect run), git log -S / -G to find when code appeared or vanished, and git grep"},
			{5, "Hooks & Automation", "Client-side hooks (pre-commit, commit-msg, pre-push): what they can and cannot enforce, sharing hooks with a team, and where server-side enforcement — protected branches and CI — has to take over"},
			{6, "Submodules & Subtrees", "Nesting repositories: how a submodule really works (a pinned commit pointer), its sharp edges, subtree as an alternative, and when plain package management beats both"},
			{7, "Disaster Recovery", "Un-losing work: recovering from a bad reset or rebase with the reflog, undoing a force-push, moving a commit made on the wrong branch, and purging a leaked secret from history — and why rotating it still matters"},
		},
	},
	{
		ID:          "release-engineering",
		Title:       "Release Engineering",
		Icon:        "🚀",
		Description: "Turn merged code into shipped software: versioning, automated releases, deployment strategies, and keeping main always releasable",
		Prereqs:     "the Git & CI/CD fundamentals course, especially Tags & Releases and Deploying with Pipelines (topics 11 and 15); the GitHub Actions in Depth track helps with the automation lessons but is not required",
		Lessons: []TrackLesson{
			{1, "Versioning Strategies", "Semantic versioning in practice, conventional commit messages as machine-readable history, pre-release and build metadata, and how versioning an application differs from versioning a library"},
			{2, "Automated Release Pipelines", "A tag-triggered pipeline that builds artifacts, generates release notes from the commit history, and publishes a GitHub Release — no human steps between pushing the tag and the finished release"},
			{3, "Deployment Strategies", "Rolling, blue-green, and canary deployments, feature flags to decouple deploying code from releasing features, and choosing a strategy by blast radius and rollback speed"},
			{4, "Environments & Promotion", "Dev → staging → production promotion pipelines, environment protection rules and manual approvals, environment-scoped secrets and configuration, and smoke tests as the gate between stages"},
			{5, "Keeping Main Releasable", "Trunk-based development in practice: small PRs, feature flags instead of release branches, a revert-first culture when main breaks, and release trains vs ship-on-green"},
		},
	},
}

// ── Claude tracks ─────────────────────────────────────
// Track order mirrors how the skills build: prompt craft applies everywhere
// and has no prerequisites, so it leads. Agentic Workflows and Extending
// Claude Code both deepen the fundamentals' automation topics (8-12). The
// API track closes because it assumes another language (Go) for its
// examples — it is the bridge from using Claude to building with it.

var claudeTracks = []Track{
	{
		ID:          "prompt-engineering",
		Title:       "Prompt Engineering",
		Icon:        "✍️",
		Description: "Go beyond good-enough prompts: structure, examples, output control, long context, and knowing whether a change actually helped",
		Lessons: []TrackLesson{
			{1, "Anatomy of a Great Prompt", "Role, context, task, constraints, and output format as distinct parts; why instructions go before data; rewriting a vague ask into a precise one"},
			{2, "Examples & Few-Shot Prompting", "Showing instead of telling: choosing representative examples, formatting them with clear delimiters (XML tags), and how one good example beats three paragraphs of description"},
			{3, "Structured Output", "Getting JSON, tables, or a fixed schema back reliably: describing the shape, providing a filled-in example, and validating and re-asking when the output doesn't conform"},
			{4, "Long-Context Strategies", "Working with big inputs: instruction placement, quoting and grounding answers in the source, asking for citations, and summarize-then-work for documents too large to reason over directly"},
			{5, "Evaluating Prompts", "Treating prompts as code: building a small eval set of inputs with expected outputs, comparing prompt variants against it, and catching regressions when tweaking a prompt that already works"},
		},
	},
	{
		ID:          "agentic-workflows",
		Title:       "Agentic Workflows",
		Icon:        "🤖",
		Description: "Design multi-step, delegated, and automated workflows where Claude works with less supervision — and verification replaces watching",
		Prereqs:     "the Claude fundamentals course, especially Plan Mode & Working in Steps, Subagents & Parallel Work, and Headless Mode & Scripting (topics 8, 11, 12) — this track deepens all three rather than re-teaching them",
		Lessons: []TrackLesson{
			{1, "Thinking in Agent Loops", "What makes work agentic: the gather-context → act → verify loop, why a verification step beats a better first attempt, and matching autonomy to the cost of a mistake"},
			{2, "Designing Multi-Step Tasks", "Decomposing a feature into ordered, independently checkable steps with explicit done criteria, deciding where human review belongs, and writing the kickoff prompt that carries the whole plan"},
			{3, "Custom Subagents", "Writing agent definitions in .claude/agents: a focused system prompt, a restricted tool set, and a model choice per agent — single-responsibility agents like a reviewer, a test-writer, or a researcher"},
			{4, "Parallel & Background Work", "Fanning independent subtasks out to subagents in one turn, running long work in the background while continuing, and isolating risky work in a git worktree"},
			{5, "Hooks as Guardrails & Automation", "Hooks that enforce rules no prompt can guarantee: blocking dangerous commands, auto-formatting after edits, requiring tests before a commit, and notifications when long work finishes"},
			{6, "Headless Pipelines", "claude -p in scripts and CI: structured JSON output, strict permission flags for unattended runs, and pipeline patterns like issue triage, changelog drafting, and scheduled maintenance"},
			{7, "Project: Automated Code Review", "A complete pipeline: a reviewer subagent with a written rubric, run headlessly by CI against each change, posting findings as a report — with a human still owning the merge"},
		},
	},
	{
		ID:          "extending",
		Title:       "Extending Claude Code",
		Icon:        "🧩",
		Description: "Package your workflows as commands, skills, hooks, and MCP servers so the whole team gets them for free",
		Prereqs:     "the Claude fundamentals course, especially Slash Commands & Skills and Settings, Permissions & Hooks (topics 9-10); lesson 5 writes a small program and assumes basic programming in any language",
		Lessons: []TrackLesson{
			{1, "Custom Slash Commands", "Commands as markdown files in .claude/commands: arguments, frontmatter, project vs personal scope, and turning your three most-retyped prompts into commands"},
			{2, "Building Skills", "A skill as a folder with a SKILL.md: writing the description that makes it trigger at the right moment, supporting files, and keeping the main file short with detail loaded on demand"},
			{3, "Hooks in Depth", "The hook lifecycle events and their matchers, reading tool input as JSON and blocking or allowing with exit codes, and debugging a hook that isn't firing"},
			{4, "MCP: Connecting External Tools", "What the Model Context Protocol is, adding existing MCP servers to give Claude new tools (databases, browsers, issue trackers), configuration scopes, and the trust decisions that come with third-party tools"},
			{5, "Building an MCP Server", "A minimal MCP server of your own: exposing one custom tool over stdio, defining its input schema, and watching Claude discover and call it"},
			{6, "Project: A Team Toolkit", "A checked-in .claude/ directory for a real repository: commands, a skill, guardrail hooks, and MCP config that a new teammate gets working on day one"},
		},
	},
	{
		ID:    "claude-api",
		Title: "The Claude API in Go",
		Icon:  "🔌",
		Description: "Call Claude from your own programs: the Messages API, streaming, conversations, tool use, and production robustness — in Go with only the standard library. " +
			"Complete the Go fundamentals and the Go HTTP track first: every lesson writes Go against a real HTTP API.",
		Prereqs: "the Claude fundamentals course topic 13 (The Claude API — the endpoint and roles are introduced there), Go fundamentals, and the Go HTTP track lessons 1-4 (clients, JSON bodies, headers and auth); Standard Library lesson 4 (encoding/json) helps but the structs used are shown in full",
		Lessons: []TrackLesson{
			{1, "Messages in Go", "Request and response structs for the Messages endpoint, sending a system prompt and user message with net/http, and handling the API's error responses by status code"},
			{2, "Streaming Responses", "stream: true and Server-Sent Events: parsing the event stream line by line with bufio, printing text deltas as they arrive, and detecting a clean finish vs a truncated one"},
			{3, "Multi-Turn Conversations", "Managing the messages array as conversation state, alternating roles correctly, and coping with context growth: trimming old turns and summarizing history"},
			{4, "Tool Use", "Defining a tool with a JSON schema, the stop_reason tool_use loop — execute the tool, append the result, call again — and handling several tool calls in one response"},
			{5, "Robustness & Cost", "Timeouts and cancellation with context, bounded retry with backoff on 429/529/5xx, counting tokens to predict cost, and caching repeated prompt prefixes"},
		},
	},
}
