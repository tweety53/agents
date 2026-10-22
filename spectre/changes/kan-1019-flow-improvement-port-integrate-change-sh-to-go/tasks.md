# kan-1019-flow-improvement-port-integrate-change-sh-to-go

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** yes — the bash header's contract text moves verbatim into `integratechange.go`'s doc comment
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

Ports `scripts/integrate-change.sh` into `flow-guard` behind a shim, proven by the unchanged bash
harness (task 1); builds the Go fixture (task 2); ports the harness's prepare/commit cases (task 3)
and land/cleanup cases (task 4); deletes the harness (task 5). `design.md` is canonical for every
decision; each task cites its entry by ID.

**The scripts at `a03f3c4b` are the specification.** `scripts/integrate-change.sh`'s header is the
contract and its body the behaviour; `scripts/test-integrate-change.sh` holds the cases. The port
reproduces every stdout and stderr line byte for byte, the stdout/stderr split, every stage mark
(key and outcome, `stopped`/`not-run-2`/`leftover` included), every state write and every exit
code. Where the script and its header disagree, stop and report — never pick one silently.

**Shim template** — a header keeping the four usage lines and pointing at `integratechange.go`, then:

```bash unverified:confirm the sibling list matches integrate-change.sh's `missing` loop plus every other "$SCRIPT_DIR/<name>.sh" it calls (prepare-workspace, check-loop-landing, check-narrative-runs) at a03f3c4b
set -uo pipefail
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "integrate-change: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
missing=""
for g in "$SCRIPT_DIR/check-base-moved.sh" "$SCRIPT_DIR/check-cleanup-complete.sh" ...; do
  [ -x "$g" ] || missing="$missing ${g##*/}"
done
[ -z "$missing" ] || { printf 'integrate-change: %s\n' "missing beside this script:$missing — run the procedure by hand" >&2; exit 2; }
FLOW_GUARD_SCRIPT_DIR="$SCRIPT_DIR"
export FLOW_GUARD_SCRIPT_DIR
flow_guard_exec integrate-change 2 "integrate-change:" "$@"
```

The bash's presence loop checks a fixed list and its other calls (`prepare-workspace.sh`,
`check-loop-landing.sh`, `check-narrative-runs.sh`) unchecked; the shim's loop keeps exactly the
bash's list so the missing-sibling message is unchanged, and spells the other three as
`"$SCRIPT_DIR/<name>.sh"` in a comment-free line rule 2 reads (an `: "$SCRIPT_DIR/…"` no-op line
per name), so rule 2 still requires them.

