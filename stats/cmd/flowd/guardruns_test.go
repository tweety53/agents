package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/records"
)

// TestIngestGuardRuns: the guard-runs file is ingested as the hook-runs file
// is -- moved aside on one tick, recorded on the next, kept after a failed
// write -- and each valid line is passed with its worktree's project key,
// "" for a worktree that no longer resolves.
func TestIngestGuardRuns(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	key, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatal(err)
	}
	gone := filepath.Join(t.TempDir(), "removed-worktree")
	path := filepath.Join(t.TempDir(), "guard-runs.jsonl")
	body := `{"guard":"check-references","worktree":"` + repo + `","exitCode":1,"outcome":"fired","durationMs":12,"recordedAt":"2026-01-01T00:00:00Z","reasons":["F2"]}
{"guard":"","worktree":"` + repo + `","outcome":"clear","recordedAt":"2026-01-01T00:00:00Z"}
{"guard":"close-task","worktree":"` + gone + `","exitCode":0,"outcome":"clear","durationMs":3,"recordedAt":"2026-01-01T00:00:01Z"}
{"guard":"c","worktree":
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	var gotKeys []string
	var got []records.GuardRun
	fail := false
	record := func(_ context.Context, keys []string, runs []records.GuardRun) error {
		if fail {
			return errors.New("store down")
		}
		gotKeys, got = append(gotKeys, keys...), append(got, runs...)
		return nil
	}
	ctx := context.Background()
	if err := ingestGuardRuns(ctx, path, record); err != nil || len(got) != 0 {
		t.Fatalf("first tick: %v, recorded %v, want nothing yet", err, got)
	}
	fail = true
	if err := ingestGuardRuns(ctx, path, record); err == nil {
		t.Fatal("failed write: no error")
	}
	fail = false
	if err := ingestGuardRuns(ctx, path, record); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Guard != "check-references" || got[0].ExitCode != 1 || got[0].Reasons[0] != "F2" || got[1].Guard != "close-task" {
		t.Fatalf("recorded %+v, want the two valid lines in file order", got)
	}
	if gotKeys[0] != key || gotKeys[1] != "" {
		t.Errorf("keys %q, want [%s \"\"]", gotKeys, key)
	}
	if _, err := os.Stat(path + ".ingest"); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("moved file left behind: %v", err)
	}
}
