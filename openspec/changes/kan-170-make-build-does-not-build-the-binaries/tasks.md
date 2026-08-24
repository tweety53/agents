# Tasks — kan-170-make-build-does-not-build-the-binaries

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

Baseline counts this plan measures against:

- Go, on the metric `cd stats && go test <pkg> -count=1 -v | grep -c '^=== RUN   Test'` —
  `./cmd/myflowd` reads 1, and `./internal/pidfile` does not exist yet and reads 0.
  <!-- measured: cd stats && go test ./cmd/myflowd/ -count=1 -v | grep -c '^=== RUN   Test' @ branch main -->
- Shell harnesses, on the metric `bash scripts/<harness> | grep -c '^ok: '` —
  `test-make-build.sh` does not exist yet and reads 0.
  <!-- measured: ls scripts/test-make-build.sh @ branch main (no such file) -->

The Go metric counts subtests, not only top-level test functions. No task below adds a
table-driven test, so each `**Baseline:**` moves by the count of `func Test` it adds — if that
stops being true, correct the number rather than reshaping the test to match it.

### 1 `make build` writes both binaries, and the docs stop naming a workaround

- [x] Change `stats/Makefile`'s `build` target to emit both binaries after the compile-all pass:

```make unverified:confirm GNU Make accepts the aligned two-space run-in on the second -o line as written
build: web-build
	go build ./...
	go build -o bin/myflowd ./cmd/myflowd
	go build -o bin/myflow  ./cmd/myflow
```

  Keep `go build ./...` and keep it **first**: it is what holds `cmd/uitest-seed` and every library
  compiling under this target, and dropping it trades the target's coverage for the two artifacts.
  Update the target's own comment — it currently describes a target that only verifies the Go build.

- [x] Add `scripts/test-make-build.sh`, an assertion harness in the shape of
  `scripts/test-check-uitest-overrides.sh`: `set -euo pipefail`, `fail`/`pass` helpers printing
  `FAIL: <case>` and `ok: <case>`, exiting non-zero on any failure. It runs `make -n build` in
  `stats/` and asserts on the printed recipe — `-n` is the whole point, since the defect is a
  missing `-o` in the recipe text and `-n` shows it without a real build, an `npm ci`, or a Vite
  run.

- [x] Add `scripts/test-make-build.sh` to `.myflow/project.md`'s `## test` list, beside the other
  `scripts/test-*.sh` harnesses.

- [x] `.myflow/project.md`'s `## run` section: drop the
  `cd stats && go build -o bin/myflowd ./cmd/myflowd && ./bin/myflowd` line's explicit build half,
  leaving `cd stats && make build` followed by `./bin/myflowd`, and correct the `make build`
  comment, which today says the target "verifies the Go build".

- [x] `stats/README.md`'s launchd install block (step 2): drop the
  `go build -o bin/myflowd ./cmd/myflowd` line that follows `cd stats && make build`. Its
  surrounding comment already says the plist's `ProgramArguments` is a fixed path — after this task
  `make build` is what writes that path.

**Files:** `stats/Makefile`, `scripts/test-make-build.sh`, `.myflow/project.md`, `stats/README.md`
**Tests:** `build target emits bin/myflowd`, `build target emits bin/myflow`,
`build target still compiles every package`
**Regression:** `build target emits bin/myflowd` and `build target emits bin/myflow` each fail if
the recipe reverts to a bare `go build ./...`, which is this ticket's first defect — the compiler
discards both binaries and the operator runs stale code.
`build target still compiles every package` fails if `go build ./...` is dropped in favour of the
two `-o` builds alone, which would silently stop compiling `cmd/uitest-seed`.
**Baseline:** before=0 after=3
<!-- predicted: bash scripts/test-make-build.sh | grep -c '^ok: ' after task 1 -->
**Build:** green
**Commit:** `fix(stats): build both binaries in make build`

### 2 `internal/pidfile`

- [x] Add `stats/internal/pidfile/pidfile.go`. Standard library only — no new dependency.

```go verified:the split below is design.md's own ordering; the single-Acquire shape this replaced was implemented, reviewed, and shown to disarm a live daemon
package pidfile

// Path returns the pidfile path for a resolved daemon port.
func Path(port int) string

// Check refuses with ErrAlreadyRunning when path names a live process whose
// name matches the recorded executable's basename. It writes nothing.
func Check(path string, logger *slog.Logger) error

// Write records this process at path and returns a Lock. It is called only
// after the listener is open, so the process that writes is the process that
// holds the port.
func Write(path string) (*Lock, error)

// Release removes the pidfile, and only when it still records this lock's
// own pid. It logs at warn on each path where it leaves the file alone.
// It is a no-op once called.
func (l *Lock) Release(logger *slog.Logger) error

var ErrAlreadyRunning = errors.New("pidfile: a daemon is already running")
```

