# kan-374-flow-run-planning-stages-on-fable-opus-fallback

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

> **Relocation:** no

Six tasks. Three are Go under `stats/` (TDD): the `planningModel` settings field, the `planner`
dispatch role, and the `flow.resolve-change` removal. Three are skill/contract prose: the setting's
resolution and `/flow-settings` question, the planner dispatch in `/flow`, and the research
subagent in `/flow-research`. `design.md` is canonical for the decisions and the planner protocol;
`proposal.md` for why. Every Go command below runs from `stats/`; every guard from the repo root.

**Baseline, measured before any edit:**

- `internal/store/settings_test.go`: 5 tests.
  <!-- measured: grep -c '^func Test' stats/internal/store/settings_test.go @ branch main (6a64963) -->
- `internal/api/settings_test.go`: 4 tests.
  <!-- measured: grep -c '^func Test' stats/internal/api/settings_test.go @ branch main (6a64963) -->
- `cmd/flow/settings_test.go`: 4 tests.
  <!-- measured: grep -c '^func Test' stats/cmd/flow/settings_test.go @ branch main (6a64963) -->
- `internal/client/client_test.go`: 41 tests (whole file, shared across every client method).
  <!-- measured: grep -c '^func Test' stats/internal/client/client_test.go @ branch main (6a64963) -->
- `cmd/flow/record_test.go`: 39 tests.
  <!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ branch main (6a64963) -->
- `internal/stages/names_test.go`: 9 tests.
  <!-- measured: grep -c '^func Test' stats/internal/stages/names_test.go @ branch main (6a64963) -->
- Latest migration is `0016_flow_settings_self_review_model.sql`; next is `0017`.
  <!-- measured: ls stats/internal/store/migrations | tail -1 @ branch main (6a64963) -->
- Contract-budget headroom that this plan's prose edits will exceed, per
  `scripts/check-contract-budget.sh` rows: `skills/flow/brainstorm.md`, `skills/flow/SKILL.md`,
  `skills/flow-settings/SKILL.md`. The task touching each file raises its row in the same commit;
  a budget raise is a visible diff in the guard by that guard's own design.
  <!-- measured: wc -c on each file against its row in scripts/check-contract-budget.sh @ branch main (6a64963) -->

- [x] 1. `planningModel` end to end: migration, store, API, client, CLI

Mirror commit `9858544` (`selfReviewModel`) layer for layer. Differences from that field: empty
resolves to the literal `fable` (a new `DefaultPlanningModel` constant beside `DefaultModel`, so a
row written before this column existed already resolves without a caller special-case), and
`ValidModels` gains `"fable": true` — it holds only sonnet, opus and haiku today, so the store would
otherwise refuse the very default this change introduces.

**`internal/store/migrations/0017_flow_settings_planning_model.sql`** — `ALTER TABLE flow_settings
ADD COLUMN planning_model TEXT NOT NULL DEFAULT ''`, with a header comment giving the same
NOT-NULL-DEFAULT-empty reason `0016` gives, plus: empty means "the store's own default,
`DefaultPlanningModel`", not "inherit `default_model`" — the two fields' empty values mean
different things and the comment says so.

**`internal/store/settings.go`**:
- `ValidModels`: add `"fable": true`; extend the comment above it to say fable is the planning
  default and a legal value for every model field.
- `const DefaultPlanningModel = "fable"`, doc comment naming what it is the fallback for.
- `Settings`: add `PlanningModel string` with a doc comment: empty is valid and means
  `DefaultPlanningModel`.
- `ValidateSettings`: `s.PlanningModel != "" && !ValidModels[s.PlanningModel]` → `ErrInvalidModel`
  wrapped with the value, same sentinel as the other two model fields.
- `PutSettings`: add `planning_model` to the `INSERT … ON CONFLICT DO UPDATE` and its args.
- `GetSettings`: select `planning_model`; the no-rows branch leaves it empty. **Do not** resolve
  empty to `fable` in the store — `GetSettings` returns what was written, exactly as it does for
  `SelfReviewModel`; resolution is the skill's job (task 4), so the wire shape stays a record of
  intent.

**`internal/api/settings.go`**: `PlanningModel string `json:"planningModel"`` on `settingsDTO`,
carried through `toSettingsDTO` and `put`.

**`internal/client/client.go`**: same field and tag on `client.Settings`.

**`cmd/flow/settings.go`**: `-planning-model` flag (`fset.String("planning-model", "", "planning's
model, e.g. fable; empty resolves to the store's default")`), optional, passed into the
`client.Settings{…}` literal; `settingsUsage`'s `settings set` line shows it.

