# kan-378-flow-deletion-only-cut-of-stale-restated-and

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** yes — tasks 1–4 move existing prose (transition rows into `pipeline.md`,
> sections B–D into `brainstorm-planner.md`, rationale into `SKILL-rationale.md`); no normative
> sentence, exit-code contract, scenario, operator prompt or rejected-alternative reason is reworded.

## Global constraints

- Cut, never paraphrase (`~/.claude/rules/be-brief.md`). A passage is deleted whole or moved
  verbatim; a sentence that stays keeps its exact wording. Where a step says "delete the clause",
  the remaining sentence is the original with that clause removed and nothing else changed.
- Every command under `## lint` in `.flow/project.md` exits clean after every task; the ones each
  task names are the ones its edits can trip. `scripts/run-guard-tests.sh` passes at task 8.
- No task lowers a `scripts/check-contract-budget.sh` row; the two rows added are the new files'
  landed byte size times 1.25, rounded down, matching the guard's own header.
- Line numbers below are as of `e7d9540`; each step also quotes the text it targets so the
  implementer finds it by content when earlier tasks have shifted the lines.

- [x] 1. Make `pipeline.md` the sole transition table

**Build:** green

**Files:**
- Modify: `skills/flow-contracts/pipeline.md`
- Modify: `skills/flow/SKILL.md`

