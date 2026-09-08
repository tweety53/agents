#!/usr/bin/env bash
# Assertion harness for check-baseline-fresh.sh / check-baseline-fresh.py.
#
# Modeled on test-check-plan-shape.sh's fixture-driven pattern: fixtures
# live under mktemp -d, the guard is invoked via a thin run_guard helper
# that captures RC/OUT, and every case ends with an explicit pass/fail
# assertion. Runs the REAL scripts/check-baseline-fresh.sh against REAL
# fixture files on disk — never a copy of its logic. Two MUTATION cases
# disable the guard's stale-verdict and absent-verdict checks (via sed
# against lines tagged `# STALE-VERDICT` and `# ABSENT-VERDICT`) and assert
# the tagged check's own MESSAGE vanishes from the mutant's output — not
# rc 0, which the guard's second layer (the unreadable-mtime verdict) may
# still legitimately produce for the same fixture. A suite that cannot
# detect the guard's own defect is not a suite.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-baseline-fresh.sh"
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

# run_guard [<changeRoot>] -> sets RC and OUT (combined stdout+stderr)
run_guard() {
  set +e
  OUT="$("$GUARD" "$@" 2>&1)"
  RC=$?
  set -e
}

# new_fixture -> FIXTURE (a git repo whose newest commit sits at epoch
# 1577836800 / 2020-01-01T00:00:00Z), CHANGE (its change root), TASKS_MD
new_fixture() {
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/baseline-fresh-test.XXXXXX")"
  DIRS+=("$FIXTURE")
  git -C "$FIXTURE" init -q
  GIT_COMMITTER_DATE="2020-01-01T00:00:00Z" \
    GIT_AUTHOR_DATE="2020-01-01T00:00:00Z" \
    git -C "$FIXTURE" commit -q --allow-empty -m base
  CHANGE="$FIXTURE/spectre/changes/baseline-fresh-fixture"
  mkdir -p "$CHANGE"
  TASKS_MD="$CHANGE/tasks.md"
}

# write_tasks <path> — a minimal one-task plan carrying a **Baseline:** field
write_tasks() {
  cat > "$1" <<'EOF'
- [ ] 1. Do the thing

**Build:** green
**Files:**
- `src/thing.go`
**Tests:** `none`
**Baseline:** before=10 after=12
**Commit:** `feat(thing): do the thing`
EOF
}

# write_config <path> — a project config declaring one results dir (pass the
# body lines as extra arguments; no arguments declares no key)
write_config() {
  local cfg="$1"
  shift
  {
    printf '# fixture project\n'
    if [ "$#" -gt 0 ]; then
      printf '\n## baseline results dirs\n\n'
      printf '%s\n' "$@"
    fi
  } > "$cfg"
}

# fresh_dir / stale_dir — set a directory's mtime after/before the fixture's
# fixed newest-commit epoch (1577836800)
fresh_dir() { mkdir -p "$1" && touch "$1"; }
stale_dir() { mkdir -p "$1" && touch -t 201901010000 "$1"; }

# case 1 — no **Baseline:** field anywhere -> RC 0, nothing to check
new_fixture
printf -- '- [ ] 1. Do the thing\n' > "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
run_guard "$CHANGE"
[ "$RC" -eq 0 ] || fail "case 1: expected RC 0, got $RC: $OUT"
pass "case 1: no Baseline field -> skip"

# case 2 — Baseline field, no .flow/project.md above the change root -> RC 0
new_fixture
write_tasks "$TASKS_MD"
run_guard "$CHANGE"
[ "$RC" -eq 0 ] || fail "case 2: expected RC 0, got $RC: $OUT"
pass "case 2: no project config -> skip"

# case 3 — Baseline field, config without the key -> RC 0
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md"
run_guard "$CHANGE"
[ "$RC" -eq 0 ] || fail "case 3: expected RC 0, got $RC: $OUT"
pass "case 3: key undeclared -> skip"

# case 4 — declared dir fresh (mtime after the fixed commit epoch) -> RC 0
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
fresh_dir "$FIXTURE/build/test-results"
run_guard "$CHANGE"
[ "$RC" -eq 0 ] || fail "case 4: expected RC 0, got $RC: $OUT"
pass "case 4: fresh dir -> pass"

# case 5 — declared dir absent -> RC 1, message names the dir
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
run_guard "$CHANGE"
[ "$RC" -eq 1 ] || fail "case 5: expected RC 1, got $RC: $OUT"
printf '%s' "$OUT" | grep -q 'build/test-results' || fail "case 5: dir not named: $OUT"
printf '%s' "$OUT" | grep -q 'does not exist' || fail "case 5: absence not stated: $OUT"
pass "case 5: absent dir -> stale"

# case 6 — declared dir stale (mtime 2019 < commit 2020) -> RC 1, both epochs
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
stale_dir "$FIXTURE/build/test-results"
run_guard "$CHANGE"
[ "$RC" -eq 1 ] || fail "case 6: expected RC 1, got $RC: $OUT"
printf '%s' "$OUT" | grep -q 'stale' || fail "case 6: staleness not stated: $OUT"
printf '%s' "$OUT" | grep -q '1577836800' || fail "case 6: commit epoch not named: $OUT"
pass "case 6: stale dir -> stale"

# case 7 — one fresh + one absent dir -> RC 1 naming only the offender
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results' 'dist/test-results'
fresh_dir "$FIXTURE/build/test-results"
run_guard "$CHANGE"
[ "$RC" -eq 1 ] || fail "case 7: expected RC 1, got $RC: $OUT"
printf '%s' "$OUT" | grep -q 'dist/test-results' || fail "case 7: offender not named: $OUT"
printf '%s' "$OUT" | grep -q 'build/test-results' && fail "case 7: fresh dir reported: $OUT"
pass "case 7: only the offender reported"

