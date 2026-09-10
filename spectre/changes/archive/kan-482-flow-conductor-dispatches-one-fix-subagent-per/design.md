## Context

`skills/flow/review-panel.md:905` already told the conductor to give every surviving finding to
**one** fix subagent as a combined list. On KAN-449 the conductor nonetheless dispatched four
background fix subagents, one per reviewer. The repo's own `gate-is-a-guard` decision — gates that
matter become guards, not better wishes — constrains the shape of the fix: the contract text is
tightened AND the obligation becomes a check at the panel's close gate.

Two facts enable the guard: the store already persists every dispatch row
(`stats/internal/store/records.go`'s `RecordDispatch`, keyed by session token, key, role, seq),
and the panel-fix role already has a canonical key shape (`panel-fix-<round>`, handshake retry on
`panel-fix-<round>-retry` per `skills/flow/implement.md`'s **The handshake**). What was missing is
a read path over dispatch rows — `flow record` carries reads for findings, verdicts, decisions and
incidents, but none for dispatches.

`spectre/specs/` is empty in this repository — the skill markdown is the contract surface — so the
change edits skill files directly and adds no spec task.

## Toggles

Resolved `default` for all three (`## execution mode`, `## implementer model`, `## review panel`).
This run executes inline by the operator's session instruction — an override of the Decide step's
`execution` result, recorded in `decision.json`'s `overrides`, never written back to any store.

## Decisions

### Enforce the single-fix-dispatch contract with prose plus a mechanical guard

**ID:** prose-plus-guard
**Status:** active
**Chosen:** tighten the fix-step wording in `skills/flow/review-panel.md` and the conductor relay
contract in `skills/flow/implement.md`, and add `check-panel-fix-single-dispatch.sh` fed by a new
`flow record dispatches` read command — prose alone already failed once at line 905, and the
repo's `gate-is-a-guard` decision puts obligations that were violated into guards.
**Considered:** prose only — smallest, but it is exactly what failed on KAN-449; guard only —
leaves the misread contract text untouched, so the next conductor re-derives the same drift before
any guard can fire.

### Guard counts dispatch rows per round at the panel's close, not before dispatch

**ID:** guard-at-panel-close
**Status:** active
**Chosen:** run the guard immediately before `flow stage end flow.review-panel`, beside
`check-panel-findings-closed.sh` — dispatch rows only exist after dispatches happen, so a
pre-dispatch check has nothing to read; the close gate is the last point where the run can still
be stopped before handoff.
**Considered:** a pre-dispatch check — nothing recorded yet to check; a mid-round check after each
dispatch — three extra call sites for the same verdict the close computes once.

### The canonical key shape is part of the check, not just the count

**ID:** key-shape-check
**Status:** active
**Chosen:** the guard verifies every panel-fix row's key matches `panel-fix-<round>` or
`panel-fix-<round>-retry` — a conductor that invents keys (`panel-fix-f1`, one per finding) would
otherwise read as four clean single-dispatch rounds.
**Considered:** counting rows only — misses the exact failure mode observed, where invented keys
defeat any per-round count.

### A read command may fail; the never-blocks rule stays write-only

**ID:** reads-may-fail
**Status:** active
**Chosen:** `flow record dispatches` exits non-zero when the store is unreachable, mirroring
`flow record findings` — an unreadable store must never render as an empty array, which would
silently pass the guard.
**Considered:** falling back to an empty array with a warning — turns a store outage into a green
verdict, the exact dishonesty the findings guard's `no-journal-excuse` decision rejects.

## Open questions

