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

# A file well under its declared budget.
mkdir -p "$FIX/within/skills/myflow-contracts"
printf 'x\n' > "$FIX/within/skills/myflow-contracts/build-green.md"
expect 'a file well under budget passes' 0 "$FIX/within"

# The same file pushed past its 4557-byte budget. 5000 bytes is over it and
# under the next-largest budget in the table, so this asserts the comparison
# rather than merely a large number.
mkdir -p "$FIX/over/skills/myflow-contracts"
head -c 5000 /dev/zero | tr '\0' 'x' > "$FIX/over/skills/myflow-contracts/build-green.md"
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

if [ "$failures" -ne 0 ]; then
  printf '%d failure(s)\n' "$failures"
  exit 1
fi
printf 'all checks passed\n'
