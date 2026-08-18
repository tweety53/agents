# scripts/lib/resolve-file.sh — resolve_file, defined once.
#
# Sourced by scripts/check-guard-symlinks.sh, scripts/plan-dispatch-bundles.sh
# and scripts/check-workspace-isolation.sh, which used to carry three
# near-identical copies of this function (plus two more, left alone below).
# The copies had already drifted before this file existed:
# check-guard-symlinks.sh's copy put `--` before every `readlink`, `dirname`
# and `cd` argument; plan-dispatch-bundles.sh's and check-workspace-isolation.sh's
# did not. KAN-73's own review panel caught the drift (see that change's
# final-review-panel.md, F1) — the same shape of defect
# scripts/lib/panel-record.sh's header records for the two guards it was
# extracted from. One definition, sourced by every guard that can safely
# reach it, is what stops that drift from happening a second time.
#
# "SAFELY REACH IT" IS THE OPERATIVE PHRASE, and not every guard carrying a
# copy of this function qualifies. scripts/preserve-session-records.sh and
# scripts/gather-self-review-context.sh keep their own inline copies —
# untouched by this file — because unlike the three guards above, neither
# ships through the skills/*/scripts/ symlink farm KAN-73 built. A guard
# reached only by hand-copying a single file into an unrelated project's own
# tooling cannot assume a sibling `lib/` travels with it; a guard shipped
# through the farm can, because the farm already symlinks `lib` as a
# directory beside every guard that needs it (KAN-73's design.md, "The
# guard-to-skill map") — the identical argument KAN-153's F7 accepted for
# check-unfinished-work.sh sourcing scripts/lib/panel-record.sh, on the
# evidence that setup.sh distributes skills, not scripts/, so a guard and the
# library it sources always travel together in this repository. Follow-up
# question, out of this change's scope: whether preserve-session-records.sh's
# and gather-self-review-context.sh's own copies should ever join this file.
#
# Not meant to be executed directly — a caller sources it and calls
# resolve_file; it sets no `set -euo pipefail` of its own and relies on the
# sourcing script's.
#
# `--` BEFORE EVERY readlink, dirname AND cd ARGUMENT — the hardened form,
# carried here rather than the un-hardened one four of the five original
# copies used, so a path beginning with `-` is never read as an option. This
# is the same discipline scripts/lib/panel-record.sh's header states for
# every guard in this repository.
#
# resolve_file <path> -> prints the path's resolved PHYSICAL location on
# stdout, or returns 1 if it cannot. macOS /bin/sh has no `readlink -f`, so
# this walks symlinks by hand, capped at 40 hops against a cycle.
resolve_file() {
  local p="$1" hops=0 target dir
  while [ -L "$p" ]; do
    hops=$((hops + 1))
    [ "$hops" -le 40 ] || return 1
    target="$(readlink -- "$p")" || return 1
    case "$target" in
      /*) p="$target" ;;
      *) p="$(dirname -- "$p")/$target" ;;
    esac
  done
  dir="$(cd -P -- "$(dirname -- "$p")" 2>/dev/null && pwd -P)" || return 1
  printf '%s\n' "$dir/${p##*/}"
}
