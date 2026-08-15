package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

func ptr[T any](v T) *T { return &v }

// jsonEqual reports whether a and b encode the same JSON value. PostgreSQL's
// JSONB type reformats whitespace and key order on storage, so a byte-exact
// comparison of round-tripped JSON is the wrong check.
func jsonEqual(t *testing.T, a, b json.RawMessage) bool {
	t.Helper()
	var va, vb any
	if err := json.Unmarshal(a, &va); err != nil {
		t.Fatalf("unmarshal %s: %v", a, err)
	}
	if err := json.Unmarshal(b, &vb); err != nil {
		t.Fatalf("unmarshal %s: %v", b, err)
	}
	return reflect.DeepEqual(va, vb)
}

// baseChange returns a fully-populated Change for a fresh project/name pair,
// so every test starts from a record with every field set.
func baseChange(projectKey, name string) store.Change {
	return store.Change{
		ProjectKey:        projectKey,
		MainCheckoutPath:  "/Users/tweety53/Projects/" + projectKey,
		Name:              name,
		State:             store.StateStarted,
		Branch:            ptr("kan-16-myflow-stats-app"),
		Worktrees:         json.RawMessage(`{"do":"/tmp/wt-do"}`),
		ArtifactURL:       ptr("https://claude.ai/artifacts/abc"),
		JiraIssue:         ptr("KAN-16"),
		PlanningEffort:    ptr("standard"),
		Models:            json.RawMessage(`{"implementer":"opus","panel":"sonnet"}`),
		ReviewPanelRoster: ptr("light"),
		PRURL:             ptr("https://github.com/example/agents/pull/1"),
		UpdatedAt:         time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC),
		UpdatedBy:         "myflow-do",
	}
}

func TestPutChangeRoundTripsEveryField(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	want := baseChange("agents", "kan-16-myflow-stats-app")
	if err := st.PutChange(ctx, want); err != nil {
		t.Fatalf("PutChange: %v", err)
	}

	got, err := st.GetChange(ctx, want.ProjectKey, want.Name)
	if err != nil {
		t.Fatalf("GetChange: %v", err)
	}

	if got.ProjectKey != want.ProjectKey {
		t.Errorf("ProjectKey = %q, want %q", got.ProjectKey, want.ProjectKey)
	}
	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}
	if got.State != want.State {
		t.Errorf("State = %q, want %q", got.State, want.State)
	}
	if got.Branch == nil || *got.Branch != *want.Branch {
		t.Errorf("Branch = %v, want %v", got.Branch, want.Branch)
	}
	if !jsonEqual(t, got.Worktrees, want.Worktrees) {
		t.Errorf("Worktrees = %s, want %s", got.Worktrees, want.Worktrees)
	}
	if got.ArtifactURL == nil || *got.ArtifactURL != *want.ArtifactURL {
		t.Errorf("ArtifactURL = %v, want %v", got.ArtifactURL, want.ArtifactURL)
	}
	if got.JiraIssue == nil || *got.JiraIssue != *want.JiraIssue {
		t.Errorf("JiraIssue = %v, want %v", got.JiraIssue, want.JiraIssue)
	}
	if got.PlanningEffort == nil || *got.PlanningEffort != *want.PlanningEffort {
		t.Errorf("PlanningEffort = %v, want %v", got.PlanningEffort, want.PlanningEffort)
	}
	if !jsonEqual(t, got.Models, want.Models) {
		t.Errorf("Models = %s, want %s", got.Models, want.Models)
	}
	if got.ReviewPanelRoster == nil || *got.ReviewPanelRoster != *want.ReviewPanelRoster {
		t.Errorf("ReviewPanelRoster = %v, want %v", got.ReviewPanelRoster, want.ReviewPanelRoster)
	}
	if got.PRURL == nil || *got.PRURL != *want.PRURL {
		t.Errorf("PRURL = %v, want %v", got.PRURL, want.PRURL)
	}
	if !got.UpdatedAt.Equal(want.UpdatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", got.UpdatedAt, want.UpdatedAt)
	}
	if got.UpdatedBy != want.UpdatedBy {
		t.Errorf("UpdatedBy = %q, want %q", got.UpdatedBy, want.UpdatedBy)
	}
}

