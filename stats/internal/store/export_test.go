package store

import "github.com/jackc/pgx/v5/pgxpool"

// PoolOf exposes a Store's pool to the external tests.
func PoolOf(s *Store) *pgxpool.Pool { return s.pool }
