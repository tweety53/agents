# Per-dispatch cost attribution — make the recorded figures true

The problem, the audit that found it and the three failure modes are in
`docs/superpowers/specs/2026-08-23-kan-212-persist-per-dispatch-cost-tokens-model-and-role-design.md`.
This file carries the decisions taken and the questions left open.

## Decisions

### The agent identifier attributes regardless of interval

**ID:** `identity-beats-interval`
**Status:** active
**Chosen:** `bestDispatchWindow` matches on `agent_id` first and unconditionally — a record whose
`agentId` equals a dispatch's recorded identifier attributes to that dispatch whether or not the
record's timestamp falls inside that dispatch's window. The interval is consulted only when no
identifier is reported on one side or the other.
**Considered:** Keeping the identifier a preference *among* windows whose interval contains the
record, as today, and fixing the intervals instead — rejected because `-started-at` and `-ended-at`
are typed by the dispatching agent at the moment it writes the mark, and the audit found values
rounded to the minute (`02:00:00`, `00:10:00`); an interval that is approximate by construction
cannot gate an exact identifier. Widening the window by a tolerance — rejected as a magic number that
would have to grow every time a run was slower than the tolerance assumed, and that reintroduces
overlap between adjacent panel rounds.

### Two dispatches recording the same identifier attribute to neither

**ID:** `duplicate-identifier-is-ambiguous`
**Status:** active
**Chosen:** Where two dispatch rows carry the same non-empty `agent_id`, a record bearing that
identifier attributes to none of them, and the ambiguity is recorded rather than resolved.
**Considered:** Breaking the tie by interval, or by the lower dispatch id — both reintroduce, one
layer down, exactly the silent wrong split this change exists to remove. An identifier that names two
dispatches has said nothing about which one incurred the record.

### An indistinguishable overlap attributes to nothing

**ID:** `ambiguous-overlap-attributes-to-none`
**Status:** active
**Chosen:** Where two or more dispatch windows contain a record and no identifier separates them, no
dispatch is credited. The affected dispatches record their cost as unattributed, with the count of
candidates that could not be told apart.
**Considered:** Keeping the latest-started rule — that is the present behaviour and the source of
`kan-295`'s wrong figures. Splitting the record's tokens evenly across the candidates — rejected: an
even split is a fabricated measurement, and `myflow-run-telemetry` already forbids recording an
unmeasured value as a value. Attributing to the whole group as one unit — rejected: the row grain is
one dispatch, and there is no group row to carry it.

### The identifier may be recorded at `end` as well as at `begin`

**ID:** `agent-id-recordable-late`
**Status:** active
**Chosen:** `myflow record dispatch end` accepts `-agent-id`, so a dispatch whose identifier is not
knowable when `begin` fires still records it. Attribution reads dispatch windows fresh on every
harvest cycle, so an identifier recorded late still attributes the records already read.
**Considered:** Moving `begin` to after the launch so the identifier is always available at insert —
rejected because the row must exist before the first harvest tick the dispatch runs through, which is
the whole reason KAN-258 split the record into two calls; delaying `begin` to obtain the identifier
would reopen the loss it closed. Requiring the identifier — rejected: two of the three supported
harnesses expose none, and an absent identifier is ordinary rather than degraded.

### The abandoned-token set is persisted, and retried on restart

**ID:** `persist-the-give-up`
**Status:** active
**Chosen:** A token the watcher gives up on is recorded in the store rather than only in
`w.gaveUpTokens`, and the watcher re-attempts persisted give-ups when it starts, scanning the
transcripts for its marks independently of the harvest offset.

**What this recovers, and what it does not — measured live, not predicted.** The retry recovers a
token's **binding**, not its past cost. After the fix, `mf-kan302-a3f9` bound 15 of 15 stage runs and
`mf-kan190-a3f7` 11 of 11, both from zero; none of those stage runs gained token figures, because the
usage records behind the harvest offset are deliberately never re-read — that is what makes
double-counting structurally impossible.

So a **future** casualty is fully recovered: the give-up persists, the scan binds the token within the
same session while its transcript is still being written, and everything from that point attributes
normally. An **already-finished** run regains its identity and no more; its cost stays unmeasured,
now stamped `session never bound` rather than silently indistinguishable from a harness that writes no
transcript. Re-attributing a newly bound token's past usage is a follow-up, not this change.
**Considered:** Leaving it in memory and raising `maxSessionTokenResolutionCycles` — rejected: it
moves the cliff without removing it, and costs every never-binding harness more work at every
restart. An explicit `myflow rebind` command — offered at the design gate and declined: a command
nobody remembers exists is not a recovery path. Re-reading a bound token's transcript region to
attribute its past usage — offered once the live measurement showed binding alone recovers no cost,
and declined for this change: an unbound session's records attributed to nothing, so it would be a
first attribution rather than a second, but it reopens exactly the double-count risk the
offset-independent scan was built to make impossible.

### Unattributed cost is stated in the ledger and in the handoff

**ID:** `unattributed-is-visible`
**Status:** active
**Chosen:** The ledger distinguishes three states — `not measured` (the harness writes no transcript),
`cost unattributed — session never bound`, and
`cost unattributed — indistinguishable from N concurrent dispatches` — and `/myflow-do`'s handoff
prints one line when the run's own cost did not attribute.
**Considered:** Persisting the fact in the store only, leaving the ledger and handoff unchanged —
rejected: the whole defect is that a loss is discoverable only by someone who already suspects it. A
dashboard indicator as well — deferred; it belongs with the aggregation views this change leaves to
KAN-198, KAN-208 and KAN-201, which is where an incomplete period actually misleads.

### No aggregation views over `dispatches`

**ID:** `no-views-in-this-change`
**Status:** active
**Chosen:** This change adds no cross-change query, endpoint or web surface over the `dispatches`
table. It repairs attribution and stops there.
**Considered:** Repairing attribution and adding the views together, and adding the views on the data
as it stands — both offered at the design gate and declined. The second would publish a per-slot cost
that attributes a panel's whole spend to whichever slot sorted first.

## Open questions

### Why `kan-302`'s session exhausted its binding window

**ID:** `kan-302-give-up-trigger`
**Status:** open
**Why it is open:** The *mechanism* is confirmed — `resolveSessionTokens`' `case 0`, the bounded
give-up after 60 cycles — but not the *trigger*. The marks are present and well-formed, the daemon
was running, and 5 minutes should have been ample.
<!-- measured: maxSessionTokenResolutionCycles (60) × cmd/myflowd's harvestInterval (5s) @ 4a305d0 -->
Reproducing it needs a live run with the
watcher's own logging retained, which this change makes possible rather than performs.
**What it affects:** Whether a fourth fix is needed beyond persisting and retrying the give-up. The
retry recovers the data either way, so nothing in this change's scope depends on the answer.

### Whether a synchronous dispatch can report its identifier at all

**ID:** `sync-dispatch-identifier`
**Status:** open
**Why it is open:** The audit confirmed an identifier is returned for an **async** agent launch, in
the parent's own tool result. Whether a synchronous dispatch's result exposes one was not
established.
**What it affects:** Nothing in this change. `/myflow-do` dispatches panel slots asynchronously, and
those are the only dispatches whose windows overlap; a synchronous dispatch attributes correctly by
interval. If a synchronous shape later needs an identifier, `-agent-id` on `end` already carries it.
