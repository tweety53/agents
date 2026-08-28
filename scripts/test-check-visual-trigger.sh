#!/usr/bin/env bash
# Assertion harness for check-visual-trigger.sh.
#
# Builds throwaway fixture trees under TMPDIR and points the guard at each,
# invoking the REAL scripts/check-visual-trigger.sh as a subprocess against
# REAL files on disk — never a copy of its logic, and never a hand-written
# expected-output string asserted without having run the guard to produce it.
#
# The changed-paths input is fed on stdin, one path per line, exactly as
# design.md's stage step 2 hands the guard `git diff --name-only` output.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-visual-trigger.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

DIRS=()
cleanup() {
  [ "${#DIRS[@]}" -eq 0 ] && return 0
  local d
  for d in "${DIRS[@]}"; do
    rm -rf "$d"
  done
}
trap cleanup EXIT

# new_root -> sets ROOT to a fresh sandbox project root carrying .flow/.
new_root() {
  ROOT="$(mktemp -d "${TMPDIR:-/tmp}/check-visual-trigger-test.XXXXXX")"
  DIRS+=("$ROOT")
  mkdir -p "$ROOT/.flow"
  CFG="$ROOT/.flow/project.md"
}

write_cfg() {
  printf '%s\n' "$1" > "$CFG"
}

# write_cfg_bom — identical to write_cfg, but a UTF-8 BOM (EF BB BF) is
# prepended as the file's literal first three bytes, exactly as
# test-check-visual-verification.sh's own copy of this helper does, and
# exactly as a BOM sits in a real file some editors and Windows tooling
# emit by default.
write_cfg_bom() {
  printf '\xef\xbb\xbf%s\n' "$1" > "$CFG"
}

# run_guard <changed-path...> -> sets RC and OUT, feeding each argument as a
# line on stdin to the real guard against $ROOT. No arguments means empty
# stdin (a diff that touched nothing).
run_guard() {
  set +e
  if [ "$#" -eq 0 ]; then
    OUT="$(: | "$GUARD" "$ROOT" 2>&1)"
  else
    OUT="$(printf '%s\n' "$@" | "$GUARD" "$ROOT" 2>&1)"
  fi
  RC=$?
  set -e
}

assert_rc() {
  local case_name="$1" want="$2"
  [ "$RC" -eq "$want" ] && pass "$case_name: exit $want" \
    || fail "$case_name: expected exit $want, got rc=$RC out=$OUT"
}

assert_out_contains() {
  local case_name="$1" needle="$2"
  case "$OUT" in
    *"$needle"*) pass "$case_name: output names '$needle'" ;;
    *) fail "$case_name: expected output to contain '$needle', got: $OUT" ;;
  esac
}

assert_out_not_contains() {
  local case_name="$1" needle="$2"
  case "$OUT" in
    *"$needle"*) fail "$case_name: expected output NOT to contain '$needle', got: $OUT" ;;
    *) pass "$case_name: output correctly omits '$needle'" ;;
  esac
}

MINIMAL_HEADER='## visual verification

| Setting | Value |
|---------|-------|
| `ui paths` | `stats/web/src/**` |
| `screenshots` | `stats/web/tests/visual` |

| Command | Runs |
|---------|------|
| `verify` | `npm run test:visual` |
| `capture` | `npx playwright test <spec>` |'

# ===========================================================================
# Case 1: no .flow/project.md at all — "not configured" is exit 2, never a
# silent no-match.
# ===========================================================================
new_root
rm -f "$CFG"
run_guard "stats/web/src/App.tsx"
assert_rc "case 1" 2

# ===========================================================================
# Case 2: .flow/project.md exists but declares no `## visual verification`
# section — exit 2, same "cannot answer" as case 1.
# ===========================================================================
new_root
write_cfg "# Project

## run

echo hi
"
run_guard "stats/web/src/App.tsx"
assert_rc "case 2" 2

# ===========================================================================
# Case 3: section declared, `ui paths` absent — exit 2.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "stats/web/src/App.tsx"
assert_rc "case 3" 2

# ===========================================================================
# Case 4: section declared, `ui paths` present but empty — exit 2.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "stats/web/src/App.tsx"
assert_rc "case 4" 2

# ===========================================================================
# Case 5: a matching changed path — exit 0.
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER"
run_guard "stats/web/src/App.tsx"
assert_rc "case 5" 0
assert_out_contains "case 5" "MATCH"

