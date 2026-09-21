# Self-review context bundle for kan-620-flow-fix-extend-the-on-top-fix-commit-route-to

found: 1 of 7 sources; skipped: 6 of 7 sources
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-620-flow-fix-extend-the-on-top-fix-commit-route-to.md (absent)
skipped: .superpowers/sdd/reviews/kan-620-flow-fix-extend-the-on-top-fix-commit-route-to-panel.md (absent)
skipped: spectre/changes/archive/kan-620-flow-fix-extend-the-on-top-fix-commit-route-to/tasks.md (absent)
skipped: spectre/changes/archive/kan-620-flow-fix-extend-the-on-top-fix-commit-route-to/design.md (absent)
skipped: spectre/changes/archive/kan-620-flow-fix-extend-the-on-top-fix-commit-route-to/narrative.md (absent)

## git log --stat

commit fddd71fa81cc0db64847f4bbdf3bb4fd8e02b687
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 22:30:00 2026 +0300

    docs(flow): take a pushed branch's gated-reviewer fix as a new commit, not a rewrite

 skills/flow/implement.md | 22 +++++++++++++++-------
 1 file changed, 15 insertions(+), 7 deletions(-)

## Session narrative

A `/flow-fast` creating run for KAN-620, filed from kan-563's deferred self-review: extend the
on-top fix-commit route to `skills/flow/implement.md`'s gated per-task reviewer fix path. All
three toggles rolled — `plan-class.sh` classified the one-task plan `micro`, so the decide
collapsed to its recorded defaults (inline, no implementer, no panel) and the parent edited the
fix path directly: a branch the remote already holds — the normal case under **Branch backup** —
now takes each task's fix as one plain pathspec-scoped commit at the tip with no sha rewritten,
the re-review pass reading its fix commit's own diff, while the existing fixup + explicit-base
autosquash mechanics, conflict rule and rewritten-sha guard re-run stay for unpushed history
only; the route is cited to **Panel re-runs** (`skills/flow/review-panel.md`) rather than
restated, and the stage-close phrase "any fix folded" became route-neutral. Judgment calls named:
the ticket's `FIX_BASE` clause was not carried over, because the panel's round diff is what
consumes that tip and this path's re-review reads per-fix-commit ranges — a recorded round tip
would be dead state; and the commit is `docs(flow)` not the plan's first-draft `fix(flow)`,
matching kan-563's own commit for the panel half of this same route. Where it struggled: picking
the citable anchor — the route paragraph is a bold lead, not a heading, so `check-references.sh`
would reject citing it directly and **Panel re-runs** is the heading that carries it; and the
flow store went unreachable mid-run, dropping every later stage mark to the local journal
(flush pending at landing, one warning line, no mark lost). The full `## lint` list, the
normative inventory (byte-identical before and after) and `run-guard-tests.sh` (77/77) all pass
in the worktree.
