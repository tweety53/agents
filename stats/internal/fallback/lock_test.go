package fallback_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/fallback"
)

// TestAppendJournalEntryDoesNotHangWhenLockHeld pins the never-block half
// of F1's fix: when journalPath's sidecar lock is held for longer than
// AppendJournalEntry's bounded wait, the append must still return promptly
// -- proceeding without the lock -- rather than waiting for the holder to
// release it. Holding the lock from the test for well longer than the
// bound and asserting AppendJournalEntry returns (and succeeds) long
// before the holder releases it is what actually pins this, rather than
// asserting only that it eventually returns -- a genuinely blocking
// implementation would also "eventually" return once the test's own
// goroutine releases the lock 300ms later, which would make that weaker
// assertion pass regardless of whether the bound was honored.
func TestAppendJournalEntryDoesNotHangWhenLockHeld(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "held-proj", "held-chg.journal")

	const holdFor = 300 * time.Millisecond

	unlock, err := fallback.LockJournal(path)
	if err != nil {
		t.Fatalf("LockJournal: %v", err)
	}
	released := make(chan struct{})
	go func() {
		time.Sleep(holdFor)
		unlock()
		close(released)
	}()
	t.Cleanup(func() { <-released })

	start := time.Now()
	body := []byte(`{"state":"STARTED","mainCheckoutPath":"/tmp/lock-test","updatedAt":"2026-08-13T10:00:00Z","updatedBy":"tester"}`)
	if err := fallback.AppendJournalEntry(path, "held-proj", "held-chg", body, time.Now()); err != nil {
		t.Fatalf("AppendJournalEntry while lock held: %v", err)
	}
	elapsed := time.Since(start)

	// The lock is not released until holdFor has passed; if
	// AppendJournalEntry returned in comfortably less than that, it must
	// have proceeded without the lock rather than waiting it out.
	if elapsed >= holdFor/2 {
		t.Fatalf("AppendJournalEntry took %v while a %v hold was in progress -- it must give up on the lock well within its bound, not wait the hold out", elapsed, holdFor)
	}

	entries, err := fallback.ReadJournalEntries(path)
	if err != nil {
		t.Fatalf("ReadJournalEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("pending entries = %d, want 1 -- the append must still succeed even without the lock", len(entries))
	}
	if entries[0].Project != "held-proj" || entries[0].Name != "held-chg" {
		t.Fatalf("entry = %+v, want project/name held-proj/held-chg", entries[0])
	}
}
