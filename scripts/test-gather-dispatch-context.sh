#!/usr/bin/env bash
# Assertion harness for gather-dispatch-context.sh. Builds a sandboxed git
# repository per case under TMPDIR; never touches the real repository tree.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated
# contract — the Requirements and Scenarios in this repository's archived
# kan-201-reduce-context-rediscovery-across-review-panel change's
# myflow-dispatch-economy spec — never against whatever the script happens to
# print.
# scripts/test-check-plan-provenance.sh's header records that suite encoding
# the guard's own defects as its specification more than once, which then
# made each defect look verified; the same mistake is possible here.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT="$SCRIPT_DIR/gather-dispatch-context.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# Every case leaves one sandboxed tree behind. Removed on exit, including on
# a failed assertion. An indexed array, not a space-separated string: sandbox
# paths come from mktemp under TMPDIR, which may contain spaces, and
# word-splitting a string would leak every sandbox whose path split.
TREES=()
cleanup() {
  [ "${#TREES[@]}" -eq 0 ] && return 0
  for tree in "${TREES[@]}"; do
    chmod -R u+w "$tree" 2>/dev/null || true
    rm -rf "$tree"
  done
}
trap cleanup EXIT

# new_repo -> sets REPO (a real git repository), CHANGE_ROOT (a fully
# populated change directory under it) and PRINCIPLES (a principles file).
# The fixture shape tasks.md records: a proposal, a design, a plan and a
# principles file. There is no specs/ leaf: under spectre a change has no
# specs/ subdirectory, and the script no longer reads one.
#
# REPO is canonicalized (`cd -P` / `pwd -P`) immediately after creation —
# the same step gather-self-review-context.sh's own F23 test fixture takes
# — because TMPDIR itself sits behind an OS-level alias on macOS
# (/var/folders/... -> /private/var/folders/...). Without this, comparing
# REPO's raw mktemp path (lexical) against its cd -P resolution (real) would
# report a symlink divergence that has nothing to do with the change-root
# or principles-path validation this suite is actually exercising.
new_repo() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/gather-dispatch-test-repo.XXXXXX")"
  REPO="$(cd -P "$REPO" && pwd -P)"
  TREES+=("$REPO")
  (
    cd "$REPO" \
      && git init -q \
      && git config user.email test@example.com \
      && git config user.name test \
      && git commit -q --allow-empty -m "init"
  )
  CHANGE_ROOT="$REPO/spectre/changes/demo"
  mkdir -p "$CHANGE_ROOT"
  printf 'PROPOSAL-BODY\n' > "$CHANGE_ROOT/proposal.md"
  printf 'DESIGN-BODY\n' > "$CHANGE_ROOT/design.md"
  printf 'TASKS-BODY\n' > "$CHANGE_ROOT/tasks.md"
  PRINCIPLES="$REPO/PRINCIPLES.md"
  printf 'PRINCIPLES-BODY\n' > "$PRINCIPLES"
}

# run_it -> sets RC and OUT, invoking with the standard four arguments.
run_it() {
  set +e
  OUT="$("$SCRIPT" "$REPO" "$CHANGE_ROOT" demo "$PRINCIPLES" 2>&1)"
  RC=$?
  set -e
}

# ===========================================================================
# CASE 1: all five sources present
# ===========================================================================

new_repo
run_it
[ "$RC" -eq 0 ] && pass "all sources present: exits 0" \
  || fail "all sources present: rc=$RC out=$OUT"
case "$OUT" in
  *PROPOSAL-BODY*) pass "all sources present: proposal content in bundle" ;;
  *) fail "all sources present: proposal content missing: $OUT" ;;
esac
case "$OUT" in
  *DESIGN-BODY*) pass "all sources present: design content in bundle" ;;
  *) fail "all sources present: design content missing: $OUT" ;;
esac
case "$OUT" in
  *TASKS-BODY*) pass "all sources present: tasks content in bundle" ;;
  *) fail "all sources present: tasks content missing: $OUT" ;;
esac
case "$OUT" in
  *PRINCIPLES-BODY*) pass "all sources present: principles content in bundle" ;;
  *) fail "all sources present: principles content missing: $OUT" ;;
esac

# ===========================================================================
# CASE 2: design.md absent
# ===========================================================================

