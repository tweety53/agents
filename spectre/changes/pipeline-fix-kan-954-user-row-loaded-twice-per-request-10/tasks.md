# pipeline-fix-kan-954-user-row-loaded-twice-per-request-10

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** a self-review finding is recorded once, however many times it is written, and the
repeats already in the store collapse onto their first row.

**Architecture:** one migration adds the cleanup and a unique expression index on
`self_review_findings`; `RecordSelfReviewFinding` writes through `ON CONFLICT` on that index.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:`
  trailer; never write `spectre/changes/`.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; it was applied to this worktree at 752f3a28 and checked there. A diff that does not
  apply is reported, never hand-merged by guesswork.
- Store tests create their own database on `flow-postgres` (port 5433); never point a test or a
  command at the live `flowd` database.
- No new inline suppressions; `cd stats && gofmt -l .` prints nothing.

## Review Focus

- The `ON CONFLICT` target matches the index expression exactly, so the statement neither errors
  nor inserts a repeat.
- A rewritten sha, a different refless note and every existing `TestSelfReviewFindings*` case
  still record a new row.

- [ ] 1. Record a self-review finding once
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Save the diff below to
    `/tmp/pf10-task1.diff`, apply its test hunk alone
    (`git apply --include='stats/internal/store/*_test.go' /tmp/pf10-task1.diff`) and run
    `cd stats && go test ./internal/store/ -run 'SelfReviewFinding' -count=1`:
    `TestSelfReviewFindingRecordedOnce` FAILS with `a repeat answered with a new row: ids [1 2 3 4 5 6 7]`
    and `TestSelfReviewFindingsMigrationCollapsesRepeats` FAILS with
    `index "self_review_findings_once" does not exist`.
  - [ ] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/internal/store/*_test.go' /tmp/pf10-task1.diff`):

````diff verified:applied with git apply to this worktree at 752f3a28; both tests fail with the test hunk alone and pass with the whole diff; go test ./internal/store/ ./internal/api/ ./internal/client/ -count=1 green; gofmt -l and go vet clean
diff --git a/stats/internal/store/migrations/0038_self_review_findings_once.sql b/stats/internal/store/migrations/0038_self_review_findings_once.sql
new file mode 100644
index 00000000..e78f0d40
--- /dev/null
+++ b/stats/internal/store/migrations/0038_self_review_findings_once.sql
@@ -0,0 +1,33 @@
+-- 0038_self_review_findings_once.sql: one row per self-review finding.
+-- A finding is its (project, change, angle, disposition) plus its ref when
+-- it carries one -- the landed sha of a fixed row, the issue key of a filed
+-- one -- or its note when it does not. A rewritten sha is a new ref, so it
+-- still adds a second row, as skills/flow-self-review/SKILL.md step 6 states.
+-- Supersedes 0035's "no uniqueness constraint": a repeated write -- the
+-- landing's own fixed: rows recorded again by hand, a write retried after a
+-- client timeout that the store had already committed -- is never a new
+-- finding. md5(note) keeps a long note inside the btree's row-size limit.
+--
+-- Existing repeats collapse onto their first row, which takes the first
+-- non-null blast radius among them.
+
+UPDATE self_review_findings a
+SET blast_radius = (
+  SELECT b.blast_radius FROM self_review_findings b
+  WHERE b.project_key = a.project_key AND b.change = a.change AND b.angle = a.angle
+    AND b.disposition = a.disposition
+    AND COALESCE(b.ref, md5(b.note)) = COALESCE(a.ref, md5(a.note))
+    AND b.blast_radius IS NOT NULL
+  ORDER BY b.id LIMIT 1
+)
+WHERE a.blast_radius IS NULL;
+
+DELETE FROM self_review_findings a
+USING self_review_findings b
+WHERE b.project_key = a.project_key AND b.change = a.change AND b.angle = a.angle
+  AND b.disposition = a.disposition
+  AND COALESCE(b.ref, md5(b.note)) = COALESCE(a.ref, md5(a.note))
+  AND b.id < a.id;
+
+CREATE UNIQUE INDEX self_review_findings_once
+  ON self_review_findings (project_key, change, angle, disposition, (COALESCE(ref, md5(note))));
diff --git a/stats/internal/store/selfreviewfindings.go b/stats/internal/store/selfreviewfindings.go
index 2f33dc8c..86cc5982 100644
--- a/stats/internal/store/selfreviewfindings.go
+++ b/stats/internal/store/selfreviewfindings.go
@@ -43,7 +43,11 @@ func validateSelfReviewFinding(f records.SelfReviewFinding) error {
 	return nil
 }
 
