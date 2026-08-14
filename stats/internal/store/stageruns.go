package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrStageRunNotFound is returned by EndStage, MergeMetrics and Price when
// no stage run exists for the given id.
var ErrStageRunNotFound = errors.New("store: stage run not found")

// ErrNilMetricsPatch is returned by MergeMetrics when patch is nil. A nil
// json.RawMessage marshals to SQL NULL, and jsonb_deep_merge(a, NULL)
// returns NULL, which the metrics column's NOT NULL constraint would then
// reject -- a caller mistake surfaced as a raw Postgres error rather than
// a typed one. No current caller passes a nil patch, but MergeMetrics
// refuses it explicitly rather than leaving that guarantee to chance.
var ErrNilMetricsPatch = errors.New("store: metrics patch must not be nil")

// ErrTooManyAttemptCollisions is returned by BeginStage when it could not
// allocate an attempt number after retrying past every concurrent
// collision it was willing to absorb. It signals contention so extreme
// that the caller, not the store, should decide what to do next -- it is
// never returned for an ordinary single collision, which BeginStage
// resolves itself by retrying.
var ErrTooManyAttemptCollisions = errors.New("store: too many concurrent attempt collisions allocating a stage run")

// stageRunsAttemptConstraint is the name given, explicitly, to the
// UNIQUE (change_id, command, stage, attempt) constraint in
// 0003_stage_runs.sql. BeginStage checks for this exact constraint name
// rather than any unique-violation, so it never mistakes an unrelated
// conflict (for example on the stage_runs_repo_root_fk foreign key, which
// pgx also reports through the same error type) for the attempt race it
// knows how to retry.
const stageRunsAttemptConstraint = "stage_runs_attempt_key"

// maxAttemptRetries bounds how many times BeginStage retries after losing
// an attempt-number race before giving up with
// ErrTooManyAttemptCollisions. It is set well above
// TestConcurrentBeginStageDoesNotCollide's 20 concurrent writers -- every
// one of those 20 can lose a race at most once each (the 21st write to
// land always succeeds, since only one writer can still be contending once
// 19 have already committed their distinct attempt numbers), so 20 writers
// need at most 19 retries in the worst case. This bound leaves real margin
// above that worst case rather than sitting exactly on it, so raising the
// test's concurrency slightly does not require raising this constant in
// lockstep.
const maxAttemptRetries = 100

// StageRun is one recorded attempt at one stage of one command, for one
// change. A nil RepoRoot means the stage belongs to the change as a whole,
// never "unknown repository" -- the same absence-is-not-a-value rule the
// metrics bag follows.
type StageRun struct {
	ID        int64
	ChangeID  int64
	RepoRoot  *string
	Harness   string
	SessionID *string
	// SessionToken is the literal, unique correlator `stage begin` wrote
	// (KAN-172, task 1) -- present whenever SessionID is not yet resolved,
	// and left in place, unused, once binding has happened, since binding
	// is one-way (design.md's kan-172 "unbinding never happens"). A nil
	// value means this run predates the sessionToken column (the store's own
	// no-backfill rows) or, in principle, was marked by a harness build
	// that never sent one.
	SessionToken *string
	Command      string
	Stage        string
	Attempt      int
	StartedAt    time.Time
	EndedAt      *time.Time
	Outcome      *string
	Metrics      json.RawMessage
}

// BeginStageInput identifies the change a stage run belongs to by its
// public identity -- project and change name, exactly as GetChange and
// PutChange do -- rather than by the internal BIGSERIAL id, which never
// crosses the package boundary.
type BeginStageInput struct {
	ProjectKey string
	ChangeName string
	// RepoRoot is set when this stage ran inside one repository of a
	// multi-repository change, and left nil when it belongs to the change
	// as a whole.
	RepoRoot  *string
	Harness   string
	SessionID *string
	// SessionToken is the literal correlator to persist alongside the run,
	// so a later harvest cycle can find it in a transcript and bind
	// SessionID (KAN-172, task 2). Whether it is required at all is an
	// application-layer decision (internal/api and cmd/myflow both enforce
	// it); the store persists whatever it is given, nil included.
	SessionToken *string
	Command      string
	Stage        string
	StartedAt    time.Time
}

