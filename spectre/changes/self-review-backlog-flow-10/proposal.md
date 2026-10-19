# self-review-backlog-flow-10

## Why

The next five live tickets in KAN's "Backlog Flow" status, oldest first, filed by earlier
self-review passes and not already fixed on `origin/main`, are planned here as one self-review fix
change (`design.md`, **Context**):

- **KAN-899** — `/flow`'s settings read discards `flow settings get`'s stderr, so a store outage
  resolves silently, and the empty roster it leaves reads as a reachable empty list rather than the
  unreachable fallback.
- **KAN-900** — integrate's clean sync rebase runs only the scoped guard-harness re-verification,
  so a rebase that is textually clean but breaks the build — kan-820's dropped `sync` import —
  lands with no check run on it.
- **KAN-902** — a guard that judges its own checkout, called by absolute path from another
  worktree, judges the main checkout instead and reports a false clean; `check-verbatim-moves` and
  `check-references` both reproduce it.
- **KAN-905** — eleven parity-test sites read a pinned historical commit with `git show`, so a
  shallow clone fails fourteen tests of the guard package.
- **KAN-907** — the ignored-signal tests of `prove-reproducer` and `mutate-and-verify` sleep half a
  second and then look, a timed check standing in for a deterministic one.

## What changes

- **SKILL.md**'s settings read leaves the CLI's stderr visible, and the run reports it and takes
  the Unreachable roster.
- `integrate-change.sh` runs the rebased worktree's `## lint` list after every clean sync rebase
  and stops `STOP: rebase-lint` on a failure; `sync-onto-base.md`, `integrate.md` and
  `finish-contract-run1.md` state it.
- `scripts/lib/flow-guard.sh` gains `flow_guard_cwd_checkout`, which the nine guards that judge
  their own checkout's text call first: from another worktree of the same repository, that
  worktree's own copy runs.
- A `historicalFile` test helper skips a parity test whose pinned commit the clone lacks, naming
  the commit.
- The two timed looks are removed; the tests' finished-run assertions, with `prove-reproducer`'s pinned to its exit status 1, are the deterministic proof.
