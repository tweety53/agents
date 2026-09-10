# kan-482-flow-conductor-dispatches-one-fix-subagent-per

## Why

During the KAN-449 `/flow` run, the conductor subagent launched four separate background fix
subagents at the review-panel fix step — one per reviewer's findings — against the contract
`skills/flow/review-panel.md` already states: give the surviving findings to **one** fix subagent
as the combined list. Splitting the fix fragmented a single-diff fix into independent fixups
against the same worktree, risked conflicting commits and duplicated work, and broke the relay
contract's no-child-in-flight rule. The operator caught it mid-run. Instructions alone failed
once; the obligation needs a mechanical check too (KAN-482).

## What changes

- `skills/flow/review-panel.md`'s fix step states the full contract: exactly one panel-fix
  dispatch per fix round carrying the combined list of every surviving open finding, awaited in
  the foreground before the round's reproducer re-runs, with the canonical `-key` shape
  (`panel-fix-<round>`, retry on `panel-fix-<round>-retry`) and a pre-dispatch self-check.
- `skills/flow/implement.md`'s conductor relay contract carries the same constraint, so the
  conductor reads it before the panel stage.
- A new read command `flow record dispatches -change <name>` prints a change's dispatch rows as
  JSON — the store already persists them; only the read was missing. No `-session-token`/`-role`
  filters: the guard filters rows itself with jq, so the verb stays `-change`-only (amended at
  implementation, F1).
- A new guard `check-panel-fix-single-dispatch.sh <worktree> <change> <session-token>` runs at the
  panel's close beside `check-panel-findings-closed.sh` and fails when any round of the run
  recorded more than the allowed panel-fix dispatches or a panel-fix key outside the canonical
  shape; exit 1 hands back to the operator.
- `skills/flow/SKILL.md`'s guard list registers the new script so the guard-presence check sees it.

Observable difference: a run whose conductor over-dispatches fix subagents is caught at the panel
close gate instead of relying on the operator noticing mid-run.
