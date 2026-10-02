package store

import (
	"context"
	"fmt"
)

// ProjectRoot is one registered project's identity and the main checkout
// path its flow calls recorded. The path is what a workspace-wide read
// scans: the store's one answer to "where do this machine's repositories
// live", learned from the same calls that created every other row, never
// configured separately.
type ProjectRoot struct {
	ProjectKey       string
	MainCheckoutPath string
}

// ProjectRoots returns every registered project's main checkout path, in
// project-key order -- every row, always; filtering is the caller's act.
// A project whose row predates its directory's removal still reads back:
// what the path points at is the scanner's concern, not the store's.
func (s *Store) ProjectRoots(ctx context.Context) ([]ProjectRoot, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT project_key, main_checkout_path
		FROM projects
		ORDER BY project_key
	`)
	if err != nil {
		return nil, fmt.Errorf("store: list project roots: %w", err)
	}
	defer rows.Close()

	var roots []ProjectRoot
	for rows.Next() {
		var r ProjectRoot
		if err := rows.Scan(&r.ProjectKey, &r.MainCheckoutPath); err != nil {
			return nil, fmt.Errorf("store: scan project root: %w", err)
		}
		roots = append(roots, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list project roots: %w", err)
	}
	return roots, nil
}
