# compact-decision-block

## Why

Operator, 2026-10-02 (gymie KAN-870): the dynamic decisions were not visible at the plan's end, and
the `## Decision` block — two tables of rolls, rules and reasons — is too long to read.

## What changes

- `flow decision render` prints `## Decision` as one short line per choice: class (with a raise
  reason), execution with implementer/fixer pairs, groups as bundle ranges with a reason only where
  a pair departs from the implementer's plus the split reason, panel (shape, each dispatch's slots
  and pair, the rerun pair, a bundle-cap skip; `default` with its roster on micro), and visual
  verification (with its reason unless `required`). Rolls, rule cells and the `planning:`/
  `reviewers:` preamble are gone.
- The `STARTED` handoff becomes `## Plan ready`: a 2–4 line summary of what the plan implements,
  the latest decision's render, an `Open questions` line and a `Jira` line only when they apply,
  and `/clear` + `/flow <name>`. The counts line, the pre-edit Jira description echo and the
  IntelliJ command are gone (`skills/flow/brainstorm.md`, `skills/flow-contracts/handoff-blocks.md`,
  `skills/flow-contracts/jira-integration.md`, `skills/flow-contracts/pipeline.md`).
- `pipeline.md`'s handoff rules: a red check the run executes is fixed in that run, never handed
  off as a question (operator, 2026-10-02).
- `setup.sh`: every `printf "$body" | <early-exiting reader>` becomes a here-string. Under
  `pipefail` the reader's early exit SIGPIPEd printf (141), so `stats/internal/setuptest` failed
  intermittently on `origin/main` (3 of 3 runs); 0 of 8 after.
- Every mention of the preamble or the rule cell is updated (`skills/flow/brainstorm-planner.md`,
  `skills/flow/document-fix.md`, `skills/flow-fast/SKILL.md`).
