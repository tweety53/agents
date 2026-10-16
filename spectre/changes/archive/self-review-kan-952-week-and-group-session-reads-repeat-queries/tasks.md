# self-review-kan-952-week-and-group-session-reads-repeat-queries

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-16-self-review-kan-952-week-and-group-session-reads-repeat-queries-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`.
- **Task 3 commits in the gymie backend worktree `/Users/tweety53/Projects/gymie-worktrees/kan-952-week-and-group-session-reads-repeat-queries`** (branch `spectre/kan-952-week-and-group-session-reads-repeat-queries`), never in this agents worktree; Tasks 1 and 2 commit in this agents worktree. Never merge or push either.
- Each task's change is the unified diff its step gives, applied with `git apply` from the root of the repository its task names — every diff was applied and checked there (`design.md`, **Context**). A diff that does not apply is reported, never hand-merged by guesswork.
- Every prose line a diff adds is copied verbatim: this change's `verbatim-moves.txt` acknowledges each new or reworded sentence by its exact text, and any other wording fails `scripts/check-verbatim-moves.sh`.
- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w .` before `gofmt -l .`. The commands below vet and test only the package Task 2 touches.
- Task 3's live run alters a database setting on the shared gymie-postgres: only ever a `gymie_<id>` workspace database, never `gymie`, and the script resets it on exit.

## Review focus

- A panel-fix row that recorded no commit — an aborted round, every existing `TestCheckPanelFixSingleDispatch` case — must stay unjudged and exit 0 (Task 2's step 4 runs those cases).
- A pushed-route fix commit carrying `Flow-Key: panel-fix-<round>` must pass, and one carrying another round's key must not (Task 2's test).
- A recorded value opening with `-` must never reach git as an option (Task 2's hex check).
- `count-request-queries.sh` pointed at the main checkout's `gymie` database must refuse before any `ALTER DATABASE` (Task 3's step 3).
- The pool's empty validation statement and connect-time `SET application_name` must not count as the request's statements (Task 3's harness).

- [x] 1. Every run writing IN_PROGRESS runs the unfinished-work guard on itself
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n 'A fix run then runs the unfinished-work guard' skills/flow/verify-and-handoff.md` and confirm one hit — the defect: only a fix run checks itself, so kan-952's creating run handed off over a panel round with no pass row.
  - [x] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply --check at 6d14366a; check-verbatim-moves.sh with this change's verbatim-moves.txt, check-references.sh, check-markdown-integrity.py and check-dispatch-paragraphs.sh exit 0 with Tasks 1 and 2 applied
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index 492fc6fd..b435e839 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -396,6 +396,7 @@ flow record dispatch end -change <name> -C <canonical-worktree> -key <key> -sess
 
-**A fix run then runs the unfinished-work guard on itself**, once per worktree of the resolved
+**Every run writing `IN_PROGRESS` then runs the unfinished-work guard on itself**, a creating run
+as well as a fix run, once per worktree of the resolved
 set, with the arguments **Run 1 — the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`) passes it —
 `check-unfinished-work.sh <worktree> <name> <canonical-worktree> <worktree>=<merge-base>...` — so an appended task left
-unrecorded or unreviewed stops this run rather than the next integrate. An `OUTSTANDING:` line is
+unrecorded or unreviewed, or a panel round dispatched with no pass row, stops this run rather than the next integrate. An `OUTSTANDING:` line is
 this run's own unfinished work: finish what its breakdown names — a missing dispatch or review is
````

  - [x] **Step 3: Verify.** From the repository root `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh` — all exit 0. Commit.

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none** — one prose paragraph; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): every run writing IN_PROGRESS runs the unfinished-work guard on itself
**After:** none

**Decision:** unfinished-work-every-in-progress

