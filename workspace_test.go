package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestMain points workspace persistence at a throwaway directory for the
// whole test process. This must be process-wide, not per-test: store* fire
// background saveWorkspaces goroutines that can outlive the test that started
// them, so a per-test redirect (or a t.Chdir) would let a straggler clobber
// the real data/workspaces.json after the test's cleanup had run.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "bootcamp-test")
	if err != nil {
		panic(err)
	}
	workspaceFile = filepath.Join(dir, "workspaces.json")
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// Exercises the workspace maps from many goroutines at once so the race
// detector (go test -race ./...) can catch any locking mistake.
func TestWorkspaceConcurrentAccess(t *testing.T) {
	const users, rounds = 4, 25
	var wg sync.WaitGroup
	for u := range users {
		user := fmt.Sprintf("user%d", u)
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := range rounds {
				storeSolution(user, "go:challenge:1:goat", fmt.Sprintf("code %d", i), "feedback")
				storeChat(user, "go:chat:1", []Message{
					{Role: "user", Content: "question"},
					{Role: "assistant", Content: fmt.Sprintf("answer %d", i)},
				})
				storeQuizWork(user, "go:quiz:1", []string{"A) yes", fmt.Sprintf("try %d", i)}, "graded")
				clearQuizWork(user, "go:quiz:2") // clearing a key nobody holds must be safe too
			}
		}()
		go func() {
			defer wg.Done()
			for range rounds {
				getWorkspace(user, "go:challenge:1:goat", "go:chat:1")
				getQuizWork(user, "go:quiz:1")
			}
		}()
	}
	wg.Wait()

	// After the dust settles every user must hold their own final round.
	for u := range users {
		user := fmt.Sprintf("user%d", u)
		sol, chat := getWorkspace(user, "go:challenge:1:goat", "go:chat:1")
		want := fmt.Sprintf("code %d", rounds-1)
		if sol.Code != want || sol.Feedback != "feedback" {
			t.Errorf("%s solution = %+v, want code %q", user, sol, want)
		}
		if len(chat) != 2 || chat[1].Role != "assistant" {
			t.Errorf("%s chat = %+v, want 2 messages ending with assistant", user, chat)
		}
		qw := getQuizWork(user, "go:quiz:1")
		wantAns := fmt.Sprintf("try %d", rounds-1)
		if len(qw.Answers) != 2 || qw.Answers[1] != wantAns || qw.Feedback != "graded" {
			t.Errorf("%s quiz work = %+v, want 2 answers ending %q", user, qw, wantAns)
		}
	}

	// clearQuizWork must actually clear, and getQuizWork must report the
	// zero value afterwards.
	clearQuizWork("user0", "go:quiz:1")
	if qw := getQuizWork("user0", "go:quiz:1"); len(qw.Answers) != 0 || qw.Feedback != "" {
		t.Errorf("after clear, quiz work = %+v, want empty", qw)
	}

	// getWorkspace must hand back a copy — mutating it can't touch the map.
	_, chat := getWorkspace("user0", "go:challenge:1:goat", "go:chat:1")
	chat[0].Content = "mutated"
	_, again := getWorkspace("user0", "go:challenge:1:goat", "go:chat:1")
	if again[0].Content == "mutated" {
		t.Error("getWorkspace returned the live chat slice, not a copy")
	}

	// getQuizWork must hand back a copy of the answers slice as well.
	qw := getQuizWork("user1", "go:quiz:1")
	qw.Answers[0] = "mutated"
	if again := getQuizWork("user1", "go:quiz:1"); again.Answers[0] == "mutated" {
		t.Error("getQuizWork returned the live answers slice, not a copy")
	}
}
