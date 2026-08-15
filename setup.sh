#!/usr/bin/env bash
# setup.sh — Install agents-data into a project for a specific harness
#
# Usage: ./setup.sh <harness> [project-dir]
#
# Harnesses: cursor | claude-code | codex | all | global
#
# Examples:
#   ./setup.sh claude-code                    # current directory
#   ./setup.sh cursor /path/to/other-project
#   ./setup.sh all /path/to/gymie
#   ./setup.sh global                         # user-level install (~/.claude, ~/.cursor)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILLS_SRC="$SCRIPT_DIR/skills"
RULES_SRC="$SCRIPT_DIR/rules"
COMMANDS_CURSOR_SRC="$SCRIPT_DIR/commands"
COMMANDS_CLAUDE_SRC="$SCRIPT_DIR/commands-claude"
HARNESS="${1:-}"
PROJECT_DIR="${2:-.}"
PROJECT_DIR="$(cd "$PROJECT_DIR" && pwd)"

CLAUDE_MD_BEGIN='<!-- myflow:begin -->'
CLAUDE_MD_END='<!-- myflow:end -->'

die() { echo "ERROR: $*" >&2; exit 1; }
info() { echo "→ $*"; }
warn() { echo "  ⚠ $*" >&2; }

# require_grep_ok <rc> <what-was-being-done>
# grep exits 0 when it matched and 1 when it did not; both are normal answers. Anything
# higher is a real failure (unreadable file, bad pattern), and every caller below uses
# grep's answer to decide whether a managed block already exists. Treating a failure as
# "found nothing" routes the run into the append branch — the one outcome that can never
# converge (see install_managed_block). So it aborts instead.
require_grep_ok() {
  (( $1 < 2 )) || die "grep exited $1 while $2 (reason above).
  Its answer decides whether a managed myflow block already exists, and guessing
  'no block' here would append a second one that no later run could reconcile."
}

