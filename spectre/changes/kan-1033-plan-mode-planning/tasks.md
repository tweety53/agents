# kan-1033-plan-mode-planning

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

- [x] 1. Guard: `plan-agent-pair` — the reproducible 40/60 roll
  - [x] **Step 1: Add `stats/internal/guard/planagentpair.go` — `Registry["plan-agent-pair"]`; usage `plan-agent-pair <seed> <n>` (`<n>` a positive integer, else exit 2 with a usage line); roll = `sha256("<seed>plan<n>")` mod 100 computed exactly as `planclass.go` computes its rolls (reuse its helper, never a copy); print `sonnet low` when roll < 40, else `opus low`, one line, exit 0. Header comment: design.md §5 and decision `low-effort-40-60`.**
  - [x] **Step 2: Add the shim `scripts/plan-agent-pair.sh` modelled on `scripts/plan-class.sh` (`flow_guard_exec plan-agent-pair 2 "plan-agent-pair:" "$@"`), executable, and the symlinks `skills/flow/scripts/plan-agent-pair.sh` and `skills/flow-plan/scripts/plan-agent-pair.sh` → `../../../scripts/plan-agent-pair.sh`.**
  - [x] **Step 3: Tests — `stats/internal/guard/plan_agent_pair_test.go`: `TestPlanAgentPair` with subtests: two fixed seeds pinned to their computed outputs (one `sonnet low`, one `opus low`); the same arguments twice give the same line; over seeds `s0`…`s999` with `n=1` the sonnet share is within 35–45% (a property of sha256, pinned once, not a tolerance on a flaky value); `n=0`, a non-number and a missing argument exit 2. Verify: `cd stats && gofmt -l . && go vet ./internal/guard/... && go test ./internal/guard/... -run 'TestPlanAgentPair' -race -count=1 && cd .. && scripts/plan-agent-pair.sh demo 1 && scripts/check-guard-symlinks.sh`.**
**Build:** green
**Files:** `stats/internal/guard/planagentpair.go`, `stats/internal/guard/plan_agent_pair_test.go`, `stats/internal/guard/planclass.go`, `scripts/plan-agent-pair.sh`, `skills/flow/scripts/plan-agent-pair.sh`, `skills/flow-plan/scripts/plan-agent-pair.sh`
**Tests:** `TestPlanAgentPair`
**Regression:** reverting the commit removes the guard: `TestPlanAgentPair` no longer compiles and the shim exits 2 naming an unknown check.
**Baseline:** before=0 after=1 <!-- measured: grep -c '^func Test' stats/internal/guard/plan_agent_pair_test.go → absent at 43b6caf7 -->
**Commit:** feat(guard): roll a planning agent's model reproducibly
**After:** none

Correction (2026-10-25): the plan said reuse `planclass.go`'s roll helper; it was a closure inside `planClass`, so the commit lifts it to the package-level `pcRoll` and `planclass.go` joins `**Files:**`. The sonnet-share subtest pins the exact count over `s0`…`s999` instead of a 35–45% range. <!-- measured: scripts/plan-agent-pair.sh over s0..s999 n=1 → 381 sonnet -->

**Decision:** low-effort-40-60

- [x] 2. CLI: the `explorer` dispatch role
  - [x] **Step 1: In `stats/cmd/flow/record.go`, add `"explorer"` to `recordRoles` and extend its comment: an Explore subagent a planning session dispatches during `flow.brainstorm` (kan-1033, design.md §1); update the usage text's `-role` line to list it.**
  - [x] **Step 2: Test in `stats/cmd/flow/record_test.go`, beside `TestRecordAcceptsPlannerRole`: `TestRecordAcceptsExplorerRole` modelled on it. `TestRecordUsageNamesEveryRole` and `TestRunRecordRolesPrintsServedVocabulary` pick the role up from `recordRoles`; adjust a literal expected list only where one exists. Verify: `cd stats && gofmt -l . && go vet ./cmd/flow/... && go test ./cmd/flow/... -run 'TestRecordAcceptsExplorerRole|TestRecordUsageNamesEveryRole|TestRunRecordRolesPrintsServedVocabulary|TestSkillDocRolesAreAllServed' -race -count=1`.**
