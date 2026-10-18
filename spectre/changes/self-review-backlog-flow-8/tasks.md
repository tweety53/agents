# self-review-backlog-flow-8

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the next five live Backlog Flow findings — KAN-821's sibling frame parts read as
departures, KAN-831's reproducer that cannot carry its own bound, KAN-837's whole-suite
re-baseline under unrelated failures, KAN-883's harness agent worktree failing the location guard,
and KAN-885's plan-class diff measured from a moved base — plus the grep-pipe test failure found
on `origin/main` while verifying them.

**Architecture:** two prose additions where their readers already look — **D. Basic Workflow #3 —
Writing plans** with the verifier's attribution paragraph, and **The loop**; three `flow-guard`
changes in `stats/internal/guard/` — `runReproducer`'s bound, `checkWorktreeLocation`'s FOREIGN
line, `planClass`'s fork point — each with its table test, its script header and, for the bound,
its two panel files; and one two-line harness fix.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-19-self-review-backlog-flow-8-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Every task runs in this worktree, `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-8`, branch `self-review-backlog-flow-8`; every `**Files:**` path is relative to its root. Never touch a main checkout.
- Every task's diff is the unified diff its fix step gives, applied with `git apply` from the worktree's root, checked in order in a scratch clone at `9df14302`. Save each by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork. Each sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Tasks 2, 4 and 5 prove their tests bite: each new test fails with the fix removed, as the task's Step 3 states, and passes with it restored.
- Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- Task 1: the verifier still blocks a frame with no sidecar line; only the parts the comment names read as expected.
- Task 2: the knob and the injected bound still win over the declaration; a malformed declaration leaves 20 seconds; `rrDeclaredBound` reads the same 10-line window as the mutation marker.
<!-- measured: grep -n -e 'rrDefaultBnd = "20"' -e 'headLines(path, 10)' stats/internal/guard/runreproducer.go @ 9df14302 → line 41 rrDefaultBnd = "20", line 503 headLines(path, 10) -->
- Task 4: a `FOREIGN:` line is never counted and never reaches `check-finish-preflight`'s `STRAY: ` relay.
- Task 5: a recorded ancestor sha classifies exactly as before; no shared history is exit 2.

- [ ] 1. Declare a sibling ticket's unbuilt frame parts expected (KAN-821)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n -F 'sibling ticket' skills/flow/brainstorm-planner.md skills/flow/visual-verify-verifier.md; echo $?`: it prints only `1`.
    <!-- measured: the grep above in a scratch clone @ 9df14302 → exit 1, no line -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 9df14302 after the earlier tasks' diffs; check-references.sh, check-installed-citations.sh, check-markdown-integrity.py and check-dispatch-paragraphs.sh exit 0, and check-verbatim-moves.sh 9df14302 fails only on sentences verbatim-moves.txt lists
diff --git a/skills/flow/brainstorm-planner.md b/skills/flow/brainstorm-planner.md
index b7acb7d6..f9518e11 100644
--- a/skills/flow/brainstorm-planner.md
+++ b/skills/flow/brainstorm-planner.md
@@ -299,6 +299,14 @@ paragraph under **The build-green tag** (`skills/flow-contracts/build-green.md`)
 > `<frame id>.png`** — a capture screenshot's name takes that shape too, so only the path form is
 > a citation `check-mockup-citations.sh` can resolve.
 
+> **Write a frame a sibling ticket finishes as this change's parts only.** When sibling tickets
+> share a mockup frame family and a frame on this change's list also draws parts a sibling ticket
+> builds, `design.md` records a decision naming the parts of that frame this change builds, and
+> the frame's `<spec>.mockups` sidecar line sits under a `#` comment naming the parts expected to
+> differ, the sibling ticket that builds them and that decision's ID — the verifier then reads
+> those parts as expected until the sibling lands, and judges every other part of the frame as
+> usual (KAN-821).
+
 > **Write a live check when the change touches a running service or persistent state — never a
 > live-verification task.** When the change under plan touches a long-running service, a daemon,
 > a store, a scheduler or anything else holding runtime state, `design.md` carries a
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index 5bac5327..c9e4ea0c 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -210,6 +210,11 @@ never an exit code the command did not return.
    carries no baseline PNG for: every departure on either is this change's, and the report names
    the missing baseline. Every line that reports the departure names its cause or its attribution
    beside it — never a bare `departure`.
+   **A part the frame's sidecar comment declares expected until a named sibling ticket lands is
+   no departure.** The line reporting the frame names that part
+   `expected — <sibling ticket>, <decision ID>` and nothing blocks on it, while every part the
+   comment does not name is measured and attributed as above — the comment being the one
+   `skills/flow/brainstorm-planner.md` plans for a frame a sibling ticket finishes (KAN-821).
 
    **A full-page match — a clean composite read, a structural match — is necessary, never
    sufficient: verify at the control level, measure rather than eyeball, and exercise the states
