package sweep_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
	"github.com/tweety53/agents/stats/internal/sweep"
)

func baseBeginInput(projectKey, name, command, stage string, startedAt time.Time) store.BeginStageInput {
	return store.BeginStageInput{
		ProjectKey: projectKey,
		ChangeName: name,
		Harness:    "claude-code",
		SessionID:  ptr("session-1"),
		Command:    command,
		Stage:      stage,
		StartedAt:  startedAt,
	}
}

// fixedClock returns a Sweeper's "now" pinned at t, so RunOnce's cutoff is
// deterministic instead of racing the wall clock -- the same reason
// store's own SweepAbandoned test computes its cutoff once, up front,
// rather than re-deriving "now" inside the assertion.
func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

// newSweeperAt builds a Sweeper over st whose RunOnce always computes its
// cutoff from at, not the real wall clock.
func newSweeperAt(st *store.Store, timeout time.Duration, at time.Time) *sweep.Sweeper {
	s := sweep.New(st, timeout, nil)
	sweep.SetClockForTest(s, fixedClock(at))
	return s
}

// TestSweepClosesSilentStages asserts that a stage run started well before
// the sweeper's silence timeout, with no end mark, is closed with outcome
// "abandoned" once RunOnce runs.
//
// Mutation check: deleting the "started_at < $1" half of SweepAbandoned's
// WHERE clause (store/stageruns.go) would still make this test pass --
// every open row would be swept regardless of age -- which is exactly why
// TestSweepLeavesActiveStagesOpen exists alongside it: only together do
// the two pin both the "old and open" and "young and open" halves of the
// predicate.
func TestSweepClosesSilentStages(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-sweep-close-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	now := time.Date(2026, 8, 13, 15, 0, 0, 0, time.UTC)
	started := now.Add(-2 * time.Hour)
	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task", started))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	s := newSweeperAt(st, time.Hour, now)
	n, err := s.RunOnce(ctx)
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if n != 1 {
		t.Fatalf("RunOnce closed %d rows, want 1", n)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	if got.Outcome == nil || *got.Outcome != "abandoned" {
		t.Errorf("Outcome = %v, want \"abandoned\"", got.Outcome)
	}
	if got.EndedAt == nil {
		t.Error("EndedAt is nil, want set by the sweep")
	}
}

// TestSweepLeavesActiveStagesOpen asserts that a stage run started more
// recently than the sweeper's silence timeout is left open: a session
// still legitimately working must never be second-guessed as abandoned.
//
// Mutation check: inverting SweepAbandoned's comparison
// (started_at < $1 to started_at > $1, or dropping the "AND started_at"
// clause entirely) makes this test fail, since the run seeded here would
// then be swept -- confirmed by running the mutation locally.
func TestSweepLeavesActiveStagesOpen(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-sweep-active-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	now := time.Date(2026, 8, 13, 15, 0, 0, 0, time.UTC)
	started := now.Add(-5 * time.Minute)
	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task", started))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	s := newSweeperAt(st, time.Hour, now)
	n, err := s.RunOnce(ctx)
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if n != 0 {
		t.Fatalf("RunOnce closed %d rows, want 0 (the run is still within its silence timeout)", n)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	if got.EndedAt != nil || got.Outcome != nil {
		t.Errorf("stage run was closed (EndedAt=%v, Outcome=%v), want still open", got.EndedAt, got.Outcome)
	}
}

// TestSweepIsIdempotent asserts that running RunOnce twice against the
// same silent stage run closes it once, and the second call is a genuine
// no-op: it neither re-closes the row (bumping its EndedAt to a later
// instant) nor reports a phantom second closure.
//
// Mutation check: replacing SweepAbandoned's WHERE clause with one that
// does not require ended_at IS NULL (for example matching on started_at
// alone) makes this test fail: the second RunOnce would report n=1 again
// and stamp a new EndedAt over the first, which the EndedAt-unchanged
// assertion below catches.
func TestSweepIsIdempotent(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-sweep-idempotent-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	now := time.Date(2026, 8, 13, 15, 0, 0, 0, time.UTC)
	started := now.Add(-2 * time.Hour)
	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task", started))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	s := newSweeperAt(st, time.Hour, now)
	first, err := s.RunOnce(ctx)
	if err != nil {
		t.Fatalf("first RunOnce: %v", err)
	}
	if first != 1 {
		t.Fatalf("first RunOnce closed %d rows, want 1", first)
	}
	afterFirst, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun after first sweep: %v", err)
	}

	// A later "now", well past the first sweep, so a bug that re-matches
	// already-closed rows would also stamp a visibly later EndedAt --
	// making the mutation this guards against unmissable rather than
	// merely a coin-flip on two calls sharing one instant.
	second, err := newSweeperAt(st, time.Hour, now.Add(10*time.Minute)).RunOnce(ctx)
	if err != nil {
		t.Fatalf("second RunOnce: %v", err)
	}
	if second != 0 {
		t.Fatalf("second RunOnce closed %d rows, want 0 (already closed)", second)
	}

	afterSecond, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun after second sweep: %v", err)
	}
	if !afterSecond.EndedAt.Equal(*afterFirst.EndedAt) {
		t.Errorf("EndedAt changed on the second sweep: %v -> %v", afterFirst.EndedAt, afterSecond.EndedAt)
	}
	if *afterSecond.Outcome != "abandoned" {
		t.Errorf("Outcome = %v, want \"abandoned\" (unchanged)", afterSecond.Outcome)
	}
}

