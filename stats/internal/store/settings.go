package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// ValidModels is the set of models the Decide step may assign a dispatch
// pair: every model field of a recorded decision must be one of these,
// which internal/api's ApplyDecisionRecord enforces at record time
// (design.md's valid-models-opus-sonnet decision).
var ValidModels = map[string]bool{
	"opus":   true,
	"sonnet": true,
}

// ValidReviewers is the fixed vocabulary a flow_settings.reviewers entry
// may take. The panel dispatches exactly the resolved list; these ids
// no longer split into a required subset and an on-demand-only subset
// (design.md's roster-from-settings decision superseded that split).
// "simple-reviewer" was retired once primary absorbed its brief, and
// "code-review-low" once it proved a shallower copy of primary's
// code-review angle; "bugbot" and "security" once the flow store showed
// security raising no finding and bugbot the lowest yield of any slot;
// "failure-modes" was promoted from the experimental exp-failure-modes
// prompt. The store rejects every retired id like any other unknown id.
// Only writes are validated: the record/aggregate paths read historical
// rows carrying a retired id unchanged, and a stored flow_settings list
// had its retired ids stripped once by
// 0033_flow_settings_drop_retired_reviewers.sql.
var ValidReviewers = map[string]bool{
	"primary":       true,
	"principles":    true,
	"failure-modes": true,
	"mutation":      true,
}

// ErrInvalidReviewer is returned by PutSettings when a Reviewers entry is
// not one of ValidReviewers.
var ErrInvalidReviewer = errors.New("store: invalid reviewer")

// DefaultReviewers is the value GetSettings reports for Reviewers when
// flow_settings holds no row yet: primary and principles,
// the same two ids skills/flow/SKILL.md's resolver falls back to when the
// store is unreachable (design.md's unreachable-falls-back-to-defaults
// decision).
var DefaultReviewers = []string{"primary", "principles"}

// Settings is the harness-wide record /flow-settings manages: which
// reviewer slots the panel dispatches by default. flow_settings
// (0015_flow_settings.sql, 0034_flow_settings_drop_model_columns.sql)
// holds exactly one row of this shape.
type Settings struct {
	Reviewers []string
}

// ValidateSettings reports whether every entry of s.Reviewers falls within
// ValidReviewers, returning ErrInvalidReviewer -- wrapped with the specific
// bad value -- for the first violation found. internal/api's own
// ValidateSettings re-exports this rather than redefining the vocabulary a
// second time, so the store and the HTTP layer can never silently diverge
// on what counts as a valid value.
func ValidateSettings(s Settings) error {
	for _, r := range s.Reviewers {
		if !ValidReviewers[r] {
			return fmt.Errorf("%w: %q", ErrInvalidReviewer, r)
		}
	}
	return nil
}

// PutSettings validates settings and upserts it as flow_settings' single
// row, overwriting whatever was recorded before. It returns
// ErrInvalidReviewer -- without writing anything -- if
// settings fails ValidateSettings.
func (s *Store) PutSettings(ctx context.Context, settings Settings) error {
	if err := ValidateSettings(settings); err != nil {
		return err
	}

	reviewers, err := json.Marshal(settings.Reviewers)
	if err != nil {
		return fmt.Errorf("store: put settings: marshal reviewers: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO flow_settings (id, reviewers)
		VALUES (TRUE, $1)
		ON CONFLICT (id) DO UPDATE SET
			reviewers  = EXCLUDED.reviewers,
			updated_at = now()
	`, json.RawMessage(reviewers))
	if err != nil {
		return fmt.Errorf("store: put settings: %w", err)
	}
	return nil
}

// GetSettings returns flow_settings' single row, or -- when no row has
// ever been written -- DefaultReviewers: a /flow run started before
// /flow-settings has ever been invoked still resolves to a defined roster
// rather than an error.
func (s *Store) GetSettings(ctx context.Context) (Settings, error) {
	var reviewers []byte
	err := s.pool.QueryRow(ctx, `
		SELECT reviewers FROM flow_settings WHERE id = TRUE
	`).Scan(&reviewers)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Settings{Reviewers: append([]string(nil), DefaultReviewers...)}, nil
		}
		return Settings{}, fmt.Errorf("store: get settings: %w", err)
	}

	var out Settings
	if err := json.Unmarshal(reviewers, &out.Reviewers); err != nil {
		return Settings{}, fmt.Errorf("store: get settings: decode reviewers: %w", err)
	}
	return out, nil
}
