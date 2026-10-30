# self-review-kan-1067-flow-allowed-pairs-experiments-stored-in-flowd

## Why

The self-review pass of `kan-1067-flow-allowed-pairs-experiments-stored-in-flowd` found six fixable
defects (evidence: `.superpowers/sdd/self-review-pass.md` in the change's worktree):

- 1-A: `check-task-records` check 2 judges only the commit carrying a task's subject, while
  `check-task-commit-fields` joins every `Task-Id: <n>` fix commit (KAN-976); once a gated fix's
  paths were declared, the first refused them at verify.
- 1-B: `close-task` runs `check-plan-shape` only for appended tasks, so a `flow tasks files add`
  widening that breaks plan shape (a test path under `Tests: none`, F4) surfaces only at verify.
- 1-C: a verify re-run keyed off the contract (`verify-fix-1b`) forms its own group in
  `check-verify-green` and leaves the blocked row latest — `VERIFY-NOT-GREEN` with no hint why.
- 1-D: the mutation slot's throwaway copy cannot compile `internal/api` tests: `experiments_test.go`
  imports `modelpolicy/embedded`, whose gitignored file only `make` writes.
- 1-E: `scripts/live-store-copy.sh` refuses a hyphenated name such as `kan-1067`.
- 1-F: task verify steps hand-pick a subset of the `## lint` guards, so `check-ask-silence.sh`
  failures reached verify.

## What changes

- `check-task-records` check 2 judges the union of the subject commit's files and every
  `Task-Id:` commit's files naming the task.
- `close-task` runs `check-plan-shape` on every close.
- `run-verify-list` refuses, exit 2 before its first command, a recording key that does not fold to
  `verify` or `verify-<repository name>`.
- `internal/api`'s experiments tests read `skills/flow-contracts/model-policy.md` directly.
- `live-store-copy.sh` admits `-` in a name and maps it to `_` in the database name only.
- Writing-plans makes a task that edits Markdown under `skills/` or `rules/` verify with the whole
  `## lint` list.
