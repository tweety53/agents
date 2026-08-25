#!/usr/bin/env bash
# Assertion harness for gather-self-review-context.sh. Builds a sandboxed git
# repository under TMPDIR per case; never touches the real repository tree.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated
# contract — Requirement: "Context is gathered deterministically, not by the
# model re-reading files" in this repository's archived kan-23-myflow-self-review
# change's delta spec — never against whatever the script happens to print.
# test-check-plan-provenance.sh's header records that suite encoding the
# guard's own defects as its specification more than once, which then made
# each defect look verified; the same mistake is possible here.
#
# NOTE on cwd (F22): the script's trust boundary (TRUSTED_REPO_ROOT) is now
# derived from its own process cwd via `git rev-parse --git-common-dir` (F23
# switched this from `--show-toplevel`, which returns a worktree's own root
# rather than the main repo's root when invoked from inside a worktree), per
# its documented caller's actual invocation shape (skills/myflow-finish/
# SKILL.md step 8 always runs it with cwd at the repo root, passing a
# RELATIVE <archived-change-path>). Every case below that expects a source to
# be found therefore invokes the script with `cd "$REPO"` and a path relative
# to it — matching the real contract — rather than an absolute path built
# from a different cwd, which the new mechanism correctly refuses.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT="$SCRIPT_DIR/gather-self-review-context.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# Every case leaves one sandboxed git repository behind. Removed on exit,
# including on a failed assertion. An indexed array, not a space-separated
# string: sandbox paths come from mktemp under TMPDIR, which may contain
# spaces, and word-splitting a string would leak every sandbox whose path
# split and `rm -rf` the fragments. Same hazard every harness under scripts/
# names in its own header; bash 3.2 has indexed arrays, only associative
# arrays are unavailable.
TREES=()
cleanup() {
  [ "${#TREES[@]}" -eq 0 ] && return 0
  for tree in "${TREES[@]}"; do
    chmod -R u+w "$tree" 2>/dev/null || true
    rm -rf "$tree"
  done
}
trap cleanup EXIT

# The relative <archived-change-path> shape this script's documented caller
# always uses (skills/myflow-finish/SKILL.md step 8).
REL="spectre/changes/archive/2026-01-01-demo"

# new_repo -> sets REPO, ARCHIVED, STATE_DIR for the change named `demo`.
# ARCHIVED is a real directory inside a git repository, since the script
# derives the repository root from its own process cwd — every invocation
# below runs with cwd at $REPO and $REL as the argument (see NOTE above).
# ARCHIVED (absolute) is kept only for building/inspecting fixture paths in
# each case, never passed directly to the script.
new_repo() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-repo.XXXXXX")"
  TREES+=("$REPO")
  (
    cd "$REPO" \
      && git init -q \
      && git config user.email test@example.com \
      && git config user.name test \
      && git commit -q --allow-empty -m "init"
  )
  ARCHIVED="$REPO/spectre/changes/archive/2026-01-01-demo"
  mkdir -p "$ARCHIVED"
  STATE_DIR="$REPO/.state"
  mkdir -p "$STATE_DIR"
}

# add_ledger / add_panel / add_tasks — write one of the three file sources.
add_ledger() {
  mkdir -p "$REPO/docs/superpowers/ledgers"
  printf 'LEDGER-BODY\n' > "$REPO/docs/superpowers/ledgers/2026-01-01-demo.md"
}
add_panel() {
  mkdir -p "$REPO/docs/superpowers/reviews"
  printf 'PANEL-BODY\n' > "$REPO/docs/superpowers/reviews/2026-01-01-demo-panel.md"
}
add_tasks() {
  printf 'TASKS-BODY\n' > "$ARCHIVED/tasks.md"
}

# add_commits — the two finish-run-1 commits plus the archive commit, using
# this repository's own established subject conventions
# (skills/myflow-contracts/pipeline.md -> Git boundaries). The implementation
# and planning commits touch REAL paths (a real-path deletion, this change's
# own review panel, pass 3, finding A) — `--allow-empty` with nothing staged
# is the only way to produce a commit that misses PLAN_SHA's live-pathspec
# filter, which made every case built on this helper exercise nothing but a
# fixture artifact once the subject-only fallback was removed. The archive
# commit stays a plain marker commit: ARCHIVE_SHA is resolved by subject
# alone and never reads a path, so giving it one would only add fixture
# complexity (a `git mv` destination shared with other fixtures' own
# archived-path setup) without covering anything PLAN_SHA/IMPL_SHA don't
# already cover realistically on their own. The archive marker goes in the
# commit BODY, not on the subject line: ARCHIVE_SUBJECT_RE is anchored at
# both ends, so a marker appended to the subject would stop it matching —
# which is the anchoring doing exactly its job.
add_commits() {
  (
    cd "$REPO" \
      && mkdir -p src \
      && printf 'IMPLEMENTATION-COMMIT-BODY\n' > src/demo-feature.txt \
      && git add src/demo-feature.txt \
      && git commit -q -m "feat(demo): IMPLEMENTATION-COMMIT-BODY" \
      && mkdir -p spectre/changes/demo \
      && printf 'PLAN-COMMIT-BODY\n' > spectre/changes/demo/proposal.md \
      && git add spectre/changes/demo/proposal.md \
      && git commit -q -m "chore(demo): plan, test guide and session records PLAN-COMMIT-BODY" \
      && git commit -q --allow-empty -m "chore(spectre): archive demo" -m "ARCHIVE-COMMIT-BODY"
  )
}

# run_it -> sets RC and OUT. Invokes with cwd at $REPO and $REL, matching the
# documented real invocation contract.
run_it() {
  set +e
  OUT="$(cd "$REPO" && "$SCRIPT" "$REL" demo "$STATE_DIR" 2>&1)"
  RC=$?
  set -e
}

# run_with -> sets RC and OUT. Like run_it, but with a caller-supplied
# <archived-change-path> (still invoked with cwd at $REPO) — for cases that
# deliberately pass a non-standard shape (trailing slash, `..`, extra
# nesting).
run_with() {
  local arg="$1"
  set +e
  OUT="$(cd "$REPO" && "$SCRIPT" "$arg" demo "$STATE_DIR" 2>&1)"
  RC=$?
  set -e
}

# ===========================================================================
# SECTION: All sources present
# ===========================================================================

new_repo
add_ledger
add_panel
add_tasks
add_commits
run_it

[ "$RC" -eq 0 ] && pass "all sources present: exits 0" \
  || fail "all sources present: rc=$RC out=$OUT"
case "$OUT" in
  *LEDGER-BODY*) pass "all sources present: ledger content in bundle" ;;
  *) fail "all sources present: ledger content missing: $OUT" ;;
esac
case "$OUT" in
  *PANEL-BODY*) pass "all sources present: panel content in bundle" ;;
  *) fail "all sources present: panel content missing: $OUT" ;;
esac
case "$OUT" in
  *TASKS-BODY*) pass "all sources present: tasks.md content in bundle" ;;
  *) fail "all sources present: tasks.md content missing: $OUT" ;;
esac
case "$OUT" in
  *"ARCHIVE-COMMIT-BODY"*) pass "all sources present: git log content in bundle" ;;
  *) fail "all sources present: git log content missing: $OUT" ;;
esac
case "$OUT" in
  *"IMPLEMENTATION-COMMIT-BODY"*) pass "all sources present: IMPL_SHA content in bundle" ;;
  *) fail "all sources present: IMPL_SHA content missing: $OUT" ;;
esac
case "$OUT" in
  *"PLAN-COMMIT-BODY"*) pass "all sources present: PLAN_SHA content in bundle" ;;
  *) fail "all sources present: PLAN_SHA content missing: $OUT" ;;
esac

# ===========================================================================
# SECTION: A missing source (1-3) is skipped, never fatal, and the rest
# still gather
# ===========================================================================

# Ledger absent
new_repo
add_panel
add_tasks
add_commits
run_it
[ "$RC" -eq 0 ] && pass "ledger absent: exits 0" \
  || fail "ledger absent: rc=$RC out=$OUT"
case "$OUT" in
  *"skipped: docs/superpowers/ledgers/demo.md (absent)"*) \
    pass "ledger absent: reported as skipped" ;;
  *) fail "ledger absent: not reported as skipped: $OUT" ;;
esac
case "$OUT" in
  *PANEL-BODY*) pass "ledger absent: panel still present" ;;
  *) fail "ledger absent: panel missing: $OUT" ;;
esac
case "$OUT" in
  *TASKS-BODY*) pass "ledger absent: tasks.md still present" ;;
  *) fail "ledger absent: tasks.md missing: $OUT" ;;
esac
case "$OUT" in
  *"ARCHIVE-COMMIT-BODY"*) pass "ledger absent: git log still present" ;;
  *) fail "ledger absent: git log missing: $OUT" ;;
esac

# Panel absent (the common case — no optional review-panel slot fired)
new_repo
add_ledger
add_tasks
add_commits
run_it
[ "$RC" -eq 0 ] && pass "panel absent: exits 0" \
  || fail "panel absent: rc=$RC out=$OUT"
case "$OUT" in
  *"skipped: docs/superpowers/reviews/demo-panel.md (absent)"*) \
    pass "panel absent: reported as skipped" ;;
  *) fail "panel absent: not reported as skipped: $OUT" ;;
