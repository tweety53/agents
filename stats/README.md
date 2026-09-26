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

## Running the tests

```bash
cd stats
go test ./...
```

Tests that need the compose stack's PostgreSQL skip themselves with a clear
message when the stack is not running, so `go test ./...` is always safe to
run without Docker.

**This stack is not needed for the Go test suite.** `go test ./...` already
gets its own isolation: every test that touches Postgres creates its own
database (see `internal/store/testsupport_test.go`), used once and never
shared with the live database, another test, or the UI-test stack below.
Nobody running `go test` needs to bring up the UI-test stack first.

## flow-guard

`flow-guard` (`cmd/flow-guard`) runs the Go ports of the repository's guard
scripts. A ported `scripts/<name>.sh` keeps its header comment as the guard's
contract and runs `flow-guard <name>` built from its own checkout
(`scripts/lib/flow-guard.sh`): it hashes the sources `flow-guard` is built
from, builds on the first call for a hash into
`${XDG_CACHE_HOME:-$HOME/.cache}/flow-guard/<hash>/flow-guard`
(`FLOW_GUARD_CACHE_DIR` overrides the directory), and execs the cached binary
after that. Nothing is installed. With no `go` to build it, or a failed build,
the shim exits that guard's own cannot-answer code — 4 for `run-reproducer`,
2 for the others — and names the cause. Before/after suite timings
for the port are recorded in the design of change
kan-760-agents-port-the-flow-guard-scripts-and-their, not here.

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

## Jira transitions

The daemon can perform the pipeline's Jira status transitions itself: the
four-position forward-only table (To Do, In Progress, In Review, Done) with
retries against a transient Atlassian outage, so a landing run is not
stranded by one failing call (KAN-571). The mechanism lives in
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

## Checking attribution by hand

KAN-16 shipped a measurement system that could not measure: every stage run
was recorded with `session_id` NULL, so the harvester attributed every
transcript offset it read to nothing, and the dashboards just looked
empty. Nothing failed loudly — it was found by a human staring at that
empty dashboard and asking why. This is the same check, made a minute's
work instead of an investigation.

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
still short of its first flushed turn, or run on a harness with no
transcript (Cursor, Codex), stays `sessionId: null` with empty `metrics`
honestly — that is the *recorded, not measured* state (see "Pricing"
above and this repository's `run-telemetry` capability), not a
symptom by itself. What is a symptom: every stage run staying unbound
minutes after it should have flushed, across every change — that is
exactly what KAN-16 looked like. On Claude Code `sessionId` is set by the mark itself from
`CLAUDE_CODE_SESSION_ID` and is present before the first harvest cycle; only `metrics` waits on the
harvester.

## Running the daemon at login

`stats/launchd/com.tweety53.flowd.plist` is a macOS user launchd agent
that starts `flowd` at login, restarts it if it exits non-zero (a clean
exit is never restarted, only a crash or failure is), and logs to
`~/Library/Logs/flowd.log`. It binds loopback only, the same guarantee
`flowd` itself enforces at startup regardless of how it is launched.

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