````

  - [ ] **Step 3: Verify.** `grep -c -F 'KAN-821' skills/flow/brainstorm-planner.md skills/flow/visual-verify-verifier.md` prints `1` for each file;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ 9df14302 → 1 and 1 -->
  - [ ] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`, `skills/flow/visual-verify-verifier.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): declare a sibling ticket's unbuilt frame parts expected
**After:** none

**Decision:** sibling-parts-sidecar-comment

- [ ] 2. Bound a reproducer by its own declared seconds (KAN-831)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n 'reproducer-bound' stats/internal/guard/runreproducer.go scripts/run-reproducer.sh skills/flow/review-panel.md skills/flow/review-panel-fix-round.md; echo $?`: it prints only `1`.
    <!-- measured: the grep above in a scratch clone @ 9df14302 → exit 1, no line -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 9df14302 after Task 1's diff; gofmt -l, go vet ./internal/guard and check-mutation-reproducer-pin.sh clean; go test ./internal/guard -run 'TestRRDeclaredBound|TestRunReproducerEnforcesDeclaredBound|TestRRDeadline|TestRunReproducer$' passes; TestRunReproducerEnforcesDeclaredBound fails, killed at the 20s bound, with rrDeclaredBound(resolved) put back to rrDefaultBnd
diff --git a/scripts/run-reproducer.sh b/scripts/run-reproducer.sh
index b9f3d9a5..43b527b3 100755
--- a/scripts/run-reproducer.sh
+++ b/scripts/run-reproducer.sh
@@ -64,6 +64,14 @@
 # ambiguity refusal (KAN-524) compares verdicts rather than raw exit codes,
 # it works unchanged under either.
 #
+# THE BOUND IS 20 SECONDS UNLESS THE REPRODUCER DECLARES ITS OWN (KAN-831).
+# A reproducer that builds or runs tests — a Gradle suite outlives 20
+# seconds — carries the exact line `# reproducer-bound: <seconds>`, a
+# positive whole number, within its first 10 lines, and this script then
+# bounds every run of it by that many seconds. A malformed declaration
+# leaves the default. RUN_REPRODUCER_BOUND_SECONDS, when set, still wins
+# over the declaration.
+#
 # THE COMPARISON IS VALID ONLY BETWEEN TWO RUNS OF THE SAME FILE. The
 # verdict a `--pre-fix-verdict` re-run compares against is meaningful only when
 # the file is byte-for-byte what the dispatch-time run read — a re-authored
diff --git a/skills/flow/review-panel-fix-round.md b/skills/flow/review-panel-fix-round.md
index 8c6718c7..e5f11397 100644
--- a/skills/flow/review-panel-fix-round.md
+++ b/skills/flow/review-panel-fix-round.md
@@ -252,10 +252,11 @@ run-reproducer.sh <worktree> "<the finding's finding-reproducer: text>"
 recorded relative path resolves, canonical first, else the one recorded worktree carrying it
 (KAN-658) — the same tree the exit-contract guard above already ran this reproducer in.
 
-**A reproducer that builds or runs tests outlives the runner's 20-second bound.** Every run of
-one — the exit-contract guard above, the per-finding run here and its re-run below — carries
-`RUN_REPRODUCER_BOUND_SECONDS=<the seconds that build or test run needs>` in front of the
-command, in a foreground Bash call whose timeout covers the bound (600000 max), so it reaches a verdict rather than a timeout.
+**A reproducer that builds or runs tests outlives the runner's 20-second bound, so it declares its
+own.** Its header carries the `# reproducer-bound: <seconds>` line that
+`skills/flow/review-panel.md` asks of it, and every run of one — the exit-contract guard above, the
+per-finding run here and its re-run below — enforces that bound, in a foreground Bash call whose
+timeout covers it (600000 max), so it reaches a verdict rather than a timeout (KAN-831).
 
 Read its exit code: **0** dispatches the finding; **1** bounces it once, back to the raising slot,
 carrying the reproducer's passing output; **2** is a refusal — recorded **unverifiable**, put to the
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index bad12d95..3e564864 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -489,7 +489,11 @@ it on stderr, exiting non-zero — never exit 0, the vacuous pass a rename must
 (KAN-839). The marker the dispatch audit reads (KAN-904): the script must reference each declared
 premise path — `check-panel-reproducer-exit-contract.sh` refuses a declared premise the
 script never names outside the `# premise:` declaration itself, the declared-but-unasserted
