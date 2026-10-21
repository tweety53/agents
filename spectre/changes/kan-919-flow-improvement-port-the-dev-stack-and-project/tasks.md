# kan-919-flow-improvement-port-the-dev-stack-and-project

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** port `check-dev-stack-fresh`, `check-spec-reach` and `check-uitest-overrides` to Go
and delete the bash libraries only they sourced (KAN-919), as `proposal.md` states.

**Architecture:** one Go file plus one Go table test per guard in `stats/internal/guard/`, the
script reduced to a `flow_guard_exec` shim; then a deletion commit for
`lib/strip-bom.sh`, `lib/sanitize-display.sh` and `lib/visual-table-cells.awk`.

**Tech Stack:** Go (`stats/`), Bash (shims).

**Spec:** `design.md` beside this file.

## Global constraints

- Every path is relative to this worktree, `/Users/tweety53/Projects/agents-worktrees/kan-919-flow-improvement-port-the-dev-stack-and-project`. Never touch the main checkout; never create a worktree.
- The port pattern is c20a7ab2 (`git show c20a7ab2`): copy its shim shape (`scripts/check-mutation-reproducer-pin.sh` — header comment kept and updated to name the Go file, `flow_guard_cwd_checkout`, `flow_guard_exec <name> 2 "<name>:" "$@"`) and its test shape (`stats/internal/guard/check_mutation_reproducer_pin_test.go` — one subtest per harness case label, a comment naming the harness and `b1b0d601`). Register each guard with `Registry["<name>"] = <func>` in an `init`, as `mutationreproducerpin.go` does.
- Parity: the Go guard prints the same stdout and stderr bytes and exits with the same codes as the bash at `b1b0d601` for every input the bash harness exercises. Read the bash with `git show b1b0d601:scripts/<name>.sh` before writing the Go. Reuse the existing Go helpers — `stripBOM`, `sanitizeDisplay`, `vtSplitCells`, `vtTrimCell`, `vtFoldCell` (`stats/internal/guard/visualsection.go`) — never a new copy.
- A guard that reads the environment or runs a subprocess does so through the `Env` it is given, so the table test can set `PATH` and variables in-process.
- Go: `gofmt -w` every Go file you change before committing; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Never start, stop or restart `flowd` on `127.0.0.1:4173` or the `flow-postgres` container.

## Review Focus

- Every bash harness case has a Go subtest with the same label and the same expected exit code and output.
- Exit 2 paths (missing file, not a directory, ambiguous section, unreadable config, missing tool) keep their exact messages.
- `check-spec-reach`: the npm/npx invocation, its working directory and the JSON parsing match the bash; `sanitizeDisplay` is applied to exactly the lines the bash piped through `sanitize_display`.
- `check-dev-stack-fresh`: the fingerprint command still runs in the worktree via the shell, its output still streams, and the `stack-stale` verdict and exit codes are unchanged.
- `check-uitest-overrides`: include resolution relative to the Makefile, the default Makefile path when no argument is given, and the `override`/`eval` cases.

- [x] 1. Port check-dev-stack-fresh to Go
  - [x] **Step 1: Read the bash and its harness.** `git show b1b0d601:scripts/check-dev-stack-fresh.sh` and `git show b1b0d601:scripts/test-check-dev-stack-fresh.sh`; list every harness case label.
  - [x] **Step 2: Write the Go test first.** `stats/internal/guard/check_dev_stack_fresh_test.go`: `TestCheckDevStackFresh`, one subtest per harness case, fixtures in `t.TempDir()`. Run it and see it fail (no guard registered).
  - [x] **Step 3: Write the guard.** `stats/internal/guard/devstackfresh.go`, `Registry["check-dev-stack-fresh"]`; run the fingerprint row through `sh -c` in the worktree with stdout and stderr streamed, as the bash does.
  - [x] **Step 4: Reduce the script to a shim** and `git rm scripts/test-check-dev-stack-fresh.sh`.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./... && go test ./internal/guard -run 'TestCheckDevStackFresh'`; `scripts/check-guard-symlinks.sh`; `scripts/check-dev-stack-fresh.sh .` prints the fingerprint line of design.md's `## Live check` and the same exit code as before.
