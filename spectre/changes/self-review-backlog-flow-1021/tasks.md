# self-review-backlog-flow-1021

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close five backlog findings — KAN-910, KAN-911, KAN-912, KAN-915 and KAN-906 — as one
self-review fix loop.

**Architecture:** a pinning table test and two prose sentences for the panel-fix guard; a
recorded-base refusal in `resolvebasebranch.go`, a landing branch in `/flow-self-review` and a
base-aware refresh in `integrate-change.sh cleanup`; three more columns in the dispatch upsert's
conflict clause; a report-file condition on the clean re-run in `panelfindingsclosed.go`; and a
pure move of cross-guard helpers into six shared files.

**Tech Stack:** Go (`stats/internal/guard`, `stats/internal/store`), bash, Markdown skills.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-21-self-review-backlog-flow-1021-design.md` in this worktree.

## Global Constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-1021`, branch `self-review-backlog-flow-1021`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- A prose diff given in a step is applied with `git apply` from the worktree's root after copying the lines between its fence markers into a scratch file outside the worktree; each was applied and checked at `1c4c3e38`. A diff that does not apply is reported, never hand-merged by guesswork. Each prose sentence lands character for character: `scripts/check-verbatim-moves.sh` matches the reworded ones against `verbatim-moves.txt` exactly.
- Run `cd stats && gofmt -w .` on Go changes before committing; `gofmt -l .` and `go vet ./...` stay clean (both after `make -s web-dist`). Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container; the store tests create and drop their own databases in it, as they already do.

## Review Focus

- Task 2: a recorded base origin still carries resolves exactly as before, an unreachable remote never refuses, and `origin/HEAD`-resolved bases never reach the `ls-remote` call.
- Task 2: a main checkout on the default branch is still refreshed onto it when the change recorded a `--base`.
- Task 3: a journal replay of a begin (identical fields) still returns the original row and seq.
- Task 4: a joined re-run (`bugbot+primary`) covers a `primary` finding when `panel-report-<round>-primary.md` is written, whatever `panel-report-<round>-bugbot.md` holds.
- Task 5: no function body, comment text or exported name changes — `git diff --stat` shows moves, and Step 1's selector is green before and after.

- [ ] 1. KAN-910 — pin the fix-dispatch shapes check-panel-fix-single-dispatch accepts
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Read the guard's role and
    key filter (`pfdPanelFixRows`, `stats/internal/guard/panelfixsingledispatch.go`) and confirm
    the four shapes below behave as the test asserts on the untouched tree: the test is a pin of
    behaviour the guard already has, not a fix.
  - [ ] **Step 2: Add the test** `TestCheckPanelFixSingleDispatchNonPanelFixShapes` to
    `stats/internal/guard/check_panel_fix_single_dispatch_test.go`, a table run through the
    file's existing `pfdExpect` helper, raw JSON rows as case 6b writes them:

```go verified:added as a scratch test file at 1c4c3e38; go test ./internal/guard/ -run TestCheckPanelFixSingleDispatchNonPanelFixShapes passes on the untouched guard
func TestCheckPanelFixSingleDispatchNonPanelFixShapes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		label, rows, needle string
		want                int
	}{
		{"a gated fix round under role implementer is no panel-fix row",
			`[{"key":"task-1+2-implementer-fix-1","role":"implementer","sessionToken":"mf-tok","agentId":"inline"}]`, "OK", 0},
		{"the parent's inline full-suite-fix row is skipped",
			`[{"key":"full-suite-fix-1","role":"panel-fix","sessionToken":"mf-tok","agentId":"inline"}]`, "OK", 0},
		{"a subagent's full-suite-fix row is refused (KAN-482)",
			`[{"key":"full-suite-fix-1","role":"panel-fix","sessionToken":"mf-tok","agentId":"a1b2c3"}]`, "full-suite-fix-1", 1},
		{"a gated fix round misrecorded under role panel-fix is refused",
			`[{"key":"task-1+2-implementer-fix-1","role":"panel-fix","sessionToken":"mf-tok","agentId":"inline"}]`, "task-1+2-implementer-fix-1", 1},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			pfdExpect(t, pfdCase{dispatches: []byte(c.rows)}, c.want, c.needle)
		})
	}
}
```

  - [ ] **Step 3: State the roles.** Apply the prose diff (the key rule's new sentence):

````diff verified:applied with git apply at 1c4c3e38; check-references.sh, check-installed-citations.sh, check-markdown-integrity.py and check-verbatim-moves.sh exit 0
diff --git a/skills/flow/review-panel-fix-round.md b/skills/flow/review-panel-fix-round.md
index e5f11397..d163b443 100644
--- a/skills/flow/review-panel-fix-round.md
+++ b/skills/flow/review-panel-fix-round.md
@@ -423,7 +423,7 @@ chunk's `-key` is exactly `panel-fix-<round>`; each further chunk appends `-<n>`
 (`panel-fix-<round>-2`, `-3`, …), contiguous from 2; the handshake retry suffixes `-retry` onto
 its own chunk's key (`panel-fix-<round>[|-<n>]-retry`), and any other key shape is a violation
 whatever the dispatch count, save the parent's own `full-suite-fix-<n>` rows recorded with
-`-agent-id inline` (`skills/flow/implement.md`). A round may not chunk freely: its chunk count is bounded by
+`-agent-id inline` (`skills/flow/implement.md`). A gated reviewer's fix round is no panel-fix row: it records under `-role implementer` (`skills/flow/gated-review-fix.md`), and an in-run verify fix under its appended task's implementer pair (`skills/flow/verify-fix-loop.md`). A round may not chunk freely: its chunk count is bounded by
 `ceil(findings raised in earlier rounds / 10)`. Before recording each `dispatch begin`, confirm that key has
 no panel-fix begin already recorded — a second begin under a fresh key is over-dispatching even
 when every key is well-formed. `check-panel-fix-single-dispatch.sh` holds every run's panel close
````

    and in `scripts/check-panel-fix-single-dispatch.sh`'s header, after the paragraph ending
    `A subagent's row under that key still counts.`, add the comment line
    `# A gated reviewer's fix round records under role implementer (gated-review-fix.md), so its`
    and `# task-<n+n>-implementer-fix-<k> rows never reach this check.`
  - [ ] **Step 4: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckPanelFixSingleDispatch' -count=1`
    passes; `gofmt -l internal/guard` prints nothing; `scripts/check-references.sh`,
    `scripts/check-installed-citations.sh`, `python3 scripts/check-markdown-integrity.py`,
    `scripts/check-guard-symlinks.sh` and `scripts/check-verbatim-moves.sh` exit 0.
  - [ ] **Step 5: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/check_panel_fix_single_dispatch_test.go`, `skills/flow/review-panel-fix-round.md`, `scripts/check-panel-fix-single-dispatch.sh`
**Tests:** `TestCheckPanelFixSingleDispatchNonPanelFixShapes`
**Regression:** `TestCheckPanelFixSingleDispatchNonPanelFixShapes` — a pin of the shapes the guard already accepts and refuses; reverting the commit removes the pin, and a later change that skips `task-…-implementer-fix-…` rows under `panel-fix` or a subagent's `full-suite-fix-<n>` would no longer fail a test.
**Baseline:** before=2 after=3
<!-- measured: grep -c '^func Test' stats/internal/guard/check_panel_fix_single_dispatch_test.go @ 1c4c3e38 -->
**Commit:** test(guard): pin the fix-dispatch shapes the panel-fix single-dispatch check accepts
**After:** none

**Decision:** non-panel-fix-roles-stated

- [ ] 2. KAN-911 — handle a `--base` change's deleted base, standalone bundle and main-checkout refresh
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In a scratch clone whose
    branch records `git config branch.<cur>.flowBase feature/gone` and whose origin carries no
    `feature/gone`, `scripts/resolve-base-branch.sh <clone>` prints `feature/gone` and exits 0.
  - [ ] **Step 2: Write the failing Go test** `TestResolveBaseBranchRefusesARecordedBaseGoneFromOrigin`
    in `stats/internal/guard/resolve_base_branch_test.go`: a copy of the fixture's `wt`
    (`rbbCopy`) recording `branch.spectre/fixture.flowBase` = `feature/gone`, which the fixture's
    remote does not carry; run the registered `resolve-base-branch` guard as the file's
    `TestResolveBaseBranch` runs it; assert exit 1, empty stdout, and stderr containing
    `recorded base 'feature/gone' is not on origin`. It fails on the untouched code (exit 0,
    stdout `feature/gone`).
  - [ ] **Step 3: Keep the existing recorded-base cases honest.** Case 14 (`recorded flowBase wins
    over origin/HEAD`) records `feature/offline-mode`, which the fixture remote does not carry and
    which the refusal would now reject: add `{"-C", seed, "push", "-q", "origin", "main:feature/offline-mode"}`
    to the fixture's seed steps so the remote carries it, and leave the case's expectation as is.
    Run every suite that records a `flowBase` against a fixture remote —
    `go test ./internal/guard/ -run 'TestResolveBaseBranch|TestKickoffWorktree|TestVerbatimMoves' -count=1`,
    `bash scripts/test-integrate-change.sh`, `bash scripts/test-check-task-records.sh`,
    `bash scripts/test-flow-task-list-landed.sh` — and give any fixture the refusal now trips the
    same remote branch, never a weaker check.
  - [ ] **Step 4: Implement the refusal** in `resolveBaseBranch`
    (`stats/internal/guard/resolvebasebranch.go`): keep the `recordedBase` result in a variable;
    after the name-validation loop, only when the base came from it, run
    `git -C <dir> -c core.askpass=true ls-remote --exit-code --heads origin refs/heads/<base>`
    under `LC_ALL=C` and `GIT_TERMINAL_PROMPT=0`; exit status 2 prints
    `resolve-base-branch: recorded base '<base>' is not on origin — re-point the change (skills/flow/SKILL.md)`
    to stderr and returns 1; any other status resolves as today. Add the case to the exit-1 list
    of `scripts/resolve-base-branch.sh`'s header and to its `A RECORDED BASE WINS.` paragraph.
  - [ ] **Step 5: Write the failing harness scenario** `refresh_on_recorded_base` in
    `scripts/test-integrate-change.sh`: the full prepare → commit → land run of the existing
    happy path, but with worktree `a` recording `branch.spectre/demo.flowBase` = `feature/x` (cut
    from and landing on `origin/feature/x`, which the scenario pushes first) and `a`'s main
    checkout on `feature/x`; check `refresh_on_recorded_base: a's main checkout on its recorded base is refreshed, not refused`
    — `$OUT` carries a `REFRESH-DONE: $WORK/a` or `REFRESH-CURRENT: $WORK/a` line and no
    `REFRESH-REFUSED: $WORK/a` line. It fails on the untouched script with
    `REFRESH-REFUSED: $WORK/a is on feature/x, not main`.
  - [ ] **Step 6: Implement the refresh base** in `cmd_cleanup` (`scripts/integrate-change.sh`):
    in the `flow.verify-merge` loop store each worktree's `$BASE` as `RBASE[i]`; in the
    `flow.refresh-main-checkout` loop take `origin/HEAD`'s branch as today, then, for each `i`
    with `REPO[i]` equal to `$m` and `RBASE[i]` equal to `git -C "$m" branch --show-current`,
    use `RBASE[i]` instead. Update the header's `refresh-main-checkout per repository` line to
    say so.
  - [ ] **Step 7: Apply the prose diff** — the recovery sentence, the landing branch and the run-2
    refresh sentence:

````diff verified:applied with git apply at 1c4c3e38; check-references.sh, check-installed-citations.sh, check-markdown-integrity.py and check-verbatim-moves.sh (with this change's verbatim-moves.txt) exit 0
diff --git a/skills/flow-contracts/finish-contract-run2.md b/skills/flow-contracts/finish-contract-run2.md
index 8580e8a1..ea5bc911 100644
--- a/skills/flow-contracts/finish-contract-run2.md
+++ b/skills/flow-contracts/finish-contract-run2.md
@@ -119,6 +119,10 @@ procedure for them is there.
    refresh-main-checkout.sh <main-checkout> <base>
    ```
 
+   A main checkout that has a `--base` change's `<base>` (`skills/flow/SKILL.md`) checked out is
+   refreshed onto that `<base>` instead — the branch the merge check resolved for that
+   repository's worktree before removing it — so it is brought forward rather than refused.
+
    The script fast-forwards the main checkout onto `origin/<base>` only when that loses nothing,
    and refuses by name otherwise; its header is canonical for the conditions. `REFRESH-REFUSED`
    and `FLOWD-RESTART-FAILED` are reported in the handoff, the latter with its `make` command for
diff --git a/skills/flow-self-review/SKILL.md b/skills/flow-self-review/SKILL.md
index f75aa4c5..d8f880c8 100644
--- a/skills/flow-self-review/SKILL.md
+++ b/skills/flow-self-review/SKILL.md
@@ -3,7 +3,7 @@ name: flow-self-review
 description: Run a change's self-review pass, a one-shot opus dispatch — inside `/flow`'s integrate and `/flow-fast`'s verify, or standalone from the context bundle an older run saved; plan in full, fix and land every finding that is not big, file the big ones, mine the flow-eval cases the operator approves, record each in the flow store, rate, write the report, delete the bundle. Standalone, not a pipeline stage. Use for /flow-self-review.
 allowed-tools: Bash(git:*), Bash(flow:*), Bash(go:*), Bash(scripts/check-self-review-report.sh:*), Bash(land-self-review-report.sh:*)
 license: MIT
-compatibility: Standalone, requires the change's default branch to be checked out and a saved context bundle at docs/self-review/<name>-context.md.
+compatibility: Standalone, requires the change's default branch, or its `--base` branch, to be checked out and a saved context bundle at docs/self-review/<name>-context.md.
 ---
 
 Run a change's self-review reasoning pass — the only one the pipeline has — over the bundle
@@ -28,8 +28,9 @@ per-change state file, and marks no `flow stage` call.
 
 `<bundle>` is `<project>/docs/self-review/<name>-context.md`. Absent: print every
 `docs/self-review/*-context.md` present (or `none pending` when there are none) and stop.
+The absent case also prints every `origin/<branch>` whose tree holds `<bundle>` (`git cat-file -e origin/<branch>:<bundle-path>`, `<bundle-path>` the bundle's path relative to `<project>`, over `git for-each-ref --format='%(refname:short)' refs/remotes/origin`), so a `--base` change's bundle (`skills/flow/SKILL.md`), landed on its `<base>`, is found from any checkout.
 
-`git branch --show-current` must be `<default-branch>` (the project's own default branch). A
+`git branch --show-current` must be `<landing-branch>`: `<default-branch>` (the project's own default branch), or a `--base` change's `<base>` — a branch other than the change's own (`spectre/<name>`, `<name>` on `/flow-fast`) whose `origin/<branch>` tree holds `<bundle>`. A
 checkout on any other branch: print the branch and stop.
 
 **Both refusals bind the standalone command only.** A pass run inside `/flow`'s integrate or
@@ -438,11 +439,11 @@ pass and the filing-and-rating prompt above is long enough that the branch is wo
 rather than trusted from step 1 alone:
 
 ```bash
-land-self-review-report.sh "<project>" "<default-branch>" \
+land-self-review-report.sh "<project>" "<landing-branch>" \
   "docs(self-review): <name> self-review report" \
   "docs/self-review/<name>-self-review.md" \
   "docs/self-review/<name>-context.md" \
-  --push "<default-branch>"
+  --push "<landing-branch>"
 ```
 
 A branch mismatch stops here — nothing is committed, pulled, or pushed. Staged work in the
@@ -457,7 +458,7 @@ End naming the report path, the rating, the shas landed, the Jira keys filed and
 ## Guardrails
 
 - **Never** push directly to a protected branch under any name but the project's own default
-  branch, and never force-push.
+  branch, or the `<landing-branch>` step 1 accepted, and never force-push.
 
 ## Commands (user-facing)
 
diff --git a/skills/flow/SKILL.md b/skills/flow/SKILL.md
index 7c5c8071..be8f2c06 100644
--- a/skills/flow/SKILL.md
+++ b/skills/flow/SKILL.md
@@ -25,6 +25,7 @@ lands on, in place of the one `origin/HEAD` points at. Kickoff records it on the
 every repository (`kickoff-worktree.sh <project> <name> <base>`, and
 `kickoff-worktree.sh --from <this change's worktree>` for every worktree added after kickoff), and `resolve-base-branch.sh` reads
 it back for every later phase, so it is never passed again.
+A recorded base gone from origin — merged and deleted — is refused by name by `resolve-base-branch.sh` (`<agents repo>/scripts/resolve-base-branch.sh`): re-point every worktree of the change at the branch that absorbed it with the three commands integrate's `STOP: base-mismatch` answer names (`skills/flow/integrate.md`), then re-run.
 
 ## Model resolution
 
````

  - [ ] **Step 8: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestResolveBaseBranch|TestKickoffWorktree|TestVerbatimMoves' -count=1`
    passes; `bash scripts/test-integrate-change.sh`, `bash scripts/test-check-task-records.sh`
    and `bash scripts/test-flow-task-list-landed.sh` pass; `gofmt -l internal/guard` prints
    nothing; `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py`, `scripts/check-guard-symlinks.sh` and
    `scripts/check-verbatim-moves.sh` exit 0.
  - [ ] **Step 9: Commit** with the subject below.

**Build:** green
**Files:** `stats/internal/guard/resolvebasebranch.go`, `stats/internal/guard/resolve_base_branch_test.go`, `scripts/resolve-base-branch.sh`, `scripts/integrate-change.sh`, `scripts/test-integrate-change.sh`, `skills/flow/SKILL.md`, `skills/flow-self-review/SKILL.md`, `skills/flow-contracts/finish-contract-run2.md`
**Allowed-collateral:** `stats/internal/guard/*_test.go`, `scripts/test-*.sh`
**Tests:** `TestResolveBaseBranchRefusesARecordedBaseGoneFromOrigin`, `refresh_on_recorded_base`
**Regression:** `TestResolveBaseBranchRefusesARecordedBaseGoneFromOrigin` — reverting resolves a base origin no longer carries, exit 0. `refresh_on_recorded_base` — reverting refreshes `a`'s main checkout onto `main` and prints `REFRESH-REFUSED`.
**Commit:** fix(guard): refuse a recorded base gone from origin and refresh a main checkout on its --base
**After:** none

**Decision:** recorded-base-gone-refused
**Decision:** self-review-landing-branch
**Decision:** refresh-onto-checked-out-base

- [ ] 3. KAN-912 — update slot, model and effort on a repeated dispatch begin
  - [ ] **Step 1: Write the failing tests** in `stats/internal/store/records_test.go`, beside
    `TestRecordDispatchReopensAnAbortedKey`:
    `TestRecordDispatchReBeginUpdatesSlotModelAndEffort` — begin key `panel-1-primary` (token
    `tok`, role `reviewer`, slot `primary`, model `opus`, effort `high`), begin it again with slot
    `principles`, model `sonnet`, effort `medium`; the returned row and `RunRecord`'s single
    dispatch carry `principles`, `sonnet`, `medium`, the same ID and seq, still open.
    `TestRecordDispatchReBeginOfAnAbortedKeyUpdatesSlotModelAndEffort` — the same, with an
    `EndDispatch` at outcome `aborted` between the two begins; the re-begun row is open and
    carries the second begin's three values. Run
    `cd stats && make -s web-dist && go test ./internal/store/ -run 'TestRecordDispatchReBegin' -count=1`:
    both fail on the untouched code, the row still on `primary`, `opus`, `high`.
  - [ ] **Step 2: Implement** in `insertDispatch` (`stats/internal/store/records.go`): add to the
    `ON CONFLICT ON CONSTRAINT … DO UPDATE SET` list, unconditionally,

```sql unverified:confirm EXCLUDED.slot already holds canonical_slot($6) and EXCLUDED.effort the defaulted value
			slot = EXCLUDED.slot,
			model = EXCLUDED.model,
			effort = EXCLUDED.effort,
```

    and extend the doc comment above `reBegin` with one paragraph: a repeated begin takes the
    new begin's slot, model and effort, last begin wins — a journal replay carries the same
    values, so it stays a no-op — so a re-dispatch on a new pair never keeps the stale one
    (KAN-912).
  - [ ] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/store/ -run 'TestRecordDispatch' -count=1`
    passes (`TestRecordDispatchIsIdempotentUnderTheSameKey` included); `go vet ./internal/store`
    and `gofmt -l internal/store` print nothing.
  - [ ] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `stats/internal/store/records.go`, `stats/internal/store/records_test.go`
**Tests:** `TestRecordDispatchReBeginUpdatesSlotModelAndEffort`, `TestRecordDispatchReBeginOfAnAbortedKeyUpdatesSlotModelAndEffort`
**Regression:** `TestRecordDispatchReBeginUpdatesSlotModelAndEffort` — reverting keeps an open row's first slot, model and effort on a repeated begin. `TestRecordDispatchReBeginOfAnAbortedKeyUpdatesSlotModelAndEffort` — reverting keeps the aborted attempt's three values on the reopened row.
**Baseline:** before=88 after=90
<!-- measured: grep -c '^func Test' stats/internal/store/records_test.go @ 1c4c3e38 -->
**Commit:** fix(store): update slot, model and effort on a repeated dispatch begin
**After:** none

**Decision:** rebegin-last-wins
**Decision:** rebegin-no-repair

- [ ] 4. KAN-915 — count a panel re-run as clean only once its report file is written
  - [ ] **Step 1: Write the failing test** `TestCheckPanelFindingsClosedRerunNeedsItsReportFile`
    in `stats/internal/guard/check_panel_findings_closed_test.go`, shaped like
    `TestCheckPanelFindingsClosedReportCounts` (its own `wt` per case, report files written under
    `.superpowers/sdd/`). Finding `F1`, Important, `fixed`, round 1, slot `primary`, unless a case
    says otherwise:
    - dispatch `panel-2-primary` (reviewer, slot `primary`, completed), no report file → exit 1,
      stderr naming `panel-report-2-primary.md`;
    - the same with `panel-report-2-primary.md` = `no verbatim report captured — timeout\n` → exit 1, naming it;
    - the same with `panel-report-2-primary.md` = `Critical: 0\nImportant: 0\n` → exit 0;
    - dispatch `panel-2-bugbot+primary` (slot `bugbot+primary`) with only
      `panel-report-2-primary.md` written → exit 0;
    - finding slot `bugbot+primary`, the same dispatch, only `panel-report-2-primary.md` written → exit 1, naming `panel-report-2-bugbot.md`.

    Run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckPanelFindingsClosedRerunNeedsItsReportFile' -count=1`:
    the missing and placeholder cases fail on the untouched code (exit 0).
  - [ ] **Step 2: Implement** in `checkPanelFindingsClosed`
    (`stats/internal/guard/panelfindingsclosed.go`): add
    `func cfcReportWritten(worktree string, round int, id string) bool` — true when
    `<worktree>/.superpowers/sdd/panel-report-<round>-<id>.md` is non-empty and does not begin
    with `no verbatim report captured`. In `covers`, a dispatch provides a required component
    only when `cfcReportWritten(worktree, d.round, component)` holds. For every finding left
    unverified for which some later completed reviewer row would have covered it but for a
    missing report, collect that report's file name, and print one more line,
    `check-panel-findings-closed: re-run(s) answered without writing their report file: <names>`,
    beside the existing unverified line. Extend the comment block above `covers` with the rule
    and KAN-915.
  - [ ] **Step 3: Keep the existing cases honest.** `TestCheckPanelFindingsClosed` and
    `TestCheckPanelFindingsClosedReportCounts` cases that expect a clean re-run to cover a fixed
    Critical or Important now need that re-run's report: write `panel-report-<round>-<id>.md`
    with a body carrying no `Critical:`/`Important:` lines (so the count check skips it) into
    the shared `wt` for `cfcDefaultDispatches`' rounds and every slot id a passing case's
    dispatches name, and into the per-case files of `ReportCounts`' passing case. A case whose
    expectation changes is a defect in the plan — report it, never edit the expectation.
  - [ ] **Step 4: Update the contract** in `scripts/check-panel-findings-closed.sh`'s header:
    the exit-0 line says each clean re-run's `panel-report-<round>-<id>.md` is written, and the
    exit-1 list adds a re-run that answered without writing it (absent, empty or the parent's
    placeholder).
  - [ ] **Step 5: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckPanelFindingsClosed' -count=1`
    passes; `gofmt -l internal/guard` prints nothing; `scripts/check-guard-symlinks.sh` exits 0.
  - [ ] **Step 6: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/panelfindingsclosed.go`, `stats/internal/guard/check_panel_findings_closed_test.go`, `scripts/check-panel-findings-closed.sh`
**Tests:** `TestCheckPanelFindingsClosedRerunNeedsItsReportFile`
**Regression:** `TestCheckPanelFindingsClosedRerunNeedsItsReportFile` — reverting counts a completed re-run with no report, or the parent's placeholder, as the clean verification, and the missing and placeholder cases exit 0.
**Baseline:** before=2 after=3
<!-- measured: grep -c '^func Test' stats/internal/guard/check_panel_findings_closed_test.go @ 1c4c3e38 -->
**Commit:** fix(guard): count a panel re-run as clean only once its report file is written
**After:** none

**Decision:** rerun-needs-report

- [ ] 5. KAN-906 — move cross-guard helpers into shared files
  - [ ] **Step 1: Record the green baseline.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'Test(Check)?(Panel|CleanupComplete|StageMark|TaskCommit|References|WorktreeLocation|GatherDispatch|RunReproducer|ChangePlan|SpecRoot|FinishPreflight|BaseMoved|UnfinishedWork|SyncPanel|WritePanel)' -count=1`
    passes before any move — the tests of every guard whose file loses a helper.
  - [ ] **Step 2: Move**, each function with its doc comment and any unexported declaration
    (type, var, const, regexp) only the moved functions use, unchanged, from its file to the
    new file named — no renames, no merges, no body or comment edits; imports follow the code:
    - `stats/internal/guard/paths.go`: `isDir`, `isFile` (`specroot.go`); `pcAbs`
      (`panelreproducers.go`); `smcAbs` (`stagemarkcalls.go`); `gdcDirname`
      (`gatherdispatch.go`); `hasDotDotSegment` (`runreproducer.go`); `siblingRoot`
      (`worktreelocation.go`).
    - `stats/internal/guard/exec.go`: `capture`, `envGit` (`resolveremotebase.go`); `lookPath`,
      `gitOut` (`changeplan.go`); `panelResolveGit` (`paneltouchedpaths.go`); `exitCode`
      (`finishpreflight.go`); `rrExitCode` (`runreproducer.go`).
    - `stats/internal/guard/flowcli.go`: `pcFlow`, `pcRaw` (`panelexitcontract.go`); `pfdRead`,
      `pfdStream` (`panelfixsingledispatch.go`).
    - `stats/internal/guard/text.go`: `lines` (`changeplan.go`); `crSort` (`references.go`);
      `ccSanitize`, `ccASCIILower` (`cleanupcomplete.go`); `smcUTF8` (`stagemarkcalls.go`).
    - `stats/internal/guard/changename.go`: `plainChangeName` (`cleanupcomplete.go`).
    - `stats/internal/guard/planparse.go`: `tcfPlanLines`, `tcfParseTask`, `tcfResolveTasksMD`,
      `tcfUnfenced` (`taskcommitfields.go`).

    Open each new file with one comment line saying what it holds, e.g.
    `// paths.go holds the path helpers several guards share.`
  - [ ] **Step 3: Update the citations.** `git grep -n -E '(isDir|isFile|pcAbs|smcAbs|gdcDirname|hasDotDotSegment|siblingRoot|capture|envGit|lookPath|gitOut|panelResolveGit|exitCode|rrExitCode|pcFlow|pcRaw|pfdRead|pfdStream|lines|crSort|ccSanitize|ccASCIILower|smcUTF8|plainChangeName|tcfPlanLines|tcfParseTask|tcfResolveTasksMD|tcfUnfenced)[^A-Za-z]' -- stats/internal/guard scripts`
    and change every comment that names a moved function beside its old file (`plainChangeName
    (cleanupcomplete.go)` and the like) to the new file; nothing else.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -w internal/guard && make -s web-dist && go vet ./... && go test ./internal/guard/ -run 'Test(Check)?(Panel|CleanupComplete|StageMark|TaskCommit|References|WorktreeLocation|GatherDispatch|RunReproducer|ChangePlan|SpecRoot|FinishPreflight|BaseMoved|UnfinishedWork|SyncPanel|WritePanel)' -count=1`
    passes and `gofmt -l .` prints nothing; `git diff --stat` shows only moves (added and removed
    line counts equal per moved block, plus the new files' header comments and import blocks);
    `scripts/check-guard-symlinks.sh` and `scripts/check-references.sh` exit 0.
  - [ ] **Step 5: Commit** with the subject below.

**Build:** green
**Files:** `stats/internal/guard/paths.go`, `stats/internal/guard/exec.go`, `stats/internal/guard/flowcli.go`, `stats/internal/guard/text.go`, `stats/internal/guard/changename.go`, `stats/internal/guard/planparse.go`, `stats/internal/guard/specroot.go`, `stats/internal/guard/panelreproducers.go`, `stats/internal/guard/stagemarkcalls.go`, `stats/internal/guard/gatherdispatch.go`, `stats/internal/guard/runreproducer.go`, `stats/internal/guard/worktreelocation.go`, `stats/internal/guard/resolveremotebase.go`, `stats/internal/guard/changeplan.go`, `stats/internal/guard/paneltouchedpaths.go`, `stats/internal/guard/finishpreflight.go`, `stats/internal/guard/panelexitcontract.go`, `stats/internal/guard/panelfixsingledispatch.go`, `stats/internal/guard/references.go`, `stats/internal/guard/cleanupcomplete.go`, `stats/internal/guard/taskcommitfields.go`
**Allowed-collateral:** `stats/internal/guard/*.go`, `scripts/*.sh`
**Tests:** **none** — a move with no behaviour change; the guard package's existing suite is the check.
**Regression:** none — no test is declared.
**Commit:** refactor(guard): move cross-guard helpers into shared files
**After:** Task 1, 4

**Decision:** guard-helpers-move-only
