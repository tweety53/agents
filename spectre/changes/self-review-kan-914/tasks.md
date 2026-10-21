# self-review-kan-914

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close KAN-914 — a workspace daemon's fresh `flow_<id>` database refuses record writes
until something seeds its `projects` row.

**Architecture:** a `Store.EnsureProject` method over the existing `ensureProject` bootstrap, and
one startup branch in flowd that calls it with `fallback.ProjectKey(".")` when the DSN names a
non-UI-test workspace database.

**Tech Stack:** Go (`stats/cmd/flowd`, `stats/internal/store`), pgx v5.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-21-self-review-kan-914-design.md` in this worktree.

## Global Constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-914`, branch `self-review-kan-914`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- Run `cd stats && gofmt -w .` on Go changes before committing; `gofmt -l .` and `go vet ./...` stay clean (both after `make -s web-dist`). Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container, nor drop the `flow` database; the store tests create and drop their own databases in it, as they already do.

## Review Focus

- The dev daemon's DSN (`config.DefaultDSN`, database `flow`) never seeds — its startup path is unchanged.
- A `flow_<id>_uitest` DSN never seeds: `cmd/uitest-seed` owns that database's rows.
- `EnsureProject` on an existing row keeps its first path, and with an empty path creates nothing (`ErrInvalidMainCheckoutPath`).
- A daemon started outside a git checkout still starts, logging one `Warn` line.

- [ ] 1. KAN-914 — seed a workspace daemon's project row at startup
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Add the store test below
    to `stats/internal/store/projects_test.go` (imports `reflect` and
    `github.com/tweety53/agents/stats/internal/records`) with its `EnsureProject` calls commented
    out, and run
    `cd stats && make -s web-dist && go test ./internal/store/ -run TestEnsureProjectLetsARecordWriteThrough -count=1`:
    the first `RecordGuardRun` returns `ErrProjectNotFound` and the second fails the same way.
    Then restore the calls.

```go verified:added to projects_test.go with EnsureProject below at b76c2cf1; go test ./internal/store/ -run 'TestEnsureProject|TestProjectRoots' -count=1 passed
// TestEnsureProjectLetsARecordWriteThrough pins KAN-914: a fresh database
// refuses a record write for a project it has no row for, and
// EnsureProject's row lets it through; a second call and an existing row
// are left as they are, and an empty path never creates one.
func TestEnsureProjectLetsARecordWriteThrough(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	run := records.GuardRun{Guard: "check-x", Worktree: "/ws/alpha", Outcome: "clear"}

	if _, err := st.RecordGuardRun(ctx, "ws-alpha-key", run); !errors.Is(err, store.ErrProjectNotFound) {
		t.Fatalf("RecordGuardRun before EnsureProject: err = %v, want ErrProjectNotFound", err)
	}
	if err := st.EnsureProject(ctx, "ws-alpha-key", ""); !errors.Is(err, store.ErrInvalidMainCheckoutPath) {
		t.Fatalf("EnsureProject with an empty path: err = %v, want ErrInvalidMainCheckoutPath", err)
	}
	for range 2 {
		if err := st.EnsureProject(ctx, "ws-alpha-key", "/ws/alpha"); err != nil {
			t.Fatalf("EnsureProject: %v", err)
		}
	}
	if err := st.EnsureProject(ctx, "ws-alpha-key", "/ws/other"); err != nil {
		t.Fatalf("EnsureProject on an existing row: %v", err)
	}
	if _, err := st.RecordGuardRun(ctx, "ws-alpha-key", run); err != nil {
		t.Fatalf("RecordGuardRun after EnsureProject: %v", err)
	}
	roots, err := st.ProjectRoots(ctx)
	if err != nil {
		t.Fatalf("ProjectRoots: %v", err)
	}
	want := []store.ProjectRoot{{ProjectKey: "ws-alpha-key", MainCheckoutPath: "/ws/alpha"}}
	if !reflect.DeepEqual(roots, want) {
		t.Fatalf("ProjectRoots = %+v, want %+v", roots, want)
	}
}
```

  - [ ] **Step 2: Implement `func (s *Store) EnsureProject(ctx context.Context, projectKey, mainCheckoutPath string) error`**
    in `stats/internal/store/projects.go`, with a doc comment naming it the bootstrap
    `PutChange` and `insertPlanStageRun` run inside their own writes, for a caller with no
    change to write:

