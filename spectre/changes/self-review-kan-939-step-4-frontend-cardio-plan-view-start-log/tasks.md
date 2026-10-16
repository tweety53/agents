# self-review-kan-939-step-4-frontend-cardio-plan-view-start-log

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** fix the eight non-`big` findings of the kan-939 step-4 self-review pass, and the root-less citation planning found.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-16-self-review-kan-939-step-4-frontend-cardio-plan-view-start-log-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`.
- **Task 7 commits in the gymie-frontend worktree `/Users/tweety53/Projects/gymie-frontend-worktrees/kan-939-step-4-frontend-cardio-plan-view-start-log`** (branch `spectre/kan-939-step-4-frontend-cardio-plan-view-start-log`), never in this agents worktree; every other task commits in this agents worktree. Never merge or push either.
- Each task's change is the unified diff its step gives: copy the fenced `diff` block's body verbatim to a file outside both repositories and `git apply` it from the root of the repository the task names. Every diff was applied in plan order and checked there. A diff that does not apply is reported, never hand-merged by guesswork.
- Every prose sentence a diff adds or rewords is already acknowledged in this change's `verbatim-moves.txt`; any other wording fails `scripts/check-verbatim-moves.sh`.
- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w .` before `gofmt -l .`. The commands below vet and test only the packages each task touches; Bash calls running `go test` set timeout 600000.

## Review focus

- A cross-repo change whose appended fix task committed only in the peer worktree must be CLEAR on signal four's commit half when every worktree's range is passed.
- A verifier row recorded before a sync onto a moved base whose new commits touch UI paths the change does not touch must still count.
- A panel run whose rounds each recorded a pass row must stay CLEAR on signal five, and a change with no `panel-` reviewer row must make no pass-row read at all.
- `flow record dispatch backfill` with a malformed `-ended-at` must write no row at all, not a begun row with no end.
- An integrate re-run that still passes `--plan-message` must stop before the reshape, not after it.
- [ ] 1. integrate's planning commit keeps its own subject and carries the outstanding work in its body
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n 'plan-message' scripts/integrate-change.sh skills/flow/integrate.md` and confirm `--plan-message) PLAN_MSG="${2:-}"` replaces the default `chore(spectre): plan $NAME` whole, and `integrate.md` says to pass the planning message "whole as `--plan-message`" — the defect: a message with no Conventional Commits subject stops `commit` at commit-split, after the reshape.
  - [ ] **Step 2: Write the failing test.** Apply only the harness hunks of the diff in step 3 (`git apply --include='scripts/test-integrate-change.sh' <the diff>`) and run `bash scripts/test-integrate-change.sh 2>&1 | tail -4`: the `outstanding:` checks FAIL.
  - [ ] **Step 3: Implement.** Apply the rest of this diff (`git apply --exclude='scripts/test-integrate-change.sh' <the diff>`):

````diff verified:applied with git apply to the worktree at a7196b56 in plan order; scripts/test-integrate-change.sh 101 passed, 0 failed; check-verbatim-moves.sh with this change's verbatim-moves.txt exits 0
diff --git a/scripts/integrate-change.sh b/scripts/integrate-change.sh
index fd683411..d0e21ed2 100755
--- a/scripts/integrate-change.sh
+++ b/scripts/integrate-change.sh
@@ -14,7 +14,7 @@
 #
 # Usage:
 #   integrate-change.sh prepare <main-checkout> <name> [--accept-foreign] [--accept-outstanding]
-#   integrate-change.sh commit  <main-checkout> <name> [<worktree>=<impl subject>]... [--plan-message <msg>]
+#   integrate-change.sh commit  <main-checkout> <name> [<worktree>=<impl subject>]... [--outstanding <text>]
 #   integrate-change.sh land    <main-checkout> <name> [--route <route>] [--pr-url <worktree>=<url|none>]...
 #   integrate-change.sh cleanup <main-checkout> <name> [--proceed <main-checkout>]...
 # Every subcommand also takes --harness <harness> and --session-token
@@ -47,6 +47,8 @@
 #          archived change without new work skips straight to commit-archive.
 #          Then the --force-with-lease branch push per worktree, in merge
 #          order — the reshape's push (Branch backup, git-boundaries.md).
+#          --outstanding <text> puts `Outstanding: <text>` in the planning
+#          commit's body, under this script's own subject.
 #          The parent appends its narrative before calling this; an
 #          `In-run pipeline fix:` line in it off the format pipeline.md's
 #          Pipeline defects found mid-run fixes stops first, STOP: narrative.
@@ -113,7 +115,12 @@ while [ $# -gt 0 ]; do
     --session-token) TOKEN="${2:-}"; shift ;;
     --accept-foreign) ACCEPT_FOREIGN=1 ;;
     --accept-outstanding) ACCEPT_OUTSTANDING=1 ;;
-    --plan-message) PLAN_MSG="${2:-}"; shift ;;
+    # The planning commit's subject is always this script's own: the
+    # outstanding work rides in its body, so a text with no Conventional
+    # Commits subject never reaches a commit-msg hook as one (it stopped
+    # commit-split after the reshape when it replaced the subject whole).
+    --outstanding) [ -n "${2:-}" ] || die "--outstanding takes the outstanding work's text"
+      PLAN_MSG="$(printf 'chore(spectre): plan %s\n\nOutstanding: %s' "$NAME" "$2")"; shift ;;
     --route) ROUTE="${2:-}"; shift ;;
     --pr-url) case "${2:-}" in ?*=?*) PR_ARGS+=("$2") ;; *) die "--pr-url takes <worktree>=<url|none>" ;; esac; shift ;;
     --proceed)
diff --git a/scripts/test-integrate-change.sh b/scripts/test-integrate-change.sh
index 9c2ddf0a..7fea70b4 100755
--- a/scripts/test-integrate-change.sh
+++ b/scripts/test-integrate-change.sh
@@ -71,6 +71,9 @@
 #   peer_fix_canon_first  peer_fix with the merge order `.` then `b`: the
 #                   rewrite lands after a was pushed, and land pushes a's
 #                   branch and base again in the same run.
+#   outstanding     commit --outstanding keeps the planning commit's own
+#                   subject and puts the text in its body; --plan-message,
+#                   which replaced the subject whole, is refused.
 #
 # Bash 3.2 is the floor: indexed arrays only.
 set -u
@@ -477,5 +480,17 @@ printf '# demo\n\n## Parts\n\n`b:demo`\n\n## Branch\n\nmain\n\n## Merge order\n\
 g -C "$WORK/a-worktrees/demo" commit -qam "chore(spectre): plan-gate demo" --amend
 peer_fix_case "peer_fix_canon_first: "
 
+# --- outstanding ---
+fixture
+sut prepare "$WORK/a" demo
+[ "$RC" -eq 0 ] || halt prepare
+printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
+sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work" --plan-message "Task 3 is unticked"
+check "outstanding: --plan-message is refused" '[ "$RC" -ne 0 ] && grep -q "unknown argument .--plan-message." "$WORK/stderr"'
+sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work" --outstanding "Task 3 is unticked"
+check "outstanding: commit exits 0" '[ "$RC" -eq 0 ]'
+check "outstanding: the planning commit keeps its own subject and carries the text in its body" \
+  '[ "$(g -C "$WORK/a-worktrees/demo" log --format=%s --grep="^Outstanding: " -1)" = "chore(spectre): plan demo" ] && [ "$(g -C "$WORK/a-worktrees/demo" log --format=%b --grep="^Outstanding: " -1)" = "Outstanding: Task 3 is unticked" ]'
+
 printf '\n%d passed, %d failed\n' "$PASS" "$FAIL"
 [ "$FAIL" -eq 0 ]
diff --git a/skills/flow/integrate.md b/skills/flow/integrate.md
index 4f2dd7a4..687f75bd 100644
--- a/skills/flow/integrate.md
+++ b/skills/flow/integrate.md
@@ -98,8 +98,9 @@ integrate-change.sh commit <main-checkout> <name> <worktree>="<type>(<module>):
 diffstat `prepare` printed under each `WORKTREE:` line, one subject per worktree the `NEXT:` line
 names. A worktree whose reshaped diff holds nothing outside `<project>/spectre/changes/` and no change directory's
 `link.md` has an empty implementation commit, so `NEXT:` names no subject for it and `commit` asks none. On
-**Continue** at **1** the planning message also lists the outstanding work: pass it whole as
-`--plan-message`. An archived change without new work takes no subject: its `NEXT:` names
+**Continue** at **1** the planning commit also lists the outstanding work: pass it as
+`--outstanding <text>`, which `commit` writes into that commit's body as `Outstanding: <text>` under
+its own `chore(spectre): plan <name>` subject. An archived change without new work takes no subject: its `NEXT:` names
 `commit` bare.
 
 `commit` reshapes, renders the ledger, runs the two-commit chain per worktree in `## Merge order`,
