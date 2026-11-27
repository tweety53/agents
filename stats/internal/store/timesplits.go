package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Time-split row kinds, PendingTimeSplit.Kind.
const (
	SplitStage    = "stage"
	SplitDispatch = "dispatch"
	// SplitParent is a stage run's parent split: the split of the session
	// transcript its session id names, stored in stage_runs.parent_split
	// (0117_parent_splits.sql).
	SplitParent = "parent"
)

// PendingTimeSplit is one row the harvest tick still has work on
// (0115_time_splits.sql): a stage run with a cc- session token and no split
// -- open and not yet linked, or ended --, an ended dispatch with an
// agent id and no split, or an ended stage run with a session id and no
// parent split (0117_parent_splits.sql). Token and TranscriptPath are a
// stage run's; AgentID a dispatch's; SessionID a dispatch's or a parent
// split's.
type PendingTimeSplit struct {
	Kind           string
	ID             int64
	Token          string
	TranscriptPath string
	AgentID        string
	SessionID      string
	StartedAt      time.Time
	EndedAt        *time.Time
}

// PendingTimeSplits returns up to limit rows of each kind still pending,
// newest first, so a backlog drains a tick at a
// time while new rows go first.
func (s *Store) PendingTimeSplits(ctx context.Context, limit int) ([]PendingTimeSplit, error) {
	rows, err := s.pool.Query(ctx, `
		(SELECT 'stage', id, session_token, COALESCE(transcript_path, ''), '', '', started_at, ended_at
		FROM stage_runs
		WHERE time_split IS NULL AND session_token LIKE 'cc-%'
		  AND (ended_at IS NOT NULL OR transcript_path IS NULL)
		ORDER BY id DESC LIMIT $1)
		UNION ALL
		(SELECT 'dispatch', id, '', '', agent_id, COALESCE(session_id, ''), started_at, ended_at
		FROM dispatches
		WHERE time_split IS NULL AND ended_at IS NOT NULL AND agent_id IS NOT NULL AND agent_id <> 'inline'
		ORDER BY id DESC LIMIT $1)
		UNION ALL
		(SELECT 'parent', id, '', '', '', session_id, started_at, ended_at
		FROM stage_runs
		WHERE parent_split IS NULL AND session_id IS NOT NULL AND ended_at IS NOT NULL
		ORDER BY id DESC LIMIT $1)
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("store: pending time splits: %w", err)
	}
	defer rows.Close()
	var out []PendingTimeSplit
	for rows.Next() {
		var p PendingTimeSplit
		if err := rows.Scan(&p.Kind, &p.ID, &p.Token, &p.TranscriptPath, &p.AgentID, &p.SessionID, &p.StartedAt, &p.EndedAt); err != nil {
			return nil, fmt.Errorf("store: pending time splits: scan: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: pending time splits: %w", err)
	}
	return out, nil
}

// LinkStageRunTranscript sets stage run id's transcript_path where none is
// set yet.
func (s *Store) LinkStageRunTranscript(ctx context.Context, id int64, path string) error {
	if _, err := s.pool.Exec(ctx, `UPDATE stage_runs SET transcript_path = $2 WHERE id = $1 AND transcript_path IS NULL`, id, path); err != nil {
		return fmt.Errorf("store: link stage run %d transcript: %w", id, err)
	}
	return nil
}

// SetTimeSplit stores split (JSON null: no transcript) on the kind row id,
// only while it has none and its window is still [from, to) -- a retime
// between the tick's read and this write leaves the row pending for the
// next tick.
func (s *Store) SetTimeSplit(ctx context.Context, kind string, id int64, from, to time.Time, split json.RawMessage) error {
	table, col := "dispatches", "time_split"
	switch kind {
	case SplitStage:
		table = "stage_runs"
	case SplitParent:
		table, col = "stage_runs", "parent_split"
	}
	if _, err := s.pool.Exec(ctx, `UPDATE `+table+` SET `+col+` = $2::jsonb
		WHERE id = $1 AND `+col+` IS NULL AND started_at = $3 AND ended_at = $4`, id, split, from, to); err != nil {
		return fmt.Errorf("store: set %s %d time split: %w", kind, id, err)
	}
	return nil
}

// storedSplit is a time_split column as read: nil for SQL NULL and for the
// JSON null of a row whose transcript was never found.
func storedSplit(b []byte) json.RawMessage {
	if b == nil || string(b) == "null" {
		return nil
	}
	return b
}
