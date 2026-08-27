package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/tweety53/agents/stats/internal/harvest"
	"github.com/tweety53/agents/stats/internal/records"
)

// ErrFindingNotFound is returned by SetFindingStatus when the change holds
// no finding under the given ref. It is typed so internal/api can answer
// 404 rather than letting a caller's typo surface as a 500 -- the
// distinction between "you asked for something that isn't there" and "this
// server is broken".
var ErrFindingNotFound = errors.New("store: finding not found")

// ErrDispatchNotFound is returned by MergeDispatchMetrics when no dispatch
// exists under the given id, and by EndDispatch when the change holds no
// dispatch under the given session token and key.
var ErrDispatchNotFound = errors.New("store: dispatch not found")

// ErrTooManyDispatchSeqCollisions is returned by RecordDispatch when it
// could not allocate a seq after retrying past every concurrent collision
// it was willing to absorb -- contention so extreme that the caller, not
// the store, should decide what to do next. It is never returned for an
// ordinary single collision, which RecordDispatch resolves itself by
// retrying. It is a separate sentinel from ErrTooManyAttemptCollisions
// rather than a reuse of it: the two name different allocations on
// different tables, and a caller shown "allocating a stage run" for a
// dispatch write would be sent looking in the wrong place.
var ErrTooManyDispatchSeqCollisions = errors.New("store: too many concurrent collisions allocating a dispatch seq")

// dispatchesSeqConstraint is the name given, explicitly, to the
// UNIQUE (change_id, seq) constraint in 0010_run_records.sql.
// RecordDispatch checks for this exact constraint name rather than for any
// unique violation, so it never mistakes an unrelated conflict for the seq
// race it knows how to retry -- the same reasoning
// stageRunsAttemptConstraint carries.
const dispatchesSeqConstraint = "dispatches_seq_key"

// dispatchesKeyConstraint is the name given, explicitly, to the
// UNIQUE (change_id, session_token, dispatch_key) constraint in
// 0012_dispatch_key.sql. RecordDispatch names it in its ON CONFLICT clause
// so that recording a dispatch is idempotent under replay -- see that
// method's own doc comment. Naming the constraint rather than listing the
// columns is the same call findingsRefConstraint records.
const dispatchesKeyConstraint = "dispatches_key_key"

// findingsRefConstraint is the name given, explicitly, to the
// UNIQUE (change_id, ref) constraint in 0010_run_records.sql. UpsertFinding
// names it in its ON CONFLICT clause, so a fix round updates the existing
// row rather than being refused, and the record of a change's findings
// never accumulates a second row for one ref.
const findingsRefConstraint = "findings_ref_key"

// maxDispatchSeqRetries bounds how many times RecordDispatch retries after
// losing a seq race before giving up with ErrTooManyDispatchSeqCollisions.
// It is set well above any concurrency the pipeline actually produces (a
// review panel dispatches three slots at once; the store's own concurrency
// test fires sixteen), on the same reasoning maxAttemptRetries records: N
// concurrent writers need at most N-1 retries in the worst case, and this
// bound leaves real margin above that rather than sitting exactly on it.
const maxDispatchSeqRetries = 100

