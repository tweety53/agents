# self-review-kan-1022-templates-step-2-be-own-templates-crud-from

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-25-self-review-kan-1022-templates-step-2-be-own-templates-crud-from-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is committed with the plan. Never merge or push.
- Each task's change is the unified diff its Step 2 gives (Step 1 for Tasks 2, 5 and 6), applied with `git apply` from this worktree's root, in task order — every diff was applied and checked there (`design.md`, **Context**). Save a diff by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork.
- Run `cd stats && gofmt -w .` before committing any Go file.
- Fixer groups (`design.md`, decision `first-review-effort`): group 1 is Tasks 1–3, group 2 is Tasks 4–6.

## Review focus

- A report with no `- wire contract:` line, or one reading `n/a`, must still refuse n/a on steps 8 and 9 and an n/a on step 10 not naming mockups (Task 1).
- A rerun at the HEAD the previous verifier row verified, and an unscoped run after a HEAD change, must still pass the finish gate (Task 3).
- A passing rerun that is not the report next after the failed one must clear nothing (Task 4).

- [ ] 1. Admit n/a on steps 8–10 beside a wire contract
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Apply Step 2's diff to the test file alone (`git apply --include='stats/internal/guard/check_verify_report_test.go' <the saved diff>`) and run `cd stats && go test ./internal/guard/ -run 'TestCheckVerifyReport$' -count=1`: `steps 8-10 n/a beside a wire contract` fails `rc=1` — the defect. Keep the applied hunk; Step 2 applies the rest with `git apply --exclude='stats/internal/guard/check_verify_report_test.go' <the saved diff>`.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at 86c20900; with it applied go test ./internal/guard/ -run TestCheckVerifyReport passes, and with verifyreport.go reverted the wire-contract case fails rc=1
diff --git a/scripts/check-verify-report.sh b/scripts/check-verify-report.sh
index 4040a292..3bf606b1 100755
--- a/scripts/check-verify-report.sh
+++ b/scripts/check-verify-report.sh
@@ -17,7 +17,10 @@
 # step, ` | `-separated, for steps 4, 7, 8, 9, 10, 11 and `motion`
 # (skills/flow/visual-verify-verifier.md). A status is admissible when it is
 #   done
-#   n/a — <reason>       steps 4 and 10 only, and `motion` when <motions named> is 0
+#   n/a — <reason>       steps 4 and 10 only, and `motion` when <motions named> is 0;
+#                        steps 8, 9 and 10 too when a `- wire contract:` line
+#                        names an endpoint — the dispatch named no view, so
+#                        nothing is captured (kan-1022's self-review)
 #   blocked — <reason>   the reason citing a non-zero `exit <n>`; step 10 only
 #                        when step 8 is blocked too or the reason cites
 #                        `compose-mockup-frames exit <n>` (KAN-1000)
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index 5fe2ac51..54a6d135 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -600,7 +600,9 @@ report's `wire contract:` line reads n/a. Otherwise, signed in as an account
 `<project>/.flow/project.md`'s `## credentials` declares, call each named endpoint on the running
 API at its worktree-resolved URL, read the response's shape and wire names against the line, then
 drive the stated acceptance behaviour through the API and read the next response — a screenshot