// BeginStage records the start of one stage run and allocates its attempt
// number for the (change, command, stage) triple, so that a fix re-run is
// recorded as a new attempt rather than overwriting the one before it.
//
// The change lookup and the attempt allocation both happen inside one
// INSERT ... SELECT statement -- there is no separate read of the current
// maximum attempt followed by a gated insert. Two concurrent callers for
// the same triple can still both compute the same next attempt before
// either commits; when that happens, the loser's insert collides with
// stage_runs' UNIQUE (change_id, command, stage, attempt) constraint, and
// BeginStage retries -- recomputing the maximum from what is now committed
// -- rather than the caller ever observing a torn or duplicated attempt
// number. This is the same shape changes.go already uses for the project
// bootstrap: let the database detect the race and report it as a typed,
// retryable condition, instead of a client-side check-then-act.
func (s *Store) BeginStage(ctx context.Context, in BeginStageInput) (StageRun, error) {
	for range maxAttemptRetries {
		run, err := s.insertStageRun(ctx, in)
		if err == nil {
			return run, nil
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return StageRun{}, fmt.Errorf("%w: %s/%s", ErrChangeNotFound, in.ProjectKey, in.ChangeName)
		}
		if isUniqueViolation(err, stageRunsAttemptConstraint) {
			continue
		}
		return StageRun{}, fmt.Errorf("store: begin stage %s/%s for %s/%s: %w", in.Command, in.Stage, in.ProjectKey, in.ChangeName, err)
	}
	return StageRun{}, fmt.Errorf("%w: %s/%s %s/%s", ErrTooManyAttemptCollisions, in.ProjectKey, in.ChangeName, in.Command, in.Stage)
}

// insertStageRun inserts one stage run. Its session_id is computed inside
// the same INSERT ... SELECT as the attempt number: the caller-supplied
// SessionID ($5) wins when given, otherwise this looks for a session
// already bound to the same session_token by an earlier run (KAN-172, task
// 4b's "a mark arriving with an already-bound token is bound at write
// time, doing no resolution work at all") -- one run's own token is
// shared, unchanged, by every mark that run makes (design.md's "one token
// per session, not one per mark"), so a second or later mark for a run
// whose token the harvester has already resolved gets its session_id at
// insert time and never enters the unresolved-session-token pool at all.
// A session_token of NULL naturally matches no row in the subquery, so no
// separate NULL guard is needed.
func (s *Store) insertStageRun(ctx context.Context, in BeginStageInput) (StageRun, error) {
	var (
		run          StageRun
		metrics      []byte
		sessionID    *string
		sessionToken *string
	)
	err := s.pool.QueryRow(ctx, `
		INSERT INTO stage_runs (
			change_id, repo_root, harness, session_id, session_token, command, stage, attempt, started_at, metrics
		)
		SELECT
			c.id, $3, $4,
			COALESCE($5, (
				SELECT sr2.session_id FROM stage_runs sr2
				WHERE sr2.session_token = $6 AND sr2.session_id IS NOT NULL
				LIMIT 1
			)),
			$6, $7, $8,
			COALESCE(
				(SELECT MAX(sr.attempt) FROM stage_runs sr
				 WHERE sr.change_id = c.id AND sr.command = $7 AND sr.stage = $8),
				0
			) + 1,
			$9, '{}'::jsonb
		FROM changes c
		WHERE c.project_key = $1 AND c.name = $2
		RETURNING id, change_id, repo_root, harness, session_id, session_token, command, stage, attempt,
		          started_at, ended_at, outcome, metrics
	`, in.ProjectKey, in.ChangeName, in.RepoRoot, in.Harness, in.SessionID, in.SessionToken, in.Command, in.Stage, in.StartedAt).
		Scan(
			&run.ID, &run.ChangeID, &run.RepoRoot, &run.Harness, &sessionID, &sessionToken, &run.Command, &run.Stage,
			&run.Attempt, &run.StartedAt, &run.EndedAt, &run.Outcome, &metrics,
		)
	if err != nil {
		return StageRun{}, err
	}
	run.SessionID = sessionID
	run.SessionToken = sessionToken
	run.Metrics = metrics
	return run, nil
}