**The check and the write are two calls, not one.** A single `Acquire` that checked and wrote
before the bind was implemented and reviewed, and it is unsafe: the write race and the bind race are
independent, so the daemon that wins the write can be the one that loses the bind, and its deferred
`Release` — pid-matched and therefore passing — deletes the pidfile of the daemon that actually holds
the port. Splitting them restores `design.md`'s ordering and closes the window outright: a daemon
that loses the bind never writes, so it has nothing to release, and the write-winner is the
bind-winner by construction.

- [x] `Path` returns `filepath.Join(os.TempDir(), fmt.Sprintf("myflowd-%d.pid", port))`. Port-derived
  so the live stack (4173) and the UI-test stack (4174) never contend for one file. It takes no
  caller-supplied path and reads no environment variable — a caller-nameable path feeding the
  removal below is the hazard `stats/Makefile`'s `UITEST_*` `override` block exists to prevent.

- [x] File contents: two lines — the decimal pid, then `os.Executable()`'s result. Anything that does
  not parse into those two lines is a stale file.

- [x] `Check`'s liveness test, in order: read and parse the file; `syscall.Kill(pid, 0)` for
  existence; then compare the live process's name against `filepath.Base(<recorded executable>)`.
  Refuse with `ErrAlreadyRunning` — naming the pid — only when both hold. Every other outcome is a
  stale file: log it at warn level with the reason and return nil. Comparing against the
  **recorded executable** rather than a hardcoded `"myflowd"` is what makes the check fire for the
  UI-test daemon, which runs as `/tmp/myflow-uitest-myflowd`. `syscall.Kill` returning `EPERM` counts
  as alive: the process exists and is owned by another user, and calling it dead would overwrite a
  file this daemon does not own.

- [x] The process-name lookup shells to `ps -o comm= -p <pid>` — macOS has no `/proc`, and `ps`
  answers on both platforms this repository builds on. Put it behind an unexported package variable
  so the tests substitute it instead of spawning real processes; do not export a hook for it.

- [x] `Release(logger *slog.Logger)` removes the file **only when it still records this lock's own
  pid**, and is safe to call twice, so `run()` can `defer` it unconditionally. A file that is
  absent, no longer parses, or names another process is left alone — and each of the two left-alone
  paths logs at warn, mirroring `Check`, so a takeover is visible in the shutdown log rather than
  silent. An absent file stays silent: that is the second call of the unconditional defer.

**Files:** `stats/internal/pidfile/pidfile.go`, `stats/internal/pidfile/pidfile_test.go`
**Tests:** `TestPathIsDerivedFromPort`, `TestCheckPassesWhenNoFileExists`,
`TestCheckRefusesWhenLiveProcessMatchesRecordedExecutable`,
`TestCheckPassesWhenRecordedPidIsNotAlive`,
`TestCheckPassesWhenLivePidIsAnUnrelatedProcess`,
`TestCheckPassesOnUnparsableFile`, `TestCheckPassesWhenTheFileCannotBeRead`,
`TestWriteRecordsPidAndExecutable`, `TestReleaseRemovesFileAndIsIdempotent`,
`TestReleaseLeavesAnotherHoldersFileInPlace`, `TestReleaseLogsWhenAnotherHolderOwnsTheFile`,
`TestReleaseLogsWhenTheFileNoLongerParses`
**Regression:** `TestPathIsDerivedFromPort` fails if the path is fixed rather than port-derived,
which is what would make `make ui-test-up` refuse while the live daemon runs.
`TestCheckRefusesWhenLiveProcessMatchesRecordedExecutable` fails if the refusal is dropped —
the ticket's second defect, a second daemon started beside a live one.
`TestCheckPassesWhenLivePidIsAnUnrelatedProcess` fails if the identity check degrades to a bare
`kill -0`, which would make a recycled pid a permanent lockout.
`TestCheckPassesWhenRecordedPidIsNotAlive` fails if a `SIGKILL`ed daemon's leftover file
becomes a lockout. `TestCheckPassesWhenNoFileExists` fails if a first-ever start is refused.
`TestWriteRecordsPidAndExecutable` fails if the executable line is dropped,
which is what the UI-test daemon's differently-named binary needs.
`TestCheckPassesWhenTheFileCannotBeRead` fails if an unreadable pidfile is propagated as an error
rather than treated as stale — the spec enumerates four stale conditions and requires every other
case to be logged, overwritten and started past, so propagating aborts every start until someone
clears the file by hand. `TestReleaseLogsWhenAnotherHolderOwnsTheFile` and
`TestReleaseLogsWhenTheFileNoLongerParses` fail if `Release` goes back to returning silently on
either path, leaving an operator no shutdown diagnostic where `Check` logs every stale-file reason
it finds.
`TestReleaseRemovesFileAndIsIdempotent` fails if `Release` panics or errors on a second call, which
`run()`'s unconditional `defer` would then hit on every clean shutdown.
`TestReleaseLeavesAnotherHoldersFileInPlace` fails if `Release` removes the file without checking it
still records this lock's own pid — task 3's `acquireStartup` defers exactly that `Release` on its
`net.Listen` failure path, so an unconditional remove would delete a live daemon's pidfile and
silently disarm the refusal for every later start.
**Baseline:** before=0 after=12
<!-- measured: cd stats && go test ./internal/pidfile/ -count=1 -v | grep -c '^=== RUN   Test' @ branch openspec/kan-170-make-build-does-not-build-the-binaries -->
**Build:** green
**Commit:** `feat(pidfile): refuse to start beside a live daemon`