**Files:** `scripts/check-dev-stack-fresh.sh`, `scripts/test-check-dev-stack-fresh.sh`, `stats/internal/guard/devstackfresh.go`, `stats/internal/guard/check_dev_stack_fresh_test.go`
**Tests:** `TestCheckDevStackFresh`
**Regression:** `TestCheckDevStackFresh` fails — no `check-dev-stack-fresh` guard registered — when the commit is reverted.
**Baseline:** before=245 after=246
<!-- measured: grep -rhoE '^func Test[A-Za-z0-9_]+\(' stats/internal/guard | wc -l @ b1b0d601 → 245 -->
**Build:** green
**After:** none
**Commit:** refactor(guard): port check-dev-stack-fresh to Go

**Decision:** task-per-guard
**Decision:** harness-cases-by-label

- [x] 2. Port check-spec-reach to Go
  - [x] **Step 1: Read the bash and its harness.** `git show b1b0d601:scripts/check-spec-reach.sh` and `git show b1b0d601:scripts/test-check-spec-reach.sh`; list every harness case label and the fake `npm`/`npx` binaries it puts on `PATH`.
  - [x] **Step 2: Write the Go test first.** `stats/internal/guard/check_spec_reach_test.go`: `TestCheckSpecReach`, one subtest per harness case, the same fake binaries written into a temp dir and put first on the `PATH` the guard's `Env` returns. Run it and see it fail.
  - [x] **Step 3: Write the guard.** `stats/internal/guard/specreach.go`, `Registry["check-spec-reach"]`.
  - [x] **Step 4: Reduce the script to a shim** and `git rm scripts/test-check-spec-reach.sh`.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./... && go test ./internal/guard -run 'TestCheckSpecReach'`; `scripts/check-guard-symlinks.sh`; `bash scripts/test-run-verify-list.sh`; `scripts/check-spec-reach.sh .` prints `Spec reach: not configured` and exits 0.
**Files:** `scripts/check-spec-reach.sh`, `scripts/test-check-spec-reach.sh`, `stats/internal/guard/specreach.go`, `stats/internal/guard/check_spec_reach_test.go`
**Tests:** `TestCheckSpecReach`
**Regression:** `TestCheckSpecReach` fails — no `check-spec-reach` guard registered — when the commit is reverted.
**Baseline:** before=246 after=247
<!-- measured: grep -rhoE '^func Test[A-Za-z0-9_]+\(' stats/internal/guard | wc -l @ b1b0d601 → 245, plus Task 1's one -->
**Build:** green
**After:** none
**Commit:** refactor(guard): port check-spec-reach to Go

Correction (2026-10-22): the plan declared byte parity for every input; the orphan lines now print in byte order, where the bash's `sort -z` collated by the operator's locale — identical under `LC_ALL=C` (measured on a fixture with `b-c`, `B` and `b\x` specs), different under `en_US.UTF-8`. Matching a locale's collation needs a collation library this module does not carry; byte order is the deterministic choice. The plan also named `npx`, which the bash never ran: the port runs `node` and `npm` exactly as the bash did.

**Decision:** task-per-guard
**Decision:** harness-cases-by-label

- [x] 3. Port check-uitest-overrides to Go
  - [x] **Step 1: Read the bash and its harness.** `git show b1b0d601:scripts/check-uitest-overrides.sh` and `git show b1b0d601:scripts/test-check-uitest-overrides.sh`; list every harness case label.
  - [x] **Step 2: Write the Go test first.** `stats/internal/guard/check_uitest_overrides_test.go`: `TestCheckUitestOverrides`, one subtest per harness case. Run it and see it fail.
  - [x] **Step 3: Write the guard.** `stats/internal/guard/uitestoverrides.go`, `Registry["check-uitest-overrides"]`; the default Makefile resolves exactly as the bash's does when no argument is given.
  - [x] **Step 4: Reduce the script to a shim** and `git rm scripts/test-check-uitest-overrides.sh`.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./... && go test ./internal/guard -run 'TestCheckUitestOverrides'`; `scripts/check-guard-symlinks.sh`; `scripts/check-uitest-overrides.sh` and `cd stats && make -s check-uitest-overrides` each print `UITEST-OVERRIDES-OK: every UITEST_* variable is override` and exit 0.
**Files:** `scripts/check-uitest-overrides.sh`, `scripts/test-check-uitest-overrides.sh`, `stats/internal/guard/uitestoverrides.go`, `stats/internal/guard/check_uitest_overrides_test.go`
**Tests:** `TestCheckUitestOverrides`
**Regression:** `TestCheckUitestOverrides` fails — no `check-uitest-overrides` guard registered — when the commit is reverted.
**Baseline:** before=247 after=248
<!-- measured: grep -rhoE '^func Test[A-Za-z0-9_]+\(' stats/internal/guard | wc -l @ b1b0d601 → 245, plus Tasks 1 and 2's two -->
**Build:** green
**After:** none
**Commit:** refactor(guard): port check-uitest-overrides to Go