Tests, RED before GREEN, each a new function beside its `SelfReviewModel` sibling so the count
below is checkable:
- `internal/store/settings_test.go`: `TestSettingsStore_RejectsUnknownPlanningModel` (rejects
  `"architect"`, accepts `""` and `"fable"`); extend `TestSettingsStore_RoundTrip` to round-trip
  `PlanningModel: "fable"` and `TestSettingsStore_RoundTripUpdates` to round-trip the empty case.
  Store tests run against a live Postgres per that file's own pattern.
- `internal/api/settings_test.go`: `TestSettingsAPI_Get_EchoesPlanningModel`; extend
  `TestSettingsAPI_Put_RejectsInvalidValue` with a bad `planningModel` → 400.
- `cmd/flow/settings_test.go`: `TestSettingsCmd_Set_WithPlanningModel`; extend
  `TestSettingsCmd_Get`'s canned JSON and struct with `planningModel`.
- `internal/client/client_test.go`: `TestSettingsRoundTripsPlanningModel`.

```bash verified:the commands stats/README.md and .flow/project.md name for this module
cd stats && go test ./internal/store/... ./internal/api/... ./internal/client/... ./cmd/flow/... -race -count=1
cd stats && go vet ./... && gofmt -l .
```

**Files:** `stats/internal/store/migrations/0017_flow_settings_planning_model.sql`,
`stats/internal/store/settings.go`, `stats/internal/store/settings_test.go`,
`stats/internal/api/settings.go`, `stats/internal/api/settings_test.go`,
`stats/internal/client/client.go`, `stats/internal/client/client_test.go`,
`stats/cmd/flow/settings.go`, `stats/cmd/flow/settings_test.go`
**Tests:** `TestSettingsStore_RejectsUnknownPlanningModel`, `TestSettingsAPI_Get_EchoesPlanningModel`,
`TestSettingsCmd_Set_WithPlanningModel`, `TestSettingsRoundTripsPlanningModel`
**Regression:** reverting this commit removes `planningModel` from the wire and store shape, so
`/flow-settings` (task 4) has no field to write and `PLANNING_MODEL` (task 4) reads nothing; the
four named tests fail to compile against the reverted struct.
**Baseline:** before=5 after=6 store tests, before=4 after=5 api tests, before=4 after=5 cmd/flow
settings tests, before=41 after=42 client tests
<!-- predicted: grep -c '^func Test' per file after task 1 -->
**Commit:** `feat(store): add planningModel to the harness-wide settings record`
**Build:** green

- [x] 2. Accept `planner` as a dispatch role

**`cmd/flow/record.go`**: append `"planner"` to `recordRoles`; extend the comment above it — the
closed set now reads `implementer · reviewer · panel-fix · red-partner · planner`, planner being
the subagent `/flow` dispatches for sections B–D of `skills/flow/brainstorm.md` and for
`flow.document-fix` (design.md `planner-role`). Update the `-role is one of:` usage line.

Tests, RED first:
- `cmd/flow/record_test.go`: `TestRecordAcceptsPlannerRole` — `record dispatch begin -role planner`
  against the `genuineDaemon` stub exits 0 and contacts the store; extend
  `TestRecordRejectsUnknownRoleWithoutContactingStore`'s accepted-role loop with `"planner"`.

```bash verified:the commands stats/README.md and .flow/project.md name for this module
cd stats && go test ./cmd/flow/... -race -count=1 -run 'TestRecord'
cd stats && go vet ./... && gofmt -l .
```