// isUniqueViolation reports whether err is a Postgres unique-violation on
// the named constraint. An empty constraint matches any unique violation.
func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return false
	}
	return constraint == "" || pgErr.ConstraintName == constraint
}

// EndStage records the end of a stage run: its end instant and outcome.
// Metrics are not touched here -- MergeMetrics is the only path that writes
// them, whether called before or after EndStage, so the harvester's token
// keys and this call's outcome keys never race to overwrite each other.
func (s *Store) EndStage(ctx context.Context, stageRunID int64, endedAt time.Time, outcome string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE stage_runs SET ended_at = $2, outcome = $3 WHERE id = $1
	`, stageRunID, endedAt, outcome)
	if err != nil {
		return fmt.Errorf("store: end stage %d: %w", stageRunID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %d", ErrStageRunNotFound, stageRunID)
	}
	return nil
}

// MergeMetrics merges patch into a stage run's metrics bag and never
// replaces it. The merge is recursive (jsonb_deep_merge, defined in
// 0003_stage_runs.sql), not a top-level JSONB concatenation: the bag is
// nested -- every token figure lives under one "tokens" object -- so a
// shallow `metrics || patch` would replace "tokens" wholesale on any write
// that touches it, discarding sibling keys an earlier, unrelated write
// already recorded there. With the recursive merge, a key already stored
// but absent from patch is left untouched at every level, not just the
// top one, and a key present in both -- at any depth -- takes patch's
// value; only a non-object value ever replaces what was there. This is how
// the harvester's token keys and a stage-end mark's outcome keys, or two
// separate token sub-keys written by different calls, coexist in the same
// bag without either call erasing what the other wrote. The merge runs as
// one atomic UPDATE, so no caller has to read the bag first to avoid
// clobbering a sibling it doesn't know about -- doing so in Go would
// reintroduce the same lost-update race BeginStage's attempt allocation
// was written to avoid.
//
// An explicit JSON null in patch overwrites the key with null rather than
// deleting it -- null is a non-object value, so it follows the same
// replace rule as any other scalar. This is deliberately not RFC 7396
// JSON Merge Patch semantics, where null means "delete this key": there is
// no delete operation here at all, only merge. Returns ErrNilMetricsPatch
// if patch itself is nil (Go nil, not the JSON literal null).
func (s *Store) MergeMetrics(ctx context.Context, stageRunID int64, patch json.RawMessage) error {
	if patch == nil {
		return ErrNilMetricsPatch
	}

	tag, err := s.pool.Exec(ctx, `
		UPDATE stage_runs SET metrics = jsonb_deep_merge(metrics, $2::jsonb) WHERE id = $1
	`, stageRunID, patch)
	if err != nil {
		return fmt.Errorf("store: merge metrics for stage run %d: %w", stageRunID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %d", ErrStageRunNotFound, stageRunID)
	}
	return nil
}

// GetStageRun returns the stage run recorded under id.
func (s *Store) GetStageRun(ctx context.Context, id int64) (StageRun, error) {
	var (
		run          StageRun
		metrics      []byte
		sessionID    *string
		sessionToken *string
	)
	err := s.pool.QueryRow(ctx, `
		SELECT id, change_id, repo_root, harness, session_id, session_token, command, stage, attempt,
		       started_at, ended_at, outcome, metrics
		FROM stage_runs
		WHERE id = $1
	`, id).Scan(
		&run.ID, &run.ChangeID, &run.RepoRoot, &run.Harness, &sessionID, &sessionToken, &run.Command, &run.Stage,
		&run.Attempt, &run.StartedAt, &run.EndedAt, &run.Outcome, &metrics,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StageRun{}, fmt.Errorf("%w: %d", ErrStageRunNotFound, id)
		}
		return StageRun{}, fmt.Errorf("store: get stage run %d: %w", id, err)
	}
	run.SessionID = sessionID
	run.SessionToken = sessionToken
	run.Metrics = metrics
	return run, nil
}

// QueryStageRuns returns the page of stage runs matching q -- its
// filters, free-text search, sort and page -- against the fixed allowlist
// in query.go, along with the total count of matching rows (ignoring q's
// page). Every stage run is joined to its owning change, so a filter,
// sort or search term naming one of the change's own fields (project,
// name, state, branch, jira_issue, planning_effort, review_panel_roster,
// pr_url, updated_at, updated_by) resolves against that join, exactly as
// it would against QueryChanges. A field query.go's allowlist does not
// recognise rejects the request with ErrUnknownField before any query
// runs.
//
// The count and the page are read inside one transaction, opened with
// queryTxOptions() -- see its doc comment, and QueryChanges', for why
// REPEATABLE READ is load-bearing here and why the 40001 serialization
// failure that isolation level can raise cannot occur on this read-only
// transaction.
func (s *Store) QueryStageRuns(ctx context.Context, q Query) ([]StageRun, int, error) {
	where, whereArgs, order, orderArgs, err := q.buildClauses(stageRunQueryable)
	if err != nil {
		return nil, 0, err
	}

	tx, err := s.pool.BeginTx(ctx, queryTxOptions())
	if err != nil {
		return nil, 0, fmt.Errorf("store: query stage runs: begin: %w", err)
	}
	// Rollback after a successful Commit is a documented no-op in pgx; its
	// error carries nothing actionable here, so it is discarded explicitly
	// rather than checked.
	defer func() { _ = tx.Rollback(ctx) }()

	var total int
	countSQL := "SELECT COUNT(*) FROM stage_runs sr JOIN changes c ON c.id = sr.change_id " + where
	if err := tx.QueryRow(ctx, countSQL, whereArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("store: query stage runs: count: %w", err)
	}

	args := append(append([]any{}, whereArgs...), orderArgs...)
	limitClause := ""
	if !q.unlimited() {
		limit, offset := q.limit(), q.offset()
		args = append(args, limit, offset)
		limitClause = fmt.Sprintf("LIMIT $%d OFFSET $%d", len(whereArgs)+len(orderArgs)+1, len(whereArgs)+len(orderArgs)+2)
	}

	sqlText := fmt.Sprintf(`
		SELECT sr.id, sr.change_id, sr.repo_root, sr.harness, sr.session_id, sr.session_token, sr.command, sr.stage,
		       sr.attempt, sr.started_at, sr.ended_at, sr.outcome, sr.metrics
		FROM stage_runs sr
		JOIN changes c ON c.id = sr.change_id
		%s
		%s
		%s
	`, where, order, limitClause)

	rows, err := tx.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("store: query stage runs: %w", err)
	}
	defer rows.Close()

	var out []StageRun
	for rows.Next() {
		var (
			run          StageRun
			metrics      []byte
			sessionID    *string
			sessionToken *string
		)
		if err := rows.Scan(
			&run.ID, &run.ChangeID, &run.RepoRoot, &run.Harness, &sessionID, &sessionToken, &run.Command, &run.Stage,
			&run.Attempt, &run.StartedAt, &run.EndedAt, &run.Outcome, &metrics,
		); err != nil {
			return nil, 0, fmt.Errorf("store: query stage runs: scan: %w", err)
		}
		run.SessionID = sessionID
		run.SessionToken = sessionToken
		run.Metrics = metrics
		out = append(out, run)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("store: query stage runs: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, 0, fmt.Errorf("store: query stage runs: commit: %w", err)
	}

	return out, total, nil
}

// UnresolvedSessionTokens returns every stage run id and its session_token
// for which session_id has not yet been bound (KAN-172, task 2; reworked
// per-run rather than per-mark in task 4b). It is the harvester's sole
// read path onto the correlator task 1 introduced -- exactly the rows
// stage_runs_unresolved_session_token (migration 0008) indexes, so this
// query costs nothing proportional to the table's full size.
//
// A run whose session_token is NULL (predates the column, or was marked
// by a harness build that never sent one) is not returned: there is
// nothing here for a harvest cycle to look for. Nor is a run whose token
// insertStageRun already resolved at insert time (KAN-172, task 4b) --
// its session_id was never NULL to begin with, so it never enters this
// pool at all, regardless of how many other runs share its token.
func (s *Store) UnresolvedSessionTokens(ctx context.Context) (map[int64]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, session_token FROM stage_runs
		WHERE session_token IS NOT NULL AND session_id IS NULL
	`)
	if err != nil {
		return nil, fmt.Errorf("store: unresolved session tokens: %w", err)
	}
	defer rows.Close()

	out := make(map[int64]string)
	for rows.Next() {
		var (
			id           int64
			sessionToken string
		)
		if err := rows.Scan(&id, &sessionToken); err != nil {
			return nil, fmt.Errorf("store: unresolved session tokens: scan: %w", err)
		}
		out[id] = sessionToken
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: unresolved session tokens: %w", err)
	}
	return out, nil
}

