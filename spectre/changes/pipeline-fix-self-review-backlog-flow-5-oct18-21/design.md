## Context

Found by `/flow` on `self-review-backlog-flow-5-oct18` (a `/flow` change, STARTED, branch
`spectre/self-review-backlog-flow-5-oct18` in agents, gymie and gymie-playwright): its first
review-panel dispatch, key `panel-0-primary+principles`, no `Key:` line, was denied with
`Loop dispatch blocked: the prompt names self-review-backlog-flow-5-oct18 but carries no ...`.
Reproduced at `52438f0d` against the real slot prompt
(`<agents>-worktrees/self-review-backlog-flow-5-oct18/.superpowers/sdd/slot-prompt-0-primary+principles.md`,
no `Key:` line, 27 mentions of `/self-review-backlog-flow-5-oct18`) piped as an Agent payload with
`cwd` `/Users/tweety53/Projects/agents` into the installed hook: deny.

Cause: `LOOP_NAME_RE` matches any token starting `self-review-`, so a `/flow` change named that way
is indistinguishable by name from a self-review loop.

What does not discriminate, measured on the dev store (read-only) and git:

- A state record: `pipeline-fix.sh start` writes a STARTED record for every loop, `updatedBy`
  `/flow`, like a `/flow` change's.
- The record's `branch` field: 22 of the 23 loop-prefixed records carry `branch` equal to their
  name and `self-review-backlog-flow-5-oct18` carries none — it would discriminate, but needs a
  flowd round-trip in a hook and fails open when the record is read in another project's store.
- The name: loop names are `self-review-<reviewed change>`, and a `/flow` change name is free text.

What does: the branch. `kickoff-worktree.sh` gives every `/flow` change the branch
`spectre/<name>`; `pipeline-fix.sh start` gives a loop the branch `<name>` itself. Of every
`pipeline-fix-*` and `self-review-*` branch in agents (local and `origin`) none has a
`spectre/<name>` twin, and none exists in gymie either; `spectre/self-review-backlog-flow-5-oct18`
exists locally and on `origin` in agents and gymie.
<!-- measured: for each pipeline-fix-*/self-review-* ref in git for-each-ref refs/heads refs/remotes/origin, rev-parse spectre/<b> and origin/spectre/<b> — no hit @ 52438f0d; git -C /Users/tweety53/Projects/gymie for-each-ref | grep spectre/(self-review|pipeline-fix)- prints only spectre/self-review-backlog-flow-5-oct18 and its origin twin -->

Blast radius: 11 files outside `spectre/` cite `enforce-loop-dispatch` or `LOOP_NAME_RE`
<!-- measured: grep -rlE 'enforce-loop-dispatch|LOOP_NAME_RE' . --exclude-dir=spectre --exclude-dir=.git | wc -l @ 52438f0d → 11 -->
(`grep -rlE 'enforce-loop-dispatch|LOOP_NAME_RE' . --exclude-dir=spectre --exclude-dir=.git`); the
fix touches two of them, the hook and its test harness. No saved memory encodes this defect.

## Decisions

### Discriminator between a loop and a /flow change of the same prefix

**ID:** spectre-branch-discriminator
**Status:** active
**Chosen:** a named token is a `/flow` change's own when `spectre/<token>` exists, as `refs/heads/` or `refs/remotes/origin/`, in the repository of the payload's `cwd` — one `git for-each-ref` call, run only when the prompt carries no Key line and names a loop-shaped token.
**Considered:** the state record's `branch` field — a flowd call inside the hook, and a record read in the dispatching session's project misses a loop filed under the agents project, so it fails open exactly where the net is meant to hold; reserving the `self-review-`/`pipeline-fix-` prefixes for loops — leaves `self-review-backlog-flow-5-oct18` blocked until it is renamed mid-run across three repositories; a marker every loop prompt carries — the net exists for the prompt that dropped its marker.
**Residual risk:** a loop sharing its exact name with an open `/flow` change's `spectre/<name>` branch is admitted without its Key line; a self-review loop's name is `self-review-<reviewed change>`, so this needs a `/flow` change named after a self-review of another — none exists (measured above).

### Fail closed when the discriminator cannot answer

**ID:** no-cwd-keeps-deny
**Status:** active
**Chosen:** no usable `cwd`, a `cwd` outside a git repository, or a git call that fails or times out keeps today's deny.
**Considered:** reading the hook process's own working directory when `cwd` is absent — the payload always carries `cwd`, and a guessed directory could admit a keyless loop dispatch.

### Every named token is checked

**ID:** every-token-checked
**Status:** active
**Chosen:** the deny fires on the first loop-shaped token that is not a `/flow` change's own, scanning them all — a prompt naming the running `/flow` change and a real loop worktree is still denied.
**Considered:** checking only the first match — a `/flow` change's name early in a prompt would hide a later loop name.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — a two-file diff whose behaviour case 13 pins both ways and whose only widening is an allow for tokens with a `spectre/` twin.
**Considered:** `flow-medium` — the hook gates loop dispatches, but the change only narrows an existing deny, and the live check and the harness cover both directions.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — the plan holds one task.
**Considered:** none — a single task has no split.

## Open questions

## Live check

From this worktree, pipe the real slot prompt above as an Agent payload (`cwd`
`/Users/tweety53/Projects/agents`) into `hooks/enforce-loop-dispatch.py`: no output, exit 0 (an
allow); the same payload into the main checkout's hook at `52438f0d` prints the deny. Then the same
payload with the prompt's `self-review-backlog-flow-5-oct18` replaced by `self-review-no-such-loop`
still prints a deny naming `self-review-no-such-loop`. An allow on that last call means the
discriminator admits every name.
