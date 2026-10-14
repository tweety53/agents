# self-review-kan-937-step-2-frontend-logic-cardio-api-models-step

## Why

The self-review pass over kan-937 found five defects, none `big` — four `flow-fix`, one `flow-cost`:

- A `/flow` fix run at `IN_PROGRESS` ran an appended task inline with no dispatch row and an unrecorded
  review, and nothing caught it until the next integrate's unfinished-work gate.
- The unfinished-work guard's signal four clears an appended task that carries an implementer row but no
  reviewer row, though its own message asks for the fix round's review.
- `skills/flow/cross-repo-worktrees.md` names link.md's `## Merge order` but not its entry syntax, and
  `spectre validate` refuses the bare `1. .` a reader writes from that prose.
- `flow self-review bundle` logs one repository's commits, so a cross-repo change's satellite
  implementation commits never reach the self-review pass.
- Inline dispatch rows are begun and ended in one Bash call after the work, a zero-width window whose cost
  the harvester cannot attribute; only inline fix rows are told not to.

## What changes

- A fix run runs `check-unfinished-work.sh` on itself before writing its `IN_PROGRESS` handoff.
- Signal four also requires a `reviewer` row begun after the appended task's implementer row.
- `cross-repo-worktrees.md` carries the literal `## Merge order` block `spectre validate` accepts.
- The self-review bundle appends each merge-order peer's `spectre/<name>` commits to its git-log source;
  `finish-contract-run1.md` says so.
- The begin-before-the-work rule covers every inline row, and `flow record dispatch end` refuses an inline
  row closed under ten seconds after its begin with no token report.
