# The panel roster follows the settings list

## Why

`flow settings set -reviewers …` writes a list that **nothing reads**. `skills/flow/SKILL.md`'s
Model resolution takes `.defaultModel` from the settings store and stops there; the panel's roster
is instead written in prose in `skills/flow/review-panel.md` as three required slots plus two
on-demand ones.

So `/flow-settings` maintains a field with no effect, and changing the roster means editing a
contract file. Adding Bugbot through `/flow-settings` appears to work and changes nothing about any
run.

## What changes

- `skills/flow/SKILL.md` resolves the roster from `.reviewers` in the same step that already
  resolves `.defaultModel`.
- `skills/flow/review-panel.md` dispatches the resolved list instead of enumerating a fixed
  three-plus-two, keeping each slot's own spawn shape.
- `skills/flow-settings/SKILL.md` drops its "three dispatched by default, two on-demand-only"
  framing, which becomes false.
- A new decision supersedes `review-panel-fixed-3` by name.

Design, decisions and rejected alternatives: `design.md` beside this file.
