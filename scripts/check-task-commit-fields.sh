#!/usr/bin/env bash
# check-task-commit-fields.sh — thin wrapper.
#
# All field-parsing and diff-checking logic lives in the Go port,
# stats/internal/guard/taskcommitfields.go, which ports
# check-task-commit-fields.py (Python 3, standard library only) — the parser
# the bash body exec'd, split out the same way check-task-build-green.sh is
# and for the same reason. That .py stays, loaded as a module by
# check-task-records.py and check-plan-shape.py. This file exists only so an
# operator's muscle memory invoking this exact filename, and
# .flow/project.md's declared commands, keep working, while the field
# grammar underneath gets a real parser.
#
# Unlike check-task-build-green.sh (which resolves WHICH tasks.md files to
# scan, zero or more of them, with no per-task identity), this guard checks
# ONE task's fields against ONE real commit, so its calling convention names
# the task and commit explicitly rather than scanning:
#
#   check-task-commit-fields.sh <worktree> <task-id> <commit-sha> [parent-sha] [canonical-worktree] [change-name]
#
# This wrapper's own job is resolving WHICH tasks.md the named task lives
# in, among the non-archived ones under
# <worktree>/spectre/changes/*/tasks.md. Zero matches there is not
# automatically a refusal any more (KAN-363 task 9) — see WHERE A LINK IS
# FOLLOWED below — and more than one ROOT change is still a refusal: two
# changes neither of which is the other's sub-change, which nothing here may
# guess between. The optional sixth argument, <change-name>, skips that
# ambiguity scan entirely (KAN-367): when given, this wrapper resolves
# directly against <worktree>/spectre/changes/<change-name> — still honoring
# its own -fix-N sibling and its own link.md, but never looking at any other
# directory under spectre/changes/ — and takes priority over the glob in
# stats/internal/guard/taskcommitfields.go.
#
# WHERE A LINK IS FOLLOWED. A satellite change directory (spectre task 1's
# `link.md`, carrying `## Part of`) has no `tasks.md` of its own by design,
# so the glob above finds nothing at all under a satellite worktree — the
# exact shape that used to be an unconditional "no tasks.md found" refusal.
# When that glob comes back empty, and exactly one LINK-ONLY change
# directory exists under `<worktree>/<spec-root>/changes/` — one whose own
# `link.md` exists and whose own `tasks.md` does not — this wrapper resolves
# that satellite's plan through `scripts/lib/change-plan.sh`'s rules (task
# 7; ported to stats/internal/guard/changeplan.go),
# passing the optional fifth argument straight through as its
# canonical-worktree, before giving up. More than one link-only directory,
# or a resolution that fails, still ends in the same "no tasks.md found"
# refusal as before.
#
# THE ABSENT-DIR SHAPE (KAN-260). A cross-repo worktree may carry no change
# directory at all — the plan lives only in the canonical repo. With the
# change name passed as the sixth argument, the named-change path hands the
# question to change_plan_path, whose absent-dir branch resolves the
# same-named plan from the SUPPLIED canonical worktree; with no canonical
# worktree, or no name, the refusal stands, because nothing here names
# where the plan lives.
#
# A LINK-ONLY DIRECTORY IS NEVER COUNTED TOWARD THE ROOT-CHANGE AMBIGUITY
# TEST in stats/internal/guard/taskcommitfields.go either, for the same reason a `<name>-fix-N` sibling is not: a
# satellite carries no `tasks.md`, so the `*/tasks.md` glob that test is
# built from never sees it — a satellite sitting beside a genuine root
# change changes nothing about which root that test resolves to.
#
# SYMLINKS. The file and directory tests used throughout the Go port, as
# `-f` and `-d` in scripts/lib/change-plan.sh, follow symlinks — so a symlink at
# `<worktree>/<spec-root>/changes/<allowlisted-name>` is followed and its
# content read as that change's plan. check-unfinished-work.sh's own header
# accepts the identical tradeoff for the identical reason; this is not a
# hole this change opens.
#
# SPECTRE'S LAYOUT GUARANTEES NOTHING LIKE "exactly one active change", and
# this header used to say it did. That assumption is what broke. A
# <name>-fix-N sub-change is a FLAT SIBLING of its parent under
# spectre/changes/ — `spectre new` refuses an id that is not a single flat
# directory name — so the glob matches the parent AND the sibling the moment
# a fix round opens one, and a wrapper refusing on "more than one" took this
# guard out of service on exactly the runs it was added for. A fix sibling is
# therefore not ambiguity: the highest-numbered one resolves, or the root when
# there is none. The mechanism, the digit test that keeps a merely
# similarly-named change out of it, and two caveats on the numbering are in
# stats/internal/guard/taskcommitfields.go.
# EXIT CODES. 0 — the commit passes. 1 — a verdict against the commit:
# violations print on stdout, one per line. 2 — could not judge: a caller or
# environment mistake (usage, unreadable plan, missing worktree, a
# git call that cannot evaluate its arguments), never a judgment about the
# commit. Every exit-2 refusal prints the same unmistakable stderr opening,
# "COULD NOT JUDGE — not a commit verdict:", so a caller mistake is never
# read at a glance as the guard refusing the commit — the KAN-170 failure,
# where a merge base mistyped by one character surfaced git's bare
# "ambiguous argument" message through this guard's exit and cost a second
# look to tell a typo from a real defect.

# Header corrected for the Go port (KAN-760): the logic lives in Go and a
# missing Python module is no longer a refusal, since the grammar is
# compiled in. The body's reasoning is in
# stats/internal/guard/taskcommitfields.go.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-task-commit-fields: COULD NOT JUDGE — not a commit verdict: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-task-commit-fields 2 "check-task-commit-fields: COULD NOT JUDGE — not a commit verdict:" "$@"
