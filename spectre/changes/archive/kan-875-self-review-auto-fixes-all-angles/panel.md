# Review panel — kan-875-self-review-auto-fixes-all-angles

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | skills/flow-self-review/SKILL.md:81 | Step 3 runs git -C <agents repo> against origin/<default-branch>, but step 1 defines <default-branch> as the project's own default branch, so a project on master or develop points at a branch the agents repo lacks. |   |
| F2 | primary | Minor | spectre/changes/kan-875-self-review-auto-fixes-all-angles/design.md:92 | The measured live-verification line quotes the warning text from d34ba4aa, which a later commit reworded. |   |
| F3 | primary | Minor | skills/flow-self-review/SKILL.md:14 | 'the one dispatch is the fix branch's reviewer' contradicts a fresh reviewer per round; design.md:31 says two dispatches on a clean pass where inline fixing gives one. |   |
| F4 | primary | Minor | skills/flow-self-review/SKILL.md:13 | design says opus throughout, but inline fixes run on whatever model the session uses, so a sonnet session lands code unattended. |   |
| F5 | primary | Minor | skills/flow-self-review/SKILL.md:79 | Step 3 does not say whether review fixes amend the per-finding commits or add new ones, leaving the recorded sha ambiguous and a dropped commit's follow-ups behind. |   |
| F6 | primary | Minor | commands-claude/flow-self-review.md:10 | commands-claude/flow-self-review.md, CLAUDE.md:66 and AGENTS.md:83 still describe a pass that only files and rates. |   |
| F7 | principles | Minor | stats/internal/store/selfreviewfindings.go:20 | The store copies the sha and issue-key regexes from the report guard, so the two can drift apart. |   |
| F8 | principles | Minor | stats/internal/store/migrations/0035_self_review_findings.sql:15 | No uniqueness constraint: a pass interrupted after step 6 and re-run records every finding twice; design.md does not state the tradeoff. |   |
| F9 | primary | Important | skills/flow-self-review/SKILL.md:116 | the Land item's push target is the project's <default-branch>, not <agents-base> |   |
| F10 | primary | Important | skills/flow-self-review/SKILL.md:88 | step 3 verifies a fresh worktree without <agents repo>'s ## worktree setup, so Verify is red on every pass |   |
| F11 | primary | Minor | spectre/changes/kan-875-self-review-auto-fixes-all-angles/design.md:105 | stale <default-branch> for the fixed sha in fixed-disposition and the Review Focus |   |
| F12 | principles | Important | skills/flow-self-review/SKILL.md:85 | the agents repo's fix-loop base branch is defined in the consumer, and the shared recipe it delegates to says otherwise |   |

findings-total: 12
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed

reproducers-total: 12
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 none — prose consistency
finding-reproducer: F4 none — policy text gap
finding-reproducer: F5 none — procedure gap
finding-reproducer: F6 none — description text
finding-reproducer: F7 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F8 none — needs a live database write sequence
finding-reproducer: F9 .superpowers/sdd/reproducers/2-primary-1.sh
finding-reproducer: F10 .superpowers/sdd/reproducers/2-primary-2.sh
finding-reproducer: F11 .superpowers/sdd/reproducers/2-primary-3.sh
finding-reproducer: F12 .superpowers/sdd/reproducers/2-principles-1.sh

## Pass log

### Round 0

- diff size: 1464 lines, under cap — proceed
- docs-only: exit 1 (scripts/check-self-review-report.sh) — roster primary+principles unchanged
- roster: compact — 79
- no addition this round — the resolved list ran alone.
- standards: CLAUDE.md
- roster: compact — 79
- diff size: 1748 lines, under cap, proceed
- docs-only: exit 1 (first non-doc path scripts/check-self-review-report.sh) — roster primary+principles dispatched
- no addition this round — the resolved list ran alone.
- late-fix reduction: not fired — condition 1, no earlier clean close (fix run)
- standards: CLAUDE.md, AGENTS.md (per project-get standards)
- base: rebased automatically onto 8dad0fc9 (no overlap)

### Round 1

- fix round 1: inline by the parent (fixer skipped — inline); diff .superpowers/sdd/fix-round-1.diff; re-run primary and principles, each alone on the rerun pair
fix-mutation: stats/internal/records/types.go — SelfReviewSha widened to match anything — TestCheckSelfReviewReport case 36; TestSelfReviewFindingRefusesBadDisposition
fix-mutation: skills/flow-self-review/SKILL.md — none — prose or comment only — no executable behaviour
fix-mutation: commands-claude/flow-self-review.md — none — prose or comment only — no executable behaviour
fix-mutation: CLAUDE.md — none — prose or comment only — no executable behaviour
fix-mutation: AGENTS.md — none — prose or comment only — no executable behaviour
fix-mutation: stats/internal/store/migrations/0035_self_review_findings.sql — none — prose or comment only — no executable behaviour
fix-mutation: spectre/changes/kan-875-self-review-auto-fixes-all-angles/design.md — none — prose or comment only — no executable behaviour
fix-mutations-total: 7

### Round 2

- fix-run panel (rounds 0-1 belong to the first run; round-0 notes timestamped now were meant for this round)
- roster: compact — 79
- diff size: 1748 lines, under cap, proceed
- docs-only: exit 1 (first non-doc path scripts/check-self-review-report.sh) — roster primary+principles dispatched
- no addition this round — the resolved list ran alone.
- late-fix reduction: not fired — condition 1
- base: rebased automatically onto 8dad0fc9 (no overlap)
- reproducer declarations for F10-F12 reformatted to <path>:<line>:<content> by the parent (format only, checks unchanged) instead of a bounce — each still demonstrates
fix-mutation: skills/flow-self-review/SKILL.md — none — skill/contract prose; no executable behaviour changed
fix-mutations-total: 1

### Round 3

- re-runs: primary (F9-F11) and principles (F12), each alone on opus/low over fix-round-3.diff — all fixed, clean