new_repo
rm -f "$CHANGE_ROOT/design.md"
run_it
[ "$RC" -eq 0 ] && pass "design absent: exits 0" \
  || fail "design absent: rc=$RC out=$OUT"
case "$OUT" in
  *"skipped: design.md (absent)"*) pass "design absent: reported as skipped" ;;
  *) fail "design absent: not reported as skipped: $OUT" ;;
esac
case "$OUT" in
  *PROPOSAL-BODY*) pass "design absent: proposal still present" ;;
  *) fail "design absent: proposal missing: $OUT" ;;
esac

# ===========================================================================
# CASE 3: every source absent
# ===========================================================================

new_repo
rm -f "$CHANGE_ROOT/proposal.md" "$CHANGE_ROOT/design.md" "$CHANGE_ROOT/tasks.md"
rm -f "$PRINCIPLES"
run_it
[ "$RC" -eq 0 ] && pass "every source absent: exits 0" \
  || fail "every source absent: rc=$RC out=$OUT"
for label in "proposal.md" "design.md" "tasks.md"; do
  case "$OUT" in
    *"skipped: $label (absent)"*) pass "every source absent: $label reported skipped" ;;
    *) fail "every source absent: $label not reported skipped: $OUT" ;;
  esac
done
case "$OUT" in
  *"skipped: $PRINCIPLES (absent)"*) pass "every source absent: principles reported skipped" ;;
  *) fail "every source absent: principles not reported skipped: $OUT" ;;
esac

# CASES 4 AND 5 ARE GONE, not renumbered: they covered two delta specs and an
# absent specs/ directory, and the delta-spec source they exercised was removed
# with the spectre cutover. The numbering below is left as it was so a case's
# name still means the same case it always did.

# ===========================================================================
# CASE 6: header carries a generated instant
# ===========================================================================

new_repo
run_it
case "$OUT" in
  *"generated: "[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]:[0-9][0-9]Z*) \
    pass "header carries a generated instant" ;;
  *) fail "header missing a generated instant: $OUT" ;;
esac

# ===========================================================================
# CASE 7: header carries the HEAD sha
# ===========================================================================

new_repo
EXPECT_SHA="$(git -C "$REPO" rev-parse --short HEAD)"
run_it
case "$OUT" in
  *"head: $EXPECT_SHA"*) pass "header carries the HEAD sha" ;;
  *) fail "header missing the HEAD sha ($EXPECT_SHA): $OUT" ;;
esac

# ===========================================================================
# CASE 8: a standards-shaped file in the change root must not leak into the
# bundle. The script globs nothing anywhere under change-root — it reads three
# leaves by exact name — so a `## standards` file dropped beside them is never
# read. The Requirement forbidding it is unchanged by the spectre cutover; only
# where such a file could plausibly be dropped changed, since there is no
# specs/ directory to drop it in any more.
# ===========================================================================

new_repo
printf 'STANDARDS-MARKER-CONTENT\n' > "$CHANGE_ROOT/standards.md"
run_it
[ "$RC" -eq 0 ] && pass "standards-shaped file present: exits 0" \
  || fail "standards-shaped file present: rc=$RC out=$OUT"
case "$OUT" in
  *STANDARDS-MARKER-CONTENT*) fail "standards-shaped file content leaked into bundle: $OUT" ;;
  *) pass "standards-shaped file content not in bundle" ;;
esac

# ===========================================================================
# CASE 9: missing fourth argument
# ===========================================================================

new_repo
set +e
OUT="$("$SCRIPT" "$REPO" "$CHANGE_ROOT" demo 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "missing fourth argument: exits 2" \
  || fail "missing fourth argument: expected exit 2, rc=$RC out=$OUT"

# ===========================================================================
# CASE 10: change name containing '/'
# ===========================================================================

new_repo
set +e
OUT="$("$SCRIPT" "$REPO" "$CHANGE_ROOT" "a/b" "$PRINCIPLES" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "change name with slash: exits 2" \
  || fail "change name with slash: expected exit 2, rc=$RC out=$OUT"
case "$OUT" in
  *"gather-dispatch-context:"*) pass "change name with slash: error is named" ;;
  *) fail "change name with slash: no named error: $OUT" ;;
esac

# ===========================================================================
# CASE 11: change name containing a glob metacharacter
# ===========================================================================

