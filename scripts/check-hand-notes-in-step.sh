#!/usr/bin/env bash
# check-hand-notes-in-step.sh — fail when the hand-maintained sections of the
# installed harness files have drifted apart.
#
# Why this exists. KAN-808: a flake-bisection pointer was appended by hand to
# BOTH `~/.claude/CLAUDE.md` and `~/.zcode/AGENTS.md` as identical appends
# outside their generated managed blocks. The managed block in each file is
# re-rendered from one source on every install, so it cannot drift; the hand
# appends around it are written twice, once per harness file, and nothing
# asserted the copies stayed identical. Every such append is a fresh
# opportunity for a one-sided edit to diverge silently — the same drift class
# kan-630's panel caught (three stale copies of one lint command across
# instruction files). This guard is that assertion.
#
# Argument-free and self-scoped, like check-installed-rules.sh: the file set
# is this checkout's own setup.sh `local managed_files=(...)` declaration —
# parsed live (kan-585), by the same parser and with the same refusals — and
# the install root is the invoking user's `$HOME`.
# CHECK_HAND_NOTES_HOME and CHECK_HAND_NOTES_SETUP_SH are explicit, opt-in
# overrides honored only when set, so the Go tests
# (stats/internal/guard/check_hand_notes_test.go) can point the guard at
# sandboxed fixtures — never set them for a normal invocation.
#
# NOT INSTALLED IS NOT STALE, and NEITHER IS A SOLO INSTALL. A checkout with
# no global install — CI, a fresh clone, a container — has nothing to be out
# of step with: HAND-NOTES-NONE, exit 0. Exactly one harness file present has
# no pair to compare: HAND-NOTES-SINGLE, exit 0 — absent is not stale, the
# same course check-installed-rules takes for a harness file that does not
# exist.
#
# COMPARISON. Everything outside the managed block of each existing file —
# the block and its delimiters stripped — must be byte-identical across all
# of them, final newline included. The installer rewrites `~/.claude/`
# pointers to `~/.zcode/` ones inside the block, which is why block content
# is never read as drift. A file ending inside an unterminated managed block
# is a violation of its own. Verdicts:
#
#   HAND-NOTES-NONE:    <home> — no global install found, nothing to check
#   HAND-NOTES-SINGLE:  <home> — one harness file present, no pair to compare
#   HAND-NOTES-OK:      <home> — N harness file(s), hand-maintained sections in step
#   HAND-NOTES-DRIFT:   <home> — N violation(s); make every harness file's hand-maintained sections identical
#
# The remedy for every violation is the same, which is why the verdict names
# it once: mirror the diverging hand section across the files it belongs to.
#
# The guard runs as the Go port in stats/internal/guard/handnotes.go. The
# binary does not live in this checkout, so this shim exports
# FLOW_GUARD_REPO_ROOT — this script's parent directory — as the checkout
# whose setup.sh is read. flow-guard is built from this checkout, never taken
# from PATH: scripts/lib/flow-guard.sh derives it, and exits 1 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-hand-notes-in-step: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 1
}
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-hand-notes-in-step 1 "check-hand-notes-in-step:" "$@"
