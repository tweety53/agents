#!/usr/bin/env bash
# test-check-contract-budget.sh — harness for check-contract-budget.sh.
#
# Every case runs against a sandboxed fixture tree under mktemp, reached through
# CHECK_CONTRACT_BUDGET_ROOT, so the harness never depends on the real
# repository's current sizes — a harness that asserted against those would go red
# every time a covered file legitimately changed.
set -euo pipefail

GUARD="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-contract-budget.sh"
[ -x "$GUARD" ] || { printf 'guard not executable: %s\n' "$GUARD" >&2; exit 2; }

# budget_table [table-file] — print the guard's budgets() rows, one per line.
#
# $GUARD is a shell script, not a bare table — budgets() wraps its rows in a
# single-quoted heredoc so raising a budget is a diff of the table alone.
# Reading the whole script treats any stray line elsewhere in it whose first two
# whitespace-separated fields happen to look like a covered path and a number as
# a row of its own; a live example broke this exact harness (F21). So when
# $table is $GUARD, only the region between budgets()'s heredoc delimiters is
# emitted; a caller-supplied bare table (F18, F22) has no such wrapper and is
# read whole, unchanged.
budget_table() {
  local table="${1:-$GUARD}"
  if [ "$table" = "$GUARD" ]; then
    sed -n "/<<'EOF'\$/,/^EOF\$/p" "$table" | sed '1d;$d'
  else
    cat "$table"
  fi
}

# budget_row_bytes <path-relative-to-repo-root> [table-file] — read a row
# straight out of the guard's own budgets() table (never executed here; the
# second argument defaults to $GUARD and exists only so the F18/F22 cases below
# can exercise the duplicate-row and unterminated-row branches against a scratch
# table without touching the guard's real one). A fixture sized off this is
# over/under its real budget BY CONSTRUCTION: it stays correct however the row
# changes. A fixture sized off a magic number copied from the table instead goes
# stale the next time that row rises — the defect the guard's own fix round
# shipped (F15).
#
# Reads the table with a bash `while read -r` comparison — the same idiom
# budget_for uses in $GUARD — never `awk -v want=`. The guard's own comment on
# budget_for documents why: a path containing a newline makes awk's -v
# assignment fail and, under `set -e`, abort the whole run rather than stay a
# per-file violation. Unreachable here, since every call below passes a fixed
# literal rather than a name read off the tree, but this harness sits beside
# the guard and should not contradict its stated convention (F19).
#
# `|| [ -n "$brel" ]` on the read keeps a table's final row from vanishing
# when that row carries no trailing newline — plain `while read -r` silently
# drops an unterminated last line, and this helper now accepts a
# caller-supplied table, so that was an undocumented precondition on
# generalized behaviour (F22).
#
# A path matched by more than one row is refused rather than silently
# resolved to the first match: this harness is reading a table it does not
# own, and an ambiguous answer is worse than a loud one (F18). budget_for in
# $GUARD does resolve to the first match — that tolerance is the guard's own
# decision and is not this finding.
budget_row_bytes() {
  local path="$1" table="${2:-$GUARD}" brel bmax bytes= found=0
  while read -r brel bmax || [ -n "$brel" ]; do
    [ "$brel" = "$path" ] || continue
    if [ "$found" -eq 1 ]; then
      printf 'duplicate budget row for %s in %s\n' "$path" "$table" >&2
      exit 2
    fi
    bytes="$bmax"
    found=1
  done < <(budget_table "$table")
  if [ "$found" -eq 0 ]; then
    printf 'no budget row for %s in %s\n' "$path" "$table" >&2
    exit 2
  fi
  printf '%s' "$bytes"
}

failures=0

# expect <label> <want-exit> <root>
expect() {
  local label="$1" want="$2" root="$3" got=0
  CHECK_CONTRACT_BUDGET_ROOT="$root" "$GUARD" >/dev/null 2>&1 || got=$?
  if [ "$got" -ne "$want" ]; then
    printf 'FAIL %s: exit %d, wanted %d\n' "$label" "$got" "$want"
    failures=$((failures + 1))
  else
    printf 'ok   %s\n' "$label"
  fi
}

FIX="$(mktemp -d)"
trap 'rm -rf "$FIX"' EXIT

