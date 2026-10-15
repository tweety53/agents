# self-review-kan-951-cors-preflight-max-age

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-16-self-review-kan-951-cors-preflight-max-age-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`.
- **Task 1 commits in the gymie backend worktree `/Users/tweety53/Projects/gymie-worktrees/kan-951-cors-preflight-max-age`** (branch `spectre/kan-951-cors-preflight-max-age`), never in this agents worktree; Tasks 2–4 commit in this agents worktree. Never merge or push either.
- Each task's change is the unified diff its step gives, applied with `git apply` from the root of the repository its task names — every diff was applied in plan order and checked there. A diff that does not apply is reported, never hand-merged by guesswork.
- Every prose line a diff adds is copied verbatim: this change's `verbatim-moves.txt` acknowledges each new or reworded sentence by its exact text, and any other wording fails `scripts/check-verbatim-moves.sh`.
- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w .` before `gofmt -l .`. `go build ./...` needs the built SPA (`internal/web/dist`); the commands below vet and test only the packages each task touches.

## Review focus

- A gymie main-checkout run with no `FRONTEND_PORT` exported must still reach the frontend on its default port — the `${FRONTEND_PORT:-3000}` fallback.
- A full-suite fix commit the guard lists after it was already pushed — the paragraph amends only an unpushed tip.
- A fix run whose since-close sha is `-` must keep the merge-base diff, never skip the stage.
- `next-ref` over refs `F9` and `F10` must print `F11`, not `F10` — numeric, not lexical.
- A change whose record has dispatches but no pass row at all must render no pass-log section.

- [ ] 1. gymie's verify and capture rows set FRONTEND_URL beside BACKEND_URL
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/gymie-worktrees/kan-951-cors-preflight-max-age`, run `grep -n '^| .verify. \|^| .capture. ' .flow/project.md` and confirm both rows open with `BACKEND_URL=` and carry no `FRONTEND_URL=` — the defect: on a worktree stack Playwright's `baseURL` falls back to the main checkout's frontend port.
  - [ ] **Step 2: Apply the fix.** From `/Users/tweety53/Projects/gymie-worktrees/kan-951-cors-preflight-max-age`, apply this diff:

````diff verified:git apply --check in the gymie worktree at f71533bf; scripts/check-visual-verification.sh reports VISUAL-OK on the section with and without it; FRONTEND_PORT=3080 renders FRONTEND_URL=http://localhost:3080
diff --git a/.flow/project.md b/.flow/project.md
index 84df81b..2ab4cca 100644
--- a/.flow/project.md
+++ b/.flow/project.md
@@ -357,8 +357,8 @@ remaining entries are fixed, and the file deleted, under KAN-970.
 | Command | Runs |
 |---------|------|
 | `setup` | `npm install && npx playwright install chromium` |
-| `verify` | `BACKEND_URL=http://localhost:${GATEWAY_PORT:-8080} npx playwright test <specs>` |
-| `capture` | `BACKEND_URL=http://localhost:${GATEWAY_PORT:-8080} npx playwright test <spec> --update-snapshots` |
+| `verify` | `FRONTEND_URL=http://localhost:${FRONTEND_PORT:-3000} BACKEND_URL=http://localhost:${GATEWAY_PORT:-8080} npx playwright test <specs>` |
+| `capture` | `FRONTEND_URL=http://localhost:${FRONTEND_PORT:-3000} BACKEND_URL=http://localhost:${GATEWAY_PORT:-8080} npx playwright test <spec> --update-snapshots` |
 | `start` | `docker compose up -d && ./gradlew devStart -PisolateOAuthPorts -PfrontendRoot=<FRONTEND_ROOT> -PadminFrontendRoot=<ADMIN_FRONTEND_ROOT>` |
 | `fingerprint` | `./gradlew -q devStackFresh` |
 | `specs` | `node scripts/specs-for-diff.mjs <frontend-root> <merge-base>` |
