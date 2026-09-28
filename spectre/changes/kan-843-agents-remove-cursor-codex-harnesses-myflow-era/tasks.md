# kan-843-agents-remove-cursor-codex-harnesses-myflow-era

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Removes what is no longer used: Cursor/Codex support (tasks 2–4), the vocabulary guard (task 5),
myflow leftovers (tasks 6–7), stats legacy data shapes (task 1), ticket special cases (task 8), then
verifies live (task 9). `design.md` is canonical for every decision; each task cites its entry by
ID.

**"Live" excludes history.** Every sweep below skips `docs/self-review/` (except task 6's two
named edits), `spectre/changes/archive/`, `KNOWN-BUGS.md`, `.idea/`, `.worktrees/` and applied
migration files (`stats/internal/store/migrations/00[0-2]*.sql`, never edited).

**Word-level matching.** `cursor` is also an ordinary word (a parser's running cursor, CSS
`cursor: pointer`, a terminal cursor). Only harness mentions are removed; the sweep regex is
`git grep -niE '(^|[^a-z])(cursor|codex)([^a-z]|$)'`, read by hand.

**Per-task verify** runs the task's own tests plus `scripts/check-references.sh`; the full lint set
from `.flow/project.md`'s `## lint`, `scripts/run-guard-tests.sh`, `cd stats && go vet ./... &&
gofmt -l . && go test ./... -race -count=1` and `cd stats/web && npx tsc -b && npm test` run in
task 9 and `flow.verify`.

Live verification: task 9 runs the real installer in a sandbox HOME and applies the new migration
to a scratch copy of the dev store, recording before/after row counts.

## Review Focus

- The migration must be exact: every legacy row rewritten, no current-shape row touched, pricing
  outcomes unchanged for every row — including the `glm-5.3-flash` flat row that ZCode's collapsed
  cache-write tokens are priced against.
- `setup.sh global` on a HOME holding an earlier install that included `~/.cursor`/`~/.codex`:
  those directories are left alone (not deleted, not written).
- Deleting a guard must not leave a dangling reference: `check-references.sh`,
  `check-guard-symlinks.sh` and `check-installed-citations.sh` all pass.

---

- [x] 1. Stats: migrate legacy rows and remove the legacy readers

**Files:** `stats/internal/store/migrations/0031_drop_legacy_shapes.sql`, `stats/internal/store/legacyshapes_test.go`, `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`, `stats/internal/store/pricing.go`, `stats/internal/store/pricing_seed.go`, `stats/internal/store/pricing_seed_test.go`, `stats/internal/store/pricing_test.go`, `stats/internal/store/stageruns_test.go`, `stats/internal/stages/synthetic.go`, `stats/internal/stages/synthetic_test.go`
**Allowed-collateral:** `stats/internal/store/testsupport_test.go`, `stats/internal/store/migrations.go`
**Tests:** `TestMigration0031RewritesLegacyRows`
**Regression:** fails if 0031 leaves a `myflow stage begin (synthetic)` row, a bare-array
`panel.dispatches`/`groups` element, or a flat pricing row with a null 1h rate — or rewrites a
current-shape row.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/store/legacyshapes_test.go 2>/dev/null | grep -cE '^func Test' @ b8faae9a -->
**After:** none
**Commit:** `feat(stats): migrate legacy row shapes and drop their readers`
**Build:** green

**Decision:** stats-legacy-migrate-then-remove

Correction (2026-09-28): the plan declared `names.go`, `cmd/flow/state.go`, `state_test.go` and `api/stages.go`; none needed an edit (they name `stages.SyntheticChangeUpdatedBy` only, never its text), so they left `**Files:**`. `pricing_seed.go` also seeds `glm-5.3-flash`'s `CacheWrite1hPerMTok` as 0 — `flowd` re-upserts seed rows at startup, and a nil 1h would undo 0031's backfill. The migration test applies 0031 through the real migrator by pre-recording it as applied, seeding at 0030, then deleting the record — no prefix-apply helper was needed.

  - [x] **Step 1: Failing test.** In `stats/internal/store/legacyshapes_test.go`, write
    `TestMigration0031RewritesLegacyRows` against a fresh test database migrated through `0030`
    only, seeded with: a `changes` row updated by `myflow stage begin (synthetic)` and one by
    `/flow`; a decision whose `panel.dispatches` mixes a bare array and an object, one whose
    `groups` is bare arrays, and one already in the current shape; a pricing row with null 1h and
    collapsed = 5m, and one with null 1h and collapsed ≠ 5m. Apply `0031`, then assert each row's
    new value, the untouched rows byte-identical, and `cache_write_per_mtok` absent from
    `information_schema.columns`. unverified:whether migrations.go can apply a prefix of the
    migration list — if not, add an unexported-through-export_test.go helper that applies
    migrations up to a named file. Run `cd stats && go test ./internal/store/ -run
    '^TestMigration0031RewritesLegacyRows$' -count=1` — expect failure.
  - [x] **Step 2: Migration.** Write `0031_drop_legacy_shapes.sql`: the `changes.updated_by`
    rename; the two `jsonb` rewrites (`jsonb_agg` over `jsonb_array_elements … WITH ORDINALITY`,
    wrapping only elements whose `jsonb_typeof` is `array`, order preserved); the pricing backfill
    where `cache_write_1h_per_mtok IS NULL AND cache_write_per_mtok = cache_write_5m_per_mtok`;
    then `ALTER TABLE pricing DROP COLUMN cache_write_per_mtok`.
  - [x] **Step 3: Readers.** `SyntheticChangeUpdatedBy` = `flow stage begin (synthetic)` (and its
    comment in `names.go`, callers' comments in `state.go`, `api/stages.go`); drop the aggregate
    SQL's bare-array `ELSE` branch and its doc comment's legacy clauses; drop
    `PricingRate.CacheWritePerMTok`, its column in every `pricing` SELECT/INSERT, its seed values,
    and restate the flat rule in `pricing.go` as "the 1h rate equals the 5m rate". Update the
    tests that seeded or asserted the old shapes.
  - [x] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./... && go test ./internal/store/
    ./internal/stages/ ./internal/api/ ./cmd/flow/ -count=1 -race`.

