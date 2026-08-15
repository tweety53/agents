## Why

The dashboard's "Every open change" table lists one change twice. Observed on 2026-08-15, project
`gymie-7c1f238a`:

| Change | State | Updated | By |
|--------|-------|---------|----|
| `kan-175-more-ui-ux-fixes` | `STARTED` | 18:06:23 | `myflow stage begin (synthetic)` |
| `kan-175` | `STARTED` | 18:05:54 | `myflow stage begin (synthetic)` |

Only `kan-175-more-ui-ux-fixes` is a real change. `kan-175` is a row nobody planned, nobody will
implement, and nothing will archive — and it carries its own `/myflow-do` next command, so it reads
as work waiting to be done.

**The cause is one placeholder in one skill.** `/myflow-fast`'s state gate marks `do.state-gate`
with `<name-or-best-guess>` (`skills/myflow-fast/SKILL.md`), and that mark fires *before*
section A resolves the Jira key into the `<key>-<slug>` change name. The daemon's begin handler
bootstraps a synthetic change row for any name the store has never heard of
(`ApplyBeginStageMark`, `stats/internal/api/stages.go`), so the guess gets a row of its own, and the
resolved name gets a second one 29 seconds later.
<!-- measured: the two rows' updatedAt values, read from the state-board API on 2026-08-15 -->

**The bootstrap is not the defect.** It is a recorded decision (kan-174, `stats/internal/stages/synthetic.go`):
a mark is never dropped, and a synthetic row is "a defect worth seeing in the data" rather than a
special case a reader has to know to look for. It is working exactly as designed here — it is
showing a defect, and the defect is upstream of it. Suppressing synthetic rows in the dashboard
would hide the one signal that made this visible.

`/myflow-do` and `/myflow-start` are unaffected: every mark they emit already carries a resolved
name. `/myflow-fast` is the only command whose mark can name something that is not a change.

## What Changes

- **The `do.state-gate` mark waits for the resolved name.** In `/myflow-fast`, the read-only
  `myflow state get <name-or-best-guess>` stays exactly as it is — a read creates nothing. Only the
  marks change: on a creating run their `begin`/`end` pair fires back to back the moment section A
  fixes the change name, immediately before `start.resolve-change`. At `IN_PROGRESS` the name is
  already resolved before the gate, so those marks stay where they are today.
- **A guard makes the rule mechanical.** `scripts/check-stage-mark-calls.sh` gains a fourth check
  beside its three existing ones: a `stage begin`'s positional change argument may not be a
  placeholder that names a guess. Without it, this change is prose that nothing verifies and the
  placeholder returns the next time someone moves a mark upward for tidiness.
- **The rule is stated once**, under **Stage marks** (`skills/myflow-contracts/pipeline.md`), beside
  the session-token and harness rules the same guard already enforces.
- **The stray row is deleted** from the shared development database.

## Capabilities

### Modified Capabilities

- `myflow-run-telemetry`: a mark's change argument names a resolved change, never a guess — the
  sibling of kan-174's "a state gate reads the state before it marks".
- `myflow-fast-command`: where `/myflow-fast`'s state-gate mark fires on a creating run.

## Impact

**Skills and contracts** — `skills/myflow-fast/SKILL.md` (State gate), and one paragraph under
**Stage marks** in `skills/myflow-contracts/pipeline.md`. Both files have budget headroom against
`scripts/check-contract-budget.sh`: pipeline.md 41042 of 44574 bytes, myflow-fast/SKILL.md 16632 of
18225 (measured 2026-08-15).

**Guards** — `scripts/check-stage-mark-calls.sh` and its harness
`scripts/test-check-stage-mark-calls.sh`.

**Data** — one stray change row in the shared development database (`myflow` on
`localhost:5433`), removed by a one-off statement rather than by new tooling.

**Not changing** — the synthetic bootstrap in `stats/internal/api/stages.go`, the
`SyntheticChangeUpdatedBy` sentinel, `state get`'s `"synthetic": true` reporting, the dashboard's
rendering of open changes, or any mark emitted by `/myflow-do`, `/myflow-start` or `/myflow-finish`.
No Go or TypeScript source changes.