- [x] 2. A panel-fix row's recorded commit is a task commit or the round's fix commit
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n 'commitSha' stats/internal/guard/panelfixsingledispatch.go` and confirm no hit — the guard never reads a row's commit, so kan-952's `panel-fix-0` row recording the plan commit `2f0fe509` passed.
  - [x] **Step 2: Write the failing test.** Apply the `stats/internal/guard/check_panel_fix_single_dispatch_test.go` hunk of the diff in step 3 alone (`git apply --include='stats/internal/guard/check_panel_fix_single_dispatch_test.go' <the diff>`) and run `cd stats && go test ./internal/guard/ -run TestCheckPanelFixSingleDispatchCommitIsATaskCommit -count=1`: it FAILS with `expected exit 1, got 0`.
  - [x] **Step 3: Implement the check.** Apply the rest of this diff:

````diff verified:git apply --check at 6d14366a; gofmt -l, go vet ./internal/guard/ and go test ./internal/guard/ -count=1 pass with it applied; TestCheckPanelFixSingleDispatchCommitIsATaskCommit fails with panelfixsingledispatch.go reverted (expected exit 1, got 0)
diff --git a/scripts/check-panel-fix-single-dispatch.sh b/scripts/check-panel-fix-single-dispatch.sh
index 87643abb..08e16038 100755
--- a/scripts/check-panel-fix-single-dispatch.sh
+++ b/scripts/check-panel-fix-single-dispatch.sh
@@ -57,6 +57,17 @@
 #     is "cannot answer" (exit 2), never a clean verdict, matching the
 #     dispatches read below.
 #
+# and each panel-fix row whose end recorded a -commit must name a commit
+# whose subject is some task's **Commit:** subject in the change's tasks.md
+# (the fold route) or which carries a `Flow-Key: <the row's key>` trailer,
+# -retry stripped (the pushed route's fix commit) -- review-panel.md's
+# -commit contract (kan-952 self-review: panel-fix-0 recorded the plan
+# commit while its fix landed as its own commit). The commit is looked up
+# in the worktree, then in each peer the spec root's peers file names; a
+# commit none of them holds, or one that is not hex, is a violation, and a
+# tasks.md that cannot be read while such a row exists is exit 2. A row
+# that recorded no commit is not judged.
+#
 # Rows of other roles and rows of other session tokens never count, nor do
 # `full-suite-fix-<n>` rows recorded with agent id `inline`: implement.md
 # records the parent's own full-suite fix rounds under role `panel-fix`, and
@@ -80,7 +91,8 @@
 #   1  at least one violation -- each offending round or key named on stderr
 #   2  cannot answer at all -- missing arguments, a non-directory worktree,
 #      a change name outside the allowlist, the store unreachable (the flow
-#      call exits non-zero), or rows that are not readable JSON
+#      call exits non-zero), rows that are not readable JSON, or a tasks.md
+#      that cannot be read while a row recorded a commit
 # flow-guard is built from this checkout, never taken from PATH:
 # scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
 # cannot-answer code) with the cause when it cannot.
diff --git a/stats/internal/guard/check_panel_fix_single_dispatch_test.go b/stats/internal/guard/check_panel_fix_single_dispatch_test.go
index 03694851..54b36e90 100644
--- a/stats/internal/guard/check_panel_fix_single_dispatch_test.go
+++ b/stats/internal/guard/check_panel_fix_single_dispatch_test.go
@@ -322,3 +322,40 @@ func TestCheckPanelFixSingleDispatch(t *testing.T) {
 		}
 	})
 }