func TestPutChangeRefusesMonotonicViolation(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("agents", "kan-16-monotonic")
	c.State = store.StateFinished
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("PutChange(FINISHED): %v", err)
	}

	regressed := c
	regressed.State = store.StateInProgress
	regressed.UpdatedAt = c.UpdatedAt.Add(time.Hour) // later timestamp must not matter

	err := st.PutChange(ctx, regressed)
	if !errors.Is(err, store.ErrMonotonicViolation) {
		t.Fatalf("PutChange(regressed) error = %v, want errors.Is(_, store.ErrMonotonicViolation)", err)
	}

	got, err := st.GetChange(ctx, c.ProjectKey, c.Name)
	if err != nil {
		t.Fatalf("GetChange after refused write: %v", err)
	}
	if got.State != store.StateFinished {
		t.Fatalf("stored state = %q after a refused write, want %q unchanged", got.State, store.StateFinished)
	}
}

// TestPutChangeRefusesOutOfOrderWriteAtSameState is F4: design.md and the
// spec's "An out-of-order write at the same state is refused" scenario
// both require the recorded instant, not just the pipeline state, to order
// same-state writes -- a state-rank-only guard would let the
// chronologically older of two same-state writes silently overwrite the
// newer record's fields, which is exactly what task 6's journal replay (it
// applies writes that can arrive out of order) would trip over.
//
// Mutation check performed by hand: reverting PutChange's WHERE clause to
// `state_rank(changes.state) <= state_rank(EXCLUDED.state)` (state rank
// only) makes this test fail -- the older write applies and UpdatedBy
// becomes "older-writer" -- confirming the test actually exercises the new
// clause rather than passing regardless of it.
func TestPutChangeRefusesOutOfOrderWriteAtSameState(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("agents", "kan-16-same-state-order")
	c.State = store.StateInProgress
	c.UpdatedAt = time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	c.UpdatedBy = "newer-writer"
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("PutChange(newer): %v", err)
	}

	older := c
	older.UpdatedAt = c.UpdatedAt.Add(-time.Hour)
	older.UpdatedBy = "older-writer"

	err := st.PutChange(ctx, older)
	if !errors.Is(err, store.ErrMonotonicViolation) {
		t.Fatalf("PutChange(older, same state) error = %v, want errors.Is(_, store.ErrMonotonicViolation)", err)
	}

	got, err := st.GetChange(ctx, c.ProjectKey, c.Name)
	if err != nil {
		t.Fatalf("GetChange after refused write: %v", err)
	}
	if got.UpdatedBy != "newer-writer" {
		t.Fatalf("stored UpdatedBy = %q after a refused out-of-order write, want %q unchanged", got.UpdatedBy, "newer-writer")
	}
	if !got.UpdatedAt.Equal(c.UpdatedAt) {
		t.Fatalf("stored UpdatedAt = %v after a refused out-of-order write, want %v unchanged", got.UpdatedAt, c.UpdatedAt)
	}
}

