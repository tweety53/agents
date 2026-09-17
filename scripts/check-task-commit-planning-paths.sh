#!/usr/bin/env bash
# check-task-commit-planning-paths.sh — fail any task commit that swept the
# planning paths (KAN-553, observed twice on kan-468: staged planning
# artifacts reached the task commit through the implementer's own `git
# commit`, and the second `git reset --hard` recovery destroyed uncommitted
# planning work alongside them).
#
# Usage:
#
#   check-task-commit-planning-paths.sh <worktree> <base>
#
# A "task commit" is a commit in `<base>..HEAD` carrying the pipeline's
# `Task-Id:` trailer — the commit an implementer makes for a plan task.
# The planning paths are the spec tree's changes directory — the leaf
# resolved per project through scripts/lib/spec-root.sh, `spectre` or
# `openspec`, never hardcoded — and `docs/superpowers/`. Anything else
# under the spec tree is implementation: a capability spec under
# `<leaf>/specs/` is exactly what a task commit SHOULD carry, and
# `docs/superpowers/` is named in full rather than as `docs/` for the same
# reason. The trailerless planning commit bare /flow makes at integrate
# sweeps exactly these paths and stays outside this contract by having no
# trailer; so does everything at or before <base>.
#
# Verdict lines, on stdout:
#
#   TASK-COMMIT-SWEEP: <short sha> <task id> <path>   per planning path a
#                                                     task commit touched
#   PLANNING-PATHS-CLEAN: <worktree> — <n> task commit(s) checked
#   PLANNING-PATHS-SWEPT: <worktree> — <n> task commit(s)
#
# Exit 0 on the clean verdict; exit 1 on the swept verdict; exit 2 with
# NOTHING on stdout when the arguments are missing, the worktree is not a
# readable git repository, HEAD or <base> does not resolve, or git refuses
# the walk — an inability is never reported as a verdict.
#
# Like the other guards that need a change in flight and a real worktree
# passed in (check-foreign-staged.sh and its siblings), this one answers a
# question about one change, not about the state of the repository's text,
# and is deliberately not a `## lint` step; its harness under
# scripts/test-check-task-commit-planning-paths.sh is what `## test` runs.
#
# Bash 3.2 is the floor: indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/spec-root.sh"

usage() {
  echo "usage: check-task-commit-planning-paths.sh <worktree> <base>" >&2
  exit 2
}

[ "$#" -eq 2 ] || usage
WT="$1"
BASE="$2"

[ -d "$WT" ] || {
  echo "check-task-commit-planning-paths.sh: not a readable directory: $WT" >&2
  exit 2
}
git -C "$WT" rev-parse --is-inside-work-tree >/dev/null 2>&1 || {
  echo "check-task-commit-planning-paths.sh: not a git repository: $WT" >&2
  exit 2
}
git -C "$WT" rev-parse --verify --quiet HEAD >/dev/null || {
  echo "check-task-commit-planning-paths.sh: HEAD does not resolve in: $WT" >&2
  exit 2
}
git -C "$WT" rev-parse --verify --quiet "$BASE^{commit}" >/dev/null || {
  echo "check-task-commit-planning-paths.sh: base does not resolve to a commit: $BASE" >&2
  exit 2
}

leaf="$(spec_root_leaf "$WT")"

checked=0
swept=0
sweep_lines=""

# One walk, one record per commit: sha NUL Task-Id value NUL. A commit with
# no Task-Id trailer reads as an empty value and is not a task commit.
while IFS= read -r -d '' sha && IFS= read -r -d '' tid; do
  tid="$(printf '%s' "$tid" | tr -d '[:space:]')"
  [ -n "$tid" ] || continue
  checked=$((checked + 1))
  commit_swept=0
  while IFS= read -r -d '' path; do
    case "$path" in
      "$leaf"/changes/*|docs/superpowers/*)
        sweep_lines="${sweep_lines}TASK-COMMIT-SWEEP: ${sha:0:12} $tid $path"$'\n'
        commit_swept=1
        ;;
    esac
  done < <(git -C "$WT" diff-tree -r --no-commit-id --name-only -z "$sha")
  [ "$commit_swept" -eq 0 ] || swept=$((swept + 1))
done < <(git -C "$WT" log --format='%H%x00%(trailers:key=Task-Id,valueonly,unfold)%x00' "$BASE..HEAD")

if [ "$swept" -gt 0 ]; then
  printf '%s' "$sweep_lines"
  printf 'PLANNING-PATHS-SWEPT: %s — %d task commit(s)\n' "$WT" "$swept"
  exit 1
fi
printf 'PLANNING-PATHS-CLEAN: %s — %d task commit(s) checked\n' "$WT" "$checked"
exit 0
