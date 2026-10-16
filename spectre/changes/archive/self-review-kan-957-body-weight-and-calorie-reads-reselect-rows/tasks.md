# self-review-kan-957-body-weight-and-calorie-reads-reselect-rows

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-16-self-review-kan-957-body-weight-and-calorie-reads-reselect-rows-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`.
- **Task 1 commits in the gymie worktree `/Users/tweety53/Projects/gymie-worktrees/kan-957-body-weight-and-calorie-reads-reselect-rows`** (branch `spectre/kan-957-body-weight-and-calorie-reads-reselect-rows`), never in this agents worktree; Tasks 2 and 3 commit in this agents worktree. Never merge or push either.
- Each task's change is the unified diff its step gives, applied with `git apply` from the root of the repository its task names — every diff was applied and checked there (`design.md`, **Context**). Task 1's diff carries one line ending in a space (`execute <unnamed>: `); extract the diff byte for byte. A diff that does not apply is reported, never hand-merged by guesswork.
- Every prose line a diff adds is copied verbatim: this change's `verbatim-moves.txt` acknowledges each new or reworded sentence by its exact text, and any other wording fails `scripts/check-verbatim-moves.sh`.
- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w .` before `gofmt -l .`. The commands below vet and test only the package Task 2 touches.

## Review focus

- An empty statement with no tab-indented continuation must stay uncounted — the pool's validation query (Task 1's harness, its first case and the new one).
- A failed bundle read and a bundle with no `found:` line must exit 2, never CLEAR (Task 2's cases).
- A `skipped: change summary (absent)` line inside a recorded summary's own text must not fire signal six — only the header counts (Task 2's case).
- **Write `IN_PROGRESS`** must record the summary before the unfinished-work guard runs, or every creating run's self-check fires on its own missing summary (Task 2's prose edit).

- [x] 1. gymie's query counter counts SQL that begins on a new line
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run `bash /private/tmp/claude-501/-Users-tweety53-Projects-gymie/19d20152-1ee3-4781-ba2d-176d4703d990/scratchpad/repro-leading-newline.sh` and confirm it prints `statements=1 begin=0 select:weigh_in=1` — the defect: the `execute <unnamed>: ` statement whose SQL sits on its continuation lines is dropped.
  - [x] **Step 2: Apply the fix.** From `/Users/tweety53/Projects/gymie-worktrees/kan-957-body-weight-and-calorie-reads-reselect-rows`, apply this diff:

````diff verified:git apply --check in the gymie worktree at f3d2675d; applied to a copy of both scripts, bash scripts/test-count-request-queries.sh passes, the repro prints statements=2 begin=0 select:weigh_in=2, and with count-request-queries.sh restored the new case fails (statements=2 … select:weigh_in=1)
diff --git a/scripts/count-request-queries.sh b/scripts/count-request-queries.sh
index 8ec5aef..716ca69 100755
--- a/scripts/count-request-queries.sh
+++ b/scripts/count-request-queries.sh
@@ -13,8 +13,10 @@
 # reaches only sessions started after it. Each <url> is called once unmeasured (the pool's reconnect
 # lands there), then once between two psql marker statements; the Postgres log is sliced between the
 # markers, never by `docker logs --since` time alone (kan-952: the time slice cut the wrong lines), and
-# only lines from the database's own sessions (pg_stat_activity pids) count. An empty statement — the
-# pool's validation query — and the driver's connect-time `SET application_name` are not counted. On exit, however it exits, the setting is reset and the
+# only lines from the database's own sessions (pg_stat_activity pids) count. An empty statement with no
+# continuation line — the pool's validation query — and the driver's connect-time `SET application_name`
+# are not counted; an empty one followed by a tab-indented continuation is SQL that begins on a new line
+# (a Kotlin `"""` native query) and takes that continuation as its text. On exit, however it exits, the setting is reset and the
 # sessions are terminated again, so the pool drops its logging connections.
 #
 # <database> must be a workspace database (`gymie_<id>`, from `./gradlew devWorkspace`'s DB_URL) —
@@ -42,6 +44,15 @@ count_slice() {
       if (!(table in selects)) order[++tables] = table
       selects[table]++
     }
+    function count(text) {
+      gsub(/^[ \t]+|[ \t;]+$/, "", text)
+      if (text == "" || text ~ /^SET application_name /) return
+      statements++
+      if (toupper(text) ~ /^BEGIN([ \t]|$)/) begins++
+      held = text
+    }
+    /^\t/ && pending { pending = 0; count($0); next }
+    { pending = 0 }
     /^\t/ && held != "" { held = held " " $0; next }
     { attribute() }
     index($0, end) { exit }
