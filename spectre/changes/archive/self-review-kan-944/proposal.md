# self-review-kan-944

## Why

KAN-944 (self-review angle `flow-automation`): gymie's `CLAUDE.md` requires a swap or a load to be
verified by recorded frames — a CDP screencast — yet gymie-playwright carries no shared recorder.
Motion strips were recorded twice by a scratch recorder kept outside the checkout, and the fidelity
specs each inline their own `Page.startScreencast` loop: `fidelity/kan-870-motion.spec.ts`,
`fidelity/kan-940-cardio-rows.spec.ts` and `fidelity/kan-983-first-frame.spec.ts` carry three
near-identical copies, each naming its frames differently. One of them,
`kan-940-cardio-rows.spec.ts`'s `borrowed-delete-*`, records under the `fidelity` project's
`reducedMotion: 'reduce'`, so its strip shows the motion snapped rather than played.

## What changes

- gymie-playwright gains `fidelity/screencast.ts`, one `recordScreencast(page, dir, action,
  { leadMs, tailMs })` that a spec — the visual verifier's, a fix run's — calls to record a motion
  strip into `<dir>/<nnn>-<ms>.png`, `<ms>` the frame's offset from the action's start.
- `fidelity/kan-870-motion.spec.ts`, `fidelity/kan-940-cardio-rows.spec.ts` and
  `fidelity/kan-983-first-frame.spec.ts` record through it; their inline recorders go.
- `kan-940-cardio-rows.spec.ts` turns reduced motion off before it records, so `borrowed-delete-*`
  plays its motion.
- `fidelity/screencast.spec.ts` covers the recorder with no running stack.
- gymie-playwright's `README.md` names the module.
