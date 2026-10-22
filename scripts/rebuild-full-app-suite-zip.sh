#!/usr/bin/env bash
# rebuild-full-app-suite-zip.sh — rebuild <screenshots root>/full-app-suite.zip
# from exactly the PNGs full-app-suite.spec.ts produced.
#
# Usage: rebuild-full-app-suite-zip.sh <worktree> <screenshots root>
#
# Exit 0 the zip is written; 1 the resolver matched no PNG (nothing written);
# 2 cannot answer — a usage error, the resolver's own exit 2, or a failing zip.
#
# The mechanism, and its full contract, is
# stats/internal/guard/rebuildfullappsuitezip.go; its test is
# stats/internal/guard/rebuild_full_app_suite_zip_test.go.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "rebuild-full-app-suite-zip: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
: "$SCRIPT_DIR/resolve-visual-screenshots.sh"
FLOW_GUARD_SCRIPT_DIR="$SCRIPT_DIR"
export FLOW_GUARD_SCRIPT_DIR
flow_guard_exec rebuild-full-app-suite-zip 2 "rebuild-full-app-suite-zip:" "$@"
