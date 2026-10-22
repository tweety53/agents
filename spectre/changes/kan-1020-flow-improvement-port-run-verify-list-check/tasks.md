# kan-1020-flow-improvement-port-run-verify-list-check

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** yes — each bash header's contract text moves verbatim into its Go file's doc comment
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

Ports three bash guards into `flow-guard` behind `flow_guard_exec` shims (tasks 1–3) and deletes two
libraries no script sources any more (task 4). `design.md` is canonical for every decision; each task cites
its entry by ID.

**The scripts at `8c54d6d3` are the specification.** Each script's header is the contract and its body the
behaviour; its `scripts/test-<name>.sh` holds the cases. The port reproduces every stdout and stderr line byte
for byte, the stdout/stderr split and every exit code. Where a script and its header disagree, stop and report —
never pick one silently.

**Shim template** — the shape of `scripts/check-spec-reach.sh`: the bash header's usage and exit-code lines
kept, a pointer at the Go file, then:

```bash unverified:confirm each shim's opening string matches the bash's own stderr prefix
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "<name>: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
: "$SCRIPT_DIR/<sibling>.sh"   # one per sibling the guard execs
FLOW_GUARD_SCRIPT_DIR="$SCRIPT_DIR"
export FLOW_GUARD_SCRIPT_DIR
flow_guard_exec <name> 2 "<name>:" "$@"
```

The Go reads every sibling path from `FLOW_GUARD_SCRIPT_DIR`, as `integratechange.go` does; an unset one is
exit 2 naming the shim. `check-narrative-runs` execs no sibling and omits both the `:` lines and the export.

**Test isolation:** every Go test calls `t.Parallel()`; fixtures are built per case in `t.TempDir()`; `PATH`
(stubs first), `FLOW_GUARD_SCRIPT_DIR` and `FLOW_GUARD_CACHE_DIR` (a per-test temp dir) pass on `Env` or the
child's own environment, never `os.Setenv`; no test touches the operator's `$HOME`, cache or this tree.

## Review Focus

- Parity subtests compare the Go guard with the bash read at `8c54d6d3`, never with the live shim.
- `rgtRun`'s change keeps `run-guard-tests`' behaviour exactly: `bash <harness>`, list-order results, one
  capture file per job.
- `run-verify-list`'s green-sha file and `flow-verify-failed` logs land in the git dir, never in the tree.

---

- [x] 1. Port check-narrative-runs.sh to flow-guard

