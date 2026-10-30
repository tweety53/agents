#!/usr/bin/env bash
# check-verify-report.sh — answer whether a flow.visual-verify verifier's
# report is complete, so the parent never accepts a partial one by reading
# its prose (KAN-870: two verifiers ran the tests and the capture, listed
# steps 9-11 under "Not done", and both reports were accepted).
#
# Usage: check-verify-report.sh <report file> <motions named> [--design <design.md>]
#
# Given --design, the report, or the `visual-verification.md` beside
# <design.md>, reading `no mockup source` while <design.md> cites a
# backticked `.html`, `.htm` or `.css` path that exists relative to the
# worktree (three directories above <report file>) leaves step 10 undone
# (kan-1017: every verifier left 7 of 11 matrix columns n/a). Step 10
# reported n/a while <design.md> cites a backticked `.html` or `.htm` path
# that exists is likewise undone, unless a wire contract names no view: an
# approved prototype is rendered as each capture's frame (kan-1038). An
# unreadable <design.md> cannot be answered (exit 2).
#
# <report file> is the relay contract's
# <abs-worktree>/.superpowers/sdd/verify-report-<key>.md
# (skills/flow/visual-verify.md); <motions named> is the count of motions the
# dispatch prompt named, 0 for `motions: none`.
#
# Every line below is read from the report's `## Report` heading to the next
# `## ` heading alone: the `## Verdict` block above it is the handback, never
# the record (decision verdict-only-handback). The verifier runs no `verify`
# (decision verifier-runs-no-verify), so step 7 is no required step and a
# `7` item on the `- steps:` line is ignored.
#
# The report's `- steps:` line carries one `<step> <status>` item per required
# step, ` | `-separated, for steps 4, 8, 9, 10, 11 and `motion`
# (skills/flow/visual-verify-verifier.md). A status is admissible when it is
#   done
#   n/a — <reason>       steps 4 and 10 only, and `motion` when <motions named> is 0;
#                        steps 8, 9 and 10 too when a `- wire contract:` line
#                        names an endpoint — the dispatch named no view, so
#                        nothing is captured (kan-1022's self-review)
#   blocked — <reason>   the reason citing a non-zero `exit <n>`; step 10 only
#                        when step 8 is blocked too or the reason cites
#                        `compose-mockup-frames exit <n>` (KAN-1000)
# and `motion done` with <motions named> above 0 also needs a
# `- motion: <n>/<n>` line, <n> that count — a strip recorded beyond the
# named motions goes on a sub-bullet, never into <n>. Every other status — absent,
# `not done`, an n/a on a required step, a block citing no failing exit — is
# an undone step. A `- fingerprint: not declared` line with no ` — <reason>`
# is undone too: step 6 reports its exit 2 with the guard's reason, which the
# prompt carries (kan-1010's self-review).
#
# A `## Report` line attributing a departure as `pre-existing` or `present at
# the merge base` that names neither `merge-base baseline: <sha>:<path>` nor
# `merge-base capture: <path>` leaves step 10 undone (kan-1015: a re-baseline
# over a defect attributed from the code path alone).
#
# With step 10 `done`, every frame id a `- <frame id>: <composite> diff=`
# line names needs its own `- <frame id> matrix:` line; each that has none is
# undone as `<frame id> has no matrix: line — step 10 is done only with every
# frame's matrix` (kan-1016: one matrix line covered nine frames).
# Each such frame needs its `- <frame id> bands:` and `- <frame id> seams:`
# lines too, and its `- <frame id> over-tolerance: <n> — #<row>: <cause>; …`
# line (decisions over-tolerance-line, by-hand-check-automated): <n> must
# equal the `over` count the frame's pairs carry in this dispatch's manifest,
# `visual-manifest-<key>.json` beside <report file>, <key> the report's own
# less any `-rest` or `-rest-<n>` segment — never the committed matrix JSON —
# and every row on that list must be named with one cause from step 10's
# closed list: `departure`, `annotation`, `transcription`, `text width`,
# `decision <ID>` with `**ID:** <ID>` in <design.md>, `present at the merge
# base` beside `merge-base baseline: ` or `merge-base capture: ` evidence,
# or `mockup source` on a row whose `keys` name `unresolved`. A missing or
# unreadable manifest, a frame in none of its pairs, a missing line, a wrong
# count, an unnamed row or a cause off the list is undone.
#
# Every `sweeps:` line's ` | `-separated items each read `<sweep> done…` or
# `<sweep> n/a — <why>`; any other leaves step 10 undone.
#
# Given --design, every matrix table in the `visual-verification.md` beside
# <design.md> — a table whose header's first cell opens `element` — is read
# cell by cell (decision matrix-kind-opens-with-class). The row's class is its
# `kind` cell's opening; a cell reading `n/a…` or `—` stands only in a column
# that does not apply to the class — `border`, `fill`, `radius`, `shadow` for
# a `text run`; none for a `control`, `row` or `card`; `content`,
# `alignment`, `border`, `radius`, `shadow` for a `rule`, `divider` or
# `fill band` — or when it reads `no mockup source`; an unknown class admits
# only `no mockup source`. An empty cell, or any other n/a cell (kan-938: a
# control's `border` left `n/a — measure failed`), leaves step 10 undone,
# naming `visual-verification.md:<line>`.
#
# Given a `## Report` line `- spec: <path>`, the PNGs
# `resolve-visual-screenshots.sh <worktree> <spec basename>` prints, the
# worktree three directories above <report file>, each need a
# `- <view>: <absolute PNG path>` line, and each such view that no
# `- <view>: <composite> diff=` line names needs a
# `per view, no frame composed: <view> sweeps:` line; each gap is undone. No
# `- spec:` line, a wire contract naming an endpoint, or a resolver exit other
# than 0 skips the check (group-allowed-pairs-roles self-review: two round-1
# reports named only the prototype-framed captures).
#
# Prints ONE verdict line to stdout:
#   VERIFY-REPORT-COMPLETE: <report file>
#   VERIFY-REPORT-INCOMPLETE: <report file> — <each undone step and its status, `; `-separated>
#
# Exit 0 complete, 1 incomplete, 2 cannot answer (a usage error, a count
# that is not a non-negative integer, a report file it
# cannot read).
#
# Go, in flow-guard (stats/internal/guard/verifyreport.go).
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-verify-report: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-verify-report 2 "check-verify-report:" "$@"
