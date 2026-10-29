package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
)

// Experiment is one allowed-pairs experiment (KAN-1067): Cells maps a role
// id to the full list of pairs replacing that role's unmarked row of
// skills/flow-contracts/model-policy.md's Allowed pairs table while the
// experiment is open. Role "main" names the parent agent's one
// <model>@<effort>. Status is derived from ClosedAt: "open" or "closed".
type Experiment struct {
	ID        int64               `json:"id"`
	Name      string              `json:"name"`
	Cells     map[string][]string `json:"cells"`
	StartedAt time.Time           `json:"startedAt"`
	ClosedAt  *time.Time          `json:"closedAt,omitempty"`
	Verdict   string              `json:"verdict,omitempty"`
	Summary   string              `json:"summary"`
	Status    string              `json:"status"`
}

// ErrInvalidExperiment is returned when an experiment fails
// ValidateExperiment.
var ErrInvalidExperiment = errors.New("store: invalid experiment")

// ErrExperimentConflict is returned by OpenExperiment when one of its roles
// is already held by an open experiment; the wrapped message names the role
// and the holder's id.
var ErrExperimentConflict = errors.New("store: experiment conflict")

// ErrExperimentNotOpen is returned by CloseExperiment for an unknown or
// already-closed experiment id.
var ErrExperimentNotOpen = errors.New("store: experiment not open")

// ErrInvalidVerdict is returned by CloseExperiment for a verdict outside
// ValidVerdicts.
var ErrInvalidVerdict = errors.New("store: invalid verdict")

// ValidVerdicts is the closing verdict vocabulary, matching the experiments
// table's CHECK constraint (0054_experiments.sql).
var ValidVerdicts = map[string]bool{"positive": true, "negative": true, "inconclusive": true}

var (
	experimentPair     = regexp.MustCompile(`^[a-z][a-z0-9.-]*/[A-Za-z][A-Za-z0-9._-]*(@[a-z]+)?$`)
	experimentMainPair = regexp.MustCompile(`^[a-z][a-z0-9.-]*@[a-z]+$`)
)

// ValidateExperiment reports, wrapping ErrInvalidExperiment, the first way e
// is malformed: an empty name, no cells, an empty role, a cell with no
// pairs, a pair not shaped <model>/<type>[@<effort>], or a "main" cell not
// holding exactly one <model>@<effort>.
func ValidateExperiment(e Experiment) error {
	if e.Name == "" {
		return fmt.Errorf("%w: empty name", ErrInvalidExperiment)
	}
	if len(e.Cells) == 0 {
		return fmt.Errorf("%w: no cells", ErrInvalidExperiment)
	}
	for _, role := range sortedRoles(e.Cells) {
		pairs := e.Cells[role]
		switch {
		case role == "":
			return fmt.Errorf("%w: empty role", ErrInvalidExperiment)
		case len(pairs) == 0:
			return fmt.Errorf("%w: role %q has no pairs", ErrInvalidExperiment, role)
		case role == "main":
			if len(pairs) != 1 || !experimentMainPair.MatchString(pairs[0]) {
				return fmt.Errorf("%w: role main takes exactly one <model>@<effort>, got %q", ErrInvalidExperiment, pairs)
			}
		default:
			for _, p := range pairs {
				if !experimentPair.MatchString(p) {
					return fmt.Errorf("%w: role %q pair %q is not <model>/<type>[@<effort>]", ErrInvalidExperiment, role, p)
				}
			}
		}
	}
	return nil
}

func sortedRoles(cells map[string][]string) []string {
	roles := make([]string, 0, len(cells))
	for r := range cells {
		roles = append(roles, r)
	}
	sort.Strings(roles)
	return roles
}

