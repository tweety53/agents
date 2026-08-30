# kan-304-myflow-ship-an-unwrapped-citation-scanner-and

## Why

`scripts/check-references.sh` (built for KAN-301) already scans this repo for citations that name
a moved section but still point at the old file. It runs project-wide under `## lint`, but only at
`/flow`'s `flow.verify` stage — after the review panel has already dispatched. Reviewers never see
its output, so a stale citation surfaces only after the panel has finished, costing a fix round it
could have caught up front. KAN-304 (filed by self-review of KAN-295, angle: what could be
automated) asks to surface this before the panel.

## What changes

- A new optional `.flow/project.md` key, `## review panel citation check`, names a single command
  to run before the review panel dispatches.
- A new guard, `check-panel-citation-trigger.sh`, decides whether to run it: only when the change's
  diff touches at least one `.md`/`.mdc` path.
- `skills/flow/review-panel.md` runs the configured command when triggered, captures its output
  verbatim, and hands every panel slot a pointer to it alongside the diff. The command's exit code
  never blocks dispatch — the real gate stays `flow.verify`'s existing lint run.
- This repo's own `.flow/project.md` declares the new key with `scripts/check-references.sh` —
  reusing the existing scan rather than writing a second one.
