# Self-review context bundle for kan-806-flow-improvement-record-the-rejected-alternative

found: 1 of 7 sources; skipped: 6 of 7 sources
note: RECORDS LOSS — a flow.review-panel stage run completed for kan-806-flow-improvement-record-the-rejected-alternative, but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-806-flow-improvement-record-the-rejected-alternative.md (absent)
skipped: .superpowers/sdd/reviews/kan-806-flow-improvement-record-the-rejected-alternative-panel.md (absent)
skipped: spectre/changes/archive/kan-806-flow-improvement-record-the-rejected-alternative/tasks.md (absent)
skipped: spectre/changes/archive/kan-806-flow-improvement-record-the-rejected-alternative/design.md (absent)
skipped: spectre/changes/archive/kan-806-flow-improvement-record-the-rejected-alternative/narrative.md (absent)

## git log --stat

commit f6c3de8559008cee9ce7904b0b03ccba4420170f
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Wed Oct 7 01:58:10 2026 +0300

    fix(skills): record the rejected alternative even for small decisions
    
    A decision as small as where a file lives is the one a later reader
    re-litigates first, because nothing else in the record explains why the
    obvious alternative was not taken. Size now never exempts a judgment
    call: design.md's Decisions section states the rule, and flow-fast's
    brainstorm stage names each call's rejected alternative and one-line
    reason in its summary, citing that section instead of restating it.
    
    KAN-806

 skills/flow-fast/SKILL.md         | 4 +++-
 skills/flow/brainstorm-planner.md | 6 ++++++
 2 files changed, 9 insertions(+), 1 deletion(-)

## Session narrative

The change adds one rule to two run-loaded sites: size never exempts a decision from recording its rejected alternative and one-line reason — brainstorm-planner.md's Decisions section states it canonically, and flow-fast's brainstorm line names each judgment call's rejected alternative and reason in its summary, citing that section instead of restating it. flow-plan/SKILL.md already required the rejected alternative for routine calls and was deliberately left untouched (rewording it would restate what it already says). Four approaches were tried and abandoned on the way. The first draft appended the new rule to the same sentence run as "**A design that forced no choices records none.**"; check-verbatim-moves then flagged that untouched sentence as reworded, because the appended bold segment changed the guard's sentence segmentation — fixed by giving the rule its own paragraph so the existing sentence stays byte-identical. The first verbatim-moves.txt copied the guard's full FAIL lines including their "FAIL ... ::" prefixes; the guard matches acknowledgement entries against the sentence after "::" only, so the entries were rewritten as bare sentences before the guard went green. The first plan-class run read files=0 and classified small, because the plan's **Files:** paths were not backticked; backticking let the parser see the two documentation files and the class fell to micro (execution inline, no panel). check-plan-shape first rejected the **Tests:** field for opening with an unquoted path; the command names were backticked. Where the run struggled most was the verbatim-moves contract itself — its acknowledgement format is specified as "exactly as the FAIL lines print it after ::" in three places, and the "after ::" tail is load-bearing rather than decoration; the first reading cost one extra guard cycle.
