# pipeline-fix-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets-43

## Why

Loop 41 forbade load reproductions outright. The operator refined the policy on 2026-10-28:
a load reproduction (backend or frontend) is allowed, but its memory must be strictly capped so the
machine is never overloaded; every process a flow run starts runs with strictly limited resources;
a flake that cannot be reproduced within the cap goes straight to a code fix on deeper analysis,
with no further reproduction loops. The policy must live in flow (rules, skills, scripts), not memory.

## What changes

- New `scripts/run-capped.sh`: runs a command in its own process group and kills the group when its
  total RSS exceeds a memory cap.
- New optional `.flow/project.md` key `## resource cap` (MB, default 4096).
- `run-verify-list` (the central `## test`/`## lint` executor) runs every command through the wrapper.
- `## run` dev-stack starts and any load reproduction go through the wrapper (contract wording).
- Loop 41's wording is rewritten: load repro allowed only under the cap with bounded workers;
  unreproducible within the cap leads straight to code with deeper analysis, the mechanism-removed
  run waived and stated. The `proofruns`/`noloadrepro` guard phrases follow.
