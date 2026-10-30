## Context

- Ticket KAN-1067; scope added 2026-10-29: main-agent experiments, combined experiments, the current roles table in the stats UI.
- Every Go reader of the table goes through one parser, `policyTable` (`stats/internal/guard/modelpolicy.go`); five call sites (check-dispatch-pair, check-loop-dispatch, plan-agent-pair, plan-class, pipeline-fix). No guard speaks HTTP; guards reach flowd through the `flow` CLI (`pcFlow`, `stats/internal/guard/flowcli.go`).
- No hook can change a running session's model. The harvester already stamps main-thread records' `modelVersion` and `effort` into stage-run and dispatch metrics (`stats/internal/harvest/attribute.go`).
- Approved prototype — the spec for the Experiments view: `.superpowers/sdd/experiments-prototype.html` in this worktree, https://claude.ai/artifact/AHiBh1cdUrfY6omS8EuDhK.

## Decisions

### Override shape

**ID:** override-whole-cells
**Status:** active
**Chosen:** an experiment maps role id → the full list of pairs replacing that role's unmarked row; marked rows (escalations) are untouched — the ticket's decision 1.
**Considered:** per-pair add/remove deltas — harder to read back and to make permanent.

### Combined experiments

**ID:** combined-experiments
**Status:** active
**Chosen:** one experiment may carry cells for any number of roles (`-cell` repeated), `main` included; its trends show the union of those roles' metric families, and the closing agent may make all, some or none of its cells permanent, its summary stating that a combined result cannot attribute an effect to one cell.
**Considered:** one role per experiment, run side by side — the operator asked for combined experiments.

### Guards read overrides through the flow CLI

**ID:** overrides-via-flow-cli
**Status:** active
**Chosen:** `policyTable` runs `flow experiment overrides` (JSON `{role: [pairs]}`) via `pcFlow` and overlays it; any non-zero exit or unparsable output → the table alone.
**Considered:** a direct HTTP client in `stats/internal/guard` — the first one there; the CLI already owns address resolution and timeouts.

### Main-agent experiments are checked and tagged, not enforced

**ID:** main-check-and-tag
**Status:** active
**Chosen:** role `main` holds one `<model>@<effort>` pair. `/flow` reads it at start and prints `⚠ experiment <id>: main agent should run <pair>; this session runs <model>@<effort>` on a mismatch, never blocking. A change counts in the experiment's "after" set only when its stage runs' harvested main-thread model and effort match.
**Considered:** a `flow claude` launch wrapper — guaranteed compliance but changes how every session starts.

### Scope and safety

**ID:** one-open-per-role
**Status:** active
**Chosen:** applies to every run from `started_at`; before/after comparison, no control arm. The store rejects opening an experiment whose roles overlap an open one (409). Closing sets `closed_at`; a closed experiment is never returned by `overrides`.
**Considered:** random arms — the ticket's decision 2 rules them out.

### Verify reds are derived, plan corrections are recorded

**ID:** metrics-records
**Status:** active
**Chosen:** verify reds = dispatch rows keyed `verify*` ended `blocked` with cause `test-failure` (already stored). Plan corrections = the count of `Correction (YYYY-MM-DD):` paragraphs in `tasks.md`, recorded by `flow tasks count` into a new `corrections` column of `change_task_counts`; `skills/flow/implement.md` has the parent run `flow tasks count` after transcribing a Correction.
**Considered:** a new verify-red event — duplicates data already stored; a bash hook in `close-task.sh` — new logic is not grown in bash.

### Experiments view on the SPA

**ID:** experiments-view
**Status:** active
**Chosen:** a new `experiments` stats view and `views/Experiments.tsx`, drawn exactly as the approved prototype.
**Considered:** markers on Review dynamics — no per-role metrics or summaries.

### The roles table flowd serves

