# self-review-backlog-flow-5-oct18b

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close five held Backlog Flow tickets — KAN-836, KAN-834, KAN-833 (prose rules), KAN-832
(gymie-frontend `verifyChange`) and KAN-827 (flowd replays the journal on every healthy ping).

**Architecture:** two one-sentence additions to `skills/flow/visual-verify-verifier.md` step 8, one
list item in `skills/flow/engineering-principles.md`, one Gradle task plus two `.flow/project.md`
sections in gymie-frontend, and one loop change in `stats/internal/reconcile/reconcile.go` with its
test and the comments and contract sentence that describe it.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-18-self-review-backlog-flow-5-oct18b-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Each task names the repository it runs in, by its worktree; every `**Files:**` path is relative to that worktree's root. Tasks 1, 2, 3 and 5 run in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-5-oct18b`; Task 4 runs in gymie-frontend's `/Users/tweety53/Projects/gymie-frontend-worktrees/self-review-backlog-flow-5-oct18b`. Both are on branch `self-review-backlog-flow-5-oct18b`; never touch a main checkout.
- A task's diff is applied with `git apply` from its worktree's root — each was applied and checked there (`design.md`, **Context**). Save it by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork. A prose task's sentence is inserted character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- Watch's steady state: an empty journal on every healthy tick logs nothing (`logReconcileResult` logs only on an error, an applied or a refused entry) and the store is not written.
- A failing ping still skips the replay — Watch never replays into a store it could not reach.
- `verifyChange` holds no source-mutating task; `ktlintFormat` stays its own pre-step.
- The two step 8 sentences read as additions beside `expectRoom`, never replacing it.

- [ ] 1. Step 8 requires a click-reached capture to assert its state and park the pointer
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -c 'parks the pointer' skills/flow/visual-verify-verifier.md` prints `0`.
  - [ ] **Step 2: Edit.** In `skills/flow/visual-verify-verifier.md` step 8, right after the sentence ending `so a missed navigation fails the capture instead of becoming its baseline.` (it ends mid-line, before `` `screenshots`'s root-not-leaf shape ``), insert one space and then this sentence, on the same line:

```text verified:applied to visual-verify-verifier.md at 02d83a03; scripts/check-verbatim-moves.sh then prints it as a FAIL line identical to verbatim-moves.txt's Task 1 entry
**A capture reached through a click also asserts the state it clicked into, and parks the pointer before it is taken** — a positive control on a drawn, state-dependent property the wrong state cannot satisfy (the selected tab's fill, the opened sheet's tag), and the pointer moved clear of every control (`page.mouse.move(0, 0)`) after the last click — so a missed click fails the capture instead of becoming its baseline, and no hover ink reaches one.
```

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.py` exit 0.
  - [ ] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`
**Tests:** none — a prose rule; `scripts/check-verbatim-moves.sh` fails if the sentence is reworded.
**Regression:** none
**Commit:** docs(visual-verify): a click-reached capture asserts its state and parks the pointer
**After:** none

**Decision:** click-capture-positive-control

- [ ] 2. Step 8 requires a capture under a dimming layer to pin its subject structurally
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -ci 'scrim' skills/flow/visual-verify-verifier.md` prints `0`.
  - [ ] **Step 2: Edit.** In `skills/flow/visual-verify-verifier.md` step 8, right after Task 1's sentence (ending `and no hover ink reaches one.`), insert one space and then this sentence, on the same line:

```text verified:applied after Task 1's sentence at 02d83a03; scripts/check-verbatim-moves.sh then prints it as a FAIL line identical to verbatim-moves.txt's Task 2 entry
**A capture whose subject sits under a scrim, or any dimming or alpha layer, also pins that subject structurally** — the presence of each tag the change draws beneath the layer and the absence of each it removes, covered tags included — because the dimming can shrink a real difference below the screenshot comparison's per-pixel threshold; a dimmed capture whose baseline predates the change is recaptured, never trusted because its diff passes.
```

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.py` exit 0.
  - [ ] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`
**Tests:** none — a prose rule; `scripts/check-verbatim-moves.sh` fails if the sentence is reworded.
**Regression:** none
**Commit:** docs(visual-verify): a capture under a dimming layer pins its subject structurally
**After:** Task 1

