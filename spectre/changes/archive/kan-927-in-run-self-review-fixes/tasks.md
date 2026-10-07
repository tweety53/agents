# kan-927-in-run-self-review-fixes

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Task 1 serves the findings as a stats view and task 2 draws it. Task 3 makes the pass's fix loop cheap and stored, and task 4 moves the pass into `/flow`'s integrate and `/flow-fast`. `design.md` is canonical for every decision, and each task cites its entry by ID.

**Global constraints**

- No task starts, stops or restarts `flowd` on `127.0.0.1:4173` or `flow-postgres` (`CLAUDE.md`).
- Go: `cd stats && gofmt -l . && go vet ./...` clean after every Go task; tests run with `-race -count=1`.
- SPA: `cd stats/web && npx tsc -b` clean after every SPA task.
- Prose tasks keep every normative sentence they don't change (cut, don't paraphrase); each reworded normative sentence is acknowledged in `spectre/changes/kan-927-in-run-self-review-fixes/verbatim-moves.txt` for `scripts/check-verbatim-moves.sh`.

- [x] 1. `self-review` stats view in the store and the API
  - [x] **Step 1: Tests first. `stats/internal/store/selfreviewfindings_test.go` — `TestSelfReviewFindingsInPeriod`: rows of two changes in two projects and two days; the period and project filters select exactly the matching rows, newest first, with `blastRadius` and `ref` carried. `stats/internal/api/stats_test.go` — `TestStatsSelfReviewView`: `GET /api/v1/stats/self-review?from=…&to=…` returns the envelope with one row per finding; a `fixed` row's `commitUrl` is `<commit base>/<ref>` when the server was built `WithCommitBase("https://github.com/o/r/commit")`, and empty for `filed`/`declined` rows or with no base. `stats/cmd/flowd/wiring_test.go` — `TestCommitBaseFromRemote`: `git@github.com:o/r.git` and `https://github.com/o/r.git` both map to `https://github.com/o/r/commit`; a non-GitHub remote or no remote maps to `""`. Run red: `cd stats && go test ./internal/store/ ./internal/api/ ./cmd/flowd/ -run 'TestSelfReviewFindingsInPeriod|TestStatsSelfReviewView|TestCommitBaseFromRemote' -count=1`.**
  - [x] **Step 2: Implement:**
    - `Store.SelfReviewFindingsInPeriod(ctx, period Period, project *string)` in `stats/internal/store/selfreviewfindings.go`, newest first, carrying `project_key`.
    - `viewSelfReview viewName = "self-review"` in `stats/internal/api/stats.go`, registered in the view list, backed by a `StatsStore` method and a DTO `{recordedAt, project, change, angle, note, disposition, ref, blastRadius, commitUrl}`.
    - The `WithCommitBase(string) Option` in `stats/internal/api/server.go`.
    - `stats/cmd/flowd/main.go` reads `git remote get-url origin` from its working directory once at start and passes the mapped base. An error maps to `""`, never fatal.

    Run green.**
  - [x] **Step 3: Verify: `cd stats && go test ./internal/store/ ./internal/api/ ./cmd/flowd/ -run 'TestSelfReviewFindingsInPeriod|TestStatsSelfReviewView|TestCommitBaseFromRemote' -race -count=1 && gofmt -l . && go vet ./...`.**
