# flow-fast-speedup

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** yes — the three panel dispatch paragraphs and the `final-review.diff` step move
> out of `skills/flow/review-panel.md` into a new `skills/flow/panel-dispatch.md`

Nine tasks, dependency order. `docs/research/kan-492.md` (section 2,
**Decisions**, and section 3, **Implementation touchpoints**) is canonical for every decision a
task implements until `/flow` folds it into this change's `design.md`; a task names the decision
it implements and what the edit must say, never a second copy of it. Task 2 is the one edit to a
`/flow` file; tasks 3–7 rewrite `skills/flow-fast/`; task 8 edits the contracts; task 9 the command
stubs.

**Baseline, measured before any edit, on `main` at `d889abe`:**

- `scripts/check-vocabulary.sh`, `scripts/check-references.sh`, `scripts/check-contract-budget.sh`
  (81 owned Markdown files within budget), `scripts/check-markdown-integrity.py`,
  `scripts/check-stage-mark-calls.sh`, `scripts/check-guard-symlinks.sh` (123 guards across 5
  skills), `scripts/check-installed-citations.sh` and `scripts/check-normative-inventory.sh` (8
  normative sentences from 81 files) all exit 0.
  <!-- measured: each script run from the repository root @ main d889abe, 2026-09-11 -->
- Byte sizes against `scripts/check-contract-budget.sh`'s `budgets()` rows: `skills/flow-fast/SKILL.md`
  10,992 of 13,631; `skills/flow-fast/brainstorm.md` 8,209 of 9,626; `skills/flow-fast/implement.md`
  4,765; `skills/flow-fast/review.md` 4,467; `skills/flow-fast/finish.md` 6,911;
  `skills/flow/review-panel.md` 63,866. Every rewrite below shrinks or holds its file except
  `finish.md` and `review.md`, which grow by the procedure and the literal block they now carry —
  measure with `wc -c` after writing and raise the row only where the file exceeds it.
  <!-- measured: wc -c on each file and the budgets() rows at scripts/check-contract-budget.sh @ main d889abe -->

**Every prose task's verify step names the guards that scan owned Markdown**
(`check-vocabulary.sh`, `check-references.sh`, `check-markdown-integrity.py`,
`check-contract-budget.sh`) **plus whichever of `check-installed-citations.sh`,
`check-guard-symlinks.sh`, `check-stage-mark-calls.sh`, `check-dispatch-paragraphs.sh` and
`check-normative-inventory.sh` its own edit could plausibly trip** — never the project's whole
`## lint` list. No task here adds a test file: every edit is Markdown or a symlink.

**Chained stage marks, every phase file (tasks 3–6):** each stage's `flow stage end` and the next
stage's `flow stage begin` are written into the same fenced Bash block as the next stage's first
command; a mark-only stage is one block carrying both its marks. The literal-token and `-harness`
placeholders are unchanged, and `check-stage-mark-calls.sh` still reads each call.

---

- [ ] 1. Delete the adopted staging note and its plan directory

`docs/research/kan-492.md` seeded this change's brainstorm; its plan and
decision (`docs/research/kan-492/`) are now this change's own `tasks.md` and
decision record — delete all of it.

  - [ ] **Step 1: Delete** `docs/research/kan-492.md` and the directory
    `docs/research/kan-492/`.
  - [ ] **Step 2: Verify** — `scripts/check-vocabulary.sh` and `scripts/check-references.sh` from the
    worktree root; both exit 0.

**Files:** `docs/research/kan-492.md`,
`docs/research/kan-492/tasks.md`,
`docs/research/kan-492/decision.json`
**Tests:** none — deletion of an adopted staging note; the verify step's guard scripts are the check
**Regression:** reverting this commit leaves a stale note and plan in the research tree after their
content has landed in this change's own artifacts, which the "delete once adopted" rule
(`skills/flow/brainstorm-planner.md` section B) exists to prevent.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `docs(flow-research): delete the adopted kan-492 staging note`
**Build:** green

