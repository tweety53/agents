# self-review-kan-941-step-6-playwright-fidelity-specs-for-the-cardio

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-18-self-review-kan-941-step-6-playwright-fidelity-specs-for-the-cardio-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` already lists the sentences Tasks 1, 3 and 6 add or reword. Never merge or push.
- Each task's change is the unified diff its Step 2 gives, applied with `git apply` from this worktree's root, in task order — every diff was applied and checked there (`design.md`, **Context**). Save a diff by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork.
- Run `cd stats && gofmt -w .` before committing any Go file; `gofmt -l .` and `go vet ./internal/guard/` stay clean.
- Run Go tests with a 600000 ms tool timeout; `go test ./internal/guard/` takes about 90 s.
  <!-- measured: cd stats && go test ./internal/guard/ -count=1 @ f4844c40 with all six diffs applied: ok in 88.3s -->

## Review focus

- A change branch that never merged anything must reshape exactly as before, and a sync merge of the base the merge base moved past must still reshape (Task 1, case 3c).
- A plan with a `**Tasks appended:**` line and no in-run task must judge exactly the same tasks as before (Task 2).
- A push of a change branch whose unpushed commits all carry a trailer, or whose branch origin does not hold yet, must go through untouched, and a checker that cannot run must never block it (Task 6, cases 105–107).
- `close-task.sh` on a branch origin does not hold yet must push as before (Task 6: every existing close-task case, whose fixture origin holds no branch).
- A plan task carrying one implementation path beside a record under `spectre/changes/` must pass (Task 5, case 38's last assertion).

- [x] 1. Stop the reshape on a merge of another branch
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Apply Step 2's diff to the test file(s) alone (`git apply --include='stats/internal/guard/reshape_branch_test.go' <the saved diff>`) and run `cd stats && go test ./internal/guard/ -run 'TestReshapeBranch' -count=1`: `3b a merged-in branch's commits stop the reshape and HEAD does not move` fails (`a foreign merge exits 1: rc=0`) — the defect. Keep the applied hunks; Step 2 applies the rest with `git apply --exclude='stats/internal/guard/reshape_branch_test.go' <the saved diff>`.
  - [x] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at f4844c40; with it applied go test ./internal/guard/ -run TestReshapeBranch passes, and with reshapebranch.go reverted case 3b fails
diff --git a/scripts/reshape-branch.sh b/scripts/reshape-branch.sh
index df2d8aee..c1b2fd14 100755
--- a/scripts/reshape-branch.sh
+++ b/scripts/reshape-branch.sh
@@ -14,6 +14,14 @@
 # calls check-planning-commit-location.sh first and stops on its verdict,
 # printing the guard's own lines.
 #
+# A merge that brought in commits <merge-base> does not reach stops the
+# reshape before HEAD moves (kan-941: a develop merge folded develop's
+# commits into the change commit): the pipeline never merges into a change
+# branch, so every commit of <merge-base>..HEAD off the first-parent chain
+# is another branch's. Prints `RESHAPE-FOREIGN-MERGE: <worktree> — <n>
+# commit(s) …`, one `  <sha12> <subject>` line per such commit and the
+# remedy line, and exits 1.
+#
 # Prints `RESHAPED: <worktree> — reset to <short sha>` and exits 0. Exits 2
 # with a reason on stderr on bad arguments or a base that does not resolve,
 # the guard's own code on a guard refusal, and git's own code on a git
diff --git a/skills/flow-contracts/finish-contract-run1.md b/skills/flow-contracts/finish-contract-run1.md
index 07c94d8b..6817ecba 100644
--- a/skills/flow-contracts/finish-contract-run1.md
+++ b/skills/flow-contracts/finish-contract-run1.md
@@ -218,7 +218,13 @@ to that merge base: every commit the run made on the branch — task, fixup and
 alike, the planning ones being the plan-gate, link, reviewer-dispatch, fix-run and
 write-in-progress commits of **Planning commits** (`git-boundaries.md`) and `/flow-plan`'s capture
 commit — folds back into the index, uncommitted. **Planning commits are squashed into one on
-landing**: they exist on the branch during the run, never on the base. Before that reset,
+landing**: they exist on the branch during the run, never on the base. **A merge of another
+branch stops the reshape**: the pipeline never merges into a change branch — its sync rebases — so
+a commit that merge base does not reach and the branch's first-parent chain does not carry is
+another branch's, and the reset would fold it under the change's subject (kan-941 folded `develop`'s
+and `main`'s commits that way). The script then prints `RESHAPE-FOREIGN-MERGE`, names each such
+commit, and exits 1 before HEAD moves; the run stops and relays its remedy — merge that branch
+into the change's base first, or rebase the change branch onto a base that holds them. Before that reset,
 `integrate-change.sh commit` pins the worktree's HEAD as `refs/flow/pre-reshape/<name>` — kept when it
 already reaches HEAD, otherwise advanced to a merge of the old pin and HEAD, pushed plain beside the branch push and never deleted by run 2 — so a
 task's own commit stays findable afterwards as the one carrying its `Task-Id: <n>` trailer
diff --git a/stats/internal/guard/reshape_branch_test.go b/stats/internal/guard/reshape_branch_test.go
index 90c48d60..4f2c54a5 100644
--- a/stats/internal/guard/reshape_branch_test.go
+++ b/stats/internal/guard/reshape_branch_test.go
@@ -151,6 +151,41 @@ func TestReshapeBranch(t *testing.T) {
 		check(t, splitGit(t, fx.main, nil, "rev-parse", "HEAD") == before, "HEAD untouched", "HEAD moved")
 	})
 
