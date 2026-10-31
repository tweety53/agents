# self-review-kan-1052-rework-the-stage-leaderboard-dashboard

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1052's seven self-review findings — F1 (Task 1), F2 (Task 2), F5 (Task 3),
F3 (Task 4), F4 (Task 5), F6 (Task 6), F16 (Task 7), as `proposal.md` names them.

**Architecture:** one `ui paths` glob with a Makefile-derived guard test; a `visual` list in
`run-verify-list` that `/flow-fast` and `/flow-interactive` call; a `GOMAXPROCS` share on the two
Go suite lines, measured before and after; a corrected key shape; a marker-based kill assertion;
one sentence in the mutation brief; one planning rule, re-vendored into spectre.

**Tech Stack:** Go (`stats/`), Bash, Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-31-self-review-kan-1052-rework-the-stage-leaderboard-dashboard-design.md`
in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject, no `Task-Id:` trailer and no `Co-Authored-By:` trailer. Never write `spectre/changes/`. Never merge or push.
- Tasks 1–7 work in the agents worktree `/Users/tweety53/Projects/agents-worktrees/kan-1052-rework-the-stage-leaderboard-dashboard` (`<wt>` below), branch `spectre/kan-1052-rework-the-stage-leaderboard-dashboard`, committing after its existing commits and rewriting none. Task 7 also works in the spectre worktree `/Users/tweety53/Projects/spectre-worktrees/self-review-kan-1052-rework-the-stage-leaderboard-dashboard` (`<swt>` below), branch `self-review-kan-1052-rework-the-stage-leaderboard-dashboard`, which the parent creates off spectre's `origin/main` before Task 7's dispatch and lands by spectre's `## default landing route`, `merge and push`, after the agents branch lands (decision `spectre-worktree-by-parent`). Every `**Files:**` path is relative to its repository's worktree root. Never touch a main checkout, and never create a worktree.
- Go: `gofmt -w` every Go file you change; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Mutation proofs restore with `scripts/break-and-prove.sh`, or on a scratch copy outside the worktree — never `git checkout --` on a file carrying uncommitted edits.
- Never stop, restart or drop `flowd` on 4173, the `flow-postgres` container or the `flow` database. A `verify` row run (Task 2) uses the disposable UI-test stack on 4174 through the row itself, with a Bash `timeout` of 600000.
- No timeout, wall-clock bound or retry count is widened anywhere (`rules/fix-determinism-at-the-source.mdc`).

## Review Focus

- Task 2: a `visual` list with no `verify` row, a `decision.json` recording `visual.verify` `skipped`, or a diff outside `ui paths` runs nothing and exits 0; a `lint` list still never runs the row; the `test` list's tail is unchanged.
- Task 3: the share is at least 1 on a one- or two-core machine; the `unchanged.sh` short-circuit of both lines still skips a suite before the prefix applies.
- Task 5: case (a) fails when the cap never kills, whatever the elapsed time.
- Task 7: the agents and spectre `SKILL.md` copies are byte-identical (`cmp`).

- [x] 1. ui paths covers the model-policy file flowd embeds
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `printf 'skills/flow-contracts/model-policy.md\n' | scripts/check-visual-trigger.sh .` from `<wt>` prints `VISUAL-TRIGGER-NO-MATCH` and exits 1.
  - [x] **Step 2: Write the test** `TestCheckVisualTriggerCoversEmbeddedSources` in `stats/internal/guard/check_visual_trigger_test.go`: read `../../Makefile` (`stats/Makefile`), collect every source path of a `cp ../<src>` command by the regular expression `cp \.\./(\S+)` — require at least one, else `t.Fatal` — and for each run `Registry["check-visual-trigger"]` with the absolute repository root (`filepath.Abs("../../..")`) as its argument and `<src>\n` on stdin; each must exit 0 with `MATCH: <src>` on stdout. Run `cd stats && make -s web-dist && go test ./internal/guard -run 'TestCheckVisualTriggerCoversEmbeddedSources' -count=1` — it fails naming `skills/flow-contracts/model-policy.md`.
  - [x] **Step 3: Fix.** In `.flow/project.md`'s `## visual verification` table, set the `ui paths` cell to `` `stats/web/src/**`, `skills/flow-contracts/model-policy.md` `` and add, to the paragraph above the table that explains the section, one sentence: "`ui paths` also names `skills/flow-contracts/model-policy.md`, which `stats/Makefile`'s `policy-copy` embeds into flowd for the Experiments view (kan-1052 self-review, F1)."
  - [x] **Step 4: Verify.** The test passes; Step 1's command prints `VISUAL-TRIGGER-MATCH` and exits 0; `scripts/check-visual-verification.sh .` and `printf 'stats/web/src/App.tsx\n' | scripts/check-visual-trigger.sh .` exit 0; `cd stats && gofmt -l internal/guard` prints nothing. Commit.