# mkroot <dir> — a fixture root carrying every scope root the corpus library
# requires, plus the contracts directory most fixtures below put a file in.
#
# scripts/lib/owned-corpus.sh refuses a root with a missing scope root rather
# than skipping it, so a fixture that creates only the directory its case cares
# about exits 2 for a reason that has nothing to do with the case. Building the
# full set here keeps every exit code below attributable to the behaviour under
# test.
mkroot() {
  mkdir -p "$1/skills/myflow-contracts" "$1/rules" "$1/spectre/specs" \
    "$1/commands" "$1/commands-claude" "$1/.flow"
}

# --- budget_row_bytes's own behaviour (F17, F18, F22) ----------------------
#
# These cases exercise the helper directly rather than through the guard —
# each branch is new this round (or new to this round's coverage) and no
# other case in this file reaches them. The scratch tables they need live
# under $FIX so their cleanup rides the same trap as every other fixture
# here, rather than needing one of its own (F23).

# The missing-row path (exit 2, a message naming the path) was new shared
# behaviour with zero coverage: replacing it with `bytes=0` left every other
# case in this file green — a demonstrated surviving mutant (F17). The `||`
# keeps the command substitution's exit 2 from killing the harness under
# `set -e`. Asserting the exact wording, not just the path, is what tells
# this refusal apart from the duplicate-row refusal below — both exit 2 and
# both name the path, so a substring-only check cannot distinguish a correct
# diagnosis from a wrong one; F20 found this same gap in the duplicate case
# and it turned out F17's case (this one) shared it.
rc=0
out="$(budget_row_bytes 'skills/does-not-exist/SKILL.md' 2>&1)" || rc=$?
if [ "$rc" -eq 2 ] && printf '%s' "$out" | grep -q '^no budget row for skills/does-not-exist/SKILL\.md in'; then
  printf 'ok   budget_row_bytes on a path with no row exits 2 and names the path\n'
else
  printf 'FAIL budget_row_bytes on a missing path: exit %s, output: %s\n' "$rc" "$out"
  failures=$((failures + 1))
fi

# A path matched by more than one row must be refused rather than silently
# resolved to the first match (F18). Asserting the "duplicate" wording, not
# just the path, is what keeps this case from passing when the helper
# actually took the missing-row branch instead (F20).
DUP_TABLE="$FIX/dup-table"
printf 'skills/dup/SKILL.md 100\nskills/dup/SKILL.md 200\n' > "$DUP_TABLE"
rc=0
out="$(budget_row_bytes 'skills/dup/SKILL.md' "$DUP_TABLE" 2>&1)" || rc=$?
if [ "$rc" -eq 2 ] && printf '%s' "$out" | grep -q '^duplicate budget row for skills/dup/SKILL\.md in'; then
  printf 'ok   budget_row_bytes refuses a path matched by more than one row\n'
else
  printf 'FAIL budget_row_bytes on a duplicated row: exit %s, output: %s\n' "$rc" "$out"
  failures=$((failures + 1))
fi

# A table whose last row carries no trailing newline. Plain `while read -r`
# drops an unterminated final line, so a row sitting there would be invisible
# to both the found and the duplicate branches above — an undocumented
# precondition on a helper that now accepts a caller-supplied table (F22).
UNTERMINATED_TABLE="$FIX/unterminated-table"
printf 'skills/unterminated/SKILL.md 999' > "$UNTERMINATED_TABLE"
rc=0
out="$(budget_row_bytes 'skills/unterminated/SKILL.md' "$UNTERMINATED_TABLE" 2>&1)" || rc=$?
if [ "$rc" -eq 0 ] && [ "$out" = "999" ]; then
  printf 'ok   budget_row_bytes reads a final row with no trailing newline\n'
else
  printf 'FAIL budget_row_bytes on an unterminated final row: exit %s, output: %s\n' "$rc" "$out"
  failures=$((failures + 1))
fi

# --- The table's own shape -------------------------------------------------

# EVERY ROW NAMES A FILE, never a directory. A directory total is not an
# acceptable substitute for per-file rows: one file can balloon while another
# shrinks with the sum unchanged, and the guard says nothing. The behavioural
# half of this property is the two-spec-file case further down; this half
# catches the table being replaced by directory totals in the first place,
# which that case cannot see until a file actually goes over.
bad_rows=0
while read -r brel _bmax; do
  [ -n "$brel" ] || continue
  case "$brel" in
    *.md | *.mdc) ;;
    *)
      printf 'FAIL budgets() row names something that is not a Markdown file: %s\n' "$brel"
      bad_rows=$((bad_rows + 1))
      ;;
  esac
