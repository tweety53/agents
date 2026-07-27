#!/usr/bin/env bash
# test-setup.sh — regression harness for setup.sh.
#
# Usage: scripts/test-setup.sh
#
# Takes no arguments. Every case runs against a throwaway HOME created under /tmp, and the
# real ~/.claude, ~/.cursor and ~/.codex are fingerprinted before the first case and again
# after the last one — a mismatch fails the run. Nothing here writes outside the sandbox.
#
# WHY THIS EXISTS. `setup.sh` writes into the user's home directory, and two data-loss
# defects reached review:
#
#   1. Delimiters in reversed order (end above begin) counted 1 begin and 1 end, took the
#      rewrite branch, and silently deleted every line after the begin marker.
#   2. An unbalanced pair took an append branch that could never converge — 1/0 became 2/1
#      became 3/2, each run adding another copy of the block to the file injected into
#      every session.
#
# Both are fixed. Nothing in the repo would have caught either coming back, and nothing
# would catch them coming back again. That is what this file is for: the cases below are
# the shapes that actually broke, plus the containment guarantees whose failure is silent
# (an opt-in rule installed globally, a rule installed that never declared itself always-on,
# a stale skill copy left shadowing its own symlink).
#
# WHAT THIS CAN AND CANNOT PROVE — read before trusting a green run.
#
#   It proves things about the installer's OBSERVABLE FILESYSTEM EFFECTS: what exists after
#   a run, what it points at, what it contains, its mode, and the exit status. That is the
#   whole of its scope.
#
#   It does NOT prove the installed content is correct or current — that a rule says the
#   right thing, that a skill's prose is coherent, or that an agent reading the managed
#   block behaves as intended. `check-vocabulary.sh` covers a slice of the prose question;
#   agent behaviour is covered by neither and is a human job.
#
#   It does not cover `check-vocabulary.sh` (that script exercises itself), nor any harness
#   the installer does not write, nor the per-project modes beyond what the containment
#   cases need.
#
#   A green run means "the shapes listed below still behave as listed". Any shape nobody
#   thought to list passes clean, exactly as with the vocabulary guard.
#
# Two source trees are used, deliberately:
#
#   - The REAL repo, for the containment guarantees, which are claims about the rules and
#     skills actually shipped here (the opt-in Kotlin rule, the real skill set).
#   - A generated FIXTURE repo — a copy of setup.sh beside synthetic skills/rules/commands
#     — for the cases that need a malformed input the real tree must never contain (a rule
#     body carrying a block delimiter, an unterminated frontmatter).
#
# Set KEEP_SANDBOX=1 to leave the sandbox behind for inspection; it is removed otherwise,
# on success and on failure alike.

set -uo pipefail

