#!/usr/bin/env bash
# check-references.sh — fail when a cross-referenced section no longer exists
# in the file it is referenced from.
#
# Rule: for every backticked .md/.mdc path that resolves to a real file, if the
# line ASSOCIATES one or more **bold tokens** with that path, at least one of
# those associated tokens must match a `#`, `##`, `###`, or `####` heading in
# that file.
#
# "Associated" is the load-bearing word. The rule originally checked every bold
# token on the line against every path on the line, which is not what a
# cross-reference looks like: a line that says "**Never** commit — see
# `rules/x.mdc`" was required to have a section called "Never". That produced 28
# failures on this repo's own tree, none of them a genuinely stale reference,
# and the suppression marker was then the only way to silence them — which in
# turn switched off the real checks sharing those lines. So a bold token counts
# only when it actually sits next to the path, in one of the shapes this repo
# writes references in:
#
#   **Section** (`path`)          — parenthesised
#   **Section** in `path`         — short connector
#   see/per/under **Section** … `path`
#   `path` — sections **A**, **B** — path first, tokens after
#
# See is_associated below for the mechanical test. Both directions of error are
# lenient by construction (an unassociated token neither satisfies nor fails a
# path), so the guard's power comes from recognising the reference shapes, not
# from blanket coverage.
#
# This is the companion guard to check-vocabulary.sh. That one greps for
# known-retired literals; this one catches a section that MOVED, which no
# literal list can know about in advance.
#
# Takes no arguments: the scan set lives here, in one place, so no call site
# can narrow it. Lines carrying `refs-guard:allow` are skipped — use it for a
# line whose bold text is emphasis rather than a section name.
set -euo pipefail

# REPO_ROOT is always resolved from the guard script's own location, exactly
# like check-vocabulary.sh's SCRIPT_DIR/REPO_ROOT — this is what makes the
# guard argument-free and self-scoped from any cwd (see the delta-spec
# scenario "The script is argument-free and self-scoped"): running it from
# /tmp, or from a subdirectory of the repo, still scans the real repo instead
# of silently scanning nothing and reporting a false "clean".
#
# CHECK_REFERENCES_ROOT is an explicit, opt-in override honored only when set
# (mirrors the MYFLOW_STATE_FILE idiom used elsewhere in this plan). It exists
# solely so the companion harness (test-check-references.sh) can point the
# guard at a sandboxed fixture tree under /tmp without touching this repo —
# never set it for a normal invocation.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -n "${CHECK_REFERENCES_ROOT:-}" ]; then
  REPO_ROOT="$CHECK_REFERENCES_ROOT"
elif [ "${CHECK_REFERENCES_ROOT+set}" = "set" ]; then
  printf 'CHECK_REFERENCES_ROOT is set but empty\n' >&2
  exit 2
else
  REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
fi

# A root that is not a directory would make every target miss and the guard
# report a clean run over zero files — a false "all clear", which is the one
# outcome a guard must never produce.
if [ ! -d "$REPO_ROOT" ]; then
  printf 'not a directory: %s\n' "$REPO_ROOT" >&2
  exit 2
fi
REPO_ROOT_PHYS="$(cd "$REPO_ROOT" && pwd -P)"
# The lexical root, used by the containment test below (lexical_norm is defined
# further down; REPO_ROOT_NORM is filled in right after it).
REPO_ROOT_ABS="$(cd "$REPO_ROOT" && pwd)"

# coverage_record / coverage_declare / coverage_report / coverage_verdict —
# per-file coverage reporting and the declared-vs-undeclared-zero decision,
# owned once in lib/coverage.sh rather than reinvented here. See that file's
# header for why (KAN-197) and check-guard-symlinks.sh for the pattern this
# guard follows.
source "$SCRIPT_DIR/lib/coverage.sh"

DEFAULT_TARGETS=(
  "rules"
  "skills"
  "commands"
  "commands-claude"
  "README.md"
  "AGENTS.md"
  "CLAUDE.md"
)

FAILURES=0

