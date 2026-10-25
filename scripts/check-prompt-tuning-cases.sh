#!/usr/bin/env bash
# check-prompt-tuning-cases.sh [dir] — is every prompt-tuning case folder
# under docs/prompt-tuning/ (or dir) complete and generic?
#
# Per case folder: every file present (case.json, finding.md, bad.diff,
# good.diff, task.md, index.html); case.json parses with every field set;
# type, reviewer_kind, scope and the target the scope maps to valid; before
# verbatim in the target at target_sha (or empty, for a scoped case creating
# its rule file); after naming no source change, Jira key, kan-<n> key or
# diff path; index.html byte-equal to `flow prompt-tuning render`. LEDGER.md,
# when present, holds only well-shaped entries. An absent root is clean.
#
# Exit codes: 0 clean, silent; 1 one `<path>: <reason>` line per violation;
# 2 cannot answer (no git, root outside a git checkout, usage).
#
# Go, in flow-guard (stats/internal/guard/prompttuningcases.go). flow-guard is
# built from this checkout, never taken from PATH: scripts/lib/flow-guard.sh
# derives it, and exits 2 with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-prompt-tuning-cases: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(flow_guard_root)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-prompt-tuning-cases 2 "check-prompt-tuning-cases:" "$@"
