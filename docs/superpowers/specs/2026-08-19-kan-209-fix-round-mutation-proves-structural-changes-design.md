# A fix round mutation-proves every behaviour it changes

**Change:** `kan-209-fix-round-mutation-proves-structural-changes`
**Jira:** KAN-209
**Date:** 2026-08-19

## The problem

A review panel's fix round is told which findings to repair. Where a finding asks for test cases,
the round proves those cases: it reverts the fix in a scratch tree, confirms the new case fails,
and restores. Where the round *also* changes structure — a delimiter, a protocol, a control flow —
alongside those cases, nothing asks it to prove that change, and it does not.

KAN-200's fix round 2 is the worked example. It repaired five findings. For G1 it added eight
harness cases and mutation-proved all eight, individually. For G3 it replaced the record protocol's
delimiter with an ASCII Unit Separator — the round's largest structural change — and proved
nothing. Reverting that protocol left all eighteen harness cases green, because the harness
contained no literal tab byte and no `\037` byte anywhere. The gap was found only in round 3, when
the Principles slot thought to ask whether the previous round had held itself to its own standard.

The asymmetry is not carelessness. It is the natural shape of "prove the thing you were told to
prove", and it recurs every time a fix round both adds tests and changes structure.

A second data point from the same run shows the discipline is load-bearing rather than ceremonial.
Fix round 3, told to mutation-prove the protocol, found that the combined revert its parent had
specified would produce a **false pass**: mutating the shared `US` value changes both the protocol
and the new control-byte check at once, so a case passes by cross-contamination rather than by the
protocol working. It split into two surgical mutations instead, and two round-4 slots independently
reproduced and confirmed that reasoning. A mutation that is not isolated proves nothing.

## The rule

**A fix round mutation-proves every executable behaviour it changes, not only the test cases it
adds.**

- **Scope is executable behaviour** — a guard script, Go, TypeScript, anything with a test that
  could fail. A prose-only or markdown-only fix round has nothing to mutate and records an explicit
  exemption rather than silence.
- **Each mutation alters one mechanism.** Where a single revert would touch state that more than one
  check reads, it splits into surgical mutations, one per mechanism. A mutation touching shared
  state can produce a false pass, which is worse than no mutation at all: it reports coverage that
  does not exist.
- **A surviving mutation is repaired inside the round.** Where no existing test fails, the round
  adds the test that catches it before it closes. It is not raised as a new finding and costs no
  extra panel pass — the round already has the change in hand, and a finding would spend a full
  round re-discovering what the round already knows.
- **The parent runs the mutations, after the fix subagent reports.** The subagent names the
  behaviours its fix changed; the parent mutates each and confirms the failure.

## Why the parent, not the subagent

`skills/myflow-do/SKILL.md` already establishes this shape one paragraph earlier: "Once the fix
subagent reports, re-run every dispatched finding's reproducer" — the parent re-runs, under the same
constraints, rather than accepting the subagent's account of its own success. The same reasoning
applies here and applies more strongly. The failure this change exists to prevent is a round
assessing its own work and finding it complete; letting the round self-certify its mutations
reproduces that failure one level down.

The cost is bounded. The subagent has already named the changed behaviours, so the parent's work is
a revert, a test run, and a restore per behaviour — not a re-derivation of what changed.

## The record

The parent writes, into the pass log entry it already keeps in
`.superpowers/sdd/final-review-panel.md` alongside the mode, the slots that ran, the diff path and
the bounced findings:

```
fix-mutation: <path> — <what was mutated> — <the test that failed>
fix-mutation: <path> — none — <reason>
fix-mutations-total: <n>
```

One line per executable behaviour the round changed. The exemption reuses the `none — <reason>`
literal form the record already uses for `finding-reproducer:`, so a reader meets one exemption
shape rather than two.

**These lines sit in the pass log and never inside the marker block.** This is a hard constraint,
not a stylistic preference. `check-unfinished-work.sh` requires every `finding-status:` marker to
occupy one unbroken run of consecutive lines — a rule added precisely so that a marker written
elsewhere cannot stand in for a missing one — and it derives the findings table's identifiers from
lines matching `^\|?[[:space:]]*F[0-9]+[[:space:]]*\|`. A new line that split the marker block, or
that resembled a table row, would break a guard this change does not otherwise touch. The
regression is cheap to test and expensive to discover late, so it is tested.

## Why no guard reads these lines

KAN-197 examined the general form of this question — "require a mutation test for every guard
change" — and rejected two mechanisms for it: a per-harness failure fixture, which every harness
already satisfied while remaining compatible with the defect it was meant to catch, and a
`Mutation:` field on every task in `tasks.md`, which puts per-task friction on every future change
to catch a defect living in the guard rather than the task.

A guard over `fix-mutation:` lines would face the same problem from the other side. To be more than
a nag it would have to decide, from a diff, whether a fix round changed executable behaviour or
edited a comment — a classification the guard cannot make, and one that is wrong in the direction
that matters: a guard that cannot tell will either fire on every prose fix round in a repository
that is mostly prose, or be silenced into vacuity.

What holds the rule instead is the record plus the bar below. The record is what makes the question
askable at all: KAN-200's gap was found because a reviewer thought to ask, and a recorded field
turns that from a lucky question into a line the next round's Principles slot reads.

## The bar

The parent does not close the fix round — and so does not reach the handoff — while any executable
behaviour the round changed carries neither a `fix-mutation:` line nor an exemption. This is the
round's own completeness condition, checked where the parent writes the record, not a new class of
finding and not a new gate in the handoff sequence.

## Scope

This binds a **fix round's** own obligations and nothing else. It does not reach task
implementation, the red-task-partner's fixups, or guard changes generally — KAN-197 examined that
broader form and rejected it.

It binds **every roster**, `light` included. The obligation belongs to the round, not to a slot, so
a preset that dispatches no Bugbot — and therefore carries no reasoned-mutation reading of the
diff — is exactly the case where the round's own proof is the only mutation reasoning that happens
at all. It adds no slot to any preset and changes no preset's required set.

## Affected surfaces

- `skills/myflow-do/SKILL.md` — the rule, in section 5, after the reproducer re-run paragraph.
- `openspec/specs/myflow-review-panel-economics/` — one new requirement, sibling to "Bugbot's brief
  includes reasoned mutation testing".
- `scripts/test-check-unfinished-work.sh` — a case proving the new record lines leave the guard's
  verdict unchanged.
- `scripts/check-contract-budget.sh` — the budget row for `skills/myflow-do/SKILL.md`, if the added
  section outgrows it.

## Decisions

Recorded in the change's `design.md` under `## Decisions`.

## Open questions

None.
