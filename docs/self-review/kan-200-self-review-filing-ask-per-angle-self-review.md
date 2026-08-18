# Self-review — kan-200-self-review-filing-ask-per-angle

**Rating:** 4/5 — good
**Date:** 2026-08-18 · **Jira:** KAN-200 · **PRs:** #14 (change), #15 (archive)
**Panel:** `light` roster, 4 rounds, 18 findings, all repaired · **Cost:** 25 dispatches, ~2.67M subagent tokens

**This is the first report written under the five-angle rule this change introduced**, and the first
subject `scripts/check-self-review-report.sh` will check.

## Problems and fixes — `myflow-fix`

- **[myflow-fix]** The plan's `Tests:` field named guard scripts inside a parenthetical, and the commit-fields guard parsed them as declared tests (commented on KAN-193) — declined
- **[myflow-fix]** The review panel record was hand-written without its required marker block, and nothing caught it until the unfinished-work gate at integration — filed: KAN-205
- **[myflow-fix]** The delta spec refused to archive after the change had already merged, because its MODIFIED blocks renamed scenarios — filed: KAN-206
- **[myflow-fix]** A requirement in the same capability still said "the four-angle report", where no review slot could see it — filed: KAN-207

## Cost — `myflow-cost`

- **[myflow-cost]** The review loop consumed 67% of the change — ~1.79M of ~2.67M tokens (commented on KAN-201) — declined
- **[myflow-cost]** Full escalation fired on every round because "the fix altered a guard's behaviour" cannot discriminate on a change whose deliverable is a guard — filed: KAN-208

## What went well — `myflow-improvement`

- **[myflow-improvement]** Mutation testing became the panel's standard and the panel enforced it recursively, catching a fix round that had not applied it to itself — filed: KAN-209
- **[myflow-improvement]** Reviewers corrected the parent twice on method rather than on code — a mutation that would have passed by cross-contamination, and a verification that cannot verify — filed: KAN-210
- **[myflow-improvement]** Three prior follow-ups were exercised and all three worked (commented on KAN-192, KAN-195 and KAN-202) — declined

## Automation candidates — `myflow-automation`

- **[myflow-automation]** Emit or validate the panel record's marker block at panel close rather than two stages downstream — filed: KAN-205
- **[myflow-automation]** Run the archive's own check while the delta is still editable, before the PR — filed: KAN-206
- **[myflow-automation]** Collapse the guard's awk-to-bash boundary, which produced three of this change's findings one manifestation at a time — filed: KAN-211

## Stats app and storage — `myflow-stats-app`

- **[myflow-stats-app]** Per-slot, per-round findings are structured data that dies with the panel record (commented on KAN-198) — declined
- **[myflow-stats-app]** Per-dispatch cost is recorded nowhere durable; every figure in this report was scraped from a session transcript — filed: KAN-212
- **[myflow-stats-app]** The guard's declared pre-rule list only shrinks, and nothing observes whether it does — filed: KAN-213

## Notes on this run

**Three of the four problems were mine**, and all three share a shape: an artifact I wrote by hand
passed every gate that ran early and failed a gate that ran late. The plan field failed at the first
task commit, the panel record at integration, the delta spec after the merge. Each is filed as
"check this where it is written, not where it is consumed".

**The panel found what it was supposed to find.** Convergence was 11 → 5 → 2 → 0. The low-effort slot
led on volume for the third change running, but the most consequential finding came from Principles
in round 3 — asking whether the previous fix round had held itself to its own mutation standard. It
had not. That is a question no finding-count metric would credit, and it is recorded on KAN-198 for
exactly that reason.

**Two panel corrections landed on me rather than on the code**: a mutation instruction I wrote would
have passed through cross-contamination and proved nothing, and a grep I used as corroboration
cannot work, because ANSI-C quoting is evaluated at runtime. Both are filed as KAN-210 — the parent
writes every dispatch and nothing else in the loop reads them.

**One count in this run was reported wrong while it was in flight.** The running total was 17
findings; reconstructing the marker block showed 18 — round 1 had 11 distinct findings after
de-duplicating overlaps between slots, not 10. A marker block written as findings arrive cannot
drift that way, which is part of KAN-205's argument.

**A wrinkle in the rule this change just landed:** the contract says a filed issue inherits *every*
label on the linked issue. KAN-200 itself carries `myflow-fix`, so inheriting literally would stamp
`myflow-fix` onto cost and stats-app tickets and corrupt the query the labels exist for. These
filings inherited the non-angle labels only, plus the filing's own angle label. The rule needs that
exception stated.
