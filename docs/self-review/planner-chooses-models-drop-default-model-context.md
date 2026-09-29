# Self-review context bundle for planner-chooses-models-drop-default-model

found: 6 of 7 sources; skipped: 1 of 7 sources
skipped: change summary (absent)

## .superpowers/sdd/ledgers/planner-chooses-models-drop-default-model.md

# SDD ledger — planner-chooses-models-drop-default-model

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — implementer

- Task: 1
- Role: implementer
- Key: task-1-implementer
- Model: opus effort=default
- Commit: dee1b048
- Outcome: completed
- Started: 2026-09-29T10:04:32Z
- Tokens: not measured

## Dispatch 2 — implementer

- Task: 2
- Role: implementer
- Key: task-2-implementer
- Model: opus effort=default
- Commit: 4b671c4145e7b71eafe87c20e72cc7e8944e2dbc
- Outcome: completed
- Started: 2026-09-29T10:07:10Z
- Tokens: not measured

## Dispatch 3 — implementer

- Task: 3
- Role: implementer
- Key: task-3-implementer
- Model: opus effort=default
- Commit: fd73e76a
- Outcome: completed
- Started: 2026-09-29T10:09:37Z
- Tokens: not measured

## Dispatch 4 — implementer

- Task: 4
- Role: implementer
- Key: task-4-implementer
- Model: opus effort=default
- Commit: efc11d37dd83a4bc7d5cb013f49945e20631fc38
- Outcome: completed
- Started: 2026-09-29T10:12:13Z
- Tokens: not measured

## Dispatch 5 — implementer

- Task: 5
- Role: implementer
- Key: task-5-implementer
- Model: opus effort=default
- Commit: 739bd92fd3f6379d81ef2247a85990af7ccd8a2e
- Outcome: completed
- Started: 2026-09-29T10:14:55Z
- Tokens: not measured

## Dispatch 6 — implementer

- Task: 6
- Role: implementer
- Key: task-6-implementer
- Model: opus effort=default
- Commit: 1599daa9c687d0c89e3925b06469e521c608bebd
- Outcome: completed
- Started: 2026-09-29T10:15:55Z
- Tokens: not measured

## Dispatch 7 — implementer

- Task: 8
- Role: implementer
- Key: task-8-implementer
- Model: opus effort=default
- Commit: 7040a386
- Outcome: completed
- Started: 2026-09-29T10:20:31Z
- Tokens: not measured

## Dispatch 8 — reviewer

- Task: no task
- Role: reviewer
- Key: task-1+2+3+4+5+6-reviewer
- Model: opus effort=high
- Commit: no commit
- Outcome: fix
- Started: 2026-09-29T10:23:54Z
- Tokens: input 142, output 6550, cache read 8510156, cache creation 202457

## Dispatch 9 — panel-fix

- Task: no task
- Role: panel-fix
- Key: task-1+4+5-implementer-fix-1
- Model: opus effort=default
- Commit: 4c4065da559a74cebee26bbddc4d42221473789d
- Outcome: completed
- Started: 2026-09-29T13:13:28Z
- Tokens: not measured

## Dispatch 10 — reviewer

- Task: no task
- Role: reviewer
- Key: task-1+4+5-reviewer-fix-1
- Model: sonnet effort=medium
- Commit: no commit
- Outcome: clean
- Started: 2026-09-29T13:14:43Z
- Tokens: input 12, output 41, cache read 140326, cache creation 33687

## Dispatch 11 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: opus effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-29T13:16:29Z
- Tokens: input 112, output 881, cache read 8445191, cache creation 218314

## Dispatch 12 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: opus effort=default
- Commit: cfdb871e99caff6dc2f796e880c305258d77fa88
- Outcome: completed
- Started: 2026-09-29T13:26:42Z
- Tokens: not measured

## Dispatch 13 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: opus effort=low
- Commit: no commit
- Diff base: 16510047
- Outcome: completed
- Started: 2026-09-29T13:29:58Z
- Tokens: not measured

## Dispatch 14 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: opus effort=low
- Commit: no commit
- Diff base: 16510047
- Outcome: completed
- Started: 2026-09-29T15:25:32Z
- Tokens: not measured

## Dispatch 15 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-2-principles
- Model: opus effort=low
- Commit: no commit
- Diff base: 3ee559fc~1
- Outcome: completed
- Started: 2026-09-29T15:27:29Z
- Tokens: input 14, output 533, cache read 143044, cache creation 23196

## Dispatch 16 — verifier

- Task: no task
- Role: verifier
- Slot: verify
- Key: verify
- Model: opus effort=high
- Commit: no commit
- Diff base: 4dafab4a
- Outcome: completed
- Started: 2026-09-29T15:28:28Z
- Tokens: not measured
## .superpowers/sdd/reviews/planner-chooses-models-drop-default-model-panel.md

# Review panel — planner-chooses-models-drop-default-model

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | skills/flow/implement.md:925 | The gated per-task reviewer — a first-pass review — takes its model from the group's or implementer pair, which Model and effort lets be sonnet; no bound pins it to opus, contradicting the operator's first-pass-review-on-opus rule. |   |
| F2 | primary | Minor | stats/internal/api/records.go:252 | encoding/json matches keys case-insensitively, last key wins, so {"implementer":{"model":"haiku"},"Implementer":"skipped — inline"} passes the refusal while jsonb readers still see haiku. |   |
| F3 | primary | Minor | skills/flow/brainstorm.md:244 | The record-time refusal does not stop the run: the Decide call site states no rule for a refused flow record decision, and dispatches read the on-disk decision.json. |   |
| F4 | primary | Minor | skills/flow-contracts/model-policy-rationale.md:24 | The rationale still says implementers run on Opus and sit at the ceiling from round 1, while the new bounds allow sonnet for implementers. |   |
| F5 | principles | Minor | stats/internal/api/records.go:311 | The refusal message hardcodes "opus, sonnet", a second copy of the set store.ValidModels owns, breaking valid-models-opus-sonnet's one-place bound. |   |
| F6 | principles | Minor | skills/flow-contracts/model-policy.md:38 | The no-recorded-pair-runs-on-opus rule is stated in two files whose lists already differ, readers cite different sources, and brainstorm-planner's canonical-for-which-model claim is too broad. |   |

findings-total: 6
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed

reproducers-total: 6
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-principles-2.sh

## Pass log

### Round 0

