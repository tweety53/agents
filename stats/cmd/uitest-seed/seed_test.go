package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	dsnutil "github.com/tweety53/agents/stats/internal/dsn"
	"github.com/tweety53/agents/stats/internal/records"
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

// seeded is the one fixture database every seed test reads: seedFixture
// writes it once, behind the pricing table flowd seeds before make
// ui-test-up runs the seeder (main.go), and no test writes to it, so each
// asserts on the same rows a per-test seed would have written.
var seeded struct {
	once   sync.Once
	dsn    string
	st     *store.Store
	err    error
	skip   error
	dbName string
}

// TestMain drops the shared fixture database once every test has run.
func TestMain(m *testing.M) {
	code := m.Run()
	if seeded.dbName != "" {
		if seeded.st != nil {
			seeded.st.Close()
		}
		if err := dropUitestDatabase(seeded.dbName); err != nil {
			fmt.Fprintf(os.Stderr, "drop test database %s: %v\n", seeded.dbName, err)
		}
	}
	os.Exit(code)
}

// seededStore returns the store on the shared fixture database, creating,
// migrating and seeding it on first use. The database's name ends in
// "_uitest", so requireUitestDatabase admits it; it is never the live "flow"
// database or the shared "flow_uitest" one. It skips cleanly when the
// compose stack is not reachable, as internal/storetest does.
func seededStore(t *testing.T) *store.Store {
	t.Helper()
	seeded.once.Do(seedOnce)
	if seeded.skip != nil {
		t.Skipf("flow-postgres compose stack not reachable: %v", seeded.skip)
	}
	if seeded.err != nil {
		t.Fatalf("seed fixture database: %v", seeded.err)
	}
	return seeded.st
}

func seedOnce() {
	ctx := context.Background()
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	adminPool, err := pgxpool.New(pingCtx, adminDSN())
	if err == nil {
		err = adminPool.Ping(pingCtx)
	}
	if err != nil {
		if adminPool != nil {
			adminPool.Close()
		}
		seeded.skip = err
		return
	}
	dbName := fmt.Sprintf("uitest_seed_test_%d_%d_uitest", os.Getpid(), time.Now().UnixNano())
	_, err = adminPool.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{dbName}.Sanitize())
	adminPool.Close()
	if err != nil {
		seeded.err = fmt.Errorf("create %s: %w", dbName, err)
		return
	}
	seeded.dbName = dbName
	seeded.dsn = testDSN(dbName)

	st, err := store.Open(ctx, seeded.dsn)
	if err != nil {
		seeded.err = fmt.Errorf("open store: %w", err)
		return
	}
	seeded.st = st
	if err := st.RunMigrations(ctx); err != nil {
		seeded.err = fmt.Errorf("run migrations: %w", err)
		return
	}
	if err := st.SeedPricing(ctx); err != nil {
		seeded.err = fmt.Errorf("seed pricing: %w", err)
		return
	}
	if err := seedFixture(ctx, st); err != nil {
		seeded.err = fmt.Errorf("seedFixture: %w", err)
	}
}

func dropUitestDatabase(dbName string) error {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, adminDSN())
	if err != nil {
		return err
	}
	defer pool.Close()
	_, err = pool.Exec(ctx, "DROP DATABASE IF EXISTS "+pgx.Identifier{dbName}.Sanitize()+" WITH (FORCE)")
	return err
}

// TestSeedPopulatesFixture proves seedFixture writes what
// the UI-test-stack requirement "The test stack starts populated"
// requires: two projects, a change in every one of the three pipeline
// states, and at least one stage run carrying token usage a cost view
// could render. It runs against the package's throwaway "..._uitest" database,
// never the live "flow" database or the shared "flow_uitest" one.
// smoke: core — the seeder a store change can break
func TestSeedPopulatesFixture(t *testing.T) {
	t.Parallel()
	st := seededStore(t)
	if err := requireUitestDatabase(seeded.dsn); err != nil {
		t.Fatalf("guard refused the test's own throwaway database: %v", err)
	}

	ctx := t.Context()

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
			wantChanges += 4 // seedInteractive's two sessions and seedSelfReviewPair's two changes, seeded beside fixtureData
		}
		if project.Key == "uitest-beta" {
			wantChanges += 2 // seedRunCosts' two, seeded beside fixtureData
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
}

