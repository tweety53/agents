# kan-491-extend-stage-vocabulary-for-flow-fast

## Why

Every stage mark `/flow-fast` is scripted to make is rejected by the CLI. `skills/flow-fast/SKILL.md`
and its phase files carry `-command '/flow-fast'` on every mark, but the stage vocabulary —
README.md's "Level 1 — the stages of each command" table and its hand-authored transcription in
`stats/internal/stages/names.go`, kept identical by `TestStagesMatchReadmeLevelOne` — defines only
`/flow`. The kan-357 run (2026-09-10) hit this on its first mark and proceeded without any stage
marks; nothing else consumes marks, so the defect is silent except at the CLI.

## What changes

`/flow-fast` joins the Commands column of exactly the rows flow-fast marks — every `flow.*` row
except `flow.design-approval`, `flow.visual-verify`, `flow.verify-cleanup` and `flow.self-review`,
the four its own Stage keys table never names — in both the README table and `stages.Table`, plus
the `FlowFast` command constant. A new drift guard derives flow-fast's key set from the SKILL.md
Stage keys table at test time and asserts the vocabulary matches it, so the skill and the
vocabulary cannot drift apart again the way they did here.

**After landing** an operator must rebuild and reinstall the CLI (`cd stats && make restart`, which
also restarts the protected dev daemon) before any `/flow-fast` mark validates.