# Temp files are registered here and removed by a single EXIT trap: unlike a
# per-function RETURN trap, this also fires when `die` aborts the script or when
# `set -e` kills it mid-rewrite, which is exactly when a leak would happen.
TMP_FILES=()
cleanup_tmp() { [[ ${#TMP_FILES[@]} -eq 0 ]] || rm -f "${TMP_FILES[@]}"; }
trap cleanup_tmp EXIT

[[ -n "$HARNESS" ]] || die "Usage: $0 <cursor|claude-code|codex|all|global> [project-dir]"
[[ -d "$SKILLS_SRC" ]] || die "skills/ directory not found at $SKILLS_SRC"

# Count of items that could not be linked. A skip must never be reportable as a
# complete install, so the closing banner and the exit status both consult this.
SKIPPED=0

# link_into <src> <dest> <label>
# Symlinks src at dest. An existing symlink is replaced. Anything else already at
# dest — a real file or a stale copied directory — is moved aside with a warning
# first: `ln -sf` over a directory would nest the link *inside* it and leave the
# stale copy shadowing the install, and over a file it would destroy hand-written
# content. Success is only ever reported for a link actually created.
#
# Where "aside" is depends on what is being displaced, and the difference matters:
#
#   - A FILE goes to `<dest>.bak`. That is safe because the harness globs
#     `commands/*.md` and `rules/*.mdc`, and `.bak` falls outside both globs.
#   - A DIRECTORY goes to a sibling `<parent>-backup/<name>-<timestamp>/`, OUTSIDE
#     the scanned tree. `<dest>.bak` would not work here: a skills directory is
#     discovered by walking the tree for SKILL.md, so `openspec-propose.bak/` stays
#     live alongside the symlink — two loadable skills declaring the same name,
#     which is precisely the stale-copy shadowing this function exists to prevent.
link_into() {
  local src="$1" dest="$2" label="$3" backup
  if [[ -L "$dest" ]]; then
    rm "$dest"
  elif [[ -e "$dest" ]]; then
    if [[ -d "$dest" ]]; then
      backup="$(dirname "$dest")-backup/$(basename "$dest")-$(date +%Y%m%d-%H%M%S)"
      mkdir -p "$(dirname "$backup")"
    else
      backup="$dest.bak"
    fi
    if [[ -e "$backup" ]]; then
      warn "$dest already exists and $backup is taken — SKIPPED, nothing linked; move one aside and re-run"
      SKIPPED=$((SKIPPED + 1))
      return 0
    fi
    mv "$dest" "$backup"
    warn "$dest already existed — moved to $backup before symlinking"
  fi
  ln -s "$src" "$dest"
  echo "  ✓ symlinked $label"
}

# finish_banner <label> <skipped-baseline>
# Closes an install mode. A run that skipped an item is not a complete install and
# must not look like one — the count is taken since the baseline so that `all`
# reports each mode's own skips rather than the running total.
finish_banner() {
  local label="$1" n=$((SKIPPED - $2))
  if (( n > 0 )); then
    echo "⚠ $label setup completed with $n skipped — see the warnings above. Those items are NOT installed."
  else
    echo "✅ $label setup complete."
  fi
}

# always_on_rules — names of the rules that apply to every project.
# The fact is a property of each rule, declared once in its own frontmatter
# (`alwaysApply: true`). Every other rule is opt-in: never installed as a file by any path,
# and reaching a project only when that project names it in its own `.myflow/project.md`,
# which a per-project install renders into the project's own managed block. See
# install_project_standards.
always_on_rules() {
  local rule_file
  for rule_file in "$RULES_SRC"/*.mdc; do
    [[ -f "$rule_file" ]] || continue
    # Read the leading YAML frontmatter only; `always` is set iff it declares
    # alwaysApply: true. (A bare `exit 0` would be overridden by END.)
    #
    # Three things this awk is careful about:
    #   - `sub(/\r$/, "")` first: in a CRLF checkout line 1 is `---\r`, which fails the
    #     `$0 != "---"` test, so an always-on rule would be silently dropped from BOTH
    #     ~/.cursor/rules and the managed block — an install that looks successful and
    #     ships none of the mandatory rules.
    #   - The value is anchored at the tail, so `alwaysApply: true_for_kotlin_only`
    #     (or any other `true…` prefix) is no longer read as `true`.
    #   - The frontmatter must actually CLOSE (`closed`) for the declaration to count. A
    #     file that opens `---` and never closes it is not frontmatter at all, and without
    #     this a prose line documenting `alwaysApply: true` in an example would install the
    #     rule globally. `NR > 20` additionally bounds the scan; real rule frontmatter
    #     closes by line 5, so nothing legitimate reaches it.
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

# prune_stale_links <target-dir>
#
# Remove symlinks in <target-dir> whose target no longer exists. Installs are additive — both
# loops below iterate the CURRENT source tree — so an entry deleted from this repo leaves its
# symlink behind at every destination it was ever installed to, forever.
#
# That is not cosmetic. A stale `~/.claude/commands/myflow-review.md` still matches the harness's
# command glob, so a command deleted from this repo goes on being offered and invoked; it then
# fails on a broken link, or worse, serves cached content naming skills that are also gone. The
# rename that retired twelve commands and fifteen skills is exactly when this bites.
#
# THREE conditions, all required. `[[ -L ]]` and `[[ ! -e ]]` alone are not enough: `[[ ! -e ]]`
# follows the link and is true whenever the target cannot be stat'd, which cannot distinguish
# "deleted from this repo" from "on a volume that is not mounted right now". A user's own symlink
# into a removable disk or a checkout they moved is broken *at this moment* and must survive.
#
# So the third condition is the load-bearing one: the link must point INTO this repo. That is
# exactly the set this installer creates, and nothing else — a link the user made to anywhere else
# is left alone however broken it looks. `readlink` is used rather than `realpath`, because the
# target does not exist and cannot be resolved; the raw stored path is what we test.
prune_stale_links() {
  local target_dir="$1" entry link_target removed=0
  [[ -d "$target_dir" ]] || return 0
  if [[ ! -r "$target_dir" ]]; then
    warn "cannot read $target_dir — skipping the stale-link prune (fix its permissions and re-run)"
    return 0
  fi
  for entry in "$target_dir"/* "$target_dir"/.[!.]*; do
    [[ -L "$entry" ]] || continue          # never touch a real file or directory
    [[ -e "$entry" ]] && continue          # still resolves — leave it
    link_target="$(readlink "$entry")"
    # Only ours: a link whose stored target is inside this repo. Anything else is the user's.
    [[ "$link_target" == "$SCRIPT_DIR"/* ]] || continue
    rm -f "$entry"
    removed=$((removed + 1))
    info "Pruned stale link $(basename "$entry") (its source in this repo no longer exists)"
  done
  (( removed == 0 )) || info "Pruned $removed stale link(s) from $target_dir"
}

install_skills() {
  local target_dir="$1" skill_dir skill_name
  info "Installing project skills into $target_dir"
  mkdir -p "$target_dir"
  prune_stale_links "$target_dir"
  for skill_dir in "$SKILLS_SRC"/*/; do
    [[ -d "$skill_dir" ]] || continue
    skill_name=$(basename "$skill_dir")
    link_into "$skill_dir" "$target_dir/$skill_name" "$skill_name"
  done
}

install_claude_code() {
  local skipped_before=$SKIPPED
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
  finish_banner "Claude Code" "$skipped_before"
  echo "   Skills → .claude/skills/  Commands → .claude/commands/"
  echo "   Next: in a Claude Code session, run /plugin install prime-radiant-inc/superpowers"
}

install_codex() {
  local skipped_before=$SKIPPED
  info "Setting up for Codex in $PROJECT_DIR"
  install_skills "$PROJECT_DIR/.codex/skills"
  if [[ ! -f "$PROJECT_DIR/AGENTS.md" ]]; then
    cp "$SCRIPT_DIR/AGENTS.md" "$PROJECT_DIR/AGENTS.md"
    info "Copied AGENTS.md to project root"
  else
    info "AGENTS.md already exists — skipping copy (diff manually if needed)"
  fi
  echo ""
  finish_banner "Codex" "$skipped_before"
  echo "   Always-on rules are NOT installed per project — run './setup.sh global' for those;"
  echo "   it writes the managed block into ~/.codex/AGENTS.md, which every Codex session reads."
  echo "   This project's OPT-IN rules are a separate matter — see the project standards step below."
  echo "   Next: install Superpowers for Codex from its fork repo."
  echo "   Enable: [features] multi_agent = true  in ~/.codex/config.toml"
}

# install_rules_cursor <target-dir>
# Installs the always-on rules, and only those — globally and per project alike.
# An opt-in rule (e.g. the Kotlin backend standard, whose globs would otherwise
# match a Compose Multiplatform or IDE-plugin repo) is never installed by path:
# whether a rule is always-on is decided by the rule, not by the caller. Opt-in rules
# reach a project through install_project_standards instead — as inlined text in the
# project's own managed block, never as a `.cursor/rules/` entry that Cursor would
# apply by glob to every file in the repo.
install_rules_cursor() {
  local target_dir="$1" rule_name rule_file
  [[ -d "$RULES_SRC" ]] || return 0
  info "Installing myflow rules into $target_dir"
  mkdir -p "$target_dir"
  while IFS= read -r rule_name; do
    rule_file="$RULES_SRC/$rule_name"
    [[ -f "$rule_file" ]] || continue
    link_into "$rule_file" "$target_dir/$rule_name" "$rule_name"
  done < <(always_on_rules)
}

# install_rules_claude <target-dir>
# Installs the FULL text of every always-on rule as ~/.claude/rules/<name>.md, plus the
# agent baseline that dispatched subagents read.
#
# The managed block in ~/.claude/CLAUDE.md carries only each rule's core excerpt and a
# `Full rule: ~/.claude/rules/<name>.md` pointer, so these links are what make that pointer
# resolve. They are symlinks into this repo, never copies: the block's core and the full text
# are then the same file, and there is no second copy to fall out of date.
#
# The `.mdc` source is linked under a `.md` name. Claude Code has no `.mdc` association, and
# the pointer a rule renders into the block names `.md`; the frontmatter that remains at the
# top of the linked file is a two-line header a reader skips.
#
# agent-baseline.md is not a rule and carries no frontmatter — it is the one file a dispatched
# agent is told to read, listing each rule in a line and pointing at these same paths. It ships
# here because it is useless unless the paths it names exist.
install_rules_claude() {
  local target_dir="$1" rule_name rule_file baseline
  [[ -d "$RULES_SRC" ]] || return 0
  info "Installing full rule texts into $target_dir"
  mkdir -p "$target_dir"
  prune_stale_links "$target_dir"
  while IFS= read -r rule_name; do
    rule_file="$RULES_SRC/$rule_name"
    [[ -f "$rule_file" ]] || continue
    link_into "$rule_file" "$target_dir/${rule_name%.mdc}.md" "${rule_name%.mdc}.md"
  done < <(always_on_rules)
  baseline="$RULES_SRC/agent-baseline.md"
  if [[ -f "$baseline" ]]; then
    link_into "$baseline" "$target_dir/agent-baseline.md" "agent-baseline.md"
  else
    warn "$baseline is missing — subagent dispatches point at a file that does not exist"
    SKIPPED=$((SKIPPED + 1))
  fi
}

# install_hooks <claude-dir>
# Installs the harness hooks, then reports on the one thing this installer will not do for
# the user: registering them.
#
# settings.json is the user's own file — model, theme, permissions, their own hooks. Editing
# JSON from bash means either a dependency this script does not have or a hand-rolled parser
# that eventually eats a config it did not understand. The hook file is installed and the
# exact snippet is printed instead; registering it is one paste, and it stays the user's call.
install_hooks() {
  local claude_dir="$1" hooks_src="$SCRIPT_DIR/hooks" hook_file hook_name settings
  [[ -d "$hooks_src" ]] || return 0
  info "Installing hooks into $claude_dir/hooks"
  mkdir -p "$claude_dir/hooks"
  prune_stale_links "$claude_dir/hooks"
  for hook_file in "$hooks_src"/*; do
    [[ -f "$hook_file" ]] || continue
    hook_name=$(basename "$hook_file")
    link_into "$hook_file" "$claude_dir/hooks/$hook_name" "$hook_name"
  done
  # `set -e` is on, so grep's normal "no match" (rc 1) must not be allowed to kill the run —
  # capture the status instead of testing it inline, then let require_grep_ok reject only a
  # real grep failure (rc ≥ 2), the same way every other caller in this script does.
  settings="$claude_dir/settings.json"
  if [[ -f "$settings" ]]; then
    local rc=0
    grep -q 'enforce-agent-baseline' "$settings" || rc=$?
    require_grep_ok "$rc" "checking $settings for the agent-baseline hook"
    (( rc == 0 )) && return 0
  fi
  echo ""
  echo "  ⚠ The agent-baseline hook is installed but NOT registered, so nothing yet enforces"
  echo "    that subagent dispatches carry the rules. Add this to \"hooks\" in $settings:"
  cat <<'SNIPPET'

    "PreToolUse": [
      {
        "matcher": "Agent|Task",
        "hooks": [
          { "type": "command", "command": "python3 \"$HOME/.claude/hooks/enforce-agent-baseline.py\"" }
        ]
      }
    ]
SNIPPET
  echo ""
}

install_commands() {
  local commands_src="$1"
  local target_dir="$2"
  local cmd_file cmd_name
  [[ -d "$commands_src" ]] || return 0
  info "Installing slash commands into $target_dir"
  mkdir -p "$target_dir"
  prune_stale_links "$target_dir"
  for cmd_file in "$commands_src"/*.md; do
    [[ -f "$cmd_file" ]] || continue
    cmd_name=$(basename "$cmd_file")
    link_into "$cmd_file" "$target_dir/$cmd_name" "$cmd_name"
  done
}

install_cursor() {
  local skipped_before=$SKIPPED
  info "Setting up for Cursor in $PROJECT_DIR"
  install_skills "$PROJECT_DIR/.cursor/skills"
  install_rules_cursor "$PROJECT_DIR/.cursor/rules"
  install_commands "$COMMANDS_CURSOR_SRC" "$PROJECT_DIR/.cursor/commands"
  echo ""
  finish_banner "Cursor" "$skipped_before"
  echo "   Skills → .cursor/skills/  Rules → .cursor/rules/  Commands → .cursor/commands/"
}

# render_managed_block [rule-name…]
#
# Render rule bodies as the content of the managed block. With no arguments the set is the
# always-on rules — the global layer. With arguments it is exactly those rules, which is how
# a project's opted-in standards are rendered into its own CLAUDE.md / AGENTS.md.
#
# One renderer, two callers, deliberately: the delimiter guard below, the frontmatter
# stripping, and the `<!-- rule: … -->` labelling are the same guarantees in both cases, and
# a second renderer would be a second place for them to drift out of.
#
# Neither Claude Code nor Codex reads `.cursor/rules/`, so this block is the only
# global rule layer for both — the rule text is inlined rather than referenced.
#
# Rendering refuses to proceed if a rule body contains a delimiter on a line of its own.
# No rule does today, so this is a guard rather than a live bug — but the failure it
# prevents is unrecoverable by the installer: the delimiter would be inlined into the
# managed block, the NEXT run would count two begins, and the run would die on a file the
# installer itself corrupted, leaving the user to hand-edit generated content. Refusing to
# write is the only outcome that keeps the file reconcilable.
render_managed_block() {
  local rule_name rule_file body has_open has_close
  local -a rule_names=()
  # Global render (no explicit rule list) is the only mode that honours core markers; see
  # the core-extraction comment in the loop below for why.
  local global_render=1
  if (( $# > 0 )); then
    rule_names=("$@")
    global_render=0
  else
    while IFS= read -r rule_name; do rule_names+=("$rule_name"); done < <(always_on_rules)
  fi
  printf '%s\n' "$CLAUDE_MD_BEGIN"
  # The provenance line has to name the command that reproduces THIS block. A reader of a
  # project's CLAUDE.md told to re-run `setup.sh global` would run the one mode that never
  # writes the file they are looking at.
  if (( $# > 0 )); then
    printf '<!-- Generated by agents/setup.sh from this project'"'"'s .myflow/project.md\n'
    printf '     ## standards section. Do not edit inside this block: it is rewritten on\n'
    printf '     every install. Put your own notes outside the delimiters. -->\n'
  else
    printf '<!-- Generated by agents/setup.sh global. Do not edit inside this block: it is\n'
    printf '     rewritten on every install. Put your own notes outside the delimiters. -->\n'
  fi
  for rule_name in ${rule_names[@]+"${rule_names[@]}"}; do
    rule_file="$RULES_SRC/$rule_name"
    [[ -f "$rule_file" ]] || continue
    # Strip the leading YAML frontmatter block, if present. A trailing \r is removed first
    # so a CRLF rule file has its frontmatter recognised (and so the rendered block is LF).
    body="$(awk '
      { sub(/\r$/, "") }
      NR==1 && $0=="---" { fm=1; next }
      fm && $0=="---"    { fm=0; next }
      !fm                { print }
    ' "$rule_file")"
    if printf '%s\n' "$body" | grep -qFx -e "$CLAUDE_MD_BEGIN" -e "$CLAUDE_MD_END"; then
      die "rule $rule_name contains a myflow block delimiter on a line of its own.
  Inlining it would put a second delimiter inside the managed block and make every
  later run die on a file this installer wrote. Indent or fence that line in
  $rule_file so it is no longer a bare delimiter, then re-run."
    fi
    # A rule may mark a CORE excerpt — the part that must be in the prompt of every session,
    # as distinct from the full text a reader opens when they need it. Globally the block
    # carries the core plus a pointer to the installed full rule, which lives once in this
    # repo and is reachable at ~/.claude/rules/<name>.md via install_rules_claude.
    #
    # Core extraction is deliberately GLOBAL-ONLY. A project's opted-in standards render into
    # that project's own CLAUDE.md, and pointing those at ~/.claude/rules/ would name a path
    # no install ever creates for an opt-in rule — so with an explicit rule list the whole
    # body is inlined exactly as before. An unmarked rule is likewise inlined whole, which is
    # what keeps myflow-manual-review.mdc and every opt-in standard rendering unchanged.
    # `if` rather than `grep … && flag=1`: under `set -e` an AND-list whose last command
    # fails IS a failing statement, so an unmarked rule — the common case — aborted the
    # render mid-block, leaving a begin delimiter with no end for the next run to trip over.
    # A grep inside an `if` condition is exempt from `set -e`, which is why every other
    # detection in this script is written this way.
    has_open=0; has_close=0
    if printf '%s\n' "$body" | grep -qFx '<!-- core -->';  then has_open=1;  fi
    if printf '%s\n' "$body" | grep -qFx '<!-- /core -->'; then has_close=1; fi
    if (( has_open != has_close )); then
      die "rule $rule_name has an unbalanced core marker. <!-- core --> and <!-- /core -->
  must both be present, each alone on its own line, or neither. One without the other
  silently renders the wrong amount of text into every managed block this installer writes.
  Fix $rule_file, then re-run."
    fi
    printf '\n<!-- rule: %s -->\n\n' "$rule_name"
    if (( global_render && has_open )); then
      # The core is everything up to the closing marker — title and any short preamble
      # included — with the marker lines themselves dropped.
      printf '%s\n' "$body" | awk '
        { sub(/\r$/, "") }
        $0 == "<!-- /core -->" { exit }
        $0 == "<!-- core -->"  { next }
        { print }
      '
      printf 'Full rule: `~/.claude/rules/%s.md`.\n' "${rule_name%.mdc}"
      continue
    fi
    # Whole-body render: the markers are scaffolding for the renderer, not content, so they
    # are dropped here too. Harmless if they survived — they are HTML comments — but a reader
    # of a project's CLAUDE.md should not meet a marker for a mode that file never uses.
    printf '%s\n' "$body" | awk '
      { sub(/\r$/, "") }
      $0 == "<!-- core -->" || $0 == "<!-- /core -->" { next }
      { print }
    '
  done
  printf '\n%s\n' "$CLAUDE_MD_END"
}

# backup_once <file> — keep a copy of <file> as it was before myflow first modified it.
#
# Written once and then never touched again. Re-copying on every run replaces the only
# irreplaceable state — the user's own file, before any generated block existed — with the
# previous generated state, so the documented recovery path survives exactly one install.
# The generated block itself is not worth preserving: re-running this installer reproduces it.
#
# `-p` so a 600-mode file holding personal instructions never gains a wider-mode copy: plain
# `cp` adopts the mode of an already-existing destination (e.g. a 644 .bak left by anything else).
backup_once() {
  local file="$1"
  if [[ -e "$file.myflow.bak" ]]; then
    info "Keeping the existing pre-myflow copy at $file.myflow.bak"
  else
    cp -p "$file" "$file.myflow.bak"
  fi
}

# generated_only <file> <begin-line> <end-line>
# True when the file's ENTIRE content is the managed block — i.e. this installer created it
# from nothing and there is no pre-myflow state to preserve. Copying such a file to
# `.myflow.bak` produces an artifact that is 100% generated content yet is announced as the
# user's "pre-myflow copy": nothing is lost, but the one file the recovery path points at
# now misrepresents what it holds. Re-running this installer reproduces generated content,
# so there is nothing here worth backing up.
generated_only() {
  local file="$1" begin_line="$2" end_line="$3" total
  total="$(wc -l <"$file" | tr -d '[:space:]')" || return 1
  [[ "$begin_line" -eq 1 && "$end_line" -eq "$total" ]]
}

# preflight_managed_block <file>
# Refuse, with context, the two shapes that would otherwise surface as a raw tool error from
# the middle of the run: `mkdir: …/.codex: File exists` when the parent is a regular file, or
# a burst of `grep: … Is a directory` followed by `cp: … is a directory` when the instruction
# file itself is a directory. Both already failed safe; neither told the user which harness
# step it was, that earlier steps had already installed, or what to do about it.
preflight_managed_block() {
  local target_file="$1" dir
  dir="$(dirname "$target_file")"
  [[ ! -e "$dir" || -d "$dir" ]] || die "$dir exists but is not a directory, so the managed
  myflow block for $target_file cannot be written. That path must be the harness's
  configuration directory. Move or remove the file at $dir, then re-run."
  [[ ! -d "$target_file" ]] || die "$target_file is a directory, but it must be the harness's
  agent instruction FILE — that is where the managed myflow block goes. Move the directory
  aside, then re-run."
}

# install_managed_block <target-file> [rule-name…]
#
# Rewrite ONLY the delimited block in an agent instruction file — globally
# (~/.claude/CLAUDE.md for Claude Code, ~/.codex/AGENTS.md for Codex) or in a project
# (<project>/CLAUDE.md and <project>/AGENTS.md, carrying that project's opted-in rules).
# Trailing arguments are the rule list, forwarded verbatim to render_managed_block; with
# none, the always-on set is rendered.
#
# Exactly three outcomes, and no other:
#   - no delimiters at all  → append a fresh block (the next run then sees 1 + 1)
#   - one begin before one end → rewrite just that span, after a .myflow.bak copy
#   - anything else → die
#
# The third case used to append too, and that is the one option that can never
# converge: 1/0 becomes 2/1 becomes 3/2, each run adding another ~600 lines of
# eventually-contradictory rules to the file injected into every session. Dying
# with the offending line numbers is chosen over auto-reconciling because the
# stray delimiter may be sitting in content the user wrote and wants to keep.
install_managed_block() {
  local target_file="$1"
  local block_tmp out_tmp begin_lines end_lines begins ends crlf_markers rc
  local cr=$'\r'
  preflight_managed_block "$target_file"
  mkdir -p "$(dirname "$target_file")"
  block_tmp="$(mktemp)"
  out_tmp="$(mktemp)"
  TMP_FILES+=("$block_tmp" "$out_tmp")
  render_managed_block "${@:2}" >"$block_tmp"

  if [[ ! -s "$target_file" ]]; then
    # Missing or empty file — write the block as the whole content.
    cat "$block_tmp" >"$target_file"
    info "Wrote managed myflow block to $target_file"
    return 0
  fi

  # Every count and every rewrite below matches a marker as a whole LF-terminated line, so a
  # CRLF file silently counts 0 begins and 0 ends and takes the append branch — leaving the
  # existing block orphaned and a second one appended, with no run able to reconcile them.
  # Detect that up front and stop, rather than duplicating the block.
  #
  # Each grep's status is captured and checked (see require_grep_ok) rather than discarded
  # with `|| true`: on a failure the substitution yields the empty string, which every test
  # below reads as zero, and "zero delimiters" is exactly the append branch.
  rc=0
  crlf_markers=$(grep -cFx -e "$CLAUDE_MD_BEGIN$cr" -e "$CLAUDE_MD_END$cr" "$target_file") || rc=$?
  require_grep_ok "$rc" "counting CRLF myflow delimiters in $target_file"
  if [[ "$crlf_markers" -gt 0 ]]; then
    die "$target_file has CRLF line endings on its myflow delimiters ($crlf_markers marker line(s)).
  This installer matches the markers as LF-terminated lines, so it would not find the
  existing block and would append a second one. Convert the file to LF endings
  (e.g. \`perl -pi -e 's/\\r\\n/\\n/' \"$target_file\"\`) and re-run."
  fi

  # `pipefail` is set, so a grep failure inside these pipelines still surfaces as the
  # substitution's status — cut and paste cannot fail on their own.
  rc=0; begin_lines="$(grep -nFx "$CLAUDE_MD_BEGIN" "$target_file" | cut -d: -f1 | paste -sd, -)" || rc=$?
  require_grep_ok "$rc" "locating the myflow begin delimiter in $target_file"
  rc=0; end_lines="$(grep -nFx "$CLAUDE_MD_END" "$target_file" | cut -d: -f1 | paste -sd, -)" || rc=$?
  require_grep_ok "$rc" "locating the myflow end delimiter in $target_file"
  rc=0; begins=$(grep -cFx "$CLAUDE_MD_BEGIN" "$target_file") || rc=$?
  require_grep_ok "$rc" "counting myflow begin delimiters in $target_file"
  rc=0; ends=$(grep -cFx "$CLAUDE_MD_END" "$target_file") || rc=$?
  require_grep_ok "$rc" "counting myflow end delimiters in $target_file"

  if [[ "$begins" -eq 0 && "$ends" -eq 0 ]]; then
    # The append is this installer's FIRST modification of a file it did not create, so the
    # backup belongs here too. Taking it only in the rewrite branch below means the copy
    # kept as "pre-myflow" already contains a generated block.
    backup_once "$target_file"
    { printf '\n'; cat "$block_tmp"; } >>"$target_file"
    info "Appended managed myflow block to $target_file (pre-myflow copy: $target_file.myflow.bak)"
    return 0
  fi

  # A single pair only counts if the begin actually precedes the end. Reversed
  # delimiters also count 1 and 1, and rewriting them deletes everything after
  # the begin marker.
  if [[ "$begins" -eq 1 && "$ends" -eq 1 && "$begin_lines" -lt "$end_lines" ]]; then
    local provenance
    if generated_only "$target_file" "$begin_lines" "$end_lines"; then
      provenance="no pre-myflow copy: this installer created the file"
    else
      backup_once "$target_file"
      provenance="pre-myflow copy: $target_file.myflow.bak"
    fi
    awk -v begin="$CLAUDE_MD_BEGIN" -v end="$CLAUDE_MD_END" -v blockfile="$block_tmp" '
      $0 == begin { inblock = 1; while ((getline l < blockfile) > 0) print l; next }
      $0 == end   { inblock = 0; next }
      !inblock    { print }
    ' "$target_file" >"$out_tmp"
    cat "$out_tmp" >"$target_file"
    info "Refreshed managed myflow block in $target_file ($provenance)"
    return 0
  fi

  die "myflow delimiters in $target_file are not one begin followed by one end
  $CLAUDE_MD_BEGIN at line(s): ${begin_lines:-none}
  $CLAUDE_MD_END at line(s): ${end_lines:-none}
  Rewriting could delete content between them and appending would never converge.
  Reconcile the file by hand — leave exactly one begin above exactly one end, or
  remove both markers — then re-run this installer."
}

# project_standards_entries <project.md>
# Print one `## standards` entry per line, in file order.
#
# An entry is a top-level list item; its value is the first backticked span on the item, or
# the item's first whitespace-delimited word when it has no backticks (both forms occur in
# real project files). Only a `## ` heading ends the section — a `### ` subsection is still
# inside it. Fenced blocks are skipped: an illustrative bullet inside a ```-fence is
# documentation, and reading it as an entry would have the installer chasing example names.
project_standards_entries() {
  awk '
    { sub(/\r$/, "") }
    /^```/                  { fence = !fence; next }
    fence                   { next }
    /^##[^#]/               { in_standards = ($0 ~ /^##[[:space:]]+standards[[:space:]]*$/); next }
    !in_standards           { next }
    /^[[:space:]]*[-*][[:space:]]+/ {
      line = $0
      sub(/^[[:space:]]*[-*][[:space:]]+/, "", line)
      if (match(line, /^`[^`]+`/)) {
        entry = substr(line, 2, RLENGTH - 2)
      } else {
        sub(/[[:space:]].*$/, "", line)
        entry = line
      }
      if (entry != "") print entry
    }
  ' "$1"
}

# install_project_standards <project-dir>
# Render the shared rules this project opted into — its `.myflow/project.md` `## standards`
# entries — into a managed block in BOTH <project>/CLAUDE.md and <project>/AGENTS.md.
#
# Opt-in rules are installed nowhere by path, on purpose: the Kotlin standard's globs would
# match every Kotlin repo on the machine. That left a project no way to actually load one
# except to paste a copy into each of its instruction files by hand — which is what gymie
# did, twice, with nothing keeping either copy in step with the rule. This is that copy,
# generated.
#
# Only entries that resolve to the SHARED library are rendered, per the resolution table in
# rules/myflow-manual-review.mdc: a bare filename (no `/`) ending in `.mdc`. An entry
# containing a `/` is a project path and is not a shared rule at all; any other bare filename
# is the project's own file, already in the repo. Always-on rules are dropped too — they
# reach every session through the global block, and rendering them again here is precisely
# the duplication this exists to remove.
install_project_standards() {
  local project_dir="$1"
  local project_md="$project_dir/.myflow/project.md"
  local skipped_before=$SKIPPED
  local entry rule_name always_list target
  local -a selected=() already_global=()

  [[ -f "$project_md" ]] || return 0

  always_list=$'\n'"$(always_on_rules)"$'\n'

  # The loop body must run in THIS shell, not a pipeline subshell: it increments SKIPPED,
  # and a skip recorded in a subshell is a skip the exit status never learns about.
  while IFS= read -r entry; do
    [[ "$entry" == *.mdc ]] || continue
    [[ "$entry" != */* ]] || continue
    # "Bare" was already established by the test above; this rejects the residue that has no
    # `/` yet is still not a filename (`..mdc`, a leading dot, a shell metacharacter), so
    # nothing but a plain name is ever concatenated onto the rule library path.
    if [[ ! "$entry" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*\.mdc$ ]]; then
      warn "$project_md names '$entry' under ## standards, which is not a plain rule filename — SKIPPED"
      SKIPPED=$((SKIPPED + 1))
      continue
    fi
    if [[ "$always_list" == *$'\n'"$entry"$'\n'* ]]; then
      already_global+=("$entry")
      continue
    fi
    if [[ ! -f "$RULES_SRC/$entry" ]]; then
      warn "$project_md names '$entry' under ## standards, but $RULES_SRC/$entry does not exist — SKIPPED, not rendered"
      SKIPPED=$((SKIPPED + 1))
      continue
    fi
    # A rule listed twice must still be rendered once; a second copy inside one block is the
    # same duplication, just closer together.
    for rule_name in ${selected[@]+"${selected[@]}"}; do
      [[ "$rule_name" != "$entry" ]] || continue 2
    done
    selected+=("$entry")
  done < <(project_standards_entries "$project_md")

  if (( ${#selected[@]} == 0 )); then
    # Nothing to render. Writing an empty block would create CLAUDE.md and AGENTS.md in
    # projects that never asked for either.
    (( SKIPPED == skipped_before )) || finish_banner "Project standards" "$skipped_before"
    return 0
  fi

  echo ""
  info "Rendering opt-in standards into $project_dir/CLAUDE.md and AGENTS.md (from .myflow/project.md)"
  for rule_name in "${selected[@]}"; do
    echo "  ✓ $rule_name"
  done
  for rule_name in ${already_global[@]+"${already_global[@]}"}; do
    echo "  – $rule_name is always-on — it already reaches every session through the global block"
  done

  # Same all-or-nothing preflight as install_global, for the same reason: validate every
  # rule body and stat every target before the first write, so a refusal cannot leave
  # CLAUDE.md rendered and AGENTS.md not — two harnesses reading the same project under
  # different rules.
  render_managed_block "${selected[@]}" >/dev/null
  for target in "$project_dir/CLAUDE.md" "$project_dir/AGENTS.md"; do
    preflight_managed_block "$target"
  done
  for target in "$project_dir/CLAUDE.md" "$project_dir/AGENTS.md"; do
    install_managed_block "$target" "${selected[@]}"
  done
  finish_banner "Project standards" "$skipped_before"
}

install_global() {
  local home_dir="${HOME:?HOME must be set}"
  local skipped_before=$SKIPPED
  # The managed-block targets, declared once so the preflight below and the install below
  # can never scan a different set than they write.
  #   - ~/.claude/CLAUDE.md — the only global rule layer Claude Code reads.
  #   - ~/.codex/AGENTS.md  — Codex reads neither .cursor/rules/ nor CLAUDE.md; without a
  #     managed block there, a global install left Codex sessions with no always-on rules at
  #     all, including the pipeline contract itself.
  local managed_files=("$home_dir/.claude/CLAUDE.md" "$home_dir/.codex/AGENTS.md")
  local managed_file
  info "Setting up myflow globally under $home_dir"

  # PREFLIGHT — everything that can refuse this run must refuse BEFORE the first symlink.
  # These checks used to fire midway: install_rules_cursor had already linked an offending
  # rule into ~/.cursor/rules/ by the time render_managed_block refused it, so Cursor picked
  # the rule up while Claude Code and Codex got no managed block at all — two harnesses
  # running the same repo under different rules, reproduced identically on every later run.
  # The refusals themselves are correct; only their ordering was wrong. Render into nothing purely
  # to validate every rule body, and stat the block targets, before installing anything.
  render_managed_block >/dev/null
  for managed_file in "${managed_files[@]}"; do
    preflight_managed_block "$managed_file"
  done

  install_skills "$home_dir/.claude/skills"
  # Cursor needs its own copy: every commands/ file resolves skills through
  # .cursor/skills/<skill>/SKILL.md, and there is no per-project copy any more.
  install_skills "$home_dir/.cursor/skills"
  # Codex too. The managed block written into ~/.codex/AGENTS.md below carries
  # myflow-manual-review.mdc, which names the /myflow-* skills throughout; without this a
  # global install told Codex the rules and left it nothing to resolve them against. The
  # per-project mode already uses .codex/skills, and projects may not keep their own copies
  # once a global install exists, so this is the only place the skills can come from.
  install_skills "$home_dir/.codex/skills"
  install_commands "$COMMANDS_CLAUDE_SRC" "$home_dir/.claude/commands"
  install_commands "$COMMANDS_CURSOR_SRC" "$home_dir/.cursor/commands"
  install_rules_cursor "$home_dir/.cursor/rules"
  # Claude Code's layer is two halves of one source: the core of each rule goes into the
  # managed block below, the full text is linked here, and the block's `Full rule:` pointer
  # is what joins them. Install the links BEFORE the block, so a pointer is never written to
  # a path that does not exist yet.
  install_rules_claude "$home_dir/.claude/rules"
  install_hooks "$home_dir/.claude"
  for managed_file in "${managed_files[@]}"; do
    install_managed_block "$managed_file"
  done
  echo ""
  finish_banner "Global" "$skipped_before"
  echo "   Skills   → $home_dir/.claude/skills/, $home_dir/.cursor/skills/ and $home_dir/.codex/skills/"
  echo "   Commands → $home_dir/.claude/commands/ and $home_dir/.cursor/commands/"
  echo "              No commands layer is installed for Codex — in a Codex session invoke a"
  echo "              skill by reading $home_dir/.codex/skills/<skill>/SKILL.md and following it."
  echo "   Rules    → $home_dir/.cursor/rules/, and the managed block in both"
  echo "              $home_dir/.claude/CLAUDE.md and $home_dir/.codex/AGENTS.md"
  echo "              Full texts → $home_dir/.claude/rules/, which the core excerpt in the"
  echo "              managed block points at. Same files, one source, no copies."
  echo "   Hooks    → $home_dir/.claude/hooks/ — installed, but registered only if"
  echo "              $home_dir/.claude/settings.json already names them (see any warning above)."
  echo "   Subagents inherit NO rules. Every dispatch must carry the two-sentence pointer to"
  echo "   $home_dir/.claude/rules/agent-baseline.md — that file is the whole rule set for a"
  echo "   dispatched agent, and it propagates itself to any depth."
  echo "   Opt-in rules (e.g. kotlin-backend-development-standard) are NOT installed globally."
  echo "   A project adopts one by naming it under ## standards in its own .myflow/project.md;"
  echo "   run './setup.sh <harness> /path/to/that/project' to render it into that project's"
  echo "   own CLAUDE.md and AGENTS.md. This mode never writes project files."
  echo ""
  echo "   Source   → $SCRIPT_DIR"
  echo "   Every link above points back into that directory, so global agent behavior"
  echo "   follows whatever is checked out there. Install from a stable main checkout —"
  echo "   from a git worktree or feature branch the links break when it is removed."
}

case "$HARNESS" in
  cursor)      install_cursor ;;
  global)      install_global ;;
  claude-code) install_claude_code ;;
  codex)       install_codex ;;
  all)
    install_cursor
    echo ""
    install_claude_code
    echo ""
    install_codex
    ;;
  *) die "Unknown harness '$HARNESS'. Choose: cursor | claude-code | codex | all | global" ;;
esac

# Opt-in standards are rendered once per run, here rather than inside each mode: `all` would
# otherwise render the same block three times and report each project skip three times. It
# runs AFTER the mode so that the CLAUDE.md / AGENTS.md a first-time `claude-code` or `codex`
# install copies in is the file the block lands in. `global` installs no project files at
# all, so it is deliberately not in this list — a user-level install must not start writing
# into whatever directory it happened to be run from.
case "$HARNESS" in
  cursor|claude-code|codex|all) install_project_standards "$PROJECT_DIR" ;;
esac

# A run that could not link something is not a successful install, and a caller (CI, a
# Makefile, a human reading $?) must be able to tell without parsing the output.
(( SKIPPED == 0 )) || exit 1
