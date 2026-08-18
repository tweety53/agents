#!/usr/bin/env bash
# check-guard-symlinks.sh — the guard that keeps KAN-73's tasks 1-3 true.
#
# Argument-free and self-scoped, exactly like check-references.sh: the scan set
# is this repository's own skills/ and scripts/ directories, resolved from this
# script's own location, so no call site can narrow it.
# CHECK_GUARD_SYMLINKS_ROOT is an explicit, opt-in override honored only when
# set (mirrors CHECK_REFERENCES_ROOT), so the companion harness
# (test-check-guard-symlinks.sh) can point this guard at a sandboxed fixture
# tree under TMPDIR without touching this repository — never set it for a
# normal invocation.
#
# Four rules, each reported by name with the offending path and line:
#
#   1. Every entry under skills/*/scripts/ is a symlink, it resolves, and its
#      target is relative — an absolute target would bake this machine's
#      checkout path into the repository.
#   2. Every guard INVOKED in a skill's own text — a leading word in a
#      ```bash/sh/zsh fenced command line, or a backtick-quoted basename a
#      nearby "Run"/"Invoke"/"Execute"/"invocation" names, or one followed by a
#      `<placeholder>` usage argument — has a symlink in that skill's own
#      scripts/ directory. A guard's sibling dependency — read from the guard's
#      OWN source rather than a hardcoded table, by grepping it for
#      `$SCRIPT_DIR/<name>` — is required exactly where the guard it belongs
#      beside is required.
#   3. No skill text carries a repository-relative `scripts/<name>` path in an
#      invoking position — a non-comment line inside a bash/sh/zsh fence, or a
#      "Run"/"Invoke"/"Execute" immediately before the backtick. Prose about
#      this repository's own guards is exempt, per task 3's boundary: it never
#      matches either shape, so the classifier never has to name the six
#      project-configured guards to leave them alone.
#   4. No guard actually shipped (present as a symlink under some skill's
#      scripts/) derives a repository root as `$SCRIPT_DIR/..` — a fixed
#      number of levels above itself. "Shipped" is read from rule 1's own scan
#      rather than a hardcoded guard list, so a newly shipped guard is covered
#      the moment it is symlinked in.
#
# Prints one violation line per finding (`path:line: message`), then ONE
# verdict line:
#   GUARD-SYMLINKS-OK:      <root> — N guard(s) across M skill(s) validated
#   GUARD-SYMLINKS-INVALID: <root> — N violation(s)
#
# Exit 0 clean, 1 with any violation, 2 when it cannot answer at all — a root
# that is not a readable directory, a skills/ or scripts/ directory that
# cannot be scanned, or a file this guard needed to read but could not. A
# refusal writes its reason to stderr and NOTHING to stdout — the caller must
# never read an absent report as a clean one.
#
# THREE DISCIPLINES CARRIED HERE, adopted from scripts/lib/panel-record.sh's
# header rather than invented fresh: `-a` on every grep (a stray NUL must not
# put grep into binary "no match" mode and read a corrupted guard as a clean
# one); the `rc > 1` split between "no match" (1, an answer) and a real
# failure to look (2+, which this guard refuses on rather than reads as
# empty); and `--` before every path handed to grep, so a guard or skill file
# named with a leading `-` is never read as an option. A path handed to awk
# is fed on stdin rather than as an argument instead of the same `--` — macOS
# awk (the BWK "one true awk" this machine ships) does not recognise `--` as
# an end-of-options marker the way grep does, and stdin sidesteps the
# leading-`-` hazard entirely rather than relying on a marker awk ignores.
#
# WHY RULE 2 IS SCOPED TO A SKILL'S OWN DIRECTORY, NOT EVERY CONTRACT ITS TEXT
# CITES. skills/myflow-contracts/*.md is shared prose, cited by path from
# every command skill for reasons that have nothing to do with which guards
# that skill runs — myflow-status cites finish-contract.md to explain how it
# COMBINES a merge-status answer, never to invoke anything in it. Expanding
# rule 2's scan into every contract a skill's text happens to name was tried
# and produced false positives on this repository's own real, correct tree
# (myflow-start and myflow-status were reported as needing guards they do not
# carry and do not run). Scoping to the skill's own files trades a small
# amount of theoretical coverage — a guard invoked ONLY inside a shared
# contract's usage-synopsis prose, naming no skill by scope — for zero false
# failures on a tree already known correct. A guard already shipped that this
# narrower scan does not detect as required is not a violation either way:
# rule 2 only ever flags a citation with NO corresponding symlink, never a
# symlink with no detected citation.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# "One level above $SCRIPT_DIR" is safe here, unlike the two guards KAN-73's
# task 1 fixed: this script is a project-configured guard (resolved through
# .myflow/project.md's own `## lint` list, exactly like check-references.sh)
# and is never itself symlinked into any skill's scripts/ directory — it has
# exactly one home, so it has exactly one answer for where it lives.
if [ -n "${CHECK_GUARD_SYMLINKS_ROOT:-}" ]; then
  ROOT="$CHECK_GUARD_SYMLINKS_ROOT"