**Decision:** scrim-structural-assertion

- [ ] 3. Testing principles name a measured-bound assertion on a drawn property as a violation
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -c 'coerced to the parent' skills/flow/engineering-principles.md` prints `0`.
  - [ ] **Step 2: Edit.** In `skills/flow/engineering-principles.md` `### Testing principles`, replace the **Violation looks like:** paragraph with exactly this, rewrapped at the file's width (the guard compares whitespace-normalised sentences):

```text verified:applied to engineering-principles.md at 02d83a03; scripts/check-verbatim-moves.sh then prints the old and new forms as FAIL lines identical to verbatim-moves.txt's Task 3 entries
**Violation looks like:** a test asserting that a mock was called rather than what changed, a sleep used for synchronisation, a test whose result depends on execution order or the wall clock, an assertion weakened rather than a bug fixed, a test deleted or `@Ignore`d in the same diff that changed the behavior it covered, or a geometry assertion on a measured or reported bound (Compose's semantics bounds, coerced to the parent's) where the property under test is what is drawn — an overflow is visible only in the rendered pixels, so the assertion passes on the unfixed code.
```

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.py` exit 0.
  - [ ] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/engineering-principles.md`
**Tests:** none — a prose rule; `scripts/check-verbatim-moves.sh` fails if the sentence is reworded.
**Regression:** none
**Commit:** docs(engineering-principles): a measured-bound assertion on a drawn property is vacuous
**After:** none

**Decision:** bounds-violation-in-principles

- [ ] 4. gymie-frontend registers verifyChange and declares its test and lint commands
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/gymie-frontend-worktrees/self-review-backlog-flow-5-oct18b`, run the `## worktree setup` line of its `.flow/project.md` if `local.properties` is missing, then `./gradlew -q verifyChange --dry-run; echo $?`: Gradle reports task `verifyChange` not found and the exit is non-zero; `/Users/tweety53/Projects/agents/scripts/project-get.sh "$PWD" lint; echo $?` reports the section absent.
  - [ ] **Step 2: Apply the fix** in that worktree:

````diff verified:git apply at gymie-frontend 755c74ea; with it applied ./gradlew -q verifyChange --dry-run lists ktlintCheck and detekt for :shared, :androidApp, :desktopApp and :webApp, and ./gradlew verifyChange :shared:desktopTest --tests 'com.gymie.shared.ui.HomeScrollReachUiTest' is BUILD SUCCESSFUL in one invocation
diff --git a/.flow/project.md b/.flow/project.md
index 3613a66d..4b140ed6 100644
--- a/.flow/project.md
+++ b/.flow/project.md
@@ -28,6 +28,29 @@ src="$(git worktree list --porcelain | awk '/^worktree /{print substr($0,10); ex
 `git worktree list --porcelain`. Where the main checkout has none either — a machine relying on
 `ANDROID_HOME` — it copies nothing and exits 0.
 
+## test
+
+```bash
+./gradlew test                                                       # full suite
+./gradlew verifyChange :shared:desktopTest --tests '<selectors>'     # one-invocation verify bundle
+```
+
+## lint
+
+```bash
+./gradlew ktlintFormat                # auto-fix formatting first
+./gradlew verifyChange                # verify — ktlintCheck + detekt across every module, one invocation
+```
+
+**A one-invocation bundle (`verifyChange`, `verify` and their like) holds read-only tasks only.**
+Gradle's parallel executor does not honour, for a task that writes sources, the ordering a dry run
+shows: gymie's kan-630 wired format-before-check four ways, each under the configuration cache and
+without it, and every one still ran a check before the format and went falsely red. A
+source-mutating step — `ktlintFormat` — is a documented pre-step run as its own invocation, as
+above, never a member of the bundle (KAN-780).
+
+Pre-approved suppressions and documented deviations live in `CONTRIBUTING.md`.
+
 ## self review
 
 defer
diff --git a/build.gradle.kts b/build.gradle.kts
index be0448a0..69c79191 100644
--- a/build.gradle.kts
+++ b/build.gradle.kts
@@ -49,4 +49,22 @@ subprojects {
     }
 }
 
