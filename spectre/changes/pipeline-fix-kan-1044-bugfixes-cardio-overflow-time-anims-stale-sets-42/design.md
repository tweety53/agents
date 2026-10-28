## Context

`stats/web/package.json` pins `vitest` `4.1.10`; its `test` script is `vitest run`.
`stats/web/vite.config.ts` sets `environment: "jsdom"`, `setupFiles` and `exclude` under `test`, and
no `pool`, `maxWorkers` or `poolOptions`, so the `forks` default applies. In vitest 4 the worker cap
is the top-level `test.maxWorkers` (`number | string`); `poolOptions.threads.maxThreads` is the
vitest 3 name and is not used.

Process isolation: `grep -rlnE "process\.(chdir|exit)|vitest-pool|poolOptions" stats/web/src
stats/web/tests` finds nothing, and `stats/web/package.json` carries no native addon. jsdom runs
per test file in either pool. Nothing in the suite needs `forks`.

Blast radius: `stats/web/vite.config.ts` and the new `stats/web/src/viteConfig.test.ts`. The fix
changes no name or section any other file cites (`grep -rl "vite.config" --exclude-dir=node_modules`
outside `spectre/changes/archive/` and `docs/` finds `.gitignore`, `stats/Makefile`,
`stats/cmd/flow/main.go` and `stats/web/package.json`, none of which reads the pool). Count: 2 files.
<!-- measured: grep -rl "vite.config" --exclude-dir=node_modules --exclude-dir=.git . | grep -v -e spectre/changes/archive -e ^./docs -->
Not `big`.

## Decisions

### Threads, two workers

**ID:** threads-two-workers
**Status:** active
**Chosen:** `test.pool: "threads"` and `test.maxWorkers: 2` in `stats/web/vite.config.ts`, with a
comment naming the orphaned-forks defect.
**Considered:** keeping `forks` with `maxWorkers: 2` — still orphans two processes on a kill.
`vmThreads` — different module isolation semantics, no need here.

### The check imports the config

**ID:** config-test
**Status:** active
**Chosen:** `stats/web/src/viteConfig.test.ts`, `// @vitest-environment node`, imports the default
export of `../vite.config` and asserts `config.test.pool === "threads"` and
`config.test.maxWorkers === 2`. It asserts the loaded value, not the file text.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-low — a two-line config change and one assertion test.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one fixer dispatch, Task 1, on `sonnet`: a literal config edit plus a test the plan
spells out.

## Open questions

None.

## Recommended approach

The two decisions above, one task, in `tasks.md`.

## Critical files

`stats/web/vite.config.ts`, `stats/web/src/viteConfig.test.ts`.

## Existing code to reuse

The existing `test` block of `stats/web/vite.config.ts`; no helper is needed.

## Verification

The Step 1 selector fails before the fix and passes after it; the full `## test` list, which runs
`stats/web`'s whole vitest suite, passes after the last task on the `threads` pool.
