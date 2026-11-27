package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"time"

	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/records"
)

// ingestLines is one tick of a local-file ingestion: a writer appends one
// JSON line per run to path and never calls flowd, so flowd collects the
// file here. A tick with path+".ingest" on disk passes its lines to record
// and removes it -- kept for the next tick when record fails; a tick without
// one renames path to it and records nothing, so a writer that opened path
// just before the rename has the whole tick to finish its write into the
// moved file. No file is a no-op.
func ingestLines(ctx context.Context, path string, record func(context.Context, [][]byte) error) error {
	taken := path + ".ingest"
	body, err := os.ReadFile(taken)
	if errors.Is(err, fs.ErrNotExist) {
		if err := os.Rename(path, taken); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}
	if err := record(ctx, bytes.Split(body, []byte("\n"))); err != nil {
		return err
	}
	return os.Remove(taken)
}

// ingestHookRuns is one tick of the hook-runs ingestion: `flow-guard hook`
// appends each run's wall time to path (records.HookRunsPath). A line that
// is torn or carries no hook, a negative duration or no time is skipped.
func ingestHookRuns(ctx context.Context, path string, record func(context.Context, []records.HookRun) error) error {
	return ingestLines(ctx, path, func(ctx context.Context, lines [][]byte) error {
		var runs []records.HookRun
		for _, line := range lines {
			var r records.HookRun
			if json.Unmarshal(line, &r) != nil || r.Hook == "" || r.DurationUS < 0 || r.At.IsZero() {
				continue
			}
			runs = append(runs, r)
		}
		return record(ctx, runs)
	})
}

// ingestGuardRuns is one tick of the guard-runs ingestion: flow-guard
// appends each dispatched guard's run to path (records.GuardRunsPath)
// instead of posting it. Each line's project is its worktree's
// (fallback.ProjectKey), resolved once per worktree per tick; a worktree
// that no longer resolves passes "", which the store keys by the change
// whose worktrees map names it. A torn line, or one with no guard, worktree
// or time, is skipped; record skips a run it refuses, as the POST it
// replaces answered 404 or 400.
func ingestGuardRuns(ctx context.Context, path string, record func(context.Context, []string, []records.GuardRun) error) error {
	return ingestLines(ctx, path, func(ctx context.Context, lines [][]byte) error {
		var keys []string
		var runs []records.GuardRun
		resolved := map[string]string{}
		for _, line := range lines {
			var r records.GuardRun
			if json.Unmarshal(line, &r) != nil || r.Guard == "" || r.Worktree == "" || r.RecordedAt.IsZero() {
				continue
			}
			key, ok := resolved[r.Worktree]
			if !ok {
				key, _, _ = fallback.ProjectKey(r.Worktree)
				resolved[r.Worktree] = key
			}
			keys, runs = append(keys, key), append(runs, r)
		}
		return record(ctx, keys, runs)
	})
}

// runIngest runs tick every interval until ctx ends, logging a failed tick.
func runIngest(ctx context.Context, what, path string, tick func(context.Context, string) error, interval time.Duration, logger *slog.Logger) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := tick(ctx, path); err != nil {
				logger.Warn("flowd: "+what+" ingest failed, retried next tick", "path", path, "error", err)
			}
		}
	}
}
