# self-review-kan-urgent-5

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-18-self-review-kan-urgent-5-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; the commit body may name the task's Jira key. Never write `spectre/changes/`. Never merge or push.
- Each task's change is the unified diff its Step 2 gives, applied with `git apply` from this worktree's root, in task order — every diff was applied and checked there (`design.md`, **Context**). Save a diff by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork.
- Run `cd stats && gofmt -w .` before committing any Go file; `gofmt -l .` and `go vet ./internal/guard/` stay clean.
- KAN-984 is disproved and carries no task (`design.md`, **Context**).

## Review focus

- A path with no symlink in it must judge exactly as before; a not-yet-existing worktree under `<repo>-worktrees/` must still be let through (Task 1).
- A single-commit task's subject mismatch must still print `commit subject … does not match declared Commit: …` (Task 2).
- A `Task-Id:` trailer an earlier change left on the base branch must never join a task's verdict, and an explicit fourth argument must still pin one range (Task 3).
- A fresh worktree's first `## test` or `## lint` line that compiles `stats/internal/web` must build the SPA, and a worktree that already has it must not rebuild (Task 4).


- [ ] 1. Judge a protected write path after its symlinks resolve
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Apply Step 2's diff to `scripts/test-protect-main-checkout.sh` alone (`git apply --include='scripts/test-protect-main-checkout.sh' <the saved diff>`) and run `bash scripts/test-protect-main-checkout.sh`: case 102 fails `wanted deny, got allow` — the defect. Keep the applied hunk; Step 2 applies the rest with `git apply --exclude='scripts/test-protect-main-checkout.sh' <the saved diff>`.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at ebc10e49; with it applied scripts/test-protect-main-checkout.sh prints all protect-main-checkout cases pass, and with the hook hunk reverted case 102 fails
diff --git a/hooks/protect-main-checkout.py b/hooks/protect-main-checkout.py
index cca7d39d..a2eda713 100755
--- a/hooks/protect-main-checkout.py
+++ b/hooks/protect-main-checkout.py
@@ -149,7 +149,10 @@ def worktree_root(d):
 
 
 def protected(path):
-    """(toplevel, branch) when path lies in a main checkout on its default branch, else None."""
+    """(toplevel, branch) when path lies in a main checkout on its default branch, else None.
+    path is judged after its symlinks resolve, as the write lands: `abspath` would collapse
+    `<link>/..` lexically, and a link into a checkout's subdirectory would hide the checkout."""
+    path = os.path.realpath(path)
     d = existing_dir(path)
     if d is None:
         return None
diff --git a/scripts/test-protect-main-checkout.sh b/scripts/test-protect-main-checkout.sh
index 114b7510..0a77aa55 100755
--- a/scripts/test-protect-main-checkout.sh
+++ b/scripts/test-protect-main-checkout.sh
@@ -103,6 +103,7 @@
 #   99. `git -C <wt> push --all $R` (--all pushes the default branch) -> deny
 #  100. `git ${OPTS} push origin HEAD` (expanded git option)         -> deny
 #  101. `$GIT push origin HEAD` (expanded git command word)          -> deny
+#  102. Bash `cat > <link into main/sub>/../f.txt` from elsewhere    -> deny (kan-963: judged after symlinks resolve)
 set -euo pipefail
 SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
 HOOK="$SCRIPT_DIR/../hooks/protect-main-checkout.py"
@@ -483,6 +484,10 @@ expect "100 an expanded git option hides the repository" deny \
 expect "101 an expanded git command word" deny \
   "$(run "$ROOT" Bash "{\"command\":$(q "\$GIT push origin HEAD")}")"
 unset LOOP_LANDING_CHECK
+mkdir -p "$MAIN/sub"
+ln -s "$MAIN/sub" "$ROOT/elsewhere/lsub"
+expect "102 a symlink into the checkout, then .., resolves before judging" deny \
+  "$(run "$ROOT" Bash "{\"command\":$(q "cat > $ROOT/elsewhere/lsub/../f.txt")}")"
 
 # 3 last: moving the main checkout off main lifts the protection
 git_q -C "$MAIN" checkout -q -b feature
