# self-review-kan-1034-spectre-optional-bundled-planning-prompts-that

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

**Goal:** close the ten non-`big` findings of kan-1034's self-review pass
(`.superpowers/sdd/self-review-pass.md`, angles 1, 2 and 4), one task each, Task 6 carrying two.

**Architecture:** four Go guards in `stats/internal/guard/` extended in place, each with a test that
fails without its fix (Tasks 1–5); four prose and config edits (Tasks 6–9), Task 9 in the spectre
repository.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-26-self-review-kan-1034-spectre-optional-bundled-planning-prompts-that-design.md`
in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Tasks 1–8 work in the agents worktree `/Users/tweety53/Projects/agents-worktrees/kan-1034-spectre-optional-bundled-planning-prompts-that`, Task 9 in the spectre worktree `/Users/tweety53/Projects/spectre-worktrees/kan-1034-spectre-optional-bundled-planning-prompts-that`, both on branch `spectre/kan-1034-spectre-optional-bundled-planning-prompts-that`; every `**Files:**` path is relative to its task's worktree root. Never touch a main checkout, never create a worktree, and never touch `spectre/changes/archive/kan-1034-spectre-optional-bundled-planning-prompts-that/`.
- A prose diff a step gives is applied with `git apply` from the worktree root after saving the lines between its fence markers to a scratch file outside the worktree; one that does not apply is reported, never hand-merged. Each added sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Run `gofmt -w` on every Go file you change before committing; `cd stats && make -s web-dist && go vet ./... && gofmt -l .` stays clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.
- Existing test cases a new rule makes fail are brought in line by giving their fixtures what the rule now requires — never by narrowing the rule.

## Review Focus

- Task 1: a re-run report missing is named once, not on both the re-run line and the new presence line; the state record is read only when a reproducer token is absent from the worktree; the `none` reproducer forms and an empty reproducer are never checked.
- Task 4: a symlink on either side is skipped with its stderr note and never followed; every other scoped path still compares and the run exits 0.
- Task 5: blank lines inside a fence never fire signal 6; an empty file never fires signal 5.

- [ ] 1. Panel close: every completed reviewer dispatch leaves a report, every fixed finding's reproducer exists

**Build:** green
**Files:** `stats/internal/guard/panelfindingsclosed.go`, `stats/internal/guard/check_panel_findings_closed_test.go`, `scripts/check-panel-findings-closed.sh`, `skills/flow/review-panel.md`
**Tests:** `TestCheckPanelFindingsClosedEveryRoleLeavesAReport`, `TestCheckPanelFindingsClosedFixedReproducerExists`
**Regression:** `TestCheckPanelFindingsClosedEveryRoleLeavesAReport` passes a completed round-0 slot with no report file, so the F7 slot's lost report is silent again; `TestCheckPanelFindingsClosedFixedReproducerExists` passes a fixed finding whose reproducer path is absent from every worktree.
**Commit:** `fix(guard): refuse a panel close missing a slot report or a fixed finding's reproducer`
**After:** none

