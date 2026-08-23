# Stop the panel's Code review slot hanging — drop the forking skill, give every slot a ceiling

## Why

Panel slot 3, **Code review (low)**, dispatches a `general-purpose` subagent told to invoke the
harness's `code-review` skill. That skill **forks its own background agent**. The parent panel
dispatcher never observes that inner agent's completion, so the slot cannot report through the
panel's contract.

Observed on KAN-295, three times. Every figure below is an observation of those runs, recorded in
KAN-302's own description; none was re-derived here.
<!-- measured: KAN-295 panel rounds 3 and 4, observed by the operator and recorded in KAN-302 -->

- **Round 3** — the inner agent bypassed the slot, returning its findings *directly to the session*.
  The slot itself stalled and was killed by the harness's stall watchdog with no result.
  It had run 39 minutes; the findings it delivered numbered 6; the watchdog's threshold is 600s.
<!-- measured: KAN-295 panel round 3 -->
- **Round 4** — the slot ran with no result and the stall watchdog never fired, because the slot
  kept emitting output while making no progress, so it sat alive and useless and the notification
  the dispatcher was waiting on never arrived. It was caught only because the operator asked
  whether it was stuck. Elapsed: **4 hours**.
<!-- measured: KAN-295 panel round 4 -->
- **Re-dispatched on the documented fallback** — a plain `general-purpose` reviewer doing the same
  checks, no skill — it returned `PANEL-SLOT-CLEAN` in **2.5 minutes**.
<!-- measured: KAN-295, the re-dispatch that followed round 4 -->


Two independent defects:

1. A slot whose work is forked to a background agent cannot report through the panel's contract.
   Slot 3 is the instance; the class is any forking skill or agent.
2. A slot that emits output while making no progress defeats a stall watchdog. Nothing in the panel
   bounds a slot's wall-clock time, and the dispatcher blocks on a completion notification that may
   never come.

The panel's fallback shape is not a degraded substitute for the skill — on the one head-to-head the
evidence carries, it was an order of magnitude faster and returned a usable result where the skill
returned none.

## What changes

- **Slot 3 becomes, unconditionally, a `general-purpose` reviewer** on `models.reviewPanel`, briefed
  to report high-confidence defects only. The `code-review` skill invocation is deleted, and the
  "where the harness offers no `code-review` skill" substitution paragraph goes with it — there is
  no longer a skill to be unavailable.
- **The slot keeps the name `Code review (low)`**, which is what `myflow record dispatch begin
  -slot` writes into the store. Existing rows stay comparable with new ones.
- **A new normative rule**: no panel slot is dispatched onto a skill or agent that forks its own
  background agent. Slot 3's change becomes an application of the rule rather than a one-off.
- **A new normative rule**: every panel slot carries a **15-minute wall-clock ceiling**. The
  dispatcher checks each in-flight slot's elapsed time rather than blocking indefinitely on a
  completion notification. On a breach the slot is stopped, its dispatch row closed with
  `-outcome timed-out`, the breach recorded in the panel record, and that one slot re-dispatched
  once. A second breach stops and asks the operator.
- **Rationale recorded, not replaced**: `SKILL-rationale.md`'s recorded reason for choosing the
  `code-review` skill stays, with the KAN-295 evidence that reversed it recorded beside it.

Documentation only. No Go, no SPA, no guard script. `-outcome` is a free-form string in the CLI,
the API and the store, so `timed-out` needs no schema change.
<!-- verified: no enum on Outcome in stats/cmd/myflow/record.go, stats/internal/api/, stats/internal/store/records.go @ 441ed2b -->

## Impact

- `skills/myflow-do/SKILL.md` — the slot table's row 3, the panel-model paragraph's substitution
  clause, the `### Code review (low)` subsection, and one new subsection carrying both rules.
- `skills/myflow-do/SKILL-rationale.md` — the reversal and its evidence.
- `openspec/specs/myflow-review-panel-roster/` — the requirement asserting slot 3 invokes the skill
  is removed and replaced, since its heading states the reversed fact.
- `openspec/specs/myflow-review-panel-economics/` — two added requirements.

No budget change. Measured after the edits landed: `skills/myflow-do/SKILL.md` is 87,388 bytes
against a 103,578-byte budget, and `skills/myflow-do/SKILL-rationale.md` is 18,793 against 20,902 —
headroom of 16,190 and 2,109.
<!-- measured: task 4's verification run, recorded in verification.md beside this file -->
The change added 1,832 and 2,058 bytes, 3,890 in all. The planning estimate said "roughly 2 KB",
against pre-edit sizes of 85,556 and 16,735; the measured figure replaces it rather than being
reconciled to it.
<!-- measured: wc -c on both files at 441ed2b, before task 2's first edit -->

Nothing about the handoff bar moves: zero open findings at any severity, under every preset.