- [ ] 2. `skills/flow/panel-dispatch.md` — the shared dispatch file

Per the note's **Review panel** decision. New file holding, moved verbatim out of
`skills/flow/review-panel.md`: the INDEPENDENT PASSES, TOOLS and MODEL HANDSHAKE dispatch
paragraphs, and the "Before writing `final-review.diff`" step with the diff command itself. Each
moved passage is deleted from `review-panel.md` and replaced there by one citation of the new
file's section — never left as a second copy. Headings in the new file are the ones the two
citing files name.

  - [ ] **Step 1: Write** `skills/flow/panel-dispatch.md` from the moved passages; **cut** each from
    `skills/flow/review-panel.md`, leaving one citation per passage.
  - [ ] **Step 2: Add its budget row** to `scripts/check-contract-budget.sh`'s `budgets()` — `wc -c`
    plus 25% headroom, beside the `skills/flow/…` rows.
  - [ ] **Step 3: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-dispatch-paragraphs.sh`, `scripts/check-installed-citations.sh` all exit 0.

**Files:** `skills/flow/panel-dispatch.md`, `skills/flow/review-panel.md`,
`scripts/check-contract-budget.sh`
**Tests:** none — Markdown relocation; the verify step's guard scripts are the check
**Regression:** reverting this commit leaves `skills/flow-fast/review.md` (task 3) citing a file
that does not exist, which `check-references.sh` reports.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `docs(flow): move the panel dispatch paragraphs into panel-dispatch.md`
**Build:** green

- [ ] 3. `skills/flow-fast/review.md` — panel-dispatch citation, narrower re-run, literal handoff

Per the note's **Review panel** and **Handoff** decisions. Rewrite: cite
`skills/flow/panel-dispatch.md` for the dispatch paragraphs and the `final-review.diff` step
(never `skills/flow/review-panel.md`); after a Critical/Major fix, re-run only the
`simple-reviewer` slot on the round's delta — `primary` keeps its pass-1 verdict — until
`check-panel-findings-closed.sh` exits 0; carry the `IN_PROGRESS` block and the run instructions
verbatim in this file instead of citing `skills/flow-contracts/handoff-blocks.md`; chained marks.

  - [ ] **Step 1: Rewrite** `skills/flow-fast/review.md` per the content above.
  - [ ] **Step 2: Raise its budget row** in `scripts/check-contract-budget.sh` if `wc -c` exceeds it.
  - [ ] **Step 3: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-stage-mark-calls.sh`, `scripts/check-dispatch-paragraphs.sh` all exit 0.

**Files:** `skills/flow-fast/review.md`, `scripts/check-contract-budget.sh`
**Tests:** none — Markdown; the verify step's guard scripts are the check
**Regression:** reverting this commit puts `/flow-fast`'s panel back on the 64 KB
`review-panel.md` load and the two-slot delta re-run.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `docs(flow-fast): cite panel-dispatch.md, re-run simple-reviewer only, carry the handoff`
**Build:** green

- [ ] 4. `skills/flow-fast/brainstorm.md` — direct write, minimal artifacts

