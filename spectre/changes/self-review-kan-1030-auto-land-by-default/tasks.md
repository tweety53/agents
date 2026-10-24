# self-review-kan-1030-auto-land-by-default

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** close kan-1030's self-review findings F1 (Task 1), F2 (Task 2) and F4 (Task 3), as
`proposal.md` names them; F3 is corrected in its own planning commit (decision
`narrative-corrected-in-planning-commit`).

**Architecture:** prose only — the before-review planning commit moved ahead of the panel's base
sync; an open-question stop and a 🤖-line print on the land-straight-away path; a land-straight-away exception to the stack start.

**Tech Stack:** Markdown, Bash guards.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-24-self-review-kan-1030-auto-land-by-default-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject, no `Task-Id:` trailer and no `Co-Authored-By:` trailer. Never merge or push. Never write `spectre/changes/self-review-kan-1030-auto-land-by-default/` — its `patches/` and `verbatim-moves.txt` are committed with the plan, and its ticks are the parent's.
- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1030-auto-land-by-default` (`<wt>` below), branch `spectre/kan-1030-auto-land-by-default`, committing after its archive commit `c0075269` and rewriting none. Every path is relative to `<wt>`. Never touch the main checkout, and never create a worktree.
- A task's diff is `spectre/changes/self-review-kan-1030-auto-land-by-default/patches/task-<n>.diff`, applied with `git apply <that path>` from `<wt>`, in task order. A patch that does not apply is reported, never hand-merged by guesswork.
- Prose lands character for character as its patch carries it: `scripts/check-verbatim-moves.sh` matches every added and reworded sentence against this change's `verbatim-moves.txt` and exits 1 on any other.

## Review Focus

- Task 1: a re-run round whose fix left new ticks still commits them before its sync; the entry citation pre-check is not re-run at a round boundary.
- Task 2: a run with both `?` and 🤖 lines stops and prints the full block once — never the 🤖 heading first; a run with neither prints nothing before integrate.
- Task 3: a run that stops on an open question still starts its stack, so its block's `Running:` section holds; a `reviewBeforeLanding: true` run is unchanged.

- [ ] 1. Commit the change's directory before the panel's base sync
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'before-review planning commit' skills/flow/review-panel.md` prints nothing, and `grep -n 'Once per worktree in this run' skills/flow/review-panel.md` prints line 26, ahead of every planning-commit mention (`grep -n 'planning commit' skills/flow/review-panel.md`, first at line 457).
    <!-- measured: those greps @ c0075269 → no hit; 26:Once per worktree…; 457:parent makes the reviewer-dispatch planning commit -->
  - [ ] **Step 2: Apply `patches/task-1.diff`**: the **The before-review planning commit comes first.** paragraph opening **Check base movement first** (`skills/flow/review-panel.md`), its round-boundary sentence (`skills/flow/review-panel-fix-round.md`), and the base-sync clause in the "Before every reviewer dispatch" row (`skills/flow-contracts/git-boundaries.md`).
    <!-- measured: git apply --check patches/task-1.diff @ c0075269 → exit 0 -->
  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-installed-citations.sh` and `scripts/check-dispatch-paragraphs.sh` exit 0. Commit.
    <!-- measured: those guards @ c0075269 with tasks 1–3 applied → exit 0 each -->

**Build:** green
**Files:** `skills/flow/review-panel.md`, `skills/flow/review-panel-fix-round.md`, `skills/flow-contracts/git-boundaries.md`
**Tests:** **none** — contract prose; `check-verbatim-moves.sh` and the reference guards cover it (decision `prose-guards-cover`)
**Regression:** none — no test is declared.
**Commit:** fix(flow): commit the change's directory before the panel's base sync
**After:** none

**Decision:** planning-commit-before-base-sync

- [ ] 2. Stop at the handoff on an open question and print auto-resolutions before landing
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'Anything else prints no block and continues' skills/flow/verify-and-handoff.md` prints line 534, and `grep -n 'reviewBeforeLanding. is .true. or' skills/flow-contracts/pipeline.md skills/flow/SKILL.md` prints nothing.
    <!-- measured: those greps @ c0075269 → 534:…; no hit -->
  - [ ] **Step 2: Apply `patches/task-2.diff`**: the `?`-line stop and the 🤖-line print (`skills/flow/verify-and-handoff.md`), the open-question clause in the **States** sentence and three transition-table cells (`skills/flow-contracts/pipeline.md`), and in the fix-run line (`skills/flow/SKILL.md`).
    <!-- measured: git apply --check patches/task-2.diff @ c0075269 with task 1 applied → exit 0 -->
  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-installed-citations.sh` exit 0. Commit.
    <!-- measured: those guards @ c0075269 with tasks 1–3 applied → exit 0 each -->

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`, `skills/flow-contracts/pipeline.md`, `skills/flow/SKILL.md`
**Tests:** **none** — contract prose; `check-verbatim-moves.sh` and the reference guards cover it (decision `prose-guards-cover`)
**Regression:** none — no test is declared.
**Commit:** fix(flow): stop at the handoff on an open question and print auto-resolutions before landing
**After:** Task 1

**Decision:** open-question-stops-auto-land

- [ ] 3. Start no stack on a run that lands straight away
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'land straight away\|lands straight away' skills/flow/verify-and-handoff.md` prints only line 527, below **Resolve the run instructions**, whose start carries no exception for such a run.
    <!-- measured: that grep @ c0075269 → 527:**Then land straight away unless review was asked for** -->
  - [ ] **Step 2: Apply `patches/task-3.diff`**: the **A run that will land straight away starts nothing.** paragraph under **Before this stage ends, start the stack** (`skills/flow/verify-and-handoff.md`).
    <!-- measured: git apply --check patches/task-3.diff @ c0075269 with tasks 1–2 applied → exit 0 -->
  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-installed-citations.sh` exit 0. Commit.
    <!-- measured: those guards @ c0075269 with tasks 1–3 applied → exit 0 each -->

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none** — contract prose; `check-verbatim-moves.sh` and the reference guards cover it (decision `prose-guards-cover`)
**Regression:** none — no test is declared.
**Commit:** perf(flow): start no stack on a run that lands straight away
**After:** Task 2

**Decision:** no-stack-start-on-auto-land
