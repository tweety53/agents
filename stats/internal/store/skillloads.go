package store

import (
	"context"
	"fmt"
	"time"
)

// skillloads.go backs the skill-load telemetry (0111_skill_loads.sql): the flow-skill-load
// hook's write and the per-path counts `flow-guard prune-candidates` reads.

// SkillLoadCount is one path's loads in a window.
type SkillLoadCount struct {
	Path  string
	Loads int
}

// RecordSkillLoad records one load of path, repo-relative, by sessionID at loadedAt (now when zero).
// An empty path is refused by the table's CHECK.
func (s *Store) RecordSkillLoad(ctx context.Context, path, sessionID string, loadedAt time.Time) error {
	if loadedAt.IsZero() {
		loadedAt = time.Now()
	}
	if _, err := s.pool.Exec(ctx, `INSERT INTO skill_loads (path, session_id, loaded_at) VALUES ($1, $2, $3)`,
		path, sessionID, loadedAt); err != nil {
		return fmt.Errorf("store: record skill load: %w", err)
	}
	return nil
}

// SkillLoadCounts is each path loaded in [from, to) with its load count, ordered by path.
func (s *Store) SkillLoadCounts(ctx context.Context, from, to time.Time) ([]SkillLoadCount, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT path, COUNT(*)::int FROM skill_loads
		WHERE loaded_at >= $1 AND loaded_at < $2
		GROUP BY path ORDER BY path`, from, to)
	if err != nil {
		return nil, fmt.Errorf("store: skill load counts: %w", err)
	}
	defer rows.Close()
	out := []SkillLoadCount{}
	for rows.Next() {
		var c SkillLoadCount
		if err := rows.Scan(&c.Path, &c.Loads); err != nil {
			return nil, fmt.Errorf("store: skill load counts: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