// RecordDispatch records one subagent dispatch for a change and allocates
// its seq -- the dispatch's append order within that change, which is what
// the rendered ledger reads in. The caller's own in.Seq is ignored; the
// allocated value comes back on the returned row.
//
// The row's own id comes back on the returned record. Seq is what a
// rendered record and a finding's DispatchSeq name a dispatch by, but
// MergeDispatchMetrics is keyed by the id, so a caller that has just
// recorded a dispatch can merge the harvester's figures into it without a
// query of its own -- and without internal/store having to grow a
// "look up the id of the row I just wrote" method that exists only because
// the write path declined to say.
//
// The change lookup and the seq allocation happen inside one
// INSERT ... SELECT, not as a read of the current maximum followed by a
// gated insert. Two concurrent callers for the same change can still both
// compute the same next seq before either commits; when that happens the
// loser's insert collides with dispatches_seq_key and RecordDispatch
// retries, recomputing the maximum from what is now committed, rather than
// either caller observing a torn or duplicated position. This is the same
// shape BeginStage's attempt allocation already uses: let the database
// detect the race and report it as a typed, retryable condition, instead of
// a client-side check-then-act.
//
// THE WRITE IS IDEMPOTENT under in.Key. A record write that could not reach
// the store is journalled and replayed later, and a *lost response* is
// indistinguishable to the caller from a store that was never reached: the
// row may already exist. Without a key to collide on, the replay would
// allocate a fresh seq and insert a second row for one logical dispatch,
// and the harvester would then attribute that dispatch's tokens across two
// rows -- one logical dispatch counted twice in every cost figure derived
// from this table. UpsertFinding and SetFindingStatus are both idempotent
// already; this is the third write, and it is the one whose duplicate costs
// money rather than merely tidiness.
//
// The key is (change, session token, key), and the reason that is the right
// key is that a replay reproduces it EXACTLY. The journal stores the
// request as it was built, so the literal the caller wrote for -key and the
// literal it wrote for -session-token both come back byte for byte; nothing
// in the key is derived from the clock, from the store's own allocation, or
// from anything else the second attempt would compute afresh. A `begin`
// replayed twice therefore reaches ON CONFLICT and returns the row already
// there, with its original seq.
//
// A dispatch recorded with an empty Key is inserted and never deduplicated:
// the constraint spans nullable columns, and SQL treats two NULLs as
// distinct. `flow record dispatch begin` requires the flag, so that path
// is not reachable from this repository's own callers.
func (s *Store) RecordDispatch(ctx context.Context, projectKey, change string, in records.Dispatch) (records.Dispatch, error) {
	for range maxDispatchSeqRetries {
		out, err := s.insertDispatch(ctx, projectKey, change, in)
		if err == nil {
			return out, nil
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return records.Dispatch{}, fmt.Errorf("%w: %s/%s", ErrChangeNotFound, projectKey, change)
		}
		if isUniqueViolation(err, dispatchesSeqConstraint) {
			continue
		}
		return records.Dispatch{}, fmt.Errorf("store: record dispatch for %s/%s: %w", projectKey, change, err)
	}
	return records.Dispatch{}, fmt.Errorf("%w: %s/%s", ErrTooManyDispatchSeqCollisions, projectKey, change)
}

// dispatchColumns is the column list every read of a dispatch row selects,
// in the order scanDispatchRow scans them. It is one constant rather than
// two identical lists because the insert and the read used to carry a copy
// each: fifteen columns and fifteen scan targets, twice, where adding a
// sixteenth to one and not the other compiles cleanly and silently reads
// the wrong column into the wrong field.
const dispatchColumns = `id, seq, stage_run_id, task_id, role, slot, model, commit_sha, outcome,
	          session_token, started_at, ended_at, metrics, notes, agent_id, dispatch_key,
	          diff_base`

// qualifiedDispatchColumns is dispatchColumns with every name qualified by
// alias, for the one statement that needs it -- an UPDATE ... FROM, where
// an unqualified `id` is ambiguous across the two tables in scope.
func qualifiedDispatchColumns(alias string) string {
	cols := strings.Split(dispatchColumns, ",")
	for i, c := range cols {
		cols[i] = alias + "." + strings.TrimSpace(c)
	}
	return strings.Join(cols, ", ")
}

// dispatchRowScanner is the one method scanDispatchRow needs, satisfied by
// both pgx.Row (a QueryRow result) and pgx.Rows (one row of a Query
// result), so the insert path and the read path share one decode.
type dispatchRowScanner interface {
	Scan(dest ...any) error
}

// scanDispatchRow decodes one row of dispatchColumns into a records.Dispatch,
// mapping every nullable column back through derefOrEmpty.
//
// It is shared by insertDispatch, EndDispatch and readDispatches
// deliberately. The insert and the read carried the same fifteen-column
// scan and the same eight derefOrEmpty assignments, and the failure that
// duplication invites is not a compile error: adding a column to one path
// leaves the other reading a row whose shape it no longer matches, or --
// worse -- reading the right shape into the wrong fields. One decode, named
// by one column list, cannot drift from itself.
func scanDispatchRow(row dispatchRowScanner) (records.Dispatch, error) {
	var (
		d            records.Dispatch
		agentID      *string
		taskID       *string
		slot         *string
		commitSHA    *string
		outcome      *string
		sessionToken *string
		notes        *string
		dispatchKey  *string
		diffBase     *string
		bag          []byte
	)
	if err := row.Scan(
		&d.ID, &d.Seq, &d.StageRunID, &taskID, &d.Role, &slot, &d.Model, &commitSHA, &outcome,
		&sessionToken, &d.StartedAt, &d.EndedAt, &bag, &notes, &agentID, &dispatchKey,
		&diffBase,
	); err != nil {
		return records.Dispatch{}, err
	}
	d.AgentID = derefOrEmpty(agentID)
	d.TaskID = derefOrEmpty(taskID)
	d.Slot = derefOrEmpty(slot)
	d.CommitSHA = derefOrEmpty(commitSHA)
	d.Outcome = derefOrEmpty(outcome)
	d.SessionToken = derefOrEmpty(sessionToken)
	d.Notes = derefOrEmpty(notes)
	d.Key = derefOrEmpty(dispatchKey)
	d.DiffBase = derefOrEmpty(diffBase)
	d.Metrics = bag
	return d, nil
}

