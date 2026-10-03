#!/usr/bin/env bash
# check-installed-rules.sh — fail when this repository's always-on rules and
# what `setup.sh global` last installed have drifted apart.
#
# Why this exists. KAN-202 added `rules/commit-scope-is-the-module.mdc` and
# listed it in `rules/agent-baseline.md`. It finished green, archived, and
# merged — and nobody re-ran the installer. `agent-baseline.md` is itself a
# symlink into this repository, so its new table row appeared in every
# dispatched agent's context the moment the change merged, pointing at
# `~/.claude/rules/commit-scope-is-the-module.md`, which no install had ever
# created. The managed block in `~/.claude/CLAUDE.md` did not carry the rule
# either, so no main session was subject to it. A rule that is written,
# reviewed, merged and unreadable is worse than one that was never added: the
# baseline table asserts it exists, and the pointer dangles. It went unnoticed
# for a day, until a commit landed on `develop` carrying exactly the scope the
# unread rule prohibits.
#
# Nothing checked the installed state against the repository, so this does.
#
# Argument-free and self-scoped, like check-references.sh and
# check-guard-symlinks.sh: the rule set is this repository's own `rules/`,
# resolved from this script's own location, and the install root is the
# invoking user's `$HOME`. CHECK_INSTALLED_RULES_HOME is an explicit, opt-in
# override honored only when set, so its tests
# (stats/internal/guard/check_installed_rules_test.go) can point this guard at a sandboxed fixture
# under TMPDIR — never set it for a normal invocation.
# CHECK_INSTALLED_RULES_SETUP_SH is the same convention for the installer
# itself, defaulting to this checkout's setup.sh, so the tests can prove
# rule 4 follows the installer's declared targets without writing the real
# one.
#
# NOT INSTALLED IS NOT STALE. A checkout with no global install — CI, a fresh
# clone, a container — has nothing to be out of date with, and failing there
# would make the guard unrunnable in exactly the places lint runs unattended.
# When `<home>/.claude/rules/` does not exist, this guard reports that it
# found no install and exits 0. A PARTIAL install is a different thing and is
# a failure: the directory existing is what says an install happened, and
# every rule missing from it after that is drift.
#
# A LINKED WORKTREE IS NOT THE INSTALLED CHECKOUT EITHER. `setup.sh global`
# links into whichever checkout it was run from, and its own closing note says
# to install from a stable main checkout because links into a worktree break
# when that worktree is removed. So every rule in an apply worktree resolves to
# the main checkout, not to this tree, and rule 1 would report all of them as
# pointing at the wrong place — failing `## lint` in every apply worktree, which
# is exactly where a `/flow*` change runs its lint. That is a false alarm
# about a correct install, not drift. This guard detects a linked worktree by
# `.git` being a file rather than a directory, reports it, and exits 0.
#
# The skip applies only when CHECK_INSTALLED_RULES_HOME is unset. That override
# means a caller is deliberately pointing the guard at a fixture home, and wants
# the real comparison run against it; short-circuiting there would make the
# tests unable to test anything from inside a worktree, which is where they run.
#
# Four rules, each reported with the path and what to do:
#
#   1. Every always-on rule in `rules/` has `<home>/.claude/rules/<name>.md`,
#      it is a symlink, it resolves, and its target is this repository's own
#      `rules/<name>.mdc`. A link pointing at some other checkout is drift of
#      the worst kind — it resolves, so nothing looks broken, while the text
#      every agent reads comes from a tree nobody is editing.
#   2. `<home>/.claude/rules/` carries no `.md` entry that is NOT a current
#      always-on rule (`agent-baseline.md` excepted, which is not a rule).
#      A rule deleted here, or one that stopped declaring `alwaysApply: true`,
#      leaves a link behind that still reads as installed.
#   3. `agent-baseline.md` is installed and resolves. It is the one file a
#      dispatched agent is told to read; every row in its table is a promise
#      that rule 1 then has to keep.
#   4. Every always-on rule has a `<!-- rule: <name>.mdc -->` marker inside
#      the managed block of each installed harness file, and each block
#      carries no marker for a rule that is no longer always-on. The marker
#      is what `render_managed_block` in setup.sh writes per rule, so it is
#      read here rather than any title or body text, which would re-derive
#      the renderer's formatting and go stale against it. The SET of harness
#      files is served too (kan-585): it is parsed from setup.sh's own
#      `local managed_files=(...)` declaration — the one line install_global
#      already comments as "declared once so the preflight and the install
#      can never scan a different set than they write" — so a harness file
#      the installer gains moves this guard first, and the guard can never
#      lag the installer with a hardcoded pair. That pair had already lagged
#      once: `~/.zcode/AGENTS.md` was added to the installer and not to this
#      guard, leaving a whole harness's block unscanned. A harness file that
#      does not exist is skipped with a
#      note, for the same reason rule 1's whole directory is: absent is not
#      stale. An unreadable setup.sh, or one whose declaration is missing,
#      empty, or carries an element the guard cannot resolve to a
#      `$home_dir/`-relative path, is a refusal: exit 1 with the reason,
#      never a guess.
#
# Prints one violation line per finding, then the verdict:
#   INSTALLED-RULES-OK:      <home> — N always-on rule(s) installed and rendered
#   INSTALLED-RULES-NONE:    <home> — no global install found, nothing to check
#   INSTALLED-RULES-STALE:   <home> — N violation(s); re-run ./setup.sh global
#
# Every violation has the same remedy, which is why the verdict names it once:
# `./setup.sh global` from this checkout.
#
# The guard runs as the Go port in stats/internal/guard/installedrules.go.
# The binary does not live in this checkout, so this shim exports
# FLOW_GUARD_REPO_ROOT — this script's parent directory, derived as the bash
# guard derived its REPO_ROOT — as the checkout whose rules/ and setup.sh are
# read. flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 1 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-installed-rules: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 1
}
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-installed-rules 1 "check-installed-rules:" "$@"