// TestSeedInteractiveSessions pins seedInteractive (KAN-1061): the
// prototype's running session lists its eight requests in their seeded
// states with its open run carrying request 7, and the waiting session its
// three landed requests.
func TestSeedInteractiveSessions(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	st := seededStore(t)
	ctx := t.Context()

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
	t.Parallel()
	st := seededStore(t)
	ctx := t.Context()

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
	t.Parallel()
	st := seededStore(t)
	ctx := t.Context()

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
	t.Parallel()
	st := seededStore(t)
	ctx := t.Context()

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
		if r.Description != "" {
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
		t.Error("no slot carries a description")
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

// TestSeedGuardRunsDrawEveryState pins the guard-runs fixture that draws
// every flow-health state (decision seed-guard-runs): thirteen guards in
// 30d, eight never fired, helpers left out, cannot-answer runs, a run-time
// leader, and runs outside 7d and outside 30d.
func TestSeedGuardRunsDrawEveryState(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	st := seededStore(t)
	ctx := t.Context()

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
		if !r.Helper && r.TotalMs > top.TotalMs {
			top = r
		}
	}
	if len(rows) != 15 {
		t.Errorf("30d guards = %d, want 15: %+v", len(rows), rows)
	}
	// The helpers are listed, flagged, and never the slowest guard.
	for _, helper := range []string{"project-get", "run-verify-list"} {
		if r, ok := byGuard[helper]; !ok || !r.Helper {
			t.Errorf("30d rows lack the flagged helper %s: %+v", helper, r)
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

// TestSeedRunCostsFixture pins the run the run-costs captures draw: two
// FINISHED uitest-beta runs inside the thirty days before fixtureNow and
// outside the pinned window, the first holding a row of every lane the
// prototype draws, so its computed result fills every totals tile.
func TestSeedRunCostsFixture(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	st := seededStore(t)
	ctx := t.Context()

	pinnedFrom := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	for _, name := range []string{"kan-301-run-costs-a", "kan-302-run-costs-b"} {
		c, err := st.GetChange(ctx, "uitest-beta", name)
		if err != nil {
			t.Fatalf("get change %s: %v", name, err)
		}
		if c.State != store.StateFinished || !c.UpdatedAt.After(fixtureNow.Add(-30*24*time.Hour)) || !c.UpdatedAt.Before(pinnedFrom) {
			t.Errorf("%s = %s updated %s, want FINISHED inside the 30 days before %s and before %s", name, c.State, c.UpdatedAt, fixtureNow, pinnedFrom)
		}
	}

	saved, err := st.GetChangeFile(ctx, "uitest-beta", "kan-301-run-costs-a", "tasks.md")
	if err != nil {
		t.Fatalf("saved tasks.md: %v", err)
	}
	openers := records.TaskOpeners([]byte(saved))
	if !strings.HasPrefix(openers["15"], "verify") || !strings.HasPrefix(openers["16"], "visual-verify") {
		t.Errorf("saved openers = %v, want task 15 a verify round and 16 a visual-verify round", openers)
	}

	srv, err := api.New(config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}, st, st, st, st,
		slog.New(slog.NewTextHandler(io.Discard, nil)), api.WithRunTimings(st), api.WithChangeFiles(st), api.WithGuardRuns(st))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/api/v1/stats/run/uitest-beta/kan-301-run-costs-a")
	if err != nil {
		t.Fatalf("GET run: %v", err)
	}
	defer resp.Body.Close()
	type value struct {
		Value float64 `json:"value"`
	}
	var got struct {
		Totals struct {
			WallClockMs, AgentTimeMs, ParentCommandsMs, IdleMs, Cost, ParentSessionCost, VisualVerifyShare value
			Dispatches, NoTokens, ParentCommandRuns, UnclaimedGuardRuns                                    int
		} `json:"totals"`
		Lanes []struct {
			Label string `json:"label"`
			Kind  string `json:"kind"`
		} `json:"lanes"`
		Counts struct {
			Findings struct {
				Value struct{ Critical, Important, Minor int } `json:"value"`
			} `json:"findings"`
			FlakyRetries struct {
				Value struct{ Retries, Runs int } `json:"value"`
			} `json:"flakyRetries"`
			Departures struct {
				Value []struct {
					Round      string `json:"round"`
					Departures int    `json:"departures"`
				} `json:"value"`
			} `json:"departures"`
		} `json:"counts"`
		CostPerTask []struct {
			Task  string  `json:"task"`
			Group string  `json:"group"`
			Cost  float64 `json:"cost"`
		} `json:"costPerTask"`
		DataQuality []struct {
			ID     string   `json:"id"`
			Count  int      `json:"count"`
			Detail []string `json:"detail"`
		} `json:"dataQuality"`
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET run = %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode run: %v", err)
	}

	tt := got.Totals
	for name, v := range map[string]float64{
		"wallClockMs": tt.WallClockMs.Value, "agentTimeMs": tt.AgentTimeMs.Value, "parentCommandsMs": tt.ParentCommandsMs.Value,
		"idleMs": tt.IdleMs.Value, "cost": tt.Cost.Value, "parentSessionCost": tt.ParentSessionCost.Value,
		"visualVerifyShare": tt.VisualVerifyShare.Value, "dispatches": float64(tt.Dispatches), "noTokens": float64(tt.NoTokens),
		"parentCommandRuns": float64(tt.ParentCommandRuns), "unclaimedGuardRuns": float64(tt.UnclaimedGuardRuns),
	} {
		if v == 0 {
			t.Errorf("totals.%s = 0, want non-zero", name)
		}
	}

	var stages, categories []string
	for _, l := range got.Lanes {
		switch l.Kind {
		case "stage":
			stages = append(stages, l.Label)
		case "category":
			categories = append(categories, l.Label)
		}
	}
	if len(stages) < 4 {
		t.Errorf("stage lanes = %v, want at least 4", stages)
	}
	for _, want := range []string{"planner", "explorer", "planned implementer", "gated reviewer", "gated fixer", "verify list",
		"visual verifier", "in-run fix (verify)", "in-run fix (visual verify)", "panel reviewer", "merger", "panel fixer"} {
		if !slices.Contains(categories, want) {
			t.Errorf("category lanes = %v, missing %q", categories, want)
		}
	}
	for _, c := range got.CostPerTask {
		if c.Cost == 0 {
			t.Errorf("costPerTask %s cost = 0, want the dispatch rows priced", c.Task)
		}
	}
	if d := got.Counts.Departures.Value; len(d) != 2 || d[0].Departures != 3 {
		t.Errorf("departures = %+v, want the two saved visual-verify reports' lines, 3 then 0", d)
	}
	if !slices.ContainsFunc(got.CostPerTask, func(c struct {
		Task  string  `json:"task"`
		Group string  `json:"group"`
		Cost  float64 `json:"cost"`
	}) bool {
		return c.Group == "13+14"
	}) {
		t.Errorf("costPerTask = %+v, want a task split from group 13+14", got.CostPerTask)
	}
	if f := got.Counts.Findings.Value; f.Critical == 0 || f.Important == 0 || f.Minor == 0 {
		t.Errorf("findings = %+v, want each severity", f)
	}
	if f := got.Counts.FlakyRetries.Value; f.Retries == 0 || f.Runs == 0 {
		t.Errorf("flaky retries = %+v, want verify runs with a flaky count", f)
	}
	dq := map[string]int{}
	for _, d := range got.DataQuality {
		dq[d.ID] = d.Count
		if d.ID == "no-tokens" && (len(d.Detail) == 0 || strings.HasSuffix(d.Detail[0], "not measured")) {
			t.Errorf("no-tokens detail = %v, want a stated reason", d.Detail)
		}
	}
	if dq["superseded-stage-runs"] == 0 || dq["no-tokens"] == 0 {
		t.Errorf("data quality = %v, want a superseded stage run and a no-tokens row", dq)
	}
}
