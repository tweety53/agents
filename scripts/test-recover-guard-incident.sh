#!/usr/bin/env bash
# Assertion harness for recover-guard-incident.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR and asserts the tool's refusals, its
# dry-run plan and its --apply effects. Never touches the real repository.
#
# THE CONTRACT THIS FILE IS THE EXECUTABLE STATEMENT OF:
#
#   exit 0   dry-run plan printed, nothing changed; or --apply completed
#   exit 1   precondition failure, cause on stderr, nothing on stdout
#   exit 2   usage error
#   --apply   REVERT_HEAD gone first, then untracked restores, never staged
#
# Shape copied from test-check-worktree-location.sh. Bash 3.2 is the floor.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="${RECOVER_GUARD_INCIDENT_TOOL:-$SCRIPT_DIR/recover-guard-incident.sh}"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

SANDBOXES=()
cleanup() {
  if [ "${#SANDBOXES[@]}" -ne 0 ]; then
    for s in "${SANDBOXES[@]}"; do
      chmod -R u+rwX "$s" 2>/dev/null || true
      rm -rf "$s"
    done
  fi
}
trap cleanup EXIT

WORK="$(mktemp -d "${TMPDIR:-/tmp}/recover-guard-incident-test.XXXXXX")"
SANDBOXES+=("$WORK")
ERRFILE="$WORK/stderr"

# run_tool <arg ...> -> OUT (stdout only), ERR (stderr only), RC. The streams
# are captured separately, never merged with 2>&1.
run_tool() {
  set +e
  OUT="$(bash "$GUARD" "$@" 2>"$ERRFILE")"
  RC=$?
  set -e
  ERR="$(cat "$ERRFILE")"
}

# run_tool_in <dir> <arg ...> -> same, but the tool runs with <dir> as its
# cwd (subdirectory-invocation cases).
run_tool_in() {
  local d="$1"
  shift
  set +e
  OUT="$(cd "$d" && bash "$GUARD" "$@" 2>"$ERRFILE")"
  RC=$?
  set -e
  ERR="$(cat "$ERRFILE")"
}

# new_repo -> REPO (physical form), a git repo with one commit, a tracked
# file two commits touch (so a conflicting revert is available), and planning
# files that only ever exist untracked (so a `git stash -u` captures them
# into the stash's third parent).
new_repo() {
  local r
  r="$(mktemp -d "${TMPDIR:-/tmp}/recover-guard-incident-repo.XXXXXX")"
  SANDBOXES+=("$r")
  git init -q "$r"
  git -C "$r" config user.email t@example.com
  git -C "$r" config user.name t
  printf 'one\n' >"$r/shared.txt"
  git -C "$r" add shared.txt
  git -C "$r" commit -q -m base
  printf 'two\n' >"$r/shared.txt"
  git -C "$r" commit -qam two
  printf 'one\n' >"$r/shared.txt"   # reverts `two` WITH conflict on commit, abort-able
  git -C "$r" commit -qam three
  REPO="$(cd "$r" && pwd -P)"
}

# start_conflicting_revert -> leaves a revert in progress (REVERT_HEAD set),
# the state the incident was recovered from.
start_conflicting_revert() {
  set +e
  git -C "$REPO" revert --no-commit HEAD~1 >/dev/null 2>&1
  set -e
}

# stash_planning_files -> the incident's stash: untracked planning files under
# the tool's default paths, captured into stash@{0}'s third parent and gone
# from the worktree afterwards. PLAN1/PLAN2 hold the contents asserted later.
P1="docs/superpowers/plans/kan-423-plan.md"
P2="spectre/changes/kan-423/tasks.md"
stash_planning_files() {
  mkdir -p "$REPO/docs/superpowers/plans" "$REPO/spectre/changes/kan-423"
  PLAN1='incident plan body
'
  PLAN2='- restore the planning files
'
  printf '%s' "$PLAN1" >"$REPO/$P1"
  printf '%s' "$PLAN2" >"$REPO/$P2"
  git -C "$REPO" stash -q -u
}

# ---------------------------------------------------------------------------
# Case 1: usage-error-exit-2 — a plain directory that is not a git repo.
# ---------------------------------------------------------------------------
NOT_A_REPO="$(mktemp -d "${TMPDIR:-/tmp}/recover-guard-incident-notrepo.XXXXXX")"
SANDBOXES+=("$NOT_A_REPO")
run_tool "$NOT_A_REPO"
[ "$RC" -eq 2 ] && pass "case 1: exits 2 on a non-repo directory" \
  || fail "case 1: expected exit 2, got rc=$RC out=$OUT err=$ERR"