- [x] 2. setup.sh: install for Claude Code and zcode only

**Files:** `setup.sh`, `scripts/test-setup.sh`, `commands/flow.md`, `commands/flow-fast.md`, `commands/flow-plan.md`, `commands/flow-self-review.md`, `commands/flow-settings.md`, `commands/flow-status.md`, `scripts/lib/owned-corpus.sh`, `scripts/test-check-normative-inventory.sh`, `skills/flow-contracts/pipeline-rationale.md`, `stats/internal/guard/check_installed_rules_test.go`, `stats/internal/guard/references.go`, `stats/internal/guard/installedcitations.go`
**Allowed-collateral:** `stats/internal/guard/libtwins_test.go`, `stats/internal/guard/check_references_test.go`, `stats/internal/guard/check_installed_citations_test.go`
**Tests:** `global install writes nothing under .cursor or .codex`, `no AGENTS.md is created for a project with no config`, `no AGENTS.md is created for a project naming no shared rule`
**Regression:** fails if a `global` run creates `~/.cursor` or `~/.codex` in the sandbox HOME, or
if `setup.sh cursor|codex|all` is still accepted.
**Baseline:** before=0 after=1
<!-- measured: cat scripts/test-setup.sh | grep -c 'global install writes nothing under .cursor or .codex' @ b8faae9a -->
**After:** none
**Commit:** `refactor(setup): install for Claude Code and zcode only`
**Build:** green

**Decision:** harnesses-claude-code-and-zcode

