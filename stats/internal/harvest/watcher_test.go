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

// fakeSessionTokenStore implements both harvest.SessionTokenBinder and
// harvest.WindowSource over the same in-memory runs map.
type fakeSessionTokenStore struct {
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
		line := fmt.Sprintf(`{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":%q,"message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"myflow stage begin -stage do.tests -session-token %s -harness claude-code"}}]}}`+"\n", sessionID, sessionToken)
		if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	writeMark(sessionAlphaPath, "session-alpha", "mf-session-token-alpha")
	writeMark(sessionBetaPath, "session-beta", "mf-session-token-beta")

	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(sessionStore), nil, harvest.WithSessionTokenBinder(sessionStore))

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
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), nil, harvest.WithSessionTokenBinder(binder))

	for i := range 3 {
		if _, err := w.RunOnce(context.Background()); err != nil {
			t.Fatalf("RunOnce (cycle %d): %v", i, err)
		}
	}
	if binder.bindCalls != 0 {
		t.Fatalf("bindCalls = %d before the sessionToken ever appears, want 0", binder.bindCalls)
	}

	path := filepath.Join(dir, "late.jsonl")
	line := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-late","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"myflow stage begin -session-token mf-later-cycle"}}]}}` + "\n"
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
		line := fmt.Sprintf(`{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":%q,"message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"myflow stage begin -session-token mf-ambiguous"}}]}}`+"\n", sessionID)
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
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), logger, harvest.WithSessionTokenBinder(binder))

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
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), logger, harvest.WithSessionTokenBinder(binder))

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
	line := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-toolate","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"myflow stage begin -session-token mf-never-appears"}}]}}` + "\n"
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
	line := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-imposter","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"myflow stage begin -session-token mf-already-bound"}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), nil, harvest.WithSessionTokenBinder(binder))

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
	content := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-same-batch","message":{"model":"claude-opus-5","usage":{"input_tokens":5,"output_tokens":1},"content":[{"type":"tool_use","name":"Bash","input":{"command":"myflow stage begin -session-token mf-same-batch"}}]}}` + "\n" +
		`{"type":"assistant","timestamp":"2025-12-01T00:00:02Z","sessionId":"session-same-batch","message":{"model":"claude-opus-5","usage":{"input_tokens":333,"output_tokens":1}}}` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(sessionStore), nil, harvest.WithSessionTokenBinder(sessionStore))

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
// layer this test's fake stands in for: insertStageRun resolves
// session_id at insert time from an already-bound token, so a stage run
// created with a token that has already resolved never enters
// UnresolvedSessionTokens at all, and RunOnce never has a reason to
// withhold its batch.
func TestSecondMarkOfAnAlreadyBoundTokenCommitsInTheSameCycle(t *testing.T) {
	dir := t.TempDir()
	started := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	boundSessionID := "session-shared"

	// Stage run 2's session_id is already resolved -- reproducing what
	// store.Store.insertStageRun now does at insert time for a run whose
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
	content := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-shared","message":{"model":"claude-opus-5","usage":{"input_tokens":77,"output_tokens":1},"content":[{"type":"tool_use","name":"Bash","input":{"command":"myflow stage begin -session-token mf-shared"}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(sessionStore), nil, harvest.WithSessionTokenBinder(sessionStore))

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
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), nil, harvest.WithSessionTokenBinder(binder))

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
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), nil, harvest.WithSessionTokenBinder(binder))

	// Cycle 1: session A's real mark is written and read while the token
	// is NOT YET pending (matchSessionTokens returns early on an empty
	// pending map), so the offset commits past it exactly as it did on
	// the live daemon before its SessionTokenBinder was wired in.
	sessionAPath := filepath.Join(dir, "session-a.jsonl")
	markLine := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-a","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"myflow stage begin -stage do.tests -session-token ` + token + ` -harness claude-code"}}]}}` + "\n"
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

// TestMarkInvocationShapeIsNotOverAnchored proves the matcher accepts the
// invocation shapes a real mark can actually take -- a leading `cd ... &&`,
// flags in a different order than the doc comment's example, and the
// `-session-token=value` form -- rather than a rewrite that only accepts
// one rigid shape.
func TestMarkInvocationShapeIsNotOverAnchored(t *testing.T) {
	dir := t.TempDir()
	token := "mf-k172-shape"
	binder := &togglableSessionTokenBinder{sessionToken: token, stageRunID: 303, pending: true}
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), nil, harvest.WithSessionTokenBinder(binder))

	path := filepath.Join(dir, "session.jsonl")
	// -harness comes before -session-token (order differs from the doc
	// comment's canonical example), and the whole thing is prefixed by a
	// cd && compound, exactly as a real caller might invoke it.
	line := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-shape","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"cd /repo && myflow stage begin -harness claude-code -stage do.tests -session-token ` + token + ` change-name"}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if binder.bindCalls != 1 {
		t.Fatalf("bindCalls = %d, want 1: a genuine mark wrapped in cd && and with reordered flags must still bind", binder.bindCalls)
	}
	if binder.bound[303] != "session-shape" {
		t.Fatalf("bound session = %q, want session-shape", binder.bound[303])
	}
}

// TestEchoedMarkExampleDoesNotBind is F5's regression test: a command that
// only PRINTS a mark-shaped string -- echo "myflow stage begin ...
// -session-token <token> ..." -- must never bind, because it is not itself
// an invocation of myflow. Before the F5 fix, strings.Fields saw
// -session-token and the token as adjacent fields regardless of the
// surrounding quotes, and stageMarkInvocationPattern's "stage begin" match
// was found inside the quoted text too, so an echoed example counted
// exactly like a real mark and silently bound whichever session printed
// it.
func TestEchoedMarkExampleDoesNotBind(t *testing.T) {
	dir := t.TempDir()
	token := "mf-k172-echoed-example"
	binder := &togglableSessionTokenBinder{sessionToken: token, stageRunID: 404, pending: true}
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{}}
	sink := newFakeHarvestSink()
	w := harvest.NewWatcher(dir, sink, harvest.NewAttributor(windows), nil, harvest.WithSessionTokenBinder(binder))

	path := filepath.Join(dir, "echoer.jsonl")
	line := `{"type":"assistant","timestamp":"2025-12-01T00:00:01Z","sessionId":"session-echoer","message":{"model":"claude-opus-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"echo \"myflow stage begin -stage do.tests -session-token ` + token + ` -harness claude-code\""}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if binder.bindCalls != 0 {
		t.Fatalf("bindCalls = %d, want 0: a command that only echoes a mark-shaped example must never bind (F5)", binder.bindCalls)
	}
}
