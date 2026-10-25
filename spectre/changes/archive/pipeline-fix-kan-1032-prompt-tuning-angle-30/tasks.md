# pipeline-fix-kan-1032-prompt-tuning-angle-30

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** `/flow` and `/flow-fast` rebase the change branch onto its base before task 1 of every
implementation run, stopping for the operator on a conflict.

**Architecture:** `skills/flow/implement.md` section 1 gains a step running
`sync-panel-base.sh --rebase` per worktree after the plan refresh, recording the new merge base
and pushing `--force-with-lease`; `/flow-fast` section 4 cites it; **Branch backup** and
`refresh-plan-base.sh`'s header follow; a repotest test pins it.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/` — this change's `verbatim-moves.txt` already lists
  every sentence the diff rewords or adds.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; it was applied to this worktree at dd2268bd and checked there. A diff that does not
  apply is reported, never hand-merged by guesswork.
- No new inline suppressions; `gofmt -l .` in `stats/` prints nothing.
- Never stop flowd or flow-postgres, never drop the `flow` DB.

## Review Focus

- The step runs before task 1 on every implementation run and after the plan refresh
  (`after-plan-refresh`).
- A conflict is never resolved and stops the run (`stop-on-conflict-and-unsettled`).
- The rewritten branch is recorded and pushed `--force-with-lease` (`record-and-force-push`).
- `/flow-fast`'s `<merge-base>` follows the rebase (`flow-fast-same-step`).

