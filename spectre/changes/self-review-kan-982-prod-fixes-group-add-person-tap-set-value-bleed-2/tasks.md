# self-review-kan-982-prod-fixes-group-add-person-tap-set-value-bleed-2

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
- Each task's change is the unified diff its step gives, applied from the repository root with `git apply` — both diffs were applied in plan order to a scratch clone at `8650ac1f` and checked there. A diff that does not apply is reported, never hand-merged by guesswork.
- Every prose line a diff adds is copied verbatim: this change's `verbatim-moves.txt` acknowledges each new or reworded sentence by its exact text, and any other wording fails `scripts/check-verbatim-moves.sh`.
- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w ./internal/guard` before `gofmt -l .`. `go build ./...` needs the built SPA (`internal/web/dist`, `make web-build`); the commands below vet and test only the package each task touches.

- [ ] 1. Run 1's gate lists every commit outside every task, and the sanctioned routes mark theirs with `Flow-Key:`
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `ls scripts/check-untasked-commits.sh` and confirm `No such file or directory`; then apply the `scripts/test-integrate-change.sh` hunks of the diff in step 2 alone (`git apply --include='scripts/test-integrate-change.sh' <the diff>`) and run `bash scripts/test-integrate-change.sh 2>&1 | grep untasked`: `FAIL - untasked: prepare stops at STOP: unfinished-work` and `FAIL - untasked: the stray commit is listed` — the defect: a trailerless code commit passes run 1's gate unlisted.
  - [ ] **Step 2: Implement.** Apply the rest of this diff (`git apply --exclude='scripts/test-integrate-change.sh' <the diff>`); it creates `skills/flow/scripts/check-untasked-commits.sh` as a symlink to `../../../scripts/check-untasked-commits.sh` and `scripts/check-untasked-commits.sh` executable:

````diff verified:applied with git apply in plan order to a scratch clone at 8650ac1f; gofmt -l and go vet ./internal/guard clean; go test ./internal/guard/ and scripts/run-guard-tests.sh pass with both tasks applied; TestCheckUntaskedCommits fails with the Task-Id or the planning-only exemption removed; the untasked case of test-integrate-change.sh fails with the integrate-change.sh hunk reverted
diff --git a/scripts/check-untasked-commits.sh b/scripts/check-untasked-commits.sh
new file mode 100755
index 00000000..f939bbbc
--- /dev/null
+++ b/scripts/check-untasked-commits.sh
@@ -0,0 +1,36 @@
+#!/usr/bin/env bash
+# check-untasked-commits.sh — list the commits of a change branch that sit
+# outside every task, for /flow's run-1 unfinished-work gate.
+#
+# Usage: check-untasked-commits.sh <worktree> <base>
+#
+# WHY THIS GUARD EXISTS. A commit the run makes outside every task — a stray
+# implementer edit, a parent's unrecorded fix — was caught only by a review
+# panel reading the log (kan-982 panel F3: "Two non-chore commits sit outside
+# every task"). Run 1's reshape then squashes it into the implementation
+# commit, and the evidence is gone.
+#
+# Every non-merge commit of <base>..HEAD is listed unless it carries a
+# non-empty `Task-Id:` trailer (a task's own commit), a non-empty `Flow-Key:`
+# trailer (a commit a run makes outside every task on a sanctioned route —
+# **Commits outside every task**, skills/flow-contracts/git-boundaries.md),
+# has a `chore` subject (the planning, link and archive commits), or touches
+# only paths under the worktree's `<spec-root>/changes/` (a plan amendment).
+#
+# Prints one verdict line to stdout:
+#   UNTASKED-OK: <n> commit(s) checked                      nothing listed — exit 0
+#   UNTASKED: <k> commit(s) carry no Task-Id: or Flow-Key: trailer
+#     <sha12> <subject>                                      one line per commit — exit 1
+# Exit 2 with no verdict line when it cannot answer: a usage error, a
+# worktree that is not a directory, or a range git cannot read.
+#
+# Go, in flow-guard (stats/internal/guard/untaskedcommits.go). flow-guard is
+# built from this checkout, never taken from PATH: scripts/lib/flow-guard.sh
+# derives it, and exits 2 (this guard's cannot-answer code) with the cause
+# when it cannot.
+set -euo pipefail
+[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
+  echo "check-untasked-commits: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
+  exit 2
+}
+flow_guard_exec check-untasked-commits 2 "check-untasked-commits:" "$@"
diff --git a/scripts/integrate-change.sh b/scripts/integrate-change.sh
index d6757a6e..fd683411 100755
--- a/scripts/integrate-change.sh
+++ b/scripts/integrate-change.sh
@@ -28,7 +28,8 @@
 #          worktree RUN2 -> NEXT is `cleanup`. Otherwise run 1: the
 #          uncommitted-archive undo, the archived decision, the unfinished-work
 #          gate (check-unfinished-work, check-visual-verify-dispatched against
-#          `git merge-base HEAD origin/$BASE`), check-base-moved and
+#          `git merge-base HEAD origin/$BASE`, check-untasked-commits from that
+#          base or an archived re-run's), check-base-moved and
 #          sync-onto-base on MOVED (running each GUARD-TEST it names), and the
 #          default landing route (project-get.sh). Prints per worktree its
 #          reshape base and the diffstat the commit subjects are written from,
@@ -90,7 +91,7 @@ say() { printf '%s\n' "$*"; }
 
 missing=""
 for g in check-base-moved check-cleanup-complete check-finish-preflight check-foreign-staged \
-  check-main-checkout-drift check-unfinished-work check-visual-verify-dispatched commit-archive \
+  check-main-checkout-drift check-unfinished-work check-untasked-commits check-visual-verify-dispatched commit-archive \
   commit-split land-self-review-report migrate-worktrees project-get refresh-main-checkout \
   remove-change-worktrees reshape-branch resolve-base-branch sync-onto-base; do
   [ -x "$SCRIPT_DIR/$g.sh" ] || missing="$missing $g.sh"
@@ -385,6 +386,14 @@ cmd_prepare() {
         1:VISUAL-VERIFY-MISSING:*) ask=1 ;;
         *) stop no-verdict "check-visual-verify-dispatched reached no verdict for ${WT[i]} (exit $RC) — report its output and ask the operator; --accept-outstanding does not pass this" ;;
       esac
+      ub="$(git -C "${WT[i]}" merge-base HEAD "origin/${BASES[i]}")"
+      [ "$ARCHIVED" -eq 1 ] && ub="$(arch_base "${WT[i]}" "$ub")"
+      run "$SCRIPT_DIR/check-untasked-commits.sh" "${WT[i]}" "$ub"
+      case "$RC:$(verdict 'UNTASKED-OK|UNTASKED')" in
+        0:UNTASKED-OK:*) ;;
+        1:UNTASKED:*) ask=1 ;;
+        *) stop no-verdict "check-untasked-commits reached no verdict for ${WT[i]} (exit $RC) — report its output and ask the operator; --accept-outstanding does not pass this" ;;
+      esac
     done
     [ "$ask" -eq 1 ] && [ -z "$ACCEPT_OUTSTANDING" ] &&
       stop unfinished-work "load skills/flow/unfinished-work-gate.md and offer its three courses; Continue or File-or-join re-runs prepare with --accept-outstanding"
diff --git a/scripts/test-integrate-change.sh b/scripts/test-integrate-change.sh
index fd38744b..69b1a4be 100755
--- a/scripts/test-integrate-change.sh
+++ b/scripts/test-integrate-change.sh
@@ -26,6 +26,9 @@
 #   no_report       land with no committed self-review report stops at
 #                   STOP: self-review-report, lands nothing and pushes no
 #                   commit made after `commit`.
+#   untasked        a trailerless code commit in b stops prepare at
+#                   STOP: unfinished-work, its subject listed; the
+#                   --accept-outstanding re-run passes the gate.
 #   no_verdict      a store outage leaves check-unfinished-work without a
 #                   verdict: STOP: no-verdict, which --accept-outstanding
 #                   does not pass.
@@ -104,7 +107,7 @@ fixture() {
     g -C "$WORK/$r" worktree add -q -b spectre/demo "$WORK/$r-worktrees/demo" main
     printf 'work in %s\n' "$r" >"$WORK/$r-worktrees/demo/work-$r.txt"
     g -C "$WORK/$r-worktrees/demo" add -A
-    g -C "$WORK/$r-worktrees/demo" commit -qm "feat($r): task 1"
+    g -C "$WORK/$r-worktrees/demo" commit -qm "feat($r): task 1" -m "Task-Id: 1"
   done
   local c="$WORK/a-worktrees/demo/spectre/changes/demo"
   mkdir -p "$c"
@@ -253,7 +256,7 @@ fixture
 to_land "no_report: " no-report
 pushed="$(git -C "$WORK/a-worktrees/demo" rev-parse HEAD)"
 printf 'later\n' >"$WORK/a-worktrees/demo/later.txt"
-g -C "$WORK/a-worktrees/demo" add later.txt && g -C "$WORK/a-worktrees/demo" commit -qm "feat(a): later"
+g -C "$WORK/a-worktrees/demo" add later.txt && g -C "$WORK/a-worktrees/demo" commit -qm "feat(a): later" -m "Task-Id: 2"
 sut land "$WORK/a" demo
 check "no_report: land stops with exit 1" '[ "$RC" -eq 1 ]'
 check "no_report: the stop is STOP: self-review-report" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: self-review-report"'
@@ -262,6 +265,16 @@ check "no_report: nothing was landed" \
 check "no_report: a's remote branch still sits at commit's HEAD" \
   '[ "$(git --git-dir="$WORK/a-origin.git" rev-parse spectre/demo)" = "$pushed" ]'
 
+# --- untasked ---
+fixture
+printf 'stray\n' >"$WORK/b-worktrees/demo/stray.txt"
+g -C "$WORK/b-worktrees/demo" add stray.txt && g -C "$WORK/b-worktrees/demo" commit -qm "test(b): a stray edit"
+sut prepare "$WORK/a" demo
+check "untasked: prepare stops at STOP: unfinished-work" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: unfinished-work"'
+check "untasked: the stray commit is listed" 'printf "%s" "$OUT" | grep -q "^  [0-9a-f]\{12\} test(b): a stray edit$"'
+sut prepare "$WORK/a" demo --accept-outstanding
+check "untasked: --accept-outstanding passes the gate" '[ "$RC" -eq 0 ]'
+
 # --- no_verdict ---
 fixture
 : >"$WORK/fail-findings"
@@ -289,7 +302,7 @@ cp -R "$SCRIPT_DIR" "$WORK/agents/scripts"
 ln -s "$(cd "$SCRIPT_DIR/../stats" && pwd -P)" "$WORK/agents/stats"
 printf '#!/usr/bin/env bash\necho gt-test-own-output\nexit 1\n' >"$WORK/agents/scripts/test-gt.sh"
 printf 'same\n' >"$WORK/b-worktrees/demo/gt.txt"
-g -C "$WORK/b-worktrees/demo" add gt.txt && g -C "$WORK/b-worktrees/demo" commit -qm "feat(b): gt"
+g -C "$WORK/b-worktrees/demo" add gt.txt && g -C "$WORK/b-worktrees/demo" commit -qm "feat(b): gt" -m "Task-Id: 2"
 printf 'same\n' >"$WORK/b/gt.txt"
 g -C "$WORK/b" add gt.txt && g -C "$WORK/b" commit -qm "base: gt" && g -C "$WORK/b" push -q origin main
 SUT_SAVED="$SUT"; SUT="$WORK/agents/scripts/integrate-change.sh"
diff --git a/skills/flow-contracts/finish-contract-run1.md b/skills/flow-contracts/finish-contract-run1.md
index 002261b5..66c55fe0 100644
--- a/skills/flow-contracts/finish-contract-run1.md
+++ b/skills/flow-contracts/finish-contract-run1.md
@@ -164,8 +164,15 @@ breakdown in; a `VISUAL-VERIFY-MISSING` line is treated exactly as `OUTSTANDING`
 breakdown and the same three courses (**The three courses**, `skills/flow/unfinished-work-gate.md`), not a second prompt. Exit 2 (cannot answer) is stop-and-ask,
 the same as a missing `check-unfinished-work.sh` verdict line above.
 
-**Load `skills/flow/unfinished-work-gate.md` only when** a worktree reported `OUTSTANDING` or
-`VISUAL-VERIFY-MISSING` — the three courses, the filing and the relay are there.
+**Then run `check-untasked-commits.sh <worktree> <base>`**, once per worktree in the same
+resolved set, `<base>` the same fork point — or, where the change is archived, the archived
+re-run's base that fork point yields. An `UNTASKED-OK` line joins `CLEAR`; an `UNTASKED` line and
+the commits it lists are treated exactly as `OUTSTANDING`, and exit 2 is stop-and-ask, the same as
+a missing `check-unfinished-work.sh` verdict line. The guard's header
+(`<agents repo>/scripts/check-untasked-commits.sh`) is canonical for what it exempts.
+
+**Load `skills/flow/unfinished-work-gate.md` only when** a worktree reported `OUTSTANDING`,
+`VISUAL-VERIFY-MISSING` or `UNTASKED` — the three courses, the filing and the relay are there.
 
 **Check whether the base branch has moved — after the unfinished-work gate, before the landing
 question.** `check-base-moved.sh <worktree> <base-ref> <recorded-merge-base|->` prints one verdict
diff --git a/skills/flow-contracts/git-boundaries.md b/skills/flow-contracts/git-boundaries.md
index cab17bd4..02efbe38 100644
--- a/skills/flow-contracts/git-boundaries.md
+++ b/skills/flow-contracts/git-boundaries.md
@@ -121,6 +121,18 @@ pathspec-scoped commit would drop.
 
 A `/flow-fast` branch's commit series is its whole record, and the route reads it back before anything lands — `check-fast-route-record.sh <worktree> <base>` in section 5's verify stage (`skills/flow-fast/SKILL.md`) asserts every commit since the merge base against the contract canonical in the guard's own header, a hit fixed and re-run under the route's commit rule.
 
+## Commits outside every task
+
+**Commits outside every task carry a `Flow-Key:` trailer.** A commit a run makes on a change
+branch outside every task — the parent's full-suite fix, a panel fix or a gated reviewer's fix
+committed on top, `flow.visual-verify`'s capture commit, the brainstorm's handoff-asset commit —
+carries `-m "Flow-Key: <key>"`, `<key>` the dispatch key its round is recorded under
+(`full-suite-fix-<n>`, `panel-fix-<round>`, `task-<n+n>-implementer-fix-<k>`) or, where the stage
+records no such row, the stage key (`flow.visual-verify`, `flow.brainstorm`). A task's own commit
+carries `Task-Id:` instead, and a planning commit neither. Run 1's unfinished-work gate lists
+every other non-`chore` commit with `check-untasked-commits.sh`
+(`skills/flow-contracts/finish-contract-run1.md`).
+
 ## Revert and back-out commits
 
 **A back-out's commit body carries the rejection reason.** When a landed-then-reverted prototype
diff --git a/skills/flow/brainstorm-planner.md b/skills/flow/brainstorm-planner.md
index a19eb802..8360c5b7 100644
--- a/skills/flow/brainstorm-planner.md
+++ b/skills/flow/brainstorm-planner.md
@@ -34,7 +34,9 @@ the design.
   are committed into the repository or their location is recorded in
   `<project>/.flow/project.md` (the `mockups` row of `## visual verification` is where a declared
   mockups directory lives — **Project configuration — visual verification**,
