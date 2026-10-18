# self-review-backlog-flow-5-oct18

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-18-self-review-backlog-flow-5-oct18-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; the commit body names the task's Jira key. Never write `spectre/changes/`. Never merge or push.
- Each task names the repository it runs in, by its worktree; every `**Files:**` path is relative to that worktree's root. Tasks 1 and 2 run in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-5-oct18`, Task 3 in gymie-playwright's `/Users/tweety53/Projects/gymie-playwright-worktrees/self-review-backlog-flow-5-oct18`, Tasks 4 and 5 in gymie's `/Users/tweety53/Projects/gymie-worktrees/self-review-backlog-flow-5-oct18`. Each worktree is on branch `spectre/self-review-backlog-flow-5-oct18`; never touch a main checkout.
- A task's diff is the unified diff its Step 2 gives, applied with `git apply` from its worktree's root, in task order — every diff was applied and checked there (`design.md`, **Context**). Save a diff by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`.

## Review focus

- The dev workspace's daemon (`flow` database) must still replay the journal at startup and on reconnect (Task 1).
- `setup.sh global` from the main checkout, and from any checkout into a sandboxed `$HOME`, must install exactly as before (Task 2).
- `LocalDevSeederIntegrationTest`'s count assertions stay exact, never widened (Task 5).


- [ ] 1. A workspace flowd never replays the shared journal
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-5-oct18`, apply Step 2's diff to the test file alone (`git apply --include='stats/cmd/flowd/wiring_test.go' <the saved diff>`) and run `cd stats && make -s web-dist && go test ./cmd/flowd -run 'TestReplaysSharedJournalOnlyIntoTheFlowDatabase' -count=1`: it fails to build, `undefined: replaysSharedJournal` — `main.go` replays unconditionally. Keep the applied hunk; Step 2 applies the rest with `git apply --exclude='stats/cmd/flowd/wiring_test.go' <the saved diff>`.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at agents 52438f0d; with it applied the test passes and go vet ./cmd/flowd is clean
diff --git a/stats/cmd/flowd/main.go b/stats/cmd/flowd/main.go
index d7bb877f..817ef57f 100644
--- a/stats/cmd/flowd/main.go
+++ b/stats/cmd/flowd/main.go
@@ -19,6 +19,7 @@ import (
 	"fmt"
 	"log/slog"
 	"net"
+	"net/url"
 	"os"
 	"os/exec"
 	"os/signal"
@@ -177,15 +178,21 @@ func run(logger *slog.Logger) error {
 	// above has already confirmed connectivity via its own ping, so this
 	// is exactly the "at startup" half of design.md's "Availability and
 	// reconciliation". The "on reconnect" half is Watch, started below.
-	reconciler := reconcile.New(st, st, st, fallback.StateRoot(), logger)
-	startupResult, startupErr := reconciler.Run(ctx)
-	logReconcileResult(logger, "startup", startupResult, startupErr)
-
+	// A workspace daemon skips both halves: the journal is shared, and an
+	// entry replayed into its throwaway database leaves the journal and is
+	// dropped with that database at cleanup (KAN-913).
 	watchCtx, stopWatch := context.WithCancel(ctx)
 	defer stopWatch()
-	go reconciler.Watch(watchCtx, st.Ping, reconnectPingInterval, func(result reconcile.Result, err error) {
-		logReconcileResult(logger, "reconnect", result, err)
-	})
+	if replaysSharedJournal(cfg.DSN) {
+		reconciler := reconcile.New(st, st, st, fallback.StateRoot(), logger)
+		startupResult, startupErr := reconciler.Run(ctx)
+		logReconcileResult(logger, "startup", startupResult, startupErr)
+		go reconciler.Watch(watchCtx, st.Ping, reconnectPingInterval, func(result reconcile.Result, err error) {
+			logReconcileResult(logger, "reconnect", result, err)
+		})
+	} else {
+		logger.Info("flowd journal replay skipped: not the dev workspace's flow database")
+	}
 
 	// Task 9: harvest session transcripts and attribute their usage to
 	// open stage windows. st (a *store.Store) satisfies both
@@ -407,6 +414,15 @@ var (
 	_ harvest.Deps = (*store.Store)(nil)
 )
 
+// replaysSharedJournal reports whether dsn names the dev workspace's own
+// `flow` database, the one store the shared journal under
+// fallback.StateRoot() is replayed into. A workspace database
+// (flow_<id>, .flow/project.md's ## workspace isolation) never is.
+func replaysSharedJournal(dsn string) bool {
+	u, err := url.Parse(dsn)
+	return err == nil && u.Path == "/flow"
+}
+
 // logReconcileResult reports one Reconciler.Run outcome. A replay failure
 // is deliberately not fatal to the daemon -- reconcile.Reconciler.Run
 // already leaves anything it could not resolve safely in the journal for
diff --git a/stats/cmd/flowd/wiring_test.go b/stats/cmd/flowd/wiring_test.go
index a70e1857..43cb1b53 100644
--- a/stats/cmd/flowd/wiring_test.go
+++ b/stats/cmd/flowd/wiring_test.go
@@ -231,3 +231,21 @@ func TestCommitBaseFromRemote(t *testing.T) {
 		}
 	}
 }
