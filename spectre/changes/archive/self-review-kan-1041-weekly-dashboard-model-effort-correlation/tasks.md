# self-review-kan-1041-weekly-dashboard-model-effort-correlation

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the seven non-`big` findings of the self-review pass of
`kan-1041-weekly-dashboard-model-effort-correlation`, and repair the store rows and archived
records they already wrote wrong.

**Architecture:** one helper in the main-checkout hook's heredoc stripper; a round-low refusal and
an inline-row floor in the record CLI, with the dev-store rows repaired through the existing
`flowd` commands and the archived ledger and panel re-rendered; two skill sentences; one condition
in the matrix's surface walk; and a new bash tool for the live store copy.

**Spec:** `design.md` beside this file; the brainstorm record is
`.superpowers/sdd/2026-10-28-self-review-kan-1041-weekly-dashboard-model-effort-correlation-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject, no `Task-Id:` trailer, and `-m "Flow-Key: <your dispatch Key>"`. Never write `spectre/changes/` save the two archived records Task 3 re-renders (decision `archived-records-in-fix`) — this change's `verbatim-moves.txt` and the archived narrative's correction are already committed with the plan. Never merge or push.
- Every task works in this worktree `/Users/tweety53/Projects/agents-worktrees/kan-1041-weekly-dashboard-model-effort-correlation`, branch `spectre/kan-1041-weekly-dashboard-model-effort-correlation`; `**Files:**` paths are relative to its root. Never touch a main checkout, and never create a worktree.
- Each task's diff is the unified diff its step gives, applied with `git apply` from the worktree's root; all seven were applied in order at `52eee152` and their tests run. Save each by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean.
- Never stop, restart or recreate the protected `flowd` on `127.0.0.1:4173` or the `flow-postgres` container, and never drop the `flow` database. A repair writes the dev store only through the commands its step names, touches only the rows it names, runs once, and its output is quoted in the report.
- Every new test fails with its task's fix reverted: run it once against the reverted source before committing and state the failure in the report.

## Review Focus

- Task 1: a heredoc whose opener runs code (`bash <<'B'`), and an unquoted body running `$(` or a backtick, still keep the other heredoc's body scanned.
- Task 2: a back-filled `begin` never runs the live round check; every existing `TestRecordDispatch*` case, the `panel-fix-1` floor cases included, still passes.
- Task 7: `down` signals only a pid whose pidfile names its own binary, and drops only `flow_<name>_live`; 4173 and 4174 are refused before any side effect.

- [x] 1. Judge a data heredoc's rest without the other heredocs' bodies
  - [x] **Step 1: Reproduce.** Save the diff below, apply its test hunk alone (`git apply --include='stats/internal/repotest/protect_main_checkout_test.go' <the saved diff>`) and run `cd stats && go test ./internal/repotest/ -run '^TestProtectMainCheckoutReadsTwoDataHeredocsAsData$' -count=1`: it fails, `wanted allow, got deny` on the two-report case.
    <!-- measured: the test with hooks/protect-main-checkout.py at HEAD, go test -run TestProtectMainCheckoutReadsTwoDataHeredocsAsData @ branch spectre/kan-1041-weekly-dashboard-model-effort-correlation → FAIL wanted allow, got deny -->
  - [x] **Step 2: Apply the fix** — the rest (`git apply --exclude='stats/internal/repotest/protect_main_checkout_test.go' <the saved diff>`):

````diff verified:applied with git apply at 52eee152; go test ./internal/repotest -run TestProtectMainCheckout passes, the new test fails with the hook reverted, and the pass report's minimal reproduction then reads allow
diff --git a/hooks/protect-main-checkout.py b/hooks/protect-main-checkout.py
index 1dc73f5b..5c3df547 100755
--- a/hooks/protect-main-checkout.py
+++ b/hooks/protect-main-checkout.py
@@ -541,6 +541,34 @@ def _split_redirect(word):
     return [m.group(1), m.group(2)] if m else [word]
 
 
+def drop_data_bodies(lines):
+    """lines with the body of every heredoc whose text the shell runs nothing of removed — a
+    quoted delimiter's, or an unquoted one holding no `$(` or backtick — keeping its opener and
+    delimiter lines, so a `bash <<'E'` opener still reads as a runner. strip_data_heredocs judges
+    one body's rest through this: another heredoc's prose (a lone `.`, an unpaired backtick) is
+    no command of the rest (kan-1041 self-review: two report heredocs in one call were denied)."""
+    out = []
+    k = 0
+    while k < len(lines):
+        line = lines[k]
+        out.append(line)
+        k += 1
+        m = HEREDOC_OPEN_RE.search(line)
+        if not m:
+            continue
+        delim = m.group(1) or m.group(2) or m.group(3)
+        tabs = line[m.start() : m.start() + 3] == "<<-"
+        end = next(
+            (j for j in range(k, len(lines)) if (lines[j].lstrip("\t") if tabs else lines[j]).rstrip("\r") == delim),
+            None,
+        )
+        if end is None or (m.group(3) and any(c in "\n".join(lines[k:end]) for c in ("$(", "`"))):
+            continue
+        out.append(lines[end])
+        k = end + 1
+    return out
+
+
 def strip_data_heredocs(command):
     """command with the body of a `cat`/`tee` heredoc to a named file removed — that body is
     data for a writer whose target the opening line already names, so a report line such as
@@ -600,7 +628,7 @@ def strip_data_heredocs(command):
             and not line.rstrip().endswith("\\")  # a continued opener line hides the rest
             and ">" not in seg  # output written to a file may be run later
             and not any(w.rsplit("/", 1)[-1] in RUNNER_WORDS for w in OPENER_WORD_SPLIT_RE.split(seg))
-            and not may_run_code("\n".join(lines[:k] + lines[end + 1 :]), None)
+            and not may_run_code("\n".join(drop_data_bodies(lines[:k]) + drop_data_bodies(lines[end + 1 :])), None)
         ):
             body = "\n".join(lines[k:end]).replace("\\`", "")
             if m.group(3) and any(c in body for c in ("$(", "`")):
@@ -621,7 +649,7 @@ def strip_data_heredocs(command):
             continue
         if target.startswith("/dev/") or "(" in target:
             continue
-        rest = "\n".join(lines[:k] + lines[end + 1 :])
+        rest = "\n".join(drop_data_bodies(lines[:k]) + drop_data_bodies(lines[end + 1 :]))
         if may_run_code(rest, target):
             if runs_as_program(rest, target) and not (m.group(3) and any(c in "\n".join(lines[k:end]) for c in ("$(", "`"))):
                 out.extend(program_git_lines("\n".join(lines[k:end])))
diff --git a/stats/internal/repotest/protect_main_checkout_test.go b/stats/internal/repotest/protect_main_checkout_test.go
index 5fa199ce..a93a1760 100644
--- a/stats/internal/repotest/protect_main_checkout_test.go
+++ b/stats/internal/repotest/protect_main_checkout_test.go
@@ -653,6 +653,33 @@ func TestProtectMainCheckoutReadsADataHeredocBodyAsData(t *testing.T) {
 	f.bash("unbalanced ) in body", "deny", "{ROOT}", "flow x <<EOF\n$(case a in a) rm {MAIN}/f.txt;; esac)\nEOF")
 }
 
+// TestProtectMainCheckoutReadsTwoDataHeredocsAsData pins the kan-1041
+// self-review: each data heredoc's rest is judged without the other's body,
+// so a lone `.` in one and an unpaired backtick in the other no longer keep
+// both scanned; an unquoted body that runs code still counts.
+func TestProtectMainCheckoutReadsTwoDataHeredocsAsData(t *testing.T) {
+	t.Parallel()
+	root, err := filepath.EvalSymlinks(t.TempDir())
+	if err != nil {
+		t.Fatal(err)
+	}
+	f := &pmFixture{t: t, home: t.TempDir(), vars: map[string]string{"ROOT": root}}
+	main := root + "/main"
+	f.vars["MAIN"], f.vars["SIB"] = main, root+"/sib"
+	f.git("init", "-q", "--bare", root+"/origin.git")
+	f.git("clone", "-q", root+"/origin.git", main)
+	f.write(main+"/f.txt", "a\n")
+	f.git("-C", main, "add", "f.txt")
+	f.git("-C", main, "commit", "-q", "-m", "init")
+	f.git("-C", main, "branch", "-M", "main")
+	f.git("-C", main, "push", "-q", "-u", "origin", "main")
+	f.git("-C", main, "remote", "set-head", "origin", "main")
+
+	f.bash("two report heredocs, a lone dot and an unpaired backtick", "allow", "{MAIN}", "cat > {SIB}/a.md <<'E'\n.\nE\ncat > {SIB}/b.md <<'E'\n`> .\nE")
+	f.bash("a second body that runs the first's file keeps it scanned", "deny", "{MAIN}", "cat > {SIB}/x.sh <<'A'\nrm {MAIN}/f.txt\nA\ncat > {SIB}/y <<B\n$(bash {SIB}/x.sh)\nB")
+	f.bash("a second heredoc fed to a shell keeps the first scanned", "deny", "{MAIN}", "cat > {SIB}/x.sh <<'A'\nrm {MAIN}/f.txt\nA\nbash <<'B'\n{SIB}/x.sh\nB")
+}
+
 // TestProtectMainCheckoutReadsASingleQuotedBacktickAsLiteral pins the
 // kan-1039 self-review: a backtick inside single quotes after a data
 // heredoc is a plain character, so the heredoc body stays data; an
````

  - [x] **Step 3: Verify.** `cd stats && gofmt -l internal/repotest && go vet ./internal/repotest && go test ./internal/repotest/ -run '^TestProtectMainCheckout' -count=1` — green; `scripts/check-python-suppressions.sh` exits 0.
  - [x] **Step 4: Commit.** The archived narrative's corrected cause (`spectre/changes/archive/kan-1041-weekly-dashboard-model-effort-correlation/narrative.md` line 16) is already in the plan commit.

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `stats/internal/repotest/protect_main_checkout_test.go`
**Tests:** `TestProtectMainCheckoutReadsTwoDataHeredocsAsData`
**Regression:** `TestProtectMainCheckoutReadsTwoDataHeredocsAsData` — reverting the commit denies the two-report write again.
**Baseline:** before=24 after=25
<!-- measured: cat stats/internal/repotest/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1041-weekly-dashboard-model-effort-correlation -->
**Commit:** fix(hooks): judge a data heredoc's rest without the other heredocs' bodies
**After:** none

- [x] 2. Refuse a panel-fix key one round low and re-key the two rows
  - [x] **Step 1: Reproduce.** Save the diff below, apply its test hunk alone (`git apply --include='stats/cmd/flow/record_test.go' <the saved diff>`) and run `cd stats && go test ./cmd/flow/ -run '^TestRecordDispatchRefusesAPanelFixKeyedARoundLow$' -count=1`: the four round-low cases fail `want 2`.
    <!-- measured: the test with record.go at HEAD @ branch spectre/kan-1041-weekly-dashboard-model-effort-correlation → 4 cases FAIL want 2 -->
  - [x] **Step 2: Apply the fix** — the rest (`git apply --exclude='stats/cmd/flow/record_test.go' <the saved diff>`), decision `panel-fix-round-refusal`:

````diff verified:applied with git apply at 52eee152; go vet ./cmd/flow and go test ./cmd/flow -count=1 pass, the new test fails with record.go reverted
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index 972cceaa..2df1c94c 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -1180,6 +1180,13 @@ func runRecordDispatchBegin(ctx context.Context, args []string, stdout, stderr i
 		}
 	}
 
+	// A live panel-fix begin after round r's reviewers began fixes round r,
+	// so its key is panel-fix-<r+1> (panelFixRoundTaken). A back-fill runs
+	// the same check against its end before it calls begin.
+	if *startedAt == "" && !panelFixKeyedARoundLow(ctx, f, projectKey, *role, *key, time.Now(), stderr) {
+		return 2
+	}
+
 	// A visual verifier is a mutator: it builds the stack from the tree and
 	// writes <changeRoot>/visual-verification/, so it never runs beside an
 	// open implementer, panel-fix or red-partner row of the change, nor one
@@ -1491,21 +1498,21 @@ func runRecordDispatchBackfill(ctx context.Context, args []string, stdout, stder
 	// panel-fix row at its round's reviewer ends, or for a full-suite fix
 	// at the newest reviewer end not after its commit (panelFixFloor;
 	// kan-1039 self-review: full-suite-fix-1 claimed the whole panel wait).
+	if err := finishRecordConnFlags(fset, &f); err != nil {
+		fmt.Fprintf(stderr, "flow: %v\n", err)
+		return 2
+	}
+	projectKey, _, err := fallback.ProjectKey(f.dir)
+	if err != nil {
+		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
+		return 1
+	}
 	var derived []string
 	if *v["agent-id"] == "inline" && *v["commit"] != "" {
 		if *v["started-at"] != "" || *v["ended-at"] != "" {
 			fmt.Fprintln(stderr, "flow: an inline back-fill with -commit takes neither -started-at nor -ended-at -- both derive from the commit")
 			return 2
 		}
-		if err := finishRecordConnFlags(fset, &f); err != nil {
-			fmt.Fprintf(stderr, "flow: %v\n", err)
-			return 2
-		}
-		projectKey, _, err := fallback.ProjectKey(f.dir)
-		if err != nil {
-			fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
-			return 1
-		}
 		roots := mutationRoots(ctx, f, projectKey)
 		endT, ok := commitTime(*v["commit"], roots)
 		var startT time.Time
@@ -1542,6 +1549,9 @@ func runRecordDispatchBackfill(ctx context.Context, args []string, stdout, stder
 		fmt.Fprintf(stderr, "flow: -started-at %s is later than -ended-at %s\n", *v["started-at"], *v["ended-at"])
 		return 2
 	}
+	if !panelFixKeyedARoundLow(ctx, f, projectKey, *v["role"], *v["key"], ended, stderr) {
+		return 2
+	}
 	if _, ok := checkDispatchEndFields(*v["outcome"], *v["cause"], *v["tokens"], stderr); !ok {
 		return 2
 	}
@@ -3537,6 +3547,37 @@ func panelFixFloor(ctx context.Context, f recordIdentityFlags, projectKey, role,
 	return floor, !floor.IsZero()
 }
 
+// panelFixKeyedARoundLow refuses, reporting false, a panel-fix-<r>[-<n>]
+// key once a reviewer row of round r began before `before` -- the fix's
+// end, or now on a live begin: that fix answers round r's findings, so its
+// round is r+1 (skills/flow/review-panel-fix-round.md; kan-1041
+// self-review: panel-fix-5 and panel-fix-7 fixed rounds 5 and 7). Any other
+// role or key, and a failed record read, pass.
+func panelFixKeyedARoundLow(ctx context.Context, f recordIdentityFlags, projectKey, role, key string, before time.Time, stderr io.Writer) bool {
+	rest, ok := strings.CutPrefix(key, "panel-fix-")
+	if role != "panel-fix" || !ok {
+		return true
+	}
+	digits, _, _ := strings.Cut(rest, "-")
+	r, err := strconv.Atoi(digits)
+	if err != nil || r < 1 {
+		return true
+	}
+	run, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.Run, error) {
+		return cl.GetRunRecord(ctx, projectKey, f.change)
+	})
+	if err != nil {
+		return true
+	}
+	for _, d := range run.Dispatches {
+		if n, ok := records.PanelDispatchRound(d); ok && n == r && d.StartedAt.Before(before) {
+			fmt.Fprintf(stderr, "flow: -key %q: panel round %d's reviewers began before this fix ended, so it fixes round %d's findings and is panel-fix-%d (skills/flow/review-panel-fix-round.md)\n", key, r, r, r+1)
+			return false
+		}
+	}
+	return true
+}
+
 // mutationRoots is the -C repository's top level plus every key of the
 // worktrees map in the change's state record.
 func mutationRoots(ctx context.Context, f recordIdentityFlags, projectKey string) []string {
diff --git a/stats/cmd/flow/record_test.go b/stats/cmd/flow/record_test.go
index 737d4153..3495ec24 100644
--- a/stats/cmd/flow/record_test.go
+++ b/stats/cmd/flow/record_test.go
@@ -6769,3 +6769,72 @@ func TestRecordDispatchRefusesPanelFixZero(t *testing.T) {
 		}
 	}
 }
+
+// TestRecordDispatchRefusesAPanelFixKeyedARoundLow pins the kan-1041
+// self-review fix: a panel-fix-<r> row is refused once round r's reviewers
+// began before the fix ended -- it fixes round r and is panel-fix-<r+1> --
+// on an inline back-fill, an explicit back-fill and a live begin alike; the
+// right key, a back-filled begin and a failed record read pass.
+func TestRecordDispatchRefusesAPanelFixKeyedARoundLow(t *testing.T) {
+	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
+	repo := recordRepo(t, "kan-258")
+	isolatedStateRoot(t)
+	fixtureCommit(t, repo, "2026-10-28T02:40:00Z")
+	sha := fixtureCommit(t, repo, "2026-10-28T02:48:46Z")
+	record := `{"change":"kan-258","dispatches":[` +
+		`{"key":"panel-5-primary","role":"reviewer","slot":"primary","startedAt":"2026-10-28T02:45:33Z","endedAt":"2026-10-28T02:47:03Z"}]}`
+	var posts int
+	recordFails := false
+	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
+		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/records/") {
+			if recordFails {
+				w.WriteHeader(http.StatusInternalServerError)
+				return
+			}
+			_, _ = w.Write([]byte(record))
+			return
+		}
+		if r.Method != http.MethodPost {
+			w.WriteHeader(http.StatusNotFound)
+			return
+		}
+		posts++
+		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"panel-fix","model":"opus"}`))
+	}))
+	t.Cleanup(srv.Close)
+	conn := []string{"-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "kan-258", "-role", "panel-fix", "-model", "opus"}
+	backfill := func(key string, extra ...string) []string {
+		return append(append(append([]string{"record", "dispatch", "backfill"}, conn...), "-key", key,
+			"-session-token", "mf-record-panel-fix-round", "-outcome", "completed"), extra...)
+	}
+	begin := func(key string, extra ...string) []string {
+		return append(append(append([]string{"record", "dispatch", "begin"}, conn...), "-key", key,
+			"-session-token", "mf-record-panel-fix-round", "-effort", "default"), extra...)
+	}
+	for _, tc := range []struct {
+		name        string
+		args        []string
+		recordFails bool
+		want        int
+	}{
+		{"inline back-fill a round low", backfill("panel-fix-5", "-agent-id", "inline", "-commit", sha), false, 2},
+		{"inline back-fill chunk a round low", backfill("panel-fix-5-2", "-agent-id", "inline", "-commit", sha), false, 2},
+		{"explicit back-fill a round low", backfill("panel-fix-5", "-started-at", "2026-10-28T02:47:03Z", "-ended-at", "2026-10-28T02:48:46Z"), false, 2},
+		{"live begin a round low", begin("panel-fix-5"), false, 2},
+		{"inline back-fill on the fix round", backfill("panel-fix-6", "-agent-id", "inline", "-commit", sha), false, 0},
+		{"live begin on the fix round", begin("panel-fix-6"), false, 0},
+		{"back-fill ending before the round began", backfill("panel-fix-5", "-started-at", "2026-10-28T02:40:00Z", "-ended-at", "2026-10-28T02:45:00Z"), false, 0},
+		{"a failed record read", backfill("panel-fix-5", "-agent-id", "inline", "-commit", sha), true, 0},
+	} {
+		posts = 0
+		recordFails = tc.recordFails
+		var stdout, stderr bytes.Buffer
+		code := run(context.Background(), tc.args, strings.NewReader(""), &stdout, &stderr)
+		if code != tc.want {
+			t.Errorf("%s: exit %d, want %d; stderr:\n%s", tc.name, code, tc.want, stderr.String())
+		}
+		if tc.want == 2 && (posts != 0 || !strings.Contains(stderr.String(), "panel-fix-6")) {
+			t.Errorf("%s: posts %d, stderr %q; want none and a message naming panel-fix-6", tc.name, posts, stderr.String())
+		}
+	}
+}
````

  - [x] **Step 3: Verify.** `cd stats && gofmt -l cmd/flow && go vet ./cmd/flow && go test ./cmd/flow/ -run '^TestRecordDispatch' -count=1` — green.
  - [x] **Step 4: Repair the rows**, in this order, from `stats/` with `FLOWD_DSN='postgres://flow:flow@localhost:5433/flow?sslmode=disable'` (decision `repair-existing-commands`):
    - `go run ./cmd/flowd rekey-dispatch 8334 panel-fix-6`
    - `go run ./cmd/flowd retime-dispatch 8334 2026-10-28T02:47:03Z 2026-10-28T02:48:46Z 21b73d9460c7912a9866b4b8f187adcb1aff45e0`
    - `go run ./cmd/flowd rekey-dispatch 8344 panel-fix-8`
    - `go run ./cmd/flowd retime-dispatch 8344 2026-10-28T03:04:14Z 2026-10-28T03:05:29Z 3e7e8429c0d22c277dde87b7463f21a470ec0720`
    - `docker exec flow-postgres psql -U flow -d flow -c "update panel_passes set note = 'fix round re-run of mutation over panel-fix-6' where id = 2535 and note = 'fix round re-run of mutation over panel-fix-5'"` — it prints `UPDATE 1`.
    <!-- measured: docker exec flow-postgres psql -U flow -d flow -tAc "select id, dispatch_key, started_at, ended_at, commit_sha from dispatches where id in (8334,8344)" and "select id, note from panel_passes where id = 2535" @ branch spectre/kan-1041-weekly-dashboard-model-effort-correlation (live dev store, read-only) -->
    Then `docker exec flow-postgres psql -U flow -d flow -tAc "select id, dispatch_key, started_at from dispatches where id in (8334,8344) order by id"` shows `panel-fix-6` from 02:47:03 and `panel-fix-8` from 03:04:14. Task 3 re-renders the archived records after its own repair.
  - [x] **Step 5: Commit.**

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchRefusesAPanelFixKeyedARoundLow`
**Regression:** `TestRecordDispatchRefusesAPanelFixKeyedARoundLow` — reverting the commit admits `panel-fix-5` after round 5's reviewers began.
**Baseline:** before=349 after=350
<!-- measured: cat stats/cmd/flow/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1041-weekly-dashboard-model-effort-correlation -->
**Commit:** fix(flow): refuse a panel-fix key one round below the round it fixes
**After:** none

- [x] 3. Floor an inline back-fill at the change's other inline rows and re-render the records
  - [x] **Step 1: Reproduce.** Save the diff below, apply its test hunk alone (`git apply --include='stats/cmd/flow/record_test.go' <the saved diff>`) and run `cd stats && go test ./cmd/flow/ -run '^TestRecordDispatchBackfillFloorsAnInlineRowAtTheOtherInlineRows$' -count=1`: it fails, the begin's `startedAt` the commit-derived `2026-10-28T02:48:46Z`.
    <!-- measured: the test with Task 2's record.go @ branch spectre/kan-1041-weekly-dashboard-model-effort-correlation → FAIL startedAt 2026-10-28T02:48:46Z -->
  - [x] **Step 2: Apply the fix** — the rest (`git apply --exclude='stats/cmd/flow/record_test.go' <the saved diff>`), decision `inline-floor`:

````diff verified:applied with git apply after Task 2's diff at 52eee152; go vet ./cmd/flow and go test ./cmd/flow -count=1 pass, the new test fails with the record.go hunks reverted
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index 2df1c94c..e382172f 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -1526,6 +1526,9 @@ func runRecordDispatchBackfill(ctx context.Context, args []string, stdout, stder
 		if floor, ok := panelFixFloor(ctx, f, projectKey, *v["role"], *v["key"], endT); ok && floor.After(startT) && !floor.After(endT) {
 			startT = floor
 		}
+		if floor, ok := inlineFloor(ctx, f, projectKey, *v["key"], endT); ok && floor.After(startT) {
+			startT = floor
+		}
 		*v["started-at"] = startT.UTC().Format(time.RFC3339)
 		*v["ended-at"] = endT.UTC().Format(time.RFC3339)
 		derived = []string{"-started-at", *v["started-at"], "-ended-at", *v["ended-at"]}
@@ -3578,6 +3581,27 @@ func panelFixKeyedARoundLow(ctx context.Context, f recordIdentityFlags, projectK
 	return true
 }
 
+// inlineFloor is the latest end, not after endT, among the change's other
+// inline rows: an inline back-fill's derived window never contains another
+// inline row of the run, since the parent's main thread belongs to one row
+// at a time (kan-1041 self-review: task-13-implementer's window held all of
+// verify-fix-3). A failed record read, or no such row, has no floor.
+func inlineFloor(ctx context.Context, f recordIdentityFlags, projectKey, key string, endT time.Time) (time.Time, bool) {
+	run, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.Run, error) {
+		return cl.GetRunRecord(ctx, projectKey, f.change)
+	})
+	if err != nil {
+		return time.Time{}, false
+	}
+	var floor time.Time
+	for _, d := range run.Dispatches {
+		if d.AgentID == "inline" && d.Key != key && d.EndedAt != nil && !d.EndedAt.After(endT) && d.EndedAt.After(floor) {
+			floor = *d.EndedAt
+		}
+	}
+	return floor, !floor.IsZero()
+}
+
 // mutationRoots is the -C repository's top level plus every key of the
 // worktrees map in the change's state record.
 func mutationRoots(ctx context.Context, f recordIdentityFlags, projectKey string) []string {
diff --git a/stats/cmd/flow/record_test.go b/stats/cmd/flow/record_test.go
index 3495ec24..2269e079 100644
--- a/stats/cmd/flow/record_test.go
+++ b/stats/cmd/flow/record_test.go
@@ -6838,3 +6838,64 @@ func TestRecordDispatchRefusesAPanelFixKeyedARoundLow(t *testing.T) {
 		}
 	}
 }
+
+// TestRecordDispatchBackfillFloorsAnInlineRowAtTheOtherInlineRows pins the
+// kan-1041 self-review fix: an inline back-fill's derived start never comes
+// before the latest end, not after its commit, among the change's other
+// inline rows, so its window holds no other inline row; a non-inline row,
+// one ending after the commit and a failed record read set no floor.
+func TestRecordDispatchBackfillFloorsAnInlineRowAtTheOtherInlineRows(t *testing.T) {
+	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
+	repo := recordRepo(t, "kan-258")
+	isolatedStateRoot(t)
+	fixtureCommit(t, repo, "2026-10-28T02:48:46Z")
+	sha := fixtureCommit(t, repo, "2026-10-28T03:02:22Z")
+	withInline := `{"change":"kan-258","dispatches":[` +
+		`{"key":"verify-fix-3","agentId":"inline","role":"verifier","startedAt":"2026-10-28T02:51:19Z","endedAt":"2026-10-28T03:02:17Z"},` +
+		`{"key":"panel-6-mutation","agentId":"a1","role":"reviewer","slot":"mutation","startedAt":"2026-10-28T02:52:00Z","endedAt":"2026-10-28T03:02:20Z"},` +
+		`{"key":"verify-fix-4","agentId":"inline","role":"verifier","startedAt":"2026-10-28T03:02:30Z","endedAt":"2026-10-28T03:08:00Z"}]}`
+	record := withInline
+	recordFails := false
+	var bodies []map[string]any
+	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
+		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/records/") {
+			if recordFails {
+				w.WriteHeader(http.StatusInternalServerError)
+				return
+			}
+			_, _ = w.Write([]byte(record))
+			return
+		}
+		if r.Method != http.MethodPost {
+			w.WriteHeader(http.StatusNotFound)
+			return
+		}
+		raw, _ := readAll(r)
+		var b map[string]any
+		_ = json.Unmarshal(raw, &b)
+		bodies = append(bodies, b)
+		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"implementer","model":"opus"}`))
+	}))
+	t.Cleanup(srv.Close)
+	for _, tc := range []struct {
+		name, record, want string
+		recordFails        bool
+	}{
+		{"floored at the other inline row's end", withInline, "2026-10-28T03:02:17Z", false},
+		{"no other inline row", `{"change":"kan-258","dispatches":[]}`, "2026-10-28T02:48:46Z", false},
+		{"a failed record read", withInline, "2026-10-28T02:48:46Z", true},
+	} {
+		bodies = nil
+		record, recordFails = tc.record, tc.recordFails
+		args := []string{"record", "dispatch", "backfill", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
+			"-change", "kan-258", "-role", "implementer", "-model", "opus", "-key", "task-13-implementer", "-task", "13",
+			"-session-token", "mf-record-inline-floor", "-outcome", "completed", "-agent-id", "inline", "-commit", sha}
+		var stdout, stderr bytes.Buffer
+		if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 0 {
+			t.Fatalf("%s: exit %d; stderr:\n%s", tc.name, code, stderr.String())
+		}
+		if len(bodies) != 2 || bodies[0]["startedAt"] != tc.want {
+			t.Errorf("%s: bodies = %v, want begin startedAt %s", tc.name, bodies, tc.want)
+		}
+	}
+}
````

  - [x] **Step 3: Verify.** `cd stats && gofmt -l cmd/flow && go vet ./cmd/flow && go test ./cmd/flow/ -run '^TestRecordDispatch' -count=1` — green.
  - [x] **Step 4: Repair the row**, from `stats/` with `FLOWD_DSN='postgres://flow:flow@localhost:5433/flow?sslmode=disable'`, after Task 2's repair: `go run ./cmd/flowd retime-dispatch 8340 2026-10-28T03:02:17Z 2026-10-28T03:02:22Z 8808872e6ba830208f2b21603036e673cd408195`.
    <!-- measured: docker exec flow-postgres psql -U flow -d flow -tAc "select id, dispatch_key, started_at, ended_at, commit_sha from dispatches where id in (8336,8340)" @ branch spectre/kan-1041-weekly-dashboard-model-effort-correlation (live dev store, read-only) -->
  - [x] **Step 5: Re-render the archived records** (decision `archived-records-in-fix`): `flow record render -kind all -repo /Users/tweety53/Projects/agents-worktrees/kan-1041-weekly-dashboard-model-effort-correlation -change kan-1041-weekly-dashboard-model-effort-correlation`, then `cp .superpowers/sdd/ledgers/kan-1041-weekly-dashboard-model-effort-correlation.md spectre/changes/archive/kan-1041-weekly-dashboard-model-effort-correlation/ledger.md` and `cp .superpowers/sdd/reviews/kan-1041-weekly-dashboard-model-effort-correlation-panel.md spectre/changes/archive/kan-1041-weekly-dashboard-model-effort-correlation/panel.md`. `grep -n 'panel-fix-[57]' spectre/changes/archive/kan-1041-weekly-dashboard-model-effort-correlation/ledger.md spectre/changes/archive/kan-1041-weekly-dashboard-model-effort-correlation/panel.md` prints nothing, and the ledger carries `Key: panel-fix-6` and `Key: panel-fix-8`.
  - [x] **Step 6: Commit.**

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `spectre/changes/archive/kan-1041-weekly-dashboard-model-effort-correlation/ledger.md`, `spectre/changes/archive/kan-1041-weekly-dashboard-model-effort-correlation/panel.md`
**Tests:** `TestRecordDispatchBackfillFloorsAnInlineRowAtTheOtherInlineRows`
**Regression:** `TestRecordDispatchBackfillFloorsAnInlineRowAtTheOtherInlineRows` — reverting the commit derives an inline window that contains another inline row.
**Baseline:** before=350 after=351
<!-- measured: cat stats/cmd/flow/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1041-weekly-dashboard-model-effort-correlation -->
**Commit:** fix(flow): floor an inline back-fill at the change's other inline rows
**After:** Task 2

- [x] 4. Drop the PIPESTATUS advice zsh cannot honour
  - [x] **Step 1: Apply** with `git apply`:

````diff verified:applied at 52eee152; check-verbatim-moves.sh 52eee152 then printed its two FAIL lines, which this change's verbatim-moves.txt acknowledges
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index b1d0be9f..5be11de3 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -96,7 +96,7 @@ its rebase branch changes what every later step reads, so it runs and is read fi
 **A guard's exit status is captured bare — a reader command's is not the guard's.** A guard's
 exit status read through a pipe is the pipe's, not the guard's: `check-panel-docs-only.sh
 <worktree> <merge-base> | tail` exits with `tail`'s 0 whatever verdict the guard reached. Capture
-the output to a file and read `$?` after, use `set -o pipefail`/`PIPESTATUS`, or run the guard
+the output to a file and read `$?` after, use `set -o pipefail`, or run the guard
 bare — so a red verdict cannot be laundered into green by the reader command.
 
 **Run the citation pre-check before rebuilding the dispatch context bundle below**:
````

  - [x] **Step 2: Verify.** `scripts/check-verbatim-moves.sh && scripts/check-references.sh && scripts/check-markdown-integrity.sh` — all exit 0.
  - [x] **Step 3: Commit.**

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** **none** — a prose edit; `check-verbatim-moves.sh` pins the sentence.
**Regression:** none — no test is declared.
**Commit:** fix(flow): drop the PIPESTATUS advice zsh leaves empty
**After:** none

- [x] 5. State the plan guards' invocation in document-fix
  - [x] **Step 1: Apply** with `git apply`:

````diff verified:applied at 52eee152; check-verbatim-moves.sh 52eee152 then printed its two FAIL lines, which this change's verbatim-moves.txt acknowledges; check-references.sh exits 0
diff --git a/skills/flow/document-fix.md b/skills/flow/document-fix.md
index b61a0934..4ed62d1f 100644
--- a/skills/flow/document-fix.md
+++ b/skills/flow/document-fix.md
@@ -123,6 +123,8 @@ flow tasks count -C <worktree> <name>
 
-**The appended plan passes the plan guards before it is committed.** Run `check-plan-shape.sh`
-and `check-plan-provenance.sh` over the appended `tasks.md` and handle each exit as **D. Basic
-Workflow #3 — Writing plans** (`skills/flow/brainstorm-planner.md`) handles it, before the
-planning commit below.
+**The appended plan passes the plan guards before it is committed.** Run
+`check-plan-shape.sh <changeRoot>/tasks.md`, a shipped guard among this skill's scripts, and the
+project's configured plan-provenance guard, at the path the project's `.flow/project.md` `## lint`
+section names — the project's own script, never one under this skill's scripts — and handle each
+exit as **D. Basic Workflow #3 — Writing plans** (`skills/flow/brainstorm-planner.md`) handles it,
+before the planning commit below.
 
````

  - [x] **Step 2: Verify.** `scripts/check-verbatim-moves.sh && scripts/check-references.sh && scripts/check-markdown-integrity.sh` — all exit 0.
  - [x] **Step 3: Commit.**

**Build:** green
**Files:** `skills/flow/document-fix.md`
**Tests:** **none** — a prose edit; `check-verbatim-moves.sh` pins the sentence.
**Regression:** none — no test is declared.
**Commit:** fix(flow): state the plan guards' invocation in document-fix
**After:** none

- [x] 6. Take a dark page's whole image as the matrix surface
  - [x] **Step 1: Reproduce.** Save the diff below, apply its test hunk alone (`git apply --include='stats/internal/guard/measure_visual_properties_test.go' <the saved diff>`) and run `cd stats && go test ./internal/guard/ -run '^TestMeasureVisualPropertiesElementsDarkPage$' -count=1`: it fails, `surface_top 200, want 0`.
    <!-- measured: the test with measurevisualproperties.go at HEAD @ branch spectre/kan-1041-weekly-dashboard-model-effort-correlation → FAIL surface_top 200, want 0 -->
  - [x] **Step 2: Apply the fix** — the rest (`git apply --exclude='stats/internal/guard/measure_visual_properties_test.go' <the saved diff>`), decision `dark-page-surface`:

````diff verified:applied with git apply at 52eee152; go test ./internal/guard -run TestMeasureVisualProperties passes, the new test fails with the fix reverted
diff --git a/stats/internal/guard/measure_visual_properties_test.go b/stats/internal/guard/measure_visual_properties_test.go
index 8e778381..125a783a 100644
--- a/stats/internal/guard/measure_visual_properties_test.go
+++ b/stats/internal/guard/measure_visual_properties_test.go
@@ -541,6 +541,25 @@ func TestMeasureVisualPropertiesElements(t *testing.T) {
 	}
 }
 
+// kan-1041 self-review: on a dark-theme page the bottom row already stops
+// the surface walk, so the whole image is the surface and its elements pair.
+func TestMeasureVisualPropertiesElementsDarkPage(t *testing.T) {
+	t.Parallel()
+	page := mvpPNG(t, 120, 200, func(x, y int) color.RGBA {
+		if x >= 20 && x < 60 && y >= 120 && y < 130 {
+			return color.RGBA{0xee, 0xee, 0xee, 0xff}
+		}
+		return color.RGBA{0x12, 0x12, 0x12, 0xff}
+	})
+	d := mvpJSON(t, "", page, page, "--scale", "1", "--props", "elements")
+	if top := mvpDig(d, "a", "elements", "surface_top"); top != float64(0) {
+		t.Fatalf("dark page: surface_top %v, want 0", top)
+	}
+	if paired := mvpDig(d, "delta", "elements_summary", "paired"); paired != float64(1) {
+		t.Fatalf("dark page: paired %v, want 1", paired)
+	}
+}
+
 // kan-1017 (sr11): `--props matrix` fills each paired element's columns in
 // one call, and a region running past the image's edge is clipped to it.
 func TestMeasureVisualPropertiesMatrix(t *testing.T) {
diff --git a/stats/internal/guard/measurevisualproperties.go b/stats/internal/guard/measurevisualproperties.go
index 8c934b61..eba6c662 100644
--- a/stats/internal/guard/measurevisualproperties.go
+++ b/stats/internal/guard/measurevisualproperties.go
@@ -1165,7 +1165,9 @@ func (e mvpElement) obj() pyObj {
 // measureElements is kan-1015's verifier matrix script (elems.py): the app
 // surface's top — the first row, walking column 3 up from the bottom, whose
 // mean channel is not above 200, plus one; 0 when that row is above row 60,
-// the whole image then the surface — and every ink run below it against the
+// the whole image then the surface, as it is when the bottom row itself is
+// dark and the walk moves no row (a dark-theme page; kan-1041 self-review:
+// every dark frame paired 0) — and every ink run below it against the
 // surface's modal colour (distance over 30), 8px in from each side.
 func (r *mvpRegion) measureElements() (int, [3]byte, []mvpElement) {
 	y := r.h - 1
@@ -1177,7 +1179,7 @@ func (r *mvpRegion) measureElements() (int, [3]byte, []mvpElement) {
 		y--
 	}
 	t := y + 1
-	if t < 60 {
+	if t < 60 || y == r.h-1 {
 		t = 0
 	}
 	x0, x1 := 8, r.w-8
````

  - [x] **Step 3: Verify.** `cd stats && gofmt -l internal/guard && go vet ./internal/guard && go test ./internal/guard/ -run '^TestMeasureVisualProperties' -count=1` — green.
  - [x] **Step 4: Commit.**

**Build:** green
**Files:** `stats/internal/guard/measurevisualproperties.go`, `stats/internal/guard/measure_visual_properties_test.go`
**Tests:** `TestMeasureVisualPropertiesElementsDarkPage`
**Regression:** `TestMeasureVisualPropertiesElementsDarkPage` — reverting the commit puts a dark page's surface top at the image height and pairs no element.
**Baseline:** before=355 after=356
<!-- measured: cat stats/internal/guard/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1041-weekly-dashboard-model-effort-correlation -->
**Commit:** fix(guard): take a dark page's whole image as the matrix surface
**After:** none

- [x] 7. Script the live check's copy of the dev store
  - [x] **Step 1: Apply** with `git apply` (it creates both scripts executable), decision `live-copy-bash`:

````diff verified:applied with git apply at 52eee152; scripts/test-live-store-copy.sh printed all cases passed, and failed with down's drop target set to flow or with the 4173 refusal removed; check-references.sh and check-workspace-isolation.sh exit 0
diff --git a/.flow/project.md b/.flow/project.md
index a55c1f59..9aedafc6 100644
--- a/.flow/project.md
+++ b/.flow/project.md
@@ -56,6 +56,14 @@ protected daemon. Neither target is isolated by the
 `## workspace isolation` section below — that section covers apply worktrees, and this stack is a
 single, main-checkout-only fixture instead.
 
+**A live check against real data runs on a copy of the dev store**, never on the store itself:
+`scripts/live-store-copy.sh up <name> [<port>]` copies the dev `flow` database into
+`flow_<name>_live`, builds `flowd` from the checkout it runs from and starts it on `<port>`
+(default 4175) against the copy, printing the `FLOW_ADDR`/`FLOW_RECORDS_ADDR` exports that point a
+session at it; `scripts/live-store-copy.sh down <name>` stops that daemon and drops only that copy.
+It refuses ports 4173 and 4174, and its header carries its exit codes. A design's `## Live check`
+names it rather than re-specifying the copy.
+
 **This is the stack `flow.visual-verify` (`skills/flow/visual-verify.md`) starts and stops.**
 That stage's step 5 probes a URL and, if nothing answers, "starts the stack from `## run`" — a
 project with more than one candidate stack has to say which one that means, so this paragraph is
diff --git a/scripts/live-store-copy.sh b/scripts/live-store-copy.sh
new file mode 100755
index 00000000..e59fd92f
--- /dev/null
+++ b/scripts/live-store-copy.sh
@@ -0,0 +1,84 @@
+#!/usr/bin/env bash
+# live-store-copy.sh up <name> [<port>] | down <name> — a design's `## Live check`
+# against real data without touching the dev stack (kan-1041 self-review: the
+# check was hand-built across about 15 calls).
+#
+#   up    copies the dev `flow` database, read-only on the source, into a
+#         scratch database flow_<name>_live (pg_dump | psql inside the
+#         flow-postgres container), builds flowd from this checkout and starts
+#         it on <port> (default 4175) against the copy, then prints the address
+#         and the FLOW_ADDR / FLOW_RECORDS_ADDR exports that point a session at it.
+#   down  stops that scratch flowd — only the pid its own pidfile names, and
+#         only when that pidfile names this script's binary — and drops
+#         flow_<name>_live. A second `down` is a no-op.
+#
+# Never the protected dev stack: port 4173 (the dev flowd) and 4174 (the
+# UI-test stack) are refused, and the only database dropped is
+# flow_<name>_live. State lives in /tmp/flow-live-<name>/; the daemon writes
+# its own pidfile at /tmp/flowd-<port>.pid (TMPDIR pinned, as ui-test-up does).
+#
+# Exit codes: 0 done, 1 a step failed (docker, copy, build, not answering),
+# 2 usage or a refusal.
+set -euo pipefail
+usage() { echo "usage: live-store-copy.sh up <name> [<port>] | down <name>" >&2; exit 2; }
+[ $# -ge 2 ] || usage
+cmd=$1 name=$2 port=${3:-4175}
+[[ $name =~ ^[a-z0-9][a-z0-9_]*$ ]] || { echo "live-store-copy: name '$name' is not [a-z0-9][a-z0-9_]*" >&2; exit 2; }
+db="flow_${name}_live"
+dir="/tmp/flow-live-$name"
+repo=$(cd "$(dirname "$0")/.." && pwd)
+pg() { docker exec -i flow-postgres "$@"; }
+port_ok() {
+  [[ $1 =~ ^[0-9]+$ ]] && [ "$1" -ge 1024 ] && [ "$1" -le 65535 ] && [ "$1" != 4173 ] && [ "$1" != 4174 ] \
+    || { echo "live-store-copy: port '$1' refused (1024-65535, never 4173 or 4174)" >&2; exit 2; }
+}
+case $cmd in
+up)
+  [ $# -le 3 ] || usage
+  port_ok "$port"
+  if [ -e "$dir" ]; then echo "live-store-copy: $dir exists -- run down $name first" >&2; exit 2; fi
+  if curl -sf "http://127.0.0.1:$port/api/v1/changes" >/dev/null 2>&1; then
+    echo "live-store-copy: port $port already answers" >&2; exit 2
+  fi
+  if [ -n "$(pg psql -U flow -d flow -tAc "select 1 from pg_database where datname='$db'")" ]; then
+    echo "live-store-copy: database $db exists -- run down $name first" >&2; exit 2
+  fi
+  mkdir -p "$dir/transcripts" "$dir/state"
+  echo "$port" >"$dir/port"
+  pg createdb -U flow "$db"
+  if ! pg pg_dump -U flow --no-owner --no-acl flow | pg psql -q -v ON_ERROR_STOP=1 -U flow -d "$db" >/dev/null; then
+    echo "live-store-copy: copy into $db failed" >&2; exit 1
+  fi
+  (cd "$repo/stats" && make -s web-dist && go build -o "$dir/flowd" ./cmd/flowd)
+  env -u FLOWD_JIRA_SITE -u FLOWD_JIRA_EMAIL -u FLOWD_JIRA_TOKEN \
+    FLOWD_PORT="$port" FLOWD_DSN="postgres://flow:flow@localhost:5433/$db?sslmode=disable" \
+    FLOW_TRANSCRIPTS_DIR="$dir/transcripts" FLOW_STATE_DIR="$dir/state" TMPDIR=/tmp \
+    nohup "$dir/flowd" >"$dir/log" 2>&1 &
+  for _ in $(seq 1 60); do
+    if curl -sf "http://127.0.0.1:$port/api/v1/changes" >/dev/null 2>&1; then
+      echo "http://127.0.0.1:$port"
+      echo "export FLOW_ADDR=http://127.0.0.1:$port FLOW_RECORDS_ADDR=http://127.0.0.1:$port"
+      exit 0
+    fi
+    sleep 1
+  done
+  echo "live-store-copy: flowd did not answer on :$port within 60s -- see $dir/log" >&2; exit 1
+  ;;
+down)
+  [ $# -eq 2 ] || usage
+  if [ -f "$dir/port" ]; then
+    port=$(cat "$dir/port")
+    port_ok "$port"
+    pidfile="/tmp/flowd-$port.pid"
+    if [ -f "$pidfile" ] && [ "$(sed -n 2p "$pidfile")" -ef "$dir/flowd" ]; then
+      pid=$(head -n1 "$pidfile")
+      kill "$pid" 2>/dev/null || true
+      for _ in $(seq 1 30); do kill -0 "$pid" 2>/dev/null || break; sleep 1; done
+      if kill -0 "$pid" 2>/dev/null; then echo "live-store-copy: flowd $pid did not exit within 30s" >&2; exit 1; fi
+    fi
+  fi
+  pg dropdb -U flow --if-exists --force "$db"
+  rm -rf "$dir"
+  ;;
+*) usage ;;
+esac
diff --git a/scripts/test-live-store-copy.sh b/scripts/test-live-store-copy.sh
new file mode 100755
index 00000000..56094b66
--- /dev/null
+++ b/scripts/test-live-store-copy.sh
@@ -0,0 +1,75 @@
+#!/usr/bin/env bash
+# test-live-store-copy.sh — live-store-copy.sh copies only into
+# flow_<name>_live, starts and stops only its own flowd, and refuses the
+# protected ports and a malformed name before any side effect (kan-1041
+# self-review). docker, make, go and curl are stubs on PATH; no database or
+# daemon is touched. Exit codes: 0 pass, 1 fail.
+set -u
+cd "$(dirname "$0")/.." || exit 1
+rc=0
+fail() { printf 'FAIL: %s\n' "$1"; rc=1; }
+stub=$(mktemp -d)
+name="lsctest$$"
+port=$((20000 + $$ % 20000))
+log="$stub/log"
+trap 'rm -rf "$stub" "/tmp/flow-live-$name" "/tmp/flowd-$port.pid"' EXIT
+export STUB="$stub" LOG="$log"
+cat >"$stub/docker" <<'S'
+#!/usr/bin/env bash
+echo "docker $*" >>"$LOG"
+S
+cat >"$stub/make" <<'S'
+#!/usr/bin/env bash
+echo "make $*" >>"$LOG"
+S
+cat >"$stub/curl" <<'S'
+#!/usr/bin/env bash
+[ -e "$STUB/up" ]
+S
+cat >"$stub/go" <<'S'
+#!/usr/bin/env bash
+echo "go $*" >>"$LOG"
+out=$3
+printf '#!/usr/bin/env bash\nprintf "%%s\\n%%s\\n" $$ "$0" >"$TMPDIR/flowd-$FLOWD_PORT.pid"\necho "$FLOWD_DSN" >>"$LOG"\ntouch "$STUB/up"\nexec sleep 300\n' >"$out"
+chmod +x "$out"
+S
+chmod +x "$stub"/*
+export PATH="$stub:$PATH"
+lsc() { scripts/live-store-copy.sh "$@" >"$stub/out" 2>&1; }
+
+for bad in "" Flow ../a "a-b;rm"; do
+  : >"$log"
+  lsc up "$bad" "$port"; [ $? -eq 2 ] || fail "up '$bad': not exit 2"
+  [ -s "$log" ] && fail "up '$bad': ran $(cat "$log")"
+done
+for p in 4173 4174 80 x; do
+  : >"$log"
+  lsc up "$name" "$p"; [ $? -eq 2 ] || fail "up --port $p: not exit 2"
+  [ -s "$log" ] && fail "up port $p: ran $(cat "$log")"
+done
+
+: >"$log"
+lsc up "$name" "$port" || fail "up: exit $? ($(cat "$stub/out"))"
+grep -qx "docker exec -i flow-postgres createdb -U flow flow_${name}_live" "$log" || fail "up: no createdb of flow_${name}_live"
+grep -qx "docker exec -i flow-postgres pg_dump -U flow --no-owner --no-acl flow" "$log" || fail "up: no read-only pg_dump of flow"
+grep -q "psql -q -v ON_ERROR_STOP=1 -U flow -d flow_${name}_live$" "$log" || fail "up: no restore into flow_${name}_live"
+grep -qx "postgres://flow:flow@localhost:5433/flow_${name}_live?sslmode=disable" "$log" || fail "up: flowd not pointed at the copy"
+grep -qF "http://127.0.0.1:$port" "$stub/out" || fail "up: address not printed"
+pid=$(head -n1 "/tmp/flowd-$port.pid" 2>/dev/null)
+kill -0 "$pid" 2>/dev/null || fail "up: scratch flowd not running"
+lsc up "$name" "$port"; [ $? -eq 2 ] || fail "second up: not exit 2"
+
+sleep 300 &
+sentinel=$!
+: >"$log"
+lsc down "$name" || fail "down: exit $? ($(cat "$stub/out"))"
+kill -0 "$pid" 2>/dev/null && fail "down: scratch flowd still running"
+kill -0 "$sentinel" 2>/dev/null || fail "down: killed a process it did not start"
+kill "$sentinel"
+grep -qx "docker exec -i flow-postgres dropdb -U flow --if-exists --force flow_${name}_live" "$log" || fail "down: no dropdb of flow_${name}_live"
+grep -E '(dropdb|createdb).* flow$|-d flow$' "$log" && fail "down: touched the flow database"
+[ -e "/tmp/flow-live-$name" ] && fail "down: state dir left"
+lsc down "$name" || fail "second down: exit $?"
+
+[ "$rc" -eq 0 ] && printf 'all cases passed\n'
+exit "$rc"
````

  - [x] **Step 2: Verify.** `scripts/test-live-store-copy.sh` prints `all cases passed`; `bash -n scripts/live-store-copy.sh`, `scripts/check-references.sh` and `scripts/check-workspace-isolation.sh` exit 0. Never run `up` or `down` against the real store here.
  - [x] **Step 3: Commit.**

**Build:** green
**Files:** `scripts/live-store-copy.sh`, `scripts/test-live-store-copy.sh`, `.flow/project.md`
**Tests:** `test-live-store-copy.sh`
**Regression:** `test-live-store-copy.sh` — reverting the commit removes the script it drives, so it cannot run.
**Baseline:** before=3 after=4
<!-- measured: ls scripts/test-*.sh | grep -c . @ branch spectre/kan-1041-weekly-dashboard-model-effort-correlation -->
**Commit:** feat(scripts): script the live check's copy of the dev store
**After:** none
