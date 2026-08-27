package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestFilterBySessionID asserts that QueryStageRuns' "session_id" filter
// (added to stageRunOwnFieldColumns by task 9, internal/harvest) actually
// resolves to the session_id column and returns the right rows.
//
// This exists because, before this test, nothing exercised that mapping
// at all: storeWindowSource.WindowsForSession (cmd/flowd/main.go) is
// the sole production caller of this filter, and every other test in
// this package either omits a session_id filter or uses the fixed
// literal "session-1" that baseBeginInput hands every stage run by
// default. A field entry mapped to the wrong column -- or transposed
// with a neighbouring one -- would have compiled, every existing test
// would have stayed green, and every stage run's token attribution would
// have silently gone to nothing (task 9's post-commit review, finding
// F6): QueryStageRuns would either error on an unrelated column type or,
// worse, simply return zero rows for every real session id the harvester
// ever asked about.
func TestFilterBySessionID(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-session-filter-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	target := fmt.Sprintf("session-target-%d", time.Now().UnixNano())
	other := fmt.Sprintf("session-other-%d", time.Now().UnixNano())

	matchIn := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	matchIn.SessionID = &target
	matchRun, err := st.BeginStage(ctx, matchIn)
	if err != nil {
		t.Fatalf("BeginStage (match): %v", err)
	}

	otherIn := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
	otherIn.SessionID = &other
	if _, err := st.BeginStage(ctx, otherIn); err != nil {
		t.Fatalf("BeginStage (other): %v", err)
	}

	rows, total, err := st.QueryStageRuns(ctx, store.Query{
		Filters: []store.Filter{{Field: "session_id", Op: store.OpEq, Value: target}},
	})
	if err != nil {
		t.Fatalf("QueryStageRuns: %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Fatalf("got %d rows (total %d) filtering by session_id=%q, want exactly 1", len(rows), total, target)
	}
	if rows[0].ID != matchRun.ID {
		t.Errorf("QueryStageRuns(session_id=%q) matched run %d, want %d", target, rows[0].ID, matchRun.ID)
	}
	if rows[0].SessionID == nil || *rows[0].SessionID != target {
		t.Errorf("matched row's SessionID = %v, want %q", rows[0].SessionID, target)
	}
}

// TestStageRunAndChangeFieldsDoNotCollide asserts that merging a stage
// run's own field allowlist with the change field allowlist (so a stage
// run query can filter, sort and search on its owning change's fields by
// join) does not lose an entry to a name collision. mergeFieldMaps is a
// plain map merge: a collision would silently let one map's entry
// overwrite the other's, which this count-based assertion would catch --
// unlike a test that merely checks each own field is *present*, which
// would still pass if two entries collided and only one survived.
func TestStageRunAndChangeFieldsDoNotCollide(t *testing.T) {
	const stageRunOwnFieldCount = 10 // command, stage, attempt, harness, outcome, started_at, ended_at, duration_ms, repo_root, session_id

	got := len(store.AllowedStageRunFields())
	want := len(store.AllowedChangeFields()) + stageRunOwnFieldCount
	if got != want {
		t.Fatalf("AllowedStageRunFields has %d entries, want %d (%d change fields + %d stage run's own) -- a name collision would silently drop one",
			got, want, len(store.AllowedChangeFields()), stageRunOwnFieldCount)
	}
}

// TestSortByEveryAllowlistedField iterates AllowedChangeFields and
// AllowedStageRunFields directly, rather than a hand-written copy of the
// list, so a field added to either allowlist without ever being exercised
// by a sort is impossible: the moment the map grows, this test grows with
// it.
func TestSortByEveryAllowlistedField(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-sort-fields-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	runStage(t, st, in, json.RawMessage(`{"cost_usd":1,"model":"opus"}`), in.StartedAt.Add(time.Minute), "completed")

	for _, field := range store.AllowedChangeFields() {
		t.Run("change/"+field, func(t *testing.T) {
			if _, _, err := st.QueryChanges(ctx, store.Query{Sort: []store.SortKey{{Field: field}}}); err != nil {
				t.Fatalf("QueryChanges sort by %q: %v", field, err)
			}
			if _, _, err := st.QueryChanges(ctx, store.Query{Sort: []store.SortKey{{Field: field, Desc: true}}}); err != nil {
				t.Fatalf("QueryChanges sort by %q DESC: %v", field, err)
			}
		})
	}

	for _, field := range store.AllowedStageRunFields() {
		t.Run("stagerun/"+field, func(t *testing.T) {
			if _, _, err := st.QueryStageRuns(ctx, store.Query{Sort: []store.SortKey{{Field: field}}}); err != nil {
				t.Fatalf("QueryStageRuns sort by %q: %v", field, err)
			}
			if _, _, err := st.QueryStageRuns(ctx, store.Query{Sort: []store.SortKey{{Field: field, Desc: true}}}); err != nil {
				t.Fatalf("QueryStageRuns sort by %q DESC: %v", field, err)
			}
		})
	}
}

func TestUnknownSortFieldIsRejected(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	_, _, err := st.QueryChanges(ctx, store.Query{Sort: []store.SortKey{{Field: "nope"}}})
	if !errors.Is(err, store.ErrUnknownField) {
		t.Fatalf("QueryChanges sort by unknown field: got %v, want ErrUnknownField", err)
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Errorf("error %q does not name the rejected field", err)
	}
	for _, name := range store.AllowedChangeFields() {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not list accepted field %q", err, name)
		}
	}
}