+// One composite invocation per verify bundle (KAN-832, porting gymie's KAN-630): ktlintCheck and
+// detekt across every module in one Gradle call — subprojects only, since the root applies both
+// plugins with `apply false` and carries neither task. Scoped tests ride alongside on the same
+// command line — `./gradlew verifyChange :shared:desktopTest --tests '...'` — so a task-close
+// verify is one invocation instead of serialized ones. Deliberately no ktlintFormat here: it
+// mutates sources, and a mutating task riding the same invocation as the checks reorders
+// arbitrarily under the parallel executor (gymie probed it: format-after-check false reds with
+// mustRunAfter wiring in every form that build accepts), while all-read-only tasks cannot race.
+// When verifyChange flags formatting drift, run ktlintFormat and re-run the bundle, as documented
+// in .flow/project.md's ## lint.
+tasks.register("verifyChange") {
+    group = "verification"
+    description = "Run ktlintCheck and detekt across every module, in one invocation."
+    subprojects.forEach { sp ->
+        dependsOn("${sp.path}:ktlintCheck", "${sp.path}:detekt")
+    }
+}
+
 apply(from = "gradle/git-hooks.gradle.kts")
````

  - [ ] **Step 3: Verify.** In that worktree: `./gradlew ktlintFormat`, then `./gradlew verifyChange :shared:desktopTest --tests 'com.gymie.shared.ui.HomeScrollReachUiTest'` is `BUILD SUCCESSFUL`; `/Users/tweety53/Projects/agents/scripts/project-get.sh "$PWD" lint` and `/Users/tweety53/Projects/agents/scripts/project-get.sh "$PWD" test` each exit 0 printing their fence. `git status --short` shows only the two files (`local.properties` is gitignored).
  - [ ] **Step 4: Commit** the two files in gymie-frontend with the subject below.

**Build:** green
**Files:** `build.gradle.kts`, `.flow/project.md`
**Tests:** none — build configuration; `./gradlew -q verifyChange --dry-run` fails without it.
**Regression:** none
**Commit:** build: register the verifyChange lint bundle and declare test and lint commands
**After:** none

**Decision:** verifychange-subprojects
**Decision:** frontend-test-lint-sections

- [ ] 5. flowd replays the journal on every healthy ping
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Save Step 2's diff, apply its test file alone (`git apply --include='stats/internal/reconcile/watch_test.go' <the saved diff>`) and run `cd stats && make -s web-dist && go test ./internal/reconcile -run 'TestWatchReplaysOnHealthyTick' -count=1`: it fails with `Watch never replayed a journal entry written while the store stayed healthy`.
  - [ ] **Step 2: Apply the fix.** Apply the rest (`git apply --exclude='stats/internal/reconcile/watch_test.go' <the saved diff>`):

````diff verified:git apply at 02d83a03; TestWatchReplaysOnHealthyTick fails with reconcile.go unchanged and passes with it; go test ./internal/reconcile -count=1 ok; go vet ./internal/reconcile ./cmd/flowd and gofmt -l clean
diff --git a/skills/flow-contracts/state-file-internals.md b/skills/flow-contracts/state-file-internals.md
index ad9029fc..79dbc735 100644
--- a/skills/flow-contracts/state-file-internals.md
+++ b/skills/flow-contracts/state-file-internals.md
@@ -58,7 +58,7 @@ already stored. Both are safe to retire without further action, which is exactly
 
 The on-disk journal beside the fallback file (`<name>.journal`) holds every write `state set` could
 not deliver, in the order it appended them. The daemon replays it — in file order — at startup and
-whenever it regains a database connection.
+on every successful database ping, every 10 seconds, so a write the CLI journalled while the database stayed up — a request timeout, a non-success answer — lands as surely as one an outage held back.
 
 Each entry is a whole-record write, applied through the same monotonic rule above: conflicts resolve
 by `updatedAt`, with the pipeline state as the tiebreaker, so a `FINISHED` record already in the
