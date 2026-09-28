# kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong — session narrative

## 2026-09-28 — creating run

- The bare `/flow` that started this change first resolved, unanswered, toward kan-842 — whose
  worktree turned out to be mid-flight in a concurrent session (commits landing during this
  session's own turn). This run stood down there and took kan-820 only on the operator's word.
- The state file read `STARTED` for the whole implementation: `flow.write-in-progress` is the
  only step entitled to advance it, so a crash mid-run leaves `STARTED` behind a fully built
  branch — worth remembering when a resumed run reads an implausibly early state.
- One store outage journaled a stage mark mid-run (`⚠ flow: store unreachable — wrote local
  journal`); it replayed without intervention.
- The base moved mid-run (5 commits on origin/main, overlapping `.flow/project.md` and
  `skills/flow/brainstorm-planner.md`). The operator chose rebase. The rebase was textually
  clean but semantically not: main's new guard-test edits and this branch's appended test
  function merged to drop the `sync` import — a broken build no exit code caught. The gated
  per-task reviewer caught it (its dispatched sha was also stale post-rebase, which is how it
  noticed); fixed as one commit on top and re-reviewed clean. Lesson priced in: a clean rebase
  proves nothing about Go imports.
- The same base movement removed `check-contract-budget.sh` from the tree; task 3's plan step
  and task 5's commit fence both needed record corrections, the latter riding the fields guard's
  own refusal.
- Task 5's file premise was wrong — the run-summary contract lives in
  `skills/flow-contracts/pipeline.md`, not `verify-and-handoff.md` — corrected through the same
  refusal-then-transcribe route.
- Of five AskUserQuestion prompts this run asked, two went unanswered (candidate resolution,
  first plan gate), two were answered (design gate, base-moved), and the plan gate answered on
  the operator's re-invocation. The panel then raised 3 Important + 2 Minor; both Importants
  were real defects in the resolution block (body normalization, outage prose), fixed in one
  round, re-run clean, one new Minor deferred.
