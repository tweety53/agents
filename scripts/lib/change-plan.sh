# scripts/lib/change-plan.sh — change_plan_path and change_plan_dir, defined
# once.
#
# One owner for "where is this change's tasks.md, really" — KAN-363 task 7,
# sourced by check-unfinished-work.sh and check-task-commit-fields.sh (tasks
# 8 and 9) in place of each composing `<worktree>/<spec-root>/changes/<name>`
# directly. KAN-343 shipped a cross-repository change by hand: the second
# repository's worktree carried no plan at all, and every guard that looked
# for one there reported the absence as a verdict — `OUTSTANDING: no plan at
# <path>`, or exit 2 with nothing further — because "no tasks.md" and "this
# change's plan lives somewhere else" were the same observation to them. This
# file is what tells the two apart, by reading `link.md` (spectre task 1's
# grammar) the way a caller of `spectre validate` would.
#
# Resolution order, per design.md's `guards-take-the-canonical-worktree-path`:
#
#   1. <worktree>/<spec-root>/changes/<name>/tasks.md exists → that path.
#   2. Otherwise, <worktree>/<spec-root>/changes/<name>/link.md exists and
#      carries `## Part of`:
#        - a canonical worktree argument was passed → its own
#          <spec-root>/changes/<canonical-id>/tasks.md, if that exists;
#          no fallback to peers when the argument was passed but the plan
#          is not there — a guard that named its own canonical worktree
#          already resolved it through flow's own worktree set, and a
#          wrong answer from that set is a fact worth failing loudly on,
#          not a cue to go looking somewhere else.
#        - no canonical worktree argument → resolve the peer name through
#          <worktree>/<spec-root>/peers and print the plan there, if it
#          exists. A peer that is declared but not present on disk is not
#          a finding (design.md's `peer-absence-is-not-a-finding`) and,
#          here, not a refusal either — it is simply unresolvable, and the
#          caller decides what that means.
#   3. Neither → unresolvable.
#
# WHY THIS DOES NOT FALL BACK TO changes/archive/ THE WAY spectre's own
# `internal/check` does for the peer side of a link (task 2's
# `independent-archive`). That fallback answers "is the counterpart change
# still findable at all", which matters to `validate` reporting archive skew.
# This function answers a narrower question for a guard running inside a
# live worktree: /flow only holds a worktree open for a change still being
# worked, so an archived canonical has no worktree left to pass as the third
# argument, and a `peers`-relative archive lookup would require duplicating
# spectre's own two-step search for a case this repository's guards do not
# reach. Narrower is deliberate here, not an oversight.
#
# CONTAINMENT applies to every name concatenated into a path in this
# function: the change name (the caller's own argument, from a
# pull-request-editable state file), and the peer name and canonical change
# id (both read out of link.md, itself a pull-request-editable file this
# function parses). All three are checked against
# check-unfinished-work.sh's own allowlist — start with a letter or digit,
# then only letters, digits, '.', '_' and '-' — character for character, so
# `../escape` in any of the three positions cannot walk this function
# outside the worktree(s) it was given. `export LC_ALL=C` is the sourcing
# script's responsibility, exactly as check-unfinished-work.sh's own header
# records for its copy of this same allowlist: this file sets no locale of
# its own and relies on the caller's.
#
# SYMLINKS: `-f` and `-d` follow symlinks, so a symlink at
# `<worktree>/<spec-root>/changes/<allowlisted-name>` is followed and its
# content returned as the change's plan — the same accepted tradeoff
# check-unfinished-work.sh's own header already ships and argues.
#
# SOURCES scripts/lib/spec-root.sh ITSELF, via its own directory, rather
# than assuming a caller already sourced it first. Every existing lib/ file
# is independent of every other, but a guard here would otherwise have to
# get the sourcing ORDER right — spec-root.sh before change-plan.sh — with
# nothing enforcing it and a wrong order failing only at call time with
# "spec_root_leaf: command not found". Sourcing its own dependency removes
# that footgun; re-sourcing spec-root.sh a second time when a caller also
# sources it directly (check-unfinished-work.sh does, for its own use) is
# harmless — the same functions are defined the same way each time.
#
# Not meant to be executed directly — a caller sources it and calls
# change_plan_path or change_plan_dir; it sets no `set -euo pipefail` of its
# own and relies on the sourcing script's.

_CHANGE_PLAN_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/spec-root.sh
source "$_CHANGE_PLAN_LIB_DIR/spec-root.sh"

# _change_plan_name_ok <value> — true iff <value> is non-empty, starts with
# a letter or digit, and contains only letters, digits, '.', '_' and '-'.
# The same allowlist check-unfinished-work.sh applies to its own <name>
# argument, character for character — kept as its own copy here rather than
# sourced from that guard, matching check-cleanup-complete.sh's own copy:
# the six-line check gains nothing from being centralized a third time, and
# this file must not depend on a guard script to do library work.
_change_plan_name_ok() {
  local v="$1"
  [ -n "$v" ] || return 1
  case "$v" in
    [!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789]* \
    | *[!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-]*)
      return 1
      ;;
  esac
  return 0
}

# _change_plan_link_part_of <link.md path> — print the `## Part of` code
# span's raw content ("<peer>:<change-id>", unvalidated) to stdout and
# return 0, or return 1 when the file carries no `## Part of` section or the
# section carries no code span. Reads line by line rather than with a single
# grep/awk pipeline so the section boundary (`## Part of` up to the next
# `## ` heading or EOF) is tracked explicitly, matching how every other
# section-scoped reader in this repository's guards works.
_change_plan_link_part_of() {
  local file="$1" in_section=0 line ref=""
  while IFS= read -r line || [ -n "$line" ]; do
    if [ "$in_section" -eq 1 ]; then
      case "$line" in
        '## '*) break ;;
      esac
      case "$line" in
        *'`'*)
          ref="${line#*\`}"
          ref="${ref%%\`*}"
          [ -n "$ref" ] && break
          ;;
      esac
    fi
    case "$line" in
      '## Part of') in_section=1 ;;
    esac
  done < "$file"
  [ -n "$ref" ] || return 1
  printf '%s\n' "$ref"
}

