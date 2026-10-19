#!/usr/bin/env bash
# check-identical-baselines.sh — name the PNGs of one directory that are
# byte-identical, so the visual-verify verifier explains each group
# (kan-998: the pad-up and pad-down baselines were one image, and only a
# reviewer saw it).
#
# Usage: resolve-visual-screenshots.sh … | check-identical-baselines.sh
#
# Takes no argument. Its input is resolve-visual-screenshots.sh's output on
# stdin: absolute PNG paths, one per line; blank lines are skipped. Paths are
# grouped by directory, then by the SHA-256 of their bytes — a capture and
# its own baseline in different directories never group.
#
# Prints to stdout:
#   IDENTICAL-NONE: <n> PNGs                          exit 0
#   identical: <dir>: <name> <name> …                 exit 1, one line per
#                                                     group, names sorted,
#                                                     lines sorted by
#                                                     directory then first name
#
# Exit 0 no group, 1 at least one group, 2 cannot answer (an argument given,
# no path on stdin, a path that cannot be read) — stdout empty, the cause on
# stderr.
#
# Go, in flow-guard (stats/internal/guard/identicalbaselines.go).
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-identical-baselines: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-identical-baselines 2 "check-identical-baselines:" "$@"
