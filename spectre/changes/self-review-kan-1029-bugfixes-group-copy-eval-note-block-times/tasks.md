# self-review-kan-1029-bugfixes-group-copy-eval-note-block-times

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1029's nine self-review findings — F1 (Task 1), F6 (Task 2), F3 (Task 3),
F4 (Task 4), F5 (Task 5), F9 (Task 6), F2 (Task 7), F7 (Task 8), F8 (Task 9), as `proposal.md`
names them.

**Architecture:** two refusals in `flow stage` (the live check on `flow.verify`'s end, the token's
UTC stamp) plus a `flowd retoken-stage-run` repair command; the mutation pass rendered; the hook's
key shapes closed; the harvester's loop expansion taught `#`/`%` strips; the zip rebuild's root
checked; two gymie dev-tooling edits; one hazard row; three live store repairs.

**Tech Stack:** Go (`stats/`), Python (`hooks/`), Bash, Markdown, Gradle Kotlin DSL (gymie).

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-24-self-review-kan-1029-bugfixes-group-copy-eval-note-block-times-design.md`
in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer, and no `Co-Authored-By:` trailer. Never write `spectre/changes/` — this change's `patches/` and `verbatim-moves.txt` are already committed with the plan. Never merge or push.
- Tasks 1–6 and 9 run in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1029-bugfixes-group-copy-eval-note-block-times` (`<agents-wt>` below), branch `self-review-kan-1029-bugfixes-group-copy-eval-note-block-times`. Task 7 runs in `/Users/tweety53/Projects/gymie-frontend-worktrees/kan-1029-bugfixes-group-copy-eval-note-block-times` and Task 8 in `/Users/tweety53/Projects/gymie-worktrees/kan-1029-bugfixes-group-copy-eval-note-block-times`, each on branch `spectre/kan-1029-bugfixes-group-copy-eval-note-block-times`, committing after its archive commit and rewriting none. Every `**Files:**` path is relative to its task's worktree root. Never touch a main checkout, and never create a worktree.
- A task's diff is `<agents-wt>/spectre/changes/self-review-kan-1029-bugfixes-group-copy-eval-note-block-times/patches/task-<n>.diff`, applied with `git apply <that path>` from the task's worktree root, in task order. Tasks 1–6 were applied in that order to `<agents-wt>` at `a988a619` and rebuilt the tested tree exactly; `git apply --check` passed for task-7 on the frontend at `a66d4aeb` and task-8 on the backend at `02afc59b`. A patch that does not apply is reported, never hand-merged by guesswork.
- Prose lands character for character as its patch carries it: `scripts/check-verbatim-moves.sh` matches every added and removed sentence against this change's `verbatim-moves.txt`; with Tasks 1–6 applied, `check-verbatim-moves.sh`, `check-markdown-integrity.sh`, `check-references.sh`, `check-installed-citations.sh`, `check-normative-inventory.sh`, `check-dispatch-paragraphs.sh`, `check-stage-mark-calls.sh`, `check-guard-symlinks.sh` and `check-python-suppressions.sh` exited 0, and `cd stats && make -s web-dist && go test ./... -count=1` and `scripts/run-guard-tests.sh` passed.
- Go: run `gofmt -w` on every Go file you change before committing; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Mutation proofs restore with `break-and-prove.sh` or `mutate-and-verify.sh`, or a saved diff re-applied with `git apply` — never `git checkout --` on a file carrying uncommitted edits.
- The gymie commits (Tasks 7 and 8) are commits outside every task of kan-1029's change (**Commits outside every task**, `skills/flow-contracts/git-boundaries.md`): add `-m "Flow-Key: self-review-kan-1029-bugfixes-group-copy-eval-note-block-times-fix"`; never touch either repository's `spectre/`. Run gymie's `adb` from `~` only (`(cd ~ && adb start-server)`), never from a worktree.
- The store repairs (Tasks 2 and 5) and the hazard row (Task 9) write the live store through `flowd` subcommands and `flow hazard add`, which open the database directly; never start, stop or restart the protected `flowd` on `127.0.0.1:4173` nor the `flow-postgres` container, and never run tests against the live database.

