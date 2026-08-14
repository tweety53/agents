package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Repo is one repository affected by a change, as recorded in change_repos.
// A nil MergeBase means no merge base has been recorded for that
// repository -- never a value to infer or compute.
type Repo struct {
	RepoRoot  string
	MergeBase *string
}

// ErrDuplicateRepoRoot is returned by PutChange when c.Repos names the same
// RepoRoot more than once. It is caller error -- a change's repository set
// has exactly one entry per repository, per its primary key -- and callers
// need a typed error to distinguish it from a transport failure or an
// unrelated database error, exactly as ErrInvalidState and
// ErrMonotonicViolation are distinguished from those.
var ErrDuplicateRepoRoot = errors.New("store: duplicate repo root in payload")

// replaceChangeRepos renders the whole repository set for changeID inside
// tx: every repository in repos is written, and one already stored but
// absent from repos is removed -- the same whole-object rule PutChange
// already follows for the change's own fields.
//
// It runs on the caller's transaction rather than opening its own, so
// PutChange can commit the change row and its repository set together. A
// reader can never observe one written without the other: PutChange's own
// rollback-on-error path covers this function's writes exactly as it
// covers the change row's. In particular, a rejection here (including
// ErrDuplicateRepoRoot) rolls back the change row's own upsert, which by
// this point has already run on tx but not yet committed.
func replaceChangeRepos(ctx context.Context, tx pgx.Tx, changeID int64, repos []Repo) error {
	if err := validateRepos(repos); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM change_repos WHERE change_id = $1`, changeID); err != nil {
		return fmt.Errorf("delete change_repos for change %d: %w", changeID, err)
	}

	for _, r := range repos {
		if _, err := tx.Exec(ctx, `
			INSERT INTO change_repos (change_id, repo_root, merge_base)
			VALUES ($1, $2, $3)
		`, changeID, r.RepoRoot, r.MergeBase); err != nil {
			return fmt.Errorf("insert change_repos %s for change %d: %w", r.RepoRoot, changeID, err)
		}
	}

	return nil
}

// validateRepos refuses a repository set that names the same RepoRoot more
// than once, with ErrDuplicateRepoRoot, before any statement runs. Checking
// in Go rather than letting the second INSERT trip change_repos' composite
// primary key means the caller gets a typed, distinguishable error instead
// of a raw constraint-violation wrapped from the driver.
func validateRepos(repos []Repo) error {
	seen := make(map[string]bool, len(repos))
	for _, r := range repos {
		if seen[r.RepoRoot] {
			return fmt.Errorf("%w: %q", ErrDuplicateRepoRoot, r.RepoRoot)
		}
		seen[r.RepoRoot] = true
	}
	return nil
}

// queryExecer is satisfied by both *pgxpool.Pool and pgx.Tx, so
// changeReposByName can run against either an ambient pool connection or
// an already-open transaction. QueryChanges (query.go) passes its own
// transaction, so its repo lookup reads the same snapshot as the count and
// page it already read there, rather than a second, later one.
type queryExecer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// changeReposByName returns every repository recorded for projectKey,
// grouped by change name, in one query -- so a caller listing many changes
// can populate every one's Repos field without querying once per row.
func changeReposByName(ctx context.Context, db queryExecer, projectKey string) (map[string][]Repo, error) {
	rows, err := db.Query(ctx, `
		SELECT c.name, cr.repo_root, cr.merge_base
		FROM change_repos cr
		JOIN changes c ON c.id = cr.change_id
		WHERE c.project_key = $1
		ORDER BY c.name, cr.repo_root
	`, projectKey)
	if err != nil {
		return nil, fmt.Errorf("list change repos for project %s: %w", projectKey, err)
	}
	defer rows.Close()

	byName := make(map[string][]Repo)
	for rows.Next() {
		var (
			name string
			r    Repo
		)
		if err := rows.Scan(&name, &r.RepoRoot, &r.MergeBase); err != nil {
			return nil, fmt.Errorf("scan change repo row for project %s: %w", projectKey, err)
		}
		byName[name] = append(byName[name], r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list change repos for project %s: %w", projectKey, err)
	}
	return byName, nil
}

// ListChangeRepos returns the repository set recorded for the given
// project and change name, ordered by repo_root so repeated calls return a
// stable order. A change with no recorded repositories, or one that does
// not exist, returns an empty slice and no error -- GetChange or
// ListChanges is the place to learn whether the change itself exists.
func (s *Store) ListChangeRepos(ctx context.Context, projectKey, name string) ([]Repo, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT cr.repo_root, cr.merge_base
		FROM change_repos cr
		JOIN changes c ON c.id = cr.change_id
		WHERE c.project_key = $1 AND c.name = $2
		ORDER BY cr.repo_root
	`, projectKey, name)
	if err != nil {
		return nil, fmt.Errorf("store: list change repos for %s/%s: %w", projectKey, name, err)
	}
	defer rows.Close()

	var repos []Repo
	for rows.Next() {
		var r Repo
		if err := rows.Scan(&r.RepoRoot, &r.MergeBase); err != nil {
			return nil, fmt.Errorf("store: scan change repo row for %s/%s: %w", projectKey, name, err)
		}
		repos = append(repos, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list change repos for %s/%s: %w", projectKey, name, err)
	}
	return repos, nil
}
