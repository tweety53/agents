# stats

The flow stats app: a PostgreSQL-backed Go service that replaces the
per-machine JSON state file, records per-stage telemetry, and serves both the
live pipeline state and aggregated statistics through a browser interface.

Design notes live in the source: each package's doc comment states the requirement it serves.

## Prerequisites

- Go 1.26.5 (`go version` to confirm).
- Docker, for the dedicated PostgreSQL stack in `docker-compose.yml`.

## The PostgreSQL stack

A dedicated `flow-postgres` container, independent of any other Postgres
stack already running on this machine (for example a `gymie-postgres`
container on port 5432). It listens on host port **5433** so the two never
collide — confirmed free on this machine before the port was chosen.

```bash
cd stats
docker compose up -d
```

This starts only `flow-postgres`; it does not stop or modify any other
running container.

## Hybrid self-review similarity (optional)

`flow self-review similar` ranks by BM25; with an embedding sidecar and a pgvector
database it fuses BM25 with embedding similarity (reciprocal rank fusion). Both
`FLOW_EMBED_URL` and `FLOW_PGVECTOR_DSN` set means hybrid; either unset means BM25
only. If the embedding half fails, the CLI prints a warning and answers BM25 only.

`make restart` and the flowd launchd agent set both variables for the live daemon. The sidecar has its own
agent, `stats/launchd/com.tweety53.flow-embedder.plist` (copy it to `~/Library/LaunchAgents/` and
`launchctl load` it, like the flowd one); the container restarts itself (`restart: unless-stopped`). The
commands below are for running them by hand without the agents.

```bash
docker compose -f stats/docker-compose.pgvector.yml up -d
uv run stats/embedder/server.py
FLOW_EMBED_URL=http://127.0.0.1:8099 FLOW_PGVECTOR_DSN='postgres://flow:flow@localhost:5434/flowvec?sslmode=disable' flowd
```

## Running the tests

```bash
cd stats
make -s web-dist   # a fresh checkout has no built SPA to embed
go test ./...
```

Tests that need the compose stack's PostgreSQL skip themselves with a clear
message when the stack is not running, so `go test ./...` is always safe to
run without Docker.

**The UI-test stack below is not needed for the Go test suite.** `go test ./...` already
gets its own isolation: every test that touches Postgres creates its own
database (see `internal/store/testsupport_test.go`), used once and never
shared with the live database, another test, or the UI-test stack below.
Nobody running `go test` needs to bring up the UI-test stack first.

## flow-guard

`flow-guard` (`cmd/flow-guard`) runs the Go ports of the repository's guard
scripts. Every guard is called as `flow-guard <name>`, which resolves on `PATH` to
the bootstrap `scripts/flow-guard`; it runs the binary built from its own checkout:
it hashes the sources `flow-guard` is built
from, builds on the first call for a hash into
`${XDG_CACHE_HOME:-$HOME/.cache}/flow-guard/<hash>/flow-guard`
(`FLOW_GUARD_CACHE_DIR` overrides the directory), and execs the cached binary
after that. Nothing is installed. With no `go` to build it, or a failed build,
the shim exits that guard's own cannot-answer code — 4 for `run-reproducer`,
2 for the others — and names the cause.

Every guard `flow-guard` dispatches is also recorded in flowd's `guard_runs`
table (`POST /api/v1/guard-runs/{project}`): the guard, its exit code
classified as `clear`, `fired` or `cannot-answer`, and its runtime, and — on a run that fired or could not answer — the distinct reason ids the guard recorded, in first-seen order, in its `reasons` column (NULL on a clear run); `SELECT guard, unnest(reasons) AS reason, count(*) FROM guard_runs GROUP BY 1, 2 ORDER BY 3 DESC` ranks them. That is
what the SPA's **Flow guards** view reads to rank guards by fires, never-fired
and run time, leaving out flow-guard's helper commands (`records.GuardHelpers`). The
write is best-effort: it goes to `FLOW_RECORDS_ADDR`, then `FLOW_ADDR`, then
the dev daemon, is bounded at 500 ms, never changes a guard's exit code, and
is dropped when the daemon is down or does not know the project (a guard run
inside a test fixture's repository). `FLOW_GUARD_TELEMETRY=off` turns it off.

### Heavy-run slots

A `run-verify-list` `test` or `visual` list and a `recapture` each hold one of
flowd's heavy-run slots for as long as they run, so at most `FLOWD_HEAVY_SLOTS`
(default 2; anything but a positive integer refuses startup) run at once across
every worktree. The slots live in flowd's memory (`internal/api/slots.go`):
`POST /api/v1/slots/acquire`, `POST /api/v1/slots/release`, and
`GET /api/v1/slots`, which lists the holders. A holder whose pid is gone, or
that went 60 seconds without renewing, is dropped. A guard that cannot reach
flowd runs without a slot (`internal/guard/heavyslot.go`).

## The UI-test stack

For testing the application's browser interface by hand, from the main
checkout, against real data that is not the operator's own — separate from
both the live daemon (port 4173, database `flow`) and any `/flow`
worktree's own isolated stack (see `.flow/project.md`'s
`## workspace isolation`).

