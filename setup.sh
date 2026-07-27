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
# (`alwaysApply: true`). Every other rule is opt-in and is never installed
# anywhere by this script; a project activates it via its `.myflow/project.md`.
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

install_skills() {
  local target_dir="$1" skill_dir skill_name
  info "Installing project skills into $target_dir"
  mkdir -p "$target_dir"
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
  echo "   Next: install Superpowers for Codex from its fork repo."
  echo "   Enable: [features] multi_agent = true  in ~/.codex/config.toml"
}

# install_rules_cursor <target-dir>
# Installs the always-on rules, and only those — globally and per project alike.
# An opt-in rule (e.g. the Kotlin backend standard, whose globs would otherwise
# match a Compose Multiplatform or IDE-plugin repo) is never installed by path:
# whether a rule is always-on is decided by the rule, not by the caller.
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

install_commands() {
  local commands_src="$1"
  local target_dir="$2"
  local cmd_file cmd_name
  [[ -d "$commands_src" ]] || return 0
  info "Installing slash commands into $target_dir"
  mkdir -p "$target_dir"
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

# Render the always-on rules as the body of the managed block.
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
  local rule_name rule_file body
  printf '%s\n' "$CLAUDE_MD_BEGIN"
  printf '<!-- Generated by agents/setup.sh global. Do not edit inside this block: it is\n'
  printf '     rewritten on every install. Put your own notes outside the delimiters. -->\n'
  while IFS= read -r rule_name; do
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
    printf '\n<!-- rule: %s -->\n\n' "$rule_name"
    printf '%s\n' "$body"
  done < <(always_on_rules)
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

# Rewrite ONLY the delimited block in an agent instruction file
# (~/.claude/CLAUDE.md for Claude Code, ~/.codex/AGENTS.md for Codex).
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
  render_managed_block >"$block_tmp"

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
  echo "   Opt-in rules (e.g. kotlin-backend-development-standard) are NOT installed globally —"
  echo "   a project activates those through its own .myflow/project.md."
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

# A run that could not link something is not a successful install, and a caller (CI, a
# Makefile, a human reading $?) must be able to tell without parsing the output.
(( SKIPPED == 0 )) || exit 1
