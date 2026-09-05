# kan-384-flow-dispatch-cost-still-unmeasured-for-most

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.

> **Relocation:** no

Two tasks. Task 1 is the fix: `flow stage begin` defaults `-session` from `CLAUDE_CODE_SESSION_ID`,
with two tests. Task 2 names the bind in the two documents that describe session binding.
`design.md` is canonical for the precedence, the fallback and the decisions; `proposal.md` for the
root cause. Every Go command below runs from `stats/`; every guard from the repo root.

**Baseline, measured before any edit:**

- `stats/cmd/flow/stage_test.go`: 11 tests.
  <!-- measured: grep -c '^func Test' stats/cmd/flow/stage_test.go @ branch main (9b6619e) -->
- `go test ./cmd/flow/` passes at main.
  <!-- measured: cd stats && go test ./cmd/flow/ @ branch main (9b6619e) -->
- `skills/flow-contracts/pipeline.md` is 24746 bytes against a `check-contract-budget.sh` row of
  36155; task 2 adds one sentence, so the row holds.
  <!-- measured: wc -c skills/flow-contracts/pipeline.md against its row in scripts/check-contract-budget.sh @ branch main (9b6619e) -->
- In a Claude Code Bash tool call, `CLAUDE_CODE_SESSION_ID` equals the transcript's `sessionId`
  (`6af6b60e-72da-4615-8e3f-75d37bdf8f9d` in both, observed from a subagent of that session), and
  is set in subagents as the root session's id.
  <!-- measured: echo $CLAUDE_CODE_SESSION_ID vs jq -r .sessionId on ~/.claude/projects/-Users-tweety53-Projects-agents/6af6b60e-72da-4615-8e3f-75d37bdf8f9d/subagents/agent-a7e4487ed26937f39.jsonl — a machine-local observation, not re-runnable at a ref -->
- `stage_test.go` already runs under a leaked `CLAUDE_CODE_SESSION_ID` when `go test` is invoked
  from a Claude Code session; no existing test asserts `sessionId` is absent from the request body,
  so the leak fails nothing today. The one existing `sessionId` assertion is
  `TestStageBeginRecordsIdentityAndInstant`, which passes `-session sess-123` explicitly.
  <!-- measured: grep -n 'sessionId' stats/cmd/flow/stage_test.go @ branch main (9b6619e) -->

- [x] 1. `flow stage begin` defaults `-session` from `CLAUDE_CODE_SESSION_ID`

**`stats/cmd/flow/stage.go`**, the `begin` path where `sessionID` is derived from `sessionFlag`
(the three lines `var sessionID *string` / `if *sessionFlag != "" {` / `sessionID = sessionFlag`
just above `req := client.BeginStageRequest{`):

```go unverified:confirm the three replaced lines read exactly as quoted above at the branch tip; the env name is the one this repository's Claude Code sessions export
var sessionID *string
switch {
case *sessionFlag != "":
	sessionID = sessionFlag
case os.Getenv("CLAUDE_CODE_SESSION_ID") != "":
	v := os.Getenv("CLAUDE_CODE_SESSION_ID")
	sessionID = &v
}
```

