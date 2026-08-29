package harvest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/harvest"
)

// copyFixtureInto copies the fixture at srcPath into dir under name,
// returning the destination path -- Watcher.RunOnce operates on a real
// filesystem tree (it walks a root for *.jsonl files), so every watcher
// test needs its own isolated directory rather than touching the
// repository's testdata directly.
func copyFixtureInto(t *testing.T, dir, name, srcPath string) string {
	t.Helper()
	data := readFixture(t, srcPath)
	dst := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", dst, err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("write fixture copy %s: %v", dst, err)
	}
	return dst
}

// openWindowForMainSession is the one window every watcher test in this
// file needs: it covers every timestamp the main-thread and sidechain
// fixtures carry for mainSessionID.
func openWindowForMainSession(stageRunID int64) map[string][]harvest.Window {
	return map[string][]harvest.Window{
		mainSessionID: {{
			StageRunID: stageRunID,
			SessionID:  mainSessionID,
			StartedAt:  time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		}},
	}
}

// fakeHarvestSink is an in-memory stand-in for *store.Store's
// GetHarvestOffset/CommitHarvestBatch pair. It reproduces two properties
// that matter for these tests:
//
//   - A committed batch's offset advance and its additive token deltas
//     become visible *together*, and a failure commits neither. It never
//     partially applies a batch -- unlike the split-transaction mutation
//     internal/store/harvest_test.go's own
//     TestCommitHarvestBatchFailurePartwayCommitsNothing guards against at
//     the real-Postgres level, this fake cannot even express that failure
//     mode, by construction (it mutates its maps only after deciding to
//     succeed).
//   - The optimistic-concurrency guard real *store.Store.CommitHarvestBatch
//     enforces (F7, task 9's post-commit review): a commit whose
//     (expectedOffset, expectedFound) no longer matches the sink's actual
//     state for that path applies nothing and returns (false, nil), not
//     an error. This fake enforces the same guard so a Watcher-level bug
//     that ignored a stale read (rather than the guard living only in
//     internal/store, invisible to these tests) would still be caught
//     here.
//
// failNextCommits, when positive, makes the next N CommitHarvestBatch
// calls that would otherwise succeed fail instead (decremented per call,
// without mutating any state) -- used to simulate a store outage spanning
// several harvest cycles.
//
// commitCount matters beyond simulating outages: Watcher.RunOnce logs and
// swallows a per-path CommitHarvestBatch failure ("harvest: commit
// failed, will retry", watcher.go) rather than surfacing it through its
// return, so a test that drives several RunOnce cycles and asserts only
// that some value *stayed constant* across them cannot tell "every cycle
// committed and the value correctly held" apart from "every cycle after
// the first silently failed, so nothing changed" -- both look identical
// to that assertion (F18, pass 4 of this change's own review panel; see
// TestSpawnDepthStaysConstantAcrossOrdinaryReSends). Any RunOnce-driven
// test whose point is a value's stability across repeated cycles must
// therefore also assert a call-count signal (commitCount, as the
// pre-existing outage tests already do) so a silently-swallowed failure
// fails the test instead of passing it.
type fakeHarvestSink struct {
	offsets        map[string]int64
	totals         map[int64]harvest.TokenDelta                // cumulative, reproducing jsonb_deep_add's effect
	modelTotals    map[int64]map[string]harvest.TokenDelta     // same, per model (task 22's "models" key)
	dispatchTotals map[int64]map[string]harvest.DispatchBucket // same, per dispatch (KAN-201's "dispatches" key) --
	// Tokens accumulate through addBucket, a genuine numeric sum,
	// unchanged from before. The four descriptor fields (agent_type,
	// description, model, spawn_depth) are applied by simple presence:
	// whichever fields a batch's MetricsPatch carries replace whatever
	// this fake already held for that dispatch, exactly once per field
	// that is ever non-empty (production only ever starts sending them
	// once hasMeta first becomes true, and resends the same constant
	// values afterward -- encodePatches' own doc comment, watcher.go).
	//
	// This fake deliberately does NOT attempt jsonb_deep_add's own
	// per-leaf, by-JSON-type sum-vs-replace rule (0005_jsonb_deep_add.sql)
	// any more (F19, this change's own review panel): an earlier version
	// of this fake carried a second, hand-written implementation of that
	// rule (mergeDeepAddFake, isJSONNumberLiteral) to catch a descriptor
	// whose wire encoding reverted to a JSON number (F12) -- and that
	// second implementation itself drifted from the real rule once
	// already (F15), decoding a quoted numeric string as if it were a
	// number and silently summing it. Two drifts in three review rounds
	// is the standing risk F19 flags: a future addition to
	// jsonb_deep_add's type handling is only caught here if someone
	// remembers to mirror it by hand. The merge-semantics coverage that
	// actually needs a JSON-type-aware rule -- spawn_depth summed instead
	// of replaced across two ordinary commits -- now runs against the
	// real function instead:
	// internal/store/harvest_test.go's
	// TestCommitHarvestBatchReplacesSpawnDepthStringAcrossTwoOrdinaryCommits
	// and TestCommitHarvestBatchSumsDispatchTokensAcrossTwoCommits. What
	// remains here is deliberately dumb.

	commitCount     int
	failNextCommits int
	failGetOffset   bool

	// forceLoseRace, when true, makes the next CommitHarvestBatch call
	// return (false, nil) unconditionally -- applying no state change --
	// regardless of whether expectedOffset/expectedFound would otherwise
	// have matched. This stands in for a concurrent harvester's commit
	// landing in the gap between this call's GetHarvestOffset read and its
	// own CommitHarvestBatch call, the race HarvestSink's own doc comment
	// describes and RunOnce's "applied=false" branch exists to handle --
	// distinct from the stale-offset case the ordinary comparison above
	// already reproduces, which requires the sink's real state to have
	// actually moved.
	forceLoseRace bool
}

func newFakeHarvestSink() *fakeHarvestSink {
	return &fakeHarvestSink{
		offsets:        map[string]int64{},
		totals:         map[int64]harvest.TokenDelta{},
		modelTotals:    map[int64]map[string]harvest.TokenDelta{},
		dispatchTotals: map[int64]map[string]harvest.DispatchBucket{},
	}
}

func (s *fakeHarvestSink) GetHarvestOffset(_ context.Context, path string) (int64, bool, error) {
	if s.failGetOffset {
		return 0, false, errors.New("fake sink: get offset unavailable")
	}
	off, ok := s.offsets[path]
	return off, ok, nil
}

func (s *fakeHarvestSink) CommitHarvestBatch(_ context.Context, path string, expectedOffset int64, expectedFound bool, newOffset int64, deltas map[int64]json.RawMessage) (bool, error) {
	if s.failNextCommits > 0 {
		s.failNextCommits--
		return false, errors.New("fake sink: commit unavailable")
	}
	if s.forceLoseRace {
		s.forceLoseRace = false
		return false, nil
	}

	curOffset, curFound := s.offsets[path]
	if curFound != expectedFound || (expectedFound && curOffset != expectedOffset) {
		// Same guard real CommitHarvestBatch enforces: the caller's
		// expected state is stale. Benign, not an error.
		return false, nil
	}

	s.commitCount++
	for stageRunID, patch := range deltas {
		var mp harvest.MetricsPatch
		if err := json.Unmarshal(patch, &mp); err != nil {
			return false, err
		}
		t := s.totals[stageRunID]
		addBucket(&t.Main, mp.Tokens.Main)
		addBucket(&t.Sidechain, mp.Tokens.Sidechain)
		s.totals[stageRunID] = t

		if len(mp.Models) > 0 {
			byModel := s.modelTotals[stageRunID]
			if byModel == nil {
				byModel = map[string]harvest.TokenDelta{}
			}
			for model, bucket := range mp.Models {
				td := byModel[model]
				addBucket(&td.Main, bucket.Tokens.Main)
				addBucket(&td.Sidechain, bucket.Tokens.Sidechain)
				byModel[model] = td
			}
			s.modelTotals[stageRunID] = byModel
		}

		if len(mp.Dispatches) > 0 {
			byDispatch := s.dispatchTotals[stageRunID]
			if byDispatch == nil {
				byDispatch = map[string]harvest.DispatchBucket{}
			}
			for agentID, bucket := range mp.Dispatches {
				db := byDispatch[agentID]
				addBucket(&db.Tokens.Main, bucket.Tokens.Main)
				addBucket(&db.Tokens.Sidechain, bucket.Tokens.Sidechain)

				// Descriptor fields replace by simple presence -- no
				// JSON-type-aware merge (see the struct's own doc
				// comment above for why that logic was removed rather
				// than reproduced a third time).
				if bucket.AgentType != "" {
					db.AgentType = bucket.AgentType
				}
				if bucket.Description != "" {
					db.Description = bucket.Description
				}
				if bucket.Model != "" {
					db.Model = bucket.Model
				}
				if bucket.SpawnDepth != nil {
					db.SpawnDepth = bucket.SpawnDepth
				}

				byDispatch[agentID] = db
			}
			s.dispatchTotals[stageRunID] = byDispatch
		}
	}
	s.offsets[path] = newOffset
	return true, nil
}

func addBucket(dst *harvest.Bucket, src harvest.Bucket) {
	dst.Input += src.Input
	dst.Output += src.Output
	dst.CacheCreation += src.CacheCreation
	dst.CacheRead += src.CacheRead
	dst.Thinking += src.Thinking
}

var _ harvest.HarvestSink = (*fakeHarvestSink)(nil)

// TestNewWatcherPanicsOnNilDeps is KAN-173 task 2's own deliverable: deps
// is a required parameter, not an optional one, so a nil value must fail
// loudly at construction time -- before any transcript is read -- rather
// than reach RunOnce and silently price nothing, bind nothing and charge
// no dispatch, the exact failure shape KAN-16 and KAN-172 both were.
func TestNewWatcherPanicsOnNilDeps(t *testing.T) {
	sink := newFakeHarvestSink()

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		harvest.NewWatcher(t.TempDir(), sink, harvest.NewAttributor(nil), nil, nil)
	}()

	if recovered == nil {
		t.Fatal("NewWatcher did not panic on a nil deps")
	}
	msg, ok := recovered.(string)
	if !ok {
		t.Fatalf("recovered value is %T (%v), want a string naming \"deps\"", recovered, recovered)
	}
	if !strings.Contains(msg, "deps") {
		t.Errorf("panic message %q does not name \"deps\"", msg)
	}
}

// TestHarvestNeedsNoDatabase is the load-bearing test for this task's
// "internal/harvest must not depend on the store directly" requirement:
// it drives a real Watcher, reading real files from disk, entirely
// through fakeWindowSource/fakeHarvestSink -- no *store.Store, no
// PostgreSQL, nothing that could fail because a database is unreachable.
func TestHarvestNeedsNoDatabase(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)

	windows := &fakeWindowSource{bySession: openWindowForMainSession(1)}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), harvest.NoDeps{}, nil)

	touched, err := w.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if touched != 1 {
		t.Fatalf("RunOnce touched %d files, want 1", touched)
	}
	if got := sink.totals[1].Main.Input; got != 109 {
		t.Errorf("tokens.main.input = %v, want 109", got)
	}
}

// TestRunOnceCommitsPerModelBucketAlongsideTotal is the watcher-level
// wiring check for task 22: the whole-thing-through-one-real-Watcher
// counterpart to attribute_test.go's TestAttributeBucketsTokensPerModel.
// The main-thread fixture is entirely claude-opus-5 (confirmed by reading
// it directly, not assumed), so the committed batch's "models" bucket for
// that one model must carry exactly the same total RunOnce already
// commits under the top-level "tokens" key -- proving encodePatches
// actually reaches HarvestSink with both, not just Total.
func TestRunOnceCommitsPerModelBucketAlongsideTotal(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)

	windows := &fakeWindowSource{bySession: openWindowForMainSession(1)}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), harvest.NoDeps{}, nil)

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	byModel, ok := sink.modelTotals[1]
	if !ok {
		t.Fatalf("no per-model totals committed for stage run 1: %v", sink.modelTotals)
	}
	opus, ok := byModel["claude-opus-5"]
	if !ok {
		t.Fatalf("no models[claude-opus-5] bucket committed: %v", byModel)
	}
	if opus.Main.Input != sink.totals[1].Main.Input {
		t.Errorf("models[claude-opus-5].main.input = %v, want %v (the whole-run total: every fixture record is claude-opus-5)", opus.Main.Input, sink.totals[1].Main.Input)
	}
}

// TestConsecutiveRunsOverUnchangedFileAddNothing is the coordinator's own
// first named test: two RunOnce calls with no new bytes between them
// must not add anything on the second call.
func TestConsecutiveRunsOverUnchangedFileAddNothing(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)

	windows := &fakeWindowSource{bySession: openWindowForMainSession(1)}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), harvest.NoDeps{}, nil)

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (first): %v", err)
	}
	firstCommitCount := sink.commitCount
	if firstCommitCount == 0 {
		t.Fatalf("first RunOnce committed nothing")
	}
	if got := sink.totals[1].Main.Input; got != 109 {
		t.Fatalf("tokens.main.input after first run = %v, want 109", got)
	}

	touched, err := w.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce (second): %v", err)
	}
	if touched != 0 {
		t.Fatalf("second RunOnce touched %d files, want 0 (nothing new to read)", touched)
	}
	if sink.commitCount != firstCommitCount {
		t.Fatalf("CommitHarvestBatch was called %d times total, want %d (unchanged file must not commit again)", sink.commitCount, firstCommitCount)
	}
	if got := sink.totals[1].Main.Input; got != 109 {
		t.Fatalf("tokens.main.input after second run = %v, want still 109", got)
	}
}

// TestFreshWatcherOverAlreadyHarvestedTranscriptAddsNothing is F10's own
// test: the restart guarantee TestRestartDoesNotDoubleCount used to prove
// against a local harvest-offsets.json file, now proved against the
// store instead -- the offset a fresh Watcher reads via
// HarvestSink.GetHarvestOffset comes from wherever a *previous* Watcher
// (a previous flowd process, in production) left it, not from
// anything this process remembers. A brand-new Watcher, built fresh
// (never having called RunOnce before) but pointed at a sink that
// already has this file's true end offset committed, must add nothing.
func TestFreshWatcherOverAlreadyHarvestedTranscriptAddsNothing(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)
	windows := func() *fakeWindowSource { return &fakeWindowSource{bySession: openWindowForMainSession(1)} }

	// A first Watcher plays the role of "the previous flowd process":
	// it harvests the file completely and then is discarded -- nothing
	// about it survives into the second Watcher below except what it
	// committed to the (shared) sink.
	sink := newFakeHarvestSink()
	firstWatcher := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows()), harvest.NoDeps{}, nil)
	if _, err := firstWatcher.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (first watcher): %v", err)
	}
	if got := sink.totals[1].Main.Input; got != 109 {
		t.Fatalf("tokens.main.input after first watcher = %v, want 109", got)
	}
	firstCommitCount := sink.commitCount

	// A brand-new Watcher value -- not the same Go struct, and carrying
	// no in-memory state of its own (this package no longer has any: no
	// OffsetState, no local file) -- built against the same sink, the way
	// a restarted flowd would reconnect to the same Postgres database.
	freshWatcher := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows()), harvest.NoDeps{}, nil)
	touched, err := freshWatcher.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce (fresh watcher): %v", err)
	}
	if touched != 0 {
		t.Fatalf("fresh watcher's RunOnce touched %d files, want 0 (the transcript was already fully harvested)", touched)
	}
	if sink.commitCount != firstCommitCount {
		t.Fatalf("CommitHarvestBatch was called %d times total, want %d (a fresh watcher must not re-add)", sink.commitCount, firstCommitCount)
	}
	if got := sink.totals[1].Main.Input; got != 109 {
		t.Fatalf("tokens.main.input after the fresh watcher = %v, want still 109 (no double-add)", got)
	}
}

