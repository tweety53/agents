package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/harvest"
	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// newRecordStore returns a freshly-migrated store together with a raw pool
// onto the same database. It composes the package's existing fixture
// (newTestDatabase, store.Open, RunMigrations) rather than introducing a
// second one -- the same shape TestSupersedeIndexExists already builds
// inline, factored out here because three of this file's cases need it.
//
// The raw pool exists because *store.Store deliberately exposes no
// raw-query escape hatch (store.go's package doc, "the only package in this
// module that builds SQL"), and three facts these tests assert are not
// reachable through the typed API at all: what schema_migrations recorded,
// which constraint a duplicate ref violates, and what a finding's own
// dispatch_id column holds -- the last of which the wire type reports back
// as a seq, never as the row id it is stored under.
func newRecordStore(t *testing.T) (*store.Store, *pgxpool.Pool) {
	t.Helper()

	dsn := newTestDatabase(t)
	ctx := context.Background()

	st, err := store.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(st.Close)
	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect raw pool: %v", err)
	}
	t.Cleanup(pool.Close)

	return st, pool
}

// baseDispatch is the minimal valid dispatch every case in this file
// starts from: the columns the schema marks NOT NULL and nothing else, so
// a case that cares about task, slot or commit sets only what it is about.
// Seq is deliberately left zero -- RecordDispatch allocates it, and a
// caller-supplied value is ignored.
func baseDispatch(role, model string) records.Dispatch {
	return records.Dispatch{
		Role:      role,
		Model:     model,
		StartedAt: time.Date(2026, 8, 22, 9, 0, 0, 0, time.UTC),
	}
}

// baseFinding is the minimal valid finding: every NOT NULL column of the
// findings table, with the optional location and reproducer left unset.
func baseFinding(ref string, round int) records.Finding {
	return records.Finding{
		Ref:      ref,
		Round:    round,
		Slot:     "principles",
		Severity: "major",
		Note:     "the handler swallows the decode error",
		Status:   "open",
	}
}

// TestRunRecordsMigrationAppliesTwiceIdempotently pins that this change
// adds exactly one migration file and that applying it twice records it
// once -- the same guarantee TestMigrationsAreIdempotent makes for the set
// as a whole, asserted here for the file this change adds so that a later
// edit splitting it in two, or making its DDL non-repeatable, fails at the
// migration rather than at whichever method first touches a half-created
// table.
//
// The count is asserted as "exactly one filename beginning 0010_" rather
// than as a literal total, so migration 0011 does not have to edit this
// test to add a table this change knows nothing about.
func TestRunRecordsMigrationAppliesTwiceIdempotently(t *testing.T) {
	st, pool := newRecordStore(t)
	ctx := context.Background()

	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("second RunMigrations call: %v", err)
	}

	var recorded int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM schema_migrations WHERE filename LIKE '0010\\_%'",
	).Scan(&recorded); err != nil {
		t.Fatalf("count schema_migrations rows for this change's migration: %v", err)
	}
	if recorded != 1 {
		t.Errorf("schema_migrations holds %d rows for 0010_*, want exactly 1 after two runs", recorded)
	}

	total, err := st.MigrationCount(ctx)
	if err != nil {
		t.Fatalf("count schema_migrations rows: %v", err)
	}
	want, err := store.EmbeddedMigrationCount()
	if err != nil {
		t.Fatalf("count embedded migration files: %v", err)
	}
	if total != want {
		t.Errorf("schema_migrations has %d rows after two runs, want %d (one per embedded migration file)", total, want)
	}

	for _, table := range []string{"dispatches", "findings"} {
		var exists bool
		if err := pool.QueryRow(ctx,
			"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)", table,
		).Scan(&exists); err != nil {
			t.Fatalf("check for table %s: %v", table, err)
		}
		if !exists {
			t.Errorf("table %s does not exist after the migration ran", table)
		}
	}
}

// TestRecordDispatchAllocatesSeqPerChange asserts the delta spec's "the
// append order SHALL be unique per change": the series restarts for a
// second change rather than being one global counter, so a change's own
// record reads 1, 2, 3 no matter what else the store holds.
func TestRecordDispatchAllocatesSeqPerChange(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-record-seq-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	seedChange(t, st, projectKey, "kan-2")

	first, err := st.RecordDispatch(ctx, projectKey, "kan-1", baseDispatch("implementer", "opus"))
	if err != nil {
		t.Fatalf("first RecordDispatch: %v", err)
	}
	if first.Seq != 1 {
		t.Errorf("first RecordDispatch seq = %d, want 1", first.Seq)
	}

	second, err := st.RecordDispatch(ctx, projectKey, "kan-1", baseDispatch("reviewer", "sonnet"))
	if err != nil {
		t.Fatalf("second RecordDispatch: %v", err)
	}
	if second.Seq != 2 {
		t.Errorf("second RecordDispatch seq = %d, want 2", second.Seq)
	}

	other, err := st.RecordDispatch(ctx, projectKey, "kan-2", baseDispatch("implementer", "opus"))
	if err != nil {
		t.Fatalf("RecordDispatch for a second change: %v", err)
	}
	if other.Seq != 1 {
		t.Errorf("RecordDispatch seq for a second change = %d, want 1 (seq is per change, not global)", other.Seq)
	}

	// Every recorded field comes back on the row the caller is handed, so
	// a caller never has to re-read to learn what it just wrote.
	in := baseDispatch("panel-fix", "unknown (agent-defined)")
	in.TaskID = "3.1"
	in.Slot = "principles"
	in.CommitSHA = "abc1234"
	in.Outcome = "completed"
	in.SessionToken = "mf-record-dispatch-round-trip"
	in.Notes = "second fix round"
	got, err := st.RecordDispatch(ctx, projectKey, "kan-1", in)
	if err != nil {
		t.Fatalf("RecordDispatch with every optional field: %v", err)
	}
	if got.Seq != 3 {
		t.Errorf("seq = %d, want 3", got.Seq)
	}
	if got.TaskID != in.TaskID || got.Slot != in.Slot || got.Role != in.Role ||
		got.CommitSHA != in.CommitSHA || got.Outcome != in.Outcome ||
		got.SessionToken != in.SessionToken || got.Notes != in.Notes {
		t.Errorf("RecordDispatch returned %+v, want every field of %+v round-tripped", got, in)
	}
	// The model is recorded intent, stored verbatim: a slot whose model the
	// dispatcher cannot read records this literal, and the store must never
	// normalise it into something that looks like a real slug.
	if got.Model != "unknown (agent-defined)" {
		t.Errorf("Model = %q, want the literal %q stored verbatim", got.Model, "unknown (agent-defined)")
	}
	if !got.StartedAt.Equal(in.StartedAt) {
		t.Errorf("StartedAt = %v, want %v", got.StartedAt, in.StartedAt)
	}
}

