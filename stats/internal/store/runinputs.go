package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/tweety53/agents/stats/internal/records"
)

// ChangeRef names one change.
type ChangeRef struct {
	ProjectKey, Name string
}

// RunInputs is one change's stored rows the per-run computation reads.
type RunInputs struct {
	Dispatches []records.Dispatch
	Findings   []records.Finding
	Stages     []StageRun
	VerifyRuns []records.VerifyRun
	// Files is the change's saved tasks.md and reports.* files, by path.
	Files map[string]string
}

// RunInputs reads every ref's RunInputs in one transaction of six queries,
// whatever the number of refs, in refs' order. A ref naming no change is
// ErrChangeNotFound.
func (s *Store) RunInputs(ctx context.Context, refs []ChangeRef) ([]RunInputs, error) {
	tx, err := s.pool.BeginTx(ctx, queryTxOptions())
	if err != nil {
		return nil, fmt.Errorf("store: run inputs: begin: %w", err)
	}
	// Rollback after a successful Commit is a documented no-op in pgx; its
	// error carries nothing actionable here, so it is discarded explicitly
	// rather than checked.
	defer func() { _ = tx.Rollback(ctx) }()

	ids, err := changeIDsOf(ctx, tx, refs)
	if err != nil {
		return nil, err
	}
	dispatches, err := readDispatchesOf(ctx, tx, ids)
	if err != nil {
		return nil, fmt.Errorf("store: run inputs: %w", err)
	}
	findings, err := readFindingsOf(ctx, tx, ids)
	if err != nil {
		return nil, fmt.Errorf("store: run inputs: %w", err)
	}
	stages, err := readStageRunsOf(ctx, tx, ids)
	if err != nil {
		return nil, fmt.Errorf("store: run inputs: %w", err)
	}
	verify, err := readVerifyRunsOf(ctx, tx, ids)
	if err != nil {
		return nil, fmt.Errorf("store: run inputs: %w", err)
	}
	files, err := readRunFilesOf(ctx, tx, ids)
	if err != nil {
		return nil, fmt.Errorf("store: run inputs: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("store: run inputs: commit: %w", err)
	}
	out := make([]RunInputs, len(ids))
	for i, id := range ids {
		out[i] = RunInputs{Dispatches: dispatches[id], Findings: findings[id], Stages: stages[id], VerifyRuns: verify[id], Files: files[id]}
	}
	return out, nil
}

// changeIDsOf resolves refs to change ids, in refs' order.
func changeIDsOf(ctx context.Context, tx pgx.Tx, refs []ChangeRef) ([]int64, error) {
	keys, names := make([]string, len(refs)), make([]string, len(refs))
	for i, r := range refs {
		keys[i], names[i] = r.ProjectKey, r.Name
	}
	rows, err := tx.Query(ctx, `
		SELECT r.i, c.id
		FROM unnest($1::text[], $2::text[]) WITH ORDINALITY AS r(pk, name, i)
		JOIN changes c ON c.project_key = r.pk AND c.name = r.name
	`, keys, names)
	if err != nil {
		return nil, fmt.Errorf("store: run inputs: changes: %w", err)
	}
	defer rows.Close()
	ids := make([]int64, len(refs))
	for rows.Next() {
		var i, id int64
		if err := rows.Scan(&i, &id); err != nil {
			return nil, fmt.Errorf("store: run inputs: changes: scan: %w", err)
		}
		ids[i-1] = id
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: run inputs: changes: %w", err)
	}
	for i, id := range ids {
		if id == 0 {
			return nil, fmt.Errorf("%w: %s/%s", ErrChangeNotFound, refs[i].ProjectKey, refs[i].Name)
		}
	}
	return ids, nil
}

func readStageRunsOf(ctx context.Context, tx pgx.Tx, ids []int64) (map[int64][]StageRun, error) {
	rows, err := tx.Query(ctx, `
		SELECT `+stageRunSelect+`
		FROM stage_runs sr
		LEFT JOIN changes c ON c.id = sr.change_id
		WHERE sr.change_id = ANY($1)
		ORDER BY sr.change_id, sr.id
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("read stage runs: %w", err)
	}
	defer rows.Close()
	runs, err := scanStageRuns(rows)
	if err != nil {
		return nil, fmt.Errorf("read stage runs: %w", err)
	}
	out := map[int64][]StageRun{}
	for _, r := range runs {
		out[r.ChangeID] = append(out[r.ChangeID], r)
	}
	return out, nil
}

func readVerifyRunsOf(ctx context.Context, tx pgx.Tx, ids []int64) (map[int64][]records.VerifyRun, error) {
	rows, err := tx.Query(ctx, `
		SELECT change_id, `+verifyRunColumns+`
		FROM verify_runs
		WHERE change_id = ANY($1)
		ORDER BY change_id, recorded_at, id
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("read verify runs: %w", err)
	}
	defer rows.Close()
	out := map[int64][]records.VerifyRun{}
	for rows.Next() {
		var id int64
		v, err := scanVerifyRun(prefixScanner{rows, &id})
		if err != nil {
			return nil, fmt.Errorf("read verify runs: scan: %w", err)
		}
		out[id] = append(out[id], v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read verify runs: %w", err)
	}
	return out, nil
}

// readRunFilesOf reads the saved files a run reads: tasks.md and the
// reports.* copies.
func readRunFilesOf(ctx context.Context, tx pgx.Tx, ids []int64) (map[int64]map[string]string, error) {
	rows, err := tx.Query(ctx, `
		SELECT change_id, path, body FROM change_files
		WHERE change_id = ANY($1) AND (path = 'tasks.md' OR path LIKE 'reports.%')
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("read change files: %w", err)
	}
	defer rows.Close()
	out := map[int64]map[string]string{}
	for rows.Next() {
		var (
			id         int64
			path, body string
		)
		if err := rows.Scan(&id, &path, &body); err != nil {
			return nil, fmt.Errorf("read change files: scan: %w", err)
		}
		if out[id] == nil {
			out[id] = map[string]string{}
		}
		out[id][path] = body
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read change files: %w", err)
	}
	return out, nil
}