new_repo
set +e
OUT="$("$SCRIPT" "$REPO" "$CHANGE_ROOT" "a*b" "$PRINCIPLES" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "change name with glob metacharacter: exits 2" \
  || fail "change name with glob metacharacter: expected exit 2, rc=$RC out=$OUT"

# ===========================================================================
# CASE 12: <change-root> reached through a symlinked ancestor
# ===========================================================================

new_repo
OUTSIDE_12="$(mktemp -d "${TMPDIR:-/tmp}/gather-dispatch-test-outside12.XXXXXX")"
TREES+=("$OUTSIDE_12")
mkdir -p "$OUTSIDE_12/changes/demo"
printf 'TOP-SECRET-12\n' > "$OUTSIDE_12/changes/demo/tasks.md"
mkdir -p "$REPO/spectre-symlink"
rm -rf "$REPO/spectre-symlink"
ln -s "$OUTSIDE_12" "$REPO/spectre-symlink"
SYMLINKED_CHANGE_ROOT="$REPO/spectre-symlink/changes/demo"
set +e
OUT="$("$SCRIPT" "$REPO" "$SYMLINKED_CHANGE_ROOT" demo "$PRINCIPLES" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "change-root through symlinked ancestor: exits 2" \
  || fail "change-root through symlinked ancestor: expected exit 2, rc=$RC out=$OUT"
case "$OUT" in
  *TOP-SECRET-12*) fail "change-root through symlinked ancestor: target content leaked: $OUT" ;;
  *) pass "change-root through symlinked ancestor: target content not in output" ;;
esac

# ===========================================================================
# CASE 13: <principles-path> that is itself a symlink — the shape a global
# install always produces. F1: this must be resolved and READ, never
# refused for being a symlink — principles-path is a trusted, caller-
# resolved argument, not attacker-influenced content the way <change-root>'s
# leaves are, so a genuine symlink divergence on it is no longer a reason to
# refuse the whole invocation.
# ===========================================================================

new_repo
OUTSIDE_13="$(mktemp -d "${TMPDIR:-/tmp}/gather-dispatch-test-outside13.XXXXXX")"
TREES+=("$OUTSIDE_13")
printf 'REAL-PRINCIPLES-BODY\n' > "$OUTSIDE_13/real-principles.md"
rm -f "$PRINCIPLES"
ln -s "$OUTSIDE_13/real-principles.md" "$PRINCIPLES"
run_it
[ "$RC" -eq 0 ] && pass "principles-path is a symlink: exits 0" \
  || fail "principles-path is a symlink: rc=$RC out=$OUT"
case "$OUT" in
  *REAL-PRINCIPLES-BODY*) pass "principles-path is a symlink: target content read" ;;
  *) fail "principles-path is a symlink: target content missing: $OUT" ;;
esac
case "$OUT" in
  *"skipped: $PRINCIPLES"*) fail "principles-path is a symlink: wrongly reported skipped: $OUT" ;;
  *) pass "principles-path is a symlink: not reported skipped" ;;
esac

# ===========================================================================
# CASE 14: <principles-path> reached through a symlinked ANCESTOR directory
# — the literal global-install shape: `setup.sh global` symlinks
# ~/.claude/skills/myflow-do to the repository's skills/myflow-do/, so
# engineering-principles.md sits inside a symlinked directory without being
# a symlink itself. Same disposition as case 13: resolved and read.
# ===========================================================================

new_repo
REAL_SKILL_DIR_14="$(mktemp -d "${TMPDIR:-/tmp}/gather-dispatch-test-realskill14.XXXXXX")"
TREES+=("$REAL_SKILL_DIR_14")
printf 'ANCESTOR-PRINCIPLES-BODY\n' > "$REAL_SKILL_DIR_14/engineering-principles.md"
SYMLINKED_SKILL_DIR_14="$(mktemp -u "${TMPDIR:-/tmp}/gather-dispatch-test-symlinkskill14.XXXXXX")"
ln -s "$REAL_SKILL_DIR_14" "$SYMLINKED_SKILL_DIR_14"
TREES+=("$SYMLINKED_SKILL_DIR_14")
ANCESTOR_PRINCIPLES_14="$SYMLINKED_SKILL_DIR_14/engineering-principles.md"
set +e
OUT="$("$SCRIPT" "$REPO" "$CHANGE_ROOT" demo "$ANCESTOR_PRINCIPLES_14" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] && pass "principles-path through symlinked ancestor: exits 0" \
  || fail "principles-path through symlinked ancestor: rc=$RC out=$OUT"
