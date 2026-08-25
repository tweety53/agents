package reconcile_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/reconcile"
	"github.com/tweety53/agents/stats/internal/store"
)

// nopStageStore satisfies api.StageStore with nothing but errors -- for
// the reconcile_test.go tests above that exercise the state journal only
// and never touch a stage mark, now that reconcile.New requires a
// StageStore alongside a ChangeStore. Mirrors
// internal/client/client_test.go's stubStageStore, which exists for the
// identical reason on the client side.
type nopStageStore struct{}

var errStageStoreNotExercised = errors.New("nopStageStore: stage marks are not exercised by this test")

func (nopStageStore) BeginStage(context.Context, store.BeginStageInput) (store.StageRun, error) {
	return store.StageRun{}, errStageStoreNotExercised
}

func (nopStageStore) EndStage(context.Context, int64, time.Time, string) error {
	return errStageStoreNotExercised
}

func (nopStageStore) MergeMetrics(context.Context, int64, json.RawMessage) error {
	return errStageStoreNotExercised
}

func (nopStageStore) QueryStageRuns(context.Context, store.Query) ([]store.StageRun, int, error) {
	return nil, 0, errStageStoreNotExercised
}

func (nopStageStore) PutChange(context.Context, store.Change) error {
	return errStageStoreNotExercised
}

var _ api.StageStore = nopStageStore{}

// stageJournalPath mirrors cmd/myflow/stage.go's own stageJournalPath
// (the state journal path with ".stage" appended) -- reproduced here
// rather than imported, since stage.go lives in a main package and
// nothing may import one. See reconcile.go's stageMarkJournalBody doc
// comment for the boundary this crosses, and why it is a literal
// convention rather than a shared constant.
func stageJournalPath(root, project, name string) string {
	return journalPath(root, project, name) + ".stage"
}

// stageMarkEnvelope is cmd/myflow/stage.go's own stageMarkJournalBody
// shape, reproduced here for the same reason stageJournalPath is: no
// import across the main-package boundary.
type stageMarkEnvelope struct {
	Kind    string `json:"kind"`
	Request any    `json:"request"`
}

// appendStageMark journals kind/req exactly as cmd/myflow/stage.go's
// journalStageMark does -- the same envelope, the same
// fallback.AppendJournalEntry call, at the same suffixed path -- so a test
// using this exercises reconcile's real decoder against the real shape the
// CLI writes, not a hand-simplified stand-in for it. req is a
// client.BeginStageRequest or client.EndStageRequest (the same exported
// types stage.go itself marshals), which is the actual shared contract
// between the two sides; only the envelope and the path suffix are
// duplicated conventions.
func appendStageMark(t *testing.T, root, project, name, kind string, req any) {
	t.Helper()
	body, err := json.Marshal(stageMarkEnvelope{Kind: kind, Request: req})
	if err != nil {
		t.Fatalf("marshal stage mark journal body: %v", err)
	}
	if err := fallback.AppendJournalEntry(stageJournalPath(root, project, name), project, name, body, time.Now()); err != nil {
		t.Fatalf("append stage mark journal entry: %v", err)
	}
}

func pendingStageCount(t *testing.T, root, project, name string) int {
	t.Helper()
	entries, err := fallback.ReadJournalEntries(stageJournalPath(root, project, name))
	if err != nil {
		t.Fatalf("read stage journal entries: %v", err)
	}
	return len(entries)
}

// seedChange writes a minimal STARTED change directly to st, so a stage
// mark's BeginStage lookup finds it rather than exercising the synthetic-
// change bootstrap path (internal/api's own tests already cover that
// path; these tests are about replay mechanics).
func seedChange(t *testing.T, st *store.Store, project, name string) {
	t.Helper()
	if err := st.PutChange(context.Background(), store.Change{
		ProjectKey:       project,
		MainCheckoutPath: "/tmp/reconcile-stage-test",
		Name:             name,
		State:            store.StateStarted,
		UpdatedAt:        time.Date(2026, 8, 13, 9, 0, 0, 0, time.UTC),
		UpdatedBy:        "tester",
	}); err != nil {
		t.Fatalf("seed change: %v", err)
	}
}

