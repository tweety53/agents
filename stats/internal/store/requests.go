package store

// requests.go — the requests table (0056_requests.sql): one row per request
// a /flow-interactive session works (KAN-1061). Both writes are upserts by
// (change, n), so a journalled replay lands on the row the first attempt
// wrote, and neither changes a row already ended (ErrRequestEnded); a request's state is never stored, only derived by requestState.

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

// ErrRequestEnded is a begin or end that would change a request already
// ended: a different title or brainstorm reason, or a different outcome,
// commit or base (kan-1070 self-review, F2: a second session numbering from
// 1 overwrote ended requests 1-5). A write identical to what is stored is
// accepted unchanged, so a journalled replay still succeeds.
var ErrRequestEnded = errors.New("store: request already ended")

// requestStarted is whether any stage run carries the request --
// requestState's started input. It is written against the requests table
// by name, so it serves a RETURNING clause, a SELECT and the live-state
// board's aggregate alike.
const requestStarted = `EXISTS (SELECT 1 FROM stage_runs sr WHERE sr.change_id = requests.change_id AND sr.request_n = requests.n)`

// requestColumns is what every request read returns, in scanRequest's
// order, its last column requestStarted.
const requestColumns = `
	requests.n, requests.title, requests.brainstorm_reason, requests.received_at, requests.ended_at,
	COALESCE(requests.outcome, ''), COALESCE(requests.commit_sha, ''), COALESCE(requests.base_sha, ''),
	` + requestStarted

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

// endedRequest reads request n of the change when it has an outcome, locking
// the row for the rest of tx; ok is false when the row is absent or open.
func endedRequest(ctx context.Context, tx pgx.Tx, projectKey, change string, n int) (r records.Request, ok bool, err error) {
	r, err = scanRequest(tx.QueryRow(ctx, `
		SELECT `+requestColumns+`
		FROM requests
		JOIN changes c ON c.id = requests.change_id
		WHERE c.project_key = $1 AND c.name = $2 AND requests.n = $3 AND requests.outcome IS NOT NULL
		FOR UPDATE OF requests
	`, projectKey, change, n))
	if errors.Is(err, pgx.ErrNoRows) {
		return records.Request{}, false, nil
	}
	if err != nil {
		return records.Request{}, false, fmt.Errorf("store: read request %d for %s/%s: %w", n, projectKey, change, err)
	}
	return r, true, nil
}

