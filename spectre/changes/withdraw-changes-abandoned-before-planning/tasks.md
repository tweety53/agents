# withdraw-changes-abandoned-before-planning

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

- [x] 1. Store: the `withdrawn` column, its invariant, migration 0031
  - [x] **Step 1: Add `stats/internal/store/migrations/0031_withdrawn.sql` with `ALTER TABLE changes ADD COLUMN withdrawn BOOLEAN NOT NULL DEFAULT FALSE;` and add the `Withdrawn bool` field to `store.Change` in `stats/internal/store/changes.go`.**
  - [x] **Step 2: Persist and scan the column in `PutChange` and the change reads. Refuse `Withdrawn == true` paired with `State != StateFinished` by returning `ErrInvalidState` (extend its wrapped message; keep the sentinel so `errors.Is` and `IsDefinitiveChangeOutcome` behave unchanged).**
  - [x] **Step 3: Tests in `stats/internal/store/changes_test.go` — `TestPutChangeWithdrawnRoundTrip` (write `FINISHED` with `Withdrawn: true`, read it back true; write a plain `FINISHED`, read back false) and `TestPutChangeRefusesWithdrawnOutsideFinished` (`STARTED` with `Withdrawn: true` → `errors.Is(err, ErrInvalidState)`, record unchanged). Verify: `cd stats && go test ./internal/store/... -race -count=1 && gofmt -l . && go vet ./internal/store/...`.**
**Build:** green
**Files:** `stats/internal/store/changes.go`, `stats/internal/store/migrations/0031_withdrawn.sql`, `stats/internal/store/changes_test.go`
**Tests:** `TestPutChangeWithdrawnRoundTrip`, `TestPutChangeRefusesWithdrawnOutsideFinished`
**Regression:** reverting the commit drops the column and the invariant: `TestPutChangeWithdrawnRoundTrip` fails (the field no longer round-trips) and `TestPutChangeRefusesWithdrawnOutsideFinished` fails (a non-FINISHED record can carry the marker).
**Baseline:** before=23 after=25
**Commit:** feat(store): withdrawn marker on finished changes
**After:** none

**Decision:** finished-withdrawn-boolean

- [x] 2. API: decode and serve `withdrawn` on the closed-schema DTO
  - [x] **Step 1: Add `Withdrawn bool \`json:"withdrawn,omitempty"\`` to `changeDTO` in `stats/internal/api/changes.go` and map it in `toChange`/the Change→DTO direction, so `DecodeChangeBody` (the one decode path live writes and journal replay share) accepts the field and reads serve it.**
  - [x] **Step 2: Tests in `stats/internal/api/changes_test.go` — `TestDecodeChangeBodyWithdrawnRoundTrip` (a body with `"withdrawn": true` decodes to `Change.Withdrawn` and encodes back with the field present; a body without it decodes false) and `TestGetChangeServesWithdrawn` (a stored withdrawn record serves `"withdrawn": true`; a plain record omits it). Run `cd stats && go test ./internal/api/... -race -count=1`.**
  - [x] **Step 3: Confirm the unknown-field refusal still stands for every other undocumented name — run the existing decode-refusal cases in the package suite. Verify: `cd stats && go test ./internal/api/... -race -count=1 && gofmt -l . && go vet ./internal/api/...`.**
**Build:** green
**Files:** `stats/internal/api/changes.go`, `stats/internal/api/changes_test.go`
**Tests:** `TestDecodeChangeBodyWithdrawnRoundTrip`, `TestGetChangeServesWithdrawn`
**Baseline:** before=21 after=23
**Regression:** reverting the commit closes the schema again: every `state set` carrying `withdrawn` is refused as a malformed body (`DisallowUnknownFields`), so the withdrawal route's state write falls to the journal and is retired — both declared tests fail.
**Commit:** feat(api): accept and serve withdrawn on changes
**After:** Task 1

**Decision:** finished-withdrawn-boolean

