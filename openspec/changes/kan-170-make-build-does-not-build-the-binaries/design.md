## Context

Brainstorming record: `docs/superpowers/specs/2026-08-24-kan-170-make-build-does-not-build-the-binaries-design.md`.
Why the two defects matter: `proposal.md`.

## The build target

```make unverified:confirm GNU Make accepts the aligned two-space run-in on the second -o line as written
build: web-build
	go build ./...
	go build -o bin/myflowd ./cmd/myflowd
	go build -o bin/myflow  ./cmd/myflow
```

`go build ./...` stays, first, so the target's compile-all coverage is not traded for the two
artifacts. `stats/bin/` is already gitignored.

The regression guard is `make -n build`: the recipe must name both `-o` outputs. A recipe that
compiles and discards again is exactly what `-n` shows.

## `internal/pidfile`

New package, no dependencies outside the standard library.

```go verified:the single-Acquire shape this replaced was implemented and reviewed; the split below is what the ordering above requires
package pidfile

// Path returns the pidfile path for a resolved daemon port.
func Path(port int) string

// Check refuses with ErrAlreadyRunning when path names a live process whose
// name matches the recorded executable's basename. It writes nothing.
func Check(path string, logger *slog.Logger) error

// Write records this process at path and returns a Lock.
func Write(path string) (*Lock, error)

// Release removes the pidfile, and only when it still records this lock's
// own pid. It logs at warn on each path where it leaves the file alone.
// It is a no-op once called.
func (l *Lock) Release(logger *slog.Logger) error

var ErrAlreadyRunning = errors.New("pidfile: a daemon is already running")
```

**The check and the write are two calls, and the listener is opened between them.** A single
`Acquire` doing both before the bind was implemented and reviewed, and it is unsafe: the write race
and the bind race are independent, so the daemon that wins the write can be the one that loses the
bind — and its deferred `Release`, correctly pid-matched, then deletes the pidfile of the daemon that
actually holds the port. Writing only after a successful bind closes that window outright: a
bind-loser has written nothing and has nothing to release.

**Path.** `filepath.Join(os.TempDir(), "myflowd-<port>.pid")`. Port-derived so the live stack (4173)
and the UI-test stack (4174) never contend for one file.

**`os.TempDir()` is not `/tmp` on macOS** — it is a per-user directory under `/var/folders/`,
measured as `/var/folders/33/1zg47r2n0zbfdtnc1yg43ych0000gn/T/` on this machine. It is stable per
user, so a daemon started from a terminal, from `make restart`, and from the launchd agent all
resolve the same path and see each other's file. What it is not is a path the Makefile can hardcode,
which is why `ui-test-up` pins `TMPDIR=/tmp` on its daemon launch.
<!-- measured: cd stats && TMPDIR= go run ./cmd/... printing os.TempDir() @ branch openspec/kan-170-make-build-does-not-build-the-binaries -->

**Contents.** Two lines: the pid, then `os.Executable()`'s result. A file that does not parse into
those two lines is a stale file.

**Liveness.** `syscall.Kill(pid, 0)` for existence, then the process's name compared against the
recorded executable's basename. The recorded path — rather than a hardcoded `"myflowd"` — is what
makes the check fire for the UI-test daemon, which runs as `/tmp/myflow-uitest-myflowd`.

The name lookup shells to `ps -o comm= -p <pid>`: macOS has no `/proc`, and `ps` answers on both
platforms this repository is built on. It sits behind an unexported package variable so tests
substitute it rather than spawning real processes.

**Release.** On graceful shutdown, and only when the file still records this lock's own pid. A
`SIGKILL`ed daemon leaves the file; the liveness check is what makes that a stale file rather than a
permanent lockout. Both paths where `Release` leaves the file alone — it no longer parses, or it
names another daemon — log at warn, mirroring `Check`, so a takeover is visible in the shutdown log.
An absent file stays silent, since that is the second call of `run()`'s unconditional defer.

**An unreadable file is stale too.** `Check` returns a refusal and nothing else: a file it cannot
read tells it nothing about whether another daemon is running, and propagating that error would
abort every start until someone cleared the file by hand — the lockout the stale-file policy exists
to rule out.

## `run()`'s ordering

```text verified:read off stats/cmd/myflowd/main.go line by line — acquireStartup at 277, api.New at 163, reconcile.New at 177, the goroutines at 211 and 221, srv.Serve at 226
config.FromEnv()
cfg.Validate()          ← non-loopback host refused, still before any listener
pidfile.Check(...)      ← refuses here, naming the pid that holds it
net.Listen("tcp", ...)  ← refuses here, before the database is touched
pidfile.Write(...)      ← only the process that holds the port writes
defer lock.Release(); defer ln.Close()
store.Open / RunMigrations / SeedPricing
web.FS / web.Handler / api.New(...)
reconciler + harvest + sweep goroutines
srv.Serve(ln)
```

**Only the first six lines are this change's own.** Everything from `store.Open` down keeps the
order it already had; it is reproduced here so the block is a true account of `run()` rather than of
this change's diff alone. In particular `api.New` runs **before** the reconciler and the background
goroutines, not after them — an earlier draft of this block and of task 3's plan said otherwise, and
a reviewer caught that the `verified:` tag was attached to an ordering nobody had read off the file.

`api.Server.Serve(net.Listener)` already exists beside `ListenAndServe`, so handing it a
pre-opened listener needs no change to `internal/api`. `cfg.Validate()` is called explicitly
because `api.New` — which calls it today — now runs after the listener is open.

## Decisions

### The check and the write are separate calls, with the bind between them