**Decision:** task-per-guard
**Decision:** harness-cases-by-label

- [x] 4. Delete strip-bom.sh, sanitize-display.sh and visual-table-cells.awk
  - [x] **Step 1: Confirm no bash or awk caller remains.** `git grep -n -e 'lib/strip-bom.sh' -e 'lib/sanitize-display.sh' -e 'lib/visual-table-cells.awk' -e strip_bom_cat -e sanitize_display -- scripts` shows only comments and the three files themselves; a remaining `source` or `awk -f` stops the task and is reported.
  - [x] **Step 2: Point the parity tests at history.** In `stats/internal/guard/libtwins_test.go`, `TestStripBOMParity`, `TestSanitizeDisplayParity` and `TestVisualTableCellsParity` read their bash with `historicalFile` at `b1b0d601` (a new `const` beside `libTwinBase`, commented with this change's name) instead of from `../../../scripts/lib/`. Run `go test ./internal/guard -run 'TestStripBOMParity|TestSanitizeDisplayParity|TestVisualTableCellsParity'` and see it pass before deleting.
  - [x] **Step 3: Delete the libraries.** `git rm scripts/lib/strip-bom.sh scripts/lib/sanitize-display.sh scripts/lib/visual-table-cells.awk`.
  - [x] **Step 4: Update citations.** Every comment naming a deleted file is rewritten to name its Go twin in `stats/internal/guard/visualsection.go` or dropped: `scripts/check-visual-verification.sh`, `scripts/check-visual-trigger.sh`, `scripts/lib/coverage.sh`, `scripts/lib/spec-root.sh`, `scripts/lib/resolve-file.sh` (its header also drops `check-dev-stack-fresh.sh` from the files that source it), `stats/internal/guard/visualsection.go`, `stats/internal/guard/resolvefile.go`.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./... && go test ./internal/guard -run 'TestStripBOMParity|TestSanitizeDisplayParity|TestVisualTableCellsParity|TestProjectGet|TestResolveFileParity'`; `scripts/check-references.sh`; `scripts/check-guard-symlinks.sh`; then the absence sweep, which prints nothing: `git grep -n -F -e 'strip-bom.sh' -e 'sanitize-display.sh' -e 'visual-table-cells.awk' -e 'strip_bom_cat' -e 'sanitize_display' -- scripts skills stats ':!stats/internal/guard/libtwins_test.go' ':!stats/internal/guard/project_get_test.go'`.
**Files:** `scripts/lib/strip-bom.sh`, `scripts/lib/sanitize-display.sh`, `scripts/lib/visual-table-cells.awk`, `scripts/lib/resolve-file.sh`, `scripts/check-visual-verification.sh`, `scripts/check-visual-trigger.sh`, `stats/internal/guard/libtwins_test.go`, `stats/internal/guard/visualsection.go`, `stats/internal/guard/resolvefile.go`
**Tests:** none — repairs existing tests: `TestStripBOMParity`, `TestSanitizeDisplayParity`, `TestVisualTableCellsParity`
**Regression:** none added; the repointed parity tests fail on any Go-twin drift from the bash at `b1b0d601`.
**Baseline:** before=248 after=248
<!-- measured: grep -rhoE '^func Test[A-Za-z0-9_]+\(' stats/internal/guard | wc -l @ b1b0d601 → 245, plus Tasks 1–3's three -->
**Build:** green
**After:** Task 1, 2, 3
**Commit:** refactor(guard): delete the bash libraries no guard sources

Correction (2026-10-22): `scripts/lib/coverage.sh` and `scripts/lib/spec-root.sh` named none of the deleted files, so they are dropped from **Files:**. The absence sweep still prints five lines the plan did not foresee — `stats/internal/guard/check_visual_trigger_test.go` (four comments describing what its harness asserted through `sanitize_display`) and `stats/internal/guard/text.go` (`ccSanitize`'s doc naming the bash function it was measured against); each names the bash function as history, not a live file, and is left as written. The parity tests' new constant is `visualLibBase` with a `visualLibAt` reader, not a direct `historicalFile` call per test.

**Decision:** delete-visual-table-cells
**Decision:** keep-resolve-file
**Decision:** parity-at-merge-base
