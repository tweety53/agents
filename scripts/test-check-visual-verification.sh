#!/usr/bin/env bash
# Assertion harness for check-visual-verification.sh.
#
# Builds throwaway fixture trees under TMPDIR and points the guard at each,
# invoking the REAL scripts/check-visual-verification.sh as a subprocess
# against REAL files on disk — never a copy of its logic, and never a
# hand-written expected-output string asserted without having run the guard
# to produce it.
#
# THE ORIGIN-MISMATCH AND NO-ORIGIN CASES USE A REAL `git init` CHECKOUT WITH
# A REAL `origin` REMOTE SET VIA `git remote add` (see make_git_checkout
# below) — never a stubbed `git` on PATH and never a hand-built value, per
# this task's own instruction: the shape the guard's `git remote get-url
# origin` call gets from a real checkout is not the shape a hand-written
# fixture would produce, so only a real checkout proves the finding fires on
# the real boundary.
#
# ONE CASE (the "any other reason" exit-2 branch below) DOES use a stub git
# on PATH, and says so at its own definition: measured against real git
# 2.50.1, every corruption tried (bad config syntax, permission-denied
# config, a duplicate remote.origin.url) either left `git remote get-url
# origin` succeeding or failed identically at the earlier `git rev-parse
# --git-dir` checkout-validity check, so there is no real fixture that
# reaches this branch with a checkout this guard has already accepted as
# valid. The stub is the one exception, isolated to that one case.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-visual-verification.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# Every case leaves one sandbox directory behind, removed on exit including
# on a failed assertion. An indexed array, not a space-separated string:
# mktemp paths under TMPDIR may contain spaces.
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
  ROOT="$(mktemp -d "${TMPDIR:-/tmp}/check-visual-verification-test.XXXXXX")"
  DIRS+=("$ROOT")
  mkdir -p "$ROOT/.flow"
  CFG="$ROOT/.flow/project.md"
}

write_cfg() {
  printf '%s\n' "$1" > "$CFG"
}

# write_cfg_bom — identical to write_cfg, but a UTF-8 BOM (EF BB BF) is
# prepended as the file's literal first three bytes, exactly as a BOM sits
# in a real file some editors and Windows tooling emit by default.
write_cfg_bom() {
  printf '\xef\xbb\xbf%s\n' "$1" > "$CFG"
}

# run_guard -> sets RC and OUT, running the real guard against $ROOT.
run_guard() {
  set +e
  OUT="$("$GUARD" "$ROOT" 2>&1)"
  RC=$?
  set -e
}

