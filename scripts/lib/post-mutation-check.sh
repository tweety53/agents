# scripts/lib/post-mutation-check.sh — snapshot_tree_state and
# check_tree_restored, the post-guard sanity check KAN-448 part 1
# mechanizes: a guard that mutates the working tree snapshots
# `git status --porcelain=v2` and `git stash list` before its first
# mutation and checks afterwards that the tree still matches the
# snapshot — any NEW stash entry or status line is residue the guard
# names and fails on, instead of leaving a conductor to retry blind
# (KAN-423's incident).
#
# Sourced, never executed. Both functions are read-only. Bash 3.2 floor:
# indexed arrays only, no associative arrays. Every empty-array expansion
# is guarded with the ${arr[@]+"${arr[@]}"} idiom — under `set -u` on
# bash 3.2 a bare "${arr[@]}" over an empty array is an unbound-variable
# error. Callers prefix their own program name on the output; this
# library prints bare findings, one per line.
#
# Only NEW entries are drift: a stash entry the snapshot recorded and
# that is still present afterwards is expected, and a deleted entry is
# not residue the contract names.

POST_MUTATION_CHECK_SENTINEL='--- post-mutation-check: stash section ---'

# snapshot_tree_state <worktree> — print the snapshot on stdout: every
# `git status --porcelain=v2` line, the sentinel, then every
# `git stash list` line. The caller captures it in a variable.
snapshot_tree_state() {
  git -C "$1" status --porcelain=v2
  printf '%s\n' "$POST_MUTATION_CHECK_SENTINEL"
  git -C "$1" stash list
}

# check_tree_restored <worktree> <snapshot> — recompute status and stash
# list, diff both against <snapshot>, and print one finding per new
# stash entry (`new stash entry: ...`) and per unexpected status line
# (`unexpected status line: ...`). Returns 0 when nothing was printed
# (the tree matches the snapshot), 1 on any drift, 2 when git itself
# failed on the recompute.
check_tree_restored() {
  local worktree="$1" snapshot="$2" line prev in_stash=0 drift=0 found
  local snap_status=() snap_stash=() now_status now_stash
  while IFS= read -r line; do
    [ -n "$line" ] || continue
    if [ "$line" = "$POST_MUTATION_CHECK_SENTINEL" ]; then
      in_stash=1
    elif [ "$in_stash" -eq 0 ]; then
      snap_status+=("$line")
    else
      snap_stash+=("$line")
    fi
  done < <(printf '%s\n' "$snapshot")

  now_status="$(git -C "$worktree" status --porcelain=v2)" || return 2
  now_stash="$(git -C "$worktree" stash list)" || return 2

  while IFS= read -r line; do
    [ -n "$line" ] || continue
    found=1
    for prev in ${snap_stash[@]+"${snap_stash[@]}"}; do
      if [ "$prev" = "$line" ]; then found=0; break; fi
    done
    if [ "$found" -eq 1 ]; then
      printf 'new stash entry: %s\n' "$line"
      drift=1
    fi
  done < <(printf '%s\n' "$now_stash")

  while IFS= read -r line; do
    [ -n "$line" ] || continue
    found=1
    for prev in ${snap_status[@]+"${snap_status[@]}"}; do
      if [ "$prev" = "$line" ]; then found=0; break; fi
    done
    if [ "$found" -eq 1 ]; then
      printf 'unexpected status line: %s\n' "$line"
      drift=1
    fi
  done < <(printf '%s\n' "$now_status")

  return "$drift"
}
