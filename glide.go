package main

// Glide support. Glide is a brand-new language developed in a sibling repo —
// no model knows it from training, so unlike every other language here the
// bootcamp cannot lean on trained knowledge. Three things make it work, all
// in this file:
//
//  1. The complete language + stdlib reference (small by design, ~4k tokens)
//     is read from the Glide repo at startup and appended to the system
//     prompt. The docs ARE the model's entire knowledge of Glide. If they
//     can't be found, Glide simply doesn't appear in the UI — the same rule
//     as an API provider without a key file.
//
//  2. Generated content is only as current as the docs it was generated
//     from, and Glide changes week to week. A short hash of the docs
//     namespaces every Glide cache key (see modelCacheKey in cache.go), so
//     changing the language regenerates content instead of teaching last
//     month's Glide.
//
//  3. Glide is the one language whose interpreter sits right there in the
//     repo, so evaluations run the student's submission for real and give
//     the model the actual output as ground truth (see glideRunBlock and
//     handleEvaluate) — every other language is judged by eye.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// glideRepo is the Glide repo checkout, overridable with GLIDE_REPO. The
// default assumes the sibling layout used on the dev machine
// (…/_projects/bootcamp next to …/_projects/glide).
var glideRepo = "../glide"

var (
	glideDocsHash string // short hash of the reference docs; "" = not loaded
	glideBin      string // interpreter binary path; "" = execution unavailable
)

// initGlide wires Glide up at startup: loads the reference docs into the
// system prompt, computes the cache-namespacing hash, and locates the
// interpreter. Called from main before the server starts serving.
func initGlide() {
	if env := os.Getenv("GLIDE_REPO"); env != "" {
		glideRepo = env
	}
	refDir := filepath.Join(glideRepo, "docs", "reference")
	langRef, errLang := os.ReadFile(filepath.Join(refDir, "language.md"))
	stdlibRef, errStd := os.ReadFile(filepath.Join(refDir, "stdlib.md"))
	if errLang != nil || errStd != nil {
		removeLanguage("glide")
		log.Printf("glide: reference docs not readable under %s — Glide hidden from the UI (set GLIDE_REPO to the repo checkout)", refDir)
		return
	}

	// 4 bytes (8 hex chars) is plenty: this only needs to change when the
	// docs change, not survive an adversary.
	sum := sha256.Sum256(append(langRef, stdlibRef...))
	glideDocsHash = hex.EncodeToString(sum[:4])

	// The languages map holds a copy of the var, so update the map entry —
	// mutating glideLanguage itself would be invisible to lookupLang.
	l := languages["glide"]
	l.SystemPrompt += "\n\n--- THE GLIDE LANGUAGE REFERENCE (your ONLY source of truth for Glide) ---\n" +
		string(langRef) +
		"\n\n--- THE GLIDE STANDARD LIBRARY REFERENCE ---\n" + string(stdlibRef)
	languages["glide"] = l

	// Absolute path: runGlide sets the command's working directory to a temp
	// dir (so interpreter errors name "solution.gld", not a /var/folders path),
	// which would break a relative binary path.
	bin, err := filepath.Abs(filepath.Join(glideRepo, "glide", "bin", "glide"))
	if err != nil {
		bin = filepath.Join(glideRepo, "glide", "bin", "glide")
	}
	if info, err := os.Stat(bin); err == nil && info.Mode()&0111 != 0 {
		glideBin = bin
		log.Printf("glide: docs@%s loaded, interpreter at %s", glideDocsHash, bin)
	} else {
		log.Printf("glide: docs@%s loaded; no interpreter at %s — evaluations will not execute code (build it: make -C %s build)",
			glideDocsHash, bin, filepath.Join(glideRepo, "glide"))
	}
}

// removeLanguage drops a language from the registry and the UI switcher, for
// languages whose startup requirements aren't met.
func removeLanguage(id string) {
	delete(languages, id)
	for ci := range categories {
		langs := categories[ci].Langs
		for i, l := range langs {
			if l == id {
				categories[ci].Langs = append(langs[:i], langs[i+1:]...)
				break
			}
		}
	}
}

// glideRunTimeout bounds one interpreter run. Glide has `for {}` — an
// accidental infinite loop in a submission must not hang the evaluation.
const glideRunTimeout = 10 * time.Second

