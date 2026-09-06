# kan-465-flow-planner-must-write-targeted-tests

## Why

`skills/flow/implement.md` §4's `**TARGETED TESTS:**` paragraph already forbids the implementer
from running a module or repository suite mid-task: it names the build tool's own selector, and
says the full `## test` list runs once per worktree at the last bundle and again in `flow.verify`.
Since KAN-441 landed that paragraph on 2026-09-05, implementers obey it — across the four changes
run after it (kan-424, kan-425, kan-454, kan-455) no task carrying named tests exceeded two
whole-module invocations, against thirteen such tasks in the seven changes before.

The remaining cost is written into the plan itself. Twenty-four task-step lines across kan-421,
kan-422, kan-423, kan-425 and kan-455 end a task's verify step with a bare
`./gradlew :shared:desktopTest`, and the implementer — correctly, per its own instructions —
follows the plan as written. kan-454's planner wrote `--tests` selectors in five steps and its
tasks ran 0–2 whole-module invocations; kan-425's planner wrote the bare module in all eight of
its tasks. The same habit is present in this repository's own plans:
`spectre/changes/archive/kan-173-make-the-daemon-s-dependencies-required-so-a/tasks.md` step 7 is
`cd stats && go test ./... -race -count=1`.

Nothing in `skills/flow/brainstorm-planner.md` section **D** — the writing-plans enrichment, which
tells the planner the spectre task shape, the plan-provenance tags and the
`flow-task-commit-fields` family — says anything about what a task's verify step may run. That is
the gap. The run it costs is the one the last bundle's FULL SUITE paragraph and `flow.verify`
already make — one 50–110 s `:shared:desktopTest` per affected task, 8–15 minutes on a 7–9-task
frontend change.
<!-- measured: every figure in this section is the survey of 77 implementer dispatches recorded in
     docs/superpowers/research/flow-gymie-implementation-speedup.md thread 3, not re-measured here -->

Source: `docs/superpowers/research/flow-gymie-implementation-speedup.md`, thread 3 and its
step-by-step item "A task's verify step names targeted tests, never the module suite".

## What changes

- `skills/flow/brainstorm-planner.md` section **D** gains one blockquote, placed immediately after
  the task-shape paragraph and before the plan-provenance load, telling the planner how to write a
  task's verify step: the lint commands the task's own `**Files:**` need, the build tool's own
  selector for each `**Tests:**` entry, never the bare module or repository suite, and lint alone
  for a task whose `**Tests:**` is `none`.
- Nothing else. No guard script, no edit to `skills/flow/implement.md`, no change to any existing
  plan — `spectre/changes/archive/` is frozen and the five gymie plans live in another repository.

## Non-goals

- **A `check-plan-shape.sh` rule enforcing the sentence.** Offered during research and declined:
  the post-KAN-441 data shows the implementer follows whatever the plan says, so correcting the
  planner's instructions is the whole fix, and a guard would be a second enforcement point for a
  failure mode the prompt now covers. If a future change's plan is found writing a bare suite
  anyway, that is the evidence that would justify the guard.
- **Any edit to `skills/flow/implement.md`.** Its TARGETED TESTS and FULL SUITE paragraphs are
  already correct. This change aligns the plan to them rather than restating them.