// TestConcurrentRecordDispatchDoesNotCollide is the case that stops
// RecordDispatch being written as a read-then-insert. Sixteen writers race
// for one change; the delta spec requires that both rows exist with
// distinct append positions and neither has overwritten the other, so
// sixteen rows must carry sixteen distinct seq values -- the same shape
// TestConcurrentBeginStageDoesNotCollide already asserts for attempt
// allocation.
func TestConcurrentRecordDispatchDoesNotCollide(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-record-concurrent-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	const writers = 16

	var wg sync.WaitGroup
	seqs := make([]int, writers)
	errs := make([]error, writers)

	for i := range writers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			callCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			got, err := st.RecordDispatch(callCtx, projectKey, "kan-1", baseDispatch("implementer", "opus"))
			errs[i] = err
			if err == nil {
				seqs[i] = got.Seq
			}
		}(i)
	}
	wg.Wait()

	seen := make(map[int]int)
	for i, err := range errs {
		if err != nil {
			t.Fatalf("writer %d: RecordDispatch: %v", i, err)
		}
		seen[seqs[i]]++
	}
	for seq := 1; seq <= writers; seq++ {
		if seen[seq] != 1 {
			t.Errorf("seq %d was allocated %d times, want exactly 1", seq, seen[seq])
		}
	}
	if len(seen) != writers {
		t.Errorf("got %d distinct seq values, want %d", len(seen), writers)
	}

	rec, err := st.RunRecord(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("RunRecord: %v", err)
	}
	if len(rec.Dispatches) != writers {
		t.Errorf("RunRecord returned %d dispatches, want %d -- a lost race overwrote a row", len(rec.Dispatches), writers)
	}
}

