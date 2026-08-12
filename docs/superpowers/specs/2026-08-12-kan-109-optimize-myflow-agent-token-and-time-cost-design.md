# KAN-109 — Optimize myflow agent token and time cost

**Date:** 2026-08-12
**Jira:** KAN-109 — "Optimize myflow agent token and time cost"

## Why

KAN-109 records a measurement taken on the KAN-71 fix-1 session: ~5.28M subagent tokens for one
change, of which 14 implementer dispatches accounted for 3.05M (58%) and two full seven-slot review
panel passes accounted for a further 1.70M. The waste the measurement isolated was not the work
itself but four structural properties of how `/myflow-do` dispatches:

- **Parallel implementers on one build directory.** Six assertions were left red at file seams and
  repaired by hand; agents idled 60s waiting on another agent's mid-edit compile; multiple Kotlin
  daemon sessions corrupted test-result XML, making whole suite runs untrustworthy and forcing
  re-runs.
- **Task slicing finer than the file graph.** Task 5.1 alone broke three assertions in files it was
  not allowed to touch. The bundles that were dispatched together (5.2+5.3+5.4, 6.1+6.2, 7.1+7.2)
  produced zero seam breakage.
- **Panel cost scaling with diff size, paid twice.** A 9877-line diff was read by two full
  seven-slot passes.
- **Conditional slots re-run against unchanged subjects.** The Security slot cost 76k tokens on
  pass 2 to confirm a comment-only backend diff.

A fifth proposal in the ticket — a stronger model for `models.panelFix` — is already in force:
**Model policy** (`skills/myflow-contracts/pipeline.md`) already defaults that role to Opus, or the
harness's strongest available model, deliberately and not to Sonnet. Nothing changes for it.

## Scope

Four rules. Two are mechanically checkable and get guard scripts; two govern agent behavior and
live in the skill and the delta specs, because there is nothing on disk for a script to measure.

Everything below applies to `/myflow-do`, and therefore to `/myflow-fast`'s implementation branch,
which runs `/myflow-do`'s sections by citation.

## 1. One implementer per worktree, serialized

`skills/myflow-do/SKILL.md` section 4 gains a rule: **at most one implementer subagent is in flight
against a given worktree at any moment.** Bundles are dispatched sequentially — the parent waits for
a dispatch to report its commit sha, runs the commit-fields guard, and dispatches review before the
next bundle starts.

Parallelism across *different* worktrees remains legal and untouched: a change affecting three
repositories still dispatches into three worktrees at once. The rule bans two dispatches sharing one
build directory, which is what produced the daemon collisions and the idle waits.

This **explicitly overrides** `superpowers:subagent-driven-development`'s parallel dispatch guidance
and `superpowers:dispatching-parallel-agents` wherever the tasks in question share a worktree,
joining the two model-policy overrides myflow already carries against the same upstream skill.

Nothing about this changes the review shape, the commit-per-task model, or the handoff bar.

## 2. Bundle dispatches by shared files

The dispatcher stops treating one `tasks.md` checkbox as one dispatch. Instead it groups tasks whose
declared `**Files:**` sets intersect, and dispatches one implementer per group.

**`scripts/plan-dispatch-bundles.sh`** is a thin Bash wrapper over
**`scripts/plan-dispatch-bundles.py`**, following the split `check-task-build-green.sh` and
`check-task-commit-fields.sh` already use for `tasks.md` parsing, and for the same reason: block
structure in Markdown deserves a real parser rather than a hand-rolled Bash ERE allowlist.

**Input:** a `tasks.md` path. With no argument, the wrapper resolves every non-archived change's
`tasks.md` under `openspec/changes/*/tasks.md`, exactly as `check-task-build-green.sh` does.

**Algorithm:** read each **unchecked** task's dotted id and its `**Files:**` field. Two tasks
sharing any declared path join the same bundle; the relation is transitive, so the grouping is the
connected components of the task-file bipartite graph, computed with union-find.
An `**Allowed-collateral:**` glob is *not* part of the join — it describes paths a sweep may touch,
not paths the task owns, and joining on it would collapse unrelated tasks into one bundle.

**Output:** one line per bundle on stdout, ordered by the bundle's lowest task id:

```
bundle 1: 1.1 1.2
bundle 2: 2.1
```

**Exit codes:** 0 bundles computed; 1 a task carries no `**Files:**` field (reported by task id — the
field is already required by `myflow-task-commit-fields`, so its absence is a plan defect, not a
reason to guess); 2 the guard cannot answer at all — unreadable file, no working `python3`.

**Where it runs:** `/myflow-do` section 4 only, at dispatch time. It is not a lint step: it computes
a grouping rather than judging a file's text, and a lint run has no change in flight to compute one
for. It is covered by its own harness under `## test`.

**Interaction with red tasks.** A task tagged `Build: red` carries `**Squash-with:** Task <N>`. When
both land in one bundle — the common case, since a red task and its green partner touch the same
files by construction — the bundle's implementer still makes one commit per task and folds the red
one into its partner exactly as section 4 already requires. When they land in different bundles, the
existing fixup-and-autosquash mechanism is unchanged; bundling does not move that rule.

**Interaction with plan order.** Bundles are dispatched in the order of their lowest task id, so a
plan whose tasks are already ordered by dependency stays ordered. A dependency that crosses a bundle
boundary is the plan's own concern and is unchanged by this rule.

## 3. Panel diff-size cap

**`scripts/check-panel-diff-size.sh <worktree> <merge-base> [cap]`** — pure Bash over git, in the
shape of `check-finish-preflight.sh`, since it needs no Markdown parsing at all.

It counts changed lines across the branch: `git -C <worktree> diff --shortstat <merge-base>` for
committed work plus the unstaged working tree, summing insertions and deletions — the same body of
text `final-review.diff` will carry.

