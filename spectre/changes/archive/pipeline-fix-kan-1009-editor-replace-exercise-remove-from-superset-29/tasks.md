# pipeline-fix-kan-1009-editor-replace-exercise-remove-from-superset-29

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter, with any `measured:` command naming no ref of its own.

**Goal:** the self-review bundle route serves a degraded bundle inside `writeTimeout` however slow
the repository's git calls are, instead of an empty reply.

**Architecture:** `selfreview.ExecRunner.Deadline` caps every git call of one bundle; the handler
sets it `selfreview.BundleBudget` out.

**Spec:** `design.md` of this change.

## Global constraints

- No new inline suppressions; `cd stats && gofmt -l .` empty and `go vet ./...` clean.
- Read-only against the dev store: the tests never call flowd or the real git of a real repository.

## Review Focus

- A call starting past the deadline runs no git (`deadline-on-exec-runner`).
- `BundleBudget + DefaultGitBound` stays inside the 30s `writeTimeout` (`budget-15s`).
- Fake runners in tests are untouched: only an `ExecRunner` gets the deadline.

- [x] 1. Bound a bundle's git calls together inside the write timeout
  - [x] **Step 1: Implement.** Add `BundleBudget` and `ExecRunner.Deadline` to
    `stats/internal/selfreview/git.go`; `Output` refuses a call past the deadline and caps its
    bound at the time left. The bundle handler (`stats/internal/api/selfreview.go`) sets the
    deadline `BundleBudget` out on an `ExecRunner`.
  - [x] **Step 2: Test.** Add `TestBundleGitBudgetCapsSlowRepository` (20 shimmed repositories,
    300ms bound, 500ms deadline, must return within 3s) and re-pin `TestGitBoundInsideWriteBudget`
    on `BundleBudget + DefaultGitBound` inside 30s.
  - [x] **Step 3: Verify.** `cd stats && gofmt -l .` prints nothing, `go vet ./...` exits 0, and
    `go test ./internal/selfreview/ -run TestBundleGitBudgetCapsSlowRepository -count=1` and
    `go test ./internal/api/ -run TestGitBoundInsideWriteBudget -count=1` pass.
  - [x] **Step 4: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/internal/selfreview/git.go`, `stats/internal/api/selfreview.go`, `stats/internal/selfreview/bundle_test.go`, `stats/internal/api/selfreview_test.go`
**Tests:** `TestBundleGitBudgetCapsSlowRepository`, `TestGitBoundInsideWriteBudget`
**Regression:** TestBundleGitBudgetCapsSlowRepository — with the deadline check disabled the shimmed bundle takes 24.4s: `Bundle took 24.429116917s past a 500ms git deadline`.
**Baseline:** before=43 after=44
<!-- measured: grep -cE '^func Test' over stats/internal/selfreview/bundle_test.go and stats/internal/api/selfreview_test.go (before at this task's parent → 43; after at its commit → 44) -->
**Commit:** fix(selfreview): bound a bundle's git calls together inside the write timeout
**After:** none

**Decision:** deadline-on-exec-runner
**Decision:** budget-15s
