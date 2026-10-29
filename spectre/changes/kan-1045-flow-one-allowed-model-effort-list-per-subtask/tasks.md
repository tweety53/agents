# kan-1045-flow-one-allowed-model-effort-list-per-subtask

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

Role ids and rows (the table task 1 writes; `—` = no marker):

| Role | Allowed pairs | Requires marker |
|---|---|---|
| `explorer` | `opus/Explore@low`, `opus/Explore@medium` | — |
| `plan-designer` | `opus/Plan@low`, `opus/Plan@medium` | — |
| `implementer`, `implementer-group`, `panel-fix`, `gated-fix`, `visual-verify-fix` | `opus/flow-low`, `opus/flow-medium`, `sonnet/flow-low`, `sonnet/flow-medium` | — |
| `panel-fix`, `gated-fix`, `visual-verify-fix` | `opus/flow-high` | `No-progress escalation` |
| `panel-reviewer` | `opus/flow-low`, `opus/flow-medium` | — |
| `panel-reviewer` | `opus/flow-high` | `Operator-approved high` |
| `panel-extra`, `panel-rerun` | `opus/flow-low`, `sonnet/flow-low` | — |
| `gated-reviewer` | `opus/flow-medium` | — |
| `gated-re-reviewer`, `visual-verifier`, `late-fix` | `opus/flow-low` | — |
| `panel-merger` | `sonnet/flow-low` | — |
| `tooling-analyst`, `self-review-pass` | `opus/flow-medium` | — |
| `self-review-plan`, `pipeline-fix-plan` | `opus/flow-planner-medium`, `opus/flow-planner-low` | — |
| `self-review-fix`, `pipeline-fix-fix` | `opus/flow-low` | — |
| `self-review-fix`, `pipeline-fix-fix` | `opus/flow-high` | `No-progress escalation` |
| `self-review-review-1`, `pipeline-fix-review-1` | `opus/flow-medium`, `opus/flow-low` | — |
| `self-review-review-r`, `pipeline-fix-review-r` | `sonnet/flow-low` | — |

Before writing a fixer row's values, confirm them against `skills/flow/gated-review-fix.md` and `skills/flow/visual-verify.md`; a contract stating a narrower pair wins and the row takes it. <!-- measured: grep -n opus skills/flow/gated-review-fix.md skills/flow/verify-fix-loop.md @ 122c3858 — the gated fix runs inline on the parent and a verify fix on the decision pair, so neither states a narrower pair -->

