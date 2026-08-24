// Package pidfile keeps a second myflowd from starting beside a live one.
// The daemon records its own pid and the path of its own executable in a
// file named after the port it is about to serve, and refuses to start
// when that file names a process that is both alive and running the
// recorded executable -- the failure this package exists to prevent is two
// daemons harvesting transcripts into one database, which the operating
// system's own EADDRINUSE only reports after the second daemon has already
// migrated and seeded.
//
// Every other state of the file is stale, not a refusal: absent,
// unreadable, unparsable, a pid that is no longer alive, or a pid recycled
// onto an unrelated program. Each is logged with its reason and then
// overwritten, so a SIGKILLed daemon leaves an inconvenience rather than a
// lockout that has to be cleared by hand.
//
// # What this check does and does not guarantee
//
// The check and the write are two calls, and the caller binds its
// listener between them: Check, then net.Listen, then Write. Check is a
// question -- it reads the file, judges it, and answers, writing nothing.
// The authoritative exclusion is the bind, so the process that records
// itself in the file is by construction the process that already holds
// the port.
//
// Two daemons starting on the same port at the same instant can therefore
// both pass Check, but only one of them ever reaches Write: the other
// fails at net.Listen having written nothing at all, and so has nothing to
// release and no way to disturb the winner's file. That ordering, not a
// lock, is what closes the race. An exclusive create (O_CREATE|O_EXCL)
// would close only the absent-file half and leave the stale-takeover half
// open, and a real lock would need flock; ordering the two calls around
// the bind needs neither.
//
// What the package adds over the bind alone is an earlier and better
// refusal in the case that actually occurs -- an operator starting a
// second daemon seconds or minutes after the first -- so the refusal
// carries the holding pid and happens before the database is opened,
// migrated and seeded.
//
// Release remains pid-matched: it removes the file only while the file
// still records its own lock's pid, so a daemon whose file has since been
// taken over by another removes nothing.
package pidfile

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// ErrAlreadyRunning is returned by Check when the pidfile names a live
// process running the recorded executable. The returned error names the
// holding pid, which is what the operator needs in order to act on it.
var ErrAlreadyRunning = errors.New("pidfile: a daemon is already running")

// filePerm is the pidfile's mode. The file carries no secret -- a pid and
// an executable path -- but nothing other than its owner has any reason to
// write it, and a writable pidfile is a handle on whatever the operator
// later signals from it.
const filePerm fs.FileMode = 0o600

// processName reports the name of the process with the given pid, without
// its directory. It shells to ps because macOS has no /proc, and ps
// answers on both platforms this daemon is built for.
//
// Verified on macOS only, where `ps -o comm=` prints the executable's full
// path and filepath.Base recovers the name the identity check compares.
// On Linux it prints the kernel's comm, truncated to 15 characters, so a
// recorded executable with a longer basename -- the UI-test daemon's
// /tmp/myflow-uitest-myflowd is 21 -- would never compare equal, and a
// live instance of it would be read as stale and started beside. This
// repository is developed on macOS and has no CI, so the plain comparison
// is what is coded; a Linux port has to widen the comparison to allow for
// that truncation.
//
// It is an unexported package variable so this package's tests can
// substitute it rather than spawning real processes under chosen names.
// Nothing outside the package may replace it: the lookup is not
// configuration, and a caller-supplied identity check is exactly the
// handle this package must not offer.
var processName = func(pid int) (string, error) {
	out, err := exec.Command("ps", "-o", "comm=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return "", fmt.Errorf("pidfile: ps for pid %d: %w", pid, err)
	}
	name := strings.TrimSpace(string(out))
	if name == "" {
		return "", fmt.Errorf("pidfile: ps reported no name for pid %d", pid)
	}
	return filepath.Base(name), nil
}

// Lock is a held pidfile. Release removes it, while it still records the
// pid Write recorded.
type Lock struct {
	path string
	pid  int
}

// Path returns the pidfile path for a resolved daemon port. It is derived
// from the port so that the live stack and the UI-test stack, which run on
// different ports, never contend for one file. It takes no caller-supplied
// path and reads no environment variable: a path a caller can name is a
// path a caller can point at someone else's process.
func Path(port int) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("myflowd-%d.pid", port))
}

