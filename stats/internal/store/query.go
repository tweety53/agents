// query.go is the entire dynamic-SQL surface of this package. A sort key
// and a filter field are SQL identifiers, and an identifier cannot be a
// bound parameter -- so any request field that ends up in an identifier
// position is resolved through the fixed maps below, never interpolated
// from caller text. A field absent from a map is a rejected request
// (ErrUnknownField), not a query built from the string. A metrics key path
// is the one identifier-shaped value this package does accept from a
// caller; it is validated against a closed syntax rule and then supplied
// as a bound parameter to a JSONB path operator (#>>), never concatenated.
//
// If a change to this package ever needs a SQL fragment built from request
// text anywhere outside this file, that is a defect in the design, not a
// gap to patch with validation -- see design.md's query-allowlist decision.
package store

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

// ErrUnknownField is returned when a filter or sort field resolves to a
// name absent from the relevant allowlist and not a validly-shaped metrics
// key path. The error names the field and the accepted alternatives.
var ErrUnknownField = errors.New("store: unknown query field")

// ErrInvalidMetricsKeyPath is returned when a field names a metrics key
// path (it has the "metrics." prefix) but fails the path's syntax rule --
// for example an empty segment, or a segment containing anything other
// than letters, digits and underscores. This is distinct from
// ErrUnknownField: the caller was attempting the metrics path syntax and
// got the syntax wrong, rather than naming a field that was never a
// candidate.
var ErrInvalidMetricsKeyPath = errors.New("store: invalid metrics key path")

// ErrUnknownFilterOp is returned when a Filter names an operator this
// package does not implement.
var ErrUnknownFilterOp = errors.New("store: unknown filter operator")

// FilterOp is one comparison a Filter can apply. The set is closed: there
// is no "raw" operator that accepts caller-supplied SQL.
type FilterOp string

const (
	OpEq      FilterOp = "eq"
	OpNeq     FilterOp = "neq"
	OpLt      FilterOp = "lt"
	OpLte     FilterOp = "lte"
	OpGt      FilterOp = "gt"
	OpGte     FilterOp = "gte"
	OpNull    FilterOp = "isnull"
	OpNotNull FilterOp = "isnotnull"
)

// Filter restricts a query to rows where Field, resolved through the
// query's allowlist, compares to Value under Op. OpNull and OpNotNull
// ignore Value.
type Filter struct {
	Field string
	Op    FilterOp
	Value any
}

// SortKey orders a query by one allowlisted field. Field undergoes the
// same resolution as a Filter's.
type SortKey struct {
	Field string
	Desc  bool
}

// Query is the caller-facing shape a request is turned into: filters, a
// free-text search term, an ordering, and a page. ListChanges and
// ListStageRuns's Query-accepting entry points -- QueryChanges and
// QueryStageRuns -- are the only places this type is interpreted into SQL.
type Query struct {
	Filters []Filter
	// Search, when non-empty, matches (case-insensitively, as a substring)
	// across the identity fields the target query type defines.
	Search string
	// Sort orders results by zero or more allowlisted fields, applied in
	// order. Whatever Sort produces, a unique tiebreaker (the row's own
	// primary key) is always appended, so the total order is stable
	// regardless of what the caller asked for.
	Sort []SortKey
	// Limit and Offset page the result. Limit <= 0 uses defaultQueryLimit;
	// Limit above maxQueryLimit is clamped to it. Offset < 0 is clamped to
	// 0. Limit == NoLimit disables paging entirely (every matching row,
	// no LIMIT clause) and requires Offset == 0 -- a non-zero Offset
	// alongside NoLimit is rejected with ErrOffsetWithNoLimit rather than
	// silently dropped, since which half of that combination was the
	// caller's mistake cannot be guessed.
	Limit  int
	Offset int
}

const (
	defaultQueryLimit = 50
	maxQueryLimit     = 500

	// NoLimit, set as Query.Limit, disables paging entirely: every matching
	// row is returned and no LIMIT clause is emitted at all. This is
	// deliberately not the zero value -- an unset Limit (0) means "use
	// defaultQueryLimit", exactly as before this constant existed, so a
	// caller must opt into an unbounded result explicitly. ListChanges is
	// the only caller that does: it has to keep returning every change for
	// a project, matching its behaviour from before QueryChanges existed.
	NoLimit = -1
)

