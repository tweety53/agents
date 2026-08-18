# KAN-200 — the self-review filing ask is per angle, and the report is checkable

**Date:** 2026-08-18
**Jira:** KAN-200
**Change:** `kan-200-self-review-filing-ask-per-angle`

## The problem

`/myflow-finish` run 2 step 8 already requires "**one** combined reasoning pass covering all four
angles … and offers a per-finding Jira filing ask before committing its report". In practice
"per-finding" has read as "per **problem**". KAN-73's run is the worked example:

- the **problems** angle produced five findings, all offered, all filed (KAN-192 … KAN-196);
- the **what went well** angle produced three items and no filing ask at all — they became tickets
  only because the operator asked afterwards (KAN-197, KAN-198, KAN-199);
- the **cost** angle produced a description of spend and nothing actionable; worked properly
  afterwards it yielded KAN-201 immediately;
- the **automation** angle listed only cross-references back to the problems angle's tickets;
  worked properly afterwards it produced KAN-202, a live defect.

Two angles were write-only and a third was a cross-reference table. Both angles that produced
nothing produced something real the moment they were actually worked — the omission was in the
execution, not in the material.

The rule that was skipped was already written down. That is the fact this design is built around: a
fifth angle and a per-angle filing ask stated only in prose would be skipped the same way, so the
report itself becomes a checkable artifact and a guard checks it.

## What changes

### 1. Five angles, stated once, each with a label

The canonical runtime statement lives in `skills/myflow-contracts/finish-contract.md` step 8 — the
file a run actually loads. `openspec/specs/myflow-self-review/spec.md` remains the requirement to
change first; `skills/myflow-finish/SKILL.md` step 8 carries only what is specific to *executing*
the procedure and cites the contract for the rest.

| # | Angle | Label |
|---|-------|-------|
| 1 | Problems encountered, and the pipeline change that would avoid them | `myflow-fix` |
| 2 | Token/time cost, and what would reduce it without quality loss | `myflow-cost` |
| 3 | What went well, and how to reproduce it | `myflow-improvement` |
| 4 | What could be automated or moved to a script | `myflow-automation` |
| 5 | What could move to the Go app or its persistent storage | `myflow-stats-app` |

Angle 5's remit covers **records and logic**: which artifacts the pipeline writes to files today
(SDD ledgers, panel records, self-review reports, the state-file fallback, task completion) would be
better held in `myflowd`'s Postgres and queryable across runs, and which derivation work now done by
a Bash guard or by the agent itself could move into the `myflow` CLI or the daemon. It deliberately
stops short of the SPA's presentation, which would turn the angle into a wishlist.

Still one combined reasoning pass — five angles, never five dispatches. Every angle produces zero or
more findings, and an angle that yields nothing says so **explicitly**, the way `## Decisions` and
`## Open questions` must be present-but-empty rather than absent. An angle that is silent is
indistinguishable from an angle that was skipped, which is exactly how KAN-73's cost angle passed
unnoticed: the report had a Cost section, so nothing *looked* missing.

### 2. Explain before filing — and the prompt only records the decision

The operator's words: *"I need a detailed explanation before filing a jira issue. Very often I don't
understand what it is about."* A filed ticket is durable and already on the board; an explanation
that arrives afterwards describes something the operator did not agree to and cannot cheaply undo.

Every finding is explained in the message body **before** any prompt fires, carrying three things:
what was observed, what breaks because of it, and what the fix would be. A choice prompt's option
text cannot carry that, so it does not try — the prompt records the decision and nothing else.

The prompt is **one multi-select per angle**, listing that angle's findings, defaulting to filing
nothing. Five prompts is the ceiling for a run; one prompt per finding would be up to fifteen, and a
single prompt across all angles loses the grouping that makes a fifteen-item list readable.

This binds **every** filing the pipeline offers, not only self-review's: run 1's follow-up filing
(`skills/myflow-contracts/jira-followups.md`) has the same shape.

### 3. A filed issue carries its angle's label

On top of the existing inherited set — every label on the change's linked issue plus `AI-generated`,
per **Labels on issues the pipeline creates** (`skills/myflow-contracts/jira-integration.md`) — a
filed finding carries the label naming the angle that produced it. "Which angles are actually
producing findings" becomes a query rather than an audit.

The back-fill the ticket asks for is already done: KAN-192 … KAN-202 each carry their angle label as
of 2026-08-18. Only the rule needs landing.

### 4. The report becomes a record, not prose

`docs/self-review/<name>-self-review.md` gains a fixed shape:

- one `##` section per angle, all five always present, in the order above;
- one line per finding, in a fixed form naming the angle label, the finding, and its disposition:

  ```markdown verified:the shape this change introduces; asserted by the guard tasks.md task 2 builds
  - **[myflow-cost]** Every panel slot gathers the same context independently — filed: KAN-201
  - **[myflow-fix]** Preflight compares against a stale local base ref — declined
  ```

- an angle with no findings carries the explicit marker instead of an empty section.

This is what makes the rule checkable at all, and it is angle 5's own first candidate: a per-finding
record with an angle, a disposition and an issue key is data, and it currently lives in seventeen
Markdown files nothing can query.

### 5. `check-self-review-report.sh`

A guard over `docs/self-review/*.md`, following this repository's established guard shape — a
`SELF-REVIEW-REPORT-OK` / `SELF-REVIEW-REPORT-INVALID` verdict line, exit 0 clean / 1 violations / 2
cannot answer — and using `scripts/lib/coverage.sh` for per-member coverage, so a report that is
checked for nothing is a visible, failing fact rather than a silent pass.

Rules:

1. every one of the five angle sections is present;
2. every section carries either at least one finding line or the explicit none-marker;
3. every finding line parses; its angle label matches the section it sits under; a finding marked
   filed names an issue key.

The seventeen reports that predate this rule — the sixteen present when this change began, plus
the KAN-197 report section 6 below recovers — are `coverage_declare`d **by name, with a reason**, in
the guard's own source — the mechanism KAN-197 established for exactly this case. A report that is
not on that list must satisfy all three rules, so a new report can never pass by being skipped, and
the declared list only ever shrinks. A format-marker scheme was rejected for the opposite property:
a new report that simply forgets the marker would be skipped rather than failed, which is the
vacuous-pass shape KAN-197 exists to stop.

The guard joins `.myflow/project.md`'s `## lint` list and its harness joins `## test`.

### 6. Carried alongside: KAN-197's stranded self-review report

Commit `0aea735` — `docs/self-review/kan-197-require-mutation-test-for-every-guard-self-review.md` —
was pushed to `chore/archive-kan-197` at 11:09Z on 2026-08-18, five minutes after PR #13 merged at
11:04Z. It exists on that branch and nowhere else, and never reached `main`.

The cause is a real gap in run 2 step 8, which says to commit "on the base branch **in the main
checkout**" and never verifies that the main checkout is *on* the base branch. It is recorded here
because this change is the one that touches step 8; the operator chose to carry the recovery on this
branch rather than open a separate PR for it.

## What is deliberately not in scope

- **The branch-verification fix itself.** Recovering the file is in scope; changing step 8 to assert
  the checkout's branch is a separate defect with its own reasoning, and folding it in here would
  mix a filing-ask change with a git-safety change.
- **Any change to the four existing angles' wording.** They are restated in one canonical place, not
  rewritten.
- **Retro-labelling.** Already complete.

## Decisions

### Angle 5 covers records and logic, not presentation

**ID:** angle-5-remit
**Status:** active
**Chosen:** records plus logic — which artifacts the pipeline writes to files today would be better
held in `myflowd`'s Postgres and queryable across runs, and which derivation work now done by a Bash
guard or by the agent could move into the `myflow` CLI or the daemon.
**Considered:** records and state only — narrower and safer, but it would exclude the derivation
work that KAN-196 and KAN-202 both turned out to be about, which is the half with live defects in
it; records, logic and the SPA's presentation — rejected because "what should the UI show" produces
wishlist items rather than pipeline changes, and the angle is meant to yield findings that can be
filed.

### The findings are explained in full before any prompt, and the prompt is one multi-select per angle

**ID:** filing-ask-shape
**Status:** active
**Chosen:** every finding is explained in the message body first — what was observed, what breaks,
what the fix would be — then one multi-select prompt per angle records the decision. Five prompts is
the ceiling for a run.
**Considered:** one prompt per finding, the literal reading of the current spec — faithful, but with
five angles it is up to fifteen sequential prompts, and the prompt's option text cannot carry the
explanation anyway; a single multi-select across all angles — fewest interruptions, but a
fifteen-option list loses the angle grouping that makes it readable, and the grouping is what the
labels exist to preserve.

### The report shape is enforced by a guard, not by prose

**ID:** report-enforcement
**Status:** active
**Chosen:** `scripts/check-self-review-report.sh` plus its harness, wired into `## lint` and
`## test`.
**Considered:** contract prose alone — rejected because it is precisely the failure mode this change
exists to fix: the four-angle rule was already written down in three places and was still not
followed; an advisory guard that reports but always exits 0 — rejected for the same reason, since an
incomplete report would still land.

### Pre-rule reports are declared by name, not detected by a marker

**ID:** legacy-report-handling
**Status:** active
**Chosen:** the seventeen pre-rule reports — the sixteen already present, plus the KAN-197 report
this change recovers — are `coverage_declare`d in the guard's own source with a
reason, the mechanism KAN-197 established for a legitimate zero. Anything not on that list must
satisfy every rule, and the list only ever shrinks.
**Considered:** a `<!-- self-review-format: 5-angle -->` marker in each new report — cheaper, but a
new report that forgets the marker is *skipped* rather than failed, which is the vacuous-pass shape
KAN-197 exists to stop; retrofitting all seventeen with "none — predates this rule" stubs — uniform,
but it writes seventeen files' worth of placeholder into the historical record to avoid one list.

### Each finding carries a parseable record line

**ID:** per-finding-record-line
**Status:** active
**Chosen:** a fixed line per finding naming its angle label, the finding, and its disposition
(`filed: <KEY>` or `declined`), so the guard can check more than section presence and a filed ticket
traces back to the angle that produced it.
**Considered:** checking section presence only — a simpler guard and a freer report format, but
nothing would connect a filed ticket to its angle except the Jira label, and the label is applied by
the same step that would be skipped.

### KAN-197's stranded report is recovered on this branch

**ID:** stranded-report-recovery
**Status:** active
**Chosen:** cherry-pick `0aea735` onto this change's branch, so the report reaches `main` with this
PR.
**Considered:** a separate recovery PR — cleaner diff, but a second PR for one file; leaving it lost
and filing the underlying bug — rejected by the operator, since the report is the record of a run
that is already finished and cannot be regenerated.

## Open questions

### Should run 2 assert that the main checkout is on the base branch before committing the report?

**ID:** run-2-base-branch-assertion
**Status:** open
**Why it is open:** deliberately scoped out of this change. It is a git-safety defect with its own
reasoning, and folding it into a filing-ask change would mix two unrelated concerns in one diff.
**What it affects:** whether a future self-review report can be stranded the same way KAN-197's was.
Recovering that one file is in scope here; preventing the next one is not.