Correction (2026-09-28): `installer-sandbox-diff.sh`, its harness and `check-installed-rules.sh` mention neither harness nor `commands/`, so they left `**Files:**`; `ownedcorpus.go` does not exist (no Go twin). `pipeline-rationale.md` joined: its `commands/` citation failed `check-installed-citations.sh` once `commands/` was deleted.

  - [x] **Step 1: Failing case.** In `scripts/test-setup.sh`, add a case labelled `global install
    writes nothing under .cursor or .codex` (fresh HOME, `run_setup … global`, `assert_absent` on
    both directories) and a case asserting `setup.sh cursor`, `codex` and `all` each exit non-zero
    with the usage line. Run `scripts/test-setup.sh` — expect those failures.
  - [x] **Step 2: Installer.** Delete `install_cursor`, `install_codex`, `install_rules_cursor`,
    `COMMANDS_CURSOR_SRC`, the `cursor`/`codex`/`all` modes and usage text, the `.cursor`/`.codex`
    lines of `install_global` and the `~/.codex/AGENTS.md` managed file; `git rm -r commands/`.
  - [x] **Step 3: Harness and dependants.** Remove every Cursor/Codex case, assert and loop entry
    from `scripts/test-setup.sh` (the mode loops, `real_home_fingerprint`'s harness list,
    `source_tree_fingerprint`'s `commands/`), and `commands/` / `.cursor` / `.codex` from
    `owned-corpus.sh` (and its Go twin if it lists them), `installer-sandbox-diff.sh` and its
    harness, `check-installed-rules.sh`, `test-check-normative-inventory.sh`, and the Go guards
    `references.go` and `installedcitations.go` with their tests.
  - [x] **Step 4: Verify.** `scripts/test-setup.sh && scripts/test-installer-sandbox-diff.sh &&
    scripts/check-references.sh && scripts/check-installed-citations.sh && cd stats && go test
    ./internal/guard/ -count=1 -race`.

- [x] 3. Docs, rules and skills: drop Cursor and Codex

**Files:** `AGENTS.md`, `README.md`, `rules/agent-baseline.md`, `rules/flow-manual-review.mdc`, `skills/README.md`, `skills/flow-contracts/pipeline.md`, `skills/flow-contracts/pipeline-rationale.md`, `skills/flow-contracts/model-policy.md`, `skills/flow-contracts/project-configuration-authoring.md`, `skills/flow-contracts/project-configuration-rationale.md`, `skills/flow/principles-reviewer-prompt.md`, `scripts/check-stage-mark-calls.sh`, `stats/internal/guard/stagemarkcalls.go`, `.flow/project.md`
**Allowed-collateral:** `stats/internal/guard/check_stage_mark_calls_test.go`
**Tests:** none — prose and a guard message; `TestCheckStageMarkCalls` covers the message
**Regression:** none — no behaviour changes
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 2
**Commit:** `docs(skills): drop Cursor and Codex from docs, rules and skills`
**Build:** green

**Decision:** harnesses-claude-code-and-zcode

Correction (2026-09-28): with `all` gone, `check-installed-citations.sh` runs `setup.sh` three times (`global`, `claude-code`, `zcode`); `.flow/project.md` says so. The principles reviewer no longer auto-detects `.cursor/rules/*.mdc` — a behaviour change despite **Regression:** none; no project under `~/Projects` carries `.cursor/rules/`. The stage-mark port-parity subtest rewrites the old harness clause of the bash pinned at d71a2327 before its byte comparison.

  - [x] **Step 1: Trim.** Remove or restate every Cursor/Codex mention in the listed files. The
    `-harness` placeholder rule stays in `pipeline.md`, `check-stage-mark-calls.sh` and
    `stagemarkcalls.go`: the values become `claude-code` or `zcode`, the reason "one skill source
    installs into `~/.claude/skills/` and `~/.zcode/skills/`". `AGENTS.md` keeps its zcode role;
    its "Where a Codex session gets its rules" text is restated for zcode or cut. `.cursor/rules`
    example paths in the project-configuration files become `.claude/rules` or are cut.
  - [x] **Step 2: Verify.** `git grep -niE '(^|[^a-z])(cursor|codex)([^a-z]|$)' -- <the listed
    files>` prints nothing but ordinary-word hits; `scripts/check-references.sh &&
    scripts/check-stage-mark-calls.sh && cd stats && go test ./internal/guard/ -run
    '^TestCheckStageMarkCalls' -count=1`.

- [x] 4. Stats: drop Cursor and Codex from comments and fixtures

**Files:** `stats/README.md`, `stats/cmd/flow/record.go`, `stats/internal/harvest/attribute.go`, `stats/internal/harvest/watcher.go`, `stats/internal/records/render.go`, `stats/internal/records/types.go`, `stats/internal/store/records.go`
**Allowed-collateral:** `stats/cmd/flow/record_test.go`, `stats/internal/harvest/attribute_test.go`, `stats/internal/harvest/endtoend_test.go`, `stats/internal/records/render_test.go`, `stats/internal/store/records_test.go`, `stats/internal/store/aggregate_test.go`, `stats/internal/api/stages_test.go`
**Tests:** none — comments and fixture values only
**Regression:** none — no behaviour changes
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2
**Commit:** `docs(stats): drop Cursor and Codex from comments and fixtures`
**Build:** green

**Decision:** harnesses-claude-code-and-zcode

Correction (2026-09-28): `installedcitations.go` left `**Files:**` — task 2 already removed its harness mentions. Fixtures use `no-transcript-harness`, not `zcode`: `api/stages.go` reads zcode transcripts, so `zcode` would change what those tests mean.

  - [x] **Step 1: Trim.** Comments naming Cursor/Codex as harnesses that write no transcript are
    restated generically ("a harness that writes no transcript") or cut. Test fixtures using
    `codex`/`cursor` as a harness value switch to `zcode` or an obviously synthetic
    `no-transcript-harness`, keeping each test's meaning (a harness with no transcript stays one).
  - [x] **Step 2: Verify.** `cd stats && gofmt -l . && go vet ./... && go test ./internal/...
    ./cmd/... -count=1 -race`.

- [x] 5. Delete the vocabulary guard

**Files:** `scripts/check-vocabulary.sh`, `CLAUDE.md`, `AGENTS.md`, `.flow/project.md`, `README.md`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-guard-symlinks.sh`, `scripts/check-normative-inventory.sh`, `scripts/check-python-suppressions.sh`, `scripts/check-references.sh`, `scripts/check-visual-verification.sh`, `scripts/lib/coverage.sh`, `skills/flow-contracts/project-configuration-rationale.md`, `stats/internal/guard/dispatchparagraphs.go`, `stats/internal/guard/guardsymlinks.go`
**Allowed-collateral:** `scripts/test-check-vocabulary.sh`, `scripts/test-check-python-suppressions.sh`, `scripts/test-lib-coverage.sh`, `scripts/test-project-get.sh`, `scripts/test-setup.sh`, `stats/internal/guard/check_guard_symlinks_test.go`, `stats/internal/guard/check_installed_citations_test.go`, `stats/internal/guard/check_task_commit_fields_test.go`
**Tests:** none — deletion; remaining guards' tests cover the references
**Regression:** none — the guard is removed
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 2, 3
**Commit:** `refactor(scripts): delete the vocabulary guard`
**Build:** green

**Decision:** delete-vocabulary-guard

  - [x] **Step 1: Delete.** `git rm scripts/check-vocabulary.sh scripts/test-check-vocabulary.sh`;
    remove every reference in the listed files (lint command lists, sibling-guard comments,
    exemption lists, coverage tables, Go guard lists and their test fixtures). A fixture that used
    the file only as an example path switches to another existing guard.
  - [x] **Step 2: Verify.** `git grep -n 'check-vocabulary\|vocab-guard' -- . ':!docs/'
    ':!spectre/changes/archive/'` prints nothing; `scripts/check-references.sh &&
    scripts/check-guard-symlinks.sh && scripts/test-lib-coverage.sh && cd stats && go test
    ./internal/guard/ -count=1 -race`.

- [x] 6. Self-review guard: no legacy labels, no exemptions

**Files:** `scripts/check-self-review-report.sh`
**Allowed-collateral:** `docs/self-review/**`, `scripts/test-check-self-review-report.sh`
**Tests:** none — cases for the removed branches are deleted; the full-corpus run is the check
**Regression:** none — the guard only loses acceptances
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** none
**Commit:** `refactor(self-review): drop legacy labels and pre-shape exemptions`
**Build:** green

**Decision:** self-review-delete-17-relabel-14

Correction (2026-09-28): 13 reports were relabelled, not 14 — `kan-197-require-mutation-test-for-every-guard-self-review.md` carried a `myflow-` label and was also one of the 17 deleted. The guard's OK line drops its "declared pre-rule" count.

  - [x] **Step 1: Reports.** `git rm` every file named in `DECLARED_REPORTS`; in every report
    `git grep -lE 'myflow-(fix|cost|improvement|automation)' docs/self-review` lists, rewrite each
    `myflow-<angle>` label to `flow-<angle>`.
    <!-- measured: git grep -lE 'myflow-(fix|cost|improvement|automation)' docs/self-review | wc -l → 14; DECLARED_REPORTS has 17 entries @ b8faae9a -->
  - [x] **Step 2: Guard.** Delete `LEGACY_ANGLE_LABELS`, its match branch, `DECLARED_REPORTS`,
    `DECLARED_REASON`, `declare_pre_rule` and the header prose describing them; delete the harness
    cases that exercised them.
  - [x] **Step 3: Verify.** `scripts/check-self-review-report.sh && scripts/test-check-self-review-report.sh`.

- [x] 7. Remove myflow leftovers

**Files:** `skills/flow/brainstorm-planner.md`, `skills/flow-contracts/finish-contract-run2.md`, `scripts/plan-dispatch-bundles.py`, `skills/flow-contracts/workspace-isolation.md`, `.gitignore`
**Allowed-collateral:** `scripts/test-prepare-workspace.sh`, `stats/cmd/flow/workspaceid_test.go`
**Tests:** none — renamed fixture values in existing tests
**Regression:** none — no behaviour changes
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 6
**Commit:** `docs(skills): remove myflow leftovers`
**Build:** green

  - [x] **Step 1: Prose.** Drop the "earlier `myflow-` spellings matched too" clauses in
    `brainstorm-planner.md` and `finish-contract-run2.md`, the myflow wording in
    `plan-dispatch-bundles.py`'s header (the two `skills/*/scripts/` entries are symlinks to it)
    and `.gitignore`'s comment.
  - [x] **Step 2: Worked example.** Rename `kan-15-parallel-myflow-do-task-lanes` to
    `kan-15-parallel-flow-task-lanes` in `workspace-isolation.md`, `test-prepare-workspace.sh` and
    `workspaceid_test.go`; recompute the expected id with `cd stats && go run ./cmd/flow
    workspace-id kan-15-parallel-flow-task-lanes` and update every place that quotes it.
  - [x] **Step 3: Verify.** `git grep -il myflow -- . ':!docs/' ':!spectre/changes/archive/'
    ':!spectre/changes/kan-843-agents-remove-cursor-codex-harnesses-myflow-era/' ':!KNOWN-BUGS.md' ':!.idea/'` prints nothing;
    Correction (2026-09-28): it prints task 1's `synthetic.go`, `synthetic_test.go`, `legacyshapes_test.go` and `0031_drop_legacy_shapes.sql`, which must name the migrated `myflow` value; `scripts/test-prepare-workspace.sh && cd stats
    && go test ./cmd/flow/ -run '^TestWorkspaceID' -count=1`.

- [x] 8. Drop ticket special cases

**Files:** `scripts/check-model-resolution-shell.sh`, `stats/internal/store/stageruns.go`
**Allowed-collateral:** `scripts/test-check-model-resolution-shell.sh`, `scripts/test-setup.sh`
**Tests:** none — a group rename, a message and a comment
**Regression:** none — no behaviour changes
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 2, 5
**Commit:** `refactor(scripts): drop ticket-named test group and messages`
**Build:** green

**Decision:** drop-ticket-special-cases

  - [x] **Step 1: Edit.** Rename `test-setup.sh`'s "KAN-369: __pycache__ churn" group and its
    tree_fingerprint comment by behaviour ("__pycache__ churn leaves the source fingerprint
    unchanged"); restate `check-model-resolution-shell.sh:100`'s message without kan-488 (and its
    harness's expectation if it matches the text); delete the `// KAN-185` comment on
    `stageRunSupersedeLockNamespace`, value unchanged.
  - [x] **Step 2: Verify.** `scripts/test-setup.sh && scripts/test-check-model-resolution-shell.sh
    && cd stats && go build ./...`.

- [x] 9. Live verification

**Files:** none
**Tests:** none — measurement task; the figures it records are the check
**Regression:** none — no commit
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2, 3, 4, 5, 6, 7, 8
**Build:** green

  - [x] **Step 1: Installer.** `SANDBOX=$(mktemp -d); HOME=$SANDBOX ./setup.sh global`; record the
    exit code and `ls -a $SANDBOX` — `.claude` and `.zcode` present, no `.cursor`, no `.codex`.
    Failure looks like either directory existing.
  - [x] **Step 2: Migration against real data.** Inside `flow-postgres`, `pg_dump` the `flow`
    database into a new scratch database `flow_kan843_verify` (the dev `flow` database is read,
    never written), record the Context counts there (205 / 16 / 4 / 1), start the worktree's
    `flowd` build against it on a spare port only long enough to run migrations (or apply `0031`
    with `psql`), re-count — expected 0 / 0 / 0 / 0 legacy rows, the same total row counts, and
    `cache_write_per_mtok` gone — then drop `flow_kan843_verify`. Failure looks like any legacy
    count above zero or a total that moved. Never stop, restart or migrate the dev `flowd` or its
    `flow` database.
  - [x] **Step 3: Full suite.** Every `## lint` command in `.flow/project.md`,
    `scripts/run-guard-tests.sh`, `cd stats && go vet ./... && gofmt -l . && go test ./... -race
    -count=1`, `cd stats/web && npx tsc -b && npm test`; record pass/fail and
    `scripts/test-setup.sh`'s wall time (KAN-844's post-change baseline) in `design.md` under a
    `## Measurements` section.