# is_fence_line <line> — true when the line is a fence delimiter (``` or ~~~),
# the single shared predicate for fence-toggle tracking. Both check_file (which
# must preserve line numbers while scanning the REFERENCING file) and
# strip_fenced_lines (used by headings_of to scan the REFERENCED file) call
# this one function, so "a line that starts a/ends a fence" is defined in
# exactly one place and the two scans can never quietly diverge on it.
is_fence_line() {
  case "$1" in
    '```'*|'~~~'*) return 0 ;;
    *) return 1 ;;
  esac
}

# strip_fenced_lines <file> — the file's lines with every fenced block (```
# or ~~~ delimited, inclusive of the delimiters) removed. Line numbers are not
# preserved — callers that need line numbers (check_file) must not use this;
# it exists for headings_of, which only needs the surviving text.
strip_fenced_lines() {
  local file="$1" in_fence=0 line
  while IFS= read -r line || [ -n "$line" ]; do
    if is_fence_line "$line"; then
      in_fence=$((1 - in_fence))
      continue
    fi
    [ "$in_fence" -eq 1 ] && continue
    printf '%s\n' "$line"
  done < "$file"
}

# headings_of <file> — every #, ##, ###, or #### heading OUTSIDE a fenced code
# block, normalized: markers stripped, backticks and emphasis removed,
# trimmed, lowercased. Fence-aware so a "# comment" inside a ```bash example
# is never mistaken for a real section heading — see strip_fenced_lines.
headings_of() {
  strip_fenced_lines "$1" \
    | sed -n 's/^#\{1,4\} //p' \
    | tr -d '`*' \
    | sed 's/^[[:space:]]*//; s/[[:space:]]*$//' \
    | tr '[:upper:]' '[:lower:]'
}

normalize_token() {
  printf '%s' "$1" | tr -d '`*' \
    | sed 's/^[[:space:]]*//; s/[[:space:]]*$//' \
    | tr '[:upper:]' '[:lower:]'
}

# mask_code_spans <line> — every "**" that occurs INSIDE a backtick-delimited
# code span, on this physical line, is replaced with "@@" (two non-asterisk
# placeholder characters). Everything else — the backticks themselves, any
# other text inside the code span, and any "**" OUTSIDE a code span — is left
# byte-for-byte unchanged.
#
# Why this exists: a line like "See **State file** and code `x**y` in
# `rules/contract.mdc`." carries a LITERAL "**" inside inline code (a glob,
# an operator, whatever) that is not a bold delimiter at all. Counting or
# pairing "**" without excluding code spans first lets that literal pair
# desync the real "**State file**" bold span next to it. Masking only the
# in-code occurrences (not the whole span) is deliberate: a bold span that
# itself WRAPS a code span, e.g. "**`prUrl: null`**", must still extract
# as "`prUrl: null`" untouched — replacing the whole code span would lose
# real content the same bug this fixes is trying to preserve.
#
# Scoped to single-backtick spans (the shape every real reference in this
# repo and the one in this fix's own test cases use); a run of 2+ backticks
# used to escape a literal backtick inside code is not specially handled.
mask_code_spans() {
  awk '
    {
      line = $0
      n = length(line)
      out = ""
      in_code = 0
      i = 1
      while (i <= n) {
        c = substr(line, i, 1)
        if (c == "`") {
          in_code = !in_code
          out = out c
          i++
          continue
        }
        if (in_code && c == "*" && substr(line, i, 2) == "**") {
          out = out "@@"
          i += 2
          continue
        }
        out = out c
        i++
      }
      print out
    }
  ' <<< "$1"
}