# ===========================================================================
# Case 6: only non-matching changed paths — exit 1, "configured and
# unmatched" is a different answer than case 1/2's "not configured".
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER"
run_guard "README.md" "stats/internal/api/handler.go"
assert_rc "case 6" 1

# ===========================================================================
# Case 7: empty stdin (no changed paths at all) — exit 1, trivially no match.
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER"
run_guard
assert_rc "case 7" 1

# ===========================================================================
# Case 8 (glob semantics): `**` spans MULTIPLE directory levels, not one
# segment.
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER"
run_guard "stats/web/src/components/dashboard/Panel.tsx"
assert_rc "case 8" 0

# ===========================================================================
# Case 9 (glob semantics): a leading `./` on the DECLARED glob is normalized
# away.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`./stats/web/src/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "stats/web/src/App.tsx"
assert_rc "case 9" 0

# ===========================================================================
# Case 10 (glob semantics): a leading `./` on the CHANGED path is normalized
# away too.
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER"
run_guard "./stats/web/src/App.tsx"
assert_rc "case 10" 0

# ===========================================================================
# Case 11 (glob semantics): an ABSOLUTE declared glob (leading `/`) is read as
# repository-relative, the same as its bare form.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`/stats/web/src/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "stats/web/src/App.tsx"
assert_rc "case 11" 0

# ===========================================================================
# Case 12 (glob semantics): an ABSOLUTE changed path never matches a relative
# glob — `git diff --name-only` never produces one, so this is a defined
# refusal, not a silent grant.
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER"
run_guard "/stats/web/src/App.tsx"
assert_rc "case 12" 1

# ===========================================================================
# Case 13 (glob semantics): a changed path outside every declared glob —
# ordinary no-match, not a special case, but pinned per tasks.md.
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER"
run_guard "README.md"
assert_rc "case 13" 1

# ===========================================================================
# Case 14 (glob semantics): a glob containing a SPACE is not split on it —
# only the comma is the separator.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/my folder/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "stats/web/my folder/App.tsx"
assert_rc "case 14" 0

# ===========================================================================
# Case 15: two comma-separated globs, only the second one matches.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`gymie-frontend/**, gymie-admin-frontend/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "gymie-admin-frontend/src/App.tsx"
assert_rc "case 15" 0

# ===========================================================================
# Case 16: the project root does not exist — exit 2.
# ===========================================================================
ROOT="/nonexistent/$$-$RANDOM"
run_guard "stats/web/src/App.tsx"
assert_rc "case 16" 2

# ===========================================================================
# Case 17: `.flow/project.md` is a directory, not a regular file — exit 2.
# ===========================================================================
new_root
rm -f "$CFG"
mkdir -p "$CFG"
run_guard "stats/web/src/App.tsx"
assert_rc "case 17" 2

# ===========================================================================
# Case 18: the file declares `## visual verification` twice — exit 2,
# ambiguous, never picking a winner.
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER

$MINIMAL_HEADER"
run_guard "stats/web/src/App.tsx"
assert_rc "case 18" 2

# ===========================================================================
# Case 19 (REPRODUCE, DON'T READ; task 19): `.flow/project.md` begins with a
# UTF-8 BOM, with `## visual verification` as the file's own first line. A
# BOM must not make a present section read as "not configured" — the
# consequence is worse here than in the other two guards, because this is
# the guard that decides whether flow.visual-verify runs at all: task 19's
# own reproduction showed this exact guard exiting 2 ("not configured")
# against a BOM-prefixed declaration before the fix.
#
# NON-VACUOUS BY CONSTRUCTION, not by a substring check on the message: a
# matching path must exit 0 with `MATCH` (case 19), and a non-matching path
# against the SAME BOM-prefixed config must exit 1, not 2 (case 19b) — only
# a parser that actually read the BOM'd section, rather than one still
# defaulting to a fixed "cannot answer" verdict, can produce two different
# exit codes from two different inputs against the same file.
# ===========================================================================
new_root
write_cfg_bom "$MINIMAL_HEADER"
run_guard "stats/web/src/App.tsx"
assert_rc "case 19" 0
assert_out_contains "case 19" "MATCH"
assert_out_not_contains "case 19" "not configured"

new_root
write_cfg_bom "$MINIMAL_HEADER"
run_guard "README.md"
assert_rc "case 19b" 1
assert_out_not_contains "case 19b" "not configured"