die() {
  echo "ERROR: $*" >&2
  exit 2
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REAL_HOME="${HOME:?HOME must be set}"

[[ -x "$REPO_ROOT/setup.sh" ]] || die "setup.sh not found or not executable at $REPO_ROOT/setup.sh"

# The delimiters are read out of setup.sh rather than restated here. A second copy would
# drift from the first, and every case below depends on matching the installer's literals
# exactly — a stale copy would make the whole delimiter group pass vacuously.
read_literal() {
  local name="$1" line
  line="$(grep -m1 "^$name=" "$REPO_ROOT/setup.sh")" || die "no $name= assignment in setup.sh"
  line="${line#*=\'}"
  printf '%s' "${line%\'}"
}
BEGIN="$(read_literal CLAUDE_MD_BEGIN)"
END="$(read_literal CLAUDE_MD_END)"
[[ -n "$BEGIN" && -n "$END" && "$BEGIN" != "$END" ]] || die "could not read the block delimiters out of setup.sh"

# The sandbox is created under /tmp explicitly, not under $TMPDIR: on macOS $TMPDIR is a
# per-user path under /var/folders, and the safety claim this harness makes is the concrete
# one — every byte it writes is under /tmp and is removed again.
SANDBOX="$(mktemp -d /tmp/myflow-test-setup.XXXXXX)" || die "cannot create the sandbox under /tmp"
cleanup() {
  if [[ -n "${KEEP_SANDBOX:-}" ]]; then
    echo "KEEP_SANDBOX set — sandbox left at $SANDBOX" >&2
    return 0
  fi
  # Belt and braces: only ever remove a path this script created under /tmp.
  [[ "$SANDBOX" == /tmp/myflow-test-setup.* ]] && rm -rf "$SANDBOX"
}
trap cleanup EXIT

# ---------------------------------------------------------------------------
# Portability helpers. macOS ships BSD stat and no sha256sum; Linux the reverse.
# ---------------------------------------------------------------------------

# Probe the stat dialect ONCE, GNU first, and bind the helpers to the winner.
#
# Order matters and the obvious order is wrong. On GNU coreutils `stat -f` means
# --file-system: it EXITS 0 on any existing file and prints `?` for every directive it
# does not recognise, so a `stat -f … || stat -c …` fallback never fires on Linux. The
# damage is not just a wrong mode: stat_line would return the same constant string for
# every file, which silently turns the real-home fingerprint at the end of this run into
# a tautology — the one assertion that must never be vacuous. BSD stat rejects `-c`
# cleanly (exit 1), so probing GNU first is safe in both directions.
if stat -c '%a' . >/dev/null 2>&1; then
  file_mode() { stat -c '%a' "$1"; }
  stat_line() { stat -c '%n %s %Y %a' "$1"; }
else
  file_mode() { stat -f '%Lp' "$1"; }
  stat_line() { stat -f '%N %z %m %Lp' "$1"; }
fi

# A missing SHA tool would make both safety fingerprints the empty string and
# assert_eq "" "" would pass, reporting "nothing outside the sandbox was touched" having
# checked nothing. Refuse to run rather than report a vacuous pass.
if command -v shasum >/dev/null 2>&1; then
  hash_stdin() { shasum -a 256; }
elif command -v sha256sum >/dev/null 2>&1; then
  hash_stdin() { sha256sum; }
else
  die "no SHA-256 tool (shasum or sha256sum) available — the safety fingerprints would be vacuous"
fi

# ---------------------------------------------------------------------------
# Assertions. Every one names what it checked, so a failure line is actionable
# without re-reading this file.
# ---------------------------------------------------------------------------

PASSED=0
FAILED=0
FAILURES=()
GROUP="(none)"

group() { GROUP="$1"; printf '\n== %s ==\n' "$1"; }
pass() { PASSED=$((PASSED + 1)); printf '  ✓ %s\n' "$1"; }
fail() {
  FAILED=$((FAILED + 1))
  FAILURES+=("$GROUP — $1")
  printf '  ✗ %s\n      %s\n' "$1" "$2" >&2
}

assert_eq() { # <desc> <expected> <actual>
  if [[ "$2" == "$3" ]]; then pass "$1"; else fail "$1" "expected [$2], got [$3]"; fi
}
assert_exists() { # <desc> <path>
  if [[ -e "$2" || -L "$2" ]]; then pass "$1"; else fail "$1" "$2 does not exist"; fi
}
# Distinct from assert_exists: this one requires a regular FILE. Used where a directory at
# the same path would be the actual defect (a managed instruction file, never a directory).
assert_file_exists() { # <desc> <path>
  if [[ -f "$2" ]]; then pass "$1"; else fail "$1" "$2 is not a regular file"; fi
}
assert_absent() { # <desc> <path>
  if [[ -e "$2" || -L "$2" ]]; then fail "$1" "$2 exists and must not"; else pass "$1"; fi
}
assert_symlink() { # <desc> <path>
  if [[ -L "$2" ]]; then pass "$1"; else fail "$1" "$2 is not a symlink"; fi
}
assert_identical() { # <desc> <path-a> <path-b>
  if cmp -s "$2" "$3"; then pass "$1"; else fail "$1" "$2 and $3 differ"; fi
}
assert_contains() { # <desc> <file> <fixed-string>
  if grep -qF -- "$3" "$2" 2>/dev/null; then pass "$1"; else fail "$1" "[$3] not found in $2"; fi
}
assert_not_contains() { # <desc> <file> <fixed-string>
  if [[ ! -e "$2" ]] || ! grep -qF -- "$3" "$2" 2>/dev/null; then pass "$1"
  else fail "$1" "[$3] found in $2 and must not be"; fi
}
assert_rc_nonzero() { # <desc> <rc>
  if (( $2 != 0 )); then pass "$1"; else fail "$1" "exit status was 0; the run should have aborted"; fi
}
assert_rc_zero() { # <desc> <rc> <log>
  if (( $2 == 0 )); then pass "$1"; else fail "$1" "exit status $2 — see ${3:-log}"; fi
}

# count_lines_matching <file> <fixed-whole-line> — how many times a line occurs verbatim.
count_lines_matching() {
  local n
  n="$(grep -cFx -- "$2" "$1" 2>/dev/null)"
  printf '%s' "${n:-0}"
}

# ---------------------------------------------------------------------------
# Sandbox plumbing
# ---------------------------------------------------------------------------

CASE_SEQ=0
HOME_DIR=""

# new_home — a fresh empty HOME for one case, published in HOME_DIR.
#
# It assigns rather than echoing on purpose. `home="$(new_home)"` would run the body in a
# subshell, so the sequence counter would never advance in the caller and every case would
# be handed the SAME directory — each one then inspecting the accumulated leftovers of the
# ones before it. That is a harness whose isolation claim is false while it still passes.
new_home() {
  CASE_SEQ=$((CASE_SEQ + 1))
  HOME_DIR="$SANDBOX/home-$CASE_SEQ"
  mkdir -p "$HOME_DIR" || die "cannot create sandbox home $HOME_DIR"
}

# run_setup <repo> <home> <mode> [project-dir] — invoke the installer against a sandboxed
# HOME. RUN_RC and RUN_LOG are published for the caller.
#
# The HOME argument is re-checked here rather than trusted: this is the single point every
# case funnels through, so one check here covers all of them, and the thing being guarded
# against (a case that forgets to sandbox) is precisely a case whose own checks are absent.
RUN_RC=0
RUN_LOG=""
run_setup() {
  local repo="$1" home="$2" mode="$3" proj="${4:-$SANDBOX/cwd}"
  [[ "$home" == "$SANDBOX"/* ]] || die "refusing to run setup.sh with HOME=$home — not inside $SANDBOX"
  [[ "$proj" == "$SANDBOX"/* ]] || die "refusing to run setup.sh with project dir $proj — not inside $SANDBOX"
  mkdir -p "$proj"
  RUN_LOG="$SANDBOX/log-$CASE_SEQ-$mode.txt"
  ( cd "$proj" && HOME="$home" "$repo/setup.sh" "$mode" "$proj" ) >"$RUN_LOG" 2>&1
  RUN_RC=$?
}

# seed_file <path> <mode> — create a file from stdin with an explicit mode.
#
# Both guards below are load-bearing, and the second one was learned the hard way: an
# earlier version of this harness seeded a path that a previous case had already left as a
# symlink into the SOURCE tree, so the seed wrote straight through the link and overwrote a
# real skill in the repo under test. Every case now gets a fresh HOME, which removes that
# particular route — but a harness whose isolation depends on nobody making that mistake
# again is not isolated. Refusing to write through a link costs one test.
seed_file() {
  seed_guard "$1"
  mkdir -p "$(dirname "$1")"
  cat >"$1"
  chmod "$2" "$1"
}

seed_guard() {
  [[ "$1" == "$SANDBOX"/* ]] || die "refusing to seed $1 — outside the sandbox $SANDBOX"
  [[ ! -L "$1" ]] || die "refusing to seed $1 — it is a symlink, and writing through it would
  modify the link's target instead of the sandbox."
  local parent="$1"
  while parent="$(dirname "$parent")"; [[ "$parent" == "$SANDBOX"/* ]]; do
    [[ ! -L "$parent" ]] || die "refusing to seed $1 — its parent $parent is a symlink, and
  writing under it would modify the link's target instead of the sandbox."
  done
}

# seed_stale_skill_dir <path> — a real skill directory sitting where a symlink must go.
seed_stale_skill_dir() {
  seed_guard "$1"
  mkdir -p "$1" || die "cannot create $1"
  seed_guard "$1/SKILL.md"
  printf '# stale copy\nSENTINEL-PREEXISTING-SKILL\n' >"$1/SKILL.md"
}

# seed_project_md <project-dir> <standards-entry>… — a `.myflow/project.md` whose
# `## standards` section lists the given entries, one bullet each, in the backticked form
# real project files use. Sections before and after it are included so the section-boundary
# parsing is exercised rather than assumed.
seed_project_md() {
  local proj="$1"; shift
  local bullets="" entry
  for entry in "$@"; do bullets+="- \`$entry\` — seeded by the harness"$'\n'; done
  seed_file "$proj/.myflow/project.md" 644 <<EOF
# myflow project configuration — fixture project

## test

\`\`\`bash
./gradlew test
\`\`\`

## standards

Files the principles reviewer receives, and the rules this project opts into.

$bullets
A bullet inside a fenced block is illustration, not an entry:

\`\`\`text
- fenced-not-an-entry.mdc
\`\`\`

## jira

none
EOF
}

# make_fixture_repo <dir> [bad-rule] — a minimal agents repo the harness fully controls.
# The real tree cannot host a rule whose body carries a block delimiter or an unterminated
# frontmatter, and those are exactly the inputs two of the guarantees are about.
make_fixture_repo() {
  local d="$1" bad="${2:-}"
  mkdir -p "$d/rules" "$d/commands" "$d/commands-claude" "$d/skills/demo-skill"
  cp "$REPO_ROOT/setup.sh" "$d/setup.sh" || die "cannot copy setup.sh into the fixture repo"
  chmod +x "$d/setup.sh"
  printf '# demo skill\n' >"$d/skills/demo-skill/SKILL.md"
  printf '# fixture CLAUDE.md\n' >"$d/CLAUDE.md"
  printf '# fixture AGENTS.md\n' >"$d/AGENTS.md"
  printf '# demo command\n' >"$d/commands/demo.md"
  printf '# demo command\n' >"$d/commands-claude/demo.md"

  # Always-on: the one rule that must install.
  printf -- '---\ndescription: fixture always-on rule\nalwaysApply: true\n---\n\n# Good\nBODY-GOOD-ALWAYS\n' \
    >"$d/rules/good-always.mdc"
  # Opt-in: declares false, must never install.
  printf -- '---\ndescription: fixture opt-in rule\nalwaysApply: false\n---\n\n# OptIn\nBODY-OPT-IN\n' \
    >"$d/rules/opt-in-false.mdc"
  # Frontmatter that never closes: the `alwaysApply: true` below is prose, not a declaration.
  printf -- '---\ndescription: fixture unterminated frontmatter\nalwaysApply: true\n\n# Unterminated\nBODY-UNTERMINATED\n' \
    >"$d/rules/unterminated.mdc"
  # A value that merely starts with `true` must not be read as `true`.
  printf -- '---\ndescription: fixture tricky value\nalwaysApply: true_for_kotlin_only\n---\n\n# Tricky\nBODY-TRICKY-PREFIX\n' \
    >"$d/rules/tricky-prefix.mdc"

  if [[ -n "$bad" ]]; then
    # An always-on rule whose BODY carries a bare begin delimiter. Inlining it would put a
    # second delimiter inside the managed block and corrupt every later run.
    printf -- '---\ndescription: fixture rule carrying a delimiter\nalwaysApply: true\n---\n\n# Bad\n%s\nBODY-BAD-DELIMITER\n' \
      "$BEGIN" >"$d/rules/bad-delimiter.mdc"
  fi
}

# ---------------------------------------------------------------------------
# Real-home fingerprint. Taken before the first case and compared after the last.
#
# Only the paths setup.sh can write are sampled — the three harness directories, their
# skills/commands/rules subtrees one level down, the two managed instruction files, and any
# .myflow.bak beside them. A full recursive listing of ~/.claude would be dominated by
# session state that changes on its own while this runs, which would make the check noisy
# and therefore ignored.
# ---------------------------------------------------------------------------

real_home_fingerprint() {
  local d p entry
  for d in .claude .cursor .codex; do
    for p in "$REAL_HOME/$d" "$REAL_HOME/$d/skills" "$REAL_HOME/$d/commands" "$REAL_HOME/$d/rules"; do
      if [[ -d "$p" ]]; then
        printf 'dir %s\n' "$p"
        for entry in "$p"/* "$p"/.*; do
          [[ -e "$entry" || -L "$entry" ]] || continue
          printf '  %s\n' "${entry##*/}"
        done
      else
        printf 'absent %s\n' "$p"
      fi
    done
    for p in "$REAL_HOME/$d/CLAUDE.md" "$REAL_HOME/$d/AGENTS.md" \
             "$REAL_HOME/$d/CLAUDE.md.myflow.bak" "$REAL_HOME/$d/AGENTS.md.myflow.bak"; do
      if [[ -e "$p" ]]; then stat_line "$p"; else printf 'absent %s\n' "$p"; fi
    done
  done
}

# The SOURCE tree is fingerprinted too, and for a reason this harness demonstrated on
# itself: a global install fills the sandbox with symlinks pointing back into this repo, so
# any sandbox path a case writes may in fact be a repo path. The seed_guard above blocks the
# route that was found; this catches the routes that were not.
source_tree_fingerprint() {
  local f
  # scripts/ and README.md are included deliberately: without them the harness cannot
  # detect damage to ITSELF or to the vocabulary guard, which is the one blind spot a
  # write-through incident would exploit twice.
  find "$REPO_ROOT/skills" "$REPO_ROOT/rules" "$REPO_ROOT/commands" "$REPO_ROOT/commands-claude" \
       "$REPO_ROOT/scripts" \
       "$REPO_ROOT/setup.sh" "$REPO_ROOT/CLAUDE.md" "$REPO_ROOT/AGENTS.md" "$REPO_ROOT/README.md" \
       -type f -print0 2>/dev/null | sort -z | while IFS= read -r -d '' f; do
    stat_line "$f"
  done
}

REAL_HOME_BEFORE="$(real_home_fingerprint | hash_stdin)"
SOURCE_TREE_BEFORE="$(source_tree_fingerprint | hash_stdin)"

echo "setup.sh regression harness"
echo "  repo    : $REPO_ROOT"
echo "  sandbox : $SANDBOX"

# ===========================================================================
# Group 1 — malformed delimiters must abort with the target untouched.
#
# These are the two data-loss defects. Each case seeds a managed file with a shape the
# installer must refuse, runs `global`, and requires both a non-zero exit AND a target that
# is byte-for-byte what it was. "Aborted" is not enough on its own: the reversed-delimiter
# defect aborted nothing and deleted content, and a future one could abort *after* writing.
# ===========================================================================

group "Malformed delimiters abort with the target byte-identical"

FIXTURE="$SANDBOX/fixture-repo"
make_fixture_repo "$FIXTURE"

# --- reversed: end above begin. Counts 1 and 1, which is what made this dangerous.
new_home; home="$HOME_DIR"
target="$home/.claude/CLAUDE.md"
seed_file "$target" 644 <<EOF
# My own notes
keep-me-above

$END
generated-looking content
$BEGIN

# More of my own notes
keep-me-below
EOF
cp -p "$target" "$SANDBOX/reversed-expected.md"
run_setup "$FIXTURE" "$home" global
assert_rc_nonzero "reversed delimiters abort the run" "$RUN_RC"
assert_contains "the abort names the delimiter shape" "$RUN_LOG" "not one begin followed by one end"
assert_identical "reversed: CLAUDE.md is byte-identical afterwards" "$target" "$SANDBOX/reversed-expected.md"
assert_absent "reversed: no .myflow.bak was written" "$target.myflow.bak"

# --- duplicated: two begins and two ends, seeded in the CODEX target so the second
# managed file is exercised too (the first one is written before this is reached).
new_home; home="$HOME_DIR"
target="$home/.codex/AGENTS.md"
seed_file "$target" 644 <<EOF
# My own notes

$BEGIN
first block
$END

$BEGIN
second block
$END
EOF
cp -p "$target" "$SANDBOX/duplicated-expected.md"
run_setup "$FIXTURE" "$home" global
assert_rc_nonzero "duplicated delimiters abort the run" "$RUN_RC"
assert_identical "duplicated: AGENTS.md is byte-identical afterwards" "$target" "$SANDBOX/duplicated-expected.md"
assert_absent "duplicated: no .myflow.bak was written" "$target.myflow.bak"

# --- CRLF markers: the installer matches LF-terminated lines, so these count zero and
# would take the append branch. It must stop and say how to fix the file.
new_home; home="$HOME_DIR"
target="$home/.claude/CLAUDE.md"
{
  printf '# My own notes\r\n\r\n'
  printf '%s\r\n' "$BEGIN"
  printf 'old block\r\n'
  printf '%s\r\n' "$END"
} >"$SANDBOX/crlf-expected.md"
mkdir -p "$(dirname "$target")"
cp -p "$SANDBOX/crlf-expected.md" "$target"
run_setup "$FIXTURE" "$home" global
assert_rc_nonzero "CRLF delimiters abort the run" "$RUN_RC"
assert_contains "the CRLF abort carries the remediation hint" "$RUN_LOG" "Convert the file to LF endings"
assert_identical "CRLF: CLAUDE.md is byte-identical afterwards" "$target" "$SANDBOX/crlf-expected.md"

# ===========================================================================
# Group 2 — repeated runs converge, and the user's own content survives them.
#
# The append-forever defect showed up only on the SECOND and THIRD run, so one run proves
# nothing here. Three runs, counting the delimiters in both managed files each time, and
# requiring run 3 to be byte-identical to run 2 — the definition of converged.
# ===========================================================================

group "Three consecutive global runs converge"

new_home; home="$HOME_DIR"
claude_md="$home/.claude/CLAUDE.md"
codex_md="$home/.codex/AGENTS.md"

for run in 1 2 3; do
  run_setup "$FIXTURE" "$home" global
  assert_rc_zero "run $run succeeds" "$RUN_RC" "$RUN_LOG"
  assert_eq "run $run: exactly one begin in CLAUDE.md" 1 "$(count_lines_matching "$claude_md" "$BEGIN")"
  assert_eq "run $run: exactly one end in CLAUDE.md" 1 "$(count_lines_matching "$claude_md" "$END")"
  assert_eq "run $run: exactly one begin in AGENTS.md" 1 "$(count_lines_matching "$codex_md" "$BEGIN")"
  assert_eq "run $run: exactly one end in AGENTS.md" 1 "$(count_lines_matching "$codex_md" "$END")"
  if [[ "$run" == 2 ]]; then
    cp -p "$claude_md" "$SANDBOX/converge-claude-run2.md"
    cp -p "$codex_md" "$SANDBOX/converge-codex-run2.md"
  fi
done
assert_identical "run 3 leaves CLAUDE.md byte-identical to run 2" "$claude_md" "$SANDBOX/converge-claude-run2.md"
assert_identical "run 3 leaves AGENTS.md byte-identical to run 2" "$codex_md" "$SANDBOX/converge-codex-run2.md"

group "Hand-written content and file mode survive every run"

new_home; home="$HOME_DIR"
claude_md="$home/.claude/CLAUDE.md"
# Mode 640 rather than 600 on purpose. 600 is what `mktemp` and a fresh `cp` both produce
# anyway, so a rewrite that replaced the file wholesale, or a backup taken with a shell
# redirect, would land on 600 by accident and these two assertions could never fail. 640
# matches nothing the installer would create on its own, so preserving it has to be
# deliberate.
seed_file "$claude_md" 640 <<'EOF'
# My own global instructions

HANDWRITTEN-ABOVE — this line predates myflow.
EOF
cp -p "$claude_md" "$SANDBOX/handwritten-original.md"

run_setup "$FIXTURE" "$home" global
assert_rc_zero "first run over a hand-written file succeeds" "$RUN_RC" "$RUN_LOG"
assert_contains "run 1 keeps the hand-written line" "$claude_md" "HANDWRITTEN-ABOVE"
assert_exists "run 1 takes a pre-myflow backup" "$claude_md.myflow.bak"
assert_identical "the backup is the file as it was before myflow touched it" \
  "$claude_md.myflow.bak" "$SANDBOX/handwritten-original.md"

# Content added AFTER the end marker exercises the rewrite branch's blast radius — this is
# what the reversed-delimiter defect destroyed.
printf '\nHANDWRITTEN-BELOW — added after the managed block.\n' >>"$claude_md"

for run in 2 3; do
  run_setup "$FIXTURE" "$home" global
  assert_rc_zero "run $run over a hand-written file succeeds" "$RUN_RC" "$RUN_LOG"
  assert_contains "run $run keeps content above the block" "$claude_md" "HANDWRITTEN-ABOVE"
  assert_contains "run $run keeps content below the block" "$claude_md" "HANDWRITTEN-BELOW"
  assert_eq "run $run: still exactly one begin" 1 "$(count_lines_matching "$claude_md" "$BEGIN")"
  assert_eq "run $run: still exactly one end" 1 "$(count_lines_matching "$claude_md" "$END")"
  assert_eq "run $run: the target keeps its 640 mode" 640 "$(file_mode "$claude_md")"
  assert_eq "run $run: the .myflow.bak keeps the 640 mode" 640 "$(file_mode "$claude_md.myflow.bak")"
  assert_identical "run $run: the pre-myflow backup is never overwritten" \
    "$claude_md.myflow.bak" "$SANDBOX/handwritten-original.md"
  if [[ "$run" == 2 ]]; then cp -p "$claude_md" "$SANDBOX/handwritten-run2.md"; fi
done
assert_identical "run 3 leaves the hand-written file byte-identical to run 2" \
  "$claude_md" "$SANDBOX/handwritten-run2.md"

# ===========================================================================
# Group 3 — all-or-nothing preflight.
#
# A rule the renderer must refuse has to be refused BEFORE anything is installed. A partial
# install is the worst outcome available: Cursor picks up rules that Claude Code and Codex
# never receive, and every later run reproduces the split identically.
# ===========================================================================

group "A rule carrying a delimiter installs nothing at all"

BAD_FIXTURE="$SANDBOX/fixture-repo-bad-rule"
make_fixture_repo "$BAD_FIXTURE" with-bad-rule
new_home; home="$HOME_DIR"
run_setup "$BAD_FIXTURE" "$home" global
assert_rc_nonzero "a rule body carrying a delimiter aborts the run" "$RUN_RC"
assert_contains "the abort names the offending rule" "$RUN_LOG" "bad-delimiter.mdc"
for leftover in \
  "$home/.claude/skills" "$home/.cursor/skills" "$home/.codex/skills" \
  "$home/.claude/commands" "$home/.cursor/commands" "$home/.cursor/rules" \
  "$home/.claude/CLAUDE.md" "$home/.codex/AGENTS.md"; do
  assert_absent "nothing installed: ${leftover#"$home"/}" "$leftover"
done

# ===========================================================================
# Group 4 — containment. Each of these fails silently if it regresses: an install that
# looks completely successful while shipping a rule it must not, or shadowing a symlink
# with a stale copy of the same skill.
# ===========================================================================

group "Only rules declaring alwaysApply: true install"

new_home; home="$HOME_DIR"
run_setup "$FIXTURE" "$home" global
assert_rc_zero "the fixture install succeeds" "$RUN_RC" "$RUN_LOG"
assert_exists "the always-on rule is installed" "$home/.cursor/rules/good-always.mdc"
assert_absent "alwaysApply: false is not installed" "$home/.cursor/rules/opt-in-false.mdc"
assert_absent "an unterminated frontmatter is not installed" "$home/.cursor/rules/unterminated.mdc"
assert_absent "alwaysApply: true_for_kotlin_only is not installed" "$home/.cursor/rules/tricky-prefix.mdc"
rule_count=0
for f in "$home/.cursor/rules"/*; do [[ -e "$f" || -L "$f" ]] && rule_count=$((rule_count + 1)); done
assert_eq "exactly one rule is installed" 1 "$rule_count"
assert_contains "the managed block carries the always-on rule body" "$home/.claude/CLAUDE.md" "BODY-GOOD-ALWAYS"
assert_not_contains "the managed block omits the opt-in rule" "$home/.claude/CLAUDE.md" "BODY-OPT-IN"
assert_not_contains "the managed block omits the unterminated rule" "$home/.claude/CLAUDE.md" "BODY-UNTERMINATED"
assert_not_contains "the managed block omits the true-prefixed rule" "$home/.claude/CLAUDE.md" "BODY-TRICKY-PREFIX"

group "The opt-in Kotlin rule is installed by no mode"

KOTLIN_RULE="kotlin-backend-development-standard.mdc"
if [[ ! -f "$REPO_ROOT/rules/$KOTLIN_RULE" ]]; then
  fail "the opt-in Kotlin rule is present to be tested" "$REPO_ROOT/rules/$KOTLIN_RULE is missing"
else
  for mode in global cursor claude-code codex all; do
    new_home; home="$HOME_DIR"
    proj="$SANDBOX/project-$CASE_SEQ"
    run_setup "$REPO_ROOT" "$home" "$mode" "$proj"
    assert_rc_zero "mode $mode installs cleanly" "$RUN_RC" "$RUN_LOG"
    hits="$(find "$home" "$proj" -name "$KOTLIN_RULE" 2>/dev/null)"
    assert_eq "mode $mode installs no $KOTLIN_RULE" "" "$hits"
    # Only `global` writes a managed block. For the other four modes the files do not
    # exist, and assert_not_contains passes on a missing file — so asserting "no Kotlin
    # rule inlined" there proves nothing. Assert the real property per mode instead:
    # global must have the block WITHOUT the opt-in rule; the rest must have no block.
    for managed in "$home/.claude/CLAUDE.md" "$home/.codex/AGENTS.md"; do
      if [[ "$mode" == "global" ]]; then
        assert_file_exists "mode $mode: ${managed#"$home"/} exists" "$managed"
        assert_not_contains "mode $mode: ${managed#"$home"/} has no inlined Kotlin rule" \
          "$managed" "<!-- rule: $KOTLIN_RULE -->"
      else
        assert_absent "mode $mode writes no managed block at ${managed#"$home"/}" "$managed"
      fi
    done
  done
fi

group "A global install populates all three skill directories with live links"

new_home; home="$HOME_DIR"
run_setup "$REPO_ROOT" "$home" global
assert_rc_zero "the real-repo global install succeeds" "$RUN_RC" "$RUN_LOG"

expected_skills=0
for d in "$REPO_ROOT/skills"/*/; do [[ -d "$d" ]] && expected_skills=$((expected_skills + 1)); done
if (( expected_skills == 0 )); then
  fail "the repo has skills to install" "$REPO_ROOT/skills contains no directories"
fi
for skills_dir in "$home/.claude/skills" "$home/.cursor/skills" "$home/.codex/skills"; do
  installed=0
  linked=0
  for entry in "$skills_dir"/*; do
    [[ -e "$entry" || -L "$entry" ]] || continue
    installed=$((installed + 1))
    [[ -L "$entry" && -d "$entry" ]] && linked=$((linked + 1))
  done
  assert_eq "${skills_dir#"$home"/} holds every skill" "$expected_skills" "$installed"
  assert_eq "${skills_dir#"$home"/} holds only resolvable symlinks" "$expected_skills" "$linked"
done

dangling="$(find "$home" -type l ! -exec test -e {} \; -print 2>/dev/null)"
assert_eq "a global install leaves zero dangling symlinks" "" "$dangling"

group "A pre-existing skill directory is moved outside the scanned tree"

new_home; home="$HOME_DIR"
victim=""
for d in "$REPO_ROOT/skills"/*/; do victim="$(basename "$d")"; break; done
[[ -n "$victim" ]] || die "cannot pick a skill name to displace"
seed_stale_skill_dir "$home/.claude/skills/$victim"

run_setup "$REPO_ROOT" "$home" global
assert_rc_zero "the install over a pre-existing skill directory succeeds" "$RUN_RC" "$RUN_LOG"
assert_symlink "the displaced skill is now a symlink" "$home/.claude/skills/$victim"
# `<name>.bak` beside it would still be discovered by a SKILL.md walk — two loadable skills
# declaring the same name. The backup has to leave the scanned tree entirely.
assert_absent "no <name>.bak is left inside the skills tree" "$home/.claude/skills/$victim.bak"
backup_hits="$(find -L "$home/.claude/skills-backup" -name SKILL.md -print0 2>/dev/null |
  xargs -0 grep -lF SENTINEL-PREEXISTING-SKILL 2>/dev/null | wc -l | tr -d '[:space:]')"
assert_eq "the stale copy is preserved outside the scanned tree" 1 "$backup_hits"
scanned_hits="$(find -L "$home/.claude/skills" -name SKILL.md -print0 2>/dev/null |
  xargs -0 grep -lF SENTINEL-PREEXISTING-SKILL 2>/dev/null | wc -l | tr -d '[:space:]')"
