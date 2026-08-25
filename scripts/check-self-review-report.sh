#!/usr/bin/env bash
# check-self-review-report.sh — every report under docs/self-review/ carries
# all five self-review angles, each section carrying either a finding line
# or the none-marker, each finding line parseable with its label matching
# its section and a filed finding naming an issue key.
#
# Usage: scripts/check-self-review-report.sh [dir]
#
# With no argument it scans `docs/self-review/`, resolved relative to the
# repo root — the one tree self-review reports live under. A caller may pass
# an explicit directory instead (the companion test harness does, against
# mktemp fixtures).
#
# WHAT THIS CATCHES. KAN-200's own report shape — five `##` sections, one
# parseable line per finding, an explicit none-marker for an empty angle —
# was, before this guard, checked by nothing but a reviewer's prose reading.
# An angle silently omitted from a report read exactly like an angle that
# produced nothing, which is how KAN-73's cost angle passed unnoticed while
# its section existed. This guard makes an incomplete or malformed report a
# loud, mechanical fact instead.
#
# THE REPORT SHAPE, in the frozen tree (openspec/changes/archive/
# 2026-08-18-kan-200-self-review-filing-ask-per-angle/tasks.md, "The
# report shape this change defines"):
#
#   ## <prose> — `<label>`
#
#   - **[<label>]** <text> — filed: <KEY>
#   - **[<label>]** <text> — declined
#
# An angle with no findings carries `_none — this angle produced no
# findings._` in place of finding lines, never both a marker and finding
# lines together. The label on a finding line must match the label in its
# own section's heading. `<KEY>` is an uppercase project key, a hyphen, and
# digits — e.g. `KAN-201`; anything else (empty, `yes`, a bare number, a
# trailing hyphen) is a violation naming the malformed key.
#
# The five `##` sections appear in the order the angle table below states,
# each exactly once — an out-of-order or duplicate section is a named
# violation. A heading at ANY level (`#`, `###`, …) ends the current
# section, not only `##`. A finding-shaped line appearing before any
# recognized section heading is a named violation rather than a silent drop.
#
# THE FIVE ANGLE LABELS, in the order the report shape states them:
#   myflow-fix, myflow-cost, myflow-improvement, myflow-automation,
#   myflow-stats-app
# (the frozen openspec/changes/archive/2026-08-18-kan-200-self-review-
# filing-ask-per-angle/specs/myflow-self-review/spec.md, "One combined
# reasoning pass" angle table).
#
# PER-REPORT COVERAGE, via scripts/lib/coverage.sh. Each report's recorded
# count is the number of section-level checks this guard actually performed
# on it: one per angle heading it FOUND (not one per angle attempted), plus
# one per finding line it found under a recognized heading. A report that is
# not recognizable as a self-review report at all — none of the five
# headings present anywhere in it — gets NO per-section "missing" findings,
# because there was nothing in it to check against; it is instead flagged
# entirely through coverage.sh's undeclared-zero mechanism, so a report
# checked for nothing is a named, failing fact rather than a silent pass
# (KAN-73's own regression shape, reproduced here as this guard's own
# harness case 7). A report that IS recognizable — at least one of the five
# headings present — but is missing one or more of the others gets an
# explicit "missing section" finding for each absent one, on top of whatever
# non-zero count its present sections contribute.
#
# THE DECLARED PRE-RULE LIST. Seventeen reports under docs/self-review/
# predate this shape and are declared, by name, with the reason
# "predates the five-angle report shape (KAN-200)" — never inferred from a
# date, a sort order, or a marker inside the report itself, because a marker
# a new report can forget to write is a mechanism for passing without being
# checked, which is the exact outcome the declaration exists to prevent (see
# the ADDED "A guard checks every self-review report" requirement in this
# change's spec). Declaration only runs on this guard's own default, bare
# invocation (no CLI argument) — mirroring check-stage-mark-calls.sh's
# `declare_expected_zeros`, gated the same way and for the same reason: the
# companion test harness always passes an explicit, sandboxed mktemp
# fixture directory, a wholly different and smaller tree where none of these
# seventeen real basenames exist, and declaring them there would make every
# one of them a coverage.sh "declared but never recorded" violation for a
# member that was simply never part of that run's corpus at all.
#
# ADOPTED, NOT INVENTED: scripts/lib/panel-record.sh's header states three
# disciplines for every guard in this repository — `-a` on every grep, the
# `rc > 1` split between "no match" and a real error, and `--` before every
# path. This guard uses no `grep` at all (section and finding-line
# extraction is done in one `awk` pass per file, then classified in bash),
# so the first and third disciplines have no call site here, exactly as
# scripts/lib/coverage.sh's own header states for the same reason. The
# posture behind them still applies: `find` and the per-file `awk` call
# below are both checked for a non-zero exit rather than folded into a
# reassuring empty result.
#
# Exit codes: 0 clean, 1 violations found (including an undeclared-zero
# report or an empty corpus), 2 cannot answer at all (a missing or
# unreadable target directory, or an internal coverage.sh call failing).
set -uo pipefail