done < <(budget_table)
if [ "$bad_rows" -eq 0 ]; then
  printf 'ok   every budgets() row names a file, not a directory\n'
else
  failures=$((failures + bad_rows))
fi

# --- Whole-guard behaviour -------------------------------------------------

# A file well under its declared budget.
mkroot "$FIX/within"
printf 'x\n' > "$FIX/within/skills/myflow-contracts/build-green.md"
expect 'a file well under budget passes' 0 "$FIX/within"

# The same file pushed past its declared budget, read straight from the
# guard's own table so this fixture is over budget by construction rather
# than by a number that must stay above a row that can rise (see F15).
mkroot "$FIX/over"
build_green_budget="$(budget_row_bytes 'skills/myflow-contracts/build-green.md')"
head -c "$((build_green_budget + 500))" /dev/zero | tr '\0' 'x' \
  > "$FIX/over/skills/myflow-contracts/build-green.md"
expect 'a file over budget fails' 1 "$FIX/over"

# A contract added later with no row in budgets(). This is the case that stops a
# new file escaping the ratchet silently.
mkroot "$FIX/undeclared"
printf 'x\n' > "$FIX/undeclared/skills/myflow-contracts/brand-new-contract.md"
expect 'a file with no budget row fails' 1 "$FIX/undeclared"

# A complete set of scope roots holding no Markdown at all. Reporting a clean
# run here is indistinguishable from a real pass, and one of the two readings is
# "the root resolved somewhere that holds none of this repository's text".
mkroot "$FIX/empty"
expect 'a corpus with no Markdown at all cannot be answered' 2 "$FIX/empty"

# A scope root missing under an otherwise real root. The corpus library refuses
# rather than skipping it: skipping would hand this guard a quietly smaller
# corpus that still looks like a complete answer, so every file under the
# missing root would drop out of the ratchet with the run still green.
mkdir -p "$FIX/nodir/skills" "$FIX/nodir/rules"
printf 'x\n' > "$FIX/nodir/skills/README.md"
expect 'a missing scope root cannot be answered' 2 "$FIX/nodir"

# A scope root replaced by a SYMLINK to a directory. Every readability test the
# corpus library applies to a scope root follows symlinks, so a link passes
# them; `find` then runs in physical mode and does not descend into a symlinked
# starting point, so that whole scope root contributes zero files — the same
# quietly smaller corpus the missing-root case above pins, reached by a
# different door. This guard's own symlink sweep does not catch it: the sweep
# filters on `*.md` and `*.mdc`, and a scope-directory link is named `skills`.
# The fixture carries a covered file with a real budget row under a DIFFERENT
# scope root, so a guard that skipped the link reports a clean BUDGET-OK over
# the remaining file rather than exiting 2 through the empty-corpus refusal for
# the wrong reason.
mkroot "$FIX/symlinked-root"
printf 'x\n' > "$FIX/symlinked-root/.flow/project.md"
mkdir -p "$FIX/symlinked-root/real-skills/myflow-contracts"
printf 'x\n' > "$FIX/symlinked-root/real-skills/myflow-contracts/build-green.md"
rm -rf "$FIX/symlinked-root/skills"
ln -s "$FIX/symlinked-root/real-skills" "$FIX/symlinked-root/skills"
expect 'a symlinked scope root cannot be answered' 2 "$FIX/symlinked-root"

# A symlinked DIRECTORY nested INSIDE a scope root, whose target holds Markdown.
# The scope-root refusal above does not reach it: it tests `$root/$dir` alone,
# and both the corpus `find` and this guard's own symlink sweep run in physical
# mode, so each walks past a symlinked subdirectory without descending into it.
# Every covered file under the link then escapes the ratchet with the run still
# green — the "escapes silently" failure this guard exists to prevent, one level
# deeper than the case above. The fixture carries a declared covered file
# outside the link, so a guard that skipped the link reports a clean BUDGET-OK
# over that file rather than exiting 2 through the empty-corpus refusal for the
# wrong reason.
mkroot "$FIX/nested-link-md"
printf 'x\n' > "$FIX/nested-link-md/skills/myflow-contracts/build-green.md"
mkdir -p "$FIX/nested-link-md/outside/hidden"
head -c 300000 /dev/zero | tr '\0' 'x' \
  > "$FIX/nested-link-md/outside/hidden/huge.md"
