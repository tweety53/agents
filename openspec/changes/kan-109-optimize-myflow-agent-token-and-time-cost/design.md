# Design — KAN-109 Optimize myflow agent token and time cost

Source: `docs/superpowers/specs/2026-08-12-kan-109-optimize-myflow-agent-token-and-time-cost-design.md`.

## Context

The measurement in KAN-109 was taken on one real change (KAN-71 fix-1) and separates cost that is
inherent to the work from cost that is an artifact of how `/myflow-do` dispatches. Four properties
account for the artifact half, and each has a distinct fix. Two of the four are mechanically
checkable from files that already exist — a plan's `**Files:**` fields and a branch's diff — and get
guard scripts. The other two govern how the parent agent paces and scopes dispatches; there is
nothing on disk for a script to read, so they live in the skill and the delta specs.

## Rule 1 — one implementer per worktree, serialized

At most one implementer subagent is in flight against a given worktree at any moment. The parent
waits for a dispatch to report its commit sha, runs `check-task-commit-fields.sh`, and dispatches
that task's review before starting the next bundle.

Parallelism across *different* worktrees is untouched: a change affecting three repositories still
dispatches into three worktrees at once. What is banned is two dispatches sharing one build
directory, which is what produced the daemon collisions, the 60s idle waits, and the corrupted
test-result XML.

This overrides `superpowers:subagent-driven-development` and `superpowers:dispatching-parallel-agents`
where the tasks share a worktree — the third override myflow carries against that upstream skill,
alongside the two under **Model policy**.

## Rule 2 — bundle dispatches by shared files

`scripts/plan-dispatch-bundles.sh` is a thin Bash wrapper over `scripts/plan-dispatch-bundles.py`,
the split `check-task-build-green.sh` and `check-task-commit-fields.sh` already use, for the same
reason: Markdown block structure deserves a parser rather than a hand-rolled Bash ERE allowlist.

**Input:** a `tasks.md` path; with no argument, every non-archived change's `tasks.md` under
`openspec/changes/*/tasks.md`, exactly as `check-task-build-green.sh` resolves them.

**Algorithm:** read each **unchecked** task's dotted id and its `**Files:**` field; join two tasks
when they declare any common path; take the transitive closure with union-find. An
`**Allowed-collateral:**` glob is not part of the join — it names paths a legitimate sweep may
touch, not paths the task owns, and joining on it would collapse unrelated tasks into one bundle.

**Output:** one line per bundle, ordered by the bundle's lowest task id, each listing that bundle's
task ids.

**Exit codes:** 0 bundles computed; 1 a task carries no `**Files:**` field, named by task id; 2 the
guard cannot answer at all.

**Where it runs:** `/myflow-do` §4 only, at dispatch time. Not a lint step — it computes a grouping
rather than judging a file's text, and a lint run has no change in flight to compute one for.

A `Build: red` task and its `Squash-with:` partner touch the same files by construction, so they
normally land in one bundle; the implementer still makes one commit per task and folds red into
green exactly as §4 already requires. Nothing about the fixup-and-autosquash mechanism moves.

## Rule 3 — panel diff-size cap

`scripts/check-panel-diff-size.sh <worktree> <merge-base> [cap]` — pure Bash over git, in the shape
of `check-finish-preflight.sh`, since it parses no Markdown. It sums insertions and deletions across
committed work and the unstaged working tree, the same body of text `final-review.diff` carries.

Default cap **2000** changed lines, keeping a full pass in the low hundreds of thousands of tokens
while leaving ordinary myflow changes untouched.
<!-- measured: KAN-109's measurement section — 948k tokens for a 7-slot pass over 9877 lines, taken on the KAN-71 fix-1 session -->

The cap is a positional third argument so a project can move it without editing the guard.

Exit 0 at or under the cap, 1 over it, 2 cannot answer.

`/myflow-do` §5 runs it immediately before writing `final-review.diff`. On exit 1 the operator gets
a named-options prompt shaped per **Operator prompts**: *proceed with the panel anyway* (default,
recommended) or *stop — I'll split the change*, which ends the run at `IN_PROGRESS` with nothing
lost. The measured count, the cap in force, and the answer are recorded in the panel record on every
run, including exit-0 runs, so the record always states the size the panel read.

Exit 2 is a stop for the same reason a workspace-isolation exit 2 is: a size the guard could not
measure is not a size under the cap.

