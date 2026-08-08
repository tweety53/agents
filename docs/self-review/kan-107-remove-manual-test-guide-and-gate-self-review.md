# Self-review — kan-107-remove-manual-test-guide-and-gate

**Change:** `kan-107-remove-manual-test-guide-and-gate`
**Jira:** KAN-107
**Date:** 2026-08-09
**Rating:** 2/5 — rough

The change landed correctly and nothing was lost, but the process that produced it was poor in two
specific ways the operator's rating reflects: the staged tree went stale and stayed stale through
four review-panel passes, and two fix instructions written by the controller were wrong in ways that
cost a full fix-and-re-review cycle each.

## 1. Problems, and the pipeline change that would prevent each

### Staging went stale, and four panel passes did not notice

`/myflow-do` stages at its section 7, before the review panel runs. Every panel fix wave afterwards
edited files that were then never re-staged. By the time the change reached the human gate, the
fourteen `docs/manual-test/` deletions — the central act of the change — were not in the index at
all, while the panel had been reading a diff built from `git diff <merge-base>`, which shows the
working tree rather than the index. The operator's gate command is `git diff --cached`, so the review
surface would have shown the change with its whole point invisible.

Bugbot found it on pass 3, by inspecting the index directly rather than the diff it had been handed.

**The change that would prevent it:** `/myflow-do` should re-stage after the panel converges rather
than before it runs, and its handoff should assert that the staged tree equals the working tree
before printing. A panel that reviews one tree while the operator is handed another is a gap no
number of review passes closes, because every slot is reading the same wrong surface.

### Two fix instructions were wrong, costing two rounds

Both were the controller's error, not the fix subagent's.

For F7, the instruction told the fixer to pass `$PLANTED/openspec/changes/clear` as a change name.
The guard's allowlist rejects any name containing `/` before it builds a path, so the "repaired" test
still exercised nothing beyond the allowlist — the exact defect the finding named. It took two
further rounds and four reviewers to establish that only a *relative* traversal name resolves onto
the planted tree once the allowlist is mutated away.

For F13, the instruction accepted a fixture that created a guide, removed it, and only then invoked
the guard — so the guard saw a tree identical to the neighbouring case's, and the case still proved
nothing. Two reviewers independently caught it.

**The change that would prevent it:** a fix instruction must name the observable that will prove the
fix, and the controller must confirm that observable is reachable before dispatching. Both errors
were a single shell command away from being caught.

### The panel record broke the guard that reads it

Reopening F7 was recorded by appending a second table row with the same identifier. The guard
requires each `F<n>` to name exactly one finding, so the record itself began reporting `OUTSTANDING`
— which would have stopped `/myflow-finish` run 1 at its unfinished-work gate.

**The change that would prevent it:** the contract should state how a finding is reopened — by
updating the existing row's note and flipping its marker, never by adding a second row.

### A self-modifying pipeline hazard, recorded without a fix

The installed `/myflow-finish` still carried the three-path staging chain while landing the change
that reduces it to two. Executed verbatim, its `reset` would have pulled the fourteen deletions out
of the index and its exclusion would have kept them out of the implementation commit, landing them
in the planning commit instead. The run used the two-path chain and said so.

No clean general fix suggests itself: a pipeline whose subject is its own contracts will always have
a window where the installed copy and the landing change disagree. Recorded so the next such change
expects it.

## 2. Token and time cost

Roughly 3.44M subagent tokens:

| Bucket | Tokens | Share |
|--------|--------|-------|
| Review panel — 4 passes × 7 slots, plus 3 fix waves | ~1.82M | 53% |
| 8 implementers plus 4 task-level fix rounds | ~1.06M | 31% |
| Per-task reviewers and re-reviews | ~0.56M | 16% |

The reductions were analysed during the run and filed as **KAN-108**. The largest single item is the
escalation trigger "the fix altered a public contract", which is vacuous in a repository that *is*
contracts: it fires on every fix round and forces a full seven-slot re-run where a targeted one would
have caught the same defects at under half the cost.

Two findings bouncing for three rounds each is the second cost driver, and it is the same root cause
as the wrong fix instructions above.

## 3. What went well, and how to reproduce it

**Mutation proof as the standard of evidence.** Three separate slots reproduced mutations
independently rather than reasoning about the code. It settled both genuine disputes of the run:
whether case 8f was load-bearing (delete the allowlist, watch the case fail with `CLEAR`), and
whether the provenance counts were 12/11 or 10/9. Bugbot withdrew its own count after re-deriving
with the guard's imported functions instead of a reimplementation.

*Reproduce it by:* requiring a mutation proof for any finding about a test, and requiring
re-derivation with the real code — imported, not reimplemented — for any finding about a number.

**Letting reviewers disagree.** F12 and F17 split the panel three-to-one on whether one clause was a
necessary invariant or a duplicated rule. The resolution reworded the guardrail sentence rather than
reverting the clause, and Lens B's dissent is recorded in the panel record rather than smoothed
away. The disagreement improved the text; harmonising it would not have.

*Reproduce it by:* never instructing a slot to defer to another, and recording surviving dissent as
dissent.

**Reports and diffs handed over as file paths.** No implementer or reviewer output entered the
controller's context, which is why a run with 30-plus subagent dispatches stayed coherent.

## 4. What could be automated

- **The staged-versus-working-tree check.** One command, and it would have caught this run's most
  serious defect four passes earlier.
- **`scripts/preserve-session-records.sh` does not cover the per-task and fix-round reports.** It
  preserves the ledger, the panel record and the proposal artifact source. The eight task reports and
  three panel-fix reports — the only record of what each implementer found, including the harness
  corrections and the mutation proofs — were preserved by hand during cleanup. Without that step,
  `git worktree remove --force` would have destroyed all eleven.
- **Panel-record bookkeeping.** The record is written by hand against a guard that validates it. A
  small helper to add a finding and flip its status would have made the duplicate-identifier defect
  impossible.

## Findings filed

None. All four candidate findings — the re-staging rule, the `preserve-session-records.sh`
extension, the reopen rule, and the fix-instruction rule — were offered and declined by the operator.
The cost findings were already captured in KAN-108 before this review ran.
