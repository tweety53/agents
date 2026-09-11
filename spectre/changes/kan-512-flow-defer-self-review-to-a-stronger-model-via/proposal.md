# kan-512-flow-defer-self-review-to-a-stronger-model-via

## Why

Run 2's step 9 always runs the self-review reasoning pass synchronously, on `SELF_REVIEW_MODEL`,
inside the same archive dispatch — there is no way to save the gathered context and run the
reasoning pass later, on a stronger model, without losing it. KAN-512 asks for a deferred path.

## What changes

`## self review` gains a third literal, `defer`: step 9 writes the gathered bundle — the four
sources `gather-self-review-context.sh` already gathers, plus the archived `design.md` and the
change's own per-invocation narrative — to one committed file,
`docs/self-review/<name>-context.md`, and dispatches nothing. A new standalone
`/flow-self-review <name>` command reads that bundle, runs the five-angle pass inline on whatever
model the session is on, files and rates findings, writes the report, deletes the bundle, and
lands both on the default branch. A new `<changeRoot>/narrative.md`, appended at
`flow.write-in-progress` and `flow.preserve-sessions`, carries the parent session's own account of
each run so a deferred pass is not limited to the four static sources. `defer` becomes the default
`## self review` setting for `agents` and `gymie`.
