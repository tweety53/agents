# kan-1021-flow-improvement-move-the-bash-hook-and-setup

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

Ports nine `scripts/test-*.sh` harnesses to Go tests and deletes them (one task each), then
rewires `pipeline-fix`'s check selection. `design.md` is canonical for every decision; each task
cites its entry by ID.

**The harness at `e77f951b` is each port's specification.** Every case it asserts becomes one Go
subtest asserting the same thing — same inputs, same expected decision/exit/output. The harness's
header comment moves to the Go file's doc comment, corrected where it describes bash plumbing.

**Package:** `stats/internal/repotest/` (new; package `repotest`, test-only files plus a
`doc.go` in Task 1). The repo root is resolved from the test's working directory
(`stats/internal/repotest` → `../../..`) via a shared helper `repoRoot(t)` in
`stats/internal/repotest/helpers_test.go`, added by Task 1 and reused by every later task.

**Test isolation:** every top-level test and subtest calls `t.Parallel()` where the harness's
cases are independent; fixtures live in `t.TempDir()`; `HOME` is set to a temp dir on every
exec'd child; no test reads or writes the operator's real `$HOME`. External tools (`python3`,
`git`, `make`, `node`) are exec'd through `os/exec` exactly as the harness ran them.

**Per-task verify:** `cd stats && gofmt -l ./internal/ && go vet ./internal/repotest/ && go test
./internal/repotest/ -run '^<Test>$' -count=1 -race -v`, then the subtest count check the task
names, then the absence sweep. The full suite runs in `flow.verify`.

## Review Focus

- Case parity: every assertion of each deleted harness has a Go subtest asserting the same thing;
  none silently dropped or weakened.
- No test touches the real `$HOME`, `~/.claude`, or the main checkout's git config.
- `pfChecks` still selects a check for a touched hook and for a touched Go file after the scripts
  are gone.

- [ ] 1. Port test-protect-main-checkout.sh

