# Self-review — kan-15-parallel-myflow-do-task-lanes

Three repositories, 36 subagent dispatches, seven review-panel passes, ten fix rounds, 84 findings.
The feature the operator asked for — `-Pservices` — was clean from the first pass that saw it. Almost
everything below is about what happened around it.

## 1. Problems, and the pipeline change that would avoid each

### P1 — The SDD ledger was one command away from being lost, every time

`scripts/preserve-session-records.sh` reads the ledger from `.superpowers/sdd/tasks/progress.md`.
`/myflow-do` writes it to `.superpowers/sdd/ledger.md`. The script did exactly what it promises —
reported the source absent and skipped it, not a failure — and the 152-line record of every dispatch,
model and fix round would have died with the worktree. It survived because the skip was noticed by
hand at run 1 and copied across manually.

**This is not specific to this change.** Any change finishing today loses its ledger the same way, and
the loss is silent: the script's `skipped: … (absent)` line is indistinguishable from the legitimate
case it exists for (a change that genuinely has no panel record).

**Fix:** make the two paths agree — either the script reads `ledger.md`, or `/myflow-do` writes where
the script reads. Preferably also: a source that is absent for *every* change is a defect, not a skip,
so the script could distinguish "this change has none" from "nothing ever writes here".

### P2 — The panel did not converge for three rounds, and nothing noticed but me

Findings per round: round 5 closed 6 and produced 8. Round 6 closed 8 and produced 13, three of them
Critical. Every one of those was against machinery inferring which port a service was on — never
against the feature. The inference could not work: a pid file says a process exists, not what it
bound; a port file says what a process *intended* to bind, written a second after spawn; `lsof` says a
port is busy, not whose it is.

The escalation ladder handled breadth correctly — it widened the roster each time — but **widening the
panel against a design that cannot work buys more findings, not fewer**. The pipeline has no notion of
a trend. It was the operator, given the counts, who chose to delete the inference instead of repairing
it again; the two rounds after that reversal went 10 → 1 → 0 Criticals.

**Fix:** a convergence check. When a fix round produces more findings than it closed — twice running,
say — stop and put the trend to the operator with the alternative, rather than escalating breadth
again. The ladder currently terminates by exhausting the roster; it should also terminate on evidence
that repair is losing.

### P3 — A comment claiming what the code does not do produced findings in four separate passes

F73, F76, F77 and F83, plus part of F82. Each was written by a fix round explaining why its repair was
sound. F73 is the sharp one: the previous round's entire seam argument rested on "both sides read
through one predicate", and `startService` called the other function directly. It survived a full
review pass because reviewers read the explanation.

F83 is the last instance and the clearest: a KDoc justified deleting a record rather than truncating it
by asserting a truncated file reads as `Recorded.Unknown`. It reads as `Recorded.None`, identical to an
absent one. The code was right; the stated reason was false, and a two-file read disproved it.

**Fix:** the principles-reviewer prompt should carry an explicit instruction to verify every
load-bearing claim in a comment against the code it describes, and to treat a false one as a finding at
the severity of the claim. Reviewers in this run did that only when told to.

### P4 — A test that could not fail shipped, and was counted as coverage

F74. The test written to cover the sibling case marked the sibling as running, but the port it held was
never probed — so it passed identically with the sibling not running, and would have stayed green
through the exact regression it existed to catch. It took an adversarial reviewer constructing the
mutation to see it.

**Fix:** when an implementer or fix round writes a test to pin a specific regression, require it to
report the mutation result — change the code the test guards, show it goes red. Fix round 9 did this
unprompted for the rewritten test and it was the strongest evidence in the run.

### P5 — One issue, two scopes, and finish moved all of it to Done

KAN-15's headline ask is parallel lanes, with a measured 20-minute saving. This change explicitly did
not build them — `proposal.md` says so — and delivered the *"IMPORTANT addition"* instead: workspace
isolation. Run 2 transitioned the whole issue to Done, because the pipeline transitions the linked
issue after archiving and that rule assumes one change per issue.

**Fix:** when a change's proposal explicitly disclaims part of its linked issue's scope, finish should
ask before Done rather than transitioning. The disclaimer is already written down in a fixed place.

### P6 — The feature shipped having never been executed

Every implementer and reviewer was forbidden `dev*` tasks, because the operator's stack was live and an
earlier dispatch had already killed their backend with an unscoped `pkill`. So `-Pservices`, the
recorded port block and every new refusal reached the human gate untried, and the operator integrated
over that knowingly.

**This change is its own fix.** Workspace isolation is what lets a worktree run its stack without
touching the operator's. The ban was the right call *before* this change existed; the next change
should be able to drop it.

