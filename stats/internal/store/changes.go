package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
)

// State is one of the three pipeline states, in the order the monotonic
// rule allows them to advance.
type State string

const (
	StateStarted    State = "STARTED"
	StateInProgress State = "IN_PROGRESS"
	StateFinished   State = "FINISHED"
)

// ErrMonotonicViolation is returned by PutChange when the write's state is
// earlier in the pipeline than the state already stored. Callers must be
// able to tell this apart from a transport failure: it is a correct
// refusal, not an error to retry or fall back on.
var ErrMonotonicViolation = errors.New("store: refused: write would move state backwards")

// ErrChangeNotFound is returned by GetChange when no record exists for the
// given project and name.
var ErrChangeNotFound = errors.New("store: change not found")

// ErrInvalidState is returned by PutChange when c.State is not one of the
// three canonical pipeline states. It is distinct from
// ErrMonotonicViolation: an invalid state is malformed input, never a
// correct refusal of a backwards move, and callers must be able to tell the
// two apart.
var ErrInvalidState = errors.New("store: invalid state")

// ErrInvalidMainCheckoutPath is returned by PutChange when c.ProjectKey
// names a project that does not yet exist and c.MainCheckoutPath is empty.
// The project row cannot be bootstrapped without it, and accepting an empty
// value would permanently poison the row instead of surfacing the caller's
// mistake.
var ErrInvalidMainCheckoutPath = errors.New("store: main checkout path required to create a new project")

// ErrInvalidMergeBase is returned by PutChange when a c.Repos entry
// records a merge base that is neither absent -- a nil MergeBase, meaning
// none recorded -- nor a 40-character lowercase hexadecimal sha. Like
// ErrInvalidState it is a fault in the caller's own content rather than a
// correct refusal, and errors.Is must be able to tell it apart from both:
// from ErrInvalidState, because a different field of the same payload is
// what is malformed, and from ErrMonotonicViolation, which reports a
// well-formed write arriving out of order. Replaying the identical
// payload produces the identical refusal, so such a write is retired
// rather than retried (IsDefinitiveChangeOutcome, internal/api/changes.go).
var ErrInvalidMergeBase = errors.New("store: invalid merge base")

// mergeBasePattern is the only shape a recorded merge base may take: a
// 40-character lowercase hexadecimal sha, exactly as git prints one.
// Uppercase is excluded deliberately rather than folded -- every producer
// in the pipeline is `git merge-base`, which emits lowercase, so an
// uppercase value is a hand-edit and worth reporting as one.
//
// stats/cmd/flow/state.go carries this same pattern for the CLI's own
// check, and the duplication is deliberate rather than an oversight: the
// two are independent defences at opposite ends of the write path
// (design.md, "Where a merge base is refused"), and sharing one constant
// would make cmd/flow import this package, pulling the store and pgx
// into the CLI's dependency graph for a single regular expression.
var mergeBasePattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

// validStates is the closed set State.IsValid checks c.State against.
var validStates = map[State]bool{
	StateStarted:    true,
	StateInProgress: true,
	StateFinished:   true,
}

// IsValid reports whether s is one of the three canonical pipeline states.
func (s State) IsValid() bool {
	return validStates[s]
}

// Change is the whole state record for one change. PutChange renders every
// field of it — a field left at its zero value is stored as absent, never
// treated as "leave the stored value alone".
type Change struct {
	ProjectKey string
	// MainCheckoutPath bootstraps the project row when it does not yet
	// exist. It is ignored once the project row is present.
	MainCheckoutPath string

	Name  string
	State State

	Branch            *string
	Worktrees         json.RawMessage
	ArtifactURL       *string
	JiraIssue         *string
	PlanningEffort    *string
	Models            json.RawMessage
	ReviewPanelRoster *string
	PRURL             *string

	// Repos is the set of repositories this change affects. It follows the
	// same whole-object rule as every other field of Change: a repository
	// stored for this change but absent from Repos is removed on write. A
	// change confined to c.ProjectKey's own repository leaves this nil or
	// empty. c.ProjectKey itself names only the project whose state
	// directory owns this record -- it is not the list of affected
	// repositories, and Repos must not be mistaken for a derivative of it.
	Repos []Repo

	UpdatedAt time.Time
	UpdatedBy string
}