Update the `-session` flag's help string (`"the harness session id, if known"`) to say it
defaults to `CLAUDE_CODE_SESSION_ID` when set. Add a doc comment above the switch, in the
file's own register, stating why: the transcript records a mark's command text before the shell
expands it, so a token written as `$TOKEN` is never found by the harvester's search
(`internal/harvest`'s `isSessionMarkCommand`); a row born with its `session_id` set skips that
search entirely (`store.UnresolvedSessionTokens` selects `session_id IS NULL`), and the token
search remains the fallback for a mark made where the variable is unset. Nothing else in
`stage.go` changes: `stage end` carries no session, and `validateSessionToken` stays as it is.

**`stats/cmd/flow/stage_test.go`**, two new tests beside
`TestStageBeginCannotDetectShellExpandedSessionToken`, each using the same `genuineDaemon`
body-capturing shape:

- `TestStageBeginDefaultsSessionFromClaudeCodeEnv`: `t.Setenv("CLAUDE_CODE_SESSION_ID",
  "6af6b60e-72da-4615-8e3f-75d37bdf8f9d")`, run `stage begin` without `-session`, decode the
  captured body, expect `got["sessionId"]` to equal that value.
- `TestStageBeginSessionFlagWinsOverClaudeCodeEnv`: same `t.Setenv`, run with `-session
  sess-flag`, expect `got["sessionId"] == "sess-flag"`.

Both tests also assert exit code 0 and that `sessionToken` in the body is the literal passed —
the env bind changes nothing about the token.

  - [x] **Step 1: Write the two tests; run them and see them fail on `sessionId` absent / wrong.**

```bash verified:the commands stats/README.md and .flow/project.md name for this module
cd stats && go test ./cmd/flow/ -run 'TestStageBegin(DefaultsSessionFromClaudeCodeEnv|SessionFlagWinsOverClaudeCodeEnv)' -count=1
```

`TestStageBeginDefaultsSessionFromClaudeCodeEnv` fails (no `sessionId` in the body);
`TestStageBeginSessionFlagWinsOverClaudeCodeEnv` passes already and stays as the guard on
precedence.

  - [x] **Step 2: Apply the `stage.go` edit above; run the package.**

```bash verified:the commands stats/README.md and .flow/project.md name for this module
cd stats && gofmt -l . && go vet ./... && go test ./cmd/flow/... -race -count=1
```

`gofmt -l .` prints nothing; the package passes with 14 tests in `stage_test.go`.
<!-- predicted: grep -c '^func Test' stats/cmd/flow/stage_test.go after task 1 -->

  - [x] **Step 3: Check it live, once, against the dev daemon.** With `flowd` running, from a
    Claude Code Bash call in the worktree, run one `flow stage begin` / `flow stage end` pair on
    this change with a fresh literal token and query the row:

```bash verified:copied from stats/README.md "Checking attribution by hand", the query shape it names
curl -s 'http://127.0.0.1:4173/api/v1/stage-runs?project=<project-key>&name=kan-384-flow-dispatch-cost-still-unmeasured-for-most&sort=-started_at&limit=1' | jq '.stageRuns[0] | {sessionId, sessionToken}'
```

`sessionId` is the session's id immediately, before any harvest cycle. Note the observed
`sessionId` in the commit body.

  - [x] **Step 4: Commit.**

**Files:** `stats/cmd/flow/stage.go`, `stats/cmd/flow/stage_test.go`
**Tests:** `TestStageBeginDefaultsSessionFromClaudeCodeEnv`,
`TestStageBeginSessionFlagWinsOverClaudeCodeEnv`, `TestStageBeginOmitsSessionWhenEnvUnset`
**Regression:** reverting this commit makes `TestStageBeginDefaultsSessionFromClaudeCodeEnv`
fail — the request body carries no `sessionId` — and every Claude Code stage run goes back to
binding only through the transcript token search, which a `-session-token $TOKEN` mark defeats.
`TestStageBeginOmitsSessionWhenEnvUnset` pins the companion case: with `-session` absent and
`CLAUDE_CODE_SESSION_ID` unset/empty, `sessionId` must be absent from the request body, not sent
as an empty string.
**Baseline:** before=11 after=14
<!-- predicted: grep -c '^func Test' stats/cmd/flow/stage_test.go after task 1 -->
**Commit:** `fix(flow): bind a stage run's session from CLAUDE_CODE_SESSION_ID at begin`
**Build:** green

- [x] 2. Name the env bind in the pipeline contract and the stats README

Prose only, one sentence each; `design.md`'s **The bind** and **Fallback** are canonical.

**`skills/flow-contracts/pipeline.md`**, the paragraph ending "`-harness` names the harness
actually running the mark — `claude-code`, `cursor` or `codex`." Append one sentence: on Claude
Code the mark also binds the stage run's session directly from the `CLAUDE_CODE_SESSION_ID` the
harness exports to every Bash call, so the token's transcript search is the fallback for a mark
made without it — the literal-token rule stands because the token is still the join key between a
change's dispatches and its stage runs. Do not reword the literal-token sentences that follow.

**`stats/README.md`**, "Checking attribution by hand", the paragraph beginning "Expect
`sessionId` to be a real session id, not `null`": add one sentence that on Claude Code
`sessionId` is set by the mark itself from `CLAUDE_CODE_SESSION_ID` and is present before the
first harvest cycle; only `metrics` waits on the harvester.

  - [x] **Step 1: Edit the two paragraphs.**
  - [x] **Step 2: Run the prose guards.**

```bash verified:the guards .flow/project.md's lint list names for skill and README prose
scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-contract-budget.sh \
  && scripts/check-installed-citations.sh && scripts/check-markdown-integrity.py \
  && scripts/check-stage-mark-calls.sh && scripts/check-dispatch-paragraphs.sh
grep -c 'CLAUDE_CODE_SESSION_ID' skills/flow-contracts/pipeline.md stats/README.md
```

Every guard exits 0; the `grep` reports 1 for each file.

  - [x] **Step 3: Commit.**

**Files:** `skills/flow-contracts/pipeline.md`, `stats/README.md`
**Tests:** none
**Regression:** reverting this commit leaves the contract describing transcript search as the
only way a stage run's session is bound, while task 1's CLI binds it at the mark — two accounts of
one mechanism.
**Baseline:** before=13 after=13
<!-- predicted: grep -c '^func Test' stats/cmd/flow/stage_test.go, unchanged by a prose-only task -->
**Commit:** `docs(flow-contracts): name the CLAUDE_CODE_SESSION_ID bind at stage begin`
**Build:** green
