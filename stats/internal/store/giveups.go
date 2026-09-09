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

// giveUpRetryHorizon bounds how long a session-token give-up stays worth
// retrying. The watcher loads PersistedGiveUps at start and re-reads the
// whole transcript corpus per harvest cycle while any of them are loaded, so
// a dead give-up is not just useless but expensive: a give-up's only purpose
// is re-finding a session mark whose usage can still attribute inside a
// stage run's [started_at, ended_at) window, and those windows close within
// minutes-to-hours -- measured here, one full retry pass over the corpus
// ran 15+ minutes at ~95% CPU for 54 give-ups aged 1-18 days whose tokens
// all named long-archived changes (kan-480). 24h is generous by an order of
// magnitude: past it, the row is deleted rather than returned, which also
// keeps the table bounded instead of growing forever.
const giveUpRetryHorizon = 24 * time.Hour

// giveUpMaxRetries bounds how many full retry generations a give-up survives.
// RecordSessionTokenGiveUp's upsert increments retries on every re-give-up
// and refreshes gave_up_at with it, so a token that keeps failing stays
// forever-fresh and would re-seed the whole-corpus retry scan on every daemon
// start no matter how old its first failure was -- the age horizon alone
// cannot stop that (kan-481). Three full bounded windows without a bind is
// measured-hopeless, not bad luck: the rows cleared ahead of this change had
// climbed to retries 10.
const giveUpMaxRetries = 3

// PersistedGiveUps returns every session token the watcher has given up on
// within the retry horizon, in no particular order -- the read the watcher
// makes at start to re-seed the tokens a restart's fresh in-memory pending
// set would otherwise never search again (0013_session_token_giveups.sql).
// Rows older than giveUpRetryHorizon are deleted by the same call: their
// stage windows are long closed, so no retry could ever attribute their
// usage, and keeping them would make every restart re-read the whole
// transcript corpus for nothing.
//
// It returns harvest.GiveUp directly, with no store-local type and no
// adapter, exactly as DispatchWindowsForSession (records.go) already
// returns harvest.DispatchWindow: the give-up type is declared in
// internal/harvest, not here, so *Store satisfies the widened
// harvest.SessionTokenBinder with no adapter either (cmd/flowd's own
// compile-time check). The dependency runs store -> harvest, never the
// reverse, which is what keeps internal/harvest's TestHarvestNeedsNoDatabase
// true even as that interface grows.
func (s *Store) PersistedGiveUps(ctx context.Context) ([]harvest.GiveUp, error) {
	cutoff := time.Now().Add(-giveUpRetryHorizon)
	if _, err := s.pool.Exec(ctx, `
		DELETE FROM session_token_giveups
		WHERE gave_up_at < $1 OR retries >= $2
	`, cutoff, giveUpMaxRetries); err != nil {
		return nil, fmt.Errorf("store: expire persisted give-ups: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT session_token, reason, retries FROM session_token_giveups
		WHERE gave_up_at >= $1 AND retries < $2
	`, cutoff, giveUpMaxRetries)
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
