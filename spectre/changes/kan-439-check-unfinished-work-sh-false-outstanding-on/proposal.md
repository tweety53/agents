# kan-439-check-unfinished-work-sh-false-outstanding-on

## Why

KAN-439 (self-review finding from KAN-423): on a cross-repo change whose project keeps the spectre
change directory in the canonical repository only, every other worktree has no change directory at
all — no `tasks.md`, no `link.md` — and `check-unfinished-work.sh` reports a false
`OUTSTANDING: no plan at …` at integrate even though run 1 passes it the canonical worktree, where
the plan exists under the same change name. Every such change forces hand verification of the task
checkboxes and the findings store instead of a verdict the operator can trust.

The link-refusal half of the issue (non-gating, never retried) was already fixed by bf684f7; the
treeless-satellite shape is what remains and is unreachable by any `link.md` mechanism, because a
repository with no spec tree has nothing for the link to be written into.

## What changes

- `scripts/lib/change-plan.sh` resolves a change with no local plan and no usable `## Part of`
  link through the supplied canonical worktree, by the same change name.
- `skills/flow/scripts/check-unfinished-work.sh` reads the findings store via the canonical
  worktree when the plan resolved there, so the satellite call's signal two sees the change's real
  findings instead of a permissive empty answer under the satellite's own project key; its
  exit-2 triage and `OUTSTANDING` shapes are unchanged.
- Both harnesses gain the cases that pin the new resolution and the refusal that must NOT be
  rescued; no flow-contract text and no project-configuration key change.
