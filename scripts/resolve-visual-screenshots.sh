#!/usr/bin/env bash
# resolve-visual-screenshots.sh — resolve a captured spec's PNGs, in a guard
# instead of prose.
#
# Usage: resolve-visual-screenshots.sh <project root> <capture spec basename>
#
# Replaces flow.visual-verify's step 9 resolution prose
# (skills/flow/visual-verify.md): "search recursively beneath
# `screenshots` for every PNG whose path contains the capture spec's own
# basename." Reading, still what a script cannot do, stays the agent's job —
# this guard only resolves WHICH paths to read.
#
# Prints one ABSOLUTE PNG path per line, nothing else, on a match — the
# stage feeds this straight into "read every captured PNG", so stdout here
# is data, not a report.
#
# THREE EXIT CODES:
#   0  at least one PNG matched; its path(s) are on stdout
#   1  zero matches — the case the stage must block on, per design.md. A
#      `screenshots` directory that does not exist YET is the identical
#      answer: `capture` creates it on its first run, so its absence is
#      `capture`'s own success path, not a guard failure.
#   2  cannot answer: usage error, the project root is not a directory,
#      `.flow/project.md` is missing/not a regular file/unreadable, the
#      heading scan failed to look, the section is declared twice, the
#      `screenshots` setting is absent or empty, or a declared `regression
#      checkout` does not resolve to an existing directory (so there is no
#      base to search relative to at all).
#
# `screenshots` RESOLVES RELATIVE TO THE REGRESSION CHECKOUT WHEN ONE IS
# DECLARED, OTHERWISE THE PROJECT ROOT — design.md's "The contract" table,
# cited rather than restated: this guard enforces it, it does not explain
# it. This guard does not otherwise validate the section's shape (that a
# declared `regression checkout`/`regression repo` pair is well formed, or
# that `regression repo` matches the checkout's real `origin`) —
# check-visual-verification.sh already does, and this guard's job is
# narrower: resolve one search root or say it cannot.
#
# THE FILTER READS THE WHOLE PATH, WITH NO SPECIAL-CASED EXCLUSION OF
# ANYTHING — a `screenshots` root wide enough to sweep in a `node_modules`
# tree still filters correctly, because "some path segment starts with the
# spec basename" is exactly as narrow as the stage's own step 9 says it is.
# Pinned rather than assumed: TestResolveVisualScreenshots
# (stats/internal/guard/resolve_visual_screenshots_test.go) plants a decoy
# PNG inside `node_modules` whose name deliberately DOES start with the spec basename
# and confirms it is found — the filter is not a `node_modules` carve-out
# that could hide a real match sitting there too (a project genuinely
# emitting PNGs into a vendored path is not this guard's problem to solve).
#
# THE MATCH IS ANCHORED AT A PATH-SEGMENT BOUNDARY, not a bare substring
# test — task 16's own panel-demonstrated defect. `visual-baseline.spec.ts`
# CONTAINS `baseline.spec.ts` as a substring, so a naive `case "$png" in
# *"$SPEC_BASENAME"*)` filter resolving `baseline.spec.ts` also returned
# `visual-baseline.spec.ts-snapshots/`'s PNGs — a different spec's own
# baseline, sitting under a directory that merely happens to end with the
# queried name. Splitting the path on `/` and testing whether any SEGMENT
# starts with `$SPEC_BASENAME` closes that: `visual-baseline.spec.ts-…`
# starts with `visual-`, not `baseline.spec.ts`, so it is correctly
# rejected, while `baseline.spec.ts-snapshots` and `baseline.spec.ts.png`
# (the node_modules decoy above) both still match, because each IS the
# start of its own segment. Only the START is anchored — the string after
# the spec basename within a segment is unconstrained on purpose, which is
# what lets both the real `-snapshots` suffix and a literal `.png` extension
# still match.
#
# THE INPUT IS ATTACKER-INFLUENCED, exactly as check-visual-verification.sh's
# own header states: `.flow/project.md` is tracked and editable in any pull
# request. Every interpolated cell that reaches an exit-2 message passes
# through sanitizeDisplay first (stats/internal/guard/visualsection.go), for
# the identical forged-verdict reason that function's comment documents. The PNG paths this guard
# prints on success are filesystem paths this guard discovered by walking a
# real directory, never text lifted out of the config file, so they carry no
# such risk and are printed unescaped.
#
# THE HEADING SCAN AND TABLE PARSER ARE check-visual-verification.sh's OWN
# CONVENTIONS, reused rather than reinvented — the Go twins in
# stats/internal/guard/visualsection.go that every visual guard shares.
#
# A LEADING UTF-8 BOM IS STRIPPED BEFORE EITHER HEADING READ OR THE TABLE
# PARSE, via stripBOM (stats/internal/guard/visualsection.go) — see its
# comment for why a BOM would otherwise make this guard read a present
# `## visual verification` heading as absent and exit 2.
#
# The body is Go: stats/internal/guard/resolvevisualscreenshots.go. Matches
# print in byte order, whatever the caller's locale.
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "resolve-visual-screenshots: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec resolve-visual-screenshots 2 "resolve-visual-screenshots:" "$@"
