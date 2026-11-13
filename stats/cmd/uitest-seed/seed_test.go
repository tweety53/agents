package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/config"
	dsnutil "github.com/tweety53/agents/stats/internal/dsn"
	"github.com/tweety53/agents/stats/internal/smoke"
	"github.com/tweety53/agents/stats/internal/store"
)

// adminDSN is the DSN used to create and drop the throwaway database this
// test seeds into. It points at the flow-postgres compose stack on host
// port 5433 -- the same stack internal/store's own tests use, and the same
// DSN config.DefaultDSN already names -- and can be overridden for
// environments that run Postgres elsewhere.
func adminDSN() string {
	if v := os.Getenv("FLOW_STATS_ADMIN_DSN"); v != "" {
		return v
	}
	return config.DefaultDSN
}

// It derives from adminDSN rather than repeating the credentials, so that
// FLOW_STATS_ADMIN_DSN moves both connections together. Hardcoding them here
// made the override only half-work -- the admin connection followed the
// environment while the per-test one stayed pinned, so a Postgres with
// different credentials failed every test while the admin step succeeded. See
// internal/store/testsupport_test.go's testDSN for the measured case.
func testDSN(dbName string) string {
	out, err := dsnutil.ForDatabase(adminDSN(), dbName)
	if err != nil {
		// Panic rather than return a best-effort string. This helper's
		// predecessor answered confidently when it could not do the job and
		// silently handed back a corrupted DSN; a test that then connects
		// reports whatever that connection says, which is the wrong question
		// answered convincingly.
		panic(fmt.Sprintf("deriving a per-test DSN: %v", err))
	}
	return out
}

// newUitestDatabase creates a uniquely-named database whose name ends in
// "_uitest" -- so requireUitestDatabase admits it -- and registers a
// cleanup that drops it once the test finishes. It never touches the live
// "flow" database or the shared "flow_uitest" one: both are named
// explicitly here only as the admin connection used to create and drop
// this test's own, disposable database. It skips cleanly, with a clear
// message, when the compose stack is not reachable, exactly as
// internal/store's own per-test-database helper does.
func newUitestDatabase(t *testing.T) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	adminPool, err := pgxpool.New(ctx, adminDSN())
	if err != nil {
		t.Skipf("flow-postgres compose stack not reachable: %v", err)
	}
	if err := adminPool.Ping(ctx); err != nil {
		adminPool.Close()
		t.Skipf("flow-postgres compose stack not reachable: %v", err)
	}

	dbName := fmt.Sprintf("uitest_seed_test_%d_%d_uitest", os.Getpid(), time.Now().UnixNano())
	ident := pgx.Identifier{dbName}.Sanitize()
	if _, err := adminPool.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
		adminPool.Close()
		t.Fatalf("create test database %s: %v", dbName, err)
	}
	adminPool.Close()

	t.Cleanup(func() {
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer dropCancel()

		dropPool, err := pgxpool.New(dropCtx, adminDSN())
		if err != nil {
			t.Logf("drop test database %s: reconnect failed: %v", dbName, err)
			return
		}
		defer dropPool.Close()

		if _, err := dropPool.Exec(dropCtx, "DROP DATABASE IF EXISTS "+ident+" WITH (FORCE)"); err != nil {
			t.Logf("drop test database %s: %v", dbName, err)
		}
	})

	return testDSN(dbName)
}

