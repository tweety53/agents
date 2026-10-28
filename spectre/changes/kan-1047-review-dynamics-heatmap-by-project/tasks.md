# kan-1047-review-dynamics-heatmap-by-project

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, split by source
> set, counted by a word-boundary-safe counter, with any `measured:` command naming no ref of its own.

**Goal:** Projects, Task size and Reviewers bars on Review dynamics, four-week default (`proposal.md`).

**Spec:** `design.md`; prototype `.superpowers/sdd/heatmap-by-project-prototype.html`.

## Global constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/kan-1047-review-dynamics-heatmap-by-project`; never the main checkout.
- `cd stats && gofmt -l .` prints nothing; `cd stats && make -s web-dist && go vet ./...` and `cd stats/web && npx tsc -b` exit 0.

- [ ] 1. Review dynamics filters by project, task size and current reviewers
  - [ ] **Step 1: Store.** `ReviewDynamics(ctx, period, project, class *string)`: `Projects` (distinct reviewed project keys in the period), `classFilterSQL` on every query as `$5`, implementer outcomes moved to `$6`.
  - [ ] **Step 2: API.** `class` in `statsQueryParams`; 400 off review-dynamics or outside micro/small/regular/big; `projects` on the DTO.
  - [ ] **Step 3: SPA.** `ViewProps.onProjectChange` wired from App; `StatsViewParams.class` sent and in `useStatsView`'s deps; Projects, Task size and Reviewers bars; `CURRENT_SLOTS` filter; default preset `"4"`.
  - [ ] **Step 4: Tests.** `TestReviewDynamicsClassAndProjects`, the class cases in `TestStatsReviewDynamics`, the vitest `review dynamics filters by project, task size and retired reviewers`; existing fixtures move to current slot names and the 28-day default.
  - [ ] **Step 5: Verify.** `cd stats && go test ./internal/store ./internal/api -run 'TestReviewDynamics|TestStatsReviewDynamics' -count=1`; `cd stats/web && npx vitest run -t 'review dynamics'`; the global-constraint lint lines. Commit.

**Build:** green
**Files:** `stats/internal/store/reviewdynamics.go`, `stats/internal/store/reviewdynamics_test.go`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/internal/api/promptchanges.go`, `stats/internal/web/embed_test.go`, `stats/internal/client/client_test.go`, `stats/web/src/App.tsx`, `stats/web/src/api.ts`, `stats/web/src/viewTypes.ts`, `stats/web/src/hooks/useStatsView.ts`, `stats/web/src/views/ReviewDynamics.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** `TestReviewDynamicsClassAndProjects`
**Regression:** `TestReviewDynamicsClassAndProjects` fails — a class no longer narrows the changes, or `Projects` is empty.
**Baseline:** before=3 after=4 <!-- measured: grep -c -E '^func Test' stats/internal/store/reviewdynamics_test.go @ origin/main -->
**Commit:** feat(stats): filter review dynamics by project, task size and current reviewers
**After:** none
