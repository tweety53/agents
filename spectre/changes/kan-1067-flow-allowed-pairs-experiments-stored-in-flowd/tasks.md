# kan-1067-flow-allowed-pairs-experiments-stored-in-flowd

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

**Goal:** Allowed-pairs experiments stored in flowd — overrides read by every table reader, main-agent and combined experiments, per-role trends and the roles table on an Experiments view, and an agent that closes them (KAN-1067).

**Architecture:** a flowd `experiments` table behind store/API/client/`flow experiment`; a shared `internal/modelpolicy` parser that the guards overlay with `flow experiment overrides` and flowd embeds; a stats aggregate and SPA view; a closing skill.

**Tech Stack:** Go (`stats/`), PostgreSQL migrations, React/TypeScript (`stats/web`), Playwright, Markdown.

**Spec:** `design.md` beside this file; the approved prototype `.superpowers/sdd/experiments-prototype.html` (https://claude.ai/artifact/AHiBh1cdUrfY6omS8EuDhK).

## Global constraints

- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1067-flow-allowed-pairs-experiments-stored-in-flowd` (`<wt>`), branch `spectre/kan-1067-flow-allowed-pairs-experiments-stored-in-flowd`; `**Files:**` paths are relative to `<wt>`. Never touch the main checkout, never create a worktree, never write `spectre/changes/`.
- Go: `gofmt -w` every Go file changed; `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean. SPA: `cd stats && make -s web-dist && cd web && npx tsc -b` exits 0.
- Store tests use the package's existing test-database helper, never the live `flow` database. Never stop, restart or drop `flowd` on 4173, the `flow-postgres` container or the `flow` database.
- Every Go test run sets a Bash `timeout` of 600000.
- Mutation proofs restore with `scripts/break-and-prove.sh` — never `git checkout --` over uncommitted edits.

## Review Focus

- Task 3: a failed or unparsable `flow experiment overrides` leaves every reader on the table alone; marked rows are never overridden; `main` never enters the guard table.
- Task 1: two concurrent opens sharing a role cannot both succeed (the check and insert are one transaction or a constraint).
- Task 13: every planner dispatch is one-shot and fresh; the parent asks the round's questions verbatim and never answers one itself; stage marks stay in order.
- Task 6: before/after windows are equal-length and a change counts once, in the window its first stage run started in; `main` cells count only matching runs.

- [ ] 1. Store experiments in flowd
  - [ ] **Step 1: Write the tests.** In `stats/internal/store/experiments_test.go`, on the settings tests' database helper: `TestExperimentOpenListClose` — `OpenExperiment(ctx, Experiment{Name:"t", Cells: map[string][]string{"plan-designer":{"opus/Plan@xhigh"}, "explorer":{"sonnet/Explore@low"}}})` returns an id and `StartedAt`; `ListExperiments(ctx, false)` returns it with `Status "open"`; `OpenOverrides(ctx)` returns both cells; `CloseExperiment(ctx, id, "negative", "s")` sets `ClosedAt`, `Verdict`, `Summary`, and `OpenOverrides` is then empty. `TestExperimentOpenRefusesOverlappingRole` — a second open naming `explorer` returns `ErrExperimentConflict` whose message names the role and the open experiment's id; after the first closes it succeeds. `TestValidateExperiment` — empty name, no cells, a cell with no pairs, a pair not shaped `<model>/<type>[@<effort>]`, a `main` cell not shaped `<model>@<effort>` or with more than one pair, an unknown verdict on close — each an error. Run `cd stats && go test ./internal/store -run 'TestExperiment|TestValidateExperiment' -count=1` — fails to compile.
  - [ ] **Step 2: Migration.** `stats/internal/store/migrations/0054_experiments.sql`, header comment in the house style citing KAN-1067: `CREATE TABLE IF NOT EXISTS experiments (id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL, cells JSONB NOT NULL, started_at TIMESTAMPTZ NOT NULL DEFAULT now(), closed_at TIMESTAMPTZ, verdict TEXT CHECK (verdict IN ('positive','negative','inconclusive')), summary TEXT NOT NULL DEFAULT '', CHECK ((closed_at IS NULL) = (verdict IS NULL)))` and `CREATE TABLE IF NOT EXISTS experiment_roles (role TEXT PRIMARY KEY, experiment_id BIGINT NOT NULL REFERENCES experiments(id))` — one row per role of an open experiment, so the primary key is the one-open-per-role guarantee (decision one-open-per-role).
  - [ ] **Step 3: Implement.** `stats/internal/store/experiments.go`: `type Experiment struct{ID int64; Name string; Cells map[string][]string; StartedAt time.Time; ClosedAt *time.Time; Verdict, Summary, Status string}` (JSON tags camelCase), `ValidateExperiment`, `ErrExperimentConflict`, `OpenExperiment` (one transaction: insert experiment, insert each role into `experiment_roles`; a unique violation → `ErrExperimentConflict` naming the role and the holder's id), `ListExperiments(ctx, openOnly bool)` newest first with `Status` derived from `closed_at`, `OpenOverrides(ctx) (map[string][]string, error)`, `CloseExperiment` (one transaction: set `closed_at=now()`, verdict, summary; delete its `experiment_roles` rows; an already-closed or unknown id → `ErrExperimentNotOpen`).
  - [ ] **Step 4: Verify.** Tests pass; `scripts/break-and-prove.sh` with the `experiment_roles` insert removed — `TestExperimentOpenRefusesOverlappingRole` fails; restore. Lint. Commit.

**Build:** green
**Files:** `stats/internal/store/migrations/0054_experiments.sql`, `stats/internal/store/experiments.go`, `stats/internal/store/experiments_test.go`
**Tests:** `TestExperimentOpenListClose`, `TestExperimentOpenRefusesOverlappingRole`, `TestValidateExperiment`
**Regression:** `TestExperimentOpenRefusesOverlappingRole` fails — two open experiments govern one role.
**Commit:** feat(store): store allowed-pairs experiments with one open experiment per role
**After:** none

**Decision:** override-whole-cells
**Decision:** combined-experiments
**Decision:** one-open-per-role

- [ ] 2. Serve experiments over the API, the client and `flow experiment`
  - [ ] **Step 1: Write the tests.** `stats/internal/api/experiments_test.go` on `settings_test.go`'s fake-store shape: `TestExperimentsRoutes` — `POST /api/v1/experiments` 201 with the id, 400 on a validation error, 409 on `ErrExperimentConflict` with its message; `GET /api/v1/experiments?open=1`; `GET /api/v1/experiments/overrides` → `{role:[pairs]}`; `POST /api/v1/experiments/{id}/close` 200, 404 on `ErrExperimentNotOpen`. `stats/cmd/flow/experiment_test.go`: `TestFlowExperiment` against an `httptest` server — `open -name t -cell 'a=x/y@low' -cell 'b=x/z'` posts both cells and prints `experiment <id> open`; a 409 exits 1 printing the server's message; `list [-open]` prints one line per experiment (`<id> <status> <started yyyy-mm-dd> <name> <role>=<pairs>;…`); `overrides` prints the JSON object compactly and `{}` when none; `close <id> -verdict v -summary s` exits 0; `close` with a bad verdict exits 2; an unreachable store exits 1 on every subcommand. Run — fails.
  - [ ] **Step 2: Implement.** `stats/internal/api/experiments.go` (handler on the `settingsHandler` pattern) and its routes in `stats/internal/api/server.go` beside `/api/v1/settings`; wire the store in `stats/cmd/flowd/main.go` as settings is; `stats/internal/client/client.go` gains `OpenExperiment`, `ListExperiments`, `ExperimentOverrides`, `CloseExperiment` with `ErrExperimentConflict` for a 409. `stats/cmd/flow/experiment.go`: `runExperiment` with `open|list|overrides|close`, store flags from `registerStoreConnFlags`, the timeout/recover wrapper of `cmd/flow/settings.go`; `stats/cmd/flow/main.go` gains `case "experiment":` and its usage line. Update `stats/cmd/flowd/wiring_test.go` if it enumerates routes.
  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/api ./cmd/flow ./cmd/flowd ./internal/client -run 'Experiment|Wiring' -count=1` passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/api/experiments.go`, `stats/internal/api/experiments_test.go`, `stats/internal/api/server.go`, `stats/internal/client/client.go`, `stats/cmd/flow/experiment.go`, `stats/cmd/flow/experiment_test.go`, `stats/cmd/flow/main.go`, `stats/cmd/flowd/main.go`
**Allowed-collateral:** `stats/cmd/flowd/wiring_test.go`, `stats/internal/api/*_test.go`
**Tests:** `TestExperimentsRoutes`, `TestFlowExperiment`
**Regression:** `TestFlowExperiment` fails — `flow experiment overrides` cannot reach the store, so no guard sees an override.
**Commit:** feat(stats): serve experiments through the API and flow experiment
**After:** Task 1

- [ ] 3. Overlay open overrides on every allowed-pairs reader
  - [ ] **Step 1: Write the tests.** `stats/internal/modelpolicy/modelpolicy_test.go`: `TestParse` moves the existing table-shape cases from the guard tests (a missing table, a 2-cell row, a row with two markers → error; the real file parses). In `stats/internal/guard/check_dispatch_pair_test.go`: `TestPolicyTableOverlaysOpenOverrides` — with a fake `flow` on `env`'s PATH printing `{"plan-designer":["opus/Plan@xhigh"],"main":["opus@medium"]}`, `policyTable` returns `plan-designer`'s unmarked row as exactly `opus/Plan@xhigh`, every marked row untouched, no `main` role; a fake exiting 1, one printing `not json`, and no `flow` on PATH each yield the table alone; an override naming a role the table lacks is ignored. `TestCheckDispatchPairAdmitsOverride` — `Role: plan-designer` on `opus/Plan@xhigh` admitted with the fake, refused without. In `stats/internal/guard/plan_agent_pair_test.go`: `TestPlanAgentPairReadsOverride` — `plan-agent-pair` prints `opus xhigh` under the fake. Run — fails.
  - [ ] **Step 2: Extract the parser.** Move `policyRow`, the parse loop and `ticks` into `stats/internal/modelpolicy/modelpolicy.go` as exported `Row{Pairs []string; Marker string}` and `Parse(b []byte) (map[string][]Row, error)`; the guard's `policyTable` reads the file and calls it. Keep `policyAllows`, `policyPair` in the guard.
  - [ ] **Step 3: Overlay.** `policyTable(env Env, repo string)`: after parsing, `pcFlow(env, false, "experiment", "overrides")`; on exit 0 and valid JSON, for each role present in the table (never `main`), replace the pairs of its first unmarked row with the override's. Update the five callers — `checkDispatchPair`, `loop.go`, `planagentpair.go`, `planclass.go`, `pipelinefix.go` — to pass their `env`. `plan-agent-pair`'s `papPair` regex admits `xhigh` and `max`.
  - [ ] **Step 4: Prove.** `scripts/break-and-prove.sh` with the overlay loop removed — `TestPolicyTableOverlaysOpenOverrides` fails; restore.
  - [ ] **Step 5: Verify.** `cd stats && go test ./internal/modelpolicy ./internal/guard -run 'TestParse|TestPolicyTable|TestCheckDispatchPair|TestPlanAgentPair|TestPlanClass|TestCheckLoopDispatch|TestPipelineFix' -count=1` passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/modelpolicy/modelpolicy.go`, `stats/internal/modelpolicy/modelpolicy_test.go`, `stats/internal/guard/modelpolicy.go`, `stats/internal/guard/loop.go`, `stats/internal/guard/planagentpair.go`, `stats/internal/guard/planclass.go`, `stats/internal/guard/pipelinefix.go`, `stats/internal/guard/check_dispatch_pair_test.go`, `stats/internal/guard/plan_agent_pair_test.go`
**Allowed-collateral:** `stats/internal/guard/*_test.go`
**Tests:** `TestParse`, `TestPolicyTableOverlaysOpenOverrides`, `TestCheckDispatchPairAdmitsOverride`, `TestPlanAgentPairReadsOverride`
**Regression:** `TestPolicyTableOverlaysOpenOverrides` fails — an open experiment governs no dispatch.
**Commit:** feat(guard): overlay open experiment overrides on the allowed-pairs table
**After:** Task 2

Correction (2026-10-30): Step 3 planned widening `papPair`'s regex to admit `xhigh` and `max`; its effort group is already `[a-z]+`, so no change shipped. The guard tests' `TestMain` also pins `FLOW_RECORDS_ADDR` to a dead port so no guard test reads the dev store's open experiments.

**Decision:** overrides-via-flow-cli
**Decision:** override-whole-cells

- [ ] 4. Embed the roles table in flowd and serve it with overrides applied
  - [ ] **Step 1: Write the tests.** `stats/internal/api/experiments_test.go`: `TestRolesRoute` — `GET /api/v1/roles` returns one entry per table row `{role, pairs, marker, override?: {pairs, experimentId}}` in table order, plus a leading `main` entry `{role:"main", pairs:[], override?}`; with the fake store's open overrides for `plan-designer` and `main`, those entries carry `override`; marked rows never do. `stats/internal/modelpolicy/modelpolicy_test.go`: `TestEmbeddedTableParses` — `Parse(Embedded)` succeeds and holds `plan-designer`. Run — fails.
  - [ ] **Step 2: Implement.** `stats/internal/modelpolicy/embed.go`: `//go:embed model-policy.md` into `Embedded []byte`, with a doc comment of `internal/web/embed.go`'s shape naming the copy step. `stats/Makefile`: a phony `policy-copy` target running `cp ../skills/flow-contracts/model-policy.md internal/modelpolicy/model-policy.md`, made a prerequisite of `web-dist`, `build` and `test`, so every lint, test and build line copies it fresh. `.gitignore` gains `stats/internal/modelpolicy/model-policy.md` with a comment. The roles handler joins `modelpolicy.Parse(modelpolicy.Embedded)` with `ListExperiments(ctx, true)`.
  - [ ] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/modelpolicy ./internal/api -run 'Roles|Embedded' -count=1` passes; `git status --short stats/internal/modelpolicy` lists no `model-policy.md`; lint. Commit.

**Build:** green
**Files:** `stats/internal/modelpolicy/modelpolicy_test.go`, `stats/internal/api/experiments.go`, `stats/internal/api/experiments_test.go`, `stats/internal/api/server.go`, `stats/Makefile`, `.gitignore`, `stats/internal/modelpolicy/embedded/embed.go`, `stats/internal/modelpolicy/embedded/embedded_test.go`
**Tests:** `TestRolesRoute`, `TestEmbeddedTableParses`
**Regression:** `TestRolesRoute` fails — the stats UI has no roles table to draw.
**Commit:** feat(stats): serve the allowed-pairs roles table with open overrides applied
**After:** Task 3

Correction (2026-10-30): the plan put `//go:embed model-policy.md` in `stats/internal/modelpolicy`, which `internal/guard` imports after Task 3; the file is gitignored and only make writes it, so plain `go build ./cmd/flow-guard` (`scripts/lib/flow-guard.sh`) would fail. The embed moves to `stats/internal/modelpolicy/embedded/` (`embed.go`, `embedded_test.go` holding `TestEmbeddedTableParses`); the copy target and the `.gitignore` entry become `stats/internal/modelpolicy/embedded/model-policy.md`.

**Decision:** roles-table-embedded

- [ ] 5. Record plan corrections with `flow tasks count`
  - [ ] **Step 1: Write the tests.** `stats/cmd/flow/tasks_test.go`: `TestTasksCountPostsCorrections` — a plan whose task entries carry two `Correction (2026-10-29):` paragraphs (plus a fenced example line `Correction (YYYY-MM-DD):`, not counted) posts `{"totalTasks":N,"corrections":2}`. `stats/internal/store/taskcounts_test.go`: `TestRecordTaskCountCorrections` stores and reads the column; an old client posting no `corrections` stores 0. Run — fails.
  - [ ] **Step 2: Implement.** `stats/internal/store/migrations/0055_task_count_corrections.sql`: `ALTER TABLE change_task_counts ADD COLUMN IF NOT EXISTS corrections INT NOT NULL DEFAULT 0`, header comment citing KAN-1067. `stats/internal/store/taskcounts.go`, its API handler and client type, and `stats/cmd/flow/tasks.go` carry `corrections` — counted as lines matching `^\s*Correction \(\d{4}-\d{2}-\d{2}\):` outside fences.
  - [ ] **Step 3: Edit the contract.** In `skills/flow/implement.md`, after the sentence ending "(kan-1020: a bare count in a Correction was refused only at verify).", add: "Once a `Correction:` paragraph is transcribed, the parent runs `flow tasks count -C <worktree> <change>`, which records the plan's correction count beside its task count — the plan-corrections metric experiments compare (KAN-1067)." Wrap at the file's width.
  - [ ] **Step 4: Verify.** Tests pass; `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0; lint. Commit.

**Build:** green
**Files:** `stats/internal/store/migrations/0055_task_count_corrections.sql`, `stats/internal/store/taskcounts.go`, `stats/internal/store/taskcounts_test.go`, `stats/cmd/flow/tasks.go`, `stats/cmd/flow/tasks_test.go`, `skills/flow/implement.md`, `stats/internal/records/types.go`, `stats/cmd/flow/tick.go`
**Allowed-collateral:** `stats/internal/api/*.go`, `stats/internal/client/client.go`
**Tests:** `TestTasksCountPostsCorrections`, `TestRecordTaskCountCorrections`
**Regression:** `TestTasksCountPostsCorrections` fails — plan corrections are never stored and the planner trend reads zero.
**Commit:** feat(stats): record a plan's correction count with flow tasks count
**After:** none

**Decision:** metrics-records

- [ ] 6. Compute experiment stats per role family
  - [ ] **Step 1: Write the tests.** `stats/internal/store/experimentstats_test.go`, seeding changes, stage runs, dispatches, findings, task counts and one experiment: `TestExperimentStatsPlannerFamily` — a `plan-designer` experiment started at T returns, for each planner metric (`reviewRounds`, `round0CritImp`, `planCorrections`, `tokens`, `wallClockMin`), `before`/`after` means over changes whose first stage run started in `[T-d, T)` / `[T, now or closed_at)`, `d` the after window's length, plus a weekly `series` of `{week, value}` and `runs: {before, after}`. `TestExperimentStatsImplementerFamily` covers `verifyReds` (rows keyed `verify%` ended `blocked` cause `test-failure`) and `findingsPerTask`. `TestExperimentStatsReviewerAndVisual` covers `findingsPerDispatch`, `lateRoundFindings`, `vvReruns`. `TestExperimentStatsMainCountsMatchingRunsOnly` — a `main=opus@medium` cell counts in `after` only a change whose stage runs' `metrics.models` holds `opus` main-thread tokens and `metrics.effort = 'medium'`; a mismatching change is excluded from `after` and reported as `excluded`. `TestExperimentStatsCombinedUnion` — a two-role experiment returns the union of both families, each metric once. Run — fails.
  - [ ] **Step 2: Implement.** `stats/internal/store/experimentstats.go`: `ExperimentStats(ctx, id) (ExperimentStatsResult, error)`, metric families per decision role-metric-families, SQL reusing `canonical_severity()` and the `ReviewDynamics` round/finding joins; tokens from `metrics->'tokens'` of the role's dispatches; wall-clock `ended_at - started_at`. The `main` match reads stage-run metrics as `stats/internal/harvest/attribute.go`'s `MetricsPatch` writes them. <!-- unverified: the key under metrics.models naming a model (alias vs model id) — read stats/internal/harvest/attribute.go and match both forms --> Route `GET /api/v1/experiments/{id}/stats` in `stats/internal/api/experiments.go`; `client.ExperimentStats`; `flow experiment stats <id>` prints the JSON.
  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/store ./internal/api ./cmd/flow -run 'ExperimentStats|Experiment' -count=1` passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/store/experimentstats.go`, `stats/internal/store/experimentstats_test.go`, `stats/internal/api/experiments.go`, `stats/internal/api/experiments_test.go`, `stats/cmd/flow/experiment.go`, `stats/cmd/flow/experiment_test.go`, `stats/internal/client/client.go`
**Tests:** `TestExperimentStatsPlannerFamily`, `TestExperimentStatsImplementerFamily`, `TestExperimentStatsReviewerAndVisual`, `TestExperimentStatsMainCountsMatchingRunsOnly`, `TestExperimentStatsCombinedUnion`
**Regression:** `TestExperimentStatsMainCountsMatchingRunsOnly` fails — a main-agent experiment's after figures count sessions that never ran its pair.
**Commit:** feat(store): compute per-role-family before/after stats for an experiment
**After:** Task 2, 4, 5

**Decision:** role-metric-families
**Decision:** main-check-and-tag

- [ ] 7. Draw the Experiments view as the approved prototype
  - [ ] **Step 1: Write the tests.** `stats/web/src/views/views.test.tsx` gains `Experiments` cases on the file's existing fetch-mock shape: the roles table lists every row, an overridden cell shows the table pairs struck through (`<s>`) and the override pairs, linking to its experiment; the list marks open/closed; selecting an experiment shows its override table, one stat panel per metric with `▲`/`▼` and the percentage vs before, one chart per metric with the start marker and both mean lines (after line coloured `--delta-down` when better, `--delta-up` when worse — lower is better for every metric), the legend, and the summary or "Open — the closing agent writes the summary." Run `cd stats/web && npx vitest run src/views/views.test.tsx -t 'Experiments'` — fails.
  - [ ] **Step 2: Implement.** `stats/web/src/views/Experiments.tsx` reproducing `.superpowers/sdd/experiments-prototype.html` exactly — layout, copy, tokens from `theme.css`, the 320×150 chart geometry, gridlines and labels, shaded window, dashed means, hover/focus point enlargement with its `title` tooltip; `api.ts` gains `fetchRoles`, `fetchExperiments`, `fetchExperimentStats`; `App.tsx` registers the `experiments` route labelled `Experiments` after `Self review`. Empty states: no experiments → the list reads "No experiments yet — open one with `flow experiment open`."; an experiment with no runs after its start → stat panels read `—`.
  - [ ] **Step 3: Verify.** `cd stats && make -s web-dist && cd web && npx tsc -b && npx vitest run src/views/views.test.tsx -t 'Experiments'` passes. Commit.

**Build:** green
**Files:** `stats/web/src/views/Experiments.tsx`, `stats/web/src/views/views.test.tsx`, `stats/web/src/api.ts`, `stats/web/src/App.tsx`
**Allowed-collateral:** `stats/web/src/styles.css`, `stats/web/src/App.test.tsx`, `stats/web/src/api.test.ts`
**Tests:** `Experiments`
**Regression:** the `Experiments` cases fail — the view or its roles table is gone.
**Commit:** feat(web): draw the Experiments view with the roles table and per-metric trends
**After:** Task 4, 6

**Decision:** experiments-view

- [ ] 8. Capture the Experiments view against the prototype
  - [ ] **Step 1: Seed.** `stats/cmd/uitest-seed/seed.go` gains `seedExperiments`: one open combined experiment (`plan-designer=opus/Plan@xhigh`, `explorer=sonnet/Explore@low`) started a week before `fixtureNow`, one closed `implementer=sonnet/flow-low` experiment with verdict `negative` and the prototype's summary text, over the fixture's existing changes.
  - [ ] **Step 2: Capture.** `stats/web/tests/visual/experiments.spec.ts` on `review-dynamics.spec.ts`'s shape, the period pinned by query params: screenshot the view with the open experiment selected, the closed one selected, and a chart point hovered and focused; each locator by role or text. Its `experiments.spec.ts.mockups` sidecar names `.superpowers/sdd/experiments-prototype.html` as the frame. Run the project's `capture` command for this spec, then `compose-mockup-frames.sh` on the sidecar and `measure-visual-properties.sh` band, seam and matrix reads on every pair; fix until clean or each leftover is named with its cause.
  - [ ] **Step 3: Verify.** The spec passes under the project's `verify` command. Commit.

**Build:** green
**Files:** `stats/cmd/uitest-seed/seed.go`, `stats/web/tests/visual/experiments.spec.ts`, `stats/web/tests/visual/experiments.spec.ts.mockups`
**Allowed-collateral:** `stats/web/tests/visual/*.spec.ts-snapshots/**`
**Tests:** `experiments.spec.ts`
**Regression:** `experiments.spec.ts` fails — the view drifted from the approved prototype.
**Commit:** test(web): capture the Experiments view against its prototype
**After:** Task 7

- [ ] 9. Warn when the main agent runs off an open experiment's pair
  - [ ] **Step 1: Write the tests.** `stats/cmd/flow/decision_test.go`: `TestDecisionParentModel` — `parent.model` is accepted and rendered on the `## Decision` block's parent line, and a `parent.effort` of `xhigh` or `max` is accepted. `stats/cmd/flow/experiment_test.go`: `TestFlowExperimentMainCheck` — `flow experiment main-check -model opus -effort high` against an open `main=opus@medium` prints `⚠ experiment <id>: main agent should run opus@medium; this session runs opus@high` and exits 0; matching or no open `main` cell prints nothing, exit 0; an unreachable store prints nothing, exit 0. Run — fails.
  - [ ] **Step 2: Implement.** `stats/cmd/flow/decision.go`: `Parent` gains `Model string`; `parent.effort` also admits `xhigh` and `max`. `stats/cmd/flow/experiment.go`: `main-check`.
  - [ ] **Step 3: Edit the contracts.** In `skills/flow/SKILL.md`'s **Model resolution**, after the `VERIFY_MODEL=opus` fence, add: "**An open main-agent experiment is checked, never enforced.** Right after resolving the above, run `flow experiment main-check -model <this session's model> -effort <this session's effort>` and print any line it prints, once; it never blocks the run (decision `main-check-and-tag`, KAN-1067)." In `skills/flow/brainstorm-planner.md`, change "`parent` (the parent's own model/effort," to "`parent` (`{model, effort}`, the parent's own model/effort,". Wrap at each file's width.
  - [ ] **Step 4: Verify.** Tests pass; `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` exit 0; lint. Commit.

**Build:** green
**Files:** `stats/cmd/flow/decision.go`, `stats/cmd/flow/decision_test.go`, `stats/cmd/flow/experiment.go`, `stats/cmd/flow/experiment_test.go`, `skills/flow/SKILL.md`, `skills/flow/brainstorm-planner.md`
**Tests:** `TestDecisionParentModel`, `TestFlowExperimentMainCheck`
**Regression:** `TestFlowExperimentMainCheck` fails — a session off the experiment's main pair runs with no warning.
**Commit:** feat(stats): warn when the main agent runs off an open experiment's pair
**After:** Task 2, 6

**Decision:** main-check-and-tag

- [ ] 10. Close experiments with `/flow-experiment-close`
  - [ ] **Step 1: Write the skill.** `skills/flow-experiment-close/SKILL.md`, frontmatter `name: flow-experiment-close`, description "Analyse an open allowed-pairs experiment from its stored before/after stats, close it with a verdict and summary, and on a positive verdict start the /flow change that makes its winning cells permanent. Standalone, not a pipeline stage. Use for /flow-experiment-close." Body: argument `<id>` (none → `flow experiment list -open`; one → it; several → AskUserQuestion); read `flow experiment stats <id>`; with fewer than 5 changes in `after`, ask before closing (**Wait** recommended); judge each metric against its before figure — lower is better for every metric — and state `positive`, `negative` or `inconclusive` with the metrics that decided it; for a combined experiment, say which cells the evidence supports and that it cannot attribute an effect to one cell; `flow experiment close <id> -verdict <v> -summary '<summary>'`; on `positive`, ask which cells to make permanent (all supported cells recommended) and invoke `/flow` with "Make experiment <id>'s winning cells permanent in the Allowed pairs table: <role>=<pairs>; …". The summary is at most 8 lines: verdict, runs before/after, each decisive metric before → after, the cells recommended. <!-- predicted: the 5-change floor and 8-line cap are thresholds chosen at plan time, not measurements -->
  - [ ] **Step 2: Index it.** `CLAUDE.md`'s skill index gains a row: `skills/flow-experiment-close/` | `/flow-experiment-close` | "Closes an allowed-pairs experiment from its stored stats with a verdict and summary; on a positive verdict starts the /flow change that makes its winning cells permanent. Standalone, not a pipeline stage". If `setup.sh` installs skills from an explicit list, add it there. <!-- unverified: whether setup.sh enumerates skills — grep setup.sh for flow-settings -->
  - [ ] **Step 3: Verify.** `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-guard-symlinks.sh` exit 0; `SANDBOX="$(mktemp -d)"; HOME="$SANDBOX" ./setup.sh global` installs `flow-experiment-close`. Commit.

**Build:** green
**Files:** `skills/flow-experiment-close/SKILL.md`, `CLAUDE.md`
**Allowed-collateral:** `setup.sh`, `AGENTS.md`
**Tests:** **none** — a skill file; the guards in Step 3 are the check.
**Regression:** none — prose.
**Commit:** feat(flow-experiment-close): close an experiment with a verdict and summary
**After:** Task 6

**Decision:** close-skill

- [ ] 11. Document experiments in the model policy and the stats README
  - [ ] **Step 1: Edit.** `skills/flow-contracts/model-policy.md`, directly after the **Allowed pairs** table, a `## Experiments` section: "**An open experiment overrides whole cells of the table above.** `flow experiment open -name <n> -cell '<role>=<pair>[,<pair>…]'…` stores it in flowd; while it is open, every reader of the table (check-dispatch-pair, check-loop-dispatch, plan-agent-pair, plan-class, pipeline-fix) admits the experiment's pairs in place of that role's unmarked row, and a marked row is never overridden. One experiment may carry several roles; a role is in at most one open experiment. Role `main` names the parent session's `<model>@<effort>`: `/flow` warns on a mismatch (**Model resolution**, `skills/flow/SKILL.md`) and only matching runs count in its after figures. flowd unreachable → the table alone. The table stays the permanent source: `/flow-experiment-close` closes an experiment and, on a positive verdict, starts the change that writes its winning cells here (KAN-1067)." `stats/README.md`: an `### Experiments` section naming `flow experiment open|list|overrides|stats|main-check|close`, the routes and the Experiments view.
  - [ ] **Step 2: Verify.** `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-installed-citations.sh` exit 0; `cd stats && make -s web-dist && go test ./internal/modelpolicy -run 'TestParse|TestEmbeddedTableParses' -count=1` passes (the parser still finds the table). Commit.

**Build:** green
**Files:** `skills/flow-contracts/model-policy.md`, `stats/README.md`
**Tests:** **none** — docs; Step 2's suites re-parse the edited table.
**Regression:** none — prose.
**Commit:** docs(model-policy): document allowed-pairs experiments
**After:** Task 3, 9, 10

- [ ] 12. Add the `planner` role to the Allowed pairs table
  - [ ] **Step 1: Write the test.** In `stats/internal/guard/check_dispatch_pair_test.go`: `TestCheckDispatchPairAdmitsPlanner` — against the real `skills/flow-contracts/model-policy.md`, a prompt with `Role: planner` on `flow-planner-medium`/`opus` and on `flow-planner-low`/`opus` is admitted, on `flow-planner-high`/`opus` and on `flow-planner-medium`/`sonnet` refused. Run `cd stats && go test ./internal/guard -run 'TestCheckDispatchPairAdmitsPlanner' -count=1` — fails.
  - [ ] **Step 2: Edit.** In `skills/flow-contracts/model-policy.md`'s **Allowed pairs** table, after the `plan-designer` row, add `| `planner` | `opus/flow-planner-medium`, `opus/flow-planner-low` | — |`.
  - [ ] **Step 3: Verify.** The test passes; `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0; lint. Commit.

**Build:** green
**Files:** `skills/flow-contracts/model-policy.md`, `stats/internal/guard/check_dispatch_pair_test.go`
**Tests:** `TestCheckDispatchPairAdmitsPlanner`
**Regression:** `TestCheckDispatchPairAdmitsPlanner` fails — a planner dispatch is refused by the agent-baseline hook.
**Commit:** feat(model-policy): add the planner role to the allowed-pairs table
**After:** Task 3, 11

**Decision:** planner-role

- [ ] 13. Run `/flow`'s brainstorming and plan writing through the planner role
  - [ ] **Step 1: Edit `skills/flow/brainstorm.md`.** Replace the section `## Run brainstorming and planning directly` up to, not including, `**Prose preceding a question is shown too, not dropped.**` with a section `## Run brainstorming and planning through the planner role` stating, per decision `planner-role`:
    - the `flow.brainstorm` begin fence, kept;
    - "**The parent never brainstorms or writes the plan itself; a `planner` dispatch does** (decision `planner-role`, KAN-1067)." Every planner dispatch is fresh and one-shot, on `subagent_type: flow-planner-<effort>` with `model` and effort from the **Allowed pairs** row `planner` (`skills/flow-contracts/model-policy.md`) — its first pair unless an open experiment overrides it — its prompt carrying `Role: planner`, `Key: plan-<n>`, the two agent-baseline sentences, the report-tersely line, the PLANNING DELEGATION paragraph (`skills/flow-self-review/SKILL.md`), the change name, the absolute worktree, the session token and `REVIEWERS`; recorded with `flow record dispatch begin -change <name> -role planner -model <m> -effort <e> -key plan-<n> -session-token mf-<literal-token>` immediately before the Agent call and `end` once it returns, `<n>` counting this run's planner dispatches from 1;
    - **1. Design round** — the planner runs section **B** of `skills/flow/brainstorm-planner.md` without plan mode, reading the ticket and every file under `<abs-worktree>/.superpowers/sdd/planner/`, dispatching its own Explore and Plan agents (it records their rows, `-role explorer|planner`, `-key explore-<n>|design-<n>`), publishing or updating the prototype a visible change needs, and writes `round-<r>.md`: the design draft, the settled prose the convergence confirm shows, then `## Questions` — each question with two to four options, one marked recommended, the convergence confirm last when nothing else is open. It asks the operator nothing;
    - **2. Relay** — the parent prints the round's settled prose, asks its questions verbatim in one AskUserQuestion call, and writes the answers verbatim to `answers-<r>.md`; it never answers a question itself. An explicit approve of the convergence confirm closes brainstorming — `flow.brainstorm` end, `flow.design-approval` begin and end, as **Convergence** (`skills/flow/brainstorm-planner.md`) marks them; any other answer dispatches round `<r+1>`. An operator message mid-planning is written to the next `answers-<r>.md` and relayed the same way;
    - **3. Artifacts and plan** — a final planner dispatch runs sections **C** and **D** of `skills/flow/brainstorm-planner.md` up to, not including, **Decide**: its `flow.create-artifacts` and `flow.writing-plans` begin marks under the run's session token, `spectre new`, the three artifacts from the approved design, writing-plans, `spectre validate`, `check-plan-shape.sh`, the configured guards and `flow tasks count`; it returns `PLAN-WRITTEN` or the guard output it could not clear, and the parent redispatches a fresh one with that output until `PLAN-WRITTEN`;
    - the parent then closes `flow.writing-plans` and runs **Decide**, the record sequence and the gate, as the rest of the file states.
    Keep the "Run each `begin` immediately before its Agent call…" paragraph. Drop the plan-mode paragraph and the `No fork … thinking stays in this session` paragraph — the planner role replaces both; keep their kan-941 reason as one sentence: "No `fork` from kickoff to the plan gate: a fork copies the whole conversation and doubles the token spend (kan-941)."
  - [ ] **Step 2: Edit `skills/flow/brainstorm-planner.md`.** In the opening paragraph, replace "run by whichever session is executing `/flow`'s brainstorming stage" with "run by `/flow`'s `planner` dispatch (**Run brainstorming and planning through the planner role**, `skills/flow/brainstorm.md`)", and add after the paragraph: "In a `planner` dispatch, every **AskUserQuestion** these sections name is written as a question of the round file instead, and the parent asks it." Wrap at the file's width.
  - [ ] **Step 3: Update `skills/flow/resume.md`** where it resumes planning at `STARTED`: a resumed run re-enters the planner loop at the next round, reading `.superpowers/sdd/planner/`. <!-- unverified: resume.md's planning-resume paragraph wording — read it before editing -->
  - [ ] **Step 4: Verify.** `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-stage-mark-calls.sh`, `scripts/check-dispatch-paragraphs.sh` and `scripts/check-guard-symlinks.sh` exit 0; `git grep -n -F -e 'Run brainstorming and planning directly' -e 'the thinking stays in this session' -- skills` prints nothing. Commit.

**Build:** green
**Files:** `skills/flow/brainstorm.md`, `skills/flow/brainstorm-planner.md`, `skills/flow/resume.md`
**Allowed-collateral:** `skills/flow/*-rationale.md`, `skills/flow-contracts/pipeline.md`, `skills/flow/stage-keys.md`
**Tests:** **none** — contract prose; Step 4's guards are the check.
**Regression:** none — prose.
**Commit:** feat(flow): run brainstorming and plan writing through the planner role
**After:** Task 12

**Decision:** planner-role