// TestSeedPopulatesFixture proves seedFixture writes what
// the UI-test-stack requirement "The test stack starts populated"
// requires: two projects, a change in every one of the three pipeline
// states, and at least one stage run carrying token usage a cost view
// could render. It runs against its own throwaway "..._uitest" database,
// never the live "flow" database or the shared "flow_uitest" one.
// smoke: core — the seeder a store change can break
func TestSeedPopulatesFixture(t *testing.T) {
	dsn := newUitestDatabase(t)
	if err := requireUitestDatabase(dsn); err != nil {
		t.Fatalf("guard refused the test's own throwaway database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	st, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()
	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	if err := seedFixture(ctx, st); err != nil {
		t.Fatalf("seedFixture: %v", err)
	}

	projectKeys := make(map[string]bool)
	statesSeen := make(map[store.State]bool)
	var stageRunsWithTokens int

	for _, project := range fixtureData() {
		projectKeys[project.Key] = true

		changes, err := st.ListChanges(ctx, project.Key)
		if err != nil {
			t.Fatalf("list changes for %s: %v", project.Key, err)
		}
		wantChanges := len(project.Changes)
		if project.Key == "uitest-alpha" {
			wantChanges += 4 // seedInteractive's two sessions and seedMainExperiment's two changes, seeded beside fixtureData
		}
		switch project.Key { // seedDecisions' runs, seeded beside fixtureData
		case "uitest-alpha":
			wantChanges += 13
		case "uitest-beta":
			wantChanges += 5
		}
		if len(changes) != wantChanges {
			t.Fatalf("project %s: got %d changes, want %d", project.Key, len(changes), wantChanges)
		}

		for _, c := range changes {
			statesSeen[c.State] = true
		}

		for _, wantChange := range project.Changes {
			got, err := st.GetChange(ctx, project.Key, wantChange.Name)
			if err != nil {
				t.Fatalf("get change %s/%s: %v", project.Key, wantChange.Name, err)
			}
			if got.State != wantChange.State {
				t.Fatalf("change %s/%s: got state %s, want %s", project.Key, wantChange.Name, got.State, wantChange.State)
			}

			runs, _, err := st.QueryStageRuns(ctx, store.Query{
				Filters: []store.Filter{
					{Field: "project", Op: store.OpEq, Value: project.Key},
					{Field: "name", Op: store.OpEq, Value: wantChange.Name},
				},
				Limit: store.NoLimit,
			})
			if err != nil {
				t.Fatalf("query stage runs for %s/%s: %v", project.Key, wantChange.Name, err)
			}
			if len(runs) != len(wantChange.Stage) {
				t.Fatalf("change %s/%s: got %d stage runs, want %d", project.Key, wantChange.Name, len(runs), len(wantChange.Stage))
			}
			for _, run := range runs {
				var bag struct {
					Tokens struct {
						Main struct {
							Input *float64 `json:"input"`
						} `json:"main"`
					} `json:"tokens"`
				}
				if err := json.Unmarshal(run.Metrics, &bag); err != nil {
					t.Fatalf("decode metrics for stage run %d: %v", run.ID, err)
				}
				if bag.Tokens.Main.Input != nil && *bag.Tokens.Main.Input > 0 {
					stageRunsWithTokens++
				}
			}
		}
	}

	if len(projectKeys) != 2 {
		t.Fatalf("got %d distinct projects, want 2", len(projectKeys))
	}
	for _, want := range []store.State{store.StateStarted, store.StateInProgress, store.StateFinished} {
		if !statesSeen[want] {
			t.Fatalf("no seeded change carries state %s", want)
		}
	}
	if stageRunsWithTokens == 0 {
		t.Fatalf("no seeded stage run carries token usage")
	}

	// The open experiment's Tokens/change and Wall-clock panels draw figures
	// on both sides of its start (KAN-1067 visual-verify fix 2).
	list, err := st.ListExperiments(ctx, true)
	if err != nil {
		t.Fatalf("list experiments: %v", err)
	}
	found := 0
	for _, e := range list {
		if e.ClosedAt != nil {
			continue
		}
		stats, err := st.ExperimentStats(ctx, e.ID)
		if err != nil {
			t.Fatalf("experiment stats %d: %v", e.ID, err)
		}
		for _, m := range stats.Metrics {
			if m.Name != "tokens" && m.Name != "wallClockMin" {
				continue
			}
			found++
			if m.Before == nil || m.After == nil || *m.Before <= 0 || *m.After <= 0 {
				t.Fatalf("open experiment %q: %s before=%v after=%v, want both positive", e.Name, m.Name, m.Before, m.After)
			}
		}
	}
	if found != 2 {
		t.Fatalf("open experiments carry %d tokens/wallClockMin metrics, want 2", found)
	}

	// The closed main-cell experiment draws selfReviewFindings on both
	// sides (KAN-1071 visual-verify fix 1): 2 before, 1 after.
	all, err := st.ListExperiments(ctx, false)
	if err != nil {
		t.Fatalf("list experiments: %v", err)
	}
	var mainSR *store.ExperimentMetric
	for _, e := range all {
		if _, ok := e.Cells["main"]; !ok {
			continue
		}
		stats, err := st.ExperimentStats(ctx, e.ID)
		if err != nil {
			t.Fatalf("experiment stats %d: %v", e.ID, err)
		}
		for i, m := range stats.Metrics {
			if m.Name == "selfReviewFindings" {
				mainSR = &stats.Metrics[i]
			}
		}
	}
	if mainSR == nil || mainSR.Before == nil || mainSR.After == nil || *mainSR.Before != 2 || *mainSR.After != 1 {
		t.Fatalf("main-cell experiment selfReviewFindings = %+v, want before 2, after 1", mainSR)
	}
}

// TestSeedInteractiveSessions pins seedInteractive (KAN-1061): the
// prototype's running session lists its eight requests in their seeded
// states with its open run carrying request 7, and the waiting session its
// three landed requests.
func TestSeedInteractiveSessions(t *testing.T) {
	smoke.Heavy(t)
	dsn := newUitestDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	st, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()
	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	if err := seedFixture(ctx, st); err != nil {
		t.Fatalf("seedFixture: %v", err)
	}

	states := func(change string) []string {
		t.Helper()
		reqs, err := st.ListRequests(ctx, "uitest-alpha", change)
		if err != nil {
			t.Fatalf("list requests for %s: %v", change, err)
		}
		var out []string
		for _, r := range reqs {
			out = append(out, fmt.Sprintf("%d %s", r.N, r.State))
		}
		return out
	}
	want := []string{"1 landed", "2 landed", "3 answered", "4 landed", "5 dropped", "6 landed", "7 in progress", "8 queued"}
	if got := states(interactiveRunning); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("%s requests = %v, want %v", interactiveRunning, got, want)
	}
	if got := states(interactiveWaiting); fmt.Sprint(got) != fmt.Sprint([]string{"1 landed", "2 landed", "3 landed"}) {
		t.Errorf("%s requests = %v, want three landed", interactiveWaiting, got)
	}

	runs, _, err := st.QueryStageRuns(ctx, store.Query{Filters: []store.Filter{
		{Field: "project", Op: store.OpEq, Value: "uitest-alpha"}, {Field: "name", Op: store.OpEq, Value: interactiveRunning},
		{Field: "ended_at", Op: store.OpNull}}, Limit: store.NoLimit})
	if err != nil {
		t.Fatalf("query open runs: %v", err)
	}
	if len(runs) != 1 || runs[0].RequestN == nil || *runs[0].RequestN != 7 || runs[0].Command != "/flow-interactive" {
		t.Fatalf("open runs = %+v, want one /flow-interactive run carrying request 7", runs)
	}
}

// TestSeedTrendFinishedChanges pins seedTrend (KAN-1051): over the twelve
// Wednesday weeks ending with fixtureNow's, /flow changes fill at least six
// weeks with an empty week between two filled ones and one week holds three
// or more distinct costs; /flow-fast fills at least two weeks; changes land
// in an agents-<8 hex> project and in one other.
func TestSeedTrendFinishedChanges(t *testing.T) {
	smoke.Heavy(t)
	dsn := newUitestDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	st, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()
	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	if err := seedFixture(ctx, st); err != nil {
		t.Fatalf("seedFixture: %v", err)
	}

	current := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC) // fixtureNow's Wednesday week
	from := current.Add(-11 * 7 * 24 * time.Hour)
	rows, err := st.TrendFinishedChanges(ctx, store.Period{From: from, To: fixtureNow})
	if err != nil {
		t.Fatalf("TrendFinishedChanges: %v", err)
	}
	flowWeek := map[time.Time]map[float64]bool{}
	fastWeeks := map[time.Time]bool{}
	agents, others := false, false
	agentsKey := regexp.MustCompile(`^agents-[0-9a-f]{8}$`)
	for _, r := range rows {
		r.Week = r.Week.UTC() // map keys compare location too
		switch r.Command {
		case "/flow":
			if flowWeek[r.Week] == nil {
				flowWeek[r.Week] = map[float64]bool{}
			}
			flowWeek[r.Week][r.CostUSD] = true
		case "/flow-fast":
			fastWeeks[r.Week] = true
		}
		if agentsKey.MatchString(r.ProjectKey) {
			agents = true
		} else {
			others = true
		}
	}
	if len(flowWeek) < 6 {
		t.Errorf("/flow fills %d weeks, want at least 6", len(flowWeek))
	}
	if len(fastWeeks) < 2 {
		t.Errorf("/flow-fast fills %d weeks, want at least 2", len(fastWeeks))
	}
	gap, three := false, false
	for i := 1; i < 11; i++ {
		w := from.Add(time.Duration(i) * 7 * 24 * time.Hour)
		prev, next := w.Add(-7*24*time.Hour), w.Add(7*24*time.Hour)
		if flowWeek[w] == nil && flowWeek[prev] != nil && flowWeek[next] != nil {
			gap = true
		}
	}
	for _, costs := range flowWeek {
		if len(costs) >= 3 {
			three = true
		}
	}
	if !gap {
		t.Error("no empty /flow week between two filled ones")
	}
	if !three {
		t.Error("no week holds three or more distinct /flow costs")
	}
	if !agents || !others {
		t.Errorf("agents project seen %v, another project seen %v; want both", agents, others)
	}
}

