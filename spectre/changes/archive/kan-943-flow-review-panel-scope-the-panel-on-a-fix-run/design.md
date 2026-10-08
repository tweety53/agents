## Context

The late-fix reduction (`skills/flow/review-panel-late-fix.md`) already narrows a fix run's
read scope to the range since the panel's last clean close, but only for a small delta that adds
no task, and only by also cutting the roster to `primary`. Any other fix run — every append, or a
non-append delta over 40 lines — takes the full path and re-reads the whole branch.
<!-- measured: grep -n "lftMaxLines = 40" stats/internal/guard/latefixtrigger.go @ 3d108317 -->

## Decisions

### Read scope of a fix run that fails only on size or scope growth

**ID:** append-scope-since-close-full-roster
**Status:** active
**Chosen:** the decided roster unchanged, every dispatch reading `late-fix.diff` (the since-close range) — the ticket's own direction, "the append's own diff plus the code it touches": every reviewer reads the code a diff touches already, the diff being the entry artifact, not the boundary.
**Considered:** `primary` alone on the delta (drop condition 4 from the late-fix reduction) — too narrow a roster for a real appended task of any size; status quo — the KAN-758 waste stands.

### Which fix runs qualify

**ID:** append-scope-any-fix-run
**Status:** active
**Chosen:** any fix run whose late-fix check fails on conditions 3 and/or 4 alone — a large non-append fix delta wastes the same whole-branch read.
**Considered:** appends only (condition 4 failed) — a second rule for the same waste, with a non-append fix over 40 lines still re-reading the branch.
<!-- measured: grep -n "lftMaxLines = 40" stats/internal/guard/latefixtrigger.go @ 3d108317 -->

### Where the verdict comes from

**ID:** append-scope-exit-3
**Status:** active
**Chosen:** a new exit 3 of `check-late-fix-trigger.sh`, which already evaluates conditions 1–5 in one pass.
**Considered:** a second guard — would re-implement conditions 1, 2 and 5 verbatim.

### The diff file

**ID:** append-scope-reuse-late-fix-diff
**Status:** active
**Chosen:** reuse `late-fix.diff`, written by `write-panel-diff.sh late-fix` and rendered with `-diff late-fix` — same range, same writer, no new code.
**Considered:** a new `append-scope.diff` name — a second writer mode and render kind for an identical range.

### Critical or Important findings under the append scope

**ID:** append-scope-no-void
**Status:** active
**Chosen:** no voiding — findings feed the ordinary fix-round loop; the full roster already read the delta, so a severe finding says nothing about the read being too narrow.
**Considered:** voiding as the late-fix reduction does — it would re-read the whole branch on the next round, which the read scope rule exists to avoid; the late-fix void rests on its roster being one slot, which does not hold here.

## Open questions

## Live check

None — the change edits a guard's verdict and the panel contract; no service or stored state.
