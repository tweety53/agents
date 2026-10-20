# self-review-kan-1000-speedups

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the six KAN-1000 speed findings of the self-review pass.

**Architecture:** prose in the canonical flow contract for each finding, one verdict change in the
`check-verify-report` Go guard with its test and shim header, and one fork deny in
`hooks/enforce-agent-baseline.py` with its harness cases.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-20-self-review-kan-1000-speedups-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Work in the worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1000-speedups`, branch `self-review-kan-1000-speedups`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- **The worktree holds a stopped attempt's uncommitted edits** to `scripts/check-verify-report.sh`, `skills/flow-contracts/operator-prompts-auto-resolution.md`, `skills/flow/verify-fix-loop.md`, `skills/flow/visual-verify-verifier.md`, `skills/flow/visual-verify.md`, `stats/internal/guard/check_verify_report_test.go` and `stats/internal/guard/verifyreport.go` (decision `draft-superseded`). Before Task 1, save them with `git diff > <your scratch dir>/kan-1000-draft.patch` — a path outside the worktree — then `git checkout -- ` those seven paths, and confirm `git status --short` lists nothing but untracked files. The task diffs carry every part of the draft this plan keeps.
- Each task's diff is the unified diff its apply step gives, applied with `git apply` from the worktree's root, in task order — the diffs share files and apply only in order. Save each by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork. Each prose sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- Task 1: a `10 blocked` citing `compose-mockup-frames exit <n>` with step 8 done stays admissible, and the existing all-blocked case (`blocked on a cited exit`) stays complete.
- Task 6: a non-fork flow dispatch is never denied by the new branch, a fork in a session that ran only `/flow-status` or `/flow-settings` is admitted, and an unreadable transcript fails open.
- Task 2: a round with no round base still runs `specs` unscoped; only a run under a round base loses the whole-suite fallback.
- Task 5: colour, a missing or extra element, and copy stay outside the allowance; the band and seam pairings' thresholds are untouched.

- [x] 1. Refuse a step 10 the verifier blocked with its captures on disk
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Save Step 2's diff, apply its
    test hunk alone (`git apply --include='stats/internal/guard/check_verify_report_test.go' <the saved diff>`)
    and run `cd stats && go test ./internal/guard/ -run 'TestCheckVerifyReport' -count=1`: the case
    `step 10 blocked with captures on disk` fails, the guard answering `VERIFY-REPORT-COMPLETE`.
    <!-- measured: the test hunk alone, go test ./internal/guard/ -run TestCheckVerifyReport @ 09bec571 → FAIL TestCheckVerifyReport/step_10_blocked_with_captures_on_disk -->
  - [x] **Step 2: Apply the fix** — the rest (`git apply --exclude='stats/internal/guard/check_verify_report_test.go' <the saved diff>`):

````diff verified:applied in order with git apply to a scratch clone at 09bec571; the scratch tree with all six applied passes the checks design.md ## Context lists
diff --git a/scripts/check-verify-report.sh b/scripts/check-verify-report.sh
index 083a3ce0..2c762839 100755
--- a/scripts/check-verify-report.sh
+++ b/scripts/check-verify-report.sh
@@ -16,7 +16,9 @@
 # (skills/flow/visual-verify-verifier.md). A status is admissible when it is
 #   done
 #   n/a — <reason>       steps 4 and 10 only, and `motion` when <motions named> is 0
-#   blocked — <reason>   the reason citing a non-zero `exit <n>`
+#   blocked — <reason>   the reason citing a non-zero `exit <n>`; step 10 only
+#                        when step 8 is blocked too or the reason cites
+#                        `compose-mockup-frames exit <n>` (KAN-1000)
 # and `motion done` with <motions named> above 0 also needs a
 # `- motion: <n>/<n>` line, <n> that count — a strip recorded beyond the
 # named motions goes on a sub-bullet, never into <n>. Every other status — absent,
diff --git a/skills/flow/visual-verify.md b/skills/flow/visual-verify.md
index 3415aeec..1fe7d6d6 100644
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -84,7 +84,8 @@ check-verify-report.sh <abs-worktree>/.superpowers/sdd/verify-report-<key>.md <m
 `<motions named>` is the count of motions this dispatch's prompt named, `0` for `motions: none`;
 the guard's header is canonical for the report's `steps:` line and its admissible statuses. Exit 0
 → the report is complete; read it. Exit 1 → **the report is incomplete and is never accepted,
-whatever its prose says**: close the dispatch `-outcome aborted` and dispatch its remainder at
+whatever its prose says**, and no fix task is planned from it, its departures included, until its
+remainder completes it (KAN-1000): close the dispatch `-outcome aborted` and dispatch its remainder at
 once, recorded under the incomplete dispatch's own key suffixed `-rest` before any `-<repository
 name>` suffix, its prompt identical to the incomplete dispatch's plus that report verbatim
 under a `## Incomplete attempt` heading and the guard's verdict line, which names the steps it
