# KAN-464 — cut review-panel cost: delta re-run scope, fix subagent owns the mutation-proof

Brainstorming record for `kan-464-flow-cut-review-panel-cost-delta-re-run-scope`, 2026-09-06.
The mechanism this round settled on is canonical in the change's own
`spectre/changes/kan-464-flow-cut-review-panel-cost-delta-re-run-scope/design.md`; this file
records how the round got there — what was measured, what was asked, and which options were
rejected.

## Where the brief came from

`docs/superpowers/research/flow-gymie-implementation-speedup.md`, thread 2. The note was not a
staged seed for this change: `skills/flow/brainstorm-planner.md`'s discovery rule looks for
`kan-464.md` and then `kan-464-*.md` under `docs/superpowers/research/`, and this note matches
neither, so it was read as the brief and left in place — thread 1 is still unconsumed. Thread 3
was consumed by KAN-465 in a previous run.

The measured shape, across seven gymie changes (kan-420 excluded as an overnight run):

- 16–83 minutes of panel per change, median 32, roughly 1000 panel-minutes in total.
- All seven ran a fix round; none closed on pass 1.
- `mutation` is the long pole of pass 1 every time (6–11 min) and raised 9 of the 33 findings.
- The delta re-run of all four slots cost 1–6 minutes per change and produced, across all seven,
  four stale `tasks.md` fields and three reviewer-labelled nits.
- Conductor glue cost 5–25 minutes per panel, including one 7-minute stretch in kan-455 where the
  conductor ran `:shared:desktopTest` four times on a 200k+ token context to mutation-prove one
  fix, and a ten-minute operator wait on refused reproducers the conductor then wrote itself.
- 11 of 33 findings carried a runnable reproducer; the rest were `none — …` because the reviewer
  stopped at a `grep` pipeline `check-panel-reproducers.sh` refuses.

## What the round checked before designing

The issue's own wording is Gradle-flavoured and generic. The current
`skills/flow/review-panel.md` was read end to end as ground truth, along with
`scripts/run-reproducer.sh`, `scripts/check-panel-reproducers.sh`,
`scripts/check-dispatch-paragraphs.sh`, `scripts/check-contract-budget.sh` and
`skills/flow/engineering-principles.md`. Three facts came out of that reading and shaped the
design:

1. **Lever C really is prompt-only.** `run-reproducer.sh` refuses an absolute path, a `..`
   segment, a leading `-`, a URL and a set of shell metacharacters, then requires the resolved
   path to be a regular file inside the worktree with execute permission. A relative
   `.superpowers/sdd/reproducers/…sh` satisfies every one of those. `check-panel-reproducers.sh`
   is lexical only and does not even check existence. Neither guard needs an edit — the note's
   claim held up against the scripts.
2. **Lever A needs a second edit the issue does not name.** **Panel re-runs**' stale-result
   definition already fails a slot's clean result when "any commit or working-tree change to
   source landed after that slot's last read", carved out today only for Bugbot and Mutation.
   Narrowing which slots re-run without generalising that carve-out would leave the handoff
   blocked on precisely the slots the new rule declines to re-run.
3. **Lever B's plan-field half is already written.** The PLAN FIELDS paragraph exists on the
   fix-subagent dispatch; kan-454 and kan-455 show it being ignored, not missing. What actually
   moves is the mutation execution.

## Round 1 — the reproducer filename

The research note proposes `<abs-worktree>/.superpowers/sdd/reproducers/<ref>.sh`. A reviewer
slot cannot write that path: `F<n>` refs are assigned by the dispatcher at `flow record finding`
time, after every slot has reported. Three options were put to the operator:

- **Slot-and-round name, no rename** — the prompt fixes
  `.superpowers/sdd/reproducers/<round>-<id>-<n>.sh`, unique by construction, recorded verbatim.
- **Parent renames to `<ref>.sh`** — matches the note's literal text at the cost of a rename plus
  a rewritten reproducer field per finding, which is the conductor glue this change exists to cut.
- **Free-form path under `reproducers/`** — least to specify, but silently collides when two slots
  choose the same name, since every slot writes into the one canonical worktree.

**Chosen: slot-and-round name, no rename.**

## Round 2 — what makes a slot re-run

"The slot(s) whose findings were actually touched by the fix" and the clause already in the file
for Bugbot/Mutation/Security ("raised a finding in the previous round or the previous round raised
a new Critical") are not the same set: they differ when a finding is withdrawn rather than fixed.
Three options:

- **Reuse the existing clause verbatim, for every slot** — lever A becomes a deletion plus a
  widening, with no second rule to keep in sync. A slot whose sole finding the operator withdrew
  re-runs once unnecessarily.
- **Only slots whose finding the round actually fixed** — matches the note's wording exactly, at
  the cost of an extra qualifier and a case to reason about.
- **Fixed findings only, no Critical escape hatch** — cheapest panel; removes the one trigger that
  makes an unrelated slot re-read a branch after a severe defect was patched.

**Chosen: reuse the existing clause verbatim, for every slot.**

## Decided without asking

**The MUTATION PROOF paragraph gets a row in `check-dispatch-paragraphs.sh`.** That guard exists
because KAN-289's root cause was an instruction living in no table, depending on the dispatcher
remembering to type it; KAN-217, KAN-263 and KAN-441 each added a row rather than a second
mechanism. With lever A cutting re-runs, the fix round's own proof is the only mutation reasoning
some runs will do at all, so it is the last paragraph that should be trimmable in silence. The
established pattern was followed rather than a new question raised.

## Not in scope

The research note's own open list — a files-per-task cap, an implementer failure budget, what a
task whose `**Tests:**` is `none` should target, lint-loop bounding, abbreviated `**Files:**`
paths, and `flow.visual-verify`'s unrecorded screenshot agent — is untouched here. So is thread 1
(the serial implementer loop), which remains staged in the research note.