# normalized_line <line> — the line with in-code "**" masked and a leading
# orphaned closing "**" stripped, ready for bold/path span scanning.
#
# A naive `grep -oE '\*\*[^*]+\*\*'` mis-pairs markers when the line opens
# with an ORPHANED closing `**` — the tail end of a bold span that started on
# the PREVIOUS physical line (prose soft-wrapped mid-span). E.g. on
# "sync** in **Jira integration** (...)", a greedy scan pairs the first `**`
# (closing the orphan) with the SECOND `**` (opening the real span), extracting
# "** in **" as the bold content and never seeing "**Jira integration**" at
# all — silently dropping a live reference instead of flagging it.
#
# `**` markers on a line with no cross-line continuation and no in-code "**"
# always occur in complete pairs, i.e. an EVEN count. An odd count means the
# first `**` is such an orphan: strip exactly that first occurrence before
# pairing, so the remaining markers on the line pair up correctly. Counting
# and pairing both run on the CODE-MASKED line (mask_code_spans above) so a
# literal "**" inside inline code never contributes to that count in the
# first place — see mask_code_spans for why that matters.
normalized_line() {
  local line="$1" masked count
  masked="$(mask_code_spans "$line")"
  count="$(printf '%s' "$masked" | grep -o '\*\*' | wc -l | tr -d '[:space:]' || true)"
  [ -n "$count" ] || count=0
  if [ $((count % 2)) -eq 1 ]; then
    masked="$(printf '%s' "$masked" | sed 's/\*\*//')"
  fi
  printf '%s\n' "$masked"
}

# associations_of <normalized-line> — every (path, bold token) pair the
# adjacency rule associates on this line, as "path<TAB>token".
#
# Adjacency test, applied to the text BETWEEN a bold span and a path span (in
# either order): the gap must be short (<= 60 characters and, ignoring ordinary
# inline code, <= 3 words), must not contain another .md/.mdc reference or
# another bold marker, and must not contain a "." — a sentence boundary between
# the two means they belong to different statements, not to one reference. That
# is what separates "**Section** in `path`" from "**Never** commit. See `path`".
associations_of() {
  awk '
    function is_associated(gap,   clean, n, w, i, c) {
      if (length(gap) > 60) return 0
      # Another .md/.mdc reference between them: they belong to different pairs.
      if (gap ~ /`[^`]*\.mdc?`/) return 0
      # Ordinary inline code between them (`jq`, `--target`) is connective
      # tissue, not a separator — drop it before judging the prose.
      clean = gap
      gsub(/`[^`]*`/, " ", clean)
      if (clean ~ /\*\*/) return 0
      if (clean ~ /\./) return 0
      n = split(clean, w, /[^A-Za-z]+/)
      c = 0
      for (i = 1; i <= n; i++) if (w[i] != "") c++
      return (c <= 3)
    }
    # A bold span is a candidate SECTION NAME only if it reads like one. Bold is
    # used for emphasis far more often than for section names in this repo
    # ("**only** in projects whose `.myflow/project.md`", "**absolute** path of
    # `engineering-principles.md`"), and an emphasis word next to a filename is
    # not a cross-reference. Every heading in the referenced files starts with a
    # capital, carries no "." and no trailing ":", so those three cheap tests
    # separate the two uses.
    #
    # Cost, stated plainly: a genuine reference written with a lowercase or
    # punctuated bold token is not checked. Heading MATCHING stays
    # case-insensitive, so this only ever decides whether a path is examined —
    # it never turns a match into a miss.
    function looks_like_section(t,   s) {
      s = t
      gsub(/`/, "", s)
      sub(/^[ \t]+/, "", s)
      sub(/[ \t]+$/, "", s)
      if (s == "") return 0
      if (s ~ /\//) return 0
      if (s ~ /\./) return 0
      if (s ~ /:$/) return 0
      return (s ~ /^[A-Z]/)
    }
    {
      line = $0
      nb = 0; np = 0
      i = 1
      while ((s = index(substr(line, i), "**")) > 0) {
        s = i + s - 1
        e = index(substr(line, s + 2), "**")
        if (e == 0) break
        e = s + 2 + e - 1
        txt = substr(line, s + 2, e - s - 2)
        if (txt != "" && txt !~ /\*/ && looks_like_section(txt)) {
          nb++; bs[nb] = s; be[nb] = e + 1; bt[nb] = txt
        }
        i = e + 2
      }
      i = 1
      while ((s = index(substr(line, i), "`")) > 0) {
        s = i + s - 1
        e = index(substr(line, s + 1), "`")
        if (e == 0) break
        e = s + 1 + e - 1
        txt = substr(line, s + 1, e - s - 1)
        if (txt ~ /\.mdc?$/) { np++; ps[np] = s; pe[np] = e; pt[np] = txt }
        i = e + 1
      }
      # A bold token belongs to at most ONE path: the nearest it is associated
      # with. Without this, a line naming .myflow/project.md and then saying
      # "(see **Project configuration** in skills/.../project-configuration.md)"
      # would demand a "Project configuration" heading in BOTH files, and fail on
      # the one the token was never about.
      for (b = 1; b <= nb; b++) {
        best = 0; bestlen = -1
        for (p = 1; p <= np; p++) {
          if (be[b] < ps[p]) gap = substr(line, be[b] + 1, ps[p] - be[b] - 1)
          else if (pe[p] < bs[b]) gap = substr(line, pe[p] + 1, bs[b] - pe[p] - 1)
          else continue
          if (!is_associated(gap)) continue
          if (bestlen < 0 || length(gap) < bestlen) { best = p; bestlen = length(gap) }
        }
        if (best > 0) print pt[best] "\t" bt[b]
      }
    }
  ' <<< "$1"
}