**Decision:** report-presence-every-dispatch
**Decision:** fixed-reproducer-resolves

  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `scripts/check-panel-findings-closed.sh /Users/tweety53/Projects/agents-worktrees/kan-1034-spectre-optional-bundled-planning-prompts-that kan-1034-spectre-optional-bundled-planning-prompts-that`:
    it prints `FINDINGS-CLOSED` although F7's recorded reproducer `.superpowers/sdd/reproducers/r1-F1-1.sh` is absent.
    <!-- measured: the same command @ bef019f7 → FINDINGS-CLOSED, exit 0 -->
  - [ ] **Step 2: Write the two failing tests** in `check_panel_findings_closed_test.go`, on the `cfcBinDispatches` stub. Add a `Reproducer` field to `cfcFinding`, emitted as `"reproducer"` when non-empty, and a `*"state get"*)` branch to the stub answering a state record JSON the test supplies (default `{"worktrees":{}}`).
    `TestCheckPanelFindingsClosedEveryRoleLeavesAReport` table: a completed `panel-0-mutation+failure-modes` reviewer dispatch with no report files exits 1 naming both `panel-report-0-mutation.md` and `panel-report-0-failure-modes.md`; with one written it names only the other; with both written as the `no verbatim report captured — <reason>` placeholder it exits 0; an empty file counts as absent; a dispatch ended `aborted`, or of role `implementer`, or keyed `panel-fix-0`, needs no report; a missing re-run report is named once on stderr, not on both lines.
    `TestCheckPanelFindingsClosedFixedReproducerExists` table: a fixed finding whose reproducer `.superpowers/sdd/reproducers/r0-x-1.sh --flag` exists in the worktree exits 0; absent from the worktree and the state record's worktrees exits 1 naming the ref and the token; absent from the worktree but present in a state-record worktree exits 0; reproducers `none — cosmetic` and empty are never checked; an open finding's absent reproducer is not named on this line (the open line already fails it); a state read that fails while a token is absent exits 2.
    Run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckPanelFindingsClosed(EveryRoleLeavesAReport|FixedReproducerExists)' -count=1`: both fail.
  - [ ] **Step 3: Implement in `panelfindingsclosed.go`.** Parse `reproducer` in `cfcParse` (string, else empty). After the re-run coverage block: for each dispatch with role `reviewer`, outcome `completed` or `clean` and `round >= 0`, each `+`-component of its slot whose `panel-report-<round>-<component>.md` is absent or empty (`os.Stat` size 0) and not already in `unreported` is collected; a non-empty list prints `check-panel-findings-closed: completed reviewer dispatch(es) left no report file: <files>` and sets `violated`. Then for each `fixed` finding whose reproducer is non-empty and not a `none` form (the predicate `panelexitcontract.go` uses), derive the path token as that guard does; when `pcExists(worktree+"/"+token)` is false, read the state record once (`pcFlow(env, false, "state", "get", "-C", worktree, name)`, `pcWorktreeKeys`, `resolvedRecordedTrees`; a failed or unparseable read is exit 2 with the cause) and look for the token in those trees; none carrying it collects `<ref> (<token>)`, printed as `check-panel-findings-closed: fixed finding(s) whose runnable reproducer resolves in no worktree of the change: <list>` and setting `violated`. Extend the comment above each block with its reason (kan-1034: F7's slot left no report and its reproducer was lost with its throwaway copy).
  - [ ] **Step 4: Update the contract.** In `scripts/check-panel-findings-closed.sh`'s `Exit codes:` header, add to exit 1: a completed reviewer dispatch left a role's `panel-report-<round>-<id>.md` absent or empty (the placeholder accepted), or a `fixed` finding's runnable reproducer resolves in no worktree of the change; add to exit 2: the state record unreadable when a reproducer token is absent from the worktree. Apply the `review-panel.md` diff:

````diff verified:git apply of this diff to the worktree at bef019f7; check-verbatim-moves.sh then FAILs exactly on the sentences verbatim-moves.txt lists, check-markdown-integrity.sh and check-references.sh exit 0
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index de802022..77c71cd2 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -795,6 +795,10 @@ yourself carrying the single line `no verbatim report captured — <reason>`.
 Every dispatched slot ends up with one, a slot that raised nothing included. **Never re-emit a
 slot's report from this context** — record its `F<n>` rows and cite the file, per **Read
 discipline**'s never-`cat`-a-report rule (`skills/flow/implement.md`).
+Confirm likewise that each runnable reproducer path a slot supplies exists in the worktree before
+recording its finding; when it does not, write the script there yourself from the demonstration
+the slot's final message gives (kan-1034: F7's mutation reproducer was lost with its throwaway
+copy).
 
 **Every finding is a row in the store. The panel record is rendered from those rows.** The parent
 records every finding itself. Every
@@ -1083,7 +1087,9 @@ in the store — nothing is deferred (**Panel re-runs** above); the line names t
 fires on the class **Panel re-runs**' ordering above added: a Critical or Important recorded
 `fixed` whose slot has no clean re-run dispatch of it in any later round — a fixed Minor needs
 none. It fires as well when a slot's report counts more Critical and Important findings on its