esac
case "$OUT" in
  *LEDGER-BODY*) pass "panel absent: ledger still present" ;;
  *) fail "panel absent: ledger missing: $OUT" ;;
esac
case "$OUT" in
  *TASKS-BODY*) pass "panel absent: tasks.md still present" ;;
  *) fail "panel absent: tasks.md missing: $OUT" ;;
esac

# tasks.md absent from the archived path
new_repo
add_ledger
add_panel
add_commits
run_it
[ "$RC" -eq 0 ] && pass "tasks.md absent: exits 0" \
  || fail "tasks.md absent: rc=$RC out=$OUT"
case "$OUT" in
  *"skipped: $REL/tasks.md (absent)"*) \
    pass "tasks.md absent: reported as skipped" ;;
  *) fail "tasks.md absent: not reported as skipped: $OUT" ;;
esac
case "$OUT" in
  *LEDGER-BODY*) pass "tasks.md absent: ledger still present" ;;
  *) fail "tasks.md absent: ledger missing: $OUT" ;;
esac
case "$OUT" in
  *PANEL-BODY*) pass "tasks.md absent: panel still present" ;;
  *) fail "tasks.md absent: panel missing: $OUT" ;;
esac

# ===========================================================================
# SECTION: All four sources absent
# ===========================================================================

new_repo
run_it
[ "$RC" -eq 0 ] && pass "all sources absent: exits 0" \
  || fail "all sources absent: rc=$RC out=$OUT"
case "$OUT" in
  *"skipped: docs/superpowers/ledgers/demo.md (absent)"*) \
    pass "all sources absent: ledger reported skipped" ;;
  *) fail "all sources absent: ledger not reported skipped: $OUT" ;;
esac
case "$OUT" in
  *"skipped: docs/superpowers/reviews/demo-panel.md (absent)"*) \
    pass "all sources absent: panel reported skipped" ;;
  *) fail "all sources absent: panel not reported skipped: $OUT" ;;
esac
case "$OUT" in
  *"skipped: $REL/tasks.md (absent)"*) \
    pass "all sources absent: tasks.md reported skipped" ;;
  *) fail "all sources absent: tasks.md not reported skipped: $OUT" ;;
esac
case "$OUT" in
  *"skipped: git log --stat (absent)"*) \
    pass "all sources absent: git log reported skipped" ;;
  *) fail "all sources absent: git log not reported skipped: $OUT" ;;
esac

# ===========================================================================
# SECTION: F1 — a symlink planted at a tracked source location is refused,
# not read. The symlink's target content must never reach the bundle.
# ===========================================================================

new_repo
OUTSIDE="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-outside.XXXXXX")"
TREES+=("$OUTSIDE")
printf 'SECRET-OUTSIDE-CONTENT\n' > "$OUTSIDE/secret.txt"
mkdir -p "$REPO/docs/superpowers/ledgers"
ln -s "$OUTSIDE/secret.txt" "$REPO/docs/superpowers/ledgers/2026-01-01-demo.md"
add_panel
add_tasks
add_commits
run_it
[ "$RC" -eq 0 ] && pass "ledger symlink escape: exits 0" \
  || fail "ledger symlink escape: rc=$RC out=$OUT"
case "$OUT" in
  *"refused: docs/superpowers/ledgers/demo.md (resolves outside the repository)"*) \
    pass "ledger symlink escape: reported as refused" ;;
  *) fail "ledger symlink escape: not reported as refused: $OUT" ;;
esac
case "$OUT" in
  *"skipped: docs/superpowers/ledgers/demo.md (absent)"*) \
    fail "ledger symlink escape: wrongly also reported as skipped: $OUT" ;;
  *) pass "ledger symlink escape: not reported as skipped" ;;
esac
case "$OUT" in
  *SECRET-OUTSIDE-CONTENT*) fail "ledger symlink escape: target content leaked into bundle: $OUT" ;;
  *) pass "ledger symlink escape: target content not in bundle" ;;
esac
case "$OUT" in
  *PANEL-BODY*) pass "ledger symlink escape: panel still present" ;;
  *) fail "ledger symlink escape: panel missing: $OUT" ;;
esac

# ===========================================================================
# SECTION: F3 — a nonexistent archived-change-path is distinguished from a
# legitimate absence
# ===========================================================================

new_repo
rm -rf "$ARCHIVED"
run_it
[ "$RC" -eq 0 ] && pass "archived path missing: exits 0" \
  || fail "archived path missing: rc=$RC out=$OUT"
case "$OUT" in
  *"note: archived-change-path '$REL' does not resolve to the expected spectre/changes/archive/<leaf> location — every source below is reported skipped for that reason"*) \
    pass "archived path missing: note printed" ;;
  *) fail "archived path missing: note not printed: $OUT" ;;
esac
case "$OUT" in
  *"skipped: docs/superpowers/ledgers/demo.md (absent)"*) \
    pass "archived path missing: ledger reported skipped" ;;
  *) fail "archived path missing: ledger not reported skipped: $OUT" ;;
esac
case "$OUT" in
  *"skipped: git log --stat (absent)"*) \
    pass "archived path missing: git log reported skipped" ;;
  *) fail "archived path missing: git log not reported skipped: $OUT" ;;
esac

# ===========================================================================
# SECTION: F12 — the archived-change-path itself is a symlink. Git tracks
# symlinks as committed blobs, so this could be a merged, tracked symlink
# pointing anywhere. Must be refused as an invalid invocation (same code path
# as "archived path missing"), and the target's tasks.md content must never
# reach the bundle.
# ===========================================================================

new_repo
rm -rf "$ARCHIVED"
OUTSIDE_DIR="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-outside-dir.XXXXXX")"
TREES+=("$OUTSIDE_DIR")
printf 'TOP-SECRET-TASKS-CONTENT\n' > "$OUTSIDE_DIR/tasks.md"
ln -s "$OUTSIDE_DIR" "$ARCHIVED"
add_commits
run_it
[ "$RC" -eq 0 ] && pass "archived path is symlink: exits 0" \
  || fail "archived path is symlink: rc=$RC out=$OUT"
case "$OUT" in
  *"note: archived-change-path '$REL' resolves through a symlink somewhere between the repository root and the leaf"*) \
    pass "archived path is symlink: distinct note printed" ;;
  *) fail "archived path is symlink: distinct note not printed: $OUT" ;;
esac
case "$OUT" in
  *TOP-SECRET-TASKS-CONTENT*) fail "archived path is symlink: target tasks.md content leaked into bundle: $OUT" ;;
  *) pass "archived path is symlink: target tasks.md content not in bundle" ;;
esac
case "$OUT" in
  *"skipped: $REL/tasks.md (absent)"*) \
    pass "archived path is symlink: tasks.md reported skipped" ;;
  *) fail "archived path is symlink: tasks.md not reported skipped: $OUT" ;;
esac
case "$OUT" in
  *"refused: 0 of 4 sources"*) pass "archived path is symlink: no source reported refused (skipped, like the missing-path case)" ;;
  *) fail "archived path is symlink: expected zero refused sources: $OUT" ;;
esac

# ===========================================================================
# SECTION: F13 — the exact same symlinked-archive-directory attack as
# "archived path is symlink" above, but invoked WITH a trailing slash on
# $ARCHIVED_PATH (the shape skills/myflow-finish/SKILL.md actually
# documents: spectre/changes/archive/<name>/). Must be refused
# identically, never leaked, regardless of the trailing slash.
# ===========================================================================

new_repo
rm -rf "$ARCHIVED"
OUTSIDE_DIR_SLASH="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-outside-slash.XXXXXX")"
TREES+=("$OUTSIDE_DIR_SLASH")
printf 'TOP-SECRET-F13-CONTENT\n' > "$OUTSIDE_DIR_SLASH/tasks.md"
ln -s "$OUTSIDE_DIR_SLASH" "$ARCHIVED"
add_commits
run_with "$REL/"
[ "$RC" -eq 0 ] && pass "archived path is symlink with trailing slash: exits 0" \
  || fail "archived path is symlink with trailing slash: rc=$RC out=$OUT"
case "$OUT" in
  *"note: archived-change-path '$REL' resolves through a symlink somewhere between the repository root and the leaf"*) \
    pass "archived path is symlink with trailing slash: distinct note printed" ;;
  *) fail "archived path is symlink with trailing slash: distinct note not printed: $OUT" ;;
esac
case "$OUT" in
  *TOP-SECRET-F13-CONTENT*) \
    fail "archived path is symlink with trailing slash: target tasks.md content leaked into bundle: $OUT" ;;
  *) pass "archived path is symlink with trailing slash: target tasks.md content not in bundle" ;;
esac
case "$OUT" in
  *"skipped: $REL/tasks.md (absent)"*) \
    pass "archived path is symlink with trailing slash: tasks.md reported skipped" ;;
  *) fail "archived path is symlink with trailing slash: tasks.md not reported skipped: $OUT" ;;
esac

# ===========================================================================
# SECTION: F14 — spectre/changes/archive itself (one level above the leaf)
# is a symlink. The leaf ($ARCHIVED, the archived-change directory) is an
# ordinary, non-symlink directory underneath it, so the F12/F13 leaf-only
# check must not be what catches this — the lexical-vs-real comparison must.
# ===========================================================================

