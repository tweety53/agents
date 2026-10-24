# self-review-kan-1022-templates-step-2-be-own-templates-crud-from

## Why

The self-review pass over gymie's kan-1022 run (a backend web-layer change on `--base
feature/templates`) found six pipeline defects, none `big`:

- A1-1: `check-verify-report` admits no `n/a` on steps 8 and 9, so a verifier on a flagged diff
  that draws nothing (a `wire contract:` dispatch) could only pass by writing `done` over a capture
  it never ran — which the kan-1022 verifier did.
- A1-2: visual-verify step 3 creates a worktree only for the regression checkout, so a `--base`
  change with no frontend worktree had its stack serve the main frontend checkout on `develop`;
  round 1 spent 18m16s and failed 28 specs on the wrong frontend.
- A1-3: the "re-run only the failed specs" precondition (an unchanged HEAD) is checked by nothing;
  kan-1022 re-ran 19 failed specs after HEAD and the served frontend both changed, leaving 76 specs
  never run against the frontend actually served.
- A1-4: the review panel's reproducer authoring pattern has no rule that a check reads the
  defect's behaviour rather than one fix route's text; F7's reproducer grepped a fix's call.
- A1-5: the mutation brief never says a reproducer's tree is `$PWD`; F8's reproducer pointed at
  the slot's removed throwaway copy.
- A4-1: `check-unfinished-work` signal nine ignores a later report whose rerun of the failed specs
  passed; integrate needed a hand-recorded false positive.

## What changes

- `check-verify-report` admits `n/a — <reason>` on steps 8, 9 and 10 beside a `- wire contract:`
  line naming an endpoint; the verifier contract says so. The kan-1022 report is repaired.
- Visual-verify step 3 gives a change whose worktree records a flowBase a worktree in every
  `## apps` repository the set lacks.
- An environment cure that changes what the stack serves voids earlier passes, and
  `check-visual-verify-dispatched` reports MISSING for a rerun of failed specs recorded at another
  HEAD than the verifier row before it.
- Signal nine counts a failure addressed when the next report passed a rerun of the failed specs.
- `review-panel.md` gains the behaviour-not-fix-text rule and the mutation reproducer's `$PWD` rule.
