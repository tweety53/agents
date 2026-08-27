// Package store is the only package in this module that builds SQL. It
// exposes typed repository methods; callers never see a query or a query
// builder.
package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool limits, per the project's adopted go-database standard. flowd is a
// single local daemon serving a handful of concurrent skill invocations, not
// a fleet of app instances sharing one database — these are sized for that
// use, not for a high-traffic service.
const (
	_maxConns        = 10
	_minConns        = 2
	_maxConnLifetime = 10 * time.Minute
	_maxConnIdleTime = 2 * time.Minute
)

// Store owns a connection pool to the flow PostgreSQL database and
// exposes the typed repository methods every caller uses instead of SQL.
type Store struct {
	pool *pgxpool.Pool
}

// Open connects to dsn, verifies the connection with a ping, and returns a
// Store backed by an explicitly-bounded pool. The caller must call Close
// when done.
func Open(ctx context.Context, dsn string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("store: parse dsn: %w", err)
	}
	cfg.MaxConns = _maxConns
	cfg.MinConns = _minConns
	cfg.MaxConnLifetime = _maxConnLifetime
	cfg.MaxConnIdleTime = _maxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("store: open pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("store: ping: %w", err)
	}
	return &Store{pool: pool}, nil
}

// Close releases the underlying connection pool.
func (s *Store) Close() {
	s.pool.Close()
}

// Ping reports whether the store's connection pool can currently reach the
// database. internal/reconcile's Reconciler.Watch uses this to detect the
// startup-down-then-up transition design.md calls "regains a database
// connection" -- the trigger for an out-of-band replay beyond the one at
// daemon startup.
func (s *Store) Ping(ctx context.Context) error {
	if err := s.pool.Ping(ctx); err != nil {
		return fmt.Errorf("store: ping: %w", err)
	}
	return nil
}

// MigrationCount reports how many migrations have been recorded as
// applied. It exists for verifying idempotency; production callers have no
// need for it.
func (s *Store) MigrationCount(ctx context.Context) (int, error) {
	var n int
	if err := s.pool.QueryRow(ctx, "SELECT count(*) FROM schema_migrations").Scan(&n); err != nil {
		return 0, fmt.Errorf("store: count schema_migrations: %w", err)
	}
	return n, nil
}
