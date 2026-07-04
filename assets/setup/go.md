# Getting Started with Go

Welcome! Before the first lesson, it's worth knowing **where to actually run Go
code**. The editor built into this bootcamp doesn't execute your code — your AI
instructor reads it and gives feedback on the logic. Running your code for real
is where the learning sticks, so set up at least one of the options below.

They're ordered easiest-first. You can start with the browser playground right
now and come back for the local install later.

## Run Go in your browser — no install

- **[goplay.tools](https://goplay.tools)** — the one to use. A community-built
  playground with autocomplete, code formatting, and a nicer layout than the
  official one. Paste in any example from a lesson, change things, break them,
  and run again.
- **[go.dev/play](https://go.dev/play)** — the official Go playground. More
  basic, but it's what people usually mean by "the playground", and its share
  links are handy for showing code to others.

## Install Go on your computer

1. Download the installer for your OS from **[go.dev/dl](https://go.dev/dl/)**
   and run it.
2. Open a terminal and check it worked:

   ```
   go version
   ```

3. Create your first project — a folder with a *module* (Go's name for a
   project) and one source file:

   ```
   mkdir hello
   cd hello
   go mod init hello
   ```

   Then create `main.go`:

   ```go
   package main

   import "fmt"

   func main() {
       fmt.Println("Hello from my own machine!")
   }
   ```

   And run it:

   ```
   go run .
   ```

That `go run .` is the command shown under the logo in the sidebar — it means
"compile and run the module in this folder".

## Set up a code editor

Once you're writing more than a few lines, a real editor pays for itself:
autocomplete, docs on hover, and mistakes underlined as you type.

### VS Code

1. Install it from **[code.visualstudio.com](https://code.visualstudio.com)**.
2. Open the Extensions panel and install **Go** (published by *Go Team at
   Google*).
3. Open any `.go` file — VS Code will offer to install the Go tools it needs
   (including `gopls`, the Go language server). Click **Install All** and
   you're done.

### Zed

1. Install it from **[zed.dev](https://zed.dev)** — a newer, very fast editor.
2. Go support is built in. With Go installed on your machine, Zed sets up the
   language server for you the first time you open a `.go` file.

Either is a great choice: VS Code has more tutorials and answers written about
it; Zed is lighter and faster. Pick one — don't agonise.

## Where to look things up

- **[pkg.go.dev/std](https://pkg.go.dev/std)** — the standard library, fully
  documented. This is *the* reference: every function, with examples. Want all
  the `fmt` printing verbs? **[pkg.go.dev/fmt](https://pkg.go.dev/fmt)**.
  Everything for strings? **[pkg.go.dev/strings](https://pkg.go.dev/strings)**.
- **[go.dev/doc](https://go.dev/doc/)** — the official documentation hub,
  including the [Tour of Go](https://go.dev/tour/), an interactive intro that
  pairs nicely with these lessons.
- **[gobyexample.com](https://gobyexample.com)** — short, annotated examples of
  almost everything this course covers. Great for a second angle on a topic.

## Suggested path

1. **Today:** use [goplay.tools](https://goplay.tools) side-by-side with the
   lessons — run every example you read.
2. **Soon:** install Go locally and get `go run .` working once.
3. **When challenges get bigger:** pick VS Code or Zed and do your challenge
   work there, then paste your solution back here to submit it.

Stuck on any step — an installer that won't run, a `command not found`, a
Windows-vs-Mac difference? Open the **💬 Ask** tab and describe what you're
seeing.
