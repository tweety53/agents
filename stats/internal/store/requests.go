package store

// requests.go — the requests table (0056_requests.sql): one row per request
// a /flow-interactive session works (KAN-1061). Both writes are upserts by
// (change, n), so a journalled replay lands on the row the first attempt
// wrote; a request's state is never stored, only derived by requestState.

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/tweety53/agents/stats/internal/records"
)

// ErrInvalidRequest is a caller mistake: a non-positive n, an outcome
// outside landed/answered/dropped, or shas that do not match the outcome
// (landed carries both, answered and dropped neither). Refused here, before
// the store is touched, so the route answers 400 rather than 500.
var ErrInvalidRequest = errors.New("store: invalid request record")

// requestColumns is what every request read returns, in scanRequest's
// order. Its last column is whether any stage run carries the request --
// requestState's started input. It is written against the requests table
// by name, so it serves a RETURNING clause and a SELECT alike.
const requestColumns = `
	requests.n, requests.title, requests.brainstorm_reason, requests.received_at, requests.ended_at,
	COALESCE(requests.outcome, ''), COALESCE(requests.commit_sha, ''), COALESCE(requests.base_sha, ''),
	EXISTS (SELECT 1 FROM stage_runs sr WHERE sr.change_id = requests.change_id AND sr.request_n = requests.n)`

func scanRequest(row rowScanner) (records.Request, error) {
	var (
		r       records.Request
		started bool
	)
	if err := row.Scan(&r.N, &r.Title, &r.BrainstormReason, &r.ReceivedAt, &r.EndedAt,
		&r.Outcome, &r.CommitSHA, &r.BaseSHA, &started); err != nil {
		return records.Request{}, err
	}
	r.State = requestState(r.Outcome, started)
	return r, nil
}

// requestState is the one derivation of a request's state (design.md's
// Store, CLI and API): an ended request is its outcome; one any stage run
// carries is in progress; any other is queued.
func requestState(outcome string, started bool) string {
	switch {
	case outcome != "":
		return outcome
	case started:
		return "in progress"
	default:
		return "queued"
	}
}

// BeginRequest records a request's receipt, or amends one already
// recorded: an empty Title or BrainstormReason leaves the stored value, and
// received_at is kept from the first write. It also bumps the change's
// updated_at to the begin's instant, so a long session stays inside the
// board's period. A nil ReceivedAt is stamped now(). An unknown
// (projectKey, change) pair is ErrChangeNotFound.
func (s *Store) BeginRequest(ctx context.Context, projectKey, change string, in records.RequestBegin) (records.Request, error) {
	if in.N <= 0 {
		return records.Request{}, fmt.Errorf("%w: n must be positive, got %d", ErrInvalidRequest, in.N)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return records.Request{}, fmt.Errorf("store: begin request: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	out, err := scanRequest(tx.QueryRow(ctx, `
		INSERT INTO requests (change_id, n, title, brainstorm_reason, received_at)
		SELECT c.id, $3, $4, $5, COALESCE($6, now())
		FROM changes c
		WHERE c.project_key = $1 AND c.name = $2
		ON CONFLICT (change_id, n) DO UPDATE SET
			title = CASE WHEN EXCLUDED.title = '' THEN requests.title ELSE EXCLUDED.title END,
			brainstorm_reason = CASE WHEN EXCLUDED.brainstorm_reason = '' THEN requests.brainstorm_reason
			                         ELSE EXCLUDED.brainstorm_reason END
		RETURNING `+requestColumns,
		projectKey, change, in.N, in.Title, in.BrainstormReason, in.ReceivedAt))
	if err != nil {
		return records.Request{}, requestWriteErr("begin", projectKey, change, err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE changes SET updated_at = GREATEST(updated_at, COALESCE($3, now()))
		WHERE project_key = $1 AND name = $2
	`, projectKey, change, in.ReceivedAt); err != nil {
		return records.Request{}, fmt.Errorf("store: begin request for %s/%s: bump change: %w", projectKey, change, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return records.Request{}, fmt.Errorf("store: begin request for %s/%s: commit: %w", projectKey, change, err)
	}
	return out, nil
}

// EndRequest records a request's end: its outcome and, for landed, the
// landed range's commit and base. An end for a request never begun inserts
// it, receipt stamped at the end's instant. A nil EndedAt is stamped now().
// An unknown (projectKey, change) pair is ErrChangeNotFound.
func (s *Store) EndRequest(ctx context.Context, projectKey, change string, in records.RequestEnd) (records.Request, error) {
	if in.N <= 0 {
		return records.Request{}, fmt.Errorf("%w: n must be positive, got %d", ErrInvalidRequest, in.N)
	}
	switch in.Outcome {
	case "landed":
		if in.CommitSHA == "" || in.BaseSHA == "" {
			return records.Request{}, fmt.Errorf("%w: landed needs both commit and base", ErrInvalidRequest)
		}
	case "answered", "dropped":
		if in.CommitSHA != "" || in.BaseSHA != "" {
			return records.Request{}, fmt.Errorf("%w: %s takes neither commit nor base", ErrInvalidRequest, in.Outcome)
		}
	default:
		return records.Request{}, fmt.Errorf("%w: outcome %q is not landed, answered or dropped", ErrInvalidRequest, in.Outcome)
	}

	out, err := scanRequest(s.pool.QueryRow(ctx, `
		INSERT INTO requests (change_id, n, received_at, ended_at, outcome, commit_sha, base_sha)
		SELECT c.id, $3, COALESCE($4, now()), COALESCE($4, now()), $5, NULLIF($6, ''), NULLIF($7, '')
		FROM changes c
		WHERE c.project_key = $1 AND c.name = $2
		ON CONFLICT (change_id, n) DO UPDATE SET
			ended_at = EXCLUDED.ended_at,
			outcome = EXCLUDED.outcome,
			commit_sha = EXCLUDED.commit_sha,
			base_sha = EXCLUDED.base_sha
		RETURNING `+requestColumns,
		projectKey, change, in.N, in.EndedAt, in.Outcome, in.CommitSHA, in.BaseSHA))
	if err != nil {
		return records.Request{}, requestWriteErr("end", projectKey, change, err)
	}
	return out, nil
}

func requestWriteErr(verb, projectKey, change string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %s/%s", ErrChangeNotFound, projectKey, change)
	}
	return fmt.Errorf("store: %s request for %s/%s: %w", verb, projectKey, change, err)
}

// ListRequests reads a change's requests in n order. An unknown
// (projectKey, change) pair reads as no requests, as ListTaskCounts does.
func (s *Store) ListRequests(ctx context.Context, projectKey, change string) ([]records.Request, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+requestColumns+`
		FROM requests
		JOIN changes c ON c.id = requests.change_id
		WHERE c.project_key = $1 AND c.name = $2
		ORDER BY requests.n
	`, projectKey, change)
	if err != nil {
		return nil, fmt.Errorf("store: list requests for %s/%s: %w", projectKey, change, err)
	}
	defer rows.Close()

	var out []records.Request
	for rows.Next() {
		r, err := scanRequest(rows)
		if err != nil {
			return nil, fmt.Errorf("store: scan request for %s/%s: %w", projectKey, change, err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list requests for %s/%s: %w", projectKey, change, err)
	}
	return out, nil
}
