## Context

The fix lives in gymie-playwright (worktree
`/Users/tweety53/Projects/gymie-playwright-worktrees/self-review-kan-944`, branch
`self-review-kan-944` off gymie-playwright `origin/main` @ `1f2b43f`). This change's artifacts live
in the agents worktree; no agents file changes.

Traced at `1f2b43f` (`git grep -n startScreencast origin/main`): four inline recorders —
`fidelity/kan-870-motion.spec.ts:77`, `fidelity/kan-940-cardio-rows.spec.ts:63`,
`fidelity/kan-983-first-frame.spec.ts:44` and `fidelity/nav-scoped-rights.spec.ts:238`. The first
three share one shape: lead wait, action, tail wait, frames written in order with their offset from
the action's start. They differ only in waits, file naming (`NNN_<ms>ms.png` in kan-870,
`<nnn>-<ms>.png` in the other two) and whether they assert reduced motion off and detach the CDP
session. `nav-scoped-rights.spec.ts` records for the whole action with no lead or tail, drops the
frames its own `page.screenshot` calls stream back, and writes flat `<name>-<nnn>.png` evidence
files.

`kan-940-cardio-rows.spec.ts`'s `borrowed-delete-${kind}` test never calls `page.emulateMedia`, so
under the `fidelity` project's `contextOptions.reducedMotion: 'reduce'` its strip records snapped
frames; `r2-hold-sweep` sets `no-preference` itself. Its earlier strips are an archived change's
historical evidence, recorded against a stack that is gone; this change fixes the recorder forward
and re-records nothing.

The agents-side visual-verify docs (`skills/flow/visual-verify-verifier.md`, the motion step)
describe the CDP screencast procedure generically, for every project, and name no scratch recorder.

Spiked at `1f2b43f`: the recorder below plus its spec, run as
`npx playwright test fidelity/screencast.spec.ts --project=fidelity` with no dev stack up, passed
in 1.1s — `page.setContent` needs no server.

## Live check

None — the change touches only Playwright test code; no running service or stored state changes.
The motion specs that call the recorder need the dev stack and run at the parent's verify.

## Decisions

### Recorder as one helper module beside the fidelity fixtures

**ID:** recorder-helper-module
**Status:** active
**Chosen:** `fidelity/screencast.ts` exporting `recordScreencast(page, dir, action, { leadMs = 250,
tailMs = 900 } = {}): Promise<number>` — asserts reduced motion off, starts the screencast
(acknowledging every frame), waits `leadMs`, runs `action`, waits `tailMs`, stops and detaches,
replaces `dir`, writes `<nnn>-<ms>.png`, asserts at least one frame and returns the frame count. The
spec that drives a motion is already TypeScript under Playwright, so a function it imports is the
whole need.
**Considered:** a standalone CLI script — rejected: a motion's trigger is spec code (a tag click, a
drag), which a script cannot take as an argument without becoming a spec runner; a module in the
repository root beside `fixtures.ts` — rejected: every caller is under `fidelity/`, and
`fidelity/frames.ts` is where that suite's shared helpers already live.

### File naming follows the verifier's strip

**ID:** strip-naming
**Status:** active
**Chosen:** `<nnn>-<ms>.png`, `<nnn>` three-digit zero-padded order, `<ms>` the frame's own
timestamp's offset from the action's start, rounded — the shape `visual-verify-verifier.md`'s
motion step writes and kan-940/kan-983 already use.
**Considered:** keeping kan-870's `NNN_<ms>ms.png` — rejected: one naming for every strip, and the
verifier's is the one the pipeline reads.

### Three specs move; nav-scoped-rights keeps its own

**ID:** migrated-specs
**Status:** active
**Chosen:** kan-870, kan-940 and kan-983 call `recordScreencast`. kan-983 starts its tag watch inside
the action it passes, and reads the watch once the recorder returns.
**Considered:** only the two specs the finding names — rejected: kan-983 is the same shape and would
stay a third copy; moving nav-scoped-rights too — rejected: it filters its own screenshots' frames
out and writes flat evidence files with no lead or tail, which the helper would need options for
that only one caller wants.

### kan-940 turns reduced motion off before recording

**ID:** kan940-motion-on
**Status:** active
**Chosen:** kan-940's `record` wrapper, when `VV_MOTION_DIR` is set, calls
`page.emulateMedia({ reducedMotion: 'no-preference' })` before `recordScreencast`; unset, it runs the
action unchanged as today. The recorder's own assertion would otherwise fail `borrowed-delete-*`.
**Considered:** the recorder emulating `no-preference` itself — rejected: a capture spec's
reduced-motion setting is its own, and the recorder silently changing it mid-test hides which
screenshot ran under which; dropping the recorder's assertion — rejected: a strip recorded under
reduced motion is the defect it exists to catch.

### No agents-side doc edit

**ID:** no-agents-doc-edit
**Status:** active
**Chosen:** no change to `skills/flow/visual-verify-verifier.md` or `skills/flow/visual-verify.md`;
gymie-playwright's `README.md` names the recorder in its `fidelity/` entry.
**Considered:** naming `fidelity/screencast.ts` in the verifier's motion step — rejected: that step
serves every project and describes no scratch recorder to replace; a gymie-playwright path in it
would be one project's detail in a shared contract.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-low — one new 25-line helper, its one spec, and three call-site swaps the plan
spells out; no guard, contract or concurrency.
**Considered:** flow-medium — rejected: nothing here needs the deeper read; the spec and the
`--list` compile check carry the risk.

### Task grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group — Task 1 alone, one fixer dispatch (`self-review-self-review-kan-944-fix`).
**Considered:** none — the plan has one task.

## Open questions