elif [ "${CHECK_GUARD_SYMLINKS_ROOT+set}" = "set" ]; then
  echo "check-guard-symlinks: CHECK_GUARD_SYMLINKS_ROOT is set but empty" >&2
  exit 2
else
  ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
fi

if [ ! -d "$ROOT" ] || [ ! -r "$ROOT" ] || [ ! -x "$ROOT" ]; then
  echo "check-guard-symlinks: $ROOT is not a readable directory — cannot scan" >&2
  exit 2
fi

SKILLS_DIR="$ROOT/skills"
SCRIPTS_DIR="$ROOT/scripts"

if [ ! -d "$SKILLS_DIR" ] || [ ! -r "$SKILLS_DIR" ] || [ ! -x "$SKILLS_DIR" ]; then
  echo "check-guard-symlinks: $SKILLS_DIR is not a readable directory — cannot scan" >&2
  exit 2
fi
if [ ! -d "$SCRIPTS_DIR" ] || [ ! -r "$SCRIPTS_DIR" ] || [ ! -x "$SCRIPTS_DIR" ]; then
  echo "check-guard-symlinks: $SCRIPTS_DIR is not a readable directory — cannot scan" >&2
  exit 2
fi

WORK="$(mktemp -d "${TMPDIR:-/tmp}/check-guard-symlinks.XXXXXX")"
cleanup() { rm -rf "$WORK"; }
trap cleanup EXIT

VIOLATIONS_FILE="$WORK/violations"
: > "$VIOLATIONS_FILE"

# violation <path> <line> <message> — every finding goes through here, so the
# report format has exactly one place to drift from.
violation() {
  printf '%s:%s: %s\n' "$1" "$2" "$3" >> "$VIOLATIONS_FILE"
}

# resolve_file <path> -> prints the path's resolved PHYSICAL location on
# stdout. Sourced from lib/resolve-file.sh rather than defined here — see
# that file's header for why this guard, unlike preserve-session-records.sh
# and gather-self-review-context.sh, may safely source a sibling instead of
# carrying its own copy.
source "$SCRIPT_DIR/lib/resolve-file.sh"

# ---------------------------------------------------------------------------
# Command skills: every directory directly under skills/ except
# myflow-contracts, which is shared prose loaded by several commands and
# SHALL NOT carry a scripts/ directory of its own (task 3's resolution rule).
# Read from the real tree rather than hardcoded, so a new command skill is
# covered the moment it is added.
# ---------------------------------------------------------------------------
COMMAND_SKILLS_FILE="$WORK/command_skills"
if ! find "$SKILLS_DIR" -mindepth 1 -maxdepth 1 -type d ! -name 'myflow-contracts' -print 2>"$WORK/find_err" | sort > "$COMMAND_SKILLS_FILE"; then
  echo "check-guard-symlinks: could not enumerate $SKILLS_DIR" >&2
  cat "$WORK/find_err" >&2
  exit 2
fi
if [ -s "$WORK/find_err" ]; then
  echo "check-guard-symlinks: error while scanning $SKILLS_DIR" >&2
  cat "$WORK/find_err" >&2
  exit 2
fi

SKILL_NAMES_FILE="$WORK/skill_names"
: > "$SKILL_NAMES_FILE"
while IFS= read -r skill_dir; do
  [ -n "$skill_dir" ] || continue
  basename -- "$skill_dir" >> "$SKILL_NAMES_FILE"
done < "$COMMAND_SKILLS_FILE"

# ---------------------------------------------------------------------------
# GUARD_SET_FILE — every basename that exists directly under the repository's
# real scripts/ directory (a file OR a directory, so "lib" counts). This is
# the vocabulary rule 2's citation scan matches against: a guard a skill's
# text cites but which is NOT YET shipped anywhere is exactly the regression
# rule 2 exists to catch, so the set has to include every REAL guard, not
# only the ones already found symlinked in somewhere.
# ---------------------------------------------------------------------------
GUARD_SET_FILE="$WORK/guard_set"
if ! find "$SCRIPTS_DIR" -mindepth 1 -maxdepth 1 -print 2>"$WORK/find_err" | while IFS= read -r e; do basename -- "$e"; done | sort -u > "$GUARD_SET_FILE"; then
  echo "check-guard-symlinks: could not enumerate $SCRIPTS_DIR" >&2
  exit 2
