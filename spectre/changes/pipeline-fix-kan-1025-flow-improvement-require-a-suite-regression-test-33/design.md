# pipeline-fix-kan-1025-flow-improvement-require-a-suite-regression-test-33

## Context

`check-panel-findings-closed` reads `<dir>/.superpowers/sdd/<name>/`. `gatePanelFindingsClosed` passes `f.dir` or `"."`; a /flow parent runs from the main checkout (`skills/flow/implement.md` forbids `cd`-led compounds), where the change's files do not exist until it lands. `changeCheckouts` (`stats/cmd/flow/record.go`) already finds every worktree carrying a change, for `requireChangeCheckout`.

## Decisions

### How the gate finds the change without -C

**ID:** panel-gate-resolves-worktree
**Status:** active
**Chosen:** when `-C` is absent, take the first checkout `changeCheckouts` lists as carrying the change that is not the main checkout, else the main checkout if it carries it, else `"."` — the prescribed command line then works unchanged from any checkout.
**Considered:** require `-C` in every contract line plus a refusal hint — rejected: four doc lines and a guard test change, and every future caller can still drop the flag; resolution closes it at the one function all callers route through.

### Planning effort for the first review

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-low — one small Go branch reusing an existing helper plus one test; planners capped at low until 2026-10-28.
**Considered:** flow-medium — rejected: nothing in the diff calls for wider reasoning.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group — Task 1 alone.
**Considered:** none — a single task.

## Open questions

## Recommended approach

Edit `gatePanelFindingsClosed` only; the docs' `-C`-less line stays correct once the gate resolves the worktree.

## Critical files

- `stats/cmd/flow/stage.go` — `gatePanelFindingsClosed`.
- `stats/cmd/flow/stage_test.go` — the new regression test beside `TestStageEndReviewPanelRunsFindingsClosed`.

## Existing code to reuse

- `changeCheckouts` in `stats/cmd/flow/record.go`.
- `gitRepo`, `isolatedStateRoot`, `genuineDaemon` test helpers and the fake-`flow` PATH stub in `TestStageEndReviewPanelRunsFindingsClosed`.

## Verification

- `cd stats && go test ./cmd/flow -run 'TestStageEndReviewPanel'` passes; the new case fails on `origin/main`.
- `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean.

## Live check

None — the gate runs in process inside the `flow` CLI; the regression test exercises it end to end against a real git worktree pair. No service or persistent state is touched.