# _change_plan_peer_root <worktree> <spec-root> <peer-name> — print the
# absolute, resolved path of <peer-name>'s tree root to stdout and return 0,
# or return 1 when <worktree>/<spec-root>/peers carries no such name, or the
# resolved path does not exist (peer declared but not present — not an
# error this function raises; the caller decides what an empty answer
# means). Paths in peers are relative to the tree's parent directory, which
# in practice means resolving relative to <worktree> itself, since every
# entry starts with `../` — the worked example in design.md's
# `peer-absence-is-not-a-finding` is the resolution this reproduces.
_change_plan_peer_root() {
  local worktree="$1" spec_root="$2" peer_name="$3"
  local peers_file="$worktree/$spec_root/peers"
  [ -f "$peers_file" ] || return 1

  local pname prel rest resolved=""
  while read -r pname prel rest || [ -n "$pname" ]; do
    [ -n "$pname" ] || continue
    if [ "$pname" = "$peer_name" ]; then
      resolved="$prel"
      break
    fi
  done < "$peers_file"
  [ -n "$resolved" ] || return 1

  ( cd "$worktree/$resolved" 2>/dev/null && pwd ) || return 1
}

# _change_plan_resolve_dir <worktree> <change-name> [canonical-worktree] —
# the shared resolution behind both public functions below. Prints the
# absolute path of the DIRECTORY that carries the resolved tasks.md (never
# the tasks.md path itself) and returns 0, or returns 1 on any of the same
# failure modes change_plan_path documents. Kept private and singular so the
# six-branch resolution order lives in exactly one place: change_plan_path
# and change_plan_dir would otherwise be two copies of the same walk, free
# to drift the moment one of them gained a case the other did not.
_change_plan_resolve_dir() {
  local worktree="$1" name="$2" canonical_worktree="${3:-}"

  _change_plan_name_ok "$name" || {
    echo "change-plan: change name '$name' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    return 1
  }

  local spec_root
  spec_root="$(spec_root_leaf "$worktree")"

  local dir="$worktree/$spec_root/changes/$name"
  if [ -f "$dir/tasks.md" ]; then
    printf '%s\n' "$dir"
    return 0
  fi

  local link="$dir/link.md"
  [ -f "$link" ] || return 1

  local ref
  ref="$(_change_plan_link_part_of "$link")" || return 1

  local peer="${ref%%:*}" changeid="${ref#*:}"
  _change_plan_name_ok "$peer" || {
    echo "change-plan: peer name '$peer' in $link is not a plain peer name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    return 1
  }
  _change_plan_name_ok "$changeid" || {
    echo "change-plan: canonical change id '$changeid' in $link is not a plain change id — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    return 1
  }

  if [ -n "$canonical_worktree" ]; then
    local canon_spec_root canon_dir
    canon_spec_root="$(spec_root_leaf "$canonical_worktree")"
    canon_dir="$canonical_worktree/$canon_spec_root/changes/$changeid"
    if [ -f "$canon_dir/tasks.md" ]; then
      printf '%s\n' "$canon_dir"
      return 0
    fi
    return 1
  fi

  local peer_root
  peer_root="$(_change_plan_peer_root "$worktree" "$spec_root" "$peer")" || return 1

  local peer_spec_root peer_dir
  peer_spec_root="$(spec_root_leaf "$peer_root")"
  peer_dir="$peer_root/$peer_spec_root/changes/$changeid"
  if [ -f "$peer_dir/tasks.md" ]; then
    printf '%s\n' "$peer_dir"
    return 0
  fi
  return 1
}

# change_plan_path <worktree> <change-name> [canonical-worktree]
#
# Prints the absolute path of the change's tasks.md and returns 0.
# Returns 1 when the change is a satellite and no canonical plan could be
# reached — the caller decides whether that is a refusal or a verdict.
change_plan_path() {
  local dir
  dir="$(_change_plan_resolve_dir "$@")" || return 1
  printf '%s\n' "$dir/tasks.md"
}

# change_plan_dir <worktree> <change-name> [canonical-worktree]
#
# Prints the absolute path of the DIRECTORY the resolved tasks.md lives in —
# <root>/<spec-root>/changes/<id>, whether <root> is <worktree> itself (a
# plain change), the supplied canonical worktree, or a peer tree reached
# through <worktree>/<spec-root>/peers — and returns 0. Returns 1 under the
# same conditions as change_plan_path.
#
# WHY A CALLER NEEDS THIS RATHER THAN `dirname` ON change_plan_path's OWN
# ANSWER. check-unfinished-work.sh's fix-sub-change sweep (task 8) globs
# `tasks.md` under a changes/ directory and matches `<id>` and `<id>-fix-*`
# there, and for a satellite that directory is never the satellite's own
# worktree — it is wherever the canonical plan actually resolved to. A
# caller could tear that answer back apart with `dirname` and `basename`,
# but this function is the same intermediate value _change_plan_resolve_dir
# already produced on the way to building change_plan_path's answer, kept
# rather than thrown away and reconstructed by every caller that needs it.
change_plan_dir() {
  _change_plan_resolve_dir "$@"
}
