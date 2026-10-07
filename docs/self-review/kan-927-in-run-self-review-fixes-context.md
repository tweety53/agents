# Self-review context bundle for kan-927-in-run-self-review-fixes

found: 6 of 7 sources; skipped: 1 of 7 sources
skipped: change summary (absent)

## .superpowers/sdd/ledgers/kan-927-in-run-self-review-fixes.md

# SDD ledger — kan-927-in-run-self-review-fixes

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — implementer

- Task: 1
- Role: implementer
- Key: task-1-implementer
- Model: claude-opus-5-5 effort=default
- Commit: 81a2ca7c
- Outcome: completed
- Started: 2026-10-07T21:24:44Z
- Tokens: not measured

## Dispatch 2 — implementer

- Task: 2
- Role: implementer
- Key: task-2-implementer
- Model: claude-opus-5-5 effort=default
- Commit: e99d560f
- Outcome: completed
- Started: 2026-10-07T21:25:02Z
- Tokens: not measured

## Dispatch 3 — implementer

- Task: 3
- Role: implementer
- Key: task-3-implementer
- Model: claude-opus-5-5 effort=default
- Commit: 70c13523
- Outcome: completed
- Started: 2026-10-07T21:26:34Z
- Tokens: not measured

## Dispatch 4 — implementer

- Task: 4
- Role: implementer
- Key: task-4-implementer
- Model: claude-opus-5-5 effort=default
- Commit: 0eab9831
- Outcome: completed
- Started: 2026-10-07T21:28:27Z
- Tokens: not measured

## Dispatch 5 — reviewer

- Task: no task
- Role: reviewer
- Key: task-1+2+3+4-reviewer
- Model: opus effort=medium
- Commit: no commit
- Outcome: fix
- Started: 2026-10-07T21:32:18Z
- Tokens: input 122, output 523, cache read 4914652, cache creation 143118

## Dispatch 6 — implementer

- Task: no task
- Role: implementer
- Key: task-3+4-implementer-fix-1
- Model: claude-opus-5-5 effort=default
- Commit: 8354648e
- Outcome: completed
- Started: 2026-10-07T21:40:34Z
- Tokens: not measured

## Dispatch 7 — reviewer

- Task: no task
- Role: reviewer
- Key: task-3+4-reviewer-fix-1
- Model: sonnet effort=medium
- Commit: no commit
- Outcome: fix
- Started: 2026-10-07T21:43:41Z
- Tokens: input 26, output 151, cache read 419798, cache creation 83777

## Dispatch 8 — implementer

- Task: no task
- Role: implementer
- Key: task-3+4-implementer-fix-2
- Model: claude-opus-5-5 effort=default
- Commit: e8ee4c76
- Outcome: completed
- Started: 2026-10-07T21:46:29Z
- Tokens: not measured

## Dispatch 9 — reviewer

- Task: no task
- Role: reviewer
- Key: task-3+4-reviewer-fix-2
- Model: sonnet effort=medium
- Commit: no commit
- Outcome: clean
- Started: 2026-10-07T21:47:05Z
- Tokens: input 102, output 567, cache read 5295108, cache creation 146015

## Dispatch 10 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: opus effort=medium
- Commit: no commit
- Outcome: completed
- Started: 2026-10-07T21:48:36Z
- Tokens: not measured

## Dispatch 11 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-0
- Model: claude-opus-5-5 effort=default
- Commit: 2f02f4c5
- Outcome: completed
- Started: 2026-10-07T21:54:46Z
- Tokens: not measured

## Dispatch 12 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-1-primary+principles
- Model: sonnet effort=low
- Commit: no commit
- Diff base: 664865d6
- Outcome: completed
- Started: 2026-10-07T21:59:08Z
- Tokens: input 10, output 29, cache read 114528, cache creation 33410

## Dispatch 13 — verifier

- Task: no task
- Role: verifier
- Key: verify
- Model: claude-opus-5-5 effort=default
- Commit: no commit
- Outcome: completed
- Started: 2026-10-07T22:00:26Z
- Tokens: not measured