- [x] 3. Skill and contract docs: the withdrawal route and its ends
  - [x] **Step 1: `skills/flow-contracts/state-file.md` — add `withdrawn` to the record's field vocabulary (boolean; absent/false = not withdrawn; written only by the withdrawal route, only paired with `FINISHED`; the CLI passes it through byte-transparently, so no CLI change).**
  - [x] **Step 2: `skills/flow-contracts/pipeline.md` — the States table's `FINISHED` row gains "or withdrawn (never planned)"; the transitions table's creating-run row gains the withdraw end; the wrong-state handoff for a `FINISHED` change distinguishes "archived" from "withdrawn".**
  - [x] **Step 3: `skills/flow/brainstorm.md` — state the route at its two offer points and its steps: per-worktree `git status --porcelain` must be empty (else stop and show it), then remove the worktree and prune, `git branch -D spectre/<name>`, `git push origin --delete spectre/<name>` (a failed remote delete is one reported line, never a blocker), then one state write `FINISHED` with `withdrawn: true` and `worktrees: {}` against a daemon that knows the field — a write that falls back warns and the route re-runs.**
  - [x] **Step 4: `skills/flow-contracts/handoff-blocks.md` (the withdrawn terminal block — names no next command), `skills/flow/SKILL.md` (the reading-the-state `FINISHED` bullet names withdrawn), `skills/flow-status/SKILL.md` (a withdrawn change is not open). If `check-contract-budget.sh` trips on an edited contract, raise that file's row in the `budgets()` table of `scripts/check-contract-budget.sh` — the declared reason, never a narrowed scope.**
  - [x] **Step 5: Verify: `scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-contract-budget.sh && scripts/check-markdown-integrity.py && scripts/check-plan-shape.sh spectre/changes/withdraw-changes-abandoned-before-planning/tasks.md`.**
**Build:** green
**Files:** `skills/flow-contracts/state-file.md`, `skills/flow-contracts/pipeline.md`, `skills/flow/brainstorm.md`, `skills/flow/brainstorm-planner.md`, `skills/flow-contracts/handoff-blocks.md`, `skills/flow/SKILL.md`, `skills/flow-status/SKILL.md`, `AGENTS.md`, `rules/flow-manual-review.mdc`, `KNOWN-BUGS.md`
**Allowed-collateral:** `scripts/check-contract-budget.sh`
**Tests:** **none**
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): the withdrawal route
**After:** none

**Decision:** both-preplan-ends

**Decision:** finished-withdrawn-boolean

**Decision:** git-first-record-last

Correction (2026-09-28): the plan declared modifying `scripts/check-contract-budget.sh` (raising
`skills/flow/brainstorm.md`'s budget row) as task 3's Allowed-collateral, and steps 4–5 name the
guard. The base's kan-842 port deleted that guard outright — no file on new `main` carries
`budgets()` — so the rebase resolve dropped our modification (`git rm`) instead of re-homing it:
the raise fed a guard that no longer exists. The task's remaining verify guards all ran green
post-rebase; no replacement budget table exists on the base to raise.

Correction (2026-09-28, panel fix round 1): the panel's round-0 findings widened this task's
surface — `AGENTS.md` and `rules/flow-manual-review.mdc` (the sanctioned digest copies gained the
reachability-ended `FINISHED (withdrawn)` line, F1), `KNOWN-BUGS.md` (a line-ref fix and the
round's deferrals, F2/F3/F7), and `stats/internal/store/changes.go` + `changes_test.go` (the
`ErrInvalidState` doc comment now names the withdrawn refusal, F4, and its message carries the
actual state, F6) — the Go hunks ride task 1's declared surface; the digest, deferral and route
hunks (brainstorm.md's step 2/3 gone-tolerance clauses, F5) are this task's.

Correction (2026-09-28): the plan declared modifying `scripts/check-contract-budget.sh` (raising
`skills/flow/brainstorm.md`'s budget row) as task 3's Allowed-collateral, and steps 4–5 name the
guard. The base's kan-842 port deleted that guard outright — no file on new `main` carries
`budgets()` — so the rebase resolve dropped our modification (`git rm`) instead of re-homing it:
the raise fed a guard that no longer exists. The task's remaining verify guards all ran green
post-rebase; no replacement budget table exists on the base to raise.