````

  - [ ] **Step 4: Verify.** `bash scripts/test-integrate-change.sh 2>&1 | tail -3` ends `101 passed, 0 failed`; `bash -n scripts/integrate-change.sh`; `scripts/check-verbatim-moves.sh` (with this change's `verbatim-moves.txt`), `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-normative-inventory.sh` exit 0 (`check-installed-citations.sh` exits 1 on `skills/flow/verify-and-handoff.md:207` alone until Task 9). Commit.

**Build:** green
**Files:** `scripts/integrate-change.sh`, `scripts/test-integrate-change.sh`, `skills/flow/integrate.md`
**Tests:** `outstanding: --plan-message is refused`, `outstanding: commit exits 0`, `outstanding: the planning commit keeps its own subject and carries the text in its body`
**Regression:** reverting the commit makes `--outstanding` an unknown argument and `--plan-message` accepted again, replacing the planning commit's subject with whatever text the parent passed.
**Baseline:** before=72 after=75
<!-- measured: grep -cE '^[[:space:]]*check "' scripts/test-integrate-change.sh @ a7196b56 (before; after with the earlier tasks' diffs and this task's applied in the worktree) -->
**Commit:** fix(scripts): integrate's planning commit keeps its subject and carries the outstanding work in its body
**After:** none

**Decision:** outstanding-in-body

- [ ] 2. The push-time untasked-commit check lives in Commits outside every task and runs before a panel fix's push
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n 'check-untasked-commits' skills/flow/implement.md skills/flow/review-panel-fix-round.md skills/flow-contracts/git-boundaries.md` and confirm the "Before it is pushed" sentence sits only in `implement.md`'s full-suite fix paragraph — the defect: a panel-fix commit can be pushed with no `Flow-Key:` trailer.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:applied with git apply to the worktree at a7196b56 in plan order; check-verbatim-moves.sh with this change's verbatim-moves.txt, check-references.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh, check-normative-inventory.sh exit 0
diff --git a/skills/flow-contracts/git-boundaries.md b/skills/flow-contracts/git-boundaries.md
index 1dfd8df8..12736743 100644
--- a/skills/flow-contracts/git-boundaries.md
+++ b/skills/flow-contracts/git-boundaries.md
@@ -134,6 +134,7 @@ records no such row, the stage key (`flow.review-panel`, `flow.visual-verify`, `
 carries `Task-Id:` instead, and a planning commit neither. Run 1's unfinished-work gate lists
 every other non-`chore` commit with `check-untasked-commits.sh`
 (`skills/flow-contracts/finish-contract-run1.md`).
+**Before such a commit is pushed, `check-untasked-commits.sh <worktree> <merge-base>` runs**; when the only sha it lists is that commit's, `git commit --amend` adds the missing trailer and the guard runs again; any other listed commit goes to run 1's unfinished-work gate, which handles `UNTASKED`.
 
 ## Revert and back-out commits
 
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index 5ee530d5..f5177af9 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -619,7 +619,7 @@ dictates (**Panel re-runs**, `skills/flow/review-panel-fix-round.md`), records t
 `dispatches` pair `-role panel-fix -model <parent model> -effort <parent effort> -agent-id inline`
 under `full-suite-fix-<n>`, and re-runs the list until it passes.
 The commit carries `-m "Flow-Key: full-suite-fix-<n>"` (**Commits outside every task**, `skills/flow-contracts/git-boundaries.md`).
-Before it is pushed, `check-untasked-commits.sh <worktree> <merge-base>` runs; when the only sha it lists is this commit's, `git commit --amend` adds the missing trailer and the guard runs again; any other listed commit goes to run 1's unfinished-work gate, which handles `UNTASKED`.
+Before it is pushed, the untasked-commit check of **Commits outside every task** (`skills/flow-contracts/git-boundaries.md`) runs.
 The pair's `end` carries `-commit` with the sha `git -C <worktree> rev-parse HEAD` prints once that commit is made, never a task's commit.
 A fix for a failure that does not recur on every run is a determinism fix, held to the PROOF RUNS paragraph above: the failure is shown returning with the fix reverted before the fix is committed, and a fix whose failure never returns is not committed.
 When the change's diff touches a `## visual verification` `ui paths` glob
diff --git a/skills/flow/review-panel-fix-round.md b/skills/flow/review-panel-fix-round.md
index 388ab662..9af2ba5b 100644
--- a/skills/flow/review-panel-fix-round.md
+++ b/skills/flow/review-panel-fix-round.md
@@ -32,7 +32,7 @@ This is the normal case: the branch is pushed with every commit (**Branch backup
 (`git add -- <the changed paths>` — a pathspec commit reads tracked paths only) and is a plain
 `git commit -m ... -- <the changed paths>` at the tip — the pathspec-scoped default (**Branch backup**, `skills/flow-contracts/git-boundaries.md`),
 so it carries only the paths the finding named, whatever else the index holds — pushed
-plain like any other commit, and every downstream commit keeps its sha. The commit carries the round's `Flow-Key:` trailer — `git commit -m <subject> -m "Flow-Key: <the round's dispatch key>" -- <the changed paths>` (**Commits outside every task**, `skills/flow-contracts/git-boundaries.md`). **A branch a
+plain like any other commit, and every downstream commit keeps its sha. The commit carries the round's `Flow-Key:` trailer — `git commit -m <subject> -m "Flow-Key: <the round's dispatch key>" -- <the changed paths>` (**Commits outside every task**, `skills/flow-contracts/git-boundaries.md`). Before it is pushed, the untasked-commit check of **Commits outside every task** (`skills/flow-contracts/git-boundaries.md`) runs. **A branch a
 round-boundary auto-rebase has rewritten pushes with `--force-with-lease`, never a bare
 `git push` and never `--force`** — the rebase replaced the shas the remote still holds, so the
 plain push above is then rejected.
````

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh` (with this change's `verbatim-moves.txt`), `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-normative-inventory.sh` exit 0 (`check-installed-citations.sh` as in Task 1). Commit.

**Build:** green
**Files:** `skills/flow-contracts/git-boundaries.md`, `skills/flow/implement.md`, `skills/flow/review-panel-fix-round.md`
**Tests:** **none** — a sentence moved into its canonical section and cited from two routes; the guards in step 3 cover it, and the check it names already exists.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): the push-time untasked-commit check covers every commit outside a task, panel fixes included
**After:** none

**Decision:** untasked-check-in-git-boundaries

- [ ] 3. Signal four reports a ticked appended task with no Task-Id commit in the change's ranges
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n 'dispatch rows are the proxy\|Task-Id' stats/internal/guard/unfinishedwork.go` and confirm signal four judges appended tasks by dispatch rows alone and reads no commit — the defect: a back-filled row satisfies it for a task whose work never landed.
  - [ ] **Step 2: Write the failing tests.** Apply only the test hunk of the diff in step 3 (`git apply --include='stats/internal/guard/check_unfinished_work_test.go' <the diff>`) and run `cd stats && go test ./internal/guard/ -run TestCheckUnfinishedWork -count=1`: the new subtests FAIL.
  - [ ] **Step 3: Implement.** Apply the rest of this diff (`git apply --exclude='stats/internal/guard/check_unfinished_work_test.go' <the diff>`):

````diff verified:applied with git apply to the worktree at a7196b56 in plan order; gofmt -l and go vet clean; go test ./internal/guard/ passes; the new subtests fail with unfinishedwork.go reverted; scripts/test-integrate-change.sh 101 passed; check-verbatim-moves.sh with this change's verbatim-moves.txt exits 0
diff --git a/scripts/check-unfinished-work.sh b/scripts/check-unfinished-work.sh
index 5dbd0506..580606b4 100755
--- a/scripts/check-unfinished-work.sh
+++ b/scripts/check-unfinished-work.sh
@@ -2,7 +2,7 @@
 # check-unfinished-work.sh — report whether a change carries unfinished work,
 # for /flow's integrate-run gate.
 #
-# Usage: check-unfinished-work.sh <worktree> <change-name> [canonical-worktree]
+# Usage: check-unfinished-work.sh <worktree> <change-name> [canonical-worktree] [<worktree>=<merge-base>]...
 #
 # Prints ONE verdict line to stdout:
 #   CLEAR: <reason>          nothing outstanding
@@ -64,6 +64,12 @@
 # row naming it: the fix run's panel reviews it under keys that name no task,
 # so any panel reviewer row of the change counts, and a reviewer row under any
 # other key does not.
+# A ticked appended task also needs a commit carrying `Task-Id: <n>` in one of
+# the `<worktree>=<merge-base>` ranges passed — one per worktree of the
+# change's resolved set, since a cross-repo fix task may commit in a peer
+# repository: a dispatch row back-filled after the fact is not a commit. A
+# plan with a ticked appended task and no range passed is exit 2, and a range
+# git cannot read is exit 2.
 #
 # All signals are counted independently and reported together on the one
 # line, so the operator sees the whole picture in one prompt rather than being
diff --git a/scripts/integrate-change.sh b/scripts/integrate-change.sh
index d0e21ed2..1fb3c0e1 100755
--- a/scripts/integrate-change.sh
+++ b/scripts/integrate-change.sh
@@ -378,10 +378,13 @@ cmd_prepare() {
     mark_end flow.unfinished-work-gate completed
   else
     ask=0
+    # Every worktree's commit range, for signal four's Task-Id search.
+    local ranges=()
+    for i in "${!WT[@]}"; do ranges+=("${WT[i]}=$(git -C "${WT[i]}" merge-base HEAD "origin/${BASES[i]}")"); done
     # A guard that reached no verdict is not OUTSTANDING: its own stop,
     # which --accept-outstanding never passes.
     for i in "${!WT[@]}"; do
-      run "$SCRIPT_DIR/check-unfinished-work.sh" "${WT[i]}" "$NAME" "$CANON"
+      run "$SCRIPT_DIR/check-unfinished-work.sh" "${WT[i]}" "$NAME" "$CANON" "${ranges[@]}"
       case "$RC:$(verdict 'CLEAR|OUTSTANDING')" in
         0:CLEAR:*) ;;
         0:OUTSTANDING:*) ask=1 ;;
diff --git a/skills/flow-contracts/finish-contract-run1.md b/skills/flow-contracts/finish-contract-run1.md
index a7962217..180cd782 100644
--- a/skills/flow-contracts/finish-contract-run1.md
+++ b/skills/flow-contracts/finish-contract-run1.md
@@ -138,13 +138,14 @@ either way: the base-moved check and **Sync the branch onto the base**
 self-review pass, and the route.
 
 **Check for unfinished work first — before the landing question and before any git action but the undo of an uncommitted archive above.**
-`check-unfinished-work.sh <worktree> <change-name> [canonical-worktree]` prints one verdict line and
+`check-unfinished-work.sh <worktree> <change-name> [canonical-worktree] [<worktree>=<merge-base>]...` prints one verdict line and
 exits 0 whenever it reached a verdict. It exits 2 with **no** verdict line when it cannot read the
 worktree. Run it once per worktree in the set found by **Resolving a change's worktrees** below —
 never a raw read of the state file's `worktrees` map, for the same reason the preflight verdict
 above does not read it raw. Pass `[canonical-worktree]` on every call in the run: the one member of
 the resolved set whose own change directory (**A change's directory**,
 `skills/flow-contracts/pipeline.md`) holds `tasks.md`.
+Pass every worktree of the resolved set as `<worktree>=<merge-base>`, its merge base `git -C <worktree> merge-base HEAD origin/<base>`, on every call.
 
 | Verdict | Meaning |
 |---------|---------|
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index 31633a5b..0f4de81d 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -396,7 +396,7 @@ flow record dispatch end -change <name> -C <canonical-worktree> -key <key> -sess
 
 **A fix run then runs the unfinished-work guard on itself**, once per worktree of the resolved
 set, with the arguments **Run 1 — the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`) passes it —
-`check-unfinished-work.sh <worktree> <name> <canonical-worktree>` — so an appended task left
+`check-unfinished-work.sh <worktree> <name> <canonical-worktree> <worktree>=<merge-base>...` — so an appended task left
 unrecorded or unreviewed stops this run rather than the next integrate. An `OUTSTANDING:` line is
 this run's own unfinished work: finish what its breakdown names — a missing dispatch or review is
 run and recorded, never recorded after the fact for work that did not run — and run the guard
diff --git a/stats/internal/guard/check_unfinished_work_test.go b/stats/internal/guard/check_unfinished_work_test.go
index 6fa15d53..50a45071 100644
--- a/stats/internal/guard/check_unfinished_work_test.go
+++ b/stats/internal/guard/check_unfinished_work_test.go
@@ -40,6 +40,19 @@ type uwFix struct {
 	env      Env
 	findings []byte
 	fErr     error
+	ranges   []string // <worktree>=<merge-base> arguments, after an empty canonical
+}
+
+// commits makes dir a repository whose base commit carries baseMsg and
+// whose one later commit carries msg -- both empty, so the fixture's files
+// stay untracked -- and returns `<dir>=<base sha>`, a range argument.
+func uwCommits(t *testing.T, dir, baseMsg, msg string) string {
+	t.Helper()
+	gitRun(t, "", "init", "-q", dir)
+	gitRun(t, dir, "commit", "-q", "--allow-empty", "-m", baseMsg)
+	base := tcfGit(t, dir, "rev-parse", "HEAD")
+	gitRun(t, dir, "commit", "-q", "--allow-empty", "-m", msg)
+	return dir + "=" + base
 }
 
 // uwResult is one run: exit code, stdout (trailing newlines stripped, as
@@ -236,6 +249,9 @@ func TestCheckUnfinishedWork(t *testing.T) {
 			if setup != nil {
 				setup(t, f)
 			}
+			if len(f.ranges) > 0 {
+				return f.run(t, append([]string{f.wt, "demo", ""}, f.ranges...)...)
+			}
 			return f.run(t, f.wt, "demo")
 		}
 	}
@@ -277,6 +293,21 @@ func TestCheckUnfinishedWork(t *testing.T) {
 		return func(t *testing.T, f *uwFix) {
 			f.plan(t, body)
 			f.setFile(t, "dispatches.json", rows+"\n")
+			f.ranges = []string{uwCommits(t, f.wt, "base", "fix: the appended task\n\nTask-Id: 7")}
+		}
+	}
+	// covered is a ticked appended task 7 with its implementer row and a
+	// later panel reviewer row: only its commit is left to judge.
+	covered := `[{"seq":1,"key":"task-7-implementer","role":"implementer","outcome":"completed"},{"seq":2,"key":"panel-2-primary","role":"reviewer","outcome":"clean"}]`
+	committedAs := func(baseMsg, msg string, peer bool) func(t *testing.T, f *uwFix) {
+		return func(t *testing.T, f *uwFix) {
+			f.plan(t, "> **Tasks appended:** 1\n\n- [x] 1. done\n- [x] 7. appended by a fix run\n")
+			f.setFile(t, "dispatches.json", covered+"\n")
+			if peer {
+				f.ranges = []string{uwCommits(t, f.wt, "base", "fix: no trailer"), uwCommits(t, t.TempDir(), "base", msg)}
+				return
+			}
+			f.ranges = []string{uwCommits(t, f.wt, baseMsg, msg)}
 		}
 	}
 	appended := "> **Tasks appended:** 1\n\n- [x] 1. done\n- [x] 2. grouped\n- [x] 7. appended by a fix run\n"
@@ -298,6 +329,26 @@ func TestCheckUnfinishedWork(t *testing.T) {
 		uwVerdict("a non-panel reviewer row does not review the appended task", "OUTSTANDING:"))
 	each(demo(planned("- [x] 1. done\n- [x] 2. grouped\n- [x] 3. grouped\n", `[{"key":"task-2-implementer","outcome":"completed"}]`)),
 		uwVerdict("plan-time tasks grouped under one key, with no Tasks appended line, are CLEAR", "CLEAR:"))
+	// 2d. Signal four's commit half — a dispatch row is not a commit (a
+	//     back-filled row named a task whose work never landed).
+	each(demo(committedAs("base", "fix: the appended task\n\nFlow-Key: panel-fix-2", false)),
+		uwVerdict("an appended task with its rows but no Task-Id commit is OUTSTANDING", "OUTSTANDING:"),
+		uwReason("the breakdown names the uncommitted appended task", "1 ticked appended task(s) with no commit carrying its `Task-Id:` trailer on the branch (Task 7)"))
+	each(demo(committedAs("fix: before the branch\n\nTask-Id: 7", "fix: no trailer", false)),
+		uwVerdict("a Task-Id commit at or before the merge base does not count", "OUTSTANDING:"))
+	each(demo(committedAs("base", "fix: the appended task\n\nTask-Id: 7", true)),
+		uwVerdict("a Task-Id commit in a peer worktree's range covers the appended task", "CLEAR:"))
+	each(func(t *testing.T) uwResult {
+		f := newUW(t)
+		f.plan(t, "> **Tasks appended:** 1\n\n- [x] 7. appended by a fix run\n")
+		f.setFile(t, "dispatches.json", covered+"\n")
+		return f.run(t, f.wt, "demo")
+	}, uwRC("ticked appended tasks with no range passed cannot answer", 2))
+	each(demo(nil), uwVerdict("a plan with no appended task needs no range", "CLEAR:"))
+	each(func(t *testing.T) uwResult {
+		f := newUW(t)
+		return f.run(t, f.wt, "demo", "", "no-equals-sign")
+	}, uwRC("a range argument not shaped <worktree>=<merge-base> is refused", 2))
 
 	// 3. Signal one — the plan.
 	each(demo(unticked),
diff --git a/stats/internal/guard/unfinishedwork.go b/stats/internal/guard/unfinishedwork.go
index fd2d2c9c..9b406eda 100644
--- a/stats/internal/guard/unfinishedwork.go
+++ b/stats/internal/guard/unfinishedwork.go
@@ -35,9 +35,21 @@ func checkUnfinishedWork(args []string, env Env, stdout, stderr io.Writer) int {
 	}
 	worktree, name, canonical := arg(0), arg(1), arg(2)
 	if worktree == "" || name == "" {
-		fmt.Fprintln(stderr, "usage: check-unfinished-work.sh <worktree> <change-name> [canonical-worktree]")
+		fmt.Fprintln(stderr, "usage: check-unfinished-work.sh <worktree> <change-name> [canonical-worktree] [<worktree>=<merge-base>]...")
 		return 2
 	}
+	// The change's commit ranges, one per worktree of its resolved set:
+	// signal four's commit half searches them for each ticked appended
+	// task's `Task-Id:` trailer.
+	var ranges [][2]string
+	for _, a := range args[min(3, len(args)):] {
+		wt, base, ok := strings.Cut(a, "=")
+		if !ok || wt == "" || base == "" {
+			fmt.Fprintf(stderr, "%s'%s' is not <worktree>=<merge-base>\n", uwPrefix, a)
+			return 2
+		}
+		ranges = append(ranges, [2]string{wt, base})
+	}
 
 	// CONTAINMENT: the change name arrives from a pull-request-editable state
 	// file and is concatenated into every path below, and passed to `flow
@@ -347,6 +359,37 @@ func checkUnfinishedWork(args []string, env Env, stdout, stderr io.Writer) int {
 			if len(unreviewed) > 0 {
 				reasons = append(reasons, strconv.Itoa(len(unreviewed))+" ticked appended task(s) with no panel reviewer row (key `panel-<round>-<slot>`) begun after its implementer row (Task "+strings.Join(unreviewed, ", ")+") — run its fix round's panel and record its pass rows and findings, before landing")
 			}
+			// A dispatch row is not a commit: a row back-filled after the
+			// fact names a task whose work never landed. Each ticked
+			// appended task also needs a commit carrying `Task-Id: <n>` in
+			// one of the change's ranges -- all of them, since a cross-repo
+			// change's fix task may commit in a peer repository.
+			if len(appended) > 0 {
+				if len(ranges) == 0 {
+					fmt.Fprintf(stderr, "%sticked appended task(s) (Task %s) and no <worktree>=<merge-base> range was passed — cannot judge their commits\n", uwPrefix, strings.Join(appended, ", "))
+					return 2
+				}
+				committed := map[string]bool{}
+				for _, r := range ranges {
+					out, ok := capture(envGit(env)("-C", r[0], "log", "--no-merges", "--format=%(trailers:key=Task-Id,valueonly)", r[1]+"..HEAD"))
+					if !ok {
+						fmt.Fprintf(stderr, "%scannot read the commits of %s..HEAD in %s — cannot determine anything\n", uwPrefix, r[1], r[0])
+						return 2
+					}
+					for _, id := range ctTaskNumberRE.FindAllString(out, -1) {
+						committed[id] = true
+					}
+				}
+				var uncommitted []string
+				for _, id := range appended {
+					if !committed[id] {
+						uncommitted = append(uncommitted, id)
+					}
+				}
+				if len(uncommitted) > 0 {
+					reasons = append(reasons, strconv.Itoa(len(uncommitted))+" ticked appended task(s) with no commit carrying its `Task-Id:` trailer on the branch (Task "+strings.Join(uncommitted, ", ")+") — commit each task's work with `Task-Id: <n>` before landing")
+				}
+			}
 		}
 	}
 