ln -s "$FIX/nested-link-md/outside/hidden" \
  "$FIX/nested-link-md/skills/nested-link"
expect 'a nested symlinked directory holding Markdown cannot be answered' 2 \
  "$FIX/nested-link-md"

# The same shape carrying NO Markdown — this repository's own
# skills/myflow-*/scripts/lib links, which point at a directory of `.sh` files.
# A symlinked directory with nothing covered under it hides no bytes from the
# ratchet, so refusing it would fail the real repository on every run for no
# escaped file. The verdict COUNT is asserted alongside the exit code: only the
# count distinguishes a guard that skipped the link from one that followed it
# and found nothing to add.
mkroot "$FIX/nested-link-nomd"
printf 'x\n' > "$FIX/nested-link-nomd/skills/myflow-contracts/build-green.md"
mkdir -p "$FIX/nested-link-nomd/outside/lib"
printf 'echo hi\n' > "$FIX/nested-link-nomd/outside/lib/helper.sh"
ln -s "$FIX/nested-link-nomd/outside/lib" \
  "$FIX/nested-link-nomd/skills/nested-link"
expect 'a nested symlinked directory holding no Markdown is skipped, not refused' \
  0 "$FIX/nested-link-nomd"
out="$(CHECK_CONTRACT_BUDGET_ROOT="$FIX/nested-link-nomd" "$GUARD" 2>&1 || true)"
if printf '%s' "$out" | grep -q '^BUDGET-OK: 1 owned Markdown file(s) within budget$'; then
  printf 'ok   a nested symlinked directory holding no Markdown is not counted\n'
else
  printf 'FAIL nested link with no Markdown: %s\n' "$out"
  failures=$((failures + 1))
fi

# A root that does not exist at all.
expect 'a nonexistent root cannot be answered' 2 "$FIX/absent"

# A covered file that cannot be READ. Every other exit-2 case here is a
# directory-shape problem; this is the only file-level one, and without it a
# regression in the wc guard ships silently — the guard would abort with wc's own
# status of 1, which this contract reserves for "over budget or undeclared", and
# send a caller off to edit budgets() when the guard could not read the file.
mkroot "$FIX/unreadable"
printf 'x\n' > "$FIX/unreadable/skills/myflow-contracts/build-green.md"
chmod 000 "$FIX/unreadable/skills/myflow-contracts/build-green.md"
if [ "$(id -u)" = "0" ]; then
  printf 'skip an unreadable file cannot be answered (running as root reads it anyway)\n'
else
  expect 'an unreadable file cannot be answered' 2 "$FIX/unreadable"
fi
chmod 644 "$FIX/unreadable/skills/myflow-contracts/build-green.md"

# A dangling .md symlink beside real files. The corpus library enumerates with
# `find -type f`, so the link is not owned content and is never followed — but
# silence would be the wrong answer here: a ratcheted file replaced by a symlink
# would drop out of the corpus with the run still green. The guard names it and
# counts it as a violation.
mkroot "$FIX/dangling"
printf 'x\n' > "$FIX/dangling/skills/myflow-contracts/build-green.md"
ln -s /nonexistent/target "$FIX/dangling/skills/myflow-contracts/aaa-ghost.md"
expect 'a dangling .md symlink is a per-file violation' 1 "$FIX/dangling"
out="$(CHECK_CONTRACT_BUDGET_ROOT="$FIX/dangling" "$GUARD" 2>&1 || true)"
if printf '%s' "$out" | grep -q 'is a symlink'; then
  printf 'ok   a dangling symlink is named as a symlink\n'
else
  printf 'FAIL unexpected message for a dangling symlink: %s\n' "$out"
  failures=$((failures + 1))
fi

