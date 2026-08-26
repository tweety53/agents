#!/usr/bin/env bash
# check-stage-mark-calls.sh — every `myflow stage begin` in a skill must carry
# a session token and a harness, the session token must be a literal, and the harness must
# NOT be — the two required flags are wrong in opposite directions, and this
# script rejects both. Every `myflow record dispatch` must carry a literal
# session token too, for the same reason and by the same two checks.
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
# `stats/cmd/flow/stage.go`'s `validateSessionToken` rejects at the CLI, and
# `internal/api/stages.go`'s `validateSessionTokenShape` rejects again at the store,
# as defence in depth.
#
# `myflow record dispatch` IS SCANNED FOR THE SESSION-TOKEN RULES TOO
# (KAN-258). The record verb takes `-session-token` for precisely the reason
# `stage begin` does: the daemon finds that literal token in the calling
# session's own transcript and binds the dispatch row to the session that
# wrote it. A substituted value there lands in every transcript as the same
# unexpanded string and so discriminates between no two sessions — the whole
# mechanism, silently gone. A guard that enforced the literal on one verb and
# not the other would be a guard with a hole, and the record verb's call
# sites are numerous (four in `skills/myflow-do/SKILL.md` alone). So both
# session-token rules — the flag is present, and its value is a literal —
# are applied to `myflow stage begin` and `myflow record dispatch` alike.
#
# The other rules stay with the verb that owns them. `-harness` is a `stage
# begin` flag and `record dispatch` has none (a dispatch row inherits its
# harness from the stage run it belongs to), so requiring one on a record
# call would demand a flag the CLI would reject. The guessed-change-name
# rule likewise stays on `stage begin`, whose begin handler is the one that
# bootstraps a change row from a name it has never seen.
#
# Only `myflow record dispatch` is scanned: `record finding`, `record
# status`, `record render` and `record journal-count` take no session token
# at all, so there is nothing on them for these rules to say.
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
# `/flow-status` marks nothing, per its own row in README.md's Level 1
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
# Exit codes: 0 every scanned call is compliant; 1 at least one
# violation was found; 2 the guard cannot answer at all (a bad path, or an
# awk/grep failure — never read as "no findings").
set -uo pipefail