**ID:** roles-table-embedded
**Status:** active
**Chosen:** the table parser moves to `stats/internal/modelpolicy` (`Parse([]byte)`), shared by the guard and flowd; `make web-dist` copies `skills/flow-contracts/model-policy.md` into `stats/internal/modelpolicy/embedded/model-policy.md` (a subpackage the guard never imports, so a plain `go build ./cmd/flow-guard` needs no make) (gitignored, like `dist`) and flowd embeds it, so the table is the one `make restart` built after the last landing.
**Considered:** flowd reading the repo file at runtime from its executable's path — breaks for any binary not under `stats/bin`.

### Metric families per role

**ID:** role-metric-families
**Status:** active
**Chosen:** planners (`planner`, `plan-designer`, `explorer`, `self-review-plan`, `pipeline-fix-plan`): review rounds/change, round-0 Critical+Important/change, plan corrections/change, tokens/change, wall-clock/change. Implementers/fixers (`implementer`, `panel-fix`, `self-review-fix`, `pipeline-fix-fix`, `late-fix`): review rounds/change, findings/task, verify reds/change, tokens and wall-clock of the role's dispatches. Reviewers (`panel-reviewer`, `panel-extra`, `panel-rerun`, `gated-reviewer`, `gated-re-reviewer`, `self-review-review-1`, `self-review-review-r`, `pipeline-fix-review-1`, `pipeline-fix-review-r`): findings raised/dispatch, late-round findings (round ≥ 2)/change, review rounds/change, tokens and wall-clock. `visual-verifier`: mismatches caught (findings by visual verify), re-runs (`visual-verify-fix-*` keys)/change, tokens and wall-clock. `main`: the planner set plus verify reds. Any other role: tokens and wall-clock.
**Considered:** one metric set for all roles — the ticket's decision 3 asks for per-family measures.

### Closing agent is a standalone skill

**ID:** close-skill
**Status:** active
**Chosen:** `skills/flow-experiment-close/SKILL.md` (`/flow-experiment-close <id>`): reads `flow experiment stats <id>`, writes a verdict (`positive|negative|inconclusive`) and summary with `flow experiment close`, and on `positive` starts `/flow` with the winning cells as the change description.
**Considered:** a flowd-side automatic close — no agent judgement, and the ticket asks for an agent.

### First experiment is opened after landing

**ID:** first-experiment-after-land
**Status:** superseded by first-experiment-with-planner
**Chosen:** one combined experiment — `plan-designer=opus/Plan@xhigh` and `explorer=sonnet/Explore@low` — opened with `flow experiment open` once the change lands and flowd restarts, named in the handoff; no task writes the dev store.
**Considered:** a migration seeding it — a store row no command made, opening on every fresh database.
**Superseded because:** the operator moved brainstorming and plan writing to a `planner` role (decision planner-role, 2026-10-29), so the xhigh-planning question now targets that role.

### Brainstorming and plan writing run in a planner role

**ID:** planner-role
**Status:** active
**Chosen:** `/flow`'s brainstorming, the three artifacts and writing-plans move off the parent into a dispatched `planner` role (Allowed pairs row `planner`: `opus/flow-planner-medium`, `opus/flow-planner-low`), so an experiment can target planning apart from the main agent. Each dispatch is one-shot: a design round writes `.superpowers/sdd/planner/round-<n>.md` (draft plus questions, each with a recommended option); the parent relays the questions verbatim, writes `answers-<n>.md` and dispatches the next round; after approval a final dispatch runs sections C and D and the plan guards. The parent keeps relaying, the stage marks around approval, Decide and the gate. `/flow-plan`, the self-review planner and the in-run fix planner are unchanged.
**Considered:** two roles (brainstormer, plan-writer) — the operator chose one; only the artifacts and tasks off-session — the operator asked for all three; a long-lived planner resumed with SendMessage — the pipeline bans resuming a finished child.

### First experiment, with the planner role

**ID:** first-experiment-with-planner
**Status:** active
**Chosen:** one combined experiment opened by the operator after landing: `planner=opus/flow-planner-xhigh`, `plan-designer=opus/Plan@xhigh`, `explorer=sonnet/Explore@low`; named in the handoff, no task writes the dev store.
**Considered:** `main=opus@xhigh` — writing-plans no longer runs on the main agent.

