# kan-395-flow-fingerprint-the-served-bundle-before-a

Linked Jira issue: KAN-395 — from KAN-29's self-review, problems-encountered angle.

## Why

`flow.visual-verify` step 4 probes the app URL and reuses whatever answers. A daemon left over
from an earlier round embeds an older SPA build (`stats/internal/web/dist` via `//go:embed`), so
every later capture screenshots the old bundle. During KAN-29's last fix round the verification
screenshot showed the very bug the fix had removed; it was caught only because the stack happened
to be restarted and re-checked. Nothing between the probe and the capture asks what is being
served, so a screenshot of a stale build counts as evidence today.

## What changes

- `## visual verification`'s commands table accepts an optional fourth command, `fingerprint`,
  that exits 0 when the served bundle is the worktree's own build. Absent is not a violation.
- `flow.visual-verify` runs it between the probe and `verify`. A mismatch restarts the stack from
  `## run` and re-checks once; a second mismatch blocks the `IN_PROGRESS` handoff with the
  command's output. The verifier report carries a `fingerprint:` row, `not declared` when no row
  exists.
- This repository declares its row: rebuild the SPA, then compare the served `/` with the
  worktree's `stats/internal/web/dist/index.html`.
