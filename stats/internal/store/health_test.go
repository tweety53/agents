package store_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

func guardRun(guard, outcome string, exit, ms int, at time.Time) records.GuardRun {
	return records.GuardRun{Guard: guard, Worktree: "/tmp/wt", ExitCode: exit, Outcome: outcome, DurationMS: ms, RecordedAt: at}
}

func TestRecordGuardRunRefusesAnUnknownProject(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	_, err := st.RecordGuardRun(context.Background(), "no-such-project", guardRun("g", "clear", 0, 1, time.Now()))
	if !errors.Is(err, store.ErrProjectNotFound) {
		t.Fatalf("err = %v, want ErrProjectNotFound", err)
	}
}

func TestRecordGuardRunRefusesAnInvalidRun(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	projectKey := fmt.Sprintf("proj-grinv-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	for name, in := range map[string]records.GuardRun{
		"no guard":         guardRun("", "clear", 0, 1, time.Now()),
		"unknown outcome":  guardRun("g", "exploded", 3, 1, time.Now()),
		"negative runtime": guardRun("g", "clear", 0, -1, time.Now()),
	} {
		if _, err := st.RecordGuardRun(context.Background(), projectKey, in); !errors.Is(err, store.ErrInvalidGuardRun) {
			t.Errorf("%s: err = %v, want ErrInvalidGuardRun", name, err)
		}
	}
}

