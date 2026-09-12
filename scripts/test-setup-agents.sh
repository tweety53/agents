#!/usr/bin/env bash
# test-setup-agents.sh — asserts setup.sh installs the three
# flow-<effort> agent definitions and the flow-review reviewer definition
# on `./setup.sh global`.
#
# Runs `setup.sh global` against a sandboxed $HOME (mktemp -d), so it never
# touches the real ~/.claude. Asserts four symlinks exist under
# $HOME/.claude/agents/, each resolves, and each frontmatter carries its own
# `effort:` line and no `model:` line — the Agent tool's dispatch-time
# `model` parameter overrides a definition's `model`, so the model axis
# needs no definition of its own while the effort axis does. flow-review
# carries neither axis: the default-toggle panel decides no effort, and its
# `tools:` allowlist must omit `Agent` — a reviewer structurally cannot fork
# (KAN-495).
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

reviewer_link="$AGENTS_DIR/flow-review.md"
if [[ ! -L "$reviewer_link" ]]; then
  fail "$reviewer_link is not a symlink"
elif [[ ! -e "$reviewer_link" ]]; then
  fail "$reviewer_link does not resolve"
else
  if ! grep -qE '^name: flow-review$' "$reviewer_link"; then
    fail "$reviewer_link frontmatter missing 'name: flow-review'"
  fi
  if ! grep -qE '^tools: ' "$reviewer_link"; then
    fail "$reviewer_link frontmatter missing a 'tools:' allowlist"
  fi
  if grep -E '^tools: ' "$reviewer_link" | grep -q 'Agent'; then
    fail "$reviewer_link 'tools:' allowlist grants Agent — a reviewer must not fork"
  fi
  if grep -E '^tools: ' "$reviewer_link" | grep -q 'Task'; then
    fail "$reviewer_link 'tools:' allowlist grants Task — a reviewer must not fork"
  fi
  if grep -qE '^model: ' "$reviewer_link"; then
    fail "$reviewer_link frontmatter still carries a 'model:' line"
  fi
  if grep -qE '^effort: ' "$reviewer_link"; then
    fail "$reviewer_link carries an 'effort:' line — the default toggle decides no effort"
  fi
fi

actual_count=$(find "$AGENTS_DIR" -maxdepth 1 -name 'flow-*.md' 2>/dev/null | wc -l | tr -d ' ')
[ "$actual_count" -eq 4 ] && pass "four symlinks present under $AGENTS_DIR" \
  || fail "expected 4 symlinks under $AGENTS_DIR, found $actual_count"

if [ "$FAILURES" -eq 0 ]; then
  echo "test-setup-agents.sh: all checks passed"
  exit 0
else
  echo "test-setup-agents.sh: $FAILURES failure(s)"
  exit 1
fi
