#!/usr/bin/env bash
# Assertion harness for preserve-session-records.sh. Builds sandboxed source
# and destination trees under TMPDIR; never touches the real repository tree.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated
# contract — the "Session records are preserved in the repository" and "Run 2
# removes the proposal artifact source" requirements in
# openspec/changes/kan-19-finish-safety-records-and-effort/specs/
# myflow-finish-cleanup/spec.md — never against observed output.
# test-check-plan-provenance.sh's header records that suite encoding the
# guard's own defects as its specification more than once, which then made each
# defect look verified.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT="$SCRIPT_DIR/preserve-session-records.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# Every case leaves two sandbox trees behind. They are removed on exit,
# including on a failed assertion. One case deliberately makes a destination
# directory unwritable, so write permission is restored before removal.
# An indexed array, not a space-separated string: sandbox paths come from
# mktemp under TMPDIR, which may contain spaces, and word-splitting a string
# would then leak every sandbox whose path split and `rm -rf` the fragments.
# This is the same hazard scripts/test-check-finish-preflight.sh's header
# names; bash 3.2 has indexed arrays, only associative arrays are unavailable.
TREES=()
cleanup() {
  # ${TREES[@]} is unset-expansion-unsafe under `set -u` on bash 3.2 when empty.
  [ "${#TREES[@]}" -eq 0 ] && return 0
  for tree in "${TREES[@]}"; do
    chmod -R u+w "$tree" 2>/dev/null || true
    rm -rf "$tree"
  done
}
trap cleanup EXIT

# new_tree -> sets WT and STATE_DIR, with all three sources present
new_tree() {
  WT="$(mktemp -d "${TMPDIR:-/tmp}/preserve-test-wt.XXXXXX")"
  STATE_DIR="$(mktemp -d "${TMPDIR:-/tmp}/preserve-test-state.XXXXXX")"
  TREES+=("$WT" "$STATE_DIR")
  mkdir -p "$WT/.superpowers/sdd/tasks"
  printf 'ledger body\n' > "$WT/.superpowers/sdd/tasks/progress.md"
  printf 'panel body\n' > "$WT/.superpowers/sdd/final-review-panel.md"
  printf '<p>artifact</p>\n' > "$STATE_DIR/demo-proposal-artifact.html"
}

# run_it -> sets RC and OUT for the change named `demo`
run_it() {
  set +e
  OUT="$("$SCRIPT" "$WT" demo "$STATE_DIR" 2>&1)"
  RC=$?
  set -e
}

# ===========================================================================
# SECTION: The three records reach the repository
# ===========================================================================

# 1. First copy places all three records under docs/superpowers/, and says so.
new_tree
run_it
[ "$RC" -eq 0 ] || fail "first copy: rc=$RC out=$OUT"
[ -n "$(find "$WT/docs/superpowers/ledgers" -name '*-demo.md' 2>/dev/null)" ] \
  && pass "ledger preserved" || fail "ledger missing: $OUT"
[ -n "$(find "$WT/docs/superpowers/reviews" -name '*-demo-panel.md' 2>/dev/null)" ] \
  && pass "panel preserved" || fail "panel missing: $OUT"
[ -n "$(find "$WT/docs/superpowers/artifacts" -name '*-demo.html' 2>/dev/null)" ] \
  && pass "artifact preserved" || fail "artifact missing: $OUT"
case "$OUT" in
  *"preserved: "*) pass "a copy is reported on stdout" ;;
  *) fail "no preserved: line reported: $OUT" ;;
esac

# 1b. The content actually arrives — a report of a copy that copied nothing
# would satisfy case 1 above.
# `|| true` because a missing destination directory is itself one of the
# outcomes under test: under `set -o pipefail` a failing `find` would abort
# the whole suite instead of failing this one case.
LEDGER_COPY="$(find "$WT/docs/superpowers/ledgers" -name '*-demo.md' 2>/dev/null | head -1 || true)"
if [ -n "$LEDGER_COPY" ] && grep -q 'ledger body' "$LEDGER_COPY"; then
  pass "the preserved ledger holds the source's content"
