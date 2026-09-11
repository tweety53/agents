#!/usr/bin/env bash
# classify-untracked.sh — the archive pre-flight that sorts a landing
# worktree's untracked entries into classes instead of leaving one flat
# refusal count behind (KAN-411; the implementation half of KAN-397).
#
# Usage: classify-untracked.sh <landing-worktree>
#
# Run immediately before prepare-archive-branch.sh, against the same
# <landing-worktree> path. The guard downstream still refuses a dirty tree —
# this script's job is to make the classes it can settle never reach that
# refusal, and to name the rest for the conductor and the operator:
#
#   capture   stray screenshot images from fix rounds (*.png, *.jpg, *.jpeg)
#             — moved to <project>/.worktrees/_scratchpad/, outside the
#             landing worktree, so the throwaway worktree's forced removal
#             cannot destroy them. A same-named file already in the
#             scratchpad is never clobbered; the incoming copy gains a
#             timestamp prefix.
#   config    local configuration — a `.claude` entry at the worktree root
#             — appended to the checkout's LOCAL exclude file
#             (<common git dir>/info/exclude), never to a committed
#             .gitignore: one project's local residue must not become every
#             clone's ignore rule. That exclude file is uncommitted but
#             shared by every worktree of this checkout — a linked
#             worktree's own <git-dir>/info/exclude is NOT read by status
#             on this machine's git (measured, 2.50.1) — so `.claude/` at
#             any worktree root, the main checkout's included, stops
#             showing in status. That scope is the point: `.claude/` is
#             exactly the residue no checkout of this repository wants
#             tracked. The files stay on disk, now invisible to status.
#   asset     everything else — reported, never touched. The conductor
#             prompts the operator: commit it to the archive branch
#             deliberately, or delete it. Until then it stays untracked and
#             prepare-archive-branch.sh keeps refusing, which is the gate
#             working as designed.
#
# Classification reads `git ls-files --others --exclude-standard --directory`:
# untracked-and-unignored entries exactly, directories collapsed to one
# `<dir>/` line. A directory entry is therefore classified whole — by its own
# name for `.claude`, otherwise as an asset, never by a file inside it; no
# whole directory is moved or deleted on the strength of one file's
# extension. Tracked modifications and staged entries are not this script's
# concern and still refuse downstream, as they always have.
#
# stdout carries one line per action taken or finding reported, and nothing
# else; a clean worktree prints exactly `CLEAN`. An absent <landing-worktree>
# prints nothing — there is nothing to classify, and the guard creates it
# fresh and clean.
#
#   Exit 0   Could answer: clean, classified, or an absent worktree.
#   Exit 2   Cannot answer — an argument is missing, the path is not a
#            directory or not a git worktree, or the repository's common
#            directory cannot be resolved. This script has no exit 1: it
#            never refuses anything, it only settles and reports; refusing
#            stays prepare-archive-branch.sh's job.
set -euo pipefail

export LC_ALL=C

LANDING="${1:-}"

if [ -z "$LANDING" ]; then
  echo "usage: classify-untracked.sh <landing-worktree>" >&2
  exit 2
fi

if [ ! -e "$LANDING" ]; then
  # Nothing to classify; prepare-archive-branch.sh will create it fresh.
  exit 0
fi

if [ ! -d "$LANDING" ]; then
  echo "classify-untracked: $LANDING is not a directory" >&2
  exit 2
fi

if ! git -C "$LANDING" rev-parse --git-dir >/dev/null 2>&1; then
  echo "classify-untracked: $LANDING is not a git worktree" >&2
  exit 2
fi

COMMON="$(git -C "$LANDING" rev-parse --path-format=absolute --git-common-dir 2>/dev/null)" || {
  echo "classify-untracked: cannot resolve the common git directory of $LANDING" >&2
  exit 2
}
SCRATCH="$(dirname "$COMMON")/.worktrees/_scratchpad"

ENTRIES=()
while IFS= read -r entry; do
  [ -n "$entry" ] || continue
  ENTRIES[${#ENTRIES[@]}]="$entry"
done <<EOF
$(git -C "$LANDING" ls-files --others --exclude-standard --directory)
EOF

# "${ENTRIES[@]}" is unset-expansion-unsafe under `set -u` on bash 3.2 when
# empty — the same floor test-prepare-archive-branch.sh's cleanup notes.
if [ "${#ENTRIES[@]}" -eq 0 ]; then
  echo "CLEAN"
  exit 0
fi

for entry in "${ENTRIES[@]}"; do
  base="$(basename "$entry")"
  case "$entry" in
    .claude | .claude/)
      EXCLUDE="$COMMON/info/exclude"
      mkdir -p "$(dirname "$EXCLUDE")"
      if ! grep -qxF "$entry" "$EXCLUDE" 2>/dev/null; then
        printf '%s\n' "$entry" >>"$EXCLUDE"
      fi
      echo "IGNORED: $entry (local exclude)"
      ;;
    *.png | *.PNG | *.jpg | *.JPG | *.jpeg | *.JPEG)
      mkdir -p "$SCRATCH"
      dest="$SCRATCH/$base"
      if [ -e "$dest" ]; then
        dest="$SCRATCH/$(date +%Y%m%d-%H%M%S)-$base"
      fi
      mv "$LANDING/$entry" "$dest"
      echo "CAPTURED: $entry -> $dest"
      ;;
    *)
      echo "ASSET: $entry (operator decides — commit deliberately or delete)"
      ;;
  esac
done