## 2. Cost, and what would cut it without losing quality

**Measured:** ~4.8M subagent tokens across 36 dispatches.
<!-- measured: summed from the per-dispatch usage lines in this session's task notifications -->

The distribution is the finding. The feature — `-Pservices` — cost three implementer dispatches and one
review. The port-block machinery cost seven panel passes and ten fix rounds.

**P2's convergence check is the saving.** The reversal happened after pass 6; the evidence for it
existed after pass 5. Stopping one pass earlier would have avoided round 6 and pass 6 outright —
roughly 1.4M tokens, or 30% of the run, with no quality lost, because round 6's work was deleted a
round later anyway.

**Cutting the roster on evidence worked and should be reusable.** At pass 5's handback the operator
chose four slots over seven, on the grounds that primary and two principles lenses had been clean twice
running. Those four then returned 13 findings including three Criticals. The lesson is not "fewer
slots" — it is that slot yield is observable within a run and can be acted on. Worth making an offered
option at a handback rather than an ad-hoc question.

**What was not waste:** the passes after the reversal. Pass 8 confirmed both Criticals closed by tracing
rather than by re-reading, pass 9 checked the new porcelain parser against real `git worktree list`
output, and pass 10 found the false KDoc claim. Each of those cost a fraction of a repair round and
caught something a repair round had introduced.

## 3. What went well, and how to reproduce it

**The operator's two decisions moved the outcome more than any agent's.** Cutting the roster, and
choosing to delete the inference rather than repair it. Both came from being shown counts and a trend
rather than a summary — the reproducible part is putting the *numbers* in the question, not the
conclusion.

**Deleting beat repairing, twice, in the same change.** Earlier the operator asked why the frontend port
could not rotate; the answer retired 786 lines and a Critical that had appeared three times. Later the
same shape recurred and 885 more lines went. Lens B named it the first time — *"it didn't reduce the
cost of a hazard, it deleted the hazard"* — and was brought back deliberately to judge the second. Both
times the prompt was a question about a constraint everyone had accepted.

**Pointing every reviewer at the repair rather than the finding.** From pass 5 on, each prompt said so
explicitly, because pass 2 had found two of its own fixes defective. Passes 8 and 9 then confirmed
closures by tracing the new call graph instead of re-reading the old finding — which is what caught F71
and F79, both introduced by the repair of something else.

**The change verified its own cleanup.** At finish, gymie's guard reported *"the workspace survivor
report for kan-15-55a6 is empty"* — the capability this change added, checking that its own resources
were gone.

## 4. Automation candidates

| Candidate | Shape | Value |
|---|---|---|
| Ledger preservation path (P1) | One-line fix in `preserve-session-records.sh` or `/myflow-do` | Stops silent loss of every change's ledger |
| Convergence check (P2) | Contract rule + a count the panel record already carries | ~30% of this run's tokens |
| Comment-claim verification (P3) | A line in `principles-reviewer-prompt.md` | Four passes' worth of findings in this run alone |
| Mutation proof for regression tests (P4) | A line in the implementer dispatch block | Catches tests that cannot fail |
| Scope-disclaimer check before Done (P5) | Finish reads `proposal.md` for an explicit disclaimer | Stops closing an issue whose headline ask is unbuilt |
| Multi-repo planning-artifact noise | `check-unfinished-work.sh` treats a repo without the planning artifacts as outstanding | Two of three repos reported OUTSTANDING for structural reasons at this finish |

The panel record already carries the count each pass needs for the convergence check —
`findings-total` and the per-pass tables — so P2 needs a rule, not new data.

## Rating and filings

**Operator's rating: 3 / 5 — mixed.** Right outcome, expensive route: the feature is sound and
covered at the unit level, but the run needed the operator to spot the non-convergence, and what
shipped has never been executed against a running stack.

Five findings were filed, each labelled `myflow` + `AI-generated` and linked to KAN-15:

| Issue | Finding |
|---|---|
| KAN-77 | P1 — `preserve-session-records.sh` loses every change's SDD ledger |
| KAN-78 | P2 — no convergence check; a losing repair escalates breadth instead of stopping |
| KAN-79 | P3 — reviewers read a comment's explanation instead of the code |
| KAN-80 | P4 — a test written to pin a regression shipped unable to fail |
| KAN-81 | multi-repo planning-artifact noise in `check-unfinished-work.sh` |

**P5 was offered and declined** — finish moving KAN-15 to Done although the change explicitly
disclaims the issue's headline scope. It is recorded here rather than tracked. It remains true of the
board: KAN-15 reads Done, and the parallel lanes it is named for do not exist.

P6 needs no issue: this change is its own fix, and the next change should be able to drop the ban on
running the applications.