// TestUpsertFindingRefusesADuplicateRef asserts the delta spec's "a
// finding's reference SHALL be unique per change, not per round": a second
// write for F1 updates the row rather than appending a second one, so the
// record of a change's findings is never cumulative.
//
// The second half asserts the constraint name itself, through a raw insert
// that bypasses the upsert. UpsertFinding keys its ON CONFLICT on that
// exact name rather than on a column list, so the name is part of this
// change's contract with Postgres, not an incidental identifier Postgres
// picked -- an edit renaming it would otherwise fail nowhere until a real
// concurrent write lost a row.
func TestUpsertFindingRefusesADuplicateRef(t *testing.T) {
	st, pool := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-record-finding-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	_, created, err := st.UpsertFinding(ctx, projectKey, "kan-1", baseFinding("F1", 0))
	if err != nil {
		t.Fatalf("first UpsertFinding: %v", err)
	}
	if !created {
		t.Error("first UpsertFinding reported the row as updated, want created -- the API answers 201 on this")
	}

	second := baseFinding("F1", 1)
	second.Note = "restated after the fix round re-read the handler"
	second.Severity = "minor"
	got, created, err := st.UpsertFinding(ctx, projectKey, "kan-1", second)
	if err != nil {
		t.Fatalf("second UpsertFinding for the same ref: %v", err)
	}
	if created {
		t.Error("second UpsertFinding for the same ref reported the row as created, want updated -- the API answers 200 on this")
	}
	if got.Note != second.Note || got.Severity != second.Severity || got.Round != second.Round {
		t.Errorf("UpsertFinding returned %+v, want the second write's own values %+v", got, second)
	}

	rec, err := st.RunRecord(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("RunRecord: %v", err)
	}
	if len(rec.Findings) != 1 {
		t.Fatalf("RunRecord returned %d findings, want exactly 1 -- the second write appended instead of updating", len(rec.Findings))
	}
	if rec.Findings[0].Note != second.Note {
		t.Errorf("stored note = %q, want the second write's %q", rec.Findings[0].Note, second.Note)
	}

	var changeID int64
	if err := pool.QueryRow(ctx,
		"SELECT id FROM changes WHERE project_key = $1 AND name = $2", projectKey, "kan-1",
	).Scan(&changeID); err != nil {
		t.Fatalf("look up change id: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO findings (change_id, ref, round, slot, severity, note, status)
		VALUES ($1, 'F1', 2, 'principles', 'major', 'a raw insert bypassing the upsert', 'open')
	`, changeID)
	if err == nil {
		t.Fatal("a raw insert of a duplicate ref succeeded, want a unique violation")
	}
	if !strings.Contains(err.Error(), "findings_ref_key") {
		t.Errorf("duplicate-ref error = %v, want it to name the constraint findings_ref_key", err)
	}
}

// warmUp opens the store pool's connections before a concurrency case
// measures anything, by firing n concurrent reads and waiting for them.
// A pool that has not yet dialled serialises the writers behind
// connection setup, which is exactly the interleaving a snapshot race
// needs and does not get.
func warmUp(t *testing.T, st *store.Store, projectKey string, n int) {
	t.Helper()

	var wg sync.WaitGroup
	ready := make(chan struct{})
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ready
			if _, err := st.RunRecord(context.Background(), projectKey, "kan-1"); err != nil {
				t.Errorf("warm-up RunRecord: %v", err)
			}
		}()
	}
	close(ready)
	wg.Wait()
}

// TestConcurrentUpsertFindingReportsCreatedExactlyOnce is the case that
// stops the created flag being read from a snapshot the conflict
// resolution never re-evaluates. Twenty writers race to be the first to
// record F1 for a change that holds no finding at all; exactly one of them
// inserted the row, so exactly one may report created.
//
// A statement-start CTE lookup fails this: every writer whose snapshot
// predates the winning insert sees no row and reports created, so the API
// answers 201 Created several times for one ref -- the exact ambiguity the
// flag exists to remove. RETURNING (xmax = 0) is per-row and evaluated at
// write time instead, so it reports what conflict resolution actually did.
//
// It is the same shape TestConcurrentRecordDispatchDoesNotCollide uses for
// seq allocation, against a real database because the race is Postgres's
// snapshot semantics and nothing else can exhibit it.
func TestConcurrentUpsertFindingReportsCreatedExactlyOnce(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-record-finding-concurrent-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	const writers = 20

	var wg sync.WaitGroup
	createds := make([]bool, writers)
	errs := make([]error, writers)

	// The writers wait on one barrier and are released together. Without
	// it they start as the loop schedules them, each one's statement
	// beginning after the previous has committed, and a snapshot race
	// that needs overlapping statements reproduces only occasionally.
	// The warm-up round before it opens the pool's connections first, so
	// the barrier releases writers that are ready to send rather than
	// writers that must still dial Postgres one after another.
	warmUp(t, st, projectKey, writers)
	ready := make(chan struct{})

	for i := range writers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			callCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			<-ready
			_, created, err := st.UpsertFinding(callCtx, projectKey, "kan-1", baseFinding("F1", i))
			errs[i] = err
			createds[i] = created
		}(i)
	}
	close(ready)
	wg.Wait()

	inserts := 0
	for i, err := range errs {
		if err != nil {
			t.Fatalf("writer %d: UpsertFinding: %v", i, err)
		}
		if createds[i] {
			inserts++
		}
	}
	if inserts != 1 {
		t.Errorf("%d of %d concurrent writers reported created, want exactly 1 -- the API would answer 201 Created %d times for one ref", inserts, writers, inserts)
	}

	rec, err := st.RunRecord(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("RunRecord: %v", err)
	}
	if len(rec.Findings) != 1 {
		t.Errorf("RunRecord returned %d findings, want exactly 1", len(rec.Findings))
	}
}

// TestSetFindingStatusUpdatesInPlace asserts the delta spec's fix-round
// scenario: F1's single row has its status updated, the change still holds
// exactly one row for F1, and nothing else about the finding moves --
// status is the only column a fix round may rewrite.
func TestSetFindingStatusUpdatesInPlace(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-record-status-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseFinding("F1", 0)
	in.Location = "internal/api/records.go:42"
	in.Reproducer = "cd stats && go test ./internal/api -run RecordDispatch"
	before, _, err := st.UpsertFinding(ctx, projectKey, "kan-1", in)
	if err != nil {
		t.Fatalf("UpsertFinding: %v", err)
	}

	if err := st.SetFindingStatus(ctx, projectKey, "kan-1", "F1", "fixed"); err != nil {
		t.Fatalf("SetFindingStatus: %v", err)
	}

	rec, err := st.RunRecord(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("RunRecord: %v", err)
	}
	if len(rec.Findings) != 1 {
		t.Fatalf("RunRecord returned %d findings, want exactly 1", len(rec.Findings))
	}
	after := rec.Findings[0]
	if after.Status != "fixed" {
		t.Errorf("status = %q, want fixed", after.Status)
	}
	want := before
	want.Status = "fixed"
	if after != want {
		t.Errorf("finding after SetFindingStatus = %+v, want only its status changed from %+v", after, before)
	}

	err = st.SetFindingStatus(ctx, projectKey, "kan-1", "F9", "fixed")
	if !errors.Is(err, store.ErrFindingNotFound) {
		t.Errorf("SetFindingStatus for an unknown ref = %v, want ErrFindingNotFound so the API can answer 404 rather than 500", err)
	}
}

// TestMergeDispatchMetricsDeepMerges pins that a dispatch's metrics bag is
// merged recursively, not concatenated: a top-level `||` loses "input"
// here, which is exactly the defect jsonb_deep_merge already exists to
// prevent for stage_runs.metrics. The harvester writes this bag in more
// than one pass, so a shallow merge would discard whatever an earlier pass
// recorded under the same parent.
//
// The third merge pins the half a recursive merge alone does not give: the
// harvester sends a per-batch *delta*, so a figure this bag already holds
// must be added to, never replaced. jsonb_deep_merge would leave "input"
// at 4 below rather than 14 -- a dispatch spanning two harvest cycles
// silently reduced to whichever batch landed last.
func TestMergeDispatchMetricsDeepMerges(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-record-metrics-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	recorded, err := st.RecordDispatch(ctx, projectKey, "kan-1", baseDispatch("implementer", "opus"))
	if err != nil {
		t.Fatalf("RecordDispatch: %v", err)
	}
	dispatchID := recorded.ID

	if err := st.MergeDispatchMetrics(ctx, dispatchID, json.RawMessage(`{"tokens":{"input":10}}`)); err != nil {
		t.Fatalf("first MergeDispatchMetrics: %v", err)
	}
	if err := st.MergeDispatchMetrics(ctx, dispatchID, json.RawMessage(`{"tokens":{"output":5}}`)); err != nil {
		t.Fatalf("second MergeDispatchMetrics: %v", err)
	}

	rec, err := st.RunRecord(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("RunRecord: %v", err)
	}
	if len(rec.Dispatches) != 1 {
		t.Fatalf("RunRecord returned %d dispatches, want 1", len(rec.Dispatches))
	}
	if !jsonEqual(t, rec.Dispatches[0].Metrics, json.RawMessage(`{"tokens":{"input":10,"output":5}}`)) {
		t.Errorf("metrics = %s, want both nested keys to survive the second merge", rec.Dispatches[0].Metrics)
	}

	if err := st.MergeDispatchMetrics(ctx, dispatchID, json.RawMessage(`{"tokens":{"input":4}}`)); err != nil {
		t.Fatalf("third MergeDispatchMetrics: %v", err)
	}
	rec, err = st.RunRecord(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("RunRecord after the third merge: %v", err)
	}
	if !jsonEqual(t, rec.Dispatches[0].Metrics, json.RawMessage(`{"tokens":{"input":14,"output":5}}`)) {
		t.Errorf("metrics = %s, want input summed to 14: the harvester sends a batch delta, not a total", rec.Dispatches[0].Metrics)
	}

	if err := st.MergeDispatchMetrics(ctx, dispatchID, nil); !errors.Is(err, store.ErrNilMetricsPatch) {
		t.Errorf("MergeDispatchMetrics with a nil patch = %v, want ErrNilMetricsPatch refused in Go rather than a NOT NULL violation", err)
	}

	if err := st.MergeDispatchMetrics(ctx, dispatchID+9999, json.RawMessage(`{"tokens":{"input":1}}`)); err == nil {
		t.Error("MergeDispatchMetrics for an unknown dispatch id succeeded, want an error naming the missing row")
	}
}

// TestRunRecordOrdersDispatchesAndFindings pins the order a render reads.
// Both are written here out of order on purpose: a record whose order came
// from insertion time rather than from seq and ref would pass a test that
// wrote them in order and fail the moment a journalled write replayed late.
func TestRunRecordOrdersDispatchesAndFindings(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-record-order-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	for i, role := range []string{"implementer", "reviewer", "panel-fix"} {
		in := baseDispatch(role, "opus")
		// Later-recorded dispatches carry earlier start instants, so an
		// ORDER BY started_at would come back reversed.
		in.StartedAt = time.Date(2026, 8, 22, 9, 0, 0, 0, time.UTC).Add(time.Duration(-i) * time.Hour)
		if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", in); err != nil {
			t.Fatalf("RecordDispatch %s: %v", role, err)
		}
	}

	for _, ref := range []string{"F3", "F1", "F2"} {
		if _, _, err := st.UpsertFinding(ctx, projectKey, "kan-1", baseFinding(ref, 0)); err != nil {
			t.Fatalf("UpsertFinding %s: %v", ref, err)
		}
	}

	rec, err := st.RunRecord(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("RunRecord: %v", err)
	}
	if rec.Change != "kan-1" {
		t.Errorf("Run.Change = %q, want kan-1", rec.Change)
	}

	wantSeqs := []int{1, 2, 3}
	gotSeqs := make([]int, 0, len(rec.Dispatches))
	for _, d := range rec.Dispatches {
		gotSeqs = append(gotSeqs, d.Seq)
	}
	if fmt.Sprint(gotSeqs) != fmt.Sprint(wantSeqs) {
		t.Errorf("dispatch seqs = %v, want %v (seq order)", gotSeqs, wantSeqs)
	}

	wantRefs := []string{"F1", "F2", "F3"}
	gotRefs := make([]string, 0, len(rec.Findings))
	for _, f := range rec.Findings {
		gotRefs = append(gotRefs, f.Ref)
	}
	if fmt.Sprint(gotRefs) != fmt.Sprint(wantRefs) {
		t.Errorf("finding refs = %v, want %v (ref order)", gotRefs, wantRefs)
	}

	// A change the store has never heard of is a distinct condition from a
	// change holding no rows: the former is ErrChangeNotFound, the latter
	// an empty record with no error, which is what lets a render report
	// "no rows for this change" rather than a failure.
	if _, err := st.RunRecord(ctx, projectKey, "kan-does-not-exist"); !errors.Is(err, store.ErrChangeNotFound) {
		t.Errorf("RunRecord for an unknown change = %v, want ErrChangeNotFound", err)
	}
	seedChange(t, st, projectKey, "kan-empty")
	empty, err := st.RunRecord(ctx, projectKey, "kan-empty")
	if err != nil {
		t.Fatalf("RunRecord for a change with no rows: %v", err)
	}
	if len(empty.Dispatches) != 0 || len(empty.Findings) != 0 {
		t.Errorf("RunRecord for a change with no rows = %+v, want an empty record", empty)
	}
}

// TestRunRecordOrdersFindingsNaturally is the case that stops the finding
// order being written as a plain lexical sort. `ORDER BY ref` alone
// returns F1, F10, F2, so a panel that raised ten findings would render
// its tenth between its first and its second -- an order no reader would
// read as an accident of collation rather than as the order the panel
// actually worked in.
//
// The refs are written out of order on purpose, so an implementation that
// happened to return insertion order could not pass by coincidence.
func TestRunRecordOrdersFindingsNaturally(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-record-natural-order-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	for _, ref := range []string{"F10", "F2", "F1"} {
		if _, _, err := st.UpsertFinding(ctx, projectKey, "kan-1", baseFinding(ref, 0)); err != nil {
			t.Fatalf("UpsertFinding %s: %v", ref, err)
		}
	}

	rec, err := st.RunRecord(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("RunRecord: %v", err)
	}
	gotRefs := make([]string, 0, len(rec.Findings))
	for _, f := range rec.Findings {
		gotRefs = append(gotRefs, f.Ref)
	}
	wantRefs := []string{"F1", "F2", "F10"}
	if fmt.Sprint(gotRefs) != fmt.Sprint(wantRefs) {
		t.Errorf("finding refs = %v, want %v -- lexical order would give [F1 F10 F2]", gotRefs, wantRefs)
	}

	// The other half of the same ordering clause: a ref no digits can be
	// read out of is ordered rather than crashing the cast the ordering
	// performs. It sorts last, among its own kind lexically, so a
	// non-conforming value written by a caller this change does not
	// control degrades to a stable position instead of taking the whole
	// record down with it.
	seedChange(t, st, projectKey, "kan-2")
	for _, ref := range []string{"F2", "unnumbered", "F1", "another"} {
		if _, _, err := st.UpsertFinding(ctx, projectKey, "kan-2", baseFinding(ref, 0)); err != nil {
			t.Fatalf("UpsertFinding %s: %v", ref, err)
		}
	}
	mixed, err := st.RunRecord(ctx, projectKey, "kan-2")
	if err != nil {
		t.Fatalf("RunRecord for the mixed refs: %v", err)
	}
	gotMixed := make([]string, 0, len(mixed.Findings))
	for _, f := range mixed.Findings {
		gotMixed = append(gotMixed, f.Ref)
	}
	wantMixed := []string{"F1", "F2", "another", "unnumbered"}
	if fmt.Sprint(gotMixed) != fmt.Sprint(wantMixed) {
		t.Errorf("finding refs = %v, want %v -- a ref carrying no digits sorts last, lexically", gotMixed, wantMixed)
	}
}

// TestUpsertFindingResolvesDispatchSeqToDispatchID asserts that the column
// the schema calls "the slot that raised it" is actually written. The wire
// shape names the raising dispatch by its seq -- the only identifier a
// caller has, since the row id is the store's own -- so the resolution
// from seq to dispatch_id has to happen in the store, and a finding whose
// DispatchSeq is nil legitimately leaves the column NULL.
func TestUpsertFindingResolvesDispatchSeqToDispatchID(t *testing.T) {
	st, pool := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-record-dispatchseq-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	raiser, err := st.RecordDispatch(ctx, projectKey, "kan-1", baseDispatch("reviewer", "sonnet"))
	if err != nil {
		t.Fatalf("RecordDispatch: %v", err)
	}

	raised := baseFinding("F1", 0)
	raised.DispatchSeq = &raiser.Seq
	if _, _, err := st.UpsertFinding(ctx, projectKey, "kan-1", raised); err != nil {
		t.Fatalf("UpsertFinding carrying a dispatch seq: %v", err)
	}
	if _, _, err := st.UpsertFinding(ctx, projectKey, "kan-1", baseFinding("F2", 0)); err != nil {
		t.Fatalf("UpsertFinding carrying no dispatch seq: %v", err)
	}

	var storedDispatchID *int64
	if err := pool.QueryRow(ctx, `
		SELECT f.dispatch_id FROM findings f
		JOIN changes c ON c.id = f.change_id
		WHERE c.project_key = $1 AND c.name = $2 AND f.ref = 'F1'
	`, projectKey, "kan-1").Scan(&storedDispatchID); err != nil {
		t.Fatalf("read F1's dispatch_id: %v", err)
	}
	if storedDispatchID == nil {
		t.Fatal("F1's dispatch_id is NULL, want the raising dispatch's own row id")
	}
	if *storedDispatchID != raiser.ID {
		t.Errorf("F1's dispatch_id = %d, want %d (the raising dispatch's row id)", *storedDispatchID, raiser.ID)
	}

	var unraisedDispatchID *int64
	if err := pool.QueryRow(ctx, `
		SELECT f.dispatch_id FROM findings f
		JOIN changes c ON c.id = f.change_id
		WHERE c.project_key = $1 AND c.name = $2 AND f.ref = 'F2'
	`, projectKey, "kan-1").Scan(&unraisedDispatchID); err != nil {
		t.Fatalf("read F2's dispatch_id: %v", err)
	}
	if unraisedDispatchID != nil {
		t.Errorf("F2's dispatch_id = %d, want NULL -- no single dispatch raised it", *unraisedDispatchID)
	}

	rec, err := st.RunRecord(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("RunRecord: %v", err)
	}
	if len(rec.Findings) != 2 {
		t.Fatalf("RunRecord returned %d findings, want 2", len(rec.Findings))
	}
	if rec.Findings[0].DispatchSeq == nil || *rec.Findings[0].DispatchSeq != raiser.Seq {
		t.Errorf("F1's DispatchSeq = %v, want %d read back as the seq the caller wrote", rec.Findings[0].DispatchSeq, raiser.Seq)
	}
	if rec.Findings[1].DispatchSeq != nil {
		t.Errorf("F2's DispatchSeq = %d, want nil", *rec.Findings[1].DispatchSeq)
	}
}

// TestRecordDispatchReturnsID asserts that a recorded dispatch comes back
// carrying the row's own id. MergeDispatchMetrics is keyed by that id, so
// without it the harvester's attribution pass could not name the row it
// just recorded through anything but a raw query -- which internal/store
// is the only package allowed to write.
func TestRecordDispatchReturnsID(t *testing.T) {
	st, pool := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-record-id-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	first, err := st.RecordDispatch(ctx, projectKey, "kan-1", baseDispatch("implementer", "opus"))
	if err != nil {
		t.Fatalf("first RecordDispatch: %v", err)
	}
	if first.ID == 0 {
		t.Fatal("RecordDispatch returned ID 0, want the inserted row's own id")
	}

	second, err := st.RecordDispatch(ctx, projectKey, "kan-1", baseDispatch("reviewer", "sonnet"))
	if err != nil {
		t.Fatalf("second RecordDispatch: %v", err)
	}
	if second.ID == first.ID {
		t.Errorf("both dispatches came back with ID %d, want distinct row ids", first.ID)
	}

	var storedID int64
	if err := pool.QueryRow(ctx, `
		SELECT d.id FROM dispatches d
		JOIN changes c ON c.id = d.change_id
		WHERE c.project_key = $1 AND c.name = $2 AND d.seq = $3
	`, projectKey, "kan-1", first.Seq).Scan(&storedID); err != nil {
		t.Fatalf("read the stored row's id: %v", err)
	}
	if storedID != first.ID {
		t.Errorf("RecordDispatch returned ID %d, want the stored row's %d", first.ID, storedID)
	}

	// The id is reachable enough to key a metrics merge, which is the
	// whole reason it is returned at all.
	if err := st.MergeDispatchMetrics(ctx, first.ID, json.RawMessage(`{"tokens":{"input":10}}`)); err != nil {
		t.Errorf("MergeDispatchMetrics keyed by the returned id: %v", err)
	}

	// RunRecord reports the same ids, so a record read back names its rows
	// the same way the write path did.
	rec, err := st.RunRecord(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("RunRecord: %v", err)
	}
	if len(rec.Dispatches) != 2 {
		t.Fatalf("RunRecord returned %d dispatches, want 2", len(rec.Dispatches))
	}
	if rec.Dispatches[0].ID != first.ID || rec.Dispatches[1].ID != second.ID {
		t.Errorf("RunRecord returned ids %d, %d, want %d, %d", rec.Dispatches[0].ID, rec.Dispatches[1].ID, first.ID, second.ID)
	}
}

// TestDispatchWindowsForSessionResolvesTheBoundToken is the query the
// second attribution pass rests on: a dispatch is attributable to a
// session only through the session_token its dispatcher was running under,
// bound to a session id on stage_runs by the harvester itself.
//
// It is a test rather than a bare method because the failure it guards
// against is silent in exactly the way this repository has already been
// bitten by (cmd/myflowd's own wiring test records the same lesson): a
// join that resolves nothing returns no windows, no window attributes no
// record, and every dispatch's metrics bag simply stays empty with nothing
// failing anywhere.
func TestDispatchWindowsForSessionResolvesTheBoundToken(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-dispatch-windows-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	const (
		sessionID = "session-dispatch-windows"
		token     = "mf-kan258-windows"
	)
	begin := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	begin.SessionID = ptr(sessionID)
	begin.SessionToken = ptr(token)
	if _, err := st.BeginStage(ctx, begin); err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	ended := time.Date(2026, 8, 22, 9, 30, 0, 0, time.UTC)
	mine := baseDispatch("implementer", "opus")
	mine.SessionToken = token
	mine.EndedAt = &ended
	recorded, err := st.RecordDispatch(ctx, projectKey, "kan-1", mine)
	if err != nil {
		t.Fatalf("RecordDispatch (bound token): %v", err)
	}

	// A dispatch under some other session's token, and one under no token
	// at all: neither has a window in this session, and the second has no
	// window anywhere.
	other := baseDispatch("reviewer", "sonnet")
	other.SessionToken = "mf-some-other-session"
	if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", other); err != nil {
		t.Fatalf("RecordDispatch (other token): %v", err)
	}
	if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", baseDispatch("reviewer", "sonnet")); err != nil {
		t.Fatalf("RecordDispatch (no token): %v", err)
	}

	windows, err := st.DispatchWindowsForSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("DispatchWindowsForSession: %v", err)
	}
	if len(windows) != 1 {
		t.Fatalf("windows = %+v, want exactly the one dispatch carrying this session's bound token", windows)
	}
	want := harvest.DispatchWindow{
		DispatchID: recorded.ID,
		StartedAt:  mine.StartedAt,
		EndedAt:    &ended,
	}
	if windows[0].DispatchID != want.DispatchID ||
		!windows[0].StartedAt.Equal(want.StartedAt) ||
		windows[0].EndedAt == nil || !windows[0].EndedAt.Equal(*want.EndedAt) {
		t.Errorf("window = %+v, want %+v", windows[0], want)
	}

	unbound, err := st.DispatchWindowsForSession(ctx, "session-nobody-bound")
	if err != nil {
		t.Fatalf("DispatchWindowsForSession for an unknown session: %v", err)
	}
	if len(unbound) != 0 {
		t.Errorf("windows for an unbound session = %+v, want none", unbound)
	}
}

// TestDispatchWindowsForSessionOrdersTiesByID pins the secondary sort key
// the windows query carries. Several dispatches recorded at the same
// instant is not an edge case here -- a review panel dispatches its slots
// at once, and the seconds resolution a caller writes -started-at at makes
// an exact tie ordinary -- and ORDER BY d.started_at alone leaves the
// order among them to whatever the plan happens to produce.
//
// The rows are perturbed before the read, deliberately: updating the
// lowest-id row rewrites its tuple at the end of the heap, so a scan in
// physical order no longer agrees with id order and a query with no
// secondary key has something to get wrong. Without that, every ordering
// would coincide and the assertion would pass whether the key were there
// or not.
//
// It states the contract; harvest's TestSameInstantDispatchWindowsResolveTheSameWayEveryTime
// is what makes attribution independent of this order regardless.
func TestDispatchWindowsForSessionOrdersTiesByID(t *testing.T) {
	st, pool := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-dispatch-window-ties-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	const (
		sessionID = "session-dispatch-window-ties"
		token     = "mf-kan258-window-ties"
	)
	begin := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	begin.SessionID = ptr(sessionID)
	begin.SessionToken = ptr(token)
	if _, err := st.BeginStage(ctx, begin); err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	// Every slot starts at the same instant: baseDispatch's own StartedAt,
	// unchanged, which is what makes this a tie rather than an ordering.
	var ids []int64
	for range 5 {
		in := baseDispatch("reviewer", "sonnet")
		in.SessionToken = token
		out, err := st.RecordDispatch(ctx, projectKey, "kan-1", in)
		if err != nil {
			t.Fatalf("RecordDispatch: %v", err)
		}
		ids = append(ids, out.ID)
	}

	if _, err := pool.Exec(ctx,
		`UPDATE dispatches SET notes = 'moved to the end of the heap' WHERE id = $1`, ids[0],
	); err != nil {
		t.Fatalf("perturb heap order: %v", err)
	}

	windows, err := st.DispatchWindowsForSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("DispatchWindowsForSession: %v", err)
	}
	if len(windows) != len(ids) {
		t.Fatalf("windows = %d, want %d", len(windows), len(ids))
	}
	for i, w := range windows {
		if w.DispatchID != ids[i] {
			t.Fatalf("windows[%d].DispatchID = %d, want %d -- tied started_at values must come back in ascending id order", i, w.DispatchID, ids[i])
		}
	}
}

// TestDispatchAgentIDRoundTripsAndAbsenceStaysAbsent pins the new column
// end to end through the typed API: a recorded agent id comes back on the
// row and on the window, and a dispatch recorded without one reports ""
// rather than a placeholder -- "not reported", which is the ordinary case
// on the two harnesses that expose no such identifier.
func TestDispatchAgentIDRoundTripsAndAbsenceStaysAbsent(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-dispatch-agent-id-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	const (
		sessionID = "session-dispatch-agent-id"
		token     = "mf-kan258-agent-id"
	)
	begin := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	begin.SessionID = ptr(sessionID)
	begin.SessionToken = ptr(token)
	if _, err := st.BeginStage(ctx, begin); err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	named := baseDispatch("reviewer", "sonnet")
	named.SessionToken = token
	named.AgentID = "agent-slot-a"
	if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", named); err != nil {
		t.Fatalf("RecordDispatch (with an agent id): %v", err)
	}

	anonymous := baseDispatch("reviewer", "sonnet")
	anonymous.SessionToken = token
	anonymous.StartedAt = named.StartedAt.Add(time.Minute)
	if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", anonymous); err != nil {
		t.Fatalf("RecordDispatch (no agent id): %v", err)
	}

	rec, err := st.RunRecord(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("RunRecord: %v", err)
	}
	if len(rec.Dispatches) != 2 {
		t.Fatalf("RunRecord returned %d dispatches, want 2", len(rec.Dispatches))
	}
	if rec.Dispatches[0].AgentID != "agent-slot-a" {
		t.Errorf("dispatch 1 AgentID = %q, want agent-slot-a", rec.Dispatches[0].AgentID)
	}
	if rec.Dispatches[1].AgentID != "" {
		t.Errorf("dispatch 2 AgentID = %q, want empty -- a harness reporting none records none", rec.Dispatches[1].AgentID)
	}

	windows, err := st.DispatchWindowsForSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("DispatchWindowsForSession: %v", err)
	}
	if len(windows) != 2 {
		t.Fatalf("windows = %d, want 2", len(windows))
	}
	if windows[0].AgentID != "agent-slot-a" {
		t.Errorf("windows[0].AgentID = %q, want agent-slot-a -- attribution matches on it and cannot read a column the query does not select", windows[0].AgentID)
	}
	if windows[1].AgentID != "" {
		t.Errorf("windows[1].AgentID = %q, want empty", windows[1].AgentID)
	}
}

// TestRenderedLedgerCallsAStoredDispatchWithNoMetricsNotMeasured renders a
// dispatch that came OUT OF THE STORE, rather than one built as a Go
// composite literal in the test, and pins that an unmeasured one still
// reads `not measured` in the ledger.
//
// The distinction is the whole point of the test. insertDispatch defaults
// an empty Metrics to the JSON object `{}` on the way in, so the shape a
// dispatch actually has between `myflow record dispatch` and the harvester
// running -- and the PERMANENT shape of every dispatch on Cursor and
// Codex, which write no transcript at all -- is two bytes, never
// zero-length and never SQL NULL. A renderer that decided "nothing was
// measured" by the bag's byte length agreed with every hand-built
// `records.Dispatch{}` in internal/records' own tests while reporting
// `input 0, output 0, cache read 0, cache creation 0` for every real row:
// a measured zero, indistinguishable from a genuine one.
//
// It lives in this package because that is where a real store is
// reachable. internal/records imports nothing but encoding/json and time,
// by design, so the round trip can only be asserted from the far side of
// it -- and asserting it from here is what makes the coverage real rather
// than self-agreeing.
func TestRenderedLedgerCallsAStoredDispatchWithNoMetricsNotMeasured(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-record-render-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	unmeasured, err := st.RecordDispatch(ctx, projectKey, "kan-1", baseDispatch("implementer", "opus"))
	if err != nil {
		t.Fatalf("record the unmeasured dispatch: %v", err)
	}

	measured := baseDispatch("implementer", "sonnet")
	measured.Metrics = json.RawMessage(`{"tokens":{"main":{"input":0,"output":0,"cache_read":0,"cache_creation":0},"sidechain":{"input":0,"output":0,"cache_read":0,"cache_creation":0}}}`)
	if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", measured); err != nil {
		t.Fatalf("record the measured dispatch: %v", err)
	}

	// The stored shape is asserted explicitly: if insertDispatch ever stops
	// defaulting to `{}` this test would otherwise keep passing while
	// covering a case the store no longer produces.
	if got := string(unmeasured.Metrics); got != "{}" {
		t.Errorf("a dispatch recorded with no metrics came back with Metrics = %q, want %q -- this test covers the store's real shape, so a change here has to be reflected in internal/records' rendering rule", got, "{}")
	}

	rec, err := st.RunRecord(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("read the run record back: %v", err)
	}
	if len(rec.Dispatches) != 2 {
		t.Fatalf("run record holds %d dispatches, want 2", len(rec.Dispatches))
	}
	if got := string(rec.Dispatches[0].Metrics); got != "{}" {
		t.Errorf("RunRecord returned Metrics = %q for the unmeasured dispatch, want %q", got, "{}")
	}

	out := records.RenderLedger(rec)
	sections := strings.Split(out, "## Dispatch ")
	if len(sections) != 3 {
		t.Fatalf("ledger renders %d dispatch sections, want 2:\n%s", len(sections)-1, out)
	}

	if line := tokensLine(t, sections[1]); line != "not measured" {
		t.Errorf("a dispatch the store holds no measurement for renders Tokens as %q, want %q -- zero is a measurement, and reporting it here is a figure the reader cannot tell from a real one:\n%s", line, "not measured", out)
	}
	want := "input 0, output 0, cache read 0, cache creation 0"
	if line := tokensLine(t, sections[2]); line != want {
		t.Errorf("a dispatch measured at exactly zero renders Tokens as %q, want %q -- an explicit zero is a real figure and must not be hidden:\n%s", line, want, out)
	}
}

// tokensLine returns the value of the one `- Tokens: ` line in a rendered
// ledger section.
func tokensLine(t *testing.T, section string) string {
	t.Helper()
	const prefix = "- Tokens: "
	for _, line := range strings.Split(section, "\n") {
		if after, ok := strings.CutPrefix(line, prefix); ok {
			return after
		}
	}
	t.Fatalf("ledger section carries no %q line:\n%s", prefix, section)
	return ""
}

// --- the begin/end pair, and what each half is for -------------------------

// dispatchPairFixture seeds a change and an open stage run whose session
// token is bound to sessionID, and returns the project key. It is the
// preamble every case below shares: a dispatch is attributable to a session
// only through that binding (DispatchWindowsForSession's own doc comment),
// so without it none of these cases can observe a window at all.
func dispatchPairFixture(t *testing.T, st *store.Store, label, sessionID, token string) string {
	t.Helper()
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-%s-%d", label, time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	begin := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	begin.SessionID = ptr(sessionID)
	begin.SessionToken = ptr(token)
	if _, err := st.BeginStage(ctx, begin); err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	return projectKey
}

// TestUsageDuringAnUnfinishedDispatchIsAttributedToIt is the case the
// begin half of the pair exists for, and it spans two packages on purpose:
// it records a dispatch that has NOT ended, reads its window straight back
// out of Postgres, and runs the real DispatchAttributor over a sidechain
// record timestamped inside it.
//
// The defect it pins is not a bug in either package alone. The harvester
// commits the transcript offset it has consumed every few seconds and never
// re-reads behind it (internal/harvest, Watcher), so a transcript record is
// offered to attribution exactly once -- at the tick that consumed it. A
// dispatch row written only when the dispatch CLOSED therefore did not
// exist for any of the ticks that ran while it was working, and every one
// of those records found no window: dropped, or credited to whichever
// earlier dispatch still had an open one. A subagent dispatch runs for
// minutes and a tick is seconds, so that was most of the usage this table
// exists to record.
//
// Reverting the insert in `begin` -- recording the dispatch at its close
// instead -- makes this test fail with the delta empty, which is exactly
// what the production defect looked like: nothing errored, the figures were
// simply not there.
func TestUsageDuringAnUnfinishedDispatchIsAttributedToIt(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()

	const (
		sessionID = "session-open-dispatch"
		token     = "mf-kan258-open-dispatch"
	)
	projectKey := dispatchPairFixture(t, st, "open-dispatch", sessionID, token)

	// Recorded as the dispatch STARTS: no end instant, no commit, no
	// outcome -- none of them knowable yet.
	in := baseDispatch("implementer", "opus")
	in.SessionToken = token
	in.Key = "task-6-implementer"
	opened, err := st.RecordDispatch(ctx, projectKey, "kan-1", in)
	if err != nil {
		t.Fatalf("RecordDispatch (begin): %v", err)
	}
	if opened.EndedAt != nil {
		t.Fatalf("a begun dispatch has EndedAt = %v, want nil -- begin records no end", opened.EndedAt)
	}

	// A harvest tick landing while the dispatch is still running, over a
	// sidechain record timestamped half an hour into it.
	deltas, err := harvest.NewDispatchAttributor(st).Attribute(ctx, []harvest.Record{{
		Timestamp:   in.StartedAt.Add(30 * time.Minute),
		SessionID:   sessionID,
		IsSidechain: true,
		Usage:       harvest.Usage{InputTokens: 1_200, OutputTokens: 340},
	}})
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	got, ok := deltas[opened.ID]
	if !ok {
		t.Fatalf("deltas = %+v, want usage attributed to dispatch %d -- a dispatch that has not ended yet still has an open window, and a harvest tick that runs mid-dispatch must find it", deltas, opened.ID)
	}
	if got.Sidechain.Input != 1_200 || got.Sidechain.Output != 340 {
		t.Errorf("delta = %+v, want the record's own 1200/340 figures", got.Sidechain)
	}
}

// TestAClosedDispatchStopsClaimingLaterUsage is the case the END half
// exists for. A dispatch row whose ended_at is NULL is an OPEN window, and
// an open window contains every instant after its start, forever: before
// `end` existed, no production path set the column at all, so every
// dispatch ever recorded went on being a candidate for every later record
// in its session.
//
// The case is deliberately a record in the GAP after a dispatch, not one
// after a later dispatch started. Where a second dispatch is open, the
// latest-start rule picks it whether or not the first was closed, and the
// stale window costs nothing observable. In the gap there is no second
// candidate, so a window that should have closed is the only one there --
// and it takes usage that belongs to nothing this table recorded.
func TestAClosedDispatchStopsClaimingLaterUsage(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()

	const (
		sessionID = "session-closed-dispatch"
		token     = "mf-kan258-closed-dispatch"
	)
	projectKey := dispatchPairFixture(t, st, "closed-dispatch", sessionID, token)

	in := baseDispatch("implementer", "opus")
	in.SessionToken = token
	in.Key = "task-6-implementer"
	opened, err := st.RecordDispatch(ctx, projectKey, "kan-1", in)
	if err != nil {
		t.Fatalf("RecordDispatch (begin): %v", err)
	}

	endedAt := in.StartedAt.Add(20 * time.Minute)
	closed, err := st.EndDispatch(ctx, projectKey, "kan-1", records.DispatchEnd{
		SessionToken: token,
		Key:          "task-6-implementer",
		CommitSHA:    "abc1234",
		Outcome:      "completed",
		EndedAt:      endedAt,
	})
	if err != nil {
		t.Fatalf("EndDispatch: %v", err)
	}
	if closed.ID != opened.ID || closed.Seq != opened.Seq {
		t.Fatalf("EndDispatch returned dispatch %d/seq %d, want the row begin wrote (%d/seq %d)", closed.ID, closed.Seq, opened.ID, opened.Seq)
	}
	if closed.EndedAt == nil || !closed.EndedAt.Equal(endedAt) {
		t.Fatalf("EndedAt = %v, want %v -- an unclosed window claims later usage forever", closed.EndedAt, endedAt)
	}
	if closed.CommitSHA != "abc1234" || closed.Outcome != "completed" {
		t.Errorf("closed row = commit %q outcome %q, want abc1234/completed", closed.CommitSHA, closed.Outcome)
	}

	deltas, err := harvest.NewDispatchAttributor(st).Attribute(ctx, []harvest.Record{{
		Timestamp:   endedAt.Add(5 * time.Minute),
		SessionID:   sessionID,
		IsSidechain: true,
		Usage:       harvest.Usage{InputTokens: 999},
	}})
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	if len(deltas) != 0 {
		t.Errorf("deltas = %+v, want none -- usage after a dispatch closed belongs to no dispatch, and a window left open would claim it", deltas)
	}
}

// TestRecordDispatchIsIdempotentUnderTheSameKey pins the third property of
// the pair: recording the same dispatch twice records ONE dispatch.
//
// A record write that could not reach the store is journalled and replayed
// later, and a lost response is indistinguishable from a store that was
// never reached -- so a replay carrying a row the store already holds is
// ordinary, not exotic. Without a key to collide on, the replay allocated a
// fresh seq and inserted a second row for one logical dispatch, and every
// cost figure derived from this table counted that dispatch twice.
// UpsertFinding and SetFindingStatus were already idempotent; this was the
// write whose duplicate costs money.
func TestRecordDispatchIsIdempotentUnderTheSameKey(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()

	const token = "mf-kan258-replay"
	projectKey := fmt.Sprintf("proj-dispatch-replay-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseDispatch("implementer", "opus")
	in.SessionToken = token
	in.Key = "task-6-implementer"

	first, err := st.RecordDispatch(ctx, projectKey, "kan-1", in)
	if err != nil {
		t.Fatalf("RecordDispatch (first attempt): %v", err)
	}
	replayed, err := st.RecordDispatch(ctx, projectKey, "kan-1", in)
	if err != nil {
		t.Fatalf("RecordDispatch (replay): %v", err)
	}
	if replayed.ID != first.ID || replayed.Seq != first.Seq {
		t.Errorf("replay produced dispatch %d/seq %d, want the original %d/seq %d", replayed.ID, replayed.Seq, first.ID, first.Seq)
	}

	rec, err := st.RunRecord(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("RunRecord: %v", err)
	}
	if len(rec.Dispatches) != 1 {
		t.Fatalf("the change holds %d dispatch rows, want 1 -- one logical dispatch, recorded twice, is one row", len(rec.Dispatches))
	}
	if rec.Dispatches[0].Key != "task-6-implementer" {
		t.Errorf("stored key = %q, want task-6-implementer", rec.Dispatches[0].Key)
	}

	// A different key under the same token is a different dispatch, so the
	// idempotency must not be "the store refuses a second dispatch".
	second := in
	second.Key = "task-7-implementer"
	if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", second); err != nil {
		t.Fatalf("RecordDispatch (a genuinely different dispatch): %v", err)
	}
	rec, err = st.RunRecord(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("RunRecord (after the second dispatch): %v", err)
	}
	if len(rec.Dispatches) != 2 {
		t.Fatalf("the change holds %d dispatch rows, want 2", len(rec.Dispatches))
	}
}

// TestEndDispatchReportsAKeyNamingNoDispatch pins that a key the change
// holds no dispatch under is reported rather than silently doing nothing.
// It is the answer internal/api turns into a 404 and the one 404 in this
// verb the CLI journals rather than refuses -- the begin it closes may
// still be queued ahead of it -- so a silent no-op here would lose the end
// entirely and leave the window open.
func TestEndDispatchReportsAKeyNamingNoDispatch(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()

	projectKey := fmt.Sprintf("proj-dispatch-end-404-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	_, err := st.EndDispatch(ctx, projectKey, "kan-1", records.DispatchEnd{
		SessionToken: "mf-kan258-nothing",
		Key:          "task-6-implementer",
		EndedAt:      time.Date(2026, 8, 22, 9, 30, 0, 0, time.UTC),
	})
	if !errors.Is(err, store.ErrDispatchNotFound) {
		t.Fatalf("EndDispatch for an unknown key = %v, want store.ErrDispatchNotFound", err)
	}
}