**Tests:** **none** — prose move, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow-contracts): carry /flow's per-state transition rows in pipeline.md`

  - [ ] **Step 1: Expand the `/flow` row in `skills/flow-contracts/pipeline.md`**

  Under `## State transitions` (line 77), replace the single row at line 81 — the one beginning
  `` | `/flow` | *(none — creates the change)* · `STARTED` · `IN_PROGRESS` | per state, exactly as `` —
  with these five rows. Each "Ends at" cell is the corresponding cell of SKILL.md's table (lines
  44–48) verbatim, apart from the first, whose parenthetical repoints from SKILL.md to
  `brainstorm.md` since the section it names lives there.

  ```markdown verified:cells copied from skills/flow/SKILL.md lines 44-48 at e7d9540
  | `/flow` | *(no state — creates the change)* | `STARTED`, same invocation continuing to `IN_PROGRESS` unless it stops early (see **Resuming at `STARTED`** in `skills/flow/brainstorm.md`) |
  | `/flow` | `STARTED` | resumes the creating run from wherever it stopped; ends at `STARTED` (still resuming) or `IN_PROGRESS` |
  | `/flow` | `IN_PROGRESS`, with an argument | fix run; state unchanged |
  | `/flow` | `IN_PROGRESS`, bare | integrate run; ends at `IN_PROGRESS` (run 1) or `FINISHED` (run 1 chained into run 2) |
  | `/flow` | `FINISHED` | wrong-state handoff — the change is archived |
  ```

  Directly under the table, before `**This table is authoritative.**`, add one line so the old
  row's **Stage keys** pointer survives:

  ```markdown verified:authored in-tree for this change
  Which phase file marks each `flow.*` key is **Stage keys** (`skills/flow/SKILL.md`).
  ```

  The `/flow-status` row at line 82 is untouched.

  - [ ] **Step 2: Delete SKILL.md's `## State transitions` section, keeping its `**No flags.**` paragraph**

  In `skills/flow/SKILL.md` delete lines 40–48: the `## State transitions` heading, the blank line,
  the table header, its separator, and the five rows. Keep lines 50–51 (`**No flags.** The only
  argument is …ignoring it.`) — they now sit directly under the "Then register this run's steps"
  paragraph, above `## Stage keys`.

  - [ ] **Step 3: Replace the stale pointer language in SKILL.md's load-first paragraph**

  In the paragraph at lines 26–31, delete everything from `` `/flow` is not yet a row in that file's own transition `` (line 28, after the sentence ending `change-name resolution.`) through `below as `/flow`'s actual contract in the meantime.` (line 31) and put this single sentence in its place, so the paragraph ends:

  ```markdown verified:authored in-tree for this change
  handoff shape and change-name resolution. Its **State transitions** table is `/flow`'s contract;
  **Stage keys** below names which phase file marks each key.
  ```

  - [ ] **Step 4: Verify the references and budget guards**

  Run: `scripts/check-references.sh && scripts/check-contract-budget.sh && scripts/check-vocabulary.sh`
  Expected: each exits 0. `pipeline.md:81` no longer cites a SKILL.md heading that does not exist.

  - [ ] **Step 5: Commit**

  ```bash verified:the commit subject is this task's Commit field
  git add skills/flow-contracts/pipeline.md skills/flow/SKILL.md
  git commit -m "docs(flow-contracts): carry /flow's per-state transition rows in pipeline.md"
  ```

- [x] 2. Cut SKILL.md's stale Stage keys disclosure and model-policy caveat; start `SKILL-rationale.md`

**Build:** green

**Files:**
- Modify: `skills/flow/SKILL.md`
- Create: `skills/flow/SKILL-rationale.md`
- Modify: `scripts/check-contract-budget.sh`
- Modify: `scripts/check-references.sh`

**Tests:** **none** — prose cut and move, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): cut the stale Stage keys disclosure; move SKILL.md rationale to SKILL-rationale.md`

  - [ ] **Step 1: Create `skills/flow/SKILL-rationale.md` with this preamble**

  ```markdown verified:authored in-tree for this change
  # flow — rationale

  Reasoning behind `skills/flow/`'s phase files: why a rejected alternative was rejected, which
  design.md decision a passage implements, and choices a task made on its own. Moved here verbatim
  from the run-loaded files; **no `/flow` run loads this file.** Each heading names the source file
  and the section the passage came from.
  ```

  Every later step in this task and in tasks 3 and 4 appends under a `## <file> — <section>`
  heading, quoting the moved text verbatim as a blockquote.

  - [ ] **Step 2: Move the Stage keys design-decision paragraph**

  In `skills/flow/SKILL.md` under `## Stage keys`, lines 55–62 are one paragraph beginning
  `Every stage `/flow` marks uses a `flow.*` key, minted fresh for this command` and ending
  `sometimes mean something different than its history records.` Append it verbatim to
  `SKILL-rationale.md` under `## SKILL.md — Stage keys`, then in SKILL.md replace the paragraph with
  its normative first clause only:

  ```markdown verified:the first clause of skills/flow/SKILL.md line 55 at e7d9540
  Every stage `/flow` marks uses a `flow.*` key.
  ```

  - [ ] **Step 3: Delete the stale known-gap paragraph**

  Delete lines 64–73 — the paragraph beginning `**This introduces a real, known gap**, disclosed
  rather than hidden:` and ending `` `commands/flow.md`, `commands-claude/flow.md`). `` — and the
  blank line after it. `stats/internal/stages/names.go` carries every `flow.*` key today, so the
  paragraph is stale rather than rationale; nothing is moved. The sentence `The full key list, in
  the order each phase file marks them:` and the table stay.

  - [ ] **Step 4: Delete the model-policy caveat paragraph**

  Delete lines 86–93, the paragraph beginning `` **`skills/flow-contracts/model-policy.md` is only partly current for `/flow`.** `` and ending `do not follow the three-role table for this command.`, plus the blank line after it. Its design-id cites (`model-default-sonnet`, `models-fields-collapse`, `settings-scope`) are recorded already in the design that introduced them and are not rejected-alternative reasons; nothing is moved. Line 155's `model-policy.md` cite for the retired per-change fields stays.

  - [ ] **Step 5: Move SKILL.md's remaining design-id clauses**

  For each of these, append the full original sentence verbatim to `SKILL-rationale.md` under the
  named heading, then delete only the parenthetical or clause from SKILL.md:

  | SKILL.md line | Heading in SKILL-rationale.md | Delete from SKILL.md |
  |---|---|---|
  | 11 | `## SKILL.md — preamble` | the parenthetical `` (design.md's `flow-rename-content-split`) `` |
  | 148 | `## SKILL.md — Model resolution` | `` — per design.md's `model-default-sonnet`: one default, chosen once per run, not three per-role defaults `` (line 148 through the end of line 149's first sentence); the sentence then ends after `` (`skills/flow/review-panel.md`) `` with a full stop |
  | 206 | `## SKILL.md — Guardrails` | the parenthetical `(design.md's "one token per session, not one per mark")` |

  - [ ] **Step 6: Add the budget row**

  In `scripts/check-contract-budget.sh`'s `budgets()` heredoc, after the `skills/flow/SKILL.md`
  row, add a row `skills/flow/SKILL-rationale.md <N>` where `<N>` is `wc -c < skills/flow/SKILL-rationale.md`
  times 1.25 rounded down, computed after task 4 has finished appending — so add the row in this task
  with today's size and raise it once, in task 4's commit, to the final landed size times 1.25.

  - [ ] **Step 7: Verify**

  Run: `scripts/check-references.sh && scripts/check-contract-budget.sh && scripts/check-vocabulary.sh && scripts/check-markdown-integrity.py`
  Expected: each exits 0; `README.md:117`, `commands/flow.md:33` and `commands-claude/flow.md:28`
  still resolve **Stage keys**.

  - [ ] **Step 8: Commit**

  ```bash verified:the commit subject is this task's Commit field
  git add skills/flow/SKILL.md skills/flow/SKILL-rationale.md scripts/check-contract-budget.sh
  git commit -m "docs(flow): cut the stale Stage keys disclosure; move SKILL.md rationale to SKILL-rationale.md"
  ```

