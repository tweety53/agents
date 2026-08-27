// Command uitest-seed populates the UI-test stack's database with a fixed
// fixture, and refuses to run at all against anything else.
package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// uitestSuffix is the one rule every destructive test-stack path checks
// its target against, per specs/myflow-ui-test-stack/spec.md's
// "Destructive test-stack paths refuse to act on any other database": the
// target's database name must end in this suffix, not merely contain it.
const uitestSuffix = "_uitest"

// requireUitestDatabase parses the database name out of dsn and refuses,
// naming the reason, unless it ends in uitestSuffix. It runs before any
// statement is issued against dsn's target -- a guard that runs after the
// first write is not a guard.
//
// A DSN naming "flow" is refused, and so is one naming
// "flow_uitest_real": the rule is a suffix match, not a substring
// match, so a database name that merely contains "uitest" in the middle
// is exactly as unprotected as the live database itself.
func requireUitestDatabase(dsn string) error {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("uitest-seed: parse dsn: %w", err)
	}

	dbName := cfg.ConnConfig.Database
	if !strings.HasSuffix(dbName, uitestSuffix) {
		return fmt.Errorf("%w: database %q does not end in %q", errRefusedTarget, dbName, uitestSuffix)
	}
	return nil
}

// errRefusedTarget is returned by requireUitestDatabase when dsn's target
// is not the UI-test database. Named so a caller (or a test) can
// distinguish "refused, deliberately" from a transport or parse failure.
var errRefusedTarget = errors.New("uitest-seed: refused")
