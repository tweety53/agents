#!/usr/bin/env bash
# check-mutation-reproducer-pin.sh — pin the mutation-reproducer marker
# literal and its line window mechanically across the two files that state
# them (KAN-624, panel F5 of kan-568's deferred self-review).
#
# Why this exists: the exact literal `# mutation-reproducer` and the
# 10-line window are each stated in skills/flow/review-panel.md (the
# reproducer-convention paragraph and the mutation-testing brief),
# scripts/run-reproducer.sh (the header comment) and the runner's Go port
# stats/internal/guard/runreproducer.go, with no mechanical pin
# keeping them together — check-dispatch-paragraphs.sh covers the required
# dispatch paragraphs but not this pair, so the two statements could drift
# silently, one prose edit or one code edit at a time. This guard is the
# pin.
#
# The source of truth is the runner's own canonical declaration — the code
# that decides the convention at run time. Since KAN-760 the runner is Go
# (scripts/run-reproducer.sh is a shim exec'ing flow-guard), so the
# declaration is two lines of stats/internal/guard/runreproducer.go: the
# marker constant `rrMarker = "<MARKER>"` and declaresMutation's
# `range headLines(path, <N>)`. The guard extracts the marker and the
# window from those lines and then requires the stated sites to agree, two
# ways:
#
#   - PRESENCE — skills/flow/review-panel.md carries the marker backticked
#     and the window stated as "first <N> lines"; scripts/run-reproducer.sh
#     carries the marker backticked (its header comment) and the same
#     window statement.
#   - NEAR-MISS ABSENCE — presence alone cannot see a single site drift
#     while another copy of the same fact still reads correctly, which is
#     the one-edit-at-a-time drift this guard exists to catch (KAN-624
#     F1). So every "first <M> lines" statement and every backticked
#     `# mutation-...` span in any of the three files must carry exactly the
#     canonical value, on a single line or wrapped across two adjacent
#     lines (the mutation-testing brief's own window statement is
#     wrapped), and each offending occurrence is reported at its own
#     line.
#
# A change to any one site without the others is a drift: exit 1, named
# per file and line, until all sites move together. Reshaping a canonical
# declaration line itself — so that no line of that shape can be found, or
# more than one can, or its marker or window is not extractable from it —
# is not a drift this guard interprets: it cannot tell what the new
# reading would be, so it exits 2 and the pin waits to be re-pointed
# deliberately, the same not-a-verdict course check-plan-shape.sh's exit 2
# takes.
#
# Argument-free and self-scoped exactly like check-dispatch-paragraphs.sh:
# the scan root is this repository's own root, this shim's parent, exported
# as FLOW_GUARD_REPO_ROOT. CHECK_MUTATION_REPRODUCER_PIN_ROOT is an explicit,
# opt-in override honored only when set (mirrors
# CHECK_DISPATCH_PARAGRAPHS_ROOT), so the Go test
# (stats/internal/guard/check_mutation_reproducer_pin_test.go) can point this
# guard at a sandboxed fixture tree without touching this repository — never
# set it for a normal invocation. The mechanism is
# stats/internal/guard/mutationreproducerpin.go.
#
# Exit codes:
#   0  clean — the marker and the window agree across all three files
#   1  drift — a stated site disagrees with the canonical declaration
#   2  cannot answer — a pinned file missing, a symlink, not a regular
#      file or unreadable; a canonical declaration line absent or
#      ambiguous; the marker or window not extractable; or
#      CHECK_MUTATION_REPRODUCER_PIN_ROOT set but empty
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-mutation-reproducer-pin: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-mutation-reproducer-pin 2 "check-mutation-reproducer-pin:" "$@"
