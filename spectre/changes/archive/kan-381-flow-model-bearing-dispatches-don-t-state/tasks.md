# kan-381-flow-model-bearing-dispatches-don-t-state

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Two docs-only tasks in dependency order: state the rule canonically in Model resolution, then
instantiate it at every model-bearing dispatch site. `design.md` is canonical for every decision.
Every touched file is an owned `.md` under `skills/`, so every task runs
`scripts/check-contract-budget.sh` and, where the guard fails on a file, raises that file's
`budgets()` row in the same commit to the file's new size plus 25% — only because the guard
failed, never to make a number.

**Baseline, measured before any edit:**

- Green on this worktree before any edit: `check-contract-budget.sh`,
  `check-dispatch-paragraphs.sh`, `check-vocabulary.sh`, `check-references.sh`,
  `check-model-resolution-shell.sh`, `check-stage-mark-calls.sh`, `check-markdown-integrity.py`
  — all exit 0.
  <!-- measured: the seven guards run from this worktree before any edit @ branch spectre/kan-381-flow-model-bearing-dispatches-don-t-state -->
- `check-plan-shape.sh`, `check-plan-provenance.sh` and `check-task-build-green.sh` exit 1 before
  enrichment: they scan this change's own scaffold `tasks.md`, which carries no fields yet. All
  three are run at the close of writing-plans and must exit 0 there and stay 0 after every task.
  <!-- measured: the three guards run from this worktree before enrichment, each exit 1 @ branch spectre/kan-381-flow-model-bearing-dispatches-don-t-state -->
- Byte sizes the budget ratchet starts from — skills/flow/SKILL.md 14687, brainstorm.md 11908,
  implement.md 31948, review-panel.md 50049, verify-and-handoff.md 31337, archive.md 18255,
  flow-research/SKILL.md 11759; tightest headroom is archive.md (row 18748) and implement.md
  (row 32500).
  <!-- measured: wc -c over the seven files, and the two rows read from scripts/check-contract-budget.sh @ branch spectre/kan-381-flow-model-bearing-dispatches-don-t-state -->
- 56 guard harnesses under `scripts/test-*.sh`; none added or changed by this plan.
  <!-- measured: ls scripts/test-*.sh | wc -l @ branch spectre/kan-381-flow-model-bearing-dispatches-don-t-state -->
- The normative inventory holds 9 sentences; it is captured before the first edit and must be
  byte-identical after the last — the added wording avoids `SHALL`/`MUST`, so no sentence enters
  or leaves the set.
  <!-- measured: scripts/check-normative-inventory.sh, 9 lines @ branch spectre/kan-381-flow-model-bearing-dispatches-don-t-state -->

---

- [x] 1. State the resolved-model rule canonically in Model resolution

**Build:** green — docs-only; every verification command runs at this task's state.
**Files:** skills/flow/SKILL.md
**Allowed-collateral:** scripts/check-contract-budget.sh
**Tests:** **none**
**Regression:** none — no tests declared; the guards named in the verify step re-run green after
this task, and the normative inventory stays byte-identical.
**Baseline:** before=56 after=56 guard harnesses under `scripts/test-*.sh`; none added.
**Commit:** feat(flow): state the resolved-model rule in model resolution

  - [x] **Step 1: Append the canonical statement to Model resolution**

    In `skills/flow/SKILL.md`, append this paragraph at the end of the **Model resolution**
    section — after the `DEFAULT_MODEL` override paragraph, before `## Reading the state`. Prose
    only: no second fenced `bash` block may appear under that heading, which
    `check-model-resolution-shell.sh` extracts by.

```markdown verified:authored in-tree for this change
**Every model-bearing dispatch states its resolved value in the run's own output, immediately
before the dispatch** — one line, `<role> model: <value> (<tier>)`. `<tier>` names the tier that
produced the value: `session override`, `project key`, `store`, or `fallback`; `fixed literal`
for `VERIFY_MODEL`, which is never resolved; `unknown (agent-defined)` for Bugbot and Security
dispatched by their own `subagent_type`, matching the dispatch record's `-model` literal; and the
model actually given for a substituted slot. A handshake-mismatch re-dispatch states again before
each re-dispatch. The line is what makes the resolution an observable claim: a dispatcher holding
a stale value prints the stale value, which an operator can challenge mid-run — the handshake
cannot catch that class, because a dispatcher holding the wrong value expects the wrong line
(KAN-381's near-miss). Each dispatch site states its own role token in place and cites this rule;
the rule is not restated there.
```

  - [x] **Step 2: Run the task's lint set**

    Run: `scripts/check-model-resolution-shell.sh && scripts/check-contract-budget.sh && scripts/check-vocabulary.sh && scripts/check-references.sh && python3 scripts/check-markdown-integrity.py`
    Expected: all exit 0. A `check-contract-budget.sh` failure on `skills/flow/SKILL.md` is
    remedied by raising that row in the same commit, per the header rule.

  - [x] **Step 3: Commit**

```bash verified:authored in-tree for this change
git add skills/flow/SKILL.md
git add scripts/check-contract-budget.sh   # only when the budget row moved
git commit -m "feat(flow): state the resolved-model rule in model resolution"
```

- [x] 2. Instantiate the statement at every model-bearing dispatch site

