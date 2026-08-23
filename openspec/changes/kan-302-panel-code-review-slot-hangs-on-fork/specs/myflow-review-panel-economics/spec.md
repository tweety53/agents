# myflow-review-panel-economics delta — kan-302-panel-code-review-slot-hangs-on-fork

## ADDED Requirements

### Requirement: No panel slot is dispatched onto a skill or agent that forks

A panel slot SHALL NOT be dispatched onto a skill or an agent that forks its own background agent to
do the work.

The parent dispatcher never observes a forked agent's completion. A slot dispatched that way
therefore cannot report through the panel's contract: its findings arrive on a surface the panel does
not read, or they do not arrive at all, and either way the slot itself neither returns a result nor
reliably ends.

This SHALL be read as a property of forking rather than of any one skill. Where a slot's intended
skill or agent is found to fork, the correct repair is to dispatch that slot on a shape that does
not fork — never to drop the slot, which **No preset moves the handoff bar** forbids.

#### Scenario: A forking skill is not dispatched

- **WHEN** a panel slot's intended skill or agent forks its own background agent
- **THEN** the slot is not dispatched onto it
- **AND** the slot runs on a shape that reports back to the dispatcher directly
- **AND** the slot is not dropped and the panel still runs its full required roster

### Requirement: A panel slot has a wall-clock ceiling and is checked for liveness

Every panel slot SHALL carry a wall-clock ceiling of **15 minutes** from its dispatch.

The dispatcher SHALL NOT block indefinitely on a slot's completion notification. It SHALL track each
in-flight slot's elapsed time and act on the ceiling itself, so that a slot which emits output while
making no progress — the case a stall watchdog cannot see — is still bounded.

This requirement SHALL be satisfied by whatever mechanism the running harness provides for observing
an in-flight dispatch, and no harness has to gain one for the requirement to hold. What is not
optional is that the dispatcher knows a slot's elapsed time and enforces the ceiling; a dispatcher
that waits on a notification alone has dropped the requirement rather than adapted it.

On a breach, the dispatcher SHALL, in order: stop the slot; close its dispatch row with
`-outcome timed-out`; record the breach in the panel record, naming the slot and the elapsed time;
and re-dispatch that one slot once. Other slots SHALL be unaffected.

A second breach of the same slot SHALL stop and put the choice to the operator — re-dispatch again,
proceed without the slot, or stop the run — and SHALL NOT be resolved silently.

That prompt SHALL take the shape **Operator prompts**
(`skills/myflow-contracts/operator-prompts.md`) fixes, which requires exactly one option marked
recommended and names it as what silence selects. **Stop the run SHALL be that option.**
Re-dispatching again on silence risks a loop, and proceeding without the slot on silence weakens
review, so neither may be what an unanswered prompt resolves to. Stopping ends the run at
`IN_PROGRESS` with the implementation committed and nothing lost.

A timed-out slot raises no finding, consumes no fix round, and SHALL NOT be counted as a clean
result. Where the operator chooses to proceed without a slot, the panel record SHALL name that slot
as not run, so a panel missing a slot is never indistinguishable from a clean one.

#### Scenario: A slot exceeds the ceiling and is re-dispatched

- **WHEN** a panel slot has run for 15 minutes without returning
- **THEN** the slot is stopped
- **AND** its dispatch row is closed with `-outcome timed-out`
- **AND** the panel record names the slot and its elapsed time
- **AND** that one slot is re-dispatched, with every other slot unaffected

#### Scenario: A second breach asks the operator

- **WHEN** the same slot exceeds the ceiling a second time
- **THEN** the dispatcher stops and asks the operator whether to re-dispatch again, proceed without
  the slot, or stop the run
- **AND** stop the run is the option marked recommended, and is what silence selects
- **AND** the breach is never resolved silently

#### Scenario: Proceeding without a slot is recorded, never silent

- **WHEN** the operator chooses to proceed without a timed-out slot
- **THEN** the panel record names that slot as not run
- **AND** the slot is not counted as a clean result for the final pass
- **AND** the zero-open-findings bar still governs every slot that did run

#### Scenario: A slot emitting output makes no difference

- **WHEN** a slot has emitted output within the last minute but has run for 15 minutes without
  returning
- **THEN** the ceiling is enforced exactly as for a silent slot