**Default cap: 2000 changed lines.** KAN-71 measured 948k tokens for seven slots at ~9877 lines;
2000 keeps a full pass in the low hundreds of thousands and leaves ordinary myflow changes
untouched. The cap is a positional third argument so a project can raise or lower it without editing
the guard.

**Exit codes:** 0 at or under the cap; 1 over the cap; 2 cannot answer — not a worktree, unresolvable
merge base, a non-numeric cap.

**Where it runs:** `/myflow-do` section 5, immediately before `final-review.diff` is written. On
exit 1 the operator gets a named-options prompt, shaped per **Operator prompts**
(`skills/myflow-contracts/operator-prompts.md`):

> **The branch diff is `<N>` changed lines, over the `<cap>`-line panel cap.**
> - **Proceed with the panel anyway** *(default, recommended)* — the panel reads the whole diff
> - **Stop — I'll split the change** — the run stops at `IN_PROGRESS` with nothing lost

The measured line count, the cap, and the operator's answer are all recorded in
`.superpowers/sdd/final-review-panel.md` either way, including on an exit-0 run, so the panel record
always states the size the panel read.

**The cap never moves the handoff bar.** Proceeding past it changes nothing about the panel's
roster, its slots, or the zero-open-findings requirement. Stopping is a stop, not a pass: the run
ends at `IN_PROGRESS` exactly as a fix round does, and the operator re-runs after splitting.

Exit 2 is a stop for the same reason a workspace-isolation exit 2 is: a size the guard could not
measure is not a size under the cap.

## 4. Conditional slots re-run only when their subject changed

Section 5's "Panel re-runs" table has two modes. **Targeted** is already scoped and is unchanged.
**Full** — the escalation mode — currently re-runs every slot in the run's roster, which is what
made the Security slot spend 76k tokens confirming a comment-only diff.

New rule, in Full mode only:

- **Required slots always re-run.** Primary, Principles, and the roster's third required slot re-read
  the rewritten `final-review.diff` on every full pass. Nothing about them is scoped.
- **A conditional slot re-runs only when its own trigger-table row still fires against
  `fix-round-N.diff`.** Security re-runs when the fix diff touches auth, tokens, crypto, secrets or
  config, query construction, path handling, deserialization, the CORS/HTTP edge, or dependencies.
  Adversarial re-runs when the fix diff touches migrations, concurrency, tested behavior, a modified
  test, or exceeds its own line threshold. Lens B and Lens C re-run on their own rows likewise.
- **A conditional slot whose trigger did not fire is not re-run**, and the panel record says
  `not re-run — subject unchanged` beside its pass-1 result.

**This is a definition of non-stale, not a waiver.** The existing bar requires the final pass to show
a non-stale clean result for every slot in the roster. A slot's result is stale when the diff it read
has since changed *in the region that slot reads*. A conditional slot's region is exactly its
trigger's subject; a fix that touches nothing in that subject leaves the slot's reading current. The
required slots have no such bounded region — they read the whole diff — which is why the scoping
applies to conditional slots alone.

The zero-open-findings bar is untouched. A conditional slot that raised an open finding on pass 1
still blocks the handoff whether or not it re-runs, and a slot that is not re-run cannot close its
own finding by not looking again.

## Delta specs

- **New capability `myflow-dispatch-economy`** — rules 1 and 2. Serialization is a dispatch
  property; bundling is a dispatch property. Neither belongs to the review panel.
- **Extended `myflow-review-panel-economics`** — rules 3 and 4. Both are properties of what the panel
  reads and how often, which is what that capability already governs.

## Collateral changes

- **`scripts/check-contract-budget.sh`** — the `skills/myflow-do/SKILL.md` row rises from 54011
  bytes to the file's post-change size plus 25%, per the guard's own stated convention. Raising the
  row is the correct response to a genuine addition; narrowing the guard is not.
- **`.myflow/project.md`** — `## test` gains `scripts/test-plan-dispatch-bundles.sh` and
  `scripts/test-check-panel-diff-size.sh`.
- **Neither new guard joins `## lint`.** Both need a change in flight and a real worktree or plan
  path passed in, which is precisely why `check-finish-preflight.sh`, `check-unfinished-work.sh`,
  `preserve-session-records.sh` and `check-cleanup-complete.sh` are already excluded. A lint step
  that cannot run against a bare tree fails on every unrelated invocation.

## Testing

Two new harnesses in the repository's existing style — `scripts/test-plan-dispatch-bundles.sh` and
`scripts/test-check-panel-diff-size.sh` — each building fixture trees under a temporary directory,
one assertion per case, and covering every exit code the script declares.

`plan-dispatch-bundles`: disjoint tasks produce one bundle each; two tasks sharing a file produce one
bundle; transitivity across three tasks produces one bundle; checked tasks are excluded; a task with
no `**Files:**` field exits 1 and names the task; `**Allowed-collateral:**` does not join bundles; an
unreadable path exits 2.

`check-panel-diff-size`: a diff under the cap exits 0; a diff over it exits 1 and prints the count;
an explicit cap argument overrides the default; unstaged work counts toward the total; a non-numeric
cap exits 2; a path that is not a worktree exits 2.

## Out of scope

- No change to `models.panelFix`, which already defaults to the strongest available model.
- No change to the review panel roster presets, the optional-slot trigger table itself, the
  escalation ladder's trigger conditions, the panel record format, or the marker-line rules.
- No change to `/myflow-start`'s planning stages or to task slicing at plan-writing time. Bundling
  happens at dispatch, deliberately: the plan's slicing serves the reader and the reviewer, and
  re-slicing it to match the file graph would make plans harder to read to save the dispatcher a
  computation it can do itself.
