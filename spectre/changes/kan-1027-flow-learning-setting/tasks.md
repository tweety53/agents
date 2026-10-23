# kan-1027-flow-learning-setting

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** a harness-wide `learning` setting (default on) and a capped topic history in flowd, read
by `/flow`'s implementation run and `/flow-fast` to print a `## Learning` section in their final
message.

**Architecture:** store (`flow_settings.learning`, `flow_learning_topics`) → API (`learning` in the
settings DTO, `GET|POST /api/v1/learning` on the settings handler) → CLI (`flow settings set
-learning`, `flow learning add|list`) → one contract file `skills/flow-contracts/learning.md` cited
by the phase files that compose and print the note.

**Spec:** `design.md` of this change.

## Global constraints

- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1027-flow-learning-setting`;
  every `**Files:**` path is relative to it. Never touch the main checkout, never create a worktree.
- Never stop flowd or flow-postgres and never drop the `flow` database; store tests run on
  `storetest.NewStore`'s own database.
- Go: `cd stats && gofmt -l . && make -s web-dist && go vet ./...` clean before each Go commit.
- Comment density, naming and idiom match the surrounding code: each exported identifier carries a
  doc comment in the style of `settings.go`'s, citing this change's decision IDs where the code
  embodies one.

## Review Focus

- `flow settings set -learning off` keeps the stored reviewer list, and `-reviewers` alone keeps
  `learning` (`settings-set-merges`).
- A fourth `flow learning add` of one topic is refused by the store, not by prose (`cap-in-store`),
  and the refusal surfaces as HTTP 409 and CLI exit 1.
- The contract states the note's composition once, in `skills/flow-contracts/learning.md`; the
  phase files cite it and restate nothing.

- [x] 1. Store the learning setting and the topic history
  - [x] **Step 1: Write the failing tests.** In `stats/internal/store/settings_test.go` add
    `TestSettingsStore_LearningDefaultsOn` (no row written → `GetSettings().Learning == true`) and
    `TestSettingsStore_LearningRoundTrip` (put `Learning: false` with a reviewer list, get it back
    equal). Create `stats/internal/store/learningtopics_test.go` with
    `TestLearningTopics_ListCounts` (record `go-generics` twice and `cqrs` once → 
    `ListLearningTopics` returns both with counts 2 and 1 and each topic's newest note),
    `TestLearningTopics_CapAtThree` (a fourth `RecordLearningTopic` of one topic returns an error
    `errors.Is` `store.ErrLearningTopicExhausted` and writes nothing) and
    `TestLearningTopics_RejectsBadSlug` (`Go Generics`, an empty string and a 65-character slug
    each return `store.ErrInvalidLearningTopic`). Run
    `cd stats && go test ./internal/store/ -run 'TestSettingsStore_Learning|TestLearningTopics' -count=1`:
    FAIL (undefined symbols).
  - [x] **Step 2: Migration.** Create `stats/internal/store/migrations/0043_flow_learning.sql`,
    header comment in the style of `0034_flow_settings_drop_model_columns.sql`:
    `ALTER TABLE flow_settings ADD COLUMN learning BOOLEAN NOT NULL DEFAULT TRUE;`,
    `CREATE TABLE flow_learning_topics (id BIGSERIAL PRIMARY KEY, topic TEXT NOT NULL, note TEXT NOT NULL, change TEXT NOT NULL, project_key TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now());`
    and `CREATE INDEX flow_learning_topics_topic ON flow_learning_topics (topic);`. The cap is
    harness-wide — counted over every project — since the operator reads one stream of notes.
  - [x] **Step 3: Settings.** In `stats/internal/store/settings.go` add `Learning bool` to
    `Settings`, write it in `PutSettings`' upsert, read it in `GetSettings`, and return
    `Learning: true` on the no-row path.
  - [x] **Step 4: Topics.** Create `stats/internal/store/learningtopics.go`: a
    `LearningTopic{Topic, Note, Change, ProjectKey string}` input, a
    `LearningTopicCount{Topic string; Count int; LastNote string}` row,
    `ErrInvalidLearningTopic`, `ErrLearningTopicExhausted`, `const MaxLearningTopicShows = 3`,
    `RecordLearningTopic(ctx, in LearningTopic) error` — slug checked against
    `^[a-z0-9]+(-[a-z0-9]+)*$` and at most 64 bytes, an empty note refused with the same invalid
    error, the count read and the insert made in one transaction holding
    `pg_advisory_xact_lock(hashtext(topic))` so two concurrent adds cannot both pass the cap — and
    `ListLearningTopics(ctx) ([]LearningTopicCount, error)` ordered by topic.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l . && make -s web-dist && go vet ./... && go test ./internal/store/ -run 'TestSettingsStore|TestLearningTopics' -count=1`
    passes, the existing `TestSettingsStore_*` cases included.