// openStageRun returns the one open (no end mark) stage run for
// project/name/command/stage, or nil. Test-local rather than reaching
// into internal/api's own unexported findOpenStageRun, so this stays a
// black-box assertion against what QueryStageRuns actually returns.
func openStageRun(t *testing.T, st *store.Store, project, name, command, stage string) *store.StageRun {
	t.Helper()
	runs, _, err := st.QueryStageRuns(context.Background(), store.Query{
		Filters: []store.Filter{
			{Field: "project", Op: store.OpEq, Value: project},
			{Field: "name", Op: store.OpEq, Value: name},
			{Field: "command", Op: store.OpEq, Value: command},
			{Field: "stage", Op: store.OpEq, Value: stage},
			{Field: "ended_at", Op: store.OpNull},
		},
	})
	if err != nil {
		t.Fatalf("QueryStageRuns: %v", err)
	}
	if len(runs) == 0 {
		return nil
	}
	return &runs[0]
}

func allStageRuns(t *testing.T, st *store.Store, project, name string) []store.StageRun {
	t.Helper()
	runs, _, err := st.QueryStageRuns(context.Background(), store.Query{
		Filters: []store.Filter{
			{Field: "project", Op: store.OpEq, Value: project},
			{Field: "name", Op: store.OpEq, Value: name},
		},
		Sort: []store.SortKey{{Field: "attempt"}},
	})
	if err != nil {
		t.Fatalf("QueryStageRuns: %v", err)
	}
	return runs
}

// --- a mark journalled while the store is unreachable is observable in
// the store after a replay ---

// TestReplayAppliesPendingStageMarks is the coordinator's own scenario:
// a begin and an end mark, journalled exactly as cmd/myflow/stage.go's
// fallback would have written them while the store was unreachable, must
// both be applied -- and retired -- once the store is reachable and Run
// replays them. Before this task's fix, nothing walked "*.journal.stage"
// at all, so this stage run would never have existed in the store; that
// is the defect this test pins shut.
func TestReplayAppliesPendingStageMarks(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	root := t.TempDir()

	seedChange(t, st, "proj-stage-replay", "chg-1")

	begin := client.BeginStageRequest{
		ProjectKey:   "proj-stage-replay",
		ChangeName:   "chg-1",
		Harness:      "claude-code",
		SessionToken: "mf-session-token-stage-replay",
		Command:      "/flow",
		Stage:        "flow.sdd-tdd",
		StartedAt:    time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC),
	}
	end := client.EndStageRequest{
		ProjectKey: "proj-stage-replay",
		ChangeName: "chg-1",
		Command:    "/flow",
		Stage:      "flow.sdd-tdd",
		EndedAt:    time.Date(2026, 8, 13, 10, 5, 0, 0, time.UTC),
		Outcome:    "completed",
		Metrics:    json.RawMessage(`{"fix_rounds":1}`),
	}
	appendStageMark(t, root, "proj-stage-replay", "chg-1", "begin", begin)
	appendStageMark(t, root, "proj-stage-replay", "chg-1", "end", end)

	rec := reconcile.New(st, st, st, root, nil)
	result, err := rec.Run(ctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Applied != 2 || result.Refused != 0 {
		t.Fatalf("Run result = %+v, want {Applied:2 Refused:0}", result)
	}

	runs := allStageRuns(t, st, "proj-stage-replay", "chg-1")
	if len(runs) != 1 {
		t.Fatalf("len(stage runs) = %d, want 1", len(runs))
	}
	run := runs[0]
	if run.EndedAt == nil {
		t.Fatal("EndedAt is nil, want the replayed end mark's instant")
	}
	if run.Outcome == nil || *run.Outcome != "completed" {
		t.Fatalf("Outcome = %v, want completed", run.Outcome)
	}
	var metrics map[string]any
	if err := json.Unmarshal(run.Metrics, &metrics); err != nil {
		t.Fatalf("decode metrics: %v", err)
	}
	if metrics["fix_rounds"] != float64(1) {
		t.Errorf("metrics.fix_rounds = %v, want 1", metrics["fix_rounds"])
	}

	if n := pendingStageCount(t, root, "proj-stage-replay", "chg-1"); n != 0 {
		t.Fatalf("pending stage journal entries after replay = %d, want 0", n)
	}
}

