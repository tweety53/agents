# kan-394-flow-check-task-build-green-sh-misses-build — design

## Context

Approved design: `docs/superpowers/specs/2026-09-04-kan-394-flow-check-task-build-green-sh-misses-build-design.md`
— canonical for the grammar, the guard's new violation, the contract edits, the test cases and
the measured context. This file carries the decision and open-question records the pipeline reads.

KAN-394, from KAN-29's self-review: `check-task-build-green.sh`'s `green|red` regex misses a
`**Build:**` line carrying prose after the keyword, and reports the miss as a missing tag. The
regex is one definition, `BUILD_TAG_RE` in `scripts/lib/plan_grammar.py`, read by both build-green
guards through `select_build_tag`.

## Decisions

### An unparseable `**Build:**` line is a malformed-tag violation, not a missing tag

**ID:** malformed-build-tag-is-its-own-finding
**Status:** active
**Chosen:** distinct violation naming the line and its value; the first `**Build:**` line in a
body decides — the guard names the cause, as it already does for an unclosed fence.
**Considered:** regex-only change, keeping `task <id> has no **Build:** tag` for a malformed line —
smaller diff, but leaves the guard naming the consequence and hiding the cause, the defect this
change exists to remove; ruled out by the operator.

### The keyword is a word-boundary prefix

**ID:** build-keyword-prefix-word-boundary
**Status:** active
**Chosen:** `^\s+(green|red)\b` on the value after `**Build:**`; trailing text ignored.
**Considered:** requiring whitespace or end-of-line after the keyword — rejects `green.` and
`green,` for no gain; `\b` is the stdlib answer and rejects `greenish` just as well.

## Open questions

None.
