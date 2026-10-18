# pipeline-fix-kan-urgent-5-oct18-20

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** a loop's satellite branch in another repository lands on the rows the loop filed under
the agents checkout's project; a loop with rows nowhere is still refused.

**Architecture:** one fallback read in `checkLoopLanding` (`stats/internal/guard/looplanding.go`),
fed by `FLOW_GUARD_REPO_ROOT`, which `scripts/check-loop-landing.sh` now exports.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; it was applied to this worktree at fe238ed0 and checked there. A diff that does not
  apply is reported, never hand-merged by guesswork.
- No new inline suppressions; `cd stats && gofmt -l .` prints nothing.
- Go tests in `stats/internal/guard` take about two minutes: run them with a 600000 ms timeout, in
  <!-- measured: cd stats && go test ./internal/guard/ -count=1 @ fe238ed0 with the diff applied — ok in 116.2s -->
  the foreground.
- Read-only against the dev store: never stop flowd or flow-postgres, never drop the `flow` DB.

## Review Focus

- A pushing repository whose project already holds the loop's rows never reads the agents
  project — every existing `TestCheckLoopLanding` case.
- A `spectre/<name>` branch never falls back (`own-branch-no-fallback`).
- `FLOW_GUARD_REPO_ROOT` unset, or naming the pushing repository itself, changes nothing.