```bash
cd stats
make ui-test-up
```

This drops and recreates the `flow_uitest` database in the same
`flow-postgres` container the live stack uses (host port 5433), starts
`flowd` on port **4174** against it, waits for the daemon to answer, and
seeds it with a fixed fixture — two projects, changes spanning `STARTED`,
`IN_PROGRESS` and `FINISHED`, and stage runs carrying token usage, so the
views render something rather than an empty interface.

**Every `make ui-test-up` resets the stack.** The database is dropped and
recreated from scratch each time, never reused from a previous session —
that is what keeps its contents known at the start of every session,
rather than accumulating the way the live database did.

Point a session or a browser tab at it:

```bash
export FLOW_ADDR=http://127.0.0.1:4174
```

`FLOW_ADDR` overrides the `flow` CLI's default daemon address (see "A
single variable targets the test stack" in this change's spec); opening
`http://127.0.0.1:4174` directly in a browser reaches the same daemon's
SPA. The live daemon on 4173 and the live database are untouched by
either bring-up or by anything written during the test session.

**`export` persists for the rest of that shell session, not just the next
command.** Once exported, *every* subsequent `flow state`/`flow stage`
invocation in that shell — including ones run much later, unrelated to UI
testing — silently targets the test stack on 4174 instead of the live
daemon on 4173, until the shell exits or the variable is unset. No
`/flow*` skill passes `-addr` explicitly, so nothing overrides this, and
a successful `state set` exits 0 with no warning either way — there is no
signal that a write landed on the test stack instead of the live one. To
avoid that:

- Prefer scoping it to one command instead of exporting it:
  `FLOW_ADDR=http://127.0.0.1:4174 flow state get <name>`, or
- Run the UI-test session in a separate subshell (`bash`, then `export`
  inside it, then `exit` when done), or
- If you do `export` it in your main shell, `unset FLOW_ADDR` as soon as
  you are done testing, before running any real `/flow*` command.

Tear it down when done:

```bash
make ui-test-down
```

This stops the daemon and drops `flow_uitest`, both idempotently — a
second `make ui-test-down` is not an error.

## Agent roles

Every role's allowed pairs live in flowd's store, seeded once by migration
`0061_role_pairs.sql`; `skills/flow-contracts/model-policy.md`'s **Allowed pairs** section is
canonical for what the tiers mean. `flow policy get` prints them (and caches them for when flowd is
down), `flow policy grants|grant|revoke -change <name> [-roles …]` manage the operator grants that
admit opt-in pairs, and flowd serves them on `GET /api/v1/policy`,
`PUT /api/v1/policy/roles/{role}`, `POST /api/v1/policy/roles` (several roles in one edit) and
`GET`/`PUT`/`DELETE /api/v1/policy/grants`.

flowd serves the current roles with their pairs on `GET /api/v1/roles` — the role ids, groups and
`Does` lines embedded from `internal/modelpolicy/embedded/model-policy.md`, which
`make policy-copy` (run by `make web-dist`) copies from the repository at build time. The stats
SPA's **Agent roles** view draws that table, grouped by the pipeline stage each role runs in, each
role's base and opt-in pairs editable and saved straight to the store.

## Role costs

`GET /api/v1/stats/role-costs?from=&to=[&project=]` folds every dispatch started in the period by
the allowed-pairs role it ran as — `api.RoleOf` derives it from the stored role and the dispatch
key, a key no rule places counting as `other` — into its dispatches, cost, summed agent time
(parallel agents add up), median run, failed count and model · effort split, grouped by pipeline stage and costliest first within one, with
the weekly spend per role. Roles are derived at read time; no stored row changes. Each headless model call a guard makes -- `pair-judge`'s
judge and `plan-check`'s decider -- is posted with that guard's run (`guard_runs.headless_*`,
migration `0082_guard_run_headless.sql`) and counted as one run of its role; calls made before
that migration were never stored. The stats SPA's
**Role costs** view draws it.