-  `skills/flow-contracts/project-configuration-visual.md`).
+  `skills/flow-contracts/project-configuration-visual.md`). A commit of the assets carries
+  `-m "Flow-Key: flow.brainstorm"` (**Commits outside every task**,
+  `skills/flow-contracts/git-boundaries.md`).
 - The design presentation does **not** end a section, or the whole design, with a "does this look
   right?" question — present the section(s) and proceed directly, section to section and then into
   artifact creation, unless the operator raises an objection during or after that presentation. This
diff --git a/skills/flow/review-panel-fix-round.md b/skills/flow/review-panel-fix-round.md
index 5caa14d5..388ab662 100644
--- a/skills/flow/review-panel-fix-round.md
+++ b/skills/flow/review-panel-fix-round.md
@@ -32,7 +32,7 @@ This is the normal case: the branch is pushed with every commit (**Branch backup
 (`git add -- <the changed paths>` — a pathspec commit reads tracked paths only) and is a plain
 `git commit -m ... -- <the changed paths>` at the tip — the pathspec-scoped default (**Branch backup**, `skills/flow-contracts/git-boundaries.md`),
 so it carries only the paths the finding named, whatever else the index holds — pushed
-plain like any other commit, and every downstream commit keeps its sha. **A branch a
+plain like any other commit, and every downstream commit keeps its sha. The commit carries the round's `Flow-Key:` trailer — `git commit -m <subject> -m "Flow-Key: <the round's dispatch key>" -- <the changed paths>` (**Commits outside every task**, `skills/flow-contracts/git-boundaries.md`). **A branch a
 round-boundary auto-rebase has rewritten pushes with `--force-with-lease`, never a bare
 `git push` and never `--force`** — the rebase replaced the shas the remote still holds, so the
 plain push above is then rejected.
diff --git a/skills/flow/scripts/check-untasked-commits.sh b/skills/flow/scripts/check-untasked-commits.sh
new file mode 120000
index 00000000..af69c741
--- /dev/null
+++ b/skills/flow/scripts/check-untasked-commits.sh
@@ -0,0 +1 @@
+../../../scripts/check-untasked-commits.sh
\ No newline at end of file
diff --git a/skills/flow/visual-verify.md b/skills/flow/visual-verify.md
index 7b1e1842..04b4371b 100644
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -218,7 +218,7 @@ states.
 12. **Commit the spec and its PNGs, and stop there.** A declared `regression checkout` receives
     them; with none declared, commit to the change's own branch instead. **The commit is
     pathspec-scoped** — `git add --` the spec, its PNGs and `<changeRoot>/visual-verification/`
-    first, then `git commit -m "<subject>" --` those same paths — the add first because the
+    first, then `git commit -m "<subject>" -m "Flow-Key: flow.visual-verify" --` those same paths — the add first because the
     stage's outputs are new untracked files, which a pathspec commit cannot pick up — carrying
     only what this stage wrote, never a bare `git commit`: the index of a main checkout
     may carry a pre-staged foreign tree (**Branch backup**, `skills/flow-contracts/git-boundaries.md`). **Resolve the
diff --git a/stats/internal/guard/check_untasked_commits_test.go b/stats/internal/guard/check_untasked_commits_test.go
new file mode 100644
index 00000000..04020fba
--- /dev/null
+++ b/stats/internal/guard/check_untasked_commits_test.go
@@ -0,0 +1,93 @@
+package guard
+
+import (
+	"bytes"
+	"os"
+	"path/filepath"
+	"strings"
+	"testing"
+)
+
+// TestCheckUntaskedCommits builds one repository per case: a base commit,
+// then the case's commits above it, and runs the guard over base..HEAD.
+func TestCheckUntaskedCommits(t *testing.T) {
+	t.Parallel()
+	type commit struct{ path, subject, trailer string }
+	cases := []struct {
+		name    string
+		commits []commit
+		code    int
+		want    []string
+		absent  []string
+	}{
+		{"task commit passes", []commit{{"src/a.kt", "feat(app): a", "Task-Id: 1"}}, 0, []string{"UNTASKED-OK: 1 commit(s) checked"}, nil},
+		{"flow-key commit passes", []commit{{"src/a.kt", "fix(app): review Minors", "Flow-Key: panel-fix-1"}}, 0, []string{"UNTASKED-OK"}, nil},
+		{"chore commit passes", []commit{{"src/a.kt", "chore(spectre): plan demo", ""}}, 0, []string{"UNTASKED-OK"}, nil},
+		{"planning-only commit passes", []commit{{"spectre/changes/demo/tasks.md", "docs(spectre): widen task 3", ""}}, 0, []string{"UNTASKED-OK"}, nil},
+		{"trailerless code commit is listed", []commit{
+			{"src/a.kt", "feat(app): a", "Task-Id: 1"},
+			{"src/b.kt", "test(update): drop checks", ""},
+		}, 1, []string{"UNTASKED: 1 commit(s) carry no Task-Id: or Flow-Key: trailer", "test(update): drop checks"}, []string{"feat(app): a"}},
+		{"planning plus code is listed", []commit{{"spectre/changes/demo/tasks.md\nsrc/c.kt", "docs(spectre): note", ""}}, 1, []string{"docs(spectre): note"}, nil},
+	}
+	for _, tc := range cases {
+		t.Run(tc.name, func(t *testing.T) {
+			t.Parallel()
+			dir := t.TempDir()
+			g := &fxGit{}
+			g.git(dir, "init", "-q", "-b", "main")
+			writeFile(t, filepath.Join(dir, "README"), "base\n")
+			g.git(dir, "add", "-A")
+			g.git(dir, "commit", "-qm", "base")
+			base := g.git(dir, "rev-parse", "HEAD")
+			for i, c := range tc.commits {
+				for _, p := range strings.Split(c.path, "\n") {
+					writeFile(t, filepath.Join(dir, p), c.subject+string(rune('a'+i))+"\n")
+				}
+				g.git(dir, "add", "-A")
+				msg := []string{"commit", "-qm", c.subject}
+				if c.trailer != "" {
+					msg = append(msg, "-m", c.trailer)
+				}
+				g.git(dir, msg...)
+			}
+			if g.err != nil {
+				t.Fatal(g.err)
+			}
+			var out, errb bytes.Buffer
+			code := checkUntaskedCommits([]string{dir, base}, Env{Getenv: os.Getenv, Dir: dir}, &out, &errb)
+			if code != tc.code {
+				t.Fatalf("exit %d, want %d\nstdout:\n%s\nstderr:\n%s", code, tc.code, out.String(), errb.String())
+			}
+			for _, w := range tc.want {
+				if !strings.Contains(out.String(), w) {
+					t.Errorf("stdout lacks %q:\n%s", w, out.String())
+				}
+			}
+			for _, a := range tc.absent {
+				if strings.Contains(out.String(), a) {
+					t.Errorf("stdout carries %q:\n%s", a, out.String())
+				}
+			}
+		})
+	}
+}
+
+// TestCheckUntaskedCommitsCannotAnswer pins exit 2 on a usage error, a
+// missing worktree and a base git cannot resolve.
+func TestCheckUntaskedCommitsCannotAnswer(t *testing.T) {
+	t.Parallel()
+	dir := t.TempDir()
+	g := &fxGit{}
+	g.git(dir, "init", "-q", "-b", "main")
+	g.git(dir, "commit", "-q", "--allow-empty", "-m", "base")
+	if g.err != nil {
+		t.Fatal(g.err)
+	}
+	for _, args := range [][]string{{dir}, {filepath.Join(dir, "absent"), "HEAD"}, {dir, "no-such-ref"}} {
+		var out, errb bytes.Buffer
+		if code := checkUntaskedCommits(args, Env{Getenv: os.Getenv, Dir: dir}, &out, &errb); code != 2 {
+			t.Errorf("args %q: exit %d, want 2 (stdout %q)", args, code, out.String())
+		}
+	}
+}
diff --git a/stats/internal/guard/untaskedcommits.go b/stats/internal/guard/untaskedcommits.go
new file mode 100644
index 00000000..365654c6
--- /dev/null
+++ b/stats/internal/guard/untaskedcommits.go
@@ -0,0 +1,85 @@
+package guard
+
+import (
+	"fmt"
+	"io"
+	"regexp"
+	"strings"
+)
+
+// check-untasked-commits lists the commits of <base>..HEAD that sit outside
+// every task, for /flow's run-1 unfinished-work gate. Its shim's header
+// (scripts/check-untasked-commits.sh) is canonical for the contract.
+func init() { Registry["check-untasked-commits"] = checkUntaskedCommits }
+
+const ucPrefix = "check-untasked-commits:"
+
+// ucChoreRE is a Conventional Commits `chore` subject: the pipeline's own
+// planning, link and archive commits.
+var ucChoreRE = regexp.MustCompile(`^chore(\([^)]*\))?!?:`)
+
+func checkUntaskedCommits(args []string, env Env, stdout, stderr io.Writer) int {
+	if len(args) != 2 || args[0] == "" || args[1] == "" {
+		fmt.Fprintf(stderr, "%s usage: check-untasked-commits.sh <worktree> <base>\n", ucPrefix)
+		return 2
+	}
+	worktree, base := args[0], args[1]
+	if !isDir(worktree) {
+		fmt.Fprintf(stderr, "%s worktree not found: %s\n", ucPrefix, worktree)
+		return 2
+	}
+	git := envGit(env)
+	out, ok := capture(git("-C", worktree, "log", "--no-merges",
+		"--format=%H%x00%s%x00%(trailers:key=Task-Id,valueonly,separator=%x2C)%x00%(trailers:key=Flow-Key,valueonly,separator=%x2C)%x1e",
+		base+"..HEAD"))
+	if !ok {
+		fmt.Fprintf(stderr, "%s git cannot read %s..HEAD in %s\n", ucPrefix, base, worktree)
+		return 2
+	}
+	planning := specRootLeaf(worktree, io.Discard) + "/changes/"
+	var listed []string
+	checked := 0
+	for _, rec := range strings.Split(out, "\x1e") {
+		f := strings.Split(strings.Trim(rec, "\n"), "\x00")
+		if len(f) != 4 {
+			continue
+		}
+		checked++
+		sha, subject := f[0], f[1]
+		if strings.TrimSpace(f[2]) != "" || strings.TrimSpace(f[3]) != "" || ucChoreRE.MatchString(subject) {
+			continue
+		}
+		paths, ok := capture(git("-C", worktree, "diff-tree", "--no-commit-id", "--no-renames", "--name-only", "-r", "--root", sha))
+		if !ok {
+			fmt.Fprintf(stderr, "%s git cannot list the paths of %s\n", ucPrefix, sha)
+			return 2
+		}
+		if ucPlanningOnly(paths, planning) {
+			continue
+		}
+		listed = append(listed, fmt.Sprintf("  %s %s", sha[:12], sanitizeDisplay(subject)))
+	}
+	if len(listed) == 0 {
+		fmt.Fprintf(stdout, "UNTASKED-OK: %d commit(s) checked\n", checked)
+		return 0
+	}
+	fmt.Fprintf(stdout, "UNTASKED: %d commit(s) carry no Task-Id: or Flow-Key: trailer\n", len(listed))
+	for _, l := range listed {
+		fmt.Fprintln(stdout, l)
+	}
+	return 1
+}
+
+// ucPlanningOnly is whether every path a commit touched lies under the
+// planning tree; a commit touching nothing is not.
+func ucPlanningOnly(paths, planning string) bool {
+	if paths == "" {
+		return false
+	}
+	for _, p := range strings.Split(paths, "\n") {
+		if !strings.HasPrefix(p, planning) {
+			return false
+		}
+	}
+	return true
+}
````

  - [ ] **Step 3: Verify.** `cd stats && gofmt -l ./internal/guard && go vet ./internal/guard && go test ./internal/guard/ -run 'TestCheckUntaskedCommits|TestCheckGuardSymlinks' -count=1` passes; `bash scripts/test-integrate-change.sh` ends `0 failed`; `scripts/test-git-config-pins.sh`, `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `python3 scripts/check-markdown-integrity.py`, `scripts/check-guard-symlinks.sh`, `scripts/check-installed-citations.sh` and `scripts/check-dispatch-paragraphs.sh` exit 0.
  - [ ] **Step 4: Commit** the paths in `**Files:**` with the subject below.

**Build:** green
**Files:** `stats/internal/guard/untaskedcommits.go`, `stats/internal/guard/check_untasked_commits_test.go`, `scripts/check-untasked-commits.sh`, `skills/flow/scripts/check-untasked-commits.sh`, `scripts/integrate-change.sh`, `scripts/test-integrate-change.sh`, `skills/flow-contracts/git-boundaries.md`, `skills/flow-contracts/finish-contract-run1.md`, `skills/flow/review-panel-fix-round.md`, `skills/flow/visual-verify.md`, `skills/flow/brainstorm-planner.md`
**Tests:** `TestCheckUntaskedCommits`, `TestCheckUntaskedCommitsCannotAnswer`
**Regression:** reverting the commit lets a trailerless, non-chore code commit pass run 1's unfinished-work gate unlisted (`TestCheckUntaskedCommits` cannot compile without the guard; the `untasked` case of `scripts/test-integrate-change.sh` fails), and the sanctioned routes stop writing `Flow-Key:`.
**Baseline:** before=175 after=177
<!-- measured: cat stats/internal/guard/*_test.go | grep -cE '^func Test[A-Za-z0-9_]*\(' @ 8650ac1f (before; after with this task's diff applied in the scratch clone) -->
**Commit:** feat(guard): run 1's gate lists commits outside every task
**After:** none

**Decision:** flow-key-trailer
**Decision:** untasked-derived-exemptions
**Decision:** untasked-at-integrate-gate
**Decision:** integrate-call-site-in-bash

- [ ] 2. `check-task-commit-fields.sh` refuses a scope naming only another repository's module
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Apply the `stats/internal/guard/check_task_commit_scope_sibling_test.go` hunk of the diff in step 2 alone (`git apply --include='stats/internal/guard/check_task_commit_scope_sibling_test.go' <the diff>`) and run `cd stats && go test ./internal/guard/ -run TestCheckScopeSibling -count=1`: it FAILS to build with `undefined: tcfCheckScopeSibling` — the defect: no check reads a declared scope against the repositories of the project, so kan-982's `fix(group)` on a gymie-playwright commit touching only `fidelity/builder.spec.ts` closed clean.
  - [ ] **Step 2: Implement.** Apply the rest of this diff:

````diff verified:applied with git apply after Task 1 to a scratch clone at 8650ac1f; gofmt -l and go vet ./internal/guard clean; go test ./internal/guard/ passes; TestCheckScopeSibling fails with the sibling refusal disabled; tcfCheckScopeSibling refuses fix(group) at gymie-playwright 7479ba1 against gymie's project.md and passes fidelity, visual and e2e
diff --git a/rules/commit-scope-is-the-module.mdc b/rules/commit-scope-is-the-module.mdc
index a36b4ae4..5d6b2148 100644
--- a/rules/commit-scope-is-the-module.mdc
+++ b/rules/commit-scope-is-the-module.mdc
@@ -19,16 +19,22 @@ never a list.
 codebase moved without opening the commit. A scope naming the change or a ticket repeats information
 the branch and the tracker already carry, and tells the reader nothing about where the change landed.
 
-## The three prohibited shapes
+## The four prohibited shapes
 
 - **The change name** — `feat(kan-239-run-2-asserts-base-branch-and-archives-via-pr): …` names the
   change, not a module.
 - **The change's bare Jira key** — `feat(kan-239): …` is the same defect with the descriptive suffix
   trimmed off.
 - **A task id** — `fix(3.2): …` names a line in a plan, not a part of the codebase.
+- **A module of another repository** — `fix(group): …` on a commit touching only gymie-playwright's
+  fidelity/builder.spec.ts, `group` being a gymie module, names where the change did not land.
+  The guard refuses a scope that is a path token of another `## apps` repository of the project
+  and is none of: a token of a path the commit touched, a word of its own repository's name, a
+  scope an earlier commit of its repository used, or a directory name of its repository — every
+  `-` part of the scope one, for a compound area such as `workout-ui`.
 
 Anything else is allowed. There is no vocabulary of legal module names to maintain; the rule refuses
-three known-wrong shapes and accepts everything else.
+four known-wrong shapes and accepts everything else.
 
 ## The guard
 
@@ -46,5 +52,8 @@ reproduces it exactly. Outside the pipeline, this rule is the only enforcement:
   passes; the check is equality, never a substring match.
 - `feat(kan-202-commit-split-and-module-scopes): …` — the change name, refused.
 - `fix(3.2): …` — a dotted task id, refused.
+- `fix(group): …` touching only gymie-playwright's fidelity/builder.spec.ts — a module of gymie
+  and nothing in gymie-playwright, refused.
+- `test(visual): …` on a snapshot capture — no repository of the project carries `visual`, passes.
 
 Applies globally, to every commit, not only inside the flow pipeline.
diff --git a/stats/internal/guard/check_task_commit_scope_sibling_test.go b/stats/internal/guard/check_task_commit_scope_sibling_test.go
new file mode 100644
index 00000000..066669cc
--- /dev/null
+++ b/stats/internal/guard/check_task_commit_scope_sibling_test.go
@@ -0,0 +1,71 @@
+package guard
+
+import (
+	"os"
+	"path/filepath"
+	"strings"
+	"testing"
+)
+
+// TestCheckScopeSibling pins tcfCheckScopeSibling over two repositories of
+// one project: gymie-playwright, whose commit touches only
+// fidelity/builder.spec.ts, and gymie, which carries a `group` module.
+func TestCheckScopeSibling(t *testing.T) {
+	t.Parallel()
+	root := t.TempDir()
+	own, sib := filepath.Join(root, "gymie-playwright"), filepath.Join(root, "gymie")
+	g := &fxGit{}
+	for _, r := range []string{own, sib} {
+		g.git("", "init", "-q", "-b", "main", r)
+	}
+	for _, p := range []string{"src/group/Group.kt", "e2e/x.ts", "helpers/h.ts", "workout-ui/w.kt", "playwright/p.ts", "build/b.kts"} {
+		writeFile(t, filepath.Join(sib, p), "x\n")
+	}
+	g.git(sib, "add", "-A")
+	g.git(sib, "commit", "-qm", "base")
+	writeFile(t, filepath.Join(own, ".flow", "project.md"),
+		"# p\n\n## apps\n\n| App | Repo root | Kind |\n|-----|-----------|------|\n| Suite | `"+own+"` | Node |\n| Backend | `"+sib+"` | Kotlin |\n")
+	for _, p := range []string{"docs/self-review/kan-30-group-workouts.md", "helpers/h.ts", "workout/a.ts", "ui/b.ts"} {
+		writeFile(t, filepath.Join(own, p), "x\n")
+	}
+	g.git(own, "add", "-A")
+	g.git(own, "commit", "-qm", "test(e2e): base")
+	writeFile(t, filepath.Join(own, "fidelity", "builder.spec.ts"), "y\n")
+	g.git(own, "add", "-A")
+	g.git(own, "commit", "-qm", "work")
+	if g.err != nil {
+		t.Fatal(g.err)
+	}
+	diffs := []tcfRepoDiff{{worktree: own, commit: "HEAD", parent: "HEAD^", changed: []string{"fidelity/builder.spec.ts"}}}
+	cases := []struct {
+		name, subject, projectDir string
+		refused                   bool
+	}{
+		{"kan-982 F16: a backend module on a playwright-only commit", "fix(group): the sheet closes", own, true},
+		{"a touched path passes", "fix(fidelity): the sheet closes", own, false},
+		{"a scope no repository carries passes", "test(visual): rebaseline", own, false},
+		{"an earlier scope of this repository passes", "fix(e2e): settle", own, false},
+		{"a directory of this repository passes", "fix(helpers): tidy", own, false},
+		{"every part a directory of this repository passes", "fix(workout-ui): tidy", own, false},
+		{"a word of this repository's name passes", "fix(playwright): config", own, false},
+		{"a sibling directory this repository lacks is refused", "chore(build): bump", own, true},
+		{"no apps table refuses nothing", "fix(group): the sheet closes", t.TempDir(), false},
+		{"no scope passes", "fix: the sheet closes", own, false},
+	}
+	for _, tc := range cases {
+		t.Run(tc.name, func(t *testing.T) {
+			t.Parallel()
+			subject := tc.subject
+			v, err := tcfCheckScopeSibling(Env{Getenv: os.Getenv, Dir: own}, tcfTask{id: "21", commit: &subject}, diffs, tc.projectDir)
+			if err != nil {
+				t.Fatal(err)
+			}
+			if got := len(v) > 0; got != tc.refused {
+				t.Fatalf("refused = %v, want %v: %q", got, tc.refused, v)
+			}
+			if tc.refused && !strings.Contains(v[0], "names a module of gymie and nothing this commit touched") {
+				t.Errorf("message %q does not name the sibling", v[0])
+			}
+		})
+	}
+}
diff --git a/stats/internal/guard/taskcommitfields.go b/stats/internal/guard/taskcommitfields.go
index 8a0123fb..ad72e2a7 100644
--- a/stats/internal/guard/taskcommitfields.go
+++ b/stats/internal/guard/taskcommitfields.go
@@ -178,7 +178,11 @@ func checkTaskCommitFields(args []string, env Env, stdout, stderr io.Writer) int
 	// forwarded. Shared by all three resolution paths (named change, satellite
 	// link, glob-path) so the dispatch lives in exactly one place.
 	judge := func(tasksMD string) int {
-		violations, err := tcfCheckTaskCommit(env, tasksMD, taskID, repos, parent)
+		projectDir := canonical
+		if projectDir == "" {
+			projectDir = worktree
+		}
+		violations, err := tcfCheckTaskCommit(env, tasksMD, taskID, repos, parent, projectDir)
 		var crash tcfCrash
 		switch {
 		case errors.As(err, &crash):
@@ -1350,7 +1354,7 @@ type tcfRepoDiff struct {
 }
 
 // tcfCheckTaskCommit is check_task_commit.
-func tcfCheckTaskCommit(env Env, tasksMD, taskID string, repos []tcfRepoCommit, parent string) ([]string, error) {
+func tcfCheckTaskCommit(env Env, tasksMD, taskID string, repos []tcfRepoCommit, parent, projectDir string) ([]string, error) {
 	raw, err := os.ReadFile(tasksMD)
 	if err != nil {
 		return nil, tcfOSError(tasksMD, err)
@@ -1474,9 +1478,147 @@ func tcfCheckTaskCommit(env Env, tasksMD, taskID string, repos []tcfRepoCommit,
 		v = append(v, tcfCheckCommitSubject(folded, tcfStrip(d.subject))...)
 	}
 	v = append(v, tcfCheckCommitScope(task, changeName)...)
+	sibling, err := tcfCheckScopeSibling(env, task, diffs, projectDir)
+	if err != nil {
+		return nil, err
+	}
+	v = append(v, sibling...)
 	return v, nil
 }
 
+// tcfCheckScopeSibling refuses a declared scope that names a module of
+// another repository of the project and nothing in this one — kan-982's
+// `fix(group)` on a commit touching only gymie-playwright's
+// fidelity/builder.spec.ts, `group` being a backend and frontend module.
+// The scope passes when it is a token of a path the commit touched, a word of
+// its repository's name, a scope an earlier commit of the repository used, a
+// directory name of the repository's tree at the commit (or every `-` part of
+// it is one), or no token of any other `## apps` repository's tree. A sibling
+// git cannot read refuses nothing. Measured at plan time over the gymie
+// repositories' scoped history: 10 of 1494 commits refused.
+func tcfCheckScopeSibling(env Env, task tcfTask, diffs []tcfRepoDiff, projectDir string) ([]string, error) {
+	if task.commit == nil {
+		return nil, nil
+	}
+	m := tcfSubjectScopeRE.FindStringSubmatch(*task.commit)
+	if m == nil || m[1] == "" {
+		return nil, nil
+	}
+	scope := strings.ToLower(m[1])
+	git := envGit(env)
+	own := map[string]bool{}
+	for _, d := range diffs {
+		if tcfPathTokens(d.changed)[scope] {
+			return nil, nil
+		}
+		common, ok := capture(git("-C", d.worktree, "rev-parse", "--path-format=absolute", "--git-common-dir"))
+		if !ok {
+			return nil, fmt.Errorf("git cannot name the repository of %s", d.worktree)
+		}
+		own[common] = true
+		if slices.Contains(tcfWords(filepath.Base(filepath.Dir(common))), scope) {
+			return nil, nil
+		}
+		subjects, ok := capture(git("-C", d.worktree, "log", "--format=%s", d.parent))
+		if !ok {
+			return nil, fmt.Errorf("git cannot read the history of %s at %s", d.worktree, d.parent)
+		}
+		for _, s := range strings.Split(subjects, "\n") {
+			if e := tcfSubjectScopeRE.FindStringSubmatch(s); e != nil && strings.ToLower(e[1]) == scope {
+				return nil, nil
+			}
+		}
+		dirs, ok := capture(git("-C", d.worktree, "ls-tree", "-r", "-d", "--name-only", d.commit))
+		if !ok {
+			return nil, fmt.Errorf("git cannot list the directories of %s at %s", d.worktree, d.commit)
+		}
+		segs := map[string]bool{}
+		for _, p := range strings.Split(dirs, "\n") {
+			for _, seg := range strings.Split(p, "/") {
+				segs[strings.ToLower(seg)] = true
+			}
+		}
+		parts := strings.Split(scope, "-")
+		if !slices.ContainsFunc(parts, func(p string) bool { return !segs[p] }) {
+			return nil, nil
+		}
+	}
+	for _, root := range tcfAppRoots(projectSection(filepath.Join(projectDir, ".flow", "project.md"), "apps")) {
+		common, ok := capture(git("-C", root, "rev-parse", "--path-format=absolute", "--git-common-dir"))
+		if !ok || own[common] {
+			continue
+		}
+		files, ok := capture(git("-C", root, "ls-tree", "-r", "--name-only", "HEAD"))
+		if !ok {
+			continue
+		}
+		if tcfPathTokens(strings.Split(files, "\n"))[scope] {
+			return []string{fmt.Sprintf("task %s: declared Commit: scope %s names a module of %s and nothing this commit touched, this repository's directories or its earlier scopes — name the area the commit moved", task.id, tcfRepr(m[1]), filepath.Base(filepath.Dir(common)))}, nil
+		}
+	}
+	return nil, nil
+}
+
+// tcfAppRoots is every absolute, backticked `Repo root` cell of an `## apps`
+// table body.
+func tcfAppRoots(apps string) []string {
+	var roots []string
+	for _, line := range strings.Split(apps, "\n") {
+		cells := strings.Split(line, "|")
+		if len(cells) < 4 {
+			continue
+		}
+		cell := strings.Trim(strings.TrimSpace(cells[2]), "`")
+		if strings.HasPrefix(cell, "/") {
+			roots = append(roots, cell)
+		}
+	}
+	return roots
+}
+
+// tcfScopeWordRE is one camelCase word of a path part; tcfScopeSepRE splits
+// a name at `-`, `_` and `.`.
+var (
+	tcfScopeWordRE = regexp.MustCompile(`[A-Z]?[a-z0-9]+|[A-Z]+`)
+	tcfScopeSepRE  = regexp.MustCompile(`[-_.]`)
+)
+
+// tcfWords is a name's `-`, `_` and `.` parts, lowercased.
+func tcfWords(name string) []string {
+	var out []string
+	for _, p := range tcfScopeSepRE.Split(strings.ToLower(name), -1) {
+		if p != "" {
+			out = append(out, p)
+		}
+	}
+	return out
+}
+
+// tcfPathTokens is every token a scope may name a path by: each segment
+// whole and without its extension, its `-`/`_`/`.` parts, and their
+// camelCase words, lowercased.
+func tcfPathTokens(paths []string) map[string]bool {
+	t := map[string]bool{}
+	for _, p := range paths {
+		for _, seg := range strings.Split(p, "/") {
+			if seg == "" {
+				continue
+			}
+			stem, _, _ := strings.Cut(seg, ".")
+			for _, x := range append([]string{seg, stem}, tcfScopeSepRE.Split(seg, -1)...) {
+				if x == "" {
+					continue
+				}
+				t[strings.ToLower(x)] = true
+				for _, w := range tcfScopeWordRE.FindAllString(x, -1) {
+					t[strings.ToLower(w)] = true
+				}
+			}
+		}
+	}
+	return t
+}
+
 // tcfCheckFiles is check_files.
 func tcfCheckFiles(task tcfTask, changed []string) []string {
 	var v []string
````

  - [ ] **Step 3: Verify.** `cd stats && gofmt -l ./internal/guard && go vet ./internal/guard && go test ./internal/guard/ -run 'TestCheckScopeSibling|TestCheckTaskCommitFields|TestTaskFieldParseMatchesPython' -count=1` passes; `scripts/test-git-config-pins.sh`, `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `python3 scripts/check-markdown-integrity.py` and `scripts/check-installed-citations.sh` exit 0.
  - [ ] **Step 4: Commit** the paths in `**Files:**` with the subject below.

**Build:** green
**Files:** `stats/internal/guard/taskcommitfields.go`, `stats/internal/guard/check_task_commit_scope_sibling_test.go`, `rules/commit-scope-is-the-module.mdc`
**Tests:** `TestCheckScopeSibling`
**Regression:** reverting the commit lets a declared `fix(group)` close clean on a commit touching only gymie-playwright's `fidelity/builder.spec.ts`, `group` being a gymie module; `TestCheckScopeSibling` cannot compile without `tcfCheckScopeSibling`.
**Baseline:** before=177 after=178
<!-- measured: cat stats/internal/guard/*_test.go | grep -cE '^func Test[A-Za-z0-9_]*\(' @ 8650ac1f with Task 1 applied (before; after with this task's diff applied in the scratch clone) -->
**Commit:** feat(guard): refuse a commit scope naming only another repository's module
**After:** Task 1

**Decision:** scope-sibling-module
**Decision:** scope-rule-fourth-shape