new_repo
OUTSIDE_ARCHIVE_DIR="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-outside-archive.XXXXXX")"
TREES+=("$OUTSIDE_ARCHIVE_DIR")
mkdir -p "$OUTSIDE_ARCHIVE_DIR/2026-01-01-demo"
printf 'TOP-SECRET-F14-CONTENT\n' > "$OUTSIDE_ARCHIVE_DIR/2026-01-01-demo/tasks.md"
rm -rf "$REPO/spectre/changes/archive"
ln -s "$OUTSIDE_ARCHIVE_DIR" "$REPO/spectre/changes/archive"
add_commits
run_it
[ "$RC" -eq 0 ] && pass "archive ancestor is symlink: exits 0" \
  || fail "archive ancestor is symlink: rc=$RC out=$OUT"
case "$OUT" in
  *"note: archived-change-path '$REL' resolves through a symlink somewhere between the repository root and the leaf"*) \
    pass "archive ancestor is symlink: distinct note printed" ;;
  *) fail "archive ancestor is symlink: distinct note not printed: $OUT" ;;
esac
case "$OUT" in
  *TOP-SECRET-F14-CONTENT*) \
    fail "archive ancestor is symlink: target tasks.md content leaked into bundle: $OUT" ;;
  *) pass "archive ancestor is symlink: target tasks.md content not in bundle" ;;
esac
case "$OUT" in
  *"skipped: $REL/tasks.md (absent)"*) \
    pass "archive ancestor is symlink: tasks.md reported skipped" ;;
  *) fail "archive ancestor is symlink: tasks.md not reported skipped: $OUT" ;;
esac
case "$OUT" in
  *"refused: 0 of 4 sources"*) \
    pass "archive ancestor is symlink: no source reported refused (skipped, like the missing-path case)" ;;
  *) fail "archive ancestor is symlink: expected zero refused sources: $OUT" ;;
esac

# ===========================================================================
# SECTION: F21 — the same must catch an ancestor symlink at hop 2
# (spectre/changes/) too, not just hop 1 (archive/, tested above as F14).
# ===========================================================================

new_repo
# add_commits runs BEFORE the ancestor is clobbered with a symlink below:
# it now stages real commits under spectre/changes/demo/ (finding A, this
# change's own review panel, pass 3), and `git add` refuses any path beyond
# a symlink — so once spectre/changes/ itself is a symlink, a real `mkdir
# -p`/`git add` under it would fail outright rather than reach the
# invocation this section exists to test. The symlink swap is a plain
# filesystem operation the script reads live off disk; it does not touch
# what git log --stat already has recorded for the commits add_commits made.
add_commits
OUTSIDE_CHANGES_DIR="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-outside-changes.XXXXXX")"
TREES+=("$OUTSIDE_CHANGES_DIR")
mkdir -p "$OUTSIDE_CHANGES_DIR/archive/2026-01-01-demo"
printf 'TOP-SECRET-F21-HOP2-CONTENT\n' > "$OUTSIDE_CHANGES_DIR/archive/2026-01-01-demo/tasks.md"
rm -rf "$REPO/spectre/changes"
ln -s "$OUTSIDE_CHANGES_DIR" "$REPO/spectre/changes"
run_it
[ "$RC" -eq 0 ] && pass "changes ancestor is symlink (hop 2): exits 0" \
  || fail "changes ancestor is symlink (hop 2): rc=$RC out=$OUT"
case "$OUT" in
  *"note: archived-change-path '$REL' resolves through a symlink somewhere between the repository root and the leaf"*) \
    pass "changes ancestor is symlink (hop 2): distinct note printed" ;;
  *) fail "changes ancestor is symlink (hop 2): distinct note not printed: $OUT" ;;
esac
case "$OUT" in
  *TOP-SECRET-F21-HOP2-CONTENT*) \
    fail "changes ancestor is symlink (hop 2): target tasks.md content leaked into bundle: $OUT" ;;
  *) pass "changes ancestor is symlink (hop 2): target tasks.md content not in bundle" ;;
esac
case "$OUT" in
  *"skipped: $REL/tasks.md (absent)"*) \
    pass "changes ancestor is symlink (hop 2): tasks.md reported skipped" ;;
  *) fail "changes ancestor is symlink (hop 2): tasks.md not reported skipped: $OUT" ;;
esac

# ===========================================================================
# SECTION: F21 — the same must catch an ancestor symlink at hop 3 (spectre/
# itself) too.
# ===========================================================================

new_repo
# add_commits runs BEFORE the ancestor is clobbered — see the matching note
# on the hop-2 section above.
add_commits
OUTSIDE_OPENSPEC_DIR="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-outside-spectre.XXXXXX")"
TREES+=("$OUTSIDE_OPENSPEC_DIR")
mkdir -p "$OUTSIDE_OPENSPEC_DIR/changes/archive/2026-01-01-demo"
printf 'TOP-SECRET-F21-HOP3-CONTENT\n' > "$OUTSIDE_OPENSPEC_DIR/changes/archive/2026-01-01-demo/tasks.md"
rm -rf "$REPO/spectre"
ln -s "$OUTSIDE_OPENSPEC_DIR" "$REPO/spectre"
run_it
[ "$RC" -eq 0 ] && pass "spectre ancestor is symlink (hop 3): exits 0" \
  || fail "spectre ancestor is symlink (hop 3): rc=$RC out=$OUT"
case "$OUT" in
  *"note: archived-change-path '$REL' resolves through a symlink somewhere between the repository root and the leaf"*) \
    pass "spectre ancestor is symlink (hop 3): distinct note printed" ;;
  *) fail "spectre ancestor is symlink (hop 3): distinct note not printed: $OUT" ;;
esac
case "$OUT" in
  *TOP-SECRET-F21-HOP3-CONTENT*) \
    fail "spectre ancestor is symlink (hop 3): target tasks.md content leaked into bundle: $OUT" ;;
  *) pass "spectre ancestor is symlink (hop 3): target tasks.md content not in bundle" ;;
esac
case "$OUT" in
  *"skipped: $REL/tasks.md (absent)"*) \
    pass "spectre ancestor is symlink (hop 3): tasks.md reported skipped" ;;
  *) fail "spectre ancestor is symlink (hop 3): tasks.md not reported skipped: $OUT" ;;
esac

# ===========================================================================
# SECTION: F22 — a symlinked ancestor at a nesting depth the old bounded
# 3-hop walk never even reached, combined with an EXTRA directory level under
# spectre/changes/archive/ that isn't part of the documented shape at all.
# spectre/ itself is a symlink, and the archived-change-path carries one
# extra "subdir/" segment before the archived-change leaf. The old bounded walk
# would have missed both the extra nesting (never checked at all — the walk
# only ever considered the fixed 3 hops for the *documented* shape) and,
# depending on how the extra segment shifted the hop count, potentially the
# ancestor symlink too. The new mechanism refuses this outright on shape
# alone (the lexical path doesn't resolve to a single-segment leaf under
# spectre/changes/archive/), before it even reaches the symlink comparison.
# ===========================================================================

new_repo
# add_commits runs BEFORE the ancestor is clobbered — see the matching note
# on the hop-2 section above.
add_commits
OUTSIDE_F22_DIR="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-outside-f22.XXXXXX")"
TREES+=("$OUTSIDE_F22_DIR")
mkdir -p "$OUTSIDE_F22_DIR/changes/archive/subdir/2026-01-01-demo"
printf 'TOP-SECRET-F22-CONTENT\n' > "$OUTSIDE_F22_DIR/changes/archive/subdir/2026-01-01-demo/tasks.md"
rm -rf "$REPO/spectre"
ln -s "$OUTSIDE_F22_DIR" "$REPO/spectre"
F22_ARG="spectre/changes/archive/subdir/2026-01-01-demo"
run_with "$F22_ARG"
[ "$RC" -eq 0 ] && pass "F22 extra nesting + symlinked spectre: exits 0" \
  || fail "F22 extra nesting + symlinked spectre: rc=$RC out=$OUT"
case "$OUT" in
  *"note: archived-change-path '$F22_ARG' does not resolve to the expected spectre/changes/archive/<leaf> location"*) \
    pass "F22 extra nesting + symlinked spectre: shape note printed" ;;
  *) fail "F22 extra nesting + symlinked spectre: shape note not printed: $OUT" ;;
esac
case "$OUT" in
  *TOP-SECRET-F22-CONTENT*) \
    fail "F22 extra nesting + symlinked spectre: target tasks.md content leaked into bundle: $OUT" ;;
  *) pass "F22 extra nesting + symlinked spectre: target tasks.md content not in bundle" ;;
esac
case "$OUT" in
  *"skipped: $F22_ARG/tasks.md (absent)"*) \
    pass "F22 extra nesting + symlinked spectre: tasks.md reported skipped" ;;
  *) fail "F22 extra nesting + symlinked spectre: tasks.md not reported skipped: $OUT" ;;
esac
case "$OUT" in
  *"refused: 0 of 4 sources"*) \
    pass "F22 extra nesting + symlinked spectre: no source reported refused (skipped, like the missing-path case)" ;;
  *) fail "F22 extra nesting + symlinked spectre: expected zero refused sources: $OUT" ;;
esac

