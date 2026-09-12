#!/usr/bin/env bash
# test-setup-agents.sh — asserts setup.sh installs the three
# flow-<effort> agent definitions on `./setup.sh global`.
#
# Runs `setup.sh global` against a sandboxed $HOME (mktemp -d), so it never
# touches the real ~/.claude. Asserts three symlinks exist under
# $HOME/.claude/agents/, each resolves, and each frontmatter carries its own
# `effort:` line and no `model:` line — the Agent tool's dispatch-time
# `model` parameter overrides a definition's `model`, so the model axis
# needs no definition of its own while the effort axis does.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
FAILURES=0

pass() { echo "  PASS: $1"; }
fail() { echo "  FAIL: $1"; FAILURES=$((FAILURES + 1)); }

SANDBOX_HOME="$(mktemp -d)"
trap 'rm -rf "$SANDBOX_HOME"' EXIT

HOME="$SANDBOX_HOME" "$REPO_DIR/setup.sh" global >/tmp/test-setup-agents.out 2>&1
RC=$?
[ "$RC" -eq 0 ] || fail "setup.sh global exited $RC — see /tmp/test-setup-agents.out"

AGENTS_DIR="$SANDBOX_HOME/.claude/agents"

EFFORTS="low medium high"

count=0
for effort in $EFFORTS; do
  count=$((count + 1))
  link="$AGENTS_DIR/flow-$effort.md"
  if [[ ! -L "$link" ]]; then
    fail "$link is not a symlink"
    continue
  fi
  if [[ ! -e "$link" ]]; then
    fail "$link does not resolve"
    continue
  fi
  if ! grep -qE '^effort: '"$effort"'$' "$link"; then
    fail "$link frontmatter missing 'effort: $effort'"
  fi
  if grep -qE '^model: ' "$link"; then
    fail "$link frontmatter still carries a 'model:' line"
  fi
done
[ "$count" -eq 3 ] || fail "expected to check 3 efforts, checked $count"

actual_count=$(find "$AGENTS_DIR" -maxdepth 1 -name 'flow-*.md' 2>/dev/null | wc -l | tr -d ' ')
[ "$actual_count" -eq 3 ] && pass "three symlinks present under $AGENTS_DIR" \
  || fail "expected 3 symlinks under $AGENTS_DIR, found $actual_count"

if [ "$FAILURES" -eq 0 ]; then
  echo "test-setup-agents.sh: all checks passed"
  exit 0
else
  echo "test-setup-agents.sh: $FAILURES failure(s)"
  exit 1
fi