assert_eq "the stale copy is no longer reachable from the skills tree" 0 "$scanned_hits"

# ===========================================================================
# Group 5 — project-level opt-in rule rendering.
#
# An opt-in rule is installed nowhere by path, so before this existed the only way a project
# could actually load one was to paste a copy into its own CLAUDE.md and AGENTS.md — two
# copies with nothing keeping them in step with the rule they came from. The installer now
# renders whatever the project's `.myflow/project.md ## standards` section names into a
# managed block in both files.
#
# Everything below is a containment or convergence claim about a file the user WROTE
# (a project's CLAUDE.md is hand-maintained far more often than ~/.claude/CLAUDE.md is), so
# the same guarantees the global block carries have to hold here too.
# ===========================================================================

group "A project's opted-in shared rule is rendered into both its instruction files"

new_home; home="$HOME_DIR"
proj="$SANDBOX/project-optin-$CASE_SEQ"
seed_project_md "$proj" "CLAUDE.md" "opt-in-false.mdc" "good-always.mdc"
# `cursor` on purpose: it copies neither CLAUDE.md nor AGENTS.md into the project, so both
# files exist below only because the standards rendering created them. Under `claude-code`
# the CLAUDE.md existence assertion could not fail whatever this feature did.
run_setup "$FIXTURE" "$home" cursor "$proj"
assert_rc_zero "the project install succeeds" "$RUN_RC" "$RUN_LOG"
for f in "$proj/CLAUDE.md" "$proj/AGENTS.md"; do
  label="${f#"$proj"/}"
  assert_file_exists "$label exists after the install" "$f"
  assert_contains "$label carries the opted-in rule body" "$f" "BODY-OPT-IN"
  assert_contains "$label labels the rule it inlined" "$f" "<!-- rule: opt-in-false.mdc -->"
  assert_eq "$label has exactly one begin" 1 "$(count_lines_matching "$f" "$BEGIN")"
  assert_eq "$label has exactly one end" 1 "$(count_lines_matching "$f" "$END")"
  # An always-on rule already reaches every session through the global block. Rendering it
  # again here is the duplication this feature exists to remove, not a harmless extra.
  assert_not_contains "$label does not re-render the always-on rule" "$f" "BODY-GOOD-ALWAYS"
  # `CLAUDE.md` is a bare non-.mdc entry — the project's own file, not a shared rule.
  assert_not_contains "$label treats a bare non-.mdc entry as no shared rule" "$f" "<!-- rule: CLAUDE.md -->"
