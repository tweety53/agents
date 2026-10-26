# self-review-kan-1036-late-round-finding-prompt-tuning

## Why

kan-1036's self-review pass raised four findings, none `big`:

- **flow-fix — late-round miss excludes round 1.** `skills/flow-self-review/SKILL.md` defines a
  late-round miss as a finding first raised "in round 2 or later", but panel rounds count from 0
  (`stats/internal/records/types.go`, `Pass.Round`; `stats/cmd/flow/record.go`), so round 1 — the
  first re-run — is wrongly left out, and "a round-1 reviewer instruction" names the wrong round.
- **flow-fix — slot-facing reproducer name is vague.** The REPRODUCER paragraph every slot reads
  says "named by round, finding id and sequence number", contradicting the canonical
  `<round>-<id>-<n>.sh` in `skills/flow/review-panel.md`. Slots used slot-local ids
  (`r1-m1-1.sh`, `r1-M1-1.sh`), which collided on case-insensitive macOS: one slot's script
  overwrote the other's. No guard catches an off-shape name.
- **flow-stats-app — no Round column.** The rendered panel findings table carries no Round
  column although every finding row stores its round, so a late-round-miss trace cannot tell which
  round first raised a finding.
- **flow-cost — verifier rows read "not measured".** The inline `verify` row
  `run-verify-list.sh` records reads `Tokens: not measured`, intermittently, on recent ledgers.

## What changes

- The late-round-miss definition counts round 1 in.
- The slot-facing REPRODUCER paragraph states `<round>-<id>-<n>.sh`, and
  `check-panel-reproducers.sh` refuses a recorded path under `.superpowers/sdd/reproducers/` off
  that shape. kan-1036's own F3 row, whose path names a script the collision overwrote, is
  re-recorded truthfully.
- The panel findings table renders a `Round` column.
- The harvest measures an inline window that held no parent-session turn as zero, so a closed
  inline row no longer reads `not measured` once the harvest has read past its end.