**Build:** green
**Files:** `.flow/project.md`, `stats/internal/guard/check_visual_trigger_test.go`
**Tests:** `TestCheckVisualTriggerCoversEmbeddedSources`
**Regression:** `TestCheckVisualTriggerCoversEmbeddedSources` fails — an edit to the embedded model-policy file runs no `verify` row again.
**Baseline:** before=3 after=4 <!-- measured: grep -c -E '^func Test' stats/internal/guard/check_visual_trigger_test.go @ fab4520a -->
**Commit:** fix(flow): cover the embedded model-policy file in ui paths
**After:** none

**Decision:** ui-paths-embedded-policy

- [x] 2. A visual list runs the verify row for /flow-fast and /flow-interactive
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `scripts/run-verify-list.sh <wt> visual` prints the usage line and exits 2.
  - [x] **Step 2: Write the test** `TestRunVerifyListVisualList` in `stats/internal/guard/run_verify_list_test.go`. Extract `TestRunVerifyListRunsTheVisualVerify`'s environment, `git` and `project` helpers and its seed→origin→`wt` fixture into one helper both tests call, keeping that test's cases and assertions unchanged. Cases, each on the fixture's project (whose `## lint` and `## test` lines `echo LIST-RAN`): a change to `web/src/a.tsx` with `verify` `echo VISUAL-RAN <specs>` exits 0, prints `VISUAL-RAN`, never `LIST-RAN`; `verify` `exit 1` exits 1; a change to `docs/a.md` exits 0, prints neither, and prints `- visual verify — not run`; a `decision.json` recording `visual.verify` `skipped` exits 0 and prints no `VISUAL-RAN`. Run `cd stats && make -s web-dist && go test ./internal/guard -run 'TestRunVerifyList(VisualList|RunsTheVisualVerify)' -count=1` — `TestRunVerifyListVisualList` fails on the usage exit.
  - [x] **Step 3: Implement** in `runVerifyList` (`stats/internal/guard/runverifylist.go`): accept `visual` beside `lint` and `test` (usage `lint|test|visual`); a `visual` list skips `project-get.sh` and the command loop, runs the existing visual block — its condition widened from `list == "test"` to `list != "lint"` — and, when `rvlVisualVerify` returns ok false, prints ``- visual verify — not run: no `verify` row, `visual.verify` skipped, or no `ui paths` change`` and exits 0. Update the file's header comment and `scripts/run-verify-list.sh`'s `Usage:` and exit-code lines to name the `visual` list: it runs the `verify` row alone, as a `test` list's tail would, and never the `## lint` or `## test` commands (kan-1052 self-review, F2).
  - [x] **Step 4: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, restore the visual block's condition to `list == "test"` — `TestRunVerifyListVisualList` fails; restore.
  - [x] **Step 5: Skills.** In section **5. Verify** of `skills/flow-fast/SKILL.md`, after the sentence naming `check-fast-route-record.sh`, add: "Then `<agents repo>/scripts/run-verify-list.sh <worktree> visual` runs the `## visual verification` `verify` row when the diff touches `ui paths` — the committed baselines a `lint` list never checks; a failure that is the change's intended look is recaptured with the section's `capture` row and committed under section 4's commit rule before the run lands (kan-1052 self-review, F2)." In step **4. Verify** of `skills/flow-interactive/SKILL.md`, insert `` `run-verify-list.sh <worktree> visual`, `` after the scoped `## test` commands in its list, before `check-fast-route-record.sh`.
  - [x] **Step 6: Verify.** The two tests pass; `cd stats && gofmt -l internal/guard` prints nothing; `scripts/run-verify-list.sh <wt> lint` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/runverifylist.go`, `stats/internal/guard/run_verify_list_test.go`, `scripts/run-verify-list.sh`, `skills/flow-fast/SKILL.md`, `skills/flow-interactive/SKILL.md`, `stats/internal/guard/pipelinefix.go`
**Tests:** `TestRunVerifyListVisualList`
**Regression:** `TestRunVerifyListVisualList` fails — `/flow-fast` and `/flow-interactive` have no command that runs the `verify` row, and a `ui paths` diff lands with stale baselines again.
**Baseline:** before=13 after=14 <!-- measured: grep -c -E '^func Test' stats/internal/guard/run_verify_list_test.go @ fab4520a -->
**Commit:** feat(guard): add a visual list that runs the verify row alone
**After:** none

**Decision:** visual-list

- [x] 3. The Go suites run on a share of the cores
  - [x] **Step 1: Measure before touching anything; do not guess.** From `<wt>`, `<scratch>` a directory outside the worktree, after `cd stats && make -s web-dist`: (i) alone — `cd stats/web && npx vitest run src/views/views.test.tsx -t 'state board' --reporter=json --outputFile=<scratch>/spa-alone.json` under `/usr/bin/time -p`; (ii) under load — in one Bash call with `timeout` 600000, start the Go load `cd stats && go test -race -count=1 -run '^TestCheck' ./internal/guard` and `cd stats && go test -race -count=1 -run '^TestRecord' ./cmd/flow` in the background, each under `/usr/bin/time -p`, wait 20 s for their compiles to start, run the vitest command of (i) writing `<scratch>/spa-before.json`, then `wait`. Record `getconf _NPROCESSORS_ONLN`, every command's `real`, `user` and `sys`, and each JSON's slowest test, `jq '[.testResults[].assertionResults[].duration] | max'`. The Go load is the two `-race` suites' heaviest packages; the vitest file holds the test that timed out.
  - [x] **Step 2: Write the test** `TestProjectTestListRunsGoSuitesOnACoreShare` in `stats/internal/guard/run_verify_list_test.go`: read `../../../.flow/project.md`, take the `## test` section's first fenced block, and for each `# suite: guard-tests` and `# suite: stats-go` marker require the command line beneath it to contain `GOMAXPROCS=$(( ($(getconf _NPROCESSORS_ONLN) + 2) / 3 ))`; a marker absent is a failure. Run `cd stats && make -s web-dist && go test ./internal/guard -run 'TestProjectTestListRunsGoSuitesOnACoreShare' -count=1` — it fails on both lines.
  - [x] **Step 3: Fix.** In `.flow/project.md`'s `## test` block, prefix the run after each `||` with the share: `GOMAXPROCS=$(( ($(getconf _NPROCESSORS_ONLN) + 2) / 3 )) scripts/run-guard-tests.sh`, and inside the `stats-go` braces `GOMAXPROCS=$(( ($(getconf _NPROCESSORS_ONLN) + 2) / 3 )) go test -race ...`. Below the block add one paragraph: the three suites run at once, so each Go suite runs on a third of the cores, rounded up — `GOMAXPROCS` bounds the go command's build parallelism, its default `-p`, and each test binary's threads and `-parallel` — beside the SPA's two worker threads (`stats/web/vite.config.ts`), so no suite starves another (kan-1052 self-review, F5: a 235 ms vitest test passed its 5 s timeout while both `-race` suites took every core).
  - [x] **Step 4: Measure again.** Repeat Step 1 (ii) with both Go load commands prefixed by the share, writing `<scratch>/spa-after.json`. Pass: the slowest `state board` test is under 2500 ms, half the 5 s vitest timeout, and the vitest `real` at most three times Step 1 (i)'s. On a fail, report the figures and stop — never raise a timeout or the share's divisor.
    <!-- predicted: 2500 ms is half vitest's default 5000 ms testTimeout; the 3x bound is chosen against the 11x slowdown full-suite-fix-2 measured (235 ms alone, 67.7 s suite) -->
  - [x] **Step 5: Verify.** The test passes; `cd stats && gofmt -l internal/guard` prints nothing. Commit, the body carrying Step 1's and Step 4's figures and the core count.

