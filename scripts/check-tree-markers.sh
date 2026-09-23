#!/usr/bin/env bash
# check-tree-markers.sh snapshot|verify <worktree> <markers-file> <snapshot-file>
#
# The guard KAN-643 asked for: a subagent's self-reported clean state is
# never evidence for anything — every dispatch that can touch the worktree
# is bracketed by CONTENT MARKERS recorded before the dispatch and
# re-grepped after it, and a mismatch is read as the subagent having
# mutated the tree. On gymie KAN-579 a reviewer destroyed never-committed
# planning artifacts and self-reported the tree restored; nothing
# independent checked the claim until the deferred self-review grepped the
# tree by hand against strings the artifacts were known to contain. This
# guard is that grep, recorded.
#
#   snapshot  run immediately BEFORE the dispatch's Agent call — greps each
#             marker and writes one <path><TAB><marker><TAB><status> line
#             per spec line to <snapshot-file>
#   verify    run when the dispatch's report file exists, BEFORE any report
#             or verdict is read or acted on — recomputes the same lines
#             and compares byte-for-byte
#
# <markers-file> carries one <path><TAB><ERE> line per artifact the
# dispatch must not damage, each ERE a string the artifact is known to
# contain — a decision ID, a task title, a section heading. Paths are
# worktree-relative and are refused before grep ever runs when they are
# absolute or carry a `..` component: the markers file names what grep
# reads, so it is a containment boundary exactly like the change-name case
# block in check-plan-unchanged.sh, and it is validated the same way —
# refused, not sanitised.
#
# MARKERS SIT BESIDE check-plan-unchanged.sh, never in its place: that
# guard asserts git's view of the plan tree only, while markers pin known
# content anywhere in the tree, independent of git entirely — content that
# was never committed (the KAN-579 artifacts) has no git answer at all.
# Both guards' verify exit 1 stops the run the same way, and for the same
# reason: the dispatch's own clean-or-restored prose is not evidence, so
# recovery is by hand from before the dispatch, never by re-running it.
#
# Exit codes:
#   0  snapshot written, or every recorded marker matches its snapshot
#   1  verify found the tree's markers changed since the snapshot — each
#      difference named on stderr
#   2  cannot answer at all — missing arguments, a bad mode, a non-directory
#      worktree, an unreadable markers file, a malformed spec line (no tab,
#      an empty path or marker), a path that is absolute or carries a `..`
#      component, a grep failure on an existing file, an unwritable
#      snapshot path (on snapshot), or a missing or unreadable snapshot
#      file (on verify)
set -euo pipefail

export LC_ALL=C

MODE="${1:-}"
WORKTREE="${2:-}"
MARKERS="${3:-}"
SNAP="${4:-}"

usage() {
  echo "check-tree-markers: usage: check-tree-markers.sh snapshot|verify <worktree> <markers-file> <snapshot-file>" >&2
  exit 2
}

[[ "$MODE" = "snapshot" || "$MODE" = "verify" ]] || usage
[[ -n "$WORKTREE" && -d "$WORKTREE" ]] || { echo "check-tree-markers: not a directory: ${WORKTREE:-<missing>}" >&2; exit 2; }
[[ -n "$MARKERS" && -f "$MARKERS" && -r "$MARKERS" ]] || {
  echo "check-tree-markers: no readable markers file: ${MARKERS:-<missing>}" >&2
  exit 2
}
[[ -n "$SNAP" ]] || usage

WORKTREE="$(cd "$WORKTREE" && pwd -P)" || { echo "check-tree-markers: worktree vanished: $WORKTREE" >&2; exit 2; }

