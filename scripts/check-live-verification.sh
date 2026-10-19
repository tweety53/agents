#!/usr/bin/env bash
# check-live-verification.sh — refuse a live-verification record that carries
# no measured before figure (kan-998: the record read "Before figure: not
# measured on a merge-base stack", and **Live check** accepted it).
#
# Usage: check-live-verification.sh <changeRoot>/live-verification.md
#
# A line NAMES a before figure when it carries the word `before`, any case, as
# a whole word (`beforehand` is not one). A table header `| Before |` counts —
# lenient by design: the guard cannot tell a header from a figure. ANY line is
# UNMEASURED when it carries `not measured`, `unmeasured` or `not taken`, any
# case, as whole words — whether or not it names `before`, so a table whose
# `Before` header sits above a `not measured` cell fails too.
#
# Prints to stdout:
#   LIVE-VERIFICATION-OK: <record>                       exit 0
#   <record>: no before figure — no line names one       exit 1
#   <record>:<n>: before figure not measured — <line>    exit 1, one per line
#
# Exit 0 a before figure is named and none is unmeasured, 1 otherwise,
# 2 cannot answer (a usage error, a record that cannot be read) — stdout empty,
# the cause on stderr.
#
# Go, in flow-guard (stats/internal/guard/liveverification.go).
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-live-verification: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-live-verification 2 "check-live-verification:" "$@"