// TestPutChangeRefusesIdenticalRetry pins F7's guarantee: a byte-identical
// retry of a Change already stored -- same state, same UpdatedAt -- is
// refused with ErrMonotonicViolation, not silently re-applied. The strict
// `>` on UpdatedAt in PutChange's WHERE clause (added for F4, to close the
// same-state-out-of-order defect) means an identical retry no longer
// matches EXCLUDED.updated_at > changes.updated_at and so falls to the
// same refusal path a genuinely superseded write takes.
//
// This is deliberately safe, not a regression: task 6's journal
// reconciliation retires an entry once the store has accepted *or
// explicitly refused* it, so a refused retry -- caused by a lost response
// to a write that actually landed, or by replaying an entry already
// applied -- is retired exactly as an accepted one would be. See put()'s
// doc comment in internal/api/changes.go for the fuller reconciliation
// story this test exists to keep honest: if a future change flips this
// WHERE clause's `>` to `>=` (reopening the boundary-timestamp collision
// F4 closed), this test fails on the name
// TestPutChangeRefusesIdenticalRetry, naming the exact guarantee that
// broke rather than leaving it to a comment no one re-reads.
func TestPutChangeRefusesIdenticalRetry(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("agents", "kan-16-identical-retry")
	c.State = store.StateInProgress
	c.UpdatedAt = time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	c.UpdatedBy = "myflow-do"
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("PutChange(first): %v", err)
	}

	// A second, byte-identical application of c -- the shape a lost-ack
	// retry or a journal replay of an already-applied entry takes.
	err := st.PutChange(ctx, c)
	if !errors.Is(err, store.ErrMonotonicViolation) {
		t.Fatalf("PutChange(identical retry) error = %v, want errors.Is(_, store.ErrMonotonicViolation)", err)
	}

	got, err := st.GetChange(ctx, c.ProjectKey, c.Name)
	if err != nil {
		t.Fatalf("GetChange after refused retry: %v", err)
	}
	if got.UpdatedBy != c.UpdatedBy || !got.UpdatedAt.Equal(c.UpdatedAt) || got.State != c.State {
		t.Fatalf("stored record changed after a refused identical retry: got %+v, want it unchanged from the first write", got)
	}
}

// TestPutChangeAppliesInOrderWriteAtSameState is the positive counterpart
// to the refusal above: a same-state write carrying a strictly later
// UpdatedAt than the one stored must still apply -- this is not a blanket
// refusal of every same-state write, only of the backwards ones.
func TestPutChangeAppliesInOrderWriteAtSameState(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("agents", "kan-16-same-state-forward")
	c.State = store.StateInProgress
	c.UpdatedAt = time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	c.UpdatedBy = "first-writer"
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("PutChange(first): %v", err)
	}

	newer := c
	newer.UpdatedAt = c.UpdatedAt.Add(time.Hour)
	newer.UpdatedBy = "second-writer"
	if err := st.PutChange(ctx, newer); err != nil {
		t.Fatalf("PutChange(newer, same state) unexpectedly refused: %v", err)
	}

	got, err := st.GetChange(ctx, c.ProjectKey, c.Name)
	if err != nil {
		t.Fatalf("GetChange: %v", err)
	}
	if got.UpdatedBy != "second-writer" {
		t.Fatalf("stored UpdatedBy = %q, want %q", got.UpdatedBy, "second-writer")
	}
}

// TestPutChangeStrictlyLaterStateAppliesRegardlessOfInstant confirms the
// pipeline state remains the primary signal a strictly-forward write is
// judged on: an earlier UpdatedAt must not block a genuine state advance,
// only an equal-state regression in time.
func TestPutChangeStrictlyLaterStateAppliesRegardlessOfInstant(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("agents", "kan-16-state-advance-old-timestamp")
	c.State = store.StateStarted
	c.UpdatedAt = time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("PutChange(STARTED): %v", err)
	}

	advanced := c
	advanced.State = store.StateInProgress
	advanced.UpdatedAt = c.UpdatedAt.Add(-time.Hour) // earlier instant, later state
	advanced.UpdatedBy = "advancer"
	if err := st.PutChange(ctx, advanced); err != nil {
		t.Fatalf("PutChange(IN_PROGRESS, earlier instant) unexpectedly refused: %v", err)
	}

	got, err := st.GetChange(ctx, c.ProjectKey, c.Name)
	if err != nil {
		t.Fatalf("GetChange: %v", err)
	}
	if got.State != store.StateInProgress {
		t.Fatalf("stored State = %q, want %q", got.State, store.StateInProgress)
	}
}

