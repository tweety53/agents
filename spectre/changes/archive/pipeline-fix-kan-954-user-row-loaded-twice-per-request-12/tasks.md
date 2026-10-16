# pipeline-fix-kan-954-user-row-loaded-twice-per-request-12

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** store-backed tests in `stats/` stop timing out under `go test ./...`: each test database
is a copy of one migrated template per package process, never a fresh database migrated per test.

**Architecture:** a new test-support package `stats/internal/storetest` owns the template lifecycle;
the six per-package helpers forward to it and each package's `TestMain` calls `storetest.Main`.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:`
  trailer; never write `spectre/changes/`.
- Each task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; both were applied to this worktree at e777ee03, in order, and checked there. A diff
  that does not apply is reported, never hand-merged by guesswork.
- Store tests create their own databases on `flow-postgres` (port 5433); never point a test or a
  command at the live `flowd` database, and never drop a database this change's code did not create.
- No deadline is raised anywhere (`design.md`, `deadlines-unchanged`).
- No new inline suppressions; `cd stats && gofmt -l .` prints nothing and `go vet ./...` passes.

## Review Focus

- `CREATE DATABASE … TEMPLATE` never runs while a session is connected to the template: the build
  closes its store before `sync.Once` returns.
- `dropLeftoverDatabases`' pattern matches only this process's `flow_test_<pid>_<n>` names — never
  the template, another process's databases, or the live `flow` database.
- `internal/harvest`'s helper still seeds pricing after opening the store.

- [x] 1. Copy store test databases from one migrated template
  - [x] **Step 1: Reproduce before touching anything; do not guess.** The flake needs the whole
    suite's cross-package load and was reproduced at plan time (`design.md`, `## Context`); here
    run `cd stats && go test -count=1 ./internal/store/ -run 'TestProjectRootsListsRegisteredCheckouts'`
    on the untouched tree and confirm it passes alone — the defect is load, not logic.
  - [x] **Step 2: Implement.** Save the diff below to `/tmp/pf12-task1.diff` and apply it
    (`git apply /tmp/pf12-task1.diff`):

````diff verified:applied with git apply to this worktree at e777ee03; go vet ./... clean with this diff alone; go test -count=1 ./internal/storetest/ ./internal/store/ green; TestMigratedDatabasesAreCopiesOfOneTemplate fails when NewMigratedDatabase migrates per database
diff --git a/stats/internal/store/testsupport_test.go b/stats/internal/store/testsupport_test.go
index 1e951300..959e0d62 100644
--- a/stats/internal/store/testsupport_test.go
+++ b/stats/internal/store/testsupport_test.go
@@ -1,126 +1,29 @@
 package store_test
 
 import (
-	"context"
-	"fmt"
-	dsnutil "github.com/tweety53/agents/stats/internal/dsn"
 	"os"
 	"testing"
-	"time"
-
-	"github.com/jackc/pgx/v5"
-	"github.com/jackc/pgx/v5/pgxpool"
 
 	"github.com/tweety53/agents/stats/internal/store"
+	"github.com/tweety53/agents/stats/internal/storetest"
 )
 
-// adminDSN is the DSN used to create and drop per-test databases. It points
-// at the flow-postgres compose stack from task 1, on host port 5433, and
-// can be overridden for environments that run Postgres elsewhere.
-func adminDSN() string {
-	if v := os.Getenv("FLOW_STATS_ADMIN_DSN"); v != "" {
-		return v
-	}
-	return "postgres://flow:flow@localhost:5433/flow?sslmode=disable"
-}
-
-// testDSN returns the DSN for a database created by newTestDatabase.
-//
-// IT DERIVES FROM adminDSN RATHER THAN HARDCODING THE SAME CREDENTIALS AGAIN,
-// so that FLOW_STATS_ADMIN_DSN moves BOTH connections together. Hardcoding them
-// here made the override only half-work: the admin connection followed the
-// environment while the per-test connection stayed pinned to the literal, so
-// pointing the suite at a Postgres with different credentials failed on every
-// test while the admin step succeeded.
-//
-// That is not hypothetical. During the flow rename this literal and adminDSN's
-// were updated together to `flow:flow@.../flow`, which is what a fresh compose
-// stack creates -- but the operator's own container is not renamed until the
-// documented cutover step. With no working override, the whole suite reached an
-// unreachable Postgres and SKIPPED: `internal/store` reported `ok` while running
-// 4 tests and skipping 155. A green package that ran almost nothing is the exact
-// failure this repository's REPRODUCE, DON'T READ rule exists to catch.
-func testDSN(dbName string) string {
-	out, err := dsnutil.ForDatabase(adminDSN(), dbName)
-	if err != nil {
-		// Panic rather than return a best-effort string. This helper's
-		// predecessor answered confidently when it could not do the job and
-		// silently handed back a corrupted DSN; a test that then connects
-		// reports whatever that connection says, which is the wrong question
-		// answered convincingly.
-		panic(fmt.Sprintf("deriving a per-test DSN: %v", err))
-	}
-	return out
-}
+// TestMain drops the package's template database once every test has run.
+func TestMain(m *testing.M) { os.Exit(storetest.Main(m)) }
 
-// newTestDatabase creates a uniquely-named, empty (unmigrated) database
-// against the compose stack's PostgreSQL, and registers a cleanup that
-// drops it once the test finishes. It skips cleanly, with a clear message,
-// when the stack is not reachable.
+// newTestDatabase creates a uniquely-named, empty (unmigrated) database and
+// registers its drop as test cleanup. It skips cleanly when the stack is not
+// reachable.
 func newTestDatabase(t *testing.T) string {
 	t.Helper()
-
-	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
-	defer cancel()
-
-	adminPool, err := pgxpool.New(ctx, adminDSN())
-	if err != nil {
-		t.Skipf("flow-postgres compose stack not reachable: %v", err)
-	}
-	if err := adminPool.Ping(ctx); err != nil {
-		adminPool.Close()
-		t.Skipf("flow-postgres compose stack not reachable: %v", err)
-	}
-
-	dbName := fmt.Sprintf("flow_test_%d_%d", os.Getpid(), time.Now().UnixNano())
-	ident := pgx.Identifier{dbName}.Sanitize()
-	if _, err := adminPool.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
-		adminPool.Close()
-		t.Fatalf("create test database %s: %v", dbName, err)
-	}
-	adminPool.Close()
-
-	t.Cleanup(func() {
-		dropCtx, dropCancel := context.WithTimeout(context.Background(), 3*time.Second)
-		defer dropCancel()
-
-		dropPool, err := pgxpool.New(dropCtx, adminDSN())
-		if err != nil {
-			t.Logf("drop test database %s: reconnect failed: %v", dbName, err)
-			return
-		}
-		defer dropPool.Close()
-
-		if _, err := dropPool.Exec(dropCtx, "DROP DATABASE IF EXISTS "+ident+" WITH (FORCE)"); err != nil {
-			t.Logf("drop test database %s: %v", dbName, err)
-		}
-	})
-
-	return testDSN(dbName)
+	return storetest.NewDatabase(t)
 }
 
