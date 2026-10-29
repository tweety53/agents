# flow-interactive: the 2026-10-29 stats-board session

The session that motivates KAN-1061 (`/flow-interactive`). One Claude Code session in the `agents`
repository, run without `/flow`, reworking the stats app's Live state board and run view while a
real `/flow` run (gymie `kan-1059-repeat-row-keep-awake-calendar-swipe`) was live on it. The
operator watched the board and asked for one small change at a time; the agent built, checked and
pushed each to `main`, then rebuilt and restarted `flowd`. The session transcript is
`~/.claude/projects/-Users-tweety53-Projects-agents/488910d1-c0d4-4c19-9e9a-f55c35d98878.jsonl`
(local to the operator's machine, ~7 MB).

## The working loop that made it fast

Per request, in one turn:

1. Read the code the request touches (grep the component, the store query, the DTO), and query the
   dev store (`docker exec flow-postgres psql …`) for the live rows behind what the operator saw.
2. `git worktree add ../agents-<slug> -b <branch> origin/main` — the main checkout is read-only.
3. Edit with small scripted replacements; add or update one test per behaviour.
4. Targeted checks only: `npx tsc -b && npx vitest run` for the SPA (with `web/node_modules`
   symlinked from the main checkout), `go vet ./...` plus `go test` of the touched packages only
   (`./internal/store/ -run <Test>`, `./internal/api/`), `gofmt -l`.
5. Commit, `git fetch && git rebase origin/main`, `git push origin <branch>:main` as a standalone
   git command, `git pull --ff-only` in the main checkout, remove the worktree.
6. `./setup.sh global && make -C stats restart`, then reload the page (Playwright) when the
   request was visual.

Mid-turn operator messages were folded into the turn in flight, never treated as a stop.

## Requests, in order, and what shipped

| # | Operator request | Shipped |
|---|------------------|---------|
| 1 | Why no effort in stage runs? | Harvester never wrote `metrics.effort` for a stage run; record the main session's effort (`Delta.Effort`, main-thread records only). |
| 2 | Rename "Running now" → "Active runs", show unfinished changes, always show the current stage | Active runs lists every non-FINISHED change; store exposes `lastStage`/`lastStageEndedAt`. |
| 3 | Input tokens 0.00M is useless | Count input + cache reads + cache writes (`readTotalInputTokens`), hint "cache reads N%". |
| 4 | What does "Errors" mean? → rename | "Tool errors" with a tooltip. |
| 5 | Remove "next" from cards; `flow.decide` makes no sense | Plain-words stage names (`web/src/lib/stageLabel.ts`). |
| 6 | Show the task list and active subagents with model-version-effort; short on the board, detailed in the run | Store: open dispatches, repo roots, committed tasks; API reads `tasks.md` from the change's worktree; run view Tasks + Agents panels; card count. |
| 7 | That one is the main session, not a subagent | `agent_id = 'inline'` marks main's own dispatches. |
| 8 | Call it "agents", show main as "main"; make it obvious which is which | Agents running table, Main row first, subagents indented with `⎿`. (A boxed MAIN/SUBAGENT tag first — "ugly" — replaced.) |
| 9 | Stage names plain in the table and timeline, id below or on hover | Stage table name + muted id; timeline lane `<title>`; cost rail. |
| 10 | Expect two rows: main waiting and the subagent | Main row whenever a stage is open; status dots: running (green pulse), waiting for subagents (amber), needs you (red blink), turn ended (grey ring). |
| 11 | Jira for visual-verify cost | KAN-1060. |
| 12 | No agents shown while main works | Main stays listed between stages (session activity of the latest run). |
| 13 | Code review cost missing | `claude-sonnet-5-5` had no pricing row, so every run that used it had no cost; seeded it from the pricing page and repriced since 2026-09-01. |
| 14 | In-run fixes not shown, on both pages | Card and Now running name the running task ("Task 6: In-run fix 1 — …") between stages. |
| 15 | Remove the tool errors column | Removed with its filter chip. |
| 16 | "Paused after …" while the session runs full speed; it must look as alive as the terminal | Session alive = transcript written in the last 90 s (`harvest_offsets.updated_at` of the session's files); dispatch keys described in plain words (`describeDispatch`). |
| 17 | Text overlap in the Now running tile | KPI grid cells `min-width: 0`, ellipsis. |
| 18 | Stage runs show no rows for the fix work | A background subagent: the store opens a stage run when a dispatch begins with none open (new `flow.pipeline-fix` stage, `stage_runs.auto_opened`, migration 0053), closes it with its last dispatch, and an explicit stage begin supersedes it. ~300 fix dispatches in 14 days had no stage. |
| 19 | "Code review round 2: primary" is off | Name only a plan task beside an open stage. |
| 20 | Current stage isn't code review, the agents are done | "agents done, main wrapping up" when every dispatch the open stage began has ended. |
| 21 | "Working" badge repeats the dot | Badge hidden for `working`. |
| 22 | Why isn't kan-1059 IN_PROGRESS? | By design `IN_PROGRESS` = staged for review; the board now shows a derived phase: Planning / Implementing / Ready for review / Finished. |
| 23 | Main's model lists the reviewer's model too | Main row reads only models with main-thread tokens (`readMainModels`). |

## What went wrong — each needs a step or a guard in `/flow-interactive`

- **A worktree was force-removed before its commit landed.** A `&&` chain failed at `go vet`, the
  commit never ran, the push pushed nothing, and the next command removed the worktree with
  `--force`, losing the edits; they were redone from scratch. Land only after confirming the
  commit exists on the remote (`git log origin/main -1` contains it), and never `--force` a worktree
  with uncommitted changes.
- **`make web-dist` refuses a symlinked `web/node_modules`** (kan-1037), so the Go build failed with
  `pattern all:dist: no matching files found` after a symlink left in place for vitest. One
  documented order: build with `make web-dist` first (it runs `npm ci`), or remove the symlink
  before it.
- **A struct-field insert silently missed** because `gofmt` had realigned the anchor line; the
  replacement script asserted nothing. Every scripted edit asserts its anchor matched.
- **A `python3 -` heredoc with a non-ASCII em dash inside an escaped string failed to parse**, and
  the whole edit batch was skipped. Scripts go to a file in the scratchpad, not inline.
- **The browser showed the old bundle after a restart**; a hash-only navigation does not reload.
  Reload (`location.reload()`) before every visual check.
- **The first visual check hit the wrong URL** (project display name instead of the project key),
  so the page loaded no stage runs and the agent briefly chased a non-bug. Open the run page by the
  board's own link or by `project_key`.
- **UI was shipped without looking at it** until the operator said "ugly". A UI request is not
  done until a screenshot of the live page was read.
- **`make -C stats restart` exited 2 once** (port race) and the next check ran against a dead
  daemon. Check the restart's exit code and that `:4173` answers.
- **The session-activity hook was installed but never registered** in `~/.claude/settings.json`,
  so the board could never show "needs you"; the auto-mode classifier blocked the agent's edit and
  the operator had to grant it. `/flow-interactive` checks the hook registrations on start and asks
  once.
- **Answering by assumption**: twice the first answer to "why don't I see X" came from reading code
  before reading the live rows. Query the store for the rows behind the screen first.

## What `/flow-interactive` keeps and what it adds

Keeps: one request → one commit on `main` within minutes; no brainstorm gate, no spectre
artifacts, no review panel per request; the operator's next message is the next request.

Adds: a change in the stats store for the session (or per request) with stage marks per request —
at least implement, verify, land — and every dispatch recorded, so the run is visible on the board
it edits; the project's lint and guard commands from `.flow/project.md` and targeted test re-runs
of only the touched packages and files; a failing test first for each behaviour change; visual
verification for UI requests; the reinstall after each landing; the failure list above as steps.
