package fallback

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// Platform note: this file's locking is built on syscall.Flock, a Unix
// advisory-lock primitive. It is available on both of this project's two
// targets, macOS and Linux (see design.md's own deployment story -- a
// docker-compose Postgres stack plus a launchd-managed daemon, neither of
// which implies Windows); there is no Windows build of myflow. This
// comment is the one place that assumption lives -- if a Windows target
// is ever added, this file is the one place that needs a
// platform-specific replacement (syscall.Flock has no equivalent in
// package syscall on windows).

// LockFilePath returns the sidecar advisory-lock file for the journal at
// journalPath: journalPath + ".lock". This file is created once, the
// first time anything locks that journal, and is never renamed or
// removed. See LockJournal's own doc comment for why a sidecar file --
// rather than an flock taken directly on the journal file itself -- is
// the only correct choice once internal/reconcile's rename-based
// compaction is in the picture.
func LockFilePath(journalPath string) string {
	return journalPath + ".lock"
}

// openLockFile opens (creating if necessary) the sidecar lock file for
// journalPath, without acquiring the flock itself.
func openLockFile(journalPath string) (*os.File, error) {
	lockPath := LockFilePath(journalPath)
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return nil, fmt.Errorf("fallback: create lock directory for %s: %w", lockPath, err)
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("fallback: open lock file %s: %w", lockPath, err)
	}
	return f, nil
}

// LockJournal blocks until it holds the exclusive advisory lock guarding
// journalPath's retire critical section (internal/reconcile's
// retirePrefix: read-current, write-temp, rename), returning a release
// function the caller must call exactly once when done. Blocking here is
// deliberate and safe: the critical section it guards is on the order of
// microseconds, and flock is released automatically by the kernel the
// instant the holding process exits or dies -- even mid-hold -- so there
// is no way a crash leaves this lock permanently held, and therefore no
// deadlock path a caller needs to defend against with a timeout of its
// own. (AppendJournalEntry, whose caller-facing guarantee is different,
// uses the bounded tryLockJournal below instead.)
//
// This locks a *sidecar* file (journalPath + ".lock"), never the journal
// file itself, and that choice is load-bearing, not incidental.  flock's
// lock is held against the underlying inode a file descriptor was opened
// against, not against the path string -- so an flock taken on an fd
// opened against the journal file would track the *pre-rename* inode, and
// the instant a retire's rename swaps a new inode in at that path (which
// is exactly what retirePrefix's compaction does), the lock silently
// stops protecting anything a future opener of that path acquires: two
// callers could both believe they hold "the" lock on the journal while
// actually holding flocks on two different, unrelated inodes. A lock file
// that is created once and never renamed does not have this failure mode
// -- every caller, for the lifetime of the journal, opens the same,
// stable inode.
//
// This also closes a race beyond the single-process one
// Reconciler.Run's own in-process mutex already serializes: two separate
// *processes* -- two myflowd instances, or a daemon racing a concurrent
// `myflow journal flush` -- retiring the same journal at the same time
// are not protected by any in-process mutex, which cannot span processes.
// This flock is what makes that cross-process case safe too; the
// in-process mutex is kept as well, since it is cheaper for the common,
// single-process case and this flock does not replace it.
func LockJournal(journalPath string) (unlock func(), err error) {
	f, err := openLockFile(journalPath)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("fallback: lock %s: %w", LockFilePath(journalPath), err)
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}, nil
}

// appendLockTimeout bounds how long AppendJournalEntry waits for
// journalPath's sidecar lock before giving up and writing without it.
//
// This is measured, not guessed: retirePrefix's own critical section
// (re-read the journal, write a temp file, fsync it, close it, rename it
// into place) is not the microseconds a first pass at this comment
// assumed -- fsync alone routinely costs low single-digit milliseconds on
// ordinary local disks, this project's own dev machine included (a
// throwaway benchmark of exactly that create+write+sync+rename sequence,
// run 500 times, averaged ~4.5ms). Sizing this timeout against a wrong,
// smaller estimate is exactly how F1's fix would have quietly reintroduced
// its own version of the bug: a bound tighter than the section it is
// meant to wait out turns "occasionally proceeds without the lock" into
// "routinely does," under exactly the sustained-contention conditions
// (`myflow journal flush` in a health-check or CI loop) F1's report named
// as the real exposure. 50ms -- roughly ten times the measured critical
// section -- leaves headroom for that section running slower than
// measured here without meaningfully being noticed: the CLI's own store
// round trip already budgets 2s (defaultTimeout, cmd/flow/state.go)
// before falling back to this path at all, so an additional worst-case
// 50ms on top of that is not the kind of latency a pipeline command's own
// human-facing timing would register. The never-block guarantee is still
// what actually bounds this -- AppendJournalEntry's caller must never be
// made to wait *unboundedly* -- and losing this race window's protection
// in the genuinely rare case of a double-timeout (lock still contended
// *and* the exact open/write race window landing in the same append) is
// strictly better than stalling the operator's pipeline on a contended
// lock.
const appendLockTimeout = 50 * time.Millisecond

// appendLockPollInterval is how often tryLockJournal retries its
// non-blocking flock attempt while waiting up to appendLockTimeout. Set
// well below the section it is polling for (see appendLockTimeout's own
// doc comment for that section's measured ~4.5ms cost) so a lock that
// frees up is noticed promptly rather than losing a large slice of
// appendLockTimeout's budget to coarse polling granularity.
const appendLockPollInterval = time.Millisecond

// tryLockJournal attempts to acquire journalPath's sidecar lock, retrying
// a non-blocking flock (LOCK_EX|LOCK_NB) every appendLockPollInterval
// until timeout elapses. Polling rather than a single blocking flock call
// racing a timer is deliberate: it keeps the caller's actual wait bounded
// by timeout to within one poll interval, with no leftover goroutine
// blocked on the kernel call past that bound (a blocking flock call has no
// timeout parameter of its own to cancel it).
//
// ok is false, with a nil unlock and a nil error, if the timeout elapsed
// without acquiring the lock -- AppendJournalEntry proceeds without it in
// that case, per the never-block guarantee appendLockTimeout's own doc
// comment names. A non-nil error means something other than contention
// went wrong opening the lock file itself (a permissions error, a full
// disk); AppendJournalEntry treats that the same way -- proceed anyway --
// since the lock is best-effort infrastructure, not a gate on the actual
// journal write.
func tryLockJournal(journalPath string, timeout time.Duration) (unlock func(), ok bool, err error) {
	f, err := openLockFile(journalPath)
	if err != nil {
		return nil, false, err
	}

	deadline := time.Now().Add(timeout)
	for {
		flockErr := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if flockErr == nil {
			return func() {
				_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
				_ = f.Close()
			}, true, nil
		}
		if !errors.Is(flockErr, syscall.EWOULDBLOCK) {
			_ = f.Close()
			return nil, false, fmt.Errorf("fallback: try-lock %s: %w", LockFilePath(journalPath), flockErr)
		}
		if time.Now().After(deadline) {
			_ = f.Close()
			return nil, false, nil
		}
		time.Sleep(appendLockPollInterval)
	}
}
