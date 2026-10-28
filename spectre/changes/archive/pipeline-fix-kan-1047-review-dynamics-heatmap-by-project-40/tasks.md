# pipeline-fix-kan-1047-review-dynamics-heatmap-by-project-40

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** break the deadlock where a fix-mutation row missed before a re-review can never be recorded.

**Architecture:** a `-late '<reason>'` flag on `flow record mutation` skips the ordering refusal and
writes the reason as a pass note of the same round; guards unchanged; one sentence in
`skills/flow/review-panel-fix-round.md`.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-28-pipeline-fix-kan-1047-review-dynamics-heatmap-by-project-40-design.md` in this worktree.

## Global constraints

- Pipeline fix: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/pipeline-fix-kan-1047-review-dynamics-heatmap-by-project-40`, branch `pipeline-fix-kan-1047-review-dynamics-heatmap-by-project-40`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- The diff below applies with `git apply` from the worktree's root; checked at `edf8e4e8`. Save it by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged. Its prose sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt`.
- Run `gofmt -w` on every Go file you change; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- Task 1: without `-late` the refusal is unchanged (exit 2, no write); with it, exactly the mutation row and one pass row are written, and a failed mutation write writes no pass row.

- [x] 1. Admit a late fix-mutation row with its reason
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Save Step 2's diff, apply its
    test hunk alone (`git apply --include='stats/cmd/flow/record_test.go' <the saved diff>`) and run
    `cd stats && make -s web-dist && go test ./cmd/flow -run TestRecordMutationRefusesAfterReReviewBegan -count=1`:
    it fails with `flag provided but not defined: -late`.
    <!-- measured: the test hunk alone, go test -run TestRecordMutationRefusesAfterReReviewBegan @ edf8e4e8 → FAIL, -late undefined -->
  - [x] **Step 2: Apply the fix** — the rest (`git apply --exclude='stats/cmd/flow/record_test.go' <the saved diff>`):

````diff verified:applied with git apply to this worktree at edf8e4e8; go test ./cmd/flow -count=1 passes, the extended test fails with record.go reverted, go vet and gofmt -l clean
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index d86b1b34..77099810 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -3332,6 +3332,7 @@ func runRecordMutation(ctx context.Context, args []string, stdout, stderr io.Wri
 	path := fset.String("path", "", "the path that was mutated (required)")
 	mutated := fset.String("mutated", "", "what was mutated -- the literal \"none\" on the exemption form (required)")
 	test := fset.String("test", "", "the test that failed -- the reason, where -mutated is none (required)")
+	late := fset.String("late", "", "admit a row for a round whose re-review already began, with the reason it was missed; recorded beside it as a pass note")
 
 	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
 		return code
@@ -3369,8 +3370,8 @@ func runRecordMutation(ctx context.Context, args []string, stdout, stderr io.Wri
 	}); err == nil {
 		now := time.Now()
 		for _, d := range run.Dispatches {
-			if n, ok := records.PanelDispatchRound(d); ok && n == *round && !d.StartedAt.After(now) {
-				fmt.Fprintf(stderr, "flow: record mutation: panel round %d's re-review began %s -- a mutation proof is recorded before the re-review it precedes (skills/flow/review-panel-fix-round.md)\n", *round, d.StartedAt.UTC().Format(time.RFC3339))
+			if n, ok := records.PanelDispatchRound(d); ok && n == *round && !d.StartedAt.After(now) && *late == "" {
+				fmt.Fprintf(stderr, "flow: record mutation: panel round %d's re-review began %s -- a mutation proof is recorded before the re-review it precedes, or with -late '<reason>' when it was missed (skills/flow/review-panel-fix-round.md)\n", *round, d.StartedAt.UTC().Format(time.RFC3339))
 				return 2
 			}
 		}
@@ -3383,7 +3384,19 @@ func runRecordMutation(ctx context.Context, args []string, stdout, stderr io.Wri
 	if callErr == nil {
 		fmt.Fprintf(stdout, "recorded: mutation %d\n", out.ID)
 	}
-	return classifyRecordWrite(callErr, projectKey, f.change, "mutation", in, stderr)
+	if code := classifyRecordWrite(callErr, projectKey, f.change, "mutation", in, stderr); code != 0 || *late == "" {
+		return code
+	}
+	// A late row says so in the pass log: the proof came after the
+	// re-review began, and why (kan-1047 pipeline fix).
+	note := records.Pass{Round: *round, Note: fmt.Sprintf("late mutation proof for %s: %s", *path, *late)}
+	pout, passErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.Pass, error) {
+		return cl.RecordPass(ctx, projectKey, f.change, note)
+	})
+	if passErr == nil {
+		fmt.Fprintf(stdout, "recorded: pass %d\n", pout.ID)
+	}
+	return classifyRecordWrite(passErr, projectKey, f.change, "pass", note, stderr)
 }
 
 // resolveRenderKinds turns -kind into the kinds to render. "all" renders