- [x] 3. Split `brainstorm.md` into a parent file and `brainstorm-planner.md`

**Build:** green

**Files:**
- Modify: `skills/flow/brainstorm.md`
- Create: `skills/flow/brainstorm-planner.md`
- Modify: `skills/flow/SKILL-rationale.md`
- Modify: `skills/flow-contracts/pipeline.md`
- Modify: `skills/flow-contracts/pipeline-rationale.md`
- Modify: `README.md`
- Modify: `scripts/check-contract-budget.sh`

**Tests:** **none** — prose move, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): split brainstorm.md into parent and planner files`

  - [ ] **Step 1: Create `skills/flow/brainstorm-planner.md`**

  Move lines 190–507 of `skills/flow/brainstorm.md` — from `## B. Basic Workflow #1 — Brainstorming`
  to the end of the file — verbatim into the new file, under this preamble:

  ```markdown verified:authored in-tree for this change
  # Brainstorm and plan — the planner's sections

  Sections **B**, **C** and **D** of the brainstorming phase, read by the planner subagent that
  **Dispatch the planner** (`skills/flow/brainstorm.md`) sends out; every "you" below addresses that
  planner. The parent never reads this file.

  ```

  `brainstorm.md` then ends at line 188 (`continues into `skills/flow/implement.md` exactly as today.`).

  - [ ] **Step 2: Repoint the dispatch prompt and the parent's section pointers**

  In `brainstorm.md`:
  - line 124–125: `the instruction to read this file's sections **B**, **C** and **D**` becomes
    `` the instruction to read `skills/flow/brainstorm-planner.md`'s sections **B**, **C** and **D** ``;
  - line 99: `Sections **B**, **C** and **D** below are the planner's work` becomes
    `` Sections **B**, **C** and **D** of `skills/flow/brainstorm-planner.md` are the planner's work ``;
  - line 84 (`resume at **B** below`) and line 87 (`resume at **D** below`): `below` becomes
    `` in `skills/flow/brainstorm-planner.md` `` in each;
  - line 162–163 `Section B's convergence confirm` and line 176 `Once `## Plan` returns` need no
    change — they name no section by bold token.

  In `brainstorm-planner.md`, `above` references to **Dispatch the planner** (lines 193, 198, 349,
  434, 505 of the original) become `` (`skills/flow/brainstorm.md`) `` — e.g. line 193's
  `mark now lives in **Dispatch the planner** above` becomes
  `` mark now lives in **Dispatch the planner** (`skills/flow/brainstorm.md`) ``. Section C's
  `per **C** below` and D's `**B**`/`**C**` references within the planner file stay.

  - [ ] **Step 3: Repoint the Convergence citations**

  Change `` (`skills/flow/brainstorm.md`) `` to `` (`skills/flow/brainstorm-planner.md`) `` on:
  - `skills/flow-contracts/pipeline.md:66` — `See **Convergence** (…)`;
  - `skills/flow-contracts/pipeline-rationale.md:11` — `harmonised belong to the command itself — **Convergence** (…)`;
  - `README.md:192` — `both honest are **Convergence** (…)`;
  - `README.md:329` — `**Convergence** (…) for the threshold`.

  `skills/flow-research/SKILL.md:38`, `skills/flow-contracts/handoff-blocks.md:92`,
  `skills/flow/implement.md:4,48,126` and `skills/flow/SKILL.md:44,79,137,168,170,198` cite
  parent-side sections and stay. `scripts/check-stage-mark-calls.sh` scans phase files by basename;
  the planner file marks no stage, so it needs no entry.

  - [ ] **Step 4: Move brainstorm.md's rationale to `SKILL-rationale.md`**

  Append each original passage verbatim under the named heading, then cut from the source as shown.
  The last two rows are in `brainstorm-planner.md` after Step 1.

  | Source (original line) | Heading | Delete from source |
  |---|---|---|
  | `brainstorm.md` 3–7 | `## brainstorm.md — preamble` | `` — the same content `/myflow-start` carried, minus the options-question round and the proposal publish, both removed per design.md (`ask-options-removed`, `publish-proposal-removed`) ``; the sentence then reads `…intertwined with spectre artifact creation, run here. Loaded by …` |
  | `brainstorm.md` 37–40 | `## brainstorm.md — A. Resolve the change and write STARTED` | `` This is design.md's `started-redefined`: `STARTED` is a kickoff marker, "the operator started this," not (as under the old `/myflow-start`) a record that a design was approved and a proposal published. `` — keep `Write it here, at the top of this phase, rather than at the bottom of it.` |
  | `brainstorm.md` 61–65 | same heading | `` (`ask-options-removed`) ``, `` (`model-default-sonnet`, `settings-scope`) `` and `` (`publish-proposal-removed`) `` — the three parentheticals only |
  | `brainstorm-planner.md` (original 204–206) | `## brainstorm-planner.md — Seed from a staged research note` | `` This implements design.md's `flow-research-staging`, the *discovery* half of open question `research-staging-mechanism` (the *write* half is `skills/flow-research/SKILL.md`'s own job): `` — the preceding sentence then ends `named below**.` and the bullet list follows |
  | `brainstorm-planner.md` (original 256, 260–263) | same heading | line 256's `` This is `flow-research-staging`'s explicit choice: seed, never skip. ``; and `` This resolves the remaining half of `research-staging-mechanism` left open by design.md: a staging note that outlives its adoption is a second, driftable copy of what the change's own `design.md` now states canonically, and `<project>/docs/superpowers/research/` is meant to hold notes still waiting for a home, not a permanent archive of every note that found one. `` — keep `Delete it as part of **C**'s artifact-creation commit …` onward |

  - [ ] **Step 5: Add the budget row**

  In `scripts/check-contract-budget.sh`'s `budgets()` heredoc, after the `skills/flow/brainstorm.md`
  row, add `skills/flow/brainstorm-planner.md <N>` with `<N>` = `wc -c < skills/flow/brainstorm-planner.md`
  times 1.25 rounded down. Leave the `skills/flow/brainstorm.md 35015` row as it is.

  - [ ] **Step 6: Verify**

  Run: `scripts/check-references.sh && scripts/check-contract-budget.sh && scripts/check-vocabulary.sh && scripts/check-markdown-integrity.py && scripts/check-stage-mark-calls.sh && scripts/check-dispatch-paragraphs.sh`
  Expected: each exits 0. Then `grep -n 'brainstorm.md' skills/flow/brainstorm-planner.md` shows
  only the preamble's and the repointed **Dispatch the planner** cites.

  - [ ] **Step 7: Commit**

  ```bash verified:the commit subject is this task's Commit field
  git add skills/flow/brainstorm.md skills/flow/brainstorm-planner.md skills/flow/SKILL-rationale.md skills/flow-contracts/pipeline.md skills/flow-contracts/pipeline-rationale.md README.md scripts/check-contract-budget.sh
  git commit -m "docs(flow): split brainstorm.md into parent and planner files"
  ```

- [x] 4. Move `integrate.md`'s rationale to `SKILL-rationale.md`

**Build:** green

**Files:**
- Modify: `skills/flow/integrate.md`
- Modify: `skills/flow/SKILL-rationale.md`
- Modify: `scripts/check-contract-budget.sh`

**Tests:** **none** — prose move, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): move integrate.md rationale to SKILL-rationale.md`

  - [ ] **Step 1: Move each passage verbatim, then cut it from `integrate.md`**

  | `integrate.md` line | Heading | Delete from `integrate.md` |
  |---|---|---|
  | 3–7 | `## integrate.md — preamble` | `` Run 1 of what was `/myflow-finish`, unchanged in procedure except two folds design.md decides: `move-in-review-fold` (the Jira "move to In Review" step becomes a sub-step of `flow.landing-routes` rather than its own mark) and this task's own resolution of open question `write-in-progress-fold` (below). `` — the preamble then reads `` Loaded by `skills/flow/SKILL.md` on a **bare** invocation at `IN_PROGRESS` — no argument. `` |
  | 105–107 | `## integrate.md — rebase` | `` — per design.md's `rebase-is-a-confirmed-choice`: the rebase never runs on its own, only after the operator picks this option, and only against the worktree(s) that actually moved `` (keep the sentence up to `per **ask-only-on-overlap**` and close it with a full stop) |
  | 119 | same heading | `` , per design.md's `never-auto-abort` `` |
  | 127 | `## integrate.md — Scoped re-verification` | `` , per design.md's `scoped-reverify-not-full-suite` `` |
  | 226–236 | `## integrate.md — landing routes` | the whole paragraph beginning `This stage carries three sub-steps under one mark, per this task's own resolution` through its final sentence; replace it with its normative first clause: `This stage carries three sub-steps under one mark: the git route, the state write, and the Jira transition.` |

  - [ ] **Step 2: Raise the `SKILL-rationale.md` budget row to its landed size**

  In `scripts/check-contract-budget.sh`, set the `skills/flow/SKILL-rationale.md` row to
  `wc -c < skills/flow/SKILL-rationale.md` times 1.25 rounded down, now that its content is final.

  - [ ] **Step 3: Verify**

  Run: `scripts/check-references.sh && scripts/check-contract-budget.sh && scripts/check-vocabulary.sh && scripts/check-markdown-integrity.py && scripts/check-stage-mark-calls.sh`
  Expected: each exits 0.

  - [ ] **Step 4: Commit**

  ```bash verified:the commit subject is this task's Commit field
  git add skills/flow/integrate.md skills/flow/SKILL-rationale.md scripts/check-contract-budget.sh
  git commit -m "docs(flow): move integrate.md rationale to SKILL-rationale.md"
  ```

