# self-review-kan-887-flow-fix-delete-the-retained-bash-twins-and

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-887's four self-review findings — F-A (a panel re-run closed `clean`), F-B
(lettered `Case <N>` labels), F-C (Go-tree audits in `pipeline-fix verify`), F-D (verify skips what
a green full-suite run already covered).

**Architecture:** three small guard changes, each pinned by one new test, and one change to
`scripts/run-verify-list.sh` with the contract text that routes the full-suite run and verify
through it.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-21-self-review-kan-887-flow-fix-delete-the-retained-bash-twins-and-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/kan-887-flow-fix-delete-the-retained-bash-twins-and`, branch `spectre/kan-887-flow-fix-delete-the-retained-bash-twins-and`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- Each task's diff is the unified diff its fix step gives, applied with `git apply` from the worktree's root; each was applied and checked in this worktree at `b9e7118b`. Save it by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- F-A: only `completed` and `clean` qualify; `fix`, `timed-out`, `aborted` and an empty outcome still do not.
- F-B: the Go regex and the Python one stay identical (`TestTaskFieldParseMatchesPython` compares their parses); a bare `case 13:` never satisfies a declared `Case 13b`.
- F-D: no skip on a dirty tree, without git, or when `git diff` fails; nothing is written into the tree; a run with any red writes no sha; a command declaring no `# inputs:` always runs unless HEAD is the green sha.

- [x] 1. Count a panel re-run closed clean
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Save Step 2's diff, apply its
    test hunk alone (`git apply --include='stats/internal/guard/check_panel_findings_closed_test.go' <the saved diff>`)
    and run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckPanelFindingsClosed$' -count=1`:
    `case 32: a re-run that ended clean covers the finding` fails.
    <!-- measured: the test hunk alone, go test -run 'TestCheckPanelFindingsClosed$' @ b9e7118b → FAIL on case 32 -->
  - [x] **Step 2: Apply the fix** — the rest (`git apply --exclude='stats/internal/guard/check_panel_findings_closed_test.go' <the saved diff>`):

````diff verified:applied with git apply to this worktree at b9e7118b; go test ./internal/guard/ -count=1 passes, case 32 fails with panelfindingsclosed.go reverted, go vet and gofmt -l clean
diff --git a/stats/internal/guard/check_panel_findings_closed_test.go b/stats/internal/guard/check_panel_findings_closed_test.go
index dda80579..0d380e12 100644
--- a/stats/internal/guard/check_panel_findings_closed_test.go
+++ b/stats/internal/guard/check_panel_findings_closed_test.go
@@ -212,6 +212,10 @@ func TestCheckPanelFindingsClosed(t *testing.T) {
 			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed", Round: 1, Slot: "primary"}, cfcFinding{Ref: "F2", Severity: "Important", Status: "fixed", Round: 2, Slot: "bugbot", Supersedes: "F1"}),
 			disp: cfcDispatchJSON(cfcDispatch{Key: "panel-3-primary", Role: "reviewer", Slot: "primary", Outcome: "completed"}),
 			want: 1, needle: []string{"in any later round: F2 (bugbot)\n"}, absent: "F1 ("},
