# kan-513-flow-bias-free-implementer-grouping-toward

Source: KAN-513

## Context

Decide step 4's free implementer grouping defaults close to 1:1 with `plan-dispatch-bundles.sh`'s
raw file-overlap bundles even when several bundles have no real parallel-execution need — bundles
are serial by default, and the two-in-flight-per-wave cap already bounds useful splitting, so a
mechanical default biased toward fewer, larger groups closes the gap the ticket reports. See
**Decisions** below for what changes and why; the analysis in the sections that follow.

## 1. Where the 1:1 grouping comes from

Step 4 of **Decide** (`skills/flow/brainstorm-planner.md`) says: run `plan-dispatch-bundles.sh`,
"group its bundles freely (no roll, no static table, no per-group ceiling; at most two implementer
dispatches in flight per wave)". Nothing mechanical proposes a grouping, so the default is wherever
the planning model's instinct lands — one group per bundle, exactly the `bundle <k>:` lines it just
read. The merge the ticket describes today happens only through a session-instruction override to
the result.

The precedent for "mechanical default the planner may override with a recorded reason" is already
in the same section: `class_mechanical` → `class` with `override`. The same shape fits groups.

## 2. The mechanical grouping rule

A new sibling script, `scripts/plan-dispatch-groups.py` behind a `plan-dispatch-groups.sh`
wrapper, imports `compute_bundles` from `plan-dispatch-bundles.py` (same stdlib-only, same
`file:line:` violation format, same exit codes 0/1/2) and prints one line per group:

```text verified:scripts/plan-dispatch-groups.py's own output, confirmed by running it against this change's own tasks.md
group <g>: <bundle ids in plan order>
```

Groups are numbered from 1 in the order of their lowest bundle id. Bundle ids are
`plan-dispatch-bundles.sh`'s own `<k>`, so the recorded `groups` field keeps today's meaning.

Definitions, over bundles `B_k` with members `M_k` (task ids) and resolved after-set `A_k`:

- `deps(k) = A_k \ M_k` — what bundle k actually waits on. A bundle's own after-set already
  contains its own members (`scripts/test-plan-dispatch-bundles.sh` case at line 82:
  `bundle 1: 1 2 / after 1: 1`), so implement.md's readiness check has always had to treat
  in-bundle ids as satisfied; the same subtraction is what makes a merged group's readiness
  well-defined.
- `members(g)` — union of `M_k` over the group's bundles; `ready(g)` — union of `deps(k)` over
  them, minus `members(g)`. A group is ready when every id in `ready(g)` has landed.

Two passes, both in plan order:

1. **Chain merge.** Walk bundles by id. Bundle k joins the first existing group g such that
   `deps(k) ∩ members(g) ≠ ∅` and `deps(k) ⊆ members(g) ∪ ready(g)`; otherwise it opens a new
   group. The first condition means k genuinely depends on something in g (so it could never run
   concurrently with g anyway); the second means everything else k waits on had already landed
   before g could start. Joining therefore leaves `ready(g)` unchanged — the group's readiness
   never moves, and the group's bundles run in plan order inside one implementer session with every
   dependency satisfied in turn. A bundle whose deps straddle two groups (a join point) opens a new
   group whose `ready` is the union it waits on.
2. **Fold to ≤2 per ready-set.** Among groups with an identical `ready` set — the same readiness
   trigger, so they become ready at the same moment — more than two are folded down to exactly
   two, alternating in plan order (1st and 3rd together, 2nd and 4th together, …). Folding two
   groups with an equal `ready` set changes neither's readiness and loses no parallelism the
   two-slot cap could have used. Groups with merely the same wave depth but different `ready` sets
   are **not** folded: two such groups can become ready at different moments and folding would
   delay one. This pass runs on the chain-merged groups, so a chain hanging off a folded member
   stays attached to it.

Worked examples (bundle ids; `→` reads "after"):

- Serial-default plan, five bundles: one group `[1,2,3,4,5]`. The ticket's case.
- `1`, `2` (both `after: none`), `3 → 1`, `4 → 2`, `5 → 3 4`: chain merge gives `[1,3]`, `[2,4]`,
  `[5]`; nothing folds (the first two share `ready = ∅` and are already two).
- Four bundles all `after: none`, then `5 → 1 2 3 4`: chain merge gives four singletons plus `[5]`;
  fold turns the four into `[1,3]` and `[2,4]`; `[5]` stays, `ready = {1,2,3,4}`.
- `1`, `2 → 1`, `3 → 1`, `4 → 2 3`: one group `[1,2,3,4]`. Bundle 2 joins group 1 (`deps = {1}`);
  bundle 3 also has `deps = {1} ⊆ members`, so it joins the same group rather than opening a second;
  bundle 4's deps are then all inside. This is the rule's known ceiling: 2 and 3 are file-disjoint
  and mutually ready, and the merge serialises one bundle of work that could have overlapped, at
  the price of one saved dispatch. The alternative — `[1,2]` and `[3]`, with `[4]` waiting on both —
  costs two extra dispatches for one bundle of overlap. Where the overlap is worth it, the planner's
  split-only override (section 3) is exactly the tool.

