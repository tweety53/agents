# self-review-kan-982-prod-fixes-group-add-person-tap-set-value-bleed-2

## Why

Two kan-982 self-review findings the operator chose to fix now, both defects only a review panel
caught, with nothing mechanical behind it:

- A commit sitting outside every task — no `Task-Id:` trailer, not a planning commit — reaches
  run 1, whose reshape squashes it into the implementation commit and erases the evidence. kan-982's
  panel raised it as F3 ("Two non-chore commits sit outside every task"); the guard the finding
  proposed was dropped once because the pipeline itself makes such commits on sanctioned routes.
- A declared `**Commit:**` scope that names a module of another repository passes
  `check-task-commit-fields.sh`. kan-982's Task 21 landed `fix(group): …` on a gymie-playwright
  commit touching only `fidelity/builder.spec.ts`; `group` is a gymie module (panel F16). The
  check the finding proposed was dropped once because matching scopes against touched paths
  refuses legitimate area scopes such as `test(visual)`.

## What changes

- A new Go guard, `check-untasked-commits.sh <worktree> <base>`, lists every non-merge commit of
  `<base>..HEAD` that carries neither a `Task-Id:` nor a `Flow-Key:` trailer, has no `chore`
  subject, and touches a path outside `<spec-root>/changes/`. Run 1's unfinished-work gate runs it
  per worktree (`scripts/integrate-change.sh` and the finish contract) and treats a listing exactly
  as `OUTSTANDING`.
- Every sanctioned route that commits outside a task — the parent's full-suite fix, a panel or
  gated reviewer's fix committed on top, `flow.visual-verify`'s capture commit and the brainstorm's
  handoff-asset commit — writes a `Flow-Key: <key>` trailer, canonical in a new **Commits outside
  every task** section of `skills/flow-contracts/git-boundaries.md`.
- `check-task-commit-fields.sh` refuses a declared scope that is a path token of another `## apps`
  repository and none of: a token of a path the commit touched, a word of the commit's repository
  name, a scope an earlier commit of that repository used, or a directory name of that repository
  (every `-` part one, for a compound area). The commit-scope rule names the fourth prohibited
  shape.
