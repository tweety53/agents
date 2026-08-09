# Self-review — kan-110-lighter-auto-code-review-by-default

**Date:** 2026-08-09
**Jira:** KAN-110
**Rating:** 4/5 — good
**Planning effort:** default · **Models:** implementation, review panel and panel fixes all Sonnet
**Panel roster:** `light` by operator override, ahead of the change that introduces it

## What the change did

A change now records a review panel roster preset — `light` (default), `standard`, `full` — in
`reviewPanelRoster`, and `/myflow-do` selects both the panel's required slots and the per-task
review's shape from it. `light` swaps Bugbot for a `Code review (low)` slot invoking the harness's
`code-review` skill; the two lighter presets offer conditional slots in one prompt rather than
auto-including them; Bugbot's brief gains reasoned mutation testing. `full` reproduces the prior
behaviour exactly, and no preset moves the handoff bar.

Ten tasks, dispatched as six implementer runs plus three fix rounds. Twelve files, 386 insertions,
49 deletions. One new test harness.

## Angle 1 — problems, and the pipeline change that would avoid them

### Planning artifacts never reach the worktree — filed as KAN-112

`/myflow-start` stages the change directory and the design doc in the main checkout without
committing. `/myflow-do` branches from HEAD, which predates them, so the worktree has no plan and
brief generation fails outright. Worked around by copying them in by hand.

It then cost a second time at finish run 1: the branch had committed the artifacts while the main
checkout still held its own staged copies of the same paths, so the merge refused. Clearing them
safely meant verifying each file byte-identical to the committed version first.

This is hit on every first run. The fix belongs in `/myflow-do` section 2 (copy them in, clear them
from the main checkout's index) or in `/myflow-start` (do not stage into a tree that will not carry
the branch).

### The ledger was silently not preserved — filed as KAN-113

`preserve-session-records.sh` reads the ledger from `.superpowers/sdd/tasks/progress.md`; the SDD
skill describes `<workspace>/progress.md` and `skills/myflow-do/SKILL.md` states no path at all. The
ledger was written to `.superpowers/sdd/progress.md` and the script reported
`skipped: … (absent)`, exit 0 — indistinguishable from a change that legitimately has no ledger.
Caught only by reading the output.

A miss that presents as a benign no-op is the dangerous shape here, and it is the same class this
repository already guards against elsewhere.

### A guard passes vacuously — filed as KAN-114

`check-task-build-green.py` matches task headings as `^### <dotted-id>`. A plan using any other
shape yields zero tasks and exits 0 having checked nothing. Measured: the archived
`2026-08-09-kan-107-remove-manual-test-guide-and-gate` plan carries eight `**Build:**` tags and the
guard sees zero tasks. It passed its pre-publish gate on that basis.

### The vocabulary carve-out took two rounds — not filed

`check-vocabulary.sh` bans the bare token `code-review`, a myflow command retired in the twelve-stage
collapse, which is byte-for-byte the harness skill the new slot invokes. The first implementer
worked around it by writing the name with a space — a broken instruction that lints clean. The
operator approved a carve-out instead.

Round 1's carve-out excluded any backtick-preceded occurrence, which blinded the guard to the retired
token in the spelling this repository actually uses for command names. It was caught by probing the
guard rather than by the reviewer, and round 2 narrowed it to a per-occurrence exemption with a new
`scripts/test-check-vocabulary.sh` proving a line carrying both an exempt mention and a genuine
retired token still reports.

The general lesson — a carve-out should ship with a test proving what it does *not* exempt — was
considered for filing and declined; round 2 built that test, and the guard's header already takes
the "not a completeness proof" stance the rule would formalise.

### Two deferred minors remain in the shipped carve-out

Both are recorded in the panel record and the ledger, and both survived the panel's judgment that
neither is worse than recorded:

- the exemption matches the singular word `skill` only, so a future mention writing the plural is a
  false positive, uncovered by any test case;
- the exemption is line-scoped, so a line wrap separating the backticked name from the following
  word produces a false hit. It fired once, during Task 9, and was fixed by reflowing the line.

### A handoff line went stale and was caught only incidentally

