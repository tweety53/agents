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
# fixture instead of editing the real Makefile into a broken state. The
# mechanism is stats/internal/guard/uitestoverrides.go; its test is
# stats/internal/guard/check_uitest_overrides_test.go, and the shim exports
# this repository's root as FLOW_GUARD_REPO_ROOT for the default path.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-uitest-overrides: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-uitest-overrides 2 "check-uitest-overrides:" "$@"
