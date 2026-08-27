package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/harvest"
	"github.com/tweety53/agents/stats/internal/store"
)

// *store.Store must satisfy harvest.SessionTokenBinder structurally -- this is
// the compile-time half of the same wiring assertion cmd/flowd/main.go
// makes for WindowSource, HarvestSink and Pricer (task 2's own file list
// keeps this task out of cmd/flowd, so it is asserted here instead,
// where UnresolvedSessionTokens and BindSession are actually defined).
var _ harvest.SessionTokenBinder = (*store.Store)(nil)

// seedChange puts a minimal, valid change under projectKey/name and returns
// nothing -- callers key every stage run call by that same project/name
// pair, exactly as BeginStage requires.
func seedChange(t *testing.T, st *store.Store, projectKey, name string) {
	t.Helper()
	ctx := context.Background()
	c := baseChange(projectKey, name)
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("seed change %s/%s: %v", projectKey, name, err)
	}
}

func baseBeginInput(projectKey, name, command, stage string) store.BeginStageInput {
	return store.BeginStageInput{
		ProjectKey: projectKey,
		ChangeName: name,
		Harness:    "claude-code",
		SessionID:  ptr("session-1"),
		Command:    command,
		Stage:      stage,
		StartedAt:  time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC),
	}
}

func TestBeginStageAllocatesAttempts(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-attempts-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")

	first, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("first BeginStage: %v", err)
	}
	if first.Attempt != 1 {
		t.Errorf("first BeginStage attempt = %d, want 1", first.Attempt)
	}

	second, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("second BeginStage: %v", err)
	}
	if second.Attempt != 2 {
		t.Errorf("second BeginStage attempt = %d, want 2", second.Attempt)
	}
	if second.ID == first.ID {
		t.Errorf("second BeginStage returned the same row id as the first: %d", second.ID)
	}

	// A different stage of the same command starts its own attempt series.
	otherStage := in
	otherStage.Stage = "review panel"
	third, err := st.BeginStage(ctx, otherStage)
	if err != nil {
		t.Fatalf("BeginStage for a different stage: %v", err)
	}
	if third.Attempt != 1 {
		t.Errorf("BeginStage attempt for a different stage = %d, want 1 (its own series)", third.Attempt)
	}
}

