## Context

Found by the self-review pass of `self-review-backlog-flow-6` (worktree
`/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-6`, `link.md` merge order `.`
then `gymie`). Task 5 of its plan landed in gymie's worktree
`/Users/tweety53/Projects/gymie-worktrees/self-review-backlog-flow-6` as `8df01571 fix(profile):
hold the friendship row lock through setLevel's check and write`. Reproduced: `scripts/check-task-records.sh`
run in the agents worktree exits 1 with `spectre/changes/self-review-backlog-flow-6/tasks.md: task 5
ticked but no commit with subject 'fix(profile): hold the friendship row lock through setLevel's
check and write' reachable from HEAD`.
<!-- measured: scripts/check-task-records.sh in /Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-6 @ cfb9b564 → exit 1, the line above -->

Cause: `check-task-records.py` builds one subject map, from the linted repository's HEAD, and
judges every ticked task against it. The plan's tasks that land in a peer repository have no commit
there by construction.

What already names the change's other worktrees: the state record's `worktrees` map. It is
`integrate-change.sh`'s primary source for the worktree set and what `changePlanStoreDir`
(`stats/internal/guard/changeplan.go`) reads to find a satellite's plan. For the running change
`flow state get -C <agents worktree> self-review-backlog-flow-6` lists both worktrees.
<!-- measured: flow state get -C /Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-6 self-review-backlog-flow-6 @ cfb9b564 → worktrees map holds the gymie and agents worktrees -->
`check-task-commit-fields.sh`'s commit map is a per-call argument its caller builds; it gives this
whole-plan lint nothing to read, so this guard reads the record the caller builds that map from.

The diff in Task 1 of `tasks.md` was applied to this worktree at `cfb9b564` and checked there: run
against the agents worktree of `self-review-backlog-flow-6` it exits 0, and `8df01571`'s seven files
match Task 5's `**Files:**` exactly.

Blast radius: 17 files outside archives cite `check-task-records` (`git grep -l check-task-records`);
<!-- measured: git grep -l check-task-records -- . ':!spectre/changes/archive' ':!spectre/changes/pipeline-fix-self-review-backlog-flow-6-23' | wc -l @ cfb9b564 → 17 -->
the fix touches two of them, the guard's Python module and its test harness. No saved memory encodes
this defect.


## Decisions

### Where a peer task's commit is looked up

**ID:** peer-worktrees-from-state-record
**Status:** active
**Chosen:** the change's state record — `flow state get -C <repo-root> <change>`'s `worktrees` map, every listed directory but the linted root, each searched from its HEAD for the subject — the record the run already keeps of the change's worktrees.
**Considered:** `link.md`'s merge order resolved through `<spec>/peers` to each peer's main checkout, then `git worktree list` matched by the linted branch's name — offline, but it assumes every repository's worktree carries the same branch name, which nothing guarantees; a `**Repo:**` field on each task — a new plan grammar every planner must write, for what the record already states; skipping a ticked task none of whose `**Files:**` exist in this tree — drops the check for every peer task instead of judging it.

### Only a cross-repo change, only on a miss

**ID:** lookup-gated-on-link-and-miss
**Status:** active
**Chosen:** the record is read only for a plan whose change directory carries a `link.md`, only once a ticked subject misses locally, and once per plan — a single-repo plan, and every test fixture without a `link.md`, never calls `flow`.
**Considered:** reading the record for every plan — a store round-trip on every lint run of every change, for nothing.

### Unanswerable lookup keeps the violation

**ID:** no-record-keeps-violation
**Status:** active
**Chosen:** no `flow` on PATH, a non-zero `flow state get`, unparseable output or no peer reaching the subject searches nothing, and today's `ticked but no commit` violation is reported.
**Considered:** exit 2 (could not judge) when the store is down — turns every cross-repo lint run without flowd into a refusal, where today's verdict is still the correct one for a task truly missing its commit.

### What a peer commit is judged on

**ID:** peer-commit-judged-in-peer
**Status:** active
**Chosen:** a commit found in a peer feeds check 2's file comparison, run with `git diff-tree` in that peer; the unticked-task check and the `expected one` commit count stay on the linted repository's own `<base>..HEAD`, since a peer's base is not this guard's to resolve.
**Considered:** resolving each peer's base to extend both checks — a second base-resolution path for a lint whose false positive is the tick check alone.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — a two-file diff to a lint guard, one gated lookup, and case 17 pinning it.
**Considered:** `flow-medium` — the guard is a `## lint` line every run executes, but the change only removes a false violation behind a `link.md` gate and keeps every other verdict unchanged.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — the plan holds one task.
**Considered:** none — a single task has no split.

## Open questions

## Live check

From this worktree, `scripts/check-task-records.sh /Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-6`
exits 0 with no output; the same call through the main checkout's guard at `cfb9b564` exits 1 with
the Task 5 line above. An exit 1 naming Task 5 after the fix means the peer lookup never ran; an
exit 2 or a Task 5 file-mismatch line means the peer commit was judged in the wrong repository.
