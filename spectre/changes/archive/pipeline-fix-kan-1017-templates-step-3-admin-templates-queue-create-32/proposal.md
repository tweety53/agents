# pipeline-fix-kan-1017-templates-step-3-admin-templates-queue-create-32

## Why

- kan-1017's `/flow` run stopped after Task 11's commit — gated review, panel re-review, verify re-run and integrate still ahead — to hand the operator a status report citing run size and listing open items that each had a recommended fix. **Fewest operator actions** (`skills/flow-contracts/pipeline.md`) named none of run size, cost, context length, remaining work or a checkpoint as a non-reason to stop, and did not forbid offering a shared-source departure as "a separate change".
- `scripts/check-installed-citations.sh` fails on `origin/main`: `skills/flow/cleanup.md:67` cites `<project>-learning/<name>.md`, which names no root. <!-- citations-guard:allow -->

## What changes

- `skills/flow-contracts/pipeline.md`: a new paragraph **A run's size is never a stop.** in **Fewest operator actions**.
- `skills/flow/cleanup.md`: the learning-note path line carries `<!-- citations-guard:allow -->`, marking it an illustration.
