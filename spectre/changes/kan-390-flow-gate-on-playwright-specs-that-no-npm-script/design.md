# kan-390-flow-gate-on-playwright-specs-that-no-npm-script — design

## Context

Approved design: `docs/superpowers/specs/2026-09-03-kan-390-flow-gate-on-playwright-specs-that-no-npm-script-design.md`
— canonical for the guard's procedure, exit-code table, call sites, plumbing and the measured
context. This file carries the decision and open-question records the pipeline reads.

KAN-390, from KAN-29's self-review: seven route A–F capture suites in `gymie-playwright` never ran
once — `test:e2e` reaches `tests/`, `test:e2e:full` reaches `tests-full/`, and the root
`kan-*.spec.ts` files `flow.visual-verify`'s `capture` writes matched neither. One (`kan-347`)
passed green against a 404 page for its whole life. Nothing in the pipeline asks whether a spec it
just wrote is reachable by anything the project runs.

## Decisions

### Reachability comes from Playwright's own `--list`, not from parsing script globs

**ID:** `reach-via-list`
**Status:** active
**Chosen:** run `npm run -s <script> -- --list --reporter=json` per `playwright test` script and
union `rootDir` + `suites[].file` — ground truth including shell expansion, regex filters,
`testDir`, `testMatch` and `testIgnore`.
**Considered:** static parsing of script arguments as globs — `gymie-playwright`'s
`test:e2e:routes` (`playwright test kan-*.spec.ts`) works only because `sh` expands the glob, and
Playwright treats the argument as a regex; a static reading calls that script unreachable on the
first real case.

### The gate runs at both `flow.verify` and `flow.visual-verify`

**ID:** `gate-at-both-stages`
**Status:** active
**Chosen:** `flow.verify` appends the guard to each verifier's command list, so a pre-existing
orphan surfaces on every change; `flow.visual-verify` step 6 runs it after `capture`, so the spec
this change just wrote must be reachable before the stage passes. Operator's choice.
**Considered:** `flow.visual-verify` only — legacy orphans surface only on UI-touching changes;
change start only (the ticket's literal wording) — a captured spec is caught one change late.

### The scan root is `regression checkout` and nothing else

**ID:** `scan-regression-checkout-only`
**Status:** active
**Chosen:** no new config key; a project without a `regression checkout` prints `Spec reach: not
configured`. Covers the one real failure. Operator's choice.
**Considered:** auto-detecting every `## apps` root carrying `@playwright/test` — more guard code
and a false-block risk on a Playwright install that is not a spec suite; a new `## e2e spec
roots` section — gymie would declare the same path twice. Add a key when a second project needs
one.

### Not configured is exit 0

**ID:** `not-configured-exit-0`
**Status:** active
**Chosen:** the guard prints `Spec reach: not configured` and exits 0, so both call sites treat it
as one more command in a list where any non-zero blocks.
**Considered:** exit 2 as `check-visual-trigger.sh` does — that guard's caller distinguishes its
exits by hand; here the verifier runs a flat command list and "nothing to check" must not block.

### JSON through `node -e`, not Python

**ID:** `json-via-node`
**Status:** active
**Chosen:** `node -e` reads `package.json` scripts and the reporter output — node is a
precondition of the checkout being a Playwright checkout at all.
**Considered:** `python3` — this repository records every Python widening of its Bash guard
toolchain as deliberate (`check-plan-provenance.py`), and none is needed here.

### Enumerate `*.spec.ts` only

**ID:** `spec-ts-only`
**Status:** active
**Chosen:** `*.spec.ts`, per the ticket.
**Considered:** Playwright's default `testMatch` (`.test.ts`, `.spec.js`, `.spec.mjs`, …) — would
pull `stats/web`'s vitest `*.test.ts` files into a Playwright question.

### Own section-extraction awk, not a shared lib

**ID:** `own-awk-extraction`
**Status:** active
**Chosen:** the guard carries its own `regression checkout` extraction over
`lib/visual-table-cells.awk`, the pattern `check-visual-trigger.sh` already set.
**Considered:** factoring `resolve-visual-screenshots.sh`'s extraction into `scripts/lib/` —
refactoring a tested guard is outside this change.

## Open questions

None.
