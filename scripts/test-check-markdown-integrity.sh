#!/usr/bin/env bash
# Assertion harness for check-markdown-integrity.py.
#
# Builds a temporary project-root-shaped directory carrying `skills/` and
# `rules/`, writes one Markdown file into it, runs the guard against that
# root, and asserts the exit code (and, where the case cares, that the
# reported output names the signal it is about). Follows
# test-check-panel-reproducers.sh's shape: a case per scenario, a helper
# that builds the sandbox, a counter, and a non-zero exit when any case
# fails.
#
# `-e` as well as `-u`/`pipefail`: without it, a failed `mktemp` in
# make_project would leave a stale path in play and the suite could keep
# going against the wrong sandbox. `expect_exit`/`expect_exit_and_names`
# bracket the guard invocation in their own `set +e`/`set -e`, matching the
# sibling harnesses, because most of these cases expect the guard to exit
# non-zero and `-e` would otherwise abort the suite on the first one.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-markdown-integrity.py"
FAILED=0

# Every case leaves one sandbox directory behind, removed on exit. An
# indexed array, not a space-separated string, for the same reason
# test-check-panel-reproducers.sh uses one: a mktemp path under TMPDIR may
# contain spaces, and word-splitting a string would leak a sandbox whose
# path split and rm -rf the fragments.
PROJECTS=()
cleanup() {
  [ "${#PROJECTS[@]}" -eq 0 ] && return 0
  for p in "${PROJECTS[@]}"; do
    chmod -R u+rwx "$p" 2>/dev/null || true
    rm -rf "$p"
  done
}
trap cleanup EXIT

make_project() {
  # $1 = relative path under the project root (e.g. skills/foo/SKILL.md
  # or rules/bar.mdc), $2 = file body. Prints the project root path.
  local rel="$1" body="$2" root
  root="$(mktemp -d "${TMPDIR:-/tmp}/check-markdown-integrity-test.XXXXXX")" || {
    printf 'make_project: mktemp failed — aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  PROJECTS+=("$root")
  mkdir -p "$(dirname "$root/$rel")"
  printf '%s\n' "$body" > "$root/$rel"
  printf '%s' "$root"
}

expect_exit() {
  # $1 = label, $2 = expected exit, $3... = command
  local label="$1" want="$2"; shift 2
  local out got
  set +e
  out="$("$@" 2>&1)"; got=$?
  set -e
  if [[ "$got" != "$want" ]]; then
    printf 'FAIL %s: expected exit %s, got %s\n%s\n' "$label" "$want" "$got" "$out"
    FAILED=1
  else
    printf 'ok: %s\n' "$label"
  fi
}

expect_exit_and_names() {
  # $1 = label, $2 = expected exit, $3 = substring the output must contain,
  # $4... = command
  local label="$1" want="$2" needle="$3"; shift 3
  local out got
  set +e
  out="$("$@" 2>&1)"; got=$?
  set -e
  if [[ "$got" != "$want" ]]; then
    printf 'FAIL %s: expected exit %s, got %s\n%s\n' "$label" "$want" "$got" "$out"
    FAILED=1
    return
  fi
  case "$out" in
    *"$needle"*) printf 'ok: %s\n' "$label" ;;
    *)
      printf 'FAIL %s: expected output to name %s, got:\n%s\n' "$label" "$needle" "$out"
      FAILED=1
      ;;
  esac
}

# ===========================================================================
# Case 1 — signal 1 (broken code span), fire: a single-backtick span whose
# intended closer is stolen by a backslash-escaped backtick earlier in the
# line, leaving a real closing backtick orphaned later on the same line.
# ===========================================================================
root="$(make_project 'skills/foo/SKILL.md' \
'This span is `foo\`bar` and now the closing backtick is orphaned.')"
expect_exit_and_names 'case 1: escaped backtick orphans the real closer — exit 1' 1 'broken code span' \
  "$GUARD" "$root"

# ===========================================================================
# Case 2 — signal 1, near-miss: double-backtick delimiters correctly wrap a
# literal single backtick. Must not fire.
# ===========================================================================
root="$(make_project 'skills/foo/SKILL.md' \
'Use ``get `foo` done`` to run it.')"
expect_exit 'case 2: double-backtick delimiter around a literal backtick — exit 0' 0 \
  "$GUARD" "$root"