@@ -49,11 +60,10 @@ count_slice() {
       pid = substr($0, RSTART + 1)
       sub(/\].*/, "", pid)
       text = substr($0, RSTART + RLENGTH)
-      gsub(/^[ \t]+|[ \t;]+$/, "", text)
-      if (index(pids, "," pid ",") && text != "" && text !~ /^SET application_name /) {
-        statements++
-        if (toupper(text) ~ /^BEGIN([ \t]|$)/) begins++
-        held = text
+      # SQL that begins on a new line logs an empty first line: the continuation below is its text.
+      if (index(pids, "," pid ",")) {
+        if (text ~ /^[ \t]*$/) pending = 1
+        else count(text)
       }
     }
     index($0, start) { inside = 1 }
diff --git a/scripts/test-count-request-queries.sh b/scripts/test-count-request-queries.sh
index 10671b5..1639938 100755
--- a/scripts/test-count-request-queries.sh
+++ b/scripts/test-count-request-queries.sh
@@ -67,6 +67,20 @@ LOG
 log1=$log
 log=$log2
 expect "selects are attributed to their FROM table in first-seen order" 8-1 200 "statements=5 begin=0 select:users=2 select:profile=1"
+log=$(cat <<'LOG'
+2026-10-16 09:00:00.000 UTC [900] LOG:  statement: SELECT 'count-request-queries:9-1:start'
+2026-10-16 09:00:00.001 UTC [200] LOG:  execute <unnamed>: select w1_0.id from weigh_in w1_0 where w1_0.user_id=$1
+2026-10-16 09:00:00.002 UTC [200] LOG:  execute <unnamed>: 
+	            SELECT * FROM weigh_in
+	             WHERE user_id = $1
+2026-10-16 09:00:00.002 UTC [200] DETAIL:  Parameters: $1 = 'a'
+2026-10-16 09:00:00.003 UTC [200] LOG:  statement: 
+2026-10-16 09:00:00.004 UTC [200] LOG:  execute S_1: select 1 from calorie_entry
+2026-10-16 09:00:00.006 UTC [901] LOG:  statement: SELECT 'count-request-queries:9-1:end'
+LOG
+)
+expect "SQL that begins on a new line counts as its continuation; an empty statement alone does not" \
+  9-1 200 "statements=3 begin=0 select:weigh_in=2 select:calorie_entry=1"
 log=$log1
 
 if printf '%s\n' "$log" | count_slice 7-9 100,101 >/dev/null 2>&1; then
````

  - [x] **Step 3: Verify.** From `/Users/tweety53/Projects/gymie-worktrees/kan-957-body-weight-and-calorie-reads-reselect-rows`: `bash scripts/test-count-request-queries.sh` exits 0 with every line `OK`, and the step 1 repro prints `statements=2 begin=0 select:weigh_in=2`. Commit.