## Dispatch 14 — implementer

- Task: no task
- Role: implementer
- Key: pipeline-fix-1
- Model: opus effort=high
- Commit: c8979d5d
- Outcome: completed
- Started: 2026-10-07T22:04:41Z
- Tokens: input 76, output 713, cache read 1871708, cache creation 81168

## Dispatch 15 — verifier

- Task: no task
- Role: verifier
- Key: visual-verify
- Model: opus effort=low
- Commit: no commit
- Outcome: completed
- Started: 2026-10-07T22:09:16Z
- Tokens: input 58, output 962, cache read 1519163, cache creation 100991
## .superpowers/sdd/reviews/kan-927-in-run-self-review-fixes-panel.md

# Review panel — kan-927-in-run-self-review-fixes

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Important | skills/flow-contracts/finish-contract-run1.md:306 | an archived re-run with new work runs the whole self-review pass again, recording every filed and declined finding a second time and re-asking the filing question; /flow-fast skips the pass instead |   |
| F2 | primary | Minor | skills/flow-contracts/finish-contract-run1.md:118 | the archived re-run base no longer recognises the old self-review context bundle subject, which git.go reservedShapes still honours |   |
| F3 | primary | Minor | spectre/changes/kan-927-in-run-self-review-fixes/design.md:78 | the live-check curl uses date-only from/to, which the API rejects as not RFC 3339 |   |
| F4 | principles | Minor | stats/internal/api/stats.go:430 | the no-token-measurement, no-model rule lives in isHealthView and again as an inline self-review special case |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 none — behaviour correct and tested; DRY only

## Pass log

### Round 0

- reachability: still reproduces — finish-contract-run1.md "Self-review is always deferred: no reasoning pass" (integrate saves a bundle only); stats/web has no self_review reference; pipeline.md in-run fixes recorded in narrative only; store: 1 fixed row of 92
- auto-resolved: where does the in-run pass run? → integrate run 1, after the archive commit, in place of saving the bundle (recommended)
- auto-resolved: where do self-review fixes land? → on the change branch when the change is in the agents repo, else on an agents worktree branch landed by its default route (recommended)
- auto-resolved: dispatch pairs for fixes? → fixer flow-medium opus (sonnet for trivial), first review flow-medium opus, re-reviews flow-medium sonnet (recommended; global rule keeps first-pass review on opus)
- auto-resolved: verify scope? → the ## lint lines the touched files need plus targeted tests; no full suite (recommended)
- auto-resolved: /flow-fast too? → yes, same in-run pass at its 5. Verify (recommended)
- auto-resolved: UI shape? → new stats view self-review: one row per finding, outcome filter, commit link from the daemon checkout origin, per-change counts derived client-side (recommended)
- auto-resolved: convergence confirm → approve the design and move on (decisions: recommended)
- auto-resolved: plan review gate → Yes (decisions: recommended)
- roster: compact — 23
- diff size 1212 lines, under cap; docs-only exit 1 (scripts/check-cleanup-complete.sh); roster dispatched: primary+principles (opus/medium); no operator-named addition this round — the resolved list ran alone; standards: CLAUDE.md, AGENTS.md
- F1-F3 reproducers re-authored (premises cited lines the fix changed); prove-reproducer.sh held both legs for each
fix-mutation: stats/internal/api/health.go — dropped viewSelfReview from isHealthView — TestHealthViewsAreNeverUnmeasured/self-review
fix-mutation: skills/flow-contracts/finish-contract-run1.md — none — prose rule, checked by the reproducers and guards
fix-mutations-total: 2

### Round 1

- primary+principles re-run on fix-round-1.diff (rerun pair sonnet/low): F1-F4 verified fixed, no new finding
## spectre/changes/archive/kan-927-in-run-self-review-fixes/tasks.md

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
## spectre/changes/archive/kan-927-in-run-self-review-fixes/design.md

## Context