- roster: compact — 48
- diff size: 3226 changed lines, under cap; docs-only: exit 1 (first non-doc path scripts/check-model-keys.sh) — resolved roster primary+principles dispatched; no addition this round — the resolved list ran alone; standards passed: CLAUDE.md, AGENTS.md

### Round 1

- fix round 1: F1-F6 fixed inline by the parent (FIX_BASE 16510047, fix-round-1.diff); reproducers flipped demonstrated->not-demonstrated on pinned shas; records.go behaviours mutation-proved (F5 test strengthened to an exact suffix after its first mutant survived)
- base CLEAR; re-run: primary alone (raised Important F1) on the rerun pair opus/low, targeted at F1-F4; principles not re-run (Minors only); no addition this round — the resolved list ran alone
- base MOVED (2 commits, no overlap): operator chose Continue, no rebase — the auto-rebase would need a force-push Branch backup reserves for integrate; integrate syncs the branch
fix-mutation: stats/internal/api/records.go — exact-key map decode reverted to the pre-fix case-insensitive struct decode — TestRecordDecisionRejectsCaseVariantKeys
fix-mutation: stats/internal/api/records.go — refusal list no longer derived from ValidModels (extra haiku entry) — TestRecordDecisionRefusalNamesValidModels
fix-mutation: skills/ — none — prose-only hunks (implement, brainstorm, brainstorm-planner, model-policy, rationale): no executable behaviour
fix-mutations-total: 3
## spectre/changes/archive/planner-chooses-models-drop-default-model/tasks.md

# planner-chooses-models-drop-default-model

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Tasks appended:** 1

`design.md` is canonical for every decision; tasks cite them by ID.

**Baseline, measured before any edit:**

- `flow_settings` holds `default_model`, `self_review_model`, `reviewers`.
  <!-- measured: grep -n default_model stats/internal/store/settings.go @ 4dafab4a -->
- 62 `func Test` across the four settings test files; 75 across `stats/internal/guard/*_test.go`;
  47 in `stats/internal/api/records_test.go`.
  <!-- measured: cat <files> | grep -c '^func Test' @ 4dafab4a -->
- Nothing in Go parses a decision body: `records.Decision.Decision` is a `json.RawMessage` and
  `ApplyDecisionRecord` checks only emptiness.
  <!-- measured: read stats/internal/api/records.go:231 @ 4dafab4a -->

**Every task's verify step** is the task's own lint lines from `.flow/project.md` `## lint` that
its `**Files:**` need (Go: `cd stats && gofmt -l . && go vet ./...`; Markdown:
`scripts/check-references.sh`, `scripts/check-markdown-integrity.py`,
`scripts/check-guard-symlinks.sh`, `scripts/check-installed-citations.sh`) plus the targeted tests
named.

## Review Focus

- An old installed `flow` binary running `flow settings set -model …` against the new daemon must
  fail loudly (400 naming the unknown field), never write a partial row — Task 1 test
  `TestSettingsAPI_Put_RejectsDefaultModelField`.
- A decision whose `implementer`/`fixer` is the string `skipped — inline` or whose `panel` is the
  string `default` must still record — Task 2 test `TestRecordDecisionAcceptsRecordedStrings`.
- A decision naming `haiku` or `fable` in any one nested pair (e.g. only `groups[1].model`) is
  refused and the error names that path — Task 2 test `TestRecordDecisionRejectsOffPolicyModel`.
- The migration on a store whose row was written before it (both model columns populated) keeps
  `reviewers` intact — Task 1 test `TestSettingsStore_MigrationDropsModelColumns`.
- No skill text still names `DEFAULT_MODEL`, `MODEL_SOURCE`, `SELF_REVIEW_MODEL`, `selfReviewModel`,
  `defaultModel`, `flow settings models`, `## model`, `## self review model` or `## self review` as a
  live mechanism — Task 7's grep sweep.

**Live verification:** Task 7 runs the worktree's own `flowd` (its isolated `flow_<id>` database)
against the migrated schema and records before/after.

---

- [x] 1. Settings store, API, client and CLI carry reviewers only