case "$OUT" in
  *ANCESTOR-PRINCIPLES-BODY*) pass "principles-path through symlinked ancestor: content read" ;;
  *) fail "principles-path through symlinked ancestor: content missing: $OUT" ;;
esac

# ===========================================================================
# CASE 15: <change-root> outside the repository (worktree)
# ===========================================================================

new_repo
OUTSIDE_14="$(mktemp -d "${TMPDIR:-/tmp}/gather-dispatch-test-outside14.XXXXXX")"
TREES+=("$OUTSIDE_14")
mkdir -p "$OUTSIDE_14/demo"
printf 'PROPOSAL-BODY\n' > "$OUTSIDE_14/demo/proposal.md"
set +e
OUT="$("$SCRIPT" "$REPO" "$OUTSIDE_14/demo" demo "$PRINCIPLES" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "change-root outside the worktree: exits 2" \
  || fail "change-root outside the worktree: expected exit 2, rc=$RC out=$OUT"

# ===========================================================================
# CASE 16: proposal.md is a symlink resolving outside change-root — a leaf
# symlink, never an ancestor. Must be refused per-source, content omitted
# from the bundle, and the run must still exit 0 (F1: the five content
# sources were previously only `[ -f ... ]`-tested, never boundary-checked).
# ===========================================================================

new_repo
OUTSIDE_15="$(mktemp -d "${TMPDIR:-/tmp}/gather-dispatch-test-outside15.XXXXXX")"
TREES+=("$OUTSIDE_15")
printf 'TOP-SECRET-15\n' > "$OUTSIDE_15/secret.md"
rm -f "$CHANGE_ROOT/proposal.md"
ln -s "$OUTSIDE_15/secret.md" "$CHANGE_ROOT/proposal.md"
run_it
[ "$RC" -eq 0 ] && pass "symlinked proposal.md outside change-root: exits 0" \
  || fail "symlinked proposal.md outside change-root: rc=$RC out=$OUT"
case "$OUT" in
  *TOP-SECRET-15*) fail "symlinked proposal.md outside change-root: target content leaked: $OUT" ;;
  *) pass "symlinked proposal.md outside change-root: target content not in bundle" ;;
esac
case "$OUT" in
  *"refused: proposal.md (resolves outside the change directory)"*) \
    pass "symlinked proposal.md outside change-root: reported refused" ;;
  *) fail "symlinked proposal.md outside change-root: not reported refused: $OUT" ;;
esac

# CASE 17 IS GONE, not renumbered: it pinned the refusal disposition for a
# symlinked delta spec, and there is no delta-spec source left to symlink.
# Case 15 above pins the same disposition on proposal.md, which is the source
# that still exists.

# ===========================================================================
# CASE 18: design.md is a symlink resolving to another file INSIDE
# change-root (not outside it). Pins that an in-tree leaf symlink is not
# over-refused by the same mechanism that refuses cases 15/16.
# ===========================================================================

new_repo
rm -f "$CHANGE_ROOT/design.md"
ln -s "$CHANGE_ROOT/tasks.md" "$CHANGE_ROOT/design.md"
run_it
[ "$RC" -eq 0 ] && pass "symlinked design.md resolving in-tree: exits 0" \
  || fail "symlinked design.md resolving in-tree: rc=$RC out=$OUT"
case "$OUT" in
  *"## design.md"*) pass "symlinked design.md resolving in-tree: heading present" ;;
  *) fail "symlinked design.md resolving in-tree: heading missing: $OUT" ;;
esac
case "$OUT" in
  *TASKS-BODY*) pass "symlinked design.md resolving in-tree: target content present" ;;
  *) fail "symlinked design.md resolving in-tree: target content missing: $OUT" ;;
esac
case "$OUT" in
  *"refused: design.md"*) fail "symlinked design.md resolving in-tree: wrongly refused: $OUT" ;;
  *) pass "symlinked design.md resolving in-tree: not refused (in-tree symlinks allowed)" ;;
esac

# ===========================================================================
# CASE 19: <worktree> passed with a trailing slash — no symlink involved.
# F2: resolve_file() (lib/resolve-file.sh:56) reduced a trailing-slash path
# to an empty basename and emitted a spurious trailing "/", a form
# lexical_normalize() never produces, so a plain trailing slash was wrongly
# refused as a symlink divergence.
# ===========================================================================

