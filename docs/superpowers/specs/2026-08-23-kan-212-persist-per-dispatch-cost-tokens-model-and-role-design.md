# Per-dispatch cost attribution — make the recorded figures true

**Jira:** KAN-212
**Date:** 2026-08-23

## What the ticket asked for, and what already exists

KAN-212 asked for per-dispatch cost — change, stage, role, model, tokens — to be persisted in the
stats store. **That landed with KAN-258 four days after the ticket was filed.** `dispatches`
(migration `0010_run_records.sql`) carries `change_id`, `stage_run_id`, `task_id`, `role`, `slot`,
`model`, `session_token`, `started_at`, `ended_at` and a `metrics` JSONB bag; `myflow record
dispatch begin | end` writes it; `/myflow-do` fires those calls at all four dispatch sites; KAN-201's
`DispatchAttributor` fills the bag from the harness transcript rather than from anything an agent
reports about itself.

Each of the ticket's three open questions is answered by what shipped:

| Open question | Answer as shipped |
|---|---|
| Who writes it | `myflow record dispatch begin`/`end`, mirroring `myflow stage` — the shape the ticket predicted |
| Whether the harness can supply it directly | The harvester reads the transcript; no agent reports its own consumption |
| Dispatches whose model is `unknown (agent-defined)` | The CLI accepts that literal, and `RenderLedger` prints it verbatim; cost is attributed independently of the model |

**So the persistence half of KAN-212 is done. This change is about the half that is not: the figures
being recorded are incomplete, and where the panel is concerned they are wrong.**

## The audit

Read directly from the live store on 2026-08-23, over every dispatch row it holds.

| Change | Dispatches | With `agent_id` | Closed | With session token | Costed |
|---|---|---|---|---|---|
| `kan-242-devstop-not-running-while-stack-alive` | 14 | 0 | 14 | 14 | 7 |
| `kan-286-numeric-keyboard-for-reps-and-seconds` | 8 | 8 | 8 | 8 | 1 |
| `kan-295-cut-pipeline-load-cost-split-by-consumer` | 20 | 0 | 20 | 20 | 12 |
| `kan-302-panel-code-review-slot-hangs-on-fork` | 12 | 0 | 12 | 12 | 0 |
| `keyboard-positioning-across-the-app` | 3 | 3 | 3 | 3 | 0 |
| **Total** | **57** | **11** | **57** | **57** | **20** |

<!-- measured: SELECT over dispatches JOIN changes, myflow store @ 2026-08-23 -->

Every row is written and closed correctly, and every row carries its session token. **35% carry any
token figure at all**, and the missing 65% is not spread evenly — it is concentrated on the review
panel, which is where most of a run's cost is incurred.

### F1 — concurrent panel slots collapse onto one row

`kan-295`, whose panel slots are dispatched together and therefore share a byte-identical window:

| seq | role | slot | window | costed |
|---|---|---|---|---|
| 3 | reviewer | Primary | 23:36:15 → 23:46:23 | **yes** |
| 4 | reviewer | Principles | 23:36:15 → 23:46:23 | no |
| 5 | reviewer | Code review (low) | 23:36:15 → 23:46:23 | no |
| 7 | reviewer | Primary | 00:12:04 → 00:25:53 | **yes** |
| 8 | reviewer | Principles | 00:12:04 → 00:25:53 | no |
| 9 | reviewer | Code review (low) | 00:12:04 → 00:25:53 | no |

<!-- measured: SELECT seq, role, slot, started_at, ended_at, (metrics ? 'tokens') over kan-295 @ 2026-08-23 -->

Exactly one slot per concurrent group is costed, and it is charged with the **whole group's**
sidechain tokens. The figure is not merely incomplete: Primary's number contains Principles' and Code
review's spend, and the ledger presents it with no qualification.

The cause is `agent_id`, the column added by migration `0011_dispatch_agent_id.sql` for exactly this
case. It is NULL on all 46 dispatches this repository has recorded, so `bestDispatchWindow` falls
through to interval containment, where every overlapping window matches and the latest-started wins.

