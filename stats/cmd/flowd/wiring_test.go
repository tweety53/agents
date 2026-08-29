package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/harvest"
	"github.com/tweety53/agents/stats/internal/pidfile"
	"github.com/tweety53/agents/stats/internal/store"
)

// The four TestAcquireStartup* tests below cover KAN-170 task 3 the same
// way TestNewTranscriptWatcherWiresBinderAndPricer covers task 7 of
// KAN-172: through the extracted constructor rather than through run()
// itself. run() opens a database, replays a journal and starts three
// background loops, so a test that drove it could not answer the one
// question this change is about -- what the daemon does *before* any of
// that. acquireStartup is exactly the prelude, so asserting on it asserts
// on the ordering.
//
// Each of them gives cfg a DSN that store.Open cannot even parse
// (pgxpool.ParseConfig rejects it before opening a socket). That is not
// decoration: it is what makes "a refused start touches no database"
// (openspec/specs/myflow-daemon-single-instance) structurally true here
// rather than merely asserted in a comment. Any store call added to the
// prelude fails these tests immediately, with no database reachable and
// none required to run them.
const unparsableDSN = "kan-170: acquireStartup must never open a store"

// startupConfig returns a Config bound to a free loopback port, with the
// pidfile redirected into the test's own temporary directory. Port 0 lets
// the kernel choose the port at bind time, so concurrent tests never race
// for a fixed one; pidfile.Path is derived from that same 0, which is why
// TMPDIR has to move as well -- otherwise every test in this file would
// share /tmp/flowd-0.pid with every other run on the machine.
func startupConfig(t *testing.T) config.Config {
	t.Helper()
	t.Setenv("TMPDIR", t.TempDir())
	return config.Config{Host: config.DefaultHost, Port: 0, DSN: unparsableDSN}
}

// discardLogger is what these tests hand acquireStartup: pidfile.Check
// logs every stale file it passes over through it, and none of these
// tests asserts on that log.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestAcquireStartupOpensListenerAndPidfile(t *testing.T) {
	cfg := startupConfig(t)

	lock, ln, err := acquireStartup(cfg, discardLogger())
	if err != nil {
		t.Fatalf("acquireStartup: %v", err)
	}
	defer func() {
		_ = ln.Close()
		_ = lock.Release(discardLogger())
	}()

	if _, ok := ln.Addr().(*net.TCPAddr); !ok {
		t.Errorf("listener address is %T, want *net.TCPAddr", ln.Addr())
	}
	data, err := os.ReadFile(pidfile.Path(cfg.Port))
	if err != nil {
		t.Fatalf("read pidfile: %v", err)
	}
	wantPid := strconv.Itoa(os.Getpid())
	if got := strings.SplitN(string(data), "\n", 2)[0]; got != wantPid {
		t.Errorf("pidfile records pid %q, want this process's %q", got, wantPid)
	}
}

func TestAcquireStartupRefusesWhenAnotherDaemonHoldsThePidfile(t *testing.T) {
	cfg := startupConfig(t)

	// The holder this test writes is the test process itself: its pid is
	// alive by construction, and ps reports it running the very
	// executable os.Executable names, so pidfile's liveness-then-identity
	// check finds a genuine live holder without spawning anything.
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("resolve this process's executable: %v", err)
	}
	held := fmt.Sprintf("%d\n%s\n", os.Getpid(), executable)
	if err := os.WriteFile(pidfile.Path(cfg.Port), []byte(held), 0o600); err != nil {
		t.Fatalf("write the holding pidfile: %v", err)
	}

	lock, ln, err := acquireStartup(cfg, discardLogger())

	if !errors.Is(err, pidfile.ErrAlreadyRunning) {
		t.Fatalf("acquireStartup error is %v, want one wrapping pidfile.ErrAlreadyRunning", err)
	}
	if lock != nil || ln != nil {
		t.Errorf("acquireStartup returned lock=%v listener=%v on a refusal, want both nil", lock, ln)
	}
	// The holder's file must survive untouched: overwriting it would
	// leave the running daemon's shutdown removing a file that no longer
	// describes it.
	got, err := os.ReadFile(pidfile.Path(cfg.Port))
	if err != nil {
		t.Fatalf("read the holding pidfile back: %v", err)
	}
	if string(got) != held {
		t.Errorf("the holder's pidfile now reads %q, want it unchanged at %q", got, held)
	}
}