Per the note's **Brainstorm** and **Artifacts** decisions and its **Spec fixes**. Section B: read
the project context, one batched `AskUserQuestion` only for true blockers, then write the three
artifacts directly — no `superpowers:brainstorming`, no `docs/superpowers/specs/` file; the staged
research-note seed stays (with its `<stem>/tasks.md` taken in D). Section C: `proposal.md` (`##
Why` / `## What changes`), `design.md` (`## Context` / `## Decisions` as short prose bullets),
`tasks.md` = `- [ ] <n>. <title>` plus `**Files:**`, `**Tests:**`, `**Commit:**` per task and
nothing else. Section D: no `superpowers:writing-plans`, no `check-plan-shape.sh` (the contradiction
with `SKILL.md`'s guard set, resolved in the guard set's favour), no provenance or build-green
tags, no `**Execution:**`/`**Relocation:**` headers; `decision.json` and `flow record decision`
unchanged. Chained marks throughout.

  - [ ] **Step 1: Rewrite** `skills/flow-fast/brainstorm.md` per the content above.
  - [ ] **Step 2: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-stage-mark-calls.sh`, `scripts/check-normative-inventory.sh` all exit 0.

**Files:** `skills/flow-fast/brainstorm.md`
**Tests:** none — Markdown; the verify step's guard scripts are the check
**Regression:** reverting this commit restores the brainstorming checklist, the specs file and the
full plan shape to `/flow-fast`.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `docs(flow-fast): write the artifacts directly, in the minimal shape`
**Build:** green

- [ ] 5. `skills/flow-fast/implement.md` — chained marks, minimal fields

Per the note's **Stage marks** and **Artifacts** decisions. The task loop reads `**Files:**`,
`**Tests:**` and `**Commit:**` only; the `**Build:** red` / `**Squash-with:**` paragraph goes
(the minimal plan never carries the tag); marks are chained.

  - [ ] **Step 1: Edit** `skills/flow-fast/implement.md` per the content above.
  - [ ] **Step 2: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-stage-mark-calls.sh` all exit 0.

**Files:** `skills/flow-fast/implement.md`
**Tests:** none — Markdown; the verify step's guard scripts are the check
**Regression:** reverting this commit makes the implement phase read plan fields task 4 no longer
writes.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `docs(flow-fast): chain the implement marks and read the minimal plan fields`
**Build:** green

- [ ] 6. `skills/flow-fast/finish.md` — the one-run finish

