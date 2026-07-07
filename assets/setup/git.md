# Getting Started with Git & GitHub

Welcome! This course is different from the language courses: there's no code to
compile. Your "programs" are **commands you run in a terminal** and, later,
**workflow files you push to GitHub**. The editor built into this bootcamp
doesn't run commands — your AI instructor reads what you submit and gives
feedback — so the learning sticks when you run everything for real in a
terminal of your own. The good news: the setup below takes ten minutes and
you'll use it for the rest of your career.

## Check whether git is already installed

Open a terminal (Terminal.app on macOS, any shell on Linux, Git Bash or
PowerShell on Windows) and run:

```
git --version
```

If that prints a version number, skip to the config step below.

## Install git

- **macOS** — run `xcode-select --install` in the terminal (installs Apple's
  command-line tools, git included), or install via
  [Homebrew](https://brew.sh) with `brew install git`.
- **Windows** — download **[Git for Windows](https://git-scm.com/download/win)**
  and run the installer. It includes *Git Bash*, a terminal where all the
  commands in this course work exactly as written.
- **Linux** — use your package manager, e.g. `sudo apt install git` (Debian/
  Ubuntu) or `sudo dnf install git` (Fedora).

## Tell git who you are (one time only)

Every commit records an author. Set your name and email once, globally:

```
git config --global user.name "Your Name"
git config --global user.email "you@example.com"
```

This is covered properly in the first lesson — doing it now just means your
first `git commit` won't stop to complain.

## Create a GitHub account

From topic 7 onward you'll push code to **[github.com](https://github.com)** —
sign up for a free account now (the free tier covers everything in this
course, including GitHub Actions). Pick a username you're happy to have on
your work; it becomes part of every repository URL you share.

## A practice repository

The best companion to these lessons is a scratch folder where breaking things
doesn't matter:

```
mkdir git-practice
cd git-practice
git init
```

Run every command from the lessons in there. Version control is learned
through the fingers — when a lesson shows `git status`, type it and read what
comes back.

## Set up a code editor (optional but recommended)

Any editor works — git doesn't care — but a good one shows diffs and conflict
markers clearly:

1. **VS Code** from
   **[code.visualstudio.com](https://code.visualstudio.com)** has excellent
   built-in git support: changed lines in the gutter, a visual diff view, and
   a merge-conflict editor you'll appreciate around topic 6. Bonus: run
   `git config --global core.editor "code --wait"` and git will use it for
   commit messages too.
2. **Zed** from **[zed.dev](https://zed.dev)** is a fast, lighter alternative
   with git support built in.

Stick to the command line for the actual git operations while you learn —
the buttons in editors and GUI apps are wrappers around the commands this
course teaches, and they make far more sense once you know what they press.

## Where to look things up

- **[git-scm.com/docs](https://git-scm.com/docs)** — the official reference
  for every command. Also available offline: `git help commit`, `git help
  rebase`, and so on.
- **[The Pro Git book](https://git-scm.com/book)** — free, excellent, and the
  standard deep reference. Chapters 1-3 pair beautifully with topics 1-6.
- **[docs.github.com](https://docs.github.com)** — GitHub's own docs, including
  the **[GitHub Actions](https://docs.github.com/actions)** section you'll
  live in from topic 13.

## Suggested path

1. **Today:** install git, set your name and email, and make `git-practice` —
   then start topic 1 and type along.
2. **Before topic 7:** create your GitHub account.
3. **From topic 13:** you'll need a real repository on GitHub to run Actions
   workflows — the practice repo, pushed up, is perfect.

Stuck on any step — an installer that won't run, a `command not found`, a
Windows-vs-Mac difference? Open the **💬 Ask** tab and describe what you're
seeing.
