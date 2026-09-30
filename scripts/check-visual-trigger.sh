#!/usr/bin/env bash
# check-visual-trigger.sh — decide whether a diff touched a project's
# declared `## visual verification` UI paths, in a guard instead of prose.
#
# Usage: check-visual-trigger.sh <project root>   (changed paths on stdin,
#                                                    one per line)
#
# Replaces flow.visual-verify's step 2 (skills/flow/verify-and-handoff.md):
# "run `git diff --name-only` and test every printed path as a glob against
# the comma-separated `ui paths`." Two runs reading the same globs
# differently was the defect this guard exists to make impossible — the
# glob semantics are the whole risk, per tasks.md task 9.
#
# THREE EXIT CODES, AND "NOT CONFIGURED" IS NEVER "NO MATCH":
#   0  at least one stdin path matched at least one `ui paths` glob
#   1  the section resolved cleanly and `ui paths` is non-empty, but no
#      stdin path (including an empty stdin — no diff at all) matched
#   2  cannot answer: the project root is not a directory, `.flow/project.md`
#      is missing, not a regular file, or unreadable, the heading scan
#      itself failed to look, the section is declared twice (ambiguous —
#      neither is read), or `ui paths` is absent or empty.
# "Not configured" (2) and "configured and unmatched" (1) are deliberately
# different answers — flow.visual-verify's step 1 and step 2 print two
# different messages and skip the stage for two different reasons.
# Every exit 2 ends its stderr with one cause token line, so a caller tells
# the two exit-2 answers apart without reading the prose above it:
#   VISUAL-TRIGGER-NOT-CONFIGURED: <root> — no visual verification section
#      `.flow/project.md` is missing, or declares no `## visual verification`
#   VISUAL-TRIGGER-CANNOT-ANSWER: <root> — <short cause>
#      every other exit 2 (a usage error prints `(no root)` as <root>)
#
# GLOB SEMANTICS, pinned because they already caused disagreement once:
#   - `**` matches any run of characters INCLUDING `/`, so it spans
#     multiple directory levels, not one segment.
#   - a bare `*` matches any run of characters EXCEPT `/` — one path
#     segment. (Every glob this repository declares today uses `**`; a
#     bare `*` is supported for completeness, not because a declaration
#     uses one.)
#   - `?` matches exactly one byte except `/`, and `*`/`**` count bytes too: the
#     port matches byte-wise whatever the caller's locale (c-locale-table-semantics).
#   - a leading `./` is stripped from BOTH the declared glob and the
#     candidate path before matching — a habit either side might write.
#   - a leading `/` on a DECLARED glob is stripped too: `.flow/project.md`
#     paths are inherently repository-root-relative, so `/stats/web/**` and
#     `stats/web/**` mean the same thing.
#   - an ABSOLUTE candidate path is never stripped and so never matches a
#     relative glob. `git diff --name-only` never emits one; a caller that
#     hands this guard one gets a defined refusal, not a silent grant.
#   - splitting `ui paths` on its separator comma never touches a SPACE
#     inside one glob — only leading/trailing whitespace around each
#     comma-separated element is trimmed.
#
# THE INPUT IS ATTACKER-INFLUENCED, exactly as check-visual-verification.sh's
# own header states: `.flow/project.md` is tracked and editable in any pull
# request. This guard never executes anything it reads — `ui paths` values
# are matched as glob patterns, never interpolated into a shell command —
# and every interpolated cell that reaches an exit-2 message or a MATCH line
# passes through sanitizeDisplay (visualsection.go, the Go twin of
# scripts/lib/sanitize-display.sh) first — see its comment for the
# forged-verdict reason.
#
# THE HEADING SCAN AND TABLE PARSER ARE check-visual-verification.sh's OWN
# CONVENTIONS, reused rather than reinvented: the same `## visual
# verification` heading regex, the same "a heading at or above this
# section's own level ends it, a deeper one is prose" rule, and the same
# `|`-row split with `\|`-escape awareness. They are the Go twins in
# stats/internal/guard/visualsection.go (vvHeadingCount, vvSectionLines,
# vtSplitCells, vtTrimCell, vtFoldCell), which carry the reasoning of
# scripts/lib/visual-table-cells.awk and its parity test.
#
# A LEADING UTF-8 BOM IS STRIPPED BEFORE EITHER HEADING READ OR THE TABLE
# PARSE, via the same file's stripBOM — see its comment for why a BOM would
# otherwise make this guard read a present `## visual verification` heading
# as absent and exit 2 ("not configured"), which is worse here than in the
# other two guards: this is the guard that decides whether flow.visual-verify
# runs at all, so that misread is a silent skip of the whole stage, not
# merely a misparse.
#
# `ui paths`' OWN PER-ELEMENT TRIM is trimGlobElement in the same file,
# shared with check-visual-verification — see its comment for the
# split-then-strip parse and the single-pass strip it performs.
#
# The guard is Go: stats/internal/guard/visualtrigger.go.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-visual-trigger: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-visual-trigger 2 "check-visual-trigger:" "$@"