# A symlink pointing OUTSIDE the tree must be refused before it is sized —
# otherwise the guard is an existence-and-size oracle for arbitrary paths, and a
# target whose name matches a budget row has its byte count printed.
mkroot "$FIX/escape"
printf 'x\n' > "$FIX/escape/skills/myflow-contracts/build-green.md"
head -c 9000 /dev/zero | tr '\0' 'x' > "$FIX/escape/outside-the-tree.txt"
ln -s "$FIX/escape/outside-the-tree.txt" "$FIX/escape/skills/myflow-contracts/SKILL.md"
expect 'a symlink escaping the tree is a violation' 1 "$FIX/escape"
out="$(CHECK_CONTRACT_BUDGET_ROOT="$FIX/escape" "$GUARD" 2>&1 || true)"
if printf '%s' "$out" | grep -q '9000'; then
  printf 'FAIL the guard printed the size of a file outside the tree: %s\n' "$out"
  failures=$((failures + 1))
else
  printf 'ok   the guard does not size a symlink target outside the tree\n'
fi

# A filename containing a newline must stay a per-file violation rather than
# aborting the whole run — the awk -v failure mode the security slot found.
mkroot "$FIX/newline"
printf 'x\n' > "$FIX/newline/skills/myflow-contracts/$(printf 'we\nird').md"
expect 'a filename containing a newline is a per-file violation' 1 "$FIX/newline"

# An override that is set but empty must be refused rather than silently falling
# back to the script's own location — a fallback there would make the harness
# scan the real repository while believing it scanned a fixture.
rc=0
CHECK_CONTRACT_BUDGET_ROOT='' "$GUARD" >/dev/null 2>&1 || rc=$?
if [ "$rc" -ne 2 ]; then
  printf 'FAIL an empty override exits %d, wanted 2\n' "$rc"
  failures=$((failures + 1))
else
  printf 'ok   an empty override is refused\n'
fi

# The verdict carries a count, which is what distinguishes a real pass from a
# run that scanned nothing.
out="$(CHECK_CONTRACT_BUDGET_ROOT="$FIX/within" "$GUARD")"
if printf '%s' "$out" | grep -q '^BUDGET-OK: 1 owned Markdown file(s) within budget$'; then
  printf 'ok   the verdict names how many files were checked\n'
else
  printf 'FAIL verdict did not carry the count: %s\n' "$out"
  failures=$((failures + 1))
fi

# --- skills/*/SKILL.md and skills/*/SKILL-rationale.md ---------------------
#
# The paths below are deliberately real repository skill paths. Each fixture's
# size is read from the guard's own budgets() table via budget_row_bytes rather
# than copied as a number, so it stays over/under its real row by construction
# as that row moves — the coupling that broke in F15.

# A skills/<x>/SKILL.md that exceeds its declared budget. skills/myflow-status
# only has a SKILL.md, no rationale sibling, which keeps this fixture simple.
mkroot "$FIX/skill-over"
mkdir -p "$FIX/skill-over/skills/myflow-status"
printf 'x\n' > "$FIX/skill-over/skills/myflow-contracts/build-green.md"
myflow_status_budget="$(budget_row_bytes 'skills/myflow-status/SKILL.md')"
head -c "$((myflow_status_budget + 1000))" /dev/zero | tr '\0' 'x' \
  > "$FIX/skill-over/skills/myflow-status/SKILL.md"
expect 'a skills/<x>/SKILL.md over budget fails' 1 "$FIX/skill-over"

# A skills/<x>/SKILL.md under a skill name that has no row in budgets() at all.
mkroot "$FIX/skill-undeclared"
mkdir -p "$FIX/skill-undeclared/skills/mystery-skill"
printf 'x\n' > "$FIX/skill-undeclared/skills/myflow-contracts/build-green.md"
printf 'x\n' > "$FIX/skill-undeclared/skills/mystery-skill/SKILL.md"
expect 'a skills/<x>/SKILL.md with no budget row fails' 1 "$FIX/skill-undeclared"

