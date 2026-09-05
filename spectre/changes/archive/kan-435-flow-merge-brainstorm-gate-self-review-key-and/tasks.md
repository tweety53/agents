# kan-435-flow-merge-brainstorm-gate-self-review-key-and

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Spec:** `design.md` beside this file — canonical for every replacement wording below; where a
step quotes new text, `design.md`'s wording wins on any difference. Three tasks, every one
`Build: green`, all documentation, no test added. Tasks 2 and 3 both edit `skills/flow/archive.md`
and `skills/flow-contracts/finish-contract-run2.md`, so they run in order. Every passage to replace
is located by its opening words, never a line number: `skills/flow/brainstorm.md` carries an
unrelated uncommitted edit in the main checkout that may land before this branch does. Byte
figures below were taken on `main` before this change; each task's verify step names the check.

- [x] 1. Merge the brainstorm convergence confirm and the design approval

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`, `skills/flow/brainstorm.md`, `skills/flow-contracts/pipeline.md`, `skills/flow-contracts/pipeline-rationale.md`
**Tests:** none — documentation; the guards in step 6 are the check
**Regression:** none — no test is added; `scripts/check-references.sh`, `scripts/check-stage-mark-calls.sh` and `scripts/check-contract-budget.sh` pass before and after
**Baseline:** before=0 after=0
<!-- measured: no test suite covers skills/ prose; the count is the number of tests this task touches @ branch main -->
<!-- predicted: unchanged after this task, on branch spectre/kan-435-flow-merge-brainstorm-gate-self-review-key-and -->
**Commit:** `docs(flow): merge the brainstorm confirm and the design approval`

  - [x] **Step 1: The confirm.** In `skills/flow/brainstorm-planner.md` **Convergence**, replace
    the blockquote opening `> **That is everything I have settled.` (two options today) with:

```markdown verified:authored in-tree for this change; design.md §1
> **That is everything I have settled. Anything still unclear before I move on?**
> - **Nothing unclear — approve the design and move on** *(recommended)*
> - **Another round — I have something** *(default — anything short of an explicit "approve and
>   move on" is treated as this)*
> - **Revise — I have a change to the design**
```

    In the paragraph that follows, `End the stage only on an explicit choice of **move on**.`
    becomes `End the stage only on an explicit choice of **approve the design and move on**.`
    The rest of that paragraph — the safe-default/recommended split, the `⚠ another round — no
    explicit answer` line — is unchanged.

  - [x] **Step 2: Revise.** Directly after that paragraph, add one paragraph: *Revise* is a round
    — it counts toward the third-round offer below exactly as *Another round* does — and differs
    only in what the planner's next turn opens with: the changed design section(s), re-presented
    before the next confirm, in place of new questions.

  - [x] **Step 3: One answer, two roles.** Replace the paragraph opening `The convergence loop's
    own exit — an explicit **move on** at the confirm above — is what closes the checklist
    itself;` with: the explicit **approve the design and move on** answer is at once the
    convergence exit that closes the checklist and the design approval the HARD GATE requires;
    the parent still marks `flow.brainstorm` end, then `flow.design-approval` begin and end,
    around that one relayed answer. Keep the four-line `bash` block that follows; change only its
    comment line to `# … the operator's approve-and-move-on answer, relayed — this is the HARD
    GATE above …`. In **The checklist**, the bullet `- **HARD GATE:** do not run \`spectre new\`
    until the user approves the design.` gains: ` Approval is the merged confirm's first option
    under **Convergence** below; no separate approval question is asked.`

  - [x] **Step 4: The relay.** In `skills/flow/brainstorm.md` **The relay**, replace `Section B's
    convergence confirm, its third-round offer, and the HARD GATE design-approval question are
    relayed the same way` with `Section B's merged convergence-and-approval confirm and its
    third-round offer are relayed the same way`, and the mark block's comment line
    `# … the operator approves the design, relayed through the planner's HARD GATE question …`
    with `# … the operator's approve-and-move-on answer, relayed through the planner's merged
    confirm — the HARD GATE …`. The paragraph about relaying prose before a `## Question` block,
    if present, is untouched.

  - [x] **Step 5: Stage exit and its rationale.** In `skills/flow-contracts/pipeline.md` **Stage
    exit — never the command's own judgment**, after `An operator who is present but silent is
    not that exception and still gets another round.` add one sentence: the same explicit answer
    may both close the checklist and grant the design approval, as **Convergence**
    (`skills/flow/brainstorm-planner.md`) defines. In `skills/flow-contracts/pipeline-rationale.md`
    **Stage exit — never the command's own judgment**, add one sentence after the existing
    paragraph: `flow.design-approval`'s median wall clock was 0 s over 49 runs in this
    repository's dev stats store on 2026-09-04 — the approval was a reflex seconds after the
    confirm — which is why one answer now serves as both.
<!-- measured: docs/superpowers/research/flow-speedup.md §7 round two, from /api/v1/stats/stage-leaderboard on the dev store @ 2026-09-04 -->

  - [x] **Step 6: Verify.** From the repository root run `scripts/check-references.sh`,
    `scripts/check-stage-mark-calls.sh`, `scripts/check-markdown-integrity.py`,
    `scripts/check-vocabulary.sh` and `scripts/check-contract-budget.sh` — all exit 0.
    Headroom on `main`: `brainstorm-planner.md` 18619 of 23248, `brainstorm.md` 10006 of 35015,
    `pipeline.md` 24595 of 36155, `pipeline-rationale.md` 14706 of 20935; if
    `check-contract-budget.sh` nevertheless trips, raise that row's budget by the overage and say
    so in the commit body. Commit.
<!-- measured: wc -c on the four files and grep of the budgets() table in scripts/check-contract-budget.sh @ branch main -->

- [x] 2. Add the `## self review` key and set this repository to `skip`

**Build:** green
**Files:** `skills/flow-contracts/project-configuration.md`, `skills/flow/archive.md`, `skills/flow-contracts/finish-contract-run2.md`, `skills/flow-contracts/handoff-blocks.md`, `.flow/project.md`
**Tests:** none — documentation; the guards in step 6 are the check
**Regression:** none — no test is added; `scripts/check-references.sh`, `scripts/check-model-keys.sh` and `scripts/check-contract-budget.sh` pass before and after
**Baseline:** before=0 after=0
<!-- measured: no test suite covers skills/ or .flow/ prose; the count is the number of tests this task touches @ branch main -->
<!-- predicted: unchanged after this task, on branch spectre/kan-435-flow-merge-brainstorm-gate-self-review-key-and -->
**Commit:** `docs(flow-contracts): add the self review project key`

  - [x] **Step 1: The contract row.** In `skills/flow-contracts/project-configuration.md`'s key
    table, insert this row directly after the `## default landing route` row:

```markdown verified:column layout copied from the key table's `## default landing route` row in skills/flow-contracts/project-configuration.md on main
| `## self review` | Optional. One of the literal bodies `run` or `skip` — this section holds that value and nothing else, never free-form prose, matching `## default landing route`'s own single-line-literal shape. Resolved by run 2's step 9 (`skills/flow/archive.md`) before its skip prompt: `skip` skips self-review without asking, `run` runs it without asking, absent asks as today. A body matching neither literal exactly is reported by name and dropped, resolving as if the key were absent. |
```

    In the paragraph opening `**\`## planning model\`'s body is matched exactly as`, append one
    sentence: `## self review`'s body is matched the same way — two literals in place of three,
    and the same report-by-name-and-drop. `scripts/check-model-keys.sh` is not extended.

  - [x] **Step 2: Resolve the key in archive.md step 9.** In `skills/flow/archive.md` step 9,
    directly before the line `The skip prompt fires first:`, add a paragraph: run
    `project-get.sh <main-checkout> "self review"` (exit 1: absent) and match the body against
    the two literals `run` / `skip` byte-for-byte after trimming leading/trailing whitespace; a
    body matching neither is reported by name and dropped, resolving as absent. `skip` ends step 9
    here, the handoff's `Self-review` line reading `skipped — project default`. `run` proceeds to
    the reasoning pass below with no prompt. Absent: the skip prompt fires as today. Then change
    `The skip prompt fires first:` to `When the key is absent, the skip prompt fires first:`.

  - [x] **Step 3: The handoff line and the guardrail.** In `skills/flow/archive.md`'s handoff
    template, replace `**Self-review:** <path> (rating: <n>/5) | skipped` with
    `**Self-review:** <path> (rating: <n>/5) | skipped | skipped — project default`. In
    `skills/flow-contracts/handoff-blocks.md`, the sentence carrying the same
    `**Self-review:** <path> (rating: <n>/5) | skipped` literal gets the same third alternative.
    In `archive.md`'s guardrails, the bullet `- **Never** ask the self-review skip prompt, a
    per-angle filing ask, or the rating question before \`FINISHED\` has been written.` becomes
    `- **Never** ask the self-review skip prompt or the filing-and-rating prompt, nor resolve the
    \`## self review\` key, before \`FINISHED\` has been written.`

  - [x] **Step 4: The canonical statement.** In `skills/flow-contracts/finish-contract-run2.md`
    step 9, after `It is skippable per run, with running it the default.` add: a project's
    `## self review` key (**Project configuration**,
    `skills/flow-contracts/project-configuration.md`) decides without asking when present and
    valid; the per-run prompt is the absent case.

  - [x] **Step 5: This repository.** In `.flow/project.md`, insert directly after the
    `## default landing route` section (before `## self review model`):

```markdown verified:authored in-tree for this change; the counts are docs/superpowers/research/flow-speedup.md §8 "Self-review key" on main
## self review

`skip`

The report series ended at kan-380: the six changes after it all answered "No" to a prompt that
fires after `FINISHED`, when the operator has walked away, and the 30 reports before it yielded 9
Jira tickets. This key ratifies that and ends the series; set `run` to bring it back.
```
<!-- measured: docs/superpowers/research/flow-speedup.md §8, from ls docs/self-review and the reports' filed-issue lines @ 2026-09-04 -->

  - [x] **Step 6: Verify.** From the repository root run `scripts/check-references.sh`,
    `scripts/check-model-keys.sh`, `scripts/check-markdown-integrity.py`,
    `scripts/check-workspace-isolation.sh`, `scripts/check-contract-budget.sh` and
    `scripts/project-get.sh . "self review"` — the first five exit 0; the last prints the section
    body (the literal and the paragraph) and exits 0. Headroom on `main`:
    `project-configuration.md` 46575 of 48175, `archive.md` 16420 of 18748,
    `finish-contract-run2.md` 30129 of 36019, `.flow/project.md` 25224 of 26450 — the first and
    last are the tight ones; if `check-contract-budget.sh` trips, raise that row's budget by the
    overage and say so in the commit body. Commit.
<!-- measured: wc -c on the files and grep of the budgets() table in scripts/check-contract-budget.sh @ branch main -->

- [x] 3. Collapse the self-review filing prompts and the rating into one call

**Build:** green
**Files:** `skills/flow-contracts/finish-contract-run2.md`, `skills/flow/archive.md`, `skills/flow-contracts/operator-prompts.md`
**Tests:** none — documentation; the guards in step 4 are the check
**Regression:** none — no test is added; `scripts/check-references.sh`, `scripts/check-self-review-report.sh` and `scripts/check-contract-budget.sh` pass before and after
**Baseline:** before=0 after=0
<!-- measured: no test suite covers skills/ prose; the count is the number of tests this task touches @ branch main -->
<!-- predicted: unchanged after this task, on branch spectre/kan-435-flow-merge-brainstorm-gate-self-review-key-and -->
**Commit:** `docs(flow-contracts): one filing-and-rating prompt for self-review`

  - [x] **Step 1: The canonical shape.** In `skills/flow-contracts/finish-contract-run2.md` step
    9, replace the paragraph opening `The filing ask is **one multi-select prompt per angle**,`
    with: the filing ask and the rating are **one `AskUserQuestion` call**. Findings from every
    angle fill up to three multi-select questions, each carrying at most three findings — every
    option prefixed with its angle's label — plus **None — file nothing** as that question's
    `(default, recommended)` option; the rating is the call's last question, options
    `5 — excellent` / `4 — good` / `3 — fine` / `2 — rough`, a `1` typed through the tool's
    free-text "Other". More than nine findings roll the overflow into one further call of the same
    shape, without the rating. Shape per **Operator prompts**
    (`skills/flow-contracts/operator-prompts.md`). Keep the sentence about a filed finding carrying
    its angle's label. Earlier in the step, `the operator's rating and the per-angle filing
    prompts still run in that session` becomes `the filing-and-rating prompt still runs in that
    session`, and in the openspec-provenance paragraph `the per-angle filing ask` becomes `the
    filing ask`.

  - [x] **Step 2: The executing wording.** In `skills/flow/archive.md` step 9, replace from `The
    filing ask is **one multi-select prompt per angle**, defaulting to filing none:` through the
    line `Then ask the operator to rate the run: **Rate this flow run, 1 (rough) to 5
    (excellent):**` with:

```markdown verified:authored in-tree for this change; design.md §3
   The filing ask and the rating are **one `AskUserQuestion` call** — one multi-select question per
   three findings, each option `<label>: <finding>`, the rating last:

   > **File any of these findings as Jira issues?**
   > - **`<label>`: <finding 1>**
   > - **`<label>`: <finding 2>**
   > - **`<label>`: <finding 3>**
   > - **None — file nothing** *(default, recommended)*
   >
   > **Rate this flow run:**
   > - **5 — excellent**
   > - **4 — good**
   > - **3 — fine**
   > - **2 — rough** — a `1` is typed through the tool's free-text "Other"

   More than nine findings roll the overflow into one further call without the rating.
```

    In the same step, `does not ask the rating question, and does not run the filing prompts`
    becomes `does not run the filing-and-rating prompt`.

  - [x] **Step 3: The multi-select variant.** In `skills/flow-contracts/operator-prompts.md`
    **The multi-select variant**, replace the last sentence — from `The two existing multi-select
    call sites choose oppositely:` to `(silence selects every fired trigger).` — with: `The one
    live multi-select call site, the self-review filing ask, chooses the empty set: silence
    selects **None — file nothing**.` The `skills/myflow-do/SKILL.md` reference goes with it.

  - [x] **Step 4: Verify.** From the repository root run `scripts/check-references.sh`,
    `scripts/check-self-review-report.sh`, `scripts/check-markdown-integrity.py`,
    `scripts/check-vocabulary.sh` and `scripts/check-contract-budget.sh` — all exit 0.
    `operator-prompts.md` is 2021 of 2432 on `main` and this step shrinks it; if
    `check-contract-budget.sh` trips on `archive.md` or `finish-contract-run2.md`, raise that
    row's budget by the overage and say so in the commit body. Commit.
<!-- measured: wc -c on the files and grep of the budgets() table in scripts/check-contract-budget.sh @ branch main -->