fi
if [ -s "$WORK/find_err" ]; then
  echo "check-guard-symlinks: error while scanning $SCRIPTS_DIR" >&2
  cat "$WORK/find_err" >&2
  exit 2
fi

# ===========================================================================
# RULE 1 — every entry under skills/*/scripts/ is a symlink, it resolves, and
# its target is relative. Also builds REAL_TARGETS_FILE, the unique resolved
# sources rule 4 reads.
# ===========================================================================
REAL_TARGETS_FILE="$WORK/real_targets"
: > "$REAL_TARGETS_FILE"

while IFS= read -r skill_dir; do
  [ -n "$skill_dir" ] || continue
  scripts_dir="$skill_dir/scripts"
  [ -e "$scripts_dir" ] || continue
  if [ ! -d "$scripts_dir" ] || [ ! -r "$scripts_dir" ] || [ ! -x "$scripts_dir" ]; then
    echo "check-guard-symlinks: $scripts_dir is not a readable directory — cannot scan" >&2
    exit 2
  fi

  entries_file="$WORK/entries"
  if ! find "$scripts_dir" -mindepth 1 -maxdepth 1 -print 2>"$WORK/find_err" | sort > "$entries_file"; then
    echo "check-guard-symlinks: could not list $scripts_dir" >&2
    cat "$WORK/find_err" >&2
    exit 2
  fi
  if [ -s "$WORK/find_err" ]; then
    echo "check-guard-symlinks: error while listing $scripts_dir" >&2
    cat "$WORK/find_err" >&2
    exit 2
  fi

  while IFS= read -r entry; do
    [ -n "$entry" ] || continue
    base="$(basename -- "$entry")"
    printf '%s\n' "$base" >> "$GUARD_SET_FILE"

    if [ ! -L "$entry" ]; then
      violation "$entry" 0 "not a symlink — every entry under skills/*/scripts/ must be a relative symlink into the repository's scripts/ directory (rule 1)"
      continue
    fi

    target="$(readlink -- "$entry")"
    case "$target" in
      /*)
        # continue, like its two sibling branches above and below — an
        # absolute (off-repo) target is ALREADY a rule 1 violation on its
        # own, and falling through would additionally feed it into
        # REAL_TARGETS_FILE, after which rule 4 greps a file outside the
        # repository this guard was asked to scan. Reported once, here.
        violation "$entry" 0 "symlink target is absolute ('$target') — every entry under skills/*/scripts/ must be a RELATIVE symlink, or it bakes this machine's checkout path into the repository (rule 1)"
        continue
        ;;
    esac

    if [ ! -e "$entry" ]; then
      violation "$entry" 0 "symlink does not resolve (target '$target') (rule 1)"
      continue
    fi

    real="$(resolve_file "$entry")" || {
      violation "$entry" 0 "cannot resolve this symlink's real physical location (rule 1)"
      continue
    }
    if [ -f "$real" ]; then
      printf '%s\n' "$real" >> "$REAL_TARGETS_FILE"
    fi
  done < "$entries_file"
done < "$COMMAND_SKILLS_FILE"

sort -u -o "$GUARD_SET_FILE" "$GUARD_SET_FILE"
sort -u -o "$REAL_TARGETS_FILE" "$REAL_TARGETS_FILE"

# ===========================================================================
# RULE 4 — no guard actually shipped derives a repository root as a fixed
# number of levels above $SCRIPT_DIR. Scoped to REAL_TARGETS_FILE, which rule
# 1 populated from what is ACTUALLY symlinked in — never a hardcoded guard
# list — so a project-configured guard that legitimately keeps the old
# `$SCRIPT_DIR/..` form (it is never reachable from anywhere but its own
# scripts/ directory) is correctly left alone.
#
# THREE SPELLINGS OF THE SAME DEFECT, not one. `$SCRIPT_DIR/..` was the only
# shape matched here originally, and it is the only one that was ever fixed
# by name — but `REPO_ROOT="$(dirname "$SCRIPT_DIR")"` and
# `cd "$SCRIPT_DIR" && cd ..` (or `; cd ..`) answer the identical question the
# identical wrong way, and neither is textually `$SCRIPT_DIR/..`. A guard
# passing rule 4 by spelling the same fixed-depth walk differently is rule 4
# defeated by the bug it exists to stop, so all three are matched here.
# ===========================================================================
FIXED_DEPTH_ROOT_RE='\$\{?SCRIPT_DIR\}?"?/\.\.'"|"'dirname([ \t]+--)?[ \t]+"?\$\{?SCRIPT_DIR\}?"?'"|"'cd[ \t]+"?\$\{?SCRIPT_DIR\}?"?"?[ \t]*(&&|;)[ \t]*cd[ \t]+\.\.'