// TestFailedCommitLeavesOffsetUnadvancedAndIsRetried is the coordinator's
// second named test and the direct behavioural guard for F1's fix: a
// commit that fails leaves the sink's offset for that file exactly where
// it was (CommitHarvestBatch's atomicity means a failed call changes
// nothing at all, not even partially), so the very next RunOnce reads
// and attempts the identical batch again -- and once the sink recovers,
// applies it exactly once.
func TestFailedCommitLeavesOffsetUnadvancedAndIsRetried(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)

	windows := &fakeWindowSource{bySession: openWindowForMainSession(1)}
	sink := newFakeHarvestSink()
	sink.failNextCommits = 1
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), harvest.NoDeps{}, nil)

	touched, err := w.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce (failing): %v", err)
	}
	if touched != 0 {
		t.Fatalf("RunOnce touched %d files despite the sink failing, want 0", touched)
	}
	if sink.commitCount != 0 {
		t.Fatalf("commitCount = %d after a failed commit, want 0", sink.commitCount)
	}
	if _, found, _ := sink.GetHarvestOffset(context.Background(), filepath.Join(dir, "session.jsonl")); found {
		t.Fatalf("offset was recorded despite the commit failing, want no row at all")
	}

	touched, err = w.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce (retry): %v", err)
	}
	if touched != 1 {
		t.Fatalf("retry RunOnce touched %d files, want 1 (the same batch, retried)", touched)
	}
	if sink.commitCount != 1 {
		t.Fatalf("commitCount = %d after the retry succeeded, want 1", sink.commitCount)
	}
	if got := sink.totals[1].Main.Input; got != 109 {
		t.Fatalf("tokens.main.input after the retry = %v, want 109 (applied exactly once, not lost and not doubled)", got)
	}
}

// TestOutageAcrossSeveralCyclesThenRecoveryMatchesCleanRun is the
// coordinator's third named test and the one that pins the whole
// guarantee: several RunOnce cycles across a simulated multi-cycle store
// outage, ending in recovery, must leave totals exactly equal to what a
// single uninterrupted run over the identical transcripts would produce
// -- no loss from an under-counted batch, no inflation from a
// double-applied one.
//
// This is deliberately built around *two* successful commits (main-thread
// bytes committed cleanly first, then an outage, then sidechain bytes
// committed once the sink recovers) rather than one eventual success
// after several failures for the whole file. An earlier version of this
// test ran every cycle before recovery against an offset stuck at 0 (no
// commit had ever succeeded yet), which collapsed the "several failed
// cycles then recovery" sequence into a single successful full-file read
// -- indistinguishable in outcome from TestHarvestNeedsNoDatabase, and
// unable to show a discrepancy under either a double-count or an
// under-count bug, because there was never more than one successful
// commit in the run to double up or skip a range between. Splitting the
// success into two genuinely separate commits, with real failures
// interposed between them, is what makes this test load-bearing: a bug
// that let the second commit start from the wrong offset (double-adding
// the main-thread bytes, or skipping the sidechain ones) would now show
// up directly in the final totals.
//
// The fixture crosses two files' worth of content (main-thread committed
// first, sidechain appended and committed second) specifically so a
// discrepancy in either direction would show up somewhere concrete: an
// under-count would leave main.input below 109 or sidechain.input below
// 8; a double-count would push either above its true value. This
// fixture's five distinct, non-round numbers per bucket also make
// accidental cancellation (a lost batch and a doubled batch of equal
// size) implausible.
func TestOutageAcrossSeveralCyclesThenRecoveryMatchesCleanRun(t *testing.T) {
	windows := func() *fakeWindowSource { return &fakeWindowSource{bySession: openWindowForMainSession(1)} }

	// The clean run: one uninterrupted RunOnce over the full, already-grown
	// file (main-thread bytes, then the appended sidechain bytes, present
	// from the start).
	cleanDir := t.TempDir()
	cleanPath := copyFixtureInto(t, cleanDir, "session.jsonl", mainThreadFixture)
	appendFixture(t, cleanPath, sidechainFixture)
	cleanSink := newFakeHarvestSink()
	cleanWatcher := harvest.NewWatcher(cleanDir, cleanSink, harvest.NewAttributor(windows()), harvest.NoDeps{}, nil)
	if _, err := cleanWatcher.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (clean): %v", err)
	}
	wantMainInput := cleanSink.totals[1].Main.Input
	wantSidechainInput := cleanSink.totals[1].Sidechain.Input
	if wantMainInput != 109 || wantSidechainInput != 8 {
		t.Fatalf("clean run totals = (main=%v, sidechain=%v), want (109, 8) -- fixture assumption broken", wantMainInput, wantSidechainInput)
	}

	// The outage-then-recovery run, built around two genuinely separate
	// successful commits (see this test's own doc comment for why one
	// success is not enough to be discriminating).
	outageDir := t.TempDir()
	outagePath := copyFixtureInto(t, outageDir, "session.jsonl", mainThreadFixture)
	outageSink := newFakeHarvestSink()
	outageWatcher := harvest.NewWatcher(outageDir, outageSink, harvest.NewAttributor(windows()), harvest.NoDeps{}, nil)

	// First successful commit: the main-thread bytes, cleanly, before any
	// outage begins.
	if _, err := outageWatcher.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (initial, main-thread): %v", err)
	}
	if outageSink.commitCount != 1 {
		t.Fatalf("commitCount after the initial run = %d, want 1", outageSink.commitCount)
	}
	if got := outageSink.totals[1].Main.Input; got != wantMainInput {
		t.Fatalf("tokens.main.input after the initial run = %v, want %v", got, wantMainInput)
	}

	// New data arrives, then the store goes down for three cycles. Every
	// attempt reads the same sidechain-only delta (the offset is stuck at
	// end-of-main-thread throughout, since none of these commits
	// succeed) and every attempt fails -- nothing is committed, and
	// nothing already committed (the main-thread total above) is
	// touched.
	appendFixture(t, outagePath, sidechainFixture)
	outageSink.failNextCommits = 3
	for range 3 {
		if _, err := outageWatcher.RunOnce(context.Background()); err != nil {
			t.Fatalf("RunOnce (during outage): %v", err)
		}
	}
	if outageSink.commitCount != 1 {
		t.Fatalf("commitCount during the outage = %d, want still 1 (no new commits succeeded)", outageSink.commitCount)
	}
	if got := outageSink.totals[1].Main.Input; got != wantMainInput {
		t.Fatalf("tokens.main.input during the outage = %v, want unchanged %v", got, wantMainInput)
	}

	// Recovery: the next cycle succeeds -- the second, genuinely separate
	// successful commit, covering exactly the sidechain bytes the first
	// commit never touched.
	if _, err := outageWatcher.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (recovery): %v", err)
	}
	if outageSink.commitCount != 2 {
		t.Fatalf("commitCount after recovery = %d, want 2 (the main-thread commit plus the sidechain commit)", outageSink.commitCount)
	}

	// One more clean cycle, matching what a real daemon's next tick would
	// do -- must add nothing further.
	if _, err := outageWatcher.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (after recovery): %v", err)
	}
	if outageSink.commitCount != 2 {
		t.Fatalf("commitCount after the settle cycle = %d, want still 2", outageSink.commitCount)
	}

	if got := outageSink.totals[1].Main.Input; got != wantMainInput {
		t.Errorf("tokens.main.input after outage+recovery = %v, want %v (matching the clean run)", got, wantMainInput)
	}
	if got := outageSink.totals[1].Sidechain.Input; got != wantSidechainInput {
		t.Errorf("tokens.sidechain.input after outage+recovery = %v, want %v (matching the clean run)", got, wantSidechainInput)
	}
	if got := outageSink.totals[1].Main.CacheRead; got != cleanSink.totals[1].Main.CacheRead {
		t.Errorf("tokens.main.cache_read after outage+recovery = %v, want %v", got, cleanSink.totals[1].Main.CacheRead)
	}
	if got := outageSink.totals[1].Sidechain.CacheCreation; got != cleanSink.totals[1].Sidechain.CacheCreation {
		t.Errorf("tokens.sidechain.cache_creation after outage+recovery = %v, want %v", got, cleanSink.totals[1].Sidechain.CacheCreation)
	}
}

// appendFixture appends the JSONL content of a second fixture onto an
// existing file, simulating a transcript a live harness is still writing
// to.
func appendFixture(t *testing.T, path, fixturePath string) {
	t.Helper()
	extra := readFixture(t, fixturePath)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open %s for append: %v", path, err)
	}
	defer f.Close()
	if _, err := f.Write(extra); err != nil {
		t.Fatalf("append to %s: %v", path, err)
	}
}

// TestDiscoverTranscriptsFindsNestedSubagentFiles guards the specific
// directory shape real Claude Code sessions use -- confirmed by reading
// live files under ~/.claude/projects/ before this package's fixtures
// were resynthesised with obviously-fake content (task 9's post-commit
// review, finding F5) -- <session>.jsonl at the project root plus
// <session>/subagents/agent-*.jsonl beside it, both ending in .jsonl and
// both required for sidechain usage to ever be attributed.
func TestDiscoverTranscriptsFindsNestedSubagentFiles(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)
	copyFixtureInto(t, dir, filepath.Join("session", "subagents", "agent-abc123.jsonl"), sidechainFixture)

	windows := &fakeWindowSource{bySession: openWindowForMainSession(1)}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), harvest.NoDeps{}, nil)

	touched, err := w.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if touched != 2 {
		t.Fatalf("RunOnce touched %d files, want 2 (the top-level session file and the nested subagent file)", touched)
	}
	if got := sink.totals[1].Sidechain.Input; got != 8 {
		t.Errorf("tokens.sidechain.input = %v, want 8 (from the nested subagents/ file)", got)
	}
}

// TestEncodePatchesCarriesDispatchDescriptors is encodePatches' own
// contract (KAN-201), driven through the one real entry point that calls
// it -- Watcher.RunOnce -- exactly the way TestRunOnceCommitsPerModelBucketAlongsideTotal
// above proves task 22's per-model encoding, since encodePatches itself
// is unexported and every test in this package is external. A subagent
// transcript's sidecar meta file, once RunOnce's per-file loop has read
// it via ReadDispatchMeta, must reach the committed patch's
// dispatches.<agentId> entry alongside that dispatch's own token
// figures -- descriptors and tokens travel together through one commit,
// even though they are read from two different files on disk.
func TestEncodePatchesCarriesDispatchDescriptors(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)
	subPath := copyFixtureInto(t, dir, filepath.Join("session", "subagents", "agent-abc123.jsonl"), sidechainFixture)
	metaPath := strings.TrimSuffix(subPath, ".jsonl") + ".meta.json"
	metaJSON := `{"agentType":"general-purpose","description":"Implement Task 3","model":"claude-sonnet-5","spawnDepth":1}`
	if err := os.WriteFile(metaPath, []byte(metaJSON), 0o644); err != nil {
		t.Fatalf("write meta sidecar: %v", err)
	}

	windows := &fakeWindowSource{bySession: openWindowForMainSession(1)}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), harvest.NoDeps{}, nil)

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	byDispatch, ok := sink.dispatchTotals[1]
	if !ok {
		t.Fatalf("no dispatch totals committed for stage run 1: %v", sink.dispatchTotals)
	}
	// sidechainFixture's own assistant lines all carry agentId
	// "agent-example0001" (confirmed by reading the fixture directly),
	// regardless of the transcript file's own basename above.
	db, ok := byDispatch["agent-example0001"]
	if !ok {
		t.Fatalf("no dispatches[agent-example0001] committed: %v", byDispatch)
	}
	if db.Tokens.Sidechain.Input != 8 {
		t.Errorf("dispatches[agent-example0001].tokens.sidechain.input = %v, want 8", db.Tokens.Sidechain.Input)
	}
	if db.AgentType != "general-purpose" {
		t.Errorf("dispatches[agent-example0001].agent_type = %q, want %q", db.AgentType, "general-purpose")
	}
	if db.Description != "Implement Task 3" {
		t.Errorf("dispatches[agent-example0001].description = %q, want %q", db.Description, "Implement Task 3")
	}
	if db.Model != "claude-sonnet-5" {
		t.Errorf("dispatches[agent-example0001].model = %q, want %q", db.Model, "claude-sonnet-5")
	}
	if db.SpawnDepth == nil || *db.SpawnDepth != "1" {
		t.Errorf("dispatches[agent-example0001].spawn_depth = %v, want \"1\"", db.SpawnDepth)
	}
}

// TestEncodePatchesPreservesGenuineSpawnDepthZero pins F5 (pass 1 of this
// change's own review panel): a top-level dispatch's sidecar genuinely
// carries "spawnDepth":0, and that must reach the committed patch as a
// present zero, distinct from TestEncodePatchesOmitsAbsentDescriptors'
// own no-sidecar case below, whose SpawnDepth must stay nil. Before the
// fix, DispatchBucket.SpawnDepth was a plain int with `omitempty`, so a
// genuine 0 and an absent value both encoded to nothing -- indistinguishable
// on the wire and, downstream, in the stored metrics bag.
func TestEncodePatchesPreservesGenuineSpawnDepthZero(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)
	subPath := copyFixtureInto(t, dir, filepath.Join("session", "subagents", "agent-abc123.jsonl"), sidechainFixture)
	metaPath := strings.TrimSuffix(subPath, ".jsonl") + ".meta.json"
	metaJSON := `{"agentType":"general-purpose","description":"Top-level dispatch","model":"claude-sonnet-5","spawnDepth":0}`
	if err := os.WriteFile(metaPath, []byte(metaJSON), 0o644); err != nil {
		t.Fatalf("write meta sidecar: %v", err)
	}

	windows := &fakeWindowSource{bySession: openWindowForMainSession(1)}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), harvest.NoDeps{}, nil)

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	byDispatch, ok := sink.dispatchTotals[1]
	if !ok {
		t.Fatalf("no dispatch totals committed for stage run 1: %v", sink.dispatchTotals)
	}
	db, ok := byDispatch["agent-example0001"]
	if !ok {
		t.Fatalf("no dispatches[agent-example0001] committed: %v", byDispatch)
	}
	if db.SpawnDepth == nil {
		t.Fatalf("dispatches[agent-example0001].spawn_depth = nil, want a present \"0\" (a real sidecar recorded it)")
	}
	if *db.SpawnDepth != "0" {
		t.Errorf("dispatches[agent-example0001].spawn_depth = %v, want \"0\"", *db.SpawnDepth)
	}
}

