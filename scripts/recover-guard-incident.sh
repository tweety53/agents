#!/usr/bin/env bash
# recover-guard-incident.sh — reproduce KAN-423's incident recovery on
# demand: abort a stuck `git revert`, then restore the untracked planning
# files the incident stash holds, unstaged, via `git show "stash@{0}^3:<f>" >
# <f>` redirects.
#
# THE ORDER IS THE POINT. `git checkout stash@{...} -- <path>` stages the
# restore, and any later `revert --abort` discards it — the incident's first
# restore attempt was lost exactly that way. Here every abort precedes every
# restore, and the redirects leave the files unstaged, so nothing after the
# abort point can discard them.
#
# Usage: recover-guard-incident.sh [--apply] [repo-dir] [path...]
#
#   --apply    execute; without it, print what would run and change nothing
#   repo-dir   defaults to the caller's cwd; must be a git repository
#   path...    planning paths to restore, relative to the repo root;
#              defaults to `docs/research` and `spectre/changes`
#
# Exit 0 on a printed dry-run plan or a completed apply; 1 on a failed
# precondition (cause on stderr, stdout empty); 2 on a usage error — an
# unknown option, a repo-dir that is not a directory, or a repo-dir that is
# not a git repository.
#
# PRECONDITIONS ARE CHECKED IN ORDER AND THE FIRST FAILURE NAMES ITS CAUSE
# ON STDERR WITH NOTHING ON STDOUT: no revert in progress (REVERT_HEAD
# missing); no stash at all; a stash created without
# `-u` (no untracked third parent — the remedy is re-stashing with -u BEFORE
# any abort, so the two causes are named differently); a restore target
# tracked in the index or HEAD, which refuses the whole run at the first
# offending file.
# A target already present untracked is restored over, named as an
# overwrite in the plan — untracked state is never clobbered silently, and
# tracked state is never clobbered at all.
#
# THE REFLOG BLOCK PRINTS IN EVERY MODE, dry-run and apply alike — the
# incident was diagnosed from `git reflog -g HEAD`'s alternating reset
# pattern, so the diagnosis is visible at the point of decision, not only
# in the mode that acts on it.
#
# Bash 3.2 is the floor: indexed arrays only, no associative arrays.
set -euo pipefail

SELF="recover-guard-incident"

die() { printf '%s: %s\n' "$SELF" "$1" >&2; exit "${2:-1}"; }

usage() {
  printf 'usage: %s [--apply] [repo-dir] [path...]\n' "$SELF"
  printf '  --apply    execute; default is a dry-run plan\n'
  printf '  repo-dir   git repository, default cwd\n'
  printf '  path...    planning paths, repo-root-relative; default docs/research spectre/changes\n'
}

APPLY=0
if [ "${1:-}" = "--help" ]; then usage; exit 0; fi
if [ "${1:-}" = "--apply" ]; then APPLY=1; shift; fi
for a in "$@"; do
  case "$a" in -*) die "unknown option: $a" 2 ;; esac
done

# F1 contract: repo-dir (or the cwd) resolves to the toplevel, so path...
# is repo-root-relative however deep inside the repo the tool is invoked.
if [ "$#" -gt 0 ]; then
  DIR="$(cd "$1" 2>/dev/null && pwd -P)" || die "not a directory: $1" 2
  shift
else
  DIR="$(pwd -P)"
fi
REPO="$(git -C "$DIR" rev-parse --show-toplevel 2>/dev/null)" \
  || die "not a git repository: $DIR" 2

[ "$#" -gt 0 ] || set -- docs/research spectre/changes

git -C "$REPO" rev-parse -q --verify REVERT_HEAD >/dev/null 2>&1 \
  || die "no revert in progress in $REPO (REVERT_HEAD missing) — nothing to recover"

if ! git -C "$REPO" rev-parse -q --verify "stash@{0}" >/dev/null 2>&1; then
  die "no stash entry in $REPO — recovery restores stash@{0}^3, which needs a stash"
fi
if ! git -C "$REPO" rev-parse -q --verify "stash@{0}^3" >/dev/null 2>&1; then
  die "stash@{0} has no untracked third parent — it was not created with 'git stash -u'; re-stash with -u before any abort"
fi

# The restore set: every file the stash's third parent holds under the given
# paths, each checked against the working tree. Tracked anywhere (index or
# HEAD) -> refuse the whole run; already present untracked -> restore over
# it, named as an overwrite; absent -> a plain restore.
FILES=""
OVERWRITES=""
for p in "$@"; do
  while IFS= read -r f; do
    [ -n "$f" ] || continue
    if [ -n "$(git -C "$REPO" ls-files -- "$f")" ] \
       || git -C "$REPO" cat-file -e "HEAD:$f" 2>/dev/null; then
      die "refusing: $f is tracked in $REPO — recovery never clobbers tracked state" 1
    fi
    FILES="${FILES}${f}
"
    if [ -e "$REPO/$f" ]; then
      OVERWRITES="${OVERWRITES}${f}
"
    fi
  done < <(git -C "$REPO" ls-tree -r --name-only "stash@{0}^3" -- "$p")
done
[ -n "$FILES" ] || die "stash@{0}^3 holds no files under: $*"

printf 'reflog diagnosis — the alternating reset pattern the incident showed:\n'
git -C "$REPO" reflog -g HEAD -n 15 || true

printf 'plan:\n'
printf '  git -C %s revert --abort\n' "$REPO"
while IFS= read -r f; do
  [ -n "$f" ] || continue
  case $'\n'"$OVERWRITES" in *$'\n'"$f"$'\n'*)
    printf '  (overwrites an existing untracked file: %s)\n' "$f"
    ;;
  esac
  printf '  mkdir -p %s/%s\n' "$REPO" "$(dirname "$f")"
  printf '  git -C %s show "stash@{0}^3:%s" > %s/%s\n' "$REPO" "$f" "$REPO" "$f"
done <<<"$FILES"

if [ "$APPLY" -eq 1 ]; then
  printf 'running: git -C %s revert --abort\n' "$REPO"
  git -C "$REPO" revert --abort
  while IFS= read -r f; do
    [ -n "$f" ] || continue
    printf 'running: git show "stash@{0}^3:%s" > %s/%s\n' "$f" "$REPO" "$f"
    mkdir -p "$REPO/$(dirname "$f")"
    git -C "$REPO" show "stash@{0}^3:$f" > "$REPO/$f"
  done <<<"$FILES"
fi