# ===========================================================================
# SECTION: F20 — a '..' path component that would desync a lexical hop-count
# walk from the real (kernel-resolved) filesystem ancestors, letting a
# symlinked ancestor slip past undetected. Reproduces the exact bypass shape:
# spectre/ itself is a symlink to an evil directory (same target shape as
# the hop-3 case above), and the archived-change-path is invoked with an
# injected '<dir>/../' segment. The new lexical-normalize step collapses
# this '..' by pure string manipulation (no filesystem access), so the
# lexical side lands on the ordinary-looking
# spectre/changes/archive/2026-01-01-demo shape and passes the shape check —
# but the real, symlink-following resolution in step 4 lands somewhere else
# entirely (through the spectre symlink), so the two diverge and the
# invocation is refused via the same lexical-vs-real comparison as every
# other symlink shape above.
# ===========================================================================

new_repo
# add_commits runs BEFORE the ancestor is clobbered — see the matching note
# on the hop-2 section above.
add_commits
rm -rf "$REPO/spectre"
OUTSIDE_DOTDOT="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-outside-dotdot.XXXXXX")"
TREES+=("$OUTSIDE_DOTDOT")
mkdir -p "$OUTSIDE_DOTDOT/changes/archive/extra"
mkdir -p "$OUTSIDE_DOTDOT/changes/archive/2026-01-01-demo"
printf 'TOP-SECRET-F20-CONTENT\n' > "$OUTSIDE_DOTDOT/changes/archive/2026-01-01-demo/tasks.md"
ln -s "$OUTSIDE_DOTDOT" "$REPO/spectre"
DOTDOT_ARG="spectre/changes/archive/extra/../2026-01-01-demo"
run_with "$DOTDOT_ARG"
[ "$RC" -eq 0 ] && pass "dotdot ancestor-walk bypass: exits 0" \
  || fail "dotdot ancestor-walk bypass: rc=$RC out=$OUT"
case "$OUT" in
  *"note: archived-change-path '$DOTDOT_ARG' resolves through a symlink somewhere between the repository root and the leaf"*) \
    pass "dotdot ancestor-walk bypass: distinct note printed" ;;
  *) fail "dotdot ancestor-walk bypass: distinct note not printed: $OUT" ;;
esac
case "$OUT" in
  *TOP-SECRET-F20-CONTENT*) \
    fail "dotdot ancestor-walk bypass: target tasks.md content leaked into bundle: $OUT" ;;
  *) pass "dotdot ancestor-walk bypass: target tasks.md content not in bundle" ;;
esac
case "$OUT" in
  *"skipped: $DOTDOT_ARG/tasks.md (absent)"*) \
    pass "dotdot ancestor-walk bypass: tasks.md reported skipped" ;;
  *) fail "dotdot ancestor-walk bypass: tasks.md not reported skipped: $OUT" ;;
esac

# ===========================================================================
# SECTION: Sanity — a well-formed path containing an unremarkable,
# '..'-free relative segment (i.e. an ordinary relative-path invocation)
# still works normally. The '..'-collapsing lexical normalization above must
# not make ordinary relative invocation collateral damage.
# ===========================================================================

new_repo
add_ledger
add_panel
add_tasks
add_commits
run_it
[ "$RC" -eq 0 ] && pass "relative path without dotdot: exits 0" \
  || fail "relative path without dotdot: rc=$RC out=$OUT"
case "$OUT" in
  *TASKS-BODY*) pass "relative path without dotdot: tasks.md content in bundle" ;;
  *) fail "relative path without dotdot: tasks.md content missing: $OUT" ;;
esac
case "$OUT" in
  *LEDGER-BODY*) pass "relative path without dotdot: ledger content in bundle" ;;
  *) fail "relative path without dotdot: ledger content missing: $OUT" ;;
esac

# ===========================================================================
# SECTION: F8 — a bare "archive" substring in an implementation commit's
# subject must not exclude it from IMPL_SHA candidacy. Fixture realistic
# (this change's own review panel, pass 3, finding A): the implementation
# commit touches a real path outside the planning paths and the planning
# commit actually writes spectre/changes/demo/, so this resolves through
# the PRIMARY path+subject query — no subject-only fallback exists to fall
# through to.
# ===========================================================================

new_repo
(
  cd "$REPO" \
    && mkdir -p src \
    && printf 'x\n' > src/archive-browsing.txt \
    && git add src/archive-browsing.txt \
    && git commit -q -m "feat(demo): implement archive browsing UI ARCHIVE-SUBSTRING-MARKER" \
    && mkdir -p spectre/changes/demo \
    && printf 'F8-PLAN\n' > spectre/changes/demo/proposal.md \
    && git add spectre/changes/demo/proposal.md \
    && git commit -q -m "chore(demo): plan, test guide and session records" \
    && git commit -q --allow-empty -m "chore(spectre): archive demo" -m "ARCHIVE-COMMIT-BODY2"
)
run_it
[ "$RC" -eq 0 ] && pass "archive substring in impl subject: exits 0" \
  || fail "archive substring in impl subject: rc=$RC out=$OUT"
case "$OUT" in
  *ARCHIVE-SUBSTRING-MARKER*) \
    pass "archive substring in impl subject: impl commit still eligible" ;;
  *) fail "archive substring in impl subject: impl commit wrongly excluded: $OUT" ;;
esac

# ===========================================================================
# SECTION: F9 — a "plan and session records" substring in an IMPLEMENTATION
# commit's subject (not the real plan commit) must not exclude it from
# IMPL_SHA candidacy. IMPL_SHA is now derived from PLAN_SHA's own first
# parent and refused only when ITS subject matches PLAN_SUBJECT_RE_OLD/NEW
# or ARCHIVE_SUBJECT_RE from the start (`[[ "$PARENT_SUBJECT" =~ ... ]]`,
# anchored at `^`) — not a bare unanchored substring match — so an
# implementation commit whose subject merely CONTAINS the phrase, without
# starting with it, must still resolve. Fixture realistic (this change's own
# review panel, pass 3, finding A): the implementation commit touches a real
# path and the planning commit actually writes spectre/changes/demo/, so
# this resolves through the PRIMARY path+subject query.
# ===========================================================================

new_repo
(
  cd "$REPO" \
    && mkdir -p src \
    && printf 'x\n' > src/onboarding-docs.txt \
    && git add src/onboarding-docs.txt \
    && git commit -q -m "feat(demo): document plan and session records for onboarding PLAN-PHRASE-SUBSTRING-MARKER" \
    && mkdir -p spectre/changes/demo \
    && printf 'F9-PLAN\n' > spectre/changes/demo/proposal.md \
    && git add spectre/changes/demo/proposal.md \
    && git commit -q -m "chore(demo): plan and session records" \
    && git commit -q --allow-empty -m "chore(spectre): archive demo" -m "PLAN-PHRASE-ARCHIVE-BODY"
)
run_it
[ "$RC" -eq 0 ] && pass "plan-phrase substring in impl subject: exits 0" \
  || fail "plan-phrase substring in impl subject: rc=$RC out=$OUT"
case "$OUT" in
  *PLAN-PHRASE-SUBSTRING-MARKER*) \
    pass "plan-phrase substring in impl subject: impl commit still resolved as IMPL_SHA" ;;
  *) fail "plan-phrase substring in impl subject: impl commit wrongly excluded from IMPL_SHA: $OUT" ;;
esac

# ===========================================================================
# SECTION: F23 — TRUSTED_REPO_ROOT must resolve to the MAIN checkout even
# when this script's own process cwd is a worktree, not the main checkout.
# `git rev-parse --show-toplevel` would return the worktree's own root here;
# `--git-common-dir` (the fix) always points at the main repo's .git.
# Simulated with a real `git worktree add`: the archived-change-path and its
# tasks.md live only in the MAIN checkout (never created inside the
# worktree), so this case only passes if TRUSTED_REPO_ROOT — and therefore
# the archive_root it's checked against — resolves to the main checkout
# regardless of cwd.
# ===========================================================================

new_repo
add_tasks
# Canonicalize $ARCHIVED the same way TRUSTED_REPO_ROOT itself gets
# canonicalized (`pwd -P`), so an OS-level path alias under $TMPDIR (e.g.
# macOS's /var/folders/... -> /private/var/folders/...) can't produce a
# spurious lexical/real mismatch unrelated to what this case is testing.
ARCHIVED_REAL_ABS="$(cd -P "$ARCHIVED" && pwd -P)"
WORKTREE_DIR="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-worktree.XXXXXX")"
rmdir "$WORKTREE_DIR"
(cd "$REPO" && git worktree add -q -b f23-worktree-branch "$WORKTREE_DIR" >/dev/null)
TREES+=("$WORKTREE_DIR")
set +e
OUT="$(cd "$WORKTREE_DIR" && "$SCRIPT" "$ARCHIVED_REAL_ABS" demo "$STATE_DIR" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] && pass "cwd is a worktree: exits 0" \
  || fail "cwd is a worktree: rc=$RC out=$OUT"
case "$OUT" in
  *TASKS-BODY*) pass "cwd is a worktree: tasks.md content in bundle (TRUSTED_REPO_ROOT resolved to main checkout)" ;;
  *) fail "cwd is a worktree: tasks.md content missing — TRUSTED_REPO_ROOT likely resolved to the worktree instead of the main checkout: $OUT" ;;
esac
(cd "$REPO" && git worktree remove -f "$WORKTREE_DIR" >/dev/null 2>&1) || true

# Sanity companion: an ordinary main-checkout invocation (cwd never a
# worktree) is unaffected by the --git-common-dir switch.
new_repo
add_ledger
add_panel
add_tasks
add_commits
run_it
[ "$RC" -eq 0 ] && pass "cwd is the main checkout (unaffected by F23): exits 0" \
  || fail "cwd is the main checkout (unaffected by F23): rc=$RC out=$OUT"
