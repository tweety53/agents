# pipeline-fix-kan-1009-editor-replace-exercise-remove-from-superset-29

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

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

### Task 1: Bound a bundle's git calls together

**Files:**
- Modify: `stats/internal/selfreview/git.go`
- Modify: `stats/internal/api/selfreview.go`
- Modify: `stats/internal/selfreview/bundle_test.go`
- Modify: `stats/internal/api/selfreview_test.go`

**Commit:** `fix(selfreview): bound a bundle's git calls together inside the write timeout`

- [ ] **Step 1:** Add `BundleBudget` and `ExecRunner.Deadline` to `git.go`; `Output` refuses a call
  past the deadline and caps its bound at the time left.
- [ ] **Step 2:** The bundle handler sets the deadline on an `ExecRunner`.
- [ ] **Step 3:** Add `TestBundleGitBudgetCapsSlowRepository` (20 shimmed repositories, 300ms
  bound, 500ms deadline, must return within 3s; 24s without the deadline check) and re-pin
  `TestGitBoundInsideWriteBudget` on `BundleBudget + DefaultGitBound`.
- [ ] **Step 4:** `cd stats && go test ./... -race -count=1` passes.
