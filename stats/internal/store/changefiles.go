package store

// changefiles.go -- the change_files table (0059_change_files.sql): the
// files of a change kept after its worktree is gone (KAN-1073), one row per
// (change, path), written by `flow change save` and read by
// `flow change file`.

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// ErrChangeFileNotFound is a known change with no saved file at the path.
var ErrChangeFileNotFound = errors.New("store: change file not found")

// changeID resolves a change's id, ErrChangeNotFound when the store has no
// such change.
func (s *Store) changeID(ctx context.Context, projectKey, change string) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `SELECT id FROM changes WHERE project_key = $1 AND name = $2`, projectKey, change).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("%w: %s/%s", ErrChangeNotFound, projectKey, change)
	}
	if err != nil {
		return 0, fmt.Errorf("store: resolve change %s/%s: %w", projectKey, change, err)
	}
	return id, nil
}

// SaveChangeFile upserts one file of a change: a second save of the same
// path replaces its body and restamps saved_at.
func (s *Store) SaveChangeFile(ctx context.Context, projectKey, change, path, body string) error {
	id, err := s.changeID(ctx, projectKey, change)
	if err != nil {
		return err
	}
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO change_files (change_id, path, body) VALUES ($1, $2, $3)
		ON CONFLICT (change_id, path) DO UPDATE SET body = EXCLUDED.body, saved_at = now()
	`, id, path, body); err != nil {
		return fmt.Errorf("store: save change file %s/%s %s: %w", projectKey, change, path, err)
	}
	return nil
}

// GetChangeFile reads one saved file of a change: ErrChangeNotFound for an
// unknown change, ErrChangeFileNotFound for a path never saved.
func (s *Store) GetChangeFile(ctx context.Context, projectKey, change, path string) (string, error) {
	id, err := s.changeID(ctx, projectKey, change)
	if err != nil {
		return "", err
	}
	var body string
	err = s.pool.QueryRow(ctx, `SELECT body FROM change_files WHERE change_id = $1 AND path = $2`, id, path).Scan(&body)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("%w: %s/%s %s", ErrChangeFileNotFound, projectKey, change, path)
	}
	if err != nil {
		return "", fmt.Errorf("store: read change file %s/%s %s: %w", projectKey, change, path, err)
	}
	return body, nil
}

// HasChangeFiles reports whether any file of the change is saved -- the
// worktree-removal gate's read.
func (s *Store) HasChangeFiles(ctx context.Context, projectKey, change string) (bool, error) {
	id, err := s.changeID(ctx, projectKey, change)
	if err != nil {
		return false, err
	}
	var has bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM change_files WHERE change_id = $1)`, id).Scan(&has); err != nil {
		return false, fmt.Errorf("store: list change files %s/%s: %w", projectKey, change, err)
	}
	return has, nil
}
