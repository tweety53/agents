# kan-1064-flow-self-review-run-the-coordination-around-the

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

Each task applies its diff exactly as written: every sentence a diff adds or rewords is listed in
`spectre/changes/kan-1064-flow-self-review-run-the-coordination-around-the/verbatim-moves.txt`, so a reworded sentence fails `check-verbatim-moves.sh`.
`<worktree>` is `/Users/tweety53/Projects/agents-worktrees/kan-1064-flow-self-review-run-the-coordination-around-the`.

- [ ] 1. Self-review fixer pair: `opus`/`sonnet` at `flow-medium`, no `flow-high` escalation row
  - [ ] **Step 1: Write the test first — apply the two test-file hunks of the diff below to `stats/internal/guard/check_loop_dispatch_test.go` (the new `TestCheckLoopDispatchSelfReviewFixPairs`, and `TestCheckLoopDispatchOwnBranchInAnotherProject`'s fix dispatch moved to `flow-medium`) and the `stats/internal/guard/pipeline_fix_test.go` hunk (`TestPipelineFixOwnBranchInAnotherProject`'s self-review `escalate` now refused with `admits no escalation`). Run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckLoopDispatchSelfReviewFixPairs|TestCheckLoopDispatchOwnBranchInAnotherProject|TestPipelineFixOwnBranchInAnotherProject' -count=1` and see all three fail against the old table.**
  - [ ] **Step 2: Apply the `skills/flow-contracts/model-policy.md` hunk: the `self-review-fix` row becomes `opus/flow-medium`, `sonnet/flow-medium` and its `opus/flow-high` `No-progress escalation` row is deleted. Re-run the Step 1 command: green.**
  - [ ] **Step 3: Verify: `cd stats && make -s web-dist && go test ./internal/modelpolicy/... -count=1 && go test ./internal/guard/ -run 'TestCheckLoopDispatch|TestPipelineFix|TestLoop' -count=1`, then `scripts/run-verify-list.sh <worktree> lint` (the task's `**Files:**` hold a Markdown file under `skills/`).**

```diff verified:applied to a scratch copy of this branch's HEAD; check-verbatim-moves.sh against HEAD then 0 violations with this change's verbatim-moves.txt, and the guard tests below ran green
diff --git a/skills/flow-contracts/model-policy.md b/skills/flow-contracts/model-policy.md
--- a/skills/flow-contracts/model-policy.md
+++ b/skills/flow-contracts/model-policy.md
@@ -139,4 +139,3 @@ roles table by them.
 | `self-review-plan` | `opus/flow-planner-medium`, `opus/flow-planner-low` | — | Plans the fixes for the self-review's findings. |
-| `self-review-fix` | `opus/flow-low` | — | Makes one planned self-review fix. |
-| `self-review-fix` | `opus/flow-high` | `No-progress escalation` | — |
+| `self-review-fix` | `opus/flow-medium`, `sonnet/flow-medium` | — | Makes one planned self-review fix. |
 | `self-review-review-1` | `opus/flow-medium`, `opus/flow-low` | — | First review of a self-review fix. |
diff --git a/stats/internal/guard/check_loop_dispatch_test.go b/stats/internal/guard/check_loop_dispatch_test.go
--- a/stats/internal/guard/check_loop_dispatch_test.go
+++ b/stats/internal/guard/check_loop_dispatch_test.go
@@ -282,3 +282,3 @@ func TestCheckLoopDispatchOwnBranchInAnotherProject(t *testing.T) {
 	gitRun(t, fx.repo, "branch", "self-review-kan-1-demo")
-	if rc, out := fx.run(t, "self-review-kan-1-demo-fix", "flow-low", "opus"); rc != 0 ||
+	if rc, out := fx.run(t, "self-review-kan-1-demo-fix", "flow-medium", "opus"); rc != 0 ||
 		!strings.Contains(fx.calls(t), "record dispatch begin -C "+fx.repo+" -change self-review-kan-1-demo") {
@@ -349 +349,33 @@ func TestCheckLoopDispatchFirstReviewEffort(t *testing.T) {
 }
+
+// TestCheckLoopDispatchSelfReviewFixPairs pins kan-1064's self-review-fix
+// row: the fixer runs on opus or sonnet at flow-medium, flow-low is off its
+// row, and with no marked row an escalated flow-high fix is refused as a
+// pair -- the no-progress answer is a fresh re-plan, never flow-high.
+func TestCheckLoopDispatchSelfReviewFixPairs(t *testing.T) {
+	t.Parallel()
+	const token = "self-review-kan-1-demo"
+	for _, tc := range []struct {
+		name, key, typ, model, prompt string
+		escalated                     bool
+		rc                            int
+		out                           string
+	}{
+		{"opus medium", token + "-fix-1", "flow-medium", "opus", "", false, 0, "LOOP-DISPATCH-OK"},
+		{"sonnet medium", token + "-fix-1", "flow-medium", "sonnet", "", false, 0, "LOOP-DISPATCH-OK"},
+		{"opus low", token + "-fix-1", "flow-low", "opus", "", false, 1, "PAIR —"},
+		{"escalated opus high", token + "-fix-1", "flow-high", "opus", "No-progress escalation\n", true, 1, "PAIR —"},
+	} {
+		t.Run(tc.name, func(t *testing.T) {
+			t.Parallel()
+			fx := lpNew(t, lpRowsJSON(token+"-", token, "plan completed", "fix completed"), token)
+			if tc.escalated {
+				gitRun(t, fx.repo, "config", "--add", "flow.escalation", token+":"+tc.key)
+			}
+			fx.prompt = tc.prompt
+			if rc, out := fx.run(t, tc.key, tc.typ, tc.model); rc != tc.rc || !strings.Contains(out, tc.out) {
+				t.Fatalf("rc %d, want %d; output %q, want it to carry %q", rc, tc.rc, out, tc.out)
+			}
+		})
+	}
+}
diff --git a/stats/internal/guard/pipeline_fix_test.go b/stats/internal/guard/pipeline_fix_test.go
--- a/stats/internal/guard/pipeline_fix_test.go
+++ b/stats/internal/guard/pipeline_fix_test.go
@@ -403,11 +403,4 @@ func TestPipelineFixOwnBranchInAnotherProject(t *testing.T) {
 	}
-	if rc, out := fx.pf(t, "escalate", token+"-fix-2"); rc != 0 {
-		t.Fatalf("escalate: rc %d: %s", rc, out)
-	}
-	if _, out := kwRun(fx.env, "", fixtureGit, "-C", proj, "config", "--get-all", "flow.escalation"); !strings.Contains(out, token+":"+token+"-fix-2") {
-		t.Fatalf("the escalation was not recorded in the caller's project: %q", out)
-	}
-	lp, _, _, msg, ok := lpFromKey(fx.env, fixtureGit, proj, token+"-fix-2")
-	if !ok || !lpEscalated(fx.env, fixtureGit, lp.main, lp, token+"-fix-2") {
-		t.Fatalf("check-loop-dispatch does not read the escalation in the caller's project: %s", msg)
+	if rc, out := fx.pf(t, "escalate", token+"-fix-2"); rc != 1 || !strings.Contains(out, "admits no escalation") {
+		t.Fatalf("escalate of a self-review fix: rc %d, want 1 naming no escalation: %s", rc, out)
 	}
```

**Build:** green
**Files:** `skills/flow-contracts/model-policy.md`, `stats/internal/guard/check_loop_dispatch_test.go`, `stats/internal/guard/pipeline_fix_test.go`
**Tests:** `TestCheckLoopDispatchSelfReviewFixPairs`
**Regression:** reverting the commit restores the `opus/flow-low` row and its `flow-high` escalation row: `TestCheckLoopDispatchSelfReviewFixPairs` fails (both `flow-medium` fixes refused `PAIR —`, the `flow-low` fix and the escalated `flow-high` fix admitted), as do the edited `TestCheckLoopDispatchOwnBranchInAnotherProject` and `TestPipelineFixOwnBranchInAnotherProject`.
**Baseline:** before=443 after=444
<!-- measured: cat stats/internal/guard/*_test.go | grep -c '^func Test' @ branch spectre/kan-1064-flow-self-review-run-the-coordination-around-the (443 at the plan's base; 444 in the scratch copy with this diff applied) -->
**Commit:** feat(flow-contracts): run the self-review fixer at flow-medium with no flow-high row
**After:** none

**Decision:** self-review-fix-pair

- [ ] 2. `/flow-self-review`: main passes file paths, never file contents
  - [ ] **Step 1: Apply the diff below to `skills/flow-self-review/SKILL.md`: `<sr-dir>` and the paths-not-contents rule open step 2; the pass names the bundle path, writes `findings.md`, `report-draft.md`, `cases/<case-id>/` and angle 7's extractions, and hands back a short summary; step 2a copies the case folders; the plan, review, fix-r and fix-v prompts name files; the `fix` row's pair and its no-escalation sentence; steps 4, 6 and 7 fill and copy the draft.**
  - [ ] **Step 2: Verify: `scripts/check-verbatim-moves.sh` (no FAIL line), then `scripts/run-verify-list.sh <worktree> lint`.**

```diff verified:applied to a scratch copy of this branch's HEAD; check-verbatim-moves.sh against HEAD then 0 violations with this change's verbatim-moves.txt, and the guard tests below ran green
diff --git a/skills/flow-self-review/SKILL.md b/skills/flow-self-review/SKILL.md
--- a/skills/flow-self-review/SKILL.md
+++ b/skills/flow-self-review/SKILL.md
@@ -36,3 +36,3 @@ checkout on any other branch: print the branch and stop.
 **Both refusals bind the standalone command only.** A pass run inside `/flow`'s integrate or
-`/flow-fast`'s verify holds its bundle in memory, runs on the change branch, and starts at step 2.
+`/flow-fast`'s verify has its bundle at `<sr-dir>/bundle.md`, runs on the change branch, and starts at step 2.
 
@@ -40,8 +40,17 @@ checkout on any other branch: print the branch and stop.
 
+**This session passes file paths, never file contents.** Every handoff between this file's steps
+is a file in `<sr-dir>`, `<state-dir>/self-review/<name>/`, `<state-dir>` being the path `flow
+state dir` prints — outside every checkout, never committed. This session reads each dispatch's
+hand-back, never the files it names, save where a step below says it reads one; every judgement —
+`big`, filing, landing, and when a step runs again — stays this session's.
+
 Dispatch the pass, in-run and standalone alike: key `self-review-<name>-pass`, on the `pass` row
-of **Allowed pairs** (`skills/flow-contracts/model-policy.md`), its prompt carrying the bundle in full (`<bundle>`
-standalone, the bundle held in memory in-run), this step's rules and step 3's classification, the
-dispatch paragraphs step 3 names and READ-ONLY REVIEW. It returns every finding with its angle, its
-evidence and its classification — `noted` or fixable, in or out of scope, and for a fixable one its
-counted blast radius and whether it is `big`. This session keeps the filing-and-rating prompt, the
+of **Allowed pairs** (`skills/flow-contracts/model-policy.md`), its prompt naming the bundle's path — `<bundle>`
+standalone, `<sr-dir>/bundle.md` in-run — and `<sr-dir>`, and carrying this step's rules and step 3's classification, the
+dispatch paragraphs step 3 names and READ-ONLY REVIEW. It writes `<sr-dir>/findings.md`: every finding under an `F<n>` id, with its angle, its
+evidence, its severity and its classification — `noted` or fixable, in or out of scope, and for a fixable one its
+counted blast radius and whether it is `big` — and a `## Offered` section laying out every `big` finding and every angle-7 candidate in step 4's explanation layout.
+It writes `<sr-dir>/report-draft.md`, step 7's report, each disposition step 3 decides written as a placeholder: `fixed: <F-id>` for a non-`big` finding, `offered: <F-id>` for a `big` one.
+Every file it writes under `<sr-dir>` or `<state-dir>/flow-eval/<name>/` is its own report file under READ-ONLY REVIEW.
+It hands back only its per-angle counts, the non-`big` finding ids, one line per `big` finding, its case ids, and whether it met a product defect. This session keeps the filing-and-rating prompt, the
 store rows and the report. The pass's dispatch row is opened under `<name>` by the hook that admits it;
@@ -90,3 +99,3 @@ every field present:
 Every path is absolute. A file that exists only in git history, or only inside a worktree that
-cleanup removes, is extracted by this session before step 4 — `git -C <repo> show <sha>:<path>`
+cleanup removes, is extracted by the pass — `git -C <repo> show <sha>:<path>`
 into `<state-dir>/flow-eval/<name>/`, `<state-dir>` being the path `flow state dir` prints for
@@ -168,3 +177,3 @@ folder:
 
-The pass returns each case's `case.json` fields (step 2a); it writes nothing.
+The pass writes each case's files, step 2a's table save `index.html`, into `<sr-dir>/cases/<case-id>/`.
 
@@ -202,6 +211,6 @@ step 3, first resolves the worktree of `<fix-branch>` — the branch named at th
 creating it with `<agents repo>/scripts/pipeline-fix.sh start <name>` when it is
-`self-review-<name>` and absent; it never writes the agents main checkout. It writes each case into that
+`self-review-<name>` and absent; it never writes the agents main checkout. It copies each `<sr-dir>/cases/<case-id>/` folder the pass wrote into that
 worktree — `<agents repo>-worktrees/<name>/docs/prompt-tuning/<case-id>/` on the change's own branch,
 `<agents repo>-worktrees/self-review-<name>/docs/prompt-tuning/<case-id>/` otherwise, `<case-id>` being
-`<name>-<n>`, `<n>` counting from 1 within the change:
+`<name>-<n>`, `<n>` counting from 1 within the change, each holding the files the pass wrote:
 
@@ -220,3 +229,3 @@ extraction does. A `reviewer` case whose `target` is `skills/flow/visual-verify-
 `mockup.png`, the mockup frame it judged that capture against, both copied into the case folder.
-Once a case's files are written, run `flow prompt-tuning render <case-dir>`, then
+Once a case's folder is copied, run `flow prompt-tuning render <case-dir>`, then
 `<agents repo>/scripts/check-prompt-tuning-cases.sh` over the new folders. A red line is fixed by
@@ -270,4 +279,3 @@ PLANNING DELEGATION below in place of NO DELEGATION — with **The
 handshake** there applying to each reply. Every dispatch is one-shot: no `SendMessage` to it once
-it returns. The no-progress escalation of **Fewest operator actions**
-(`skills/flow-contracts/pipeline.md`) still runs on `flow-high`.
+it returns. The `fix` row carries no no-progress escalation: a fix that makes no progress is answered by the fresh re-plan above, never by `flow-high`.
 
@@ -281,3 +289,3 @@ every finding the plan carries is mechanical, this session recording the one-lin
 dispatch; the `flow-planner-<effort>` definitions are the `flow-<effort>` ones plus `Agent`
-(decision `planner-carries-agent`, kan-1033) — **the fixer pair** the `fix` row and **the review pair**
+(decision `planner-carries-agent`, kan-1033) — **the fixer pair** the `fix` row, `opus` by default and `sonnet` only for a fix that is simple and predictable, this session recording the one-line reason with the dispatch — and **the review pair**
 the two `review` rows: the first review on `flow-medium` or `flow-low` as the planner recorded —
@@ -314,3 +322,3 @@ any other project's run, `self-review-<name>` based at `origin/<agents-base>`.
    own branch its worktree is already set up. It then dispatches key `self-review-<name>-plan` on the planner pair, its
-   prompt carrying every non-`big` finding with its evidence and counted blast radius. The planner
+   prompt naming `<sr-dir>/findings.md` and the non-`big` finding ids, whose evidence and counted blast radius that file carries. The planner
    plans them all in full as one spectre change, `self-review-<name>`, in that worktree, per
@@ -364,4 +372,5 @@ any other project's run, `self-review-<name>` based at `origin/<agents-base>`.
    and one `Important: <n>` line counting them.
-3. **Fix the review's findings** — key `self-review-<name>-fix-<r>` on the fixer pair, its prompt carrying the
-   review's report verbatim. Each fix is folded into the commit of the task it fixes
+   Its prompt names `<sr-dir>/review-<r>.md` as its report file: it writes its report there and hands back only those two lines, and this session closes it with `pipeline-fix.sh end <key> <sr-dir>/review-<r>.md`.
+3. **Fix the review's findings** — key `self-review-<name>-fix-<r>` on the fixer pair, its prompt naming
+   `<sr-dir>/review-<r>.md`. Each fix is folded into the commit of the task it fixes
    (`git commit --fixup <that commit>`, then `git rebase --autosquash <fix-base>`), so
@@ -383,6 +392,5 @@ any other project's run, `self-review-<name>` based at `origin/<agents-base>`.
    alone carries the ticks, and is made once. This session then runs
-   `<agents repo>/scripts/pipeline-fix.sh verify <fix worktree>`. A red line is fixed by a fresh
+   `<agents repo>/scripts/pipeline-fix.sh verify <fix worktree> > <sr-dir>/verify-<v>.log 2>&1`, `<v>` counting verify runs from 1, and reads only its exit code and its `red:` lines. A red line is fixed by a fresh
    dispatch of step 3, key
-   `self-review-<name>-fix-v<v>`, its prompt carrying the failing command and its output
-   verbatim; then step 2, then this step's lint lines and tests again, `<v>` plus one. The
+   `self-review-<name>-fix-v<v>`, its prompt naming the failing command and `<sr-dir>/verify-<v>.log`; then step 2, then this step's lint lines and tests again, `<v>` plus one. The
    autosquash keeps the archive commit last; a task commit dropped after it has its task unticked
@@ -435,2 +443,4 @@ explanation, so the prompt records the decision only: a filed issue is durable,
 explanation arriving afterward describes something the operator did not agree to.
+This session first fills each `fixed: <F-id>` placeholder of `<sr-dir>/report-draft.md` with the sha `git log --format='%h %s'` shows for its task's commit subject, rewriting one whose finding step 3 left unfixed as `offered: <F-id>`.
+It then prints that body from `<sr-dir>/findings.md`'s `## Offered` section and the draft's `fixed:` lines, never composing either.
 
@@ -526,2 +536,5 @@ checkout there: print `⚠ flow-eval: skipped — <reason>` and mine nothing. A
 
+First this session fills each `offered: <F-id>` placeholder of `<sr-dir>/report-draft.md` with `filed: <KEY>` from step 5 or `declined`.
+It reads the filled draft once, here, and records the rows from its lines.
+
 One call per finding, every disposition alike, before the report is written — save a `fixed`
@@ -542,2 +555,3 @@ A row is stated as recorded only once `flow self-review findings -change <name>`
 
+The report is the filled `<sr-dir>/report-draft.md`, copied to its path below; this session never composes the eight angles.
 Write `<project>/docs/self-review/<name>-self-review.md` in the shape
```

**Build:** green
**Files:** `skills/flow-self-review/SKILL.md`
**Tests:** **none**
**Regression:** none — the task declares no tests; `check-verbatim-moves.sh` fails on any sentence the implementer words differently from `verbatim-moves.txt`.
**Baseline:** before=0 after=0
**Commit:** feat(flow-self-review): pass file paths, never file contents, through the self-review loop
**After:** Task 1

**Decision:** paths-not-contents

**Decision:** handoff-dir-state-dir

**Decision:** no-review-merger

- [ ] 3. The in-run bundle goes to `<sr-dir>/bundle.md`, never held in memory
  - [ ] **Step 1: Apply the diff below to `skills/flow-contracts/finish-contract-run1.md` (**Run the self-review pass**), `skills/flow/integrate.md` (step 4) and `skills/flow-fast/SKILL.md` (the self-review section): each writes the bundle to `<sr-dir>/bundle.md`, never printed, instead of holding it in memory.**
  - [ ] **Step 2: Absence sweep — `git grep -n -F -e 'holds the whole in memory' -e 'in memory; run the pass' -e 'Assemble the bundle in memory' -e 'holds its bundle in memory' -e 'held in memory in-run' -- skills/flow-contracts/finish-contract-run1.md skills/flow/integrate.md skills/flow-fast/SKILL.md skills/flow-self-review/SKILL.md` prints nothing.**
  - [ ] **Step 3: Verify: `scripts/check-verbatim-moves.sh` (no FAIL line), then `scripts/run-verify-list.sh <worktree> lint`.**

```diff verified:applied to a scratch copy of this branch's HEAD; check-verbatim-moves.sh against HEAD then 0 violations with this change's verbatim-moves.txt, and the guard tests below ran green
diff --git a/skills/flow-contracts/finish-contract-run1.md b/skills/flow-contracts/finish-contract-run1.md
--- a/skills/flow-contracts/finish-contract-run1.md
+++ b/skills/flow-contracts/finish-contract-run1.md
@@ -355,5 +355,5 @@ and its `git log --stat` lines leave out every `*.png` path. A source that is ab
 `skipped: <source> (absent)` inside the bundle, never fatal.
-The session appends `## Session narrative` (one paragraph it writes for run 1 itself; the
+The session writes that document to `<sr-dir>/bundle.md` (**2. Read the bundle and run the eight angles**, `skills/flow-self-review/SKILL.md`), never printing it, and appends `## Session narrative` there (one paragraph it writes for run 1 itself; the
 archived `narrative.md` is already a bundle section, and a change predating the narrative rule
-has the bundle report it skipped) and holds the whole in memory — it is never written to
+has the bundle report it skipped) — it is never written to
 `<project>/docs/self-review/` or committed.
diff --git a/skills/flow-fast/SKILL.md b/skills/flow-fast/SKILL.md
--- a/skills/flow-fast/SKILL.md
+++ b/skills/flow-fast/SKILL.md
@@ -310,3 +310,3 @@ A report already committed on `<name>` (`<project>/docs/self-review/<name>-self-
 a fix re-run never records its findings twice.
-Assemble the bundle in memory, never written to `<project>/docs/self-review/` or committed, as, in
+Assemble the bundle in `<sr-dir>/bundle.md` (**2. Read the bundle and run the eight angles**, `skills/flow-self-review/SKILL.md`), never printed, never written to `<project>/docs/self-review/` or committed, as, in
 order: the stdout of `flow self-review
diff --git a/skills/flow/integrate.md b/skills/flow/integrate.md
--- a/skills/flow/integrate.md
+++ b/skills/flow/integrate.md
@@ -160,3 +160,3 @@ A `flow-fix` row `flow self-review findings -change <name>` lists as `fixed` who
 Per **Run the self-review pass** (`skills/flow-contracts/finish-contract-run1.md`), canonical for it.
-Hold the bundle's stdout, then `## Session narrative`, in memory; run the pass over it; then
+Write the bundle's stdout, then `## Session narrative`, to `<sr-dir>/bundle.md`, never printing it; run the pass over it; then
 commit the report in `<canonical-worktree>`, not pushing here:
```

**Build:** green
**Files:** `skills/flow-contracts/finish-contract-run1.md`, `skills/flow/integrate.md`, `skills/flow-fast/SKILL.md`
**Tests:** **none**
**Regression:** none — the task declares no tests; the Step 2 sweep prints a hit if any in-memory wording is left.
**Baseline:** before=0 after=0
**Commit:** feat(flow): write the in-run self-review bundle to the state dir
**After:** Task 2

**Decision:** paths-not-contents