-`Critical: <n>` and `Important: <n>` lines than the store holds for that round and slot. Either way, return to the
+`Critical: <n>` and `Important: <n>` lines than the store holds for that round and slot. It fires too
+when a completed reviewer dispatch left one of its roles' `panel-report-<round>-<id>.md` absent or
+empty, or when a `fixed` finding's runnable reproducer resolves in no worktree of the change. Either way, return to the
 handback loop above for them. Exit 2 stops the run.
 
 Beside it, run
````

  - [ ] **Step 5: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckPanelFindingsClosed|TestCheckUnfinishedWork|TestCheckLateFixTrigger|TestCheckPanelFixSingleDispatch' -count=1` passes; `gofmt -l .` and `go vet ./...` clean; `scripts/check-markdown-integrity.sh`, `scripts/check-references.sh` and `scripts/check-verbatim-moves.sh` exit 0. Re-run Step 1's command: it now exits 1 naming `F7 (.superpowers/sdd/reproducers/r1-F1-1.sh)` — expected until the parent recreates that reproducer (decision `kan-1034-repair`); report the line, do not create the file.

- [ ] 2. Plan shape: the bold `**none**` opens the test-repair form

**Build:** green
**Files:** `stats/internal/guard/planparse.go`, `stats/internal/guard/check_plan_shape_test.go`
**Tests:** `Case 44`
**Regression:** `Case 44` fails: `**Tests:** **none** — repairs existing tests:` over a test-shaped Files entry is refused by F4 again.
**Commit:** `fix(guard): read the bold none as the test-repair form`
**After:** none

**Decision:** bold-none-repair

  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Save the fixture below as `<scratch>/tasks.md` and run `scripts/check-plan-shape.sh <scratch>/tasks.md`: exit 1, the F4 contradiction line.
    <!-- measured: scripts/check-plan-shape.sh on this fixture @ bef019f7 → exit 1, F4 -->

````markdown verified:check-plan-shape.sh exit 1 on it at bef019f7
- [ ] 1. Repair

**Files:** `scripts/test-check-foo.sh`
**Tests:** **none** — repairs existing tests: `TestX`
**After:** none
**Commit:** fix: x
````

  - [ ] **Step 2: Add Case 44** in `TestCheckPlanShape`, right after case 34's lines, and watch it fail:

````go verified:applied at bef019f7; go test -run 'TestCheckPlanShape$' fails on it with the regex unchanged and passes with Step 3
	// Case 44 (KAN-1034): the bold `**none**` repair form, which
	// brainstorm-planner.md says opens the field, is the repair form too.
	r = repair(" **none** — repairs existing tests: `TestLoadSets`")
	want("case 44: the bold repair form over a test-shaped Files entry passes", r.rc == 0, r)
	want("case 44: no output", r.out == "", r)
````

  - [ ] **Step 3: Fix `tcfRepairsOpenRE`** in `planparse.go` — `[*_]*` after `none\b` — so the declaration reads as below, and its comment's first line reads "a Tests value opening with `none` (bold `**none**` included), an optional em dash,":

````go verified:applied at bef019f7; go test -run 'TestCheckPlanShape$' passes with Case 44
var tcfRepairsOpenRE = regexp.MustCompile(`(?i)^[` + tcfWS + `*_]*none\b[*_]*[` + tcfWS + `]*(?:[—–]|--)?[` + tcfWS + `]*repairs[` + tcfWS + `]+existing[` + tcfWS + `]+tests[` + tcfWS + `]*:`)
````

  - [ ] **Step 4: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckPlanShape$|TestCheckTaskCommitFields' -count=1` passes; Step 1's command exits 0; `gofmt -l .` and `go vet ./...` clean.

- [ ] 3. Plan shape: refuse a Baseline `measured:` comment the parser cannot read (F20)