diff --git a/stats/cmd/flowd/main.go b/stats/cmd/flowd/main.go
index 817ef57f..13122eed 100644
--- a/stats/cmd/flowd/main.go
+++ b/stats/cmd/flowd/main.go
@@ -52,14 +52,13 @@ const shutdownGrace = 15 * time.Second
 // to matter.
 const harvestInterval = 5 * time.Second
 
-// reconnectPingInterval is how often Reconciler.Watch pings the store to
-// detect design.md's "regains a database connection" transition -- the
-// trigger for a replay beyond the one this daemon already runs once at
-// startup. It is not on the pipeline's own hot path (nothing here blocks a
-// CLI request), so it favors an interval short enough that an outage's
-// journal entries land in the store again soon after the database comes
-// back, without polling aggressively enough to matter against a local
-// Postgres instance.
+// reconnectPingInterval is how often Reconciler.Watch pings the store and,
+// on every ping that succeeds, replays the journal -- the replay beyond the
+// one this daemon already runs once at startup. It is not on the pipeline's
+// own hot path (nothing here blocks a CLI request), so it favors an interval
+// short enough that an outage's journal entries land in the store soon
+// after the database or the daemon's HTTP surface comes back, without
+// polling aggressively enough to matter against a local Postgres instance.
 const reconnectPingInterval = 10 * time.Second
 
 // sweepInterval is how often the abandoned-stage sweeper scans for stage
