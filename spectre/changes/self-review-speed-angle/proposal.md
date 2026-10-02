# self-review-speed-angle

## Why

Operator, 2026-10-02: "add one more angle to self-review -> what can be sped up". The five angles cover tokens and effort (`flow-cost`) but nothing asks where a run's elapsed time went.

## What changes

- `skills/flow-self-review/SKILL.md`: angle 6, `flow-speed`, and the clause setting it apart from angle 2. Every "five angles" in the run-loaded corpus becomes "six angles".
- `scripts/check-self-review-report.sh`: an angle after the original five is demanded of a report only once that report carries at least one angle after the five, so the reports under `docs/self-review/` written before angle 6 stay valid. Harness cases 28 and 31 pin both directions.