**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordAcceptsPlannerRole`
**Regression:** reverting this commit makes the planner dispatch record (task 5) exit 2 as a
caller mistake before the store is contacted; `TestRecordAcceptsPlannerRole` fails on exit code.
**Baseline:** before=39 after=40 record tests
<!-- predicted: grep -c '^func Test' stats/cmd/flow/record_test.go after task 2 -->
**Commit:** `feat(record): accept the planner dispatch role`
**Build:** green

- [x] 3. Fold `flow.resolve-change` into `flow.kickoff`

Remove the key from every place it is written; historical rows in `stage_runs` stay, and only
writes validate against the table (`stages.Validate` is called from `stage begin`/`end` and the
API's mark handlers alone), so the dashboard keeps rendering them.

- **`README.md`** Level 1 table: delete the `flow.resolve-change` row.
- **`stats/internal/stages/names.go`**: delete the row. `TestStagesMatchReadmeLevelOne` pins the two
  identical, so edit both in this task and run it RED between the two edits to prove it watches.
- **`stats/internal/stages/names_test.go`** `TestFlowStatusHasNoDocumentedStages`: replace the
  example key `flow.resolve-change` with `flow.kickoff` — the assertion is about the command having
  no documented stages, not about that key.
- **`skills/flow/SKILL.md`** Stage keys table: drop `flow.resolve-change` from the
  `skills/flow/brainstorm.md` row.
- **`skills/flow/brainstorm.md`** section A: delete the `flow.resolve-change` begin/end pair; the
  `flow.kickoff` end mark stays as the last line of that block.

```bash verified:the commands stats/README.md and .flow/project.md name for this module
cd stats && go test ./internal/stages/... -race -count=1
scripts/check-stage-mark-calls.sh && scripts/check-references.sh
grep -rn "resolve-change" README.md skills stats --exclude-dir=archive ; test $? -eq 1
```

**Files:** `README.md`, `stats/internal/stages/names.go`, `stats/internal/stages/names_test.go`,
`skills/flow/SKILL.md`, `skills/flow/brainstorm.md`
**Tests:** `TestFlowStatusHasNoDocumentedStages` (existing; its example key changes in this commit) —
the README parity test is exercised RED between the README and Go edits, no new test function
**Regression:** reverting this commit restores a stage whose mark brackets nothing on every
creating run; `TestStagesMatchReadmeLevelOne` fails if only one of README or `names.go` is reverted.
**Baseline:** before=9 after=9 stages tests
<!-- predicted: grep -c '^func Test' stats/internal/stages/names_test.go after task 3 -->
**Commit:** `refactor(stages): fold flow.resolve-change into flow.kickoff`
**Build:** green

- [x] 4. `PLANNING_MODEL`: resolution, the project key, and the `/flow-settings` question

Prose only — no Go, no tests.

**`skills/flow-contracts/project-configuration.md`**: add a `## planning model` row to the optional
keys table beside `## default landing route`, same single-line-literal shape: the body is one
member of `<agents repo>/stats/internal/store/settings.go`'s `ValidModels` and nothing else,
trimmed and matched byte for byte; absent falls through to the settings store; a body matching no
member is reported by name (quoting what was found) and dropped. Add the one-paragraph matching
rule beside the landing route's.