**Build:** green
**Files:** `stats/internal/store/selfreviewfindings.go`, `stats/internal/store/selfreviewfindings_test.go`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/internal/api/server.go`, `stats/cmd/flowd/main.go`, `stats/cmd/flowd/wiring_test.go`, `stats/internal/web/embed_test.go`, `stats/internal/client/client_test.go`
**Allowed-collateral:** `stats/internal/api/*_test.go`, `stats/README.md`
**Tests:** `TestSelfReviewFindingsInPeriod`, `TestStatsSelfReviewView`, `TestCommitBaseFromRemote`
**Regression:** reverting the commit removes the view, so the route answers unknown-view; all three tests are deleted with it, and the SPA's fetch for the view fails.
**Baseline:** before=449 after=452
<!-- measured: grep -h '^func Test' stats/internal/store/*_test.go stats/internal/api/*_test.go stats/cmd/flowd/*_test.go | wc -l @ branch spectre/kan-927-in-run-self-review-fixes -->
**Commit:** feat(stats): serve self-review findings as a stats view with commit links
**After:** none

**Decision:** self-review-view

Correction (2026-10-08): the plan declared seven files; the commit also touches `stats/internal/web/embed_test.go` and `stats/internal/client/client_test.go`, whose fakes implement `api.StatsStore` and must grow `SelfReviewFindingsInPeriod` to compile. The view also skips the `unmeasured` probe, as the health views do: findings carry no token measurement.

- [x] 2. "Self-review fixes" view in the SPA
  - [x] **Step 1: Test first. `stats/web/src/views/views.test.tsx` — add an envelope for `self-review` and a case `renders self-review fixes`:
    - A `fixed` row renders its ref as a link to its `commitUrl`. A `filed` row renders its key as text.
    - The outcome column is filterable.
    - The per-change table shows `fixed`/`filed`/`declined` counts derived from the rows.
    - An empty period shows `No self-review findings in this period.`

    Run red: `cd stats/web && npx vitest run src/views/views.test.tsx -t 'renders self-review fixes'`.**
  - [x] **Step 2: Implement:**
    - Add `"self-review"` to `ViewName`, the view list and a `SelfReviewRow` type in `stats/web/src/api.ts`.
    - Add `stats/web/src/views/SelfReview.tsx` on the `Reviewers.tsx` pattern: a `ViewFrame` titled `Self-review fixes`, two stat panels (Fixed, Filed), the findings `DataTable` (Recorded, Change, Angle, Finding, Outcome — filterable, Ref), and a per-change `DataTable` (Change, Fixed, Filed, Declined).
    - Add the label `Self-review fixes` and the component to `stats/web/src/App.tsx`.

    Run green.**
  - [x] **Step 3: Verify: `cd stats/web && npx vitest run src/views/views.test.tsx -t 'renders self-review fixes' && npx tsc -b`.**
**Build:** green
**Files:** `stats/web/src/api.ts`, `stats/web/src/views/SelfReview.tsx`, `stats/web/src/App.tsx`, `stats/web/src/views/views.test.tsx`, `stats/web/src/views/RunDetail.test.tsx`
**Allowed-collateral:** `stats/web/src/App.test.tsx`, `stats/web/src/api.test.ts`, `stats/web/tests/visual/**`
**Tests:** `renders self-review fixes`
**Regression:** reverting the commit drops the tab and the view; `renders self-review fixes` is deleted with it.
**Baseline:** before=17 after=18
<!-- measured: grep -cE '^\s*(it|test)\(' stats/web/src/views/views.test.tsx @ branch spectre/kan-927-in-run-self-review-fixes -->
**Commit:** feat(stats-web): add the Self-review fixes view
**After:** Task 1

**Decision:** self-review-view

Correction (2026-10-08): the plan declared four files; the commit also touches `stats/web/src/views/RunDetail.test.tsx`, whose `adds no view` case pins `VIEW_NAMES` at 8 and now counts 9. The visual suite (`tests/visual/full-app-suite.spec.ts`) gains no entry here: a new capture needs a baseline, which `flow.visual-verify` owns.

- [x] 3. Cheap, stored fix loop: `/flow-self-review` step 3 and in-run pipeline fixes
  - [x] **Step 1: `skills/flow-self-review/SKILL.md` step 3:**
    - The fixer is one `flow-medium` dispatch on `opus`, or on `sonnet` when every fix is a literal edit with nothing it can break (reason recorded with the dispatch).
    - The first review is `flow-medium` `opus`; every re-review is `flow-medium` `sonnet`.
    - **Verify** runs only the `.flow/project.md` `## lint` lines the touched files need, plus the tests covering them (`go test` of the touched packages, the `scripts/test-*.sh` harness of a touched script, `npx vitest run <file>` of a touched SPA file). The full `## lint` and `## test` lists are gone from it.
    - **Land once** becomes: when the pass runs inside a run, the fixes go where `design.md` `fixes-ride-the-change` says. Standalone, they keep today's merge-and-push.
    - Step 1's default-branch refusal applies to the standalone command only.
    - The header's model sentence names the new pairs.
  - [x] **Step 2: `skills/flow-contracts/pipeline.md` **Pipeline defects found mid-run**:**
    - **Every dispatch in the loop is one-shot** runs the fix on `flow-medium` `opus` (or `sonnet` per the same rule), the first review on `flow-medium` `opus`, and re-reviews on `flow-medium` `sonnet`.
    - **Record each one** adds that a landed fix also calls `flow self-review finding -change <name> -angle flow-fix -disposition fixed -ref <sha> -blast-radius <N>`. A deferred one gets no row until the pass resolves it.
    - **Fewest operator actions**' `flow-high` no-progress escalation is unchanged.
  - [x] **Step 3: Acknowledge every reworded normative sentence in `spectre/changes/kan-927-in-run-self-review-fixes/verbatim-moves.txt`. Verify: `scripts/check-references.sh && scripts/check-verbatim-moves.sh && scripts/check-markdown-integrity.py && scripts/check-dispatch-paragraphs.sh && scripts/check-normative-inventory.sh && scripts/check-installed-citations.sh`.**
**Build:** green
**Files:** `skills/flow-self-review/SKILL.md`, `skills/flow-contracts/pipeline.md`
**Allowed-collateral:** `skills/flow-self-review/SKILL-rationale.md`, `skills/flow-contracts/pipeline-rationale.md`, `commands-claude/flow-self-review.md`
**Tests:** none — prose; the guards in Step 3 check it
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow-self-review): run the fix loop on medium effort with targeted checks, and store in-run fixes
**After:** none

**Decision:** cheap-dispatch-pairs

**Decision:** targeted-verify

**Decision:** every-fix-a-row

Correction (2026-10-08): the plan declared `spectre/changes/kan-927-in-run-self-review-fixes/verbatim-moves.txt` in `**Files:**`; a task commit never stages `spectre/changes/` (FLOW — COMMIT-PER-TASK), so the acknowledgement file rides the planning commits and is dropped from `**Files:**`.

- [x] 4. Run the pass inside `/flow`'s integrate and `/flow-fast`'s verify
  - [x] **Step 1: `skills/flow-contracts/finish-contract-run1.md` — **Save the self-review context bundle** becomes **Run the self-review pass**:**
    - After the archive commit and before any route, the session assembles the bundle (`flow self-review bundle -change <name>` plus `## Session narrative`) without committing it.
    - It runs `/flow-self-review`'s steps 2–6 (`skills/flow-self-review/SKILL.md`, canonical) inline: the six angles, the fix loop on the cheap pairs, the filing-and-rating ask, and the store rows.
    - It commits `docs/self-review/<name>-self-review.md` on `spectre/<name>` through `land-self-review-report.sh` (no `--push`; the route pushes).
    - The fixes land per `fixes-ride-the-change`.
    - "Self-review is always deferred" and the context-bundle commit are removed. A bundle committed by an older run is still consumed by `/flow-self-review` (fallback).
  - [x] **Step 2: `skills/flow/integrate.md` **4.** — the `flow.self-review` stage block runs that pass instead of writing and committing `<name>-context.md`. Its heading names the pass.**
  - [x] **Step 3: `skills/flow-fast/SKILL.md` **5. Verify** — "Self-review is always deferred" becomes the same in-run pass. The report, not the bundle, is committed and pushed with `land-self-review-report.sh … --push "<name>"`.**
  - [x] **Step 4: Update every other statement the move makes false — `grep -rn "always deferred\|self-review context bundle" README.md skills/ commands-claude/` (leaving the panel's own unrelated "context bundle") — including `skills/flow-contracts/finish-contract-run2.md`, `skills/flow/cleanup.md` and `skills/flow-contracts/artifacts-registry.md`. Acknowledge reworded normative sentences in `verbatim-moves.txt`. Verify: `scripts/check-references.sh && scripts/check-verbatim-moves.sh && scripts/check-markdown-integrity.py && scripts/check-stage-mark-calls.sh && scripts/check-guard-symlinks.sh && scripts/check-normative-inventory.sh && scripts/check-installed-citations.sh && scripts/check-self-review-report.sh`.**
**Build:** green
**Files:** `skills/flow-contracts/finish-contract-run1.md`, `skills/flow/integrate.md`, `skills/flow-fast/SKILL.md`, `skills/flow-contracts/git-boundaries.md`, `skills/flow-contracts/SKILL.md`, `stats/internal/stages/names.go`
**Allowed-collateral:** `skills/flow-contracts/finish-contract-run2.md`, `skills/flow-contracts/finish-contract-rationale.md`, `skills/flow-contracts/artifacts-registry.md`, `skills/flow/cleanup.md`, `skills/flow-self-review/SKILL.md`, `commands-claude/*.md`, `README.md`, `skills/README.md`, `AGENTS.md`, `CLAUDE.md`, `templates/*.md`, `skills/flow/scripts/*.sh`, `skills/flow-fast/scripts/*.sh`
**Tests:** none — prose; the guards in Step 4 check it
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): run the self-review pass inside integrate and flow-fast, landing its fixes with the change
**After:** Task 3

**Decision:** in-run-pass-at-integrate

**Decision:** fixes-ride-the-change

**Decision:** filing-ask-stays

Correction (2026-10-08): the plan declared `spectre/changes/kan-927-in-run-self-review-fixes/verbatim-moves.txt` in `**Files:**`; a task commit never stages `spectre/changes/` (FLOW — COMMIT-PER-TASK), so the acknowledgement file rides the planning commits and is dropped from `**Files:**`.

Correction (2026-10-08): the commit also touches `skills/flow-contracts/git-boundaries.md` and `skills/flow-contracts/SKILL.md`, which both stated that run 1 commits the context bundle, and `stats/internal/stages/names.go`, whose `flow.self-review` display name `README.md` mirrors. The archived re-run base now keys on the report commit's subject, and the report's `**Deferred:**` line names `the in-run bundle`.
