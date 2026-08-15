#!/usr/bin/env bash
# check-stage-mark-calls.sh — every `myflow stage begin` in a skill must carry
# a session token and a harness, the session token must be a literal, and the harness must
# NOT be — the two required flags are wrong in opposite directions, and this
# script rejects both.
#
# Usage: scripts/check-stage-mark-calls.sh [path ...]          # dirs or files
#
# With no arguments it scans `skills/`, resolved relative to the repo root —
# the one tree the marking skills live under. A caller may pass explicit
# paths instead (the companion test harness does, against mktemp fixtures).
#
# WHAT THIS CATCHES. KAN-172 task 1 gave `myflow stage begin` two required
# flags: `-session-token`, a literal token the daemon later finds in the calling
# session's own transcript to bind that stage run's session_id, and
# `-harness`, the harness recording the mark. Neither is enforced by prose:
# four skills carried a `myflow stage begin` invocation with neither flag for
# the whole of KAN-16's life before this guard existed (tasks.md, "This guard
# is the deliverable of the task, not the edits"). This script makes that
# defect loud rather than silent, forever.
#
# It also enforces design.md's second decision, "the session token is a literal,
# never a shell substitution": `-session-token "mf-$(date +%s)-$$"` is recorded by
# the transcript *before* the shell expands it, so every session would record
# the identical string and it would identify nothing. A `-session-token` value
# containing `$(`, a backtick, or `$` immediately followed by a shell
# variable-name character is rejected here, in the skill source, before a
# reader ever has the chance to run it — the same three shapes
# `stats/cmd/myflow/stage.go`'s `validateSessionToken` rejects at the CLI, and
# `internal/api/stages.go`'s `validateSessionTokenShape` rejects again at the store,
# as defence in depth.
#
# `-harness` FAILS THE OPPOSITE WAY: a value that IS a literal is the defect.
# These skill files are one source installed into `~/.claude/skills/`,
# `~/.cursor/skills/` and `~/.codex/skills/` alike (CLAUDE.md's "installed
# alongside the skills in every harness"), so a hardcoded `-harness
# claude-code` in the shared source mislabels every Cursor and Codex run as
# Claude Code — masking the very thing the harness field exists to record:
# that Cursor and Codex write no transcript, so their runs are *explicitly
# unavailable* rather than zero. The skill source must instead carry a
# placeholder the agent fills in at call time, exactly as `-session-token`'s literal
# token is filled in at call time — this guard requires the `-harness` value
# to be written as a bracketed placeholder (`<harness>`) and rejects any bare
# value, the same way it rejects a substituted session token.
#
# `myflow stage end` is deliberately NOT checked: pipeline.md's "Stage marks"
# section states plainly that `stage end` carries no harness of its own (the
# harness is recorded once at `begin` and is immutable), and it takes no
# `-session-token` either — attribution happens once, at `begin`. Only lines that
# invoke `stage begin` are examined.
#
# `/myflow-status` marks nothing, per its own row in README.md's Level 1
# table, so a `myflow stage begin` appearing in its SKILL.md would itself be
# a defect this guard would (correctly) catch — there is no exemption for it.
#
# It also enforces that a `stage begin`'s change argument is never a guess.
# The guard cannot know whether `<name>` is resolved at a given call site,
# but it can know that a bracketed placeholder whose own text says "guess"
# is not — `<name-or-best-guess>` and any restatement of it. KAN-182: on
# 2026-08-15 `/myflow-fast`'s state gate marked `do.state-gate` with
# `<name-or-best-guess>` before its change name was resolved; the daemon's
# begin handler bootstraps a change row for any name it has never seen
# (`ApplyBeginStageMark`, `stats/internal/api/stages.go`), so the guess got
# a row of its own — `kan-175` and `kan-175-more-ui-ux-fixes` appeared as
# two open changes, 29 seconds apart. A mark writes, so a guessed name
# bootstraps a change row that outlives the run.
#
# The check does not try to identify "the change argument" at all. Finding
# the last token on a shell command line requires parsing shell syntax, and
# this guard kept getting that parsing wrong: KAN-182 round 1 added
# quote-awareness to a naive last-token extraction so a quoted or
# commented-out placeholder would still be caught, and that parsing
# introduced three more bypasses (round 2) — a quoted value with an internal
# space that a space-only split truncated before the placeholder, an escaped
# quote inside a session token that desynced the quote-tracker so a later
# `#` was read as an unquoted comment and discarded the rest of the line,
# and a tab before the change argument that a space-only split never saw as
# a token boundary. Extraction is deleted rather than patched a fourth time.
# Instead the check searches the WHOLE assembled command text for a
# bracketed placeholder whose own text says "guess" (`<...guess...>`,
# case-insensitive) — strictly stronger than extracting a token, and immune
# to quoting, spacing and comments because it never tries to reason about
# any of them. One behaviour is worth stating plainly: a guessed placeholder
# written inside a trailing comment is reported too — a `stage begin` line
# that mentions `<name-or-best-guess>` anywhere is a call site to fix.
#
# MULTI-LINE INVOCATIONS. A real call is frequently wrapped across several
# lines with a trailing `\` continuation, exactly as an ordinary shell
# command would be. This guard joins a `stage begin` line with every
# following line while the previous physical line ends in `\`, then checks
# the assembled logical command as one string — a check against the raw
# physical lines alone would miss `-session-token`/`-harness` written on a
# continuation line, which is precisely how these calls are written today.
#
# Exit codes: 0 every `stage begin` call is compliant; 1 at least one
# violation was found; 2 the guard cannot answer at all (a bad path, or an
# awk/grep failure — never read as "no findings").
set -uo pipefail

