## Why

The stats app records stage runs and attributes no usage to them, so every token and currency
figure is empty in practice. Measured against a live session on 2026-08-14: **two stage runs
recorded, 2,988 transcript offsets consumed, zero metrics.**

A harvested message binds to a stage run by **session id first, timestamp second**
(`internal/harvest/attribute.go`): a record with an empty session id is skipped outright, and window
matching requires the two ids to be equal. Every stage run written so far has `session_id` NULL. The
harvester is reading every transcript correctly and attributing all of it to nothing.

The cause is in the pipeline contract: *"Where the harness exposes a session identifier, `stage
begin` **may** carry it via `-session`."* Optional — because Claude Code does not put the session id
in the environment, so a skill has nothing to read. But the entire telemetry half of the application
is inert without it, which makes it load-bearing rather than optional. **KAN-16 shipped a
measurement system that cannot measure.**

Two smaller defects compound it: `/myflow-fast` has no stage vocabulary to speak of, and `harness`
is `unknown` on every run.

## What Changes

- **BREAKING (contract)**: `stage begin` gains a required `-nonce` and a required `-harness`. The
  pipeline contract changes from *may carry a session* to *must carry a nonce*. Every `/myflow-*`
  skill's mark calls are updated.
- **New**: a stage run binds to its session **after** the fact. The skill writes a literal unique
  nonce into the `stage begin` command; that command text lands in the calling session's own
  transcript; the daemon finds it there and binds the run's `session_id` to that transcript's
  `sessionId`. From that point attribution is unchanged.

  **The nonce must be a literal the skill writes, never a shell substitution.** The transcript
  records the command as handed to the tool, before the shell runs, so `-nonce "mf-$(date +%s)"`
  would be recorded verbatim and identically in every session — discriminating nothing. This is the
  single detail on which the whole mechanism turns.
- **New**: resolution is **bounded**. A nonce not found within a fixed window is abandoned and its
  run stays honestly unattributed. On a harness with no transcript — Cursor, Codex — every mark takes
  that path once and cheaply, rather than becoming permanent rescan work.
- **BREAKING (vocabulary)**: every stage gains a **stable key** alongside its human-readable name.
  Today a stage's identity *is* its prose name, compared byte for byte — so
  `validate the project's `## workspace isolation` section, then export what it declares` is an
  identifier, carrying backticks and apostrophes into every shell invocation, and **rewording it
  silently splits that stage's history in two**. Keys are declared, never derived from the name, so a
  name can be improved without breaking continuity. Marks pass the key; the interface shows the name.
- **Changed**: `README.md`'s **Level 1 — the stages of each command** table is audited in full.
  `/myflow-fast`'s row is prose, not stages: splitting on the arrow separator yields 11 discrete
  stages for `/myflow-do` and **2** for `/myflow-fast`, one of which is a 400-character sentence. A
  whole `/myflow-fast` run therefore records one 28-minute "stage". Every command's row is checked
  and repaired.
- **Changed**: a stage run recorded but never attributed reads as **recorded, not measured** —
  distinct from "no data for this period" and from a measured zero. The interface currently draws
  that distinction with two arms and needs three; without the third, a misconfiguration is
  indistinguishable from a quiet week, which is exactly how this defect stayed invisible.

**Not changing:** the exactly-once harvest offsets, the pricing path, the never-block guarantee, the
three pipeline states, and any already-recorded data. **No backfill** — the handful of unattributed
runs are all from KAN-16's own development and are worth nothing.

## Capabilities

### Modified Capabilities

- `myflow-run-telemetry`: how a stage run is bound to the session whose usage it should receive, and
  what happens when it cannot be. The stage-mark obligation gains the nonce and the harness.
- `myflow-stats-views`: the third arm of the absence distinction — recorded-but-unmeasured.

## Impact

**Contracts** — `skills/myflow-contracts/pipeline.md`'s **Stage marks** section: `may carry a
session` becomes `must carry a nonce and a harness`.

**Skills** — `myflow-start`, `myflow-do`, `myflow-finish`, `myflow-fast` each update their mark calls.
`myflow-status` marks nothing and is untouched.

**Code** — `stats/cmd/myflow` (the new flags), `stats/internal/api` (accepting them),
`stats/internal/store` (a `session_nonce` column, a `stage_key` column, and the binding write),
`stats/internal/harvest` (finding nonces and binding), `stats/internal/stages` (the key vocabulary),
`stats/web` (keys for grouping, names for display, and the third arm).

**Already-recorded stage runs keep their prose stage and gain no key.** They are two rows from this
change's own development; migrating them would be inventing history.

**Docs** — `README.md`'s Level 1 table, audited across every command.

**Measured, not assumed** — every figure above was read from the live store and the live transcripts
on 2026-08-14, not estimated.
