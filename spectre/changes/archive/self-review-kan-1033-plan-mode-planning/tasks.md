# self-review-kan-1033-plan-mode-planning

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1033's self-review findings 1.2 (Task 1), 1.3 (Task 2), 1.4 (Task 3),
4.1 (Task 4), 4.2 (Task 5) and 5.1 (Task 6), as `proposal.md` names them.

**Architecture:** `check-untasked-commits` refuses a `pipeline-fix-` Flow-Key on a change branch;
one rationale clause corrected; `gsPlaceholder` takes a word prefix; `run-verify-list` and
`panel-round-begins` record the verify row and the round's pass row themselves; the Bash hook
denies a bare `cat` redirect; three kan-1033 store rows retimed.

**Tech Stack:** Go (`stats/`), Python (`hooks/`), Bash, Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-25-self-review-kan-1033-plan-mode-planning-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject, no `Task-Id:` trailer and no `Co-Authored-By:` trailer. Never write `spectre/changes/` — this change's `patches/` and `verbatim-moves.txt` are committed with the plan. Never merge or push.
- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1033-plan-mode-planning` (`<agents-wt>` below), branch `spectre/kan-1033-plan-mode-planning`. Every `**Files:**` path is relative to it. Never touch a main checkout, and never create a worktree.
- Tasks 1, 2, 4 and 6 apply `<agents-wt>/spectre/changes/self-review-kan-1033-plan-mode-planning/patches/task-<n>.diff` with `git apply <that path>` from `<agents-wt>` for their prose; `git apply --check` passed for all four at `fbd8bbd5`. A patch that does not apply is reported, never hand-merged by guesswork. Everything else is written from the steps.
- Prose lands character for character as its patch carries it: `scripts/check-verbatim-moves.sh` matches every added, reworded or removed sentence against this change's `verbatim-moves.txt`.
- Go: run `gofmt -w` on every Go file you change before committing; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0. Every Go test command below runs from `<agents-wt>/stats`.
- Mutation proofs restore with `break-and-prove.sh` or `mutate-and-verify.sh`, or a saved copy restored with `cp` — never `git checkout --` on a file carrying uncommitted edits.
- Task 4's store repair writes the live store through `go run ./cmd/flowd retime-dispatch` and `go run ./cmd/flowd restamp-dispatch` from `<agents-wt>/stats`, which open the database directly; never start, stop or restart the protected `flowd` on `127.0.0.1:4173` nor the `flow-postgres` container, and never run tests against the live database.
- The kan-1033 parent transcript is `/Users/tweety53/.claude/projects/-Users-tweety53-Projects-agents/1533d335-98df-4487-b1cc-eeb767026cc8.jsonl` (`<parent>` below).

## Review Focus

- Task 1: a `Flow-Key: pipeline-fix-…` commit on a `pipeline-fix-<name>-<k>` branch is never listed; a detached HEAD is judged as a change branch; a `Flow-Key: panel-fix-1` commit still passes.
- Task 3: `" --help"` and `" foo bar"` still fail the placeholder test, so no plain-prose span joins the required set.
- Task 4: without the five trailing arguments nothing is recorded and no `flow` binary is needed; a non-zero list records no `end`; a failed record call is printed on stderr and never changes the list's exit code.
- Task 5: a heredoc, `cat > f < in`, `x | cat > f` and `cat a > f` are allowed; a hook fault still fails open.
- Task 6: a round whose every begin failed records no pass row; a failed pass record exits 1 with its output relayed.

- [x] 1. Refuse a pipeline-fix commit on a change branch
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In a scratch repo (`git init -q -b main`, one base commit, `git checkout -qb spectre/demo`, a `src/a.kt` commit `fix(app): a` with `-m 'Flow-Key: pipeline-fix-mid-run-question'`), `scripts/check-untasked-commits.sh <scratch> main` prints `UNTASKED-OK: 1 commit(s) checked` and exits 0.
    <!-- unverified: run the scratch reproduction; the guard's code (untaskedcommits.go) skips every commit with a non-empty Flow-Key -->
  - [x] **Step 2: Write the failing test** `TestCheckUntaskedCommitsRefusesAPipelineFixKey` in `stats/internal/guard/check_untasked_commits_test.go`, on the fixture shape of `TestCheckUntaskedCommits` with the commits made on a named branch: on `spectre/demo`, a `fix(app): a` commit with `Flow-Key: pipeline-fix-mid-run-question` exits 1, its stdout carries `UNTASKED: 1 commit(s)` and a line holding `fix(app): a — Flow-Key: pipeline-fix-… on a change branch`; the same commit on `pipeline-fix-demo-3` exits 0 with `UNTASKED-OK`; on `spectre/demo` a `Flow-Key: panel-fix-1` commit exits 0. Run `go test ./internal/guard -run 'TestCheckUntaskedCommitsRefusesAPipelineFixKey' -count=1` — it fails.
  - [x] **Step 3: Implement** in `stats/internal/guard/untaskedcommits.go`: read the branch once with `git -C <worktree> symbolic-ref --short -q HEAD` (empty on a detached HEAD); a commit whose `Flow-Key:` value opens `pipeline-fix-` is listed as `  <sha12> <subject> — Flow-Key: pipeline-fix-… on a change branch` unless the branch matches `^pipeline-fix-.+-[0-9]+$`. The verdict line's text is unchanged.
  - [x] **Step 4: Apply `patches/task-1.diff`** from `<agents-wt>`: the shim header's exception paragraph (`scripts/check-untasked-commits.sh`) and the **Commits outside every task** sentence (`skills/flow-contracts/git-boundaries.md`).
  - [x] **Step 5: Prove the test pins the fix.** Save a copy of `untaskedcommits.go`, drop the `pipeline-fix-` listing — the Step 2 test fails; restore the copy.
  - [x] **Step 6: Verify.** `gofmt -l internal/guard` prints nothing; `go test ./internal/guard -run 'TestCheckUntaskedCommits' -count=1` passes; from `<agents-wt>`, `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-references.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/untaskedcommits.go`, `stats/internal/guard/check_untasked_commits_test.go`, `scripts/check-untasked-commits.sh`, `skills/flow-contracts/git-boundaries.md`
**Tests:** `TestCheckUntaskedCommitsRefusesAPipelineFixKey`
**Regression:** `TestCheckUntaskedCommitsRefusesAPipelineFixKey` fails — an in-run pipeline fix committed on a change branch passes the push hook, `close-task.sh` and run 1's gate again.
**Baseline:** before=3 after=4 <!-- measured: grep -c -E '^func Test' stats/internal/guard/check_untasked_commits_test.go @ fbd8bbd5 → 3 -->
**Commit:** fix(guard): refuse a pipeline-fix commit on a change branch
**After:** none

**Decision:** untasked-refuses-pipeline-fix-key

- [x] 2. Correct the effort-parameter rationale of the flow-effort family
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `grep -c "while \`effort\` has no dispatch-time parameter" skills/flow/SKILL-rationale.md` prints `1`.
    <!-- measured: that grep @ fbd8bbd5 → 1 -->
  - [x] **Step 2: Apply `patches/task-2.diff`** from `<agents-wt>`: the corrected clause (`skills/flow/SKILL-rationale.md`).
  - [x] **Step 3: Verify.** The Step 1 grep prints `0`; `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-references.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/SKILL-rationale.md`
**Tests:** none — a rationale clause; no run loads the file.
**Regression:** none — no test; reverting restores the stale clause.
**Commit:** docs(flow): correct the effort-parameter rationale of the flow-effort family
**After:** none

**Decision:** effort-rationale-is-stale

- [x] 3. Read a literal-prefixed placeholder as an invocation
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `gsPlaceholder.MatchString(" fp-<literal-token> <n>")` is false at `stats/internal/guard/guardsymlinks.go:101`.
    <!-- measured: scratch go run of the pattern @ fbd8bbd5 → false -->
  - [x] **Step 2: Write the failing test** `TestGsPlaceholderLiteralPrefix` in `stats/internal/guard/check_guard_symlinks_test.go`: `gsPlaceholder.MatchString` is true for `" fp-<literal-token> <n>"`, `" <session-token> <n>"`, `" targets <w>"`, `" [--clean <c>] <f>"`, and false for `" --help"`, `" foo bar"`, `""`. Run `go test ./internal/guard -run 'TestGsPlaceholderLiteralPrefix' -count=1` — it fails on the first case.
  - [x] **Step 3: Implement.** The last element of `gsPlaceholder` becomes `[ \t]+[A-Za-z0-9._-]*<`; extend the comment above it with the `fp-<literal-token>` case (kan-1033).
  - [x] **Step 4: Verify.** `gofmt -l internal/guard` prints nothing; `go test ./internal/guard -run 'TestGsPlaceholderLiteralPrefix|TestCheckGuardSymlinks' -count=1` passes; from `<agents-wt>`, `scripts/check-guard-symlinks.sh` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/guardsymlinks.go`, `stats/internal/guard/check_guard_symlinks_test.go`
