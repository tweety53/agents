# self-review-kan-1023-templates-step-2-fe-save-as-template-template

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1023's self-review findings — F1 (Task 1), F2 (Task 2), F4 (Tasks 3 and 4),
as `proposal.md` names them; F3 is disproved in `design.md` and plans no task.

**Architecture:** one regex class narrowed in `check-unfinished-work`; one sentence in the
verify-fix loop; the harvester's keyed stamp taught the `Task <ids>/<n>` description; a settle pass
for closed agentless and inline rows; `flowd restamp-dispatch` widened to repair them; eight live
store rows repaired.

**Tech Stack:** Go (`stats/`), Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-25-self-review-kan-1023-templates-step-2-fe-save-as-template-template-design.md`
in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer, and no `Co-Authored-By:` trailer. Never write `spectre/changes/` — this change's `patches/` and `verbatim-moves.txt` are already committed with the plan. Never merge or push.
- Every task runs in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1023-templates-step-2-fe-save-as-template-template` (`<agents-wt>` below), branch `self-review-kan-1023-templates-step-2-fe-save-as-template-template`. Every `**Files:**` path is relative to it. Never touch a main checkout, and never create a worktree.
- Tasks 1 and 2 apply `<agents-wt>/spectre/changes/self-review-kan-1023-templates-step-2-fe-save-as-template-template/patches/task-<n>.diff` with `git apply <that path>` from `<agents-wt>`; `git apply --check` passed for both at `ef04314f`. A patch that does not apply is reported, never hand-merged by guesswork. Tasks 3 and 4 are written from their steps.
- Prose lands character for character as its patch carries it: `scripts/check-verbatim-moves.sh` matches every added sentence against this change's `verbatim-moves.txt`.
- Go: run `gofmt -w` on every Go file you change before committing; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0. Every Go test command below is run from `<agents-wt>/stats`.
- Mutation proofs restore with `break-and-prove.sh` or `mutate-and-verify.sh`, or a saved copy restored with `cp` — never `git checkout --` on a file carrying uncommitted edits.
- The store repairs (Tasks 3 and 4) write the live store through `go run ./cmd/flowd restamp-dispatch`, which opens the database directly (default DSN `postgres://flow:flow@localhost:5433/flow?sslmode=disable`); never start, stop or restart the protected `flowd` on `127.0.0.1:4173` nor the `flow-postgres` container, and never run tests against the live database.
- The gymie run's parent transcript is `/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/e905b265-37ae-40b2-9776-72e831e83ebc.jsonl` (`<parent>` below); its agent transcripts sit in `…/e905b265-37ae-40b2-9776-72e831e83ebc/subagents/`.

## Review Focus

- Task 1: a spec path with a directory, a dash and a dot (`fidelity/full-app-suite.spec.ts:110`) still captures whole.
- Task 3: `Tasks 2+3+5+8/12 (review)` resolves to `task-2+3+5+8-`; a prefix `task-13-` never stamps `task-130-implementer`; a description opening with a key is unchanged.
- Task 4: a closed agentless row with an Agent launch or denial inside its interval is never stamped inline; an open row, or one ended after the harvested records, is kept for a later batch; a zero share writes a `tokens` key.