else
  fail "preserved ledger content: '$LEDGER_COPY'"
fi

# 1c. The destination name is date-stamped, per the spec's
# `docs/superpowers/ledgers/<date>-<name>.md`. Asserted as a shape
# (YYYY-MM-DD) rather than against today's date, so the case is not tied to
# the wall clock beyond the digit grouping the path contract states.
case "${LEDGER_COPY##*/}" in
  [0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]-demo.md) \
    pass "the preserved ledger is named <date>-<name>.md" ;;
  *) fail "preserved ledger name is not <date>-<name>.md: ${LEDGER_COPY##*/}" ;;
esac

# ===========================================================================
# SECTION: Idempotency — a fix round refreshes rather than duplicates
# ===========================================================================

# 2. A re-copy overwrites in place and creates no second dated file. The first
# copy is renamed to a fixed past date so the case cannot pass merely because
# both runs happen to land on the same day.
new_tree
run_it
LEDGER="$(find "$WT/docs/superpowers/ledgers" -name '*-demo.md' 2>/dev/null | head -1 || true)"
if [ -z "$LEDGER" ]; then
  fail "re-copy: the first copy wrote no ledger to rename: $OUT"
else
  mv "$LEDGER" "${LEDGER%/*}/1999-01-01-demo.md"
  printf 'ledger body v2\n' > "$WT/.superpowers/sdd/tasks/progress.md"
  run_it
  COUNT="$(find "$WT/docs/superpowers/ledgers" -name '*-demo.md' 2>/dev/null | wc -l | tr -d ' ' || true)"
  [ "$COUNT" = "1" ] && pass "re-copy does not duplicate" || fail "re-copy made $COUNT ledger files"
  grep -q 'v2' "$WT/docs/superpowers/ledgers/1999-01-01-demo.md" \
    && pass "re-copy overwrites the existing dated path" \
    || fail "re-copy did not overwrite the existing dated path"
fi

# 2b. The existing-file search must not adopt ANOTHER change's preserved
# record. `*-demo.md` matches `2020-01-01-other-demo.md`, so a bare glob would
# overwrite a different change's ledger — silent data loss. Only a
# date-prefixed file for this exact change name may be reused.
new_tree
mkdir -p "$WT/docs/superpowers/ledgers"
printf 'other change ledger\n' > "$WT/docs/superpowers/ledgers/2020-01-01-other-demo.md"
run_it
grep -q 'other change ledger' "$WT/docs/superpowers/ledgers/2020-01-01-other-demo.md" \
  && pass "another change's preserved ledger is not overwritten" \
  || fail "another change's ledger was overwritten: $OUT"
[ -n "$(find "$WT/docs/superpowers/ledgers" -name '*-demo.md' ! -name '*-other-demo.md')" ] \
  && pass "a fresh dated file is written for this change" \
  || fail "no ledger written for this change: $OUT"

# ===========================================================================
# SECTION: A missing source is skipped, never fatal
# ===========================================================================

# 3. Each source missing independently is skipped, reported, and non-fatal —
# and the other two are still preserved, since a partial gap must not become a
# total one.
for missing in ledger panel artifact; do
  new_tree
  case "$missing" in
    ledger) rm "$WT/.superpowers/sdd/tasks/progress.md" ;;
    panel) rm "$WT/.superpowers/sdd/final-review-panel.md" ;;
    artifact) rm "$STATE_DIR/demo-proposal-artifact.html" ;;
  esac
  run_it
  [ "$RC" -eq 0 ] && pass "missing $missing is not fatal" \
    || fail "missing $missing: rc=$RC out=$OUT"
  case "$OUT" in
    *skipped:*) pass "missing $missing is reported" ;;
    *) fail "missing $missing was not reported: $OUT" ;;
  esac
  REMAINING="$(find "$WT/docs/superpowers" -type f 2>/dev/null | wc -l | tr -d ' ' || true)"
  [ "$REMAINING" = "2" ] && pass "missing $missing: the other two records are still preserved" \
    || fail "missing $missing: preserved $REMAINING of the remaining 2 records"
done

# ===========================================================================
# SECTION: Failure is loud — a copy that was attempted and could not be
# written exits non-zero, unlike an absent source
# ===========================================================================