// TestSeedPricesReviewerDispatches: kan-102's primary and principles panel
// dispatches carry the agent ids its flow.review-panel run's priced buckets
// key, so Reviewers reads their spend (decision reviewers-seed-agent-ids).
func TestSeedPricesReviewerDispatches(t *testing.T) {
	smoke.Heavy(t)
	dsn := newUitestDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	st, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()
	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	if err := seedFixture(ctx, st); err != nil {
		t.Fatalf("seedFixture: %v", err)
	}

	period := store.Period{
		From: time.Date(2026, 7, 16, 9, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 8, 15, 9, 0, 0, 0, time.UTC),
	}
	project := "uitest-alpha"
	rows, err := st.Reviewers(ctx, period, &project, nil)
	if err != nil {
		t.Fatalf("Reviewers: %v", err)
	}
	bySlot := map[string]store.ReviewerRow{}
	for _, r := range rows {
		bySlot[r.Slot] = r
	}
	const eps = 1e-9
	for slot, cost := range map[string]float64{"primary": 0.15, "principles": 0.1} {
		got, ok := bySlot[slot]
		if !ok {
			t.Fatalf("no Reviewers row for %q (rows %+v)", slot, rows)
		}
		if got.PricedDispatches != 1 || math.Abs(got.CostUSD-cost) > eps {
			t.Errorf("%s = priced %d cost %v, want priced 1 cost %v", slot, got.PricedDispatches, got.CostUSD, cost)
		}
	}
}

