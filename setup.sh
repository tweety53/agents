#!/usr/bin/env bash
# setup.sh — Install agents-data into a project for a specific harness
#
# Usage: ./setup.sh <harness> [project-dir]
#
# Harnesses: cursor | claude-code | codex | all
#
# Examples:
#   ./setup.sh claude-code                    # current directory
#   ./setup.sh cursor /path/to/other-project
#   ./setup.sh all /path/to/gymie

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILLS_SRC="$SCRIPT_DIR/skills"
RULES_SRC="$SCRIPT_DIR/rules"
COMMANDS_CURSOR_SRC="$SCRIPT_DIR/commands"
COMMANDS_CLAUDE_SRC="$SCRIPT_DIR/commands-claude"
HARNESS="${1:-}"
PROJECT_DIR="${2:-.}"
PROJECT_DIR="$(cd "$PROJECT_DIR" && pwd)"

die() { echo "ERROR: $*" >&2; exit 1; }
info() { echo "→ $*"; }

[[ -n "$HARNESS" ]] || die "Usage: $0 <cursor|claude-code|codex|all> [project-dir]"
[[ -d "$SKILLS_SRC" ]] || die "skills/ directory not found at $SKILLS_SRC"

install_skills() {
  local target_dir="$1"
  info "Installing project skills into $target_dir"
  mkdir -p "$target_dir"
  for skill_dir in "$SKILLS_SRC"/*/; do
    skill_name=$(basename "$skill_dir")
    dest="$target_dir/$skill_name"
    if [[ -L "$dest" ]]; then
      rm "$dest"
    fi
    ln -sf "$skill_dir" "$dest"
    echo "  ✓ symlinked $skill_name"
  done
}

install_claude_code() {
  info "Setting up for Claude Code in $PROJECT_DIR"
  install_skills "$PROJECT_DIR/.claude/skills"
  install_commands "$COMMANDS_CLAUDE_SRC" "$PROJECT_DIR/.claude/commands"
  if [[ ! -f "$PROJECT_DIR/CLAUDE.md" ]]; then
    cp "$SCRIPT_DIR/CLAUDE.md" "$PROJECT_DIR/CLAUDE.md"
    info "Copied CLAUDE.md to project root"
  else
    info "CLAUDE.md already exists — skipping copy (diff manually if needed)"
  fi
  echo ""
  echo "✅ Claude Code setup complete."
  echo "   Skills → .claude/skills/  Commands → .claude/commands/"
  echo "   Next: in a Claude Code session, run /plugin install prime-radiant-inc/superpowers"
}

install_codex() {
  info "Setting up for Codex in $PROJECT_DIR"
  install_skills "$PROJECT_DIR/.codex/skills"
  if [[ ! -f "$PROJECT_DIR/AGENTS.md" ]]; then
    cp "$SCRIPT_DIR/AGENTS.md" "$PROJECT_DIR/AGENTS.md"
    info "Copied AGENTS.md to project root"
  else
    info "AGENTS.md already exists — skipping copy (diff manually if needed)"
  fi
  echo ""
  echo "✅ Codex setup complete."
  echo "   Next: install Superpowers for Codex from its fork repo."
  echo "   Enable: [features] multi_agent = true  in ~/.codex/config.toml"
}

install_rules_cursor() {
  local target_dir="$1"
  [[ -d "$RULES_SRC" ]] || return 0
  info "Installing myflow rules into $target_dir"
  mkdir -p "$target_dir"
  for rule_file in "$RULES_SRC"/*.mdc; do
    [[ -f "$rule_file" ]] || continue
    rule_name=$(basename "$rule_file")
    dest="$target_dir/$rule_name"
    if [[ -L "$dest" ]]; then
      rm "$dest"
    fi
    ln -sf "$rule_file" "$dest"
    echo "  ✓ symlinked $rule_name"
  done
}

install_commands() {
  local commands_src="$1"
  local target_dir="$2"
  [[ -d "$commands_src" ]] || return 0
  info "Installing slash commands into $target_dir"
  mkdir -p "$target_dir"
  for cmd_file in "$commands_src"/*.md; do
    [[ -f "$cmd_file" ]] || continue
    cmd_name=$(basename "$cmd_file")
    dest="$target_dir/$cmd_name"
    if [[ -L "$dest" ]]; then
      rm "$dest"
    fi
    ln -sf "$cmd_file" "$dest"
    echo "  ✓ symlinked $cmd_name"
  done
}

install_cursor() {
  info "Setting up for Cursor in $PROJECT_DIR"
  install_skills "$PROJECT_DIR/.cursor/skills"
  install_rules_cursor "$PROJECT_DIR/.cursor/rules"
  install_commands "$COMMANDS_CURSOR_SRC" "$PROJECT_DIR/.cursor/commands"
  echo ""
  echo "✅ Cursor setup complete."
  echo "   Skills → .cursor/skills/  Rules → .cursor/rules/  Commands → .cursor/commands/"
}

case "$HARNESS" in
  cursor)      install_cursor ;;
  claude-code) install_claude_code ;;
  codex)       install_codex ;;
  all)
    install_cursor
    echo ""
    install_claude_code
    echo ""
    install_codex
    ;;
  *) die "Unknown harness '$HARNESS'. Choose: cursor | claude-code | codex | all" ;;
esac
