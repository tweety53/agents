package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/tweety53/agents/stats/internal/harvest"
	"github.com/tweety53/agents/stats/internal/records"
)

// ErrFindingNotFound is returned by SetFindingStatus when the change holds
// no finding under the given ref. It is typed so internal/api can answer
// 404 rather than letting a caller's typo surface as a 500 -- the
// distinction between "you asked for something that isn't there" and "this
// server is broken".
var ErrFindingNotFound = errors.New("store: finding not found")

// ErrDeferredNotMinor is returned by SetFindingStatus when a `deferred
// <reason>` status is set against a finding whose severity is not Minor.
// design.md's `deferred <reason>` section: a Critical or Important finding is
// always fixed, never deferred, so the store -- not just the CLI's
// validator, which cannot see a finding's severity -- refuses the write
// here rather than storing a status the finding is not allowed to carry.
var ErrDeferredNotMinor = errors.New("store: deferred status is Minor-only")

// ErrFindingLinkInvalid is returned by UpsertFinding when a finding's
// lineage carries a link the change cannot honour: a supersedes or
// regression-of ref the change holds no finding under, or a finding
// linked to itself. It is typed for the same reason ErrFindingNotFound is
// -- internal/api answers 400 on it, so a mistyped ref is a definitive
// CLI refusal rather than a write that journals for a replay that could
// never succeed -- and it names the offending ref, because a refusal that
// did not say which hop was bad would send the caller back to the same
// guesswork this column exists to end.
var ErrFindingLinkInvalid = errors.New("store: finding lineage link is invalid")

// ErrCategoryNotDeferred is returned by SetFindingStatus and UpsertFinding
// when a non-empty deferral category rides on a status that is not
// `deferred`. The category names the mechanism a finding was deferred for
// (0025_finding_deferral_category.sql), so on any other status it is a
// contradiction rather than a value -- refused here, where this package's
// own cross-column rules live, so neither the route nor a replayed write
// can land one.
var ErrCategoryNotDeferred = errors.New("store: deferral category is deferred-only")

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

// findingsSupersedesFK, findingsRegressionOfFK,
// findingsSupersedesNotSelfCheck and findingsRegressionOfNotSelfCheck are
// the names given, explicitly, to the two lineage foreign keys and the two
// self-link CHECKs in 0024_finding_lineage.sql. UpsertFinding checks for
// these exact constraint names when it translates a foreign-key or CHECK
// violation into ErrFindingLinkInvalid, so it knows WHICH link was refused
// and can name the ref that hop carried -- the same
// constraint-name-is-contract reasoning findingsRefConstraint records. All
// four are constants, not literals at the switch: a migration renaming a
// constraint then fails here as a wrong-name mismatch the tests catch,
// never as a silently unmatched case that surfaces as a 500.
const (
	findingsSupersedesFK             = "findings_supersedes_fk"
	findingsRegressionOfFK           = "findings_regression_of_fk"
	findingsSupersedesNotSelfCheck   = "findings_supersedes_not_self"
	findingsRegressionOfNotSelfCheck = "findings_regression_of_not_self"
)

// decisionsSessionConstraint is the name given, explicitly, to the
// UNIQUE (change_id, session_token) constraint in 0019_decisions.sql.
// RecordDecision names it in its ON CONFLICT clause for the same reason
// findingsRefConstraint is named rather than left as a column list: a
// replayed write reaches the row a run already recorded instead of being
// refused or silently matching some other unique index over the same
// columns.
const decisionsSessionConstraint = "decisions_session_key"