**Build:** green
**Files:** `stats/internal/guard/planshape.go`, `stats/internal/guard/check_plan_shape_test.go`
**Tests:** `Case 45`
**Regression:** `Case 45` fails: a Baseline comment `<!-- measured: grep -c x f → 1 at abc -->` passes the shape guard and its re-measurement is silently disabled again.
**Commit:** `fix(guard): refuse a Baseline measured comment the parser cannot read`
**After:** Task 2

**Decision:** measured-comment-f20

  - [ ] **Step 1: Reproduce before touching anything; do not guess.** A one-task plan whose Baseline reads `before=1 after=2 <!-- measured: grep -c x a_test.go → 1 at e093bad -->` passes `scripts/check-plan-shape.sh`.
    <!-- measured: scripts/check-plan-shape.sh on that one-task plan @ bef019f7 → exit 0 -->
  - [ ] **Step 2: Add Case 45** in `TestCheckPlanShape`, just before `// Case 43 (F18)`, reusing case 41's `chain` and `m`; it fails until Step 3:

````go verified:applied at bef019f7 with Step 3; go test -run 'TestCheckPlanShape$' passes and the mutation line proves the rule bites
	// Case 45 (F20): a Baseline measured: comment with no ` @ <ref>` is
	// refused; the read form beside it in case 41 passes.
	md = chain("**Baseline:** before=1 after=2 <!-- measured: grep -c x f → 1 at abc -->", "**Baseline:** before=2 after=3 "+m)
	r = one(md)
	want("case 45 (F20): an unreadable measured: comment fails", r.rc == 1 && has(r, "task 1's **Baseline:** carries a `<!-- measured: … -->` comment with no ` @ <ref>`"), r)
	mutation("F20", md, md, "with no ` @ <ref>`", "case 45")
````

  - [ ] **Step 3: Add F20 to `planshape.go`.** In the header list, before `F19`:

````go verified:applied at bef019f7; gofmt -l clean
//	F20 A Baseline value (the joined field) carrying a `<!-- measured: … -->`
//	    comment tcfMeasuredRE cannot read -- no ` @ <ref>` before its `-->`:
//	    the parser takes no command from it, so the re-measurement is
//	    silently disabled (kan-1034's Task 2). Reported at the Baseline line.
````

    In the per-task checks, directly before `// F4 and F6 read the continuation-joined Tests value`:

````go verified:applied at bef019f7; go vet ./internal/guard/ clean and TestCheckPlanShape passes
	for _, c := range psMeasuredCommentRE.FindAllString(strings.Join(joined["Baseline"], " "), -1) {
		if !tcfMeasuredRE.MatchString(c) && p.on("F20") {
			v = append(v, p.at(psFieldLine(body, start, "Baseline"), "task "+id+"'s **Baseline:** carries a `<!-- measured: … -->` comment with no ` @ <ref>`, so no command is read from it and the re-measurement never runs — write `<!-- measured: <command> @ <ref> -->`"))
		}
	}
````

    And directly above `// baselineConvention is F12:`:

````go verified:applied at bef019f7; gofmt -l clean
// psMeasuredCommentRE is any `<!-- measured: … -->` comment, read or not;
// F20 holds each one against tcfMeasuredRE.
var psMeasuredCommentRE = regexp.MustCompile(`<!--[` + tcfWS + `]*measured:.*?-->`)
````

  - [ ] **Step 4: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckPlanShape' -count=1` passes; Step 1's plan now exits 1 naming F20's message; `scripts/check-plan-shape.sh` with no argument exits 0 on the branch; `gofmt -l .` and `go vet ./...` clean.

- [ ] 4. Relocation comparison: skip a symlinked scoped path instead of refusing the comparison

**Build:** green
**Files:** `stats/internal/guard/relocationcomparison.go`, `stats/internal/guard/generate_relocation_comparison_test.go`
**Tests:** `TestGenerateRelocationComparisonSkipsSymlinks`
**Regression:** `TestGenerateRelocationComparisonSkipsSymlinks` fails: a plan scoping one symlink beside a moved paragraph exits 2 and writes no table.
**Commit:** `fix(guard): skip a symlinked scoped path in the relocation comparison`
**After:** none

