# pipeline-fix-kan-941-step-6-playwright-fidelity-specs-for-the-cardio-22

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** `flow self-review bundle` elides element × property matrix rows under a matrix heading
at any level, so kan-941 step 6's bundle fits the 1048576-byte cap.

**Architecture:** one predicate in `elideMatrixRows` (`stats/internal/selfreview/git.go`) — a
heading is a matrix heading when its text holds `matrix`, any case — plus a new test and the run-1
contract sentence.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/` (this change's `verbatim-moves.txt` is already
  committed with the plan).
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; it was applied to this worktree at 02d83a03 and checked there. A diff that does not
  apply is reported, never hand-merged by guesswork.
- No new inline suppressions; `cd stats && gofmt -l .` prints nothing.
- Read-only against the dev store: never stop flowd or flow-postgres, never drop the `flow` DB.

## Review Focus

- A `###` and a `##` matrix heading now elide their rows; the `####` case, fenced lines and a
  non-matrix table (`### Departures`) are kept as before (`matrix-heading-by-name`).
- The contract sentence matches the code's rule.

- [ ] 1. Elide matrix rows under a matrix heading at any level
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Save the diff below to
    `/tmp/pf22-task1.diff`, apply its test hunk alone
    (`git apply --include='stats/internal/selfreview/bundle_test.go' /tmp/pf22-task1.diff`) and run
    `cd stats && go test ./internal/selfreview/ -run 'TestBundleElidesMatrixRows' -count=1`:
    `--- FAIL: TestBundleElidesMatrixRowsUnderAnyMatrixHeading`, the rows `ROW-H3` and `ROW-H2`
    still in the bundle; `TestBundleElidesMatrixRowsAndPNGStatLines` passes.
  - [ ] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/internal/selfreview/bundle_test.go' /tmp/pf22-task1.diff`):