# ===========================================================================
# Case 20 (KAN-359, multi-glob): two comma-separated globs, EACH individually
# backticked — the shape the defect actually breaks. Case 15 above wraps the
# WHOLE value in a single pair of backticks, which the whole-cell strip
# leaves with no backticks at all after splitting and so never exercised the
# bug; this fixture backticks each element the way both real declarations
# (this repository's and Gymie's) actually write one. A path matching only
# the SECOND glob must still match, and the reported glob must read clean —
# no stray backtick — which is what makes this non-vacuous rather than a
# substring check on "MATCH".
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`a/**\`, \`src/gateway/src/main/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "src/gateway/src/main/resources/application.yml"
assert_rc "case 20" 0
assert_out_contains "case 20" 'matched `src/gateway/src/main/**`'

# ===========================================================================
# Case 21 (KAN-359, multi-glob): three individually-backticked globs — the
# middle and last elements each carry a stray backtick on both sides under
# the old whole-cell-strip order; a path matching only the THIRD proves both
# strays were removed, not just the outermost pair.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`a/**\`, \`b/**\`, \`stats/web/src/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "stats/web/src/App.tsx"
assert_rc "case 21" 0
assert_out_contains "case 21" 'matched `stats/web/src/**`'

# ===========================================================================
# Case 22 (KAN-359, multi-glob): a mix of a BARE element and a backticked
# one in the same value. Both must resolve to a clean glob — the bare one
# unaffected, and the backticked one stripped of its own backticks only.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | a/**, \`stats/web/src/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "a/foo.txt"
assert_rc "case 22" 0
assert_out_contains "case 22" 'matched `a/**`'

run_guard "stats/web/src/App.tsx"
assert_rc "case 22b" 0
assert_out_contains "case 22b" 'matched `stats/web/src/**`'

# ===========================================================================
# Case 23 (KAN-359, multi-glob + space): a glob containing a space, as one
# individually-backticked element of a multi-glob value — pins that the
# per-element strip added for the multi-glob fix still never touches an
# INTERIOR space, the same invariant case 14 pins for a single-glob value.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`a/**\`, \`stats/web/my folder/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "stats/web/my folder/App.tsx"
assert_rc "case 23" 0
assert_out_contains "case 23" 'matched `stats/web/my folder/**`'

# ===========================================================================
# Case 24 (KAN-359 task 4, glob_to_ere's stale `n`): `glob_to_ere` declares
# `local g="$1" out="" i=0 n=${#g} c nc` — every word on a `local` command is
# expanded BEFORE `local` runs, so `${#g}` measured the CALLER's `g`, not the
# `$1` being bound in the same statement. The per-element split loop above
# uses `g` as its own loop variable and leaves it holding the LAST raw
# glob's trimmed value once that loop ends, so every `glob_to_ere` call
# afterward silently truncated its OWN glob to that leftover length.
#
# NON-VACUOUS BY CONSTRUCTION, not a short-glob case: the value below puts a
# 76-character glob FIRST and a 4-character one LAST, so the stale leftover
# `n` (4, from the short trailing glob) is far shorter than the first glob's
# real 76. A changed path that only the LONG glob covers must still match,
# and the reported text must name that exact glob — a case built only from
# short globs (case 15, case 20's `a/**`, …) cannot catch this, which is
# exactly why nothing did until now.
# ===========================================================================
new_root
LONG_GLOB='src/very/long/path/that/exceeds/any/plausible/caller/variable/length/here/**'
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`$LONG_GLOB\`, \`x/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "src/very/long/path/that/exceeds/any/plausible/caller/variable/length/here/App.tsx"
assert_rc "case 24" 0
assert_out_contains "case 24" "matched \`$LONG_GLOB\`"

# ===========================================================================
# Case 25 (KAN-359 task 5, trim_glob_element's own single-pass strip): a
# MIDDLE element (neither first nor last in the comma list) padded with
# 3,000 backtick characters on each side. Middle, not boundary, on purpose:
# the shared table parser's own whole-cell `trimcell` (visual-table-cells.awk)
# already strips a leading/trailing run of whitespace and backticks off the
# WHOLE cell before the comma split ever runs, so padding placed at the very
# start of the first element or the very end of the last element never
# reaches trim_glob_element at all — a middle element's own padding is the
# only padding trimcell cannot see, since it only touches the cell's two
# boundaries.
#
# NON-VACUOUS BY CONSTRUCTION: a changed path matching ONLY the padded
# element's real glob must still match, and the reported glob text must read
# clean — no stray backtick or whitespace — which only a parser that
# actually stripped the padding, rather than one choking on it or leaving it
# attached, can produce.
# ===========================================================================
new_root
PAD="$(printf '`%.0s' $(seq 1 3000))"
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`a/**\`, ${PAD}gymie-admin-frontend/**${PAD}, \`stats/web/src/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "gymie-admin-frontend/src/App.tsx"
assert_rc "case 25" 0
assert_out_contains "case 25" 'matched `gymie-admin-frontend/**`'

# ===========================================================================
# Case 26 (KAN-359 task 9, split_cells's own char-at-a-time accumulation):
# the SAME shape as case 25 — a middle element padded so the whole-cell
# `trimcell` (visual-table-cells.awk) cannot reach the padding before
# split_cells's own per-cell scan runs — but at 60,000 backticks per side,
# the exact fixture task 9's own measurement used, and with a wall-clock
# bound instead of only a correctness assertion. `split_cells` used to build
# each cell with `cur = cur ch`, one character at a time, and every one of
# those copies the whole accumulated cell — quadratic in cell length; a
# second, larger, unrelated cost (a `$`-anchored regex match against a
# computed string being pathologically slow on this awk — see
# visual-table-cells.awk's own `trimcell` header) sat behind it. On the
# guard as it stood before this task's commit, THIS fixture costs ~15s; a
# fixed guard costs a small fraction of a second. 5 seconds is generous on
# both sides — comfortably above every fixed-guard run measured while
# writing this case (all under 0.1s) and comfortably below the ~15s a
# regression in either bug would cost.
#
# NON-VACUOUS BY CONSTRUCTION, same as case 25: the changed path matches
# ONLY the padded element's real glob, and the reported glob text must read
# clean, so a parser that silently truncated or mis-split the padded cell
# would fail the correctness assertion even if it happened to run fast.
# ===========================================================================
new_root
BIGPAD="$(printf '`%.0s' $(seq 1 60000))"
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`a/**\`, ${BIGPAD}gymie-admin-frontend/**${BIGPAD}, \`stats/web/src/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
CASE26_START="$SECONDS"
run_guard "gymie-admin-frontend/src/App.tsx"
CASE26_ELAPSED="$((SECONDS - CASE26_START))"
assert_rc "case 26" 0
assert_out_contains "case 26" 'matched `gymie-admin-frontend/**`'
if [ "$CASE26_ELAPSED" -le 5 ]; then
  pass "case 26: 60,000-backtick interior padding resolved in ${CASE26_ELAPSED}s (<= 5s)"
else
  fail "case 26: 60,000-backtick interior padding took ${CASE26_ELAPSED}s, expected <= 5s — the quadratic cell splitter (or the trimcell \$-anchor cost behind it) is back"
fi

# ===========================================================================
# Case 27 (KAN-359 task 10, split_cells's escape semantics): a cell holding
# `\|` — an escaped pipe — must NOT split the row. Read directly, `split_cells`
# unescapes `\|` to a literal `|` and keeps scanning the same cell; a splitter
# that instead treats every `|` as a delimiter would break this ROW into
# three cells (key, value-part-1, value-part-2), so the awk key/value scanner
# (`if (n != 2) next`) would never recognise it as `ui paths` at all.
#
# PLACED IN THIS HARNESS, NOT A NEW ONE: `visual-table-cells.awk` is a shared
# lib with no test file of its own — every one of its three callers is the
# only way to reach it, and this guard's own end-to-end pipeline (a real
# `.flow/project.md`, a real glob match) is what task 9's cases already use
# to exercise it. The escaped pipe sits INSIDE the value cell, one element,
# not at either boundary the whole-cell `trimcell` strips first — the same
# boundary-vs-interior distinction task 8's own finding turned on for
# `trim_glob_element` — so this fixture reaches `split_cells`'s escape
# branches rather than being trimmed away before they run.
#
# NON-VACUOUS BY CONSTRUCTION: the glob resolves to `a|b/**` only if the `\|`
# was unescaped to a literal `|` rather than splitting the cell; a changed
# path containing that same literal `|` character must match it, and the
# reported glob text must read `a|b/**` — clean, no stray backslash.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`a\|b/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard 'a|b/App.tsx'
assert_rc "case 27" 0
assert_out_contains "case 27" 'matched `a|b/**`'

# ===========================================================================
# Case 28 (KAN-359 task 10, split_cells's escape semantics): `\\` — an
# escaped backslash — must resolve to ONE literal backslash, not two and not
# zero. `split_cells` unescapes `\\` to a single `\` and continues; the value
# has no `|` at all here, so this pins the unescape itself rather than the
# split decision case 27 pins.
#
# The guard's own `sanitize_display` (scripts/lib/sanitize-display.sh)
# renders every literal backslash it prints as TWO visible backslash
# characters — a display-safety measure, not the value the guard matched
# against — so the assertion below expects that doubled form; the real glob
# and the real changed path each carry exactly one backslash.
#
# NON-VACUOUS BY CONSTRUCTION: a changed path with exactly one literal
# backslash matches the resolved glob only if `\\` collapsed to exactly one
# backslash — two would over-match nothing (`glob_to_ere` escapes a literal
# backslash to require exactly one), zero would silently drop it.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`a\\\\b/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard 'a\b/App.tsx'
assert_rc "case 28" 0
assert_out_contains "case 28" 'matched `a\\b/**`'

# ===========================================================================
# Case 29 (KAN-359 task 10, split_cells's escape semantics): a trailing lone
# backslash — one not followed by `|` or `\` — must be kept as a literal
# backslash, and the scan must terminate cleanly rather than hang or drop it.
# The glob's own last character IS the backslash, immediately followed by the
# cell's closing backtick, so this is the backslash-then-ordinary-character
# branch of `split_cells`, exercised at the point where nothing legitimate
# follows it to form a pair.
#
# NON-VACUOUS BY CONSTRUCTION: the resolved glob must still end in a literal
# backslash for a changed path ending in one to match — a mutation that
# treats the backslash as always pairing with whatever follows (here, the
# closing backtick) silently drops the backslash and closes the cell one byte
# early instead, which case 28's `\\` case cannot catch (no `|` or `\`
# follows their nc there) but this one does. Same doubled-backslash display
# caveat as case 28.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`a/**\\\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard 'a/App.tsx\'
assert_rc "case 29" 0
assert_out_contains "case 29" 'matched `a/**\\`'

# ===========================================================================
# Case 30 (KAN-359 task 10, split_cells's escape semantics): `\\|` — an
# escaped backslash immediately followed by a REAL delimiter — MUST split.
# The first two characters unescape to one literal backslash (case 28's own
# rule) and are consumed as a pair; the `|` immediately after is then fresh,
# unescaped input and splits the row exactly as an ordinary `|` would.
#
# This is the mirror of case 27: same two bytes read in the other order (`|`
# first there, `\\` first here) must resolve oppositely, which is what proves
# the escape scan advances by the right number of characters per pair rather
# than by a fixed stride.
#
# NON-VACUOUS BY CONSTRUCTION: splitting here breaks the "ui paths" row into
# three cells instead of two, so the awk key/value scanner's `n != 2` check
# never recognises it as `ui paths` — the guard reports the value absent
# (exit 2), not a mismatched or truncated glob. A splitter that failed to
# split here would instead resolve a one-element `ui paths` and exit 0 or 1.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`a\\\\|b/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard 'a/App.tsx'
assert_rc "case 30" 2
assert_out_contains "case 30" '`ui paths` is absent or empty'

# ===========================================================================
# Case 31 (KAN-359 task 10, split_cells's escape semantics): a lone backslash
# followed by an ordinary character in the MIDDLE of a glob. Case 29 already
# covers the backslash-then-ordinary-character branch, but only where that
# ordinary character is the cell's closing backtick — which `trimcell` strips
# whether the splitter consumed it or not. That placement makes case 29 blind
# to the branch's own advance distance.
#
# NON-VACUOUS BY CONSTRUCTION: `split_cells` must leave the character after a
# lone backslash for the next iteration (`rest = substr(rest, bs + 1)`) rather
# than consuming it as part of a pair. Advancing one byte too far here drops
# the `b`, resolving `a\c/**`, which no longer matches the changed path — so
# this case reds where all four of cases 27-30 stay green. Same doubled-
# backslash display caveat as case 28.
#
# Found by the review panel's mutation-testing slot, which reported that
# `rest = substr(rest, bs + 1)` -> `bs + 2` left all three harnesses of this
# shared lib green. Reproduced before writing this case, and confirmed red
# after.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`a\\bc/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard 'a\bc/App.tsx'
assert_rc "case 31" 0
assert_out_contains "case 31" 'matched `a\\bc/**`'

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
