# Self-review context bundle for kan-679-flow-improvement-gate-the-draw-at-the-root-never

found: 1 of 7 sources; skipped: 6 of 7 sources
note: RECORDS LOSS — a flow.review-panel stage run completed for kan-679-flow-improvement-gate-the-draw-at-the-root-never, but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-679-flow-improvement-gate-the-draw-at-the-root-never.md (absent)
skipped: .superpowers/sdd/reviews/kan-679-flow-improvement-gate-the-draw-at-the-root-never-panel.md (absent)
skipped: spectre/changes/archive/kan-679-flow-improvement-gate-the-draw-at-the-root-never/tasks.md (absent)
skipped: spectre/changes/archive/kan-679-flow-improvement-gate-the-draw-at-the-root-never/design.md (absent)
skipped: spectre/changes/archive/kan-679-flow-improvement-gate-the-draw-at-the-root-never/narrative.md (absent)

## git log --stat

commit ba66aacf4f38d6f81f3f6dec0e41cc9c2ceb8cb2
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Wed Oct 7 01:53:53 2026 +0300

    docs(briefs): gate the draw at the root, never enumerate exits

 ...e-the-draw-at-the-root-never-enumerate-exits.md | 33 ++++++++++++++++++++++
 1 file changed, 33 insertions(+)

## Session narrative

The run resolved KAN-679 — a flow-improvement ticket recording, from kan-517's third fix round, the practice of gating a property at its draw site instead of resetting it at every enumerated exit — and judged the deliverable to be one brief in the lessons home, after `flow lesson resolve` on the topic returned no brief and no narrative mention carrying it, which per the lessons contract makes this run the one positioned to promote it. One alternative was considered and set aside: wiring the practice into a run-loaded file (a skill section or a citation from `review-panel.md` or `implement.md`) so runs would meet it directly. That was rejected because nothing consumes briefs mechanically yet, a citation would add a load directive to the run-loaded corpus for no behavioural gain, and the brief format already carries an origin section a later resolve surfaces in full — build-the-simplest-thing, not omission. plan-class classified the plan micro (1 task, 1 file, inline, no panel), the brief was written project-agnostically with the kan-517 history as the Why and the mutation proof as the load-bearing check, and the project's whole lint list ran green in the worktree; the `## test` list scoped to nothing because the diff names no Go package, SPA module or guard harness. Nothing struggled: the only friction was confirming before writing that `docs/briefs/` sits outside the run-loaded corpus `check-verbatim-moves.sh` guards, so new brief prose cannot fail that guard.
