# Design — kan-395-flow-fingerprint-the-served-bundle-before-a

## Context

Bounded change; `proposal.md` carries why, this file carries how. The stage that owns the
procedure is **Visual verification** in `skills/flow/verify-and-handoff.md`; the section it reads
is canonical in `skills/flow-contracts/project-configuration.md` and mechanically enforced by
`scripts/check-visual-verification.sh`, whose `Command` vocabulary is closed.

## The `fingerprint` command

One optional row in the commands table:

| Command | Required | Meaning |
|---------|----------|---------|
| `fingerprint` | no | Exits 0 when the app the probe answered from is serving the worktree's own build, non-zero otherwise. The project owns how — what a bundle's identity is differs per stack. |

`check-visual-verification.sh` admits the name and requires nothing of it, exactly as it treats
`setup`. Its violation message for an unknown command names all four.

## The stage step

Inserted after the probe/start step and before `verify`, renumbering the rest; the verifier runs
it as part of its steps:

1. No row declared → report `fingerprint: not declared`, continue.
2. Run `fingerprint`. Exit 0 → report `fingerprint: exit 0`, continue.
3. Non-zero → stop the stack, start it from `## run`, and record that this stage started it (so
   the final stop step stops it). Run `fingerprint` again.
4. Exit 0 → report `fingerprint: mismatch → restarted → exit 0`, continue. Non-zero → report
   `fingerprint: mismatch → restarted → exit <n>` with the command's output, and **block**.

The report template gains the `- fingerprint:` row; **Blocking** gains "a `fingerprint` that
still fails after the restart".

## This repository's row

```text verified:the cell as written in task 4 of tasks.md, run against the working tree with the four diffs applied — check-visual-verification.sh . VISUAL-OK at c93da10
cd stats/web && npm run build && curl -sf http://127.0.0.1:4174/ | cmp -s - ../internal/web/dist/index.html
```

Vite content-hashes its asset filenames and rewrites `index.html` to reference them, so
`index.html` is the bundle's fingerprint; `web.Handler` serves the embedded file byte-for-byte at
`/`. Rebuilding first makes the comparison mean "the worktree's source" rather than "whatever
`dist/` last held". `make ui-test-up` rebuilds too, so a stage-started stack matches on the first
check and only a reused stack pays the restart.

## Files

- `skills/flow/verify-and-handoff.md` — the step, report row, `Blocking`
- `skills/flow-contracts/project-configuration.md` — the row; "three rows" becomes four
- `scripts/check-visual-verification.sh` — vocabulary and message
- `scripts/test-check-visual-verification.sh` — one case: `fingerprint` accepted, absent is silent
- `.flow/project.md` — this repository's row
- `scripts/check-contract-budget.sh` — only if a file outgrows its budget

## Decisions

### Detect staleness with a project-declared fingerprint, then restart once

**ID:** fingerprint-restart-then-block
**Status:** active
**Chosen:** optional `fingerprint` command; a mismatch restarts the stack from `## run` and
re-checks, a second mismatch blocks — the served bundle is proven, not assumed, and the one
repair that fixed KAN-29 runs automatically.
**Considered:** unconditional restart in step 4 — smallest diff, but trusts `## run` to rebuild
rather than proving the served bundle matches, and drops "a stack the operator had running is left
alone". A shipped `check-served-bundle.sh <url> <file>` — a script plus harness to replace one
project-declared line, for a comparison that is project-specific by nature. Block on the first
mismatch without restarting — every fix round against a reused stack would block once for the
operator to restart by hand.

### The row is optional and its absence is silent

**ID:** fingerprint-optional
**Status:** active
**Chosen:** absent is not a lint violation; the stage reports `fingerprint: not declared` so the
gap is visible in every handoff.
**Considered:** required — breaks lint in every consuming project without a served bundle
(CLI-only, static site).

## Open questions
