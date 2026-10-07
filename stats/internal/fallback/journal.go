package fallback

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry is one journal record: a whole-object write that could not reach
// the store. Task 6's reconciler replays a journal by reading every Entry
// in file order and re-issuing Body as the PUT it originally was -- Body
// is exactly the bytes `state set` was given, the same shape as the
// on-disk state file, so replay needs no translation step.
//
// The format is JSON Lines: one Entry per line, appended and never
// rewritten in place. Replay order is file order -- each append is a
// single O_APPEND write, which POSIX guarantees is atomic with respect to
// other appends to the same file, so the sequence entries land in a
// journal is exactly the sequence concurrent CLI invocations issued their
// writes in. That is deliberately not recorded as a separate field:
// parallel worktrees against the same project are a real, expected shape
// of this pipeline (not a hypothetical), so concurrent `flow state set`
// processes appending to one journal is the normal case, and any field
// computed from "entries already in the file" (a prior version of this
// type carried a Sequence assigned as len(existing)+1) is a read-then-act
// race across processes -- twenty concurrent appends produced fourteen
// entries claiming to be first and four claiming to be fifteenth. Rather
// than add locking to defend a field whose only job was documenting an
// order the file's own byte order already gives for free, the field is
// removed.
//
// RecordedAt exists only for operator-facing diagnostics; the authority
// reconciliation actually resolves conflicts by is each entry's own
// Body.updatedAt, per design.md's "Availability and reconciliation"
// ("conflicts resolve by updated_at, with the monotonic-state rule as the
// tiebreaker").
type Entry struct {
	RecordedAt time.Time       `json:"recordedAt"`
	Project    string          `json:"project"`
	Name       string          `json:"name"`
	Body       json.RawMessage `json:"body"`
}

// AppendJournalEntry appends one Entry to the journal at path, creating the
// file and its parent directory if needed.
//
// Concurrent CLI invocations against the same journal -- parallel
// worktrees taking the fallback path at the same time -- are expected, not
// hypothetical, so appender-versus-appender safety still relies only on
// O_APPEND's atomicity guarantee: no read-modify-write step between two
// appenders, no in-memory state shared across calls.
//
// What this function *does* take, before opening path, is journalPath's
// sidecar advisory lock (LockJournal, lock.go), held across open+write.
// This exists for a different race than appender-versus-appender:
// appender-versus-*retirer*. internal/reconcile's retirePrefix compacts
// the journal by writing a temp file and renaming it over path; if this
// function's os.OpenFile below resolves the pre-rename inode, and its
// Write lands after retirePrefix re-read that inode, the appended bytes go
// to an inode the rename then unlinks -- lost silently, since every caller
// discards this function's error. retirePrefix holds the same lock across
// its own re-read/write-temp/rename section, so holding it here closes
// that window outright.
//
// The wait for the lock is unbounded on purpose. A previous version gave
// up after 50ms and wrote unlocked, on the premise that the retire section
// never runs that long; it does -- File.Sync on macOS is F_FULLFSYNC, and
// under load (twelve parallel race-detector runs of
// TestConcurrentAppendVersusRetirePreservesEveryEntry) roughly 35 of 600
// appends timed out per run, and every one of 52 runs lost between 4 and
// 24 entries. The wait is bounded in practice by retirePrefix's
// local-disk section, and flock is released by the kernel the instant
// its holder exits or dies, so no crash leaves an appender waiting.
//
// A failure to open the lock file itself (permissions, a full disk) still
// proceeds without the lock: retirePrefix opens the same file and fails
// the same way, refusing to retire, so no rename can race this write.
func AppendJournalEntry(path, project, name string, body []byte, recordedAt time.Time) (err error) {
	entry := Entry{
		RecordedAt: recordedAt.UTC(),
		Project:    project,
		Name:       name,
		Body:       append(json.RawMessage(nil), body...),
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("fallback: encode journal entry: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("fallback: create journal directory for %s: %w", path, err)
	}

	if unlock, lockErr := LockJournal(path); lockErr == nil {
		defer unlock()
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("fallback: open journal %s: %w", path, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("fallback: close journal %s: %w", path, cerr)
		}
	}()

	if _, err = f.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("fallback: append journal entry to %s: %w", path, err)
	}
	return nil
}

// ReadJournalEntries reads every Entry in path, in file order -- which is
// replay order, per Entry's own doc comment. A journal that does not yet
// exist is not an error -- it reads as no pending entries, exactly like a
// change that has never taken the fallback path.
func ReadJournalEntries(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("fallback: read journal %s: %w", path, err)
	}

	var entries []Entry
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("fallback: decode journal entry in %s: %w", path, err)
		}
		entries = append(entries, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("fallback: scan journal %s: %w", path, err)
	}
	return entries, nil
}
