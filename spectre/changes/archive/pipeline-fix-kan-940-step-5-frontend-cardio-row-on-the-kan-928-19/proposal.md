# pipeline-fix-kan-940-step-5-frontend-cardio-row-on-the-kan-928-19

## Why

The operator ruled on 2026-10-17 — "update flow not to ask such questions again" — on two asks
kan-940's run made that have one answer every time:

- After a sync's targeted recapture, tests in specs the change never touched failed on the
  base's own baselines ("… — week 1" titles the running backend no longer draws), and the run asked
  whether to re-baseline them. The answer is always yes.
- Run 2's `integrate-change.sh cleanup` stopped with `STOP: disclose` on residue nothing is lost
  by removing: `.DS_Store` files, files already preserved as identical in the main checkout, a
  `.env` byte-identical to the main checkout's, and screenshot-spec output PNGs.

## What changes

- A screenshot mismatch in a spec the change never touched is re-baselined in the run, without
  asking: only the failing tests are recaptured, only the PNGs they named are committed, as their
  own `test(fidelity): re-baseline …` commit on the change branch. A failure that is no screenshot
  mismatch is reported, never re-baselined. Canonical in **the sweep**
  (`skills/flow-contracts/known-bugs.md`), cited from `skills/flow/sync-onto-base.md` and
  `skills/flow/visual-verify-verifier.md`.
- `remove-change-worktrees`'s check 4 counts a `.DS_Store` anywhere, and an ignored `.png` at the
  root of a Playwright checkout, as regeneratable: neither stops the cleanup. Entries identical in
  the main checkout already never stopped it; only genuinely unique, unclassified work does.
