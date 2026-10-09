# flow-scoped-reverify

## Why

Operator, 2026-10-09, after gymie KAN-924's in-run verify-fix loop: "re-verify only what a round touched (the Quick mode spec plus the four motion strips, about 3 minutes), and run the full suite once at the end. … The pipeline could also bound the loop: one-frame glitches the recording finds after the first fix round get batched into one task instead of opening a round each." and "create a subagent to always do that in flow".

Evidence from that run:

- Four `flow.visual-verify` rounds each ran the whole Playwright suite (~334 tests, ~25 min). `specs` was computed from the frontend's merge base, so a fix round touching one runner file still selected every spec.
  <!-- measured: cannot be re-run — quoted from the operator's dispatch, gymie KAN-924's verifier reports @ 2026-10-09 -->
- Each round's re-run found a smaller one-frame departure in a motion strip, opening another round.
- Two verifier reports were rejected by `check-verify-report.sh` on format alone — a `7 blocked` with no `— <reason> exit <n>`, and a `motion: 5/4` line counting a strip beyond the four named motions — and each needed a remainder dispatch.

## What changes

- `skills/flow/verify-fix-loop.md`, **The loop**: one round appends one task carrying every defect its report names. Step 4's `visual-verify` re-run is scoped to the round's own diff in every worktree a verifier of this run already verified: the prompt names a round base — the HEAD step 3 read for that worktree's previous verifier dispatch — which the verifier substitutes for `specs`' `<merge-base>`, and `motions:` names only the motions that diff adds, changes or removes plus every motion the round's defects name. A scoped re-run that comes back clean is followed by one full, unscoped `flow.visual-verify` on the same HEAD and stack, keyed `-fix-<k>-full`; a defect it finds opens the next round. **No cap** is unchanged.
- `skills/flow/visual-verify.md`: step 3 reads each worktree's HEAD in the same Bash call as its guards; the verifier prompt carries the round base when the loop scopes the re-run, and a REPORT FORMAT paragraph pointing the verifier at `check-verify-report.sh`'s header, naming the two forms that cost KAN-924 its remainders, and asking for every departure in one pass. The motion-naming paragraph cites the loop for the scoped subset.
- `skills/flow/visual-verify-verifier.md`: step 7 substitutes the round base when the prompt names one; a `specs` with no `<merge-base>` runs its full list and the report says `specs: not scoped`. The report template's `motion:` line counts only named motions, an extra strip going on an uncounted sub-bullet.
- `scripts/check-verify-report.sh` header: states the sub-bullet rule; two Go test cases pin it (`stats/internal/guard/check_verify_report_test.go`).
- `skills/flow/implement.md` closed dispatch list: the verifier row's key shape gains `-fix-<k>-full`.
- `skills/flow-contracts/project-configuration-visual.md`, `specs` row: a fix round's re-run substitutes its round base for `<merge-base>`.

No new dispatch role: the verifier subagent already exists, and every change is to what the parent dispatches it to do.
