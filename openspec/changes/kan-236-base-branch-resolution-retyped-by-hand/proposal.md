## Why

The block that resolves a repository's base branch — fetch, `refs/remotes/origin/HEAD`, the
`git remote show origin` fallback, and three assertions — exists only as fenced prose inside
`skills/myflow-contracts/finish-contract.md`, and every `/myflow-finish` run retypes it. The
contract's own comments say what a wrong answer costs: a base ref resolved from `HEAD`'s upstream
compares `openspec/<name>` against `origin/openspec/<name>`, which is true the moment the branch is
pushed, so an unmerged change is reported merged — after which run 2 archives it and
`--force`-removes its worktree. That is the most destructive step in the pipeline, gated by a block
that is copied by hand.

`/myflow-status` hand-derives the same value with no stated source at all: its merge-status step 3
runs `git merge-base --is-ancestor <branch> origin/<base>` and names nothing that produces `<base>`,
so the assertion that catches the misresolution is not merely retyped there — it is absent.

## What Changes

- **New guard `resolve-base-branch.sh`** owning the whole resolution: the wrapped fetch, both
  resolution paths, and all three assertions — non-detached `HEAD`, non-empty base, and base ≠ the
  current branch. It prints the bare branch name on stdout and nothing else, so a caller composes it
  directly, and writes every refusal to stderr with an empty stdout under a four-value exit contract
  that separates a named refusal, an unreadable tree, and a repository with no `origin` remote.
- **The finish contract's fenced block becomes a one-line invocation.** Its prose stays: the
  `HEAD@{upstream}` warning and the no-remote message are the reasoning a reader needs in order not
  to reintroduce the block, and both are re-pointed at the guard.
- **`/myflow-status` resolves its base through the same guard**, in the worktree it is already
  visiting. A non-zero exit maps to **inconclusive**, never to a stop — the command stays read-only
  and never blocks.
- **`/myflow-status` acquires its first guard**, which reverses three claims that currently assert
  it has none: its "No guard-presence check here" block, its "carries no `scripts/` directory"
  sentence, and `check-guard-symlinks.sh`'s `declare_if_present "myflow-status"` call.
- **`check-finish-preflight.sh`'s signature is unchanged.** It keeps taking `<base-ref>` and still
  resolves nothing itself; only its header comment moves from "the resolution stays in the finish
  contract" to "the resolution stays in the sibling guard."
- **A mutation-proving test harness** `scripts/test-resolve-base-branch.sh`, added to
  `.myflow/project.md`'s `## test` list.

No behaviour visible to an operator changes on a correct run: the same base branch resolves, and the
same refusals are reported in the same order.

## Capabilities

### New Capabilities

- `myflow-base-branch-resolution`: how the base branch is resolved — one guard owning the fetch,
  both resolution paths, the three assertions and the name validation; its stdout/stderr and exit
  contract; the rule that every consumer invokes it rather than re-deriving it; and which directory
  a consumer passes it.

### Modified Capabilities

*(none — no existing requirement's behaviour changes. `myflow-finish-cleanup`'s "the base branch is
resolved, never assumed" scenario stays true as written; this change states where that resolution
lives, which no existing requirement fixes. `myflow-contract-distribution`'s guard-reachability
requirements are complied with, not altered.)*

## Impact

| Path | Change |
|------|--------|
| `scripts/resolve-base-branch.sh` | new guard |
| `scripts/test-resolve-base-branch.sh` | new test harness |
| `skills/myflow-finish/scripts/resolve-base-branch.sh` | new relative symlink |
| `skills/myflow-status/scripts/resolve-base-branch.sh` | new directory, new relative symlink |
| `skills/myflow-fast/scripts/resolve-base-branch.sh` | new relative symlink |
| `skills/myflow-contracts/finish-contract.md` | fenced block → invocation; prose kept, re-pointed |
| `skills/myflow-finish/SKILL.md` | guard-presence paragraph; base-resolution paragraph re-pointed |
| `skills/myflow-status/SKILL.md` | step 3 invokes the guard; three no-guard claims reversed |
| `scripts/check-finish-preflight.sh` | header comment re-pointed |
| `scripts/check-guard-symlinks.sh` | `declare_if_present "myflow-status"` removed |
| `scripts/test-check-guard-symlinks.sh` | its F4c fixture's stand-in renamed off `myflow-status`, which that declaration's removal breaks |
| `.myflow/project.md` | `## test` gains the new harness |

**Not in scope.** KAN-224 (`myflow workspace-id`) is the same shape of finding for the workspace id
and stays its own change; nothing here presumes its route. Adding a `-` self-resolution mode to
`check-finish-preflight.sh` was considered and rejected — alongside the guard it would ship two
mechanisms for one job.
