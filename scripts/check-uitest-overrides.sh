#!/usr/bin/env bash
# check-uitest-overrides.sh — fail when a UITEST_* Makefile variable is
# assigned without `override`.
#
# Why this exists: none of stats/Makefile's ui-test-up / ui-test-down targets
# take parameters, and several of the UITEST_* variables they use are read
# straight into a `dropdb --force`, an unqualified `rm -rf`, or a `kill` on a
# pid read from a file. A plain assignment is silently overridden by a
# command-line or environment value (`make ui-test-up UITEST_LOG=/etc/passwd`,
# or an exported UITEST_LOG under `make -e`), which lets a caller aim any of
# those at an arbitrary path or process. Fixing that has recurred three times
# already, one variable at a time (UITEST_DB, then UITEST_ID, then the rest) —
# this guard checks the whole class at once so a variable added later cannot
# reintroduce the same defect unnoticed.
#
# Rule: every line assigning a UITEST_* variable in the scanned file(s) must
# begin with `override`.
#
# WHAT "ASSIGNS" MEANS. Earlier versions of this guard matched only a plain
# `:=` anchored at column 0, which passed every one of these while GNU Make
# still treats each as an ordinary, command-line-overridable assignment:
#   - a leading SPACE before the variable name (`  UITEST_EVIL := ...`) — a
#     leading TAB is different: that is a recipe line, evaluated by the
#     shell, not by Make, so it is deliberately NOT treated as an assignment
#     here;
#   - `?=` (conditional), `+=` (append), `=` (recursive), `::=` (POSIX
#     simple).
# The pattern below matches Make's assignment grammar instead: optional
# leading spaces, `UITEST_<NAME>`, optional whitespace, then any of those
# five operators. A line is flagged unless the UITEST_* name is immediately
# preceded on the line by `override ` — i.e. `override UITEST_X := ...`
# never matches the violation pattern, because the line does not begin
# (after leading spaces) with `UITEST_`.
#
# THREE MORE FORMS, added after a live check found each beat the guard
# above while still winning a command-line override in GNU Make:
#
#   - `export UITEST_X := ...` / `unexport UITEST_X` — the plain-assignment
#     pattern above requires the line to *begin* (after leading spaces)
#     with `UITEST_`, so `export ` in front hides it from that pattern,
#     while GNU Make still lets a command-line assignment win. The accepted
#     protected form is `override export UITEST_X := ...` (override
#     first) — confirmed against GNU Make 4.4.1 that `override` there still
#     beats a command-line reassignment; the macOS-shipped GNU Make 3.81
#     mishandles `override`+`export` combined regardless of order, which is
#     a bug in that decades-old release, not a reason to trust the
#     combined form less — verified separately, not fixed here (out of
#     this guard's scope, and not a UITEST_* codepath).
#   - `$(eval UITEST_X := ...)` — text (`$(eval `) precedes `UITEST_` on
#     the line, same evasion as `export` above. The accepted protected
#     form is `$(eval override UITEST_X := ...)`.
#   - `define UITEST_X` / `endef` — a `define`d variable is an ordinary
#     recursive variable to Make and is beatable from the command line
#     exactly like `:=`, confirmed live against stats/Makefile's
#     UITEST_SAFE_RMRF macro (`make -n ui-test-down
#     'UITEST_SAFE_RMRF=@echo INJECTED $(1)'` replaced the entire safety
#     `case` block with the injected text). The accepted protected form is
#     `override define UITEST_X` / `endef`. Lines inside a `define`...
#     `endef` body are macro text, not Make assignments, and are not
#     scanned as if they were — only the `define` line itself is checked
#     for `override`.
#
# INCLUDED FILES. `include` / `-include` / `sinclude` directives naming a
# literal path (no `$(...)` expansion) are followed recursively, resolved
# relative to the directory of the file that names them, so a UITEST_*
# variable moved into an included file is still caught. A target that
# contains `$` — a Make variable or function reference this guard cannot
# evaluate without running Make itself — is reported as unresolved and
# skipped; an included file that does not exist on disk is likewise reported
# and skipped. LIMITATION, stated here rather than left silent: wildcarded
# includes (`include $(wildcard *.mk)`) and any include path built from a
# Make variable are not followed, so a UITEST_* variable hidden behind one of
# those would not be caught. stats/Makefile has no `include` directive today.
#
# Usage: check-uitest-overrides.sh [<makefile>]
# With no argument, checks stats/Makefile in this repository — the one real
# caller (`make test`, `make check-uitest-overrides`). An explicit argument
# exists so the guard's own test harness can point it at a throwaway
# fixture instead of editing the real Makefile into a broken state.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MAKEFILE="${1:-$REPO_ROOT/stats/Makefile}"

if [ ! -f "$MAKEFILE" ]; then
  echo "check-uitest-overrides: $MAKEFILE not found" >&2
  exit 2
fi

VIOLATIONS=0

# SCANNED tracks resolved, canonical file paths already scanned so that a
# cycle of `include` directives (A includes B, B includes A) terminates
# instead of recursing forever, and so a file reachable by two different
# include chains is not double-reported.
SCANNED=()

already_scanned() {
  local target="$1" f
  for f in "${SCANNED[@]}"; do
    [ "$f" = "$target" ] && return 0
  done
  return 1
}

