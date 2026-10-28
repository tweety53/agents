# pipeline-fix-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets-42

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** stop `stats/web`'s vitest from leaving orphaned fork workers holding gigabytes after its
parent is killed.

**Architecture:** switch the vitest pool to `threads`, capped at two workers, in
`stats/web/vite.config.ts`, with one vitest test asserting the loaded config.

**Spec:** `design.md` beside this file.

## Global constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/pipeline-fix-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets-42`, branch `pipeline-fix-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets-42`; `**Files:**` are relative to its root. Never touch the main checkout; never write `spectre/changes/`.
- One commit per task with its `**Commit:**` subject, no `Task-Id:` trailer. Never merge or push.
- vitest is `4.1.10`: the option is `test.maxWorkers`, never `poolOptions`.
- Run vitest with a Bash timeout of 600000. Never start, stop or restart the protected `flowd`.

## Review Focus

- Task 1: the test fails with `pool` back on its default or `maxWorkers` removed.

- [ ] 1. Run stats/web's vitest on two worker threads
  - [ ] **Step 1: Write the failing test** `stats/web/src/viteConfig.test.ts`:

```ts verified:vite.config.ts default-exports defineConfig({...}) with a test block at plan time
// @vitest-environment node
import { describe, expect, it } from "vitest";
import config from "../vite.config";

// A killed parent vitest orphans `forks` workers to PID 1 (~180 MB each);
// threads die with the parent.
describe("vite.config test pool", () => {
  it("runs on two worker threads", () => {
    expect(config.test?.pool).toBe("threads");
    expect(config.test?.maxWorkers).toBe(2);
  });
});
```
    Run `cd stats/web && npx vitest run -t "runs on two worker threads"`; it fails.
  - [ ] **Step 2: Fix** `stats/web/vite.config.ts`: in the `test` block, before `environment`, add

```ts verified:test block keys environment, setupFiles, exclude read at plan time
    // The default `forks` pool spawns one process per CPU; a killed parent
    // (Bash timeout, stopped agent) orphans them to PID 1, ~2.5 GB held.
    // Threads die with the parent; two cap the footprint.
    pool: "threads",
    maxWorkers: 2,
```
    If `tsc` rejects the test's `config.test` access on the `UserConfig` type, add
    `/// <reference types="vitest/config" />` as the config's first line; nothing else.
  - [ ] **Step 3: Verify.** `cd stats/web && npx tsc --noEmit -p .` clean; the Step 1 command passes.
  - [ ] **Step 4: Commit** the two files.

**Build:** green
**Files:** `stats/web/vite.config.ts`, `stats/web/src/viteConfig.test.ts`
**Tests:** `vite.config test pool > runs on two worker threads`
**Regression:** `runs on two worker threads` — reverting the pool or the cap fails it.
**Baseline:** before=0 after=1
<!-- measured: grep -cE '^\s*it\(' stats/web/src/viteConfig.test.ts @ branch pipeline-fix-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets-42 -->
**Commit:** fix(stats): run the web vitest suite on two worker threads
**After:** none

**Decision:** threads-two-workers