- [x] 1. Rebase the change branch onto its base before task 1
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Save the diff below to
    `/tmp/pf30-task1.diff`, apply its test hunk alone
    (`git apply --include='stats/*' /tmp/pf30-task1.diff`) and run
    `cd stats && go test ./internal/repotest -run TestBaseSyncBeforeTaskOne -count=1`: FAIL, with
    `skills/flow/implement.md section 1 lacks` lines, `still defers the rebase to integrate`, and
    `skills/flow-fast/SKILL.md section 4 lacks`.
  - [x] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/*' /tmp/pf30-task1.diff`):

````diff verified:applied with git apply to this worktree at dd2268bd; TestBaseSyncBeforeTaskOne passes with it and fails with the skill edits reverted; check-references, check-verbatim-moves, check-markdown-integrity, check-guard-symlinks, check-dispatch-paragraphs, check-installed-citations, check-normative-inventory, check-ask-silence, check-hand-notes-in-step, gofmt -l and go vet exit 0
diff --git a/scripts/refresh-plan-base.sh b/scripts/refresh-plan-base.sh
index a9678267..5fc28ff6 100755
--- a/scripts/refresh-plan-base.sh
+++ b/scripts/refresh-plan-base.sh
@@ -28,7 +28,8 @@
 # answer — usage, a merge base that is not a commit, an unreadable
 # <tasks.md>, or a base branch resolve-base-branch cannot resolve (its own
 # reason precedes this guard's on stderr). It names, it never rebases: the
-# change branch is synced onto the base at integrate.
+# change branch is rebased onto the base right after it, by sync-panel-base.sh
+# --rebase (skills/flow/implement.md).
 #
 # flow-guard is built from this checkout, never taken from PATH:
 # scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
diff --git a/skills/flow-contracts/git-boundaries.md b/skills/flow-contracts/git-boundaries.md
index 46ffc4dc..ab9bc1d4 100644
--- a/skills/flow-contracts/git-boundaries.md
+++ b/skills/flow-contracts/git-boundaries.md
@@ -86,7 +86,9 @@ call that made or verified it. The one deferred push is a visual-verify commit i
 origin/<branch>`. The one rewrite is integrate's reshape (`reshape-branch.sh`, then
 the two commits), so that run's push is `git push --force-with-lease origin <branch>`, made once the
 archive commit lands and before the self-review pass, so a run that stops before its route still
-leaves the remote on the rewritten branch; every other push is plain.
+leaves the remote on the rewritten branch; every other push is plain, save a branch a base-sync
+rebase rewrote, which pushes `--force-with-lease` (the base-sync step of **1. Load context and
+validate the plan**, `skills/flow/implement.md`).
 
 **The planning path** is the one **Handoff output** (`skills/flow-contracts/pipeline.md`) names.
 `/flow`'s implement phase clears it from the index and only then stages with it excluded by
diff --git a/skills/flow-fast/SKILL.md b/skills/flow-fast/SKILL.md
index fa3dd092..3f01fbf7 100644
--- a/skills/flow-fast/SKILL.md
+++ b/skills/flow-fast/SKILL.md
@@ -70,7 +70,8 @@ basename is what keys the rolls — and `<project>/.superpowers/` is listed in
 `<project>/.git/info/exclude`. Every `skills/flow/` section cited below runs as written, with
 these substitutions and no others: `<changeRoot>` as above; `ff-<literal-token>` wherever it
 reads `mf-<literal-token>`; the one worktree as the resolved set, its `<merge-base>` the sha
-`git -C <worktree> rev-parse origin/<base>` prints right after section 3 creates it,
+`git -C <worktree> rev-parse origin/<base>` prints right after section 3 creates it — or, once
+section 4's base sync rebased the branch, the sha its last `REBASED:` line names —
 standing wherever a cited section reads the working notes' merge base; and nothing written to
 `<project>/spectre/` or a state file — `flow record render` is
 skipped, the finding rows are the record. The scripts a cited section calls run as it says.
@@ -214,6 +215,10 @@ flow stage begin -command '/flow-fast' -stage flow.load-context -harness <harnes
 Compose the learning note here per **Learning** (`skills/flow-contracts/learning.md`), write its
 file and print its absolute path; the final report repeats the path.
 
+The branch is then rebased onto its base before task 1 runs, as the base-sync step of **1. Load
+context and validate the plan** (`skills/flow/implement.md`) states, with no `flow state
+add-worktree` call and `<name>` as the pushed branch.
+
 The project's instruction file is already in this session's context (`<project>/CLAUDE.md` on
 Claude Code, `<project>/AGENTS.md` on ZCode and Muse); never read either one here. Read
 `<project>/.flow/project.md`'s `## lint`, `## test`, `## handoff` and `## default landing route` sections, the
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index 1e7e0b5d..b474b498 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -267,9 +267,29 @@ the plan's `**Files:**` declares and each spec's text at the moved base (`SPEC-A
 gone there); exit 2 stops the run. The script's header (`<agents repo>/scripts/refresh-plan-base.sh`)
 is canonical for its output. On `MOVED:`, **name every route or requirement the plan adds that the
 moved base now already carries**. A collision is reconciled before task 1 runs, under **A pivot
-reconciles the three artifacts together** (section 4 below) — never discovered mid-task. The step
-names, it never rebases: the change branch is synced onto the base at integrate (**Sync the branch
-onto the base**, `skills/flow/sync-onto-base.md`).
+reconciles the three artifacts together** (section 4 below) — never discovered mid-task.
+
+**The change branch is rebased onto its base before task 1 runs**, on every implementation run —
+creating, resumed or fix — once the plan refresh above has named what moved (kan-1032: the base
+moved 13 commits during planning, and the overlap surfaced only at the review panel's base sync,
+after every task). Once per worktree in this run's resolved set, with its working-notes merge base:
+
+```bash
+sync-panel-base.sh --rebase <worktree> <merge-base>
+```
+
+The script's header (`<agents repo>/scripts/sync-panel-base.sh`) is canonical for its lines. Exit 0
+is settled. After a `REBASED:` line, that worktree's working-notes merge base becomes the sha the
+last `REBASED:` line names, and the run records it and pushes the rewritten branch:
+
+```bash
+flow state add-worktree <name> <worktree> <sha>
+git -C <worktree> push --force-with-lease origin spectre/<name>
+```
+
+Exit 1 is a `CONFLICT`, handled as the **Conflict** bullet of **Check base movement first**
+(`skills/flow/review-panel.md`) states, and the run stops for the operator. Exit 2 or 3 stops the
+run with its output named.
 
 **Whether there is anything left to implement is read off the task checkboxes**, from
 `spectre list --json`'s `{"changes":[{"id","done","total"}]}` for this change. A change already
diff --git a/stats/internal/repotest/base_sync_before_task_one_test.go b/stats/internal/repotest/base_sync_before_task_one_test.go
new file mode 100644
index 00000000..c80a4955
--- /dev/null
+++ b/stats/internal/repotest/base_sync_before_task_one_test.go
@@ -0,0 +1,57 @@
+package repotest
+
+// TestBaseSyncBeforeTaskOne pins the rebase onto the base that /flow's and
+// /flow-fast's implementation runs make before task 1 (kan-1032: the base moved
+// 13 commits during planning, and only the review panel's base sync, after
+// every task, met the overlap).
+
+import (
+	"os"
+	"path/filepath"
+	"strings"
+	"testing"
+)
+
+// section is the text of the "## <heading>" section of a Markdown file, up to
+// the next level-2 heading.
+func section(t *testing.T, rel, heading string) string {
+	t.Helper()
+	b, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
+	if err != nil {
+		t.Fatal(err)
+	}
+	s := string(b)
+	i := strings.Index(s, "\n## "+heading)
+	if i < 0 {
+		t.Fatalf("%s: no %q section", rel, "## "+heading)
+	}
+	s = s[i+1:]
+	if j := strings.Index(s[3:], "\n## "); j >= 0 {
+		s = s[:j+3]
+	}
+	return s
+}
+
+func TestBaseSyncBeforeTaskOne(t *testing.T) {
+	t.Parallel()
+	const step = "**The change branch is rebased onto its base before task 1 runs**"
+	impl := section(t, "skills/flow/implement.md", "1. Load context and validate the plan")
+	for _, want := range []string{
+		step,
+		"sync-panel-base.sh --rebase <worktree> <merge-base>",
+		"flow state add-worktree <name> <worktree> <sha>",
+		"git -C <worktree> push --force-with-lease origin spectre/<name>",
+	} {
+		if !strings.Contains(impl, want) {
+			t.Errorf("skills/flow/implement.md section 1 lacks %q", want)
+		}
+	}
+	if strings.Contains(impl, "it never rebases") {
+		t.Errorf("skills/flow/implement.md section 1 still defers the rebase to integrate")
+	}
+	const cite = "rebased onto its base before task 1 runs, as the base-sync step of **1. Load context and validate the plan** (`skills/flow/implement.md`)"
+	fast := strings.Join(strings.Fields(section(t, "skills/flow-fast/SKILL.md", "4. Implement")), " ")
+	if !strings.Contains(fast, cite) {
+		t.Errorf("skills/flow-fast/SKILL.md section 4 lacks %q", cite)
+	}
+}
````

  - [x] **Step 3: Verify.** `cd stats && go test ./internal/repotest -run TestBaseSyncBeforeTaskOne
    -count=1` passes; `scripts/check-references.sh`, `scripts/check-verbatim-moves.sh`,
    `scripts/check-markdown-integrity.sh`, `scripts/check-dispatch-paragraphs.sh`,
    `scripts/check-installed-citations.sh`, `scripts/check-normative-inventory.sh` exit 0;
    `cd stats && gofmt -l . && go vet ./internal/repotest` prints nothing.
  - [x] **Step 4: Commit** the five files with the subject below.

**Build:** green
**Files:** `skills/flow/implement.md`, `skills/flow-fast/SKILL.md`, `skills/flow-contracts/git-boundaries.md`, `scripts/refresh-plan-base.sh`, `stats/internal/repotest/base_sync_before_task_one_test.go`
**Tests:** `TestBaseSyncBeforeTaskOne` in `stats/internal/repotest/base_sync_before_task_one_test.go`
**Regression:** `TestBaseSyncBeforeTaskOne` — reverting the skill edits fails it: `skills/flow/implement.md section 1 still defers the rebase to integrate`.
**Commit:** feat(flow): rebase the change branch onto its base before task 1
**After:** none

**Decision:** reuse-sync-panel-base-rebase
**Decision:** after-plan-refresh
**Decision:** stop-on-conflict-and-unsettled
**Decision:** record-and-force-push
**Decision:** flow-fast-same-step
**Decision:** doc-pin-test