Per the note's **Finish**, **Cleanup**, **Jira** and **Stage marks** decisions and the touchpoints
list's first three bullets. Rewrite as `/flow-fast`'s own procedure, merge-and-push only: the three
guards (`check-finish-preflight.sh`, `check-unfinished-work.sh`, `check-base-moved.sh`, one chained
call, verdict tables cited from `skills/flow-contracts/finish-contract-run1.md`; a `RUN2` verdict is
the wrong-state handoff); `git reset --soft <recorded-merge-base>` and `commit-split.sh`'s chain
(cited from `skills/flow-contracts/git-boundaries.md`); `prepare-archive-branch.sh
<project>/.worktrees/_landing-<name> <base> <base>`; `git merge --no-ff spectre/<name>` in the
landing worktree; `spectre archive <name>` there, on `<base>`; `check-archive-scope.sh`; the
archive commit; one `git push origin <base>`; cleanup = `## stop` (bounded) and
`check-worktree-processes.sh` (cwd outside every worktree) as gates, then `worktree remove
--force`, `branch -d`, `prune`, the project's workspace `remove` command via `flow workspace-id`;
write `FINISHED`; one Jira transition to `Done`; remove the landing worktree. No landing question,
no `## default landing route` read, no change-branch push, no remote delete, no `chore/archive-`
branch, no `flow record render`, no `check-cleanup-complete.sh`. All twenty finish-side keys marked,
mapped as the touchpoints list states, chained.

  - [ ] **Step 1: Rewrite** `skills/flow-fast/finish.md` per the content above.
  - [ ] **Step 2: Raise its budget row** in `scripts/check-contract-budget.sh` if `wc -c` exceeds it.
  - [ ] **Step 3: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-stage-mark-calls.sh`, `scripts/check-normative-inventory.sh` all exit 0.

**Files:** `skills/flow-fast/finish.md`, `scripts/check-contract-budget.sh`
**Tests:** none — Markdown; the verify step's guard scripts are the check
**Regression:** reverting this commit restores the two-run finish with its second landing, archive
branch, second push and six-check cleanup.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `docs(flow-fast): finish in one run — merge, archive on base, push once`
**Build:** green

- [ ] 7. `skills/flow-fast/SKILL.md` and its guard symlinks

Per the touchpoints list's **Stage keys** and **Guard set** bullets and the note's **Finish**
decision. Stage-key table: one `finish.md` row carrying all twelve finish keys (no run 1 / run 2
split). Guard set: `check-archive-scope.sh` joins, `check-cleanup-complete.sh` leaves — the
presence-check sentence, the **Guard set** list and the dropped-guard list all move together.
Guardrails: never ask the landing question, never read `## default landing route`, never push the
change branch, never open a PR — merge-and-push is the only route. Model resolution unchanged.
Symlinks: add `skills/flow-fast/scripts/check-archive-scope.sh`, remove
`skills/flow-fast/scripts/check-cleanup-complete.sh`.

  - [ ] **Step 1: Edit** `skills/flow-fast/SKILL.md` per the content above.
  - [ ] **Step 2: Symlinks** — `ln -s ../../../scripts/check-archive-scope.sh` into
    `skills/flow-fast/scripts/`; `git rm` the `check-cleanup-complete.sh` link there.
  - [ ] **Step 3: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-guard-symlinks.sh`, `scripts/check-stage-mark-calls.sh` all exit 0.

**Files:** `skills/flow-fast/SKILL.md`, `skills/flow-fast/scripts/check-archive-scope.sh`,
`skills/flow-fast/scripts/check-cleanup-complete.sh`
**Tests:** none — Markdown and symlinks; the verify step's guard scripts are the check
**Regression:** reverting this commit leaves `finish.md` (task 6) calling a guard the router does
not presence-check, which `check-guard-symlinks.sh` reports.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `docs(flow-fast): map the one-run finish keys and swap the archive-scope guard in`
**Build:** green

- [ ] 8. Contracts — finish contracts, Jira transitions, artifacts registry

Per the note's **Spec fixes** and **Jira** decisions and the touchpoints list's registry and
Transitions bullets. `finish-contract-run1.md`: the artifact-copy skip's `/myflow-fast` sentence
becomes `/flow-fast`. `finish-contract-run2.md`: the check-4 `/myflow-fast` override paragraph is
deleted (`/flow-fast` skips checks 1–4 outright and runs no run 2); the step-7 and step-9
`/flow-fast` exemption sentences are deleted for the same reason. `jira-integration.md`
**Transitions**: a `/flow-fast` row — `In Progress` at kickoff, `Done` after the push. 
`artifacts-registry.md`: the *Remote branch*, *Archive branch* and *Rendered ledger and panel
record* rows each gain the clause that `/flow-fast` creates none.

  - [ ] **Step 1: Edit** the four contract files per the content above.
  - [ ] **Step 2: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-installed-citations.sh`, `scripts/check-normative-inventory.sh` all exit 0.

**Files:** `skills/flow-contracts/finish-contract-run1.md`,
`skills/flow-contracts/finish-contract-run2.md`, `skills/flow-contracts/jira-integration.md`,
`skills/flow-contracts/artifacts-registry.md`
**Tests:** none — Markdown; the verify step's guard scripts are the check
**Regression:** reverting this commit leaves the contracts describing a `/flow-fast` run 2, a
check-4 override under a retired name, and no Jira row for the command.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `docs(flow-contracts): describe /flow-fast's one-run finish, Jira row and registry rows`
**Build:** green

- [ ] 9. `commands/flow-fast.md` and `commands-claude/flow-fast.md`

Per the note's **Finish** and **Brainstorm** decisions. Both stubs describe the command as it now
runs: direct-write brainstorm, minimal artifacts, one-run merge-and-push finish with no landing
question. Same **Input** / **When done** shape as today.

  - [ ] **Step 1: Edit** both stubs per the content above.
  - [ ] **Step 2: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-installed-citations.sh` all exit 0.

**Files:** `commands/flow-fast.md`, `commands-claude/flow-fast.md`
**Tests:** none — Markdown; the verify step's guard scripts are the check
**Regression:** reverting this commit leaves the command stubs promising a landing question and a
run 2 the skill no longer has.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `docs(commands): describe /flow-fast's direct brainstorm and one-run finish`
**Build:** green