# case 8 — two stale dirs -> RC 1, both named in the one report
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results' 'dist/test-results'
stale_dir "$FIXTURE/build/test-results"
stale_dir "$FIXTURE/dist/test-results"
run_guard "$CHANGE"
[ "$RC" -eq 1 ] || fail "case 8: expected RC 1, got $RC: $OUT"
printf '%s' "$OUT" | grep -q 'build/test-results' || fail "case 8: first dir not named: $OUT"
printf '%s' "$OUT" | grep -q 'dist/test-results' || fail "case 8: second dir not named: $OUT"
pass "case 8: both offenders in one report"

# case 9 — duplicate key declarations -> RC 2
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
printf '\n## baseline results dirs\n\ndist/test-results\n' >> "$FIXTURE/.flow/project.md"
run_guard "$CHANGE"
[ "$RC" -eq 2 ] || fail "case 9: expected RC 2, got $RC: $OUT"
pass "case 9: duplicate key -> cannot answer"

# case 10 — key declared with an empty body -> RC 2
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md"
printf '\n## baseline results dirs\n' >> "$FIXTURE/.flow/project.md"
run_guard "$CHANGE"
[ "$RC" -eq 2 ] || fail "case 10: expected RC 2, got $RC: $OUT"
pass "case 10: empty key body -> cannot answer"

# case 11 — declared path escaping the project root -> RC 2
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" '../outside-results'
run_guard "$CHANGE"
[ "$RC" -eq 2 ] || fail "case 11: expected RC 2, got $RC: $OUT"
pass "case 11: escaping path -> cannot answer"

# case 12 — a **Baseline:** line inside a fenced example is not a field -> RC 0
new_fixture
FENCE='```'
{
  printf -- '- [ ] 1. Do the thing\n\n'
  printf -- '**Build:** green\n\n'
  printf -- '%s\n' "$FENCE"
  printf -- '**Baseline:** before=1 after=2\n'
  printf -- '%s\n' "$FENCE"
} > "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
run_guard "$CHANGE"
[ "$RC" -eq 0 ] || fail "case 12: expected RC 0, got $RC: $OUT"
pass "case 12: fenced Baseline line ignored"

# case 13 — usage: no arguments -> RC 2
run_guard
[ "$RC" -eq 2 ] || fail "case 13: expected RC 2, got $RC: $OUT"
pass "case 13: usage -> exit 2"

# cases 14–15 — mutations: each verdict depends on its own check. The
# assertion is the house one (test-check-plan-shape.sh's): the tagged
# check's own MESSAGE must vanish from the mutant's output — not rc 0,
# which the guard's second layer (the unreadable-mtime verdict) may still
# legitimately produce for the same fixture.
assert_mutation_removes_finding() {
  local tag="$1" fixture_change="$2" label="$3" needle="$4"
  local mutant_dir out rc
  mutant_dir="$(mktemp -d "${TMPDIR:-/tmp}/baseline-fresh-mutant.XXXXXX")"
  DIRS+=("$mutant_dir")
  mkdir -p "$mutant_dir/lib"
  ln -s "$SCRIPT_DIR/check-task-commit-fields.py" "$mutant_dir/check-task-commit-fields.py"
  ln -s "$SCRIPT_DIR/lib/plan_grammar.py" "$mutant_dir/lib/plan_grammar.py"
  cp "$SCRIPT_DIR/check-baseline-fresh.sh" "$mutant_dir/check-baseline-fresh.sh"
  chmod +x "$mutant_dir/check-baseline-fresh.sh"
  sed "/# ${tag}\$/s/if .*/if False:  # ${tag} (mutated)/" \
    "$SCRIPT_DIR/check-baseline-fresh.py" > "$mutant_dir/check-baseline-fresh.py"
  grep -q "${tag} (mutated)" "$mutant_dir/check-baseline-fresh.py" \
    || { fail "$label: sed edit did not apply"; return 0; }
  set +e
  out="$("$mutant_dir/check-baseline-fresh.sh" "$fixture_change" 2>&1)"
  rc=$?
  set -e
  printf '%s' "$out" | grep -qF -- "$needle" \
    && fail "$label: mutant still reports the finding (rc=$rc): $out"
  pass "$label"
}

# case 14 — stale verdict disabled: the case-6 fixture's stale-source message gone
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
stale_dir "$FIXTURE/build/test-results"
STALE_CHANGE="$CHANGE"
assert_mutation_removes_finding 'STALE-VERDICT' "$STALE_CHANGE" 'case 14: stale mutation' 'stale source'

# case 15 — absent verdict disabled: the case-5 fixture's absence message gone
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
ABSENT_CHANGE="$CHANGE"
assert_mutation_removes_finding 'ABSENT-VERDICT' "$ABSENT_CHANGE" 'case 15: absent mutation' 'does not exist'

# case 16 — dir mtime exactly at the commit epoch -> RC 0: the comparison is
# >=, and this case pins that boundary so a `<` -> `<=` mutation is caught
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
mkdir -p "$FIXTURE/build/test-results"
touch -t "$(date -r 1577836800 +%Y%m%d%H%M)" "$FIXTURE/build/test-results"
run_guard "$CHANGE"
[ "$RC" -eq 0 ] || fail "case 16: expected RC 0, got $RC: $OUT"
pass "case 16: boundary mtime == commit epoch -> fresh"

printf '%s\n' "failures: $FAILURES"
[ "$FAILURES" -eq 0 ]
