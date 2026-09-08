# kan-393-flow-give-guards-a-canonical-worktree-concept

KAN-393 · Flow: give guards a canonical-worktree concept for cross-repo changes — close the last
named gap, in `gather-dispatch-context.sh`.

## Why

KAN-29 ran as a three-repo change before any cross-repo mechanism existed, and its self-review
recorded three guard failures. KAN-363 already closed two: `check-unfinished-work.sh` no longer
reports `OUTSTANDING` from a satellite worktree whose plan lives in the canonical repo, and
`check-task-commit-fields.sh` resolves a satellite's task through the link — both take an optional
`[canonical-worktree]` argument and follow `link.md` through `scripts/lib/change-plan.sh`. The
first bullet's other half survived: `gather-dispatch-context.sh` has no link awareness at all.

Under KAN-363's pointer-tree model a satellite change directory carries only `link.md`, so a
bundle gathered for a satellite worktree reports `proposal.md`, `design.md` and `tasks.md` all as
`skipped: … (absent)` and the satellite implementer is dispatched with no plan in the bundle. The
hand workaround — passing the canonical worktree and change dir as the first two arguments —
silently mislabels the bundle: the `project commands` section carries the canonical repository's
lint/test/run commands instead of the satellite repository's own, the `incidents` section reads
the canonical project's incident log, and the header `head:` line names the canonical sha. That is
the "hand-substituted every time" cost KAN-393 records.

The Jira's literal proposal — read the canonical worktree from `.flow/project.md` — predates
KAN-363's shipped design and is superseded by it: the canonical repository is a per-change
property (KAN-343's was `gymie`, KAN-363's was `agents`), which a static per-project key in one
repository's `.flow/project.md` cannot name.

## What changes

`gather-dispatch-context.sh` accepts an optional seventh argument `[canonical-worktree]` — the
same treatment KAN-363 gave the other two guards — and resolves the change's plan through
`scripts/lib/change-plan.sh`. A satellite worktree's bundle carries the canonical
`proposal.md`/`design.md`/`tasks.md`, labeled `(canonical <peer>:<change-id>)`, while keeping its
own worktree's `project commands`, `incidents` and `head:` sha. A plain change resolves to its own
directory and produces byte-identical bundles to today's. A satellite whose canonical plan cannot
be reached skips the three plan sections with a distinct label and still exits 0 — the bundle is
advisory and never gates a run, the deliberate inverse of `check-unfinished-work.sh`'s refusal for
the same condition.

`implement.md`'s per-bundle gather step and `review-panel.md`'s rebuild step pass the canonical
worktree on every call, resolved per the rule `finish-contract-run1.md` already states for
`check-unfinished-work.sh`. `test-gather-dispatch-context.sh` gains the satellite shapes. No
capability spec changes (`spectre/specs/` is empty); `spectre` (the CLI repo) is untouched —
the link mechanism this change consumes already shipped.

Fix 1 (`check-installed-citations.sh` blocked flow.verify): the two `scripts/gather-dispatch-context.sh`
citations tasks 3 and 4 added to `implement.md` and `review-panel.md` name no root; both are
prefixed `<agents repo>/`, the form sibling prose in those files already uses. Prose only — no
behavior change.
