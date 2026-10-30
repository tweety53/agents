# kan-1061-flow-add-flow-interactive-fast-conversational

## Why

- An operator watching a live page asks for one small change at a time and wants each one shipped
  within minutes. `/flow` (plan gate, panel, archive) and `/flow-fast` (one change per invocation,
  one worktree per change) both cost more per ask than the ask is worth.
- The 2026-10-29 stats-board session (`docs/briefs/flow-interactive-session-2026-10-29.md`) ran
  that loop by hand, outside every pipeline: fast, but unrecorded in the store, and it paid for
  failures no step or guard caught — a worktree force-removed before its commit landed, UI shipped
  without looking, an old bundle screenshotted after a restart, a page opened by a display name
  instead of its project key, a failed `make -C stats restart` read as success, a session-activity
  hook never registered, answers given from code without the live rows behind them.
- KAN-1061 asks for a conversational command that keeps that loop's speed, records every request's
  lifecycle in the store well enough to analyse it, shows the session on the live board, and turns
  the brief's failure list into steps and guards.

## What changes

- A new command, `/flow-interactive`: one invocation is one session — one store change, one Jira
  issue (filed at start unless declined, Done at the explicit stop with a session summary
  appended), one git worktree — and a stream of requests, each landed on the base branch on its own
  by merge and push. The main session does each request inline; no review panel per request, no
  spectre artifacts, no plan file.
- Every request runs Look → (rarely) Brainstorm → Implement → Verify → Land → Reinstall → UI check
  (when the diff touches `ui paths`) → End, each step a stage run carrying the request number.
  Mid-turn messages queue as the next request; only an explicit stop ends the session, running one
  self-review pass over the whole session, cleanup and `FINISHED`. A session is resumable by Jira
  key or by change name.
- flowd stores requests (`requests` table, `stage_runs.request_n`); `flow record request
  begin|end` records them, journalled when the store is unreachable; `flow stage begin -request
  <n>` tags a stage run with its request.
- `flow record request end -outcome landed` refuses a commit that is not on `origin/<base>`, and,
  when the request's diff matches the project's `ui paths`, a request whose latest live-page UI
  check did not pass.
- `/flow-interactive` joins the stage vocabulary on eleven existing `flow.*` keys; its
  `flow.write-finished` moves the change to `FINISHED`.
- The live state board shows an interactive session as the approved prototype draws it: an
  `Interactive` pill, `Request <n>: <title>` as the running line, `requests <ended>/<total>`,
  `Waiting for the next request` between requests, a Requests panel in the Tasks slot, the
  per-request Main steps and an `R<n>` column in Stage runs.
- The self-review bundle renders a session's requests and their landed commits in place of the
  (empty) session branch log.
- Hooks: the agent-baseline hook's begun-row check reads the session's latest stage run of any
  command — closing a fail-open for `/flow-fast` dispatches — and recognises `/flow-interactive`;
  the prompt hook reminds an unfinished interactive session that a plain message is its next
  request or its stop.
- The command is registered everywhere the other commands are: its skill and command file, the
  skill indexes, the Jira contract's Resolution, Transitions and Description lines, the pipeline's
  Command surface, and the reinstall rule after each landed request.

## Fix 1 — in-run verify

- `flow.verify`'s lint list failed `scripts/check-installed-citations.sh`: `skills/flow-interactive/SKILL.md:275: citation `.superpowers/` names no root`. The resume clause added by the task-12 review Minor cites the gitignored directory without a root; it is rooted as `<project>/.superpowers/`.

## Fix 2 — in-run visual verify

- `flow.visual-verify` found the interactive run view departing from the approved prototype (`.superpowers/sdd/planner/prototype-board.html`): the Requests panel sits narrow beside Agents/Main instead of in a right column with Cost by stage (`.work { grid-template-columns: minmax(0, 3fr) minmax(0, 2fr) }`); the Agents Status cell prints "running" beside the dot where the prototype draws the dot alone; Stage runs shows 5 rows and a "Show all" chip where the prototype shows 6 rows and no chip. The four card captures in `interactive.spec.ts` are clipped element shots where verification needs viewport captures.
