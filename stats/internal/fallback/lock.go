package fallback

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Platform note: this file's locking is built on syscall.Flock, a Unix
// advisory-lock primitive. It is available on both of this project's two
// targets, macOS and Linux (see design.md's own deployment story -- a
// docker-compose Postgres stack plus a launchd-managed daemon, neither of
// which implies Windows); there is no Windows build of flow. This
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
// own. AppendJournalEntry takes this same lock around its open+write.
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
// *processes* -- two flowd instances, or a daemon racing a concurrent
// `flow journal flush` -- retiring the same journal at the same time
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
