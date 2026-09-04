# Design — kan-404-flow-keep-the-mutation-testing-brief-as-a

## Context

KAN-29's self-review credits the panel's highest-severity findings (F11, F24–F26) to the
general-purpose reviewer that stood in for Bugbot when the harness offered no `bugbot` agent
type — `skills/flow/review-panel.md`'s **An unspawnable id is substituted, not skipped** dispatched
it carrying Bugbot's mutation-testing brief, and sabotage-proofing (moving a click off its target,
overlaying an earlier commit's tree) caught what green runs hid. That reviewer exists today only as
a fallback: `ValidReviewers` (`stats/internal/store/settings.go`) carries five ids, and the mutation
brief reaches the panel only through `bugbot`'s own substitution path. An operator who wants it on
purpose — on a harness that does offer Bugbot, or beside it — has no id to name in
`/flow-settings`. This change adds `mutation` as a sixth, opt-in `ValidReviewers` id, and factors
the brief and its throwaway-worktree sequence into sections `bugbot` and `mutation` both bind.

## The slot

`mutation` is the sixth entry of `ValidReviewers`. Its roster row:

| id | Slot | How to spawn | Model |
|---|------|---------------|-------|
| `mutation` | **Mutation** — sabotage-proofing | general-purpose + the mutation-testing brief, own throwaway worktree copy per repository, one dispatch per round naming every copy | `DEFAULT_MODEL` |

It is dispatched by the parent like Principles and Code review (low): a general-purpose subagent
given a prompt, so `-model` records `DEFAULT_MODEL` (or this run's override), never
`unknown (agent-defined)`. `-slot` records `Mutation`. It reads no diff file — it mutates and runs
the tree — so on a fix round it follows the rule Bugbot and Security follow: re-run only when it
raised a finding in the previous round or the previous round raised a new Critical.

## What `bugbot` and `mutation` share

The sections "Bugbot's mutation-testing brief" and "Bugbot's throwaway worktree" in
`skills/flow/review-panel.md` become "The mutation-testing brief" and "The throwaway worktree",
each opening with the two slot ids it binds. The worktree copy is `<worktree>-<slot>-<round>`
(`<worktree>-bugbot-3`, `<worktree>-mutation-3`), so a roster carrying both slots produces two
distinct copies per repository per round, each removed unconditionally when its own dispatch
closes. `artifacts-registry.md`'s row for the copy takes the same name.

The brief's list of mutations grows from four to six: flip a condition, drop a guard, move a
boundary, remove a branch, move an interaction off its target, overlay an earlier commit's tree and
run the tests. Both slots carry the six.

## What does not change

`bugbot` keeps its own agent type, its `unknown (agent-defined)` model record, and its substitution
rule. `DefaultReviewers` stays `primary, principles, code-review-low`. Security stays unisolated.

## Decisions

### The mutation brief becomes its own reviewer id rather than replacing `bugbot`

**ID:** mutation-is-a-sixth-id
**Status:** active
**Chosen:** add `mutation` to `ValidReviewers`; `bugbot` unchanged — an operator can name the
general-purpose mutation reviewer deliberately, on any harness, beside or instead of Bugbot.
**Considered:** repurpose `bugbot` to always dispatch general-purpose + brief — smallest diff, but
the real Bugbot agent is never used again where it exists and the id name no longer says what runs.
Add `mutation` to `DefaultReviewers` as well — one more dispatch per round on every run in every
project; the operator chose opt-in.

### The brief grows the two sabotage shapes KAN-29 credits

**ID:** brief-gains-two-shapes
**Status:** active
**Chosen:** add "move an interaction off its target" and "overlay an earlier commit's tree and run
the tests" to the brief's list, for both slots — the slot exists because those shapes found what
green runs hid, so it is briefed to do them.
**Considered:** leave the brief as written, since the shapes are instances of "for each behaviour
the diff changes, mutate it" — rejected because a brief that names four shapes and not these two
steers the reviewer toward the four.

## Open questions

None.