+	t.Run("3b a merged-in branch's commits stop the reshape and HEAD does not move", func(t *testing.T) {
+		t.Parallel()
+		fx := rbNew(t)
+		w := fx.wt
+		splitGit(t, fx.main, nil, "checkout", "-q", "-b", "develop")
+		writeFile(t, fx.main+"/src/d.txt", "develop\n")
+		splitGit(t, fx.main, nil, "add", "-A", "--", "src")
+		splitGit(t, fx.main, nil, "-c", "user.name=dev", "-c", "user.email=dev@example.com", "commit", "-q", "-m", "feat(src): develop work")
+		writeFile(t, w+"/src/a.txt", "task\n")
+		fx.commitAs(t, "impl", "feat(src): task\n\nTask-Id: 1", "src")
+		splitGit(t, w, nil, "-c", "user.name=impl", "-c", "user.email=impl@example.com", "merge", "-q", "--no-ff", "-m", "chore: merge develop", "develop")
+		before := splitGit(t, w, nil, "rev-parse", "HEAD")
+		rc, out := fx.reshape(t, w, w, "demo", fx.base)
+		check(t, rc == 1, "a foreign merge exits 1", "rc=%d out=%s", rc, out)
+		check(t, strings.Contains(out, "RESHAPE-FOREIGN-MERGE: "+w+" — 1 commit(s)") && strings.Contains(out, "feat(src): develop work"),
+			"the verdict names the merged-in commit", "out=%s", out)
+		check(t, !strings.Contains(out, "feat(src): task"), "the change's own commit is not named", "out=%s", out)
+		check(t, splitGit(t, w, nil, "rev-parse", "HEAD") == before, "HEAD untouched", "HEAD moved")
+	})
+
+	t.Run("3c a sync merge of the base the merge base moved to reshapes", func(t *testing.T) {
+		t.Parallel()
+		fx := rbNew(t)
+		w := fx.wt
+		writeFile(t, fx.main+"/src/m.txt", "base moved\n")
+		splitGit(t, fx.main, nil, "add", "-A", "--", "src")
+		splitGit(t, fx.main, nil, "-c", "user.name=dev", "-c", "user.email=dev@example.com", "commit", "-q", "-m", "feat(src): base work")
+		moved := splitGit(t, fx.main, nil, "rev-parse", "HEAD")
+		writeFile(t, w+"/src/a.txt", "task\n")
+		fx.commitAs(t, "impl", "feat(src): task\n\nTask-Id: 1", "src")
+		splitGit(t, w, nil, "-c", "user.name=impl", "-c", "user.email=impl@example.com", "merge", "-q", "--no-ff", "-m", "chore: merge main", "main")
+		rc, out := fx.reshape(t, w, w, "demo", moved)
+		check(t, rc == 0 && strings.HasSuffix(out, "\nRESHAPED: "+w+" — reset to "+moved[:12]), "reshape verdict", "rc=%d out=%s", rc, out)
+	})
+
 	t.Run("4 a base that does not resolve cannot answer", func(t *testing.T) {
 		t.Parallel()
 		fx := rbNew(t)
diff --git a/stats/internal/guard/reshapebranch.go b/stats/internal/guard/reshapebranch.go
index f72c53ee..f177e754 100644
--- a/stats/internal/guard/reshapebranch.go
+++ b/stats/internal/guard/reshapebranch.go
@@ -31,6 +31,37 @@ func reshapeBranch(args []string, env Env, stdout, stderr io.Writer) int {
 		return 2
 	}
 	baseSHA := strings.TrimRight(b.String(), "\n")
+	// A FOREIGN MERGE STOPS THE RESHAPE (kan-941): a commit a merge's side
+	// parent brought that <merge-base> does not reach would fold into the
+	// change's implementation commit under the change's subject. The
+	// pipeline never merges into a change branch -- its sync rebases -- so
+	// any such commit is foreign: every commit of <merge-base>..HEAD off
+	// the first-parent chain.
+	all, ok1 := capture(git("-C", wt, "rev-list", baseSHA+"..HEAD"))
+	own, ok2 := capture(git("-C", wt, "rev-list", "--first-parent", baseSHA+"..HEAD"))
+	if !ok1 || !ok2 {
+		fmt.Fprintf(stderr, "reshape-branch.sh: git cannot list %s..HEAD in %s\n", baseSHA[:12], wt)
+		return 2
+	}
+	mine := map[string]bool{}
+	for _, c := range strings.Fields(own) {
+		mine[c] = true
+	}
+	var foreign []string
+	for _, c := range strings.Fields(all) {
+		if !mine[c] {
+			foreign = append(foreign, c)
+		}
+	}
+	if len(foreign) > 0 {
+		fmt.Fprintf(stdout, "RESHAPE-FOREIGN-MERGE: %s — %d commit(s) a merge brought in that %s does not reach would fold into the change commit:\n", wt, len(foreign), baseSHA[:12])
+		for _, c := range foreign {
+			subj, _ := capture(git("-C", wt, "log", "-1", "--format=%s", c))
+			fmt.Fprintf(stdout, "  %s %s\n", c[:12], sanitizeDisplay(strings.TrimSpace(subj)))
+		}
+		fmt.Fprintln(stdout, "  merge that branch into the change's base first, or rebase the change branch onto a base that holds them, then re-run")
+		return 1
+	}
 	reset := git("-C", wt, "reset", "-q", "--soft", baseSHA)
 	reset.Stdout, reset.Stderr = stderr, stderr
 	if rc := exitCode(reset.Run()); rc != 0 {
````

  - [x] **Step 3: Verify.** `cd stats && go test ./internal/guard/ -run 'TestReshapeBranch' -count=1` passes; `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go vet ./internal/guard/` exits 0; `scripts/check-installed-citations.sh`, `scripts/check-verbatim-moves.sh` and `scripts/check-markdown-integrity.py` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/reshapebranch.go`, `stats/internal/guard/reshape_branch_test.go`, `scripts/reshape-branch.sh`, `skills/flow-contracts/finish-contract-run1.md`
**Tests:** `3b a merged-in branch's commits stop the reshape and HEAD does not move`, `3c a sync merge of the base the merge base moved to reshapes`
**Regression:** reverting `reshapeBranch`'s foreign-commit check lets a merged-in `develop` commit fold into the index, so case 3b fails `a foreign merge exits 1`; case 3c fails if the check refuses a merge whose side the merge base already reaches.
**Baseline:** before=6 after=8
<!-- measured: grep -c 't.Run("' stats/internal/guard/reshape_branch_test.go @ f4844c40 (before; after with this task's diff applied) -->
**Commit:** fix(guard): stop the reshape on a merge of another branch
**After:** none

**Decision:** reshape-foreign-merge-stop

- [x] 2. Judge in-run fix tasks in the unfinished-work gate
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Apply Step 2's diff to the test file(s) alone (`git apply --include='stats/internal/guard/check_unfinished_work_test.go' <the saved diff>`) and run `cd stats && go test ./internal/guard/ -run 'TestCheckUnfinishedWork' -count=1`: `an in-run fix task with no dispatch row is OUTSTANDING` and `the breakdown names the undispatched in-run fix task` fail — the defect. Keep the applied hunks; Step 2 applies the rest with `git apply --exclude='stats/internal/guard/check_unfinished_work_test.go' <the saved diff>`.
  - [x] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at f4844c40 after Task 1; with it applied go test ./internal/guard/ -run 'TestCheckUnfinishedWork|TestCloseTask' passes, and with unfinishedwork.go reverted the two in-run OUTSTANDING cases fail
diff --git a/scripts/check-unfinished-work.sh b/scripts/check-unfinished-work.sh
index 192b5804..2e0c27fa 100755
--- a/scripts/check-unfinished-work.sh
+++ b/scripts/check-unfinished-work.sh
@@ -55,7 +55,9 @@
 #
 # SIGNAL FOUR: A TICKED APPENDED TASK NO DISPATCH ROW NAMES IS UNFINISHED
 # (kan-979). Only the last <n> tasks of the plan are judged, <n> from its
-# `**Tasks appended:** <n>` header line (absent: 0, the signal never fires);
+# `**Tasks appended:** <n>` header line (absent: 0), and every task titled
+# `In-run fix <k> — …` wherever it sits — verify-fix-loop.md's loop appends
+# those without raising the count (kan-941: tasks 12-16 went unjudged);
 # a row names a task by its `taskId` or by the `+`-joined ids of a
 # `task-<ids>[-…]` key. Plan-time tasks run grouped or inline and carry no row
 # of their own, so they are never judged.
diff --git a/scripts/close-task.sh b/scripts/close-task.sh
index 61ece831..d7a439da 100755
--- a/scripts/close-task.sh
+++ b/scripts/close-task.sh
@@ -15,7 +15,8 @@
 #      end for <key> did not reach the store; run: flow record dispatch end
 #      …` (the full command) as the last stderr line, turning exit 0 into 3.
 #      Then, for every task it closes that the plan's `**Tasks appended:**`
-#      line counts, `flow record dispatches -change <name> -C <canonical>`:
+#      line counts or whose title opens `In-run fix <k> — `,
+#      `flow record dispatches -change <name> -C <canonical>`:
 #      an appended task no row names is refused at step 4 (exit 1); a
 #      store that cannot be read refuses nothing.
 #   2. check-task-commit-fields.sh <canonical> <id> <sha|map> "" <canonical>
diff --git a/stats/internal/guard/check_unfinished_work_test.go b/stats/internal/guard/check_unfinished_work_test.go
index d0a4eafa..100187be 100644
--- a/stats/internal/guard/check_unfinished_work_test.go
+++ b/stats/internal/guard/check_unfinished_work_test.go
@@ -333,6 +333,14 @@ func TestCheckUnfinishedWork(t *testing.T) {
 		uwVerdict("a reviewer row begun before the implementer row does not review the appended task", "OUTSTANDING:"))
 	each(demo(planned(appended, `[{"seq":1,"key":"task-7-implementer","role":"implementer","outcome":"completed"},{"seq":2,"key":"pipeline-fix-1-review-1","role":"reviewer","outcome":"clean"}]`)),
 		uwVerdict("a non-panel reviewer row does not review the appended task", "OUTSTANDING:"))
+	// An in-run fix task (verify-fix-loop.md step 1) never raises
+	// **Tasks appended:**, and is judged all the same (kan-941's tasks 12-16).
+	inRun := "- [x] 1. done\n- [x] 7. In-run fix 1 — flow.visual-verify: the menu shadow\n"
+	each(demo(planned(inRun, `[{"seq":1,"key":"task-1-implementer","role":"implementer","outcome":"completed"}]`)),
+		uwVerdict("an in-run fix task with no dispatch row is OUTSTANDING", "OUTSTANDING:"),
+		uwReason("the breakdown names the undispatched in-run fix task", "1 ticked appended task(s) with no dispatch row (Task 7)"))
+	each(demo(planned(inRun, `[{"seq":1,"key":"task-7-implementer","role":"implementer","outcome":"completed"},{"seq":2,"key":"panel-2-primary","role":"reviewer","outcome":"clean"}]`)),
+		uwVerdict("an in-run fix task with its implementer row and a later reviewer row is CLEAR", "CLEAR:"))
 	each(demo(planned("- [x] 1. done\n- [x] 2. grouped\n- [x] 3. grouped\n", `[{"key":"task-2-implementer","outcome":"completed"}]`)),
 		uwVerdict("plan-time tasks grouped under one key, with no Tasks appended line, are CLEAR", "CLEAR:"))
 	// 2e. Signal five — a panel round dispatched with no pass row (kan-939:
diff --git a/stats/internal/guard/unfinishedwork.go b/stats/internal/guard/unfinishedwork.go
index 1f6371f8..1f8966ba 100644
--- a/stats/internal/guard/unfinishedwork.go
+++ b/stats/internal/guard/unfinishedwork.go
@@ -459,8 +459,8 @@ func checkUnfinishedWork(args []string, env Env, stdout, stderr io.Writer) int {
 		// round that recorded nothing (kan-979's fix-round Task 7 reached
 		// integrate with no dispatch row at all). A row names a task by its
 		// taskId or by the `+`-joined ids of a task-<ids>[-…] key. Only the tasks a
-		// fix run appended — the last `**Tasks appended:** <n>` of the plan —
-		// are judged: each is implemented alone, so it carries a row of its
+		// fix run appended — the last `**Tasks appended:** <n>` of the plan,
+		// and every `In-run fix <k> — …` task (uwAppendedIn) — are judged: each is implemented alone, so it carries a row of its
 		// own, while plan-time tasks run grouped under one task-<n> key or
 		// inline with one row per bundle and most have none. It needs the
 		// plan read above. An appended task also needs a `reviewer` row whose
@@ -590,7 +590,10 @@ func uwDerivedRange(env Env, worktree string) ([2]string, bool) {
 var (
 	uwTask          = regexp.MustCompile(`^- \[([ xX])\] ([0-9]+)\. `)
 	uwTasksAppended = regexp.MustCompile(`^(?:> )?\*\*Tasks appended:\*\* ([0-9]+)`)
-	uwTaskKey       = regexp.MustCompile(`^task-([0-9]+(?:\+[0-9]+)*)(?:-|$)`)
+	// uwInRunFix is verify-fix-loop.md's in-run fix task title, appended
+	// without raising **Tasks appended:** (kan-941: tasks 12-16 unjudged).
+	uwInRunFix = regexp.MustCompile(`^- \[[ xX]\] [0-9]+\. In-run fix [0-9]+ — `)
+	uwTaskKey  = regexp.MustCompile(`^task-([0-9]+(?:\+[0-9]+)*)(?:-|$)`)
 )
 
 // uwRowTasks is the task ids a dispatch row names: its taskId, and the
@@ -617,20 +620,24 @@ func uwAppended(plan string, tickedOnly bool) ([]string, bool) {
 	return uwAppendedIn(string(b), tickedOnly), true
 }
 
-// uwAppendedIn is uwAppended over a plan's text.
+// uwAppendedIn is uwAppended over a plan's text. A task titled
+// `In-run fix <k> — …` (verify-fix-loop.md step 1) counts as appended
+// wherever it sits, since that loop never raises **Tasks appended:**.
 func uwAppendedIn(plan string, tickedOnly bool) []string {
 	appended := 0
 	var tasks [][]string
+	var inRun []bool
 	for l := range strings.Lines(plan) {
 		if m := uwTasksAppended.FindStringSubmatch(l); m != nil {
 			appended, _ = strconv.Atoi(m[1])
 		} else if m := uwTask.FindStringSubmatch(l); m != nil {
 			tasks = append(tasks, m)
+			inRun = append(inRun, uwInRunFix.MatchString(l))
 		}
 	}
 	var ids []string
-	for _, m := range tasks[max(0, len(tasks)-appended):] {
-		if !tickedOnly || m[1] != " " {
+	for i, m := range tasks {
+		if (i >= len(tasks)-appended || inRun[i]) && (!tickedOnly || m[1] != " ") {
 			ids = append(ids, m[2])
 		}
 	}
````

  - [x] **Step 3: Verify.** `cd stats && go test ./internal/guard/ -run 'TestCheckUnfinishedWork|TestCloseTask' -count=1` passes; `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go vet ./internal/guard/` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `scripts/check-unfinished-work.sh`, `scripts/close-task.sh`
**Tests:** `an in-run fix task with no dispatch row is OUTSTANDING`, `the breakdown names the undispatched in-run fix task`, `an in-run fix task with its implementer row and a later reviewer row is CLEAR`
**Regression:** reverting `uwAppendedIn`'s `uwInRunFix` match leaves an `In-run fix <k> — …` task unjudged, so the two OUTSTANDING cases read CLEAR and fail.
**Baseline:** before=89 after=92
<!-- measured: grep -cE 'uw(Verdict|Reason)\("' stats/internal/guard/check_unfinished_work_test.go @ f4844c40 (before; after with this task's diff applied) -->
**Commit:** fix(guard): judge in-run fix tasks in the unfinished-work gate
**After:** none

**Decision:** in-run-fix-tasks-appended

- [x] 3. Flag a structural unfinished-work verdict once, at the canonical worktree
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `grep -n -- '-C <worktree>' skills/flow/unfinished-work-gate.md` prints the per-worktree call (line 86), and `sed -n '/func (s \*Store) FlagVerdictFalsePositive/,/^}/p' stats/internal/store/records.go` shows the query keyed on project, change and guard alone — the defect.
  - [x] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at f4844c40 after Tasks 1-2; with it applied scripts/check-installed-citations.sh, scripts/check-verbatim-moves.sh and scripts/check-markdown-integrity.py exit 0
diff --git a/skills/flow/unfinished-work-gate.md b/skills/flow/unfinished-work-gate.md
index 2949a179..eb08fa1c 100644
--- a/skills/flow/unfinished-work-gate.md
+++ b/skills/flow/unfinished-work-gate.md
@@ -78,12 +78,17 @@ The signals that produce that list are the script's own.
 ## A verdict the operator calls structural
 
 When the operator's answer says the verdict was verified structural — the plan held in another
-worktree with every task ticked and no open finding — run, once per worktree that reported
-`OUTSTANDING` and before the landing question:
+worktree with every task ticked and no open finding — run once, before the landing question:
 
 ```bash verified:the flag set is design.md §5's; the call shape mirrors the `flow record verdict` invocation in scripts/check-unfinished-work.sh
 flow record verdict false-positive -change <name> -guard check-unfinished-work \
-  -reason "<the operator's reason, verbatim>" -C <worktree>
+  -reason "<the operator's reason, verbatim>" -C <canonical-worktree>
 ```
 
+Once, and at the canonical worktree, whichever worktrees reported `OUTSTANDING`: the guard records
+every worktree's verdict under the canonical project (THE STORE CALLS ANCHOR AT THE PLAN'S PROJECT,
+`<agents repo>/scripts/check-unfinished-work.sh`), and the command flags that project's newest verdict for the
+change — a satellite's `-C` names a project holding none, and the store refuses it (kan-941:
+"dispatch not found" for both satellites).
+
 A write that falls back to the journal is one warning line and the run continues.
````

  - [x] **Step 3: Verify.** `scripts/check-installed-citations.sh`, `scripts/check-verbatim-moves.sh` and `scripts/check-markdown-integrity.py` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/unfinished-work-gate.md`
**Tests:** **none** — a contract-text fix: the store already flags one verdict per change and guard, and the call's `-C` is the only thing that changes
**Regression:** none — no test; reverting the text sends the gate back to one refused call per satellite worktree.
**Commit:** docs(flow): flag a structural unfinished-work verdict once at the canonical worktree
**After:** none

**Decision:** structural-verdict-canonical

- [x] 4. Read only the first-parent chain in check-untasked-commits
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Apply Step 2's diff to the test file(s) alone (`git apply --include='stats/internal/guard/check_untasked_commits_test.go' <the saved diff>`) and run `cd stats && go test ./internal/guard/ -run 'TestCheckUntaskedCommits' -count=1`: `TestCheckUntaskedCommitsSkipsAMergedBranch` fails with `feat(y): develop work` listed (exit 1) — the defect. Keep the applied hunks; Step 2 applies the rest with `git apply --exclude='stats/internal/guard/check_untasked_commits_test.go' <the saved diff>`.
  - [x] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at f4844c40 after Tasks 1-3; with it applied go test ./internal/guard/ -run TestCheckUntaskedCommits passes, and with untaskedcommits.go reverted TestCheckUntaskedCommitsSkipsAMergedBranch fails
diff --git a/scripts/check-untasked-commits.sh b/scripts/check-untasked-commits.sh
index f939bbbc..0d5519c2 100755
--- a/scripts/check-untasked-commits.sh
+++ b/scripts/check-untasked-commits.sh
@@ -10,7 +10,9 @@
 # every task"). Run 1's reshape then squashes it into the implementation
 # commit, and the evidence is gone.
 #
-# Every non-merge commit of <base>..HEAD is listed unless it carries a
+# Every non-merge commit on the first-parent chain of <base>..HEAD is listed
+# — a merged-in branch's commits are that branch's own, never the change's
+# (kan-941: a develop merge listed develop's commits) — unless it carries a
 # non-empty `Task-Id:` trailer (a task's own commit), a non-empty `Flow-Key:`
 # trailer (a commit a run makes outside every task on a sanctioned route —
 # **Commits outside every task**, skills/flow-contracts/git-boundaries.md),
diff --git a/stats/internal/guard/check_untasked_commits_test.go b/stats/internal/guard/check_untasked_commits_test.go
index 04020fba..b7f17358 100644
--- a/stats/internal/guard/check_untasked_commits_test.go
+++ b/stats/internal/guard/check_untasked_commits_test.go
@@ -73,6 +73,34 @@ func TestCheckUntaskedCommits(t *testing.T) {
 	}
 }
 
+// TestCheckUntaskedCommitsSkipsAMergedBranch pins --first-parent: a
+// trailerless commit a merged-in branch brought is not the change's own
+// (kan-941's develop merge listed develop's commits as UNTASKED).
+func TestCheckUntaskedCommitsSkipsAMergedBranch(t *testing.T) {
+	t.Parallel()
+	dir := t.TempDir()
+	g := &fxGit{}
+	g.git(dir, "init", "-q", "-b", "main")
+	g.git(dir, "commit", "-q", "--allow-empty", "-m", "base")
+	base := g.git(dir, "rev-parse", "HEAD")
+	g.git(dir, "checkout", "-q", "-b", "develop")
+	writeFile(t, filepath.Join(dir, "y.kt"), "y\n")
+	g.git(dir, "add", "-A")
+	g.git(dir, "commit", "-qm", "feat(y): develop work")
+	g.git(dir, "checkout", "-q", "main")
+	writeFile(t, filepath.Join(dir, "a.kt"), "a\n")
+	g.git(dir, "add", "-A")
+	g.git(dir, "commit", "-qm", "feat(app): a", "-m", "Task-Id: 1")
+	g.git(dir, "merge", "-q", "--no-ff", "-m", "chore: merge develop", "develop")
+	if g.err != nil {
+		t.Fatal(g.err)
+	}
+	var out, errb bytes.Buffer
+	if code := checkUntaskedCommits([]string{dir, base}, Env{Getenv: os.Getenv, Dir: dir}, &out, &errb); code != 0 || strings.Contains(out.String(), "develop work") {
+		t.Fatalf("exit %d, want 0 with the merged branch's commit unlisted\nstdout:\n%s\nstderr:\n%s", code, out.String(), errb.String())
+	}
+}
+
 // TestCheckUntaskedCommitsCannotAnswer pins exit 2 on a usage error, a
 // missing worktree and a base git cannot resolve.
 func TestCheckUntaskedCommitsCannotAnswer(t *testing.T) {
diff --git a/stats/internal/guard/untaskedcommits.go b/stats/internal/guard/untaskedcommits.go
index 365654c6..e62d929d 100644
--- a/stats/internal/guard/untaskedcommits.go
+++ b/stats/internal/guard/untaskedcommits.go
@@ -29,7 +29,9 @@ func checkUntaskedCommits(args []string, env Env, stdout, stderr io.Writer) int
 		return 2
 	}
 	git := envGit(env)
-	out, ok := capture(git("-C", worktree, "log", "--no-merges",
+	// --first-parent: a merged-in branch's commits are that branch's own,
+	// never this change's -- reshape-branch refuses such a merge outright.
+	out, ok := capture(git("-C", worktree, "log", "--no-merges", "--first-parent",
 		"--format=%H%x00%s%x00%(trailers:key=Task-Id,valueonly,separator=%x2C)%x00%(trailers:key=Flow-Key,valueonly,separator=%x2C)%x1e",
 		base+"..HEAD"))
 	if !ok {
````

  - [x] **Step 3: Verify.** `cd stats && go test ./internal/guard/ -run 'TestCheckUntaskedCommits' -count=1` passes; `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go vet ./internal/guard/` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/untaskedcommits.go`, `stats/internal/guard/check_untasked_commits_test.go`, `scripts/check-untasked-commits.sh`
**Tests:** `TestCheckUntaskedCommitsSkipsAMergedBranch`
**Regression:** reverting `--first-parent` lists the merged-in branch's trailerless commit as UNTASKED, so `TestCheckUntaskedCommitsSkipsAMergedBranch` fails.
**Baseline:** before=2 after=3
<!-- measured: grep -c '^func Test' stats/internal/guard/check_untasked_commits_test.go @ f4844c40 (before; after with this task's diff applied) -->
**Commit:** fix(guard): read only the first-parent chain in check-untasked-commits
**After:** none

**Decision:** untasked-first-parent

- [x] 5. Refuse a plan task whose files are all planning paths
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Apply Step 2's diff to the test file(s) alone (`git apply --include='scripts/test-check-plan-shape.sh' <the saved diff>`) and run `bash scripts/test-check-plan-shape.sh`: `case 38 (F14): rc=0 out=` and the case 38 mutation fail — the defect. Keep the applied hunks; Step 2 applies the rest with `git apply --exclude='scripts/test-check-plan-shape.sh' <the saved diff>`.
  - [x] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at f4844c40 after Tasks 1-4; with it applied bash scripts/test-check-plan-shape.sh prints all cases passed and scripts/check-plan-shape.sh exits 0, and with check-plan-shape.py reverted case 38 fails
diff --git a/scripts/check-plan-shape.py b/scripts/check-plan-shape.py
index 1dacc13b..022f48c4 100755
--- a/scripts/check-plan-shape.py
+++ b/scripts/check-plan-shape.py
@@ -1,11 +1,11 @@
 #!/usr/bin/env python3
 """check-plan-shape.py — check that a `tasks.md` is SHAPED so the guard
 that will later judge its commits (`check-task-commit-fields.py`) can
-actually read it. Thirteen findings, F1-F13 (F1-F6's canonical definitions,
+actually read it. Fourteen findings, F1-F14 (F1-F6's canonical definitions,
 decisions and rationale: `spectre/changes/kan-121-run-the-guards-own-
 parsers-over-tasks-md-at-plan/design.md` — do not restate them here, a
 second copy is a Single Source of Truth violation, the same class of drift
-stats/internal/guard/planprovenance.go's ported docstring warns against; F7-F13's canonical
+stats/internal/guard/planprovenance.go's ported docstring warns against; F7-F14's canonical
 definitions are the rows in the Findings list below).
 
 This guard IMPORTS the real parsers rather than reimplementing their
@@ -34,7 +34,7 @@ Exit codes:
   0  clean — every task in the file is shaped so the real parsers read it
      without ambiguity (including a file with zero findings; a file with
      zero TASKS is F5, not clean — see below).
-  1  one or more of F1-F13 found. Printed one per line as
+  1  one or more of F1-F14 found. Printed one per line as
      `file:line: message`, naming the task id (F5 excepted: a file with no
      tasks has no task id to name).
   2  invocation error — wrong argument count, or the file cannot be read.
@@ -159,6 +159,13 @@ field-shaped line inside a worked example is not a declaration.
       escape) after it.
       Reported at that line, naming the task.
 
+  F14 A task whose `**Files:**` paths all lie under the planning paths —
+      `spectre/changes/`, `openspec/changes/` or `docs/superpowers/`
+      (`PLANNING_PREFIXES`). Its commit cannot carry a `Task-Id:`:
+      check-task-commit-planning-paths.sh refuses a task commit on those
+      paths, so the task lands as a planning commit and is ticked by hand
+      (kan-941's Task 4, `visual-sweep.md`). Reported at the task line.
+
 A duplicate task id (two task lines sharing one id) is check-task-build-
 green.py's own violation, not this guard's: `collect_task_ids` de-
 duplicates by first occurrence, and `select_task`/`parse_task_fields`
@@ -265,6 +272,10 @@ INDENTED_FIELD_RE = re.compile(r"^[ \t]+" + FIELD_RE.pattern.lstrip("^"))
 # a "test-shaped path".
 TEST_SHAPED_PATTERNS = ("test-*", "*_test.go", "*.test.*", "*.spec.ts")
 
+# PLANNING_PREFIXES — F14's planning paths: check-task-commit-planning-paths'
+# `<leaf>/changes/` for both spec leaves, and `docs/superpowers/`.
+PLANNING_PREFIXES = ("spectre/changes/", "openspec/changes/", "docs/superpowers/")
+
 # MEASURED_GREP_TEST_RE — F13's `grep [opts] '@Test'` shape inside a
 # `measured:` comment: group 1 the option tokens, the lookahead rejecting a
 # pattern that continues as a longer word or a word-boundary escape.
@@ -497,6 +508,14 @@ def _check_task(
             f"{path}:{task_line}: task {task_id} has no **Files:** line, "
             "or one naming no file at all"
         )
+    if fields.files and all(f.startswith(PLANNING_PREFIXES) for f in fields.files):  # F14
+        violations.append(
+            f"{path}:{task_line}: task {task_id}'s **Files:** all lie under "
+            "the planning paths (spectre/changes/, openspec/changes/, "
+            "docs/superpowers/), so its commit cannot carry a `Task-Id:` "
+            "— name the file the task's work lands in, or make the record "
+            "a planning commit with no task"
+        )
 
     # NONE_OPEN_RE and REPAIRS_OPEN_RE are applied to fields.tests_value —
     # the SAME continuation-joined string parse_task_fields itself builds
diff --git a/scripts/test-check-plan-shape.sh b/scripts/test-check-plan-shape.sh
index adc40342..d280fd1c 100755
--- a/scripts/test-check-plan-shape.sh
+++ b/scripts/test-check-plan-shape.sh
@@ -8,7 +8,7 @@
 # REAL fixture files on disk — never a copy of its logic, never its Python
 # internals imported and asserted on directly.
 #
-# One case per finding (F1-F11), a clean-plan case, a no-arg aggregation
+# One case per finding (F1-F14), a clean-plan case, a no-arg aggregation
 # case, and three exit-2 cases (missing sibling module, unreadable file,
 # usage error). Additionally, for each of F1-F6 and F11, a MUTATION case: a
 # throwaway copy of check-plan-shape.py with that finding's own check
@@ -1042,6 +1042,27 @@ sed -i.bak "s/grep -c -m1 --include=\*w\* '@Test'/grep -cw --include=*.kt '@Test
 run_guard "$TASKS_MD"
 [ "$RC" -eq 0 ] && pass "case 37 (F13): grep -cw --include=*.kt '@Test' passes" || fail "case 37 (F13) -cw --include: rc=$RC out=$OUT"
 
+# ===========================================================================
+# Case 38 (F14): a task whose Files all lie under the planning paths cannot
+# carry a Task-Id: (kan-941's Task 4); one implementation path clears it.
+# ===========================================================================
+new_fixture
+{
+  printf -- '- [ ] 1. Sweep the frames\n\n'
+  printf '**Files:** `spectre/changes/x/visual-sweep.md`, `docs/superpowers/notes.md`\n'
+  printf '**Tests:** none\n'
+  printf '**Commit:** `docs(spectre): sweep`\n'
+} > "$TASKS_MD"
+run_guard "$TASKS_MD"
+case "$RC:$OUT" in
+  1:*"task 1's **Files:** all lie under the planning paths"*) pass "case 38 (F14): a planning-paths-only task fails" ;;
+  *) fail "case 38 (F14): rc=$RC out=$OUT" ;;
+esac
+assert_mutation_removes_finding "F14" "$TASKS_MD" "all lie under the planning paths" "case 38 (F14) mutation"
+sed -i.bak 's#`docs/superpowers/notes.md`#`docs/visual-sweep.md`#' "$TASKS_MD"
+run_guard "$TASKS_MD"
+[ "$RC" -eq 0 ] && pass "case 38 (F14): one implementation path passes" || fail "case 38 (F14) with an implementation path: rc=$RC out=$OUT"
+
 # ===========================================================================
 # Unlanded archives: the scan cannot list the archived changes the base does
 # not carry -> exit 2, never a scan that silently skips them. A copy of the
````

  - [x] **Step 3: Verify.** `bash scripts/test-check-plan-shape.sh` prints `all cases passed`; `scripts/check-plan-shape.sh` and `scripts/check-python-suppressions.sh` exit 0. Commit.

**Build:** green
**Files:** `scripts/check-plan-shape.py`, `scripts/test-check-plan-shape.sh`
**Tests:** `case 38 (F14): a planning-paths-only task fails`, `case 38 (F14): one implementation path passes`
**Regression:** reverting the `# F14` check lets the planning-paths-only fixture exit 0, so `case 38 (F14): a planning-paths-only task fails` fails; the second case fails if the check refuses a task with one implementation path.
**Baseline:** before=25 after=27
<!-- measured: grep -cE '^(case "\$RC|\[ "\$RC" -eq 0 \] && pass)' scripts/test-check-plan-shape.sh @ f4844c40 (before; after with this task's diff applied) -->
**Commit:** fix(scripts): refuse a plan task whose files are all planning paths
**After:** none

**Decision:** plan-shape-planning-only-task

- [x] 6. Refuse a change-branch push carrying a commit outside every task with no Flow-Key
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Apply Step 2's diff to the test file(s) alone (`git apply --include='scripts/test-protect-main-checkout.sh' --include='stats/internal/guard/close_task_test.go' <the saved diff>`) and run `bash scripts/test-protect-main-checkout.sh`: it fails `104 a change-branch push carrying an untasked commit: wanted deny, got allow`, and `cd stats && go test ./internal/guard/ -run 'TestCloseTask' -count=1` fails `untasked unpushed commit: exit 1, no tick, no push` — the defect. Keep the applied hunks; Step 2 applies the rest with `git apply --exclude='scripts/test-protect-main-checkout.sh' --exclude='stats/internal/guard/close_task_test.go' <the saved diff>`.
  - [x] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at f4844c40 after Tasks 1-5; with it applied bash scripts/test-protect-main-checkout.sh prints all protect-main-checkout cases pass and go test ./internal/guard/ -run TestCloseTask passes, and with the hook and closetask.go reverted cases 104 and the close-task refusal fail
diff --git a/hooks/protect-main-checkout.py b/hooks/protect-main-checkout.py
index c81e4c97..a7b3a7f2 100755
--- a/hooks/protect-main-checkout.py
+++ b/hooks/protect-main-checkout.py
@@ -35,6 +35,13 @@ denied too unless scripts/check-loop-landing.sh, in the agents checkout this hoo
 from, admits it — flowd holding the loop's completed plan, fix and clean last review. That one
 check fails closed: a checker that cannot answer or cannot run denies the push.
 
+It also denies a push of a `/flow` change branch (`spectre/<name>`) whose unpushed commits —
+`origin/spectre/<name>..HEAD` — hold one outside every task with neither a `Task-Id:` nor a
+`Flow-Key:` trailer, as scripts/check-untasked-commits.sh judges them (**Commits outside every
+task**, skills/flow-contracts/git-boundaries.md; kan-941 pushed four such commits). It checks only a
+push of HEAD's commit to a branch origin already holds, and fails open: a checker that cannot
+answer, cannot run or times out lets the push through, and run 1's gate still lists the commit.
+
 Both landing checks read the push from the command text, so a push the hook cannot read is denied
 rather than let through: one whose repository (`-C <path>`, a `cd`), remote or refspec goes through
 a shell expansion the hook cannot resolve — `$W`, `"${B}:main"`, `$(git branch --show-current)`, a
@@ -93,6 +100,12 @@ LOOP_BRANCH_RE = re.compile(r"^(pipeline-fix-.+-[0-9]+|self-review-.+|spectre/.+
 LOOP_CHECK_TIMEOUT = 240
 LOOP_DEADLINE = None
 
+# The untasked-commits checker and its budget; overridable for the hook's own harness only.
+UNTASKED_CHECK = os.environ.get("UNTASKED_COMMITS_CHECK") or os.path.join(
+    os.path.dirname(os.path.dirname(os.path.realpath(__file__))), "scripts", "check-untasked-commits.sh"
+)
+UNTASKED_CHECK_TIMEOUT = 60
+
 # The loop-landing checker; overridable for the hook's own harness only.
 LOOP_CHECK = os.environ.get("LOOP_LANDING_CHECK") or os.path.join(
     os.path.dirname(os.path.dirname(os.path.realpath(__file__))), "scripts", "check-loop-landing.sh"
@@ -513,6 +526,35 @@ def loop_landing(args, target):
     return None
 
 
+def untasked_push(args, target):
+    """The deny reason when the push sends HEAD's commit to a `spectre/` branch origin holds and
+    check-untasked-commits.sh lists an unpushed commit carrying neither trailer, else None."""
+    if target is None:
+        return None
+    head = git(target, "rev-parse", "--verify", "-q", "HEAD")
+    for src, dst in push_refspecs(args, target):
+        remote = f"refs/remotes/origin/{dst}"
+        if not dst.startswith("spectre/") or not git(target, "rev-parse", "--verify", "-q", remote):
+            continue
+        if not head or git(target, "rev-parse", "--verify", "-q", f"{src}^{{commit}}") != head:
+            continue
+        try:
+            r = subprocess.run(
+                [UNTASKED_CHECK, target, remote], capture_output=True, text=True, timeout=UNTASKED_CHECK_TIMEOUT
+            )
+        except (OSError, subprocess.SubprocessError):
+            return None
+        if r.returncode == 1:
+            return (
+                f"Blocked: a commit this push sends to {dst} sits outside every task with neither a "
+                f"`Task-Id:` nor a `Flow-Key:` trailer:\n{r.stdout}\n"
+                "Add `-m \"Flow-Key: <key>\"` to it — `git commit --amend` for HEAD, a reword for an "
+                "earlier one — `<key>` its round's dispatch key or stage key (**Commits outside every "
+                "task**, skills/flow-contracts/git-boundaries.md), then push again."
+            )
+    return None
+
+
 def bash_hits(command, cwd, landings=None):
     """Paths a Bash command writes into or git-mutates, resolved against cwd and `cd`.
 
@@ -584,6 +626,9 @@ def bash_hits(command, cwd, landings=None):
                     loop = None if unreadable else loop_landing(args, target)
                     if loop:
                         landings.append(loop)
+                    untasked = None if unreadable else untasked_push(args, target)
+                    if untasked:
+                        landings.append(untasked)
                 i = j + 1
                 continue
             if tok == "sed" and any(t.startswith("-i") for t in tokens[i + 1 : i + 4]):
diff --git a/scripts/close-task.sh b/scripts/close-task.sh
index d7a439da..b03d7baf 100755
--- a/scripts/close-task.sh
+++ b/scripts/close-task.sh
@@ -23,7 +23,11 @@
 #      <name> for every task, and its commit's — every map pair's — Task-Id:
 #      trailer must name the task or a task its Build: red fold joins (exit 1
 #      otherwise): the trailer is how a task's commit is found again.
-#   3. check-task-commit-planning-paths.sh <canonical> <merge-base>.
+#   3. check-task-commit-planning-paths.sh <canonical> <merge-base>, then
+#      check-untasked-commits.sh <worktree> refs/remotes/origin/spectre/<name>
+#      once per distinct worktree whose branch origin already holds: the push
+#      at step 7 sends those commits, and one outside every task carries its
+#      `Flow-Key:` trailer before it leaves (kan-941).
 #   4. Any guard exit 2 → exit 2; else any exit 1 → exit 1; else an
 #      appended task step 1 found no row naming → exit 1. Nothing is ticked
 #      or pushed either way, so a refused commit is never pushed.
diff --git a/scripts/test-protect-main-checkout.sh b/scripts/test-protect-main-checkout.sh
index b0c10587..2c51697c 100755
--- a/scripts/test-protect-main-checkout.sh
+++ b/scripts/test-protect-main-checkout.sh
@@ -493,6 +493,42 @@ ln -s "$MAIN/sub" "$ROOT/elsewhere/lsub2"
 expect "103 rm and mv of a symlink into the checkout act on the link, outside it" allow \
   "$(run "$ROOT" Bash "{\"command\":$(q "rm $ROOT/elsewhere/lsub2 && mv $ROOT/elsewhere/lsub $ROOT/elsewhere/l3")}")"
 
+# 104-107: a push of a change branch runs check-untasked-commits over its
+# unpushed commits (kan-941); a stub stands in for the guard.
+UNTASKED_RC="$ROOT/untasked-rc"
+UNTASKED_LOG="$ROOT/untasked-log"
+cat >"$ROOT/untaskedcheck" <<STUB
+#!/usr/bin/env bash
+printf '%s\n' "\$@" >"$UNTASKED_LOG"
+echo "UNTASKED: 1 commit(s) carry no Task-Id: or Flow-Key: trailer"
+exit "\$(cat "$UNTASKED_RC")"
+STUB
+chmod +x "$ROOT/untaskedcheck"
+export UNTASKED_COMMITS_CHECK="$ROOT/untaskedcheck"
+CWT="$ROOT/main-worktrees/chg2"
+git_q -C "$MAIN" worktree add "$CWT" -b spectre/chg2 origin/main
+git_q -C "$CWT" push -q -u origin spectre/chg2
+echo m >"$CWT/m.txt"
+git_q -C "$CWT" add m.txt
+git_q -C "$CWT" commit -q -m "fix(app): review Minors"
+echo 1 >"$UNTASKED_RC"
+expect "104 a change-branch push carrying an untasked commit" deny \
+  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $CWT push origin spectre/chg2")}")"
+[ "$(tr '\n' ' ' <"$UNTASKED_LOG")" = "$CWT refs/remotes/origin/spectre/chg2 " ] \
+  && pass "104 checker args" || fail "104 checker args: $(tr '\n' ' ' <"$UNTASKED_LOG")"
+echo 0 >"$UNTASKED_RC"
+expect "105 a change-branch push the checker clears" allow \
+  "$(run "$ROOT" Bash "{\"command\":$(q "cd $CWT && git push")}")"
+echo 1 >"$UNTASKED_RC"
+rm -f "$UNTASKED_LOG"
+git_q -C "$CWT" checkout -q -b spectre/chg3
+expect "106 a change branch origin does not hold yet" allow \
+  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $CWT push -u origin spectre/chg3")}")"
+[ ! -e "$UNTASKED_LOG" ] && pass "106 checker not run" || fail "106 checker ran"
+expect "107 a checker that cannot run fails open" allow \
+  "$(UNTASKED_COMMITS_CHECK="$ROOT/missing" run "$ROOT" Bash "{\"command\":$(q "git -C $CWT push origin HEAD:spectre/chg2")}")"
+unset UNTASKED_COMMITS_CHECK
+
 # 3 last: moving the main checkout off main lifts the protection
 git_q -C "$MAIN" checkout -q -b feature
 expect "3 Write in main checkout on a feature branch" allow \
diff --git a/skills/flow-contracts/git-boundaries.md b/skills/flow-contracts/git-boundaries.md
index 12736743..c75b035b 100644
--- a/skills/flow-contracts/git-boundaries.md
+++ b/skills/flow-contracts/git-boundaries.md
@@ -134,7 +134,14 @@ records no such row, the stage key (`flow.review-panel`, `flow.visual-verify`, `
 carries `Task-Id:` instead, and a planning commit neither. Run 1's unfinished-work gate lists
 every other non-`chore` commit with `check-untasked-commits.sh`
 (`skills/flow-contracts/finish-contract-run1.md`).
-**Before such a commit is pushed, `check-untasked-commits.sh <worktree> <merge-base>` runs**; when the only sha it lists is that commit's, `git commit --amend` adds the missing trailer and the guard runs again; any other listed commit goes to run 1's unfinished-work gate, which handles `UNTASKED`.
+**Before such a commit is pushed, `check-untasked-commits.sh` judges every commit the push sends**,
+mechanically at both points that push a change branch mid-run: `hooks/protect-main-checkout.py`
+denies a typed push of HEAD to `spectre/<name>` while it lists a commit origin's copy of that branch lacks,
+and `close-task.sh` refuses (exit 1, nothing ticked or pushed) on the same range per worktree
+before its push. A listed HEAD takes the trailer through `git commit --amend`, an earlier commit
+through a reword; the push or the close is then made again. A commit already on origin is not
+judged there — run 1's unfinished-work gate lists it and handles `UNTASKED` (kan-941 pushed four
+such commits before either check existed).
 
 ## Revert and back-out commits
 
diff --git a/stats/internal/guard/close_task_test.go b/stats/internal/guard/close_task_test.go
index d71045b5..a6fc7084 100644
--- a/stats/internal/guard/close_task_test.go
+++ b/stats/internal/guard/close_task_test.go
@@ -159,6 +159,32 @@ func TestCloseTaskOrder(t *testing.T) {
 		}
 		ctEqual(t, fx.calls(t), nil)
 	})
+	t.Run("untasked unpushed commit: exit 1, no tick, no push", func(t *testing.T) {
+		t.Parallel()
+		fx := ctFixture(t, true, "0", "0")
+		gitRun(t, fx.wt, "push", "-q", "-u", "origin", "spectre/demo")
+		writeFile(t, fx.wt+"/m.txt", "m\n")
+		gitRun(t, fx.wt, "add", "--", "m.txt")
+		gitRun(t, fx.wt, "commit", "-q", "-m", "fix(app): review Minors")
+		s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
+		r := fx.run(fx.wt, "demo", fx.base, "1:"+s1)
+		if r.rc != 1 || !strings.Contains(r.stdout, "UNTASKED: 1 commit(s) carry no Task-Id: or Flow-Key: trailer") {
+			t.Fatalf("exit %d, want 1 and the UNTASKED verdict\n%s", r.rc, r.out)
+		}
+		ctEqual(t, fx.calls(t), nil)
+	})
+	t.Run("a Flow-Key unpushed commit pushes", func(t *testing.T) {
+		t.Parallel()
+		fx := ctFixture(t, true, "0", "0")
+		gitRun(t, fx.wt, "push", "-q", "-u", "origin", "spectre/demo")
+		writeFile(t, fx.wt+"/m.txt", "m\n")
+		gitRun(t, fx.wt, "add", "--", "m.txt")
+		gitRun(t, fx.wt, "commit", "-q", "-m", "fix(app): review Minors", "-m", "Flow-Key: panel-fix-1")
+		s1 := fx.commit(t, fx.wt, "one", "1", "a.txt", "a\n")
+		if r := fx.run(fx.wt, "demo", fx.base, "1:"+s1); r.rc != 0 {
+			t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
+		}
+	})
 	t.Run("undeclared path from a refusal fires the gate", func(t *testing.T) {
 		t.Parallel()
 		fx := ctFixture(t, true, "0", "0")
diff --git a/stats/internal/guard/closetask.go b/stats/internal/guard/closetask.go
index c4931a55..dc3e79a7 100644
--- a/stats/internal/guard/closetask.go
+++ b/stats/internal/guard/closetask.go
@@ -119,6 +119,16 @@ func closeTask(args []string, env Env, stdout, stderr io.Writer) (exit int) {
 		worst = max(worst, ctCheckTrailer(env, stdout, stderr, canonical, name, t.id, t.commit))
 	}
 	worst = max(worst, Registry["check-task-commit-planning-paths"]([]string{canonical, mergeBase}, env, stdout, stderr))
+	// The push below sends every commit origin lacks: one outside every task
+	// needs its Flow-Key: trailer before it leaves (git-boundaries.md,
+	// Commits outside every task; kan-941 pushed four with none). A worktree
+	// whose branch origin does not hold yet has nothing to judge.
+	for _, wt := range worktrees {
+		remote := "refs/remotes/origin/spectre/" + name
+		if _, ok := capture(envGit(env)("-C", wt, "rev-parse", "--verify", "-q", remote)); ok {
+			worst = max(worst, Registry["check-untasked-commits"]([]string{wt, remote}, env, stdout, stderr))
+		}
+	}
 	switch worst {
 	case 0:
 	case 1:
````

  - [x] **Step 3: Verify.** `bash scripts/test-protect-main-checkout.sh` prints `all protect-main-checkout cases pass`; `cd stats && go test ./internal/guard/ -run 'TestCloseTask' -count=1` passes; `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go vet ./internal/guard/` exits 0; `scripts/check-python-suppressions.sh`, `scripts/check-installed-citations.sh`, `scripts/check-verbatim-moves.sh` and `scripts/check-markdown-integrity.py` exit 0. Commit.

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `scripts/test-protect-main-checkout.sh`, `stats/internal/guard/closetask.go`, `stats/internal/guard/close_task_test.go`, `scripts/close-task.sh`, `skills/flow-contracts/git-boundaries.md`, `scripts/check-untasked-commits.sh`, `stats/internal/guard/untaskedcommits.go`
**Tests:** `104 a change-branch push carrying an untasked commit`, `105 a change-branch push the checker clears`, `106 a change branch origin does not hold yet`, `107 a checker that cannot run fails open`, `untasked unpushed commit: exit 1, no tick, no push`, `a Flow-Key unpushed commit pushes`
**Regression:** reverting `untasked_push` lets case 104's push through (`wanted deny, got allow`); reverting close-task's `check-untasked-commits` loop pushes the trailerless commit, so `untasked unpushed commit: exit 1, no tick, no push` fails; cases 105–107 and `a Flow-Key unpushed commit pushes` fail if either check refuses a push it must let through.
**Baseline:** before=99 after=105
<!-- measured: cat scripts/test-protect-main-checkout.sh stats/internal/guard/close_task_test.go | grep -cE '^expect |t\.Run\("' @ f4844c40 (before; after with this task's diff applied) -->
**Commit:** fix(hooks): refuse a change-branch push carrying a commit with no task or Flow-Key trailer
**After:** Task 2, 4

**Decision:** flow-key-at-push