func TestPutChangeIsWholeObject(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("agents", "kan-16-whole-object")
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("first PutChange: %v", err)
	}

	// A second write that omits fields the first write set (nil pointers,
	// empty JSON) must still overwrite them, because PutChange renders the
	// whole record rather than merging. The caller is responsible for
	// carrying forward fields it does not own; the store must not "help"
	// by preserving what a payload leaves out.
	overwrite := store.Change{
		ProjectKey: c.ProjectKey,
		Name:       c.Name,
		State:      store.StateInProgress,
		Worktrees:  json.RawMessage(`{}`),
		UpdatedAt:  c.UpdatedAt.Add(time.Minute),
		UpdatedBy:  "myflow-finish",
	}
	if err := st.PutChange(ctx, overwrite); err != nil {
		t.Fatalf("second PutChange: %v", err)
	}

	got, err := st.GetChange(ctx, c.ProjectKey, c.Name)
	if err != nil {
		t.Fatalf("GetChange: %v", err)
	}
	if got.Branch != nil {
		t.Errorf("Branch = %v after an omitting write, want nil (whole-object write must clear it)", *got.Branch)
	}
	if got.ArtifactURL != nil {
		t.Errorf("ArtifactURL = %v after an omitting write, want nil", *got.ArtifactURL)
	}
	if got.JiraIssue != nil {
		t.Errorf("JiraIssue = %v after an omitting write, want nil", *got.JiraIssue)
	}
	if got.PRURL != nil {
		t.Errorf("PRURL = %v after an omitting write, want nil", *got.PRURL)
	}
	if got.UpdatedBy != "myflow-finish" {
		t.Errorf("UpdatedBy = %q, want %q", got.UpdatedBy, "myflow-finish")
	}
}

func TestListChangesFiltersByProject(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	a1 := baseChange("proj-a", "change-1")
	a2 := baseChange("proj-a", "change-2")
	b1 := baseChange("proj-b", "change-1")

	for _, c := range []store.Change{a1, a2, b1} {
		if err := st.PutChange(ctx, c); err != nil {
			t.Fatalf("PutChange(%s/%s): %v", c.ProjectKey, c.Name, err)
		}
	}

	got, err := st.ListChanges(ctx, "proj-a")
	if err != nil {
		t.Fatalf("ListChanges: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListChanges(proj-a) returned %d changes, want 2", len(got))
	}
	for _, c := range got {
		if c.ProjectKey != "proj-a" {
			t.Errorf("ListChanges(proj-a) returned a change from project %q", c.ProjectKey)
		}
	}
}

func TestSameNameInTwoProjectsCoexist(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	a := baseChange("proj-x", "shared-name")
	b := baseChange("proj-y", "shared-name")
	b.State = store.StateInProgress

	if err := st.PutChange(ctx, a); err != nil {
		t.Fatalf("PutChange(proj-x): %v", err)
	}
	if err := st.PutChange(ctx, b); err != nil {
		t.Fatalf("PutChange(proj-y): %v", err)
	}

	gotA, err := st.GetChange(ctx, "proj-x", "shared-name")
	if err != nil {
		t.Fatalf("GetChange(proj-x): %v", err)
	}
	gotB, err := st.GetChange(ctx, "proj-y", "shared-name")
	if err != nil {
		t.Fatalf("GetChange(proj-y): %v", err)
	}

	if gotA.State != store.StateStarted {
		t.Errorf("proj-x state = %q, want %q", gotA.State, store.StateStarted)
	}
	if gotB.State != store.StateInProgress {
		t.Errorf("proj-y state = %q, want %q", gotB.State, store.StateInProgress)
	}
}

func TestGetChangeUnknownReturnsNotFound(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	_, err := st.GetChange(ctx, "no-such-project", "no-such-change")
	if !errors.Is(err, store.ErrChangeNotFound) {
		t.Fatalf("GetChange error = %v, want errors.Is(_, store.ErrChangeNotFound)", err)
	}
}

