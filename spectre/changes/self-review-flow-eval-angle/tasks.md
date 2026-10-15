# self-review-flow-eval-angle

> **Execution:** `/flow-self-review`'s fix loop implements this plan. Mark a task's own checkbox
> when `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter, with any `measured:` command naming no ref of its own.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`.
- Each task's change is the unified diff its step gives, applied from the repository root with `git apply` — the three diffs were applied in plan order to a scratch clone at `5a1601fa` and reproduce the planner's working tree. A diff that does not apply is reported, never hand-merged by guesswork.
- Every prose line a diff adds is copied verbatim: this change's `verbatim-moves.txt` acknowledges each new or reworded sentence by its exact text, and any other wording fails `scripts/check-verbatim-moves.sh`.
- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w <package dir>` before `gofmt -l`. `go build ./...` needs the built SPA (`internal/web/dist`, `make web-build`); the commands below vet and test only the package each task touches.
- Tasks 1 and 3 both edit `skills/flow-contracts/finish-contract-run1.md` and `skills/flow-fast/SKILL.md`, and Tasks 2 and 3 both edit `skills/flow-self-review/SKILL.md`: run them in order, one after another.

- [ ] 1. The self-review bundle serves the archived `visual-verification.md`
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Apply the `stats/internal/selfreview/bundle_test.go` hunks of the diff in step 2 alone (`git apply --include='stats/internal/selfreview/bundle_test.go' <the diff>`) and run `cd stats && go test ./internal/selfreview/ -run TestBundleServesArchivedVisualVerification -count=1`: it FAILS (`bundle missing "## spectre/changes/archive/demo/visual-verification.md"`) — the defect: a visual candidate's record never reaches the self-review pass.
  <!-- measured: the step-1 command with only the bundle_test.go hunks applied, in a scratch clone @ 5a1601fa (exit 1, --- FAIL: TestBundleServesArchivedVisualVerification) -->
  - [ ] **Step 2: Implement.** Apply the rest of this diff (`git apply --exclude='stats/internal/selfreview/bundle_test.go' <the diff>`):

````diff verified:applied with git apply in plan order to a scratch clone at 5a1601fa, reproducing the planner's working tree byte for byte; gofmt -l and go vet clean; go test ./internal/guard/ ./internal/selfreview/ and scripts/run-guard-tests.sh pass with all three tasks applied
diff --git a/skills/flow-contracts/finish-contract-run1.md b/skills/flow-contracts/finish-contract-run1.md
index 66c55fe0..a4f15cc8 100644
--- a/skills/flow-contracts/finish-contract-run1.md
+++ b/skills/flow-contracts/finish-contract-run1.md
@@ -291,7 +291,7 @@ bundle with
 `flow self-review bundle -change <name>`: flowd assembles the whole bundle — the ledger and the
 panel record rendered from the store, or, when the store yields no render for that record, read
 from the copies the archive commit carries, the archived `tasks.md`,
-`design.md` and `narrative.md`
+`design.md`, `narrative.md` and `visual-verification.md`
 read out of `spectre/<name>` in the repository the command resolves from its
 own location (the main checkout its working directory sits in — the store carries no repository
 roots for the pipeline's changes), and the `git log --stat` of every implementation, planning and
diff --git a/skills/flow-fast/SKILL.md b/skills/flow-fast/SKILL.md
index 3c603866..d454892d 100644
--- a/skills/flow-fast/SKILL.md
+++ b/skills/flow-fast/SKILL.md
@@ -286,8 +286,8 @@ a fix re-run never records its findings twice.
 Assemble the bundle in memory, never written to `<project>/docs/self-review/` or committed, as, in
 order: the stdout of `flow self-review
 bundle -change <name>` run from the worktree — of the sources `/flow`'s bundle judges a change
-by, a `/flow-fast` run can never have tasks.md, design.md and narrative.md (no spectre
-artifacts and no archive), so those three report `skipped` by design, never silently;
+by, a `/flow-fast` run can never have tasks.md, design.md, narrative.md and visual-verification.md (no spectre
+artifacts and no archive), so those four report `skipped` by design, never silently;
 the git-log source resolves from the main checkout the command resolves, as the change branch's
 commits; the ledger and panel render from the store when the run wrote rows — then
 `## Session narrative`, one paragraph this session writes on what it
diff --git a/stats/internal/selfreview/bundle_test.go b/stats/internal/selfreview/bundle_test.go
index d3945d27..82f8c522 100644
--- a/stats/internal/selfreview/bundle_test.go
+++ b/stats/internal/selfreview/bundle_test.go
@@ -46,8 +46,9 @@ func TestBundleAssemblyRendersStoreSources(t *testing.T) {
 			"/repo show spectre/demo:spectre/changes/archive/demo/tasks.md": "# demo tasks\n",
 		},
 		failOn: map[string]bool{
-			"/repo show spectre/demo:spectre/changes/archive/demo/design.md":    true,
-			"/repo show spectre/demo:spectre/changes/archive/demo/narrative.md": true,
+			"/repo show spectre/demo:spectre/changes/archive/demo/design.md":              true,
+			"/repo show spectre/demo:spectre/changes/archive/demo/narrative.md":           true,
+			"/repo show spectre/demo:spectre/changes/archive/demo/visual-verification.md": true,
 		},
 	}
 