if [ -s "$REAL_TARGETS_FILE" ]; then
  while IFS= read -r real; do
    [ -n "$real" ] || continue
    if [ ! -r "$real" ]; then
      violation "$real" 0 "cannot read this guard's real source to check for a fixed-depth root derivation (rule 4)"
      continue
    fi
    set +e
    matches="$(grep -anE -- "$FIXED_DEPTH_ROOT_RE" "$real")"
    rc=$?
    set -e
    if [ "$rc" -ge 2 ]; then
      violation "$real" 0 "grep exited $rc while scanning for a fixed-depth root derivation — a failure to look, not an absence (rule 4)"
      continue
    fi
    if [ "$rc" -eq 0 ]; then
      while IFS= read -r m; do
        [ -n "$m" ] || continue
        lnno="${m%%:*}"
        violation "$real" "$lnno" "derives a repository root as \$SCRIPT_DIR/.. — this guard is reachable from more than one directory once shipped, and a fixed number of levels above itself silently answers a skill directory instead of the repository root (rule 4)"
      done <<VIOEOF
$matches
VIOEOF
    fi
  done < "$REAL_TARGETS_FILE"
fi

# ===========================================================================
# RULE 3 — no skill text carries a repository-relative scripts/<name> path in
# an invoking position. Scanned across every .md file under skills/, fence-
# aware: a non-comment line inside a bash/sh/zsh fence, or a "Run"/"Invoke"/
# "Execute" immediately before the backtick, is an invocation; everything
# else — including a plain fenced example with no such language tag, and a
# sentence describing what a guard reads or does — is prose and is left
# alone. The offending line's own text is never echoed into the report: this
# repository's skill text is tracked and editable in any pull request, so the
# report names path and line only, never the attacker-influenced content.
# ===========================================================================
ALL_MD_FILE="$WORK/all_md"
if ! find "$SKILLS_DIR" -type f -name '*.md' -print 2>"$WORK/find_err" | sort > "$ALL_MD_FILE"; then
  echo "check-guard-symlinks: could not enumerate .md files under $SKILLS_DIR" >&2
  cat "$WORK/find_err" >&2
  exit 2
fi
if [ -s "$WORK/find_err" ]; then
  echo "check-guard-symlinks: error while scanning $SKILLS_DIR for .md files" >&2
  cat "$WORK/find_err" >&2
  exit 2
fi

