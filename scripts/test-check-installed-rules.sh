#!/usr/bin/env bash
# Assertion harness for check-installed-rules.sh.
#
# Builds throwaway install trees under a sandboxed TMPDIR and points the guard
# at each with CHECK_INSTALLED_RULES_HOME — never touches the invoking user's
# real ~/.claude, and never leaves a fixture behind.
#
# WHY THIS FILE EXISTS, and what makes it evidence. Every case below breaks
# one property of a KNOWN-GOOD install and asserts the guard fails on exactly
# that break. A harness that only asserted the clean tree passes would be
# satisfied by `exit 0`, which is the guard this repository already had for
# the installed rule set: none. Case `missing_link` is the KAN-202 defect
# itself, reproduced — a rule present in rules/ and listed in the baseline
# table, with no link in the install and no marker in the managed block.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
GUARD="$SCRIPT_DIR/check-installed-rules.sh"
RULES_SRC="$REPO_ROOT/rules"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

DIRS=()
cleanup() {
  [ "${#DIRS[@]}" -eq 0 ] && return 0
  local d
  for d in "${DIRS[@]}"; do rm -rf "$d"; done
}
trap cleanup EXIT

# always_on -> the always-on rule basenames, taken from the guard itself so
# this harness cannot disagree with it about which rules are in scope.
always_on() {
  CHECK_INSTALLED_RULES_HOME="/nonexistent-$$" "$GUARD" >/dev/null 2>&1 || true
  local f
  for f in "$RULES_SRC"/*.mdc; do
    [ -f "$f" ] || continue
    if awk '
      { sub(/\r$/, "") }
      NR == 1 && $0 != "---" { exit }
      NR > 1  && $0 == "---" { closed = 1; exit }
      NR > 20 { exit }
      NR > 1  && /^alwaysApply:[[:space:]]*true[[:space:]]*$/ { always = 1 }
      END { exit !(always && closed) }
    ' "$f"; then basename "$f"; fi
  done
}

mapfile -t RULES < <(always_on)
[ "${#RULES[@]}" -gt 1 ] || { echo "harness: need at least two always-on rules to test with" >&2; exit 1; }
FIRST="${RULES[0]}"

# new_home -> sets HOME_DIR to a complete, known-good fixture install.
new_home() {
  HOME_DIR="$(mktemp -d "${TMPDIR:-/tmp}/check-installed-rules-test.XXXXXX")"
  DIRS+=("$HOME_DIR")
  mkdir -p "$HOME_DIR/.claude/rules" "$HOME_DIR/.codex"
  local r
  for r in "${RULES[@]}"; do
    ln -s "$RULES_SRC/$r" "$HOME_DIR/.claude/rules/${r%.mdc}.md"
  done
  ln -s "$RULES_SRC/agent-baseline.md" "$HOME_DIR/.claude/rules/agent-baseline.md"
  render_block "$HOME_DIR/.claude/CLAUDE.md" "${RULES[@]}"
  render_block "$HOME_DIR/.codex/AGENTS.md" "${RULES[@]}"
}

# render_block <file> <rule>... -> a managed block carrying one marker per rule.
render_block() {
  local out="$1"; shift
  { echo "# fixture"; local r; for r in "$@"; do printf '\n<!-- rule: %s -->\n\nbody\n' "$r"; done; } > "$out"
}

# run_guard -> sets RC and OUT for the current HOME_DIR.
run_guard() {
  set +e
  OUT="$(CHECK_INSTALLED_RULES_HOME="$HOME_DIR" "$GUARD" 2>&1)"
  RC=$?
  set -e
}

expect_stale() {
  local name="$1" needle="$2"
  run_guard
  if [ "$RC" -eq 0 ]; then
    fail "$name: guard passed a broken install"
  elif ! printf '%s' "$OUT" | grep -qF "$needle"; then
    fail "$name: failed, but no message naming '$needle'; got: $OUT"
  else
    pass "$name"
  fi
}

# --- baseline: a complete install passes ------------------------------------
new_home
run_guard
if [ "$RC" -ne 0 ]; then
  fail "clean: a complete install was reported stale: $OUT"
else
  pass "clean: a complete install passes"
fi

# --- rule 1: the KAN-202 defect --------------------------------------------
new_home
rm "$HOME_DIR/.claude/rules/${FIRST%.mdc}.md"
sed -i.bak "/<!-- rule: $FIRST -->/d" "$HOME_DIR/.claude/CLAUDE.md"
sed -i.bak "/<!-- rule: $FIRST -->/d" "$HOME_DIR/.codex/AGENTS.md"
expect_stale "missing_link: an always-on rule that was never installed" "was never installed"

# --- rule 1: a copy instead of a symlink ------------------------------------
new_home
rm "$HOME_DIR/.claude/rules/${FIRST%.mdc}.md"
cp "$RULES_SRC/$FIRST" "$HOME_DIR/.claude/rules/${FIRST%.mdc}.md"
expect_stale "copy_not_link: a copied rule file" "is not a symlink"

# --- rule 1: a dangling link ------------------------------------------------
new_home
rm "$HOME_DIR/.claude/rules/${FIRST%.mdc}.md"
ln -s "$RULES_SRC/no-such-rule.mdc" "$HOME_DIR/.claude/rules/${FIRST%.mdc}.md"
expect_stale "dangling: a link whose target is gone" "dangling symlink"

# --- rule 1: a link into some other checkout --------------------------------
new_home
OTHER="$(mktemp -d "${TMPDIR:-/tmp}/other-checkout.XXXXXX")"; DIRS+=("$OTHER")
mkdir -p "$OTHER/rules"; cp "$RULES_SRC/$FIRST" "$OTHER/rules/$FIRST"
rm "$HOME_DIR/.claude/rules/${FIRST%.mdc}.md"
ln -s "$OTHER/rules/$FIRST" "$HOME_DIR/.claude/rules/${FIRST%.mdc}.md"
expect_stale "foreign_checkout: a link resolving into another tree" "not this checkout's"

# --- rule 2: a rule that is no longer always-on -----------------------------
new_home
ln -s "$RULES_SRC/$FIRST" "$HOME_DIR/.claude/rules/retired-rule.md"
expect_stale "retired: a link for a rule this checkout no longer has" "is not an always-on rule here"

# --- rule 3: the baseline every dispatch points at --------------------------
new_home
rm "$HOME_DIR/.claude/rules/agent-baseline.md"
expect_stale "no_baseline: the agent baseline is not installed" "every subagent dispatch points at it"

# --- rule 4: installed, but not rendered into the managed block -------------
new_home
sed -i.bak "/<!-- rule: $FIRST -->/d" "$HOME_DIR/.claude/CLAUDE.md"
expect_stale "unrendered_claude: linked but absent from CLAUDE.md" "no session reads that rule"

new_home
sed -i.bak "/<!-- rule: $FIRST -->/d" "$HOME_DIR/.codex/AGENTS.md"
expect_stale "unrendered_codex: linked but absent from AGENTS.md" "no session reads that rule"

# --- rule 4: a marker for a rule that no longer exists ----------------------
new_home
printf '\n<!-- rule: retired-rule.mdc -->\n\nbody\n' >> "$HOME_DIR/.claude/CLAUDE.md"
expect_stale "stale_marker: a rendered rule this checkout does not have" "which is not an always-on rule here"

# --- a linked worktree is skipped, not failed --------------------------------
# setup.sh links into the checkout it ran from, so in an apply worktree every
# rule resolves to the main checkout. Rule 1 would call all of them misplaced and
# fail `## lint` in every worktree a /myflow-* change runs in. Detected by `.git`
# being a file.
new_home
WT_REPO="$(mktemp -d "${TMPDIR:-/tmp}/check-installed-rules-wt.XXXXXX")"; DIRS+=("$WT_REPO")
mkdir -p "$WT_REPO/scripts/lib" "$WT_REPO/rules"
cp "$GUARD" "$WT_REPO/scripts/"
cp "$RULES_SRC"/*.mdc "$RULES_SRC/agent-baseline.md" "$WT_REPO/rules/" 2>/dev/null
printf 'gitdir: /somewhere/.git/worktrees/x\n' > "$WT_REPO/.git"
set +e
OUT="$(HOME="$HOME_DIR" "$WT_REPO/scripts/$(basename "$GUARD")" 2>&1)"
RC=$?
set -e
if [ "$RC" -ne 0 ]; then
  fail "worktree: a linked worktree was reported stale (rc=$RC): $OUT"
elif ! printf '%s' "$OUT" | grep -qF "INSTALLED-RULES-WORKTREE"; then
  fail "worktree: exited 0 but did not say it was a linked worktree; got: $OUT"
else
  pass "worktree: a linked worktree is skipped, not failed"
fi

# --- not installed is not stale ---------------------------------------------
HOME_DIR="$(mktemp -d "${TMPDIR:-/tmp}/check-installed-rules-empty.XXXXXX")"; DIRS+=("$HOME_DIR")
run_guard
if [ "$RC" -ne 0 ]; then
  fail "no_install: an absent install was reported stale: $OUT"
elif ! printf '%s' "$OUT" | grep -qF "INSTALLED-RULES-NONE"; then
  fail "no_install: exited 0 but did not say no install was found; got: $OUT"
else
  pass "no_install: an absent install is reported, not failed"
fi

# --- a partial install IS stale ---------------------------------------------
new_home
rm "$HOME_DIR/.claude/rules/${FIRST%.mdc}.md"
run_guard
if [ "$RC" -eq 0 ]; then
  fail "partial: an install missing one rule passed"
else
  pass "partial: an install missing one rule is stale, not absent"
fi

if [ "$FAILURES" -ne 0 ]; then
  printf 'TEST-CHECK-INSTALLED-RULES-FAIL: %d case(s) failed\n' "$FAILURES"
  exit 1
fi
printf 'TEST-CHECK-INSTALLED-RULES-OK: every case passed\n'
exit 0
