#!/usr/bin/env bash
# test-setup-agents.sh — asserts setup.sh installs the nine
# flow-<model>-<effort> agent definitions on `./setup.sh global`.
#
# Runs `setup.sh global` against a sandboxed $HOME (mktemp -d), so it never
# touches the real ~/.claude. Asserts nine symlinks exist under
# $HOME/.claude/agents/, each resolves, and each frontmatter carries a
# `model:` line and an `effort:` line.
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

MODELS="sonnet opus haiku"
EFFORTS="low medium high"

count=0
for model in $MODELS; do
  for effort in $EFFORTS; do
    count=$((count + 1))
    link="$AGENTS_DIR/flow-$model-$effort.md"
    if [[ ! -L "$link" ]]; then
      fail "$link is not a symlink"
      continue
    fi
    if [[ ! -e "$link" ]]; then
      fail "$link does not resolve"
      continue
    fi
    if ! grep -qE '^model: '"$model"'$' "$link"; then
      fail "$link frontmatter missing 'model: $model'"
    fi
    if ! grep -qE '^effort: '"$effort"'$' "$link"; then
      fail "$link frontmatter missing 'effort: $effort'"
    fi
  done
done
[ "$count" -eq 9 ] || fail "expected to check 9 combinations, checked $count"

actual_count=$(find "$AGENTS_DIR" -maxdepth 1 -name 'flow-*.md' 2>/dev/null | wc -l | tr -d ' ')
[ "$actual_count" -eq 9 ] && pass "nine symlinks present under $AGENTS_DIR" \
  || fail "expected 9 symlinks under $AGENTS_DIR, found $actual_count"

if [ "$FAILURES" -eq 0 ]; then
  echo "test-setup-agents.sh: all checks passed"
  exit 0
else
  echo "test-setup-agents.sh: $FAILURES failure(s)"
  exit 1
fi