- [ ] 1. Read a satellite loop branch's rows under the agents checkout's project
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** From this worktree run
    `scripts/check-loop-landing.sh /Users/tweety53/Projects/gymie self-review-kan-urgent-5-oct18`:
    exit 1, `LOOP-LANDING-REFUSED: PLAN —`, `FIX —`, `REVIEW —` (read-only; the dev store holds
    the finished loop's rows under the agents project). Save the diff below to
    `/tmp/pf20-task1.diff`, apply its test hunk alone
    (`git apply --include='stats/internal/guard/*_test.go' /tmp/pf20-task1.diff`) and run
    `cd stats && go test ./internal/guard/ -run 'TestCheckLoopLandingSatellite' -count=1`:
    `rows_under_the_agents_project` FAILs; the other two subtests pass.
  - [ ] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/internal/guard/*_test.go' /tmp/pf20-task1.diff`):

````diff verified:applied with git apply to this worktree at fe238ed0; TestCheckLoopLandingSatellite/rows_under_the_agents_project fails with looplanding.go untouched and all subtests pass with it; go test ./internal/guard/ -count=1 green; gofmt -l and go vet clean; check-guard-symlinks and check-references exit 0; the live call against gymie exits 0 LOOP-LANDING-OK
diff --git a/scripts/check-loop-landing.sh b/scripts/check-loop-landing.sh
index fd3fd0d1..a437c7df 100755
--- a/scripts/check-loop-landing.sh
+++ b/scripts/check-loop-landing.sh
@@ -5,7 +5,12 @@
 # Only once flowd holds, under the loop's change (its branch name), a
 # completed plan dispatch, a completed first fix (`fix` or `fix-g<g>`), at
 # least one review, every row closed, the highest-numbered review closed
-# `clean` — no Critical, no Important — and no `fix-v<v>` after it.
+# `clean` — no Critical, no Important — and no `fix-v<v>` after it. A
+# loop branch whose repository's project holds no rows for it — a
+# satellite, the loop's branch in another repository it carried a task into —
+# is read under the project of the agents checkout this script sits in, where
+# a loop run from that checkout files them; a FINISHED loop's rows still
+# admit its satellites, which land after its own branch.
 # hooks/protect-main-checkout.py runs it on a landing push typed by hand;
 # pipeline-fix.sh land runs the same check before its own push.
 #
@@ -23,4 +28,6 @@ SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
   echo "check-loop-landing: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
   exit 2
 }
+FLOW_GUARD_REPO_ROOT="$(flow_guard_root)"
+export FLOW_GUARD_REPO_ROOT
 flow_guard_exec check-loop-landing 2 "check-loop-landing:" "$@"
diff --git a/stats/internal/guard/check_loop_landing_test.go b/stats/internal/guard/check_loop_landing_test.go
index f288a3b8..ca24833a 100644
--- a/stats/internal/guard/check_loop_landing_test.go
+++ b/stats/internal/guard/check_loop_landing_test.go
@@ -3,6 +3,7 @@ package guard
 import (
 	"bytes"
 	"fmt"
+	"path/filepath"
 	"strings"
 	"testing"
 )
@@ -86,3 +87,51 @@ func TestCheckLoopLanding(t *testing.T) {
 		})
 	}
 }
+
+// A satellite -- the loop's branch in another repository it carried a task
+// into -- lands on the rows the loop filed under the agents checkout's
+// project (FLOW_GUARD_REPO_ROOT) when its own repository's project has none;
+// a loop with rows nowhere is still refused, and a change's own branch never
+// reads another project's rows.
+func TestCheckLoopLandingSatellite(t *testing.T) {
+	t.Parallel()
+	const sr = "self-review-kan-1-demo"
+	clean := lpRowsJSON(sr+"-", sr, "plan completed", "fix completed", "review-1 clean")
+	for _, tc := range []struct {
+		name, branch, agentsRows string
+		rc                       int
+		out                      string
+	}{
+		{name: "rows under the agents project", branch: sr, agentsRows: clean, out: "LOOP-LANDING-OK"},
+		{name: "rows nowhere", branch: sr, agentsRows: `[]`, rc: 1, out: "PLAN — no completed " + sr + "-plan"},
+		{name: "a change branch reads only its own project", branch: "spectre/kan-1-demo", agentsRows: clean, out: "LOOP-LANDING-NA"},
+	} {
+		t.Run(tc.name, func(t *testing.T) {
+			t.Parallel()
+			fx := lpNew(t, `[]`)
+			sat := filepath.Dir(fx.repo) + "/gymie"
+			gitRun(t, "", "init", "-q", "-b", "main", sat)
+			gitRun(t, sat, "commit", "-q", "--allow-empty", "-m", "init")
+			writeFile(t, fx.rows+".agents", tc.agentsRows)
+			bin := filepath.Dir(fx.repo) + "/bin"
+			writeExec(t, bin+"/flow", `#!/usr/bin/env bash
+case "$*" in
+  *"-C `+fx.repo+`"*) exec cat `+fx.rows+`.agents ;;
+  *) exec cat `+fx.rows+` ;;
+esac
+`)
+			env := fx.env
+			env.Getenv = func(k string) string {
+				if k == "FLOW_GUARD_REPO_ROOT" {
+					return fx.repo
+				}
+				return fx.env.Getenv(k)
+			}
+			var o, e bytes.Buffer
+			rc := checkLoopLanding([]string{sat, tc.branch}, env, &o, &e)
+			if out := o.String() + e.String(); rc != tc.rc || !strings.Contains(out, tc.out) {
+				t.Fatalf("rc %d, want %d; output %q, want it to carry %q", rc, tc.rc, out, tc.out)
+			}
+		})
+	}
+}
diff --git a/stats/internal/guard/looplanding.go b/stats/internal/guard/looplanding.go
index fb296d08..55fc9a50 100644
--- a/stats/internal/guard/looplanding.go
+++ b/stats/internal/guard/looplanding.go
@@ -10,8 +10,10 @@ import (
 
 // check-loop-landing.sh <repo> <branch>: may a pipeline-fix-* or
 // self-review-* branch land on the default branch? Only once flowd holds,
-// for its loop, a completed plan, a completed first fix and a review, every
-// row closed, the highest review `clean` -- no Critical or Important -- and
+// for its loop -- under the pushing repository's project, or, for a loop
+// branch with none there, under the project of the agents checkout this
+// guard is installed from -- a completed plan, a completed first fix and a
+// review, every row closed, the highest review `clean` -- no Critical or Important -- and
 // no red-line fix (`fix-v<v>`) after it. protect-main-checkout.py runs it on
 // a landing push typed by hand; pipeline-fix.sh land runs it before its own.
 func init() {
@@ -129,6 +131,16 @@ func checkLoopLanding(args []string, env Env, stdout, stderr io.Writer) int {
 		fmt.Fprintf(stderr, "%s: flow record dispatches -change %s failed -- cannot answer (is flowd up?)\n", self, lp.change)
 		return 2
 	}
+	// A loop run from the agents checkout files every row under that checkout's
+	// project, so its satellite branches -- the same loop branch in another
+	// repository it carried a task into -- find their rows there.
+	if root := env.Getenv("FLOW_GUARD_REPO_ROOT"); len(rows) == 0 && !lpOwnBranch(args[1]) && root != "" {
+		if agents, ok := lpMain(env, git, root); ok && agents != main {
+			if r, ok := lpRows(env, agents, lp); ok {
+				rows = r
+			}
+		}
+	}
 	if lpOwnBranch(args[1]) && len(rows) == 0 {
 		fmt.Fprintf(stdout, "LOOP-LANDING-NA %s\n", args[1])
 		return 0
````

  - [ ] **Step 3: Verify.** `cd stats && gofmt -l ./internal/guard && go vet ./internal/guard/ && go test ./internal/guard/ -run 'CheckLoopLanding|PipelineFix' -count=1`
    — gofmt prints nothing, the rest exit 0; `scripts/check-guard-symlinks.sh` and
    `scripts/check-references.sh` exit 0; the Step 1 live call now exits 0 with
    `LOOP-LANDING-OK self-review-kan-urgent-5-oct18`, and
    `scripts/check-loop-landing.sh /Users/tweety53/Projects/gymie self-review-no-such-loop` still
    exits 1.
  - [ ] **Step 4: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/looplanding.go`, `stats/internal/guard/check_loop_landing_test.go`, `scripts/check-loop-landing.sh`
**Tests:** `TestCheckLoopLandingSatellite`
**Regression:** reverting the commit reads the satellite's rows under the pushing repository's project alone — `TestCheckLoopLandingSatellite/rows_under_the_agents_project` fails, refused with `PLAN —`, `FIX —` and `REVIEW —` though the agents project holds a clean loop.
**Baseline:** before=1 after=2
<!-- measured: grep -cE '^func Test' stats/internal/guard/check_loop_landing_test.go (before at fe238ed0; after with this task's diff applied) -->
**Commit:** fix(guard): land a loop's satellite branch on the agents project's rows
**After:** none

**Decision:** fallback-to-agents-project
**Decision:** own-branch-no-fallback
**Decision:** fallback-read-failure-keeps-refusal