diff --git a/stats/cmd/flow/record_test.go b/stats/cmd/flow/record_test.go
index b749dd7e..a2d5da4a 100644
--- a/stats/cmd/flow/record_test.go
+++ b/stats/cmd/flow/record_test.go
@@ -7124,6 +7124,7 @@ func TestRecordMutationRefusesAfterReReviewBegan(t *testing.T) {
 	record := `{"change":"kan-258","dispatches":[` +
 		`{"key":"panel-1-primary","agentId":"a1","role":"reviewer","slot":"primary","startedAt":"2026-10-28T14:38:10Z"}]}`
 	posts := 0
+	lastBody := ""
 	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
 		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/records/") {
 			_, _ = w.Write([]byte(record))
@@ -7134,6 +7135,8 @@ func TestRecordMutationRefusesAfterReReviewBegan(t *testing.T) {
 			return
 		}
 		posts++
+		body, _ := io.ReadAll(r.Body)
+		lastBody = string(body)
 		w.WriteHeader(http.StatusCreated)
 		_, _ = w.Write([]byte(`{"id":12,"round":2,"path":"render.go","mutated":"x","test":"y"}`))
 	}))
@@ -7156,4 +7159,13 @@ func TestRecordMutationRefusesAfterReReviewBegan(t *testing.T) {
 	if code := run(context.Background(), args("2"), strings.NewReader(""), &stdout, &stderr); code != 0 || posts != 1 {
 		t.Fatalf("round 2: exit %d, posts %d; want 0 and one (stderr %q)", code, posts, stderr.String())
 	}
+	// The late form admits round 1 and records its reason as a pass note
+	// (kan-1047 pipeline fix).
+	stderr.Reset()
+	if code := run(context.Background(), append(args("1"), "-late", "missed before re-review"), strings.NewReader(""), &stdout, &stderr); code != 0 || posts != 3 {
+		t.Fatalf("late round 1: exit %d, posts %d; want 0 and three (stderr %q)", code, posts, stderr.String())
+	}
+	if !strings.Contains(lastBody, "late mutation proof for render.go: missed before re-review") {
+		t.Errorf("late round 1: pass body %q carries no late note", lastBody)
+	}
 }
diff --git a/skills/flow/review-panel-fix-round.md b/skills/flow/review-panel-fix-round.md
index d208d8ae..b7536d35 100644
--- a/skills/flow/review-panel-fix-round.md
+++ b/skills/flow/review-panel-fix-round.md
@@ -371,3 +371,5 @@ fix-mutations-total: <n>
 Each row is recorded before the round's re-review begins, and `flow record mutation` refuses one
-after (kan-1046 self-review, 1d).
+after (kan-1046 self-review, 1d). A row missed before the re-review is recorded with
+`flow record mutation -late '<reason>'`, which admits it at its own round and records the reason as
+a pass note of that round, `late mutation proof for <path>: <reason>` (kan-1047 pipeline fix).
 
````

  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./cmd/flow -run TestRecordMutationRefusesAfterReReviewBegan -count=1` passes; `go vet ./cmd/flow` and `gofmt -l ./cmd/flow` are clean; `scripts/check-verbatim-moves.sh` and `scripts/check-references.sh` exit 0.
  - [x] **Step 4: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `skills/flow/review-panel-fix-round.md`
**Tests:** `TestRecordMutationRefusesAfterReReviewBegan` (extended with the late case)
<!-- measured: grep -ho '^func Test[A-Za-z0-9_]*' stats/cmd/flow/*_test.go | wc -l → 360 -->
**Baseline:** before=360 after=360
**Commit:** fix(flow): admit a late fix-mutation row with its reason
