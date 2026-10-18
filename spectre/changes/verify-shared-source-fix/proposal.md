# verify-shared-source-fix

## Why

Operator, 2026-10-18, after gymie kan-941's `/flow` run. Three causes:

- The parent stopped and asked whether to fix a visual-verify departure because its root cause was a shared, app-wide style: the fix would move every baseline. The operator wants such a fix made automatically; asking stays for a prompt with no clear option or an irreversible or outward-facing action.
- The verifier left checks "not covered" and the run handed them to the operator: a shadow `measure-visual-properties.sh` could not measure because the element reaches the frame's edge, an incomplete element × property matrix, and skipped sweeps.
- Playwright's bare `--update-snapshots` is its `changed` preset, which keeps a PNG whose diff is under the threshold, so a subtle colour fix's recapture came back unchanged.

Operator, same day: a hand-written verifier dispatch ran on `flow-high` although `skills/flow/visual-verify.md` fixes it at `flow-low`/`opus`, and a parent ran scoped in-run fixes on `flow-high` where the recorded decision said inline. Prose did not stop either; a hook denial does.

## What changes

- `skills/flow-contracts/operator-prompts-auto-resolution.md`: a mockup departure whose root cause is a shared style, token or component is auto-resolved like any other.
- `skills/flow/verify-fix-loop.md`: such a defect is fixed at the shared source; the round's task recaptures every baseline the fix moves (`--update-snapshots=all`) and every re-run names the moved frames under `moved:` so they are compared against their mockups.
- `skills/flow/visual-verify-verifier.md`: step 8's Playwright `capture` appends `--update-snapshots=all` (measured against `@playwright/test` 1.62.1: bare keeps a sub-threshold PNG, `=all` rewrites it and writes a missing one with exit 0, and the last flag wins); the frame list takes the `moved:` frames; a matrix cell the pixels cannot answer — an exit 1 no widening resolved, a `clipped` side — is measured from the mockup's source (its HTML's CSS), n/a only with no source in the tree; step 10 is `done` only with a full matrix and every sweep `done` or `n/a — <why it does not apply>`.
- `skills/flow/visual-verify.md`: a gap the parent's reconciliation finds (a frame, bands/seams/matrix line, matrix row, cell or sweep missing) makes the report incomplete and dispatches its remainder, instead of blocking as a departure routed to the fix loop.
- `skills/flow/sync-onto-base.md`: the rebase's targeted recapture appends `--update-snapshots=all`.
- `measure-visual-properties` (Go, `stats/internal/guard/measurevisualproperties.go`): a given region whose side lies on the image's edge with the control running past it reports that side `clipped` and measures the others, instead of exiting 1.
- `hooks/enforce-agent-baseline.py` (harness `scripts/test-enforce-agent-baseline-hook.sh`):
  - denies a visual-verify verifier or tooling-analyst dispatch off the agent type and model its contract fixes (`flow-low`/`opus` and `flow-medium`/`opus`; `flow-high`/`glm-5.3-flash` on zcode). The role is recognised by the line its prompt opens with — `Role: visual-verify verifier`, `Role: visual-verify tooling analyst` — which `skills/flow/visual-verify.md` and `skills/flow/visual-verify-tooling-analysis.md` now require; a prompt merely citing either file is not that role.
  - denies any `flow-high` dispatch but the no-progress escalation — its prompt opening `No-progress escalation`, or a self-review or pipeline-fix loop key — and a pass-1 panel dispatch the operator approved at `high`, its prompt opening `Operator-approved high`.
- `skills/flow/visual-verify-tooling-analysis.md`: the tooling analyst moves from `flow-high` to `flow-medium`.
- `skills/flow/implement.md`: `flow-high` is dispatched in exactly those two cases, never chosen by hand. `skills/flow-contracts/pipeline.md`: the escalation prompt opens with its marker line. `skills/flow-self-review/SKILL.md`: the planner's `flow-high` re-plan is the no-progress escalation only.
- `skills/flow/brainstorm-planner.md`: `high` reaches a pass-1 panel dispatch only through the operator's Yes at the plan gate — the fix-run re-Decide route is removed — and that dispatch's prompt opens with `Operator-approved high`.
- `measure-visual-properties`: when a control covers two of the crop's corners (a bar on the image's top edge, a full-height side panel), each other corner colour is tried as the background.
