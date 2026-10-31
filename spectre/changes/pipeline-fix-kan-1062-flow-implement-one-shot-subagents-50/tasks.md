# pipeline-fix-kan-1062-flow-implement-one-shot-subagents-50

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter, with any `measured:` command naming no ref of its own.

**Goal:** `check-panel-fix-single-dispatch.sh` accepts a chunk row's fix commit keyed to its round.

**Architecture:** `pfdCheckCommits` gains one accepted trailer, `pfdRoundKey(base)`.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; a diff that does not apply is reported, never hand-merged by guesswork.
- No new inline suppressions; never stop flowd or flow-postgres.

## Review Focus

- A chunk row whose commit carries another round's key is still refused (`accept-round-key`).

- [ ] 1. Accept a chunk row's fix commit keyed to its round
  - [ ] **Step 1: Reproduce.** Apply the test hunk alone
    (`git apply --include='stats/internal/guard/*_test.go' /tmp/pf50-task1.diff`, the diff below saved
    there) and run `cd stats && go test ./internal/guard -run ChunkRowKeyedToItsRound`: FAIL, stderr
    `panel-fix-1-2 records commit … carries no Flow-Key: panel-fix-1-2 trailer`.
  - [ ] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/internal/guard/*_test.go' /tmp/pf50-task1.diff`):

````diff verified:applied with git apply to this worktree at 5302fcb2; go test ./internal/guard -run 'TestCheckPanelFixSingleDispatch|TestPfd' passes with it and the new test fails with panelfixsingledispatch.go reverted; gofmt -l, go vet and check-references exit clean
diff --git a/scripts/check-panel-fix-single-dispatch.sh b/scripts/check-panel-fix-single-dispatch.sh
index 16346883..584d669c 100755
--- a/scripts/check-panel-fix-single-dispatch.sh
+++ b/scripts/check-panel-fix-single-dispatch.sh
@@ -61,7 +61,9 @@
 # and each panel-fix row whose end recorded a -commit must name a commit
 # whose subject is some task's **Commit:** subject in the change's tasks.md
 # (the fold route) or which carries a `Flow-Key: <the row's key>` trailer,
-# -retry stripped (the pushed route's fix commit) -- review-panel.md's
+# -retry stripped, or its round's key `panel-fix-<r>` for a chunk row
+# `panel-fix-<r>-<n>` (the pushed route's fix commit, keyed to its round as
+# review-panel-fix-round.md states) -- review-panel.md's
 # -commit contract (kan-952 self-review: panel-fix-0 recorded the plan
 # commit while its fix landed as its own commit). The commit is looked up
 # in the worktree, then in each peer the spec root's peers file names; a
diff --git a/stats/internal/guard/check_panel_fix_single_dispatch_test.go b/stats/internal/guard/check_panel_fix_single_dispatch_test.go
index e7036a36..878f9f79 100644
--- a/stats/internal/guard/check_panel_fix_single_dispatch_test.go
+++ b/stats/internal/guard/check_panel_fix_single_dispatch_test.go
@@ -418,6 +418,33 @@ func TestCheckPanelFixSingleDispatchAcceptsATrailerOneRoundLow(t *testing.T) {
 	pfdExpect(t, pfdCase{dispatches: row("panel-fix-5", head("panel-fix-3")), args: args}, 1, "carries no Flow-Key: panel-fix-5 trailer")
 }
 
+// TestCheckPanelFixSingleDispatchAcceptsAChunkRowKeyedToItsRound pins that a
+// chunk row panel-fix-<r>-<n> whose fix commit carries the round's key
+// panel-fix-<r> passes, as review-panel-fix-round.md keys a fix commit to its
+// round; another round's key is still refused (kan-1062 run, panel-fix-1-2).
+func TestCheckPanelFixSingleDispatchAcceptsAChunkRowKeyedToItsRound(t *testing.T) {
+	t.Parallel()
+	wt := t.TempDir()
+	gitRun(t, wt, "init", "-q")
+	writeFile(t, filepath.Join(wt, "spectre/changes/demo/tasks.md"),
+		"# demo\n\n- [ ] 1. Add alpha\n\n**Commit:** feat(alpha): add alpha\n")
+	head := func(trailer string) string {
+		gitRun(t, wt, "commit", "-q", "--allow-empty", "-m", "fix(alpha): review findings", "-m", "Flow-Key: "+trailer)
+		out, err := exec.Command(fixtureGit, "-C", wt, "rev-parse", "HEAD").Output()
+		if err != nil {
+			t.Fatal(err)
+		}
+		return strings.TrimSpace(string(out))
+	}
+	chunked := func(sha string) []byte {
+		return []byte(`[{"key":"panel-fix-1","role":"panel-fix","sessionToken":"mf-tok","commitSha":"` + head("panel-fix-1") +
+			`"},{"key":"panel-fix-1-2","role":"panel-fix","sessionToken":"mf-tok","commitSha":"` + sha + `"}]`)
+	}
+	args := []string{wt, "demo", "mf-tok"}
+	pfdExpect(t, pfdCase{dispatches: chunked(head("panel-fix-1")), findings: pfdFindings(11, 0), args: args}, 0, "OK")
+	pfdExpect(t, pfdCase{dispatches: chunked(head("panel-fix-2")), findings: pfdFindings(11, 0), args: args}, 1, "carries no Flow-Key: panel-fix-1-2 trailer")
+}
+
 // TestPfdPriorFindingsSkipsGatedReviewerRows pins that a gated per-task
 // reviewer's round-0 rows do not raise a fix round's chunk bound.
 func TestPfdPriorFindingsSkipsGatedReviewerRows(t *testing.T) {
diff --git a/stats/internal/guard/panelfixsingledispatch.go b/stats/internal/guard/panelfixsingledispatch.go
index 4ed3abe8..86fb1c4e 100644
--- a/stats/internal/guard/panelfixsingledispatch.go
+++ b/stats/internal/guard/panelfixsingledispatch.go
@@ -339,7 +339,9 @@ func pfdPanelFixRows(raw []byte, token string) (keys, foreign []pfdRow, ok bool)
 
 // pfdCheckCommits names each row whose recorded commit carries neither a
 // task's Commit: subject nor a Flow-Key: trailer naming the row's key (its
-// -retry stripped) or that key one round low, looked up in the worktree and
+// -retry stripped), its round's key (a chunk row panel-fix-<r>-<n>'s
+// -<n> stripped -- review-panel-fix-round.md's fix commit names its round,
+// kan-1062 run's panel-fix-1-2) or that key one round low, looked up in the worktree and
 // then in each peer its spec root's peers file names. The one-round-low
 // trailer is the row signal eleven of check-unfinished-work back-fills under
 // the real round's key for a commit keyed a round low (kan-1059 self-review,
@@ -390,13 +392,24 @@ func pfdCheckCommits(env Env, worktree, name string, fixRows []pfdRow, stderr io
 		}
 		base, _ := strings.CutSuffix(r.key, "-retry")
 		low := pfdOneRoundLow(base)
-		if !subjects[meta[0]] && !slices.Contains(meta[1:], base) && !slices.Contains(meta[1:], r.key) && (low == "" || !slices.Contains(meta[1:], low)) {
+		round := pfdRoundKey(base)
+		if !subjects[meta[0]] && !slices.Contains(meta[1:], base) && !slices.Contains(meta[1:], r.key) && !slices.Contains(meta[1:], round) && (low == "" || !slices.Contains(meta[1:], low)) {
 			violations = append(violations, fmt.Sprintf("%s records commit %s (%q), which is no task's Commit: subject and carries no Flow-Key: %s trailer -- -commit is the task commit the fixup was folded into, or the fix commit itself", r.key, r.commit, meta[0], base))
 		}
 	}
 	return violations, true
 }
 
+// pfdRoundKey maps panel-fix-<r>-<n> to panel-fix-<r>; any other key to itself.
+func pfdRoundKey(key string) string {
+	rest, ok := strings.CutPrefix(key, "panel-fix-")
+	if !ok {
+		return key
+	}
+	round, _, _ := strings.Cut(rest, "-")
+	return "panel-fix-" + round
+}
+
 // pfdOneRoundLow maps panel-fix-<r>[-<n>] to panel-fix-<r-1>[-<n>]; "" when
 // the key is not that shape or r is 0.
 func pfdOneRoundLow(key string) string {
````
  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/guard -run 'TestCheckPanelFixSingleDispatch|TestPfd'`
    passes; `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean;
    `scripts/check-references.sh` exits 0.
  - [ ] **Step 4: Commit** the three files with the subject below.

**Build:** green
**Files:** `scripts/check-panel-fix-single-dispatch.sh`, `stats/internal/guard/check_panel_fix_single_dispatch_test.go`, `stats/internal/guard/panelfixsingledispatch.go`
**Tests:** `TestCheckPanelFixSingleDispatchAcceptsAChunkRowKeyedToItsRound`
**Regression:** TestCheckPanelFixSingleDispatchAcceptsAChunkRowKeyedToItsRound — reverting the commit refuses a chunk row `panel-fix-1-2` whose commit carries `Flow-Key: panel-fix-1`.
**Baseline:** before=5 after=6
<!-- measured: grep -c '^func Test' stats/internal/guard/check_panel_fix_single_dispatch_test.go @ 5302fcb2 → 5 -->
**Commit:** fix(guard): accept a chunk row's fix commit keyed to its round
**After:** none

**Decision:** accept-round-key
