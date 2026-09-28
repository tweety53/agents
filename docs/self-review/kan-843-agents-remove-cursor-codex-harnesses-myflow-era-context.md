# Self-review context bundle for kan-843-agents-remove-cursor-codex-harnesses-myflow-era

found: 6 of 7 sources; skipped: 1 of 7 sources
skipped: change summary (absent)

## .superpowers/sdd/ledgers/kan-843-agents-remove-cursor-codex-harnesses-myflow-era.md

# SDD ledger — kan-843-agents-remove-cursor-codex-harnesses-myflow-era

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — implementer

- Task: 1
- Role: implementer
- Key: task-1-implementer
- Model: opus effort=high
- Commit: c4c2d12e37b10ec7b1da1a41ab4eb948d6727c16
- Outcome: completed
- Started: 2026-09-27T23:31:52Z
- Tokens: input 104, output 946, cache read 3870861, cache creation 150860

## Dispatch 2 — implementer

- Task: no task
- Role: implementer
- Key: task-2+3+5+8-implementer
- Model: opus effort=medium
- Commit: d80fc3a0
- Outcome: completed
- Started: 2026-09-27T23:31:52Z
- Tokens: input 184, output 5987, cache read 12037763, cache creation 220580

## Dispatch 3 — implementer

- Task: no task
- Role: implementer
- Key: task-6+7-implementer
- Model: opus effort=low
- Commit: 3ae22eff
- Outcome: completed
- Started: 2026-09-27T23:31:52Z
- Tokens: input 74, output 2920, cache read 1894925, cache creation 86148

## Dispatch 4 — reviewer

- Task: 1
- Role: reviewer
- Key: task-1-reviewer
- Model: opus effort=high
- Commit: no commit
- Outcome: clean
- Started: 2026-09-27T23:44:23Z
- Tokens: input 64, output 3718, cache read 1802825, cache creation 86421

## Dispatch 5 — reviewer

- Task: no task
- Role: reviewer
- Key: task-2+3+5-reviewer
- Model: opus effort=medium
- Commit: no commit
- Outcome: clean
- Started: 2026-09-27T23:56:50Z
- Tokens: input 66, output 573, cache read 2619234, cache creation 126141

## Dispatch 6 — reviewer

- Task: no task
- Role: reviewer
- Key: task-6+7-reviewer
- Model: opus effort=low
- Commit: no commit
- Outcome: clean
- Started: 2026-09-27T23:56:50Z
- Tokens: input 24, output 192, cache read 368765, cache creation 41016

## Dispatch 7 — implementer

- Task: 4
- Role: implementer
- Key: task-4-implementer
- Model: opus effort=low
- Commit: 68586556
- Outcome: completed
- Started: 2026-09-27T23:56:50Z
- Tokens: input 22, output 527, cache read 398535, cache creation 49505

## Dispatch 8 — reviewer

- Task: 4
- Role: reviewer
- Key: task-4-reviewer
- Model: opus effort=low
- Commit: no commit
- Outcome: clean
- Started: 2026-09-28T00:01:48Z
- Tokens: input 10, output 87, cache read 104785, cache creation 26039

## Dispatch 9 — implementer

- Task: 9
- Role: implementer
- Key: task-9-implementer
- Model: opus effort=medium
- Commit: no commit
- Outcome: completed
- Started: 2026-09-28T00:01:48Z
- Tokens: input 48, output 2477, cache read 739859, cache creation 43051

## Dispatch 10 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: opus effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-28T00:16:56Z
- Tokens: input 148, output 4035, cache read 11651967, cache creation 262150

## Dispatch 11 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-0-principles-reproducer-bounce
- Model: opus effort=low
- Commit: no commit
- Outcome: completed
- Started: 2026-09-28T00:33:34Z
- Tokens: input 26, output 317, cache read 286347, cache creation 29564

## Dispatch 12 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: opus effort=medium
- Commit: 9e208a61
- Outcome: completed
- Started: 2026-09-28T00:35:52Z
- Tokens: input 54, output 3543, cache read 1230888, cache creation 63025

