# kan-263-myflow-forbid-backgrounded-builds-in-the-shared

## Why

Self-review finding from KAN-253 (angle: token/time cost). Four implementer agents ran a build with
`run_in_background` and then ended their turn waiting on it, producing no result — one did it
twice. Each stall cost a resume cycle. It was fixed ad hoc, per-prompt, after the third stall; the
instruction never made it into the reusable dispatch text.

## What changes

A new `FOREGROUND BUILDS` paragraph — forbidding a dispatched agent from ending its turn with a
build/test/long-running command still running in the background — is added once, as canonical
text, to the implementer dispatch preamble in `skills/flow/implement.md`. The two other sites that
already carry that preamble by citation (the per-task reviewer dispatch and the review panel's slot
dispatch) gain a one-line citation to it, the same pattern already used for `CONTEXT BUNDLE`. The
panel-fix subagent dispatch (`skills/flow/review-panel.md`), which is implementer-shaped and can
also run builds, gets the same citation.

No behavior outside dispatch-prompt text changes.