// TestBeginStageRoundTripsSessionToken pins KAN-172, task 1's own
// acceptance criterion: the store persists whatever correlator BeginStage
// is given, and returns it unchanged both from BeginStage itself and from
// a later GetStageRun -- the two reads a harvest cycle (task 2) and a
// caller re-fetching the row would each perform.
func TestBeginStageRoundTripsSessionToken(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-session-token-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	in.SessionID = nil
	in.SessionToken = ptr("mf-session-token-round-trip")

	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if run.SessionToken == nil || *run.SessionToken != "mf-session-token-round-trip" {
		t.Errorf("BeginStage SessionToken = %v, want mf-session-token-round-trip", run.SessionToken)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	if got.SessionToken == nil || *got.SessionToken != "mf-session-token-round-trip" {
		t.Errorf("GetStageRun SessionToken = %v, want mf-session-token-round-trip", got.SessionToken)
	}
	// SessionID stays unset -- writing a sessionToken records only the
	// correlator, never a guessed session (design.md's kan-172 "a session
	// is never guessed"); binding it is task 2's job, not BeginStage's.
	if got.SessionID != nil {
		t.Errorf("SessionID = %v, want nil -- BeginStage must not resolve a session from a sessionToken", got.SessionID)
	}
}

// TestUnresolvedSessionTokensReturnsOnlyRowsAwaitingBinding is
// UnresolvedSessionTokens' own acceptance test: a row with a sessionToken and no
// session_id is returned; a row with neither, and a row already bound
// (both sessionToken and session_id set), are not -- exactly the partial
// index's own WHERE clause (stage_runs_unresolved_session_token, migration
// 0008), proven against the real predicate rather than assumed to match
// it.
func TestUnresolvedSessionTokensReturnsOnlyRowsAwaitingBinding(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-unresolved-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	unresolved := baseBeginInput(projectKey, "kan-1", "/myflow-do", "unresolved stage")
	unresolved.SessionID = nil
	unresolved.SessionToken = ptr("mf-unresolved")
	unresolvedRun, err := st.BeginStage(ctx, unresolved)
	if err != nil {
		t.Fatalf("BeginStage (unresolved): %v", err)
	}

	noToken := baseBeginInput(projectKey, "kan-1", "/myflow-do", "no sessionToken at all")
	noToken.SessionID = nil
	noToken.SessionToken = nil
	if _, err := st.BeginStage(ctx, noToken); err != nil {
		t.Fatalf("BeginStage (no sessionToken): %v", err)
	}

	alreadyBound := baseBeginInput(projectKey, "kan-1", "/myflow-do", "already bound")
	alreadyBound.SessionID = ptr("session-already-bound")
	alreadyBound.SessionToken = ptr("mf-already-bound")
	if _, err := st.BeginStage(ctx, alreadyBound); err != nil {
		t.Fatalf("BeginStage (already bound): %v", err)
	}

	got, err := st.UnresolvedSessionTokens(ctx)
	if err != nil {
		t.Fatalf("UnresolvedSessionTokens: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("UnresolvedSessionTokens returned %d rows, want exactly 1: %v", len(got), got)
	}
	if sessionToken, ok := got[unresolvedRun.ID]; !ok || sessionToken != "mf-unresolved" {
		t.Errorf("UnresolvedSessionTokens = %v, want {%d: mf-unresolved}", got, unresolvedRun.ID)
	}
}

// TestBindSessionSetsSessionID is BindSession's positive case: a stage
// run with no session_id gets exactly the session it is told to bind.
func TestBindSessionSetsSessionID(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-bind-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "to be bound")
	in.SessionID = nil
	in.SessionToken = ptr("mf-to-bind")
	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	bound, err := st.BindSession(ctx, "mf-to-bind", "session-newly-bound")
	if err != nil {
		t.Fatalf("BindSession: %v", err)
	}
	if bound != 1 {
		t.Fatalf("bound = %d on the first bind, want 1", bound)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	if got.SessionID == nil || *got.SessionID != "session-newly-bound" {
		t.Errorf("SessionID = %v, want session-newly-bound", got.SessionID)
	}
}

// TestBindSessionBindsEveryRunSharingAToken is KAN-172 task 4b's own
// acceptance criterion for the store half of "one token per session, not
// one per mark": when a session's token has been generated once and
// carried on several marks, each producing its own stage run row, a
// single BindSession call for that token binds every one of those rows at
// once -- not just whichever one a caller happened to name. This is the
// mechanism task 4b changes BindSession's own signature (stage run id ->
// token) to make possible.
func TestBindSessionBindsEveryRunSharingAToken(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-bind-shared-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	const token = "mf-shared-across-marks"
	first := baseBeginInput(projectKey, "kan-1", "/myflow-do", "first mark")
	first.SessionID = nil
	first.SessionToken = ptr(token)
	firstRun, err := st.BeginStage(ctx, first)
	if err != nil {
		t.Fatalf("BeginStage (first): %v", err)
	}

	second := baseBeginInput(projectKey, "kan-1", "/myflow-do", "second mark")
	second.SessionID = nil
	second.SessionToken = ptr(token)
	secondRun, err := st.BeginStage(ctx, second)
	if err != nil {
		t.Fatalf("BeginStage (second): %v", err)
	}

	bound, err := st.BindSession(ctx, token, "session-shared")
	if err != nil {
		t.Fatalf("BindSession: %v", err)
	}
	if bound != 2 {
		t.Fatalf("bound = %d, want 2 (both runs sharing the token)", bound)
	}

	for _, id := range []int64{firstRun.ID, secondRun.ID} {
		got, err := st.GetStageRun(ctx, id)
		if err != nil {
			t.Fatalf("GetStageRun(%d): %v", id, err)
		}
		if got.SessionID == nil || *got.SessionID != "session-shared" {
			t.Errorf("run %d SessionID = %v, want session-shared", id, got.SessionID)
		}
	}
}

// TestBindSessionIsOneWay is design.md's "unbinding never happens",
// pinned directly against the store: a second BindSession call against
// a token that is already bound reports zero rows bound and leaves the
// original session untouched, even when it names a different session
// entirely -- binding never overwrites what a first call already set.
func TestBindSessionIsOneWay(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-oneway-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "bound once")
	in.SessionID = nil
	in.SessionToken = ptr("mf-one-way")
	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	if _, err := st.BindSession(ctx, "mf-one-way", "session-first"); err != nil {
		t.Fatalf("BindSession (first): %v", err)
	}

	bound, err := st.BindSession(ctx, "mf-one-way", "session-second")
	if err != nil {
		t.Fatalf("BindSession (second): %v", err)
	}
	if bound != 0 {
		t.Fatalf("bound = %d on the second bind, want 0", bound)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	if got.SessionID == nil || *got.SessionID != "session-first" {
		t.Errorf("SessionID = %v, want unchanged session-first (binding is one-way)", got.SessionID)
	}
}

// TestBeginStageResolvesSessionIDFromAnAlreadyBoundToken is KAN-172 task
// 4b's other acceptance criterion: "a mark arriving with an already-bound
// token is bound at write time, doing no resolution work at all". Once a
// token has been bound to a session (by an earlier BindSession call, the
// harvester's own job), a later BeginStage call carrying that same token
// gets its session_id set immediately, from insertStageRunAndSupersede's
// own COALESCE, and never appears in UnresolvedSessionTokens at all.
func TestBeginStageResolvesSessionIDFromAnAlreadyBoundToken(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-resolve-at-insert-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	const token = "mf-already-resolved"
	first := baseBeginInput(projectKey, "kan-1", "/myflow-do", "first mark")
	first.SessionID = nil
	first.SessionToken = ptr(token)
	if _, err := st.BeginStage(ctx, first); err != nil {
		t.Fatalf("BeginStage (first): %v", err)
	}
	if _, err := st.BindSession(ctx, token, "session-resolved"); err != nil {
		t.Fatalf("BindSession: %v", err)
	}

	second := baseBeginInput(projectKey, "kan-1", "/myflow-do", "second mark")
	second.SessionID = nil
	second.SessionToken = ptr(token)
	secondRun, err := st.BeginStage(ctx, second)
	if err != nil {
		t.Fatalf("BeginStage (second): %v", err)
	}
	if secondRun.SessionID == nil || *secondRun.SessionID != "session-resolved" {
		t.Fatalf("second run SessionID = %v, want session-resolved (resolved at insert time)", secondRun.SessionID)
	}

	unresolved, err := st.UnresolvedSessionTokens(ctx)
	if err != nil {
		t.Fatalf("UnresolvedSessionTokens: %v", err)
	}
	if _, ok := unresolved[secondRun.ID]; ok {
		t.Errorf("second run appears in UnresolvedSessionTokens = %v, want absent: its session was already known at insert time", unresolved)
	}
}

// TestBeginStageSupersedesAnEarlierOpenRunOfTheSameSession is
// design.md's begin-supersedes-open-runs decision, pinned against the
// store: a begin mark closes every still-open run of its own session
// token that started no later than it, recording outcome "superseded"
// and an end instant equal to the new run's own start -- the incident
// itself, reduced: an orphan stage of one change (kan-175) left open by
// a dropped end mark, and the live stage of another change (kan-184)
// that the same session moved on to.
func TestBeginStageSupersedesAnEarlierOpenRunOfTheSameSession(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-supersede-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-175")
	seedChange(t, st, projectKey, "kan-184")

	const token = "mf-supersede-token"

	a := baseBeginInput(projectKey, "kan-175", "/myflow-fast", "finish.verify-merge")
	a.SessionToken = ptr(token)
	a.StartedAt = time.Date(2026, 8, 16, 10, 11, 1, 0, time.UTC)
	runA, err := st.BeginStage(ctx, a)
	if err != nil {
		t.Fatalf("BeginStage (A): %v", err)
	}

	b := baseBeginInput(projectKey, "kan-184", "/myflow-do", "SDD + TDD per task")
	b.SessionToken = ptr(token)
	b.StartedAt = a.StartedAt.Add(2 * time.Hour)
	runB, err := st.BeginStage(ctx, b)
	if err != nil {
		t.Fatalf("BeginStage (B): %v", err)
	}

	gotA, err := st.GetStageRun(ctx, runA.ID)
	if err != nil {
		t.Fatalf("GetStageRun (A): %v", err)
	}
	if gotA.Outcome == nil || *gotA.Outcome != "superseded" {
		t.Errorf("run A outcome = %v, want superseded", gotA.Outcome)
	}
	if gotA.EndedAt == nil || !gotA.EndedAt.Equal(b.StartedAt) {
		t.Errorf("run A EndedAt = %v, want %v (B's started_at)", gotA.EndedAt, b.StartedAt)
	}

	gotB, err := st.GetStageRun(ctx, runB.ID)
	if err != nil {
		t.Fatalf("GetStageRun (B): %v", err)
	}
	if gotB.EndedAt != nil {
		t.Errorf("run B EndedAt = %v, want nil (still open)", gotB.EndedAt)
	}
	if gotB.Outcome != nil {
		t.Errorf("run B Outcome = %v, want nil (still open)", gotB.Outcome)
	}
}

// TestBeginStageLeavesAnOpenRunThatStartedLaterAlone is
// design.md's supersede-guarded-on-start-order decision: a journalled
// begin mark replayed later carries its original, older start instant,
// so it must not close a run that genuinely started after it -- doing so
// would close the live stage the session is actually in, the same class
// of damage this change exists to stop.
func TestBeginStageLeavesAnOpenRunThatStartedLaterAlone(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-replay-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	const token = "mf-replay-token"

	live := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	live.SessionToken = ptr(token)
	live.StartedAt = time.Date(2026, 8, 16, 11, 0, 0, 0, time.UTC)
	liveRun, err := st.BeginStage(ctx, live)
	if err != nil {
		t.Fatalf("BeginStage (live): %v", err)
	}

	replay := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
	replay.SessionToken = ptr(token)
	replay.StartedAt = time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	if _, err := st.BeginStage(ctx, replay); err != nil {
		t.Fatalf("BeginStage (replay): %v", err)
	}

	got, err := st.GetStageRun(ctx, liveRun.ID)
	if err != nil {
		t.Fatalf("GetStageRun (live): %v", err)
	}
	if got.EndedAt != nil || got.Outcome != nil {
		t.Errorf("live run closed by an older replayed begin: EndedAt=%v Outcome=%v, want both nil", got.EndedAt, got.Outcome)
	}
}

// TestBeginStageSupersedesAnOpenRunStartedAtTheSameInstant pins the
// supersede UPDATE's `started_at <= $2` guard at its equal-instant
// boundary -- the shape a real journal replay actually carries: the CLI
// captures StartedAt once, before the RPC attempt (cmd/flow/stage.go's
// runStageBegin), and internal/reconcile.go's applyStageMarkEntry replays
// that same original instant unchanged, so the replay's StartedAt equals
// the attempt it replays rather than falling strictly before it (see
// attribute.go's bestWindow doc comment for the same correction).
// TestBeginStageLeavesAnOpenRunThatStartedLaterAlone already covers the
// strictly-earlier replay start; this covers the equal one.
func TestBeginStageSupersedesAnOpenRunStartedAtTheSameInstant(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-same-instant-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	const token = "mf-same-instant-token"
	startedAt := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)

	first := baseBeginInput(projectKey, "kan-1", "/myflow-do", "first attempt")
	first.SessionToken = ptr(token)
	first.StartedAt = startedAt
	firstRun, err := st.BeginStage(ctx, first)
	if err != nil {
		t.Fatalf("BeginStage (first): %v", err)
	}

	replay := baseBeginInput(projectKey, "kan-1", "/myflow-do", "replayed begin")
	replay.SessionToken = ptr(token)
	replay.StartedAt = startedAt
	if _, err := st.BeginStage(ctx, replay); err != nil {
		t.Fatalf("BeginStage (replay): %v", err)
	}

	got, err := st.GetStageRun(ctx, firstRun.ID)
	if err != nil {
		t.Fatalf("GetStageRun (first): %v", err)
	}
	if got.Outcome == nil || *got.Outcome != "superseded" {
		t.Errorf("first run Outcome = %v, want superseded", got.Outcome)
	}
	if got.EndedAt == nil || !got.EndedAt.Equal(startedAt) {
		t.Errorf("first run EndedAt = %v, want %v (the shared started_at)", got.EndedAt, startedAt)
	}
}

// TestBeginStageWithNoSessionTokenSupersedesNothing is
// design.md's supersede-keyed-on-token decision, pinned in both
// directions: an open run recorded with no session token stays open
// when another run begins, and a begin mark carrying no session token
// closes nothing -- a NULL token matches no row through SQL's own `=`,
// so two NULL tokens never appear to match each other.
func TestBeginStageWithNoSessionTokenSupersedesNothing(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-notoken-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	noToken := baseBeginInput(projectKey, "kan-1", "/myflow-do", "no token open run")
	noToken.SessionToken = nil
	noToken.StartedAt = time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	noTokenRun, err := st.BeginStage(ctx, noToken)
	if err != nil {
		t.Fatalf("BeginStage (no token): %v", err)
	}

	// A begin carrying a real token must not close the no-token run.
	withToken := baseBeginInput(projectKey, "kan-1", "/myflow-do", "with token")
	withToken.SessionToken = ptr("mf-notoken-other")
	withToken.StartedAt = noToken.StartedAt.Add(time.Hour)
	if _, err := st.BeginStage(ctx, withToken); err != nil {
		t.Fatalf("BeginStage (with token): %v", err)
	}

	got, err := st.GetStageRun(ctx, noTokenRun.ID)
	if err != nil {
		t.Fatalf("GetStageRun (no token run): %v", err)
	}
	if got.EndedAt != nil || got.Outcome != nil {
		t.Errorf("no-token run was closed by a begin carrying a token: EndedAt=%v Outcome=%v, want both nil", got.EndedAt, got.Outcome)
	}

	// A begin carrying no token itself must close nothing either.
	anotherNoToken := baseBeginInput(projectKey, "kan-1", "/myflow-do", "another no-token begin")
	anotherNoToken.SessionToken = nil
	anotherNoToken.StartedAt = withToken.StartedAt.Add(time.Hour)
	if _, err := st.BeginStage(ctx, anotherNoToken); err != nil {
		t.Fatalf("BeginStage (another no token): %v", err)
	}

	gotAgain, err := st.GetStageRun(ctx, noTokenRun.ID)
	if err != nil {
		t.Fatalf("GetStageRun (no token run, again): %v", err)
	}
	if gotAgain.EndedAt != nil || gotAgain.Outcome != nil {
		t.Errorf("no-token run was closed by a begin carrying no token: EndedAt=%v Outcome=%v, want both nil", gotAgain.EndedAt, gotAgain.Outcome)
	}
}

func TestBeginStageUnknownChangeIsNotFound(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	_, err := st.BeginStage(ctx, baseBeginInput("no-such-project", "no-such-change", "/myflow-do", "SDD + TDD per task"))
	if !errors.Is(err, store.ErrChangeNotFound) {
		t.Fatalf("BeginStage(unknown change) error = %v, want errors.Is(_, store.ErrChangeNotFound)", err)
	}
}

// TestConcurrentBeginStageDoesNotCollide reproduces the failure a
// check-then-insert attempt allocator would have: many goroutines call
// BeginStage for the exact same (change, command, stage) triple at once.
// If the next attempt were computed by a separate read followed by a
// gated insert, two goroutines could both read the same current maximum
// and insert the same attempt number, tripping the unique constraint for
// one of them (a bug) or -- worse, if the code swallowed that error --
// silently losing one attempt's row. BeginStage instead allocates the
// attempt inside the insert and retries on the constraint's own
// unique-violation, so every goroutine must come away with an attempt
// number, and those numbers must be exactly {1..N} with no duplicate and
// no gap.
func TestConcurrentBeginStageDoesNotCollide(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-concurrent-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	const writers = 30
	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")

	var wg sync.WaitGroup
	attempts := make([]int, writers)
	errs := make([]error, writers)

	for i := range writers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			callCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			run, err := st.BeginStage(callCtx, in)
			errs[i] = err
			if err == nil {
				attempts[i] = run.Attempt
			}
		}(i)
	}
	wg.Wait()

	seen := make(map[int]int)
	for i, err := range errs {
		if err != nil {
			t.Fatalf("writer %d: BeginStage: %v", i, err)
		}
		seen[attempts[i]]++
	}
	for attempt := 1; attempt <= writers; attempt++ {
		if seen[attempt] != 1 {
			t.Errorf("attempt %d was allocated %d times, want exactly 1", attempt, seen[attempt])
		}
	}
	if len(seen) != writers {
		t.Errorf("got %d distinct attempt numbers, want %d", len(seen), writers)
	}
}

