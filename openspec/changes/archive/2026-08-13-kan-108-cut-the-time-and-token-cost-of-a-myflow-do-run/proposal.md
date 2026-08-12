# Cut the time and token cost of a `/myflow-do` run

## Why

KAN-108 records a measurement taken on the KAN-107 run, which spent roughly 3.44M subagent tokens:
the review panel accounted for ~1.82M of that (53%), the eight implementers plus their fix rounds
~1.06M, and the per-task reviewers ~0.56M. Two structural causes are isolated, ranked by measured
saving.

**The escalation trigger `the fix altered a public contract` is vacuous in this repository.** This
repository *is* contracts, so the clause fires on every fix round and never discriminates, forcing a
full re-run each time. Pass 3 of that run cost ~430k tokens; a targeted re-run of the primary slot
plus the one slot holding an open finding would have caught the same two defects at ~190k.

**A wrong fix instruction is dispatched as readily as a right one.** Two findings bounced for three
rounds each because the instruction itself was wrong — one told the fixer to pass an absolute path
that the guard's name allowlist rejects before any path is built, so the repaired test still proved
nothing; the other accepted a fixture teardown that ran before the guard was invoked, leaving the
case a duplicate of its neighbour. Running the check each instruction described, and observing that
it passed, would have caught both before a single fix round was spent.

KAN-109 touched the same section of `skills/myflow-do/SKILL.md` and left both of these open — its
proposal states the escalation ladder's trigger conditions are untouched, and it added no
verification ahead of the fix dispatch. What KAN-109 did change is what a *Full* escalation costs
once it fires, so part of the first item's gap is already collected. The remaining gap is that the
trigger forces Full rather than Targeted on every fix round here.

## What Changes

- **The vacuous clause is reworded.** In `/myflow-do`'s auto-escalate trigger list, `a public
  contract` becomes `a guard's behaviour`. The set becomes: the fix touched a file outside the set
  named in the findings; the fix diff exceeds ~150 changed lines; the fix altered a delta spec, a
  migration, or a guard's behaviour; a targeted re-run surfaced a new Critical finding; three or more
  fix rounds have already run. That set appears in no capability spec today, so it becomes a
  requirement — otherwise the next reword has nothing to check it against.
- **A fix instruction is not dispatched until its reproducer fails.** Every slot's dispatch prompt
  SHALL require, per finding, either a runnable command demonstrating the defect or `none —
  <reason>`. Before dispatching the fix subagent, `/myflow-do` runs each supplied reproducer in the
  worktree and requires a non-zero exit. On exit 0 the finding is bounced **once** to the slot that
  raised it, carrying the passing output; a second passing reproducer stops the run at the operator
  handback already in section 5. A bounce consumes no fix round, on the precedent section 4 sets for
  `check-task-commit-fields.sh`. A finding recorded `none — <reason>` is dispatched without a run,
  and `none — not supplied by <slot>` is legal and dispatched unverified, visible in the record.
- **The panel record gains a reproducer marker block.** A `reproducers-total: <n>` line and one
  `finding-reproducer: F<n> <command | none — reason>` line per finding, in a block of its own —
  separate from the `finding-status:` block, whose lines `check-unfinished-work.sh` requires to
  occupy one unbroken span.
- **A guard enforces the field.** New `scripts/check-panel-reproducers.sh <worktree>`, anchored
  regexes over the marker block only: exit 0 every `F<n>` has exactly one well-formed reproducer line
  and the count matches, exit 1 violations found, exit 2 cannot answer. It does not join `## lint` —
  it needs a change in flight and a real worktree, the same reason `check-panel-diff-size.sh` is
  excluded — and is covered by its own harness under `## test`.
- **Collateral:** `principles-reviewer-prompt.md` and `adversarial-reviewer-prompt.md` gain the
  reproducer requirement; `.myflow/project.md`'s `## test` list gains the new harness. No contract
  budget re-anchor is needed — `skills/myflow-do/SKILL.md` is at 46,899 bytes against a 58,623-byte
  budget.

## Added during implementation

**A runner, `scripts/run-reproducer.sh`**, added at the operator's direction after the review panel's
fifth pass. The change as first proposed left every rule about *running* a reproducer as prose in
`skills/myflow-do/SKILL.md`: the containment of each token, the ban on shell metacharacters, the direct
exec, the bound, and the timed-out and surviving-process dispositions. Eight findings across three
panel passes were defects in that prose, and each pass produced another, because a rule nothing
implements can only be re-read rather than tested — most plainly when a fix for one of them cited
`ps -o sid=`, a flag that does not exist on this platform. The runner turns those rules into a script
with its own harness, so the constraints are enforced where they were described, and
`skills/myflow-do/SKILL.md` cites an invocation instead of carrying a procedure. Two panel findings are
closed by it directly: the quote characters missing from the banned set, and the invalid session-kill
citation.

## Capabilities

### Modified Capabilities

- `myflow-review-panel-economics`: gains **Requirement: The auto-escalate trigger set is stated and
  discriminates** (the reworded set, and why a clause that fires on every fix round is not a
  trigger), **Requirement: A fix instruction is verified by a failing reproducer before dispatch**
  (the field, the pre-dispatch run, the one bounce, the operator handback, and the exemption forms)
  and **Requirement: A bounced finding consumes no fix round** (the guard-class accounting). The
  existing zero-open-findings bar, the findings table, the `finding-status:` marker rules, the diff
  cap and the roster are untouched.

## Impact

- Files: `skills/myflow-do/SKILL.md` (the **Panel re-runs** subsection and the slot dispatch briefs
  in section 5), `skills/myflow-do/principles-reviewer-prompt.md`,
  `skills/myflow-do/adversarial-reviewer-prompt.md`, new `scripts/check-panel-reproducers.sh` + new
  `scripts/test-check-panel-reproducers.sh`, `.myflow/project.md`'s `## test` list.
- No change to the state file shape, the three-state machine, the review panel roster presets, the
  optional-slot trigger table, the panel record's findings table or `finding-status:` rules, the
  zero-open-findings bar, the panel diff cap, or any `/myflow-start` / `/myflow-finish` /
  `/myflow-status` behaviour. `/myflow-fast` inherits everything here through the `/myflow-do`
  sections it cites, with no edit of its own.
- Design:
  `docs/superpowers/specs/2026-08-12-kan-108-cut-the-time-and-token-cost-of-a-myflow-do-run-design.md`.