// TestSeedDrawsReviewerStates: the default 30d Reviewers window over
// uitest-alpha holds every state the approved prototype draws -- a roster
// description on the experimental slot, an experimental slot priced with no findings ("found
// nothing"), a non-zero withdrawn share, different heads for the Severity
// and False alarms ranks, and enough rows to run past a 1024px viewport.
func TestSeedDrawsReviewerStates(t *testing.T) {
	smoke.Heavy(t)
	dsn := newUitestDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	st, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()
	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	if err := seedFixture(ctx, st); err != nil {
		t.Fatalf("seedFixture: %v", err)
	}

	period := store.Period{From: fixtureNow.Add(-30 * 24 * time.Hour), To: fixtureNow}
	project := "uitest-alpha"
	rows, err := st.Reviewers(ctx, period, &project, nil)
	if err != nil {
		t.Fatalf("Reviewers: %v", err)
	}
	if len(rows) < 13 {
		t.Errorf("rows = %d, want at least 13", len(rows))
	}
	var described, found, withdrawn bool
	var severest, alarmiest store.ReviewerRow
	wpd := func(r store.ReviewerRow) float64 {
		return float64(10*r.Critical+4*r.Important+r.Minor) / float64(r.Dispatches)
	}
	share := func(r store.ReviewerRow) float64 {
		n := r.Critical + r.Important + r.Minor + r.Withdrawn
		if n == 0 {
			return -1
		}
		return float64(r.Withdrawn) / float64(n)
	}
	for _, r := range rows {
		if r.Experimental && r.Description != "" {
			described = true
		}
		if severest.Slot == "" || wpd(r) > wpd(severest) {
			severest = r
		}
		if alarmiest.Slot == "" || share(r) > share(alarmiest) {
			alarmiest = r
		}
		if r.Experimental && r.PricedDispatches > 0 && r.Critical+r.Important+r.Minor == 0 {
			found = true
		}
		if r.Withdrawn > 0 {
			withdrawn = true
		}
	}
	if !described {
		t.Error("no experimental slot carries its roster description")
	}
	if severest.Slot == alarmiest.Slot {
		t.Errorf("%s heads both the Severity and the False alarms rank; want them to differ", severest.Slot)
	}
	if !found {
		t.Error("no priced experimental slot that found nothing")
	}
	if !withdrawn {
		t.Error("no slot with a withdrawn finding")
	}
}