die() {
  echo "check-stage-mark-calls: $*" >&2
  exit 2
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [[ $# -gt 0 ]]; then
  TARGETS=("$@")
else
  TARGETS=("$REPO_ROOT/skills")
fi

for t in "${TARGETS[@]}"; do
  [[ -e "$t" ]] || die "no such file or directory: $t"
done

# assemble_calls <file> — print one line per `stage begin` invocation found in
# <file>, as "<start-line>\t<joined command text>", continuation lines folded
# in. Reads a real file's physical lines with awk, one pass, no shell loop
# over its content (a shell `while read` would mangle a backslash the same
# way the continuation logic here has to preserve deliberately).
assemble_calls() {
  local file="$1"
  awk '
    BEGIN { instage = 0; buf = ""; startline = 0 }
    function stripcr(s) { sub(/\r$/, "", s); return s }
    {
      line = stripcr($0)
      if (instage) {
        buf = buf " " line
        if (line ~ /\\[[:space:]]*$/) {
          sub(/\\[[:space:]]*$/, "", buf)
          next
        }
        printf "%d\t%s\n", startline, buf
        instage = 0
        buf = ""
        next
      }
      if (line ~ /^[[:space:]]*myflow stage begin([[:space:]]|\\|$)/) {
        if (line ~ /\\[[:space:]]*$/) {
          instage = 1
          startline = NR
          buf = line
          sub(/\\[[:space:]]*$/, "", buf)
          next
        }
        printf "%d\t%s\n", NR, line
      }
    }
    END {
      # A file ending mid-continuation is a malformed fixture/skill, not a
      # silently dropped call -- report it as its own finding downstream by
      # emitting what was gathered rather than swallowing it.
      if (instage) printf "%d\t%s\n", startline, buf
    }
  ' "$file"
}

# session_token_value <command text> — the first `-session-token`'s argument, single- or
# double-quoted or bare, or empty if `-session-token` does not appear at all. Kept as
# its own function because the shape check below and the presence check both
# need it, and a single extraction keeps the two answers consistent.
session_token_value() {
  local text="$1"
  printf '%s\n' "$text" | grep -oE -- "-session-token[[:space:]]+('[^']*'|\"[^\"]*\"|[^[:space:]]+)" | head -n1 \
    | sed -E "s/^-session-token[[:space:]]+//; s/^'(.*)'\$/\1/; s/^\"(.*)\"\$/\1/"
}

# harness_value <command text> — the first `-harness`'s argument, same
# extraction shape as session_token_value above, or empty if `-harness` does not
# appear at all.
harness_value() {
  local text="$1"
  printf '%s\n' "$text" | grep -oE -- "-harness[[:space:]]+('[^']*'|\"[^\"]*\"|[^[:space:]]+)" | head -n1 \
    | sed -E "s/^-harness[[:space:]]+//; s/^'(.*)'\$/\1/; s/^\"(.*)\"\$/\1/"
}

# guess_placeholder <command text> — the first bracketed placeholder whose
# own text says "guess" (`<...guess...>`, case-insensitive), found ANYWHERE
# in the assembled command text, or empty if none appears. No shell parsing
# at all: no token extraction, no quote-tracking, no comment-stripping — the
# regex is applied to the whole line, so quoting, internal whitespace, tabs
# and trailing comments cannot hide a placeholder from it (KAN-182 F6/F7/F8).
guess_placeholder() {
  local text="$1"
  printf '%s\n' "$text" | grep -oE -- '<[^<>]*[Gg][Uu][Ee][Ss][Ss][^<>]*>' | head -n1
}

VIOLATIONS=0
CHECKED=0

for target in "${TARGETS[@]}"; do
  if [[ -d "$target" ]]; then
    while IFS= read -r -d '' f; do
      FILES+=("$f")
    done < <(find "$target" -type f -name '*.md' -print0)
  else
    FILES=("$target")
  fi

  for f in "${FILES[@]:-}"; do
    [[ -n "${f:-}" ]] || continue
    while IFS=$'\t' read -r lineno cmd; do
      [[ -n "$lineno" ]] || continue
      CHECKED=$((CHECKED + 1))

      has_session_token=1
      if ! printf '%s' "$cmd" | grep -qE -- '(^|[[:space:]])-session-token([[:space:]]|=)'; then
        has_session_token=0
        printf '%s:%s: `myflow stage begin` carries no -session-token -- required so the daemon can later bind this stage run to the session that marked it (design.md, "bind after the fact, by a correlator the caller writes")\n' "$f" "$lineno"
        VIOLATIONS=$((VIOLATIONS + 1))
      fi

      has_harness=1
      if ! printf '%s' "$cmd" | grep -qE -- '(^|[[:space:]])-harness([[:space:]]|=)'; then
        has_harness=0
        printf '%s:%s: `myflow stage begin` carries no -harness -- required so a recorded run states which harness marked it\n' "$f" "$lineno"
        VIOLATIONS=$((VIOLATIONS + 1))
      fi

      if [[ "$has_harness" -eq 1 ]]; then
        harness="$(harness_value "$cmd")"
        if [[ ! "$harness" =~ ^\<[^\<\>]+\>$ ]]; then
          printf '%s:%s: -harness %q is a hardcoded literal, not a placeholder -- this skill source is one file installed into `~/.claude/skills/`, `~/.cursor/skills/` and `~/.codex/skills/` alike, so a fixed value mislabels every harness but the one it names; write a bracketed placeholder (e.g. -harness <harness>) that the agent fills in with the harness actually running the command\n' "$f" "$lineno" "$harness"
          VIOLATIONS=$((VIOLATIONS + 1))
        fi
      fi

      if [[ "$has_session_token" -eq 1 ]]; then
        session_token="$(session_token_value "$cmd")"
        case "$session_token" in
          *'$('*)
            printf '%s:%s: -session-token %q contains a command substitution "$(...)" -- the transcript records the command text before the shell expands it, so this value would be recorded identically by every caller and identify nothing; write a literal token instead\n' "$f" "$lineno" "$session_token"
            VIOLATIONS=$((VIOLATIONS + 1))
            ;;
          *'`'*)
            printf '%s:%s: -session-token %q contains a backtick -- backtick command substitution is expanded by the shell after the transcript already recorded the unexpanded text; write a literal token instead\n' "$f" "$lineno" "$session_token"
            VIOLATIONS=$((VIOLATIONS + 1))
            ;;
          *)
            if printf '%s' "$session_token" | grep -qE '\$[A-Za-z_]'; then
              printf '%s:%s: -session-token %q contains a shell variable reference ($VAR) -- the transcript records the command text before the shell expands it, so this value would be recorded identically by every caller and identify nothing; write a literal token instead\n' "$f" "$lineno" "$session_token"
              VIOLATIONS=$((VIOLATIONS + 1))
            fi
            ;;
        esac
      fi

      guess="$(guess_placeholder "$cmd")"
      if [[ -n "$guess" ]]; then
        printf '%s:%s: `stage begin` names a guess, not a resolved change (%s) -- a mark writes, so a guessed name bootstraps a change row that outlives the run; wait until the change name is resolved before marking\n' "$f" "$lineno" "$guess"
        VIOLATIONS=$((VIOLATIONS + 1))
      fi
    done < <(assemble_calls "$f")
  done
  unset FILES
done

if [[ "$VIOLATIONS" -gt 0 ]]; then
  echo "check-stage-mark-calls: $VIOLATIONS violation(s) across $CHECKED \`stage begin\` call(s) checked" >&2
  exit 1
fi

echo "✓ Stage-mark-calls guard: clean ($CHECKED \`stage begin\` call(s) checked)"
exit 0
