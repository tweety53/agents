# self-review-kan-1018-week-programs-plan-a-week

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the four non-`big` findings of `kan-1018-week-programs-plan-a-week`'s self-review pass.

**Architecture:** one refusal in the hook's commit-then-push branch, one arm in the dispatch upsert,
one signal in `check-unfinished-work`, one doc correction in gymie.

**Spec:** `design.md` beside this file.

## Global constraints

- One commit per task with its `**Commit:**` subject and a `Flow-Key: <dispatch key>` trailer, no `Task-Id:` trailer. Never write `spectre/changes/` in the agents worktree. Never merge or push.
- Tasks 1–3 work in `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1018-week-programs-plan-a-week`, branch `self-review-kan-1018-week-programs-plan-a-week`; their `**Files:**` are relative to its root. Task 4 works in `/Users/tweety53/Projects/gymie-worktrees/kan-1018-week-programs-plan-a-week`, branch `spectre/kan-1018-week-programs-plan-a-week`. Never touch a main checkout; never create a worktree.
- Run `gofmt -w` on every Go file changed; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.
- Go test commands run as `cd stats && make -s web-dist && go test <pkg> -run '<name>' -count=1`, Bash timeout 600000.

## Review Focus

- Task 1: a commit with a readable trailered message followed by a push in one call still passes; the commit-only refusal still never judges an unreadable message.
- Task 2: a re-begin with a later instant, or on an ended row, leaves `started_at` alone.
- Task 3: a round whose fixed findings carry a `none — <reason>` mutation row passes; withdrawn findings never demand a row.

