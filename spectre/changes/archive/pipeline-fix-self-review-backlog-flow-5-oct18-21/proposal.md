# pipeline-fix-self-review-backlog-flow-5-oct18-21

## Why

`hooks/enforce-loop-dispatch.py` denies every Agent dispatch whose prompt names a token shaped
like a loop's branch (`pipeline-fix-<name>-<k>`, `self-review-<name>`) but carries no loop `Key:`
line. A `/flow` change may itself be named `self-review-…` — the running
`self-review-backlog-flow-5-oct18` is — and then every dispatch of that run (review panel,
verifier, any other) names its worktree and is denied as a loop dispatch that dropped its Key line.
No dispatch of such a run can ever go out.

## What changes

- The keyless-name deny skips a named token that is a `/flow` change's own: a `spectre/<token>`
  branch, local or on `origin`, in the repository of the payload's `cwd`. A loop's branch is its
  bare name, never `spectre/<name>`, so a loop dispatch that dropped its Key line is still denied.
- `scripts/test-enforce-loop-dispatch-hook.sh` gains case 13, the `/flow`-change allow and its
  deny twin without the branch.
- Once landed and pulled into the main checkout the hook is installed from,
  `self-review-backlog-flow-5-oct18`'s panel dispatch goes out unchanged.