func (q Query) limit() int {
	switch {
	case q.Limit <= 0:
		return defaultQueryLimit
	case q.Limit > maxQueryLimit:
		return maxQueryLimit
	default:
		return q.Limit
	}
}

func (q Query) offset() int {
	if q.Offset < 0 {
		return 0
	}
	return q.Offset
}

// unlimited reports whether q asked for every matching row, with no LIMIT
// clause at all, via the NoLimit sentinel.
func (q Query) unlimited() bool { return q.Limit == NoLimit }

// ErrOffsetWithNoLimit is returned when a Query sets both Limit == NoLimit
// and a non-zero Offset. NoLimit already means "no LIMIT clause at all";
// silently dropping the caller's Offset in that case would mean the
// request was not honoured and nothing said so. Rather than guess which
// half of the caller's intent was the mistake, this rejects the
// combination outright -- the same posture query.go takes with every
// other malformed request.
var ErrOffsetWithNoLimit = errors.New("store: offset is meaningless with NoLimit; leave it 0")

// validate checks the parts of q that are wrong regardless of which
// queryable it runs against -- currently just the NoLimit/Offset
// combination. Field-shaped errors (ErrUnknownField,
// ErrInvalidMetricsKeyPath) are still caught by buildClauses, which needs
// the target queryable's allowlist and so cannot run here.
func (q Query) validate() error {
	if q.unlimited() && q.Offset != 0 {
		return ErrOffsetWithNoLimit
	}
	return nil
}

// metricsFieldPrefix is the literal prefix that marks a field as a metrics
// key path (for example "metrics.tokens.main.cache_read" -- a real path a
// harvested stage run's metrics bag actually carries; the harvester never
// writes a bare "tokens.cache_read" scalar, only "tokens.main.<field>",
// "tokens.sidechain.<field>" and "models.<model>.tokens.*", see
// internal/harvest.TokenDelta) rather than a static column name. A field
// must match this prefix exactly -- byte for byte,
// including the trailing dot -- before its remainder is even considered
// for path validation; anything else falls through to the static
// allowlist lookup and is rejected as an unknown field if not found there.
const metricsFieldPrefix = "metrics."

// metricsKeySegmentRE is the syntax rule a single path segment must match:
// this is the entire grammar a metrics key path is allowed to have. The
// path this validates is never assembled into SQL text -- it is bound as
// a parameter to the #>> operator (metricsTextExpr/metricsNumericExpr
// below), so SQL syntax inside a segment is inert regardless of what this
// regex admits. What the regex actually defends is the shape of the key
// vocabulary itself: it bounds a segment to identifier-like characters
// plus '-' (the one character every model name in pricing_seed.go
// requires, e.g. "claude-sonnet-5") so a caller-controlled path stays a
// small, predictable shape rather than admitting arbitrary bytes -- not
// because arbitrary bytes could reach SQL syntax, but because the bound
// parameter still flows into a JSONB path array and a Postgres query
// plan, and a segment made of quotes, brackets, arrows, whitespace or
// semicolons has no legitimate key it could ever name.
var metricsKeySegmentRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)

// Bounds on a metrics key path's shape. Every key this package's own
// vocabulary defines (design.md's metrics bag table) is one or two
// segments of a handful of characters each; these bounds are sized with
// generous headroom above that, not tuned to the exact vocabulary, so a
// legitimate future key never needs the bound raised. What they exist to
// stop is unbounded caller-controlled work reaching the query planner --
// a path with thousands of segments, or one multi-kilobyte segment --
// once task 11 feeds this straight from HTTP parameters. The check lives
// here, next to the syntax it bounds, rather than at the HTTP layer: a
// caller that reaches this function by any route gets the same limit.
const (
	maxMetricsKeyPathSegments  = 8
	maxMetricsKeySegmentLength = 64
)