- [x] 1. Refuse a commit-then-push to spectre/* whose commit message the hook cannot read
  - [x] **Step 1: Write the failing test** `TestCommitThenPushWithUnreadableMessageIsDenied` in `stats/internal/repotest/protect_main_checkout_test.go`, following the file's existing harness: in a scratch repo on branch `spectre/x`, feed the hook `git commit -F msg.txt && git push origin spectre/x`, and also `git commit -m "$(cat msg)" && git push origin HEAD:spectre/x`; expect a deny naming "commit and push in two calls". Add a passing case: `git commit -F msg.txt` alone is allowed. Run it; it fails.
    <!-- unverified:the file's harness can drive a commit-then-push command against a scratch spectre/ branch -->
  - [x] **Step 2: Fix.** In `hooks/protect-main-checkout.py`, extract the unreadable-message test of `commit_lacks_trailer` (`UNREAD_MESSAGE_OPTS` prefixes, no message, `$`/backtick/leading-quote message) into `commit_message_unread(args)`; `commit_lacks_trailer` returns False when it is true (unchanged behaviour). In the commit-then-push loop (`for ct, cargs in commits:`), when `dsts` is non-empty, the commit is on `target`, and `commit_message_unread(cargs)`, append: `Blocked: this command commits and then pushes to <dst>, and the hook cannot read that commit's message (-F, -C, --amend, or a shell expansion), so it cannot tell whether it carries a `Task-Id:` or `Flow-Key:` trailer. Commit and push in two calls — the push check then reads the commit (**Commits outside every task**, skills/flow-contracts/git-boundaries.md).` then `break`.
  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/repotest/ -run 'TestCommitThenPushWithUnreadableMessageIsDenied|ProtectMainCheckout' -count=1` passes; `python3 -m py_compile hooks/protect-main-checkout.py`.
  - [x] **Step 4: Commit** the two files.

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `stats/internal/repotest/protect_main_checkout_test.go`
**Tests:** `TestCommitThenPushWithUnreadableMessageIsDenied`
**Regression:** `TestCommitThenPushWithUnreadableMessageIsDenied` — reverting lets `git commit -F … && git push origin spectre/x` through and the test's deny assertion fails.
**Baseline:** before=4 after=5
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/repotest/protect_main_checkout_test.go @ d8471213 -->
**Commit:** fix(hooks): refuse a commit-then-push whose message the hook cannot read
**After:** none

**Decision:** p1-deny-unreadable-commit-push

- [x] 2. Let a backfill move an open dispatch row's begin back
  - [x] **Step 1: Write the failing test** `TestBeginDispatchMovesAnOpenRowsBeginEarlier` in `stats/internal/store/records_test.go`, following the file's existing dispatch fixtures: begin inline row K at T, begin K again at T-6s, end K with `CommitAt` T-3s and EndedAt T+10s; expect the end to succeed and `started_at == T-6s`. Second case in the same test: begin K2 at T, re-begin at T+5s — `started_at` stays T. Run it; the first case fails with `ErrInlineBegunAfterCommit`.
  - [x] **Step 2: Fix.** In `insertDispatch`'s ON CONFLICT clause (`stats/internal/store/records.go`), make the `started_at` arm `CASE WHEN `+reBegin+` THEN EXCLUDED.started_at WHEN dispatches.ended_at IS NULL AND dispatches.outcome IS NULL AND EXCLUDED.started_at < dispatches.started_at THEN EXCLUDED.started_at ELSE dispatches.started_at END`, and add one sentence to the comment above `reBegin`: an open row takes an earlier re-begin's instant, so `flow record dispatch backfill` repairs a row begun after its commit.
  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/store/ -run 'TestBeginDispatch|TestEndDispatch' -count=1` passes; `go vet ./internal/store && gofmt -l internal/store` prints nothing.
  - [x] **Step 4: Commit** the two files.

**Build:** green
**Files:** `stats/internal/store/records.go`, `stats/internal/store/records_test.go`
**Tests:** `TestBeginDispatchMovesAnOpenRowsBeginEarlier`
**Regression:** `TestBeginDispatchMovesAnOpenRowsBeginEarlier` — reverting keeps the late begin and the end refuses with `ErrInlineBegunAfterCommit`.
**Baseline:** before=106 after=107
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/store/records_test.go @ d8471213 -->
**Commit:** fix(store): let a backfill move an open dispatch row's begin back
**After:** none

**Decision:** p2-open-row-earlier-begin

- [x] 3. Refuse landing when a panel fix round recorded no fix-mutation row
  - [x] **Step 1: Confirm the round mapping.** Read how `records.Mutation.Round` is recorded (`skills/flow/review-panel-fix-round.md`, `flow record` mutation subcommand in `stats/cmd/flow/record.go`): whether a fix of round-N findings records its mutations under round N or N+1. Use that mapping below.
    <!-- unverified:the round a fix-mutation row carries relative to the round its fixed findings were raised in -->
  - [x] **Step 2: Write the failing test** `TestUnfinishedWorkRefusesAFixedRoundWithNoMutationRow` in `stats/internal/guard/check_unfinished_work_test.go`, on the file's existing stubbed `flow record` harness: findings round 0 `fixed` ×2, round 1 `withdrawn reason` ×1, no mutation rows → refuse naming round 0; add a mutation row (`none — doc-only`) for the mapped round → passes; only withdrawn findings → passes. Run it; it fails.
  - [x] **Step 3: Fix.** In `stats/internal/guard/unfinishedwork.go`, beside the open-findings signal, read the change's mutation rows through `pcFlow` (the same family as `record findings`; `record passes` if it carries mutations, else the run render's JSON), collect each round holding a finding whose status is exactly `fixed`, and append the reason `fixed finding(s) in round(s) <list> and no fix-mutation row for that fix round — record each fix's fix-mutation: line (skills/flow/review-panel-fix-round.md) before landing`. A store read failure exits 2 like the findings read.
  - [x] **Step 4: Measure against history.** Feed the new signal (recorded fixtures in the test are enough) the the mutation and findings rows of two recent archived changes that recorded mutations (`flow record passes -change <name>` read-only) and confirm no refusal; report the names.
  - [x] **Step 5: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestUnfinishedWork|TestCheckUnfinishedWork' -count=1` passes; `go vet ./internal/guard && gofmt -l internal/guard` prints nothing.
  - [x] **Step 6: Commit** the two files.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `stats/cmd/flow/record.go`, `skills/flow/review-panel-fix-round.md`, `spectre/changes/self-review-kan-1018-week-programs-plan-a-week/verbatim-moves.txt`
**Tests:** `TestUnfinishedWorkRefusesAFixedRoundWithNoMutationRow`
**Regression:** `TestUnfinishedWorkRefusesAFixedRoundWithNoMutationRow` — reverting lets a change with fixed findings and no mutation rows land.
**Baseline:** before=10 after=11
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/guard/check_unfinished_work_test.go @ d8471213 -->
**Commit:** fix(guard): refuse landing when a fix round recorded no fix-mutation row
**After:** none

**Decision:** p3-mutation-row-per-fixed-round

- [ ] 4. Correct kan-1018's live check status and round-0 counts
  - [ ] **Step 1: Edit** `/Users/tweety53/Projects/gymie-worktrees/kan-1018-week-programs-plan-a-week/spectre/changes/archive/kan-1018-week-programs-plan-a-week/live-verification.md` line 7: `200, member program created` → `201, member program created`. Edit `/Users/tweety53/Projects/gymie-worktrees/kan-1018-week-programs-plan-a-week/spectre/changes/archive/kan-1018-week-programs-plan-a-week/narrative.md`: `seven Importants and five Minors` → `six Importants and six Minors`.
    <!-- measured: sed -n 7p live-verification.md and grep -n "seven Importants" narrative.md @ spectre/kan-1018-week-programs-plan-a-week; WeekProgramController.kt:44 ResponseEntity.created -->
  - [ ] **Step 2: Verify.** `git -C /Users/tweety53/Projects/gymie-worktrees/kan-1018-week-programs-plan-a-week grep -n -F -e '| 200, member program created' -e 'seven Importants and five Minors' -- spectre/changes/archive/kan-1018-week-programs-plan-a-week` prints nothing.
  - [ ] **Step 3: Commit** the two files in that worktree on `spectre/kan-1018-week-programs-plan-a-week`.

**Build:** green
**Files:** `/Users/tweety53/Projects/gymie-worktrees/kan-1018-week-programs-plan-a-week/spectre/changes/archive/kan-1018-week-programs-plan-a-week/live-verification.md`, `/Users/tweety53/Projects/gymie-worktrees/kan-1018-week-programs-plan-a-week/spectre/changes/archive/kan-1018-week-programs-plan-a-week/narrative.md`
**Tests:** none — documentation correction.
**Regression:** none — no test.
**Commit:** docs(spectre): correct kan-1018's live check status and round-0 counts
**After:** none
