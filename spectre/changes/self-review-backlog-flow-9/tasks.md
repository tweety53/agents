# self-review-backlog-flow-9

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the next five live Backlog Flow findings — KAN-886's four untested ported-guard
refusals, KAN-889's untested commit-map refusals and link route, KAN-890's resumed fix round read
as a retry with no original, KAN-896's list-item block cut that flags an untouched sentence as
reworded, and KAN-898's reviewer-row paragraph that names a pair the store refuses on `zcode`.

**Architecture:** two test-only tasks in `stats/internal/guard/` that pin existing refusals; two
`flow-guard` changes in the same package — `pfdPanelFixRows`' foreign-token originals with the
script header that states the contract, and `vmBlocks`' list-item continuation — each with its
table tests; and one sentence in `skills/flow/review-panel.md`.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-19-self-review-backlog-flow-9-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Every task runs in this worktree, `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-9`, branch `self-review-backlog-flow-9`; every `**Files:**` path is relative to its root. Never touch a main checkout.
- Every task's diff is the unified diff its fix step gives, applied with `git apply` from the worktree's root, checked in order in a scratch clone at `373988df`. Save each by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork. Task 5's sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Every test-bearing task proves its tests bite: each new test fails with the mutation its RED step names, and passes with it restored.
- Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- Task 1: each subtest's needle is the refusal's own line, so the guard's later `not a directory` or verdict path cannot satisfy it.
- Task 3: case 6's two foreign originals still pass; a round only another token touched is never judged; a foreign `-retry` is no original.
- Task 4: an unindented line after a list item still opens its own block; `TestVMSentences` is unchanged.

