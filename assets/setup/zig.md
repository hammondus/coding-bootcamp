# Getting Started with Zig

Welcome! Before the first lesson, it's worth knowing **where to actually run
Zig code**. The editor built into this bootcamp doesn't execute your code —
your AI instructor reads it and gives feedback on the logic. Running your code
for real is where the learning sticks.

Zig is a young language (pre-1.0), so its browser playgrounds are thinner than
Go's or JavaScript's — installing it locally is the recommended path, and it's
a small, quick install.

## Try Zig in your browser — no install

- **[godbolt.org](https://godbolt.org)** (Compiler Explorer) — pick **Zig** as
  the language. It's aimed at inspecting compiled output, but the *Execute*
  view runs your program and shows its output, which is all you need for the
  early lessons.

Because playground support is limited, don't stay here long — do the local
install below early.

## Install Zig on your computer

1. Download a prebuilt package for your OS from
   **[ziglang.org/download](https://ziglang.org/download/)**, or use a package
   manager (`brew install zig` on macOS, `winget install zig.zig` on Windows).
2. Check it worked:

   ```
   zig version
   ```

3. Run a single file — for early lessons this is all you need. Create
   `main.zig`:

   ```zig
   const std = @import("std");

   pub fn main() void {
       std.debug.print("Hello from my own machine!\n", .{});
   }
   ```

   And run it:

   ```
   zig run main.zig
   ```

4. Later (the *Build System & Testing* topic), you'll create a real project:

   ```
   mkdir hello
   cd hello
   zig init
   zig build run
   ```

**A version warning:** Zig is still changing between releases, so a code
snippet from an old blog post may not compile on today's compiler. If something
from the wider internet fails mysteriously, check what Zig version it was
written for — and ask in the **💬 Ask** tab.

## Set up a code editor

Zig's language server is called **ZLS** — it gives you autocomplete, hover
docs, and inline errors in either editor below.

### VS Code

1. Install it from **[code.visualstudio.com](https://code.visualstudio.com)**.
2. Open the Extensions panel and install **Zig Language** (published by
   *ziglang*).
3. Open a `.zig` file — the extension will offer to install ZLS for you.
   Accept, and you're done.

### Zed

1. Install it from **[zed.dev](https://zed.dev)** — a newer, very fast editor.
2. Open the extensions list (`zed: extensions` in the command palette) and
   install the **Zig** extension. It downloads and manages ZLS for you.

## Where to look things up

- **[ziglang.org/documentation](https://ziglang.org/documentation/)** — the
  official language reference: every keyword and builtin (`@import`, `@as`, …)
  on one page. Use your browser's find-in-page liberally.
- **[Standard library docs](https://ziglang.org/documentation/master/std/)** —
  searchable reference for `std`: look up `std.debug.print`, `std.ArrayList`,
  `std.mem.eql`, and friends.
- **[zig.guide](https://zig.guide)** — a community tutorial that walks the same
  ground as this course; good for a second explanation of a tricky topic.
- **[Ziglings](https://codeberg.org/ziglings/exercises)** — tiny broken
  programs you fix one at a time. Excellent extra practice alongside the
  challenges here.

## Suggested path

1. **Today:** install Zig (it's quick) and get `zig run main.zig` working.
2. **Soon:** set up VS Code or Zed with ZLS so errors show up as you type —
   Zig's explicitness makes a language server especially valuable.
3. **When you reach Build System & Testing:** switch from single files to
   `zig init` projects.

Stuck on any step — an install problem, a confusing compile error, a
Windows-vs-Mac difference? Open the **💬 Ask** tab and describe what you're
seeing.
