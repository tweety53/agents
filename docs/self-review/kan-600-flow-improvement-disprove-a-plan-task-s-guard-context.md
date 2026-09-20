# Self-review context bundle for kan-600-flow-improvement-disprove-a-plan-task-s-guard

found: 1 of 7 sources; skipped: 6 of 7 sources
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-600-flow-improvement-disprove-a-plan-task-s-guard.md (absent)
skipped: .superpowers/sdd/reviews/kan-600-flow-improvement-disprove-a-plan-task-s-guard-panel.md (absent)
skipped: spectre/changes/archive/kan-600-flow-improvement-disprove-a-plan-task-s-guard/tasks.md (absent)
skipped: spectre/changes/archive/kan-600-flow-improvement-disprove-a-plan-task-s-guard/design.md (absent)
skipped: spectre/changes/archive/kan-600-flow-improvement-disprove-a-plan-task-s-guard/narrative.md (absent)

## git log --stat

commit cdb06c0e5ddd82edcbe23345a07ae3ed6b58ef6c
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sun Sep 20 22:15:37 2026 +0300

    docs(planner): run a task's guard-premise during planning

 skills/flow/brainstorm-planner.md | 11 +++++++++++
 1 file changed, 11 insertions(+)

## Session narrative

Implemented kan-600 in one micro-classed run: a single paragraph in brainstorm-planner.md's
writing-plans section makes a task premise about a guard's current behaviour a plan-time
measurement — run the guard, cite its output with a `measured:` comment, strike a disproved
task before the panel reads the plan. The run ate its own dog food: the plan's premises (budget
ratchet green at the merge base, no existing rule) were measured and cited in tasks.md before
implementation. The struggle was the landing: verify surfaced a pre-existing
check-installed-citations.sh red in finish-contract-run2.md — two rootless citations on
origin/main, already fixed by the operator's staged-but-unlanded edits to the same file — and
the run held there rather than push a red gate or collide with that in-flight work, until the
operator said to finish.
