# Known bugs

- `scripts/test-check-cleanup-complete.sh` — the fixture survivors commands' timing bounds (the
  5-second inside-bound cases and the 10-second escaper-race case) lose their races when the
  machine is under heavy external load — observed at load average ~24 on 10 cores, where
  `./survivors.sh` spawns slower than the bound and the group kill can beat a fork — failing 3-4
  timing cases that pass on an idle machine; the guard under test and every assertion are
  correct, the bounds are simply not load-proof (KAN-376 raised one bound for the same reason) —
  introduced by 18feb597 (refactor(flow): work only in worktrees, back the branch up remotely,
  drop the myflow legacy), an ancestor of main.

- `tests/visual/runs.spec.ts › runs › renders runs grouped by change, expanded to main session and
  dispatches` — `runs-darwin.png` was captured before the expanded run grew a Stage/Wall clock
  sub-table, so the page now renders 1195px tall against a 1072px baseline — introduced by
  dd995076 (feat(stats): show per-stage wall clock in the runs view detail), an ancestor of main.