- [ ] 1. Pin the four untested ported-guard refusals (KAN-886)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n -F -e 'FLOW_GUARD_REPO_ROOT is unset' -e 'cannot read the state record' -e 'a null status' stats/internal/guard/check_references_test.go stats/internal/guard/check_installed_citations_test.go stats/internal/guard/check_panel_reproducers_test.go stats/internal/guard/check_unfinished_work_test.go; echo $?`: it prints only `1`.
    <!-- measured: the grep above in this worktree @ 373988df → exit 1, no line -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 373988df; gofmt -l internal/guard clean; go test ./internal/guard -run 'TestCheckPanelReproducers|TestCheckUnfinishedWork|TestCheckReferences|TestCheckInstalledCitations' passes; each new subtest fails with its refusal removed as Step 3 names
diff --git a/stats/internal/guard/check_installed_citations_test.go b/stats/internal/guard/check_installed_citations_test.go
index c198d138..7ec3abbf 100644
--- a/stats/internal/guard/check_installed_citations_test.go
+++ b/stats/internal/guard/check_installed_citations_test.go
@@ -233,6 +233,14 @@ func TestCheckInstalledCitations(t *testing.T) {
 		t.Parallel()
 		refused(t, cicRun(t, "", nil), "CHECK_INSTALLED_CITATIONS_ROOT")
 	})
+	// KAN-886: with no override and no shim-exported root, the guard names
+	// the missing variable rather than scanning its working directory.
+	t.Run("no override and FLOW_GUARD_REPO_ROOT unset: exit 2, stderr names it, stdout empty", func(t *testing.T) {
+		t.Parallel()
+		var out, errOut bytes.Buffer
+		rc := checkInstalledCitations(nil, crEnv(t.TempDir(), nil), &out, &errOut)
+		refused(t, cicResult{rc, out.String(), errOut.String()}, cicName+": FLOW_GUARD_REPO_ROOT is unset — run scripts/check-installed-citations.sh, which sets it")
+	})
 	t.Run("root is not a directory: exit 2, stdout empty", func(t *testing.T) {
 		t.Parallel()
 		refused(t, cicRun(t, filepath.Join(t.TempDir(), "does-not-exist"), nil), "")
diff --git a/stats/internal/guard/check_panel_reproducers_test.go b/stats/internal/guard/check_panel_reproducers_test.go
index 154d7f51..d6459d6e 100644
--- a/stats/internal/guard/check_panel_reproducers_test.go
+++ b/stats/internal/guard/check_panel_reproducers_test.go
@@ -229,4 +229,12 @@ func TestCheckPanelReproducers(t *testing.T) {
 		got, out := cprRun(t, cprBin(t, "exit 0", "exit 1"), wt, "demo", findings)
 		cprExpect(t, got, out, 2, "cannot determine anything")
 	})
+	// KAN-886: `flow state get` printing a JSON null at exit 0 decodes into
+	// a nil map with no error; without the nil check the guard read the
+	// findings and answered REPRODUCERS-OK on a change it never saw.
+	t.Run("case 32: a null state record is cannot-answer, never clean", func(t *testing.T) {
+		t.Parallel()
+		got, out := cprRun(t, cprBin(t, "echo null; exit 0", "exit 1"), wt, "demo", findings)
+		cprExpect(t, got, out, 2, "cannot read the state record for 'demo' — cannot determine anything")
+	})
 }
diff --git a/stats/internal/guard/check_references_test.go b/stats/internal/guard/check_references_test.go
index dfc35fd9..fa1d349c 100644
--- a/stats/internal/guard/check_references_test.go
+++ b/stats/internal/guard/check_references_test.go
@@ -105,6 +105,11 @@ func TestCheckReferences(t *testing.T) {
 				clean := crFixture(t, map[string]string{ntp: "## State file\n\nbody\n", demo: refState})
 				return crEnv(root, map[string]string{"FLOW_GUARD_REPO_ROOT": clean})
 			}},
+		// KAN-886: with no override and no shim-exported root, the guard
+		// names the missing variable rather than scanning an empty root.
+		{label: "no override and FLOW_GUARD_REPO_ROOT unset exits 2 naming it",
+			env: func(t *testing.T, root string) Env { return crEnv(root, nil) }, rc: 2,
+			has: []string{"check-references: FLOW_GUARD_REPO_ROOT is unset — run scripts/check-references.sh, which sets it"}},
 		{label: "in-code ** does not desync a live reference", files: map[string]string{ntp: "## State file\n\nbody\n",
 			demo: "See **State file** and code `x**y` in `rules/never-touch-production.mdc`.\n"}},
 		{label: "in-code ** does not mask a genuinely stale reference", files: map[string]string{ntp: "## Something else\n\nbody\n",
diff --git a/stats/internal/guard/check_unfinished_work_test.go b/stats/internal/guard/check_unfinished_work_test.go
index 100187be..55950329 100644
--- a/stats/internal/guard/check_unfinished_work_test.go
+++ b/stats/internal/guard/check_unfinished_work_test.go
@@ -594,6 +594,9 @@ func TestCheckUnfinishedWork(t *testing.T) {
 	for _, bad := range []struct{ label, json string }{
 		{"empty output", ""}, {"an object", `{}`}, {"an object of findings", `{"a":{"status":"open"}}`},
 		{"a null element", `[null]`}, {"two arrays", `[{"status":"open"}] []`},
+		// KAN-886: a null status decodes into "" with no error, which
+		// would be counted as one open finding instead of refused.
+		{"a finding with a null status", `[{"ref":"F1","status":null}]`},
 	} {
 		each(demo(func(t *testing.T, f *uwFix) { f.findings = []byte(bad.json) }),
 			uwRC("malformed findings refuse: "+bad.label+" exits 2", 2),
````

  - [ ] **Step 3: RED, then GREEN.** Remove each refusal in turn and run its guard's test, restoring it after:
    `|| state == nil` in `stats/internal/guard/panelreproducers.go`'s state decode → `TestCheckPanelReproducers` fails on `case 32`;
    `!pcIsString(f["status"]) || ` in `uwOpenFindings` (`stats/internal/guard/unfinishedwork.go`) → `TestCheckUnfinishedWork` fails on the three `a finding with a null status` subtests;
    `root == ""` replaced by `false` in the `FLOW_GUARD_REPO_ROOT` lookup of `stats/internal/guard/references.go`, then of `stats/internal/guard/installedcitations.go` → `TestCheckReferences` and `TestCheckInstalledCitations` each fail on their new subtest. With every refusal restored, all four pass.
    <!-- measured: the four mutations above, one at a time, in a scratch clone @ 373988df after the diff → each named subtest FAIL; restored → ok -->
  - [ ] **Step 4: Verify.** `cd stats && gofmt -w . && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0;
    `cd stats && go test ./internal/guard -run 'TestCheckPanelReproducers|TestCheckUnfinishedWork|TestCheckReferences|TestCheckInstalledCitations'` passes.
  - [ ] **Step 5: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/check_panel_reproducers_test.go`, `stats/internal/guard/check_unfinished_work_test.go`, `stats/internal/guard/check_references_test.go`, `stats/internal/guard/check_installed_citations_test.go`
**Tests:** `case 32: a null state record is cannot-answer, never clean`, `a finding with a null status`, `no override and FLOW_GUARD_REPO_ROOT unset exits 2 naming it`, `no override and FLOW_GUARD_REPO_ROOT unset: exit 2, stderr names it, stdout empty`
**Regression:** reverting the commit removes the four subtests, and each refusal can again be deleted with every test of its guard green.
**Commit:** test(guard): pin the four ported-guard refusals no test covered
**After:** none

**Decision:** pin-ported-refusals

- [ ] 2. Cover the commit map's refusals and its link-route satellite (KAN-889)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -c -F -e 'map entry is not' -e 'named twice in the commit map' stats/internal/guard/check_task_commit_fields_test.go`: it prints `0`.
    <!-- measured: the grep above in this worktree @ 373988df → 0 -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 373988df after Task 1's diff; gofmt -l internal/guard clean; go test ./internal/guard -run TestCheckTaskCommitFields passes; cases 160, 161 and 162 each fail with the mutation Step 3 names