# Two skills whose SKILL.md files carry different budgets in the real table —
# skills/myflow-research/SKILL.md (a small budget) and skills/myflow-do/SKILL.md (a
# much larger one). skills/myflow-do/SKILL.md is sized to the midpoint between
# the two real rows, read from the guard itself: above myflow-research's row,
# below myflow-do's row, by construction rather than by a number that must stay
# between two rows that can each move independently. If the table were still
# keyed on the bare basename "SKILL.md" — the collision this key fixes —
# every skills/*/SKILL.md would be checked against one shared row, and this
# fixture would fail against the small budget that basename collision would
# hand it. Keyed on the path relative to the repository root, each file is
# checked against its own row and both pass.
mkroot "$FIX/skill-distinct"
mkdir -p "$FIX/skill-distinct/skills/myflow-research" "$FIX/skill-distinct/skills/myflow-do"
printf 'x\n' > "$FIX/skill-distinct/skills/myflow-contracts/build-green.md"
printf 'x\n' > "$FIX/skill-distinct/skills/myflow-research/SKILL.md"
myflow_research_budget="$(budget_row_bytes 'skills/myflow-research/SKILL.md')"
myflow_do_budget="$(budget_row_bytes 'skills/myflow-do/SKILL.md')"

# The fixture only discriminates the basename-collision bug when mid_size
# lands STRICTLY BETWEEN the two rows. If the rows ever converge, integer
# division makes mid_size collapse onto myflow_research_budget and the case
# below would keep reporting "ok" even with the collision bug back in place —
# F15's own defect class (a fixture that stops discriminating when its
# premise erodes), reintroduced by the round that fixed F15 (F16). Assert the
# premise rather than assume it, and fail the whole harness loudly, naming
# both rows, when it does not hold: a fixture that cannot be constructed to
# discriminate must say so, never quietly pass.
if [ "$((myflow_do_budget - myflow_research_budget))" -lt 2 ]; then
  printf 'cannot construct a discriminating fixture: skills/myflow-research/SKILL.md is %s bytes and skills/myflow-do/SKILL.md is %s bytes — too close together for a strict midpoint\n' \
    "$myflow_research_budget" "$myflow_do_budget" >&2
  exit 2
fi
mid_size=$((myflow_research_budget + (myflow_do_budget - myflow_research_budget) / 2))
head -c "$mid_size" /dev/zero | tr '\0' 'x' > "$FIX/skill-distinct/skills/myflow-do/SKILL.md"
expect 'two skills with different budgets are each checked against their own row' \
  0 "$FIX/skill-distinct"

# A SKILL-rationale.md over budget. Sized from the guard's own row (see F15: a
# hardcoded 20000 here fell under the row when a fix round raised it, and this
# case stopped catching an over-budget file at all).
mkroot "$FIX/rationale"
mkdir -p "$FIX/rationale/skills/myflow-do"
printf 'x\n' > "$FIX/rationale/skills/myflow-contracts/build-green.md"
myflow_do_rationale_budget="$(budget_row_bytes 'skills/myflow-do/SKILL-rationale.md')"
head -c "$((myflow_do_rationale_budget + 1000))" /dev/zero | tr '\0' 'x' \
  > "$FIX/rationale/skills/myflow-do/SKILL-rationale.md"
expect 'a SKILL-rationale.md over budget fails' 1 "$FIX/rationale"

# A SKILL-rationale.md with no budget row — the undeclared case for the same tree.
mkroot "$FIX/rationale-new"
mkdir -p "$FIX/rationale-new/skills/brand-new"
printf 'x\n' > "$FIX/rationale-new/skills/myflow-contracts/build-green.md"
printf 'x\n' > "$FIX/rationale-new/skills/brand-new/SKILL-rationale.md"
expect 'a SKILL-rationale.md with no budget row fails' 1 "$FIX/rationale-new"

# Each covered file is counted exactly once. Deleting a de-duplication step, or
# enumerating one tree twice, leaves every exit code above green while the
# verdict's count silently inflates, so this case asserts the COUNT.
mkroot "$FIX/nodouble"
printf 'x\n' > "$FIX/nodouble/skills/myflow-contracts/build-green.md"
printf 'x\n' > "$FIX/nodouble/skills/myflow-contracts/SKILL.md"
expect 'a contracts-dir SKILL.md passes' 0 "$FIX/nodouble"
out="$(CHECK_CONTRACT_BUDGET_ROOT="$FIX/nodouble" "$GUARD")"
if printf '%s' "$out" | grep -q '^BUDGET-OK: 2 owned Markdown file(s) within budget$'; then
  printf 'ok   each covered file is counted once, not twice\n'
else
  printf 'FAIL double-count: %s\n' "$out"
  failures=$((failures + 1))
