# flow-scoped-reverify

## Why

Operator, 2026-10-09, after gymie KAN-924's in-run verify-fix loop: "re-verify only what a round touched (the Quick mode spec plus the four motion strips, about 3 minutes), and run the full suite once at the end. … The pipeline could also bound the loop: one-frame glitches the recording finds after the first fix round get batched into one task instead of opening a round each." and "create a subagent to always do that in flow".

Evidence from that run:

- Four `flow.visual-verify` rounds each ran the whole Playwright suite (~334 tests, ~25 min). `specs` was computed from the frontend's merge base, so a fix round touching one runner file still selected every spec.
  <!-- measured: cannot be re-run — quoted from the operator's dispatch, gymie KAN-924's verifier reports @ 2026-10-09 -->
- Each round's re-run found a smaller one-frame departure in a motion strip, opening another round.
- Two verifier reports were rejected by `check-verify-report.sh` on format alone — a `7 blocked` with no `— <reason> exit <n>`, and a `motion: 5/4` line counting a strip beyond the four named motions — and each needed a remainder dispatch.

## What changes

- `skills/flow/verify-fix-loop.md`, **The loop**: one round appends one task carrying every defect its report names. Step 4 re-runs both stages from their `begin` marks. Its `visual-verify` re-run is scoped to the round's own diff in every worktree with a round base — the HEAD step 12 read after that worktree's previous verifier dispatch: the prompt names that base, which the verifier substitutes for `specs`' `<merge-base>` in step 7, and `motions:` names only the motions that diff adds, changes or removes plus every motion the round's defects name. A worktree whose round diff is empty is not dispatched; one with no round base — never verified this run, or a sha the parent no longer holds — runs unscoped. A scoped re-run that comes back clean is followed by one full, unscoped `flow.visual-verify` from its `begin` mark, keyed `-fix-<k>-full`; a defect it finds opens the next round. **No cap** is unchanged.
- `skills/flow/visual-verify.md`: step 12 reads each worktree's HEAD last, after its commit; the verifier prompt carries the round base when the loop scopes the re-run, and a REPORT FORMAT paragraph pointing the verifier at `check-verify-report.sh`'s header and asking for every departure in one pass. The motion-naming paragraph cites the loop for the scoped subset. **A missed defect**: a scoped round passed only the motions it named and the specs its `verify` ran.
- `skills/flow/visual-verify-verifier.md`: step 7 substitutes the round base when the prompt names one, for `verify` alone — an empty output then runs no `verify` spec, and step 8's capture spec still comes from the merge base; a `specs` with no `<merge-base>` runs its full list and the report says `specs: not scoped`. The report template's `motion:` line counts only named motions, an extra strip going on a sub-bullet.
- `scripts/check-verify-report.sh` header: states the sub-bullet rule — documentation of what the guard already accepts, so no code change.
- `skills/flow/implement.md` closed dispatch list: the verifier row's key shape gains `-fix-<k>-full`.
- `skills/flow-contracts/project-configuration-visual.md`, `specs` row: a fix round's re-run substitutes its round base for `<merge-base>`.

No new dispatch role: the verifier subagent already exists, and every change is to what the parent dispatches it to do.
