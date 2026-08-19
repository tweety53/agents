#!/usr/bin/env bash
# test-check-contract-budget.sh — harness for check-contract-budget.sh.
#
# Every case runs against a sandboxed fixture tree under mktemp, reached through
# CHECK_CONTRACT_BUDGET_ROOT, so the harness never depends on the real
# repository's current sizes — a harness that asserted against those would go red
# every time a contract legitimately changed.
set -euo pipefail

GUARD="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-contract-budget.sh"
[ -x "$GUARD" ] || { printf 'guard not executable: %s\n' "$GUARD" >&2; exit 2; }

# budget_row_bytes <path-relative-to-repo-root> [table-file] — read a row
# straight out of the guard's own budgets() table (a plain heredoc in $GUARD,
# never executed here; the second argument defaults to $GUARD and exists only
# so the F18/F22 cases below can exercise the duplicate-row and
# unterminated-row branches against a scratch table without touching the
# guard's real one). A fixture sized off this is over/under its real budget
# BY CONSTRUCTION: it stays correct however the row changes. A fixture sized
# off a magic number copied from the table instead goes stale the next time
# that row rises — the defect the guard's own fix round shipped (F15).
#
# $GUARD is a shell script, not a bare table — budgets() wraps its rows in a
# single-quoted heredoc so raising a budget is a diff of the table alone.
# Reading the whole script, as this helper used to, treats any stray line
# elsewhere in it whose first two whitespace-separated fields happen to look
# like a covered path and a number as a row of its own; a live example broke
# this exact harness (F21). So when $table is $GUARD, only the region
# between budgets()'s heredoc delimiters is fed to the comparison below; a
# caller-supplied bare table (F18, F22) has no such wrapper and is read
# whole, unchanged.
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
  done < <(
    if [ "$table" = "$GUARD" ]; then
      sed -n "/<<'EOF'\$/,/^EOF\$/p" "$table" | sed '1d;$d'
    else
      cat "$table"
    fi
  )
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

# A file well under its declared budget.
mkdir -p "$FIX/within/skills/myflow-contracts"
printf 'x\n' > "$FIX/within/skills/myflow-contracts/build-green.md"
expect 'a file well under budget passes' 0 "$FIX/within"

# The same file pushed past its declared budget, read straight from the
# guard's own table so this fixture is over budget by construction rather
# than by a number that must stay above a row that can rise (see F15).
mkdir -p "$FIX/over/skills/myflow-contracts"
build_green_budget="$(budget_row_bytes 'skills/myflow-contracts/build-green.md')"
head -c "$((build_green_budget + 500))" /dev/zero | tr '\0' 'x' \
  > "$FIX/over/skills/myflow-contracts/build-green.md"
expect 'a file over budget fails' 1 "$FIX/over"

# A contract added later with no row in budgets(). This is the case that stops a
# new file escaping the ratchet silently.
mkdir -p "$FIX/undeclared/skills/myflow-contracts"
printf 'x\n' > "$FIX/undeclared/skills/myflow-contracts/brand-new-contract.md"
expect 'a file with no budget row fails' 1 "$FIX/undeclared"

# A contracts directory that exists but holds no .md at all. Without the glob
# guard this reports clean, which is the false-clean this check exists to catch.
mkdir -p "$FIX/empty/skills/myflow-contracts"
expect 'an empty contracts directory cannot be answered' 2 "$FIX/empty"

# No contracts directory under an otherwise real root.
mkdir -p "$FIX/nodir"
expect 'a missing contracts directory cannot be answered' 2 "$FIX/nodir"

# A root that does not exist at all.
expect 'a nonexistent root cannot be answered' 2 "$FIX/absent"

# A file the glob matched but cannot be READ. Every other exit-2 case here is a
# directory-shape problem; this is the only file-level one, and without it a
# regression in the wc guard ships silently — the guard would abort with wc's own
# status of 1, which this contract reserves for "over budget or undeclared", and
# send a caller off to edit budgets() when the guard could not read the file.
mkdir -p "$FIX/unreadable/skills/myflow-contracts"
printf 'x\n' > "$FIX/unreadable/skills/myflow-contracts/build-green.md"
chmod 000 "$FIX/unreadable/skills/myflow-contracts/build-green.md"
if [ "$(id -u)" = "0" ]; then
  printf 'skip an unreadable file cannot be answered (running as root reads it anyway)\n'
else
  expect 'an unreadable file cannot be answered' 2 "$FIX/unreadable"
fi
chmod 644 "$FIX/unreadable/skills/myflow-contracts/build-green.md"