// TestConcurrentBeginStageForOneSessionLeavesOneOpenRun is fix round 4's
// F8, demonstrated: design.md's supersede-serialised-per-session decision
// exists because atomicity is not serialisability. Every writer here
// shares the *same* session token and the *same* StartedAt -- the shape
// the reconciler's replay actually produces (a live begin racing a
// replayed copy of the identical journalled request) -- rather than
// distinct, increasing StartedAt values: with a shared StartedAt, the
// supersede UPDATE's own `started_at <= $2` guard is satisfied regardless
// of which writer's transaction happens to commit first, so exactly one
// open run is the only possible correct outcome no matter how the
// goroutines interleave. Distinct StartedAt values would not give that
// guarantee -- the guard is deliberately asymmetric (an older-started
// replay must never supersede a run that genuinely started later), so a
// begin that commits after another one with a *later* StartedAt already
// went open would leave both open, which is correct per
// supersede-guarded-on-start-order and not what this test is about.
//
// Each writer's own (command, stage) is distinct, though -- unlike
// TestConcurrentBeginStageDoesNotCollide, which shares one triple on
// purpose to exercise attempt allocation. Sharing a triple here would let
// BeginStage's own attempt-collision retry loop (stageRunsAttemptConstraint)
// absorb most of the contention before it ever reaches the supersede at
// all: measured against the pre-fix code, 150 writers sharing one triple
// left exactly one run open on every run, while 30 writers each on their
// own triple reliably left several open, because nothing but the
// supersede itself was left to serialise them.
//
// Without the transaction's advisory lock, two writers whose inserts
// happen to overlap in time under READ COMMITTED can each fail to see
// the other's still-uncommitted row when their own supersede UPDATE
// runs, so neither closes the other and both stay open -- this is what
// this test fails on before the lock exists.
func TestConcurrentBeginStageForOneSessionLeavesOneOpenRun(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-concurrent-supersede-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	const writers = 30
	const token = "mf-concurrent-supersede-token"
	startedAt := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)

	var wg sync.WaitGroup
	ids := make([]int64, writers)
	errs := make([]error, writers)
	// start is closed once every writer has reached its own starting
	// block, so all `writers` calls actually race the database at once
	// instead of trickling in staggered by ordinary goroutine scheduling
	// -- without this, the transactions rarely overlap enough on a fast
	// local connection to exercise the race at all.
	var ready sync.WaitGroup
	ready.Add(writers)
	start := make(chan struct{})

	for i := range writers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ready.Done()
			<-start

			callCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			in := baseBeginInput(projectKey, "kan-1", "/myflow-do", fmt.Sprintf("stage-%d", i))
			in.SessionToken = ptr(token)
			in.StartedAt = startedAt
			run, err := st.BeginStage(callCtx, in)
			errs[i] = err
			if err == nil {
				ids[i] = run.ID
			}
		}(i)
	}
	ready.Wait()
	close(start)
	wg.Wait()

	openCount, supersededCount := 0, 0
	for i, err := range errs {
		if err != nil {
			t.Fatalf("writer %d: BeginStage: %v", i, err)
		}
		got, err := st.GetStageRun(ctx, ids[i])
		if err != nil {
			t.Fatalf("GetStageRun %d: %v", ids[i], err)
		}
		switch {
		case got.EndedAt == nil && got.Outcome == nil:
			openCount++
		case got.Outcome != nil && *got.Outcome == "superseded" && got.EndedAt != nil && got.EndedAt.Equal(startedAt):
			supersededCount++
		default:
			t.Errorf("run %d in unexpected state: EndedAt=%v Outcome=%v", ids[i], got.EndedAt, got.Outcome)
		}
	}
	if openCount != 1 {
		t.Errorf("open runs = %d, want exactly 1 (every writer shared one session token and start instant)", openCount)
	}
	if supersededCount != writers-1 {
		t.Errorf("superseded runs = %d, want %d", supersededCount, writers-1)
	}
}

