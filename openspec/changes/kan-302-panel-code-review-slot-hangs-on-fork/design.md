# Design — kan-302-panel-code-review-slot-hangs-on-fork

## Where the rules live

Both new rules are prose in `skills/myflow-do/SKILL.md` §5, enforced by the dispatching agent.

That is the only layer that can enforce them. A guard script has no handle on an in-flight subagent
— it could only read a timestamp file the agent itself writes, which is the agent enforcing the
ceiling with extra steps and one more thing to go stale. The harness's own stall watchdog cannot
enforce it either: round 4's evidence is precisely that the watchdog does not fire on a slot
emitting output while making no progress, and myflow is installed into three harnesses whose
watchdogs it does not configure.

The ceiling is therefore stated against the mechanism, not against one harness's tool — the same
form **Progress visibility** (`skills/myflow-contracts/pipeline.md`) already uses for the task list.
What the rule requires is that the dispatcher not block indefinitely and that it know each in-flight
slot's elapsed time; how a given harness lets it check is that harness's business.

## Rule 1 — no panel slot is dispatched onto a forking skill or agent

A slot whose work is forked to a background agent cannot report through the panel's contract. The
parent never observes that agent's completion, so the slot's findings either arrive on a surface the
panel does not read — which is what round 3 did, delivering 6 findings to the session while the slot
itself stalled — or never arrive at all.

Stated generally rather than as a property of slot 3, because the defect is a property of forking,
not of one skill. Slots 1 and 4 are dispatched by `subagent_type` onto agent definitions this
pipeline does not read; the rule is what a future dispatcher can cite before pointing a slot at one
that forks.

## Rule 2 — every panel slot carries a wall-clock ceiling

**15 minutes.** The observed good run was 2.5 minutes; the harness's stall watchdog is 600s and did
not fire; the hung slot ran 4 hours.
<!-- measured: KAN-295 panel rounds 3-4 and the re-dispatch after round 4, recorded in KAN-302 -->
15 minutes clears a slow pass over a large `final-review.diff` with room to spare and is far below
anything a reader would call "still working."
<!-- predicted: no pass over a large final-review.diff has been timed; 15 is chosen as a ceiling
     well clear of the one good run measured, not derived from a distribution -->


On a breach, in order: stop the slot; close its dispatch row with `-outcome timed-out`; record the
breach in the panel record; re-dispatch that one slot once. A second breach on the same slot stops
and puts the choice to the operator.

The ladder is a retry, not a waiver. A slot that never returned cannot be dropped — **No preset
moves the handoff bar** (`skills/myflow-do/SKILL.md`) already fixes zero open findings at any
severity, and "an unavailable harness feature is not a way to weaken review" is the existing
statement of the same principle. So proceeding without the slot is only ever the operator's explicit
choice, and the panel record names the slot as not run when they make it.

`-outcome timed-out` costs nothing: the flag is a free-form string in the CLI, the API and the
store, with no enum anywhere on the path.
<!-- verified: stats/cmd/myflow/record.go:450 declares -outcome as a free string; no Outcome validation in stats/internal/api/ or stats/internal/store/records.go @ 441ed2b -->

## Rule 2's interaction with the fix-round ladder

A timed-out slot is not a finding. It raises none, so it adds nothing to the union of open findings
and consumes no fix round. What it does affect is the final pass's requirement of a non-stale clean
result for every slot in the roster: a slot the operator chose to proceed without has no clean
result, and the panel record saying so is what keeps that visible rather than reading as a clean
panel.

## Slot 3's new shape

Unconditionally a `general-purpose` subagent on `models.reviewPanel` — Sonnet by default — briefed
to report high-confidence defects only, reporting `F<n>` rows and marker lines like every other
slot. Its model is named by the dispatcher, so its dispatch row records that real model and never
`unknown (agent-defined)`, exactly as before.

Everything the old subsection said about the skill goes: the invocation, the effort level, the
"return findings in your report back rather than where the skill displays them" instruction (there
is no longer a host surface to displace them to), and the harness-availability fallback.

## Decisions

### Slot 3's skill path is deleted rather than kept as an opt-in

**ID:** `slot3-delete-skill-path`
**Status:** active
**Chosen:** Delete it — slot 3 is unconditionally the plain reviewer. One shape, no branch, no
substitution to record, and no way to reach the hang again.
**Considered:** Keep the skill path behind an operator instruction or a recorded field — rejected
because it preserves a branch the evidence identifies as the broken one, and a route back to a hang
is not worth the deeper reading the skill might have offered.

