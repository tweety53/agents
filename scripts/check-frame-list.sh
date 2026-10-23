#!/usr/bin/env bash
# check-frame-list.sh — reconcile the mockup frames design.md declares against
# the capture specs' `.mockups` sidecars and a flow.visual-verify verifier's
# report, so a declared frame no capture reached is named rather than lost
# (KAN-549: design.md declared 22 frames, two rounds reported 15/17 composed,
# and nothing compared the two).
#
# Usage: check-frame-list.sh <design.md> <mockups dir> <report file>
#
# <design.md> is the change's <changeRoot>/design.md; <mockups dir> the
# resolved `mockups` root (skills/flow-contracts/project-configuration-visual.md);
# <report file> the verifier's
# <abs-worktree>/.superpowers/sdd/verify-report-<key>.md
# (skills/flow/visual-verify.md).
#
# THE DECLARED LIST is derived, never written down twice: every
# `<frame id>.png` directly inside <mockups dir> whose <frame id> design.md
# names as a whole token — bounded on each side by the start or end of a line
# or by a character that is not a letter, digit, `_` or `-`. A range
# `<prefix><a> … <prefix><b>` (joined by `…`, `–` or `...`, each end optionally
# a path ending in `/`, a `.png` suffix and backticks) names every
# `<prefix><n>` from <a> to <b>; ends with different prefixes name only
# themselves. A design line opening with `**Considered:**` names no frame:
# its alternatives were rejected.
#
# THE SIDECARS are `<spec>.mockups` beside every spec the report's
# `- spec: <path>` lines name; a missing sidecar contributes no frame. A
# sidecar line is `<screenshot name> <frame id>`, `#` comments and blank lines
# ignored.
#
# A GAP is, for a declared frame:
#   no sidecar line: <frame id>   no read sidecar maps a screenshot to it
#   no report line: <frame id>    no `- <frame id>:` or `- <frame id> <…>:`
#                                 line — the id may follow other ids joined
#                                 by ` / `, `- <id> / <frame id> <…>:` — and
#                                 the `- frames:` line does not name it
# and, for the report:
#   no frames line                the report carries no `- frames: <n>/<m>`
#                                 line while design.md declares a frame
#   frames <n>/<m> below <k> declared   <m> is smaller than the declared count
#
# Prints ONE verdict line to stdout:
#   FRAME-LIST-COMPLETE: <report file> — <k> declared
#   FRAME-LIST-GAP: <report file> — <each gap, `; `-separated>
#
# Exit 0 no gap, 1 at least one gap, 2 cannot answer (a usage error, an
# unreadable design.md or report, a mockups dir that is not a directory).
#
# Go, in flow-guard (stats/internal/guard/framelist.go).
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-frame-list: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-frame-list 2 "check-frame-list:" "$@"
