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
