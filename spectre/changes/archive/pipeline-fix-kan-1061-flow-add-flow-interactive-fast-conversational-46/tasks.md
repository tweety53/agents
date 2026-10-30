# pipeline-fix-kan-1061-flow-add-flow-interactive-fast-conversational-46

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** the `flow` skill ships `check-plan-provenance.sh`, its `implement.md` invocation is one
`check-guard-symlinks.sh` detects, and the shim works when called through the symlink.

**Architecture:** a symlink in `skills/flow/scripts/`, a reworded sentence in
`skills/flow/implement.md`, and the shim's root derived by `flow_guard_root`.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; it was applied to this worktree at f3cf9bd9 and checked there. A diff that does not
  apply is reported, never hand-merged by guesswork.
- No new inline suppressions; `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits
  clean.
- Read-only against the dev store: never stop flowd or flow-postgres, never drop the `flow` DB.

## Review Focus

- The symlink and the reworded sentence land together: either alone makes
  `check-guard-symlinks.sh` exit 1 (rule 2 without the symlink, rule 6 without the wording).
- `skills/flow/scripts/check-plan-provenance.sh` run from the worktree root scans the repository,
  not `skills/flow`.

- [x] 1. Ship check-plan-provenance.sh in the flow skill
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Save the diff below to
    `/tmp/pf46-task1.diff` and apply its `implement.md` hunk alone
    (`git apply --include='skills/flow/implement.md' /tmp/pf46-task1.diff`): `scripts/check-guard-symlinks.sh`
    exits 1 with `invokes check-plan-provenance.sh here, but skill "flow" carries no symlink (rule 2)`.
  - [x] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='skills/flow/implement.md' /tmp/pf46-task1.diff`):

````diff verified:applied with git apply to this worktree at f3cf9bd9; check-guard-symlinks exits 0 with it and 1 with the symlink removed; skills/flow/scripts/check-plan-provenance.sh exits 0 with it and 2 with the shim hunk reverted; go test ./internal/guard passes; check-references and check-dispatch-paragraphs exit 0
diff --git a/scripts/check-plan-provenance.sh b/scripts/check-plan-provenance.sh
index d700dbf9..7613ef61 100755
--- a/scripts/check-plan-provenance.sh
+++ b/scripts/check-plan-provenance.sh
@@ -40,7 +40,9 @@
 #
 # The default root is resolved here. The Python guard took REPO_ROOT from
 # its own file's location; flow-guard runs from a build cache, so this shim
-# resolves the checkout from its own location instead and hands it over in
+# resolves the checkout through flow_guard_root instead — a call through a
+# skill's scripts/ symlink answers the repository, never the skill
+# directory — and hands it over in
 # CHECK_PLAN_PROVENANCE_ROOT, only when that is unset — an explicit value,
 # empty included, reaches the guard untouched.
 set -euo pipefail
@@ -49,7 +51,7 @@ set -euo pipefail
   exit 2
 }
 flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
-root="$(cd -P -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)" || {
+root="$(flow_guard_root)" || {
   echo "check-plan-provenance: cannot resolve the checkout beside ${BASH_SOURCE[0]}" >&2
   exit 2
 }
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index efaea3fc..2aa3ebbe 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -896,8 +896,8 @@ the plan did not name — the parent appends a paragraph opening `Correction (YY
 that task's entry in `tasks.md`, at the task-close boundary where the task's other corrections
 are transcribed (before the guard runs on the commit), stating what the plan declared, what
 shipped instead, and why. The disclosure route is the same as any correction's: the implementer
-reports the deviation, the parent transcribes it; inline, the session is both halves. A `Correction:` paragraph is plan prose `check-plan-provenance.sh` reads: run that guard when
-transcribing it, and give each numeric claim its `measured:` comment or word it without a count
+reports the deviation, the parent transcribes it; inline, the session is both halves. A `Correction:` paragraph is plan prose the plan-provenance guard reads:
+Run `check-plan-provenance.sh` when transcribing it, and give each numeric claim its `measured:` comment or word it without a count
 (kan-1020: a bare count in a Correction was refused only at verify). Once a `Correction:`
 paragraph is transcribed, the parent runs `flow tasks count -C <worktree> <change>`, which
 records the plan's correction count beside its task count — the plan-corrections metric
diff --git a/skills/flow/scripts/check-plan-provenance.sh b/skills/flow/scripts/check-plan-provenance.sh
new file mode 120000
index 00000000..1dfa6318
--- /dev/null
+++ b/skills/flow/scripts/check-plan-provenance.sh
@@ -0,0 +1 @@
+../../../scripts/check-plan-provenance.sh
\ No newline at end of file
````

  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean and
    `go test ./internal/guard -run 'TestCheckGuardSymlinks|TestCheckPlanProvenance'` passes;
    `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh` and
    `scripts/check-dispatch-paragraphs.sh` exit 0; `skills/flow/scripts/check-plan-provenance.sh`
    run from the worktree root exits 0 with `all provenance stated`.
  - [x] **Step 4: Commit** the three files with the subject below.

**Build:** green
**Files:** `scripts/check-plan-provenance.sh`, `skills/flow/implement.md`, `skills/flow/scripts/check-plan-provenance.sh`
**Tests:** `TestCheckGuardSymlinks`, `TestCheckPlanProvenance`
**Regression:** TestCheckGuardSymlinks — its own-tree case runs check-guard-symlinks.sh on this repository, which exits 1 if the symlink is removed while the sentence invokes the guard.
**Baseline:** before=0 after=0
<!-- measured: grep -cE '^func Test' stats/internal/guard/check_guard_symlinks_test.go @ f3cf9bd9 → unchanged by this task, which adds no test function; before=0 after=0 counts added tests -->
**Commit:** fix(scripts): ship check-plan-provenance.sh in the flow skill
**After:** none

**Decision:** ship-provenance-symlink