## Bottlenecks

`flow stats bottlenecks [-last N] [-project key] [-json]` reads
`GET /api/v1/stats/bottlenecks?last=N[&project=]` and ranks where the last N changes (default 20,
ordered by their latest stage run) spent wall-clock time, from rows already recorded: ended
`stage_runs` by stage key, ended `dispatches` by role and by `dispatch_class(role)`, `guard_runs` by
guard (claimed by worktree as the per-run page claims them) — a run's work, `duration_ms` less
`wait_ms` (the time it spent on a heavy slot or blocked in `await-rows`, migration
`0112_guard_run_wait.sql`), the wait ranked apart under `guard-wait` and the `scripts/flow-guard`
bootstrap before the run (`bootstrap_ms`, `0113_guard_run_bootstrap.sql`, from the
`$EPOCHREALTIME` stamp it exports as `FLOW_GUARD_START`) under `guard-bootstrap` —, `hook_runs` by
hook (`0114_hook_runs.sql`: each `flow-guard hook <name>` run appends its wall time to the state
root's `hook-runs.jsonl` with no flowd call — about 17 µs a run — and the dev daemon ingests the file
on its harvest tick; a hook run carries no project or worktree, so it counts once when it ended inside
any ranked change's wall clock) and `verify_runs` by label. Each
dimension prints its top 10 — name, runs, total, share of the changes' summed wall clock (first
stage-run start to last end), p50, p95 — then one line of the top 5 time sinks across dimensions,
role class left out. Dimensions overlap, so shares do not add up across them. Ranking only; no UI.

## Trace