- [ ] 1. Exclude opening punctuation from the verify-failure spec capture
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `<agents-wt>`, `sed -n 1070p stats/internal/guard/unfinishedwork.go` prints ``uwSpecRE      = regexp.MustCompile(`([^\s:]+\.spec\.ts):(\d+)`)`` — a class admitting `(`, `[`, backtick and quotes.
    <!-- measured: sed -n 1070p stats/internal/guard/unfinishedwork.go @ ef04314f → the line above; a scratch go run of that pattern over "(kan-341-route-b.spec.ts:224:5) and `fidelity/a.spec.ts:3`" → "(kan-341-route-b.spec.ts", "`fidelity/a.spec.ts" -->
  - [ ] **Step 2: Apply `patches/task-1.diff`** from `<agents-wt>`: the narrowed `uwSpecRE` (`stats/internal/guard/unfinishedwork.go`) and `TestUWVerifyFailuresSpecExcludesOpeningPunctuation` plus the `slices` import (`stats/internal/guard/check_unfinished_work_test.go`).
    <!-- measured: git apply --check patches/task-1.diff @ ef04314f → exit 0 -->
  - [ ] **Step 3: Prove the test pins the fix.** Save a copy of `unfinishedwork.go`, set `uwSpecRE` back to ``regexp.MustCompile(`([^\s:]+\.spec\.ts):(\d+)`)`` — `go test ./internal/guard -run 'TestUWVerifyFailuresSpecExcludesOpeningPunctuation' -count=1` fails; restore the copy.
    <!-- measured: that test @ ef04314f without the regex change → FAIL; with it → ok -->
  - [ ] **Step 4: Verify.** `gofmt -l internal/guard` prints nothing; `go test ./internal/guard -run 'TestUWVerifyFailures|TestCheckUnfinishedWork' -count=1` passes. Commit.
    <!-- measured: go test ./internal/guard -run 'TestUWVerifyFailures|TestCheckUnfinishedWork' -count=1 @ ef04314f with the patch → ok -->

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`
**Tests:** `TestUWVerifyFailuresSpecExcludesOpeningPunctuation`
**Regression:** `TestUWVerifyFailuresSpecExcludesOpeningPunctuation` fails — a parenthesised or quoted Playwright location is keyed with its stray character again, and a clean citation never clears it.
**Baseline:** before=9 after=10 <!-- measured: grep -c -E '^func Test' stats/internal/guard/check_unfinished_work_test.go @ ef04314f → 9, and 10 with the patch -->
**Commit:** fix(guard): exclude opening punctuation from the verify-failure spec capture
**After:** none

**Decision:** spec-regex-excludes-opening-punctuation

- [ ] 2. Require a test for every new branch a visual fix adds
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -c 'Every new branch the fix adds' skills/flow/verify-fix-loop.md` prints `0`; step 2 of **The loop** asks for the departure's permanent probe and nothing for a branch beyond it.
    <!-- measured: that grep @ ef04314f → 0 -->
  - [ ] **Step 2: Apply `patches/task-2.diff`** from `<agents-wt>`: the **Every new branch the fix adds lands a test of its own in the same commit.** sentence after the permanent-probe paragraph (`skills/flow/verify-fix-loop.md`).
    <!-- measured: git apply --check patches/task-2.diff @ ef04314f → exit 0 -->
  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-references.sh`, `scripts/check-installed-citations.sh` and `scripts/check-normative-inventory.sh` exit 0. Commit.
    <!-- measured: those five @ ef04314f with the patch applied → check-verbatim-moves exit 1 on exactly the two sentences verbatim-moves.txt lists, the other four exit 0 -->

**Build:** green
**Files:** `skills/flow/verify-fix-loop.md`
**Tests:** none — a prose rule; `check-verbatim-moves.sh` pins its text.
**Regression:** none — no test; reverting the commit drops the rule and `check-verbatim-moves.sh` stays green.
**Commit:** fix(flow): require a test for every new branch a visual fix adds
**After:** none

**Decision:** visual-fix-tests-every-new-branch

- [ ] 3. Stamp a launch described by its task ids onto its task row, and restamp row 7560
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `sed -n 590p stats/internal/harvest/transcript.go` prints `Key:       firstWord(raw.ToolUseResult.Description),`; `docker exec flow-postgres psql -U flow -d flow -tc "select id, dispatch_key, coalesce(agent_id,''), metrics ? 'tokens' from dispatches where id = 7560"` prints `7560 | task-13-implementer |  | f`; line 1868 of `<parent>` is the launch described `Task 13/13 (visual-verify fixes)` whose result names agent `afdfbfaf17eebe758`.
    <!-- measured: those three reads @ ef04314f, dev store 2026-10-25 → as stated; no dispatches row of the change holds afdfbfaf17eebe758 -->
  - [ ] **Step 2: Write the failing tests.** In `stats/internal/store/records_test.go`, `TestStampSessionDispatchAgentStampsATaskPrefix`, built like `TestStampSessionDispatchAgentFillsTheSessionsLatestEmptyRow`: in one session seed agentless rows `task-13-implementer` (ended), `task-130-implementer` (open) and `task-13-reviewer` (open, newest); `StampSessionDispatchAgent(ctx, session, "task-13-", "a1")` returns true and stamps `task-13-reviewer` alone; with `task-13-reviewer` stamped, a second call returns false (the ended `task-13-implementer` and `task-130-implementer` stay empty). In `stats/internal/harvest/watcher_test.go`, `TestWatcherStampsATaskDescribedLaunchByItsPrefix`, built like `TestWatcherStampsALaunchByItsDescriptionKey`: a transcript with a lone begin `-key task-13-implementer`, two denied Agent calls 32 s and 59 s after it, then a launch described `Task 13/13 (visual-verify fixes)` 85 s after it — the fake records `StampSessionDispatchAgent` with key `task-13-` and the launch's agent id, and no `StampDispatchAgent` call; a second case described `Tasks 2+3+5+8/12 (review)` records key `task-2+3+5+8-`. Run both; they fail.
  - [ ] **Step 3: Implement.** (a) `stats/internal/harvest/transcript.go`: `AgentLaunch.Key` comes from a new `launchKey(description)` — a description matching `^Tasks? (\d+(?:\+\d+)*)/\d+(?:\s|$)` gives `task-<ids>-`, anything else its first word as today; update `Key`'s doc comment to say so. (b) `stats/internal/store/records.go`, `StampSessionDispatchAgent`: the row filter reads `(d.dispatch_key = $2 OR (right($2, 1) = '-' AND starts_with(d.dispatch_key, $2) AND d.ended_at IS NULL))`, its doc comment naming the prefix form and why it is open rows only (the hook's own open-row rule; a closed agentless row is never a launch's). (c) `stats/internal/harvest/watcher.go`, `stampKeyedLaunches`: a begin is dropped when its key is used or starts with a used key ending in `-`.
  - [ ] **Step 4: Prove the tests pin the fix.** Save copies first. Make `launchKey` return `firstWord(description)` — `go test ./internal/harvest -run 'TestWatcherStampsATaskDescribedLaunchByItsPrefix' -count=1` fails; restore. Drop the `OR (…)` clause from the SQL — `go test ./internal/store -run 'TestStampSessionDispatchAgentStampsATaskPrefix' -count=1` fails; restore.
  - [ ] **Step 5: Verify.** `gofmt -l internal` prints nothing; `go test ./internal/harvest ./internal/store -run 'TestStampSessionDispatchAgent|TestWatcherStamps' -count=1` passes; `make -s web-dist && go vet ./...` exits 0. Commit.
  - [ ] **Step 6: Repair row 7560**, after the commit: `make -s web-dist`, then `go run ./cmd/flowd restamp-dispatch 7560 /Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/e905b265-37ae-40b2-9776-72e831e83ebc/subagents/agent-afdfbfaf17eebe758.jsonl`.
    <!-- measured: that agent file and its .meta.json (description `Task 13/13 (visual-verify fixes)`) exist @ 2026-10-25 -->
  - [ ] **Step 7: Verify the repair.** Step 1's query prints `7560 | task-13-implementer | afdfbfaf17eebe758 | t`; report its output.

**Build:** green
**Files:** `stats/internal/harvest/transcript.go`, `stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`
**Tests:** `TestStampSessionDispatchAgentStampsATaskPrefix`, `TestWatcherStampsATaskDescribedLaunchByItsPrefix`
**Regression:** `TestWatcherStampsATaskDescribedLaunchByItsPrefix` fails — a `Task <ids>/<n>` launch is keyed `Task` again and its row rides time pairing; `TestStampSessionDispatchAgentStampsATaskPrefix` fails — a prefix key stamps nothing, or stamps a closed or `task-130-` row.
**Baseline:** before=170 after=172 <!-- measured: cat stats/internal/store/records_test.go stats/internal/harvest/watcher_test.go | grep -c -E '^func Test' @ ef04314f → 170 -->
**Commit:** fix(harvest): stamp a launch described by its task ids onto its task row
**After:** none

**Decision:** task-description-stamps-by-prefix

**Decision:** k1023-rows-repair

- [ ] 4. Settle closed agentless and inline rows, and restamp the seven inline rows
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `docker exec flow-postgres psql -U flow -d flow -tc "select id, dispatch_key, coalesce(agent_id,''), metrics ? 'tokens' from dispatches where id in (7562,7567,7577,7613,7616,7621,7624) order by id"` prints five empty agent ids (7562, 7567, 7577, 7613, 7621), `inline` for 7616 and 7624, and `f` on every row; `go run ./cmd/flowd restamp-dispatch 7562 <parent>` is refused (the row is not inline).
    <!-- measured: that query @ dev store 2026-10-25 → as stated -->
    <!-- unverified:restamp-refusal — the refusal is read from restampDispatch's doc comment and RestampInlineDispatch's `agent_id = 'inline'` filter; run it first -->
  - [ ] **Step 2: Write the failing tests.** In `stats/internal/store/records_test.go`: `TestDispatchEndedAtReadsTheLatestRow` — the new `DispatchEndedAt(ctx, token, key)` returns the latest row's `ended_at` and agent id, a zero time for an open row, and `found == false` for no row; `TestMarkDispatchInlineFillsOnlyAnEmptyID` — the new `MarkDispatchInline(ctx, id)` sets `inline` on an agentless row (true) and leaves a row naming an agent untouched (false). In `stats/internal/harvest/watcher_test.go`: `TestWatcherSettlesClosedRows`, built like the inline-backfill tests (`inlineBackfillDeps`), with cases — an agentless row ended with no launch or denial between its begin and its end is stamped `inline` and restamped with its share; an agentless row with an Agent launch inside that interval is left untouched; an inline row whose window holds no parent record is restamped with a patch carrying a `tokens` key of zeros; an open row, and a row ended after the batch's newest record, are kept and settled by a later batch. Run them; they fail.
  - [ ] **Step 3: Implement.** (a) `stats/internal/store/records.go`: `DispatchEndedAt` (latest row under `(session_token, dispatch_key)`, `ORDER BY started_at DESC, id DESC`) and `MarkDispatchInline` (`UPDATE dispatches SET agent_id = 'inline' WHERE id = $1 AND COALESCE(agent_id, '') = ''`). (b) `stats/internal/harvest/deps.go`: `DispatchEndedAt` on `DispatchAgentStamper` and its `NoDeps` answer (not found). (c) `stats/internal/harvest/watcher.go`: a per-path bounded list of settle candidates — every agentless begin `dispatchBeginsFromCommand` yields, taken before `stampKeyedLaunches` drops any, plus every `dispatch begin … -agent-id inline` invocation (a sibling of `inlineBackfillsFromCommand`) — and a per-path bounded list of every launch and denial instant seen, each bound a named constant carrying a `ponytail:` comment (oldest dropped; a candidate dropped stays unsettled, the ordinary silence). A new `settleClosedRows`, called right after `measureInlineBackfills` with the same committed offset, settles each candidate whose row has ended at or before the batch's newest record: agentless with no launch or denial instant in `[begin, ended_at]` → `StampDispatchAgent(token, key, "inline")`; then, inline, `InlineDispatchID` → `ClosedInlinePatch` over the transcript up to the committed offset → `RestampInlineDispatch`; the candidate is dropped once settled. `ClosedInlinePatch` is `InlineDispatchPatch` with `ErrNoInlineShare` answered by a zero patch whose JSON carries a `tokens` key. (d) `stats/cmd/flowd/main.go`, `restampDispatch`'s parent-transcript branch: `MarkDispatchInline` first (a row naming another agent stays refused by `RestampInlineDispatch`), then `ClosedInlinePatch`; update its doc comment and usage line.
  - [ ] **Step 4: Prove the tests pin the fix.** Save copies first. Make `settleClosedRows` return at once — `go test ./internal/harvest -run 'TestWatcherSettlesClosedRows' -count=1` fails; restore. Make the launch-inside check always false — the launch-inside case fails; restore. Make `ClosedInlinePatch` return `ErrNoInlineShare` unchanged — the zero-share case fails; restore. Change `MarkDispatchInline`'s filter to `WHERE id = $1` — `go test ./internal/store -run 'TestMarkDispatchInlineFillsOnlyAnEmptyID' -count=1` fails; restore.
  - [ ] **Step 5: Verify.** `gofmt -l .` prints nothing; `go test ./internal/harvest ./internal/store -run 'TestDispatchEndedAtReadsTheLatestRow|TestMarkDispatchInlineFillsOnlyAnEmptyID|TestWatcherSettlesClosedRows|TestWatcherMeasuresABackfilledInlineRow' -count=1` passes and `go build ./cmd/flowd` exits 0; `make -s web-dist && go vet ./...` exits 0. Commit.
  - [ ] **Step 6: Repair the seven rows**, after the commit: `make -s web-dist`, then for each of 7562, 7567, 7577, 7613, 7616, 7621 and 7624, `go run ./cmd/flowd restamp-dispatch <id> /Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/e905b265-37ae-40b2-9776-72e831e83ebc.jsonl`.
  - [ ] **Step 7: Verify the repair.** Step 1's query prints `inline` and `t` on all seven rows; with `G=/Users/tweety53/Projects/gymie-worktrees/kan-1023-templates-step-2-fe-save-as-template-template`, `flow record render -change kan-1023-templates-step-2-fe-save-as-template-template -kind ledger -repo $G -C $G` then `grep -c 'not measured' $G/.superpowers/sdd/ledgers/kan-1023-templates-step-2-fe-save-as-template-template.md` prints `0`; report both outputs.
    <!-- measured: that render and grep @ dev store 2026-10-25 → 8 -->

**Build:** green
**Files:** `stats/internal/harvest/deps.go`, `stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/cmd/flowd/main.go`
**Tests:** `TestDispatchEndedAtReadsTheLatestRow`, `TestMarkDispatchInlineFillsOnlyAnEmptyID`, `TestWatcherSettlesClosedRows`
**Regression:** `TestWatcherSettlesClosedRows` fails — a closed agentless row no launch ran stays unmeasured, a row a launch ran is mis-stamped inline, or a zero share renders "not measured" again; `TestDispatchEndedAtReadsTheLatestRow` fails — the settle pass reads the wrong row's end; `TestMarkDispatchInlineFillsOnlyAnEmptyID` fails — the repair overwrites an agent's id.
**Baseline:** before=172 after=175 <!-- measured: cat stats/internal/store/records_test.go stats/internal/harvest/watcher_test.go | grep -c -E '^func Test' @ ef04314f → 170, 172 after Task 3 -->
**Commit:** fix(harvest): settle closed agentless and inline dispatch rows
**After:** Task 3

**Decision:** settle-closed-rows

**Decision:** k1023-rows-repair