````

  - [ ] **Step 3: Verify.** From this agents worktree run `scripts/check-visual-verification.sh /Users/tweety53/Projects/gymie-worktrees/kan-951-cors-preflight-max-age` (exit 0, `VISUAL-OK`), and `FRONTEND_PORT=3080 bash -c 'echo FRONTEND_URL=http://localhost:${FRONTEND_PORT:-3000}'` prints `FRONTEND_URL=http://localhost:3080`. Commit in `/Users/tweety53/Projects/gymie-worktrees/kan-951-cors-preflight-max-age`: `git -C /Users/tweety53/Projects/gymie-worktrees/kan-951-cors-preflight-max-age commit -m "fix(flow): verify and capture rows point Playwright at the worktree's frontend port" -- .flow/project.md`.

**Build:** green
**Files:** `.flow/project.md` (gymie repository, worktree `/Users/tweety53/Projects/gymie-worktrees/kan-951-cors-preflight-max-age`)
**Tests:** **none** — a configuration row; step 3's checks cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** fix(flow): verify and capture rows point Playwright at the worktree's frontend port
**After:** none

**Decision:** frontend-url-in-verify-rows

- [ ] 2. The full-suite fix carries its trailer, records its own sha and proves a determinism fix
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n 'Flow-Key: full-suite-fix\|check-untasked-commits' skills/flow/implement.md` and confirm no line of **The parent runs the full suite — never an implementer.** names either — the defect: kan-951's full-suite fix commit carried no `Flow-Key:` trailer and its row recorded Task 1's sha.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:applied with git apply to the worktree at edda7db8 in plan order; check-verbatim-moves.sh with this change's verbatim-moves.txt, check-references.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh, check-guard-symlinks.sh exit 0
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index 32381a81..3a4524dd 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -618,6 +618,10 @@ cause, as inline fixes are applied: it edits, commits on the route the branch's
 dictates (**Panel re-runs**, `skills/flow/review-panel-fix-round.md`), records the round as one
 `dispatches` pair `-role panel-fix -model <parent model> -effort <parent effort> -agent-id inline`
 under `full-suite-fix-<n>`, and re-runs the list until it passes.
+The commit carries `-m "Flow-Key: full-suite-fix-<n>"` (**Commits outside every task**, `skills/flow-contracts/git-boundaries.md`).
+Before it is pushed, `check-untasked-commits.sh <worktree> <merge-base>` runs; a commit it lists is still the unpushed tip, so `git commit --amend` adds the missing trailer and the guard runs again.
+The pair's `end` carries `-commit` with the sha `git -C <worktree> rev-parse HEAD` prints once that commit is made, never a task's commit.
+A fix for a failure that does not recur on every run is a determinism fix, held to the PROOF RUNS paragraph above: the failure is shown returning with the fix reverted before the fix is committed, and a fix whose failure never returns is not committed.
 When the change's diff touches a `## visual verification` `ui paths` glob
 (`git -C <worktree> diff --name-only <merge-base>..HEAD | check-visual-trigger.sh <worktree>` exits 0
 for any worktree in the resolved set, or the `--repo` run step 2 of **Visual verification**
````

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-guard-symlinks.sh`, `scripts/check-normative-inventory.sh` — all exit 0. Commit.

**Build:** green
**Files:** `skills/flow/implement.md`
**Tests:** **none** — four prose sentences; the guards in step 3 cover them, and the trailer's own guard `check-untasked-commits.sh` already exists.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): a full-suite fix carries its Flow-Key trailer, its own sha and its determinism proof
**After:** none

**Decision:** full-suite-fix-obligations

- [ ] 3. A fix run's visual trigger reads the fix run's own diff
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n 'since-close' skills/flow/verify-and-handoff.md` and confirm it prints nothing — the defect: Visual verification's trigger reads `<merge-base>..HEAD` on a fix run too, so kan-951's fix run touching two gateway test files re-ran the whole visual verify.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:applied with git apply to the worktree at edda7db8 in plan order; check-verbatim-moves.sh with this change's verbatim-moves.txt, check-references.sh, check-markdown-integrity.py, check-installed-citations.sh exit 0
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index 2a4157c6..08ddf621 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -203,6 +203,9 @@ this pipeline restates it. Resolve once per worktree in this run's resolved set,
    Its exit 0 makes the canonical worktree survive this step as the canonical worktree's own exit 0
    does.
 
+   **On a fix run, both calls read the fix run's own diff.** A worktree that has a since-close sha this run — the one **The late-fix reduction** (`skills/flow/review-panel-late-fix.md`) read for it — feeds them `git -C <worktree> diff --name-only <since-close-sha>..HEAD` in place of `<merge-base>..HEAD`, so a fix run whose delta touches no `ui paths` skips this stage at step 2 as exit 1 does.
+   A worktree whose since-close sha is `-` keeps `<merge-base>..HEAD`.
+
 **A `skipped` decision ends the stage at step 2.** When this run's decision
 (`<abs-worktree>/.superpowers/sdd/decision.json`) records `visual.verify` `skipped` (step 6 of
 **Decide**, `skills/flow/brainstorm-planner.md`), every worktree surviving step 2 prints
````

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`, `scripts/check-stage-mark-calls.sh` — all exit 0. Commit.

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none** — one prose paragraph; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): a fix run's visual trigger reads the diff since the panel's last clean close
**After:** none

**Decision:** fix-run-visual-trigger-since-close

- [ ] 4. Finding refs continue across runs and a panel round with no pass row still renders
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `cd stats && go run ./cmd/flow record next-ref -change x 2>&1 | head -3` and confirm it is refused as an unknown subcommand, and `grep -n 'Dispatches' internal/records/render.go | head` shows `passLogRounds` reads no dispatch row — the defect: kan-951's third run recorded `F1` again and rewrote the creating run's `F1`, and round 1's `panel-1-primary+principles` left no trace in the pass log.
  - [ ] **Step 2: Write the failing tests.** Apply only the test hunks of the diff in step 3 (`git apply --include='stats/cmd/flow/record_test.go' --include='stats/internal/records/render_test.go' <the diff>`) and run `cd stats && go test ./cmd/flow/ -run TestRunRecordNextRef -count=1` and `go test ./internal/records/ -run TestRenderPanelPassLogShowsADispatchedRoundWithNoPassRow -count=1`: both FAIL.
  - [ ] **Step 3: Implement.** Apply the rest of this diff:

````diff verified:applied with git apply to the worktree at edda7db8 in plan order; gofmt -l and go vet clean; go test ./internal/guard/ ./cmd/flow/ ./internal/records/ pass; both new tests fail with record.go and render.go reverted; check-verbatim-moves.sh with this change's verbatim-moves.txt exits 0
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index 98eb94a6..b75a6d66 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -277,6 +277,8 @@ const recordUsage = `usage: flow record dispatch begin [-addr url] [-timeout dur
                              -change name
        flow record next-round [-addr url] [-timeout dur] [-C dir]
                              -change name
+       flow record next-ref [-addr url] [-timeout dur] [-C dir]
+                             -change name
        flow record verdict  [-addr url] [-timeout dur] [-C dir]
                              -change name -guard guard -worktree path -verdict line
        flow record verdict false-positive [-addr url] [-timeout dur] [-C dir]
@@ -338,6 +340,13 @@ than the highest -round any pass, finding or mutation row of it carries, or
 memory, so a later run never reuses a round an earlier run recorded under.
 A failed read exits 1 and prints no number.
 
+next-ref prints the finding reference a run records its first finding
+under: F<n>, n one more than the highest F<n> any finding row of the change
+carries, or F1 when it has none. A ref is unique per change and a finding
+recorded under a ref already held rewrites that row, so a later run that
+numbered from F1 again would overwrite an earlier run's findings. A failed
+read exits 1 and prints nothing.
+
 journal-count prints how many of those journalled writes are still pending
 for a change -- one decimal count on stdout, and "unknown" where no count
 could be produced. It never blocks either, and for a sharper reason: it
@@ -568,6 +577,8 @@ func runRecord(ctx context.Context, args []string, stdin io.Reader, stdout, stde
 		return runRecordDispatches(ctx, args[1:], stdout, stderr)
 	case "next-round":
 		return runRecordNextRound(ctx, args[1:], stdout, stderr)
+	case "next-ref":
+		return runRecordNextRef(ctx, args[1:], stdout, stderr)
 	case "verdict":
 		return runRecordVerdict(ctx, args[1:], stdout, stderr)
 	case "verdicts":
@@ -2046,41 +2057,73 @@ func runRecordDispatches(ctx context.Context, args []string, stdout, stderr io.W
 // 0 for a change with none -- the round a panel run opens with, read from the
 // store so a later run cannot reuse an earlier run's round.
 func runRecordNextRound(ctx context.Context, args []string, stdout, stderr io.Writer) int {
-	fset := flag.NewFlagSet("flow record next-round", flag.ContinueOnError)
+	run, code, ok := readNextRecord(ctx, "next-round", args, stderr)
+	if !ok {
+		return code
+	}
+
+	next := 0
+	for _, r := range run.Passes {
+		next = max(next, r.Round+1)
+	}
+	for _, r := range run.Findings {
+		next = max(next, r.Round+1)
+	}
+	for _, r := range run.Mutations {
+		next = max(next, r.Round+1)
+	}
+	fmt.Fprintln(stdout, next)
+	return 0
+}
+
+// runRecordNextRef implements `flow record next-ref`: F<n>, n one more than
+// the highest F<n> any finding row of the change carries, or F1 for a change
+// with none -- the ref a run records its first finding under, read from the
+// store because the store upserts on (change, ref): kan-951's third run
+// numbered from F1 again and its F1 rewrote the creating run's F1.
+func runRecordNextRef(ctx context.Context, args []string, stdout, stderr io.Writer) int {
+	run, code, ok := readNextRecord(ctx, "next-ref", args, stderr)
+	if !ok {
+		return code
+	}
+
+	next := 1
+	for _, f := range run.Findings {
+		if n, err := strconv.Atoi(strings.TrimPrefix(f.Ref, "F")); err == nil {
+			next = max(next, n+1)
+		}
+	}
+	fmt.Fprintf(stdout, "F%d\n", next)
+	return 0
+}
+
+// readNextRecord reads the change's run record for next-round and next-ref:
+// a change the store has never heard of is an empty record, any other
+// failed read exits 1 with nothing on stdout.
+func readNextRecord(ctx context.Context, name string, args []string, stderr io.Writer) (records.Run, int, bool) {
+	fset := flag.NewFlagSet("flow record "+name, flag.ContinueOnError)
 	fset.SetOutput(stderr)
 	var f recordIdentityFlags
 	registerRecordIdentityFlags(fset, &f)
 
 	if ok, code := parseRecordFlags(fset, &f, args, stderr); !ok {
-		return code
+		return records.Run{}, code, false
 	}
 
 	projectKey, _, err := fallback.ProjectKey(f.dir)
 	if err != nil {
 		fmt.Fprintf(stderr, "flow: resolve project key: %v\n", err)
-		return 1
+		return records.Run{}, 1, false
 	}
 
 	run, callErr := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.Run, error) {
 		return cl.GetRunRecord(ctx, projectKey, f.change)
 	})
 	if callErr != nil && !errors.Is(callErr, client.ErrNotFound) {
-		fmt.Fprintf(stderr, "flow: next-round: %v\n", callErr)
-		return 1
-	}
-
-	next := 0
-	for _, r := range run.Passes {
-		next = max(next, r.Round+1)
-	}
-	for _, r := range run.Findings {
-		next = max(next, r.Round+1)
+		fmt.Fprintf(stderr, "flow: %s: %v\n", name, callErr)
+		return records.Run{}, 1, false
 	}
-	for _, r := range run.Mutations {
-		next = max(next, r.Round+1)
-	}
-	fmt.Fprintln(stdout, next)
-	return 0
+	return run, 0, true
 }
 
 // runRecordVerdict implements `flow record verdict`: one guard's verdict on
diff --git a/stats/cmd/flow/record_test.go b/stats/cmd/flow/record_test.go
index e5e16639..f2b4c3b5 100644
--- a/stats/cmd/flow/record_test.go
+++ b/stats/cmd/flow/record_test.go
@@ -4657,6 +4657,52 @@ func TestRunRecordNextRound(t *testing.T) {
 	})
 }
 
