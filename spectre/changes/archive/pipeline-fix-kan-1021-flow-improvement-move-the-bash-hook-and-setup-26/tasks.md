# pipeline-fix-kan-1021-flow-improvement-move-the-bash-hook-and-setup-26

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** integrate run 1's archived re-run base runs past every self-review loop commit, so a run 1
resumed mid-loop never folds the loop's fix commits into a new implementation commit.

**Architecture:** `archBase` in `stats/internal/guard/integratechange.go` asks a new
`selfReviewBase` first, which walks `<merge-base>..HEAD` oldest-first over the loop's commits;
the contract sentence in `skills/flow-contracts/finish-contract-run1.md` and its rationale follow.

**Spec:** `design.md` of this change.

## Global constraints

- Every task runs in
  `/Users/tweety53/Projects/agents-worktrees/pipeline-fix-kan-1021-flow-improvement-move-the-bash-hook-and-setup-26`; every
  `**Files:**` path is relative to it. Never touch the main checkout and never create a worktree.
- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/` — its `verbatim-moves.txt` already lists the
  contract sentences this change rewords.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; it was applied to this worktree at e77f951b and checked there. A diff that does not
  apply is reported, never hand-merged by guesswork.
- Never stop flowd or flow-postgres, never drop the `flow` DB; the test runs on its own fixtures.

## Review Focus

- A branch with no self-review loop commit keeps today's base: `archive <name>`, else the merge
  base (`self-review-loop-base`); every existing `TestIntegrateChange*` case unchanged.
- A fix run's `chore(spectre): plan <name>` commit, bare or suffixed, ends the loop, so its work
  stays new work (`self-review-loop-base`).
- The contract sentence and the code state the same base.

- [x] 1. Run an archived re-run's reshape base past the self-review loop
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Save the diff below to
    `/tmp/pf26-task1.diff`, apply its test hunk alone
    (`git apply --include='stats/internal/guard/integrate_change_prepare_test.go' /tmp/pf26-task1.diff`)
    and from `stats/` run `make -s web-dist && go test ./internal/guard -run
    'TestIntegrateChangePrepareCommit/archived_rerun_self_review'`: FAIL, output carrying
    `ARCHIVED: demo — with new work` and `STOP: unfinished-work`.
  - [x] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/internal/guard/integrate_change_prepare_test.go' /tmp/pf26-task1.diff`):