**Build:** green
**Files:** `scripts/count-request-queries.sh`, `scripts/test-count-request-queries.sh` (gymie repository, worktree `/Users/tweety53/Projects/gymie-worktrees/kan-957-body-weight-and-calorie-reads-reselect-rows`)
**Tests:** `SQL that begins on a new line counts as its continuation; an empty statement alone does not`
**Regression:** reverting `count-request-queries.sh` makes the case print `statements=2 begin=0 select:weigh_in=1 select:calorie_entry=1` where `statements=3 begin=0 select:weigh_in=2 select:calorie_entry=1` is expected.
**Baseline:** before=3 after=4
<!-- measured: grep -c '^expect ' scripts/test-count-request-queries.sh @ f3d2675d (before; after with this task's diff applied to a copy) -->
**Commit:** fix(scripts): count SQL that begins on a new line in the request query counter
**After:** none

**Decision:** leading-newline-continuation

- [x] 2. A change with no recorded summary is unfinished work
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -rn 'record summary' skills scripts hooks` and confirm the only hit is the prescription in `skills/flow-contracts/pipeline.md`, and `grep -n 'self-review' stats/internal/guard/unfinishedwork.go` and confirm no hit — the defect: nothing refuses a change whose store holds no summary row.
  - [x] **Step 2: Write the failing tests.** Apply the `stats/internal/guard/check_unfinished_work_test.go` hunks of the diff in step 3 alone (`git apply --include='stats/internal/guard/check_unfinished_work_test.go' <the diff>`) and run `cd stats && go test ./internal/guard/ -run TestCheckUnfinishedWork -count=1`: it FAILS on the four cases `a change with no recorded summary is OUTSTANDING`, `the breakdown names the summary and its record command`, `a bundle with no found: line cannot answer` and `a failed bundle read cannot answer`.
  - [x] **Step 3: Implement the signal.** Apply the rest of this diff:

````diff verified:git apply --check at d1085fb4; gofmt -l, go vet ./internal/guard/ and go test ./internal/guard/ -count=1 pass with it applied; with unfinishedwork.go reverted the four cases step 2 names fail; bash scripts/test-integrate-change.sh prints 161 passed, 0 failed; check-verbatim-moves.sh with this change's verbatim-moves.txt, check-references.sh, check-markdown-integrity.py and check-dispatch-paragraphs.sh exit 0 with Tasks 2 and 3 applied
diff --git a/scripts/check-unfinished-work.sh b/scripts/check-unfinished-work.sh
index 5cbf97d0..7ab4a46e 100755
--- a/scripts/check-unfinished-work.sh
+++ b/scripts/check-unfinished-work.sh
@@ -80,6 +80,14 @@
 # like the dispatch read. Like signal three, only the worktree holding the
 # plan reads them.
 #
+# SIGNAL SIX: A CHANGE WITH NO RECORDED SUMMARY IS UNFINISHED (kan-957
+# self-review: 35 of 36 gymie changes landed with no summary row). The
+# summary is read through `flow self-review bundle -change <name> -C
+# <anchor>`, whose header — the lines above its first `## ` section — names
+# it `skipped: change summary (absent)` when the store holds none. A failed
+# read, or a header with no `found:` line, is exit 2, like the dispatch read.
+# Like signal three, only the worktree holding the plan reads it.
+#
 # All signals are counted independently and reported together on the one
 # line, so the operator sees the whole picture in one prompt rather than being
 # sent back around the loop one signal at a time.
diff --git a/scripts/test-integrate-change.sh b/scripts/test-integrate-change.sh
index 9ca68a3b..8e5bd310 100755
--- a/scripts/test-integrate-change.sh
+++ b/scripts/test-integrate-change.sh
@@ -166,6 +166,7 @@ case "\$1 \$2" in
   "stage begin"|"stage end") echo "\$*" >>"$WORK/stages.log" ;;
   "record findings") [ -f "$WORK/fail-findings" ] && exit 1; echo '[]' ;;
   "self-review findings") echo '[]' ;;
+  "self-review bundle") printf 'found: 1 of 1 sources; skipped: 0 of 1 sources\n' ;;
   "self-review finding") echo "\$*" >>"$WORK/fixed.log" ;;
   "record dispatches") [ -f "$WORK/loop-rows" ] && exec cat "$WORK/loop-rows"; cat "$WORK/dispatches.json" 2>/dev/null || echo '[{"key":"task-1-implementer","outcome":"completed"}]' ;;
   "record decisions") echo '[]' ;;