// TestStageEndReplayWithNoOpenRunRetiresCleanly is the coordinator's
// closed-run scenario: an end mark for a run that no longer has an open
// attempt (here, because no begin mark for that triple was ever applied --
// the same observable state a sweeper-closed run leaves behind) must be
// retired, not retried forever and not silently treated as success.
func TestStageEndReplayWithNoOpenRunRetiresCleanly(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	root := t.TempDir()

	seedChange(t, st, "proj-stage-noopen", "chg-1")

	end := client.EndStageRequest{
		ProjectKey: "proj-stage-noopen",
		ChangeName: "chg-1",
		Command:    "/flow",
		Stage:      "flow.sdd-tdd",
		EndedAt:    time.Date(2026, 8, 13, 10, 5, 0, 0, time.UTC),
		Outcome:    "completed",
	}
	appendStageMark(t, root, "proj-stage-noopen", "chg-1", "end", end)

	rec := reconcile.New(st, st, st, root, nil)
	result, err := rec.Run(ctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Applied != 0 || result.Refused != 1 {
		t.Fatalf("Run result = %+v, want {Applied:0 Refused:1} -- a no-open-run end mark is a definitive refusal, not success and not left pending", result)
	}
	if n := pendingStageCount(t, root, "proj-stage-noopen", "chg-1"); n != 0 {
		t.Fatalf("pending stage journal entries after replay = %d, want 0 -- an unresolvable end mark must still be retired, never retried forever", n)
	}
	if len(allStageRuns(t, st, "proj-stage-noopen", "chg-1")) != 0 {
		t.Error("a stage run was created for an end mark with nothing to end -- ApplyEndStageMark must never fabricate one")
	}
}

// TestUndocumentedStageMarkReplayRetiresAsRefused pins the same
// closed-form refusal for a begin mark naming a stage README.md no longer
// documents by the time replay runs (or never did, for a journal written
// by a stale binary) -- api.IsDefinitiveMarkOutcome's *stages.ErrUnknownStage
// branch is what keeps this from stalling every entry behind it forever.
func TestUndocumentedStageMarkReplayRetiresAsRefused(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	root := t.TempDir()

	seedChange(t, st, "proj-stage-undoc", "chg-1")

	begin := client.BeginStageRequest{
		ProjectKey: "proj-stage-undoc",
		ChangeName: "chg-1",
		Harness:    "claude-code",
		Command:    "/flow",
		Stage:      "a stage nobody documented",
		StartedAt:  time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC),
	}
	appendStageMark(t, root, "proj-stage-undoc", "chg-1", "begin", begin)

	rec := reconcile.New(st, st, st, root, nil)
	result, err := rec.Run(ctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Applied != 0 || result.Refused != 1 {
		t.Fatalf("Run result = %+v, want {Applied:0 Refused:1}", result)
	}
	if n := pendingStageCount(t, root, "proj-stage-undoc", "chg-1"); n != 0 {
		t.Fatalf("pending stage journal entries after replay = %d, want 0", n)
	}
	if len(allStageRuns(t, st, "proj-stage-undoc", "chg-1")) != 0 {
		t.Error("a stage run was recorded for an undocumented stage name")
	}
}

// --- interrupted stage replay resumes rather than loses or duplicates ---

// injectingStageStore wraps a real api.StageStore, forwarding every call
// except BeginStage for one chosen change name, which fails
// failsRemaining times with a plain, non-definitive error -- the stage
// journal's equivalent of reconcile_test.go's injectingStore for the state
// journal.
type injectingStageStore struct {
	inner api.StageStore

	mu             sync.Mutex
	failChangeName string
	failsRemaining int
	beginCalls     int
}

var errInjectedStageTransportFailure = errors.New("injected stage transport failure")

func (s *injectingStageStore) BeginStage(ctx context.Context, in store.BeginStageInput) (store.StageRun, error) {
	s.mu.Lock()
	s.beginCalls++
	shouldFail := in.ChangeName == s.failChangeName && s.failsRemaining > 0
	if shouldFail {
		s.failsRemaining--
	}
	s.mu.Unlock()

	if shouldFail {
		return store.StageRun{}, errInjectedStageTransportFailure
	}
	return s.inner.BeginStage(ctx, in)
}

func (s *injectingStageStore) EndStage(ctx context.Context, id int64, endedAt time.Time, outcome string) error {
	return s.inner.EndStage(ctx, id, endedAt, outcome)
}

func (s *injectingStageStore) MergeMetrics(ctx context.Context, id int64, patch json.RawMessage) error {
	return s.inner.MergeMetrics(ctx, id, patch)
}

func (s *injectingStageStore) QueryStageRuns(ctx context.Context, q store.Query) ([]store.StageRun, int, error) {
	return s.inner.QueryStageRuns(ctx, q)
}

func (s *injectingStageStore) PutChange(ctx context.Context, c store.Change) error {
	return s.inner.PutChange(ctx, c)
}

var _ api.StageStore = (*injectingStageStore)(nil)

// TestStageReplayInterruptedResumesWithoutDuplicating exercises the stage
// journal's version of TestInterruptedReplayResumesWithoutDuplicating: a
// transport-shaped failure on the second of two pending begin marks must
// leave that entry (and only that entry) in the journal, and a second Run
// -- once the failure clears -- must apply it, without ever re-sending the
// first (already-applied) entry to the store.
func TestStageReplayInterruptedResumesWithoutDuplicating(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	root := t.TempDir()

	seedChange(t, st, "proj-stage-interrupt", "chg-a")
	seedChange(t, st, "proj-stage-interrupt", "chg-b")

	appendStageMark(t, root, "proj-stage-interrupt", "chg-a", "begin", client.BeginStageRequest{
		ProjectKey: "proj-stage-interrupt", ChangeName: "chg-a", Harness: "claude-code",
		SessionToken: "mf-session-token-interrupt-chg-a",
		Command:      "/flow", Stage: "flow.sdd-tdd", StartedAt: time.Now(),
	})
	appendStageMark(t, root, "proj-stage-interrupt", "chg-b", "begin", client.BeginStageRequest{
		ProjectKey: "proj-stage-interrupt", ChangeName: "chg-b", Harness: "claude-code",
		SessionToken: "mf-session-token-interrupt-chg-b",
		Command:      "/flow", Stage: "flow.sdd-tdd", StartedAt: time.Now(),
	})

	wrapped := &injectingStageStore{inner: st, failChangeName: "chg-b", failsRemaining: 1}
	rec := reconcile.New(st, wrapped, st, root, nil)

	result1, err := rec.Run(ctx)
	if err != nil {
		t.Fatalf("Run (interrupted): %v", err)
	}
	if result1.Applied != 1 {
		t.Fatalf("Run (interrupted) result = %+v, want Applied:1 (only chg-a)", result1)
	}
	if n := pendingStageCount(t, root, "proj-stage-interrupt", "chg-a"); n != 0 {
		t.Fatalf("chg-a pending stage entries = %d, want 0 (accepted, must be retired)", n)
	}
	if n := pendingStageCount(t, root, "proj-stage-interrupt", "chg-b"); n != 1 {
		t.Fatalf("chg-b pending stage entries = %d, want 1 (unresolved, must remain)", n)
	}
	if openStageRun(t, st, "proj-stage-interrupt", "chg-a", "/flow", "flow.sdd-tdd") == nil {
		t.Error("chg-a has no open stage run after the first Run -- it should have been applied")
	}
	if openStageRun(t, st, "proj-stage-interrupt", "chg-b", "/flow", "flow.sdd-tdd") != nil {
		t.Error("chg-b has an open stage run after the first Run -- the injected failure should have prevented it")
	}

	result2, err := rec.Run(ctx)
	if err != nil {
		t.Fatalf("Run (resumed): %v", err)
	}
	if result2.Applied != 1 {
		t.Fatalf("Run (resumed) result = %+v, want Applied:1 (chg-b)", result2)
	}
	if n := pendingStageCount(t, root, "proj-stage-interrupt", "chg-b"); n != 0 {
		t.Fatalf("chg-b pending stage entries after resume = %d, want 0", n)
	}

	if len(allStageRuns(t, st, "proj-stage-interrupt", "chg-a")) != 1 {
		t.Error("chg-a was applied more than once across the two Run calls")
	}
	if len(allStageRuns(t, st, "proj-stage-interrupt", "chg-b")) != 1 {
		t.Error("chg-b should have exactly one stage run after resuming, not a duplicate from the failed first attempt")
	}
}