// TestAcquireStartupWritesNoPidfileWhenListenFails pins the *ordering* of
// the prelude, not merely its cleanup. The write comes after the bind, so
// a daemon that loses the bind has written nothing and has nothing to
// release. Moving the write back ahead of the bind fails this test: the
// listen-failure path releases nothing -- it has no lock to release -- so
// a pidfile written before the bind is simply left behind.
//
// The distinction matters because the ordering, not the cleanup, is what
// protects the daemon that won the port: a bind-loser that wrote the file
// would hold a pid-matched lock, and its own Release would then delete the
// winner's pidfile and silently disarm every later start's refusal.
func TestAcquireStartupWritesNoPidfileWhenListenFails(t *testing.T) {
	cfg := startupConfig(t)

	// Occupy a port, then ask acquireStartup for that same one. The bind
	// is the middle step of the prelude, so this exercises the path where
	// the pidfile check has passed and the listener then fails.
	occupied, err := net.Listen("tcp", net.JoinHostPort(config.DefaultHost, "0"))
	if err != nil {
		t.Fatalf("occupy a port: %v", err)
	}
	defer func() { _ = occupied.Close() }()
	cfg.Port = occupied.Addr().(*net.TCPAddr).Port

	lock, ln, err := acquireStartup(cfg, discardLogger())

	if err == nil {
		t.Fatalf("acquireStartup succeeded on an occupied port; want a bind failure")
	}
	if lock != nil || ln != nil {
		t.Errorf("acquireStartup returned lock=%v listener=%v on a failed bind, want both nil", lock, ln)
	}
	if _, err := os.Stat(pidfile.Path(cfg.Port)); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("pidfile %s exists after a failed bind (stat error %v); the write must come after the bind, so a bind-loser writes nothing at all",
			pidfile.Path(cfg.Port), err)
	}
}

func TestAcquireStartupRefusesNonLoopbackHostBeforeListening(t *testing.T) {
	cfg := startupConfig(t)
	cfg.Host = "10.0.0.1"

	lock, ln, err := acquireStartup(cfg, discardLogger())

	// The assertion is on the *identity* of the error, not merely that
	// one occurred: binding 10.0.0.1 would fail on most hosts anyway, so
	// only ErrNonLoopbackHost distinguishes "refused by the loopback rule"
	// from "the kernel happened to reject the address".
	if !errors.Is(err, config.ErrNonLoopbackHost) {
		t.Fatalf("acquireStartup error is %v, want one wrapping config.ErrNonLoopbackHost", err)
	}
	if lock != nil || ln != nil {
		t.Errorf("acquireStartup returned lock=%v listener=%v for a non-loopback host, want both nil", lock, ln)
	}
	if _, err := os.Stat(pidfile.Path(cfg.Port)); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("pidfile %s was written before the loopback check refused the config (stat error %v)",
			pidfile.Path(cfg.Port), err)
	}
}

// TestDaemonWiresTheRealStore replaces TestNewTranscriptWatcherWiresBinderAndPricer
// (deleted in KAN-173 task 2, which removed the Has* accessors it asserted
// on). It closes the gap task 2's compile-time guarantee cannot cover:
// harvest.NewWatcher now panics if deps is nil, but nothing stops
// newTranscriptWatcher from passing a non-nil, wrong deps -- and
// harvest.NoDeps{} is the wrong value that most resembles a right one, since
// a daemon wired with it compiles, runs, harvests every transcript, and
// prices nothing, binds nothing and charges no dispatch, silently and green.
//
// Observed: temporarily editing newTranscriptWatcher to
// `return harvest.NewWatcher(root, st, attributor, harvest.NoDeps{}, logger)`
// made this test fail with:
//
//	wiring_test.go:209: the daemon wired harvest.NoDeps: nothing is priced, no session token is bound, no dispatch is charged
//
// Reverting that edit (back to passing st as deps) made it pass. Both runs
// used `go test ./cmd/flowd/ -race -count=1 -run TestDaemonWiresTheRealStore -v`.
func TestDaemonWiresTheRealStore(t *testing.T) {
	var st *store.Store // never dereferenced: newTranscriptWatcher only stores it behind Deps and HarvestSink.
	w := newTranscriptWatcher(t.TempDir(), st, harvest.NewAttributor(nil), nil)

	f := reflect.ValueOf(w).Elem().FieldByName("deps")
	if !f.IsValid() {
		t.Fatal("harvest.Watcher has no deps field: this test no longer checks anything")
	}
	if f.IsNil() {
		t.Fatal("the daemon's watcher has a nil deps")
	}
	if f.Elem().Type() == reflect.TypeOf(harvest.NoDeps{}) {
		t.Error("the daemon wired harvest.NoDeps: nothing is priced, no session token is bound, no dispatch is charged")
	}
}