func TestMergeMetricsPreservesOtherKeys(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-merge-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	// The harvester writes token keys first.
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"tokens":{"main":{"input":100,"output":50}}}`)); err != nil {
		t.Fatalf("MergeMetrics (tokens): %v", err)
	}
	// A stage-end mark writes outcome keys afterward.
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"duration_ms":4200,"fast_mode":false}`)); err != nil {
		t.Fatalf("MergeMetrics (duration): %v", err)
	}
	// A later merge overwrites a key it shares, and leaves the rest alone.
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"tokens":{"main":{"input":150,"output":50,"cache_read":10}}}`)); err != nil {
		t.Fatalf("MergeMetrics (updated tokens): %v", err)
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
		t.Errorf("metrics lost duration_ms after a later merge touched only tokens: %s", got.Metrics)
	}
	if _, ok := bag["fast_mode"]; !ok {
		t.Errorf("metrics lost fast_mode after a later merge touched only tokens: %s", got.Metrics)
	}

	var bagTokens struct {
		Main struct {
			Input     float64 `json:"input"`
			Output    float64 `json:"output"`
			CacheRead float64 `json:"cache_read"`
		} `json:"main"`
	}
	if err := json.Unmarshal(bag["tokens"], &bagTokens); err != nil {
		t.Fatalf("unmarshal tokens: %v", err)
	}
	if bagTokens.Main.Input != 150 {
		t.Errorf("tokens.main.input = %v, want 150 (last write wins on a shared key)", bagTokens.Main.Input)
	}
	if bagTokens.Main.CacheRead != 10 {
		t.Errorf("tokens.main.cache_read = %v, want 10", bagTokens.Main.CacheRead)
	}
}

// TestMergeMetricsDeepMergesNestedSiblings reproduces the finding against
// the shallow `metrics || patch` implementation: two writers each touch a
// different sub-key of the same nested "tokens" object -- one the
// harvester's main-thread token counts, the other its sidechain counts,
// exactly as two separate CommitHarvestBatch calls would. A shallow
// top-level concatenation replaces "tokens" wholesale on the second write,
// discarding the first writer's main bucket. The merge must combine the
// nested object's keys instead, so neither writer needs to know what the
// other already stored there.
func TestMergeMetricsDeepMergesNestedSiblings(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-deepmerge-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"tokens":{"main":{"input":5,"output":9}}}`)); err != nil {
		t.Fatalf("MergeMetrics (main): %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"tokens":{"sidechain":{"input":0}}}`)); err != nil {
		t.Fatalf("MergeMetrics (sidechain): %v", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}

	var bag struct {
		Tokens struct {
			Main struct {
				Input  *float64 `json:"input"`
				Output *float64 `json:"output"`
			} `json:"main"`
			Sidechain struct {
				Input *float64 `json:"input"`
			} `json:"sidechain"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}

	if bag.Tokens.Main.Input == nil || *bag.Tokens.Main.Input != 5 {
		t.Errorf("tokens.main.input = %v, want 5 -- a later write to tokens.sidechain must not erase it", bag.Tokens.Main.Input)
	}
	if bag.Tokens.Main.Output == nil || *bag.Tokens.Main.Output != 9 {
		t.Errorf("tokens.main.output = %v, want 9 -- a later write to tokens.sidechain must not erase it", bag.Tokens.Main.Output)
	}
	if bag.Tokens.Sidechain.Input == nil || *bag.Tokens.Sidechain.Input != 0 {
		t.Errorf("tokens.sidechain.input = %v, want 0 (a recorded zero, not absence)", bag.Tokens.Sidechain.Input)
	}
}

// TestMergeMetricsNonObjectValueReplaces asserts the other half of the
// merge rule: when patch's value at a key is not itself an object, it
// replaces whatever was stored there -- including an object -- rather than
// being merged into it. There is nothing under a non-object value to
// preserve.
func TestMergeMetricsNonObjectValueReplaces(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-nonobjreplace-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"tokens":{"main":{"input":5,"output":9}}}`)); err != nil {
		t.Fatalf("MergeMetrics (object): %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"tokens":false}`)); err != nil {
		t.Fatalf("MergeMetrics (non-object replaces): %v", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag map[string]json.RawMessage
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}
	if string(bag["tokens"]) != "false" {
		t.Errorf(`tokens = %s, want "false" -- a non-object patch value must replace the stored object entirely`, bag["tokens"])
	}
}

// TestMergeMetricsRejectsNilPatch asserts that a nil patch is refused with
// a typed error before it ever reaches SQL. A nil json.RawMessage
// marshals to SQL NULL, and jsonb_deep_merge(metrics, NULL) returns NULL,
// which stage_runs.metrics' NOT NULL constraint would then reject as a raw
// Postgres error -- MergeMetrics catches this in Go instead.
func TestMergeMetricsRejectsNilPatch(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-nilpatch-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	err = st.MergeMetrics(ctx, run.ID, nil)
	if !errors.Is(err, store.ErrNilMetricsPatch) {
		t.Fatalf("MergeMetrics(nil patch) error = %v, want errors.Is(_, store.ErrNilMetricsPatch)", err)
	}

	// The stage run's metrics must be untouched, not corrupted or nulled.
	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	if string(got.Metrics) != "{}" {
		t.Errorf("metrics = %s after a rejected nil patch, want unchanged {}", got.Metrics)
	}
}

func TestMergeMetricsUnknownStageRunIsNotFound(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	err := st.MergeMetrics(ctx, 99999999, json.RawMessage(`{"a":1}`))
	if !errors.Is(err, store.ErrStageRunNotFound) {
		t.Fatalf("MergeMetrics(unknown id) error = %v, want errors.Is(_, store.ErrStageRunNotFound)", err)
	}
}

func TestEndStageRecordsOutcome(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-end-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	endedAt := run.StartedAt.Add(5 * time.Minute)
	if err := st.EndStage(ctx, run.ID, endedAt, "completed"); err != nil {
		t.Fatalf("EndStage: %v", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	if got.EndedAt == nil || !got.EndedAt.Equal(endedAt) {
		t.Errorf("EndedAt = %v, want %v", got.EndedAt, endedAt)
	}
	if got.Outcome == nil || *got.Outcome != "completed" {
		t.Errorf("Outcome = %v, want completed", got.Outcome)
	}
}

func TestEndStageUnknownStageRunIsNotFound(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	err := st.EndStage(ctx, 99999999, time.Now(), "completed")
	if !errors.Is(err, store.ErrStageRunNotFound) {
		t.Fatalf("EndStage(unknown id) error = %v, want errors.Is(_, store.ErrStageRunNotFound)", err)
	}
}

// TestEndStageRefusesARunAlreadyClosed is fix round 4's F9, pinned at the
// store: design.md's end-mark-cannot-resurrect decision. A row EndStage
// has already closed must not be reopened, or have its outcome silently
// replaced, by a second call landing on the same id -- the shape a
// supersede racing a slow end mark's lookup-then-close gap produces
// (ApplyEndStageMark, internal/api/stages.go). The second call must
// report a distinct, typed refusal rather than success, and must leave
// the row exactly as the first call left it.
func TestEndStageRefusesARunAlreadyClosed(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-end-twice-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	firstEnd := run.StartedAt.Add(5 * time.Minute)
	if err := st.EndStage(ctx, run.ID, firstEnd, "completed"); err != nil {
		t.Fatalf("EndStage (first): %v", err)
	}

	secondEnd := run.StartedAt.Add(10 * time.Minute)
	err = st.EndStage(ctx, run.ID, secondEnd, "superseded")
	if !errors.Is(err, store.ErrStageRunAlreadyClosed) {
		t.Fatalf("EndStage (second) error = %v, want errors.Is(_, store.ErrStageRunAlreadyClosed)", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	if got.EndedAt == nil || !got.EndedAt.Equal(firstEnd) {
		t.Errorf("EndedAt = %v, want %v (the first, definitive close, untouched by the second call)", got.EndedAt, firstEnd)
	}
	if got.Outcome == nil || *got.Outcome != "completed" {
		t.Errorf("Outcome = %v, want completed (the first call's outcome, not the second call's)", got.Outcome)
	}
}

// TestSweepAbandonedClosesSilentStages asserts that SweepAbandoned closes
// only stage runs that are both open (no end mark) and started before the
// silence cutoff -- never a stage run that already ended, and never one
// that started after the cutoff.
func TestSweepAbandonedClosesSilentStages(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-sweep-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	silent := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	silent.StartedAt = time.Now().Add(-2 * time.Hour)
	silentRun, err := st.BeginStage(ctx, silent)
	if err != nil {
		t.Fatalf("BeginStage(silent): %v", err)
	}

	ended := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
	ended.StartedAt = time.Now().Add(-2 * time.Hour)
	endedRun, err := st.BeginStage(ctx, ended)
	if err != nil {
		t.Fatalf("BeginStage(ended): %v", err)
	}
	if err := st.EndStage(ctx, endedRun.ID, time.Now().Add(-90*time.Minute), "completed"); err != nil {
		t.Fatalf("EndStage(ended): %v", err)
	}

	recent := baseBeginInput(projectKey, "kan-1", "/myflow-do", "finish")
	recent.StartedAt = time.Now().Add(-1 * time.Minute)
	recentRun, err := st.BeginStage(ctx, recent)
	if err != nil {
		t.Fatalf("BeginStage(recent): %v", err)
	}

	cutoff := time.Now().Add(-1 * time.Hour)
	n, err := st.SweepAbandoned(ctx, cutoff)
	if err != nil {
		t.Fatalf("SweepAbandoned: %v", err)
	}
	if n != 1 {
		t.Fatalf("SweepAbandoned closed %d rows, want 1 (only the silent, still-open one)", n)
	}

	got, err := st.GetStageRun(ctx, silentRun.ID)
	if err != nil {
		t.Fatalf("GetStageRun(silent): %v", err)
	}
	if got.Outcome == nil || *got.Outcome != "abandoned" {
		t.Errorf("silent stage run outcome = %v, want abandoned", got.Outcome)
	}
	if got.EndedAt == nil {
		t.Errorf("silent stage run has no EndedAt after being swept")
	}

	gotEnded, err := st.GetStageRun(ctx, endedRun.ID)
	if err != nil {
		t.Fatalf("GetStageRun(ended): %v", err)
	}
	if gotEnded.Outcome == nil || *gotEnded.Outcome != "completed" {
		t.Errorf("already-ended stage run outcome changed to %v, want it left as completed", gotEnded.Outcome)
	}

	gotRecent, err := st.GetStageRun(ctx, recentRun.ID)
	if err != nil {
		t.Fatalf("GetStageRun(recent): %v", err)
	}
	if gotRecent.Outcome != nil {
		t.Errorf("recent, still-open stage run outcome = %v, want nil (not swept: started after cutoff)", gotRecent.Outcome)
	}
}

func TestPriceFreezesCostAndVersion(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	in.StartedAt = time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	// cache_creation_1h (not the collapsed "cache_creation") is what Price
	// now reads: this run's cache-creation usage is entirely a 1-hour
	// write, which task 23 exists to price at its own rate rather than
	// the 5-minute one. "cache_creation" itself is still present,
	// unchanged in meaning, purely to prove Price leaves it untouched.
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"claude-opus-5": {
				"tokens": {"main": {"input": 1000000, "output": 500000, "cache_creation": 200000, "cache_creation_1h": 200000, "cache_read": 4000000}}
			}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics(tokens): %v", err)
	}

	oldRate := store.PricingRate{
		Model:               "claude-opus-5",
		EffectiveFrom:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok:        3,
		OutputPerMTok:       15,
		CacheWritePerMTok:   3.75,
		CacheWrite5mPerMTok: 3.75,
		CacheWrite1hPerMTok: ptr(3.75),
		CacheReadPerMTok:    0.3,
	}
	if err := st.PutPricing(ctx, oldRate); err != nil {
		t.Fatalf("PutPricing(old): %v", err)
	}

	if err := st.Price(ctx, run.ID); err != nil {
		t.Fatalf("Price: %v", err)
	}

	priced, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag struct {
		CostUSD float64 `json:"cost_usd"`
		Models  map[string]struct {
			CostUSD        float64 `json:"cost_usd"`
			PricingVersion string  `json:"pricing_version"`
			Tokens         struct {
				Main struct {
					Input         float64 `json:"input"`
					Output        float64 `json:"output"`
					CacheCreation float64 `json:"cache_creation"`
					CacheRead     float64 `json:"cache_read"`
				} `json:"main"`
			} `json:"tokens"`
		} `json:"models"`
	}
	if err := json.Unmarshal(priced.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal priced metrics: %v", err)
	}

	wantCost := 1*3 + 0.5*15 + 0.2*3.75 + 4*0.3
	opus, ok := bag.Models["claude-opus-5"]
	if !ok {
		t.Fatalf("models.claude-opus-5 missing from priced metrics: %s", priced.Metrics)
	}
	if diff := opus.CostUSD - wantCost; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("models.claude-opus-5.cost_usd = %v, want %v", opus.CostUSD, wantCost)
	}
	wantVersion := "claude-opus-5@2026-01-01T00:00:00Z"
	if opus.PricingVersion != wantVersion {
		t.Errorf("models.claude-opus-5.pricing_version = %q, want %q", opus.PricingVersion, wantVersion)
	}
	// The single-model run's whole cost is this one bucket's cost, so the
	// top-level total (their sum) matches it exactly.
	if diff := bag.CostUSD - wantCost; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("cost_usd = %v, want %v", bag.CostUSD, wantCost)
	}
	// Token counts survive pricing untouched, so history can be re-priced.
	if opus.Tokens.Main.Input != 1000000 || opus.Tokens.Main.Output != 500000 ||
		opus.Tokens.Main.CacheCreation != 200000 || opus.Tokens.Main.CacheRead != 4000000 {
		t.Errorf("token counts changed after Price: %+v", opus.Tokens)
	}

	// A later, higher rate takes effect for a subsequent Price call, but a
	// figure already frozen and never re-priced does not move under the
	// reader: re-fetching the row without calling Price again must not
	// reflect the new rate.
	newRate := oldRate
	newRate.EffectiveFrom = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	newRate.InputPerMTok = 30
	if err := st.PutPricing(ctx, newRate); err != nil {
		t.Fatalf("PutPricing(new): %v", err)
	}

	stillOld, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun after new pricing published: %v", err)
	}
	var stillBag struct {
		CostUSD float64 `json:"cost_usd"`
	}
	if err := json.Unmarshal(stillOld.Metrics, &stillBag); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if diff := stillBag.CostUSD - wantCost; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("cost_usd moved to %v after a newer pricing row was published without re-pricing; want it to stay frozen at %v", stillBag.CostUSD, wantCost)
	}
}

// TestPriceTwoModelsAgainstTwoRates covers step 3's central case: a
// mixed-model run (the review panel's own parent-vs-reviewer split) is
// priced per bucket against each model's own rate, and the top-level
// cost_usd is their sum -- a mixed run's real total, not one model's
// rate applied to every token.
func TestPriceTwoModelsAgainstTwoRates(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-twomodel-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
	in.StartedAt = time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"claude-opus-5":   {"tokens": {"main": {"input": 1000000}}},
			"claude-sonnet-5": {"tokens": {"sidechain": {"input": 2000000}}}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}

	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-opus-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 3, OutputPerMTok: 15, CacheWritePerMTok: 3.75, CacheReadPerMTok: 0.3,
	}); err != nil {
		t.Fatalf("PutPricing(opus): %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-sonnet-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 1, OutputPerMTok: 5, CacheWritePerMTok: 1.25, CacheReadPerMTok: 0.1,
	}); err != nil {
		t.Fatalf("PutPricing(sonnet): %v", err)
	}

	if err := st.Price(ctx, run.ID); err != nil {
		t.Fatalf("Price: %v", err)
	}

	priced, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag struct {
		CostUSD float64 `json:"cost_usd"`
		Models  map[string]struct {
			CostUSD float64 `json:"cost_usd"`
		} `json:"models"`
	}
	if err := json.Unmarshal(priced.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	const eps = 1e-9
	wantOpus := 1.0 * 3
	wantSonnet := 2.0 * 1
	if diff := bag.Models["claude-opus-5"].CostUSD - wantOpus; diff > eps || diff < -eps {
		t.Errorf("models.claude-opus-5.cost_usd = %v, want %v", bag.Models["claude-opus-5"].CostUSD, wantOpus)
	}
	if diff := bag.Models["claude-sonnet-5"].CostUSD - wantSonnet; diff > eps || diff < -eps {
		t.Errorf("models.claude-sonnet-5.cost_usd = %v, want %v", bag.Models["claude-sonnet-5"].CostUSD, wantSonnet)
	}
	wantTotal := wantOpus + wantSonnet
	if diff := bag.CostUSD - wantTotal; diff > eps || diff < -eps {
		t.Errorf("cost_usd = %v, want %v (the sum of both models' own bucket cost, not one rate applied to every token)", bag.CostUSD, wantTotal)
	}
}

