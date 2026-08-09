# Self-review — kan-100-myflow-get-rid-of-staging-use-commits

## Problems, and the pipeline change that would avoid them

**The new worktree didn't inherit `/myflow-start`'s untracked planning artifacts.** `openspec/new
change` and the proposal/design/specs/tasks files `/myflow-start` writes are staged-only in the
main checkout — never committed, per this pipeline's own git boundaries. `git worktree add` only
carries committed history, so the freshly created worktree had no `openspec/changes/<name>/` at
all; `sdd-workspace` failed with "no such plan file" until the files were copied over by hand.
**Fix:** `/myflow-do` section 2 (isolate the workspace) should copy the untracked planning
artifacts from the main checkout into the new worktree as a normal step, not leave it as something
a run has to discover.

**A concurrent, unrelated change's `/myflow-finish` run 2 swept up two of this change's leftover
untracked files into its own archive commit on `main`.** `kan-111-myflow-fast`'s archive-and-sync
commit added `openspec/changes/kan-100-.../tasks.md` (stale, unchecked) and
`docs/superpowers/specs/2026-08-09-kan-100-...-design.md` to `main` — neither belongs to
`kan-111`. Both were caused by a `git add -A`-shaped step in run 2's archive commit that isn't
scoped narrowly enough, so it picked up sibling untracked files sitting in the working tree from
another change's still-open `/myflow-start`. This produced a real add/add merge conflict when this
change's own branch merged. **Fix:** `/myflow-finish` run 2's archive-commit step should stage only
the specific paths it touches (`openspec/specs/`, the archive destination directory,
`docs/self-review/` when relevant) instead of a broad add, so one change's finish run can never
commit another change's in-flight leftovers.

**`tasks.md`'s checkbox format requirement isn't stated anywhere obvious.** The first draft of this
plan used one `### N.N` heading + prose + a `**Verify:**` line per task, with no `- [ ]` checkbox
lines — matching `build-green.md`'s placement rule for tags, but not `openspec`'s actual parser,
which tracks completion via checkboxes and reported "blocked: contains no tasks" until the whole
file was restructured. **Fix (optional, lower priority):** either state the checkbox requirement
explicitly in writing-plans' own guidance, or add a guard at `/myflow-start` publish time that
fails fast when a `tasks.md` has zero `- [ ]` lines, so this is caught before `/myflow-do` ever
starts rather than as a first-thing "blocked" surprise.

## Cost

This was one of the largest `/myflow-do` runs by dispatch count: 61 checkbox items, ~18 real
implementer+reviewer dispatch pairs (several merged red/green task groups), 6 fix rounds, 4
addenda for gaps reviewers surfaced mid-run, and a 3-slot final panel plus one fix wave. Several of
the smallest dispatch units (2.5, 5.1, 5.2, 6.1, 6.3) were single-paragraph or single-line prose
edits that still cost a full implementer-dispatch + reviewer-dispatch round trip. For a
documentation/contract-editing change like this one, batching several small, genuinely independent
prose edits into fewer combined dispatches — the way the merged `Build: red` pairs already forced
some grouping — would meaningfully cut total dispatch count without losing review coverage, since
the reviewer for a 3-line change and the reviewer for a 500-line change cost roughly the same fixed
dispatch overhead regardless of diff size.

## What went well

The merged `Build: red`/`Squash-with:` task-pairing mechanism worked cleanly throughout — every
pairing landed as one reviewed unit with no partial-state gaps. Reviewers repeatedly caught real,
concrete gaps outside their assigned task's literal scope (the stale `CLAUDE.md`/`SKILL.md`
summaries, the self-referential `Squash-with:` migration this plan's own `tasks.md` needed once its
own guard changed) and those were resolved immediately as small addenda rather than deferred or
ignored — the addendum pattern (register a task, dispatch, review, ledger) kept the main sequence
moving without losing the finding. The merge-conflict handling at finish (the swept-up files,
the genuine `check-contract-budget.sh` content conflict) was investigated and confirmed
byte-identical or resolved as a content union before touching anything, and the anomaly was
surfaced to the operator rather than silently resolved.

## What could be automated

- Fold the untracked-planning-artifact copy into `/myflow-do` section 2 or
  `superpowers:using-git-worktrees`, per the first problem above.
- A small `ledger-append` helper (or a documented one-liner) instead of ad hoc `printf >>` calls for
  every SDD ledger line, to keep the format consistent without relying on getting the shell quoting
  right every time.

## Rating

**3/5** (operator's rating). The three findings above — filed as KAN-118, KAN-119, KAN-120 — are
the concrete gaps behind that score: the worktree not inheriting planning artifacts, and
`/myflow-finish` run 2's overly broad archive-commit staging both cost real mid-run time to diagnose
and work around rather than being handled by the pipeline itself, and the tasks.md checkbox
requirement gap forced a full plan restructure before implementation could even start.
