package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

// Runs real Glide code through the sibling repo's interpreter. Skipped when
// the binary hasn't been built (make -C ../glide/glide build) so the suite
// still passes on a machine without the Glide checkout.
func TestGlideRunBlock(t *testing.T) {
	// Absolute, as initGlide guarantees: runGlide changes working directory.
	bin, err := filepath.Abs(filepath.Join(glideRepo, "glide", "bin", "glide"))
	if err != nil {
		t.Fatal(err)
	}
	old := glideBin
	glideBin = bin
	t.Cleanup(func() { glideBin = old })

	if glideRunBlock(context.Background(), "fn main() {\n}\n") == "" {
		t.Skipf("no glide interpreter at %s — skipping", bin)
	}

	tests := []struct {
		name, code, want string
	}{
		{"clean run output", "fn main() {\n    println(\"ok {1 + 1}\")\n}\n", "ok 2"},
		{"clean run status", "fn main() {\n}\n", "exited cleanly"},
		{"parse error is reported as failure", "fn main() { let x = }\n", "FAILED"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := glideRunBlock(context.Background(), tc.code)
			if !strings.Contains(got, tc.want) {
				t.Errorf("run block does not contain %q:\n%s", tc.want, got)
			}
		})
	}
}

// The Glide cache namespace must react to doc changes and leave other
// languages' keys untouched.
func TestGlideCacheNamespace(t *testing.T) {
	old := glideDocsHash
	t.Cleanup(func() { glideDocsHash = old })

	glideDocsHash = "aaaa1111"
	a := modelCacheKey("u", "glide:lesson:1")
	goKey := modelCacheKey("u", "go:lesson:1")
	glideDocsHash = "bbbb2222"
	b := modelCacheKey("u", "glide:lesson:1")

	if a == b {
		t.Errorf("glide key did not change with the docs hash: %q", a)
	}
	if !strings.Contains(a, "aaaa1111") {
		t.Errorf("glide key %q does not carry the docs hash", a)
	}
	if goKey != "go:lesson:1" {
		t.Errorf("non-glide key was rewritten: %q", goKey)
	}
}
