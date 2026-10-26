# self-review-kan-1018-week-programs-plan-a-week

## Why

The self-review pass of `/flow` change `kan-1018-week-programs-plan-a-week` found four non-`big`
defects (evidence: `.superpowers/sdd/self-review-pass.md` in the gymie worktree
`/Users/tweety53/Projects/gymie-worktrees/kan-1018-week-programs-plan-a-week`):

- P1: `hooks/protect-main-checkout.py` lets a command that commits with a message it cannot read
  (`-F`, `-C`, `--amend`, a `$(…)` message) and pushes to `spectre/*` in the same call through —
  the push check runs before the commit exists. A task commit with no `Task-Id:` trailer reached
  origin that way (gymie `863f872a`).
- P2: `flow record dispatch backfill` cannot repair an inline row begun after its commit: the
  re-begin keeps the late `started_at`, so the end refuses again.
- P3: no gate checks that a panel fix round recorded its `fix-mutation:` proof; none of
  kan-1018's three fix rounds had one and the change landed.
- P4: kan-1018's archived records misstate the live check's status (200 for a 201) and round 0's
  counts (seven Importants and five Minors for six and six).

## What changes

- The hook refuses a commit-then-push to `spectre/*` whose commit message it cannot read, telling
  the caller to commit and push in two calls.
- A re-begin carrying an earlier instant moves an open (never-ended) dispatch row's `started_at`
  back, so backfill's derived instant replaces a late begin.
- `check-unfinished-work` refuses landing when a panel round has a `fixed` finding and no
  `fix-mutation` row for that round.
- kan-1018's archived `live-verification.md` and `narrative.md` state the recorded figures.