````

  - [ ] **Step 4: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckUnfinishedWork|TestCloseTask' -count=1` — `gofmt -l` prints nothing, the tests pass; `bash scripts/test-integrate-change.sh 2>&1 | tail -1` reads `101 passed, 0 failed`; `scripts/check-verbatim-moves.sh` (with this change's `verbatim-moves.txt`), `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-normative-inventory.sh` exit 0 (`check-installed-citations.sh` as in Task 1). Commit.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `scripts/check-unfinished-work.sh`, `scripts/integrate-change.sh`, `skills/flow-contracts/finish-contract-run1.md`, `skills/flow/verify-and-handoff.md`
**Tests:** `an appended task with its rows but no Task-Id commit is OUTSTANDING`, `the breakdown names the uncommitted appended task`, `a Task-Id commit at or before the merge base does not count`, `a Task-Id commit in a peer worktree's range covers the appended task`, `ticked appended tasks with no range passed cannot answer`, `a plan with no appended task needs no range`, `a range argument not shaped <worktree>=<merge-base> is refused`
**Regression:** reverting the commit lets a ticked appended task whose only evidence is a back-filled dispatch row pass the integrate gate with no commit carrying its `Task-Id:` trailer, and the range arguments integrate passes are refused as unknown.
**Baseline:** before=79 after=86
<!-- measured: grep -cE '(uwVerdict|uwRC|uwReason)\("' stats/internal/guard/check_unfinished_work_test.go @ a7196b56 (before; after with the earlier tasks' diffs and this task's applied in the worktree) -->
**Commit:** fix(flow-guard): the unfinished-work gate needs each ticked appended task's Task-Id commit
**After:** Task 1

**Decision:** task-id-commit-across-ranges

- [ ] 4. A verifier row counts only for the HEAD it verified
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n 'found = true\|diffBase' stats/internal/guard/visualverifydispatched.go` and confirm a completed verifier row sets `found` whatever its `diffBase` — the defect: UI-path commits landed after the verifier still pass `check-visual-verify-dispatched`.
  - [ ] **Step 2: Write the failing tests.** Apply only the test hunks of the diff in step 3 (`git apply --include='stats/internal/guard/check_visual_verify_dispatched_test.go' <the diff>`) and run `cd stats && go test ./internal/guard/ -run TestCheckVisualVerifyDispatched -count=1`: cases 9a–9d FAIL.
  - [ ] **Step 3: Implement.** Apply the rest of this diff (`git apply --exclude='stats/internal/guard/check_visual_verify_dispatched_test.go' <the diff>`):

````diff verified:applied with git apply to the worktree at a7196b56 in plan order; gofmt -l and go vet clean; go test ./internal/guard/ passes; cases 9a-9d fail with visualverifydispatched.go reverted; check-verbatim-moves.sh with this change's verbatim-moves.txt exits 0
diff --git a/scripts/check-visual-verify-dispatched.sh b/scripts/check-visual-verify-dispatched.sh
index c5aa1d7d..ea695936 100755
--- a/scripts/check-visual-verify-dispatched.sh
+++ b/scripts/check-visual-verify-dispatched.sh
@@ -87,6 +87,21 @@
 # verifier from a dead one. See **The return** (skills/flow/implement.md)
 # for why only `completed` closes a dispatch's own story.
 #
+# A VERIFIER COUNTS ONLY FOR THE HEAD IT VERIFIED (kan-939 self-review: a
+# completed row answered OK however many UI commits landed after it). Such a
+# row is evidence only when its `diffBase` — the HEAD the verifier was
+# dispatched against, `-diff-base` on its `dispatch begin`
+# (skills/flow/visual-verify.md, **Recording**) — is a commit sha this
+# worktree resolves, and no path of this change's own `<merge-base>..HEAD`
+# diff that changed between that sha and HEAD matches the `ui paths`, asked
+# as the trigger above is, the canonical worktree's included. A row with no
+# `diffBase`, a ref name such as HEAD in place of a sha, or a sha this
+# worktree cannot resolve verified nothing here. With qualifying rows and
+# none fresh, the verdict is
+#   VISUAL-VERIFY-MISSING: a verifier dispatch is recorded for '<name>' but
+#   none verified a HEAD after this change's last UI-path commit — …
+# and a git diff that fails against a row's sha is exit 2.
+#
 # A SKIP DECIDED AT PLANNING IS NOT A SILENT SKIP. With no qualifying
 # verifier row, the guard reads `flow record decisions -change <name> -C
 # <canonical worktree when given>` (newest row first) and answers
diff --git a/skills/flow/visual-verify.md b/skills/flow/visual-verify.md
index 04b4371b..05e4bb15 100644
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -71,6 +71,7 @@ own system prompt>`.
 opus -effort low`, `-key visual-verify`, suffixed `-<worktree basename>` when this run's resolved set holds
 more than one worktree — the pair's semantics are section 4 of `skills/flow/implement.md`, cited
 here, not restated.
+Its `begin` carries `-diff-base` with the sha `git -C <worktree> rev-parse HEAD` prints before the dispatch, the HEAD it verifies: `check-visual-verify-dispatched.sh` counts a verifier row only for that HEAD, and only while no UI path the change touches has changed since.
 
 **Every report is checked for completeness before anything else reads it:**
 
diff --git a/stats/internal/guard/check_visual_verify_dispatched_test.go b/stats/internal/guard/check_visual_verify_dispatched_test.go
index b83ab6c2..cd0ef544 100644
--- a/stats/internal/guard/check_visual_verify_dispatched_test.go
+++ b/stats/internal/guard/check_visual_verify_dispatched_test.go
@@ -30,6 +30,7 @@ const (
 	vvdNoUI      = "VISUAL-VERIFY-OK: no UI paths touched\n"
 	vvdOK        = "VISUAL-VERIFY-OK: UI paths touched and a completed (or ended-but-unclosed) verifier dispatch is recorded for 'demo'\n"
 	vvdMissing   = "VISUAL-VERIFY-MISSING: this change's diff touched a declared UI path but no completed or ended 'verifier' dispatch (key starting 'visual-verify') is recorded for 'demo' — flow.visual-verify's stage marks were written with no verifier ever dispatched, its report was never read to completion, or the dispatch's closing outcome was lost to a session restart\n"
+	vvdStale     = "VISUAL-VERIFY-MISSING: a verifier dispatch is recorded for 'demo' but none verified a HEAD after this change's last UI-path commit — a row counts only with the -diff-base it verified, and only when no UI path this change touches changed since\n"
 	vvdSection   = "# project\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `app/src/**` |\n"
 )
 
@@ -123,11 +124,12 @@ var vvdMasters = sync.OnceValues(func() (map[string]vvdMaster, error) {
 	return masters, nil
 })
 