// PutChange renders the whole record: every field of c is written, and a
// field c leaves at its zero value overwrites whatever was stored, exactly
// as a whole-object write must. The project row is created if absent, using
// c.MainCheckoutPath — which PutChange refuses with
// ErrInvalidMainCheckoutPath if empty and the project does not yet exist,
// rather than silently persisting an unusable row. The write is refused
// with ErrInvalidState if c.State is not one of the three canonical states,
// and with ErrInvalidMergeBase if any c.Repos entry records a merge base
// that is not a sha. Both content checks run before the transaction opens,
// so a refusal has nothing to roll back.
//
// A write is refused with ErrMonotonicViolation when it would move the
// record backwards in either dimension the design names: to a state
// earlier in the pipeline than the one stored, or — at the *same* state —
// to a c.UpdatedAt earlier than the one already recorded. c.UpdatedAt is
// the primary ordering and the pipeline state is only the tiebreaker
// (design.md, "conflicts resolve by updatedAt, with the monotonic-state
// rule as the tiebreaker"); a naive state-rank-only guard would let two
// same-state writes apply in either order and let the chronologically
// older one silently overwrite the newer record's fields, which is exactly
// the case task 6's journal replay exists to guard against: replay's whole
// purpose is applying writes that can arrive out of order.
func (s *Store) PutChange(ctx context.Context, c Change) error {
	if !c.State.IsValid() {
		return fmt.Errorf("%w: %q", ErrInvalidState, c.State)
	}

	// Before the transaction opens, so a refusal cannot leave a partial
	// write behind to be rolled back. c.Repos already arrives sorted by
	// RepoRoot from internal/api's reposFromWorktrees, so a payload
	// carrying several bad values names the same one on every run without
	// this loop sorting anything itself.
	for _, r := range c.Repos {
		if r.MergeBase != nil && !mergeBasePattern.MatchString(*r.MergeBase) {
			return fmt.Errorf("%w: repo %s: %q is neither null nor a 40-character lowercase hex sha", ErrInvalidMergeBase, r.RepoRoot, *r.MergeBase)
		}
	}

	worktrees := c.Worktrees
	if worktrees == nil {
		worktrees = json.RawMessage(`{}`)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("store: begin PutChange: %w", err)
	}
	// Rollback after a successful Commit is a documented no-op in pgx; its
	// error carries nothing actionable here, so it is discarded explicitly
	// rather than checked.
	defer func() { _ = tx.Rollback(ctx) }()

	// Bootstrap the project row atomically: INSERT ... ON CONFLICT DO
	// NOTHING RETURNING tells us, in one round trip, whether *this* call
	// was the one that created the row. A separate SELECT EXISTS followed
	// by a conditional INSERT would be check-then-act — two concurrent
	// first-writers for the same brand-new project could both observe
	// "absent" and both attempt the INSERT, and the loser would hit a bare
	// unique-violation instead of a typed error. ON CONFLICT DO NOTHING
	// makes the losing writer's INSERT a silent no-op at the database
	// level, so no caller ever sees that violation.
	var createdProjectKey string
	err = tx.QueryRow(ctx, `
		INSERT INTO projects (project_key, main_checkout_path)
		VALUES ($1, $2)
		ON CONFLICT (project_key) DO NOTHING
		RETURNING project_key
	`, c.ProjectKey, c.MainCheckoutPath).Scan(&createdProjectKey)
	switch {
	case err == nil:
		// This call created the project row. Refuse it — rolling back the
		// whole transaction, including this insert — if it did so with no
		// usable checkout path.
		if c.MainCheckoutPath == "" {
			return fmt.Errorf("%w: project %q", ErrInvalidMainCheckoutPath, c.ProjectKey)
		}
	case errors.Is(err, pgx.ErrNoRows):
		// The project already existed (or another concurrent writer just
		// created it); this call's MainCheckoutPath, if any, is ignored.
	default:
		return fmt.Errorf("store: create project %s: %w", c.ProjectKey, err)
	}

	var id int64
	err = tx.QueryRow(ctx, `
		INSERT INTO changes (
			project_key, name, state, branch, worktrees, artifact_url,
			jira_issue, planning_effort, models, review_panel_roster,
			pr_url, updated_at, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		ON CONFLICT (project_key, name) DO UPDATE SET
			state               = EXCLUDED.state,
			branch              = EXCLUDED.branch,
			worktrees           = EXCLUDED.worktrees,
			artifact_url        = EXCLUDED.artifact_url,
			jira_issue          = EXCLUDED.jira_issue,
			planning_effort     = EXCLUDED.planning_effort,
			models              = EXCLUDED.models,
			review_panel_roster = EXCLUDED.review_panel_roster,
			pr_url              = EXCLUDED.pr_url,
			updated_at          = EXCLUDED.updated_at,
			updated_by          = EXCLUDED.updated_by
		WHERE state_rank(EXCLUDED.state) > state_rank(changes.state)
		   OR (state_rank(EXCLUDED.state) = state_rank(changes.state)
		       AND EXCLUDED.updated_at > changes.updated_at)
		RETURNING id
	`,
		c.ProjectKey, c.Name, string(c.State), c.Branch, worktrees, c.ArtifactURL,
		c.JiraIssue, c.PlanningEffort, c.Models, c.ReviewPanelRoster,
		c.PRURL, c.UpdatedAt, c.UpdatedBy,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMonotonicViolation
		}
		return fmt.Errorf("store: put change %s/%s: %w", c.ProjectKey, c.Name, err)
	}

	// The repository set is written on this same transaction, before
	// Commit, so a reader can never observe the change row with a
	// half-written -- or stale -- repository set: both land together or
	// neither does.
	if err := replaceChangeRepos(ctx, tx, id, c.Repos); err != nil {
		return fmt.Errorf("store: put change repos %s/%s: %w", c.ProjectKey, c.Name, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("store: commit PutChange: %w", err)
	}
	return nil
}

