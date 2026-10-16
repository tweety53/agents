# pipeline-fix-kan-954-user-row-loaded-twice-per-request-14

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** a loop's `review-1` is admitted on the effort its plan's `first-review-effort` decision
records, written bare (`low`, as the skill names it) or as `flow-low`/`flow-medium`, and a recorded
decision that names no effort is refused instead of silently run on `flow-medium`.

**Architecture:** one function, `lpFirstReviewType` in `stats/internal/guard/loop.go`, and its one
caller in `checkLoopDispatch`; the refusal list in `scripts/check-loop-dispatch.sh`'s header.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; it was applied to this worktree at be156cf7 and checked there. A diff that does not
  apply is reported, never hand-merged by guesswork.
- No new inline suppressions; `cd stats && gofmt -l .` prints nothing.
- Go tests in `stats/internal/guard` take over a minute: run them with a 600000 ms timeout, in the
  <!-- measured: cd stats && go test ./internal/guard/ -count=1 @ be156cf7 with the diff applied — ok in 101.7s -->
  foreground.

## Review Focus

- The bare check runs before the backticked one, and both read only the `first-review-effort`
  entry's own `**Chosen:**` line — never `**Considered:**`, never the next entry.
- A plan with no `first-review-effort` decision, or no readable `design.md`, still admits
  `review-1` on `flow-medium` only — every existing `TestCheckLoopDispatch` case.
- The `EFFORT` refusal fires only for a recorded entry naming no effort.

