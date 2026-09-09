# kan-474-flow-brainstorm-planner-should-batch

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Three prose edits, one per file. `design.md` is canonical for every decision and for the wording
each edit carries; a task below names the lines it replaces and what the replacement must say, not
a second copy of the text. No task adds a test: every file here is Markdown, and the checks are the
repository's guard scripts run on the edited file.

**Baseline, measured before any edit:**

- Every guard in `.flow/project.md`'s `## lint` that scans owned Markdown passes on the clean
  branch: `check-vocabulary.sh`, `check-references.sh`, `check-stage-mark-calls.sh`,
  `check-installed-citations.sh`, `check-markdown-integrity.py`, `check-plan-provenance.sh`,
  `check-contract-budget.sh` (71 files within budget), `check-dispatch-paragraphs.sh` (8 sites).
  <!-- measured: each script run from the worktree root, exit 0 @ branch spectre/kan-474-flow-brainstorm-planner-should-batch -->
- The three edited files sit inside their `check-contract-budget.sh` budgets with room for every
  edit below: `skills/flow/brainstorm-planner.md` 20,225 of 23,248 bytes,
  `skills/flow/brainstorm.md` 11,908 of 35,015, `skills/flow-research/SKILL.md` 11,759 of 13,047.
  <!-- measured: wc -c on each file, and the budgets() rows in scripts/check-contract-budget.sh @ branch spectre/kan-474-flow-brainstorm-planner-should-batch -->
- The one-question cadence has exactly three in-tree sources, and no spec cites the relay contract
  (`spectre/specs/` is empty): `superpowers:brainstorming` (installed plugin, not in this
  repository) v6.3.0 lines 169–171, `skills/flow/brainstorm.md` lines 123–126 and 156–160,
  `skills/flow-research/SKILL.md` line 126.
  <!-- measured: grep -rn -i 'one question\|one at a time\|AskUserQuestion' skills/ rules/ and ls spectre/specs @ branch spectre/kan-474-flow-brainstorm-planner-should-batch -->

**Every task that grows an owned `.md` file runs `scripts/check-contract-budget.sh` in its verify
step.** The budget is a ratchet: none of the three files is expected to trip it, and a row is raised
only if the guard actually fails on that file, in that task's own commit.

---

- [x] 1. Planner batches pending questions into one turn

Edit `skills/flow/brainstorm-planner.md`, section **B**, in two places.

**The checklist.** After the existing bullet that ends "a scoped override of
`superpowers:brainstorming`'s hard design-approval gate, `/flow` only." (line 99 on the clean
branch), add one bullet stating: ask every pending question whose wording does not depend on another
pending answer in the same turn, each as its own `## Question` block, up to four blocks per turn; a
question that only makes sense once another is answered waits for the next turn; the convergence
confirm and the third-round offer may be the last block of such a turn; this is a scoped override of
`superpowers:brainstorming`'s "Only one question per message", `/flow` only.

**Convergence.** After the paragraph beginning "*Revise* is a round" (lines 131–133), add one
paragraph stating what a batched confirm means: when the confirm rides along with a round's
questions, **approve the design and move on** folds that turn's answers into the design and ends
the stage; an answer to an accompanying question that names something new opens another round
regardless of the confirm's choice; the silence default and its `⚠ another round — no explicit
answer` marker are unchanged.

Do not touch the relay contract's own wording — task 2 owns `brainstorm.md`.

  - [x] **Step 1: Add the checklist bullet** beneath the design-presentation override bullet, same
    indentation and bold style as its siblings.
  - [x] **Step 2: Add the batched-confirm paragraph** under **Convergence**, after the *Revise*
    paragraph and before "When no answer is possible at all".
  - [x] **Step 3: Verify** — from the worktree root run `scripts/check-markdown-integrity.py`,
    `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-installed-citations.sh` and `scripts/check-contract-budget.sh`; all exit 0.

**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** none — prose edit; the verify step's guard scripts are the check
**Regression:** reverting this commit leaves the planner bound to `superpowers:brainstorming`'s
one-question-per-message rule, so every clarifying question is again its own turn and its own
full-context resend.
**Baseline:** before=0 after=0 — no test file changes; the guards above exit 0 before and after
<!-- measured: the guard runs recorded in the baseline block above @ branch spectre/kan-474-flow-brainstorm-planner-should-batch -->
**Commit:** `feat(flow): batch the planner's independent questions into one turn`
**Build:** green

