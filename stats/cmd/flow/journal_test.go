package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/store"
	"github.com/tweety53/agents/stats/internal/storetest"
)

// TestMain drops the package's template database once every test has run.
func TestMain(m *testing.M) { os.Exit(storetest.Main(m)) }

// newTestStoreDSN returns the DSN of a throwaway, migrated database -- a
// copy of the package's one migrated template (internal/storetest) --
// because runJournalFlush opens its own store from a DSN string (it is the
// one CLI command that talks to the store directly; see journal.go's own doc
// comment). It skips cleanly when the stack is not reachable and registers
// the database's drop as test cleanup.
func newTestStoreDSN(t *testing.T) string {
	t.Helper()
	return storetest.NewMigratedDatabase(t)
}

// TestJournalFlushCommandReplaysOnDemand pins `flow journal flush`
// end-to-end: a pending journal entry on disk is replayed into a real
// store when the command runs, reported on stdout, and retired from the
// journal -- without flowd running at all, since this command talks to
// the store directly (journal.go's own doc comment).
func TestJournalFlushCommandReplaysOnDemand(t *testing.T) {
	dsn := newTestStoreDSN(t)
	root := t.TempDir()

	journalPath := filepath.Join(root, "proj-flush", "chg-flush.journal")
	body := []byte(`{"state":"STARTED","mainCheckoutPath":"/tmp/journal-flush-test","updatedAt":"2026-08-13T10:00:00Z","updatedBy":"flow-start"}`)
	if err := fallback.AppendJournalEntry(journalPath, "proj-flush", "chg-flush", body, time.Now()); err != nil {
		t.Fatalf("append journal entry: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := runJournal(context.Background(),
		[]string{"flush", "-root", root, "-dsn", dsn},
		&stdout, &stderr)

	if code != 0 {
		t.Fatalf("runJournal flush exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if got := stdout.String(); !bytes.Contains([]byte(got), []byte("1 applied")) {
		t.Fatalf("stdout = %q, want it to report 1 applied", got)
	}

	entries, err := fallback.ReadJournalEntries(journalPath)
	if err != nil {
		t.Fatalf("read journal entries after flush: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("pending entries after flush = %d, want 0", len(entries))
	}

	openCtx := context.Background()
	st, err := store.Open(openCtx, dsn)
	if err != nil {
		t.Fatalf("reopen store to verify: %v", err)
	}
	defer st.Close()

	got, err := st.GetChange(openCtx, "proj-flush", "chg-flush")
	if err != nil {
		t.Fatalf("GetChange after flush: %v", err)
	}
	if got.State != store.StateStarted {
		t.Fatalf("stored state = %v, want STARTED", got.State)
	}
}

// TestJournalFlushCommandUnknownSubcommand asserts `flow journal <bogus>`
// is rejected with usage rather than silently doing nothing.
func TestJournalFlushCommandUnknownSubcommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runJournal(context.Background(), []string{"bogus"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("expected usage/error output on stderr")
	}
}

// TestJournalFlushCommandConnectFailure asserts a dead DSN is reported and
// exits non-zero rather than silently doing nothing -- unlike state
// get/set, this command has no fallback to take instead (journal.go's own
// doc comment: there is nothing to fall back *to* for an explicit
// "reconcile now" request).
func TestJournalFlushCommandConnectFailure(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := runJournal(context.Background(),
		[]string{"flush", "-root", root, "-dsn", "postgres://flow:flow@127.0.0.1:1/flow?sslmode=disable", "-timeout", "200ms"},
		&stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1; stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}
