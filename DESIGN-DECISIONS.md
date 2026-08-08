# Design decisions

Records the *why* behind choices where several routes existed, so reviewers
(human or Claude) get the big picture. Newest at the bottom.

## Glide support (2026-08)

Glide is Craig's own in-development language (sibling repo, `GLIDE_REPO`,
default `../glide`). No model knows it from training, which forced several
decisions that differ from every other language here:

- **The reference docs ride in the system prompt — no separate AI-facing
  doc.** The Glide repo's `docs/reference/language.md` + `stdlib.md`
  (~16KB, ~4k tokens) are read at startup and appended verbatim to the
  instructor preamble. A hand-maintained "bootcamp.md" rewrite of the same
  facts was considered and rejected: it would be a third copy of the
  language truth that silently drifts, and the terse ✓/○ table format is
  already near-ideal prompt material. Reading the same files Craig
  maintains makes sync automatic — when the language changes, the docs
  change in the same commit (a standing rule in the Glide repo), and the
  bootcamp picks them up on next start.
- **Missing docs hide the language.** If the reference can't be read,
  `initGlide` removes Glide from the registry — same pattern as an API
  provider without a key file. A Glide tab that generates hallucinated
  lessons would be worse than no tab.
- **Cache keys carry a docs hash.** Generated content is only as current
  as the docs it was generated from, and Glide changes weekly. A short
  hash of the docs namespaces every `glide:*` cache key inside
  `modelCacheKey` — the same chokepoint as the model namespace, so quiz
  answers stay paired with the exact quiz text they answered. Old entries
  orphan in `cache.json` rather than being purged; local disk is cheap and
  purging would add code.
- **Evaluations execute the submission.** Glide's interpreter is sitting
  in the sibling repo, so `handleEvaluate` runs the student's code
  (`glide run`, 10s timeout, 8KB output cap) and gives the model the real
  output as ground truth — the only language where "code cannot be
  executed" doesn't apply. Running arbitrary submissions is acceptable
  because the app is explicitly local-only (see Security in README) and
  today's Glide stdlib can read files but not reach the network. Generated
  *lesson* code is not auto-validated the same way: lessons stream to the
  browser as they generate, so a validate-and-retry loop doesn't fit the
  architecture — the Regenerate button is the recovery path, and the ✓/○
  rules in the prompt keep the failure rate down.
- **▶ Run in the editor is interpreter-only, no LLM.** `/api/run` executes
  the editor contents (same `runGlide` path, same timeout/output caps as
  evaluation) and returns raw output, so students iterate before spending
  an evaluation. The button only appears for languages where the server
  reports `canRun` — capability is decided server-side, not assumed by the
  frontend.
- **Editor highlighting is a textarea + mirror, not an editor library.**
  The solution editor stays a plain `<textarea>` (native undo, IME,
  selection) with a highlight.js-colored `<pre>` mirror absolutely
  positioned behind it and transparent text in front. CodeMirror/Monaco
  would do this better but violate the no-dependency rule for what is
  cosmetic polish; the mirror reuses the hljs setup lessons already ship.
  The trap to preserve: both layers must keep identical font geometry and
  wrapping (see style.css), or the caret drifts off the letters.
- **✓-only teaching is a prompt rule, not a doc edit.** The docs mark
  implemented (✓) vs designed-only (○) features; the Glide system prompt
  instructs the model to teach only ✓ and mention ○ as "coming later".
  Topic skills lines repeat the sharpest traps (no break/continue, no
  literal match patterns) because challenge prompts quote skills lines
  directly.