- [x] 1. The **Allowed pairs** table, its parser and the `check-dispatch-pair` guard
  - [x] **Step 1: Add `## Allowed pairs` to `skills/flow-contracts/model-policy.md` after `## Model policy`: one sentence that it is canonical for which `{model, effort}` every flow dispatch may run on, that a role row with a marker admits its pairs only when the prompt carries the marker as a whole line, and that a pair is `` `<model>/<subagent_type>` `` or `` `<model>/<Explore|Plan>@<effort>` ``; then the table above, one row per role id (a cell listing several ids is split, one row each).**
  - [x] **Step 2: Write `stats/internal/guard/modelpolicy.go`: `policyTable(repo string) (map[string][]policyRow, error)` parsing the `| Role | Allowed pairs | Requires marker |` table (backticked tokens per cell, as `loop.go`'s `lpTable` does; a missing table or a malformed row is an error); `policyAllows(rows map[string][]policyRow, role, typ, model, effort, prompt string) error` — `nil` when an unmarked row of the role lists the pair, or a marked row lists it and its marker is a whole line of `prompt`; otherwise an error naming the role's allowed pairs. Register `check-dispatch-pair` in `Registry` from `init()`: args `<agents-repo> <subagent_type> <model> [<effort>]`, prompt on stdin, role from the first line matching `^Role: (\S+)$`; exit 0 printing `DISPATCH-PAIR-OK <role>`, exit 1 printing `DISPATCH-PAIR-REFUSED: <why>` (no `Role:` line, unknown role, pair not allowed, marker missing), exit 2 on usage error or unreadable table.**
  - [x] **Step 3: Add `scripts/check-dispatch-pair.sh`, the three-line shim `flow_guard_exec check-dispatch-pair 2 "check-dispatch-pair:" "$@"`, shaped as `scripts/check-loop-dispatch.sh`.**
  - [x] **Step 4: Tests in `stats/internal/guard/modelpolicy_test.go`, all against the real repo file: `TestPolicyTableParsesRepo` (every role id above present), `TestPolicyAllowsMarkedRowOnlyWithMarker` (`panel-fix` on `opus/flow-high` refused without and admitted with `No-progress escalation`), `TestPolicyTableNamesNoBannedPair` (no row names `haiku`, `fable` or an `xhigh` type), `TestCheckDispatchPairExitCodes` (0 on an allowed pair, 1 on no `Role:` line, unknown role and wrong pair, 2 on a missing repo). Verify: `cd stats && gofmt -l . && go vet ./internal/guard/... && go test ./internal/guard/ -run 'TestPolicy|TestCheckDispatchPair' -count=1`.**
**Build:** green
**Files:** `skills/flow-contracts/model-policy.md`, `stats/internal/guard/modelpolicy.go`, `stats/internal/guard/modelpolicy_test.go`, `scripts/check-dispatch-pair.sh`
**Allowed-collateral:** `skills/*/scripts/check-dispatch-pair.sh`
**Tests:** `TestPolicyTableParsesRepo`, `TestPolicyAllowsMarkedRowOnlyWithMarker`, `TestPolicyTableNamesNoBannedPair`, `TestCheckDispatchPairExitCodes`
**Regression:** reverting drops the table and the guard: all four tests fail (no table to parse, no guard registered).
**Baseline:** before=0 after=4 <!-- measured: cat stats/internal/guard/modelpolicy_test.go 2>/dev/null | grep -c '^func Test' @ 122c3858 -->
**Commit:** feat(guard): one allowed-pairs table per dispatch role and its check-dispatch-pair guard
**After:** none

**Decision:** policy-table-in-model-policy

**Decision:** escalation-as-marked-rows

- [x] 2. `check-loop-dispatch` reads the table; the self-review pairs table goes
  - [x] **Step 1: In `stats/internal/guard/loop.go` delete `lpSkill`, `lpRow` and `lpTable`; map the loop key's suffix to a role id (`pass`→`-pass`, `plan`/`plan-<n>`→`-plan`, `fix`/`fix-<r>`/`fix-g<g>`/`fix-v<v>`→`-fix`, `review-1`→`-review-1`, `review-<r>`→`-review-r`, prefixed `self-review` or `pipeline-fix`), and replace the PAIR check with `policyAllows` from task 1, passing the prompt the hook hands in; keep the ESCALATION check (a `flow-high` fix needs a completed earlier round) and the review-1 `**Chosen:**` effort check unchanged.**
  - [x] **Step 2: In `skills/flow-self-review/SKILL.md` delete the `| Key suffix |` table and replace "the table `check-loop-dispatch.sh` reads" and "off its row of the table above" with a citation of **Allowed pairs** (`skills/flow-contracts/model-policy.md`), cutting every sentence that restates a pair the table now holds.**
  - [x] **Step 3: Update `stats/internal/guard/check_loop_dispatch_test.go` fixtures that wrote a SKILL.md table to write the model-policy table instead; add `TestLoopDispatchReadsPolicyTable` (a fixture table whose `self-review-review-r` row names `opus/flow-low` admits that pair and refuses `sonnet/flow-low`). Verify: `cd stats && gofmt -l . && go vet ./internal/guard/... && go test ./internal/guard/ -run 'TestLoopDispatch|TestCheckLoopDispatch' -count=1`, then `git grep -n -F -e '| Key suffix |' -e lpTable -- stats/internal/guard skills/flow-self-review` prints nothing.**
**Build:** green
**Files:** `stats/internal/guard/loop.go`, `stats/internal/guard/check_loop_dispatch_test.go`, `skills/flow-self-review/SKILL.md`, `hooks/enforce-loop-dispatch.py`, `scripts/check-loop-dispatch.sh`, `stats/internal/guard/pipelinefix.go`, `stats/internal/guard/pipeline_fix_test.go`
**Tests:** `TestLoopDispatchReadsPolicyTable`
**Regression:** reverting restores the SKILL.md table read: `TestLoopDispatchReadsPolicyTable` fails because the fixture's policy row no longer decides the pair.
**Baseline:** before=7 after=8 <!-- measured: grep -c '^func Test' stats/internal/guard/check_loop_dispatch_test.go @ 122c3858 -->
**Commit:** refactor(guard): check loop dispatch pairs against the allowed-pairs table
**After:** Task 1

Correction (2026-10-29): the plan named `loop.go`, its test and SKILL.md alone; `pipelinefix.go` (and its test) also read `lpTable`/`lpRow` for `escalate`, so it moved onto `policyTable` too, and `hooks/enforce-loop-dispatch.py` now hands the prompt to `check-loop-dispatch.sh` on stdin, since a marked escalation row needs the prompt's `No-progress escalation` line; the guard's header was brought in line.

**Decision:** loop-suffix-maps-to-role

- [x] 3. `plan-agent-pair` prints the table's planning pair
  - [x] **Step 1: `stats/internal/guard/planagentpair.go` prints the first pair of the `plan-designer` row as `<model> <effort>` (today `opus low`) through `policyTable`, the repo resolved as `flow_guard_root` resolves it for the other guards; an unreadable table exits 2.**
  - [x] **Step 2: Add `TestPlanAgentPairFollowsPolicy` to `stats/internal/guard/plan_agent_pair_test.go` (a fixture table whose `plan-designer` row starts `opus/Plan@medium` prints `opus medium`). Verify: `cd stats && gofmt -l . && go vet ./internal/guard/... && go test ./internal/guard/ -run TestPlanAgentPair -count=1`.**
**Build:** green
**Files:** `stats/internal/guard/planagentpair.go`, `stats/internal/guard/plan_agent_pair_test.go`, `scripts/plan-agent-pair.sh`
**Tests:** `TestPlanAgentPairFollowsPolicy`
**Regression:** reverting restores the constant: the fixture's `medium` is ignored and the test fails.
**Baseline:** before=1 after=2 <!-- measured: grep -c '^func Test' stats/internal/guard/plan_agent_pair_test.go @ 122c3858 -->
**Commit:** refactor(guard): read the planning-agent pair from the allowed-pairs table
**After:** Task 1

Correction (2026-10-29): the plan named the Go file and its test alone; `scripts/plan-agent-pair.sh` now exports `FLOW_GUARD_REPO_ROOT` through `flow_guard_root`, which resolves the `skills/*/scripts/` symlinks the guard is called through, and its header names the table in place of the constant.

- [x] 4. `plan-class`'s extra dispatches pick from the table
  - [x] **Step 1: `stats/internal/guard/planclass.go` `pcExtra` picks between the `panel-extra` row's `sonnet/flow-low` and `opus/flow-low` pairs by the roll — `sonnet` below 50 as today — reading the table through `policyTable`; the printed `extra:` line is unchanged in shape.**
  - [x] **Step 2: Add `TestPlanClassExtraFollowsPolicy` to `stats/internal/guard/plan_class_test.go` (a fixture `panel-extra` row naming only `opus/flow-low` prints `opus low` for a roll below 50). Verify: `cd stats && gofmt -l . && go vet ./internal/guard/... && go test ./internal/guard/ -run TestPlanClass -count=1`.**
**Build:** green
**Files:** `stats/internal/guard/planclass.go`, `stats/internal/guard/plan_class_test.go`
**Tests:** `TestPlanClassExtraFollowsPolicy`
**Regression:** reverting restores the hard-coded sonnet pick: the test fails printing `sonnet low`.
**Baseline:** before=6 after=7 <!-- measured: grep -c '^func Test' stats/internal/guard/plan_class_test.go @ 122c3858 -->
**Commit:** refactor(guard): pick the extra-dispatch pair from the allowed-pairs table
**After:** Task 1

- [x] 5. Every pair statement becomes a citation; every dispatch template carries its `Role:` line
  - [x] **Step 1: In each file of `**Files:**` below, replace each sentence stating which model or effort a role runs on with a citation of **Allowed pairs** (`skills/flow-contracts/model-policy.md`) — cut, never paraphrased — keeping every roll band, ordering rule and decision-recording rule (brainstorm-planner.md's effort_roll bands, the micro literal `opus`, the `Operator-approved high` gate, pipeline.md's no-progress escalation procedure, which now names "the marked row of its fixer role" in place of "on `opus` at effort `high` (`flow-high`)").**
  - [x] **Step 2: Add a `Role: <role id>` line to every dispatch prompt template the files carry — implement.md's closed list gains a role-id column — and rename the five existing markers (`Role: visual-verify tooling analyst`→`Role: tooling-analyst`, `Role: visual-verify verifier`→`Role: visual-verifier`, `Role: panel findings merger`→`Role: panel-merger`, `Role: gated per-task reviewer`→`Role: gated-reviewer`, `Role: gated per-task re-reviewer`→`Role: gated-re-reviewer`).**
  - [x] **Step 3: Add `stats/internal/repotest/role_lines_test.go` with `TestEveryRoleLineIsAPolicyRow`: every `Role: <id>` line under `skills/` names a role of the **Allowed pairs** table. Verify: `cd stats && gofmt -l . && go test ./internal/repotest/ -run TestEveryRoleLineIsAPolicyRow -count=1`, `scripts/check-references.sh`, every other `scripts/check-*.sh` guard `.flow/project.md`'s `## lint` names, then `git grep -n -F -e 'Role: visual-verify tooling analyst' -e 'Role: visual-verify verifier' -e 'Role: panel findings merger' -e 'Role: gated per-task' -- skills` prints nothing.**
**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`, `skills/flow/review-panel.md`, `skills/flow/implement.md`, `skills/flow/visual-verify.md`, `skills/flow/visual-verify-tooling-analysis.md`, `skills/flow/gated-review-fix.md`, `skills/flow/review-panel-late-fix.md`, `skills/flow-contracts/pipeline.md`, `skills/flow-contracts/model-policy.md`, `stats/internal/repotest/role_lines_test.go`, `skills/flow/SKILL-rationale.md`, `stats/internal/guard/installedcitations.go`, `stats/internal/guard/check_installed_citations_test.go`
**Allowed-collateral:** `skills/flow/*.md`, `skills/flow-contracts/*.md`, `skills/spectre-planning/SKILL.md`, `spectre/**/spectre-planning/SKILL.md`, `AGENTS.md`, `rules/*.mdc`
**Tests:** `TestEveryRoleLineIsAPolicyRow`
**Regression:** reverting restores the old marker names and pair prose: `TestEveryRoleLineIsAPolicyRow` fails on `Role: panel findings merger` and the others.
**Baseline:** before=0 after=1 <!-- measured: cat stats/internal/repotest/role_lines_test.go 2>/dev/null | grep -c '^func Test' @ 122c3858 -->
**Commit:** docs(flow): cite the allowed-pairs table and give every dispatch its Role line
**After:** Task 1, 2

Correction (2026-10-29): `review-panel-fix-round.md`, `flow-plan/SKILL.md`, `flow-fast/SKILL.md` and `flow-self-review/SKILL.md` needed no edit — the first two inherit the Role line from **Explore and design agents** and **Dispatch sites — the parent's closed list**, flow-fast states no pair of its own, and task 2 already cut the self-review table; the Role line is stated once, as the closed list's new **Role line** column, rather than in each template. `check-installed-citations` read a `opus/flow-low` pair as an unrooted path, so `installedcitations.go` gained an anchored model-pair exclusion with a reporting case for `opus/flow-low.md`; `SKILL-rationale.md`'s two marker citations were renamed; the reworded sentences are acknowledged in this change's `verbatim-moves.txt`.

**Decision:** role-line-on-every-dispatch

- [x] 6. The agent-baseline hook enforces the table
  - [x] **Step 1: In `hooks/enforce-agent-baseline.py` delete `PINNED_ROLES`, `FLOW_HIGH_MARKERS`, `FLOW_XHIGH_TYPES`, `pinned_role_reason`, `flow_high_reason` and `flow_xhigh_reason`; add `dispatch_pair_reason(prompt, tool_input)` that, in a flow run (`in_flow_run`), for every dispatch but the exempt `fork` — Explore and Plan included — runs `<repo>/scripts/check-dispatch-pair.sh <repo> <subagent_type> <model> [<effort>]` with the prompt on stdin (`<repo>` as `enforce-loop-dispatch.py` resolves it) and denies with the guard's output on exit 1 or 2. On ZCode (`/.zcode/` in `__file__`) a known `Role:` line is still required but the pair is checked against `ZCODE_PAIR` only.**
  - [x] **Step 2: In `stats/internal/repotest/enforce_agent_baseline_hook_test.go` rewrite the pinned-role and flow-high cases onto role ids, and add `TestEnforceAgentBaselineHookRefusesOffPolicyDispatch` (in a flow run: denied with no `Role:` line, an unknown role, a wrong pair, `flow-high` without `No-progress escalation`, and `flow-xhigh`; admitted on an allowed pair) and `TestEnforceAgentBaselineHookZcodePairForKnownRole`. Verify: `cd stats && gofmt -l . && go test ./internal/repotest/ -run TestEnforceAgentBaselineHook -count=1`, then `git grep -n -F -e PINNED_ROLES -e FLOW_HIGH_MARKERS -e FLOW_XHIGH_TYPES -- hooks stats` prints nothing.**
**Build:** green
**Files:** `hooks/enforce-agent-baseline.py`, `stats/internal/repotest/enforce_agent_baseline_hook_test.go`
**Tests:** `TestEnforceAgentBaselineHookRefusesOffPolicyDispatch`, `TestEnforceAgentBaselineHookZcodePairForKnownRole`
**Regression:** reverting restores the three hard-coded rules: a `Role:`-less or off-table dispatch is admitted and `TestEnforceAgentBaselineHookRefusesOffPolicyDispatch` fails.
**Baseline:** before=7 after=9 <!-- measured: grep -c '^func Test' stats/internal/repotest/enforce_agent_baseline_hook_test.go @ 122c3858 -->
**Commit:** feat(hooks): refuse a dispatch off the allowed-pairs table
**After:** Task 1, 5

**Decision:** refuse-unresolved-role

**Decision:** guard-exit-2-denies