````

  - [ ] **Step 3: Verify.** `bash scripts/test-protect-main-checkout.sh` prints `all protect-main-checkout cases pass`; `scripts/check-python-suppressions.sh` exits 0. Commit.

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `scripts/test-protect-main-checkout.sh`
**Tests:** `102 a symlink into the checkout, then .., resolves before judging`
**Regression:** reverting `protected()`'s `realpath` lets `cat > <link into main/sub>/../f.txt` through, so case 102 fails `wanted deny, got allow`.
**Baseline:** before=89 after=90
<!-- measured: grep -c '^expect ' scripts/test-protect-main-checkout.sh @ ebc10e49 (before; after with this task's diff applied) -->
**Commit:** fix(hooks): judge a protected write path after its symlinks resolve
**After:** none

**Decision:** protect-realpath

- [ ] 2. Let a cross-repo task's commits each name their repository's scope
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Apply Step 2's diff to the test file alone (`git apply --include='stats/internal/guard/check_task_commit_fields_test.go' <the saved diff>`) and run `cd stats && go test ./internal/guard/ -run 'TestCheckTaskCommitFields/case_158' -count=1`: `case 158: a cross-repo task's second commit may name its own scope` fails — the defect. Keep the applied hunks; Step 2 applies the rest with `git apply --exclude='stats/internal/guard/check_task_commit_fields_test.go' <the saved diff>`.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at ebc10e49 after Task 1; with it applied go test ./internal/guard/ passes, and with taskcommitfields.go reverted case 158 fails
diff --git a/scripts/check-task-commit-fields.sh b/scripts/check-task-commit-fields.sh
index b23e4683..52bb9b9f 100755
--- a/scripts/check-task-commit-fields.sh
+++ b/scripts/check-task-commit-fields.sh
@@ -30,8 +30,10 @@
 # is refused. The verdict merges across the listed repositories: the union
 # of changed paths against `Files:`/`Allowed-collateral:`, the concatenated
 # diffs against `Tests:`, the summed `@Test` counts against `Baseline:`,
-# the declared `Commit:` subject required on every listed commit, and the
-# tree check satisfied by any listed tree. `**Baseline:**`'s one recorded
+# the declared `Commit:` subject on one listed commit exactly and on one
+# commit of every listed repository with its scope free — each repository's
+# commit may name its own module (KAN-972) — and the tree check satisfied by
+# any listed tree. `**Baseline:**`'s one recorded
 # measurement command addresses a single repository, so with more than one
 # pair that check skips — never a verdict (the kan-100 rule). One pair
 # behaves exactly as the single-sha form; a real revision never contains
diff --git a/stats/internal/guard/check_task_commit_fields_test.go b/stats/internal/guard/check_task_commit_fields_test.go
index 967b24eb..3b3472c1 100644
--- a/stats/internal/guard/check_task_commit_fields_test.go
+++ b/stats/internal/guard/check_task_commit_fields_test.go
@@ -1325,6 +1325,13 @@ After the change: 186 tests
 **Tests:** ¤test_alpha¤
 **Commit:** add alpha for real
 **Build:** green
+`),
+	"158": bt(`- [ ] 1. A task whose commits name each repository's module
+
+**Files:** ¤alpha.txt¤, ¤beta.txt¤
+**Tests:** ¤test_alpha¤
+**Commit:** feat(alpha): add alpha for real
+**Build:** green
 `),
 	"153": bt(`- [ ] 1. A declared test only a removal hunk carries
 
@@ -2985,6 +2992,28 @@ func TestCheckTaskCommitFields(t *testing.T) {
 			res := r.run("157", sha)
 			ok(t, "case 157: the repair form declares no diff-demanded names, so a name absent from the diff passes", res.rc == 0, res)
 		}},
+		{"case 158", func(t *testing.T) {
+			// KAN-972: each repository's commit names its own module.
+			commit := func(b1Subject, b2Subject string) tcfRes {
+				b1, b2, a := tcfMapRepos(t, "158")
+				writeFile(t, b1+"/alpha.txt", "def test_alpha(): pass\n")
+				tcfGit(t, b1, "add", "alpha.txt")
+				tcfGit(t, b1, "commit", "-q", "-m", b1Subject)
+				writeFile(t, b2+"/beta.txt", "b\n")
+				tcfGit(t, b2, "add", "beta.txt")
+				tcfGit(t, b2, "commit", "-q", "-m", b2Subject)
+				sha1, sha2 := tcfGit(t, b1, "rev-parse", "HEAD"), tcfGit(t, b2, "rev-parse", "HEAD")
+				return fx.run(b1, "1", b1+"="+sha1+","+b2+"="+sha2, "", a, "x-repo-change")
+			}
+			res := commit("feat(alpha): add alpha for real", "feat(beta): add alpha for real")
+			ok(t, "case 158: a cross-repo task's second commit may name its own scope", res.rc == 0, res)
+			res = commit("feat(gamma): add alpha for real", "feat(beta): add alpha for real")
+			ok(t, "case 158: a cross-repo task with no commit carrying the declared scope fails", res.rc == 1, res)
+			ok(t, "case 158: names the exact-subject rule", has(res.out, "no listed commit carries the declared Commit: 'feat(alpha): add alpha for real' exactly"), res)
+			res = commit("feat(alpha): add alpha for real", "fix(beta): add alpha for real")
+			ok(t, "case 158: a second commit of another type still fails", res.rc == 1, res)
+			ok(t, "case 158: names that commit's subject", has(res.out, "commit subject 'fix(beta): add alpha for real' does not match"), res)
+		}},
 	}
 	for _, c := range cases {
 		t.Run(c.name, func(t *testing.T) {
diff --git a/stats/internal/guard/taskcommitfields.go b/stats/internal/guard/taskcommitfields.go
index ad72e2a7..6872ab44 100644
--- a/stats/internal/guard/taskcommitfields.go
+++ b/stats/internal/guard/taskcommitfields.go
@@ -1340,9 +1340,9 @@ func tcfRunGit(env Env, worktree string, args ...string) (string, error) {
 // the task's commits — and merges the verdict across them: the union of
 // changed paths against Files:/Allowed-collateral:, the concatenated diffs
 // against Tests:, the summed @Test counts against Baseline:, the declared
-// Commit: subject required on every listed commit, and the tree check
-// satisfied by any listed tree. One pair behaves exactly as the
-// single-repository form always did.
+// Commit: subject per tcfCheckSubjects, and the tree check satisfied by any
+// listed tree. One pair behaves exactly as the single-repository form always
+// did.
 //
 // tcfRepoDiff is one repository's slice of that verdict: the pair the
 // commit map (or the single-commit form) named, its derived parent, and
@@ -1474,9 +1474,7 @@ func tcfCheckTaskCommit(env Env, tasksMD, taskID string, repos []tcfRepoCommit,
 	for _, name := range missingAll {
 		v = append(v, fmt.Sprintf("task %s: declared test %s not found in the tree at %s — %s", task.id, name, strings.Join(shas, ", "), tcfTestsParseRule))
 	}
-	for _, d := range diffs {
-		v = append(v, tcfCheckCommitSubject(folded, tcfStrip(d.subject))...)
-	}
+	v = append(v, tcfCheckSubjects(folded, diffs)...)
 	v = append(v, tcfCheckCommitScope(task, changeName)...)
 	sibling, err := tcfCheckScopeSibling(env, task, diffs, projectDir)
 	if err != nil {
@@ -2091,6 +2089,51 @@ func tcfCheckCommitSubject(task tcfTask, actual string) []string {
 	return []string{fmt.Sprintf("task %s: commit subject %s does not match declared Commit: %s", task.id, tcfRepr(actual), tcfRepr(*task.commit))}
 }
 
+// tcfCheckSubjects holds the listed commits to the declared Commit: subject:
+// one of them carries it exactly, and every repository listed carries it on
+// one of its commits with the scope free, since one task's commits in two
+// repositories each name their own repository's module (KAN-972: KAN-924
+// Task 29's frontend and playwright commits had to share one subject, so it
+// dropped its scope). A repository with none is reported on its first
+// commit, in tcfCheckCommitSubject's words.
+func tcfCheckSubjects(task tcfTask, diffs []tcfRepoDiff) []string {
+	if task.commit == nil {
+		return nil
+	}
+	var v []string
+	exact := false
+	var seen []string
+	for _, d := range diffs {
+		exact = exact || tcfStrip(d.subject) == *task.commit
+		if slices.Contains(seen, d.worktree) {
+			continue
+		}
+		seen = append(seen, d.worktree)
+		if !slices.ContainsFunc(diffs, func(o tcfRepoDiff) bool {
+			return o.worktree == d.worktree && tcfUnscoped(tcfStrip(o.subject)) == tcfUnscoped(*task.commit)
+		}) {
+			v = append(v, tcfCheckCommitSubject(task, tcfStrip(d.subject))...)
+		}
+	}
+	if exact || v != nil {
+		return v
+	}
+	if len(seen) == 1 {
+		return tcfCheckCommitSubject(task, tcfStrip(diffs[0].subject))
+	}
+	return []string{fmt.Sprintf("task %s: no listed commit carries the declared Commit: %s exactly — one repository's commit names the declared scope, another's may name its own", task.id, tcfRepr(*task.commit))}
+}
+
+// tcfUnscoped is subject without its `(<scope>)`, `feat(a)!: x` read as
+// `feat!: x`; a subject with no scope is returned as it is.
+func tcfUnscoped(subject string) string {
+	m := tcfSubjectScopeRE.FindStringSubmatchIndex(subject)
+	if m == nil {
+		return subject
+	}
+	return subject[:m[2]-1] + subject[m[3]+1:]
+}
+
 // Moved from check-task-commit-fields.py's SUBJECT_SCOPE_RE (KAN-760 deleted the
 // Python copy once this port replaced it; panel round 0, F9):
 //
````

  - [ ] **Step 3: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckTaskCommitFields|TestCloseTask' -count=1` passes with `gofmt -l .` printing nothing. Commit.

**Build:** green
**Files:** `stats/internal/guard/taskcommitfields.go`, `stats/internal/guard/check_task_commit_fields_test.go`, `scripts/check-task-commit-fields.sh`
**Tests:** `case 158: a cross-repo task's second commit may name its own scope`
**Regression:** reverting `taskcommitfields.go` requires the declared subject on every listed commit again, so case 158's `feat(beta)` commit fails the task.
**Baseline:** before=143 after=144
<!-- measured: grep -c '{"case [0-9]' stats/internal/guard/check_task_commit_fields_test.go @ ebc10e49 (before; after with this task's diff applied) -->
**Commit:** fix(guard): let a cross-repo task's commits each name their repository's scope
**After:** none

**Decision:** subject-scope-per-repository

- [ ] 3. Judge a task over every commit carrying its Task-Id trailer
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Apply Step 2's diff to the test file alone (`git apply --include='stats/internal/guard/check_task_commit_fields_test.go' <the saved diff>`) and run `cd stats && go test ./internal/guard/ -run 'TestCheckTaskCommitFields/case_159' -count=1`: case 159 fails naming `alpha.txt is declared in Files: but the commit does not touch it` — the defect. Keep the applied hunks; Step 2 applies the rest with `git apply --exclude='stats/internal/guard/check_task_commit_fields_test.go' <the saved diff>`.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at ebc10e49 after Tasks 1-2; with it applied go test ./internal/guard/ passes, with taskcommitfields.go at Task 2's state case 159 fails, and with the merge-base bound removed case 159 fails naming gamma.txt
diff --git a/scripts/check-task-commit-fields.sh b/scripts/check-task-commit-fields.sh
index 52bb9b9f..430e1ec2 100755
--- a/scripts/check-task-commit-fields.sh
+++ b/scripts/check-task-commit-fields.sh
@@ -34,8 +34,17 @@
 # commit of every listed repository with its scope free — each repository's
 # commit may name its own module (KAN-972) — and the tree check satisfied by
 # any listed tree. `**Baseline:**`'s one recorded
-# measurement command addresses a single repository, so with more than one
-# pair that check skips — never a verdict (the kan-100 rule). One pair
+# measurement command addresses a single commit's range, so with more than
+# one commit that check skips — never a verdict (the kan-100 rule).
+#
+# EVERY COMMIT CARRYING THE TASK'S TRAILER (KAN-976). Without a fourth
+# argument, every other commit from a listed repository's merge base to its
+# listed commit whose `Task-Id:` trailer names the task joins the verdict as
+# one more commit of that repository — a gated review's on-top fix commit
+# carries its task's trailer, and judged alone it touched none of the task's
+# declared files. The merge base is HEAD's with the branch's recorded
+# `flowBase`, else origin/HEAD, else origin/main, else main; with none of
+# them resolvable the listed commit stands alone. One pair
 # behaves exactly as the single-sha form; a real revision never contains
 # "=", which is what tells the two forms apart.
 #
diff --git a/stats/internal/guard/check_task_commit_fields_test.go b/stats/internal/guard/check_task_commit_fields_test.go
index 3b3472c1..4ddd1167 100644
--- a/stats/internal/guard/check_task_commit_fields_test.go
+++ b/stats/internal/guard/check_task_commit_fields_test.go
@@ -1332,6 +1332,13 @@ After the change: 186 tests
 **Tests:** ¤test_alpha¤
 **Commit:** feat(alpha): add alpha for real
 **Build:** green
+`),
+	"159": bt(`- [ ] 1. A task finished by an on-top fix commit
+
+**Files:** ¤alpha.txt¤, ¤beta.txt¤
+**Tests:** ¤test_alpha¤
+**Commit:** add alpha
+**Build:** green
 `),
 	"153": bt(`- [ ] 1. A declared test only a removal hunk carries
 
@@ -2992,6 +2999,21 @@ func TestCheckTaskCommitFields(t *testing.T) {
 			res := r.run("157", sha)
 			ok(t, "case 157: the repair form declares no diff-demanded names, so a name absent from the diff passes", res.rc == 0, res)
 		}},
+		{"case 159", func(t *testing.T) {
+			// KAN-976: the task's commit, another task's, then a gated
+			// review's on-top fix carrying the same Task-Id: trailer; a
+			// trailer below the merge base is never read.
+			r := fx.repo(t, "change-a")
+			r.plan(tcfPlans["159"])
+			r.git("branch", "-M", "main")
+			r.commit("old change\n\nTask-Id: 1", "gamma.txt", "g\n")
+			r.git("checkout", "-q", "-b", "spectre/change-a")
+			r.commit("add alpha\n\nTask-Id: 1", "alpha.txt", "def test_alpha(): pass\n")
+			r.commit("other task\n\nTask-Id: 2", "delta.txt", "d\n")
+			fix := r.commit("fix review finding\n\nTask-Id: 1", "beta.txt", "b\n")
+			res := r.run("1", fix)
+			ok(t, "case 159: an on-top fix commit is judged with its task's commit", res.rc == 0, res)
+		}},
 		{"case 158", func(t *testing.T) {
 			// KAN-972: each repository's commit names its own module.
 			commit := func(b1Subject, b2Subject string) tcfRes {
diff --git a/stats/internal/guard/taskcommitfields.go b/stats/internal/guard/taskcommitfields.go
index 6872ab44..f377039d 100644
--- a/stats/internal/guard/taskcommitfields.go
+++ b/stats/internal/guard/taskcommitfields.go
@@ -1378,6 +1378,26 @@ func tcfCheckTaskCommit(env Env, tasksMD, taskID string, repos []tcfRepoCommit,
 		return squashViolations, nil
 	}
 
+	// Every other commit of a listed repository's branch carrying the task's
+	// Task-Id: trailer joins the verdict as one more pair (KAN-976): a gated
+	// review's on-top fix commit carries its task's trailer, and judged
+	// alone it touched none of the task's declared files. An explicit parent
+	// pins the one range the caller means and adds nothing.
+	if parent == "" {
+		var all []tcfRepoCommit
+		for _, repo := range repos {
+			more, err := tcfTaskIDCommits(env, repo.worktree, repo.commit, taskID)
+			if err != nil {
+				return nil, err
+			}
+			all = append(all, repo)
+			for _, sha := range more {
+				all = append(all, tcfRepoCommit{repo.worktree, sha})
+			}
+		}
+		repos = all
+	}
+
 	diffs := make([]tcfRepoDiff, 0, len(repos))
 	for _, repo := range repos {
 		d := tcfRepoDiff{worktree: repo.worktree, commit: repo.commit}
@@ -1439,8 +1459,8 @@ func tcfCheckTaskCommit(env Env, tasksMD, taskID string, repos []tcfRepoCommit,
 	}
 	v = append(v, counts...)
 	if len(diffs) == 1 {
-		// The recorded measurement command addresses one repository; with
-		// more than one pair there is no single tree it measures, so the
+		// The recorded measurement command addresses one commit's range; with
+		// more than one commit there is no single range it measures, so the
 		// check skips — never a verdict (the kan-100 rule).
 		measured, err := tcfCheckBaselineMeasured(env, task, diffs[0].worktree, diffs[0].parent, diffs[0].commit, tcfMeasuredTimeout)
 		if err != nil {
@@ -1484,6 +1504,36 @@ func tcfCheckTaskCommit(env Env, tasksMD, taskID string, repos []tcfRepoCommit,
 	return v, nil
 }
 
+// tcfTaskIDCommits is every commit from worktree's base to commit, commit
+// itself aside, whose Task-Id: trailer names taskID, newest first. The base
+// is vmDefaultBase's merge base, so a trailer an earlier change left on the
+// base branch is never read; a worktree with no resolvable base adds none.
+func tcfTaskIDCommits(env Env, worktree, commit, taskID string) ([]string, error) {
+	base, _, ok := vmDefaultBase(envGit(env), worktree)
+	if !ok {
+		return nil, nil
+	}
+	self, err := tcfRunGit(env, worktree, "rev-parse", commit+"^{commit}")
+	if err != nil {
+		return nil, err
+	}
+	out, err := tcfRunGit(env, worktree, "log", "--no-merges", "--format=%H %(trailers:key=Task-Id,valueonly,separator=%x2C)", base+".."+commit)
+	if err != nil {
+		return nil, err
+	}
+	var shas []string
+	for _, line := range tcfSplitLines(out) {
+		sha, ids, _ := strings.Cut(line, " ")
+		if sha == "" || sha == tcfStrip(self) {
+			continue
+		}
+		if slices.ContainsFunc(strings.Split(ids, ","), func(id string) bool { return strings.TrimSpace(id) == taskID }) {
+			shas = append(shas, sha)
+		}
+	}
+	return shas, nil
+}
+
 // tcfCheckScopeSibling refuses a declared scope that names a module of
 // another repository of the project and nothing in this one — kan-982's
 // `fix(group)` on a commit touching only gymie-playwright's
````

  - [ ] **Step 3: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckTaskCommitFields|TestCloseTask' -count=1` passes with `gofmt -l .` printing nothing. Commit.

**Build:** green
**Files:** `stats/internal/guard/taskcommitfields.go`, `stats/internal/guard/check_task_commit_fields_test.go`, `scripts/check-task-commit-fields.sh`
**Tests:** `case 159: an on-top fix commit is judged with its task's commit`
**Regression:** reverting `taskcommitfields.go` judges case 159's fix commit alone, so it fails on `alpha.txt`, `test_alpha` and the subject.
**Baseline:** before=144 after=145
<!-- measured: grep -c '{"case [0-9]' stats/internal/guard/check_task_commit_fields_test.go @ ebc10e49 (before with Task 2's diff applied; after with this task's diff applied too) -->
**Commit:** fix(guard): judge a task over every commit carrying its Task-Id trailer
**After:** Task 2

**Decision:** task-id-commit-union

- [ ] 4. Build the SPA only when a test or lint line needs it
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Apply Step 2's diff to `scripts/test-make-build.sh` alone (`git apply --include='scripts/test-make-build.sh' <the saved diff>`) and run `bash scripts/test-make-build.sh`: `web-dist builds the SPA only when it is absent` fails — the defect. Keep the applied hunk; Step 2 applies the rest with `git apply --exclude='scripts/test-make-build.sh' <the saved diff>`.
  - [ ] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply at ebc10e49 after Tasks 1-3; with it applied the whole lint list of .flow/project.md exits 0 and scripts/run-guard-tests.sh passes, and with stats/Makefile reverted the web-dist case fails
diff --git a/.flow/project-rationale.md b/.flow/project-rationale.md
index 8b43b4e3..6d7de863 100644
--- a/.flow/project-rationale.md
+++ b/.flow/project-rationale.md
@@ -24,11 +24,14 @@ database before anyone noticed.
 `not by a number pasted here` — which is the point:
 a written duration went stale twice before this paragraph stopped carrying one.
 
-## .flow/project.md — worktree setup
+## .flow/project.md — test: `make -s web-dist`
 
 kan-389's verifier hit exactly that, and its conductor then ran
 the whole `## test` list itself to get past it — a fresh-worktree fact, not a branch defect, and
-the reason this key exists.
+the reason a `## worktree setup` running `make web-build` once existed. That setup ran `npm ci` and
+the SPA build on every creating run, a docs-only change's included, though `go list -deps
+./cmd/flow ./cmd/flow-guard` names no `internal/web` (KAN-987); the lines that need the build now
+make it lazily instead.
 
 ## .flow/project.md — lint
 
diff --git a/.flow/project.md b/.flow/project.md
index 84aabfc2..b42c07cb 100644
--- a/.flow/project.md
+++ b/.flow/project.md
@@ -69,10 +69,15 @@ forbids any agent action touching it at all.
 
 ```bash
 scripts/run-guard-tests.sh
-cd stats && go test ./... -race -count=1
-cd stats/web && npm test
+cd stats && make -s web-dist && go test ./... -race -count=1
+cd stats && make -s web-dist && cd web && npm test
 ```
 
+`stats/internal/web/dist/` is gitignored and `stats/internal/web/embed.go`'s `//go:embed all:dist`
+refuses to compile without it, so `make -s web-dist` (`stats/Makefile`) opens every line here and in
+`## lint` that compiles `internal/web` or needs `stats/web/node_modules`: it runs `make web-build`,
+`npm ci && npm run build` in `stats/web`, only when the build is absent.
+
 `scripts/run-guard-tests.sh` discovers every `scripts/test-*.sh` by glob and runs them concurrently
 through `scripts/lib/parallel.sh`; a new harness added to `scripts/` is picked up automatically, with
 no edit needed here. It points `FLOW_GUARD_CACHE_DIR` at its own temp directory, so a ported
@@ -100,17 +105,6 @@ here since it is the one guard in this repository paying for a subprocess rather
 scan.
 <!-- measured: time scripts/check-installed-citations.sh >/dev/null @ branch kan-102-citations-resolve-to-installed-paths -->
 
-## worktree setup
-
-```bash
-cd stats && make web-build
-```
-
-`stats/internal/web/dist/` is gitignored and `stats/internal/web/embed.go`'s `//go:embed all:dist`
-refuses to compile without it, so a fresh worktree's first `go test ./...` or `go build ./...`
-fails until the SPA is built once. `make web-build` is `npm ci && npm run build` in `stats/web`
-(`stats/Makefile`), the same prerequisite `make test` and `make build` already carry.
-
 ## lint
 
 ```bash
@@ -140,8 +134,8 @@ scripts/check-verbatim-moves.sh
 scripts/check-ask-silence.sh
 scripts/check-worktree-location.sh "$(git worktree list --porcelain | awk '/^worktree /{print substr($0,10); exit}')"
 cd stats && gofmt -l .
-cd stats && go vet ./...
-cd stats/web && npx tsc -b
+cd stats && make -s web-dist && go vet ./...
+cd stats && make -s web-dist && cd web && npx tsc -b
 ```
 
 **There is no auto-fix command for the guard scripts** (`scripts/check-*`) — every one of them
diff --git a/AGENTS.md b/AGENTS.md
index ef64a4dc..de8d4fd9 100644
--- a/AGENTS.md
+++ b/AGENTS.md
@@ -43,8 +43,8 @@ cd stats && gofmt -w .                                  # auto-fix, Go source on
                                                           # this repository has an auto-fix command
 scripts/check-references.sh                               # plus every other scripts/check-*.sh guard
                                                           # named in <project>/.flow/project.md's `## lint`
-cd stats && go vet ./... && gofmt -l .                   # must exit clean before claiming Go work done
-cd stats/web && npx tsc -b                                # must exit clean before claiming SPA work done
+cd stats && make -s web-dist && go vet ./... && gofmt -l .  # must exit clean before claiming Go work done
+cd stats && make -s web-dist && cd web && npx tsc -b      # must exit clean before claiming SPA work done
 ```
 
 ---
diff --git a/CLAUDE.md b/CLAUDE.md
index 57147b43..b853f028 100644
--- a/CLAUDE.md
+++ b/CLAUDE.md
@@ -20,8 +20,8 @@ cd stats && gofmt -w .                                  # auto-fix, Go source on
                                                           # this repository has an auto-fix command
 scripts/check-references.sh                               # plus every other scripts/check-*.sh guard
                                                           # named in <project>/.flow/project.md's `## lint`
-cd stats && go vet ./... && gofmt -l .                   # must exit clean before claiming Go work done
-cd stats/web && npx tsc -b                                # must exit clean before claiming SPA work done
+cd stats && make -s web-dist && go vet ./... && gofmt -l .  # must exit clean before claiming Go work done
+cd stats && make -s web-dist && cd web && npx tsc -b      # must exit clean before claiming SPA work done
 ```
 
 ---
diff --git a/scripts/test-make-build.sh b/scripts/test-make-build.sh
index b35fdda6..33f38a08 100755
--- a/scripts/test-make-build.sh
+++ b/scripts/test-make-build.sh
@@ -282,6 +282,27 @@ else
   pass "restart depends on build to produce \$(LIVE_BIN), and does not rebuild it"
 fi
 
+# web-dist builds the SPA only when internal/web/dist/index.html is absent
+# (KAN-987). Run under `make -n` over a copy of the Makefile, so the case
+# reads which recipe would run without running npm.
+WEB_DIST_DIR="$(mktemp -d "${TMPDIR:-/tmp}/test-make-build-web-dist.XXXXXX")"
+cp "$STATS_DIR/Makefile" "$WEB_DIST_DIR/Makefile"
+mkdir -p "$WEB_DIST_DIR/internal/web/dist"
+touch "$WEB_DIST_DIR/internal/web/dist/index.html"
+WEB_DIST_BUILT="$(make -s -n -C "$WEB_DIST_DIR" web-dist 2>&1 || true)"
+rm "$WEB_DIST_DIR/internal/web/dist/index.html"
+WEB_DIST_MISSING="$(make -s -n -C "$WEB_DIST_DIR" web-dist 2>&1 || true)"
+rm -rf "$WEB_DIST_DIR"
+if grep -q 'npm ci' <<<"$WEB_DIST_BUILT"; then
+  fail "web-dist builds the SPA only when it is absent: with index.html present it runs:
+$WEB_DIST_BUILT"
+elif ! grep -q 'npm ci' <<<"$WEB_DIST_MISSING"; then
+  fail "web-dist builds the SPA only when it is absent: with no index.html it runs no npm ci:
+$WEB_DIST_MISSING"
+else
+  pass "web-dist builds the SPA only when it is absent"
+fi
+
 if [ "$FAILURES" -ne 0 ]; then
   printf '%s case(s) failed\n' "$FAILURES" >&2
   exit 1
diff --git a/stats/Makefile b/stats/Makefile
index e682393a..2c62ff43 100644
--- a/stats/Makefile
+++ b/stats/Makefile
@@ -1,4 +1,4 @@
-.PHONY: test web-test web-build build up down restart ui-test-up ui-test-down check-uitest-overrides
+.PHONY: test web-test web-build web-dist build up down restart ui-test-up ui-test-down check-uitest-overrides
 
 # Runs the Go test suite. Tests that need the compose stack's PostgreSQL
 # skip themselves with a clear message when it is not running.
@@ -37,6 +37,16 @@ web-test:
 web-build:
 	cd web && npm ci && npm run build
 
+# Builds the SPA only when internal/web/dist holds no index.html. The
+# .flow/project.md `## test` and `## lint` lines that compile internal/web
+# or need web/node_modules run it first, so a worktree builds the SPA the
+# first time one of them runs and a change that runs none never does
+# (KAN-987).
+web-dist: internal/web/dist/index.html
+
+internal/web/dist/index.html:
+	$(MAKE) web-build
+
 # Both binaries: the Vite build, then `go build`, in that order -- never
 # the other way around -- so the embedded assets `go build` bakes in are
 # always the ones just produced, never a stale dist/ left over from an
````

  - [ ] **Step 3: Verify.** `bash scripts/test-make-build.sh` prints `make-build: all cases pass`; `scripts/check-references.sh`, `scripts/check-markdown-integrity.py` and `scripts/check-installed-citations.sh` exit 0; `cd stats && make -s web-dist && go vet ./...` exits 0. Commit.

**Build:** green
**Files:** `stats/Makefile`, `scripts/test-make-build.sh`, `.flow/project.md`, `.flow/project-rationale.md`, `CLAUDE.md`, `AGENTS.md`
**Tests:** `web-dist builds the SPA only when it is absent`
**Regression:** reverting `stats/Makefile` drops the `web-dist` target, so the case finds no `npm ci` with `index.html` absent and every `make -s web-dist` line fails.
**Baseline:** before=5 after=6
<!-- measured: grep -c 'pass "' scripts/test-make-build.sh @ ebc10e49 (before; after with this task's diff applied) -->
**Commit:** fix(stats): build the SPA only when a test or lint line needs it
**After:** none

**Decision:** web-dist-lazy
