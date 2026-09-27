#!/usr/bin/env bash
# check-model-keys.sh — validate `.flow/project.md`'s `## self review model`
# key against the store's own `ValidModels` set.
#
# Usage: check-model-keys.sh [<project root> ...]
#
# With no arguments it checks the repository this script ships in (mirrors
# check-workspace-isolation.sh's own convention). Each argument is a PROJECT
# ROOT — the directory holding `.flow/project.md` — never the file itself.
#
# **Project configuration** (`skills/flow-contracts/project-configuration.md`)
# is canonical: both keys are optional, and each body — leading/trailing
# whitespace trimmed, nothing else normalized — must match exactly one
# `ValidModels` member and nothing else. A resolver reading a mismatched body
# reports it by name and falls back as if the key were absent; THIS guard is
# stricter on purpose — it exists so that fallback never has to happen in the
# first place, so a mismatch here is a hard failure, not a silent drop.
#
# The valid-model set is asked, not regexed: `flow settings models` prints
# the harness's compiled-in `ValidModels` set (the same store.ValidModels
# the daemon serves), one name per line, and the checkout's own
# settings.go is cross-checked against it. When the two agree, the CLI set
# governs. When they drift — an installed binary predating the checkout —
# the source set governs the verdict, so a stale install cannot fail a
# value the checkout declares, and the drift is announced with each
# side's unique members. The map-literal parse is also the offline
# fallback when the CLI cannot answer at all.
#
# A key that is absent is not a violation — absence is a supported, valid
# state per project-configuration.md's own "Optional" rows.
#
# Exit 0 when every present key is valid, 1 when any is invalid, 2 when it
# cannot answer at all (no usable `flow settings models` answer AND a
# settings.go that is missing/unreadable or yields no members, or a project
# root that is not a readable directory).
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-model-keys: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-model-keys 2 "check-model-keys:" "$@"