@@ -177,7 +176,8 @@ func run(logger *slog.Logger) error {
 	// Task 6: replay the write-ahead journal now, at startup -- store.Open
 	// above has already confirmed connectivity via its own ping, so this
 	// is exactly the "at startup" half of design.md's "Availability and
-	// reconciliation". The "on reconnect" half is Watch, started below.
+	// reconciliation". The "on reconnect" half is Watch, started below, which
+	// replays on every healthy ping.
 	// A workspace daemon skips both halves: the journal is shared, and an
 	// entry replayed into its throwaway database leaves the journal and is
 	// dropped with that database at cleanup (KAN-913).
diff --git a/stats/internal/reconcile/reconcile.go b/stats/internal/reconcile/reconcile.go
index fd113b21..c5209e6b 100644
--- a/stats/internal/reconcile/reconcile.go
+++ b/stats/internal/reconcile/reconcile.go
@@ -1,6 +1,6 @@
 // Package reconcile replays the CLI's write-ahead journal
 // (internal/fallback) into the store (internal/store): at daemon startup,
-// whenever the daemon regains a database connection, and on demand via
+// on every healthy ping of Watch's ticker, and on demand via
 // `flow journal flush`. It consumes task 5's journal format and task 2's
 // change repository directly -- it does not go through internal/api's HTTP
 // surface -- but decodes a journal entry's Body with the exact same
@@ -226,19 +226,14 @@ func (r *Reconciler) Run(ctx context.Context) (Result, error) {
 
 // Watch runs until ctx is done, pinging the store every interval via ping
 // and calling Run -- reporting its outcome through onResult, which may be
-// nil -- whenever ping transitions from failing to succeeding. That
-// transition is design.md's "regains a database connection"
-// (`Reconciler.Run` alone only covers the "at startup" half of that
-// requirement; a caller wires Watch for the "on reconnect" half).
-//
-// Watch assumes a connection is already up when it starts: the daemon's
-// own startup sequence already replays once via a direct Run call, after
-// store.Open's own ping has already succeeded, so Watch's first job is
-// detecting the *next* failure-to-success transition, not replaying
-// redundantly for the connection it started with.
+// nil -- on every tick whose ping succeeds. That covers design.md's
+// "regains a database connection" (`Reconciler.Run` alone only covers the
+// "at startup" half; a caller wires Watch for the "on reconnect" half) and
+// also a mark the CLI journalled while the database stayed up -- a brief
+// HTTP-level outage, a request timeout -- which no failing ping ever
+// precedes (KAN-827). Run retires entries idempotently and an empty journal
+// costs one directory walk, so a replay per healthy tick is cheap.
 func (r *Reconciler) Watch(ctx context.Context, ping func(context.Context) error, interval time.Duration, onResult func(Result, error)) {
-	healthy := true
-
 	ticker := time.NewTicker(interval)
 	defer ticker.Stop()
 
@@ -250,17 +245,12 @@ func (r *Reconciler) Watch(ctx context.Context, ping func(context.Context) error
 			pingCtx, cancel := context.WithTimeout(ctx, interval)
 			err := ping(pingCtx)
 			cancel()
-
 			if err != nil {
-				healthy = false
 				continue
 			}
-			if !healthy {
-				healthy = true
-				result, runErr := r.Run(ctx)
-				if onResult != nil {
-					onResult(result, runErr)
-				}
+			result, runErr := r.Run(ctx)
+			if onResult != nil {
+				onResult(result, runErr)
 			}
 		}
 	}
diff --git a/stats/internal/reconcile/watch_test.go b/stats/internal/reconcile/watch_test.go
new file mode 100644
index 00000000..91934fb2
--- /dev/null
+++ b/stats/internal/reconcile/watch_test.go
@@ -0,0 +1,55 @@
+package reconcile_test
+
+import (
+	"context"
+	"testing"
+	"time"
+
+	"github.com/tweety53/agents/stats/internal/reconcile"
+	"github.com/tweety53/agents/stats/internal/store"
+)
+
+// TestWatchReplaysOnHealthyTick pins KAN-827: a mark the CLI journalled
+// during a brief HTTP-level outage, while the database stayed up, is
+// replayed by Watch's next healthy tick -- not only on a ping that goes
+// from failing to succeeding, which such an outage never produces.
+func TestWatchReplaysOnHealthyTick(t *testing.T) {
+	st := newTestStore(t)
+	root := t.TempDir()
+	ctx, cancel := context.WithCancel(context.Background())
+	done := make(chan struct{})
+	// Stop Watch and wait for it before the store's cleanup closes it.
+	defer func() { cancel(); <-done }()
+
+	applied := make(chan reconcile.Result, 1)
+	rec := reconcile.New(st, st, st, root, nil)
+	go func() {
+		defer close(done)
+		rec.Watch(ctx, func(context.Context) error { return nil }, 10*time.Millisecond, func(r reconcile.Result, err error) {
+			if err == nil && r.Applied > 0 {
+				select {
+				case applied <- r:
+				default:
+				}
+			}
+		})
+	}()
+
+	appendEntry(t, root, "proj-watch", "chg-watch", "STARTED", "2026-10-18T10:00:00Z", "flow-start")
+
+	select {
+	case <-applied:
+	case <-time.After(5 * time.Second):
+		t.Fatal("Watch never replayed a journal entry written while the store stayed healthy")
+	}
+	got, err := st.GetChange(context.Background(), "proj-watch", "chg-watch")
+	if err != nil {
+		t.Fatalf("GetChange: %v", err)
+	}
+	if got.State != store.StateStarted {
+		t.Fatalf("stored state = %q, want STARTED", got.State)
+	}
+	if n := pendingCount(t, root, "proj-watch", "chg-watch"); n != 0 {
+		t.Fatalf("pending entries after Watch replay = %d, want 0", n)
+	}
+}
````

  - [ ] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/reconcile -run 'TestWatchReplaysOnHealthyTick' -count=1` passes; `cd stats && go vet ./internal/reconcile ./cmd/flowd && gofmt -l internal/reconcile cmd/flowd` prints nothing; `scripts/check-verbatim-moves.sh` and `scripts/check-references.sh` exit 0.
  - [ ] **Step 4: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/internal/reconcile/reconcile.go`, `stats/internal/reconcile/watch_test.go`, `stats/cmd/flowd/main.go`, `skills/flow-contracts/state-file-internals.md`
**Tests:** `TestWatchReplaysOnHealthyTick`
**Regression:** `TestWatchReplaysOnHealthyTick` — reverting the commit restores the transition-only replay: a journal entry written while the ping stays healthy is never replayed, and the test fails on its deadline.
**Baseline:** before=33 after=34
<!-- measured: cat stats/internal/reconcile/*_test.go | grep -c '^func Test[A-Za-z0-9_]*(t \*testing\.T)' @ 02d83a03 → 33; with watch_test.go added → 34 -->
**Commit:** fix(reconcile): replay the journal on every healthy ping, not only on reconnect
**After:** none

**Decision:** watch-replay-every-healthy-tick