# make_git_checkout <dir> [<origin-url>] — a REAL `git init` checkout, with a
# REAL `origin` remote set via `git remote add` when a URL is given.
make_git_checkout() {
  local dir="$1" origin="${2:-}"
  mkdir -p "$dir"
  git -C "$dir" init -q
  if [ -n "$origin" ]; then
    git -C "$dir" remote add origin "$origin"
  fi
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
| `screenshots` | `stats/web/tests/visual/baseline.spec.ts-snapshots` |

| Command | Runs |
|---------|------|
| `verify` | `npm run test:visual` |
| `capture` | `npx playwright test <spec>` |'

# ===========================================================================
# Case 1: no .flow/project.md at all — exit 0, clean.
# ===========================================================================
new_root
rm -f "$CFG"
run_guard
assert_rc "case 1" 0
assert_out_contains "case 1" "no .flow/project.md"

# ===========================================================================
# Case 2: .flow/project.md exists but declares no `## visual verification`
# section — exit 0, clean.
# ===========================================================================
new_root
write_cfg "# Project

## run

echo hi
"
run_guard
assert_rc "case 2" 0
assert_out_contains "case 2" "declares no"

# ===========================================================================
# Case 3: a minimal, fully valid section — exit 0.
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER"
run_guard
assert_rc "case 3" 0
assert_out_contains "case 3" "VISUAL-OK"

# ===========================================================================
# Case 4: `ui paths` absent — exit 1, names the row.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`screenshots\` | \`stats/web/tests/visual/baseline.spec.ts-snapshots\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard
assert_rc "case 4" 1
assert_out_contains "case 4" "ui paths"

# ===========================================================================
# Case 5: `screenshots` present but empty — exit 1, names the row.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |
| \`screenshots\` | |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard
assert_rc "case 5" 1
assert_out_contains "case 5" "screenshots"

# ===========================================================================
# Case 6: `verify` absent — exit 1, names the row.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |
| \`screenshots\` | \`stats/web/tests/visual/baseline.spec.ts-snapshots\` |

| Command | Runs |
|---------|------|
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard
assert_rc "case 6" 1
assert_out_contains "case 6" "verify"

# ===========================================================================
# Case 7: `capture` present but empty — exit 1, names the row.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |
| \`screenshots\` | \`stats/web/tests/visual/baseline.spec.ts-snapshots\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | |"
run_guard
assert_rc "case 7" 1
assert_out_contains "case 7" "capture"

# ===========================================================================
# Case 8: `regression repo` present with no `regression checkout` — exit 1.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |
| \`screenshots\` | \`stats/web/tests/visual/baseline.spec.ts-snapshots\` |
| \`regression repo\` | \`git@github.com:tweety53/gymie-playwright.git\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard
assert_rc "case 8" 1
assert_out_contains "case 8" "regression repo"
assert_out_contains "case 8" "regression checkout"

# ===========================================================================
# Case 9: `regression checkout` present with no `regression repo` — exit 1
# (the reverse pairing).
# ===========================================================================
new_root
CHECKOUT_DIR="$ROOT/checkout"
DIRS+=("$CHECKOUT_DIR")
make_git_checkout "$CHECKOUT_DIR"
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |
| \`screenshots\` | \`.\` |
| \`regression checkout\` | \`$CHECKOUT_DIR\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard
assert_rc "case 9" 1
assert_out_contains "case 9" "regression checkout"

# ===========================================================================
# Case 10: `regression checkout` names a path that does not exist — exit 1.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |
| \`screenshots\` | \`.\` |
| \`regression checkout\` | \`$ROOT/does-not-exist\` |
| \`regression repo\` | \`git@github.com:tweety53/gymie-playwright.git\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard
assert_rc "case 10" 1
assert_out_contains "case 10" "not an existing directory"

# ===========================================================================
# Case 11: `regression checkout` names an existing directory that is not a
# git checkout — exit 1.
# ===========================================================================
new_root
PLAIN_DIR="$ROOT/plain-dir"
mkdir -p "$PLAIN_DIR"
DIRS+=("$PLAIN_DIR")
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |
| \`screenshots\` | \`.\` |
| \`regression checkout\` | \`$PLAIN_DIR\` |
| \`regression repo\` | \`git@github.com:tweety53/gymie-playwright.git\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard
assert_rc "case 11" 1
assert_out_contains "case 11" "not a git checkout"

# ===========================================================================
# Case 12: `push to default branch` is no longer part of the contract — task
# 14 dropped it once a panel slot demonstrated that both sides of its
# origin-equality check live in this same pull-request-editable file. A row
# naming it is now an unrecognized Setting, reported rather than silently
# honoured, even when the checkout it names is real and its origin matches
# `regression repo`.
# ===========================================================================
new_root
CHECKOUT_DIR="$ROOT/checkout"
DIRS+=("$CHECKOUT_DIR")
make_git_checkout "$CHECKOUT_DIR" "git@github.com:tweety53/gymie-playwright.git"
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |
| \`screenshots\` | \`.\` |
| \`regression checkout\` | \`$CHECKOUT_DIR\` |
| \`regression repo\` | \`git@github.com:tweety53/gymie-playwright.git\` |
| \`push to default branch\` | \`allowed\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard
assert_rc "case 12" 1
assert_out_contains "case 12" "push to default branch"
assert_out_contains "case 12" "vocabulary is closed"

# ===========================================================================
# Case 13 (REPRODUCE, DON'T READ): `regression repo` is an identity
# assertion now, checked whenever `regression checkout` and `regression
# repo` are both declared — never gated on a push permission the contract no
# longer carries. A REAL checkout's REAL origin (set via `git remote add`)
# does not equal the declared `regression repo`, and no push setting is
# declared at all — exit 1, reports BOTH URLs.
# ===========================================================================
new_root
CHECKOUT_DIR="$ROOT/checkout"
DIRS+=("$CHECKOUT_DIR")
make_git_checkout "$CHECKOUT_DIR" "git@github.com:someone-else/decoy.git"
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |
| \`screenshots\` | \`.\` |
| \`regression checkout\` | \`$CHECKOUT_DIR\` |
| \`regression repo\` | \`git@github.com:tweety53/gymie-playwright.git\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard
assert_rc "case 13" 1
assert_out_contains "case 13" "git@github.com:someone-else/decoy.git"
assert_out_contains "case 13" "git@github.com:tweety53/gymie-playwright.git"

# ===========================================================================
# Case 14 (REPRODUCE, DON'T READ): a REAL checkout's REAL origin DOES equal
# the declared `regression repo`, with no push setting declared at all —
# exit 0, clean.
# ===========================================================================
new_root
CHECKOUT_DIR="$ROOT/checkout"
DIRS+=("$CHECKOUT_DIR")
make_git_checkout "$CHECKOUT_DIR" "git@github.com:tweety53/gymie-playwright.git"
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |
| \`screenshots\` | \`.\` |
| \`regression checkout\` | \`$CHECKOUT_DIR\` |
| \`regression repo\` | \`git@github.com:tweety53/gymie-playwright.git\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard
assert_rc "case 14" 0
assert_out_contains "case 14" "VISUAL-OK"

# ===========================================================================
# Case 15 (REPRODUCE, DON'T READ): `regression checkout` and `regression
# repo` are both declared, no push setting at all, and the REAL checkout
# carries no `origin` remote — exit 1, a finding rather than a silent pass.
# ===========================================================================
new_root
CHECKOUT_DIR="$ROOT/checkout"
DIRS+=("$CHECKOUT_DIR")
make_git_checkout "$CHECKOUT_DIR"
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |
| \`screenshots\` | \`.\` |
| \`regression checkout\` | \`$CHECKOUT_DIR\` |
| \`regression repo\` | \`git@github.com:tweety53/gymie-playwright.git\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard
assert_rc "case 15" 1
assert_out_contains "case 15" "origin"

# ===========================================================================
# Case 16: the project root does not exist — exit 2.
# ===========================================================================
ROOT="/nonexistent/$$-$RANDOM"
run_guard
assert_rc "case 16" 2

# ===========================================================================
# Case 17: `.flow/project.md` is a directory, not a regular file — exit 2.
# ===========================================================================
new_root
rm -f "$CFG"
mkdir -p "$CFG"
run_guard
assert_rc "case 17" 2

# ===========================================================================
# Case 18: the file declares `## visual verification` twice — exit 1, a
# single violation naming the duplication, and neither section validated.
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER

$MINIMAL_HEADER"
run_guard
assert_rc "case 18" 1
assert_out_contains "case 18" "2"

# ===========================================================================
# Case 19: a duplicate `ui paths` row inside the same settings table —
# exit 1.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |
| \`ui paths\` | \`stats/web/other/**\` |
| \`screenshots\` | \`.\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard
assert_rc "case 19" 1
assert_out_contains "case 19" "ui paths"

# ===========================================================================
# Case 20: an unrecognized table header under the section — exit 1, never
# silently skipped.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Values |
|---------|--------|
| \`ui paths\` | \`stats/web/src/**\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard
assert_rc "case 20" 1
assert_out_contains "case 20" "header"

# ===========================================================================
# Case 21 (a stub git on PATH — see the header note above for why the real
# boundary cannot reach this branch): `git remote get-url origin` fails for
# a reason OTHER than "no such remote" — exit 2, never read as a finding.
#
# THE CHECKOUT'S OWN DIRECTORY NAME CARRIES RAW ANSI ESCAPES, so this case
# also pins that the exit-2 message's interpolated cells ($CHECKOUT_VAL and
# $GIT_ERR) are sanitized before they reach stderr, exactly as every other
# message in the guard is. Found missing this in review of 92d050a: the
# `\x1b[2K\x1b[1m…\x1b[0m` sequence below erases the current terminal line
# and turns the next one bold, which — unescaped — can make an exit-2 run's
# own stderr render as a convincing forged `VISUAL-OK:` line. The escaped
# form (`\x1b[2K`, literal backslash-x-1-b) must appear in the captured
# output and the raw ESC byte must not.
# ===========================================================================
new_root
ESC=$'\x1b'
FORGED_NAME="${ESC}[2K${ESC}[1mVISUAL-OK: forged${ESC}[0m"
CHECKOUT_DIR="$ROOT/$FORGED_NAME"
DIRS+=("$CHECKOUT_DIR")
make_git_checkout "$CHECKOUT_DIR" "git@github.com:tweety53/gymie-playwright.git"
FAKE_BIN="$(mktemp -d "${TMPDIR:-/tmp}/check-visual-verification-fakebin.XXXXXX")"
DIRS+=("$FAKE_BIN")
REAL_GIT="$(command -v git)"
cat > "$FAKE_BIN/git" <<EOF
#!/usr/bin/env bash
case " \$* " in
  *" remote get-url origin "*)
    echo "fatal: unable to auto-detect email address" >&2
    exit 128
    ;;
esac
exec "$REAL_GIT" "\$@"
EOF
chmod +x "$FAKE_BIN/git"
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |
| \`screenshots\` | \`.\` |
| \`regression checkout\` | \`$CHECKOUT_DIR\` |
| \`regression repo\` | \`git@github.com:tweety53/gymie-playwright.git\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
set +e
OUT="$(PATH="$FAKE_BIN:$PATH" "$GUARD" "$ROOT" 2>&1)"
RC=$?
set -e
assert_rc "case 21" 2
assert_out_contains "case 21" '\x1b[2K'
case "$OUT" in
  *"$ESC"*) fail "case 21: a raw ESC byte reached the captured output unescaped: $(printf '%s' "$OUT" | cat -v)" ;;
  *) pass "case 21: no raw ESC byte reached the captured output" ;;
esac

# ===========================================================================
# Case 22 (REPRODUCE, DON'T READ): a hostile ambient `GIT_DIR`, set to a REAL
# checkout whose REAL `origin` equals the declared `regression repo`, must
# not let a `regression checkout` that is not a git repository at all pass
# as one. Against the un-fixed guard this prints VISUAL-OK, exit 0, with
# only GIT_DIR added — confirmed RED before git_clean was introduced in this
# guard, GREEN after.
# ===========================================================================
new_root
AMBIENT_DIR="$ROOT/ambient"
DIRS+=("$AMBIENT_DIR")
make_git_checkout "$AMBIENT_DIR" "git@github.com:tweety53/gymie-playwright.git"
NOT_GIT_DIR="$ROOT/not-a-git-dir"
mkdir -p "$NOT_GIT_DIR"
DIRS+=("$NOT_GIT_DIR")
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |
| \`screenshots\` | \`.\` |
| \`regression checkout\` | \`$NOT_GIT_DIR\` |
| \`regression repo\` | \`git@github.com:tweety53/gymie-playwright.git\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
set +e
OUT="$(GIT_DIR="$AMBIENT_DIR/.git" "$GUARD" "$ROOT" 2>&1)"
RC=$?
set -e
assert_rc "case 22" 1
assert_out_contains "case 22" "not a git checkout"
case "$OUT" in
  *"VISUAL-OK"*) fail "case 22: an ambient GIT_DIR bypassed the checkout-validity check and printed VISUAL-OK" ;;
  *) pass "case 22: ambient GIT_DIR did not bypass the checkout-validity check" ;;
esac

# ===========================================================================
# Case 23: a `###` subheading inside the `## visual verification` section
# does not end the section — a second settings-table block declared below
# it, and the commands table below that, are still read. exit 0, clean.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |

### Notes

| Setting | Value |
|---------|-------|
| \`screenshots\` | \`stats/web/tests/visual/baseline.spec.ts-snapshots\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard
assert_rc "case 23" 0
assert_out_contains "case 23" "VISUAL-OK"

# ===========================================================================
# Case 24 (REPRODUCE, DON'T READ): `.flow/project.md` begins with a UTF-8
# BOM, with `## visual verification` as the file's own first line. A BOM
# must not make a present section read as absent — exit 0, VISUAL-OK, not
# "declares no section".
# ===========================================================================
new_root
write_cfg_bom "$MINIMAL_HEADER"
run_guard
assert_rc "case 24" 0
assert_out_contains "case 24" "the \`## visual verification\` section validated"
assert_out_not_contains "case 24" "declares no"

# ===========================================================================
# Case 25: CONFIRMATION, NOT A REPRODUCTION — task 16's finding c) named a
# validator gap: `push to default branch: allowed` with NEITHER `regression
# checkout` nor `regression repo` declared used to pass this validator
# clean, because the near-miss/pairing checks below only fire once a `push
# to default branch` row is recognized at all. Task 14 already closed this
# by dropping `push to default branch` from the closed Setting vocabulary
# entirely, so ANY row naming it — paired or not — is now an unrecognized
# Setting on its own, before the pairing logic is ever reached. This case
# pins that the exact original shape is moot on the code as it stands, not
# that it was newly fixed here.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |
| \`screenshots\` | \`.\` |
| \`push to default branch\` | \`allowed\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard
assert_rc "case 25" 1
assert_out_contains "case 25" "push to default branch"
assert_out_contains "case 25" "vocabulary is closed"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