// TestSeedDecisionsDrawsEveryState pins decision decisions-seed-isolated:
// the eighteen decisions runs sit inside the default 30d window at
// fixtureNow, draw every state the Decisions prototype shows, and stay out
// of PINNED_QUERY's window.
func TestSeedDecisionsDrawsEveryState(t *testing.T) {
	smoke.Heavy(t)
	dsn := newUitestDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	st, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()
	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	if err := seedFixture(ctx, st); err != nil {
		t.Fatalf("seedFixture: %v", err)
	}

	seeded := regexp.MustCompile(`^kan-6\d\d-decisions$`)
	rows, err := st.Decisions(ctx, store.Period{From: fixtureNow.Add(-30 * 24 * time.Hour), To: fixtureNow}, nil)
	if err != nil {
		t.Fatalf("Decisions: %v", err)
	}
	var n int
	projects, models, classes := map[string]bool{}, map[string]bool{}, map[string]bool{}
	bigGroupings := map[string]bool{}
	fallback, twoRounds, twoCritical := false, false, false
	for _, r := range rows {
		if !seeded.MatchString(r.Change) {
			continue
		}
		n++
		projects[r.Project] = true
		models[r.ImplementerModel] = true
		classes[r.Class] = true
		if r.Class == "big" {
			bigGroupings[r.Grouping] = true
		}
		fallback = fallback || r.Fallbacks == 1
		twoRounds = twoRounds || r.FixRounds == 2
		twoCritical = twoCritical || r.Critical == 2
		switch r.Change {
		case "kan-604-decisions":
			if r.Important != 1 {
				t.Errorf("run 4 Important = %d, want 1 (withdrawn excluded)", r.Important)
			}
		case "kan-611-decisions":
			if r.Minor != 0 {
				t.Errorf("run 11 Minor = %d, want 0 (withdrawn excluded)", r.Minor)
			}
		}
	}
	if n != 18 {
		t.Fatalf("seeded decisions rows = %d, want 18", n)
	}
	for _, p := range []string{"uitest-alpha", "uitest-beta"} {
		if !projects[p] {
			t.Errorf("no seeded row in project %s", p)
		}
	}
	for _, m := range []string{"skipped — inline", "skipped - inline", "inline"} {
		if !models[m] {
			t.Errorf("no seeded row with implementer %q", m)
		}
	}
	for _, c := range []string{"big", "regular", "small", ""} {
		if !classes[c] {
			t.Errorf("no seeded row with class %q", c)
		}
	}
	if !fallback || !twoRounds || !twoCritical {
		t.Errorf("fallback=%v fixRounds2=%v critical2=%v, want all true", fallback, twoRounds, twoCritical)
	}
	if !bigGroupings["free"] || !bigGroupings["static"] {
		t.Errorf("big groupings = %v, want free and static", bigGroupings)
	}

	pinned, err := st.Decisions(ctx, store.Period{
		From: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC),
	}, nil)
	if err != nil {
		t.Fatalf("Decisions over the pinned window: %v", err)
	}
	for _, r := range pinned {
		if seeded.MatchString(r.Change) {
			t.Errorf("seeded row %s is inside the pinned window", r.Change)
		}
	}
}

