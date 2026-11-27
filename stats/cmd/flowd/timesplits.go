package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"time"

	"github.com/tweety53/agents/stats/internal/harvest"
	"github.com/tweety53/agents/stats/internal/store"
)

// timeSplitBatch bounds each tick's rows of each kind, so a backlog (every
// dispatch recorded before 0115_time_splits.sql) drains a tick at a time.
const timeSplitBatch = 200

// timeSplitSettle is how long an ended stage run whose transcript has no
// line at or past its end waits before its split is stored anyway: the line
// closing the run's last gap -- its own `flow stage end` call's result --
// normally lands within seconds; only a session that died never writes one.
const timeSplitSettle = 15 * time.Minute

// timeSplitStore is what harvestTimeSplits needs from the store.
type timeSplitStore interface {
	PendingTimeSplits(ctx context.Context, limit int) ([]store.PendingTimeSplit, error)
	LinkStageRunTranscript(ctx context.Context, id int64, path string) error
	SetTimeSplit(ctx context.Context, kind string, id int64, from, to time.Time, split json.RawMessage) error
}

// harvestTimeSplits is one tick of the time-split harvest
// (0115_time_splits.sql): every pending stage run is linked to its token's
// transcript under root (harvest.TokenTranscriptPath), and every ended
// pending row -- a stage run once its transcript has a line at or past its
// end or timeSplitSettle has passed, a dispatch at once, its agent having
// finished writing -- gets its transcript's harvest.SplitGaps over its
// window stored, or JSON null when no transcript exists. Each transcript is
// read once per tick. A row whose transcript read fails stays pending.
func harvestTimeSplits(ctx context.Context, root string, st timeSplitStore, now time.Time) error {
	rows, err := st.PendingTimeSplits(ctx, timeSplitBatch)
	if err != nil {
		return err
	}
	gapsOf := map[string][]harvest.Gap{}
	var errs []error
	for _, r := range rows {
		path, found := r.TranscriptPath, r.TranscriptPath != ""
		switch {
		case r.Kind == store.SplitStage && !found:
			path, _, found, err = harvest.TokenTranscriptPath(root, r.Token)
			if err == nil && found {
				err = st.LinkStageRunTranscript(ctx, r.ID, path)
			}
		case r.Kind == store.SplitDispatch:
			path, found, err = harvest.AgentTranscriptPath(root, r.SessionID, r.AgentID)
			if err == nil && !found && r.SessionID != "" {
				path, found, err = harvest.AgentTranscriptPath(root, "", r.AgentID)
			}
		}
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if r.EndedAt == nil {
			continue
		}
		split := json.RawMessage("null")
		if found {
			gaps, seen := gapsOf[path]
			if !seen {
				if gaps, err = harvest.ReadTranscriptGaps(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
					errs = append(errs, err)
					continue
				}
				gapsOf[path] = gaps
			}
			covered := len(gaps) > 0 && !gaps[len(gaps)-1].To.Before(*r.EndedAt)
			if r.Kind == store.SplitStage && !covered && now.Sub(*r.EndedAt) < timeSplitSettle {
				continue
			}
			if split, err = json.Marshal(harvest.SplitGaps(gaps, r.StartedAt, *r.EndedAt)); err != nil {
				errs = append(errs, err)
				continue
			}
		}
		if err := st.SetTimeSplit(ctx, r.Kind, r.ID, r.StartedAt, *r.EndedAt, split); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// runTimeSplitHarvest runs harvestTimeSplits every interval until ctx ends,
// logging a failed tick.
func runTimeSplitHarvest(ctx context.Context, root string, st timeSplitStore, interval time.Duration, logger *slog.Logger) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := harvestTimeSplits(ctx, root, st, time.Now()); err != nil {
				logger.Warn("flowd: time-split harvest failed, retried next tick", "error", err)
			}
		}
	}
}
