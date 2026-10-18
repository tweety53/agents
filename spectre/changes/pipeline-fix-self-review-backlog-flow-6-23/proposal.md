# pipeline-fix-self-review-backlog-flow-6-23

## Why

`scripts/check-task-records.sh`, a `## lint` line of `.flow/project.md`, reports a ticked task as a
violation when no commit reachable from the linted repository's HEAD carries its `**Commit:**`
subject. A cross-repo change keeps its plan in the canonical repository while some of its tasks land
in a peer repository's worktree, so every such ticked task is a false violation and the lint line
stays red for the whole run. The running `self-review-backlog-flow-6` hits it: its Task 5 landed in
gymie as `8df01571`, and the guard in the agents worktree exits 1 naming that task.

## What changes

- For a change whose directory carries a `link.md`, a ticked task whose subject this repository
  does not reach is looked up in the change's other worktrees — the `worktrees` map of the change's
  state record — and a commit found there is judged, files and all, in that repository.
- With no record, no `flow` CLI, or no peer carrying the subject, the violation is reported as
  before.
- `scripts/test-check-task-records.sh` gains case 17: a ticked task whose commit lives only in a
  peer worktree exits 0.