[ -z "$OUT" ] && pass "case 1: writes nothing to stdout" \
  || fail "case 1: emitted something on stdout: $OUT"
[ -n "$ERR" ] && pass "case 1: names the failure on stderr" \
  || fail "case 1: stderr was empty"
case "$ERR" in
  *" 2") fail "case 1: stderr leaks the exit code into the message: $ERR" ;;
  *) pass "case 1: stderr carries the cause only, not the exit code" ;;
esac

# ---------------------------------------------------------------------------
# Case 1b: unknown-flag-usage-error — an unknown flag after repo-dir is a
# usage error (exit 2), not a planning path falling through to exit 1.
# ---------------------------------------------------------------------------
new_repo
stash_planning_files
run_tool "$REPO" --bogus
[ "$RC" -eq 2 ] && pass "case 1b: unknown flag exits 2" \
  || fail "case 1b: expected exit 2, got rc=$RC out=$OUT err=$ERR"
[ -z "$OUT" ] && pass "case 1b: stdout empty on the usage error" \
  || fail "case 1b: stdout carried: $OUT"
case "$ERR" in
  *"--bogus"*) pass "case 1b: stderr names the offending flag" ;;
  *) fail "case 1b: stderr does not name --bogus: $ERR" ;;
esac

# ---------------------------------------------------------------------------
# Case 2: no-revert-refusal — no revert in progress, nothing changed.
# ---------------------------------------------------------------------------
new_repo
stash_planning_files
run_tool "$REPO"
[ "$RC" -eq 1 ] && pass "case 2: exits 1 with no revert in progress" \
  || fail "case 2: expected exit 1, got rc=$RC out=$OUT err=$ERR"
[ -z "$OUT" ] && pass "case 2: refusal leaves stdout empty" \
  || fail "case 2: stdout carried: $OUT"
case "$ERR" in
  *REVERT_HEAD*) pass "case 2: stderr names the missing REVERT_HEAD" ;;
  *) fail "case 2: stderr does not name REVERT_HEAD: $ERR" ;;
esac
if git -C "$REPO" rev-parse -q --verify REVERT_HEAD >/dev/null 2>&1; then
  fail "case 2: REVERT_HEAD appeared out of nowhere"
else
  pass "case 2: still no revert in progress"
fi

# ---------------------------------------------------------------------------
# Case 3: stash-without-third-parent-refusal — the refusal distinguishes the
# cause, revert still in progress.
# ---------------------------------------------------------------------------
new_repo
printf 'x\n' >>"$REPO/shared.txt"
git -C "$REPO" stash -q   # tracked-only stash: no untracked third parent
start_conflicting_revert
run_tool "$REPO"
[ "$RC" -eq 1 ] && pass "case 3: exits 1 on a third-parent-less stash" \
  || fail "case 3: expected exit 1, got rc=$RC out=$OUT err=$ERR"
[ -z "$OUT" ] && pass "case 3: refusal leaves stdout empty" \
  || fail "case 3: stdout carried: $OUT"
case "$ERR" in
  *"third parent"*) pass "case 3: stderr names the missing third parent" ;;
  *) fail "case 3: stderr does not name the third parent: $ERR" ;;
esac
case "$ERR" in
  *"no stash entry"*) fail "case 3: the no-third-parent cause is indistinguishable from a missing stash: $ERR" ;;
  *) pass "case 3: cause is distinct from the missing-stash message" ;;
esac
git -C "$REPO" rev-parse -q --verify REVERT_HEAD >/dev/null 2>&1 \
  && pass "case 3: revert still in progress" \
  || fail "case 3: the refusal aborted the revert"

# ---------------------------------------------------------------------------
# Case 4: dry-run-changes-nothing — lists the plan, changes nothing.
# ---------------------------------------------------------------------------
new_repo
stash_planning_files
start_conflicting_revert
run_tool "$REPO"
[ "$RC" -eq 0 ] && pass "case 4: dry-run exits 0" \
  || fail "case 4: expected exit 0, got rc=$RC out=$OUT err=$ERR"
case "$OUT" in
  *"revert --abort"*) pass "case 4: plan lists the revert --abort" ;;
  *) fail "case 4: no revert --abort in the plan: $OUT" ;;