`skills/myflow-do/SKILL.md` does pass `-agent-id <id>`, but `begin` fires before the Agent tool is
called and the skill treats the id as something the harness may or may not expose. It does expose it:
an async agent launch returns

```text
agentId: a392afd1eacbdfebc
```

in the **parent's own tool result, at launch**. The kan-302 run dispatched all ten of its panel agents
that way. The id is available and simply never captured.

### F2 — a whole session never binds, and nothing says so

`kan-302`'s `/myflow-do` session, token `mf-kan302-a3f9`: **15 stage runs — including `do.sdd-tdd`
and `do.review-panel` — none bound to a session**, and therefore 12 dispatches with no cost at either
grain. The later finish session `mf-kan302-b7e2` bound normally, so this is per-session, not
per-change.

The marks are not at fault: all 30 are present in that session's transcript in the shape
`isSessionMarkCommand` recognises, leading the command and mid-`&&`-chain alike. This is
`resolveSessionTokens`' `case 0` — the bounded give-up after `maxSessionTokenResolutionCycles`
(60 cycles ≈ 5 minutes).

What makes it a defect rather than an accepted bound is what happens next:

- `w.gaveUpTokens` is **in process memory**, so the token is never retried, even though its transcript
  still holds every mark needed to bind it;
- the only report is `w.warn`, and on this machine `myflowd` runs outside launchd, so that stream is
  not written anywhere;
- the store records nothing, the ledger says `not measured` — indistinguishable from a harness that
  writes no transcript — and the `/myflow-do` handoff says nothing at all.

### F3 — a correct `agent_id` that still does not attribute

Gymie's `kan-286` recorded an `agent_id` on all 8 dispatches, and every one of those ids appears in a
transcript. **One is costed.** Two of its windows:

```text
a9b17029fcec7d2f1   02:00:00 → 04:44:34
a05e47499777b21c7   00:02:15 → 00:10:00
```

<!-- measured: SELECT agent_id, started_at, ended_at over kan-286 @ 2026-08-23 -->

`02:00:00` and `00:10:00` are not measured instants. `-started-at` and `-ended-at` are typed by the
dispatching agent, so **a dispatch's window is approximate by construction** — while `bestDispatchWindow`
only considers windows whose interval *contains* the record and treats the agent id as a preference
among them. An exact identifier is gated behind a hand-typed guess.

## What this change does

1. **Identity beats interval.** A record whose `agentId` equals a dispatch's recorded `agent_id`
   attributes to that dispatch regardless of interval. The interval becomes the fallback used only
   when no id is reported on either side — which is every dispatch on Cursor and Codex, and every
   non-panel dispatch, where it is exactly correct because those windows do not overlap.

2. **Refuse to guess.** Where two or more windows contain a record and no identifier separates them,
   attribute to none and record the dispatch's cost as unattributed. This is the dispatch-grain
   counterpart of the session-grain rule `myflow-run-telemetry` already states — *"no session is
   recorded for that stage run, and the ambiguity is reported rather than resolved by choosing"* —
   which today has no analogue one grain down.

3. **Capture the id.** `/myflow-do` reads `agentId` from the launch result and passes `-agent-id` on
   `begin`; `myflow record dispatch end` gains `-agent-id` so an id that becomes known late is still
   recorded.

4. **Persist the give-up and retry on restart.** The abandoned-token set moves from process memory to
   the store, and the watcher re-attempts persisted give-ups when it starts.

5. **Say so where someone is looking.** The ledger tells three states apart, and `/myflow-do`'s
   handoff prints one line when the run's own cost did not attribute.

## What this change does not do

**No aggregation views over `dispatches`.** Nothing in `internal/store/aggregate.go` reads that table
today; every view aggregates `stage_runs` alone, and the only reader of dispatch rows is
`RenderLedger`. KAN-198, KAN-208 and KAN-201 each want a different cross-change view, and each should
specify its own — on data that is correct by the time they do. Building a per-slot findings-per-token
view on today's attribution would publish a confident number attributing a whole panel's cost to
whichever slot sorted first.
