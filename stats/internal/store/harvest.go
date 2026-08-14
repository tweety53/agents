package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// GetHarvestOffset returns the last committed byte offset for
// transcriptPath, and whether CommitHarvestBatch has ever been called for
// it. A transcript this store has never heard of reads as (0, false, nil)
// -- exactly like a transcript file the harvester has never seen -- not
// an error, since that is the ordinary state of every transcript the
// first time it is read.
func (s *Store) GetHarvestOffset(ctx context.Context, transcriptPath string) (int64, bool, error) {
	var offset int64
	err := s.pool.QueryRow(ctx, `
		SELECT byte_offset FROM harvest_offsets WHERE transcript_path = $1
	`, transcriptPath).Scan(&offset)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("store: get harvest offset for %s: %w", transcriptPath, err)
	}
	return offset, true, nil
}

// CommitHarvestBatch applies deltas -- zero or more per-stage-run
// additive token patches -- and advances transcriptPath's committed byte
// offset from expectedOffset to newOffset, in one transaction. Either
// both effects land or neither does.
//
// This atomicity is the entire point of this method existing rather than
// two separate calls (an additive metrics write, then an offset advance,
// or the reverse): whichever order two separate calls used, a crash or a
// store outage landing between them left one artifact of a batch without
// the other. Offset-then-metrics silently *under-counted* that batch
// forever, on the routine condition (the store briefly unreachable) this
// whole change exists to survive -- not a rare crash window, but the
// ordinary shape of the outages the never-block guarantee is built
// around (task 9's post-commit review, findings F1 and its follow-up).
// Metrics-then-offset had the opposite failure: a successful additive
// write followed by a failed offset advance would be retried with the
// same delta, adding it a second time, silently inflating the total.
// Neither ordering is safe on its own; one transaction removes the
// choice rather than picking between them.
//
// expectedOffset and expectedFound are exactly what an earlier
// GetHarvestOffset call for transcriptPath returned, and this method's
// offset advance is guarded on them via optimistic concurrency: the
// UPDATE (or, when expectedFound is false, the INSERT) only succeeds if
// the row is still in the state the caller read it in. This closes a
// race nothing else in this method's shape prevents: GetHarvestOffset
// reads outside any transaction, so two callers -- two myflowd processes,
// one stale, is the concrete case this guards against, the same shape
// task 6's cross-process retire race already had to be closed rather
// than assumed away -- can both read the same offset, compute
// overlapping deltas from the same bytes, and both attempt to commit.
// Without a guard, jsonb_deep_add would sum that overlapping usage
// twice, and whichever commit lands last would set the final offset,
// which could even *regress* below the other committer's newOffset and
// cause the overlap to be re-read and added a third time on the next
// cycle. With the guard, only the first commit to reach Postgres
// succeeds; the second finds its expected row state gone and applies
// nothing at all (applied=false, err=nil -- not a failure, since losing
// this race is the correct, ordinary outcome for whichever caller reads
// stale state; see this method's own applied return value below).
//
// deltas may be empty -- a batch that read new bytes but attributed
// nothing to any open window (every message in it fell outside every
// registered window, or was a non-assistant type) still needs its
// offset advanced, so the same bytes are never re-read. An empty deltas
// map commits only the offset row.
//
// A patch value of nil (Go nil, not the JSON literal null) for any
// stage run is refused with ErrNilMetricsPatch before the transaction
// touches SQL, for the same reason MergeMetrics refuses it: a nil
// json.RawMessage marshals to SQL NULL, and jsonb_deep_add(a, NULL)
// returns NULL, which stage_runs.metrics' NOT NULL constraint would
// otherwise reject as a raw Postgres error.
//
// applied reports whether the batch was actually committed. applied is
// false with a nil error precisely when the optimistic-concurrency guard
// above did not match -- the caller lost a race with a concurrent
// committer for the same transcriptPath, and should simply re-read
// (GetHarvestOffset) and retry on its next cycle rather than treat this
// as a failure to log or alarm on. A non-nil error is a genuine failure
// (the connection dropped, an unknown stage run id, and so on).
func (s *Store) CommitHarvestBatch(ctx context.Context, transcriptPath string, expectedOffset int64, expectedFound bool, newOffset int64, deltas map[int64]json.RawMessage) (applied bool, err error) {
	for stageRunID, patch := range deltas {
		if patch == nil {
			return false, fmt.Errorf("store: commit harvest batch for %s, stage run %d: %w", transcriptPath, stageRunID, ErrNilMetricsPatch)
		}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("store: commit harvest batch for %s: begin: %w", transcriptPath, err)
	}
	// Rollback after a successful Commit is a documented no-op in pgx; its
	// error carries nothing actionable here, so it is discarded explicitly
	// rather than checked -- the same pattern every other transactional
	// method in this package already uses (query.go's queryTxOptions
	// callers, changes.go's project bootstrap). It is also what discards
	// a guard-mismatch's tentative writes: this function never calls
	// Commit on that path, so everything the transaction touched before
	// the guard was checked rolls back automatically.
	defer func() { _ = tx.Rollback(ctx) }()

	// The guard runs first, before any metrics write: a caller that has
	// already lost the race should do as little work as possible before
	// finding that out, and must not have touched stage_runs at all by
	// the time it does.
	var guardTag pgconn.CommandTag
	if expectedFound {
		guardTag, err = tx.Exec(ctx, `
			UPDATE harvest_offsets SET byte_offset = $2, updated_at = now()
			WHERE transcript_path = $1 AND byte_offset = $3
		`, transcriptPath, newOffset, expectedOffset)
	} else {
		guardTag, err = tx.Exec(ctx, `
			INSERT INTO harvest_offsets (transcript_path, byte_offset, updated_at)
			VALUES ($1, $2, now())
			ON CONFLICT (transcript_path) DO NOTHING
		`, transcriptPath, newOffset)
	}
	if err != nil {
		return false, fmt.Errorf("store: commit harvest batch for %s: advance offset: %w", transcriptPath, err)
	}
	if guardTag.RowsAffected() == 0 {
		// Lost the race: transcriptPath's row no longer matches what the
		// caller read (or, for a first-ever commit, someone else created
		// it first). Not an error -- see this method's own doc comment.
		return false, nil
	}

	for stageRunID, patch := range deltas {
		tag, err := tx.Exec(ctx, `
			UPDATE stage_runs SET metrics = jsonb_deep_add(metrics, $2::jsonb) WHERE id = $1
		`, stageRunID, patch)
		if err != nil {
			return false, fmt.Errorf("store: commit harvest batch for %s: add token metrics for stage run %d: %w", transcriptPath, stageRunID, err)
		}
		if tag.RowsAffected() == 0 {
			return false, fmt.Errorf("store: commit harvest batch for %s: %w: %d", transcriptPath, ErrStageRunNotFound, stageRunID)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("store: commit harvest batch for %s: commit: %w", transcriptPath, err)
	}
	return true, nil
}
