## Context

KAN-1020, the next slice of the bash-to-Go guard port (KAN-760, KAN-841, KAN-842, KAN-919, KAN-921, KAN-1019).
New guard logic is Go in flow-guard, bash only as a `flow_guard_exec` shim. The guard-test runner
(`stats/internal/guard/runguardtests.go`) accepts a ported `check-*.sh` without a bash companion only when its
file carries `flow_guard_exec <name> ` and `stats/internal/guard/<name, - as _>_test.go` exists.

`git grep` at `8c54d6d3` finds no script that sources `lib/resolve-file.sh` or `lib/spec-root.sh`: the remaining
references are comments and three tests — `libtwins_test.go`'s `TestResolveFileParity` sources the live
`resolve-file.sh`; `aside_planning_artifacts_test.go`, `check_workspace_isolation_test.go` and
`check_guard_symlinks_test.go` already read their copies at historical refs.

## Decisions

### Port each script into flow-guard behind a shim

**ID:** guards-in-flow-guard
**Status:** active
**Chosen:** one `stats/internal/guard/<name>.go` per script registering `Registry["<name>"]`, the bash header's
contract moved into its doc comment; the script becomes a shim in the shape of `scripts/check-spec-reach.sh`
(header pointing at the Go file, `flow_guard_exec <name> 2 "<name>:" "$@"`), keeping its `$SCRIPT_DIR/<sibling>.sh`
mentions so `check-guard-symlinks.sh` rule 2 still requires each sibling it execs.
**Considered:** keeping the bash — rejected by the ticket and by the new-guards-in-Go practice.

### The bash at 8c54d6d3 is the oracle

**ID:** bash-at-8c54d6d3-is-the-oracle
**Status:** active
**Chosen:** each task first runs the unchanged bash harness against the shim, then its Go test runs the bash
script read with `historicalFile(t, "../../..", "8c54d6d3", …)` and the Go guard over the same fixtures,
comparing stdout, stderr and exit byte for byte; the harness is then deleted.
**Considered:** hand-pinned expected strings only — rejected: a transcription can drift from what the bash did,
and every earlier port used the historical-ref parity shape.

### run-verify-list reuses rgtRun

**ID:** verify-list-on-rgtrun
**Status:** active
**Chosen:** `rgtRun` takes one argv per job in place of a harness path (`run-guard-tests` passes `bash <harness>`);
`run-verify-list`'s `test` list passes `bash -c <command>` per non-skipped command with `jobs = len(commands)`,
so every command still runs at once as the bash `&` did; stdin is nil (`/dev/null`). The `lint` list stays one
command at a time.
**Considered:** a second pool inside `runverifylist.go` — rejected: the ticket names `rgtRun`, and a second
pool is the duplicate this port exists to avoid. Bounding the test list by `JOBS` — rejected: a behaviour change
the ticket did not ask for.

### rebuild-full-app-suite-zip keeps exec'ing zip

**ID:** zip-via-exec
**Status:** active
**Chosen:** pin each PNG's mtime with `os.Chtimes` to 1980-01-01 00:00 local time (what `touch -t 198001010000`
sets) and exec `zip -q -X -j -@ <zip>` with the list on stdin.
**Considered:** `archive/zip` — rejected: its local headers and extra fields differ from Info-ZIP's, so the archive
bytes change from the bash's and the first rebuild after landing commits a different zip over unchanged images.

### check-narrative-runs execs the flow CLI

**ID:** narrative-runs-exec-flow
**Status:** active
**Chosen:** exec `flow stage sessions -C <wt> -stage <key> <change>` from PATH, as the bash did, so the harness's
`flow` stub drives the Go test unchanged.
**Considered:** an `Env` hook — rejected: no second caller, and the stub-on-PATH fixture already isolates it.

### Delete both libraries

**ID:** delete-dead-libs
**Status:** active
**Chosen:** delete `scripts/lib/resolve-file.sh` and `scripts/lib/spec-root.sh`; `TestResolveFileParity` reads
`resolve-file.sh` at `8c54d6d3` into a temp dir and sources that copy; every comment naming either as a live file
says it is retired and names its Go twin.
**Considered:** porting them — nothing to port: `resolvefile.go` and `specroot.go` already are the ports and no
script sources the bash.

## Open questions

## Live check

None — the change ports guard scripts with byte-for-byte parity proven by tests; no running service or persistent
state is touched.