````diff verified:applied with git apply to this worktree at e77f951b; the new case fails with integratechange.go untouched and passes with it; go test -run TestIntegrateChange, go vet ./..., gofmt -l ., check-references, check-markdown-integrity, check-normative-inventory and check-verbatim-moves (with this change's verbatim-moves.txt) exit clean
diff --git a/skills/flow-contracts/finish-contract-rationale.md b/skills/flow-contracts/finish-contract-rationale.md
index 0d58a6de..b63a261c 100644
--- a/skills/flow-contracts/finish-contract-rationale.md
+++ b/skills/flow-contracts/finish-contract-rationale.md
@@ -56,6 +56,15 @@ the archive commit the re-run's base and hide that work. Archived-ness is read f
 canonical repository alone because only it holds the change directory; a satellite worktree has
 nothing to judge it by.
 
+The base runs past the self-review loop's own commits, not only its report, because a run 1
+resumed mid-loop — after `chore(spectre): plan self-review-<name>` and its fix commits, before the
+report — otherwise took `chore(spectre): archive <name>` as its base, folded the loop's fix
+commits into a new implementation commit and lost their shas, and carried the loop's plan files
+under `chore(spectre): plan <name>` (kan-1021). A fix commit carries no marker subject, so the
+loop is every commit from its plan commit on; a fix run's own planning commit,
+`chore(spectre): plan <name>` with or without its ` — <stage>` suffix, ends it, since a fix run
+commits that before its first task and its work is new work.
+
 The merge-and-push route pushes `spectre/<name>` before `<base>`, so `origin/spectre/<name>` never
 stays at its pre-reshape tip. Run 2's `git branch -d` does not depend on it: `remove-change-worktrees.sh`
 points the branch's upstream at `origin/<base>` first, because `-d` refuses a branch not merged into
diff --git a/skills/flow-contracts/finish-contract-run1.md b/skills/flow-contracts/finish-contract-run1.md
index b1db0a8e..3cab9cd7 100644
--- a/skills/flow-contracts/finish-contract-run1.md
+++ b/skills/flow-contracts/finish-contract-run1.md
@@ -128,13 +128,16 @@ worktree's `<project>/spectre/changes/archive/<name>/` exists and its
 somewhere after it, and any fix run since wrote into the archived directory (**A change's
 directory**, `skills/flow-contracts/pipeline.md`).
 
-**An archived re-run's base** is the newest commit in `<merge-base>..HEAD` whose subject is
-`docs(self-review): <name> self-review report` — or, from a run before KAN-927,
-`docs(self-review): <name> self-review context bundle` — else the newest whose subject is
-`chore(spectre): archive <name>`, else `<merge-base>` itself — a satellite worktree whose run 1
+**An archived re-run's base** is the newest commit of the self-review loop in
+`<merge-base>..HEAD`: its `docs(self-review): <name> self-review report` commit — or, from a run
+before KAN-927, `docs(self-review): <name> self-review context bundle` — its
+`chore(spectre): archive self-review-<name>` commit, or, from its
+`chore(spectre): plan self-review-<name>` commit on, every commit above that plan up to a fix
+run's first `chore(spectre): plan <name>` planning commit, which opens new work. Else it is the
+newest whose subject is `chore(spectre): archive <name>`, else `<merge-base>` itself — a satellite worktree whose run 1
 predates KAN-815 holds neither — `<merge-base>` being the merge base the reshape below would otherwise take. Everything at
-or below it is kept untouched — a reshape from that merge base itself would fold the archive and
-report commits into the implementation commit. The re-run **carries new work** when a commit sits
+or below it is kept untouched — a reshape from that merge base itself would fold the archive, the
+self-review loop's plan and fix commits and the report into the implementation commit. The re-run **carries new work** when a commit sits
 above the canonical worktree's base — every fix run commits its plan there, whichever repositories
 its tasks touched — or any worktree in the set holds an uncommitted tracked change, judged at the
 unfinished-work gate, before this run's own narrative append.
diff --git a/stats/internal/guard/integrate_change_prepare_test.go b/stats/internal/guard/integrate_change_prepare_test.go
index b50cbecd..da62ab3a 100644
--- a/stats/internal/guard/integrate_change_prepare_test.go
+++ b/stats/internal/guard/integrate_change_prepare_test.go
@@ -73,6 +73,20 @@ g -C "$WORK/a-worktrees/demo" add more.txt && g -C "$WORK/a-worktrees/demo" comm
 sut prepare "$WORK/a" demo
 check "archived_rerun: prepare passes the gate" '[ "$RC" -eq 0 ]'
 check "archived_rerun: b's reshaped commit is not listed" '! printf "%s" "$OUT" | grep -q "feat(b): do the work"'`},
+		{"archived_rerun_self_review", `# --- archived_rerun_self_review: the self-review loop's commits sit below the base ---
+fixture
+to_land "archived_rerun_self_review: " no-report
+A="$WORK/a-worktrees/demo"
+mkdir -p "$A/spectre/changes/self-review-demo"
+printf '# tasks\n' >"$A/spectre/changes/self-review-demo/tasks.md"
+g -C "$A" add spectre/changes/self-review-demo && g -C "$A" commit -qm "chore(spectre): plan self-review-demo"
+printf 'fix\n' >"$A/sr-fix.txt"
+g -C "$A" add sr-fix.txt && g -C "$A" commit -qm "fix(a): a self-review fix"
+sut prepare "$WORK/a" demo
+check "archived_rerun_self_review: a resumed run 1 holds no new work" 'printf "%s\n" "$OUT" | grep -q "^ARCHIVED: demo — without new work"'
+g -C "$A" commit -q --allow-empty -m "chore(spectre): plan demo — fix-run tasks"
+sut prepare "$WORK/a" demo
+check "archived_rerun_self_review: a fix run's plan commit above the loop is new work" 'printf "%s\n" "$OUT" | grep -q "^ARCHIVED: demo — with new work"'`},
 		{"no_verdict", `# --- no_verdict ---
 fixture
 : >"$WORK/fail-findings"
diff --git a/stats/internal/guard/integratechange.go b/stats/internal/guard/integratechange.go
index ac3f7190..ab2fb158 100644
--- a/stats/internal/guard/integratechange.go
+++ b/stats/internal/guard/integratechange.go
@@ -791,19 +791,47 @@ func (c *ic) baseOf(i int) {
 	}
 }
 