// TestSpawnDepthStaysConstantAcrossOrdinaryReSends is the RunOnce-level
// wiring proof that encodePatches re-sends a dispatch's descriptors,
// including spawn_depth, on every ordinary batch its transcript grows by
// (encodePatches' own "if hasMeta" branch runs on every commit, not
// once) -- the sidecar is present *from before the first cycle*, unlike
// TestBackfillsDispatchMetaWhenSidecarArrivesLate's late-arrival case
// (F4/F11).
//
// This test does NOT exercise jsonb_deep_add's own per-leaf,
// by-JSON-type sum-vs-replace rule any more: fakeHarvestSink's
// descriptor handling is deliberately dumb (its own doc comment, above),
// so this test cannot distinguish "replaced" from "summed" the way it
// once did as F12's regression test (pass 2 of this change's own review
// panel). That merge-semantics coverage now runs against the real
// migration function instead --
// internal/store/harvest_test.go's
// TestCommitHarvestBatchReplacesSpawnDepthStringAcrossTwoOrdinaryCommits
// (F19, this change's own review panel, relocating what had drifted
// twice as a hand-written fake: F12, then F15). What this test still
// pins is RunOnce's own wiring: a stable spawn_depth across repeated
// cycles is only meaningful alongside a call-count signal (F18, pass 4
// of this change's own review panel) -- see the cross-check below.
func TestSpawnDepthStaysConstantAcrossOrdinaryReSends(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)
	subPath := copyFixtureInto(t, dir, filepath.Join("session", "subagents", "agent-abc123.jsonl"), sidechainFixture)
	metaPath := strings.TrimSuffix(subPath, ".jsonl") + ".meta.json"
	metaJSON := `{"agentType":"general-purpose","description":"Recurring dispatch","model":"claude-sonnet-5","spawnDepth":1}`
	if err := os.WriteFile(metaPath, []byte(metaJSON), 0o644); err != nil {
		t.Fatalf("write meta sidecar: %v", err)
	}

	windows := &fakeWindowSource{bySession: openWindowForMainSession(1)}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), harvest.NoDeps{}, nil)

	// Cycle 1: sidecar already present, so descriptors -- including
	// spawn_depth -- are sent alongside this batch's tokens.
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 1): %v", err)
	}

	// Two more ordinary cycles, each with genuinely new transcript bytes
	// (never a "nothing new to read" cycle, which would route through
	// maybeBackfillDispatchMeta instead): every one re-sends the sidecar's
	// same constant spawn_depth:1.
	for cycle := 2; cycle <= 3; cycle++ {
		appendFixture(t, subPath, sidechainFixture)
		if _, err := w.RunOnce(context.Background()); err != nil {
			t.Fatalf("RunOnce (cycle %d): %v", cycle, err)
		}
	}

	// F18 (pass 4 of this change's own review panel): a stable final value
	// alone cannot distinguish "stayed constant because every commit
	// succeeded" from "stayed constant because cycles 2 and 3 both failed
	// and RunOnce silently swallowed it" (watcher.go's "harvest: commit
	// failed, will retry" branch never surfaces a per-path commit error
	// through RunOnce's return -- see fakeHarvestSink's own doc comment
	// above). The cross-check: four commits happened (main-thread once
	// plus sidechain once in cycle 1, then one more sidechain commit for
	// each of the two ordinary re-send cycles) -- if a commit had
	// silently failed instead, commitCount would be short and this would
	// catch it even though *db.SpawnDepth alone would not.
	if sink.commitCount != 4 {
		t.Fatalf("sink.commitCount = %d, want 4 (main-thread + sidechain in cycle 1, plus one more sidechain commit per ordinary re-send cycle) -- a lower count means a commit silently failed and the spawn_depth check below would not have caught it", sink.commitCount)
	}
	db, ok := sink.dispatchTotals[1]["agent-example0001"]
	if !ok {
		t.Fatalf("no dispatches[agent-example0001] committed: %v", sink.dispatchTotals[1])
	}
	if db.SpawnDepth == nil {
		t.Fatalf("dispatches[agent-example0001].spawn_depth = nil, want a present \"1\"")
	}
	if *db.SpawnDepth != "1" {
		t.Errorf("dispatches[agent-example0001].spawn_depth after 3 ordinary re-sends = %v, want \"1\" (constant, never summed)", *db.SpawnDepth)
	}
}

// TestEncodePatchesOmitsAbsentDescriptors is the negative half: a
// subagent transcript with no sidecar meta file must still get its
// tokens attributed and committed under dispatches.<agentId>, with every
// descriptor field left empty -- never invented -- exactly the rule
// ReadDispatchMeta's own doc comment states and DispatchBucket's absence
// case (attribute.go) requires.
func TestEncodePatchesOmitsAbsentDescriptors(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)
	copyFixtureInto(t, dir, filepath.Join("session", "subagents", "agent-abc123.jsonl"), sidechainFixture)
	// Deliberately no agent-abc123.meta.json sidecar written.

	windows := &fakeWindowSource{bySession: openWindowForMainSession(1)}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), harvest.NoDeps{}, nil)

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	byDispatch, ok := sink.dispatchTotals[1]
	if !ok {
		t.Fatalf("no dispatch totals committed for stage run 1: %v", sink.dispatchTotals)
	}
	db, ok := byDispatch["agent-example0001"]
	if !ok {
		t.Fatalf("no dispatches[agent-example0001] committed: %v", byDispatch)
	}
	if db.Tokens.Sidechain.Input != 8 {
		t.Errorf("dispatches[agent-example0001].tokens.sidechain.input = %v, want 8 (tokens attributed even with no sidecar)", db.Tokens.Sidechain.Input)
	}
	if db.AgentType != "" || db.Description != "" || db.Model != "" || db.SpawnDepth != nil {
		t.Errorf("dispatches[agent-example0001] descriptors = %+v, want all absent -- no meta sidecar exists, so nothing must be invented", db)
	}
}

// TestBackfillsDispatchMetaWhenSidecarArrivesLate is F4's own regression
// test (pass 1 of this change's own review panel): a sidecar written
// *after* its transcript has already been fully harvested -- so a later
// RunOnce cycle reads no new bytes for it at all -- must still get its
// descriptors committed once the sidecar shows up, without touching the
// tokens that were already committed in the earlier cycle. Before the
// fix, a "nothing new to read" cycle never called ReadDispatchMeta again
// for that path, so hasMeta=false was permanent.
func TestBackfillsDispatchMetaWhenSidecarArrivesLate(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)
	subPath := copyFixtureInto(t, dir, filepath.Join("session", "subagents", "agent-abc123.jsonl"), sidechainFixture)
	// Deliberately no sidecar yet -- it "arrives" only after the first
	// harvest cycle has already fully consumed this transcript.

	windows := &fakeWindowSource{bySession: openWindowForMainSession(1)}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), harvest.NoDeps{}, nil)

	// Cycle 1: tokens are attributed and committed; no sidecar exists yet,
	// so descriptors are absent -- exactly TestEncodePatchesOmitsAbsentDescriptors'
	// own case.
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 1): %v", err)
	}
	db := sink.dispatchTotals[1]["agent-example0001"]
	if db.Tokens.Sidechain.Input != 8 {
		t.Fatalf("cycle 1: dispatches[agent-example0001].tokens.sidechain.input = %v, want 8", db.Tokens.Sidechain.Input)
	}
	if db.AgentType != "" || db.SpawnDepth != nil {
		t.Fatalf("cycle 1: descriptors = %+v, want all absent (no sidecar yet)", db)
	}

	// Cycle 2, with no new transcript bytes: still no sidecar, so nothing
	// backfills yet -- the tokens must stay exactly where cycle 1 left
	// them, not re-attributed.
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 2, still no sidecar): %v", err)
	}
	db = sink.dispatchTotals[1]["agent-example0001"]
	if db.Tokens.Sidechain.Input != 8 {
		t.Fatalf("cycle 2: dispatches[agent-example0001].tokens.sidechain.input = %v, want 8 (unchanged)", db.Tokens.Sidechain.Input)
	}

	// The sidecar "arrives" now, well after its transcript was fully
	// harvested.
	metaPath := strings.TrimSuffix(subPath, ".jsonl") + ".meta.json"
	metaJSON := `{"agentType":"general-purpose","description":"Late sidecar","model":"claude-sonnet-5","spawnDepth":1}`
	if err := os.WriteFile(metaPath, []byte(metaJSON), 0o644); err != nil {
		t.Fatalf("write late meta sidecar: %v", err)
	}

	// Cycle 3, still no new transcript bytes: this is the cycle that must
	// notice the now-present sidecar and backfill its descriptors.
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 3, sidecar now present): %v", err)
	}
	db = sink.dispatchTotals[1]["agent-example0001"]
	if db.AgentType != "general-purpose" {
		t.Errorf("cycle 3: agent_type = %q, want %q", db.AgentType, "general-purpose")
	}
	if db.Description != "Late sidecar" {
		t.Errorf("cycle 3: description = %q, want %q", db.Description, "Late sidecar")
	}
	if db.Model != "claude-sonnet-5" {
		t.Errorf("cycle 3: model = %q, want %q", db.Model, "claude-sonnet-5")
	}
	if db.SpawnDepth == nil || *db.SpawnDepth != "1" {
		t.Errorf("cycle 3: spawn_depth = %v, want \"1\"", db.SpawnDepth)
	}
	// The critical guard: descriptors landed, but the token figures are
	// untouched -- backfilling must never re-read or re-add usage.
	if db.Tokens.Sidechain.Input != 8 {
		t.Errorf("cycle 3: tokens.sidechain.input = %v, want 8 (backfill must not double-count tokens)", db.Tokens.Sidechain.Input)
	}
}

// TestBackfillDispatchMetaPricesStageRunAfterCommit is F11's own
// regression test (pass 2 of this change's own review panel): a
// pricer-wired sibling of TestBackfillsDispatchMetaWhenSidecarArrivesLate
// above, which wires no pricer at all and so cannot see this class of
// defect. This dispatch's only ordinary pricing pass ran in cycle 1,
// while its model was still empty -- pricing.go's `if db.Model == ""
// { continue }` skipped it there, permanently, unless something prices
// this stage run again once a model actually lands. Before the fix,
// maybeBackfillDispatchMeta committed the backfilled descriptors but
// never called Pricer.Price, so that dispatch's cost_usd stayed absent
// forever even after its model became known.
func TestBackfillDispatchMetaPricesStageRunAfterCommit(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)
	subPath := copyFixtureInto(t, dir, filepath.Join("session", "subagents", "agent-abc123.jsonl"), sidechainFixture)
	// Deliberately no sidecar yet -- it "arrives" only after the first
	// harvest cycle has already fully consumed this transcript, exactly
	// like TestBackfillsDispatchMetaWhenSidecarArrivesLate.

	windows := &fakeWindowSource{bySession: openWindowForMainSession(1)}
	sink := newFakeHarvestSink()
	pricer := &fakePricer{}
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), pricingDeps{p: pricer}, nil)

	// Cycle 1: tokens committed for both files (the main-thread session
	// and the subagent transcript, both attributed to stage run 1, each
	// committed and priced separately by RunOnce's own per-file loop) --
	// no sidecar yet, so this dispatch's model is empty and pricing.go
	// leaves its dispatches.<agentId> bucket unpriced, but the stage run
	// itself is still priced once per file committed (the ordinary
	// commit path's own pricing pass, watcher.go:474-480).
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 1): %v", err)
	}
	for _, id := range pricer.priced {
		if id != 1 {
			t.Fatalf("pricer.priced after cycle 1 = %v, want every call for stage run 1", pricer.priced)
		}
	}
	baseline := len(pricer.priced)
	if baseline == 0 {
		t.Fatalf("pricer.priced after cycle 1 = %v, want at least one call", pricer.priced)
	}

	// Cycle 2, with no new transcript bytes: still no sidecar, so nothing
	// backfills yet, and nothing new to price either.
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 2, still no sidecar): %v", err)
	}
	if len(pricer.priced) != baseline {
		t.Fatalf("pricer.priced after cycle 2 = %v, want still exactly %d calls (nothing new committed)", pricer.priced, baseline)
	}

	// The sidecar "arrives" now, well after its transcript was fully
	// harvested.
	metaPath := strings.TrimSuffix(subPath, ".jsonl") + ".meta.json"
	metaJSON := `{"agentType":"general-purpose","description":"Late sidecar","model":"claude-sonnet-5","spawnDepth":1}`
	if err := os.WriteFile(metaPath, []byte(metaJSON), 0o644); err != nil {
		t.Fatalf("write late meta sidecar: %v", err)
	}

	// Cycle 3, still no new transcript bytes: this is the cycle that must
	// notice the now-present sidecar, backfill its descriptors, and price
	// this stage run one more time now that a model is finally on record
	// for the dispatch.
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 3, sidecar now present): %v", err)
	}
	if len(pricer.priced) != baseline+1 {
		t.Fatalf("pricer.priced after cycle 3 = %v, want %d calls (one more, for stage run 1, after the backfill commit)", pricer.priced, baseline+1)
	}
	if pricer.priced[len(pricer.priced)-1] != 1 {
		t.Fatalf("pricer.priced's last call after cycle 3 = %v, want stage run 1", pricer.priced)
	}
}

// TestBackfillSurvivesAnInterveningBatchWithNoDispatchDescriptors pins
// F29 (pass 7 of this change's own review panel): the "hasMeta" branch
// that clears a path's pendingDispatchMeta entries used to delete the
// *whole* per-path map, not only the stageRunIDs this particular batch's
// deltas actually carried descriptors for. A batch can legitimately have
// newOffset != offset while parsing zero assistant records at all --
// interleaved, non-assistant lines (tool_use/tool_result in a real
// transcript; here, the fixture's own non-"assistant"-typed "attachment"
// line) do this routinely, and CommitHarvestBatch's own doc comment
// documents deltas as possibly empty and still applying. Before the fix,
// such a batch -- landing after the sidecar has arrived, so hasMeta is
// true, but delivering no dispatch descriptors of its own -- wiped an
// earlier batch's still-pending entry outright, and
// maybeBackfillDispatchMeta never got a chance to revisit it: those
// descriptors were lost permanently, even though the sidecar was sitting
// right there the whole time.
func TestBackfillSurvivesAnInterveningBatchWithNoDispatchDescriptors(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)
	subPath := copyFixtureInto(t, dir, filepath.Join("session", "subagents", "agent-abc123.jsonl"), sidechainFixture)
	// Deliberately no sidecar yet.

	windows := &fakeWindowSource{bySession: openWindowForMainSession(1)}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), harvest.NoDeps{}, nil)

	// Cycle 1: tokens are attributed and committed; no sidecar exists yet,
	// so pendingDispatchMeta[subPath] records stage run 1's entry.
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 1): %v", err)
	}
	db := sink.dispatchTotals[1]["agent-example0001"]
	if db.Tokens.Sidechain.Input != 8 {
		t.Fatalf("cycle 1: dispatches[agent-example0001].tokens.sidechain.input = %v, want 8", db.Tokens.Sidechain.Input)
	}

	// The sidecar arrives now.
	metaPath := strings.TrimSuffix(subPath, ".jsonl") + ".meta.json"
	metaJSON := `{"agentType":"general-purpose","description":"Late sidecar","model":"claude-sonnet-5","spawnDepth":1}`
	if err := os.WriteFile(metaPath, []byte(metaJSON), 0o644); err != nil {
		t.Fatalf("write late meta sidecar: %v", err)
	}

	// Cycle 2: new bytes land on this same path, but they carry zero
	// assistant records -- a single non-"assistant"-typed line, the same
	// shape ReadNewRecords already skips (transcript.go's own
	// recordTypeAssistant check). newOffset != offset (there is new
	// content to consume), yet Attribute(records) returns an empty deltas
	// map, so this batch delivers no dispatch descriptors for stage run 1
	// at all -- even though hasMeta is now true.
	f, err := os.OpenFile(subPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open %s for append: %v", subPath, err)
	}
	nonAssistantLine := `{"parentUuid":"uuid-sub-0001","isSidechain":true,"agentId":"agent-example0001","attachment":{"type":"deferred_tools_delta","addedNames":["ExampleTool"],"addedLines":[],"removedNames":[],"readdedNames":[]},"type":"attachment","uuid":"uuid-sub-0099","timestamp":"2026-01-01T00:05:04.000Z","userType":"external","entrypoint":"cli","cwd":"/synthetic/project","sessionId":"session-main-a1b2c3","version":"0.0.0-fixture","gitBranch":"example-branch"}` + "\n"
	if _, err := f.WriteString(nonAssistantLine); err != nil {
		t.Fatalf("append non-assistant line to %s: %v", subPath, err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close %s: %v", subPath, err)
	}
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 2, intervening batch with no dispatch descriptors): %v", err)
	}
	// The intervening batch must not have re-priced or re-attributed
	// anything -- descriptors are still absent, tokens unchanged.
	db = sink.dispatchTotals[1]["agent-example0001"]
	if db.Tokens.Sidechain.Input != 8 {
		t.Fatalf("cycle 2: tokens.sidechain.input = %v, want unchanged 8", db.Tokens.Sidechain.Input)
	}
	if db.AgentType != "" {
		t.Fatalf("cycle 2: agent_type = %q, want still absent -- this batch delivered no dispatch descriptors", db.AgentType)
	}

	// Cycle 3, no new transcript bytes: this is the "nothing new to read"
	// branch that calls maybeBackfillDispatchMeta. Before the fix, cycle
	// 2's hasMeta=true had already wiped stage run 1's pending entry
	// outright, so this cycle would find nothing left to backfill and the
	// descriptors would stay absent forever, even with the sidecar sitting
	// right there.
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 3, backfill): %v", err)
	}
	db = sink.dispatchTotals[1]["agent-example0001"]
	if db.AgentType != "general-purpose" {
		t.Errorf("cycle 3: agent_type = %q, want %q -- the pending entry must have survived cycle 2's intervening batch", db.AgentType, "general-purpose")
	}
	if db.Model != "claude-sonnet-5" {
		t.Errorf("cycle 3: model = %q, want %q", db.Model, "claude-sonnet-5")
	}
	if db.SpawnDepth == nil || *db.SpawnDepth != "1" {
		t.Errorf("cycle 3: spawn_depth = %v, want \"1\"", db.SpawnDepth)
	}
	if db.Tokens.Sidechain.Input != 8 {
		t.Errorf("cycle 3: tokens.sidechain.input = %v, want unchanged 8 (backfill must not double-count tokens)", db.Tokens.Sidechain.Input)
	}
}

