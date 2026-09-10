# kan-463-flow-planner-splits-desktoptest-ui-tests-into

## Why

In the gymie repos, 63% of implementer wall clock is model latency times turn count, not Gradle
(12%). The worst tail is an implementer grinding on Compose `desktopTest` UI tests inside the task
that also carries the feature's source files, re-reading a 450–585k-token context on every
iteration — kan-425 task 6, three source files bundled beside the `desktopTest` file, looping on
`aRevokedGrantRemovesTheLoggingButtonWithoutNavigating`, is the worst of them:
56 minutes, 300 API calls, 44 Gradle runs. kan-423 and kan-455 carry the same task shape.
<!-- measured: every figure in this section is the survey of 78 implementer dispatches recorded in
     docs/superpowers/research/flow-gymie-implementation-speedup.md thread 1, not re-measured here -->

`/flow` has no granularity rule of its own. `skills/flow/brainstorm-planner.md` section **D** hands
task splitting entirely to `superpowers:writing-plans`, and the de-facto gymie pattern fuses
"feature plus its `desktopTest` file" into one task — kan-423 has eleven of twelve frontend tasks
carrying exactly one `desktopTest` file beside 3–8 source files. Nothing the planner reads says
otherwise.

Source: `docs/superpowers/research/flow-gymie-implementation-speedup.md`, thread 1 and its
step-by-step item "Split `desktopTest` UI tests into their own follow-on task".

## What changes

- `skills/flow/brainstorm-planner.md` section **D** gains one blockquote, placed immediately after
  the verify-step paragraph KAN-465 added and before the plan-provenance load, telling the planner
  to write a feature whose tests live in a UI-test source set of their own (Compose Multiplatform's
  `desktopTest` named as the example) as two tasks: the feature task with its source and unit-test
  (`commonTest`) files, then one follow-on task per feature task whose `**Files:**` names only that
  feature's UI-test file(s), all of them. The two tasks are file-disjoint, so
  `plan-dispatch-bundles.sh` keeps them separate bundles and the UI-test iteration starts in a
  fresh implementer at a small context.
- Nothing else. No guard script, no edit to `scripts/plan-dispatch-bundles.py` or
  `skills/flow/implement.md`, no change to any existing plan.

## Non-goals

- **A `check-plan-shape.sh` rule detecting a task whose `**Files:**` mixes a UI-test path with
  source paths.** Declined by the operator during brainstorming: the fix statement is planner
  prose, and the guard would be a new heuristic (which paths count as UI-test paths, per project)
  with its own false positives. If a future plan is found fusing the two anyway, that is the
  evidence that would justify it.
- **A files-per-task cap and an implementer failure budget.** Both declined during research
  (`flow-gymie-implementation-speedup.md` `## Open / undesigned`) and not re-opened here.
- **Any change to the pre-panel full-suite run.** The last bundle's FULL SUITE paragraph and
  `flow.verify` still run the resolved `## test` list once each; the operator confirmed that gate.