**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordAcceptsExplorerRole`
**Regression:** reverting the commit drops the role: `TestRecordAcceptsExplorerRole` fails because `-role explorer` is rejected before the store is contacted.
**Baseline:** before=131 after=132 <!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ 43b6caf7 → 131 -->
**Commit:** feat(stats): accept the explorer dispatch role
**After:** none

**Decision:** flow-plan-records-nothing

- [x] 3. Guard: the plan-mode sections of `design.md` in `check-plan-shape`
  - [x] **Step 1: In `stats/internal/guard/planshape.go`, when the scanned `tasks.md`'s leading `>` header block carries the line `> **Design shape:** plan-mode`, read the sibling `design.md` and report each missing heading of `## Recommended approach`, `## Critical files`, `## Existing code to reuse`, `## Verification` as `<design.md path>: missing section <heading> (Design shape: plan-mode)`; an absent `design.md` is one such line naming the file. Without the header line nothing changes. Extend the header comment with the rule and decision `design-shape-line`.**
  - [x] **Step 2: Tests in `stats/internal/guard/check_plan_shape_test.go`: `TestPlanShapeDesignSections` — a fixture with the header line and all four headings passes; the same without `## Verification` fails naming it; a fixture without the header line and without the headings passes. Verify: `cd stats && gofmt -l . && go vet ./internal/guard/... && go test ./internal/guard/... -run 'TestPlanShapeDesignSections' -race -count=1 && cd .. && scripts/check-plan-shape.sh`.**
**Build:** green
**Files:** `stats/internal/guard/planshape.go`, `stats/internal/guard/check_plan_shape_test.go`
**Tests:** `TestPlanShapeDesignSections`
**Regression:** reverting the commit drops the rule: `TestPlanShapeDesignSections` fails because a plan-mode change without `## Verification` passes.
**Baseline:** before=2 after=3 <!-- measured: grep -c '^func Test' stats/internal/guard/check_plan_shape_test.go @ 43b6caf7 → 2 -->
**Commit:** feat(guard): require the plan-mode design sections
**After:** none

**Decision:** design-shape-line

- [x] 4. Planner definitions, the pairs table and PLANNING DELEGATION
  - [x] **Step 1: Add `agents/flow-planner-low.md`, `agents/flow-planner-medium.md` and `agents/flow-planner-high.md`, each a copy of `agents/flow-low.md` / `agents/flow-medium.md` / `agents/flow-high.md` with `name` changed, `Agent` appended to `tools:`, and the description naming the planner role and decision `planner-carries-agent`.**
  - [x] **Step 2: `skills/flow-self-review/SKILL.md` — the pairs table's `plan`, `plan-<n>` row admits `flow-planner-medium` and `flow-planner-low` (medium by default, low with a recorded one-line reason when every finding is mechanical) with escalation `flow-planner-high`; the sentence on the no-progress re-plan names `flow-planner-high`; item 1's planner prompt carries a new **PLANNING DELEGATION:** paragraph in place of NO DELEGATION, worded per design.md §7 (fresh Explore and Plan agents only, pairs from `plan-agent-pair.sh <loop key> <n>`, never a fork, never an editing agent, every dispatch carrying the two agent-baseline sentences, every choice auto-resolved). State where `flow-medium` carries no Skill tool that `flow-planner-*` doesn't either.**
  - [x] **Step 3: `skills/flow-contracts/pipeline.md` **Pipeline defects found mid-run** — where it names the planner's type, name `flow-planner-low`/`-medium`/`-high`. `scripts/check-dispatch-paragraphs.sh` — add the `**PLANNING DELEGATION:**` row for `skills/flow-self-review/SKILL.md` to its table and exempt the plan dispatch from the NO DELEGATION row there.**
  - [x] **Step 4: Test in `stats/internal/guard/check_loop_dispatch_test.go`: `TestLoopDispatchPlannerTypes` — a `Key: self-review-x-plan` dispatch on `flow-planner-medium`/`opus` or `flow-planner-low`/`opus` is admitted, one on `flow-medium` is refused naming the row; `self-review-x-plan-2` on `flow-planner-high` is admitted only after an escalation. Verify: `cd stats && gofmt -l . && go test ./internal/guard/... -run 'TestLoopDispatchPlannerTypes|TestCheckDispatchParagraphs' -race -count=1 && cd .. && scripts/check-dispatch-paragraphs.sh && scripts/check-references.sh && scripts/check-markdown-integrity.sh`.**
**Build:** green
**Files:** `agents/flow-planner-low.md`, `agents/flow-planner-medium.md`, `agents/flow-planner-high.md`, `skills/flow-self-review/SKILL.md`, `stats/internal/guard/dispatchparagraphs.go`, `stats/internal/guard/loop.go`, `stats/internal/guard/check_loop_dispatch_test.go`
**Allowed-collateral:** `setup.sh`, `stats/internal/guard/check_dispatch_paragraphs_test.go`, `skills/flow/SKILL-rationale.md`
**Tests:** `TestLoopDispatchPlannerTypes`
**Regression:** reverting the commit restores `flow-medium` in the table: `TestLoopDispatchPlannerTypes` fails because the `flow-planner-medium` plan dispatch is refused.
**Baseline:** before=6 after=7 <!-- measured: grep -c '^func Test' stats/internal/guard/check_loop_dispatch_test.go @ 43b6caf7 → 6 -->
**Commit:** feat(flow-self-review): let the fix planner spawn planning agents
**After:** Task 1

