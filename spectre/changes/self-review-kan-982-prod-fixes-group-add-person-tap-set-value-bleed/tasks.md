# self-review-kan-982-prod-fixes-group-add-person-tap-set-value-bleed

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`.
- Each task's change is the unified diff its step gives, applied from the repository root with `git apply` — every diff was applied in plan order to a scratch clone at `bc44caac` and checked there. A diff that does not apply is reported, never hand-merged by guesswork.
- Every prose line a diff adds is copied verbatim: this change's `verbatim-moves.txt` acknowledges each new or reworded sentence by its exact text, and any other wording fails `scripts/check-verbatim-moves.sh`.
- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w .` before `gofmt -l .`. `go build ./...` needs the built SPA (`internal/web/dist`); the commands below vet and test only the packages each task touches.

- [ ] 1. A back-filled dispatch row carries its real start and end instants
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `cd stats && go run ./cmd/flow record dispatch begin -started-at 2026-10-15T12:08:50Z -change x -role reviewer -model opus -key k -session-token mf-x 2>&1 | head -3` and confirm `flag provided but not defined: -started-at` — the defect: a back-fill cannot state when its dispatch ran, so the daemon stamps the back-fill's own instant.
  - [ ] **Step 2: Write the failing test.** Apply the `stats/cmd/flow/record_test.go` hunk of the diff in step 3 alone (`git apply --include='stats/cmd/flow/record_test.go' <the diff>`) and run `cd stats && go test ./cmd/flow/ -run TestRecordDispatchBackfillInstants -count=1`: it FAILS with `exit code = 2, want 0`.
  - [ ] **Step 3: Implement the flags.** Apply the rest of this diff:

````diff verified:applied with git apply in plan order to a scratch clone at bc44caac; go vet, gofmt -l and go test ./internal/guard/ ./cmd/flow/ pass with every task applied; TestRecordDispatchBackfillInstants fails with record.go reverted (exit code = 2, want 0)
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index c5f659a7..98eb94a6 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -253,11 +253,11 @@ func validateFindingRef(flagName, ref string) error {
 const recordUsage = `usage: flow record dispatch begin [-addr url] [-timeout dur] [-C dir]
                              -change name [-task id] -role role [-slot name]
                              -model model [-agent-id id|none] [-diff-base sha]
-                             -key key -session-token token
+                             [-started-at instant] -key key -session-token token
        flow record dispatch end   [-addr url] [-timeout dur] [-C dir]
                              -change name -key key -session-token token
                              [-commit sha] [-outcome outcome] [-cause cause]
-                             [-agent-id id] [-tokens k=v,...]
+                             [-agent-id id] [-tokens k=v,...] [-ended-at instant]
        flow record finding  [-addr url] [-timeout dur] [-C dir]
                              -change name -ref F<n> [-round n] -slot name
                              -severity sev [-location loc] -status status
@@ -954,6 +954,7 @@ func runRecordDispatchBegin(ctx context.Context, args []string, stdout, stderr i
 	diffBase := fset.String("diff-base", "", "the sha the diff this dispatch was given was computed from, where it was given a delta -- optional, since an implementer and a slot reading the whole diff record none")
 	key := fset.String("key", "", "this dispatch's own literal label, unique within the run -- what the end call closes, and what makes a replayed write land on one row (required)")
 	sessionToken := fset.String("session-token", "", "the run's own literal session token, unchanged from the mark that opened the run -- never a shell substitution (required)")
+	startedAt := fset.String("started-at", "", "a back-filled row's real start instant, RFC 3339 -- optional: omitted, the daemon stamps the moment it writes the row")
 
 	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
 		return code
@@ -961,6 +962,10 @@ func runRecordDispatchBegin(ctx context.Context, args []string, stdout, stderr i
 	if !requireChangeCheckout(f.dir, f.change, stderr) {
 		return 2
 	}
+	started, ok := parseBackfillInstant("-started-at", *startedAt, stderr)
+	if !ok {
+		return 2
+	}
 	if !requireRecordFlags(stderr,
 		[2]string{"-role", *role},
 		[2]string{"-model", *model},
@@ -1007,11 +1012,12 @@ func runRecordDispatchBegin(ctx context.Context, args []string, stdout, stderr i
 		Effort:       *effort,
 		DiffBase:     *diffBase,
 		SessionToken: *sessionToken,
+		StartedAt:    started,
 	}
 	out, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.Dispatch, error) {
 		return cl.RecordDispatch(ctx, projectKey, f.change, in)
 	})
-	if callErr != nil {
+	if callErr != nil && in.StartedAt.IsZero() {
 		in.StartedAt = attemptAt
 	}
 	if callErr == nil {
@@ -1056,6 +1062,7 @@ func runRecordDispatchEnd(ctx context.Context, args []string, stdout, stderr io.
 	cause := fset.String("cause", "", "why the outcome is blocked -- one of: "+strings.Join(recordCauses, ", ")+"; required with -outcome blocked, refused with any other outcome")
 	agentID := fset.String("agent-id", "", "the harness's own identifier for the dispatched subagent, where begin could not carry it -- optional, and never clears an identifier begin already recorded")
 	tokens := fset.String("tokens", "", "the token usage a same-session (inline) dispatch spent, as comma-separated key=value pairs over: input, output, cache-read, cache-creation (optional -- for a harness that writes no transcript the harvester could credit the inline row from)")
+	endedAt := fset.String("ended-at", "", "a back-filled row's real end instant, RFC 3339 -- optional: omitted, the daemon stamps the moment it closes the row")
 
 	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
 		return code
@@ -1063,6 +1070,10 @@ func runRecordDispatchEnd(ctx context.Context, args []string, stdout, stderr io.
 	if !requireChangeCheckout(f.dir, f.change, stderr) {
 		return 2
 	}
+	ended, ok := parseBackfillInstant("-ended-at", *endedAt, stderr)
+	if !ok {
+		return 2
+	}
 	if !requireRecordFlags(stderr,
 		[2]string{"-key", *key},
 		[2]string{"-session-token", *sessionToken},
@@ -1126,11 +1137,12 @@ func runRecordDispatchEnd(ctx context.Context, args []string, stdout, stderr io.
 		Cause:        *cause,
 		AgentID:      *agentID,
 		Tokens:       report,
+		EndedAt:      ended,
 	}
 	out, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.Dispatch, error) {
 		return cl.EndDispatch(ctx, projectKey, f.change, in)
 	})
-	if callErr != nil {
+	if callErr != nil && in.EndedAt.IsZero() {
 		in.EndedAt = attemptAt
 	}
 	if callErr == nil {
@@ -1139,6 +1151,24 @@ func runRecordDispatchEnd(ctx context.Context, args []string, stdout, stderr io.
 	return classifyDispatchEnd(callErr, projectKey, f.change, in, stderr)
 }
 
+// parseBackfillInstant reads -started-at or -ended-at: empty is the zero
+// instant, which the daemon stamps at its own clock (KAN-324); anything
+// else must be RFC 3339. The flags exist for a back-fill -- a row recorded
+// after its dispatch ran, whose daemon-stamped instants would all read the
+// back-fill's own moment (kan-982: five rows at 17:47:20Z) -- and the
+// daemon keeps a supplied instant exactly as it keeps a journal replay's.
+func parseBackfillInstant(flagName, v string, stderr io.Writer) (time.Time, bool) {
+	if v == "" {
+		return time.Time{}, true
+	}
+	t, err := time.Parse(time.RFC3339, v)
+	if err != nil {
+		fmt.Fprintf(stderr, "flow: %s %q is not an RFC 3339 instant\n", flagName, v)
+		return time.Time{}, false
+	}
+	return t, true
+}
+
 // classifyDispatchEnd is classifyRecordWrite with ONE exception, and the
 // exception is the whole reason it exists: a 404 from this route is
 // journalled rather than reported -- when, and only when, the record
diff --git a/stats/cmd/flow/record_test.go b/stats/cmd/flow/record_test.go
index 493410ca..e5e16639 100644
--- a/stats/cmd/flow/record_test.go
+++ b/stats/cmd/flow/record_test.go
@@ -4824,3 +4824,48 @@ func TestRequireChangeCheckoutRefusesNonChangeNames(t *testing.T) {
 		t.Error("real change refused")
 	}
 }
+
+// TestRecordDispatchBackfillInstants pins the kan-982 self-review fix: a
+// back-filled row carries the dispatch's real instants. -started-at reaches
+// begin's body as startedAt and -ended-at end's as endedAt; an instant that
+// is not RFC 3339 is a usage refusal, exit 2, before the store is called.
+func TestRecordDispatchBackfillInstants(t *testing.T) {
+	const start, end = "2026-10-15T12:08:50Z", "2026-10-15T12:31:07Z"
+	if sent := dispatchBeginBody(t, "-started-at", start); sent["startedAt"] != start {
+		t.Errorf("begin startedAt = %v, want %s", sent["startedAt"], start)
+	}
+
+	repo := recordRepo(t, "kan-258")
+	isolatedStateRoot(t)
+	var gotBody []byte
+	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
+		gotBody, _ = readAll(r)
+		w.WriteHeader(http.StatusOK)
+		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"reviewer","model":"sonnet","endedAt":"` + end + `"}`))
+	}))
+	defer srv.Close()
+	endArgs := []string{"record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
+		"-change", "kan-258", "-key", "panel-primary", "-session-token", "mf-record-backfill"}
+	var stdout, stderr bytes.Buffer
+	if code := run(context.Background(), append(endArgs, "-ended-at", end), strings.NewReader(""), &stdout, &stderr); code != 0 {
+		t.Fatalf("end exit = %d, want 0; stderr:\n%s", code, stderr.String())
+	}
+	var sent map[string]any
+	if err := json.Unmarshal(gotBody, &sent); err != nil || sent["endedAt"] != end {
+		t.Errorf("end endedAt = %v (%v), want %s", sent["endedAt"], err, end)
+	}
+
+	for _, args := range [][]string{
+		{"record", "dispatch", "begin", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
+			"-change", "kan-258", "-role", "reviewer", "-model", "sonnet", "-key", "k", "-session-token", "mf-record-backfill",
+			"-started-at", "15 Oct 12:08"},
+		append(append([]string{}, endArgs...), "-ended-at", "yesterday"),
+	} {
+		stdout.Reset()
+		stderr.Reset()
+		gotBody = nil
+		if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "is not an RFC 3339 instant") || gotBody != nil {
+			t.Errorf("%s: exit %d, store called %v, stderr %q; want 2, no call, an RFC 3339 refusal", args[2], code, gotBody != nil, stderr.String())
+		}
+	}
+}
````

  - [ ] **Step 4: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./cmd/flow/ && go test ./cmd/flow/ -run 'TestRecordDispatchBackfillInstants|TestRecordDispatch|TestDispatchEnd' -count=1` — `gofmt -l` prints nothing, the tests pass. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchBackfillInstants`
