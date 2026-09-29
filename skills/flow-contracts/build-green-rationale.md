# Build green — rationale

This file is the reasoning behind **Build green** (`skills/flow-contracts/build-green.md`).
**A `/flow*` run never loads it — appendices are for whoever edits a contract.**

## build-green.md — The canonical statement

`the guard script's own module docstring points here rather than restating the rule` — a second copy is the same Single Source of Truth violation `<agents repo>/stats/internal/guard/planprovenance.go`'s docstring warns against.

## build-green.md — The build-green tag

This is the same placement rule the guard script's own docstring states as a regex — this file states it in prose, and the two are required to describe the same rule, the column-0 anchor included.

## build-green.md — The guard's scope

`where its report is blocking rather than advisory` — (KAN-538).

## Moved by KAN-856

Verbatim passages KAN-856 moved out of the planning session's run-loaded files; each is the reason behind a rule that stays where it was.

### build-green.md — The build-green tag (KAN-856)

Indenting the fields along with the steps is the
natural reading of "the body sits beneath its task", and it is wrong in a way only
`check-plan-shape.sh` catches kindly: `spectre validate` reports no findings, because an indented `**Build:**` line is no more a
task line to spectre than a step is, while this file's own guard reports
`task <id> has no **Build:** tag` — naming the consequence and hiding the cause, since the tag is
there, one column short of where its regex looks. Measured on a one-task plan written both ways.
The shape guard names the cause instead — `task <id> carries a **Build:** line indented past
column 0` — and runs unconditionally at plan time.

A first line whose value opens with neither keyword
(`**Build:** yellow`, `**Build:** greenish`, a bare backticked path) is a **malformed tag** — its
own violation, reported by that line and its value, never as a missing tag (which would name the
consequence and hide the cause, exactly as the unclosed-fence finding refuses to), and never
overridden by a well-formed `**Build:**` line further down.

### build-green.md — What the guard does not do (KAN-856)

This is the
same accepted limit as **What the guard does not do**
(`skills/flow-contracts/plan-provenance-guard.md`): a script can confirm a claim was written down, not
that the claim is correct.