**Files:** `stats/internal/store/migrations/0043_flow_learning.sql`, `stats/internal/store/settings.go`, `stats/internal/store/settings_test.go`, `stats/internal/store/learningtopics.go`, `stats/internal/store/learningtopics_test.go`
**Tests:** `TestSettingsStore_LearningDefaultsOn`, `TestSettingsStore_LearningRoundTrip`, `TestLearningTopics_ListCounts`, `TestLearningTopics_CapAtThree`, `TestLearningTopics_RejectsBadSlug`
**Regression:** `TestSettingsStore_LearningDefaultsOn` and `TestSettingsStore_LearningRoundTrip` fail without the column and field; the three `TestLearningTopics_*` fail without `learningtopics.go`.
**After:** none
**Commit:** `feat(store): record the learning setting and capped topic history`

**Decision:** learning-in-settings
**Decision:** learning-default-on
**Decision:** topics-in-flowd
**Decision:** cap-in-store
**Decision:** topic-slug

**Build:** green

- [x] 2. Serve the learning setting and the topic history
  - [x] **Step 1: Write the failing tests.** In `stats/internal/api/settings_test.go` add
    `TestSettingsHandler_LearningRoundTrip` (PUT `{"reviewers":["primary"],"learning":false}` then
    GET returns `"learning":false`) and `TestSettingsHandler_LearningAbsentReadsOn` (PUT without
    `learning` stores `Learning: true`). Create `stats/internal/api/learning_test.go` with
    `TestLearningHandler_AddAndList` (POST `{"topic":"cqrs","note":"n","change":"c","projectKey":"p"}`
    → 201; GET → `[{"topic":"cqrs","count":1,"lastNote":"n"}]`),
    `TestLearningHandler_ExhaustedIs409` and `TestLearningHandler_BadTopicIs400`, on the fake
    store `settings_test.go` already uses, extended with the two new methods. Run
    `cd stats && make -s web-dist && go test ./internal/api/ -run 'TestSettingsHandler_Learning|TestLearningHandler' -count=1`:
    FAIL.
  - [x] **Step 2: Implement.** In `stats/internal/api/settings.go` add an optional `learning`
    field to `settingsDTO` (a `*bool`, absent read as `true`), and give `SettingsStore`
    `RecordLearningTopic` and `ListLearningTopics` — the learning endpoints share the settings
    handler's store, so `api.New`'s signature is unchanged. Create `stats/internal/api/learning.go`
    with the handler's `listLearning` and `addLearning`: `ErrInvalidLearningTopic` → 400 naming the
    value, `ErrLearningTopicExhausted` → 409, anything else through `mapStoreError`. Register
    `GET /api/v1/learning` and `POST /api/v1/learning` in `stats/internal/api/server.go` beside the
    settings routes. In `stats/internal/client/client.go` add `Learning bool` to `client.Settings`
    and `RecordLearningTopic` / `ListLearningTopics`, a 409 surfacing as
    `client.ErrLearningTopicExhausted` and a 400 as `client.ErrLearningTopicRejected`; give
    `stubStageStore` in `stats/internal/client/client_test.go` the two methods so it still
    satisfies `api.SettingsStore`.
  - [x] **Step 3: Verify.** `cd stats && gofmt -l . && make -s web-dist && go vet ./... && go test ./internal/api/ ./internal/client/ -run 'Settings|Learning' -count=1`
    passes.