# 4. An unwritable destination directory fails the run and names the path, and
# the sources it CAN write are still preserved.
new_tree
mkdir -p "$WT/docs/superpowers/ledgers"
chmod a-w "$WT/docs/superpowers/ledgers"
run_it
chmod u+w "$WT/docs/superpowers/ledgers"
[ "$RC" -ne 0 ] && pass "an unwritable destination exits non-zero" \
  || fail "unwritable destination: expected non-zero, rc=$RC out=$OUT"
case "$OUT" in
  *"docs/superpowers/ledgers"*) pass "the failure names the destination it could not write" ;;
  *) fail "the failure does not name the destination: $OUT" ;;
esac
[ -n "$(find "$WT/docs/superpowers/reviews" -name '*-demo-panel.md' 2>/dev/null)" ] \
  && pass "one unwritable destination does not abandon the other records" \
  || fail "panel not preserved after a ledger write failure: $OUT"

# ===========================================================================
# SECTION: The destination must stay inside the worktree
# ===========================================================================

# 4b. A destination directory that is a SYMLINK out of the worktree is refused,
# not followed. The three directories under docs/superpowers/ are ordinary
# tracked repo paths, editable in any pull request, and this script runs
# automatically from /myflow-do's push path and /myflow-finish run 1 — so a
# symlink there is a destination-controlled arbitrary-file-write primitive.
# `mkdir -p` is a no-op on an existing symlink and `cp` follows it, so nothing
# else in the script would notice.
for linked in ledgers reviews artifacts; do
  new_tree
  OUTSIDE="$(mktemp -d "${TMPDIR:-/tmp}/preserve-test-outside.XXXXXX")"
  TREES+=("$OUTSIDE")
  mkdir -p "$WT/docs/superpowers"
  ln -s "$OUTSIDE" "$WT/docs/superpowers/$linked"
  run_it
  [ -z "$(find "$OUTSIDE" -type f 2>/dev/null)" ] \
    && pass "symlinked $linked: nothing is written outside the worktree" \
    || fail "symlinked $linked: a record was written outside the worktree: $(find "$OUTSIDE" -type f)"
  [ "$RC" -ne 0 ] && pass "symlinked $linked: exits non-zero" \
    || fail "symlinked $linked: expected non-zero, rc=$RC out=$OUT"
  case "$OUT" in
    *"preserve-session-records:"*"$linked"*) pass "symlinked $linked: the refusal names the destination" ;;
    *) fail "symlinked $linked: refusal does not name the destination: $OUT" ;;
  esac
  # The other two records are still preserved: one untrusted destination must
  # not cost the records that can be written safely, exactly as an unwritable
  # destination does not in case 4.
  REMAINING="$(find "$WT/docs/superpowers" -type f 2>/dev/null | wc -l | tr -d ' ' || true)"
  [ "$REMAINING" = "2" ] && pass "symlinked $linked: the other two records are still preserved" \
    || fail "symlinked $linked: preserved $REMAINING of the remaining 2 records"
done

# 4c. A symlink pointing INSIDE the worktree is fine — the boundary is the
# worktree, not the use of a symlink, and refusing it would break a repository
# that legitimately links one docs directory to another.
new_tree
mkdir -p "$WT/docs/superpowers" "$WT/real-ledgers"
ln -s "$WT/real-ledgers" "$WT/docs/superpowers/ledgers"
run_it
[ "$RC" -eq 0 ] && [ -n "$(find "$WT/real-ledgers" -name '*-demo.md' 2>/dev/null)" ] \
  && pass "a symlink inside the worktree is followed, not refused" \
  || fail "in-worktree symlink: rc=$RC out=$OUT"

# 4c-ii. The reported destination is the RESOLVED directory, not the argument
# the script was handed. This is the observable consequence of building the
# destination path from the already-verified resolved directory instead of from
# the unresolved argument: had `cp` been left to re-follow the symlink itself,
# the check would have validated one path while the write went through another,
# and the reported line would name the unvalidated one. The reported path is
# therefore also the evidence that check and write agree.
WT_REAL="$(cd -P "$WT" && pwd -P)"
case "$OUT" in
  *"preserved: $WT_REAL/real-ledgers/"*) \
    pass "the reported destination is the resolved directory, not the symlink reached through it" ;;
  *) fail "in-worktree symlink: the reported destination is not the resolved directory: $OUT" ;;
