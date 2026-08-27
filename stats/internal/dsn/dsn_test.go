package dsn

import (
	"errors"
	"strings"
	"testing"
)

// TestForDatabaseHandlesTheShapesStringSurgeryCorrupted pins the four inputs the
// hand-rolled predecessor silently got wrong. Each case is a real DSN shape an
// operator overriding FLOW_STATS_ADMIN_DSN can plausibly have, and each one
// produced a corrupted connection string with no error and no panic before this
// package existed -- the review panel reproduced all four.
func TestForDatabaseHandlesTheShapesStringSurgeryCorrupted(t *testing.T) {
	cases := []struct {
		name string
		base string
		want string
	}{
		{
			// LastIndex("/") landed inside the query value, rewriting the search
			// path and leaving the database untouched.
			name: "query value containing a slash",
			base: "postgres://flow:flow@localhost:5433/flow?options=-c%20search_path=a/b",
			want: "postgres://flow:flow@localhost:5433/scratch?options=-c%20search_path=a/b",
		},
		{
			// The rewrite landed inside the host parameter of a socket DSN.
			name: "socket host parameter",
			base: "postgres:///flow?host=/var/run/postgresql",
			want: "postgres:///scratch?host=/var/run/postgresql",
		},
		{
			// No slash anywhere, so the predecessor returned the base unchanged
			// and every test then connected to the WRONG database.
			name: "libpq keyword form",
			base: "host=localhost port=5433 dbname=flow sslmode=disable",
			want: "host=localhost port=5433 dbname=scratch sslmode=disable",
		},
		{
			name: "ordinary URL",
			base: "postgres://flow:flow@localhost:5433/flow?sslmode=disable",
			want: "postgres://flow:flow@localhost:5433/scratch?sslmode=disable",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ForDatabase(c.base, "scratch")
			if err != nil {
				t.Fatalf("ForDatabase(%q) returned error: %v", c.base, err)
			}
			if got != c.want {
				t.Errorf("ForDatabase(%q)\n got %q\nwant %q", c.base, got, c.want)
			}
		})
	}
}

// TestForDatabaseRoutesOnSchemeNotOnASubstring pins the regression the review
// panel found in this package's FIRST implementation: routing on `://` appearing
// anywhere rather than on a leading scheme.
//
// A keyword-form DSN can carry `://` inside a value — a certificate path, an
// options string, a vault URL. Routed into the URL branch, `url.Parse` accepts
// it with an empty scheme and puts the WHOLE string in Path, which the rewrite
// then replaces wholesale. The measured output for the first case below was
// exactly "/scratch", with a nil error: host, credentials and every other
// keyword silently discarded.
func TestForDatabaseRoutesOnSchemeNotOnASubstring(t *testing.T) {
	cases := []struct {
		name string
		base string
		want string
	}{
		{
			name: "keyword form with a scheme inside a certificate path",
			base: "host=localhost dbname=flow sslcert=/certs/a://b.pem",
			want: "host=localhost dbname=scratch sslcert=/certs/a://b.pem",
		},
		{
			name: "keyword form with an https value",
			base: "host=localhost dbname=flow sslrootcert=https://x/ca.pem",
			want: "host=localhost dbname=scratch sslrootcert=https://x/ca.pem",
		},
		{
			// A `dbname=` inside someone else's quoted value is not the real one;
			// the unquoted one after it is. Getting this wrong either refuses a
			// valid DSN or rewrites the wrong keyword, and both have happened.
			name: "dbname inside another keyword's quoted value is skipped",
			base: "host=localhost application_name='a dbname=b' dbname=real",
			want: "host=localhost application_name='a dbname=b' dbname=scratch",
		},
		{
			// A quote belonging to another keyword must not block the rewrite.
			// Refusing these was the panel's finding: both are ordinary libpq.
			name: "quoted options value does not block the rewrite",
			base: "host=localhost dbname=flow options='-c search_path=myschema'",
			want: "host=localhost dbname=scratch options='-c search_path=myschema'",
		},
		{
			name: "quoted application_name does not block the rewrite",
			base: "host=localhost dbname=flow application_name='my app'",
			want: "host=localhost dbname=scratch application_name='my app'",
		},
		{
			name: "postgresql scheme is a URL",
			base: "postgresql://flow:flow@localhost:5433/flow?sslmode=disable",
			want: "postgresql://flow:flow@localhost:5433/scratch?sslmode=disable",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ForDatabase(c.base, "scratch")
			if err != nil {
				t.Fatalf("ForDatabase(%q) returned error: %v", c.base, err)
			}
			if got != c.want {
				t.Errorf("ForDatabase(%q)\n got %q\nwant %q", c.base, got, c.want)
			}
		})
	}
}

// TestForDatabaseRefusesRatherThanGuessing pins the other half of the contract.
// The predecessor's defining fault was answering confidently when it could not
// do the job; every case here must produce an error, never a best-effort string.
func TestForDatabaseRefusesRatherThanGuessing(t *testing.T) {
	cases := []struct {
		name   string
		base   string
		dbName string
	}{
		{"empty base", "", "scratch"},
		{"whitespace-only base", "   ", "scratch"},
		{"empty database name", "postgres://flow@localhost/flow", ""},
		{"keyword form with no dbname", "host=localhost port=5433 sslmode=disable", "scratch"},
		{"quoted dbname", "host=localhost dbname='my db' sslmode=disable", "scratch"},
		{"backslash-escaped value", `host=localhost dbname=fl\\ ow`, "scratch"},
		{"no scheme, no dbname keyword", "://nonsense", "scratch"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ForDatabase(c.base, c.dbName)
			if err == nil {
				t.Fatalf("ForDatabase(%q, %q) = %q, want an error", c.base, c.dbName, got)
			}
			if got != "" {
				t.Errorf("ForDatabase returned %q alongside its error; it must return no DSN at all", got)
			}
		})
	}
}

// TestForDatabaseEmptyBaseIsDistinguishable pins ErrEmpty as its own sentinel.
// An empty DSN nearly always means an environment variable was exported empty
// rather than left unset, and that is worth telling apart from a parse failure.
func TestForDatabaseEmptyBaseIsDistinguishable(t *testing.T) {
	if _, err := ForDatabase("", "scratch"); !errors.Is(err, ErrEmpty) {
		t.Fatalf("ForDatabase(\"\") error = %v, want ErrEmpty", err)
	}
}

// TestForDatabasePreservesCredentialsAndPort guards the property every caller
// actually depends on: only the database changes.
func TestForDatabasePreservesCredentialsAndPort(t *testing.T) {
	base := "postgres://someone:s3cr3t@db.example.internal:6543/original?sslmode=require&application_name=x"
	got, err := ForDatabase(base, "scratch_1")
	if err != nil {
		t.Fatalf("ForDatabase returned error: %v", err)
	}
	for _, must := range []string{"someone:s3cr3t", "db.example.internal:6543", "sslmode=require", "application_name=x", "/scratch_1"} {
		if !strings.Contains(got, must) {
			t.Errorf("ForDatabase(%q) = %q, missing %q", base, got, must)
		}
	}
	if strings.Contains(got, "/original") {
		t.Errorf("ForDatabase(%q) = %q, still names the original database", base, got)
	}
}
