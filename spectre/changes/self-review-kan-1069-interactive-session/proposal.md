# self-review-kan-1069-interactive-session

## Why

The self-review pass of `/flow-interactive` change `kan-1069-interactive-session` found five
non-`big` defects:

1. `flow record request end -outcome landed` checks only the UI check, so eleven of the change's
   twelve requests landed with no `flow.verify` stage run behind "the request never lands red".
2. A UI check whose begin and end were chained in one call satisfies the same guard, against step
   7 of `skills/flow-interactive/SKILL.md`.
3. The run page unmounts — a blank SPA — for every change whose latest decision stores
   `overrides` as an object, because `DecisionsPanel`'s `lines()` calls `.flatMap` on it.
4. On the live state board, a failed finished-changes read is silent once the open poll has
   answered: the Finished tile and "Cost in period" read "—" with no message.
5. Findings the per-task gated reviewer raises are never recorded with `flow record finding`, so
   the run page's Findings KPI cannot count them and carries a "review panel only" hint instead.

## What changes

- The request-end guard also requires, on every `landed` end, the request's latest
  `/flow-interactive` `flow.verify` stage run to have ended `completed`, and refuses a `completed`
  UI check that ended less than one second after it began; the skill's **Request-end guard**
  states both.
- `DecisionsPanel` reads `overrides` only when it is an array.
- The state board shows a read error beside its heading whenever a read fails, the finished read
  included.
- The `/flow` parent records each gated reviewer pass's findings under slot `gated-reviewer` and
  closes them when the fix re-review is clean; the panel's guards, its rendered record and
  `flow record next-round` skip those rows, and the Findings KPI drops its "review panel only" hint.