**Files:** `stats/internal/guard/narrativeruns.go`, `stats/internal/guard/check_narrative_runs_test.go`, `scripts/check-narrative-runs.sh`, `scripts/test-check-narrative-runs.sh`
**Tests:** `TestCheckNarrativeRuns`
**Regression:** fails if any case of the deleted `scripts/test-check-narrative-runs.sh` regresses — the OK and MISSING verdict lines, an absent narrative counted as zero sections, an unreadable narrative, a `flow` call that fails or prints a non-integer (exit 2 with its stderr line) — or if the Go output differs from the bash's at `8c54d6d3`.
**Baseline:** before=269 after=270
<!-- measured: cat stats/internal/guard/*_test.go | grep -cE '^func Test[A-Za-z0-9_]*\(t \*testing\.T\)' @ 8c54d6d3 -->
**After:** none
**Commit:** `refactor(flow-guard): port check-narrative-runs to Go`
**Build:** green

**Decision:** guards-in-flow-guard
**Decision:** bash-at-8c54d6d3-is-the-oracle
**Decision:** narrative-runs-exec-flow

  - [x] **Step 1: Record the oracle.** `bash scripts/test-check-narrative-runs.sh` on the untouched tree — exit 0, every line `ok:`.
  - [x] **Step 2: Port.** Write `narrativeruns.go`: `Registry["check-narrative-runs"]`, the bash header verbatim as its doc comment, `flow stage sessions` exec'd via PATH with `-C <wt>`, the two section counts as the same `^## .* — fix run` / `^## .* — integrate run` line tests.
  - [x] **Step 3: Shim.** Replace `scripts/check-narrative-runs.sh` per the template; re-run Step 1's command against the shim — exit 0. A failing assertion is a port defect: fix the Go, never the harness.
  - [x] **Step 4: Go test.** Write `check_narrative_runs_test.go`: `TestCheckNarrativeRuns`, one `t.Run` per harness case (labels kept), each running `Registry["check-narrative-runs"]` in process and the bash read at `8c54d6d3` over the same stub `flow` and narrative, asserting both the harness's expectation and byte-for-byte parity. Delete `scripts/test-check-narrative-runs.sh`.
  - [x] **Step 5: Break and prove.** Change the MISSING line's text in `narrativeruns.go`, see the test fail, restore it.
  - [x] **Step 6: Verify.** `cd stats && gofmt -l ./internal/guard && go vet ./internal/guard && go test ./internal/guard -run 'TestCheckNarrativeRuns' -race -count=1`; `bash scripts/check-guard-symlinks.sh` and `bash scripts/check-references.sh` exit 0.

- [x] 2. Port rebuild-full-app-suite-zip.sh to flow-guard

**Files:** `stats/internal/guard/rebuildfullappsuitezip.go`, `stats/internal/guard/rebuild_full_app_suite_zip_test.go`, `scripts/rebuild-full-app-suite-zip.sh`, `scripts/test-rebuild-full-app-suite-zip.sh`
**Tests:** `TestRebuildFullAppSuiteZip`
**Regression:** fails if any case of the deleted `scripts/test-rebuild-full-app-suite-zip.sh` regresses — the zip holding exactly the resolved PNGs, a stale PNG dropped, a rebuild over unchanged images byte-identical, a resolver exit 1 removing the zip and exiting 1, a resolver exit 2 leaving the old zip and exiting 2, a usage error — or if the Go output differs from the bash's at `8c54d6d3`.
**Baseline:** before=270 after=271
<!-- measured: cat stats/internal/guard/*_test.go | grep -cE '^func Test[A-Za-z0-9_]*\(t \*testing\.T\)' @ 8c54d6d3 -->
**After:** Task 1
**Commit:** `refactor(flow-guard): port rebuild-full-app-suite-zip to Go`
**Build:** green

**Decision:** guards-in-flow-guard
**Decision:** bash-at-8c54d6d3-is-the-oracle
**Decision:** zip-via-exec

  - [x] **Step 1: Record the oracle.** `FLOW_GUARD_CACHE_DIR=$(mktemp -d) bash scripts/test-rebuild-full-app-suite-zip.sh` on the untouched tree — exit 0.
  - [x] **Step 2: Port.** Write `rebuildfullappsuitezip.go`: `Registry["rebuild-full-app-suite-zip"]`, header as doc comment, `resolve-visual-screenshots.sh` exec'd from `FLOW_GUARD_SCRIPT_DIR`, mtimes pinned with `os.Chtimes` to `time.Date(1980,1,1,0,0,0,0,time.Local)`, `zip -q -X -j -@` exec'd with the list on stdin, exit codes as the bash's.
  - [x] **Step 3: Shim.** Replace the script per the template, with `: "$SCRIPT_DIR/resolve-visual-screenshots.sh"`; re-run Step 1's command against the shim — exit 0.
  - [x] **Step 4: Go test.** Write `rebuild_full_app_suite_zip_test.go`: `TestRebuildFullAppSuiteZip`, one `t.Run` per harness case, each asserting the harness's expectation and byte-for-byte stdout/stderr/exit parity with the bash at `8c54d6d3` (its `resolve-visual-screenshots.sh` sibling the live shim); the byte-identical case compares the Go zip with the bash zip too. Delete `scripts/test-rebuild-full-app-suite-zip.sh`.
  - [x] **Step 5: Break and prove.** Drop the `os.Chtimes` call, see the byte-identical case fail, restore it.
  - [x] **Step 6: Verify.** `cd stats && gofmt -l ./internal/guard && go vet ./internal/guard && go test ./internal/guard -run 'TestRebuildFullAppSuiteZip' -race -count=1`; `bash scripts/check-guard-symlinks.sh` and `bash scripts/check-references.sh` exit 0.

- [x] 3. Port run-verify-list.sh to flow-guard on rgtRun

**Files:** `stats/internal/guard/runverifylist.go`, `stats/internal/guard/run_verify_list_test.go`, `stats/internal/guard/runguardtests.go`, `scripts/run-verify-list.sh`, `scripts/test-run-verify-list.sh`
**Tests:** `TestRunVerifyList`
**Regression:** fails if any case of the deleted `scripts/test-run-verify-list.sh` regresses — list parsing and `# inputs:` lines, the template refusal (exit 2), exit 3 on an undeclared list, the 20-line truncation, the data-race line, `full output:` logs under `flow-verify-failed`, the green-sha write and skips, concurrent `test` commands reported in list order, `lint`'s appended `check-spec-reach.sh` — or if the Go output differs from the bash's at `8c54d6d3`; and `TestRunGuardTests*` fail if `rgtRun`'s argv change alters `run-guard-tests`.
**Baseline:** before=271 after=272
<!-- measured: cat stats/internal/guard/*_test.go | grep -cE '^func Test[A-Za-z0-9_]*\(t \*testing\.T\)' @ 8c54d6d3 -->
**After:** Task 2
**Commit:** `refactor(flow-guard): port run-verify-list to Go`
**Build:** green

**Decision:** guards-in-flow-guard
**Decision:** bash-at-8c54d6d3-is-the-oracle
**Decision:** verify-list-on-rgtrun

  - [x] **Step 1: Record the oracle.** `FLOW_GUARD_CACHE_DIR=$(mktemp -d) bash scripts/test-run-verify-list.sh` on the untouched tree — exit 0.
  - [x] **Step 2: rgtRun takes argv.** Change `rgtRun(ctx, harnesses []string, …)` to take `[][]string` argv, `exec.CommandContext(ctx, argv[0], argv[1:]...)`; `runGuardTests` builds `{"bash", h}` per harness. `go test ./internal/guard -run 'TestRunGuardTests' -count=1` stays green.
  - [x] **Step 3: Port.** Write `runverifylist.go`: `Registry["run-verify-list"]`, header as doc comment; `project-get.sh` and `check-spec-reach.sh` exec'd from `FLOW_GUARD_SCRIPT_DIR`; the fence/`# inputs:` parse as the awk's line tests; `test` commands through `rgtRun` with `{"bash","-c",cmd}` and `jobs = len` of the non-skipped commands, `Dir` the worktree; `lint` commands one at a time; reporting, truncation, race line, logs and green sha exactly as the bash.
  - [x] **Step 4: Shim.** Replace the script per the template, with `: "$SCRIPT_DIR/project-get.sh" "$SCRIPT_DIR/check-spec-reach.sh"`; re-run Step 1's command against the shim — exit 0.
  - [x] **Step 5: Go test.** Write `run_verify_list_test.go`: `TestRunVerifyList`, one `t.Run` per harness case, each asserting the harness's expectation and byte-for-byte parity with the bash at `8c54d6d3` (siblings the live shims), with the temp `run-verify-list.XXXXXX` and worktree paths normalised the same way on both sides. Delete `scripts/test-run-verify-list.sh`.
  - [x] **Step 6: Break and prove.** Change the truncation threshold to 21 in `runverifylist.go`, see the test fail, restore it.
  - [x] **Step 7: Verify.** `cd stats && gofmt -l ./internal/guard && go vet ./internal/guard && go test ./internal/guard -run 'TestRunVerifyList|TestRunGuardTests' -race -count=1`; `bash scripts/check-guard-symlinks.sh` and `bash scripts/check-references.sh` exit 0.

Correction (2026-10-22): Step 6's mutation (the threshold raised by one) survived the harness's cases — its only long output truncates under either threshold. `TestRunVerifyList` gained a subtest one line past the threshold, beyond the harness, which the mutation fails.

- [x] 4. Delete lib/resolve-file.sh and lib/spec-root.sh

**Files:** `scripts/lib/resolve-file.sh`, `scripts/lib/spec-root.sh`, `stats/internal/guard/libtwins_test.go`
**Allowed-collateral:** `stats/**/*.go`, `scripts/*.sh`
**Tests:** none — repairs existing tests: `TestResolveFileParity`, now reading `resolve-file.sh` at `8c54d6d3`.
**Regression:** `TestResolveFileParity` fails if `resolveFile` drifts from the bash `resolve_file` as it stood at `8c54d6d3`; `check-references.sh` fails on any path left citing a deleted file.
**Baseline:** before=272 after=272
<!-- measured: cat stats/internal/guard/*_test.go | grep -cE '^func Test[A-Za-z0-9_]*\(t \*testing\.T\)' @ 8c54d6d3 -->
**After:** Task 3
**Commit:** `refactor(scripts): delete the unsourced resolve-file and spec-root libraries`
**Build:** green

**Decision:** delete-dead-libs

  - [x] **Step 1: Confirm no loader.** `git grep -nE '(\.|source) .*lib/(resolve-file|spec-root)\.sh' -- scripts skills hooks setup.sh stats` prints nothing.
  - [x] **Step 2: Repoint the test.** In `libtwins_test.go`, `TestResolveFileParity` writes `historicalFile(t, "../../..", "8c54d6d3", "scripts/lib/resolve-file.sh")` into a temp dir and sources that copy (`bashRun` on its path), the doc comment naming the ref.
  - [x] **Step 3: Delete and update comments.** `git rm scripts/lib/resolve-file.sh scripts/lib/spec-root.sh`; every comment in `stats/**/*.go` and `scripts/*.sh` naming either as a live file (`resolvefile.go`, `specroot.go`, `changeplan.go`, `gatherdispatch.go`, `helpers_test.go`, `stats/cmd/flow/tasks.go`, `stats/internal/selfreview/git.go`, `scripts/aside-planning-artifacts.sh`, `scripts/check-task-commit-planning-paths.sh`) says it is retired and names its Go twin.
  - [x] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./... && go test ./internal/guard -run 'TestResolveFileParity|TestSpecRoot' -count=1`; `bash scripts/check-references.sh` and `bash scripts/check-guard-symlinks.sh` exit 0.
  - [x] **Step 5: Absence sweep.** `git grep -n -F -e 'scripts/lib/resolve-file.sh' -e 'scripts/lib/spec-root.sh' -- scripts skills hooks setup.sh` prints nothing.