## 3. The planner's freedom: split only

Step 4 becomes: run `plan-dispatch-groups.sh <changeRoot>/tasks.md`; its output is
`groups_mechanical`. The planner may **split** a mechanical group into more groups — a chain it
judges too long for one implementer's context, or the example-4 case above where a real parallel
wave is worth a dispatch — with a one-line `groups_override` reason, and never merge across the
mechanical result: merging two groups the script kept apart would collapse a real parallel wave.
`groups` is the result actually dispatched; `groups_override` is `null` when the planner took the
mechanical grouping verbatim; `groups_reason` stays and now defaults to the literal `mechanical`.
Session-instruction overrides keep their existing `overrides` slot and are not `groups_override`.

No plan-time group-size ceiling: one serial chain is one group however long. A ceiling was
considered (tasks per group, files per group, or a runtime handback on low context) and rejected —
the implementer commits per task, so an over-long group loses at most one task's work on a context
failure, and the split-only override covers a chain the planner can see is too long.

`skills/flow-fast/` is untouched: its decision records `groups: null` and it never dispatches.
`stats/` is untouched: the decisions view projects `groups` alone
(`stats/internal/store/aggregate_test.go`, `TestDecisionsRendersGrouping`), and the two new
sibling fields are ignored by it.

## 4. implement.md readiness wording

**Waves** (`skills/flow/implement.md`) says a group is ready "when every id in the union of its
bundles' `after <k>:` lines has landed". Read literally, a group containing bundle 1 and bundle 2,
whose `after 2:` names task 1, waits on itself. That has always been resolved by practice, never by
text; with chain-merged groups the case becomes the norm rather than the exception, so the sentence
gains the subtraction: ready when every id in that union **that is not itself a member of the
group** has landed.

## Decisions

### Mechanical default with split-only override for implementer groups

**ID:** mechanical-groups-split-only
**Status:** active
**Chosen:** A deterministic script (`plan-dispatch-groups.py`, chain-merge + fold-to-≤2) computes
the default grouping; the planner may only split it further, never merge across it — mirrors
`class_mechanical`/`override`.
**Considered:**
- Leaving grouping fully free (today's behavior) — rejected: this is the exact problem the ticket
  reports, an unbounded 1:1 default with no bias toward fewer dispatches.
- Letting the planner merge across the mechanical result too — rejected: a merge across it could
  collapse a real parallel wave the script deliberately kept split (worked example 4).
- A plan-time group-size ceiling (tasks/files per group) — rejected: the implementer commits per
  task, so an over-long group loses at most one task's work on a context failure; the split-only
  override already covers a chain the planner judges too long.

### implement.md readiness excludes a group's own members

**ID:** waves-readiness-self-exclusion
**Status:** active
**Chosen:** State explicitly that a group's readiness excludes ids that are members of the group
itself.
**Considered:** Leaving the wording as-is and relying on practice — rejected: merged groups make
the self-wait reading the norm rather than the exception, so the text must say what practice
already does.

### A fully-seeded research note skips brainstorming's interactive checklist entirely

**ID:** fully-seeded-note-skips-checklist
**Status:** active
**Chosen:** When a found staging note carries all three of: the note itself, a sibling
`<stem>/tasks.md` and a sibling `<stem>/decision.json` (**Seed from a staged research note, if one
exists**, `skills/flow/brainstorm-planner.md`), the interactive checklist and the merged
convergence-and-approval confirm are both skipped — present the note's parsed structure, then
continue directly into **C** (artifact creation). This is a second, narrow exception to **Stage
exit — never the command's own judgment**'s existing rule (`skills/flow-contracts/pipeline.md`),
alongside its "no channel to ask through" exception: `/flow-plan`'s own investigate-then-ask
process (at least two rounds per topic, an explicit convergence check) already gathered and
confirmed the content that produced the note, plan and decision, so there is no fresh operator
judgment left to gate on for a *fully* seeded note. A note missing either sibling file — a note
alone, or a note with only a plan — still runs the full checklist exactly as today; this exception
never applies to a partial seed. The `flow.brainstorm` end and `flow.design-approval` begin/end
marks still fire, back-to-back with no interactive gap between them, to keep stage bookkeeping
consistent with every other run.
**Considered:**
- Skipping the checklist whenever any note is found, seeded plan/decision or not — rejected: a bare
  note (no plan, no decision) answers less, and the checklist is exactly what surfaces what it does
  not cover; narrowing the exception to the fully-seeded case is what keeps it safe.
- Leaving today's "seeding never skips the interactive round" rule as the only path — rejected per
  the user's explicit request: a fully-seeded note (this change's own `kan-513` note being the
  concrete case) already re-litigates content the operator settled in the `/flow-plan` session that
  produced it, so the redundant confirm round is pure overhead.

## Open questions

None.
