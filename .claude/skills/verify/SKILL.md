---
name: verify
description: Build, launch, and drive the bootcamp web app headlessly to verify changes at the browser surface.
---

# Verifying the bootcamp app

## Build & launch

```bash
go build -o bootcamp . && go vet ./...
# without API key (UI + non-streaming endpoints only):
PORT=8191 ./bootcamp -dev > /tmp/bootcamp-test.log 2>&1 &
# with API key (needed for any /api/*chat|lesson|challenge|evaluate|hint streaming):
ANTHROPIC_API_KEY=$(cat claude.key) PORT=8191 ./bootcamp -dev > /tmp/bootcamp-test.log 2>&1 &
```

`-dev` bypasses login (auto-auth as user "dev"), so no cookie dance is needed —
curl works directly against every endpoint. Use a non-default port (8181 is the
user's own running instance). Stop with `lsof -ti :8191 | xargs kill` —
background jobs don't survive across shell invocations, so kill by port.

## Driving the UI

No system playwright install. Use `playwright-core` (npm i in the scratchpad)
with the cached headless shell as `executablePath`:

```
~/Library/Caches/ms-playwright/chromium_headless_shell-*/chrome-headless-shell-mac-arm64/chrome-headless-shell
```

Give the page ~1.5s after `goto` for dev auto-login + topic load. Useful
anchors: `#topic-list` (fundamentals sidebar), `.tab[data-tab="..."]` (Lesson/
Challenge/Ask tabs), `#header-badge` / `#header-title`, `#lesson-output`,
`#progress-label`, `.lang-btn[data-lang="css"]` (language switcher).

## Gotchas

- Streaming endpoints without a key return an SSE `error` event, not an HTTP
  error — check the body, not the status.
- One real chat round-trip costs a few hundred tokens of the user's key; do it
  once when the change touches a chat/streaming path, not per probe.
- Server logs land in the file you redirected to; check it when a request
  seems to hang.
