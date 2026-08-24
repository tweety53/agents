# KAN-170 — `make build` does not build the binaries, and myflowd processes are left running

Two operational papercuts from KAN-16's self-review, both in `stats/`.

## Problem 1 — `make build` refreshes nothing

`stats/Makefile`'s `build` target runs bare `go build ./...`. With three main packages
(`cmd/myflowd`, `cmd/myflow`, `cmd/uitest-seed`) Go discards every binary, so `stats/bin/myflowd`
and `stats/bin/myflow` are never written. Anyone following the documented build runs a stale
binary with no indication. Hit in two separate sessions during KAN-16.

The working recipe already exists one target away: `restart:` does the explicit
`go build -o bin/myflowd ./cmd/myflowd` and `go build -o $(LIVE_CLI) ./cmd/myflow`.

## Problem 2 — stray myflowd processes

Three myflowd processes outlived their sessions during KAN-16. One ran for 4h40m on a pre-fix
binary, harvesting real session transcripts into the development database, and was found only
because it held `:4173` and made a later start fail with a bare
`bind: address already in use`.

Two compounding factors:

1. **Nothing prevents a second daemon.** The launchd agent is deliberately not loaded, so every
   one of these was started by hand.
2. **The collision is discovered last.** `cmd/myflowd/main.go`'s `run()` opens the store, runs
   migrations, seeds pricing and starts the harvest + sweep goroutines *before* it calls
   `srv.ListenAndServe()`. A colliding start therefore migrates, seeds and begins scanning
   transcripts against the live database before it gives up.

## Design

### `make build` produces both binaries

```make
build: web-build
	go build ./...
	go build -o bin/myflowd ./cmd/myflowd
	go build -o bin/myflow  ./cmd/myflow
```

`go build ./...` is kept, first, so `cmd/uitest-seed` and every library still get compiled — the
target's compile-all coverage is not traded away for the two artifacts. `bin/` is already
gitignored (`stats/bin/`).

`stats/README.md` and `.myflow/project.md`'s `## run` section stop documenting the explicit
`go build -o` workaround and name `make build` instead, so the documented command and the working
command become the same command.

### myflowd owns a port-derived pidfile

`run()`'s order becomes:

```text
config.FromEnv()
  → pidfile check      ← refuses here, before anything else exists
  → net.Listen(addr)   ← refuses here, before the database is touched
  → write pidfile
  → store.Open / RunMigrations / SeedPricing
  → harvest + sweep goroutines
  → srv.Serve(ln)
```

Binding before opening the store is what makes a collision cheap: today it costs a migration run,
a pricing seed and a transcript scan against the live database before it fails.

**Path:** `/tmp/myflowd-<port>.pid`. Port-derived, not fixed, because the live stack (4173) and the
UI-test stack (4174) are explicitly designed to run at the same time; one shared file would make
`make ui-test-up` refuse to start whenever the live daemon was up.

**Contents:** the pid and the daemon's own `os.Executable()` path.

**Liveness check:** refuse only when the recorded pid is alive **and** its process name matches the
recorded executable's basename. The executable path is what makes this work for the UI-test daemon,
which runs as `/tmp/myflow-uitest-myflowd` rather than `myflowd` — a hardcoded `"myflowd"` match
would silently never fire for it. Anything else — a dead pid, or a pid recycled onto an unrelated
process — is a stale file: log it, overwrite it, start.

**Removal:** on graceful shutdown. A `SIGKILL`ed daemon leaves the file behind; the identity check
above is what makes that harmless rather than a permanent lockout.

### The UI-test stack reads myflowd's pidfile

`ui-test-up` drops its own `echo $! > $(UITEST_PIDFILE)`, and `UITEST_PIDFILE` becomes
`/tmp/myflowd-$(UITEST_PORT).pid` — still `override`, per the Makefile's own RULE that every
`UITEST_*` variable carries it. One writer, one file, no chance of the Makefile's copy and the
daemon's copy disagreeing.

### `restart:` keeps its port-driven kill

`restart:` continues to find the daemon by `lsof -ti tcp:$(LIVE_PORT)`, per its own recorded
reasoning that a `kill` driven by "who holds this port" has no caller-nameable handle. Its
"NO PIDFILE, deliberately" comment is rewritten: the premise — that a pidfile *this target* wrote
would be absent for a hand-started daemon — no longer holds now that the daemon writes its own on
every start, however it was started.

## Decisions

### Which stray-process mechanism

**ID:** stray-process-mechanism
**Status:** active
**Chosen:** a pidfile myflowd writes itself, refusing to start when a live one names a running
myflowd — it prevents the second daemon rather than only explaining the collision afterwards.
**Considered:** reporting the holding pid and its start time on `EADDRINUSE` — cheaper and
stateless, but it diagnoses a daemon that has already been started rather than stopping it; both
together — two mechanisms covering one failure, and the pidfile's known weakness is not repaired by
adding a second mechanism beside it.

### Who writes the pidfile

**ID:** pidfile-writer
**Status:** active
**Chosen:** myflowd itself, so every start leaves one however it was started.
**Considered:** the Makefile, as `ui-test-up` does today — rejected because `restart:`'s own
recorded objection applies to it exactly: a hand-started daemon writes no Makefile pidfile, so the
file is absent precisely when it is needed.

### Pidfile path

**ID:** pidfile-path
**Status:** active
**Chosen:** `/tmp/myflowd-<port>.pid`, derived from the resolved port.
**Considered:** a fixed `/tmp/myflowd.pid` — would make the live and UI-test daemons refuse each
other, breaking `make ui-test-up` while the live stack runs; a `MYFLOWD_PIDFILE` env override —
configurability nobody asked for, and a caller-nameable path feeding a `kill` is the exact hazard
the Makefile's `UITEST_*` `override` block exists to prevent.

### Stale-pidfile policy

**ID:** stale-pidfile-policy
**Status:** active
**Chosen:** verify identity, then take over — refuse only when the pid is alive and its process
name matches the recorded executable's basename.
**Considered:** refusing on any live pid (`kill -0` alone) — a recycled pid belonging to any
unrelated process would make myflowd unstartable until the file was deleted by hand.

### Bind order

**ID:** bind-before-store
**Status:** active
**Chosen:** bind the listener before opening the store, and hand it to `srv.Serve`.
**Considered:** leaving the bind last and only improving the error text — a smaller diff, but a
colliding start still migrates, seeds and begins harvesting into the live database before failing.

### `go build ./...` in the build target

**ID:** keep-compile-all
**Status:** active
**Chosen:** keep it, ahead of the two `-o` builds, so `cmd/uitest-seed` and every library stay
compiled by `make build`.
**Considered:** the two `-o` builds alone — smallest recipe, but it drops `cmd/uitest-seed` from
the target's coverage; adding `bin/uitest-seed` as a third artifact — `ui-test-up` invokes it via
`go run`, so the binary would have no consumer.

## Open questions

*(none)*
