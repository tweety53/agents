package store

// taskcounts.go — the change_task_counts table (0026_task_counts.sql):
// one append-only row per observation of a change plan's total task
// count. Nothing here upserts; the series is the trend, so every insert
// is a new observation exactly as recordPass's routes treat their rows.

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/tweety53/agents/stats/internal/records"
)

// ErrInvalidTaskCount is a caller mistake: a total that is not a positive
// integer counts no plan, and the CHECK constraint in the migration would
// refuse it anyway -- refused here first, before the store is touched, so
// the route answers 400 rather than 500.
var ErrInvalidTaskCount = errors.New("store: task count needs a positive total")

// RecordTaskCount records one observation of a change plan's total task
// count. An unknown (projectKey, change) pair is ErrChangeNotFound, the
// same resolution RecordDecision applies.
func (s *Store) RecordTaskCount(ctx context.Context, projectKey, change string, in records.TaskCount) (records.TaskCount, error) {
	if in.TotalTasks <= 0 {
		return records.TaskCount{}, fmt.Errorf("%w: got total %d", ErrInvalidTaskCount, in.TotalTasks)
	}

	var out records.TaskCount
	err := s.pool.QueryRow(ctx, `
		INSERT INTO change_task_counts (change_id, total_tasks)
		SELECT c.id, $3
		FROM changes c
		WHERE c.project_key = $1 AND c.name = $2
		RETURNING id, total_tasks, observed_at
	`, projectKey, change, in.TotalTasks).
		Scan(&out.ID, &out.TotalTasks, &out.ObservedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return records.TaskCount{}, fmt.Errorf("%w: %s/%s", ErrChangeNotFound, projectKey, change)
		}
		return records.TaskCount{}, fmt.Errorf("store: record task count for %s/%s: %w", projectKey, change, err)
	}
	return out, nil
}

// ListTaskCounts reads a change's observations oldest first -- the order
// the planned (first) and appended (latest minus first) derivations and
// the state board's lateral joins both read the series in.
func (s *Store) ListTaskCounts(ctx context.Context, projectKey, change string) ([]records.TaskCount, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.id, t.total_tasks, t.observed_at
		FROM change_task_counts t
		JOIN changes c ON c.id = t.change_id
		WHERE c.project_key = $1 AND c.name = $2
		ORDER BY t.observed_at, t.id
	`, projectKey, change)
	if err != nil {
		return nil, fmt.Errorf("store: list task counts for %s/%s: %w", projectKey, change, err)
	}
	defer rows.Close()

	var out []records.TaskCount
	for rows.Next() {
		var tc records.TaskCount
		if err := rows.Scan(&tc.ID, &tc.TotalTasks, &tc.ObservedAt); err != nil {
			return nil, fmt.Errorf("store: scan task count for %s/%s: %w", projectKey, change, err)
		}
		out = append(out, tc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list task counts for %s/%s: %w", projectKey, change, err)
	}
	return out, nil
}
