# self-review-backlog-flow-7

## Why

The next five live tickets in KAN's "Backlog Flow" status, oldest first, filed by earlier
self-review passes, are planned here as one self-review fix change (`design.md`, **Context**):

- **KAN-771** — a served self-review bundle names no daemon build, so a source the stale
  protected `flowd` cannot serve reads `skipped: <source> (absent)`, the same line a real absence
  prints.
- **KAN-790** — a mockup departure measured and fixed in visual verification leaves no test
  behind; kan-577's growing pixel probe was the exception, not a rule.
- **KAN-801** — nothing decides a departure whose frames disagree on a shared component, so each
  one becomes a per-item question; kan-741's "the component's own source wins" decided a whole
  backlog.
- **KAN-804** — kan-741's admin-frontend serving worktree carried no commits and never reached
  the state record, so run 2's per-worktree checks never ran on it.
- **KAN-815** — run 1 archives only in the canonical repository, so a satellite's
  `spectre/changes/<name>/link.md` stays unarchived and run 2 reports `LEFTOVER`.

## What changes

- `flowd` stamps its VCS build into every self-review bundle it serves, as the bundle's first
  `note:` line.
- **The loop** (`skills/flow/verify-fix-loop.md`) gives the in-run fix implementer the PIXEL
  PROBE paragraph, requires a permanent probe for every mockup departure the round fixes, and
  classifies departures before the round's task is written: agreeing frames are fixed, a frame
  that disagrees with a shared component's own design source closes as a mockup conflict.
- **2. Isolate the workspace** (`skills/flow/implement.md`) names a serving stack among the
  needs that run `kickoff-worktree.sh --from`, and states a commit-less worktree is recorded like
  any other.
- `integrate-change.sh commit` archives and commits every satellite's change directory after the
  canonical one; **Archive on the change branch** (`skills/flow-contracts/finish-contract-run1.md`)
  says so.
