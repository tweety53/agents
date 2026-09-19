# Self-review context bundle for kan-593-panel-preflight-bundle-build-fails-loudly

found: 2 of 6 sources; skipped: 4 of 6 sources
skipped: spectre/changes/archive/kan-593-panel-preflight-bundle-build-fails-loudly/tasks.md (absent)
skipped: spectre/changes/archive/kan-593-panel-preflight-bundle-build-fails-loudly/design.md (absent)
skipped: spectre/changes/archive/kan-593-panel-preflight-bundle-build-fails-loudly/narrative.md (absent)
skipped: git log --stat (absent)

## .superpowers/sdd/ledgers/kan-593-panel-preflight-bundle-build-fails-loudly.md

# SDD ledger — kan-593-panel-preflight-bundle-build-fails-loudly

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: a8605db6990c8a40bfaea933d583e61851aec5c5
- Outcome: completed
- Started: 2026-09-19T19:31:21Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-19T19:52:26Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-593-panel-preflight-bundle-build-fails-loudly-panel.md

# Review panel — kan-593-panel-preflight-bundle-build-fails-loudly

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | .superpowers/sdd/kan-593-panel-preflight-bundle-build-fails-loudly/tasks.md:47 | Task 4 names budget rows for skills/flow/review-panel.md AND scripts/check-dispatch-paragraphs.sh, but the budget guard covers owned .md/.mdc files only - the .sh row can never exist; the diff correctly raises one row while the task field no longer matches it and nothing records the deviation |   |
| F2 | primary | Minor | scripts/check-contract-budget.sh:150 | Budget row 109572 matches no documented derivation: the guard's rule is landed size plus 25 percent, and the headroom measured 24.1 percent |   |
| F3 | principles | Minor | scripts/check-contract-budget.sh:150 | Same defect as primary F2: declared row 109572 does not equal landed size plus 25 percent, violating the derivation the guard's own header and .flow/project.md state |   |

findings-total: 3
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed

reproducers-total: 3
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-principles-1.sh

## Pass log

### Round 0

- roster: compact — 0
- diff size: 262 lines, under cap — proceeding
- docs-only: exit 1 — first non-documentation path scripts/check-contract-budget.sh; resolved roster runs unchanged

### Round 1

- principles: not re-run — raised a Minor only (F3), fixed inline; a Minor triggers no re-run
- round close: all three reproducers exit 0; primary re-run verdicts fixed/fixed, no new defects; fix diff touches scripts/check-contract-budget.sh, the finding-named path

## Session narrative

This `/flow-fast` run fixed KAN-593: a failed dispatch-context-bundle build during the review panel's pre-flight now stops or takes an explicit operator override, never a silent downgrade. All three toggles resolved `dynamic`, so the run wrote its plan to `.superpowers/sdd/`, classified it (`small`, compact panel), and executed inline. The contract change was one labeled blockquote in `skills/flow/review-panel.md`; it is pinned by a new `check-dispatch-paragraphs.sh` entry (24th site) with six new harness cases (78–83), and the `check-contract-budget.sh` ratchet row moved deliberately. The compact panel (primary+principles, one bundled dispatch) raised one Important — the plan's task 4 named a budget row the `.md`/`.mdc`-only guard can never carry — and two Minors sharing one defect (the row's 109572 value came from a stale pre-rebase size). The base moved twice mid-run and rebased clean both times; the second rebase is what grew `review-panel.md` and invalidated the row. The fix round corrected the plan text and recomputed the row from the landed size (110363 = ceil(88290 × 1.25)); the primary re-run verified both fixes with all three reproducers flipping 1→0. Where this run struggled: the first lint pass caught a rootless path citation in the new paragraph (fixed by pointing at the rebuild's output path in prose, harness fixtures re-synced), and three store-write shapes had to be learned from the CLI's own usage (`record finding` takes no `-session-token`, no `-theme`, and uses `-location`/`-ref`/`-status`). Verification: the full lint list green, the guard suite 75/75, both guard mutations proven to bite.