fi

# --- The widened corpus ----------------------------------------------------
#
# These cases exercise the widening from 13 covered files to every owned
# Markdown file. Each was run against the pre-widening guard before being
# trusted: the three exit-1 cases below reported exit 0 there, because the file
# they name was not covered at all.

# THE FROZEN openspec/specs/ TREE IS NOT COVERED, however large a file placed
# there is. It was covered before this guard's corpus moved to spectre/specs/
# — 462 KB of text a change under the old tree could grow without this guard
# ever noticing was the reason it was widened onto that path in the first
# place — but that tree is history now: scripts/lib/owned-corpus.sh's scope
# roots no longer name it, so owned_corpus_files never enumerates a file
# under it, and this guard cannot ratchet what it never sees. Modeled on the
# "excluded tree" cases below: the verdict COUNT is asserted, not just the
# exit code, because a guard that covered this file would fail it for having
# no row, while a guard that merely counted it and found a row would pass
# with the count inflated — only the count tells the two apart.
mkroot "$FIX/frozen-tree"
mkdir -p "$FIX/frozen-tree/openspec/specs/myflow-build-green"
printf 'x\n' > "$FIX/frozen-tree/skills/myflow-contracts/build-green.md"
head -c 300000 /dev/zero | tr '\0' 'x' \
  > "$FIX/frozen-tree/openspec/specs/myflow-build-green/spec.md"
expect 'a file under the frozen openspec/specs/ tree is not covered' 0 "$FIX/frozen-tree"
out="$(CHECK_CONTRACT_BUDGET_ROOT="$FIX/frozen-tree" "$GUARD" 2>&1 || true)"
if printf '%s' "$out" | grep -q '^BUDGET-OK: 1 owned Markdown file(s) within budget$'; then
  printf 'ok   the frozen tree contributes nothing to the count\n'
else
  printf 'FAIL the frozen tree was counted: %s\n' "$out"
  failures=$((failures + 1))
fi

# THE LIVE spectre/specs/ TREE IS COVERED. Its own positive counterpart to the
# frozen-tree case above: a file placed under spectre/specs/ with no budget
# row fails exactly the way any other covered-but-undeclared file does. This
# is the only case in the suite that would notice spectre/specs/ silently
# falling out of OWNED_CORPUS_SCOPE_DIRS — every other case that touches that
# root (mkroot's own skeleton, or a plain directory-existence check) is
# satisfied by an EMPTY spectre/specs/, which today's real spectre/specs/
# actually is (a bare .gitkeep, no budget rows of its own yet). Mutation-
# proved below, in the report: deleting spectre/specs from the scope array
# turns this case's exit 1 into exit 0, and restoring it turns it back.
mkroot "$FIX/live-spec-tree"
mkdir -p "$FIX/live-spec-tree/spectre/specs"
printf 'x\n' > "$FIX/live-spec-tree/skills/myflow-contracts/build-green.md"
printf 'x\n' > "$FIX/live-spec-tree/spectre/specs/some-capability.md"
expect 'a file under the live spectre/specs/ tree is covered' 1 "$FIX/live-spec-tree"
out="$(CHECK_CONTRACT_BUDGET_ROOT="$FIX/live-spec-tree" "$GUARD" 2>&1 || true)"
if printf '%s' "$out" | grep -q '^spectre/specs/some-capability\.md: no budget declared'; then
  printf 'ok   the live spec-tree file is named\n'
else
  printf 'FAIL the live spec-tree file was not named: %s\n' "$out"
  failures=$((failures + 1))
fi

# A DIRECTORY TOTAL IS NOT A SUBSTITUTE FOR PER-FILE ROWS. Two real covered
# files sharing one directory (skills/myflow-contracts/), deliberately picked
# for a huge budget gap between them: build-green.md a few hundred bytes over
# its own (small) row, finish-contract.md a two-byte stub next to its own row
# — more than ten times larger. Their combined size is far below any
# plausible total for the two rows summed, so a guard that resolved
# skills/myflow-contracts/ to one combined directory budget would report both
# clean. Replacing budget_for's per-file lookup with a directory-prefix match
# does exactly that: this case goes from exit 1 to exit 0 under that
# mutation, while the guard as written names the one file that grew.
mkroot "$FIX/dirtotal"
build_green_budget="$(budget_row_bytes 'skills/myflow-contracts/build-green.md')"
head -c "$((build_green_budget + 500))" /dev/zero | tr '\0' 'x' \
  > "$FIX/dirtotal/skills/myflow-contracts/build-green.md"
