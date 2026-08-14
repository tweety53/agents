package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestGetHarvestOffsetUnknownTranscriptReadsAsZero asserts the "never
// seen before" case reads as a plain, error-free zero -- exactly the
// state of every transcript file the first time the harvester reads it.
func TestGetHarvestOffsetUnknownTranscriptReadsAsZero(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	offset, found, err := st.GetHarvestOffset(ctx, "/nonexistent/transcript.jsonl")
	if err != nil {
		t.Fatalf("GetHarvestOffset: %v", err)
	}
	if found {
		t.Errorf("found = true for a transcript never committed, want false")
	}
	if offset != 0 {
		t.Errorf("offset = %d, want 0", offset)
	}
}

// TestCommitHarvestBatchAddsAndAdvancesTogether is the direct guard
// against F1 (task 9's post-commit review, and its follow-up): a batch's
// additive token deltas and its transcript's advanced offset are visible
// together after CommitHarvestBatch returns, and a second batch's deltas
// add onto the first rather than replacing them -- the additive
// guarantee jsonb_deep_add provides, paired atomically with the offset
// that makes replaying a batch, or not replaying it, always the correct
// choice.
func TestCommitHarvestBatchAddsAndAdvancesTogether(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-commitbatch-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	path := "/synthetic/session.jsonl"

	applied, err := st.CommitHarvestBatch(ctx, path, 0, false, 1000, map[int64]json.RawMessage{
		run.ID: json.RawMessage(`{"tokens":{"main":{"input":10,"output":5}}}`),
	})
	if err != nil {
		t.Fatalf("CommitHarvestBatch (first): %v", err)
	}
	if !applied {
		t.Fatalf("CommitHarvestBatch (first) applied = false, want true")
	}
	offset, found, err := st.GetHarvestOffset(ctx, path)
	if err != nil {
		t.Fatalf("GetHarvestOffset (after first): %v", err)
	}
	if !found || offset != 1000 {
		t.Fatalf("offset after first commit = (%d, found=%v), want (1000, true)", offset, found)
	}

	applied, err = st.CommitHarvestBatch(ctx, path, 1000, true, 2500, map[int64]json.RawMessage{
		run.ID: json.RawMessage(`{"tokens":{"main":{"input":3,"output":2}}}`),
	})
	if err != nil {
		t.Fatalf("CommitHarvestBatch (second): %v", err)
	}
	if !applied {
		t.Fatalf("CommitHarvestBatch (second) applied = false, want true")
	}
	offset, found, err = st.GetHarvestOffset(ctx, path)
	if err != nil {
		t.Fatalf("GetHarvestOffset (after second): %v", err)
	}
	if !found || offset != 2500 {
		t.Fatalf("offset after second commit = (%d, found=%v), want (2500, true)", offset, found)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag struct {
		Tokens struct {
			Main struct {
				Input  float64 `json:"input"`
				Output float64 `json:"output"`
			} `json:"main"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}
	if bag.Tokens.Main.Input != 13 {
		t.Errorf("tokens.main.input = %v, want 13 (10 + 3, added across two batches)", bag.Tokens.Main.Input)
	}
	if bag.Tokens.Main.Output != 7 {
		t.Errorf("tokens.main.output = %v, want 7 (5 + 2)", bag.Tokens.Main.Output)
	}
}

// TestCommitHarvestBatchEmptyDeltasStillAdvancesOffset covers the batch
// that read new bytes but attributed nothing -- every message fell
// outside every registered window, or was a non-assistant type. The
// offset must still advance, or those same bytes are re-read forever.
func TestCommitHarvestBatchEmptyDeltasStillAdvancesOffset(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	path := "/synthetic/no-matches.jsonl"

	applied, err := st.CommitHarvestBatch(ctx, path, 0, false, 500, map[int64]json.RawMessage{})
	if err != nil {
		t.Fatalf("CommitHarvestBatch (empty deltas): %v", err)
	}
	if !applied {
		t.Fatalf("CommitHarvestBatch (empty deltas) applied = false, want true")
	}
	offset, found, err := st.GetHarvestOffset(ctx, path)
	if err != nil {
		t.Fatalf("GetHarvestOffset: %v", err)
	}
	if !found || offset != 500 {
		t.Fatalf("offset = (%d, found=%v), want (500, true)", offset, found)
	}
}

// TestCommitHarvestBatchFailurePartwayCommitsNothing is the atomicity
// guarantee's own test, and the one the coordinator's mutation targets
// directly: when one delta in a multi-stage-run batch fails (an unknown
// stage run id, here), *no* delta from that batch is applied and the
// offset does not advance -- not the deltas that happened to be
// processed before the failing one. A batch that is retried after this
// kind of failure must start from exactly where it was, never from
// "some of it landed".
//
// Map iteration order in Go is randomised, so this seeds two *known
// good* stage runs alongside the unknown id, runs the assertion, and
// checks that neither good stage run's metrics changed -- regardless of
// which position in iteration order the failing id happened to occupy.
func TestCommitHarvestBatchFailurePartwayCommitsNothing(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-commitbatch-partial-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	runA, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage (A): %v", err)
	}
	runB, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel"))
	if err != nil {
		t.Fatalf("BeginStage (B): %v", err)
	}

	const unknownStageRunID = 99999999
	path := "/synthetic/partial-batch.jsonl"

	_, err = st.CommitHarvestBatch(ctx, path, 0, false, 777, map[int64]json.RawMessage{
		runA.ID:           json.RawMessage(`{"tokens":{"main":{"input":10}}}`),
		unknownStageRunID: json.RawMessage(`{"tokens":{"main":{"input":20}}}`),
		runB.ID:           json.RawMessage(`{"tokens":{"main":{"input":30}}}`),
	})
	if !errors.Is(err, store.ErrStageRunNotFound) {
		t.Fatalf("CommitHarvestBatch error = %v, want errors.Is(_, store.ErrStageRunNotFound)", err)
	}

	// Neither known-good stage run's metrics changed...
	for name, run := range map[string]store.StageRun{"A": runA, "B": runB} {
		got, err := st.GetStageRun(ctx, run.ID)
		if err != nil {
			t.Fatalf("GetStageRun (%s): %v", name, err)
		}
		if string(got.Metrics) != "{}" {
			t.Errorf("stage run %s metrics = %s after a failed batch, want unchanged {}", name, got.Metrics)
		}
	}

	// ...and the offset never advanced, so a retry starts the same batch
	// from scratch rather than resuming partway through it.
	_, found, err := st.GetHarvestOffset(ctx, path)
	if err != nil {
		t.Fatalf("GetHarvestOffset: %v", err)
	}
	if found {
		t.Fatalf("offset was recorded for %s despite a failed batch, want no row at all", path)
	}
}

// TestCommitHarvestBatchRejectsNilPatch mirrors MergeMetrics' own guard:
// a nil json.RawMessage anywhere in deltas is refused before the
// transaction opens, for the same reason (jsonb_deep_add(a, NULL)
// returns SQL NULL, which stage_runs.metrics' NOT NULL constraint would
// otherwise reject as a raw Postgres error rather than a typed one).
func TestCommitHarvestBatchRejectsNilPatch(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-commitbatch-nilpatch-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	_, err = st.CommitHarvestBatch(ctx, "/synthetic/nil-patch.jsonl", 0, false, 42, map[int64]json.RawMessage{run.ID: nil})
	if !errors.Is(err, store.ErrNilMetricsPatch) {
		t.Fatalf("CommitHarvestBatch(nil patch) error = %v, want errors.Is(_, store.ErrNilMetricsPatch)", err)
	}
}

// TestHarvestBatchPreservesStageEndOutcomeKeys is F9's own test: the
// exact interaction jsonb_deep_merge (task 3) was created to fix, now
// checked for jsonb_deep_add specifically, since nothing in this suite
// exercised it before this test -- a stage-end mark's outcome keys
// (written through MergeMetrics) and a later harvest commit's token
// deltas (written through CommitHarvestBatch, a different SQL function)
// on the *same* stage run's metrics bag must both survive, in either
// write order.
func TestHarvestBatchPreservesStageEndOutcomeKeys(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-outcome-then-harvest-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	// The stage-end path: EndStage plus a MergeMetrics call carrying
	// outcome keys, exactly as ApplyEndStageMark (internal/api/stages.go)
	// does in production.
	endedAt := run.StartedAt.Add(90 * time.Second)
	if err := st.EndStage(ctx, run.ID, endedAt, "completed"); err != nil {
		t.Fatalf("EndStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"duration_ms":90000,"fast_mode":false}`)); err != nil {
		t.Fatalf("MergeMetrics (outcome keys): %v", err)
	}

	// Then a harvest batch adds token deltas for the same stage run, the
	// way a live daemon's Watcher would once the transcript for this
	// already-ended stage is read.
	applied, err := st.CommitHarvestBatch(ctx, "/synthetic/outcome-then-harvest.jsonl", 0, false, 4096, map[int64]json.RawMessage{
		run.ID: json.RawMessage(`{"tokens":{"main":{"input":42,"output":17}}}`),
	})
	if err != nil {
		t.Fatalf("CommitHarvestBatch: %v", err)
	}
	if !applied {
		t.Fatalf("CommitHarvestBatch applied = false, want true")
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag map[string]json.RawMessage
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}
	if _, ok := bag["duration_ms"]; !ok {
		t.Errorf("metrics lost duration_ms after a harvest batch touched only tokens: %s", got.Metrics)
	}
	if _, ok := bag["fast_mode"]; !ok {
		t.Errorf("metrics lost fast_mode after a harvest batch touched only tokens: %s", got.Metrics)
	}
	var tokens struct {
		Main struct {
			Input  float64 `json:"input"`
			Output float64 `json:"output"`
		} `json:"main"`
	}
	if err := json.Unmarshal(bag["tokens"], &tokens); err != nil {
		t.Fatalf("unmarshal tokens: %v", err)
	}
	if tokens.Main.Input != 42 {
		t.Errorf("tokens.main.input = %v, want 42", tokens.Main.Input)
	}
	if tokens.Main.Output != 17 {
		t.Errorf("tokens.main.output = %v, want 17", tokens.Main.Output)
	}

	// And the reverse order: a second stage run harvested first, then
	// given its outcome keys, must show the same survival.
	run2, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel"))
	if err != nil {
		t.Fatalf("BeginStage (run2): %v", err)
	}
	if _, err := st.CommitHarvestBatch(ctx, "/synthetic/harvest-then-outcome.jsonl", 0, false, 2048, map[int64]json.RawMessage{
		run2.ID: json.RawMessage(`{"tokens":{"main":{"input":7}}}`),
	}); err != nil {
		t.Fatalf("CommitHarvestBatch (run2): %v", err)
	}
	if err := st.EndStage(ctx, run2.ID, run2.StartedAt.Add(time.Minute), "completed"); err != nil {
		t.Fatalf("EndStage (run2): %v", err)
	}
	if err := st.MergeMetrics(ctx, run2.ID, json.RawMessage(`{"duration_ms":60000}`)); err != nil {
		t.Fatalf("MergeMetrics (run2 outcome keys): %v", err)
	}

	got2, err := st.GetStageRun(ctx, run2.ID)
	if err != nil {
		t.Fatalf("GetStageRun (run2): %v", err)
	}
	var bag2 map[string]json.RawMessage
	if err := json.Unmarshal(got2.Metrics, &bag2); err != nil {
		t.Fatalf("unmarshal metrics (run2): %v", err)
	}
	if _, ok := bag2["duration_ms"]; !ok {
		t.Errorf("run2 metrics lost duration_ms: %s", got2.Metrics)
	}
	var tokens2 struct {
		Main struct {
			Input float64 `json:"input"`
		} `json:"main"`
	}
	if err := json.Unmarshal(bag2["tokens"], &tokens2); err != nil {
		t.Fatalf("unmarshal tokens (run2): %v", err)
	}
	if tokens2.Main.Input != 7 {
		t.Errorf("run2 tokens.main.input = %v, want 7 (harvested before the outcome keys were written)", tokens2.Main.Input)
	}
}

// TestCommitHarvestBatchConcurrentCallersOnlyOneApplies is F7's own test:
// two CommitHarvestBatch calls for the same transcriptPath, both reading
// the same starting offset and computing overlapping deltas -- exactly
// what two myflowd processes (a stale one alongside a freshly started
// one; nothing prevents this) racing over the same transcript would
// produce. Exactly one call must apply; the other must return
// applied=false with a nil error; and the stored metrics must equal a
// single application, never the sum of both (which jsonb_deep_add would
// otherwise produce with no guard against it) and never neither (which a
// naive "first write wins, second write silently overwrites" scheme
// could also produce for the offset row).
func TestCommitHarvestBatchConcurrentCallersOnlyOneApplies(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-commitbatch-race-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	path := "/synthetic/race.jsonl"
	// Both callers read the same "nothing committed yet" starting state
	// and compute their own (here, identical) delta for the same
	// overlapping byte range, exactly as two harvesters that both read
	// the transcript from offset 0 would.
	const startOffset = 0
	const startFound = false
	const contendedNewOffset = 500

	results := make([]struct {
		applied bool
		err     error
	}, 2)
	var wg sync.WaitGroup
	for i := range 2 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			applied, err := st.CommitHarvestBatch(ctx, path, startOffset, startFound, contendedNewOffset, map[int64]json.RawMessage{
				run.ID: json.RawMessage(`{"tokens":{"main":{"input":10,"output":5}}}`),
			})
			results[i].applied, results[i].err = applied, err
		}(i)
	}
	wg.Wait()

	appliedCount := 0
	for i, r := range results {
		if r.err != nil {
			t.Fatalf("caller %d: unexpected error: %v", i, r.err)
		}
		if r.applied {
			appliedCount++
		}
	}
	if appliedCount != 1 {
		t.Fatalf("appliedCount = %d, want exactly 1 (one caller wins the race, the other applies nothing)", appliedCount)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag struct {
		Tokens struct {
			Main struct {
				Input  float64 `json:"input"`
				Output float64 `json:"output"`
			} `json:"main"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}
	if bag.Tokens.Main.Input != 10 {
		t.Errorf("tokens.main.input = %v, want 10 (a single application, not double-counted)", bag.Tokens.Main.Input)
	}
	if bag.Tokens.Main.Output != 5 {
		t.Errorf("tokens.main.output = %v, want 5", bag.Tokens.Main.Output)
	}

	offset, found, err := st.GetHarvestOffset(ctx, path)
	if err != nil {
		t.Fatalf("GetHarvestOffset: %v", err)
	}
	if !found || offset != contendedNewOffset {
		t.Fatalf("offset = (%d, found=%v), want (%d, true)", offset, found, contendedNewOffset)
	}
}

// TestCommitHarvestBatchConcurrentCallersOnExistingRowOnlyOneApplies is
// F11's own test: the previous concurrency test
// (TestCommitHarvestBatchConcurrentCallersOnlyOneApplies) races two
// callers with expectedFound=false, which only ever exercises the
// INSERT ... ON CONFLICT DO NOTHING branch of CommitHarvestBatch's
// guard -- the first-ever commit for a transcript. Every commit after
// the first goes through the UPDATE ... WHERE byte_offset = $3 branch
// instead, and nothing exercised that branch under real contention: the
// coordinator found, by deleting "AND byte_offset = $3" from the UPDATE,
// that the entire suite stayed green. That guard protects the more
// important race in practice -- a steady-state collision between two
// myflowd processes happens on every cycle they overlap, where a
// first-ever-commit collision happens once per transcript.
//
// This test seeds an existing harvest_offsets row first (a real prior
// commit, not a synthetic INSERT), then races two callers that both
// read that same row's offset and both pass expectedFound=true with
// overlapping deltas -- exactly two harvesters that both read the
// transcript's already-committed offset and compute overlapping usage
// from the same new bytes.
func TestCommitHarvestBatchConcurrentCallersOnExistingRowOnlyOneApplies(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-commitbatch-race-update-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	path := "/synthetic/race-update.jsonl"

	// A real prior commit -- this is what puts the UPDATE branch's guard
	// (rather than the INSERT branch's) in play for everything that
	// follows.
	const seededOffset = 1000
	applied, err := st.CommitHarvestBatch(ctx, path, 0, false, seededOffset, map[int64]json.RawMessage{
		run.ID: json.RawMessage(`{"tokens":{"main":{"input":1}}}`),
	})
	if err != nil {
		t.Fatalf("CommitHarvestBatch (seed): %v", err)
	}
	if !applied {
		t.Fatalf("CommitHarvestBatch (seed) applied = false, want true")
	}

	// Both callers read the same seeded offset and race to commit
	// overlapping deltas against it, with expectedFound=true -- the
	// UPDATE branch.
	const contendedNewOffset = 2500

	results := make([]struct {
		applied bool
		err     error
	}, 2)
	var wg sync.WaitGroup
	for i := range 2 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			applied, err := st.CommitHarvestBatch(ctx, path, seededOffset, true, contendedNewOffset, map[int64]json.RawMessage{
				run.ID: json.RawMessage(`{"tokens":{"main":{"input":10,"output":5}}}`),
			})
			results[i].applied, results[i].err = applied, err
		}(i)
	}
	wg.Wait()

	appliedCount := 0
	for i, r := range results {
		if r.err != nil {
			t.Fatalf("caller %d: unexpected error: %v", i, r.err)
		}
		if r.applied {
			appliedCount++
		}
	}
	if appliedCount != 1 {
		t.Fatalf("appliedCount = %d, want exactly 1 (one caller wins the race, the other applies nothing)", appliedCount)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag struct {
		Tokens struct {
			Main struct {
				Input  float64 `json:"input"`
				Output float64 `json:"output"`
			} `json:"main"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}
	// 1 from the seed commit, plus a single 10 from whichever racer won --
	// never 20 (both racers' 10 summed) and never just the seed's 1 (both
	// racers lost, which appliedCount already rules out).
	if bag.Tokens.Main.Input != 11 {
		t.Errorf("tokens.main.input = %v, want 11 (1 from the seed + a single 10, not double-counted)", bag.Tokens.Main.Input)
	}
	if bag.Tokens.Main.Output != 5 {
		t.Errorf("tokens.main.output = %v, want 5 (a single application, not double-counted)", bag.Tokens.Main.Output)
	}

	offset, found, err := st.GetHarvestOffset(ctx, path)
	if err != nil {
		t.Fatalf("GetHarvestOffset: %v", err)
	}
	if !found || offset != contendedNewOffset {
		t.Fatalf("offset = (%d, found=%v), want (%d, true)", offset, found, contendedNewOffset)
	}
}

// TestCommitHarvestBatchStaleExpectedOffsetAppliesNothing is the
// sequential twin of the concurrency tests above, covering the case the
// coordinator's review asked to be checked explicitly: a caller whose
// expectedOffset is *older* than what is actually stored -- not because
// it lost a live race, but because it read the offset before some other
// commit had already advanced it, and is only now getting around to
// committing its own (now-stale) batch. This is exactly the shape that
// would let the stored offset *regress*, re-exposing already-harvested
// bytes to be read and added again on the next cycle, if it were not
// refused.
func TestCommitHarvestBatchStaleExpectedOffsetAppliesNothing(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-commitbatch-stale-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	path := "/synthetic/stale-offset.jsonl"

	// The row starts at offset 1000 (a real prior commit)...
	if _, err := st.CommitHarvestBatch(ctx, path, 0, false, 1000, map[int64]json.RawMessage{
		run.ID: json.RawMessage(`{"tokens":{"main":{"input":1}}}`),
	}); err != nil {
		t.Fatalf("CommitHarvestBatch (first): %v", err)
	}
	// ...then advances again, to 2000, by a second, unrelated commit --
	// standing in for "another harvester (or another cycle of this same
	// one) already moved the offset forward while this caller was still
	// computing its own batch".
	if _, err := st.CommitHarvestBatch(ctx, path, 1000, true, 2000, map[int64]json.RawMessage{
		run.ID: json.RawMessage(`{"tokens":{"main":{"input":2}}}`),
	}); err != nil {
		t.Fatalf("CommitHarvestBatch (second): %v", err)
	}

	// A stale caller now shows up, believing the offset is still 1000 (a
	// smaller value than what is actually stored) and tries to commit a
	// batch that would advance it to only 1500 -- both older than the
	// current, real state.
	applied, err := st.CommitHarvestBatch(ctx, path, 1000, true, 1500, map[int64]json.RawMessage{
		run.ID: json.RawMessage(`{"tokens":{"main":{"input":999}}}`),
	})
	if err != nil {
		t.Fatalf("CommitHarvestBatch (stale): %v", err)
	}
	if applied {
		t.Fatalf("CommitHarvestBatch (stale) applied = true, want false -- a caller with an older expectedOffset than what is stored must not apply")
	}

	// The offset must not have regressed to the stale caller's 1500...
	offset, found, err := st.GetHarvestOffset(ctx, path)
	if err != nil {
		t.Fatalf("GetHarvestOffset: %v", err)
	}
	if !found || offset != 2000 {
		t.Fatalf("offset = (%d, found=%v), want (2000, true) -- it must not have regressed to the stale caller's 1500", offset, found)
	}

	// ...and the stale caller's delta (999) must not have been applied.
	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag struct {
		Tokens struct {
			Main struct {
				Input float64 `json:"input"`
			} `json:"main"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}
	if bag.Tokens.Main.Input != 3 {
		t.Errorf("tokens.main.input = %v, want 3 (1 + 2 from the two real commits; the stale caller's 999 must not have been added)", bag.Tokens.Main.Input)
	}
}
