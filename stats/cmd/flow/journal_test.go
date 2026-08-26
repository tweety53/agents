package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/store"
)

// adminDSN is the DSN used to create and drop per-test databases -- the
// same dedicated myflow-postgres compose stack every other package's tests
// use.
func adminDSN() string {
	if v := os.Getenv("FLOW_STATS_ADMIN_DSN"); v != "" {
		return v
	}
	return "postgres://myflow:myflow@localhost:5433/myflow?sslmode=disable"
}

// newTestStoreDSN creates a uniquely-named, migrated database against the
// compose stack and returns its DSN -- runJournalFlush opens its own store
// from a DSN string (it is the one CLI command that talks to the store
// directly; see journal.go's own doc comment), so this test needs the
// connection string itself, not a *store.Store. It skips cleanly when the
// stack is not reachable and registers a cleanup that drops the database.
func newTestStoreDSN(t *testing.T) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	adminPool, err := pgxpool.New(ctx, adminDSN())
	if err != nil {
		t.Skipf("myflow-postgres compose stack not reachable: %v", err)
	}
	if err := adminPool.Ping(ctx); err != nil {
		adminPool.Close()
		t.Skipf("myflow-postgres compose stack not reachable: %v", err)
	}

	dbName := fmt.Sprintf("myflow_test_%d_%d", os.Getpid(), time.Now().UnixNano())
	ident := pgx.Identifier{dbName}.Sanitize()
	if _, err := adminPool.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
		adminPool.Close()
		t.Fatalf("create test database %s: %v", dbName, err)
	}
	adminPool.Close()

	t.Cleanup(func() {
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer dropCancel()
		dropPool, err := pgxpool.New(dropCtx, adminDSN())
		if err != nil {
			t.Logf("drop test database %s: reconnect failed: %v", dbName, err)
			return
		}
		defer dropPool.Close()
		if _, err := dropPool.Exec(dropCtx, "DROP DATABASE IF EXISTS "+ident+" WITH (FORCE)"); err != nil {
			t.Logf("drop test database %s: %v", dbName, err)
		}
	})

	dsn := fmt.Sprintf("postgres://myflow:myflow@localhost:5433/%s?sslmode=disable", dbName)

	openCtx, openCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer openCancel()
	st, err := store.Open(openCtx, dsn)
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	if err := st.RunMigrations(openCtx); err != nil {
		st.Close()
		t.Fatalf("run migrations: %v", err)
	}
	st.Close()

	return dsn
}

// TestJournalFlushCommandReplaysOnDemand pins `myflow journal flush`
// end-to-end: a pending journal entry on disk is replayed into a real
// store when the command runs, reported on stdout, and retired from the
// journal -- without myflowd running at all, since this command talks to
// the store directly (journal.go's own doc comment).
func TestJournalFlushCommandReplaysOnDemand(t *testing.T) {
	dsn := newTestStoreDSN(t)
	root := t.TempDir()

	journalPath := filepath.Join(root, "proj-flush", "chg-flush.journal")
	body := []byte(`{"state":"STARTED","mainCheckoutPath":"/tmp/journal-flush-test","updatedAt":"2026-08-13T10:00:00Z","updatedBy":"myflow-start"}`)
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

	openCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
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

// TestJournalFlushCommandUnknownSubcommand asserts `myflow journal <bogus>`
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
		[]string{"flush", "-root", root, "-dsn", "postgres://myflow:myflow@127.0.0.1:1/myflow?sslmode=disable", "-timeout", "200ms"},
		&stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1; stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}