**Tests:** `TestGsPlaceholderLiteralPrefix`
**Regression:** `TestGsPlaceholderLiteralPrefix` fails — a citation whose first argument is a literal-prefixed placeholder leaves the required set, and its missing symlink reads OK again.
**Baseline:** before=4 after=5 <!-- measured: grep -c -E '^func Test' stats/internal/guard/check_guard_symlinks_test.go @ fbd8bbd5 → 4 -->
**Commit:** fix(guard): read a literal-prefixed placeholder as an invocation
**After:** none

**Decision:** placeholder-word-prefix

- [x] 4. Record the inline verify row from run-verify-list
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `flow record dispatches -change kan-1033-plan-mode-planning -C /Users/tweety53/Projects/agents` shows id 7650 `verify` from 19:01:46+02:00 to 19:26:46+02:00, enclosing id 7647 `verify-fix-1` and id 7651 `verify-2`; `grep -c '"dispatch"' stats/internal/guard/runverifylist.go` prints `0`.
    <!-- measured: flow record dispatches -change kan-1033-plan-mode-planning @ the live store, read 2026-10-25, not re-runnable once repaired → rows 7647, 7650, 7651 as stated; the grep @ fbd8bbd5 → 0 -->
  - [x] **Step 2: Write the failing test** `TestRunVerifyListRecordsTheVerifyRow` in `stats/internal/guard/run_verify_list_test.go`, with a `flow` stub on `PATH` logging its argv (the stub pattern of `TestPanelRoundBeginsRecordsLiteralKeys`): `runVerifyList` with `<wt> lint kan-1 verify mf-k1 opus low` over a passing list logs `record dispatch begin -C <wt> -change kan-1 -role verifier -model opus -effort low -agent-id inline -key verify -session-token mf-k1` before the list's first command and `record dispatch end -C <wt> -change kan-1 -key verify -session-token mf-k1 -outcome completed` after its last; a failing list logs the begin and no end and still exits 1; the two-argument form logs nothing; three to six arguments exit 2.
  - [x] **Step 3: Implement** in `stats/internal/guard/runverifylist.go`: accept two or seven arguments (usage `run-verify-list.sh <worktree> lint|test [<change> <key> <session-token> <model> <effort>]`, in the header comment, both usage lines and `scripts/run-verify-list.sh`'s `# Usage:` line); with seven, resolve `flow` with `lookPath(env, "flow")` (absent: exit 2 before any command runs), record the begin before the first command, and the end only when the list's exit is 0. A record call that fails prints `run-verify-list: flow record dispatch <begin|end> failed: <its output>` on stderr and leaves the exit code as the list decided.
  - [x] **Step 4: Apply `patches/task-4.diff`** from `<agents-wt>`: the new **Recording** paragraph and block (`skills/flow/verify-and-handoff.md`).
  - [x] **Step 5: Prove the test pins the fix.** Save a copy of `runverifylist.go`, drop the begin call — the Step 2 test fails; restore the copy.
  - [x] **Step 6: Repair kan-1033's rows.** From `<agents-wt>/stats`, retime each row to the instants `<parent>` holds, then restamp its figures over `<parent>`:

```bash verified:instants read from the parent transcript's tool_use and tool_result timestamps, 2026-10-25
go run ./cmd/flowd retime-dispatch 7647 2026-10-25T17:06:21Z 2026-10-25T17:08:18Z 5249db0cce81c79b3f621af52d6016dc93b32e94
go run ./cmd/flowd retime-dispatch 7650 2026-10-25T17:00:56Z 2026-10-25T17:06:17Z ""
go run ./cmd/flowd retime-dispatch 7651 2026-10-25T17:08:20Z 2026-10-25T17:13:38Z ""
P=/Users/tweety53/.claude/projects/-Users-tweety53-Projects-agents/1533d335-98df-4487-b1cc-eeb767026cc8.jsonl
for id in 7647 7650 7651; do go run ./cmd/flowd restamp-dispatch $id "$P"; done
```

    Then the Step 1 read shows the three rows at the new instants, no two overlapping.
    <!-- unverified: confirm go run ./cmd/flowd reaches the live database with the default DSN, as self-review-kan-1023's Task 3 did -->
  - [x] **Step 7: Verify.** `gofmt -l internal/guard` prints nothing; `go test ./internal/guard -run 'TestRunVerifyList' -count=1` passes; from `<agents-wt>`, `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-references.sh`, `scripts/check-guard-symlinks.sh` and `scripts/check-installed-citations.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/runverifylist.go`, `stats/internal/guard/run_verify_list_test.go`, `scripts/run-verify-list.sh`, `skills/flow/verify-and-handoff.md`
**Tests:** `TestRunVerifyListRecordsTheVerifyRow`
**Regression:** `TestRunVerifyListRecordsTheVerifyRow` fails — the verify row is hand-recorded again, and a forgotten begin is back-filled with a guessed start.
**Baseline:** before=1 after=2 <!-- measured: grep -c -E '^func Test' stats/internal/guard/run_verify_list_test.go @ fbd8bbd5 → 1 -->
**Commit:** fix(guard): record the inline verify row from run-verify-list
**After:** none

**Decision:** verify-list-records-row

- [x] 5. Deny a bare cat redirect that waits on stdin
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `printf '%s' '{"tool_name":"Bash","tool_input":{"command":"cat > /tmp/x.go"},"cwd":"/tmp"}' | python3 hooks/protect-main-checkout.py` prints nothing (allowed).
    <!-- unverified: run it; main() denies only landings and protected paths -->
  - [x] **Step 2: Write the failing test** `TestProtectMainCheckoutDeniesABareCatRedirect` in `stats/internal/repotest/protect_main_checkout_test.go`, driving the hook as `TestProtectMainCheckout` does: denied — `cat > /tmp/x.go`, `cat >> f`, `cat >f; echo`, `a && cat > f`, `false || cat > f`, `(cat > f)`; allowed — `cat > f <<'EOF'` with a body and `EOF` line, `cat > f < in`, `x | cat > f`, `cat a > f`, `cat -n f > g`, `echo cat > f`, `concat > f`. Each denial's reason names the Write tool and a heredoc. Run `go test ./internal/repotest -run 'TestProtectMainCheckoutDeniesABareCatRedirect' -count=1` — it fails.
  - [x] **Step 3: Implement** in `hooks/protect-main-checkout.py`'s Bash branch, before `bash_hits`: a module-level pattern compiled with `re.M`, `(?:^|[;&(]|\|\|)\s*cat\s*>>?\s*[^\s<>|;&()]+\s*(?=$|[;&)]|\|\|)`, denying with `permissionDecisionReason` `a bare \`cat > <file>\` reads stdin, which a Bash call never has, so it waits until the call times out — write the file with the Write tool, or give cat a heredoc (\`cat > <file> <<'EOF'\`)`. Mark the ceiling with a `# ponytail:` comment: a heredoc body line opening with `cat > <file>` is denied too.
  - [x] **Step 4: Verify.** `go test ./internal/repotest -run 'TestProtectMainCheckout' -count=1` passes. Commit.

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `stats/internal/repotest/protect_main_checkout_test.go`
**Tests:** `TestProtectMainCheckoutDeniesABareCatRedirect`
**Regression:** `TestProtectMainCheckoutDeniesABareCatRedirect` fails — a bare `cat >` runs again and hangs its Bash call.
**Baseline:** before=2 after=3 <!-- measured: grep -c -E '^func Test' stats/internal/repotest/protect_main_checkout_test.go @ fbd8bbd5 → 2 -->
**Commit:** fix(hooks): deny a bare cat redirect that waits on stdin
**After:** none

**Decision:** bare-cat-in-protect-hook

- [x] 6. Record a panel round's pass row with its begins
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `grep -c '"pass"' stats/internal/guard/panelroundbegins.go` prints `0`: the script records begins only.
    <!-- measured: that grep @ fbd8bbd5 → 0 -->
  - [x] **Step 2: Write the failing test** `TestPanelRoundBeginsRecordsThePassRow` in `stats/internal/guard/panel_round_begins_test.go`, on the stub of `TestPanelRoundBeginsRecordsLiteralKeys`: args `/wt kan-1 4 mf-k1 primary:opus:low:a515c200 test-sanity+call-efficiency:sonnet:low` log, after the two begins, `record pass -C /wt -change kan-1 -round 4 -note dispatched: panel-4-primary on opus/low over a515c200; panel-4-test-sanity+call-efficiency on sonnet/low`; with a refused begin (`bad:opus:low primary:opus:low`) the note names `panel-4-primary` alone; with every begin refused no `record pass` line is logged; a stub refusing `record pass` makes the call exit 1. Run `go test ./internal/guard -run 'TestPanelRoundBeginsRecordsThePassRow' -count=1` — it fails.
  - [x] **Step 3: Implement** in `stats/internal/guard/panelroundbegins.go`: after the begins, when at least one succeeded, one `flow record pass -C <wt> -change <change> -round <round> -note <note>` call, the note `dispatched: ` plus each succeeded slot's `panel-<round>-<slot> on <model>/<effort>[ over <diff-base>]` joined by `; `; its output printed as `pass: <output>`; a failure sets exit 1. Update the header comment.
  - [x] **Step 4: Apply `patches/task-6.diff`** from `<agents-wt>`: the round's begins sentence (`skills/flow/review-panel.md`).
  - [x] **Step 5: Verify.** `gofmt -l internal/guard` prints nothing; `go test ./internal/guard -run 'TestPanelRoundBegins' -count=1` passes; from `<agents-wt>`, `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.sh`, `scripts/check-references.sh` and `scripts/check-guard-symlinks.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/panelroundbegins.go`, `stats/internal/guard/panel_round_begins_test.go`, `skills/flow/review-panel.md`
**Tests:** `TestPanelRoundBeginsRecordsThePassRow`
**Regression:** `TestPanelRoundBeginsRecordsThePassRow` fails — a round's pass row is hand-recorded again, and a round can reach landing without one.
**Baseline:** before=1 after=2 <!-- measured: grep -c -E '^func Test' stats/internal/guard/panel_round_begins_test.go @ fbd8bbd5 → 1 -->
**Commit:** fix(guard): record a panel round's pass row with its begins
**After:** none

**Decision:** round-begins-records-pass
