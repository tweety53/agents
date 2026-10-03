#!/usr/bin/env bash
# prepare-workspace.sh — validate a worktree's `## workspace isolation`
# declaration and export the workspace variables it derives.
#
# Usage: prepare-workspace.sh <worktree>
#
# This is exactly what `/flow`'s implement phase used to do by hand, in prose:
# run check-workspace-isolation.sh against the worktree first — fail loudly
# and stop if it fails — then derive and export the workspace variables the
# project's `## workspace isolation` section declares, resolved against the
# workspace id. Every rule this script applies is stated once, and canonical,
# elsewhere: the id derivation and what it derives under **The workspace id**
# and **What the id derives** (skills/flow-contracts/workspace-isolation.md),
# and the `## workspace isolation` row shapes — the four `In a workspace`
# forms, the `<id>`, `<id_underscored>` and `<value:VARIABLE>` tokens — under
# **Project configuration — workspace isolation** (skills/flow-contracts/project-configuration-isolation.md).
# This script re-derives none of that reasoning; it applies the rule.
#
# Prints one `KEY=value` line per exported variable to stdout, one per
# declared row that carries a value (see the cache-index exception below), so
# a caller reads exactly what was exported without re-deriving anything
# itself.
#
# THE WORKSPACE ID NEEDS THE CHANGE NAME, AND THIS SCRIPT TAKES ONLY A
# WORKTREE. The id is derived from the change name and from nothing else, per
# **The workspace id**, so this script reads the name from the worktree's own
# branch: every apply worktree is created on `spectre/<name>`, per
# skills/flow/implement.md, and that branch is the one place the name is
# already recorded where a script can read it without being handed it
# separately. A worktree not on such a branch — detached HEAD, or checked out
# on something else — cannot be resolved, and this script refuses rather than
# guessing at a name.
#
# THE CACHE INDEX IS THE ONE ROW THIS SCRIPT DOES NOT EXPORT. Per **The cache
# index** (skills/flow-contracts/workspace-isolation.md), a `cache index`
# row is claimed by probing the project's own cache, not derived from the id —
# and the registry in skills/flow-contracts/artifacts-registry.md names `/flow`,
# by probing, as what claims it "when it exports the workspace's variables".
# Probing means holding a client for whatever cache technology the project
# actually runs, which is exactly the kind of project-specific knowledge a
# script shipped into every project alike cannot carry. So a declared
# `cache index` row is reported by name on stderr — never silently dropped —
# and exports nothing; claiming it stays the operator's or the agent's own
# step, done against the real service, same as it was before this script
# existed.
#
# Exit 0 when the section is absent (no-op: nothing exported, nothing
# printed) or when isolation resolved and exported cleanly. When
# check-workspace-isolation.sh itself fails, this script's exit code is the
# guard's own — 1 (a dropped row) or 2 (the guard could not answer at all) —
# and the guard's stdout is relayed verbatim before exiting; its stderr is
# inherited, not captured, so it is already in front of the operator by the
# time this script's own exit code lands. Exit 2 additionally covers a
# worktree this script cannot derive a change name from, and one whose
# check-workspace-isolation.sh cannot be found.
#
# A project declaring no `## workspace isolation` section is the ordinary
# case for a repository with no runnable application, and is never reported
# as a misconfiguration, per **The empty id**
# (skills/flow-contracts/workspace-isolation.md).
#
# Ported to Go (KAN-873): the body's reasoning for every step is in
# stats/internal/guard/prepareworkspace.go, beside the code. The guard still
# execs check-workspace-isolation.sh from beside this script, as the bash
# did: FLOW_GUARD_WORKSPACE_ISOLATION is exported as its path, so a missing or
# non-executable sibling keeps its exit 2.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "prepare-workspace: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_WORKSPACE_ISOLATION="$SCRIPT_DIR/check-workspace-isolation.sh"
export FLOW_GUARD_WORKSPACE_ISOLATION
flow_guard_exec prepare-workspace 2 "prepare-workspace:" "$@"