case "$OUT" in
  *TASKS-BODY*) pass "cwd is the main checkout (unaffected by F23): tasks.md content in bundle" ;;
  *) fail "cwd is the main checkout (unaffected by F23): tasks.md content missing: $OUT" ;;
esac

# ===========================================================================
# SECTION: F26 — a literal glob metacharacter in the archived-change-path
# must not trigger pathname (glob) expansion during the step-2 lexical
# component-stack walk. Reproduced by placing sibling directories in $REPO
# whose names would surface in the bundle's note text if `for part in
# $abs_input` (unquoted) had glob-expanded instead of `IFS=/ read -ra`
# field-splitting.
# ===========================================================================

new_repo
mkdir -p "$REPO/spectre/changes/archive/GLOB-SIBLING-ONE"
mkdir -p "$REPO/spectre/changes/archive/GLOB-SIBLING-TWO"
GLOB_ARG="spectre/changes/archive/*"
run_with "$GLOB_ARG"
[ "$RC" -eq 0 ] && pass "F26 literal glob in archived-change-path: exits 0" \
  || fail "F26 literal glob in archived-change-path: rc=$RC out=$OUT"
case "$OUT" in
  *"does not resolve to the expected spectre/changes/archive/<leaf> location"*) \
    pass "F26 literal glob in archived-change-path: refused as shape (no bypass)" ;;
  *) fail "F26 literal glob in archived-change-path: not refused as shape: $OUT" ;;
esac
case "$OUT" in
  *GLOB-SIBLING-ONE* | *GLOB-SIBLING-TWO*) \
    fail "F26 literal glob in archived-change-path: sibling directory name leaked into output — glob expanded: $OUT" ;;
  *) pass "F26 literal glob in archived-change-path: no sibling directory name leaked (glob did not expand)" ;;
esac

# ===========================================================================
# SECTION: F25 — a leaf that is a symlink pointing at a plain FILE (not a
# directory) must be reported with reason "symlink", not "shape" — it fails
# `[ -d ]` the same way a missing path does, but the real issue is the
# symlink, not the shape.
# ===========================================================================

new_repo
rmdir "$ARCHIVED"
printf 'F25-SYMLINK-TARGET-FILE\n' > "$REPO/f25-target-file"
ln -s "$REPO/f25-target-file" "$ARCHIVED"
run_it
[ "$RC" -eq 0 ] && pass "F25 leaf symlink to a file: exits 0" \
  || fail "F25 leaf symlink to a file: rc=$RC out=$OUT"
case "$OUT" in
  *"note: archived-change-path '$REL' resolves through a symlink somewhere between the repository root and the leaf"*) \
    pass "F25 leaf symlink to a file: reported as symlink, not shape" ;;
  *) fail "F25 leaf symlink to a file: not reported as symlink: $OUT" ;;
esac
case "$OUT" in
  *F25-SYMLINK-TARGET-FILE*) fail "F25 leaf symlink to a file: target file content leaked into bundle: $OUT" ;;
  *) pass "F25 leaf symlink to a file: target file content not in bundle" ;;
esac

# ===========================================================================
# SECTION: New plan-commit subject wording ("plan and session records",
# without "test guide") must resolve as PLAN_SHA and must NOT be picked up
# by IMPL_SHA. The old-old subject ("plan, test guide and session records")
# is kept as its own fixture elsewhere in this file (add_commits, and the F8
# section above) as the back-compat proof; this section is the new-wording
# counterpart, proving both subjects resolve correctly side by side.
#
# This is also the dedicated old-era, end-to-end fixture finding A (this
# change's own review panel, pass 3) calls for: a real implementation
# commit touching a real path, a real planning commit that actually writes
# spectre/changes/demo/, under the old subject, and a real archive
# `git mv` to the archived location — proving old-era resolution goes
# through the PRIMARY path+subject query. There is no longer a subject-only
# fallback to fall through to (finding A deleted it as unreachable for any
# genuine historical commit — see gather-self-review-context.sh's own
# header note), so any successful resolution below IS resolution through
# the primary query, by construction.
# ===========================================================================

new_repo
(
  cd "$REPO" \
    && mkdir -p src \
    && printf 'x\n' > src/new-wording-feature.txt \
    && git add src/new-wording-feature.txt \
    && git commit -q -m "feat(demo): NEW-WORDING-IMPLEMENTATION-COMMIT-BODY" \
    && mkdir -p spectre/changes/demo \
    && printf 'NEW-WORDING-PLAN\n' > spectre/changes/demo/proposal.md \
    && git add spectre/changes/demo/proposal.md \
    && git commit -q -m "chore(demo): plan and session records NEW-WORDING-PLAN-COMMIT-BODY" \
    && git mv spectre/changes/demo/proposal.md "$REL/proposal.md" \
    && git commit -q -m "chore(spectre): archive demo" -m "NEW-WORDING-ARCHIVE-COMMIT-BODY"
)
run_it
[ "$RC" -eq 0 ] && pass "new-wording plan commit: exits 0" \
  || fail "new-wording plan commit: rc=$RC out=$OUT"
case "$OUT" in
  *NEW-WORDING-PLAN-COMMIT-BODY*) \
    pass "new-wording plan commit: plan commit content is present in the bundle" ;;
  *) fail "new-wording plan commit: plan commit content missing from the bundle: $OUT" ;;
esac
# The specific failure this section exists to catch (not just "PLAN_SHA was
# found"): if IMPL_SHA's exclusion grep is not widened to match the new
# wording too, IMPL_SHA's head-1-of-most-recent-first selection picks the
# plan commit itself — it is more recent than the true implementation
# commit and, unexcluded, outranks it. The true implementation commit is
# then never reached at all. So the decisive check is that the
# implementation commit's own body marker made it into the bundle —
# proving IMPL_SHA resolved to the real implementation commit, not to the
# plan commit a second time.
case "$OUT" in
  *NEW-WORDING-IMPLEMENTATION-COMMIT-BODY*) \
    pass "new-wording plan commit: implementation commit resolved as IMPL_SHA (not shadowed by the plan commit)" ;;
  *) fail "new-wording plan commit: implementation commit not resolved as IMPL_SHA — the plan commit likely shadowed it: $OUT" ;;
esac

# ===========================================================================
# SECTION: task 11 (KAN-202) — commit resolution under the module-scope
# subjects finish run 1 now writes. The implementation and planning commit
# subjects no longer carry the change name at all (the module-scope commit
# convention), so PLAN_SHA and IMPL_SHA can no longer be resolved by a
# subject grep keyed on the change name — path-based resolution (task 12)
# must do it instead. ARCHIVE_SHA now carries a module scope too —
# `chore(spectre): archive <name>` — and is kept change-specific by the
# change name in its DESCRIPTION plus a both-ends-anchored grep, which the
# dedicated distractor section below this one proves.
# ===========================================================================

# add_module_scope_commits — the implementation and planning commits back to
# back, in that order, exactly as commit-split.sh produces them: the
# implementation commit touches a real file outside the planning paths under
# a module-scoped subject; the planning commit touches THIS change's own
# LIVE directory (spectre/changes/demo/tasks.md) — the signal path-based
# resolution now keys on exclusively (pass 2, finding F: PLAN_SHA is
# searched only at the live path, never the archived one, since `git log
# -- <path>` follows a commit's historical tree through a later rename) —
# under the fixed literal subject finish run 1 now writes, with no change
# name in it at all. `git mv` to the archived location and the archive
# commit follow, matching the REAL run-1-then-run-2 sequence rather than
# committing directly at the archived path, which never happens in a real
# run (see the dedicated section below this one for why that distinction
# matters).
add_module_scope_commits() {
  (
    cd "$REPO" \
      && mkdir -p scripts/some-module spectre/changes/demo \
      && printf 'x\n' > scripts/some-module/foo.sh \
      && git add scripts/some-module/foo.sh \
      && git commit -q -m "feat(some-module): MODULE-SCOPE-IMPL-BODY" \
      && printf 'TASKS-MODULE-SCOPE\n' > spectre/changes/demo/tasks.md \
      && git add spectre/changes/demo/tasks.md \
      && git commit -q -m "chore(spectre): plan and session records MODULE-SCOPE-PLAN-BODY" \
      && git mv spectre/changes/demo "$REL" \
      && git commit -q -m "chore(spectre): archive demo" -m "MODULE-SCOPE-ARCHIVE-BODY"
  )
}

new_repo
add_module_scope_commits
run_it
[ "$RC" -eq 0 ] && pass "module-scope subjects: exits 0" \
  || fail "module-scope subjects: rc=$RC out=$OUT"
case "$OUT" in
  *MODULE-SCOPE-IMPL-BODY*) \
    pass "module-scope subjects: implementation commit resolved as IMPL_SHA (the planning commit's own first parent)" ;;
  *) fail "module-scope subjects: implementation commit not resolved as IMPL_SHA: $OUT" ;;
esac
case "$OUT" in
  *MODULE-SCOPE-PLAN-BODY*) pass "module-scope subjects: planning commit resolved as PLAN_SHA by path" ;;
  *) fail "module-scope subjects: planning commit not resolved as PLAN_SHA: $OUT" ;;
