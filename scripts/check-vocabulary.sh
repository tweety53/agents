#!/usr/bin/env bash
# check-vocabulary.sh — drift guard for retired myflow vocabulary.
#
# Usage: scripts/check-vocabulary.sh [path ...]               # dirs or files
#
# With no arguments it scans the project's own documentation and command trees, resolved
# relative to the repo root — see DEFAULT_TARGETS below. That default is the ONE place the
# invocation lives: the rule file and the review skill say "run the script", they do not
# each carry a copy of the path list that can drift from this one.
#
# Scans the given directories and files for retired pipeline-stage vocabulary and retired
# review-panel vocabulary. Every hit is reported as file:line. Exits non-zero if
# either check fails, so it can gate CI or run on demand after a rename.
#
# A path that exists as neither a file nor a directory is a hard error (exit 2), never a
# silent skip: a CI job pointed at a renamed directory would otherwise stay green forever
# while checking nothing. So is a grep that fails for any reason other than "no match" —
# an unreadable directory or a broken pattern reported as "✓ clean" is that same vacuous
# pass one layer down.
#
# WHAT THIS CAN AND CANNOT DO — read before trusting a green run. This is a regression
# test for a fixed list of literals we know were retired. It proves those exact strings
# are absent. It does NOT prove a rename is complete: any paraphrase outside the list
# ("the full panel", "every reviewer slot", a roster written in a new order) passes clean,
# and so does every literal retired by the NEXT rename until someone adds it here. A green
# run means "the strings we listed have not come back" — nothing wider. Reviewing the diff
# for the rename's actual completeness is still a human job.
#
# Previously this lived inside a project-side file-edit hook, which had to fail
# open and so could only warn. As a standalone script it can fail loudly.

set -uo pipefail

die() {
  echo "ERROR: $*" >&2
  exit 2
}

