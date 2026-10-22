#!/usr/bin/env bash
# rebuild-full-app-suite-zip.sh — rebuild <screenshots root>/full-app-suite.zip
# from exactly the PNGs full-app-suite.spec.ts produced. The rebuild lived only
# as inline bash in step 8 of skills/flow/visual-verify-verifier.md, so an
# in-run recapture never ran it and kan-1003's zip went stale twice; the
# verifier, the in-run fix loop (skills/flow/verify-fix-loop.md) and the base
# sync's recapture (skills/flow/sync-onto-base.md) all call this one script.
#
# Usage: rebuild-full-app-suite-zip.sh <worktree> <screenshots root>
#
# The PNG list is `resolve-visual-screenshots.sh <worktree> full-app-suite.spec.ts`.
# The old zip is removed first, so the archive never keeps a PNG the suite no
# longer produces. Every PNG's mtime is pinned to 1980-01-01 before zipping,
# since the zip stores it: a rebuild over unchanged images is byte-identical
# and commits nothing.
#
# Exit 0 the zip is written; 1 the resolver matched no PNG (nothing written);
# 2 cannot answer — a usage error, the resolver's own exit 2, or a failing zip.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this script needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
if [ "$#" -ne 2 ]; then
  echo "usage: rebuild-full-app-suite-zip.sh <worktree> <screenshots root>" >&2
  exit 2
fi
zip_path="$2/full-app-suite.zip"
rc=0
list=$("$SCRIPT_DIR/resolve-visual-screenshots.sh" "$1" full-app-suite.spec.ts) || rc=$?
# A resolver that cannot answer (exit 2) leaves the old zip alone: only a resolved list replaces it.
[ "$rc" -eq 2 ] && exit 2
rm -f "$zip_path"
[ "$rc" -eq 0 ] || exit "$rc"
printf '%s\n' "$list" | tr '\n' '\0' | xargs -0 touch -t 198001010000
printf '%s\n' "$list" | zip -q -X -j -@ "$zip_path" || exit 2
