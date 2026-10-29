# Review panel — kan-1045-flow-one-allowed-model-effort-list-per-subtask

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Round | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|---|
| F1 | 0 | primary | Critical | scripts/check-dispatch-pair.sh:21 | run-guard-tests.sh refuses: check-dispatch-pair.sh has no companion test — its tests live in modelpolicy_test.go, not check_dispatch_pair_test.go |   |
| F2 | 0 | primary | Important | skills/flow-contracts/model-policy.md:96 | the verify fix loop's no-progress escalation has no allowed role: implementer has no marked opus/flow-high row |   |
| F3 | 0 | primary | Important | skills/flow-self-review/SKILL.md:341 | a loop planner's Explore and Plan dispatches are refused: PLANNING DELEGATION never names the Role line |   |
| F4 | 0 | principles | Important | hooks/enforce-agent-baseline.py:120 | the hook's ZCode branch is a second runtime parser of the allowed-pairs table |   |
| F5 | 0 | principles | Minor | skills/flow-contracts/model-policy.md:96 | six table rows serve no dispatch site (implementer-group, gated-fix, visual-verify-fix) |   |
| F6 | 0 | call-efficiency | Important | hooks/enforce-agent-baseline.py:303 | dispatch_pair_reason re-runs in_flow_run, reading the transcript and the store twice per dispatch |   |
| F7 | 0 | mutation | Important | stats/internal/guard/modelpolicy.go:64 | no test catches removing the malformed-row validation |   |
| F8 | 0 | mutation | Important | hooks/enforce-agent-baseline.py:163 | no test catches the guard's exit 2 being let through |   |

findings-total: 8
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed

reproducers-total: 8
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-principles-2.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-test-sanity+call-efficiency-1.sh
finding-reproducer: F7 .superpowers/sdd/reproducers/0-mutation-1.sh
finding-reproducer: F8 .superpowers/sdd/reproducers/0-mutation-2.sh

## Pass log

### Round 0

- roster: compact — 61
- diff size: 1632 lines, under cap — proceed
- docs-only: exit 1 — hooks/enforce-agent-baseline.py; roster primary+principles plus both extra dispatches
- no addition this round — the resolved list ran alone.
- standards: CLAUDE.md, AGENTS.md
- context bundle: rebuilt
- dispatched: panel-0-primary+principles on opus/medium; panel-0-test-sanity+call-efficiency on opus/low
- dispatched: panel-0-mutation+failure-modes on opus/low
- merger: not dispatched — every slot handed back its findings summary directly; reports read for verdicts only
- verdict: fix (1 Critical, 6 Important)

### Round 1

- dispatched: panel-1-primary+principles+call-efficiency on sonnet/low over 49794eb4f7c513933a5fe278a966ac77792b9c7f; panel-1-mutation on sonnet/low over 49794eb4f7c513933a5fe278a966ac77792b9c7f
- verdict: clean
fix-mutation: stats/internal/guard/modelpolicy.go — malformed-row check replaced by if false — TestPolicyTableRefusesMalformedRows
fix-mutation: hooks/enforce-agent-baseline.py — deny only on exit 1 — TestEnforceAgentBaselineHookRefusesOffPolicyDispatch
fix-mutations-total: 2
