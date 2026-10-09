# The in-run fix loop

Loaded by the load directive under **Verify** (`skills/flow/verify-and-handoff.md`) only when the
final report of `flow.verify` or `flow.visual-verify` carries a fixable defect. Every "you" here
addresses the parent (**The parent orchestrates directly**, `skills/flow/implement.md`).

## The loop

**A fixable defect the verify stages find is fixed by this run, never handed to the operator as a
question.** The fixable defects are a lint or test failure **Inline verify**
(`skills/flow/verify-and-handoff.md`) routes here, and every item the **Blocking** paragraph of
**Steps 3–13** (`skills/flow/visual-verify.md`) routes here. The choice "fix it in-run, or hand
off with it open?" is never asked: the run takes **fix it in-run** under **Auto-resolution**
(`skills/flow-contracts/operator-prompts.md`), recorded and named on the handoff's
`**Decisions:**` lines as that section states, one entry per defect, the defect named as the
report names it.

One round runs these steps in order:

1. **Transcribe the fix.** Run **Documenting a fix, before implementing it**
   (`skills/flow/document-fix.md`) with the stage's final `## Report` — every defect it names,
   verbatim — as the fix instructions. Neither of its two asks is made: the fix is appended to
   `proposal.md` and `tasks.md`, and `**Tasks appended:**` is not raised, since it counts appends
   made at the human gate. The appended task carries the field family the verification-change
   paragraph of **D. Basic Workflow #3 — Writing plans** (`skills/flow/brainstorm-planner.md`)
   states for a found defect, and its title reads `In-run fix <k> — <stage>: <the defects>`, `<k>`
   this stage's round in this run. Its planning commit and the re-decide run as that file states.
   **One round appends one task**, carrying every defect the report names — never a task or a
   round per defect.
2. **Implement it.** **4. Execute (SDD + TDD)** (`skills/flow/implement.md`) executes the appended
   task exactly as a plan-time task: the decision's implementer pair, `close-task.sh` with its
   guard and its review gate. The implementer's dispatch also carries the MUTATION PROOF
   paragraph (`skills/flow/review-panel.md`), and the round closes only when every reported
   `fix-mutation:` line is checked against the fix diff the way **The fix round mutation-proves
   what it changed** (`skills/flow/review-panel-fix-round.md`) walks them — a mutant nothing
   kills is a new fixable defect this loop takes another round on.
3. **Re-review the delta.** Open `flow.review-panel` again and run one dispatch: `primary` alone,
   reading `late-fix.diff` over the range since the panel's last clean close, written, paired,
   ceilinged and recorded as **The late-fix reduction** (`skills/flow/review-panel-late-fix.md`)
   states its dispatch, its prompt carrying the FIX-ROUND SCOPE paragraph of **Panel re-runs**
   (`skills/flow/review-panel-fix-round.md`). That file's finding, voiding and staleness rules
   apply as written — a clean or Minor-only read leaves every slot it did not dispatch current.
   `check-late-fix-trigger.sh` is not run: its scope-growth condition denies every appended task the `primary`-alone reduction,
   and this task's premise is a defect a verifier measured, not an operator's unverified flag.
   The stage then closes through its own close (**Review panel**, `skills/flow/review-panel.md`).
4. **Re-capture and re-run.** Run **Verify** (`skills/flow/verify-and-handoff.md`) again from its
   `begin` mark, then **Visual verification** from its `begin` mark, so the fix's own lint, tests and
   captures are what the handoff reports. This round's `verify` and `visual-verify` dispatch keys
   carry the suffix `-fix-<k>`, before any `-<worktree basename>` suffix. **The `visual-verify`
   re-run is scoped to the round's own diff** in every worktree with a round base — the HEAD step
   12 of **Steps 3–13** (`skills/flow/visual-verify.md`) read after that worktree's previous
   verifier dispatch: its prompt names that base, which the verifier's step 7 substitutes for
   `specs`' `<merge-base>`, and its `motions:` names only the motions `git diff <round
   base>..HEAD` adds, changes or removes, plus every motion this round's defects name. A worktree
   whose round diff is empty is not dispatched. A worktree with no round base — no verifier of
   this run has verified it, or the parent no longer holds the sha — runs the stage unscoped.
   The confirming re-run
   verifies a fresh build, never a stale stack: before this round's re-run reports the fix
   confirmed live, rebuild the stack from the branch HEAD — current source through the project's
   own build commands, brought up the way step 5 of **Steps 3–13** (`skills/flow/visual-verify.md`)
   starts one — and re-run the defect's exact scenario end to end against that build. The
   fingerprint step there proves only that a stack serves its worktree's build, never that the
   build is current source, and verification against a stale build is not verification. The
   rebuild is itself a defect-finder: a defect the fresh build exposes that the fix's own task
   does not cover is reported the way any stage-found defect is, never absorbed into the fix's
   confirmation.

**One full run closes the loop.** A scoped re-run whose final report carries no fixable defect is
followed by **Visual verification** once more from its `begin` mark, unscoped — no round base,
every motion of the change — its dispatch keys suffixed `-fix-<k>-full` before any `-<worktree
basename>` suffix. A fixable defect it reports opens the next round; its clean report is the one the
handoff reports.

## No cap

**No round count ends the loop.** **Fewest operator actions** (`skills/flow-contracts/pipeline.md`)
is canonical for that, and for a round that makes no progress.