**Decision:** relocation-skip-symlink

  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Read `relocationcomparison.go`'s worktree-symlink branch in `generateRelocationComparison` and `grcGitShow`'s mode-120000 branch: both return exit 2 for the whole run. The existing cases ff, gg and the `pin: a symlink at the merge-base only exits 2 naming it` case in `TestGenerateRelocationComparison` pin that refusal.
  - [ ] **Step 2: Write `TestGenerateRelocationComparisonSkipsSymlinks`** with the file's own `repo`/`gen`/`plan` helpers (lift them out of `TestGenerateRelocationComparison` only if they are closures it cannot reach): a plan scoping `src.md`, `dst.md` (a paragraph moved between them) and `link.md` — a worktree symlink — exits 0 with a `moved` row and stderr `generate-relocation-comparison: skipped symlink <worktree>/link.md`; and the same with `link.md` a symlink only at the merge-base (regular file in the worktree) exits 0 with stderr `generate-relocation-comparison: skipped symlink <sha>:link.md` and no row from `link.md`. It fails until Step 3.
  - [ ] **Step 3: Implement.** In `generateRelocationComparison`, the worktree `os.Lstat` symlink branch prints `<grcLabel>skipped symlink <wtPath>` to stderr and `continue`s. In `grcGitShow`, the mode-120000 branch returns a sentinel error `errGrcSymlink` wrapped with the spec (`fmt.Errorf("%s: %w", spec, errGrcSymlink)`); the caller, on `errors.Is(err, errGrcSymlink)`, prints `<grcLabel>skipped symlink <spec>` and `continue`s before appending either side's passages. Update the header: drop the symlink clauses from exit 2, and in Algorithm step 3 say a symlink on either side is skipped with that stderr note so neither side ever resolves a link. Update cases ff, gg and the merge-base pin to expect exit 0, the skip note and no row for the link.
  - [ ] **Step 4: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestGenerateRelocationComparison' -count=1` passes; `gofmt -l .` and `go vet ./...` clean.

- [ ] 5. Markdown integrity: flag a missing final newline and doubled blank lines

**Build:** green
**Files:** `stats/internal/guard/markdownintegrity.go`, `stats/internal/guard/check_markdown_integrity_test.go`, `scripts/check-markdown-integrity.sh`, `skills/flow-contracts/workspace-isolation.md`, `skills/flow/SKILL-rationale.md`
**Tests:** `TestCheckMarkdownIntegrityLayoutSignals`
**Regression:** `TestCheckMarkdownIntegrityLayoutSignals` fails: a skill file with no final newline, or two blank lines in a row outside a fence, exits 0.
**Commit:** `feat(guard): flag a missing final newline and doubled blank lines in skill Markdown`
**After:** none

**Decision:** markdown-layout-signals

  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `scripts/check-markdown-integrity.sh` exits 0 on the branch while `skills/flow-contracts/workspace-isolation.md` ends without a newline and `skills/flow/SKILL-rationale.md` carries a doubled blank line ending at line 154 and another ending at line 288.
    <!-- measured: tail -c1 per file and an awk scan for a blank line after a blank line outside ``` fences, over git ls-files 'skills/**/*.md' 'rules/*.mdc' @ bef019f7 → those three sites only -->
  - [ ] **Step 2: Write `TestCheckMarkdownIntegrityLayoutSignals`** on `mdiProject`/`mdiRun`: `"Text.\n"` exits 0; `"Text."` exits 1 with `skills/foo/SKILL.md:1: missing final newline:`; `"One.\n\n\nTwo.\n"` exits 1 with `:3: consecutive blank lines:`; a fence holding two blank lines in a row exits 0; an empty file exits 0. It fails until Step 3. In `TestCheckMarkdownIntegrity`, `expect` appends `"\n"` to a body that does not end in one, so the existing cases keep testing their own signal; a case whose `want` text pins a line count is checked unchanged.
  - [ ] **Step 3: Implement in `mdiScan`.** Signal 5 — text non-empty and not ending in `\n`: one finding at the last line, `missing final newline`, `the file does not end in a newline`. Signal 6 — a blank line (`mdiStrip` empty) directly after a blank line, neither inside a `fence` block's `start`..`end` range: one finding at the second line, `consecutive blank lines`, `two blank lines in a row; one separates blocks`. Extend the doc comment's signal list with both, and in `scripts/check-markdown-integrity.sh` replace "the four signals" (twice) with "the six signals" and add the two to its opening sentence.
  - [ ] **Step 4: Fix the three sites.** Append the final newline to `skills/flow-contracts/workspace-isolation.md`; delete one blank line of each doubled pair in `skills/flow/SKILL-rationale.md` (before `## Moved by the 2026-09-22 prompt audit — incident records and recorded reasons` and before `## Moved by the 2026-09-22 prompt audit (pass A) — review-panel.md and implement.md`).
  - [ ] **Step 5: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckMarkdownIntegrity' -count=1` passes; `scripts/check-markdown-integrity.sh` and `scripts/check-verbatim-moves.sh` exit 0; `gofmt -l .` and `go vet ./...` clean.