esac
for f in "$P1" "$P2"; do
  case "$OUT" in
    *"show \"stash@{0}^3:$f\""*) pass "case 4: plan restores $f via a git show redirect" ;;
    *) fail "case 4: no git show redirect for $f in: $OUT" ;;
  esac
done
case "$OUT" in
  *"reflog diagnosis"*) pass "case 4: stdout carries the reflog diagnosis block" ;;
  *) fail "case 4: no reflog diagnosis in: $OUT" ;;
esac
git -C "$REPO" rev-parse -q --verify REVERT_HEAD >/dev/null 2>&1 \
  && pass "case 4: REVERT_HEAD untouched" \
  || fail "case 4: the dry-run aborted the revert"
[ ! -e "$REPO/$P1" ] && [ ! -e "$REPO/$P2" ] \
  && pass "case 4: no planning file exists on disk" \
  || fail "case 4: a planning file appeared on disk"
[ -z "$(git -C "$REPO" diff --cached)" ] \
  && pass "case 4: nothing staged" \
  || fail "case 4: the dry-run staged something"

# ---------------------------------------------------------------------------
# Case 5: apply-restores-unstaged-after-abort — aborts first, restores
# unstaged.
# ---------------------------------------------------------------------------
new_repo
stash_planning_files
start_conflicting_revert
run_tool --apply "$REPO"
[ "$RC" -eq 0 ] && pass "case 5: apply exits 0" \
  || fail "case 5: expected exit 0, got rc=$RC out=$OUT err=$ERR"
# The ordering is normative (design decision redirect-restore-after-single-
# abort): everything printed before the FIRST restore run-line must carry
# the abort run-line. A tool that restores first and aborts after prints a
# restore line with no abort before it.
case "${OUT%%"running: git show"*}" in
  *"running: git -C"*"revert --abort"*) pass "case 5: the abort precedes every restore" ;;
  *) fail "case 5: a restore line printed before the abort line: $OUT" ;;
esac
if git -C "$REPO" rev-parse -q --verify REVERT_HEAD >/dev/null 2>&1; then
  fail "case 5: REVERT_HEAD still present after apply"
else
  pass "case 5: REVERT_HEAD gone"
fi
# $(cat) strips the trailing newline on BOTH sides, so the comparison is
# against the fixture with its newline stripped the same way.
[ "$(cat "$REPO/$P1")" = "$(printf '%s' "$PLAN1")" ] && pass "case 5: $P1 back with stash content" \
  || fail "case 5: $P1 content wrong: $(cat "$REPO/$P1")"
[ "$(cat "$REPO/$P2")" = "$(printf '%s' "$PLAN2")" ] && pass "case 5: $P2 back with stash content" \
  || fail "case 5: $P2 content wrong: $(cat "$REPO/$P2")"
UNSTAGED=1
for f in "$P1" "$P2"; do
  case "$(git -C "$REPO" status --porcelain -- "$f")" in
    "?? "*) ;;
    *) UNSTAGED=0 ;;
  esac
done
[ "$UNSTAGED" -eq 1 ] && pass "case 5: every restored file is untracked" \
  || fail "case 5: a restored file is not untracked: $(git -C "$REPO" status --porcelain)"
[ -z "$(git -C "$REPO" diff --cached)" ] \
  && pass "case 5: nothing staged — the incident's property holds" \
  || fail "case 5: the apply staged the restored files"

# ---------------------------------------------------------------------------
# Case 6: tracked-target-refusal — a tracked file at a restore target
# refuses the whole run.
# ---------------------------------------------------------------------------
new_repo
stash_planning_files
start_conflicting_revert
mkdir -p "$REPO/$(dirname "$P1")"
printf 'tracked content\n' >"$REPO/$P1"
git -C "$REPO" add "$P1"
run_tool "$REPO"
[ "$RC" -eq 1 ] && pass "case 6: exits 1 on a tracked target" \
  || fail "case 6: expected exit 1, got rc=$RC out=$OUT err=$ERR"
[ -z "$OUT" ] && pass "case 6: refusal leaves stdout empty" \
  || fail "case 6: stdout carried: $OUT"
case "$ERR" in
  *"$P1"*) pass "case 6: stderr names the tracked file" ;;
  *) fail "case 6: stderr does not name $P1: $ERR" ;;
esac
git -C "$REPO" rev-parse -q --verify REVERT_HEAD >/dev/null 2>&1 \
  && pass "case 6: revert still in progress" \
  || fail "case 6: the refusal aborted the revert"
[ "$(cat "$REPO/$P1")" = "tracked content" ] \
  && pass "case 6: tracked content untouched" \
  || fail "case 6: tracked content was clobbered"