esac

# ===========================================================================
# SECTION: The sources must stay inside the root they come from
# ===========================================================================

# 4c-iii. A SOURCE that is a symlink out of its root is refused, not read.
# Validating only the destination leaves an arbitrary-file-READ: all three
# source paths are plantable — the two under `.superpowers/` are gitignored but
# `.gitignore` only gates untracked paths, so a symlink forced in with
# `git add -f` lands on disk on checkout — and `[ -f ]` follows symlinks while
# `cp` reads through them. The result would then be committed and pushed by
# /myflow-finish run 1, which runs this script before `git add -A`.
#
# The refusal is deliberately NOT the `skipped:` path. `skipped:` means "this
# change legitimately has no such record"; this is a source that exists and
# cannot be trusted, so it is reported on stderr and is fatal to its own copy.
for linked in ledger panel artifact; do
  new_tree
  OUTSIDE="$(mktemp -d "${TMPDIR:-/tmp}/preserve-test-outside.XXXXXX")"
  TREES+=("$OUTSIDE")
  printf 'PLANTED-SECRET-VALUE\n' > "$OUTSIDE/planted"
  case "$linked" in
    ledger) SRC="$WT/.superpowers/sdd/tasks/progress.md" ;;
    panel) SRC="$WT/.superpowers/sdd/final-review-panel.md" ;;
    artifact) SRC="$STATE_DIR/demo-proposal-artifact.html" ;;
  esac
  rm -f "$SRC"
  ln -s "$OUTSIDE/planted" "$SRC"
  run_it
  [ -z "$(grep -rl 'PLANTED-SECRET-VALUE' "$WT/docs" 2>/dev/null || true)" ] \
    && pass "symlinked $linked source: the planted content never reaches the repository" \
    || fail "symlinked $linked source: planted content was preserved into the repository: $OUT"
  [ "$RC" -ne 0 ] && pass "symlinked $linked source: exits non-zero" \
    || fail "symlinked $linked source: expected non-zero, rc=$RC out=$OUT"
  case "$OUT" in
    *"preserve-session-records:"*"$SRC"*) pass "symlinked $linked source: the refusal names the source" ;;
    *) fail "symlinked $linked source: refusal does not name the source: $OUT" ;;
  esac
  case "$OUT" in
    *"skipped: $SRC"*) fail "symlinked $linked source: reported as absent rather than refused: $OUT" ;;
    *) pass "symlinked $linked source: not reported as merely absent" ;;
  esac
  # As with an unwritable or an untrusted destination, the other two records are
  # still attempted: one untrusted source must not cost the records that are fine.
  REMAINING="$(find "$WT/docs/superpowers" -type f 2>/dev/null | wc -l | tr -d ' ' || true)"
  [ "$REMAINING" = "2" ] && pass "symlinked $linked source: the other two records are still preserved" \
    || fail "symlinked $linked source: preserved $REMAINING of the remaining 2 records"
done

# 4c-iv. A source symlink pointing INSIDE its own root is followed, not refused.
# The boundary is the root, not the use of a symlink — the same rule the
# destination side applies in case 4c.
new_tree
mkdir -p "$WT/real-records"
printf 'in-tree ledger body\n' > "$WT/real-records/progress.md"
rm -f "$WT/.superpowers/sdd/tasks/progress.md"
ln -s "$WT/real-records/progress.md" "$WT/.superpowers/sdd/tasks/progress.md"
run_it
LEDGER_COPY="$(find "$WT/docs/superpowers/ledgers" -name '*-demo.md' 2>/dev/null | head -1 || true)"
if [ "$RC" -eq 0 ] && [ -n "$LEDGER_COPY" ] && grep -q 'in-tree ledger body' "$LEDGER_COPY"; then
  pass "a source symlink inside its root is followed, not refused"
