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
# override honored only when set, so the companion harness
# (test-check-installed-rules.sh) can point this guard at a sandboxed fixture
# under TMPDIR — never set it for a normal invocation.
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
# is exactly where a `/myflow-*` change runs its lint. That is a false alarm
# about a correct install, not drift. This guard detects a linked worktree by
# `.git` being a file rather than a directory, reports it, and exits 0.
#
# The skip applies only when CHECK_INSTALLED_RULES_HOME is unset. That override
# means a caller is deliberately pointing the guard at a fixture home, and wants
# the real comparison run against it; short-circuiting there would make the
# harness unable to test anything from inside a worktree, which is where it runs.
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
#      the managed block of each installed harness file — `.claude/CLAUDE.md`
#      and `.codex/AGENTS.md` — and each block carries no marker for a rule
#      that is no longer always-on. The marker is what `render_managed_block`
#      in setup.sh writes per rule, so it is read here rather than any title
#      or body text, which would re-derive the renderer's formatting and go
#      stale against it. A harness file that does not exist is skipped with a
#      note, for the same reason rule 1's whole directory is: absent is not
#      stale.
#
# Prints one violation line per finding, then the verdict:
#   INSTALLED-RULES-OK:      <home> — N always-on rule(s) installed and rendered
#   INSTALLED-RULES-NONE:    <home> — no global install found, nothing to check
#   INSTALLED-RULES-STALE:   <home> — N violation(s); re-run ./setup.sh global
#
# Every violation has the same remedy, which is why the verdict names it once:
# `./setup.sh global` from this checkout.
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
RULES_SRC="$REPO_ROOT/rules"
HOME_DIR="${CHECK_INSTALLED_RULES_HOME:-$HOME}"
RULES_DIR="$HOME_DIR/.claude/rules"

VIOLATIONS=0
violation() {
  echo "check-installed-rules: $*" >&2
  VIOLATIONS=$((VIOLATIONS + 1))
}

# always_on_rules — the same frontmatter test setup.sh's own always_on_rules()
# applies, kept identical on purpose: a rule this guard calls always-on that
# the installer does not would demand a link setup.sh never creates, and the
# two would disagree forever with no way to satisfy both. The awk is a copy
# because setup.sh is not sourceable — it runs an install on load.
always_on_rules() {
  local rule_file
  for rule_file in "$RULES_SRC"/*.mdc; do
    [ -f "$rule_file" ] || continue
    if awk '
      { sub(/\r$/, "") }
      NR == 1 && $0 != "---"                             { exit }
      NR > 1  && $0 == "---"                             { closed = 1; exit }
      NR > 20                                            { exit }
      NR > 1  && /^alwaysApply:[[:space:]]*true[[:space:]]*$/ { always = 1 }
      END                                                { exit !(always && closed) }
    ' "$rule_file"; then
      basename "$rule_file"
    fi
  done
}

if [ ! -d "$RULES_SRC" ]; then
  echo "check-installed-rules: $RULES_SRC does not exist — this is not an agents checkout" >&2
  exit 1
fi

mapfile -t ALWAYS_ON < <(always_on_rules)
if [ "${#ALWAYS_ON[@]}" -eq 0 ]; then
  echo "check-installed-rules: $RULES_SRC declares no always-on rule — refusing to report an empty install as clean" >&2
  exit 1
fi

if [ -z "${CHECK_INSTALLED_RULES_HOME:-}" ] && [ -f "$REPO_ROOT/.git" ]; then
  printf 'INSTALLED-RULES-WORKTREE: %s is a linked worktree — the install tracks the main checkout, nothing to check here\n' "$REPO_ROOT"
  exit 0
fi

if [ ! -d "$RULES_DIR" ]; then
  printf 'INSTALLED-RULES-NONE: %s — no global install found, nothing to check\n' "$HOME_DIR"
  exit 0
fi

# Rule 1 — every always-on rule is linked, resolves, and points back here.
for rule_name in "${ALWAYS_ON[@]}"; do
  installed="$RULES_DIR/${rule_name%.mdc}.md"
  want="$RULES_SRC/$rule_name"
  if [ ! -L "$installed" ]; then
    if [ -e "$installed" ]; then
      violation "$installed is not a symlink — a copy goes stale the next time the rule is edited"
    else
      violation "$installed is missing — rules/$rule_name is always-on but was never installed"
    fi
    continue
  fi
  if [ ! -e "$installed" ]; then
    violation "$installed is a dangling symlink — it points at $(readlink "$installed"), which does not exist"
    continue
  fi
  got="$(cd "$(dirname "$installed")" && cd "$(dirname "$(readlink "$installed")")" 2>/dev/null && pwd)/$(basename "$(readlink "$installed")")"
  if [ "$got" != "$want" ]; then
    violation "$installed points at $got, not this checkout's $want"
  fi
done

# Rule 2 — nothing installed that is no longer an always-on rule.
for installed in "$RULES_DIR"/*.md; do
  [ -e "$installed" ] || [ -L "$installed" ] || continue
  base="$(basename "$installed")"
  [ "$base" = "agent-baseline.md" ] && continue
  found=0
  for rule_name in "${ALWAYS_ON[@]}"; do
    [ "${rule_name%.mdc}.md" = "$base" ] && { found=1; break; }
  done
  if [ "$found" -eq 0 ]; then
    violation "$installed is installed but rules/${base%.md}.mdc is not an always-on rule here — a stale link still reads as installed"
  fi
done

# Rule 3 — the baseline every dispatched agent is told to read.
baseline="$RULES_DIR/agent-baseline.md"
if [ ! -e "$baseline" ]; then
  violation "$baseline is missing or dangling — every subagent dispatch points at it"
fi

# Rule 4 — the managed block in each installed harness file renders each rule.
HARNESS_FILES=(".claude/CLAUDE.md" ".codex/AGENTS.md")
for rel in "${HARNESS_FILES[@]}"; do
  harness="$HOME_DIR/$rel"
  if [ ! -f "$harness" ]; then
    echo "check-installed-rules: $harness does not exist, not scanned" >&2
    continue
  fi
  for rule_name in "${ALWAYS_ON[@]}"; do
    if ! grep -qF "<!-- rule: $rule_name -->" "$harness"; then
      violation "$harness carries no managed-block marker for $rule_name — no session reads that rule"
    fi
  done
  while IFS= read -r marker; do
    found=0
    for rule_name in "${ALWAYS_ON[@]}"; do
      [ "$marker" = "$rule_name" ] && { found=1; break; }
    done
    if [ "$found" -eq 0 ]; then
      violation "$harness renders $marker, which is not an always-on rule here"
    fi
  done < <(grep -o '<!-- rule: [^ ]*\.mdc -->' "$harness" | sed -E 's/^<!-- rule: (.*) -->$/\1/' | sort -u)
done

if [ "$VIOLATIONS" -ne 0 ]; then
  printf 'INSTALLED-RULES-STALE: %s — %d violation(s); re-run ./setup.sh global from %s\n' \
    "$HOME_DIR" "$VIOLATIONS" "$REPO_ROOT"
  exit 1
fi

printf 'INSTALLED-RULES-OK: %s — %d always-on rule(s) installed and rendered\n' \
  "$HOME_DIR" "${#ALWAYS_ON[@]}"
exit 0