func TestUnknownFilterFieldIsRejected(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	_, _, err := st.QueryStageRuns(ctx, store.Query{
		Filters: []store.Filter{{Field: "nope", Op: store.OpEq, Value: "x"}},
	})
	if !errors.Is(err, store.ErrUnknownField) {
		t.Fatalf("QueryStageRuns filter by unknown field: got %v, want ErrUnknownField", err)
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Errorf("error %q does not name the rejected field", err)
	}
}

// TestSortInjectionAttemptIsRejected uses real injection payloads as sort
// field names -- the position that would land in an ORDER BY clause if
// this package ever built one from caller text. Each assertion checks the
// *specific* typed error (ErrUnknownField), not merely that some error
// came back: a test that accepted any error would also pass against a
// build that rejected the payload for the wrong reason, or that silently
// fell back to a default sort instead of refusing the request. The final
// GetChange call separately proves the database itself is unharmed --
// the seeded row survives every payload untouched, and the changes table
// still exists to be queried at all.
func TestSortInjectionAttemptIsRejected(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-injection-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	payloads := []string{
		"name; DROP TABLE changes",
		"name) UNION SELECT",
		`name' OR '1'='1`,
		"name -- comment",
		"name/*",
		"metrics->'x'",
		`"name"`,
	}

	for _, p := range payloads {
		t.Run(p, func(t *testing.T) {
			_, _, err := st.QueryChanges(ctx, store.Query{Sort: []store.SortKey{{Field: p}}})
			if !errors.Is(err, store.ErrUnknownField) {
				t.Fatalf("sort by %q: got %v, want ErrUnknownField", p, err)
			}
			// The error renders the field with %q, so compare against the
			// same rendering rather than the raw payload -- a payload that
			// itself contains a double quote (like `"name"`) would
			// otherwise never match a literal substring check.
			if !strings.Contains(err.Error(), fmt.Sprintf("%q", p)) {
				t.Errorf("error %q does not name the rejected field %q", err, p)
			}

			_, _, err = st.QueryStageRuns(ctx, store.Query{Filters: []store.Filter{{Field: p, Op: store.OpEq, Value: "x"}}})
			if !errors.Is(err, store.ErrUnknownField) {
				t.Fatalf("filter by %q: got %v, want ErrUnknownField", p, err)
			}
		})
	}

	got, err := st.GetChange(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("GetChange after injection attempts: %v (the changes table should be untouched)", err)
	}
	if got.Name != "kan-1" {
		t.Errorf("seeded change altered by an injection attempt: got name %q", got.Name)
	}
}