done
assert_contains "the run reports which rule it rendered" "$RUN_LOG" "opt-in-false.mdc"

group "Project rendering is idempotent and preserves hand-written content"

new_home; home="$HOME_DIR"
proj="$SANDBOX/project-idem-$CASE_SEQ"
seed_project_md "$proj" "opt-in-false.mdc"
seed_file "$proj/CLAUDE.md" 640 <<'EOF'
# Gymie-like project instructions

PROJECT-HANDWRITTEN — this line predates the managed block.
EOF
cp -p "$proj/CLAUDE.md" "$SANDBOX/project-handwritten-original.md"
for run in 1 2 3; do
  run_setup "$FIXTURE" "$home" all "$proj"
  assert_rc_zero "project run $run succeeds" "$RUN_RC" "$RUN_LOG"
  assert_contains "project run $run keeps the hand-written line" "$proj/CLAUDE.md" "PROJECT-HANDWRITTEN"
  assert_eq "project run $run: exactly one begin in CLAUDE.md" 1 "$(count_lines_matching "$proj/CLAUDE.md" "$BEGIN")"
  assert_eq "project run $run: exactly one begin in AGENTS.md" 1 "$(count_lines_matching "$proj/AGENTS.md" "$BEGIN")"
  assert_eq "project run $run: CLAUDE.md keeps its 640 mode" 640 "$(file_mode "$proj/CLAUDE.md")"
  if [[ "$run" == 2 ]]; then
    cp -p "$proj/CLAUDE.md" "$SANDBOX/project-idem-claudefile-run2.md"
    cp -p "$proj/AGENTS.md" "$SANDBOX/project-idem-agentsfile-run2.md"
  fi