**Test isolation:** every Go test calls `t.Parallel()`; fixture repositories are built per case in
`t.TempDir()`; `PATH` (the `flow` and `spectre` stubs first), `FLOW_GUARD_SCRIPT_DIR` (this
checkout's real `scripts/`), `FLOW_GUARD_CACHE_DIR` (a per-test temp dir) and
`FLOW_GUARD_TELEMETRY=off` pass on `Env`, never through the process environment; no test touches
the operator's real `$HOME`, cache or this repository's tree.

## Review Focus

- Every `STOP:` closes the open stage mark with the bash's outcome before printing, and exits 1.
- Every child process runs with `Dir` `/`, save `spectre archive`, `gh pr create` and the
  workspace rows, which run in their own directory (`child-dir-root`).
- `land`'s merge-and-push route continues into `cleanup` in the same process, as `cmd_cleanup`
  was called from `cmd_land`.
- `visual_gate` stops at the first MISSING worktree; no later worktree is judged.
- Pushes keep `--force-with-lease` on the branch, plain on the pin and on `HEAD:<base>`; the pin
  on origin is never overwritten unless the local pin descends from it.

---

- [x] 1. Port integrate-change.sh to flow-guard behind a shim

**Files:** `stats/internal/guard/integratechange.go`, `stats/internal/guard/integratechange_prepare.go`, `stats/internal/guard/integratechange_commit.go`, `stats/internal/guard/integratechange_land.go`, `stats/internal/guard/integratechange_cleanup.go`, `scripts/integrate-change.sh`
**Tests:** **none** — the unchanged `scripts/test-integrate-change.sh` is this task's regression, run against the shim.
**Regression:** fails if any assertion of `scripts/test-integrate-change.sh` regresses when it drives the shim — any output line, exit code, stage mark, state write, commit, push or removal the bash made.
**Baseline:** before=0 after=0
<!-- measured: cat stats/internal/guard/integrate_change*_test.go 2>/dev/null | grep -cE '^func Test' @ a03f3c4b -->
**After:** none
**Commit:** `refactor(flow-guard): port integrate-change to Go`
**Build:** green

**Decision:** integrate-change-in-flow-guard
**Decision:** siblings-execd-from-shim-dir
**Decision:** child-dir-root
**Decision:** bash-harness-is-the-oracle
**Decision:** header-moves-to-doc-comment

  - [x] **Step 1: Record the oracle.** Run the harness on the untouched tree and keep its output:
    `FLOW_GUARD_CACHE_DIR=$(mktemp -d) FLOW_GUARD_TELEMETRY=off bash scripts/test-integrate-change.sh`
    — expect its summary line `215 passed, 0 failed`, exit 0.
    <!-- measured: FLOW_GUARD_CACHE_DIR=<tmp> FLOW_GUARD_TELEMETRY=off bash scripts/test-integrate-change.sh | tail -1 @ a03f3c4b -->
  - [x] **Step 2: Port.** Write `integratechange.go` (`Registry["integrate-change"]`, argument
    parsing exactly as the bash's `while` loop with its `die` messages and exit 2, the shared
    helpers — `run`/`verdict`, `mark_begin`/`mark_end`/`stop`, `load_set`, `main_of`,
    `main_of_gone`, `find_canonical` via `specroot.go`, `merge_order`, `base_of`, `arch_base`,
    `impl_paths`, `link_only`, `write_state`, `guard_tests`, `rebase_lint`, the visual gate,
    `push_branch`, `push_pin`, `pin_pre_reshape` — and the bash header's contract verbatim as its
    doc comment) and one file per subcommand. Read every sibling path from `FLOW_GUARD_SCRIPT_DIR`;
    an unset one is exit 2 naming the shim, as `finishpreflight.go` does for its variable. Keep
    each `jq` program's semantics in Go's `encoding/json`, each `awk`/`grep` filter as the same
    line test, and the `bash -c` workspace rows as `bash -c`.
  - [x] **Step 3: Shim.** Replace `scripts/integrate-change.sh` with the shim template above,
    the presence loop carrying the bash's exact list.
  - [x] **Step 4: Prove parity.** Re-run Step 1's command against the shim — the same summary
    line, exit 0. Any failing assertion is a port defect: fix the Go, never the harness.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l ./internal/guard && go vet ./internal/guard ./cmd/flow-guard`
    prints nothing and exits 0; `bash scripts/check-guard-symlinks.sh` and
    `bash scripts/check-references.sh` exit 0.

- [x] 2. Build the Go fixture and port the happy path

**Files:** `stats/internal/guard/integrate_change_test.go`
**Tests:** `TestIntegrateChangeHappyPath`
**Regression:** fails if prepare → commit → land on the merge-and-push route stops reaching
`FINISHED`: both origins' `main` carrying the implementation commit (a's also the planning and
archive commits), the worktrees removed, the record `FINISHED` with an empty map, the pin
surviving run 2, and the stage log running from `flow.preflight` to `flow.refresh-main-checkout`.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/integrate_change*_test.go 2>/dev/null | grep -cE '^func Test' @ a03f3c4b -->
**After:** Task 1
**Commit:** `test(flow-guard): port integrate-change's fixture and happy path to Go`
**Build:** green

**Decision:** go-tests-in-process-two-files

  - [x] **Step 1: Fixture.** Port `fixture`, `g`, `sut`, `to_land`, `halt`'s intent (a step that
    did not exit 0 fails the test with its stdout and stderr), `ui_a` and `flag` from
    `scripts/test-integrate-change.sh` into helpers in `integrate_change_test.go`: the same two
    repositories `a` and `b`, bare origins, worktrees at `<repo>-worktrees/demo`, the same
    `flow` and `spectre` stub scripts and state record, git's dates and identity pinned as `g`
    pins them. `sut` calls `Registry["integrate-change"]` in process with the isolation `Env`
    above, capturing stdout and stderr separately.
  - [x] **Step 2: Happy path.** Port `happy_path`, one `t.Run` per `check` label, labels kept.
  - [x] **Step 3: Break and prove.** Change one `NEXT:` line's text in
    `integratechange_prepare.go`, see the test fail, restore it.
  - [x] **Step 4: Verify.** `cd stats && gofmt -l ./internal/guard && go vet ./internal/guard && go test ./internal/guard -run 'TestIntegrateChangeHappyPath'`.

Correction (2026-10-22): the plan declared the fixture and each case's steps ported into Go helpers; what shipped keeps the harness's fixture, `g`, `to_land`, `peer_fix_case`, `ui_a`, `flag`, every case body and every `check` expression verbatim as bash, run per case as its own bash session whose `sut` hands argv and environment back to the test, which runs `Registry["integrate-change"]` in process. The SUT still runs in process under `t.Parallel()` and `go test -cover`, each `check` label is its own subtest, and no setup step or assertion was re-expressed by hand — a transcription into Go would have restated about 1000 lines of setup the bash already proves, with a new place for each to drift.

- [x] 3. Port the prepare and commit cases to Go

