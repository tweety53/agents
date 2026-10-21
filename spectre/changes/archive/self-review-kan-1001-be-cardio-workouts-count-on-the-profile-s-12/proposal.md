# self-review-kan-1001-be-cardio-workouts-count-on-the-profile-s-12

## Why

The self-review pass of gymie change `kan-1001-be-cardio-workouts-count-on-the-profile-s-12` found
six fixable defects:

- `scripts/pipeline-fix.sh end self-review-<name>-pass completed` fails — 404 "no dispatch row … and
  no journalled begin" from the change's worktree, "no checkout carries change" from the agents
  repository — because `check-loop-dispatch` admits the pass off its pair and, by design, records no
  row for it, while step 3 of the self-review skill has the parent close each loop dispatch.
- `hooks/protect-main-checkout.py` keeps a `cat > $T <<'EOF'` data heredoc's body scanned when `$T`
  is merely named again as an argument (`check-plan-shape.sh $C/tasks.md`) or when any path-named
  command (`~/.claude/skills/flow/scripts/check-plan-shape.sh`, `./gradlew`) appears in the command,
  so prose such as `> **Execution:** …; …` in a plan write is denied as a redirect.
- The self-review bundle (`stats/internal/selfreview/git.go`, `archivedFiles`) serves the archived
  `tasks.md`, `design.md`, `narrative.md` and `visual-verification.md`, never `live-verification.md`,
  so a pass judging a live check never sees its before/after figures.
- gymie's `linkEnvFile` runs before `devStart` alone; `dbSeed` and `devWorkspaceCreate` seed through
  `seedDatabase`, which reads `DEV_INVITE_EMAIL` from `.env` — absent in a fresh worktree — so a
  worktree seeded before its first `devStart` carries no developer invite and no friend cast.
- gymie's before-figure recipe (`.flow/project.md`, `## run`) hand-copies ten `:app:bootRun`
  arguments that `seedDatabase` already assembles.
- The same recipe has no teardown for a workspace created in `../<name>-before`: kan-1001 left
  `kan-1001-be-9051`, and five more orphaned workspaces sit on this machine.

## What changes

- `pipeline-fix.sh end <…-pass> <outcome>` prints `PIPELINE-FIX-END-NA <key>: the pass is
  pair-checked, never recorded; nothing to close` and exits 0, calling no `flow`.
- A data heredoc's body stays scanned only when the rest of the command may run it: a runner word,
  a substitution, or a command word that may name the heredoc's target (same spelling, same file
  name, or a name an expansion hides). The target as a plain argument, and a path-named command
  naming another file, are data.
- The bundle serves archived `live-verification.md` after `visual-verification.md`.
- gymie's `dbSeed` and `devWorkspaceCreate` depend on `linkEnvFile`, pinned by
  `scripts/test-env-link-wiring.sh` in CI.
- gymie's `./gradlew -q devBootRunArgs -PbootPort=<port>` prints the before backend's `--args=`
  from the builder the seed run uses; the recipe points at it.
- The recipe removes the before worktree's workspace with `devWorkspaceRemove` before
  `git worktree remove`, and the six orphaned workspaces no live worktree owns are removed.