// validateMetricsKeyPath checks field against metricsFieldPrefix,
// metricsKeySegmentRE and the size bounds above, and, if it passes,
// returns the path with the prefix stripped and split on ".". The
// returned path is what gets bound as a parameter to the #>> operator --
// it is never assembled into SQL text.
func validateMetricsKeyPath(field string) ([]string, error) {
	rest := strings.TrimPrefix(field, metricsFieldPrefix)
	if rest == field || rest == "" {
		return nil, fmt.Errorf("%w: %q", ErrInvalidMetricsKeyPath, field)
	}
	segments := strings.Split(rest, ".")
	if len(segments) > maxMetricsKeyPathSegments {
		return nil, fmt.Errorf("%w: %q has %d path segments, more than the %d allowed",
			ErrInvalidMetricsKeyPath, field, len(segments), maxMetricsKeyPathSegments)
	}
	for _, seg := range segments {
		if len(seg) > maxMetricsKeySegmentLength {
			return nil, fmt.Errorf("%w: %q has a segment longer than the %d characters allowed",
				ErrInvalidMetricsKeyPath, field, maxMetricsKeySegmentLength)
		}
		if !metricsKeySegmentRE.MatchString(seg) {
			return nil, fmt.Errorf("%w: %q", ErrInvalidMetricsKeyPath, field)
		}
	}
	return segments, nil
}

// queryable describes one target table's query surface: the allowlist
// mapping accepted field names to real column expressions, the column
// (if any) holding its metrics bag, the columns free-text search compares
// against, and the column that breaks ties in a total order.
//
// Every string in fields' values, in searchColumns and in idColumn is
// package-authored -- copied verbatim from the literals below -- and never
// derived from a caller. That is what makes it safe to interpolate them
// directly into SQL text: they are not caller input at all, merely
// selected by caller input.
type queryable struct {
	idColumn      string
	fields        map[string]string
	metricsColumn string // "" if this target has no metrics bag
	searchColumns []string
}

// changeFieldColumns is the allowlist for every field a caller may filter
// or sort by on the changes table -- including, by inclusion in
// stageRunFieldColumns below, on a stage run's owning change via join.
// "PR URL presence" is delivered by OpNull/OpNotNull against pr_url
// directly rather than a separate boolean pseudo-field: presence is
// already exactly what those two operators test.
var changeFieldColumns = map[string]string{
	"project":             "c.project_key",
	"name":                "c.name",
	"state":               "c.state",
	"branch":              "c.branch",
	"jira_issue":          "c.jira_issue",
	"planning_effort":     "c.planning_effort",
	"review_panel_roster": "c.review_panel_roster",
	"pr_url":              "c.pr_url",
	"updated_at":          "c.updated_at",
	"updated_by":          "c.updated_by",
}

// stageRunOwnFieldColumns is the allowlist for fields that belong to a
// stage run itself. duration_ms is computed rather than stored: a run
// with no ended_at yet produces NULL here, not a fabricated zero, the same
// absence-is-not-a-value rule the metrics bag follows.
var stageRunOwnFieldColumns = map[string]string{
	"command":     "sr.command",
	"stage":       "sr.stage",
	"attempt":     "sr.attempt",
	"harness":     "sr.harness",
	"outcome":     "sr.outcome",
	"started_at":  "sr.started_at",
	"ended_at":    "sr.ended_at",
	"duration_ms": "(EXTRACT(EPOCH FROM (sr.ended_at - sr.started_at)) * 1000)",
	"repo_root":   "sr.repo_root",
	// session_id was not part of task 3.1's original allowlist -- it was
	// added by task 9 (internal/harvest) because attribution's own
	// non-negotiable rule ("a message belongs to the open stage run whose
	// session_id matches") has no other way to ask the store "which stage
	// runs exist for this session" without either a bespoke SQL method
	// (a second, narrower dynamic-SQL surface this file's own header
	// forbids: "the entire dynamic-SQL surface of this package") or
	// building the query outside internal/store entirely (which the
	// global constraint "internal/store is the only package that builds
	// SQL" forbids even more directly). Reusing this file's existing,
	// already-reviewed allowlist mechanism -- built by task 3.1 explicitly
	// for future fields like this one -- is the smaller, more consistent
	// change of the two.
	"session_id": "sr.session_id",
}

