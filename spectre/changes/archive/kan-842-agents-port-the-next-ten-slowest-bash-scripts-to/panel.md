# Review panel — kan-842-agents-port-the-next-ten-slowest-bash-scripts-to

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | spectre/changes/kan-842-agents-port-the-next-ten-slowest-bash-scripts-to/design.md:180 | design.md and proposal.md still say the contract-budget guard is ported and that owned-corpus gets a Go twin; the removal is recorded only in tasks.md's task-10 Correction |   |
| F2 | primary | Minor | spectre/changes/kan-842-agents-port-the-next-ten-slowest-bash-scripts-to/design.md:127 | the Measurements parity and After figures were taken at 9a6aa00a, before the merge, the KAN-809 port and the contract-budget removal |   |
| F3 | primary | Minor | stats/internal/guard/installedrules.go:105 | with a relative CHECK_INSTALLED_RULES_HOME the port answers OK where the c5379c0a bash answered STALE; no Correction records this divergence |   |
| F4 | principles | Important | spectre/changes/kan-842-agents-port-the-next-ten-slowest-bash-scripts-to/design.md:180 | design.md's active decisions kan842-helper-twins and scope-ten-next-scripts still claim ownedcorpus.go and the contract-budget port, which tasks.md records as removed |   |
| F5 | principles | Minor | scripts/plan-class.sh:72 | the shim derives and exports FLOW_GUARD_REPO_ROOT, but nothing plan-class runs reads it |   |
| F6 | principles | Minor | stats/internal/guard/taskreviewersingledispatch.go:105 | the FLOW_GUARD_SELF to sibling-directory resolution is written twice, so the KNOWN-BUGS lexical-clean defect must be fixed twice |   |
| F7 | principles | Minor | stats/internal/guard/taskreviewersingledispatch.go:304 | the bash exported LC_ALL=C to plan-dispatch-*.sh; the port runs them under the caller's locale |   |

findings-total: 7
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 withdrawn deliberate divergence: the port follows the header contract where the c5379c0a bash had a readlink-after-cd bug; recorded as a task-7 Correction
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed

reproducers-total: 7
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-principles-2.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-principles-3.sh
finding-reproducer: F7 .superpowers/sdd/reproducers/0-principles-4.sh

## Pass log

### Round 0

- roster: compact — primary+principles, one dispatch opus/high (decision)
- no addition this round — the resolved list ran alone
- diff size: 15268 changed lines, over cap — proceeded unasked
- docs-only: exit 1 — first non-doc path scripts/check-contract-budget.sh; roster unchanged
- base: CLEAR at 4a278320 after operator-chosen merge of origin/main (d09aaaa2)
- standards: CLAUDE.md, AGENTS.md

### Round 1

- fix round 1: F1+F4 (one defect) fixed inline by the parent in design.md/proposal.md; F2,F3,F5,F6,F7 deferred to KNOWN-BUGS.md per the operator rule; re-run primary and principles alone on opus/low, reading fix-round-1.diff
- round 1: base CLEAR at 675a55a94674a0fbe59f998e060b2491940d4b24 after merging origin/main; re-run primary (F1) and principles (F4), each alone on opus/low, reading fix-round-1.diff (c1374eb45da3b3fb45fd214a61b6cb787e5e88fb..HEAD^1)
- round 1 re-runs: primary clean, principles clean — no new finding
fix-mutation: spectre/changes/kan-842-agents-port-the-next-ten-slowest-bash-scripts-to/design.md — none — doc-only fix; the reproducers 0-primary-1.sh and 0-principles-1.sh flipped demonstrated→not-demonstrated
fix-mutation: stats/internal/guard/taskreviewersingledispatch.go — dropped the LC_ALL=C pin — TestCheckTaskReviewerSingleDispatch/plan-dispatch siblings run under LC_ALL=C
fix-mutation: stats/internal/guard/taskreviewersingledispatch.go — guardSelfDir returns the parent directory — TestCheckTaskReviewerSingleDispatch, TestCheckVisualVerifyDispatched
fix-mutation: scripts/plan-class.sh — none — nothing reads the removed lines
fix-mutations-total: 4
