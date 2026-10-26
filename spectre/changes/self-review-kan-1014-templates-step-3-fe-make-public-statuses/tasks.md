# self-review-kan-1014-templates-step-3-fe-make-public-statuses

> **Execution:** `/flow-self-review`'s fixer implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-26-self-review-kan-1014-templates-step-3-fe-make-public-statuses-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` here. Never merge or push.
- Every refusal added exits 2 with a message naming its cause and remedy, and each new test fails with its fix reverted.
- Run `cd stats && gofmt -w .` before committing any Go file; `gofmt -l .` and `go vet ./...` stay clean (`make -s web-dist` first, per `.flow/project.md`).
- Task 3 commits in the gymie worktree `/Users/tweety53/Projects/gymie-worktrees/kan-1014-templates-step-3-fe-make-public-statuses` on its branch `spectre/kan-1014-templates-step-3-fe-make-public-statuses`, never in gymie's main checkout.

## Review focus

- A seven-argument call must record exactly as before, with `-C <worktree>` (Task 1).
- `none — <reason>` and relative reproducers must still be accepted (Task 2).

- [ ] 1. Record run-verify-list's verify row against the change's canonical worktree
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Read `stats/internal/guard/runverifylist.go:117-135`: the recording argv is `record dispatch <verb> -C <wt>`, `wt` the worktree under verify, and `requireChangeCheckout` (`stats/cmd/flow/record.go`) refuses a `-C` repository holding no `spectre/changes/<name>`.
  - [ ] **Step 2: Write the failing test.** In `stats/internal/guard/run_verify_list_test.go` add `TestRunVerifyListRecordsAgainstTheCanonicalWorktree`, in `TestRunVerifyListRecordsTheVerifyRow`'s shape: two scratch repositories, the change directory in the second only; call with eight arguments (the eighth the second repository) and assert every recorded `flow` argv carries `-C <second>`; a seven-argument call still carries `-C <worktree>`; six or nine arguments exit 2. Run `cd stats && go test ./internal/guard/ -run TestRunVerifyListRecordsAgainstTheCanonicalWorktree`: it fails (eight arguments is usage).
  - [ ] **Step 3: Implement.** In `runVerifyList` accept 2, 7 or 8 arguments; the recording `-C` target is `args[7]` when present, else `wt`. Usage text and header comment name `[<canonical-worktree>]`. In `scripts/run-verify-list.sh` the `# Usage:` line becomes `run-verify-list.sh <worktree> lint|test [<change> <key> <session-token> <model> <effort> [<canonical-worktree>]]` and the exit-2 line names "none, all five, or all six". In `skills/flow/verify-and-handoff.md` (the block at lines 153-157) both calls end `[<canonical worktree>]`, the manual end becomes `flow record dispatch end -C <canonical worktree> -change <name> …`, and one sentence after the block: "`<canonical worktree>` is the worktree holding `spectre/changes/<name>`; pass it when verifying any other worktree of a multi-repo change."
  - [ ] **Step 4: Verify.** `cd stats && go test ./internal/guard/ -run 'TestRunVerifyList'` passes; `scripts/check-markdown-integrity.sh` and `scripts/check-references.sh` exit 0; gofmt/vet clean. Commit.

**Build:** green
**Files:** `stats/internal/guard/runverifylist.go`, `stats/internal/guard/run_verify_list_test.go`, `scripts/run-verify-list.sh`, `skills/flow/verify-and-handoff.md`
**Tests:** `TestRunVerifyListRecordsAgainstTheCanonicalWorktree`
**Commit:** `fix(guard): let run-verify-list record against the change's canonical worktree`
**After:** none
**Regression:** without the 8th argument an eight-argument call exits 2 as usage, so the test's first case fails.

- [ ] 2. Refuse an absolute reproducer path at record time
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `validateFindingReproducer` (`stats/cmd/flow/record.go:242-261`) returns nil for `/abs/path/x.sh`, while `panelfindingsclosed.go:280,298` joins worktree+"/"+token and can never resolve it.
  - [ ] **Step 2: Write the failing test.** In `stats/cmd/flow/record_test.go` add `TestRecordFindingRejectsAbsoluteReproducer`, in `TestRecordFindingRejectsBareNoneReproducer`'s shape: `-reproducer '/tmp/x.sh arg'` exits 2 with stderr naming "absolute path" and "relative to the worktree"; `scripts/run-verify-list.sh /abs` and `none — prose only` are accepted. Run `cd stats && go test ./cmd/flow/ -run TestRecordFindingRejectsAbsoluteReproducer`: it fails.
  - [ ] **Step 3: Implement.** In `validateFindingReproducer`, after the `none` branch, when the first field starts with `/` return `fmt.Errorf("-reproducer token '%s' is an absolute path — a reproducer names a path relative to the worktree", fields[0])`, worded as `runreproducer.go:169`. Rewrite the comment's last sentence: an absolute first token is refused here (decision `finding-reproducer-relative`); metacharacters, `..` segments and later tokens stay a guard-side check.
  - [ ] **Step 4: Verify.** `cd stats && go test ./cmd/flow/ -run 'TestRecordFinding'` passes; gofmt/vet clean. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordFindingRejectsAbsoluteReproducer`
**Commit:** `fix(records): refuse an absolute reproducer path in flow record finding`
**After:** none
**Regression:** without the check `/tmp/x.sh arg` records with exit 0, so the test's first case fails wanting 2.

- [ ] 3. State that gymie desktopTest test names are camelCase
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `gymie-frontend/detekt.yml:37-40` sets `FunctionNaming` `functionPattern: '[a-z][a-zA-Z0-9]*'`; detekt's default excludes cover `commonTest` only, so backticked `desktopTest` names fail; gymie's `.flow/project.md` says nothing of it.
  - [ ] **Step 2: Implement.** In `/Users/tweety53/Projects/gymie-worktrees/kan-1014-templates-step-3-fe-make-public-statuses/.flow/project.md`, `## lint`, after the paragraph ending "(KAN-780)." add the paragraph: "**`desktopTest` test functions are camelCase.** gymie-frontend's detekt `FunctionNaming` (`functionPattern '[a-z][a-zA-Z0-9]*'`) applies to `desktopTest` — detekt's default excludes cover `commonTest` only — so a backticked test name fails `verifyChange`."
  - [ ] **Step 3: Verify.** `git -C /Users/tweety53/Projects/gymie-worktrees/kan-1014-templates-step-3-fe-make-public-statuses diff --stat` shows `.flow/project.md` only. Commit there on `spectre/kan-1014-templates-step-3-fe-make-public-statuses`.

**Build:** green
**Files:** `.flow/project.md`
**Tests:** none
**Commit:** `docs(flow): desktopTest test names are camelCase under detekt FunctionNaming`
**After:** none
**Regression:** none — documentation only.