diff --git a/skills/flow/integrate.md b/skills/flow/integrate.md
index 9162d864..82f92fef 100644
--- a/skills/flow/integrate.md
+++ b/skills/flow/integrate.md
@@ -54,6 +54,10 @@ the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`).
   breakdown, the three-course prompt and what each course does are there. **Continue** and
   **File or join…** re-run `prepare` with `--accept-outstanding`; **Stop** leaves the mark the
   script closed `stopped`.
+  A breakdown naming `no change summary recorded` is closed by recording the summary first — the
+  `**Summary:**` bullets the last run printed, or, where this session never saw them, bullets
+  written from the change's `tasks.md` and `narrative.md` — with the command the breakdown names,
+  then re-running `prepare`.
 - **`STOP: no-verdict`** → a guard of the gate printed no verdict line, or exited non-zero without
   one: stop and ask, reporting its output. `--accept-outstanding` never passes it.
 - **`STOP: visual-verify`** → a worktree printed `VISUAL-VERIFY-MISSING`: the UI about to land was
diff --git a/stats/internal/guard/check_unfinished_work_test.go b/stats/internal/guard/check_unfinished_work_test.go
index 2296085f..f256cfe3 100644
--- a/stats/internal/guard/check_unfinished_work_test.go
+++ b/stats/internal/guard/check_unfinished_work_test.go
@@ -71,6 +71,8 @@ func newUWBare(t *testing.T, wt string) *uwFix {
 	f.setFile(t, "dispatches.json", `[{"key":"task-1-implementer","outcome":"completed"}]`+"\n")
 	// A pass row for every round a fixture's panel rows run in (0..2).
 	f.setFile(t, "passes.json", `[{"id":1,"round":0,"note":"roster: full"},{"id":2,"round":1,"note":"n"},{"id":3,"round":2,"note":"n"}]`+"\n")
+	// A recorded change summary: the bundle header signal six reads.
+	f.setFile(t, "bundle.json", "# Self-review context bundle for demo\n\nfound: 1 of 1 sources; skipped: 0 of 1 sources\n\n## change summary\n\nwhat changed\n")
 	f.setFindings(t)
 	get := pathEnv(f.bin)
 	f.env = Env{Getenv: get, Dir: wt,
@@ -353,6 +355,20 @@ func TestCheckUnfinishedWork(t *testing.T) {
 		f.setFile(t, "passes.fail", "")
 	}), uwRC("a failed pass-row read cannot answer", 2))
 
+	// 2f. Signal six — no change summary recorded (kan-957: 35 of 36 gymie
+	//     changes landed with none).
+	bundle := func(body string) func(t *testing.T, f *uwFix) {
+		return func(t *testing.T, f *uwFix) { f.setFile(t, "bundle.json", body) }
+	}
+	each(demo(bundle("# Self-review context bundle for demo\n\nfound: 0 of 1 sources; skipped: 1 of 1 sources\nskipped: change summary (absent)\n\n")),
+		uwVerdict("a change with no recorded summary is OUTSTANDING", "OUTSTANDING:"),
+		uwReason("the breakdown names the summary and its record command", "no change summary recorded — record the run's summary with `flow record summary -change demo -C "))
+	each(demo(bundle("# Self-review context bundle for demo\n\nfound: 1 of 1 sources; skipped: 0 of 1 sources\n\n## change summary\n\nskipped: change summary (absent)\n")),
+		uwVerdict("the skipped line inside a recorded summary's text is not the header's", "CLEAR:"))
+	each(demo(bundle("not a bundle\n")), uwRC("a bundle with no found: line cannot answer", 2))
+	each(demo(func(t *testing.T, f *uwFix) { f.setFile(t, "bundle.fail", "") }),
+		uwRC("a failed bundle read cannot answer", 2))
+
 	// 2d. Signal four's commit half — a dispatch row is not a commit (a
 	//     back-filled row named a task whose work never landed).
 	each(demo(committedAs("base", "fix: the appended task\n\nFlow-Key: panel-fix-2", false)),
diff --git a/stats/internal/guard/unfinishedwork.go b/stats/internal/guard/unfinishedwork.go
index f7805736..37729a03 100644
--- a/stats/internal/guard/unfinishedwork.go
+++ b/stats/internal/guard/unfinishedwork.go
@@ -357,6 +357,33 @@ func checkUnfinishedWork(args []string, env Env, stdout, stderr io.Writer) int {
 				reasons = append(reasons, strconv.Itoa(len(nums))+" panel round(s) dispatched with no pass row ("+strings.Join(named, "; ")+") — record each round's pass rows with `flow record pass -change "+name+" -round <round>` before landing")
 			}
 		}
+		// SIGNAL SIX — no change summary recorded (kan-957 self-review: 35 of
+		// 36 gymie changes landed with no summary row although pipeline.md
+		// prescribes one per run). The summary is read through the
+		// self-review bundle, the one read of it the CLI has: its header —
+		// the lines above its first `## ` source section — carries a `found:`
+		// line and names the summary `skipped: change summary (absent)` when the
+		// store holds none. A failed read, or a payload whose header carries
+		// no `found:` line, is refused like the dispatch read above.
+		bundleOut, bundleErr, bundleRC := pcFlow(env, false, "self-review", "bundle", "-change", name, "-C", anchor)
+		if bundleRC != 0 {
+			fmt.Fprintf(stderr, "%scannot read the self-review bundle for '%s' from the store — cannot determine anything: %s\n", uwPrefix, name, strings.TrimRight(string(bundleErr), "\n"))
+			return 2
+		}
+		var headerLines []string
+		for _, l := range strings.Split(string(bundleOut), "\n") {
+			if strings.HasPrefix(l, "## ") {
+				break
+			}
+			headerLines = append(headerLines, l)
+		}
+		if !slices.ContainsFunc(headerLines, func(l string) bool { return strings.HasPrefix(l, "found: ") }) {
+			fmt.Fprintf(stderr, "%sthe self-review bundle for '%s' carries no found: line — cannot determine anything\n", uwPrefix, name)
+			return 2
+		}
+		if slices.Contains(headerLines, "skipped: change summary (absent)") {
+			reasons = append(reasons, "no change summary recorded — record the run's summary with `flow record summary -change "+name+" -C "+anchor+" -file -` before landing")
+		}
 		// SIGNAL FOUR — a ticked appended task no dispatch row names: a fix
 		// round that recorded nothing (kan-979's fix-round Task 7 reached
 		// integrate with no dispatch row at all). A row names a task by its
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index c2b6b4bd..0d832147 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -396,2 +396,4 @@ flow record dispatch end -change <name> -C <canonical-worktree> -key <key> -sess
 
