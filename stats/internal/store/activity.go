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

// EndSessionPermission reports a subagent's tool call in session sessionID.
// A permission notification carries no agent, so a subagent's prompt sets
// the session's own state, and the main thread -- waiting on that subagent --
// may report nothing for minutes. A subagent's tool call after it shows the
// prompt was answered, so a permission state becomes working, detail naming
// the call. Any other state is left alone -- a subagent's call answers no
// open question and wakes no idle main thread -- and no row is created.
// ponytail: any subagent's call ends the wait, since the notification names
// none; a prompt still open while a sibling subagent works reads as working.
func (s *Store) EndSessionPermission(ctx context.Context, sessionID, detail string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE session_activity SET state = 'working', detail = $2, updated_at = now()
		WHERE session_id = $1 AND state = 'permission'
	`, sessionID, detail)
	if err != nil {
		return fmt.Errorf("store: end session permission: %w", err)
	}
	return nil
}
