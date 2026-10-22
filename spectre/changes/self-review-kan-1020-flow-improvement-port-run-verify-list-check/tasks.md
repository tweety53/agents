# self-review-kan-1020-flow-improvement-port-run-verify-list-check

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** close kan-1020's three non-big self-review findings — 1 and 3 (Task 1), 2 (Task 2), as
`proposal.md` numbers them. Finding 3's record repair is committed with this plan (decision
`repair-task-2-correction-at-plan-time`); Task 1's verify re-checks it.

**Architecture:** two sentences of skill prose, each acknowledged in this change's
`verbatim-moves.txt`.

**Tech Stack:** Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-22-self-review-kan-1020-flow-improvement-port-run-verify-list-check-design.md`
in this worktree.

## Global constraints

- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1020-flow-improvement-port-run-verify-list-check` (`<wt>` below), on branch `spectre/kan-1020-flow-improvement-port-run-verify-list-check`; every `**Files:**` path is relative to it. Never touch the main checkout and never create a worktree.
- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Write every sentence a task quotes exactly as quoted, with the line wraps shown — `spectre/changes/self-review-kan-1020-flow-improvement-port-run-verify-list-check/verbatim-moves.txt` acknowledges those sentences verbatim, and `scripts/check-verbatim-moves.sh` fails any other wording.

## Review Focus

- Task 1: the sentence sits inside the Correction paragraph, after "inline, the session is both halves.", not as a paragraph of its own.
- Task 2: the `pipeline.md` load paragraph above the new sentence is byte-identical to the base.

- [ ] 1. Run the provenance guard when transcribing a Correction
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In a scratch clone of `<wt>` (`git clone -q <wt> "$TMPDIR/r1020"`), reword Task 3's Correction in its `spectre/changes/archive/kan-1020-flow-improvement-port-run-verify-list-check/tasks.md` so "gained a subtest one line past the threshold" carries a bare count instead, and run `scripts/check-plan-provenance.sh` in the clone: it reports that line as "numeric claim with no measured:/predicted: provenance comment" and exits 1. Never edit `<wt>/spectre/changes/`.
  - [ ] **Step 2: Write the sentence** per decision `correction-provenance-at-transcription` in `skills/flow/implement.md`, at the end of the paragraph opening "**A deviation from the plan records as a dated `Correction:` paragraph.**": the line ending "inline, the session is both halves." becomes the first line of the block below, and the block's other two lines follow it.

```markdown verified:applied to a scratch clone at 243eca24; scripts/check-verbatim-moves.sh with this change's verbatim-moves.txt exits 0
reports the deviation, the parent transcribes it; inline, the session is both halves. A `Correction:` paragraph is plan prose `check-plan-provenance.sh` reads: run that guard when
transcribing it, and give each numeric claim its `measured:` comment or word it without a count
(kan-1020: a bare count in a Correction was refused only at verify).
```

  - [ ] **Step 3: Verify.** `scripts/check-plan-provenance.sh` exits 0 in `<wt>` — it scans kan-1020's archived `tasks.md`, whose Task 2 now carries the Correction committed with this plan; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-installed-citations.sh`, `scripts/check-normative-inventory.sh` and `scripts/check-ask-silence.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/implement.md`
**Tests:** none — prose rule
**Regression:** none — prose; Step 3's provenance run covers the repaired record.
**Commit:** docs(flow): run the provenance guard when transcribing a Correction
**After:** none

**Decision:** correction-provenance-at-transcription
**Decision:** repair-task-2-correction-at-plan-time

- [ ] 2. Load contract files through the file-read tool, never cat
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `wc -c skills/flow-contracts/pipeline.md skills/flow-contracts/finish-contract-run1.md` shows both files larger than the shell tool's output cap; `grep -n "cat\|Read tool\|file-read" skills/flow/SKILL.md` finds no load instruction naming the tool.
  - [ ] **Step 2: Write the sentence** per decision `contract-load-via-read` in `skills/flow/SKILL.md`, as the four lines of the block below inserted right after the line ending "**Stage keys** (`skills/flow/stage-keys.md`) names which phase file marks each key." and before the blank line that closes that paragraph.

```markdown verified:applied to a scratch clone at 243eca24; scripts/check-verbatim-moves.sh with this change's verbatim-moves.txt exits 0
Load every contract file with the harness's file-read tool — Read on Claude Code — in offset
chunks wherever one read cannot return the whole file, never with `cat`: `pipeline.md` and
`finish-contract-run1.md` exceed the shell tool's output cap, and a `cat` load arrives truncated
(kan-1020).
```

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-installed-citations.sh`, `scripts/check-normative-inventory.sh`, `scripts/check-ask-silence.sh`, `scripts/check-stage-mark-calls.sh` and `scripts/check-self-review-report.sh` exit 0; the task's diff of `skills/flow/SKILL.md` shows only the four added lines. Commit.

**Build:** green
**Files:** `skills/flow/SKILL.md`
**Tests:** none — prose rule
**Regression:** none — prose.
**Commit:** docs(flow): load contract files through the file-read tool, never cat
**After:** none

**Decision:** contract-load-via-read
