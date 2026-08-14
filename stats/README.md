# stats

The myflow stats app: a PostgreSQL-backed Go service that replaces the
per-machine JSON state file, records per-stage telemetry, and serves both the
live pipeline state and aggregated statistics through a browser interface.

Full design: `openspec/changes/kan-16-myflow-stats-app/design.md`.

## Prerequisites

- Go 1.26.5 (`go version` to confirm).
- Docker, for the dedicated PostgreSQL stack in `docker-compose.yml`.

## The PostgreSQL stack

A dedicated `myflow-postgres` container, independent of any other Postgres
stack already running on this machine (for example a `gymie-postgres`
container on port 5432). It listens on host port **5433** so the two never
collide — confirmed free on this machine before the port was chosen.

```bash
cd stats
docker compose up -d
```

This starts only `myflow-postgres`; it does not stop or modify any other
running container.

## Running the tests

```bash
cd stats
go test ./...
```

Tests that need the compose stack's PostgreSQL skip themselves with a clear
message when the stack is not running, so `go test ./...` is always safe to
run without Docker.

## Pricing

`myflowd` seeds the published Anthropic per-model rates (`internal/store
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

While a `myflow-postgres` stack and `myflowd` are both running (see above),
and while a real `/myflow-*` command is mid-run (so a change and at least
one stage mark already exist):

```bash
# 1. Find the change's most recent stage run.
curl -s 'http://127.0.0.1:4173/api/v1/stage-runs?project=<project-key>&name=<change-name>&sort=-started_at&limit=1' | jq .

# 2. Wait one harvest cycle (5s, cmd/myflowd's harvestInterval) -- longer
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
above and this repository's `myflow-run-telemetry` capability), not a
symptom by itself. What is a symptom: every stage run staying unbound
minutes after it should have flushed, across every change — that is
exactly what KAN-16 looked like.

## Running the daemon at login

`stats/launchd/com.tweety53.myflowd.plist` is a macOS user launchd agent
that starts `myflowd` at login, restarts it if it exits non-zero (a clean
exit is never restarted, only a crash or failure is), and logs to
`~/Library/Logs/myflowd.log`. It binds loopback only, the same guarantee
`myflowd` itself enforces at startup regardless of how it is launched.

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
go build -o bin/myflowd ./cmd/myflowd

# 3. Install and load the agent.
cp stats/launchd/com.tweety53.myflowd.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.tweety53.myflowd.plist

# 4. Confirm it's up.
curl -s http://127.0.0.1:4173/api/v1/changes
```

To stop it:

```bash
launchctl unload ~/Library/LaunchAgents/com.tweety53.myflowd.plist
```

Unloading stops the daemon without removing the binary or the Postgres
stack; every `myflow` CLI call falls back to the on-disk journal while it
is down, per the never-block guarantee, and replays automatically the
next time the daemon starts.

To update the binary after pulling new code, rebuild it with the same two
commands from step 2 above, then either `launchctl kickstart -k` the
agent or unload/reload it — `KeepAlive` does not pick up a replaced
binary on its own.