**`skills/flow/SKILL.md`** Model resolution: add `PLANNING_MODEL` to the resolution block —
`jq -r '.planningModel // empty'` from `SETTINGS_JSON`, then the project key when present and
valid wins, then the literal `fable` when both are empty or the store is unreachable (naming the
literal a fallback exactly as `DEFAULT_MODEL`'s `sonnet` is). State the three roles it governs —
the planner dispatch, `flow.document-fix`'s planner dispatch, and `/flow-research`'s research
subagent — and that a plain-language session instruction overrides it for the run, recorded with
the dispatch and never written back. Raise this file's `check-contract-budget.sh` row.

**`skills/flow-settings/SKILL.md`**: description and intro name the fourth field; step 1 prints
`planning model: <planningModel, or "(fable — store default)" when empty>`; step 2 adds a
"Planning model" question offering `ValidModels` plus "Store default (fable)" (maps to `""`) and
"keep current"; step 3 always passes `-planning-model` explicitly, same reasoning as
`-self-review-model`. Raise this file's budget row.

**`commands-claude/flow-settings.md`** and **`commands/flow-settings.md`**: description lines name
the planning model alongside the other three fields (check both budget rows).

```bash verified:the guards .flow/project.md's lint list names for skill prose
scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-contract-budget.sh \
  && scripts/check-installed-citations.sh && scripts/check-markdown-integrity.py
```

**Files:** `skills/flow-contracts/project-configuration.md`, `skills/flow/SKILL.md`,
`skills/flow-settings/SKILL.md`, `commands-claude/flow-settings.md`, `commands/flow-settings.md`,
`scripts/check-contract-budget.sh`
**Tests:** none
**Regression:** reverting this commit leaves `planningModel` writable (task 1) but unread — the
planner dispatch (task 5) has no `PLANNING_MODEL` to name, and the project key is not a key.
**Baseline:** before=0 after=0 (no test-count-bearing files touched)
**Commit:** `docs(flow-settings): add the planning model setting and its project override`
**Build:** green

- [x] 5. Run planning on the planner subagent

Prose only. design.md's **The planner protocol** is canonical; this task writes it into the
skill in the places a run reads.

**`skills/flow/brainstorm.md`**:
- New section between A and B, **Dispatch the planner**: resolve `PLANNING_MODEL` (task 4); record
  `flow record dispatch begin -change <name> -role planner -model <PLANNING_MODEL> -key planner
  -session-token mf-<literal-token> -started-at <ts>` before the dispatch; dispatch one subagent
  with the Agent tool's `model` set to `PLANNING_MODEL`, the prompt carrying the agent-baseline
  pointer verbatim, the change name, the Jira key and issue text, the project root, `<changeRoot>`,
  and the instruction to read this file's sections B–D and follow them as the planner. The prompt
  states the relay contract: the planner cannot ask the operator; it ends a turn with exactly one
  `## Question` block, or with `## Design`, `## Artifacts` or `## Plan` at the three returns; the
  first line of its first reply is `Model: <the model named in its own system prompt>`.
- The handshake: compare that line with `PLANNING_MODEL`; on a mismatch, `flow record dispatch
  end … -outcome fallback`, re-dispatch once with `model: opus` and a fresh `begin` with `-model
  opus`; on a second mismatch continue on what answered, report it in the run's own output, and
  record that model. A mark or record never blocks.
- The relay: the parent asks each `## Question` verbatim through **AskUserQuestion** and resumes
  the planner with the answer via `SendMessage`. B's convergence confirm, third-round offer and the
  HARD GATE approval are relayed the same way, and the parent marks `flow.brainstorm` end and
  `flow.design-approval` begin/end around the approval exactly as today.
- Sections B, C and D are re-addressed to the planner where they say "you": the checklist runs in
  the planner; the seeded-note discovery and its deletion, `spectre new`, the three artifacts and
  the writing-plans enrichment are the planner's; the guards at the end of D run in the planner
  and their output comes back in `## Plan`. The parent keeps every `flow stage` mark, calling
  `flow.create-artifacts` begin/end around the `## Artifacts` return and `flow.writing-plans`
  begin/end around the `## Plan` return.
- After `## Plan`: `flow record dispatch end -change <name> -key planner -session-token
  mf-<literal-token> -outcome completed -ended-at <ts> -agent-id <id>`.
- Raise this file's budget row.

**`skills/flow/implement.md`** section 3: `flow.document-fix` dispatches the planner the same way
(`-key planner-fix-<n>`, `<n>` the fix run's ordinal), the "where should it go" prompt relayed;
the planner writes the proposal/tasks or sub-change; the Jira description sync stays in the parent.

**`skills/flow-contracts/model-policy.md`**: rewrite the planning paragraph — planning runs in a
planner subagent on `PLANNING_MODEL` (`fable` by default, `opus` on a verified fallback), parent
on sonnet; rewrite the Claude Code bullet — a command's `model:` frontmatter applies to the turn it
starts and nothing later, so it enforces no session model for a multi-turn run; models are set at
dispatch time, which is what every harness supports.

**`commands/flow.md`**: replace the "keep this session on Sonnet" paragraph with one sentence:
planning and implementation both run in subagents whose models are set at dispatch time, so the
session model is not load-bearing. **`commands-claude/flow.md`**: one clause in the creating-run
sentence — brainstorming runs in a planner subagent on the configured planning model.

```bash verified:the guards .flow/project.md's lint list names for skill prose
scripts/check-stage-mark-calls.sh && scripts/check-dispatch-paragraphs.sh \
  && scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-contract-budget.sh \
  && scripts/check-installed-citations.sh && scripts/check-markdown-integrity.py
```

**Files:** `skills/flow/brainstorm.md`, `skills/flow/implement.md`,
`skills/flow-contracts/model-policy.md`, `commands/flow.md`, `commands-claude/flow.md`,
`scripts/check-contract-budget.sh`
**Tests:** none
**Regression:** reverting this commit puts brainstorming and writing-plans back in the sonnet
session with no dispatch record; `flow.document-fix` runs in the parent again.
**Baseline:** before=0 after=0 (no test-count-bearing files touched)
**Commit:** `docs(flow): run planning on the planner subagent`
**Build:** green

- [x] 6. Run `/flow-research` on the planning model

Prose only.

**`skills/flow-research/SKILL.md`**: a new section after **The Stance**, **The research
subagent**: the parent resolves `PLANNING_MODEL` (**Model resolution**, `skills/flow/SKILL.md`),
dispatches one research subagent on it with this skill as its instructions and the topic, the
agent-baseline pointer verbatim, and the same relay contract as the planner — one `## Question`
per turn, the parent asks and resumes; the same `Model:` handshake with the `opus` fallback. The
subagent reads the tree, writes the staging note or offers the `design.md` capture through the
relay, and the guardrails bind it as they bind the session. No dispatch record — there is no
change to record against.

**`commands-claude/flow-research.md`** and **`commands/flow-research.md`**: one sentence each —
the thinking runs in a research subagent on the configured planning model. Check both budget rows.

```bash verified:the guards .flow/project.md's lint list names for skill prose
scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-contract-budget.sh \
  && scripts/check-installed-citations.sh && scripts/check-markdown-integrity.py
```

**Files:** `skills/flow-research/SKILL.md`, `commands-claude/flow-research.md`,
`commands/flow-research.md`, `scripts/check-contract-budget.sh`
**Allowed-collateral:** `scripts/check-contract-budget.sh`
**Tests:** none
**Regression:** reverting this commit leaves `/flow-research` thinking on the sonnet session while
`/flow`'s planning runs on `PLANNING_MODEL` — one setting no longer governs both.
**Baseline:** before=0 after=0 (no test-count-bearing files touched)
**Commit:** `docs(flow-research): run research on the planning-model subagent`
**Build:** green