# A dangling *.md symlink beside real files. Exit 2 is correct — the guard truly
# cannot answer for that entry — but WHICH 2 matters: `-e` alone is false for a
# broken symlink, so the empty-glob sentinel used to fire and report "no .md
# files in <dir>" while a real contract file sat beside it unexamined. A guard
# that reports the directory is empty when it is not is the false report this
# repository least tolerates, so this case asserts the message, not just the code.
mkdir -p "$FIX/dangling/skills/myflow-contracts"
printf 'x\n' > "$FIX/dangling/skills/myflow-contracts/build-green.md"
ln -s /nonexistent/target "$FIX/dangling/skills/myflow-contracts/aaa-ghost.md"
# aaa- sorts first, so the glob reaches the broken link before the real file.
expect 'a dangling .md symlink is a per-file violation' 1 "$FIX/dangling"
out="$(CHECK_CONTRACT_BUDGET_ROOT="$FIX/dangling" "$GUARD" 2>&1 || true)"
if printf '%s' "$out" | grep -q 'is a symlink'; then
  printf 'ok   a dangling symlink is named as a symlink, not as an empty directory\n'
else
  printf 'FAIL unexpected message for a dangling symlink: %s\n' "$out"
  failures=$((failures + 1))
fi

# A symlink pointing OUTSIDE the tree must be refused before it is sized —
# otherwise the guard is an existence-and-size oracle for arbitrary paths, and a
# target whose name matches a budget row has its byte count printed.
mkdir -p "$FIX/escape/skills/myflow-contracts"
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
mkdir -p "$FIX/newline/skills/myflow-contracts"
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
if printf '%s' "$out" | grep -q '^BUDGET-OK: 1 contract file(s) within budget$'; then
  printf 'ok   the verdict names how many files were checked\n'
else
  printf 'FAIL verdict did not carry the count: %s\n' "$out"
  failures=$((failures + 1))
fi

# --- Widened scope: skills/*/SKILL.md and skills/*/SKILL-rationale.md ------
#
# These three cases exercise the glob widening from task 5.1. Each fixture
# still needs a real, in-budget skills/myflow-contracts/*.md file so the
# contracts-directory requirement is satisfied — the point of these cases is
# the skills/ glob, not the contracts one.
#
# The paths below are deliberately real repository skill paths. Each fixture's
# size is read from the guard's own budgets() table via budget_row_bytes rather
# than copied as a number, so it stays over/under its real row by construction
# as that row moves — the coupling that broke in F15.

# A skills/<x>/SKILL.md that exceeds its declared budget. skills/myflow-status
# only has a SKILL.md, no rationale sibling, which keeps this fixture simple.
mkdir -p "$FIX/skill-over/skills/myflow-contracts" "$FIX/skill-over/skills/myflow-status"
printf 'x\n' > "$FIX/skill-over/skills/myflow-contracts/build-green.md"
myflow_status_budget="$(budget_row_bytes 'skills/myflow-status/SKILL.md')"
head -c "$((myflow_status_budget + 1000))" /dev/zero | tr '\0' 'x' \
  > "$FIX/skill-over/skills/myflow-status/SKILL.md"
expect 'a skills/<x>/SKILL.md over budget fails' 1 "$FIX/skill-over"

# A skills/<x>/SKILL.md under a skill name that has no row in budgets() at all.
mkdir -p "$FIX/skill-undeclared/skills/myflow-contracts" \
  "$FIX/skill-undeclared/skills/mystery-skill"
printf 'x\n' > "$FIX/skill-undeclared/skills/myflow-contracts/build-green.md"
printf 'x\n' > "$FIX/skill-undeclared/skills/mystery-skill/SKILL.md"
expect 'a skills/<x>/SKILL.md with no budget row fails' 1 "$FIX/skill-undeclared"

# Two skills whose SKILL.md files carry different budgets in the real table —
# skills/openspec-explore/SKILL.md (a small budget) and skills/myflow-do/SKILL.md (a
# much larger one). skills/myflow-do/SKILL.md is sized to the midpoint between
# the two real rows, read from the guard itself: above openspec-explore's row,
# below myflow-do's row, by construction rather than by a number that must stay
# between two rows that can each move independently. If the table were still
# keyed on the bare basename "SKILL.md" — the collision this task fixes —
# every skills/*/SKILL.md would be checked against one shared row, and this
# fixture would fail against the small budget that basename collision would
# hand it. Keyed on the path relative to the repository root, each file is
# checked against its own row and both pass.
mkdir -p "$FIX/skill-distinct/skills/myflow-contracts" \
  "$FIX/skill-distinct/skills/openspec-explore" "$FIX/skill-distinct/skills/myflow-do"