esac
case "$OUT" in
  *MODULE-SCOPE-ARCHIVE-BODY*) \
    pass "module-scope subjects: archive commit still resolved as ARCHIVE_SHA under its module scope" ;;
  *) fail "module-scope subjects: archive commit not resolved as ARCHIVE_SHA: $OUT" ;;
esac

# A DIFFERENT change's planning commit — textually IDENTICAL
# ("chore(spectre): plan and session records" is now the same literal
# subject across every change) and landed LATER in this repository's
# history than demo's own. This is the case that distinguishes path-scoped
# resolution from a merely widened subject grep: a subject-only
# --max-count=1 match would return this later, unrelated commit instead of
# demo's own, per the operator's stated reason for rejecting that
# alternative.
new_repo
add_module_scope_commits
(
  cd "$REPO" \
    && mkdir -p spectre/changes/archive/2026-01-02-demo2 \
    && printf 'OTHER-CHANGE-TASKS\n' > spectre/changes/archive/2026-01-02-demo2/tasks.md \
    && git add spectre/changes/archive/2026-01-02-demo2/tasks.md \
    && git commit -q -m "chore(spectre): plan and session records OTHER-CHANGE-PLAN-BODY"
)
run_it
[ "$RC" -eq 0 ] && pass "module-scope subjects (later distractor plan commit present): exits 0" \
  || fail "module-scope subjects (later distractor plan commit present): rc=$RC out=$OUT"
case "$OUT" in
  *MODULE-SCOPE-PLAN-BODY*) \
    pass "module-scope subjects (later distractor plan commit present): own planning commit still resolved by path" ;;
  *) fail "module-scope subjects (later distractor plan commit present): own planning commit not resolved: $OUT" ;;
esac
case "$OUT" in
  *MODULE-SCOPE-IMPL-BODY*) \
    pass "module-scope subjects (later distractor plan commit present): own implementation commit still resolved as IMPL_SHA" ;;
  *) fail "module-scope subjects (later distractor plan commit present): own implementation commit not resolved: $OUT" ;;
esac
case "$OUT" in
  *"OTHER-CHANGE-PLAN-BODY"* | *"OTHER-CHANGE-TASKS"*) \
    fail "module-scope subjects (later distractor plan commit present): a different change's later planning commit wrongly resolved instead of demo's own: $OUT" ;;
  *) pass "module-scope subjects (later distractor plan commit present): a different change's later planning commit correctly excluded" ;;
esac

# ===========================================================================
# A DIFFERENT change's ARCHIVE commit, landed LATER, whose name has demo's
# as a PREFIX — `demo-fix-1`, the exact shape a nested fix sub-change takes.
# Since the spectre cutover the archive subject carries a module scope
# (`chore(spectre): archive <name>`) shared by every change, so the only
# thing separating demo's archive commit from this one is the change name in
# the description AND the fact that ARCHIVE_SUBJECT_RE is anchored at both
# ends. Drop either anchor and `archive demo` matches `archive demo-fix-1`
# too; --max-count=1 then returns this later, unrelated commit and the
# self-review bundle quietly describes the wrong change.
# ===========================================================================

new_repo
add_module_scope_commits
(
  cd "$REPO" \
    && git commit -q --allow-empty -m "chore(spectre): archive demo-fix-1" -m "SIBLING-ARCHIVE-BODY"
)
run_it
[ "$RC" -eq 0 ] && pass "prefix-sibling archive commit present: exits 0" \
  || fail "prefix-sibling archive commit present: rc=$RC out=$OUT"
case "$OUT" in
  *MODULE-SCOPE-ARCHIVE-BODY*) \
    pass "prefix-sibling archive commit present: demo's own archive commit still resolved" ;;
  *) fail "prefix-sibling archive commit present: demo's own archive commit not resolved: $OUT" ;;
esac
case "$OUT" in
  *SIBLING-ARCHIVE-BODY*) \
    fail "prefix-sibling archive commit present: a prefix-sibling change's later archive commit wrongly resolved instead of demo's own: $OUT" ;;
  *) pass "prefix-sibling archive commit present: a prefix-sibling change's later archive commit correctly excluded" ;;
esac

# ===========================================================================
# The REAL run-1-then-run-2 sequence: finish run 1 commits the planning
# artifacts at the LIVE location, spectre/changes/<name>/, because the change
# is not archived yet; run 2 only later renames that directory under
# spectre/changes/archive/. So the commit PLAN_SHA must find is one that
# touched the LIVE path — and, since pass 2's finding F, the LIVE path is the
# ONLY pathspec PLAN_SHA is ever searched at; the prior ALSO-searched archived
# location was removed because `git log -- <path>` already follows a commit's
# historical tree through the later `git mv`, so the live pathspec alone
# finds it.
#
# Found by the parent's own mutation pass during the fix round that added
# path-based resolution, and repaired here rather than filed.
new_repo
(
  cd "$REPO" \
    && mkdir -p scripts/some-module spectre/changes/demo \
    && printf 'x\n' > scripts/some-module/foo.sh \
    && git add scripts/some-module/foo.sh \
    && git commit -q -m "feat(some-module): LIVE-PATH-IMPL-BODY" \
    && printf 'TASKS-LIVE\n' > spectre/changes/demo/tasks.md \
    && git add spectre/changes/demo/tasks.md \
    && git commit -q -m "chore(spectre): plan and session records LIVE-PATH-PLAN-BODY" \
    && git mv spectre/changes/demo "$REL" \
    && git commit -q -m "chore(spectre): archive demo" -m "LIVE-PATH-ARCHIVE-BODY"
)
run_it
[ "$RC" -eq 0 ] && pass "run-1-live-then-run-2-archive: exits 0" \
  || fail "run-1-live-then-run-2-archive: rc=$RC out=$OUT"
case "$OUT" in
  *LIVE-PATH-PLAN-BODY*) \
    pass "run-1-live-then-run-2-archive: planning commit resolved through PLAN_PATH_LIVE, the path it actually touched" ;;
  *) fail "run-1-live-then-run-2-archive: planning commit not resolved: $OUT" ;;
esac
case "$OUT" in
  *LIVE-PATH-IMPL-BODY*) \
    pass "run-1-live-then-run-2-archive: implementation commit resolved as the planning commit's first parent" ;;
  *) fail "run-1-live-then-run-2-archive: implementation commit not resolved: $OUT" ;;
esac

# ===========================================================================
# SECTION: pass 2, finding E — a LATER commit that merely touches the same
# (now archived) directory, under a subject that is NOT a plan-commit
# subject, must never outrank the real planning commit. Before the fix,
# PLAN_SHA was matched by path alone, so this later commit — the most
# recent commit touching the path — became PLAN_SHA and the real planning
# commit disappeared from the bundle entirely.
# ===========================================================================
new_repo
(
  cd "$REPO" \
    && mkdir -p scripts/some-module spectre/changes/demo \
    && printf 'x\n' > scripts/some-module/foo.sh \
    && git add scripts/some-module/foo.sh \
    && git commit -q -m "feat(some-module): E-IMPL-BODY" \
    && printf 'TASKS-E\n' > spectre/changes/demo/tasks.md \
    && git add spectre/changes/demo/tasks.md \
    && git commit -q -m "chore(spectre): plan and session records E-PLAN-BODY" \
    && git mv spectre/changes/demo "$REL" \
    && git commit -q -m "chore(spectre): archive demo" -m "E-ARCHIVE-BODY" \
    && printf 'TASKS-E-FIXED-TYPO\n' > "$REL/tasks.md" \
    && git add "$REL/tasks.md" \
    && git commit -q -m "docs(spectre): fix a typo in an archived plan"
)
run_it
[ "$RC" -eq 0 ] && pass "finding E: exits 0" \
  || fail "finding E: rc=$RC out=$OUT"
case "$OUT" in
  *E-PLAN-BODY*) pass "finding E: the real planning commit is still resolved as PLAN_SHA" ;;
  *) fail "finding E: the real planning commit was not resolved: $OUT" ;;
esac
case "$OUT" in
  *"fix a typo in an archived plan"*) \
    fail "finding E: the later, non-plan-subject commit touching the same path wrongly won instead: $OUT" ;;
  *) pass "finding E: the later, non-plan-subject commit did not win" ;;
esac
case "$OUT" in
  *E-IMPL-BODY*) pass "finding E: the real implementation commit is still resolved as IMPL_SHA" ;;
  *) fail "finding E: the real implementation commit was not resolved: $OUT" ;;
esac

# ===========================================================================
# SECTION: pass 2, finding G — when finish run 1 skipped the implementation
# commit (nothing was staged outside the planning paths), the planning
# commit's parent is whatever preceded the whole sequence — here, a merge
# commit. Before the fix, the guard accepted any parent whose subject was
# not one of three reserved shapes, so a merge commit was reported as this
# change's own implementation commit: the "confident wrong answer" this
# heuristic exists to prevent. This case had ZERO prior coverage — deleting
# the whole refusal guard failed no test.
# ===========================================================================
new_repo
ORIG_BRANCH="$(cd "$REPO" && git symbolic-ref --short HEAD)"
(
  cd "$REPO" \
    && git checkout -q -b other-branch \
    && git commit -q --allow-empty -m "feat(other): unrelated work on a side branch" \
    && git checkout -q "$ORIG_BRANCH" \
    && git merge -q --no-ff -m "Merge pull request #42 from someone/some-other-change" other-branch \
    && mkdir -p spectre/changes/demo \
    && printf 'TASKS-G\n' > spectre/changes/demo/tasks.md \
    && git add spectre/changes/demo/tasks.md \
    && git commit -q -m "chore(spectre): plan and session records G-PLAN-BODY" \
    && git mv spectre/changes/demo "$REL" \
    && git commit -q -m "chore(spectre): archive demo" -m "G-ARCHIVE-BODY"
)
run_it
[ "$RC" -eq 0 ] && pass "finding G: exits 0" \
  || fail "finding G: rc=$RC out=$OUT"
