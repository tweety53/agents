## REMOVED Requirements

### Requirement: A script performs the mechanical state write

**Reason**: The seven commands the script served — `/myflow-start-done`,
`/myflow-do-manual-review`, `/myflow-do-done`, `/myflow-do-fix-manual-review`,
`/myflow-do-fix-done`, `/myflow-manual-test-done`, `/myflow-review-done` — no longer exist. In
the three-state model the human gate is a property of the state, so there is no confirmation for
a command to record and nothing for a script to advance.

**Migration**: None required. Advancing happens as a side effect of the next pipeline command:
run `/myflow-do` or `/myflow-finish` where a `*-done` command was previously run. `state-advance.sh` is deleted with the `myflow-state-advance` skill directory that carries it.

### Requirement: The script escalates with distinct exit codes

**Reason**: The escalation contract existed so a cheap script could handle the happy path and
hand judgment cases to a skill. With the script and its callers removed, there is no caller to
escalate and no skill to escalate to.

**Migration**: None required. State validation now happens inside each pipeline command, which
already loads the state machine contract for its own work.

### Requirement: Dynamic target resolution matches the fix re-entry contract

**Reason**: Fix re-entry is removed. `/myflow-do` never moves the state except from `STARTED` to
`IN_PROGRESS`, so there is no `originStage` to resolve a dynamic target from.

**Migration**: None required. A fix leaves the state unchanged; re-run an earlier command to
re-open an earlier gate.

### Requirement: Commands run the script first and the skill only on escalation

**Reason**: Neither the script nor its commands survive.

**Migration**: None required.

### Requirement: Self-heal narrowing is documented at the point of change

**Reason**: This requirement constrained how the script's narrowed self-heal was documented.
Self-heal now lives entirely in the state machine contract, documented once there.

**Migration**: None required. Self-heal behaviour is specified in the `myflow-state-machine`
capability.