-premise that reads a vacuous green. Carry the premise rule on every slot's dispatch prompt. **The exemption form is available to Minor
+premise that reads a vacuous green. Carry the premise rule on every slot's dispatch prompt.
+**A reproducer that builds or runs tests also declares its own bound**: the exact line
+`# reproducer-bound: <seconds>`, a positive whole number covering that build or test run, in the
+same first-10-lines window — `run-reproducer.sh` enforces it on every run in place of its
+20-second default (KAN-831). **The exemption form is available to Minor
 findings only: an Important-severity finding must carry a runnable command** — one that
 `check-panel-reproducers.sh` accepts and the parent can run — and the guard rejects the exemption
 at Important. A demonstrating command needing a pipe, a
diff --git a/stats/internal/guard/runreproducer.go b/stats/internal/guard/runreproducer.go
index e9f9114a..de734441 100644
--- a/stats/internal/guard/runreproducer.go
+++ b/stats/internal/guard/runreproducer.go
@@ -259,7 +259,13 @@ func runReproducer(args []string, env Env, stdout, stderr io.Writer) int {
 	// (skills/flow/review-panel-fix-round.md). A larger bound only lets a run
 	// finish; it never changes what the run's exit code means. The Go tests
 	// inject Env.ReproducerBound and Env.ReproducerGrace instead.
-	bound, boundLabel := rrDeadline(env.ReproducerBound, env.Getenv("RUN_REPRODUCER_BOUND_SECONDS"), rrDefaultBnd)
+	//
+	// A reproducer that declares its own bound (rrDeclaredBound, KAN-831)
+	// replaces the 20-second default with it, so every run of that file --
+	// the exit-contract guard's, the per-finding run and its re-run -- is
+	// bounded alike without each caller remembering the knob; the knob
+	// still wins over the declaration.
+	bound, boundLabel := rrDeadline(env.ReproducerBound, env.Getenv("RUN_REPRODUCER_BOUND_SECONDS"), rrDeclaredBound(resolved))
 	grace, _ := rrDeadline(env.ReproducerGrace, env.Getenv("RUN_REPRODUCER_GRACE_SECONDS"), rrDefaultGrc)
 	// RUN_REPRODUCER_BOUND_FILE, when set, also expires the bound the moment that
 	// path exists. Test-only, and monotone the same way: it can end the wait
@@ -508,6 +514,28 @@ func declaresMutation(path string) bool {
 	return false
 }
 
+// rrBoundPrefix opens the line a reproducer declares its own bound with
+// (KAN-831): `# reproducer-bound: <seconds>`, within the same first 10
+// lines the mutation declaration reads.
+const rrBoundPrefix = "# reproducer-bound: "
+
+// rrDeclaredBound is the seconds a reproducer's own header declares, a
+// positive whole number on a line that is exactly rrBoundPrefix and that
+// number, within its first 10 lines; rrDefaultBnd when none does -- a
+// malformed declaration included, so the run is never left unbounded.
+func rrDeclaredBound(file string) string {
+	for _, l := range headLines(file, 10) {
+		v, ok := strings.CutPrefix(l, rrBoundPrefix)
+		if !ok || v == "" || strings.Trim(v, "0123456789") != "" {
+			continue
+		}
+		if n, err := strconv.Atoi(v); err == nil && n > 0 {
+			return v
+		}
+	}
+	return rrDefaultBnd
+}
+
 // rrDeadline is a RUN_REPRODUCER_* knob: an injected duration wins;
 // otherwise the env value when it is all digits, else the default. The label
 // is the number the bash prints before "s".
diff --git a/stats/internal/guard/runreproducer_test.go b/stats/internal/guard/runreproducer_test.go
index a5c2d94b..d01c765c 100644
--- a/stats/internal/guard/runreproducer_test.go
+++ b/stats/internal/guard/runreproducer_test.go
@@ -452,6 +452,45 @@ func TestRRDeadlineParsesEnvKnob(t *testing.T) {
 	}
 }
 
