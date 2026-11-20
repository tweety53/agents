package store

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5"
)

// RolePolicy is one role's pairs (kan-1103, decision pairs-in-db): Base is
// always admitted; Optin only on an `Operator pair: <role>` line plus a live
// grant (decision optin-grant). Each list
// keeps the order it was written in -- Base[0] is the role's default.
type RolePolicy struct {
	Base  []string `json:"base"`
	Optin []string `json:"optin"`
}

// ErrInvalidPair is returned by ReplaceRolePairs for a pair not shaped
// <model>/<type>[@<effort>], an empty role, or a pair named twice in a tier.
var ErrInvalidPair = errors.New("store: invalid pair")

// ErrGrantNotFound is returned by Revoke when change holds no grant for role.
var ErrGrantNotFound = errors.New("store: grant not found")

// ValidPair reports whether p is shaped <model>/<type>[@<effort>] -- the
// only check an edited pair gets (decision syntax-only-validation).
func ValidPair(p string) bool { return pairShape.MatchString(p) }

var pairShape = regexp.MustCompile(`^[a-z][a-z0-9.-]*/[A-Za-z][A-Za-z0-9._-]*(@[a-z]+)?$`)

// RolePairs returns every role's pairs, keyed by role.
func (s *Store) RolePairs(ctx context.Context) (map[string]RolePolicy, error) {
	rows, err := s.pool.Query(ctx, `SELECT role, tier, pair FROM role_pairs ORDER BY role, tier, position`)
	if err != nil {
		return nil, fmt.Errorf("store: role pairs: %w", err)
	}
	defer rows.Close()
	out := map[string]RolePolicy{}
	for rows.Next() {
		var role, tier, pair string
		if err := rows.Scan(&role, &tier, &pair); err != nil {
			return nil, fmt.Errorf("store: role pairs: %w", err)
		}
		p := out[role]
		if tier == "base" {
			p.Base = append(p.Base, pair)
		} else {
			p.Optin = append(p.Optin, pair)
		}
		out[role] = p
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: role pairs: %w", err)
	}
	for role, p := range out {
		out[role] = p.nonNil()
	}
	return out, nil
}

func (p RolePolicy) nonNil() RolePolicy {
	if p.Base == nil {
		p.Base = []string{}
	}
	if p.Optin == nil {
		p.Optin = []string{}
	}
	return p
}

// ReplaceRolePairs replaces role's base and opt-in pairs whole, in one
// transaction. A malformed pair, or one named twice in a tier, refuses the
// whole write with ErrInvalidPair.
func (s *Store) ReplaceRolePairs(ctx context.Context, role string, base, optin []string) error {
	if role == "" {
		return fmt.Errorf("%w: empty role", ErrInvalidPair)
	}
	for tier, pairs := range map[string][]string{"base": base, "optin": optin} {
		seen := map[string]bool{}
		for _, p := range pairs {
			if !ValidPair(p) {
				return fmt.Errorf("%w: role %q %s pair %q is not <model>/<type>[@<effort>]", ErrInvalidPair, role, tier, p)
			}
			if seen[p] {
				return fmt.Errorf("%w: role %q %s pair %q named twice", ErrInvalidPair, role, tier, p)
			}
			seen[p] = true
		}
	}
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		// Serialize replaces of one role so two concurrent writes never merge.
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('role_pairs:' || $1))`, role); err != nil {
			return fmt.Errorf("store: replace role pairs %q: %w", role, err)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM role_pairs WHERE role = $1`, role); err != nil {
			return fmt.Errorf("store: replace role pairs %q: %w", role, err)
		}
		for tier, pairs := range map[string][]string{"base": base, "optin": optin} {
			for i, p := range pairs {
				if _, err := tx.Exec(ctx, `
					INSERT INTO role_pairs (role, tier, pair, position) VALUES ($1, $2, $3, $4)
				`, role, tier, p, i); err != nil {
					return fmt.Errorf("store: replace role pairs %q: %w", role, err)
				}
			}
		}
		return nil
	})
}

// Grant records the operator's say-so that change's dispatches of each role
// may run that role's opt-in pairs. Granting twice is a no-op.
func (s *Store) Grant(ctx context.Context, change string, roles []string) error {
	if change == "" || len(roles) == 0 {
		return fmt.Errorf("%w: a grant needs a change and at least one role", ErrInvalidPair)
	}
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		for _, role := range roles {
			if role == "" {
				return fmt.Errorf("%w: empty role", ErrInvalidPair)
			}
			if _, err := tx.Exec(ctx, `INSERT INTO policy_grants (change, role) VALUES ($1, $2) ON CONFLICT DO NOTHING`, change, role); err != nil {
				return fmt.Errorf("store: grant %q %q: %w", change, role, err)
			}
		}
		return nil
	})
}

// Revoke removes change's grants for roles; every role must hold one, or
// nothing is removed and the error wraps ErrGrantNotFound.
func (s *Store) Revoke(ctx context.Context, change string, roles []string) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		for _, role := range roles {
			tag, err := tx.Exec(ctx, `DELETE FROM policy_grants WHERE change = $1 AND role = $2`, change, role)
			if err != nil {
				return fmt.Errorf("store: revoke %q %q: %w", change, role, err)
			}
			if tag.RowsAffected() == 0 {
				return fmt.Errorf("%w: change %q holds no grant for role %q", ErrGrantNotFound, change, role)
			}
		}
		return nil
	})
}

// Grants returns the roles change holds a live grant for, sorted; a change
// recorded FINISHED holds none -- its grants lapse with it.
func (s *Store) Grants(ctx context.Context, change string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT g.role FROM policy_grants g
		WHERE g.change = $1
		  AND NOT EXISTS (SELECT 1 FROM changes c WHERE c.name = g.change AND c.state = 'FINISHED')
		ORDER BY g.role
	`, change)
	if err != nil {
		return nil, fmt.Errorf("store: grants of %q: %w", change, err)
	}
	out, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("store: grants of %q: %w", change, err)
	}
	return out, nil
}