- [ ] 6. Planner: check a relocation's citations against its destination, never cite a branch sha

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** none — prose rules for the planner; `scripts/check-verbatim-moves.sh` checks the sentences
**Commit:** `docs(flow): check a relocation's citations and never cite a change-branch sha`
**After:** none

**Decision:** rules-in-brainstorm-planner

  - [ ] **Step 1: Apply the diff:**

````diff verified:git apply of this diff to the worktree at bef019f7; check-verbatim-moves.sh then FAILs exactly on the sentences verbatim-moves.txt lists, check-markdown-integrity.sh and check-references.sh exit 0
diff --git a/skills/flow/brainstorm-planner.md b/skills/flow/brainstorm-planner.md
index f9e26cc2..b5bd2ccf 100644
--- a/skills/flow/brainstorm-planner.md
+++ b/skills/flow/brainstorm-planner.md
@@ -201,6 +201,12 @@ and cites its output in the task as a `measured:` comment per **Plan provenance*
 exit code or figures the run produced. A premise the run disproves strikes the task during
 planning, before the panel reads the plan.
 
+**Committed content never cites a commit sha on the change's own branch.** Integrate's run 1
+reshapes that branch (**Run 1 — the branch is not merged**, `skills/flow-contracts/finish-contract-run1.md`),
+so a sha a task's committed file names stops resolving; cite a path on the default branch and the
+ticket instead (kan-1034: task 2's vendored `UPSTREAM.md` named the task 1 commit, unreachable once
+the branch was reshaped).
+
 **A task premise about a harness or tool is spiked live at plan time, never assumed.** When a
 task written here rests on how a harness or tool behaves — that a config key takes effect where
 the task puts it, how a clock or timer behaves, what a dev-stack rebuild picks up — planning runs