**ID:** split-check-and-write
**Status:** active
**Chosen:** `Check` before `net.Listen`, `Write` after it.
**Considered:** a single `Acquire` doing both before the bind — implemented and reviewed first, and
rejected once its reviewer reproduced the disarm: the write race and the bind race are independent,
so a bind-loser can be the write-winner, and its own pid-matched `Release` then deletes the live
daemon's pidfile. Closing that with `O_CREATE|O_EXCL` covers only the absent-file half and leaves the
stale-takeover half open; a real lock would need `flock`. Ordering the two calls around the bind
needs neither.

### Which stray-process mechanism

**ID:** stray-process-mechanism
**Status:** active
**Chosen:** a pidfile `myflowd` writes itself, refusing to start when a live one names a running
myflowd — it prevents the second daemon rather than only explaining the collision afterwards.
**Considered:** reporting the holding pid and its start time on `EADDRINUSE` — cheaper and
stateless, but it diagnoses a daemon that has already started rather than stopping it; both
together — two mechanisms for one failure, and the pidfile's known weakness is not repaired by
adding a second mechanism beside it.

### Who writes the pidfile

**ID:** pidfile-writer
**Status:** active
**Chosen:** `myflowd` itself, so every start leaves one however it was started.
**Considered:** the Makefile, as `ui-test-up` does today — rejected because `restart:`'s own
recorded objection applies to it exactly: a hand-started daemon writes no Makefile pidfile, so the
file is absent precisely when it is needed.

### Pidfile path

**ID:** pidfile-path
**Status:** superseded by pidfile-path-tmpdir
**Chosen:** `/tmp/myflowd-<port>.pid`, derived from the resolved port.
**Considered:** a fixed `/tmp/myflowd.pid` — would make the live and UI-test daemons refuse each
other, breaking `make ui-test-up` while the live stack runs; a `MYFLOWD_PIDFILE` environment
override — configurability nobody asked for, and a caller-nameable path feeding a `kill` is the
hazard the Makefile's `UITEST_*` `override` block exists to prevent.

### Pidfile path — the temporary directory is `$TMPDIR`, not `/tmp`

**ID:** pidfile-path-tmpdir
**Status:** active
**Chosen:** `filepath.Join(os.TempDir(), "myflowd-<port>.pid")`, with `ui-test-up` pinning
`TMPDIR=/tmp` on its daemon launch so the Makefile can name that one file as a literal.
**Considered:** hardcoding `/tmp` in `Path` — matches every other literal path in `stats/Makefile`
and would need no pin, but the package's own tests would then write into the machine's real `/tmp`,
and a second unexported override just for them buys back the complexity the pin avoids;
`override UITEST_PIDFILE := $(TMPDIR)/myflowd-$(UITEST_PORT).pid` — re-exposes the hazard the
`override` block exists to prevent, since `make ui-test-down TMPDIR=<anything>` would then aim a
`kill` and an `rm -f` at a caller-named path, and it expands to an empty prefix when `TMPDIR` is
unset while Go still falls back to `/tmp`.
**Why the original was wrong:** decision `pidfile-path` said `/tmp/myflowd-<port>.pid` in prose while
the plan said `os.TempDir()` in code. On macOS those are different directories, so the Makefile's
literal named a file the daemon never wrote, and every `[ -f $(UITEST_PIDFILE) ]` guard would have
taken its "no daemon" branch — leaving `ui-test-down`'s unforced `dropdb` racing a live daemon,
silently and intermittently.

### Stale-pidfile policy

**ID:** stale-pidfile-policy
**Status:** active
**Chosen:** verify identity, then take over — refuse only when the pid is alive **and** its process
name matches the recorded executable's basename.
**Considered:** refusing on any live pid (`kill -0` alone) — a recycled pid on an unrelated process
would make `myflowd` unstartable until the file was deleted by hand.

### Bind order

**ID:** bind-before-store
**Status:** active
**Chosen:** bind the listener before opening the store, and hand it to `srv.Serve`.
**Considered:** leaving the bind last and only improving the error text — a smaller diff, but a
colliding start still migrates, seeds and begins harvesting into the live database before failing.

### `go build ./...` in the build target

**ID:** keep-compile-all
**Status:** active
**Chosen:** keep it, ahead of the two `-o` builds.
**Considered:** the two `-o` builds alone — smallest recipe, but it drops `cmd/uitest-seed` from the
target's coverage; adding `bin/uitest-seed` as a third artifact — `ui-test-up` invokes it via
`go run`, so the binary would have no consumer.

### `restart:` keeps its port-driven kill

**ID:** restart-kill-stays-port-driven
**Status:** active
**Chosen:** leave `restart:` finding the daemon by `lsof -ti tcp:$(LIVE_PORT)`, and rewrite only its
"NO PIDFILE, deliberately" comment.
**Considered:** switching the kill to the new pidfile — rejected on the target's own recorded
reasoning, which this change does not overturn: a `kill` driven by a caller-nameable file is the
hazard, and one driven by "who holds this port" has no such handle.

## Open questions

### `run()` has no unit test, so three of its guarantees rest on review alone

**ID:** run-has-no-unit-test
**Status:** open
**Why it is open:** `run()` opens a database and starts three background loops, so it cannot be
exercised without a live PostgreSQL instance. Every test this change adds goes through
`acquireStartup` instead, which is why the prelude's own ordering is pinned but `run()`'s use of it
is not.
**What it affects:** three mutations survive the suite — dropping `defer lock.Release()`, dropping
`defer ln.Close()`, and reverting `srv.Serve(ln)` to `srv.ListenAndServe()`. The third is a real
regression if it ever happens: a fresh, unclaimed listener would bypass the port `acquireStartup`
just secured. It was caught by the compiler before `defer ln.Close()` existed — `ln` was then
otherwise unused — so adding that defer removed the only thing standing against it. Closing this
needs a `run()` that can be driven against a test database, which is a larger change than this one.