## Review Focus

- Task 1: a `flow.verify` end with `-outcome` other than `completed`, a `/flow-fast` `flow.verify` end and every other stage pass through untouched.
- Task 2: a stamped token whose name part carries dashes (`mf-k1029-x-20261024T002513Z`) is accepted; a token with the stamp mid-string is refused.
- Task 3: a delta or fix-round render with mutation still names each other pass's diff file; a mutation-alone render carries no ENTRY CONTEXT and no touched list.
- Task 4: every key shape the skills prescribe — `task-4+5+6-reviewer-fix-1`, `panel-2-primary+principles-cite`, `panel-fix-3-retry`, `visual-verify-2-gymie`, `tooling-analysis-1` — still passes.
- Task 6: a root given with a trailing slash still accepts the PNGs under it.

- [ ] 1. Gate flow.verify's stage end on the live check
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `<agents-wt>`, `grep -c 'check-live-verification' stats/cmd/flow/stage.go` prints `0`, and `docker exec flow-postgres psql -U flow -d flow -tc "select outcome from stage_runs where id = 17075"` prints `completed` — kan-1029's `flow.verify` ended with no gate.
    <!-- measured: grep -c 'check-live-verification' stats/cmd/flow/stage.go @ a988a619 → 0; stage_runs 17075 → completed -->
  - [ ] **Step 2: Apply `patches/task-1.diff`** from `<agents-wt>`: `gateLiveVerification` and its call in `runStageEnd` (`stats/cmd/flow/stage.go`), `TestStageEndFlowVerifyRunsLiveCheck` (`stats/cmd/flow/stage_test.go`), the **Close the stage** paragraph and its one-call block (`skills/flow/verify-and-handoff.md`), and the header pointer in `scripts/check-live-verification.sh`.
    <!-- measured: git apply --index patches/task-1.diff @ a988a619 → exit 0 -->
  - [ ] **Step 3: Prove the test pins the fix.** With `break-and-prove.sh` (or a saved diff re-applied), replace `if code := gateLiveVerification(f, *outcome, stderr); code != 0 {` with `if code := 0; code != 0 {` — `cd stats && go test ./cmd/flow -run 'TestStageEndFlowVerifyRunsLiveCheck' -count=1` fails two subtests; restore.
    <!-- measured: that flip @ a988a619 with the patch → 2 subtests FAIL; restored → ok -->
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l cmd/flow` prints nothing; `cd stats && go test ./cmd/flow -run 'TestStageEnd' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-references.sh`, `scripts/check-installed-citations.sh` and `scripts/check-stage-mark-calls.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/cmd/flow/stage.go`, `stats/cmd/flow/stage_test.go`, `skills/flow/verify-and-handoff.md`, `scripts/check-live-verification.sh`
**Tests:** `TestStageEndFlowVerifyRunsLiveCheck`
**Regression:** `TestStageEndFlowVerifyRunsLiveCheck` fails — a completed `flow.verify` end with a named live check and no record reaches the store again.
**Baseline:** before=36 after=37 <!-- measured: grep -c -E '^func Test' stats/cmd/flow/stage_test.go @ a988a619 → 36, and 37 with the patch -->
**Commit:** fix(stats): run the live-verification check in flow.verify's stage end
**After:** none

**Decision:** verify-end-runs-live-check

- [ ] 2. Refuse a stage-mark token without its UTC stamp, and retoken rows 17078 and 17079
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `docker exec flow-postgres psql -U flow -d flow -tc "select id, session_token from stage_runs where id in (17078,17079) order by id"` prints `mf-k1029-PLACEHOLDER` for both, and `grep -c 'retoken-stage-run' stats/cmd/flowd/main.go` prints `0`.
    <!-- measured: that query @ dev store 2026-10-24 → 17078, 17079 mf-k1029-PLACEHOLDER; grep @ a988a619 → 0 -->
  - [ ] **Step 2: Apply `patches/task-2.diff`** from `<agents-wt>`, after Task 1's commit: `validateSessionTokenStamp` and its calls in `runStageBegin` and `runStageWrap` plus the usage line (`stats/cmd/flow/stage.go`); every unstamped token literal in `stats/cmd/flow/stage_test.go` given the suffix `-20261024T000000Z`, and `TestStageBeginRefusesUnstampedSessionToken`; `store.RetokenStageRun` (`stats/internal/store/stageruns.go`) and `TestRetokenStageRunSetsTheToken` (`stats/internal/store/stageruns_test.go`); the `retoken-stage-run` subcommand (`stats/cmd/flowd/main.go`); the refusal sentence in **Stage marks** (`skills/flow-contracts/pipeline.md`).
    <!-- measured: git apply --index patches/task-2.diff @ a988a619 after task-1.diff → exit 0 -->
  - [ ] **Step 3: Prove the tests pin the fix.** With `break-and-prove.sh` (or a saved diff re-applied): make `validateSessionTokenStamp` `return nil` first — `cd stats && go test ./cmd/flow -run 'TestStageBeginRefusesUnstampedSessionToken' -count=1` fails; restore. Change `RetokenStageRun`'s SQL to ``UPDATE stage_runs SET session_token = session_token WHERE id = $1 AND $2::text = $2::text`` — `cd stats && go test ./internal/store -run 'TestRetokenStageRunSetsTheToken' -count=1` fails; restore.
    <!-- measured: both flips @ a988a619 with tasks 1-6 → the refusal test fails its begin subtests; the retoken test fails -->
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l .` prints nothing; `cd stats && go test ./cmd/flow -count=1` passes; `cd stats && go test ./internal/store -run 'TestRetokenStageRunSetsTheToken' -count=1` passes; `cd stats && make -s web-dist && go vet ./...` exits 0; `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-references.sh` and `scripts/check-stage-mark-calls.sh` exit 0. Commit.
    <!-- measured: go test ./cmd/flow -count=1 and the store test @ a988a619 with task-1 and task-2 → ok -->
  - [ ] **Step 5: Repair the two rows** from `<agents-wt>/stats`, after the commit: `make -s web-dist`, then `go run ./cmd/flowd retoken-stage-run 17078 mf-k1029-20261024T002513Z` and `go run ./cmd/flowd retoken-stage-run 17079 mf-k1029-20261024T002513Z` (the default DSN is `flow-postgres`'s, `postgres://flow:flow@localhost:5433/flow?sslmode=disable`).
  - [ ] **Step 6: Verify the repair.** Step 1's query prints `mf-k1029-20261024T002513Z` for both rows; report its output.

**Build:** green
**Files:** `stats/cmd/flow/stage.go`, `stats/cmd/flow/stage_test.go`, `stats/cmd/flowd/main.go`, `stats/internal/store/stageruns.go`, `stats/internal/store/stageruns_test.go`, `skills/flow-contracts/pipeline.md`
**Tests:** `TestStageBeginRefusesUnstampedSessionToken`, `TestRetokenStageRunSetsTheToken`
**Regression:** `TestStageBeginRefusesUnstampedSessionToken` fails — a placeholder token opens stage runs again; `TestRetokenStageRunSetsTheToken` fails — `flowd retoken-stage-run` stops changing the row.
**Baseline:** before=92 after=94 <!-- measured: cat stats/cmd/flow/stage_test.go stats/internal/store/stageruns_test.go | grep -c -E '^func Test' @ a988a619 → 91, 92 after task-1, and 94 with task-2 -->
**Commit:** fix(stats): refuse a stage-mark token without its UTC stamp
**After:** Task 1

**Decision:** stage-token-utc-stamp-enforced

**Decision:** k1029-stage-rows-retoken

- [ ] 3. Render the mutation pass from the mutation-testing brief
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n "mutation's pass is never rendered" skills/flow/review-panel.md scripts/render-slot-prompt.sh` prints one line in each, and `renderslotprompt.go`'s slot loop drops `mutation` with an empty `case`.
    <!-- measured: that grep @ a988a619 → review-panel.md:370, render-slot-prompt.sh:42 -->
  - [ ] **Step 2: Apply `patches/task-3.diff`** from `<agents-wt>`: the `## PASS mutation` render and `rspSection` (`stats/internal/guard/renderslotprompt.go`); `TestRenderSlotPromptRendersMutationPass`, carrying the mutation cases out of `TestRenderSlotPromptUnresolvedPlaceholder` with their assertions inverted (`stats/internal/guard/render_slot_prompt_test.go`); the header paragraph (`scripts/render-slot-prompt.sh`); the rendered-mutation sentence (`skills/flow/review-panel.md`).
    <!-- measured: git apply --index patches/task-3.diff @ a988a619 after tasks 1-2 → exit 0 -->
  - [ ] **Step 3: Prove the test pins the fix.** With `break-and-prove.sh` (or a saved diff re-applied), insert `continue` as the first statement of the `if id == "mutation" {` block in `renderslotprompt.go` — `cd stats && go test ./internal/guard -run 'TestRenderSlotPromptRendersMutationPass' -count=1` fails; restore.
    <!-- measured: that flip @ a988a619 with tasks 1-6 → TestRenderSlotPromptRendersMutationPass FAIL -->
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestRenderSlotPrompt' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-references.sh`, `scripts/check-installed-citations.sh` and `scripts/check-dispatch-paragraphs.sh` exit 0. Commit.
    <!-- measured: those commands @ a988a619 with tasks 1-6 → ok -->

**Build:** green
**Files:** `stats/internal/guard/renderslotprompt.go`, `stats/internal/guard/render_slot_prompt_test.go`, `scripts/render-slot-prompt.sh`, `skills/flow/review-panel.md`
**Tests:** `TestRenderSlotPromptRendersMutationPass`
**Regression:** `TestRenderSlotPromptRendersMutationPass` fails — a bundle naming mutation renders no mutation pass, and the slot runs only its other roles.
**Baseline:** before=5 after=6 <!-- measured: grep -c -E '^func Test' stats/internal/guard/render_slot_prompt_test.go @ a988a619 → 5, and 6 with the patch -->
**Commit:** fix(guard): render the mutation pass from the mutation-testing brief
**After:** none

**Decision:** render-mutation-pass

- [ ] 4. Accept only the closed list's key shapes as a dispatch description
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Pipe `{"tool_name":"Agent","transcript_path":"<a file holding {"type":"user","message":{"content":"<command-name>/flow</command-name>"}}>","tool_input":{"subagent_type":"flow-low","model":"opus","description":"verify-fix-1 m3 m4 departures","prompt":"read agent-baseline.md … verbatim"}}` into `python3 hooks/enforce-agent-baseline.py`: it prints nothing (allowed).
    <!-- measured: ROW_DESCRIPTION @ a988a619 matches `[a-z][a-z0-9+._-]*(?:\s|$)`, which `verify-fix-1 ` satisfies -->
  - [ ] **Step 2: Apply `patches/task-4.diff`** from `<agents-wt>`: `ROW_KEY` and the narrowed `ROW_DESCRIPTION` plus the docstring line (`hooks/enforce-agent-baseline.py`); cases 13i–13l (`stats/internal/repotest/enforce_agent_baseline_hook_test.go`); the key-shape sentence (`skills/flow/implement.md`).
    <!-- measured: git apply --index patches/task-4.diff @ a988a619 after tasks 1-3 → exit 0 -->
  - [ ] **Step 3: Prove the cases pin the fix.** With `break-and-prove.sh` (or a saved diff re-applied), set `ROW_DESCRIPTION` back to `re.compile(r"Tasks? \d+(?:\+\d+)*/\d+(?:\s|$)|[a-z][a-z0-9+._-]*(?:\s|$)")` — `cd stats && go test ./internal/repotest -run 'TestEnforceAgentBaselineHook' -count=1` fails cases 13i and 13j; restore.
    <!-- measured: that flip @ a988a619 with tasks 1-6 → cases 13i and 13j FAIL -->
  - [ ] **Step 4: Verify.** `python3 -m py_compile hooks/enforce-agent-baseline.py` exits 0; `cd stats && go test ./internal/repotest -run 'TestEnforceAgentBaselineHook' -count=1` passes; `scripts/check-python-suppressions.sh`, `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-references.sh` exit 0. Commit.
    <!-- measured: those commands @ a988a619 with tasks 1-6 → ok -->

**Build:** green
**Files:** `hooks/enforce-agent-baseline.py`, `stats/internal/repotest/enforce_agent_baseline_hook_test.go`, `skills/flow/implement.md`
**Tests:** `13i a /flow dispatch whose first word is off the closed list`, `13j a /flow dispatch opening with an inline key`, `13k a /flow verifier opening with its repository-suffixed key`, `13l a /flow panel-fix chunk`
**Regression:** cases 13i and 13j fail — a launch described by a word off the closed list goes out with no dispatches row again; 13k and 13l fail — a prescribed verifier or panel-fix key is refused.
**Baseline:** before=30 after=34 <!-- measured: grep -c -E '^[[:space:]]+\{"[0-9]+[a-z]? ' stats/internal/repotest/enforce_agent_baseline_hook_test.go @ a988a619 → 30, and 34 with the patch -->
**Commit:** fix(hooks): accept only the closed list's key shapes as a dispatch description
**After:** none

**Decision:** hook-closed-list-keys

- [ ] 5. Read strips of a loop variable in a dispatch begin, and restamp rows 7367 and 7368
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `docker exec flow-postgres psql -U flow -d flow -tc "select id, dispatch_key, coalesce(agent_id,'') from dispatches where id in (7367,7368) order by id"` prints both rows with an empty agent id; the kan-1029 transcript `/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/63fb1038-b4d0-4297-93cd-31d4590242ef.jsonl` holds the begin `for k in "9:task-9-implementer" "11:task-11-implementer"; do flow record dispatch begin … -task ${k%%:*} … -key ${k#*:} …; done`.
    <!-- measured: that query @ dev store 2026-10-24 → 7367 task-9-implementer '', 7368 task-11-implementer ''; the grep of the transcript → the loop above -->
  - [ ] **Step 2: Apply `patches/task-5.diff`** from `<agents-wt>`, after Task 4's commit: the strip branch in `substituteLoopVar` and `stripLoopWord` (`stats/internal/harvest/watcher.go`); the `prefix and suffix strips of the loop variable resolve the key` case in `TestWatcherStampsKan925LoopBegins` (`stats/internal/harvest/watcher_test.go`); the loop-variable sentence (`skills/flow/implement.md`).
    <!-- measured: git apply --index patches/task-5.diff @ a988a619 after tasks 1-4 → exit 0 -->
  - [ ] **Step 3: Prove the case pins the fix.** With `break-and-prove.sh` (or a saved diff re-applied), make `stripLoopWord` `return word` first — `cd stats && go test ./internal/harvest -run 'TestWatcherStampsKan925LoopBegins' -count=1` fails the new case; restore.
    <!-- measured: that flip @ a988a619 with tasks 1-6 → the strips case FAILs -->
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/harvest` prints nothing; `cd stats && go test ./internal/harvest -run 'TestWatcherStampsKan925LoopBegins' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-references.sh` exit 0. Commit.
    <!-- measured: those commands @ a988a619 with tasks 1-6 → ok -->
  - [ ] **Step 5: Repair the two rows** from `<agents-wt>/stats`, after the commit: `make -s web-dist`, then `go run ./cmd/flowd restamp-dispatch 7367 /Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/63fb1038-b4d0-4297-93cd-31d4590242ef/subagents/agent-a6a4cc4cd5fa75a38.jsonl` and `go run ./cmd/flowd restamp-dispatch 7368 /Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/63fb1038-b4d0-4297-93cd-31d4590242ef/subagents/agent-ac28906c875a56d4f.jsonl`.
    <!-- measured: both agent files exist; their .meta.json descriptions are `Tasks 9+10/11 (Save focus, repeat both)` and `Task 11/11 (web-edge tests)`; first lines 2026-10-23T20:20:58.975Z and .988Z; no dispatches row holds either id -->
  - [ ] **Step 6: Verify the repair.** Step 1's query prints `a6a4cc4cd5fa75a38` for 7367 and `ac28906c875a56d4f` for 7368; report its output.

**Build:** green
**Files:** `stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go`, `skills/flow/implement.md`
**Tests:** `prefix and suffix strips of the loop variable resolve the key`
**Regression:** the case fails — a begin keyed `${k#*:}` inside a loop claims no launch, and its row stays unattributed again.
**Baseline:** before=13 after=14 <!-- measured: grep -c -E '^[[:space:]]+name: "' stats/internal/harvest/watcher_test.go @ a988a619 → 13, and 14 with the patch -->
**Commit:** fix(harvest): read prefix and suffix strips of a loop variable in a dispatch begin
**After:** Task 4

**Decision:** harvest-loop-var-strips

**Decision:** k1029-dispatch-rows-restamp

- [ ] 6. Refuse a screenshots root no resolved PNG lies under
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `ls -la /Users/tweety53/Projects/gymie/full-app-suite.zip` shows the stray zip, and `rebuildFullAppSuiteZip` (`stats/internal/guard/rebuildfullappsuitezip.go`) compares no resolved PNG with `<screenshots root>` before zipping.
    <!-- measured: ls @ 2026-10-24 → the file exists; the function @ a988a619 checks only that the root is absolute -->
  - [ ] **Step 2: Apply `patches/task-6.diff`** from `<agents-wt>`: the root check (`stats/internal/guard/rebuildfullappsuitezip.go`), `case 8 a root no PNG lies under` (`stats/internal/guard/rebuild_full_app_suite_zip_test.go`) and the exit-2 line of both headers (`scripts/rebuild-full-app-suite-zip.sh` and the Go file).
    <!-- measured: git apply --index patches/task-6.diff @ a988a619 after tasks 1-5 → exit 0 -->
  - [ ] **Step 3: Prove the case pins the fix.** With `break-and-prove.sh` (or a saved diff re-applied), change `if rc == 0 {` above the root loop to `if false {` — `cd stats && go test ./internal/guard -run 'TestRebuildFullAppSuiteZip' -count=1` fails case 8; restore.
    <!-- measured: that flip @ a988a619 with tasks 1-6 → case 8 FAILs -->
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestRebuildFullAppSuiteZip' -count=1` passes; `scripts/check-guard-symlinks.sh` exits 0. Commit.
    <!-- measured: those commands @ a988a619 with tasks 1-6 → ok -->

**Build:** green
**Files:** `stats/internal/guard/rebuildfullappsuitezip.go`, `stats/internal/guard/rebuild_full_app_suite_zip_test.go`, `scripts/rebuild-full-app-suite-zip.sh`
**Tests:** `case 8 a root no PNG lies under`
**Regression:** case 8 fails — a root in another checkout gets a stray `full-app-suite.zip` again.
**Baseline:** before=7 after=8 <!-- measured: grep -c -E 't\.Run\("case [0-9]+' stats/internal/guard/rebuild_full_app_suite_zip_test.go @ a988a619 → 7, and 8 with the patch -->
**Commit:** fix(guard): refuse a screenshots root no resolved PNG lies under
**After:** none

**Decision:** zip-root-holds-the-pngs

- [ ] 7. Drop the iOS simulator test run from gymie-frontend's suite
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/gymie-frontend-worktrees/kan-1029-bugfixes-group-copy-eval-note-block-times`, the `## test` fence line of `.flow/project.md` run with `--dry-run` appended, piped through `grep -E 'linkDebugTestIosSimulatorArm64|iosSimulatorArm64Test '`, lists both tasks.
    <!-- measured: that dry run @ a66d4aeb → :shared:linkDebugTestIosSimulatorArm64 SKIPPED, :shared:iosSimulatorArm64Test SKIPPED -->
  - [ ] **Step 2: Apply `patches/task-7.diff`** from that worktree's root (`git apply <agents-wt>/spectre/changes/self-review-kan-1029-bugfixes-group-copy-eval-note-block-times/patches/task-7.diff`): the `## test` line and its one-paragraph reason (`.flow/project.md`).
    <!-- measured: git apply --check patches/task-7.diff @ a66d4aeb → exit 0 -->
  - [ ] **Step 3: Verify.** The new `## test` fence line run with `--dry-run` appended, piped through `grep -E 'IosSimulatorArm64'`, lists `:shared:compileTestKotlinIosSimulatorArm64` and neither `:shared:linkDebugTestIosSimulatorArm64` nor `:shared:iosSimulatorArm64Test`. Commit with `-m "Flow-Key: self-review-kan-1029-bugfixes-group-copy-eval-note-block-times-fix"`.
    <!-- measured: that dry run @ a66d4aeb → compileTestKotlinIosSimulatorArm64 listed; link and test tasks absent -->

**Build:** green
**Files:** `.flow/project.md`
**Tests:** **none** — a configuration line; Step 3's dry run shows the graph.
**Regression:** none — no test is declared.
**Commit:** build(flow): keep the iOS test compile and drop its Xcode-bound run from the suite
**After:** none

**Decision:** frontend-suite-drops-ios-link

- [ ] 8. Reverse the workspace's resolved ports in devPhone
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/gymie-worktrees/kan-1029-bugfixes-group-copy-eval-note-block-times`, `grep -n 'listOf(3000, 8080, 9000)' gradle/dev-lifecycle.gradle.kts` prints the `devPhone` line.
    <!-- measured: that grep @ 02afc59b → line 2609 -->
  - [ ] **Step 2: Apply `patches/task-8.diff`** from that worktree's root (`git apply <agents-wt>/spectre/changes/self-review-kan-1029-bugfixes-group-copy-eval-note-block-times/patches/task-8.diff`): the `devPhone` task and its comment (`gradle/dev-lifecycle.gradle.kts`).
    <!-- measured: git apply --check patches/task-8.diff @ 02afc59b → exit 0 -->
  - [ ] **Step 3: Verify.** `./gradlew ktlintFormat`, then `./gradlew verifyChange` exits 0; `./gradlew tasks --group 'gymie dev' | grep devPhone` prints `devPhone - Reverse this workspace's frontend and gateway ports and 9000 from the USB-connected Android phone`. Commit with `-m "Flow-Key: self-review-kan-1029-bugfixes-group-copy-eval-note-block-times-fix"`.
    <!-- measured: ktlintFormat, verifyChange and the tasks listing @ 02afc59b with the patch → clean; the line above -->

**Build:** green
**Files:** `gradle/dev-lifecycle.gradle.kts`
**Tests:** **none** — a Gradle task body; no test harness reaches `dev-lifecycle.gradle.kts`'s tasks.
**Regression:** none — no test is declared.
**Commit:** fix(dev-lifecycle): reverse the workspace's resolved ports in devPhone
**After:** none

**Decision:** devphone-resolved-ports

- [ ] 9. Add gymie's hazard for `./gradlew --stop`
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `flow hazards -C /Users/tweety53/Projects/gymie` prints `[]`.
    <!-- measured: flow hazards -C /Users/tweety53/Projects/gymie @ dev store 2026-10-24 → [] -->
  - [ ] **Step 2: Add the row:** `flow hazard add -C /Users/tweety53/Projects/gymie -name gradle-stop-kills-dev-stack -applies all -text '`./gradlew --stop` stops every Gradle daemon of the wrapper'"'"'s version on this machine, the dev stack'"'"'s frontend among them: devStart runs it as `./gradlew :webApp:wasmJsBrowserDevelopmentRun`, so --stop takes the frontend down mid-check (kan-1029: the live-check verifier killed it that way). Never run --stop while a dev stack is up; stop one build with Ctrl-C, and the stack with devStop.'`
  - [ ] **Step 3: Verify.** Step 1's command prints one row named `gradle-stop-kills-dev-stack` with `applies` `all`; report it.

**Build:** green
**Files:** none — a store write, no file changes
**Tests:** none — verified by Step 3's read
**Regression:** none — no code changes
**Commit:** none — this task commits nothing; it adds one flow-store row
**After:** none

**Decision:** gymie-gradle-stop-hazard