RULE3_AWK='
BEGIN {
  fence_len = 0; fence_lang = ""
  # The six project-configured guards named in design.md, "Two families of
  # guard, and only one of them ships": these are resolved through a
  # project'"'"'s own .myflow/project.md, never invoked by a command directly,
  # so prose naming them keeps its repository-relative form legitimately —
  # rule 3 governs INVOKED guards only. Keyed by name so membership is a
  # single lookup rather than an alternation repeated at every call site.
  EXEMPT["check-references.sh"] = 1
  EXEMPT["check-vocabulary.sh"] = 1
  EXEMPT["check-plan-provenance.sh"] = 1
  EXEMPT["check-task-build-green.sh"] = 1
  EXEMPT["check-contract-budget.sh"] = 1
  EXEMPT["check-stage-mark-calls.sh"] = 1
}
{
  raw = $0
  line = raw
  sub(/^[ \t]+/, "", line)
  if (fence_len == 0) {
    if (line ~ /^```/) {
      n = 0
      while (substr(line, n + 1, 1) == "`") n++
      fence_len = n
      fence_lang = substr(line, n + 1)
      gsub(/^[ \t]+/, "", fence_lang)
      gsub(/[ \t]+$/, "", fence_lang)
      next
    }
    scan_prose(raw, FNR)
    next
  }
  n = 0
  while (substr(line, n + 1, 1) == "`") n++
  trail = substr(line, n + 1)
  gsub(/[ \t]+$/, "", trail)
  if (n >= fence_len && trail == "") { fence_len = 0; fence_lang = ""; next }
  # Anchored to a full word — "bash", "sh" or "zsh" followed by whitespace or
  # end of string — never a prefix match. Unanchored, this also matched
  # "shell" and "shellsession" by accident (both start with "sh"), which were
  # never a declared member of the set this guard scans; "console" and
  # "text" are deliberately still excluded, since every real invocation in
  # this repository is written in a bash/sh/zsh fence and nowhere else.
  if (fence_lang ~ /^(bash|sh|zsh)([ \t]|$)/) scan_fence(raw, FNR)
  next
}
function is_exempt(matched_span,   name) {
  name = matched_span
  sub(/^scripts\//, "", name)
  sub(/\/.*/, "", name)
  return (name in EXEMPT)
}
function scan_fence(line, lineno,   trimmed) {
  trimmed = line
  sub(/^[ \t]+/, "", trimmed)
  if (trimmed ~ /^#/) return
  if (match(trimmed, /scripts\/[A-Za-z0-9._\/-]+/)) {
    if (is_exempt(substr(trimmed, RSTART, RLENGTH))) return
    printf "%d\n", lineno
  }
}
function scan_prose(line, lineno,   m) {
  if (match(line, /(Run|run|Invoke|invoke|Execute|execute)[ \t]+`scripts\/[A-Za-z0-9._\/-]+/)) {
    m = substr(line, RSTART, RLENGTH)
    sub(/.*`/, "", m)
    if (is_exempt(m)) return
    printf "%d\n", lineno
  }
}
'

while IFS= read -r mdfile; do
  [ -n "$mdfile" ] || continue
  if [ ! -r "$mdfile" ]; then
    echo "check-guard-symlinks: cannot read $mdfile (rule 3)" >&2
    exit 2
  fi
  if ! awk "$RULE3_AWK" < "$mdfile" > "$WORK/rule3_out" 2>"$WORK/rule3_err"; then
    echo "check-guard-symlinks: awk failed scanning $mdfile for rule 3" >&2
    cat "$WORK/rule3_err" >&2
    exit 2
  fi
  if [ -s "$WORK/rule3_err" ]; then
    echo "check-guard-symlinks: awk reported an error scanning $mdfile for rule 3" >&2
    cat "$WORK/rule3_err" >&2
    exit 2
  fi
  while IFS= read -r lnno; do
    [ -n "$lnno" ] || continue
    violation "$mdfile" "$lnno" "a repository-relative scripts/<name> path appears in an invoking position (a bash-fenced command, or an imperative Run/Invoke/Execute) — name the guard by basename instead, per the resolution rule in skills/myflow-contracts/pipeline.md (rule 3)"
  done < "$WORK/rule3_out"
done < "$ALL_MD_FILE"

# ===========================================================================
# RULE 2 — every guard invoked in a skill's own text has a symlink in that
# skill's own scripts/ directory, sibling dependencies included. Level-0
# citations come from the skill's own *.md files; a required guard's
# siblings are then derived from THAT GUARD'S OWN SOURCE (a $SCRIPT_DIR/<name>
# grep), never from a hardcoded table — task 2's map, checked rather than
# trusted, per the task.
# ===========================================================================
CITATION_AWK='
BEGIN {
  while ((getline g < guardfile) > 0) if (g != "") guards[g] = 1
  close(guardfile)
  fence_len = 0
  fence_lang = ""
}
{
  raw = $0
  line = raw
  sub(/^[ \t]+/, "", line)
  if (fence_len == 0) {
    if (line ~ /^```/) {
      n = 0
      while (substr(line, n + 1, 1) == "`") n++
      fence_len = n
      fence_lang = substr(line, n + 1)
      gsub(/^[ \t]+/, "", fence_lang)
      gsub(/[ \t]+$/, "", fence_lang)
      next
    }
    scan_prose(raw, FNR)
    next
  }
  n = 0
  while (substr(line, n + 1, 1) == "`") n++
  trail = substr(line, n + 1)
  gsub(/[ \t]+$/, "", trail)
  if (n >= fence_len && trail == "") { fence_len = 0; fence_lang = ""; next }
  # Anchored to a full word, exactly as rule 3 above — see that pass'"'"'s
  # comment for why "shell" and "shellsession" must not match "sh" as a
  # prefix, and why "console"/"text" stay out of the set on purpose.
  if (fence_lang ~ /^(bash|sh|zsh)([ \t]|$)/) scan_fence(raw, FNR)
  next
}
function scan_fence(line, lineno,   trimmed, rest, w) {
  trimmed = line
  sub(/^[ \t]+/, "", trimmed)
  if (trimmed ~ /^#/) return
  # The leading token alone missed two real invocation shapes: `./guard.sh
  # <args>` (the leading token is `./guard.sh`, which this class never
  # matched past the `.`) and `bash guard.sh <args>` / `sh guard.sh` / `zsh
  # guard.sh` (an interpreter named explicitly, with the guard as the SECOND
  # token). Both are tried in turn; `./` is stripped from either position
  # before the basename test, since `bash ./guard.sh` combines them.
  rest = trimmed
  sub(/^\.\//, "", rest)
  if (match(rest, /^[A-Za-z0-9._-]+/)) {
    w = substr(rest, RSTART, RLENGTH)
    if (w in guards) { printf "%s\t%d\n", w, lineno; return }
  }
  if (match(trimmed, /^(bash|sh|zsh)[ \t]+/)) {
    rest = substr(trimmed, RLENGTH + 1)
    sub(/^\.\//, "", rest)
    if (match(rest, /^[A-Za-z0-9._-]+/)) {
      w = substr(rest, RSTART, RLENGTH)
      if (w in guards) printf "%s\t%d\n", w, lineno
    }
  }
}
function scan_prose(line, lineno,    s, pos, m, start, m2, span, base, restspan, prefix, matched, arr, n2, i, from, tailwords) {
  s = line
  pos = 1
  while (1) {
    m = index(substr(s, pos), "`")
    if (m == 0) break
    start = pos + m
    m2 = index(substr(s, start), "`")
    if (m2 == 0) break
    span = substr(s, start, m2 - 1)
    if (match(span, /^[A-Za-z0-9._-]+/)) {
      base = substr(span, RSTART, RLENGTH)
      restspan = substr(span, RLENGTH + 1)
      if (base in guards) {
        matched = 0
        if (restspan ~ /^[ \t]+</) matched = 1
        if (!matched) {
          prefix = substr(line, 1, start - 2)
          n2 = split(prefix, arr, /[ \t]+/)
          from = n2 - 3
          if (from < 1) from = 1
          tailwords = ""
          for (i = from; i <= n2; i++) tailwords = tailwords " " tolower(arr[i])
          if (tailwords ~ /(run|invoke|invocation|invoking|execute)/) matched = 1
        }
        if (matched) printf "%s\t%d\n", base, lineno
      }
    }
    # Advance to just past the OPENING backtick of THIS attempt, not past
    # both. A stray, unpaired backtick earlier in the line (a markdown typo,
    # or a genuine apostrophe-adjacent character) makes `start` land on that
    # stray mark and `m2` land on the NEXT real backtick — the one that was
    # meant to OPEN the following citation. Skipping to `start + m2` consumes
    # that backtick as a CLOSER here and leaves nothing to pair the real
    # citation'"'"'s own closing backtick with, dropping it silently.
    # Resyncing to `start` retries from the very next backtick as a fresh
    # opening candidate instead, so one stray mark costs at most one empty
    # non-match rather than a real citation.
    pos = start
  }
}
'

REQUIRED_FILE="$WORK/required"
: > "$REQUIRED_FILE"

# DELEGATE_AWK — for a skill that invokes no guard of its own and instead
# CHAINS another command's stages verbatim (myflow-fast is, today, the only
# one), its required set has to come from somewhere other than its own text,
# or rule 2 checks nothing for it at all — task 5's own file-level comment on
# "WHY RULE 2 IS SCOPED TO A SKILL'S OWN DIRECTORY" already rejected scanning
# every contract a skill's prose happens to CITE, because myflow-start and
# myflow-status cite contracts and other skills' sections constantly without
# invoking anything in them (confirmed above: myflow-status alone cites
# `/myflow-do`, `/myflow-finish` and `/myflow-start` by slash-command over a
# dozen times, purely in explanatory prose). A bare citation of another
# command is therefore not a safe delegation signal.
#
# What IS safe: every command skill that invokes at least one guard of its
# own states so in a paragraph beginning "**Check guard presence.**" (task 4's
# own convention — myflow-do and myflow-finish enumerate guard basenames
# there; myflow-status instead says, in the same slot, "No guard-presence
# check here — this command invokes no guard."). myflow-fast's OWN copy of
# that exact paragraph is the one place in its text that names OTHER
# commands' presence checks by slash-command — "confirm every guard named in
# `/myflow-do`'s, `/myflow-finish`'s and `/myflow-status`'s own presence
# checks ... is present in `skills/myflow-fast/scripts/`" — so scoping the
# slash-command scan to THAT paragraph specifically, rather than the whole
# file, is what tells a real delegation apart from the ordinary cross-command
# prose every skill carries. A command with no such paragraph names no
# delegate, and contributes nothing.
DELEGATE_AWK='
BEGIN { in_para = 0 }
/^\*\*Check guard presence\.\*\*/ { in_para = 1 }
in_para {
  if ($0 ~ /^[ \t]*$/) { in_para = 0; next }
  s = $0
  while (match(s, /`\/myflow-[A-Za-z0-9_-]+`/)) {
    tok = substr(s, RSTART + 2, RLENGTH - 3)
    print tok
    s = substr(s, RSTART + RLENGTH)
  }
  next
}
'

DELEGATES_FILE="$WORK/delegates"
: > "$DELEGATES_FILE"

while IFS= read -r skill_dir; do
  [ -n "$skill_dir" ] || continue
  skill="$(basename -- "$skill_dir")"

  own_md_file="$WORK/own_md"
  if ! find "$skill_dir" -maxdepth 1 -type f -name '*.md' -print 2>"$WORK/find_err" | sort > "$own_md_file"; then
    echo "check-guard-symlinks: could not enumerate .md files under $skill_dir" >&2
    cat "$WORK/find_err" >&2
    exit 2
  fi
  if [ -s "$WORK/find_err" ]; then
    echo "check-guard-symlinks: error while scanning $skill_dir for .md files" >&2
    cat "$WORK/find_err" >&2
    exit 2
  fi

  while IFS= read -r mdfile; do
    [ -n "$mdfile" ] || continue
    if [ ! -r "$mdfile" ]; then
      echo "check-guard-symlinks: cannot read $mdfile (rule 2)" >&2
      exit 2
    fi
    if ! awk -v guardfile="$GUARD_SET_FILE" "$CITATION_AWK" < "$mdfile" > "$WORK/cite_out" 2>"$WORK/cite_err"; then
      echo "check-guard-symlinks: awk failed scanning $mdfile for rule 2" >&2
      cat "$WORK/cite_err" >&2
      exit 2
    fi
    if [ -s "$WORK/cite_err" ]; then
      echo "check-guard-symlinks: awk reported an error scanning $mdfile for rule 2" >&2
      cat "$WORK/cite_err" >&2
      exit 2
    fi
    while IFS="$(printf '\t')" read -r guard lineno; do
      [ -n "$guard" ] || continue
      printf '%s\t%s\t%s\t%s\n' "$skill" "$guard" "$mdfile" "$lineno" >> "$REQUIRED_FILE"
    done < "$WORK/cite_out"

    if ! awk "$DELEGATE_AWK" < "$mdfile" > "$WORK/delegate_out" 2>"$WORK/delegate_err"; then
      echo "check-guard-symlinks: awk failed scanning $mdfile for delegation" >&2
      cat "$WORK/delegate_err" >&2
      exit 2
    fi
    if [ -s "$WORK/delegate_err" ]; then
      echo "check-guard-symlinks: awk reported an error scanning $mdfile for delegation" >&2
      cat "$WORK/delegate_err" >&2
      exit 2
    fi
    while IFS= read -r delegate; do
      [ -n "$delegate" ] || continue
      [ "$delegate" != "$skill" ] || continue
      printf '%s\t%s\n' "$skill" "$delegate" >> "$DELEGATES_FILE"
    done < "$WORK/delegate_out"
  done < "$own_md_file"
done < "$COMMAND_SKILLS_FILE"

sort -u -o "$DELEGATES_FILE" "$DELEGATES_FILE"

# Resolve one level of delegation. REQUIRED_FILE at this point holds ONLY
# level-0 (directly cited) entries for every skill, which is exactly what a
# delegate'"'"'s "own" required set means — a transitive delegate-of-a-delegate
# does not exist among today'"'"'s command skills, and adding a second round here
# would be speculative generality for a case nothing exercises.
DELEGATED_FILE="$WORK/delegated_required"
: > "$DELEGATED_FILE"
while IFS="$(printf '\t')" read -r skill delegate; do
  [ -n "$skill" ] || continue
  [ -n "$delegate" ] || continue
  [ -d "$SKILLS_DIR/$delegate" ] || continue
  awk -F "$(printf '\t')" -v sk="$delegate" -v newsk="$skill" -v via="$delegate" \
    '$1 == sk { printf "%s\t%s\t%s\t%s\t%s\n", newsk, $2, $3, $4, via }' \
    "$REQUIRED_FILE" >> "$DELEGATED_FILE"
done < "$DELEGATES_FILE"
if [ -s "$DELEGATED_FILE" ]; then
  cat "$DELEGATED_FILE" >> "$REQUIRED_FILE"
fi

# Sibling closure — read from each required guard's OWN real source, one
# round per frontier level, capped so a cycle in $SCRIPT_DIR references
# cannot spin this guard forever.
while IFS= read -r skill; do
  [ -n "$skill" ] || continue
  seen_file="$WORK/seen.$skill"
  frontier_file="$WORK/frontier.$skill"
  awk -F "$(printf '\t')" -v sk="$skill" '$1 == sk { print $2 }' "$REQUIRED_FILE" | sort -u > "$frontier_file"
  sort -u -o "$seen_file" "$frontier_file" 2>/dev/null || cp "$frontier_file" "$seen_file"

  round=0
  while [ -s "$frontier_file" ]; do
    round=$((round + 1))
    [ "$round" -le 40 ] || {
      echo "check-guard-symlinks: sibling-dependency closure for $skill did not terminate within 40 rounds — a cycle in \$SCRIPT_DIR references" >&2
      exit 2
    }
    next_frontier="$WORK/next_frontier.$skill"
    : > "$next_frontier"
    while IFS= read -r g; do
      [ -n "$g" ] || continue
      real="$SCRIPTS_DIR/$g"
      [ -f "$real" ] || continue
      if [ ! -r "$real" ]; then
        echo "check-guard-symlinks: cannot read $real to derive sibling dependencies (rule 2)" >&2
        exit 2
      fi
      set +e
      sibs="$(grep -aoE -- '\$\{?SCRIPT_DIR\}?"?/[A-Za-z0-9._-]+' "$real")"
      rc=$?
      set -e
      if [ "$rc" -ge 2 ]; then
        echo "check-guard-symlinks: grep exited $rc scanning $real for sibling dependencies (rule 2)" >&2
        exit 2
      fi
      [ "$rc" -eq 0 ] || continue
      while IFS= read -r s; do
        [ -n "$s" ] || continue
        sib="${s##*/}"
        [ -n "$sib" ] || continue
        # `-a` and the rc>=2 refusal, per this file's own header discipline —
        # both were missing here even though the two other reads of a guard's
        # source in this same file (rule 4's scan, and the sibling-source
        # read a few lines above) already carry them. Without `-a`, a
        # corrupted seen_file reads as "nothing seen yet" rather than a
        # failure to look; without the rc split, `!` folds "could not check"
        # into "not seen", the same fail-open direction closed everywhere
        # else in this guard.
        set +e
        grep -aqxF -- "$sib" "$seen_file"
        seen_rc=$?
        set -e
        if [ "$seen_rc" -ge 2 ]; then
          echo "check-guard-symlinks: grep exited $seen_rc while checking $seen_file for a previously seen sibling dependency (rule 2)" >&2
          exit 2
        fi
        if [ "$seen_rc" -ne 0 ]; then
          printf '%s\n' "$sib" >> "$seen_file"
          printf '%s\n' "$sib" >> "$next_frontier"
          printf '%s\t%s\t%s\t%s\n' "$skill" "$sib" "$real" "sibling" >> "$REQUIRED_FILE"
        fi
      done <<VIOEOF
$sibs
VIOEOF
    done < "$frontier_file"
    mv "$next_frontier" "$frontier_file"
  done
done < "$SKILL_NAMES_FILE"

# One violation per (skill, guard) pair, at its first-discovered citation.
REQUIRED_UNIQUE="$WORK/required_unique"
awk -F "$(printf '\t')" '{ key = $1 "\t" $2; if (!(key in seen)) { seen[key] = 1; print } }' "$REQUIRED_FILE" > "$REQUIRED_UNIQUE"

while IFS="$(printf '\t')" read -r skill guard citing_file citing_line via; do
  [ -n "$skill" ] || continue
  have_path="$SKILLS_DIR/$skill/scripts/$guard"
  if [ ! -e "$have_path" ] && [ ! -L "$have_path" ]; then
    if [ "$citing_line" = "sibling" ]; then
      violation "$citing_file" 0 "$guard is a sibling dependency this guard resolves from its own directory, but skill \"$skill\" carries no symlink at skills/$skill/scripts/$guard (rule 2)"
    elif [ -n "$via" ]; then
      violation "$citing_file" "$citing_line" "invokes $guard here, and skill \"$skill\" delegates to \"$via\"'s own guard presence check (task 4's convention), but carries no symlink at skills/$skill/scripts/$guard (rule 2)"
    else
      violation "$citing_file" "$citing_line" "invokes $guard here, but skill \"$skill\" carries no symlink at skills/$skill/scripts/$guard (rule 2)"
    fi
  fi
done < "$REQUIRED_UNIQUE"

# ===========================================================================
# Verdict.
# ===========================================================================
GUARD_COUNT="$(wc -l < "$GUARD_SET_FILE" | tr -d ' ')"
SKILL_COUNT="$(wc -l < "$SKILL_NAMES_FILE" | tr -d ' ')"

if [ -s "$VIOLATIONS_FILE" ]; then
  N_FOUND="$(wc -l < "$VIOLATIONS_FILE" | tr -d ' ')"
  cat "$VIOLATIONS_FILE"
  printf 'GUARD-SYMLINKS-INVALID: %s — %s violation(s)\n' "$ROOT" "$N_FOUND"
  exit 1
fi

printf 'GUARD-SYMLINKS-OK: %s — %s guard(s) across %s skill(s) validated\n' "$ROOT" "$GUARD_COUNT" "$SKILL_COUNT"
exit 0