## Dispatch 13 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: opus effort=low
- Commit: no commit
- Diff base: 49c91492e055a4930d0cc6ad133d9e1a3b28fd9f
- Outcome: completed
- Started: 2026-09-28T00:40:31Z
- Tokens: input 14, output 754, cache read 148027, cache creation 28813

## Dispatch 14 — verifier

- Task: no task
- Role: verifier
- Key: verify
- Model: claude-opus-5-5 effort=default
- Commit: no commit
- Outcome: completed
- Started: 2026-09-28T00:42:18Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-843-agents-remove-cursor-codex-harnesses-myflow-era-panel.md

# Review panel — kan-843-agents-remove-cursor-codex-harnesses-myflow-era

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | stats/internal/store/migrations/0031_drop_legacy_shapes.sql:41 | The pricing backfill treats collapsed = 5m as published-flat, but 0007 copied the collapsed column into the 5m rate on every older row, so 0031 invents a 1h rate equal to 5m on every pre-0007 row: 1h cache writes that used to refuse now price at the 5m rate, understating cost. |   |
| F2 | primary | Important | skills/flow/brainstorm-planner.md:18 | The verify-the-defect-still-exists step now fires only on flow-fix/flow-cost labels, while Jira issues filed under myflow- labels were never relabelled, so a /flow run on one skips the check. |   |
| F3 | principles | Minor | stats/internal/store/migrations/0031_drop_legacy_shapes.sql:41 | glm-5.3-flash's 1h rate has two writers (the startup seed upsert and 0031's backfill), and the backfill's collapsed = 5m rule is a second definition of flat beside the pricing code's 1h = 5m. |   |
| F4 | principles | Minor | scripts/test-setup.sh:945 | The nothing-to-render group now seeds AGENTS.md and lost the no-AGENTS.md-is-created assertion, so a mutant that creates AGENTS.md in a project with no config passes the harness. |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 withdrawn premise unsupported: Jira holds no open myflow-fix/myflow-cost issue (only KAN-264, myflow-automation, which the check never covered); new issues carry flow- labels
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-principles-2.sh

## Pass log

### Round 0

- roster: compact — 33
- diff size: 5528 changed lines, over cap — proceeded unasked; docs-only: exit 1 (.gitignore) — full roster primary+principles; no operator-added slot this round — the resolved list ran alone; base rebased onto 4bb54bd5 (operator chose Rebase; KNOWN-BUGS.md append conflict resolved keeping both sides on operator instruction); standards: CLAUDE.md, AGENTS.md
- F2 withdrawn on primary-source check: JQL labels in (myflow-*) AND statusCategory != Done returns only KAN-264 (myflow-automation). F1: operator chose "Drop the backfill" — design decision stats-legacy-migrate-then-remove superseded.
- F4 bounced once to principles: malformed premise (4bb54bd5: prefix); repaired, now demonstrated

### Round 1

- panel-fix-1 (opus/medium) fixed F1,F3,F4 in f9e8f326+9e208a61; F3,F4 reproducers flipped (pinned); F1 pinned re-run refused ambiguous (premise names the deleted backfill line) — routed to primary for re-authoring; plan-unchanged flagged the fixer's PLAN FIELDS Tests: edit to task 2, hand-verified
- primary re-run (opus/low, fix-round-1.diff): F1 fixed, no new finding; reproducer re-authored and proven both legs (sha 57b44888…); principles not re-run — raised only Minors, both fixed
fix-mutation: stats/internal/store/migrations/0031_drop_legacy_shapes.sql — re-inserted the 1h backfill UPDATE — TestMigration0031RewritesLegacyRows
fix-mutation: scripts/test-setup.sh — setup.sh install_project_standards creates AGENTS.md when absent — no AGENTS.md is created for a project with no config
fix-mutations-total: 2
## spectre/changes/archive/kan-843-agents-remove-cursor-codex-harnesses-myflow-era/tasks.md

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
## spectre/changes/archive/kan-843-agents-remove-cursor-codex-harnesses-myflow-era/design.md

# Design — kan-843-agents-remove-cursor-codex-harnesses-myflow-era

## Context

Measured on `main` at `b8faae9a` (2026-09-28):