printf 'x\n' > "$FIX/skill-distinct/skills/myflow-contracts/build-green.md"
printf 'x\n' > "$FIX/skill-distinct/skills/openspec-explore/SKILL.md"
openspec_explore_budget="$(budget_row_bytes 'skills/openspec-explore/SKILL.md')"
myflow_do_budget="$(budget_row_bytes 'skills/myflow-do/SKILL.md')"

# The fixture only discriminates the basename-collision bug when mid_size
# lands STRICTLY BETWEEN the two rows. If the rows ever converge, integer
# division makes mid_size collapse onto openspec_explore_budget and the case
# below would keep reporting "ok" even with the collision bug back in place —
# F15's own defect class (a fixture that stops discriminating when its
# premise erodes), reintroduced by the round that fixed F15 (F16). Assert the
# premise rather than assume it, and fail the whole harness loudly, naming
# both rows, when it does not hold: a fixture that cannot be constructed to
# discriminate must say so, never quietly pass.
if [ "$((myflow_do_budget - openspec_explore_budget))" -lt 2 ]; then
  printf 'cannot construct a discriminating fixture: skills/openspec-explore/SKILL.md is %s bytes and skills/myflow-do/SKILL.md is %s bytes — too close together for a strict midpoint\n' \
    "$openspec_explore_budget" "$myflow_do_budget" >&2
  exit 2
fi
mid_size=$((openspec_explore_budget + (myflow_do_budget - openspec_explore_budget) / 2))
head -c "$mid_size" /dev/zero | tr '\0' 'x' > "$FIX/skill-distinct/skills/myflow-do/SKILL.md"
expect 'two skills with different budgets are each checked against their own row' \
  0 "$FIX/skill-distinct"

# A SKILL-rationale.md over budget. Without this the rationale half of the widened
# scope has no coverage at all: mutating that glob to a typo left all other cases
# green while every SKILL-rationale.md in the repo silently stopped being guarded —
# exactly the "escapes the ratchet silently" failure this guard exists to prevent.
# Sized from the guard's own row (see F15: a hardcoded 20000 here fell under the
# row when a fix round raised it, and this case stopped catching an over-budget
# file at all).
mkdir -p "$FIX/rationale/skills/myflow-do" "$FIX/rationale/skills/myflow-contracts"
printf 'x\n' > "$FIX/rationale/skills/myflow-contracts/build-green.md"
myflow_do_rationale_budget="$(budget_row_bytes 'skills/myflow-do/SKILL-rationale.md')"
head -c "$((myflow_do_rationale_budget + 1000))" /dev/zero | tr '\0' 'x' \
  > "$FIX/rationale/skills/myflow-do/SKILL-rationale.md"
expect 'a SKILL-rationale.md over budget fails' 1 "$FIX/rationale"

# A SKILL-rationale.md with no budget row — the undeclared case for the same glob.
mkdir -p "$FIX/rationale-new/skills/brand-new" "$FIX/rationale-new/skills/myflow-contracts"
printf 'x\n' > "$FIX/rationale-new/skills/myflow-contracts/build-green.md"
printf 'x\n' > "$FIX/rationale-new/skills/brand-new/SKILL-rationale.md"
expect 'a SKILL-rationale.md with no budget row fails' 1 "$FIX/rationale-new"

# skills/myflow-contracts/SKILL.md is matched by BOTH globs — the required contracts
# loop and the skills/*/SKILL.md loop. The prefix skip is what stops it being counted
# twice. Deleting that skip leaves every other case green while the verdict's count
# silently inflates, so this case asserts the COUNT, not just the exit code.
mkdir -p "$FIX/nodouble/skills/myflow-contracts"
printf 'x\n' > "$FIX/nodouble/skills/myflow-contracts/build-green.md"
printf 'x\n' > "$FIX/nodouble/skills/myflow-contracts/SKILL.md"
expect 'a contracts-dir SKILL.md passes' 0 "$FIX/nodouble"
out="$(CHECK_CONTRACT_BUDGET_ROOT="$FIX/nodouble" "$GUARD")"
if printf '%s' "$out" | grep -q '^BUDGET-OK: 2 contract file(s) within budget$'; then
  printf 'ok   a contracts-dir SKILL.md is counted once, not twice\n'
else
  printf 'FAIL double-count: %s\n' "$out"
  failures=$((failures + 1))
fi

if [ "$failures" -ne 0 ]; then
  printf '%d failure(s)\n' "$failures"
  exit 1
fi
printf 'all checks passed\n'