@@ -286,6 +292,12 @@ comparison (generated later in the pipeline, by `generate-relocation-comparison.
 run before the review panel's `final-review.diff`) to the union
 of every task's own `**Files:**` field across the plan.
 
+A relocation whose destination is consumed outside this repository — a skill another repository
+vendors — checks, before the plan classes a paragraph as movable, that every file, guard and stage
+the paragraph cites resolves in what that destination ships (kan-1034: flow-only paragraphs classed
+generic moved into the vendored `skills/spectre-planning/SKILL.md`, and a fix round moved about 230
+lines back).
+
 Seed `> **Design shape:** plan-mode` into the same block of every plan: on a `tasks.md`
 carrying it, `check-plan-shape.sh` refuses a sibling `design.md` missing any of the four plan sections
 (**C. Create the change and its artifacts**, `skills/spectre-planning/SKILL.md`) (decision `design-shape-line`, kan-1033).
````

  - [ ] **Step 2: Verify.** `scripts/check-markdown-integrity.sh`, `scripts/check-references.sh` and `scripts/check-verbatim-moves.sh` exit 0.

- [ ] 7. Brainstorm: record a planning dispatch's begin before its Agent call

**Build:** green
**Files:** `skills/flow/brainstorm.md`
**Tests:** none — one prose sentence; `scripts/check-verbatim-moves.sh` checks it
**Commit:** `docs(flow): begin a planning dispatch's record before its Agent call`
**After:** none

  - [ ] **Step 1: Apply the diff:**

````diff verified:git apply of this diff to the worktree at bef019f7; check-verbatim-moves.sh then FAILs exactly on the sentences verbatim-moves.txt lists, check-markdown-integrity.sh and check-references.sh exit 0
diff --git a/skills/flow/brainstorm.md b/skills/flow/brainstorm.md
index a446062e..381e304e 100644
--- a/skills/flow/brainstorm.md
+++ b/skills/flow/brainstorm.md
@@ -124,6 +124,10 @@ flow record dispatch begin -change <name> -role explorer|planner -model <m> -eff
 flow record dispatch end -change <name> -key explore-<n>|design-<n> -session-token mf-<literal-token> -outcome completed
 ```
 
+Run each `begin` immediately before its Agent call and its `end` once that agent returns — never
+both after it returns, which records a window of zero seconds (kan-1034: both explorer rows read
+0s).
+
 Plan mode (**Plan mode**, `skills/spectre-planning/SKILL.md`) is entered right after
 `flow.brainstorm` begins — `STARTED` and the worktree already exist — and before the Explore phase.
 On **Approve**, save the design under `<project>-worktrees/<name>/.superpowers/sdd/` and continue
````

  - [ ] **Step 2: Verify.** `scripts/check-markdown-integrity.sh`, `scripts/check-references.sh` and `scripts/check-verbatim-moves.sh` exit 0.

- [ ] 8. Skill index: name the re-vendor duty for spectre's copy

**Build:** green
**Files:** `CLAUDE.md`, `AGENTS.md`
**Tests:** none — an index row
**Commit:** `docs: name the re-vendor duty in the spectre-planning index row`
**After:** none

  - [ ] **Step 1: Edit both rows identically.** In the `skills/spectre-planning/` row of `CLAUDE.md` and of `AGENTS.md`, replace `and spectre vendors it for planning without flow |` with `and spectre vendors it for planning without flow — an edit here is re-copied byte for byte into spectre's vendored copy in the same change, as that copy's UPSTREAM.md says |`.
  - [ ] **Step 2: Verify.** `git grep -n -F -e "re-copied byte for byte into spectre's vendored copy" -- CLAUDE.md AGENTS.md` prints one line per file; `scripts/check-references.sh` exits 0.

- [ ] 9. spectre: declare the lint and test lists

**Build:** green
**Files:** `.flow/project.md`
**Tests:** none — project configuration; `run-verify-list.sh` running it is the check
**Commit:** `chore(flow): declare spectre's lint and test lists`
**After:** none

**Decision:** spectre-gofmt-fails

  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `/Users/tweety53/Projects/agents-worktrees/kan-1034-spectre-optional-bundled-planning-prompts-that/scripts/project-get.sh /Users/tweety53/Projects/spectre-worktrees/kan-1034-spectre-optional-bundled-planning-prompts-that lint` exits 1, and `… test` likewise.
    <!-- measured: project-get.sh <spectre worktree> lint|test, per self-review-pass.md angle 1 @ spectre 0b9073a → exit 1 -->
  - [ ] **Step 2: Append to spectre's `.flow/project.md`**, after `## default landing route`'s body, in the spectre worktree:

````markdown verified:each command exited 0 in the spectre worktree at 0b9073a
## lint

`gofmt -w .` is the auto-fix, run before the list. `gofmt -l` exits 0 whatever it lists, so the
list's line fails through `grep` when a file is unformatted.

```bash
! gofmt -l . | grep .
go vet ./...
```

## test

```bash
go test ./...
```
````

  - [ ] **Step 3: Verify.** From the agents worktree, `scripts/run-verify-list.sh /Users/tweety53/Projects/spectre-worktrees/kan-1034-spectre-optional-bundled-planning-prompts-that lint` and `… test` each exit 0.
