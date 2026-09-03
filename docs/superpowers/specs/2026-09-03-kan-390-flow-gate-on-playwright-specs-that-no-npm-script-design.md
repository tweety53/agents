# kan-390-flow-gate-on-playwright-specs-that-no-npm-script — design

## Context

KAN-390, from KAN-29's self-review: seven route A–F capture suites in `gymie-playwright` never
ran once. `test:e2e` reaches `tests/`, `test:e2e:full` reaches `tests-full/`, and the root
`kan-*.spec.ts` files matched neither; one (`kan-347`) passed green against a 404 page for its
whole life. The ticket proposes enumerating every `*.spec.ts` and diffing it against what the
project's scripts reach, failing a gate on orphans.

No staged research note existed (`docs/superpowers/research/kan-390.md`, `kan-390-*.md` absent).
`spectre/specs/` is empty, so no capability spec is edited.

Measured against `main` at `d31ebd9`:

- `gymie-playwright` is gymie's `regression checkout` (`## visual verification`,
  `/Users/tweety53/Projects/gymie/.flow/project.md`). Its `verify` is `npm run test:visual`,
  reaching `tests/visual-baseline.spec.ts` alone; `flow.visual-verify`'s `capture` writes a
  per-change `kan-<n>-*.spec.ts` at the checkout root, which nothing named in `.flow/project.md`
  runs afterwards.
- `gymie-playwright` commit `c183ddf` added `test:e2e:routes: playwright test kan-*.spec.ts`.
  That works only because npm runs the script through `sh`, which expands the glob before
  Playwright sees it: Playwright treats every positional argument as a regular expression matched
  against absolute paths (`packages/playwright/src/program.ts`, help text), and quoted
  `'kan-*.spec.ts' --list` lists 0 files. A static re-implementation of "which globs the scripts
  reach" would have to model both shell expansion and Playwright's regex filter, and would have
  called that script unreachable.
- `npm run -s <script> -- --list --reporter=json` lists reached files without a browser, in
  about 0.6s per script (Playwright 1.62.1). The JSON carries `config.rootDir` and one
  `suites[].file` per file, relative to `rootDir`.
- This repository's own Playwright dir, `stats/web`, runs a bare `playwright test` (`test:visual`)
  that reaches everything under `testDir`; it has no orphan problem and no key naming it.
- `skills/flow/verify-and-handoff.md` is 26051 bytes against a 26100-byte row in
  `scripts/check-contract-budget.sh`; any added prose raises that row.

## 1. The guard

`scripts/check-spec-reach.sh <project-root>`, shipped: symlinked into `skills/flow/scripts/`,
cited by basename per **Guard resolution** (`skills/flow-contracts/pipeline.md`), listed in
`skills/flow/SKILL.md`'s guard-presence union.

Procedure:

1. Resolve `regression checkout` from `<project-root>/.flow/project.md`'s `## visual
   verification` settings table — the guard's own awk over `lib/visual-table-cells.awk`, the same
   heading regex and "a heading at or above this section's level ends it" rule
   `check-visual-trigger.sh` and `resolve-visual-screenshots.sh` each carry, a BOM stripped via
   `lib/strip-bom.sh` first. A missing `.flow/project.md`, no `## visual verification` section,
   or no `regression checkout` row prints `Spec reach: not configured` and exits 0.
2. Enumerate `*.spec.ts` under the checkout with `find`, pruning `node_modules` and `.git`.
3. Read `package.json` `scripts` (via `node -e`, since node is present wherever Playwright is —
   no Python widening) and keep every script whose command contains the substring
   `playwright test`.
4. For each kept script, run `npm run -s <script> -- --list --reporter=json` from the checkout,
   with `PLAYWRIGHT_JSON_OUTPUT_NAME`, `PLAYWRIGHT_JSON_OUTPUT_DIR` and `PLAYWRIGHT_JSON_OUTPUT_FILE`
   unset so the reporter writes stdout; union `config.rootDir` joined with each `suites[].file`.
5. Print `<relative path>: reached by no package.json script` per orphan, or `Spec reach: <N>
   spec(s), all reached`.

Exit codes:

| Exit | Meaning |
|------|---------|
| 0 | every enumerated spec is reached, or the project is not configured (`Spec reach: not configured` printed) |
| 1 | at least one orphan, one line each |
| 2 | cannot answer: `<project-root>` not a directory; `## visual verification` declared twice; the checkout is not a directory; no `package.json` in it; no script contains `playwright test`; `node` or `npm` absent; a `--list` run exits non-zero (its output is relayed) or prints unparsable JSON |