printf 'x\n' > "$FIX/dirtotal/skills/myflow-contracts/finish-contract.md"
expect 'a large sibling does not absorb a file over its own row' 1 "$FIX/dirtotal"
out="$(CHECK_CONTRACT_BUDGET_ROOT="$FIX/dirtotal" "$GUARD" 2>&1 || true)"
if printf '%s' "$out" | grep -q '^skills/myflow-contracts/build-green\.md: ' \
  && printf '%s' "$out" | grep -q '^BUDGET-FAIL: 1 file(s) over budget or undeclared$'; then
  printf 'ok   exactly the file over its own row is reported\n'
else
  printf 'FAIL directory-total substitution not ruled out: %s\n' "$out"
  failures=$((failures + 1))
fi

# An undeclared .md at the repository root fails. The root is a covered
# location and is scanned non-recursively.
mkroot "$FIX/root-undeclared"
printf 'x\n' > "$FIX/root-undeclared/skills/myflow-contracts/build-green.md"
printf 'x\n' > "$FIX/root-undeclared/NOTES.md"
expect 'an undeclared root .md fails' 1 "$FIX/root-undeclared"
out="$(CHECK_CONTRACT_BUDGET_ROOT="$FIX/root-undeclared" "$GUARD" 2>&1 || true)"
if printf '%s' "$out" | grep -q '^NOTES\.md: no budget declared'; then
  printf 'ok   the undeclared root file is named\n'
else
  printf 'FAIL the undeclared root file was not named: %s\n' "$out"
  failures=$((failures + 1))
fi

# A file under an excluded tree is not covered, however large and however
# undeclared. The exclusions are structural — a path component, never a list of
# filenames — so these fixtures use ordinary names inside excluded directories.
# The verdict COUNT is asserted, not just the exit code: a guard that covered
# these two files would fail on them for having no row, but a guard that merely
# counted them and found rows would pass with the count inflated, and only the
# count tells the two apart. Removing the exclusion from
# scripts/lib/owned-corpus.sh takes this case from exit 0 to exit 1.
mkroot "$FIX/excluded"
mkdir -p "$FIX/excluded/skills/vendor/node_modules/pkg" \
  "$FIX/excluded/skills/scratch/.superpowers"
printf 'x\n' > "$FIX/excluded/skills/myflow-contracts/build-green.md"
head -c 300000 /dev/zero | tr '\0' 'x' \
  > "$FIX/excluded/skills/vendor/node_modules/pkg/huge.md"
head -c 300000 /dev/zero | tr '\0' 'x' \
  > "$FIX/excluded/skills/scratch/.superpowers/notes.md"
expect 'a file under an excluded tree is not covered' 0 "$FIX/excluded"
out="$(CHECK_CONTRACT_BUDGET_ROOT="$FIX/excluded" "$GUARD" 2>&1 || true)"
if printf '%s' "$out" | grep -q '^BUDGET-OK: 1 owned Markdown file(s) within budget$'; then
  printf 'ok   an excluded tree contributes nothing to the count\n'
else
  printf 'FAIL an excluded tree was counted: %s\n' "$out"
  failures=$((failures + 1))
fi

# A symlink inside an excluded tree is not a violation either. The symlink sweep
# runs over the same scope roots as the corpus and applies the same exclusion
# predicate; skipping that predicate would make a vendored tree's own symlinks
# fail this repository's ratchet.
mkroot "$FIX/excluded-link"
printf 'x\n' > "$FIX/excluded-link/skills/myflow-contracts/build-green.md"
mkdir -p "$FIX/excluded-link/skills/vendor/node_modules/pkg"
ln -s /nonexistent/target "$FIX/excluded-link/skills/vendor/node_modules/pkg/ghost.md"
expect 'a symlink inside an excluded tree is not a violation' 0 "$FIX/excluded-link"

if [ "$failures" -ne 0 ]; then
  printf '%d failure(s)\n' "$failures"
  exit 1
fi
printf 'all checks passed\n'
