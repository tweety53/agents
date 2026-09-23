#!/usr/bin/env bash
# check-plan-unchanged.sh snapshot|verify <worktree> <change-name> <snapshot-file>
#
# The guard KAN-635 asked for: the change's plan tree survives every
# reviewer dispatch, asserted by git — never by the reviewer's own prose.
# On gymie KAN-579 the gated per-task reviewer ran `git checkout <sha> -- .`
# mid-review, silently destroyed the never-committed design.md/tasks.md
# working-tree edits (task commits never carry plan paths, per
# check-task-commit-planning-paths.sh, so the working tree held the only
# current copy), and then self-reported the tree restored — a claim nothing
# checked, and a false one. implement.md's **The plan tree survives every
# reviewer dispatch** and review-panel.md's round-dispatch paragraph wire
# the two modes around every review dispatch:
#
#   snapshot  run immediately BEFORE the dispatch's Agent call, after the
#             parent has committed any uncommitted plan edits — writes the
#             plan tree's git status to <snapshot-file>
#   verify    run when the dispatch's report file(s) exist, BEFORE any
#             verdict is read or acted on — recomputes the same status and
#             compares byte-for-byte
#
# THE GUARD IS SCOPED TO THE PLAN TREE (`spectre/changes/<name>/`),
# deliberately, and never asserts HEAD or the rest of the worktree: a
# gated reviewer bundle flies BESIDE the next implementer group (implement.md's
# bundled dispatch), whose source edits and commits legitimately move
# everything else while the reviewer is in flight — a whole-worktree assert
# would fire on every normal run, and the incident's own loss was the plan
# tree ("source code was never at risk", the KAN-635 issue records). The
# status includes untracked files, so a file planted under the plan tree is
# a violation too; the plan directory being absent at snapshot and still
# absent at verify is clean (nothing was there to lose).
#
# THE SNAPSHOT IS THE EVIDENCE, and verify is the only restore claim this
# pipeline accepts: the dispatching parent never reads the reviewer's
# "restored"/"clean" prose as an answer to what happened to the tree. Verify
# exit 1 therefore stops the run — the reviewer reported about a tree it
# also changed — and recovery is by hand from the pre-dispatch commit (which
# holds the plan tree precisely because the parent commits before it
# snapshots), never by re-running the reviewer.
#
# THE CHANGE NAME IS A PATHSPEC COMPONENT here, so the same containment
# `case` block as check-task-reviewer-single-dispatch.sh's own copy (itself
# duplicated, on purpose, from check-panel-findings-closed.sh) refuses
# anything but a plain change name before git ever sees it.
#
# Exit codes:
#   0  snapshot written, or the plan tree matches its snapshot
#   1  verify found the plan tree changed since the snapshot — each
#      difference named on stderr
#   2  cannot answer at all — missing arguments, a bad mode, a non-directory
#      worktree, a change name outside the allowlist, the worktree not being
#      a git repository, a git failure, or (on verify) a missing or
#      unreadable snapshot file or an unwritable snapshot path (on snapshot)
set -euo pipefail

export LC_ALL=C

MODE="${1:-}"
WORKTREE="${2:-}"
NAME="${3:-}"
SNAP="${4:-}"

usage() {
  echo "check-plan-unchanged: usage: check-plan-unchanged.sh snapshot|verify <worktree> <change-name> <snapshot-file>" >&2
  exit 2
}

[[ "$MODE" = "snapshot" || "$MODE" = "verify" ]] || usage
[[ -n "$WORKTREE" && -d "$WORKTREE" ]] || { echo "check-plan-unchanged: not a directory: ${WORKTREE:-<missing>}" >&2; exit 2; }
[[ -n "$NAME" ]] || usage
[[ -n "$SNAP" ]] || usage

# CONTAINMENT, identical to check-task-reviewer-single-dispatch.sh's own
# copy -- see that script's header for why this `case` block stays
# duplicated rather than centralized. The name lands in a git pathspec, so
# `../escape` and a glob metacharacter are hazards here exactly as they are
# in a state-file-derived change name there.
case "$NAME" in
  [!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789]* \
  | *[!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-]*)
    echo "check-plan-unchanged: change name '$NAME' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    exit 2
    ;;
esac

# Canonicalise before git runs against it, and refuse rather than proceed
# if it vanished between the `-d` check above and here.
WORKTREE="$(cd "$WORKTREE" && pwd -P)" || { echo "check-plan-unchanged: worktree vanished: $WORKTREE" >&2; exit 2; }

plan_status() {
  git -C "$WORKTREE" status --porcelain --untracked-files=all -- "spectre/changes/$NAME"
}

case "$MODE" in
  snapshot)
    if ! plan_status > "$SNAP"; then
      rm -f "$SNAP"
      echo "check-plan-unchanged: git status failed in $WORKTREE -- cannot answer (is it a git repository?)" >&2
      exit 2
    fi
    echo "PLAN-SNAPSHOT-OK: spectre/changes/$NAME status pinned at $SNAP"
    ;;
  verify)
    [[ -f "$SNAP" && -r "$SNAP" ]] || {
      echo "check-plan-unchanged: no readable snapshot at $SNAP -- cannot answer; snapshot runs before the dispatch, never after it" >&2
      exit 2
    }
    current="$(mktemp "${TMPDIR:-/tmp}/check-plan-unchanged-verify.XXXXXX")" || {
      echo "check-plan-unchanged: mktemp failed -- cannot answer" >&2
      exit 2
    }
    trap 'rm -f "$current"' EXIT
    if ! plan_status > "$current"; then
      echo "check-plan-unchanged: git status failed in $WORKTREE -- cannot answer (is it a git repository?)" >&2
      exit 2
    fi
    if cmp -s "$SNAP" "$current"; then
      echo "PLAN-UNCHANGED-OK: spectre/changes/$NAME matches its pre-dispatch snapshot"
      exit 0
    fi
    echo "check-plan-unchanged: spectre/changes/$NAME changed across the review dispatch -- the reviewer's own clean-or-restored claim is not evidence; stop and recover by hand from the pre-dispatch commit" >&2
    diff "$SNAP" "$current" | sed 's/^/check-plan-unchanged: /' >&2
    exit 1
    ;;
esac
