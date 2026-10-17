# kan-991-panel-effort-roll-high-gated

## Context

The effort policy lives in prose in `skills/flow/brainstorm-planner.md`'s **Model and effort**;
`plan-class.sh` already prints `effort_roll` (sha256 of `<name>effort` mod 100) and nothing in
`flow record decision` checks a panel dispatch's effort. The gated per-task reviewer's effort is
restated in `skills/flow/implement.md`. Implementer, fixer, group, rerun-pair and extra-dispatch
effort rules are out of scope and unchanged.

## Decisions

### Panel effort comes from the existing effort roll

**ID:** panel-effort-roll-60-40
**Status:** active
**Chosen:** a pass-1 panel dispatch and a gated per-task reviewer bundle take `low` when
`effort_roll < 60`, `medium` otherwise; the `≥ 90` planner-choice band no longer applies to them —
reuses the roll `plan-class.sh` already prints, so no Go or script change.
**Considered:** a new independent roll (`<name>paneleffort`) printed by `plan-class.sh` — a Go
change plus a ninth output line for no behaviour the shared roll lacks; keeping the planner's free
choice above 90 — leaves the ungated `high` path the ticket removes.

### High is operator-gated

**ID:** panel-high-operator-gated
**Status:** active
**Chosen:** the decision records the roll's effort; where the planner judges a dispatch's hard
seam warrants `high`, the plan gate's AskUserQuestion carries one extra question per such
dispatch — **Raise `<dispatch>` to `high` — `<seam>`?**, **No** *(default, recommended)* / **Yes** —
and only an explicit Yes re-records that pair at `high`. A fix run's re-Decide may set `high` on a
panel dispatch when the fix instructions report that an earlier, lower-effort attempt failed to
solve the same issue, the pair's `reason` naming that attempt.
**Considered:** a third option on the gate question itself — the gate has exactly two options by
contract; enforcing in `flow record decision` — a Go change the ticket did not ask for, prose
suffices for a planner-written record.

### Display-only formatting is decided, not asked

**ID:** display-formatting-auto-decided
**Status:** active
**Chosen:** a bullet in brainstorming's checklist: display-only formatting the request leaves open —
numeric precision, empty-cell placeholders and the like — is chosen by the planner and recorded
under `## Decisions`, never asked; a mockup or ticket that states it still governs.
**Considered:** amending `rules/design-mockups-are-specs.mdc` — that rule governs drawn states,
not formatting a frame leaves unstated, and is global to every project.

## Open questions

## Live check

None — the change edits skill prose only; no running service or persistent state is touched.