// fakePricer stands in for *store.Store.Price (harvest.Pricer): it
// records every stage run id it was asked to price, in call order, and
// -- when failIDs names one -- returns an error for that id instead of
// succeeding, so TestPricingFailureIsNotFatal can simulate exactly the
// case Pricer's own doc comment calls out (a pricing failure must be
// logged and skipped, never fatal to the harvest pass that triggered it).
type fakePricer struct {
	priced  []int64
	failIDs map[int64]bool
}

func (p *fakePricer) Price(_ context.Context, stageRunID int64) error {
	p.priced = append(p.priced, stageRunID)
	if p.failIDs[stageRunID] {
		return errors.New("fake pricer: pricing unavailable")
	}
	return nil
}

var _ harvest.Pricer = (*fakePricer)(nil)

// pricingDeps satisfies harvest.Deps by delegating Price to p and leaving
// every other method as harvest.NoDeps' no-op (KAN-173) -- the test
// sites that used to configure only harvest.WithPricer(pricer) collapse
// to this.
type pricingDeps struct {
	harvest.NoDeps
	p harvest.Pricer
}

func (d pricingDeps) Price(ctx context.Context, stageRunID int64) error {
	return d.p.Price(ctx, stageRunID)
}

// sessionBinderDeps satisfies harvest.Deps by delegating every
// SessionTokenBinder method to binder and leaving the rest as
// harvest.NoDeps' no-op (KAN-173) -- the test sites that used to
// configure only harvest.WithSessionTokenBinder(binder) collapse to
// this. Every method here is an explicit override, not an embedded
// interface, so it always wins over harvest.NoDeps' own promoted method
// of the same name regardless of embedding depth -- an embedded
// harvest.SessionTokenBinder field alongside harvest.NoDeps would instead
// make every overlapping method ambiguous and unpromoted, so the struct
// would silently fail to satisfy harvest.Deps at all.
type sessionBinderDeps struct {
	harvest.NoDeps
	binder harvest.SessionTokenBinder
}

func (d sessionBinderDeps) UnresolvedSessionTokens(ctx context.Context) (map[int64]string, error) {
	return d.binder.UnresolvedSessionTokens(ctx)
}

func (d sessionBinderDeps) BindSession(ctx context.Context, sessionToken string, sessionID string) (int64, error) {
	return d.binder.BindSession(ctx, sessionToken, sessionID)
}

func (d sessionBinderDeps) RecordSessionTokenGiveUp(ctx context.Context, token, reason string, at time.Time) error {
	return d.binder.RecordSessionTokenGiveUp(ctx, token, reason, at)
}

func (d sessionBinderDeps) PersistedGiveUps(ctx context.Context) ([]harvest.GiveUp, error) {
	return d.binder.PersistedGiveUps(ctx)
}

func (d sessionBinderDeps) MarkDispatchesUnattributedByID(ctx context.Context, ids []int64, reason string, candidates int) error {
	return d.binder.MarkDispatchesUnattributedByID(ctx, ids, reason, candidates)
}

func (d sessionBinderDeps) MarkDispatchesUnattributed(ctx context.Context, token, reason string, candidates int) error {
	return d.binder.MarkDispatchesUnattributed(ctx, token, reason, candidates)
}

// sessionBinderAndDispatchDeps extends sessionBinderDeps with the
// dispatch-grain pass's two dependencies -- the two watcher_test.go sites
// that used to configure both harvest.WithSessionTokenBinder and
// harvest.WithDispatchAttribution collapse to this. Its two explicit
// methods below win over sessionBinderDeps' own promoted
// harvest.NoDeps-sourced no-ops for the same reason sessionBinderDeps'
// own methods win over harvest.NoDeps: an explicit method always beats a
// promoted one, whatever its depth.
type sessionBinderAndDispatchDeps struct {
	sessionBinderDeps
	windows harvest.DispatchWindowSource
	sink    harvest.DispatchMetricsSink
}

func (d sessionBinderAndDispatchDeps) DispatchWindowsForSession(ctx context.Context, sessionID string) ([]harvest.DispatchWindow, error) {
	return d.windows.DispatchWindowsForSession(ctx, sessionID)
}

func (d sessionBinderAndDispatchDeps) MergeDispatchMetrics(ctx context.Context, dispatchID int64, patch json.RawMessage) error {
	return d.sink.MergeDispatchMetrics(ctx, dispatchID, patch)
}

// TestRunOncePricesTouchedStageRunsAfterCommit is task 23 step 5's own
// load-bearing test: the wiring assertion, not a test of Price's
// correctness in isolation (which stageruns_test.go already covers
// directly). This is the exact class of test task 23's own plan calls
// out as missing before this task -- "every pricing test seeds its own
// rates and calls Price directly ... nothing asserted that anything in
// production calls it" -- so it drives a real Watcher.RunOnce end to end
// and asserts the touched stage run was handed to Pricer, proving the
// wiring exists rather than merely that Price itself works.
//
// Written and run against pre-task-23 code (no WithPricer option, no
// pricer field, no call site in RunOnce) it fails to compile at all --
// exactly the "watch it fail against current code" this task's own
// instructions require, and a stronger signal than a runtime failure
// would have been: the wiring this test guards did not exist in any form.
func TestRunOncePricesTouchedStageRunsAfterCommit(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)

	windows := &fakeWindowSource{bySession: openWindowForMainSession(1)}
	sink := newFakeHarvestSink()
	pricer := &fakePricer{}
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), pricingDeps{p: pricer}, nil)

	touched, err := w.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if touched != 1 {
		t.Fatalf("RunOnce touched %d files, want 1", touched)
	}
	if len(pricer.priced) != 1 || pricer.priced[0] != 1 {
		t.Fatalf("pricer.priced = %v, want exactly [1] (stage run 1, priced once after its batch committed)", pricer.priced)
	}

	// A second RunOnce over the same, now fully-harvested file commits
	// nothing new (TestConsecutiveRunsOverUnchangedFileAddNothing already
	// covers this at the commit level) and so must not price again either
	// -- pricing rides on "this batch actually committed", not on "this
	// stage run exists".
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (second): %v", err)
	}
	if len(pricer.priced) != 1 {
		t.Errorf("pricer.priced after a second, empty RunOnce = %v, want still exactly one call", pricer.priced)
	}
}

// TestLostRaceSkipsCountAndPricing is F4's own test: a CommitHarvestBatch
// that reports (applied=false, err=nil) -- this call lost a race with a
// concurrent harvester for the same file, per HarvestSink's own doc
// comment -- must not increment RunOnce's touchedFiles count and must not
// hand anything to Pricer, with a pricer configured. Both would double
// something real: touchedFiles would overcount how many files this call
// actually advanced, and Price would be asked to price a stage run whose
// batch this call never actually committed (the concurrent winner's batch
// already has, or will, trigger its own Price call). Neutralising the
// `!applied` guard passes every other test in this file, since none of
// them ever put the fake sink into forceLoseRace -- this is the one that
// must fail without it.
func TestLostRaceSkipsCountAndPricing(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)

	windows := &fakeWindowSource{bySession: openWindowForMainSession(1)}
	sink := newFakeHarvestSink()
	sink.forceLoseRace = true
	pricer := &fakePricer{}
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), pricingDeps{p: pricer}, nil)

	touched, err := w.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if touched != 0 {
		t.Errorf("RunOnce touched %d files, want 0: this call lost the commit race, so nothing here should count as touched", touched)
	}
	if len(pricer.priced) != 0 {
		t.Errorf("pricer.priced = %v, want empty: Price must not run for a batch that never actually committed", pricer.priced)
	}
	if sink.commitCount != 0 {
		t.Fatalf("sink.commitCount = %d, want 0 (forceLoseRace must have prevented any state change)", sink.commitCount)
	}
}

// TestRunOnceWithNoPricerConfiguredStillCommits is the negative
// companion: a Watcher built with no WithPricer option (every call site
// before this task, and every other test in this file) must keep working
// exactly as before -- pricing is additive, not a new requirement on
// every caller.
func TestRunOnceWithNoPricerConfiguredStillCommits(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)

	windows := &fakeWindowSource{bySession: openWindowForMainSession(1)}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), harvest.NoDeps{}, nil)

	touched, err := w.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if touched != 1 {
		t.Fatalf("RunOnce touched %d files, want 1", touched)
	}
}

// TestPricingFailureIsNotFatal proves the other half of Pricer's own doc
// comment: a Price call that fails must not fail RunOnce, must not
// prevent other stage runs in the same batch from being priced, and must
// not stop the batch's commit from having already landed (which it always
// has by the time Price is ever called -- Price runs strictly after
// CommitHarvestBatch reports applied).
func TestPricingFailureIsNotFatal(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)
	copyFixtureInto(t, dir, filepath.Join("session", "subagents", "agent-abc123.jsonl"), sidechainFixture)

	// Two open windows for the two stage runs the two files' records will
	// land in, so this batch touches two distinct stage runs and
	// exercises "one fails, the other still gets priced" for real.
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		mainSessionID: {{
			StageRunID: 1,
			SessionID:  mainSessionID,
			StartedAt:  time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		}},
	}}
	sink := newFakeHarvestSink()
	pricer := &fakePricer{failIDs: map[int64]bool{1: true}}
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), pricingDeps{p: pricer}, nil)

	touched, err := w.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v, want nil -- a Pricer failure must not fail the harvest pass", err)
	}
	if touched != 2 {
		t.Fatalf("RunOnce touched %d files, want 2 (both files' batches still committed despite the pricing failure)", touched)
	}
	if len(pricer.priced) != 2 {
		t.Fatalf("pricer.priced = %v, want 2 calls (pricing is attempted for every touched stage run even after one call fails)", pricer.priced)
	}
	// The commit itself is unaffected by the pricing failure: the offset
	// advanced and the totals landed exactly as they would with no pricer
	// configured at all.
	if got := sink.totals[1].Main.Input; got != 109 {
		t.Errorf("tokens.main.input = %v, want 109 (the commit succeeded regardless of the pricing failure)", got)
	}
}

// --- KAN-172 task 2: sessionToken resolution ---------------------------------

// crossedTokenRun is one stage run's mutable state inside
// fakeSessionTokenStore -- exactly what UnresolvedSessionTokens, BindSession
// and WindowsForSession all need to answer consistently with each
// other, the same coupling a real *store.Store enforces through one
// shared stage_runs table. This is deliberately a tighter fake than
// fakeWindowSource plus a standalone SessionTokenBinder fake (used by most
// tests below): the crossed-session-token test needs a stage run's window to
// only exist once its session_id has actually been bound, exactly like
// a real WindowsForSession query keyed on session_id would behave, so a
// binding bug shows up as wrong attribution, not just a wrong map entry.
type crossedTokenRun struct {
	sessionToken string
	sessionID    *string
	startedAt    time.Time
}

// noopGiveUpAndAmbiguityRecorder is the shared no-op stand-in for the four
// SessionTokenBinder methods (task 6, tasks.md) that record give-ups and
// dispatch-grain ambiguity -- RecordSessionTokenGiveUp, PersistedGiveUps,
// MarkDispatchesUnattributedByID and MarkDispatchesUnattributed. Embedded
// by both fakeSessionTokenStore and togglableSessionTokenBinder below,
// neither of whose own tests drives a give-up or an ambiguity, so both
// need these four to exist only to keep satisfying the widened
// harvest.SessionTokenBinder, never to record anything.
type noopGiveUpAndAmbiguityRecorder struct{}

func (noopGiveUpAndAmbiguityRecorder) RecordSessionTokenGiveUp(_ context.Context, _, _ string, _ time.Time) error {
	return nil
}

func (noopGiveUpAndAmbiguityRecorder) PersistedGiveUps(_ context.Context) ([]harvest.GiveUp, error) {
	return nil, nil
}

func (noopGiveUpAndAmbiguityRecorder) MarkDispatchesUnattributedByID(_ context.Context, _ []int64, _ string, _ int) error {
	return nil
}

func (noopGiveUpAndAmbiguityRecorder) MarkDispatchesUnattributed(_ context.Context, _, _ string, _ int) error {
	return nil
}

// fakeSessionTokenStore implements both harvest.SessionTokenBinder and
// harvest.WindowSource over the same in-memory runs map.
type fakeSessionTokenStore struct {
	noopGiveUpAndAmbiguityRecorder
	runs map[int64]*crossedTokenRun
}

func (f *fakeSessionTokenStore) UnresolvedSessionTokens(_ context.Context) (map[int64]string, error) {
	out := make(map[int64]string)
	for id, r := range f.runs {
		if r.sessionID == nil {
			out[id] = r.sessionToken
		}
	}
	return out, nil
}

// BindSession binds every run sharing sessionToken that is not already
// bound -- reproducing store.Store.BindSession's own "one UPDATE, every
// matching row" shape (KAN-172, task 4b), not just whichever single run a
// caller happened to name.
func (f *fakeSessionTokenStore) BindSession(_ context.Context, sessionToken string, sessionID string) (int64, error) {
	var bound int64
	for _, r := range f.runs {
		if r.sessionToken != sessionToken || r.sessionID != nil {
			continue
		}
		sid := sessionID
		r.sessionID = &sid
		bound++
	}
	return bound, nil
}

func (f *fakeSessionTokenStore) WindowsForSession(_ context.Context, sessionID string) ([]harvest.Window, error) {
	var out []harvest.Window
	for id, r := range f.runs {
		if r.sessionID != nil && *r.sessionID == sessionID {
			out = append(out, harvest.Window{
				StageRunID: id,
				SessionID:  sessionID,
				StartedAt:  r.startedAt,
			})
		}
	}
	return out, nil
}

var (
	_ harvest.SessionTokenBinder = (*fakeSessionTokenStore)(nil)
	_ harvest.WindowSource       = (*fakeSessionTokenStore)(nil)
)