// TestPriceOneUnpriceableBucketOmitsTopLevelTotal is step 3's other
// central case: when one of a mixed run's model buckets has no rate in
// effect, Price writes the bucket it could price, omits the one it could
// not, and omits the top-level cost_usd entirely rather than writing a
// partial sum that reads like a complete one -- a partial total is
// indistinguishable from a correct one at every layer above it, which
// makes it the most dangerous possible output here.
func TestPriceOneUnpriceableBucketOmitsTopLevelTotal(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-partial-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"claude-opus-5": {"tokens": {"main": {"input": 1000000}}},
			"no-such-model": {"tokens": {"sidechain": {"input": 500000}}}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-opus-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 3, OutputPerMTok: 15, CacheWritePerMTok: 3.75, CacheReadPerMTok: 0.3,
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	err = st.Price(ctx, run.ID)
	if !errors.Is(err, store.ErrPricingNotFound) {
		t.Fatalf("Price(one bucket unpriceable) error = %v, want errors.Is(_, store.ErrPricingNotFound)", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag map[string]json.RawMessage
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}
	if _, ok := bag["cost_usd"]; ok {
		t.Errorf("cost_usd = %s, want no top-level cost_usd key at all when one model bucket could not be priced", bag["cost_usd"])
	}

	var models map[string]json.RawMessage
	if err := json.Unmarshal(bag["models"], &models); err != nil {
		t.Fatalf("unmarshal models: %v", err)
	}
	var opus struct {
		CostUSD float64 `json:"cost_usd"`
	}
	if err := json.Unmarshal(models["claude-opus-5"], &opus); err != nil {
		t.Fatalf("unmarshal models.claude-opus-5: %v", err)
	}
	if opus.CostUSD != 3 {
		t.Errorf("models.claude-opus-5.cost_usd = %v, want 3 (the bucket that could be priced is still written)", opus.CostUSD)
	}
	var noSuch struct {
		CostUSD *float64 `json:"cost_usd"`
	}
	if err := json.Unmarshal(models["no-such-model"], &noSuch); err != nil {
		t.Fatalf("unmarshal models.no-such-model: %v", err)
	}
	if noSuch.CostUSD != nil {
		t.Errorf("models.no-such-model.cost_usd = %v, want absent (no rate was ever in effect for it)", *noSuch.CostUSD)
	}
}

func TestPriceWithoutTokensIsUnavailable(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-unavail-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	// Harness with no transcript: no models, no tokens recorded at all.

	err = st.Price(ctx, run.ID)
	if !errors.Is(err, store.ErrTokensUnavailable) {
		t.Fatalf("Price(no tokens) error = %v, want errors.Is(_, store.ErrTokensUnavailable)", err)
	}
}

// TestPriceWithoutChargeableFieldIsUnavailable covers the case where a
// "models" bucket is present but every model in it carries only
// thread-attribution bookkeeping (thinking tokens; TokenDelta's own
// Bucket type carries a "thinking" field that is real, recorded usage
// but never billed) with none of the four chargeable fields
// (input/output/cache_creation/cache_read). Without a check for this,
// Price's cost accumulator never advances past its zero initialiser and
// a literal cost_usd: 0 is written, indistinguishable from a genuinely
// free stage. Price must refuse instead.
func TestPriceWithoutChargeableFieldIsUnavailable(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-nocharge-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"models":{"claude-opus-5":{"tokens":{"main":{"thinking":50}}}}}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-opus-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 3, OutputPerMTok: 15, CacheWritePerMTok: 3.75, CacheReadPerMTok: 0.3,
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	err = st.Price(ctx, run.ID)
	if !errors.Is(err, store.ErrTokensUnavailable) {
		t.Fatalf("Price(tokens present, no chargeable field) error = %v, want errors.Is(_, store.ErrTokensUnavailable)", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag map[string]json.RawMessage
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}
	if _, ok := bag["cost_usd"]; ok {
		t.Errorf("cost_usd was written (%s) for a run with no chargeable token field; want no cost_usd key at all", bag["cost_usd"])
	}
}

func TestPriceWithoutPricingRowIsNotFound(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-nopricing-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"models":{"no-such-model":{"tokens":{"main":{"input":10}}}}}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}

	err = st.Price(ctx, run.ID)
	if !errors.Is(err, store.ErrPricingNotFound) {
		t.Fatalf("Price(no pricing row) error = %v, want errors.Is(_, store.ErrPricingNotFound)", err)
	}
}