-// RecordSelfReviewFinding validates and inserts one self-review finding.
+// RecordSelfReviewFinding validates and records one self-review finding,
+// once: a repeat of a recorded finding -- same project, change, angle and
+// disposition, and the same ref, or the same note when it carries no ref --
+// writes no second row and returns the recorded one, its missing blast
+// radius filled from the repeat.
 func (s *Store) RecordSelfReviewFinding(ctx context.Context, projectKey string, in records.SelfReviewFinding) (records.SelfReviewFinding, error) {
 	if err := validateSelfReviewFinding(in); err != nil {
 		return records.SelfReviewFinding{}, err
@@ -52,9 +56,11 @@ func (s *Store) RecordSelfReviewFinding(ctx context.Context, projectKey string,
 	if err := s.pool.QueryRow(ctx, `
 		INSERT INTO self_review_findings (project_key, change, angle, note, disposition, ref, blast_radius)
 		VALUES ($1, $2, $3, $4, $5, $6, $7)
-		RETURNING id, recorded_at
+		ON CONFLICT (project_key, change, angle, disposition, (COALESCE(ref, md5(note))))
+		DO UPDATE SET blast_radius = COALESCE(self_review_findings.blast_radius, EXCLUDED.blast_radius)
+		RETURNING id, note, blast_radius, recorded_at
 	`, projectKey, in.Change, in.Angle, in.Note, in.Disposition, nullIfEmpty(in.Ref), in.BlastRadius,
-	).Scan(&out.ID, &out.RecordedAt); err != nil {
+	).Scan(&out.ID, &out.Note, &out.BlastRadius, &out.RecordedAt); err != nil {
 		return records.SelfReviewFinding{}, fmt.Errorf("store: record self-review finding for %s/%s: %w", projectKey, in.Change, err)
 	}
 	return out, nil
diff --git a/stats/internal/store/selfreviewfindings_test.go b/stats/internal/store/selfreviewfindings_test.go
index 231fc0c0..12555e15 100644
--- a/stats/internal/store/selfreviewfindings_test.go
+++ b/stats/internal/store/selfreviewfindings_test.go
@@ -174,3 +174,104 @@ func TestSelfReviewFindingsInPeriod(t *testing.T) {
 		t.Fatalf("day 2, project A = %v, want %v", got, want)
 	}
 }
+
+// TestSelfReviewFindingRecordedOnce asserts a repeated write of one finding
+// -- the same sha recorded again with another note, the same refless note
+// recorded again -- stores no second row and answers with the recorded one,
+// its missing blast radius filled from the repeat, while a rewritten sha and
+// a different refless note are new findings.
+func TestSelfReviewFindingRecordedOnce(t *testing.T) {
+	st, _, projectKey := newHazardStore(t)
+	ctx := context.Background()
+
+	one := 1
+	writes := []records.SelfReviewFinding{
+		{Change: "kan-1", Angle: "flow-fix", Note: "the landing's own row", Disposition: "fixed", Ref: "14a900ef"},
+		{Change: "kan-1", Angle: "flow-fix", Note: "recorded again by hand", Disposition: "fixed", Ref: "14a900ef", BlastRadius: &one},
+		{Change: "kan-1", Angle: "flow-fix", Note: "recorded again by hand", Disposition: "fixed", Ref: "14a900ef", BlastRadius: &one},
+		{Change: "kan-1", Angle: "flow-fix", Note: "the landing's own row", Disposition: "fixed", Ref: "b0d0cdd5"},
+		{Change: "kan-1", Angle: "flow-cost", Note: "an observation", Disposition: "noted"},
+		{Change: "kan-1", Angle: "flow-cost", Note: "an observation", Disposition: "noted"},
+		{Change: "kan-1", Angle: "flow-cost", Note: "another observation", Disposition: "noted"},
+	}
+	var ids []int64
+	for _, f := range writes {
+		got, err := st.RecordSelfReviewFinding(ctx, projectKey, f)
+		if err != nil {
+			t.Fatalf("RecordSelfReviewFinding(%+v): %v", f, err)
+		}
+		ids = append(ids, got.ID)
+	}
+	if ids[1] != ids[0] || ids[2] != ids[0] || ids[5] != ids[4] {
+		t.Errorf("a repeat answered with a new row: ids %v", ids)
+	}
+
+	got, err := st.ListSelfReviewFindings(ctx, projectKey, "kan-1")
+	if err != nil {
+		t.Fatalf("ListSelfReviewFindings: %v", err)
+	}
+	var keys []string
+	for _, f := range got {
+		keys = append(keys, f.Ref+"/"+f.Note)
+	}
+	want := []string{"14a900ef/the landing's own row", "b0d0cdd5/the landing's own row", "/an observation", "/another observation"}
+	if !slices.Equal(keys, want) {
+		t.Fatalf("rows = %v, want %v", keys, want)
+	}
+	if got[0].BlastRadius == nil || *got[0].BlastRadius != 1 {
+		t.Errorf("first row's blast radius = %v, want 1 from the repeat", got[0].BlastRadius)
+	}
+}
+
+// TestSelfReviewFindingsMigrationCollapsesRepeats asserts
+// 0038_self_review_findings_once.sql collapses rows written twice before it
+// onto their first row, which takes the first non-null blast radius among
+// them, and leaves distinct findings alone.
+func TestSelfReviewFindingsMigrationCollapsesRepeats(t *testing.T) {
+	st, pool, projectKey := newHazardStore(t)
+	ctx := context.Background()
+
+	if _, err := pool.Exec(ctx, `DROP INDEX self_review_findings_once`); err != nil {
+		t.Fatalf("drop the 0038 index: %v", err)
+	}
+	for _, r := range []struct {
+		note, disposition, ref string
+		blast                  *int
+	}{
+		{"the landing's own row", "fixed", "a4a32afc", nil},
+		{"recorded again by hand", "fixed", "a4a32afc", ptr(16)},
+		{"recorded a third time", "fixed", "a4a32afc", ptr(15)},
+		{"another fix", "fixed", "766e24c4", ptr(2)},
+		{"an observation", "noted", "", nil},
+		{"an observation", "noted", "", nil},
+	} {
+		if _, err := pool.Exec(ctx, `
+			INSERT INTO self_review_findings (project_key, change, angle, note, disposition, ref, blast_radius)
+			VALUES ($1, 'kan-1', 'flow-fix', $2, $3, NULLIF($4, ''), $5)
+		`, projectKey, r.note, r.disposition, r.ref, r.blast); err != nil {
+			t.Fatalf("seed %q: %v", r.note, err)
+		}
+	}
+	if _, err := pool.Exec(ctx, `DELETE FROM schema_migrations WHERE filename = '0038_self_review_findings_once.sql'`); err != nil {
+		t.Fatalf("unrecord migration: %v", err)
+	}
+	if err := st.RunMigrations(ctx); err != nil {
+		t.Fatalf("RunMigrations: %v", err)
+	}
+
+	got, err := st.ListSelfReviewFindings(ctx, projectKey, "kan-1")
+	if err != nil {
+		t.Fatalf("ListSelfReviewFindings: %v", err)
+	}
+	var keys []string
+	for _, f := range got {
+		keys = append(keys, f.Ref+"/"+f.Note)
+	}
+	want := []string{"a4a32afc/the landing's own row", "766e24c4/another fix", "/an observation"}
+	if !slices.Equal(keys, want) {
+		t.Fatalf("rows = %v, want %v", keys, want)
+	}
+	if got[0].BlastRadius == nil || *got[0].BlastRadius != 16 {
+		t.Errorf("kept row's blast radius = %v, want 16, the first non-null among its repeats", got[0].BlastRadius)
+	}
+}
````

  - [ ] **Step 3: Verify.** `cd stats && gofmt -l ./internal/store && go vet ./internal/store/ && go test ./internal/store/ -run 'SelfReviewFinding|Migration' -count=1`
    — gofmt prints nothing, the rest exit 0.
  - [ ] **Step 4: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/store/selfreviewfindings.go`, `stats/internal/store/selfreviewfindings_test.go`, `stats/internal/store/migrations/0038_self_review_findings_once.sql`
**Tests:** `TestSelfReviewFindingRecordedOnce`, `TestSelfReviewFindingsMigrationCollapsesRepeats`
**Regression:** reverting the commit makes every write a new row and drops the migration — `TestSelfReviewFindingRecordedOnce` fails with `a repeat answered with a new row`, and `TestSelfReviewFindingsMigrationCollapsesRepeats` fails on the missing `self_review_findings_once` index.
**Baseline:** before=3 after=5
<!-- measured: grep -cE '^func Test' stats/internal/store/selfreviewfindings_test.go (before at 752f3a28; after with this task's diff applied) -->
**Commit:** fix(store): record a self-review finding once
**After:** none

**Decision:** finding-key
**Decision:** upsert-at-store
**Decision:** migration-cleanup
