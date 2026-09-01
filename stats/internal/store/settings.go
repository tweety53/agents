package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// ValidModels is the harness's fixed set of known model identifiers a
// flow_settings.default_model value may take (design.md's
// model-default-sonnet decision: implementer, fixer and reviewer all
// default to one of these, uniformly). "fable" is the literal
// skills/flow/SKILL.md's planning-model resolver falls back to -- it must
// be a legal value for every model field, not just planning_model, or the
// store would refuse the very default this field introduces.
var ValidModels = map[string]bool{
	"sonnet": true,
	"opus":   true,
	"haiku":  true,
	"fable":  true,
}

// ValidReviewers is the fixed vocabulary a flow_settings.reviewers entry
// may take. The panel dispatches exactly the resolved list; these five ids
// no longer split into a required subset and an on-demand-only subset
// (design.md's roster-from-settings decision superseded that split).
var ValidReviewers = map[string]bool{
	"primary":         true,
	"principles":      true,
	"code-review-low": true,
	"bugbot":          true,
	"security":        true,
}

// ErrInvalidModel is returned by PutSettings when DefaultModel is not one
// of ValidModels.
var ErrInvalidModel = errors.New("store: invalid model")

// ErrInvalidReviewer is returned by PutSettings when a Reviewers entry is
// not one of ValidReviewers.
var ErrInvalidReviewer = errors.New("store: invalid reviewer")

// DefaultModel is the value GetSettings reports for DefaultModel when
// flow_settings holds no row yet -- the model-default-sonnet decision's
// default, so a /flow run started before /flow-settings has ever been run
// still resolves to a defined default rather than an error.
const DefaultModel = "sonnet"

// DefaultReviewers is the value GetSettings reports for Reviewers when
// flow_settings holds no row yet: primary, principles and code-review-low,
// the same three ids skills/flow/SKILL.md's resolver falls back to when the
// store is unreachable (design.md's unreachable-falls-back-to-defaults
// decision).
var DefaultReviewers = []string{"primary", "principles", "code-review-low"}

// Settings is the harness-wide record /flow-settings manages: which model
// implements, fixes and reviews by default, which model self-review runs
// on, which model planning runs on, and which reviewer slots the panel
// dispatches by default. flow_settings (0015_flow_settings.sql,
// 0016_flow_settings_self_review_model.sql,
// 0017_flow_settings_planning_model.sql) holds exactly one row of this
// shape.
type Settings struct {
	DefaultModel string
	// SelfReviewModel is the model /flow's archive-phase self-review
	// reasoning pass runs on. Unlike DefaultModel, empty is a valid value
	// here -- it means "inherit DefaultModel", not "unset".
	SelfReviewModel string
	// PlanningModel is the model /flow's planning stages run on. Empty is
	// a valid value here too, but -- unlike SelfReviewModel -- it does not
	// mean "inherit DefaultModel": it means the literal "fable",
	// skills/flow/SKILL.md's own fallback. Resolving that fallback is the
	// skill's job, not GetSettings'; the wire shape stays a record of intent.
	PlanningModel string
	Reviewers     []string
}

// ValidateSettings reports whether s.DefaultModel, s.SelfReviewModel (when
// non-empty), s.PlanningModel (when non-empty) and every entry of
// s.Reviewers fall within the harness's fixed enums, returning
// ErrInvalidModel or ErrInvalidReviewer -- wrapped with the specific bad
// value -- for the first violation found. internal/api's own
// ValidateSettings re-exports this rather than redefining the enums a
// second time, so the store and the HTTP layer (task 2) can never
// silently diverge on what counts as a valid value.
func ValidateSettings(s Settings) error {
	if !ValidModels[s.DefaultModel] {
		return fmt.Errorf("%w: %q", ErrInvalidModel, s.DefaultModel)
	}
	if s.SelfReviewModel != "" && !ValidModels[s.SelfReviewModel] {
		return fmt.Errorf("%w: %q", ErrInvalidModel, s.SelfReviewModel)
	}
	if s.PlanningModel != "" && !ValidModels[s.PlanningModel] {
		return fmt.Errorf("%w: %q", ErrInvalidModel, s.PlanningModel)
	}
	for _, r := range s.Reviewers {
		if !ValidReviewers[r] {
			return fmt.Errorf("%w: %q", ErrInvalidReviewer, r)
		}
	}
	return nil
}

// PutSettings validates settings and upserts it as flow_settings' single
// row, overwriting whatever was recorded before. It returns
// ErrInvalidModel or ErrInvalidReviewer -- without writing anything -- if
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
		INSERT INTO flow_settings (id, default_model, self_review_model, planning_model, reviewers)
		VALUES (TRUE, $1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			default_model      = EXCLUDED.default_model,
			self_review_model  = EXCLUDED.self_review_model,
			planning_model     = EXCLUDED.planning_model,
			reviewers          = EXCLUDED.reviewers,
			updated_at         = now()
	`, settings.DefaultModel, settings.SelfReviewModel, settings.PlanningModel, json.RawMessage(reviewers))
	if err != nil {
		return fmt.Errorf("store: put settings: %w", err)
	}
	return nil
}

// GetSettings returns flow_settings' single row, or -- when no row has
// ever been written -- DefaultModel and DefaultReviewers: a /flow run
// started before /flow-settings has ever been invoked still resolves to a
// defined default rather than an error (this package's own DefaultModel
// and DefaultReviewers doc comments).
func (s *Store) GetSettings(ctx context.Context) (Settings, error) {
	var defaultModel, selfReviewModel, planningModel string
	var reviewers []byte
	err := s.pool.QueryRow(ctx, `
		SELECT default_model, self_review_model, planning_model, reviewers FROM flow_settings WHERE id = TRUE
	`).Scan(&defaultModel, &selfReviewModel, &planningModel, &reviewers)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Settings{
				DefaultModel: DefaultModel,
				Reviewers:    append([]string(nil), DefaultReviewers...),
			}, nil
		}
		return Settings{}, fmt.Errorf("store: get settings: %w", err)
	}

	out := Settings{DefaultModel: defaultModel, SelfReviewModel: selfReviewModel, PlanningModel: planningModel}
	if err := json.Unmarshal(reviewers, &out.Reviewers); err != nil {
		return Settings{}, fmt.Errorf("store: get settings: decode reviewers: %w", err)
	}
	return out, nil
}