-// archBase: an archived re-run's base.
+// archBase: an archived re-run's base — past the self-review loop's
+// commits when it ran, else the archive commit, else mb.
 func (c *ic) archBase(w, mb string) string {
 	log, _ := c.git("-C", w, "log", "--format=%H %s", mb+"..HEAD")
-	if s := icFirstSubject(log, "docs(self-review): "+c.name+" self-review report", "docs(self-review): "+c.name+" self-review context bundle"); s != "" {
+	if s := c.selfReviewBase(log); s != "" {
 		return s
 	}
-	log, _ = c.git("-C", w, "log", "--format=%H %s", mb+"..HEAD")
 	if s := icFirstSubject(log, "chore(spectre): archive "+c.name); s != "" {
 		return s
 	}
 	return mb
 }
 
+// selfReviewBase: the newest commit of the self-review loop in log (newest
+// first) — its report or bundle, its archive, or, from its plan commit on,
+// every commit up to a fix run's own planning commit, which opens new work.
+func (c *ic) selfReviewBase(log string) string {
+	plan := "chore(spectre): plan " + c.name
+	marks := []string{
+		"docs(self-review): " + c.name + " self-review report",
+		"docs(self-review): " + c.name + " self-review context bundle",
+		"chore(spectre): archive self-review-" + c.name,
+	}
+	base, loop := "", false
+	l := icLines(log)
+	for i := len(l) - 1; i >= 0; i-- {
+		h, s, _ := strings.Cut(l[i], " ")
+		switch {
+		case s == "chore(spectre): plan self-review-"+c.name:
+			base, loop = h, true
+		case icHas(marks, s):
+			base, loop = h, false
+		case s == plan || strings.HasPrefix(s, plan+" — "):
+			loop = false
+		case loop:
+			base = h
+		}
+	}
+	return base
+}
+
 // icFirstSubject: the hash of the first `<hash> <subject>` line whose
 // subject is one of subjects.
 func icFirstSubject(log string, subjects ...string) string {
````

  - [x] **Step 3: Verify.** From `stats/`: `gofmt -l .` prints nothing; `make -s web-dist && go vet
    ./...` exits 0; `go test ./internal/guard -run 'TestIntegrateChange'` exits 0. From the
    worktree root: `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`,
    `scripts/check-normative-inventory.sh` and `scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 4: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/integratechange.go`, `stats/internal/guard/integrate_change_prepare_test.go`, `skills/flow-contracts/finish-contract-run1.md`, `skills/flow-contracts/finish-contract-rationale.md`
**Tests:** `archived_rerun_self_review`
**Regression:** `archived_rerun_self_review` — reverting the commit makes a run 1 resumed above the self-review loop's plan and fix commits report `ARCHIVED: demo — with new work`, so its reshape would fold the loop's fix commits into a new implementation commit.
**Commit:** fix(flow-guard): run an archived re-run's reshape base past the self-review loop
**After:** none

**Decision:** self-review-loop-base
**Decision:** prepare-case-test
