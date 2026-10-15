# self-review-kan-938-step-3-frontend-cardio-type-picker-on-add-a

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`.
- Every patch below was spiked at `25c1f07c` in this order; apply each verbatim with `git apply`, never retype it. Task 4's prose is acknowledged by exact text in this change's `verbatim-moves.txt` — a reworded sentence fails `scripts/check-verbatim-moves.sh`.
- Go in `stats/`: run `cd stats && gofmt -w <dirs>` before `gofmt -l .`.

- [ ] 1. close-task ends the implementer's dispatch row under the canonical worktree
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n '"record", "dispatch", "end"' stats/internal/guard/closetask.go` shows the call built with no `"-C"`, beside `record decisions` and `tasks tick`, which pass `-C`, canonical — the defect (F1).
  - [ ] **Step 2: Apply the patch.** Save the block below verbatim to `/tmp/self-review-kan-938-step-3-frontend-cardio-type-picker-on-add-a-1.patch` and run `git -C <worktree> apply /tmp/self-review-kan-938-step-3-frontend-cardio-type-picker-on-add-a-1.patch`; it must apply cleanly on the previous task's commit.

````diff verified:git apply in chain order from 25c1f07c, then the task's targeted tests, gofmt -l and go vet green; the new test fails with the fix removed
diff --git a/stats/internal/guard/close_task_test.go b/stats/internal/guard/close_task_test.go
index b7b7f2f3..3ea8969d 100644
--- a/stats/internal/guard/close_task_test.go
+++ b/stats/internal/guard/close_task_test.go
@@ -128,7 +128,7 @@ func TestCloseTaskOrder(t *testing.T) {
 			}
 		}
 		ctEqual(t, fx.calls(t), []string{
-			"flow record dispatch end -change demo -key task-1-implementer -session-token mf-t -commit " + s2 + " -outcome completed",
+			"flow record dispatch end -change demo -C " + fx.wt + " -key task-1-implementer -session-token mf-t -commit " + s2 + " -outcome completed",
 			"flow record decisions -change demo -C " + fx.wt,
 			"flow tasks tick -C " + fx.wt + " demo 1",
 			"git -C " + fx.wt + " push origin spectre/demo",
@@ -246,7 +246,7 @@ func TestCloseTaskExitContract(t *testing.T) {
 				return []string{"-end-key", "k", "-session-token", "mf-t", fx.wt, "demo", fx.base, "1:" + s1}
 			},
 			0, func(fx *ctFx) []string {
-				return []string{"flow record dispatch end -change demo -key k -session-token mf-t -outcome completed",
+				return []string{"flow record dispatch end -change demo -C " + fx.wt + " -key k -session-token mf-t -outcome completed",
 					"flow tasks tick -C " + fx.wt + " demo 1", "git -C " + fx.wt + " push origin spectre/demo"}
 			}},
 		{"task argument without a sha", true, "0", "0",
diff --git a/stats/internal/guard/closetask.go b/stats/internal/guard/closetask.go
index 944df6eb..695c2a0a 100644
--- a/stats/internal/guard/closetask.go
+++ b/stats/internal/guard/closetask.go
@@ -71,7 +71,7 @@ func closeTask(args []string, env Env, stdout, stderr io.Writer) int {
 
 	// A record never blocks: a failed dispatch end is reported, never a stop.
 	if *endKey != "" {
-		rec := []string{"record", "dispatch", "end", "-change", name, "-key", *endKey, "-session-token", *token}
+		rec := []string{"record", "dispatch", "end", "-change", name, "-C", canonical, "-key", *endKey, "-session-token", *token}
 		if *endCommit != "" {
 			rec = append(rec, "-commit", *endCommit)
 		}
````

  - [ ] **Step 3: Verify.** `cd stats && gofmt -w internal/guard cmd/flow && gofmt -l . && go vet ./...`, then `cd stats && go test ./internal/guard -run 'TestCloseTask' -count=1` — exit 0. Commit with the subject below, no `Task-Id:` trailer.

The patch's test change is the two existing call-log expectations in `TestCloseTaskOrder` and `TestCloseTaskExitContract`, which now expect `-C <wt>`; reverting the fix fails both.

**Build:** green
**Files:** `stats/internal/guard/closetask.go`, `stats/internal/guard/close_task_test.go`
**Tests:** none — repairs existing tests: TestCloseTaskOrder, TestCloseTaskExitContract
**Regression:** none — the task adds no test; two existing expectations change
**Baseline:** before=3 after=3
<!-- measured: grep -c '^func Test' stats/internal/guard/close_task_test.go @ 25c1f07c -->
**Commit:** fix(guard): close-task ends the dispatch row under the canonical worktree
**After:** none

**Decision:** close-task-end-under-canonical

- [ ] 2. Signal four's message and header name the panel reviewer key it counts
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `stats/internal/guard/unfinishedwork.go`, signal four sets `lastReview` only when `strings.HasPrefix(d.Key, "panel-")`, while its unreviewed message reads `no reviewer row begun after its implementer row` and `scripts/check-unfinished-work.sh`'s SIGNAL FOUR header reads `any reviewer row of the change counts` — the defect (F2).
  - [ ] **Step 2: Apply the patch.** Save the block below verbatim to `/tmp/self-review-kan-938-step-3-frontend-cardio-type-picker-on-add-a-2.patch` and run `git -C <worktree> apply /tmp/self-review-kan-938-step-3-frontend-cardio-type-picker-on-add-a-2.patch`; it must apply cleanly on the previous task's commit.

````diff verified:git apply in chain order from 25c1f07c, then the task's targeted tests, gofmt -l and go vet green; the new test fails with the fix removed
diff --git a/scripts/check-unfinished-work.sh b/scripts/check-unfinished-work.sh
index d12248c7..5dbd0506 100755
--- a/scripts/check-unfinished-work.sh
+++ b/scripts/check-unfinished-work.sh
@@ -59,9 +59,11 @@
 # a row names a task by its `taskId` or by the `+`-joined ids of a
 # `task-<ids>[-…]` key. Plan-time tasks run grouped or inline and carry no row
 # of their own, so they are never judged.
-# A named appended task also needs a `reviewer` row whose seq is greater than
-# the earliest non-reviewer row naming it: the fix run's panel reviews it
-# under keys that name no task, so any reviewer row of the change counts.
+# A named appended task also needs a panel reviewer row — role `reviewer`, key
+# `panel-<round>-<slot>` — whose seq is greater than the earliest non-reviewer
+# row naming it: the fix run's panel reviews it under keys that name no task,
+# so any panel reviewer row of the change counts, and a reviewer row under any
+# other key does not.
 #
 # All signals are counted independently and reported together on the one
 # line, so the operator sees the whole picture in one prompt rather than being
diff --git a/stats/internal/guard/check_unfinished_work_test.go b/stats/internal/guard/check_unfinished_work_test.go
index ca1be141..6fa15d53 100644
--- a/stats/internal/guard/check_unfinished_work_test.go
+++ b/stats/internal/guard/check_unfinished_work_test.go
@@ -291,7 +291,7 @@ func TestCheckUnfinishedWork(t *testing.T) {
 		uwVerdict("a +-joined key covers the appended task it names", "CLEAR:"))
 	each(demo(planned(appended, `[{"seq":1,"key":"task-7-implementer","role":"implementer","outcome":"completed"}]`)),
 		uwVerdict("an appended task with an implementer row and no reviewer row is OUTSTANDING", "OUTSTANDING:"),
-		uwReason("the breakdown names the unreviewed appended task", "1 ticked appended task(s) with no reviewer row begun after its implementer row (Task 7)"))
+		uwReason("the breakdown names the unreviewed appended task", "1 ticked appended task(s) with no panel reviewer row (key `panel-<round>-<slot>`) begun after its implementer row (Task 7)"))
 	each(demo(planned(appended, `[{"seq":1,"key":"panel-1-primary","role":"reviewer","outcome":"clean"},{"seq":2,"key":"task-7-implementer","role":"implementer","outcome":"completed"}]`)),
 		uwVerdict("a reviewer row begun before the implementer row does not review the appended task", "OUTSTANDING:"))
 	each(demo(planned(appended, `[{"seq":1,"key":"task-7-implementer","role":"implementer","outcome":"completed"},{"seq":2,"key":"pipeline-fix-1-review-1","role":"reviewer","outcome":"clean"}]`)),
diff --git a/stats/internal/guard/unfinishedwork.go b/stats/internal/guard/unfinishedwork.go
index e18fcd36..bf32396f 100644
--- a/stats/internal/guard/unfinishedwork.go
+++ b/stats/internal/guard/unfinishedwork.go
@@ -349,7 +349,7 @@ func checkUnfinishedWork(args []string, env Env, stdout, stderr io.Writer) int {
 				reasons = append(reasons, strconv.Itoa(len(missing))+" ticked appended task(s) with no dispatch row (Task "+strings.Join(missing, ", ")+") — record each task's dispatch, and its fix round's pass rows and findings, before landing")
 			}
 			if len(unreviewed) > 0 {
-				reasons = append(reasons, strconv.Itoa(len(unreviewed))+" ticked appended task(s) with no reviewer row begun after its implementer row (Task "+strings.Join(unreviewed, ", ")+") — run its fix round's panel and record its pass rows and findings, before landing")
+				reasons = append(reasons, strconv.Itoa(len(unreviewed))+" ticked appended task(s) with no panel reviewer row (key `panel-<round>-<slot>`) begun after its implementer row (Task "+strings.Join(unreviewed, ", ")+") — run its fix round's panel and record its pass rows and findings, before landing")
 			}
 		}
 	}
````

  - [ ] **Step 3: Verify.** `cd stats && gofmt -w internal/guard cmd/flow && gofmt -l . && go vet ./...`, then `cd stats && go test ./internal/guard -run 'TestCheckUnfinishedWork' -count=1` and `scripts/check-references.sh` — exit 0. Commit with the subject below, no `Task-Id:` trailer.

The patch updates the existing `uwReason("the breakdown names the unreviewed appended task", …)` expectation, which fails with the old message.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `scripts/check-unfinished-work.sh`
**Tests:** none — repairs existing tests: TestCheckUnfinishedWork
**Regression:** none — the task adds no test; one existing expectation changes
**Baseline:** before=1 after=1
<!-- measured: grep -c '^func Test' stats/internal/guard/check_unfinished_work_test.go @ 25c1f07c -->
**Commit:** fix(guard): signal four's message names the panel reviewer key it counts
**After:** none

**Decision:** signal-four-names-panel-key

- [ ] 3. close-task warns when it closes an appended task no dispatch row names
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'dispatches' stats/internal/guard/closetask.go` prints nothing: close-task never reads dispatch rows, so an appended task closes silently with none (F6).
  - [ ] **Step 2: Apply the patch.** Save the block below verbatim to `/tmp/self-review-kan-938-step-3-frontend-cardio-type-picker-on-add-a-3.patch` and run `git -C <worktree> apply /tmp/self-review-kan-938-step-3-frontend-cardio-type-picker-on-add-a-3.patch`; it must apply cleanly on the previous task's commit.

````diff verified:git apply in chain order from 25c1f07c, then the task's targeted tests, gofmt -l and go vet green; the new test fails with the fix removed
diff --git a/stats/internal/guard/close_task_test.go b/stats/internal/guard/close_task_test.go
index 3ea8969d..442da88b 100644
--- a/stats/internal/guard/close_task_test.go
+++ b/stats/internal/guard/close_task_test.go
@@ -35,8 +35,8 @@ var ctPlan = bt(`- [ ] 1. One
 `)
 
 type ctFx struct {
-	wt, peer, base, log, decisions string
-	env                            Env
+	wt, peer, base, log, decisions, dispatches string
+	env                                        Env
 }
 
 // ctRepo is a repository on spectre/demo with one base commit and a bare
@@ -57,17 +57,18 @@ func ctRepo(t *testing.T, dir string, origin bool) {
 // ctFixture builds the canonical worktree (plan on disk, never committed),
 // a peer repository, and the flow stub: tickRC and recordRC are the exit
 // codes of `flow tasks tick` and `flow record …`; `flow record decisions`
-// prints fx.decisions when a test wrote it, nothing otherwise.
+// and `flow record dispatches` print fx.decisions and fx.dispatches when a
+// test wrote them, nothing otherwise.
 func ctFixture(t *testing.T, origin bool, tickRC, recordRC string) *ctFx {
 	t.Helper()
 	root := t.TempDir()
-	fx := &ctFx{wt: root + "/wt", peer: root + "/peer", log: root + "/calls.log", decisions: root + "/decisions.json"}
+	fx := &ctFx{wt: root + "/wt", peer: root + "/peer", log: root + "/calls.log", decisions: root + "/decisions.json", dispatches: root + "/dispatches.json"}
 	ctRepo(t, fx.wt, origin)
 	ctRepo(t, fx.peer, origin)
 	writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", ctPlan)
 	fx.base = tcfGit(t, fx.wt, "rev-parse", "HEAD")
 	bin := root + "/bin"
-	writeExec(t, bin+"/flow", "#!/bin/sh\necho \"flow $*\" >> "+fx.log+"\ncase \"$1 $2\" in\n\"tasks tick\") exit "+tickRC+";;\n\"record decisions\") cat "+fx.decisions+" 2>/dev/null; exit "+recordRC+";;\nrecord*) exit "+recordRC+";;\nesac\n")
+	writeExec(t, bin+"/flow", "#!/bin/sh\necho \"flow $*\" >> "+fx.log+"\ncase \"$1 $2\" in\n\"tasks tick\") exit "+tickRC+";;\n\"record decisions\") cat "+fx.decisions+" 2>/dev/null; exit "+recordRC+";;\n\"record dispatches\") cat "+fx.dispatches+" 2>/dev/null; exit "+recordRC+";;\nrecord*) exit "+recordRC+";;\nesac\n")
 	writeExec(t, bin+"/git", "#!/bin/sh\ncase \" $* \" in *\" push \"*) echo \"git $*\" >> "+fx.log+";; esac\nexec "+fixtureGit+" \"$@\"\n")
 	fx.env = Env{Getenv: pathEnv(bin), Dir: fx.wt}
 	return fx
@@ -356,3 +357,38 @@ func TestCloseTaskTaskIdTrailer(t *testing.T) {
 		})
 	}
 }
+
+// Signal four at close time: closing an appended task no dispatch row names
+// warns, never blocks; a plan-time task never reads the store at all.
+func TestCloseTaskWarnsOnAnUndispatchedAppendedTask(t *testing.T) {
+	t.Parallel()
+	for _, c := range []struct {
+		name, rows string
+		warn       bool
+	}{
+		{"an appended task no row names warns", `[{"key":"task-1-implementer","taskId":""}]`, true},
+		{"an appended task its key names is silent", `[{"key":"task-2-implementer","taskId":""}]`, false},
+		{"an appended task its taskId names is silent", `[{"key":"impl","taskId":"2"}]`, false},
+	} {
+		t.Run(c.name, func(t *testing.T) {
+			t.Parallel()
+			fx := ctFixture(t, true, "0", "0")
+			writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", "> **Tasks appended:** 2\n\n"+ctPlan)
+			writeFile(t, fx.dispatches, c.rows)
+			s2 := fx.commit(t, fx.wt, "two", "2", "b.txt", "b\n")
+			r := fx.run(fx.wt, "demo", fx.base, "2:"+s2)
+			if r.rc != 0 {
+				t.Fatalf("exit %d, want 0 — a warning never blocks\n%s", r.rc, r.out)
+			}
+			const want = "close-task: WARNING — appended task 2 has no dispatch row naming it"
+			if got := strings.Contains(r.err, want); got != c.warn {
+				t.Errorf("warning present %v, want %v\n%s", got, c.warn, r.out)
+			}
+			ctEqual(t, fx.calls(t), []string{
+				"flow record dispatches -change demo -C " + fx.wt,
+				"flow tasks tick -C " + fx.wt + " demo 2",
+				"git -C " + fx.wt + " push origin spectre/demo",
+			})
+		})
+	}
+}
diff --git a/stats/internal/guard/closetask.go b/stats/internal/guard/closetask.go
index 695c2a0a..f085a2cb 100644
--- a/stats/internal/guard/closetask.go
+++ b/stats/internal/guard/closetask.go
@@ -2,6 +2,7 @@ package guard
 
 import (
 	"bytes"
+	"encoding/json"
 	"flag"
 	"fmt"
 	"io"
@@ -80,6 +81,18 @@ func closeTask(args []string, env Env, stdout, stderr io.Writer) int {
 		}
 	}
 
+	// SIGNAL FOUR at close time (scripts/check-unfinished-work.sh): an
+	// appended task no dispatch row names is warned about here, while its
+	// row can still be recorded, rather than first refused at integrate. A
+	// warning, never a stop: a record never blocks.
+	var closed []string
+	for _, t := range tasks {
+		closed = append(closed, t.id)
+	}
+	for _, id := range ctUndispatched(env, canonical, name, closed) {
+		fmt.Fprintf(stderr, "close-task: WARNING — appended task %s has no dispatch row naming it; record its dispatch (`flow record dispatch begin … -task %s`) before integrate, whose unfinished-work gate refuses it\n", id, id)
+	}
+
 	// Both commit guards run before anything is ticked or pushed, so a
 	// refused commit is never pushed (design.md: close-task-push-after-guards).
 	worst := 0
@@ -229,3 +242,49 @@ func ctCheckTrailer(env Env, stdout, stderr io.Writer, canonical, name, id, comm
 	}
 	return rc
 }
+
+// ctUndispatched is signal four's predicate (uwAppended, uwRowTasks) at
+// close time: the closed tasks the plan's **Tasks appended:** line marks as
+// appended that no dispatch row of the change names. Nil when no closed task
+// is appended -- the store is then never read -- or when the plan or the
+// store cannot be read: a warning that cannot judge stays silent.
+func ctUndispatched(env Env, canonical, name string, closed []string) []string {
+	tasksMD, _ := tcfResolveTasksMD(env, io.Discard, "close-task", canonical, name, canonical)
+	if tasksMD == "" {
+		return nil
+	}
+	appended, ok := uwAppended(tasksMD, false)
+	if !ok {
+		return nil
+	}
+	var judged []string
+	for _, id := range closed {
+		if slices.Contains(appended, id) {
+			judged = append(judged, id)
+		}
+	}
+	if len(judged) == 0 {
+		return nil
+	}
+	out, _, rc := pcFlow(env, false, "record", "dispatches", "-change", name, "-C", canonical)
+	var rows []struct {
+		Key    string `json:"key"`
+		TaskID string `json:"taskId"`
+	}
+	if rc != 0 || json.Unmarshal(out, &rows) != nil {
+		return nil
+	}
+	named := map[string]bool{}
+	for _, d := range rows {
+		for _, id := range uwRowTasks(d.Key, d.TaskID) {
+			named[id] = true
+		}
+	}
+	var missing []string
+	for _, id := range judged {
+		if !named[id] {
+			missing = append(missing, id)
+		}
+	}
+	return missing
+}
diff --git a/stats/internal/guard/unfinishedwork.go b/stats/internal/guard/unfinishedwork.go
index bf32396f..fd2d2c9c 100644
--- a/stats/internal/guard/unfinishedwork.go
+++ b/stats/internal/guard/unfinishedwork.go
@@ -323,17 +323,13 @@ func checkUnfinishedWork(args []string, env Env, stdout, stderr io.Writer) int {
 					}
 					continue
 				}
-				ids := []string{d.TaskID}
-				if m := uwTaskKey.FindStringSubmatch(d.Key); m != nil {
-					ids = append(ids, strings.Split(m[1], "+")...)
-				}
-				for _, id := range ids {
+				for _, id := range uwRowTasks(d.Key, d.TaskID) {
 					if s, seen := first[id]; id != "" && (!seen || d.Seq < s) {
 						first[id] = d.Seq
 					}
 				}
 			}
-			appended, ok := uwAppendedTicked(primary)
+			appended, ok := uwAppended(primary, true)
 			if !ok {
 				return unreadable(primary)
 			}
@@ -385,12 +381,23 @@ var (
 	uwTaskKey       = regexp.MustCompile(`^task-([0-9]+(?:\+[0-9]+)*)(?:-|$)`)
 )
 
-// uwAppendedTicked returns the ids of every one of plan's ticked appended
-// tasks — the last n column-0 tasks, n from the `**Tasks appended:** <n>`
-// header line (blockquoted or not), 0 when absent — in plan order, whatever
-// rows name them: the caller judges coverage. False when the plan cannot be
-// read.
-func uwAppendedTicked(plan string) ([]string, bool) {
+// uwRowTasks is the task ids a dispatch row names: its taskId, and the
+// `+`-joined ids of a task-<ids>[-…] key. Signal four and close-task's
+// close-time warning share it, so the two cannot drift.
+func uwRowTasks(key, taskID string) []string {
+	ids := []string{taskID}
+	if m := uwTaskKey.FindStringSubmatch(key); m != nil {
+		ids = append(ids, strings.Split(m[1], "+")...)
+	}
+	return ids
+}
+
+// uwAppended returns the ids of plan's appended tasks — the last n
+// column-0 tasks, n from the `**Tasks appended:** <n>` header line
+// (blockquoted or not), 0 when absent — in plan order, only the ticked ones
+// when tickedOnly, whatever rows name them: the caller judges coverage.
+// False when the plan cannot be read.
+func uwAppended(plan string, tickedOnly bool) ([]string, bool) {
 	b, err := os.ReadFile(plan)
 	if err != nil {
 		return nil, false
@@ -404,13 +411,13 @@ func uwAppendedTicked(plan string) ([]string, bool) {
 			tasks = append(tasks, m)
 		}
 	}
-	var missing []string
+	var ids []string
 	for _, m := range tasks[max(0, len(tasks)-appended):] {
-		if m[1] != " " {
-			missing = append(missing, m[2])
+		if !tickedOnly || m[1] != " " {
+			ids = append(ids, m[2])
 		}
 	}
-	return missing, true
+	return ids, true
 }
 
 // uwCountUnticked is count_unticked: column-0 `- [ ]` lines (the script
````

  - [ ] **Step 3: Verify.** `cd stats && gofmt -w internal/guard cmd/flow && gofmt -l . && go vet ./...`, then `cd stats && go test ./internal/guard -run 'TestCloseTask|TestCheckUnfinishedWork' -count=1` — exit 0. Commit with the subject below, no `Task-Id:` trailer.

**Build:** green
**Files:** `stats/internal/guard/closetask.go`, `stats/internal/guard/close_task_test.go`, `stats/internal/guard/unfinishedwork.go`
**Tests:** `TestCloseTaskWarnsOnAnUndispatchedAppendedTask`
**Regression:** `TestCloseTaskWarnsOnAnUndispatchedAppendedTask` — reverting the commit removes `ctUndispatched`: no `record dispatches` call is logged and no warning is printed, so all three subtests fail.
**Baseline:** before=3 after=4
<!-- measured: grep -c '^func Test' stats/internal/guard/close_task_test.go @ 25c1f07c -->
**Commit:** feat(guard): close-task warns on an appended task no dispatch row names
**After:** Task 1, 2

**Decision:** close-task-warns-undispatched

- [ ] 4. A cross-repo visual-verify re-run scopes specs by the frontend worktree's own round base
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `skills/flow/visual-verify-verifier.md` step 7 reads `or for the round base, when the prompt names one`, and `skills/flow/verify-fix-loop.md` step 4 reads `its prompt names that base` — neither says whose round base `specs` takes when the dispatched worktree is not `<frontend-root>` (F3).
  - [ ] **Step 2: Apply the patch.** Save the block below verbatim to `/tmp/self-review-kan-938-step-3-frontend-cardio-type-picker-on-add-a-4.patch` and run `git -C <worktree> apply /tmp/self-review-kan-938-step-3-frontend-cardio-type-picker-on-add-a-4.patch`; it must apply cleanly on the previous task's commit.

````diff verified:git apply in chain order from 25c1f07c; check-verbatim-moves.sh, check-references.sh, check-installed-citations.sh and check-markdown-integrity.py exit 0 with this change's verbatim-moves.txt
diff --git a/skills/flow/verify-fix-loop.md b/skills/flow/verify-fix-loop.md
index ded3a709..50a9b95c 100644
--- a/skills/flow/verify-fix-loop.md
+++ b/skills/flow/verify-fix-loop.md
@@ -52,9 +52,11 @@ One round runs these steps in order:
    passed on that same HEAD. **The `visual-verify`
    re-run is scoped to the round's own diff** in every worktree with a round base — the HEAD step
    12 of **Steps 3–13** (`skills/flow/visual-verify.md`) read after that worktree's previous
-   verifier dispatch: its prompt names that base, which the verifier's step 7 substitutes for
-   `specs`' `<merge-base>`, and its `motions:` names only the motions `git diff <round
-   base>..HEAD` adds, changes or removes, plus every motion this round's defects name. A worktree
+   verifier dispatch: its `motions:` names only the motions `git diff <round
+   base>..HEAD` adds, changes or removes, plus every motion this round's defects name. Its prompt
+   names the `<frontend-root>` worktree's own round base, which the verifier's step 7 substitutes
+   for `specs`' `<merge-base>` — never another worktree's of the change; when the
+   `<frontend-root>` worktree has no round base, the prompt names none and `specs` runs unscoped. A worktree
    whose round diff is empty is not dispatched. A worktree with no round base — no verifier of
    this run has verified it, or the parent no longer holds the sha — runs the stage unscoped.
    The confirming re-run
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index 15fe9c0d..8d10cfb3 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -7,7 +7,8 @@ motion step and the `## Report` template, run as written.
 4. **Run `setup`, if declared.** A non-zero exit blocks, printing the command verbatim.
 7. **Run `verify`.** When `specs` is declared, run it first from the `regression checkout`, with
    `<frontend-root>` substituted for the worktree and `<merge-base>` for its merge base — or for
-   the round base, when the prompt names one — and substitute the spec paths it prints, space-separated, for `verify`'s `<specs>`: the mapped specs
+   the `<frontend-root>` worktree's own round base, when the prompt names one, never another
+   worktree's — and substitute the spec paths it prints, space-separated, for `verify`'s `<specs>`: the mapped specs
    replace the checked-in suite, never extend it. An absent `specs`, an empty output or a non-zero
    exit substitutes nothing, and `verify` runs its whole suite. A `verify` with no `<specs>` runs as
    declared, and with `specs` declared the report names `specs: not passed — verify has no <specs>`.
````

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-installed-citations.sh`, `scripts/check-markdown-integrity.py` — all exit 0. Commit with the subject below, no `Task-Id:` trailer.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`, `skills/flow/verify-fix-loop.md`
**Tests:** **none** — prose; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): a visual-verify re-run scopes specs by the frontend worktree's own round base
**After:** none

**Decision:** frontend-root-round-base

- [ ] 5. flow record dispatch end refuses an outcome outside the contracts' set
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n '"outcome", ""' stats/cmd/flow/record.go` shows `-outcome` registered as free text (`e.g. completed`) with no check in `runRecordDispatchEnd` — the defect (F8).
  - [ ] **Step 2: Apply the patch.** Save the block below verbatim to `/tmp/self-review-kan-938-step-3-frontend-cardio-type-picker-on-add-a-5.patch` and run `git -C <worktree> apply /tmp/self-review-kan-938-step-3-frontend-cardio-type-picker-on-add-a-5.patch`; it must apply cleanly on the previous task's commit.

````diff verified:git apply in chain order from 25c1f07c, then the task's targeted tests, gofmt -l and go vet green; the new test fails with the fix removed
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index 1fe403b6..b33cdb8a 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -70,6 +70,15 @@ var recordEfforts = []string{"low", "medium", "high", "default"}
 // explain (KAN-510).
 var recordCauses = []string{"environment", "test-failure", "missing-fixture"}
 
+// recordOutcomes is the closed set `dispatch end -outcome` accepts: every
+// outcome a contract closes a dispatch row with -- completed, clean and fix
+// (a gated reviewer's verdict), blocked, aborted, fallback (a model
+// mismatch) and timed-out (a panel slot's breach). A free-text outcome
+// (`done`, `approve`) reads as no contract's word, so it is refused rather
+// than mapped; an empty one stays allowed, and signal three of
+// check-unfinished-work reports the row it leaves open.
+var recordOutcomes = []string{"completed", "clean", "fix", "blocked", "aborted", "fallback", "timed-out"}
+
 // tokenReportKeys maps the hyphenated keys `-tokens` accepts onto the
 // TokenReport field each one fills -- the bag's four figures the ledger's
 // token line renders, spelled the way the ledger itself spells them. The
@@ -1039,7 +1048,7 @@ func runRecordDispatchEnd(ctx context.Context, args []string, stdout, stderr io.
 	key := fset.String("key", "", "the literal label this dispatch's begin call carried (required)")
 	sessionToken := fset.String("session-token", "", "the run's own literal session token, unchanged from the mark that opened the run -- never a shell substitution (required)")
 	commit := fset.String("commit", "", "the commit sha this dispatch produced")
-	outcome := fset.String("outcome", "", "how the dispatch ended, e.g. completed")
+	outcome := fset.String("outcome", "", "how the dispatch ended -- one of: "+strings.Join(recordOutcomes, ", "))
 	cause := fset.String("cause", "", "why the outcome is blocked -- one of: "+strings.Join(recordCauses, ", ")+"; required with -outcome blocked, refused with any other outcome")
 	agentID := fset.String("agent-id", "", "the harness's own identifier for the dispatched subagent, where begin could not carry it -- optional, and never clears an identifier begin already recorded")
 	tokens := fset.String("tokens", "", "the token usage a same-session (inline) dispatch spent, as comma-separated key=value pairs over: input, output, cache-read, cache-creation (optional -- for a harness that writes no transcript the harvester could credit the inline row from)")
@@ -1053,6 +1062,10 @@ func runRecordDispatchEnd(ctx context.Context, args []string, stdout, stderr io.
 	) {
 		return 2
 	}
+	if *outcome != "" && !slices.Contains(recordOutcomes, *outcome) {
+		fmt.Fprintf(stderr, "flow: -outcome %q is not one of: %s\n", *outcome, strings.Join(recordOutcomes, ", "))
+		return 2
+	}
 	if *outcome == "blocked" && *cause == "" {
 		fmt.Fprintf(stderr, "flow: -cause is required when -outcome is blocked -- one of: %s\n", strings.Join(recordCauses, ", "))
 		return 2
diff --git a/stats/cmd/flow/record_test.go b/stats/cmd/flow/record_test.go
index 29c0139e..1abd2acb 100644
--- a/stats/cmd/flow/record_test.go
+++ b/stats/cmd/flow/record_test.go
@@ -981,6 +981,39 @@ func TestDispatchEndCauseUnknown(t *testing.T) {
 	}
 }
 
+// TestDispatchEndOutcomeClosedSet pins the closed outcome vocabulary: a
+// free-text outcome no contract names (`done`, `approve`) is refused before
+// the store is contacted, its usage naming the accepted set, and every
+// contract outcome reaches the store.
+func TestDispatchEndOutcomeClosedSet(t *testing.T) {
+	for _, outcome := range []string{"done", "approve", "Completed"} {
+		t.Run("refused="+outcome, func(t *testing.T) {
+			code, _, stderr, contacted, _ := causeEndRun(t, "-outcome", outcome)
+			if code != 2 {
+				t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
+			}
+			if contacted {
+				t.Error("the store was contacted for an outcome outside the set -- it must be refused first")
+			}
+			if !strings.Contains(stderr.String(), "completed, clean, fix, blocked, aborted, fallback, timed-out") {
+				t.Errorf("stderr does not name the accepted outcomes:\n%s", stderr.String())
+			}
+		})
+	}
+	for _, outcome := range recordOutcomes {
+		t.Run("accepted="+outcome, func(t *testing.T) {
+			extra := []string{"-outcome", outcome}
+			if outcome == "blocked" {
+				extra = append(extra, "-cause", "environment")
+			}
+			code, _, stderr, contacted, _ := causeEndRun(t, extra...)
+			if code != 0 || !contacted {
+				t.Fatalf("exit code = %d, contacted = %v, want 0 and the store contacted; stderr:\n%s", code, contacted, stderr.String())
+			}
+		})
+	}
+}
+
 // TestDispatchEndCauseAccepted pins the happy path: `-outcome blocked`
 // with a cause from the set reaches the wire carrying it.
 func TestDispatchEndCauseAccepted(t *testing.T) {
````

  - [ ] **Step 3: Verify.** `cd stats && gofmt -w internal/guard cmd/flow && gofmt -l . && go vet ./...`, then `cd stats && go test ./cmd/flow -run 'TestDispatchEnd' -count=1` — exit 0. Commit with the subject below, no `Task-Id:` trailer.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestDispatchEndOutcomeClosedSet`
**Regression:** `TestDispatchEndOutcomeClosedSet` — reverting the commit removes `recordOutcomes`, so the test fails to compile; with the check alone removed, `done`, `approve` and `Completed` reach the store and the three refusal subtests fail.
**Baseline:** before=99 after=100
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ 25c1f07c -->
**Commit:** fix(flow-cli): dispatch end refuses an outcome outside the contracts' set
**After:** none

**Decision:** dispatch-outcome-closed-set

- [ ] 6. Run-record writes refuse a change no checkout of the -C repository carries
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** From the agents checkout, `flow record dispatch end -change no-such-change -key x -session-token mf-x -outcome completed -addr http://127.0.0.1:1 -timeout 300ms` with `FLOW_STATE_DIR` set to a fresh temporary directory journals the write and exits 0 — the defect (F7): a row for a change this repository never held is accepted under its project key.
  - [ ] **Step 2: Apply the patch.** Save the block below verbatim to `/tmp/self-review-kan-938-step-3-frontend-cardio-type-picker-on-add-a-6.patch` and run `git -C <worktree> apply /tmp/self-review-kan-938-step-3-frontend-cardio-type-picker-on-add-a-6.patch`; it must apply cleanly on the previous task's commit.

````diff verified:git apply in chain order from 25c1f07c, then the task's targeted tests, gofmt -l and go vet green; the new test fails with the fix removed
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index b33cdb8a..bec04f1c 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -11,6 +11,7 @@ import (
 	"io/fs"
 	"net/http"
 	"os"
+	"os/exec"
 	"path/filepath"
 	"slices"
 	"sort"
@@ -957,6 +958,9 @@ func runRecordDispatchBegin(ctx context.Context, args []string, stdout, stderr i
 	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
 		return code
 	}
+	if !requireChangeCheckout(f.dir, f.change, stderr) {
+		return 2
+	}
 	if !requireRecordFlags(stderr,
 		[2]string{"-role", *role},
 		[2]string{"-model", *model},
@@ -1056,6 +1060,9 @@ func runRecordDispatchEnd(ctx context.Context, args []string, stdout, stderr io.
 	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
 		return code
 	}
+	if !requireChangeCheckout(f.dir, f.change, stderr) {
+		return 2
+	}
 	if !requireRecordFlags(stderr,
 		[2]string{"-key", *key},
 		[2]string{"-session-token", *sessionToken},
@@ -1235,6 +1242,9 @@ func runRecordFinding(ctx context.Context, args []string, stdout, stderr io.Writ
 	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
 		return code
 	}
+	if !requireChangeCheckout(f.dir, f.change, stderr) {
+		return 2
+	}
 	if !requireRecordFlags(stderr,
 		[2]string{"-ref", *ref},
 		[2]string{"-slot", *slot},
@@ -1323,6 +1333,9 @@ func runRecordStatus(ctx context.Context, args []string, stdout, stderr io.Write
 	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
 		return code
 	}
+	if !requireChangeCheckout(f.dir, f.change, stderr) {
+		return 2
+	}
 	if !requireRecordFlags(stderr, [2]string{"-ref", *ref}, [2]string{"-status", *status}) {
 		return 2
 	}
@@ -2576,6 +2589,9 @@ func runRecordPass(ctx context.Context, args []string, stdout, stderr io.Writer)
 	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
 		return code
 	}
+	if !requireChangeCheckout(f.dir, f.change, stderr) {
+		return 2
+	}
 	if !requireRecordFlags(stderr, [2]string{"-note", *note}) {
 		return 2
 	}
@@ -2668,3 +2684,34 @@ func resolveRenderKinds(kind string) ([]string, error) {
 	}
 	return nil, fmt.Errorf("-kind %q is not one of: %s, all", kind, strings.Join(records.Kinds(), ", "))
 }
+
+// requireChangeCheckout refuses a run-record write -- a dispatch row, a
+// finding, a finding's status, a pass -- for a change no checkout of the
+// repository holding dir carries: none of its worktrees, the main checkout
+// among them, holds spectre/changes/<change>, spectre/changes/archive/<change>
+// or a /flow-fast run's .superpowers/sdd/<change>. The project key derives
+// from dir, so a write run from the wrong repository -- the agents checkout,
+// a second repo's worktree, a call that dropped -C -- would otherwise file
+// the row under a project the change's own reads never look at. Every
+// worktree is searched, not dir's alone: a /flow parent runs from the main
+// checkout while the change lives only in its worktree until it lands.
+func requireChangeCheckout(dir, change string, stderr io.Writer) bool {
+	out, err := exec.Command("git", "-C", dir, "worktree", "list", "--porcelain").Output()
+	if err != nil {
+		fmt.Fprintf(stderr, "flow: list the checkouts of %s: %v\n", dir, err)
+		return false
+	}
+	for line := range strings.Lines(string(out)) {
+		wt, ok := strings.CutPrefix(strings.TrimRight(line, "\n"), "worktree ")
+		if !ok {
+			continue
+		}
+		for _, root := range []string{"spectre/changes", "spectre/changes/archive", ".superpowers/sdd"} {
+			if fi, err := os.Stat(filepath.Join(wt, root, change)); err == nil && fi.IsDir() {
+				return true
+			}
+		}
+	}
+	fmt.Fprintf(stderr, "flow: no checkout of the repository holding %s carries change %q (spectre/changes/, spectre/changes/archive/ or .superpowers/sdd/) -- refused, so the row is not filed under the wrong project; pass -C <the change's canonical worktree>\n", dir, change)
+	return false
+}
diff --git a/stats/cmd/flow/record_test.go b/stats/cmd/flow/record_test.go
index 1abd2acb..f1129024 100644
--- a/stats/cmd/flow/record_test.go
+++ b/stats/cmd/flow/record_test.go
@@ -58,7 +58,7 @@ func recordJournalEntries(t *testing.T, repo, change string) (entries []fallback
 // would make the client's own `created` flag dead weight.
 func TestRecordWritePrintsOneLineAndExitsZero(t *testing.T) {
 	t.Run("dispatch", func(t *testing.T) {
-		repo := gitRepo(t)
+		repo := recordRepo(t, "kan-258")
 		isolatedStateRoot(t)
 
 		var gotPath string
@@ -133,7 +133,7 @@ func TestRecordWritePrintsOneLineAndExitsZero(t *testing.T) {
 	})
 
 	t.Run("finding created", func(t *testing.T) {
-		repo := gitRepo(t)
+		repo := recordRepo(t, "kan-258")
 		isolatedStateRoot(t)
 
 		var gotBody []byte
@@ -176,7 +176,7 @@ func TestRecordWritePrintsOneLineAndExitsZero(t *testing.T) {
 	})
 
 	t.Run("finding updated", func(t *testing.T) {
-		repo := gitRepo(t)
+		repo := recordRepo(t, "kan-258")
 		isolatedStateRoot(t)
 
 		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
@@ -201,7 +201,7 @@ func TestRecordWritePrintsOneLineAndExitsZero(t *testing.T) {
 	})
 
 	t.Run("status", func(t *testing.T) {
-		repo := gitRepo(t)
+		repo := recordRepo(t, "kan-258")
 		isolatedStateRoot(t)
 
 		var gotMethod, gotPath string
@@ -288,7 +288,7 @@ func TestRecordWriteFallsBackToJournalAndExitsZero(t *testing.T) {
 
 	for _, tc := range cases {
 		t.Run(tc.name, func(t *testing.T) {
-			repo := gitRepo(t)
+			repo := recordRepo(t, "kan-258")
 			isolatedStateRoot(t)
 
 			// The connection flags are spliced in after the verb words
@@ -387,7 +387,7 @@ func TestRecordDispatchJournalCarriesTheAttemptInstant(t *testing.T) {
 
 	for _, tc := range cases {
 		t.Run(tc.name, func(t *testing.T) {
-			repo := gitRepo(t)
+			repo := recordRepo(t, "kan-258")
 			isolatedStateRoot(t)
 
 			verbLen := 0
@@ -461,7 +461,7 @@ func TestRecordMissingRequiredFlagExitsTwoWithoutJournalling(t *testing.T) {
 
 	for _, tc := range cases {
 		t.Run(tc.name, func(t *testing.T) {
-			repo := gitRepo(t)
+			repo := recordRepo(t, "kan-258")
 			isolatedStateRoot(t)
 
 			args := append([]string{}, tc.args[:2]...)
@@ -486,7 +486,7 @@ func TestRecordMissingRequiredFlagExitsTwoWithoutJournalling(t *testing.T) {
 // `stage begin` checks its stage key: were it checked after, the fallback
 // would swallow the caller's own mistake as if it were a store outage.
 func TestRecordRejectsUnknownRoleWithoutContactingStore(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	contacted := false
@@ -525,7 +525,7 @@ func TestRecordRejectsUnknownRoleWithoutContactingStore(t *testing.T) {
 // role `/flow` records for the subagent it dispatches to plan
 // brainstorm.md sections B-D and flow.document-fix.
 func TestRecordAcceptsPlannerRole(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	contacted := false
@@ -556,7 +556,7 @@ func TestRecordAcceptsPlannerRole(t *testing.T) {
 // that runs implement.md, review-panel.md and verify-and-handoff.md as
 // the conductor (design.md's conductor-runs-implementation-half).
 func TestRecordAcceptsConductorRole(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	contacted := false
@@ -587,7 +587,7 @@ func TestRecordAcceptsConductorRole(t *testing.T) {
 // skills/flow/verify-and-handoff.md dispatches for flow.verify and
 // flow.visual-verify.
 func TestRecordAcceptsVerifierRole(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	contacted := false
@@ -619,7 +619,7 @@ func TestRecordAcceptsVerifierRole(t *testing.T) {
 // between no two sessions, so the harvest binding it exists for would bind
 // nothing.
 func TestRecordRejectsSessionTokenSubstitution(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	contacted := false
@@ -694,7 +694,7 @@ func TestRecordDispatchSendsAgentIDOnlyWhenGiven(t *testing.T) {
 
 	send := func(t *testing.T, extra ...string) map[string]any {
 		t.Helper()
-		repo := gitRepo(t)
+		repo := recordRepo(t, "kan-258")
 		isolatedStateRoot(t)
 
 		var gotBody []byte
@@ -752,7 +752,7 @@ func TestRecordDispatchSendsAgentIDOnlyWhenGiven(t *testing.T) {
 // than being written out twice.
 func dispatchBeginBody(t *testing.T, extra ...string) map[string]any {
 	t.Helper()
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-327")
 	isolatedStateRoot(t)
 
 	var gotBody []byte
@@ -840,7 +840,7 @@ func TestDispatchEndAcceptsAgentID(t *testing.T) {
 
 	send := func(t *testing.T, extra ...string) map[string]any {
 		t.Helper()
-		repo := gitRepo(t)
+		repo := recordRepo(t, "kan-258")
 		isolatedStateRoot(t)
 
 		var gotBody []byte
@@ -902,7 +902,7 @@ func TestDispatchEndAcceptsAgentID(t *testing.T) {
 // minimal end request, varying only `-outcome` and `-cause`.
 func causeEndRun(t *testing.T, extra ...string) (int, bytes.Buffer, bytes.Buffer, bool, string) {
 	t.Helper()
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-510")
 	isolatedStateRoot(t)
 
 	contacted := false
@@ -1017,7 +1017,7 @@ func TestDispatchEndOutcomeClosedSet(t *testing.T) {
 // TestDispatchEndCauseAccepted pins the happy path: `-outcome blocked`
 // with a cause from the set reaches the wire carrying it.
 func TestDispatchEndCauseAccepted(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-510")
 	isolatedStateRoot(t)
 
 	var gotBody []byte
@@ -1662,7 +1662,7 @@ func TestRecordJournalCountAbsentJournalCountsZero(t *testing.T) {
 // A count that disagreed would report one more pending write than a replay
 // will ever apply -- at a gate whose whole job is to be believable.
 func TestRecordJournalCountIgnoresPartialTrailingLine(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	projectKey, _, err := fallback.ProjectKey(repo)
@@ -1809,7 +1809,7 @@ func TestCostStatusNeverBlocks(t *testing.T) {
 // change's record its dispatch landed -- and end names its row by the key
 // and token begin carried, never by that seq.
 func TestRecordDispatchBeginPrintsTheAllocatedSeqAndEndClosesIt(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	var endPath string
@@ -1888,7 +1888,7 @@ func TestRecordDispatchBeginPrintsTheAllocatedSeqAndEndClosesIt(t *testing.T) {
 // it would lose the end and leave the window that begin opened open
 // forever, which is the defect the end call exists to prevent.
 func TestRecordDispatchEndJournalsAKeyTheStoreDoesNotKnowYet(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	var beginOut, beginErr bytes.Buffer
@@ -1940,7 +1940,7 @@ func TestRecordDispatchEndJournalsAKeyTheStoreDoesNotKnowYet(t *testing.T) {
 // unreachable" and journalled an entry that 404s on every replay). It is
 // reported, names the status, journals nothing and exits non-zero.
 func TestRecordDispatchEndRefusesA404WithNoQueuedBegin(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
@@ -1971,7 +1971,7 @@ func TestRecordDispatchEndRefusesA404WithNoQueuedBegin(t *testing.T) {
 // exception above: a 404 from any other record write stays definitive, so
 // the exception cannot quietly become the rule.
 func TestRecordFindingStillRefusesAnUnknownRef(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
@@ -2002,7 +2002,7 @@ func TestRecordFindingStillRefusesAnUnknownRef(t *testing.T) {
 // after, the fallback would swallow the caller's own mistake as if it were
 // a store outage.
 func TestRecordFindingRejectsUnknownStatus(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	contacted := false
@@ -2043,7 +2043,7 @@ func TestRecordFindingRejectsUnknownStatus(t *testing.T) {
 func TestRecordFindingRejectsWithdrawnWithNoReason(t *testing.T) {
 	for _, status := range []string{"withdrawn", "withdrawn   "} {
 		t.Run(status, func(t *testing.T) {
-			repo := gitRepo(t)
+			repo := recordRepo(t, "kan-258")
 			isolatedStateRoot(t)
 
 			contacted := false
@@ -2435,7 +2435,7 @@ func TestRecordIncidentsPrintsArray(t *testing.T) {
 // TestRecordFindingRejectsEmptyReproducer pins that an empty -reproducer is
 // refused before the store is contacted.
 func TestRecordFindingRejectsEmptyReproducer(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	contacted := false
@@ -2475,7 +2475,7 @@ func TestRecordFindingRejectsEmptyReproducer(t *testing.T) {
 func TestRecordFindingRejectsBareNoneReproducer(t *testing.T) {
 	for _, reproducer := range []string{"none", "none  "} {
 		t.Run(reproducer, func(t *testing.T) {
-			repo := gitRepo(t)
+			repo := recordRepo(t, "kan-258")
 			isolatedStateRoot(t)
 
 			contacted := false
@@ -2525,7 +2525,7 @@ func TestRecordFindingAcceptsWellFormedStatusAndReproducer(t *testing.T) {
 
 	for _, c := range cases {
 		t.Run(c.name, func(t *testing.T) {
-			repo := gitRepo(t)
+			repo := recordRepo(t, "kan-258")
 			isolatedStateRoot(t)
 
 			contacted := false
@@ -2558,7 +2558,7 @@ func TestRecordFindingAcceptsWellFormedStatusAndReproducer(t *testing.T) {
 // the same status vocabulary check as `record finding`, before the store
 // is contacted.
 func TestRecordStatusRejectsUnknownStatus(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	contacted := false
@@ -2595,7 +2595,7 @@ func TestRecordStatusRejectsUnknownStatus(t *testing.T) {
 func TestRecordStatusRejectsWithdrawnWithNoReason(t *testing.T) {
 	for _, status := range []string{"withdrawn", "withdrawn   "} {
 		t.Run(status, func(t *testing.T) {
-			repo := gitRepo(t)
+			repo := recordRepo(t, "kan-258")
 			isolatedStateRoot(t)
 
 			contacted := false
@@ -2634,7 +2634,7 @@ func TestRecordStatusRejectsWithdrawnWithNoReason(t *testing.T) {
 // carrying ref/status/reproducer -- the only fields a guard consuming this
 // verb needs.
 func TestRecordFindingsPrintsJSONArray(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "demo")
 	isolatedStateRoot(t)
 
 	body := `{"change":"demo","dispatches":[],"findings":[
@@ -2923,7 +2923,7 @@ func TestRunRecordDispatchesUnreachableStoreFails(t *testing.T) {
 // pipeline starts with .[] and crashes on a top-level null under set -euo
 // pipefail.
 func TestRunRecordDispatchesNullDispatchesPrintsEmptyArray(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "demo")
 	isolatedStateRoot(t)
 
 	body := `{"change":"demo","findings":[{"ref":"F1","round":0,"slot":"Bugbot","severity":"Minor","location":"a.go:1","note":"n1","status":"fixed","reproducer":"none — prose only"}],"dispatches":null}`
@@ -3261,7 +3261,7 @@ var roleLiteralRe = regexp.MustCompile(`(?:^|[^A-Za-z])-role\s+([a-z][a-z-]*)`)
 // taking the never-block fallback path would journal a write a replay
 // could only ever be refused for a second time.
 func TestRunRecordDispatchBeginRejectsUnknownEffort(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	contacted := false
@@ -3328,7 +3328,7 @@ func TestRecordDispatchBeginAcceptsEffort(t *testing.T) {
 // to .../passes carrying round and note verbatim, one stdout line naming the
 // allocated row id, exit 0.
 func TestRecordPassCommand(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	var gotPath string
@@ -3440,7 +3440,7 @@ func TestRecordPassAndMutationMissingFlagsExitTwo(t *testing.T) {
 		{"mutation without test", []string{"record", "mutation", "-change", "kan-258", "-path", "p", "-mutated", "none"}},
 	} {
 		t.Run(tc.name, func(t *testing.T) {
-			repo := gitRepo(t)
+			repo := recordRepo(t, "kan-258")
 			isolatedStateRoot(t)
 
 			var stdout, stderr bytes.Buffer
@@ -3544,7 +3544,7 @@ func TestRecordFindingLineageFlags(t *testing.T) {
 
 	for _, c := range cases {
 		t.Run(c.name, func(t *testing.T) {
-			repo := gitRepo(t)
+			repo := recordRepo(t, "kan-507")
 			isolatedStateRoot(t)
 
 			var body records.Finding
@@ -3581,7 +3581,7 @@ func TestRecordFindingLineageFlags(t *testing.T) {
 // reads, and `flow record findings` prints it back. The fake daemon keeps
 // the posted finding and serves it on the run-record read.
 func TestRecordFindingTaskRoundTrips(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-9")
 	isolatedStateRoot(t)
 
 	var stored json.RawMessage
@@ -3651,7 +3651,7 @@ func TestValidateFindingCategory(t *testing.T) {
 // call, category included, so a replay of a categorised deferral re-lands
 // the category instead of clearing it.
 func TestRecordStatusJournalsCategory(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	var stdout, stderr bytes.Buffer
@@ -3694,7 +3694,7 @@ func TestRecordStatusJournalsCategory(t *testing.T) {
 // review panel round 0): -category reaches the daemon's POST body, so a
 // finding recorded already-deferred lands with its category in one write.
 func TestRecordFindingSendsCategory(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	var gotBody []byte
@@ -3735,7 +3735,7 @@ func TestRecordFindingSendsCategory(t *testing.T) {
 // before any network call -- a dead addr proves the refusal, since a fall
 // through to the write path would journal and exit 0.
 func TestRecordVerbsValidateCategoryBeforeTheStore(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-258")
 	isolatedStateRoot(t)
 
 	cases := []struct {
@@ -3785,7 +3785,7 @@ func TestRecordVerbsValidateCategoryBeforeTheStore(t *testing.T) {
 // passes the validator and falls through to the network path, journalling
 // as every unreachable write does.
 func TestRecordFindingRefusesADeferredNonMinorBeforeTheStore(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-508", "kan-508-ok")
 	isolatedStateRoot(t)
 
 	for _, tc := range []struct{ name, severity string }{
@@ -4042,7 +4042,7 @@ func TestRecordSubstitutionsReadback(t *testing.T) {
 // store, while the pattern's normalization stays the store's business --
 // the CLI forwards verbatim what the caller typed.
 func TestRecordFindingPatternFlag(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "kan-416")
 	isolatedStateRoot(t)
 
 	var gotPattern string
@@ -4271,7 +4271,7 @@ func TestDispatchEndTokensParserRefusals(t *testing.T) {
 func TestDispatchEndTokensCarriedOnWire(t *testing.T) {
 	send := func(t *testing.T, extra ...string) map[string]any {
 		t.Helper()
-		repo := gitRepo(t)
+		repo := recordRepo(t, "kan-525")
 		isolatedStateRoot(t)
 
 		var gotBody []byte
@@ -4616,7 +4616,7 @@ func TestRecordHandoffLinesReadsNoFindings(t *testing.T) {
 // the highest round any pass, finding or mutation row carries, 0 for a
 // change the store has never heard of, and no number on a failed read.
 func TestRunRecordNextRound(t *testing.T) {
-	repo := gitRepo(t)
+	repo := recordRepo(t, "demo")
 	isolatedStateRoot(t)
 
 	for _, tc := range []struct {
@@ -4672,7 +4672,7 @@ func TestRecordFindingRejectsMalformedRef(t *testing.T) {
 		{"-regression-of", "KAN-9"},
 	} {
 		t.Run(tc.flag+"="+tc.value, func(t *testing.T) {
-			repo := gitRepo(t)
+			repo := recordRepo(t, "kan-964")
 			isolatedStateRoot(t)
 
 			contacted := false
@@ -4721,7 +4721,7 @@ func TestRecordPassRefusesMarkerNote(t *testing.T) {
 		"  finding-status: F9 fixed",
 	} {
 		t.Run(note, func(t *testing.T) {
-			repo := gitRepo(t)
+			repo := recordRepo(t, "kan-258")
 			isolatedStateRoot(t)
 
 			var stdout, stderr bytes.Buffer
@@ -4741,3 +4741,69 @@ func TestRecordPassRefusesMarkerNote(t *testing.T) {
 		})
 	}
 }
+
+// TestRecordWriteRefusesAChangeNoCheckoutCarries pins the project-key
+// sanity check: a run-record write for a change no checkout of the -C
+// repository carries is refused before the store is contacted, so it never
+// files the row under the wrong project; a change carried by a linked
+// worktree, an archive or a /flow-fast .superpowers/sdd root is accepted
+// from the main checkout.
+func TestRecordWriteRefusesAChangeNoCheckoutCarries(t *testing.T) {
+	for _, c := range []struct {
+		name, root string
+		linked     bool
+		want       int
+	}{
+		{"no checkout carries the change", "", false, 2},
+		{"a linked worktree carries it", "spectre/changes/kan-7", true, 0},
+		{"the archive carries it", "spectre/changes/archive/kan-7", false, 0},
+		{"a flow-fast root carries it", ".superpowers/sdd/kan-7", true, 0},
+	} {
+		t.Run(c.name, func(t *testing.T) {
+			repo := gitRepo(t)
+			isolatedStateRoot(t)
+			holder := repo
+			if c.linked {
+				holder = filepath.Join(t.TempDir(), "wt")
+				if out, err := exec.Command("git", "-C", repo, "worktree", "add", "-q", "-b", "kan-7", holder).CombinedOutput(); err != nil {
+					t.Fatalf("git worktree add: %v\n%s", err, out)
+				}
+			}
+			if c.root != "" {
+				if err := os.MkdirAll(filepath.Join(holder, c.root), 0o755); err != nil {
+					t.Fatal(err)
+				}
+			}
+			contacted := false
+			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
+				contacted = true
+				w.WriteHeader(http.StatusOK)
+				_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"verifier","model":"sonnet","endedAt":"2026-01-02T03:44:05Z"}`))
+			}))
+			defer srv.Close()
+			var stdout, stderr bytes.Buffer
+			code := run(context.Background(), []string{"record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
+				"-change", "kan-7", "-key", "k", "-session-token", "mf-kan7-end", "-outcome", "completed"}, strings.NewReader(""), &stdout, &stderr)
+			if code != c.want || contacted != (c.want == 0) {
+				t.Fatalf("exit %d, contacted %v; want exit %d, contacted %v\nstderr:\n%s", code, contacted, c.want, c.want == 0, stderr.String())
+			}
+			if c.want != 0 && !strings.Contains(stderr.String(), `carries change "kan-7"`) {
+				t.Errorf("stderr does not name the change:\n%s", stderr.String())
+			}
+		})
+	}
+}
+
+// recordRepo is gitRepo carrying spectre/changes/<change> for every change
+// named: a run-record write refuses a change no checkout of its repository
+// carries (requireChangeCheckout).
+func recordRepo(t *testing.T, changes ...string) string {
+	t.Helper()
+	repo := gitRepo(t)
+	for _, c := range changes {
+		if err := os.MkdirAll(filepath.Join(repo, "spectre", "changes", c), 0o755); err != nil {
+			t.Fatal(err)
+		}
+	}
+	return repo
+}
````

  - [ ] **Step 3: Verify.** `cd stats && gofmt -w internal/guard cmd/flow && gofmt -l . && go vet ./...`, then `cd stats && go test ./cmd/flow -run 'TestRecord|TestDispatchEnd' -count=1` — exit 0. Commit with the subject below, no `Task-Id:` trailer.

The patch also moves every existing `record_test.go` test that writes through these verbs from `gitRepo(t)` to `recordRepo(t, <its -change literals>)`, which creates `spectre/changes/<change>` for each.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordWriteRefusesAChangeNoCheckoutCarries`
**Regression:** `TestRecordWriteRefusesAChangeNoCheckoutCarries` — reverting the commit lets the write for a change no checkout carries reach the store, so its first subtest fails.
**Baseline:** before=100 after=101
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ 25c1f07c with tasks 1–5 applied -->
**Commit:** feat(flow-cli): run-record writes refuse a change no checkout carries
**After:** Task 5

**Decision:** change-checkout-check