**Regression:** reverting the commit refuses `-started-at` and `-ended-at` as undefined flags (exit 2), so a back-filled row is again stamped with the back-fill's own instant.
**Baseline:** before=102 after=103
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ bc44caac (before; after with this task's diff applied in the scratch clone) -->
**Commit:** feat(flow-cli): a back-filled dispatch row carries its real start and end instants
**After:** none

**Decision:** backfill-instant-flags

- [ ] 2. close-task refuses an appended task no dispatch row names
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Read the SIGNAL FOUR block in `closeTask` (`stats/internal/guard/closetask.go`): it prints `close-task: WARNING — appended task …` and goes on to tick and push — the defect; kan-982 ticked Tasks 19–21 past it.
  - [ ] **Step 2: Write the failing test.** Apply the `stats/internal/guard/close_task_test.go` hunk of the diff in step 3 alone and run `cd stats && go test ./internal/guard/ -run TestCloseTaskRefusesAnUndispatchedAppendedTask -count=1`: it FAILS with `refusal present false, want true`.
  - [ ] **Step 3: Implement the refusal.** Apply the rest of this diff — the guard, its shim's header and the exit-1 clause of `implement.md`'s task-close step:

````diff verified:applied with git apply in plan order to a scratch clone at bc44caac; go vet, gofmt -l and go test ./internal/guard/ ./cmd/flow/ pass with every task applied; TestCloseTaskRefusesAnUndispatchedAppendedTask fails with closetask.go reverted (refusal present false, want true); check-verbatim-moves.sh (with this change's verbatim-moves.txt), check-references.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh and check-normative-inventory.sh exit 0
diff --git a/scripts/close-task.sh b/scripts/close-task.sh
index afdf11b9..a4ac7c31 100755
--- a/scripts/close-task.sh
+++ b/scripts/close-task.sh
@@ -11,12 +11,17 @@
 #   1. With -end-key: `flow record dispatch end -change <name> -key <key>
 #      -session-token <token> [-commit <sha>] -outcome completed`. A record
 #      never blocks: a failed call prints a warning and the sequence goes on.
+#      Then, for every task it closes that the plan's `**Tasks appended:**`
+#      line counts, `flow record dispatches -change <name> -C <canonical>`:
+#      an appended task no row names is refused at step 4 (exit 1); a
+#      store that cannot be read refuses nothing.
 #   2. check-task-commit-fields.sh <canonical> <id> <sha|map> "" <canonical>
 #      <name> for every task, and its commit's — every map pair's — Task-Id:
 #      trailer must name the task or a task its Build: red fold joins (exit 1
 #      otherwise): the trailer is how a task's commit is found again.
 #   3. check-task-commit-planning-paths.sh <canonical> <merge-base>.
-#   4. Any guard exit 2 → exit 2; else any exit 1 → exit 1. Nothing is ticked
+#   4. Any guard exit 2 → exit 2; else any exit 1 → exit 1; else an
+#      appended task step 1 found no row naming → exit 1. Nothing is ticked
 #      or pushed either way, so a refused commit is never pushed.
 #   5. check-review-gate.sh <canonical> <id> <sha|map> <canonical> <name>
 #      [<the task's -undeclared paths>…] for every task; its FIRE:/QUIET:
@@ -37,7 +42,8 @@
 #
 # Exit 0 closed (read each FIRE: line: on big that task's tick waits for its
 # reviewer; elsewhere it is the panel's review focus); 1 a commit guard refused — re-commit, then re-run without the
-# -end-* flags; 2 stop: a guard could not judge, a tick or push failed, or
+# -end-* flags — or an appended task has no dispatch row — record it, then
+# re-run without the -end-* flags; 2 stop: a guard could not judge, a tick or push failed, or
 # a usage error.
 #
 # flow-guard is built from this checkout, never taken from PATH:
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index 9f239694..32381a81 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -675,7 +675,10 @@ entry, one or more bundles `plan-dispatch-bundles.sh` emits. At each boundary, i
    a task commit carrying no `Task-Id:` trailer naming the task —
    is the parent's own fix, with nothing ticked or pushed: it re-commits that task itself,
    without the swept paths and carrying `Task-Id: <n>`, and re-runs `close-task.sh` without the `-end-*` flags before
-   anything below; **exit 2 — a guard's not-a-verdict close (an unreadable plan, a task it
+   anything below; **exit 1 with `close-task: RECORD —`** is an appended task no dispatch row
+   names — record its implementer dispatch (`flow record dispatch begin … -task <n>` and its
+   `end`, a back-fill passing the dispatch's real `-started-at` and `-ended-at`), then re-run
+   `close-task.sh` without the `-end-*` flags; **exit 2 — a guard's not-a-verdict close (an unreadable plan, a task it
    cannot resolve, a commit range git cannot resolve, a usage error), or a failed tick or push —
    stops the run**: re-committing cannot repair an inability, so it is never folded into
    the re-commit loop. Exit 0 prints every task's `FIRE:` or `QUIET:` gate line.
diff --git a/stats/internal/guard/close_task_test.go b/stats/internal/guard/close_task_test.go
index 442da88b..c4dcc43c 100644
--- a/stats/internal/guard/close_task_test.go
+++ b/stats/internal/guard/close_task_test.go
@@ -359,14 +359,15 @@ func TestCloseTaskTaskIdTrailer(t *testing.T) {
 }
 
 // Signal four at close time: closing an appended task no dispatch row names
-// warns, never blocks; a plan-time task never reads the store at all.
-func TestCloseTaskWarnsOnAnUndispatchedAppendedTask(t *testing.T) {
+// is refused, exit 1, nothing ticked or pushed (kan-982 self-review); a
+// plan-time task never reads the store at all.
+func TestCloseTaskRefusesAnUndispatchedAppendedTask(t *testing.T) {
 	t.Parallel()
 	for _, c := range []struct {
 		name, rows string
 		warn       bool
 	}{
-		{"an appended task no row names warns", `[{"key":"task-1-implementer","taskId":""}]`, true},
+		{"an appended task no row names is refused", `[{"key":"task-1-implementer","taskId":""}]`, true},
 		{"an appended task its key names is silent", `[{"key":"task-2-implementer","taskId":""}]`, false},
 		{"an appended task its taskId names is silent", `[{"key":"impl","taskId":"2"}]`, false},
 	} {
@@ -377,12 +378,19 @@ func TestCloseTaskWarnsOnAnUndispatchedAppendedTask(t *testing.T) {
 			writeFile(t, fx.dispatches, c.rows)
 			s2 := fx.commit(t, fx.wt, "two", "2", "b.txt", "b\n")
 			r := fx.run(fx.wt, "demo", fx.base, "2:"+s2)
-			if r.rc != 0 {
-				t.Fatalf("exit %d, want 0 — a warning never blocks\n%s", r.rc, r.out)
-			}
-			const want = "close-task: WARNING — appended task 2 has no dispatch row naming it"
+			const want = "close-task: appended task 2 has no dispatch row naming it"
 			if got := strings.Contains(r.err, want); got != c.warn {
-				t.Errorf("warning present %v, want %v\n%s", got, c.warn, r.out)
+				t.Errorf("refusal present %v, want %v\n%s", got, c.warn, r.out)
+			}
+			if c.warn {
+				if r.rc != 1 || !strings.Contains(r.err, "close-task: RECORD —") {
+					t.Fatalf("exit %d, want 1 and the RECORD line\n%s", r.rc, r.out)
+				}
+				ctEqual(t, fx.calls(t), []string{"flow record dispatches -change demo -C " + fx.wt})
+				return
+			}
+			if r.rc != 0 {
+				t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
 			}
 			ctEqual(t, fx.calls(t), []string{
 				"flow record dispatches -change demo -C " + fx.wt,
diff --git a/stats/internal/guard/closetask.go b/stats/internal/guard/closetask.go
index f085a2cb..27e4ccdf 100644
--- a/stats/internal/guard/closetask.go
+++ b/stats/internal/guard/closetask.go
@@ -82,15 +82,18 @@ func closeTask(args []string, env Env, stdout, stderr io.Writer) int {
 	}
 
 	// SIGNAL FOUR at close time (scripts/check-unfinished-work.sh): an
-	// appended task no dispatch row names is warned about here, while its
-	// row can still be recorded, rather than first refused at integrate. A
-	// warning, never a stop: a record never blocks.
+	// appended task no dispatch row names is refused here, nothing ticked
+	// or pushed, while its row can still be recorded. A warning was ticked
+	// past (kan-982: Tasks 19-21 closed with none) and surfaced only at
+	// integrate's gate. A store that cannot be read refuses nothing --
+	// ctUndispatched reads it as no finding: a record never blocks.
 	var closed []string
 	for _, t := range tasks {
 		closed = append(closed, t.id)
 	}
-	for _, id := range ctUndispatched(env, canonical, name, closed) {
-		fmt.Fprintf(stderr, "close-task: WARNING — appended task %s has no dispatch row naming it; record its dispatch (`flow record dispatch begin … -task %s`) before integrate, whose unfinished-work gate refuses it\n", id, id)
+	undispatched := ctUndispatched(env, canonical, name, closed)
+	for _, id := range undispatched {
+		fmt.Fprintf(stderr, "close-task: appended task %s has no dispatch row naming it — record its dispatch (`flow record dispatch begin … -task %s` and its `end`; a back-fill passes -started-at and -ended-at), then re-run without the -end-* flags\n", id, id)
 	}
 
 	// Both commit guards run before anything is ticked or pushed, so a
@@ -110,6 +113,10 @@ func closeTask(args []string, env Env, stdout, stderr io.Writer) int {
 		fmt.Fprintln(stderr, "close-task: STOP — a commit guard could not judge; nothing was ticked or pushed")
 		return 2
 	}
+	if len(undispatched) > 0 {
+		fmt.Fprintln(stderr, "close-task: RECORD — an appended task has no dispatch row; nothing was ticked or pushed")
+		return 1
+	}
 
 	var quiet, all, fired []string
 	for _, t := range tasks {
````

  - [ ] **Step 4: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCloseTask' -count=1`, then from the repository root `bash -n scripts/close-task.sh`, `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh` — all exit 0 with `gofmt -l` printing nothing. Commit.

**Build:** green
**Files:** `stats/internal/guard/closetask.go`, `stats/internal/guard/close_task_test.go`, `scripts/close-task.sh`, `skills/flow/implement.md`
**Tests:** `TestCloseTaskRefusesAnUndispatchedAppendedTask`
**Regression:** reverting the commit lets close-task exit 0 and tick and push an appended task no dispatch row names, printing only a warning.
**Baseline:** before=4 after=4
<!-- measured: grep -c '^func Test' stats/internal/guard/close_task_test.go @ bc44caac (before; after with this task's diff applied in the scratch clone — the test is renamed, not added) -->
**Commit:** fix(guard): close-task refuses an appended task no dispatch row names
**After:** Task 1

**Decision:** close-task-refuses-undispatched

- [ ] 3. Integrate's visual-verify check matches a satellite against the canonical `ui paths`
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Read `checkVisualVerifyDispatched` (`stats/internal/guard/visualverifydispatched.go`) and the `check-visual-verify-dispatched.sh` call in `scripts/integrate-change.sh`'s `cmd_prepare`: the trigger is asked of each worktree's own `.flow/project.md` only and the store is read `-C <worktree>` — the defect: a frontend worktree with no section is "not configured" though the canonical project names its files `gymie-frontend/**`.
  - [ ] **Step 2: Write the failing test.** Apply the `stats/internal/guard/check_visual_verify_dispatched_test.go` hunks of the diff in step 3 alone and run `cd stats && go test ./internal/guard/ -run 'TestCheckVisualVerifyDispatched' -count=1`: the new cases of `TestCheckVisualVerifyDispatchedCanonicalUIPaths` and case 8d's usage line FAIL.
  - [ ] **Step 3: Implement the canonical match.** Apply the rest of this diff — the guard, its shim's header, integrate's call and the finish contract's **Run 1** line:

````diff verified:applied with git apply in plan order to a scratch clone at bc44caac; go vet, gofmt -l and go test ./internal/guard/ ./cmd/flow/ pass with every task applied; three cases of TestCheckVisualVerifyDispatchedCanonicalUIPaths fail with visualverifydispatched.go reverted; check-verbatim-moves.sh (with this change's verbatim-moves.txt), check-references.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh and check-normative-inventory.sh exit 0
diff --git a/scripts/check-visual-verify-dispatched.sh b/scripts/check-visual-verify-dispatched.sh
index 454ac8f3..b6e2b925 100755
--- a/scripts/check-visual-verify-dispatched.sh
+++ b/scripts/check-visual-verify-dispatched.sh
@@ -3,7 +3,7 @@
 # touched a declared `## visual verification` UI path actually dispatched
 # `flow.visual-verify`'s own verifier, for /flow's run-1 finish gate.
 #
-# Usage: check-visual-verify-dispatched.sh <worktree> <change-name> <merge-base>
+# Usage: check-visual-verify-dispatched.sh <worktree> <change-name> <merge-base> [<canonical-worktree>]
 #
 # WHY THIS GUARD EXISTS. `flow.visual-verify` (skills/flow/visual-verify.md)
 # is the one stage built to catch a defect that is obvious the moment a page
@@ -50,7 +50,15 @@
 #      below. The trigger reads `.flow/project.md` from the
 #      worktree itself, matching flow.visual-verify's own per-worktree
 #      resolution; this guard is called once per worktree in the run's
-#      resolved set, exactly as that stage is.
+#      resolved set, exactly as that stage is. Given a canonical worktree
+#      other than <worktree>, a diff the worktree's own answer does not
+#      match is also matched against the canonical worktree's `ui paths`,
+#      each path prefixed with the worktree's repository name — its main
+#      checkout's basename, as check-visual-trigger.sh --repo prefixes it —
+#      and the better answer stands, 0 over 1 over 2; every store call is
+#      then made under the canonical worktree, whose project the run's rows
+#      were recorded in. A canonical worktree that is not a directory, or a
+#      repository git cannot name, is exit 2.
 #
 # THE DISPATCH CHECK READS THE STORE, NOT A RENDERED FILE — the same
 # posture check-panel-fix-single-dispatch.sh and check-panel-reproducers.sh
diff --git a/scripts/integrate-change.sh b/scripts/integrate-change.sh
index 76fa34ca..d6757a6e 100755
--- a/scripts/integrate-change.sh
+++ b/scripts/integrate-change.sh
@@ -379,7 +379,7 @@ cmd_prepare() {
         0:OUTSTANDING:*) ask=1 ;;
         *) stop no-verdict "check-unfinished-work reached no verdict for ${WT[i]} (exit $RC) — report its output and ask the operator; --accept-outstanding does not pass this" ;;
       esac
-      run "$SCRIPT_DIR/check-visual-verify-dispatched.sh" "${WT[i]}" "$NAME" "$(git -C "${WT[i]}" merge-base HEAD "origin/${BASES[i]}")"
+      run "$SCRIPT_DIR/check-visual-verify-dispatched.sh" "${WT[i]}" "$NAME" "$(git -C "${WT[i]}" merge-base HEAD "origin/${BASES[i]}")" "$CANON"
       case "$RC:$(verdict 'VISUAL-VERIFY-OK|VISUAL-VERIFY-MISSING')" in
         0:VISUAL-VERIFY-OK:*) ;;
         1:VISUAL-VERIFY-MISSING:*) ask=1 ;;
diff --git a/skills/flow-contracts/finish-contract-run1.md b/skills/flow-contracts/finish-contract-run1.md
index eb17e5be..002261b5 100644
--- a/skills/flow-contracts/finish-contract-run1.md
+++ b/skills/flow-contracts/finish-contract-run1.md
@@ -155,10 +155,11 @@ A missing verdict line is not a verdict. Treat it exactly as the preflight scrip
 above: stop and ask the operator. The exit code is checked as well as the line, because a caller
 that greps for `CLEAR` in empty output finds nothing.
 
-**Then run `check-visual-verify-dispatched.sh <worktree> <change-name> $(git -C <worktree> merge-base HEAD origin/$BASE)`**,
+**Then run `check-visual-verify-dispatched.sh <worktree> <change-name> $(git -C <worktree> merge-base HEAD origin/$BASE) <canonical-worktree>`**,
 once per worktree in the same resolved set — the fork point itself, because this runs before the
 base-moved check below can name a rebased one, and the state-file value is stale after an earlier
-rebase. A `VISUAL-VERIFY-OK` line joins `CLEAR` and folds no further
+rebase. The canonical worktree makes a satellite's diff match the canonical `ui paths` as well as
+its own, as **Visual verification** (`skills/flow/verify-and-handoff.md`) step 2 matches it. A `VISUAL-VERIFY-OK` line joins `CLEAR` and folds no further
 breakdown in; a `VISUAL-VERIFY-MISSING` line is treated exactly as `OUTSTANDING` — it feeds the same
 breakdown and the same three courses (**The three courses**, `skills/flow/unfinished-work-gate.md`), not a second prompt. Exit 2 (cannot answer) is stop-and-ask,
 the same as a missing `check-unfinished-work.sh` verdict line above.
diff --git a/stats/internal/guard/check_visual_verify_dispatched_test.go b/stats/internal/guard/check_visual_verify_dispatched_test.go
index 98b71c2b..a202511e 100644
--- a/stats/internal/guard/check_visual_verify_dispatched_test.go
+++ b/stats/internal/guard/check_visual_verify_dispatched_test.go
@@ -331,7 +331,7 @@ func TestCheckVisualVerifyDispatched(t *testing.T) {
 		{"case 8d: an empty change name exits 2",
 			vvdCase{args: func(wt, base string) []string { return []string{wt, "", base} }}, 2, "",
 			func(vvdResult) string {
-				return p + "usage: check-visual-verify-dispatched.sh <worktree> <change-name> <merge-base>\n"
+				return p + "usage: check-visual-verify-dispatched.sh <worktree> <change-name> <merge-base> [<canonical-worktree>]\n"
 			}, nil},
 		{"case 8e: an empty merge-base exits 2",
 			vvdCase{args: func(wt, _ string) []string { return []string{wt, "demo", ""} }}, 2, "",
@@ -502,3 +502,62 @@ func TestCheckVisualVerifyDispatchedInProcessTrigger(t *testing.T) {
 		})
 	}
 }
+
+// TestCheckVisualVerifyDispatchedCanonicalUIPaths pins the kan-982
+// self-review fix: with a canonical worktree, a satellite whose own project
+// declares no section is matched against the canonical `ui paths` by its
+// repository's name, and the store is read under the canonical worktree.
+// Without the argument the same satellite is "not configured".
+func TestCheckVisualVerifyDispatchedCanonicalUIPaths(t *testing.T) {
+	t.Parallel()
+	masters, err := vvdMasters()
+	if err != nil {
+		t.Fatal(err)
+	}
+	m := masters["plain-docs"] // no section; its change touches docs/note.md
+	for _, c := range []struct {
+		name, uiPaths, rows string
+		canonical           bool
+		code                int
+		out                 string
+	}{
+		{"no canonical worktree: not configured", "`gymie-frontend/**`", "[]", false, 0, vvdNotConfig},
+		{"canonical ui paths name the satellite: no verifier is MISSING", "`gymie-frontend/**`", "[]", true, 1, vvdMissing},
+		{"canonical ui paths name the satellite: a verifier is OK", "`gymie-frontend/**`", string(vvdRows("visual-verify", "verifier", "completed")), true, 0, vvdOK},
+		{"canonical ui paths name another repository: no UI paths touched", "`other/**`", "[]", true, 0, vvdNoUI},
+		{"a canonical worktree that is not a directory cannot answer", "`gymie-frontend/**`", "[]", true, 2, ""},
+	} {
+		t.Run(c.name, func(t *testing.T) {
+			t.Parallel()
+			top, err := filepath.EvalSymlinks(t.TempDir())
+			if err != nil {
+				t.Fatal(err)
+			}
+			wt, canon, bin := top+"/gymie-frontend", top+"/gymie", top+"/bin"
+			gdcCopyTree(t, m.dir, wt)
+			writeFile(t, canon+"/.flow/project.md", strings.Replace(vvdSection, "`app/src/**`", c.uiPaths, 1))
+			writeExec(t, bin+"/flow", vvdFlowRows(c.rows, "[]"))
+			env := Env{Dir: top, Getenv: func(k string) string {
+				if k == "PATH" {
+					return bin + ":" + os.Getenv("PATH")
+				}
+				return os.Getenv(k)
+			}}
+			args := []string{wt, "demo", m.base}
+			if c.canonical {
+				args = append(args, canon)
+			}
+			if c.code == 2 {
+				args = append(args[:3], top+"/no-such-worktree")
+			}
+			var out, errb bytes.Buffer
+			code := Registry["check-visual-verify-dispatched"](args, env, &out, &errb)
+			if code != c.code || out.String() != c.out {
+				t.Fatalf("exit %d stdout %q stderr %q, want %d %q", code, out.String(), errb.String(), c.code, c.out)
+			}
+			if calls, _ := os.ReadFile(bin + "/args"); c.code == 1 && !strings.Contains(string(calls), "record dispatches -change demo -C "+canon) {
+				t.Errorf("store calls %q, want the dispatches read under the canonical worktree %s", calls, canon)
+			}
+		})
+	}
+}
diff --git a/stats/internal/guard/visualverifydispatched.go b/stats/internal/guard/visualverifydispatched.go
index 6b928537..70666cbd 100644
--- a/stats/internal/guard/visualverifydispatched.go
+++ b/stats/internal/guard/visualverifydispatched.go
@@ -37,7 +37,7 @@ func checkVisualVerifyDispatched(args []string, env Env, stdout, stderr io.Write
 		}
 		return ""
 	}
-	worktreeArg, name, base := arg(0), arg(1), arg(2)
+	worktreeArg, name, base, canonicalArg := arg(0), arg(1), arg(2), arg(3)
 	if worktreeArg == "" || !isDir(pcAbs(env, worktreeArg)) {
 		shown := worktreeArg
 		if shown == "" {
@@ -47,7 +47,7 @@ func checkVisualVerifyDispatched(args []string, env Env, stdout, stderr io.Write
 		return 2
 	}
 	if name == "" {
-		fmt.Fprintf(stderr, "%susage: check-visual-verify-dispatched.sh <worktree> <change-name> <merge-base>\n", vvdPrefix)
+		fmt.Fprintf(stderr, "%susage: check-visual-verify-dispatched.sh <worktree> <change-name> <merge-base> [<canonical-worktree>]\n", vvdPrefix)
 		return 2
 	}
 	if base == "" {
@@ -89,7 +89,51 @@ func checkVisualVerifyDispatched(args []string, env Env, stdout, stderr io.Write
 	// diff's paths go to visualTrigger (an empty diff is one empty line, as
 	// the bash's `printf '%s\n'` piped it), its stdout and stderr discarded,
 	// its three exit codes read as-is.
-	triggerExit := visualTrigger(env, worktree, func() ([]string, error) { return strings.Split(changed, "\n"), nil }, io.Discard, io.Discard)
+	paths := strings.Split(changed, "\n")
+	triggerExit := visualTrigger(env, worktree, func() ([]string, error) { return paths, nil }, io.Discard, io.Discard)
+
+	// A SATELLITE IS ALSO MATCHED AGAINST THE CANONICAL `ui paths` (kan-982
+	// self-review): a cross-repo change's frontend worktree may declare no
+	// section of its own while the canonical project names its files as
+	// `<repo>/**`, so asking each worktree alone passed integrate with every
+	// verifier blocked. Its paths are prefixed with its repository's name --
+	// its main checkout's basename, check-visual-trigger.sh's --repo -- and
+	// the better of the two answers stands: 0 over 1 over 2. The store is
+	// read under the canonical worktree, the project the run's rows were
+	// recorded in; a second repository's worktree resolves another project
+	// key (unfinishedwork.go's anchor carries the reasoning).
+	anchor := worktree
+	if canonicalArg != "" {
+		canonical, err := filepath.EvalSymlinks(pcAbs(env, canonicalArg))
+		if err != nil || !isDir(canonical) {
+			fmt.Fprintf(stderr, "%scanonical worktree is not a directory: %s\n", vvdPrefix, canonicalArg)
+			return 2
+		}
+		if canonical != worktree {
+			anchor = canonical
+			if triggerExit != 0 {
+				common, ok := capture(envGit(env)("-C", worktree, "rev-parse", "--path-format=absolute", "--git-common-dir"))
+				if !ok {
+					fmt.Fprintf(stderr, "%sgit cannot name the repository of %s — cannot answer\n", vvdPrefix, worktree)
+					return 2
+				}
+				repo := filepath.Base(filepath.Dir(strings.TrimSpace(common))) + "/"
+				prefixed := make([]string, len(paths))
+				for i, p := range paths {
+					if p != "" {
+						p = repo + p
+					}
+					prefixed[i] = p
+				}
+				switch visualTrigger(env, canonical, func() ([]string, error) { return prefixed, nil }, io.Discard, io.Discard) {
+				case 0:
+					triggerExit = 0
+				case 1:
+					triggerExit = 1
+				}
+			}
+		}
+	}
 	switch triggerExit {
 	case 2:
 		// The project declares no `## visual verification` section at all.
@@ -104,7 +148,7 @@ func checkVisualVerifyDispatched(args []string, env Env, stdout, stderr io.Write
 	// A store the call could not reach is "cannot answer", never "no rows":
 	// an outage that read as zero dispatches would pass every round it was
 	// blind to.
-	rows, err := pfdRead(env, env.Dispatches, "dispatches", name, worktree)
+	rows, err := pfdRead(env, env.Dispatches, "dispatches", name, anchor)
 	if err != nil {
 		fmt.Fprintf(stderr, "%sflow record dispatches failed for '%s' — cannot answer\n", vvdPrefix, name)
 		return 2
@@ -128,7 +172,7 @@ func checkVisualVerifyDispatched(args []string, env Env, stdout, stderr io.Write
 			_ = env.Verdict(name, guardName, worktree, verdict)
 			return
 		}
-		pcFlow(env, false, "record", "verdict", "-change", name, "-guard", guardName, "-worktree", worktree, "-verdict", verdict, "-C", worktree)
+		pcFlow(env, false, "record", "verdict", "-change", name, "-guard", guardName, "-worktree", worktree, "-verdict", verdict, "-C", anchor)
 	}
 	if found {
 		verdict := "VISUAL-VERIFY-OK: UI paths touched and a completed (or ended-but-unclosed) verifier dispatch is recorded for '" + name + "'"
@@ -141,7 +185,7 @@ func checkVisualVerifyDispatched(args []string, env Env, stdout, stderr io.Write
 	// operator saw that row and its reason at the plan gate, so the skip is
 	// not silent (kan-30's failure) and is not this guard's to second-guess.
 	// The read fails closed exactly as the dispatch read does.
-	decisions, err := pfdRead(env, env.Decisions, "decisions", name, worktree)
+	decisions, err := pfdRead(env, env.Decisions, "decisions", name, anchor)
 	if err != nil {
 		fmt.Fprintf(stderr, "%sflow record decisions failed for '%s' — cannot answer\n", vvdPrefix, name)
 		return 2
@@ -163,7 +207,7 @@ func checkVisualVerifyDispatched(args []string, env Env, stdout, stderr io.Write
 	var priorErr error
 	if env.Verdicts != nil {
 		prior, priorErr = env.Verdicts(guardName)
-	} else if out, _, rc := pcFlow(env, false, "record", "verdicts", "-guard", guardName, "-false-positive", "-C", worktree); rc == 0 {
+	} else if out, _, rc := pcFlow(env, false, "record", "verdicts", "-guard", guardName, "-false-positive", "-C", anchor); rc == 0 {
 		prior = out
 	} else {
 		priorErr = fmt.Errorf("flow exit %d", rc)
````

  - [ ] **Step 4: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckVisualVerifyDispatched' -count=1`, then from the repository root `bash -n scripts/integrate-change.sh scripts/check-visual-verify-dispatched.sh`, `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh` — all exit 0 with `gofmt -l` printing nothing. Commit.

**Build:** green
**Files:** `stats/internal/guard/visualverifydispatched.go`, `stats/internal/guard/check_visual_verify_dispatched_test.go`, `scripts/check-visual-verify-dispatched.sh`, `scripts/integrate-change.sh`, `skills/flow-contracts/finish-contract-run1.md`
**Tests:** `TestCheckVisualVerifyDispatchedCanonicalUIPaths`
**Regression:** reverting the commit ignores the canonical worktree argument, so a satellite diff only the canonical `ui paths` name reads `VISUAL-VERIFY-OK: not configured` and integrate passes with no verifier.
**Baseline:** before=2 after=3
<!-- measured: grep -c '^func Test' stats/internal/guard/check_visual_verify_dispatched_test.go @ bc44caac (before; after with this task's diff applied in the scratch clone) -->
**Commit:** fix(guard): integrate's visual-verify check matches a satellite against the canonical ui paths
**After:** none

**Decision:** visual-dispatched-canonical

- [ ] 4. A load-flake fix closes after two more full test runs
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Read step 2 of **The loop** in `skills/flow/verify-fix-loop.md`: nothing runs the full list again after a fix for a test that passes alone, so each further load flake waits for the next full-suite round — the defect.
  - [ ] **Step 2: Add the paragraph.** Apply this diff:

````diff verified:applied with git apply in plan order to a scratch clone at bc44caac; check-verbatim-moves.sh (with this change's verbatim-moves.txt), check-references.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh and check-normative-inventory.sh exit 0
--- a/skills/flow/verify-fix-loop.md
+++ b/skills/flow/verify-fix-loop.md
@@ -34,6 +34,11 @@
    `fix-mutation:` line is checked against the fix diff the way **The fix round mutation-proves
    what it changed** (`skills/flow/review-panel-fix-round.md`) walks them — a mutant nothing
    kills is a new fixable defect this loop takes another round on.
+   **A task fixing a test that passes alone and fails only in the full suite closes after two
+   more full runs.** Such load flakes surface one per full-suite round, so after the task's commit
+   and before `close-task.sh` the parent runs the worktree's whole `## test` list twice more, and
+   every further failure those runs show is fixed in this same task — its implementer's further
+   commit carrying the same `Task-Id:` — never left to the next round.
 3. **Re-review the delta.** Open `flow.review-panel` again and run one dispatch: `primary` alone,
    reading `late-fix.diff` over the range since the panel's last clean close, written, paired,
    ceilinged and recorded as **The late-fix reduction** (`skills/flow/review-panel-late-fix.md`)
````

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-normative-inventory.sh` — all exit 0. Commit.

**Build:** green
**Files:** `skills/flow/verify-fix-loop.md`
**Tests:** **none** — one prose paragraph; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): a load-flake fix closes after two more full test runs
**After:** none

**Decision:** load-flake-two-more-runs

- [ ] 5. A determinism fix whose flake does not reproduce is dropped
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'reproduce' skills/flow/verify-fix-loop.md skills/flow/implement.md` — no line says what happens to a fix whose flake never failed without it — the defect: kan-982 Task 17 committed one (0 failures in 8 runs without it).
    <!-- measured: read from the kan-982 self-review findings file (machine-local, not re-runnable) @ 2026-10-15 -->
  - [ ] **Step 2: Add the paragraph.** Apply this diff:

````diff verified:applied with git apply in plan order to a scratch clone at bc44caac; check-verbatim-moves.sh (with this change's verbatim-moves.txt), check-references.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh and check-normative-inventory.sh exit 0
--- a/skills/flow/verify-fix-loop.md
+++ b/skills/flow/verify-fix-loop.md
@@ -39,6 +39,10 @@
    and before `close-task.sh` the parent runs the worktree's whole `## test` list twice more, and
    every further failure those runs show is fixed in this same task — its implementer's further
    commit carrying the same `Task-Id:` — never left to the next round.
+   **A determinism fix whose flake does not reproduce is not committed.** When the task's runs
+   reproducing a flake, sized as the PROOF RUNS paragraph (`skills/flow/implement.md`) sizes them,
+   see no failure without the fix, the fix is dropped and the task closes with no code change, its
+   report stating the measured rate — runs and failures — in place of a fix for a flake nobody saw.
 3. **Re-review the delta.** Open `flow.review-panel` again and run one dispatch: `primary` alone,
    reading `late-fix.diff` over the range since the panel's last clean close, written, paired,
    ceilinged and recorded as **The late-fix reduction** (`skills/flow/review-panel-late-fix.md`)
````

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh` — all exit 0. Commit.

**Build:** green
**Files:** `skills/flow/verify-fix-loop.md`
**Tests:** **none** — one prose paragraph; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): drop a determinism fix whose flake does not reproduce
**After:** Task 4

**Decision:** drop-unreproduced-determinism-fix

- [ ] 6. Motion frames carry their offset and two triggers compare at one sampling
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Read **The motion step** in `skills/flow/visual-verify-verifier.md`: frames are written `<nnn>.png` with no instant and nothing asks for the sibling trigger at the same sampling — the defect behind kan-982 Task 20's false "Back draws in-between frames".
  - [ ] **Step 2: Edit the paragraph.** Apply this diff:

````diff verified:applied with git apply in plan order to a scratch clone at bc44caac; check-verbatim-moves.sh (with this change's verbatim-moves.txt), check-references.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh and check-normative-inventory.sh exit 0
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index 8d10cfb3..a9e27b99 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -477,12 +477,19 @@ trigger until the view settles — a CDP screencast (`Page.startScreencast`, eac
 with `Page.screencastFrameAck`), or `page.screenshot` sampled back-to-back where no screencast is
 available — while driving the trigger (a tap, a navigation, a drag through `page.mouse` in small
 steps). Write the frames, in order, to `<changeRoot>/visual-verification/motion/<motion
-id>/<nnn>.png`, read every one, and judge the strip frame by frame: **every frame shows the view
+id>/<nnn>-<ms>.png`, `<ms>` the frame's offset in milliseconds from the trigger — the screencast
+frame's own timestamp, or the clock read as a sampled screenshot is taken — read every one, and
+judge the strip frame by frame: **every frame shows the view
 being left or the view being entered, never an empty, default or wrong-but-plausible rendering
 between them; chrome (a title, a bar) and content change in the same frame; a removed motion's
 strip shows no intermediate frame; a gesture's content follows the pointer in every frame and
 settles where the motion's end state says** — plus any frame rule the project's own instructions
 state. A frame that breaks one is a departure, named by its index and what it shows.
+**A difference between two triggers is judged at one sampling.** Before reporting that one
+trigger draws a frame its sibling does not — Back against the forward navigation, a gesture
+against the tap it replaces — record the sibling in the same session by the same capture method
+and compare the two strips by their `<ms>` offsets; a frame one strip lacks only because its
+sampling skipped that instant is no departure.
 
 11. **Write `<changeRoot>/visual-verification.md`** — one entry per view: its screenshot
     path, resolved by the same recursive search step 9 used, and what was seen; and, per composed
````

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh` — all exit 0. Commit.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`
**Tests:** **none** — prose in the verifier's motion step; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): name motion frames by offset and compare triggers at one sampling
**After:** none

**Decision:** motion-frame-offsets