// OpenExperiment validates e and, in one transaction, inserts it and claims
// each of its roles in experiment_roles. A role an open experiment already
// holds refuses the whole open with ErrExperimentConflict naming the role
// and the holder's id. It returns e with ID, StartedAt and Status set.
func (s *Store) OpenExperiment(ctx context.Context, e Experiment) (Experiment, error) {
	if err := ValidateExperiment(e); err != nil {
		return Experiment{}, err
	}
	cells, err := json.Marshal(e.Cells)
	if err != nil {
		return Experiment{}, fmt.Errorf("store: open experiment: marshal cells: %w", err)
	}
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `
			INSERT INTO experiments (name, cells) VALUES ($1, $2)
			RETURNING id, started_at
		`, e.Name, json.RawMessage(cells)).Scan(&e.ID, &e.StartedAt); err != nil {
			return fmt.Errorf("store: open experiment: %w", err)
		}
		// Sorted, so two concurrent opens lock their shared roles in one order.
		for _, role := range sortedRoles(e.Cells) {
			tag, err := tx.Exec(ctx, `
				INSERT INTO experiment_roles (role, experiment_id) VALUES ($1, $2)
				ON CONFLICT (role) DO NOTHING
			`, role, e.ID)
			if err != nil {
				return fmt.Errorf("store: open experiment: claim role %q: %w", role, err)
			}
			if tag.RowsAffected() == 0 {
				var holder int64
				if err := tx.QueryRow(ctx, `SELECT experiment_id FROM experiment_roles WHERE role = $1`, role).Scan(&holder); err != nil {
					return fmt.Errorf("%w: role %q is held by an open experiment", ErrExperimentConflict, role)
				}
				return fmt.Errorf("%w: role %q is held by open experiment %d", ErrExperimentConflict, role, holder)
			}
		}
		return nil
	})
	if err != nil {
		return Experiment{}, err
	}
	e.Status = "open"
	return e, nil
}

// ListExperiments returns experiments newest first -- only the open ones
// when openOnly is set.
func (s *Store) ListExperiments(ctx context.Context, openOnly bool) ([]Experiment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, cells, started_at, closed_at, COALESCE(verdict, ''), summary
		FROM experiments
		WHERE NOT $1 OR closed_at IS NULL
		ORDER BY started_at DESC, id DESC
	`, openOnly)
	if err != nil {
		return nil, fmt.Errorf("store: list experiments: %w", err)
	}
	defer rows.Close()
	out := []Experiment{}
	for rows.Next() {
		var e Experiment
		var cells []byte
		if err := rows.Scan(&e.ID, &e.Name, &cells, &e.StartedAt, &e.ClosedAt, &e.Verdict, &e.Summary); err != nil {
			return nil, fmt.Errorf("store: list experiments: %w", err)
		}
		if err := json.Unmarshal(cells, &e.Cells); err != nil {
			return nil, fmt.Errorf("store: list experiments: decode cells of %d: %w", e.ID, err)
		}
		e.Status = "open"
		if e.ClosedAt != nil {
			e.Status = "closed"
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list experiments: %w", err)
	}
	return out, nil
}

// OpenOverrides returns the cells of every open experiment merged into one
// role -> pairs map; experiment_roles guarantees no two open experiments
// share a role.
func (s *Store) OpenOverrides(ctx context.Context) (map[string][]string, error) {
	open, err := s.ListExperiments(ctx, true)
	if err != nil {
		return nil, err
	}
	out := map[string][]string{}
	for _, e := range open {
		for role, pairs := range e.Cells {
			out[role] = pairs
		}
	}
	return out, nil
}

// CloseExperiment, in one transaction, closes open experiment id with
// verdict and summary and releases its roles. An unknown or already-closed
// id returns ErrExperimentNotOpen; a verdict outside ValidVerdicts returns
// ErrInvalidVerdict without writing.
func (s *Store) CloseExperiment(ctx context.Context, id int64, verdict, summary string) error {
	if !ValidVerdicts[verdict] {
		return fmt.Errorf("%w: %q (want positive, negative or inconclusive)", ErrInvalidVerdict, verdict)
	}
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE experiments SET closed_at = now(), verdict = $2, summary = $3
			WHERE id = $1 AND closed_at IS NULL
		`, id, verdict, summary)
		if err != nil {
			return fmt.Errorf("store: close experiment %d: %w", id, err)
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("%w: %d", ErrExperimentNotOpen, id)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM experiment_roles WHERE experiment_id = $1`, id); err != nil {
			return fmt.Errorf("store: close experiment %d: release roles: %w", id, err)
		}
		return nil
	})
}
