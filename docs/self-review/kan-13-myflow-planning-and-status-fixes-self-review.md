# Self-review — kan-13-myflow-planning-and-status-fixes

**Rating: 4/5**

## Problems, and the pipeline change that would avoid them

1. **Creating the worktree didn't bring the `STARTED`-state planning artifacts with it.**
   `/myflow-start` leaves `proposal.md`/`design.md`/`tasks.md`/`specs/` **staged but uncommitted**
   in the main checkout. `git worktree add` branches from `HEAD` — the last **commit** — so the new
   worktree had no `openspec/changes/<name>/` directory at all. This wasn't caught until the first
   `uncommitted-review-package` call failed with "no such plan file." The fix was ad-hoc: copy the
   files into the worktree by hand, then unstage-and-delete them from the main checkout to avoid a
   duplicate, stale copy sitting there for the rest of the run. **Pipeline change:** `/myflow-do`'s
   "Isolate the workspace" step should explicitly say to move (not just create) the staged planning
   artifacts from the main checkout into the new worktree as part of worktree setup, before any
   implementer is dispatched — this is exactly the shape of gap this self-review mechanism exists to
   surface.
2. **The archive move used shell `mv` + a narrow `git add`, which silently dropped the deletion of
   the old path.** `mv old new && git add new/` stages the new location as an addition but leaves the
   old location's removal unstaged, because `git add` on a path that no longer exists on disk does
   nothing. This produced a first commit that *duplicated* the change instead of *moving* it, caught
   only by re-running `git status` afterward and requiring a second, follow-up commit to fix.
   **Pipeline change:** either use `git mv` directly (which stages both sides atomically), or make
   `/myflow-finish`'s own archive-and-commit snippet stage the whole `openspec/changes/` subtree
   (`git add -A -- openspec/changes/`) rather than only the new archive path.
3. **The SDD progress ledger I kept (`.superpowers/sdd/ledger.md`) didn't match the path
   `preserve-session-records.sh` looks for** (`.superpowers/sdd/tasks/progress.md`, per
   `subagent-driven-development`'s `sdd-workspace` naming convention), so it was never preserved and
   was lost with the worktree at cleanup. The substantive record survived elsewhere (the panel record
   was preserved; this conversation transcript has the per-group ledger lines), but the raw file is
   gone. **Pipeline change:** `/myflow-do` should either use the exact conventional ledger path when
   it tracks SDD progress outside the harness's own task list, or `preserve-session-records.sh`
   should be told the actual path in use rather than assuming the convention.

## Cost

Eleven subagent dispatches total: 4 implementer groups (tightly-coupled task bundling — 12 plan
tasks compressed into 4 dispatches rather than one per checkbox), 6 parallel review-panel slots, and
1 fix dispatch. The panel breadth (6 of 7 possible slots) was driven by objective, stated thresholds
(>300 changed lines, >200 changed lines, path/file handling) rather than a subjective call, so it
wasn't over-dispatched for this diff's actual size — and it found a real Critical bug (no
fenced-code-block awareness) that a narrower panel plausibly would have missed, since only the
Adversarial slot's specific "test theater" hunting turned it up. The parallel dispatch of all 6 panel
slots in one message, rather than sequentially, was the single biggest wall-clock saver in the run.
Controller-performed direct verification (grep + Python repro scripts) instead of dispatching a
second reviewer for the scoped re-review was a deliberate, cheaper substitute — every pass-1 finding
already carried a concrete repro command in the panel table, so re-running those directly cost a few
tool calls instead of a full reviewer dispatch, with no loss of rigor.

## What went well

- Grouping tightly-coupled tasks (1.1–1.3+5.1, 2.1–2.2, 3.1–3.2, 4.1–4.2) into 4 implementer
  dispatches instead of 12 kept review overhead proportionate without losing per-group scoping.
- Dogfooding the new guard against the change's own `tasks.md` at three separate points (after the
  first implementation, after the fix round, and via the guardrail's own future `/myflow-start` runs)
  caught nothing wrong here but is a cheap, repeatable check worth keeping as a habit for any
  self-referential guard change in this repo.
- Independently reproducing every panel finding with a standalone repro command before dispatching
  the fix, and again after the fix landed, meant the "panel clean" claim in the handoff rests on
  direct evidence rather than trusting either the panel's or the fixer's own report.

## Automation candidates

- The worktree-setup gap (finding 1) and the archive-move gap (finding 2) are both concrete,
  reproducible pipeline defects, not one-off mistakes — worth fixing in the shared skills rather than
  re-discovering per run.
- The ledger-path mismatch (finding 3) is smaller but has the same shape: a convention that exists
  only as an unstated assumption between two different files.

## Findings filed

- Worktree setup should move staged planning artifacts into the new worktree — **filed as
  [KAN-56](https://tweety53.atlassian.net/browse/KAN-56)**, linked to KAN-13.
- Archive move should use `git mv` or stage the whole `openspec/changes/` subtree — **filed as
  [KAN-57](https://tweety53.atlassian.net/browse/KAN-57)**, linked to KAN-13.
- SDD ledger path convention mismatch with `preserve-session-records.sh` — **filed as
  [KAN-58](https://tweety53.atlassian.net/browse/KAN-58)**, linked to KAN-13.
