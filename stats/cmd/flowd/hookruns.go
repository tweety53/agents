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

	"github.com/tweety53/agents/stats/internal/records"
)

// ingestHookRuns is one tick of the hook-runs ingestion: `flow-guard hook`
// appends each run's wall time to path (records.HookRunsPath) and never
// calls flowd, so flowd collects the file here. A tick with path+".ingest"
// on disk records its valid lines through record and removes it -- kept for
// the next tick when record fails; a tick without one renames path to it and
// records nothing, so a hook that opened path just before the rename has the
// whole tick to finish its write into the moved file. A line that is torn or
// carries no hook, a negative duration or no time is skipped. No file is a
// no-op.
func ingestHookRuns(ctx context.Context, path string, record func(context.Context, []records.HookRun) error) error {
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
	var runs []records.HookRun
	for _, line := range bytes.Split(body, []byte("\n")) {
		var r records.HookRun
		if json.Unmarshal(line, &r) != nil || r.Hook == "" || r.DurationUS < 0 || r.At.IsZero() {
			continue
		}
		runs = append(runs, r)
	}
	if err := record(ctx, runs); err != nil {
		return err
	}
	return os.Remove(taken)
}

// runHookRunsIngest runs ingestHookRuns every interval until ctx ends,
// logging a failed tick.
func runHookRunsIngest(ctx context.Context, path string, record func(context.Context, []records.HookRun) error, interval time.Duration, logger *slog.Logger) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := ingestHookRuns(ctx, path, record); err != nil {
				logger.Warn("flowd: hook-runs ingest failed, retried next tick", "path", path, "error", err)
			}
		}
	}
}
