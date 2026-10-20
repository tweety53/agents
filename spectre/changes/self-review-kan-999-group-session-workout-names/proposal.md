# self-review-kan-999-group-session-workout-names

## Why

The self-review pass of `kan-999-group-session-workout-names` found four fixable defects:

- `spec_root_leaf` (`scripts/lib/spec-root.sh`) and its Go port `specRootLeaf`
  (`stats/internal/guard/specroot.go`) print the "carries both spectre/changes/ and
  openspec/changes/ … invisible to this guard" note whenever `openspec/changes/` exists. gymie froze
  its `openspec/` tree as history — `openspec/changes/` holds only `archive/` — so every guard call
  there prints a warning about changes that do not exist.
- kan-999's task 1 hit detekt `TooManyFunctions` on three types already at the threshold: the plan
  added members to them under a "no new port" constraint that the lint then broke. Section D of
  `skills/flow/brainstorm-planner.md` never tells a planner to count a grown type's members against
  the project's size rules.
- `RenderLedger` (`stats/internal/records/render.go`) never renders `Dispatch.EndedAt`, so every
  rendered ledger shows when a dispatch started and never when it ended or how long it took.
- kan-999's live check hand-seeded rows through `psql` — three constraint snags and a silently
  rolled-back `psql -c` transaction — and its cleanup `delete from group_session` wiped `dbSeed`'s
  ten sessions; its before-figure stack was hand-built. gymie's `.flow/project.md` says nothing about
  either.

## What changes

- The both-trees note fires only while `openspec/changes/` holds an entry other than `archive/`;
  the leaf returned is unchanged.
- Section D of `skills/flow/brainstorm-planner.md` carries one rule: a task adding a member to an
  existing class or interface counts that type's members against the size-rule thresholds and plans
  the split when there is no headroom.
- A ledger renders `- Ended: <RFC3339> (<duration>)` after `- Started:` for every dispatch whose end
  is recorded.
- gymie's `.flow/project.md` gains two paragraphs beside the `count-request-queries.sh` one: live
  check fixtures from `dbSeed`'s rows, created through the API, cleaned up by id, `dbSeed` after any
  wider delete; and the exact before-figure backend recipe.
