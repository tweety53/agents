package pidfile

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// uitestExecutable is a stand-in for the UI-test stack's differently-named
// copy of the daemon. Recording it rather than the test binary's own path
// is what proves the identity check compares against the recorded
// executable instead of a fixed program name.
const uitestExecutable = "/tmp/flow-uitest-flowd"

// discardLogger is the logger every test hands Check: the stale-file
// warnings are behavior this package owes its operator, not something the
// tests assert on, so they are written nowhere.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// capturingLogger returns a logger and the buffer it writes to, for the
// two tests that assert Release says something. Everything else discards:
// the stale-file warnings are behavior this package owes its operator, and
// only the paths where the operator has no other signal are pinned here.
func capturingLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})), &buf
}

// stubProcessName substitutes the package's process-name lookup for the
// duration of one test, so no test has to spawn a real process with a
// chosen name. It restores the real lookup afterwards, which is why no
// test in this file runs in parallel.
func stubProcessName(t *testing.T, fn func(pid int) (string, error)) {
	t.Helper()
	previous := processName
	processName = fn
	t.Cleanup(func() { processName = previous })
}

// rejectProcessNameLookup fails the test if the process-name lookup is
// consulted at all. Check tests liveness first, so a file whose pid is
// dead -- or that never parsed -- must be declared stale without one.
func rejectProcessNameLookup(t *testing.T) {
	t.Helper()
	stubProcessName(t, func(pid int) (string, error) {
		t.Errorf("process-name lookup consulted for pid %d; the file was already stale", pid)
		return "", errors.New("must not be called")
	})
}

// writePidfile writes the two-line form Check reads: the decimal pid,
// then the recorded executable's path.
func writePidfile(t *testing.T, path string, pid int, executable string) {
	t.Helper()
	contents := fmt.Sprintf("%d\n%s\n", pid, executable)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("writing the pidfile under test: %v", err)
	}
}

// readPidfile returns the pidfile's lines with the trailing newline
// removed, so a test can assert on what Write recorded.
func readPidfile(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the pidfile back: %v", err)
	}
	return strings.Split(strings.TrimRight(string(data), "\n"), "\n")
}

// assertFileUnchanged fails the test unless path still holds exactly the
// bytes it held when want was read. Every Check test uses it: Check is a
// question, and a question that writes is the defect this package's split
// into Check and Write exists to prevent -- a daemon that writes before
// the bind can lose the bind and then delete the winner's file.
func assertFileUnchanged(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s back after Check: %v", path, err)
	}
	if string(got) != string(want) {
		t.Errorf("Check rewrote %s: it now reads %q, want it unchanged at %q", path, got, want)
	}
}

// deadPid returns a process id that has certainly exited and been reaped,
// for the case where a daemon was killed without removing its file.
func deadPid(t *testing.T) int {
	t.Helper()
	cmd := exec.Command("sh", "-c", "exit 0")
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting the short-lived process: %v", err)
	}
	pid := cmd.Process.Pid
	if err := cmd.Wait(); err != nil {
		t.Fatalf("waiting for the short-lived process: %v", err)
	}
	return pid
}

func TestPathIsDerivedFromPort(t *testing.T) {
	live := Path(4173)
	uitest := Path(4174)

	if want := filepath.Join(os.TempDir(), "flowd-4173.pid"); live != want {
		t.Errorf("Path(4173) = %q, want %q", live, want)
	}
	if want := filepath.Join(os.TempDir(), "flowd-4174.pid"); uitest != want {
		t.Errorf("Path(4174) = %q, want %q", uitest, want)
	}
	if live == uitest {
		t.Errorf("Path(4173) and Path(4174) both returned %q; the live and UI-test stacks would contend for one file", live)
	}
}

func TestCheckPassesWhenNoFileExists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flowd-4173.pid")
	rejectProcessNameLookup(t)

	if err := Check(path, discardLogger()); err != nil {
		t.Fatalf("Check on an absent file: %v, want nil -- a first-ever start must not be refused", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("Check created %s (stat error %v); it must write nothing, so that a daemon which then loses the bind has nothing to release", path, err)
	}
}

func TestCheckRefusesWhenLiveProcessMatchesRecordedExecutable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flowd-4174.pid")
	holder := os.Getpid()
	writePidfile(t, path, holder, uitestExecutable)
	held, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the holding pidfile: %v", err)
	}
	stubProcessName(t, func(int) (string, error) {
		return filepath.Base(uitestExecutable), nil
	})

	err = Check(path, discardLogger())

	if !errors.Is(err, ErrAlreadyRunning) {
		t.Fatalf("Check error = %v, want ErrAlreadyRunning", err)
	}
	if want := strconv.Itoa(holder); !strings.Contains(err.Error(), want) {
		t.Errorf("Check error %q does not name the holding pid %s", err, want)
	}
	assertFileUnchanged(t, path, held)
}

func TestCheckPassesWhenRecordedPidIsNotAlive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flowd-4173.pid")
	writePidfile(t, path, deadPid(t), uitestExecutable)
	stale, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the stale pidfile: %v", err)
	}
	rejectProcessNameLookup(t)

	if err := Check(path, discardLogger()); err != nil {
		t.Fatalf("Check over a killed daemon's file: %v, want nil -- a SIGKILLed daemon must leave an inconvenience, not a lockout", err)
	}
	assertFileUnchanged(t, path, stale)
}

