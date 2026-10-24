#!/usr/bin/env bash
# check-live-verification.sh — refuse a live-verification record that carries
# no measured before figure (kan-998: the record read "Before figure: not
# measured on a merge-base stack", and **Live check** accepted it).
#
# Usage: check-live-verification.sh <changeRoot>/live-verification.md
#        check-live-verification.sh <changeRoot>
#
# Given the change directory instead — the gate flow.verify's stage end runs
# in-process (stats/cmd/flow/stage.go, gateLiveVerification;
# skills/flow/verify-and-handoff.md, **Close the stage**) — it
# reads <changeRoot>/design.md: a design naming no live check (no `## Live
# check` section, or one whose first line opens `None`, `No ` or `Not
# applicable`) prints `LIVE-CHECK-NONE: <changeRoot>`, exit 0; a design naming
# one judges <changeRoot>/live-verification.md as below, a missing record
# printing `<record>: missing — design.md names a live check`, exit 1; a
# directory with no design.md cannot answer, exit 2 (kan-1008's self-review:
# the stage end checked nothing, and the live check first ran at integrate).
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