// TestConcurrentSweepersDoNotDoubleCloseOrCollide covers the concurrency
// half of "the sweeper must not fight the harvester": two Sweeper
// instances (modelling two flowd processes, or one racing itself)
// sweeping the same set of silent stage runs at once must, in total,
// close each row exactly once -- never report the same row closed twice,
// and never error.
//
// This exercises SweepAbandoned's reliance on Postgres's own per-row
// locking and read-committed re-check (no client-side check-then-act).
// This store has had real concurrency defects of that check-then-act
// shape before -- found in review, not documented in design.md, which
// says nothing about them -- which is the reason this task's own
// instructions call for a real test here rather than an argument that the
// SQL "should" be safe.
func TestConcurrentSweepersDoNotDoubleCloseOrCollide(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-sweep-concurrent-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	const rows = 20
	now := time.Date(2026, 8, 13, 15, 0, 0, 0, time.UTC)
	started := now.Add(-2 * time.Hour)
	for i := range rows {
		_, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", fmt.Sprintf("stage-%d", i), started))
		if err != nil {
			t.Fatalf("BeginStage %d: %v", i, err)
		}
	}

	const sweepers = 5
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		total int64
		errs  []error
	)
	for range sweepers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n, err := newSweeperAt(st, time.Hour, now).RunOnce(ctx)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				return
			}
			total += n
		}()
	}
	wg.Wait()

	for _, err := range errs {
		t.Errorf("concurrent RunOnce error: %v", err)
	}
	if total != rows {
		t.Errorf("concurrent sweepers closed %d rows in total, want exactly %d (no double-count, none missed)", total, rows)
	}
}

// TestSweepRacingLiveEndDoesNotCorruptOutcome asserts the second
// concurrency requirement this task names explicitly: a sweep racing a
// live `stage end` for the same row must never leave the row in an
// inconsistent state (an outcome with no end instant, or vice versa), and
// must never lose the fact that the row was closed at all -- whichever of
// the two writes commits first determines the final outcome, and that is
// the only property asserted here, deliberately: this project's design
// (internal/api/stages.go's ApplyEndStageMark doc comment) already commits
// to "no correctness hazard, at most one spurious statistic", not to a
// specific winner between a live end and a sweep that happen to race.
//
// Which of the two calls actually wins is no longer "whoever commits
// last, silently" (fix round 4, design.md's end-mark-cannot-resurrect):
// EndStage's UPDATE now requires ended_at IS NULL, guarding against
// exactly the class of race this test drives, not only the supersede race
// the decision names first. If the sweep commits first, the live
// EndStage's now-guarded UPDATE finds the row already closed and returns
// store.ErrStageRunAlreadyClosed instead of silently overwriting
// "abandoned" with "completed" -- a typed, coherent refusal, not a torn
// state. If the live end commits first, SweepAbandoned's own UPDATE
// (unaffected by this task -- it was already a single, self-contained
// statement filtered on ended_at IS NULL, never a separate lookup-then-write
// pair) simply excludes the now-closed row from what it touches. Either
// way the row ends up with exactly one coherent outcome; this test checks
// that outcome is the one whichever call actually reported succeeding.
func TestSweepRacingLiveEndDoesNotCorruptOutcome(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-sweep-race-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	now := time.Date(2026, 8, 13, 15, 0, 0, 0, time.UTC)
	started := now.Add(-2 * time.Hour)
	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task", started))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	s := newSweeperAt(st, time.Hour, now)

	var wg sync.WaitGroup
	wg.Add(2)
	var sweepErr, endErr error
	go func() {
		defer wg.Done()
		_, sweepErr = s.RunOnce(ctx)
	}()
	go func() {
		defer wg.Done()
		endErr = st.EndStage(ctx, run.ID, now.Add(-time.Minute), "completed")
	}()
	wg.Wait()

	if sweepErr != nil {
		t.Fatalf("sweep RunOnce: %v", sweepErr)
	}
	// endErr is nil when the live end won the race, and
	// store.ErrStageRunAlreadyClosed when the sweep won it and closed the
	// row first -- both are the coherent, expected outcomes of this race;
	// anything else is a real failure.
	if endErr != nil && !errors.Is(endErr, store.ErrStageRunAlreadyClosed) {
		t.Fatalf("live EndStage: %v", endErr)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	if got.EndedAt == nil || got.Outcome == nil {
		t.Fatalf("stage run left inconsistent: EndedAt=%v, Outcome=%v -- a race must produce one coherent outcome, never a torn one", got.EndedAt, got.Outcome)
	}
	wantOutcome := "completed"
	if endErr != nil {
		wantOutcome = "abandoned"
	}
	if *got.Outcome != wantOutcome {
		t.Fatalf("Outcome = %q, want %q (whichever of sweep/live-end actually reported winning this race)", *got.Outcome, wantOutcome)
	}
}
