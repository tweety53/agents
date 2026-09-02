# kan-378-flow-deletion-only-cut-of-stale-restated-and — design

## Context

Seeded from the staged research note `docs/superpowers/research/kan-378.md` (deleted on adoption,
per **C** of the planner's brainstorm file). The note's measurement thread, its parent-session and
dispatch-fan-out findings, and its step-by-step breakdown are folded into this file; its second
change (mechanics-to-code) is KAN-379 and out of scope here. The note's superpowers figures were
taken against an older plugin; the installed superpowers 6.3.0 skills are what this file measures.

Brainstorming classified this as bounded: every file to cut already exists and no new flow is
introduced.

## Scope rule

Deletion-only. A passage is deleted when it is stale or restates what another run-loaded file
states canonically; it is moved verbatim when it records why a rejected alternative was rejected.
Nothing is paraphrased: a normative sentence re-said differently is a changed requirement no guard
detects, where a deletion is visible in the diff. What survives unchanged: states, transitions,
stage keys, guard names, field names, exit codes, ordering constraints, operator-prompt wording.

## Cuts, per file

- `skills/flow/SKILL.md` — the "not yet a row in that file's own transition table" pointer language
  in the load-first paragraph is replaced by one sentence pointing at `pipeline.md`'s table; the
  **State transitions** section is deleted with its `**No flags.**` paragraph kept; in **Stage
  keys** the design-decision paragraph and the stale known-gap paragraph are cut and the key table
  stays, since `README.md`, `commands/flow.md` and `commands-claude/flow.md` cite that heading; the
  "`model-policy.md` is only partly current" caveat is deleted; the later `model-policy.md` cite for
  the retired per-change fields stays.
- `skills/flow-contracts/pipeline.md` — the `/flow` row of **State transitions** becomes one row
  per accepted state, carrying SKILL.md's row text verbatim and keeping the **Stage keys** pointer.
- Project `CLAUDE.md` — `### /flow commands summary` is deleted whole, including the paragraph that
  declared the digest a deliberate copy; the skill index, the mandatory rules, **How to invoke a
  skill** and **Superpowers general skills** stay.
- `skills/flow/brainstorm.md` — sections B, C and D move verbatim to
  `skills/flow/brainstorm-planner.md`; the dispatch prompt's "read this file's sections B, C and D"
  names the new file; **Convergence** citations in `pipeline.md`, `pipeline-rationale.md` and
  `README.md` repoint to the new file.
- `skills/flow/implement.md` — "Invoke **superpowers:subagent-driven-development**, dispatching one
  implementer per bundle" becomes "Dispatch one implementer per bundle"; the skill-sequence table
  row and the parallel-dispatch override sentence that name the skill stay, since they state what
  `implement.md` overrides; the `REQUIRED READING` paragraph points at the bundle's engineering
  principles section.
- `skills/flow/principles-reviewer-prompt.md` — the roster-history paragraph is deleted; the
  `[STANDARDS_PATHS]` step keeps its entry-form and containment rule and loses the
  `project-configuration.md` file cite.
- Rationale — every line in `SKILL.md`, `brainstorm.md` and `integrate.md` that cites a design.md
  decision id, records "this task's own" choice, or explains a rejected alternative moves verbatim
  to `skills/flow/SKILL-rationale.md`, under a heading naming its source file and section; the
  source keeps only the normative clause. `workspace-isolation.md` and `project-configuration.md`
  already point at their own siblings and need no move.
- `rules/flow-manual-review.mdc` — `core` markers enclose the title, the three-line trigger, its two
  explanatory paragraphs, and **Load the pipeline before acting** through the "Never act on a
  remembered version" paragraph; the contract table and the installation note remain in the full
  rule only. `./setup.sh global` re-renders the managed block.
- `scripts/check-contract-budget.sh` — two rows added at landed size plus a quarter; no row lowered.

## Measurement

Load sets, each a sum of `wc -c` over the files that set reads, tokens approximated as bytes
divided by four. The script that produced both columns is the fenced block in `tasks.md` task 8;
before-figures were taken at `e7d9540` (main, before any edit), after-figures at the branch tip.

| Set | Files | Before | After |
|---|---|---|---|
| A — creating-run parent | always-on (`~/.claude/CLAUDE.md`, project `CLAUDE.md`, `rules/agent-baseline.md`) + router (`skills/flow/SKILL.md`, `pipeline.md`) + `brainstorm.md`, `implement.md`, `review-panel.md`, `verify-and-handoff.md`, `principles-reviewer-prompt.md` + every contract those cite at point of use + superpowers `subagent-driven-development` (before only), `requesting-code-review`, `using-git-worktrees` | 342387 bytes, ~85596 tok | 281245 bytes, ~70311 tok |
| B — per-subagent fixed overhead | `~/.claude/CLAUDE.md`, project `CLAUDE.md`, `MEMORY.md`, `rules/agent-baseline.md` | 32086 bytes, ~8021 tok | 25694 bytes, ~6423 tok |
| C — planner | B + `brainstorm.md` (before) / `brainstorm-planner.md` (after) + superpowers `brainstorming`, `writing-plans` + `plan-provenance.md`, `build-green.md` | 114221 bytes, ~28555 tok | 97362 bytes, ~24340 tok |
| D — principles slot | B + `engineering-principles.md` + `project-configuration.md` (before only) | 85770 bytes, ~21442 tok | 34679 bytes, ~8669 tok |
| E — implementer | B + `engineering-principles.md` once (after) or twice (before) + superpowers `test-driven-development` | 59071 bytes, ~14767 tok | 43694 bytes, ~10923 tok |
<!-- measured: the measure.sh block in tasks.md task 8, run as `measure.sh <repo> before` @ e7d9540 -->
<!-- measured: the measure.sh block in tasks.md task 8, run as measure.sh <repo> after @ branch spectre/kan-378-flow-deletion-only-cut-of-stale-restated-and -->

Set B is the same bytes for every dispatch; after this change it shrinks by what the `core` marker
removes from the managed block and by `CLAUDE.md`'s deleted summary. Set A's largest single cut is
the `subagent-driven-development` skill no longer invoked.

## Verification

- Every command under `## lint` in `.flow/project.md` exits clean after every task, and
  `scripts/run-guard-tests.sh` passes.
- `./setup.sh global` re-rendered; `scripts/check-installed-rules.sh` and
  `scripts/check-installed-citations.sh` clean against the live install.
- `scripts/check-references.sh` clean — every cited heading, including the repointed **Convergence**
  and the kept **Stage keys**, still resolves.
- `git diff main --stat` shows only the files this design names.

## Decisions

### Where `/flow`'s per-state transition rows live

**ID:** transitions-move-to-pipeline
**Status:** active
**Chosen:** expand `pipeline.md`'s `/flow` row into five per-state rows, delete SKILL.md's
**State transitions** section and its dangling pointer language — `pipeline.md` becomes the sole
canonical source, as the issue intends.
**Considered:** keep SKILL.md's table and cut only `CLAUDE.md`'s copy — rejected because it leaves
three copies (global stub, SKILL.md, `pipeline.md`'s pointer) and `pipeline.md` would still not be
canonical for its own command.

### Where rationale from files with no `-rationale.md` sibling goes

**ID:** flow-rationale-single-appendix
**Status:** active
**Chosen:** one new `skills/flow/SKILL-rationale.md` for the whole `skills/flow/` skill, covering
`SKILL.md`, `brainstorm.md` and `integrate.md`; matches the `skills/<name>/SKILL-rationale.md`
shape `check-contract-budget.sh` already names; one new budget row.
**Considered:** one sibling per phase file — rejected as three files and three budget rows for
the same content; appending to `skills/flow-contracts/pipeline-rationale.md` — rejected because it
mixes skill-phase reasoning into a contract's appendix.

### The slot-brief item is already satisfied

**ID:** slot-briefs-already-clean
**Status:** active
**Chosen:** record that no fenced slot prompt carries the wall-clock ceiling, the no-forking rule
or the re-dispatch procedure — all three live only in `review-panel.md`, which the dispatcher alone
reads — and cut only `principles-reviewer-prompt.md`'s roster-history paragraph and its
`project-configuration.md` cite.
**Considered:** also moving the `**Placeholders:**` block into `review-panel.md` — rejected because
it already sits outside the fenced prompt and is read only by the dispatcher, so the move saves
nothing per dispatch.

### The planner file's name

**ID:** planner-file-name
**Status:** active
**Chosen:** `skills/flow/brainstorm-planner.md` — keeps the `brainstorm` prefix so the relationship
to `brainstorm.md` is visible in the filename.
**Considered:** `skills/flow/plan.md` — rejected as hiding which phase file it was split from.

## Open questions

