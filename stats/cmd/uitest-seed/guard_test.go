package main

import (
	"strings"
	"testing"
)

// TestGuardRefusesNonUitestDSN is the guard's own point: every DSN whose
// database name does not *end in* "_uitest" is refused, with a reason,
// before any statement would be issued. The live DSN in URL form is the
// failure this task exists to close off; the same DSN in keyword form
// ("host=... dbname=myflow") is included too, since pgxpool.ParseConfig
// accepts both forms and the guard must refuse the target regardless of
// which syntax named it. The "myflow_uitest_real" case is deliberately
// included even though the database name contains "uitest" -- the
// requirement is a suffix match, not a substring match, so a database
// named "..._real" is exactly as unprotected as "myflow" itself and must
// refuse too.
func TestGuardRefusesNonUitestDSN(t *testing.T) {
	cases := []struct {
		name string
		dsn  string
	}{
		{
			name: "the live DSN",
			dsn:  "postgres://myflow:myflow@localhost:5433/myflow?sslmode=disable",
		},
		{
			name: "the live DSN in keyword form",
			dsn:  "host=localhost port=5433 user=myflow password=myflow dbname=myflow sslmode=disable",
		},
		{
			name: "a DSN whose database name merely contains uitest in the middle",
			dsn:  "postgres://myflow:myflow@localhost:5433/myflow_uitest_real?sslmode=disable",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := requireUitestDatabase(tc.dsn)
			if err == nil {
				t.Fatalf("requireUitestDatabase(%q): expected refusal, got nil", tc.dsn)
			}
			if !strings.Contains(err.Error(), "_uitest") {
				t.Fatalf("requireUitestDatabase(%q): error %q does not name the reason", tc.dsn, err)
			}
		})
	}
}

// TestGuardAcceptsUitestDSN is the guard's other half: a DSN whose
// database name ends in "_uitest" is admitted.
func TestGuardAcceptsUitestDSN(t *testing.T) {
	if err := requireUitestDatabase("postgres://myflow:myflow@localhost:5433/myflow_uitest?sslmode=disable"); err != nil {
		t.Fatalf("requireUitestDatabase: expected acceptance, got %v", err)
	}
}

// TestGuardRefusesUnparsableDSN makes sure a malformed target is refused
// rather than silently treated as unnamed and let through.
func TestGuardRefusesUnparsableDSN(t *testing.T) {
	if err := requireUitestDatabase("not a dsn at all"); err == nil {
		t.Fatalf("requireUitestDatabase: expected refusal for an unparsable DSN, got nil")
	}
}