diff --git a/stats/internal/guard/check_task_commit_fields_test.go b/stats/internal/guard/check_task_commit_fields_test.go
index 113adf4d..ebe3441f 100644
--- a/stats/internal/guard/check_task_commit_fields_test.go
+++ b/stats/internal/guard/check_task_commit_fields_test.go
@@ -3040,6 +3040,41 @@ func TestCheckTaskCommitFields(t *testing.T) {
 			res = commit("feat(alpha): add alpha for real", "feat(x-repo-change): add alpha for real")
 			ok(t, "case 158: a second commit's change-name scope fails", res.rc == 1 && has(res.out, "scope 'x-repo-change' names the change"), res)
 		}},
+		{"case 160", func(t *testing.T) {
+			// KAN-889: a map entry missing its sha is refused at the
+			// argument boundary, never read as a commit to judge.
+			b1, _, a := tcfMapRepos(t, "147")
+			res := fx.run(b1, "1", b1+"=", "", a, "x-repo-change")
+			ok(t, "case 160: a map entry with no sha refuses", res.rc == 2, res)
+			ok(t, "case 160: it names the malformed entry", has(res.out, "map entry is not <worktree>=<sha>: \""+b1+"=\""), res)
+		}},
+		{"case 161", func(t *testing.T) {
+			// KAN-889: one worktree named twice would judge its commit
+			// twice; the map refuses it.
+			b1, _, a := tcfMapRepos(t, "147")
+			sha1 := tcfGit(t, b1, "rev-parse", "HEAD")
+			res := fx.run(b1, "1", b1+"="+sha1+","+b1+"="+sha1, "", a, "x-repo-change")
+			ok(t, "case 161: a worktree named twice in the map refuses", res.rc == 2, res)
+			ok(t, "case 161: it names the repeated worktree", has(res.out, "worktree named twice in the commit map: "+b1), res)
+		}},
+		{"case 162", func(t *testing.T) {
+			// KAN-889: the first repository is a satellite whose change
+			// directory carries link.md, so the plan resolves through the
+			// link route to the canonical worktree, not the absent-dir one.
+			b1, b2, a := tcfMapRepos(t, "147")
+			writeFile(t, b1+"/spectre/changes/x-repo-change/link.md", "## Part of\n\n`peerx:x-repo-change`\n")
+			tcfGit(t, b1, "add", "spectre")
+			tcfGit(t, b1, "commit", "-q", "-m", "link")
+			writeFile(t, b1+"/alpha.txt", "def test_alpha(): pass\n")
+			tcfGit(t, b1, "add", "alpha.txt")
+			tcfGit(t, b1, "commit", "-q", "-m", "add alpha for real")
+			writeFile(t, b2+"/beta.txt", "b\n")
+			tcfGit(t, b2, "add", "beta.txt")
+			tcfGit(t, b2, "commit", "-q", "-m", "add alpha for real")
+			sha1, sha2 := tcfGit(t, b1, "rev-parse", "HEAD"), tcfGit(t, b2, "rev-parse", "HEAD")
+			res := fx.run(b1, "1", b1+"="+sha1+","+b2+"="+sha2, "", a, "x-repo-change")
+			ok(t, "case 162: a link-route satellite resolves the canonical plan under the commit map", res.rc == 0, res)
+		}},
 	}
 	for _, c := range cases {
 		t.Run(c.name, func(t *testing.T) {
````

  - [ ] **Step 3: RED, then GREEN.** Apply each mutation in turn, restoring it after, and run
    `cd stats && go test ./internal/guard -run 'TestCheckTaskCommitFields/case_16'`:
    `wt == "" || sha == "" || ` removed from the map parse in `stats/internal/guard/taskcommitfields.go` → `case 160` fails;
    the `slices.ContainsFunc(repos, …)` condition replaced by `false` → `case 161` fails;
    in `changePlanDir`'s link branch (`stats/internal/guard/changeplan.go`) the `if isFile(canonDir + "/tasks.md") { return canonDir, 0 }` block removed → `case 162` fails. Restored, all pass.
    <!-- measured: the three mutations above, one at a time, in a scratch clone @ 373988df after the diff → case 160, case 161, case 162 FAIL respectively; restored → ok -->
  - [ ] **Step 4: Verify.** `cd stats && gofmt -w . && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0;
    `cd stats && go test ./internal/guard -run 'TestCheckTaskCommitFields|TestTaskFieldParseMatchesPython'` passes.
  - [ ] **Step 5: Commit** the file with the subject below.

**Build:** green
**Files:** `stats/internal/guard/check_task_commit_fields_test.go`
**Tests:** `case 160: a map entry with no sha refuses`, `case 161: a worktree named twice in the map refuses`, `case 162: a link-route satellite resolves the canonical plan under the commit map`
**Regression:** reverting the commit removes cases 160–162, and the malformed-entry and duplicate-worktree refusals and the map's link route can again be removed with the guard's tests green.
**Baseline:** before=139 after=142 <!-- measured: grep -c -E '^[[:space:]]+\{"case [0-9]+", func' stats/internal/guard/check_task_commit_fields_test.go @ 373988df → 139, and 142 after the diff in a scratch clone -->
**Commit:** test(guard): cover the commit map's refusals and link-route satellite
**After:** none

**Decision:** commit-map-cases

- [ ] 3. Read another token's original as a resumed fix round's own (KAN-890)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n 'foreign' stats/internal/guard/panelfixsingledispatch.go; echo $?`: it prints only `1` — the guard keeps no other token's row.
    <!-- measured: the grep above in this worktree @ 373988df → exit 1, no line -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 373988df after Tasks 1-2's diffs; gofmt -l and go vet ./internal/guard clean; go test ./internal/guard -run TestCheckPanelFixSingleDispatch passes; cases 18 and 19 fail with the guard file at 373988df and with the mutation Step 3 names
diff --git a/scripts/check-panel-fix-single-dispatch.sh b/scripts/check-panel-fix-single-dispatch.sh
index 08e16038..2461c4fa 100755
--- a/scripts/check-panel-fix-single-dispatch.sh
+++ b/scripts/check-panel-fix-single-dispatch.sh
@@ -68,6 +68,12 @@
 # tasks.md that cannot be read while such a row exists is exit 2. A row
 # that recorded no commit is not judged.
 #
+# A FIX ROUND RESUMED UNDER A FRESH SESSION TOKEN keeps its original
+# (KAN-890): an original another session token of the change recorded under
+# a key answers whether that key's original exists -- for a `-retry`, for a
+# chunked round's bare key and for its chunk contiguity -- and nothing
+# else; every count above stays this token's.
+#
 # Rows of other roles and rows of other session tokens never count, nor do
 # `full-suite-fix-<n>` rows recorded with agent id `inline`: implement.md
 # records the parent's own full-suite fix rounds under role `panel-fix`, and
diff --git a/stats/internal/guard/check_panel_fix_single_dispatch_test.go b/stats/internal/guard/check_panel_fix_single_dispatch_test.go
index 54b36e90..5e1cdd72 100644
--- a/stats/internal/guard/check_panel_fix_single_dispatch_test.go
+++ b/stats/internal/guard/check_panel_fix_single_dispatch_test.go
@@ -146,6 +146,13 @@ func TestCheckPanelFixSingleDispatch(t *testing.T) {
 			pfdCase{dispatches: d("panel-fix-1", pf, "mf-tok", "panel-fix-1-f2", pf, "mf-tok"), findings: pfdFindings(24, 0)}, 1, "panel-fix-1-f2"},
 		{"case 17: findings-store-unreachable-exits-2",
 			pfdCase{dispatches: []byte("[]"), fErr: errors.New("exit status 1")}, 2, "flow record findings failed for 'demo' -- cannot answer"},
+		// KAN-890: a fix round resumed under a fresh session token.
+		{"case 18: a retry whose original another token of the change recorded exits 0",
+			pfdCase{dispatches: d("panel-fix-1", pf, "mf-old", "panel-fix-1-retry", pf, "mf-tok")}, 0, ""},
+		{"case 19: a resumed chunked round whose bare key and first chunk another token recorded exits 0",
+			pfdCase{dispatches: d("panel-fix-1", pf, "mf-old", "panel-fix-1-2", pf, "mf-old", "panel-fix-1-3", pf, "mf-tok"), findings: pfdFindings(24, 0)}, 0, ""},
+		{"case 20: another token's retry is no original for this token's retry",
+			pfdCase{dispatches: d("panel-fix-1-retry", pf, "mf-old", "panel-fix-1-retry", pf, "mf-tok")}, 1, "round panel-fix-1 carries 1 -retry dispatch(es) and no original"},
 	}
 	for _, tc := range cases {
 		t.Run(tc.label, func(t *testing.T) {
diff --git a/stats/internal/guard/panelfixsingledispatch.go b/stats/internal/guard/panelfixsingledispatch.go
index b2101b71..e7866a3f 100644
--- a/stats/internal/guard/panelfixsingledispatch.go
+++ b/stats/internal/guard/panelfixsingledispatch.go
@@ -46,11 +46,13 @@ var pfdFullSuiteRE = regexp.MustCompile(`^full-suite-fix-[0-9]+$`)
 
 // pfdBase is one panel-fix base key (the key with a trailing -retry
 // stripped) and its counts, kept in first-seen order as the bash's parallel
-// arrays were.
+// arrays were. orig and retries count this session token's rows; foreign
+// counts the originals other session tokens of the change recorded under the
+// same key, which answer only whether the key's original exists (KAN-890).
 type pfdBase struct {
-	key           string
-	round, chunk  int64
-	orig, retries int
+	key                    string
+	round, chunk           int64
+	orig, retries, foreign int
 }
 
 func checkPanelFixSingleDispatch(args []string, env Env, stdout, stderr io.Writer) int {
@@ -124,9 +126,10 @@ func checkPanelFixSingleDispatch(args []string, env Env, stdout, stderr io.Write
 		return 2
 	}
 
-	// Each panel-fix row of THIS token. A payload that is not the shape the
-	// verb documents is the same cannot-answer posture as a failed call.
-	fixRows, ok := pfdPanelFixRows(rows, token)
+	// Each panel-fix row of THIS token, and apart from them every other
+	// token's. A payload that is not the shape the verb documents is the same
+	// cannot-answer posture as a failed call.
+	fixRows, foreignRows, ok := pfdPanelFixRows(rows, token)
 	if !ok {
 		fmt.Fprintf(stderr, "%sdispatch rows were not readable JSON -- cannot answer\n", pfdPrefix)
 		return 2
@@ -140,13 +143,7 @@ func checkPanelFixSingleDispatch(args []string, env Env, stdout, stderr io.Write
 	var violations []string
 	var bases []*pfdBase
 	index := map[string]*pfdBase{}
-	for _, row := range fixRows {
-		key := row.key
-		if !pfdKeyRE.MatchString(key) {
-			violations = append(violations, fmt.Sprintf("out-of-shape panel-fix key '%s' -- the canonical shape is panel-fix-<round>[-<chunk>][-retry]", key))
-			continue
-		}
-		base, retry := strings.CutSuffix(key, "-retry")
+	baseOf := func(base string) *pfdBase {
 		b := index[base]
 		if b == nil {
 			// rest is "<round>" or "<round>-<chunk>"; parsed base 10 (the
@@ -157,17 +154,42 @@ func checkPanelFixSingleDispatch(args []string, env Env, stdout, stderr io.Write
 			index[base] = b
 			bases = append(bases, b)
 		}
+		return b
+	}
+	for _, row := range fixRows {
+		key := row.key
+		if !pfdKeyRE.MatchString(key) {
+			violations = append(violations, fmt.Sprintf("out-of-shape panel-fix key '%s' -- the canonical shape is panel-fix-<round>[-<chunk>][-retry]", key))
+			continue
+		}
+		base, retry := strings.CutSuffix(key, "-retry")
+		b := baseOf(base)
 		if retry {
 			b.retries++
 		} else {
 			b.orig++
 		}
 	}
+	// A FIX ROUND RESUMED UNDER A FRESH SESSION TOKEN (KAN-890): kan-789's
+	// resumed round recorded panel-fix-1-retry under its new token while
+	// panel-fix-1 sat under the token that started the round, and the
+	// retry-alone rule read a retry with no original. Another token's
+	// original of a key therefore answers whether that key's original
+	// exists -- for a retry, a chunked round's bare key and its chunk
+	// contiguity -- and nothing else: the dispatch counts stay this token's,
+	// so two runs that each dispatched a round once are never one round
+	// dispatched twice, and a round no row of this token touched is never
+	// judged.
+	for _, row := range foreignRows {
+		if pfdKeyRE.MatchString(row.key) && !strings.HasSuffix(row.key, "-retry") {
+			baseOf(row.key).foreign++
+		}
+	}
 
 	for _, b := range bases {
 		total := b.orig + b.retries
 		switch {
-		case b.retries > 0 && b.orig == 0:
+		case b.retries > 0 && b.orig == 0 && b.foreign == 0:
 			violations = append(violations, fmt.Sprintf("round %s carries %d -retry dispatch(es) and no original -- a retry is never a round's only dispatch", b.key, b.retries))
 		case b.orig > 1:
 			violations = append(violations, fmt.Sprintf("round %s carries %d panel-fix dispatches -- the fix goes to ONE subagent as the combined list, never one per reviewer, slot or finding", b.key, b.orig))
@@ -184,7 +206,7 @@ func checkPanelFixSingleDispatch(args []string, env Env, stdout, stderr io.Write
 	var rounds []int64
 	seen := map[int64]bool{}
 	for _, b := range bases {
-		if !seen[b.round] {
+		if b.orig+b.retries > 0 && !seen[b.round] {
 			seen[b.round] = true
 			rounds = append(rounds, b.round)
 		}
@@ -204,7 +226,7 @@ func checkPanelFixSingleDispatch(args []string, env Env, stdout, stderr io.Write
 		if len(chunks) == 0 {
 			continue
 		}
-		if bare == nil || bare.orig == 0 {
+		if bare == nil || bare.orig+bare.foreign == 0 {
 			violations = append(violations, fmt.Sprintf("round panel-fix-%d carries chunked dispatches but no panel-fix-%d original -- chunk numbering starts at the round's own key", r, r))
 		}
 
@@ -213,7 +235,7 @@ func checkPanelFixSingleDispatch(args []string, env Env, stdout, stderr io.Write
 		// proves nothing about numbering.
 		var present []int64
 		for _, b := range chunks {
-			if b.orig >= 1 {
+			if b.orig+b.foreign >= 1 {
 				present = append(present, b.chunk)
 			}
 		}
@@ -300,20 +322,20 @@ type pfdRow struct{ key, commit string }
 
 // pfdPanelFixRows is `.[] | select((.sessionToken // "") == $tok and
 // (.role // "") == "panel-fix")` then each row's `.key // ""` as jq -r
-// printed it, beside the row's commitSha. A null row selects nothing; any other non-object row, or a
+// printed it, beside the row's commitSha; the panel-fix rows of every other
+// session token come back apart, as foreign. A null row selects nothing; any other non-object row, or a
 // top-level value that is not an array, is jq failing. (jq iterated a
 // top-level object's values in insertion order; that shape is refused here
 // rather than guessed at -- the verb prints an array.)
-func pfdPanelFixRows(raw []byte, token string) ([]pfdRow, bool) {
+func pfdPanelFixRows(raw []byte, token string) (keys, foreign []pfdRow, ok bool) {
 	vals, ok := pfdStream(raw)
 	if !ok {
-		return nil, false
+		return nil, nil, false
 	}
-	var keys []pfdRow
 	for _, v := range vals {
 		arr, ok := v.([]any)
 		if !ok {
-			return nil, false
+			return nil, nil, false
 		}
 		for _, e := range arr {
 			if e == nil {
@@ -321,9 +343,9 @@ func pfdPanelFixRows(raw []byte, token string) ([]pfdRow, bool) {
 			}
 			row, ok := e.(map[string]any)
 			if !ok {
-				return nil, false
+				return nil, nil, false
 			}
-			if row["sessionToken"] != token || row["role"] != "panel-fix" {
+			if row["role"] != "panel-fix" {
 				continue
 			}
 			key := pfdRawText(row["key"])
@@ -334,11 +356,15 @@ func pfdPanelFixRows(raw []byte, token string) ([]pfdRow, bool) {
 			if row["agentId"] == "inline" && pfdFullSuiteRE.MatchString(key) {
 				continue
 			}
+			if row["sessionToken"] != token {
+				foreign = append(foreign, pfdRow{key: key})
+				continue
+			}
 			commit, _ := row["commitSha"].(string)
 			keys = append(keys, pfdRow{key: key, commit: commit})
 		}
 	}
-	return keys, true
+	return keys, foreign, true
 }
 
 // pfdCheckCommits names each row whose recorded commit carries neither a
````

  - [ ] **Step 3: RED, then GREEN.** In `stats/internal/guard/panelfixsingledispatch.go` replace `foreign = append(foreign, pfdRow{key: key})` with `_ = foreign`:
    `cd stats && go test ./internal/guard -run TestCheckPanelFixSingleDispatch` fails on `case 18` and `case 19`. Restore the line: it passes.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -w . && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0;
    `cd stats && go test ./internal/guard -run TestCheckPanelFixSingleDispatch` passes; `scripts/check-guard-symlinks.sh` and `scripts/check-references.sh` exit 0.
  - [ ] **Step 5: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/panelfixsingledispatch.go`, `stats/internal/guard/check_panel_fix_single_dispatch_test.go`, `scripts/check-panel-fix-single-dispatch.sh`
**Tests:** `case 18: a retry whose original another token of the change recorded exits 0`, `case 19: a resumed chunked round whose bare key and first chunk another token recorded exits 0`, `case 20: another token's retry is no original for this token's retry`
**Regression:** reverting the commit drops every other token's row again — cases 18 and 19 fail, 18 on the retry-alone violation and 19 on the chunked round's missing bare key; case 20 passes either way, pinning that a foreign retry never counts.
**Baseline:** before=24 after=27 <!-- measured: grep -c -E '^[[:space:]]+\{"case ' stats/internal/guard/check_panel_fix_single_dispatch_test.go @ 373988df → 24, and 27 after the diff in a scratch clone -->
**Commit:** fix(guard): read another token's original as a resumed fix round's own
**After:** none

**Decision:** foreign-token-original

- [ ] 4. Join a list item's wrapped lines into its verbatim-moves block (KAN-896)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** With `<scratch>` a scratch directory outside the worktree, run
    `git worktree add -q --detach <scratch>/k813 f603fe9b && (cd stats && go build -o <scratch>/flow-guard ./cmd/flow-guard) && FLOW_GUARD_REPO_ROOT=<scratch>/k813 <scratch>/flow-guard check-verbatim-moves d4b55fa8^ | grep -c '^FAIL deleted or reworded'; git worktree remove --force <scratch>/k813`:
    it prints `2` — the two halves of kan-813's untouched `Run **Verify**` sentence, `f603fe9b` being kan-813's tree once its one genuine reword was restored.
    <!-- measured: the commands above, under the session scratchpad, in a scratch clone @ 373988df → 2; after the diff → 0 -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 373988df after Tasks 1-3's diffs; gofmt -l internal/guard clean; go test ./internal/guard -run 'TestVMSentences|TestCheckVerbatimMoves' passes; both new tests fail with the guard file at 373988df and with the mutation Step 3 names; scripts/check-verbatim-moves.sh 373988df reports 0 violations over the tree
diff --git a/stats/internal/guard/check_verbatim_moves_test.go b/stats/internal/guard/check_verbatim_moves_test.go
index f326eb71..b0303fd1 100644
--- a/stats/internal/guard/check_verbatim_moves_test.go
+++ b/stats/internal/guard/check_verbatim_moves_test.go
@@ -60,6 +60,7 @@ const (
 	// A new run-loaded sentence a flow-fast change acknowledges in its own
 	// change root instead of spectre's.
 	vmAdded = "The confirming re-run verifies a fresh build, never a stale stack."
+	vmLoop  = "skills/flow/verify-fix-loop.md"
 )
 
 var vmBase = map[string]string{
@@ -154,6 +155,15 @@ func TestCheckVerbatimMoves(t *testing.T) {
 				vmPipe: vmBase[vmPipe] + "\n" + vmAdded + "\n",
 			},
 			rc: 1, want: []string{"FAIL new run-loaded text (paraphrase?) :: " + vmAdded}},
+		// KAN-896: kan-813's shape — a sentence inserted ahead of one that
+		// wraps across a list item's first line, the item rewrapped.
+		{name: "a pure insertion into a wrapped list item flags only the inserted sentence",
+			base: map[string]string{vmLoop: "# Loop\n\n4. **Re-capture and re-run.** Run **Verify** again from its\n   `begin` mark, then more follows here.\n"},
+			edit: map[string]string{
+				vmLoop: "# Loop\n\n4. **Re-capture and re-run.** The inserted sentence leads the step now. Run\n   **Verify** again from its `begin` mark, then more follows here.\n",
+				"spectre/changes/demo/verbatim-moves.txt": "The inserted sentence leads the step now.\n",
+			},
+			rc: 0, want: []string{"ok   acknowledged new text :: The inserted sentence leads the step now.", "0 violation(s)"}, reject: []string{"FAIL"}},
 		{name: "a deleted file fails every sentence it held",
 			remove: []string{vmPipe},
 			rc:     1, want: []string{"FAIL deleted or reworded :: " + vmWhy}},
@@ -346,3 +356,26 @@ func TestVMSentences(t *testing.T) {
 		t.Fatalf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
 	}
 }
+
+// TestVMSentencesJoinListItemContinuation pins KAN-896's block rule: a list
+// item's indented continuation lines join it, quoted or not, so a sentence
+// wrapped across the item's first line is one sentence; an unindented line
+// after an item still opens a block of its own.
+func TestVMSentencesJoinListItemContinuation(t *testing.T) {
+	t.Parallel()
+	in := "4. **A longer lead-in.** A sentence that wraps across\n   the item's first line. A second one follows.\n" +
+		"> - a quoted item whose sentence\n>   wraps onto the next line.\n" +
+		"- an item that stands alone\nAn unindented line opens its own block.\n"
+	got := vmSentences(in)
+	want := []string{
+		"**A longer lead-in.",
+		"A sentence that wraps across the item's first line.",
+		"A second one follows.",
+		"a quoted item whose sentence wraps onto the next line.",
+		"an item that stands alone",
+		"An unindented line opens its own block.",
+	}
+	if strings.Join(got, "\n") != strings.Join(want, "\n") {
+		t.Fatalf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
+	}
+}
diff --git a/stats/internal/guard/verbatimmoves.go b/stats/internal/guard/verbatimmoves.go
index f4bdec27..12afb755 100644
--- a/stats/internal/guard/verbatimmoves.go
+++ b/stats/internal/guard/verbatimmoves.go
@@ -371,8 +371,13 @@ func vmAcks(root string, archived []string) (map[string]bool, error) {
 }
 
 // vmBlocks is the prototype's blocks(): paragraphs joined across soft wraps;
-// a heading, table row, rule or list item stands alone; a fenced line stands
-// alone; blockquote markers are dropped.
+// a heading, table row or rule stands alone; a fenced line stands alone;
+// blockquote markers are dropped. A list item is the one departure from the
+// prototype (KAN-896): it opens a block its indented continuation lines join,
+// where the prototype stood its first line alone — that cut a sentence
+// wrapped across the item's first line in two, so text inserted ahead of it
+// on that line moved the cut and the untouched sentence read as reworded
+// (kan-813). An unindented line after an item opens a block of its own.
 func vmBlocks(text string) []string {
 	var out, cur []string
 	flush := func() {
@@ -381,8 +386,15 @@ func vmBlocks(text string) []string {
 			cur = nil
 		}
 	}
-	fence := false
+	fence, item := false, false
 	for _, l := range strings.Split(text, "\n") {
+		// Indentation is read past any blockquote markers, so a quoted
+		// item's continuation (`>   more`) joins it as a plain one does.
+		t := strings.TrimLeft(l, " \t")
+		indented := t != l
+		if q := vmQuote.FindString(t); q != "" {
+			indented = strings.TrimLeft(t[len(q):], " \t") != t[len(q):]
+		}
 		s := strings.TrimSpace(l)
 		if strings.HasPrefix(s, "```") || strings.HasPrefix(s, "~~~") {
 			flush()
@@ -396,11 +408,19 @@ func vmBlocks(text string) []string {
 		s = vmQuote.ReplaceAllString(s, "")
 		if s == "" || vmBreak.MatchString(s) {
 			flush()
-			if s != "" {
-				out = append(out, vmListMarker.ReplaceAllString(s, ""))
+			item = vmListMarker.MatchString(s)
+			switch {
+			case item:
+				cur = append(cur, vmListMarker.ReplaceAllString(s, ""))
+			case s != "":
+				out = append(out, s)
 			}
 			continue
 		}
+		if item && !indented {
+			flush()
+		}
+		item = item && indented
 		cur = append(cur, s)
 	}
 	flush()
````

  - [ ] **Step 3: RED, then GREEN.** In `vmBlocks` (`stats/internal/guard/verbatimmoves.go`) replace `cur = append(cur, vmListMarker.ReplaceAllString(s, ""))` with `out = append(out, vmListMarker.ReplaceAllString(s, ""))`:
    `cd stats && go test ./internal/guard -run 'TestVMSentences|TestCheckVerbatimMoves'` fails on `TestVMSentencesJoinListItemContinuation` and `a pure insertion into a wrapped list item flags only the inserted sentence`. Restore the line: it passes.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -w . && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0;
    `cd stats && go test ./internal/guard -run 'TestVMSentences|TestCheckVerbatimMoves'` passes; Step 1's count prints `0`.
  - [ ] **Step 5: Commit** the two files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/verbatimmoves.go`, `stats/internal/guard/check_verbatim_moves_test.go`
**Tests:** `TestVMSentencesJoinListItemContinuation`, `a pure insertion into a wrapped list item flags only the inserted sentence`
**Regression:** reverting the commit stands a list item's first line alone again — the first test sees the wrapped sentence cut at the line break, and the second fails on two "deleted or reworded" lines for the untouched sentence.
**Baseline:** before=24 after=26 <!-- measured: grep -c -E '^func Test[A-Za-z0-9_]+\(|^[[:space:]]+\{name: "' stats/internal/guard/check_verbatim_moves_test.go @ 373988df → 24, and 26 after the diff in a scratch clone -->
**Commit:** fix(guard): join a list item's wrapped lines in verbatim-moves blocks
**After:** none

**Decision:** list-item-continuation-block

- [ ] 5. State the mapped pair on the reviewer row's record paragraph (KAN-898)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -c -F 'KAN-898' skills/flow/review-panel.md`: it prints `0`.
    <!-- measured: the grep above in this worktree @ 373988df → 0 -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 373988df after Tasks 1-4's diffs; check-references.sh, check-installed-citations.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh, check-normative-inventory.sh and check-ask-silence.sh exit 0, and check-verbatim-moves.sh 373988df fails only on the sentence verbatim-moves.txt lists
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index 3e564864..daa66737 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -449,6 +449,9 @@ a `default` panel and the dispatch's own model from the decision's
 fix-round re-run on a decided panel, `panel.rerun_dispatch`'s model. `-effort` likewise:
 `low` on a `default` panel, and the dispatch's own effort on a decided panel,
 `panel.rerun_dispatch`'s `low` on a re-run. The always-on extra dispatch is the one exception to both: its pass-1 `-model` and `-effort` are its `extra_dispatch` pair, `opus`/`low` when the decision has none.
+On harness `zcode` every pair this paragraph names is recorded as `glm-5.3-flash` / `high`
+instead — the one pair the store accepts on a `zcode` session token — and on harness `muse` as
+the session's own (**Harness mapping**, `skills/flow-contracts/model-policy.md`; KAN-898).
 
 **`-agent-id` is never typed on a slot's row, never invented** — the daemon captures the launch identifier,
 pairing each launch with the begin nearest it in time;
````

  - [ ] **Step 3: Verify.** `grep -c -F 'KAN-898' skills/flow/review-panel.md` prints `1`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh` and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ 373988df → 1 -->
  - [ ] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — no test is declared.
**Commit:** docs(flow): state the mapped pair on the reviewer row's record paragraph
**After:** none

**Decision:** reviewer-row-mapped-pair
