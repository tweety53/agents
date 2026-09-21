# Self-review context bundle for kan-703-flow-cost-pass-8-was-silently-demoted-to-a

found: 1 of 7 sources; skipped: 6 of 7 sources
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-703-flow-cost-pass-8-was-silently-demoted-to-a.md (absent)
skipped: .superpowers/sdd/reviews/kan-703-flow-cost-pass-8-was-silently-demoted-to-a-panel.md (absent)
skipped: spectre/changes/archive/kan-703-flow-cost-pass-8-was-silently-demoted-to-a/tasks.md (absent)
skipped: spectre/changes/archive/kan-703-flow-cost-pass-8-was-silently-demoted-to-a/design.md (absent)
skipped: spectre/changes/archive/kan-703-flow-cost-pass-8-was-silently-demoted-to-a/narrative.md (absent)

## git log --stat

commit c4f3951324481dae0858c422a9e1d3066a3293e0
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 22 01:45:45 2026 +0300

    flow: cap panel whole-branch re-reads and surface demotions in the clean claim
    
    Rerun policy full's final whole-branch pass repeated after every clean fix
    round with no bound: kan-573's panel ran eight of them (fifteen rounds,
    convergence declared on the last), kan-580 five plus an operator-approved
    sixth. The mechanism now adds two unasked — the first final pass and its
    one repeat — and a third runs only on one explicit operator choice put at
    the cap, whose silent default closes the panel and marks the Panel: line;
    no fourth is dispatched whatever the third finds. Targeted delta re-runs
    stay uncapped. This settles in the shape KAN-662 names.
    
    When a budget rule — the sonnet/haiku restriction from the third
    whole-roster read onward — substituted a weaker pair on the pass whose
    clean result closes the panel, the Panel: line now says so in a demoted:
    field naming the pass, the pair it ran, and the pair the normal resolution
    would have given: clean on pass 8 and clean on opus are not the same
    evidence. KAN-703.

 skills/flow-contracts/handoff-blocks.md |  2 +-
 skills/flow/review-panel.md             | 30 +++++++++++++++++++++++++++++-
 skills/flow/verify-and-handoff.md       | 11 ++++++++++-
 3 files changed, 40 insertions(+), 3 deletions(-)

## Session narrative

A `/flow-fast` run (inline by the command's own terms — the project's three `dynamic` toggles were
read and deliberately not engaged, because the invoked command states "implement inline in this
session… no subagent, no review panel, no decision record"). The session resolved KAN-703,
settled the cap design against the recorded history — KAN-662's named shape, the reverted
`cc9167b` prototype (per-repeat prompt, default "run again"), and the operator's later
`30f7b9f` cheapen-the-passes rule — then edited three contract files: `review-panel.md` (the
two-unasked-passes cap with one explicit operator choice at it, and the demotion-is-evidence
obligation beside the restriction rule), and the `Panel:` line in `verify-and-handoff.md` and the
canonical `handoff-blocks.md` template (new `demoted:` and `rerun cap:` fields). The whole
`## lint` list ran green in the worktree, including the fresh-worktree `make web-build` the
`## worktree setup` section demands before `go vet`; the scoped test suite
(`run-guard-tests.sh`) passed 78/78. Where it struggled: the rejected prototype left no recorded
rejection reason, so the ask's polarity (default closes the panel, ⚠ on silence) was reasoned
from the operator-prompts contract and KAN-662's wording rather than from a recorded position;
and `check-normative-inventory.sh` printed 7 lines where `.flow/project.md` prose expects
"roughly a thousand" — reproduced identically on the main checkout, so pre-existing and left
alone rather than investigated as a regression.