+// TestRRDeclaredBound pins the `# reproducer-bound: <seconds>` header
+// (KAN-831): a positive whole number on an exact line within the first 10
+// lines is the bound; anything else leaves the 20-second default.
+func TestRRDeclaredBound(t *testing.T) {
+	t.Parallel()
+	dir := t.TempDir()
+	for i, c := range []struct {
+		body, want string
+	}{
+		{"#!/usr/bin/env bash\n# reproducer-bound: 600\nexit 1\n", "600"},
+		{"#!/usr/bin/env bash\nexit 1\n", "20"},
+		{strings.Repeat("# filler\n", 9) + "# reproducer-bound: 300\n", "300"},
+		{strings.Repeat("# filler\n", 10) + "# reproducer-bound: 300\n", "20"},
+		{"# reproducer-bound: 5s\n", "20"},
+		{"# reproducer-bound: 0\n", "20"},
+		{"# reproducer-bound: 90 seconds\n", "20"},
+		{"# reproducer-bound:\n", "20"},
+	} {
+		p := filepath.Join(dir, fmt.Sprintf("r%d.sh", i))
+		writeFile(t, p, c.body)
+		if got := rrDeclaredBound(p); got != c.want {
+			t.Errorf("case %d %q: rrDeclaredBound = %q, want %q", i, c.body, got, c.want)
+		}
+	}
+}
+
+// TestRunReproducerEnforcesDeclaredBound: with no knob and no injected
+// bound, a reproducer declaring `# reproducer-bound: 1` is killed at 1s, not
+// at the 20-second default (KAN-831).
+func TestRunReproducerEnforcesDeclaredBound(t *testing.T) {
+	t.Parallel()
+	wt := rrWorktree(t)
+	rrFixture(t, wt, "scripts/declares-bound.sh", "# reproducer-bound: 1\nsleep 30")
+	env := rrEnv(t)
+	env.ReproducerBound = 0
+	got, out := rrRun(t, env, wt, "scripts/declares-bound.sh")
+	rrExpect(t, got, out, 3, "at the 1s bound")
+}
+
 // TestRunReproducerChildPWDIsWorktree: the bash guard's `cd -- "$WORKTREE"`
 // gave the reproducer PWD=<worktree>. A non-bash reproducer is used because
 // bash resets PWD on its own.
````

  - [ ] **Step 3: RED, then GREEN.** In `stats/internal/guard/runreproducer.go` put the bound call's third argument back to `rrDefaultBnd`:
    `cd stats && go test ./internal/guard -run TestRunReproducerEnforcesDeclaredBound` fails, the reproducer killed at the 20s bound. Restore `rrDeclaredBound(resolved)`: it passes.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -w . && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0;
    `cd stats && go test ./internal/guard -run 'TestRRDeclaredBound|TestRunReproducerEnforcesDeclaredBound|TestRRDeadline|TestRunReproducer$'` passes;
    `scripts/check-mutation-reproducer-pin.sh`, `scripts/check-references.sh`, `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
  - [ ] **Step 5: Commit** the five files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/runreproducer.go`, `stats/internal/guard/runreproducer_test.go`, `scripts/run-reproducer.sh`, `skills/flow/review-panel.md`, `skills/flow/review-panel-fix-round.md`
**Tests:** `TestRRDeclaredBound`, `TestRunReproducerEnforcesDeclaredBound`
**Regression:** reverting the commit removes `rrDeclaredBound` — both tests stop compiling; with only the bound call put back to `rrDefaultBnd`, `TestRunReproducerEnforcesDeclaredBound` sees the reproducer killed at 20s, not 1s.
**Baseline:** before=6 after=8 <!-- measured: grep -c -E '^func Test[A-Za-z0-9_]+\(' stats/internal/guard/runreproducer_test.go @ 9df14302 → 6, and 8 after the diff in a scratch clone -->
**Commit:** feat(guard): bound a reproducer by its own declared seconds
**After:** none

**Decision:** reproducer-declared-bound

- [ ] 3. Re-baseline per capture when the suite fails for other reasons (KAN-837)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n -F 'per capture' skills/flow/verify-fix-loop.md; echo $?`: it prints only `1`.
    <!-- measured: the grep above in a scratch clone @ 9df14302 → exit 1, no line -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 9df14302 after the earlier tasks' diffs; check-references.sh, check-installed-citations.sh, check-markdown-integrity.py and check-dispatch-paragraphs.sh exit 0, and check-verbatim-moves.sh 9df14302 fails only on sentences verbatim-moves.txt lists
diff --git a/skills/flow/verify-fix-loop.md b/skills/flow/verify-fix-loop.md
index fbf2c9a0..03a87761 100644
--- a/skills/flow/verify-fix-loop.md
+++ b/skills/flow/verify-fix-loop.md
@@ -28,6 +28,16 @@ run included, names under `moved:` in its prompt every frame a `<spec>.mockups`
 spec whose baseline moved — the verifier compares each against its mockup exactly as it does the
 change's own frames, so a screen the shared fix regressed is a departure the next round fixes.
 
+**A suite that also fails for a reason the fix did not cause is re-baselined per capture, never per
+flag run** — `--update-snapshots=all` would write the unrelated failures' captures over their
+baselines too. The suite runs without the flag; for each failing capture, every changed-pixel
+blob of its diff image is classified against the fix's expected footprint — where the shared
+source draws and how large its change is — and only a capture whose every blob fits has its
+`-actual.png` copied over its baseline, each copied baseline read before it is staged. A test
+stops at its first failing capture, so the run repeats until no failure inside the footprint
+remains, and a capture with a blob outside it is a defect per **The sweep**
+(`skills/flow-contracts/known-bugs.md`), never copied (KAN-837).
+
 **Mockup departures are classified before the round's task is written.** Where every frame that
 draws the departing component agrees, the departure is a defect this round fixes. Where frames
 disagree on a shared component, the component's own design source wins — the frame that draws it
````

  - [ ] **Step 3: Verify.** `grep -c -F 'KAN-837' skills/flow/verify-fix-loop.md` prints `1`;
    `scripts/check-references.sh`, `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ 9df14302 → 1 -->
  - [ ] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/verify-fix-loop.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): re-baseline per capture when the suite fails for other reasons
