#!/usr/bin/env bash
# preserve-session-records.sh — copy a change's session records out of the
# gitignored worktree and into the repository, so they survive worktree
# removal.
#
# Usage: preserve-session-records.sh <worktree> <change-name> <state-dir>
#
# Every path it touches arrives as an argument, so it runs correctly from any
# working directory.
#
# Prints one line per source: "preserved: <dest>" or "skipped: <src> (absent)".
# A MISSING SOURCE IS NEVER FATAL. A change may legitimately have no panel
# record, and a preservation step able to block an integration would be a worse
# failure than the gap it closes. Exit non-zero when a copy was attempted and
# could not be written, and when a source or a destination was REFUSED as
# untrustworthy (see the protections below) — and even then the remaining
# sources are still attempted, because one bad path must not cost the other
# records. Every non-zero path names itself on stderr; a `skipped:` line on
# stdout is the only outcome that is silent, and it is the only one that means
# nothing was wrong.
#
# The destination date is fixed at the FIRST copy: an existing file for this
# change is reused, so a fix round overwrites in place instead of leaving one
# dated duplicate per round.
#
# THREE PATH PROTECTIONS ARE EXPLICIT HERE RATHER THAN INCIDENTAL, because this
# script runs automatically from /myflow-do's push path and /myflow-finish
# run 1, and every path it reads and writes is built from an argument:
#   1. The change name must match an allowlist of characters a real change name
#      uses — one leading alphanumeric, then letters, digits, `.`, `_` and `-`.
#      Two hazards make one check. A name containing `/` was previously blocked
#      only by an accident of string concatenation (the `$TODAY-` prefix shares
#      the path component with the start of the name, so a leading `../` became
#      part of a literal directory name), and a refactor could remove that
#      accident without anyone noticing it was load-bearing. A name carrying a
#      glob metacharacter defeats the digit-anchored `find` below: `*` matched
#      and overwrote a DIFFERENT change's preserved record.
#   2. Every destination directory is required to resolve INSIDE the worktree.
#      The three directories under docs/superpowers/ are ordinary tracked repo
#      paths, editable in any pull request. If one is a symlink, `mkdir -p` is a
#      no-op and `cp` follows it, writing the record outside the repository.
#   3. Every SOURCE is required to resolve inside the root it comes from — the
#      worktree for the two `.superpowers/` sources, the state directory for the
#      artifact. Validating only the destination left an arbitrary-file-READ:
#      `[ -f ]` follows symlinks and `cp` reads through them, and `.superpowers/`
#      being gitignored is no protection, because `.gitignore` only gates
#      untracked paths — a symlink forced in with `git add -f` lands on disk when
#      the branch is checked out, and run 1 then commits and pushes whatever it
#      pointed at. A source that resolves outside its root is a REFUSAL: it is
#      reported on stderr and non-zero for that source, deliberately not the
#      silent `skipped:` path, which means "this change legitimately has no such
#      record".
#
# CHECK AND WRITE USE THE SAME RESOLVED PATH. Both boundary checks resolve a
# path and then act through the resolved form, never through the argument, so
# `cp` cannot re-follow a symlink component the check already validated. What
# remains is narrower and stated rather than claimed away: the resolved
# directory itself could still be replaced between the check and the write.
# There is no way to close that with `cp` alone; the point of resolving first is
# that no *component* of the path is re-followable.
set -euo pipefail

WORKTREE="${1:-}"
NAME="${2:-}"
STATE_DIR="${3:-}"

if [ -z "$WORKTREE" ] || [ -z "$NAME" ] || [ -z "$STATE_DIR" ]; then
  echo "usage: preserve-session-records.sh <worktree> <change-name> <state-dir>" >&2
  exit 2
fi