### 3 `run()` acquires the pidfile and binds before it opens the store

- [x] In `stats/cmd/myflowd/main.go`, extract the startup prelude into a function `run` calls, so it
  is reachable from `wiring_test.go` without a database:

```go verified:config.Config.Validate is exported and was called from package main in this task's first implementation
// acquireStartup claims the daemon's port, then records the pidfile, before
// anything touches the database. On any failure it releases whatever it
// already claimed and returns no listener.
func acquireStartup(cfg config.Config, logger *slog.Logger) (*pidfile.Lock, net.Listener, error)
```

  Its body, in order: `cfg.Validate()`, `pidfile.Check(pidfile.Path(cfg.Port), logger)`,
  `net.Listen("tcp", cfg.Addr())`, `pidfile.Write(pidfile.Path(cfg.Port))`. **The write comes after
  the bind, and that order is the point** — see task 2's own note on why a check-and-write before the
  bind lets a bind-loser delete the winner's file. A failed `net.Listen` therefore has no lock to
  release and must write nothing at all; a failed `pidfile.Write` closes the listener before
  returning.

- [x] Reorder `run()`'s **prelude** to: `config.FromEnv()` → `acquireStartup(cfg, logger)` →
  `defer lock.Release()` → `defer ln.Close()`, ahead of `store.Open`. Everything below that —
  `RunMigrations`, `SeedPricing`, `web.FS`/`web.Handler`, `api.New`, the reconciler, the harvest and
  sweep goroutines, and finally `srv.Serve(ln)` — keeps the order it already has and is not this
  task's to move. Replace the
  `srv.ListenAndServe()` call in the serve goroutine with `srv.Serve(ln)`;
  `api.Server.Serve(net.Listener)` already exists beside `ListenAndServe` and needs no change to
  `internal/api`. The `defer ln.Close()` covers the paths between the listener opening and
  `srv.Serve` taking ownership of it — `store.Open`, the migrations, the seeding, `web.FS` and
  `api.New` each `return err` with the port still held, and every other resource in this function is
  released by an explicit defer. `Serve` closes the listener itself on shutdown, so the deferred
  close then returns an already-closed error and is discarded.

- [x] `cfg.Validate()` is called explicitly in `acquireStartup` because `api.New` — which calls it
  today — now runs *after* the listener is open. Keep `api.New`'s own call; it is cheap and it keeps
  `internal/api` correct for any other caller. Update the comment at the `api.New` call site, which
  currently says a non-loopback host is refused there.

- [x] Update `main.go`'s package doc comment: it describes the bind as loopback-only, and now also
  states that the daemon claims its port before its database.