// TestCrossedSessionTokensBindEachRunToItsOwnSession is the crossed-session-token test
// design.md and tasks.md both single out as the point of this task: two
// transcripts, two stage runs, discovered in filename order that is
// deliberately the *opposite* of which session each sessionToken actually
// belongs to ("a-session-beta.jsonl" sorts and is processed before
// "b-session-alpha.jsonl", yet stage run 1's sessionToken lives in the alpha
// file). A resolver that bound by "whichever transcript is discovered or
// processed first/last" (design.md's rejected "newest transcript"
// alternative, generalised) rather than by matching the literal sessionToken
// would bind both runs to the same wrong session here; this test only
// passes because binding is driven by the sessionToken actually found in each
// transcript's own recorded command text.
//
// Binding and attribution are proven together, across two cycles: cycle
// 1 contains only the two `stage begin -session-token ...` marks (no other
// usage) and binds both runs; cycle 2 appends real usage to each
// transcript now that a window exists for each bound session, and the
// resulting per-run totals must reflect only that run's own session's
// tokens -- proving neither run received the other's usage, not merely
// that BindSession was called with the right arguments.
func TestCrossedSessionTokensBindEachRunToItsOwnSession(t *testing.T) {
	dir := t.TempDir()
	started := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)

	sessionStore := &fakeSessionTokenStore{runs: map[int64]*crossedTokenRun{
		1: {sessionToken: "mf-session-token-alpha", startedAt: started},
		2: {sessionToken: "mf-session-token-beta", startedAt: started},
	}}

	sessionAlphaPath := filepath.Join(dir, "b-session-alpha.jsonl")
	sessionBetaPath := filepath.Join(dir, "a-session-beta.jsonl")

	writeMark := func(path, sessionID, sessionToken string) {
		line := fmt.Sprintf(`{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":%q,"message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -stage do.tests -session-token %s -harness claude-code"}}]}}`+"\n", sessionID, sessionToken)
		if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	writeMark(sessionAlphaPath, "session-alpha", "mf-session-token-alpha")
	writeMark(sessionBetaPath, "session-beta", "mf-session-token-beta")

	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(sessionStore), sessionBinderDeps{binder: sessionStore}, nil)

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 1, binding): %v", err)
	}

	if got := sessionStore.runs[1].sessionID; got == nil || *got != "session-alpha" {
		t.Fatalf("stage run 1 session = %v, want session-alpha", got)
	}
	if got := sessionStore.runs[2].sessionID; got == nil || *got != "session-beta" {
		t.Fatalf("stage run 2 session = %v, want session-beta", got)
	}

	appendUsage := func(path, sessionID string, inputTokens int) {
		line := fmt.Sprintf(`{"type":"assistant","timestamp":"2025-12-01T00:00:02Z","sessionId":%q,"message":{"model":"claude-opus-5","usage":{"input_tokens":%d,"output_tokens":1}}}`+"\n", sessionID, inputTokens)
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatalf("open %s for append: %v", path, err)
		}
		defer f.Close()
		if _, err := f.WriteString(line); err != nil {
			t.Fatalf("append to %s: %v", path, err)
		}
	}
	appendUsage(sessionAlphaPath, "session-alpha", 111)
	appendUsage(sessionBetaPath, "session-beta", 222)

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 2, attribution): %v", err)
	}

	if got := sink.totals[1].Main.Input; got != 111 {
		t.Errorf("stage run 1 tokens.main.input = %d, want 111 (session-alpha's own usage, not session-beta's)", got)
	}
	if got := sink.totals[2].Main.Input; got != 222 {
		t.Errorf("stage run 2 tokens.main.input = %d, want 222 (session-beta's own usage, not session-alpha's)", got)
	}
}

// countingSessionTokenBinder is a simpler SessionTokenBinder-only fake for the tests
// below, which are about the resolution *lifecycle* (bounded give-up,
// ambiguity, already-bound) rather than about attribution -- they pair
// it with an empty fakeWindowSource, since no test here asserts on
// token totals.
type countingSessionTokenBinder struct {
	sessionToken string
	stageRunID   int64
	bound        map[int64]string
	bindCalls    int

	// giveUpState is this fake's own upserted give-up state, keyed by
	// token, reproducing RecordSessionTokenGiveUp's real upsert semantics
	// (store.Store.RecordSessionTokenGiveUp's own doc comment, giveups.go:
	// retries starts at 0 on a token's first recording and rises by one
	// on every later one). A test that wants to simulate "already given
	// up once before, in a prior process" seeds this directly, so the
	// very next call this fake receives increments from that seed rather
	// than from zero.
	giveUpState map[string]*harvest.GiveUp
	// giveUpCalls is one entry per RecordSessionTokenGiveUp call, in call
	// order, snapshotting giveUpState's value for that token right after
	// the call -- what the tests below assert reasons, tokens and retry
	// counts against.
	giveUpCalls []harvest.GiveUp

	// seededGiveUps is what PersistedGiveUps returns -- a test simulating
	// a binder already holding a persisted give-up (task 5's own words,
	// tasks.md) sets this before building the Watcher.
	seededGiveUps         []harvest.GiveUp
	persistedGiveUpsCalls int

	// unattributedCalls is one entry per MarkDispatchesUnattributedByID
	// call, in call order.
	unattributedCalls []unattributedCall

	// unattributedTokenCalls is one entry per MarkDispatchesUnattributed
	// (the token form) call, in call order. unattributedTokenErr, when
	// set, is what every such call returns -- TestGiveUpStampsItsOwnDispatches'
	// "stamp failure never blocks the give-up" subtest uses it to prove a
	// stamp error is reported and never stops RecordSessionTokenGiveUp's
	// own outcome from standing.
	unattributedTokenCalls []unattributedTokenCall
	unattributedTokenErr   error
}

// unattributedCall is one MarkDispatchesUnattributedByID call's
// arguments, predicted from store.Store.MarkDispatchesUnattributedByID
// (records.go) -- ids, not a session token, since a dispatch-grain
// ambiguity names specific rows (task 6's own corrections, tasks.md).
type unattributedCall struct {
	ids        []int64
	reason     string
	candidates int
}

// unattributedTokenCall is one MarkDispatchesUnattributed call's
// arguments -- the token form (task 6.1, tasks.md, "stamp the dispatches
// of a session that never bound"), keyed by session token rather than by
// dispatch id, since a give-up leaves every dispatch of that session
// uncosted, not just specific rows.
type unattributedTokenCall struct {
	token      string
	reason     string
	candidates int
}

func (c *countingSessionTokenBinder) UnresolvedSessionTokens(_ context.Context) (map[int64]string, error) {
	if _, ok := c.bound[c.stageRunID]; ok {
		return map[int64]string{}, nil
	}
	return map[int64]string{c.stageRunID: c.sessionToken}, nil
}

// BindSession takes sessionToken now, not a stage run id (KAN-172, task
// 4b) -- this single-run fake still records the result against its one
// stageRunID, since every test using it carries only one run per token.
func (c *countingSessionTokenBinder) BindSession(_ context.Context, sessionToken string, sessionID string) (int64, error) {
	c.bindCalls++
	if sessionToken != c.sessionToken {
		return 0, fmt.Errorf("counting session token binder: unknown token %q", sessionToken)
	}
	if c.bound == nil {
		c.bound = map[int64]string{}
	}
	if _, already := c.bound[c.stageRunID]; already {
		return 0, nil
	}
	c.bound[c.stageRunID] = sessionID
	return 1, nil
}

// RecordSessionTokenGiveUp reproduces store.Store.RecordSessionTokenGiveUp's
// own upsert: a token recorded for the first time starts at retries 0; a
// token already in giveUpState increments it, exactly as the store's own
// `ON CONFLICT ... retries = retries + 1` does (giveups.go).
func (c *countingSessionTokenBinder) RecordSessionTokenGiveUp(_ context.Context, token, reason string, _ time.Time) error {
	if c.giveUpState == nil {
		c.giveUpState = map[string]*harvest.GiveUp{}
	}
	g, ok := c.giveUpState[token]
	if !ok {
		g = &harvest.GiveUp{Token: token}
		c.giveUpState[token] = g
	} else {
		g.Retries++
	}
	g.Reason = reason
	c.giveUpCalls = append(c.giveUpCalls, *g)
	return nil
}

// PersistedGiveUps returns seededGiveUps, recording that it was called --
// what TestPersistedGiveUpIsRetriedOnStart and TestRetryStillBounded
// assert against to prove a fresh Watcher actually reads persisted
// give-ups rather than merely happening to retry by coincidence of a
// fresh in-memory map.
func (c *countingSessionTokenBinder) PersistedGiveUps(_ context.Context) ([]harvest.GiveUp, error) {
	c.persistedGiveUpsCalls++
	return c.seededGiveUps, nil
}

// MarkDispatchesUnattributedByID records its call rather than doing
// anything with ids, reason or candidates -- this fake is never asked to
// answer a later read of them, only to prove they were passed.
func (c *countingSessionTokenBinder) MarkDispatchesUnattributedByID(_ context.Context, ids []int64, reason string, candidates int) error {
	c.unattributedCalls = append(c.unattributedCalls, unattributedCall{ids: ids, reason: reason, candidates: candidates})
	return nil
}

// MarkDispatchesUnattributed records its call (the token form, task 6.1)
// and returns unattributedTokenErr, letting a test simulate a stamp
// failure without that failure ever reaching RecordSessionTokenGiveUp's
// own outcome.
func (c *countingSessionTokenBinder) MarkDispatchesUnattributed(_ context.Context, token, reason string, candidates int) error {
	c.unattributedTokenCalls = append(c.unattributedTokenCalls, unattributedTokenCall{token: token, reason: reason, candidates: candidates})
	return c.unattributedTokenErr
}

var _ harvest.SessionTokenBinder = (*countingSessionTokenBinder)(nil)

// TestSessionTokenResolvesOnALaterCycleWithinTheBound is task 2 step 3's first
// named test: a sessionToken that appears several cycles after the mark, but
// still well inside the bounded window, must still bind -- the bound
// exists to cap wasted work on a sessionToken that will never appear, not to
// shrink the window a live but slow-flushing transcript has to land in.
func TestSessionTokenResolvesOnALaterCycleWithinTheBound(t *testing.T) {
	dir := t.TempDir()
	binder := &countingSessionTokenBinder{sessionToken: "mf-later-cycle", stageRunID: 42}
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, nil)

	for i := range 3 {
		if _, err := w.RunOnce(context.Background()); err != nil {
			t.Fatalf("RunOnce (cycle %d): %v", i, err)
		}
	}
	if binder.bindCalls != 0 {
		t.Fatalf("bindCalls = %d before the sessionToken ever appears, want 0", binder.bindCalls)
	}

	path := filepath.Join(dir, "late.jsonl")
	line := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-late","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -session-token mf-later-cycle"}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (binding cycle): %v", err)
	}
	if binder.bindCalls != 1 {
		t.Fatalf("bindCalls = %d, want 1", binder.bindCalls)
	}
	if binder.bound[42] != "session-late" {
		t.Fatalf("bound session = %q, want session-late", binder.bound[42])
	}
}

// TestSessionTokenMatchedByTwoSessionsRecordsNoSessionAndStopsRetrying is the
// ambiguity scenario: the same sessionToken found in two different sessions'
// transcripts must never be bound to either -- "a session is never
// guessed" (design.md) -- must be reported (logged), and must not be
// re-reported or retried on a later cycle: ambiguity cannot resolve
// itself by waiting, so it is treated as terminal exactly like the
// bounded give-up.
func TestSessionTokenMatchedByTwoSessionsRecordsNoSessionAndStopsRetrying(t *testing.T) {
	dir := t.TempDir()
	binder := &countingSessionTokenBinder{sessionToken: "mf-ambiguous", stageRunID: 7}

	writeMark := func(name, sessionID string) {
		line := fmt.Sprintf(`{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":%q,"message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -session-token mf-ambiguous"}}]}}`+"\n", sessionID)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(line), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	writeMark("one.jsonl", "session-one")
	writeMark("two.jsonl", "session-two")

	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
	sink := newFakeHarvestSink()
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, logger)

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if binder.bindCalls != 0 {
		t.Fatalf("bindCalls = %d, want 0: an ambiguous sessionToken must never be bound", binder.bindCalls)
	}
	if !strings.Contains(logBuf.String(), "more than one session") {
		t.Fatalf("log output = %q, want a message reporting the ambiguity", logBuf.String())
	}

	logBuf.Reset()
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (second cycle): %v", err)
	}
	if binder.bindCalls != 0 {
		t.Fatalf("bindCalls after a second cycle = %d, want still 0", binder.bindCalls)
	}
	if logBuf.Len() != 0 {
		t.Fatalf("log output after the second cycle = %q, want empty: ambiguity is reported once, not every cycle", logBuf.String())
	}
}

// TestSessionTokenStopsBeingScannedAfterBoundedGiveUp is task 2 step 3's
// "actually stops" test: a sessionToken whose transcript never appears must be
// abandoned after maxSessionTokenResolutionCycles cycles (60, watcher.go),
// logged exactly once, and -- the guard that would catch a give-up that
// merely stops *logging* rather than stops *looking* -- must not bind
// even if its transcript line shows up afterward.
func TestSessionTokenStopsBeingScannedAfterBoundedGiveUp(t *testing.T) {
	dir := t.TempDir()
	binder := &countingSessionTokenBinder{sessionToken: "mf-never-appears", stageRunID: 99}
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
	sink := newFakeHarvestSink()
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, logger)

	// 60 empty cycles: nothing to find, so the give-up bound is reached
	// on the last of these.
	for i := range 60 {
		if _, err := w.RunOnce(context.Background()); err != nil {
			t.Fatalf("RunOnce (cycle %d): %v", i, err)
		}
	}
	if got := strings.Count(logBuf.String(), "giving up"); got != 1 {
		t.Fatalf("give-up logged %d times after 60 cycles, want exactly 1: %q", got, logBuf.String())
	}

	// The sessionToken's transcript line finally appears -- too late. It must
	// not be looked for any more.
	path := filepath.Join(dir, "toolate.jsonl")
	line := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-toolate","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -session-token mf-never-appears"}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	logBuf.Reset()

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (after give-up): %v", err)
	}
	if binder.bindCalls != 0 {
		t.Fatalf("bindCalls = %d, want 0: a sessionToken is never resolved once this Watcher has abandoned it", binder.bindCalls)
	}
	if strings.Contains(logBuf.String(), "giving up") {
		t.Fatalf("give-up logged again on a later cycle, want it logged exactly once ever: %q", logBuf.String())
	}
}

// TestAlreadyBoundRunIsNeverReconsidered is the one-way-binding guard at
// the Watcher level: a stage run UnresolvedSessionTokens no longer reports
// (because its session_id is already set -- store.Store.UnresolvedSessionTokens'
// own WHERE clause, stageruns.go) must never be handed to BindSession
// again, even when a transcript carrying its sessionToken is right there to be
// read. Binding is one-way (design.md); this is what "never" means in
// practice, one layer up from the store's own WHERE-guarded UPDATE.
func TestAlreadyBoundRunIsNeverReconsidered(t *testing.T) {
	dir := t.TempDir()
	binder := &countingSessionTokenBinder{sessionToken: "mf-already-bound", stageRunID: 5}
	binder.bound = map[int64]string{5: "session-original"}

	path := filepath.Join(dir, "session.jsonl")
	line := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-imposter","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -session-token mf-already-bound"}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, nil)

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if binder.bindCalls != 0 {
		t.Fatalf("bindCalls = %d, want 0: an already-bound run must never be reconsidered", binder.bindCalls)
	}
	if binder.bound[5] != "session-original" {
		t.Fatalf("bound session = %q, want unchanged session-original", binder.bound[5])
	}
}