// ErrInvalidDecision is returned by RecordDecision when the session token
// is empty or the decision body is not valid JSON -- the store-level
// counterpart to api.ErrInvalidRecord, declared here rather than reused
// from internal/api because internal/store imports nothing above it in the
// dependency graph.
var ErrInvalidDecision = errors.New("store: invalid decision")

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
	          cause, session_token, started_at, ended_at, metrics, notes, agent_id, dispatch_key,
	          diff_base, effort`

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
		cause        *string
		sessionToken *string
		notes        *string
		dispatchKey  *string
		diffBase     *string
		bag          []byte
	)
	if err := row.Scan(
		&d.ID, &d.Seq, &d.StageRunID, &taskID, &d.Role, &slot, &d.Model, &commitSHA, &outcome,
		&cause, &sessionToken, &d.StartedAt, &d.EndedAt, &bag, &notes, &agentID, &dispatchKey,
		&diffBase, &d.Effort,
	); err != nil {
		return records.Dispatch{}, err
	}
	d.AgentID = derefOrEmpty(agentID)
	d.TaskID = derefOrEmpty(taskID)
	d.Slot = derefOrEmpty(slot)
	d.CommitSHA = derefOrEmpty(commitSHA)
	d.Outcome = derefOrEmpty(outcome)
	d.Cause = derefOrEmpty(cause)
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
	// effort has no nullIfEmpty treatment: the column is NOT NULL, and the
	// vocabulary's own "default" word IS the absent case -- an absent
	// effort is a known fact, never an unknown one (design.md's
	// Enforcement), so a caller that recorded none is defaulted here
	// rather than left to a SQL DEFAULT the explicit column list would
	// never reach.
	effort := in.Effort
	if effort == "" {
		effort = "default"
	}

	return scanDispatchRow(s.pool.QueryRow(ctx, `
		INSERT INTO dispatches (
			change_id, stage_run_id, seq, task_id, role, slot, model, commit_sha, outcome,
			session_token, started_at, ended_at, metrics, notes, agent_id, dispatch_key,
			diff_base, effort
		)
		SELECT
			c.id, $3,
			COALESCE((SELECT MAX(d.seq) FROM dispatches d WHERE d.change_id = c.id), 0) + 1,
			$4, $5, $6, $7, $8, $9, $10, $11, $12, $13::jsonb, $14, $15, $16, $17, $18
		FROM changes c
		WHERE c.project_key = $1 AND c.name = $2
		ON CONFLICT ON CONSTRAINT `+dispatchesKeyConstraint+` DO UPDATE SET
			dispatch_key = EXCLUDED.dispatch_key
		RETURNING `+dispatchColumns+`
	`,
		projectKey, change, in.StageRunID, nullIfEmpty(in.TaskID), in.Role, nullIfEmpty(in.Slot),
		in.Model, nullIfEmpty(in.CommitSHA), nullIfEmpty(in.Outcome), nullIfEmpty(in.SessionToken),
		in.StartedAt, in.EndedAt, metrics, nullIfEmpty(in.Notes), nullIfEmpty(in.AgentID),
		nullIfEmpty(in.Key), nullIfEmpty(in.DiffBase), effort,
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
//
// cause is the one opposite case, and deliberately so: it exists only
// beside an outcome of `blocked`, so it takes the outcome's own plain
// last-write-wins SET rather than agent_id's COALESCE -- an end that
// records a non-blocked outcome with no cause clears a stale cause, and a
// row never goes on claiming a block its outcome no longer reports
// (KAN-510).
func (s *Store) EndDispatch(ctx context.Context, projectKey, change string, in records.DispatchEnd) (records.Dispatch, error) {
	out, err := scanDispatchRow(s.pool.QueryRow(ctx, `
		UPDATE dispatches d
		SET commit_sha = $5, outcome = $6, ended_at = $7, agent_id = COALESCE($8, d.agent_id),
		    cause = $9
		FROM changes c
		WHERE c.id = d.change_id AND c.project_key = $1 AND c.name = $2
		  AND d.session_token = $3 AND d.dispatch_key = $4
		RETURNING `+qualifiedDispatchColumns("d"),
		projectKey, change, in.SessionToken, in.Key,
		nullIfEmpty(in.CommitSHA), nullIfEmpty(in.Outcome), in.EndedAt, nullIfEmpty(in.AgentID),
		nullIfEmpty(in.Cause),
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
// A finding recorded with a deferral category whose status is not
// `deferred` is refused with ErrCategoryNotDeferred before the statement
// runs: the category names the mechanism a finding was deferred for, so on
// any other status it is a caller contradiction, the same shape of mistake
// the store's Minor-only rule refuses one layer out. The conflict update
// rewrites the column from EXCLUDED like every other restated field, so a
// fix round restating F1 as fixed clears a category an earlier deferral
// had set.
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
		supersedes  *string
		regression  *string
		category    *string
		created     bool
	)

	if in.Category != "" && !strings.HasPrefix(in.Status, "deferred") {
		return records.Finding{}, false, fmt.Errorf("%w: %s in %s/%s", ErrCategoryNotDeferred, in.Ref, projectKey, change)
	}

	err := s.pool.QueryRow(ctx, `
		INSERT INTO findings (
			change_id, dispatch_id, ref, round, slot, severity, location, note, status, deferral_category, reproducer,
			supersedes, regression_of
		)
		SELECT
			c.id,
			(SELECT d.id FROM dispatches d WHERE d.change_id = c.id AND d.seq = $4::int),
			$3, $5, $6, $7, $8, $9, $10, $12, $11, $13, $14
		FROM changes c
		WHERE c.project_key = $1 AND c.name = $2
		ON CONFLICT ON CONSTRAINT `+findingsRefConstraint+` DO UPDATE SET
			dispatch_id   = EXCLUDED.dispatch_id,
			round         = EXCLUDED.round,
			slot          = EXCLUDED.slot,
			severity      = EXCLUDED.severity,
			location      = EXCLUDED.location,
			note          = EXCLUDED.note,
			status        = EXCLUDED.status,
			deferral_category = EXCLUDED.deferral_category,
			reproducer    = EXCLUDED.reproducer,
			supersedes    = EXCLUDED.supersedes,
			regression_of = EXCLUDED.regression_of
		RETURNING
			ref,
			(SELECT d.seq FROM dispatches d WHERE d.id = findings.dispatch_id),
			round, slot, severity, location, note, status, deferral_category, reproducer, supersedes, regression_of,
			xmax = 0
	`,
		projectKey, change, in.Ref, in.DispatchSeq, in.Round, in.Slot, in.Severity,
		nullIfEmpty(in.Location), in.Note, in.Status, nullIfEmpty(in.Reproducer),
		nullIfEmpty(in.Category), nullIfEmpty(in.Supersedes), nullIfEmpty(in.RegressionOf),
	).Scan(&out.Ref, &dispatchSeq, &out.Round, &out.Slot, &out.Severity, &location,
		&out.Note, &out.Status, &category, &reproducer, &supersedes, &regression, &created)
	if err != nil {
		// A lineage link the change cannot honour is the caller's typo,
		// not the store's failure: the FK refuses a ref the change does
		// not hold and the CHECK refuses a self-link, and both are
		// translated here into the typed error naming the offending ref,
		// so the refusal survives to the caller instead of surfacing as a
		// raw Postgres message from a 500.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch {
			case pgErr.Code == "23503" && pgErr.ConstraintName == findingsSupersedesFK:
				return records.Finding{}, false, fmt.Errorf("%w: finding %s in %s/%s supersedes %s, which this change does not hold", ErrFindingLinkInvalid, in.Ref, projectKey, change, in.Supersedes)
			case pgErr.Code == "23503" && pgErr.ConstraintName == findingsRegressionOfFK:
				return records.Finding{}, false, fmt.Errorf("%w: finding %s in %s/%s is a regression of %s, which this change does not hold", ErrFindingLinkInvalid, in.Ref, projectKey, change, in.RegressionOf)
			case pgErr.Code == "23514" && pgErr.ConstraintName == findingsSupersedesNotSelfCheck:
				return records.Finding{}, false, fmt.Errorf("%w: finding %s in %s/%s cannot supersede itself", ErrFindingLinkInvalid, in.Ref, projectKey, change)
			case pgErr.Code == "23514" && pgErr.ConstraintName == findingsRegressionOfNotSelfCheck:
				return records.Finding{}, false, fmt.Errorf("%w: finding %s in %s/%s cannot be a regression of itself", ErrFindingLinkInvalid, in.Ref, projectKey, change)
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return records.Finding{}, false, fmt.Errorf("%w: %s/%s", ErrChangeNotFound, projectKey, change)
		}
		return records.Finding{}, false, fmt.Errorf("store: upsert finding %s for %s/%s: %w", in.Ref, projectKey, change, err)
	}

	out.DispatchSeq = dispatchSeq
	out.Location = derefOrEmpty(location)
	out.Reproducer = derefOrEmpty(reproducer)
	out.Category = derefOrEmpty(category)
	out.Supersedes = derefOrEmpty(supersedes)
	out.RegressionOf = derefOrEmpty(regression)
	return out, created, nil
}

// SetFindingStatus updates one finding's status and its deferral category,
// and nothing else -- the whole of what a fix round changes about a finding
// it has resolved. A ref the change holds no finding under is
// ErrFindingNotFound, not a silent no-op, so a caller's typo is reported
// rather than looking like a successful update.
//
// The category rides beside a `deferred <reason>` status and nowhere else:
// a non-empty category on any other status is refused with
// ErrCategoryNotDeferred before the statement runs, and a category-less
// call clears whatever an earlier deferral wrote, since the column's only
// meaning is the one its status gives it.
//
// A `deferred <reason>` status carries the extra `AND f.severity ILIKE 'Minor'`
// clause design.md's `deferred <reason>` section requires, so the update
// itself never lands a deferral against a Critical or Important finding.
// Severity is free text a caller writes verbatim (this package's
// Reviewers/Decisions queries already match it case-insensitively), so the
// guard is case-insensitive too -- a lowercase "minor" is still Minor. Zero
// rows affected is ambiguous for that status alone -- the ref may not
// exist, or it may exist at the wrong severity -- so a second, read-only
// SELECT distinguishes the two: a row found is ErrDeferredNotMinor, no row
// is ErrFindingNotFound, the same sentinel every other status's zero-row
// case already returns.
func (s *Store) SetFindingStatus(ctx context.Context, projectKey, change, ref, status, category string) error {
	deferred := strings.HasPrefix(status, "deferred")

	if category != "" && !deferred {
		return fmt.Errorf("%w: %s in %s/%s", ErrCategoryNotDeferred, ref, projectKey, change)
	}

	query := `
		UPDATE findings f
		SET status = $4, deferral_category = $5
		FROM changes c
		WHERE c.id = f.change_id AND c.project_key = $1 AND c.name = $2 AND f.ref = $3`
	if deferred {
		query += ` AND f.severity ILIKE 'Minor'`
	}

	tag, err := s.pool.Exec(ctx, query, projectKey, change, ref, status, nullIfEmpty(category))
	if err != nil {
		return fmt.Errorf("store: set finding %s status for %s/%s: %w", ref, projectKey, change, err)
	}
	if tag.RowsAffected() > 0 {
		return nil
	}

	if deferred {
		var severity string
		err := s.pool.QueryRow(ctx, `
			SELECT f.severity
			FROM findings f
			JOIN changes c ON c.id = f.change_id
			WHERE c.project_key = $1 AND c.name = $2 AND f.ref = $3
		`, projectKey, change, ref).Scan(&severity)
		switch {
		case err == nil:
			return fmt.Errorf("%w: %s in %s/%s has severity %s, not Minor", ErrDeferredNotMinor, ref, projectKey, change, severity)
		case !errors.Is(err, pgx.ErrNoRows):
			return fmt.Errorf("store: set finding %s status for %s/%s: %w", ref, projectKey, change, err)
		}
	}
	return fmt.Errorf("%w: %s in %s/%s", ErrFindingNotFound, ref, projectKey, change)
}

// RecordDecision records one run's dynamic decision, or replaces the one
// already recorded under the same session token for the same change --
// exactly UpsertFinding's shape, moved from ref to session token as the key
// design.md's decisions-jsonb-row decision names. The returned bool is
// created, read from the returned row's xmax the same way UpsertFinding
// reads it: a row this statement genuinely inserted carries xmax = 0, and a
// row ON CONFLICT DO UPDATE reached instead carries the updating
// transaction's id.
//
// An unknown (projectKey, change) pair is ErrChangeNotFound. An empty
// SessionToken or a Decision that is not valid JSON is ErrInvalidDecision,
// checked before the statement runs so the store never round-trips to
// Postgres for a write jsonb would reject anyway.
func (s *Store) RecordDecision(ctx context.Context, projectKey, change string, in records.Decision) (records.Decision, bool, error) {
	if in.SessionToken == "" {
		return records.Decision{}, false, fmt.Errorf("%w: sessionToken is required", ErrInvalidDecision)
	}
	if !json.Valid(in.Decision) {
		return records.Decision{}, false, fmt.Errorf("%w: decision must be valid JSON", ErrInvalidDecision)
	}

	var (
		out     records.Decision
		created bool
	)
	err := s.pool.QueryRow(ctx, `
		INSERT INTO decisions (change_id, session_token, decision)
		SELECT c.id, $3, $4
		FROM changes c
		WHERE c.project_key = $1 AND c.name = $2
		ON CONFLICT ON CONSTRAINT `+decisionsSessionConstraint+` DO UPDATE SET
			decision    = EXCLUDED.decision,
			recorded_at = now()
		RETURNING id, session_token, recorded_at, decision, xmax = 0
	`, projectKey, change, in.SessionToken, []byte(in.Decision)).
		Scan(&out.ID, &out.SessionToken, &out.RecordedAt, &out.Decision, &created)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return records.Decision{}, false, fmt.Errorf("%w: %s/%s", ErrChangeNotFound, projectKey, change)
		}
		return records.Decision{}, false, fmt.Errorf("store: record decision for %s/%s: %w", projectKey, change, err)
	}
	return out, created, nil
}

// ListDecisions reads a change's recorded decisions, newest first -- what
// `flow record decisions -change <name>` reports, and what a resumed run
// reads to recover the roster and model a stopped run already chose.
func (s *Store) ListDecisions(ctx context.Context, projectKey, change string) ([]records.Decision, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT d.id, d.session_token, d.recorded_at, d.decision
		FROM decisions d
		JOIN changes c ON c.id = d.change_id
		WHERE c.project_key = $1 AND c.name = $2
		ORDER BY d.recorded_at DESC, d.id DESC
	`, projectKey, change)
	if err != nil {
		return nil, fmt.Errorf("store: list decisions for %s/%s: %w", projectKey, change, err)
	}
	defer rows.Close()

	var out []records.Decision
	for rows.Next() {
		var d records.Decision
		if err := rows.Scan(&d.ID, &d.SessionToken, &d.RecordedAt, &d.Decision); err != nil {
			return nil, fmt.Errorf("store: scan decision for %s/%s: %w", projectKey, change, err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list decisions for %s/%s: %w", projectKey, change, err)
	}
	return out, nil
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

// DispatchWindowsForAgent returns every dispatch row recorded with agentID,
// ordered by (started_at, id) -- the agent-keyed counterpart of
// DispatchWindowsForSession above. One resumed agent shares its agentId
// across several dispatch rows (the conductor and per-task fix rounds do
// exactly that), and the agent-file attribution pass
// (internal/harvest's attributeAgentFileRecords) splits that agent's
// transcript usage among its own rows by this order.
//
// The ordering and the empty-agentId filter are the same rules the session
// query carries: (started_at, id) because seconds-resolution hand-typed
// starts make exact ties ordinary, and an empty agent_id means "not
// reported" and never matches a lookup -- including an empty one, which
// this query's caller can never make, the filter keeping the rule true at
// the store regardless.
//
// The window shape and the direct harvest.DispatchWindow return follow the
// session query's own reasoning: *store.Store satisfies
// harvest.AgentWindowSource with no adapter (cmd/flowd asserts harvest.Deps
// at compile time), and internal/harvest keeps importing nothing from
// internal/store.
func (s *Store) DispatchWindowsForAgent(ctx context.Context, agentID string) ([]harvest.DispatchWindow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT d.id, d.started_at, d.ended_at, d.agent_id
		FROM dispatches d
		WHERE d.agent_id = $1 AND d.agent_id <> ''
		ORDER BY d.started_at, d.id
	`, agentID)
	if err != nil {
		return nil, fmt.Errorf("store: dispatch windows for agent %s: %w", agentID, err)
	}
	defer rows.Close()

	var out []harvest.DispatchWindow
	for rows.Next() {
		var scanned *string
		var w harvest.DispatchWindow
		if err := rows.Scan(&w.DispatchID, &w.StartedAt, &w.EndedAt, &scanned); err != nil {
			return nil, fmt.Errorf("store: dispatch windows for agent %s: scan: %w", agentID, err)
		}
		w.AgentID = derefOrEmpty(scanned)
		out = append(out, w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: dispatch windows for agent %s: %w", agentID, err)
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
// All five reads share one REPEATABLE READ, read-only transaction
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
	passes, err := readPanelPasses(ctx, tx, changeID)
	if err != nil {
		return records.Run{}, fmt.Errorf("store: run record for %s/%s: %w", projectKey, change, err)
	}
	mutations, err := readPanelMutations(ctx, tx, changeID)
	if err != nil {
		return records.Run{}, fmt.Errorf("store: run record for %s/%s: %w", projectKey, change, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return records.Run{}, fmt.Errorf("store: run record for %s/%s: commit: %w", projectKey, change, err)
	}

	return records.Run{Change: change, Dispatches: dispatches, Findings: findings,
		Passes: passes, Mutations: mutations}, nil
}

// RecordPass records one pass-log entry -- a pass-by-pass metadata line of
// the review panel's record -- against a change. Append-only, exactly the
// guard-log writes' shape: the parent records each fact once as it
// arises, and a replayed write can at worst duplicate a line, the same
// cosmetic risk guard_verdicts already carries -- no dedup key exists for
// "the same fact phrased identically". An unknown (projectKey, change)
// pair is ErrChangeNotFound, the same sentinel every change-scoped write
// here returns.
func (s *Store) RecordPass(ctx context.Context, projectKey, change string, in records.Pass) (records.Pass, error) {
	var out records.Pass
	err := s.pool.QueryRow(ctx, `
		INSERT INTO panel_passes (change_id, round, note)
		SELECT c.id, $3, $4
		FROM changes c
		WHERE c.project_key = $1 AND c.name = $2
		RETURNING id, round, note
	`, projectKey, change, in.Round, in.Note).Scan(&out.ID, &out.Round, &out.Note)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return records.Pass{}, fmt.Errorf("%w: %s/%s", ErrChangeNotFound, projectKey, change)
		}
		return records.Pass{}, fmt.Errorf("store: record pass for %s/%s: %w", projectKey, change, err)
	}
	return out, nil
}

// RecordMutation records one fix-mutation: line of the fix round's
// mutation proof -- the path, what was mutated, and the test that failed
// (or, on the contract's exemption form, mutated "none" and test carrying
// the reason). Append-only and Round-scoped like RecordPass; the
// fix-mutations-total count is the round's own row count and is rendered,
// never stored. An unknown (projectKey, change) pair is ErrChangeNotFound.
func (s *Store) RecordMutation(ctx context.Context, projectKey, change string, in records.Mutation) (records.Mutation, error) {
	var out records.Mutation
	err := s.pool.QueryRow(ctx, `
		INSERT INTO panel_mutations (change_id, round, path, mutated, test)
		SELECT c.id, $3, $4, $5, $6
		FROM changes c
		WHERE c.project_key = $1 AND c.name = $2
		RETURNING id, round, path, mutated, test
	`, projectKey, change, in.Round, in.Path, in.Mutated, in.Test).
		Scan(&out.ID, &out.Round, &out.Path, &out.Mutated, &out.Test)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return records.Mutation{}, fmt.Errorf("%w: %s/%s", ErrChangeNotFound, projectKey, change)
		}
		return records.Mutation{}, fmt.Errorf("store: record mutation for %s/%s: %w", projectKey, change, err)
	}
	return out, nil
}

// readPanelPasses reads a change's pass-log entries in (round, id) order,
// inside the caller's transaction -- RunRecord's REPEATABLE READ, so a
// pass recorded mid-read cannot appear in a record whose other rows
// predate it.
func readPanelPasses(ctx context.Context, tx pgx.Tx, changeID int64) ([]records.Pass, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, round, note
		FROM panel_passes
		WHERE change_id = $1
		ORDER BY round, id
	`, changeID)
	if err != nil {
		return nil, fmt.Errorf("read panel passes: %w", err)
	}
	defer rows.Close()

	var out []records.Pass
	for rows.Next() {
		var p records.Pass
		if err := rows.Scan(&p.ID, &p.Round, &p.Note); err != nil {
			return nil, fmt.Errorf("read panel passes: scan: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read panel passes: %w", err)
	}
	return out, nil
}

// readPanelMutations is readPanelPasses' mutation-proof counterpart: a
// change's fix-mutation rows in (round, id) order, same transaction.
func readPanelMutations(ctx context.Context, tx pgx.Tx, changeID int64) ([]records.Mutation, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, round, path, mutated, test
		FROM panel_mutations
		WHERE change_id = $1
		ORDER BY round, id
	`, changeID)
	if err != nil {
		return nil, fmt.Errorf("read panel mutations: %w", err)
	}
	defer rows.Close()

	var out []records.Mutation
	for rows.Next() {
		var m records.Mutation
		if err := rows.Scan(&m.ID, &m.Round, &m.Path, &m.Mutated, &m.Test); err != nil {
			return nil, fmt.Errorf("read panel mutations: scan: %w", err)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read panel mutations: %w", err)
	}
	return out, nil
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
		SELECT f.ref, d.seq, f.round, f.slot, f.severity, f.location, f.note, f.status, f.deferral_category, f.reproducer,
		       f.supersedes, f.regression_of
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
			supersedes  *string
			regression  *string
			category    *string
		)
		if err := rows.Scan(&f.Ref, &dispatchSeq, &f.Round, &f.Slot, &f.Severity, &location,
			&f.Note, &f.Status, &category, &reproducer, &supersedes, &regression); err != nil {
			return nil, fmt.Errorf("read findings: scan: %w", err)
		}
		f.DispatchSeq = dispatchSeq
		f.Location = derefOrEmpty(location)
		f.Reproducer = derefOrEmpty(reproducer)
		f.Category = derefOrEmpty(category)
		f.Supersedes = derefOrEmpty(supersedes)
		f.RegressionOf = derefOrEmpty(regression)
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read findings: %w", err)
	}
	return out, nil
}

// verdictColumns is the column list every read of a guard_verdicts row
// selects, in the order scanVerdictRow scans them -- the same
// one-list-shared-by-every-reader shape dispatchColumns gives insertDispatch,
// EndDispatch and readDispatches. Callers reading through a join qualify it
// with qualifiedVerdictColumns; FlagVerdictFalsePositive's UPDATE has only
// one table in scope and uses it bare.
const verdictColumns = `id, guard, worktree, verdict, recorded_at,
	          false_positive, false_positive_reason, flagged_at`

// qualifiedVerdictColumns is verdictColumns with every name qualified by
// alias, for RecordVerdict's and ListVerdicts' joins -- the same reasoning
// qualifiedDispatchColumns carries.
func qualifiedVerdictColumns(alias string) string {
	cols := strings.Split(verdictColumns, ",")
	for i, c := range cols {
		cols[i] = alias + "." + strings.TrimSpace(c)
	}
	return strings.Join(cols, ", ")
}

// scanVerdictRow decodes one row of verdictColumns, plus a change name
// joined in ahead of it, into a records.Verdict.
func scanVerdictRow(row dispatchRowScanner) (records.Verdict, error) {
	var (
		v      records.Verdict
		change string
		reason *string
	)
	if err := row.Scan(
		&change, &v.ID, &v.Guard, &v.Worktree, &v.Verdict, &v.RecordedAt,
		&v.FalsePositive, &reason, &v.FlaggedAt,
	); err != nil {
		return records.Verdict{}, err
	}
	v.Change = change
	v.FalsePositiveReason = derefOrEmpty(reason)
	return v, nil
}

// RecordVerdict records one guard's verdict against a change and worktree.
// Unlike RecordDispatch it allocates no seq and dedups on nothing: a guard
// that runs the same check twice (retried by an operator, or re-entered
// mid-flight) is two distinct verdicts, not one replayed write, so every
// call inserts a new row.
//
// RecordedAt is the caller's own timestamp, exactly as in.StartedAt is on
// RecordDispatch -- the CLI stamps time.Now() before the call, rather than
// this method defaulting it, so a journalled write replayed later still
// carries the moment the guard actually ran rather than the moment the
// journal was drained.
//
// An unknown (projectKey, change) pair is ErrChangeNotFound, the same
// sentinel RecordDispatch returns for the same shape of failure.
func (s *Store) RecordVerdict(ctx context.Context, projectKey, change string, in records.Verdict) (records.Verdict, error) {
	row := s.pool.QueryRow(ctx, `
		WITH ins AS (
			INSERT INTO guard_verdicts (change_id, guard, worktree, verdict, recorded_at)
			SELECT c.id, $3, $4, $5, $6
			FROM changes c
			WHERE c.project_key = $1 AND c.name = $2
			RETURNING `+verdictColumns+`
		)
		SELECT $2, `+qualifiedVerdictColumns("ins")+`
		FROM ins
	`, projectKey, change, in.Guard, in.Worktree, in.Verdict, in.RecordedAt)

	out, err := scanVerdictRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return records.Verdict{}, fmt.Errorf("%w: %s/%s", ErrChangeNotFound, projectKey, change)
		}
		return records.Verdict{}, fmt.Errorf("store: record verdict for %s/%s: %w", projectKey, change, err)
	}
	return out, nil
}

// FlagVerdictFalsePositive marks the most recent verdict a change recorded
// for the named guard as a false positive, with the operator's reason.
//
// "The most recent one" is deliberate -- see design.md's
// flag-latest-verdict decision -- because the operator at the gate has no
// row id to name: they know the change, the guard and their own
// adjudication, and the newest verdict for that pair is always the one the
// gate just showed them.
//
// Re-flagging the same row is not an error: it overwrites reason and
// flagged_at, so an operator correcting their own wording does not have to
// go through a store-side unflag first.
//
// A (change, guard) pair the store holds no verdict for is
// ErrDispatchNotFound -- not a new sentinel, but a reuse of the one
// EndDispatch already returns for "no row answers to this key", since a
// flag naming no verdict is the same shape of failure: an operator, or a
// guard's advisory call, pointing at a row that never existed.
func (s *Store) FlagVerdictFalsePositive(ctx context.Context, projectKey, change string, in records.VerdictFlag) (records.Verdict, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE guard_verdicts gv
		SET false_positive = true, false_positive_reason = $4, flagged_at = now()
		WHERE gv.id = (
			SELECT gv2.id
			FROM guard_verdicts gv2
			JOIN changes c ON c.id = gv2.change_id
			WHERE c.project_key = $1 AND c.name = $2 AND gv2.guard = $3
			ORDER BY gv2.recorded_at DESC, gv2.id DESC
			LIMIT 1
		)
		RETURNING `+verdictColumns+`
	`, projectKey, change, in.Guard, in.Reason)

	out, err := scanFlaggedVerdictRow(row, change)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return records.Verdict{}, fmt.Errorf("%w: verdict for guard %q in %s/%s", ErrDispatchNotFound, in.Guard, projectKey, change)
		}
		return records.Verdict{}, fmt.Errorf("store: flag verdict false positive for %s/%s guard %q: %w", projectKey, change, in.Guard, err)
	}
	return out, nil
}

