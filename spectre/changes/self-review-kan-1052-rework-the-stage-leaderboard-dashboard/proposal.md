# self-review-kan-1052-rework-the-stage-leaderboard-dashboard

## Why

The self-review pass of `kan-1052-rework-the-stage-leaderboard-dashboard` found seven fixable
defects (`findings.md` of the pass, F1–F6 and F16):

- F1: `.flow/project.md`'s `ui paths` misses `skills/flow-contracts/model-policy.md`, which flowd
  embeds and the Experiments view renders, so an edit to it ran no `verify` row and left five
  Experiments baselines stale.
- F2: `/flow-fast` and `/flow-interactive` land a `ui paths` diff without running the
  `## visual verification` `verify` row, so stale baselines and a swallowed hover border reached
  `main` unseen.
- F3: `skills/flow/implement.md` gives the implementer key as `task-<n>-implementer`, but a group's
  launch needs `task-<ids>-implementer`, and the hook denied the documented form.
- F4: `scripts/test-run-capped.sh` case (a) asserts a wall-clock bound that fails under load though
  the kill it checks happened.
- F5: the `## test` list's concurrent Go suites each take every core, starving the SPA suite until
  a 235 ms vitest test passed its 5 s timeout.
  <!-- measured: implementer-report-full-suite-fix-2.md of kan-1052 — 235 ms alone, suite 67.7 s concurrent, vitest 5 s timeout -->
- F6: the mutation slot's own prompt never says not to symlink a shared `node_modules` into its
  throwaway copy.
- F16: the planning prompts carry no rule that a plan-specified test's fixture must tell the rule
  from its nearest wrong alternative; two such tests passed wrong code.

## What changes

- `ui paths` names `skills/flow-contracts/model-policy.md`; a guard test requires every source
  `stats/Makefile` copies into flowd's embed to match `ui paths`.
- `run-verify-list.sh <worktree> visual` runs the `verify` row alone when the diff touches
  `ui paths`; `/flow-fast` and `/flow-interactive` run it before landing and recapture an intended
  drift with `capture`.
- `implement.md` and two code comments give the implementer key as `task-<ids>-implementer`, ids
  `+`-joined in plan order.
- `test-run-capped.sh` case (a) proves the kill by a completion marker the killed child never
  writes, with no wall-clock bound.
- The `guard-tests` and `stats-go` lines of `## test` run under a `GOMAXPROCS` share of the cores,
  measured before and after; a guard test pins it.
- The mutation-testing brief tells the slot to run `npm ci` in its copy and never symlink a shared
  `node_modules`.
- `skills/spectre-planning/SKILL.md` gains a discriminating-fixture rule, re-copied into spectre's
  vendored copy.
