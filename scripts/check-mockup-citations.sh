#!/usr/bin/env bash
# check-mockup-citations.sh — refuse a plan citing a mockup frame no file in
# the project's mockups root holds (KAN-558, KAN-565: kan-469's design.md and
# tasks.md cited P6.png, no such frame existed, and the gap surfaced only when
# visual verification skipped the comparison).
#
# Usage: check-mockup-citations.sh <project root> <mockups root> <file>...
#
# <project root> is the directory `.flow/project.md` lives in (a worktree's
# root); <mockups root> the `mockups` value of its `## visual verification`
# table, relative to <project root>, a trailing `/` ignored
# (skills/flow-contracts/project-configuration-visual.md); each <file> a plan
# artifact — <changeRoot>/design.md and <changeRoot>/tasks.md.
#
# A CITATION is a `<mockups root>/<frame id>.png` token, <frame id> made of
# letters, digits, `_` and `-`, not preceded by a letter, digit, `_`, `.`,
# `/` or `-` — so `other/<mockups root>/X.png` is no citation. A bare
# `<frame id>.png` is not one either: plans name capture screenshots that
# way too, and nothing in the token tells the two apart. Writing plans
# (skills/flow/brainstorm-planner.md, section D) has a plan cite each frame
# by its path for that reason.
#
# Prints ONE verdict line to stdout:
#   MOCKUP-CITATIONS-RESOLVED: <n> cited
#   MOCKUP-CITATIONS-MISSING: <file>:<line> <citation>; …
#
# Exit 0 every citation names a regular file, 1 at least one does not,
# 2 cannot answer (a usage error, a <mockups root> that is not a directory
# under <project root>, an unreadable <file>).
#
# Go, in flow-guard (stats/internal/guard/mockupcitations.go).
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-mockup-citations: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-mockup-citations 2 "check-mockup-citations:" "$@"