done
assert_identical "project run 3 leaves CLAUDE.md byte-identical to run 2" \
  "$proj/CLAUDE.md" "$SANDBOX/project-idem-claudefile-run2.md"
assert_identical "project run 3 leaves AGENTS.md byte-identical to run 2" \
  "$proj/AGENTS.md" "$SANDBOX/project-idem-agentsfile-run2.md"
assert_identical "the project's pre-myflow copy is the file as the user wrote it" \
  "$proj/CLAUDE.md.myflow.bak" "$SANDBOX/project-handwritten-original.md"

group "Which modes render project standards"

for mode in cursor claude-code codex all; do
  new_home; home="$HOME_DIR"
  proj="$SANDBOX/project-mode-$CASE_SEQ"
  seed_project_md "$proj" "opt-in-false.mdc"
  run_setup "$FIXTURE" "$home" "$mode" "$proj"
  assert_rc_zero "mode $mode installs cleanly over a project with standards" "$RUN_RC" "$RUN_LOG"
  assert_contains "mode $mode renders the rule into CLAUDE.md" "$proj/CLAUDE.md" "BODY-OPT-IN"
  assert_contains "mode $mode renders the rule into AGENTS.md" "$proj/AGENTS.md" "BODY-OPT-IN"
done

# `global` installs no project files at all, so it must not create these either — otherwise a
# user-level install would start writing into whatever directory it happened to be run from.
new_home; home="$HOME_DIR"
proj="$SANDBOX/project-global-$CASE_SEQ"
seed_project_md "$proj" "opt-in-false.mdc"
run_setup "$FIXTURE" "$home" global "$proj"
assert_rc_zero "global installs cleanly beside a project with standards" "$RUN_RC" "$RUN_LOG"
assert_absent "global writes no project CLAUDE.md" "$proj/CLAUDE.md"
assert_absent "global writes no project AGENTS.md" "$proj/AGENTS.md"

