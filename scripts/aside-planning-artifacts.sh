#!/usr/bin/env bash
# aside-planning-artifacts.sh — set the planning paths aside before a
# pipeline rebase, restore them after (KAN-628, observed on kan-361 — a fix
# round's autosquash had to cross uncommitted design.md/tasks.md edits — and
# again on kan-579, where uncommitted planning edits from
# flow.write-in-progress blocked the pre-landing rebase. Both runs
# improvised the same temporary-WIP-commit dance; this helper is what the
# runs cite instead).
#
# Usage:
#
#   aside-planning-artifacts.sh aside <worktree>
#   aside-planning-artifacts.sh restore <worktree>
#
# The planning paths are the spec tree's changes directory — the leaf
# resolved per project through scripts/lib/spec-root.sh, `spectre` or
# `openspec`, never hardcoded — and `docs/superpowers/`, the same two paths
# check-task-commit-planning-paths.sh sweeps for. Each enters the pathspec
# only when the directory exists in <worktree>: a pathspec naming an absent
# directory would have git refuse the stash outright. `aside` stashes the
# paths' uncommitted state — tracked modifications and untracked files
# alike, `git stash push --include-untracked` — under a message carrying the
# literal marker `aside-planning-artifacts`, and nothing else: dirty
# implementation paths are left exactly where they are, for
# check-unfinished-work.sh and the guards that own them, and a worktree
# whose planning paths are already clean is reported and left untouched.
#
# `restore` refuses outright while a rebase, merge, cherry-pick or am is
# still in progress — a stash popped over an unresolved sequencer state
# would fold four kinds of confusion into one worktree. Once the rebase has
# finished or aborted, it pops the top stash only when its message carries
# the marker: an operator's own stash is never popped, and one pushed after
# the aside shadows the helper's entry, so restore reports NONE and leaves
# every entry in place — `git stash list` is always the recovery path. A
# conflicted apply (the rebase moved a planning file the aside also changed)
# is reported with exit 1 and the stash kept, which git itself does on a
# conflicted pop.
#
# Verdict lines, on stdout:
#
#   PLANNING-ARTIFACTS-CLEAN: <worktree> — nothing to set aside
#   PLANNING-ARTIFACTS-ASIDE: <worktree> — <short sha>
#   PLANNING-ARTIFACTS-RESTORED: <worktree> — <short sha>
#   PLANNING-ARTIFACTS-CONFLICT: <worktree> — <short sha> kept
#   PLANNING-ARTIFACTS-NONE: <worktree> — no aside stash on top
#
# Exit 0 on every answered verdict; exit 1 on PLANNING-ARTIFACTS-CONFLICT;
# exit 2 with NOTHING on stdout when the arguments are missing or unknown,
# <worktree> is not a git repository, git refuses an operation, or restore
# is called with a rebase/merge/cherry-pick/am still in progress — an
# inability is never reported as a verdict.
#
# Bash 3.2 is the floor: indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/spec-root.sh"

MARKER="aside-planning-artifacts"
STASH_MESSAGE="$MARKER: planning paths set aside for a rebase"

usage() {
  echo "usage: aside-planning-artifacts.sh aside <worktree>" >&2
  echo "       aside-planning-artifacts.sh restore <worktree>" >&2
  exit 2
}

refuse() {
  echo "aside-planning-artifacts.sh: $1" >&2
  exit 2
}

[ "$#" -eq 2 ] || usage
ACTION="$1"
WT="$2"

[ -d "$WT" ] || refuse "not a readable directory: $WT"
git -C "$WT" rev-parse --is-inside-work-tree >/dev/null 2>&1 ||
  refuse "not a git repository: $WT"

# The planning pathspec, built by existing-directory probes: repositioning
# the positional parameters keeps the paths as separate arguments for both
# the status call and the stash call, with no word-splitting of a joined
# string. Neither path existing is answered as CLEAN before anything runs —
# a project with neither a spec tree nor docs/superpowers/ has nothing this
# helper could set aside.
leaf="$(spec_root_leaf "$WT")"
have_spec=0; [ -d "$WT/$leaf/changes" ] && have_spec=1
have_docs=0; [ -d "$WT/docs/superpowers" ] && have_docs=1
if [ "$have_spec" -eq 1 ] && [ "$have_docs" -eq 1 ]; then
  set -- "$leaf/changes" "docs/superpowers"
elif [ "$have_spec" -eq 1 ]; then
  set -- "$leaf/changes"
elif [ "$have_docs" -eq 1 ]; then
  set -- "docs/superpowers"
else
  if [ "$ACTION" = "aside" ]; then
    printf 'PLANNING-ARTIFACTS-CLEAN: %s — nothing to set aside\n' "$WT"
    exit 0
  fi
  set --
fi

case "$ACTION" in
  aside)
    dirt="$(git -C "$WT" status --porcelain --untracked-files=normal -- "$@")" ||
      refuse "git status refused the planning-path check in: $WT"
    if [ -z "$dirt" ]; then
      printf 'PLANNING-ARTIFACTS-CLEAN: %s — nothing to set aside\n' "$WT"
      exit 0
    fi
    # git's own "Saved working directory..." line goes to stderr so the
    # stdout protocol above stays the helper's alone.
    git -C "$WT" stash push --include-untracked -m "$STASH_MESSAGE" -- "$@" 1>&2 ||
      refuse "git stash push refused in: $WT"
    sha="$(git -C "$WT" rev-parse --verify --short 'stash@{0}')" ||
      refuse "the stash it just pushed does not resolve in: $WT"
    printf 'PLANNING-ARTIFACTS-ASIDE: %s — %s\n' "$WT" "${sha:0:12}"
    exit 0
    ;;
  restore)
    # Refuse over an unresolved sequencer state. --git-path answers
    # relative to <worktree> here, so the existence probe prefixes it.
    for f in rebase-merge rebase-apply MERGE_HEAD CHERRY_PICK_HEAD; do
      p="$(git -C "$WT" rev-parse --git-path "$f")" || refuse "git rev-parse --git-path refused: $f"
      case "$p" in /*) ;; *) p="$WT/$p" ;; esac
      [ -e "$p" ] && refuse "restore refused — a rebase/merge/cherry-pick is still in progress in: $WT"
    done
    # The list is captured whole and cut to its first line afterwards, never
    # piped through `head -n 1`: under `set -o pipefail` git's SIGPIPE when
    # head exits before the list's last chunk is written reads as a refusal
    # (exit 2, nothing on stdout) on any repo whose stash list spans several
    # pipe writes.
    top="$(git -C "$WT" stash list --format='%H %gs')" ||
      refuse "git stash list refused in: $WT"
    top="${top%%$'\n'*}"
    if [ -z "$top" ]; then
      printf 'PLANNING-ARTIFACTS-NONE: %s — no aside stash on top\n' "$WT"
      exit 0
    fi
    sha="${top%% *}"
    subject="${top#* }"
    case "$subject" in
      *"$MARKER"*)
        if git -C "$WT" stash pop 1>&2; then
          printf 'PLANNING-ARTIFACTS-RESTORED: %s — %s\n' "$WT" "${sha:0:12}"
          exit 0
        fi
        # git keeps the stash entry on a conflicted pop; report it and
        # leave the recovery to the caller through git stash list.
        printf 'PLANNING-ARTIFACTS-CONFLICT: %s — %s kept\n' "$WT" "${sha:0:12}"
        exit 1
        ;;
      *)
        printf 'PLANNING-ARTIFACTS-NONE: %s — no aside stash on top\n' "$WT"
        exit 0
        ;;
    esac
    ;;
  *)
    usage
    ;;
esac
