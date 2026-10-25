## Context

kan-1033's self-review pass raised six non-`big` findings (`proposal.md`); this change fixes all
six on the change's own branch, `spectre/kan-1033-plan-mode-planning`, in its own worktree, per
step 3 of `skills/flow-self-review/SKILL.md`. The evidence for each is the pass file's sections
1.2, 1.3, 1.4, 4.1, 4.2 and 5.1. The brainstorm record is
`.superpowers/sdd/2026-10-25-self-review-kan-1033-plan-mode-planning-design.md` in this worktree.

What the planning measured beyond the pass:

- **1.2.** `git log --all --grep='Flow-Key: pipeline-fix'` in the agents checkout finds one commit
  ever, `ba68637c` — the defect itself; a pipeline-fix loop's own fixer commits carry no
  `Flow-Key:`. kan-1033's branch has since been reshaped (`b1860ac5` holds `ba68637c`'s content),
  so its trailer no longer exists on the branch. Integrate's `check-loop-landing.sh` runs after
  that reshape (**Run 1 — the branch is not merged**, `skills/flow-contracts/finish-contract-run1.md`);
  `check-untasked-commits.sh` runs before it — at every typed push of `spectre/<name>`, in
  `close-task.sh`, and in run 1's unfinished-work gate. The pipeline-fix loop records its rows under
  the change `pipeline-fix-<name>-<k>` (`lpFromKey`, `stats/internal/guard/loop.go`), never under
  the change it runs beside, so kan-1033's row `pipeline-fix-mid-run-question` (role `panel-fix`,
  filed under kan-1033) is the bypass's record. Its content was re-read by panel round 3
  (`primary+principles`, pass row 3: "re-read the post-panel commits (pipeline rule, hook test
  seam)"), so nothing it wrote is left to repair.
- **1.3.** The dispatching session's Agent tool lists an `effort` parameter ("Reasoning effort for
  this agent. Set this ONLY when the user, or instructions such as CLAUDE.md or a skill, explicitly
  ask…"), quoted in this planner's dispatch. The planning texts (`skills/flow/brainstorm-planner.md`,
  `skills/flow/brainstorm.md`, `skills/flow-plan/SKILL.md`, `skills/flow-self-review/SKILL.md`)
  are therefore right, and the defect is the rationale's clause at
  `skills/flow/SKILL-rationale.md:741`. The built-in `Explore`/`Plan` types need no `effort:`
  definition, since the dispatch-time parameter sets it; the `low` half of decision
  `low-effort-40-60` applies as written. A loop dispatch's effort is read off its `subagent_type`
  by `check-loop-dispatch.sh` (`stats/internal/guard/loop.go`, `strings.TrimPrefix(typ, "flow-")`),
  which is why the `flow-<effort>` family still carries `effort:`.
- **1.4.** A scratch `go run` of the current `gsPlaceholder` and of the fix's pattern over the
  pass's cases: `" fp-<literal-token> <n>"` false → true; `" <session-token> <n>"`,
  `" targets <w>"` and `" [--clean <c>] <f>"` true → true; `" --help"` and `" foo bar"` false →
  false.
- **4.1.** The parent transcript
  (`/Users/tweety53/.claude/projects/-Users-tweety53-Projects-agents/1533d335-98df-4487-b1cc-eeb767026cc8.jsonl`)
  holds the real instants: the first verify's Bash call ran from 17:00:56Z (lint and test groups
  started together) to its result at 17:06:17Z; the fix work began at 17:06:21Z and its commit
  `5249db0c` landed by 17:08:18Z; `verify-2` ran inside the 17:08:17Z call from 17:08:20Z to its
  result at 17:13:38Z. The back-fill at 17:26:45Z stamped `verify` (id 7650) as
  `date -u -v-25M` to now. `flowd retime-dispatch <id> <started> <ended> <commit>`
  (`stats/cmd/flowd/main.go`, store `RetimeDispatch`) sets a row's span outright, and
  `flowd restamp-dispatch <id> <parent-transcript>` re-apportions an inline row's figures.
  `flow record dispatch begin` on a key whose row is still open keeps that row's start (the
  `reBegin` clause of `insertDispatch`, `stats/internal/store/records.go`), and `end` writes
  outcome and `ended_at` last-write-wins (`EndDispatch`).