// scanFlaggedVerdictRow decodes a verdictColumns row that carries no joined
// change name of its own -- FlagVerdictFalsePositive's UPDATE ... RETURNING
// already knows the change from its caller, unlike a read that joins it in.
func scanFlaggedVerdictRow(row dispatchRowScanner, change string) (records.Verdict, error) {
	var (
		v      records.Verdict
		reason *string
	)
	if err := row.Scan(
		&v.ID, &v.Guard, &v.Worktree, &v.Verdict, &v.RecordedAt,
		&v.FalsePositive, &reason, &v.FlaggedAt,
	); err != nil {
		return records.Verdict{}, err
	}
	v.Change = change
	v.FalsePositiveReason = derefOrEmpty(reason)
	return v, nil
}

// ListVerdicts reads a project's guard verdicts, newest first. guard == ""
// means every guard; falsePositiveOnly restricts to rows an operator has
// flagged. The two filters compose independently, since the guard prompt
// in check-unfinished-work.sh's advisory line needs exactly "this guard,
// flagged only" while a future per-project view would want "every guard,
// flagged only".
func (s *Store) ListVerdicts(ctx context.Context, projectKey, guard string, falsePositiveOnly bool) ([]records.Verdict, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.name, `+qualifiedVerdictColumns("gv")+`
		FROM guard_verdicts gv
		JOIN changes c ON c.id = gv.change_id
		WHERE c.project_key = $1
		  AND ($2 = '' OR gv.guard = $2)
		  AND (NOT $3 OR gv.false_positive)
		ORDER BY gv.recorded_at DESC, gv.id DESC
	`, projectKey, guard, falsePositiveOnly)
	if err != nil {
		return nil, fmt.Errorf("store: list verdicts for %s: %w", projectKey, err)
	}
	defer rows.Close()

	var out []records.Verdict
	for rows.Next() {
		v, err := scanVerdictRow(rows)
		if err != nil {
			return nil, fmt.Errorf("store: list verdicts for %s: scan: %w", projectKey, err)
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list verdicts for %s: %w", projectKey, err)
	}
	return out, nil
}

// incidentColumns is the column list every read of an incidents row
// selects, in the order scanIncidentRow scans them. ListIncidents qualifies
// it with qualifiedIncidentColumns; RecordIncident's plain INSERT ...
// RETURNING, with only one table in scope, uses it bare.
const incidentColumns = `id, guard, symptom, recovery, minutes_lost, occurred_at`

// qualifiedIncidentColumns is incidentColumns with every name qualified by
// alias, for ListIncidents' join -- the same reasoning
// qualifiedDispatchColumns carries.
func qualifiedIncidentColumns(alias string) string {
	cols := strings.Split(incidentColumns, ",")
	for i, c := range cols {
		cols[i] = alias + "." + strings.TrimSpace(c)
	}
	return strings.Join(cols, ", ")
}

// scanIncidentRow decodes one row of incidentColumns, plus a change name
// (possibly absent) joined in ahead of it, into a records.Incident.
func scanIncidentRow(row dispatchRowScanner) (records.Incident, error) {
	var (
		i      records.Incident
		change *string
	)
	if err := row.Scan(
		&change, &i.ID, &i.Guard, &i.Symptom, &i.Recovery, &i.MinutesLost, &i.OccurredAt,
	); err != nil {
		return records.Incident{}, err
	}
	i.Change = derefOrEmpty(change)
	return i, nil
}

// RecordIncident records one per-project incident: a guard, what went
// wrong, the recovery taken and how many minutes it cost.
//
// in.Change, when non-empty, must name a change the project already holds
// -- an unknown name is ErrChangeNotFound, the same sentinel RecordDispatch
// returns for the same shape of failure -- and resolves to change_id. Left
// empty, change_id is NULL: an incident can be recorded against a project
// with no change in flight, or after the change that produced it has
// archived, which is why incidents carries project_key of its own rather
// than deriving it through a change the way guard_verdicts does.
//
// in.OccurredAt left at the zero time takes the column's own now() default
// rather than storing the zero time literally, so a hand-written incident
// that omits it is stamped with when it was actually recorded.
func (s *Store) RecordIncident(ctx context.Context, projectKey string, in records.Incident) (records.Incident, error) {
	var changeID *int64
	if in.Change != "" {
		if err := s.pool.QueryRow(ctx,
			`SELECT id FROM changes WHERE project_key = $1 AND name = $2`,
			projectKey, in.Change,
		).Scan(&changeID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return records.Incident{}, fmt.Errorf("%w: %s/%s", ErrChangeNotFound, projectKey, in.Change)
			}
			return records.Incident{}, fmt.Errorf("store: resolve change for incident in %s: %w", projectKey, err)
		}
	}

	var occurredAt *time.Time
	if !in.OccurredAt.IsZero() {
		occurredAt = &in.OccurredAt
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO incidents (project_key, change_id, guard, symptom, recovery, minutes_lost, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, now()))
		RETURNING `+incidentColumns+`
	`, projectKey, changeID, in.Guard, in.Symptom, in.Recovery, in.MinutesLost, occurredAt)

	var out records.Incident
	if err := row.Scan(&out.ID, &out.Guard, &out.Symptom, &out.Recovery, &out.MinutesLost, &out.OccurredAt); err != nil {
		return records.Incident{}, fmt.Errorf("store: record incident for %s: %w", projectKey, err)
	}
	out.Change = in.Change
	return out, nil
}

// ListIncidents reads a project's incidents, newest first.
func (s *Store) ListIncidents(ctx context.Context, projectKey string) ([]records.Incident, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.name, `+qualifiedIncidentColumns("i")+`
		FROM incidents i
		LEFT JOIN changes c ON c.id = i.change_id
		WHERE i.project_key = $1
		ORDER BY i.occurred_at DESC, i.id DESC
	`, projectKey)
	if err != nil {
		return nil, fmt.Errorf("store: list incidents for %s: %w", projectKey, err)
	}
	defer rows.Close()

	var out []records.Incident
	for rows.Next() {
		i, err := scanIncidentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("store: list incidents for %s: scan: %w", projectKey, err)
		}
		out = append(out, i)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list incidents for %s: %w", projectKey, err)
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
