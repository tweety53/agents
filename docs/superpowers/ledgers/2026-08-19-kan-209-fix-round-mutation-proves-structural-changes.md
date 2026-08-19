# SDD ledger — kan-209-fix-round-mutation-proves-structural-changes

Command: `/myflow-fast`. Roster: `light`. Recorded models: implementation `sonnet`,
review panel `sonnet`, panel fixes `sonnet` — the defaults `/myflow-fast` records on a creating run,
with no operator override given this run.

Worktree: `/Users/tweety53/Projects/agents-worktrees/kan-209-fix-round-mutation-proves-structural-changes`
Branch: `openspec/kan-209-fix-round-mutation-proves-structural-changes`
Merge base: `c6bd1cb`

## Dispatches

### Task 1 — `scripts/test-check-unfinished-work.sh`, the record lines are inert to the guard

- **Implementer:** general-purpose subagent, model: `sonnet` (recorded `models.implementation`).
- **Commit:** `5a9b100`, `Task-Id: 1`.
- **Result:** case 12 added; 122 → 125 `ok:` assertions. `check-task-commit-fields.sh` exit 0.
- **Mutation proof (two isolated mutations, restored between each):**
  1. A `fix-mutation:` line moved between two `finding-status:` markers → the guard's `M_SPAN`
     contiguity check fired, flipping the CLEAR assertion to `OUTSTANDING:`.
  2. A `fix-mutation:` line reshaped as `| F9 | fix-mutation | scripts/check-thing.sh |` → the
     guard's `ROW_IDS`/`MARKER_IDS` comparison fired, flipping the same assertion.
- **Reported deviation from the plan's prose:** the plan predicted mutation 2 would break the
  `assert_reason "1 open finding(s)"` assertion. It does not — the guard joins fired reasons into
  one semicolon-separated line, so an already-`OUTSTANDING:` fixture absorbs the extra reason. Both
  mutations therefore surface through the CLEAR assertion rather than through two different ones.
  **Referred to the per-task reviewer** to judge whether both mechanisms remain independently
  pinned or whether one is now unpinned.
- **Per-task review:** combined spec + quality reviewer, model: `sonnet` (recorded
  `models.reviewPanel`), `light` roster — one combined reviewer per task.

### Task 2 — `skills/myflow-do/SKILL.md`, the rule and its reasoning

- **Implementer:** general-purpose subagent, model: `sonnet` (recorded `models.implementation`).
- **Base:** `5a9b100`.
- **Commit:** `5176654`, `Task-Id: 2`. `check-task-commit-fields.sh` exit 0.
- **Result:** the rule in `skills/myflow-do/SKILL.md` (between the reproducer re-run paragraph and
  the bounce paragraph, as the plan specified) and its reasoning in `SKILL-rationale.md`. All nine
  repository guards exit 0. No budget row raised — `SKILL.md` 64021 → 67158 (row 71317),
  `SKILL-rationale.md` 14132 → 15600 (row 15721, 121 bytes headroom).
- **Per-task review:** combined reviewer, model: `sonnet`, `light` roster — **clean, 0 findings**.
  All nine delta-spec scenarios traced to specific lines; the Bugbot-disposition contradiction hunt
  came back unmistakable.

### Task 1 fix round — per-task review findings

- **Reviewer verdict:** 3 findings (1 Critical, 1 Major, 1 Minor), one root cause.
  - **F1 Critical** — case 12 does not pin the marker-contiguity (`M_SPAN`) mechanism it claims to.
    Its `CLEAR:` fixture carries a **single** marker, so "a line between two markers" cannot occur
    there; its `OUTSTANDING:` fixture's assertions are insensitive to an extra joined reason. Only
    the `ROW_IDS`/`MARKER_IDS` mechanism was genuinely fail-by-construction.
  - **F2 Major** — `assert_reason "1 open finding(s)"` is vacuous: it passes with the pass log
    omitted, present, and under both mutations.
  - **F3 Minor** — the case header cites case 9 for contiguity; contiguity is asserted in 4s-0 and
    4s-0b, and row correspondence in the 4q/4r subgroup.
- **Parent verification:** F1 confirmed by direct read of the committed case (`write_panel 1 fixed`
  — one marker). F3 confirmed by locating the `assert_reason "must be one unbroken block"` lines.
- **Root cause:** the original mutations were run against fixtures the **committed** case does not
  contain — the implementer bumped the fixture to two markers to perform mutation 1, then committed
  the one-marker version. This is the exact defect class KAN-209 exists to prevent, reproduced
  inside the change that lands the rule against it.
- **Fix subagent:** general-purpose, model: `sonnet` (recorded `models.panelFix`), folded into
  Task 1's commit via `git commit --fixup=5a9b100` + `git rebase --autosquash`.
