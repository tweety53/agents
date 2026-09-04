# kan-402-flow-one-combined-panel-diff-per-round-and

## Context

`proposal.md` carries the diagnosis: the per-worktree dispatch rule in `skills/flow/review-panel.md`
doubles every slot on a two-repository change, and the implementer bundle's plan section is the
whole `tasks.md` rather than the dispatched task's block. No capability spec in `spectre/specs/`
governs either behaviour; the three files below are the whole surface.

## Finding 1 — one dispatch per slot per round

All of this is prose in `skills/flow/review-panel.md`. No script changes.

### The combined diff

`<canonical-worktree>/.superpowers/sdd/final-review.diff` is written once per round from every
worktree in the change's resolved set (**Resolving a change's worktrees**,
`skills/flow-contracts/worktree-resolution.md`), in resolved order:

```sh unverified:confirm the header line survives `git apply --stat`-style readers a slot might run on the file; it is a comment to a human reader, never applied
: > <canonical-worktree>/.superpowers/sdd/final-review.diff
for each <worktree> in the resolved set:
  printf '# worktree: %s — merge base %s\n' "<abs-worktree>" "<merge-base>" >> final-review.diff
  git -C <worktree> diff <merge-base> >> final-review.diff
```

`<merge-base>` is each worktree's own working-notes merge base, exactly the value the per-worktree
`check-base-moved.sh` call already reads. A single-worktree change writes the same shape with one
header: one rule, no special case, and every slot's prompt can always say "the diff is sectioned by
worktree".

### One dispatch per slot

Each included slot is dispatched once per round, in the canonical worktree, key
`panel-<round>-<slot>` — the key `review-panel.md` already states; KAN-29's `-backend`/`-frontend`
suffix was the operator's improvisation around a rule that forced two dispatches. `flow record
dispatch begin -diff-base` takes one sha; a delta dispatch records the canonical worktree's held
sha, and the panel record (`final-review-panel.md`) names every worktree's sha beside the diff path.

### Deltas on re-runs

A slot's last-reviewed sha becomes **per slot per worktree**: each dispatch sets that slot's sha in
every worktree to the HEAD it was dispatched against. `slot-delta-<round>-<slot>.diff` is combined
the same way as `final-review.diff`, one section per worktree, from that worktree's held sha. A slot
whose delta is empty in **every** worktree is not dispatched — the existing `not re-run — nothing
new since its last read` record. A worktree with no held sha for that slot contributes its whole
`git diff <merge-base>` section, under the existing no-held-sha rule.

### Cap and docs-only across worktrees

`check-panel-diff-size.sh <worktree> <sha>` runs once per worktree (per distinct held sha on a
re-run, as today), unchanged. **The gating count is the sum across worktrees**, since one slot now
reads all of them; an over-cap sum puts the existing over-cap prompt, naming the sum and the
per-worktree counts. `check-panel-docs-only.sh` runs once per worktree; the reduction to `primary`
fires only when **every** worktree exits 0, and the first non-documentation path any worktree
printed is the one recorded.

### Bugbot and Security

Bugbot's throwaway-worktree snippet runs once per worktree, producing `<worktree>-bugbot-<round>`
for each; **one** Bugbot dispatch lists every copy, and the removal loop runs over all of them.
Security is **one** dispatch naming every worktree path. Neither reads the diff file; both are
told, as every slot is, that the change spans the listed worktrees.

### Findings across worktrees

When the resolved set holds more than one worktree, every slot's prompt carries: a finding's
`file:line` is prefixed with the worktree's basename (`gymie-frontend:src/Foo.tsx:42`), and its
reproducer runs from that worktree. With one worktree, nothing changes: `file:line` stays
repo-relative. The citation pre-check and the relocation comparison stay per worktree — each writes
its own file under that worktree's `.superpowers/sdd/` — and the prompt names every such file that
exists.

## Finding 2 — the plan section scoped to the dispatched task(s)

### `gather-dispatch-context.sh`

Usage becomes `gather-dispatch-context.sh <worktree> <change-root> <name> <principles-path>
<output-path> [<task-ids>]`. `<task-ids>` is a comma-separated list of flat integers. When absent,
output is byte-identical to today. When present, the `## tasks.md` section is:

1. the plan header — every line before the first task line;
2. each named task's block, in **document order** regardless of the order the ids were given —
   from its task line to the next task line, the next line matching `^#{2,3}(\s|$)`, or end of
   file.

A task line is `^- \[[ x]\] (\d+)\. ` outside a fenced block, and fences (three or more backticks or
tildes, toggling) are opaque — the grammar `scripts/plan-dispatch-bundles.py` states, so the ids the
planner printed are the ids this script finds. The extraction is one `awk` pass in the script; it
reads the plan once and never re-parses a task's fields. A named id with no task line in the plan
exits 2 with `task <id> not found in tasks.md` — an implementer with no task text is worse than a
stopped run. The scoped section enters `render_body` in place of the full file, so the hash sidecar
and the skip-when-unchanged path work per output path with no change.