die() {
  echo "check-stage-mark-calls: $*" >&2
  exit 2
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# coverage_record / coverage_declare / coverage_report / coverage_verdict —
# per-file coverage reporting and the declared-vs-undeclared-zero decision,
# owned once in lib/coverage.sh rather than reinvented here. See that file's
# header for why (KAN-197) and check-guard-symlinks.sh / check-references.sh
# for the pattern this guard follows.
source "$SCRIPT_DIR/lib/coverage.sh"

if [[ $# -gt 0 ]]; then
  TARGETS=("$@")
else
  TARGETS=("$REPO_ROOT/skills")
fi

for t in "${TARGETS[@]}"; do
  [[ -e "$t" ]] || die "no such file or directory: $t"
done

# assemble_calls <file> — print one line per checked invocation found in
# <file>, as "<start-line>\t<verb>\t<joined command text>", continuation
# lines folded in. Two verbs are checked, `stage begin` and `record
# dispatch`; the verb is carried through so the rules below can be applied
# per verb rather than re-derived from the command text. Reads a real file's
# physical lines with awk, one pass, no shell loop over its content (a shell
# `while read` would mangle a backslash the same way the continuation logic
# here has to preserve deliberately).
assemble_calls() {
  local file="$1"
  awk '
    BEGIN { incall = 0; buf = ""; startline = 0; verb = "" }
    function stripcr(s) { sub(/\r$/, "", s); return s }
    {
      line = stripcr($0)
      if (incall) {
        buf = buf " " line
        if (line ~ /\\[[:space:]]*$/) {
          sub(/\\[[:space:]]*$/, "", buf)
          next
        }
        printf "%d\t%s\t%s\n", startline, verb, buf
        incall = 0
        buf = ""
        next
      }
      verb = ""
      if (line ~ /^[[:space:]]*myflow stage begin([[:space:]]|\\|$)/) {
        verb = "myflow stage begin"
      } else if (line ~ /^[[:space:]]*myflow record dispatch([[:space:]]|\\|$)/) {
        verb = "myflow record dispatch"
      }
      if (verb != "") {
        if (line ~ /\\[[:space:]]*$/) {
          incall = 1
          startline = NR
          buf = line
          sub(/\\[[:space:]]*$/, "", buf)
          next
        }
        printf "%d\t%s\t%s\n", NR, verb, line
      }
    }
    END {
      # A file ending mid-continuation is a malformed fixture/skill, not a
      # silently dropped call -- report it as its own finding downstream by
      # emitting what was gathered rather than swallowing it.
      if (incall) printf "%d\t%s\t%s\n", startline, verb, buf
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

# KAN-197 FIX ROUND, F1 — the corpus is narrowed to the members CAPABLE of
# carrying a `stage begin` call, not every Markdown file under skills/. Before
# this fix the scan glob was '*.md', which reached 33 files while only a
# SKILL.md or pipeline.md can ever hold a checked invocation (a contract
# doc, a rationale file or a reviewer-prompt file structurally cannot) — so 28
# of 33 declared entries were declared not because their zero was meaningful
# but because they were never candidates, and every new rationale or
# reviewer-prompt file added under skills/ would need a line here purely to
# stay green. Narrowing the FIND pattern below to these two filenames removes
# that whole non-candidate class from the corpus rather than declaring around
# it. Detection is preserved: a SKILL.md or pipeline.md that loses all its
# marks is still a member of this narrower corpus and still fires — see
# test-check-stage-mark-calls.sh's mutation case for the check.
#
# EXPECTED-ZERO FILES — after narrowing, only two SKILL.md files under
# skills/ genuinely carry no checked call site of their own — neither a
# `stage begin` nor a `record dispatch` (measured 2026-08-18 for `stage
# begin` and re-measured 2026-08-22 for `record dispatch`, by grepping each
# directly against this guard's own call-detection pattern), plus
# flow-status/SKILL.md declared separately
# below with its own by-contract reason. Each gets its OWN one-line
# justification for WHY its zero is legitimate by design (KAN-197 F2),
# not a shared string attesting only to how the zero was measured.
# Two PARALLEL indexed arrays, matched by position — never an associative
# array, which is bash-4-only and this repository's floor is bash 3.2 (see
# scripts/lib/coverage.sh's own header for the same constraint).
EXPECTED_ZERO_FILES=(
  "skills/myflow-contracts/SKILL.md"
  "skills/flow-research/SKILL.md"
  "skills/flow-settings/SKILL.md"
  "skills/flow/SKILL.md"
)
EXPECTED_ZERO_REASONS=(
  "the contracts index — shared prose loaded by several command skills; it is never itself run as a command, so it marks no stage and dispatches no subagent of its own"
  "a thinking-partner research mode with no implementation or verification stage to mark and no subagent to dispatch — the same reason check-guard-symlinks.sh declares it expected-zero"
  "a standalone settings command with no per-change state, no implementation or verification stage to mark, and no subagent to dispatch — the same reason check-guard-symlinks.sh declares it expected-zero"
  "a legitimate zero-mark router file — it resolves state and dispatches into the topic file (brainstorm.md, implement.md, review-panel.md, verify-and-handoff.md, integrate.md, archive.md) that owns the phase in force; every flow.* mark lives in one of those phase files, never in this one"
)

# declare_expected_zeros — called ONLY for the guard's own default, full-
# corpus scan (no explicit CLI targets). This guard's declared list is a
# hardcoded set of THIS REPOSITORY's own paths; a caller-supplied target
# (the companion test harness always passes one, pointed at a sandboxed
# mktemp fixture — see new_fixture above) scans a wholly different, smaller
# tree where none of these paths exist. Declaring them there would make
# every one of them a KAN-197 F3 "declared but never recorded" violation
# even though they are simply not part of that run's corpus at all — not a
# real staleness. Gating on "no CLI args" keeps F3's protection exactly where
# it means something: this guard's own real, default invocation, exercised
# bare by test-check-stage-mark-calls.sh's own case 24.
declare_expected_zeros() {
  local i f
  for i in "${!EXPECTED_ZERO_FILES[@]}"; do
    f="${EXPECTED_ZERO_FILES[$i]}"
    if ! coverage_declare "$f" "${EXPECTED_ZERO_REASONS[$i]}"; then
      die "coverage_declare failed for '$f' (see stderr above)"
    fi
  done
  # flow-status marks nothing BY CONTRACT, not merely as a measured fact
  # like the files above: it is a read-only status report, and a stage mark
  # or a dispatch record it wrote would record work nobody did. Declared with its own reason
  # rather than folded into the generic list, per this task's own note.
  if ! coverage_declare "skills/flow-status/SKILL.md" \
    "read-only status report; writes no stage marks by contract — a mark here would record work nobody did"; then
    die "coverage_declare failed for 'skills/flow-status/SKILL.md' (see stderr above)"
  fi
}

VIOLATIONS=0
CHECKED=0
if [[ $# -eq 0 ]]; then
  declare_expected_zeros
fi

for target in "${TARGETS[@]}"; do
  if [[ -d "$target" ]]; then
    # KAN-197 F1: narrowed to the two filenames capable of ever carrying a
    # checked call — see the corpus comment above EXPECTED_ZERO_FILES.
    FILES=()
    while IFS= read -r -d '' f; do
      FILES+=("$f")
    done < <(find "$target" -type f \( -name 'SKILL.md' -o -name 'pipeline.md' \) -print0)
    if [[ ${#FILES[@]} -eq 0 ]]; then
      # KAN-197 F7: a directory target that enumerates to ZERO candidate
      # files produced no file for the loop below to iterate, so
      # coverage_record was never called at all for it -- the target
      # vanished from coverage entirely rather than being caught, the exact
      # vacuous-pass shape this whole change exists to close. Record the
      # TARGET itself with a zero count so it becomes a corpus member like
      # any other -- undeclared, it is now a violation naming itself. This
      # is the same "record regardless of what follows" discipline
      # check-vocabulary.sh's collect_hits already carries for its own
      # per-target FILES_COUNT_OUT ("Set regardless of what follows,
      # including the early return").
      if ! coverage_record "${target#"$REPO_ROOT"/}" 0; then
        die "coverage_record failed for target '$target' (see stderr above)"
      fi
      unset FILES
      continue
    fi
  else
    FILES=("$target")
  fi

  for f in "${FILES[@]:-}"; do
    [[ -n "${f:-}" ]] || continue
    file_checked=0
    while IFS=$'\t' read -r lineno verb cmd; do
      [[ -n "$lineno" ]] || continue
      CHECKED=$((CHECKED + 1))
      file_checked=$((file_checked + 1))

      has_session_token=1
      if ! printf '%s' "$cmd" | grep -qE -- '(^|[[:space:]])-session-token([[:space:]]|=)'; then
        has_session_token=0
        printf '%s:%s: `%s` carries no -session-token -- required so the daemon can later bind this record to the session that wrote it (design.md, "bind after the fact, by a correlator the caller writes")\n' "$f" "$lineno" "$verb"
        VIOLATIONS=$((VIOLATIONS + 1))
      fi

      # `-harness` is a `stage begin` flag and `myflow record dispatch` has
      # none: a dispatch row inherits its harness from the stage run it
      # belongs to, so requiring one here would demand a flag the CLI would
      # reject. The rule is applied to the verb that takes it, not to every
      # scanned call.
      has_harness=1
      if [[ "$verb" == "myflow stage begin" ]] \
        && ! printf '%s' "$cmd" | grep -qE -- '(^|[[:space:]])-harness([[:space:]]|=)'; then
        has_harness=0
        printf '%s:%s: `myflow stage begin` carries no -harness -- required so a recorded run states which harness marked it\n' "$f" "$lineno"
        VIOLATIONS=$((VIOLATIONS + 1))
      fi

      if [[ "$verb" == "myflow stage begin" && "$has_harness" -eq 1 ]]; then
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

      guess=""
      [[ "$verb" == "myflow stage begin" ]] && guess="$(guess_placeholder "$cmd")"
      if [[ -n "$guess" ]]; then
        printf '%s:%s: `stage begin` names a guess, not a resolved change (%s) -- a mark writes, so a guessed name bootstraps a change row that outlives the run; wait until the change name is resolved before marking\n' "$f" "$lineno" "$guess"
        VIOLATIONS=$((VIOLATIONS + 1))
      fi
    done < <(assemble_calls "$f")
    # KAN-197 coverage: how many checked calls — `stage begin` and `record
    # dispatch` together — this guard actually found IN THIS FILE, per
    # scripts/lib/coverage.sh; distinct from CHECKED above, which is the
    # corpus-wide total. A file the glob reaches but which carries no checked
    # call at all is "nothing was checked here," visible even on a run that
    # goes on to report clean.
    #
    # KAN-197 F8: the return is checked rather than ignored. This guard runs
    # under `set -uo pipefail` with no `-e`, so an unchecked failure here
    # (e.g. the same file enumerated twice) would be silently swallowed and
    # the verdict would go on to report clean, contradicting its own
    # breakdown — exactly the shape F8 found live in this file before this
    # fix.
    if ! coverage_record "${f#"$REPO_ROOT"/}" "$file_checked"; then
      die "coverage_record failed for '$f' (see stderr above)"
    fi
  done
  unset FILES
done

# COVERAGE — per-file count of the calls this guard actually found, via
# scripts/lib/coverage.sh. A file whose coverage is zero and is
# not in EXPECTED_ZERO_FILES above is folded into the ordinary violation
# count here, exactly as check-references.sh and check-guard-symlinks.sh do
# it — never a separate exit status.
if ! coverage_verdict_out="$(coverage_verdict)"; then
  while IFS= read -r cvline; do
    [[ -n "$cvline" ]] || continue
    printf '%s:0: %s\n' "${cvline%%:*}" "${cvline#*: }"
    VIOLATIONS=$((VIOLATIONS + 1))
  done <<<"$coverage_verdict_out"
fi

if [[ "$VIOLATIONS" -gt 0 ]]; then
  echo "check-stage-mark-calls: $VIOLATIONS violation(s) across $CHECKED call site(s) checked" >&2
  exit 1
fi

echo "✓ Stage-mark-calls guard: clean ($CHECKED call site(s) checked)"
coverage_frag="$(coverage_report)"
[[ -n "$coverage_frag" ]] && printf '  %s\n' "$coverage_frag"
exit 0