- [x] 5. Drop the subagent-driven-development invocation and the implementer's second principles read

**Build:** green

**Files:**
- Modify: `skills/flow/implement.md`

**Tests:** **none** — prose cut, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): drop the subagent-driven-development invocation and the second principles read`

  - [ ] **Step 1: Replace the invocation sentence**

  `skills/flow/implement.md:217` — `Invoke **superpowers:subagent-driven-development**, dispatching one implementer per bundle from:` becomes:

  ```markdown verified:authored in-tree for this change
  Dispatch one implementer per bundle from:
  ```

  Line 11's skill-sequence row and lines 182–183's override sentence stay: they record what
  `implement.md` overrides, and cutting them would leave the override unexplained.

  - [ ] **Step 2: Repoint the required reading at the bundle**

  Lines 252–253 — `` > **REQUIRED READING:** `engineering-principles.md` — your implementation must satisfy these `` / `> principles; the panel's principles reviewer checks the diff against them.` — become:

  ```markdown verified:authored in-tree for this change
  > **REQUIRED READING:** the engineering principles section of the context bundle below — your
  > implementation must satisfy these principles; the panel's principles reviewer checks the diff
  > against them.
  ```

  - [ ] **Step 3: Verify**

  Run: `scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-dispatch-paragraphs.sh && scripts/check-contract-budget.sh`
  Expected: each exits 0.

  - [ ] **Step 4: Commit**

  ```bash verified:the commit subject is this task's Commit field
  git add skills/flow/implement.md
  git commit -m "docs(flow): drop the subagent-driven-development invocation and the second principles read"
  ```

