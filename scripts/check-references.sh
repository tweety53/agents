#!/usr/bin/env bash
# check-references.sh — fail when a cross-referenced section no longer exists
# in the file it is referenced from.
#
# Rule: for every backticked .md/.mdc path that resolves to a real file, if the
# line ASSOCIATES one or more **bold tokens** with that path, at least one of
# those associated tokens must match a `#`, `##`, `###`, or `####` heading in
# that file.
#
# "Associated" is the load-bearing word. The rule originally checked every bold
# token on the line against every path on the line, which is not what a
# cross-reference looks like: a line that says "**Never** commit — see
# `rules/x.mdc`" was required to have a section called "Never". That produced 28
# failures on this repo's own tree, none of them a genuinely stale reference,
# and the suppression marker was then the only way to silence them — which in
# turn switched off the real checks sharing those lines. So a bold token counts
# only when it actually sits next to the path, in one of the shapes this repo
# writes references in:
#
#   **Section** (`path`)          — parenthesised
#   **Section** in `path`         — short connector
#   see/per/under **Section** … `path`
#   `path` — sections **A**, **B** — path first, tokens after
#
# See crAssociated in stats/internal/guard/references.go for the mechanical
# test. Both directions of error are lenient by construction (an unassociated
# token neither satisfies nor fails a path), so the guard's power comes from
# recognising the reference shapes, not from blanket coverage.
#
# Takes no arguments: the scan set lives in one place (crTargets in
# stats/internal/guard/references.go), so no call site can narrow it. Lines
# carrying `refs-guard:allow` are skipped — use it for a line whose bold text
# is emphasis rather than a section name.
#
# Ported to Go (KAN-778): the body's reasoning for every step is in
# stats/internal/guard/references.go. The binary does not live in this
# checkout, so this shim exports FLOW_GUARD_REPO_ROOT — this script's parent
# directory, derived as the bash guard derived its REPO_ROOT — as the root
# scanned when CHECK_REFERENCES_ROOT is unset.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-references: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-references 2 "check-references:" "$@"