**Files:** `stats/internal/api/settings.go`, `stats/internal/api/settings_test.go`, `stats/internal/api/learning.go`, `stats/internal/api/learning_test.go`, `stats/internal/api/server.go`, `stats/internal/client/client.go`, `stats/internal/client/client_test.go`, `stats/internal/api/changes_test.go`, `stats/internal/web/embed_test.go`

Correction (2026-10-23): the plan declared the fake-store extension in `settings_test.go` alone; the fake's topic slice is a field of `fakeStore`, declared in `stats/internal/api/changes_test.go`, and `stats/internal/web/embed_test.go`'s own `fakeStore` also implements `api.SettingsStore`, so both gained the two methods' support to keep compiling.
**Tests:** `TestSettingsHandler_LearningRoundTrip`, `TestSettingsHandler_LearningAbsentReadsOn`, `TestLearningHandler_AddAndList`, `TestLearningHandler_ExhaustedIs409`, `TestLearningHandler_BadTopicIs400`
**Regression:** the two settings cases fail without `learning` in the DTO; the three learning cases fail without `learning.go` and its routes.
**After:** Task 1
**Commit:** `feat(api): serve the learning setting and topic history`

**Decision:** settings-set-merges
**Decision:** cap-in-store

**Build:** green

- [x] 3. Add `flow settings set -learning` and `flow learning`
  - [x] **Step 1: Write the failing tests.** In `stats/cmd/flow/settings_test.go` add
    `TestSettingsSet_LearningKeepsReviewers` (store holds `primary,mutation` and learning on;
    `set -learning off` → stored reviewers unchanged, learning false),
    `TestSettingsSet_ReviewersKeepLearning` (learning off stored; `set -reviewers primary` keeps it
    off), `TestSettingsSet_NeedsAFlag` (no flag → exit 2) and
    `TestSettingsSet_LearningRejectsBadValue` (`-learning maybe` → exit 2). Create
    `stats/cmd/flow/learning_test.go` with `TestLearningAdd_Records`,
    `TestLearningAdd_ExhaustedExits1` (stderr names the three-times cap) and
    `TestLearningList_PrintsCounts` (one JSON array line), on the fake-server pattern
    `settings_test.go` already uses. Run
    `cd stats && make -s web-dist && go test ./cmd/flow/ -run 'TestSettingsSet_|TestLearning' -count=1`: FAIL.
  - [x] **Step 2: Implement.** In `stats/cmd/flow/settings.go` add `-learning on|off` beside
    `-reviewers`, at least one required; `set` reads the current record with `getSettings`,
    overrides the given fields and puts the whole record; update `settingsUsage`. Create
    `stats/cmd/flow/learning.go` with `runLearning`:
    `flow learning add <topic> -note '<one line>' -change <name> [-C dir] [-addr url] [-timeout dur]`
    (project key resolved from `-C` as `flow self-review finding` resolves it; exit 0 on write,
    1 on a store refusal or failure, 2 on usage) and `flow learning list [-addr url] [-timeout dur]`
    (prints the JSON array; exit 1 when the store is unreachable). Route `learning` in
    `stats/cmd/flow/main.go` and add its two lines to the top-level usage.
  - [x] **Step 3: Verify.** `cd stats && gofmt -l . && make -s web-dist && go vet ./... && go test ./cmd/flow/ -run 'Settings|Learning' -count=1`
    passes.