**Build:** green — docs-only; every verification command runs at this task's state.
**Files:** `skills/flow/archive.md`, `skills/flow/brainstorm.md`, `skills/flow/implement.md`,
`skills/flow/review-panel.md`, `skills/flow/verify-and-handoff.md`, `skills/flow-research/SKILL.md`,
`scripts/check-model-statements.sh`, `scripts/test-check-model-statements.sh`,
`scripts/check-guard-symlinks.sh`, `.flow/project.md`
**Allowed-collateral:** scripts/check-contract-budget.sh
**Tests:** `scripts/test-check-model-statements.sh`
**Regression:** reverting the commit removes the model-statement teeth — deleting a statement
line is green again, which F1's reproducer demonstrates. `check-dispatch-paragraphs.sh`'s
required paragraphs are untouched, and the normative inventory stays byte-identical.
**Baseline:** before=56 after=57 guard harnesses under `scripts/test-*.sh`; one added
(`test-check-model-statements.sh`, 8 cases — case 8 deletes each row's own literal, read from the guard's source, so a dropped row shrinks coverage loudly).
**Commit:** feat(flow): state the resolved model before every dispatch

  - [x] **Step 1: archive.md — self-review dispatch (step 9)**

    Open the paragraph beginning `**On anything but No, the combined reasoning pass runs as a
    subagent, on \`SELF_REVIEW_MODEL\`**` and insert this sentence as its opening sentence:

```markdown verified:authored in-tree for this change
State `self-review model: <value> (<tier>)` — the resolved `SELF_REVIEW_MODEL` and the tier that
produced it, per **Model resolution** (`skills/flow/SKILL.md`) — in this session's own output
immediately before the dispatch and again before the mismatch re-dispatch.
```

  - [x] **Step 2: brainstorm.md — planner dispatch**

    Insert immediately before the paragraph beginning `Dispatch one subagent with the Agent
    tool's \`model\` parameter set to \`PLANNING_MODEL\``:

```markdown verified:authored in-tree for this change
State `planner model: <PLANNING_MODEL> (<tier>)` — the tier that produced the value, per
**Model resolution** (`skills/flow/SKILL.md`) — in this run's own output immediately before the
dispatch.
```

  - [x] **Step 3: implement.md — conductor dispatch**

    Insert immediately before the conductor-dispatch paragraph that sets the Agent tool's
    `model` parameter to `DEFAULT_MODEL`:

```markdown verified:authored in-tree for this change
State `conductor model: <value> (<tier>)` — the resolved `DEFAULT_MODEL` (or this run's recorded
session override) and the tier that produced it, per **Model resolution**
(`skills/flow/SKILL.md`) — in the run's own output immediately before the dispatch.
```

  - [x] **Step 4: implement.md — `flow.document-fix` planner dispatch**

    Insert immediately before that section's planner dispatch (the one recorded with
    `-role planner`):

```markdown verified:authored in-tree for this change
State `planner model: <PLANNING_MODEL> (<tier>)` — the tier that produced the value, per
**Model resolution** (`skills/flow/SKILL.md`) — in the run's own output immediately before the
dispatch.
```

  - [x] **Step 5: review-panel.md — panel slots and the panel-fix subagent**

    Insert one sentence beside the paragraph beginning `**Every slot's dispatch is recorded**`:

```markdown verified:authored in-tree for this change
State each dispatched slot's model in the run's own output immediately before its dispatch —
`<slot> model: <value> (<tier>)`, tier per **Model resolution** (`skills/flow/SKILL.md`); a slot
dispatched by its own `subagent_type` states `unknown (agent-defined)`, and a substituted slot
states the model actually given.
```

    And insert one sentence immediately before the fix-subagent dispatch paragraph (**Dispatch it
    on `DEFAULT_MODEL`**):

```markdown verified:authored in-tree for this change
State `panel-fix model: <DEFAULT_MODEL> (<tier>)` — the tier that produced the value, per
**Model resolution** (`skills/flow/SKILL.md`) — in the run's own output immediately before the
dispatch.
```

  - [x] **Step 6: verify-and-handoff.md — both verifier dispatches**

    Insert one sentence immediately before each stage's dispatch (the **Verify** stage's and
    **Visual verification**'s), each naming its own stage:

```markdown verified:authored in-tree for this change
State `verify model: sonnet (fixed literal)` — `VERIFY_MODEL` is the fixed literal `sonnet`, read
from neither the settings store nor the project key (**Model resolution**,
`skills/flow/SKILL.md`) — in the run's own output immediately before this stage's dispatch.
```

  - [x] **Step 7: flow-research/SKILL.md — research subagent dispatch**

    Insert immediately before the research dispatch (the subagent launched with the Agent tool's
    `model` parameter set to the resolved planning model):

```markdown verified:authored in-tree for this change
State `research model: <PLANNING_MODEL> (<tier>)` — the tier that produced the value, per
**Model resolution** (`skills/flow/SKILL.md`) — in the session's own output immediately before
the dispatch.
```

  - [x] **Step 8: Run the task's lint set**

    Run: `scripts/check-contract-budget.sh && scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-dispatch-paragraphs.sh && scripts/check-stage-mark-calls.sh && python3 scripts/check-markdown-integrity.py`
    Expected: all exit 0. A `check-contract-budget.sh` failure on any touched file is remedied by
    raising that file's row in the same commit, per the header rule.

  - [x] **Step 9: Commit**

```bash verified:authored in-tree for this change
git add skills/flow/archive.md skills/flow/brainstorm.md skills/flow/implement.md \
  skills/flow/review-panel.md skills/flow/verify-and-handoff.md skills/flow-research/SKILL.md
git add scripts/check-contract-budget.sh   # only when a budget row moved
git commit -m "feat(flow): state the resolved model before every dispatch"
```
