# pipeline-fix-kan-1025-flow-improvement-require-a-suite-regression-test-33

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

- [x] 1. Review-panel end resolves the change's worktree when -C is absent
  - [x] **Step 1: In `stats/cmd/flow/stage_test.go`, add `TestStageEndReviewPanelResolvesWorktreeWithoutC`: build a repo with `gitRepo(t)`, add a linked worktree (`git -C <repo> worktree add <wt>`), create `<wt>/.superpowers/sdd/kan-1/` with every closed-finding fixture the existing `TestStageEndReviewPanelRunsFindingsClosed` "every finding closed ends" case uses, have its fake `flow` stub report one completed reviewer dispatch whose `panel-report-0-primary.md` exists only under `<wt>/.superpowers/sdd/kan-1/`, `t.Chdir(<repo>)`, and run `stage end` with no `-C`. Assert exit 0 and the daemon reached.**
  - [x] **Step 2: Run `cd stats && go test ./cmd/flow -run TestStageEndReviewPanelResolvesWorktreeWithoutC` and confirm it fails (the guard reads the main checkout and reports the report file missing).**
  - [x] **Step 3: In `stats/cmd/flow/stage.go` `gatePanelFindingsClosed`, when `f.dir == ""`, call `changeCheckouts(".", f.name)` and use the first carrying checkout other than `main`, else `main` if it carries the change, else `"."`; pass that dir to the guard and as `guard.Env.Dir`. Update the doc comment to name kan-1025.**
  - [x] **Step 4: Verify: `cd stats && gofmt -w ./cmd/flow && go vet ./cmd/flow && go test ./cmd/flow -run 'TestStageEndReviewPanel'` exits 0.**
**Build:** green
**Files:** `stats/cmd/flow/stage.go`, `stats/cmd/flow/stage_test.go`
**Tests:** `TestStageEndReviewPanelResolvesWorktreeWithoutC`, `TestStageEndReviewPanelRunsFindingsClosed`
**Regression:** reverting the commit makes `TestStageEndReviewPanelResolvesWorktreeWithoutC` fail: the guard reads the main checkout's `.superpowers/sdd/` and refuses.
**Commit:** fix(flow): resolve the change's worktree for the review-panel end gate
**After:** none

**Decision:** panel-gate-resolves-worktree
