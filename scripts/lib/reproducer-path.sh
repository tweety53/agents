#!/usr/bin/env bash
# reproducer-path.sh
#
# The single source of truth for the recorded-form refusals on a
# reproducer's worktree-relative path: the path must be relative to the
# worktree, must not begin with a dash, and must carry no `..` segment.
# prove-reproducer.sh sources and applies this check at its door, before it
# copies the script into a scratch worktree — a `..` or an absolute path
# there would point the copy step outside the tree the leg runs in, the
# same escape the refusals exist to prevent.
#
# run-reproducer.sh and check-panel-reproducers.sh keep their own, older
# copies of these refusals beside their wider lexical sets (shell
# metacharacters, URLs, NUL bytes, resolved-symlink containment) — the same
# split reproducer-metachars.sh records for its own extraction: only what
# is verbatim-identical moves here; each of those scripts applies the ban
# in its own context and at its own layer, and their wider sets are not
# this rule.
#
# Usage: `source "<dir of this file>/reproducer-path.sh"`, then
# `reproducer_path_refusal <path>` prints nothing and returns 0 when the
# path is acceptable, or one reason line on stdout and returns 1 when it is
# not. Not executable on its own.
reproducer_path_refusal() {
  local rel="$1"
  case "$rel" in
    "") printf 'reproducer path is empty\n'
       return 1 ;;
    /*) printf 'reproducer path must be relative to the worktree, got an absolute path: %s\n' "$rel"
        return 1 ;;
    -*) printf 'reproducer path may not begin with a dash: %s\n' "$rel"
        return 1 ;;
    ..|../*|*/..|*/../*) printf 'reproducer path may not contain a .. segment: %s\n' "$rel"
                         return 1 ;;
  esac
  return 0
}