// insertDispatch is one attempt at RecordDispatch's insert, separated so
// the retry loop above reads as the policy it is. A lost race for a seq
// surfaces here as a unique violation on dispatches_seq_key, which the loop
// keys on.
//
// A collision on dispatches_key_key is a different thing entirely and is
// NOT a race: it means this exact dispatch has already been recorded, so
// the statement resolves it here rather than returning it to the loop. The
// DO UPDATE writes the conflict key back onto itself -- a genuine no-op --
// purely so that RETURNING yields the existing row; DO NOTHING would return
// no row at all and leave the caller unable to tell a replayed write from a
// change that does not exist.
func (s *Store) insertDispatch(ctx context.Context, projectKey, change string, in records.Dispatch) (records.Dispatch, error) {
	metrics := in.Metrics
	if len(metrics) == 0 {
		metrics = json.RawMessage(`{}`)
	}

	return scanDispatchRow(s.pool.QueryRow(ctx, `
		INSERT INTO dispatches (
			change_id, stage_run_id, seq, task_id, role, slot, model, commit_sha, outcome,
			session_token, started_at, ended_at, metrics, notes, agent_id, dispatch_key,
			diff_base
		)
		SELECT
			c.id, $3,
			COALESCE((SELECT MAX(d.seq) FROM dispatches d WHERE d.change_id = c.id), 0) + 1,
			$4, $5, $6, $7, $8, $9, $10, $11, $12, $13::jsonb, $14, $15, $16, $17
		FROM changes c
		WHERE c.project_key = $1 AND c.name = $2
		ON CONFLICT ON CONSTRAINT `+dispatchesKeyConstraint+` DO UPDATE SET
			dispatch_key = EXCLUDED.dispatch_key
		RETURNING `+dispatchColumns+`
	`,
		projectKey, change, in.StageRunID, nullIfEmpty(in.TaskID), in.Role, nullIfEmpty(in.Slot),
		in.Model, nullIfEmpty(in.CommitSHA), nullIfEmpty(in.Outcome), nullIfEmpty(in.SessionToken),
		in.StartedAt, in.EndedAt, metrics, nullIfEmpty(in.Notes), nullIfEmpty(in.AgentID),
		nullIfEmpty(in.Key), nullIfEmpty(in.DiffBase),
	))
}

