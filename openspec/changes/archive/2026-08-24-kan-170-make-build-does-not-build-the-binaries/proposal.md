## Why

`stats/Makefile`'s `build` target runs bare `go build ./...`. With three main packages Go discards
every binary, so `stats/bin/myflowd` and `stats/bin/myflow` are never written and anyone following
the documented build runs a stale binary with no indication. Hit in two separate sessions during
KAN-16; the working recipe already exists one target away, inside `restart:`.

Separately, three `myflowd` processes outlived their sessions during KAN-16. One ran for 4h40m on a
pre-fix binary, harvesting real session transcripts into the development database, and was found
only because it held `:4173` and made a later start fail with a bare `bind: address already in
use`. Nothing prevents a second daemon, and the collision is discovered last: `run()` opens the
store, migrates, seeds pricing and starts the harvest and sweep goroutines *before* it binds, so a
colliding start writes to the live database before it gives up.

## What Changes

- `make build` produces `bin/myflowd` and `bin/myflow` explicitly, keeping `go build ./...` ahead of
  them so `cmd/uitest-seed` and every library stay compiled by the target.
- `stats/README.md` and `.myflow/project.md`'s `## run` section stop documenting the explicit
  `go build -o` workaround alongside `make build` and name `make build` alone.
- `myflowd` writes a pidfile at `$TMPDIR/myflowd-<port>.pid`, recording its pid and its own
  executable path, and refuses to start when that file names a live process whose name matches the
  recorded executable. A dead pid, or a pid recycled onto an unrelated process, is a stale file it
  logs, overwrites and starts past. **On macOS `$TMPDIR` is a per-user directory under
  `/var/folders/`, not `/tmp`** — which is why `ui-test-up` pins `TMPDIR=/tmp` for the daemon it
  starts, so the Makefile can name that file as a literal.
- `myflowd` binds its listener before it opens the store, so a refused start touches no database.
- `ui-test-up` stops writing its own pidfile; `UITEST_PIDFILE` becomes the daemon-written
  `/tmp/myflowd-$(UITEST_PORT).pid`, still `override`, with `TMPDIR=/tmp` pinned on the daemon launch
  so that literal is the file the daemon actually writes.
- `restart:`'s "NO PIDFILE, deliberately" comment is rewritten — its premise no longer holds — while
  its kill stays port-driven.

## Capabilities

### New Capabilities

- `myflow-stats-build`: the stats build target produces the runnable binaries it documents, rather
  than compiling and discarding them.
- `myflow-daemon-single-instance`: the daemon refuses to start beside a live one, and reaches that
  refusal before it touches the database.

### Modified Capabilities

*(none)* — `ui-test-up` swapping its own pidfile for the daemon-written one changes which file the
tear-down reads, not any requirement `myflow-ui-test-stack` states. Its coexistence requirement is
covered from the other side, by `myflow-daemon-single-instance`'s two-daemons-on-different-ports
scenario.

## Impact

- `stats/Makefile` — the `build` target; `ui-test-up`/`ui-test-down`'s `UITEST_PIDFILE`; `restart:`'s
  pidfile comment.
- `stats/internal/pidfile/` — new package: path derivation, acquire, release, stale detection.
- `stats/cmd/myflowd/main.go` — `run()`'s ordering, and the pidfile acquire/release.
- `scripts/test-make-build.sh` — new harness, added to `.myflow/project.md`'s `## test`.
- `stats/README.md`, `.myflow/project.md` — the build and run documentation.
- No migration, no schema change, no new dependency.