**After:** none

**Decision:** per-capture-blob-rebaseline

- [ ] 4. Name a harness agent worktree FOREIGN, never STRAY (KAN-883)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `d=$(mktemp -d) && git -C "$d" init -q r && git -C "$d/r" commit -q --allow-empty -m i && git -C "$d/r" worktree add -q "$d/r/.claude/worktrees/agent-a1" -b a1 && scripts/check-worktree-location.sh "$d/r"; echo $?; rm -rf "$d"`:
    it prints a `STRAY:` line for `.claude/worktrees/agent-a1`, a `LOCATION-STRAY:` line counting `1`, and `1`.
    <!-- measured: the command above, under the scratchpad instead of mktemp's default, in a scratch clone @ 9df14302 → STRAY …/r/.claude/worktrees/agent-a1 (refs/heads/a1), LOCATION-STRAY … — 1, exit 1 -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 9df14302 after Tasks 1-3's diffs; gofmt -l and go vet ./internal/guard clean; go test ./internal/guard -run 'TestCheckWorktreeLocation|TestCheckFinishPreflight' passes; three TestCheckWorktreeLocation subtests fail with the harness case's prefix made unmatchable
diff --git a/scripts/check-worktree-location.sh b/scripts/check-worktree-location.sh
index 1bba66a4..5e2a28ce 100755
--- a/scripts/check-worktree-location.sh
+++ b/scripts/check-worktree-location.sh
@@ -9,6 +9,7 @@
 #
 # Prints one STRAY line per offender, then ONE verdict line:
 #   STRAY: <path> (<branch>|detached)      per worktree outside <project>-worktrees/
+#   FOREIGN: <path> (<branch>|detached)    per harness agent worktree, never counted
 #   LOCATION-OK: <project>                 every worktree is at or under it
 #   LOCATION-STRAY: <project> — <n>        <n> worktree(s) are not
 #
@@ -33,6 +34,13 @@
 # test would report <project>-worktrees-old/x as in-tree. <project>/.worktrees/,
 # the retired in-repo layout, is STRAY like any other path.
 #
+# A HARNESS AGENT'S WORKTREE IS NOT FLOW'S (KAN-883). Claude Code creates a
+# subagent's isolated worktree under <project>/.claude/worktrees/ and locks
+# it while that session lives; no flow command creates one there, and a run
+# cannot remove another session's. Each is named on a FOREIGN line, never
+# counted, so it neither stops a kickoff nor reaches check-finish-preflight's
+# STRAY relay. Every other path outside <project>-worktrees/ is STRAY.
+#
 # flow-guard is built from this checkout, never taken from PATH:
 # scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
 # cannot-answer code) with the cause when it cannot.
