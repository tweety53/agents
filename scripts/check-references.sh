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
if [ -n "${CHECK_REFERENCES_ROOT:-}" ]; then
  REPO_ROOT="$CHECK_REFERENCES_ROOT"
elif [ "${CHECK_REFERENCES_ROOT+set}" = "set" ]; then
  printf 'CHECK_REFERENCES_ROOT is set but empty\n' >&2
  exit 2
else
  REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
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
# itself WRAPS a code span, e.g. "**`fastPath: true`**", must still extract  # vocab-guard:allow
# as "`fastPath: true`" untouched — replacing the whole code span would  # vocab-guard:allow lose
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
  local file="$1" lineno=0 in_fence=0 line
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
      # docs/manual-test/<name>.md are legitimate and must not fail the guard.
      [ -n "$resolved" ] || continue

      local heads matched token
      heads="$(headings_of "$resolved")"
      [ -n "$heads" ] || continue
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
}

main() {
  local target scanned=0
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

  if [ "$FAILURES" -ne 0 ]; then
    printf '\n%d stale reference(s) found.\n' "$FAILURES" >&2
    printf 'Fix the reference, or mark the line refs-guard:allow if the bold text is emphasis.\n' >&2
    exit 1
  fi
  printf 'check-references: all referenced sections resolve\n'
}

main "$@"