- [x] 6. Trim the principles slot brief

**Build:** green

**Files:**
- Modify: `skills/flow/principles-reviewer-prompt.md`

**Tests:** **none** — prose cut, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): trim the principles slot brief`

  - [ ] **Step 1: Delete the roster-history paragraph**

  Delete lines 6–9 — `**All three principle groups always apply**, and this template takes no
  parameter selecting among them. Earlier rosters made … to name which was in force.` — and the
  blank line after. The roster table in `review-panel.md:55` already states that all three groups
  always apply.

  - [ ] **Step 2: Drop the `project-configuration.md` cite**

  Lines 175–177 — `resolved to absolute paths **per the entry-form table and the` / `` containment rule in `skills/flow-contracts/project-configuration.md`**, `` / `which is canonical.` — become one clause: `resolved to absolute paths per the `## standards` entry-form table and containment rule.` The three bullets that follow (`Entries are not paths to use as-is: …`) stay, since they are the rule the slot applies.

  - [ ] **Step 3: Verify**

  Run: `scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-contract-budget.sh`
  Expected: each exits 0.

  - [ ] **Step 4: Commit**

  ```bash verified:the commit subject is this task's Commit field
  git add skills/flow/principles-reviewer-prompt.md
  git commit -m "docs(flow): trim the principles slot brief"
  ```

- [x] 7. Carry only flow's trigger in the managed block; drop `CLAUDE.md`'s `/flow commands summary`

**Build:** green

**Files:**
- Modify: `rules/flow-manual-review.mdc`
- Modify: `CLAUDE.md`
- Modify: `scripts/check-references.sh`

**Tests:** **none** — rule marker and prose cut, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(rules): carry only flow's trigger in the managed block`

  - [ ] **Step 1: Add the `core` markers to `rules/flow-manual-review.mdc`**

  Insert `<!-- core -->` on its own line after the frontmatter's closing `---` (line 5, so the
  marker becomes line 6, before the blank line and `# flow — /flow, start to finish`), and
  `<!-- /core -->` on its own line after the paragraph ending `stale copy of the state machine —
  the failure this split exists to prevent.` (line 41), before the blank line and `Narrower
  contracts sit beside it`. The other always-on rules place their markers exactly this way
  (`rules/be-brief.mdc:8,44`).

  - [ ] **Step 2: Delete `### /flow commands summary` from `CLAUDE.md`**

  Delete lines 88–136 — from the `### /flow commands summary` heading through the paragraph ending
  `and a gap found there routes back to a fix run anyway.` — and the blank line after, so
  `### How to invoke a skill` follows the skill index table directly.

  - [ ] **Step 3: Re-render the managed block and verify the install**

  Run: `./setup.sh global && scripts/check-installed-rules.sh && scripts/check-installed-citations.sh`
  Expected: `INSTALLED-RULES-OK` and the citations check clean. Then
  `grep -c 'Load it before' ~/.claude/CLAUDE.md` prints `0` — the contract table is no longer in
  the block — and `grep -c 'fixed three-state pipeline' ~/.claude/CLAUDE.md` prints `1`.

  - [ ] **Step 4: Verify the remaining guards**

  Run: `scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-contract-budget.sh && scripts/check-markdown-integrity.py && scripts/check-guard-symlinks.sh`
  Expected: each exits 0.

  - [ ] **Step 5: Commit**

  ```bash verified:the commit subject is this task's Commit field
  git add rules/flow-manual-review.mdc CLAUDE.md
  git commit -m "docs(rules): carry only flow's trigger in the managed block"
  ```