## Rule 4 — conditional slots re-run only when their subject changed

Targeted re-runs are already scoped and are unchanged. In **Full** escalation mode:

- required slots always re-run against the rewritten `final-review.diff`;
- a conditional slot re-runs only when its own row in the optional-slot trigger table still fires
  against `fix-round-N.diff`;
- a conditional slot whose trigger did not fire is not re-run, and its pass-1 result stands, recorded
  as `not re-run — subject unchanged`.

This is a definition of non-stale rather than a waiver. A result is stale when the diff it read has
since changed *in the region that slot reads*. A conditional slot's region is exactly its trigger's
subject; a fix touching nothing in that subject leaves its reading current. Required slots have no
bounded region — they read the whole diff — which is why the scoping reaches conditional slots
alone. The zero-open-findings bar is untouched: a conditional slot that raised an open finding still
blocks the handoff whether or not it re-runs, and a slot that is not re-run cannot close its own
finding by not looking again.

## Decisions

### Enforce the mechanically-checkable rules with guard scripts rather than prose alone

**ID:** guards-not-prose
**Status:** active
**Chosen:** two guard scripts, each with a `test-*.sh` harness — a real ratchet for the two rules
that can be measured from files.
**Considered:** prose and delta specs only — cheapest, and consistent with how dispatch rules are
carried today, but leaves the diff-size cap and the bundling grouping as things an agent may or may
not do with no way to tell after the fact; a mixed option scripting only the cap — rejected because
the bundling computation is the more error-prone of the two to do by eye.

### The over-cap response is a prompt, not a hard block

**ID:** cap-asks-rather-than-blocks
**Status:** active
**Chosen:** a named-options prompt with *proceed* as the default and recommended answer.
**Considered:** a hard block until the change is split — strands work already implemented in a
worktree, turning a cost optimization into a wall; an advisory line only — zero friction and zero
enforcement, which leaves the measured waste exactly where it was.

### Cap at 2000 changed lines

**ID:** cap-two-thousand
**Status:** active
**Chosen:** 2000, roughly a fifth of the measured diff, keeping a full seven-slot pass in the low
hundreds of thousands of tokens.
<!-- measured: KAN-109's measurement section — 948k tokens for a 7-slot pass over 9877 lines -->

**Considered:** 1000 — would prompt on a fair share of ordinary myflow changes, training the
operator to dismiss it; 3000 — only genuinely large redesigns would trip it, which leaves most of
the measured waste unaddressed.

### Bundles are computed at dispatch time, not at plan-writing time

**ID:** bundles-at-dispatch
**Status:** active
**Chosen:** `/myflow-do` §4 runs the script; `/myflow-start`'s writing-plans stage is untouched.
**Considered:** running it during writing-plans so the plan author re-slices tasks to match the file
graph — rejected because a plan's slicing serves the reader and the reviewer, and re-slicing it to
match the file graph would make plans harder to read to save the dispatcher a computation it can do
itself; adding it to `## lint` as well — rejected because a lint run has no change in flight and
would print bundles for whatever happens to be open.

### Serialization and re-run scoping stay unscripted

**ID:** behavior-rules-unscripted
**Status:** active
**Chosen:** skill prose plus delta specs for rules 1 and 4.
**Considered:** a third guard auditing the SDD ledger for concurrent dispatches into one worktree —
rejected because the ledger carries no timestamps, so the audit would have no signal to read.

### Split the delta specs across a new capability and an existing one

**ID:** dispatch-economy-capability
**Status:** active
**Chosen:** new `myflow-dispatch-economy` for rules 1 and 2; extend
`myflow-review-panel-economics` for rules 3 and 4.
**Considered:** one new capability carrying all four — simpler to find, but splits panel rules
across two capabilities; folding all four into `myflow-review-panel-economics` — serialization and
bundling are implementer-dispatch concerns, not panel ones, so that is the wrong home.

### Raise the contract-budget row rather than offsetting the growth

**ID:** budget-raised-deliberately
**Status:** active
**Chosen:** re-anchor `skills/myflow-do/SKILL.md`'s row from its post-change size plus 25%, as the
guard's own convention requires, as the last content task.
**Considered:** cutting elsewhere in the file to stay under 54011 — risks deleting content to
satisfy a ratchet, which is precisely what the guard's own header warns a target-first budget
causes.

## Open questions

(none)
