# kan-1036-late-round-finding-prompt-tuning

## Why

- A finding a review slot first raises on round 3, on code that same slot already reviewed unchanged in rounds 1 and 2, means the earlier rounds missed it — and the self-review's prompt-tuning angle only looks at such a miss when it passes the severity bar and its cause happens to be read as an instruction gap.
- The earlier rounds' miss is the signal worth investigating; today nothing requires the pass to trace why it happened (KAN-1036).

## What changes

- Angle 8 of the self-review pass (`skills/flow-self-review/SKILL.md`) names a **late-round miss** and requires it traced, whatever its severity, by comparing that slot's earlier dispatches with the round that raised it.
- A generic instruction gap becomes a `reviewer` miss case naming the earlier rounds; any other cause is an angle-8 finding classified by step 3's classification.
