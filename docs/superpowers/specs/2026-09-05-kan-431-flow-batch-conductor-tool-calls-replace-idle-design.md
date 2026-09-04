# Design — kan-431-flow-batch-conductor-tool-calls-replace-idle

## Context

kan-404's conductor transcript made 191 API calls at about $0.047 each — 29 of them idle polls
(`echo "still waiting"` every one to two seconds) and most of the rest one guard, one record or one
mark per turn (`docs/superpowers/research/flow-speedup.md`, section 6 round three). The polls have a
contract cause, not a habit: **4. Execute** (`skills/flow/implement.md`) says "never end a turn with a
child in flight", and **Dispatch the conductor** closes a turn that ends without `## Question`,
`## Stage` or `## Handoff` as `aborted`, so the conductor cannot yield to the harness's completion
notification and fills the gap with tool calls. Nothing forbids batching; one sentence assumes it
does not happen.

## The wait

Every wait on a child is one foreground Bash call, re-issued while the condition is still false:

```bash verified:a foreground sleep and a bounded loop ran in this harness's Bash tool while planning this change
for i in $(seq 1 110); do test -s <report> && break; sleep 5; done; test -s <report> && echo ready || echo still-running
```

- A foreground `sleep` works in the Claude Code Bash tool (checked by running one); one call is
  capped at 10 minutes, so the loop is bounded under it and the call names its outcome. A
  <!-- measured: the Bash tool's own description, "timeout is in milliseconds: default 120000, max 600000", and review-panel.md:330 @ 41688f0 -->
  `still-running` re-issues the wait; the panel's 15-minute ceiling (**No forking, and a wall-clock
  ceiling on every slot**, `skills/flow/review-panel.md`) is tracked across the calls.
- **One condition for every child kind:** `test -s <report path>`. Panel slots already write
  `<abs-worktree>/.superpowers/sdd/panel-report-<round>-<id>.md`. The implementer, the fix
  subagent and the verifier gain the same REPORT FILE paragraph — written as the child's last act,
  after its commit and its final test run:
  - implementer: `<abs-worktree>/.superpowers/sdd/implementer-report-<k>.md` (`<k>` the bundle;
    a resumed fix under `implementer-report-<k>-fix-<n>.md`)
  - fix subagent: `<abs-worktree>/.superpowers/sdd/panel-fix-report-<round>.md`
  - verifier: `<abs-worktree>/.superpowers/sdd/verify-report-<key>.md` (`<key>` the dispatch key,
    `verify` or `visual-verify`, suffixed as the key is)

## The batches

- **Task boundary** (`implement.md` §4, bundle N+1 just committed): one Bash — `record dispatch end`
  for the implementer, `check-task-commit-fields.sh` per new sha, `gather-dispatch-context.sh` for
  bundle N+2 and its `test -f`; one message launching bundle N's re-reviewer (when any), bundle N+1's
  reviewers and bundle N+2's implementer; one Bash with every `record begin` — the very next action
  after the launches return, which satisfies "recorded immediately after the launch returns, before
  any other action"; one wait on the reviewer's report; one Bash with `record end` and `flow tasks
  tick`. A guard failure still sends the task back before any launch — every verdict is printed,
  and only the dispatch depends on them.
- **Panel pre-work** (`review-panel.md`): base movement, citation pre-check, bundle rebuild,
  relocation comparison, diff-size, docs-only, `final-review.diff`, the throwaway worktrees and the
  `[PRINCIPLES_PATH]`/`[STANDARDS_PATHS]` resolution in one Bash; every slot launched in one
  message; `record begin`s in one Bash; the wait is one call for the whole round (`test -s` on
  every report), then the `record end`s in one; every `record finding` in one; every
  `run-reproducer.sh` (`; echo "F<n>: exit $?"` after each so each exit stays visible) and every
  worktree removal in one; the fix launch and its `record begin`; on its return the reproducer
  re-runs and the fix diff in one; every `record status … fixed` in one after the diff walk;
  `check-panel-findings-closed.sh` with the stage `end` mark.
- **Verify and the tail** (`verify-and-handoff.md`): `prepare-workspace.sh`, both `project-get.sh`
  calls and the stage `begin` in one; verifier launch; wait, `record end` and the ledger render in
  one; each later stage's `begin` rides its first command and its `end` its last.
- **Stage marks**: `flow stage begin` rides the stage's first command and `flow stage end` its last,
  in the same Bash call. The `## Stage` returns are unchanged — the tail relay is the research note's
  cost lever 4, a separate item.

## Where it is stated

One **Turn discipline** paragraph in **4. Execute** (`skills/flow/implement.md`), beside "Never end
a turn with a child in flight": independent calls share one Bash call and independent launches one
message; a wait is the loop above; the report-file condition and the three new paths; marks ride
adjacent work. `skills/flow/review-panel.md` and `skills/flow/verify-and-handoff.md` cite it once
and state only their own batches, per this repository's non-repetition rule.

`review-panel.md`'s sentence "Record it per finding, as its verdict is reached, not batched at the
round's end" becomes "recorded in the same call as every other verdict this turn reached, never
deferred to the round's end" — its purpose, that an aborted round leaves verified findings closed,
survives one call recording what the turn established.

## What does not change

`scripts/gather-dispatch-context.sh`, `scripts/check-task-commit-fields.sh` and
`scripts/run-reproducer.sh`: they are what the batched calls chain, and each prints its own
verdict. `scripts/check-dispatch-paragraphs.sh` (paragraph presence) and
`scripts/check-stage-mark-calls.sh` (flag literals) are unaffected. `skills/flow/review-panel.md`
sits 6 bytes under its `scripts/check-contract-budget.sh` row; the row is raised, the guard's own
documented response to a genuine addition.

## Decisions

### The wait is a bounded foreground loop, not a yielded turn

**ID:** foreground-bounded-until
**Status:** active
**Chosen:** one foreground `for … sleep 5` loop per wait, bounded under the Bash tool's 10-minute
cap and re-issued on `still-running` — one call per wait, the relay contract untouched.
**Considered:** ending the turn and letting the harness's completion notification re-invoke the
conductor — the parent closes a blockless turn as `aborted` (**Dispatch the conductor**), so the
relay contract and the parent would both change; an unbounded `until` — killed silently at the cap
with no outcome word.

### Every child kind is awaited on a report file

**ID:** report-file-wait-condition
**Status:** active
**Chosen:** `test -s <report>` for implementer, fix subagent, reviewer, slot and verifier alike,
with the three new REPORT FILE paragraphs.
**Considered:** the research note's "new commit" condition for implementers
(`git rev-list --count` reaching the bundle's task count) — fires before the implementer's final
test run and report, so the next implementer could be launched against a worktree the previous one
is still using, against **4. Execute**'s one-implementer-in-flight rule.

### Batching is stated once, in implement.md

**ID:** turn-discipline-in-implement
**Status:** active
**Chosen:** one paragraph in **4. Execute**, cited from the other two phase files.
**Considered:** restating the loop and the rule in each of the three files — three copies of one
rule, the drift this repository's non-repetition rule exists to prevent; a new
`skills/flow-contracts/` file — a fourth read for one paragraph.

## Open questions

None.