// TestBindMarkAndFirstUsageInSameBatchAreBothAttributed is the defect a
// post-commit review found in this task's first pass: Claude Code
// flushes a turn's transcript entries together, so the `stage begin
// -session-token ...` mark and that same turn's own usage arrive in the *same*
// newly-read batch -- not a later one. A Watcher that attributes a
// batch and commits it (advancing the offset) before sessionToken resolution
// has had a chance to bind the session that batch just revealed loses
// that usage permanently: nothing ever re-reads bytes the offset has
// already moved past. Because a mark's own turn is frequently a stage's
// largest, this was not a rare edge case -- it was every stage's first
// turn, every time.
//
// This test writes both the mark (as part of an assistant message that
// also carries that message's own usage) and a second message's usage
// into one file, present from the very first read, then drives two
// cycles: the first must bind without losing the batch that revealed
// the sessionToken, and the second must show the usage that batch carried, not
// zero.
func TestBindMarkAndFirstUsageInSameBatchAreBothAttributed(t *testing.T) {
	dir := t.TempDir()
	started := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	sessionStore := &fakeSessionTokenStore{runs: map[int64]*crossedTokenRun{
		1: {sessionToken: "mf-same-batch", startedAt: started},
	}}

	path := filepath.Join(dir, "session.jsonl")
	content := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-same-batch","message":{"model":"claude-opus-5","usage":{"input_tokens":5,"output_tokens":1},"content":[{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -session-token mf-same-batch"}}]}}` + "\n" +
		`{"type":"assistant","timestamp":"2025-12-01T00:00:02Z","sessionId":"session-same-batch","message":{"model":"claude-opus-5","usage":{"input_tokens":333,"output_tokens":1}}}` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(sessionStore), sessionBinderDeps{binder: sessionStore}, nil)

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 1): %v", err)
	}
	if got := sessionStore.runs[1].sessionID; got == nil || *got != "session-same-batch" {
		t.Fatalf("stage run 1 session after cycle 1 = %v, want session-same-batch", got)
	}

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 2): %v", err)
	}

	if got := sink.totals[1].Main.Input; got != 338 {
		t.Errorf("tokens.main.input = %d, want 338 (5 from the mark's own turn + 333 from the same batch's second message -- neither lost)", got)
	}
}

// TestSecondMarkOfAnAlreadyBoundTokenCommitsInTheSameCycle is KAN-172 task
// 4b's own pinning test: "a run whose token is already bound makes a
// further mark, and its batch is committed in the same cycle rather than
// withheld." Under the per-mark shape this replaces, this test fails --
// confirmed directly, before this rework, against the equivalent
// pre-4b fixture: two stage runs sharing one correlator value (run 1
// already bound to "session-shared", run 2 still unbound) drove
// UnresolvedNonces to keep treating run 2's mark as needing its own
// resolution, so its batch -- carrying real usage in the very same
// message as the mark -- was withheld rather than committed.
//
// Task 4b's mechanism is what changes the outcome here, at the store
// layer this test's fake stands in for: insertStageRunAndSupersede resolves
// session_id at insert time from an already-bound token, so a stage run
// created with a token that has already resolved never enters
// UnresolvedSessionTokens at all, and RunOnce never has a reason to
// withhold its batch.
func TestSecondMarkOfAnAlreadyBoundTokenCommitsInTheSameCycle(t *testing.T) {
	dir := t.TempDir()
	started := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	boundSessionID := "session-shared"

	// Stage run 2's session_id is already resolved -- reproducing what
	// store.Store.insertStageRunAndSupersede now does at insert time for a run whose
	// token has already bound (task 4b), rather than something this
	// Watcher-level fake resolves itself. This fake's UnresolvedSessionTokens
	// only ever reports a run whose sessionID is nil, so run 2 is never
	// pending in the first place.
	// Run 1's window starts an hour after the mark's own message timestamp
	// -- a later stage in the same, already-bound session -- so only run
	// 2's window is open when that message arrives; this isolates the
	// attribution to run 2 without depending on any attempt-based
	// tie-break between two simultaneously open windows.
	sessionStore := &fakeSessionTokenStore{runs: map[int64]*crossedTokenRun{
		1: {sessionToken: "mf-shared", sessionID: &boundSessionID, startedAt: started.Add(time.Hour)},
		2: {sessionToken: "mf-shared", sessionID: &boundSessionID, startedAt: started},
	}}

	path := filepath.Join(dir, "session.jsonl")
	content := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-shared","message":{"model":"claude-opus-5","usage":{"input_tokens":77,"output_tokens":1},"content":[{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -session-token mf-shared"}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(sessionStore), sessionBinderDeps{binder: sessionStore}, nil)

	touched, err := w.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if touched != 1 {
		t.Fatalf("RunOnce touched %d files, want 1 (committed in the same cycle, not withheld)", touched)
	}
	if got := sink.totals[2].Main.Input; got != 77 {
		t.Fatalf("stage run 2 tokens.main.input = %d, want 77 committed in the same cycle (not withheld)", got)
	}
}

// togglableSessionTokenBinder is a countingSessionTokenBinder with one
// extra knob: whether its one stage run is reported as pending at all.
// The F4 silent-misattribution test (below) needs to simulate a run whose
// token was NOT pending during an earlier cycle (so a transcript line
// naming it was read and its offset consumed as ordinary content, never
// scanned as a candidate) and only becomes pending afterward -- exactly
// the shape the live daemon was in immediately after F1 was fixed and
// restarted: the mark's own bytes were already behind the read offset
// from before the binder existed, and only later-written bytes were ever
// scanned against the newly-pending token.
type togglableSessionTokenBinder struct {
	noopGiveUpAndAmbiguityRecorder
	sessionToken string
	stageRunID   int64
	pending      bool
	bound        map[int64]string
	boundOrder   []string
	bindCalls    int
}

func (b *togglableSessionTokenBinder) UnresolvedSessionTokens(_ context.Context) (map[int64]string, error) {
	if !b.pending {
		return map[int64]string{}, nil
	}
	if _, ok := b.bound[b.stageRunID]; ok {
		return map[int64]string{}, nil
	}
	return map[int64]string{b.stageRunID: b.sessionToken}, nil
}

func (b *togglableSessionTokenBinder) BindSession(_ context.Context, sessionToken string, sessionID string) (int64, error) {
	b.bindCalls++
	if sessionToken != b.sessionToken {
		return 0, fmt.Errorf("togglable session token binder: unknown token %q", sessionToken)
	}
	if b.bound == nil {
		b.bound = map[int64]string{}
	}
	b.bound[b.stageRunID] = sessionID
	b.boundOrder = append(b.boundOrder, sessionID)
	return 1, nil
}

var _ harvest.SessionTokenBinder = (*togglableSessionTokenBinder)(nil)

// TestCommandMerelyMentioningTokenDoesNotBind is F4's first required
// test, written to fail against the pre-fix matcher: matchSessionTokens
// matched by bare strings.Contains, so a diagnostic command that only
// MENTIONS a pending sessionToken -- a grep for it, a log dump, a database
// query, an echo -- counted exactly like a genuine
// `stage begin -session-token ...` invocation. Here the sessionToken is
// pending and never appears in any genuine mark at all; the only
// occurrence anywhere is a grep's argument. A correct matcher must never
// bind on that alone.
func TestCommandMerelyMentioningTokenDoesNotBind(t *testing.T) {
	dir := t.TempDir()
	binder := &togglableSessionTokenBinder{sessionToken: "mf-mention-only", stageRunID: 101, pending: true}
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, nil)

	path := filepath.Join(dir, "mentioner.jsonl")
	line := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-mentioner","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"grep 'mf-mention-only' ~/.claude/projects/*/*.jsonl"}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if binder.bindCalls != 0 {
		t.Fatalf("bindCalls = %d, want 0: a command that merely mentions the sessionToken must never bind it (grep is not a mark)", binder.bindCalls)
	}
}

// TestMentionAfterOwnMarkIsConsumedDoesNotMisattribute is F4's severity
// proof, not just its symptom: session A genuinely marks a stage with
// sessionToken; A's own mark bytes are read and committed on a cycle
// before the token is pending at all (mirroring the live sequence -- the
// SessionTokenBinder was wired in later, after F1 was fixed and the
// daemon restarted, so the file offset had already moved past the real
// mark). Once the token becomes pending, the only occurrence left for
// this Watcher to find is session B's transcript, which merely MENTIONS
// A's token in a diagnostic command -- never a mark of its own. A matcher
// that matches by bare substring has nothing else to bind to and silently
// attributes A's stage run to B; the fix must leave it unbound instead.
func TestMentionAfterOwnMarkIsConsumedDoesNotMisattribute(t *testing.T) {
	dir := t.TempDir()
	token := "mf-k172-silent-misattribution"
	binder := &togglableSessionTokenBinder{sessionToken: token, stageRunID: 202, pending: false}
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, nil)

	// Cycle 1: session A's real mark is written and read while the token
	// is NOT YET pending (matchSessionTokens returns early on an empty
	// pending map), so the offset commits past it exactly as it did on
	// the live daemon before its SessionTokenBinder was wired in.
	sessionAPath := filepath.Join(dir, "session-a.jsonl")
	markLine := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-a","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -stage do.tests -session-token ` + token + ` -harness claude-code"}}]}}` + "\n"
	if err := os.WriteFile(sessionAPath, []byte(markLine), 0o644); err != nil {
		t.Fatalf("write %s: %v", sessionAPath, err)
	}
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 1, pre-pending): %v", err)
	}
	if binder.bindCalls != 0 {
		t.Fatalf("bindCalls after cycle 1 = %d, want 0 (token was not yet pending)", binder.bindCalls)
	}

	// The token becomes pending now -- session A's mark bytes are already
	// behind the read offset and will never be scanned again.
	binder.pending = true

	// Cycle 2: session B's transcript, new to this cycle, merely mentions
	// A's token in a diagnostic command.
	sessionBPath := filepath.Join(dir, "session-b.jsonl")
	mentionLine := `{"type":"assistant","timestamp":"2025-12-01T00:00:02Z","sessionId":"session-b","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"grep '` + token + `' /tmp/dispatch.log"}}]}}` + "\n"
	if err := os.WriteFile(sessionBPath, []byte(mentionLine), 0o644); err != nil {
		t.Fatalf("write %s: %v", sessionBPath, err)
	}
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 2, mention only): %v", err)
	}

	if binder.bindCalls != 0 {
		t.Fatalf("bindCalls after cycle 2 = %d, want 0: stage run 202 must NOT bind to session-b, which only mentioned the token", binder.bindCalls)
	}
	for _, sid := range binder.boundOrder {
		if sid == "session-b" {
			t.Fatalf("stage run 202 was bound to %q, a session that only mentioned the token in a grep -- silent misattribution", sid)
		}
	}
}

// runMarkCommand builds a one-file transcript carrying a single Bash
// tool_use command and runs one Watcher.RunOnce cycle over it, returning
// the binder so callers can assert on bindCalls/bound. Shared by every
// shape test below (KAN-174) so each table entry is only the command text
// and the assertion, not the transcript/watcher plumbing.
func runMarkCommand(t *testing.T, token string, stageRunID int64, sessionID, command string) *togglableSessionTokenBinder {
	t.Helper()
	dir := t.TempDir()
	binder := &togglableSessionTokenBinder{sessionToken: token, stageRunID: stageRunID, pending: true}
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, nil)

	path := filepath.Join(dir, "session.jsonl")
	encoded, err := json.Marshal(command)
	if err != nil {
		t.Fatalf("encode command: %v", err)
	}
	line := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"` + sessionID + `","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":` + string(encoded) + `}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	return binder
}

// TestMarkRecognizedWhereverItSitsInTheCommand is KAN-174's regression
// corpus: real marks are emitted inside shell blocks carrying variable
// assignments, directory changes and other statements before the
// `flow stage begin`/`stage end` invocation itself, per design.md
// ("recognise a mark by its invocation, not by its position") and the
// spec's "A mark inside a larger shell block" scenario. Every shape here
// was measured against the merged (pre-fix) matcher and found rejected --
// the multi-line and variable-assignment cases in particular are the
// defect this change repairs, not incidental coverage.
func TestMarkRecognizedWhereverItSitsInTheCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{
			name:    "bare invocation",
			command: "flow stage begin -stage do.tests -session-token TOKEN -harness claude-code",
		},
		{
			name:    "behind cd &&",
			command: "cd /repo && flow stage begin -stage do.tests -session-token TOKEN -harness claude-code",
		},
		{
			name:    "after variable assignments on the same line",
			command: "N=kan; T=mf-x; flow stage begin -stage do.tests -session-token TOKEN -harness claude-code",
		},
		{
			name:    "after variable assignments and a cd, same line",
			command: "N=kan; T=mf-x; cd /repo && flow stage begin -stage do.tests -session-token TOKEN -harness claude-code",
		},
		{
			name:    "on a later line after variable assignments and a bare cd",
			command: "N=kan; T=mf-x; cd /repo\nflow stage begin -stage do.tests -session-token TOKEN -harness claude-code",
		},
		{
			name:    "on a later line of a multi-statement block, one statement per line",
			command: "WT=/repo\ncd \"$WT\"\nflow stage begin -stage do.tests -session-token TOKEN -harness claude-code",
		},
		{
			name:    "flags reordered, -harness before -session-token",
			command: "cd /repo && flow stage begin -harness claude-code -stage do.tests -session-token TOKEN change-name",
		},
		{
			name:    "token quoted",
			command: `flow stage begin -stage do.tests -session-token 'TOKEN' -harness claude-code`,
		},
		{
			name:    "-session-token=value form",
			command: "flow stage begin -stage do.tests -session-token=TOKEN -harness claude-code",
		},
		{
			name:    "-session-token=value form, quoted",
			command: `flow stage begin -stage do.tests -session-token="TOKEN" -harness claude-code`,
		},
		{
			name:    "line-continuation form (backslash-newline), as skills/myflow-do/SKILL.md emits",
			command: "flow stage begin -command '/myflow-do' \\\n  -stage do.workspace-export \\\n  -harness claude-code \\\n  -session-token TOKEN \\\n  change-name",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := fmt.Sprintf("mf-k174-shape-%d", i)
			sessionID := fmt.Sprintf("session-shape-%d", i)
			command := strings.ReplaceAll(tt.command, "TOKEN", token)
			binder := runMarkCommand(t, token, int64(i), sessionID, command)
			if binder.bindCalls != 1 {
				t.Fatalf("bindCalls = %d, want 1: a genuine mark must bind regardless of where it sits in the command (%s)", binder.bindCalls, tt.name)
			}
			if binder.bound[int64(i)] != sessionID {
				t.Fatalf("bound session = %q, want %q", binder.bound[int64(i)], sessionID)
			}
		})
	}
}

// TestCommandsThatOnlyMentionTokenNeverBind is F4's negative corpus,
// broadened for KAN-174: removing the position anchor must not reopen the
// bare-containment defect F4 fixed. None of these commands perform the
// invocation -- they only carry the token as data -- so the
// `-session-token`-value check (requirement 2 of isSessionMarkCommand)
// must still refuse every one of them.
func TestCommandsThatOnlyMentionTokenNeverBind(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{
			name:    "grep for the token",
			command: "grep 'TOKEN' ~/.claude/projects/*/*.jsonl",
		},
		{
			name:    "psql query naming the token",
			command: `psql -U flow -d flow -tAc "SELECT id FROM stage_runs WHERE session_token='TOKEN'"`,
		},
		{
			name:    "piped cat",
			command: "cat /tmp/dispatch.log | grep TOKEN",
		},
		{
			name:    "bare mention with no stage-begin/end at all",
			command: "echo TOKEN",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := fmt.Sprintf("mf-k174-mention-%d", i)
			sessionID := fmt.Sprintf("session-mention-%d", i)
			command := strings.ReplaceAll(tt.command, "TOKEN", token)
			binder := runMarkCommand(t, token, int64(1000+i), sessionID, command)
			if binder.bindCalls != 0 {
				t.Fatalf("bindCalls = %d, want 0: a command that only mentions the token must never bind (%s)", binder.bindCalls, tt.name)
			}
		})
	}
}