// stageRunFieldColumns is stageRunOwnFieldColumns plus changeFieldColumns
// -- "the owning change's fields by join", per tasks.md -- computed once at
// package init. The two maps share no keys, verified by
// TestStageRunAndChangeFieldsDoNotCollide.
var stageRunFieldColumns = mergeFieldMaps(stageRunOwnFieldColumns, changeFieldColumns)

func mergeFieldMaps(maps ...map[string]string) map[string]string {
	out := make(map[string]string, len(maps[0]))
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

// changeQueryable and stageRunQueryable are the two query surfaces
// QueryChanges and QueryStageRuns build their SQL against.
var changeQueryable = queryable{
	idColumn:      "c.id",
	fields:        changeFieldColumns,
	metricsColumn: "",
	searchColumns: []string{"c.name", "c.jira_issue", "c.branch", "c.updated_by"},
}

var stageRunQueryable = queryable{
	idColumn:      "sr.id",
	fields:        stageRunFieldColumns,
	metricsColumn: "sr.metrics",
	searchColumns: []string{
		"c.name", "c.jira_issue", "c.branch", "c.updated_by",
		"sr.command", "sr.stage", "sr.repo_root",
	},
}

// AllowedChangeFields reports every field name QueryChanges accepts in a
// Filter or a SortKey, sorted. It exists so a test asserting "every
// allowlisted field works" can iterate the allowlist itself rather than a
// hand-written copy of it that a later addition could silently bypass.
func AllowedChangeFields() []string { return sortedFieldNames(changeFieldColumns) }

// AllowedStageRunFields reports every field name QueryStageRuns accepts in
// a Filter or a SortKey, sorted -- including the owning change's fields by
// join. See AllowedChangeFields.
func AllowedStageRunFields() []string { return sortedFieldNames(stageRunFieldColumns) }

func sortedFieldNames(fields map[string]string) []string {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// newUnknownFieldError names field and the fields this queryable accepts,
// noting the metrics key path syntax too when this queryable has a
// metrics bag.
func newUnknownFieldError(context, field string, qa queryable) error {
	accepted := strings.Join(sortedFieldNames(qa.fields), ", ")
	if qa.metricsColumn != "" {
		accepted += `, or a metrics key path ("metrics.<key>[.<key>...]")`
	}
	return fmt.Errorf("%w: %s %q not recognised; accepted: %s", ErrUnknownField, context, field, accepted)
}

// resolvedField is what a field name resolves to: either a trusted,
// package-authored SQL column expression (expr), or -- for a metrics key
// path -- the validated path to bind as a parameter to the #>> operator.
type resolvedField struct {
	expr      string
	isMetrics bool
	keyPath   []string
}

// resolveField is the one place a request's field name is turned into
// something usable in a query. context ("filter" or "sort") only shapes
// the error message. Every return path is either a lookup in a
// package-authored map or a syntax-validated metrics path bound as a
// parameter later -- never a caller string placed where SQL expects an
// identifier.
func resolveField(context string, qa queryable, field string) (resolvedField, error) {
	if col, ok := qa.fields[field]; ok {
		return resolvedField{expr: col}, nil
	}
	if qa.metricsColumn != "" && strings.HasPrefix(field, metricsFieldPrefix) {
		path, err := validateMetricsKeyPath(field)
		if err != nil {
			return resolvedField{}, err
		}
		return resolvedField{isMetrics: true, keyPath: path}, nil
	}
	return resolvedField{}, newUnknownFieldError(context, field, qa)
}

// binder accumulates bound query parameters and hands back their
// placeholders ($N), continuing numbering from start so a WHERE clause's
// parameters and an ORDER BY clause's parameters can be combined into one
// query without colliding.
type binder struct {
	start int
	args  []any
}

func (b *binder) bind(v any) string {
	b.args = append(b.args, v)
	return fmt.Sprintf("$%d", b.start+len(b.args))
}

// metricsTextExpr returns the SQL text extraction of a metrics key path
// -- (<column> #>> $N::text[]) -- binding path as the parameter. It
// yields NULL when the key is absent, which is what makes absence order
// and compare distinctly from a recorded zero: NULL is never equal to,
// nor less than, nor greater than, any value, including 0.
func metricsTextExpr(b *binder, metricsColumn string, path []string) string {
	return fmt.Sprintf("(%s #>> %s::text[])", metricsColumn, b.bind(path))
}

// metricsNumericExpr is metricsTextExpr cast to numeric, for the ordering
// comparisons (lt/lte/gt/gte) and for sorting, where the metrics bag's
// "interesting numbers" need a numeric rather than lexical order.
func metricsNumericExpr(b *binder, metricsColumn string, path []string) string {
	return fmt.Sprintf("(%s #>> %s::text[])::numeric", metricsColumn, b.bind(path))
}

// buildFilterSQL renders one Filter as a boolean SQL expression, binding
// every value it contains. A metrics-field equality/inequality compares
// against the JSON scalar's text representation (fmt.Sprint(f.Value)); a
// metrics-field ordering comparison casts to numeric, since ordering only
// makes sense on the bag's numeric keys.
func buildFilterSQL(b *binder, qa queryable, f Filter) (string, error) {
	rf, err := resolveField("filter", qa, f.Field)
	if err != nil {
		return "", err
	}

	numericOp := f.Op == OpLt || f.Op == OpLte || f.Op == OpGt || f.Op == OpGte

	var expr string
	switch {
	case rf.isMetrics && numericOp:
		expr = metricsNumericExpr(b, qa.metricsColumn, rf.keyPath)
	case rf.isMetrics:
		expr = metricsTextExpr(b, qa.metricsColumn, rf.keyPath)
	default:
		expr = rf.expr
	}

	switch f.Op {
	case OpEq, OpNeq:
		val := f.Value
		if rf.isMetrics {
			val = fmt.Sprint(f.Value)
		}
		op := "="
		if f.Op == OpNeq {
			op = "<>"
		}
		return expr + " " + op + " " + b.bind(val), nil
	case OpLt:
		return expr + " < " + b.bind(f.Value), nil
	case OpLte:
		return expr + " <= " + b.bind(f.Value), nil
	case OpGt:
		return expr + " > " + b.bind(f.Value), nil
	case OpGte:
		return expr + " >= " + b.bind(f.Value), nil
	case OpNull:
		return expr + " IS NULL", nil
	case OpNotNull:
		return expr + " IS NOT NULL", nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnknownFilterOp, f.Op)
	}
}

// likeEscapeChar is the escape character buildSearchSQL declares via an
// explicit ESCAPE clause, so ILIKE's own default escape behaviour is never
// relied on implicitly.
const likeEscapeChar = `\`

// likeMetacharReplacer escapes the three characters that are meaningful to
// ILIKE -- the escape character itself, then its two wildcards -- so a
// literal "%", "_" or "\" in a caller's search term is matched as that
// literal character rather than as a wildcard. The escape character is
// replaced first: escaping it after the wildcards would double-escape the
// backslashes the wildcard replacements just introduced.
var likeMetacharReplacer = strings.NewReplacer(
	likeEscapeChar, likeEscapeChar+likeEscapeChar,
	"%", likeEscapeChar+"%",
	"_", likeEscapeChar+"_",
)

// buildSearchSQL renders q's free-text term as one OR-joined group of
// case-insensitive substring matches across qa.searchColumns, with the
// term bound once and reused across every column. The term is escaped
// before binding so "%" and "_" in it match themselves literally rather
// than acting as ILIKE wildcards -- the spec promises a substring match,
// not a pattern match. It returns "" if the term is empty.
func buildSearchSQL(b *binder, qa queryable, term string) string {
	if term == "" || len(qa.searchColumns) == 0 {
		return ""
	}
	ph := b.bind("%" + likeMetacharReplacer.Replace(term) + "%")
	parts := make([]string, len(qa.searchColumns))
	for i, col := range qa.searchColumns {
		parts[i] = col + " ILIKE " + ph + ` ESCAPE '` + likeEscapeChar + `'`
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

// buildWhereSQL renders q's filters and search term into one WHERE clause
// (or "" if q carries neither), binding every value through b.
func buildWhereSQL(b *binder, qa queryable, q Query) (string, error) {
	var conds []string
	for _, f := range q.Filters {
		cond, err := buildFilterSQL(b, qa, f)
		if err != nil {
			return "", err
		}
		conds = append(conds, cond)
	}
	if s := buildSearchSQL(b, qa, q.Search); s != "" {
		conds = append(conds, s)
	}
	if len(conds) == 0 {
		return "", nil
	}
	return "WHERE " + strings.Join(conds, " AND "), nil
}

// buildOrderSQL renders q's sort keys into an ORDER BY clause, always
// ending with qa.idColumn ASC as a unique tiebreaker so the total order is
// stable no matter what q.Sort asked for -- a row can never appear on two
// pages, or be skipped, as a caller pages through equal sort keys. Every
// user-requested term carries NULLS LAST explicitly, in both directions,
// so a record missing the sorted key -- a metrics key most of all -- sorts
// as absent in one consistent place, distinct from a recorded zero,
// rather than one that silently moves depending on ASC vs DESC.
func buildOrderSQL(b *binder, qa queryable, q Query) (string, error) {
	terms := make([]string, 0, len(q.Sort)+1)
	for _, sk := range q.Sort {
		rf, err := resolveField("sort", qa, sk.Field)
		if err != nil {
			return "", err
		}
		var expr string
		if rf.isMetrics {
			expr = metricsNumericExpr(b, qa.metricsColumn, rf.keyPath)
		} else {
			expr = rf.expr
		}
		dir := "ASC"
		if sk.Desc {
			dir = "DESC"
		}
		terms = append(terms, fmt.Sprintf("%s %s NULLS LAST", expr, dir))
	}
	terms = append(terms, qa.idColumn+" ASC")
	return "ORDER BY " + strings.Join(terms, ", "), nil
}

// buildClauses renders q against qa into a WHERE clause and its own
// parameters (suitable for a COUNT(*) query on its own), and an ORDER BY
// clause and its own parameters numbered to continue immediately after the
// WHERE clause's (suitable for appending to the same query, followed by a
// LIMIT/OFFSET numbered to continue after both). Every error path returns
// before any SQL runs, per query-allowlist: a rejected field never reaches
// the database.
func (q Query) buildClauses(qa queryable) (whereSQL string, whereArgs []any, orderSQL string, orderArgs []any, err error) {
	if err := q.validate(); err != nil {
		return "", nil, "", nil, err
	}
	wb := &binder{}
	whereSQL, err = buildWhereSQL(wb, qa, q)
	if err != nil {
		return "", nil, "", nil, err
	}
	ob := &binder{start: len(wb.args)}
	orderSQL, err = buildOrderSQL(ob, qa, q)
	if err != nil {
		return "", nil, "", nil, err
	}
	return whereSQL, wb.args, orderSQL, ob.args, nil
}

// queryTxOptions is the one transaction options value QueryChanges and
// QueryStageRuns each pass to BeginTx, so the isolation level and access
// mode that guard their count-then-page consistency live in exactly one
// place. IsoLevel: pgx.RepeatableRead is load-bearing, not a hardening
// nice-to-have: the pool's session default, READ COMMITTED, takes a fresh
// snapshot per statement rather than per transaction, so a write landing
// between the count and the page -- even on the same still-open
// transaction -- would make total disagree with what the page actually
// contains (TestRepeatableReadPreventsCountDrift reproduces this
// directly). AccessMode: pgx.ReadOnly documents that neither method ever
// writes inside this transaction, which is also why the 40001
// serialization failure REPEATABLE READ can raise for a writing or
// locking-read transaction cannot occur here -- see QueryChanges' own
// comment for the full reasoning.
//
// TestQueryTxOptionsUsesRepeatableRead asserts this function's return
// value directly, so a change or drop of either field here is caught
// immediately -- deterministically, not by racing a concurrent writer
// against QueryChanges' own few-microsecond window between statements,
// which would be a flaky test for no better guarantee.
func queryTxOptions() pgx.TxOptions {
	return pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}
}

// QueryTxOptionsForTest exposes queryTxOptions() to this package's
// external tests (package store_test, which cannot see an unexported
// function). It exists purely so TestQueryTxOptionsUsesRepeatableRead can
// assert the exact value QueryChanges and QueryStageRuns build their
// transactions with; it is not part of this package's API for any other
// caller.
func QueryTxOptionsForTest() pgx.TxOptions { return queryTxOptions() }