scan_file() {
  local file="$1"
  local base_dir
  base_dir="$(cd "$(dirname "$file")" && pwd)"
  local line_no=0
  local line
  # Set while scanning the body of a `define UITEST_X` ... `endef` block —
  # skips every other check until `endef`, since that body is macro text,
  # not Make assignments; only the `define` line itself is checked.
  local in_define=0

  while IFS= read -r line || [ -n "$line" ]; do
    line_no=$((line_no + 1))

    if [ "$in_define" -eq 1 ]; then
      if [[ "$line" =~ ^[\ ]*endef([[:space:]]|$) ]]; then
        in_define=0
      fi
      continue
    fi

    # `define UITEST_X` (optionally `override define UITEST_X`) — see the
    # header comment. A `define`d variable is an ordinary recursive
    # variable to Make, beatable from the command line exactly like `:=`,
    # unless `override` precedes `define`.
    if [[ "$line" =~ ^[\ ]*(override[[:space:]]+)?define[[:space:]]+(UITEST_[A-Za-z0-9_]*) ]]; then
      in_define=1
      if [ -z "${BASH_REMATCH[1]}" ]; then
        echo "check-uitest-overrides: $file:$line_no: ${BASH_REMATCH[2]} is defined without \`override\` — a caller can reassign it on the command line or via the environment" >&2
        VIOLATIONS=$((VIOLATIONS + 1))
      fi
      continue
    fi

    # `include`/`-include`/`sinclude`, only when the directive itself is not
    # inside a recipe (a recipe line begins with a TAB, never a plain space).
    if [[ "$line" =~ ^[\ ]*(-include|sinclude|include)[[:space:]]+(.+)$ ]]; then
      local targets="${BASH_REMATCH[2]}"
      local t resolved
      for t in $targets; do
        if [[ "$t" == *'$'* ]]; then
          echo "check-uitest-overrides: $file:$line_no: include target '$t' uses Make variable/function expansion this guard cannot resolve — not scanned" >&2
          continue
        fi
        case "$t" in
          /*) resolved="$t" ;;
          *) resolved="$base_dir/$t" ;;
        esac
        if [ ! -f "$resolved" ]; then
          echo "check-uitest-overrides: $file:$line_no: included file not found, not scanned: $resolved" >&2
          continue
        fi
        resolved="$(cd "$(dirname "$resolved")" && pwd)/$(basename "$resolved")"
        already_scanned "$resolved" && continue
        SCANNED+=("$resolved")
        scan_file "$resolved"
      done
      continue
    fi

    # `export UITEST_X := ...` / `unexport UITEST_X := ...`, optionally
    # `override export UITEST_X := ...` — see the header comment. Checked
    # before the plain-assignment pattern below because that pattern
    # requires the line to start with `UITEST_`, which an `export`/
    # `unexport` keyword in front defeats.
    if [[ "$line" =~ ^[\ ]*(override[[:space:]]+)?(export|unexport)[[:space:]]+(UITEST_[A-Za-z0-9_]*)[[:space:]]*(::=|:=|\+=|\?=|=) ]]; then
      if [ -z "${BASH_REMATCH[1]}" ]; then
        echo "check-uitest-overrides: $file:$line_no: ${BASH_REMATCH[3]} is assigned via \`${BASH_REMATCH[2]}\` without \`override\` — a caller can reassign it on the command line or via the environment" >&2
        VIOLATIONS=$((VIOLATIONS + 1))
      fi
      continue
    fi

    # Assignment grammar: optional leading spaces (never a tab — that is a
    # recipe line), UITEST_<NAME>, optional whitespace, then one of Make's
    # five assignment operators. A line beginning with `override ` does not
    # match here, because the match requires UITEST_ to start the line
    # (after leading spaces only).
    if [[ "$line" =~ ^[\ ]*UITEST_[A-Za-z0-9_]*[[:space:]]*(::=|:=|\+=|\?=|=) ]]; then
      local var
      var="$(printf '%s' "$line" | sed -E 's/^[ ]*(UITEST_[A-Za-z0-9_]*).*/\1/')"
      echo "check-uitest-overrides: $file:$line_no: $var is assigned without \`override\` — a caller can reassign it on the command line or via the environment" >&2
      VIOLATIONS=$((VIOLATIONS + 1))
    fi

    # `$(eval UITEST_X := ...)`, optionally `$(eval override UITEST_X :=
    # ...)` — see the header comment. Searched anywhere on the line, not
    # anchored to its start, since the evading text is `$(eval ` itself
    # sitting in front of `UITEST_`.
    if [[ "$line" =~ \$\(eval[[:space:]]+(override[[:space:]]+)?(UITEST_[A-Za-z0-9_]*)[[:space:]]*(::=|:=|\+=|\?=|=) ]]; then
      if [ -z "${BASH_REMATCH[1]}" ]; then
        echo "check-uitest-overrides: $file:$line_no: ${BASH_REMATCH[2]} is assigned inside \`\$(eval ...)\` without \`override\` — a caller can reassign it on the command line or via the environment" >&2
        VIOLATIONS=$((VIOLATIONS + 1))
      fi
    fi
  done < "$file"
}

SCANNED+=("$MAKEFILE")
scan_file "$MAKEFILE"

if [ "$VIOLATIONS" -ne 0 ]; then
  printf 'UITEST-OVERRIDES-FAIL: %d UITEST_* variable(s) assigned without override\n' "$VIOLATIONS"
  exit 1
fi

printf 'UITEST-OVERRIDES-OK: every UITEST_* variable is override\n'
exit 0