die() {
  echo "check-self-review-report: $*" >&2
  exit 2
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# coverage_record / coverage_declare / coverage_report / coverage_verdict —
# per-report coverage reporting and the declared-vs-undeclared-zero
# decision, owned once in lib/coverage.sh rather than reinvented here. See
# that file's header for why (KAN-197) and check-stage-mark-calls.sh for the
# pattern this guard follows.
source "$SCRIPT_DIR/lib/coverage.sh"

[[ $# -le 1 ]] || die "usage: $0 [dir]"

if [[ $# -eq 1 ]]; then
  TARGET="$1"
else
  TARGET="$REPO_ROOT/docs/self-review"
fi

[[ -e "$TARGET" ]] || die "no such directory: $TARGET"
[[ -d "$TARGET" ]] || die "not a directory: $TARGET"
[[ -r "$TARGET" ]] || die "cannot read directory: $TARGET"

# The five angle labels, in the report shape's own order.
ANGLE_LABELS=(myflow-fix myflow-cost myflow-improvement myflow-automation myflow-stats-app)
ANGLE_LABELS_JOINED="${ANGLE_LABELS[*]}"

# Regex patterns are kept in variables and referenced unquoted in `[[ =~ ]]`
# below rather than written inline: bash's quote-removal strips a literal
# `\]`/`\*` written directly after `=~` before the regex engine ever sees
# it (`\]\*\*` becomes `]**`, which glibc's regex compiler then rejects as
# "repetition-operator operand invalid") — a variable reference is not
# subject to that removal, so the escapes survive intact.
#
# FINDING_SHAPE_RE is the LOOSE predicate: a line that is attempting to be a
# finding line, well-formed or not — a leading `-`, any amount (including
# zero) of horizontal whitespace, then `**[`. FINDING_LINE_RE is the STRICT
# predicate a shape must additionally satisfy to be accepted as a
# well-formed finding: exactly one space after the `-`, a `]**`, one space,
# then text. Both scan_report's awk (the ORPHAN check, for a finding-shaped
# line before any heading) and the in-section check below (for a
# finding-shaped line that fails the strict pattern) test against this same
# FINDING_SHAPE_RE variable, so the two can never disagree about what counts
# as finding-shaped — before this fix they used separately hardcoded
# patterns that did disagree, letting a malformed line slip through silently
# whenever it happened to fall inside a recognized section.
FINDING_SHAPE_RE='^-[[:space:]]*\*\*\['
# awk's `-v` assignment runs the same C-style backslash-escape processing as
# an awk string literal, and — like glibc's regex compiler above, a
# different tool hitting the same class of problem — silently drops the
# backslash from any escape sequence it does not itself recognize (`\*` and
# `\[` are not on that list), corrupting the pattern before scan_report's
# dynamic regexp ever sees it. Doubling every backslash first survives that:
# `\\*` is a recognized escape (a literal backslash), so what awk stores is
# the same `\*` this variable already holds — derived here, not
# hand-maintained a second time, so the two representations cannot drift.
FINDING_SHAPE_RE_AWK="${FINDING_SHAPE_RE//\\/\\\\}"
FINDING_LINE_RE='^-[[:space:]]\*\*\[([^]]+)\]\*\*[[:space:]](.+)$'
DISPOSITION_FILED_RE='^filed:[[:space:]]*(.*)$'
# An uppercase project key, a hyphen, and digits — e.g. `KAN-201`. Anything
# else naming itself `filed:` (empty, `yes`, `201`, `KAN-`) is a violation.
ISSUE_KEY_RE='^[A-Z][A-Z0-9]*-[0-9]+$'

# The seventeen reports present before this shape existed (measured against
# docs/self-review/ once task 1 restored the seventeenth, KAN-197's own
# report) — written out by name, per this guard's own header above.
DECLARED_REPORTS=(
  "kan-100-myflow-get-rid-of-staging-use-commits-self-review.md"
  "kan-106-slim-the-myflow-skills-cut-meta-prose-extract-self-review.md"
  "kan-107-remove-manual-test-guide-and-gate-self-review.md"
  "kan-108-cut-the-time-and-token-cost-of-a-myflow-do-run-self-review.md"
  "kan-109-optimize-myflow-agent-token-and-time-cost-self-review.md"
  "kan-110-lighter-auto-code-review-by-default-self-review.md"
  "kan-111-myflow-fast-self-review.md"
  "kan-13-myflow-planning-and-status-fixes-self-review.md"
  "kan-15-parallel-myflow-do-task-lanes-self-review.md"
  "kan-153-kan-108-follow-up-self-review.md"
  "kan-16-myflow-stats-app-self-review.md"
  "kan-197-require-mutation-test-for-every-guard-self-review.md"
  "kan-23-myflow-self-review-self-review.md"
  "kan-73-install-guard-scripts-alongside-skills-self-review.md"
  "kan-82-cut-myflow-per-command-token-overhead-self-review.md"
  "kan-87-cut-per-command-load-further-self-review.md"
  "kan-95-slim-the-myflow-contract-files-self-review.md"
)
DECLARED_REASON="predates the five-angle report shape (KAN-200)"

# declare_pre_rule — called ONLY for this guard's own default, full-corpus
# scan (no explicit CLI target). See the header comment above for why.
declare_pre_rule() {
  local b
  for b in "${DECLARED_REPORTS[@]}"; do
    if ! coverage_declare "docs/self-review/$b" "$DECLARED_REASON"; then
      die "coverage_declare failed for 'docs/self-review/$b' (see stderr above)"
    fi
  done
}

# `$# -eq 0` is 1 only for this guard's own default, full-corpus scan (no
# explicit CLI target) — the one mode where declare_pre_rule below actually
# runs and registers the seventeen basenames with coverage.sh. The main loop
# further down consults the same test again (F8): is_declared_basename's own
# list is a static, hardcoded array present regardless of mode, so checking
# it without also checking `$# -eq 0` would let a file merely SHARING one of
# the seventeen basenames skip every content check in EXPLICIT-directory mode
# too — the mode the companion test harness always uses, against a wholly
# different sandboxed tree where no coverage_declare call for that name was
# ever made. Read directly rather than cached in a flag variable: this script
# never calls `shift` or `set --`, so `$#` cannot change between here and the
# main loop's own check of it, and a flag would only restate what `$#`
# already says.
if [[ $# -eq 0 ]]; then
  declare_pre_rule
fi

# is_declared_basename <basename> -> exit 0 if <basename> is one of the
# seventeen pre-rule reports declared above, 1 otherwise. Only meaningful
# when `$# -eq 0` — see the comment above.
is_declared_basename() {
  local target="$1" b
  for b in "${DECLARED_REPORTS[@]}"; do
    [[ "$b" == "$target" ]] && return 0
  done
  return 1
}

# scan_report <file> -> prints one record per line of output, fields
# separated by ASCII Unit Separator (0x1F, "\037" — never tab): "HEADING<US>
# <label><US><lineno>" for each of the five angle headings found, "BODY<US>
# <label><US><lineno><US><line>" for every line under a recognized heading up
# to the next heading (any level) or EOF, "ORPHAN<US><US><lineno><US>
# <line>" (empty label field) for a finding-shaped line that appears before
# any recognized heading has been seen (F6 — such a line was previously
# dropped in silence), and "CTRLBYTE<US><US><lineno>" (empty label, no
# trailing content field) for a report line that itself contains a raw 0x1F
# byte — checked BEFORE any other classification, so a heading, body, or
# orphan line carrying one is never misparsed as a different kind. The raw
# line is deliberately NOT relayed in this record: a literal 0x1F inside it
# would itself act as a field delimiter one field short, corrupting the
# record protocol's own framing rather than surviving as content.
#
# US rather than tab (G3 — round 2): the original protocol joined fields with
# a literal tab and split them back apart with `IFS=$'\t' read`, but tab is
# an IFS-*whitespace* character, so bash's read collapses RUNS of it and
# trims it at a record's edges. That silently ate a literal tab inside the
# raw line carried in the last field (`rest`) — once at a trailing tab on a
# `filed: <KEY>` line (letting a malformed disposition through validation
# unchanged), and once at a leading tab on an indented finding-shaped body
# line (collapsing it into the adjacent field delimiter and making the line
# read as if un-indented, so a section that should have failed "neither a
# finding nor the none-marker" silently passed instead). US is not an
# IFS-whitespace character and never appears in a self-review report's prose,
# so `IFS=$'\037' read` below splits on it literally: runs are not collapsed,
# edges are not trimmed, and the `rest` field survives byte-for-byte for
# every byte a report line can actually carry — which excludes 0x1F itself.
# A raw 0x1F in the line is not relayed through `rest` at all: it is caught
# by the CTRLBYTE check below, before this protocol's own framing would
# otherwise have to treat it as an unplanned field boundary (a trailing one
# is dropped by `read` the same way a trailing IFS-whitespace run is, which
# would silently forgive a none-marker line differing from the real one only
# by a trailing 0x1F). This replaces round 1's F6 fix, which gave the
# ORPHAN record a non-empty placeholder label as a workaround for the same
# root cause — the label field here is genuinely empty, because the protocol
# itself no longer loses empty fields.
#
# A `##` heading whose text contains none of the five backtick-quoted labels
# resets the current section to none, so lines under it are attributed to no
# angle at all. A heading at ANY OTHER level (`#`, `###`, `####`, …) also
# ends the current section — it only does not itself become a new one, since
# the report shape's own headings are `##` (F3: a `###` heading used to fall
# through neither check and be silently folded into the previous section's
# body). A trailing `\r` is stripped from every line before any pattern is
# tested, so a CRLF-saved report is read the same as an LF one (F9).
#
# The ORPHAN test uses `shape` (passed in as FINDING_SHAPE_RE_AWK, the
# backslash-doubled relay of FINDING_SHAPE_RE — see that variable's own
# comment for why the doubling is necessary) — the same loose finding-shape
# predicate the in-section check below tests a BODY line against (G2 —
# round 2) — so the two can never again disagree about what counts as
# finding-shaped.
#
# One `awk` pass per file rather than five separate `grep` passes, so the
# five checks below cannot disagree about where one section ends and the
# next begins.
scan_report() {
  local file="$1"
  awk -v labels="$ANGLE_LABELS_JOINED" -v shape="$FINDING_SHAPE_RE_AWK" '
    BEGIN { n = split(labels, L, " "); cur = ""; US = sprintf("%c", 31) }
    {
      line = $0
      sub(/\r$/, "", line)
      if (index(line, US) > 0) {
        printf "CTRLBYTE%s%s%s%d\n", US, "", US, NR
        next
      }
      if (line ~ /^##[[:space:]]/) {
        cur = ""
        for (i = 1; i <= n; i++) {
          pat = "`" L[i] "`"
          if (index(line, pat) > 0) {
            cur = L[i]
            printf "HEADING%s%s%s%d\n", US, cur, US, NR
            break
          }
        }
        next
      }
      if (line ~ /^#+[[:space:]]/) {
        cur = ""
        next
      }
      if (cur != "") {
        printf "BODY%s%s%s%d%s%s\n", US, cur, US, NR, US, line
      } else if (line ~ shape) {
        printf "ORPHAN%s%s%d%s%s\n", US, US, NR, US, line
      }
    }
  ' "$file"
}

VIOLATIONS=0
CHECKED_REPORTS=0
DECLARED_COUNT=0

# find's own exit status is captured directly, not folded through a pipe
# (F2): `find ... | sort -z` in a process substitution loses find's status
# entirely — only sort's is visible via $? after the loop, which is why an
# untraversable directory (readable but not searchable) used to be silently
# read as "empty corpus" and reported at exit 1 instead of exit 2 ("cannot
# answer"). Writing find's own output to a temp file lets its exit code be
# checked with nothing else in between, exactly as this guard's own header
# above already claims it is.
FIND_TMP="$(mktemp "${TMPDIR:-/tmp}/check-self-review-report-find.XXXXXX")" \
  || die "cannot create a temp file for find's output"
trap 'rm -f "$FIND_TMP"' EXIT

if ! find "$TARGET" -maxdepth 1 -type f -name '*.md' -print0 >"$FIND_TMP"; then
  die "find failed while scanning '$TARGET' (permission denied, or another find error)"
fi

FILES=()
while IFS= read -r -d '' f; do
  FILES+=("$f")
done < <(sort -z "$FIND_TMP")
rm -f "$FIND_TMP"
trap - EXIT

for f in "${FILES[@]:-}"; do
  [[ -n "${f:-}" ]] || continue
  rel="${f#"$REPO_ROOT"/}"
  base="$(basename -- "$f")"

  if [[ $# -eq 0 ]] && is_declared_basename "$base"; then
    if ! coverage_record "$rel" 0; then
      die "coverage_record failed for '$rel' (see stderr above)"
    fi
    DECLARED_COUNT=$((DECLARED_COUNT + 1))
    continue
  fi

  # Reset per-report state: one slot per angle label, walked by index —
  # bash 3.2 is the floor (macOS's own /bin/bash), matching
  # scripts/lib/coverage.sh's own constraint.
  HEADING_FOUND=(0 0 0 0 0)
  HEADING_LINE=(0 0 0 0 0)
  HAS_NONE=(0 0 0 0 0)
  FINDING_COUNT=(0 0 0 0 0)
  # LAST_IDX tracks the highest angle index whose heading has been seen so
  # far, so a heading appearing before it in ANGLE_LABELS order is named as
  # out of order (F4). -1 means no heading seen yet.
  LAST_IDX=-1

  scan_rc=0
  scan_out="$(scan_report "$f")" || scan_rc=$?
  if [[ "$scan_rc" -ne 0 ]]; then
    die "awk failed while scanning '$f' (exit $scan_rc)"
  fi

  while IFS=$'\037' read -r kind label lineno rest; do
    [[ -n "$kind" ]] || continue

    if [[ "$kind" == "ORPHAN" ]]; then
      printf '%s:%s: finding line appears before any recognized section heading\n' "$rel" "$lineno"
      VIOLATIONS=$((VIOLATIONS + 1))
      continue
    fi

    if [[ "$kind" == "CTRLBYTE" ]]; then
      printf '%s:%s: unsupported control byte (0x1F, ASCII Unit Separator) in report content\n' \
        "$rel" "$lineno"
      VIOLATIONS=$((VIOLATIONS + 1))
      continue
    fi

    idx=-1
    for i in "${!ANGLE_LABELS[@]}"; do
      [[ "${ANGLE_LABELS[$i]}" == "$label" ]] && { idx="$i"; break; }
    done
    [[ "$idx" -ge 0 ]] || continue

    case "$kind" in
      HEADING)
        if [[ "${HEADING_FOUND[$idx]}" -eq 1 ]]; then
          printf '%s:%s: duplicate section for angle `%s` (first seen at line %s)\n' \
            "$rel" "$lineno" "$label" "${HEADING_LINE[$idx]}"
          VIOLATIONS=$((VIOLATIONS + 1))
        fi
        if [[ "$idx" -lt "$LAST_IDX" ]]; then
          printf '%s:%s: section for angle `%s` is out of order (expected after `%s`)\n' \
            "$rel" "$lineno" "$label" "${ANGLE_LABELS[$LAST_IDX]}"
          VIOLATIONS=$((VIOLATIONS + 1))
        fi
        HEADING_FOUND[$idx]=1
        HEADING_LINE[$idx]="$lineno"
        [[ "$idx" -gt "$LAST_IDX" ]] && LAST_IDX="$idx"
        ;;
      BODY)
        if [[ "$rest" == "_none — this angle produced no findings._" ]]; then
          HAS_NONE[$idx]=1
        elif [[ "$rest" =~ $FINDING_LINE_RE ]]; then
          # A well-formed finding line: exactly one space after the `-`.
          # (G2 — round 2): checked BEFORE the loose FINDING_SHAPE_RE below,
          # so a well-formed line is never also reported as malformed.
          finding_label="${BASH_REMATCH[1]}"
          finding_text="${BASH_REMATCH[2]}"
          FINDING_COUNT[$idx]=$((FINDING_COUNT[$idx] + 1))

          if [[ "$finding_label" != "$label" ]]; then
            printf '%s:%s: finding line label `%s` does not match its section `%s`\n' \
              "$rel" "$lineno" "$finding_label" "$label"
            VIOLATIONS=$((VIOLATIONS + 1))
          fi

          disposition="${finding_text##*— }"
          if [[ "$disposition" == "declined" ]]; then
            :
          elif [[ "$disposition" =~ $DISPOSITION_FILED_RE ]]; then
            key="${BASH_REMATCH[1]}"
            if [[ -z "$key" ]]; then
              printf '%s:%s: finding line marked filed with no issue key\n' "$rel" "$lineno"
              VIOLATIONS=$((VIOLATIONS + 1))
            elif [[ ! "$key" =~ $ISSUE_KEY_RE ]]; then
              printf '%s:%s: finding line marked filed with a malformed issue key `%s` (want an uppercase project key, a hyphen, digits — e.g. KAN-201)\n' \
                "$rel" "$lineno" "$key"
              VIOLATIONS=$((VIOLATIONS + 1))
            fi
          else
            printf '%s:%s: finding line disposition is neither `filed: <KEY>` nor `declined`\n' \
              "$rel" "$lineno"
            VIOLATIONS=$((VIOLATIONS + 1))
          fi
        elif [[ "$rest" =~ $FINDING_SHAPE_RE ]]; then
          # Finding-shaped (a `-`, optional whitespace, `**[`) but not
          # well-formed — e.g. more than one space after the `-` (G2 —
          # round 2). Without this branch such a line was silently dropped:
          # not counted as a finding, not flagged, invisible to every check
          # below. The same FINDING_SHAPE_RE the ORPHAN check above tests a
          # pre-heading line against is used here for an in-section one, so
          # the two checks cannot disagree about what counts as
          # finding-shaped — a plain prose line (no `-**[` prefix at all)
          # still matches neither pattern and stays ignored, exactly as
          # ordinary body prose should.
          printf '%s:%s: finding-shaped line is malformed (want "- **[%s]** <text> — filed: <KEY>" or "- **[%s]** <text> — declined")\n' \
            "$rel" "$lineno" "$label" "$label"
          VIOLATIONS=$((VIOLATIONS + 1))
        fi
        ;;
    esac
  done <<<"$scan_out"

  found_count=0
  for i in "${!ANGLE_LABELS[@]}"; do
    [[ "${HEADING_FOUND[$i]}" -eq 1 ]] && found_count=$((found_count + 1))
  done

  count=0
  if [[ "$found_count" -gt 0 ]]; then
    for i in "${!ANGLE_LABELS[@]}"; do
      label="${ANGLE_LABELS[$i]}"
      if [[ "${HEADING_FOUND[$i]}" -eq 0 ]]; then
        printf '%s: missing section for angle `%s`\n' "$rel" "$label"
        VIOLATIONS=$((VIOLATIONS + 1))
        continue
      fi
      count=$((count + 1))
      if [[ "${HAS_NONE[$i]}" -eq 0 && "${FINDING_COUNT[$i]}" -eq 0 ]]; then
        printf '%s:%s: section for angle `%s` carries neither a finding line nor the none-marker\n' \
          "$rel" "${HEADING_LINE[$i]}" "$label"
        VIOLATIONS=$((VIOLATIONS + 1))
      fi
      if [[ "${HAS_NONE[$i]}" -eq 1 && "${FINDING_COUNT[$i]}" -gt 0 ]]; then
        printf '%s:%s: section for angle `%s` carries both the none-marker and %s finding line(s)\n' \
          "$rel" "${HEADING_LINE[$i]}" "$label" "${FINDING_COUNT[$i]}"
        VIOLATIONS=$((VIOLATIONS + 1))
      fi
      count=$((count + FINDING_COUNT[i]))
    done
  fi

  if ! coverage_record "$rel" "$count"; then
    die "coverage_record failed for '$rel' (see stderr above)"
  fi
  CHECKED_REPORTS=$((CHECKED_REPORTS + 1))
done

# COVERAGE — a report checked for nothing (found_count above was 0, so no
# per-section "missing" finding fired for it at all) is folded into the
# ordinary violation count here, via scripts/lib/coverage.sh, exactly as
# check-stage-mark-calls.sh and check-guard-symlinks.sh do it — never a
# separate exit status. This is also how an empty corpus (no *.md file
# found at all) is reported: coverage_record was never called for any
# member, and coverage_verdict's own empty-corpus check fires.
if ! coverage_verdict_out="$(coverage_verdict)"; then
  while IFS= read -r cvline; do
    [[ -n "$cvline" ]] || continue
    printf '%s:0: %s\n' "${cvline%%:*}" "${cvline#*: }"
    VIOLATIONS=$((VIOLATIONS + 1))
  done <<<"$coverage_verdict_out"
fi

if [[ "$VIOLATIONS" -gt 0 ]]; then
  echo "check-self-review-report: $VIOLATIONS violation(s) across $CHECKED_REPORTS report(s) checked, $DECLARED_COUNT declared pre-rule" >&2
  exit 1
fi

TOTAL_REPORTS=$((CHECKED_REPORTS + DECLARED_COUNT))
printf 'SELF-REVIEW-REPORT-OK: %s — %s report(s), %s checked, %s declared pre-rule\n' \
  "$TARGET" "$TOTAL_REPORTS" "$CHECKED_REPORTS" "$DECLARED_COUNT"
coverage_frag="$(coverage_report)"
[[ -n "$coverage_frag" ]] && printf '  %s\n' "$coverage_frag"
exit 0
