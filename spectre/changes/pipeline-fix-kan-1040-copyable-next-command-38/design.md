# Design — pipeline-fix-kan-1040-copyable-next-command-38

## Context

See `proposal.md`: `land`'s `NEXT:` line is the parent's only cue to print the Finished
handoff, and it names neither the template nor its never-omitted lines.

## Decisions

**ID:** next-line-names-template
**Chosen:** extend `land`'s `NEXT:` line with the template path and the two never-omitted lines.
**Rejected:** `land` resolving and printing the `Learning:` and `Prompt tuning:` values itself —
it needs the learning-note location and the agents repo's main checkout, both outside `land`'s
inputs; more code for the same failure the named template already closes. Add it if the line is
dropped again with the template named.

**ID:** stop-hook-blocks-incomplete-handoff
**Chosen:** `hooks/flow-stop-guard.py`, already the Stop hook for every session, reads the turn's
last assistant text from the transcript. A text carrying a `## Finished` line, or either of
`**Learning:**` / `**Prompt tuning:**` (`/flow-fast`'s landing report has no heading, so carrying
one of the two lines is its only recognisable shape), that lacks one of the two lines blocks with
a reason naming the missing line(s) and `skills/flow/cleanup.md`. The check needs no store, so it
runs first; it is skipped when `stop_hook_active` is true, so a turn is blocked at most once for
it and never trapped; the hook's fail-open wrapper still covers it.
**Rejected:** treating the printed handoff as unguardable chat output — a Stop hook reads the
transcript's last assistant message, and the operator asked to prevent the drop, not only cue it.
**Rejected:** a `/flow-fast` report that drops both lines is not caught — nothing else marks it as
the landing report; add a heading to `skills/flow-fast/SKILL.md` if that drop is seen.

**ID:** land-prints-no-values
**Chosen:** `land` still does not print the `Learning:` and `Prompt tuning:` values. Reason: the
Stop hook now refuses the turn end until both lines are present, and resolving the values needs
the learning-note location and the agents repo's main checkout, both outside `land`'s inputs.

**ID:** flow-fast-unchanged
**Chosen:** no `/flow-fast` change. Its report template sits in the skill it runs from, which
already says both lines are carried; there is no `NEXT:` hand-off to amend.

**ID:** first-review-effort
**Chosen:** flow-low — a one-string change and a small hook branch, each with checks verified
applied and failing without the fix.

**ID:** fixer-groups
**Chosen:** one group, tasks 1 and 2 — disjoint files, both small.