- **4.2.** A Python spike of the hook's pattern over 14 commands matched every expected verdict:
  denied `cat > /tmp/x.go`, `cat >> f`, `cat >f; echo`, `a && cat > f`, `false || cat > f`,
  `(cat > f)`; allowed a heredoc, `cat > f < in`, `x | cat > f`, `cat a > f`, `cat -n f > g`,
  `echo cat > f`, `cat >> f <<EOF`, `concat > f`.
- **5.1.** `flow record passes -change kan-1033-plan-mode-planning` holds one row per round 0–4
  now, so nothing is left to repair. A pass row is `round` plus a free-text `note`
  (`runRecordPass`, `stats/cmd/flow/record.go`); signal five of `check-unfinished-work` needs only
  a row for the round. What a round's `panel-<r>-<slot>` begins know — each slot and its
  model/effort/diff base — is derivable; roster reductions, bounces and diff-size notes are not,
  and stay the parent's `flow record pass` lines.

## Decisions

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-low — the operator's cap: opus at low effort for every flow dispatch, the
self-review's planning and fixers included, until 2026-10-28.
**Considered:** flow-medium, the default when none is recorded — over the cap.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** two groups, in series — group 1 Tasks 1, 2, 3, 5 (the guard, rationale and hook
fixes, each confined to its own files); group 2 Tasks 4 and 6 (the two scripts that start
recording rows, and the live store repair).
**Considered:** one group of six — one fixer would carry the store repair beside four unrelated
fixes; one group per task — six dispatches against the low-effort cap for fixes of a few lines.

### Refuse the bypass in check-untasked-commits

**ID:** untasked-refuses-pipeline-fix-key
**Status:** active
**Chosen:** `check-untasked-commits` lists a commit whose `Flow-Key:` opens `pipeline-fix-` unless
the worktree's branch is itself a `pipeline-fix-<name>-<k>` loop branch — the guard already runs
before every push of a change branch and before run 1's reshape, while the trailer still exists.
**Considered:** a refusal in `check-loop-landing` (`lpFromBranch`), as the pass proposed — it runs
after the reshape squashes the trailer away, and a row-based variant would refuse kan-1033's own
landing on a row no command turns into a loop; a refusal in `flow record dispatch begin` for a
`pipeline-fix-` key under a non-loop change — the parent can skip recording, while the commit is
what lands.

### The rationale is the stale text

**ID:** effort-rationale-is-stale
**Status:** active
**Chosen:** correct the clause in `skills/flow/SKILL-rationale.md` — the Agent tool takes a
dispatch-time `effort`, so the planning texts passing `plan-agent-pair.sh`'s effort stand.
**Considered:** rewrite the planning texts and add read-only `effort: low` Explore/Plan
definitions — built on the rationale's premise, which the Agent tool's own schema disproves.

### Word prefix before the placeholder

**ID:** placeholder-word-prefix
**Status:** active
**Chosen:** the last element of `gsPlaceholder` becomes `[ \t]+[A-Za-z0-9._-]*<`.
**Considered:** dropping the anchor and matching any `<` in the span — would read every `a<b`
comparison in a backticked span as an invocation.

### run-verify-list records the row

**ID:** verify-list-records-row
**Status:** active
**Chosen:** `run-verify-list.sh <worktree> lint|test [<change> <key> <session-token> <model>
<effort>]`; with the five, the script records `begin` before its first command and `end -outcome
completed` after its last when the list exited 0. Both groups run with the same five, so the first
begin stamps the start and the later end the finish; a group that exited non-zero records no end
and the parent ends the row after its re-run course. Kan-1033's rows `verify-fix-1` (7647),
`verify` (7650) and `verify-2` (7651) are retimed to the transcript's instants and restamped.
**Considered:** recording from the `test` group alone — misses a lint group that ran longer;
ending `blocked` from the script on a non-zero exit — the one inline re-run is the parent's, so
the script cannot know the row's outcome.

### Deny a bare cat redirect in the existing Bash hook

**ID:** bare-cat-in-protect-hook
**Status:** active
**Chosen:** the Bash branch of `hooks/protect-main-checkout.py`, already registered on `Bash` by
`setup.sh`, denies a command line where `cat`, at the start of a command, redirects into a file
with no operand, heredoc or input redirect after it.
**Considered:** a new hook registered in `setup.sh` — a new file, registration and unregistered-hook
warning for one pattern.

### panel-round-begins records the pass row

