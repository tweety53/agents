# KAN-395 — fingerprint the served bundle before a screenshot counts as evidence

## Problem

`flow.visual-verify` step 4 probes the app URL and reuses whatever answers. A daemon left over
from an earlier round embeds an older SPA build (`stats/internal/web/dist` via `//go:embed`), so
every later capture screenshots the old bundle. During KAN-29's last fix round the verification
screenshot showed the very bug the fix had removed; it was caught only because the stack happened
to be restarted and re-checked. Nothing between the probe and the capture asks what is being
served.

## Design

**An optional `fingerprint` command row** in `## visual verification`'s commands table. The
project declares a command that exits 0 when the served bundle is the worktree's own build and
non-zero otherwise. The closed `Command` vocabulary becomes `setup`, `verify`, `capture`,
`fingerprint`. Absent is not a violation — a project with no served bundle declares nothing — and
`check-visual-verification.sh` stays silent about it, exactly as it does for `setup`.

**A new stage step between the probe (step 4) and `verify` (step 5).** With a `fingerprint` row
declared, the verifier runs it. Exit 0 continues. Non-zero: stop the stack, start it from `## run`
— recorded as stage-started so the final step stops it — and run `fingerprint` again. A second
non-zero **blocks** the `IN_PROGRESS` handoff with the command's output. With no row declared, the
step reports `fingerprint: not declared` and continues.

**Report row.** `- fingerprint: not declared | exit 0 | mismatch → restarted → exit <n>`, and
`Blocking` names the second mismatch.

**This repository's row.**

```text
cd stats/web && npm run build && curl -sf http://127.0.0.1:4174/ | cmp -s - ../internal/web/dist/index.html
```

Vite content-hashes its asset filenames and rewrites `index.html` to reference them, so
`index.html` is the bundle's fingerprint; `web.Handler` serves the embedded file byte-for-byte at
`/`. Rebuilding first is what makes the comparison mean "the worktree's source", not "whatever
`dist/` last held" — `make ui-test-up` rebuilds too, so a stage-started stack matches on the
first check and only a reused stack pays the restart.

## Alternatives

- **Unconditional restart in step 4** — smallest diff, but trusts `## run` to rebuild rather
  than proving the served bundle matches, and drops "a stack the operator had running is left
  alone".
- **A shipped `check-served-bundle.sh <url> <file>`** — a script plus harness to replace one
  project-declared line; the comparison is project-specific by nature.
- **Required `fingerprint` row** — would break lint in every consuming project without a served
  bundle.

## Files

- `skills/flow/verify-and-handoff.md` — the step, report row, `Blocking`
- `skills/flow-contracts/project-configuration.md` — the row, "three rows" → four
- `scripts/check-visual-verification.sh` — vocabulary; `scripts/test-check-visual-verification.sh`
  — one case: `fingerprint` accepted, absent is silent
- `.flow/project.md` — this repository's row
- `scripts/check-contract-budget.sh` — only if a file outgrows its budget