// TestPutChangeRejectsInvalidState asserts that a state outside the three
// canonical values (STARTED, IN_PROGRESS, FINISHED) is refused with
// ErrInvalidState rather than silently persisted. Before this check, a
// first-time insert bypasses the ON CONFLICT ... WHERE guard entirely — the
// monotonic rule only fires on a conflict — so state_rank's -1 for an
// unknown value never gets a chance to refuse anything.
func TestPutChangeRejectsInvalidState(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("agents", "kan-16-invalid-state")
	c.State = store.State("BOGUS_STATE")

	err := st.PutChange(ctx, c)
	if !errors.Is(err, store.ErrInvalidState) {
		t.Fatalf("PutChange(BOGUS_STATE) error = %v, want errors.Is(_, store.ErrInvalidState)", err)
	}

	if _, getErr := st.GetChange(ctx, c.ProjectKey, c.Name); !errors.Is(getErr, store.ErrChangeNotFound) {
		t.Fatalf("GetChange after a rejected write = %v, want errors.Is(_, store.ErrChangeNotFound) (nothing should have been persisted)", getErr)
	}
}

// TestPutChangeRejectsEmptyMainCheckoutPathForNewProject asserts the chosen
// resolution of F3: bootstrapping a project row for the first time requires
// a non-empty MainCheckoutPath, refused with ErrInvalidMainCheckoutPath
// rather than silently accepting an empty string that permanently poisons
// the row (main_checkout_path is NOT NULL but not checked non-empty).
func TestPutChangeRejectsEmptyMainCheckoutPathForNewProject(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("brand-new-project", "kan-16-empty-checkout-path")
	c.MainCheckoutPath = ""

	err := st.PutChange(ctx, c)
	if !errors.Is(err, store.ErrInvalidMainCheckoutPath) {
		t.Fatalf("PutChange(empty MainCheckoutPath, new project) error = %v, want errors.Is(_, store.ErrInvalidMainCheckoutPath)", err)
	}

	if _, getErr := st.GetChange(ctx, c.ProjectKey, c.Name); !errors.Is(getErr, store.ErrChangeNotFound) {
		t.Fatalf("GetChange after a rejected write = %v, want errors.Is(_, store.ErrChangeNotFound) (nothing should have been persisted)", getErr)
	}
}

// TestPutChangeAllowsEmptyMainCheckoutPathForExistingProject confirms the
// empty-path guard is scoped to bootstrapping: once a project row exists, a
// write that leaves MainCheckoutPath at its zero value (because the caller
// isn't the one that owns that field) must not be refused.
func TestPutChangeAllowsEmptyMainCheckoutPathForExistingProject(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	first := baseChange("agents", "kan-16-existing-project")
	if err := st.PutChange(ctx, first); err != nil {
		t.Fatalf("first PutChange: %v", err)
	}

	second := first
	second.MainCheckoutPath = ""
	second.State = store.StateInProgress
	second.UpdatedAt = first.UpdatedAt.Add(time.Minute)
	if err := st.PutChange(ctx, second); err != nil {
		t.Fatalf("PutChange with empty MainCheckoutPath against an existing project: %v", err)
	}
}

// TestPutChangeInvalidStateIsNotConflatedWithMonotonicViolation is F5: a
// caller sending a malformed state (wrong case, typo) against a change that
// already has a stored state must get ErrInvalidState, never
// ErrMonotonicViolation — the two mean different things to the caller and
// errors.Is must be able to tell them apart.
func TestPutChangeInvalidStateIsNotConflatedWithMonotonicViolation(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("agents", "kan-16-case-mismatch")
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("first PutChange: %v", err)
	}

	malformed := c
	malformed.State = store.State("in_progress") // wrong case
	malformed.UpdatedAt = c.UpdatedAt.Add(time.Minute)

	err := st.PutChange(ctx, malformed)
	if !errors.Is(err, store.ErrInvalidState) {
		t.Errorf("PutChange(wrong-case state) error = %v, want errors.Is(_, store.ErrInvalidState)", err)
	}
	if errors.Is(err, store.ErrMonotonicViolation) {
		t.Errorf("PutChange(wrong-case state) error = %v, must not also satisfy errors.Is(_, store.ErrMonotonicViolation)", err)
	}
}