new_repo
set +e
OUT="$("$SCRIPT" "$REPO/" "$CHANGE_ROOT" demo "$PRINCIPLES" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] && pass "worktree with trailing slash: exits 0" \
  || fail "worktree with trailing slash: rc=$RC out=$OUT"
case "$OUT" in
  *PROPOSAL-BODY*) pass "worktree with trailing slash: proposal content present" ;;
  *) fail "worktree with trailing slash: proposal content missing: $OUT" ;;
esac

# ===========================================================================
# CASE 20: <worktree> passed as a relative path (".", invoked with cwd
# inside the worktree) — no symlink involved. Same root cause as case 19:
# resolve_file(".") appended a spurious "." basename that
# lexical_normalize() never produces.
# ===========================================================================

new_repo
set +e
OUT="$(cd "$REPO" && "$SCRIPT" "." "$CHANGE_ROOT" demo "$PRINCIPLES" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] && pass "worktree as relative path '.': exits 0" \
  || fail "worktree as relative path '.': rc=$RC out=$OUT"
case "$OUT" in
  *PROPOSAL-BODY*) pass "worktree as relative path '.': proposal content present" ;;
  *) fail "worktree as relative path '.': proposal content missing: $OUT" ;;
esac


# ===========================================================================
# CASE 21: proposal.md is a symlink cycle (F36, pass 7 of this change's own
# review panel). This is NOT a test of the "unresolvable" refusal path --
# a genuine cycle fails `[ -f "$path" ]` itself (proven directly: bash's -f
# follows symlinks to a real regular file or reports false, and a cycle
# reports false), so add_fixed_source's own absence check routes it to
# "skipped", never as far as resolve_file() at all. What this case pins is
# exactly that routing: the fix must not have disturbed the ordinary
# missing/dangling-symlink disposition while adding the unresolvable-vs-
# outside distinction. (resolve_file() reaching REFUSED_REASONS=unresolvable
# in production requires a resolve_file()-internal failure -- e.g. a
# component vanishing between this check and resolve_file()'s own walk --
# that "[ -f ]" already ruled out here and that a synchronous, single-
# process test cannot force deterministically; see this script's own
# REFUSED_REASONS doc comment.)
# ===========================================================================

new_repo
rm -f "$CHANGE_ROOT/proposal.md"
ln -s "$CHANGE_ROOT/proposal-loop-b.md" "$CHANGE_ROOT/proposal.md"
ln -s "$CHANGE_ROOT/proposal.md" "$CHANGE_ROOT/proposal-loop-b.md"
run_it
[ "$RC" -eq 0 ] && pass "symlink-cycle proposal.md: exits 0" \
  || fail "symlink-cycle proposal.md: rc=$RC out=$OUT"
case "$OUT" in
  *TOP-SECRET*) fail "symlink-cycle proposal.md: unexpected leaked content: $OUT" ;;
  *) pass "symlink-cycle proposal.md: no content leaked" ;;
esac
case "$OUT" in
  *"skipped: proposal.md (absent)"*) \
    pass "symlink-cycle proposal.md: reported skipped (absent), not refused" ;;
  *) fail "symlink-cycle proposal.md: not reported skipped: $OUT" ;;
esac
case "$OUT" in
  *"refused: proposal.md"*) \
    fail "symlink-cycle proposal.md: wrongly reported as refused: $OUT" ;;
  *) pass "symlink-cycle proposal.md: not misreported as refused" ;;
esac

# ===========================================================================
# CASE 22: .flow/project.md declares ## lint and ## test sections -> the
# bundle carries a "## project commands" heading whose body reproduces both
# sections' fenced command blocks verbatim.
# ===========================================================================

new_repo
mkdir -p "$REPO/.flow"
cat > "$REPO/.flow/project.md" <<'EOF'
# Project configuration

## lint

```bash
scripts/check-lint-marker.sh
```

## test

```bash
scripts/check-test-marker.sh
```
EOF
run_it
if [ "$RC" -eq 0 ] && case "$OUT" in
  *"## project commands"*"scripts/check-lint-marker.sh"*"scripts/check-test-marker.sh"*) true ;;
  *) false ;;
esac; then
  pass "project.md with lint and test: bundle reproduces both sections' fenced blocks verbatim"