+
+// TestReplaysSharedJournalOnlyIntoTheFlowDatabase pins KAN-913: a
+// workspace daemon's own flow_<id> database never replays the shared
+// journal, while the dev workspace's flow database still does.
+func TestReplaysSharedJournalOnlyIntoTheFlowDatabase(t *testing.T) {
+	cases := map[string]bool{
+		config.DefaultDSN: true,
+		"postgres://flow:flow@localhost:5433/flow_kan_9_x?sslmode=disable":   false,
+		"postgres://flow:flow@localhost:5433/flow_ui_uitest?sslmode=disable": false,
+		"postgres://flow:flow@localhost:5433/flowd?sslmode=disable":          false,
+		"host=localhost port=5433 dbname=flow user=flow":                     false,
+	}
+	for dsn, want := range cases {
+		if got := replaysSharedJournal(dsn); got != want {
+			t.Errorf("replaysSharedJournal(%q) = %v, want %v", dsn, got, want)
+		}
+	}
+}
````

  - [ ] **Step 3: Verify.** In `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-5-oct18/stats`: `gofmt -l ./cmd/flowd` prints nothing, `go vet ./cmd/flowd` exits 0, and `go test ./cmd/flowd -run 'TestReplaysSharedJournalOnlyIntoTheFlowDatabase' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/cmd/flowd/main.go`, `stats/cmd/flowd/wiring_test.go`
**Tests:** `TestReplaysSharedJournalOnlyIntoTheFlowDatabase`
**Regression:** reverting the commit removes `replaysSharedJournal`, so the test no longer builds and a workspace daemon replays the shared journal again.
**Baseline:** before=6 after=7 <!-- measured: grep -c '^func Test' stats/cmd/flowd/wiring_test.go @ 52438f0d -->
**Commit:** fix(stats): replay the shared journal only into the dev workspace's flow database
**After:** none

**Decision:** journal-replay-flow-db-only