// TestConcurrentPutChangeBootstrapsNewProjectWithoutRace is F9: several
// goroutines call PutChange for the same brand-new project key at once,
// each with a distinct change name and a valid MainCheckoutPath. A
// check-then-insert bootstrap (SELECT EXISTS, then a conditional INSERT)
// lets every goroutine observe "absent" before any of them commits, so all
// but one INSERT the same project_key and the losers get a bare
// unique-violation instead of a successful write. The bootstrap must be a
// single atomic statement so every concurrent first-writer either succeeds
// outright or defers cleanly to whichever writer actually created the row.
func TestConcurrentPutChangeBootstrapsNewProjectWithoutRace(t *testing.T) {
	st := newTestStore(t)

	const writers = 5
	projectKey := fmt.Sprintf("brand-new-concurrent-%d", time.Now().UnixNano())

	var wg sync.WaitGroup
	errs := make([]error, writers)

	for i := range writers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			c := baseChange(projectKey, fmt.Sprintf("change-%d", i))
			errs[i] = st.PutChange(ctx, c)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("writer %d: PutChange: %v", i, err)
		}
	}

	ctx := context.Background()
	got, err := st.ListChanges(ctx, projectKey)
	if err != nil {
		t.Fatalf("ListChanges: %v", err)
	}
	if len(got) != writers {
		t.Fatalf("ListChanges(%s) returned %d changes, want %d (one per concurrent writer)", projectKey, len(got), writers)
	}
}

// --- ProjectKeysByDisplayName -------------------------------------------
//
// ProjectKeysByDisplayName is the server-side twin of
// stats/web/src/lib/projectLabel.ts: it derives a display name from
// project_key in SQL by trimming the same documented suffix -- a trailing
// "-" plus exactly eight lowercase hex characters -- so the two sides
// agree on what a key "displays as" by construction rather than by
// comment. These tests seed real project rows through PutChange (the only
// way a project row is created) and read them back by display name.

func TestProjectKeysByDisplayNameReturnsBothKeysForSharedDisplayName(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	base := fmt.Sprintf("proj-shared-%d", time.Now().UnixNano())
	keyA := base + "-a740d89c"
	keyB := base + "-7c1f238a"

	if err := st.PutChange(ctx, baseChange(keyA, "kan-1")); err != nil {
		t.Fatalf("PutChange %s: %v", keyA, err)
	}
	if err := st.PutChange(ctx, baseChange(keyB, "kan-1")); err != nil {
		t.Fatalf("PutChange %s: %v", keyB, err)
	}

	got, err := st.ProjectKeysByDisplayName(ctx, base)
	if err != nil {
		t.Fatalf("ProjectKeysByDisplayName(%s): %v", base, err)
	}
	sort.Strings(got)
	want := []string{keyA, keyB}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ProjectKeysByDisplayName(%s) = %v, want %v", base, got, want)
	}
}

func TestProjectKeysByDisplayNameReturnsOneKeyForUniqueDisplayName(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	base := fmt.Sprintf("proj-unique-%d", time.Now().UnixNano())
	key := base + "-a740d89c"

	if err := st.PutChange(ctx, baseChange(key, "kan-1")); err != nil {
		t.Fatalf("PutChange %s: %v", key, err)
	}

	got, err := st.ProjectKeysByDisplayName(ctx, base)
	if err != nil {
		t.Fatalf("ProjectKeysByDisplayName(%s): %v", base, err)
	}
	want := []string{key}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ProjectKeysByDisplayName(%s) = %v, want %v", base, got, want)
	}
}

func TestProjectKeysByDisplayNameReturnsNoneForUnknownName(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	base := fmt.Sprintf("proj-unknown-%d", time.Now().UnixNano())

	got, err := st.ProjectKeysByDisplayName(ctx, base)
	if err != nil {
		t.Fatalf("ProjectKeysByDisplayName(%s): %v", base, err)
	}
	if len(got) != 0 {
		t.Fatalf("ProjectKeysByDisplayName(%s) = %v, want none", base, got)
	}
}