Drop both model settings from the whole settings stack in one green commit (they compile
together).

  - [x] **Step 1: Write the failing tests.**
    - `stats/internal/store/settings_test.go`: delete `TestSettingsStore_RejectsUnknownModel` and
      `TestSettingsStore_RejectsUnknownSelfReviewModel`; add
      `TestSettingsStore_MigrationDropsModelColumns` — against the test DB after migrations, query
      `information_schema.columns` for `flow_settings` and assert exactly `id`, `reviewers`
      remain; before migrating `0034` (use the existing migration-test helper the
      `MigrationDropsRetiredReviewers` test uses) insert a row with both model columns set and a
      reviewers list, and assert `GetSettings` returns that reviewers list unchanged after.
      Update the round-trip tests to `store.Settings{Reviewers: …}`.
    - `stats/internal/api/settings_test.go`: delete `TestSettingsAPI_Get_EchoesSelfReviewModel`;
      add `TestSettingsAPI_Put_RejectsDefaultModelField` — PUT `{"defaultModel":"opus",
      "reviewers":["primary"]}` → 400, body names `defaultModel`, store untouched. `Get` asserts
      the body is exactly `{"reviewers":[…]}`.
    - `stats/internal/client/client_test.go`: delete `TestSettingsRoundTripsSelfReviewModel`;
      rename `TestSettingsPutBodyCarriesNoPlanningModelKey` to
      `TestSettingsPutBodyCarriesNoModelKeys` and assert the PUT body has no `defaultModel`,
      `selfReviewModel` or `planningModel` key.
    - `stats/cmd/flow/settings_test.go`: delete `TestSettingsCmd_Set_WithSelfReviewModel` and
      `TestSettingsCmd_Models`; add `TestSettingsCmd_Set_RejectsModelFlag` (`set -model opus
      -reviewers primary` → exit 2, stderr names `-model`) and `TestSettingsCmd_ModelsIsUnknown`
      (`settings models` → exit 2, `unknown settings command "models"`). `Set_Valid`/`Get` use
      `-reviewers` alone.
  - [x] **Step 2: Run them; they fail** — `cd stats && go test ./internal/store ./internal/api
    ./internal/client ./cmd/flow -run 'Settings' -count=1`.
  - [x] **Step 3: Implement.**
    - New `stats/internal/store/migrations/0034_flow_settings_drop_model_columns.sql`:
      `ALTER TABLE flow_settings DROP COLUMN default_model, DROP COLUMN self_review_model;` with a
      header comment citing `remove-default-model` and `remove-self-review-model`.
    - `stats/internal/store/settings.go`: `ValidModels` = `{"opus": true, "sonnet": true}`, its
      comment rewritten (the planner bound `ApplyDecisionRecord` enforces — decision
      `valid-models-opus-sonnet`); delete `DefaultModel`, `ErrInvalidModel`; `Settings` is
      `{Reviewers []string}`; `ValidateSettings`, `PutSettings`, `GetSettings` read/write
      `reviewers` only.
    - `stats/internal/api/settings.go`: `settingsDTO` is `{Reviewers []string \`json:"reviewers"\`}`;
      drop the `ErrInvalidModel` branch.
    - `stats/internal/client/client.go`: `Settings` is `{Reviewers}`; `ErrSettingsRejected`'s
      comment says "an unknown reviewer value".
    - `stats/cmd/flow/settings.go`: usage `flow settings get …` / `flow settings set … -reviewers
      a,b,c`; delete `-model`, `-self-review-model`, `runSettingsModels` and the `models` case;
      `-reviewers` alone is required. `stats/cmd/flow/main.go`: drop `settings models` from help.
  - [x] **Step 4: Run the tests; they pass**, then `cd stats && gofmt -l . && go vet ./...`.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/store/migrations/0034_flow_settings_drop_model_columns.sql`,
`stats/internal/store/settings.go`, `stats/internal/store/settings_test.go`,
`stats/internal/api/settings.go`, `stats/internal/api/settings_test.go`,
`stats/internal/client/client.go`, `stats/internal/client/client_test.go`,
`stats/cmd/flow/settings.go`, `stats/cmd/flow/settings_test.go`, `stats/cmd/flow/main.go`
**Tests:** `TestSettingsStore_MigrationDropsModelColumns`,
`TestSettingsAPI_Put_RejectsDefaultModelField`, `TestSettingsPutBodyCarriesNoModelKeys`,
`TestSettingsCmd_Set_RejectsModelFlag`, `TestSettingsCmd_ModelsIsUnknown`
**Regression:** reverting brings back `default_model`/`self_review_model`: the migration test
finds the columns, the API accepts `defaultModel`, the PUT body carries model keys, `-model` parses
and `settings models` answers.
**Baseline:** before=62 after=60
<!-- measured: cat stats/internal/store/settings_test.go stats/internal/api/settings_test.go stats/internal/client/client_test.go stats/cmd/flow/settings_test.go | grep -c '^func Test' @ 4dafab4a -->
**Commit:** `feat(store): drop default_model and self_review_model from flow_settings`
**After:** none
**Build:** green

**Decision:** remove-default-model
**Decision:** remove-self-review-model
**Decision:** valid-models-opus-sonnet

Correction (2026-09-29): **Build:** green holds for the task's own tests, not the tree: shrinking
`ValidModels` here left `scripts/check-model-keys.sh` (on `.flow/project.md`'s `fable` value) and
`TestCheckModelKeys*` red at this commit and Task 2's, until Task 3's commit retired both. The
branch was already pushed, so the order stands and this records it.

- [x] 2. Refuse an off-policy model when a decision is recorded

  - [x] **Step 1: Write the failing tests** in `stats/internal/api/records_test.go`:
    - `TestRecordDecisionRejectsOffPolicyModel` — table over the five paths
      (`implementer.model`, `fixer.model`, `panel.dispatches[0].model`,
      `panel.rerun_dispatch.model`, `groups[1].model`), each set to `haiku` in an otherwise valid
      body (and one case `fable`); `ApplyDecisionRecord` returns `ErrInvalidRecord`, the message
      contains the path and value, and the fake store records nothing.
    - `TestRecordDecisionAcceptsRecordedStrings` — a micro body (`implementer`/`fixer`
      `"skipped — inline"`, `panel` `"default"`, `groups` `null`) and a full `sdd` body on
      `opus`/`sonnet` both record.
  - [x] **Step 2: Run them; they fail** — `cd stats && go test ./internal/api -run
    'RecordDecision' -count=1`.
  - [x] **Step 3: Implement** in `stats/internal/api/records.go`: a `checkDecisionModels(body
    json.RawMessage) error` that unmarshals into a minimal struct of `json.RawMessage` fields
    (`implementer`, `fixer`, `panel`, `groups`); a field that decodes as a JSON string or `null` is
    skipped; an object's `model` (and `panel.dispatches[].model`, `panel.rerun_dispatch.model`,
    `groups[].model`) must be in `store.ValidModels`, else `fmt.Errorf("%w: %s model %q is not
    one of opus, sonnet", ErrInvalidRecord, path, model)`. An absent `model` key in an object is
    refused the same way (`""`). Call it from `ApplyDecisionRecord` after the emptiness check. A
    body that is not a JSON object is refused as `ErrInvalidRecord`.
  - [x] **Step 4: Run the tests; they pass**; `gofmt -l`, `go vet`.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/api/records.go`, `stats/internal/api/records_test.go`
**Tests:** `TestRecordDecisionRejectsOffPolicyModel`, `TestRecordDecisionAcceptsRecordedStrings`
**Regression:** reverting lets a decision naming `haiku`/`fable` record: the reject test's cases
record a row.
**Baseline:** before=47 after=49
<!-- measured: grep -c '^func Test' stats/internal/api/records_test.go @ 4dafab4a -->
**Commit:** `feat(api): refuse a recorded decision whose model is not opus or sonnet`
**After:** Task 1
**Build:** green

**Decision:** refuse-off-policy-models

Correction (2026-09-29): as for Task 1, this commit's tree is still red on
`scripts/check-model-keys.sh` and `TestCheckModelKeys*`; Task 3's commit retires both.

Correction (2026-09-29): the panel's round-1 fix added `TestRecordDecisionRejectsCaseVariantKeys`
(the check reads exact keys — `encoding/json` struct decoding folds case) and
`TestRecordDecisionRefusalNamesValidModels` (the refusal lists `store.ValidModels`), on top of
this task's commit; `records_test.go` carries 51 tests after them.
<!-- measured: grep -c '^func Test' stats/internal/api/records_test.go @ 95c9967f -->

- [x] 3. Retire the model-key and model-resolution guards and the three project keys

Must land before Tasks 4 and 6: `check-model-resolution-shell.sh` extracts the blocks those tasks
rewrite.

  - [x] **Step 1:** `git rm scripts/check-model-keys.sh scripts/check-model-resolution-shell.sh
    scripts/test-check-model-resolution-shell.sh stats/internal/guard/modelkeys.go
    stats/internal/guard/check_model_keys_test.go`.
  - [x] **Step 2:** remove the `"check-model-keys.sh": {"lib"},` entry from
    `stats/internal/guard/check_guard_symlinks_test.go`; drop `check-model-keys.sh` from
    `scripts/lib/project-section.sh`'s header comment; in `.flow/project.md` delete the two
    `## lint` lines and the `## self review`, `## self review model` and `## model` sections; in
    `.flow/project-rationale.md` delete any passage about those keys or guards; delete
    `KNOWN-BUGS.md`'s three entries citing `modelkeys.go` and `check-model-resolution-shell.sh`;
    in `skills/flow-contracts/project-configuration-rationale.md` delete the
    `check-model-keys.sh` passage (around line 200) and the rationale for the three keys.
  - [x] **Step 3: Verify** — `cd stats && go vet ./... && go test ./internal/guard -count=1`,
    `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`.
  - [x] **Step 4: Commit.**

**Files:** `scripts/check-model-keys.sh`, `scripts/check-model-resolution-shell.sh`,
`scripts/test-check-model-resolution-shell.sh`, `stats/internal/guard/modelkeys.go`,
`stats/internal/guard/check_model_keys_test.go`,
`stats/internal/guard/check_guard_symlinks_test.go`, `scripts/lib/project-section.sh`,
`.flow/project.md`, `.flow/project-rationale.md`, `KNOWN-BUGS.md`,
`skills/flow-contracts/project-configuration-rationale.md`,
`stats/internal/guard/workspaceisolation.go`
**Tests:** `TestShimSiblingsDeclared` — its sibling map loses the `check-model-keys.sh` entry
**Regression:** reverting restores the map entry and the guard; `TestShimSiblingsDeclared` still passes with both, so the check is that it passes without them.
**Baseline:** before=75 after=73
<!-- measured: cat stats/internal/guard/*_test.go | grep -c '^func Test' @ 4dafab4a -->
**Commit:** `chore(scripts): retire the model-key and model-resolution guards`
**After:** Task 1
**Build:** green

**Decision:** remove-default-model
**Decision:** remove-self-review-model
**Decision:** self-review-deferred-only

Correction (2026-09-29): the plan declared the five guard files deletable as-is; `modelkeys.go`
also defined `mkSpace`, which `workspaceisolation.go` calls, so the helper moved there as
`wiSpace` and that file joined **Files:**. `KNOWN-BUGS.md` lost twelve entries, not three: every
entry whose subject this change deletes — the two guards, `check_model_keys_test.go`, the
`## self review model` wording at `project-configuration.md:91`, archive's self-review-key
snippet at `archive.md:191`, and `SKILL.md:74`'s `MODEL_SOURCE` paragraph (Tasks 4 and 6 remove
those subjects).

- [x] 4. The Decide step chooses every pair's model; DEFAULT_MODEL goes

  - [x] **Step 1:** `skills/flow/brainstorm-planner.md` —
    - step 2: records the **fixer** pair on every non-micro class; the implementer pair only when
      step 1 is `sdd` (else `skipped — inline`);
    - step 3: rerun pair's `model` chosen per **Model and effort**, `effort` `low`;
    - step 4: group pairs chosen per **Model and effort**;
    - **Model and effort**: `model` ∈ {`opus`, `sonnet`} with the bounds in `design.md` **Model
      choice**, stated once here as canonical; micro and any dispatch with no recorded pair run on
      the literal `opus`; `flow record decision` refuses any other model;
    - JSON shape: drop `resolved`; the `## Decision` block's `↳ fixer` row shows the pair on inline
      runs; the preamble's `models:` line becomes `reviewers <REVIEWERS>`.
  - [x] **Step 2:** `skills/flow/SKILL.md` **Model resolution** — the block keeps
    `SETTINGS_JSON`, `REVIEWERS`, `VERIFY_MODEL=opus`; delete `PROJECT_MODEL`/`DEFAULT_MODEL`/
    `MODEL_SOURCE`, the `SELF_REVIEW_MODEL` paragraph, and "DEFAULT_MODEL is the model for all four
    roles"; the session-override paragraph overrides "the decision's pair(s)" for this run.
    `skills/flow/SKILL-rationale.md`: delete what explains the removed resolution.
  - [x] **Step 3:** `skills/flow-contracts/model-policy.md` — every role runs on its decision pair;
    delete **Where `DEFAULT_MODEL` comes from**; micro/unrecorded → `opus`; keep `VERIFY_MODEL`,
    zcode mapping; drop `SELF_REVIEW_MODEL` from the mapping list. `model-policy-rationale.md`:
    same deletions.
  - [x] **Step 4:** readers — `skills/flow/brainstorm.md` (line ~61 "models are resolved per run
    from the settings store" → "chosen per dispatch by the Decide step"; line ~207 drop
    `DEFAULT_MODEL`), `skills/flow-plan/SKILL.md` (~186), `skills/flow-contracts/pipeline.md`
    (~280: the summary names the decision's models, not a resolved default), `README.md` (~52, ~130,
    ~250: `/flow-settings` manages reviewer slots; the Decide step picks models),
    `commands-claude/flow.md` (~13: slots dispatched on the decision's models).
  - [x] **Step 5: Verify** — `grep -n 'DEFAULT_MODEL\|MODEL_SOURCE\|SELF_REVIEW_MODEL' <files>`
    prints nothing; `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`,
    `scripts/check-guard-symlinks.sh`, `scripts/check-installed-citations.sh`,
    `scripts/check-normative-inventory.sh` diffed against a capture taken before the first edit
    (differences only where a sentence was about the removed mechanism).
  - [x] **Step 6: Commit.**

**Files:** `skills/flow/brainstorm-planner.md`, `skills/flow/SKILL.md`,
`skills/flow/SKILL-rationale.md`, `skills/flow-contracts/model-policy.md`,
`skills/flow-contracts/model-policy-rationale.md`, `skills/flow/brainstorm.md`,
`skills/flow-plan/SKILL.md`, `skills/flow-contracts/pipeline.md`, `README.md`,
`commands-claude/flow.md`
**Tests:** none — prose contract; Task 7's sweep is the check
**Regression:** none — no executable behaviour; reverting restores the one-model rule.
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**Commit:** `feat(flow): the Decide step chooses every dispatch's model`
**After:** Task 3
**Build:** green

**Decision:** planner-chooses-models
**Decision:** fixer-pair-sdd-only
**Decision:** remove-default-model

Correction (2026-09-29): the plan had step 2 record the fixer pair on every non-micro class
(`inline-fixer-planner-chosen`); the gated review found inline runs have no panel-fix dispatch,
so the operator superseded it with `fixer-pair-sdd-only` and the fix round restored the
sdd-only fixer pair. The same round cut the README's restated model bounds to a citation of
**Model and effort**.

- [x] 5. Dispatch sites read the decision's pair

  - [x] **Step 1:** `skills/flow/implement.md` (~31, ~98, ~499, ~726, ~928): the handshake compares
    against the dispatch's recorded model (or this run's override); ~31 resolves `REVIEWERS` only;
    the gate-fired bundle at ~928 runs on the decision's implementer pair, `opus`/`default` when
    none is recorded.
  - [x] **Step 2:** `skills/flow/review-panel.md` (~159, ~306, ~412, ~1236, ~1330): a `default`
    panel and the no-decision dispatch run on `opus`; panel-fix runs on the decision's fixer pair
    (every non-micro class), `opus` on micro; `-model` in records is the dispatch's model.
  - [x] **Step 3:** the three reviewer prompts' `model:` comment
    (`primary-`, `principles-`, `failure-modes-reviewer-prompt.md`): "`opus` on a `default` panel,
    the decision's dispatch model otherwise". `skills/flow/visual-verify.md` (~16, ~73, ~142,
    ~145): drop the `DEFAULT_MODEL` contrasts; the tooling analyst runs on `opus`.
  - [x] **Step 4: Verify** — the grep of Task 4 step 5 over these files prints nothing; the
    Markdown lint lines.
  - [x] **Step 5: Commit.**

**Files:** `skills/flow/implement.md`, `skills/flow/review-panel.md`,
`skills/flow/primary-reviewer-prompt.md`, `skills/flow/principles-reviewer-prompt.md`,
`skills/flow/failure-modes-reviewer-prompt.md`, `skills/flow/visual-verify.md`
**Tests:** none — prose contract; Task 7's sweep is the check
**Regression:** none — no executable behaviour.
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**Commit:** `feat(flow): dispatch every role on its decision pair`
**After:** Task 4
**Build:** green

**Decision:** planner-chooses-models
**Decision:** fixer-pair-sdd-only

Correction (2026-09-29): the panel-fix sentence the plan implied ("the fixer pair on every
non-micro class, `opus` on micro") was superseded with `fixer-pair-sdd-only`; the fix round made
the panel-fix subagent run on the `sdd` decision's fixer pair alone.

- [x] 6. Self-review is deferred, always; /flow-settings manages reviewers only

  - [x] **Step 1:** `skills/flow/archive.md` step 9 — keep the bundle fetch, `## Session
    narrative`, the context-file write and `land-self-review-report.sh … context bundle`; delete
    the `SELF_REVIEW_MODEL` block, the `## self review` key read, the skip prompt, the inline
    reasoning pass, the filing/rating ask and the report write; the handoff `Self-review` line is
    `deferred — docs/self-review/<name>-context.md` only (and in `## Finished`).
    `skills/flow-contracts/finish-contract-run2.md` step 9 to match (canonical: the deferred bundle;
    the angles and rating now live in `/flow-self-review` alone).
  - [x] **Step 2:** `skills/flow-fast/SKILL.md` — the self-review section always writes and lands
    the bundle (drop "`defer` or nothing", the key read and the mark-through branch); line ~59 and
    ~188-190 drop `DEFAULT_MODEL` and the `## self review` key; its model text points at the Decide
    step. `commands-claude/flow-fast.md` to match.
  - [x] **Step 3:** `skills/flow-self-review/SKILL.md` and `commands-claude/flow-self-review.md` —
    drop "a `/flow` or `/flow-fast` run deferred" framing that implies alternatives and the
    same-run-pass comparison; the pass runs on this session's model.
  - [x] **Step 4:** `skills/flow-contracts/project-configuration.md` — delete the `## self review`,
    `## self review model`, `## model` rows and their names from the match-and-drop list (~94-95).
    `skills/flow-contracts/git-boundaries.md`, `skills/flow-contracts/handoff-blocks-rationale.md`:
    drop self-review report / run-skip references that no longer happen in `/flow`.
  - [x] **Step 5:** `skills/flow-settings/SKILL.md` and `commands-claude/flow-settings.md` — the
    record is `reviewers` alone; the show block, the ask and `flow settings set -reviewers "<list>"`. The skill-index rows in
    `CLAUDE.md`, `AGENTS.md` and `skills/README.md` read "reviewer slots" only.
  - [x] **Step 6: Verify** — `grep -n 'self review model\|selfReviewModel\|SELF_REVIEW_MODEL\|##
    self review\|defaultModel\|settings models\|skipped — project default' <files>` prints nothing;
    `scripts/check-self-review-report.sh`, the Markdown lint lines.
  - [x] **Step 7: Commit.**

**Files:** `skills/flow/archive.md`, `skills/flow-contracts/finish-contract-run2.md`,
`skills/flow-fast/SKILL.md`, `commands-claude/flow-fast.md`, `skills/flow-self-review/SKILL.md`,
`commands-claude/flow-self-review.md`, `skills/flow-contracts/project-configuration.md`,
`skills/flow-contracts/git-boundaries.md`, `skills/flow-contracts/handoff-blocks-rationale.md`,
`skills/flow-settings/SKILL.md`, `commands-claude/flow-settings.md`, `CLAUDE.md`, `AGENTS.md`,
`skills/README.md`, `skills/flow-contracts/jira-integration.md`,
`scripts/check-self-review-report.sh`
**Allowed-collateral:** `scripts/test-check-self-review-report.sh`
**Tests:** none — prose contract plus comment-only harness edits; Task 7's sweep and the harness run are the check
**Regression:** none — no executable behaviour.
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**Commit:** `feat(flow): self-review is always deferred; settings hold reviewers only`
**After:** Task 3
**Build:** green

**Decision:** self-review-deferred-only
**Decision:** remove-self-review-model
**Decision:** remove-default-model

Correction (2026-09-29): moving the angles out of run 2 step 9 moved the five-angle table, the
filing bounds, the explain-first rule and the prompt shape into `skills/flow-self-review/SKILL.md`,
now canonical for them. `check-self-review-report.sh` parses that table at run time, so its default
source path moved with it (its harness passes, and a broken table makes it refuse), and
`jira-integration.md`'s citation of the table followed; the guard and the citation joined the
task's declared files, the harness's comment-only edits its allowed collateral.

- [x] 7. Live verification and the whole-tree sweep

Reproduce before touching anything; do not guess.

  - [x] **Step 1: Before** (on the worktree's isolated store, `FLOWD_DSN`/`FLOW_ADDR` from
    `.flow/project.md` `## workspace isolation`): with the pre-change binary, `flow settings get`
    prints `defaultModel`, `selfReviewModel`, `reviewers`; record those figures.
  - [x] **Step 2: After** — rebuild `flowd` and `flow` from the worktree, restart the worktree's
    own `flowd` (never the dev workspace's on 4173): `flow settings get` prints `{"reviewers":[…]}`
    with the same list; `flow settings set -model opus -reviewers primary` exits 2;
    `flow record decision` with a body whose `groups[0].model` is `haiku` exits non-zero naming
    that path; the same body on `sonnet` records. **Not working looks like:** a `defaultModel`
    key still printed, a `haiku` decision recorded, or reviewers lost by the migration.
  - [x] **Step 3: Sweep** — `grep -rn 'DEFAULT_MODEL\|MODEL_SOURCE\|SELF_REVIEW_MODEL\|
    selfReviewModel\|defaultModel\|settings models\|## self review\|## model\b\|check-model-keys\|
    check-model-resolution-shell\|default model\|self-review model' -i --exclude-dir=spectre --exclude-dir=docs --exclude-dir=node_modules
    --exclude-dir=dist .` prints nothing but the migrations' historical comments; any hit becomes an
    appended task.
  - [x] **Step 4:** the full `## lint` and `## test` lists pass.

**Files:** none
**Tests:** none — measurement task; the figures it records are the check
**Regression:** none — no commit
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 2, 5, 6
**Build:** green

- [x] 8. Sweep hit: the self-review harness comment names the retired key

Task 7's sweep found `scripts/test-check-self-review-report.sh:820` still saying the bundle is
written "on `## self review: defer`"; it is written on every run now.

  - [x] **Step 1:** reword the comment to "a bundle step 9 writes on every run".
  - [x] **Step 2: Verify** — `scripts/test-check-self-review-report.sh` passes; Task 7's sweep
    prints no hit outside migrations, the retirement tests and `## Model …` headings.
  - [x] **Step 3: Commit.**

**Files:** `scripts/test-check-self-review-report.sh`
**Tests:** `Case 25` — the case whose comment this rewords
**Regression:** none — no executable behaviour.
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**Commit:** `docs(scripts): the self-review bundle is written on every run`
**After:** Task 6, 7
**Build:** green
## spectre/changes/archive/planner-chooses-models-drop-default-model/design.md

# Design — planner-chooses-models-drop-default-model

## Context

- `ea661bc6` pinned every Decide pair to `DEFAULT_MODEL`; the operator wants the planner to choose
  models, `default_model` and `self_review_model` removed, and self-review deferred-only.

## Model choice

- The Decide step (`skills/flow/brainstorm-planner.md`) assigns `{model, effort, reason}` to every
  pair: implementer, fixer, each `panel.dispatches[]` entry, `panel.rerun_dispatch`, each `groups[]`
  entry. `model` ∈ {`opus`, `sonnet`}.
- Bounds:
  - implementer, fixer, groups — `opus` by default; `sonnet` only for simple, predictable work
    near-certain to succeed first try, the `reason` saying so;
  - a pair carrying a hard seam (concurrency, platform interop, data model, performance-sensitive
    code) — `opus`, always;
  - first-pass panel dispatches — `opus`, always;
  - rerun pair — `opus` or `sonnet`, effort `low` fixed; no must-differ-from-pass-1 constraint (the
    defect `ea661bc6` fixed stays fixed).
- Step 2 records the implementer and **fixer** pairs on `sdd` only; both stay `skipped — inline`
  on an inline decision, whose panel fixes the parent applies on its own model
  (`fixer-pair-sdd-only`).
- **Micro** records no pairs; its panel and every dispatch with no recorded pair
  (review-panel.md's no-decision dispatch) run on the literal `opus`.
- A plain-language session instruction replaces the affected pairs for this run, recorded in
  `overrides`. The model handshake compares against the dispatch's own recorded (or overridden)
  model.
- Unchanged: `VERIFY_MODEL=opus`, the zcode harness mapping, effort rules.
- Removed from `decision.json`: `resolved`. Removed from the Decide preamble and run summary:
  `DEFAULT_MODEL`/`MODEL_SOURCE`; the preamble's `models:` line becomes `reviewers <REVIEWERS>`.

## Record-time refusal

- `ApplyDecisionRecord` (`stats/internal/api/records.go`) parses the decision body's model fields:
  `implementer.model`, `fixer.model`, `panel.dispatches[].model`, `panel.rerun_dispatch.model`,
  `groups[].model`. A string-valued `implementer`/`fixer`/`panel` (`skipped — inline`, `default`)
  and a `null` `groups` carry no model and pass.
- A model outside `store.ValidModels` → `ErrInvalidRecord` (HTTP 400) naming the JSON path and the
  value. Writes only; historical rows are never re-checked.
- Server-side, so every client is covered, not only the CLI.

## Settings store

- Migration `0034` drops `flow_settings.default_model` and `flow_settings.self_review_model`.
  `flow_settings` keeps `reviewers`.
- `store.Settings` = `{Reviewers}`; `DefaultModel`, `ErrInvalidModel` on settings writes,
  `SelfReviewModel` go. `ValidModels` = {`opus`, `sonnet`}.
- API/client DTO drops `defaultModel`, `selfReviewModel`. CLI drops `-model`,
  `-self-review-model` and `flow settings models`.
- Compatibility: an old `flow` binary sending `defaultModel` is refused (400, unknown field — the
  decoder is strict); an old reader sees the field absent. `./setup.sh global` reinstalls.

## Self-review

- Deferred is the only path. `/flow` archive step 9 (`skills/flow/archive.md`,
  `skills/flow-contracts/finish-contract-run2.md`) and `/flow-fast` always fetch the bundle, add
  `## Session narrative`, write `docs/self-review/<name>-context.md` and land it.
- Removed: the run/skip/defer prompt, the inline reasoning pass, archive's filing/rating ask and
  report write, `SELF_REVIEW_MODEL` resolution, the `## self review` and `## self review model`
  keys.
- The handoff `Self-review` line always reads `deferred — docs/self-review/<name>-context.md`.
- `/flow-self-review` unchanged in behaviour; its wording that compares against a "same-run pass"
  is updated.

## Guards removed

- `scripts/check-model-keys.sh` + `stats/internal/guard/modelkeys.go` + tests: no model key left.
- `scripts/check-model-resolution-shell.sh` + `scripts/test-check-model-resolution-shell.sh`: both
  blocks it extracts lose their model resolution; `REVIEWERS` is a one-line jq read.
- Both leave `.flow/project.md`'s `## lint` list; their `KNOWN-BUGS.md` entries are removed.

## Decisions

### Planner chooses every dispatch's model

**ID:** planner-chooses-models
**Status:** active
**Chosen:** the Decide step picks `model` per pair within {opus, sonnet} — matches the operator's
global model rule, which varies model by role and task difficulty.
**Considered:** keep `DEFAULT_MODEL` for every pair (status quo since `ea661bc6`) — cannot express
opus-by-default with sonnet for simple work; `DEFAULT_MODEL` as fallback only — rejected by the
operator in favour of removing the setting entirely; `DEFAULT_MODEL` as a ceiling/floor — adds a
knob nobody asked for.

### Remove default_model everywhere

**ID:** remove-default-model
**Status:** active
**Chosen:** drop the store column, API/CLI fields and the `## model` project key; literal `opus`
where no pair is recorded — nothing reads the setting once the planner chooses.
**Considered:** remove from skills only, keep the column — leaves a dead setting `/flow-settings`
still offers.

### Inline fixer is planner-chosen

**ID:** inline-fixer-planner-chosen
**Status:** superseded by fixer-pair-sdd-only
**Chosen:** step 2 records a fixer pair on every non-micro class — the inline panel-fix dispatch
otherwise has no model source.
**Considered:** literal `opus` for the inline fixer — loses the sonnet-for-simple-fixes option.
**Superseded because:** the gated per-task review found no inline panel-fix dispatch exists — an
inline run's parent applies panel fixes itself on its own model (`implement.md` **Inline — the
parent implements**), so an inline fixer pair governs nothing.

### Fixer pair on sdd only

**ID:** fixer-pair-sdd-only
**Status:** active
**Chosen:** step 2 records the fixer pair only when execution is `sdd`, the one mode with a
panel-fix subagent; inline runs record `skipped — inline` — operator's answer at the task review
(2026-09-29).
**Considered:** inline runs dispatch a panel-fix subagent on the fixer pair — adds a dispatch and
its cost to every inline fix round.

### Refuse off-policy models at record time

**ID:** refuse-off-policy-models
**Status:** active
**Chosen:** `ApplyDecisionRecord` refuses any model outside `ValidModels` = {opus, sonnet} —
enforced where every client writes, not only stated in skill text.
**Considered:** skill text only — an unenforced bound drifts; CLI-side check — misses non-CLI
clients.

### Self-review is deferred only

**ID:** self-review-deferred-only
**Status:** active
**Chosen:** `/flow` and `/flow-fast` always save the bundle; `/flow-self-review` is the only pass,
run on the model the operator picks for that session.
**Considered:** deferred as default with run/skip kept — operator: deferred is "the only option".

### Remove self_review_model

**ID:** remove-self-review-model
**Status:** active
**Chosen:** drop the column, fields and `## self review model` key — deprecated; the deferred pass
runs on the session's hand-picked model.
**Considered:** none raised.

### ValidModels becomes the planner bound

**ID:** valid-models-opus-sonnet
**Status:** active
**Chosen:** shrink `ValidModels` to {opus, sonnet}; with both model settings gone its only reader
is the decision refusal, so the bound lives in one place.
**Considered:** a separate planner set beside a wider `ValidModels` — two sets, one with no
reader.

## Open questions
## spectre/changes/archive/planner-chooses-models-drop-default-model/narrative.md

# planner-chooses-models-drop-default-model — session narrative

## 2026-09-29 — creating run

- **Resumed at `STARTED`, inline execution.** This run's own model resolution was the last to read
  a `defaultModel`: the store answered, so the recorded decision carries `opus`/`store` — the
  mechanism this change removes.
- **Build-red intermediate tasks.** Tasks 1 and 2 removed symbols (`DefaultModel`,
  `ErrInvalidModel`) whose last callers only Task 3 deleted, so the tree was red between them; the
  plan's Correction paragraphs record it. `mkSpace` lived in the deleted `modelkeys.go` and moved to
  `workspaceisolation.go` as `wiSpace`.
- **Guard hits along the way.** `check-references` (rewraps), `check-installed-citations` (a
  redundant line in `archive.md`), `check-self-review-report` (the angle table moved into
  `skills/flow-self-review/SKILL.md`; the guard was repointed), `check-plan-shape` (a test path under
  `Tests: none`, a Correction line opening with `**Files:**`, a missing `After:`), and
  `check-plan-provenance` (the 51-test count needed its `measured:` comment).
- **Operator decision — fixer pair.** Inline runs have no panel-fix dispatch, so the recorded
  `inline-fixer-planner-chosen` decision was false; the operator chose to supersede it with
  `fixer-pair-sdd-only`.
- **Operator decision — base moved.** `origin/main` moved 2 commits with no overlap at panel round
  1; the operator chose to continue without a rebase. Integrate's sync rebases.
- **Panel.** Round 0 raised F1 (Important: the gated reviewer could run on sonnet) and F2–F6
  (Minor). All fixed. F6's first fix left two citations at the forwarding pointer; the principles
  re-run caught it, fixed in round 2. F5's mutation first survived a `Contains` assertion; the test
  was pinned with `HasSuffix`.
- **Environment.** zsh does not word-split unquoted variables, which broke the first pathspec
  commit; `run-reproducer.sh` takes its flags after the positionals and pins the reproducer file's
  sha256, and a reproducer needs its execute bit. A `flow stage begin` issued from the skills
  directory left `flow.verify`'s run unmatched by its `end`; the next stage's `begin` superseded it.
- **Gated-review fixer key.** `check-panel-fix-single-dispatch.sh` flagged the gated per-task fix
  row key `task-1+4+5-implementer-fix-1` as out of the panel-fix shape; it is not a panel-fix
  dispatch, and the prompt was auto-resolved on Continue.

## 2026-09-29 — integrate run

- Preflight: `STAGED-CLEAN`, `DRIFT-CLEAN`, `RUN1` against `origin/main`.
- Unfinished-work gate: `CLEAR`; visual verify `VISUAL-VERIFY-OK` (no UI paths).
- Base: `check-base-moved.sh` `CLEAR` — no rebase needed.
- Route: merge and push, from the project's configured default, not asked.
- Environment: one `flow stage end` issued from `~/.claude/skills` failed to resolve the project key
  (not a git checkout); re-issued from the main checkout.
## git log --stat

commit fabb28740fe7de8d1508a74ff01fe14ec32fd6b5
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 29 22:08:45 2026 +0300

    feat(flow): the planner chooses every dispatch's model; drop the default model

 .flow/project-rationale.md                         |   9 -
 .flow/project.md                                   |  16 +-
 AGENTS.md                                          |   4 +-
 CLAUDE.md                                          |   4 +-
 KNOWN-BUGS.md                                      |  15 +-
 README.md                                          |   9 +-
 commands-claude/flow-fast.md                       |   4 +-
 commands-claude/flow-self-review.md                |  10 +-
 commands-claude/flow-settings.md                   |   5 +-
 commands-claude/flow.md                            |   4 +-
 scripts/check-model-keys.sh                        |  49 ---
 scripts/check-model-resolution-shell.sh            | 193 ----------
 scripts/check-self-review-report.sh                |  11 +-
 scripts/lib/project-section.sh                     |   4 +-
 scripts/test-check-model-resolution-shell.sh       |  95 -----
 scripts/test-check-self-review-report.sh           |   9 +-
 skills/README.md                                   |   6 +-
 skills/flow-contracts/finish-contract-run2.md      |  98 +----
 skills/flow-contracts/git-boundaries.md            |   2 +-
 skills/flow-contracts/handoff-blocks-rationale.md  |   4 +-
 skills/flow-contracts/jira-integration.md          |   4 +-
 skills/flow-contracts/model-policy-rationale.md    |  22 +-
 skills/flow-contracts/model-policy.md              |  62 ++-
 skills/flow-contracts/pipeline.md                  |   6 +-
 .../project-configuration-rationale.md             |   4 -
 skills/flow-contracts/project-configuration.md     |   7 +-
 skills/flow-fast/SKILL.md                          |  21 +-
 skills/flow-plan/SKILL.md                          |   2 +-
 skills/flow-self-review/SKILL.md                   |  48 ++-
 skills/flow-settings/SKILL.md                      |  56 +--
 skills/flow/SKILL-rationale.md                     |  11 +-
 skills/flow/SKILL.md                               |  62 +--
 skills/flow/archive.md                             |  89 +----
 skills/flow/brainstorm-planner.md                  |  43 ++-
 skills/flow/brainstorm.md                          |   9 +-
 skills/flow/failure-modes-reviewer-prompt.md       |   2 +-
 skills/flow/implement.md                           |  26 +-
 skills/flow/primary-reviewer-prompt.md             |   2 +-
 skills/flow/principles-reviewer-prompt.md          |   2 +-
 skills/flow/review-panel.md                        |  14 +-
 skills/flow/visual-verify.md                       |  11 +-
 stats/cmd/flow/main.go                             |   1 -
 stats/cmd/flow/settings.go                         |  49 +--
 stats/cmd/flow/settings_test.go                    |  94 ++---
 stats/internal/api/records.go                      |  91 +++++
 stats/internal/api/records_test.go                 | 103 +++++
 stats/internal/api/settings.go                     |  33 +-
 stats/internal/api/settings_test.go                | 132 +++----
 stats/internal/client/client.go                    |  14 +-
 stats/internal/client/client_test.go               |  87 +----
 stats/internal/guard/check_guard_symlinks_test.go  |   1 -
 stats/internal/guard/check_model_keys_test.go      | 428 ---------------------
 stats/internal/guard/modelkeys.go                  | 303 ---------------
 stats/internal/guard/workspaceisolation.go         |   9 +-
 .../0034_flow_settings_drop_model_columns.sql      |   9 +
 stats/internal/store/settings.go                   |  95 ++---
 stats/internal/store/settings_test.go              | 145 ++++---
 57 files changed, 687 insertions(+), 1961 deletions(-)

commit c99408dcb38cbbf46400f598d8c059558bededb5
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 29 22:08:45 2026 +0300

    chore(spectre): plan

 .../planner-chooses-models-drop-default-model/narrative.md       | 9 +++++++++
 1 file changed, 9 insertions(+)

commit 1f080ac1f7d2738109563c1b48972184bd105f6c
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 29 22:09:38 2026 +0300

    chore(spectre): archive planner-chooses-models-drop-default-model

 .../design.md                                      |   0
 .../ledger.md                                      | 189 +++++++++++++++++++++
 .../narrative.md                                   |   0
 .../panel.md                                       |  45 +++++
 .../proposal.md                                    |   0
 .../tasks.md                                       |   0
 .../verbatim-moves.txt                             |   0
 7 files changed, 234 insertions(+)

## Session narrative

Integrate and archive ran in one invocation on the project's configured merge-and-push route: preflight `RUN1`, unfinished-work `CLEAR`, base unmoved, reshape kept 8 planning commits, two-commit split, local `--no-ff` merge in the landing worktree, archive, cleanup verified `COMPLETE`. Three stage marks were mishandled by the session: one `flow stage end` issued from `~/.claude/skills` could not resolve the project key and was re-issued; `flow.commit-two`'s end and `flow.landing-routes`'s begin were never issued (the former superseded by the next begin, the latter's end reported no open run). Self-review took `defer` because this change itself retired the `## self review` key and the run/skip prompt — the installed (pre-merge) archive.md still prompts on an absent key, while the merged one always defers; `project-get.sh` warned that the main checkout's `.flow/project.md` diverged from HEAD, the expected ghost from the local merge moving `main` before step 11's refresh.