The input is attacker-influenced exactly as `check-visual-verification.sh`'s header states:
`.flow/project.md` and the checkout's `package.json` are tracked and editable in any pull request.
The guard runs `npm run` on scripts the project already runs through `verify` and `capture` — the
same trust level, no widening — and every interpolated value reaching an error message passes
through `lib/sanitize-display.sh`.

Stated limits, in the header: a script reaching Playwright indirectly (`npm run other`) is not
followed and its files count as orphans; `.test.ts`, `.spec.js` and `.spec.mjs` are not enumerated
— the ticket names `*.spec.ts`, and Playwright's wider default `testMatch` would pull `stats/web`'s
vitest `*.test.ts` files into a Playwright question.

## 2. The two call sites

`flow.verify` (`skills/flow/verify-and-handoff.md`): the conductor appends `check-spec-reach.sh
<worktree>` after the `## test` commands in each verifier's ordered command list. It is one more
command in that list: the verifier runs it in order, reports its exit and output in `## Report`,
and a non-zero exit blocks the handoff exactly as a failed command does. Not configured is exit 0
with its one line, so a project with no `regression checkout` passes through.

`flow.visual-verify`, step 6: after `capture` succeeds, run `check-spec-reach.sh <worktree>`; exit
1 or 2 blocks, added to **Blocking**'s list and as a `- spec reach: exit <n>` report line. Steps
keep their numbers.

Both call sites, so a pre-existing orphan surfaces on every change and the spec this change's
`capture` just wrote must be reachable before the stage passes.

## 3. Repository plumbing

- `skills/flow/scripts/check-spec-reach.sh -> ../../../scripts/check-spec-reach.sh`
  (`check-guard-symlinks.sh` rule 2; `lib/` is already a directory symlink).
- `skills/flow/SKILL.md` guard-presence list gains `check-spec-reach.sh`.
- `.flow/project.md` `## lint` gains `scripts/check-spec-reach.sh .` — this repository declares no
  `regression checkout`, so the line exercises the not-configured path, exit 0.
- `skills/flow-contracts/project-configuration.md` `## visual verification`: one sentence noting
  `regression checkout` is also the root `check-spec-reach.sh` scans.
- `scripts/check-contract-budget.sh` rows raised for every owned file this change grows past its
  row, to the new size plus 25% per the guard's own convention.
- `scripts/test-check-spec-reach.sh`: fixture trees with a `.flow/project.md`, a checkout carrying
  `package.json` and spec files, and a fake `npm` first on `PATH` that prints canned JSON per
  script name — exercising script selection, the union, orphan reporting and every exit code
  without Playwright installed. `run-guard-tests.sh` discovers it by glob. One manual smoke against
  `/Users/tweety53/Projects/gymie` (its checkout with real Playwright) is a plan verification step,
  not a harness case.

## Decisions

- **`--list` over static glob analysis.** Ground truth from Playwright, including shell expansion,
  regex filters, `testDir`, `testMatch` and `testIgnore`. Rejected: parsing script arguments as
  globs — `test:e2e:routes` proves it wrong on the first real case.
- **Both call sites** over `flow.visual-verify` only (legacy orphans surface only on UI-touching
  changes) and over change-start only (the ticket's literal wording; a captured spec would be caught
  one change late). Operator's choice.
- **`regression checkout` only** as the scan root, no new key. Rejected: auto-detecting every
  `## apps` root with `@playwright/test` (more guard code, false-block risk on a non-suite
  Playwright install); a new `## e2e spec roots` section (gymie declares the same path twice).
  Operator's choice; add a key when a second project needs one.
- **Not configured is exit 0**, unlike `check-visual-trigger.sh`'s exit 2, because at both call
  sites the guard is one command in a list where any non-zero blocks; "nothing to check" must not.
- **`node -e` for JSON**, not Python: node is a precondition of the checkout being a Playwright
  checkout at all, and this repository records every Python widening as deliberate.
- **`*.spec.ts` only**, per the ticket; see section 1's limits.
- **Own awk extraction** rather than factoring `resolve-visual-screenshots.sh`'s into a lib: the
  repository already accepted per-guard copies of the section scan (`check-visual-trigger.sh`),
  and refactoring a tested guard is outside this change.

## Open questions

None.