+// TestRunRecordNextRef pins the ref a run records its first finding under:
+// one past the highest F<n> any finding row carries, compared as numbers,
+// F1 for a change the store has never heard of, and nothing on a failed
+// read -- kan-951's third run numbered from F1 and overwrote the creating
+// run's F1.
+func TestRunRecordNextRef(t *testing.T) {
+	repo := recordRepo(t, "demo")
+	isolatedStateRoot(t)
+
+	for _, tc := range []struct {
+		name, body string
+		status     int
+		want       string
+	}{
+		{"highest ref plus one", `{"change":"demo","dispatches":[],
+		  "findings":[{"ref":"F9","round":0,"slot":"primary","severity":"Minor","note":"n","status":"fixed"},
+		              {"ref":"F10","round":3,"slot":"primary","severity":"Minor","note":"n","status":"open"}]}`, http.StatusOK, "F11"},
+		{"no rows", ``, http.StatusNotFound, "F1"},
+	} {
+		t.Run(tc.name, func(t *testing.T) {
+			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
+				w.WriteHeader(tc.status)
+				_, _ = w.Write([]byte(tc.body))
+			}))
+			defer srv.Close()
+			var stdout, stderr bytes.Buffer
+			code := run(context.Background(),
+				[]string{"record", "next-ref", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "demo"},
+				strings.NewReader(""), &stdout, &stderr)
+			if code != 0 || strings.TrimSpace(stdout.String()) != tc.want {
+				t.Fatalf("exit %d stdout %q, want 0 and %s; stderr:\n%s", code, stdout.String(), tc.want, stderr.String())
+			}
+		})
+	}
+
+	t.Run("unreachable store prints nothing", func(t *testing.T) {
+		var stdout, stderr bytes.Buffer
+		code := run(context.Background(),
+			[]string{"record", "next-ref", "-addr", deadPortAddr(t), "-timeout", "500ms", "-C", repo, "-change", "demo"},
+			strings.NewReader(""), &stdout, &stderr)
+		if code == 0 || stdout.Len() != 0 {
+			t.Fatalf("exit %d stdout %q, want non-zero and empty", code, stdout.String())
+		}
+	})
+}
+
 // TestRecordFindingRejectsMalformedRef pins that a -ref, -supersedes or
 // -regression-of not shaped F<positive int> is refused with exit 2 before
 // the store is contacted -- kan-964 stored two rows keyed by commit shas,