func TestCheckPassesWhenLivePidIsAnUnrelatedProcess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flowd-4173.pid")
	writePidfile(t, path, os.Getpid(), uitestExecutable)
	recycled, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the recycled-pid file: %v", err)
	}
	stubProcessName(t, func(int) (string, error) {
		return "vim", nil
	})

	if err := Check(path, discardLogger()); err != nil {
		t.Fatalf("Check over a recycled pid: %v, want nil -- a bare kill -0 would make this a permanent lockout", err)
	}
	assertFileUnchanged(t, path, recycled)
}

func TestCheckPassesOnUnparsableFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flowd-4173.pid")
	garbage := []byte("not a pidfile at all\n")
	if err := os.WriteFile(path, garbage, 0o600); err != nil {
		t.Fatalf("writing the unparsable file: %v", err)
	}
	rejectProcessNameLookup(t)

	if err := Check(path, discardLogger()); err != nil {
		t.Fatalf("Check over an unparsable file: %v, want nil", err)
	}
	assertFileUnchanged(t, path, garbage)
}

func TestCheckPassesWhenTheFileCannotBeRead(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root reads a 0o000 file, so the unreadable case cannot be staged here")
	}
	path := filepath.Join(t.TempDir(), "flowd-4173.pid")
	// A perfectly valid pidfile naming this live test binary: if Check
	// could read it, it would refuse. What it cannot do is read it, and
	// an unreadable file is not one of the four conditions the spec lets
	// it refuse on -- it is one of the "every other case" it must log,
	// overwrite and start past.
	writePidfile(t, path, os.Getpid(), uitestExecutable)
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatalf("making the pidfile unreadable: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
	rejectProcessNameLookup(t)

	if err := Check(path, discardLogger()); err != nil {
		t.Fatalf("Check on an unreadable file: %v, want nil -- a file this daemon cannot read is stale, not a refusal, and propagating the error aborts every start until it is cleared by hand", err)
	}
}

func TestWriteRecordsPidAndExecutable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flowd-4173.pid")

	lock, err := Write(path)
	if err != nil {
		t.Fatalf("Write on an absent file: %v", err)
	}
	if lock == nil {
		t.Fatal("Write returned a nil lock and no error")
	}

	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("resolving this test binary's path: %v", err)
	}
	lines := readPidfile(t, path)
	if len(lines) != 2 {
		t.Fatalf("pidfile has %d lines (%q), want 2", len(lines), lines)
	}
	if want := strconv.Itoa(os.Getpid()); lines[0] != want {
		t.Errorf("recorded pid = %q, want %q", lines[0], want)
	}
	if lines[1] != executable {
		t.Errorf("recorded executable = %q, want %q", lines[1], executable)
	}
}

func TestReleaseRemovesFileAndIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flowd-4173.pid")
	lock, err := Write(path)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if err := lock.Release(discardLogger()); err != nil {
		t.Fatalf("first Release: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("pidfile still present after Release (stat error %v)", err)
	}
	if err := lock.Release(discardLogger()); err != nil {
		t.Errorf("second Release: %v, want nil -- run() defers it unconditionally", err)
	}
}

func TestReleaseLeavesAnotherHoldersFileInPlace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flowd-4173.pid")
	lock, err := Write(path)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	// A daemon that won the race for this port overwrites the file
	// between this lock's Write and its Release. Releasing must not
	// disarm the winner: the file no longer describes this lock.
	winner := os.Getpid() + 1
	writePidfile(t, path, winner, uitestExecutable)

	if err := lock.Release(discardLogger()); err != nil {
		t.Fatalf("Release over another holder's file: %v", err)
	}

	lines := readPidfile(t, path)
	if len(lines) != 2 || lines[0] != strconv.Itoa(winner) || lines[1] != uitestExecutable {
		t.Errorf("Release removed or rewrote another holder's pidfile: %q", lines)
	}
}

func TestReleaseLogsWhenAnotherHolderOwnsTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flowd-4173.pid")
	lock, err := Write(path)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	// The same takeover TestReleaseLeavesAnotherHoldersFileInPlace stages.
	// That test pins what Release does to the file; this one pins what it
	// tells the operator, who otherwise watches a clean shutdown remove
	// nothing and is told nothing about why.
	winner := os.Getpid() + 1
	writePidfile(t, path, winner, uitestExecutable)
	logger, logged := capturingLogger()

	if err := lock.Release(logger); err != nil {
		t.Fatalf("Release over another holder's file: %v", err)
	}

	line := logged.String()
	if line == "" {
		t.Fatalf("Release removed nothing and logged nothing; a shutdown whose pidfile was taken over must say so")
	}
	if !strings.Contains(line, "level=WARN") {
		t.Errorf("Release logged %q, want it at warn level, as Check logs every stale-file reason it finds", line)
	}
	if !strings.Contains(line, strconv.Itoa(winner)) {
		t.Errorf("Release logged %q, want it to name the pid %d that now holds the file", line, winner)
	}
}

func TestReleaseLogsWhenTheFileNoLongerParses(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flowd-4173.pid")
	lock, err := Write(path)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := os.WriteFile(path, []byte("not a pidfile\n"), 0o600); err != nil {
		t.Fatalf("overwriting the pidfile with unparsable contents: %v", err)
	}
	logger, logged := capturingLogger()

	if err := lock.Release(logger); err != nil {
		t.Fatalf("Release over an unparsable file: %v", err)
	}

	line := logged.String()
	if line == "" {
		t.Fatalf("Release left an unparsable file in place and logged nothing; the file it wrote has been replaced by something it cannot read, and only the log can say so")
	}
	if !strings.Contains(line, "level=WARN") {
		t.Errorf("Release logged %q, want it at warn level", line)
	}
}