// TestSeedGuardRunsDrawEveryState pins the guard-runs fixture that draws
// every flow-health state (decision seed-guard-runs): thirteen guards in
// 30d, eight never fired, helpers left out, cannot-answer runs, a run-time
// leader, and runs outside 7d and outside 30d.
func TestSeedGuardRunsDrawEveryState(t *testing.T) {
	smoke.Heavy(t)
	dsn := newUitestDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	st, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()
	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	if err := seedFixture(ctx, st); err != nil {
		t.Fatalf("seedFixture: %v", err)
	}

	window := func(days int) store.Period {
		return store.Period{From: fixtureNow.Add(-time.Duration(days) * 24 * time.Hour), To: fixtureNow}
	}
	rows, err := st.GuardHealth(ctx, window(30), nil)
	if err != nil {
		t.Fatalf("GuardHealth: %v", err)
	}
	byGuard := map[string]store.GuardHealthRow{}
	var neverFired int
	var top store.GuardHealthRow
	for _, r := range rows {
		byGuard[r.Guard] = r
		if r.Fired == 0 {
			neverFired++
		}
		if r.TotalMs > top.TotalMs {
			top = r
		}
	}
	if len(rows) != 13 {
		t.Errorf("30d guards = %d, want 13: %+v", len(rows), rows)
	}
	for _, helper := range []string{"project-get", "run-verify-list"} {
		if _, ok := byGuard[helper]; ok {
			t.Errorf("30d rows carry the helper %s", helper)
		}
	}
	if neverFired != 8 {
		t.Errorf("never-fired guards = %d, want 8", neverFired)
	}
	if r := byGuard["check-references"]; r.Runs != 31 || r.Fired != 4 {
		t.Errorf("check-references = %+v, want Runs 31 Fired 4", r)
	}
	if r := byGuard["check-task-commit-fields"]; r.Runs != 25 || r.CannotAnswer != 6 {
		t.Errorf("check-task-commit-fields = %+v, want Runs 25 CannotAnswer 6", r)
	}
	if r := byGuard["check-guard-calls"]; r.Runs != 8 || r.CannotAnswer != 2 {
		t.Errorf("check-guard-calls = %+v, want Runs 8 CannotAnswer 2", r)
	}
	if top.Guard != "run-guard-tests" || top.TotalMs != 4200000 {
		t.Errorf("largest TotalMs = %+v, want run-guard-tests 4200000", top)
	}

	week, err := st.GuardHealth(ctx, window(7), nil)
	if err != nil {
		t.Fatalf("GuardHealth 7d: %v", err)
	}
	for _, r := range week {
		if r.Guard == "check-spec-reach" {
			t.Errorf("7d window carries check-spec-reach: %+v", r)
		}
	}
	quarter, err := st.GuardHealth(ctx, window(90), nil)
	if err != nil {
		t.Fatalf("GuardHealth 90d: %v", err)
	}
	if !slices.ContainsFunc(quarter, func(r store.GuardHealthRow) bool { return r.Guard == "check-ask-silence" }) {
		t.Error("90d window has no check-ask-silence row")
	}

	projects, err := st.GuardHealthProjects(ctx, window(30))
	if err != nil {
		t.Fatalf("GuardHealthProjects: %v", err)
	}
	want := []store.ProjectGuardRuns{{ProjectKey: "uitest-alpha", Runs: 205}, {ProjectKey: "uitest-beta", Runs: 18}}
	if !slices.Equal(projects, want) {
		t.Errorf("GuardHealthProjects = %+v, want %+v", projects, want)
	}
}