- [x] 8. Re-measure to the leaf and run every guard

**Build:** green

**Files:**
- Modify: `spectre/changes/kan-378-flow-deletion-only-cut-of-stale-restated-and/design.md`

**Tests:** **none** — measurement and verification, no code
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(spectre): record kan-378 after-figures per load set`

  - [ ] **Step 1: Write the measurement script to the scratchpad and run it**

  Save this as `measure.sh` in the session scratchpad (never in the repository), make it
  executable, and run `measure.sh <repo-root> after` from the branch tip. Running it with `before`
  at `e7d9540` reproduces the before column in `design.md`.

  ```bash verified:run at e7d9540 with the before argument; its output is design.md's before column
  #!/usr/bin/env bash
  # measure.sh <repo-root> before|after — sum wc -c per load set; tokens = bytes/4
  set -euo pipefail
  R="$1"; MODE="$2"
  SP=~/.claude/plugins/cache/claude-plugins-official/superpowers/6.3.0/skills
  C="$R/skills/flow-contracts"; F="$R/skills/flow"
  ALWAYS=(~/.claude/CLAUDE.md "$R/CLAUDE.md" "$R/rules/agent-baseline.md")
  SUB=("${ALWAYS[@]}" ~/.claude/projects/-Users-tweety53-Projects-agents/memory/MEMORY.md)
  if [ "$MODE" = before ]; then PLANNER="$F/brainstorm.md"; PRIN_EXTRA="$C/project-configuration.md"; IMPL_EXTRA="$F/engineering-principles.md"; SDD="$SP/subagent-driven-development/SKILL.md";
  else PLANNER="$F/brainstorm-planner.md"; PRIN_EXTRA=""; IMPL_EXTRA=""; SDD=""; fi
  A=("${ALWAYS[@]}" "$F/SKILL.md" "$C/pipeline.md" "$F/brainstorm.md" "$F/implement.md" "$F/review-panel.md" "$F/verify-and-handoff.md" "$F/principles-reviewer-prompt.md"
     "$C/jira-integration.md" "$C/operator-prompts.md" "$C/artifacts-registry.md" "$C/workspace-isolation.md" "$C/worktree-resolution.md" "$C/model-policy.md" "$C/project-configuration.md" "$C/git-boundaries.md" "$C/session-records.md"
     $SDD "$SP/requesting-code-review/SKILL.md" "$SP/using-git-worktrees/SKILL.md")
  B=("${SUB[@]}")
  Cset=("${SUB[@]}" "$PLANNER" "$SP/brainstorming/SKILL.md" "$SP/writing-plans/SKILL.md" "$C/plan-provenance.md" "$C/build-green.md")
  D=("${SUB[@]}" "$F/engineering-principles.md" $PRIN_EXTRA)
  E=("${SUB[@]}" "$F/engineering-principles.md" $IMPL_EXTRA "$SP/test-driven-development/SKILL.md")
  sum() { local t=0; for f in "$@"; do t=$((t + $(wc -c < "$f"))); done; printf '%d bytes ~%d tok\n' "$t" $((t/4)); }
  printf 'A creating-run parent: '; sum "${A[@]}"
  printf 'B per-subagent fixed:  '; sum "${B[@]}"
  printf 'C planner:             '; sum "${Cset[@]}"
  printf 'D principles slot:     '; sum "${D[@]}"
  printf 'E implementer:         '; sum "${E[@]}"
  ```

  The `~/.claude/CLAUDE.md` it reads is the live managed block, so task 7's `./setup.sh global`
  must have run first. Set B therefore counts the re-rendered block, which is what every subagent
  actually inherits.

  - [ ] **Step 2: Record the after column in `design.md`**

  In `## Measurement`, replace each `recorded by task 8` cell with the script's output for that
  set, and add a second provenance comment directly under the existing one:
  `<!-- measured: the measure.sh block in tasks.md task 8, run as measure.sh <repo> after @ branch spectre/kan-378-flow-deletion-only-cut-of-stale-restated-and -->`
  (write the branch name exactly as `git branch --show-current` prints it).

  - [ ] **Step 3: Run every configured lint command and the guard tests**

  Run every command under `## lint` in `.flow/project.md`, then `scripts/run-guard-tests.sh`.
  Expected: every command exits 0. `scripts/check-normative-inventory.sh` prints a set, not a
  verdict; compare its output against the same command at `e7d9540` (`git stash`-free: run it in a
  `git worktree add /tmp/kan378-base e7d9540` checkout) and confirm no `SHALL`/`MUST` sentence
  disappeared — every line in the base output is present in the branch output, the moved
  **No forking** sentences included.

  - [ ] **Step 4: Confirm the diff touches only the files this plan names**

  Run: `git diff --stat e7d9540 -- . ':!spectre/changes' ':!docs/superpowers/research'`
  Expected: exactly `CLAUDE.md`, `README.md`, `rules/flow-manual-review.mdc`,
  `scripts/check-contract-budget.sh`, `scripts/check-references.sh`, `skills/flow/SKILL.md`,
  `skills/flow/SKILL-rationale.md`, `skills/flow/brainstorm.md`, `skills/flow/brainstorm-planner.md`,
  `skills/flow/implement.md`, `skills/flow/integrate.md`, `skills/flow/principles-reviewer-prompt.md`,
  `skills/flow-contracts/pipeline.md`, `skills/flow-contracts/pipeline-rationale.md`.
  `scripts/check-references.sh` was added to this list mid-run: tasks 2 and 7 each needed it for
  the new-file/deleted-section reference-coverage guard (see their own `**Files:**` fields).

  - [ ] **Step 5: Commit**

  ```bash verified:the commit subject is this task's Commit field
  git add spectre/changes/kan-378-flow-deletion-only-cut-of-stale-restated-and/design.md
  git commit -m "docs(spectre): record kan-378 after-figures per load set"
  ```
