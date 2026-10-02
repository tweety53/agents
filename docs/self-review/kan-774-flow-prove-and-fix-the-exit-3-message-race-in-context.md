# Self-review context bundle for kan-774-flow-prove-and-fix-the-exit-3-message-race-in

found: 1 of 7 sources; skipped: 6 of 7 sources
note: RECORDS LOSS — a flow.review-panel stage run completed for kan-774-flow-prove-and-fix-the-exit-3-message-race-in, but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-774-flow-prove-and-fix-the-exit-3-message-race-in.md (absent)
skipped: .superpowers/sdd/reviews/kan-774-flow-prove-and-fix-the-exit-3-message-race-in-panel.md (absent)
skipped: spectre/changes/archive/kan-774-flow-prove-and-fix-the-exit-3-message-race-in/tasks.md (absent)
skipped: spectre/changes/archive/kan-774-flow-prove-and-fix-the-exit-3-message-race-in/design.md (absent)
skipped: spectre/changes/archive/kan-774-flow-prove-and-fix-the-exit-3-message-race-in/narrative.md (absent)

## git log --stat

commit 8f4d1c92b616ddcada364896dcadfdd19477514c
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Oct 3 00:44:35 2026 +0300

    docs(known-bugs): record the run-reproducer exit-3 survivor-message race, fixed at source

 KNOWN-BUGS.md | 17 +++++++++++++++++
 1 file changed, 17 insertions(+)

## Session narrative

KAN-774 asked this run to prove and fix the exit-3 message race that `scripts/test-run-reproducer.sh` cases 13/14/18 showed once during kan-676's verify, and to record the flake. Reading the tree showed the fix already on main: the Go port 117b6ae1 replaced the bash sweep, whose survivor-naming read sat AFTER each pid's SIGKILL (introduced by 0cf0b173) and lost to launchd's reaper under load, with `rrSweep` reading liveness BEFORE the kill and `TestRunReproducerSurvivorNamedWhenReapedAtKill` pinning the losing interleaving — one day after the flake was observed. The prove half therefore became a mutation proof of the existing pin: `break-and-prove.sh` reordered the read after the kill and the pin test failed with `survivors [], want [64801]`, exactly the observed shape, then passed restored; the sweep-bearing cases 10/13/14/16/18 ran green on the untouched tree. The record half went into KNOWN-BUGS.md as historical, not live — the entry names the deleted bash harness, the mechanism, the introducing commit, and the fixing commit with its pin, so no future run re-diagnoses the flake or mistakes it for a live bug. Judgment calls: the fix predating the change reduced this run's fix half to proving the fix is real; `flow tasks count` was skipped because it reads a spectre path a flow-fast run never writes; the bundle's RECORDS LOSS note is expected here — the micro decision dispatches nothing, so the store legitimately holds no dispatch rows.