# lexical_norm <path> — the path with `.` and `..` resolved TEXTUALLY, without
# touching the filesystem. Purely lexical on purpose: containment must be
# decided from the shape of the reference alone, so the guard's verdict cannot
# depend on whether the out-of-tree target happens to exist on this machine.
lexical_norm() {
  local p="$1" part joined
  case "$p" in /*) ;; *) p="$PWD/$p" ;; esac
  local -a out=()
  local IFS=/
  for part in $p; do
    case "$part" in
      ''|.) ;;
      ..) [ "${#out[@]}" -gt 0 ] && unset "out[$(( ${#out[@]} - 1 ))]" ;;
      *) out+=("$part") ;;
    esac
  done
  joined="${out[*]:-}"
  printf '/%s\n' "$joined"
}

REPO_ROOT_NORM="$(lexical_norm "$REPO_ROOT_ABS")"

# contained <candidate> — print the candidate normalized to an absolute path,
# but only when it lies inside REPO_ROOT. Repository Markdown is
# attacker-influenceable (any pull request can add a line), and the path in a
# backtick span is read from disk — without this, `` `../../outside/target.md` ``
# is opened, which is a heading-existence oracle over the whole filesystem and a
# stall on a large file.
#
# Two tests, in this order:
#
#  1. LEXICAL, always. `..` is resolved textually and the result must sit under
#     the root. This runs before any `stat`, so a crafted reference is refused
#     identically whether or not the target exists — otherwise the same
#     repository would warn on a laptop and pass in CI, and the attacker-probe
#     case (a reference to a path that may or may not be there) would be the
#     quiet one.
#  2. PHYSICAL, when the target exists. `pwd -P` resolves symlinks, so a link
#     inside the tree cannot point the guard out of it.
contained() {
  local candidate="$1" norm dir base phys
  norm="$(lexical_norm "$candidate")"
  case "$norm" in
    "$REPO_ROOT_NORM"/*) ;;
    *) return 1 ;;
  esac
  dir="$(dirname "$norm")"
  base="$(basename "$norm")"
  if [ -d "$dir" ]; then
    phys="$(cd "$dir" && pwd -P)/$base"
    case "$phys" in
      "$REPO_ROOT_PHYS"/*) ;;
      *) return 1 ;;
    esac
  fi
  printf '%s\n' "$norm"
}

check_file() {
  local file="$1" lineno=0 in_fence=0 line checked=0
  while IFS= read -r line || [ -n "$line" ]; do
    lineno=$((lineno + 1))

    if is_fence_line "$line"; then
      in_fence=$((1 - in_fence))
      continue
    fi
    [ "$in_fence" -eq 1 ] && continue
    case "$line" in *refs-guard:allow*) continue ;; esac
    case "$line" in *'**'*) ;; *) continue ;; esac
    case "$line" in *'`'*) ;; *) continue ;; esac

    local pairs
    pairs="$(associations_of "$(normalized_line "$line")")"
    [ -n "$pairs" ] || continue

    # The set of paths this line associates at least one bold token with.
    local paths path
    paths="$(printf '%s\n' "$pairs" | cut -f1 | sort -u)"

    while IFS= read -r path; do
      [ -n "$path" ] || continue

      # An absolute path is its own only candidate; a relative one is tried
      # against the repository root and against the referring file's directory.
      local candidates=()
      case "$path" in
        /*) candidates=("$path") ;;
        *)  candidates=("$REPO_ROOT/$path" "$(dirname "$file")/$path") ;;
      esac

      local resolved="" escaped=0 candidate safe
      for candidate in "${candidates[@]}"; do
        if safe="$(contained "$candidate")"; then
          if [ -f "$safe" ]; then
            resolved="$safe"
            break
          fi
        else
          escaped=1
        fi
      done
      # A reference pointing outside the repository is a defect in the referring
      # file, not a note in passing: this guard's contract is that a clean exit
      # means every reference was checked, and an unreadable one was not.
      if [ -z "$resolved" ] && [ "$escaped" -eq 1 ]; then
        printf '%s:%d: reference %s resolves outside the repository root; not read\n' \
          "${file#"$REPO_ROOT"/}" "$lineno" "$path"
        FAILURES=$((FAILURES + 1))
        continue
      fi
      # A path that resolves to nothing is out of scope: templated paths like
      # openspec/changes/<name>/tasks.md are legitimate and must not fail the guard.
      [ -n "$resolved" ] || continue

      local heads matched token
      heads="$(headings_of "$resolved")"
      [ -n "$heads" ] || continue
      # This is a CHECKED reference: a bold token associated with a path that
      # resolved to a real, headed file, and was actually compared against
      # that file's headings — see scripts/lib/coverage.sh's header for why
      # this count, not the mere fact that check_file ran, is what "coverage"
      # means here. A file with none of these is a file this guard read but
      # verified nothing in, indistinguishable from the outside from a rule
      # that silently stopped checking it.
      checked=$((checked + 1))
      matched=0
      while IFS= read -r token; do
        [ -n "$token" ] || continue
        token="$(normalize_token "$token")"
        # The path itself is often bolded; that is a file reference, not a
        # section name, so it never counts as a match.
        case "$token" in */*) continue ;; esac
        if printf '%s\n' "$heads" | grep -qxF "$token"; then
          matched=1
          break
        fi
      done < <(printf '%s\n' "$pairs" | awk -F'\t' -v p="$path" '$1 == p { print $2 }')
      if [ "$matched" -eq 0 ]; then
        printf '%s:%d: no bold token resolves to a heading in %s\n' \
          "${file#"$REPO_ROOT"/}" "$lineno" "$path"
        FAILURES=$((FAILURES + 1))
      fi
    done <<EOF