````diff verified:applied with git apply to this worktree at 02d83a03; go test ./internal/selfreview/ passes with it and the new test fails with git.go untouched; gofmt -l, go vet ./internal/selfreview/, check-references, check-markdown-integrity, check-normative-inventory and check-verbatim-moves (with this change's verbatim-moves.txt) exit 0
diff --git a/skills/flow-contracts/finish-contract-run1.md b/skills/flow-contracts/finish-contract-run1.md
index bc237a34..07c94d8b 100644
--- a/skills/flow-contracts/finish-contract-run1.md
+++ b/skills/flow-contracts/finish-contract-run1.md
@@ -310,8 +310,8 @@ roots for the pipeline's changes), and the `git log --stat` of every implementat
 archive commit of the change, oldest first, followed by each peer the archived `link.md`'s
 `## Merge order` names, its `spectre/<name>` commits against its base under a `### <peer>`
 heading — and prints it as one Markdown document. Its `visual-verification.md` keeps every heading and
-drops the table rows under each `####` element × property matrix heading, and its `git log --stat`
-lines leave out every `*.png` path. A source that is absent is reported
+drops the table rows under each heading naming a matrix — any level, `matrix` in its text, any
+case — and its `git log --stat` lines leave out every `*.png` path. A source that is absent is reported
 `skipped: <source> (absent)` inside the bundle, never fatal.
 The session appends `## Session narrative` (one paragraph it writes for run 1 itself; the
 archived `narrative.md` is already a bundle section, and a change predating the narrative rule
diff --git a/stats/internal/selfreview/bundle_test.go b/stats/internal/selfreview/bundle_test.go
index e0502210..aa80de1e 100644
--- a/stats/internal/selfreview/bundle_test.go
+++ b/stats/internal/selfreview/bundle_test.go
@@ -217,6 +217,34 @@ func TestBundleServesArchivedVisualVerification(t *testing.T) {
 	}
 }
 
+func TestBundleElidesMatrixRowsUnderAnyMatrixHeading(t *testing.T) {
+	repo := gitRepo(t)
+	commitAll(t, repo, "a.go", "package a\n", "feat(demo): a")
+	writeArchiveBranch(t, repo, "demo", map[string]string{
+		"tasks.md": "# demo tasks\n",
+		"visual-verification.md": "# Visual verification — demo\n\n## Element × property matrices\n\n" +
+			"### cardio-add-workout:cardio-add-workout-A1 matrix (A1)\n\n| element | x |\n|---|---|\n| ROW-H3 | 1 |\n\n" +
+			"## Matrix\n\n| element | x |\n|---|---|\n| ROW-H2 | 1 |\n\n" +
+			"### Departures\n\n| element | departure |\n|---|---|\n| bar | colour |\n",
+	})
+
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, ExecRunner{})
+	if err != nil {
+		t.Fatalf("Bundle: %v", err)
+	}
+
+	for _, want := range []string{"### cardio-add-workout:cardio-add-workout-A1 matrix (A1)", "## Matrix", "_(3 matrix rows elided)_", "| bar | colour |"} {
+		if !strings.Contains(bundle, want) {
+			t.Errorf("bundle missing %q:\n%s", want, bundle)
+		}
+	}
+	for _, gone := range []string{"ROW-H3", "ROW-H2"} {
+		if strings.Contains(bundle, gone) {
+			t.Errorf("bundle still carries matrix row %q:\n%s", gone, bundle)
+		}
+	}
+}
+
 func TestBundleElidesMatrixRowsAndPNGStatLines(t *testing.T) {
 	repo := gitRepo(t)
 	if err := os.WriteFile(filepath.Join(repo, "shot.png"), []byte("\x89PNG"), 0o644); err != nil {
diff --git a/stats/internal/selfreview/git.go b/stats/internal/selfreview/git.go
index b88348d0..dea031ce 100644
--- a/stats/internal/selfreview/git.go
+++ b/stats/internal/selfreview/git.go
@@ -454,9 +454,10 @@ func dropPNGStatLines(s string) string {
 	return b.String()
 }
 
-// elideMatrixRows drops every table row between a `####` heading and the
-// next heading, leaving one `_(<n> matrix rows elided)_` line per run. A
-// line inside a ``` fence is never a heading or a row.
+// elideMatrixRows drops every table row between a heading naming a matrix
+// (its text holds `matrix`, any case, at any level) and the next heading,
+// leaving one `_(<n> matrix rows elided)_` line per run. A line inside a ```
+// fence is never a heading or a row.
 func elideMatrixRows(md string) string {
 	var b strings.Builder
 	inMatrix, inFence, run := false, false, 0
@@ -474,7 +475,7 @@ func elideMatrixRows(md string) string {
 			flush()
 		} else if strings.HasPrefix(l, "#") {
 			flush()
-			inMatrix = strings.HasPrefix(l, "#### ")
+			inMatrix = strings.Contains(strings.ToLower(l), "matrix")
 		} else if inMatrix && strings.HasPrefix(strings.TrimSpace(l), "|") {
 			run++
 			continue
````

  - [ ] **Step 3: Verify.** `cd stats && gofmt -l .` prints nothing;
    `cd stats && make -s web-dist && go vet ./internal/selfreview/` exits 0;
    `cd stats && go test ./internal/selfreview/ -run 'TestBundleElidesMatrixRows' -count=1` passes
    both tests; `scripts/check-references.sh`, `scripts/check-markdown-integrity.py` and
    `scripts/check-verbatim-moves.sh` exit 0.
  - [ ] **Step 4: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/selfreview/git.go`, `stats/internal/selfreview/bundle_test.go`, `skills/flow-contracts/finish-contract-run1.md`
**Tests:** `TestBundleElidesMatrixRowsUnderAnyMatrixHeading`
**Regression:** `TestBundleElidesMatrixRowsUnderAnyMatrixHeading` — reverting the commit keeps the rows under `### cardio-add-workout:cardio-add-workout-A1 matrix (A1)` and `## Matrix` in the bundle: the test fails on `ROW-H3` and `ROW-H2`.
**Baseline:** before=28 after=29
<!-- measured: grep -c '^func Test' stats/internal/selfreview/bundle_test.go (before at 02d83a03; after with this task's diff applied) -->
**Commit:** fix(selfreview): elide matrix rows under a matrix heading at any level
**After:** none

**Decision:** matrix-heading-by-name
