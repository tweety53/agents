package harvest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// DefaultTranscriptsRootEnv, when set, overrides DefaultTranscriptsRoot --
// a test's own isolated root, the same pattern
// internal/fallback.StateRoot's MYFLOW_STATE_DIR already uses for the
// state directory.
const DefaultTranscriptsRootEnv = "MYFLOW_TRANSCRIPTS_DIR"

// DefaultTranscriptsRoot returns the directory Claude Code writes session
// transcripts under: DefaultTranscriptsRootEnv when set, otherwise
// "~/.claude/projects" resolved against the current user's home
// directory.
func DefaultTranscriptsRoot() (string, error) {
	if v := os.Getenv(DefaultTranscriptsRootEnv); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("harvest: resolve home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "projects"), nil
}

// HarvestSink is where a batch's results -- attributed token deltas and
// the transcript's newly consumed byte offset -- are committed together.
//
// Defined here, at the consumer, per go-interface-design: internal/harvest
// never imports internal/store, so this package -- and
// TestHarvestNeedsNoDatabase in particular -- is testable with nothing
// but a fake satisfying these two methods, no PostgreSQL required. The
// daemon wires a real implementation backed directly by *store.Store,
// whose GetHarvestOffset and CommitHarvestBatch methods are written to
// match this interface exactly (structural typing, no adapter needed --
// cmd/myflowd/main.go asserts this at compile time).
//
// This interface, and the offset it makes this package's sole authority
// for, replaces an earlier design (this package's own git history) in
// which Watcher kept a local harvest-offsets.json file and committed
// metrics and the offset as two separate calls. Task 9's post-commit
// review found that design unsafe under both possible orderings of those
// two calls -- offset-then-metrics silently under-counted a batch on
// every store outage, which is the routine condition this whole change
// exists to survive, not a rare crash window; metrics-then-offset risked
// the opposite, a retried batch adding its usage twice. CommitHarvestBatch
// removes the choice by making both effects one atomic write, which is
// also what makes a purely local, uncommitted offset file redundant: the
// only offset that matters is the one Postgres holds beside the totals it
// governs.
//
// CommitHarvestBatch's expectedOffset/expectedFound parameters exist for
// a second reason, found in the same review's follow-up (F7): reading
// the offset and committing the batch are still two separate calls, so
// nothing prevents two Watchers -- two myflowd processes, one stale
// alongside a freshly started one -- from both reading the same offset,
// computing overlapping deltas, and both attempting to commit. Passing
// back exactly what GetHarvestOffset returned lets the store guard the
// commit with optimistic concurrency instead: only the caller whose
// expected state still matches actually commits.
type HarvestSink interface {
	// GetHarvestOffset returns the last committed byte offset for
	// transcriptPath, and whether any batch has ever been committed for
	// it (false, with offset 0, for a transcript never seen before).
	GetHarvestOffset(ctx context.Context, transcriptPath string) (offset int64, found bool, err error)
	// CommitHarvestBatch atomically adds deltas (additive token patches,
	// keyed by stage run id; may be empty) onto whatever is already
	// stored, and advances transcriptPath's committed offset from
	// (expectedOffset, expectedFound) -- exactly what a prior
	// GetHarvestOffset call returned -- to newOffset. Either both effects
	// land or neither does.
	//
	// applied reports whether the batch was actually committed. A false
	// return with a nil error means the offset had already moved by the
	// time this call reached the store -- a concurrent committer won the
	// race for the same transcriptPath -- which is not a failure: the
	// caller should simply re-read (GetHarvestOffset) and try again on
	// its next cycle. A non-nil error is a genuine failure.
	CommitHarvestBatch(ctx context.Context, transcriptPath string, expectedOffset int64, expectedFound bool, newOffset int64, deltas map[int64]json.RawMessage) (applied bool, err error)
}

// Pricer prices one stage run's already-committed metrics -- called only
// after a harvest batch's CommitHarvestBatch call has reported the batch
// actually applied (RunOnce, below), never from inside that commit's own
// transaction: the commit's atomicity is what makes harvesting
// exactly-once, and pricing is a pure, idempotent recomputation from
// metrics already durably stored, so it has nothing to contribute to that
// atomicity and everything to lose by extending it (a slow or failing
// pricing pass would then block the harvest it has no business blocking).
//
// Defined here, at the consumer, per go-interface-design, like
// WindowSource and HarvestSink above: internal/harvest never imports
// internal/store, so *store.Store.Price (whose signature already matches
// this interface exactly, no adapter needed) is wired in only by the
// daemon (cmd/myflowd/main.go, via WithPricer).
type Pricer interface {
	Price(ctx context.Context, stageRunID int64) error
}

// Watcher periodically scans a transcripts root for *.jsonl files and
// harvests whatever bytes are new since each one's last committed offset,
// attributing them via its Attributor and committing the result -- both
// the token deltas and the advanced offset -- through its HarvestSink in
// one atomic call per file. It never talks to PostgreSQL directly --
// only through WindowSource (via Attributor), HarvestSink, and (when
// configured) Pricer.
type Watcher struct {
	root       string
	sink       HarvestSink
	attributor *Attributor
	logger     *slog.Logger
	pricer     Pricer
}

// WatcherOption configures optional Watcher behaviour not every caller
// needs -- currently just WithPricer. Following the same functional-option
// shape internal/api.WithSPA already uses (cmd/myflowd/main.go) rather
// than growing NewWatcher's positional parameter list keeps every
// existing call site (this package's own tests included) compiling
// unchanged: a Watcher built with no options simply never prices, exactly
// as it behaved before this task.
type WatcherOption func(*Watcher)

// WithPricer configures the Watcher to price every stage run a batch
// touches, once that batch's CommitHarvestBatch call has reported it
// applied (RunOnce's own doc comment explains why after, and only after).
func WithPricer(p Pricer) WatcherOption {
	return func(w *Watcher) { w.pricer = p }
}

// NewWatcher builds a Watcher over root (scanned recursively for
// *.jsonl files), sink (where offsets are read from and results are
// committed to) and attributor (how records become deltas). logger may
// be nil. opts configures optional behaviour -- see WithPricer.
func NewWatcher(root string, sink HarvestSink, attributor *Attributor, logger *slog.Logger, opts ...WatcherOption) *Watcher {
	w := &Watcher{root: root, sink: sink, attributor: attributor, logger: logger}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

// RunOnce performs a single scan-and-harvest pass: every *.jsonl file
// under the watcher's root is read from its last committed offset
// (HarvestSink.GetHarvestOffset), attributed, and committed -- offset and
// deltas together, atomically -- through HarvestSink.CommitHarvestBatch.
//
// A failure at any point for one file (reading the transcript, looking
// up its committed offset, attributing its records, or committing the
// result) is logged and this loop moves on to the next file, rather than
// aborting the whole pass -- one bad file must never starve every other
// session's harvesting. Because nothing about a file's committed state
// changes until CommitHarvestBatch succeeds, a failure anywhere before
// that call leaves the file exactly as it was: the next RunOnce reads
// the same bytes from the same offset and tries again, with no risk of
// either losing that batch's usage or adding it twice.
//
// A commit that reports applied=false with a nil error (HarvestSink's own
// doc comment) is not logged as a failure at all: it means a concurrent
// harvester already advanced this file's offset first, which is the
// ordinary, correct outcome of losing that race, not an error condition
// -- the next RunOnce simply re-reads the current state and tries again.
//
// RunOnce returns the number of files whose newly read records were
// successfully committed by this call.
func (w *Watcher) RunOnce(ctx context.Context) (int, error) {
	files, err := discoverTranscripts(w.root)
	if err != nil {
		return 0, err
	}

	touchedFiles := 0
	for _, path := range files {
		if ctx.Err() != nil {
			return touchedFiles, ctx.Err()
		}

		offset, found, err := w.sink.GetHarvestOffset(ctx, path)
		if err != nil {
			w.warn("harvest: get committed offset failed, will retry", "path", path, "error", err)
			continue
		}

		records, newOffset, err := ReadNewRecords(path, offset)
		if err != nil {
			w.warn("harvest: read transcript failed, will retry", "path", path, "error", err)
			continue
		}
		if newOffset == offset {
			continue // nothing new (or only a partial trailing line) since last time.
		}

		deltas, err := w.attributor.Attribute(ctx, records)
		if err != nil {
			w.warn("harvest: attribute failed, will retry", "path", path, "error", err)
			continue
		}

		patches, err := encodePatches(deltas)
		if err != nil {
			w.warn("harvest: encode failed, will retry", "path", path, "error", err)
			continue
		}

		applied, err := w.sink.CommitHarvestBatch(ctx, path, offset, found, newOffset, patches)
		if err != nil {
			w.warn("harvest: commit failed, will retry", "path", path, "error", err)
			continue
		}
		if !applied {
			// Lost a race with a concurrent harvester for this file --
			// benign, not a warning; the next cycle re-reads and retries.
			continue
		}
		touchedFiles++

		// Price every stage run this batch touched -- deliberately after
		// CommitHarvestBatch has already reported applied, never inside
		// it (Pricer's own doc comment explains why). A pricing failure
		// (no rate for a model, no chargeable tokens yet, or a transient
		// store error) is logged and skipped, never fatal to this pass
		// and never a reason to retry the batch itself: the metrics are
		// already durably committed either way, and the next harvest
		// cycle -- or a future re-price -- gets another chance.
		if w.pricer != nil {
			for stageRunID := range deltas {
				if err := w.pricer.Price(ctx, stageRunID); err != nil {
					w.warn("harvest: price stage run failed, will retry next cycle", "stage_run_id", stageRunID, "error", err)
				}
			}
		}
	}
	return touchedFiles, nil
}

// encodePatches marshals each stage run's Delta into the json.RawMessage
// shape HarvestSink.CommitHarvestBatch's deltas parameter expects: the
// whole-run total under "tokens" (Delta.Total, unchanged meaning) and,
// when this batch carried any, the per-model breakdown under "models"
// (Delta.Models, task 22) -- omitted entirely rather than encoded as an
// empty object when a batch attributed nothing to any model, so it never
// masks a stage run's own "models" key already holding real data from an
// earlier batch (json.RawMessage(nil) participates in no merge at all,
// where an encoded "models":{} would still be a value jsonb_deep_add has
// to reconcile against).
func encodePatches(deltas map[int64]Delta) (map[int64]json.RawMessage, error) {
	patches := make(map[int64]json.RawMessage, len(deltas))
	for stageRunID, delta := range deltas {
		mp := MetricsPatch{Tokens: delta.Total, Speed: delta.Speed}
		if len(delta.Models) > 0 {
			mp.Models = make(map[string]ModelBucket, len(delta.Models))
			for model, td := range delta.Models {
				mp.Models[model] = ModelBucket{Tokens: td}
			}
		}
		patch, err := json.Marshal(mp)
		if err != nil {
			return nil, fmt.Errorf("harvest: encode metrics patch for stage run %d: %w", stageRunID, err)
		}
		patches[stageRunID] = patch
	}
	return patches, nil
}

// Run calls RunOnce every interval until ctx is done. A single failed
// pass is logged and does not stop the loop -- the daemon must keep
// harvesting future transcript growth even if one pass hit a transient
// error.
func (w *Watcher) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := w.RunOnce(ctx); err != nil && w.logger != nil {
				w.logger.Warn("harvest: run failed", "error", err)
			}
		}
	}
}

func (w *Watcher) warn(msg string, args ...any) {
	if w.logger != nil {
		w.logger.Warn(msg, args...)
	}
}

// discoverTranscripts walks root recursively and returns every *.jsonl
// file found, sorted for deterministic processing order. This picks up
// both a session's own top-level file
// (~/.claude/projects/<project>/<session>.jsonl) and its per-session
// subagents/agent-*.jsonl files, which carry sidechain messages under
// the same top-level sessionId -- confirmed by reading a real
// subagents/*.jsonl file directly rather than assumed. A root that does
// not exist yet is not an error: it reads as no transcripts found,
// exactly like a machine that has not run Claude Code yet.
func discoverTranscripts(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".jsonl") {
			out = append(out, path)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("harvest: scan transcripts root %s: %w", root, err)
	}
	sort.Strings(out)
	return out, nil
}