else
  fail "in-root source symlink: rc=$RC out=$OUT"
fi

# ===========================================================================
# SECTION: Invocation contract
# ===========================================================================

# 4d. A change name is checked against an allowlist, and anything else is
# rejected outright. Two distinct hazards make one check:
#   * `/` — traversal is blocked today only INCIDENTALLY, because `$TODAY-`
#     shares the path component with the start of the name, so a leading `../`
#     becomes part of a literal directory name. That is an accident of string
#     concatenation, not a protection, and the source path
#     `$STATE_DIR/$NAME-proposal-artifact.html` has no such accident to rely on.
#   * glob metacharacters — the name flows into `find -name
#     "[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]-${NAME}${suffix}"`, whose
#     digit anchoring exists to stop one change adopting another's file but
#     assumes the name itself carries no metacharacters. `*` matches every
#     change's preserved record; see case 4d-ii.
# Change names are lowercased Jira keys plus kebab-case slugs, so the allowlist
# costs nothing a real name needs.
for badname in ../escape a/b /abs '*' 'a*b' 'a?b' '[0-9]' 'a b' 'a;b' '$(id)' '-lead' '.lead'; do
  new_tree
  set +e
  OUT="$("$SCRIPT" "$WT" "$badname" "$STATE_DIR" 2>&1)"
  RC=$?
  set -e
  [ "$RC" -eq 2 ] && pass "change name '$badname' exits 2" \
    || fail "change name '$badname': expected exit 2, rc=$RC out=$OUT"
  case "$OUT" in
    *"preserve-session-records:"*) pass "change name '$badname' is named in the error" ;;
    *) fail "change name '$badname': no named error: $OUT" ;;
  esac
  [ -z "$(find "$WT/docs" -type f 2>/dev/null)" ] \
    && pass "change name '$badname' wrote nothing" \
    || fail "change name '$badname' wrote files: $(find "$WT/docs" -type f)"
done

# 4d-i2. The allowlist answers the same under every locale, which is a property
# separate from what it answers. Protection 1 spelled its bracket expressions
# with RANGES, and a bracket range is a COLLATING range: measured on bash 3.2
# (Darwin 25.5.0), every name below was refused under `LC_ALL=C` and ADMITTED
# under `en_US.UTF-8` by that spelling — so the accepted input set of a
# containment gate moved with the operator's environment, and an admitted name
# then reached `find -name` and a constructed path with nothing else in the way.
# The characters are enumerated now, and a literal list has no endpoints for a
# collation order to reorder.
#
# THE PROBES ARE FOUR DIFFERENT HAZARDS PLUS THE ONE THE OTHER SUITES USE, not
# five spellings of one: a Latin letter with a diacritic, a ligature, a Roman
# numeral, a fullwidth capital, and `İstanbul-test`. A locale this machine lacks
# is skipped rather than forced — `LC_ALL` naming a missing locale silently falls
# back to C, and the loop would then pass by not running, which is the vacuous
# green this suite is written against.
PSR_LOCALES=(C)
for psr_loc in en_US.UTF-8 tr_TR.UTF-8; do
  if [ "$(LC_ALL="$psr_loc" locale charmap 2>/dev/null)" = "UTF-8" ]; then
    PSR_LOCALES+=("$psr_loc")
  else
    printf 'skip: the allowlist under LC_ALL=%s (this machine has no such locale)\n' "$psr_loc"
  fi
done
for psr_name in "écho" "ﬀoo" "ⅰx" "Ａbc" "İstanbul-test"; do
  for psr_loc in "${PSR_LOCALES[@]}"; do
    new_tree
    set +e
    # The assignment is a PREFIX on an external command, which is the one form
    # of it that cannot leak into the cases after this loop.
    OUT="$(LC_ALL="$psr_loc" LANG="$psr_loc" "$SCRIPT" "$WT" "$psr_name" "$STATE_DIR" 2>&1)"
    RC=$?
    set -e
    [ "$RC" -eq 2 ] && pass "change name '$psr_name' exits 2 under LC_ALL=$psr_loc" \
      || fail "change name '$psr_name' under LC_ALL=$psr_loc: expected exit 2, rc=$RC out=$OUT"
    [ -z "$(find "$WT/docs" -type f 2>/dev/null)" ] \
      && pass "change name '$psr_name' wrote nothing under LC_ALL=$psr_loc" \
      || fail "change name '$psr_name' under LC_ALL=$psr_loc wrote files: $(find "$WT/docs" -type f)"
  done