### `implement.md`

The single pre-dispatch gather becomes one gather per dispatch bundle, immediately before that
bundle's implementer:

```sh unverified:confirm `plan-dispatch-bundles.sh` prints `bundle <k>: <ids>` with space-separated ids, so the `tr` below yields the comma list
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  <worktree>/.superpowers/sdd/dispatch-context-bundle-<k>.md <ids-comma-separated>
```

`<k>` and the ids are the `bundle <k>: <ids>` line the existing `plan-dispatch-bundles.sh` call
printed. The CONTEXT BUNDLE paragraph names `dispatch-context-bundle-<k>.md` and states that its
plan section carries the plan header and this bundle's tasks only. The fix-subagent and panel
gathers in `review-panel.md` keep the five-argument call and the `dispatch-context.md` path.

### Tests

`scripts/test-gather-dispatch-context.sh` gains cases, appended after the existing ones: scoped to
one id yields the header plus that block only; two ids given in reverse order come out in document
order; a `- [ ] 9. ` line inside a fence is not a task and a request for id 9 exits 2; an unknown
id exits 2 with the message above; the five-argument call's output is byte-identical to the
scoped call's non-plan sections.

## Decisions

### One combined diff per round rather than one per worktree

**ID:** combined-diff-per-round
**Status:** active
**Chosen:** one `final-review.diff` sectioned by worktree, one dispatch per slot per round — halves
the panel on a two-repository change and puts cross-repository seams in one reviewer's view.
**Considered:** *keep per-worktree dispatch* — rejected: it is the doubling KAN-402 reports, and it
hides exactly the seams a multi-repository change is most likely to get wrong.

### Bugbot and Security dispatched once, told every worktree

**ID:** bugbot-security-one-dispatch
**Status:** active
**Chosen:** one throwaway worktree per repository, one Bugbot dispatch listing them all; one
Security dispatch listing every worktree path. KAN-29's eight dispatches become four.
**Considered:** *combine only the diff-reading slots, keep Bugbot and Security per worktree* —
rejected by the operator: it saves two of eight instead of four, and neither slot needs a separate
prompt to run one repository's tests when the prompt can name both.

### `-diff-base` carries the canonical worktree's sha

**ID:** diff-base-canonical-sha
**Status:** active
**Chosen:** the record CLI's single `-diff-base` takes the canonical worktree's held sha; every
worktree's sha is written in the panel record.
**Considered:** *extend `flow record dispatch begin` to take several shas* — rejected: a Go and
schema change to carry a value the panel record already holds, for a field the ledger renders as one
line.

### The cap gates on the sum across worktrees

**ID:** cap-sum-across-worktrees
**Status:** active
**Chosen:** `check-panel-diff-size.sh` unchanged, run per worktree; the parent sums.
**Considered:** *teach the guard to take several worktree/merge-base pairs* — rejected: the sum is
one line of arithmetic and the guard's existing tests and cap rationale stay untouched.

### Scope the plan section, not the change's files

**ID:** scope-tasks-not-files
**Status:** active
**Chosen:** the bundle's `## tasks.md` section is scoped to the dispatched task ids; nothing else
in the bundle changes.
**Considered:** *scope by `**Files:**` as the issue says* — rejected: the bundle carries no source
files, so there is nothing to scope. *Drop finding 2 until per-dispatch cost is attributable
(KAN-401/KAN-414)* — rejected by the operator: the plan-section cut is measurable now (156 KB of a
210 KB bundle on KAN-29) and costs one optional argument.

### An unknown task id is exit 2

**ID:** unknown-task-id-exit-2
**Status:** active
**Chosen:** a named id with no task line stops the gather with exit 2.
**Considered:** *report it as a skipped source and exit 0, like a missing `design.md`* — rejected:
a missing design is a legitimate absence; a missing task means the caller passed the wrong ids and
the implementer would be dispatched with no task text.

### One output path per dispatch bundle

**ID:** per-bundle-output-path
**Status:** active
**Chosen:** `dispatch-context-bundle-<k>.md`, `k` the bundle number `plan-dispatch-bundles.sh`
printed; the full-plan `dispatch-context.md` stays for the panel and fix subagents.
**Considered:** *overwrite `dispatch-context.md` per bundle* — rejected: a reviewer dispatched
beside the next implementer reads the same path, and the skip-when-unchanged sidecar would rebuild
on every alternation.

## Open questions

None — the two forks were answered during brainstorming and are recorded above.