# ===========================================================================
# Case 3 — signal 1, fire: an odd backtick count outside a fence — one lone
# backtick that never gets a matching delimiter of the same length.
# ===========================================================================
root="$(make_project 'skills/foo/SKILL.md' \
'This line has one stray backtick ` here and nothing closes it.')"
expect_exit_and_names 'case 3: odd backtick count outside a fence — exit 1' 1 'broken code span' \
  "$GUARD" "$root"

# ===========================================================================
# Case 4 — signal 2 (orphaned blockquote body), fire: a `> ` line follows
# un-marked prose that does not end a sentence — the blockquote's first
# line lost its own marker.
# ===========================================================================
root="$(make_project 'skills/foo/SKILL.md' \
'The following important point stands out
> and continues here without its own marker.')"
expect_exit_and_names 'case 4: unmarked prose torn from its blockquote — exit 1' 1 'orphaned blockquote body' \
  "$GUARD" "$root"

# ===========================================================================
# Case 5 — signal 2, near-miss: the same shape, but the prose line ends a
# complete sentence before the blockquote begins. Must not fire.
# ===========================================================================
root="$(make_project 'skills/foo/SKILL.md' \
'The following point stands out.
> and this blockquote begins fresh from a real point.')"
expect_exit 'case 5: blockquote after a completed sentence — exit 0' 0 \
  "$GUARD" "$root"

# ===========================================================================
# Case 6 — signal 3 (unterminated paragraph), fire: a paragraph whose last
# line ends with no terminal punctuation at all.
# ===========================================================================
root="$(make_project 'skills/foo/SKILL.md' \
'This paragraph starts fine but then just stops without')"
expect_exit_and_names 'case 6: paragraph torn mid-clause — exit 1' 1 'unterminated paragraph' \
  "$GUARD" "$root"

# ===========================================================================
# Case 7 — signal 3 (and signal 1), near-miss: the same torn-looking text
# and a lone backtick, each sitting inside a fence, a table cell and a list
# item — none of which are scanned as prose. Must not fire.
# ===========================================================================
root="$(make_project 'skills/foo/SKILL.md' \
'```text
This looks unterminated and has a lone backtick ` here too
```

| Col | text without period |
|-----|----------------------|
| a   | b                    |

