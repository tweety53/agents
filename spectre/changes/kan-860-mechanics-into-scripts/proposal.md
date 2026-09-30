# kan-860-mechanics-into-scripts

**Jira:** KAN-860 (epic KAN-851). **Audits:** the MECHANICS rows of
`docs/prompt-audit-2026-09-29/audit-*.md`. **Blocked by:** KAN-852 (the verbatim-move guard). It
is done, and KAN-855–KAN-859 (the session trims) have landed, so no moved text is edited twice.

## Why

The run-loaded files carry mechanical recipes that a session re-reads every turn and re-types into
shell calls: derivations, record sequences, diff writes, worktree snapshots and cleanup checks. A
script runs each one identically every time. The prompt keeps one call line and its exit contract.

## What changes

Each recipe moves into Go — `flow-guard` (`stats/internal/guard/`, with a `scripts/<name>.sh`
`flow_guard_exec` shim) or a `flow` CLI verb (`stats/cmd/flow/`) — with table tests that pin
parity against the recipe it replaces. The prompt keeps one call line plus the exit contract.
Every reworded prompt line is listed in `verbatim-moves.txt`.

The scope is the audits' medium-confidence MECHANICS rows plus the rows KAN-860 names
explicitly. Design.md lists which rows landed, by group, and which were left, with the reason.

| Group | Prompt files | Rows |
|---|---|---|
| Planning | `brainstorm.md`, `brainstorm-planner.md` | BP-21, BP-22, BP-23, BR-M1, MX4 |
| Implementation | `implement.md`, `review-panel-optional-slots.md` | MX1, MX2/OS05, MX3, MX6, MX7 |
| Review panel | `review-panel*.md`, the reviewer prompts | RP34, RP35, the late-fix trigger, `render-slot-prompt.sh` |
| Verify | `visual-verify.md`, `verify-and-handoff.md` | VV-18, VH-24, VH-25 |
| Finish | `integrate.md`, `archive.md`, `finish-contract-run2.md`, `flow-fast/SKILL.md` | MI1/MA2, MA1, MR2, FF1 |

## Out of scope

- Low-confidence rows the issue does not name (BR-M2, BR-M3, BP-24, MX5, RP36, RP37, JI-13,
  JI-14, VH-26, WI-12, MR1, MA3). Design.md gives the reasons.
- Every-session rows K18/P17. They belong to KAN-855's audit, not to KAN-860's scope list.