// TestPriceChargesFiveMinuteAndOneHourWritesAtTheirOwnRates is task 23's
// central defect, priced directly: a run whose cache-creation usage is
// split between a 5-minute write and a 1-hour write must charge each
// portion at its own rate, not one collapsed rate applied to the whole
// total -- the 37.5% understatement this task's plan-provenance note
// measures if the two ever get collapsed onto the cheaper 5m rate.
func TestPriceChargesFiveMinuteAndOneHourWritesAtTheirOwnRates(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-split-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"claude-opus-5": {
				"tokens": {"main": {"cache_creation": 3000000, "cache_creation_5m": 1000000, "cache_creation_1h": 2000000}}
			}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-opus-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 5, OutputPerMTok: 25,
		CacheWritePerMTok: 6.25, CacheWrite5mPerMTok: 6.25, CacheWrite1hPerMTok: ptr(10.0),
		CacheReadPerMTok: 0.50,
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	if err := st.Price(ctx, run.ID); err != nil {
		t.Fatalf("Price: %v", err)
	}

	priced, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag struct {
		CostUSD float64 `json:"cost_usd"`
	}
	if err := json.Unmarshal(priced.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// 1 Mtok @ 6.25 (5m) + 2 Mtok @ 10 (1h) = 6.25 + 20 = 26.25. Collapsing
	// onto the 5m rate alone would give 3*6.25 = 18.75, a 28.6% understatement
	// of this bucket's true cost -- the exact failure mode this test exists
	// to catch.
	wantCost := 1*6.25 + 2*10.0
	if diff := bag.CostUSD - wantCost; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("cost_usd = %v, want %v", bag.CostUSD, wantCost)
	}
}

