package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/tweety53/agents/stats/internal/records"
)

// TestIngestHookRuns: a tick moves the hook-runs file aside and records
// nothing, so a hook that opened it just before still writes into the
// moved file; the next tick records the moved file's valid lines (a torn or
// invalid line is skipped) and removes it. A failed write keeps the moved
// file for the tick after; no file at all is a no-op.
func TestIngestHookRuns(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "hook-runs.jsonl")
	var got []records.HookRun
	fail := false
	record := func(_ context.Context, runs []records.HookRun) error {
		if fail {
			return errors.New("store down")
		}
		got = append(got, runs...)
		return nil
	}
	ctx := context.Background()
	if err := ingestHookRuns(ctx, path, record); err != nil {
		t.Fatalf("no file: %v", err)
	}
	body := `{"hook":"a","us":1500,"at":"2026-01-01T00:00:00Z"}
{"hook":"","us":1,"at":"2026-01-01T00:00:00Z"}
{"hook":"b","us":-1,"at":"2026-01-01T00:00:00Z"}
{"hook":"c","us":
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ingestHookRuns(ctx, path, record); err != nil || len(got) != 0 {
		t.Fatalf("first tick: %v, recorded %v, want nothing yet", err, got)
	}
	fail = true
	if err := ingestHookRuns(ctx, path, record); err == nil {
		t.Fatal("failed write: no error")
	}
	fail = false
	if err := ingestHookRuns(ctx, path, record); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Hook != "a" || got[0].DurationUS != 1500 {
		t.Errorf("recorded %+v, want hook a's run alone", got)
	}
	if _, err := os.Stat(path + ".ingest"); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("moved file left behind: %v", err)
	}
}
