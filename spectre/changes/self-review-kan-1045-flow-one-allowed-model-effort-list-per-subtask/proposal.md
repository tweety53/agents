# self-review-kan-1045-flow-one-allowed-model-effort-list-per-subtask

## Why

The self-review pass of `kan-1045-flow-one-allowed-model-effort-list-per-subtask` found six fixable
defects:

- F-A: `.flow/project.md` names a ported guard's Go companion test `stats/internal/guard/<name>_test.go`,
  while `scripts/run-guard-tests.sh` wants `<name, - replaced by _>_test.go`; the plan followed the
  project text, wrote `modelpolicy_test.go`, and the panel raised a Critical on it.
- F-B: `flow tasks files` can only `add`, so narrowing a `**Files:**` field that
  `check-task-commit-fields` refuses ("declared in Files: but the commit does not touch it") forces
  the hand edit of `tasks.md` that `skills/flow/implement.md` forbids.
- F-C: the findings merger was skipped in both panel rounds against
  `skills/flow/review-panel.md`'s mandatory rule, and nothing refused the fix round that began
  without a `panel-<r>-merger` row.
- F-D: `skills/flow/implement.md`, `scripts/pipeline-fix.sh` and two guard-test comments still cite
  "the pairs table" — deleted from `skills/flow-self-review/SKILL.md` by this change — instead of
  **Allowed pairs** (`skills/flow-contracts/model-policy.md`).
- F-E: the repotest hook tests build flow-guard once per hook call, because `runHook` gives every
  call a fresh `HOME` and no `FLOW_GUARD_CACHE_DIR`; one failure under load was closed on one green
  re-run.
- F-F: `skills/flow-fast/SKILL.md` cites the dispatch closed list for the `description` only, never
  for the `Role:` line the agent-baseline hook refuses a flow-fast dispatch without.

## What changes

- `.flow/project.md` states the companion test's `-`→`_` name with the `check_dispatch_pair_test.go`
  example.
- `flow tasks files <change> <id> remove <path>…` narrows one task's `**Files:**` field, and
  `skills/flow/implement.md` names it for a narrowing.
- `flow record dispatch begin` refuses a `panel-fix-<r>` row (live or back-filled) and a live
  reviewer row of panel round `<r>` while round `<r-1>` holds no completed `panel-<r-1>-merger`
  row; `skills/flow/review-panel.md` states the refusal.
- The four stale citations name **Allowed pairs** (`skills/flow-contracts/model-policy.md`).
- `runHook` pins one flow-guard cache, built once per package run and removed after it, the
  `guardCache` pattern of `stats/internal/guard/helpers_test.go`.
- `skills/flow-fast/SKILL.md` sends the reader to the closed list for the `description` and the
  `Role:` line.
