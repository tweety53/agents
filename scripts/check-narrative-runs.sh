#!/usr/bin/env bash
# check-narrative-runs.sh <worktree> <change> <narrative.md>
#
# Compares a change's narrative with the runs the flow store marked for it
# (decision narrative-runs-from-stage-marks). The fix runs are the sessions
# `flow stage sessions -stage flow.document-fix` counts, the integrate runs
# those of `-stage flow.preserve-sessions`; the narrative must hold at least
# as many `## <date> — fix run` and `## <date> — integrate run` sections. A
# narrative that does not exist holds zero sections.
#
# Exit 0: NARRATIVE-RUNS-OK — every marked run has its section.
# Exit 1: NARRATIVE-RUNS-MISSING — a fix or integrate run lacks its section.
# Exit 2: usage error, a store read that did not print one integer, or a
#         narrative that exists and cannot be read.
set -euo pipefail

[ "$#" -eq 3 ] || { echo "usage: check-narrative-runs.sh <worktree> <change> <narrative.md>" >&2; exit 2; }
wt="$1" change="$2" nf="$3"

sessions() {
  local out
  out="$(flow stage sessions -C "$wt" -stage "$1" "$change")" || { echo "check-narrative-runs: flow stage sessions -stage $1 failed" >&2; exit 2; }
  [[ "$out" =~ ^[0-9]+$ ]] || { echo "check-narrative-runs: flow stage sessions -stage $1 printed '$out', not one integer" >&2; exit 2; }
  printf '%s' "$out"
}
f="$(sessions flow.document-fix)"
i="$(sessions flow.preserve-sessions)"

hf=0 hi=0
if [ -e "$nf" ]; then
  [ -r "$nf" ] || { echo "check-narrative-runs: cannot read $nf" >&2; exit 2; }
  hf="$(grep -cE '^## .* — fix run' "$nf" || true)"
  hi="$(grep -cE '^## .* — integrate run' "$nf" || true)"
fi

if [ "$hf" -lt "$f" ] || [ "$hi" -lt "$i" ]; then
  echo "NARRATIVE-RUNS-MISSING: $nf holds $hf '— fix run' section(s) for $f fix run(s) and $hi '— integrate run' section(s) for $i integrate run(s) the store marked"
  exit 1
fi
echo "NARRATIVE-RUNS-OK: $f fix run(s), $i integrate run(s), each with its section"
