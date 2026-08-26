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
# ADOPTED, NOT INVENTED: every guard in this repository holds to three
# disciplines — `-a` on every grep, the `rc > 1` split between "no match"
# and a real error, and `--` before every path. This guard uses neither
# `grep` nor `awk` at all — each report is read line by line in bash and
# every line is classified at the point it is read — so the first and third
# disciplines have no call site here, exactly as scripts/lib/coverage.sh's
# own header states for the same reason. The posture behind them still
# applies: `find`'s enumeration and each report file's own read are the two
# remaining I/O operations this guard has to face, and each is checked by its
# own exit status rather than folded into a reassuring empty result. Neither
# is checked by a precondition standing in for it — the status of the
# operation, or nothing. What that still leaves uncovered — a read that fails
# partway through a file it opened successfully — is recorded as a known
# limit of the all-bash shape under KAN-211.
#
# NO RECORD PROTOCOL, DELIBERATELY (KAN-211). This guard used to classify
# each report in one `awk` pass, serialize the result as records whose fields
# were joined with ASCII Unit Separator (0x1F, never tab), and re-parse those
# records back in bash. Three of KAN-200's findings — G2, G3 and H2 — were
# taxes on that one seam, each a manifestation of the same fact: a value had
# to survive a hand-written encoding across two languages with different
# escaping and splitting rules. H2 reopened the class one fix round after G3
# supposedly closed it, which is why KAN-211 collapsed the boundary rather
# than patching a fourth manifestation. Do not reintroduce one: the
# classification bodies below are already bash and already correct, and a
# second language has nothing to do here but re-encode what bash can read
# directly. A raw 0x1F byte in a report is ordinary content now — it was
# rejected only because it was that protocol's field delimiter, and the rule
# was retired with the protocol that created it.
#
# Exit codes: 0 clean, 1 violations found (including an undeclared-zero
# report or an empty corpus), 2 cannot answer at all (a missing or
# unreadable target directory, an unreadable report file inside it, or an
# internal coverage.sh call failing).
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
# then text. Both the orphan check below (for a finding-shaped line before
# any heading) and the in-section check beside it (for a finding-shaped line
# that fails the strict pattern) test against this same FINDING_SHAPE_RE
# variable, so the two can never disagree about what counts as
# finding-shaped — before this fix they used separately hardcoded
# patterns that did disagree, letting a malformed line slip through silently
# whenever it happened to fall inside a recognized section.
FINDING_SHAPE_RE='^-[[:space:]]*\*\*\['
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

  # Each line is classified where it is read — no intermediate representation
  # and no second language, per this file's own NO RECORD PROTOCOL note above.
  # A `##` heading whose text carries one of the five backtick-quoted labels
  # opens that angle's section; a `##` heading carrying none of them resets the
  # current section to none, so lines under it are attributed to no angle at
  # all. A heading at ANY OTHER level (`#`, `###`, `####`, …) also ends the
  # current section — it only does not itself become a new one, since the
  # report shape's own headings are `##` (F3: a `###` heading used to fall
  # through neither check and be silently folded into the previous section's
  # body). A trailing `\r` is stripped from every line before any pattern is
  # tested, so a CRLF-saved report is read the same as an LF one (F9).
  #
  # Redirected from the file rather than piped into: the body assigns to
  # VIOLATIONS and to the four per-angle arrays above, and a pipeline would run
  # it in a subshell whose assignments are then discarded.
  #
  # `cur` is the label of the section currently open ("" for none) and `idx`
  # its index into ANGLE_LABELS. The two are set together and `idx` is read
  # only where `cur` is already non-empty, so it needs no reset here — unlike
  # LAST_IDX above, which IS read before any heading is seen and so carries a
  # -1 sentinel.
  #
  # Whether the read happened at all is answered by the loop's OWN exit
  # status, captured immediately after `done < "$f"` below — the same
  # discipline `find`'s check above already follows, and for the same reason:
  # check the operation, never a precondition standing in for it. A failed
  # `< "$f"` redirection aborts the compound command before the body runs even
  # once, so the `while` exits non-zero; a loop that reached end-of-file exits
  # with the status of the last command its body ran, which is 0 by
  # construction (see the `:` closing the body). Measured on bash 3.2.57 — the
  # floor, macOS's own /bin/bash — and on bash 5.3.15: 0 for a normal file, an
  # empty file and a file with no trailing newline; 1 for a file the
  # redirection could not open.
  #
  # Without that check bash writes "Permission denied" to stderr, the body
  # never runs, every per-report counter stays at its reset zero, and the
  # report is folded into `coverage_record "$rel" 0` — reported as an
  # undeclared-zero coverage violation at exit 1, exactly like a genuinely
  # unrecognizable report, instead of the "cannot answer" this guard's header
  # promises.
  #
  # A `[[ -r "$f" ]]` precheck stood here and was removed (KAN-211 round 2):
  # it tests `access()`'s permission bits rather than the `open()` that
  # actually matters, and it opens a window in which the file can change
  # between the test and the read. The loop's own status has neither problem.

  lineno=0
  cur=""
  while IFS= read -r line || [[ -n "$line" ]]; do
    lineno=$((lineno + 1))
    line="${line%$'\r'}"

    if [[ "$line" =~ ^##[[:space:]] ]]; then
      cur=""
      for i in "${!ANGLE_LABELS[@]}"; do
        quoted_label='`'"${ANGLE_LABELS[$i]}"'`'
        if [[ "$line" == *"$quoted_label"* ]]; then
          cur="${ANGLE_LABELS[$i]}"
          idx="$i"
          break
        fi
      done
      [[ -n "$cur" ]] || continue

      if [[ "${HEADING_FOUND[$idx]}" -eq 1 ]]; then
        printf '%s:%s: duplicate section for angle `%s` (first seen at line %s)\n' \
          "$rel" "$lineno" "$cur" "${HEADING_LINE[$idx]}"
        VIOLATIONS=$((VIOLATIONS + 1))
      fi
      if [[ "$idx" -lt "$LAST_IDX" ]]; then
        printf '%s:%s: section for angle `%s` is out of order (expected after `%s`)\n' \
          "$rel" "$lineno" "$cur" "${ANGLE_LABELS[$LAST_IDX]}"
        VIOLATIONS=$((VIOLATIONS + 1))
      fi
      HEADING_FOUND[$idx]=1
      HEADING_LINE[$idx]="$lineno"
      [[ "$idx" -gt "$LAST_IDX" ]] && LAST_IDX="$idx"
      continue
    fi

    if [[ "$line" =~ ^#+[[:space:]] ]]; then
      cur=""
      continue
    fi

    if [[ -z "$cur" ]]; then
      # A finding-shaped line appearing before any recognized section heading
      # (F6 — such a line was previously dropped in silence). Tested against
      # the same FINDING_SHAPE_RE the in-section malformed check below uses,
      # so the two can never disagree about what counts as finding-shaped
      # (G2 — round 2).
      if [[ "$line" =~ $FINDING_SHAPE_RE ]]; then
        printf '%s:%s: finding line appears before any recognized section heading\n' "$rel" "$lineno"
        VIOLATIONS=$((VIOLATIONS + 1))
      fi
      continue
    fi

    if [[ "$line" == "_none — this angle produced no findings._" ]]; then
      HAS_NONE[$idx]=1
    elif [[ "$line" =~ $FINDING_LINE_RE ]]; then
      # A well-formed finding line: exactly one space after the `-`.
      # (G2 — round 2): checked BEFORE the loose FINDING_SHAPE_RE below,
      # so a well-formed line is never also reported as malformed.
      finding_label="${BASH_REMATCH[1]}"
      finding_text="${BASH_REMATCH[2]}"
      FINDING_COUNT[$idx]=$((FINDING_COUNT[$idx] + 1))

      if [[ "$finding_label" != "$cur" ]]; then
        printf '%s:%s: finding line label `%s` does not match its section `%s`\n' \
          "$rel" "$lineno" "$finding_label" "$cur"
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
    elif [[ "$line" =~ $FINDING_SHAPE_RE ]]; then
      # Finding-shaped (a `-`, optional whitespace, `**[`) but not
      # well-formed — e.g. more than one space after the `-` (G2 —
      # round 2). Without this branch such a line was silently dropped:
      # not counted as a finding, not flagged, invisible to every check
      # below. The same FINDING_SHAPE_RE the orphan check above tests a
      # pre-heading line against is used here for an in-section one, so
      # the two checks cannot disagree about what counts as
      # finding-shaped — a plain prose line (no `-**[` prefix at all)
      # still matches neither pattern and stays ignored, exactly as
      # ordinary body prose should.
      printf '%s:%s: finding-shaped line is malformed (want "- **[%s]** <text> — filed: <KEY>" or "- **[%s]** <text> — declined")\n' \
        "$rel" "$lineno" "$cur" "$cur"
      VIOLATIONS=$((VIOLATIONS + 1))
    fi

    # Close the body with an explicit no-op so the loop's exit status means
    # exactly one thing. bash reports a `while` with the status of the last
    # command its body ran, so without this the classification chain above
    # would decide it — harmless today, since every arm ends in a `printf` or
    # an assignment and a chain matching no arm is itself 0, but one arm ever
    # ending in a failing test or a `(( x++ ))` that evaluates to zero would
    # silently turn a finished read into "cannot read report file". Every
    # early exit from this body is a `continue`, which leaves 0 (measured on
    # bash 3.2.57 and 5.3.15), so every iteration ends at 0 and only a failed
    # redirection can make the loop non-zero.
    :
  done < "$f"
  read_rc=$?
  [[ "$read_rc" -eq 0 ]] || die "cannot read report file: $f"

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