+**Record the `**Summary:**` bullets next**, per **Summary and live-stack line, before every handoff** (`skills/flow-contracts/pipeline.md`), since the guard below finds a change with no recorded summary unfinished.
+
 **Every run writing `IN_PROGRESS` then runs the unfinished-work guard on itself**, a creating run
@@ -461,4 +463,2 @@ a blocked step, a refused start, a `## Question` — prints as its own rule stat
 
-**The `**Summary:**` bullets are recorded before the block prints**, per **Summary and live-stack line, before every handoff** (`skills/flow-contracts/pipeline.md`).
-
 **The `Running:` section is never omitted.** Its URLs are the ones **Resolve the run
````

  - [x] **Step 4: Verify.** From the repository root: `cd stats && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run TestCheckUnfinishedWork -count=1` passes; `bash scripts/test-integrate-change.sh` ends `0 failed`; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh` and `scripts/check-stage-mark-calls.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `scripts/check-unfinished-work.sh`, `scripts/test-integrate-change.sh`, `skills/flow/verify-and-handoff.md`, `skills/flow/integrate.md`
**Tests:** `a change with no recorded summary is OUTSTANDING`, `the breakdown names the summary and its record command`, `the skipped line inside a recorded summary's text is not the header's`, `a bundle with no found: line cannot answer`, `a failed bundle read cannot answer`
**Regression:** reverting the commit lets a change whose store holds no summary row print `CLEAR:`, and a failed bundle read go unnoticed — the first, second, fourth and fifth cases fail; the third pins that only the bundle's header is read.
**Baseline:** before=96 after=101
<!-- measured: grep -c 'uwRC(\|uwVerdict(\|uwReason(' stats/internal/guard/check_unfinished_work_test.go @ d1085fb4 (before; after with this task's diff applied) -->
**Commit:** fix(guard): a change with no recorded summary is unfinished work
**After:** none

**Decision:** summary-is-unfinished-work

- [x] 3. A live check's before figure comes from a detached worktree at the merge base
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n 'detach' skills/flow/verify-and-handoff.md` and confirm no hit — the defect: **Live check** names no source for the before figure, so kan-957's run checked base sources out over its change worktree.
  - [x] **Step 2: Apply the fix.** Apply this diff:

````diff verified:git apply --check at d1085fb4, alone and after Task 2's diff; check-verbatim-moves.sh with this change's verbatim-moves.txt, check-references.sh and check-markdown-integrity.py exit 0 with Tasks 2 and 3 applied
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index 0d832147..67983ca6 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -130,7 +130,11 @@ calls: when the stack's worktree-resolved URLs do not answer, or `check-dev-stac
 (`skills/flow/implement.md`) does; then exercise what the section names and write each
 before/after figure, beside the failure look the section states, to `live-verification.md` in
 the change's directory (**A change's directory**, `skills/flow-contracts/pipeline.md`), a
-planning path the write-in-progress planning commit carries. The `## Report` gains one line, `live check — <matched | failure look>`. A figure that
+planning path the write-in-progress planning commit carries. **A before figure is measured from a
+detached worktree at the merge base** — `git -C <worktree> worktree add --detach
+<project>-worktrees/<name>-before <merge-base>`, the base's stack run from there and the worktree
+removed with `git worktree remove` once the figure is written — never by checking base sources out
+over the change worktree. The `## Report` gains one line, `live check — <matched | failure look>`. A figure that
 shows the failure look is the branch's own defect and takes **The loop**
 (`skills/flow/verify-fix-loop.md`); a stack that will not start ends the turn as a failed
 command's environment cause does. The verifier row's `end` (**Recording.**, above) is recorded
````

  - [x] **Step 3: Verify.** From the repository root `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.py` — all exit 0. Commit.

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none** — one prose sentence; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): a live check's before figure comes from a detached worktree at the merge base
**After:** Task 2

**Decision:** live-check-before-from-detached-worktree