group "A project with nothing to render is left alone, silently"

# No .myflow/project.md at all.
new_home; home="$HOME_DIR"
proj="$SANDBOX/project-noconfig-$CASE_SEQ"
mkdir -p "$proj"
run_setup "$FIXTURE" "$home" cursor "$proj"
assert_rc_zero "a project with no .myflow/project.md installs cleanly" "$RUN_RC" "$RUN_LOG"
assert_absent "no CLAUDE.md is created for a project with no config" "$proj/CLAUDE.md"
assert_absent "no AGENTS.md is created for a project with no config" "$proj/AGENTS.md"

# A project.md that names only its own files — no shared rule to render, so no block.
new_home; home="$HOME_DIR"
proj="$SANDBOX/project-noshared-$CASE_SEQ"
seed_project_md "$proj" "CLAUDE.md" "CONTRIBUTING.md" "docs/standards/api.mdc"
run_setup "$FIXTURE" "$home" cursor "$proj"
assert_rc_zero "a project naming no shared rule installs cleanly" "$RUN_RC" "$RUN_LOG"
assert_absent "no block is written when no entry resolves to the shared library" "$proj/CLAUDE.md"
# `docs/standards/api.mdc` contains a `/`, so it is a project path — form 3, never the
# shared library. Resolving it there would be the containment bypass.
assert_not_contains "a slashed .mdc entry is not looked up in the shared library" "$RUN_LOG" "docs/standards/api.mdc"