// EndDispatch closes the dispatch named by in.SessionToken and in.Key,
// writing the three facts that are knowable only at its close: the commit
// it produced, how it ended, and when.
//
// IT IS THE HALF THAT CLOSES THE ATTRIBUTION WINDOW, and that is why it
// exists as a call of its own rather than as columns on the insert. A
// dispatch row whose ended_at is NULL is an OPEN window
// (harvest.DispatchWindow.contains), and an open window goes on containing
// every later timestamp forever: a dispatch that is never closed keeps
// claiming usage that belongs to the dispatches that followed it. Before
// this call existed no production path set ended_at at all, so every window
// this table produced was open.
//
// It is idempotent by construction: replaying it writes the same three
// values onto the same row. A key naming no row is ErrDispatchNotFound
// rather than a silent no-op -- see the CLI's own classification for why
// that is a *retryable* answer for this one write and not a definitive one,
// the begin it closes possibly still sitting in the journal ahead of it.
//
// agent_id is written only when in.AgentID is non-empty, via
// COALESCE($8, d.agent_id) rather than a plain assignment: `begin` may
// already have recorded a real identifier for this row, and on Claude Code
// the harness reports one only once the dispatch has launched, so `end`
// has to be able to carry it too. A plain `SET agent_id = $8` would clear
// that identifier on every ordinary end call that omits it -- destroying
// the very thing this parameter exists to capture -- since nullIfEmpty
// turns "" into SQL NULL and COALESCE is what lets a NULL argument here
// mean "leave the column as it is" instead of "set it to NULL". Where
// in.AgentID names a different identifier than `begin` already recorded,
// end's value wins: this call has no notion of a conflict, only of
// "supplied" versus "omitted".
func (s *Store) EndDispatch(ctx context.Context, projectKey, change string, in records.DispatchEnd) (records.Dispatch, error) {
	out, err := scanDispatchRow(s.pool.QueryRow(ctx, `
		UPDATE dispatches d
		SET commit_sha = $5, outcome = $6, ended_at = $7, agent_id = COALESCE($8, d.agent_id)
		FROM changes c
		WHERE c.id = d.change_id AND c.project_key = $1 AND c.name = $2
		  AND d.session_token = $3 AND d.dispatch_key = $4
		RETURNING `+qualifiedDispatchColumns("d"),
		projectKey, change, in.SessionToken, in.Key,
		nullIfEmpty(in.CommitSHA), nullIfEmpty(in.Outcome), in.EndedAt, nullIfEmpty(in.AgentID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return records.Dispatch{}, fmt.Errorf("%w: %s/%s key %q", ErrDispatchNotFound, projectKey, change, in.Key)
		}
		return records.Dispatch{}, fmt.Errorf("store: end dispatch %q for %s/%s: %w", in.Key, projectKey, change, err)
	}
	return out, nil
}

// UpsertFinding records one review-panel finding, or updates the one
// already recorded under the same ref for the same change, and reports
// which of the two it did. A ref is unique per change, not per round: a
// fix round restating F1 rewrites F1's row, so a change's findings never
// accumulate a second row for one reference.
//
// The created result exists because nothing outside this statement can
// know it. internal/api answers 201 for an insert and 200 for a replace,
// so a caller can tell the round that first raised a finding from a round
// that restated it, and the only place that distinction is observable is
// inside the upsert itself.
//
// It is read from the returned row's xmax. A row this statement genuinely
// inserted carries xmax = 0; a row ON CONFLICT DO UPDATE reached instead
// carries the updating transaction's id, so (xmax = 0) is exactly "the
// insert won". The test is per row and is evaluated as the row is
// written, which is what makes it correct under concurrency.
//
// A lookup of the ref before the insert -- a CTE beside it, or a separate
// SELECT -- is not. Both read the statement's own start snapshot, and
// conflict resolution never re-evaluates it, so every writer racing to be
// the first to record a new ref sees no row and every one of them reports
// created. TestConcurrentUpsertFindingReportsCreatedExactlyOnce is that
// race, and it is not theoretical: twenty concurrent writers reported
// created two to three times each run against a real Postgres, which
// would have the API answer 201 Created repeatedly for one ref -- the
// ambiguity this flag exists to remove.
//
// The conflict is resolved by naming findings_ref_key rather than a column
// list, so the clause fails loudly if that constraint is ever renamed
// instead of quietly matching some other unique index that happens to cover
// the same columns.
//
// dispatch_id is resolved from in.DispatchSeq in the same statement, by a
// scalar subquery over the same change's dispatches. Seq is the only
// identifier a caller has -- the row id is this package's own bookkeeping,
// and internal/api never sees one -- so the translation belongs here
// rather than in a caller that would have to query for it first. A nil
// DispatchSeq, and a seq the change holds no dispatch under, both leave
// the column NULL: a finding no single dispatch raised is a legitimate
// case, and this column records the raising slot where it is known rather
// than refusing the finding where it is not.
func (s *Store) UpsertFinding(ctx context.Context, projectKey, change string, in records.Finding) (records.Finding, bool, error) {
	var (
		out         records.Finding
		dispatchSeq *int
		location    *string
		reproducer  *string
		created     bool
	)

	err := s.pool.QueryRow(ctx, `
		INSERT INTO findings (
			change_id, dispatch_id, ref, round, slot, severity, location, note, status, reproducer
		)
		SELECT
			c.id,
			(SELECT d.id FROM dispatches d WHERE d.change_id = c.id AND d.seq = $4::int),
			$3, $5, $6, $7, $8, $9, $10, $11
		FROM changes c
		WHERE c.project_key = $1 AND c.name = $2
		ON CONFLICT ON CONSTRAINT `+findingsRefConstraint+` DO UPDATE SET
			dispatch_id = EXCLUDED.dispatch_id,
			round       = EXCLUDED.round,
			slot        = EXCLUDED.slot,
			severity    = EXCLUDED.severity,
			location    = EXCLUDED.location,
			note        = EXCLUDED.note,
			status      = EXCLUDED.status,
			reproducer  = EXCLUDED.reproducer
		RETURNING
			ref,
			(SELECT d.seq FROM dispatches d WHERE d.id = findings.dispatch_id),
			round, slot, severity, location, note, status, reproducer,
			xmax = 0
	`,
		projectKey, change, in.Ref, in.DispatchSeq, in.Round, in.Slot, in.Severity,
		nullIfEmpty(in.Location), in.Note, in.Status, nullIfEmpty(in.Reproducer),
	).Scan(&out.Ref, &dispatchSeq, &out.Round, &out.Slot, &out.Severity, &location,
		&out.Note, &out.Status, &reproducer, &created)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return records.Finding{}, false, fmt.Errorf("%w: %s/%s", ErrChangeNotFound, projectKey, change)
		}
		return records.Finding{}, false, fmt.Errorf("store: upsert finding %s for %s/%s: %w", in.Ref, projectKey, change, err)
	}

	out.DispatchSeq = dispatchSeq
	out.Location = derefOrEmpty(location)
	out.Reproducer = derefOrEmpty(reproducer)
	return out, created, nil
}

