## Why

Per-dispatch cost is already persisted — KAN-258 shipped the `dispatches` table, the
`myflow record dispatch` verb and the harvester's second attribution pass, four days after KAN-212
was filed. What is not true is the figures.

An audit of every dispatch row the store holds found **20 of 57 costed**, and the missing 65%
concentrated on the review panel, where most of a run's spend happens. Three failure modes, all
silent:

- **Concurrent panel slots collapse onto one row.** Slots dispatched together share a byte-identical
  window; with no `agent_id`, attribution falls through to interval containment and one slot is
  charged with the whole group's tokens while its siblings record zero. Not incomplete — **wrong**,
  and presented unqualified.
- **A whole session can fail to bind, and nothing says so.** `kan-302`'s `/myflow-do` session left 15
  stage runs and 12 dispatches uncosted at both grains. The abandoned-token set lives in process
  memory, so it is never retried even though the transcript still holds all 30 marks; the only report
  is a log line to a stream that, with `myflowd` run outside launchd, goes nowhere.
- **A correct `agent_id` still need not attribute.** `-started-at` and `-ended-at` are typed by the
  dispatching agent — the audit found windows ending at `00:10:00` and starting at `02:00:00` — so a
  dispatch's interval is approximate by construction, while an exact identifier is gated behind it.

Three open tickets — KAN-198, KAN-208, KAN-201 — each propose a cross-change view over this data.
Every one would publish a number attributing a panel's whole cost to whichever slot sorted first.

The evidence, per failure mode, is in
`docs/superpowers/specs/2026-08-23-kan-212-persist-per-dispatch-cost-tokens-model-and-role-design.md`.

## What Changes

- **Identity beats interval.** A record whose `agentId` equals a dispatch's recorded `agent_id`
  attributes to that dispatch regardless of interval. The interval drops to the fallback it should
  always have been: used only when no identifier is reported on either side.
- **Ambiguity attributes to nothing.** Where two or more windows contain a record and no identifier
  separates them, no dispatch is credited, and the affected dispatches record their cost as
  unattributed — the dispatch-grain counterpart of the session-grain ambiguity rule
  `myflow-run-telemetry` already states.
- **The identifier is captured.** `/myflow-do` reads `agentId` from the harness's launch result and
  passes `-agent-id` on `begin`; `myflow record dispatch end` gains `-agent-id`, so an identifier that
  becomes known late is still recorded.
- **The give-up is persisted and retried.** The abandoned-token set moves out of process memory into
  the store, and the watcher re-attempts persisted give-ups on start, scanning for their marks
  independently of the harvest offset. This recovers a token's **binding**, not its past cost:
  measured live, `mf-kan302-a3f9` bound 15 of 15 stage runs and `mf-kan190-a3f7` 11 of 11, both from
  zero, and neither gained token figures. A **future** casualty is therefore fully recovered — the
  scan binds it while its transcript is still being written, and everything after attributes normally
  — while an already-finished run regains identity alone. Re-attributing past usage for a newly bound
  token is a follow-up.
- **Unattributed cost is visible.** The SDD ledger distinguishes `not measured`,
  `cost unattributed — session never bound` and
  `cost unattributed — indistinguishable from N concurrent dispatches`; `/myflow-do`'s handoff prints
  one line when the run's own cost did not attribute.

**Not in this change:** no aggregation views over `dispatches`. KAN-198, KAN-208 and KAN-201 each
specify the view they need, against data that is correct by then.

## Impact

- **Affected specs:** `myflow-run-record`, `myflow-run-telemetry`
- **Affected code:** `stats/internal/harvest/` (attribution and the watcher), `stats/internal/store/`
  (a persisted give-up, plus one migration), `stats/internal/records/render.go` (the ledger's three
  states), `stats/cmd/myflow/record.go` (`-agent-id` on `end`), `skills/myflow-do/SKILL.md` (capture
  the identifier, and the handoff line)
- **Behaviour that changes for existing rows:** a dispatch previously credited with a concurrent
  group's whole spend stops being credited with it. Change-level and stage-level totals are computed
  by a separate pass and are unaffected — what stops is the false split, not the total.