**Files:** `stats/cmd/myflowd/main.go`, `stats/cmd/myflowd/wiring_test.go`
**Tests:** `TestAcquireStartupOpensListenerAndPidfile`,
`TestAcquireStartupRefusesWhenAnotherDaemonHoldsThePidfile`,
`TestAcquireStartupWritesNoPidfileWhenListenFails`,
`TestAcquireStartupRefusesNonLoopbackHostBeforeListening`
**Regression:** `TestAcquireStartupRefusesWhenAnotherDaemonHoldsThePidfile` fails if the pidfile
check is not wired into startup at all, leaving `internal/pidfile` dead code.
`TestAcquireStartupWritesNoPidfileWhenListenFails` fails if the write is moved back ahead of the
bind — the ordering itself, which no earlier version of this task's tests pinned, so a reorder went
undetected. A bind-loser that writes is the defect: its own `Release` then deletes the pidfile of the
daemon holding the port.
`TestAcquireStartupRefusesNonLoopbackHostBeforeListening` fails if `cfg.Validate()` is dropped from
the prelude — the loopback-only rule would then be enforced only after a listener was already open.
`TestAcquireStartupOpensListenerAndPidfile` fails if the prelude regains a store call, which is what
`myflow-daemon-single-instance`'s "a refused start touches no database" requires it not to have.
**Baseline:** before=1 after=5
<!-- predicted: cd stats && go test ./cmd/myflowd/ -count=1 -v | grep -c '^=== RUN   Test' after task 3 -->
**Build:** green
**Commit:** `fix(myflowd): claim the port and pidfile before opening the store`

### 4 The UI-test stack reads the daemon's pidfile, and `restart:` loses a stale comment

- [x] `stats/Makefile`: change `override UITEST_PIDFILE := /tmp/myflow-uitest.pid` to
  `override UITEST_PIDFILE := /tmp/myflowd-$(UITEST_PORT).pid` — the file `myflowd` now writes
  itself. Keep `override`: the Makefile's own RULE is that every `UITEST_*` variable carries it, and
  `scripts/check-uitest-overrides.sh` enforces that over the whole class. Move the assignment below
  `UITEST_PORT`'s if it is not already, so the reference resolves in reading order.

- [x] `ui-test-up`: drop the `& echo $! > $(UITEST_PIDFILE)` half of the daemon launch, leaving the
  background start alone. The daemon writes the file now, so a second copy written here could only
  disagree with it. Keep `rm -f $(UITEST_PIDFILE)` before the launch: it clears a file left by a
  daemon this target killed, and `myflowd` overwriting a stale file does not make removing one wrong.

- [x] `ui-test-up`'s stop-previous-daemon block and `ui-test-down`'s stop block read the same
  `$(UITEST_PIDFILE)`; neither needs a change beyond the path. Update the comment above `ui-test-up`
  that explains the two mechanisms, and `ui-test-down`'s comment about the bounded wait, to say the
  pid comes from the daemon's own file.

- [x] `restart:`: rewrite the `# NO PIDFILE, deliberately.` paragraph. Its premise — that a pidfile
  *this target* wrote would be absent for a hand-started daemon — no longer holds, because the daemon
  writes one on every start however it was started. The kill **stays port-driven**: the recorded
  reasoning that a `kill` driven by a caller-nameable file is the hazard, and one driven by "who
  holds this port" has no such handle, is not overturned by this change. Say that, rather than
  deleting the paragraph.

- [x] Extend `scripts/test-make-build.sh` with two cases over `stats/Makefile`'s text.

**Files:** `stats/Makefile`, `scripts/test-make-build.sh`
**Tests:** `ui-test-up does not write its own pidfile`,
`UITEST_PIDFILE is override and derived from UITEST_PORT`,
`ui-test pid reads take only the pidfile's pid line`
**Regression:** `ui-test-up does not write its own pidfile` fails if the `echo $! >` is restored,
recreating the two-writers-one-process split this task removes.
`UITEST_PIDFILE is override and derived from UITEST_PORT` fails if the path is hardcoded back to
`/tmp/myflow-uitest.pid` — the tear-down would then signal a pid nothing writes — or if `override`
is dropped, which `scripts/check-uitest-overrides.sh` also catches and which this case asserts from
the consuming side.
`ui-test pid reads take only the pidfile's pid line` fails if any of the three pid reads goes back
to `cat`, which captures both lines of the daemon-written pidfile: `kill -0` then rejects the
embedded newline as an illegal pid and always fails, so a live previous daemon is never detected —
`ui-test-up` launches a second daemon that cannot bind, and `ui-test-down` skips its SIGTERM and
drops the database out from under a live connection. The case *evaluates* the extracted reads
against a fixture two-line pidfile naming a live process rather than matching their spelling, since
the defect was spelled exactly as intended.
**Baseline:** before=3 after=6
<!-- measured: bash scripts/test-make-build.sh | grep -c '^ok: ' @ branch openspec/kan-170-make-build-does-not-build-the-binaries -->
**Build:** green
**Commit:** `fix(stats): point the ui-test stack at the daemon's own pidfile`
