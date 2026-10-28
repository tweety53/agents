package store

import (
	"context"
	"fmt"
)

// SetSessionActivity records what session sessionID is doing now: one of
// working, question, permission or idle (the table's CHECK refuses any
// other), with detail naming what it waits on. updated_at moves only when
// state changes, so a repeated working report keeps the time the session
// started working and the board's wait reads from it.
func (s *Store) SetSessionActivity(ctx context.Context, sessionID, state, detail string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO session_activity (session_id, state, detail) VALUES ($1, $2, $3)
		ON CONFLICT (session_id) DO UPDATE SET
			updated_at = CASE WHEN session_activity.state = EXCLUDED.state
			                  THEN session_activity.updated_at ELSE now() END,
			state = EXCLUDED.state,
			detail = EXCLUDED.detail
	`, sessionID, state, detail)
	if err != nil {
		return fmt.Errorf("store: set session activity: %w", err)
	}
	return nil
}