group "A named rule that does not exist is reported and skipped"

new_home; home="$HOME_DIR"
proj="$SANDBOX/project-missing-$CASE_SEQ"
seed_project_md "$proj" "no-such-rule.mdc" "opt-in-false.mdc"
run_setup "$FIXTURE" "$home" claude-code "$proj"
assert_contains "the missing rule is reported by name" "$RUN_LOG" "no-such-rule.mdc"
assert_contains "the rules that do exist are still rendered" "$proj/CLAUDE.md" "BODY-OPT-IN"
assert_not_contains "a bullet inside a fenced block is not read as an entry" "$RUN_LOG" "fenced-not-an-entry.mdc"
# Reported, not fatal: the rest of the install completed. But a run that could not deliver
# something the project asked for is not a clean install either, and `$?` is where that has
# to show — same contract as every other skip in this installer.
assert_rc_nonzero "a missing rule makes the run report a skip in its exit status" "$RUN_RC"
assert_contains "both instruction files still got the rules that do exist" "$proj/AGENTS.md" "BODY-OPT-IN"

group "The delimiter guards fire on a project file exactly as on a global one"

new_home; home="$HOME_DIR"
proj="$SANDBOX/project-reversed-$CASE_SEQ"
seed_project_md "$proj" "opt-in-false.mdc"
seed_file "$proj/CLAUDE.md" 644 <<EOF
# Project notes