# Temp files are registered here and removed by a single EXIT trap, so they are cleaned up on
# `die` and on a signal too, not only on the happy path.
TMP_FILES=()
cleanup_tmp() { [[ ${#TMP_FILES[@]} -eq 0 ]] || rm -f "${TMP_FILES[@]}"; }
trap cleanup_tmp EXIT

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# coverage_record / coverage_declare / coverage_report / coverage_verdict —
# per-target coverage reporting and the declared-vs-undeclared-zero decision,
# owned once in lib/coverage.sh rather than reinvented here. See that file's
# header for why (KAN-197) and check-guard-symlinks.sh / check-references.sh
# for the pattern this guard follows. A "target" here is a member of TREES
# below (a DEFAULT_TARGETS entry, or a caller-supplied path) — the unit
# collect_hits already scans one at a time, so it is the natural corpus
# member for this guard, distinct from an individual FILE within a target
# (which collect_hits enumerates and greps in one batch, never one at a time).
source "$SCRIPT_DIR/lib/coverage.sh"

# The default scan set: everything in this repo that can carry the vocabulary. Kept here
# and nowhere else, so callers only ever have to say `scripts/check-vocabulary.sh`.
#
# `openspec/specs` is not in this list because that tree is FROZEN — history, this repository's
# convention no matter which guard is asking (see scripts/lib/owned-corpus.sh's corpus
# definition for the guards that literally source it; this one does not — it has always had its
# own DEFAULT_TARGETS, never that library's scope roots — but the same tree gets the same answer
# either way), never linted again no matter what vocabulary it still carries. That is reason
# enough on its own today. It is also, separately,
# a record of an earlier decision made while the tree was still live: scanning `openspec/specs`
# was tried and reverted back then, because a requirement that forbids a retired term has to
# name that term to do its job (e.g. "SHALL NOT contain `gates`, `tested`, ..."), so a live spec
# written correctly still tripped a bare vocabulary scan. There was no honest fix for that false
# positive — rewording deletes the requirement, and a suppression marker is forbidden by this
# repo's own lint policy. Don't re-add `openspec/specs` here for either reason; if drift in a
# live spec needs catching, it needs a check that understands requirement structure, not a flat
# grep over retired literals.
DEFAULT_TARGETS=(skills rules commands commands-claude scripts README.md AGENTS.md CLAUDE.md)

if [[ $# -gt 0 ]]; then
  TREES=("$@")
else
  cd "$REPO_ROOT" || die "cannot enter repo root $REPO_ROOT"
  TREES=("${DEFAULT_TARGETS[@]}")
fi

# EXPECTED-ZERO TARGETS — established by enumerating every DEFAULT_TARGETS
# member against the real tree (2026-08-18, at df9d5dd): skills (77 files),
# rules (12), commands (6), commands-claude (5), scripts (57), README.md
# (1), AGENTS.md (1), CLAUDE.md (1) — every one of them non-zero. Left
# empty rather than populated with an invented member: this guard's own
# die() above already refuses a target that does not exist as a
# configuration error, not a corpus member, and nothing in this scan set
# legitimately enumerates to zero files today. Declared here, never
# inferred — exactly like check-references.sh's and
# check-guard-symlinks.sh's own lists — so if a target genuinely becomes
# empty later, its name is added here, not inferred from the tree.
EXPECTED_ZERO_TARGETS=()
EXPECTED_ZERO_REASON="enumerates to zero files under this guard's own find -L scan — declared here rather than treated as an error, for a target legitimately expected to sometimes carry none"

for target in "${TREES[@]}"; do
  [[ -e "$target" ]] || die "no such file or directory: $target"
done

# The opt-out marker.  Some lines must name retired vocabulary to do their job: the legacy-stage
# migration mapping has to spell out the old values it maps FROM, the paragraph describing this
# guard has to quote the tokens it bans, and this script has to list them to search for them.
# Those are documentation, not drift.  Such a line carries a trailing `<!-- vocab-guard:allow -->`
# (or `# vocab-guard:allow` in a script), which both checks honour.  The marker is deliberately
# per-line rather than per-file: exempting a whole file would let real drift hide behind one
# documentary sentence.  The filter below therefore matches the marker only in the CONTENT field
# of grep's output — a marker in a *path* must never exempt everything under it.
#
# Separating the two needs an unambiguous boundary between the path field and the content field,
# and a colon is not one: `grep -n` emits `path:line:content`, so a directory literally named
# `d:0:vocab-guard:allow/` supplies a `:<digits>:` AND the marker from the PATH alone and exempts
# its whole subtree, no matter where the pattern is anchored. `--null` makes grep terminate the
# filename with a NUL instead, which cannot occur inside a path or a text line; SEP below is that
# NUL after `tr` maps it to a byte the shell can carry through a variable. So the filter can
# require the marker to sit strictly after the separator — i.e. in the content.
SEP=$'\001'
MARKER_FILTER="^[^$SEP]*$SEP"'[0-9]+:.*vocab-guard:allow'

# abs_dir <dir> — the physical absolute path of a directory, links resolved. `pwd -P` is used
# because `realpath` is absent from stock macOS.
abs_dir() { (cd "$1" 2>/dev/null && pwd -P); }

# Files reached through a symlink are required to resolve under the scan root or under this
# repo — see the `-L` note in collect_hits below.
REPO_ROOT_REAL="$(abs_dir "$REPO_ROOT")" || die "cannot resolve repo root $REPO_ROOT"

# collect_hits <path> <grep-args...> — search one path, minus opted-out lines.
#
# The result is published in the global HITS_OUT rather than written to stdout, because callers
# would otherwise have to capture it with `$(...)` — and a `die` inside a command substitution
# exits only that subshell, so the run would carry on and print "✓ clean" after the very failure
# it was meant to abort on.
#
# Files are enumerated by `find -L` rather than by grep's own `-r`/`-R`, because neither flag
# traverses symlinked directories portably: macOS BSD grep 2.6.0-FreeBSD descends a symlinked
# subdirectory under NEITHER -r nor -R (verified), while GNU grep and ugrep do under -R. Every
# installed tree (~/.claude/skills, ~/.cursor/skills) is a farm of symlinked directories, so
# relying on grep's recursion means "✓ clean" over a tree that was never opened. `find -L`
# follows links on both platforms.
#
# What `-L` costs, and why the walk is bounded: following links means the walk is NOT confined
# to the named target. A single committed link — `skills/anything/link -> /`, or `-> $HOME` —
# turns a bare invocation into a filesystem-wide scan that reports hits at paths outside the
# repo entirely (verified: a link under skills/ reported a file in another tree as if it were
# part of the target). So the escape is bounded rather than the feature removed:
#   - `-xdev` keeps the walk off other filesystems (mounted volumes, network shares).
#   - Every enumerated file is resolved and required to live under the scan root or under this
#     repo — the second root is what keeps the installed-farm case working, since those links
#     resolve back into this checkout. Anything else aborts, naming the path. Silently dropping
#     it would be a scan reporting "✓ clean" over files it never opened, i.e. this guard's own
#     failure mode; the user is told to scan that root directly instead.
#
# A symlink LOOP is left to find, which is loud on both platforms rather than silent: BSD find
# reports the loop and exits 0, GNU find prints "File system loop detected" and exits 1, so the
# rc check below turns the GNU case into a hard exit 2. Worth naming because the trees this
# scans are symlink farms.
#
# `-H` is passed because GNU grep omits the filename when handed a single file, which would both
# strip the path from every reported hit and silently disable the marker filter above; BSD grep
# includes it. `-H` makes the output shape the same everywhere.
#
# `-a` is passed so a file containing a NUL is still reported as `file:line:content`. Without it
# grep collapses the whole file to "Binary file … matches", which breaks the header's promise
# that every hit carries a location. `-I` (skip binary files) was rejected as the other way to
# make the shape uniform: it makes a file that IS scanned today silently unscanned, which is the
# vacuous pass this guard exists to prevent — a garbled content line is the cheaper failure.
#
# Failures abort instead of being swallowed. find's status is checked (unreadable directory,
# broken symlink, symlink loop) and grep's status is checked — 0 (matched) and 1 (no match) are
# normal, >= 2 is a real error — for the marker filter as well as for the search itself. Neither
# stderr is redirected, so the reason stays visible. A swallowed error reported as "✓ clean" is
# the same vacuous pass this guard exists to prevent.
HITS_OUT=""
FILES_COUNT_OUT=""
collect_hits() {
  local target="$1"; shift
  local list_tmp raw rc filtered file root dir last_dir="" last_real=""
  local -a files=()
  HITS_OUT=""
  FILES_COUNT_OUT=""

  # For a file target the root is its directory: the file was named explicitly, so it cannot
  # be an escape. For a directory target it is the directory itself.
  if [[ -d "$target" ]]; then
    root="$(abs_dir "$target")" || die "cannot resolve scan target '$target'"
  else
    root="$(abs_dir "$(dirname "$target")")" || die "cannot resolve scan target '$target'"
  fi

  list_tmp="$(mktemp)"
  TMP_FILES+=("$list_tmp")
  # `__pycache__` is pruned rather than scanned. It holds COMPILED BYTECODE that
  # Python regenerates from the .py files this guard already reads, and a .pyc
  # embeds its source's docstrings verbatim -- so a stale one reports retired
  # vocabulary from a version of the source that no longer exists, at a
  # `file:line` pointing into a binary nobody can edit. It is gitignored and
  # untracked: not part of the corpus by this repository's own reckoning.
  # Found when the KAN-289 literals below matched a docstring inside a .pyc left
  # over from before that rename, while every real source file was already clean.
  find -L "$target" -xdev -name __pycache__ -prune -o -type f -print0 >"$list_tmp"
  rc=$?
  (( rc == 0 )) || die "find exited $rc while enumerating '$target' (reason above) — refusing to report a clean run"
  while IFS= read -r -d '' file; do
    # Resolve per directory, not per file: find walks depth-first, so consecutive entries
    # almost always share a parent and the subshell is paid once per directory.
    dir="${file%/*}"
    [[ "$dir" != "$file" ]] || dir="."
    if [[ "$dir" != "$last_dir" ]]; then
      last_dir="$dir"
      last_real="$(abs_dir "$dir")" || die "cannot resolve '$dir' while enumerating '$target'"
    fi
    if [[ "$last_real" != "$root" && "$last_real" != "$root"/* &&
          "$last_real" != "$REPO_ROOT_REAL" && "$last_real" != "$REPO_ROOT_REAL"/* ]]; then
      die "'$file' is reached through a symlink that leaves the scanned tree (it resolves to
  $last_real/$(basename "$file"), outside both '$root' and $REPO_ROOT_REAL).
  Following it would scan an unbounded amount of the filesystem and report hits at paths
  that are not part of the target. Remove the link, or scan that root directly."
    fi
    files+=("$file")
  done <"$list_tmp"
  # Set regardless of what follows (including the early return right below):
  # this is the KAN-197 coverage count for this target — how many files this
  # guard actually found to scan, not whether any of them carried a hit. A
  # target enumerating to zero files is "nothing was checked here," visible
  # even on a run that goes on to report clean.
  FILES_COUNT_OUT=${#files[@]}
  [[ ${#files[@]} -gt 0 ]] || return 0

  raw="$(grep -aHn --null "$@" "${files[@]}" | tr '\0' "$SEP")"
  rc=$?
  (( rc < 2 )) || die "grep exited $rc while scanning '$target' (reason above) — refusing to report a clean run"
  [[ -n "$raw" ]] || return 0
  filtered="$(printf '%s\n' "$raw" | grep -vE "$MARKER_FILTER")"
  rc=$?
  (( rc < 2 )) || die "grep exited $rc while applying the vocab-guard:allow filter for '$target' (reason above) — refusing to report a clean run"
  [[ -n "$filtered" ]] || return 0
  HITS_OUT="$(printf '%s\n' "$filtered" | tr "$SEP" ':')"
}

# Guard: retired stage vocabulary must never reappear in any command tree.
#
# A past rename updated the stage vocabulary in the rule file and the skills but never swept the
# COMMAND files, leaving four commands gating on stages that no longer existed and contradicting
# the skills they delegate to. This check makes that class of drift loud instead of silent.
#
# The legal states are owned by `skills/flow-contracts/pipeline.md` — see its `## State
# transitions` section. They are deliberately not copied here; a second list would drift from
# the first.
#
# `checkpoint` is deliberately NOT in the list below, though it was a retired flag. It is an
# ordinary English word that appears legitimately in prose about review checkpoints, so matching
# it as a literal produces hits that are not drift — and the only way to silence those is a
# `vocab-guard:allow` marker on a line that is telling the truth, which teaches the guard to lie.
# It is swept by hand instead.
#
# Legitimate exceptions, deliberately allowed:
#   - `requesting-code-review` — a real external Superpowers skill; never rename it. The pattern
#     already excludes it structurally (the `-` before `code-review` fails `[^-]`), so no  # vocab-guard:allow
#     whole-line filter is needed — one that dropped the line would also hide a genuine retired
#     token sitting on the same line.
#   - `code-review` — a real harness-provided review skill (`skills/myflow-do/SKILL.md` section 5  # vocab-guard:allow
#     invokes it for the `light` roster's third panel slot); never rename it. It is, byte for
#     byte, the same string as the myflow command retired in the twelve-stage collapse, so the
#     guard cannot tell "the skill" from "the old command" by the string alone — only by how each
#     is written. Every legitimate reference to the skill in this repository's prose writes it as
#     its own token in backticks immediately followed by the word `skill` — "the harness's  # vocab-guard:allow
#     `code-review` skill", "the harness offers no `code-review` skill" — never any other way.
#     A first attempt at this exception widened the pattern's prefix class from `[^-]` to  # vocab-guard:allow
#     `` [^-`] ``, excluding a backtick the same way `requesting-code-review` above excludes a
#     hyphen. That was wrong and was reverted: a prefix exclusion cannot tell "the skill" from
#     "the old command" either — it blinds the guard to BOTH, including the backtick-fenced
#     spelling this repository actually uses for the retired command everywhere else (e.g. "Run
#     the `code-review` command to advance the stage." exited 0 under that pattern — measured, not  # vocab-guard:allow
#     hypothetical). So the prefix class here is `(^|[^-])`, identical to  # vocab-guard:allow
#     `requesting-code-review`'s, and admits nothing by prefix that the guard did not already
#     admit before either exception existed.
#     The exemption is instead scoped to the exact shape above, per OCCURRENCE rather than per
#     line — a whole-line filter would hide a genuine retired token sitting on the same line as a
#     legitimate skill mention, which is exactly what this header's own opening paragraph on
#     `checkpoint` warns against. Concretely: every substring of a candidate hit's content
#     matching a backtick, `code-review`, a backtick, whitespace, and the word `skill` is removed,  # vocab-guard:allow
#     and the pattern is re-applied to what remains. If nothing left still matches, the line was
#     drift-free and is dropped. If a `code-review` token remains — bare, or backtick-fenced but  # vocab-guard:allow
#     followed by anything other than `skill` — the hit still stands, on the SAME line that also
#     carried the exempt mention.
#     What this still cannot catch: the retired command written as its own token in backticks
#     immediately followed by the word `skill` — that one shape is genuinely indistinguishable  # vocab-guard:allow
#     from the real skill mention it exists to admit, and passes clean. Every other spelling,
#     including a backtick-fenced one followed by any word other than `skill`, is still caught.
#   - `code-review-low` — a live flow_settings.reviewers slot id (`ValidReviewers` in  # vocab-guard:allow
#     stats/internal/store/settings.go); never rename it. A TRAILING `[^-]|$` class, mirroring  # vocab-guard:allow
#     `requesting-code-review`'s LEADING `[^-]` exclusion, was tried first and reverted: measured
#     against `code-review-heavy`, `code-review-related`, `code-review-only` and  # vocab-guard:allow
#     `code-review-panel` — none of them a live id — the trailing class let all four pass clean,  # vocab-guard:allow
#     because `\bcode-review\b` already places a word boundary right at the hyphen that follows,  # vocab-guard:allow
#     and "is the next character a hyphen" is true of every compound suffix, not just `-low`.
#     There is no ERE lookahead to ask "followed by `-low` specifically, and nothing more" in the
#     pattern itself, so the pattern instead matches `code-review` as a bare word with no  # vocab-guard:allow
#     trailing condition — exactly as `code-review-heavy` and its siblings above demand — and the  # vocab-guard:allow
#     one live exemption is carried per OCCURRENCE, the same mechanism the `code-review` skill  # vocab-guard:allow
#     exemption above already established: after `$pattern` collects a hit, every substring
#     matching `code-review-low` bounded by a non-hyphen character (or start/end of string) on
#     each side is stripped, and the pattern is re-applied to what remains. `code-review-low-tier`  # vocab-guard:allow
#     — not a live id — keeps its trailing hyphen through that strip and so still matches.
#   - Any line carrying the marker `vocab-guard:allow` — see above.
check_retired_stage_vocabulary() {
  # Retired by the twelve-stage → three-state rename (KAN-8): the twelve stage values, the
  # removed and renamed commands, the retired skill, and every removed flag.
  local pattern='awaiting-review|awaiting-test|(^|[^-])\bcode-review\b'   # vocab-guard:allow
  pattern+='|awaiting-proposal-review|proposal-done|awaiting-do-review'   # vocab-guard:allow
  pattern+='|do-review-started|do-done|awaiting-fix-review'               # vocab-guard:allow
  pattern+='|fix-review-started|awaiting-manual-test|manual-test-done'    # vocab-guard:allow
  pattern+='|awaiting-pr-review|review-done'                              # vocab-guard:allow
  pattern+='|myflow-full|myflow-fast-path|myflow-manual-test'             # vocab-guard:allow
  pattern+='|myflow-start-fix|myflow-start-done|myflow-do-fix'            # vocab-guard:allow
  pattern+='|myflow-do-done|myflow-do-manual-review|myflow-review-done'   # vocab-guard:allow
  pattern+='|myflow-state-advance|automerge|skip-manual-test'             # vocab-guard:allow
  pattern+='|skip-review|skip-propose|propose-only|full-panel'            # vocab-guard:allow
  pattern+='|commit-during-apply'                                         # vocab-guard:allow
  # Retired by the five-state → three-state collapse: the test and review commands folded  # vocab-guard:allow
  # into /myflow-do and /myflow-finish. They were still live commands when the list above was
  # first written, which is why they arrive separately.
  # Bounded so `myflow-test-setup` (a sandbox prefix in test-setup.sh) and the separately
  # listed `myflow-review-done` / `myflow-fast-path` are not matched twice or spuriously.  # vocab-guard:allow
  pattern+='|myflow-test([^-]|$)|myflow-review([^-]|$)'                     # vocab-guard:allow
  #
  # RETIRED BY THE myflow->flow RENAME (KAN-289). Five literals, and the list is
  # deliberately this short. A flat ban on the word `myflow` was measured first
  # and rejected: 852 lines in this guard's own scan set carry it, and almost all
  # of them are legitimate -- 478 are `/myflow-*` command literals that are VALUES
  # in the live stage_runs table, 115 are the self-review angle labels whose
  # legacy spelling check-self-review-report.sh must keep recognising, 57 are the
  # `.myflow` the hard-cutover detection exists to look for, and 15 are the
  # managed-block delimiters setup.sh compares byte-for-byte against files already
  # in the operator's home. Banning the word would demand a suppression marker on
  # roughly eight hundred lines, which is precisely the shape this file's own
  # `myflow-fast` note calls "teaching the guard to lie".
  #
  # What is listed below instead is every spelling that measured ZERO legitimate
  # occurrences in the live corpus after the rename -- so any future hit is drift
  # rather than history, and needs no marker to sit beside it. Measured, not
  # assumed: each was counted before being added.
  pattern+='|myflowd|myflow-postgres|myflow-contracts'                      # vocab-guard:allow
  pattern+='|MYFLOWD_|MYFLOW_'                                              # vocab-guard:allow
  # `myflow-fast` was ALSO retired by that same five-state collapse — but KAN-111
  # (operator-approved 2026-08-09) ships a real, live command of the identical name: keep the
  # new command's name, narrow this guard instead of renaming it.
  #
  # An excluded-character-class boundary was tried first, the same idiom the `code-review`  # vocab-guard:allow
  # exception above already established for this script's plain ERE (`grep -E`, no lookaround):
  # match `myflow-fast` as a bare word only when not immediately preceded by `/` (a path
  # segment) or a backtick (an inline-code start), and not immediately followed by `/` or `.` (a
  # path continuation) — `(^|[^/\`])myflow-fast([^-/.]|$)`. That correctly admits the new
  # command's path and backtick-fenced spellings (`skills/myflow-fast/`,
  # `commands/myflow-fast.md`, `` `myflow-fast` ``) and its own bare command mention
  # `/myflow-fast` (excluded because the leading `/` is indistinguishable from a path segment).
  # But it does NOT admit the new command's plain-word self-references, and this skill genuinely
  # has them — `skills/myflow-fast/SKILL.md` itself, measured against this exact pattern, still
  # trips on its own frontmatter (`name: myflow-fast`) and its own prose (`Using myflow-fast for
  # change`, `myflow-fast does not publish one`), none of which is preceded by `/` or a backtick
  # or followed by `/` or `.`. That is the identical shape the retired command's own bare
  # mention had, so no boundary-class technique can admit one without admitting the other — the
  # two are lexically the same string in the same context.
  #
  # So `myflow-fast` is handled the same way `checkpoint` and the `effort` VALUES above are:
  # dropped from this mechanical list and swept by hand, because it now collides with ordinary,
  # legitimate language the way they do, and the only way to silence that collision here is a
  # `vocab-guard:allow` marker on a line that would be telling the truth about a live command,
  # which teaches the guard to lie. `myflow-fast-path` remains listed on its own line above and  # vocab-guard:allow
  # is unaffected — it was always a distinct, separately-spelled retired name, not a collision.
  # What this drop costs, honestly: a bare, non-path, non-backtick reintroduction of the retired
  # command (e.g. "run myflow-fast next" with no leading `/`) now passes this guard clean, same
  # as a bare "checkpoint" or a bare "effort" value would. Per this script's own header, that was
  # already true of any paraphrase — this just names one more shape the fixed-literal list can't
  # safely cover once the literal itself is legitimate vocabulary.
  # Retired FIELD and GATE vocabulary. Omitting these is how a wholly stale file — the contracts
  # index, which still described stage boundaries and Gates B/C/D — passed a clean run.
  pattern+='|gates\.[a-zA-Z]|originStage|fastPath|REVIEWED_TREE|MERGE_BASE'   # vocab-guard:allow
  pattern+='|Gate [ABCD]\b|monotonic gates'                                   # vocab-guard:allow
  # Retired by the planning-effort rename (KAN-26): the state file's `effort` key became
  # `planningEffort`. Only the KEY is listed, and it is matched where a JSON field can actually
  # stand, so the ordinary English word is untouched: "planning effort" in prose, "best-effort
  # reconstruction" in the Jira contract, and an archived change slug ending in it are all
  # legitimate, and none of them is a field.
  #
  # The field shape is matched with the tolerance a hand-typed reintroduction actually has:
  # either quote style around the key, and optional whitespace between the closing quote and
  # the colon. The first version of this entry matched one exact spelling — double quotes,
  # colon flush against them — which is how every real JSON example in this repository is
  # written, and so a single-quoted key, or one with a space before its colon, passed clean.
  # Both were reproduced in a sandbox copy of the scan set before this line was widened and
  # are caught after it. That gap sat inside the header's stated limit (a fixed list of
  # literals, not a completeness proof), but the entry exists precisely to stop a hand-typed
  # regression, so the QUOTED spellings a hand types are the ones it covers — which is not all
  # of them, and the ones it does not are named under WHAT IS NOT MATCHED below.
  #
  # WHAT THE TWO ALTERNATIVES BELOW MATCH. Both require the key to be QUOTED — either quote
  # style — and additionally one of two JSON contexts, because quote-plus-colon alone is not one:
  # measured in a sandbox, the truthful sentence `Two things determine 'effort': the level and
  # the model.` tripped an earlier form of this entry and exited 1. Nothing in this tree hit that,
  # so it was dormant rather than broken — but by the reasoning recorded above for leaving
  # `checkpoint` out and below for leaving the level VALUES out, a pattern that can fire on a line
  # telling the truth is one whose only silencer is a `vocab-guard:allow` marker that lies. So
  # each alternative requires a JSON context on one side of the quoted key:
  #   1. the quoted key stands where a field stands — at the start of a line, after `{`, or after
  #      the backtick that opens an inline-code span (whitespace between is tolerated); or
  #   2. the value reads as JSON — a quoted string, `null`, or a `<placeholder>`.
  # Either is enough, so a field is caught by (1) inside a block and by (2) mid-sentence.
  #
  # A `,` sat in (1)'s anchor class and was removed for the same reason the quote requirement
  # exists. Measured in a sandbox, with either quote style around the key: it made the sentence
  # `Three settings exist, 'effort': low, medium, high are the names.` exit 1 — prose, not drift,
  # and dormant only because nothing in this tree happens to be written that way.
  # What removing it costs was measured rather than argued: a comma-preceded field whose value is
  # a quoted string, `null` or a `<placeholder>` is still caught by (2), which tests the value
  # rather than what precedes the key; what is no longer caught is a quoted key after a comma
  # whose value is ALSO unquoted, which is not JSON and not a line this pipeline writes — it is
  # the false-positive shape itself. Named rather than left to be discovered.
  #
  # WHAT IS NOT MATCHED — measured, not inferred, and named here because it is the shape a
  # hand-typed line in this repository most often has. An UNQUOTED key is matched by neither
  # alternative: `effort: low`, `- effort: low`, an inline-code `effort: null`, a bare backticked
  # `effort`, and the house-style sentence this rename deleted (a run of backticked field names
  # in prose) all pass clean. That is not a regression from any earlier form of this entry —
  # every one of them required the quotes too, so the unquoted coverage never existed. Widening
  # to reach it was considered and deliberately not taken: unquoted, `effort` is the ordinary
  # English word this repository uses in prose, so the alternation would fire on truthful lines
  # exactly as `checkpoint` above and the level VALUES below would. The unquoted spellings are
  # swept by hand, on the reasoning the VALUES paragraph below records and within the limit this
  # script's header states.
  #
  # The retired VALUES `medium` and `high` are deliberately NOT listed, for the same reason
  # `checkpoint` above is not: they are ordinary English words used throughout this repository,
  # and `Medium` is also a Jira priority name, so matching them as literals produces hits that are
  # not drift — and the only way to silence those is a `vocab-guard:allow` marker on a line that
  # is telling the truth, which teaches the guard to lie. Per this script's header the check
  # proves a fixed list of literals is absent, not that a rename is complete; the values are
  # swept by hand.
  pattern+="|([{\`]|^)[[:space:]]*['\"]effort['\"][[:space:]]*:"          # vocab-guard:allow
  pattern+="|['\"]effort['\"][[:space:]]*:[[:space:]]*(['\"]|null|<)"     # vocab-guard:allow
  local hits="" tree
  for tree in "${TREES[@]}"; do
    collect_hits "$tree" -E "$pattern"
    # Recorded here, once — not duplicated in check_retired_panel_vocabulary's
    # own identical loop below, which would scan the same targets a second
    # time and trip coverage_record's own already-recorded rejection. Both
    # checks share one enumeration of the same TREES, so one recording is
    # the whole answer for both.
    #
    # KAN-197 F8: the return is checked rather than ignored. This file runs
    # under `set -uo pipefail` with no `-e`, so an unchecked rejection here
    # (e.g. a future caller adding a second scan of the same tree) would be
    # silently swallowed and the run would go on to print "✓ clean" while its
    # own coverage breakdown contradicted that verdict — the exact shape F8
    # found live in this file before this fix.
    if ! coverage_record "$tree" "$FILES_COUNT_OUT"; then
      die "coverage_record failed for '$tree' (see stderr above)"
    fi
    [[ -z "$HITS_OUT" ]] || hits+="$HITS_OUT"$'\n'
  done

  # Drop this pattern's two occurrence-level exemptions (see "Legitimate exceptions" above) from
  # each candidate hit before deciding whether the line is real drift. Done here, once, on the
  # assembled "path:line:content" hits — not baked into $pattern itself — because both exemptions
  # are per-occurrence, not per-line:
  #   1. every substring matching a backtick, `code-review`, a backtick, whitespace, and the word  # vocab-guard:allow
  #      `skill` is removed;
  #   2. every substring matching `code-review-low`, bounded by a non-hyphen character or  # vocab-guard:allow
  #      start/end of string on each side, is removed — `code-review-low-tier` keeps its  # vocab-guard:allow
  #      trailing hyphen and is untouched by this step, so it still matches below.
  # The pattern is then re-applied to what remains. A hit that still matches after both removals
  # carries a genuine retired token — bare, backtick-fenced but followed by something other than  # vocab-guard:allow
  # `skill`, or a `code-review`-prefixed compound other than `-low` — and stands unchanged (the  # vocab-guard:allow
  # ORIGINAL content is reported, not the stripped version, so the printed hit still shows the
  # real line). A hit that matches nothing after the removals was drift-free and is dropped.
  #
  # Removal 2 loops to a fixed point rather than running once. `sed -E .../g` matches
  # non-overlapping spans only: a match's trailing `[^-]` boundary consumes the very character  # vocab-guard:allow
  # the next occurrence needs as ITS leading boundary, so two or more `code-review-low`  # vocab-guard:allow
  # occurrences separated by a single space strip only every other one in one pass — the
  # untouched survivor then still matches $pattern and is reported as false drift. Comma- or
  # period-separated repeats never collided (each occurrence gets its own boundary character),
  # which is why that shape alone was measured clean before this fix. Every successful match
  # removes exactly the 15 literal characters of `code-review-low` and puts back the SAME  # vocab-guard:allow
  # boundary characters it consumed, so the string strictly shortens on any pass that still
  # matches; looping until a pass changes nothing therefore always terminates, and what remains
  # at that point is only the genuinely non-exempt shape (e.g. a hyphenated compound like
  # `code-review-low-tier`, which the boundary classes never touch in any pass).  # vocab-guard:allow
  if [[ -n "$hits" ]]; then
    local filtered_hits="" hit_line content stripped next rc
    while IFS= read -r hit_line; do
      [[ -n "$hit_line" ]] || continue
      # "path:line:content" — strip the shortest "anything:digits:" prefix to isolate content.
      # Repo paths never contain a colon, so this split is unambiguous.
      content="${hit_line#*:*:}"
      stripped="$(printf '%s' "$content" | sed -E 's/`code-review`[[:space:]]+skill([^A-Za-z]|$)/\1/g')"   # vocab-guard:allow
      while :; do
        next="$(printf '%s' "$stripped" | sed -E 's/(^|[^-])code-review-low([^-]|$)/\1\2/g')"   # vocab-guard:allow
        [[ "$next" == "$stripped" ]] && break
        stripped="$next"
      done
      printf '%s' "$stripped" | grep -qE "$pattern"
      rc=$?
      (( rc < 2 )) || die "grep exited $rc while re-testing a hit after stripping exempt \`code-review\` occurrences — refusing to report a clean run"   # vocab-guard:allow
      (( rc != 0 )) || filtered_hits+="$hit_line"$'\n'
    done <<<"$hits"
    hits="$filtered_hits"
  fi

  if [[ -n "$hits" ]]; then
    printf '\n⚠ Retired myflow state vocabulary found:\n%s\n' "$hits" >&2
    printf 'State gates must match the `## State transitions` table in\n' >&2
    printf 'skills/flow-contracts/pipeline.md.\n' >&2
    return 1
  fi

  echo "✓ Stage-vocabulary guard: clean"
  return 0
}

# Guard: retired review-panel vocabulary must never reappear.
#
# The review panel was reduced and its reviewer slots renamed. Any surviving mention of a removed
# reviewer, the old panel size, or a retired pass name means one layer of the docs/commands/skills
# was left behind by the rename. The terms cover the ways the old roster was actually written —
# spelled-out slot names, roster fragments ("… + Senior", ", Conventions"), and every spelling of  # vocab-guard:allow
# the old panel size — because a narrow list passes clean over real drift.
#
# Matching stays case-SENSITIVE, and both capitalisations of the reviewer slot are listed
# separately as a result. Case-insensitive matching was tried and rejected: it flags the
# adversarial reviewer's own persona line ("You are an adversarial senior engineer"), which is
# prose about a reviewer's attitude, not the retired roster slot. The retired slot was always
# written capitalised.
#
# This remains a blacklist: see the header's "WHAT THIS CAN AND CANNOT DO". A paraphrase nobody
# thought to list still passes.
check_retired_panel_vocabulary() {
  local terms=(
    'Conventions & hygiene'   # vocab-guard:allow
    'senior-engineer-reviewer' # vocab-guard:allow
    'conventions-reviewer'    # vocab-guard:allow
    'all six'                 # vocab-guard:allow
    '1 primary + 5'           # vocab-guard:allow
    '1+5 agents'              # vocab-guard:allow
    '+ Senior'                # vocab-guard:allow
    ', Conventions'           # vocab-guard:allow
    'Economic Senior'         # vocab-guard:allow
    'two-agent'               # vocab-guard:allow
    '2-agent'                 # vocab-guard:allow
    'Architect pass'          # vocab-guard:allow
    ', Senior'                # vocab-guard:allow
    'Senior engineer'         # vocab-guard:allow
    'Senior Engineer'         # vocab-guard:allow
    'six-agent'               # vocab-guard:allow
    '6-agent'                 # vocab-guard:allow
    '5 additional'            # vocab-guard:allow
    'five additional'         # vocab-guard:allow
    'all 6'                   # vocab-guard:allow
  )
  local args=() term hits="" tree
  for term in "${terms[@]}"; do args+=(-e "$term"); done

  for tree in "${TREES[@]}"; do
    collect_hits "$tree" -F "${args[@]}"
    [[ -z "$HITS_OUT" ]] || hits+="$HITS_OUT"$'\n'
  done

  if [[ -n "$hits" ]]; then
    printf '\n⚠ Retired myflow review-panel vocabulary found:\n%s\n' "$hits" >&2
    printf 'The panel roster must match the rule file'"'"'s review-panel section.\n' >&2
    return 1
  fi

  echo "✓ Panel-vocabulary guard: clean"
  return 0
}

status=0
check_retired_stage_vocabulary || status=1
check_retired_panel_vocabulary || status=1

# COVERAGE — per-target count of how many files this guard actually found to
# scan, via scripts/lib/coverage.sh. A target enumerating to zero files and
# absent from EXPECTED_ZERO_TARGETS is folded into the overall exit status
# here, exactly as check-references.sh and check-guard-symlinks.sh do it —
# never a separate exit code of its own.
if [[ ${#EXPECTED_ZERO_TARGETS[@]} -gt 0 ]]; then
  for target in "${EXPECTED_ZERO_TARGETS[@]}"; do
    # KAN-197 F8: the return is checked rather than ignored, for the same
    # reason as coverage_record above — this file has no `-e` to fall back
    # on either.
    if ! coverage_declare "$target" "$EXPECTED_ZERO_REASON"; then
      die "coverage_declare failed for '$target' (see stderr above)"
    fi
  done
fi

if ! coverage_verdict_out="$(coverage_verdict)"; then
  # KAN-197 F4: trailing newline restored. Command substitution strips the
  # newline coverage_verdict's own last printed line ends with, and this
  # format string did not add one back, so the block ran into whatever
  # printed next on the same terminal line.
  printf '\n⚠ Coverage violation(s) — a scan target checked nothing:\n%s\n' "$coverage_verdict_out" >&2
  status=1
elif [[ "$status" -eq 0 ]]; then
  coverage_frag="$(coverage_report)"
  [[ -n "$coverage_frag" ]] && printf '  %s\n' "$coverage_frag"
fi

exit "$status"
