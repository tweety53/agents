# Self-review

## After a change lands

Every change ends with one review of the run that built it, not of the product. A single agent
reads a bundle of what happened and reports under five angles, each always present even when
empty:

| Angle | Asks |
|-------|------|
| Problems | What went wrong, and what change to the pipeline would avoid it? |
| What went well | What worked, and how is it reproduced? |
| Instruction gaps | Which missing instruction cost a review or visual-verify fix round, and what prompt update would have avoided it? |
| Cost | What spent tokens or time working or waiting, and could be cut without losing quality? |
| Automation | Which guard, hook, script or other automation would have avoided a problem or a manual step? |

What happens to a finding:

- **A problem with a small fix is fixed and landed without asking.**
- **A big one**, a fix touching 60 or more files, is not fixed in the pass; it is suggested.
- **The other four angles are suggested, never applied**: you approve or decline each on the
  dashboard's Self-review view. Only Critical and Important findings are suggested.
- **A new or extended guard, check or hook is never built by the pass**; it is suggested.

Canonical: [`skills/flow-self-review/SKILL.md`](../skills/flow-self-review/SKILL.md).

## Fixes made mid-run

A run that hits a defect in the pipeline itself, in a skill, a guard, the CLI or the service, does
not stop to ask. It fixes the defect within the run and continues; only a fix touching 60 or more
files is deferred to self-review.

Canonical: **Pipeline defects found mid-run** in
[`skills/flow-contracts/pipeline.md`](../skills/flow-contracts/pipeline.md).