// glideRunOutputLimit caps how much interpreter output goes into the prompt,
// so a print-in-a-loop submission can't blow the request up.
const glideRunOutputLimit = 8 * 1024

// runGlide executes code with the real interpreter and reports what happened:
// the combined output (capped), a one-line human-readable result, and whether
// the run failed. err is reserved for infrastructure problems (no temp file);
// a program that doesn't run is failed=true, not an err.
func runGlide(ctx context.Context, code string) (output, result string, failed bool, err error) {
	// The code runs as solution.gld inside its own temp dir, with the
	// interpreter's working directory set there — so error messages say
	// "solution.gld: line 3: …" instead of leaking a /var/folders path into
	// the student's face and the evaluation prompt.
	dir, err := os.MkdirTemp("", "bootcamp-glide-")
	if err != nil {
		return "", "", false, err
	}
	defer os.RemoveAll(dir)
	if err := os.WriteFile(filepath.Join(dir, "solution.gld"), []byte(code), 0644); err != nil {
		return "", "", false, err
	}

	runCtx, cancel := context.WithTimeout(ctx, glideRunTimeout)
	defer cancel()
	cmd := exec.CommandContext(runCtx, glideBin, "run", "solution.gld")
	cmd.Dir = dir
	out, runErr := cmd.CombinedOutput()

	result = "the program exited cleanly (status 0)"
	switch {
	case runCtx.Err() == context.DeadlineExceeded:
		failed = true
		result = fmt.Sprintf("KILLED — still running after %v (infinite loop?)", glideRunTimeout)
	case runErr != nil:
		// Covers parse errors, runtime panics, non-zero os.exit — the
		// interpreter's own message is in the captured output.
		failed = true
		result = fmt.Sprintf("the program FAILED (%v)", runErr)
	}

	if len(out) > glideRunOutputLimit {
		out = append(out[:glideRunOutputLimit], "\n… (output truncated)"...)
	}
	return string(out), result, failed, nil
}

// glideRunBlock executes a submission with the real interpreter and formats
// the result as a prompt section for the evaluation. Returns "" when the
// interpreter isn't available, in which case the evaluation falls back to
// judging by eye like every other language.
func glideRunBlock(ctx context.Context, code string) string {
	if glideBin == "" {
		return ""
	}
	display, result, _, err := runGlide(ctx, code)
	if err != nil {
		log.Printf("glide run: %v", err)
		return ""
	}
	if display == "" {
		display = "(no output)"
	}

	return fmt.Sprintf(`The submission was executed with the real Glide interpreter (glide run, no
arguments, no stdin). Result: %s.

Output:
`+"```\n%s\n```"+`

Treat this run as ground truth over your own reading of the code: an error
above means the code does not run, however plausible it looks — quote the
interpreter's message in **Issues Found**. A clean run means the syntax is
valid, but whether the output satisfies the requirements is still yours to
judge. If the program needed arguments or input, the run won't show its full
behavior — say so and judge the logic from the code.`, result, display)
}

// langCanRun reports whether the ▶ Run button works for a language — today
// only Glide, and only when its interpreter was found at startup.
func langCanRun(id string) bool {
	return id == "glide" && glideBin != ""
}

// handleRunCode is the ▶ Run button: execute the student's editor contents
// and return the real output, so they can iterate before submitting for
// evaluation. No LLM involved — this is just the interpreter.
func handleRunCode(w http.ResponseWriter, r *http.Request, user string) {
	var req struct {
		Lang string `json:"lang"`
		Code string `json:"code"`
	}
	if !decodePOST(w, r, &req) {
		return
	}
	if _, ok := lookupLang(w, req.Lang); !ok {
		return
	}
	if !langCanRun(req.Lang) {
		jsonErr(w, 400, "running code isn't supported for this language")
		return
	}
	if strings.TrimSpace(req.Code) == "" {
		jsonErr(w, 400, "nothing to run")
		return
	}
	output, result, failed, err := runGlide(r.Context(), req.Code)
	if err != nil {
		log.Printf("glide run: %v", err)
		jsonErr(w, 500, "could not run the code")
		return
	}
	jsonOK(w, map[string]any{"output": output, "result": result, "failed": failed})
}