case "$OUT" in
  *G-PLAN-BODY*) pass "finding G: the planning commit is still resolved as PLAN_SHA" ;;
  *) fail "finding G: the planning commit was not resolved: $OUT" ;;
esac
case "$OUT" in
  *"Merge pull request #42"*) \
    fail "finding G: the merge commit was wrongly reported as the implementation commit: $OUT" ;;
  *) pass "finding G: the merge commit was not reported as the implementation commit (IMPL_SHA resolves nothing)" ;;
esac

# ===========================================================================
# SECTION: pass 2, finding G — each acceptance condition pinned ALONE.
# The finding-G fixture above merges an EMPTY branch, so its merge commit is
# refused by the non-merge test AND by the outside-path test at the same
# time. That makes each condition redundant there: dropping either one alone
# still failed no test. Both mutants were found by the parent's own mutation
# pass and are repaired here, with one fixture that isolates each condition.
# ===========================================================================

# Isolates OUTSIDE-PATH: a NON-merge parent that touches only a planning
# path. The non-merge test passes it; only the outside-path test refuses it.
new_repo
(
  cd "$REPO" \
    && mkdir -p docs/superpowers/ledgers spectre/changes/demo \
    && printf 'unrelated\n' > docs/superpowers/ledgers/someone-else.md \
    && git add docs/superpowers/ledgers/someone-else.md \
    && git commit -q -m "chore(records): PLANNING-ONLY-PARENT-BODY" \
    && printf 'TASKS-OUTSIDE\n' > spectre/changes/demo/tasks.md \
    && git add spectre/changes/demo/tasks.md \
    && git commit -q -m "chore(spectre): plan and session records OUTSIDE-PLAN-BODY" \
    && git mv spectre/changes/demo "$REL" \
    && git commit -q -m "chore(spectre): archive demo"
)
run_it
case "$OUT" in
  *OUTSIDE-PLAN-BODY*) pass "condition isolation: planning commit resolved" ;;
  *) fail "condition isolation: planning commit not resolved: $OUT" ;;
esac
case "$OUT" in
  *PLANNING-ONLY-PARENT-BODY*) \
    fail "condition isolation: a non-merge parent touching only planning paths was wrongly accepted as IMPL_SHA: $OUT" ;;
  *) pass "condition isolation: a non-merge parent touching only planning paths is refused (outside-path test alone)" ;;
esac

# Isolates NON-MERGE: a merge commit that DOES touch a path outside the
# planning paths. The outside-path test would pass it; only the non-merge
# test refuses it.
new_repo
ISO_BRANCH="$(cd "$REPO" && git symbolic-ref --short HEAD)"
(
  cd "$REPO" \
    && git checkout -q -b iso-side \
    && mkdir -p scripts/iso \
    && printf 'y\n' > scripts/iso/bar.sh \
    && git add scripts/iso/bar.sh \
    && git commit -q -m "feat(iso): side work touching scripts" \
    && git checkout -q "$ISO_BRANCH" \
    && git merge -q --no-ff -m "Merge branch iso-side MERGING-PARENT-BODY" iso-side \
    && mkdir -p spectre/changes/demo \
    && printf 'TASKS-MERGE\n' > spectre/changes/demo/tasks.md \
    && git add spectre/changes/demo/tasks.md \
    && git commit -q -m "chore(spectre): plan and session records MERGE-PLAN-BODY" \
    && git mv spectre/changes/demo "$REL" \
    && git commit -q -m "chore(spectre): archive demo"
)
run_it
case "$OUT" in
  *MERGE-PLAN-BODY*) pass "condition isolation: planning commit resolved past a merging parent" ;;
  *) fail "condition isolation: planning commit not resolved past a merging parent: $OUT" ;;
esac
case "$OUT" in
  *MERGING-PARENT-BODY*) \
    fail "condition isolation: a merge commit touching outside paths was wrongly accepted as IMPL_SHA: $OUT" ;;
  *) pass "condition isolation: a merge commit touching outside paths is refused (non-merge test alone)" ;;
esac


# ===========================================================================
# SECTION: pass 2, finding G (positive control) — a genuine, non-merge
# implementation commit that touches a path outside spectre/ and
# docs/superpowers/ is still accepted as IMPL_SHA. Proves the new
# non-merge/outside-path heuristic does not just refuse everything.
# ===========================================================================
new_repo
(
  cd "$REPO" \
    && mkdir -p scripts/some-module spectre/changes/demo \
    && printf 'x\n' > scripts/some-module/foo.sh \
    && git add scripts/some-module/foo.sh \
    && git commit -q -m "feat(some-module): G-POSITIVE-IMPL-BODY" \
    && printf 'TASKS-G-POSITIVE\n' > spectre/changes/demo/tasks.md \
    && git add spectre/changes/demo/tasks.md \
    && git commit -q -m "chore(spectre): plan and session records G-POSITIVE-PLAN-BODY" \
    && git mv spectre/changes/demo "$REL" \
    && git commit -q -m "chore(spectre): archive demo" -m "G-POSITIVE-ARCHIVE-BODY"
)
run_it
[ "$RC" -eq 0 ] && pass "finding G positive control: exits 0" \
  || fail "finding G positive control: rc=$RC out=$OUT"
case "$OUT" in
  *G-POSITIVE-IMPL-BODY*) \
    pass "finding G positive control: a real non-merge, outside-path implementation commit is accepted as IMPL_SHA" ;;
  *) fail "finding G positive control: the real implementation commit was not resolved: $OUT" ;;
esac

# ===========================================================================
# SECTION: the planning commit's parent is the repository's own ROOT commit
# (this change's own review panel, pass 3, finding C). PARENT_TOUCHES_OUTSIDE
# was computed with `git diff --name-only "${PLAN_PARENT}^..${PLAN_PARENT}"`,
# which fails with "fatal: ambiguous argument" when PLAN_PARENT has no `^`
# parent to diff against (i.e. it IS the root commit) — and the `2>/dev/null`
# on that call swallowed the failure silently, leaving PARENT_TOUCHES_OUTSIDE
# at its default 0 and dropping a genuine implementation commit. Built as a
# bespoke two-commit repository rather than via new_repo() (which always
# commits its own "init" first, so the implementation commit would never be
# the actual root commit) — deliberately no "init" commit here.
# ===========================================================================

ROOT_REPO="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-root-repo.XXXXXX")"
TREES+=("$ROOT_REPO")
(
  cd "$ROOT_REPO" \
    && git init -q \
    && git config user.email test@example.com \
    && git config user.name test \
    && mkdir -p scripts/some-module spectre/changes/demo \
    && printf 'x\n' > scripts/some-module/foo.sh \
    && git add scripts/some-module/foo.sh \
    && git commit -q -m "feat(some-module): ROOT-COMMIT-IMPL-BODY" \
    && printf 'TASKS-ROOT\n' > spectre/changes/demo/tasks.md \
    && git add spectre/changes/demo/tasks.md \
    && git commit -q -m "chore(spectre): plan and session records ROOT-COMMIT-PLAN-BODY" \
    && mkdir -p spectre/changes/archive \
    && git mv spectre/changes/demo spectre/changes/archive/2026-01-01-demo \
    && git commit -q -m "chore(spectre): archive demo" -m "ROOT-COMMIT-ARCHIVE-BODY"
)
REPO="$ROOT_REPO"
ARCHIVED="$REPO/spectre/changes/archive/2026-01-01-demo"
STATE_DIR="$REPO/.state"
mkdir -p "$STATE_DIR"
run_it
[ "$RC" -eq 0 ] && pass "root commit as implementation commit: exits 0" \
  || fail "root commit as implementation commit: rc=$RC out=$OUT"
case "$OUT" in
  *ROOT-COMMIT-IMPL-BODY*) \
    pass "root commit as implementation commit: implementation commit resolved as IMPL_SHA despite having no parent to diff against" ;;
  *) fail "root commit as implementation commit: implementation commit NOT resolved — PARENT_TOUCHES_OUTSIDE likely silently stayed 0 because \"git diff A^..A\" fails on a root commit: $OUT" ;;
esac
case "$OUT" in
  *ROOT-COMMIT-PLAN-BODY*) pass "root commit as implementation commit: planning commit still resolved as PLAN_SHA" ;;
  *) fail "root commit as implementation commit: planning commit not resolved: $OUT" ;;
esac

# SECTION: <repo-root> — the optional fourth positional argument
# (Requirement: "Context is gathered deterministically, not by the model
# re-reading files", scenarios "The repo root is omitted", "The repo root is
# supplied" and "A malformed repo root is an invocation error", in this
# repository's archived kan-239-run-2-asserts-base-branch-and-archives-via-pr
# change's delta spec). One assertion per case, each folding its
# exit code and its content/emptiness check into a single pass/fail — no
# case here duplicates a check already made above this section.
# ===========================================================================

