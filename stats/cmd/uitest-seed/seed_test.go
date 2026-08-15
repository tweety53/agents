package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/store"
)

// adminDSN is the DSN used to create and drop the throwaway database this
// test seeds into. It points at the myflow-postgres compose stack on host
// port 5433 -- the same stack internal/store's own tests use, and the same
// DSN config.DefaultDSN already names -- and can be overridden for
// environments that run Postgres elsewhere.
func adminDSN() string {
	if v := os.Getenv("MYFLOW_STATS_ADMIN_DSN"); v != "" {
		return v
	}
	return config.DefaultDSN
}

func testDSN(dbName string) string {
	return fmt.Sprintf("postgres://myflow:myflow@localhost:5433/%s?sslmode=disable", dbName)
}

// newUitestDatabase creates a uniquely-named database whose name ends in
// "_uitest" -- so requireUitestDatabase admits it -- and registers a
// cleanup that drops it once the test finishes. It never touches the live
// "myflow" database or the shared "myflow_uitest" one: both are named
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
		t.Skipf("myflow-postgres compose stack not reachable: %v", err)
	}
	if err := adminPool.Ping(ctx); err != nil {
		adminPool.Close()
		t.Skipf("myflow-postgres compose stack not reachable: %v", err)
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
// specs/myflow-ui-test-stack/spec.md's "The test stack starts populated"
// requires: two projects, a change in every one of the three pipeline
// states, and at least one stage run carrying token usage a cost view
// could render. It runs against its own throwaway "..._uitest" database,
// never the live "myflow" database or the shared "myflow_uitest" one.
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
		if len(changes) != len(project.Changes) {
			t.Fatalf("project %s: got %d changes, want %d", project.Key, len(changes), len(project.Changes))
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
