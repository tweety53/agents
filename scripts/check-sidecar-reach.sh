#!/usr/bin/env bash
# check-sidecar-reach.sh — pre-flight every `.mockups` sidecar line against
# the capture spec beside it, before the capture runs (kan-998: stale sidecar
# lines surfaced only at compose-mockup-frames.sh, after a 13-minute capture).
#
# Usage: check-sidecar-reach.sh <capture root> [<mockups root>]
#
# Walks <capture root>, never descending into `node_modules`, `.git` or
# `.worktrees`. A SIDECAR is a file ending `.mockups`; its spec is that path
# without the suffix. A sidecar line is `<screenshot name> <frame id>`; blank
# and `#` lines are skipped.
#
# A name is REACHED when it equals a string literal of the spec ('…', "…" or
# `…`), or fully matches a backtick template literal containing `${`, each
# `${…}` hole standing for one or more characters
# (`cardio-guide-${frame}.png` reaches `cardio-guide-G1.png`).
# Stated limit: a name built outside any literal (concatenation, a computed
# variable), or through a `${…}` hole holding a `}` (the hole ends at its
# first `}`), reads as unreached.
#
# With <mockups root>, a reached line carrying a second field <frame id> must
# also name a frame that exists: <mockups root>/<frame id>.png (kan-1007). A
# line with one field is not judged on its frame.
#
# Prints to stdout, paths relative to <capture root>, sorted by sidecar then
# line:
#   SIDECAR-REACH-OK: <n> lines in <m> sidecars                      exit 0
#   <sidecar>:<line>: <name> — taken by no screenshot in <spec>      exit 1
#   <sidecar>:<line>: <name> — <spec> does not exist                 exit 1
#   <sidecar>:<line>: <name> — frame <id> has no <mockups root>/<id>.png  exit 1
# (<mockups root> printed as given).
#
# Exit 0 every line reached, 1 at least one not, 2 cannot answer (a usage
# error, a capture or mockups root that is not a directory, an unreadable
# sidecar) — stdout
# empty, the cause on stderr.
#
# Go, in flow-guard (stats/internal/guard/sidecarreach.go).
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-sidecar-reach: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-sidecar-reach 2 "check-sidecar-reach:" "$@"