```go verified:compiled and passed the Step 1 test at b76c2cf1 (pgx v5.10.0 has BeginFunc)
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		return ensureProject(ctx, tx, projectKey, mainCheckoutPath)
	})
```

    Run the Step 1 command again: it passes.
  - [ ] **Step 3: Write the predicate test** `TestSeedsWorkspaceProjectOnlyIntoAWorkspaceDatabase`
    in `stats/cmd/flowd/wiring_test.go`, beside `TestReplaysSharedJournalOnlyIntoTheFlowDatabase`
    and shaped like it, a table over `seedsWorkspaceProject(dsn)`:

```go verified:passed against the Step 4 predicate at b76c2cf1 with go test ./cmd/flowd/ -run TestSeedsWorkspaceProject -count=1
	cases := map[string]bool{
		config.DefaultDSN: false,
		"postgres://flow:flow@localhost:5433/flow_kan_9_x?sslmode=disable":   true,
		"postgres://flow:flow@localhost:5433/flow_ui_uitest?sslmode=disable": false,
		"postgres://flow:flow@localhost:5433/flowd?sslmode=disable":          false,
		"host=localhost port=5433 dbname=flow_kan_9_x user=flow":             false,
	}
```

  - [ ] **Step 4: Implement** in `stats/cmd/flowd/main.go`
    `func seedsWorkspaceProject(dsn string) bool` beside `replaysSharedJournal` — true when
    `url.Parse(dsn)` succeeds and its path starts `/flow_` and does not end `_uitest` — and, in
    `run` after `st.SeedPricing` and before the SPA wiring, a block that, when
    `seedsWorkspaceProject(cfg.DSN)`, calls `fallback.ProjectKey(".")` then
    `st.EnsureProject(ctx, key, mainCheckout)`, logging
    `logger.Warn("flowd did not seed the workspace project row", …)` with the error (and the key,
    once known) on either failure and continuing. Its comment cites KAN-914 and names the
    working directory as the checkout `originRemote` already reads. Run
    `cd stats && make -s web-dist && go test ./cmd/flowd/ -run 'TestSeedsWorkspaceProject|TestReplaysSharedJournal' -count=1`:
    passes.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -w . && make -s web-dist && go vet ./... && gofmt -l .`
    exits 0 and prints nothing; rerun the Step 2 and Step 4 test commands. Commit.

**Build:** green
**Files:** `stats/internal/store/projects.go`, `stats/internal/store/projects_test.go`, `stats/cmd/flowd/main.go`, `stats/cmd/flowd/wiring_test.go`
**Tests:** `TestEnsureProjectLetsARecordWriteThrough`, `TestSeedsWorkspaceProjectOnlyIntoAWorkspaceDatabase`
**Regression:** `TestEnsureProjectLetsARecordWriteThrough` — reverting removes `EnsureProject` and the test no longer compiles; without the seed the second `RecordGuardRun` returns `ErrProjectNotFound`. `TestSeedsWorkspaceProjectOnlyIntoAWorkspaceDatabase` — reverting removes the predicate and the startup seed with it.
**Baseline:** before=11 after=13
<!-- measured: cat stats/internal/store/projects_test.go stats/cmd/flowd/wiring_test.go | grep -c '^func Test' @ b76c2cf1 -->
**Commit:** fix(flowd): seed a workspace daemon's project row at startup
**After:** none

**Decision:** seed-at-workspace-daemon-startup
**Decision:** seed-scope-workspace-only
**Decision:** seed-failure-logged