// BindSession sets session_id to sessionID on every stage run carrying
// sessionToken that has not already been bound -- not just the one stage
// run whose mark first revealed the token to the harvester (KAN-172, task
// 4b's "a resolving token binds every run carrying it, not just the one
// that revealed it" -- the whole economy of moving from one correlator per
// mark to one per session). The UPDATE's own WHERE session_id IS NULL is
// what makes binding one-way per row: a run that already carries a
// session_id -- bound by an earlier cycle, resolved at insert time
// (insertStageRun's own doc comment), or (in principle) by a concurrent
// harvester -- is left untouched rather than overwritten. This is the same
// "let the database detect and report the race, don't check-then-act in
// Go" shape BeginStage's attempt allocation already uses.
//
// bound reports how many rows this call actually updated. Zero, with no
// error, is not a failure: every run carrying the token may already have
// been bound (by an earlier cycle, or because every later mark for that
// run resolved its session_id at insert time and never needed this path
// at all) -- there is nothing left to retry.
func (s *Store) BindSession(ctx context.Context, sessionToken string, sessionID string) (bound int64, err error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE stage_runs SET session_id = $2 WHERE session_token = $1 AND session_id IS NULL
	`, sessionToken, sessionID)
	if err != nil {
		return 0, fmt.Errorf("store: bind session for session token: %w", err)
	}
	return tag.RowsAffected(), nil
}

// SweepAbandoned closes every stage run that is still open (no end mark)
// and whose session has gone silent: started before silentBefore, with no
// activity to indicate otherwise. It sets ended_at to now and outcome to
// "abandoned", and returns how many rows it closed.
//
// An abandoned stage is a statistic worth having, not an error to
// suppress: the rework-rate view reads the "abandoned" outcome directly,
// exactly as it reads any other recorded outcome.
func (s *Store) SweepAbandoned(ctx context.Context, silentBefore time.Time) (int64, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE stage_runs
		SET ended_at = now(), outcome = 'abandoned'
		WHERE ended_at IS NULL AND started_at < $1
	`, silentBefore)
	if err != nil {
		return 0, fmt.Errorf("store: sweep abandoned stages: %w", err)
	}
	return tag.RowsAffected(), nil
}
