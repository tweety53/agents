# pipeline-fix-kan-1060-flow-cost-visual-verify-now-costs-more-than-49

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** `origin/main`'s Playwright visual suite is green, with every project named by a Project chip
on the live state board.

**Architecture:** `ui-test-up` copies the model policy; a read-only `GET /api/v1/projects` feeds the
board's Project chips until finished changes load; stale specs and baselines follow intended UI.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- No retry, sleep or wider timeout or screenshot threshold (`fix-determinism-at-the-source`). No new
  inline suppressions.
- Never stop flowd (4173) or flow-postgres (5433), never drop the `flow` DB. The UI-test stack on 4174
  may be started and stopped; `make -C stats ui-test-down` before the turn ends.
- A screenshot is re-recorded only after its diff PNG shows no move beyond those `design.md`'s Context
  table names (`stale-specs-follow-intent`).

## Review Focus

- Chips before the finished load name every registered project with a `—` count, and fall back to
  in-period keys on fetch error (`project-chips-name-every-project`).
- A chip press loads finished changes; nothing else new triggers that load (`70594ba2` intent kept).
- Each re-recorded PNG's change matches a named intended commit (`stale-specs-follow-intent`).

- [x] 1. Let ui-test-up build flowd on a fresh worktree
  - [x] **Step 1: Reproduce.** In a fresh worktree, `rm -f stats/internal/modelpolicy/embedded/model-policy.md && make -C stats ui-test-up`: fails with `pattern model-policy.md: no matching files found`.
  - [x] **Step 2: Implement.** In `stats/Makefile`, change `ui-test-up: web-build check-uitest-overrides` to `ui-test-up: web-build policy-copy check-uitest-overrides`.
  - [x] **Step 3: Verify.** Rerun Step 1's command: the stack comes up; then `make -C stats ui-test-down`. From the repository root every `scripts/check-*.sh` guard named in `.flow/project.md`'s `## lint` exits 0.
  - [x] **Step 4: Commit** with the subject below.

**Build:** green
**Files:** `stats/Makefile`
**Tests:** none — build tooling: `make -C stats ui-test-up` is its own check
**Regression:** `make -C stats ui-test-up` with the policy copy removed — reverting the commit fails it with `pattern model-policy.md: no matching files found`.
**Baseline:** go before=1938 after=1938; web before=310 after=310

**Commit:** fix(stats): copy the model policy before ui-test-up builds flowd
**After:** none

**Decision:** ui-test-up-policy-copy

- [x] 2. Name every project as a Project chip before finished changes load
  - [x] **Step 1: Failing tests first.** Add `TestProjectsRoute` in `stats/internal/api/` asserting `GET /api/v1/projects` returns `{"projects":[...]}` in key order from a store stub's `ProjectRoots`. Add a `views.test.tsx` case "names every project as a chip before finished changes load": with the board's open poll holding only project A and `/api/v1/projects` returning A and B, the `Project` group shows a `B` chip whose count reads `—`, and pressing it issues the finished state-board read and then shows B's finished count. Both fail.
  - [x] **Step 2: Implement.** Route `GET /api/v1/projects` in `stats/internal/api/server.go` to a handler that calls `ProjectRoots` and writes the keys (errors via `mapStoreError`). In `stats/web/src/views/StateBoard.tsx`: fetch it once on mount; while `finishedLoaded` is false, `projects` is the union of the fetched keys and the in-period keys and each chip count renders `—`; once loaded, unchanged. Choosing a project sets `wantFinished`. A failed fetch leaves the in-period keys alone.
  - [x] **Step 3: Verify.** From `stats/`: `make -s web-dist && go vet ./... && gofmt -l .` clean, `go test -run TestProjects ./internal/api/` ok, `cd web && npx tsc -b` clean and `npx vitest run -t "names every project"` ok. With the UI-test stack up, `npx playwright test tests/visual/controls.spec.ts -g "Project chip"` passes. Every `scripts/check-*.sh` guard in `.flow/project.md`'s `## lint` exits 0.
  - [x] **Step 4: Commit** with the subject below.

**Build:** green
**Files:** `stats/internal/api/lessons.go`, `stats/internal/api/projects_test.go`, `stats/internal/api/server.go`, `stats/web/src/api.ts`, `stats/web/src/views/StateBoard.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** the projects-route API test; the `views.test.tsx` chip-before-load case
**Regression:** `controls.spec.ts` "a Project chip narrows the rows and the state tiles" — reverting the commit times it out on the missing `uitest-beta` chip.
**Baseline:** go before=1938 after=1939; web before=310 after=311

**Commit:** fix(stats): name every project as a Project chip before finished changes load
**After:** Task 1

**Decision:** project-chips-name-every-project

- [x] 3. Bring the visual specs and baselines in line with the intended board and run view
  - [x] **Step 1: Specs.** In `stats/web/tests/visual/controls.spec.ts` "Updated All widens…", click the `▾Finished` group row and wait for its count to leave `—` before reading the tile. In `interactive.spec.ts` "interactive run view", pass `exact: true` to `getByRole("region", { name: "Main" })`.
  - [x] **Step 2: Inspect.** With the UI-test stack up, `npx playwright test` from `stats/web`; open every failing `*-diff.png` and confirm each move is one `design.md`'s Context table names. Any other move is fixed in code, not re-recorded.
  - [x] **Step 3: Re-record** only the screenshots of Step 2: `npx playwright test --update-snapshots <spec> -g <title>` per test.
  - [x] **Step 4: Verify.** `npx playwright test` from `stats/web`: all pass, twice. `make -C stats ui-test-down`. Every `scripts/check-*.sh` guard in `.flow/project.md`'s `## lint` exits 0. The commit body lists each re-recorded PNG and the move it carries.
  - [x] **Step 5: Commit** with the subject below.

**Build:** green
**Files:** `stats/web/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png`, `stats/web/tests/visual/baseline.spec.ts-snapshots/run-detail-darwin.png`, `stats/web/tests/visual/baseline.spec.ts-snapshots/run-detail-hover-darwin.png`, `stats/web/tests/visual/combinations.spec.ts-snapshots/combination-chips-plus-tile-darwin.png`, `stats/web/tests/visual/controls.spec.ts`, `stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-run-detail-darwin.png`, `stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-state-board-darwin.png`, `stats/web/tests/visual/interactive.spec.ts`, `stats/web/tests/visual/interactive.spec.ts-snapshots/interactive-run-view-darwin.png`, `stats/web/tests/visual/interactive.spec.ts-snapshots/interactive-running-card-darwin.png`, `stats/web/tests/visual/interactive.spec.ts-snapshots/interactive-running-card-focus-darwin.png`, `stats/web/tests/visual/interactive.spec.ts-snapshots/interactive-running-card-hover-darwin.png`, `stats/web/tests/visual/interactive.spec.ts-snapshots/interactive-waiting-card-darwin.png`
**Tests:** none — repairs existing tests: `controls.spec.ts` "Updated All widens…", `interactive.spec.ts` "interactive run view", and the re-recorded screenshot tests
**Regression:** `npx playwright test` against the UI-test stack — reverting the commit fails the tests named in `design.md`'s Context table.
**Baseline:** go before=1939 after=1939; web before=311 after=311

**Commit:** test(web): follow the intended state board and run view in the visual suite
**After:** Task 2

**Decision:** stale-specs-follow-intent