+
+// TestCheckPanelFixSingleDispatchCommitIsATaskCommit pins review-panel.md's
+// -commit contract: the task commit the fixup folded into, or the fix commit
+// carrying the round's Flow-Key: trailer. A row recording any other commit
+// -- kan-952's panel-fix-0 recorded its plan commit -- exits 1 naming it.
+func TestCheckPanelFixSingleDispatchCommitIsATaskCommit(t *testing.T) {
+	t.Parallel()
+	wt := t.TempDir()
+	gitRun(t, wt, "init", "-q")
+	writeFile(t, filepath.Join(wt, "spectre/changes/demo/tasks.md"),
+		"# demo\n\n- [ ] 1. Add alpha\n\n**Commit:** feat(alpha): add alpha\n")
+	head := func(msg ...string) string {
+		args := []string{"commit", "-q", "--allow-empty"}
+		for _, m := range msg {
+			args = append(args, "-m", m)
+		}
+		gitRun(t, wt, args...)
+		out, err := exec.Command(fixtureGit, "-C", wt, "rev-parse", "HEAD").Output()
+		if err != nil {
+			t.Fatal(err)
+		}
+		return strings.TrimSpace(string(out))
+	}
+	plan := head("chore(spectre): plan demo")
+	task := head("feat(alpha): add alpha")
+	fix := head("fix(alpha): review findings", "Flow-Key: panel-fix-0")
+	other := head("fix(alpha): another round", "Flow-Key: panel-fix-1")
+	row := func(sha string) []byte {
+		return []byte(`[{"key":"panel-fix-0","role":"panel-fix","sessionToken":"mf-tok","commitSha":"` + sha + `"}]`)
+	}
+	args := []string{wt, "demo", "mf-tok"}
+	pfdExpect(t, pfdCase{dispatches: row(task), args: args}, 0, "OK")
+	pfdExpect(t, pfdCase{dispatches: row(fix), args: args}, 0, "OK")
+	pfdExpect(t, pfdCase{dispatches: row(other), args: args}, 1, "carries no Flow-Key: panel-fix-0 trailer")
+	pfdExpect(t, pfdCase{dispatches: row(plan), args: args}, 1, "panel-fix-0 records commit "+plan)
+	pfdExpect(t, pfdCase{dispatches: row("0123456789abcdef0123456789abcdef01234567"), args: args}, 1, "no repository of the change holds")
+}
diff --git a/stats/internal/guard/panelfixsingledispatch.go b/stats/internal/guard/panelfixsingledispatch.go
index 9d23aac9..b2101b71 100644
--- a/stats/internal/guard/panelfixsingledispatch.go
+++ b/stats/internal/guard/panelfixsingledispatch.go
@@ -5,8 +5,10 @@ import (
 	"encoding/json"
 	"fmt"
 	"io"
+	"os"
 	"path/filepath"
 	"regexp"
+	"slices"
 	"sort"
 	"strconv"
 	"strings"
@@ -36,6 +38,10 @@ const pfdPrefix = "check-panel-fix-single-dispatch: "
 // pins the verdicts measured from bash.
 var pfdKeyRE = regexp.MustCompile(`^panel-fix-[0-9]+(-[0-9]+)?(-retry)?$`)
 
+// pfdShaRE is a recorded commit git may be handed: hex only, so a value
+// opening with `-` never reaches git as an option.
+var pfdShaRE = regexp.MustCompile(`^[0-9a-f]{4,64}$`)
+
 var pfdFullSuiteRE = regexp.MustCompile(`^full-suite-fix-[0-9]+$`)
 
 // pfdBase is one panel-fix base key (the key with a trailing -retry
@@ -120,7 +126,7 @@ func checkPanelFixSingleDispatch(args []string, env Env, stdout, stderr io.Write
 
 	// Each panel-fix row of THIS token. A payload that is not the shape the
 	// verb documents is the same cannot-answer posture as a failed call.
-	keys, ok := pfdPanelFixKeys(rows, token)
+	fixRows, ok := pfdPanelFixRows(rows, token)
 	if !ok {
 		fmt.Fprintf(stderr, "%sdispatch rows were not readable JSON -- cannot answer\n", pfdPrefix)
 		return 2
@@ -134,7 +140,8 @@ func checkPanelFixSingleDispatch(args []string, env Env, stdout, stderr io.Write
 	var violations []string
 	var bases []*pfdBase
 	index := map[string]*pfdBase{}
-	for _, key := range keys {
+	for _, row := range fixRows {
+		key := row.key
 		if !pfdKeyRE.MatchString(key) {
 			violations = append(violations, fmt.Sprintf("out-of-shape panel-fix key '%s' -- the canonical shape is panel-fix-<round>[-<chunk>][-retry]", key))
 			continue
@@ -235,6 +242,17 @@ func checkPanelFixSingleDispatch(args []string, env Env, stdout, stderr io.Write
 		}
 	}
 
+	// A panel-fix row's -commit is the task commit its fixup folded into, or
+	// on the pushed route the fix commit itself, carrying the round's
+	// Flow-Key: trailer (review-panel.md). Any other commit is the wrong
+	// record -- kan-952's panel-fix-0 recorded its plan commit while its fix
+	// landed as 50af0ddf.
+	commitViolations, ok := pfdCheckCommits(env, worktree, name, fixRows, stderr)
+	if !ok {
+		return 2
+	}
+	violations = append(violations, commitViolations...)
+
 	if len(violations) > 0 {
 		for _, v := range violations {
 			fmt.Fprintf(stderr, "%s%s\n", pfdPrefix, v)
@@ -276,18 +294,22 @@ func pfdStream(raw []byte) ([]any, bool) {
 	}
 }
 
-// pfdPanelFixKeys is `.[] | select((.sessionToken // "") == $tok and
+// pfdRow is one panel-fix row of the run: its key, and the commitSha its
+// end recorded ("" when none).
+type pfdRow struct{ key, commit string }
+
+// pfdPanelFixRows is `.[] | select((.sessionToken // "") == $tok and
 // (.role // "") == "panel-fix")` then each row's `.key // ""` as jq -r
-// printed it. A null row selects nothing; any other non-object row, or a
+// printed it, beside the row's commitSha. A null row selects nothing; any other non-object row, or a
 // top-level value that is not an array, is jq failing. (jq iterated a
 // top-level object's values in insertion order; that shape is refused here
 // rather than guessed at -- the verb prints an array.)
-func pfdPanelFixKeys(raw []byte, token string) ([]string, bool) {
+func pfdPanelFixRows(raw []byte, token string) ([]pfdRow, bool) {
 	vals, ok := pfdStream(raw)
 	if !ok {
 		return nil, false
 	}
-	var keys []string
+	var keys []pfdRow
 	for _, v := range vals {
 		arr, ok := v.([]any)
 		if !ok {
@@ -312,12 +334,91 @@ func pfdPanelFixKeys(raw []byte, token string) ([]string, bool) {
 			if row["agentId"] == "inline" && pfdFullSuiteRE.MatchString(key) {
 				continue
 			}
-			keys = append(keys, key)
+			commit, _ := row["commitSha"].(string)
+			keys = append(keys, pfdRow{key: key, commit: commit})
 		}
 	}
 	return keys, true
 }
 
+// pfdCheckCommits names each row whose recorded commit carries neither a
+// task's Commit: subject nor a Flow-Key: trailer naming the row's key (its
+// -retry stripped), looked up in the worktree and then in each peer its
+// spec root's peers file names. Rows that recorded no commit are not judged;
+// false is cannot answer -- a tasks.md that does not resolve or read.
+func pfdCheckCommits(env Env, worktree, name string, fixRows []pfdRow, stderr io.Writer) ([]string, bool) {
+	var judged []pfdRow
+	for _, r := range fixRows {
+		if r.commit != "" {
+			judged = append(judged, r)
+		}
+	}
+	if len(judged) == 0 {
+		return nil, true
+	}
+	tasksMD, _ := tcfResolveTasksMD(env, io.Discard, "check-panel-fix-single-dispatch", worktree, name, worktree)
+	raw, err := os.ReadFile(tasksMD)
+	var plan []string
+	if err == nil {
+		plan, err = tcfPlanLines(raw)
+	}
+	if err != nil {
+		fmt.Fprintf(stderr, "%scannot read tasks.md for '%s' -- cannot answer\n", pfdPrefix, name)
+		return nil, false
+	}
+	subjects := map[string]bool{}
+	for _, id := range tcfTaskIDs(plan) {
+		if t, ok := tcfParseTask(plan, id); ok && t.commit != nil {
+			subjects[*t.commit] = true
+		}
+	}
+	repos := append([]string{worktree}, pfdPeerRoots(env, worktree)...)
+	var violations []string
+	for _, r := range judged {
+		var meta []string
+		for _, repo := range repos {
+			if !pfdShaRE.MatchString(r.commit) {
+				break
+			}
+			if out, err := tcfRunGit(env, repo, "log", "-1", "--format=%s%n%(trailers:key=Flow-Key,valueonly)", r.commit, "--"); err == nil {
+				meta = strings.Split(strings.TrimSpace(out), "\n")
+				break
+			}
+		}
+		if meta == nil {
+			violations = append(violations, fmt.Sprintf("%s records commit %s, which no repository of the change holds -- -commit is the commit the fix landed in", r.key, r.commit))
+			continue
+		}
+		base, _ := strings.CutSuffix(r.key, "-retry")
+		if !subjects[meta[0]] && !slices.Contains(meta[1:], base) && !slices.Contains(meta[1:], r.key) {
+			violations = append(violations, fmt.Sprintf("%s records commit %s (%q), which is no task's Commit: subject and carries no Flow-Key: %s trailer -- -commit is the task commit the fixup was folded into, or the fix commit itself", r.key, r.commit, meta[0], base))
+		}
+	}
+	return violations, true
+}
+
+// pfdPeerRoots is every peer checkout <worktree>/<spec root>/peers names,
+// resolved as changePlanPeerRoot resolves one; a peer that does not resolve
+// is skipped.
+func pfdPeerRoots(env Env, worktree string) []string {
+	leaf := specRootLeaf(worktree, io.Discard)
+	b, err := os.ReadFile(worktree + "/" + leaf + "/peers")
+	if err != nil {
+		return nil
+	}
+	var roots []string
+	for _, line := range lines(b) {
+		f := strings.Fields(line)
+		if len(f) == 0 {
+			continue
+		}
+		if root, ok := changePlanPeerRoot(env, worktree, leaf, f[0]); ok {
+			roots = append(roots, root)
+		}
+	}
+	return roots
+}
+
 // pfdRawText is `.key // ""` as jq -r printed it and `$(...)` captured it:
 // a string raw with its NUL bytes dropped and then trailing newlines
 // stripped, null/false as "", anything else as its JSON text. One
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index 9edf1c38..f1d5d1e2 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -978,3 +978,5 @@ flow record dispatch end -change <name> -key panel-fix-<round>[-<chunk>] \
 
-`-commit` is the task commit the fixup was folded into.
+`-commit` is the task commit the fixup was folded into, or, where the branch is already on the
+remote (**Panel re-runs**, `skills/flow/review-panel-fix-round.md`), the fix commit itself, carrying
+the round's `Flow-Key:` trailer.
 
@@ -1010,2 +1012,6 @@ violation, a key outside the canonical shape — and is a prompt, resolved on it
 
+A violation naming a row's recorded commit is not that prompt: the row recorded a commit the fix
+did not land in. The parent ends the row again under its own key and token with the commit the
+fix landed in as `-commit`, and runs the guard again.
+
 Exit 2 stops the run — a close the guard cannot answer for is not a clean close.
````

  - [x] **Step 4: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckPanelFixSingleDispatch' -count=1`, then from the repository root `bash -n scripts/check-panel-fix-single-dispatch.sh`, `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh` — all exit 0 with `gofmt -l` printing nothing. Commit.

**Build:** green
**Files:** `stats/internal/guard/panelfixsingledispatch.go`, `stats/internal/guard/check_panel_fix_single_dispatch_test.go`, `scripts/check-panel-fix-single-dispatch.sh`, `skills/flow/review-panel.md`
**Tests:** `TestCheckPanelFixSingleDispatchCommitIsATaskCommit`
**Regression:** reverting the commit lets a panel-fix row record any commit — kan-952's plan commit included — and the guard prints its OK line.
**Baseline:** before=1 after=2
<!-- measured: grep -c '^func Test' stats/internal/guard/check_panel_fix_single_dispatch_test.go @ 6d14366a (before; after with this task's diff applied) -->
**Commit:** fix(guard): a panel-fix row's recorded commit is a task commit or the round's fix commit
**After:** none

**Decision:** panel-fix-commit-is-landing

- [x] 3. gymie counts SQL statements and BEGINs per HTTP request against a worktree database
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/gymie-worktrees/kan-952-week-and-group-session-reads-repeat-queries`, run `ls scripts/count-request-queries.sh` and confirm `No such file or directory`, and `grep -n 'count-request-queries' .flow/project.md` and confirm no hit — the defect: no tool counts a request's SQL, so kan-924 and kan-952 did it by hand.
  - [x] **Step 2: Apply the fix.** From `/Users/tweety53/Projects/gymie-worktrees/kan-952-week-and-group-session-reads-repeat-queries`, apply this diff:

````diff verified:git apply --check in the gymie worktree at e289b126; bash scripts/test-count-request-queries.sh passes and fails with the statement filter removed; run live against gymie_kan_952_week_fd3b on backend 8361 it printed http=200 statements=0 begin=0 for /actuator/health and left the database setting reset
diff --git a/.flow/project.md b/.flow/project.md
index 2ab4cca7..f2f787cf 100644
--- a/.flow/project.md
+++ b/.flow/project.md
@@ -54,2 +54,7 @@ migration: it realigns the checksum either way, on whichever database it is poin
 
+A live check that counts the SQL statements or transactions an HTTP request runs (kan-924, kan-952)
+uses `scripts/count-request-queries.sh <database> [<curl-option>...] -- <url>...` against this
+worktree's database — the name in `DB_URL` from `./gradlew devWorkspace`; the script's header states
+the method.
+
 Single services, when the full stack is not wanted:
diff --git a/scripts/count-request-queries.sh b/scripts/count-request-queries.sh
new file mode 100755
index 00000000..3e686ecb
--- /dev/null
+++ b/scripts/count-request-queries.sh
@@ -0,0 +1,116 @@
+#!/usr/bin/env bash
+# Counts the SQL statements and the BEGINs each HTTP request runs against a worktree database.
+#
+#   usage: count-request-queries.sh <database> [<curl-option>...] -- <url>...
+#
+# For every <url> it prints one tab-separated line: `<url>  http=<code>  statements=<n>  begin=<n>`.
+#
+# How: `log_statement=all` is set on <database> alone (`ALTER DATABASE … SET`) and that database's open
+# sessions are terminated, so the backend's pool reconnects with the setting — a per-database setting
+# reaches only sessions started after it. Each <url> is called once unmeasured (the pool's reconnect
+# lands there), then once between two psql marker statements; the Postgres log is sliced between the
+# markers, never by `docker logs --since` time alone (kan-952: the time slice cut the wrong lines), and
+# only lines from the database's own sessions (pg_stat_activity pids) count. An empty statement — the
+# pool's validation query — and the driver's connect-time `SET application_name` are not counted. On exit, however it exits, the setting is reset and the
+# sessions are terminated again, so the pool drops its logging connections.
+#
+# <database> must be a workspace database (`gymie_<id>`, from `./gradlew devWorkspace`'s DB_URL) —
+# never the main checkout's shared `gymie`. The container is $PG_CONTAINER, default gymie-postgres.
+# Auth is the caller's: pass curl options such as -H 'Authorization: Bearer …' before `--`.
+#
+# Exit codes: 0 every url counted; 1 a curl call failed; 2 usage, a database that is not a workspace
+# database, a psql call that failed, or an end marker that never reached the log.
+
+set -euo pipefail
+
+# count_slice <nonce> <pids>: reads a Postgres log on stdin and prints `statements=<n> begin=<n>` for
+# the lines between the nonce's start and end markers whose pid is in <pids> (comma-separated).
+count_slice() {
+  awk -v start="count-request-queries:$1:start" -v end="count-request-queries:$1:end" -v pids=",$2," '
+    index($0, end) { exit }
+    inside && match($0, /\[[0-9]+\] LOG:  (statement|execute [^:]*): /) {
+      pid = substr($0, RSTART + 1)
+      sub(/\].*/, "", pid)
+      text = substr($0, RSTART + RLENGTH)
+      gsub(/^[ \t]+|[ \t;]+$/, "", text)
+      if (index(pids, "," pid ",") && text != "" && text !~ /^SET application_name /) {
+        statements++
+        if (toupper(text) ~ /^BEGIN([ \t]|$)/) begins++
+      }
+    }
+    index($0, start) { inside = 1 }
+    END { printf "statements=%d begin=%d\n", statements, begins }
+  '
+}
+
+# Sourced rather than executed? Stop here: scripts/test-count-request-queries.sh wants count_slice only.
+if [ "${BASH_SOURCE[0]}" != "${0}" ]; then
+  return 0
+fi
+
+usage="usage: count-request-queries.sh <database> [<curl-option>...] -- <url>..."
+db=${1:?$usage}
+shift
+if ! [[ $db =~ ^gymie_[a-z0-9_]+$ ]]; then
+  echo "count-request-queries: '$db' is not a workspace database (gymie_<id>)" >&2
+  exit 2
+fi
+curl_opts=()
+while [ $# -gt 0 ] && [ "$1" != "--" ]; do
+  curl_opts+=("$1")
+  shift
+done
+if [ $# -lt 2 ]; then
+  echo "$usage" >&2
+  exit 2
+fi
+shift
+container=${PG_CONTAINER:-gymie-postgres}
+
+psql_on() {
+  docker exec "$container" psql -U gymie -d "$1" -v ON_ERROR_STOP=1 -qtAc "$2" || {
+    echo "count-request-queries: psql failed on $1" >&2
+    exit 2
+  }
+}
+
+terminate_sessions() {
+  psql_on postgres "SELECT count(pg_terminate_backend(pid)) FROM pg_stat_activity WHERE datname = '$db'" >/dev/null
+}
+
+reset_logging() {
+  psql_on postgres "ALTER DATABASE \"$db\" RESET log_statement"
+  terminate_sessions
+}
+
+psql_on postgres "ALTER DATABASE \"$db\" SET log_statement = 'all'"
+trap reset_logging EXIT
+terminate_sessions
+since=$(($(date +%s) - 1))
+
+i=0
+for url in "$@"; do
+  i=$((i + 1))
+  nonce="$$-$i"
+  curl -sS -o /dev/null "${curl_opts[@]}" "$url" || exit 1
+  psql_on "$db" "SELECT 'count-request-queries:$nonce:start'" >/dev/null
+  code=$(curl -sS -o /dev/null -w '%{http_code}' "${curl_opts[@]}" "$url") || exit 1
+  psql_on "$db" "SELECT 'count-request-queries:$nonce:end'" >/dev/null
+  pids=$(psql_on postgres "SELECT string_agg(pid::text, ',') FROM pg_stat_activity WHERE datname = '$db'")
+  # The log is written as the statements run; poll until the end marker is in it rather than
+  # assuming it already is.
+  log=""
+  for _ in $(seq 1 25); do
+    log=$(docker logs --since "$since" "$container" 2>&1)
+    case $log in *"count-request-queries:$nonce:end"*) break ;; esac
+    sleep 0.2
+  done
+  case $log in
+    *"count-request-queries:$nonce:end"*) ;;
+    *)
+      echo "count-request-queries: the end marker for $url never reached the log" >&2
+      exit 2
+      ;;
+  esac
+  printf '%s\thttp=%s\t%s\n' "$url" "$code" "$(printf '%s\n' "$log" | count_slice "$nonce" "$pids" | tr ' ' '\t')"
+done
diff --git a/scripts/test-count-request-queries.sh b/scripts/test-count-request-queries.sh
new file mode 100755
index 00000000..714c9ad3
--- /dev/null
+++ b/scripts/test-count-request-queries.sh
@@ -0,0 +1,55 @@
+#!/usr/bin/env bash
+# Regression checks for scripts/count-request-queries.sh's log slicing and counting.
+#
+# Sources the script — which returns early when sourced, defining count_slice and running nothing — and
+# feeds it a recorded Postgres log. The live half (ALTER DATABASE, curl, docker logs) needs a running
+# worktree stack; the slice is where a miscount hides silently, so it is the half tested here.
+
+set -euo pipefail
+
+repo_root=$(cd "$(dirname "$0")/.." && pwd)
+# shellcheck source=/dev/null
+source "$repo_root/scripts/count-request-queries.sh"
+
+fail=0
+
+expect() {
+  local label=$1 nonce=$2 pids=$3 expected=$4 actual
+  actual=$(printf '%s\n' "$log" | count_slice "$nonce" "$pids")
+  if [ "$actual" != "$expected" ]; then
+    echo "FAIL $label" >&2
+    echo "  expected: [$expected]" >&2
+    echo "  actual:   [$actual]" >&2
+    fail=1
+    return
+  fi
+  echo "OK $label"
+}
+
+log=$(cat <<'LOG'
+2026-10-16 07:50:43.394 UTC [100] LOG:  statement: SET application_name = 'PostgreSQL JDBC Driver'
+2026-10-16 07:50:43.400 UTC [100] LOG:  execute S_1: select u.id from users u where u.id=$1
+2026-10-16 07:50:43.497 UTC [900] LOG:  statement: SELECT 'count-request-queries:7-1:start'
+2026-10-16 07:50:43.500 UTC [100] LOG:  statement: 
+2026-10-16 07:50:43.501 UTC [100] LOG:  execute S_1: BEGIN
+2026-10-16 07:50:43.502 UTC [100] LOG:  execute <unnamed>: select w.id,
+	w.name from workout w where w.owner=$1
+2026-10-16 07:50:43.502 UTC [100] DETAIL:  Parameters: $1 = 'a1'
+2026-10-16 07:50:43.503 UTC [101] LOG:  statement: SET application_name = 'PostgreSQL JDBC Driver'
+2026-10-16 07:50:43.504 UTC [101] LOG:  execute S_2: begin read only
+2026-10-16 07:50:43.505 UTC [101] LOG:  execute S_3: select 1 from access_grant
+2026-10-16 07:50:43.506 UTC [100] LOG:  execute S_4: COMMIT
+2026-10-16 07:50:43.507 UTC [555] LOG:  statement: select 1 from another_workspace
+2026-10-16 07:50:43.508 UTC [100] LOG:  statement: BEGINNING_IS_NOT_BEGIN()
+2026-10-16 07:50:43.591 UTC [901] LOG:  statement: SELECT 'count-request-queries:7-1:end'
+2026-10-16 07:50:43.600 UTC [100] LOG:  execute S_1: BEGIN
+2026-10-16 07:50:43.601 UTC [100] LOG:  execute S_5: select 2
+LOG
+)
+
+expect "counts statements and BEGINs between the markers only, from the database's own sessions" \
+  7-1 100,101 "statements=6 begin=2"
+expect "a pid outside the list is another database's traffic" 7-1 101 "statements=2 begin=1"
+expect "an unknown nonce slices nothing" 7-9 100,101 "statements=0 begin=0"
+
+exit "$fail"
````

  - [x] **Step 3: Verify.** From `/Users/tweety53/Projects/gymie-worktrees/kan-952-week-and-group-session-reads-repeat-queries`: `bash -n scripts/count-request-queries.sh`, `bash scripts/test-count-request-queries.sh` (three `OK` lines, exit 0), and `scripts/count-request-queries.sh gymie -- http://localhost:8080/` exits 2 with `'gymie' is not a workspace database` before touching the database. Commit in that worktree: `git -C /Users/tweety53/Projects/gymie-worktrees/kan-952-week-and-group-session-reads-repeat-queries add scripts/count-request-queries.sh scripts/test-count-request-queries.sh .flow/project.md && git -C /Users/tweety53/Projects/gymie-worktrees/kan-952-week-and-group-session-reads-repeat-queries commit -m "chore(scripts): count SQL statements and BEGINs per HTTP request against a worktree database" -- scripts/count-request-queries.sh scripts/test-count-request-queries.sh .flow/project.md`.

**Build:** green
**Files:** `scripts/count-request-queries.sh`, `scripts/test-count-request-queries.sh`, `.flow/project.md` (gymie repository, worktree `/Users/tweety53/Projects/gymie-worktrees/kan-952-week-and-group-session-reads-repeat-queries`)
**Tests:** `test-count-request-queries.sh`
**Regression:** reverting the commit removes the script and its harness; with the harness kept and `count_slice`'s pid or empty-statement filter removed, its first two cases fail (`statements=8` where `statements=6` is expected).
**Baseline:** before=0 after=3
<!-- predicted: grep -c '^expect "' scripts/test-count-request-queries.sh in the gymie worktree after this task's commit; no harness exists before it -->
**Commit:** chore(scripts): count SQL statements and BEGINs per HTTP request against a worktree database
**After:** none

**Decision:** request-query-counter