`skills/myflow-do/SKILL.md`'s handoff template still read `**Panel:** clean — required: primary +
Bugbot + Principles`, which the lighter rosters make false. No task's brief covered it; an
implementer flagged it as out of its own scope, and it was folded into a fix round. Nothing in the
pipeline links "the roster changed" to "the handoff template names a roster" — recorded as an
observation rather than a filed change, since the answer here was reviewer diligence rather than
mechanism.

## Angle 2 — cost, and what would reduce it without losing quality

Twenty-one subagent dispatches: six implementer runs, three fix rounds, nine task reviews and
re-reviews, three panel slots. All on Sonnet.

What kept it down:

- **Grouping coupled tasks.** Tasks 2–5 all edit `skills/myflow-do/SKILL.md` and went to one
  implementer; Tasks 7–8 are two small carry-forward edits and went to another. Six dispatches for
  ten tasks, and the grouping was forced by file coupling rather than chosen for economy — the two
  motives happened to agree.
- **One combined reviewer per task** instead of the spec-compliance and code-quality pair, which is
  the `light` preset's own per-task shape applied ahead of the change landing. Roughly halved the
  review dispatches. Worth naming plainly: this is the feature under review being used to review
  itself, and the saving is therefore evidence *for* the feature only as far as the panel's clean
  result supports it.
- **Running the guard and test sweep in the controller** rather than dispatching Task 10. It is
  verification, which section 7 of `/myflow-do` performs anyway.

What cost most avoidably:

- **The carve-out round trip.** Round 1's too-broad fix cost a dispatch, the controller's probe, and
  a second dispatch carrying a new harness. A negative test required up front would have collapsed
  it to one round.
- **Contract reading.** `finish-contract.md` (444 lines) and `plan-provenance.md` (366 lines) were
  both read in full mid-run, and both were needed. This is the standing per-command cost the
  contract-economy work exists to bound, not a fault of this run.

## Angle 3 — what went well, and how to reproduce it

- **Probing a guard rather than trusting a subagent's account of it.** The implementer's claim that
  `code-review` collided with the retired-token regex was correct; its workaround was wrong, and the
  first carve-out was wrong in a different way. Both were found by running the guard against a
  three-line probe file. Reproduce: whenever a subagent reports working around a guard, run that
  guard against a minimal case yourself before accepting the workaround.
- **Asking before overriding a required slot, and again before declining fired triggers.** The
  operator's instruction to run three reviewers arrived before anyone knew the diff would be 433
  lines with a guard's matching logic rewritten. Surfacing which triggers fired, and what each would
  have read, kept a standing instruction from silently becoming a skipped review.
- **Amending the plan the moment it was contradicted.** The Global Constraint said
  `check-vocabulary.sh` is not modified. Rather than proceeding against it, `proposal.md` and
  `tasks.md` were amended in place with the collision, the operator's decision and the two rejected
  alternatives — before the fix was dispatched.
- **Verifying byte-identity before a destructive step.** Clearing the main checkout's staged
  duplicates was only safe because each file was first proved identical to the version the merge
  would restore.

## Angle 4 — what could be automated or moved to a script

- **The worktree plan copy** (KAN-112) is pure mechanism and belongs in a script or an explicit
  skill step, not in a controller's judgment.
- **The ledger path** (KAN-113) is a one-line agreement between a producer and a consumer that
  currently exists in neither.
- **Vacuous-pass detection** (KAN-114) generalises past `check-task-build-green`: any guard that
  iterates a set should report the size of the set it found, so zero is visible in the output rather
  than inferable only by reading the input.
- **Panel-record scaffolding.** The findings table, the marker block and the total line were written
  by hand and are entirely mechanical given a findings list. A helper that emits a well-formed record
  from a list would remove a class of format error the guard exists to catch.

## Findings filed

| Finding | Outcome |
|---------|---------|
| Planning artifacts never reach the worktree | filed — KAN-112 |
| `preserve-session-records.sh` ledger path mismatch reports a silent skip | filed — KAN-113 |
| `check-task-build-green` passes vacuously on unmatched headings | filed — KAN-114 |
| A new vocabulary carve-out should ship with a negative test | declined — round 2 built the test; the guard's header already takes the stance |
| The stale `**Panel:**` handoff line | not filed — an observation, with no mechanism to propose |
| Two deferred minors in the shipped carve-out | not filed — recorded in the panel record and the ledger |