func TestFilterByMetricsKeyPath(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-metrics-filter-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	// "tokens.main.cache_read" is the real wire shape a harvested batch
	// writes (internal/harvest.TokenDelta: {main: Bucket, sidechain:
	// Bucket}, each Bucket an object with its own "cache_read" field) --
	// never a bare "tokens.cache_read" scalar, which no real stage run's
	// metrics bag has ever carried.
	matchIn := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	matchRun := runStage(t, st, matchIn, json.RawMessage(`{"tokens":{"main":{"cache_read":100}}}`), matchIn.StartedAt.Add(time.Minute), "completed")

	otherIn := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
	otherRun := runStage(t, st, otherIn, json.RawMessage(`{"tokens":{"main":{"cache_read":250}}}`), otherIn.StartedAt.Add(time.Minute), "completed")

	unrelatedIn := baseBeginInput(projectKey, "kan-1", "/myflow-do", "finish")
	runStage(t, st, unrelatedIn, json.RawMessage(`{"tokens":{"main":{"input":9}}}`), unrelatedIn.StartedAt.Add(time.Minute), "completed")

	rows, total, err := st.QueryStageRuns(ctx, store.Query{
		Filters: []store.Filter{{Field: "metrics.tokens.main.cache_read", Op: store.OpEq, Value: 100}},
	})
	if err != nil {
		t.Fatalf("QueryStageRuns: %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Fatalf("got %d rows (total %d), want exactly 1", len(rows), total)
	}
	if rows[0].ID != matchRun.ID {
		t.Errorf("QueryStageRuns matched run %d, want %d", rows[0].ID, matchRun.ID)
	}

	// A numeric comparison operator exercises the metrics numeric cast
	// path, distinct from the equality path above which compares text.
	rows, total, err = st.QueryStageRuns(ctx, store.Query{
		Filters: []store.Filter{{Field: "metrics.tokens.main.cache_read", Op: store.OpGt, Value: 150}},
	})
	if err != nil {
		t.Fatalf("QueryStageRuns (gt): %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Fatalf("gt filter got %d rows (total %d), want exactly 1", len(rows), total)
	}
	if rows[0].ID != otherRun.ID {
		t.Errorf("QueryStageRuns (gt) matched run %d, want %d", rows[0].ID, otherRun.ID)
	}
}

// TestMalformedMetricsKeyPathIsRejected asserts the syntax rule a metrics
// key path must satisfy, distinguishing this failure (ErrInvalidMetricsKeyPath)
// from an ordinary unknown field (ErrUnknownField): the caller was
// attempting the metrics path syntax, not naming some unrelated field.
func TestMalformedMetricsKeyPathIsRejected(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	malformed := []string{
		"metrics.",
		"metrics..cache_read",
		"metrics.tokens.",
		"metrics.tokens cache_read",
		"metrics.tokens->cache_read",
		"metrics.1cache_read",
		"metrics.tokens.-cache_read",
		"metrics.tokens.cache_read;DROP TABLE stage_runs",
		`metrics.tokens.'cache_read'`,
	}
	for _, field := range malformed {
		t.Run(field, func(t *testing.T) {
			_, _, err := st.QueryStageRuns(ctx, store.Query{Sort: []store.SortKey{{Field: field}}})
			if !errors.Is(err, store.ErrInvalidMetricsKeyPath) {
				t.Fatalf("sort by %q: got %v, want ErrInvalidMetricsKeyPath", field, err)
			}
			if errors.Is(err, store.ErrUnknownField) {
				t.Errorf("sort by %q: got ErrUnknownField too -- must be exactly ErrInvalidMetricsKeyPath", field)
			}
		})
	}
}

// TestModelMetricsKeyPathIsAccepted asserts that a real model-scoped
// metrics key path -- the shape task 22's models bucket actually uses,
// hyphens and all (pricing_seed.go's model names: claude-sonnet-5,
// claude-opus-5, claude-haiku-4-5, claude-opus-4-8) -- is accepted as a
// sort key rather than rejected. The path is bound as a parameter to the
// #>> operator (metricsTextExpr/metricsNumericExpr in query.go), never
// assembled into SQL text, so admitting '-' into the segment grammar
// widens what a caller may name, not what SQL a caller may inject; this
// test exercises that end to end against the real store, not just the
// validator.
func TestModelMetricsKeyPathIsAccepted(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-model-metric-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	run := runStage(t, st, in, json.RawMessage(`{"models":{"claude-sonnet-5":{"cost_usd":1.5}}}`), in.StartedAt.Add(time.Minute), "completed")

	rows, _, err := st.QueryStageRuns(ctx, store.Query{
		Sort: []store.SortKey{{Field: "metrics.models.claude-sonnet-5.cost_usd"}},
	})
	if err != nil {
		t.Fatalf("QueryStageRuns: %v", err)
	}
	if len(rows) != 1 || rows[0].ID != run.ID {
		t.Fatalf("got %d rows, want 1 row for run %d", len(rows), run.ID)
	}
}

// TestAbsentMetricSortsDistinctlyFromZero seeds one stage run whose
// metrics record tokens.main.cache_read as 0 and one whose metrics record
// tokens.main under a different key entirely (never mentioning
// cache_read), then sorts by metrics.tokens.main.cache_read ascending --
// the real wire shape a harvested batch writes (internal/harvest.Bucket),
// not a bare "tokens.cache_read" scalar no real stage run ever carries.
// If absence were ever coerced to 0, the two runs would compare equal and
// their relative order would be undefined (tiebreaker only,
// indistinguishable from a bug); instead the absent run must sort after
// the recorded zero, per the NULLS LAST convention buildOrderSQL
// documents. absentRun is seeded before zeroRun so it holds the lower
// BIGSERIAL id: buildOrderSQL's unconditional idColumn ASC tiebreaker
// would then order a tie as [absent, zero], distinct from the expected
// [zero, absent] -- so a coerce-to-zero (or NULL-both) regression that
// ties on the primary key is caught by that mismatch instead of the
// tiebreaker coincidentally reproducing the right answer.
func TestAbsentMetricSortsDistinctlyFromZero(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-absent-metric-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	absentIn := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
	absentRun := runStage(t, st, absentIn, json.RawMessage(`{"tokens":{"main":{"cache_creation":5}}}`), absentIn.StartedAt.Add(time.Minute), "completed")

	zeroIn := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	zeroRun := runStage(t, st, zeroIn, json.RawMessage(`{"tokens":{"main":{"cache_read":0}}}`), zeroIn.StartedAt.Add(time.Minute), "completed")

	if absentRun.ID >= zeroRun.ID {
		t.Fatalf("fixture invariant broken: absentRun (id=%d) must be seeded before zeroRun (id=%d) — "+
			"otherwise buildOrderSQL's unconditional idColumn ASC tiebreak can mask a coerce-to-zero regression",
			absentRun.ID, zeroRun.ID)
	}

	rows, _, err := st.QueryStageRuns(ctx, store.Query{
		Sort: []store.SortKey{{Field: "metrics.tokens.main.cache_read"}},
	})
	if err != nil {
		t.Fatalf("QueryStageRuns: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	if rows[0].ID != zeroRun.ID || rows[1].ID != absentRun.ID {
		t.Errorf("order = [%d, %d], want [zero=%d, absent=%d]: absent must sort distinctly from a recorded zero",
			rows[0].ID, rows[1].ID, zeroRun.ID, absentRun.ID)
	}
}

// TestAbsentMetricSortsDistinctlyFromZeroDescending is
// TestAbsentMetricSortsDistinctlyFromZero's DESC counterpart. Postgres's
// own default flips NULLS placement between ASC and DESC (NULLS LAST for
// ASC, NULLS FIRST for DESC); buildOrderSQL overrides that default and
// pins NULLS LAST unconditionally in both directions, per its own doc
// comment. Without a DESC-specific test, a regression that dropped the
// explicit "NULLS LAST" only from the DESC branch would fall back to
// Postgres's default (NULLS FIRST) -- silently inverting the absent-value
// guarantee for every descending sort -- while the ASC test above kept
// passing and the suite stayed green throughout. As in the ASC test,
// absentRun is seeded before zeroRun so it holds the lower id: the
// idColumn tiebreaker is always ASC regardless of this sort's DESC
// direction, so a tie still orders [absent, zero] against the expected
// [zero, absent], and a coerce-to-zero regression is caught here too
// rather than only in the ASC twin.
func TestAbsentMetricSortsDistinctlyFromZeroDescending(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-absent-metric-desc-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	absentIn := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
	absentRun := runStage(t, st, absentIn, json.RawMessage(`{"tokens":{"main":{"cache_creation":5}}}`), absentIn.StartedAt.Add(time.Minute), "completed")

	zeroIn := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	zeroRun := runStage(t, st, zeroIn, json.RawMessage(`{"tokens":{"main":{"cache_read":0}}}`), zeroIn.StartedAt.Add(time.Minute), "completed")

	if absentRun.ID >= zeroRun.ID {
		t.Fatalf("fixture invariant broken: absentRun (id=%d) must be seeded before zeroRun (id=%d) — "+
			"otherwise buildOrderSQL's unconditional idColumn ASC tiebreak can mask a coerce-to-zero regression",
			absentRun.ID, zeroRun.ID)
	}

	rows, _, err := st.QueryStageRuns(ctx, store.Query{
		Sort: []store.SortKey{{Field: "metrics.tokens.main.cache_read", Desc: true}},
	})
	if err != nil {
		t.Fatalf("QueryStageRuns: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	// DESC still orders the recorded zero before the absent run: NULLS
	// LAST means "absent sorts after every recorded value", regardless of
	// which direction "after" points in.
	if rows[0].ID != zeroRun.ID || rows[1].ID != absentRun.ID {
		t.Errorf("DESC order = [%d, %d], want [zero=%d, absent=%d]: absent must sort last (NULLS LAST) in both directions",
			rows[0].ID, rows[1].ID, zeroRun.ID, absentRun.ID)
	}
}

// TestSearchMatchesAcrossIdentityFields embeds one unique, unguessable
// term into exactly one identity field at a time -- change name, Jira
// issue, branch, updated by, command, stage, and repository root -- and
// asserts a search for that term returns exactly the one stage run
// carrying it. Because the term is unique per case and matched nowhere
// else in the (freshly created, per-test) database, a search returning
// anything other than that single row proves either a missing column in
// the search or an over-broad match.
func TestSearchMatchesAcrossIdentityFields(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	unique := func(label string) string {
		return fmt.Sprintf("srch%s%d", label, time.Now().UnixNano())
	}

	assertFoundExactly := func(t *testing.T, term string, wantID int64) {
		t.Helper()
		rows, total, err := st.QueryStageRuns(ctx, store.Query{Search: term})
		if err != nil {
			t.Fatalf("QueryStageRuns search %q: %v", term, err)
		}
		if total != 1 || len(rows) != 1 {
			t.Fatalf("search %q returned %d rows (total %d), want exactly 1", term, len(rows), total)
		}
		if rows[0].ID != wantID {
			t.Errorf("search %q matched run %d, want %d", term, rows[0].ID, wantID)
		}
	}

	t.Run("change name", func(t *testing.T) {
		term := unique("name")
		projectKey := "proj-" + term
		seedChange(t, st, projectKey, term)
		run := runStage(t, st, baseBeginInput(projectKey, term, "/myflow-do", "SDD + TDD per task"), nil, time.Now(), "completed")
		assertFoundExactly(t, term, run.ID)
	})

	t.Run("jira issue", func(t *testing.T) {
		term := unique("jira")
		projectKey := "proj-" + term
		c := baseChange(projectKey, "kan-1")
		c.JiraIssue = ptr(term)
		if err := st.PutChange(ctx, c); err != nil {
			t.Fatalf("PutChange: %v", err)
		}
		run := runStage(t, st, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"), nil, time.Now(), "completed")
		assertFoundExactly(t, term, run.ID)
	})

	t.Run("branch", func(t *testing.T) {
		term := unique("branch")
		projectKey := "proj-" + term
		c := baseChange(projectKey, "kan-1")
		c.Branch = ptr(term)
		if err := st.PutChange(ctx, c); err != nil {
			t.Fatalf("PutChange: %v", err)
		}
		run := runStage(t, st, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"), nil, time.Now(), "completed")
		assertFoundExactly(t, term, run.ID)
	})

	t.Run("updated by", func(t *testing.T) {
		term := unique("updby")
		projectKey := "proj-" + term
		c := baseChange(projectKey, "kan-1")
		c.UpdatedBy = term
		if err := st.PutChange(ctx, c); err != nil {
			t.Fatalf("PutChange: %v", err)
		}
		run := runStage(t, st, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"), nil, time.Now(), "completed")
		assertFoundExactly(t, term, run.ID)
	})

	t.Run("command", func(t *testing.T) {
		term := unique("cmd")
		projectKey := "proj-" + term
		seedChange(t, st, projectKey, "kan-1")
		run := runStage(t, st, baseBeginInput(projectKey, "kan-1", term, "SDD + TDD per task"), nil, time.Now(), "completed")
		assertFoundExactly(t, term, run.ID)
	})

	t.Run("stage", func(t *testing.T) {
		term := unique("stage")
		projectKey := "proj-" + term
		seedChange(t, st, projectKey, "kan-1")
		run := runStage(t, st, baseBeginInput(projectKey, "kan-1", "/myflow-do", term), nil, time.Now(), "completed")
		assertFoundExactly(t, term, run.ID)
	})

	t.Run("repository root", func(t *testing.T) {
		term := unique("repo")
		projectKey := "proj-" + term
		c := baseChange(projectKey, "kan-1")
		c.Repos = []store.Repo{{RepoRoot: term}}
		if err := st.PutChange(ctx, c); err != nil {
			t.Fatalf("PutChange: %v", err)
		}
		in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
		in.RepoRoot = ptr(term)
		run := runStage(t, st, in, nil, time.Now(), "completed")
		assertFoundExactly(t, term, run.ID)
	})
}

// TestSearchTermMetacharactersAreEscaped asserts that "_" and "%" in a
// search term match themselves literally rather than acting as ILIKE
// wildcards (any-one-character and any-run-of-characters, respectively).
// Each case seeds a "decoy" row that a real record would only match if the
// metacharacter were left unescaped, and asserts the search returns
// exactly the literal match and not the decoy -- a test that only checked
// "the literal match is found" would still pass against an unescaped
// implementation, since an unescaped "_" or "%" matches its literal
// character too, just not *only* its literal character.
func TestSearchTermMetacharactersAreEscaped(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-search-escape-%d", time.Now().UnixNano())

	seedBranch := func(t *testing.T, name, branch string) {
		t.Helper()
		c := baseChange(projectKey, name)
		c.Branch = ptr(branch)
		if err := st.PutChange(ctx, c); err != nil {
			t.Fatalf("PutChange %s: %v", name, err)
		}
	}

	assertOnlyMatch := func(t *testing.T, term, wantName string) {
		t.Helper()
		rows, total, err := st.QueryChanges(ctx, store.Query{
			Filters: []store.Filter{{Field: "project", Op: store.OpEq, Value: projectKey}},
			Search:  term,
		})
		if err != nil {
			t.Fatalf("QueryChanges search %q: %v", term, err)
		}
		if total != 1 || len(rows) != 1 {
			t.Fatalf("search %q got %d rows (total %d), want exactly 1", term, len(rows), total)
		}
		if rows[0].Name != wantName {
			t.Errorf("search %q matched %q, want %q", term, rows[0].Name, wantName)
		}
	}

	// "_" matches any single character when unescaped, so an unescaped
	// search for "feature_x" would also match "featureXx".
	seedBranch(t, "kan-underscore", "feature_x")
	seedBranch(t, "kan-decoy-underscore", "featureXx")
	assertOnlyMatch(t, "feature_x", "kan-underscore")

	// "%" matches any run of characters when unescaped, so an unescaped
	// search for "50%done" would also match "50XXXdone".
	seedBranch(t, "kan-percent", "50%done")
	seedBranch(t, "kan-decoy-percent", "50XXXdone")
	assertOnlyMatch(t, "50%done", "kan-percent")

	// A literal backslash in the term must survive escaping without
	// breaking the ESCAPE clause itself. The seeded change's name is
	// deliberately unrelated to the search term -- naming it e.g.
	// "kan-backslash" would let the assertion pass off the *name* column
	// alone: Postgres's ESCAPE grammar treats an escape character before an
	// ordinary character permissively (escape+'s' reads as a literal 's',
	// dropping the backslash), so 'kan-backslash' ILIKE '%back\slash%'
	// ESCAPE '\' is true even with no real backslash character anywhere --
	// proven directly against Postgres. With an unrelated name, the only
	// way this assertion can pass is if the branch's actual backslash
	// character survived escaping and matched.
	seedBranch(t, "kan-esc-check", `back\slash`)
	assertOnlyMatch(t, `back\slash`, "kan-esc-check")
}

// TestMetricsKeyPathSizeIsBounded asserts the size bounds
// validateMetricsKeyPath enforces alongside its syntax rule: a
// syntactically valid path is still rejected once it has more segments,
// or a longer segment, than the fixed caps allow.
func TestMetricsKeyPathSizeIsBounded(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	tooManySegments := "metrics." + strings.Repeat("a.", 20) + "z" // 21 segments
	_, _, err := st.QueryStageRuns(ctx, store.Query{Sort: []store.SortKey{{Field: tooManySegments}}})
	if !errors.Is(err, store.ErrInvalidMetricsKeyPath) {
		t.Fatalf("sort by an over-long path (21 segments): got %v, want ErrInvalidMetricsKeyPath", err)
	}

	tooLongSegment := "metrics." + strings.Repeat("a", 200)
	_, _, err = st.QueryStageRuns(ctx, store.Query{Sort: []store.SortKey{{Field: tooLongSegment}}})
	if !errors.Is(err, store.ErrInvalidMetricsKeyPath) {
		t.Fatalf("sort by a 200-character segment: got %v, want ErrInvalidMetricsKeyPath", err)
	}

	// A path right at the documented shape a real metrics key uses --
	// "metrics.tokens.main.cache_read", three segments of ordinary length,
	// the wire shape internal/harvest.TokenDelta actually writes -- must
	// still work: the bound must not have crept below anything this
	// package's own vocabulary needs.
	if _, _, err := st.QueryStageRuns(ctx, store.Query{Sort: []store.SortKey{{Field: "metrics.tokens.main.cache_read"}}}); err != nil {
		t.Errorf("sort by metrics.tokens.main.cache_read: %v, want no error", err)
	}
}

// TestNoLimitWithNonZeroOffsetIsRejected asserts Query.Limit == NoLimit
// combined with a non-zero Offset is a rejected request rather than a
// silently ignored one: NoLimit already means "no LIMIT clause, no
// paging", so an Offset alongside it can never take effect, and letting it
// through without a word would hide a caller's mistaken assumption that
// paging still applied.
func TestNoLimitWithNonZeroOffsetIsRejected(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	_, _, err := st.QueryChanges(ctx, store.Query{Limit: store.NoLimit, Offset: 5})
	if !errors.Is(err, store.ErrOffsetWithNoLimit) {
		t.Fatalf("QueryChanges with NoLimit and Offset=5: got %v, want ErrOffsetWithNoLimit", err)
	}

	_, _, err = st.QueryStageRuns(ctx, store.Query{Limit: store.NoLimit, Offset: 5})
	if !errors.Is(err, store.ErrOffsetWithNoLimit) {
		t.Fatalf("QueryStageRuns with NoLimit and Offset=5: got %v, want ErrOffsetWithNoLimit", err)
	}

	// NoLimit with the zero-value Offset is exactly ListChanges' own usage
	// and must still work.
	if _, _, err := st.QueryChanges(ctx, store.Query{Limit: store.NoLimit}); err != nil {
		t.Errorf("QueryChanges with NoLimit and Offset=0: %v, want no error", err)
	}
}

// TestQueryTxOptionsUsesRepeatableRead is the actual guard on this
// package's production wiring: it asserts queryTxOptions()'s return value
// directly, field by field, rather than trying to observe its effect by
// racing a concurrent writer against QueryChanges' or QueryStageRuns' own
// count-then-page window -- a few microseconds, so that race would be
// probabilistic and the test would be flaky for no better guarantee. This
// is deterministic instead: since both QueryChanges and QueryStageRuns
// call queryTxOptions() rather than constructing their own pgx.TxOptions
// inline, a change to -- or accidental drop of -- the isolation level in
// production fails this test immediately, every run, with no timing
// dependency at all.
//
// TestRepeatableReadPreventsCountDrift, below, is the complementary half:
// it proves REPEATABLE READ actually has the effect this test's value
// claims it does. Neither proves the other's claim; both are needed.
func TestQueryTxOptionsUsesRepeatableRead(t *testing.T) {
	want := pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}
	got := store.QueryTxOptionsForTest()
	if got != want {
		t.Fatalf("queryTxOptions() = %+v, want %+v -- QueryChanges and QueryStageRuns rely on exactly this isolation level and access mode for count/page consistency", got, want)
	}
}

// TestRepeatableReadPreventsCountDrift proves the Postgres mechanism
// queryTxOptions() relies on -- REPEATABLE READ fixing a transaction's
// snapshot once, at its first statement, rather than per statement -- does
// what its name promises. It is *not* a guard on this package's own
// wiring: it hand-opens its own transaction with a parameterised isolation
// level rather than calling QueryChanges or QueryStageRuns, so reverting
// IsoLevel: pgx.RepeatableRead out of both production BeginTx calls does
// not make this test fail (confirmed: it still passes with that reverted).
// TestQueryTxOptionsUsesRepeatableRead, above, is what catches that
// regression -- this test exists only to establish that the isolation
// level it asserts on actually behaves as documented, at READ COMMITTED
// (the pool's session default), each statement inside an open transaction
// takes its *own* snapshot, so a commit landing between two statements on
// the same still-open transaction is visible to the second one. Only
// REPEATABLE READ fixes the snapshot once, at the transaction's first
// statement, so every later statement on it is blind to anything
// committed afterward.
//
// This does not go through QueryChanges itself: QueryChanges' own count
// and page statements are microseconds apart, so racing a concurrent
// writer against that window from outside would be probabilistic, not a
// reliable regression test. Instead it opens a transaction with the same
// shape and isolation level QueryChanges uses, runs the same COUNT query
// twice on it with a real committed write from a second connection in
// between, and checks whether the second read saw that write -- a
// deterministic reproduction of the mechanism, not a race against timing.
func TestRepeatableReadPreventsCountDrift(t *testing.T) {
	dsn := newTestDatabase(t)
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	st, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()
	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	projectKey := fmt.Sprintf("proj-isolation-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	// countBeforeAndAfterConcurrentInsert opens a transaction at iso,
	// counts changes in projectKey, has a second connection (st, its own
	// pool) insert and commit one more matching change while the
	// transaction is still open, then counts again on that same
	// transaction.
	countBeforeAndAfterConcurrentInsert := func(t *testing.T, iso pgx.TxIsoLevel, insertName string) (before, after int) {
		t.Helper()

		tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: iso, AccessMode: pgx.ReadOnly})
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()

		countSQL := "SELECT COUNT(*) FROM changes WHERE project_key = $1"
		if err := tx.QueryRow(ctx, countSQL, projectKey).Scan(&before); err != nil {
			t.Fatalf("count before: %v", err)
		}

		if err := st.PutChange(ctx, baseChange(projectKey, insertName)); err != nil {
			t.Fatalf("concurrent PutChange: %v", err)
		}

		if err := tx.QueryRow(ctx, countSQL, projectKey).Scan(&after); err != nil {
			t.Fatalf("count after: %v", err)
		}
		return before, after
	}

	t.Run("READ COMMITTED sees the concurrent commit", func(t *testing.T) {
		before, after := countBeforeAndAfterConcurrentInsert(t, pgx.ReadCommitted, "kan-rc")
		if before == after {
			t.Fatalf("READ COMMITTED before=%d after=%d, want them to differ -- this subtest documents the bug the fix closes, and it failing means the drift is gone even at READ COMMITTED, which would be surprising", before, after)
		}
	})

	t.Run("REPEATABLE READ is blind to the concurrent commit", func(t *testing.T) {
		before, after := countBeforeAndAfterConcurrentInsert(t, pgx.RepeatableRead, "kan-rr")
		if before != after {
			t.Fatalf("REPEATABLE READ before=%d after=%d, want them equal: a transaction at this isolation level must not see a commit that landed after it began", before, after)
		}
	})
}

// TestPaginationIsStableUnderEqualSortKeys seeds several stage runs that
// all share the same value for the sorted field (outcome), so their
// relative order depends entirely on the id tiebreaker buildOrderSQL
// always appends. Paging through with a small page size must visit every
// row exactly once -- neither a duplicate (the tiebreaker missing or
// non-unique) nor a gap (an off-by-one in LIMIT/OFFSET).
func TestPaginationIsStableUnderEqualSortKeys(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-page-stable-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	const n = 7
	want := make(map[int64]bool, n)
	for i := 0; i < n; i++ {
		in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
		run := runStage(t, st, in, nil, in.StartedAt.Add(time.Minute), "completed")
		want[run.ID] = true
	}

	seen := make(map[int64]bool, n)
	const pageSize = 3
	for offset := 0; ; offset += pageSize {
		rows, total, err := st.QueryStageRuns(ctx, store.Query{
			Filters: []store.Filter{{Field: "project", Op: store.OpEq, Value: projectKey}},
			Sort:    []store.SortKey{{Field: "outcome"}},
			Limit:   pageSize,
			Offset:  offset,
		})
		if err != nil {
			t.Fatalf("QueryStageRuns offset %d: %v", offset, err)
		}
		if total != n {
			t.Fatalf("total = %d, want %d", total, n)
		}
		if len(rows) == 0 {
			break
		}
		for _, r := range rows {
			if seen[r.ID] {
				t.Fatalf("stage run %d returned on two pages", r.ID)
			}
			seen[r.ID] = true
		}
	}

	if len(seen) != n {
		t.Fatalf("paged through %d distinct rows, want %d", len(seen), n)
	}
	for id := range want {
		if !seen[id] {
			t.Errorf("stage run %d never appeared on any page", id)
		}
	}
}

// TestPageBoundaryLosesNoRow exercises the same guarantee as
// TestPaginationIsStableUnderEqualSortKeys from the other query type
// (changes) and with a row count that does not divide evenly by the page
// size, so the last page is a partial page -- the boundary most likely to
// lose or duplicate a row in an off-by-one.
func TestPageBoundaryLosesNoRow(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-page-boundary-%d", time.Now().UnixNano())

	const n = 5
	want := make(map[string]bool, n)
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("kan-%d", i)
		if err := st.PutChange(ctx, baseChange(projectKey, name)); err != nil {
			t.Fatalf("PutChange %s: %v", name, err)
		}
		want[name] = true
	}

	seen := make(map[string]bool, n)
	const pageSize = 2
	for offset := 0; ; offset += pageSize {
		rows, total, err := st.QueryChanges(ctx, store.Query{
			Filters: []store.Filter{{Field: "project", Op: store.OpEq, Value: projectKey}},
			Sort:    []store.SortKey{{Field: "state"}},
			Limit:   pageSize,
			Offset:  offset,
		})
		if err != nil {
			t.Fatalf("QueryChanges offset %d: %v", offset, err)
		}
		if total != n {
			t.Fatalf("total = %d, want %d", total, n)
		}
		if len(rows) == 0 {
			break
		}
		for _, c := range rows {
			if seen[c.Name] {
				t.Fatalf("change %q returned on two pages", c.Name)
			}
			seen[c.Name] = true
		}
	}

	if len(seen) != n {
		t.Fatalf("paged through %d distinct rows, want %d", len(seen), n)
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("change %q never appeared on any page", name)
		}
	}
}
