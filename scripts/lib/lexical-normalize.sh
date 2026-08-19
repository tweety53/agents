# scripts/lib/lexical-normalize.sh — lexically_collapse, defined once.
#
# Extracted (F34, pass 7 of this change's own review panel) from two
# byte-for-byte-identical component-stack walks: gather-dispatch-context.sh's
# own lexical_normalize() and gather-self-review-context.sh's
# validate_archived_path() step 2. Each function's own doc comment already
# said the only real difference between them was what a relative input gets
# joined onto before the walk runs — gather-dispatch-context.sh joins to
# this process's own cwd; gather-self-review-context.sh joins to a
# separately-derived trusted root, never a raw cwd read (its own step 2
# comment explains why: sidestepping an OS-level path alias such as macOS's
# /tmp -> /private/tmp) — which is exactly the "make it absolute" half of
# the job, not the "collapse . and .. once it is absolute" half this file
# factors out. That second half never actually differed, so hand-copying it
# carried the same drift risk within_root's own extraction (F22, this
# change's own review panel) already closed for the boundary check — a
# future fix to the walk (an edge case in `..` handling, say) applied to
# one copy and missed in the other.
#
# Sourced, not carried as an inline copy, by both callers above — the same
# "SAFELY REACH IT" criterion scripts/lib/resolve-file.sh's own header
# states, and the same one scripts/lib/within-root.sh already relies on:
# both callers ship through the skills/*/scripts/ symlink farm
# (gather-dispatch-context.sh via skills/myflow-do/scripts/ and
# skills/myflow-fast/scripts/; gather-self-review-context.sh via
# skills/myflow-fast/scripts/ and skills/myflow-finish/scripts/), each
# alongside its own `lib` symlink into scripts/lib/, so a guard reached
# through either entry point finds this file exactly where it expects it.
#
# Not meant to be executed directly — a caller sources it and calls
# lexically_collapse; it sets no `set -euo pipefail` of its own and relies
# on the sourcing script's.
#
# lexically_collapse <abs-path> -> prints <abs-path> with every `.` and
# `..` component removed, by pure string manipulation (a component-stack
# walk: push each `/`-separated segment, pop on `..`, never past the root,
# skip `.` and empty segments) that touches the filesystem nowhere, so it
# cannot itself be fooled by a symlink. <abs-path> is assumed to already
# begin with "/" — making a relative input absolute, and against what root,
# is each caller's own decision, made before calling this, and stays there:
# the two callers disagree about it (cwd vs a trusted root) for reasons
# specific to each, which is exactly why that half was left out of this
# shared function rather than folded in and parameterized.
#
# `IFS=/ read -ra`, never an unquoted `for part in $abs_path` (both
# original copies already carried this fix, each citing the other's own
# finding number): bash applies pathname (glob) expansion as well as
# word-splitting to an unquoted expansion, so a `*`/`?`/`[...]` in the
# input would expand against the process's cwd and corrupt this "pure
# string manipulation" step with real filesystem entries; `read` only
# field-splits on $IFS and never globs, closing that gap entirely.
lexically_collapse() {
  local abs="$1"
  local -a stack=() parts=()
  local part
  IFS=/ read -ra parts <<< "$abs"
  for part in "${parts[@]}"; do
    case "$part" in
      "" | ".") continue ;;
      "..")
        if [ "${#stack[@]}" -gt 0 ]; then
          unset "stack[$((${#stack[@]} - 1))]"
        fi
        ;;
      *) stack+=("$part") ;;
    esac
  done

  local result="" seg i
  i=0
  while [ "$i" -lt "${#stack[@]}" ]; do
    seg="${stack[$i]}"
    [ -n "$seg" ] && result="$result/$seg"
    i=$((i + 1))
  done
  [ -n "$result" ] || result="/"
  printf '%s\n' "$result"
}