- `setup.sh` modes: `cursor | claude-code | codex | zcode | all | global`. zcode reuses
  `commands-claude/`, copies `AGENTS.md` per project and renders `~/.zcode/AGENTS.md`.
- `scripts/check-vocabulary.sh` guards retired myflow stage names and the old panel roster only.
- `scripts/check-self-review-report.sh` scans every file under `docs/self-review/`; it exempts 17
  named pre-shape reports and accepts `myflow-<angle>` labels, which 14 reports carry.
- Dev store (`flow` DB, read-only query): 205 `changes` rows with `updated_by = 'myflow stage begin
  (synthetic)'`; 16 decisions with bare-array `panel.dispatches` elements and 4 with bare-array
  `groups` elements (of 77); 1 pricing row (`glm-5.3-flash`) with a null 1h rate, its collapsed
  `cache_write_per_mtok` equal to its 5m rate (0 = 0).

## Decisions

### Keep Claude Code and zcode only

**ID:** harnesses-claude-code-and-zcode
**Status:** active
**Chosen:** delete Cursor and Codex support outright — `commands/`, `install_cursor`,
`install_codex`, `install_rules_cursor`, the `cursor`/`codex`/`all` modes, `~/.cursor`/`~/.codex`
installs and the `~/.codex/AGENTS.md` block — and trim every live mention. `AGENTS.md` stays
(zcode). The `-harness` placeholder rule stays, its reason restated for `~/.claude` and `~/.zcode`.
**Considered:** keeping `all` as an alias for claude-code+zcode — rejected, nobody uses it and
`global` already covers both.

### Delete the vocabulary guard

**ID:** delete-vocabulary-guard
**Status:** active
**Chosen:** delete `scripts/check-vocabulary.sh` and `scripts/test-check-vocabulary.sh` and every
reference (lint lists, sibling guards, Go guards, coverage lib, tests).
**Considered:** porting it to Go — rejected, the vocabulary it guards is retired.

### Self-review reports: delete 17, relabel 14

**ID:** self-review-delete-17-relabel-14
**Status:** active
**Chosen:** the guard loses its legacy-label branch and exemption list; the 17 pre-shape reports
are deleted (git history keeps them) and `myflow-<angle>` → `flow-<angle>` in the 14 others.
**Considered:** keeping the exemptions (legacy stays); deleting all 31 reports — the operator chose
this option.

### Stats: migrate legacy rows, then remove the readers

**ID:** stats-legacy-migrate-then-remove
**Status:** superseded by stats-legacy-migrate-no-pricing-backfill
**Chosen:** one data migration and the code removal together:
- `changes.updated_by` `'myflow stage begin (synthetic)'` → `'flow stage begin (synthetic)'`;
  `stages.SyntheticChangeUpdatedBy` follows.
- `decisions`: every bare-array element of `panel.dispatches` becomes `{"slots": <array>}`, of
  `groups` becomes `{"bundles": <array>}`; the aggregate SQL drops its bare-array branch.
- `pricing`: where the 1h rate is null and the collapsed column equals the 5m rate, set 1h = 5m;
  drop `cache_write_per_mtok`. `PricingRate.CacheWritePerMTok`, its seed values and the
  "nil 1h + collapsed column equals 5m" flat rule go; the flat rule is "1h equals 5m". Every row
  prices exactly as before.
- Untouched: the decision `implementer` string form (still written today as
  `"skipped — inline"`) and generic absent-key handling (`fixer`).
The migration reaches the dev store when the operator next starts `flowd` on the new binary; no
agent restarts it. Earlier migration files are never edited.
**Considered:** leaving the compat readers — the operator chose migrate-and-remove.
**Superseded because:** the panel (F1) showed the pricing backfill's `collapsed = 5m` test matches
every pre-0007 row, since 0007 copied the collapsed column into the 5m rate on all of them, so it
would invent a 1h rate and under-price 1h cache writes there. "Every row prices exactly as before"
cannot hold for that shape either way: the old schema meant "flat for an unknown split, refuse an
explicit 1h", which one column cannot express.

### Stats: migrate legacy rows without a pricing backfill