diff --git a/stats/internal/records/render.go b/stats/internal/records/render.go
index ce2526a7..e92d04e1 100644
--- a/stats/internal/records/render.go
+++ b/stats/internal/records/render.go
@@ -10,6 +10,7 @@ import (
 	"regexp"
 	"slices"
 	"sort"
+	"strconv"
 	"strings"
 	"time"
 )
@@ -197,15 +198,34 @@ func RenderPanel(r Run) string {
 	// record whose change carried no pass rows -- every panel rendered
 	// before this section existed, and every panel that ran clean without
 	// one -- renders nothing here, byte-identical to the old output.
-	if len(r.Passes) > 0 || len(r.Mutations) > 0 {
+	//
+	// A round a panel dispatch ran in renders even when the parent recorded
+	// no pass row for it, naming the dispatch keys: kan-951's round 1 ran
+	// panel-1-primary+principles and recorded nothing, and its record showed
+	// Round 0 alone, as if the round had never run.
+	rounds := passLogRounds(r)
+	if len(rounds) > 0 {
 		b.WriteString("\n## Pass log\n")
-		for _, round := range passLogRounds(r) {
+		for _, round := range rounds {
 			fmt.Fprintf(&b, "\n### Round %d\n\n", round)
+			passes := 0
 			for _, p := range r.Passes {
 				if p.Round == round {
+					passes++
 					fmt.Fprintf(&b, "- %s\n", neutraliseMarkers(p.Note))
 				}
 			}
+			if passes == 0 {
+				var keys []string
+				for _, d := range r.Dispatches {
+					if n, ok := panelDispatchRound(d); ok && n == round {
+						keys = append(keys, neutraliseMarkers(d.Key))
+					}
+				}
+				if len(keys) > 0 {
+					fmt.Fprintf(&b, "- no pass row recorded — dispatched: %s\n", strings.Join(keys, ", "))
+				}
+			}
 			mutations := 0
 			for _, m := range r.Mutations {
 				if m.Round != round {
@@ -226,26 +246,59 @@ func RenderPanel(r Run) string {
 
 // passLogRounds lists the distinct rounds a change's pass-log rows carry,
 // ascending -- the rows themselves decide, so a round holding only passes,
-// or only mutations, still renders.
+// or only mutations, still renders -- plus every round a panel dispatch ran
+// in, once the change has any pass-log row, so a round whose pass rows were
+// never recorded still shows. A change with no pass-log row at all renders
+// no section, as before.
 func passLogRounds(r Run) []int {
+	if len(r.Passes) == 0 && len(r.Mutations) == 0 {
+		return nil
+	}
 	seen := map[int]bool{}
 	var nums []int
-	for _, p := range r.Passes {
-		if !seen[p.Round] {
-			seen[p.Round] = true
-			nums = append(nums, p.Round)
+	add := func(n int) {
+		if !seen[n] {
+			seen[n] = true
+			nums = append(nums, n)
 		}
 	}
+	for _, p := range r.Passes {
+		add(p.Round)
+	}
 	for _, m := range r.Mutations {
-		if !seen[m.Round] {
-			seen[m.Round] = true
-			nums = append(nums, m.Round)
+		add(m.Round)
+	}
+	for _, d := range r.Dispatches {
+		if n, ok := panelDispatchRound(d); ok {
+			add(n)
 		}
 	}
 	sort.Ints(nums)
 	return nums
 }
 
+// panelDispatchRound reads the round out of a panel bundle's dispatch key,
+// panel-<round>-<slot|slot+slot> (skills/flow/implement.md's dispatch
+// table); any other row -- a panel-fix-<round> fixer included -- is not one.
+func panelDispatchRound(d Dispatch) (int, bool) {
+	if d.Role != "reviewer" {
+		return 0, false
+	}
+	rest, ok := strings.CutPrefix(d.Key, "panel-")
+	if !ok {
+		return 0, false
+	}
+	digits, _, ok := strings.Cut(rest, "-")
+	if !ok {
+		return 0, false
+	}
+	n, err := strconv.Atoi(digits)
+	if err != nil || n < 0 || strconv.Itoa(n) != digits {
+		return 0, false
+	}
+	return n, true
+}
+
 // dispatchMetrics is the part of a dispatch's metrics bag the ledger
 // reads: the token figures internal/harvest's second attribution pass
 // merges in, under the bag's top-level "tokens" key, plus the
diff --git a/stats/internal/records/render_test.go b/stats/internal/records/render_test.go
index b8e781ee..75763167 100644
--- a/stats/internal/records/render_test.go
+++ b/stats/internal/records/render_test.go
@@ -844,6 +844,36 @@ func TestRenderPanelPassLog(t *testing.T) {
 	}
 }
 
+// TestRenderPanelPassLogShowsADispatchedRoundWithNoPassRow pins that a
+// round a panel bundle ran in renders even when no pass row was recorded for
+// it, naming its dispatch keys -- kan-951's round 1 ran
+// panel-1-primary+principles, recorded nothing, and its record showed Round 0
+// alone. A panel-fix row and a non-panel reviewer add no round.
+func TestRenderPanelPassLogShowsADispatchedRoundWithNoPassRow(t *testing.T) {
+	out := records.RenderPanel(records.Run{
+		Change: "demo",
+		Dispatches: []records.Dispatch{
+			{Key: "panel-0-primary+principles", Role: "reviewer"},
+			{Key: "panel-1-primary+principles", Role: "reviewer"},
+			{Key: "panel-fix-2", Role: "panel-fix"},
+			{Key: "task-1-reviewer", Role: "reviewer"},
+		},
+		Passes: []records.Pass{{ID: 1, Round: 0, Note: "roster: compact — 39"}},
+	})
+
+	section := out[strings.Index(out, "## Pass log"):]
+	want := "\n### Round 1\n\n- no pass row recorded — dispatched: panel-1-primary+principles\n"
+	if !strings.Contains(section, want) {
+		t.Errorf("pass-log section missing %q:\n%s", want, section)
+	}
+	if strings.Contains(section, "### Round 2") {
+		t.Errorf("a panel-fix row rendered a round of its own:\n%s", section)
+	}
+	if strings.Contains(section, "dispatched: panel-0") {
+		t.Errorf("round 0 has a pass row and must not render the no-pass-row line:\n%s", section)
+	}
+}
+
 // TestRenderPanelWithoutPassLogIsUnchanged pins that a record whose change
 // carries no pass-log rows renders exactly as it did before the section
 // existed -- every panel rendered before this change, and every run whose
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index f9645b96..7f52bdff 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -719,6 +719,8 @@ flow record finding -change <name> -ref F<n> -round <r> -slot <slot> -severity <
 
 `-round` is `0` for the initial panel and `1..n` for a fix round. `-task` is the task id the
-finding names, omitted where it names none. `-ref` is unique within the change;
-the store's own constraint enforces it. **`-note` carries the reviewer's own sentence naming the
+finding names, omitted where it names none. `-ref` is unique within the change, and a finding
+recorded under a ref the change already holds rewrites that row. **A run's first finding ref is the
+one `flow record next-ref -change <name>` prints**, never `F1` carried from session memory or an
+earlier run's numbering, and each later finding of the run is one more. **`-note` carries the reviewer's own sentence naming the
 defect, not a dispatcher restatement.** Where the reviewer's wording runs long, quote the sentence
 that names the defect and leave the rest to the report file. **A fix round updates the finding it
````

  - [ ] **Step 4: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./cmd/flow/ ./internal/records/ && go test ./cmd/flow/ -run 'TestRunRecordNextRef|TestRunRecordNextRound' -count=1 && go test ./internal/records/ -count=1` — `gofmt -l` prints nothing, the tests pass; then `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py` exit 0. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/internal/records/render.go`, `stats/internal/records/render_test.go`, `skills/flow/review-panel.md`
**Tests:** `TestRunRecordNextRef`, `TestRenderPanelPassLogShowsADispatchedRoundWithNoPassRow`
**Regression:** reverting the commit makes `flow record next-ref` an unknown subcommand, so a later run numbers from `F1` and rewrites an earlier run's finding, and drops a dispatched round with no pass row from the rendered pass log.
**Baseline:** before=136 after=138
<!-- measured: cat stats/cmd/flow/record_test.go stats/internal/records/render_test.go | grep -c '^func Test' @ edda7db8 (before; after with this task's diff applied in the worktree) -->
**Commit:** fix(flow-cli): finding refs continue across runs and a panel round with no pass row still renders
**After:** none

**Decision:** next-ref-from-store
**Decision:** render-dispatched-round