else
  fail "project.md with lint and test: rc=$RC out=$OUT"
fi

# ===========================================================================
# CASE 23: .flow/project.md exists but declares none of ## lint / ## test /
# ## run -> reported as skipped, the same shape as an absent source.
# ===========================================================================

new_repo
mkdir -p "$REPO/.flow"
cat > "$REPO/.flow/project.md" <<'EOF'
# Project configuration

## apps

Nothing here concerns lint, test or run.
EOF
run_it
if [ "$RC" -eq 0 ] && case "$OUT" in
  *"skipped: project commands (absent)"*) true ;;
  *) false ;;
esac && case "$OUT" in
  *"## project commands"*) false ;;
  *) true ;;
esac; then
  pass "project.md with no lint/test/run: reported skipped, no project commands heading"
else
  fail "project.md with no lint/test/run: rc=$RC out=$OUT"
fi

# ===========================================================================
# CASE 24: no .flow/project.md at all -> reported as skipped, not a script
# error.
# ===========================================================================

new_repo
run_it
if [ "$RC" -eq 0 ] && case "$OUT" in
  *"skipped: project commands (absent)"*) true ;;
  *) false ;;
esac; then
  pass "no project.md: reported skipped, not a script error"
else
  fail "no project.md: rc=$RC out=$OUT"
fi

# ===========================================================================
# CASE 25: .flow/project.md's ## lint section stops at the next top-level
# "## " heading (## test) rather than swallowing it. Regression test for a
# mutated section-boundary exit condition (/^## / -> /^### /) that would
# make the lint section run to EOF and absorb the test section's content.
# ===========================================================================

new_repo
mkdir -p "$REPO/.flow"
cat > "$REPO/.flow/project.md" <<'EOF'
# Project configuration

## lint

LINT-ONLY-MARKER

## test

TEST-ONLY-MARKER
EOF
run_it
LINT_BLOCK="$(printf '%s\n' "$OUT" | sed -n '/^### lint$/,/^### test$/p')"
if [ "$RC" -eq 0 ] && case "$OUT" in
  *TEST-ONLY-MARKER*) true ;;
  *) false ;;
esac && case "$LINT_BLOCK" in
  *TEST-ONLY-MARKER*) false ;;
  *) true ;;
esac; then
  pass "project.md lint section stops at next ## heading: test content not swallowed into lint"
else
  fail "project.md lint section stops at next ## heading: rc=$RC out=$OUT"
fi

# ===========================================================================
# CASE 26: .flow/project.md's ## lint section body is whitespace-only ->
# treated as absent, same as a missing section. Regression test for a
# weakened emptiness check ([ -n "$section" ] instead of stripping whitespace
# first) that would wrongly emit a populated but blank "### lint" heading.
# ===========================================================================

new_repo
mkdir -p "$REPO/.flow"
printf '# Project configuration\n\n## lint\n   \t   \n\n## test\n\nscripts/check-test-marker.sh\n' \
  > "$REPO/.flow/project.md"
run_it
if [ "$RC" -eq 0 ] && case "$OUT" in
  *"### test"*) true ;;
  *) false ;;
esac && case "$OUT" in
  *"### lint"*) false ;;
  *) true ;;
esac; then
  pass "project.md whitespace-only lint section: treated as absent, no ### lint heading"
else
  fail "project.md whitespace-only lint section: rc=$RC out=$OUT"
fi

# ===========================================================================
# CASE 27: .flow/project.md has a "## linter" heading (a superstring of
# "## lint") but no exact "## lint" heading -> not captured as the lint
# section. Regression test for a weakened key match ($0 == key -> $0 ~ key)
# that would let "## linter" match the "lint" key as a substring/regex.
# ===========================================================================

new_repo
mkdir -p "$REPO/.flow"
cat > "$REPO/.flow/project.md" <<'EOF'
# Project configuration

## linter

scripts/should-not-be-captured.sh
EOF
run_it
if [ "$RC" -eq 0 ] && case "$OUT" in
  *"skipped: project commands (absent)"*) true ;;
  *) false ;;
esac && case "$OUT" in
  *"scripts/should-not-be-captured.sh"*) false ;;
  *) true ;;
esac; then
  pass "project.md '## linter' heading: not captured as the lint section"
else
  fail "project.md '## linter' heading: rc=$RC out=$OUT"
fi

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'gather-dispatch-context: all cases pass\n'