**ID:** stats-legacy-migrate-no-pricing-backfill
**Status:** active
**Chosen:** as `stats-legacy-migrate-then-remove`, except that 0031 does not backfill any 1h rate;
it only drops `cache_write_per_mtok`. The pricing seed is the single writer of
`glm-5.3-flash`'s 1h rate (0, set by the startup upsert that runs right after migrations). A
pre-0007 row the seed does not cover keeps a null 1h rate and so refuses an unknown-split cache
write rather than pricing it at the 5m rate: cost is never understated. The dev store has no such
row (task 9 measured it).
**Considered:** narrowing the backfill to `glm-5.3-flash` — rejected, it duplicates the seed;
keeping the backfill as designed — rejected, it understates 1h cache writes on pre-0007 rows.

### Ticket special cases

**ID:** drop-ticket-special-cases
**Status:** active
**Chosen:** remove code that special-cases a ticket or names a test by ticket: the self-review
exemption list, `test-setup.sh`'s "KAN-369" group (renamed by behaviour), the kan-488 wording in
`check-model-resolution-shell.sh`'s message, the `// KAN-185` comment beside an unchanged value.
**Considered:** stripping all ~3,000 inline `(KAN-NNN)` citations — the operator kept them.

## Open questions

## Measurements

Task 9, at the branch tip (2026-09-28):

- **Installer:** `HOME=<sandbox> ./setup.sh global` → exit 0; the sandbox holds `.claude`, `.zcode` (and `.zshrc`), no `.cursor`, no `.codex`. With a pre-existing `.cursor/x` and `.codex/y`, both are byte-identical afterwards and nothing new appears under either.
- **0031 on a copy of the dev store** (`flow_kan843_verify`, applied through the real migrator, then dropped):

  | Measure | Before | After |
  |---|---|---|
  | `changes` updated by `myflow stage begin (synthetic)` | 205 | 0 |
  | decisions with a bare-array `panel.dispatches` element | 16 | 0 |
  | decisions with a bare-array `groups` element | 4 | 0 |
  | flat pricing rows with a null 1h rate | 1 | 0 |
  | total `changes` / `decisions` / `pricing` | 441 / 78 / 9 | 441 / 78 / 9 |

  `cache_write_per_mtok` is gone; `glm-5.3-flash`'s 1h rate went NULL → 0 (= its 5m rate). Re-pricing all 553 glm stage runs: 545 byte-identical; the other 8 were never priced in the live store and price the same under main's code (4bb54bd5); run 4114 fails identically on both (`<synthetic>` model has no pricing).