-// newTestStore creates a uniquely-named database, applies every migration
-// to it, and registers a cleanup that closes the store and drops the
-// database once the test finishes. It skips cleanly, with a clear message,
-// when the stack is not reachable.
+// newTestStore opens a store on a throwaway, migrated database -- a copy of
+// the package's one migrated template (internal/storetest) -- and registers
+// its Close and the database's drop as test cleanup. It skips cleanly when
+// the stack is not reachable.
 func newTestStore(t *testing.T) *store.Store {
 	t.Helper()
-
-	dsn := newTestDatabase(t)
-
-	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
-	defer cancel()
-
-	st, err := store.Open(ctx, dsn)
-	if err != nil {
-		t.Fatalf("open test store: %v", err)
-	}
-	if err := st.RunMigrations(ctx); err != nil {
-		st.Close()
-		t.Fatalf("run migrations: %v", err)
-	}
-
-	t.Cleanup(st.Close)
-
-	return st
+	return storetest.NewStore(t)
 }
diff --git a/stats/internal/storetest/storetest.go b/stats/internal/storetest/storetest.go
new file mode 100644
index 00000000..b4702ce1
--- /dev/null
+++ b/stats/internal/storetest/storetest.go
@@ -0,0 +1,242 @@
+// Package storetest gives tests a throwaway, migrated database on the
+// flow-postgres compose stack. It is imported by _test.go files only.
+//
+// A migrated database is a copy of one template database the package
+// process migrates once, never a fresh database migrated per test. Migrating
+// per test was the contention the suite flaked on: every test of every
+// package under `go test ./...` ran all of internal/store's migrations --
+// one transaction and one commit flush each -- against the one shared
+// server, and under that load a 3-second setup deadline ran out mid
+// migration ("run migrations: ... timeout: context deadline exceeded").
+// CREATE DATABASE ... TEMPLATE copies the migrated files instead, so a test
+// pays one database copy and the server sees one migration run per
+// package process.
+package storetest
+
+import (
+	"context"
+	"fmt"
+	"os"
+	"sync"
+	"testing"
+	"time"
+
+	"github.com/jackc/pgx/v5"
+	"github.com/jackc/pgx/v5/pgxpool"
+
+	dsnutil "github.com/tweety53/agents/stats/internal/dsn"
+	"github.com/tweety53/agents/stats/internal/store"
+)
+
+// setupTimeout bounds each step of a test's database setup. It is the
+// reachability bound the per-package helpers this package replaces each
+// carried; it is not raised here.
+const setupTimeout = 3 * time.Second
+
+// AdminDSN is the DSN used to create and drop test databases: the
+// flow-postgres compose stack on host port 5433, overridable through
+// FLOW_STATS_ADMIN_DSN for environments that run Postgres elsewhere.
+func AdminDSN() string {
+	if v := os.Getenv("FLOW_STATS_ADMIN_DSN"); v != "" {
+		return v
+	}
+	return "postgres://flow:flow@localhost:5433/flow?sslmode=disable"
+}
+
+// DSN returns AdminDSN with its database replaced by dbName, so
+// FLOW_STATS_ADMIN_DSN moves the admin and the per-test connections
+// together. It panics rather than hand back a corrupted DSN: a test that
+// then connects reports whatever that connection says.
+func DSN(dbName string) string {
+	out, err := dsnutil.ForDatabase(AdminDSN(), dbName)
+	if err != nil {
+		panic(fmt.Sprintf("deriving a per-test DSN: %v", err))
+	}
+	return out
+}
+
+var (
+	templateOnce sync.Once
+	templateName string
+	templateErr  error
+	unreachable  error
+)
+
+// Main runs the package's tests, then drops the template, if a test built
+// one, and every test database this process created whose cleanup drop
+// failed under load and would otherwise outlive the run. Every
+// package calling NewStore, NewMigratedDatabase or NewDatabase calls it from
+// TestMain: `os.Exit(storetest.Main(m))`.
+func Main(m *testing.M) int {
+	code := m.Run()
+	ctx, cancel := context.WithTimeout(context.Background(), setupTimeout)
+	defer cancel()
+	if templateName != "" {
+		if err := dropDatabase(ctx, templateName); err != nil {
+			fmt.Fprintf(os.Stderr, "storetest: drop template database %s: %v\n", templateName, err)
+		}
+	}
+	if err := dropLeftoverDatabases(ctx); err != nil {
+		fmt.Fprintf(os.Stderr, "storetest: drop leftover test databases: %v\n", err)
+	}
+	return code
+}
+
+// dropLeftoverDatabases drops every test database this process created
+// that still exists. It runs once no test is left to use one. An unreachable
+// stack means nothing was created, so it is not an error.
+func dropLeftoverDatabases(ctx context.Context) error {
+	admin, err := connectAdmin(ctx)
+	if err != nil {
+		return nil
+	}
+	defer admin.Close()
+
+	rows, err := admin.Query(ctx, "SELECT datname FROM pg_database WHERE datname ~ $1",
+		fmt.Sprintf(`^flow_test_%d_[0-9]+$`, os.Getpid()))
+	if err != nil {
+		return err
+	}
+	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
+	if err != nil {
+		return err
+	}
+	for _, name := range names {
+		if _, err := admin.Exec(ctx, "DROP DATABASE IF EXISTS "+pgx.Identifier{name}.Sanitize()+" WITH (FORCE)"); err != nil {
+			return fmt.Errorf("drop %s: %w", name, err)
+		}
+	}
+	return nil
+}
+
+// NewDatabase creates an empty, unmigrated database and registers its drop
+// as test cleanup. It skips the test when the stack is not reachable.
+func NewDatabase(t testing.TB) string {
+	t.Helper()
+	return createDatabase(t, "")
+}
+
+// NewMigratedDatabase creates a database carrying every embedded migration,
+// copied from the package's template, and registers its drop as test
+// cleanup. It skips the test when the stack is not reachable.
+func NewMigratedDatabase(t testing.TB) string {
+	t.Helper()
+	templateOnce.Do(buildTemplate)
+	if unreachable != nil {
+		t.Skipf("flow-postgres compose stack not reachable: %v", unreachable)
+	}
+	if templateErr != nil {
+		t.Fatalf("build template database: %v", templateErr)
+	}
+	return createDatabase(t, templateName)
+}
+
+// NewStore opens a store on a NewMigratedDatabase and registers its Close
+// as test cleanup.
+func NewStore(t testing.TB) *store.Store {
+	t.Helper()
+	dsn := NewMigratedDatabase(t)
+
+	ctx, cancel := context.WithTimeout(context.Background(), setupTimeout)
+	defer cancel()
+
+	st, err := store.Open(ctx, dsn)
+	if err != nil {
+		t.Fatalf("open test store: %v", err)
+	}
+	t.Cleanup(st.Close)
+	return st
+}
+
+// buildTemplate creates and migrates the package's template database, then
+// closes every connection to it: CREATE DATABASE ... TEMPLATE refuses a
+// template another session is connected to.
+func buildTemplate() {
+	ctx, cancel := context.WithTimeout(context.Background(), setupTimeout)
+	defer cancel()
+
+	admin, err := connectAdmin(ctx)
+	if err != nil {
+		unreachable = err
+		return
+	}
+	name := uniqueName("flow_template")
+	_, err = admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize())
+	admin.Close()
+	if err != nil {
+		templateErr = fmt.Errorf("create %s: %w", name, err)
+		return
+	}
+	templateName = name
+
+	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), setupTimeout)
+	defer migrateCancel()
+
+	st, err := store.Open(migrateCtx, DSN(name))
+	if err != nil {
+		templateErr = fmt.Errorf("open %s: %w", name, err)
+		return
+	}
+	defer st.Close()
+	if err := st.RunMigrations(migrateCtx); err != nil {
+		templateErr = fmt.Errorf("migrate %s: %w", name, err)
+	}
+}
+
+func createDatabase(t testing.TB, template string) string {
+	t.Helper()
+
+	ctx, cancel := context.WithTimeout(context.Background(), setupTimeout)
+	defer cancel()
+
+	admin, err := connectAdmin(ctx)
+	if err != nil {
+		t.Skipf("flow-postgres compose stack not reachable: %v", err)
+	}
+	defer admin.Close()
+
+	dbName := uniqueName("flow_test")
+	stmt := "CREATE DATABASE " + pgx.Identifier{dbName}.Sanitize()
+	if template != "" {
+		stmt += " TEMPLATE " + pgx.Identifier{template}.Sanitize()
+	}
+	if _, err := admin.Exec(ctx, stmt); err != nil {
+		t.Fatalf("create test database %s: %v", dbName, err)
+	}
+
+	t.Cleanup(func() {
+		dropCtx, dropCancel := context.WithTimeout(context.Background(), setupTimeout)
+		defer dropCancel()
+		if err := dropDatabase(dropCtx, dbName); err != nil {
+			t.Logf("drop test database %s: %v", dbName, err)
+		}
+	})
+
+	return DSN(dbName)
+}
+
+func connectAdmin(ctx context.Context) (*pgxpool.Pool, error) {
+	pool, err := pgxpool.New(ctx, AdminDSN())
+	if err != nil {
+		return nil, err
+	}
+	if err := pool.Ping(ctx); err != nil {
+		pool.Close()
+		return nil, err
+	}
+	return pool, nil
+}
+
+func dropDatabase(ctx context.Context, dbName string) error {
+	admin, err := connectAdmin(ctx)
+	if err != nil {
+		return fmt.Errorf("reconnect: %w", err)
+	}
+	defer admin.Close()
+	_, err = admin.Exec(ctx, "DROP DATABASE IF EXISTS "+pgx.Identifier{dbName}.Sanitize()+" WITH (FORCE)")
+	return err
+}
+
+func uniqueName(prefix string) string {
+	return fmt.Sprintf("%s_%d_%d", prefix, os.Getpid(), time.Now().UnixNano())
+}
diff --git a/stats/internal/storetest/storetest_internal_test.go b/stats/internal/storetest/storetest_internal_test.go
new file mode 100644
index 00000000..f69b8e08
--- /dev/null
+++ b/stats/internal/storetest/storetest_internal_test.go
@@ -0,0 +1,39 @@
+package storetest
+
+import (
+	"context"
+	"testing"
+
+	"github.com/jackc/pgx/v5"
+)
+
+// TestDropLeftoverDatabasesRemovesALeftover pins Main's sweep: a test database
+// whose cleanup never dropped it -- the drop timed out under load -- is
+// gone once dropLeftoverDatabases has run.
+func TestDropLeftoverDatabasesRemovesALeftover(t *testing.T) {
+	ctx, cancel := context.WithTimeout(context.Background(), setupTimeout)
+	defer cancel()
+
+	admin, err := connectAdmin(ctx)
+	if err != nil {
+		t.Skipf("flow-postgres compose stack not reachable: %v", err)
+	}
+	defer admin.Close()
+
+	leftover := uniqueName("flow_test")
+	if _, err := admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{leftover}.Sanitize()); err != nil {
+		t.Fatalf("create leftover database: %v", err)
+	}
+
+	if err := dropLeftoverDatabases(ctx); err != nil {
+		t.Fatalf("dropLeftoverDatabases: %v", err)
+	}
+
+	var exists bool
+	if err := admin.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", leftover).Scan(&exists); err != nil {
+		t.Fatalf("look up leftover database: %v", err)
+	}
+	if exists {
+		t.Fatalf("database %s outlived dropLeftoverDatabases", leftover)
+	}
+}
diff --git a/stats/internal/storetest/storetest_test.go b/stats/internal/storetest/storetest_test.go
new file mode 100644
index 00000000..fb4d97e3
--- /dev/null
+++ b/stats/internal/storetest/storetest_test.go
@@ -0,0 +1,69 @@
+package storetest_test
+
+import (
+	"context"
+	"os"
+	"testing"
+	"time"
+
+	"github.com/jackc/pgx/v5"
+
+	"github.com/tweety53/agents/stats/internal/store"
+	"github.com/tweety53/agents/stats/internal/storetest"
+)
+
+func TestMain(m *testing.M) { os.Exit(storetest.Main(m)) }
+
+// TestMigratedDatabasesAreCopiesOfOneTemplate pins the mechanism the suite's
+// setup rests on: two migrated databases carry every embedded migration
+// with identical applied_at stamps, which only a copy of one migrated
+// template produces. A database migrated per test stamps each migration
+// with its own transaction's now(), so the stamps differ and this fails.
+func TestMigratedDatabasesAreCopiesOfOneTemplate(t *testing.T) {
+	want, err := store.EmbeddedMigrationCount()
+	if err != nil {
+		t.Fatal(err)
+	}
+	first := migrationStamps(t, storetest.NewMigratedDatabase(t))
+	second := migrationStamps(t, storetest.NewMigratedDatabase(t))
+
+	if len(first) != want {
+		t.Fatalf("migrated database records %d migrations, want %d", len(first), want)
+	}
+	for name, stamp := range first {
+		if !second[name].Equal(stamp) {
+			t.Fatalf("migration %s applied at %v in one database and %v in the other: each was migrated on its own, not copied from the template",
+				name, stamp, second[name])
+		}
+	}
+}
+
+func migrationStamps(t *testing.T, dsn string) map[string]time.Time {
+	t.Helper()
+	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
+	defer cancel()
+
+	conn, err := pgx.Connect(ctx, dsn)
+	if err != nil {
+		t.Fatalf("connect: %v", err)
+	}
+	defer conn.Close(context.Background())
+
+	rows, err := conn.Query(ctx, "SELECT filename, applied_at FROM schema_migrations")
+	if err != nil {
+		t.Fatalf("read schema_migrations: %v", err)
+	}
+	stamps := make(map[string]time.Time)
+	for rows.Next() {
+		var name string
+		var at time.Time
+		if err := rows.Scan(&name, &at); err != nil {
+			t.Fatalf("scan schema_migrations: %v", err)
+		}
+		stamps[name] = at
+	}
+	if err := rows.Err(); err != nil {
+		t.Fatalf("read schema_migrations: %v", err)
+	}
+	return stamps
+}
````

  - [x] **Step 3: Prove the seam.** In `stats/internal/storetest/storetest.go`, temporarily replace
    `NewMigratedDatabase`'s `return createDatabase(t, templateName)` with an empty
    `createDatabase(t, "")` followed by `store.Open` + `RunMigrations` on it, run
    `cd stats && go test -count=1 -run TestMigratedDatabasesAreCopiesOfOneTemplate ./internal/storetest/`
    and see it FAIL with `applied at … in one database and … in the other`; restore the file
    (`git diff --stat` shows only the diff's own changes) and see it pass.
  - [x] **Step 4: Verify.** `cd stats && gofmt -l ./internal/storetest ./internal/store && go vet ./... && go test -count=1 ./internal/storetest/ && go test -count=1 ./internal/store/ -run 'TestProjectRootsListsRegisteredCheckouts|TestConcurrentBeginStageForOneSessionLeavesOneOpenRun|TestPriceFreezesCostAndVersion'`
    — gofmt prints nothing, the rest exit 0.
  - [x] **Step 5: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/internal/storetest/storetest.go`, `stats/internal/storetest/storetest_test.go`, `stats/internal/storetest/storetest_internal_test.go`, `stats/internal/store/testsupport_test.go`
**Tests:** `TestMigratedDatabasesAreCopiesOfOneTemplate`, `TestDropLeftoverDatabasesRemovesALeftover`
**Regression:** reverting the commit removes the template — `TestMigratedDatabasesAreCopiesOfOneTemplate` (with `NewMigratedDatabase` migrating per database) fails with `applied at … in one database and … in the other`, and `internal/store` setups go back to migrating per test and timing out under whole-suite load; without the sweep, `TestDropLeftoverDatabasesRemovesALeftover` fails with `outlived dropLeftoverDatabases`.
**Baseline:** before=0 after=2
<!-- measured: cat stats/internal/storetest/*_test.go 2>/dev/null | grep -E '^func Test' | grep -vc '^func TestMain(' (before at e777ee03; after with this task's diff applied) -->
**Commit:** fix(storetest): copy store test databases from one migrated template
**After:** none

**Decision:** template-per-process
**Decision:** deadlines-unchanged
**Decision:** storetest-package
**Decision:** main-sweeps-leftovers

- [x] 2. Switch the other store-backed packages to the template
  - [x] **Step 1: Implement.** Save the diff below to `/tmp/pf12-task2.diff` and apply it
    (`git apply /tmp/pf12-task2.diff`):

````diff verified:applied with git apply to this worktree at e777ee03 on top of Task 1's diff; go vet clean; the whole stats suite green 2 of 2 plain and 2 of 2 with -race at load averages 17–24
diff --git a/stats/cmd/flow/journal_test.go b/stats/cmd/flow/journal_test.go
index 00c6b09c..484e0f53 100644
--- a/stats/cmd/flow/journal_test.go
+++ b/stats/cmd/flow/journal_test.go
@@ -3,88 +3,28 @@ package main
 import (
 	"bytes"
 	"context"
-	"fmt"
-	dsnutil "github.com/tweety53/agents/stats/internal/dsn"
 	"os"
 	"path/filepath"
 	"testing"
 	"time"
 
-	"github.com/jackc/pgx/v5"
-	"github.com/jackc/pgx/v5/pgxpool"
-
 	"github.com/tweety53/agents/stats/internal/fallback"
 	"github.com/tweety53/agents/stats/internal/store"
+	"github.com/tweety53/agents/stats/internal/storetest"
 )
 
-// adminDSN is the DSN used to create and drop per-test databases -- the
-// same dedicated flow-postgres compose stack every other package's tests
-// use.
-func adminDSN() string {
-	if v := os.Getenv("FLOW_STATS_ADMIN_DSN"); v != "" {
-		return v
-	}
-	return "postgres://flow:flow@localhost:5433/flow?sslmode=disable"
-}
+// TestMain drops the package's template database once every test has run.
+func TestMain(m *testing.M) { os.Exit(storetest.Main(m)) }
 
-// newTestStoreDSN creates a uniquely-named, migrated database against the
-// compose stack and returns its DSN -- runJournalFlush opens its own store
-// from a DSN string (it is the one CLI command that talks to the store
-// directly; see journal.go's own doc comment), so this test needs the
-// connection string itself, not a *store.Store. It skips cleanly when the
-// stack is not reachable and registers a cleanup that drops the database.
+// newTestStoreDSN returns the DSN of a throwaway, migrated database -- a
+// copy of the package's one migrated template (internal/storetest) --
+// because runJournalFlush opens its own store from a DSN string (it is the
+// one CLI command that talks to the store directly; see journal.go's own doc
+// comment). It skips cleanly when the stack is not reachable and registers
+// the database's drop as test cleanup.
 func newTestStoreDSN(t *testing.T) string {
 	t.Helper()
-
-	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
-	defer cancel()
-
-	adminPool, err := pgxpool.New(ctx, adminDSN())
-	if err != nil {
-		t.Skipf("flow-postgres compose stack not reachable: %v", err)
-	}
-	if err := adminPool.Ping(ctx); err != nil {
-		adminPool.Close()
-		t.Skipf("flow-postgres compose stack not reachable: %v", err)
-	}
-
-	dbName := fmt.Sprintf("flow_test_%d_%d", os.Getpid(), time.Now().UnixNano())
-	ident := pgx.Identifier{dbName}.Sanitize()
-	if _, err := adminPool.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
-		adminPool.Close()
-		t.Fatalf("create test database %s: %v", dbName, err)
-	}
-	adminPool.Close()
-
-	t.Cleanup(func() {
-		dropCtx, dropCancel := context.WithTimeout(context.Background(), 3*time.Second)
-		defer dropCancel()
-		dropPool, err := pgxpool.New(dropCtx, adminDSN())
-		if err != nil {
-			t.Logf("drop test database %s: reconnect failed: %v", dbName, err)
-			return
-		}
-		defer dropPool.Close()
-		if _, err := dropPool.Exec(dropCtx, "DROP DATABASE IF EXISTS "+ident+" WITH (FORCE)"); err != nil {
-			t.Logf("drop test database %s: %v", dbName, err)
-		}
-	})
-
-	dsn := dsnForDatabase(dbName)
-
-	openCtx, openCancel := context.WithTimeout(context.Background(), 3*time.Second)
-	defer openCancel()
-	st, err := store.Open(openCtx, dsn)
-	if err != nil {
-		t.Fatalf("open test store: %v", err)
-	}
-	if err := st.RunMigrations(openCtx); err != nil {
-		st.Close()
-		t.Fatalf("run migrations: %v", err)
-	}
-	st.Close()
-
-	return dsn
+	return storetest.NewMigratedDatabase(t)
 }
 
 // TestJournalFlushCommandReplaysOnDemand pins `flow journal flush`
@@ -167,22 +107,3 @@ func TestJournalFlushCommandConnectFailure(t *testing.T) {
 		t.Fatalf("exit code = %d, want 1; stdout=%q stderr=%q", code, stdout.String(), stderr.String())
 	}
 }
-
-// dsnForDatabase rewrites adminDSN's database segment, so FLOW_STATS_ADMIN_DSN
-// moves the admin connection and every per-test connection together. Repeating
-// the credentials here instead made the override half-work: the admin step
-// followed the environment while the per-test connection stayed pinned to the
-// literal, so a Postgres with different credentials failed every test while the
-// admin step succeeded.
-func dsnForDatabase(dbName string) string {
-	out, err := dsnutil.ForDatabase(adminDSN(), dbName)
-	if err != nil {
-		// Panic rather than return a best-effort string. This helper's
-		// predecessor answered confidently when it could not do the job and
-		// silently handed back a corrupted DSN; a test that then connects
-		// reports whatever that connection says, which is the wrong question
-		// answered convincingly.
-		panic(fmt.Sprintf("deriving a per-test DSN: %v", err))
-	}
-	return out
-}
diff --git a/stats/internal/api/stats_test.go b/stats/internal/api/stats_test.go
index 1cb36c11..7f22134f 100644
--- a/stats/internal/api/stats_test.go
+++ b/stats/internal/api/stats_test.go
@@ -5,7 +5,6 @@ import (
 	"encoding/json"
 	"errors"
 	"fmt"
-	dsnutil "github.com/tweety53/agents/stats/internal/dsn"
 	"io"
 	"net/http"
 	"net/http/httptest"
@@ -15,13 +14,11 @@ import (
 	"testing"
 	"time"
 
-	"github.com/jackc/pgx/v5"
-	"github.com/jackc/pgx/v5/pgxpool"
-
 	"github.com/tweety53/agents/stats/internal/api"
 	"github.com/tweety53/agents/stats/internal/config"
 	"github.com/tweety53/agents/stats/internal/records"
 	"github.com/tweety53/agents/stats/internal/store"
+	"github.com/tweety53/agents/stats/internal/storetest"
 )
 
 // --- fakeStore's StatsStore surface -----------------------------------
@@ -1515,67 +1512,16 @@ func TestStatsViewTimesOutCleanly(t *testing.T) {
 // numbers it is told are the right ones.
 // ==========================================================================
 
-func statsAdminDSN() string {
-	if v := os.Getenv("FLOW_STATS_ADMIN_DSN"); v != "" {
-		return v
-	}
-	return "postgres://flow:flow@localhost:5433/flow?sslmode=disable"
-}
+// TestMain drops the package's template database once every test has run.
+func TestMain(m *testing.M) { os.Exit(storetest.Main(m)) }
 
-// newIntegrationStore creates a uniquely-named, migrated database against
-// the compose stack's PostgreSQL and returns a *store.Store backed by it,
-// registering cleanup. It skips cleanly, with a clear message, when the
-// stack is not reachable -- exactly internal/store's own testsupport_test.go
-// pattern, reproduced here because that helper is unexported in package
-// store_test and this package cannot import it.
+// newIntegrationStore opens a store on a throwaway, migrated database -- a copy of
+// the package's one migrated template (internal/storetest) -- and registers
+// its Close and the database's drop as test cleanup. It skips cleanly when
+// the stack is not reachable.
 func newIntegrationStore(t *testing.T) *store.Store {
 	t.Helper()
-
-	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
-	defer cancel()
-
-	adminPool, err := pgxpool.New(ctx, statsAdminDSN())
-	if err != nil {
-		t.Skipf("flow-postgres compose stack not reachable: %v", err)
-	}
-	if err := adminPool.Ping(ctx); err != nil {
-		adminPool.Close()
-		t.Skipf("flow-postgres compose stack not reachable: %v", err)
-	}
-
-	dbName := fmt.Sprintf("flow_apitest_%d_%d", os.Getpid(), time.Now().UnixNano())
-	ident := pgx.Identifier{dbName}.Sanitize()
-	if _, err := adminPool.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
-		adminPool.Close()
-		t.Fatalf("create test database %s: %v", dbName, err)
-	}
-	adminPool.Close()
-
-	t.Cleanup(func() {
-		dropCtx, dropCancel := context.WithTimeout(context.Background(), 3*time.Second)
-		defer dropCancel()
-		dropPool, err := pgxpool.New(dropCtx, statsAdminDSN())
-		if err != nil {
-			t.Logf("drop test database %s: reconnect failed: %v", dbName, err)
-			return
-		}
-		defer dropPool.Close()
-		if _, err := dropPool.Exec(dropCtx, "DROP DATABASE IF EXISTS "+ident+" WITH (FORCE)"); err != nil {
-			t.Logf("drop test database %s: %v", dbName, err)
-		}
-	})
-
-	dsn := dsnForDatabase(dbName)
-	st, err := store.Open(ctx, dsn)
-	if err != nil {
-		t.Fatalf("open test store: %v", err)
-	}
-	if err := st.RunMigrations(ctx); err != nil {
-		st.Close()
-		t.Fatalf("run migrations: %v", err)
-	}
-	t.Cleanup(st.Close)
-	return st
+	return storetest.NewStore(t)
 }
 
 // newIntegrationTestServer wires a server whose every store dependency is
@@ -1938,25 +1884,6 @@ func TestCostPerChangeAcceptsChangeAlone(t *testing.T) {
 // which this package cannot reach.
 func sptr(v string) *string { return &v }
 
-// dsnForDatabase rewrites statsAdminDSN's database segment, so FLOW_STATS_ADMIN_DSN
-// moves the admin connection and every per-test connection together. Repeating
-// the credentials here instead made the override half-work: the admin step
-// followed the environment while the per-test connection stayed pinned to the
-// literal, so a Postgres with different credentials failed every test while the
-// admin step succeeded.
-func dsnForDatabase(dbName string) string {
-	out, err := dsnutil.ForDatabase(statsAdminDSN(), dbName)
-	if err != nil {
-		// Panic rather than return a best-effort string. This helper's
-		// predecessor answered confidently when it could not do the job and
-		// silently handed back a corrupted DSN; a test that then connects
-		// reports whatever that connection says, which is the wrong question
-		// answered convincingly.
-		panic(fmt.Sprintf("deriving a per-test DSN: %v", err))
-	}
-	return out
-}
-
 // TestStatsSelfReviewView asserts the self-review view's wire shape: one row
 // per finding, and a fixed row's commitUrl built from the server's commit
 // base -- empty for filed and declined rows, and for every row when the
diff --git a/stats/internal/harvest/endtoend_test.go b/stats/internal/harvest/endtoend_test.go
index bfb7084a..5632cda0 100644
--- a/stats/internal/harvest/endtoend_test.go
+++ b/stats/internal/harvest/endtoend_test.go
@@ -28,111 +28,35 @@ import (
 	"context"
 	"encoding/json"
 	"fmt"
-	dsnutil "github.com/tweety53/agents/stats/internal/dsn"
 	"os"
 	"path/filepath"
 	"testing"
 	"time"
 
-	"github.com/jackc/pgx/v5"
-	"github.com/jackc/pgx/v5/pgxpool"
-
 	"github.com/tweety53/agents/stats/internal/harvest"
 	"github.com/tweety53/agents/stats/internal/store"
+	"github.com/tweety53/agents/stats/internal/storetest"
 )
 
 // --- real-store test support, mirroring internal/store/testsupport_test.go ---
 
-func e2eAdminDSN() string {
-	if v := os.Getenv("FLOW_STATS_ADMIN_DSN"); v != "" {
-		return v
-	}
-	return "postgres://flow:flow@localhost:5433/flow?sslmode=disable"
-}
+// TestMain drops the package's template database once every test has run.
+func TestMain(m *testing.M) { os.Exit(storetest.Main(m)) }
 
-// Derives from e2eAdminDSN so FLOW_STATS_ADMIN_DSN moves both connections
-// together; see internal/store/testsupport_test.go's testDSN for why repeating
-// the credentials here makes the override only half-work.
-func e2eTestDSN(dbName string) string {
-	out, err := dsnutil.ForDatabase(e2eAdminDSN(), dbName)
-	if err != nil {
-		// Panic rather than return a best-effort string. This helper's
-		// predecessor answered confidently when it could not do the job and
-		// silently handed back a corrupted DSN; a test that then connects
-		// reports whatever that connection says, which is the wrong question
-		// answered convincingly.
-		panic(fmt.Sprintf("deriving a per-test DSN: %v", err))
-	}
-	return out
-}
-
-// newEndToEndStore creates a uniquely-named, migrated database against the
-// compose stack and returns a *store.Store whose Close and whose database
-// drop are both registered as test cleanup. It skips cleanly, with a clear
-// message, when the stack is not reachable -- the same guarantee
-// internal/store's and internal/reconcile's own newTestStore helpers give
-// their packages, duplicated here rather than exported, for the same
-// reason internal/reconcile's copy gives (its own doc comment): those
-// helpers live in package store_test, deliberately internal to that
-// package.
+// newEndToEndStore opens a store on a throwaway, migrated database -- a copy of
+// the package's one migrated template (internal/storetest) -- and registers
+// its Close and the database's drop as test cleanup, then seeds pricing the
+// way cmd/flowd does at startup. It skips cleanly when the stack is not
+// reachable.
 func newEndToEndStore(t *testing.T) *store.Store {
 	t.Helper()
+	st := storetest.NewStore(t)
 
 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
 	defer cancel()
-
-	adminPool, err := pgxpool.New(ctx, e2eAdminDSN())
-	if err != nil {
-		t.Skipf("flow-postgres compose stack not reachable: %v", err)
-	}
-	if err := adminPool.Ping(ctx); err != nil {
-		adminPool.Close()
-		t.Skipf("flow-postgres compose stack not reachable: %v", err)
-	}
-
-	dbName := fmt.Sprintf("flow_test_e2e_%d_%d", os.Getpid(), time.Now().UnixNano())
-	ident := pgx.Identifier{dbName}.Sanitize()
-	if _, err := adminPool.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
-		adminPool.Close()
-		t.Fatalf("create test database %s: %v", dbName, err)
-	}
-	adminPool.Close()
-
-	t.Cleanup(func() {
-		dropCtx, dropCancel := context.WithTimeout(context.Background(), 3*time.Second)
-		defer dropCancel()
-
-		dropPool, err := pgxpool.New(dropCtx, e2eAdminDSN())
-		if err != nil {
-			t.Logf("drop test database %s: reconnect failed: %v", dbName, err)
-			return
-		}
-		defer dropPool.Close()
-
-		if _, err := dropPool.Exec(dropCtx, "DROP DATABASE IF EXISTS "+ident+" WITH (FORCE)"); err != nil {
-			t.Logf("drop test database %s: %v", dbName, err)
-		}
-	})
-
-	dsn := e2eTestDSN(dbName)
-
-	openCtx, openCancel := context.WithTimeout(context.Background(), 3*time.Second)
-	defer openCancel()
-
-	st, err := store.Open(openCtx, dsn)
-	if err != nil {
-		t.Fatalf("open test store: %v", err)
-	}
-	if err := st.RunMigrations(openCtx); err != nil {
-		st.Close()
-		t.Fatalf("run migrations: %v", err)
-	}
-	t.Cleanup(st.Close)
-
-	if err := st.SeedPricing(openCtx); err != nil {
+	if err := st.SeedPricing(ctx); err != nil {
 		t.Fatalf("seed pricing: %v", err)
 	}
-
 	return st
 }
 
diff --git a/stats/internal/reconcile/testsupport_test.go b/stats/internal/reconcile/testsupport_test.go
index 470d8858..f7539bfe 100644
--- a/stats/internal/reconcile/testsupport_test.go
+++ b/stats/internal/reconcile/testsupport_test.go
@@ -1,111 +1,21 @@
 package reconcile_test
 
 import (
-	"context"
-	"fmt"
-	dsnutil "github.com/tweety53/agents/stats/internal/dsn"
 	"os"
 	"testing"
-	"time"
-
-	"github.com/jackc/pgx/v5"
-	"github.com/jackc/pgx/v5/pgxpool"
 
 	"github.com/tweety53/agents/stats/internal/store"
+	"github.com/tweety53/agents/stats/internal/storetest"
 )
 
-// adminDSN is the DSN used to create and drop per-test databases -- the
-// same dedicated flow-postgres compose stack internal/store's own tests
-// use (task 1), overridable for environments that run Postgres elsewhere.
-func adminDSN() string {
-	if v := os.Getenv("FLOW_STATS_ADMIN_DSN"); v != "" {
-		return v
-	}
-	return "postgres://flow:flow@localhost:5433/flow?sslmode=disable"
-}
-
-// It derives from adminDSN rather than repeating the credentials, so that
-// FLOW_STATS_ADMIN_DSN moves both connections together. Hardcoding them here
-// made the override only half-work -- the admin connection followed the
-// environment while the per-test one stayed pinned, so a Postgres with
-// different credentials failed every test while the admin step succeeded. See
-// internal/store/testsupport_test.go's testDSN for the measured case.
-func testDSN(dbName string) string {
-	out, err := dsnutil.ForDatabase(adminDSN(), dbName)
-	if err != nil {
-		// Panic rather than return a best-effort string. This helper's
-		// predecessor answered confidently when it could not do the job and
-		// silently handed back a corrupted DSN; a test that then connects
-		// reports whatever that connection says, which is the wrong question
-		// answered convincingly.
-		panic(fmt.Sprintf("deriving a per-test DSN: %v", err))
-	}
-	return out
-}
+// TestMain drops the package's template database once every test has run.
+func TestMain(m *testing.M) { os.Exit(storetest.Main(m)) }
 
-// newTestStore creates a uniquely-named, migrated database against the
-// compose stack, returning a *store.Store whose Close and whose database
-// drop are both registered as test cleanup. It skips cleanly, with a clear
-// message, when the stack is not reachable -- mirroring
-// internal/store/testsupport_test.go's own helper exactly, duplicated here
-// rather than exported from internal/store: that package's test helpers
-// are deliberately internal (package store_test), and reconcile's tests
-// need the same real-Postgres guarantee design.md's testing strategy
-// requires ("store is tested against a real PostgreSQL ... never a mock")
-// for the tests that exercise the actual monotonic guard.
+// newTestStore opens a store on a throwaway, migrated database -- a copy of
+// the package's one migrated template (internal/storetest) -- and registers
+// its Close and the database's drop as test cleanup. It skips cleanly when
+// the stack is not reachable.
 func newTestStore(t *testing.T) *store.Store {
 	t.Helper()
-
-	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
-	defer cancel()
-
-	adminPool, err := pgxpool.New(ctx, adminDSN())
-	if err != nil {
-		t.Skipf("flow-postgres compose stack not reachable: %v", err)
-	}
-	if err := adminPool.Ping(ctx); err != nil {
-		adminPool.Close()
-		t.Skipf("flow-postgres compose stack not reachable: %v", err)
-	}
-
-	dbName := fmt.Sprintf("flow_test_%d_%d", os.Getpid(), time.Now().UnixNano())
-	ident := pgx.Identifier{dbName}.Sanitize()
-	if _, err := adminPool.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
-		adminPool.Close()
-		t.Fatalf("create test database %s: %v", dbName, err)
-	}
-	adminPool.Close()
-
-	t.Cleanup(func() {
-		dropCtx, dropCancel := context.WithTimeout(context.Background(), 3*time.Second)
-		defer dropCancel()
-
-		dropPool, err := pgxpool.New(dropCtx, adminDSN())
-		if err != nil {
-			t.Logf("drop test database %s: reconnect failed: %v", dbName, err)
-			return
-		}
-		defer dropPool.Close()
-
-		if _, err := dropPool.Exec(dropCtx, "DROP DATABASE IF EXISTS "+ident+" WITH (FORCE)"); err != nil {
-			t.Logf("drop test database %s: %v", dbName, err)
-		}
-	})
-
-	dsn := testDSN(dbName)
-
-	openCtx, openCancel := context.WithTimeout(context.Background(), 3*time.Second)
-	defer openCancel()
-
-	st, err := store.Open(openCtx, dsn)
-	if err != nil {
-		t.Fatalf("open test store: %v", err)
-	}
-	if err := st.RunMigrations(openCtx); err != nil {
-		st.Close()
-		t.Fatalf("run migrations: %v", err)
-	}
-	t.Cleanup(st.Close)
-
-	return st
+	return storetest.NewStore(t)
 }
diff --git a/stats/internal/sweep/testsupport_test.go b/stats/internal/sweep/testsupport_test.go
index 18da1018..eada7eb0 100644
--- a/stats/internal/sweep/testsupport_test.go
+++ b/stats/internal/sweep/testsupport_test.go
@@ -2,112 +2,24 @@ package sweep_test
 
 import (
 	"context"
-	"fmt"
-	dsnutil "github.com/tweety53/agents/stats/internal/dsn"
 	"os"
 	"testing"
 	"time"
 
-	"github.com/jackc/pgx/v5"
-	"github.com/jackc/pgx/v5/pgxpool"
-
 	"github.com/tweety53/agents/stats/internal/store"
+	"github.com/tweety53/agents/stats/internal/storetest"
 )
 
-// adminDSN is the DSN used to create and drop per-test databases -- the
-// same dedicated flow-postgres compose stack internal/store's own tests
-// use (task 1), overridable for environments that run Postgres elsewhere.
-func adminDSN() string {
-	if v := os.Getenv("FLOW_STATS_ADMIN_DSN"); v != "" {
-		return v
-	}
-	return "postgres://flow:flow@localhost:5433/flow?sslmode=disable"
-}
-
-// It derives from adminDSN rather than repeating the credentials, so that
-// FLOW_STATS_ADMIN_DSN moves both connections together. Hardcoding them here
-// made the override only half-work -- the admin connection followed the
-// environment while the per-test one stayed pinned, so a Postgres with
-// different credentials failed every test while the admin step succeeded. See
-// internal/store/testsupport_test.go's testDSN for the measured case.
-func testDSN(dbName string) string {
-	out, err := dsnutil.ForDatabase(adminDSN(), dbName)
-	if err != nil {
-		// Panic rather than return a best-effort string. This helper's
-		// predecessor answered confidently when it could not do the job and
-		// silently handed back a corrupted DSN; a test that then connects
-		// reports whatever that connection says, which is the wrong question
-		// answered convincingly.
-		panic(fmt.Sprintf("deriving a per-test DSN: %v", err))
-	}
-	return out
-}
+// TestMain drops the package's template database once every test has run.
+func TestMain(m *testing.M) { os.Exit(storetest.Main(m)) }
 
-// newTestStore creates a uniquely-named, migrated database against the
-// compose stack, returning a *store.Store whose Close and whose database
-// drop are both registered as test cleanup. It skips cleanly, with a clear
-// message, when the stack is not reachable -- mirroring
-// internal/store/testsupport_test.go's own helper and
-// internal/reconcile/testsupport_test.go's copy of it: sweep's own tests
-// need the same real-Postgres guarantee design.md's testing strategy
-// requires ("store is tested against a real PostgreSQL ... never a mock")
-// for the concurrency behaviour under test, which a mock could not
-// reproduce.
+// newTestStore opens a store on a throwaway, migrated database -- a copy of
+// the package's one migrated template (internal/storetest) -- and registers
+// its Close and the database's drop as test cleanup. It skips cleanly when
+// the stack is not reachable.
 func newTestStore(t *testing.T) *store.Store {
 	t.Helper()
-
-	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
-	defer cancel()
-
-	adminPool, err := pgxpool.New(ctx, adminDSN())
-	if err != nil {
-		t.Skipf("flow-postgres compose stack not reachable: %v", err)
-	}
-	if err := adminPool.Ping(ctx); err != nil {
-		adminPool.Close()
-		t.Skipf("flow-postgres compose stack not reachable: %v", err)
-	}
-
-	dbName := fmt.Sprintf("flow_test_%d_%d", os.Getpid(), time.Now().UnixNano())
-	ident := pgx.Identifier{dbName}.Sanitize()
-	if _, err := adminPool.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
-		adminPool.Close()
-		t.Fatalf("create test database %s: %v", dbName, err)
-	}
-	adminPool.Close()
-
-	t.Cleanup(func() {
-		dropCtx, dropCancel := context.WithTimeout(context.Background(), 3*time.Second)
-		defer dropCancel()
-
-		dropPool, err := pgxpool.New(dropCtx, adminDSN())
-		if err != nil {
-			t.Logf("drop test database %s: reconnect failed: %v", dbName, err)
-			return
-		}
-		defer dropPool.Close()
-
-		if _, err := dropPool.Exec(dropCtx, "DROP DATABASE IF EXISTS "+ident+" WITH (FORCE)"); err != nil {
-			t.Logf("drop test database %s: %v", dbName, err)
-		}
-	})
-
-	dsn := testDSN(dbName)
-
-	openCtx, openCancel := context.WithTimeout(context.Background(), 3*time.Second)
-	defer openCancel()
-
-	st, err := store.Open(openCtx, dsn)
-	if err != nil {
-		t.Fatalf("open test store: %v", err)
-	}
-	if err := st.RunMigrations(openCtx); err != nil {
-		st.Close()
-		t.Fatalf("run migrations: %v", err)
-	}
-	t.Cleanup(st.Close)
-
-	return st
+	return storetest.NewStore(t)
 }
 
 // seedChange puts a minimal, valid change under projectKey/name so a
````

  - [x] **Step 2: Verify.** `cd stats && gofmt -l ./internal/reconcile ./internal/sweep ./internal/harvest ./internal/api ./cmd/flow && go vet ./... && go test -count=1 ./internal/harvest/ -run 'TestMarkedStageBindsToRealSessionEndToEnd' && go test -count=1 ./internal/api/ -run 'TestMultiRepoChangeIsOneRowInEveryView|TestCostPerChangeAcceptsChangeAlone' && go test -count=1 ./cmd/flow/ -run 'TestJournalFlushCommand' && go test -count=1 ./internal/reconcile/ ./internal/sweep/`
    — gofmt prints nothing, the rest exit 0.
  - [x] **Step 3: Commit** the five files with the subject below.

**Build:** green
**Files:** `stats/internal/reconcile/testsupport_test.go`, `stats/internal/sweep/testsupport_test.go`, `stats/internal/harvest/endtoend_test.go`, `stats/internal/api/stats_test.go`, `stats/cmd/flow/journal_test.go`
**Tests:** none — repairs existing tests: every store-backed test in `internal/reconcile`, `internal/sweep`, `internal/harvest`, `internal/api` and `cmd/flow` now gets a template copy; Task 1's `storetest` tests cover the shared mechanism.
**Regression:** reverting the commit puts these five packages back on per-test migration, whose load slows every package's setup under whole-suite load — the api and reconcile failures reproduced at plan time.
**Baseline:** before=646 after=646
<!-- measured: cat stats/internal/reconcile/*_test.go stats/internal/sweep/*_test.go stats/internal/harvest/*_test.go stats/internal/api/*_test.go stats/cmd/flow/*_test.go | grep -E '^func Test' | grep -vc '^func TestMain(' (before at e777ee03 with Task 1's diff; after with both diffs applied) -->
**Commit:** fix(stats): give every store-backed test package a template copy
**After:** Task 1

**Decision:** template-per-process
**Decision:** storetest-package
