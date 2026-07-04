package main

// ── Go tracks ─────────────────────────────────────────
// To add a new Go track: append a Track{} block to goTracks.
// To add a track for a new language: create a new var here and assign it
// in the Language definition in languages.go.

var goTracks = []Track{
	{
		ID:          "http",
		Title:       "HTTP",
		Icon:        "🌐",
		Description: "Build robust HTTP clients and servers, from simple requests to production-ready APIs",
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
			{4, "JSON over HTTP", "Combining std.json with HTTP to build a simple data API"},
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