- [x] 2. Parent relays a planner turn's `## Question` blocks as one AskUserQuestion call

Edit `skills/flow/brainstorm.md` in two places.

**The relay contract** (lines 123–126): "exactly one `## Question` block" becomes "one to four
`## Question` blocks" — each the question plus named options when it has any — or, at the three
returns, one of `## Design`, `## Artifacts`, `## Plan` and nothing else. The handshake sentence
about the first line stays.

**The relay** (paragraph beginning "**The relay.**", lines 156–160): the parent puts every
`## Question` block of the planner's turn into a single **AskUserQuestion** call — one question per
block, in block order, each block's named options as that question's options — and resumes the
planner via **SendMessage** with every answer, labelled by block order. The sentence that section
B's merged confirm and third-round offer are relayed the same way stays; the following paragraph
about prose preceding the blocks stays, with "a `## Question` block" read as "the turn's
`## Question` blocks".

The dispatch prompt the parent writes to the planner is described by this file's own text, so no
separate prompt template needs editing — confirm this with `grep -n 'exactly one' skills/flow/` after
the edit: no hit remains in `brainstorm.md`.

  - [x] **Step 1: Reword the relay contract** to one to four blocks.
  - [x] **Step 2: Reword the relay paragraph** to one call per turn, all blocks, answers labelled by
    block.
  - [x] **Step 3: Verify** — from the worktree root run `scripts/check-markdown-integrity.py`,
    `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-dispatch-paragraphs.sh`, `scripts/check-stage-mark-calls.sh`,
    `scripts/check-installed-citations.sh` and `scripts/check-contract-budget.sh`; all exit 0, and
    `grep -n 'exactly one' skills/flow/brainstorm.md` prints nothing.

**Files:** `skills/flow/brainstorm.md`
**Tests:** none — prose edit; the verify step's guard scripts are the check
**Regression:** reverting this commit leaves the parent asking one AskUserQuestion per block, so a
planner turn carrying several blocks (task 1) is relayed as several calls and the batching buys
nothing.
**Baseline:** before=0 after=0 — no test file changes; the guards above exit 0 before and after
<!-- measured: the guard runs recorded in the baseline block above @ branch spectre/kan-474-flow-brainstorm-planner-should-batch -->
**Commit:** `feat(flow): relay a planner turn's question blocks as one AskUserQuestion call`
**Build:** green

- [x] 3. `/flow-research` batches independent questions into one AskUserQuestion call

Edit `skills/flow-research/SKILL.md` line 126: replace the sentence "Ask one question at a time."
with one stating that every pending question whose wording does not depend on another pending
answer is asked in one **AskUserQuestion** call, up to four per call, and a dependent question waits
for the next turn. The rest of that paragraph — don't funnel the user through a fixed line of
questioning, surface the interesting directions — stays as it is.

  - [x] **Step 1: Replace the sentence.**
  - [x] **Step 2: Verify** — from the worktree root run `scripts/check-markdown-integrity.py`,
    `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-installed-citations.sh` and `scripts/check-contract-budget.sh`; all exit 0, and
    `grep -n 'one question at a time' skills/flow-research/SKILL.md` prints nothing.

**Files:** `skills/flow-research/SKILL.md`
**Tests:** none — prose edit; the verify step's guard scripts are the check
**Regression:** reverting this commit leaves `/flow-research` asking one question per turn, the same
per-question full-context resend the ticket measured in the planner.
**Baseline:** before=0 after=0 — no test file changes; the guards above exit 0 before and after
<!-- measured: the guard runs recorded in the baseline block above @ branch spectre/kan-474-flow-brainstorm-planner-should-batch -->
**Commit:** `feat(flow-research): batch independent questions into one AskUserQuestion call`
**Build:** green
**After:** none