$paths
EOF
  done < "$file"
  # KAN-197 F8: the return is checked rather than ignored, matching the audit
  # this task also applied to check-vocabulary.sh and check-stage-mark-calls.sh
  # (the two guards where an unchecked failure here would have been silently
  # swallowed under `set -uo pipefail` with no `-e`). This file already runs
  # under `set -euo pipefail`, so a failure would abort regardless, but an
  # explicit check names the real cause instead of relying on -e's own,
  # context-free abort.
  if ! coverage_record "${file#"$REPO_ROOT"/}" "$checked"; then
    printf 'check-references: coverage_record failed for %s (see stderr above)\n' "${file#"$REPO_ROOT"/}" >&2
    exit 2
  fi
}

# EXPECTED-ZERO FILES — established by running this guard's own association
# and resolution logic against the real tree (2026-08-18, at df9d5dd), never
# guessed: each of these carries no `**Bold**`-adjacent `.md`/`.mdc` path in
# any of the shapes is_associated recognizes, so this guard genuinely
# verifies nothing inside it. Declared here, once, rather than inferred from
# the tree — inferring it would restate the very assumption a silently
# uncovered file already encodes, which is exactly the case this exists to
# fail instead of pass.
#
# KAN-197 F2: batched by CATEGORY rather than one flat list sharing a single
# reason string. The panel's Primary slot found the original single reason
# attested only to HOW the zero was measured ("ran the association logic
# against this file"), never to WHY it is legitimate BY DESIGN — weaker than
# this requirement's own words, "legitimately checks nothing". Each category
# below states the actual shape that makes its members' bold/path pairs never
# associate, verified per category rather than guessed:
#
#   command-dispatch stub — every path it cites sits INSIDE the same bold
#   span as the verb citing it (e.g. "**load `skills/.../pipeline.md`
#   first**"); looks_like_section rejects any bold span containing "/", so no
#   candidate section name is ever formed next to the path.
#
#   rule file — a path citation sits in a Markdown table cell (rules/agent-
#   baseline.md's rule table) separated from any bold text by far more than
#   the adjacency window, or the file cites no path in a bold-adjacent shape
#   at all.
#
#   contract/index doc — cites other files as a plain parenthetical backtick
#   path or a `[label](path)` Markdown link (skills/myflow-contracts/SKILL.md's
#   own table), never as a bold token adjacent to the path.
#
#   reviewer-prompt file — deliberately self-contained; three of the five
#   cite no .md/.mdc path anywhere, and the two that do (engineering-
#   principles.md, principles-reviewer-prompt.md) never pair the citation
#   with an adjacent bold section name.
#
#   rationale/exploration doc — prose-only; any path citation sits inside the
#   same bold span as its citing verb (the command-dispatch-stub shape), or
#   with no bold nearby at all.
EXPECTED_ZERO_COMMAND_DISPATCH_STUBS=(
  "commands-claude/myflow-do.md"
  "commands-claude/myflow-fast.md"
  "commands-claude/myflow-start.md"
  "commands-claude/myflow-status.md"
  "commands/myflow-do.md"
  "commands/myflow-fast.md"
  "commands/myflow-start.md"
  "commands/myflow-status.md"
  "commands/opsx-explore.md"
)
EXPECTED_ZERO_COMMAND_DISPATCH_REASON="command-dispatch stub — every path it cites sits inside the SAME bold span as the verb citing it (e.g. \"**load \`path\` first**\"); looks_like_section rejects any bold span containing '/', so no candidate section name ever forms adjacent to the path"