- **Full suite:** every `## lint` command, `scripts/run-guard-tests.sh` (57/57, 147s), `stats/web` `npm test` and `tsc -b` pass. `go test ./... -race -count=1` passes except `internal/reconcile`'s `TestConcurrentAppendVersusRetirePreservesEveryEntry` — the pre-existing race already in `KNOWN-BUGS.md` (kan-842); the package is untouched here and 5 reruns pass.
- **`scripts/test-setup.sh` wall time:** 25s (KAN-844's post-change baseline).
## spectre/changes/archive/kan-843-agents-remove-cursor-codex-harnesses-myflow-era/narrative.md

# kan-843-agents-remove-cursor-codex-harnesses-myflow-era — session narrative

## 2026-09-28 — creating run

- Resumed at `STARTED` with the plan already committed; implemented in three waves (groups 1, 2 and 4 in throwaway worktrees, then group 3, then task 9's live checks).
- Implementers disclosed `**Files:**` corrections on tasks 1, 2 and 4 (declared files needing no edit; `pipeline-rationale.md` joining task 2 after `commands/` deletion broke its citation); each was judged legitimate on the guard's refusal and transcribed as a dated Correction.
- `check-plan-unchanged.sh` flagged two dispatches whose only plan-tree change was the parent's own tick (task 4, mid-review) and the fixer's PLAN FIELDS edit; both hand-verified from git.
- The first full-suite run hit the known `internal/reconcile` append-vs-retire flake (KNOWN-BUGS, kan-842); untouched here, 5 reruns clean, and the post-fix verify run was fully green.
- At panel entry `origin/main` had moved 20 commits (kan-798, kan-839) overlapping only `KNOWN-BUGS.md`. The operator chose Rebase; the append-vs-append conflict was then resolved on the operator's instruction by keeping both sides, and the branch was force-pushed with lease.
- Panel round 0: F1 (Important) — 0031's pricing backfill would fabricate a 1h rate on every pre-0007 row. The operator chose to drop the backfill; design decision superseded by `stats-legacy-migrate-no-pricing-backfill`. F2 withdrawn after a JQL check found no open `myflow-fix`/`myflow-cost` issue. F3/F4 (Minors) fixed beside F1. F4's reproducer bounced once for a malformed premise; F1's was re-authored after the fix deleted its premise line.
- Task 9 measured 205/16/4/1 legacy rows → 0 on a scratch copy of the dev store (dropped afterwards); 0031 reaches the dev store only when the operator next restarts `flowd`.

## 2026-09-28 — integrate run

- Preflight `RUN1`; unfinished-work and visual-verify gates clear; main checkout staged-clean and drift-clean.
- `origin/main` had moved 19 commits (the withdraw-changes change among them), overlapping `AGENTS.md`, `KNOWN-BUGS.md`, `rules/flow-manual-review.mdc`, `skills/flow-contracts/pipeline.md` and `skills/flow/brainstorm-planner.md`. Rebased without a prompt; only `KNOWN-BUGS.md` conflicted, twice — task-1's defer commit appended beside upstream's kan-797 entries (kept both), and the pricing fix commit removed its own now-fixed entry (kept upstream's entries, dropped that one line). The other four overlaps merged cleanly.
- Resolution triggered the full `## lint` and `## test` lists: every guard, `go vet`, `gofmt`, `tsc`, 57 guard harnesses, Go tests and SPA tests passed, except `check-task-records.sh`, which fails identically on `main` at `db691488` — the unarchived `withdraw-changes-abandoned-before-planning` change's ticked tasks name per-task commit subjects its squash-merged landing does not carry. Not introduced here; left to that change's archive.
- Landing route `merge and push` taken from the project's configured default, not asked.

## 2026-09-28 — integrate run

Preflight clean: no foreign staged work or drift on the main checkout, `RUN1`. The unfinished-work
gate reported `CLEAR` and `VISUAL-VERIFY-OK` (no UI paths touched).

`origin/main` had moved 31 commits since the recorded merge base, overlapping 35 paths. The rebase
onto `b6e19e5f` stopped once, on `KNOWN-BUGS.md`: main appended kan-820's deferred Minors and this
change appended its own deferred Minors at the same spot. Resolved by keeping both sets of entries,
main's first. Nothing was set aside — the worktree had no uncommitted planning artifacts.

Because the rebase needed resolution, the whole `## lint` and `## test` lists ran. Every test suite
passed (57/57 guard harnesses, `go test -race`, 170 SPA tests) and every lint step exited clean
except `check-task-records.sh`, which also fails on `main` itself: it flags ticked tasks whose
per-task commits a previous integrate collapsed (kan-820, withdraw-changes-abandoned-before-planning,
and this change). A structural false positive of squashed history, not something the rebase
introduced. The re-run of `check-base-moved.sh` against the rebased tip reported `CLEAR`.

Landing route: merge and push, from the project's configured default, not asked.
## git log --stat

commit 52f9660b0418d61a7d603ff037d55a5622124ac5
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 04:09:15 2026 +0300

    refactor(setup): drop Cursor and Codex harnesses and myflow-era leftovers

 .flow/project.md                                   |  11 +-
 .gitignore                                         |   2 +-
 AGENTS.md                                          |  48 +-
 CLAUDE.md                                          |   4 +-
 KNOWN-BUGS.md                                      |   7 +
 README.md                                          |  33 +-
 commands/flow-fast.md                              |  39 --
 commands/flow-plan.md                              |  14 -
 commands/flow-self-review.md                       |  26 -
 commands/flow-settings.md                          |  22 -
 commands/flow-status.md                            |  21 -
 commands/flow.md                                   |  48 --
 ...08-24-kan-284-store-rejects-same-state-write.md |  34 +-
 ...w-get-rid-of-staging-use-commits-self-review.md |  75 ---
 ...ow-skills-cut-meta-prose-extract-self-review.md | 125 -----
 ...emove-manual-test-guide-and-gate-self-review.md | 126 -----
 ...nd-token-cost-of-a-myflow-do-run-self-review.md | 165 -------
 ...myflow-agent-token-and-time-cost-self-review.md | 109 ----
 ...hter-auto-code-review-by-default-self-review.md | 160 ------
 .../self-review/kan-111-myflow-fast-self-review.md |  75 ---
 ...myflow-planning-and-status-fixes-self-review.md |  77 ---
 ...15-parallel-myflow-do-task-lanes-self-review.md | 171 -------
 .../kan-153-kan-108-follow-up-self-review.md       | 131 -----
 .../kan-16-myflow-stats-app-self-review.md         | 120 -----
 ...uild-does-not-build-the-binaries-self-review.md |  42 +-
 ...re-mutation-test-for-every-guard-self-review.md | 112 -----
 ...self-review-filing-ask-per-angle-self-review.md |  44 +-
 ...-rediscovery-across-review-panel-self-review.md |  32 +-
 ...2-commit-split-and-module-scopes-self-review.md |  38 +-
 ...patch-cost-tokens-model-and-role-self-review.md |  36 +-
 .../kan-23-myflow-self-review-self-review.md       | 106 ----
 ...ranch-resolution-retyped-by-hand-self-review.md |  30 +-
 .../kan-258-store-native-run-record-self-review.md |  38 +-
 ...an-265-be-brief-in-repo-markdown-self-review.md |  38 +-
 ...line-load-cost-split-by-consumer-self-review.md |  40 +-
 ...l-code-review-slot-hangs-on-fork-self-review.md |  30 +-
 ...ew-model-overridable-per-project-self-review.md |  30 +-
 ...l-guard-scripts-alongside-skills-self-review.md | 127 -----
 ...kan-77-sdd-ledger-canonical-path-self-review.md |  40 +-
 ...yflow-per-command-token-overhead-self-review.md | 186 -------
 ...-87-cut-per-command-load-further-self-review.md | 198 --------
 ...5-slim-the-myflow-contract-files-self-review.md | 134 -----
 rules/agent-baseline.md                            |   2 +-
 rules/flow-manual-review.mdc                       |   4 +-
 scripts/check-dispatch-paragraphs.sh               |   7 +-
 scripts/check-guard-symlinks.sh                    |   2 +-
 scripts/check-model-resolution-shell.sh            |   2 +-
 scripts/check-normative-inventory.sh               |   6 +-
 scripts/check-python-suppressions.sh               |   3 +-
 scripts/check-references.sh                        |   4 -
 scripts/check-self-review-report.sh                | 145 +-----
 scripts/check-stage-mark-calls.sh                  |  12 +-
 scripts/check-visual-verification.sh               |   3 +-
 scripts/check-vocabulary.sh                        | 549 ---------------------
 scripts/lib/coverage.sh                            |   7 +-
 scripts/lib/owned-corpus.sh                        |   4 +-
 scripts/plan-dispatch-bundles.py                   |   2 +-
 scripts/test-check-normative-inventory.sh          |   5 +-
 scripts/test-check-python-suppressions.sh          |   5 +-
 scripts/test-check-self-review-report.sh           |  73 +--
 scripts/test-check-vocabulary.sh                   | 275 -----------
 scripts/test-lib-coverage.sh                       |   4 +-
 scripts/test-prepare-workspace.sh                  |  18 +-
 scripts/test-project-get.sh                        |   8 +-
 scripts/test-setup.sh                              | 168 ++++---
 setup.sh                                           | 154 ++----
 skills/README.md                                   |   2 +-
 skills/flow-contracts/finish-contract-run2.md      |   6 -
 skills/flow-contracts/model-policy.md              |   5 -
 skills/flow-contracts/pipeline-rationale.md        |   9 +-
 skills/flow-contracts/pipeline.md                  |   9 +-
 .../project-configuration-authoring.md             |   2 +-
 .../project-configuration-rationale.md             |   7 +-
 skills/flow-contracts/workspace-isolation.md       |  18 +-
 skills/flow/brainstorm-planner.md                  |   4 +-
 skills/flow/principles-reviewer-prompt.md          |   6 +-
 stats/README.md                                    |   4 +-
 stats/cmd/flow/record.go                           |   2 +-
 stats/cmd/flow/record_test.go                      |   2 +-
 stats/cmd/flow/workspaceid_test.go                 |   8 +-
 stats/internal/api/stages_test.go                  |  14 +-
 stats/internal/guard/check_guard_symlinks_test.go  |   2 +-
 .../guard/check_installed_citations_test.go        |  17 +-
 stats/internal/guard/check_installed_rules_test.go |  52 +-
 .../internal/guard/check_stage_mark_calls_test.go  |  13 +-
 .../guard/check_task_commit_fields_test.go         |   2 +-
 stats/internal/guard/dispatchparagraphs.go         |   6 +-
 stats/internal/guard/guardsymlinks.go              |   3 +-
 stats/internal/guard/installedcitations.go         |  21 +-
 stats/internal/guard/references.go                 |   6 +-
 stats/internal/guard/stagemarkcalls.go             |   2 +-
 stats/internal/harvest/attribute.go                |  10 +-
 stats/internal/harvest/attribute_test.go           |   4 +-
 stats/internal/harvest/endtoend_test.go            |   8 +-
 stats/internal/harvest/watcher.go                  |   6 +-
 stats/internal/records/render.go                   |   2 +-
 stats/internal/records/render_test.go              |   4 +-
 stats/internal/records/types.go                    |   4 +-
 stats/internal/stages/synthetic.go                 |   5 +-
 stats/internal/stages/synthetic_test.go            |  19 +-
 stats/internal/store/aggregate.go                  |  25 +-
 stats/internal/store/aggregate_test.go             |  51 +-
 stats/internal/store/legacyshapes_test.go          | 144 ++++++
 .../store/migrations/0031_drop_legacy_shapes.sql   |  44 ++
 stats/internal/store/pricing.go                    |  37 +-
 stats/internal/store/pricing_seed.go               |  21 +-
 stats/internal/store/pricing_seed_test.go          |  15 +-
 stats/internal/store/pricing_test.go               |  45 +-
 stats/internal/store/records.go                    |   2 +-
 stats/internal/store/records_test.go               |   4 +-
 stats/internal/store/stageruns.go                  |   2 +-
 stats/internal/store/stageruns_test.go             |  35 +-
 112 files changed, 870 insertions(+), 4245 deletions(-)

commit d210225b00d32d0a5af25938c971280a4e9c060d
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 04:09:15 2026 +0300

    chore(spectre): plan

 .../narrative.md                                      | 19 +++++++++++++++++++
 1 file changed, 19 insertions(+)

commit 0e478c6b673ede1ba556a9714e61d2b6b06a97fb
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 04:10:34 2026 +0300

    chore(spectre): archive kan-843-agents-remove-cursor-codex-harnesses-myflow-era

 .../design.md                                      |   0
 .../ledger.md                                      | 162 +++++++++++++++++++++
 .../narrative.md                                   |   0
 .../panel.md                                       |  39 +++++
 .../proposal.md                                    |   0
 .../tasks.md                                       |   0
 6 files changed, 201 insertions(+)

## Session narrative

Run 2 ran as the same-invocation continuation of the merge-and-push route. Run 1 found `origin/main`
31 commits ahead; the rebase stopped once on `KNOWN-BUGS.md`, where both sides appended deferred
Minors at the same spot, and was resolved by keeping both. The full lint and test lists then passed
except `check-task-records.sh`, which fails identically on `main` because integrate collapses the
per-task commits whose subjects the guard looks for. An earlier integrate run on this change had
already reshaped the branch, so this run's reshape re-collapsed a single implementation commit and
re-committed it with a module scope (`refactor(setup)`) in place of the earlier scopeless subject.
Run 2 archived the change, preserved the ledger and panel record, removed the worktree, local and
remote branches (the workspace database did not exist), verified cleanup `COMPLETE`, wrote
`FINISHED` and moved KAN-843 to Done.