$END
generated-looking content
$BEGIN

keep-me-below
EOF
cp -p "$proj/CLAUDE.md" "$SANDBOX/project-reversed-expected.md"
run_setup "$FIXTURE" "$home" claude-code "$proj"
assert_rc_nonzero "reversed delimiters in a project file abort the run" "$RUN_RC"
assert_contains "the project abort names the delimiter shape" "$RUN_LOG" "not one begin followed by one end"
assert_identical "reversed: the project CLAUDE.md is byte-identical afterwards" \
  "$proj/CLAUDE.md" "$SANDBOX/project-reversed-expected.md"
assert_absent "reversed: no project .myflow.bak was written" "$proj/CLAUDE.md.myflow.bak"

group "The real Kotlin standard reaches a project that opts into it"

new_home; home="$HOME_DIR"
proj="$SANDBOX/project-kotlin-$CASE_SEQ"
seed_project_md "$proj" "CLAUDE.md" "$KOTLIN_RULE"
run_setup "$REPO_ROOT" "$home" claude-code "$proj"
assert_rc_zero "the real-repo project install succeeds" "$RUN_RC" "$RUN_LOG"
for f in "$proj/CLAUDE.md" "$proj/AGENTS.md"; do
  label="${f#"$proj"/}"
  assert_contains "$label inlines the Kotlin standard" "$f" "<!-- rule: $KOTLIN_RULE -->"
  assert_contains "$label carries the Kotlin standard's text" "$f" "Kotlin Backend Development Standard"
  # The frontmatter is stripped exactly as it is for the global block; leaving `globs:` or
  # `alwaysApply:` in the rendered body would feed a Cursor-only header to every harness.
  assert_not_contains "$label strips the rule's frontmatter" "$f" "alwaysApply: false"
done

# ===========================================================================
# Safety close-out. Every case above ran with a sandboxed HOME; this proves it.
# ===========================================================================

group "Nothing outside the sandbox was touched"

REAL_HOME_AFTER="$(real_home_fingerprint | hash_stdin)"
assert_eq "~/.claude, ~/.cursor and ~/.codex are unchanged" "$REAL_HOME_BEFORE" "$REAL_HOME_AFTER"
SOURCE_TREE_AFTER="$(source_tree_fingerprint | hash_stdin)"
assert_eq "the repo's own skills, rules and commands are unchanged" "$SOURCE_TREE_BEFORE" "$SOURCE_TREE_AFTER"

# ===========================================================================

printf '\n----------------------------------------\n'
if (( FAILED == 0 )); then
  printf '✓ PASS — %d assertions, 0 failures\n' "$PASSED"
  exit 0
fi
printf '✗ FAIL — %d passed, %d failed\n\n' "$PASSED" "$FAILED"
for f in "${FAILURES[@]}"; do printf '  • %s\n' "$f"; done
exit 1
