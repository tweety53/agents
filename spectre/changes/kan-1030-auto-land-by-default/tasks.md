# kan-1030-auto-land-by-default

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

- [ ] 1. Store: the `review_before_landing` column, migration 0045
  - [ ] **Step 1: Add `stats/internal/store/migrations/0045_review_before_landing.sql` — a header comment naming the field's meaning (true keeps `/flow` at the `IN_PROGRESS` review handoff; false/absent auto-lands) and `ALTER TABLE changes ADD COLUMN review_before_landing BOOLEAN NOT NULL DEFAULT FALSE;` — and add `ReviewBeforeLanding bool` (doc comment) to `store.Change` in `stats/internal/store/changes.go`.**
  - [ ] **Step 2: Persist it in `PutChange` (INSERT column list, `ON CONFLICT ... SET review_before_landing = EXCLUDED.review_before_landing`, argument list) and scan it in every change read that already scans `withdrawn` (the two SELECTs and the scan helper).**
  - [ ] **Step 3: Test in `stats/internal/store/changes_test.go`, beside `TestPutChangeWithdrawnRoundTrip`: `TestPutChangeReviewBeforeLandingRoundTrip` — write `STARTED` with `ReviewBeforeLanding: true`, read back true; write `IN_PROGRESS` (later `UpdatedAt`) with it false, read back false. Verify: `cd stats && go test ./internal/store/... -run 'TestPutChangeReviewBeforeLandingRoundTrip' -race -count=1 && gofmt -l . && go vet ./internal/store/...`.**
**Build:** green
**Files:** `stats/internal/store/changes.go`, `stats/internal/store/migrations/0045_review_before_landing.sql`, `stats/internal/store/changes_test.go`
**Tests:** `TestPutChangeReviewBeforeLandingRoundTrip`
**Regression:** reverting the commit drops the column: `TestPutChangeReviewBeforeLandingRoundTrip` fails because the field no longer round-trips.
**Baseline:** before=25 after=26 <!-- measured: grep -c '^func Test' stats/internal/store/changes_test.go @ ebfb5504 → 25 -->
**Commit:** feat(store): record whether a change keeps the review handoff
**After:** none

**Decision:** review-flag-in-state-record

- [ ] 2. API: decode and serve `reviewBeforeLanding`
  - [ ] **Step 1: Add `ReviewBeforeLanding bool \`json:"reviewBeforeLanding,omitempty"\`` to `changeDTO` in `stats/internal/api/changes.go` and map it in both directions, beside `Withdrawn`, so `DecodeChangeBody` accepts the field and reads serve it.**
  - [ ] **Step 2: Tests in `stats/internal/api/changes_test.go`, modelled on the withdrawn pair: `TestDecodeChangeBodyReviewBeforeLandingRoundTrip` (a body with `"reviewBeforeLanding": true` decodes true and encodes back with the field; a body without it decodes false) and `TestGetChangeServesReviewBeforeLanding` (a stored true record serves the field; a plain record omits it). Verify: `cd stats && go test ./internal/api/... -run 'ReviewBeforeLanding' -race -count=1 && gofmt -l . && go vet ./internal/api/...`.**
**Build:** green
**Files:** `stats/internal/api/changes.go`, `stats/internal/api/changes_test.go`
**Tests:** `TestDecodeChangeBodyReviewBeforeLandingRoundTrip`, `TestGetChangeServesReviewBeforeLanding`
**Regression:** reverting the commit closes the schema again: a `state set` carrying `reviewBeforeLanding` is refused as a malformed body (`DisallowUnknownFields`) and falls to the journal — both declared tests fail.
**Baseline:** before=23 after=25 <!-- measured: grep -c '^func Test' stats/internal/api/changes_test.go @ ebfb5504 → 23 -->
**Commit:** feat(api): serve a change's review-before-landing flag
**After:** Task 1

**Decision:** review-flag-in-state-record

- [ ] 3. Contracts and skill: land straight after implementation unless the flag is set
  - [ ] **Step 1: `skills/flow-contracts/state-file.md` — add `reviewBeforeLanding` to the record's shape and field vocabulary: boolean, absent/false = auto-land, true = stop at the `IN_PROGRESS` review handoff; written by any `/flow` run of the change when the operator asks (decision `override-any-time`), carried forward verbatim otherwise; the CLI passes it through byte-transparently.**
  - [ ] **Step 2: `skills/flow-contracts/pipeline.md` — the **States** block and table and the **State transitions** rows for `/flow` at `STARTED` and the fix run now end at `IN_PROGRESS` only when `reviewBeforeLanding` is true, otherwise continue into integrate in the same invocation (ending as an integrate run ends); rewrite **Every invocation is re-entrant** per decision `reentrancy-rule-narrowed`; the `/clear` bullet of **Handoff output** applies only to a handoff actually printed at `IN_PROGRESS`. Add the reasoning under `### Every command is re-entrant` in `skills/flow-contracts/pipeline-rationale.md`.**
  - [ ] **Step 3: `skills/flow/verify-and-handoff.md` **Write `IN_PROGRESS`** — after the state write and stage end, re-read the record (`flow state get`); `reviewBeforeLanding` true prints the existing block as today; otherwise print no block (the summary is already recorded) and continue at **Deciding which run this is** (`skills/flow/integrate.md`) in the same turn. `skills/flow/integrate.md`'s opening line names that chain as its second entry. Carry `reviewBeforeLanding` forward in the state-write field list.**
  - [ ] **Step 4: `skills/flow/SKILL.md` — a short section **Review-before-landing instruction** (decision `override-any-time`): a plain-language instruction to review before landing, or to land without review, at any point of a `/flow` run (planning included) is written at once via `flow state get <name> | jq '.reviewBeforeLanding = <bool>' | flow state set <name>`, noted in the narrative and a `Decisions:` bullet; latest wins; never itself an integrate trigger. Update the **Reading the state** bullets that describe ends at `IN_PROGRESS`.**
  - [ ] **Step 5: `rules/flow-manual-review.mdc` — the state diagram's `(STARTED, planned)` and fix lines read `→ IN_PROGRESS or FINISHED` with "lands straight away unless review-before-landing is set"; re-render its copies `AGENTS.md`, `templates/AGENTS.md` and the diagram in `skills/README.md` exactly as `scripts/check-installed-rules.sh` expects.**
  - [ ] **Step 6: Verify: `scripts/check-references.sh && scripts/check-markdown-integrity.sh && scripts/check-installed-rules.sh && scripts/check-normative-inventory.sh && scripts/check-verbatim-moves.sh && scripts/check-ask-silence.sh && scripts/check-plan-shape.sh spectre/changes/kan-1030-auto-land-by-default/tasks.md`.**
**Build:** green
**Files:** `skills/flow-contracts/state-file.md`, `skills/flow-contracts/pipeline.md`, `skills/flow-contracts/pipeline-rationale.md`, `skills/flow/verify-and-handoff.md`, `skills/flow/integrate.md`, `skills/flow/SKILL.md`, `rules/flow-manual-review.mdc`, `AGENTS.md`, `templates/AGENTS.md`, `skills/README.md`
**Allowed-collateral:** `skills/flow-contracts/handoff-blocks.md`, `skills/flow/implement.md`, `skills/flow-status/SKILL.md`, `CLAUDE.md`
**Tests:** **none**
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** feat(flow): land straight after implementation unless review is asked for
**After:** none

**Decision:** auto-land-default

**Decision:** fix-runs-follow-flag

**Decision:** override-any-time

**Decision:** reentrancy-rule-narrowed
