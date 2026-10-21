# self-review-kan-944

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close KAN-944 — one checked-in CDP frame recorder in gymie-playwright that the fidelity
specs, the visual verifier and fix runs record motion strips through.

**Architecture:** one helper module, `fidelity/screencast.ts`, beside `fidelity/frames.ts`; three
specs swap their inline recorders for it; one stack-free spec covers it.

**Global Constraints:**

- Task 1 runs in gymie-playwright's worktree
  `/Users/tweety53/Projects/gymie-playwright-worktrees/self-review-kan-944`, branch
  `self-review-kan-944` (exists; `npm ci` done). Every `**Files:**` path is relative to that
  worktree's root. Never touch a main checkout.
- gymie-playwright declares no lint command; `npx playwright test --list` is the compile check.
- The motion specs need the dev stack; no step starts it. Only `fidelity/screencast.spec.ts` runs.

- [x] 1. Record motion strips through one shared CDP recorder (KAN-944)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `git grep -n -e 'Page.startScreencast' -- fidelity/` and `ls fidelity/screencast.ts`: four hits
    (kan-870, kan-940, kan-983, nav-scoped-rights) and no such file.
    <!-- measured: git grep -n startScreencast origin/main @ gymie-playwright 1f2b43f → 4 hits; fidelity/ holds cardio.ts, fixtures.ts, frames.ts, offline-l5.ts and no screencast.ts -->
  - [x] **Step 2: Write the failing spec** `fidelity/screencast.spec.ts`:

```ts verified:ran npx playwright test fidelity/screencast.spec.ts --project=fidelity with the Step 3 module, no stack, gymie-playwright 1f2b43f, 1 passed
/**
 * `recordScreencast` (./screencast.ts) on a page with no app behind it: a CSS transition, a
 * stale file in the strip directory, and the strip read back.
 */
import fs from 'fs';
import { expect, test } from '@playwright/test';
import { recordScreencast } from './screencast';

test.use({ contextOptions: { reducedMotion: 'no-preference' } });

test('recordScreencast writes an ordered nnn-ms strip and replaces the old one', async ({ page }, info) => {
  await page.setContent('<div id="b" style="width:40px;height:40px;background:red;transition:transform 400ms linear"></div>');
  const dir = info.outputPath('strip');
  fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(`${dir}/stale.png`, '');
  const n = await recordScreencast(page, dir, async () => {
    await page.evaluate(() => {
      document.getElementById('b')!.style.transform = 'translateX(200px)';
    });
  }, { tailMs: 600 });
  const files = fs.readdirSync(dir);
  expect(files).toHaveLength(n);
  expect(n).toBeGreaterThan(1);
  files.forEach((f, i) => expect(f).toMatch(new RegExp(`^${String(i).padStart(3, '0')}-(-?\\d+)\\.png$`)));
});
```
    <!-- measured: this spec with the Step 3 module, npx playwright test fidelity/screencast.spec.ts --project=fidelity, no dev stack @ gymie-playwright 1f2b43f → 1 passed (1.1s) -->

    Run `npx playwright test fidelity/screencast.spec.ts --project=fidelity`: it fails, `./screencast`
    not found.
  - [x] **Step 3: Add `fidelity/screencast.ts`** (decisions `recorder-helper-module`, `strip-naming`):

```ts verified:same spike as Step 2, gymie-playwright 1f2b43f, 1 passed
import fs from 'fs';
import path from 'path';
import { expect, type Page } from '@playwright/test';

/**
 * Records every composited frame by CDP screencast (`Page.startScreencast`, each frame acknowledged)
 * for [leadMs] before [action], while it runs, and for [tailMs] after it, then replaces [dir] with
 * the strip `<nnn>-<ms>.png`, `<ms>` the frame's own timestamp's offset from the action's start —
 * the shape flow's visual verifier reads. Reduced motion is checked off first, so a run that
 * silently snapped is a failure, not a strip of rest frames. Returns the frame count.
 *
 * gymie's CLAUDE.md verifies a swap or a load by recorded frames: every motion strip — a fidelity
 * spec's, the visual verifier's, a fix run's — is recorded through this, never a recorder of its own.
 */
export async function recordScreencast(
  page: Page,
  dir: string,
  action: () => Promise<void>,
  { leadMs = 250, tailMs = 900 }: { leadMs?: number; tailMs?: number } = {},
): Promise<number> {
  expect(await page.evaluate(() => matchMedia('(prefers-reduced-motion: reduce)').matches), 'reduced motion must be off').toBe(false);
  const cdp = await page.context().newCDPSession(page);
  const frames: { data: string; t: number }[] = [];
  cdp.on('Page.screencastFrame', (f) => {
    frames.push({ data: f.data, t: f.metadata.timestamp ?? Date.now() / 1000 });
    cdp.send('Page.screencastFrameAck', { sessionId: f.sessionId }).catch(() => {});
  });
  await cdp.send('Page.startScreencast', { format: 'png', everyNthFrame: 1 });
  await page.waitForTimeout(leadMs);
  const t0 = Date.now() / 1000;
  await action();
  await page.waitForTimeout(tailMs);
  await cdp.send('Page.stopScreencast');
  await cdp.detach();
  fs.rmSync(dir, { recursive: true, force: true });
  fs.mkdirSync(dir, { recursive: true });
  frames.forEach((f, i) =>
    fs.writeFileSync(path.join(dir, `${String(i).padStart(3, '0')}-${Math.round((f.t - t0) * 1000)}.png`), Buffer.from(f.data, 'base64')),
  );
  expect(frames.length, `${path.basename(dir)}: the screencast delivered frames`).toBeGreaterThan(0);
  return frames.length;
}
```

    Run Step 2's command again: it passes.
  - [x] **Step 4: Move the three specs onto it** (decisions `migrated-specs`, `kan940-motion-on`).
    - `fidelity/kan-870-motion.spec.ts`: delete its `record` function and doc comment; keep `OUT`;
      add `import { recordScreencast } from './screencast';` and define
      `const record = (page: Page, name: string, action: () => Promise<void>, tailMs = 900) => recordScreencast(page, path.join(OUT, name), action, { tailMs }).then(() => {});`
      so every call site stays unchanged. Its header's frame path sentence names
      `<nnn>-<ms>.png`.
    - `fidelity/kan-940-cardio-rows.spec.ts`: its `record` keeps its doc comment and becomes

