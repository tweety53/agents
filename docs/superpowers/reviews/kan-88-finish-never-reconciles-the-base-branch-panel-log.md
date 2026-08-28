# Review panel — kan-88-finish-never-reconciles-the-base-branch

## Pass 1 (round 0) — full roster

**Mode:** initial panel. Every slot in the resolved roster ran; no re-run scoping applies.

**Roster resolved from the settings store:** `primary`, `principles`, `code-review-low`, `bugbot`.
No addition this round — the resolved list ran alone. The operator named no additional slot in this
run's argument or in the session before this stage.

**Substitution:** the `bugbot` id's own agent type is not offered by this harness (Claude Code's
pre-dispatch agent-type listing carries no `bugbot`), so that slot ran as a **general-purpose**
subagent carrying Bugbot's brief plus the mutation-testing requirement, per **An unspawnable id is
substituted, not skipped**. Its dispatch is recorded with `-slot Bugbot` and `-model sonnet` — the
model actually given, never `unknown (agent-defined)`.

**Model:** `sonnet` for every slot — the settings store's `defaultModel`. No session-instruction
override was given this run.

**Diff size:** `check-panel-diff-size.sh <worktree> c761592` measured **672** against the cap in
force and exited 0 (`under cap`). No operator question was reached, so none was answered.

**Diff every slot read:** `.superpowers/sdd/final-review.diff`, 793 lines, written from
`git diff c761592` — the whole branch, staged and unstaged. Bugbot's substitute additionally read
the working tree directly, as its brief requires.

**Standards passed to the Principles slot:** `CLAUDE.md` and `AGENTS.md`, the two `## standards`
entries `.flow/project.md` declares. Principles path:
`/Users/tweety53/.claude/skills/flow/engineering-principles.md`.

## Pass 2 (round 1) — Full mode, escalated

