# self-review-kan-1000-speedups

## Why

The KAN-1000 gymie run (2026-10-20) spent most of its wall-clock in the visual-verify fix loop,
for six reasons the operator asked to fix together:

1. The loop's implementer judged its mockup fixes by eye, so each verifier round found a new
   small offset the previous fix had left.
2. Verifiers skipped step 10 (`10 not done`, or `10 blocked` on another step's exit with every
   capture on disk), and the parent planned fix tasks from those partial reports, so departures
   surfaced one round at a time.
3. Scoped rounds re-ran the whole regression suite, and three rounds in a row ran unscoped
   `-full` re-runs.
4. A feature-branch base that had fallen behind develop produced a round of failures before the
   lag was found.
5. Sub-dp position offsets a font's glyph metrics cause had no sanctioned end state, so the loop
   kept reopening them.
6. The parent forked itself to run a self-review loop, carrying its whole context into a dispatch
   that needed a change list and file paths, and bypassing the rule that the parent drives every
   keyed dispatch.

## What changes

- The loop's implementer dispatch for a mockup departure carries the frame-measurement
  instruction, and the implementer reports its measured numbers (`skills/flow/verify-fix-loop.md`).
- `check-verify-report.sh` refuses a `10 blocked` unless step 8 is blocked too or the cause is
  `compose-mockup-frames`' own exit; an incomplete report is never planned from
  (`skills/flow/visual-verify.md`).
- Under a round base the verifier's `verify` never runs the whole suite, and the round a `-full`
  run's defect opens re-runs scoped (`skills/flow/visual-verify-verifier.md`,
  `skills/flow/verify-fix-loop.md`).
- The parent checks a feature-branch base against the branch it tracks before the first verify
  stage (`skills/flow-contracts/operator-prompts-auto-resolution.md`).
- A position departure of at most 1dp may stand after a fix round tried it, recorded and listed
  (`skills/flow/visual-verify.md`, `skills/flow/visual-verify-verifier.md`).
- No flow phase or loop dispatches a fork; `hooks/enforce-agent-baseline.py` denies one in a flow
  session (`skills/flow/implement.md`, `skills/flow-self-review/SKILL.md`).