## Open questions

## Recommended approach

Store first (migration, store, API, client, CLI), then the shared table parser and the `policyTable` overlay, then the main-agent check, the corrections record, the stats aggregate and view, the closing skill and the docs. Last, the planner role and its contract. Rejected alternatives: direct HTTP (overrides-via-flow-cli), launch wrapper (main-check-and-tag), review-dynamics markers (experiments-view), runtime file read (roles-table-embedded).

## Critical files

- `stats/internal/store/migrations/0054_experiments.sql`, `0055_task_count_corrections.sql` — new table/column.
- `stats/internal/store/experiments.go`, `experimentstats.go` — CRUD, overlap check, stats aggregate.
- `stats/internal/store/taskcounts.go` — corrections column.
- `stats/internal/modelpolicy/` — shared parser and embedded table.
- `stats/internal/api/experiments.go`, `server.go` — routes.
- `stats/internal/client/client.go` — client methods.
- `stats/cmd/flow/experiment.go`, `main.go`, `tasks.go`, `decision.go` — CLI.
- `stats/internal/guard/modelpolicy.go` and its five callers — overlay.
- `stats/Makefile`, `stats/.gitignore` — the copy step.
- `stats/web/src/views/Experiments.tsx`, `App.tsx`, `api.ts` — SPA.
- `stats/cmd/uitest-seed/seed.go`, `stats/web/tests/visual/experiments.spec.ts` — fixture and capture.
- `skills/flow/brainstorm.md` — the planner-role protocol.
- `skills/flow-experiment-close/SKILL.md`, `skills/flow/SKILL.md`, `skills/flow/implement.md`, `skills/flow/brainstorm-planner.md`, `skills/flow-contracts/model-policy.md`, `CLAUDE.md`, `stats/README.md` — contracts and docs.

## Existing code to reuse

- `store.PutSettings`/`GetSettings` (`stats/internal/store/settings.go`), `settingsHandler` (`stats/internal/api/settings.go`), `client.GetSettings`, `cmd/flow/settings.go` — the entity template.
- `pcFlow` (`stats/internal/guard/flowcli.go`) — CLI call from guards.
- `ReviewDynamics` (`stats/internal/store/reviewdynamics.go`) — finding/round SQL and `canonical_severity()`.
- `change_task_counts` (`stats/internal/store/taskcounts.go`) and `flow tasks count`.
- `internal/web/embed.go` — the build-copied embed pattern.
- `ViewFrame`, `Panel`, `StatPanel` (`stats/web/src/components/`) and `theme.css` tokens.

## Verification

- `cd stats && make -s web-dist && go vet ./... && gofmt -l .` clean; `go test ./...` green; `cd stats/web && npx tsc -b && npx vitest run` green; the Experiments capture spec matches the prototype.
- Live check below.

## Live check

On `scripts/live-store-copy.sh up kan-1067` (a copy of the dev store, its own flowd):
1. `flow experiment open -name t -cell 'planner=opus/flow-planner-xhigh' -cell 'plan-designer=opus/Plan@xhigh' -cell 'explorer=sonnet/Explore@low'` → id printed; a second open naming `explorer` → exit 1 naming the conflict.
2. `scripts/plan-agent-pair.sh x 1` against the copy prints `opus xhigh`; before step 1 it printed a `low|medium` pair. Failure looks like the same pair before and after.
3. check-dispatch-pair with `Role: explorer` on `sonnet/Explore@low`, and with `Role: planner` on `opus/flow-planner-xhigh`, is admitted; with `FLOW_ADDR` at a dead port it is refused (table fallback).
4. `flow experiment stats <id>` prints non-zero `before` figures from real history; all-zero before figures = failure.
5. `flow experiment close <id> -verdict negative -summary s`; `flow experiment overrides` prints `{}`.
6. The Experiments view at the copy's port shows the roles table, with the override struck through while open, and the experiment's before figures.