@@ -73,6 +74,7 @@ func TestBundleAssemblyRendersStoreSources(t *testing.T) {
 	for _, label := range []string{
 		"spectre/changes/archive/demo/design.md",
 		"spectre/changes/archive/demo/narrative.md",
+		"spectre/changes/archive/demo/visual-verification.md",
 		"git log --stat",
 	} {
 		if !strings.Contains(bundle, "skipped: "+label+" (absent)") {
@@ -96,7 +98,7 @@ func TestBundleAssemblySkipsAbsentSources(t *testing.T) {
 
 	// A change the store has never heard of renders nothing: the ledger
 	// needs dispatch rows and the panel needs any row at all.
-	if !strings.Contains(bundle, "found: 0 of 7 sources; skipped: 7 of 7 sources") {
+	if !strings.Contains(bundle, "found: 0 of 8 sources; skipped: 8 of 8 sources") {
 		t.Errorf("summary line wrong:\n%s", bundle)
 	}
 	if strings.Contains(bundle, "## .superpowers/sdd/ledgers/demo.md") {
@@ -186,6 +188,35 @@ func TestBundleAssemblyReadsArchivedFilesThroughGit(t *testing.T) {
 	}
 }
 
+// TestBundleServesArchivedVisualVerification pins the flow-eval angle's
+// visual source: an archived visual-verification.md is served as its own
+// section, the record a visual candidate's screenshots and findings are
+// read from.
+func TestBundleServesArchivedVisualVerification(t *testing.T) {
+	repo := gitRepo(t)
+	writeArchiveBranch(t, repo, "demo", map[string]string{
+		"tasks.md":               "# demo tasks\n",
+		"visual-verification.md": "# Visual verification — demo\n",
+	})
+
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, ExecRunner{})
+	if err != nil {
+		t.Fatalf("Bundle: %v", err)
+	}
+
+	for _, want := range []string{
+		"## spectre/changes/archive/demo/visual-verification.md",
+		"# Visual verification — demo",
+	} {
+		if !strings.Contains(bundle, want) {
+			t.Errorf("bundle missing %q:\n%s", want, bundle)
+		}
+	}
+	if strings.Contains(bundle, "skipped: spectre/changes/archive/demo/visual-verification.md") {
+		t.Errorf("visual-verification.md archived, yet reported skipped:\n%s", bundle)
+	}
+}
+
 func TestBundleAssemblyDerivesFinishCommits(t *testing.T) {
 	repo := gitRepo(t)
 	// The archive commit sits on the change branch above the implementation
@@ -245,7 +276,7 @@ func TestBundleAssemblyGitLogFallsBackToChangeBranch(t *testing.T) {
 	}
 	// One found source — the git log; every other source a flow-fast run
 	// has no archive to read and no store rows to render.
-	if !strings.Contains(bundle, "found: 1 of 7 sources; skipped: 6 of 7 sources") {
+	if !strings.Contains(bundle, "found: 1 of 8 sources; skipped: 7 of 8 sources") {
 		t.Errorf("summary line wrong:\n%s", bundle)
 	}
 }
@@ -321,10 +352,12 @@ func TestBundleAssemblyFallsBackToCommittedRecords(t *testing.T) {
 		"## spectre/changes/archive/demo/panel.md",
 		"# Review panel — demo",
 		// tasks.md and the archive commit's git log resolve besides the two
-		// fallback sections; design.md and narrative.md stay skipped.
-		"found: 4 of 7 sources",
+		// fallback sections; design.md, narrative.md and
+		// visual-verification.md stay skipped.
+		"found: 4 of 8 sources",
 		"skipped: spectre/changes/archive/demo/design.md (absent)",
 		"skipped: spectre/changes/archive/demo/narrative.md (absent)",
+		"skipped: spectre/changes/archive/demo/visual-verification.md (absent)",
 	} {
 		if !strings.Contains(bundle, want) {
 			t.Errorf("bundle missing %q:\n%s", want, bundle)
@@ -728,7 +761,7 @@ func TestBundleServesRecordedSummaryFirst(t *testing.T) {
 	for _, want := range []string{
 		"## change summary",
 		"what changed and why",
-		"found: 1 of 7 sources",
+		"found: 1 of 8 sources",
 	} {
 		if !strings.Contains(bundle, want) {
 			t.Errorf("bundle missing %q:\n%s", want, bundle)
@@ -738,7 +771,7 @@ func TestBundleServesRecordedSummaryFirst(t *testing.T) {
 
 // TestBundleSkipsAbsentSummary asserts the missing-source shape: a change
 // with no recorded summary reports it skipped like any other absent
-// source, in the same words, and the found line counts the seven sources.
+// source, in the same words, and the found line counts the eight sources.
 func TestBundleSkipsAbsentSummary(t *testing.T) {
 	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}
 
@@ -747,7 +780,7 @@ func TestBundleSkipsAbsentSummary(t *testing.T) {
 		t.Fatalf("Bundle: %v", err)
 	}
 
-	if !strings.Contains(bundle, "found: 0 of 7 sources; skipped: 7 of 7 sources") {
+	if !strings.Contains(bundle, "found: 0 of 8 sources; skipped: 8 of 8 sources") {
 		t.Errorf("summary line wrong:\n%s", bundle)
 	}
 	if !strings.Contains(bundle, "skipped: change summary (absent)") {
@@ -774,7 +807,7 @@ func TestBundleNamesRecordsLossWhenPanelRanWithoutDispatches(t *testing.T) {
 	}
 	// The note is a note, not a source: the found/skipped line still
 	// counts exactly the seven real sources.
-	if !strings.Contains(bundle, "found: 0 of 7 sources; skipped: 7 of 7 sources") {
+	if !strings.Contains(bundle, "found: 0 of 8 sources; skipped: 8 of 8 sources") {
 		t.Errorf("found/skipped line disturbed by the note:\n%s", bundle)
 	}
 	if strings.Contains(bundle, "skipped: note:") {
diff --git a/stats/internal/selfreview/git.go b/stats/internal/selfreview/git.go
index 7f7f6958..1d164147 100644
--- a/stats/internal/selfreview/git.go
+++ b/stats/internal/selfreview/git.go
@@ -446,6 +446,13 @@ func trimmedOutput(b []byte, err error) string {
 	return strings.TrimSpace(string(b))
 }
 
+// archivedFiles are the change's archived files the bundle serves, in
+// order. visual-verification.md is the record the flow-eval angle's visual
+// candidates are read from (skills/flow-self-review/SKILL.md, angle 7); a
+// change with no visual verification reports it skipped like any absent
+// source.
+var archivedFiles = []string{"tasks.md", "design.md", "narrative.md", "visual-verification.md"}
+
 // Bundle assembles the whole self-review context bundle for change as one
 // Markdown document: the header, the found/skipped summary line, one
 // `skipped: <label> (absent)` line per absent source, and one `## <label>`
@@ -475,7 +482,7 @@ func trimmedOutput(b []byte, err error) string {
 // answer for. A change no supplied repository archives — a /flow-fast
 // change, never archived by design — still gets its git-log source, read
 // off the change branch in the first readable supplied repository, the
-// main checkout the caller resolved (KAN-583); the three spectre sources
+// main checkout the caller resolved (KAN-583); the archivedFiles sources
 // stay skipped for it. The git-log source also carries each merge-order
 // peer's change-branch log under `### <peer>`. A repository git cannot read at all is reported in
 // a `note:`
@@ -551,13 +558,13 @@ func Bundle(change string, run records.Run, summary string, summaryFound bool, p
 	}
 
 	if repo != "" {
-		for _, file := range []string{"tasks.md", "design.md", "narrative.md"} {
+		for _, file := range archivedFiles {
 			label := archiveDir + "/" + change + "/" + file
 			content, err := show(g, repo, rev, label)
 			add(label, string(content), err == nil)
 		}
 	} else {
-		for _, file := range []string{"tasks.md", "design.md", "narrative.md"} {
+		for _, file := range archivedFiles {
 			add(archiveDir+"/"+change+"/"+file, "", false)
 		}
 	}
````

  - [ ] **Step 3: Verify.** `cd stats && gofmt -l ./internal/selfreview && go vet ./internal/selfreview && go test ./internal/selfreview/ -run 'TestBundle' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `python3 scripts/check-markdown-integrity.py` and `scripts/check-installed-citations.sh` exit 0.
  - [ ] **Step 4: Commit** the paths in `**Files:**` with the subject below.

**Build:** green
**Files:** `stats/internal/selfreview/git.go`, `stats/internal/selfreview/bundle_test.go`, `skills/flow-contracts/finish-contract-run1.md`, `skills/flow-fast/SKILL.md`
**Tests:** `TestBundleServesArchivedVisualVerification`
**Regression:** reverting the commit drops `visual-verification.md` from the bundle: `TestBundleServesArchivedVisualVerification` fails, and the found-line assertions read `of 7 sources` again.
**Baseline:** before=26 after=27
<!-- measured: cat stats/internal/selfreview/*_test.go | grep -cE '^func Test[A-Za-z0-9_]*\(' @ 5a1601fa (before; after with this task's diff applied in the scratch clone) -->
**Commit:** feat(selfreview): serve the archived visual-verification.md in the bundle
**After:** none

**Decision:** bundle-serves-visual-verification

- [ ] 2. The report guard holds six-angle reports to six angles, and the table gains angle 7
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Apply the `skills/flow-self-review/SKILL.md` and `stats/internal/guard/check_self_review_report_test.go` hunks of the diff in step 2 alone (`git apply --include='skills/flow-self-review/SKILL.md' --include='stats/internal/guard/check_self_review_report_test.go' <the diff>`) and run `cd stats && go test ./internal/guard/ -run TestCheckSelfReviewReport -count=1`: cases 7, 37 and 39 FAIL — the defect: with a seventh table row, every six-angle report, the real tree's 65 included, is reported missing `flow-eval`, and no list can hold them to six.
  <!-- measured: the step-1 command with only those hunks applied, after Task 1, in a scratch clone @ 5a1601fa (exit 1; case 7, case 37 and case 39 fail) -->
  - [ ] **Step 2: Implement.** Apply the rest of this diff (`git apply --exclude='skills/flow-self-review/SKILL.md' --exclude='stats/internal/guard/check_self_review_report_test.go' <the diff>`). Then regenerate the list from the tree this commit sits on, so a report landed on the base since `5a1601fa` is held too: `ls docs/self-review | grep '\.md$' | grep -v -- '-context.md$' | grep -vxFf docs/self-review/five-angle-reports.txt | LC_ALL=C sort > docs/self-review/six-angle-reports.txt` — at `5a1601fa` it writes exactly the 65 lines below.
  <!-- measured: that command | grep -c . @ 5a1601fa (65) -->

````diff verified:applied with git apply in plan order to a scratch clone at 5a1601fa, reproducing the planner's working tree byte for byte; gofmt -l and go vet clean; go test ./internal/guard/ ./internal/selfreview/ and scripts/run-guard-tests.sh pass with all three tasks applied
diff --git a/docs/self-review/six-angle-reports.txt b/docs/self-review/six-angle-reports.txt
new file mode 100644
index 00000000..5974156d
--- /dev/null
+++ b/docs/self-review/six-angle-reports.txt
@@ -0,0 +1,65 @@
+kan-550-flow-improvement-on-a-miss-dispatch-tooling-self-review.md
+kan-627-flow-fix-spectre-validate-is-red-on-develop-so-self-review.md
+kan-677-flow-fix-research-note-task-copies-under-docs-self-review.md
+kan-678-flow-fix-baseline-fields-ran-on-three-self-review.md
+kan-694-flow-improvement-when-a-live-report-persists-self-review.md
+kan-721-flow-improvement-after-the-third-finding-of-one-self-review.md
+kan-765-scope-fix-round-re-reviews-to-the-fix-diff-self-review.md
+kan-766-flow-live-check-plan-assumptions-and-size-proof-self-review.md
+kan-767-flow-serve-the-flow-state-directory-from-one-self-review.md
+kan-768-classify-pycache-pyc-and-tsbuildinfo-as-self-review.md
+kan-770-flow-enforce-that-a-finding-s-fixed-status-is-self-review.md
+kan-772-flow-codify-the-silence-rule-for-unanswered-mid-self-review.md
+kan-774-flow-prove-and-fix-the-exit-3-message-race-in-self-review.md
+kan-775-flow-require-a-reverted-prototype-s-rejection-self-review.md
+kan-777-parent-shell-discipline-in-contracts-self-review.md
+kan-778-agents-port-the-next-five-slowest-guards-to-the-self-review.md
+kan-781-record-failed-approaches-in-the-run-narrative-self-review.md
+kan-782-flow-automation-setup-sh-should-render-the-self-review.md
+kan-784-flow-fix-the-file-path-close-rule-cannot-close-self-review.md
+kan-785-flow-fix-batch-a-fix-round-s-operator-handbacks-self-review.md
+kan-786-check-task-commit-fields-skipped-cross-repo-self-review.md
+kan-788-flow-cost-start-runs-on-the-cheap-panel-shape-self-review.md
+kan-789-flow-improvement-mutation-kill-every-fix-self-review.md
+kan-791-flow-improvement-every-withdrawn-finding-carries-self-review.md
+kan-792-never-exempt-a-fix-round-from-re-review-self-review.md
+kan-794-flow-fix-plan-shape-f4-and-commit-fields-self-review.md
+kan-795-flow-fix-reproducer-exit-contract-guard-cannot-self-review.md
+kan-797-flow-fix-default-landing-route-never-resolves-self-review.md
+kan-798-flow-fix-protect-main-checkout-py-false-self-review.md
+kan-800-flow-improvement-defects-become-tasks-the-self-review.md
+kan-805-flow-fix-process-lessons-live-only-in-per-run-self-review.md
+kan-807-flow-improvement-scope-verification-to-the-diff-self-review.md
+kan-808-flow-automation-guard-agents-md-and-claude-md-self-review.md
+kan-809-flow-fix-stage-marks-lost-on-session-self-review.md
+kan-811-flow-improvement-a-mid-run-contract-change-self-review.md
+kan-812-flow-improvement-dated-correction-blocks-and-self-review.md
+kan-813-flow-improvement-re-verify-against-a-fresh-head-self-review.md
+kan-816-flow-fix-the-archive-step-s-set-pair-loop-does-self-review.md
+kan-817-flow-fix-the-done-when-named-gitignored-self-review.md
+kan-819-flow-cost-standing-instructions-keep-self-review.md
+kan-820-flow-cost-a-fix-dispatch-went-out-on-the-wrong-self-review.md
+kan-823-flow-fix-prepare-archive-branch-sh-fails-self-review.md
+kan-823-prepare-archive-branch-sh-fails-silently-self-review.md
+kan-826-flow-improvement-attribute-every-unexplained-self-review.md
+kan-829-flow-improvement-fix-pipeline-defects-at-the-self-review.md
+kan-838-flow-fast-record-guard-self-review.md
+kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix-self-review.md
+kan-840-flow-improvement-mark-known-ceilings-in-code-self-review.md
+kan-841-agents-port-the-next-ten-slowest-bash-scripts-to-self-review.md
+kan-842-agents-port-the-next-ten-slowest-bash-scripts-to-self-review.md
+kan-843-agents-remove-cursor-codex-harnesses-myflow-era-self-review.md
+kan-844-agents-speed-up-scripts-test-setup-sh-self-review.md
+kan-850-agents-port-the-visual-verify-scripts-to-the-go-self-review.md
+kan-860-mechanics-into-scripts-self-review.md
+kan-862-flow-fix-review-minors-inline-during-the-run-self-review.md
+kan-864-agents-flow-base-branch-override-self-review.md
+kan-873-agents-port-the-next-five-slowest-bash-guards-to-self-review.md
+kan-875-self-review-auto-fixes-all-angles-self-review.md
+kan-876-flow-speed-up-the-e2e-fidelity-live-verification-self-review.md
+kan-916-per-worktree-lsp-self-review.md
+kan-925-most-dispatch-rows-never-get-agent-id-self-review.md
+kan-943-flow-review-panel-scope-the-panel-on-a-fix-run-self-review.md
+kan-959-call-efficiency-skill-and-reviewer-slot-self-review.md
+planner-chooses-models-drop-default-model-self-review.md
+retire-bugbot-security-promote-failure-modes-self-review.md
diff --git a/scripts/check-self-review-report.sh b/scripts/check-self-review-report.sh
index 7f1eb7fb..0883eae8 100755
--- a/scripts/check-self-review-report.sh
+++ b/scripts/check-self-review-report.sh
@@ -58,14 +58,16 @@
 # table the way a hardcoded copy would. (The original five came from the
 # kan-200 self-review spec's "One combined reasoning pass" angle table.)
 #
-# FIVE-ANGLE REPORTS ARE A FROZEN LIST. `five-angle-reports.txt` in the
+# EARLIER REPORTS ARE FROZEN LISTS. `five-angle-reports.txt` in the
 # scanned directory names, one basename per line, every report written
-# before angle 6 (`flow-speed`) existed. A listed report is checked against
-# the first five angles of the table only; every report not on the list
-# must carry every angle the table serves. The list is frozen: it never
-# gains an entry, so a new report cannot be exempted from a later angle. A
-# directory with no list exempts nothing. A list that exists but cannot be
-# read is "cannot answer" (exit 2), never an empty exemption set.
+# before angle 6 (`flow-speed`) existed, and `six-angle-reports.txt` every
+# report written after it and before angle 7 (`flow-eval`). A listed report
+# is checked against the first five, or six, angles of the table only;
+# every report on neither list must carry every angle the table serves. A
+# list is frozen: it never gains an entry, so a new report cannot be
+# exempted from a later angle. A directory with no list exempts nothing. A
+# list that exists but cannot be read is "cannot answer" (exit 2), never an
+# empty exemption set.
 #
 # PER-REPORT COVERAGE, via coverage.go (scripts/lib/coverage.sh's Go twin).
 # Each report's recorded
diff --git a/skills/flow-self-review/SKILL.md b/skills/flow-self-review/SKILL.md
index c996408c..d1256d8d 100644
--- a/skills/flow-self-review/SKILL.md
+++ b/skills/flow-self-review/SKILL.md
@@ -56,6 +56,7 @@ present-but-empty for an angle with no findings, never omitted:
 | 4 | What could be automated or moved to a script | `flow-automation` |
 | 5 | What could move to the Go app or its persistent storage | `flow-stats-app` |
 | 6 | What can be sped up — wall-clock time the run spent waiting or working that a pipeline change would cut | `flow-speed` |
+| 7 | What would make a good flow-eval case — a task, a review finding or a visual pair from this run | `flow-eval` |
 
 Angle 5's remit covers the records the pipeline writes to files today and the derivation work
 now done in Bash or by the agent — **not** what the SPA should display. Angle 6 is elapsed time, where angle 2 is what a step spends: serial steps that could run in parallel, slow guards, builds or test runs, waits, and redundant re-runs. A silent angle and a
diff --git a/stats/internal/guard/check_self_review_report_test.go b/stats/internal/guard/check_self_review_report_test.go
index 5cabbdb1..11ce67fc 100644
--- a/stats/internal/guard/check_self_review_report_test.go
+++ b/stats/internal/guard/check_self_review_report_test.go
@@ -36,9 +36,12 @@ var (
 	csrAuto  = csrSec("Automation", "flow-automation", csrNone)
 	csrStats = csrSec("Stats app", "flow-stats-app", csrNone)
 	csrSpeed = csrSec("Speed", "flow-speed", csrNone)
-	// csrFive is a report as written before angle 6 existed.
+	csrEval  = csrSec("Flow-eval cases", "flow-eval", csrNone)
+	// csrFive is a report as written before angle 6 existed, csrSix one as
+	// written before angle 7 existed.
 	csrFive      = strings.Join([]string{csrFix, csrCost, csrImp, csrAuto, csrStats}, "\n")
-	csrCompliant = csrFive + "\n" + csrSpeed
+	csrSix       = csrFive + "\n" + csrSpeed
+	csrCompliant = csrSix + "\n" + csrEval
 	// csrNoneFive is the five original sections, each carrying the marker.
 	csrNoneFive = strings.Join([]string{csrSec("Cost", "flow-cost", csrNone), csrImp, csrAuto, csrStats}, "\n")
 )
@@ -371,25 +374,59 @@ func TestCheckSelfReviewReport(t *testing.T) {
 		}},
 		{func(t *testing.T) csrResult {
 			return csrRun(t, nil, csrReport(t, strings.Join([]string{csrFix,
-				csrSec("Cost", "flow-cost", "- **[flow-cost]** Every panel slot gathers the same context independently — fixed: abc1234"), csrImp, csrAuto, csrStats, csrSpeed}, "\n")))
+				csrSec("Cost", "flow-cost", "- **[flow-cost]** Every panel slot gathers the same context independently — fixed: abc1234"), csrImp, csrAuto, csrStats, csrSpeed, csrEval}, "\n")))
 		}, []check{
 			{"case 34: a fixed finding naming a sha exits 0", rcIs(0)},
 		}},
 		{func(t *testing.T) csrResult {
 			return csrRun(t, nil, csrReport(t, strings.Join([]string{csrFix,
-				csrSec("Cost", "flow-cost", "- **[flow-cost]** Every panel slot gathers the same context independently — fixed:"), csrImp, csrAuto, csrStats, csrSpeed}, "\n")))
+				csrSec("Cost", "flow-cost", "- **[flow-cost]** Every panel slot gathers the same context independently — fixed:"), csrImp, csrAuto, csrStats, csrSpeed, csrEval}, "\n")))
 		}, []check{
 			{"case 35: a fixed finding with no sha is caught", rcIs(1)},
 			{"case 35: finding names the missing sha", has("fixed", "no sha")},
 		}},
 		{func(t *testing.T) csrResult {
 			return csrRun(t, nil, csrReport(t, strings.Join([]string{csrFix,
-				csrSec("Cost", "flow-cost", "- **[flow-cost]** Every panel slot gathers the same context independently — fixed: KAN-9"), csrImp, csrAuto, csrStats, csrSpeed}, "\n")))
+				csrSec("Cost", "flow-cost", "- **[flow-cost]** Every panel slot gathers the same context independently — fixed: KAN-9"), csrImp, csrAuto, csrStats, csrSpeed, csrEval}, "\n")))
 		}, []check{
 			{"case 36: a fixed finding naming a non-sha is caught", rcIs(1)},
 			{"case 36: finding names the malformed sha", has("fixed", "malformed sha", "KAN-9")},
 		}},
 
+		{func(t *testing.T) csrResult {
+			return csrRun(t, nil, csrDir(t, map[string]string{"fixture-self-review.md": csrSix,
+				"six-angle-reports.txt": "other-self-review.md\nfixture-self-review.md\n"}))
+		}, []check{
+			{"case 37: a listed six-angle report exits 0", rcIs(0)},
+		}},
+		{func(t *testing.T) csrResult {
+			return csrRun(t, nil, csrDir(t, map[string]string{"fixture-self-review.md": csrSix,
+				"six-angle-reports.txt": "old-fixture-self-review.md\n"}))
+		}, []check{
+			{"case 38: an unlisted six-angle report is caught", rcIs(1)},
+			{"case 38: the missing-section finding names flow-eval", has("missing section for angle", "flow-eval")},
+		}},
+		{func(t *testing.T) csrResult {
+			dir := csrReport(t, csrCompliant)
+			mkdir(t, filepath.Join(dir, "six-angle-reports.txt"))
+			return csrRun(t, nil, dir)
+		}, []check{
+			{"case 39: an unreadable six-angle list is 'cannot answer'", rcIs(2)},
+			{"case 39: the die message names the unreadable list", has("cannot read the six-angle report list")},
+		}},
+		{func(t *testing.T) csrResult {
+			return csrRun(t, nil, csrDir(t, map[string]string{"fixture-self-review.md": csrSix,
+				"five-angle-reports.txt": "fixture-self-review.md\n", "six-angle-reports.txt": "fixture-self-review.md\n"}))
+		}, []check{
+			{"case 40: a report on both lists is held to the first list's five angles", rcIs(0)},
+		}},
+		{func(t *testing.T) csrResult {
+			return csrRun(t, nil, csrDir(t, map[string]string{"fixture-self-review.md": csrFive,
+				"six-angle-reports.txt": "fixture-self-review.md\n"}))
+		}, []check{
+			{"case 41: a five-angle report on the six-angle list is caught", rcIs(1)},
+			{"case 41: the missing-section finding names flow-speed", has("missing section for angle", "flow-speed")},
+		}},
 		// Beyond the harness: the exact output of a clean run and of a
 		// violation, the refusals it never reached, and the collation.
 		{func(t *testing.T) csrResult {
@@ -399,7 +436,7 @@ func TestCheckSelfReviewReport(t *testing.T) {
 			return r
 		}, []check{
 			{"a clean run prints the verdict and the coverage fragment on stdout", func(r csrResult) bool {
-				return r.out == "SELF-REVIEW-REPORT-OK: <dir> — 1 report(s) checked\n  <dir>/fixture-self-review.md 8\n" && r.err == ""
+				return r.out == "SELF-REVIEW-REPORT-OK: <dir> — 1 report(s) checked\n  <dir>/fixture-self-review.md 9\n" && r.err == ""
 			}},
 		}},
 		{func(t *testing.T) csrResult {
diff --git a/stats/internal/guard/selfreviewreport.go b/stats/internal/guard/selfreviewreport.go
index 8a1ab438..af69e068 100644
--- a/stats/internal/guard/selfreviewreport.go
+++ b/stats/internal/guard/selfreviewreport.go
@@ -47,14 +47,23 @@ var (
 
 const (
 	srrNoneMarker = "_none — this angle produced no findings._"
-	// srrOriginalAngles is the table's size when every report on the frozen
-	// five-angle list was written (the script header's FIVE-ANGLE REPORTS
-	// ARE A FROZEN LIST).
-	srrOriginalAngles = 5
 )
 
+// srrFrozenLists are the frozen report lists (the script header's EARLIER
+// REPORTS ARE FROZEN LISTS): each file in the scanned directory names the
+// reports written while the table held angles angles, which is all a listed
+// report is checked against. A basename on more than one list takes the
+// first.
+var srrFrozenLists = []struct {
+	file, name string
+	angles     int
+}{
+	{"five-angle-reports.txt", "five-angle", 5},
+	{"six-angle-reports.txt", "six-angle", 6},
+}
+
 // checkSelfReviewReport is scripts/check-self-review-report.sh's body. Its
-// contract — the report shape, the served angle labels, the frozen list,
+// contract — the report shape, the served angle labels, the frozen lists,
 // coverage and the exit codes — is the script's header; the reasoning
 // behind each step is here.
 //
@@ -116,20 +125,24 @@ func checkSelfReviewReport(args []string, env Env, stdout, stderr io.Writer) int
 		return die("the canonical angle table yielded no labels: %s", contract)
 	}
 
-	// The frozen list, newline-wrapped so a basename is matched whole. A
-	// list that exists but cannot be read is "cannot answer", never an empty
-	// exemption set; the read failing is reported as the bash's `cat` did.
-	listPath := target + "/five-angle-reports.txt"
-	fiveAngle := "\n"
-	if _, err := os.Stat(abs(listPath)); err == nil {
-		b, err := os.ReadFile(abs(listPath))
-		if err != nil {
-			// gsFindErr renders the line a tool prints for a failed path; cat's is
-			// find's with its own name.
-			fmt.Fprint(stderr, "cat: "+strings.TrimPrefix(gsFindErr(listPath, err), "find: "))
-			return die("cannot read the five-angle report list: %s", listPath)
+	// The frozen lists, each newline-wrapped so a basename is matched whole.
+	// A list that exists but cannot be read is "cannot answer", never an
+	// empty exemption set; the read failing is reported as the bash's `cat`
+	// did.
+	frozen := make([]string, len(srrFrozenLists))
+	for i, l := range srrFrozenLists {
+		listPath := target + "/" + l.file
+		frozen[i] = "\n"
+		if _, err := os.Stat(abs(listPath)); err == nil {
+			b, err := os.ReadFile(abs(listPath))
+			if err != nil {
+				// gsFindErr renders the line a tool prints for a failed path; cat's is
+				// find's with its own name.
+				fmt.Fprint(stderr, "cat: "+strings.TrimPrefix(gsFindErr(listPath, err), "find: "))
+				return die("cannot read the %s report list: %s", l.name, listPath)
+			}
+			frozen[i] = "\n" + strings.TrimRight(string(b), "\n") + "\n"
 		}
-		fiveAngle = "\n" + strings.TrimRight(string(b), "\n") + "\n"
 	}
 
 	files, rc := srrReports(env, target, abs(target), stderr)
@@ -287,16 +300,19 @@ func checkSelfReviewReport(args []string, env Env, stdout, stderr io.Writer) int
 			}
 		}
 		required := n
-		if strings.Contains(fiveAngle, "\n"+base+"\n") {
-			required = srrOriginalAngles
+		for i, l := range srrFrozenLists {
+			if strings.Contains(frozen[i], "\n"+base+"\n") {
+				required = l.angles
+				break
+			}
 		}
 
 		// A report with no angle heading at all gets no per-section findings
 		// — there was nothing in it to check against — and is flagged
 		// through coverage's undeclared zero instead (KAN-73's regression
 		// shape). A recognizable report gets a "missing section" finding for
-		// each absent angle, a listed five-angle report for the first five
-		// only.
+		// each absent angle, a report on a frozen list for that list's first
+		// angles only.
 		count := 0
 		if found > 0 {
 			for j, label := range labels {
````

  - [ ] **Step 3: Verify.** `cd stats && gofmt -l ./internal/guard && go vet ./internal/guard && go test ./internal/guard/ -run TestCheckSelfReviewReport -count=1` passes; `scripts/check-self-review-report.sh` prints `SELF-REVIEW-REPORT-OK` and exits 0; `scripts/check-verbatim-moves.sh` exits 0.
  - [ ] **Step 4: Commit** the paths in `**Files:**` with the subject below.

**Build:** green
**Files:** `stats/internal/guard/selfreviewreport.go`, `stats/internal/guard/check_self_review_report_test.go`, `scripts/check-self-review-report.sh`, `docs/self-review/six-angle-reports.txt`, `skills/flow-self-review/SKILL.md`
**Tests:** Case 37, Case 38, Case 39, Case 40, Case 41
**Regression:** reverting the commit removes angle 7 from the table and the second frozen list; with the row kept and the list support reverted, case 7 (the real tree), case 37 (a listed six-angle report) and case 39 (an unreadable six-angle list) fail.
**Baseline:** before=35 after=40
<!-- measured: grep -oE '"case [0-9]+:' stats/internal/guard/check_self_review_report_test.go | sort -u | grep -c . @ 5a1601fa (before; after with this task's diff applied in the scratch clone) -->
**Commit:** feat(guard): hold six-angle self-review reports to their six angles
**After:** Task 1

**Decision:** six-angle-frozen-list

- [ ] 3. `/flow-self-review` runs angle 7, offers its candidates for mining and mines the chosen ones
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -c 'mining' skills/flow-self-review/SKILL.md` after Task 2: it prints `0` — nothing names a mining ask, a mining step or a candidate template, only the table row names `flow-eval`; `grep -n 'step 5 of `/flow-self-review`' skills/flow-contracts/finish-contract-run1.md skills/flow-fast/SKILL.md` lists two citations of step 5 (filing) for the landing push's shas, which item 5 of step 3 states.
  <!-- measured: git show HEAD:skills/flow-self-review/SKILL.md | grep -c 'mining' @ 5a1601fa (0) -->
  - [ ] **Step 2: Implement.** Apply this diff (one line of context, so no context line reads as a Markdown fence):

````diff verified:applied with git apply in plan order to a scratch clone at 5a1601fa, reproducing the planner's working tree byte for byte; gofmt -l and go vet clean; go test ./internal/guard/ ./internal/selfreview/ and scripts/run-guard-tests.sh pass with all three tasks applied
diff --git a/commands-claude/flow-self-review.md b/commands-claude/flow-self-review.md
index 60780f16..55ddf94b 100644
--- a/commands-claude/flow-self-review.md
+++ b/commands-claude/flow-self-review.md
@@ -9,4 +9,4 @@ Follow that skill exactly. **Standalone, not a pipeline stage** — it takes one
 no per-change state file, and marks no `flow stage` call. It reads the context bundle an
-older `/flow` or `/flow-fast` run saved — newer runs run the same pass in-run — dispatches the six-angle reasoning pass, plans, fixes and
-lands every finding that is not big, files the big ones, rates, records each finding in the flow
+older `/flow` or `/flow-fast` run saved — newer runs run the same pass in-run — dispatches the seven-angle reasoning pass, plans, fixes and
+lands every finding that is not big, files the big ones, mines the flow-eval cases the operator approves, rates, records each finding in the flow
 store, writes the report, deletes the bundle, and lands both on the default branch.
diff --git a/skills/flow-contracts/finish-contract-run1.md b/skills/flow-contracts/finish-contract-run1.md
index a4f15cc8..a7962217 100644
--- a/skills/flow-contracts/finish-contract-run1.md
+++ b/skills/flow-contracts/finish-contract-run1.md
@@ -307,3 +307,3 @@ has the bundle report it skipped) and holds the whole in memory — it is never
 Over that bundle the session runs steps 2–6 of `/flow-self-review`
-(`skills/flow-self-review/SKILL.md`), canonical for the six angles, the fix loop and its
+(`skills/flow-self-review/SKILL.md`), canonical for the seven angles, the fix loop and its
 dispatch pairs, where the fixes land, the filing-and-rating prompt and the store rows. It then
@@ -318,3 +318,3 @@ land-self-review-report.sh "<canonical-worktree>" "spectre/<name>" \
 
-The route's landing push carries the fix commits' shas per step 5 of `/flow-self-review`.
+The route's landing push carries the fix commits' shas per item 5 of step 3 of `/flow-self-review`, and on merge and push is followed by step 5's mining of the chosen flow-eval candidates, whose cases, branch and shas the run's final report names.
 
diff --git a/skills/flow-contracts/operator-prompts-auto-resolution.md b/skills/flow-contracts/operator-prompts-auto-resolution.md
index e92cdab6..477ea405 100644
--- a/skills/flow-contracts/operator-prompts-auto-resolution.md
+++ b/skills/flow-contracts/operator-prompts-auto-resolution.md
@@ -60,3 +60,4 @@ These are asked, or stop with `## Question`, exactly as their call sites state:
 - **Outward-facing or irreversible actions.** Push, merge, opening a PR, archiving, deleting, and
-  every Jira write — the follow-up filing and the self-review filing included.
+  every Jira write — the follow-up filing and the self-review filing included — and the
+  self-review's flow-eval mining.
   The global rules require these confirmed.
diff --git a/skills/flow-fast/SKILL.md b/skills/flow-fast/SKILL.md
index d454892d..b83ce41f 100644
--- a/skills/flow-fast/SKILL.md
+++ b/skills/flow-fast/SKILL.md
@@ -299,3 +299,3 @@ lessons home (**Process lessons**, `skills/flow-contracts/lessons.md`).
 Over that bundle run steps 2–6 of `/flow-self-review` (`skills/flow-self-review/SKILL.md`),
-canonical for the six angles, the fix loop, where the fixes land — on `<name>` itself
+canonical for the seven angles, the fix loop, where the fixes land — on `<name>` itself
 when `<worktree>` is `<agents repo>`'s — the filing-and-rating prompt and the store rows. Write
@@ -312,3 +312,3 @@ land-self-review-report.sh "<worktree>" "<name>" \
 The report and the fixes land with the change on every route. **7. Land**'s landing push carries the fix
-commits' shas per step 5 of `/flow-self-review`, in whichever invocation makes it. A staged index holding anything
+commits' shas per item 5 of step 3 of `/flow-self-review`, in whichever invocation makes it. A staged index holding anything
 beyond the chain's own path refuses the commit (`LAND-FOREIGN-STAGED`) — clear the staging or land
diff --git a/skills/flow-self-review/SKILL.md b/skills/flow-self-review/SKILL.md
index d1256d8d..9d873e90 100644
--- a/skills/flow-self-review/SKILL.md
+++ b/skills/flow-self-review/SKILL.md
@@ -2,4 +2,4 @@
 name: flow-self-review
-description: Run a change's self-review pass, a one-shot opus dispatch — inside `/flow`'s integrate and `/flow-fast`'s verify, or standalone from the context bundle an older run saved; plan in full, fix and land every finding that is not big, file the big ones, record each in the flow store, rate, write the report, delete the bundle. Standalone, not a pipeline stage. Use for /flow-self-review.
-allowed-tools: Bash(git:*), Bash(flow:*), Bash(scripts/check-self-review-report.sh:*), Bash(land-self-review-report.sh:*)
+description: Run a change's self-review pass, a one-shot opus dispatch — inside `/flow`'s integrate and `/flow-fast`'s verify, or standalone from the context bundle an older run saved; plan in full, fix and land every finding that is not big, file the big ones, mine the flow-eval cases the operator approves, record each in the flow store, rate, write the report, delete the bundle. Standalone, not a pipeline stage. Use for /flow-self-review.
+allowed-tools: Bash(git:*), Bash(flow:*), Bash(go:*), Bash(scripts/check-self-review-report.sh:*), Bash(land-self-review-report.sh:*)
 license: MIT
@@ -12,3 +12,3 @@ Run a change's self-review reasoning pass — the only one the pipeline has —
 (`skills/flow-fast/SKILL.md`) assembles: inside that run, or, as this standalone command, from the
-bundle an older run saved on the change branch before landing. This file is canonical for the six angles, what is fixed and what may be filed, the
+bundle an older run saved on the change branch before landing. This file is canonical for the seven angles, what is fixed, what may be filed and what may be mined, the
 filing-and-rating prompt, the store record and the report. **Neither the pass nor its fixes run
@@ -37,3 +37,3 @@ checkout on any other branch: print the branch and stop.
 
-### 2. Read the bundle and run the six angles
+### 2. Read the bundle and run the seven angles
 
@@ -47,3 +47,3 @@ store rows and the report.
 
-The pass runs the six-angle table over the bundle, every angle explicit —
+The pass runs the seven-angle table over the bundle, every angle explicit —
 present-but-empty for an angle with no findings, never omitted:
@@ -64,2 +64,22 @@ skipped angle are indistinguishable to a reader, which is why an empty angle say
 
+Angle 7 lists the run's tasks, review findings and visual-verification pairs that would make good
+flow-eval cases — what `case add impl|review|visual` mines, per `<agents repo>/../flow-eval/README.md`,
+the `flow-eval` checkout beside `<agents repo>` — one candidate per case, in its kind's template,
+every field present:
+
+- **impl** — the task, as `tasks.md` records it; the resulting code, an excerpt under its
+  worktree-resolved `path:line`; why it is a good candidate.
+- **review** — each finding with its severity, its location and the reviewer's own sentence; the
+  bad code; the fixed code, with the fixing commit's sha; why it is a good candidate.
+- **visual** — the actual app screenshot and the mockup screenshot, each by absolute path; the
+  findings; why it is a good candidate.
+
+Every path is absolute. A file that exists only in git history, or only inside a worktree that
+cleanup removes, is extracted by this session before step 4 — `git -C <repo> show <sha>:<path>`
+into `<state-dir>/flow-eval/<name>/`, `<state-dir>` being the path `flow state dir` prints for
+that repository — and cited there. A candidate names no defect: it is recorded `noted`, never fixed or filed — the one `noted`
+item offered, for mining in step 4. Angle 7 reads its candidates from the run's own work in any repository the
+run touched, product code included, and reads beyond the bundle only for the commits and files
+the bundle names.
+
 An `In-run pipeline fix:` line in the bundle (**Pipeline defects found mid-run**,
@@ -68,3 +88,3 @@ is never offered for filing; one that reads `deferred` is an angle-1 finding lik
 
-**A finding is fixed or filed only from the six angles.** A finding about the pipeline itself is
+**A finding is fixed or filed only from the seven angles.** A finding about the pipeline itself is
 fixed in step 3 unless it is `big`, and offered under its angle when it is. A finding that names no defect — what
@@ -76,6 +96,6 @@ scope: it is never offered or filed, and goes to a fix run instead — a pass in
 here, at the end of this step (**the sweep**, `skills/flow-contracts/known-bugs.md`). The report carries no
-section beyond the six angles and the rating. The filing prompt is never waived: a pass with
-no operator to answer it files nothing and records every offered finding `declined`.
+section beyond the seven angles and the rating. The filing prompt is never waived: a pass with
+no operator to answer it files and mines nothing and records every offered finding `declined`.
 
-**One combined pass** — never six separate reads. The pass covers what the bundle holds and
+**One combined pass** — never seven separate reads. The pass covers what the bundle holds and
 nothing beyond it — say so in the report's `**Deferred:**` line.
@@ -227,10 +247,13 @@ body, naming its landed sha; it is not offered. A prompt's option text cannot ca
 explanation, so the prompt records the decision only: a filed issue is durable, and an
-explanation arriving afterward describes something the operator did not agree to. The filing ask
-and the rating are **one `AskUserQuestion` call**, shape per **Operator prompts**
+explanation arriving afterward describes something the operator did not agree to. The filing ask,
+the mining ask and the rating are **one `AskUserQuestion` call**, shape per **Operator prompts**
 (`skills/flow-contracts/operator-prompts.md`): up to three multi-select questions of
-three offered findings each, every option prefixed with its angle's label, plus **None — file nothing**
-as the default; the rating last, `5 — excellent` / `4 — good` / `3 — fine` / `2 — rough`, a `1`
-typed through the tool's free-text "Other". More than nine findings roll the overflow into one
-further call of the same shape, without the rating. With no finding to offer, the call
-carries the rating alone.
+three offered items each — findings to file or angle-7 candidates to mine, never both in one
+question — every option prefixed with its angle's label, plus **None — file nothing** or
+**None — mine nothing** as the default; the rating last, `5 — excellent` / `4 — good` / `3 — fine` / `2 — rough`, a `1`
+typed through the tool's free-text "Other". More than nine offered items roll the overflow into one
+further call of the same shape, without the rating. With nothing to offer, the call
+carries the rating alone. A candidate is mined only when the operator selects it: silence, a
+stalled prompt and **Auto-resolution** (`skills/flow-contracts/operator-prompts-auto-resolution.md`)
+mine nothing. An angle-7 candidate's explanation in the message body is its template.
 
@@ -249,4 +272,5 @@ plainly. A safe default may resolve silence to the empty set (an explicit "None"
 recommended) or to the full set (every listed option, silence needing no option of its own to name
-it) — whichever matches what the options actually control. The one live multi-select call site,
-the self-review filing ask, chooses the empty set: silence selects **None — file nothing**.
+it) — whichever matches what the options actually control. The live multi-select call sites,
+the self-review filing and mining asks, choose the empty set: silence selects
+**None — file nothing** and **None — mine nothing**.
 
@@ -258,2 +282,25 @@ this session: print `⚠ Jira: skipped — <reason>` and record that finding `de
 
+Mine each chosen candidate into the `flow-eval` checkout once the change's commits are on its
+repository's default branch — standalone, at once; in a run, after the landing push of item 5 of
+step 3 on merge and push. On any other route nothing is mined: each chosen candidate's
+`case add` command is named in step 8 for after the merge. The mining runs in a worktree, never
+the `flow-eval` main checkout, one branch per pass —
+`git -C <agents repo>/../flow-eval worktree add -b flow-eval-cases-<name> <agents repo>/../flow-eval-worktrees/flow-eval-cases-<name> origin/main`
+— from that worktree's root, one `case add` per candidate, per `<agents repo>/../flow-eval/docs/cases.md`:
+
+```bash
+go run ./cmd/flow-eval case add impl --repo <code checkout> [--plan-repo <plan checkout>] --change <name> --task <n>
+go run ./cmd/flow-eval case add review --repo <code checkout> [--plan-repo <plan checkout>] --change <name>
+go run ./cmd/flow-eval case add visual --id <id> --mockup <mockup path> --capture <capture path> --verdict match|mismatch [--defect '<text>' [--region x,y,w,h]]…
+```
+
+`--repo` and `--plan-repo` are the repositories' main checkouts, standalone, or the change's
+worktrees in a run, whose `case.json` `source.repo` is then set to the main checkout's directory
+name. Edit each draft as that file's "What to check and edit" asks, from what the
+candidate's evidence answers; run `case validate` the same way; and commit each case
+alone as `feat(cases): add <role> case <id>`. The branch is never merged or pushed: the
+operator's review of its commits is the label approval the README names. No `flow-eval`
+checkout there: print `⚠ flow-eval: skipped — <reason>` and mine nothing. A `case add` or
+`case validate` that fails leaves its candidate unmined, its output named in the final report.
+
 ### 6. Record every finding in the flow store
@@ -276,3 +323,3 @@ warning line and the pass continues — the report below is the durable record.
 Write `<project>/docs/self-review/<name>-self-review.md` in the shape
-`<project>/scripts/check-self-review-report.sh` checks — one section per angle, all six present; each
+`<project>/scripts/check-self-review-report.sh` checks — one section per angle, all seven present; each
 finding one line naming its angle's label, the finding, and its disposition (`fixed: <sha>`, `filed: <KEY>`,
@@ -285,2 +332,5 @@ title:
 
+An angle-7 candidate is a `noted` line naming its kind and whether the operator chose to mine
+it, its template's fields beneath it as indented sub-bullets.
+
 Delete `<bundle>`; a pass inside a run has no bundle file and commits the report alone, per its
@@ -293,3 +343,3 @@ Commit both paths in one commit through the landing chain's one script — the s
 `skills/flow/integrate.md` step 4 lands the report with, differing in the asserted branch,
-the removed context-bundle path and the `--push` — since the round-trip through the six-angle
+the removed context-bundle path and the `--push` — since the round-trip through the seven-angle
 pass and the filing-and-rating prompt above is long enough that the branch is worth re-checking
@@ -312,3 +362,3 @@ commit local and this run names it — never retried around.
 
-End naming the report path, the rating, the shas landed and the Jira keys filed (each `none` when empty), and each product defect the pass met (**the sweep**, `skills/flow-contracts/known-bugs.md`).
+End naming the report path, the rating, the shas landed, the Jira keys filed and the flow-eval cases mined, by branch and sha (each `none` when empty), each chosen candidate left unmined with its `case add` command, and each product defect the pass met (**the sweep**, `skills/flow-contracts/known-bugs.md`).
 
````

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `python3 scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`, `scripts/check-ask-silence.sh`, `scripts/check-dispatch-paragraphs.sh` and `scripts/check-self-review-report.sh` exit 0; `scripts/check-normative-inventory.sh` prints the same lines before and after the diff.
  - [ ] **Step 4: Commit** the paths in `**Files:**` with the subject below.

**Build:** green
**Files:** `skills/flow-self-review/SKILL.md`, `skills/flow-contracts/finish-contract-run1.md`, `skills/flow-fast/SKILL.md`, `commands-claude/flow-self-review.md`, `skills/flow-contracts/operator-prompts-auto-resolution.md`
**Tests:** **none** — prose; the guards in step 3 are the check
**Regression:** reverting the commit leaves angle 7 a bare table row: no template, no mining ask, no mining step, no report shape, and the six-angle wording and the two step-5 citations back.
**Baseline:** before=0 after=0
<!-- predicted: no tests in this task — Markdown prose only -->
**Commit:** feat(flow-self-review): add a seventh angle, flow-eval case candidates mined on approval
**After:** Task 1, 2

**Decision:** candidate-disposition-noted
**Decision:** mining-ask-in-step-4-call
**Decision:** mine-after-landing
**Decision:** mine-on-flow-eval-branch
**Decision:** flow-eval-beside-agents
**Decision:** durable-extract-state-dir
**Decision:** flow-eval-angle-reads-product-code