EXPECTED_ZERO_RULE_FILES=(
  "rules/agent-baseline.md"
  "rules/be-brief.mdc"
  "rules/build-the-simplest-thing.mdc"
  "rules/context7.mdc"
  "rules/dependency-versions.mdc"
  "rules/design-mockups-are-specs.mdc"
  "rules/dispatch-carries-the-baseline.mdc"
  "rules/lint-fix-priority.mdc"
  "rules/never-touch-production.mdc"
  "rules/no-direct-pushes-to-main.mdc"
)
EXPECTED_ZERO_RULE_FILES_REASON="rule file — its own path citations (where present) sit in a Markdown table cell or plain prose, separated from any bold text by more than the adjacency window this guard's is_associated allows, or cite no path in a bold-adjacent shape at all"

EXPECTED_ZERO_CONTRACT_DOCS=(
  "skills/myflow-contracts/build-green.md"
  "skills/myflow-contracts/operator-prompts.md"
  "skills/myflow-contracts/SKILL.md"
)
EXPECTED_ZERO_CONTRACT_DOCS_REASON="contract/index doc — cites other files as a plain parenthetical backtick path or a [label](path) Markdown link, never as a bold token adjacent to the path"

EXPECTED_ZERO_REVIEWER_PROMPTS=(
  "skills/myflow-do/adversarial-reviewer-prompt.md"
  "skills/myflow-do/bug-hunter-reviewer-prompt.md"
  "skills/myflow-do/engineering-principles.md"
  "skills/myflow-do/principles-reviewer-prompt.md"
  "skills/myflow-do/security-reviewer-prompt.md"
)
EXPECTED_ZERO_REVIEWER_PROMPTS_REASON="reviewer-prompt file, deliberately self-contained — most cite no .md/.mdc path anywhere, and the rest never pair a citation with an adjacent bold section name"

