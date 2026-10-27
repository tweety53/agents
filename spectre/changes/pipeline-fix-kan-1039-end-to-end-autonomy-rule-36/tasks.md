# pipeline-fix-kan-1039-end-to-end-autonomy-rule-36

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

**Goal:** `flow record dispatch begin` accepts a change whose worktree is on branch
`spectre/<change>` before `spectre new` has run.

**Architecture:** one extra match in `changeCheckouts` (`stats/cmd/flow/record.go`) plus one Go test.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- Read-only against the dev store: never stop flowd or flow-postgres.
- `recordEfforts` stays unchanged (`xhigh-effort-out-of-scope`).

## Review Focus

- The branch match compares the full ref `refs/heads/spectre/<change>`, never a suffix.
- `gatePanelFindingsClosed` and the inline-row resolution still pick the change's own worktree.

- [ ] 1. Accept a worktree on branch spectre/<change> as carrying the change
  - [ ] **Step 1: Write the failing test.** In `stats/cmd/flow/record_test.go` add
    `TestRequireChangeCheckoutAcceptsTheChangeBranch`: `repo := gitRepo(t)`, commit once, then
    `git -C repo worktree add -b spectre/kan-9 <t.TempDir()>/wt`; assert
    `requireChangeCheckout(repo, "kan-9", io.Discard)` is true and
    `requireChangeCheckout(repo, "kan-10", io.Discard)` is false. Run
    `cd stats && go test ./cmd/flow -run 'TestRequireChangeCheckoutAcceptsTheChangeBranch'`: it fails.
  - [ ] **Step 2: Implement.** In `changeCheckouts`, track the current entry's worktree and, on a
    `branch refs/heads/spectre/<change>` line, append that worktree to `carrying` unless already
    appended by the directory check. Update `requireChangeCheckout`'s and `changeCheckouts`'
    comments and the refusal message to name the `spectre/<change>` branch.
  - [ ] **Step 3: Verify.** `cd stats && gofmt -w . && gofmt -l .` prints nothing;
    `cd stats && make -s web-dist && go vet ./...` exits 0;
    `cd stats && go test ./cmd/flow -run 'TestRequireChangeCheckout'` passes.
  - [ ] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `stats/cmd/flow/record_test.go` (`TestRequireChangeCheckoutAcceptsTheChangeBranch`)
**Regression:** TestRequireChangeCheckoutAcceptsTheChangeBranch — reverting the branch match fails it with "a worktree on spectre/kan-9 refused".
**Commit:** fix(flow-cli): accept a worktree on the change's branch as carrying the change
**After:** none

**Decision:** branch-carries-change
**Decision:** no-retroactive-repair
