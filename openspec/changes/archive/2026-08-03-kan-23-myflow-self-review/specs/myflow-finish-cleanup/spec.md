## ADDED Requirements

### Requirement: Run 2 invokes self-review after writing FINISHED

Immediately after `FINISHED` is written, run 2 SHALL invoke self-review as step 8. Self-review's own
behavior — the skip prompt, context gathering, the reasoning pass, per-finding Jira filing, the
operator rating, the report and the handoff line — is defined in full by the `myflow-self-review`
capability and is not restated here. This requirement states only the invocation point: after
`FINISHED`, before the terminal handoff prints.

Self-review SHALL NOT be able to prevent the `FINISHED` write, since it never runs before that write
succeeds, and a failure inside it SHALL NOT move the change off `FINISHED`.

#### Scenario: Step 8 follows the FINISHED write

- **WHEN** run 2 completes step 7 and `FINISHED` is written
- **THEN** step 8 (self-review, per the `myflow-self-review` capability) runs before the terminal
  handoff is printed

#### Scenario: A run 2 that never reaches FINISHED never runs self-review

- **WHEN** run 2 stops at step 6 on a cleanup leftover
- **THEN** step 8 does not run, and the change remains `IN_PROGRESS`