**Mode:** Full, and **not** because the operator asked. Fix round 1's diff touched
`scripts/test-check-finish-preflight.sh` and `scripts/test-check-base-moved.sh`, which are outside
the file set F1–F3 named (both findings' locations name the guards themselves), and a fix touching a
file outside the findings' named set is a stated automatic escalation trigger. No question was put
to the operator, per that trigger's own "do not ask, and say why in the record".

**Who ran:** every slot in the resolved roster — Primary, Principles, Code review (low), Bugbot
(substituted as general-purpose, as in pass 1). No slot outside the resolved roster was added; the
operator named none.

**Diffs they read:** Primary the rewritten whole-branch `final-review.diff` (906 lines, measured 779
by `check-panel-diff-size.sh`, under cap). Principles and Code review (low) each their own delta,
`git diff e39ce94 HEAD` — 136 lines — at `.superpowers/sdd/slot-delta-1-Principles.diff` and
`.superpowers/sdd/slot-delta-1-CodeReviewLow.diff`. Bugbot read no diff file, per its slot.

**No finding was bounced this pass** — every finding recorded in round 0 carried the `none — <reason>`
exemption form rather than a runnable command, so none was eligible for a reproducer run.

**F1–F3 closure, and the operator decision behind it.** The contract requires a fix's diff to touch
at least one path the finding named before that finding can close. F1–F3 are surviving mutants: the
guards' own code was already correct, so the only correct repair is a test in the paired harness, and
the rule read literally can never be satisfied for this finding class. The dispatcher put exactly
that to the operator rather than waiving the rule on its own judgment. The operator chose: the paired
harness `scripts/test-<guard>.sh` counts as within the path set a finding located in
`scripts/<guard>.sh` names. F1–F3 recorded fixed on that basis. **The contract itself still says
otherwise and is worth a follow-up**, since the next surviving-mutant finding hits the same wall.

**Mutation proof of what fix round 1 changed**, re-run by the dispatcher rather than taken from the
fix subagent's report:

```text
fix-mutation: scripts/check-finish-preflight.sh — collapsed the ANCESTOR_RC -eq 1 / -ne 0 two-step so any nonzero merge-base exit reads as RUN1 — case 8c failed (3 assertions)
fix-mutation: scripts/check-base-moved.sh — dropped the failure guard on the COUNT assignment — case 9b failed (3 assertions)
fix-mutation: scripts/check-base-moved.sh — moved the overlap cap boundary from -gt 10 to -ge 10 — case 6b failed (1 assertion)
fix-mutations-total: 3
```

**Findings raised this pass:** F4 (Principles, Important — the failing-git shim block written out
three times), F5 and F6 (Primary, Minor — `tasks.md` baselines left stale by fix round 1), F7, F8 and
F9 (Bugbot, Important/Important/Minor — three further surviving mutants). F5 and F6 were repaired by
the dispatcher, since they are planning-artifact edits the fix subagent is forbidden to make. The
rest go to fix round 2.

**A convergence risk, named rather than discovered later.** F8's own report notes that
`COMMITTED_RAW`, `STAGED_RAW` and `UNSTAGED_RAW` share `MOVED_RAW`'s shape and its gap. Repairing one
capture per round would find a sibling every round without end, so fix round 2 is dispatched to cover
**every** capture guard of that shape at once, not the single one the finding names.

## Pass 3 (round 2) — Full mode, escalated again

**Mode:** Full, automatic again. Fix round 2 touched both test harnesses and added
`scripts/lib/test-git-shim.sh` — outside the file set F4/F7/F8/F9 named, and over ~150 changed lines
while adding a new file. Two independent triggers, neither requiring a question.

**Who ran:** the full resolved roster again — Primary, Principles, Code review (low), Bugbot
(substituted as general-purpose). No slot was added; the operator named none.

**Diffs:** Primary the rewritten `final-review.diff` (1002 lines measured, under cap); Principles and
Code review (low) their own 353-line delta `git diff 85fae8f HEAD`; Bugbot no diff file.

**A bookkeeping error in this run's own record, disclosed rather than corrected silently:** the
dispatcher recorded `panel-2-CodeReviewLow`'s dispatch end before that slot had reported. The row's
end timestamp is therefore earlier than the slot's real completion, and it carries no agent id. The
slot's findings are recorded in full and its report is captured verbatim; only the timing on that one
row is wrong.

**F15 — recorded, investigated, and not reproduced.** Code review (low) reported the shim-based cases
as intermittently failing, quoting failures such as `unreadable commit count: expected exit 2, got
rc=0 out=MOVED: …`, and attributed them to a race in the shim's `exec`-the-real-git technique. Two
earlier slots this run had seen the same intermittency and set it aside as non-reproducible. The
dispatcher tested it directly rather than accepting either reading:

- 20 consecutive serial runs of `scripts/test-check-base-moved.sh`: **0 failures**.
- 24 runs with six suites executing concurrently — the condition under which a genuine race in a
  shared technique would surface: **0 failures**.
- Deliberately dropping `check-base-moved.sh`'s `COUNT` capture guard and running the suite once
  reproduces the reported failure text **character for character**, including the empty commit count
  in `—  commits on main`.

The cause is therefore established: panel slots were mutating the guard scripts in the shared
worktree while other slots ran the suites against them. It is an artifact of how this run dispatched
concurrent mutating reviewers into one worktree, not a defect in the branch. **The dispatch design
caused it**, which is worth carrying forward: a mutation-testing slot and a suite-running slot must
not share a worktree concurrently.

**Findings raised this pass:** F10 (Principles, Minor), F11–F13 (Primary, Minor — `tasks.md`
baselines stale a second time), F14 (Bugbot, Important — the shim's matcher precision unpinned,
confirmed by the dispatcher), F15 (Code review (low), Important — refuted above).

**Operator decision on convergence.** Three passes found no live defect; every finding has been
coverage or bookkeeping, and each fix round's test-file edits re-trigger Full escalation, which then
surfaces another marginal gap. The dispatcher put that to the operator rather than deviating on its
own judgment. The operator chose: fix F10 and F14, then re-run **targeted** — Primary plus the slots
that raised them — rather than a fourth Full pass. **This is an authorised deviation from the
automatic-escalation rule, recorded here as one.**

## Pass 4 (round 3) — Targeted, by operator authorisation

**Mode:** Targeted, not Full. Three fix rounds had run, which is itself an automatic Full-escalation
trigger, and the operator authorised running targeted instead — recorded under pass 3. Targeted here
means Primary as the integration check plus every slot that raised a finding last pass, which was
all four; the saving is that Principles, Code review (low) and Bugbot read a 294-line delta rather
than the whole branch.

**Findings:** F16 (Primary, Minor — task 1's `Files:` omitted the shim library its commit adds;
repaired by the dispatcher, a planning artifact the fix subagent may not touch), F17 and F18
(Bugbot, both Minor). Principles and Code review (low) both clean. Code review (low) this pass
correctly checked whether the guard scripts differed from HEAD before drawing a conclusion, and
declined to report machine-load noise as a defect — the calibration its previous pass needed.

**Bugbot also raised one Info item, recorded here rather than as a finding:** `--end-of-options` on
`check-base-moved.sh`'s range-argument `rev-list` and `diff` calls is decorative, because the range
always leads with a resolved sha and so can never begin with `-`. That is the same conclusion F9
reached for `resolve-remote-base.sh`, proven this time against a real repository carrying both a
dash-prefixed remote ref and a dash-prefixed local branch. No action taken.

**The handback, and the operator's decisions.** F17 and F18 went back to the operator one at a time,
per the contract's non-convergence handback. F17 is a genuine regress — nothing tests the assertion
helpers, and any test for `assert_shim_fired` rests on an untested assertion one level up; the slot
that raised it said plainly it was not a shippable defect. The dispatcher put that regress to the
operator rather than withdrawing it on its own judgment. The operator chose another round on both.

## Fix rounds 4 and 5

**Round 4** closed F18 (idempotent real-git capture, and the header's overclaim corrected) and F17
(`scripts/test-lib-test-git-shim.sh`, driving `assert_shim_fired` against a prepared and an
unprepared directory, with its own header stating what it does **not** establish so no later reader
mistakes it for having closed the regress).

**Round 5 closed F19, which the dispatcher raised against round 4's own work.** Round 4's report
stated honestly that its new harness did not cover the capture guard, and verification confirmed it:
reverting the idempotency passed all three suites. Held to the same standard as every other finding
this run, that is a surviving mutant, so it was recorded and fixed rather than let through because
the round that produced it had already reported success.

```text
fix-mutation: scripts/lib/test-git-shim.sh — broadened the shim's exact match to a substring match — case 15 failed (3 assertions)
fix-mutation: scripts/test-check-finish-preflight.sh — pointed case 8b's shim at an argument no call carries — the case failed on assert_shim_fired plus 3 pre-existing assertions
fix-mutation: scripts/lib/test-git-shim.sh — moved real-git resolution from source time to per-call — the shim-chaining case failed (2 assertions)
fix-mutation: scripts/lib/test-git-shim.sh — made assert_shim_fired unconditionally pass — test-lib-test-git-shim.sh case 1 failed
fix-mutation: scripts/lib/test-git-shim.sh — reverted the idempotent real-git capture to an unconditional assignment — test-lib-test-git-shim.sh case 3 failed
fix-mutations-total: 5
```

Every one re-run by the dispatcher rather than taken on the fix subagent's report. All three suites
then ran 30 consecutive times with zero failures.

## Closing state

Nineteen findings raised across four passes; all nineteen closed. **No pass found a live defect in
the shipped guards** — every finding was test coverage, a planning-artifact inaccuracy, or, once, a
report that did not reproduce. The guards' logic has been correct since its first commit; what four
fix rounds bought is that it is now pinned, so a later edit that breaks it fails a test instead of
shipping.