done

# 4d-ii. The consequence the allowlist exists to prevent: a glob metacharacter
# in the change name defeats the digit-anchored existing-file search and lets
# one invocation adopt — and overwrite — a DIFFERENT change's preserved record.
# Case 2b proves the anchoring works for an ordinary name; this proves the
# anchoring is not the whole protection on its own.
new_tree
mkdir -p "$WT/docs/superpowers/ledgers"
printf 'otherchange ledger\n' > "$WT/docs/superpowers/ledgers/2020-01-01-otherchange.md"
set +e
OUT="$("$SCRIPT" "$WT" '*' "$STATE_DIR" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "a glob change name exits 2" \
  || fail "glob change name: expected exit 2, rc=$RC out=$OUT"
grep -q 'otherchange ledger' "$WT/docs/superpowers/ledgers/2020-01-01-otherchange.md" \
  && pass "a glob change name does not overwrite another change's record" \
  || fail "a glob change name overwrote another change's record: $OUT"

# 4d-iii. The names real changes use are accepted. An allowlist that rejected a
# legitimate name would break every change instead of protecting it, so the
# shapes `/myflow-start` produces are asserted alongside the rejections: a
# lowercased Jira key plus a kebab slug, a slug alone, and the `-fix-N`
# sub-change form.
for goodname in kan-19-finish-safety-records-and-effort plain-slug kan-19-x-fix-2 a; do
  new_tree
  printf '<p>artifact</p>\n' > "$STATE_DIR/$goodname-proposal-artifact.html"
  set +e
  OUT="$("$SCRIPT" "$WT" "$goodname" "$STATE_DIR" 2>&1)"
  RC=$?
  set -e
  [ "$RC" -eq 0 ] && pass "change name '$goodname' is accepted" \
    || fail "change name '$goodname': expected exit 0, rc=$RC out=$OUT"
  COUNT="$(find "$WT/docs/superpowers" -type f 2>/dev/null | wc -l | tr -d ' ' || true)"
  [ "$COUNT" = "3" ] && pass "change name '$goodname' preserves all three records" \
    || fail "change name '$goodname' preserved $COUNT of 3 records: $OUT"
done

# 5. Missing arguments are programmer error: exit 2 with a usage line, never a
# silent no-op that would read as "nothing to preserve".
for args in 0 1 2; do
  set +e
  case "$args" in
    0) OUT="$("$SCRIPT" 2>&1)" ;;
    1) OUT="$("$SCRIPT" /tmp 2>&1)" ;;
    2) OUT="$("$SCRIPT" /tmp demo 2>&1)" ;;
  esac
  RC=$?
  set -e
  [ "$RC" -eq 2 ] && pass "$args argument(s) exits 2" || fail "$args args: rc=$RC out=$OUT"
  case "$OUT" in
    *usage:*) pass "$args argument(s) prints usage" ;;
    *) fail "$args args: no usage line: $OUT" ;;
  esac
done

# 6. Runnable from any directory: the script takes every path it touches as an
# argument, so its own cwd is irrelevant. /myflow-finish invokes it from the
# worktree, /myflow-do from wherever the session sits.
new_tree
ELSEWHERE="$(mktemp -d "${TMPDIR:-/tmp}/preserve-test-cwd.XXXXXX")"
TREES+=("$ELSEWHERE")
set +e
OUT="$(cd "$ELSEWHERE" && "$SCRIPT" "$WT" demo "$STATE_DIR" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] && [ -n "$(find "$WT/docs/superpowers/ledgers" -name '*-demo.md' 2>/dev/null)" ] \
  && pass "runs correctly from an unrelated working directory" \
  || fail "unrelated cwd: rc=$RC out=$OUT"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'preserve-session-records: all cases pass\n'