case "$(git -C "$REPO" status --porcelain -- "$P1")" in
  "A  "*) pass "case 6: the staged file is still staged" ;;
  *) fail "case 6: the staged file's index state changed: $(git -C "$REPO" status --porcelain -- "$P1")" ;;
esac
case "$ERR" in
  *" 1") fail "case 6: stderr leaks the exit code into the message: $ERR" ;;
  *) pass "case 6: refusal message carries the cause only" ;;
esac

# ---------------------------------------------------------------------------
# Case 7: subdirectory-invocation-anchors-at-root — repo-dir (explicit or
# cwd) resolves to the toplevel, so path... is repo-root-relative.
# ---------------------------------------------------------------------------
new_repo
stash_planning_files
start_conflicting_revert
mkdir -p "$REPO/sub"
run_tool_in "$REPO/sub"
[ "$RC" -eq 0 ] && pass "case 7: cwd-in-subdirectory dry-run exits 0" \
  || fail "case 7: expected exit 0, got rc=$RC out=$OUT err=$ERR"
for f in "$P1" "$P2"; do
  case "$OUT" in
    *"> $REPO/$f"*) pass "case 7: $f anchored at the repo root from cwd" ;;
    *) fail "case 7: restore for $f not anchored at $REPO: $OUT" ;;
  esac
done
run_tool "$REPO/sub"
[ "$RC" -eq 0 ] && pass "case 7: explicit subdirectory repo-dir exits 0" \
  || fail "case 7: explicit subdir expected exit 0, got rc=$RC out=$OUT err=$ERR"
for f in "$P1" "$P2"; do
  case "$OUT" in
    *"> $REPO/$f"*) pass "case 7: $f anchored at the repo root from explicit dir" ;;
    *) fail "case 7: explicit-dir restore for $f not anchored at $REPO: $OUT" ;;
  esac
done
git -C "$REPO" rev-parse -q --verify REVERT_HEAD >/dev/null 2>&1 \
  && pass "case 7: REVERT_HEAD untouched" \
  || fail "case 7: the dry-run aborted the revert"

# ---------------------------------------------------------------------------
# Case 8: overwrite-naming — an existing untracked target is named as an
# overwrite, an absent sibling is not, and the mark is never inside a
# command line (the plan stays paste-runnable).
# ---------------------------------------------------------------------------
new_repo
mkdir -p "$REPO/a" "$REPO/za"
printf 'aa\n' >"$REPO/a/notes.md"
printf 'za\n' >"$REPO/za/notes.md"
git -C "$REPO" stash -q -u
start_conflicting_revert
mkdir -p "$REPO/za"   # za/notes.md back untracked; a/notes.md stays absent
printf 'za\n' >"$REPO/za/notes.md"
run_tool "$REPO" a za
[ "$RC" -eq 0 ] && pass "case 8: dry-run with an existing untracked target exits 0" \
  || fail "case 8: expected exit 0, got rc=$RC out=$OUT err=$ERR"
OW=$(printf '%s\n' "$OUT" | grep -c 'overwrites an existing untracked file' || true)
[ "$OW" -eq 1 ] && pass "case 8: exactly one overwrite mark (za.md, not absent a.md)" \
  || fail "case 8: expected 1 overwrite mark, got $OW: $OUT"
case "$OUT" in
  *"(overwrites an existing untracked file: za/notes.md)"*) \
    pass "case 8: the mark names za/notes.md on its own line" ;;
  *) fail "case 8: no standalone overwrite line naming za/notes.md: $OUT" ;;
esac
BAD=$(printf '%s\n' "$OUT" | grep '^  mkdir' | grep -c '(' || true)
[ "$BAD" -eq 0 ] && pass "case 8: every mkdir line parses as a command" \
  || fail "case 8: $BAD mkdir line(s) carry annotation text: $OUT"

# ---------------------------------------------------------------------------
# Case 9: HEAD-tracked-target-refusal — a target tracked in HEAD only
# (committed, then removed from the index) refuses the whole run.
# ---------------------------------------------------------------------------
new_repo
mkdir -p "$REPO/$(dirname "$P1")"
printf 'committed\n' >"$REPO/$P1"
git -C "$REPO" add "$P1"
git -C "$REPO" commit -qam sneak
git -C "$REPO" rm -q --cached "$P1"   # tracked in HEAD, untracked in worktree
stash_planning_files
git -C "$REPO" rm -q --cached "$P1"   # `stash -u` restores the index to HEAD,
                                       # re-tracking $P1 — remove it from the
                                       # index again so only the cat-file
                                       # HEAD-tracked half of the guard can fire