Correction (2026-10-25): the dispatch-paragraph table lives in `stats/internal/guard/dispatchparagraphs.go`, not `scripts/check-dispatch-paragraphs.sh` (a shim), so the `planning` entry and its `skills/flow-self-review/SKILL.md` site land there; that file names no NO DELEGATION site to exempt. Its block reader now drops leading spaces before `>`, since the paragraph is a blockquote indented under item 1. `stats/internal/guard/loop.go` maps `flow-planner-<effort>` to its effort on the recorded row (it recorded `default`). `skills/flow-contracts/pipeline.md` names no planner type — it cites the pairs table — so it is unchanged.

**Decision:** planner-carries-agent

**Decision:** three-entry-points

- [x] 5. Planning texts: subagents, Explore phase, competing designs, plan sections, plan mode
  - [x] **Step 1: `skills/flow/brainstorm.md` — rewrite the dispatch-ban paragraph per design.md §1: a `fork` stays banned with its recorded reason; fresh subagents through the Agent tool are allowed during planning; the thinking stays in this session; `/flow` records each planning dispatch with `flow record dispatch begin/end` under `flow.brainstorm` (role `explorer` for Explore, `planner` for Plan, model and effort from `plan-agent-pair.sh <name> <n>`). Add the Claude Code plan-mode wrapper of design.md §6 for the creating run: `EnterPlanMode` after `flow.brainstorm` begins, `ExitPlanMode` after convergence closes as the design approval, reject-with-feedback reopening convergence, a declined consent one line.**
  - [x] **Step 2: `skills/flow/brainstorm-planner.md` — the opening paragraph's "no dispatched subagent, no relay" now reads that no one relays sections B–D, and subagents are allowed per §1. Add a subsection `### Explore and design agents` to **B. The checklist** carrying design.md §2 and §3 in full (when, how many, one message so they run concurrently, the briefing, the pair roll, read the critical files before asking, the recommendation with its alternatives recorded under `## Decisions`, auto-pick for the self-review and in-run fix planners). On Claude Code, the HARD GATE bullet names `ExitPlanMode` as the approval (decision `exit-plan-mode-is-approval`). In **C**, add the four sections of design.md §4 to `design.md`'s required headings and the `> **Design shape:** plan-mode` line to `tasks.md`'s header block.**
  - [x] **Step 3: `skills/flow-plan/SKILL.md` — "No subagent is dispatched" now allows fresh subagents per `brainstorm-planner.md`'s new subsection, the seed for `plan-agent-pair.sh` being this session's token; "No `flow record dispatch` call" stays, with decision `flow-plan-records-nothing`'s reason. Add the plan-mode wrapper of design.md §6 for `/flow-plan` (enter at session start; `ExitPlanMode` only at a capturable design, before the capture question; a research-only session never calls it). **The Fixed Section Structure** gains the four plan sections.**
  - [x] **Step 4: `skills/flow/SKILL-rationale.md` — under the existing no-dispatch reasoning, record why the ban narrows to forks (decision `agent-tool-everywhere`), and decisions `exit-plan-mode-is-approval` and `extend-not-replace` with their rejected alternatives.**
  - [x] **Step 5: Verify: `scripts/check-references.sh && scripts/check-markdown-integrity.sh && scripts/check-normative-inventory.sh && scripts/check-verbatim-moves.sh && scripts/check-ask-silence.sh && scripts/check-stage-mark-calls.sh && scripts/check-dispatch-paragraphs.sh && scripts/check-installed-citations.sh && scripts/check-plan-shape.sh spectre/changes/kan-1033-plan-mode-planning/tasks.md && (cd stats && go test ./cmd/flow/... -run 'TestSkillDocRolesAreAllServed' -count=1)`.**
**Build:** green
**Files:** `skills/flow/brainstorm.md`, `skills/flow/brainstorm-planner.md`, `skills/flow-plan/SKILL.md`, `skills/flow/SKILL-rationale.md`, `stats/internal/guard/installedcitations.go`, `skills/flow-self-review/scripts/plan-agent-pair.sh`
**Allowed-collateral:** `skills/flow-plan/SKILL-rationale.md`, `skills/flow/SKILL.md`, `rules/flow-manual-review.mdc`, `skills/flow-contracts/pipeline.md`
**Tests:** **none**
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** feat(flow): plan with Explore and Plan agents and Claude Code plan mode
**After:** Task 1, 2, 3, 4

Correction (2026-10-25): `check-installed-citations.sh` refused the three new `agents/flow-planner-*.md` as undeclared zero-citation files, so `installedcitations.go` declares them beside `flow-<effort>`; and `check-guard-symlinks.sh` refused `skills/flow-self-review/SKILL.md`'s PLANNING DELEGATION citation of `plan-agent-pair.sh` with no symlink, so that skill gains one. Both are task 4's surface, found only once task 5's text cited the guard.

**Decision:** extend-not-replace

**Decision:** all-four-practices

**Decision:** agent-tool-everywhere

**Decision:** exit-plan-mode-is-approval

**Decision:** design-shape-line
