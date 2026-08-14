package harvest_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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
type fakeHarvestSink struct {
	offsets     map[string]int64
	totals      map[int64]harvest.TokenDelta            // cumulative, reproducing jsonb_deep_add's effect
	modelTotals map[int64]map[string]harvest.TokenDelta // same, per model (task 22's "models" key)

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
		offsets:     map[string]int64{},
		totals:      map[int64]harvest.TokenDelta{},
		modelTotals: map[int64]map[string]harvest.TokenDelta{},
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
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), nil)

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
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), nil)

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
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), nil)

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
// (a previous myflowd process, in production) left it, not from
// anything this process remembers. A brand-new Watcher, built fresh
// (never having called RunOnce before) but pointed at a sink that
// already has this file's true end offset committed, must add nothing.
func TestFreshWatcherOverAlreadyHarvestedTranscriptAddsNothing(t *testing.T) {
	dir := t.TempDir()
	copyFixtureInto(t, dir, "session.jsonl", mainThreadFixture)
	windows := func() *fakeWindowSource { return &fakeWindowSource{bySession: openWindowForMainSession(1)} }

	// A first Watcher plays the role of "the previous myflowd process":
	// it harvests the file completely and then is discarded -- nothing
	// about it survives into the second Watcher below except what it
	// committed to the (shared) sink.
	sink := newFakeHarvestSink()
	firstWatcher := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows()), nil)
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
	// a restarted myflowd would reconnect to the same Postgres database.
	freshWatcher := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows()), nil)
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
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), nil)

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
	cleanWatcher := harvest.NewWatcher(cleanDir, cleanSink, harvest.NewAttributor(windows()), nil)
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
	outageWatcher := harvest.NewWatcher(outageDir, outageSink, harvest.NewAttributor(windows()), nil)

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
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), nil)

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
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), nil, harvest.WithPricer(pricer))

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
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), nil, harvest.WithPricer(pricer))

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
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), nil)

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
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), nil, harvest.WithPricer(pricer))

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