# Protection 1. One check for both hazards, so there is no second place for a
# name rule to drift to. A name outside the allowlist is programmer error, not a
# recoverable input: exit 2 like the usage failure above, before any directory is
# created or any source is read. Change names are a lowercased Jira key plus a
# kebab-case slug, or the slug alone (skills/myflow-contracts/jira-integration.md
# defines the shape), so the allowlist costs nothing a real name needs.
#
# THE ALLOWED CHARACTERS ARE ENUMERATED RATHER THAN WRITTEN AS RANGES, and that
# is the whole of this gate's locale independence. `A-Z` inside a bracket
# expression is a COLLATING range, not a byte range, so what it admits is
# whatever the ambient locale's collation puts between the two endpoints.
# Measured on bash 3.2 (Darwin 25.5.0) with this file's previous
# `[!A-Za-z0-9]* | *[!A-Za-z0-9._-]*`: under `LC_ALL=C` and `ru_RU.UTF-8` the
# names `écho`, `İstanbul`, `ﬀoo`, `ⅰx`, `Ａbc` and `ⅹ` were all refused; under
# `en_US.UTF-8`, `de_DE.UTF-8` and `tr_TR.UTF-8` every one of them was ADMITTED.
# A containment gate whose accepted set changes with the operator's environment
# is not a containment gate, and this one is the only thing standing between a
# pull-request-editable change name and a `find -name` pattern and a constructed
# path below.
#
# A LITERAL LIST HAS NO ENDPOINTS, so there is nothing for a collation order to
# reorder: membership is membership in every locale. That is why the fix is not
# `export LC_ALL=C`, which check-unfinished-work.sh does carry for reasons of its
# own — pinning the whole script's locale would reach every other locale-aware
# thing it does, and in check-cleanup-complete.sh it would be EXPORTED into the
# project's own `survivors` command, changing the locale a project's tooling runs
# under. The enumeration touches nothing outside these two patterns and needs no
# environment at all, so one written form is correct in all three guards. The
# accepted set is unchanged from what the range form accepted under `LC_ALL=C`,
# which is the set every harness already asserts.
#
# The `-` stays LAST in the second pattern, where a bracket expression reads it
# as a literal rather than as the start of a range.
case "$NAME" in
  [!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789]* \
  | *[!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-]*)
    echo "preserve-session-records: change name '$NAME' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    exit 2
    ;;
esac

# Protection 2's and 3's reference points. Resolved once, with symlinks
# followed, so every comparison below is between two fully-resolved paths.
WORKTREE_REAL="$(cd -P "$WORKTREE" 2>/dev/null && pwd -P)" || {
  echo "preserve-session-records: worktree $WORKTREE is not a readable directory" >&2
  exit 2
}

# The state directory is the artifact source's root. Unlike the worktree it is
# not required to exist: a change that never published a proposal has none, and
# that is the `skipped:` case, not a failure. An empty root value means no
# source under it can be trusted, which `preserve` treats as a refusal.
STATE_DIR_REAL="$(cd -P "$STATE_DIR" 2>/dev/null && pwd -P)" || STATE_DIR_REAL=""

TODAY="$(date -u +%Y-%m-%d)"

# resolve_file <path> — print <path> with every symlink resolved, both on the
# final component and on each directory component. `readlink -f` would do this
# in one call on GNU but is not portable to the BSD readlink macOS ships, so the
# final component is walked here and the directory components are left to
# `cd -P`. Fails on a symlink loop rather than spinning.
resolve_file() {
  local p="$1" hops=0 target dir
  while [ -L "$p" ]; do
    hops=$((hops + 1))
    [ "$hops" -le 40 ] || return 1
    target="$(readlink "$p")" || return 1
    case "$target" in
      /*) p="$target" ;;
      *) p="$(dirname "$p")/$target" ;;
    esac
  done
  dir="$(cd -P "$(dirname "$p")" 2>/dev/null && pwd -P)" || return 1
  printf '%s\n' "$dir/${p##*/}"
}