- `finish-contract-run1.md` **Save the self-review context bundle**: "Self-review is always deferred: no reasoning pass and no prompt run here." Integrate commits `docs/self-review/<name>-context.md`; `/flow-fast` 5. Verify does the same.
- `/flow-self-review` (KAN-875) fixes every non-`big` finding, but only when someone runs it by hand on the default branch. About 16 bundles are pending, and none has been processed since auto-fix shipped.
- Every fix, review and re-review it dispatches runs on `opus` at `flow-high`. So do the in-run pipeline fixes (`pipeline.md` **Pipeline defects found mid-run**, KAN-829).
- In-run pipeline fixes are recorded only as a narrative line. `self_review_findings` holds 92 rows, 1 of them `fixed`.
- The store serves `GET /api/v1/self-review/{project}/{change}/findings` (one change). `stats/web` has no view of it.
- The operator's intent (KAN-927): every non-`big` finding, from any angle, is fixed during the run, fast and cheap. Every fix is stored, and a UI view shows them.

## Decisions

### The pass runs in the integrate run

**ID:** in-run-pass-at-integrate
**Status:** active
**Chosen:** integrate run 1 runs the six-angle pass inline after the archive commit and before the route, on the integrate session's model, over the same bundle `flow self-review bundle` assembles (held in memory, not committed). `/flow-fast` runs the same pass at **5. Verify**. The report `docs/self-review/<name>-self-review.md` is committed on the change branch and lands with the change. `/flow-self-review <name>` stays as the fallback, for a bundle an older run saved.
**Considered:** at the end of implementation, before the `IN_PROGRESS` handoff — the archive, narrative and full ledger don't exist yet, and the pass would run again on every fix run; keep the pass standalone and only auto-invoke it — it is still a separate session on the default branch, so the fixes still don't land with the change.

### Where self-review fixes land

**ID:** fixes-ride-the-change
**Status:** active
**Chosen:** a change whose canonical repository is `<agents repo>` takes the fix commits on `spectre/<name>`, after the archive commit, so they ride the run's own route. Any other project's change puts them on `self-review-<name>` in an `<agents repo>` worktree, landed by `<agents repo>`'s `## default landing route` as **Pipeline defects found mid-run** step 4 does.
**Considered:** always a separate `<agents repo>` branch — a second merge to `main` for every agents change, which the operator rejected ("the fixes land with the change"); always the change branch — a gymie branch cannot carry `<agents repo>` files.

### Cheap dispatch pairs

**ID:** cheap-dispatch-pairs
**Status:** active
**Chosen:**
- One fixer for the whole batch of non-`big` findings: `flow-medium`, `opus` by default, `sonnet` when every fix is a literal edit with nothing it can break (the reason recorded with the dispatch).
- The first review: `flow-medium` `opus`.
- Every re-review after a fix: `flow-medium` `sonnet`.
- The same pairs apply to in-run pipeline fixes (**Pipeline defects found mid-run**).
- The no-progress escalation in **Fewest operator actions** stays `flow-high`.
**Considered:** `sonnet` for the first review too — the operator's message allows it, but the global model rule keeps every first-pass review on `opus`; `flow-high` everywhere (today) — the cost the operator objected to.

### Targeted verification

**ID:** targeted-verify
**Status:** active
**Chosen:** after a clean review, the session runs only the `## lint` lines the touched files need, plus the tests covering them (`go test` on the touched packages with `-run` where a test is named, the `scripts/test-*.sh` harness of a touched script, `npx vitest run <file>` for a touched SPA file). No full suite. A red line loops back through the fixer, as today.
**Considered:** the full `## lint` and `## test` list (today) — about 4 minutes per pass, the cost the operator objected to.
<!-- measured: kan-916 self-review verify, 2026-10-07 — run-guard-tests.sh "57s wall", store package alone 94.9s, full lint list before them; machine-local timing, not re-runnable at a ref -->

### Every fix is a store row

