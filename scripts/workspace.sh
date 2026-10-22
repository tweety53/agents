#!/usr/bin/env bash
# workspace.sh — thin shim over `flow-guard workspace`
# (stats/internal/guard/workspace.go), whose header states the
# `create`/`remove`/`survivors` triple .flow/project.md's
# `## workspace isolation` section names, its arguments and exit codes.
#
# Usage:
#   scripts/workspace.sh create <id>
#   scripts/workspace.sh remove <id> [--force]
#   scripts/workspace.sh survivors <id>
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this command's
# refusal code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "workspace: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec workspace 2 "workspace:" "$@"