+		{label: "case 32: a re-run that ended clean covers the finding",
+			json: cfcJSON(cfcFinding{Ref: "F1", Status: "fixed", Slot: "primary"}),
+			disp: cfcDispatchJSON(cfcDispatch{Key: "panel-1-primary", Role: "reviewer", Slot: "primary", Outcome: "clean"}),
+			want: 0, needle: []string{"FINDINGS-CLOSED"}, clean: true},
 	}
 	for _, c := range cases {
 		t.Run(c.label, func(t *testing.T) {
diff --git a/stats/internal/guard/panelfindingsclosed.go b/stats/internal/guard/panelfindingsclosed.go
index e239f738..881a5774 100644
--- a/stats/internal/guard/panelfindingsclosed.go
+++ b/stats/internal/guard/panelfindingsclosed.go
@@ -140,14 +140,17 @@ func checkPanelFindingsClosed(args []string, env Env, stdout, stderr io.Writer)
 	// A FINDING IS RECORDED `fixed` ONLY AFTER THE RE-RUN THAT VERIFIES IT
 	// (review-panel.md's **Recording findings**, KAN-770), and the store's
 	// only witness of that re-run is the raising slot's own dispatch row at
-	// a later round, ended `completed`. A `fixed` finding whose slot has
+	// a later round, ended `completed` or `clean`. A `fixed` finding whose slot has
 	// none stands verified before its verification existed, which is the
 	// defect this class exists to catch: the ordering was luck, not
 	// discipline. A slot's re-run qualifies when its row carries role
 	// `reviewer`, a key of the shape panel-<round>-... (a handshake retry's
 	// trailing -retry tolerated; a panel-fix or task key never matches), a
 	// round strictly greater than the finding's, an outcome exactly
-	// `completed` -- a timed-out or never-ended dispatch is not a clean
+	// `completed` or `clean` -- review-panel.md closes a slot `completed`,
+	// while `flow record dispatch end` also takes a gated reviewer's
+	// `clean`, and 50 panel rows in the store end that way (KAN-887's
+	// self-review); a timed-out or never-ended dispatch is not a clean
 	// re-run -- and a slot whose `+`-components cover every component of the
 	// finding's: a bundled re-run covers its members, a solo re-run of one
 	// member does not cover a joined finding. Withdrawn findings claim no
@@ -173,7 +176,7 @@ func checkPanelFindingsClosed(args []string, env Env, stdout, stderr io.Writer)
 		required := strings.Split(slot, "+")
 		var missing []string
 		for _, d := range dispatches {
-			if d.role != "reviewer" || d.outcome != "completed" || d.round <= round {
+			if d.role != "reviewer" || (d.outcome != "completed" && d.outcome != "clean") || d.round <= round {
 				continue
 			}
 			provides := strings.Split(d.slot, "+")
````

  - [x] **Step 3: Verify** — `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean with no `gofmt -l` output, then
    `cd stats && go test ./internal/guard/ -run 'TestCheckPanelFindingsClosed' -count=1` passes.

**Build:** green
**Files:** `stats/internal/guard/panelfindingsclosed.go`, `stats/internal/guard/check_panel_findings_closed_test.go`
**Tests:** Case 32
**Regression:** Case 32 — reverting the commit leaves a re-run closed `clean` uncounted, and the case exits 1 naming F1 as fixed with no clean re-run.
**Baseline:** before=31 after=32
<!-- measured: grep -c 'label: "case ' stats/internal/guard/check_panel_findings_closed_test.go @ b9e7118b -->
**Commit:** fix(guard): count a panel re-run closed clean
**After:** none

**Decision:** clean-rerun-counts

- [x] 2. Accept a lettered Case label
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Save Step 2's diff, apply its
    test hunks alone (`git apply --include='stats/internal/guard/check_task_commit_fields_test.go' --include='scripts/test-check-plan-shape.sh' <the saved diff>`),
    then run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckTaskCommitFields$/case_163' -count=1`
    (fails: `the bare number does not satisfy a lettered label`) and `scripts/test-check-plan-shape.sh`
    (fails: `case 11b` with `declares nothing`).
    <!-- measured: the test hunks alone, both commands @ b9e7118b → FAIL on case 163 and case 11b -->
  - [x] **Step 2: Apply the fix** — the rest (`git apply --exclude='stats/internal/guard/check_task_commit_fields_test.go' --exclude='scripts/test-check-plan-shape.sh' <the saved diff>`):

````diff verified:applied with git apply to this worktree at b9e7118b; go test ./internal/guard/ -count=1 and scripts/test-check-plan-shape.sh pass, case 163 fails with taskcommitfields.go reverted and case 11b with the Python regex reverted, go vet and gofmt -l clean
diff --git a/scripts/check-task-commit-fields.py b/scripts/check-task-commit-fields.py
index 1ae184a8..17142d7b 100755
--- a/scripts/check-task-commit-fields.py
+++ b/scripts/check-task-commit-fields.py
@@ -278,7 +278,8 @@ check-plan-shape.py's F4 exempts the annotated value for the same reason.
 Only an opening `none`
 counts: the word appearing mid-sentence, as in "added `test_alpha`, none
 of the other paths change", does not. Short of that, every `Case <N>`
-label mentioned anywhere in the field's text becomes a declared test named
+label mentioned anywhere in the field's text — `<N>` a number with an
+optional letter suffix, `Case 13b` — becomes a declared test named
 "Case <N>", and it is checked by asking whether `Case <N>` (any case, e.g.
 as a `# Case <N>:` comment or a `case <N>:` pass/fail line — the
 convention this repository's own test harnesses already use, per
@@ -361,7 +362,7 @@ FIELD_RE = re.compile(
     r":\*\*\s*(.*)$"
 )
 BACKTICK_RE = re.compile(r"`([^`]+)`")
-CASE_LABEL_RE = re.compile(r"\bCase\s+(\d+)\b", re.IGNORECASE)
+CASE_LABEL_RE = re.compile(r"\bCase\s+(\d+[a-z]?)\b", re.IGNORECASE)
 # NONE_OPEN_RE — a `Tests:` value that OPENS with the literal `none`:
 # case-insensitive, leading Markdown emphasis and whitespace stripped. This
 # is the one definition of the shape; check-plan-shape.py imports it rather
diff --git a/scripts/test-check-plan-shape.sh b/scripts/test-check-plan-shape.sh
index d280fd1c..562552f1 100755
--- a/scripts/test-check-plan-shape.sh
+++ b/scripts/test-check-plan-shape.sh
@@ -327,6 +327,19 @@ new_fixture
 run_guard "$TASKS_MD"
 [ "$RC" -eq 0 ] && pass "case 11: a Case N label satisfies Tests:" || fail "case 11: rc=$RC out=$OUT"
 
+# ===========================================================================
+# Case 11b: a Tests: field carrying a lettered Case label (`Case 13b`, the
+# shape 15 Go test files name their cases) -> F6 does not fire.
+# ===========================================================================
+new_fixture
+{
+  printf -- '- [ ] 1. Lettered case-labeled Tests field\n\n'
+  printf '**Files:** `a.py`\n\n'
+  printf '**Tests:** Case 13b pins the new behaviour\n'
+} > "$TASKS_MD"
+run_guard "$TASKS_MD"
+[ "$RC" -eq 0 ] && pass "case 11b: a lettered Case label satisfies Tests:" || fail "case 11b: rc=$RC out=$OUT"
+
 # ===========================================================================
 # Case 12: a Tests: field carrying a backtick token -> F6 does not fire.
 # ===========================================================================
diff --git a/skills/flow/brainstorm-planner.md b/skills/flow/brainstorm-planner.md
index 7e72f9a2..2d27bcca 100644
--- a/skills/flow/brainstorm-planner.md
+++ b/skills/flow/brainstorm-planner.md
@@ -405,7 +405,8 @@ with `**Build:**` per **The build-green tag**
   and their like — is expanded by the writer, never left in the field for
   the guard to expand.
 - `**Tests:**` — the names of the tests this task adds. **The field is parsed, not read:** a
-  `Case <N>` label is checked by that label alone and backticks beside it parse as nothing;
+  `Case <N>` label — `<N>` a number with an optional letter suffix, `Case 13b` — is checked by
+  that label alone and backticks beside it parse as nothing;
   otherwise every backticked token becomes a declared test name the commit's diff and the tree
   are searched for; any other prose in it is invisible to the check — write what covers the task
   outside the field, never inside it. A task adding none writes a field opening with the literal
diff --git a/stats/internal/guard/check_task_commit_fields_test.go b/stats/internal/guard/check_task_commit_fields_test.go
index ebe3441f..1064155b 100644
--- a/stats/internal/guard/check_task_commit_fields_test.go
+++ b/stats/internal/guard/check_task_commit_fields_test.go
@@ -1369,6 +1369,13 @@ After the change: 186 tests
 **Tests:** none — repairs existing tests: ¤test_alpha¤
 **Commit:** repair alpha
 **Build:** green
+`),
+	"163": bt(`- [ ] 163. A lettered case label
+
+**Files:** ¤guard_test.go¤
+**Tests:** Case 13b: a lettered label is checked by that label
+**Commit:** add lettered case
+**Build:** green
 `),
 }
 
@@ -3075,6 +3082,18 @@ func TestCheckTaskCommitFields(t *testing.T) {
 			res := fx.run(b1, "1", b1+"="+sha1+","+b2+"="+sha2, "", a, "x-repo-change")
 			ok(t, "case 162: a link-route satellite resolves the canonical plan under the commit map", res.rc == 0, res)
 		}},
+		{"case 163", func(t *testing.T) {
+			r := fx.repo(t, "change-a")
+			r.plan(tcfPlans["163"])
+			sha := r.commit("add lettered case", "guard_test.go", "t.Run(\"case 13b: a lettered label\", nil)\n")
+			res := r.run("163", sha)
+			ok(t, "case 163: a lettered Case label found in the diff passes", res.rc == 0, res)
+			r2 := fx.repo(t, "change-a")
+			r2.plan(tcfPlans["163"])
+			bare := r2.commit("add lettered case", "guard_test.go", "t.Run(\"case 13: the bare number\", nil)\n")
+			res = r2.run("163", bare)
+			ok(t, "case 163: the bare number does not satisfy a lettered label", res.rc == 1 && has(res.out, "Case 13b", "not found in the diff"), res)
+		}},
 	}
 	for _, c := range cases {
 		t.Run(c.name, func(t *testing.T) {
diff --git a/stats/internal/guard/taskcommitfields.go b/stats/internal/guard/taskcommitfields.go
index f18a5e90..86ab4eec 100644
--- a/stats/internal/guard/taskcommitfields.go
+++ b/stats/internal/guard/taskcommitfields.go
@@ -284,7 +284,7 @@ func tcfSelectTask(lines []string, taskID string) *tcfTaskBody {
 
 var (
 	tcfBacktickRE  = regexp.MustCompile("`([^`]+)`")
-	tcfCaseLabelRE = regexp.MustCompile(`(?i)\bCase[` + tcfWS + `]+([0-9]+)\b`)
+	tcfCaseLabelRE = regexp.MustCompile(`(?i)\bCase[` + tcfWS + `]+([0-9]+[a-z]?)\b`)
 	tcfNoneOpenRE  = regexp.MustCompile(`(?i)^[` + tcfWS + `*_]*none\b`)
 	tcfCamelRE     = regexp.MustCompile(`\b[A-Za-z][a-z0-9]*(?:[A-Z][a-z0-9]*[a-z][a-z0-9]*)+\b`)
 )
@@ -1274,7 +1274,7 @@ func tcfCheckDeclaredFiles(task tcfTask, changed []string) []string {
 // must not arrive under a rule under which it was never declared.
 //
 // tcfTestsParseRule is TESTS_PARSE_RULE.
-const tcfTestsParseRule = "a **Tests:** field is parsed, not read: a `Case <N>` label is checked by that label alone and backticks beside it parse as nothing, any other backticked token is a declared test name, the tree check also treats a bare camelCase identifier as one, and a task adding no tests opens the field with `none`; coverage prose belongs outside the field"
+const tcfTestsParseRule = "a **Tests:** field is parsed, not read: a `Case <N>` label, `<N>` a number with an optional letter suffix (`Case 13b`), is checked by that label alone and backticks beside it parse as nothing, any other backticked token is a declared test name, the tree check also treats a bare camelCase identifier as one, and a task adding no tests opens the field with `none`; coverage prose belongs outside the field"
 
 // Moved from check-task-commit-fields.py's DIFF_LEADING_MARKER_RE (KAN-760 deleted the
 // Python copy once this port replaced it; panel round 0, F9):
````

  - [x] **Step 3: Verify** — `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean with no `gofmt -l` output;
    `cd stats && go test ./internal/guard/ -run 'TestCheckTaskCommitFields|TestTaskFieldParseMatchesPython' -count=1`
    passes; `scripts/test-check-plan-shape.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.

**Build:** green
**Files:** `stats/internal/guard/taskcommitfields.go`, `stats/internal/guard/check_task_commit_fields_test.go`, `scripts/check-task-commit-fields.py`, `scripts/test-check-plan-shape.sh`, `skills/flow/brainstorm-planner.md`
**Tests:** Case 163, Case 11b
**Regression:** Case 163 — reverting the Go regex makes a commit carrying only `case 13:` pass a task declaring `Case 13b`; Case 11b — reverting the Python regex makes `check-plan-shape` report a `**Tests:** Case 13b` field as declaring nothing.
**Commit:** fix(guard): accept a lettered Case label
**After:** none

**Decision:** lettered-case-label

- [x] 3. Select the Go-tree audits when a Go file is touched
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Save Step 2's diff, apply its
    test hunk alone (`git apply --include='stats/internal/guard/pipeline_fix_test.go' <the saved diff>`)
    and run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestPipelineFixChecksGoAudits' -count=1`:
    it fails with `a touched Go file: … lacks scripts/test-flow-addr-declaration.sh`.
    <!-- measured: the test hunk alone, go test -run TestPipelineFixChecksGoAudits @ b9e7118b → FAIL, lacks scripts/test-flow-addr-declaration.sh -->
  - [x] **Step 2: Apply the fix** — the rest (`git apply --exclude='stats/internal/guard/pipeline_fix_test.go' <the saved diff>`):

````diff verified:applied with git apply to this worktree at b9e7118b; go test ./internal/guard/ -count=1 passes, TestPipelineFixChecksGoAudits fails with the goTouched loop emptied, go vet and gofmt -l clean
diff --git a/stats/internal/guard/pipeline_fix_test.go b/stats/internal/guard/pipeline_fix_test.go
index b103caa2..4e30d01c 100644
--- a/stats/internal/guard/pipeline_fix_test.go
+++ b/stats/internal/guard/pipeline_fix_test.go
@@ -4,6 +4,7 @@ import (
 	"bytes"
 	"os"
 	"path/filepath"
+	"slices"
 	"strings"
 	"testing"
 )
@@ -196,6 +197,32 @@ func TestPipelineFixVerify(t *testing.T) {
 	}
 }
 
+// A touched Go file selects the harnesses auditing the whole Go tree, which
+// no Go file's name leads to; a touched script alone does not.
+func TestPipelineFixChecksGoAudits(t *testing.T) {
+	t.Parallel()
+	wt := t.TempDir()
+	writeFile(t, wt+"/.flow/project.md", pfProject)
+	audits := []string{"scripts/test-flow-addr-declaration.sh", "scripts/test-git-config-pins.sh"}
+	for _, h := range audits {
+		writeExec(t, wt+"/"+h, "#!/usr/bin/env bash\n")
+	}
+	pe := pfEnv{projectGet: tcfScriptsDir(t) + "/project-get.sh"}
+	env := Env{Getenv: os.Getenv}
+	cmds, rc := pfChecks(env, pe, wt, []string{"stats/internal/guard/x.go"})
+	for _, h := range audits {
+		if rc != 0 || !slices.Contains(cmds, h) {
+			t.Fatalf("a touched Go file: rc %d, %q lacks %s", rc, cmds, h)
+		}
+	}
+	cmds, _ = pfChecks(env, pe, wt, []string{"scripts/foo.sh"})
+	for _, h := range audits {
+		if slices.Contains(cmds, h) {
+			t.Fatalf("a touched script alone: %q selects %s", cmds, h)
+		}
+	}
+}
+
 func TestPipelineFixLand(t *testing.T) {
 	pfIdentity(t)
 	reviewed := lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed", "fix completed", "review-1 clean")
diff --git a/stats/internal/guard/pipelinefix.go b/stats/internal/guard/pipelinefix.go
index f4ae6748..6b7fdd65 100644
--- a/stats/internal/guard/pipelinefix.go
+++ b/stats/internal/guard/pipelinefix.go
@@ -309,12 +309,19 @@ func pfTouched(env Env, pe pfEnv, wt, base string) ([]string, bool) {
 	return strings.Fields(out), rc == 0
 }
 
+// pfGoAudits is every scripts/ harness auditing stats/ Go source as a whole:
+// test-flow-addr-declaration.sh reads stats/cmd/flow, test-git-config-pins.sh
+// stats/internal/guard.
+var pfGoAudits = []string{"scripts/test-flow-addr-declaration.sh", "scripts/test-git-config-pins.sh"}
+
 // pfChecks is the lint lines and tests the touched paths need: every guard
 // line of `## lint`, its `cd stats` lines only for touched Go and its
 // `cd stats/web` lines only for a touched SPA file; `go test` of each
 // touched package, the scripts/test-*.sh harness of a touched script or
 // hook, and vitest of a touched SPA test file or the test beside a touched
-// source file.
+// source file. A touched Go file also selects pfGoAudits: harnesses that
+// audit the whole Go tree, named after no Go file, which main landed red
+// without twice (KAN-887's self-review).
 func pfChecks(env Env, pe pfEnv, wt string, touched []string) ([]string, int) {
 	get := exec.Command(pe.projectGet, wt, "lint")
 	var body bytes.Buffer
@@ -346,6 +353,13 @@ func pfChecks(env Env, pe pfEnv, wt string, touched []string) ([]string, int) {
 			cmds = append(cmds, c)
 		}
 	}
+	if goTouched {
+		for _, h := range pfGoAudits {
+			if exists(h) {
+				add(h)
+			}
+		}
+	}
 	for _, p := range touched {
 		base := filepath.Base(p)
 		stem := strings.TrimSuffix(base, filepath.Ext(base))
````

  - [x] **Step 3: Verify** — `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean with no `gofmt -l` output, then
    `cd stats && go test ./internal/guard/ -run 'TestPipelineFix' -count=1` passes.

**Build:** green
**Files:** `stats/internal/guard/pipelinefix.go`, `stats/internal/guard/pipeline_fix_test.go`
**Tests:** `TestPipelineFixChecksGoAudits`
**Regression:** `TestPipelineFixChecksGoAudits` — reverting the commit leaves both audit harnesses out of the checks a touched Go file selects, and the test fails naming the missing harness.
**Baseline:** before=10 after=11
<!-- measured: grep -c '^func Test' stats/internal/guard/pipeline_fix_test.go @ b9e7118b -->
**Commit:** fix(guard): run the Go-tree audits for a touched Go file
**After:** none

**Decision:** go-audits-on-go

- [x] 4. Skip verify tests a green run already covered
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Save Step 2's diff, apply its
    harness hunk alone (`git apply --include='scripts/test-run-verify-list.sh' <the saved diff>`) and run
    `scripts/test-run-verify-list.sh`: it fails on `green sha`, `every command skipped at the green sha`,
    `a command re-ran at the green sha` and `a command whose inputs the delta skips is skipped`.
    <!-- measured: the harness hunk alone, scripts/test-run-verify-list.sh | grep -c FAIL → 4 @ b9e7118b -->
  - [x] **Step 2: Apply the fix** — the rest (`git apply --exclude='scripts/test-run-verify-list.sh' <the saved diff>`):

````diff verified:applied with git apply to this worktree at b9e7118b; scripts/test-run-verify-list.sh and scripts/test-git-config-pins.sh pass, the harness fails on the unchanged run-verify-list.sh, check-verbatim-moves.sh clean with verbatim-moves.txt
diff --git a/.flow/project.md b/.flow/project.md
index cbf7be6f..1442c118 100644
--- a/.flow/project.md
+++ b/.flow/project.md
@@ -70,6 +70,7 @@ forbids any agent action touching it at all.
 ```bash
 scripts/run-guard-tests.sh
 cd stats && make -s web-dist && go test ./... -race -count=1
+# inputs: stats/web/ stats/Makefile
 cd stats && make -s web-dist && cd web && npm test
 ```
 
@@ -78,6 +79,10 @@ refuses to compile without it, so `make -s web-dist` (`stats/Makefile`) opens ev
 `## lint` that compiles `internal/web` or needs `stats/web/node_modules`: it runs `make web-build`,
 `npm ci && npm run build` in `stats/web`, only when the build is absent.
 
+Only the SPA line declares `# inputs:` (`skills/flow-contracts/project-configuration.md`): the guard
+harnesses and the Go tests read files across the whole repository — `stats/internal/guard`'s tests
+open `skills/`, `scripts/` and `setup.sh` through `../../../` — so both always run.
+
 `scripts/run-guard-tests.sh` discovers every `scripts/test-*.sh` by glob and runs them concurrently
 through `scripts/lib/parallel.sh`; a new harness added to `scripts/` is picked up automatically, with
 no edit needed here. It points `FLOW_GUARD_CACHE_DIR` at its own temp directory, so a ported
diff --git a/scripts/run-verify-list.sh b/scripts/run-verify-list.sh
index 2eda39b6..3dddff30 100755
--- a/scripts/run-verify-list.sh
+++ b/scripts/run-verify-list.sh
@@ -9,7 +9,9 @@
 #
 # THE LIST is every non-blank, non-`#` line inside a ```bash fence of
 # `project-get.sh <worktree> <list>`'s output, in the order printed; `lint`
-# appends `check-spec-reach.sh <worktree>`. A line holding a `<word>`
+# appends `check-spec-reach.sh <worktree>`. A `# inputs: <pathspec>...`
+# line of the `test` list declares what the one command below it reads; every
+# other `#` line is a comment. A line holding a `<word>`
 # placeholder — `./gradlew verifyChange :<module>:test --tests '<selectors>'`
 # — is a template, not a command: the shell reads `<module>` as a redirect,
 # so the list is refused before anything runs (kan-983 self-review). Each command runs from
@@ -19,6 +21,20 @@
 # <worktree> first; every command inherits them. This script retries
 # nothing — the one inline re-run of a failing command stays the caller's.
 #
+# THE GREEN SHA (KAN-887's self-review: verify re-ran the whole list minutes
+# after a green full-suite run, the delta a test rename). A `test` run that
+# exits 0 on a tree clean before and after it — `git status --porcelain
+# --untracked-files=all` empty — writes HEAD to the worktree's own git dir
+# (`git rev-parse --git-path flow-test-green`), never into the tree. A later
+# `test` run on a clean tree skips every command when HEAD is that sha, and
+# otherwise skips each command whose `# inputs:` pathspecs `git diff --quiet
+# <sha> HEAD` finds untouched. A command declaring no inputs reads everything
+# and always runs; a dirty tree, no git, no recorded sha or a failing diff
+# runs the whole list. A skip passes and prints instead of output:
+#   - `<command>` — skipped, no input changed since <sha> passed the list
+# ponytail: the sha ignores the exported KEY=value lines; a run under other
+# isolation values reuses the green, so key the file by them if they ever vary.
+#
 # OUTPUT, one block per command in the order run, stdout and stderr together:
 #   - `<command>` — exit <n>
 #     <each output line, indented two spaces; past 20 lines,
@@ -50,13 +66,22 @@ case $? in
   *) exit 2 ;;
 esac
 
+# INPUTS[i] is the pathspec list the `# inputs:` line above CMDS[i] named.
 CMDS=()
+INPUTS=()
+PENDING=""
 while IFS= read -r line; do
+  if [[ "$line" =~ ^[[:space:]]*#\ inputs:(.*)$ ]]; then
+    PENDING="${BASH_REMATCH[1]}"
+    continue
+  fi
   CMDS+=("$line")
+  INPUTS+=("$PENDING")
+  PENDING=""
 done < <(printf '%s\n' "$BODY" | awk '
   /^```bash[[:space:]]*$/ { f = 1; next }
   /^```/                  { f = 0; next }
-  f && NF && $1 !~ /^#/
+  f && NF && ($1 !~ /^#/ || /^[[:space:]]*# inputs:/)
 ')
 
 for i in "${!CMDS[@]}"; do
@@ -66,6 +91,27 @@ for i in "${!CMDS[@]}"; do
   fi
 done
 
+HEAD_SHA=""
+SINCE=""
+GREEN_FILE=""
+if [ "$LIST" = test ] && HEAD_SHA="$(git -C "$WT" rev-parse -q --verify HEAD 2>/dev/null)" &&
+  STATUS="$(git -C "$WT" status --porcelain --untracked-files=all 2>/dev/null)" && [ -z "$STATUS" ] &&
+  GREEN_FILE="$(git -C "$WT" rev-parse --path-format=absolute --git-path flow-test-green 2>/dev/null)"; then
+  SINCE="$(cat "$GREEN_FILE" 2>/dev/null || true)"
+else
+  HEAD_SHA=""
+fi
+
+# unchanged <inputs> -> 0 when a recorded green sha covers the command.
+unchanged() {
+  local specs
+  [ -n "$SINCE" ] || return 1
+  [ "$SINCE" = "$HEAD_SHA" ] && return 0
+  [ -n "${1// /}" ] || return 1
+  read -ra specs <<<"$1"
+  git -C "$WT" diff --quiet "$SINCE" "$HEAD_SHA" -- "${specs[@]}" 2>/dev/null
+}
+
 OUT="$(mktemp "${TMPDIR:-/tmp}/run-verify-list.XXXXXX")"
 trap 'rm -f "$OUT"' EXIT
 
@@ -87,8 +133,16 @@ run() {
 }
 
 for i in "${!CMDS[@]}"; do
+  if unchanged "${INPUTS[$i]}"; then
+    printf -- '- `%s` — skipped, no input changed since %s passed the list\n' "${CMDS[$i]}" "$SINCE"
+    continue
+  fi
   run "${CMDS[$i]}" "${CMDS[$i]}"
 done
+if [ -n "$HEAD_SHA" ] && [ "$FAILED" -eq 0 ] &&
+  STATUS="$(git -C "$WT" status --porcelain --untracked-files=all 2>/dev/null)" && [ -z "$STATUS" ]; then
+  printf '%s\n' "$HEAD_SHA" >"$GREEN_FILE"
+fi
 if [ "$LIST" = lint ]; then
   run "check-spec-reach.sh $WT" "\"$SCRIPT_DIR/check-spec-reach.sh\" \"$WT\""
 fi
diff --git a/scripts/test-run-verify-list.sh b/scripts/test-run-verify-list.sh
index b19feb3d..3092ed49 100755
--- a/scripts/test-run-verify-list.sh
+++ b/scripts/test-run-verify-list.sh
@@ -81,6 +81,47 @@ expect_rc "a placeholder line refuses the list" 2
 expect_has "the refusal names the line" "test line is a template, not a command — move it out of the fence: ./gradlew verifyChange :<module>:test --tests '<selectors>'"
 expect_lacks "nothing runs before the refusal" 'ran-first'
 
+# THE GREEN SHA: a git worktree whose `test` list counts its own runs.
+G="$ROOT/git"
+mkdir -p "$G/.flow" "$G/web" "$G/go"
+cat > "$G/.flow/project.md" <<'CFG'
+## test
+
+```bash
+# inputs: web/
+echo ran >> "$(git rev-parse --git-dir)/web-runs"; echo web
+# inputs: go/
+echo go
+echo all
+```
+CFG
+echo a > "$G/web/a"
+echo b > "$G/go/b"
+git -C "$G" init -q
+git -C "$G" -c user.name=t -c user.email=t@t add -A
+git -C "$G" -c user.name=t -c user.email=t@t commit -q -m one
+SHA1="$(git -C "$G" rev-parse HEAD)"
+run_bin "$G" test
+expect_rc "a first green run" 0
+expect_has "a first run runs every command" '- `echo all` — exit 0'
+GREEN="$(git -C "$G" rev-parse --path-format=absolute --git-path flow-test-green)"
+[ "$(cat "$GREEN" 2>/dev/null)" = "$SHA1" ] && pass "a green run records HEAD in the git dir" || fail "green sha: $(cat "$GREEN" 2>/dev/null)"
+[ -z "$(git -C "$G" status --porcelain)" ] && pass "the record leaves the tree clean" || fail "the record dirtied the tree"
+run_bin "$G" test
+expect_rc "a re-run at the green sha" 0
+expect_has "every command skipped at the green sha" "- \`echo all\` — skipped, no input changed since $SHA1 passed the list"
+[ "$(wc -l < "$(git -C "$G" rev-parse --absolute-git-dir)/web-runs")" -eq 1 ] && pass "nothing re-ran at the green sha" || fail "a command re-ran at the green sha"
+echo c > "$G/go/b"
+git -C "$G" -c user.name=t -c user.email=t@t commit -q -am two
+run_bin "$G" test
+expect_has "a command whose inputs the delta skips is skipped" "- \`echo ran >> \"\$(git rev-parse --git-dir)/web-runs\"; echo web\` — skipped, no input changed since $SHA1"
+expect_has "a command whose inputs the delta touches runs" $'- `echo go` — exit 0'
+expect_has "a command declaring no inputs runs" $'- `echo all` — exit 0'
+echo dirty > "$G/web/a"
+run_bin "$G" test
+expect_has "a dirty tree runs the whole list" $'echo web` — exit 0'
+git -C "$G" checkout -q -- web/a
+
 run_bin "$ROOT" build
 expect_rc "an unknown list" 2
 run_bin "$ROOT/missing" lint
diff --git a/skills/flow-contracts/project-configuration.md b/skills/flow-contracts/project-configuration.md
index 1dd3ab54..977ca857 100644
--- a/skills/flow-contracts/project-configuration.md
+++ b/skills/flow-contracts/project-configuration.md
@@ -24,7 +24,7 @@ exactly what that row specifies and nothing else. The row is authoritative.
 | `## run` | How to start the stack and individual services locally — the actual commands, including any parameter that must point at another app's root. |
 | `## stop` | Optional. The command that stops the project's local stack, run by bare `/flow` before its stack-stopped check so nothing holds a port or file handle open in a worktree about to be removed. Same shape as `## worktree setup`: one command per line inside a fence, read literally and run in order from the worktree; prose outside a fence is never run. **Absent means that check is skipped, not failed** — cleanup proceeds on the strength of the other two checks. A body with no fenced command declares no command and is skipped the same way. A project with no stack should say so here rather than omit the key, so its absence is a recorded fact rather than an oversight. |
 | `## credentials` | Local-development-only sign-in credentials (which app, user, password) and what seeds them. Never deployed-environment secrets. |
-| `## test` | The command(s) that run the project's tests. Every non-`#` line inside its bash fence runs literally, as `## lint`'s do: a template with a `<word>` placeholder, such as a verify bundle's `:<module>:test` selector, belongs in prose outside the fence, and `<agents repo>/scripts/run-verify-list.sh` refuses the list while one stands inside it. |
+| `## test` | The command(s) that run the project's tests. Every non-`#` line inside its bash fence runs literally, as `## lint`'s do: a template with a `<word>` placeholder, such as a verify bundle's `:<module>:test` selector, belongs in prose outside the fence, and `<agents repo>/scripts/run-verify-list.sh` refuses the list while one stands inside it. A `# inputs:` line naming pathspecs declares what the one command below it reads, so `run-verify-list.sh` skips that command when nothing it reads changed since the list last passed; a command declaring none reads everything. |
 | `## worktree setup` | Optional. A fenced command block run once per worktree, from the worktree root, immediately after the run creates the worktree — **2. Isolate the workspace** (`skills/flow/implement.md`) on `/flow`, the kickoff section (`skills/flow-fast/SKILL.md`) on a `/flow-fast` creating run — and before anything else touches the tree — the place for a build that a gitignored, embedded artifact needs before the project's first `go test`/`go build` can succeed. Absent means nothing runs. Same shape as `## lint` and `## test`: one command per line inside the fence, read literally and run in order; resolved through `project-get.sh <worktree> "worktree setup"`, whose exit 1 is the absent case. |
 | `## lint` | The command(s) that verify lint, and the auto-fix command to run first. |
 | `## standards` | The project's own written standards: the files the principles reviewer receives, plus any opt-in shared rule the project has adopted. This same list is both the opt-in list and the reviewer's standards list. Absent, no standards resolve: the reviewer gets an empty list, never auto-detected files. |
diff --git a/skills/flow/SKILL-rationale.md b/skills/flow/SKILL-rationale.md
index 2291db93..7885920c 100644
--- a/skills/flow/SKILL-rationale.md
+++ b/skills/flow/SKILL-rationale.md
@@ -178,6 +178,7 @@ text after it is the passage, verbatim.
 
 - *…the baseline is the project's own statement that the failure exists on an unmodified tree* — (KAN-547)
 - *`environment` where the environment itself failed twice* — KAN-510
+- *The test group skips what the list's last green run on a clean tree already covered* — KAN-887's self-review: verify re-ran the whole `## test` list about 14 minutes after a green full-suite run when the only delta was a test-name rename. The skip keys on the tree, never on time: a green sha is reused only for a command whose declared inputs the delta leaves untouched.
 
 ### visual-verify.md — The verifier dispatch
 
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index 7ec32fdd..a5deed4f 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -638,8 +638,10 @@ Every implementer dispatch also carries:
 **The parent runs the full suite — never an implementer.** Once the plan-last group — the group
 holding the last `bundle <k>` line `plan-dispatch-bundles.sh` printed — has passed the guard, by
 direct commit or by its wave's final pick, **the parent itself** runs the resolved `## test` list
-once on the canonical worktree, in the foreground, in the order the context bundle carries it, its
-output through `tail`. Every failure, wherever it lies, **the parent fixes itself** at its root
+once on the canonical worktree, in the foreground, as `run-verify-list.sh <canonical-worktree> test`
+— by hand, in the order the context bundle carries it, only where that exits 3 — its output
+through `tail`; the green sha that run records is what lets **Inline verify**
+(`skills/flow/verify-and-handoff.md`) skip what no later commit touched. Every failure, wherever it lies, **the parent fixes itself** at its root
 cause, as inline fixes are applied: it edits, commits on the route the branch's push state
 dictates (**Panel re-runs**, `skills/flow/review-panel-fix-round.md`), records the round as one
 `dispatches` pair `-role panel-fix -model <parent model> -effort <parent effort> -agent-id inline`
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index 0fb3d876..363e8c93 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -107,6 +107,10 @@ call; the lint and test run, this run's `flow record dispatch begin`, and the le
 are one more.
 
 `run-verify-list.sh <worktree> lint` and `run-verify-list.sh <worktree> test` are the two groups: each runs its list as above, stdin detached, and prints that list's `## Report` lines, its exit codes canonical in its header — never a runner built by hand.
+The test group skips what the list's last green run on a clean tree already covered: every command
+when HEAD is the sha that run recorded, otherwise each command whose `# inputs:` pathspecs the delta
+since that sha leaves untouched; a skipped command's `## Report` line says so, and the script's
+header is canonical for the rule.
 
 **Inline verify — a failing command.** A non-zero exit from any command in the list earns **one**
 inline re-run of that command — the environmental-flake case. A second non-zero exit from the same
````

  - [x] **Step 3: Verify** — `scripts/test-run-verify-list.sh`, `scripts/test-git-config-pins.sh`,
    `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`,
    `scripts/check-verbatim-moves.sh` and `scripts/project-get.sh . test` exit 0, the last printing
    the `# inputs: stats/web/ stats/Makefile` line above the SPA test line.

**Build:** green
**Files:** `scripts/run-verify-list.sh`, `scripts/test-run-verify-list.sh`, `skills/flow/verify-and-handoff.md`, `skills/flow/implement.md`, `skills/flow-contracts/project-configuration.md`, `skills/flow/SKILL-rationale.md`, `.flow/project.md`
**Tests:** `every command skipped at the green sha`, `a command whose inputs the delta skips is skipped`
**Regression:** `every command skipped at the green sha` — reverting the commit re-runs the whole list at the sha a green run recorded; `a command whose inputs the delta skips is skipped` — reverting it runs a command whose declared inputs the delta never touched.
**Commit:** feat(scripts): skip verify tests a green run already covered
**After:** none

**Decision:** green-sha-git-dir
**Decision:** inputs-comment
**Decision:** full-suite-through-script
**Decision:** agents-inputs