// BeginRequest records a request's receipt, or amends one already
// recorded: an empty Title or BrainstormReason leaves the stored value, and
// received_at is kept from the first write. It also bumps the change's
// updated_at to the begin's instant, so a long session stays inside the
// board's period. A nil ReceivedAt is stamped now(). An unknown
// (projectKey, change) pair is ErrChangeNotFound. On an ended request a
// non-empty title or reason that differs from the stored one is
// ErrRequestEnded; any other begin returns the stored row unchanged.
func (s *Store) BeginRequest(ctx context.Context, projectKey, change string, in records.RequestBegin) (records.Request, error) {
	if in.N <= 0 {
		return records.Request{}, fmt.Errorf("%w: n must be positive, got %d", ErrInvalidRequest, in.N)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return records.Request{}, fmt.Errorf("store: begin request: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if stored, ok, err := endedRequest(ctx, tx, projectKey, change, in.N); err != nil {
		return records.Request{}, err
	} else if ok {
		if (in.Title != "" && in.Title != stored.Title) || (in.BrainstormReason != "" && in.BrainstormReason != stored.BrainstormReason) {
			return records.Request{}, fmt.Errorf("%w: %s/%s request %d ended %s; a begin may not change its title or reason", ErrRequestEnded, projectKey, change, in.N, stored.Outcome)
		}
		return stored, nil
	}

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
// An unknown (projectKey, change) pair is ErrChangeNotFound. On an ended
// request a different outcome, commit or base is ErrRequestEnded; the
// identical end returns the stored row unchanged.
func (s *Store) EndRequest(ctx context.Context, projectKey, change string, in records.RequestEnd) (records.Request, error) {
	if err := validateRequestEnd(in.N, in.Outcome, in.CommitSHA, in.BaseSHA); err != nil {
		return records.Request{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return records.Request{}, fmt.Errorf("store: end request: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if stored, ok, err := endedRequest(ctx, tx, projectKey, change, in.N); err != nil {
		return records.Request{}, err
	} else if ok {
		if in.Outcome != stored.Outcome || in.CommitSHA != stored.CommitSHA || in.BaseSHA != stored.BaseSHA {
			return records.Request{}, fmt.Errorf("%w: %s/%s request %d ended %s; an end may not change its outcome, commit or base", ErrRequestEnded, projectKey, change, in.N, stored.Outcome)
		}
		return stored, nil
	}

	out, err := scanRequest(tx.QueryRow(ctx, `
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
	if err := tx.Commit(ctx); err != nil {
		return records.Request{}, fmt.Errorf("store: end request for %s/%s: commit: %w", projectKey, change, err)
	}
	return out, nil
}

// validateRequestEnd refuses a non-positive n and an outcome/sha shape
// outside landed (both shas), answered and dropped (neither).
func validateRequestEnd(n int, outcome, commit, base string) error {
	if n <= 0 {
		return fmt.Errorf("%w: n must be positive, got %d", ErrInvalidRequest, n)
	}
	switch outcome {
	case "landed":
		if commit == "" || base == "" {
			return fmt.Errorf("%w: landed needs both commit and base", ErrInvalidRequest)
		}
	case "answered", "dropped":
		if commit != "" || base != "" {
			return fmt.Errorf("%w: %s takes neither commit nor base", ErrInvalidRequest, outcome)
		}
	default:
		return fmt.Errorf("%w: outcome %q is not landed, answered or dropped", ErrInvalidRequest, outcome)
	}
	return nil
}

// PutRequest inserts or overwrites request r.N of the change whole -- title,
// brainstorm reason, receipt, end, outcome, commit and base -- bypassing
// ErrRequestEnded. It is `flowd put-request`'s repair write (kan-1070
// self-review, F2), never a pipeline path. The shape is validated as
// EndRequest validates it; an unknown change is ErrChangeNotFound.
func (s *Store) PutRequest(ctx context.Context, projectKey, change string, r records.Request) error {
	if err := validateRequestEnd(r.N, r.Outcome, r.CommitSHA, r.BaseSHA); err != nil {
		return err
	}
	var n int
	err := s.pool.QueryRow(ctx, `
		INSERT INTO requests (change_id, n, title, brainstorm_reason, received_at, ended_at, outcome, commit_sha, base_sha)
		SELECT c.id, $3, $4, $5, $6, $7, $8, NULLIF($9, ''), NULLIF($10, '')
		FROM changes c
		WHERE c.project_key = $1 AND c.name = $2
		ON CONFLICT (change_id, n) DO UPDATE SET
			title = EXCLUDED.title,
			brainstorm_reason = EXCLUDED.brainstorm_reason,
			received_at = EXCLUDED.received_at,
			ended_at = EXCLUDED.ended_at,
			outcome = EXCLUDED.outcome,
			commit_sha = EXCLUDED.commit_sha,
			base_sha = EXCLUDED.base_sha
		RETURNING n
	`, projectKey, change, r.N, r.Title, r.BrainstormReason, r.ReceivedAt, r.EndedAt, r.Outcome, r.CommitSHA, r.BaseSHA).Scan(&n)
	if err != nil {
		return requestWriteErr("put", projectKey, change, err)
	}
	return nil
}

// SetStageRunRequest sets stage run id's request to n -- `flowd
// set-stage-request`'s repair write. A non-positive n is ErrInvalidRequest;
// an id matching no stage run is an error naming it.
func (s *Store) SetStageRunRequest(ctx context.Context, id int64, n int) error {
	if n <= 0 {
		return fmt.Errorf("%w: n must be positive, got %d", ErrInvalidRequest, n)
	}
	tag, err := s.pool.Exec(ctx, `UPDATE stage_runs SET request_n = $2 WHERE id = $1`, id, n)
	if err != nil {
		return fmt.Errorf("store: set stage run %d request: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("store: set stage run %d request: no such stage run", id)
	}
	return nil
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
