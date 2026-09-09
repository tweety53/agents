# kan-260-guards-are-blind-to-cross-repo-changes

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

- [x] 1. change-plan.sh resolves a same-named plan from the canonical worktree when the local
change directory is absent entirely
**Build:** green
**Files:** `scripts/lib/change-plan.sh`, `scripts/test-lib-change-plan.sh`
**Tests:** `absent local dir resolves the canonical worktree's same-named plan`, `absent local dir without a canonical worktree stays unresolvable`, `satellite dir with link.md and no tasks.md never takes the absent-dir branch`
**Regression:** reverting this commit makes `check-unfinished-work.sh` report `OUTSTANDING: no plan at …` and `check-task-commit-fields.sh` exit 2 for a cross-repo worktree holding no change dir, even when the canonical worktree is supplied.
**Baseline:** before=32 after=40
<!-- measured: bash scripts/test-lib-change-plan.sh @ branch spectre/kan-260-guards-are-blind-to-cross-repo-changes -->
<!-- measured: bash scripts/test-lib-change-plan.sh after task 1 @ branch spectre/kan-260-guards-are-blind-to-cross-repo-changes -->
**Commit:** fix(scripts): resolve same-named plan from canonical worktree when local change dir absent

  - [x] **Step 1:** in `_change_plan_resolve_dir`, between the local `tasks.md` check and the
    `link.md` branch: when `[ ! -d "$dir" ]` and a canonical worktree was supplied and
    `<canonical>/<spec-root>/changes/<name>/tasks.md` exists (spec root via
    `spec_root_leaf "$canonical_worktree"`), print that dir and return 0.
  - [x] **Step 2:** update the library header's resolution-order comment (new branch between
    today's 1 and 2, gated on the absent local dir, with the
    `no-fallback-for-linked-satellites` rationale in one line).
  - [x] **Step 3:** add the three harness cases above as real two-repo mktemp fixtures per the
    harness header's own convention; run `bash scripts/test-lib-change-plan.sh` clean.

- [x] 2. check-task-commit-fields.sh reaches a verdict when the plan lives only in the canonical
worktree
**Build:** green
**Files:** `scripts/check-task-commit-fields.sh`, `scripts/test-check-task-commit-fields.sh`
**Tests:** `named change with no local dir resolves through the canonical worktree`, `named change with no local dir and no canonical worktree still refuses`, `no change-name argument and no local dir still refuses through the glob path`
**Regression:** reverting this commit returns the named-change path to `no tasks.md found for change '<name>'` exit 2 on a cross-repo worktree with no change dir, so per-task field verification silently never runs there again.
**Baseline:** before=199 after=204
<!-- measured: bash scripts/test-check-task-commit-fields.sh @ branch spectre/kan-260-guards-are-blind-to-cross-repo-changes -->
<!-- measured: bash scripts/test-check-task-commit-fields.sh after task 2 @ branch spectre/kan-260-guards-are-blind-to-cross-repo-changes -->
**Commit:** fix(scripts): task-commit-fields verifies a plan living only in the canonical worktree

  - [x] **Step 1:** in the named-change block, when `$ROOT_DIR` does not exist at all, resolve
    `change_plan_path "$WORKTREE" "$CHANGE_NAME" "$CANONICAL_WORKTREE"` before the refusal; keep
    the refusal for every shape the lib cannot resolve.
  - [x] **Step 2:** extend the wrapper header's WHERE-A-LINK-IS-FOLLOWED narration with the
    absent-dir shape in one paragraph.
  - [x] **Step 3:** add the three harness cases above as real two-repo fixtures (canonical repo
    with the plan, second repo's worktree without any change dir); run
    `bash scripts/test-check-task-commit-fields.sh` clean.

- [x] 3. check-unfinished-work.sh anchors its store calls at the plan-owning project
**Build:** green
**Files:** `scripts/check-unfinished-work.sh`, `scripts/test-check-unfinished-work.sh`
**Tests:** `case 26: the findings query anchors at the canonical plan dir`, `case 26: the verdict write anchors at the canonical plan dir`, `case 26: the verdict still names the judged worktree`, `case 26b: the findings query anchors at the judged worktree`, `case 26c: the fall-through findings query anchors at the judged worktree`
**Regression:** reverting this commit restores the false CLEAR on the satellite worktree's own verdict line (findings read under the second repo's project key) and records its verdict where the canonical project's tools never look.
**Baseline:** before=84 after=94
<!-- measured: bash scripts/test-check-unfinished-work.sh @ branch spectre/kan-260-guards-are-blind-to-cross-repo-changes -->
<!-- measured: bash scripts/test-check-unfinished-work.sh after task 3 @ branch spectre/kan-260-guards-are-blind-to-cross-repo-changes -->
**Commit:** fix(scripts): anchor unfinished-work store calls at the plan-owning project

  - [x] **Step 1:** derive the store anchor once after signal one resolves: the resolved plan's
    directory when the plan landed outside `$WORKTREE`, else `$WORKTREE`; on the no-plan
    fall-through keep `$WORKTREE`. Use it for the `record findings` query, `record verdict` and
    `record verdicts` `-C` values; `-worktree` keeps naming `$WORKTREE`.
  - [x] **Step 2:** extend the header's SIGNAL TWO paragraph with the anchoring rule and why the
    plan's repo owns the records.
  - [x] **Step 3:** add the three harness cases above; the stub `flow` must capture and assert
    the `-C` it received for each shape; run `bash scripts/test-check-unfinished-work.sh` clean.