// SetFindingStatus updates one finding's status and nothing else -- the
// whole of what a fix round changes about a finding it has resolved. A ref
// the change holds no finding under is ErrFindingNotFound, not a silent
// no-op, so a caller's typo is reported rather than looking like a
// successful update.
func (s *Store) SetFindingStatus(ctx context.Context, projectKey, change, ref, status string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE findings f
		SET status = $4
		FROM changes c
		WHERE c.id = f.change_id AND c.project_key = $1 AND c.name = $2 AND f.ref = $3
	`, projectKey, change, ref, status)
	if err != nil {
		return fmt.Errorf("store: set finding %s status for %s/%s: %w", ref, projectKey, change, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s in %s/%s", ErrFindingNotFound, ref, projectKey, change)
	}
	return nil
}

// MergeDispatchMetrics merges patch into a dispatch's metrics bag and never
// replaces it, recursively and as one atomic UPDATE rather than a
// caller-side read-modify-write -- see MergeMetrics' doc comment for why a
// stage run's bag needs both, which this column needs for the same reasons.
//
// The merge is jsonb_deep_add (0005_jsonb_deep_add.sql), not
// jsonb_deep_merge: structurally the same recursive merge, with the one
// added case that two numbers at the same key sum instead of the second
// replacing the first. That case is the whole difference between a
// dispatch's figures being right and being the last batch's only. This
// column has exactly one writer -- the harvester's second attribution pass
// (internal/harvest, DispatchAttributor) -- and it sends a *batch delta*,
// never a cumulative total, for the same reason CommitHarvestBatch does:
// the harvester reads a transcript incrementally and holds no running
// total anywhere, Postgres being the only place one exists. A dispatch
// long enough to span two harvest cycles -- which is most of them, the
// cycle being far shorter than a subagent's run -- would otherwise end up
// recording only whichever batch landed last. Every non-numeric leaf still
// resolves last-write-wins exactly as it did under jsonb_deep_merge, so a
// descriptor or an outcome key merged into this bag behaves unchanged.
//
// Returns ErrNilMetricsPatch if patch itself is nil (Go nil, not the JSON
// literal null): jsonb_deep_add(metrics, NULL) returns NULL, which this
// column's NOT NULL constraint would reject as a raw Postgres error rather
// than as the caller mistake it is.
func (s *Store) MergeDispatchMetrics(ctx context.Context, dispatchID int64, patch json.RawMessage) error {
	if patch == nil {
		return ErrNilMetricsPatch
	}

	tag, err := s.pool.Exec(ctx, `
		UPDATE dispatches SET metrics = jsonb_deep_add(metrics, $2::jsonb) WHERE id = $1
	`, dispatchID, patch)
	if err != nil {
		return fmt.Errorf("store: merge metrics for dispatch %d: %w", dispatchID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %d", ErrDispatchNotFound, dispatchID)
	}
	return nil
}

// MarkDispatchesUnattributed stamps every dispatch recorded under token
// with the reason its session's cost could not be attributed --
// RecordSessionTokenGiveUp's dispatch-grain counterpart
// (0013_session_token_giveups.sql). Zero dispatches under token is not an
// error: a give-up can be recorded before any dispatch under that token
// was ever written.
//
// The stamp merges via the jsonb "||" operator, not jsonb_deep_add: || is a
// top-level, non-recursive merge, so it replaces the whole "unattributed"
// key wholesale while leaving "tokens" and every other sibling key
// untouched -- an existing tokens figure survives, exactly as it did under
// jsonb_deep_add, so a dispatch that was measured and then had its session
// given up is still not resolved by destroying the measurement.
// unattributed.candidates is a snapshot fact, not a delta: this method
// fires again on a retried give-up still ambiguous after a restart, and a
// repeat stamp with the same candidates count must leave it unchanged
// rather than summing it, which is exactly what jsonb_deep_add did instead
// (F18, review panel round 1) -- "||" is idempotent under an identical
// repeat stamp, and a later stamp with a different reason replaces the
// stale reason and candidates together rather than merging them.
//
// candidates is included in the payload only when positive -- it names how
// many sessions an ambiguous match was torn between, and is meaningless
// for a give-up that never found a candidate at all. A non-positive value
// omits the key entirely rather than writing a 0 that would read as a
// measurement.
func (s *Store) MarkDispatchesUnattributed(ctx context.Context, token, reason string, candidates int) error {
	unattributed := map[string]any{"reason": reason}
	if candidates > 0 {
		unattributed["candidates"] = candidates
	}
	patch, err := json.Marshal(map[string]any{"unattributed": unattributed})
	if err != nil {
		return fmt.Errorf("store: marshal unattributed patch for session token %s: %w", token, err)
	}

	if _, err := s.pool.Exec(ctx, `
		UPDATE dispatches SET metrics = metrics || $2::jsonb WHERE session_token = $1
	`, token, patch); err != nil {
		return fmt.Errorf("store: mark dispatches unattributed for session token %s: %w", token, err)
	}
	return nil
}

// MarkDispatchesUnattributedByID stamps exactly the dispatches named by
// ids with the reason their cost could not be attributed -- the
// dispatch-grain second pass's own ambiguity counterpart of
// MarkDispatchesUnattributed above, added beside it rather than in place
// of it (task 6's own corrections, tasks.md): the two reasons have
// genuinely different scopes. MarkDispatchesUnattributed stamps every
// dispatch under a session token, which is correct only for "session
// never bound" -- every dispatch of that session really is uncosted. A
// dispatch-grain ambiguity (bestDispatchWindow, internal/harvest,
// attribute.go) names specific rows -- the candidates a record's agent id
// or timestamp could not tell apart -- and stamping every dispatch under
// their shared session would also stamp siblings that attributed
// correctly. Zero ids is a no-op, not an error: nothing to stamp is not a
// failure.
//
// The stamp merges via the jsonb "||" operator, for the same reason
// MarkDispatchesUnattributed's own does: a dispatch that was measured and
// then found ambiguous is a contradiction this method must not resolve by
// destroying the measurement, so an existing "tokens" key survives the
// merge untouched alongside the new "unattributed" key -- and a repeat
// stamp for the same candidate set, which a multi-minute review panel
// round produces on every 5s harvest cycle that still sees it, leaves
// candidates unchanged instead of summing it (F18, review panel round 1).
//
// candidates is included in the payload only when positive, for the same
// reason MarkDispatchesUnattributed's own is: a non-positive value omits
// the key entirely rather than writing a 0 that would read as a
// measurement.
func (s *Store) MarkDispatchesUnattributedByID(ctx context.Context, ids []int64, reason string, candidates int) error {
	if len(ids) == 0 {
		return nil
	}

	unattributed := map[string]any{"reason": reason}
	if candidates > 0 {
		unattributed["candidates"] = candidates
	}
	patch, err := json.Marshal(map[string]any{"unattributed": unattributed})
	if err != nil {
		return fmt.Errorf("store: marshal unattributed patch for dispatch ids %v: %w", ids, err)
	}

	if _, err := s.pool.Exec(ctx, `
		UPDATE dispatches SET metrics = metrics || $2::jsonb WHERE id = ANY($1)
	`, ids, patch); err != nil {
		return fmt.Errorf("store: mark dispatches unattributed by id %v: %w", ids, err)
	}
	return nil
}

// DispatchWindowsForSession returns every dispatch attributable to
// sessionID, as the harvest.DispatchWindow shape attribution needs and
// nothing more -- the dispatch-grain counterpart of the stage windows
// storeWindowSource resolves through QueryStageRuns (cmd/flowd/main.go).
//
// A dispatch names its session the same way a stage run does: by the
// literal session_token its dispatcher was running under, which is the
// dispatching session's token and not a token of the subagent's own -- a
// subagent's usage is written into its dispatcher's transcript, so that is
// the session whose records can ever fall inside this window. The token is
// bound to a session id on stage_runs alone (0008_stage_run_session_token.sql),
// by the harvester itself, so this reads the binding back from there
// rather than duplicating it onto dispatches: one fact, one place, and a
// dispatch recorded before its token bound becomes attributable the moment
// it does, with nothing to backfill.
//
// agent_id comes back on the window because attribution matches on it
// before it reaches the interval rule: two dispatches running concurrently
// against one session have overlapping intervals, and the identifier is
// the only thing that separates them. A NULL is read back as "", which
// means "the harness reported none" and never matches another absent id --
// bestDispatchWindow (internal/harvest/attribute.go) holds that rule, and
// this query's job is only to stop the column from being silently
// unavailable to it.
//
// The rows are ordered by (started_at, id), not by started_at alone.
// Several dispatches sharing one started_at is ordinary here -- a review
// panel dispatches its slots at once -- and with no secondary key which of
// the tied rows came back first was left to the plan, so a harvest cycle
// could resolve a tie one way and the next cycle the other. The id is the
// tie-break bestDispatchWindow states for itself, so the two agree rather
// than one of them depending on the other.
//
// A dispatch carrying no session token is not returned at all. There is no
// window to build for it -- nothing states which transcript its records
// would be in -- and inventing one would be exactly the guess this
// requirement exists to remove.
//
// Returning harvest.DispatchWindow directly, rather than a store-local
// type an adapter converts, is what lets *store.Store satisfy
// harvest.DispatchWindowSource with no adapter at all (cmd/flowd asserts
// it at compile time). The dependency runs store -> harvest, which is the
// direction that keeps internal/harvest importing nothing from
// internal/store -- the property TestHarvestNeedsNoDatabase rests on.
func (s *Store) DispatchWindowsForSession(ctx context.Context, sessionID string) ([]harvest.DispatchWindow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT d.id, d.started_at, d.ended_at, d.agent_id
		FROM dispatches d
		WHERE d.session_token IN (
			SELECT sr.session_token FROM stage_runs sr
			WHERE sr.session_id = $1 AND sr.session_token IS NOT NULL
		)
		ORDER BY d.started_at, d.id
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("store: dispatch windows for session %s: %w", sessionID, err)
	}
	defer rows.Close()

	var out []harvest.DispatchWindow
	for rows.Next() {
		var agentID *string
		var w harvest.DispatchWindow
		if err := rows.Scan(&w.DispatchID, &w.StartedAt, &w.EndedAt, &agentID); err != nil {
			return nil, fmt.Errorf("store: dispatch windows for session %s: scan: %w", sessionID, err)
		}
		w.AgentID = derefOrEmpty(agentID)
		out = append(out, w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: dispatch windows for session %s: %w", sessionID, err)
	}
	return out, nil
}

// RunRecord returns one change's whole derived record: its dispatches in
// seq order and its findings in the order the numbers their refs spell,
// which is the order every rendering of the record reads them in.
//
// A change the store has never heard of is ErrChangeNotFound; a change that
// exists and holds no rows is an empty record with no error. The two are
// deliberately distinct, because the render's own report distinguishes them
// -- "the store holds no rows of this kind for this change" is a value, not
// a failure, and it must never be reachable by mistyping a change name.
//
// All three reads share one REPEATABLE READ, read-only transaction
// (queryTxOptions), so a write landing mid-read cannot produce a record
// whose findings reference a dispatch its dispatch list does not contain.
func (s *Store) RunRecord(ctx context.Context, projectKey, change string) (records.Run, error) {
	tx, err := s.pool.BeginTx(ctx, queryTxOptions())
	if err != nil {
		return records.Run{}, fmt.Errorf("store: run record for %s/%s: begin: %w", projectKey, change, err)
	}
	// Rollback after a successful Commit is a documented no-op in pgx; its
	// error carries nothing actionable here, so it is discarded explicitly
	// rather than checked.
	defer func() { _ = tx.Rollback(ctx) }()

	var changeID int64
	if err := tx.QueryRow(ctx,
		`SELECT id FROM changes WHERE project_key = $1 AND name = $2`, projectKey, change,
	).Scan(&changeID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return records.Run{}, fmt.Errorf("%w: %s/%s", ErrChangeNotFound, projectKey, change)
		}
		return records.Run{}, fmt.Errorf("store: run record for %s/%s: %w", projectKey, change, err)
	}

	dispatches, err := readDispatches(ctx, tx, changeID)
	if err != nil {
		return records.Run{}, fmt.Errorf("store: run record for %s/%s: %w", projectKey, change, err)
	}
	findings, err := readFindings(ctx, tx, changeID)
	if err != nil {
		return records.Run{}, fmt.Errorf("store: run record for %s/%s: %w", projectKey, change, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return records.Run{}, fmt.Errorf("store: run record for %s/%s: commit: %w", projectKey, change, err)
	}

	return records.Run{Change: change, Dispatches: dispatches, Findings: findings}, nil
}

func readDispatches(ctx context.Context, tx pgx.Tx, changeID int64) ([]records.Dispatch, error) {
	rows, err := tx.Query(ctx, `
		SELECT `+dispatchColumns+`
		FROM dispatches
		WHERE change_id = $1
		ORDER BY seq
	`, changeID)
	if err != nil {
		return nil, fmt.Errorf("read dispatches: %w", err)
	}
	defer rows.Close()

	var out []records.Dispatch
	for rows.Next() {
		d, err := scanDispatchRow(rows)
		if err != nil {
			return nil, fmt.Errorf("read dispatches: scan: %w", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read dispatches: %w", err)
	}
	return out, nil
}

// readFindings orders findings by the number their refs spell rather than
// by the refs themselves. Lexically, F10 sorts between F1 and F2, so a
// panel that raised ten findings would render its tenth second -- an order
// a reader would take for the order the panel actually worked in. A ref no
// digits can be read out of sorts last (NULLS LAST) and lexically among
// its own kind, so a value this change does not control degrades to a
// stable position instead of failing the cast and taking the whole record
// down with it.
//
// dispatch_id is read back out as the raising dispatch's seq, through a
// LEFT JOIN so a finding no dispatch raised still returns. Seq is the
// identifier the wire shape carries; the row id never leaves this package.
func readFindings(ctx context.Context, tx pgx.Tx, changeID int64) ([]records.Finding, error) {
	rows, err := tx.Query(ctx, `
		SELECT f.ref, d.seq, f.round, f.slot, f.severity, f.location, f.note, f.status, f.reproducer
		FROM findings f
		LEFT JOIN dispatches d ON d.id = f.dispatch_id
		WHERE f.change_id = $1
		ORDER BY NULLIF(regexp_replace(f.ref, '\D', '', 'g'), '')::int NULLS LAST, f.ref
	`, changeID)
	if err != nil {
		return nil, fmt.Errorf("read findings: %w", err)
	}
	defer rows.Close()

	var out []records.Finding
	for rows.Next() {
		var (
			f           records.Finding
			dispatchSeq *int
			location    *string
			reproducer  *string
		)
		if err := rows.Scan(&f.Ref, &dispatchSeq, &f.Round, &f.Slot, &f.Severity, &location,
			&f.Note, &f.Status, &reproducer); err != nil {
			return nil, fmt.Errorf("read findings: scan: %w", err)
		}
		f.DispatchSeq = dispatchSeq
		f.Location = derefOrEmpty(location)
		f.Reproducer = derefOrEmpty(reproducer)
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read findings: %w", err)
	}
	return out, nil
}

// nullIfEmpty maps an empty string to SQL NULL. The record wire types carry
// their optional fields as plain strings, so "" is how an absent task id,
// slot, commit or location arrives; storing it as NULL keeps absence
// distinct from a recorded empty value in the table itself, which is what a
// query filtering on `task_id IS NULL` depends on.
func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// derefOrEmpty is nullIfEmpty's inverse on the way out: a NULL column reads
// back as the empty string the wire type omits.
func derefOrEmpty(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