- [ ] 2. Refuse a global install into the operator's home from a linked worktree
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-5-oct18`, apply Step 2's diff to the two test files alone (`git apply --include='stats/internal/setuptest/*' <the saved diff>`) and run `cd stats && go test ./internal/setuptest -run 'TestGlobalInstallRefusedFromLinkedWorktree' -count=1`: it fails `✗ global from a linked worktree exits non-zero`, `✗ the refusal names the main checkout's command` and `✗ nothing is linked into $HOME` — `setup.sh` installs from the worktree. Keep the applied hunks; Step 2 applies the rest with `git apply --exclude='stats/internal/setuptest/*' <the saved diff>`.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at agents 52438f0d; with it applied go test ./internal/setuptest passes, and with setup.sh reverted the new test fails its three refusal checks
diff --git a/setup.sh b/setup.sh
index 744169b9..a289d492 100755
--- a/setup.sh
+++ b/setup.sh
@@ -1216,6 +1216,17 @@ install_global() {
   echo "   from a git worktree or feature branch the links break when it is removed."
 }
 
+# A global install into the operator's own home links it into this checkout, so from a linked
+# worktree every session's rules break the moment cleanup removes it (KAN-773). A sandboxed
+# $HOME — every test and sandbox-diff run — is not the operator's and installs as before. The
+# home is the password database's (bash's ~user), SETUP_REAL_HOME standing in for it in a test.
+real_home="${SETUP_REAL_HOME:-$(eval echo "~$(id -un)")}"
+if [ "$HARNESS" = global ] && [ "$HOME" = "$real_home" ] \
+  && common_dir="$(git -C "$SCRIPT_DIR" rev-parse --path-format=absolute --git-common-dir 2>/dev/null)" \
+  && [ "$(git -C "$SCRIPT_DIR" rev-parse --absolute-git-dir)" != "$common_dir" ]; then
+  die "$SCRIPT_DIR is a linked worktree — a global install from it breaks when it is removed. Run from the main checkout: $(dirname "$common_dir")/setup.sh global"
+fi
+
 case "$HARNESS" in
   global)      install_global ;;
   claude-code) install_claude_code ;;
diff --git a/stats/internal/setuptest/harnesses_test.go b/stats/internal/setuptest/harnesses_test.go
index 7d460cce..277e2798 100644
--- a/stats/internal/setuptest/harnesses_test.go
+++ b/stats/internal/setuptest/harnesses_test.go
@@ -3,6 +3,7 @@ package setuptest
 import (
 	"io/fs"
 	"os"
+	"os/exec"
 	"path/filepath"
 	"sort"
 	"strings"
@@ -216,3 +217,39 @@ func TestZcodeCompactWindowEnvBlock(t *testing.T) {
 	g.assertIdentical("an unmanaged Z_COMPACT_WINDOW line is left byte-identical", zshrc, expected)
 	g.assertContains("the skip is explained in the log", g.log, "outside the managed block")
 }
+
+// TestGlobalInstallRefusedFromLinkedWorktree pins KAN-773: `global` into the operator's own
+// home (SETUP_REAL_HOME here) from a linked worktree exits non-zero naming the main checkout's
+// command and links nothing; the main checkout still installs there, and the worktree still
+// installs into any other home — the sandboxed one every other test here uses.
+func TestGlobalInstallRefusedFromLinkedWorktree(t *testing.T) {
+	g := newGroup(t)
+	main := filepath.Join(g.dir, "fixture-repo-main")
+	if err := makeFixtureRepo(main, false); err != nil {
+		t.Fatal(err)
+	}
+	wt := filepath.Join(g.dir, "fixture-repo-wt")
+	for _, args := range [][]string{
+		{"init", "-q", "-b", "main"},
+		{"add", "-A"},
+		{"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "fixture"},
+		{"worktree", "add", "-q", wt},
+	} {
+		if out, err := exec.Command("git", append([]string{"-C", main}, args...)...).CombinedOutput(); err != nil {
+			t.Fatalf("git %v: %v\n%s", args, err, out)
+		}
+	}
+
+	home := g.newHome()
+	g.env = []string{"SETUP_REAL_HOME=" + home}
+	g.runSetup(wt, home, "global", "")
+	g.assertRCNonzero("global from a linked worktree exits non-zero", g.rc)
+	g.assertContains("the refusal names the main checkout's command", g.log, filepath.Join(main, "setup.sh")+" global")
+	g.assertAbsent("nothing is linked into $HOME", filepath.Join(home, ".claude/rules"))
+
+	g.runSetup(main, home, "global", "")
+	g.assertRCZero("global from the main checkout succeeds", g.rc, g.log)
+
+	g.runSetup(wt, g.newHome(), "global", "")
+	g.assertRCZero("global from the worktree into a sandboxed home succeeds", g.rc, g.log)
+}
diff --git a/stats/internal/setuptest/helpers_test.go b/stats/internal/setuptest/helpers_test.go
index aaf09887..343173c2 100644
--- a/stats/internal/setuptest/helpers_test.go
+++ b/stats/internal/setuptest/helpers_test.go
@@ -20,8 +20,9 @@ type group struct {
 	t   *testing.T
 	dir string // $SANDBOX/<TestName>; every path this group writes lies under it
 	seq int
-	rc  int    // the last runSetup's exit status
-	log string // the last runSetup's combined-output log
+	rc  int      // the last runSetup's exit status
+	log string   // the last runSetup's combined-output log
+	env []string // extra environment every later runSetup passes the installer
 }
 
 // newGroup marks the test parallel and gives it its own sub-sandbox.
@@ -76,7 +77,7 @@ func (g *group) runSetup(repo, home, mode, proj string) {
 	defer cancel()
 	cmd := exec.CommandContext(ctx, filepath.Join(repo, "setup.sh"), mode, proj)
 	cmd.Dir = proj
-	cmd.Env = withHome(os.Environ(), home)
+	cmd.Env = append(withHome(os.Environ(), home), g.env...)
 	cmd.Stdout, cmd.Stderr = f, f
 	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
 	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
````

  - [ ] **Step 3: Verify.** In `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-5-oct18/stats`: `gofmt -l ./internal/setuptest` prints nothing, `go vet ./internal/setuptest` exits 0, and `go test ./internal/setuptest -run 'TestGlobalInstallRefusedFromLinkedWorktree|TestReinstallPrunesDeletedSourceLinks' -count=1` passes — the second a sandboxed-home install from this worktree, which must still succeed. Commit.

**Build:** green
**Files:** `setup.sh`, `stats/internal/setuptest/harnesses_test.go`, `stats/internal/setuptest/helpers_test.go`
**Tests:** `TestGlobalInstallRefusedFromLinkedWorktree`
**Regression:** reverting the `setup.sh` hunk lets a linked worktree link the operator's home into itself, failing the test's three refusal checks.
**Baseline:** before=7 after=8 <!-- measured: grep -c '^func Test' stats/internal/setuptest/harnesses_test.go @ 52438f0d -->
**Commit:** fix(setup): refuse a global install into the operator's home from a linked worktree
**After:** none

**Decision:** setup-global-worktree-refusal

- [ ] 3. A gymie-playwright worktree installs its own dependencies
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/gymie-playwright-worktrees/self-review-backlog-flow-5-oct18`, run `/Users/tweety53/Projects/agents/scripts/project-get.sh "$PWD" 'worktree setup'`: exit 1, no section — kickoff runs no install here.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at gymie-playwright eb0dbb6; with it applied project-get.sh exits 0 printing the npm ci fence, and npm ci resolves @playwright/test inside the worktree
diff --git a/.flow/project.md b/.flow/project.md
new file mode 100644
index 0000000..8b0253e
--- /dev/null
+++ b/.flow/project.md
@@ -0,0 +1,14 @@
+# flow project configuration — gymie-playwright
+
+The Playwright regression checkout `gymie`'s `.flow/project.md` declares under `## apps`. Every other
+key is that project's; this file carries only what a worktree of this repository needs itself.
+
+## worktree setup
+
+A worktree owns its installed dependencies: without its own `node_modules`, Node resolves
+`@playwright/test` from the main checkout's, and a capture or test then runs another tree's
+toolchain while its result is attributed to the branch (KAN-818).
+
+```bash
+npm ci
+```
````

  - [ ] **Step 3: Verify.** In `/Users/tweety53/Projects/gymie-playwright-worktrees/self-review-backlog-flow-5-oct18`: `/Users/tweety53/Projects/agents/scripts/project-get.sh "$PWD" 'worktree setup'` exits 0 and prints the `npm ci` fence; `npm ci` exits 0; `node -e "console.log(require.resolve('@playwright/test'))"` prints a path under `/Users/tweety53/Projects/gymie-playwright-worktrees/self-review-backlog-flow-5-oct18/node_modules`. Commit.

**Build:** green
**Files:** `.flow/project.md`
**Tests:** **none** — configuration; Step 3's `project-get.sh` run is its check.
**Regression:** reverting the commit leaves the section undeclared, so kickoff skips the install and the worktree resolves the main checkout's `@playwright/test`.
**Commit:** chore(flow): install a worktree's own dependencies at setup
**After:** none

**Decision:** playwright-worktree-npm-ci

- [ ] 4. gymie's lint section states the read-only bundle rule
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/gymie-worktrees/self-review-backlog-flow-5-oct18`, run `grep -c 'read-only tasks only' .flow/project.md`: `0`.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at gymie eb06468c
diff --git a/.flow/project.md b/.flow/project.md
index bd46089e..e304cf52 100644
--- a/.flow/project.md
+++ b/.flow/project.md
@@ -467,2 +467,9 @@ database section; a seeded account's weigh-ins and targets are replaced wholesal
 
+**A one-invocation bundle (`verifyChange`, `verify` and their like) holds read-only tasks only.**
+Gradle's parallel executor does not honour, for a task that writes sources, the ordering a dry run
+shows: kan-630 wired format-before-check four ways, each under the configuration cache and without
+it, and every one still ran a check before the format and went falsely red. A source-mutating step
+— `ktlintFormat` — is a documented pre-step run as its own invocation, as above, never a member of
+the bundle (KAN-780).
+
 Pre-approved suppressions and documented deviations live in `CONTRIBUTING.md`.
````

  - [ ] **Step 3: Verify.** In `/Users/tweety53/Projects/gymie-worktrees/self-review-backlog-flow-5-oct18`: `grep -c 'read-only tasks only' .flow/project.md` prints `1`, and `/Users/tweety53/Projects/agents/scripts/project-get.sh "$PWD" lint` exits 0 printing both fenced commands. Commit.

**Build:** green
**Files:** `.flow/project.md`
**Tests:** **none** — prose only.
**Regression:** reverting the commit drops the constraint from where bundle authors read the verify commands.
**Commit:** docs(flow): a one-invocation verify bundle holds read-only tasks only
**After:** Task 3

**Decision:** gymie-lint-read-only-bundle

- [ ] 5. The seeder test pins the invite email unset
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/gymie-worktrees/self-review-backlog-flow-5-oct18`, run `DEV_INVITE_EMAIL=dev@example.com ./gradlew :app:test --tests 'com.gymie.dev.LocalDevSeederIntegrationTest'`: `seedIsIdempotent()` fails `expected: 3L but was: 21L`.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at gymie eb06468c; with it applied the Step 1 command is BUILD SUCCESSFUL and ktlintCheck is clean
diff --git a/src/app/src/test/kotlin/com/gymie/dev/LocalDevSeederIntegrationTest.kt b/src/app/src/test/kotlin/com/gymie/dev/LocalDevSeederIntegrationTest.kt
index db90a8b4..633dac1e 100644
--- a/src/app/src/test/kotlin/com/gymie/dev/LocalDevSeederIntegrationTest.kt
+++ b/src/app/src/test/kotlin/com/gymie/dev/LocalDevSeederIntegrationTest.kt
@@ -35,6 +35,9 @@ class LocalDevSeederIntegrationTest : AppIntegrationTest() {
         @JvmStatic
         fun registerProperties(registry: DynamicPropertyRegistry) {
             registry.add("app.seed.enabled") { "false" }
+            // Pinned unset: application-local.yml reads DEV_INVITE_EMAIL, which adds the
+            // developer's friend cast to the counts these tests pin exactly (KAN-799).
+            registry.add("app.dev.invite-email") { "" }
         }
     }
 
````

  - [ ] **Step 3: Verify.** In `/Users/tweety53/Projects/gymie-worktrees/self-review-backlog-flow-5-oct18`: the Step 1 command, `DEV_INVITE_EMAIL` still exported, is `BUILD SUCCESSFUL`; `./gradlew ktlintFormat` then `./gradlew verifyChange :app:test --tests 'com.gymie.dev.LocalDevSeederIntegrationTest'` passes. Commit.

**Build:** green
**Files:** `src/app/src/test/kotlin/com/gymie/dev/LocalDevSeederIntegrationTest.kt`
**Tests:** **none** — the existing `seedIsIdempotent` is the check: red under `DEV_INVITE_EMAIL` before, green after.
**Regression:** reverting the commit makes `seedIsIdempotent` fail again whenever `DEV_INVITE_EMAIL` is exported.
**Commit:** fix(dev): pin the invite email unset in the seeder test
**After:** none

**Decision:** seeder-test-pins-invite-email
