## Context

`scripts/integrate-change.sh` at `a03f3c4b` is the specification: its header is the contract (four
subcommands, their flags, the `NEXT:`/`STOP:`/`JIRA:` closing line, exit 0/1/2) and its body is the
behaviour. `scripts/test-integrate-change.sh` at the same commit drives the real script and the
real guards over a two-repository fixture, stubbing only the external `flow` and `spectre` CLIs on
`PATH`; its cases are the behaviour the port must keep.

Findings that shape the plan:

- Every guard the script sequences is a `flow_guard_exec` shim except `run-verify-list.sh` and
  `check-narrative-runs.sh`, which are bash.
- check-guard-symlinks rule 2 learns a guard's siblings from `$SCRIPT_DIR/<name>` spellings in the
  guard's own file; today each call site spells one. The precedent for a Go guard that execs a
  sibling is `check-finish-preflight.sh`: the shim exports the sibling's path and the Go execs it.
- `lib/spec-root.sh`'s `spec_root_leaf` has a Go counterpart, `specroot.go`.
- Live mentions of `integrate-change.sh` (skills, `stagemarkcalls.go`, `unfinishedwork.go`, the
  `check-stage-mark-calls.sh` header) name the shim, which keeps its path, so none changes.
  `test-integrate-change.sh` is named live only by itself and by the open change
  `spectre/changes/faster-integrate/proposal.md`, another change's artifact, left as it is.

## Decisions

### Where the port lives

**ID:** integrate-change-in-flow-guard
**Status:** active
**Chosen:** `Registry["integrate-change"]` in `stats/internal/guard/`, one file for the shared
helpers and argument parsing (`integratechange.go`) and one per subcommand
(`integratechange_prepare.go`, `integratechange_commit.go`, `integratechange_land.go`,
`integratechange_cleanup.go`) — the bash already falls into those four bodies, and one file of the
whole would be the package's largest by far.
**Considered:** a single `integratechange.go` — rejected, a file nobody can review in one sitting;
a separate `stats/cmd/integrate-change` binary — rejected, a second build path and cache for one
caller (kan-921's `runner-in-flow-guard`).

### How the sequenced guards run

**ID:** siblings-execd-from-shim-dir
**Status:** active
**Chosen:** each sequenced guard is exec'd as its sibling script, `<dir>/<guard>.sh`, where `<dir>`
is `FLOW_GUARD_SCRIPT_DIR`, exported by the shim as its own `$SCRIPT_DIR`. The shim keeps the
sibling-presence check, spelling each sibling `"$SCRIPT_DIR/<name>.sh"` once so check-guard-symlinks
rule 2 still requires every one, with its missing-sibling message and exit 2 unchanged. Their
stdout is relayed as the bash's `run` did, their stderr passed straight through.
**Considered:** calling the ported guards' `Registry` functions in process — rejected: two of them
are bash, each shim sets its own `FLOW_GUARD_*` environment the Go would have to replicate, and
the script's contract is that it re-implements no guard; exporting one path per sibling, as
`check-finish-preflight.sh` does — rejected, many exports for one directory.

### Working directory of every child

**ID:** child-dir-root
**Status:** active
**Chosen:** every child process runs with `Dir: "/"` — what the bash's `cd /` gave every command
after it, so `remove-change-worktrees`' check 6 never counts this process's cwd — save the calls
the bash ran in a subshell `cd` (`spectre archive`, `gh pr create`, the workspace rows), which keep
their own directory. The Go process never changes its own working directory.
**Considered:** `os.Chdir("/")` — rejected, process-wide state an in-process test would inherit.

### The parity oracle

**ID:** bash-harness-is-the-oracle
**Status:** active
**Chosen:** the port's commit leaves `scripts/test-integrate-change.sh` unchanged and passing
against the shim — the harness exercises the script by its path, so it then runs the Go — which is
the byte-for-byte proof the ticket asks for, made with the bash at `a03f3c4b`'s own assertions. Only
once the Go tests carry every case is the harness deleted.
**Considered:** capturing golden output strings per case (kan-921's `coverage-golden-strings`) —
rejected, the fixture's paths and shas are per-run temp values, so goldens would need a
normalisation layer the harness's own assertions already make unnecessary.

### Shape of the Go tests

**ID:** go-tests-in-process-two-files
**Status:** superseded by go-tests-bash-session-sut-in-process
**Chosen:** the harness's fixture becomes a Go helper building the same two-repository fixture
with the same `flow` and `spectre` stubs on `PATH`; each case calls `Registry["integrate-change"]`
in process with an env carrying the stubs' `PATH` and `FLOW_GUARD_SCRIPT_DIR` pointed at the real
`scripts/`, and runs under `t.Parallel()` as the harness ran each case in its own subshell. The
cases split across `integrate_change_test.go` (fixture, prepare and commit cases) and
`integrate_change_land_test.go` (land and cleanup cases), so two implementers can port them side
by side.
**Considered:** running the shim as a subprocess — rejected, every case would pay the shim's
build-cache check and the test would not exercise the code under `go test -cover`.
**Superseded because:** implementation found a Go transcription of the harness's fixture and case bodies would restate about 1000 lines of setup the bash already proves.

### Shape of the Go tests, as shipped

**ID:** go-tests-bash-session-sut-in-process
**Status:** active
**Chosen:** each case is a `t.Run` with `t.Parallel()` running the harness's own fixture, helpers, case body and `check` expressions verbatim as one bash session; its `sut` hands argv and environment back to the test, which runs `Registry["integrate-change"]` in process. Each `check` label is a subtest. The cases split across `integrate_change_test.go` (driver and happy path), `integrate_change_prepare_test.go` and `integrate_change_land_test.go`.
**Considered:** transcribing every setup step and assertion into Go (go-tests-in-process-two-files) — rejected, a hand restatement with a new place for each to drift; the SUT runs in process either way.

### Where the contract text lives

**ID:** header-moves-to-doc-comment
**Status:** active
**Chosen:** the bash header's contract (usage, each subcommand's paragraph, the output and exit
contract) moves into `integratechange.go`'s doc comment verbatim; the shim's header keeps the usage
lines and points at it, as `run-guard-tests.sh` does.
**Considered:** keeping the full header in the shim — rejected, it would describe code that no
longer lives beside it.

## Open questions

## Live check

Run the shim's `prepare` on this change from its own worktree at one HEAD twice — once with the
bash at `a03f3c4b` (`git show a03f3c4b:scripts/integrate-change.sh` saved beside the scripts) and
once through the shim — recording each run's full output and exit code; they must be identical.
Then this change's own integrate runs through the shim — prepare, commit, land — the port carrying
a real change end to end. Not working looks like a shim build error, an exit 2 naming a missing
sibling, any line or exit code differing between the two prepare runs, or a stop the bash would
not have made.