- [ ] 1. Read the first-review effort as the skill names it, and refuse a recorded one naming none
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Save the diff below to
    `/tmp/pf14-task1.diff`, apply its test hunk alone
    (`git apply --include='stats/internal/guard/*_test.go' /tmp/pf14-task1.diff`) and run
    `cd stats && go test ./internal/guard/ -run 'TestCheckLoopDispatchFirstReviewEffort' -count=1`:
    six subtests FAIL (`bare_low`, `bare_low_refuses_flow-medium`, `backticked_low_in_prose`,
    `bare_value_before_a_backticked_one`, `no_effort_named`,
    `no_Chosen_line,_the_next_entry's_never_read`); `backticked_medium` passes.
  - [ ] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/internal/guard/*_test.go' /tmp/pf14-task1.diff`):

````diff verified:applied with git apply to this worktree at be156cf7; six of the new test's seven subtests fail with loop.go untouched and all pass with it; go test ./internal/guard/ -count=1 green; gofmt -l and go vet clean
diff --git a/scripts/check-loop-dispatch.sh b/scripts/check-loop-dispatch.sh
index 94b0d9b1..dc133df3 100755
--- a/scripts/check-loop-dispatch.sh
+++ b/scripts/check-loop-dispatch.sh
@@ -12,6 +12,8 @@
 #   NOT-STARTED  a pipeline-fix key with no pipeline-fix-<name>-<k> branch
 #                (scripts/pipeline-fix.sh start makes it);
 #   PAIR         subagent_type or model off the suffix's row;
+#   EFFORT       a review-1 whose loop plan records a first-review-effort
+#                decision naming no effort on its **Chosen:** line;
 #   ESCALATION   the row's escalation type with no recorded escalation
 #                (scripts/pipeline-fix.sh escalate);
 #   KEY-REUSED   the key already has a row;
diff --git a/stats/internal/guard/check_loop_dispatch_test.go b/stats/internal/guard/check_loop_dispatch_test.go
index 1d43695c..79cf9452 100644
--- a/stats/internal/guard/check_loop_dispatch_test.go
+++ b/stats/internal/guard/check_loop_dispatch_test.go
@@ -228,3 +228,38 @@ func TestCheckLoopDispatchOwnFastBranchReadsLoopPlan(t *testing.T) {
 		t.Fatalf("rc %d: %s", rc, out)
 	}
 }
+
+// review-1's first-review-effort reads the effort as a planner writes it --
+// bare or backticked low/medium as the skill names it, or flow-low/flow-medium
+// -- from that decision's own entry, and refuses an entry naming none rather
+// than falling back to flow-medium.
+func TestCheckLoopDispatchFirstReviewEffort(t *testing.T) {
+	t.Parallel()
+	const head = "## Decisions\n\n### First review\n\n**ID:** first-review-effort\n**Status:** active\n"
+	for _, tc := range []struct {
+		name, chosen, typ string
+		rc                int
+		out               string
+	}{
+		{name: "bare low", chosen: "**Chosen:** low — dev-only tooling, no hard seam.\n**Considered:** medium — no seam here.\n", typ: "flow-low", out: "LOOP-DISPATCH-OK"},
+		{name: "bare low refuses flow-medium", chosen: "**Chosen:** low — dev-only tooling.\n", typ: "flow-medium", rc: 1, out: "runs on flow-low"},
+		{name: "backticked low in prose", chosen: "**Chosen:** the first review runs at effort `low` on `opus`: a test-only seam.\n", typ: "flow-low", out: "LOOP-DISPATCH-OK"},
+		{name: "backticked medium", chosen: "**Chosen:** `medium` — a gate every run passes.\n**Considered:** `low` — too thin.\n", typ: "flow-medium", out: "LOOP-DISPATCH-OK"},
+		{name: "bare value before a backticked one", chosen: "**Chosen:** low — rather than `medium`.\n", typ: "flow-low", out: "LOOP-DISPATCH-OK"},
+		{name: "no effort named", chosen: "**Chosen:** the cheap one.\n", typ: "flow-medium", rc: 1, out: "EFFORT —"},
+		{name: "no Chosen line, the next entry's never read", chosen: "\n### Grouping\n\n**ID:** fixer-grouping\n**Chosen:** one fixer at `low`.\n", typ: "flow-low", rc: 1, out: "EFFORT —"},
+	} {
+		t.Run(tc.name, func(t *testing.T) {
+			t.Parallel()
+			fx := lpNew(t, lpPlanDone, lpBranch)
+			gitRun(t, fx.repo, "checkout", "-q", lpBranch)
+			writeFile(t, fx.repo+"/spectre/changes/"+lpBranch+"/design.md", head+tc.chosen)
+			gitRun(t, fx.repo, "add", "-A")
+			gitRun(t, fx.repo, "commit", "-q", "-m", "plan")
+			gitRun(t, fx.repo, "checkout", "-q", "main")
+			if rc, out := fx.run(t, "pipeline-fix-6-review-1", tc.typ, "opus"); rc != tc.rc || !strings.Contains(out, tc.out) {
+				t.Fatalf("rc %d, want %d; output %q, want it to carry %q", rc, tc.rc, out, tc.out)
+			}
+		})
+	}
+}
diff --git a/stats/internal/guard/loop.go b/stats/internal/guard/loop.go
index b53d6eed..95211ead 100644
--- a/stats/internal/guard/loop.go
+++ b/stats/internal/guard/loop.go
@@ -262,8 +262,12 @@ func lpEarlierRound(lp lpLoop, suffix string, rows []lpDispatch) bool {
 
 var (
 	lpFirstReviewID = regexp.MustCompile(`\*\*ID:\*\*\s*first-review-effort\b`)
+	lpEntryID       = regexp.MustCompile(`\*\*ID:\*\*`)
 	lpChosen        = regexp.MustCompile(`\*\*Chosen:\*\*[^\n]*`)
-	lpReviewType    = regexp.MustCompile("`(flow-(?:low|medium))`")
+	// The effort a **Chosen:** line names: a bare low or medium opening the
+	// value, else its first backticked low, medium, flow-low or flow-medium.
+	lpChosenBare = regexp.MustCompile(`^\*\*Chosen:\*\*\s*(low|medium)\b`)
+	lpChosenTick = regexp.MustCompile("`(?:flow-)?(low|medium)`")
 )
 
 // lpFirstReviewType is the subagent_type the planner recorded for review-1 in
@@ -271,8 +275,10 @@ var (
 // line). The loop's plan is always spectre/changes/<loop branch name>/ --
 // self-review-<name> on the change's own branch too -- read from the loop's
 // branch, or for a self-review on the change's own branch from spectre/<name>
-// (/flow) or <name> (/flow-fast); flow-medium when it recorded none.
-func lpFirstReviewType(env Env, git, repo string, lp lpLoop) string {
+// (/flow) or <name> (/flow-fast); flow-medium when it recorded none. ok is
+// false when the decision is recorded but its **Chosen:** line names no
+// effort: a refusal, never a silent flow-medium.
+func lpFirstReviewType(env Env, git, repo string, lp lpLoop) (typ string, ok bool) {
 	ref := lp.token
 	for _, b := range []string{"spectre/" + lp.change, lp.change} {
 		if lpBranchExists(env, git, repo, ref) {
@@ -281,17 +287,21 @@ func lpFirstReviewType(env Env, git, repo string, lp lpLoop) string {
 		ref = b
 	}
 	rc, body := kwRun(env, "", git, "-C", repo, "show", "refs/heads/"+ref+":spectre/changes/"+lp.token+"/design.md")
-	if rc != 0 {
-		return "flow-medium"
+	loc := lpFirstReviewID.FindStringIndex(body)
+	if rc != 0 || loc == nil {
+		return "flow-medium", true
 	}
-	if loc := lpFirstReviewID.FindStringIndex(body); loc != nil {
-		if c := lpChosen.FindString(body[loc[1]:]); c != "" {
-			if m := lpReviewType.FindStringSubmatch(c); m != nil {
-				return m[1]
-			}
+	entry := body[loc[1]:]
+	if next := lpEntryID.FindStringIndex(entry); next != nil {
+		entry = entry[:next[0]]
+	}
+	c := lpChosen.FindString(entry)
+	for _, re := range []*regexp.Regexp{lpChosenBare, lpChosenTick} {
+		if m := re.FindStringSubmatch(c); m != nil {
+			return "flow-" + m[1], true
 		}
 	}
-	return "flow-medium"
+	return "", false
 }
 
 func checkLoopDispatch(args []string, env Env, stdout, stderr io.Writer) int {
@@ -344,7 +354,11 @@ func checkLoopDispatch(args []string, env Env, stdout, stderr io.Writer) int {
 			return refuse(fmt.Sprintf("ESCALATION — %s on %s needs a completed earlier round and its no-progress escalation recorded first: scripts/pipeline-fix.sh escalate %s", key, typ, key))
 		}
 	} else if suffix == "review-1" {
-		if want := lpFirstReviewType(env, git, lp.main, lp); typ != want {
+		want, ok := lpFirstReviewType(env, git, lp.main, lp)
+		if !ok {
+			return refuse(fmt.Sprintf("EFFORT — %s: the loop's design.md first-review-effort decision names no effort on its **Chosen:** line; write low, medium, `flow-low` or `flow-medium` there", key))
+		}
+		if typ != want {
 			return refuse(fmt.Sprintf("PAIR — %s runs on %s, as the loop's design.md first-review-effort decision records (flow-medium when none); this dispatch carries %q", key, want, typ))
 		}
 	}
````

  - [ ] **Step 3: Verify.** `cd stats && gofmt -l ./internal/guard && go vet ./internal/guard/ && go test ./internal/guard/ -run 'CheckLoopDispatch|LoopTable|PipelineFix' -count=1`
    — gofmt prints nothing, the rest exit 0.
  - [ ] **Step 4: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/loop.go`, `stats/internal/guard/check_loop_dispatch_test.go`, `scripts/check-loop-dispatch.sh`
**Tests:** `TestCheckLoopDispatchFirstReviewEffort`
**Regression:** reverting the commit matches only a backticked `flow-low`/`flow-medium` again — `TestCheckLoopDispatchFirstReviewEffort` fails: the bare and backticked-bare `low` forms are refused as `flow-medium`, and a recorded entry naming no effort, or missing its `**Chosen:**` line, is admitted on `flow-medium` instead of refused with `EFFORT —`.
**Baseline:** before=5 after=6
<!-- measured: grep -cE '^func Test' stats/internal/guard/check_loop_dispatch_test.go (before at be156cf7; after with this task's diff applied) -->
**Commit:** fix(guard): read a loop's first-review effort as the skill names it
**After:** none

**Decision:** bare-effort-accepted
**Decision:** unparseable-refuses
**Decision:** entry-bounded
