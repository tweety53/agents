# self-review-speed-angle

## Why

Operator, 2026-10-02: "add one more angle to self-review -> what can be sped up". The five angles cover tokens and effort (`flow-cost`) but nothing asks where a run's elapsed time went.

## What changes

- `skills/flow-self-review/SKILL.md`: angle 6, `flow-speed`, and the clause setting it apart from angle 2. Every "five angles" in the run-loaded corpus becomes "six angles".
- `scripts/check-self-review-report.sh`: the reports under `docs/self-review/` written before angle 6 are named in a frozen `docs/self-review/five-angle-reports.txt` and checked against the first five angles; every other report carries all six.
- `skills/flow/withdrawal.md`, `skills/flow/brainstorm-planner.md`: an issue labelled `flow-speed` gets the reachability check `flow-fix` and `flow-cost` issues get — a speed finding names a pipeline change exactly like a cost one.
