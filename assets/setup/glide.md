# Getting Started with Glide

Glide is a brand-new programming language being developed right here — the
compiler you'll use is built from source in the Glide repo. There are no
installers, package managers, or online playgrounds yet: you build the
interpreter once with Go, and you're set.

## What you need

- The **Glide repo** checked out locally (the bootcamp reads it from the
  path in `GLIDE_REPO`, default `../glide` next to this bootcamp).
- A **Go toolchain** (1.22+) to build the interpreter: `brew install go`
  on macOS, or [go.dev/dl](https://go.dev/dl) elsewhere.

## Build the interpreter

```bash
cd ../glide/glide     # the interpreter lives in glide/ inside the repo
make build            # produces bin/glide
```

Put it on your PATH, or shell-alias it:

```bash
alias glide="$PWD/bin/glide"
```

## Your first program

Glide source files use the `.gld` extension. Create `hello.gld`:

```glide
fn main() {
    let name = "world"
    println("hello, {name}")
}
```

Run it:

```bash
glide run hello.gld
```

No semicolons, mandatory braces, and strings always interpolate `{expr}` —
if that printed `hello, world`, you're ready for Topic 1.

## Running tests

Glide has testing built into the language — no framework to install.
`test` blocks live in the same file as the code:

```glide
fn double(n: Int) -> Int {
    n * 2
}

test "double doubles" {
    expect(double(4) == 8)
}
```

```bash
glide test hello.gld
```

## Editor setup

There's no Glide editor extension yet. Treat `.gld` files as plain text —
or tell your editor to use Rust highlighting for them, which is close enough
to be pleasant. The interpreter's error messages, `glide run`, and `glide
test` are the whole toolchain today.

## A heads-up

Glide is under active development: the language reference the lessons are
generated from changes as the language grows, and lessons regenerate to
match. If something you learned last month now behaves differently, that's
the language evolving — ask in the 💬 chat tab and you'll get an answer
grounded in the current reference.