### The ceiling is 15 minutes
<!-- predicted: 15 is a chosen ceiling, not a measured duration -->

**ID:** `ceiling-fifteen-minutes`
**Status:** active
**Chosen:** 15 minutes — clears the 2.5-minute observed good run and a slow pass over a large diff,
well under anything that reads as stuck.
<!-- measured: the 2.5-minute figure is KAN-295's re-dispatch; the clearance is predicted, no large-diff pass having been timed -->
**Considered:** 30 minutes — rejected as too long a wait before anyone learns a slot is hung.
<!-- predicted: 30 and 10 are the bracketing alternatives considered, not measured outcomes -->
10 minutes — rejected as too close to the harness's own 600s watchdog, risking the kill of a
<!-- measured: the 600s threshold is the harness's own, observed on KAN-295 round 3 -->
legitimately slow pass over a large `final-review.diff`.


### A breach re-dispatches once before asking

**ID:** `breach-retry-once`
**Status:** active
**Chosen:** Stop, record, re-dispatch that one slot once; a second breach asks the operator. The
ticket's own evidence is a re-dispatch on the fallback shape returning in 2.5 minutes, so one
automatic retry is the cheapest thing that would have resolved the observed incident.
<!-- measured: KAN-295, the re-dispatch that followed round 4 -->

**Considered:** Ask on the first breach — rejected as an operator interrupt for something one retry
usually clears. Stop the run on the first breach — rejected because it turns one flaky slot into a
stopped run with the implementation committed and the panel incomplete.

### The second-breach prompt's safe default is stop the run

**ID:** `breach-prompt-default-stop`
**Status:** active
**Chosen:** Stop the run, marked recommended, and what silence selects.
**Considered:** Naming no default at all — rejected, and this was the design's own defect: an
earlier draft said the dispatcher "makes no choice of its own here," which contradicts **Operator
prompts** (`skills/myflow-contracts/operator-prompts.md`) outright. That contract is canonical for
prompt shape and requires exactly one option marked recommended, named as what silence selects. A
disclosed default that fires with a ⚠ marker is not the dispatcher's own judgment; it is the call
site's declared safe answer, which is the thing that keeps a breach from being resolved silently.
Re-dispatch again — rejected as the default because silence would loop. Proceed without the slot —
rejected as the default because silence would weaken review, which **No preset moves the handoff
bar** forbids. Found by task 3's per-task reviewer, as finding F1.

### The ceiling reaches panel dispatches only

**ID:** `ceiling-scope-panel-only`
**Status:** active
**Chosen:** §5's slots. That is where the evidence is, and implementer dispatches are serialised per
worktree and report a commit sha, so a hang there is visible by a different route.
**Considered:** Every subagent dispatch, implementers and panel-fix subagents included — rejected as
speculative: no evidence yet that implementers need it, and it would touch a second section for a
defect nobody has observed there.

### The forking rule is stated generally, not as a slot-3 fact

**ID:** `forking-rule-general`
**Status:** active
**Chosen:** A normative sentence in §5 covering any slot, with slot 3's change as its application.
**Considered:** Fix slot 3 alone — rejected because the next slot pointed at a forking skill
reintroduces the same hang with nothing to cite against it. The general rule costs one sentence.

### The slot keeps its name

**ID:** `slot3-keep-name`
**Status:** active
**Chosen:** `Code review (low)`. It still describes what the slot does, and it is the value
`myflow record dispatch begin -slot` writes into the store, so keeping it leaves the existing
dispatch history comparable with new rows.
**Considered:** Rename it to drop the implied skill and effort level — rejected because it splits
the store's `-slot` values across old and new runs, forcing any query over dispatch history to know
both names, for a gain that is only in how the name reads.

### The old rationale is recorded as reversed, not deleted

**ID:** `rationale-records-reversal`
**Status:** active
**Chosen:** `SKILL-rationale.md` keeps why the `code-review` skill was chosen and records the
KAN-295 evidence that reversed it.
**Considered:** Delete the paragraph — rejected: **Cut, never paraphrase** (`rules/be-brief.mdc`)
names a recorded reason a rejected alternative was rejected as one of the things never cut, and the
reason the skill looked right is exactly what stops it being re-proposed.

## Open questions

*(none — every question this stage raised was answered)*