start_conflicting_revert
run_tool "$REPO"
[ "$RC" -eq 1 ] && pass "case 9: exits 1 on a HEAD-tracked target" \
  || fail "case 9: expected exit 1, got rc=$RC out=$OUT err=$ERR"
[ -z "$OUT" ] && pass "case 9: refusal leaves stdout empty" \
  || fail "case 9: stdout carried: $OUT"
case "$ERR" in
  *"$P1"*) pass "case 9: stderr names the HEAD-tracked file" ;;
  *) fail "case 9: stderr does not name $P1: $ERR" ;;
esac
git -C "$REPO" rev-parse -q --verify REVERT_HEAD >/dev/null 2>&1 \
  && pass "case 9: revert still in progress" \
  || fail "case 9: the refusal aborted the revert"

# ---------------------------------------------------------------------------
# Case 10: custom-paths-replace-defaults — explicit path... replaces the
# default paths and resolves repo-root-relative.
# ---------------------------------------------------------------------------
new_repo
mkdir -p "$REPO/notes"
printf 'incident note\n' >"$REPO/notes/incident.md"
git -C "$REPO" stash -q -u
start_conflicting_revert
run_tool "$REPO" notes
[ "$RC" -eq 0 ] && pass "case 10: custom path dry-run exits 0" \
  || fail "case 10: expected exit 0, got rc=$RC out=$OUT err=$ERR"
case "$OUT" in
  *"show \"stash@{0}^3:notes/incident.md\""*) \
    pass "case 10: plan restores the custom path's file" ;;
  *) fail "case 10: no show line for notes/incident.md: $OUT" ;;
esac
case "$OUT" in
  *"> $REPO/notes/incident.md"*) pass "case 10: custom path resolved repo-root-relative" ;;
  *) fail "case 10: custom path not anchored at $REPO: $OUT" ;;
esac
case "$OUT" in
  *"stash@{0}^3:docs/superpowers"*) fail "case 10: custom path did not replace the defaults: $OUT" ;;
  *) pass "case 10: default paths absent from the plan" ;;
esac

# ---------------------------------------------------------------------------
# Case 11: empty-restore-set-refusal — a path the stash's third parent does
# not hold refuses (exit 1) before the abort, even with --apply.
# ---------------------------------------------------------------------------
new_repo
stash_planning_files
start_conflicting_revert
run_tool --apply "$REPO" spectre/nope
[ "$RC" -eq 1 ] && pass "case 11: exits 1 when the restore set is empty" \
  || fail "case 11: expected exit 1, got rc=$RC out=$OUT err=$ERR"
[ -z "$OUT" ] && pass "case 11: refusal leaves stdout empty" \
  || fail "case 11: stdout carried: $OUT"
case "$ERR" in
  *"holds no files"*) pass "case 11: stderr names the empty-restore-set cause" ;;
  *) fail "case 11: stderr does not name the cause: $ERR" ;;
esac
git -C "$REPO" rev-parse -q --verify REVERT_HEAD >/dev/null 2>&1 \
  && pass "case 11: --apply aborted nothing — revert still in progress" \
  || fail "case 11: --apply ran the abort on a refused plan"
[ ! -e "$REPO/$P1" ] && [ ! -e "$REPO/$P2" ] \
  && pass "case 11: nothing restored" \
  || fail "case 11: a file appeared on disk"

# ---------------------------------------------------------------------------
# Case 12: reflog-block-bounded-at-15 — the diagnosis block carries the last
# 15 entries when the reflog holds more.
# ---------------------------------------------------------------------------
new_repo
stash_planning_files
i=0
while [ "$i" -lt 14 ]; do
  git -C "$REPO" commit -q --allow-empty -m "fill$i"
  i=$((i + 1))
done
start_conflicting_revert
run_tool "$REPO"
[ "$RC" -eq 0 ] && pass "case 12: dry-run exits 0" \
  || fail "case 12: expected exit 0, got rc=$RC out=$OUT err=$ERR"
REFLOG=$(printf '%s\n' "$OUT" | grep -c 'HEAD@{' || true)
[ "$REFLOG" -eq 15 ] && pass "case 12: diagnosis block carries exactly 15 entries" \
  || fail "case 12: expected 15 reflog entries, got $REFLOG"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'recover-guard-incident: all cases pass\n'