**Build:** green
**Files:** `.flow/project.md`, `stats/internal/guard/run_verify_list_test.go`
**Tests:** `TestProjectTestListRunsGoSuitesOnACoreShare`
**Regression:** `TestProjectTestListRunsGoSuitesOnACoreShare` fails — both `-race` suites take every core again, and the SPA suite starves under the concurrent list.
**Baseline:** before=14 after=15 <!-- measured: grep -c -E '^func Test' stats/internal/guard/run_verify_list_test.go @ fab4520a -->
**Commit:** fix(flow): run the Go suites on a share of the cores
**After:** Task 1, 2

**Decision:** go-suites-core-share

- [x] 4. The implementer key names every task of the group
  - [x] **Step 1: Edit.** In `skills/flow/implement.md`, replace the implementer key `task-<n>-implementer` with `task-<ids>-implementer` in the Dispatch-sites row "implementer, one per group" (adding after it ", ids `+`-joined in plan order"), in the inline-row sentence (`task-<ids>-implementer…`), in both `flow record dispatch` snippets of **The parent records each dispatch in two calls**, in the `-key` sentence after them, and in the implementer prompt's `Key:` line. Leave `task-<n>-implementer-recommit-<k>` unchanged. In `stats/internal/records/render.go` (`keyTask`'s comment), `stats/internal/records/render_test.go` (the comment above the `taskOf` cases) and `stats/internal/store/records.go` (`ErrInlineCommitMissing`'s comment), write the shape as `task-<ids>-implementer`.
  - [x] **Step 2: Verify.** `git grep -n -F -e 'task-<n>-implementer' -- skills/flow/implement.md stats/internal/records/render.go stats/internal/records/render_test.go stats/internal/store/records.go` prints only the `-recommit-<k>` occurrences; `cd stats && gofmt -l internal` prints nothing; `scripts/run-verify-list.sh <wt> lint` exits 0. Commit.

**Build:** green
**Files:** `skills/flow/implement.md`, `stats/internal/records/render.go`, `stats/internal/records/render_test.go`, `stats/internal/store/records.go`
**Tests:** none — repairs existing tests: `TestRenderLedgerDerivesTheTaskFromTheKey` (its comment only); the shape is documentation of what `hooks/enforce-agent-baseline.py` already enforces (decision `implementer-key-ids`).
**Regression:** **none** — no behaviour changes.
**Commit:** docs(flow): give the group implementer key as task-<ids>-implementer
**After:** none

**Decision:** implementer-key-ids

- [x] 5. run-capped case (a) proves the kill by a marker, not a clock
  - [x] **Step 1: Edit** `scripts/test-run-capped.sh` case (a): drop `start`, `el` and the `-le 6` test; run `python3 -c "$alloc; open('$tmp/done-a', 'w').close()"` under `bash "$rc" -m 64 --`, and assert `[ "$s" = 137 ] && [ ! -e "$tmp/done-a" ] && grep -q 'CAPPED:' "$tmp/err"`, failing as `FAIL (a): status=$s completed=<yes|no>`. Rename the case comment to `# (a) capped before completion, CAPPED: line`, and add one comment line: the marker, written only once the allocator finishes, proves the kill without a wall-clock bound that load breaks (kan-1052 self-review, F4).
  - [x] **Step 2: Prove the mechanism.** Copy `scripts/` to a scratch directory outside the worktree. (i) In the copy's `run-capped.sh`, make the cap check never fire; the copy's `test-run-capped.sh` prints `FAIL (a)` naming `completed=yes`. (ii) Restore it, and delay the copy's first RSS sample by 7 s; the new case (a) passes, and the old case (a) from `git show HEAD:scripts/test-run-capped.sh`, run against the same copy, fails on its elapsed bound. Delete the copy.
  - [x] **Step 3: Verify.** `bash -n scripts/test-run-capped.sh` exits 0; `bash scripts/test-run-capped.sh` prints `ok: run-capped.sh`. Commit, the body stating Step 2's two results.

**Build:** green
**Files:** `scripts/test-run-capped.sh`
**Tests:** none — repairs existing tests: `test-run-capped.sh` case (a); Step 2 proves its new assertion.
**Regression:** case (a) fails again under load though the cap killed the child.
**Commit:** test(scripts): prove run-capped's kill by a completion marker, not a clock
**After:** none

**Decision:** run-capped-completion-marker

- [x] 6. The mutation brief carries the node_modules rule
  - [x] **Step 1: Write the test** `TestRenderSlotPromptMutationPassForbidsSharedNodeModules` in `stats/internal/guard/render_slot_prompt_test.go`, on `TestRenderSlotPromptRendersMutationPass`'s fixture: render `mutation` alone from the real skill directory and require the rendered file to contain ``Never symlink a shared `node_modules` into your copy`` and `npm ci`. Run `cd stats && make -s web-dist && go test ./internal/guard -run 'TestRenderSlotPromptMutationPassForbidsSharedNodeModules' -count=1` — it fails.
  - [x] **Step 2: Fix.** Append to the last paragraph of **The mutation-testing brief** in `skills/flow/review-panel.md`: "Never symlink a shared `node_modules` into your copy: run `npm ci` inside the copy, which installs its own — over a symlink it empties the shared directory (kan-1037; kan-1052 self-review, F6)."
  - [x] **Step 3: Verify.** The test and `TestRenderSlotPromptRendersMutationPass` pass; `cd stats && gofmt -l internal/guard` prints nothing; `scripts/run-verify-list.sh <wt> lint` exits 0. Commit.

**Build:** green
**Files:** `skills/flow/review-panel.md`, `stats/internal/guard/render_slot_prompt_test.go`
**Tests:** `TestRenderSlotPromptMutationPassForbidsSharedNodeModules`
**Regression:** `TestRenderSlotPromptMutationPassForbidsSharedNodeModules` fails — the mutation slot's prompt again lacks the rule, and a slot may link the canonical `node_modules` into its copy.
**Baseline:** before=7 after=8 <!-- measured: grep -c -E '^func Test' stats/internal/guard/render_slot_prompt_test.go @ fab4520a -->
**Commit:** docs(flow): tell the mutation slot to npm ci its copy, never symlink node_modules
**After:** none

**Decision:** mutation-brief-node-modules

- [x] 7. A planned test's fixture rules out the nearest wrong version
  - [x] **Step 1: Edit** `skills/spectre-planning/SKILL.md` in `<wt>`: in section **D. Basic Workflow #3 — Writing plans**, after the `>` block **Write a spec so it never waits out a production timing.**, add a blank line and this `>` block, each line prefixed `> `: "**Write each plan-specified test so the nearest wrong version fails it.** When a task's step spells out a test's fixture or its step order, choose them so the rule under test and the nearest plausible wrong version of it give different results: a sort's fixture holds its items in an order its sort key does not; a test that a step keeps earlier state sets that state before the step, never after it. A fixture the wrong version also passes cannot catch the defect it was written for (kan-1052: a fixture whose key order matched its spend order passed a sort on the wrong key, and a test that clicked a Period chip before choosing any other chip passed a click that reset the others; both cost a review fix round)."
  - [x] **Step 2: Verify in agents.** `scripts/run-verify-list.sh <wt> lint` exits 0. Commit in `<wt>`.
  - [x] **Step 3: Re-vendor.** `cp <wt>/skills/spectre-planning/SKILL.md <swt>/.claude/skills/spectre-planning/SKILL.md`; in `<swt>/.claude/skills/spectre-planning/UPSTREAM.md` replace the `- Landed:` line with `- Landed: agents main with kan-1052's self-review fixes, 2026-10-31`. `cmp <wt>/skills/spectre-planning/SKILL.md <swt>/.claude/skills/spectre-planning/SKILL.md` prints nothing.
  - [x] **Step 4: Verify in spectre.** From `<swt>`: `gofmt -l . | grep .` prints nothing, `go vet ./...` and `go test -run 'TestPlugin' ./...` exit 0. Commit in `<swt>` with the same subject.

**Build:** green
**Files:** `skills/spectre-planning/SKILL.md`
**Allowed-collateral:** `.claude/skills/spectre-planning/*` — the spectre commit's two files, `SKILL.md` and `UPSTREAM.md`, in `<swt>`
**Tests:** **none** — no mechanical check can judge whether a fixture discriminates (decision `discriminating-fixture-rule`).
**Regression:** **none** — prose rule; `cmp` keeps the vendored copy identical.
**Commit:** docs(spectre-planning): require a planned test's fixture to fail the nearest wrong version
**After:** none

**Decision:** discriminating-fixture-rule
**Decision:** spectre-worktree-by-parent