- This list item also stops without terminal punctuation and has ` a backtick')"
expect_exit 'case 7: torn-looking text inside fence/table/list-item — exit 0' 0 \
  "$GUARD" "$root"

# ===========================================================================
# Case 8 — signal 4 (dangling promise), near-miss: a paragraph ending in a
# colon immediately followed by a list. The colon's promise is kept. Must
# not fire.
# ===========================================================================
root="$(make_project 'skills/foo/SKILL.md' \
'The steps are as follows:

- Step one
- Step two')"
expect_exit 'case 8: colon followed by a list — exit 0' 0 \
  "$GUARD" "$root"

# ===========================================================================
# Case 9 — signal 4, fire: a paragraph ending in a colon as the last block
# before a heading — nothing ever supplies what the colon promised.
# ===========================================================================
root="$(make_project 'skills/foo/SKILL.md' \
'The following applies here as well:

## Next heading')"
expect_exit_and_names 'case 9: colon promise with nothing following before a heading — exit 1' 1 'dangling promise' \
  "$GUARD" "$root"

# ===========================================================================
# Case 10 — signal 4, near-miss: a section's last paragraph ends with an
# ordinary "See X (`f.md`) for why ..." citation, terminated by a full
# stop, immediately before a heading. This repository ends hundreds of
# sections exactly this way, correctly — signal 4 must be structural
# (colon-and-nothing-follows) and never phrase-based, or this near-miss
# would fire on all of them with no legal way to quiet it.
# ===========================================================================
root="$(make_project 'skills/foo/SKILL.md' \
'This section explains the rationale in detail. See **Widget** (`widget.md`) for why this design was chosen.

## Next heading')"
expect_exit 'case 10: a citation ending in a full stop before a heading — exit 0' 0 \
  "$GUARD" "$root"

# ===========================================================================
# Case 11 — environment: an unreadable file inside the scanned scope is a
# refusal (exit 2), not a silent "clean".
# ===========================================================================
root="$(make_project 'skills/foo/SKILL.md' 'Some ordinary prose that ends with a period.')"
chmod 000 "$root/skills/foo/SKILL.md"
expect_exit_and_names 'case 11: an unreadable file is cannot-answer — exit 2' 2 'cannot' \
  "$GUARD" "$root"
chmod 644 "$root/skills/foo/SKILL.md"

# ===========================================================================
# Case 12 — environment: a scope root that does not exist is a refusal
# (exit 2), never treated as an empty, clean tree.
# ===========================================================================
expect_exit 'case 12: a nonexistent scope root — exit 2' 2 \
  "$GUARD" "/nonexistent-check-markdown-integrity-root-$$"

# ===========================================================================
# Cases 13-16 — near-misses added after calibrating this guard against this
# repository's own skills/ and rules/ tree (required by the change that
# introduced it). The first calibration pass found four real shapes this
# repository uses correctly and constantly, each of which the original
# four-signal rule misread as damage; each is now a structural exemption
# in the guard itself, and pinned here so a future edit cannot reintroduce
# the false-positive class without a failing test.
# ===========================================================================

# Case 13 — YAML frontmatter: every skill and rule file in this repository
# opens with a `---`-delimited frontmatter block. Its own closing `---`
# line must never read as an unterminated paragraph.
root="$(make_project 'skills/foo/SKILL.md' \
'---
name: foo
description: bar
---

# Title

Some closing prose that ends properly.')"
expect_exit 'case 13: YAML frontmatter is not scanned as prose — exit 0' 0 \
  "$GUARD" "$root"

# Case 14 — a paragraph ending with no terminal punctuation directly
# introducing a fenced command block ("...run", blank line, ```bash ...```)
# is this repository's own established style throughout skills/myflow-do/
# SKILL.md, not a torn sentence — the fence supplies what the paragraph
# led into.
root="$(make_project 'skills/foo/SKILL.md' \
'Run the following command

```bash
echo hi
```')"
expect_exit 'case 14: a paragraph introducing a fenced command — exit 0' 0 \
  "$GUARD" "$root"

# Case 15 — a bold pseudo-heading directly above the list it introduces,
# with no blank line between them and no terminal punctuation of its own
# ("**Explore the problem space**" followed immediately by "- Ask..."),
# is this repository's own established style in skills/myflow-research/
# SKILL.md, not an orphaned marker.
root="$(make_project 'skills/foo/SKILL.md' \
'**Explore the problem space**
- Ask clarifying questions
- Challenge assumptions')"
expect_exit 'case 15: a bold label directly above the list it introduces — exit 0' 0 \
  "$GUARD" "$root"

# Case 16 — a CommonMark indented code block (four-space indent, no fence
# delimiters) is code, not prose, even though it carries no terminal
# punctuation of its own — skills/myflow-contracts/plan-provenance.md
# uses exactly this shape for its own worked examples.
root="$(make_project 'skills/foo/SKILL.md' \
'This text introduces an indented example below.

    literal example line with no terminal punctuation')"
expect_exit 'case 16: an indented code block is not scanned as prose — exit 0' 0 \
  "$GUARD" "$root"

# ===========================================================================
# Cases 17-18 — the signal-3 exemption discriminator (F1 fix): a paragraph
# before a fence/list/table is exempt from signal 3 only when its own last
# word could plausibly end a clause. A paragraph that trails off on a
# closed-class word (a preposition, conjunction, article, auxiliary) is
# reported even though a fence follows it, because the fence is unrelated
# to the torn sentence, not a supply for it.
# ===========================================================================

# Case 17 — signal 3, fire: a paragraph genuinely torn mid-clause — its
# last word ("without") is a dangling preposition — sitting in front of an
# unrelated fence. The fence does not complete this paragraph; the guard
# must not read the mere adjacency as "introduce, then supply" and must
# still report the tear.
root="$(make_project 'skills/foo/SKILL.md' \
'This paragraph starts fine but then just stops without

```bash
echo x
```')"
expect_exit_and_names 'case 17: torn paragraph before an unrelated fence — exit 1' 1 'unterminated paragraph' \
  "$GUARD" "$root"

# Case 18 — signal 3, near-miss: case 14's introducing-before-fence shape
# must keep passing under the narrowed exemption — its last word
# ("command") is an ordinary noun, not a closed-class dangling word, so the
# paragraph reads as a complete clause and the fence genuinely supplies
# what it introduced.
root="$(make_project 'skills/foo/SKILL.md' \
'Run the following command

```bash
echo hi
```')"
expect_exit 'case 18: an introducing paragraph before its own fence still passes — exit 0' 0 \
  "$GUARD" "$root"

# ===========================================================================
# Cases 19-20 — the signal-3 dangling-word discriminator, second pass (F4
# fix): a paragraph that ends on a dangling word in front of a fence/list/
# table is exempt after all when the sentence *resumes* past that block —
# the "spans an embedded block" shape this repository uses for real
# (skills/myflow-contracts/plan-provenance.md:211, :246;
# skills/myflow-do/SKILL.md:133), where a sentence deliberately breaks
# around a worked example and picks back up afterward. Case 17 above still
# pins the case where nothing follows the fence to resume it; these two pin
# the case where something does, and the case where what follows starts a
# new sentence instead of resuming the old one.
# ===========================================================================

# Case 19 — signal 3, near-miss: the paragraph's last word ("and") is a
# dangling conjunction, same shape as case 17, but this time the block past
# the fence resumes the sentence in lower case ("resumes right where..."),
# so the fence is a genuine mid-sentence interruption, not proof of a tear.
root="$(make_project 'skills/foo/SKILL.md' \
'This paragraph spans an embedded example and

```bash
echo x
```

resumes right where it left off.')"
expect_exit 'case 19: dangling word before a fence, resumed afterward — exit 0' 0 \
  "$GUARD" "$root"

# Case 20 — signal 3, fire: same dangling last word ("and") and the same
# fence, but the block past the fence opens with a capital letter — a new
# sentence, not a resumption. Having *something* after the fence is not
# enough on its own; the guard must still report the tear.
root="$(make_project 'skills/foo/SKILL.md' \
'This paragraph looks like it tears off after the fence and

```bash
echo x
```

New sentence begins here instead of resuming.')"
expect_exit_and_names 'case 20: dangling word before a fence, followed by a new sentence — exit 1' 1 'unterminated paragraph' \
  "$GUARD" "$root"

# ===========================================================================
# Case 21 — signal 2 near-miss (F3 fix): a heading immediately followed by
# a blockquote. A heading is marked, not un-marked prose, and heading text
# routinely carries no terminal punctuation of its own — this must not
# read as an orphaned blockquote marker.
# ===========================================================================
root="$(make_project 'skills/foo/SKILL.md' \
'## A heading with no terminal punctuation
> a blockquote that begins right after a heading')"
expect_exit 'case 21: a blockquote immediately after a heading — exit 0' 0 \
  "$GUARD" "$root"

# ===========================================================================
# Cases 22-24 — signal 2 near-misses (F1 fix): a blockquote directly
# preceded by a fence, a list item, or a table must not read as orphaned.
# find_orphaned_blockquotes previously excluded only "blockquote" and
# "heading" from the predecessor-kind test, even though the guard already
# defines COMPLETING_KINDS ("fence", "list_item", "table") for exactly this
# purpose elsewhere (the signal-3 "introduce, then supply" exemption). A
# tilde-fence, a tight list item, or a table sitting directly above a `> `
# block is valid Markdown and must not be reported.
# ===========================================================================

# Case 22 — a fenced block directly above a blockquote.
root="$(make_project 'skills/foo/SKILL.md' \
'~~~text
echo hi
~~~
> a blockquote that begins right after a fence')"
expect_exit 'case 22: a blockquote immediately after a fence — exit 0' 0 \
  "$GUARD" "$root"

# Case 23 — a list item directly above a blockquote.
root="$(make_project 'skills/foo/SKILL.md' \
'- a list item with no terminal punctuation
> a blockquote that begins right after a list item')"
expect_exit 'case 23: a blockquote immediately after a list item — exit 0' 0 \
  "$GUARD" "$root"

# Case 24 — a table directly above a blockquote.
root="$(make_project 'skills/foo/SKILL.md' \
'| Col | Text |
|-----|------|
> a blockquote that begins right after a table')"
expect_exit 'case 24: a blockquote immediately after a table — exit 0' 0 \
  "$GUARD" "$root"

# ===========================================================================
# Case 25 — signal 2, fire (F1 regression guard): an actually-orphaned
# blockquote after un-marked prose — none of fence, list_item, table, or
# heading — must still be reported. The newly-excluded kinds must not widen
# the exemption past what COMPLETING_KINDS actually names.
# ===========================================================================
root="$(make_project 'skills/foo/SKILL.md' \
'This important point stands out but was never marked
> and continues here without its own marker.')"
expect_exit_and_names 'case 25: unmarked prose (not a completing kind) torn from its blockquote — exit 1' 1 'orphaned blockquote body' \
  "$GUARD" "$root"

if [ "$FAILED" -ne 0 ]; then
  printf 'test-check-markdown-integrity: one or more cases failed\n' >&2
  exit 1
fi
printf 'test-check-markdown-integrity: all 25 cases pass\n'
