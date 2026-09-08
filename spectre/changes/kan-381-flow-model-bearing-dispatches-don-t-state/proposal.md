# kan-381-flow-model-bearing-dispatches-don-t-state

## Why

KAN-381 (self-review finding from KAN-380, problems-and-fixes angle): at archive time the
dispatcher read the settings store's raw `selfReviewModel` value instead of re-resolving
`SELF_REVIEW_MODEL` through the three-tier order, nearly dispatching the self-review subagent on
`opus` instead of the resolved `fable` — caught only because the operator asked "why opus?"
mid-run. The handshake cannot catch this class: the dispatcher believed the wrong value was
right, so the subagent's honest `Model:` line matched the wrong expectation and the run would
have proceeded silently on the wrong model.

## What changes

- Every model-bearing dispatch site states its resolved value and the tier that produced it in
  the run's own output, immediately before dispatching: `<role> model: <value> (<tier>)`.
- The rule is stated once, in `skills/flow/SKILL.md`'s Model resolution; each dispatch site adds
  one sentence instantiating its own role token.
- Sites covered: the planner dispatch (creating run and `flow.document-fix`), the conductor,
  every review-panel slot per round, the panel-fix subagent, both verifiers, the self-review
  dispatch (step 9 and its mismatch re-dispatch), and `/flow-research`'s research subagent.
