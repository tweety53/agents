# Design — kan-384-flow-dispatch-cost-still-unmeasured-for-most

## Context

Bounded change; no separate spec file. `proposal.md` carries why; this file carries how.
`flow stage begin` (`stats/cmd/flow/stage.go`) already has a `-session` flag whose value goes out as
`BeginStageRequest.SessionID`; the API (`stats/internal/api/stages.go`, `stageBeginRequest.SessionID`)
and the store pass it through to `stage_runs.session_id`. `UnresolvedSessionTokens` selects
`WHERE session_id IS NULL`, `BindSession` updates only such rows, and `DispatchWindowsForSession`
resolves dispatches by `stage_runs.session_token → session_id`. A row inserted with `session_id`
set therefore needs nothing downstream. The variable's value was verified equal to the transcript
`sessionId` from a subagent's Bash call in session `6af6b60e-…` (it is the root session's id in a
subagent, which is also what every subagent transcript line's `sessionId` carries).

## The bind

In `stage.go`'s `begin` path, where `sessionID` is derived from `sessionFlag`:

```go unverified:the surrounding lines are stage.go:309-311 at main (9b6619e); the env name is what this session's Bash environment exposes
var sessionID *string
switch {
case *sessionFlag != "":
	sessionID = sessionFlag
case os.Getenv("CLAUDE_CODE_SESSION_ID") != "":
	v := os.Getenv("CLAUDE_CODE_SESSION_ID")
	sessionID = &v
}
```

Precedence: `-session` flag > `CLAUDE_CODE_SESSION_ID` > nil. `stage end` is untouched —
attribution happens once, at `begin`. `flow record dispatch` is untouched — dispatches carry no
`session_id`; they resolve through the stage run's token.

## Fallback

A mark made where the variable is unset (Cursor, Codex, a shell outside the harness) sends no
`sessionId`, and KAN-172's transcript search runs exactly as today. The literal-token contract in
`skills/flow-contracts/pipeline.md` stands unchanged for that reason and because the token is the
dispatch↔stage-run join key regardless of how the session was bound.

## Files

`stats/cmd/flow/stage.go`, `stats/cmd/flow/stage_test.go`, `skills/flow-contracts/pipeline.md`
(one sentence in the session-token paragraph), `stats/README.md` (one sentence under "Checking
attribution by hand").

## Testing

Two new tests in `stats/cmd/flow/stage_test.go`, shaped like
`TestStageBeginCannotDetectShellExpandedSessionToken` (a `genuineDaemon` capturing the request body):
one sets `CLAUDE_CODE_SESSION_ID` via `t.Setenv` and expects the body's `sessionId` to equal it;
the other sets both the variable and `-session` and expects the flag's value. Existing tests that
must keep seeing no `sessionId` clear the variable with `t.Setenv("CLAUDE_CODE_SESSION_ID", "")`
where the harness's own environment would otherwise leak in.
`cd stats && gofmt -l . && go vet ./... && go test ./cmd/flow/... -race -count=1`, plus the prose
guards `.flow/project.md`'s `## lint` names for the two Markdown edits.

## Decisions

### Bind from `CLAUDE_CODE_SESSION_ID` at the mark, not by fixing the transcript match

**ID:** env-bind-at-stage-begin
**Status:** active
**Chosen:** `flow stage begin` defaults `-session` from `CLAUDE_CODE_SESSION_ID` — one `Getenv`,
every downstream pass already handles a pre-bound row, and the whole failure class (substituted
tokens, the 60-cycle window, session ambiguity) disappears on Claude Code.
**Considered:** teaching `isSessionMarkCommand` to resolve `$VAR` from a same-command `VAR=…`
assignment — fixes only that one shape, keeps the harvest race and give-up window, more code in
the harvester. A `PreToolUse` hook denying `-session-token $…` — a new installed hook to police a
writer the env bind makes irrelevant.

### Historic unbound rows are left as recorded

**ID:** no-backfill
**Status:** active
**Chosen:** forward-only; `session never bound` on past rows is the honest record of what the
pipeline could measure then.
**Considered:** a one-off backfill re-binding given-up tokens by finding `TOKEN=<tok>` assignments
in transcripts — harvester code for a one-time repair, and every such row's dispatch windows would
then be attributed from transcript offsets the store has already consumed past.

## Open questions

None recorded.
