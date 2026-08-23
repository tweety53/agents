package store

import (
	"context"
	"fmt"
	"time"

	"github.com/tweety53/agents/stats/internal/harvest"
)

// RecordSessionTokenGiveUp persists that the watcher has given up on token,
// upserting rather than inserting: a token recorded a second time (the
// watcher retrying it after a restart and still failing) updates the
// existing row's reason and instant and increments retries, rather than
// being rejected or silently duplicated. retries starts at 0 on the first
// recording -- nothing has retried this token yet -- and rises by exactly
// one on every subsequent call, which is what PersistedGiveUps' caller
// reads to tell a fresh give-up from one the watcher has already retried.
func (s *Store) RecordSessionTokenGiveUp(ctx context.Context, token, reason string, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO session_token_giveups (session_token, reason, gave_up_at, retries)
		VALUES ($1, $2, $3, 0)
		ON CONFLICT (session_token) DO UPDATE
		SET reason = EXCLUDED.reason,
		    gave_up_at = EXCLUDED.gave_up_at,
		    retries = session_token_giveups.retries + 1
	`, token, reason, at)
	if err != nil {
		return fmt.Errorf("store: record give-up for session token %s: %w", token, err)
	}
	return nil
}

// PersistedGiveUps returns every session token the watcher has ever given
// up on, in no particular order -- the read the watcher makes at start to
// re-seed the tokens a restart's fresh in-memory pending set would
// otherwise never search again (0013_session_token_giveups.sql).
//
// It returns harvest.GiveUp directly, with no store-local type and no
// adapter, exactly as DispatchWindowsForSession (records.go) already
// returns harvest.DispatchWindow: the give-up type is declared in
// internal/harvest, not here, so *Store satisfies the widened
// harvest.SessionTokenBinder with no adapter either (cmd/myflowd's own
// compile-time check). The dependency runs store -> harvest, never the
// reverse, which is what keeps internal/harvest's TestHarvestNeedsNoDatabase
// true even as that interface grows.
func (s *Store) PersistedGiveUps(ctx context.Context) ([]harvest.GiveUp, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT session_token, reason, retries FROM session_token_giveups
	`)
	if err != nil {
		return nil, fmt.Errorf("store: list persisted give-ups: %w", err)
	}
	defer rows.Close()

	var giveUps []harvest.GiveUp
	for rows.Next() {
		var g harvest.GiveUp
		if err := rows.Scan(&g.Token, &g.Reason, &g.Retries); err != nil {
			return nil, fmt.Errorf("store: scan persisted give-up: %w", err)
		}
		giveUps = append(giveUps, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list persisted give-ups: %w", err)
	}
	return giveUps, nil
}