// TestEchoedMarkExampleIsAnAcceptedResidual is design.md's "accept the
// echoed-example false positive, and say so" decision, locked in as a
// test rather than left as prose: dropping the position anchor
// (isSessionMarkCommand's own doc comment) means a command that only
// PRINTS a mark-shaped string -- echo "flow stage begin ...
// -session-token <token> ..." -- now satisfies both remaining
// requirements (contains "stage begin"/"stage end"; binds the token as
// -session-token's value) and binds, exactly like a genuine invocation.
//
// This is deliberate, not a regression of F5: the alternative is a
// position anchor, and every anchor tried (KAN-172's `cd ... &&`-only
// anchor, chosen alternative "anchor flow at a command position") either
// re-rejects a real multi-line mark or is defeated by the same newline the
// real shapes already contain. The asymmetry is the argument (design.md):
// a false negative here is silent and total -- no stage run ever binds,
// exactly the defect this change repairs -- while this false positive
// needs a mark-shaped string carrying a *currently pending* token, and
// where two sessions match it, the ambiguity rule already refuses to bind
// rather than choosing.
func TestEchoedMarkExampleIsAnAcceptedResidual(t *testing.T) {
	token := "mf-k174-echoed-example"
	binder := runMarkCommand(t, token, 404, "session-echoer",
		`echo "flow stage begin -stage do.tests -session-token `+token+` -harness claude-code"`)
	if binder.bindCalls != 1 {
		t.Fatalf("bindCalls = %d, want 1: an echoed mark-shaped example is an accepted residual false positive (design.md), not rejected", binder.bindCalls)
	}
	if binder.bound[404] != "session-echoer" {
		t.Fatalf("bound session = %q, want session-echoer", binder.bound[404])
	}
}

// fakeDispatchMetricsSink is DispatchMetricsSink's minimal in-memory
// stand-in for TestAmbiguousDispatchIsStamped below -- it only needs to
// prove whether a merge happened, not to reproduce jsonb_deep_add's own
// merge semantics the way fakeHarvestSink does for the stage grain.
type fakeDispatchMetricsSink struct {
	merged map[int64]int // dispatchID -> call count
}

func (s *fakeDispatchMetricsSink) MergeDispatchMetrics(_ context.Context, dispatchID int64, _ json.RawMessage) error {
	if s.merged == nil {
		s.merged = map[int64]int{}
	}
	s.merged[dispatchID]++
	return nil
}

var _ harvest.DispatchMetricsSink = (*fakeDispatchMetricsSink)(nil)

// TestGiveUpIsPersisted is task 5 steps 1 and 2 (tasks.md): both of
// resolveSessionTokens' give-up branches -- case 0, the bounded window
// exhausted, and the default branch, the token matched more than one
// session -- must persist the give-up through the binder before this
// Watcher's own in-memory gaveUpTokens set stops it searching, and each
// branch's reason must be its own, not a shared string reused for both
// (design.md's "say plainly where cost could not be attributed" needs the
// two failures told apart, not merged into one).
//
// Reverting task 6 leaves the give-up in memory only: neither subtest's
// RecordSessionTokenGiveUp assertion has anything to pass against, since
// nothing in watcher.go calls it yet.
func TestGiveUpIsPersisted(t *testing.T) {
	var exhaustedReason, ambiguousReason string

	t.Run("bounded window exhausted", func(t *testing.T) {
		dir := t.TempDir()
		const token = "mf-never-appears-persisted"
		binder := &countingSessionTokenBinder{sessionToken: token, stageRunID: 501}
		windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
		sink := newFakeHarvestSink()
		w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, nil)

		// 60 empty cycles: nothing to find, so the give-up bound
		// (maxSessionTokenResolutionCycles, watcher.go) is reached on the
		// last of these -- the same drive
		// TestSessionTokenStopsBeingScannedAfterBoundedGiveUp already
		// uses for the in-memory-only half of this behaviour.
		for i := range 60 {
			if _, err := w.RunOnce(context.Background()); err != nil {
				t.Fatalf("RunOnce (cycle %d): %v", i, err)
			}
		}

		if len(binder.giveUpCalls) != 1 {
			t.Fatalf("RecordSessionTokenGiveUp called %d times after the bounded window, want exactly 1: %v", len(binder.giveUpCalls), binder.giveUpCalls)
		}
		if got := binder.giveUpCalls[0].Token; got != token {
			t.Fatalf("persisted give-up token = %q, want %q", got, token)
		}
		exhaustedReason = binder.giveUpCalls[0].Reason
		if exhaustedReason == "" {
			t.Fatalf("persisted give-up reason is empty, want a reason naming the bounded window")
		}

		// One more cycle must not persist the give-up again: the
		// in-memory gaveUpTokens set is what stops this same process
		// re-searching within one run (task 6 step 2, tasks.md); the
		// store is only what survives the process, not what is
		// re-written every idle cycle after.
		if _, err := w.RunOnce(context.Background()); err != nil {
			t.Fatalf("RunOnce (extra cycle): %v", err)
		}
		if len(binder.giveUpCalls) != 1 {
			t.Fatalf("RecordSessionTokenGiveUp called %d times after an extra cycle, want still 1", len(binder.giveUpCalls))
		}
	})

	t.Run("ambiguous", func(t *testing.T) {
		dir := t.TempDir()
		const token = "mf-ambiguous-persisted"
		binder := &countingSessionTokenBinder{sessionToken: token, stageRunID: 502}

		writeMark := func(name, sessionID string) {
			line := fmt.Sprintf(`{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":%q,"message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -session-token `+token+`"}}]}}`+"\n", sessionID)
			if err := os.WriteFile(filepath.Join(dir, name), []byte(line), 0o644); err != nil {
				t.Fatalf("write %s: %v", name, err)
			}
		}
		writeMark("one.jsonl", "session-one")
		writeMark("two.jsonl", "session-two")

		windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
		sink := newFakeHarvestSink()
		w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, nil)

		if _, err := w.RunOnce(context.Background()); err != nil {
			t.Fatalf("RunOnce: %v", err)
		}

		if len(binder.giveUpCalls) != 1 {
			t.Fatalf("RecordSessionTokenGiveUp called %d times for the ambiguous token, want exactly 1: %v", len(binder.giveUpCalls), binder.giveUpCalls)
		}
		if got := binder.giveUpCalls[0].Token; got != token {
			t.Fatalf("persisted give-up token = %q, want %q", got, token)
		}
		ambiguousReason = binder.giveUpCalls[0].Reason
		if ambiguousReason == "" {
			t.Fatalf("persisted give-up reason is empty, want a reason naming the ambiguity")
		}
	})

	if exhaustedReason != "" && ambiguousReason != "" && exhaustedReason == ambiguousReason {
		t.Fatalf("both give-up reasons are %q -- the exhausted-window and ambiguous-match cases must persist distinct reasons, not a shared one", exhaustedReason)
	}
}

// TestGiveUpStampsItsOwnDispatches is task 6.1 (tasks.md, "Stamp a
// given-up session's own dispatches"): task 6 wired the give-up itself
// into RecordSessionTokenGiveUp and a dispatch-grain ambiguity into
// MarkDispatchesUnattributedByID, but left nothing stamping the
// dispatches of a session that gave up -- the "cost unattributed --
// session never bound" state task 8 renders had no producer at all.
// Both of resolveSessionTokens' give-up branches must call
// MarkDispatchesUnattributed (the token form) immediately after
// RecordSessionTokenGiveUp, carrying that branch's own reason -- a
// session that never bound leaves every one of its dispatches uncosted,
// which the token form expresses; the id form (MarkDispatchesUnattributedByID)
// names specific rows for a narrower, dispatch-grain ambiguity and is not
// a substitute for it.
func TestGiveUpStampsItsOwnDispatches(t *testing.T) {
	t.Run("bounded window exhausted", func(t *testing.T) {
		dir := t.TempDir()
		const token = "mf-never-appears-stamped"
		binder := &countingSessionTokenBinder{sessionToken: token, stageRunID: 801}
		windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
		sink := newFakeHarvestSink()
		w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, nil)

		for i := range 60 {
			if _, err := w.RunOnce(context.Background()); err != nil {
				t.Fatalf("RunOnce (cycle %d): %v", i, err)
			}
		}

		if len(binder.giveUpCalls) != 1 {
			t.Fatalf("RecordSessionTokenGiveUp called %d times, want exactly 1: %v", len(binder.giveUpCalls), binder.giveUpCalls)
		}
		if len(binder.unattributedTokenCalls) != 1 {
			t.Fatalf("MarkDispatchesUnattributed called %d times, want exactly 1: %v", len(binder.unattributedTokenCalls), binder.unattributedTokenCalls)
		}
		got := binder.unattributedTokenCalls[0]
		if got.token != token {
			t.Fatalf("stamped token = %q, want %q", got.token, token)
		}
		if got.reason != binder.giveUpCalls[0].Reason {
			t.Fatalf("stamped reason = %q, want %q (the same reason the give-up itself recorded)", got.reason, binder.giveUpCalls[0].Reason)
		}
	})

	t.Run("ambiguous", func(t *testing.T) {
		dir := t.TempDir()
		const token = "mf-ambiguous-stamped"
		binder := &countingSessionTokenBinder{sessionToken: token, stageRunID: 802}

		writeMark := func(name, sessionID string) {
			line := fmt.Sprintf(`{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":%q,"message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -session-token `+token+`"}}]}}`+"\n", sessionID)
			if err := os.WriteFile(filepath.Join(dir, name), []byte(line), 0o644); err != nil {
				t.Fatalf("write %s: %v", name, err)
			}
		}
		writeMark("one.jsonl", "session-one")
		writeMark("two.jsonl", "session-two")

		windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
		sink := newFakeHarvestSink()
		w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, nil)

		if _, err := w.RunOnce(context.Background()); err != nil {
			t.Fatalf("RunOnce: %v", err)
		}

		if len(binder.giveUpCalls) != 1 {
			t.Fatalf("RecordSessionTokenGiveUp called %d times, want exactly 1: %v", len(binder.giveUpCalls), binder.giveUpCalls)
		}
		if len(binder.unattributedTokenCalls) != 1 {
			t.Fatalf("MarkDispatchesUnattributed called %d times, want exactly 1: %v", len(binder.unattributedTokenCalls), binder.unattributedTokenCalls)
		}
		got := binder.unattributedTokenCalls[0]
		if got.token != token {
			t.Fatalf("stamped token = %q, want %q", got.token, token)
		}
		if got.reason != binder.giveUpCalls[0].Reason {
			t.Fatalf("stamped reason = %q, want %q (the same reason the give-up itself recorded)", got.reason, binder.giveUpCalls[0].Reason)
		}
		if got.candidates != 2 {
			t.Fatalf("stamped candidates = %d, want 2 (the two sessions this token matched)", got.candidates)
		}
	})

	t.Run("stamp failure never blocks the give-up", func(t *testing.T) {
		dir := t.TempDir()
		const token = "mf-stamp-fails"
		binder := &countingSessionTokenBinder{sessionToken: token, stageRunID: 803}
		binder.unattributedTokenErr = errors.New("stamp store outage")
		windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
		sink := newFakeHarvestSink()
		var logBuf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))
		w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, logger)

		for i := range 60 {
			if _, err := w.RunOnce(context.Background()); err != nil {
				t.Fatalf("RunOnce (cycle %d): %v", i, err)
			}
		}

		if len(binder.giveUpCalls) != 1 {
			t.Fatalf("RecordSessionTokenGiveUp called %d times despite the stamp erroring, want exactly 1: a stamp failure must never block the give-up itself from being recorded: %v", len(binder.giveUpCalls), binder.giveUpCalls)
		}
		if len(binder.unattributedTokenCalls) != 1 {
			t.Fatalf("MarkDispatchesUnattributed called %d times, want exactly 1 even though it errored", len(binder.unattributedTokenCalls))
		}
		if !strings.Contains(logBuf.String(), "stamp unattributed dispatches failed") || !strings.Contains(logBuf.String(), "stamp store outage") {
			t.Fatalf("log output = %q, want it to report both the stamp failure and the injected error: a stamp failure that is silently dropped is not \"never blocks\" -- the run continuing is only half of that guarantee, the failure being visible is the other half", logBuf.String())
		}

		// The give-up stands even though its stamp failed: one more cycle
		// must not re-record either call -- w.gaveUpTokens still stops
		// this process re-searching a token it has already given up on,
		// stamp error or not.
		if _, err := w.RunOnce(context.Background()); err != nil {
			t.Fatalf("RunOnce (extra cycle): %v", err)
		}
		if len(binder.giveUpCalls) != 1 {
			t.Fatalf("RecordSessionTokenGiveUp called %d times after an extra cycle, want still 1", len(binder.giveUpCalls))
		}
		if len(binder.unattributedTokenCalls) != 1 {
			t.Fatalf("MarkDispatchesUnattributed called %d times after an extra cycle, want still 1", len(binder.unattributedTokenCalls))
		}
	})
}