EXPECTED_ZERO_RATIONALE_DOCS=(
  "skills/myflow-fast/SKILL-rationale.md"
  "skills/openspec-explore/SKILL.md"
)
EXPECTED_ZERO_RATIONALE_DOCS_REASON="rationale/exploration doc, prose-only — any path citation sits inside the same bold span as its citing verb, or with no bold nearby at all"

# declare_category <reason> <file...> — declares every <file> with <reason>,
# but ONLY when <file> exists under the CURRENT REPO_ROOT (KAN-197 F3
# compatibility: REPO_ROOT is this guard's own real location by default, or a
# sandboxed CHECK_REFERENCES_ROOT fixture under the companion test harness —
# see that variable's own header comment above. This guard's declared lists
# are this repository's own real paths; a sandboxed fixture scans a
# different, smaller tree where most of them do not exist at all. Declaring
# a path that is not part of the CURRENT run's corpus would make it a
# KAN-197 F3 "declared but never recorded" violation for every such fixture
# — not a real staleness, just a scope mismatch. Existence-gating keeps F3's
# protection meaningful for this guard's real, default run (every path below
# genuinely exists there today) without that false-positive noise.
declare_category() {
  local reason="$1"; shift
  local f
  for f in "$@"; do
    [ -f "$REPO_ROOT/$f" ] || continue
    if ! coverage_declare "$f" "$reason"; then
      printf 'check-references: coverage_declare failed for %s (see stderr above)\n' "$f" >&2
      exit 2
    fi
  done
}

declare_expected_zeros() {
  declare_category "$EXPECTED_ZERO_COMMAND_DISPATCH_REASON" "${EXPECTED_ZERO_COMMAND_DISPATCH_STUBS[@]:-}"
  declare_category "$EXPECTED_ZERO_RULE_FILES_REASON" "${EXPECTED_ZERO_RULE_FILES[@]:-}"
  declare_category "$EXPECTED_ZERO_CONTRACT_DOCS_REASON" "${EXPECTED_ZERO_CONTRACT_DOCS[@]:-}"
  declare_category "$EXPECTED_ZERO_REVIEWER_PROMPTS_REASON" "${EXPECTED_ZERO_REVIEWER_PROMPTS[@]:-}"
  declare_category "$EXPECTED_ZERO_RATIONALE_DOCS_REASON" "${EXPECTED_ZERO_RATIONALE_DOCS[@]:-}"
}

main() {
  local target scanned=0
  declare_expected_zeros
  for target in "${DEFAULT_TARGETS[@]}"; do
    local full="$REPO_ROOT/$target"
    [ -e "$full" ] || continue
    if [ -d "$full" ]; then
      while IFS= read -r f; do
        check_file "$f"
        scanned=$((scanned + 1))
      done < <(find "$full" -type f \( -name '*.md' -o -name '*.mdc' \) | sort)
    else
      check_file "$full"
      scanned=$((scanned + 1))
    fi
  done

  if [ "$scanned" -eq 0 ]; then
    printf 'no Markdown files found under %s — refusing to report a clean run\n' \
      "$REPO_ROOT" >&2
    exit 2
  fi

  # COVERAGE — per-file count of what this guard actually verified, via
  # scripts/lib/coverage.sh. A file whose coverage is zero and is not in
  # EXPECTED_ZERO_FILES above is folded into the ordinary stale-reference
  # violations rather than a separate exit status, exactly as
  # check-guard-symlinks.sh does it.
  local coverage_verdict_out
  if ! coverage_verdict_out="$(coverage_verdict)"; then
    while IFS= read -r cvline; do
      [ -n "$cvline" ] || continue
      printf '%s:0: %s\n' "${cvline%%:*}" "${cvline#*: }"
      FAILURES=$((FAILURES + 1))
    done <<<"$coverage_verdict_out"
  fi

  if [ "$FAILURES" -ne 0 ]; then
    printf '\n%d stale reference(s) found.\n' "$FAILURES" >&2
    printf 'Fix the reference, or mark the line refs-guard:allow if the bold text is emphasis.\n' >&2
    exit 1
  fi
  printf 'check-references: all referenced sections resolve\n'
  local frag
  frag="$(coverage_report)"
  [ -n "$frag" ] && printf '  %s\n' "$frag"
}

main "$@"