diff --git a/stats/internal/guard/check_worktree_location_test.go b/stats/internal/guard/check_worktree_location_test.go
index c2975857..93d2137f 100644
--- a/stats/internal/guard/check_worktree_location_test.go
+++ b/stats/internal/guard/check_worktree_location_test.go
@@ -50,7 +50,7 @@ func TestCheckWorktreeLocation(t *testing.T) {
 
 	verdicts := []struct {
 		name   string
-		setup  func(t *testing.T, repo string) string // returns the expected STRAY lines
+		setup  func(t *testing.T, repo string) string // returns the expected STRAY and FOREIGN lines
 		wantRC int
 	}{
 		{"case 1: ok with no worktree", func(*testing.T, string) string { return "" }, 0},
@@ -82,6 +82,18 @@ func TestCheckWorktreeLocation(t *testing.T) {
 			gitRun(t, r, "worktree", "add", "-q", "--detach", r+"-worktrees-old/x")
 			return "STRAY: " + r + "-worktrees-old/x (detached)\n"
 		}, 1},
+		// KAN-883: a harness agent's worktree under <repo>/.claude/worktrees/
+		// is named FOREIGN and never counted, beside a real stray that is.
+		{"harness agent worktree is FOREIGN, not counted", func(t *testing.T, r string) string {
+			d := physicalTempDir(t)
+			gitRun(t, r, "worktree", "add", "-q", r+"/.claude/worktrees/agent-a1", "-b", "worktree-agent-a1")
+			gitRun(t, r, "worktree", "add", "-q", "--detach", d+"/x")
+			return "FOREIGN: " + r + "/.claude/worktrees/agent-a1 (refs/heads/worktree-agent-a1)\nSTRAY: " + d + "/x (detached)\n"
+		}, 1},
+		{"only a harness agent worktree -> ok", func(t *testing.T, r string) string {
+			gitRun(t, r, "worktree", "add", "-q", r+"/.claude/worktrees/agent-a1", "-b", "worktree-agent-a1")
+			return "FOREIGN: " + r + "/.claude/worktrees/agent-a1 (refs/heads/worktree-agent-a1)\n"
+		}, 0},
 		// Beyond the harness: two strays are counted, in porcelain order.
 		{"two strays counted", func(t *testing.T, r string) string {
 			d := physicalTempDir(t)
@@ -98,7 +110,7 @@ func TestCheckWorktreeLocation(t *testing.T) {
 			strays := tc.setup(t, repo)
 			want := strays + "LOCATION-OK: " + repo + "\n"
 			if tc.wantRC == 1 {
-				want = strays + "LOCATION-STRAY: " + repo + " — " + strconv.Itoa(strings.Count(strays, "\n")) + "\n"
+				want = strays + "LOCATION-STRAY: " + repo + " — " + strconv.Itoa(strings.Count(strays, "STRAY: ")) + "\n"
 			}
 			r := runGuard("check-worktree-location", []string{repo}, env)
 			if r.rc != tc.wantRC || r.stdout != want || r.err != "" {
diff --git a/stats/internal/guard/worktreelocation.go b/stats/internal/guard/worktreelocation.go
index a5fcd2d0..20538bd5 100644
--- a/stats/internal/guard/worktreelocation.go
+++ b/stats/internal/guard/worktreelocation.go
@@ -61,7 +61,13 @@ func checkWorktreeLocation(args []string, env Env, stdout, stderr io.Writer) int
 	// The appended newline restores the blank line that terminates the LAST
 	// record, which capture's trailing-newline strip removed (the bash's
 	// `printf '%s\n\n'`), so the final worktree's record reaches the check.
-	inTree := siblingRoot(root)
+	// A HARNESS AGENT'S WORKTREE IS NOT FLOW'S (KAN-883). Claude Code
+	// creates a subagent's isolated worktree under <project>/.claude/worktrees/
+	// and locks it while that session lives; no flow command creates one
+	// there, and a run cannot remove another session's. Such a worktree is
+	// named on a FOREIGN line and never counted, so it neither stops a
+	// kickoff nor reaches check-finish-preflight's STRAY relay.
+	inTree, harness := siblingRoot(root), filepath.Join(root, ".claude", "worktrees")+"/"
 	n, count := 0, 0
 	w, b := "", "detached"
 	for _, line := range strings.Split(porcelain+"\n", "\n") {
@@ -74,7 +80,11 @@ func checkWorktreeLocation(args []string, env Env, stdout, stderr io.Writer) int
 				b = f[1]
 			}
 		case line == "":
-			if n > 1 && w != inTree && !strings.HasPrefix(w, inTree+"/") {
+			switch {
+			case n <= 1 || w == inTree || strings.HasPrefix(w, inTree+"/"):
+			case strings.HasPrefix(w, harness):
+				fmt.Fprintf(stdout, "FOREIGN: %s (%s)\n", w, b)
+			default:
 				fmt.Fprintf(stdout, "STRAY: %s (%s)\n", w, b)
 				count++
 			}
````

  - [ ] **Step 3: RED, then GREEN.** In `stats/internal/guard/worktreelocation.go` change `case strings.HasPrefix(w, harness):` to `case strings.HasPrefix(w, harness+"never"):`:
    `cd stats && go test ./internal/guard -run TestCheckWorktreeLocation` fails on both new subtests. Restore the line: it passes.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -w . && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0;
    `cd stats && go test ./internal/guard -run 'TestCheckWorktreeLocation|TestCheckFinishPreflight'` passes; Step 1's command prints a `FOREIGN:` line, a `LOCATION-OK:` line and `0`.
  - [ ] **Step 5: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/worktreelocation.go`, `stats/internal/guard/check_worktree_location_test.go`, `scripts/check-worktree-location.sh`
**Tests:** `harness agent worktree is FOREIGN, not counted`, `only a harness agent worktree -> ok`
**Regression:** reverting the commit counts the harness agent worktree as STRAY — both subtests fail, the first on its `FOREIGN:` line and count, the second on its exit code.
**Baseline:** before=13 after=15 <!-- measured: grep -c -E '^[[:space:]]+\{"' stats/internal/guard/check_worktree_location_test.go @ 9df14302 → 13, and 15 after the diff in a scratch clone -->
**Commit:** fix(guard): name a harness agent worktree FOREIGN, never STRAY
**After:** none

**Decision:** harness-worktree-foreign

- [ ] 5. Measure plan-class's diff side from the fork point (KAN-885)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `d=$(mktemp -d) && git -C "$d" init -q r && git -C "$d/r" commit -q --allow-empty -m i && git -C "$d/r" checkout -q -b moved && seq 1 40 > "$d/r/src.go" && git -C "$d/r" add -A && git -C "$d/r" commit -q -m moved && git -C "$d/r" checkout -q - && printf -- '- [ ] 1. Docs\n**Files:** `docs/note.md`\n' > "$d/tasks.md" && scripts/plan-class.sh "$d/tasks.md" 1 "$d/r" moved | grep '^class:'; rm -rf "$d"`:
    it prints `class: small` — the moved base's `src.go` counted as this change's surface.
    <!-- measured: the command above, under the scratchpad instead of mktemp's default, in a scratch clone @ 9df14302 → class: small; after the diff → class: micro -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 9df14302 after Tasks 1-4's diffs; gofmt -l and go vet ./internal/guard clean; go test ./internal/guard -run TestPlanClass passes; the moved-base subtest fails with the mergebase = fork assignment replaced by _ = fork
diff --git a/scripts/plan-class.sh b/scripts/plan-class.sh
index 7cfd9b87..01f4cd2f 100755
--- a/scripts/plan-class.sh
+++ b/scripts/plan-class.sh
@@ -68,7 +68,11 @@
 #            committed-since-merge-base, staged and unstaged) entirely
 #            documentation and at most planclass.go's pcMicroLineCap (20)
 #            changed lines; an empty
-#            touched surface passes vacuously. Without the optional arguments
+#            touched surface passes vacuously. Both are measured from
+#            `git merge-base HEAD <merge-base>`, never from <merge-base>
+#            itself, so a base ref that moved after the worktree was cut
+#            counts none of its own new commits (KAN-885); a <merge-base>
+#            sharing no history with HEAD is exit 2. Without the optional arguments
 #            micro never fires, so the two-argument form classifies exactly as
 #            before it existed. Added by
 #            kan-617-flow-cost-the-decide-record-ceremony-runs-full so a
diff --git a/stats/internal/guard/plan_class_test.go b/stats/internal/guard/plan_class_test.go
index abfbc7da..d0690953 100644
--- a/stats/internal/guard/plan_class_test.go
+++ b/stats/internal/guard/plan_class_test.go
@@ -246,6 +246,32 @@ func TestPlanClass(t *testing.T) {
 			writeFile(t, r.dir+"/note.md", pcSeq(11, 20))
 			classIs(t, run(pcTasks(t, "micro-churn-single-count", pcDocs), "1", r.dir, mb), "micro")
 		}},
+		// KAN-885: a base ref that moved after the worktree was cut is
+		// measured from HEAD's fork point with it, so the base's own new
+		// code commit is not this change's surface.
+		{"moved base ref measured from the fork point -> micro", func(t *testing.T) {
+			r := newRepo(t)
+			var g fxGit
+			g.git(r.dir, "checkout", "-q", "-b", "moved")
+			writeFile(t, r.dir+"/src.go", pcSeq(1, 40))
+			g.git(r.dir, "add", "-A")
+			g.git(r.dir, "commit", "-q", "-m", "base moved")
+			g.git(r.dir, "checkout", "-q", "-")
+			if g.err != nil {
+				t.Fatal(g.err)
+			}
+			classIs(t, run(pcTasks(t, "micro-moved-base", pcDocs), "1", r.dir, "moved"), "micro")
+		}},
+		{"merge base with no shared history -> exit 2", func(t *testing.T) {
+			r := newRepo(t)
+			var g fxGit
+			// A parentless commit of the empty tree: a root of its own.
+			root := g.git(r.dir, "commit-tree", "-m", "unrelated root", "4b825dc642cb6eb9a060e54bf8d69288fbee4904")
+			if g.err != nil {
+				t.Fatal(g.err)
+			}
+			exit2(t, run(pcTasks(t, "micro-unrelated-base", pcDocs), "1", r.dir, root))
+		}},
 		{"missing file -> exit 2", func(t *testing.T) {
 			exit2(t, run("/nonexistent/tasks.md", "1"))
 		}},
diff --git a/stats/internal/guard/planclass.go b/stats/internal/guard/planclass.go
index 93f75cf6..502a7d66 100644
--- a/stats/internal/guard/planclass.go
+++ b/stats/internal/guard/planclass.go
@@ -257,6 +257,18 @@ func planClass(args []string, env Env, stdout, stderr io.Writer) int {
 		if !ok || !panelValidateWorktree(env, "plan-class", worktree, mergebase, gitBin, stderr) {
 			return 2
 		}
+		// The diff side is measured from HEAD's fork point with the given
+		// base, never from the base itself (KAN-885): a ref the caller
+		// resolved after the base moved — origin/main fetched mid-run — is
+		// not an ancestor of HEAD, and diffing against its tip counted the
+		// base's own new commits as this change's surface. A recorded
+		// worktree sha, already an ancestor, is its own fork point.
+		fork, ok := capture(panelGit(env, gitBin, worktree, "merge-base", "HEAD", mergebase))
+		if !ok {
+			fmt.Fprintf(stderr, "plan-class: HEAD and '%s' share no merge base in %s\n", mergebase, worktree)
+			return 2
+		}
+		mergebase = fork
 		surface, ok := panelTouchedPaths(env, "plan-class", worktree, mergebase, gitBin, stderr)
 		if !ok {
 			return 2
````

  - [ ] **Step 3: RED, then GREEN.** In `stats/internal/guard/planclass.go` replace `mergebase = fork` with `_ = fork`:
    `cd stats && go test ./internal/guard -run TestPlanClass` fails on `moved base ref measured from the fork point -> micro`. Restore the line: it passes.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -w . && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0;
    `cd stats && go test ./internal/guard -run TestPlanClass` passes; Step 1's command prints `class: micro`.
  - [ ] **Step 5: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/planclass.go`, `stats/internal/guard/plan_class_test.go`, `scripts/plan-class.sh`
**Tests:** `moved base ref measured from the fork point -> micro`, `merge base with no shared history -> exit 2`
**Regression:** reverting the commit measures from the moved ref itself — the first subtest classifies `small`, and the second classifies instead of exiting 2.
**Baseline:** before=41 after=43 <!-- measured: grep -c -E '^[[:space:]]+\{"' stats/internal/guard/plan_class_test.go @ 9df14302 → 41, and 43 after the diff in a scratch clone -->
**Commit:** fix(guard): measure plan-class's diff side from the fork point
**After:** none

**Decision:** plan-class-fork-point

- [ ] 6. Grep check-task-records harness output through here-strings
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `cd stats && go test ./internal/guard -run TestNoPipeIntoEarlyExitGrep`: it fails naming `test-check-task-records.sh:392` and `:411`.
    <!-- measured: the test above in a scratch clone @ 9df14302 → FAIL, test-check-task-records.sh:392 and :411 -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 9df14302 after Tasks 1-5's diffs; go test ./internal/guard -run TestNoPipeIntoEarlyExitGrep passes and bash scripts/test-check-task-records.sh prints all cases passed
diff --git a/scripts/test-check-task-records.sh b/scripts/test-check-task-records.sh
index 6e0c8324..631af73d 100755
--- a/scripts/test-check-task-records.sh
+++ b/scripts/test-check-task-records.sh
@@ -389,7 +389,7 @@ DIRS+=("$STUB")
 printf '#!/bin/sh\nexit 1\n' >"$STUB/flow"
 chmod +x "$STUB/flow"
 PATH="$STUB:$PATH" run_guard "$FIX"
-if [ "$RC" -eq 1 ] && printf '%s' "$OUT" | grep -q "no commit with subject 'fix(peer): one'"; then
+if [ "$RC" -eq 1 ] && grep -q "no commit with subject 'fix(peer): one'" <<<"$OUT"; then
   pass "case 18 unreadable state record keeps the violation"
 else
   fail "case 18 failing flow: rc=$RC out=$OUT"
@@ -408,7 +408,7 @@ DIRS+=("$STUB")
 printf '#!/bin/sh\nprintf %s\n' "'{\"worktrees\":{\"$PLAIN\":\"x\"}}\\n'" >"$STUB/flow"
 chmod +x "$STUB/flow"
 PATH="$STUB:$PATH" run_guard "$FIX"
-if [ "$RC" -eq 1 ] && printf '%s' "$OUT" | grep -q "no commit with subject 'fix(peer): one'"; then
+if [ "$RC" -eq 1 ] && grep -q "no commit with subject 'fix(peer): one'" <<<"$OUT"; then
   pass "case 19 non-repository record path is skipped"
 else
   fail "case 19 non-repo path: rc=$RC out=$OUT"
````

  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/guard -run TestNoPipeIntoEarlyExitGrep` passes;
    `bash scripts/test-check-task-records.sh` ends `all cases passed`.
  - [ ] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `scripts/test-check-task-records.sh`
**Tests:** none — repairs existing tests: cases 18 and 19 of `scripts/test-check-task-records.sh`, whose two lines `TestNoPipeIntoEarlyExitGrep` flags.
**Regression:** none — no test is declared.
**Commit:** fix(scripts): grep check-task-records harness output through here-strings
**After:** none