// TestPriceUnknownCacheSplitIsUnpriceable is the absent-split guard: a
// bucket carrying cache_creation_unknown (this store's encoding for "the
// harvester recorded cache-creation usage but not which rate applied",
// internal/harvest.Bucket's own doc comment) must never be priced by
// guessing a rate -- the whole model bucket is treated exactly like one
// with no pricing row at all: unpriceable, and the top-level cost_usd is
// omitted.
func TestPriceUnknownCacheSplitIsUnpriceable(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-unknownsplit-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"claude-opus-5": {
				"tokens": {"main": {"input": 1000000, "cache_creation": 500000, "cache_creation_unknown": 500000}}
			}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-opus-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 5, OutputPerMTok: 25,
		CacheWritePerMTok: 6.25, CacheWrite5mPerMTok: 6.25, CacheWrite1hPerMTok: ptr(10.0),
		CacheReadPerMTok: 0.50,
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	err = st.Price(ctx, run.ID)
	if !errors.Is(err, store.ErrPricingNotFound) {
		t.Fatalf("Price(unknown cache split) error = %v, want errors.Is(_, store.ErrPricingNotFound)", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag map[string]json.RawMessage
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}
	if _, ok := bag["cost_usd"]; ok {
		t.Errorf("cost_usd = %s, want no top-level cost_usd: this run's only model bucket has an unpriceable cache split", bag["cost_usd"])
	}
	var models map[string]json.RawMessage
	if err := json.Unmarshal(bag["models"], &models); err != nil {
		t.Fatalf("unmarshal models: %v", err)
	}
	var opus struct {
		CostUSD *float64 `json:"cost_usd"`
	}
	if err := json.Unmarshal(models["claude-opus-5"], &opus); err != nil {
		t.Fatalf("unmarshal models.claude-opus-5: %v", err)
	}
	if opus.CostUSD != nil {
		t.Errorf("models.claude-opus-5.cost_usd = %v, want absent -- guessing which rate applied to the unknown-split portion is exactly what this rule forbids", *opus.CostUSD)
	}
}

// TestPriceFastModeUsesFastRate covers task 23's third defect: a run
// recorded at "fast" speed must be charged the model's fast input/output
// rate, not the standard one.
func TestPriceFastModeUsesFastRate(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-fast-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"speed": "fast",
		"models": {
			"claude-opus-5": {"tokens": {"main": {"input": 1000000, "output": 1000000}}}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-opus-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 5, OutputPerMTok: 25,
		CacheWritePerMTok: 6.25, CacheWrite5mPerMTok: 6.25, CacheWrite1hPerMTok: ptr(10.0),
		CacheReadPerMTok:  0.50,
		FastInputPerMTok:  ptr(10.0),
		FastOutputPerMTok: ptr(50.0),
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	if err := st.Price(ctx, run.ID); err != nil {
		t.Fatalf("Price: %v", err)
	}

	priced, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag struct {
		CostUSD float64 `json:"cost_usd"`
	}
	if err := json.Unmarshal(priced.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	wantCost := 1*10.0 + 1*50.0 // fast rates, not the standard 5/25.
	if diff := bag.CostUSD - wantCost; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("cost_usd = %v, want %v (fast-mode input/output rate)", bag.CostUSD, wantCost)
	}
}

// TestPriceFastModeWithoutFastRateIsUnpriceable covers the other half:
// a run recorded at "fast" speed against a model with no published
// fast-mode rate (Sonnet 5 and Haiku 4.5, per pricing_seed.go's own
// table) must not be silently priced at the standard rate.
func TestPriceFastModeWithoutFastRateIsUnpriceable(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-fastnorate-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"speed": "fast",
		"models": {
			"claude-sonnet-5": {"tokens": {"main": {"input": 1000000}}}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-sonnet-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 2, OutputPerMTok: 10,
		CacheWritePerMTok: 2.5, CacheWrite5mPerMTok: 2.5, CacheWrite1hPerMTok: ptr(4.0),
		CacheReadPerMTok: 0.20,
		// No fast rate published for this model -- FastInputPerMTok and
		// FastOutputPerMTok stay nil.
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	err = st.Price(ctx, run.ID)
	if !errors.Is(err, store.ErrPricingNotFound) {
		t.Fatalf("Price(fast speed, no fast rate) error = %v, want errors.Is(_, store.ErrPricingNotFound)", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag map[string]json.RawMessage
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}
	if _, ok := bag["cost_usd"]; ok {
		t.Errorf("cost_usd = %s, want no top-level cost_usd: claude-sonnet-5 has no fast-mode rate", bag["cost_usd"])
	}
}

// TestPriceDispatchGetsCostThroughSamePricingPath covers KAN-201's own
// defect fix: a dispatch's cost must be computed through the exact same
// rate resolution and chargeableTokens.cost arithmetic as its owning
// model bucket -- not an implied average rate scaled from the model
// bucket's blended total, which overstates a cache-heavy dispatch and
// understates an output-heavy one (specs/myflow-stats-views/spec.md, "Per-
// dispatch cost SHALL be derived through the same pricing path"). This
// also pins the double-counting guard: the top-level cost_usd is the
// model buckets' own sum, unaffected by dispatch pricing, even though a
// dispatch's tokens are already inside that same model bucket.
func TestPriceDispatchGetsCostThroughSamePricingPath(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-dispatch-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "5. The review panel"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"claude-sonnet-5": {"tokens": {"sidechain": {"input": 1000000, "output": 200000}}}
		},
		"dispatches": {
			"agent-1": {
				"tokens": {"sidechain": {"input": 800000, "output": 160000}},
				"model": "claude-sonnet-5"
			}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-sonnet-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 1, OutputPerMTok: 5, CacheWritePerMTok: 1.25, CacheReadPerMTok: 0.1,
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	if err := st.Price(ctx, run.ID); err != nil {
		t.Fatalf("Price: %v", err)
	}

	priced, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag struct {
		CostUSD float64 `json:"cost_usd"`
		Models  map[string]struct {
			CostUSD float64 `json:"cost_usd"`
		} `json:"models"`
		Dispatches map[string]struct {
			CostUSD float64 `json:"cost_usd"`
		} `json:"dispatches"`
	}
	if err := json.Unmarshal(priced.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	const eps = 1e-9
	wantModelCost := 1.0*1 + 0.2*5     // the model bucket's own, blended total (1M input, 200K output)
	wantDispatchCost := 0.8*1 + 0.16*5 // priced through the same rate, on the dispatch's own tokens alone (800K input, 160K output)
	if diff := bag.Models["claude-sonnet-5"].CostUSD - wantModelCost; diff > eps || diff < -eps {
		t.Errorf("models.claude-sonnet-5.cost_usd = %v, want %v", bag.Models["claude-sonnet-5"].CostUSD, wantModelCost)
	}
	dispatch, ok := bag.Dispatches["agent-1"]
	if !ok {
		t.Fatalf("dispatches.agent-1 missing from priced metrics: %s", priced.Metrics)
	}
	if diff := dispatch.CostUSD - wantDispatchCost; diff > eps || diff < -eps {
		t.Errorf("dispatches.agent-1.cost_usd = %v, want %v", dispatch.CostUSD, wantDispatchCost)
	}
	// Double-counting guard: the top-level total is the model buckets' own
	// sum, unaffected by having also priced the dispatch view of the same
	// tokens -- pricing both must not roughly double the run's total.
	if diff := bag.CostUSD - wantModelCost; diff > eps || diff < -eps {
		t.Errorf("cost_usd = %v, want %v (the model bucket's own cost; dispatch cost must not be added on top)", bag.CostUSD, wantModelCost)
	}
}

// TestPriceDispatchIsPricedWhenEveryModelBucketIsUnpriceable pins F28
// (pass 7 of this change's own review panel): before the fix, Price
// returned as soon as it found the "models" bag entirely unpriceable
// (priced == 0), before the dispatches loop below it ever ran -- so a
// dispatch whose own sidecar-declared model DID have a valid pricing row
// still got no cost_usd, purely because some other, unrelated model
// happened to head this run's "models" bag with no rate at all. That
// defeats per-dispatch attribution's whole purpose exactly when the two
// model sources diverge, which is the case they exist to expose. This
// pins the dispatch getting priced regardless, while every other rule
// stays exactly as documented: ErrPricingNotFound is still returned,
// naming the unpriceable model; no top-level cost_usd is written; and the
// dispatch's cost is never folded into it.
func TestPriceDispatchIsPricedWhenEveryModelBucketIsUnpriceable(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-dispatch-priced-when-models-unpriceable-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "5. The review panel"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"unpriced-model": {"tokens": {"main": {"input": 1000000}}}
		},
		"dispatches": {
			"agent-1": {
				"tokens": {"sidechain": {"input": 800000, "output": 160000}},
				"model": "claude-sonnet-5"
			}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-sonnet-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 1, OutputPerMTok: 5, CacheWritePerMTok: 1.25, CacheReadPerMTok: 0.1,
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}
	// Deliberately no pricing row for "unpriced-model" -- the run's own
	// "models" bag is entirely unpriceable, while the dispatch's own
	// recorded model (claude-sonnet-5) has a valid rate.

	err = st.Price(ctx, run.ID)
	if !errors.Is(err, store.ErrPricingNotFound) {
		t.Fatalf("Price() error = %v, want errors.Is(_, store.ErrPricingNotFound): the models bag's only model has no rate", err)
	}

	priced, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag struct {
		CostUSD    *float64 `json:"cost_usd"`
		Dispatches map[string]struct {
			CostUSD float64 `json:"cost_usd"`
		} `json:"dispatches"`
	}
	if err := json.Unmarshal(priced.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if bag.CostUSD != nil {
		t.Errorf("cost_usd = %v, want absent: the only model bucket is unpriceable", *bag.CostUSD)
	}
	dispatch, ok := bag.Dispatches["agent-1"]
	if !ok {
		t.Fatalf("dispatches.agent-1 missing from priced metrics: %s -- its own sidecar-declared model has a valid pricing row and must be priced regardless of whether the models bag priced anything at all", priced.Metrics)
	}
	const eps = 1e-9
	wantDispatchCost := 0.8*1 + 0.16*5
	if diff := dispatch.CostUSD - wantDispatchCost; diff > eps || diff < -eps {
		t.Errorf("dispatches.agent-1.cost_usd = %v, want %v", dispatch.CostUSD, wantDispatchCost)
	}
}

// TestPriceDegradesOnMalformedDispatchesRatherThanFailing pins F6 (pass 1
// of this change's own review panel): a "dispatches" value that does not
// even decode as map[string]dispatchBucket must not abort the whole
// Price call -- the models loop's already-computed modelsPatch and
// top-level cost_usd must still be written. Before the fix, the decode
// error returned immediately, before patchFields was even built,
// discarding everything already priced. No current writer produces this
// shape (encodePatches, harvest/watcher.go, always emits a well-formed
// object) -- this is latent robustness, exercised here by writing the
// malformed shape directly through MergeMetrics.
func TestPriceDegradesOnMalformedDispatchesRatherThanFailing(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-malformed-dispatches-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "5. The review panel"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"claude-sonnet-5": {"tokens": {"sidechain": {"input": 1000000, "output": 200000}}}
		},
		"dispatches": "not-an-object"
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-sonnet-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 1, OutputPerMTok: 5, CacheWritePerMTok: 1.25, CacheReadPerMTok: 0.1,
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	if err := st.Price(ctx, run.ID); err != nil {
		t.Fatalf("Price: %v, want nil -- a malformed dispatches value must degrade, not fail the whole call", err)
	}

	priced, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag struct {
		CostUSD float64 `json:"cost_usd"`
		Models  map[string]struct {
			CostUSD float64 `json:"cost_usd"`
		} `json:"models"`
	}
	if err := json.Unmarshal(priced.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	const eps = 1e-9
	wantModelCost := 1.0*1 + 0.2*5
	if diff := bag.Models["claude-sonnet-5"].CostUSD - wantModelCost; diff > eps || diff < -eps {
		t.Errorf("models.claude-sonnet-5.cost_usd = %v, want %v -- must still be written despite the malformed dispatches value", bag.Models["claude-sonnet-5"].CostUSD, wantModelCost)
	}
	if diff := bag.CostUSD - wantModelCost; diff > eps || diff < -eps {
		t.Errorf("cost_usd = %v, want %v -- the already-computed total must not be discarded", bag.CostUSD, wantModelCost)
	}
}

// TestPriceDispatchLookupErrorDoesNotDiscardModelsResult pins F13 (pass 2
// of this change's own review panel): a non-ErrPricingNotFound error from
// resolveRate in the *dispatches* loop -- a genuine store failure, not
// "no rate in effect" -- must not discard the models loop's
// already-computed modelsPatch and total, the same "worse than writing
// nothing" mistake F6's fix (the test above) prevents for a malformed
// dispatches value, reached here through a different trigger. Before the
// fix, this loop's `return err` fired before patchFields was even built,
// throwing away everything the models loop above had already priced.
//
// A JSON string small enough to satisfy jsonb (unlike an embedded NUL
// byte, which Postgres' jsonb input rejects outright regardless of
// query, so it can never even reach MergeMetrics) cannot itself make a
// later `model = $1` lookup fail -- any string that survives jsonb
// storage round-trips as valid UTF-8, and text equality never errors on
// that. The seam that *is* reachable is the pricing row itself: this
// test seeds one, by raw SQL rather than through PutPricing (whose
// PricingRate.InputPerMTok is a Go float64 and so cannot even represent
// this value), whose input_per_mtok is a genuine NUMERIC Postgres
// accepts on INSERT -- 1e309 -- but pgx's numeric-to-float64 scan
// rejects on SELECT with strconv.ParseFloat's own "value out of range",
// a real, deterministic, non-ErrNoRows error. A pricing rate with a
// misplaced decimal point is exactly the kind of bad row a real seed
// mistake could produce.
func TestPriceDispatchLookupErrorDoesNotDiscardModelsResult(t *testing.T) {
	dsn := newTestDatabase(t)
	ctx := context.Background()

	st, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	if err := st.RunMigrations(ctx); err != nil {
		st.Close()
		t.Fatalf("run migrations: %v", err)
	}
	t.Cleanup(st.Close)

	rawPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("open raw pool for direct seeding: %v", err)
	}
	t.Cleanup(rawPool.Close)

	projectKey := fmt.Sprintf("proj-price-dispatch-lookup-error-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "5. The review panel"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"claude-sonnet-5": {"tokens": {"sidechain": {"input": 1000000, "output": 200000}}}
		},
		"dispatches": {
			"agent-1": {
				"tokens": {"sidechain": {"input": 800000, "output": 160000}},
				"model": "overflow-model"
			}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-sonnet-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 1, OutputPerMTok: 5, CacheWritePerMTok: 1.25, CacheReadPerMTok: 0.1,
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}
	if _, err := rawPool.Exec(ctx, `
		INSERT INTO pricing (model, effective_from, input_per_mtok, output_per_mtok, cache_write_per_mtok, cache_write_5m_per_mtok, cache_read_per_mtok)
		VALUES ('overflow-model', $1, '1e309', 1, 1, 1, 1)
	`, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("seed overflow pricing row: %v", err)
	}

	if err := st.Price(ctx, run.ID); err != nil {
		t.Fatalf("Price: %v, want nil -- a dispatch lookup error must degrade, not fail the whole call", err)
	}

	priced, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag struct {
		CostUSD float64 `json:"cost_usd"`
		Models  map[string]struct {
			CostUSD float64 `json:"cost_usd"`
		} `json:"models"`
		Dispatches map[string]struct {
			CostUSD *float64 `json:"cost_usd"`
		} `json:"dispatches"`
	}
	if err := json.Unmarshal(priced.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	const eps = 1e-9
	wantModelCost := 1.0*1 + 0.2*5
	if diff := bag.Models["claude-sonnet-5"].CostUSD - wantModelCost; diff > eps || diff < -eps {
		t.Errorf("models.claude-sonnet-5.cost_usd = %v, want %v -- must still be written despite the dispatch lookup error", bag.Models["claude-sonnet-5"].CostUSD, wantModelCost)
	}
	if diff := bag.CostUSD - wantModelCost; diff > eps || diff < -eps {
		t.Errorf("cost_usd = %v, want %v -- the already-computed total must not be discarded", bag.CostUSD, wantModelCost)
	}
	if dispatch, ok := bag.Dispatches["agent-1"]; ok && dispatch.CostUSD != nil {
		t.Errorf("dispatches.agent-1.cost_usd = %v, want absent -- its own model lookup failed", *dispatch.CostUSD)
	}
}

// TestPriceDispatchWithNoRecordedModelGetsNoCost covers the sidecar-absent
// case: a dispatch whose meta sidecar was never found records no model at
// all (DispatchBucket's own doc comment, attribute.go), and Price must not
// price it -- absence is never a value, never a fabricated 0 and never a
// neighbouring dispatch's rate.
func TestPriceDispatchWithNoRecordedModelGetsNoCost(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-dispatch-nomodel-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "5. The review panel"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"claude-sonnet-5": {"tokens": {"sidechain": {"input": 1000000}}}
		},
		"dispatches": {
			"agent-2": {"tokens": {"sidechain": {"input": 200000}}}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-sonnet-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 1, OutputPerMTok: 5, CacheWritePerMTok: 1.25, CacheReadPerMTok: 0.1,
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	if err := st.Price(ctx, run.ID); err != nil {
		t.Fatalf("Price: %v", err)
	}

	priced, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag map[string]json.RawMessage
	if err := json.Unmarshal(priced.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	var dispatches map[string]json.RawMessage
	if err := json.Unmarshal(bag["dispatches"], &dispatches); err != nil {
		t.Fatalf("unmarshal dispatches: %v", err)
	}
	var agent2 struct {
		CostUSD *float64 `json:"cost_usd"`
	}
	if err := json.Unmarshal(dispatches["agent-2"], &agent2); err != nil {
		t.Fatalf("unmarshal dispatches.agent-2: %v", err)
	}
	if agent2.CostUSD != nil {
		t.Errorf("dispatches.agent-2.cost_usd = %v, want absent -- this dispatch recorded no model at all", *agent2.CostUSD)
	}
}

// TestPriceDispatchModelWithNoPricingRowGetsNoCost covers a dispatch whose
// own recorded model has no pricing row in effect: it gets no cost_usd,
// not 0 and not a neighbouring model's rate -- and this must not stop
// Price from succeeding overall when every model bucket it is actually
// responsible for still prices cleanly.
func TestPriceDispatchModelWithNoPricingRowGetsNoCost(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-dispatch-norate-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "5. The review panel"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"claude-sonnet-5": {"tokens": {"sidechain": {"input": 1000000}}}
		},
		"dispatches": {
			"agent-3": {"tokens": {"sidechain": {"input": 200000}}, "model": "no-such-model"}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-sonnet-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 1, OutputPerMTok: 5, CacheWritePerMTok: 1.25, CacheReadPerMTok: 0.1,
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	if err := st.Price(ctx, run.ID); err != nil {
		t.Fatalf("Price: %v, want success -- the run's own model bucket priced fine, a dispatch's unpriceable model must not fail the whole call", err)
	}

	priced, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag map[string]json.RawMessage
	if err := json.Unmarshal(priced.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	var dispatches map[string]json.RawMessage
	if err := json.Unmarshal(bag["dispatches"], &dispatches); err != nil {
		t.Fatalf("unmarshal dispatches: %v", err)
	}
	var agent3 struct {
		CostUSD *float64 `json:"cost_usd"`
	}
	if err := json.Unmarshal(dispatches["agent-3"], &agent3); err != nil {
		t.Fatalf("unmarshal dispatches.agent-3: %v", err)
	}
	if agent3.CostUSD != nil {
		t.Errorf("dispatches.agent-3.cost_usd = %v, want absent -- no pricing row was ever in effect for %q", *agent3.CostUSD, "no-such-model")
	}
}

// TestPriceDispatchWithUnknownCacheSplitGetsNoCost covers the refusal
// case at dispatch granularity: a dispatch bucket carrying
// cache_creation_unknown must be refused, not partially priced --
// chargeableTokens.cost's own doc comment states this rule, and it
// applies identically whether the bucket being priced is a model's or a
// dispatch's.
func TestPriceDispatchWithUnknownCacheSplitGetsNoCost(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-dispatch-unknownsplit-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "5. The review panel"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"claude-sonnet-5": {"tokens": {"sidechain": {"input": 1000000}}}
		},
		"dispatches": {
			"agent-4": {
				"tokens": {"sidechain": {"cache_creation_unknown": 500000}},
				"model": "claude-sonnet-5"
			}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-sonnet-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 1, OutputPerMTok: 5, CacheWritePerMTok: 1.25, CacheReadPerMTok: 0.1,
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	if err := st.Price(ctx, run.ID); err != nil {
		t.Fatalf("Price: %v", err)
	}

	priced, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag map[string]json.RawMessage
	if err := json.Unmarshal(priced.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	var dispatches map[string]json.RawMessage
	if err := json.Unmarshal(bag["dispatches"], &dispatches); err != nil {
		t.Fatalf("unmarshal dispatches: %v", err)
	}
	var agent4 struct {
		CostUSD *float64 `json:"cost_usd"`
	}
	if err := json.Unmarshal(dispatches["agent-4"], &agent4); err != nil {
		t.Fatalf("unmarshal dispatches.agent-4: %v", err)
	}
	if agent4.CostUSD != nil {
		t.Errorf("dispatches.agent-4.cost_usd = %v, want absent -- its cache-creation split is unknown", *agent4.CostUSD)
	}
}

// TestStageRunsNullRepoRootSkipsFKCheck verifies the design's claim, tagged
// unverified in design.md, that a NULL repo_root inserts cleanly with no
// matching change_repos row at all -- because Postgres' default FK match
// semantics (MATCH SIMPLE) skip the check entirely whenever any column of
// a composite foreign key is NULL.
func TestStageRunsNullRepoRootSkipsFKCheck(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-nullrepo-%d", time.Now().UnixNano())
	// No repos on this change at all.
	c := baseChange(projectKey, "kan-1")
	c.Repos = nil
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("PutChange: %v", err)
	}

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	in.RepoRoot = nil
	if _, err := st.BeginStage(ctx, in); err != nil {
		t.Fatalf("BeginStage with nil RepoRoot and no change_repos rows: %v", err)
	}
}

// TestStageRunsRepoRootMustMatchChangeRepos verifies the other half of the
// same unverified DDL claim: a non-NULL repo_root that names a repository
// the change does not have is rejected by the foreign key, rather than
// silently accepted.
func TestStageRunsRepoRootMustMatchChangeRepos(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-badrepo-%d", time.Now().UnixNano())
	c := baseChange(projectKey, "kan-1")
	c.Repos = []store.Repo{{RepoRoot: "/repo/a"}}
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("PutChange: %v", err)
	}

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	in.RepoRoot = ptr("/repo/does-not-exist")
	_, err := st.BeginStage(ctx, in)
	if err == nil {
		t.Fatalf("BeginStage with a repo_root absent from change_repos succeeded, want a foreign-key rejection")
	}
}

// TestSupersedeIndexExists pins that migration 0009 is embedded, applied,
// and names its index what this change agreed on -- the part a later edit
// could silently drop. An index is a performance property and a query plan
// is not stable enough to assert on (EXPLAIN output varies with planner
// statistics and row counts), so this checks pg_indexes directly rather
// than the supersede UPDATE's plan.
//
// Matching the index name alone (fix round 5, finding F18) would still
// pass if the index were rebuilt under the same name over a different
// column, or without the `WHERE ended_at IS NULL` predicate that is this
// migration's whole point -- indexdef carries both, so this asserts the
// definition, not just that something by this name exists.
//
// *store.Store exposes no raw-query escape hatch (by design -- see
// store.go's package doc, "the only package in this module that builds
// SQL"), so this opens its own connection to the same freshly-migrated
// database, the same shape TestRepeatableReadPreventsCountDrift
// (query_test.go) already uses for a query the typed API does not serve.
func TestSupersedeIndexExists(t *testing.T) {
	dsn := newTestDatabase(t)
	ctx := context.Background()

	st, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()
	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	var indexdef string
	err = pool.QueryRow(ctx,
		"SELECT indexdef FROM pg_indexes WHERE indexname = $1", "stage_runs_open_session_token",
	).Scan(&indexdef)
	if err != nil {
		t.Fatalf("query pg_indexes: %v", err)
	}
	if !strings.Contains(indexdef, "(session_token)") {
		t.Errorf("indexdef = %q, want it to index (session_token)", indexdef)
	}
	if !strings.Contains(indexdef, "WHERE (ended_at IS NULL)") {
		t.Errorf("indexdef = %q, want the partial predicate WHERE (ended_at IS NULL)", indexdef)
	}
}