**ID:** round-begins-records-pass
**Status:** active
**Chosen:** after its begins, `panel-round-begins.sh` records one pass row for the round, `-note
'dispatched: <key> on <model>/<effort>[ over <diff-base>]; …'` naming every slot whose begin
succeeded; the parent's own pass lines carry only what no row holds.
**Considered:** deriving the pass row inside `check-unfinished-work` from the reviewer rows — the
guard would stop requiring a pass row at all, and the round's log would still be empty.

## Open questions

## Recommended approach

Six independent fixes, each at the seam the finding's evidence points at, chosen in the decisions
above: `untasked-refuses-pipeline-fix-key` over a `check-loop-landing` refusal,
`effort-rationale-is-stale` over rewriting the planning texts, `placeholder-word-prefix` over an
unanchored match, `verify-list-records-row` over a test-only recording, `bare-cat-in-protect-hook`
over a new hook, and `round-begins-records-pass` over a derived signal. Prose edits ride committed
patches whose sentences `verbatim-moves.txt` lists.

## Critical files

- `stats/internal/guard/untaskedcommits.go` — the untasked-commit verdict (Task 1).
- `stats/internal/guard/check_untasked_commits_test.go` — its tests (Task 1).
- `scripts/check-untasked-commits.sh` — the guard's contract header (Task 1).
- `skills/flow-contracts/git-boundaries.md` — **Commits outside every task** (Task 1).
- `skills/flow/SKILL-rationale.md` — the `flow-<effort>` family's reasoning (Task 2).
- `stats/internal/guard/guardsymlinks.go` — `gsPlaceholder` (Task 3).
- `stats/internal/guard/check_guard_symlinks_test.go` — its tests (Task 3).
- `stats/internal/guard/runverifylist.go` — the verify-list runner (Task 4).
- `stats/internal/guard/run_verify_list_test.go` — its tests (Task 4).
- `scripts/run-verify-list.sh` — the runner's shim and usage line (Task 4).
- `skills/flow/verify-and-handoff.md` — **Recording** under **Inline verify** (Task 4).
- `hooks/protect-main-checkout.py` — the Bash PreToolUse hook (Task 5).
- `stats/internal/repotest/protect_main_checkout_test.go` — its tests (Task 5).
- `stats/internal/guard/panelroundbegins.go` — the round's begins (Task 6).
- `stats/internal/guard/panel_round_begins_test.go` — its tests (Task 6).
- `skills/flow/review-panel.md` — the round's begins sentence (Task 6).

## Existing code to reuse

- `envGit`, `capture` (`stats/internal/guard`) — Task 1's branch read.
- `lookPath`, `kwRun` (`stats/internal/guard/panelroundbegins.go`'s pattern) — the `flow` calls
  of Tasks 4 and 6, and the PATH-stub test pattern of `TestPanelRoundBeginsRecordsLiteralKeys`.
- `flowd retime-dispatch` and `flowd restamp-dispatch` (`stats/cmd/flowd/main.go`) — Task 4's
  store repair.
- `bash_hits`'s caller in `main()` of `hooks/protect-main-checkout.py` — Task 5 adds its deny
  beside the landing deny.

## Verification

Each task's own verify step, then from `<agents-wt>/stats`: `gofmt -l .` prints nothing,
`make -s web-dist && go vet ./...` exits 0, and `go test ./internal/guard ./internal/repotest
-count=1` passes; from `<agents-wt>`: `scripts/check-verbatim-moves.sh`,
`scripts/check-guard-symlinks.sh`, `scripts/check-references.sh` and
`scripts/check-markdown-integrity.sh` exit 0.

## Live check

From this worktree with every task committed, before landing; the protected `flowd` on
`127.0.0.1:4173` and `flow-postgres` are never stopped or restarted:

1. **Rows repaired.** `flow record dispatches -change kan-1033-plan-mode-planning -C
   /Users/tweety53/Projects/agents` — before: `verify` 19:01:46–19:26:46 +02:00 enclosing
   `verify-fix-1` and `verify-2`; after: `verify-fix-1` 19:06:21–19:08:18, `verify`
   19:00:56–19:06:17, `verify-2` 19:08:20–19:13:38, no two overlapping. Not working looks like
   `verify` still ending 19:26:46.
2. **Verify row recorded.** Stand-in for a live run: `TestRunVerifyListRecordsTheVerifyRow`
   against a `flow` stub; not working looks like no `record dispatch begin` line in the stub's log.
3. **Pass row recorded.** Stand-in: `TestPanelRoundBeginsRecordsThePassRow` against a `flow` stub;
   not working looks like no `record pass` line after the begins.