// TestPersistedGiveUpBindsFromAFullyConsumedTranscript is task 6.2
// (tasks.md, "bind a retried token by scanning, not by waiting for new
// bytes") -- the gap task 12's own live restart measured, which none of
// task 5's or task 6's own tests caught because their fake binder and
// fresh sink both start every transcript at offset 0: the real defect
// only shows once a transcript's harvest_offsets already sits at that
// file's own EOF -- kan-302's own measured state -- *before* the token
// that transcript's mark carries is retried. matchSessionTokens
// (watcher.go) only ever scans bytes newly read this cycle, so nothing
// in the ordinary path can find a mark sitting behind an offset that will
// never move again.
//
// Every subtest here pre-seeds the fake sink's own offset for a
// transcript to that file's exact size (os.Stat), reproducing "already
// fully consumed" directly rather than driving 60+ real RunOnce cycles
// to get there.
func TestPersistedGiveUpBindsFromAFullyConsumedTranscript(t *testing.T) {
	t.Run("binds", func(t *testing.T) {
		dir := t.TempDir()
		const token = "mf-consumed-retry"
		const stageRunID = int64(901)
		binder := &countingSessionTokenBinder{sessionToken: token, stageRunID: stageRunID}
		binder.seededGiveUps = []harvest.GiveUp{{Token: token, Reason: "session-never-bound", Retries: 1}}

		path := filepath.Join(dir, "consumed.jsonl")
		// The mark and this same line's own usage arrive together, the
		// ordinary shape Claude Code flushes a turn in (RunOnce's own doc
		// comment, "withholding a batch that revealed a sessionToken") --
		// carrying real usage here, alongside a real open window for the
		// session this mark identifies, is what makes the property-1
		// assertions below meaningful: a wrongly-attributing scan would
		// have somewhere to attribute these 1000+500 tokens to, not merely
		// nowhere to put them.
		line := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-consumed","message":{"model":"claude-opus-5","usage":{"input_tokens":1000,"output_tokens":500},"content":[{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -session-token ` + token + `"}}]}}` + "\n"
		if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}

		windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
			"session-consumed": {{
				StageRunID: stageRunID,
				SessionID:  "session-consumed",
				StartedAt:  time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			}},
		}}
		sink := newFakeHarvestSink()
		// Simulate a prior process having already fully harvested this
		// transcript, before this token's give-up was ever persisted:
		// harvest_offsets for path is already at EOF.
		sink.offsets[path] = info.Size()

		w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, nil)

		if _, err := w.RunOnce(context.Background()); err != nil {
			t.Fatalf("RunOnce: %v", err)
		}

		if binder.persistedGiveUpsCalls != 1 {
			t.Fatalf("PersistedGiveUps called %d times, want exactly 1", binder.persistedGiveUpsCalls)
		}
		if binder.bindCalls != 1 {
			t.Fatalf("bindCalls = %d, want 1: a persisted give-up whose mark sits behind the harvest offset must still bind", binder.bindCalls)
		}
		if binder.bound[stageRunID] != "session-consumed" {
			t.Fatalf("bound session for stage run %d = %q, want session-consumed", stageRunID, binder.bound[stageRunID])
		}

		// Property 1 ("never re-attribute usage"): the scan must produce
		// no usage records and must not move harvest_offsets. commitCount
		// staying 0 proves CommitHarvestBatch was never called at all --
		// stronger than checking totals, which could stay zero merely
		// because nothing happened to be wired up to receive them.
		if sink.commitCount != 0 {
			t.Fatalf("commitCount = %d, want 0: binding a retried token must never attribute or commit this transcript's usage a second time", sink.commitCount)
		}
		gotOffset, found, err := sink.GetHarvestOffset(context.Background(), path)
		if err != nil {
			t.Fatalf("GetHarvestOffset: %v", err)
		}
		if !found || gotOffset != info.Size() {
			t.Fatalf("harvest offset for %s = (%d, %v), want (%d, true) unchanged", path, gotOffset, found, info.Size())
		}
		if _, ok := sink.totals[stageRunID]; ok {
			t.Fatalf("stage run %d has a committed token total, want none: this transcript's usage must never be attributed by this scan", stageRunID)
		}
	})

	t.Run("ambiguous still refuses", func(t *testing.T) {
		dir := t.TempDir()
		const token = "mf-consumed-ambiguous"
		binder := &countingSessionTokenBinder{sessionToken: token, stageRunID: 902}
		binder.seededGiveUps = []harvest.GiveUp{{Token: token, Reason: "session-never-bound", Retries: 1}}

		writeConsumed := func(name, sessionID string) string {
			line := fmt.Sprintf(`{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":%q,"message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -session-token `+token+`"}}]}}`+"\n", sessionID)
			p := filepath.Join(dir, name)
			if err := os.WriteFile(p, []byte(line), 0o644); err != nil {
				t.Fatalf("write %s: %v", name, err)
			}
			return p
		}
		one := writeConsumed("one.jsonl", "session-one")
		two := writeConsumed("two.jsonl", "session-two")

		windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
		sink := newFakeHarvestSink()
		for _, p := range []string{one, two} {
			info, err := os.Stat(p)
			if err != nil {
				t.Fatalf("stat %s: %v", p, err)
			}
			sink.offsets[p] = info.Size() // both already fully consumed
		}

		var logBuf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))
		w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, logger)

		if _, err := w.RunOnce(context.Background()); err != nil {
			t.Fatalf("RunOnce: %v", err)
		}

		if binder.bindCalls != 0 {
			t.Fatalf("bindCalls = %d, want 0: a retried token whose marks sit in two already-consumed transcripts must still refuse to bind", binder.bindCalls)
		}
		if !strings.Contains(logBuf.String(), "more than one session") {
			t.Fatalf("log output = %q, want a message reporting the ambiguity", logBuf.String())
		}
	})

	t.Run("mention only never binds even behind the offset", func(t *testing.T) {
		dir := t.TempDir()
		const token = "mf-consumed-mention-only"
		binder := &countingSessionTokenBinder{sessionToken: token, stageRunID: 903}
		binder.seededGiveUps = []harvest.GiveUp{{Token: token, Reason: "session-never-bound", Retries: 1}}

		path := filepath.Join(dir, "mentioned.jsonl")
		line := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-mentioning","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"grep '` + token + `' /tmp/dispatch.log"}}]}}` + "\n"
		if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}

		windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
		sink := newFakeHarvestSink()
		sink.offsets[path] = info.Size() // already consumed, like the real mark's own transcript would be

		w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, nil)
		if _, err := w.RunOnce(context.Background()); err != nil {
			t.Fatalf("RunOnce: %v", err)
		}
		if binder.bindCalls != 0 {
			t.Fatalf("bindCalls = %d, want 0: a command that only mentions the token, even behind the offset, must never bind it (KAN-172 F4)", binder.bindCalls)
		}
	})
}

// TestPersistedGiveUpIsRetriedOnStart is task 5 step 3 (tasks.md), the
// kan-302 recovery case: a token a prior process gave up on, whose
// binder now reports it through PersistedGiveUps, must be searched for
// again once a fresh Watcher's first cycle reads that persisted state --
// and, its transcript now carrying the mark, must bind.
//
// The PersistedGiveUps-was-called assertion is what makes this test
// genuinely exercise the recovery path rather than pass by coincidence:
// a fresh Watcher's own in-memory gaveUpTokens starts empty regardless of
// any persisted state, so "it binds" alone would not distinguish
// deliberate recovery from an accident of construction. Reverting task 6
// fails the call-count assertion first, because nothing in watcher.go
// calls PersistedGiveUps at all.
func TestPersistedGiveUpIsRetriedOnStart(t *testing.T) {
	dir := t.TempDir()
	const token = "mf-recovered"
	binder := &countingSessionTokenBinder{sessionToken: token, stageRunID: 601}
	binder.seededGiveUps = []harvest.GiveUp{{Token: token, Reason: "session-never-bound", Retries: 1}}

	path := filepath.Join(dir, "recovered.jsonl")
	line := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-recovered","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"flow stage begin -session-token ` + token + `"}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, nil)

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if binder.persistedGiveUpsCalls != 1 {
		t.Fatalf("PersistedGiveUps called %d times, want exactly 1: a fresh Watcher must read the persisted give-ups at start so this token re-enters the pending set instead of relying on an empty in-memory map by accident", binder.persistedGiveUpsCalls)
	}
	if binder.bindCalls != 1 {
		t.Fatalf("bindCalls = %d, want 1: the persisted give-up's token must be searched for again now that its mark exists", binder.bindCalls)
	}
	if binder.bound[601] != "session-recovered" {
		t.Fatalf("bound session for stage run 601 = %q, want session-recovered", binder.bound[601])
	}
}

// TestRetryStillBounded is task 5 step 4 (tasks.md): a token retried
// after a restart, whose transcript still carries no mark, must give up
// again after the same bounded window as the first attempt -- not loop
// forever now that it has already been retried once -- and that second
// give-up's retries figure must rise from whatever the persisted state
// already recorded.
func TestRetryStillBounded(t *testing.T) {
	dir := t.TempDir()
	const token = "mf-retried-and-still-missing"
	binder := &countingSessionTokenBinder{sessionToken: token, stageRunID: 701}
	binder.seededGiveUps = []harvest.GiveUp{{Token: token, Reason: "session-never-bound", Retries: 1}}
	// giveUpState is seeded to match seededGiveUps: this token has
	// already been recorded twice before this test's own retry (retries:
	// 1 -- store.Store.RecordSessionTokenGiveUp's own doc comment,
	// giveups.go, "retries starts at 0 on the first recording"), so the
	// retry this test drives is its third recording and must land at
	// retries: 2.
	binder.giveUpState = map[string]*harvest.GiveUp{
		token: {Token: token, Reason: "session-never-bound", Retries: 1},
	}

	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), sessionBinderDeps{binder: binder}, nil)

	// A fresh bounded window, exactly as long as the first attempt's own
	// (task 6 step 3, tasks.md, "w.tokenCycles starts fresh") --
	// maxSessionTokenResolutionCycles, 60 (watcher.go). Checkpointing one
	// cycle short of the bound, before driving the final cycle, pins the
	// window's actual *length* rather than only its upper bound: an
	// implementation that seeded a retried token's own cycle counter from
	// somewhere other than zero -- from the persisted retries figure, say
	// -- could give up well before cycle 60 and still pass a check that
	// only asserted "has given up by cycle 60".
	for i := range 59 {
		if _, err := w.RunOnce(context.Background()); err != nil {
			t.Fatalf("RunOnce (cycle %d): %v", i, err)
		}
	}
	if binder.persistedGiveUpsCalls != 1 {
		t.Fatalf("PersistedGiveUps called %d times, want exactly 1 (at start): a token retried across a restart must still be searched for on one fresh bounded window, not re-read on every cycle", binder.persistedGiveUpsCalls)
	}
	if len(binder.giveUpCalls) != 0 {
		t.Fatalf("RecordSessionTokenGiveUp called %d times after 59 of the retry's 60 cycles, want 0: the retry's bounded window must be exactly as long as the first attempt's, not shorter", len(binder.giveUpCalls))
	}

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (cycle 60): %v", err)
	}
	if len(binder.giveUpCalls) != 1 {
		t.Fatalf("RecordSessionTokenGiveUp called %d times after the retry's own bounded window, want exactly 1: %v", len(binder.giveUpCalls), binder.giveUpCalls)
	}
	if got := binder.giveUpCalls[0].Retries; got != 2 {
		t.Fatalf("retry's give-up retries = %d, want 2: this is the token's third recording, and retries rises by one on every recording after the first (giveups.go)", got)
	}

	// The retry is itself bounded exactly as the first attempt was: one
	// more cycle beyond this second window must not persist a third
	// give-up.
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce (extra cycle): %v", err)
	}
	if len(binder.giveUpCalls) != 1 {
		t.Fatalf("RecordSessionTokenGiveUp called %d times after an extra cycle, want still 1", len(binder.giveUpCalls))
	}
}

// TestAmbiguousDispatchIsStamped is task 5 step 5 (tasks.md): where the
// dispatch-grain second pass's bestDispatchWindow (attribute.go) refuses
// to attribute a record because it matched more than one dispatch by
// agent id -- the identity pass's own ambiguous case,
// TestDispatchDuplicateAgentIDAttributesToNeither's scenario in
// attribute_test.go, driven here through a real Watcher rather than
// Attribute directly -- the Watcher must stamp every candidate as
// unattributed with the ambiguity's reason and candidate count, and must
// never merge dispatch metrics for a record it refused to attribute.
func TestAmbiguousDispatchIsStamped(t *testing.T) {
	dir := t.TempDir()
	binder := &countingSessionTokenBinder{}

	dispatchWindows := &fakeDispatchWindowSource{bySession: map[string][]harvest.DispatchWindow{
		mainSessionID: {
			panelWindow(t, 1, "agent-one", "2026-01-01T00:10:00Z"),
			panelWindow(t, 2, "agent-one", "2026-01-01T00:10:00.5Z"),
		},
	}}
	dispatchSink := &fakeDispatchMetricsSink{}

	line := fmt.Sprintf(`{"type":"assistant","timestamp":"2026-01-01T00:10:30Z","sessionId":%q,"isSidechain":true,"agentId":"agent-one","message":{"model":"claude-opus-5","usage":{"input_tokens":5,"output_tokens":1}}}`+"\n", mainSessionID)
	if err := os.WriteFile(filepath.Join(dir, "panel.jsonl"), []byte(line), 0o644); err != nil {
		t.Fatalf("write panel.jsonl: %v", err)
	}

	stageWindows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(stageWindows), sessionBinderAndDispatchDeps{
		sessionBinderDeps: sessionBinderDeps{binder: binder},
		windows:           dispatchWindows,
		sink:              dispatchSink,
	}, nil)

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if len(dispatchSink.merged) != 0 {
		t.Fatalf("dispatch metrics merged = %v, want none: two dispatches recording the same agent id is ambiguous, and an ambiguous record must not attribute to either", dispatchSink.merged)
	}
	if len(binder.unattributedCalls) != 1 {
		t.Fatalf("MarkDispatchesUnattributedByID called %d times, want exactly 1: %v", len(binder.unattributedCalls), binder.unattributedCalls)
	}
	got := binder.unattributedCalls[0]
	if got.reason == "" {
		t.Fatalf("unattributed reason is empty, want a reason naming the ambiguity")
	}
	if got.candidates != 2 {
		t.Fatalf("unattributed candidates = %d, want 2 (the two dispatches the record's agent id matched)", got.candidates)
	}
	if !reflect.DeepEqual(got.ids, []int64{1, 2}) {
		t.Fatalf("unattributed ids = %v, want [1 2] -- the ambiguity names specific dispatch rows, not the session token", got.ids)
	}
}

// TestDispatchThatAttributedIsNeverStampedUnattributed is the regression
// test for review finding F2 (this change's own review panel):
// TestAmbiguousDispatchIsStamped above only ever drives a batch whose
// ambiguous candidates receive no tokens at all, so it never caught a
// dispatch that is *both* merged with real tokens *and* named in the same
// batch's ambiguous stamp -- contradicting task 6 step 4's own words, "Do
// not stamp a dispatch that attributed."
//
// The two panel windows below share agent id "agent-one", exactly as
// TestAmbiguousDispatchIsStamped's do, so a record carrying that id is
// ambiguous between dispatch 1 and dispatch 2 regardless of its own
// timestamp (bestDispatchWindow's identity pass ignores the interval
// entirely, attribute.go). This batch adds a second, sidechain record
// besides that one: no agent id at all, timestamped at 00:10:00.2 --
// inside dispatch 1's own interval but before dispatch 2's opens at
// 00:10:00.5 -- so the interval pass (bestDispatchWindow's fallback)
// attributes it cleanly to dispatch 1 alone, in the very same batch that
// also reports [1, 2] as ambiguous. Dispatch 1 must receive its merged
// tokens and must never appear in the unattributed stamp; dispatch 2,
// which this batch attributed nothing to, is the only one left to stamp
// -- still carrying the ambiguity's own full candidate count (2), not the
// filtered id count (1).
func TestDispatchThatAttributedIsNeverStampedUnattributed(t *testing.T) {
	dir := t.TempDir()
	binder := &countingSessionTokenBinder{}

	dispatchWindows := &fakeDispatchWindowSource{bySession: map[string][]harvest.DispatchWindow{
		mainSessionID: {
			panelWindow(t, 1, "agent-one", "2026-01-01T00:10:00Z"),
			panelWindow(t, 2, "agent-one", "2026-01-01T00:10:00.5Z"),
		},
	}}
	dispatchSink := &fakeDispatchMetricsSink{}

	ambiguousLine := fmt.Sprintf(`{"type":"assistant","timestamp":"2026-01-01T00:10:30Z","sessionId":%q,"isSidechain":true,"agentId":"agent-one","message":{"model":"claude-opus-5","usage":{"input_tokens":5,"output_tokens":1}}}`+"\n", mainSessionID)
	cleanLine := fmt.Sprintf(`{"type":"assistant","timestamp":"2026-01-01T00:10:00.2Z","sessionId":%q,"isSidechain":true,"message":{"model":"claude-opus-5","usage":{"input_tokens":99,"output_tokens":7}}}`+"\n", mainSessionID)
	if err := os.WriteFile(filepath.Join(dir, "panel.jsonl"), []byte(ambiguousLine+cleanLine), 0o644); err != nil {
		t.Fatalf("write panel.jsonl: %v", err)
	}

	stageWindows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(stageWindows), sessionBinderAndDispatchDeps{
		sessionBinderDeps: sessionBinderDeps{binder: binder},
		windows:           dispatchWindows,
		sink:              dispatchSink,
	}, nil)

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if dispatchSink.merged[1] != 1 {
		t.Fatalf("dispatch 1 merged = %d, want 1: the clean record's tokens must still reach the sink", dispatchSink.merged[1])
	}

	for _, call := range binder.unattributedCalls {
		for _, id := range call.ids {
			if id == 1 {
				t.Fatalf("dispatch 1 stamped unattributed (%+v) despite receiving real tokens in the same batch -- task 6 step 4: %q", call, "Do not stamp a dispatch that attributed")
			}
		}
	}

	if len(binder.unattributedCalls) != 1 {
		t.Fatalf("MarkDispatchesUnattributedByID called %d times, want exactly 1 (for dispatch 2 alone): %v", len(binder.unattributedCalls), binder.unattributedCalls)
	}
	got2 := binder.unattributedCalls[0]
	if !reflect.DeepEqual(got2.ids, []int64{2}) {
		t.Fatalf("unattributed ids = %v, want [2] -- dispatch 1 attributed in this batch and must be filtered out of the stamp", got2.ids)
	}
	if got2.candidates != 2 {
		t.Fatalf("unattributed candidates = %d, want 2 -- the true size of the ambiguity, not the filtered id count", got2.candidates)
	}
}