-// vvdRows is the harness's dispatch_json: rows from key/role/outcome triples.
+// vvdRows is the harness's dispatch_json: rows from key/role/outcome triples,
+// each verifying the case worktree's HEAD (vvdRun's @HEAD@).
 func vvdRows(vals ...string) []byte {
 	out := []map[string]string{}
 	for i := 0; i+2 < len(vals); i += 3 {
-		out = append(out, map[string]string{"key": vals[i], "role": vals[i+1], "outcome": vals[i+2]})
+		out = append(out, map[string]string{"key": vals[i], "role": vals[i+1], "outcome": vals[i+2], "diffBase": "@HEAD@"})
 	}
 	b, _ := json.Marshal(out)
 	return b
@@ -178,6 +180,12 @@ func vvdRun(t *testing.T, c vvdCase) vvdResult {
 		t.Fatal(err)
 	}
 	gdcCopyTree(t, m.dir, wt)
+	// A row names the HEAD it verified as @HEAD@ or the merge base as
+	// @BASE@: the shas exist only once the master is built.
+	head := tcfGit(t, wt, "rev-parse", "HEAD")
+	shas := strings.NewReplacer("@HEAD@", head, "@BASE@", m.base)
+	c.dispatches = []byte(shas.Replace(string(c.dispatches)))
+	c.flowStub = shas.Replace(c.flowStub)
 	bin := t.TempDir()
 	if c.flowStub != "" {
 		writeExec(t, bin+"/flow", c.flowStub)
@@ -271,9 +279,9 @@ func TestCheckVisualVerifyDispatched(t *testing.T) {
 	p := vvdPrefix
 	bad := vvdPrefix + "dispatch rows were not readable JSON — cannot answer\n"
 	const (
-		ended   = `[{"key":"visual-verify","role":"verifier","endedAt":"2026-09-20T10:00:00Z"}]`
-		blocked = `[{"key":"visual-verify","role":"verifier","outcome":"blocked","endedAt":"2026-09-20T10:00:00Z"}]`
-		open    = `[{"key":"visual-verify","role":"verifier"}]`
+		ended   = `[{"key":"visual-verify","role":"verifier","diffBase":"@HEAD@","endedAt":"2026-09-20T10:00:00Z"}]`
+		blocked = `[{"key":"visual-verify","role":"verifier","diffBase":"@HEAD@","outcome":"blocked","endedAt":"2026-09-20T10:00:00Z"}]`
+		open    = `[{"key":"visual-verify","role":"verifier","diffBase":"@HEAD@"}]`
 		flag    = `[{"falsePositive":true,"falsePositiveReason":"the verifier ran; the mark was lost to a session restart","change":"demo","flaggedAt":"2026-09-20T14:00:00Z"}]`
 	)
 	hint := func(n, last string) func(vvdResult) string {
@@ -319,6 +327,12 @@ func TestCheckVisualVerifyDispatched(t *testing.T) {
 		{"case 6b: an open row is not evidence, exit 1", vvdCase{dispatches: []byte(open)}, 1, vvdMissing, nil, nil},
 		{"case 6b: the verdict is VISUAL-VERIFY-MISSING", vvdCase{dispatches: []byte(open)}, 1, vvdMissing, nil, missingCalls},
 		{"case 7: a foreign key prefix is not evidence, exit 1", vvdCase{dispatches: vvdRows("some-other-key", v, done)}, 1, vvdMissing, nil, nil},
+		// ---- a verifier counts only for the HEAD it verified (kan-939) ----
+		{"case 9a: a verifier of a HEAD before the change's UI commit is stale, exit 1", vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","outcome":"completed","diffBase":"@BASE@"}]`)}, 1, vvdStale, nil, nil},
+		{"case 9b: a verifier row with no diffBase verified nothing, exit 1", vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","outcome":"completed"}]`)}, 1, vvdStale, nil, nil},
+		{"case 9c: a diffBase naming a ref, not a sha, verified nothing, exit 1", vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","outcome":"completed","diffBase":"HEAD"}]`)}, 1, vvdStale, nil, nil},
+		{"case 9d: a diffBase this worktree cannot resolve verified nothing here, exit 1", vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","outcome":"completed","diffBase":"0123456789abcdef0123456789abcdef01234567"}]`)}, 1, vvdStale, nil, nil},
+		{"case 9e: a later verifier of HEAD beside a stale one is OK", vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","outcome":"completed","diffBase":"@BASE@"},{"key":"visual-verify-2","role":"verifier","outcome":"completed","diffBase":"@HEAD@"}]`)}, 0, vvdOK, nil, nil},
 		{"case 8a: an unreachable store exits 2", vvdCase{dErr: errors.New("flow exit 1")}, 2, "",
 			func(vvdResult) string { return p + "flow record dispatches failed for 'demo' — cannot answer\n" }, nil},
 		{"case 8a: no verdict on an inability to answer", vvdCase{dErr: errors.New("flow exit 1")}, 2, "",
@@ -390,15 +404,15 @@ func TestCheckVisualVerifyDispatched(t *testing.T) {
 
 		// ---- branches the harness never reached, pinned from the bash -----
 		{"an ended row with an empty-string outcome is unclosed evidence",
-			vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","outcome":"","endedAt":"x"}]`)}, 0, vvdOK, nil, okCalls},
+			vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","diffBase":"@HEAD@","outcome":"","endedAt":"x"}]`)}, 0, vvdOK, nil, okCalls},
 		{"an ended row with a false outcome is unclosed evidence (jq's // reads false as absent)",
-			vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","outcome":false,"endedAt":"x"}]`)}, 0, vvdOK, nil, nil},
+			vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","diffBase":"@HEAD@","outcome":false,"endedAt":"x"}]`)}, 0, vvdOK, nil, nil},
 		{"an endedAt of false is an end (jq's != null)",
-			vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","endedAt":false}]`)}, 0, vvdOK, nil, nil},
+			vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","diffBase":"@HEAD@","endedAt":false}]`)}, 0, vvdOK, nil, nil},
 		{"an endedAt of null is still open",
-			vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","endedAt":null}]`)}, 1, vvdMissing, nil, nil},
+			vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","diffBase":"@HEAD@","endedAt":null}]`)}, 1, vvdMissing, nil, nil},
 		{"an ended row outside the key prefix is not evidence",
-			vvdCase{dispatches: []byte(`[{"key":"visual","role":"verifier","endedAt":"x"}]`)}, 1, vvdMissing, nil, nil},
+			vvdCase{dispatches: []byte(`[{"key":"visual","role":"verifier","diffBase":"@HEAD@","endedAt":"x"}]`)}, 1, vvdMissing, nil, nil},
 		{"an ended non-verifier row is not evidence",
 			vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"reviewer","endedAt":"x"}]`)}, 1, vvdMissing, nil, nil},
 		{"hint: a false-positive answer that is not an array prints nothing", vvdCase{dispatches: []byte("[]"), flags: []byte(`{"a":[1]}`)}, 1, vvdMissing, nil, nil},
@@ -437,16 +451,16 @@ func TestCheckVisualVerifyDispatched(t *testing.T) {
 		{"a UI file renamed away still touches its UI path (--no-renames)", vvdCase{repo: "rename", dispatches: []byte("[]")}, 1, vvdMissing, nil, nil},
 		{"a completed verifier row among others is found", vvdCase{dispatches: vvdRows("visual-verify", v, "aborted", "task-1-reviewer", "reviewer", done, "visual-verify-retry", v, done)}, 0, vvdOK, nil, nil},
 		{"the outcome must be exactly completed", vvdCase{dispatches: vvdRows("visual-verify", v, "Completed")}, 1, vvdMissing, nil, nil},
-		{"a null row selects nothing", vvdCase{dispatches: []byte(`[null,{"key":"visual-verify","role":"verifier","outcome":"completed"}]`)}, 0, vvdOK, nil, nil},
-		{"rows missing fields read as empty", vvdCase{dispatches: []byte(`[{},{"role":"verifier","key":false,"outcome":"completed"},{"role":"verifier","key":null}]`)}, 1, vvdMissing, nil, nil},
+		{"a null row selects nothing", vvdCase{dispatches: []byte(`[null,{"key":"visual-verify","role":"verifier","diffBase":"@HEAD@","outcome":"completed"}]`)}, 0, vvdOK, nil, nil},
+		{"rows missing fields read as empty", vvdCase{dispatches: []byte(`[{},{"role":"verifier","diffBase":"@HEAD@","key":false,"outcome":"completed"},{"role":"verifier","diffBase":"@HEAD@","key":null}]`)}, 1, vvdMissing, nil, nil},
 		{"a non-string key on a non-verifier row is never read", vvdCase{dispatches: []byte(`[{"role":"reviewer","key":1}]`)}, 1, vvdMissing, nil, nil},
-		{"a non-string key on a verifier row is cannot-answer", vvdCase{dispatches: []byte(`[{"role":"verifier","key":1}]`)}, 2, "", func(vvdResult) string { return bad }, nil},
+		{"a non-string key on a verifier row is cannot-answer", vvdCase{dispatches: []byte(`[{"role":"verifier","diffBase":"@HEAD@","key":1}]`)}, 2, "", func(vvdResult) string { return bad }, nil},
 		{"a non-object row is cannot-answer", vvdCase{dispatches: []byte(`["visual-verify"]`)}, 2, "", func(vvdResult) string { return bad }, nil},
-		{"a top-level object is cannot-answer", vvdCase{dispatches: []byte(`{"a":{"key":"visual-verify","role":"verifier","outcome":"completed"}}`)}, 2, "", func(vvdResult) string { return bad }, nil},
+		{"a top-level object is cannot-answer", vvdCase{dispatches: []byte(`{"a":{"key":"visual-verify","role":"verifier","diffBase":"@HEAD@","outcome":"completed"}}`)}, 2, "", func(vvdResult) string { return bad }, nil},
 		{"an empty store answer is cannot-answer", vvdCase{dispatches: []byte("")}, 2, "", func(vvdResult) string { return bad }, nil},
 		{"two JSON values are cannot-answer", vvdCase{dispatches: []byte("[] []")}, 2, "", func(vvdResult) string { return bad }, nil},
 		{"real flow on PATH answers the store read with the canonical worktree",
-			vvdCase{flowStub: vvdFlowRows(`[{"key":"visual-verify","role":"verifier","outcome":"completed"}]`, "[]")}, 0, vvdOK, nil, okCalls},
+			vvdCase{flowStub: vvdFlowRows(`[{"key":"visual-verify","role":"verifier","diffBase":"@HEAD@","outcome":"completed"}]`, "[]")}, 0, vvdOK, nil, okCalls},
 		{"real flow: MISSING records its verdict, reads the prior false positives and hints",
 			vvdCase{flowStub: vvdFlowRows("[]", flag)}, 1, vvdMissing, flagHint, missingCalls},
 		{"real flow exiting non-zero is cannot-answer",
@@ -546,7 +560,7 @@ func TestCheckVisualVerifyDispatchedCanonicalUIPaths(t *testing.T) {
 				wt = linked
 			}
 			writeFile(t, canon+"/.flow/project.md", strings.Replace(vvdSection, "`app/src/**`", c.uiPaths, 1))
-			writeExec(t, bin+"/flow", vvdFlowRows(c.rows, "[]"))
+			writeExec(t, bin+"/flow", vvdFlowRows(strings.ReplaceAll(c.rows, "@HEAD@", tcfGit(t, wt, "rev-parse", "HEAD")), "[]"))
 			env := Env{Dir: top, Getenv: func(k string) string {
 				if k == "PATH" {
 					return bin + ":" + os.Getenv("PATH")
diff --git a/stats/internal/guard/visualverifydispatched.go b/stats/internal/guard/visualverifydispatched.go
index 70666cbd..8ac9348c 100644
--- a/stats/internal/guard/visualverifydispatched.go
+++ b/stats/internal/guard/visualverifydispatched.go
@@ -6,6 +6,7 @@ import (
 	"fmt"
 	"io"
 	"path/filepath"
+	"regexp"
 	"strings"
 	"syscall"
 )
@@ -89,9 +90,7 @@ func checkVisualVerifyDispatched(args []string, env Env, stdout, stderr io.Write
 	// diff's paths go to visualTrigger (an empty diff is one empty line, as
 	// the bash's `printf '%s\n'` piped it), its stdout and stderr discarded,
 	// its three exit codes read as-is.
-	paths := strings.Split(changed, "\n")
-	triggerExit := visualTrigger(env, worktree, func() ([]string, error) { return paths, nil }, io.Discard, io.Discard)
-
+	//
 	// A SATELLITE IS ALSO MATCHED AGAINST THE CANONICAL `ui paths` (kan-982
 	// self-review): a cross-repo change's frontend worktree may declare no
 	// section of its own while the canonical project names its files as
@@ -102,37 +101,49 @@ func checkVisualVerifyDispatched(args []string, env Env, stdout, stderr io.Write
 	// read under the canonical worktree, the project the run's rows were
 	// recorded in; a second repository's worktree resolves another project
 	// key (unfinishedwork.go's anchor carries the reasoning).
-	anchor := worktree
+	anchor, canonical := worktree, ""
 	if canonicalArg != "" {
-		canonical, err := filepath.EvalSymlinks(pcAbs(env, canonicalArg))
-		if err != nil || !isDir(canonical) {
+		c, err := filepath.EvalSymlinks(pcAbs(env, canonicalArg))
+		if err != nil || !isDir(c) {
 			fmt.Fprintf(stderr, "%scanonical worktree is not a directory: %s\n", vvdPrefix, canonicalArg)
 			return 2
 		}
-		if canonical != worktree {
-			anchor = canonical
-			if triggerExit != 0 {
-				common, ok := capture(envGit(env)("-C", worktree, "rev-parse", "--path-format=absolute", "--git-common-dir"))
-				if !ok {
-					fmt.Fprintf(stderr, "%sgit cannot name the repository of %s — cannot answer\n", vvdPrefix, worktree)
-					return 2
-				}
-				repo := filepath.Base(filepath.Dir(strings.TrimSpace(common))) + "/"
-				prefixed := make([]string, len(paths))
-				for i, p := range paths {
-					if p != "" {
-						p = repo + p
-					}
-					prefixed[i] = p
-				}
-				switch visualTrigger(env, canonical, func() ([]string, error) { return prefixed, nil }, io.Discard, io.Discard) {
-				case 0:
-					triggerExit = 0
-				case 1:
-					triggerExit = 1
-				}
+		if c != worktree {
+			anchor, canonical = c, c
+		}
+	}
+	// ask is the trigger's answer for paths: 0, 1 or 2, or -1 when git
+	// cannot name the satellite's repository (reported, exit 2).
+	ask := func(paths []string) int {
+		exit := visualTrigger(env, worktree, func() ([]string, error) { return paths, nil }, io.Discard, io.Discard)
+		if canonical == "" || exit == 0 {
+			return exit
+		}
+		common, ok := capture(envGit(env)("-C", worktree, "rev-parse", "--path-format=absolute", "--git-common-dir"))
+		if !ok {
+			fmt.Fprintf(stderr, "%sgit cannot name the repository of %s — cannot answer\n", vvdPrefix, worktree)
+			return -1
+		}
+		repo := filepath.Base(filepath.Dir(strings.TrimSpace(common))) + "/"
+		prefixed := make([]string, len(paths))
+		for i, p := range paths {
+			if p != "" {
+				p = repo + p
 			}
+			prefixed[i] = p
 		}
+		switch visualTrigger(env, canonical, func() ([]string, error) { return prefixed, nil }, io.Discard, io.Discard) {
+		case 0:
+			return 0
+		case 1:
+			return 1
+		}
+		return exit
+	}
+	paths := strings.Split(changed, "\n")
+	triggerExit := ask(paths)
+	if triggerExit < 0 {
+		return 2
 	}
 	switch triggerExit {
 	case 2:
@@ -153,11 +164,55 @@ func checkVisualVerifyDispatched(args []string, env Env, stdout, stderr io.Write
 		fmt.Fprintf(stderr, "%sflow record dispatches failed for '%s' — cannot answer\n", vvdPrefix, name)
 		return 2
 	}
-	found, ok := vvdCompletedVerifier(rows)
+	heads, ok := vvdCompletedVerifiers(rows)
 	if !ok {
 		fmt.Fprintf(stderr, "%sdispatch rows were not readable JSON — cannot answer\n", vvdPrefix)
 		return 2
 	}
+	// A VERIFIER COUNTS ONLY FOR THE HEAD IT VERIFIED (kan-939 self-review):
+	// a completed row once answered OK however many UI commits landed after
+	// it. Each row's `diffBase` is that HEAD (visual-verify.md's
+	// **Recording**); the row counts when git resolves it in this worktree
+	// and none of this change's own touched paths that changed since is a
+	// UI path. Intersecting with the change's own paths keeps a base the
+	// branch was rebased onto from reading as this change's UI work. A row
+	// with no `diffBase`, or one this worktree cannot resolve -- another
+	// worktree's verifier -- verified nothing here.
+	own := map[string]bool{}
+	for _, p := range paths {
+		own[p] = true
+	}
+	found := false
+	for _, head := range heads {
+		if !vvdSHA.MatchString(head) {
+			continue
+		}
+		if _, ok := capture(envGit(env)("-C", worktree, "rev-parse", "--verify", "--quiet", head+"^{commit}")); !ok {
+			continue
+		}
+		since, ok := capture(envGit(env, "LC_ALL=C")("-C", worktree, "diff", "--no-renames", "--name-only", head, "HEAD"))
+		if !ok {
+			fmt.Fprintf(stderr, "%sgit diff failed against verified HEAD '%s' in %s — cannot answer\n", vvdPrefix, head, worktree)
+			return 2
+		}
+		var changedSince []string
+		for _, p := range strings.Split(since, "\n") {
+			if own[p] && p != "" {
+				changedSince = append(changedSince, p)
+			}
+		}
+		if len(changedSince) == 0 {
+			changedSince = []string{""}
+		}
+		exit := ask(changedSince)
+		if exit < 0 {
+			return 2
+		}
+		if exit != 0 {
+			found = true
+			break
+		}
+	}
 
 	// THE VERDICT IS RECORDED, AND PRIOR FALSE POSITIVES ARE ADVISORY -- the
 	// habit unfinishedwork.go carries. The write's result is discarded: a
@@ -202,6 +257,9 @@ func checkVisualVerifyDispatched(args []string, env Env, stdout, stderr io.Write
 		return 0
 	}
 	verdict := "VISUAL-VERIFY-MISSING: this change's diff touched a declared UI path but no completed or ended 'verifier' dispatch (key starting 'visual-verify') is recorded for '" + name + "' — flow.visual-verify's stage marks were written with no verifier ever dispatched, its report was never read to completion, or the dispatch's closing outcome was lost to a session restart"
+	if len(heads) > 0 {
+		verdict = "VISUAL-VERIFY-MISSING: a verifier dispatch is recorded for '" + name + "' but none verified a HEAD after this change's last UI-path commit — a row counts only with the -diff-base it verified, and only when no UI path this change touches changed since"
+	}
 	record(verdict)
 	var prior []byte
 	var priorErr error
@@ -351,10 +409,11 @@ func vvdLastLine(v json.RawMessage) (string, bool) {
 	return reason + " (" + change + ", " + string(flagged) + ")", true
 }
 
-// vvdCompletedVerifier is the bash's `[.[] | select((.role // "") ==
-// "verifier" and ((.key // "") | startswith("visual-verify")) and
-// ((.outcome // "") == "completed" or ((.endedAt != null) and ((.outcome //
-// "") == ""))))] | length` >= 1. A row counts when its role is `verifier`,
+// vvdCompletedVerifiers is the `diffBase` of every row the bash's `[.[] |
+// select((.role // "") == "verifier" and ((.key // "") |
+// startswith("visual-verify")) and ((.outcome // "") == "completed" or
+// ((.endedAt != null) and ((.outcome // "") == ""))))]` selects, "" for a
+// row that carries none. A row qualifies when its role is `verifier`,
 // its key starts with the stage's canonical prefix (the key's exact shape is
 // a different guard's job), and it carries one of two evidence shapes: its
 // outcome is exactly `completed`, or -- the session-continuation signature
@@ -373,23 +432,23 @@ func vvdLastLine(v json.RawMessage) (string, bool) {
 // verifier row whose key is neither a string nor null/false (jq failing, the
 // bash exiting 5 under set -e). A top-level object is refused as
 // pfdPanelFixKeys's comment explains.
-func vvdCompletedVerifier(raw []byte) (bool, bool) {
+func vvdCompletedVerifiers(raw []byte) ([]string, bool) {
 	vals, ok := pfdStream(raw)
 	if !ok || len(vals) != 1 {
-		return false, false
+		return nil, false
 	}
 	arr, ok := vals[0].([]any)
 	if !ok {
-		return false, false
+		return nil, false
 	}
-	found := false
+	var heads []string
 	for _, e := range arr {
 		if e == nil {
 			continue
 		}
 		row, ok := e.(map[string]any)
 		if !ok {
-			return false, false
+			return nil, false
 		}
 		if row["role"] != "verifier" {
 			continue
@@ -399,12 +458,12 @@ func vvdCompletedVerifier(raw []byte) (bool, bool) {
 		case nil:
 		case bool:
 			if k {
-				return false, false
+				return nil, false
 			}
 		case string:
 			key = k
 		default:
-			return false, false
+			return nil, false
 		}
 		// `// ""` reads null, false and absent as the empty string.
 		outcome := row["outcome"]
@@ -412,8 +471,14 @@ func vvdCompletedVerifier(raw []byte) (bool, bool) {
 			outcome = ""
 		}
 		if strings.HasPrefix(key, "visual-verify") && (outcome == "completed" || (row["endedAt"] != nil && outcome == "")) {
-			found = true
+			head, _ := row["diffBase"].(string)
+			heads = append(heads, head)
 		}
 	}
-	return found, true
+	return heads, true
 }
+
+// vvdSHA is a commit id as `git rev-parse HEAD` prints it, abbreviated or
+// whole -- never a ref name such as HEAD, which would verify whatever the
+// worktree holds when the guard runs.
+var vvdSHA = regexp.MustCompile(`^[0-9a-f]{7,40}$`)
````

  - [ ] **Step 4: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckVisualVerifyDispatched' -count=1` — `gofmt -l` prints nothing, the tests pass; `scripts/check-verbatim-moves.sh` (with this change's `verbatim-moves.txt`), `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-normative-inventory.sh` exit 0 (`check-installed-citations.sh` as in Task 1). Commit.

**Build:** green
**Files:** `stats/internal/guard/visualverifydispatched.go`, `stats/internal/guard/check_visual_verify_dispatched_test.go`, `scripts/check-visual-verify-dispatched.sh`, `skills/flow/visual-verify.md`
**Tests:** `case 9a: a verifier of a HEAD before the change's UI commit is stale, exit 1`, `case 9b: a verifier row with no diffBase verified nothing, exit 1`, `case 9c: a diffBase naming a ref, not a sha, verified nothing, exit 1`, `case 9d: a diffBase this worktree cannot resolve verified nothing here, exit 1`, `case 9e: a later verifier of HEAD beside a stale one is OK`
**Regression:** reverting the commit lets a verifier row recorded before the change's last UI-path commit — or with no recorded HEAD at all — pass the integrate gate.
**Baseline:** before=38 after=43
<!-- measured: grep -cE '^[[:space:]]*\{"case [0-9]' stats/internal/guard/check_visual_verify_dispatched_test.go @ a7196b56 (before; after with the earlier tasks' diffs and this task's applied in the worktree) -->
**Commit:** fix(flow-guard): a verifier dispatch counts only for the HEAD it verified
**After:** none

**Decision:** verifier-counts-for-its-head

- [ ] 5. A back-filled dispatch is one `flow record dispatch backfill` call with its real instants required
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `cd stats && go run ./cmd/flow record dispatch backfill 2>&1 | head -2` and confirm it is refused as an unknown record dispatch command, and `grep -n 'back-fill' skills/flow/implement.md` shows the real-instants rule only on close-task's RECORD path — the defect: back-filled rows take the recording moment as both instants.
  - [ ] **Step 2: Write the failing test.** Apply only the test hunk of the diff in step 3 (`git apply --include='stats/cmd/flow/record_test.go' <the diff>`) and run `cd stats && go test ./cmd/flow/ -run TestRecordDispatchBackfill -count=1`: `TestRecordDispatchBackfill` FAILS.
  - [ ] **Step 3: Implement.** Apply the rest of this diff (`git apply --exclude='stats/cmd/flow/record_test.go' <the diff>`):

````diff verified:applied with git apply to the worktree at a7196b56 in plan order; gofmt -l and go vet clean; go test ./cmd/flow/ and the close-task tests pass; TestRecordDispatchBackfill fails with record.go reverted; check-verbatim-moves.sh with this change's verbatim-moves.txt exits 0
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index f5177af9..d7d88161 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -374,6 +374,7 @@ unique within the run's session token — `task-<n>-implementer`, reused identic
 flat integer id, omitted for a dispatch against no single task. `-session-token` takes a literal,
 never a shell substitution. The start and end instants are the daemon's own — never
 caller inputs.
+**A dispatch recorded only after it finished is a back-fill, and is one call** — `flow record dispatch backfill`, carrying `begin`'s flags and `end`'s together with the dispatch's real `-started-at` and `-ended-at`, both required — never a `begin` and `end` pair, whose instants would read the back-fill's own moment.
 
 **`-model` is the model this dispatch was actually given — the group's own `model` from the
 decision's `groups` entry** (chosen per **Model and effort**, `skills/flow/brainstorm-planner.md`,
@@ -680,8 +681,8 @@ entry, one or more bundles `plan-dispatch-bundles.sh` emits. At each boundary, i
    is the parent's own fix, with nothing ticked or pushed: it re-commits that task itself,
    without the swept paths and carrying `Task-Id: <n>`, and re-runs `close-task.sh` without the `-end-*` flags before
    anything below; **exit 1 with `close-task: RECORD —`** is an appended task no dispatch row
-   names — record its implementer dispatch (`flow record dispatch begin … -task <n>` and its
-   `end`, a back-fill passing the dispatch's real `-started-at` and `-ended-at`), then re-run
+   names — record its implementer dispatch (`flow record dispatch backfill … -task <n>`, the
+   back-fill above), then re-run
    `close-task.sh` without the `-end-*` flags; **exit 2 — a guard's not-a-verdict close (an unreadable plan, a task it
    cannot resolve, a commit range git cannot resolve, a usage error), or a failed tick or push —
    stops the run**: re-committing cannot repair an inability, so it is never folded into
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index b75a6d66..a5a569a4 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -258,6 +258,12 @@ const recordUsage = `usage: flow record dispatch begin [-addr url] [-timeout dur
                              -change name -key key -session-token token
                              [-commit sha] [-outcome outcome] [-cause cause]
                              [-agent-id id] [-tokens k=v,...] [-ended-at instant]
+       flow record dispatch backfill [-addr url] [-timeout dur] [-C dir]
+                             -change name [-task id] -role role [-slot name]
+                             -model model [-agent-id id|none] [-diff-base sha]
+                             -started-at instant -ended-at instant
+                             -key key -session-token token [-commit sha]
+                             -outcome outcome [-cause cause] [-tokens k=v,...]
        flow record finding  [-addr url] [-timeout dur] [-C dir]
                              -change name -ref F<n> [-round n] -slot name
                              -severity sev [-location loc] -status status
@@ -483,6 +489,12 @@ invisible to every harvest tick the dispatch ran through -- that usage is
 dropped or credited to an unrelated earlier dispatch -- and a row never
 closed leaves an open window that goes on claiming its successors' tokens.
 
+A dispatch recorded only after it finished is a back-fill, and is ONE
+"backfill" call: begin's flags and end's together, with -started-at,
+-ended-at and -outcome required, so the row carries the dispatch's real
+instants rather than the moment it was recorded. It validates both
+instants before writing either half, then runs begin and end.
+
 -key names the dispatch within its run. It is what "end" closes, and what
 makes "begin" idempotent: a write whose response was lost is journalled
 and replayed carrying the identical key, so it updates the row the first
@@ -886,6 +898,8 @@ func runRecordDispatchVerb(ctx context.Context, args []string, stdout, stderr io
 		return runRecordDispatchBegin(ctx, args[1:], stdout, stderr)
 	case "end":
 		return runRecordDispatchEnd(ctx, args[1:], stdout, stderr)
+	case "backfill":
+		return runRecordDispatchBackfill(ctx, args[1:], stdout, stderr)
 	default:
 		fmt.Fprintf(stderr, "flow: unknown record dispatch command %q\n", args[0])
 		fmt.Fprint(stderr, recordUsage)
@@ -1162,6 +1176,58 @@ func runRecordDispatchEnd(ctx context.Context, args []string, stdout, stderr io.
 	return classifyDispatchEnd(callErr, projectKey, f.change, in, stderr)
 }
 
+// runRecordDispatchBackfill implements `flow record dispatch backfill`: a
+// row recorded after its dispatch ran, in one call whose real instants are
+// required (kan-939 self-review: back-filled begin/end pairs took the
+// recording moment as both instants, since only close-task's message asked
+// for -started-at and -ended-at). It routes each flag it was given to begin,
+// end or both and runs the two, begin first; both instants are validated
+// before either is sent, so a bad -ended-at never leaves a half-written row.
+func runRecordDispatchBackfill(ctx context.Context, args []string, stdout, stderr io.Writer) int {
+	fset := flag.NewFlagSet("flow record dispatch backfill", flag.ContinueOnError)
+	fset.SetOutput(stderr)
+	var f recordIdentityFlags
+	registerRecordIdentityFlags(fset, &f)
+	beginOnly := []string{"task", "role", "slot", "model", "effort", "agent-id", "diff-base", "started-at"}
+	endOnly := []string{"commit", "outcome", "cause", "tokens", "ended-at"}
+	v := map[string]*string{}
+	for _, n := range append(append([]string{"key", "session-token"}, beginOnly...), endOnly...) {
+		v[n] = fset.String(n, "", "as on `dispatch begin` or `dispatch end`")
+	}
+	if err := fset.Parse(args); err != nil || fset.NArg() != 0 {
+		return 2
+	}
+	if !requireRecordFlags(stderr,
+		[2]string{"-started-at", *v["started-at"]},
+		[2]string{"-ended-at", *v["ended-at"]},
+		[2]string{"-outcome", *v["outcome"]},
+	) {
+		return 2
+	}
+	if _, ok := parseBackfillInstant("-started-at", *v["started-at"], stderr); !ok {
+		return 2
+	}
+	if _, ok := parseBackfillInstant("-ended-at", *v["ended-at"], stderr); !ok {
+		return 2
+	}
+	var begin, end []string
+	fset.Visit(func(fl *flag.Flag) {
+		arg := []string{"-" + fl.Name, fl.Value.String()}
+		switch {
+		case slices.Contains(beginOnly, fl.Name):
+			begin = append(begin, arg...)
+		case slices.Contains(endOnly, fl.Name):
+			end = append(end, arg...)
+		default:
+			begin, end = append(begin, arg...), append(end, arg...)
+		}
+	})
+	if code := runRecordDispatchBegin(ctx, begin, stdout, stderr); code != 0 {
+		return code
+	}
+	return runRecordDispatchEnd(ctx, end, stdout, stderr)
+}
+
 // parseBackfillInstant reads -started-at or -ended-at: empty is the zero
 // instant, which the daemon stamps at its own clock (KAN-324); anything
 // else must be RFC 3339. The flags exist for a back-fill -- a row recorded
diff --git a/stats/cmd/flow/record_test.go b/stats/cmd/flow/record_test.go
index f2b4c3b5..17411996 100644
--- a/stats/cmd/flow/record_test.go
+++ b/stats/cmd/flow/record_test.go
@@ -4871,6 +4871,57 @@ func TestRequireChangeCheckoutRefusesNonChangeNames(t *testing.T) {
 	}
 }
 
+// TestRecordDispatchBackfill pins the kan-939 self-review fix: a back-fill
+// is one `dispatch backfill` call whose real instants are required. It
+// sends begin carrying -started-at, then end carrying -ended-at and the
+// outcome; a call missing either instant, or carrying one that is not RFC
+// 3339, is exit 2 before the store is called -- never half a row.
+func TestRecordDispatchBackfill(t *testing.T) {
+	const start, end = "2026-10-15T12:08:50Z", "2026-10-15T12:31:07Z"
+	repo := recordRepo(t, "kan-258")
+	isolatedStateRoot(t)
+	var bodies []map[string]any
+	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
+		raw, _ := readAll(r)
+		var b map[string]any
+		_ = json.Unmarshal(raw, &b)
+		bodies = append(bodies, b)
+		w.WriteHeader(http.StatusOK)
+		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"implementer","model":"opus"}`))
+	}))
+	defer srv.Close()
+	base := []string{"record", "dispatch", "backfill", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
+		"-change", "kan-258", "-task", "7", "-role", "implementer", "-model", "opus", "-agent-id", "none",
+		"-key", "task-7-implementer", "-session-token", "mf-record-backfill", "-commit", "abc1234"}
+	var stdout, stderr bytes.Buffer
+	args := append(append([]string{}, base...), "-started-at", start, "-ended-at", end, "-outcome", "completed")
+	if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 0 {
+		t.Fatalf("exit %d, want 0; stderr:\n%s", code, stderr.String())
+	}
+	if len(bodies) != 2 {
+		t.Fatalf("store calls = %d, want begin then end: %v", len(bodies), bodies)
+	}
+	if bodies[0]["startedAt"] != start || bodies[0]["taskId"] != "7" || bodies[0]["endedAt"] != nil {
+		t.Errorf("begin body = %v, want startedAt %s, taskId 7, no endedAt", bodies[0], start)
+	}
+	if bodies[1]["endedAt"] != end || bodies[1]["outcome"] != "completed" || bodies[1]["commitSha"] != "abc1234" {
+		t.Errorf("end body = %v, want endedAt %s, outcome completed, commitSha abc1234", bodies[1], end)
+	}
+
+	for name, extra := range map[string][]string{
+		"no -ended-at":          {"-started-at", start, "-outcome", "completed"},
+		"no -started-at":        {"-ended-at", end, "-outcome", "completed"},
+		"no -outcome":           {"-started-at", start, "-ended-at", end},
+		"a malformed -ended-at": {"-started-at", start, "-ended-at", "yesterday", "-outcome", "completed"},
+	} {
+		bodies = nil
+		stderr.Reset()
+		if code := run(context.Background(), append(append([]string{}, base...), extra...), strings.NewReader(""), &stdout, &stderr); code != 2 || bodies != nil {
+			t.Errorf("%s: exit %d, store calls %d; want 2 and none (stderr %q)", name, code, len(bodies), stderr.String())
+		}
+	}
+}
+
 // TestRecordDispatchBackfillInstants pins the kan-982 self-review fix: a
 // back-filled row carries the dispatch's real instants. -started-at reaches
 // begin's body as startedAt and -ended-at end's as endedAt; an instant that
diff --git a/stats/internal/guard/closetask.go b/stats/internal/guard/closetask.go
index 27e4ccdf..61f0f8b9 100644
--- a/stats/internal/guard/closetask.go
+++ b/stats/internal/guard/closetask.go
@@ -93,7 +93,7 @@ func closeTask(args []string, env Env, stdout, stderr io.Writer) int {
 	}
 	undispatched := ctUndispatched(env, canonical, name, closed)
 	for _, id := range undispatched {
-		fmt.Fprintf(stderr, "close-task: appended task %s has no dispatch row naming it — record its dispatch (`flow record dispatch begin … -task %s` and its `end`; a back-fill passes -started-at and -ended-at), then re-run without the -end-* flags\n", id, id)
+		fmt.Fprintf(stderr, "close-task: appended task %s has no dispatch row naming it — record its dispatch (`flow record dispatch backfill … -task %s -started-at <instant> -ended-at <instant> -outcome <outcome>`), then re-run without the -end-* flags\n", id, id)
 	}
 
 	// Both commit guards run before anything is ticked or pushed, so a
````

  - [ ] **Step 4: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./cmd/flow/ ./internal/guard/ && go test ./cmd/flow/ -run 'TestRecordDispatchBackfill|TestRecordMissingRequiredFlag' -count=1 && go test ./internal/guard/ -run TestCloseTask -count=1` — `gofmt -l` prints nothing, the tests pass; `scripts/check-verbatim-moves.sh` (with this change's `verbatim-moves.txt`), `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-normative-inventory.sh` exit 0 (`check-installed-citations.sh` as in Task 1). Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/internal/guard/closetask.go`, `skills/flow/implement.md`
**Tests:** `TestRecordDispatchBackfill`
**Regression:** reverting the commit makes `flow record dispatch backfill` an unknown command, so a back-fill goes back to a `begin`/`end` pair whose instants read the recording moment.
**Baseline:** before=104 after=105
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ a7196b56 (before; after with the earlier tasks' diffs and this task's applied in the worktree) -->
**Commit:** feat(flow-cli): a back-filled dispatch is one call carrying its real instants
**After:** Task 2

**Decision:** backfill-is-one-call

- [ ] 6. The unfinished-work gate reports a panel round dispatched with no pass row
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `cd stats && go run ./cmd/flow record passes 2>&1 | head -2` and confirm it is refused as an unknown record command, and `grep -n 'pass rows' internal/guard/unfinishedwork.go` shows "pass rows have no read verb" — the defect: a dispatched panel round that recorded no pass row reaches integrate unflagged.
  - [ ] **Step 2: Write the failing tests.** Apply only the test hunks of the diff in step 3 (`git apply --include='stats/cmd/flow/record_test.go' --include='stats/internal/guard/check_unfinished_work_test.go' <the diff>`) and run `cd stats && go test ./cmd/flow/ -run TestRunRecordPasses -count=1` and `go test ./internal/guard/ -run TestCheckUnfinishedWork -count=1`: both FAIL.
  - [ ] **Step 3: Implement.** Apply the rest of this diff (`git apply --exclude='stats/cmd/flow/record_test.go' --exclude='stats/internal/guard/check_unfinished_work_test.go' <the diff>`):

````diff verified:applied with git apply to the worktree at a7196b56 in plan order; gofmt -l and go vet clean; go test ./cmd/flow/ ./internal/guard/ ./internal/records/ pass; the new signal-five subtests fail with unfinishedwork.go reverted; scripts/test-integrate-change.sh 101 passed
diff --git a/scripts/check-unfinished-work.sh b/scripts/check-unfinished-work.sh
index 580606b4..5cbf97d0 100755
--- a/scripts/check-unfinished-work.sh
+++ b/scripts/check-unfinished-work.sh
@@ -71,6 +71,15 @@
 # plan with a ticked appended task and no range passed is exit 2, and a range
 # git cannot read is exit 2.
 #
+# SIGNAL FIVE: A PANEL ROUND DISPATCHED WITH NO PASS ROW IS UNFINISHED
+# (kan-939 self-review). Every round a reviewer row keyed
+# `panel-<round>-<slot>` ran in needs a pass row of that round, read with
+# `flow record passes -change <name> -C <anchor>` — every panel run and
+# fix-round pass records one (skills/flow/review-panel.md). The pass rows are
+# read only when such a round exists; a failed or unreadable read is exit 2,
+# like the dispatch read. Like signal three, only the worktree holding the
+# plan reads them.
+#
 # All signals are counted independently and reported together on the one
 # line, so the operator sees the whole picture in one prompt rather than being
 # sent back around the loop one signal at a time.
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index a5a569a4..37471c7e 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -281,6 +281,8 @@ const recordUsage = `usage: flow record dispatch begin [-addr url] [-timeout dur
                              -name pattern
        flow record dispatches [-addr url] [-timeout dur] [-C dir]
                              -change name
+       flow record passes   [-addr url] [-timeout dur] [-C dir]
+                             -change name
        flow record next-round [-addr url] [-timeout dur] [-C dir]
                              -change name
        flow record next-ref [-addr url] [-timeout dur] [-C dir]
@@ -587,6 +589,8 @@ func runRecord(ctx context.Context, args []string, stdin io.Reader, stdout, stde
 		return runRecordFindingPattern(ctx, args[1:], stdout, stderr)
 	case "dispatches":
 		return runRecordDispatches(ctx, args[1:], stdout, stderr)
+	case "passes":
+		return runRecordPasses(ctx, args[1:], stdout, stderr)
 	case "next-round":
 		return runRecordNextRound(ctx, args[1:], stdout, stderr)
 	case "next-ref":
@@ -2075,7 +2079,32 @@ func runRecordFindingPattern(ctx context.Context, args []string, stdout, stderr
 // never rendered as an empty array a caller could mistake for "no
 // dispatches."
 func runRecordDispatches(ctx context.Context, args []string, stdout, stderr io.Writer) int {
-	fset := flag.NewFlagSet("flow record dispatches", flag.ContinueOnError)
+	return runRecordRunRows(ctx, "dispatches", args, stdout, stderr, func(r records.Run) any {
+		if r.Dispatches == nil {
+			return []records.Dispatch{}
+		}
+		return r.Dispatches
+	})
+}
+
+// runRecordPasses implements `flow record passes`: the change's pass-log
+// rows as one JSON array, `dispatches`' twin -- the read
+// check-unfinished-work needs to tell a panel round that recorded no pass
+// row (kan-939 self-review).
+func runRecordPasses(ctx context.Context, args []string, stdout, stderr io.Writer) int {
+	return runRecordRunRows(ctx, "passes", args, stdout, stderr, func(r records.Run) any {
+		if r.Passes == nil {
+			return []records.Pass{}
+		}
+		return r.Passes
+	})
+}
+
+// runRecordRunRows is a read verb printing one kind of the change's run
+// record rows as a JSON array: a change the store has never heard of is
+// `[]`, any other failed read exits 1 with nothing on stdout.
+func runRecordRunRows(ctx context.Context, verb string, args []string, stdout, stderr io.Writer, rows func(records.Run) any) int {
+	fset := flag.NewFlagSet("flow record "+verb, flag.ContinueOnError)
 	fset.SetOutput(stderr)
 	var f recordIdentityFlags
 	registerRecordIdentityFlags(fset, &f)
@@ -2099,19 +2128,15 @@ func runRecordDispatches(ctx context.Context, args []string, stdout, stderr io.W
 		// The store was reached and has never heard of this change --
 		// "no rows" is a fact, not a failure, and prints as an empty
 		// array below exactly as the findings verb treats it.
-		run = records.Run{Change: f.change, Dispatches: []records.Dispatch{}}
+		run = records.Run{Change: f.change}
 	default:
-		fmt.Fprintf(stderr, "flow: dispatches: %v\n", callErr)
+		fmt.Fprintf(stderr, "flow: %s: %v\n", verb, callErr)
 		return 1
 	}
 
-	if run.Dispatches == nil {
-		run.Dispatches = []records.Dispatch{}
-	}
-
-	body, err := json.Marshal(run.Dispatches)
+	body, err := json.Marshal(rows(run))
 	if err != nil {
-		fmt.Fprintf(stderr, "flow: encode dispatches: %v\n", err)
+		fmt.Fprintf(stderr, "flow: encode %s: %v\n", verb, err)
 		return 1
 	}
 	fmt.Fprintln(stdout, string(body))
diff --git a/stats/cmd/flow/record_test.go b/stats/cmd/flow/record_test.go
index 17411996..e7171929 100644
--- a/stats/cmd/flow/record_test.go
+++ b/stats/cmd/flow/record_test.go
@@ -4657,6 +4657,48 @@ func TestRunRecordNextRound(t *testing.T) {
 	})
 }
 
+// TestRunRecordPasses pins `flow record passes`: the change's pass rows as
+// one JSON array, `[]` for a change the store has never heard of, and
+// nothing on stdout with a non-zero exit when the store cannot be read --
+// check-unfinished-work's signal five reads it.
+func TestRunRecordPasses(t *testing.T) {
+	repo := recordRepo(t, "demo")
+	isolatedStateRoot(t)
+	for _, tc := range []struct {
+		name, body string
+		status     int
+		want       string
+	}{
+		{"pass rows", `{"change":"demo","dispatches":[],"passes":[{"id":1,"round":0,"note":"roster: full"}]}`, http.StatusOK, `[{"id":1,"round":0,"note":"roster: full"}]`},
+		{"no pass rows", `{"change":"demo","dispatches":[]}`, http.StatusOK, `[]`},
+		{"unknown change", ``, http.StatusNotFound, `[]`},
+	} {
+		t.Run(tc.name, func(t *testing.T) {
+			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
+				w.WriteHeader(tc.status)
+				_, _ = w.Write([]byte(tc.body))
+			}))
+			defer srv.Close()
+			var stdout, stderr bytes.Buffer
+			code := run(context.Background(),
+				[]string{"record", "passes", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "demo"},
+				strings.NewReader(""), &stdout, &stderr)
+			if code != 0 || strings.TrimSpace(stdout.String()) != tc.want {
+				t.Fatalf("exit %d stdout %q, want 0 and %s; stderr:\n%s", code, stdout.String(), tc.want, stderr.String())
+			}
+		})
+	}
+	t.Run("unreachable store prints nothing", func(t *testing.T) {
+		var stdout, stderr bytes.Buffer
+		code := run(context.Background(),
+			[]string{"record", "passes", "-addr", deadPortAddr(t), "-timeout", "500ms", "-C", repo, "-change", "demo"},
+			strings.NewReader(""), &stdout, &stderr)
+		if code == 0 || stdout.Len() != 0 {
+			t.Fatalf("exit %d stdout %q, want non-zero and empty", code, stdout.String())
+		}
+	})
+}
+
 // TestRunRecordNextRef pins the ref a run records its first finding under:
 // one past the highest F<n> any finding row carries, compared as numbers,
 // F1 for a change the store has never heard of, and nothing on a failed
diff --git a/stats/internal/guard/check_unfinished_work_test.go b/stats/internal/guard/check_unfinished_work_test.go
index 50a45071..3d8a155c 100644
--- a/stats/internal/guard/check_unfinished_work_test.go
+++ b/stats/internal/guard/check_unfinished_work_test.go
@@ -69,6 +69,8 @@ func newUWBare(t *testing.T, wt string) *uwFix {
 	writeExec(t, f.bin+"/flow", uwStub)
 	f.setFile(t, "verdicts.json", "[]\n")
 	f.setFile(t, "dispatches.json", `[{"key":"task-1-implementer","outcome":"completed"}]`+"\n")
+	// A pass row for every round a fixture's panel rows run in (0..2).
+	f.setFile(t, "passes.json", `[{"id":1,"round":0,"note":"roster: full"},{"id":2,"round":1,"note":"n"},{"id":3,"round":2,"note":"n"}]`+"\n")
 	f.setFindings(t)
 	get := pathEnv(f.bin)
 	f.env = Env{Getenv: get, Dir: wt,
@@ -329,6 +331,28 @@ func TestCheckUnfinishedWork(t *testing.T) {
 		uwVerdict("a non-panel reviewer row does not review the appended task", "OUTSTANDING:"))
 	each(demo(planned("- [x] 1. done\n- [x] 2. grouped\n- [x] 3. grouped\n", `[{"key":"task-2-implementer","outcome":"completed"}]`)),
 		uwVerdict("plan-time tasks grouped under one key, with no Tasks appended line, are CLEAR", "CLEAR:"))
+	// 2e. Signal five — a panel round dispatched with no pass row (kan-939:
+	//     round 1 ran panel-1-primary+principles and recorded nothing).
+	withPasses := func(rows, passes string) func(t *testing.T, f *uwFix) {
+		return func(t *testing.T, f *uwFix) {
+			f.setFile(t, "dispatches.json", rows+"\n")
+			f.setFile(t, "passes.json", passes+"\n")
+		}
+	}
+	panelRows := `[{"seq":1,"key":"panel-0-primary+principles","role":"reviewer","outcome":"clean"},{"seq":2,"key":"panel-fix-1","role":"panel-fix","outcome":"completed"},{"seq":3,"key":"panel-1-primary+principles","role":"reviewer","outcome":"clean"},{"seq":4,"key":"task-1-reviewer","role":"reviewer","outcome":"clean"}]`
+	each(demo(withPasses(panelRows, `[{"id":1,"round":0,"note":"roster: full"}]`)),
+		uwVerdict("a panel round dispatched with no pass row is OUTSTANDING", "OUTSTANDING:"),
+		uwReason("the breakdown names the round and its keys", "1 panel round(s) dispatched with no pass row (round 1: panel-1-primary+principles)"))
+	each(demo(withPasses(panelRows, `[{"id":1,"round":0,"note":"roster: full"},{"id":2,"round":1,"note":"n"}]`)),
+		uwVerdict("every dispatched panel round with a pass row is CLEAR", "CLEAR:"))
+	each(demo(withPasses(`[{"seq":1,"key":"task-1-reviewer","role":"reviewer","outcome":"clean"},{"seq":2,"key":"panel-fix-3","role":"panel-fix","outcome":"completed"}]`, "not json")),
+		uwVerdict("no panel reviewer row reads no pass rows", "CLEAR:"))
+	each(demo(withPasses(panelRows, "not json")), uwRC("unreadable pass rows cannot answer", 2))
+	each(demo(func(t *testing.T, f *uwFix) {
+		withPasses(panelRows, "[]")(t, f)
+		f.setFile(t, "passes.fail", "")
+	}), uwRC("a failed pass-row read cannot answer", 2))
+
 	// 2d. Signal four's commit half — a dispatch row is not a commit (a
 	//     back-filled row named a task whose work never landed).
 	each(demo(committedAs("base", "fix: the appended task\n\nFlow-Key: panel-fix-2", false)),
diff --git a/stats/internal/guard/unfinishedwork.go b/stats/internal/guard/unfinishedwork.go
index 9b406eda..b1b0185b 100644
--- a/stats/internal/guard/unfinishedwork.go
+++ b/stats/internal/guard/unfinishedwork.go
@@ -7,9 +7,12 @@ import (
 	"io"
 	"os"
 	"regexp"
+	"slices"
 	"strconv"
 	"strings"
 	"unicode/utf8"
+
+	"github.com/tweety53/agents/stats/internal/records"
 )
 
 // checkUnfinishedWork is scripts/check-unfinished-work.sh: that script's
@@ -311,11 +314,52 @@ func checkUnfinishedWork(args []string, env Env, stdout, stderr io.Writer) int {
 		if len(unclosed) > 0 {
 			reasons = append(reasons, strconv.Itoa(len(unclosed))+" dispatch row(s) with no outcome ("+strings.Join(unclosed, ", ")+") — close each with "+strings.Join(closeCmds, ", "))
 		}
+		// SIGNAL FIVE — a panel round dispatched with no pass row (kan-939
+		// self-review: round 1 ran panel-1-primary+principles and recorded
+		// nothing, and the gate cleared it). Every panel run and fix-round
+		// pass records pass rows (review-panel.md), so a round some
+		// `panel-<round>-…` reviewer row ran in and no pass row carries is
+		// unfinished. Pass rows are read only when such a round exists, and
+		// a failed read is refused like the dispatch read above.
+		rounds := map[int][]string{}
+		for _, d := range dispatches {
+			if n, ok := records.PanelDispatchRound(records.Dispatch{Key: d.Key, Role: d.Role}); ok {
+				rounds[n] = append(rounds[n], d.Key)
+			}
+		}
+		if len(rounds) > 0 {
+			passOut, passErr, passRC := pcFlow(env, false, "record", "passes", "-change", name, "-C", anchor)
+			if passRC != 0 {
+				fmt.Fprintf(stderr, "%scannot read pass rows for '%s' from the store — cannot determine anything: %s\n", uwPrefix, name, strings.TrimRight(string(passErr), "\n"))
+				return 2
+			}
+			var passes []struct {
+				Round int `json:"round"`
+			}
+			if json.Unmarshal(passOut, &passes) != nil || passes == nil {
+				fmt.Fprintf(stderr, "%spass rows for '%s' are not one JSON array — cannot determine anything\n", uwPrefix, name)
+				return 2
+			}
+			for _, p := range passes {
+				delete(rounds, p.Round)
+			}
+			if len(rounds) > 0 {
+				var nums []int
+				for n := range rounds {
+					nums = append(nums, n)
+				}
+				slices.Sort(nums)
+				var named []string
+				for _, n := range nums {
+					named = append(named, "round "+strconv.Itoa(n)+": "+strings.Join(rounds[n], ", "))
+				}
+				reasons = append(reasons, strconv.Itoa(len(nums))+" panel round(s) dispatched with no pass row ("+strings.Join(named, "; ")+") — record each round's pass rows with `flow record pass -change "+name+" -round <round>` before landing")
+			}
+		}
 		// SIGNAL FOUR — a ticked appended task no dispatch row names: a fix
 		// round that recorded nothing (kan-979's fix-round Task 7 reached
 		// integrate with no dispatch row at all). A row names a task by its
-		// taskId or by the `+`-joined ids of a task-<ids>[-…] key; pass rows
-		// have no read verb, so dispatch rows are the proxy. Only the tasks a
+		// taskId or by the `+`-joined ids of a task-<ids>[-…] key. Only the tasks a
 		// fix run appended — the last `**Tasks appended:** <n>` of the plan —
 		// are judged: each is implemented alone, so it carries a row of its
 		// own, while plan-time tasks run grouped under one task-<n> key or
diff --git a/stats/internal/records/render.go b/stats/internal/records/render.go
index e92d04e1..ebf1bd8d 100644
--- a/stats/internal/records/render.go
+++ b/stats/internal/records/render.go
@@ -218,7 +218,7 @@ func RenderPanel(r Run) string {
 			if passes == 0 {
 				var keys []string
 				for _, d := range r.Dispatches {
-					if n, ok := panelDispatchRound(d); ok && n == round {
+					if n, ok := PanelDispatchRound(d); ok && n == round {
 						keys = append(keys, neutraliseMarkers(d.Key))
 					}
 				}
@@ -269,7 +269,7 @@ func passLogRounds(r Run) []int {
 		add(m.Round)
 	}
 	for _, d := range r.Dispatches {
-		if n, ok := panelDispatchRound(d); ok {
+		if n, ok := PanelDispatchRound(d); ok {
 			add(n)
 		}
 	}
@@ -277,10 +277,11 @@ func passLogRounds(r Run) []int {
 	return nums
 }
 
-// panelDispatchRound reads the round out of a panel bundle's dispatch key,
+// PanelDispatchRound reads the round out of a panel bundle's dispatch key,
 // panel-<round>-<slot|slot+slot> (skills/flow/implement.md's dispatch
 // table); any other row -- a panel-fix-<round> fixer included -- is not one.
-func panelDispatchRound(d Dispatch) (int, bool) {
+// check-unfinished-work's panel-round signal reads rounds the same way.
+func PanelDispatchRound(d Dispatch) (int, bool) {
 	if d.Role != "reviewer" {
 		return 0, false
 	}
````

  - [ ] **Step 4: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./cmd/flow/ ./internal/guard/ ./internal/records/ && go test ./cmd/flow/ -run 'TestRunRecordPasses|Dispatches' -count=1 && go test ./internal/guard/ -run TestCheckUnfinishedWork -count=1 && go test ./internal/records/ -count=1` — `gofmt -l` prints nothing, the tests pass; `bash scripts/test-integrate-change.sh 2>&1 | tail -1` reads `101 passed, 0 failed`; `scripts/check-references.sh` exits 0. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `stats/internal/records/render.go`, `scripts/check-unfinished-work.sh`
**Tests:** `TestRunRecordPasses`, `a panel round dispatched with no pass row is OUTSTANDING`, `the breakdown names the round and its keys`, `every dispatched panel round with a pass row is CLEAR`, `no panel reviewer row reads no pass rows`, `unreadable pass rows cannot answer`, `a failed pass-row read cannot answer`
**Regression:** reverting the commit removes `flow record passes` and lets a panel round that ran and recorded no pass row pass the integrate gate.
**Baseline:** before=192 after=199
<!-- measured: cat stats/cmd/flow/record_test.go stats/internal/guard/check_unfinished_work_test.go | grep -cE '^func Test|(uwVerdict|uwRC|uwReason)\("' @ a7196b56 (before; after with the earlier tasks' diffs and this task's applied in the worktree) -->
**Commit:** fix(flow-guard): the unfinished-work gate reports a panel round dispatched with no pass row
**After:** Task 3, 5

**Decision:** passes-read-verb

- [ ] 7. gymie-frontend's full suite runs the desktop UI tests
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/gymie-frontend-worktrees/kan-939-step-4-frontend-cardio-plan-view-start-log`, run `grep -n '^## test' .flow/project.md` and confirm it prints nothing — the defect: with no `## test` the pipeline's full suite is auto-detected as Gradle's `test` task alone, which does not depend on `:shared:desktopTest` (design.md's Context records the dry runs).
  - [ ] **Step 2: Apply the fix.** From `/Users/tweety53/Projects/gymie-frontend-worktrees/kan-939-step-4-frontend-cardio-plan-view-start-log`, apply this diff:

````diff verified:git apply --check in the gymie-frontend worktree at 7eb56b16; ./gradlew test :shared:desktopTest --dry-run lists :shared:desktopTest there; scripts/project-get.sh reads the section back
--- a/.flow/project.md
+++ b/.flow/project.md
@@ -28,6 +28,16 @@
 `git worktree list --porcelain`. Where the main checkout has none either — a machine relying on
 `ANDROID_HOME` — it copies nothing and exits 0.
 
+## test
+
+```bash
+./gradlew test :shared:desktopTest   # full suite — `test` alone never runs the desktop UI tests
+```
+
+`:shared:desktopTest` is named on its own because `./gradlew test` does not depend on it
+(`./gradlew test --dry-run` lists no `desktopTest` task), so the full suite would otherwise skip
+every Compose UI test under `shared/src/desktopTest`.
+
 ## self review
 
 defer
````

  - [ ] **Step 3: Verify.** From this agents worktree, `scripts/project-get.sh /Users/tweety53/Projects/gymie-frontend-worktrees/kan-939-step-4-frontend-cardio-plan-view-start-log test` prints the section's fenced command line, the one naming `:shared:desktopTest`. Commit in the gymie-frontend worktree: `git -C /Users/tweety53/Projects/gymie-frontend-worktrees/kan-939-step-4-frontend-cardio-plan-view-start-log commit -m "chore(flow-project): the full suite runs the desktop UI tests" -- .flow/project.md`.

**Build:** green
**Files:** `.flow/project.md` (gymie-frontend repository, worktree `/Users/tweety53/Projects/gymie-frontend-worktrees/kan-939-step-4-frontend-cardio-plan-view-start-log`)
**Tests:** **none** — a configuration section; step 3's checks cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** chore(flow-project): the full suite runs the desktop UI tests
**After:** none

**Decision:** frontend-test-names-desktop

- [ ] 8. A full Claude model id is recorded under its alias
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n 'Model:        \*model' stats/cmd/flow/record.go` and confirm `begin` sends `-model` as given — the defect: a parent's inline row records `claude-opus-5-5` beside dispatches recorded as `opus`.
  - [ ] **Step 2: Write the failing test.** Apply only the test hunk of the diff in step 3 (`git apply --include='stats/cmd/flow/record_test.go' <the diff>`) and run `cd stats && go test ./cmd/flow/ -run TestRecordDispatchBeginModelAlias -count=1`: it FAILS.
  - [ ] **Step 3: Implement.** Apply the rest of this diff (`git apply --exclude='stats/cmd/flow/record_test.go' <the diff>`):

````diff verified:applied with git apply to the worktree at a7196b56 in plan order; gofmt -l and go vet clean; go test ./cmd/flow/ passes; TestRecordDispatchBeginModelAlias fails with record.go reverted
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index 37471c7e..f30ffefa 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -1006,6 +1006,7 @@ func runRecordDispatchBegin(ctx context.Context, args []string, stdout, stderr i
 	if *agentID == "none" {
 		*agentID = ""
 	}
+	*model = recordModelAlias(*model)
 	if !slices.Contains(recordRoles, *role) {
 		fmt.Fprintf(stderr, "flow: -role %q is not one of: %s\n", *role, strings.Join(recordRoles, ", "))
 		return 2
@@ -1180,6 +1181,19 @@ func runRecordDispatchEnd(ctx context.Context, args []string, stdout, stderr io.
 	return classifyDispatchEnd(callErr, projectKey, f.change, in, stderr)
 }
 
+// recordModelAlias maps a full Claude model id to the alias dispatches are
+// recorded under (kan-939 self-review: a parent's inline row recorded
+// claude-opus-5-5 while every dispatch it made recorded opus, so one model
+// read as two). Every other value is kept verbatim.
+func recordModelAlias(model string) string {
+	for _, alias := range []string{"opus", "sonnet"} {
+		if strings.HasPrefix(model, "claude-"+alias+"-") {
+			return alias
+		}
+	}
+	return model
+}
+
 // runRecordDispatchBackfill implements `flow record dispatch backfill`: a
 // row recorded after its dispatch ran, in one call whose real instants are
 // required (kan-939 self-review: back-filled begin/end pairs took the
diff --git a/stats/cmd/flow/record_test.go b/stats/cmd/flow/record_test.go
index e7171929..51d79479 100644
--- a/stats/cmd/flow/record_test.go
+++ b/stats/cmd/flow/record_test.go
@@ -4913,6 +4913,22 @@ func TestRequireChangeCheckoutRefusesNonChangeNames(t *testing.T) {
 	}
 }
 
+// TestRecordDispatchBeginModelAlias pins that a full Claude model id is
+// recorded under its alias -- kan-939's parent rows read claude-opus-5-5
+// beside dispatches' opus -- and any other model verbatim.
+func TestRecordDispatchBeginModelAlias(t *testing.T) {
+	for given, want := range map[string]string{
+		"claude-opus-5-5":   "opus",
+		"claude-sonnet-4-6": "sonnet",
+		"opus":              "opus",
+		"glm-5.3-flash":     "glm-5.3-flash",
+	} {
+		if sent := dispatchBeginBody(t, "-model", given); sent["model"] != want {
+			t.Errorf("-model %s recorded as %v, want %s", given, sent["model"], want)
+		}
+	}
+}
+
 // TestRecordDispatchBackfill pins the kan-939 self-review fix: a back-fill
 // is one `dispatch backfill` call whose real instants are required. It
 // sends begin carrying -started-at, then end carrying -ended-at and the
````

  - [ ] **Step 4: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./cmd/flow/ && go test ./cmd/flow/ -run 'TestRecordDispatchBeginModelAlias|TestRecordDispatchBackfill' -count=1` — `gofmt -l` prints nothing, the tests pass. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchBeginModelAlias`
**Regression:** reverting the commit records `claude-opus-*` and `claude-sonnet-*` verbatim again, splitting one model across two values in the store.
**Baseline:** before=106 after=107
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ a7196b56 (before; after with the earlier tasks' diffs and this task's applied in the worktree) -->
**Commit:** fix(flow-cli): a full Claude model id is recorded under its alias
**After:** Task 5, 6

**Decision:** model-alias-on-record

- [ ] 9. verify-and-handoff's guard citation names its root
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `scripts/check-installed-citations.sh 2>&1 | grep 'names no root'` and confirm it reports `skills/flow/verify-and-handoff.md:207` citing `scripts/check-visual-verify-dispatched.sh` — the defect: `origin/main` fails this guard.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:applied with git apply to the worktree at a7196b56 in plan order; check-installed-citations.sh reports INSTALLED-CITATIONS-OK; check-verbatim-moves.sh with this change's verbatim-moves.txt exits 0
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index 0f4de81d..492fc6fd 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -204,7 +204,7 @@ this pipeline restates it. Resolve once per worktree in this run's resolved set,
    does.
 
    **On a fix run, both calls read the fix run's own diff.** A worktree that has a since-close sha this run — the one **The late-fix reduction** (`skills/flow/review-panel-late-fix.md`) read for it — feeds them `git -C <worktree> diff --name-only <since-close-sha>..HEAD` in place of `<merge-base>..HEAD`, so a fix run whose delta touches no `ui paths` skips this stage at step 2 as exit 1 does.
-   A worktree whose since-close sha is `-` keeps `<merge-base>..HEAD`, and so does every worktree when the change's newest verifier dispatch row (the newest row `flow record dispatches -change <name>` returns with `role` `verifier` and a `key` starting with `visual-verify` — the rows `scripts/check-visual-verify-dispatched.sh` counts) does not have `outcome` `completed` — an earlier run's UI change that was never verified may lie before the since-close sha.
+   A worktree whose since-close sha is `-` keeps `<merge-base>..HEAD`, and so does every worktree when the change's newest verifier dispatch row (the newest row `flow record dispatches -change <name>` returns with `role` `verifier` and a `key` starting with `visual-verify` — the rows `<agents repo>/scripts/check-visual-verify-dispatched.sh` counts) does not have `outcome` `completed` — an earlier run's UI change that was never verified may lie before the since-close sha.
 
 **A `skipped` decision ends the stage at step 2.** When this run's decision
 (`<abs-worktree>/.superpowers/sdd/decision.json`) records `visual.verify` `skipped` (step 6 of
````

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh` (with this change's `verbatim-moves.txt`), `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-normative-inventory.sh` all exit 0, `check-installed-citations.sh` printing `INSTALLED-CITATIONS-OK`. Commit.

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none** — one citation's root; `check-installed-citations.sh` in step 3 covers it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): verify-and-handoff's guard citation names its root
**After:** Task 3

**Decision:** citation-root-fix
