# kan-384-flow-dispatch-cost-still-unmeasured-for-most

Jira: KAN-384.

## Why

- KAN-380's ledger reads `session never bound` on all 17 dispatch rows and all 28 stage runs of
  its run carry `session_id` NULL, although KAN-172 (bind a stage run's session from a literal
  token found in the transcript) and KAN-212 (persist per-dispatch cost) both landed.
- Root cause, verified in the store and the transcript: the session batched its marks into one
  Bash call as `TOKEN=mf-380a` … `flow stage begin … -session-token $TOKEN`. The transcript records
  the command before the shell expands it, so `internal/harvest`'s `isSessionMarkCommand` compares
  the field `$TOKEN` against the literal `mf-380a` and never matches; after 60 harvest cycles the
  token is given up and every dispatch under it is stamped `session never bound`. The CLI's
  `validateSessionToken` cannot see the substitution — argv is already expanded.
- Not one run's mistake: roughly 600 of the 6,600 `-session-token` invocations in this machine's
  transcripts use the `$VAR` shape, and unbound stage runs per day since 2026-08-27 are
  20, 85, 44, 18, 14, 65. Every cost-angle self-review of an affected change is a none-marker by
  construction.

## What changes

- **`flow stage begin` binds the session at the mark.** Every Bash tool call in Claude Code —
  parent and subagent alike — carries `CLAUDE_CODE_SESSION_ID`, equal to the transcript's own
  `sessionId`. When `-session` is empty and that variable is set, the mark sends it as the stage
  run's `sessionId`. The row is born bound; the harvester's transcript search, its 60-cycle
  window and its give-up never engage for it, and `DispatchWindowsForSession` attributes the
  run's dispatches through the token join it already uses.
- **Transcript-token binding stays as the fallback** for a mark made without the variable. No
  harvester, API or store change; the literal-token rule stands — the token is still the
  dispatch↔stage-run correlator.
- **Two sentences of documentation**: the session-token paragraph in
  `skills/flow-contracts/pipeline.md` and "Checking attribution by hand" in `stats/README.md` name
  the env bind.
- **Historic unbound rows stay as they are** — the fix is forward-only.