```ts unverified:run r2-hold-sweep and borrowed-delete with VV_MOTION_DIR set against the dev stack
async function record(page: Page, id: string, act: () => Promise<void>): Promise<void> {
  const root = process.env.VV_MOTION_DIR;
  if (!root) return act();
  // The fidelity project reduces motion; a strip recorded under it shows the motion snapped.
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  await recordScreencast(page, path.join(root, id), act, { leadMs: 400, tailMs: 1200 });
}
```

      importing `path` and `recordScreencast`; drop the `fs` import if nothing else uses it.
    - `fidelity/kan-983-first-frame.spec.ts`: its `record` becomes

```ts unverified:npx playwright test --list compiles it; its first-frame tests need the dev stack
async function record(page: Page, id: string, action: () => Promise<void>, tailMs = 1200): Promise<string[][]> {
  let stop!: () => Promise<string[][]>;
  const dir = path.join(OUT, id);
  await recordScreencast(page, dir, async () => {
    stop = await watchTags(page);
    await action();
  }, { tailMs });
  const seen = await stop();
  fs.writeFileSync(path.join(dir, 'tags.json'), JSON.stringify(seen));
  return seen;
}
```

      importing `recordScreencast`.
    - `README.md`: in the `fidelity/` bullet, add one sentence: "`fidelity/screencast.ts`'s
      `recordScreencast` records a motion strip — `<nnn>-<ms>.png` by CDP screencast, reduced motion
      off — and every strip, a spec's, the visual verifier's or a fix run's, is recorded through it."
  - [x] **Step 5: Verify.** `git grep -n -e 'Page.startScreencast' -- fidelity/` prints exactly two
    lines, `fidelity/screencast.ts` and `fidelity/nav-scoped-rights.spec.ts`;
    `npx playwright test --list --project=fidelity fidelity/kan-870-motion.spec.ts fidelity/kan-940-cardio-rows.spec.ts fidelity/kan-983-first-frame.spec.ts fidelity/screencast.spec.ts`
    prints `Total: 24 tests in 4 files`; `npx playwright test fidelity/screencast.spec.ts --project=fidelity`
    passes.
    <!-- measured: npx playwright test --list --project=fidelity on the three specs @ gymie-playwright 1f2b43f → Total: 23 tests in 3 files -->
  - [x] **Step 6: Commit** the six files with the subject below.

**Build:** green
**Files:** `fidelity/screencast.ts`, `fidelity/screencast.spec.ts`, `fidelity/kan-870-motion.spec.ts`, `fidelity/kan-940-cardio-rows.spec.ts`, `fidelity/kan-983-first-frame.spec.ts`, `README.md`
**Tests:** `recordScreencast writes an ordered nnn-ms strip and replaces the old one`
**Regression:** reverting the commit removes `fidelity/screencast.ts`; the spec's import fails and the three specs return to their own inline recorders.
**Baseline:** before=0 after=1 <!-- measured: cat fidelity/screencast.spec.ts 2>/dev/null | grep -c "^test('" @ gymie-playwright 1f2b43f → 0 -->
**Commit:** test(fidelity): record motion strips through one shared CDP recorder
**After:** none

**Decision:** recorder-helper-module
**Decision:** strip-naming
**Decision:** migrated-specs
**Decision:** kan940-motion-on