# CONTAINMENT: the markers file names what grep reads. An absolute path or
# any `..` component would read outside the worktree, so the spec is
# refused rather than sanitised — the parent fixes its marker list, the
# guard never widens it.
check_spec() {
  local path="$1" marker="$2" component
  [[ -n "$path" ]] || { echo "check-tree-markers: empty path in markers file" >&2; exit 2; }
  case "$path" in
    /*) echo "check-tree-markers: absolute path '$path' in markers file -- refused" >&2; exit 2 ;;
  esac
  IFS=/ read -ra components <<< "$path"
  for component in "${components[@]}"; do
    if [[ "$component" = ".." ]]; then
      echo "check-tree-markers: path '$path' carries a '..' component -- refused" >&2
      exit 2
    fi
  done
  [[ -n "$marker" ]] || { echo "check-tree-markers: empty marker for '$path' in markers file" >&2; exit 2; }
}

# One snapshot line per spec line: <path><TAB><marker><TAB><status>, where
# status is the literal `missing` or the grep -cE line count. grep -c
# exits 1 on a zero count -- that is a valid answer, recorded as 0; any
# other failure on an existing file cannot be answered at all.
marker_line() {
  local path="$1" marker="$2" file count rc
  check_spec "$path" "$marker"
  file="$WORKTREE/$path"
  if [[ ! -f "$file" ]]; then
    printf '%s\t%s\tmissing\n' "$path" "$marker"
    return 0
  fi
  count="$(grep -cE -- "$marker" "$file" 2>/dev/null)" && rc=0 || rc=$?
  if [[ "$rc" -gt 1 ]]; then
    echo "check-tree-markers: grep failed on $file -- cannot answer" >&2
    exit 2
  fi
  [[ -n "$count" ]] || count=0
  printf '%s\t%s\t%s\n' "$path" "$marker" "$count"
}

# Validate every spec line BEFORE anything is written: the snapshot file
# must never come into existence half-made, and a malformed line must exit
# 2 from the current shell, never from a process substitution (where an
# exit would be swallowed and an empty marker set would verify clean).
read_specs() {
  local line path marker
  while IFS= read -r line || [[ -n "$line" ]]; do
    [[ -n "$line" ]] || continue
    path="${line%%$'\t'*}"
    marker="${line#*$'\t'}"
    [[ "$marker" != "$line" ]] || {
      echo "check-tree-markers: line without a tab: $line" >&2
      exit 2
    }
    check_spec "$path" "$marker"
    printf '%s\n' "$line"
  done < "$MARKERS"
}

record_all() {
  local line path marker
  while IFS= read -r line; do
    path="${line%%$'\t'*}"
    marker="${line#*$'\t'}"
    marker_line "$path" "$marker"
  done
}

SPECS="$(mktemp "${TMPDIR:-/tmp}/check-tree-markers-specs.XXXXXX")" || {
  echo "check-tree-markers: mktemp failed -- cannot answer" >&2
  exit 2
}
OUT=""
CURRENT=""
trap 'rm -f -- "$SPECS"; [ -z "$OUT" ] || rm -f -- "$OUT"; [ -z "$CURRENT" ] || rm -f -- "$CURRENT"' EXIT
read_specs > "$SPECS"

case "$MODE" in
  snapshot)
    OUT="$(mktemp "${TMPDIR:-/tmp}/check-tree-markers-snap.XXXXXX")" || {
      echo "check-tree-markers: mktemp failed -- cannot answer" >&2
      exit 2
    }
    if ! record_all < "$SPECS" > "$OUT"; then
      echo "check-tree-markers: could not record the markers -- cannot answer" >&2
      exit 2
    fi
    if ! mv -- "$OUT" "$SNAP"; then
      echo "check-tree-markers: could not write the snapshot at $SNAP" >&2
      exit 2
    fi
    OUT=""
    echo "TREE-MARKERS-SNAPSHOT-OK: markers pinned at $SNAP"
    ;;
  verify)
    [[ -f "$SNAP" && -r "$SNAP" ]] || {
      echo "check-tree-markers: no readable snapshot at $SNAP -- cannot answer; snapshot runs before the dispatch, never after it" >&2
      exit 2
    }
    CURRENT="$(mktemp "${TMPDIR:-/tmp}/check-tree-markers-verify.XXXXXX")" || {
      echo "check-tree-markers: mktemp failed -- cannot answer" >&2
      exit 2
    }
    if ! record_all < "$SPECS" > "$CURRENT"; then
      echo "check-tree-markers: could not recompute the markers -- cannot answer" >&2
      exit 2
    fi
    if cmp -s "$SNAP" "$CURRENT"; then
      echo "TREE-MARKERS-OK: every recorded marker matches its pre-dispatch snapshot"
      exit 0
    fi
    echo "check-tree-markers: the tree's recorded markers changed across the dispatch -- the dispatch's own clean-or-restored claim is not evidence; stop and recover by hand from before the dispatch" >&2
    diff "$SNAP" "$CURRENT" | sed 's/^/check-tree-markers: /' >&2
    exit 1
    ;;
esac
