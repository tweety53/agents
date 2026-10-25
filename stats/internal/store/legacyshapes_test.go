package store_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestMigration0031RewritesLegacyRows seeds every row shape 0031 rewrites --
// a synthetic change row stamped by the retired "myflow" writer, bare-array
// "panel.dispatches" and "groups" elements -- next to current-shape rows of
// each kind, plus pricing rows with a null 1h rate whose collapsed column
// equals the 5m rate (the shape 0007 left on every pre-0007 row), on a
// database migrated through 0030 only, then applies 0031 and asserts the
// rewritten values, the untouched rows byte-identical (a null 1h rate stays
// null: 0031 backfills no rate), and the collapsed pricing column gone.
//
// The migrator applies every embedded file it has not recorded, so 0031 is
// held back by recording it as applied before the first run and releasing
// it after seeding.
// ponytail: any migration numbered after 0031 also runs before seeding; add a
// prefix-applying migrator hook if one of them ever reshapes these tables.
func TestMigration0031RewritesLegacyRows(t *testing.T) {
	const held = "0031_drop_legacy_shapes.sql"
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := newTestDatabase(t)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("open raw pool: %v", err)
	}
	t.Cleanup(pool.Close)
	st, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(st.Close)

	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("exec %q: %v", sql, err)
		}
	}
	text := func(sql string, args ...any) string {
		t.Helper()
		var s *string
		if err := pool.QueryRow(ctx, sql, args...).Scan(&s); err != nil {
			t.Fatalf("query %q: %v", sql, err)
		}
		if s == nil {
			return "<null>"
		}
		return *s
	}

	exec(`CREATE TABLE schema_migrations (filename TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`)
	exec(`INSERT INTO schema_migrations (filename) VALUES ($1)`, held)
	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("migrate through 0030: %v", err)
	}

	exec(`INSERT INTO projects (project_key, main_checkout_path) VALUES ('p', '/tmp/p')`)
	exec(`INSERT INTO changes (project_key, name, state, updated_at, updated_by) VALUES
		('p', 'synthetic', 'STARTED', '2026-09-01T00:00:00Z', 'myflow stage begin (synthetic)'),
		('p', 'current', 'IN_PROGRESS', '2026-09-02T00:00:00Z', '/flow')`)
	exec(`INSERT INTO decisions (change_id, session_token, decision)
		SELECT id, tok, doc::jsonb FROM changes, (VALUES
			('mixed', '{"panel": {"grouping": "free", "dispatches": [["primary", "principles"], {"slots": ["bugbot"], "model": "opus", "effort": "high"}]}, "groups": null}'),
			('groups', '{"panel": "default", "groups": [[1, 2], [3]], "implementer": "skipped — inline"}'),
			('current', '{"panel": {"dispatches": [{"slots": ["primary"], "model": "opus", "effort": "high"}]}, "groups": [{"bundles": [1], "model": "opus", "effort": "high"}]}')
		) AS v(tok, doc) WHERE name = 'current'`)
	exec(`INSERT INTO pricing (model, effective_from, input_per_mtok, output_per_mtok,
			cache_write_per_mtok, cache_write_5m_per_mtok, cache_write_1h_per_mtok, cache_read_per_mtok)
		VALUES
		('flat', '2026-01-01T00:00:00Z', 0.075, 0.25, 0, 0, NULL, 0.015),
		('pre-0007', '2025-06-01T00:00:00Z', 3, 15, 3.75, 3.75, NULL, 0.3),
		('unset-split', '2026-01-01T00:00:00Z', 1, 5, 1.25, 0, NULL, 0.1),
		('split', '2026-01-01T00:00:00Z', 5, 25, 6.25, 6.25, 10, 0.5)`)

	changeRow := `SELECT (to_jsonb(c) - 'updated_by')::text FROM changes c WHERE name = $1`
	changeBy := `SELECT updated_by FROM changes WHERE name = $1`
	decisionDoc := `SELECT decision::text FROM decisions WHERE session_token = $1`
	pricingRow := `SELECT (to_jsonb(p) - 'cache_write_per_mtok' - 'cache_write_1h_per_mtok')::text FROM pricing p WHERE model = $1`
	pricing1h := `SELECT cache_write_1h_per_mtok::text FROM pricing WHERE model = $1`

	beforeChange := map[string]string{"synthetic": text(changeRow, "synthetic"), "current": text(changeRow, "current")}
	beforeCurrentDecision := text(decisionDoc, "current")
	beforePricing := map[string]string{}
	for _, m := range []string{"flat", "pre-0007", "unset-split", "split"} {
		beforePricing[m] = text(pricingRow, m)
	}

	exec(`DELETE FROM schema_migrations WHERE filename = $1`, held)
	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("apply 0031: %v", err)
	}

	for name, want := range map[string]string{"synthetic": "flow stage begin (synthetic)", "current": "/flow"} {
		if got := text(changeBy, name); got != want {
			t.Errorf("changes %q updated_by = %q, want %q", name, got, want)
		}
		if got := text(changeRow, name); got != beforeChange[name] {
			t.Errorf("changes %q other columns changed:\n before %s\n after  %s", name, beforeChange[name], got)
		}
	}

	wantDecision := map[string]string{
		"mixed":  `{"panel": {"grouping": "free", "dispatches": [{"slots": ["primary", "principles"]}, {"slots": ["bugbot"], "model": "opus", "effort": "high"}]}, "groups": null}`,
		"groups": `{"panel": "default", "groups": [{"bundles": [1, 2]}, {"bundles": [3]}], "implementer": "skipped — inline"}`,
	}
	for tok, want := range wantDecision {
		wantCanonical := text(`SELECT $1::jsonb::text`, want)
		if got := text(decisionDoc, tok); got != wantCanonical {
			t.Errorf("decision %q =\n %s\nwant\n %s", tok, got, wantCanonical)
		}
	}
	if got := text(decisionDoc, "current"); got != beforeCurrentDecision {
		t.Errorf("current-shape decision rewritten:\n before %s\n after  %s", beforeCurrentDecision, got)
	}

	for model, want := range map[string]string{"flat": "<null>", "pre-0007": "<null>", "unset-split": "<null>", "split": "10"} {
		if got := text(pricing1h, model); got != want {
			t.Errorf("pricing %q cache_write_1h_per_mtok = %s, want %s", model, got, want)
		}
		if got := text(pricingRow, model); got != beforePricing[model] {
			t.Errorf("pricing %q other columns changed:\n before %s\n after  %s", model, beforePricing[model], got)
		}
	}

	var collapsed int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM information_schema.columns
		WHERE table_name = 'pricing' AND column_name = 'cache_write_per_mtok'`).Scan(&collapsed); err != nil {
		t.Fatalf("query information_schema: %v", err)
	}
	if collapsed != 0 {
		t.Errorf("pricing.cache_write_per_mtok still present after 0031")
	}
}

// TestMigration0047DropsLoopDispatchTwins pins kan-1032's self-review: 0047 deletes only an
// agentless pipeline-fix row whose same-key twin sits under its own loop change of the same
// project, nulls the finding pointing at it, and is a no-op on a second run.
func TestMigration0047DropsLoopDispatchTwins(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := newTestDatabase(t)
	st, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(st.Close)
	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("open raw pool: %v", err)
	}
	t.Cleanup(pool.Close)
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("exec %q: %v", sql, err)
		}
	}

	exec(`INSERT INTO projects (project_key, main_checkout_path) VALUES ('p', '/tmp/p')`)
	exec(`INSERT INTO changes (project_key, name, state, updated_at, updated_by) VALUES
		('p', 'kan-1', 'IN_PROGRESS', now(), '/flow'),
		('p', 'pipeline-fix-kan-1-7', 'IN_PROGRESS', now(), '/flow'),
		('p', 'kan-2', 'IN_PROGRESS', now(), '/flow'),
		('p', 'kan-3', 'IN_PROGRESS', now(), '/flow'),
		('p', 'pipeline-fix-kan-3-foo-4', 'IN_PROGRESS', now(), '/flow')`)
	exec(`INSERT INTO dispatches (change_id, seq, role, model, started_at, dispatch_key, agent_id)
		SELECT c.id, v.seq, 'planner', 'opus', now(), v.k, v.a FROM changes c JOIN (VALUES
			('kan-1', 1, 'pipeline-fix-7-plan', NULL),
			('kan-1', 2, 'pipeline-fix-8-plan', NULL),
			('kan-1', 3, 'pipeline-fix-7-fix', 'a3'),
			('pipeline-fix-kan-1-7', 1, 'pipeline-fix-7-plan', 'a1'),
			('pipeline-fix-kan-1-7', 2, 'pipeline-fix-7-fix', 'a2'),
			('kan-2', 1, 'pipeline-fix-7-plan', NULL),
			('kan-3', 1, 'pipeline-fix-4-plan', NULL),
			('pipeline-fix-kan-3-foo-4', 1, 'pipeline-fix-4-plan', 'a4')
		) AS v(ch, seq, k, a) ON c.name = v.ch`)
	exec(`INSERT INTO findings (change_id, dispatch_id, ref, round, slot, severity, note, status)
		SELECT d.change_id, d.id, 'F1', 1, 'primary', 'Minor', 'n', 'open'
		FROM dispatches d JOIN changes c ON c.id = d.change_id
		WHERE c.name = 'kan-1' AND d.dispatch_key = 'pipeline-fix-7-plan'`)

	sql, err := os.ReadFile("migrations/0047_drop_loop_dispatch_twins.sql")
	if err != nil {
		t.Fatal(err)
	}
	exec(string(sql))
	exec(string(sql))

	rows, err := pool.Query(ctx, `SELECT c.name || '/' || d.dispatch_key FROM dispatches d
		JOIN changes c ON c.id = d.change_id ORDER BY 1`)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		got = append(got, s)
	}
	rows.Close()
	want := "kan-1/pipeline-fix-7-fix kan-1/pipeline-fix-8-plan kan-2/pipeline-fix-7-plan kan-3/pipeline-fix-4-plan pipeline-fix-kan-1-7/pipeline-fix-7-fix pipeline-fix-kan-1-7/pipeline-fix-7-plan pipeline-fix-kan-3-foo-4/pipeline-fix-4-plan"
	if g := strings.Join(got, " "); g != want {
		t.Errorf("rows =\n %s\nwant\n %s", g, want)
	}
	var dispatchID *int64
	if err := pool.QueryRow(ctx, `SELECT dispatch_id FROM findings WHERE ref = 'F1'`).Scan(&dispatchID); err != nil {
		t.Fatal(err)
	}
	if dispatchID != nil {
		t.Errorf("finding dispatch_id = %d, want NULL", *dispatchID)
	}
}