**Files:** `stats/cmd/flow/settings.go`, `stats/cmd/flow/settings_test.go`, `stats/cmd/flow/learning.go`, `stats/cmd/flow/learning_test.go`, `stats/cmd/flow/main.go`
**Tests:** `TestSettingsSet_LearningKeepsReviewers`, `TestSettingsSet_ReviewersKeepLearning`, `TestSettingsSet_NeedsAFlag`, `TestSettingsSet_LearningRejectsBadValue`, `TestLearningAdd_Records`, `TestLearningAdd_ExhaustedExits1`, `TestLearningList_PrintsCounts`
**Regression:** the settings cases fail when `set` replaces instead of merging or lacks `-learning`; the learning cases fail without the `learning` command.
**After:** Task 2
**Commit:** `feat(flow): add the learning setting flag and the learning command`

**Decision:** settings-set-merges
**Decision:** topics-in-flowd

**Build:** green

- [x] 4. Compose and print the learning note
  - [x] **Step 1: The contract.** Create `skills/flow-contracts/learning.md`, canonical for the
    note: when it runs (the runs `learning-scope` names), reading the setting
    (`flow settings get | jq .learning`; an unreachable store prints no note and never blocks),
    choosing a topic (`flow learning list`, exhausted topics skipped; related first — the plan's
    language, a design approach its tasks use, or agentic-coding / LLM / AI-infrastructure
    practice the run exercises — else a random programming-language or system-design pattern),
    the note's shape (a theory paragraph of at most about 150 words, a coding puzzle with its
    answer behind a `<details>` block where the topic admits one, two or three study links each
    checked with `curl -sfI --max-time 5 <url>` and dropped on a non-zero exit), recording it
    (`flow learning add <topic> -note '<the topic in under 80 characters>' -change <name> -C <worktree>`;
    a refusal picks another topic), and the printed section's shape: a `## Learning — <topic>`
    heading, the theory, a `**Puzzle:**` line where one applies, and a `**Study:**` line of links.
  - [x] **Step 2: Cite it.** One sentence each, restating nothing of the contract:
    `skills/flow/implement.md` **1. Load context and validate the plan** — on an implementation
    run, not a fix run, compose the note per **Learning** (`skills/flow-contracts/learning.md`)
    once the plan validates; `skills/flow/verify-and-handoff.md` **Write `IN_PROGRESS`** and
    `skills/flow-contracts/handoff-blocks.md`'s `IN_PROGRESS` rendering — a `## Learning` section,
    printed by the run only, after `Decisions`; `skills/flow-fast/SKILL.md` — compose at **4.
    Implement** and print in the final report; `skills/flow-settings/SKILL.md` — `learning` shown
    and set with `flow settings set -learning on|off`. Add `learning.md` to the `flow-contracts`
    row of the skill index in `CLAUDE.md` and `AGENTS.md`.
  - [x] **Step 3: Verify.** `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`,
    `scripts/check-normative-inventory.sh` and `scripts/check-installed-citations.sh` exit 0; fix
    every hit.

**Files:** `skills/flow-contracts/learning.md`, `skills/flow/implement.md`, `skills/flow/verify-and-handoff.md`, `skills/flow-contracts/handoff-blocks.md`, `skills/flow-fast/SKILL.md`, `skills/flow-settings/SKILL.md`, `CLAUDE.md`, `AGENTS.md`, `templates/CLAUDE.md`, `templates/AGENTS.md`

Correction (2026-10-23): the plan named `CLAUDE.md` and `AGENTS.md` alone; their skill-index row is rendered from `templates/CLAUDE.md` and `templates/AGENTS.md`, which carry the same row, so both templates gained the `learning.md` entry too.
**Tests:** none — prose; the guards in Step 3 check its references.
**Regression:** none — no test declared.
**After:** Task 3
**Commit:** `feat(flow): print a learning note at implementation start`

**Decision:** note-in-handoff
**Decision:** learning-scope
**Decision:** compose-inline
**Decision:** no-spa-view

**Build:** green
