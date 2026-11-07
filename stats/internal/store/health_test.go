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
	st := newTestStore(t)
	_, err := st.RecordGuardRun(context.Background(), "no-such-project", guardRun("g", "clear", 0, 1, time.Now()))
	if !errors.Is(err, store.ErrProjectNotFound) {
		t.Fatalf("err = %v, want ErrProjectNotFound", err)
	}
}

func TestRecordGuardRunRefusesAnInvalidRun(t *testing.T) {
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