# preserve <source> <source-root> <dest-dir> <suffix>
# <suffix> is what follows the change name, e.g. ".md" or "-panel.md".
preserve() {
  local src="$1" src_root="$2" dest_dir="$3" suffix="$4" existing dest dest_real src_real
  if [ ! -f "$src" ]; then
    echo "skipped: $src (absent)"
    return 0
  fi
  # Protection 3. The source exists, so from here a refusal is a failure and
  # never a skip — the distinction the caller and both call sites rely on.
  if [ -z "$src_root" ]; then
    echo "preserve-session-records: refusing to read $src — its root directory does not resolve" >&2
    return 1
  fi
  src_real="$(resolve_file "$src")" || {
    echo "preserve-session-records: refusing to read $src — it does not resolve to a real path" >&2
    return 1
  }
  case "$src_real" in
    "$src_root"/*) : ;;
    *)
      echo "preserve-session-records: refusing to read $src — it resolves to $src_real, outside $src_root" >&2
      return 1
      ;;
  esac
  if ! mkdir -p "$dest_dir"; then
    echo "preserve-session-records: could not create $dest_dir" >&2
    return 1
  fi
  # Protection 2. Refuse a destination that does not resolve under the
  # worktree. Compared as a path BOUNDARY, not a string prefix: a bare prefix
  # test accepts `/foo/bar-evil` against `/foo/bar`. A refusal is a failure
  # (return 1), not a skip — a skip is reserved for a source that is legitimately
  # absent, whereas this is a write that cannot be trusted. The remaining
  # sources are still attempted, as they are for any other write failure.
  dest_real="$(cd -P "$dest_dir" 2>/dev/null && pwd -P)" || {
    echo "preserve-session-records: could not resolve $dest_dir" >&2
    return 1
  }
  case "$dest_real" in
    "$WORKTREE_REAL" | "$WORKTREE_REAL"/*) : ;;
    *)
      echo "preserve-session-records: refusing to write $src — destination $dest_dir resolves to $dest_real, outside the worktree $WORKTREE_REAL" >&2
      return 1
      ;;
  esac
  # Everything from here uses $dest_real, never $dest_dir: the boundary check
  # above validated the resolved directory, so searching or writing through the
  # unresolved argument would let `cp` re-follow the symlink components the check
  # just cleared — a check on one path and a write through another.
  #
  # The date is matched digit by digit rather than with a leading `*`: a bare
  # `*-${NAME}${suffix}` also matches a DIFFERENT change whose name ends in
  # this one ("2020-01-01-other-demo.md" for the change "demo"), and reusing
  # that path would overwrite another change's preserved record. That anchoring
  # holds only because Protection 1 has already excluded glob metacharacters
  # from $NAME. Sorted so the choice among several is deterministic rather than
  # filesystem order.
  existing="$(find "$dest_real" -maxdepth 1 \
    -name "[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]-${NAME}${suffix}" \
    2>/dev/null | sort | head -1 || true)"
  if [ -n "$existing" ]; then
    dest="$existing"
  else
    dest="$dest_real/$TODAY-$NAME$suffix"
  fi
  if ! cp "$src_real" "$dest"; then
    echo "preserve-session-records: could not write $dest" >&2
    return 1
  fi
  echo "preserved: $dest"
}

# The three sources. The ledger's `tasks/` component is not a coincidence:
# superpowers derives that directory from the plan file's basename, and under
# myflow the plan is always `tasks.md`, so `.superpowers/sdd/tasks/progress.md`
# is stable for every myflow change. The panel record's path is the one
# skills/myflow-do/SKILL.md writes.
RC=0
preserve "$WORKTREE/.superpowers/sdd/tasks/progress.md" "$WORKTREE_REAL" \
  "$WORKTREE/docs/superpowers/ledgers" ".md" || RC=1
preserve "$WORKTREE/.superpowers/sdd/final-review-panel.md" "$WORKTREE_REAL" \
  "$WORKTREE/docs/superpowers/reviews" "-panel.md" || RC=1
preserve "$STATE_DIR/$NAME-proposal-artifact.html" "$STATE_DIR_REAL" \
  "$WORKTREE/docs/superpowers/artifacts" ".html" || RC=1

exit "$RC"