**Files:** `stats/internal/guard/integrate_change_prepare_test.go`
**Tests:** `TestIntegrateChangePrepareCommit`
**Regression:** fails if any prepare or commit behaviour the harness pinned regresses:
`narrative_format`, `narrative_runs`, `foreign_staged`, `untasked`, `archived_rerun`,
`no_verdict`, `sync_conflict`, `guard_test`, `rebase_lint`, `rebase_lint_env`,
`rebase_lint_rewrite`, `rebase_lint_dirty`, `nested_link`, `mixed_roots`, `planning_only`,
`satellite_archive`, `link_only_canon`, `commit_pushes`, `pre_reshape_pin`, `rerun_pin`,
`base_mismatch`, `outstanding`, `drift_switch`, `drift_switch_stops`, `drift_switch_refused`.
**Baseline:** before=1 after=2
<!-- measured: cat stats/internal/guard/integrate_change*_test.go 2>/dev/null | grep -cE '^func Test' @ a03f3c4b -->
**After:** Task 2
**Commit:** `test(flow-guard): port integrate-change's prepare and commit cases to Go`
**Build:** green

**Decision:** go-tests-in-process-two-files

  - [x] **Step 1: Port.** One table entry per case named in **Regression**, each a `t.Run` with
    `t.Parallel()` on its own fixture, one subtest per `check` label with the label kept verbatim,
    reusing task 2's helpers.
  - [x] **Step 2: Break and prove.** Remove the `--accept-outstanding` check in
    `integratechange_prepare.go`'s unfinished-work stop, see `untasked` fail, restore it.
  - [x] **Step 3: Verify.** `cd stats && gofmt -l ./internal/guard && go vet ./internal/guard && go test ./internal/guard -run 'TestIntegrateChangePrepareCommit'`.

- [x] 4. Port the land, cleanup and visual-gate cases to Go

**Files:** `stats/internal/guard/integrate_change_land_test.go`
**Tests:** `TestIntegrateChangeLandCleanup`
**Regression:** fails if any land or cleanup behaviour the harness pinned regresses: `no_report`,
`loop_landing`, `rebase_lint_land`, `gh_open_merged`, `disclose`, `rerun_leftover`,
`no_state_dir`, `pr_url`, `peer_fix`, `peer_fix_canon_first`, `visual_stop`, `visual_archived`,
`visual_rebased`, `visual_resync`, `visual_resync_reread`, `visual_reread`, `visual_uncommitted`,
`refresh_on_recorded_base`, `orphan_sweep`.
**Baseline:** before=2 after=3
<!-- measured: cat stats/internal/guard/integrate_change*_test.go 2>/dev/null | grep -cE '^func Test' @ a03f3c4b -->
**After:** Task 2
**Commit:** `test(flow-guard): port integrate-change's land and cleanup cases to Go`
**Build:** green

**Decision:** go-tests-in-process-two-files

  - [x] **Step 1: Port.** One table entry per case named in **Regression**, each a `t.Run` with
    `t.Parallel()` on its own fixture, one subtest per `check` label with the label kept verbatim,
    reusing task 2's helpers; `peer_fix_case` becomes a helper both peer-fix cases call.
  - [x] **Step 2: Break and prove.** Drop `record_fixed_rows`' call in
    `integratechange_land.go`, see `peer_fix` fail, restore it.
  - [x] **Step 3: Verify.** `cd stats && gofmt -l ./internal/guard && go vet ./internal/guard && go test ./internal/guard -run 'TestIntegrateChangeLandCleanup'`.

- [x] 5. Delete the bash harness

**Files:** `scripts/test-integrate-change.sh`, `stats/internal/guard/integrate_change_test.go`, `stats/internal/guard/integrate_change_prepare_test.go`, `stats/internal/guard/integrate_change_land_test.go`
**Tests:** none — repairs existing tests: deletes `scripts/test-integrate-change.sh`, whose cases tasks 2–4 carry.
**Regression:** none of its own; a case of the harness no Go test carries would be lost, which
Step 1 rules out.
**Baseline:** before=3 after=3
<!-- measured: cat stats/internal/guard/integrate_change*_test.go 2>/dev/null | grep -cE '^func Test' @ a03f3c4b -->
**After:** Task 3, 4
**Commit:** `test(scripts): drop integrate-change's bash harness for its Go tests`
**Build:** green

Correction (2026-10-22): the absence sweep found the three Go test files' header comments citing the deleted harness by path; this task's commit rewords them to name it as the former bash harness, so `**Files:**` widens to those three files.

  - [x] **Step 1: Coverage check.** Every `# --- <case> ---` marker of
    `scripts/test-integrate-change.sh` names a case in task 2's, 3's or 4's **Regression** field
    and a `t.Run` in the Go tests; a missing one is ported before the delete.
  - [x] **Step 2: Delete** `scripts/test-integrate-change.sh`.
  - [x] **Step 3: Absence sweep.** `git grep -n -F -e 'test-integrate-change' -- scripts stats skills rules .flow README.md`
    prints nothing.
  - [x] **Step 4: Verify.** `bash scripts/check-references.sh` and `bash scripts/check-guard-symlinks.sh` exit 0;
    `cd stats && go test ./internal/guard -run 'TestIntegrateChange'` passes.