// Check returns an error wrapping ErrAlreadyRunning when path names a live
// process running the recorded executable, and nil in every other case --
// logging why the file was treated as stale before the caller goes on to
// bind and overwrite it. That is the whole of its error contract: the only
// error it ever returns is the refusal. A file it cannot even read is
// reported as stale rather than returned, because a start blocked by a
// file whose contents nobody has seen is a lockout, not an exclusion.
//
// It writes nothing at all, in either case: it is called before the
// listener is open, and a caller that loses the bind must be left with
// nothing to undo.
//
// The order of its steps matters: read, then parse, then liveness, then
// identity, each step cheaper than the one it guards.
func Check(path string, logger *slog.Logger) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		// A file this daemon cannot read -- wrong owner, no read bit --
		// is stale for the same reason an unparsable one is: it tells
		// this daemon nothing about whether another is running, and a
		// refusal is owed only to a file that names a live matching
		// process. Propagating it instead would abort every start until
		// someone cleared the file by hand, which is exactly the lockout
		// this package's stale-file policy exists to rule out.
		logger.Warn("stale pidfile, overwriting",
			"path", path, "reason", "the file could not be read", "error", err)
		return nil
	}

	pid, executable, err := parse(data)
	if err != nil {
		logger.Warn("stale pidfile, overwriting", "path", path, "reason", err)
		return nil
	}
	if !alive(pid) {
		logger.Warn("stale pidfile, overwriting",
			"path", path, "pid", pid, "reason", "the recorded process is no longer alive")
		return nil
	}
	name, err := processName(pid)
	if err != nil {
		logger.Warn("stale pidfile, overwriting",
			"path", path, "pid", pid, "reason", "the live process's name could not be read", "error", err)
		return nil
	}
	// The comparison is against the executable recorded in the file, not
	// a fixed "myflowd": the UI-test stack runs the same daemon from a
	// differently-named copy, and a fixed name would not recognise it.
	if want := filepath.Base(executable); name != want {
		logger.Warn("stale pidfile, overwriting",
			"path", path, "pid", pid, "process", name, "expected", want,
			"reason", "the recorded pid now belongs to an unrelated process")
		return nil
	}
	return fmt.Errorf("%w: pid %d is running %s", ErrAlreadyRunning, pid, executable)
}

// Write records this process in the pidfile at path and returns a Lock the
// caller releases at shutdown. It is called only once the caller's
// listener is open, so the process that writes the file is the process
// that holds the port -- see this package's own doc comment for why that
// ordering, rather than a lock, is what makes the file trustworthy.
//
// It overwrites whatever Check judged stale. Check having passed is the
// caller's obligation, not something Write re-verifies: a second read here
// would be a second, weaker copy of the same judgement.
func Write(path string) (*Lock, error) {
	executable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("pidfile: resolve this process's executable: %w", err)
	}
	pid := os.Getpid()
	contents := fmt.Sprintf("%d\n%s\n", pid, executable)
	if err := os.WriteFile(path, []byte(contents), filePerm); err != nil {
		return nil, fmt.Errorf("pidfile: write %s: %w", path, err)
	}
	return &Lock{path: path, pid: pid}, nil
}

// Release removes the pidfile, but only while the file still records this
// lock's own pid. A file that is already gone, no longer parses, or now
// names another process is not this lock's to remove -- the last of those
// is a daemon that has since taken the port over, and deleting its file
// would leave the next start with nothing to refuse against.
//
// It follows that Release is idempotent and safe to defer unconditionally:
// the second call finds the file it removed absent and does nothing.
//
// It takes a logger for the same reason Check does, and logs at the same
// level: every path on which it declines to remove the file it was given
// is a path on which the daemon's own file was replaced under it while it
// ran, and an operator reading a shutdown that removed nothing has no
// other way to learn that. The absent-file case is silent -- that is the
// second call of an unconditional defer, and the ordinary end of a clean
// shutdown.
func (l *Lock) Release(logger *slog.Logger) error {
	data, err := os.ReadFile(l.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("pidfile: read %s: %w", l.path, err)
	}
	pid, _, err := parse(data)
	if err != nil {
		logger.Warn("pidfile left in place at shutdown, it is no longer this daemon's",
			"path", l.path, "pid", l.pid, "reason", err)
		return nil
	}
	if pid != l.pid {
		logger.Warn("pidfile left in place at shutdown, it is no longer this daemon's",
			"path", l.path, "pid", l.pid, "recorded", pid,
			"reason", "another daemon has taken this port over and recorded itself")
		return nil
	}
	if err := os.Remove(l.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("pidfile: remove %s: %w", l.path, err)
	}
	return nil
}

// parse reads the pidfile's two lines -- the decimal pid, then the
// recorded executable's path. Anything else is a stale file, reported as
// the reason it did not parse.
func parse(data []byte) (int, string, error) {
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 2 {
		return 0, "", fmt.Errorf("the file has %d lines, want 2", len(lines))
	}
	pid, err := strconv.Atoi(lines[0])
	if err != nil {
		return 0, "", fmt.Errorf("the first line %q is not a pid", lines[0])
	}
	if pid <= 0 {
		return 0, "", fmt.Errorf("the recorded pid %d is not a process id", pid)
	}
	if lines[1] == "" {
		return 0, "", errors.New("the file records no executable")
	}
	return pid, lines[1], nil
}

// alive reports whether a process with this pid exists. Signal 0 performs
// the permission and existence checks and delivers nothing. EPERM counts
// as alive: the process exists, it just belongs to another user, and
// treating it as dead would let this daemon overwrite a file it does not
// own.
func alive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