`flow stats trace [-project key] [-no-transcripts] [-o file] <change>` prints one change as
[Chrome Trace Event Format](https://docs.google.com/document/d/1CvAClvFfyA5R-PhYUmn5OOQtYMH4h6I0nSsKchNAySU)
JSON, one event per line, from `GET /api/v1/stats/trace?project=P&change=C`: the change's
`stage_runs`, `requests`, `dispatches`, `guard_runs` (claimed by worktree, as **Bottlenecks** claims
them) and `verify_runs` (a guard or verify run starts at its recorded end less its duration). The
change is one process (`process_name` `<project>/<change>`); its threads are lanes, top to bottom —
`request`, `stage`, one per dispatch named `<role> <dispatch key>` (`#<seq>` without a key), that
dispatch's `<lane> tools`, `guard`, `guard wait` (a guard run that waited, from its start for the
length of its `wait_ms`), `guard bootstrap` (ending at the run's start), `verify` — each named by `thread_name` and ordered by
`thread_sort_index`. Every row is a complete event (`"ph":"X"`, `ts`/`dur` in microseconds from the
change's first row, its start in `otherData.t0`); a row overlapping the previous one in its lane
spills into `<lane> (2)`, … A row still open ends at the trace's last instant with `"open": true` in
its args. Tool calls come from each dispatch's own subagent transcript, found as the harvester finds
it (`~/.claude/projects/*/*/subagents/agent-<id>.jsonl`, `FLOW_TRANSCRIPTS_DIR` overriding the root):
one span per `tool_use` to its `tool_result`, named by tool and a short form of its input (`Bash: go
test ./...`), kept inside its dispatch — a call started outside it is dropped, one running past its
end cut there; `-no-transcripts` leaves the tool spans out. Every stage run and dispatch row carries
the time split flowd stored for it: its args carry `llmMs`, `toolMs` (`toolMsBy` per tool, Bash by
command word) and `humanMs`. The rule (`harvest.TimeSplit`): each gap between two consecutive
timestamped transcript lines belongs to the line that ends it — an assistant line ends LLM time, a
user line answering a `tool_use` ends tool time, any other user line (a user message after an
assistant turn) ends human time; gaps never overlap, so the three sum to the row's covered span.
`GET /api/v1/stats/run/{project}/{change}` sums the stored dispatch splits into
`totals.agentTimeSplit`, and splits `totals.idleMs` the same way into `totals.idleSplit`: the
gaps of the parent transcripts the change's stage runs link, each kept only where it lies inside an
idle / parent segment, so its three sum to the share of idle those transcripts cover. The output is
stable for the same rows.

flowd's harvest tick stores the split (`cmd/flowd/timesplits.go`, `0115_time_splits.sql`). A stage
run is linked to the transcript of the agent that marked it, derived from its
`cc-<session id>-<agent>` session token — the session's own `<session>.jsonl` for `main`, the
agent's `subagents/agent-<id>.jsonl` otherwise; an `mf-`/`fi-` token gets no link and no split.
An ended dispatch gets its agent transcript's split at once; an ended stage run once its transcript
has a line at or past its end, or 15 minutes after it. Each row is computed once; a retime or
restamp of either instant clears it for the next tick to recompute. An open row has no split.

To look at it, write it to a file and drop that file on <https://ui.perfetto.dev> (or load it in
`chrome://tracing`):

```bash
flow stats trace -o /tmp/trace.json kan-1167-dispatch-recording-guard
```

## Confidence estimates

`flow confidence record` files a below-100% estimate with its gaps and tags (`flow confidence tags`
lists those in use). flowd serves them on `/api/v1/confidence/{project}/{change}`,
`/api/v1/confidence-tags` and `/api/v1/stats/confidence-estimates`; the SPA's **Confidence** view
approves or declines each (`PATCH /api/v1/confidence/{project}/estimates/{id}/review`).

## Jira transitions

The daemon can perform the pipeline's Jira status transitions itself: the
four-position forward-only table (To Do, In Progress, In Review, Done) with
retries against a transient Atlassian outage, so a landing run is not
stranded by one failing call. The mechanism lives in
`internal/jira`; flowd serves it at `POST /api/v1/jira/transition`, and the
`flow` CLI drives it:

```bash
flow jira transition KAN-571 "In Review"
```

`<target>` is a position name in any of its usual spellings — `In Review`
and `Code Review` resolve alike. An issue already at or past the target is
a success reported as `already <status> (no transition)`; nothing ever
moves backward, and a status matching no position is refused with its name
rather than guessed at.

The block is off until the daemon is given credentials. All three
environment variables are one unit — set them together or not at all;
`flowd` refuses to start on any partial combination:

| Variable | Meaning |
|----------|---------|
| `FLOWD_JIRA_SITE` | the site's base URL, e.g. `https://tweety53.atlassian.net` |
| `FLOWD_JIRA_EMAIL` | the Atlassian account's email |
| `FLOWD_JIRA_TOKEN` | that account's API token (id.atlassian.com → Security → API tokens) |

An unconfigured daemon starts normally and answers every transition with a
503 naming the variables — the pipeline treats any such failure as one
`Jira: skipped — <reason>` line and carries on, per the never-blocking
Jira contract. When run under the launchd agent, add the three variables
as an `EnvironmentVariables` dictionary in the installed plist (an
operator step, like loading the agent itself); when run from a shell,
export them first.

## Pricing

`flowd` seeds the published Anthropic per-model rates (`internal/store
/pricing_seed.go`, read from
https://platform.claude.com/docs/en/about-claude/pricing on 2026-08-14)
into the `pricing` table at every startup — an upsert, so this is a no-op
on every run after the first. Once a harvest batch commits, the daemon
prices every stage run it touched: each model's tokens are charged at
that model's own rate in effect at the run's start, with cache-creation
writes split into their two real rates (a 5-minute write costs 1.25x base
input, a 1-hour write costs 2x — collapsing them onto one rate silently
misprices whichever kind dominates), and fast-mode input/output rates
applied when the harness recorded `speed: "fast"`. A model absent from
the seed, a cache-creation write whose split was never recorded, or a
fast-speed run against a model with no published fast rate all price as
**unavailable**, never at a guessed or invented rate — the same
absence-is-not-a-value rule this store's metrics bag follows everywhere
else. See `internal/store/pricing.go` and `internal/store/pricing_seed.go`
for the mechanism, and `internal/harvest/attribute.go`'s `Bucket` for how
the cache-creation split reaches the metrics bag in the first place.

A dispatch is priced at the rate of the model its transcript says served
it (`signals.served_models`, the key with the most turns), never the alias
it was given. To price stage runs harvested before a pricing change again,
run `flow pricing reprice -since <YYYY-MM-DD>`: it prints how many runs
priced and how many were skipped for want of tokens or a rate.

Each dispatch row is also priced at its served model's rate in effect at
its start, set on its own `metrics.cost_usd` and repriced by
`flow pricing reprice`. Readers show it as a breakdown only, never added to
a stage total, since the stage run's `models.<m>` already holds those
tokens.

A transcript record goes to the stage window of its session that contains
its timestamp; a record between two stage windows goes to the stage that
began last before it, up to the change's finish, and a record before the
session's first mark goes to none. The spend the earlier rule dropped
between windows is back-filled once, when `flowd` starts, from the
transcripts still on disk — only the bytes already harvested, before the
watcher's first cycle — and the `harvest_backfills` row `stage-gap-cost`
records that it ran, so it never runs again.

The run costs views count the parent's own command time from
`guard_runs`: a guard run belongs to the change whose `-worktrees/<change>`
segment the worktree it was pointed at carries — the first argument that
is a directory under such a path, else its working directory. Parent
commands are those runs' union outside every dispatch row, and idle /
parent is the wall clock left over.

## Checking attribution by hand

A stage run recorded with `session_id` NULL leaves the harvester attributing
transcript offsets to nothing: the dashboards just look empty and nothing fails
loudly.

While a `flow-postgres` stack and `flowd` are both running (see above),
and while a real `/flow*` command is mid-run (so a change and at least
one stage mark already exist):

```bash
# 1. Find the change's most recent stage run.
curl -s 'http://127.0.0.1:4173/api/v1/stage-runs?project=<project-key>&name=<change-name>&sort=-started_at&limit=1' | jq .

# 2. Wait one harvest cycle (5s, cmd/flowd's harvestInterval) -- longer
#    if the mark's own turn hasn't flushed to the transcript yet.
sleep 6

# 3. Re-query the same stage run by id.
curl -s 'http://127.0.0.1:4173/api/v1/stage-runs?id=<id>' | jq '.stageRuns[0] | {sessionId, metrics}'
```

Expect `sessionId` to be a real session id, not `null`, and `metrics` to
carry a non-empty `tokens` object and a `cost_usd` — not `{}`. A stage
still short of its first flushed turn, or run on a harness that writes no
transcript, stays `sessionId: null` with empty `metrics`
honestly — that is the *recorded, not measured* state (see "Pricing"
above and this repository's `run-telemetry` capability), not a
symptom by itself. What is a symptom: every stage run staying unbound
minutes after it should have flushed, across every change. On Claude Code `sessionId` is set by the mark itself from
`CLAUDE_CODE_SESSION_ID` and is present before the first harvest
cycle; only `metrics` waits on the harvester.

A dispatch row begun late is corrected with `flow record dispatch retime`, which requires
`-reason`; the row's first original instants are kept in its `retimed` column across later
retimes, beside the latest reason. There is no `flowd` verb for it. A gated reviewer's or merger's
plan check runs at its hand-back: `hooks/await-own-agents.py` runs `flow record dispatch handback`
when the agent stops, and the parent's later `flow record dispatch end` trusts the marker it
leaves, so a plan edit after the hand-back no longer blocks the end.

## Running the daemon at login

`stats/launchd/com.tweety53.flowd.plist` is a macOS user launchd agent
that starts `flowd` at login, restarts it if it exits non-zero (a clean
exit is never restarted, only a crash or failure is), and logs to
`~/Library/Logs/flowd.log`. It binds loopback only, the same guarantee
`flowd` itself enforces at startup regardless of how it is launched.

`make restart` also installs the `flow` CLI to `~/.local/bin/flow` and runs `flowd` outside launchd instead, appending its log to `/tmp/flow-live.log` (`LIVE_LOG`, `stats/Makefile`); while it does, `~/Library/Logs/flowd.log` is not written. Postgres logs lock waits and statements slower than one second (`ALTER SYSTEM`, kan-1040), so a store stall leaves evidence in `docker logs flow-postgres`.

**Loading this agent is a manual operator step.** No skill or command in
this repository loads it — an agent left running unattended during this
change's development harvested 2,961 transcript offsets into the database
before anyone noticed, which is exactly the failure mode a
skill-triggered `launchctl load` would risk repeating. Install it
yourself:

```bash
# 1. Bring up the dedicated PostgreSQL stack (see above).
cd stats && docker compose up -d

# 2. Build the daemon binary the agent points at. The plist's
#    ProgramArguments is a fixed path, so build to that exact location —
#    the main checkout, not a worktree.
cd stats && make build

# 3. Install and load the agent.
cp stats/launchd/com.tweety53.flowd.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.tweety53.flowd.plist

# 4. Confirm it's up.
curl -s http://127.0.0.1:4173/api/v1/changes
```

To stop it:

```bash
launchctl unload ~/Library/LaunchAgents/com.tweety53.flowd.plist
```

Unloading stops the daemon without removing the binary or the Postgres
stack; every `flow` CLI call falls back to the on-disk journal while it
is down, per the never-block guarantee, and replays automatically the
next time the daemon starts.

To update the binary after pulling new code, rebuild it with the same two
commands from step 2 above, then either `launchctl kickstart -k` the
agent or unload/reload it — `KeepAlive` does not pick up a replaced
binary on its own.