-comparison is no substitute for it.
+comparison is no substitute for it. With a `wire contract:` line the dispatch names no view, so
+steps 8, 9 and 10 read n/a — wire contract, no view drawn — never `done` over a capture that
+never ran (kan-1022's self-review).
 
 **No required step is left undone for want of time or context.** The report's `steps:` line gives
 each of steps 4, 7–11 and the motion step a status `check-verify-report.sh` admits — its header
diff --git a/stats/internal/guard/check_verify_report_test.go b/stats/internal/guard/check_verify_report_test.go
index 00090cf6..82b399bf 100644
--- a/stats/internal/guard/check_verify_report_test.go
+++ b/stats/internal/guard/check_verify_report_test.go
@@ -37,6 +37,9 @@ func TestCheckVerifyReport(t *testing.T) {
 		{"motions named, step n/a", allDone, "1", 1, "motion n/a — no motions named"},
 		{"motions named, no motion line", "- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion done\n", "2", 1, "motion: no `- motion: 2/2` line"},
 		{"motions named, one entry short", "- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion done\n- motion: 1/2 — swipe: clean\n", "2", 1, "motion: no `- motion: 2/2` line"},
+		// kan-1022: a wire-contract dispatch names no view, so 8-10 are n/a, never `done`.
+		{"steps 8-10 n/a beside a wire contract", "- wire contract: GET /api/templates/{id} — shape matched\n- steps: 4 done | 7 done | 8 n/a — wire contract, no view drawn | 9 n/a — wire contract, no view drawn | 10 n/a — wire contract, no view drawn | 11 done | motion n/a — none\n", "0", 0, "COMPLETE"},
+		{"steps 8-9 n/a with no wire contract", "- wire contract: n/a — no wire contract named\n- steps: 4 done | 7 done | 8 n/a — no view drawn | 9 n/a — no view drawn | 10 done | 11 done | motion n/a — none\n", "0", 1, "8 n/a — no view drawn; 9 n/a — no view drawn"},
 		{"missing report", "", "0", 2, "cannot read"},
 	}
 	for _, tc := range cases {
diff --git a/stats/internal/guard/verifyreport.go b/stats/internal/guard/verifyreport.go
index d9ebc7c1..d981ba7f 100644
--- a/stats/internal/guard/verifyreport.go
+++ b/stats/internal/guard/verifyreport.go
@@ -37,6 +37,10 @@ var (
 	// (kan-1015: `specs` printed 2 under the round base, the verifier ran 95).
 	vrVerifyLine  = regexp.MustCompile(`^\s*- verify:\s*(.*)$`)
 	vrMappedSpecs = regexp.MustCompile(`\b(\d+) mapped specs\b`)
+	// A `- wire contract:` line naming an endpoint: the dispatch named no
+	// view, so steps 8-10 have nothing to capture (kan-1022's self-review: a
+	// verifier wrote them `done` over a capture it never ran).
+	vrWireContract = regexp.MustCompile(`^\s*- wire contract:\s*(.*)$`)
 )
 
 func checkVerifyReport(args []string, env Env, stdout, stderr io.Writer) int {
@@ -66,6 +70,7 @@ func checkVerifyReport(args []string, env Env, stdout, stderr io.Writer) int {
 	status := map[string]string{}
 	motionLine, bareFingerprint := false, false
 	inReport, unevidenced, verify := false, "", ""
+	noView := false
 	for _, line := range strings.Split(string(data), "\n") {
 		if m := vrVerifyLine.FindStringSubmatch(line); m != nil && verify == "" {
 			verify = strings.TrimSpace(m[1])
@@ -90,6 +95,9 @@ func checkVerifyReport(args []string, env Env, stdout, stderr io.Writer) int {
 			motionLine = true
 		}
 		bareFingerprint = bareFingerprint || vrBareFingerprint.MatchString(line)
+		if m := vrWireContract.FindStringSubmatch(line); m != nil && !strings.HasPrefix(strings.TrimSpace(m[1]), "n/a") {
+			noView = true
+		}
 	}
 	var undone []string
 	if bareFingerprint {
@@ -107,7 +115,7 @@ func checkVerifyReport(args []string, env Env, stdout, stderr io.Writer) int {
 	}
 	for _, id := range vrSteps {
 		s, ok := status[id]
-		naAllowed := id == "4" || id == "10" || (id == "motion" && motions == 0)
+		naAllowed := id == "4" || id == "10" || (id == "motion" && motions == 0) || (noView && (id == "8" || id == "9"))
 		reason, isNA := strings.CutPrefix(s, "n/a — ")
 		cause, isBlocked := strings.CutPrefix(s, "blocked — ")
 		switch {
@@ -116,8 +124,9 @@ func checkVerifyReport(args []string, env Env, stdout, stderr io.Writer) int {
 		case s == "done" && id == "motion" && motions > 0 && !motionLine:
 			undone = append(undone, fmt.Sprintf("motion: no `- motion: %d/%d` line", motions, motions))
 		case s == "done":
-		// Step 10 is n/a only when no mockups are declared, so its reason must say so.
-		case isNA && naAllowed && strings.TrimSpace(reason) != "" && (id != "10" || strings.Contains(reason, "mockups")):
+		// Step 10 is n/a only when no mockups are declared, so its reason must say so,
+		// or when a wire contract names no view to compose.
+		case isNA && naAllowed && strings.TrimSpace(reason) != "" && (id != "10" || noView || strings.Contains(reason, "mockups")):
 		// Step 10 measures the captures step 8 wrote, so it blocks only when step 8 did or
 		// on compose-mockup-frames' own exit (KAN-1000: two verifiers skipped step 10 with
 		// every capture on disk, and departures surfaced one round at a time).
````

  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/guard/ -run 'TestCheckVerifyReport' -count=1` passes; `cd stats && gofmt -l .` prints nothing and `cd stats && go vet ./internal/guard/` exits 0; `scripts/check-verbatim-moves.sh`, `scripts/check-installed-citations.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.
  - [ ] **Step 4: Repair the kan-1022 record.** In `/Users/tweety53/Projects/gymie-worktrees/kan-1022-templates-step-2-be-own-templates-crud-from/.superpowers/sdd/verify-report-visual-verify-2.md` (gitignored; nothing is committed in gymie) replace the `- steps:` line with `- steps: 4 done | 7 done | 8 n/a — wire contract, no view drawn | 9 n/a — wire contract, no view drawn | 10 n/a — wire contract, no view drawn | 11 done | motion n/a — no motions named`, then run `scripts/check-verify-report.sh /Users/tweety53/Projects/gymie-worktrees/kan-1022-templates-step-2-be-own-templates-crud-from/.superpowers/sdd/verify-report-visual-verify-2.md 0` from this worktree: it prints `VERIFY-REPORT-COMPLETE`. A file no longer on disk is reported, never recreated.

**Build:** green
**Files:** `stats/internal/guard/verifyreport.go`, `stats/internal/guard/check_verify_report_test.go`, `scripts/check-verify-report.sh`, `skills/flow/visual-verify-verifier.md`
**Tests:** `steps 8-10 n/a beside a wire contract`, `steps 8-9 n/a with no wire contract`
**Regression:** reverting `verifyreport.go` refuses n/a on steps 8 and 9 again, so `steps 8-10 n/a beside a wire contract` fails `rc=1`.
**Baseline:** before=30 after=32
<!-- measured: grep -cE '^[[:space:]]+\{"' stats/internal/guard/check_verify_report_test.go @ 86c20900 (before; after with this task's diff applied) -->
**Commit:** fix(guard): admit n/a on steps 8 to 10 beside a wire contract
**After:** none

**Decision:** wire-contract-na
**Decision:** repair-kan-1022-report

- [ ] 2. Give a flowBase change a worktree in every apps repository
  - [ ] **Step 1: Apply the fix.** Apply this diff:

````diff verified:git apply at 86c20900 after Task 1; check-verbatim-moves.sh then fails only on sentences this change's verbatim-moves.txt lists
diff --git a/skills/flow/visual-verify.md b/skills/flow/visual-verify.md
index 3dd010d8..4f195550 100644
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -207,7 +207,12 @@ verifier reports on a `wire contract` line blocks as a defect in a captured scre
    missing checkout, ~4.45M tokens spent for nothing. The recipe's merge-order record runs with
    it, its planning commit included (**Isolate the workspace — additional worktrees**,
    `skills/flow/cross-repo-worktrees.md`): a repository the canonical `link.md`'s `## Merge order`
-   omits is one run 1 never sequences and the self-review bundle never reads (kan-998). With the
+   omits is one run 1 never sequences and the self-review bundle never reads (kan-998). **A change
+   whose worktree records a flowBase gets a worktree in every `## apps` repository the set holds
+   none in, the same way** — the `kickoff-worktree.sh --from` call and its merge-order record,
+   before the checks below — so the stack serves `<base>`'s build, never a main checkout on the
+   default branch: kan-1022's stack served the `develop` frontend to a change based on
+   feature/templates, and 28 specs failed on it. With the
    parent's own Bash call, before the
    dispatch below, run the four checks — ports held only by their own app, the base-URL
    configuration resolving to the workspace's values, origin allowed, one Playwright checkout per
````

  - [ ] **Step 2: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-installed-citations.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/visual-verify.md`
**Tests:** **none** — a rule sentence; no guard reads it.
**Commit:** fix(visual-verify): give a flowBase change a worktree in every apps repository
**After:** none

**Decision:** flowbase-apps-worktrees

- [ ] 3. Refuse a failed-spec rerun recorded across HEADs
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Apply Step 2's diff to the test file alone (`git apply --include='stats/internal/guard/check_visual_verify_dispatched_test.go' <the saved diff>`) and run `cd stats && go test ./internal/guard/ -run 'TestCheckVisualVerifyDispatchedRefusesARerunAcrossHeads' -count=1`: `a rerun across a HEAD change` gets exit 0 and the OK verdict — the defect. Keep the applied hunks; Step 2 applies the rest with `git apply --exclude='stats/internal/guard/check_visual_verify_dispatched_test.go' <the saved diff>`.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at 86c20900 after Task 2; with it applied go test ./internal/guard/ -run TestCheckVisualVerifyDispatched passes, and with visualverifydispatched.go reverted the rerun-across-HEADs case gets exit 0
diff --git a/scripts/check-visual-verify-dispatched.sh b/scripts/check-visual-verify-dispatched.sh
index 8a849ad1..e97b45ec 100755
--- a/scripts/check-visual-verify-dispatched.sh
+++ b/scripts/check-visual-verify-dispatched.sh
@@ -40,6 +40,17 @@
 #   base <sha> — the loop's closing unscoped run is not recorded
 #   (skills/flow/verify-fix-loop.md, **One full run closes the loop**) — HEAD <sha>
 #
+# Nor may its report, `<worktree>/.superpowers/sdd/verify-report-<key>.md`,
+# carry a `- verify:` item naming a `rerun of` failed specs when the
+# `visual-verify…` verifier row recorded just before it, of any outcome,
+# carries a different `-diff-base` — a rerun stands only at the HEAD that
+# row verified (kan-1022 self-review); the verdict is
+#   VISUAL-VERIFY-MISSING: the newest verifier row '<key>' re-ran only failed
+#   specs at <diff-base> where the verifier row before it verified <diff-base>
+#   — the unscoped run at this HEAD is not recorded
+#   (skills/flow/visual-verify.md) — HEAD <sha>
+# A report no longer on disk reads as no rerun.
+#
 # Exit 0 on a VISUAL-VERIFY-OK verdict, exit 1 on VISUAL-VERIFY-MISSING —
 # UI paths were touched and no qualifying verifier dispatch is recorded;
 # the verdict line carries the answer — and exit 2 when it cannot answer at
diff --git a/skills/flow/visual-verify.md b/skills/flow/visual-verify.md
index 4f195550..9f47963b 100644
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -124,7 +124,12 @@ verified and that report's `verify:` exit was non-zero, the prompt adds `rerun s
 spec paths that report's `verify` output names as failing; the verifier's step 7 runs those alone
 (`skills/flow/visual-verify-verifier.md`), and their exit is the `verify` exit — a non-zero one
 still blocks, exactly as above: kan-983's `visual-verify-2` re-ran all 87 specs, ~14 minutes, to
-clear one the previous run had already passed 30 times out of 30. The ledger render and this stage's `end` mark follow
+clear one the previous run had already passed 30 times out of 30. **An environment cure that
+changes what the stack serves — a served root switched, a worktree added to the set — voids every
+earlier pass**, so the next re-dispatch runs unscoped; `check-visual-verify-dispatched.sh` reports
+a newest verifier whose report re-ran failed specs under a `-diff-base` other than the verifier row
+before it as MISSING (kan-1022: a rerun of 19 failed specs left 76 never run against the served
+feature/templates frontend). The ledger render and this stage's `end` mark follow
 whichever report was last.
 
 **A final report that is a block closes its dispatch `-outcome blocked -cause <cause>`, never
diff --git a/stats/internal/guard/check_visual_verify_dispatched_test.go b/stats/internal/guard/check_visual_verify_dispatched_test.go
index e82e6c7f..f59b65fb 100644
--- a/stats/internal/guard/check_visual_verify_dispatched_test.go
+++ b/stats/internal/guard/check_visual_verify_dispatched_test.go
@@ -188,6 +188,7 @@ type vvdCase struct {
 	noRecord   bool                      // no spectre/changes/demo/visual-verification.md in the worktree
 	// archivedRecord: the record at spectre/changes/archive/demo/ instead
 	archivedRecord bool
+	files          map[string]string // written into the case worktree, path relative to it
 }
 
 type vvdResult struct {
@@ -225,6 +226,9 @@ func vvdRun(t *testing.T, c vvdCase) vvdResult {
 	case !c.noRecord:
 		writeFile(t, wt+"/spectre/changes/demo/visual-verification.md", "# record\n")
 	}
+	for p, body := range c.files {
+		writeFile(t, wt+"/"+p, body)
+	}
 	// A row names the HEAD it verified as @HEAD@ or the merge base as
 	// @BASE@: the shas exist only once the master is built.
 	head := tcfGit(t, wt, "rev-parse", "HEAD")
@@ -759,3 +763,32 @@ func TestCheckVisualVerifyDispatchedScopesTheNewestVerifierPerWorktree(t *testin
 		t.Fatalf("got exit %d stdout %q stderr %q, want exit 0 and the OK verdict", r.code, r.out, r.err)
 	}
 }
+
+// TestCheckVisualVerifyDispatchedRefusesARerunAcrossHeads pins kan-1022's
+// self-review: a newest verifier whose report re-ran only the failed specs
+// at a -diff-base other than the verifier row before it -- a blocked one
+// included -- is MISSING; the same rerun at that row's own HEAD, and an
+// unscoped run after a HEAD change, pass.
+func TestCheckVisualVerifyDispatchedRefusesARerunAcrossHeads(t *testing.T) {
+	t.Parallel()
+	const rerun = "## Report\n- verify: exit 0 — rerun of 19 failed specs at unchanged HEAD\n"
+	rows := func(prev string) []byte {
+		return []byte(`[` +
+			`{"key":"visual-verify","role":"verifier","outcome":"blocked","diffBase":"` + prev + `","startedAt":"2026-10-24T10:00:00Z"},` +
+			`{"key":"visual-verify-2","role":"verifier","outcome":"completed","diffBase":"@HEAD@","startedAt":"2026-10-24T11:00:00Z"}]`)
+	}
+	for _, c := range []struct {
+		name, prev, report string
+		code               int
+		want               string
+	}{
+		{"a rerun across a HEAD change", "@BASE@", rerun, 1, "VISUAL-VERIFY-MISSING: the newest verifier row 'visual-verify-2' re-ran only failed specs at "},
+		{"a rerun at the same HEAD", "@HEAD@", rerun, 0, "VISUAL-VERIFY-OK: UI paths touched"},
+		{"an unscoped run across a HEAD change", "@BASE@", "## Report\n- verify: exit 0\n", 0, "VISUAL-VERIFY-OK: UI paths touched"},
+	} {
+		r := vvdRun(t, vvdCase{dispatches: rows(c.prev), files: map[string]string{".superpowers/sdd/verify-report-visual-verify-2.md": c.report}})
+		if r.code != c.code || !strings.HasPrefix(r.out, c.want) {
+			t.Fatalf("%s: got exit %d stdout %q stderr %q, want exit %d and %q", c.name, r.code, r.out, r.err, c.code, c.want)
+		}
+	}
+}
diff --git a/stats/internal/guard/visualverifydispatched.go b/stats/internal/guard/visualverifydispatched.go
index 865b6494..2f0c60a6 100644
--- a/stats/internal/guard/visualverifydispatched.go
+++ b/stats/internal/guard/visualverifydispatched.go
@@ -5,6 +5,7 @@ import (
 	"encoding/json"
 	"fmt"
 	"io"
+	"os"
 	"path/filepath"
 	"regexp"
 	"strings"
@@ -304,6 +305,28 @@ func checkVisualVerifyDispatched(args []string, env Env, stdout, stderr io.Write
 		fmt.Fprintln(stdout, verdict)
 		return 1
 	}
+	// A RERUN OF FAILED SPECS STANDS ONLY AT THE HEAD THE ROW BEFORE IT
+	// VERIFIED (kan-1022 self-review): a newest verifier whose report re-ran
+	// only the specs that failed, at a -diff-base other than the verifier row
+	// before it, left every spec that passed there unrun at this HEAD
+	// (skills/flow/visual-verify.md, **A re-dispatch at an unchanged HEAD
+	// re-runs only the specs that failed**). A report no longer on disk
+	// reads as no rerun.
+	if found {
+		prev, ok := vvdHeadBefore(rows, newestKey)
+		report, err := os.ReadFile(worktree + "/.superpowers/sdd/verify-report-" + newestKey + ".md")
+		if ok && prev != newest.head && err == nil && vvdRerun.Match(report) {
+			headSHA, ok := capture(envGit(env)("-C", worktree, "rev-parse", "HEAD"))
+			if !ok {
+				fmt.Fprintf(stderr, "%sgit cannot name HEAD in %s — cannot answer\n", vvdPrefix, worktree)
+				return 2
+			}
+			verdict := "VISUAL-VERIFY-MISSING: the newest verifier row '" + newestKey + "' re-ran only failed specs at " + newest.head + " where the verifier row before it verified " + prev + " — the unscoped run at this HEAD is not recorded (skills/flow/visual-verify.md) — HEAD " + headSHA
+			record(verdict)
+			fmt.Fprintln(stdout, verdict)
+			return 1
+		}
+	}
 	if found {
 		verdict := "VISUAL-VERIFY-OK: UI paths touched and a completed (or ended-but-unclosed) verifier dispatch is recorded for '" + name + "'"
 		record(verdict)
@@ -571,6 +594,35 @@ func vvdCompletedVerifiers(raw []byte) (rows []vvdVerifier, ok bool) {
 	return rows, true
 }
 
+// vvdRerun is a report's `- verify:` item naming a rerun of failed specs.
+var vvdRerun = regexp.MustCompile(`(?m)^\s*- verify:.*\brerun of\b`)
+
+// vvdHeadBefore is the diffBase of the `visual-verify…` verifier row, of any
+// outcome, recorded just before the last row keyed key; ok is false when no
+// such row precedes it or the rows do not parse.
+func vvdHeadBefore(raw []byte, key string) (head string, ok bool) {
+	vals, ok := pfdStream(raw)
+	if !ok || len(vals) != 1 {
+		return "", false
+	}
+	arr, _ := vals[0].([]any)
+	prev, found := "", false
+	head, ok = "", false
+	for _, e := range arr {
+		row, _ := e.(map[string]any)
+		k, _ := row["key"].(string)
+		if row["role"] != "verifier" || !strings.HasPrefix(k, "visual-verify") {
+			continue
+		}
+		base, _ := row["diffBase"].(string)
+		if k == key {
+			head, ok = prev, found
+		}
+		prev, found = base, true
+	}
+	return head, ok
+}
+
 // vvdVerifier is one qualifying verifier row: its key, the HEAD(s) it
 // verified (diffBase), its start instant and its scoped round base.
 type vvdVerifier struct{ key, head, start, roundBase string }
````

  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/guard/ -run 'TestCheckVisualVerifyDispatched' -count=1` passes; `cd stats && gofmt -l .` prints nothing and `cd stats && go vet ./internal/guard/` exits 0; `scripts/check-verbatim-moves.sh`, `scripts/check-installed-citations.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/visualverifydispatched.go`, `stats/internal/guard/check_visual_verify_dispatched_test.go`, `scripts/check-visual-verify-dispatched.sh`, `skills/flow/visual-verify.md`
**Tests:** `TestCheckVisualVerifyDispatchedRefusesARerunAcrossHeads`
**Regression:** reverting `visualverifydispatched.go` passes a rerun of failed specs recorded at another HEAD, so `a rerun across a HEAD change` gets exit 0.
**Baseline:** before=9 after=10
<!-- measured: grep -c '^func Test' stats/internal/guard/check_visual_verify_dispatched_test.go @ 86c20900 (before; after with this task's diff applied) -->
**Commit:** fix(guard): refuse a failed-spec rerun recorded across HEADs
**After:** Task 2

**Decision:** rerun-same-head-gate

- [ ] 4. Count a passing rerun of failed specs in signal nine
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Apply Step 2's diff to the test file alone (`git apply --include='stats/internal/guard/check_unfinished_work_test.go' <the saved diff>`) and run `cd stats && go test ./internal/guard/ -run 'TestCheckUnfinishedWorkVerifyFailureRerunPassed' -count=1`: `a later passing rerun addresses it` fails — the defect. Keep the applied hunk; Step 2 applies the rest with `git apply --exclude='stats/internal/guard/check_unfinished_work_test.go' <the saved diff>`.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at 86c20900 after Task 3; with it applied go test ./internal/guard/ -run TestCheckUnfinishedWork passes, and with unfinishedwork.go reverted the later-passing-rerun case fails
diff --git a/scripts/check-unfinished-work.sh b/scripts/check-unfinished-work.sh
index 305683cd..0c5817a2 100755
--- a/scripts/check-unfinished-work.sh
+++ b/scripts/check-unfinished-work.sh
@@ -133,7 +133,9 @@
 # commit, or when an earlier false-positive verdict of this change names
 # `<spec>:<line>` in its reason, no path character before it and no digit
 # after (kan-1010's self-review) — a store that cannot answer clears
-# nothing. Each other failure is one reason, `verify failure not addressed:
+# nothing — or when the report next after it by mtime opens its `- verify:`
+# item `exit 0 — rerun of`, a rerun of exactly those failed specs that passed
+# (kan-1022's self-review). Each other failure is one reason, `verify failure not addressed:
 # <spec>:<line>[ <name>.png] (<report>)`; an unreadable report is exit 2.
 #
 # SIGNAL TEN: A MULTI-REPOSITORY CHANGE WHOSE MERGE ORDER MISSES A REPOSITORY
diff --git a/stats/internal/guard/check_unfinished_work_test.go b/stats/internal/guard/check_unfinished_work_test.go
index 4a9243eb..8080ac7d 100644
--- a/stats/internal/guard/check_unfinished_work_test.go
+++ b/stats/internal/guard/check_unfinished_work_test.go
@@ -1083,6 +1083,49 @@ func TestCheckUnfinishedWorkVerifyFailureNotAddressed(t *testing.T) {
 	}
 }
 
+// TestCheckUnfinishedWorkVerifyFailureRerunPassed pins kan-1022's
+// self-review: a failed report whose next report, by mtime, passed a rerun
+// of the failed specs is addressed; a passing rerun before the failure, or a
+// failing next report, addresses nothing.
+func TestCheckUnfinishedWorkVerifyFailureRerunPassed(t *testing.T) {
+	t.Parallel()
+	const failed = "- verify: exit 1\n  1) fidelity/templates.spec.ts:80 opens the template — timeout\n"
+	const reason = "verify failure not addressed: fidelity/templates.spec.ts:80"
+	first, second := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(2026, 1, 1, 13, 0, 0, 0, time.UTC)
+	for _, c := range []struct {
+		name           string
+		failAt, nextAt time.Time
+		next           string
+		wantIn         bool
+	}{
+		{"a later passing rerun addresses it", first, second, "- verify: exit 0 — rerun of 19 failed specs at unchanged HEAD\n", false},
+		{"an earlier passing rerun does not", second, first, "- verify: exit 0 — rerun of 19 failed specs at unchanged HEAD\n", true},
+		{"a later failing rerun does not", first, second, "- verify: exit 1 — rerun of 19 failed specs at unchanged HEAD\n", true},
+	} {
+		t.Run(c.name, func(t *testing.T) {
+			t.Parallel()
+			f := newUW(t)
+			gitRun(t, "", "init", "-q", f.wt)
+			gitRun(t, f.wt, "commit", "-q", "--allow-empty", "-m", "base")
+			base := tcfGit(t, f.wt, "rev-parse", "HEAD")
+			for _, r := range []struct {
+				name, body string
+				at         time.Time
+			}{{"visual-verify", failed, c.failAt}, {"visual-verify-2", c.next, c.nextAt}} {
+				rp := f.wt + "/.superpowers/sdd/verify-report-" + r.name + ".md"
+				writeFile(t, rp, "## Report\n"+r.body)
+				if err := os.Chtimes(rp, r.at, r.at); err != nil {
+					t.Fatal(err)
+				}
+			}
+			r := f.run(t, f.wt, "demo", "", f.wt+"="+base)
+			if r.rc != 0 || strings.Contains(r.out, reason) != c.wantIn {
+				t.Fatalf("exit %d stdout %q stderr %q; want exit 0 and %q present=%v", r.rc, r.out, r.err, reason, c.wantIn)
+			}
+		})
+	}
+}
+
 // TestCheckUnfinishedWorkVerifyFailureFlaggedEarlier pins kan-1010's
 // self-review: screenshots.spec.ts:52 was flagged a false positive three
 // times, each flag landing on that run's newest row only. An earlier
diff --git a/stats/internal/guard/unfinishedwork.go b/stats/internal/guard/unfinishedwork.go
index 2000fa02..2495b431 100644
--- a/stats/internal/guard/unfinishedwork.go
+++ b/stats/internal/guard/unfinishedwork.go
@@ -529,7 +529,10 @@ func checkUnfinishedWork(args []string, env Env, stdout, stderr io.Writer) int {
 			narrative, _ := os.ReadFile(filepath.Dir(primary) + "/narrative.md")
 			var flagged []string
 			flaggedRead := false
-			for _, rp := range reports {
+			// The reports in mtime order: a report whose next one passed a
+			// rerun of its failed specs is addressed (kan-1022's self-review).
+			bodies, infos := make([]string, len(reports)), make([]os.FileInfo, len(reports))
+			for i, rp := range reports {
 				b, err := os.ReadFile(rp)
 				if err != nil {
 					return unreadable(rp)
@@ -538,8 +541,18 @@ func checkUnfinishedWork(args []string, env Env, stdout, stderr io.Writer) int {
 				if err != nil {
 					return unreadable(rp)
 				}
-				for _, fl := range uwVerifyFailures(string(b)) {
-					addressed := false
+				bodies[i], infos[i] = string(b), st
+			}
+			order := make([]int, len(reports))
+			for i := range order {
+				order[i] = i
+			}
+			slices.SortStableFunc(order, func(a, b int) int { return infos[a].ModTime().Compare(infos[b].ModTime()) })
+			for k, i := range order {
+				rp, st := reports[i], infos[i]
+				rerunPassed := k+1 < len(order) && uwRerunPassed.MatchString(bodies[order[k+1]])
+				for _, fl := range uwVerifyFailures(bodies[i]) {
+					addressed := rerunPassed
 					for _, c := range commits {
 						if c.at < st.ModTime().Unix() {
 							continue
@@ -1052,8 +1065,10 @@ var uwMergeOrderEntryRE = regexp.MustCompile("^\\s*\\d+\\.\\s+`([^`]+)`")
 type uwVerifyFailure struct{ spec, line, png string }
 
 var (
-	uwSpecRE = regexp.MustCompile(`([^\s:]+\.spec\.ts):(\d+)`)
-	uwPngRE  = regexp.MustCompile(`([A-Za-z0-9_.-]+)\.png`)
+	// A report's `- verify:` item passing a rerun of the failed specs.
+	uwRerunPassed = regexp.MustCompile(`(?m)^- verify: exit 0 — rerun of\b`)
+	uwSpecRE      = regexp.MustCompile(`([^\s:]+\.spec\.ts):(\d+)`)
+	uwPngRE       = regexp.MustCompile(`([A-Za-z0-9_.-]+)\.png`)
 )
 
 // uwVerifyFailures reads a verify report's failed `- verify: exit <n>` item
````

  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/guard/ -run 'TestCheckUnfinishedWork' -count=1` passes; `cd stats && gofmt -l .` prints nothing and `cd stats && go vet ./internal/guard/` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `scripts/check-unfinished-work.sh`
**Tests:** `TestCheckUnfinishedWorkVerifyFailureRerunPassed`
**Regression:** reverting `unfinishedwork.go` leaves a failure whose next report passed its rerun outstanding, so `a later passing rerun addresses it` fails.
**Baseline:** before=8 after=9
<!-- measured: grep -c '^func Test' stats/internal/guard/check_unfinished_work_test.go @ 86c20900 (before; after with this task's diff applied) -->
**Commit:** fix(guard): count a passing rerun of failed specs in signal nine
**After:** none

**Decision:** signal-nine-rerun

- [ ] 5. A reproducer reads the defect's behaviour, never a fix's text
  - [ ] **Step 1: Apply the fix.** Apply this diff:

````diff verified:git apply at 86c20900 after Task 4; go test ./internal/guard/ -run TestRenderSlotPrompt passes with it applied
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index e18b6512..b3b4cdea 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -561,7 +561,11 @@ declarations sit in the window.** The pre-fix leg reads the tree that leg runs i
 defect-logic reads cwd-relative to that leg's worktree, never an archived HEAD of the live
 worktree, which is what lets `prove-reproducer.sh` hold both legs. The `# demonstrates:` and
 `# premise:` declarations sit inside the first-10-lines window the checker reads, above the
-script's body, so the window rule above resolves them without a bounce.
+script's body, so the window rule above resolves them without a bounce. **A reproducer's check
+reads the defect's behaviour, never one fix route's text**: it runs the path the defect takes, or
+reads the defect's own code, so any fix that removes the defect turns it to defect not
+demonstrated — a grep for one fix's call, or for call-site text a fix leaves unchanged, makes the
+parent re-author it (kan-1022's F7 grepped `catalog.byKeys(keys)`).
 
 **A reproducer is authored, or repaired, only with both recorded exits.** Before the finding's
 reproducer line is recorded, its author has executed the script in both directions and recorded
````

  - [ ] **Step 2: Verify.** `cd stats && go test ./internal/guard/ -run 'TestRenderSlotPrompt' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-installed-citations.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** **none** — a brief sentence; the render test guards its placeholders.
**Commit:** fix(review-panel): a reproducer reads the defect's behaviour, never a fix's text
**After:** none

- [ ] 6. A mutation reproducer's tree is the worktree it runs in
  - [ ] **Step 1: Apply the fix.** Apply this diff:

````diff verified:git apply at 86c20900 after Task 5; go test ./internal/guard/ -run TestRenderSlotPrompt passes with it applied, where a wording naming a placeholder failed it
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index b3b4cdea..32415374 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -752,7 +752,10 @@ a test. A mutation no test catches is a
 withdrawn with a reason under the fix round's handback. A surviving mutant's reproducer carries the exact line
 `# mutation-reproducer` within its first 10 lines — the declaration
 `run-reproducer.sh` reads as the mutation convention, since the build succeeding
-with the mutation landed is the bug present, the reverse of the generic exit-code contract.
+with the mutation landed is the bug present, the reverse of the generic exit-code contract. Its
+tree is `$PWD`, the worktree `run-reproducer.sh` passes — never the slot's throwaway copy it was
+authored in, which is gone once the dispatch closes (kan-1022's F8 pointed at that copy and
+demonstrated nothing).
 
 ### Principles
 
````

  - [ ] **Step 2: Verify.** `cd stats && go test ./internal/guard/ -run 'TestRenderSlotPrompt' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-installed-citations.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** **none** — a brief sentence; the render test guards its placeholders.
**Commit:** fix(review-panel): a mutation reproducer's tree is the worktree it runs in
**After:** Task 5
