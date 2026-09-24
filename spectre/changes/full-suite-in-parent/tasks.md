# full-suite-in-parent

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Two prose tasks: move the full `## test` run from the plan-last implementer to the parent
(task 1), then remove every resume of a finished subagent (task 2).
`design.md` is canonical for the mechanism and both decisions.

No live-verification task: the change edits prose contracts only, and no running service or
persistent state is touched.

**Baseline, measured before any edit:**

- The FULL SUITE paragraph and every sentence citing it sit at `implement.md:685-697`, `:944-950`,
  `:990`, and `brainstorm-planner.md:241`, `:254`; no guard pins their wording.
  <!-- measured: git grep -n -E 'FULL SUITE|full-suite|full suite|plan-last|last bundle' -- ':!spectre/changes/archive' @ f7a825b -->
- `check-contract-budget.sh` passes; `implement.md` is 71071 bytes against a 77479 row,
  `brainstorm-planner.md` 41268 against 53065.
  <!-- measured: scripts/check-contract-budget.sh; wc -c; grep -n in scripts/check-contract-budget.sh @ f7a825b -->

---

- [x] 1. The parent runs the full suite after the last group

**Files:** `skills/flow/implement.md`, `skills/flow/brainstorm-planner.md`
**Tests:** none — prose; the lint guards in the verify step are the check
**Regression:** reverting this commit puts FULL SUITE back into the plan-last implementer's
dispatch, so a suite run over 5 minutes again expires that subagent's prompt cache.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**After:** none
**Commit:** `feat(flow): run the full suite in the parent, not the plan-last implementer`
**Build:** green

**Decision:** full-suite-runs-in-parent
**Decision:** parent-fixes-own-file-failure

  - [x] **Step 1: `skills/flow/implement.md` — the dispatch.** Delete the whole block that opens
    "**The plan-last group's implementer dispatch — … alone also carries:**" and its `> **FULL
    SUITE:**` quote (lines 685-697). Replace it with one parent-side paragraph stating the
    mechanism `design.md`'s **Mechanism** section fixes: after the last group's guard passes, the
    parent runs the resolved `## test` list once on the canonical worktree, foreground, in context
    bundle order, output through `tail`; a failure in a file the plan-last group's tasks declare
    in `**Files:**` the parent fixes inline, commits on the push-state route (**Panel re-runs**,
    `skills/flow/review-panel.md`), records as `-role panel-fix -agent-id inline` under key
    `full-suite-fix-<n>`, and re-runs the list; any other failure is the last-boundary
    `## Question` stop. Leave the TARGETED TESTS paragraph untouched.
  - [x] **Step 2: `skills/flow/implement.md` — the last boundary (lines 944-950).** Replace "the
    last implementer's report carries no `## Full suite` failure" with the parent's full-suite run
    having passed (after any own-file fix); replace "A report that records a full-suite failure"
    with a full-suite failure outside the plan-last group's files. Keep the `## Question`
    consequence and "the panel never runs on a red branch" verbatim.
  - [x] **Step 3: `skills/flow/implement.md` line 990.** "the full-suite run after a shared wave"
    → "the full-suite run after the last group".
  - [x] **Step 4: `skills/flow/brainstorm-planner.md` lines 241-242 and 254-255.** "the last
    bundle's FULL SUITE paragraph (`skills/flow/implement.md`)" → "the parent's full-suite run
    after the last group (`skills/flow/implement.md`)"; "the last bundle's FULL SUITE run still
    covers the pair" → "the parent's full-suite run after the last group still covers the pair".
  - [x] **Step 5: Verify.** `git grep -n -E 'FULL SUITE|## Full suite' -- skills/` exits 1;
    `scripts/check-references.sh`, `scripts/check-vocabulary.sh` and
    `scripts/check-contract-budget.sh` exit 0; `scripts/test-check-dispatch-paragraphs.sh | tail -3`
    passes.

- [x] 2. No subagent is ever resumed — the parent fixes inline

**Files:** `skills/flow/implement.md`
**Tests:** none — prose; the lint guards in the verify step are the check
**Regression:** reverting this commit restores the four resume sites, so a guard refusal, a pick
conflict or a gated `fix` verdict again re-wakes a finished implementer whose cache has expired.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**After:** Task 1
**Commit:** `feat(flow): never resume a finished subagent; the parent fixes inline`
**Build:** green

**Decision:** no-subagent-resume

  - [x] **Step 1: The rule.** Under **Dispatch sites — the parent's closed list**, after **The
    self-check**, add one paragraph: every dispatch is one-shot, the parent never `SendMessage`s a
    finished child, and whatever that child's work still needs is the parent's own inline fix —
    citing why (a subagent's prompt cache lives five minutes).
  - [x] **Step 2: The wave pick** (the **As wave members return** paragraph). "hands the group
    back to its own implementer — its throwaway worktree rebased … re-pick" → the parent resolves
    the conflict or re-commits in the canonical worktree itself and re-runs the guard; drop "or
    after handback resolves".
  - [x] **Step 3: The guard verdicts** (boundary step 2). "sends that task back to the **same
    implementer**, which re-commits" → the parent re-commits that task itself; "the same
    implementer re-committing without the swept paths" → the parent re-committing without them.
  - [x] **Step 4: The gated fix** (**The gated per-task reviewer**). "**A fix resumes the task's
    own group's implementer** (`SendMessage`; …)" → the parent applies every `fix` report of the
    group itself, recorded as one inline pair under `task-<n+n>-implementer-fix-<k>`; "writes
    `implementer-report-<k>-fix-<n>.md`" dropped; "the implementer resolves it by hand" → the
    parent; delete the now-redundant closing sentence "On an inline run the parent applies the
    fixes itself …, and the re-review is still a dispatch." keeping "the re-review is a fresh
    reviewer dispatch" stated once.
  - [x] **Step 5: REPORT FILE.** Drop "; a resumed fix writes `implementer-report-<k>-fix-<n>.md`
    instead".
  - [x] **Step 6: Verify.** `git grep -n -E 'SendMessage\)|same implementer|own implementer|resumed
    fix' -- skills/flow/implement.md` exits 1; `scripts/check-references.sh`,
    `scripts/check-vocabulary.sh`, `scripts/check-contract-budget.sh` exit 0;
    `scripts/test-check-dispatch-paragraphs.sh | tail -3` passes.