# run_bad_root <repo-root-arg> -> sets RC, STDOUT and STDERR separately,
# never merged, for a fourth argument expected to be rejected before any
# gathering starts. cwd is $REPO and $REL the archived-change-path in every
# case below — only the fourth argument varies. Streams are kept apart
# because "gathers nothing" means stdout is genuinely empty, not merely free
# of a marker string; a 2>&1 merge would hide that behind stderr's own text.
run_bad_root() {
  local repo_root_arg="$1" errfile
  errfile="$(mktemp "${TMPDIR:-/tmp}/gather-test-stderr.XXXXXX")"
  set +e
  STDOUT="$(cd "$REPO" && "$SCRIPT" "$REL" demo "$STATE_DIR" "$repo_root_arg" 2>"$errfile")"
  RC=$?
  set -e
  STDERR="$(cat "$errfile")"
  rm -f "$errfile"
}

# repo-root-omitted: three arguments, cwd inside the repository — unchanged
# behaviour, byte for byte, matching every case above this section that
# never passes a fourth argument.
new_repo
add_ledger
add_panel
add_tasks
add_commits
run_it
if [ "$RC" -eq 0 ] && case "$OUT" in *LEDGER-BODY*) true ;; *) false ;; esac; then
  pass "repo-root-omitted: three-argument invocation is unchanged"
else
  fail "repo-root-omitted: rc=$RC out=$OUT"
fi

# repo-root-supplied: cwd OUTSIDE $REPO, a fourth argument naming $REPO's own
# canonical (symlink-free) path. The only way to prove the argument is
# actually used: a cwd outside the repository has no --git-common-dir anchor
# to fall back on at all, so the bundle is gathered only if the fourth
# argument itself becomes the trusted repository root.
new_repo
add_ledger
add_panel
add_tasks
add_commits
REPO_REAL_ABS="$(cd -P "$REPO" && pwd -P)"
ARCHIVED_REAL_ABS="$REPO_REAL_ABS/spectre/changes/archive/2026-01-01-demo"
OUTSIDE_CWD="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-outside-cwd.XXXXXX")"
TREES+=("$OUTSIDE_CWD")
set +e
OUT="$(cd "$OUTSIDE_CWD" && "$SCRIPT" "$ARCHIVED_REAL_ABS" demo "$STATE_DIR" "$REPO_REAL_ABS" 2>&1)"
RC=$?
set -e
if [ "$RC" -eq 0 ] && case "$OUT" in *LEDGER-BODY*) true ;; *) false ;; esac; then
  pass "repo-root-supplied: bundle gathered against the supplied root from outside the repository"
else
  fail "repo-root-supplied: rc=$RC out=$OUT"
fi

# repo-root-supplied-relative-archived-path (F3, this change's own review
# panel): cwd OUTSIDE $REPO, a fourth argument naming $REPO's own canonical
# root, and $REL — a RELATIVE archived-change-path, the exact shape
# skills/myflow-finish/SKILL.md documents — rather than the absolute path
# the "repo-root-supplied" case above uses. Step 2 (lexical normalize) joins
# a relative $ARCHIVED_PATH onto $trusted_root (the supplied root here), so
# step 4's semantic resolution of the SAME relative $ARCHIVED_PATH must also
# resolve against $trusted_root, not against the actual process cwd
# ($OUTSIDE_CWD, which has no such directory at all) — otherwise the two
# resolutions diverge, ARCHIVED_REAL comes back empty, and every source is
# wrongly reported skipped even though the repo-root argument was honored.
new_repo
add_ledger
add_panel
add_tasks
add_commits
REPO_REAL_ABS="$(cd -P "$REPO" && pwd -P)"
OUTSIDE_CWD_REL="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-outside-cwd-rel.XXXXXX")"
TREES+=("$OUTSIDE_CWD_REL")
set +e
OUT="$(cd "$OUTSIDE_CWD_REL" && "$SCRIPT" "$REL" demo "$STATE_DIR" "$REPO_REAL_ABS" 2>&1)"
RC=$?
set -e
if [ "$RC" -eq 0 ] && case "$OUT" in *LEDGER-BODY*) true ;; *) false ;; esac; then
  pass "repo-root-supplied-relative-archived-path: bundle gathered (relative archived-change-path resolved against the supplied root)"
else
  fail "repo-root-supplied-relative-archived-path: rc=$RC out=$OUT"
fi

# repo-root-supplied-relative-archived-path-symlink: the exact same
# override shape as the case above (cwd outside the repo, a relative
# archived-change-path, a supplied repo-root), but the archived directory is
# a symlink to somewhere outside the repository (the F12 attack). The
# symlink defence must still catch this when step 4 is resolving against
# $trusted_root instead of raw process cwd — proving the fix did not weaken
# the lexical-vs-real divergence check.
new_repo
rm -rf "$ARCHIVED"
OUTSIDE_DIR_ROOT_OVERRIDE="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-outside-dir-root-override.XXXXXX")"
TREES+=("$OUTSIDE_DIR_ROOT_OVERRIDE")
printf 'TOP-SECRET-ROOT-OVERRIDE-CONTENT\n' > "$OUTSIDE_DIR_ROOT_OVERRIDE/tasks.md"
ln -s "$OUTSIDE_DIR_ROOT_OVERRIDE" "$ARCHIVED"
add_commits
REPO_REAL_ABS="$(cd -P "$REPO" && pwd -P)"
OUTSIDE_CWD_REL_SYMLINK="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-outside-cwd-rel-symlink.XXXXXX")"
TREES+=("$OUTSIDE_CWD_REL_SYMLINK")
set +e
OUT="$(cd "$OUTSIDE_CWD_REL_SYMLINK" && "$SCRIPT" "$REL" demo "$STATE_DIR" "$REPO_REAL_ABS" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] && pass "repo-root-supplied-relative-archived-path-symlink: exits 0" \
  || fail "repo-root-supplied-relative-archived-path-symlink: rc=$RC out=$OUT"
case "$OUT" in
  *"note: archived-change-path '$REL' resolves through a symlink somewhere between the repository root and the leaf"*) \
    pass "repo-root-supplied-relative-archived-path-symlink: distinct note printed" ;;
  *) fail "repo-root-supplied-relative-archived-path-symlink: distinct note not printed: $OUT" ;;
esac
case "$OUT" in
  *TOP-SECRET-ROOT-OVERRIDE-CONTENT*) \
    fail "repo-root-supplied-relative-archived-path-symlink: target tasks.md content leaked into bundle: $OUT" ;;
  *) pass "repo-root-supplied-relative-archived-path-symlink: target tasks.md content not in bundle" ;;
esac

# repo-root-relative: a relative fourth argument is an invocation error.
new_repo
add_ledger
add_tasks
run_bad_root "relative-root"
if [ "$RC" -eq 2 ] && [ -z "$STDOUT" ]; then
  pass "repo-root-relative: exits 2 with nothing gathered"
else
  fail "repo-root-relative: rc=$RC stdout=$STDOUT stderr=$STDERR"
fi

# repo-root-symlinked: a fourth argument reaching through a symlink is an
# invocation error — the same lexical/real divergence test already applied
# to <archived-change-path>.
new_repo
add_ledger
add_tasks
REPO_REAL_ABS="$(cd -P "$REPO" && pwd -P)"
SYMLINKED_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-symlinked-root.XXXXXX")"
rmdir "$SYMLINKED_ROOT"
ln -s "$REPO_REAL_ABS" "$SYMLINKED_ROOT"
TREES+=("$SYMLINKED_ROOT")
run_bad_root "$SYMLINKED_ROOT"
if [ "$RC" -eq 2 ] && [ -z "$STDOUT" ]; then
  pass "repo-root-symlinked: exits 2 with nothing gathered"
else
  fail "repo-root-symlinked: rc=$RC stdout=$STDOUT stderr=$STDERR"
fi

# repo-root-not-a-repo: an absolute, symlink-free directory that is not a
# git repository root is an invocation error.
new_repo
add_ledger
add_tasks
NOT_A_REPO="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-not-a-repo.XXXXXX")"
TREES+=("$NOT_A_REPO")
NOT_A_REPO_REAL="$(cd -P "$NOT_A_REPO" && pwd -P)"
run_bad_root "$NOT_A_REPO_REAL"
if [ "$RC" -eq 2 ] && [ -z "$STDOUT" ]; then
  pass "repo-root-not-a-repo: exits 2 with nothing gathered"
else
  fail "repo-root-not-a-repo: rc=$RC stdout=$STDOUT stderr=$STDERR"
fi

# ===========================================================================
# SECTION: Invocation contract — a bad change name is rejected before any
# path is touched
# ===========================================================================

new_repo
for badname in '../escape' 'a/b' '/abs' '*' 'a*b'; do
  set +e
  OUT="$("$SCRIPT" "$ARCHIVED" "$badname" "$STATE_DIR" 2>&1)"
  RC=$?
  set -e
  [ "$RC" -eq 2 ] && pass "change name '$badname' exits 2" \
    || fail "change name '$badname': expected exit 2, rc=$RC out=$OUT"
  case "$OUT" in
    *"gather-self-review-context:"*) pass "change name '$badname' is named in the error" ;;
    *) fail "change name '$badname': no named error: $OUT" ;;
  esac
done

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'gather-self-review-context: all cases pass\n'