**Files:** `stats/internal/repotest/protect_main_checkout_test.go`, `scripts/test-protect-main-checkout.sh`, `stats/internal/repotest/doc.go`, `stats/internal/repotest/helpers_test.go`
**Tests:** `TestProtectMainCheckout`
**Regression:** `TestProtectMainCheckout` fails if any behaviour `scripts/test-protect-main-checkout.sh` pinned at `e77f951b` regresses.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/repotest/*_test.go 2>/dev/null | grep -cE '^func Test' @ e77f951b -->
**After:** none
**Commit:** `test(repotest): port test-protect-main-checkout.sh to Go`
**Build:** green

**Decision:** tests-only
**Decision:** delete-not-shim
**Decision:** repotest-package
**Decision:** case-parity

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-protect-main-checkout.sh` to `TestProtectMainCheckout` in `stats/internal/repotest/protect_main_checkout_test.go`, one subtest per case, named after the harness's own case label. Run `cd stats && go test ./internal/repotest/ -run '^TestProtectMainCheckout$' -count=1` — expect it to fail only where an assertion was mis-ported; fix the port until green against the untouched hook/source (the code under test does not change in this task).
  - [ ] **Step 2: Package scaffolding.** `doc.go` carries the package doc (what `repotest` holds: Go ports of the repo-level bash harnesses; tests exec the Python hooks and repo tools, never import them). `helpers_test.go` carries `repoRoot(t)` and a `runHook(t, hook string, payload []byte, env ...string) (stdout, stderr string, rc int)` that execs `python3 <repo>/hooks/<hook>` with `HOME` set to `t.TempDir()` — the shared exec seam Tasks 2–4 reuse.
  - [ ] **Step 3: Parity check.** `cd stats && go test ./internal/repotest/ -run '^TestProtectMainCheckout$' -count=1 -race -v | grep -c -- '--- PASS: TestProtectMainCheckout/'` equals the harness's own case count, read from `bash scripts/test-protect-main-checkout.sh` run once before deleting it (count its `ok:`/`PASS` lines, or its case list where it prints none).
  - [ ] **Step 4: Delete** `scripts/test-protect-main-checkout.sh`.
  - [ ] **Step 5: Absence sweep.** `git grep -n -F -e 'test-protect-main-checkout.sh' -- stats scripts hooks mods README.md CLAUDE.md AGENTS.md .flow` prints nothing.
- [ ] 2. Port test-enforce-agent-baseline-hook.sh

**Files:** `stats/internal/repotest/enforce_agent_baseline_hook_test.go`, `scripts/test-enforce-agent-baseline-hook.sh`, `README.md`
**Tests:** `TestEnforceAgentBaselineHook`
**Regression:** `TestEnforceAgentBaselineHook` fails if any behaviour `scripts/test-enforce-agent-baseline-hook.sh` pinned at `e77f951b` regresses.
**Baseline:** before=1 after=2
<!-- measured: cat stats/internal/repotest/*_test.go 2>/dev/null | grep -cE '^func Test' @ e77f951b -->
**After:** Task 1
**Commit:** `test(repotest): port test-enforce-agent-baseline-hook.sh to Go`
**Build:** green

**Decision:** tests-only
**Decision:** delete-not-shim
**Decision:** repotest-package
**Decision:** case-parity

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-enforce-agent-baseline-hook.sh` to `TestEnforceAgentBaselineHook` in `stats/internal/repotest/enforce_agent_baseline_hook_test.go`, one subtest per case, named after the harness's own case label. Run `cd stats && go test ./internal/repotest/ -run '^TestEnforceAgentBaselineHook$' -count=1` — expect it to fail only where an assertion was mis-ported; fix the port until green against the untouched hook/source (the code under test does not change in this task).
  - [ ] **Step 2: Doc reference.** `README.md`'s hook bullet cites `stats/internal/repotest/enforce_agent_baseline_hook_test.go` in place of the deleted script.
  - [ ] **Step 3: Parity check.** `cd stats && go test ./internal/repotest/ -run '^TestEnforceAgentBaselineHook$' -count=1 -race -v | grep -c -- '--- PASS: TestEnforceAgentBaselineHook/'` equals the harness's own case count, read from `bash scripts/test-enforce-agent-baseline-hook.sh` run once before deleting it (count its `ok:`/`PASS` lines, or its case list where it prints none).
  - [ ] **Step 4: Delete** `scripts/test-enforce-agent-baseline-hook.sh`.
  - [ ] **Step 5: Absence sweep.** `git grep -n -F -e 'test-enforce-agent-baseline-hook.sh' -- stats scripts hooks mods README.md CLAUDE.md AGENTS.md .flow` prints nothing.
- [ ] 3. Port test-enforce-loop-dispatch-hook.sh

**Files:** `stats/internal/repotest/enforce_loop_dispatch_hook_test.go`, `scripts/test-enforce-loop-dispatch-hook.sh`
**Tests:** `TestEnforceLoopDispatchHook`
**Regression:** `TestEnforceLoopDispatchHook` fails if any behaviour `scripts/test-enforce-loop-dispatch-hook.sh` pinned at `e77f951b` regresses.
**Baseline:** before=2 after=3
<!-- measured: cat stats/internal/repotest/*_test.go 2>/dev/null | grep -cE '^func Test' @ e77f951b -->
**After:** Task 1
**Commit:** `test(repotest): port test-enforce-loop-dispatch-hook.sh to Go`
**Build:** green

**Decision:** tests-only
**Decision:** delete-not-shim
**Decision:** repotest-package
**Decision:** case-parity

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-enforce-loop-dispatch-hook.sh` to `TestEnforceLoopDispatchHook` in `stats/internal/repotest/enforce_loop_dispatch_hook_test.go`, one subtest per case, named after the harness's own case label. Run `cd stats && go test ./internal/repotest/ -run '^TestEnforceLoopDispatchHook$' -count=1` — expect it to fail only where an assertion was mis-ported; fix the port until green against the untouched hook/source (the code under test does not change in this task).
  - [ ] **Step 2: Stub checker.** The harness's stub checker (writes its argv and `pwd -P` to files, exits with a chosen code) is written into `t.TempDir()` per subtest and passed through `LOOP_DISPATCH_CHECK`, as the harness does.
  - [ ] **Step 3: Parity check.** `cd stats && go test ./internal/repotest/ -run '^TestEnforceLoopDispatchHook$' -count=1 -race -v | grep -c -- '--- PASS: TestEnforceLoopDispatchHook/'` equals the harness's own case count, read from `bash scripts/test-enforce-loop-dispatch-hook.sh` run once before deleting it (count its `ok:`/`PASS` lines, or its case list where it prints none).
  - [ ] **Step 4: Delete** `scripts/test-enforce-loop-dispatch-hook.sh`.
  - [ ] **Step 5: Absence sweep.** `git grep -n -F -e 'test-enforce-loop-dispatch-hook.sh' -- stats scripts hooks mods README.md CLAUDE.md AGENTS.md .flow` prints nothing.
- [ ] 4. Port test-flow-active-change-hook.sh

**Files:** `stats/internal/repotest/flow_active_change_hook_test.go`, `scripts/test-flow-active-change-hook.sh`
**Tests:** `TestFlowActiveChangeHook`
**Regression:** `TestFlowActiveChangeHook` fails if any behaviour `scripts/test-flow-active-change-hook.sh` pinned at `e77f951b` regresses.
**Baseline:** before=3 after=4
<!-- measured: cat stats/internal/repotest/*_test.go 2>/dev/null | grep -cE '^func Test' @ e77f951b -->
**After:** Task 1
**Commit:** `test(repotest): port test-flow-active-change-hook.sh to Go`
**Build:** green

**Decision:** tests-only
**Decision:** delete-not-shim
**Decision:** repotest-package
**Decision:** case-parity

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-flow-active-change-hook.sh` to `TestFlowActiveChangeHook` in `stats/internal/repotest/flow_active_change_hook_test.go`, one subtest per case, named after the harness's own case label. Run `cd stats && go test ./internal/repotest/ -run '^TestFlowActiveChangeHook$' -count=1` — expect it to fail only where an assertion was mis-ported; fix the port until green against the untouched hook/source (the code under test does not change in this task).
  - [ ] **Step 2: Fixtures.** Every transcript/env fixture the harness builds is built under `t.TempDir()` by the subtest that uses it.
  - [ ] **Step 3: Parity check.** `cd stats && go test ./internal/repotest/ -run '^TestFlowActiveChangeHook$' -count=1 -race -v | grep -c -- '--- PASS: TestFlowActiveChangeHook/'` equals the harness's own case count, read from `bash scripts/test-flow-active-change-hook.sh` run once before deleting it (count its `ok:`/`PASS` lines, or its case list where it prints none).
  - [ ] **Step 4: Delete** `scripts/test-flow-active-change-hook.sh`.
  - [ ] **Step 5: Absence sweep.** `git grep -n -F -e 'test-flow-active-change-hook.sh' -- stats scripts hooks mods README.md CLAUDE.md AGENTS.md .flow` prints nothing.
- [ ] 5. Port test-git-config-pins.sh

**Files:** `stats/internal/repotest/git_config_pins_test.go`, `scripts/test-git-config-pins.sh`
**Tests:** `TestGitConfigPins`
**Regression:** `TestGitConfigPins` fails if any behaviour `scripts/test-git-config-pins.sh` pinned at `e77f951b` regresses.
**Baseline:** before=4 after=5
<!-- measured: cat stats/internal/repotest/*_test.go 2>/dev/null | grep -cE '^func Test' @ e77f951b -->
**After:** Task 1
**Commit:** `test(repotest): port test-git-config-pins.sh to Go`
**Build:** green

**Decision:** tests-only
**Decision:** delete-not-shim
**Decision:** repotest-package
**Decision:** case-parity

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-git-config-pins.sh` to `TestGitConfigPins` in `stats/internal/repotest/git_config_pins_test.go`, one subtest per case, named after the harness's own case label. Run `cd stats && go test ./internal/repotest/ -run '^TestGitConfigPins$' -count=1` — expect it to fail only where an assertion was mis-ported; fix the port until green against the untouched hook/source (the code under test does not change in this task).
  - [ ] **Step 2: Self-exclusion.** The harness excludes itself (`SELF_NAME`) from its own audit scan; the Go test excludes `stats/internal/repotest/git_config_pins_test.go` the same way.
  - [ ] **Step 3: Parity check.** `cd stats && go test ./internal/repotest/ -run '^TestGitConfigPins$' -count=1 -race -v | grep -c -- '--- PASS: TestGitConfigPins/'` equals the harness's own case count, read from `bash scripts/test-git-config-pins.sh` run once before deleting it (count its `ok:`/`PASS` lines, or its case list where it prints none).
  - [ ] **Step 4: Delete** `scripts/test-git-config-pins.sh`.
  - [ ] **Step 5: Absence sweep.** `git grep -n -F -e 'test-git-config-pins.sh' -- stats scripts hooks mods README.md CLAUDE.md AGENTS.md .flow` prints nothing.
- [ ] 6. Port test-flow-addr-declaration.sh

**Files:** `stats/internal/repotest/flow_addr_declaration_test.go`, `scripts/test-flow-addr-declaration.sh`
**Tests:** `TestFlowAddrDeclaration`
**Regression:** `TestFlowAddrDeclaration` fails if any behaviour `scripts/test-flow-addr-declaration.sh` pinned at `e77f951b` regresses.
**Baseline:** before=5 after=6
<!-- measured: cat stats/internal/repotest/*_test.go 2>/dev/null | grep -cE '^func Test' @ e77f951b -->
**After:** Task 1
**Commit:** `test(repotest): port test-flow-addr-declaration.sh to Go`
**Build:** green

**Decision:** tests-only
**Decision:** delete-not-shim
**Decision:** repotest-package
**Decision:** case-parity

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-flow-addr-declaration.sh` to `TestFlowAddrDeclaration` in `stats/internal/repotest/flow_addr_declaration_test.go`, one subtest per case, named after the harness's own case label. Run `cd stats && go test ./internal/repotest/ -run '^TestFlowAddrDeclaration$' -count=1` — expect it to fail only where an assertion was mis-ported; fix the port until green against the untouched hook/source (the code under test does not change in this task).
  - [ ] **Step 2: Sources read.** Reads `.flow/project.md` and `stats/cmd/flow` from `repoRoot(t)`, as the harness does.
  - [ ] **Step 3: Parity check.** `cd stats && go test ./internal/repotest/ -run '^TestFlowAddrDeclaration$' -count=1 -race -v | grep -c -- '--- PASS: TestFlowAddrDeclaration/'` equals the harness's own case count, read from `bash scripts/test-flow-addr-declaration.sh` run once before deleting it (count its `ok:`/`PASS` lines, or its case list where it prints none).
  - [ ] **Step 4: Delete** `scripts/test-flow-addr-declaration.sh`.
  - [ ] **Step 5: Absence sweep.** `git grep -n -F -e 'test-flow-addr-declaration.sh' -- stats scripts hooks mods README.md CLAUDE.md AGENTS.md .flow` prints nothing.
- [ ] 7. Port test-make-build.sh

**Files:** `stats/internal/repotest/make_build_test.go`, `scripts/test-make-build.sh`, `stats/Makefile`
**Tests:** `TestMakeBuild`
**Regression:** `TestMakeBuild` fails if any behaviour `scripts/test-make-build.sh` pinned at `e77f951b` regresses.
**Baseline:** before=6 after=7
<!-- measured: cat stats/internal/repotest/*_test.go 2>/dev/null | grep -cE '^func Test' @ e77f951b -->
**After:** Task 1
**Commit:** `test(repotest): port test-make-build.sh to Go`
**Build:** green

**Decision:** tests-only
**Decision:** delete-not-shim
**Decision:** repotest-package
**Decision:** case-parity

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-make-build.sh` to `TestMakeBuild` in `stats/internal/repotest/make_build_test.go`, one subtest per case, named after the harness's own case label. Run `cd stats && go test ./internal/repotest/ -run '^TestMakeBuild$' -count=1` — expect it to fail only where an assertion was mis-ported; fix the port until green against the untouched hook/source (the code under test does not change in this task).
  - [ ] **Step 2: Makefile comments.** The four `stats/Makefile` comments naming `scripts/test-make-build.sh` name `stats/internal/repotest/make_build_test.go` instead; recipe lines unchanged. `make -n` cases run in a copy or with `-C <repo>/stats` read-only, as the harness does — never a real build into the tree.
  - [ ] **Step 3: Parity check.** `cd stats && go test ./internal/repotest/ -run '^TestMakeBuild$' -count=1 -race -v | grep -c -- '--- PASS: TestMakeBuild/'` equals the harness's own case count, read from `bash scripts/test-make-build.sh` run once before deleting it (count its `ok:`/`PASS` lines, or its case list where it prints none).
  - [ ] **Step 4: Delete** `scripts/test-make-build.sh`.
  - [ ] **Step 5: Absence sweep.** `git grep -n -F -e 'test-make-build.sh' -- stats scripts hooks mods README.md CLAUDE.md AGENTS.md .flow` prints nothing.
- [ ] 8. Port test-flow-task-list-landed.sh

**Files:** `stats/internal/repotest/flow_task_list_landed_test.go`, `scripts/test-flow-task-list-landed.sh`, `mods/flow-task-list/hooks/landed.ts`
**Tests:** `TestFlowTaskListLanded`
**Regression:** `TestFlowTaskListLanded` fails if any behaviour `scripts/test-flow-task-list-landed.sh` pinned at `e77f951b` regresses.
**Baseline:** before=7 after=8
<!-- measured: cat stats/internal/repotest/*_test.go 2>/dev/null | grep -cE '^func Test' @ e77f951b -->
**After:** Task 1
**Commit:** `test(repotest): port test-flow-task-list-landed.sh to Go`
**Build:** green

**Decision:** tests-only
**Decision:** delete-not-shim
**Decision:** repotest-package
**Decision:** case-parity

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-flow-task-list-landed.sh` to `TestFlowTaskListLanded` in `stats/internal/repotest/flow_task_list_landed_test.go`, one subtest per case, named after the harness's own case label. Run `cd stats && go test ./internal/repotest/ -run '^TestFlowTaskListLanded$' -count=1` — expect it to fail only where an assertion was mis-ported; fix the port until green against the untouched hook/source (the code under test does not change in this task).
  - [ ] **Step 2: Runner and comment.** Runs `landed.ts` under `node` exactly as the harness does (same flags), against the same temp-repo fixture; skips with the harness's reason if it had one when `node` is absent, else fails. `landed.ts`'s line-2 comment names the Go test instead of the script.
  - [ ] **Step 3: Parity check.** `cd stats && go test ./internal/repotest/ -run '^TestFlowTaskListLanded$' -count=1 -race -v | grep -c -- '--- PASS: TestFlowTaskListLanded/'` equals the harness's own case count, read from `bash scripts/test-flow-task-list-landed.sh` run once before deleting it (count its `ok:`/`PASS` lines, or its case list where it prints none).
  - [ ] **Step 4: Delete** `scripts/test-flow-task-list-landed.sh`.
  - [ ] **Step 5: Absence sweep.** `git grep -n -F -e 'test-flow-task-list-landed.sh' -- stats scripts hooks mods README.md CLAUDE.md AGENTS.md .flow` prints nothing.
- [ ] 9. Port test-setup-agents.sh and drop the test-setup.sh shim

**Files:** `stats/internal/setuptest/agents_test.go`, `stats/internal/setuptest/main_test.go`, `scripts/test-setup-agents.sh`, `scripts/test-setup.sh`
**Tests:** `TestSetupAgents`
**Regression:** `TestSetupAgents` fails if `setup.sh global` stops installing or resolving the three `flow-<effort>` agents, or their frontmatter gains a `model:` line.
**Baseline:** before=41 after=42
<!-- measured: cat stats/internal/setuptest/*_test.go | grep -cE '^func Test' @ e77f951b -->
**After:** none
**Commit:** `test(setuptest): port test-setup-agents.sh to Go`
**Build:** green

**Decision:** delete-not-shim
**Decision:** repotest-package
**Decision:** case-parity

  - [ ] **Step 1: Failing test.** Port every case of `scripts/test-setup-agents.sh` to `TestSetupAgents` in `stats/internal/setuptest/agents_test.go`, reusing the package's existing sandbox helpers (`helpers_test.go`) for `setup.sh global` against a temp `HOME`. Run `cd stats && go test ./internal/setuptest/ -run '^TestSetupAgents$' -count=1 -v` — green against the untouched `setup.sh`.
  - [ ] **Step 2: Parity check.** Subtest count equals the harness's assertion count (three symlinks resolve; each `flow-<effort>` frontmatter has its `effort:` line and no `model:` line).
  - [ ] **Step 3: Delete** `scripts/test-setup-agents.sh` and `scripts/test-setup.sh` (the latter already only ran this package); `stats/internal/setuptest/main_test.go`'s package doc names `## test`'s Go suite as its runner instead of the script.
  - [ ] **Step 4: Absence sweep.** `git grep -n -F -e 'test-setup-agents.sh' -e 'scripts/test-setup.sh' -- stats/internal/setuptest scripts hooks mods README.md CLAUDE.md AGENTS.md .flow` prints nothing.
- [ ] 10. Rewire pipeline-fix check selection to the Go packages

**Files:** `stats/internal/guard/pipelinefix.go`, `stats/internal/guard/pipeline_fix_test.go`
**Tests:** `TestPipelineFixChecksHooks`
**Regression:** `TestPipelineFixChecksHooks` fails if a touched hook selects no check; the rewritten `TestPipelineFixChecksGoAudits` fails if a touched Go file no longer runs the two audits.
**Baseline:** before=11 after=12
<!-- measured: cat stats/internal/guard/pipeline_fix_test.go | grep -cE '^func Test' @ e77f951b -->
**After:** Task 5, 6
**Commit:** `fix(guard): select repotest packages in pipeline-fix checks`
**Build:** green

**Decision:** pipelinefix-packages

  - [ ] **Step 1: Failing test.** In `stats/internal/guard/pipeline_fix_test.go`: rewrite `TestPipelineFixChecksGoAudits` to assert a touched Go file selects one command running the `stats/internal/repotest` package filtered to `TestGitConfigPins` and `TestFlowAddrDeclaration` (no fixture scripts written) and a touched shell script under `scripts/` alone does not; add `TestPipelineFixChecksHooks` asserting a touched `hooks/protect-main-checkout.py` selects the whole `stats/internal/repotest` package. Run — expect failure.
  - [ ] **Step 2: Implement.** In `stats/internal/guard/pipelinefix.go`: `pfGoAudits` becomes that one filtered package command (doc comment updated: the audits live in `stats/internal/repotest/`); the `hooks/` prefix gets its own `case` adding the whole-package `stats/internal/repotest` command, leaving the `scripts/` stem lookup for remaining `test-*.sh` harnesses.
  - [ ] **Step 3: Green.** `cd stats && go test ./internal/guard/ -run '^TestPipelineFixChecks' -count=1 -race -v`.
  - [ ] **Step 4: Absence sweep.** `git grep -n -F -e 'test-flow-addr-declaration.sh' -e 'test-git-config-pins.sh' -- stats/internal/guard` prints nothing.
