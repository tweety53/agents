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

- `scripts/test-run-reproducer.sh` cases 13/14/18 — the kill sweep's survivor-naming read lost to
  launchd's reaper: it read `kill -0` on each pid AFTER its `kill -KILL`, counting on the reaper
  being slower than the read, and under suite load the reaper won — the SIGKILLed orphan was
  reaped before the read, the survivor went unnamed, and the exit-3 message lost its `surviving
  process pid(s)` clause (cases 13/14 fell through the survivor branch to exit 1 "defect not
  demonstrated"; case 18 kept exit 3 but printed the no-survivor timeout message) — observed once
  during kan-676's verify, classified pre-existing, and left unrecorded until KAN-774 — introduced
  by 0cf0b173 (feat(kan-108-cut-the-time-and-token-cost-of-a-myflow-do-run): verify a fix
  instruction's reproducer before dispatch, and narrow the vacuous escalation trigger), an
  ancestor of main. Historical, not live: fixed at source by 117b6ae1 (feat(stats): run the five
  slowest guards as one Go flow-guard binary built from the checkout), whose Go sweep reads each
  pid's liveness BEFORE its SIGKILL — nothing reaps a live process, so the read cannot lose — and
  whose `TestRunReproducerSurvivorNamedWhenReapedAtKill` pins the losing interleaving; the same
  commit deleted the bash harness this entry names, so no current test can carry the failure.
<!-- measured: scripts/break-and-prove.sh stats/internal/guard/runreproducer.go --patch <the reorder to the post-kill read> -- go -C stats test ./internal/guard/ -run TestRunReproducerSurvivorNamedWhenReapedAtKill -count=1 @ branch kan-774-flow-prove-and-fix-the-exit-3-message-race-in — survivors [], want [64801] with the read after the kill; PASS restored -->


- `skills/flow/review-panel.md` MUTATION PROOF paragraph — the cited `break-and-prove.sh <file>
  (--sed <expr> | --patch <patch>) -- <test-command>` signature omits `[--clean <command>]`, the
  forced clean re-run the script's header says exists because Gradle skips a re-run and records a
  stale green; a fixer on a Gradle project following the citation alone falls into that trap —
  Minor, deferred from the kan-916 self-review fix review — introduced by aee79446 (docs(flow):
  name break-and-prove.sh in the MUTATION PROOF paragraph).