diff --git a/stats/internal/guard/check_verify_report_test.go b/stats/internal/guard/check_verify_report_test.go
index 7a97b718..cd6b1b56 100644
--- a/stats/internal/guard/check_verify_report_test.go
+++ b/stats/internal/guard/check_verify_report_test.go
@@ -27,6 +27,9 @@ func TestCheckVerifyReport(t *testing.T) {
 		{"a step left out", "- steps: 4 done | 7 done | 8 done | 10 done | 11 done | motion n/a — none\n", "0", 1, "9 absent"},
 		{"n/a on a required step", "- steps: 4 done | 7 done | 8 done | 9 n/a — out of time | 10 done | 11 done | motion n/a — none\n", "0", 1, "9 n/a — out of time"},
 		{"n/a with no reason", "- steps: 4 n/a | 7 done | 8 done | 9 done | 10 done | 11 done | motion n/a — none\n", "0", 1, "4 n/a"},
+		// KAN-1000: step 10 declared blocked on another step's exit while step 8 captured.
+		{"step 10 blocked with captures on disk", "- steps: 4 done | 7 blocked — verify exit 1 | 8 done | 9 done | 10 blocked — verify exit 1 | 11 done | motion n/a — none\n", "0", 1, "10 blocked — verify exit 1 (step 8 captured)"},
+		{"step 10 blocked on compose's own exit", "- steps: 4 done | 7 done | 8 done | 9 done | 10 blocked — compose-mockup-frames exit 1 | 11 done | motion n/a — none\n", "0", 0, "COMPLETE"},
 		{"blocked with no cited exit", "- steps: 4 done | 7 done | 8 done | 9 blocked — ran out of context | 10 done | 11 done | motion n/a — none\n", "0", 1, "9 blocked — ran out of context"},
 		{"blocked on exit 0", "- steps: 4 done | 7 done | 8 done | 9 blocked — capture exit 0 | 10 done | 11 done | motion n/a — none\n", "0", 1, "9 blocked — capture exit 0"},
 		{"motions named, step n/a", allDone, "1", 1, "motion n/a — no motions named"},
diff --git a/stats/internal/guard/verifyreport.go b/stats/internal/guard/verifyreport.go
index e2047e88..6c5727b6 100644
--- a/stats/internal/guard/verifyreport.go
+++ b/stats/internal/guard/verifyreport.go
@@ -21,6 +21,8 @@ var (
 	vrStepsLine = regexp.MustCompile(`^\s*- steps:\s*(.*)$`)
 	vrMotion    = regexp.MustCompile(`^\s*- motion:\s*(\d+)/(\d+)\b`)
 	vrCitedExit = regexp.MustCompile(`\bexit [1-9][0-9]*\b`)
+	// Step 10's own blocking exit (skills/flow/visual-verify-verifier.md step 10).
+	vrComposeExit = regexp.MustCompile(`\bcompose-mockup-frames(\.sh)? exit [1-9][0-9]*\b`)
 )
 
 func checkVerifyReport(args []string, env Env, stdout, stderr io.Writer) int {
@@ -65,6 +67,11 @@ func checkVerifyReport(args []string, env Env, stdout, stderr io.Writer) int {
 			undone = append(undone, fmt.Sprintf("motion: no `- motion: %d/%d` line", motions, motions))
 		case s == "done":
 		case isNA && naAllowed && strings.TrimSpace(reason) != "":
+		// Step 10 measures the captures step 8 wrote, so it blocks only when step 8 did or
+		// on compose-mockup-frames' own exit (KAN-1000: two verifiers skipped step 10 with
+		// every capture on disk, and departures surfaced one round at a time).
+		case isBlocked && id == "10" && !strings.HasPrefix(status["8"], "blocked — ") && !vrComposeExit.MatchString(cause):
+			undone = append(undone, strings.TrimSpace(id+" "+s)+" (step 8 captured)")
 		case isBlocked && vrCitedExit.MatchString(cause):
 		default:
 			undone = append(undone, strings.TrimSpace(id+" "+s))
````

  - [x] **Step 3: Verify.** `cd stats && gofmt -l internal/guard && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckVerifyReport' -count=1` passes, `gofmt -l` printing nothing; then from the root `scripts/check-references.sh && scripts/check-installed-citations.sh && scripts/check-markdown-integrity.py && scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 4: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/verifyreport.go`, `stats/internal/guard/check_verify_report_test.go`, `scripts/check-verify-report.sh`, `skills/flow/visual-verify.md`
**Tests:** `step 10 blocked with captures on disk`, `step 10 blocked on compose's own exit`
**Regression:** `step 10 blocked with captures on disk` — reverting the commit admits `10 blocked — verify exit 1` beside `8 done`, and the case fails on `VERIFY-REPORT-COMPLETE`; `step 10 blocked on compose's own exit` pins that compose's own exit stays admissible.
**Baseline:** before=16 after=18
<!-- measured: grep -cE '^[[:space:]]+\{"' stats/internal/guard/check_verify_report_test.go @ 09bec571 → 16, and → 18 with Task 1 applied -->
**Commit:** fix(guard): refuse a step 10 the verifier blocked with its captures on disk
**After:** none

**Decision:** step10-block-admissible
**Decision:** not-done-already-enforced
**Decision:** no-plan-from-incomplete

- [x] 2. Keep a scoped round's verify to the round's own specs
  - [x] **Step 1: Apply** the diff:

````diff verified:applied in order with git apply to a scratch clone at 09bec571; the scratch tree with all six applied passes the checks design.md ## Context lists
diff --git a/skills/flow/verify-fix-loop.md b/skills/flow/verify-fix-loop.md
index e9a6787b..3077df99 100644
--- a/skills/flow/verify-fix-loop.md
+++ b/skills/flow/verify-fix-loop.md
@@ -127,8 +127,10 @@ One round runs these steps in order:
 **One full run closes the loop.** A scoped re-run whose final report carries no fixable defect is
 followed by **Visual verification** once more from its `begin` mark, unscoped — no round base,
 every motion of the change — its dispatch keys suffixed `-fix-<k>-full` before any `-<repository
-name>` suffix. A fixable defect it reports opens the next round; its clean report is the one the
-handoff reports.
+name>` suffix. A fixable defect it reports opens the next round, whose step-4 re-run is scoped
+like any round's, its round base the HEAD this full run verified: a `-full` run follows only a
+clean scoped re-run, never another round's fix (KAN-1000: three rounds in a row re-ran unscoped).
+Its clean report is the one the handoff reports.
 
 ## No cap
 
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index a8db84a2..32d45ee3 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -19,7 +19,12 @@ never an exit code the command did not return.
    Under a round base the output feeds `verify` alone, and an empty one runs no `verify` spec — the
    report's `verify:` line reads `not run — no spec since the round base` and step 7 reports `done`; step 8 takes its capture
    spec from `specs` run against the merge base.
-   A `specs` with no `<merge-base>` cannot take a round base: it runs as declared, its full list,
+   **Under a round base `verify` never runs the whole suite**: a non-zero `specs` exit, or a
+   `specs` with no `<merge-base>`, runs no `verify` spec, and the report's `verify:` line reads
+   `not run — specs cannot scope to the round base`; the whole suite runs once, in the loop's
+   closing unscoped run (**The loop**, `skills/flow/verify-fix-loop.md`)
+   (KAN-1000: every round re-ran all 88 specs, about 14 minutes each).
+   Without a round base, a `specs` with no `<merge-base>` runs as declared, its full list,
    and the report names `specs: not scoped — specs has no <merge-base>`.
    **A prompt naming `rerun specs:` runs those specs alone** — substituted for `<specs>` in place
    of `specs`' output, which is not run — and the report's `verify:` line reads `exit <n> — rerun
````

  - [x] **Step 2: Verify.** From the root, `scripts/check-references.sh && scripts/check-installed-citations.sh && scripts/check-markdown-integrity.py && scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 3: Commit** the two files with the subject below.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`, `skills/flow/verify-fix-loop.md`
**Tests:** **none** — prose rules; `check-verbatim-moves.sh` covers their acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): keep a scoped round's verify to the round's own specs
**After:** Task 1

**Decision:** scoped-verify-never-whole
**Decision:** gymie-specs-fallback-out-of-scope

- [x] 3. Have the loop's implementer measure its fix against the frame
  - [x] **Step 1: Apply** the diff:

````diff verified:applied in order with git apply to a scratch clone at 09bec571; the scratch tree with all six applied passes the checks design.md ## Context lists
diff --git a/skills/flow/verify-fix-loop.md b/skills/flow/verify-fix-loop.md
index 3077df99..3a1964b4 100644
--- a/skills/flow/verify-fix-loop.md
+++ b/skills/flow/verify-fix-loop.md
@@ -70,6 +70,15 @@ One round runs these steps in order:
    a test that drives the real composable or component, never a hand-copied constant, and asserts
    the value the verifier measured against the frame — so a later change that moves it again fails
    a test instead of waiting for the next visual verification (KAN-790).
+   **The implementer measures before it hands back.** When the round's defects include a mockup
+   departure, its dispatch names the absolute path of `skills/flow/visual-verify-verifier.md`, the
+   capture spec with its `<spec>.mockups` sidecar, the resolved `mockups` root and `mockup frame`
+   value, and the project's `capture` command, and tells it: on every view its fix touched,
+   recapture against the stack rebuilt from its fix, as step 4's confirming re-run rebuilds it, and
+   run step 10's measurements on each — `compose-mockup-frames.sh` and the band, seam and matrix
+   reads of `measure-visual-properties.sh` — fixing until they come back clean or each leftover is
+   named with its cause, and report every view's measured numbers (KAN-1000: rounds 2–5 each found
+   a new 1–3px offset the previous fix had judged by eye, 45–70 minutes a round).
    **A task fixing a test that passes alone and fails only in the full suite closes after two
    more full runs.** Such load flakes surface one per full-suite round, so after the task's commit
    and before `close-task.sh` the parent runs the worktree's whole `## test` list twice more, and
````

  - [x] **Step 2: Verify.** From the root, `scripts/check-references.sh && scripts/check-installed-citations.sh && scripts/check-markdown-integrity.py && scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 3: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/verify-fix-loop.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): have the loop's implementer measure its fix against the frame
**After:** Task 2

**Decision:** implementer-measures-prose

- [x] 4. Check a feature-branch base before the first verify stage
  - [x] **Step 1: Apply** the diff:

````diff verified:applied in order with git apply to a scratch clone at 09bec571; the scratch tree with all six applied passes the checks design.md ## Context lists
diff --git a/skills/flow-contracts/operator-prompts-auto-resolution.md b/skills/flow-contracts/operator-prompts-auto-resolution.md
index 426adf23..28177c9a 100644
--- a/skills/flow-contracts/operator-prompts-auto-resolution.md
+++ b/skills/flow-contracts/operator-prompts-auto-resolution.md
@@ -41,7 +41,12 @@ cases is a stop for nothing (KAN-1000):
 - **Failures that are already on the base:** attribute them to their cause first, then fix them in this
   run (**The sweep**, `skills/flow-contracts/known-bugs.md`). If the cause is a base branch that has
   fallen behind the branch it tracks (a feature branch that never took in develop or main), the
-  recommended option is to sync that base and re-run. Pushing the synced base is still confirmed,
+  recommended option is to sync that base and re-run. The parent checks this at the start of
+  whichever of `flow.verify` and `flow.visual-verify` runs first, before any failure arrives: for
+  each repository whose recorded base is a feature branch, after `git fetch origin`, a
+  `git rev-list --count origin/<base>..origin/<the branch it tracks>` above 0 — develop or main
+  holding commits the base lacks — syncs the base first (KAN-1000: a stale feature/cardio base
+  cost 61 failures and about an hour). Pushing the synced base is still confirmed,
   per **What still stops**.
 - **A departure from the mockup:** match the mockup (`rules/design-mockups-are-specs.mdc`). An element
   the mockup does not draw stays only when the screen cannot work without it. Record that reason
````

  - [x] **Step 2: Verify.** From the root, `scripts/check-references.sh && scripts/check-installed-citations.sh && scripts/check-markdown-integrity.py && scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 3: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow-contracts/operator-prompts-auto-resolution.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — no test is declared.
**Commit:** docs(flow-contracts): check a feature-branch base before the first verify stage
**After:** none

**Decision:** base-check-in-auto-resolution

- [x] 5. Let a 1dp position departure stand only after a fix round tried it
  - [x] **Step 1: Apply** the diff:

````diff verified:applied in order with git apply to a scratch clone at 09bec571; the scratch tree with all six applied passes the checks design.md ## Context lists
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index 32d45ee3..ccde28db 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -231,6 +231,9 @@ never an exit code the command did not return.
    carries no baseline PNG for: every departure on either is this change's, and the report names
    the missing baseline. Every line that reports the departure names its cause or its attribution
    beside it — never a bare `departure`.
+   A departure within 1dp is reported like any other, its measured value beside it: whether it
+   stands is the parent's call under the Blocking paragraph of **Steps 3–13**
+   (`skills/flow/visual-verify.md`), never yours (KAN-1000).
    **A part the frame's sidecar comment declares expected until a named sibling ticket lands is
    no departure.** The line reporting the frame names that part
    `expected — <sibling ticket>, <decision ID>` and nothing blocks on it, while every part the
diff --git a/skills/flow/visual-verify.md b/skills/flow/visual-verify.md
index 1fe7d6d6..8e926529 100644
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -327,6 +327,14 @@ with no strip under `<changeRoot>/visual-verification/motion/` or no entry in
 `visual-verification.md`, and a frame its strip shows breaking the motion step's rule, block as a
 departure would.**
 
+**A position departure of at most 1dp — 2px at 2x — on a band or an element may stand, and only
+when all of these hold:** a fix round has already tried to close it at its source; the run records
+why no easy fix exists, such as a font's glyph metrics; `visual-verification.md` carries it beside
+the frame with its measured value and that reason; and the handoff's `**Decisions:**` lines name it.
+It is never the first answer, and colour, a missing or extra element, and copy get no tolerance
+(KAN-1000: the operator states this approximation as the requirement, the one widening
+`rules/fix-determinism-at-the-source.mdc` admits).
+
 **Every blocking item but the pre-flight, a stack that could not be started and the fingerprint
 is the branch's own defect, never a `## Question`:** it takes **The loop**
 (`skills/flow/verify-fix-loop.md`). Those three are the environment's and end the stage as their
````

  - [x] **Step 2: Verify.** From the root, `scripts/check-references.sh && scripts/check-installed-citations.sh && scripts/check-markdown-integrity.py && scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 3: Commit** the two files with the subject below.

**Build:** green
**Files:** `skills/flow/visual-verify.md`, `skills/flow/visual-verify-verifier.md`
**Tests:** **none** — prose rules; `check-verbatim-moves.sh` covers their acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): let a 1dp position departure stand only after a fix round tried it
**After:** Task 1, 2

**Decision:** allowance-in-blocking

- [x] 6. Deny a fork dispatch in a flow session
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Save Step 2's diff, apply its
    harness hunk alone (`git apply --include='scripts/test-enforce-agent-baseline-hook.sh' <the saved diff>`)
    and run `bash scripts/test-enforce-agent-baseline-hook.sh`: cases `10 fork naming a flow skill`
    and `10c fork in a /flow session` fail, the hook admitting both.
    <!-- measured: the harness hunk alone, bash scripts/test-enforce-agent-baseline-hook.sh @ 09bec571 → FAIL 10 and 10c, exit 1 -->
  - [x] **Step 2: Apply the fix** — the rest (`git apply --exclude='scripts/test-enforce-agent-baseline-hook.sh' <the saved diff>`):

````diff verified:applied in order with git apply to a scratch clone at 09bec571; the scratch tree with all six applied passes the checks design.md ## Context lists
diff --git a/hooks/enforce-agent-baseline.py b/hooks/enforce-agent-baseline.py
index 37237a81..dd5e9ae9 100755
--- a/hooks/enforce-agent-baseline.py
+++ b/hooks/enforce-agent-baseline.py
@@ -17,19 +17,37 @@ merely cites the role's contract file -- an implementer, a self-review planner,
 escalation carrying verifier reports -- is not that role. Prose in
 skills/flow/visual-verify.md did not stop a hand-written verifier dispatch on flow-high.
 
+It denies every fork in a flow session -- one whose transcript records a /flow, /flow-fast,
+/flow-plan or /flow-self-review command, or whose prompt names a flow skill: a flow dispatch is a
+fresh agent whose prompt carries the paths it needs (KAN-1000: a parent forked itself into a
+self-review loop and dragged ~300k tokens of its own context into it).
+
 Fails open: any unexpected input, parse error, or internal fault exits 0 and lets the
 dispatch through. A guardrail on a convention must never be the reason work stops.
 """
 
 import json
+import re
 import sys
 
 # The tool that spawns a subagent. Current builds name it "Agent"; older ones "Task".
 DISPATCH_TOOLS = {"Agent", "Task"}
 
-# A fork inherits the parent's full context, CLAUDE.md included, so it needs no pointer.
+# A fork inherits the parent's full context, CLAUDE.md included, so it needs no pointer --
+# outside a flow session, where it is denied instead (fork_reason).
 EXEMPT_SUBAGENT_TYPES = {"fork"}
 
+# A flow run's slash command as the harness writes it into the session transcript.
+FLOW_COMMAND = re.compile(rb"<command-name>/flow(?:-fast|-plan|-self-review)?</command-name>")
+
+FORK_REASON = (
+    "Dispatch blocked: no flow phase, and no loop a phase or the self-review runs, dispatches "
+    "subagent_type \"fork\" (skills/flow/implement.md, KAN-1000).\n\n"
+    "Re-send this call as a fresh agent on the pair its contract names, its prompt carrying "
+    "what it needs -- paths to the run's artifacts, reports and records -- never this "
+    "session's context."
+)
+
 # Both must appear in the prompt: the file being pointed at, and the propagation clause
 # that makes the receiving agent a carrier. Either alone is not the rule.
 REQUIRED_MARKERS = ("agent-baseline.md", "verbatim")
@@ -48,7 +66,8 @@ DENY_REASON = (
     "Re-send this call with the following two sentences included verbatim in `prompt`:\n\n"
     f"{REQUIRED_TEXT}\n\n"
     "The second sentence is not optional: it is what carries the rules past depth 1.\n"
-    "Exempt: subagent_type \"fork\", which inherits this session's context already."
+    "Exempt: subagent_type \"fork\" outside a flow session, which inherits this session's "
+    "context already."
 )
 
 # (marker line, subagent_type, model, the contract that fixes the pair). The marker is
@@ -110,6 +129,22 @@ def pinned_role_reason(prompt: str, tool_input: dict) -> "str | None":
     return None
 
 
+def fork_reason(payload: dict, tool_input: dict) -> "str | None":
+    prompt = tool_input.get("prompt")
+    if isinstance(prompt, str) and "skills/flow" in prompt:
+        return FORK_REASON
+    path = payload.get("transcript_path")
+    if not isinstance(path, str):
+        return None
+    try:
+        with open(path, "rb") as f:
+            if any(FLOW_COMMAND.search(line) for line in f):
+                return FORK_REASON
+    except OSError:
+        return None
+    return None
+
+
 def deny(reason: str) -> None:
     json.dump(
         {
@@ -140,6 +175,9 @@ def main() -> int:
         return 0
 
     if tool_input.get("subagent_type") in EXEMPT_SUBAGENT_TYPES:
+        reason = fork_reason(payload, tool_input)
+        if reason:
+            deny(reason)
         return 0
 
     prompt = tool_input.get("prompt")
diff --git a/scripts/test-enforce-agent-baseline-hook.sh b/scripts/test-enforce-agent-baseline-hook.sh
index 32e6764f..0a26ab7a 100755
--- a/scripts/test-enforce-agent-baseline-hook.sh
+++ b/scripts/test-enforce-agent-baseline-hook.sh
@@ -19,7 +19,11 @@
 #  8b. the same from ~/.zcode/hooks on flow-high /
 #      glm-5.3-flash, a type ZCode does not offer (KAN-609) -> deny, names general-purpose
 #   9. the same from ~/.zcode/hooks on flow-low / opus     -> deny, names glm-5.3-flash
-#  10. a fork dispatch carrying the verifier marker        -> allow
+#  10. a fork dispatch carrying the verifier marker        -> deny (names a flow skill)
+#  10b. a fork naming no flow skill, no transcript          -> allow
+#  10c. the same fork in a session whose transcript records
+#       a /flow command                                     -> deny
+#  10d. the same fork, transcript recording only /flow-status -> allow
 #  11. malformed JSON                                      -> allow, exit 0
 #  12. a planner or implementer prompt citing the verifier
 #      or analyst file, with no marker line, off the
@@ -53,6 +57,12 @@ if len(sys.argv)>3: ti["model"]=sys.argv[3]
 print(json.dumps({"tool_name":"Agent","tool_input":ti}))' "$@"
 }
 
+# payload_in <transcript> <prompt> <subagent_type>: a payload naming its session transcript
+payload_in() {
+  python3 -c 'import json,sys
+print(json.dumps({"tool_name":"Agent","transcript_path":sys.argv[1],"tool_input":{"prompt":sys.argv[2],"subagent_type":sys.argv[3]}}))' "$@"
+}
+
 # run_hook <hook path> <stdin>: sets RC and OUT
 run_hook() {
   set +e
@@ -125,7 +135,18 @@ run_hook "$SANDBOX/.zcode/hooks/enforce-agent-baseline.py" "$(payload "$VERIFIER
 expect_deny "9 zcode verifier on the claude pair" 'model: \"glm-5.3-flash\"'
 
 run_hook "$HOOK" "$(payload "$VERIFIER" fork)"
-expect_allow "10 fork exempt"
+expect_deny "10 fork naming a flow skill" 'subagent_type \"fork\"'
+
+run_hook "$HOOK" "$(payload 'summarise the thread above' fork)"
+expect_allow "10b fork outside a flow session"
+
+printf '%s\n' '{"message":{"content":"<command-name>/flow</command-name>"}}' >"$SANDBOX/flow.jsonl"
+run_hook "$HOOK" "$(payload_in "$SANDBOX/flow.jsonl" 'run the loop above' fork)"
+expect_deny "10c fork in a /flow session" 'KAN-1000'
+
+printf '%s\n' '{"message":{"content":"<command-name>/flow-status</command-name>"}}' >"$SANDBOX/status.jsonl"
+run_hook "$HOOK" "$(payload_in "$SANDBOX/status.jsonl" 'run the loop above' fork)"
+expect_allow "10d fork after /flow-status alone"
 
 run_hook "$HOOK" '{not json'
 expect_allow "11 malformed JSON"
diff --git a/skills/flow-self-review/SKILL.md b/skills/flow-self-review/SKILL.md
index 9d88f248..f75aa4c5 100644
--- a/skills/flow-self-review/SKILL.md
+++ b/skills/flow-self-review/SKILL.md
@@ -143,7 +143,7 @@ design that holds, the next step is effort `xhigh` — and an `xhigh` dispatch i
 without the operator's approval, asked first with the finding, the objection and what `high`
 returned; a refusal leaves the finding offered for filing in step 4. **This session plans,
 changes and reviews nothing itself** (`design.md`, `one-shot-subagent-fix-loop`), save item 4's
-ticks and archive commit and **The base sync** below: each numbered step below that does is a fresh dispatch on the pair it names, its prompt carrying the MODEL HANDSHAKE,
+ticks and archive commit and **The base sync** below: each numbered step below that does is a fresh dispatch on the pair it names, never a fork of this session (**Dispatch sites — the parent's closed list**, `skills/flow/implement.md`), its prompt carrying the MODEL HANDSHAKE,
 TOOLS, FOREGROUND BUILDS and NO DELEGATION paragraphs of **4. Execute (SDD + TDD)**
 (`skills/flow/implement.md`) verbatim — a review's READ-ONLY REVIEW too — with **The
 handshake** there applying to each reply. Every dispatch is one-shot: no `SendMessage` to it once
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index 08f8c759..7ec32fdd 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -76,6 +76,13 @@ other name.
 **The self-check.** Before any Agent-tool call, the parent names which row above the call is. A
 call that names no row is not made.
 
+**No row is ever a fork.** No `/flow*` phase, and no loop a phase or the self-review runs,
+dispatches `subagent_type: "fork"`: every flow dispatch is a fresh agent whose prompt carries what
+it needs — paths to the run's artifacts, reports and records — never the parent's context, and
+`hooks/enforce-agent-baseline.py` denies a fork in a flow session (KAN-1000: a parent forked itself
+to run a self-review loop, dragging about 300k tokens of uncached context into a dispatch that
+needed a change list and file paths).
+
 **Every dispatch's Agent-tool `description` names what it carries, in the operator's status-line
 shape.** A dispatch carrying plan tasks — an implementer, a gated reviewer bundle — names them
 by plan id over the plan's task count, then a few words in parentheses: `Task 3/22 (port guard)`
````

  - [x] **Step 3: Verify.** `bash scripts/test-enforce-agent-baseline-hook.sh` prints `PASS`; then from the root `scripts/check-python-suppressions.sh && scripts/check-references.sh && scripts/check-installed-citations.sh && scripts/check-markdown-integrity.py && scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 4: Commit** the four files with the subject below.

**Build:** green
**Files:** `hooks/enforce-agent-baseline.py`, `scripts/test-enforce-agent-baseline-hook.sh`, `skills/flow/implement.md`, `skills/flow-self-review/SKILL.md`
**Tests:** `10b fork outside a flow session`, `10c fork in a /flow session`, `10d fork after /flow-status alone`
**Regression:** `10c fork in a /flow session` — reverting the commit admits a fork in a session whose transcript records `/flow`, and the case fails with the hook's empty output; case `10 fork naming a flow skill` fails the same way; `10b` and `10d` pin that a fork outside a flow run stays admitted.
**Baseline:** before=18 after=21
<!-- measured: grep -cE '^expect_(allow|deny) "' scripts/test-enforce-agent-baseline-hook.sh @ 09bec571 → 18, and → 21 with Task 6 applied -->
**Commit:** fix(hooks): deny a fork dispatch in a flow session
**After:** none

**Decision:** fork-denied-in-flow-session