// GetChange returns the record for the given project and change name, or
// ErrChangeNotFound if none exists. The returned Change's Repos field is
// populated from change_repos exactly as PutChange wrote it.
func (s *Store) GetChange(ctx context.Context, projectKey, name string) (Change, error) {
	c, err := scanChange(s.pool.QueryRow(ctx, `
		SELECT project_key, name, state, branch, worktrees, artifact_url,
		       jira_issue, planning_effort, models, review_panel_roster,
		       pr_url, updated_at, updated_by
		FROM changes
		WHERE project_key = $1 AND name = $2
	`, projectKey, name))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Change{}, ErrChangeNotFound
		}
		return Change{}, fmt.Errorf("store: get change %s/%s: %w", projectKey, name, err)
	}

	repos, err := s.ListChangeRepos(ctx, projectKey, name)
	if err != nil {
		return Change{}, fmt.Errorf("store: get change repos %s/%s: %w", projectKey, name, err)
	}
	c.Repos = repos

	return c, nil
}

// ListChanges returns every change recorded for projectKey, in name order,
// with no filtering, search or paging beyond that project scope -- every
// matching row, always. It exists only because tasks 2 and 2.1's tests
// already depend on this exact signature; it is a thin wrapper around
// QueryChanges (a project filter, a name sort, and Query.NoLimit) rather
// than a second SQL path, so the two can never drift against each other.
//
// New callers should reach for QueryChanges directly -- it is the one that
// filters, searches, sorts and pages, and this wrapper's unbounded result
// is very much the exception, not the model to copy.
func (s *Store) ListChanges(ctx context.Context, projectKey string) ([]Change, error) {
	changes, _, err := s.QueryChanges(ctx, Query{
		Filters: []Filter{{Field: "project", Op: OpEq, Value: projectKey}},
		Sort:    []SortKey{{Field: "name"}},
		Limit:   NoLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("store: list changes for %s: %w", projectKey, err)
	}
	return changes, nil
}

// ProjectKeySuffixPattern is the regular expression, in POSIX/RE2 syntax
// (valid both as a Go regexp and as a PostgreSQL regexp_replace pattern),
// that matches the trailing "-" plus exactly eight lowercase hexadecimal
// characters a project key carries to disambiguate two same-named
// checkouts (State file, skills/flow-contracts/state-file.md).
//
// This package's own SQL (below) and internal/api/stats.go's
// looksLikeProjectKey both build their pattern from this one constant
// (panel round 1, F1) rather than keeping a second, silently-driftable
// copy in the same Go module -- unlike stats/web/src/lib/projectLabel.ts,
// which is a real language boundary this constant cannot reach across:
// that TypeScript copy is the one duplicate that must stay a duplicate,
// and its own doc comment names this constant's Go twin by file path for
// exactly the reason this one names it back.
const ProjectKeySuffixPattern = `-[0-9a-f]{8}$`

// projectKeysByDisplayNameQuery is built from ProjectKeySuffixPattern via
// fmt.Sprintf, not a query parameter: PostgreSQL's regexp_replace takes its
// pattern as a plain SQL argument like any other, so binding it as $2
// alongside displayName's $1 would work equally well, but building the
// query text once here keeps the single source of truth visible next to
// the query it derives, without adding a second bound parameter whose
// only caller ever passes the same fixed literal.
var projectKeysByDisplayNameQuery = fmt.Sprintf(`
	SELECT project_key
	FROM projects
	WHERE regexp_replace(project_key, '%s', '') = $1
	ORDER BY project_key
`, ProjectKeySuffixPattern)

// ProjectKeysByDisplayName returns every project key whose display name --
// project_key with ProjectKeySuffixPattern's suffix removed -- equals
// displayName. It is the server-side twin of
// stats/web/src/lib/projectLabel.ts: the same suffix, anchored at the end
// the same way, trimmed here in SQL instead of in the client, so the two
// sides agree on what a key "displays as" by construction rather than by
// keeping a comment in sync.
//
// Zero results means no project has that display name. More than one
// means the name is ambiguous -- two projects share a basename and differ
// only in the disambiguating hash -- which internal/api's resolution
// helper (stats/internal/api/stats.go) turns into a 400 naming the
// candidates, rather than picking one silently.
func (s *Store) ProjectKeysByDisplayName(ctx context.Context, displayName string) ([]string, error) {
	rows, err := s.pool.Query(ctx, projectKeysByDisplayNameQuery, displayName)
	if err != nil {
		return nil, fmt.Errorf("store: project keys by display name %q: %w", displayName, err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("store: project keys by display name %q: scan: %w", displayName, err)
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: project keys by display name %q: %w", displayName, err)
	}
	return keys, nil
}

// QueryChanges returns the page of changes matching q -- its filters,
// free-text search, sort and page -- against the fixed allowlist in
// query.go, along with the total count of matching rows (ignoring q's
// page, for a caller building pagination controls). A field named in q
// that query.go's allowlist does not recognise rejects the request with
// ErrUnknownField before any query runs, per the query-allowlist design
// decision: this is the entry point task 11 builds a Query from request
// parameters and calls.
//
// The count and the page are read inside one transaction, opened with
// queryTxOptions() -- see its doc comment for why REPEATABLE READ is
// load-bearing here (a write landing between the count and the page, even
// on the same still-open transaction, would otherwise make total disagree
// with what the page actually contains) and why the 40001 serialization
// failure that isolation level can raise cannot occur on this read-only
// transaction. Drift *across* separate calls is inherent to offset
// pagination and is not what this guards; drift *within* one is the bug
// this closes.
//
// Each returned Change's Repos field is populated inside that same
// transaction, from one extra query per distinct project represented in
// the page -- changeReposByName, exactly as the pre-QueryChanges ListChanges
// used it -- rather than one query per row, so a wide page never becomes an
// N+1.
func (s *Store) QueryChanges(ctx context.Context, q Query) ([]Change, int, error) {
	where, whereArgs, order, orderArgs, err := q.buildClauses(changeQueryable)
	if err != nil {
		return nil, 0, err
	}

	tx, err := s.pool.BeginTx(ctx, queryTxOptions())
	if err != nil {
		return nil, 0, fmt.Errorf("store: query changes: begin: %w", err)
	}
	// Rollback after a successful Commit is a documented no-op in pgx; its
	// error carries nothing actionable here, so it is discarded explicitly
	// rather than checked.
	defer func() { _ = tx.Rollback(ctx) }()

	var total int
	countSQL := "SELECT COUNT(*) FROM changes c " + where
	if err := tx.QueryRow(ctx, countSQL, whereArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("store: query changes: count: %w", err)
	}

	args := append(append([]any{}, whereArgs...), orderArgs...)
	limitClause := ""
	if !q.unlimited() {
		limit, offset := q.limit(), q.offset()
		args = append(args, limit, offset)
		limitClause = fmt.Sprintf("LIMIT $%d OFFSET $%d", len(whereArgs)+len(orderArgs)+1, len(whereArgs)+len(orderArgs)+2)
	}

	sqlText := fmt.Sprintf(`
		SELECT c.project_key, c.name, c.state, c.branch, c.worktrees, c.artifact_url,
		       c.jira_issue, c.planning_effort, c.models, c.review_panel_roster,
		       c.pr_url, c.updated_at, c.updated_by
		FROM changes c
		%s
		%s
		%s
	`, where, order, limitClause)

	rows, err := tx.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("store: query changes: %w", err)
	}
	defer rows.Close()

	var changes []Change
	for rows.Next() {
		c, err := scanChange(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("store: query changes: scan: %w", err)
		}
		changes = append(changes, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("store: query changes: %w", err)
	}

	projectKeys := make(map[string]bool, len(changes))
	for _, c := range changes {
		projectKeys[c.ProjectKey] = true
	}
	for pk := range projectKeys {
		reposByName, err := changeReposByName(ctx, tx, pk)
		if err != nil {
			return nil, 0, fmt.Errorf("store: query changes: repos for %s: %w", pk, err)
		}
		for i := range changes {
			if changes[i].ProjectKey == pk {
				changes[i].Repos = reposByName[changes[i].Name]
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, 0, fmt.Errorf("store: query changes: commit: %w", err)
	}

	return changes, total, nil
}

// rowScanner is satisfied by both pgx.Row (QueryRow) and pgx.Rows (Query),
// so scanChange serves both GetChange and ListChanges.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanChange(row rowScanner) (Change, error) {
	var (
		c         Change
		state     string
		worktrees []byte
		models    []byte
	)
	if err := row.Scan(
		&c.ProjectKey, &c.Name, &state, &c.Branch, &worktrees, &c.ArtifactURL,
		&c.JiraIssue, &c.PlanningEffort, &models, &c.ReviewPanelRoster,
		&c.PRURL, &c.UpdatedAt, &c.UpdatedBy,
	); err != nil {
		return Change{}, err
	}
	c.State = State(state)
	c.Worktrees = json.RawMessage(worktrees)
	if models != nil {
		c.Models = json.RawMessage(models)
	}
	return c, nil
}