**ID:** every-fix-a-row
**Status:** active
**Chosen:**
- An in-run pipeline fix, landed or deferred, calls `flow self-review finding -change <name> -angle flow-fix -disposition fixed -ref <sha> -blast-radius <N>` when landed. A deferred one gets no row until the pass decides it.
- The pass records every finding as today.
- The narrative line stays as a copy.
- No new disposition and no schema change.
**Considered:** a `deferred` disposition — the pass already resolves a deferred fix to `fixed`/`filed`/`declined` in the same run.

### The filing ask and the rating stay

**ID:** filing-ask-stays
**Status:** active
**Chosen:** the pass asks the one combined filing-and-rating question at integrate (shape per today's step 4). Silence files nothing. A bare `/flow` invocation means an operator is present.
**Considered:** filing `big` findings without asking — a Jira create is outward-facing, and the operator wants findings explained first.

### Self-review fixes view

**ID:** self-review-view
**Status:** active
**Chosen:**
- A new stats view `self-review` ("Self-review fixes"), served as `GET /api/v1/stats/self-review` with the existing period/project filters. One row per finding: recorded at, change, angle, finding, outcome, ref.
- `fixed` renders the ref as a commit link built from the daemon checkout's `origin` remote (GitHub https form, `commitUrl` in the DTO), or as plain text when the remote is not GitHub.
- The DataTable's column filter on outcome serves as the outcome filter.
- A second table holds per-change fixed/filed/declined counts, derived client-side from the same rows.
**Considered:** a section inside RunDetail — per change only, with no cross-change view; a Jira link for `filed` — the daemon knows no Jira base URL.

## Live check

- flowd on `127.0.0.1:4173` (the dev workspace's, never restarted by the run; the operator reloads it):
  - `curl -s 'http://127.0.0.1:4173/api/v1/stats/self-review?from=2026-10-01T00:00:00Z&to=2026-10-09T00:00:00Z'` returns the 96 existing rows (92 plus kan-916's 4).
  - kan-916's two `fixed` rows carry a `commitUrl` under `github.com/tweety53/agents/commit/`.
- The SPA at that address shows the "Self-review fixes" tab with those rows, and the per-change table shows kan-916 with 2/2/0.
- Failure looks like: 404 on the route, an empty table while the DB holds rows, or `fixed` rows with no link.

## Open questions
## spectre/changes/archive/kan-927-in-run-self-review-fixes/narrative.md

# kan-927-in-run-self-review-fixes — session narrative

## 2026-10-08 — creating run

Resumed at `STARTED` with a ready four-task plan; `small` class, inline execution, one bundled `primary+principles` panel.

- **Plan corrections at task close.** Three tasks' `**Files:**` fields were short: task 1 also had to teach two more `api.StatsStore` fakes (`internal/web/embed_test.go`, `internal/client/client_test.go`) the new method; task 2 had to move `RunDetail.test.tsx`'s pinned `VIEW_NAMES` count from 8 to 9; task 4 also touched `git-boundaries.md`, `flow-contracts/SKILL.md` and `stages/names.go`. Tasks 3 and 4 also declared `verbatim-moves.txt`, a `spectre/changes/` path no task commit may carry — dropped from `**Files:**`; it rides the planning commits.
- **Stale fix shas — three gated-review rounds.** The first design recorded each change-branch fix's sha at the pass. The reviewer showed any later rebase (`/flow-fast`'s §7 rebase, run 1's rejected-push re-sync) leaves those shas outside history. First attempt — re-read shas and amend the report before every push, record rows after "the push" — was rejected: "the push" was ambiguous on merge-and-push, the raw amend bypassed the landing script's guards, and on `/flow-fast` the in-memory findings were gone by the time a later invocation pushed. Final shape: the committed report is the record; the landing push (named per route), in whichever invocation makes it, re-reads shas by subject, commits a rewritten report through `land-self-review-report.sh`, then records `fixed` rows from the report's lines. A squash/rebase PR merge is a stated ceiling.
- **Re-run duplicates.** `/flow-fast` fix re-runs and `/flow`'s archived re-runs both re-ran the pass, duplicating store rows and re-asking the filing question; both now skip the pass once a report is committed. New work on such a re-run gets the review panel, not a second pass.
- **Panel round 0** raised 1 Important (the archived re-run above) and 3 Minors; all fixed inline, re-run clean. Reproducers 1–3 had to be re-authored: their premises cited the very lines the fix changed.
- **Full Go suite, second run:** `reconcile.TestConcurrentAppendVersusRetirePreservesEveryEntry` lost 1 of 600 entries under full-suite load (15/15 green in isolation). The in-run pipeline fix found the cause — `AppendJournalEntry` writes unlocked after a 50ms lock wait, and a macOS `F_FULLFSYNC` retire outlasts it — and fixed it on `fix-journal-retire-race` (c8979d5d) by waiting for the lock unbounded. That reverses a recorded design choice (bounded wait, loss acceptable), so it is deferred to the operator rather than landed.
- **Live check** ran against a worktree flowd on 4693 (its database had to be created with `scripts/workspace.sh create` and the project row seeded; the first start failed on the missing database). The worktree flowd replays the shared state-dir journals by design; both journals on disk were empty, so nothing was consumed. The 4173 figures in design.md wait for the operator's reload.
- **Visual verify** captured only the empty view: the UI-test seed has no `agents-a740d89c` project row, so seeding findings there failed the FK. The verifier also made a test-only edit (`exact: true` on the full-suite nav locator) outside its relay contract.

In-run pipeline fix: deferred — fallback journal append writes unlocked after a 50ms wait and loses entries to a concurrent retire (blast radius 5 files)

## 2026-10-08 — integrate run

- **Preflight:** `RUN1`; main checkout staged-clean and drift-clean; no retired-layout worktree to migrate.
- **Unfinished-work gate:** `CLEAR` — every plan item checked, no open finding; visual verify dispatched (`VISUAL-VERIFY-OK`).
- **Base:** `origin/main` had not moved since the recorded merge base `3b4aa4dc`; no rebase.
- **Route:** merge and push, from the project's configured default.
- **Self-review:** run on the installed (pre-change) integrate, so the pass is deferred to a saved bundle; this change's in-run pass applies from the next integrate on.
## git log --stat

commit e62c2d687baf86a4a542153498d5f93307dc3bdd
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Oct 8 01:59:43 2026 +0300

    feat(flow): run the self-review pass inside integrate and flow-fast, fix findings cheaply, and show them in a stats view

 AGENTS.md                                          |   2 +-
 CLAUDE.md                                          |   2 +-
 README.md                                          |   4 +-
 commands-claude/flow-fast.md                       |   4 +-
 commands-claude/flow-self-review.md                |   6 +-
 scripts/check-cleanup-complete.sh                  |   5 +-
 scripts/check-self-review-report.sh                |   7 +-
 skills/README.md                                   |   2 +-
 skills/flow-contracts/SKILL.md                     |   2 +-
 skills/flow-contracts/artifacts-registry.md        |   2 +-
 skills/flow-contracts/finish-contract-rationale.md |   4 +-
 skills/flow-contracts/finish-contract-run1.md      |  52 +++++++----
 skills/flow-contracts/finish-contract-run2.md      |   2 +-
 skills/flow-contracts/git-boundaries.md            |   2 +-
 skills/flow-contracts/pipeline.md                  |  15 +++-
 skills/flow-fast/SKILL.md                          |  36 ++++----
 skills/flow-self-review/SKILL.md                   |  90 +++++++++++++------
 skills/flow/cleanup.md                             |   4 +-
 skills/flow/integrate.md                           |  15 ++--
 stats/cmd/flowd/main.go                            |  32 ++++++-
 stats/cmd/flowd/wiring_test.go                     |  22 +++++
 stats/internal/api/health.go                       |  19 ++--
 stats/internal/api/health_test.go                  |   4 +-
 stats/internal/api/server.go                       |  17 +++-
 stats/internal/api/stats.go                        |  45 +++++++++-
 stats/internal/api/stats_test.go                   |  97 ++++++++++++++++++++
 stats/internal/client/client_test.go               |   4 +
 stats/internal/selfreview/bundle_test.go           |  52 ++++++-----
 stats/internal/selfreview/git.go                   |  22 ++---
 stats/internal/stages/names.go                     |   2 +-
 stats/internal/store/selfreviewfindings.go         |  36 ++++++++
 stats/internal/store/selfreviewfindings_test.go    |  77 ++++++++++++++++
 stats/internal/web/embed_test.go                   |   4 +
 stats/web/src/App.tsx                              |   3 +
 stats/web/src/api.ts                               |  18 +++-
 stats/web/src/views/RunDetail.test.tsx             |   4 +-
 stats/web/src/views/SelfReview.tsx                 | 100 +++++++++++++++++++++
 stats/web/src/views/views.test.tsx                 |  38 ++++++++
 stats/web/tests/visual/full-app-suite.spec.ts      |   3 +-
 .../full-cache-efficiency-darwin.png               | Bin 80178 -> 81225 bytes
 .../full-decisions-darwin.png                      | Bin 74831 -> 75857 bytes
 .../full-flow-health-darwin.png                    | Bin 139282 -> 140400 bytes
 .../full-reviewers-darwin.png                      | Bin 66692 -> 67659 bytes
 .../full-run-detail-darwin.png                     | Bin 93766 -> 94823 bytes
 .../full-runs-darwin.png                           | Bin 108174 -> 109317 bytes
 .../full-self-review-darwin.png                    | Bin 0 -> 78631 bytes
 .../full-stage-leaderboard-darwin.png              | Bin 74823 -> 75873 bytes
 .../full-state-board-darwin.png                    | Bin 99137 -> 100277 bytes
 .../full-trend-darwin.png                          | Bin 73108 -> 74244 bytes
 stats/web/tests/visual/full-app-suite.zip          | Bin 747326 -> 829021 bytes
 stats/web/tests/visual/self-review.spec.ts         |  22 +++++
 .../self-review-empty-darwin.png                   | Bin 0 -> 78631 bytes
 52 files changed, 734 insertions(+), 143 deletions(-)

commit d071736513659794e7428308cd34b40ce1cb416e
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Oct 8 01:59:43 2026 +0300

    chore(spectre): plan kan-927-in-run-self-review-fixes

 .../kan-927-in-run-self-review-fixes/design.md     |  83 ++++++++++++
 .../live-verification.md                           |  15 +++
 .../kan-927-in-run-self-review-fixes/narrative.md  |  23 ++++
 .../kan-927-in-run-self-review-fixes/proposal.md   |  18 +++
 .../kan-927-in-run-self-review-fixes/tasks.md      | 126 ++++++++++++++++++
 .../verbatim-moves.txt                             | 144 +++++++++++++++++++++
 .../visual-verification.md                         |  42 ++++++
 7 files changed, 451 insertions(+)

commit c008a399a7ac22e3df40976a50bfc61bcf20e57f
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Thu Oct 8 01:59:50 2026 +0300

    chore(spectre): archive kan-927-in-run-self-review-fixes

 .../kan-927-in-run-self-review-fixes/design.md     |   0
 .../kan-927-in-run-self-review-fixes/ledger.md     | 172 +++++++++++++++++++++
 .../live-verification.md                           |   0
 .../kan-927-in-run-self-review-fixes/narrative.md  |   0
 .../kan-927-in-run-self-review-fixes/panel.md      |  46 ++++++
 .../kan-927-in-run-self-review-fixes/proposal.md   |   0
 .../kan-927-in-run-self-review-fixes/tasks.md      |   0
 .../verbatim-moves.txt                             |   0
 .../visual-verification.md                         |   0
 9 files changed, 218 insertions(+)

## Session narrative

Integrate run 1 on 2026-10-08: preflight `RUN1`, main checkout clean, unfinished-work gate `CLEAR` with visual verify recorded, `origin/main` unmoved since `3b4aa4dc` so no rebase. The branch was reshaped into one implementation commit and one planning commit, archived, and landed by the project's default merge-and-push route. This run executed the installed integrate procedure, which predates this change, so the self-review pass is deferred to this bundle rather than run in-run.
