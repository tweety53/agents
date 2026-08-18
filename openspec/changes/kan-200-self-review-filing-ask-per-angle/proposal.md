# Self-review's filing ask covers every angle, and the report becomes checkable

## Why

`/myflow-finish` run 2 step 8 already requires one combined pass over all four self-review angles
and "a per-finding Jira filing ask". In practice "per-finding" has read as "per **problem**".

KAN-73's run is the worked example: the problems angle produced five findings, all offered and all
filed (KAN-192 … KAN-196); the **what went well** angle produced three items and no filing ask at
all; the **cost** angle produced a description of spend and nothing actionable; the **automation**
angle listed only cross-references back to the problems angle's tickets. Two of four angles were
write-only and a third was a cross-reference table.

Both angles that produced nothing produced something real the moment they were actually worked —
KAN-201 from cost (including a ~30x difference in findings-per-token between panel slots) and
KAN-202 from automation (a live defect: `/myflow-fast` bypasses `commit-split.sh` entirely). The
omission was in the execution, not in the material.

The rule that was skipped **was already written down**. That is the fact this change is built
around: a fifth angle and a per-angle filing ask stated only in prose would be skipped the same
way. So the report itself becomes a checkable artifact, and a guard checks it.

## What Changes

- **A fifth angle.** What could move to the Go app or its persistent storage — records the pipeline
  writes to files today, and derivation work now done in Bash or by the agent. Label
  `myflow-stats-app`.
- **The filing ask is per angle, not per problem.** Every angle produces zero or more findings; the
  ask covers every angle's findings; an angle that yields nothing says so explicitly rather than
  falling silent.
- **Explain before filing.** Every finding is explained in the message body — what was observed,
  what breaks, what the fix would be — before any prompt fires. The prompt records the decision and
  nothing else, as one multi-select per angle. This binds run 1's follow-up filing too.
- **Each filed issue carries its angle's label**, on top of the inherited set.
- **The report gains a fixed shape**: five sections always present, one parseable line per finding
  carrying its angle label, disposition and issue key.
- **`scripts/check-self-review-report.sh`** checks that shape over every report, with the sixteen
  pre-rule reports declared by name, per KAN-197's coverage pattern.
- **Carried alongside:** KAN-197's self-review report, stranded on a merged-and-deleted branch.

## Impact

- Affected specs: `myflow-self-review` (two requirements modified, two added)
- Affected code: `skills/myflow-contracts/finish-contract.md`,
  `skills/myflow-contracts/jira-integration.md`, `skills/myflow-contracts/jira-followups.md`,
  `skills/myflow-finish/SKILL.md`, `scripts/check-self-review-report.sh` (new),
  `scripts/test-check-self-review-report.sh` (new), `scripts/check-contract-budget.sh`,
  `.myflow/project.md`