// TestRecordGuardRunStoresReasons pins guard_runs.reasons: the run's ids
// come back in the order recorded, an absent or empty list is stored NULL
// (read back nil, never an empty slice), and an empty id is refused
// (decisions reasons-array, reason-nullable).
func TestRecordGuardRunStoresReasons(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	projectKey := fmt.Sprintf("proj-grrsn-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	ctx := context.Background()

	in := guardRun("check-references", "fired", 1, 1, time.Now())
	in.Reasons = []string{"F2", "plan-step-files"}
	got, err := st.RecordGuardRun(ctx, projectKey, in)
	if err != nil {
		t.Fatalf("RecordGuardRun: %v", err)
	}
	if !slices.Equal(got.Reasons, []string{"F2", "plan-step-files"}) {
		t.Errorf("reasons = %q, want [F2 plan-step-files]", got.Reasons)
	}

	for name, reasons := range map[string][]string{"nil": nil, "empty": {}} {
		in := guardRun("check-references", "clear", 0, 1, time.Now())
		in.Reasons = reasons
		got, err := st.RecordGuardRun(ctx, projectKey, in)
		if err != nil {
			t.Fatalf("%s: RecordGuardRun: %v", name, err)
		}
		if got.Reasons != nil {
			t.Errorf("%s: reasons = %#v, want nil (stored NULL)", name, got.Reasons)
		}
	}

	in = guardRun("check-references", "fired", 1, 1, time.Now())
	in.Reasons = []string{"F2", ""}
	if _, err := st.RecordGuardRun(ctx, projectKey, in); !errors.Is(err, store.ErrInvalidGuardRun) {
		t.Errorf("empty id: err = %v, want ErrInvalidGuardRun", err)
	}
}

func recordGuardRuns(t *testing.T, st *store.Store, projectKey string, runs ...records.GuardRun) {
	t.Helper()
	for _, r := range runs {
		if _, err := st.RecordGuardRun(context.Background(), projectKey, r); err != nil {
			t.Fatalf("RecordGuardRun %s %s: %v", projectKey, r.Guard, err)
		}
	}
}

// TestGuardHealthPerGuard pins the guard-health view's per-guard row: counts
// by outcome, the median and summed duration, the latest fired instant (not
// the latest run), the window and project filters, and helper exclusion.
func TestGuardHealthPerGuard(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	projA := fmt.Sprintf("proj-gh-pa-%d", suffix)
	projB := fmt.Sprintf("proj-gh-pb-%d", suffix)
	seedChange(t, st, projA, "kan-1")
	seedChange(t, st, projB, "kan-1")

	day := time.Date(2033, 4, 10, 0, 0, 0, 0, time.UTC)
	at := func(h int) time.Time { return day.Add(time.Duration(h) * time.Hour) }
	recordGuardRuns(t, st, projA,
		guardRun("check-a", "fired", 1, 100, at(1)),
		guardRun("check-a", "fired", 1, 900, at(2)),
		guardRun("check-a", "clear", 0, 200, at(3)),
		guardRun("check-b", "cannot-answer", 2, 50, at(4)),
		guardRun("check-b", "clear", 0, 150, at(5)),
		guardRun("project-get", "fired", 1, 10, at(6)),
		guardRun("check-a", "fired", 1, 100, at(-1)),
	)
	recordGuardRuns(t, st, projB, guardRun("check-a", "fired", 1, 100, at(7)))

	period := store.Period{From: day, To: day.Add(24 * time.Hour)}
	got, err := st.GuardHealth(ctx, period, &projA)
	if err != nil {
		t.Fatalf("GuardHealth: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("GuardHealth(A) = %+v, want check-a and check-b only", got)
	}
	a, b := got[0], got[1]
	lastA := at(2)
	if a.Guard != "check-a" || a.Runs != 3 || a.Fired != 2 || a.CannotAnswer != 0 ||
		a.MedianMs != 200 || a.TotalMs != 1200 || a.LastFiredAt == nil || !a.LastFiredAt.Equal(lastA) {
		t.Errorf("check-a = %+v (LastFiredAt %v), want Runs 3 Fired 2 CannotAnswer 0 MedianMs 200 TotalMs 1200 LastFiredAt %v", a, a.LastFiredAt, lastA)
	}
	if b.Guard != "check-b" || b.Runs != 2 || b.Fired != 0 || b.CannotAnswer != 1 ||
		b.MedianMs != 100 || b.TotalMs != 200 || b.LastFiredAt != nil {
		t.Errorf("check-b = %+v, want Runs 2 Fired 0 CannotAnswer 1 MedianMs 100 TotalMs 200 LastFiredAt nil", b)
	}

	all, err := st.GuardHealth(ctx, period, nil)
	if err != nil {
		t.Fatalf("GuardHealth(nil): %v", err)
	}
	for _, r := range all {
		if r.Guard == "project-get" {
			t.Errorf("GuardHealth(nil) carries the helper project-get: %+v", r)
		}
		if r.Guard == "check-a" && (r.Runs != 4 || r.Fired != 3) {
			t.Errorf("GuardHealth(nil) check-a = %+v, want Runs 4 Fired 3 (both projects)", r)
		}
	}
}

// TestGuardHealthProjectsCountsGuardRuns pins the Project chips' counts:
// helpers excluded, count descending before key.
func TestGuardHealthProjectsCountsGuardRuns(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	projZ := fmt.Sprintf("proj-gh-z-%d", suffix)
	projA := fmt.Sprintf("proj-gh-a-%d", suffix)
	seedChange(t, st, projZ, "kan-1")
	seedChange(t, st, projA, "kan-1")

	at := time.Date(2033, 5, 10, 9, 0, 0, 0, time.UTC)
	recordGuardRuns(t, st, projZ,
		guardRun("check-a", "clear", 0, 1, at),
		guardRun("check-a", "fired", 1, 1, at),
		guardRun("check-b", "clear", 0, 1, at),
		guardRun("project-get", "fired", 1, 1, at),
		guardRun("project-get", "clear", 0, 1, at),
	)
	recordGuardRuns(t, st, projA, guardRun("check-a", "clear", 0, 1, at))

	period := store.Period{From: time.Date(2033, 5, 10, 0, 0, 0, 0, time.UTC), To: time.Date(2033, 5, 11, 0, 0, 0, 0, time.UTC)}
	got, err := st.GuardHealthProjects(ctx, period)
	if err != nil {
		t.Fatalf("GuardHealthProjects: %v", err)
	}
	want := []store.ProjectGuardRuns{{ProjectKey: projZ, Runs: 3}, {ProjectKey: projA, Runs: 1}}
	if !slices.Equal(got, want) {
		t.Errorf("GuardHealthProjects = %+v, want %+v", got, want)
	}
}

// TestGuardHealthWindowBounds pins the half-open window both queries share:
// a run exactly at From counts, a run exactly at To does not.
func TestGuardHealthWindowBounds(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()
	proj := fmt.Sprintf("proj-gh-bounds-%d", time.Now().UnixNano())
	seedChange(t, st, proj, "kan-1")
	from := time.Date(2033, 6, 10, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)
	recordGuardRuns(t, st, proj,
		guardRun("check-edge", "fired", 1, 1, from),
		guardRun("check-edge", "fired", 1, 1, to),
	)
	period := store.Period{From: from, To: to}

	got, err := st.GuardHealth(ctx, period, &proj)
	if err != nil {
		t.Fatalf("GuardHealth: %v", err)
	}
	if len(got) != 1 || got[0].Runs != 1 {
		t.Errorf("GuardHealth = %+v, want check-edge with Runs 1 (From in, To out)", got)
	}
	projects, err := st.GuardHealthProjects(ctx, period)
	if err != nil {
		t.Fatalf("GuardHealthProjects: %v", err)
	}
	want := []store.ProjectGuardRuns{{ProjectKey: proj, Runs: 1}}
	if !slices.Equal(projects, want) {
		t.Errorf("GuardHealthProjects = %+v, want %+v (From in, To out)", projects, want)
	}
}

// TestRecordGuardRunFindsAPeerWorktreesProject: a run posted under a project
// key the store does not hold -- a guard whose cwd is a peer repository's
// worktree, or a wave copy of one -- is recorded under the project of the
// change whose worktrees map names that worktree; a worktree no change names
// is still refused.
func TestRecordGuardRunFindsAPeerWorktreesProject(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-grpeer-%d", time.Now().UnixNano())
	c := baseChange(projectKey, "kan-1")
	c.Worktrees = []byte(`{"/w/app-worktrees/kan-1":"aaa","/w/peer-worktrees/kan-1":"bbb"}`)
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatal(err)
	}
	for _, wt := range []string{"/w/peer-worktrees/kan-1", "/w/peer-worktrees/kan-1/fidelity", "/w/peer-worktrees/kan-1-wave-group-2"} {
		in := guardRun("check-references", "clear", 0, 1, time.Now())
		in.Worktree = wt
		if _, err := st.RecordGuardRun(ctx, "peer-0000beef", in); err != nil {
			t.Errorf("%s: err = %v, want the run recorded", wt, err)
		}
	}
	rows, err := st.GuardHealth(ctx, store.Period{From: time.Now().Add(-time.Hour), To: time.Now().Add(time.Hour)}, &projectKey)
	if err != nil || len(rows) != 1 || rows[0].Runs != 3 {
		t.Errorf("guard health = %+v, %v; want one guard with 3 runs under the change's project", rows, err)
	}
	in := guardRun("g", "clear", 0, 1, time.Now())
	in.Worktree = "/w/peer-worktrees/kan-10"
	if _, err := st.RecordGuardRun(ctx, "peer-0000beef", in); !errors.Is(err, store.ErrProjectNotFound) {
		t.Errorf("unnamed worktree: err = %v, want ErrProjectNotFound", err)
	}
}

// TestChangeGuardRunsMatchesTheWorktreeSegment pins ChangeGuardRuns
// (decision parent-time-guard-runs): a run is the change's when its worktree
// carries the segment -worktrees/<change> followed by "/" or the end, in any
// repository's worktree directory and under any project -- never a change
// whose name only starts with it (kan-1-ab) -- and only inside the window; a
// main-checkout run of the asked project in the window is counted unclaimed,
// another project's is not.
func TestChangeGuardRunsMatchesTheWorktreeSegment(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()
	p := fmt.Sprintf("proj-cgr-p-%d", time.Now().UnixNano())
	q := fmt.Sprintf("proj-cgr-q-%d", time.Now().UnixNano())
	seedChange(t, st, p, "kan-1-a")
	seedChange(t, st, q, "kan-9")
	from := time.Date(2034, 3, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	at := from.Add(30 * time.Minute)
	run := func(wt string, when time.Time) records.GuardRun {
		r := guardRun("check-references", "clear", 0, 1000, when)
		r.Worktree = wt
		return r
	}
	recordGuardRuns(t, st, p,
		run("/x/gymie-worktrees/kan-1-a", at),
		run("/x/gymie-frontend-worktrees/kan-1-a/sub", at.Add(time.Minute)),
		run("/x/gymie-worktrees/kan-1-ab", at),
		run("/x/gymie", at),
		run("/x/gymie-worktrees/kan-1-a", to.Add(time.Hour)),
	)
	recordGuardRuns(t, st, q, run("/y/other", at), run("/z/other-worktrees/kan-1-a", at.Add(2*time.Minute)))

	gs, err := st.ChangeGuardRuns(ctx, []store.GuardWindow{
		{ProjectKey: p, Change: "kan-1-a", From: from, To: to},
		{ProjectKey: q, Change: "kan-9", From: from, To: to},
	})
	if err != nil {
		t.Fatalf("ChangeGuardRuns: %v", err)
	}
	if len(gs) != 2 {
		t.Fatalf("len = %d, want one result per window", len(gs))
	}
	if len(gs[1].Claimed) != 0 || gs[1].Unclaimed != 1 {
		t.Errorf("second window = %+v, want no claimed and 1 unclaimed", gs[1])
	}
	unclaimed := gs[0].Unclaimed
	var got []string
	for _, r := range gs[0].Claimed {
		got = append(got, r.Worktree)
	}
	want := []string{"/x/gymie-worktrees/kan-1-a", "/x/gymie-frontend-worktrees/kan-1-a/sub", "/z/other-worktrees/kan-1-a"}
	if !slices.Equal(got, want) {
		t.Errorf("claimed = %v, want %v", got, want)
	}
	if unclaimed != 1 {
		t.Errorf("unclaimed = %d, want 1 (the main-checkout run of %s)", unclaimed, p)
	}
}
