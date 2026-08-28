# Review panel — kan-217-myflow-forward-the-reviewer-s-verbatim-report

The findings table and the marker block are rendered from the store by `flow record render`.
This file is the pass log.

## Pass 1 (round 0) — the resolved roster

Mode: initial panel. Roster resolved from the settings store: `primary`, `principles`,
`code-review-low`, `bugbot`. No addition this round — the resolved list ran alone.

`bugbot`'s own agent type is not offered by this harness, so it was dispatched as a
**general-purpose substitute** carrying the Bugbot mutation-testing brief, per **An unspawnable id
is substituted, not skipped**. Its `-model` records the model actually given (`sonnet`), never
`unknown (agent-defined)`.

`check-panel-diff-size.sh` exited 0 — under cap. Diff: 1205 lines over 8 files, merge base
`8875344`. Every slot read `.superpowers/sdd/final-review.diff`.

Each slot's report was written verbatim to `.superpowers/sdd/panel-report-0-<id>.md` — this change's
own mechanism, applied to the run that reviewed it.

Findings raised: 8. Primary raised none.

## Pass 2 (round 1) — fix round

One fix subagent, `sonnet`, given all eight findings as a structured block plus the four verbatim
reports as fact. It was told explicitly that Primary and Code review (low) disagreed on F3 and that
it had to resolve the disagreement itself rather than adopt either reading.

Fixups folded with `git commit --fixup` + `git rebase --autosquash`: F1 into task 2's commit, F2–F8
into task 3's.

### The fix round's mutation proof

Every executable behaviour the fix changed, mutated by the parent — not by the fix subagent — with
the fix restored afterwards and `git diff` confirmed empty each time.

fix-mutation: `scripts/check-dispatch-paragraphs.sh` — folded the set-but-empty root refusal into the default-root branch — case 12 failed
fix-mutation: `scripts/check-dispatch-paragraphs.sh` — removed the not-a-regular-file refusal — case 13 failed
fix-mutation: `scripts/check-dispatch-paragraphs.sh` — removed the not-readable refusal — case 14 failed
fix-mutation: `scripts/check-dispatch-paragraphs.sh` — lowered the review-panel.md REPRODUCE site's min-blocks from 1 to 0 — case 2's min-blocks assertion failed
fix-mutation: `scripts/check-dispatch-paragraphs.sh` — lowered the implement.md site's min-blocks from 2 to 1 — case 4's min-blocks assertion failed
fix-mutation: `scripts/check-dispatch-paragraphs.sh` — reverted the label-boundary stop in `extract_block_text` — case 15 failed
fix-mutation: `skills/flow-contracts/artifacts-registry.md` — none — the deletion of a restating sentence is not an executable behaviour
fix-mutation: `scripts/test-check-dispatch-paragraphs.sh` — none — a comment rewrite is not an executable behaviour
fix-mutation: `scripts/test-check-plan-shape.sh` — none — removing a false citation from a comment is not an executable behaviour
fix-mutations-total: 9

No surviving mutant. The harness grew 11 cases / 22 assertions to 15 cases / 32 assertions; no case
or assertion was removed or weakened, so no hunk needs a covering test named for it.

### F2 did not close

F2's own reproducer, re-run against the fixed tree, still exits 0. The fix subagent stopped the
continuation walk at a line carrying a new known label, which closes two required blocks glued
together by a missing blank line — a different fixture from the one that raised the finding. It
stated the original case out of scope and verified it still passes clean; the parent confirmed that
independently.

Handed back to the operator rather than withdrawn: only an operator withdrawal records a reason.

## Pass 3 (round 2) — F2 alone

Targeted. One fix subagent, `sonnet`, given F2 alone, its verbatim report, the round-1 bounce
against the same defect identity, and an explicit statement that "this cannot be closed cleanly" was
an acceptable outcome of the round.

It changed nothing and reported that no Markdown-structural signal separates a long legitimate
wrapped paragraph from one diluted with foreign prose, because in F2's fixture they are
structurally the same thing. It enumerated four candidate tightenings — a bare `>` separator line, a
sentence-boundary split, a line-count cap, and full literal-paragraph matching — and rejected each
with its own false-rejection risk rather than shipping one.

The parent verified the load-bearing claim rather than accepting it: `implement.md`'s real
REPRODUCE, DON'T READ paragraph carries `exercise the real thing` in its first sentence and
`a fake or a hand-built value` in its second, so a sentence-boundary rule would have to span
sentences anyway — which is most of what makes the bleed possible.

The operator withdrew F2 with a reason. No executable behaviour changed this round, so there was
nothing to mutation-prove.

fix-mutations-total: 0
