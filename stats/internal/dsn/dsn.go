// Package dsn rewrites the database segment of a PostgreSQL connection string.
//
// WHY THIS PACKAGE EXISTS. The Go suites each create a uniquely-named scratch
// database and then connect to it, so every one of them needs "the admin DSN,
// but pointing at this other database". That derivation must come from ONE
// place, and it must not be attempted by hand.
//
// It was, twice, and both attempts were defects the review panel caught:
//
//   - First the per-test DSN repeated the credentials as a literal while the
//     admin DSN read an environment override. Setting FLOW_STATS_ADMIN_DSN then
//     moved only half the connections: the admin step succeeded and every test
//     failed, so the override that exists to point the suite at a different
//     Postgres could not actually do it.
//   - Then the derivation was written as string surgery -- LastIndex("/") for
//     the database segment, Index("?") for the query -- and copied into seven
//     files. That silently produced a CORRUPTED DSN, with no error and no
//     panic, for inputs an operator overriding the default is unusually likely
//     to have: a query string containing a slash (`?options=-c
//     search_path=a/b`), a socket-host DSN (`postgres:///db?host=/var/run/...`,
//     where the rewrite lands inside the host parameter), and a libpq
//     key=value DSN (`host=... dbname=...`), where it silently did nothing at
//     all. An empty DSN returned an empty DSN.
//
// Both failures share a shape: a helper that answers confidently when it should
// refuse. So this one parses properly and returns an error rather than a
// best-effort string, and it lives in a real package so the seven call sites
// share one implementation instead of seven copies of one bug.
package dsn

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ErrEmpty is returned for an empty base DSN. It is a distinct error because an
// empty DSN almost always means an environment variable was set to the empty
// string rather than left unset, which is a configuration mistake worth naming
// rather than folding into a generic parse failure.
var ErrEmpty = errors.New("dsn: base DSN is empty")

// dbnameKeyword matches the `dbname=<value>` keyword of a libpq key=value
// connection string. libpq accepts both URL and key=value forms, and the two
// need different rewrites; the value is delimited by whitespace, since this
// package does not attempt to support libpq's single-quoted values (see
// ForDatabase's refusal below).
var dbnameKeyword = regexp.MustCompile(`(^|\s)dbname=([^\s]*)`)

// ForDatabase returns base with its database changed to dbName.
//
// It accepts both connection-string forms libpq accepts:
//
//   - a URL, `postgres://user:pass@host:port/dbname?params`, where the database
//     is the path and everything else -- credentials, host, port, query -- is
//     preserved exactly, including a query value that itself contains a slash;
//   - a keyword string, `host=... dbname=... sslmode=...`, where only the
//     `dbname` keyword's value is replaced.
//
// It returns an error rather than a best-effort string when it cannot do that
// safely: an empty base, a base it cannot parse, or a keyword string carrying a
// single-quoted or escaped `dbname` value, which this package deliberately does
// not try to rewrite. Returning a wrong DSN is worse than refusing, because the
// caller connects to it and reports whatever that connection says.
func ForDatabase(base, dbName string) (string, error) {
	if strings.TrimSpace(base) == "" {
		return "", ErrEmpty
	}
	if dbName == "" {
		return "", errors.New("dsn: database name is empty")
	}

	// Route on a LEADING SCHEME, never on `://` appearing anywhere: a keyword
	// DSN can carry one inside a value. See this package's doc comment for what
	// that mistake actually did.
	if !strings.HasPrefix(base, "postgres://") && !strings.HasPrefix(base, "postgresql://") {
		// Keyword form. Refuse only what actually blocks the rewrite -- the
		// `dbname` value being quoted or escaped -- NOT a quote anywhere in the
		// string.
		//
		// The first attempt refused on `strings.ContainsAny(base, "'\\")`, which
		// rejected ordinary overrides an operator would plausibly write:
		// `options='-c search_path=myschema'` and `application_name='my app'`
		// both carry a quote nowhere near `dbname`, and both were refused. That
		// contradicted this function's own doc comment, which promised to refuse
		// a quoted *dbname value*. Trading silent wrongness for an override
		// surface narrower than the documented one is its own defect; measured
		// by running both shapes through it.
		return rewriteKeywordDBName(base, dbName)
	}

	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("dsn: parsing %q: %w", base, err)
	}
	// There is deliberately NO second "did it parse with a scheme?" guard here.
	// One was written, as belt and braces, and it was dead code: anything that
	// reaches this point began with a literal `postgres://` or `postgresql://`,
	// and Go's url.Parse takes its scheme from exactly that prefix, so u.Scheme
	// can never be empty; anything malformed enough to lose it fails at Parse
	// above instead. Measured by deleting the guard and re-running the suite --
	// nothing failed -- and by parsing the shapes that could plausibly reach it.
	//
	// It is called out rather than silently omitted because an unreachable check
	// whose comment claims it prevents silent corruption is worse than no check:
	// it reads as protection during review and provides none. The routing above
	// is what makes the guarantee, so that is where the reasoning belongs.
	// url.Parse keeps the leading slash in Path. Setting it this way replaces the
	// database and leaves RawQuery, User and Host untouched, which is the whole
	// point: a query value containing a slash is query, not path.
	u.Path = "/" + dbName
	return u.String(), nil
}

// rewriteKeywordDBName replaces the value of the first `dbname` keyword that is
// not itself inside a quoted value.
//
// Quote position is what distinguishes the two cases the panel found. A quote
// belonging to some OTHER keyword (`options='-c search_path=a/b'`) must not stop
// the rewrite; a `dbname=` appearing INSIDE such a quoted value
// (`application_name='x dbname=y' dbname=real`) must not be mistaken for the
// real one. Both are decided by whether an odd number of quotes precedes the
// match, so both fall out of one rule rather than two special cases.
func rewriteKeywordDBName(base, dbName string) (string, error) {
	for _, m := range dbnameKeyword.FindAllStringSubmatchIndex(base, -1) {
		start, valStart, valEnd := m[0], m[4], m[5]
		if strings.Count(base[:start], "'")%2 != 0 {
			continue // this `dbname=` sits inside someone else's quoted value
		}
		value := base[valStart:valEnd]
		if strings.HasPrefix(value, "'") || strings.Contains(value, `\`) {
			return "", fmt.Errorf("dsn: refusing to rewrite a quoted or escaped dbname value in %q", base)
		}
		return base[:valStart] + dbName + base[valEnd:], nil
	}
	return "", fmt.Errorf("dsn: no unquoted dbname keyword to rewrite in %q", base)
}
