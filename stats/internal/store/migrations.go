package store

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const migrationsDir = "migrations"

// EmbeddedMigrationCount reports how many migration files are embedded in
// this build. It exists so a test asserting idempotency (running the
// migrator twice records each file exactly once) can check against the
// actual embedded set instead of a hardcoded literal that has to be
// hand-edited every time a migration is added.
func EmbeddedMigrationCount() (int, error) {
	entries, err := fs.ReadDir(migrationsFS, migrationsDir)
	if err != nil {
		return 0, fmt.Errorf("store: read embedded migrations: %w", err)
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() {
			n++
		}
	}
	return n, nil
}

// migrationsLockKey is an arbitrary, fixed advisory-lock key. Its only
// requirement is that it not collide with a key some other subsystem of
// this database takes — nothing here depends on its value beyond that.
const migrationsLockKey = 725_016_001

// dbConn is satisfied by both *pgxpool.Pool and *pgxpool.Conn, so
// migrationApplied and applyMigration can run either against the pool or
// against the single connection RunMigrations holds the advisory lock on.
type dbConn interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

// RunMigrations applies every embedded migration that has not yet been
// recorded as applied, in lexical filename order, inside its own
// transaction, and records it in schema_migrations. Running it again is a
// no-op.
//
// The whole run is wrapped in a Postgres advisory lock held on one
// connection, so concurrent callers against the same fresh database
// serialize instead of racing DDL from two sessions at once.
func (s *Store) RunMigrations(ctx context.Context) (err error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("store: acquire connection for migrations: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", int64(migrationsLockKey)); err != nil {
		return fmt.Errorf("store: acquire migration lock: %w", err)
	}
	defer func() {
		// pg_advisory_unlock runs with context.Background rather than ctx:
		// a canceled or timed-out ctx must not leave the lock held forever.
		if _, unlockErr := conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", int64(migrationsLockKey)); unlockErr != nil {
			unlockErr = fmt.Errorf("store: release migration lock: %w", unlockErr)
			if err != nil {
				err = errors.Join(err, unlockErr)
			} else {
				err = unlockErr
			}
		}
	}()

	if _, execErr := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename   TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); execErr != nil {
		return fmt.Errorf("store: create schema_migrations: %w", execErr)
	}

	entries, readErr := fs.ReadDir(migrationsFS, migrationsDir)
	if readErr != nil {
		return fmt.Errorf("store: read embedded migrations: %w", readErr)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		applied, appliedErr := migrationApplied(ctx, conn, name)
		if appliedErr != nil {
			return appliedErr
		}
		if applied {
			continue
		}

		sqlBytes, readFileErr := migrationsFS.ReadFile(path.Join(migrationsDir, name))
		if readFileErr != nil {
			return fmt.Errorf("store: read migration %s: %w", name, readFileErr)
		}

		if applyErr := applyMigration(ctx, conn, name, string(sqlBytes)); applyErr != nil {
			return fmt.Errorf("store: apply migration %s: %w", name, applyErr)
		}
	}

	return nil
}

func migrationApplied(ctx context.Context, conn dbConn, name string) (bool, error) {
	var exists bool
	err := conn.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE filename = $1)", name,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("store: check migration %s: %w", name, err)
	}
	return exists, nil
}

func applyMigration(ctx context.Context, conn dbConn, name, sqlText string) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	// Rollback after a successful Commit is a documented no-op in pgx; its
	// error carries nothing actionable here, so it is discarded explicitly
	// rather than checked.
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, sqlText); err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (filename) VALUES ($1)", name); err != nil {
		return fmt.Errorf("record: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
