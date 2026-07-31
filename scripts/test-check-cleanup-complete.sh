#!/usr/bin/env bash
# Assertion harness for check-cleanup-complete.sh. Builds throwaway git
# repositories and state directories under a sandboxed TMPDIR and asserts the
# guard's verdict and exit status. Never MODIFIES the real repository tree —
# with one deliberate read: the registry-coupling case reads
# skills/myflow-contracts/pipeline.md, because the thing under test there is the
# guard's agreement with the real registry and a fixture copy of it would be a
# third place for the same list to drift.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated
# contract in openspec/specs/myflow-finish-cleanup/spec.md — the requirement
# "Run 2 verifies that cleanup actually completed", which lives in this
# change's delta spec until it is synced there, plus the cleanup registry it
# points at. Never assert against observed output.
# test-check-plan-provenance.sh's header records that suite encoding the
# guard's own defects as its specification more than once, which then made each
# defect look verified.
#
# WHY THE FIXTURE IS A REAL REPOSITORY. Three of the five registry rows this
# guard reads — the worktree registration, the local branch and the
# remote-tracking ref — exist only inside git's own bookkeeping. A fixture tree
# of plain files could assert the other two and would report a passing suite
# over a guard that never asked git anything.
#
# WHY SOME CASES ASSERT ON THE REASON TEXT. The requirement asks for the
# leftover to be NAMED, not merely for the word LEFTOVER. A prefix assertion
# alone cannot tell "this row survived" from "some row survived", so a guard
# that collapsed all five rows into one generic sentence would pass every prefix
# case while leaving the operator nothing to act on. The needles below are
# therefore short and behavioural — evidence that the named row reached the
# breakdown, not a transcript of its prose.
#
# WHY THIS DUPLICATES test-check-unfinished-work.sh's HELPERS instead of
# sharing them. The two guards answer different questions at different points
# in the run and are deliberately kept independent; a shared harness library
# would couple their two suites so that a change to one guard's contract could
# only be made by editing a file the other one also runs. The duplication is
# the cheaper of the two, and it is recorded here rather than left unexplained.
#
# Bash 3.2 is the floor, as test-check-finish-preflight.sh's header records:
# indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-cleanup-complete.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# An indexed array, not a space-separated string: sandbox paths come from
# mktemp under TMPDIR, which may contain spaces, and word-splitting a string
# would then rm -rf the fragments.
SANDBOXES=()
cleanup() {
  # ${SANDBOXES[@]} is unset-expansion-unsafe under `set -u` on bash 3.2 when empty.
  [ "${#SANDBOXES[@]}" -eq 0 ] && return 0
  for s in "${SANDBOXES[@]}"; do rm -rf "$s"; done
}
trap cleanup EXIT

WORK="$(mktemp -d "${TMPDIR:-/tmp}/cleanup-complete-test.XXXXXX")"
SANDBOXES+=("$WORK")
ERRFILE="$WORK/stderr"

# run_guard <repo> <change-name> <state-dir> -> sets OUT (stdout only), ERR, RC.
# The two streams are captured SEPARATELY rather than merged with 2>&1,
# because the contract distinguishes them: a refusal puts its message on
# stderr and must leave stdout empty, and a merged capture cannot tell an
# empty stdout from a stdout carrying the message.
run_guard() {
  set +e
  OUT="$("$GUARD" "$@" 2>"$ERRFILE")"
  RC=$?
  set -e
  ERR="$(cat "$ERRFILE")"
}

# assert_verdict <expected-prefix> <label> — exit 0, exactly one stdout line,
# and that line begins with the expected verdict token.
assert_verdict() {
  local want="$1" label="$2" lines
  if [ "$RC" -ne 0 ]; then
    fail "$label: expected exit 0, got rc=$RC out=$OUT err=$ERR"
    return 0
  fi
  if [ -z "$OUT" ]; then
    fail "$label: expected a verdict line, got empty stdout (err=$ERR)"
    return 0
  fi
  lines="$(printf '%s\n' "$OUT" | wc -l | tr -d ' ')"
  if [ "$lines" != "1" ]; then
    fail "$label: expected exactly one stdout line, got $lines: $OUT"
    return 0
  fi
  case "$OUT" in
    "$want"*) pass "$label" ;;
    *) fail "$label: expected a line beginning $want, got: $OUT" ;;
  esac
}

# assert_reason <needle> <label> — the breakdown names this registry row.
assert_reason() {
  case "$OUT" in
    *"$1"*) pass "$2" ;;
    *) fail "$2: the breakdown does not name the row ($1): $OUT" ;;
  esac
}

# assert_no_reason <needle> <label> — the breakdown does NOT name this row.
# A leftover that is reported is only useful if it is reported accurately: a
# guard that named a row which is in fact gone sends the operator to remove
# something that is not there, and every prefix assertion still passes while it
# does. The five rows are read from five different places precisely so they can
# be attributed one at a time.
assert_no_reason() {
  case "$OUT" in
    *"$1"*) fail "$2: the breakdown names a row that is gone ($1): $OUT" ;;
    *) pass "$2" ;;
  esac
}

# assert_no_verdict <label> — the refusal shape: non-zero, nothing on stdout,
# and the guard's own name on stderr. The needle carries the colon
# deliberately. Without it the shell's own "…/check-cleanup-complete.sh: No
# such file or directory" satisfies the case, so it passes while the guard does
# not exist — a vacuous assertion that proves nothing about the guard's own
# reporting.
assert_no_verdict() {
  local label="$1"
  [ "$RC" -ne 0 ] && pass "$label: exits non-zero" \
    || fail "$label: expected a non-zero exit, got rc=$RC out=$OUT"
  [ -z "$OUT" ] && pass "$label: writes nothing to stdout" \
    || fail "$label: emitted a verdict line: $OUT"
  case "$ERR" in
    *"check-cleanup-complete: "*) pass "$label: names the failure on stderr" ;;
    *) fail "$label: no named message on stderr: $ERR" ;;
  esac
}

# new_fixture -> sets REPO and STATE to a repository and state directory after
# a SUCCESSFUL run-2 cleanup of the change "demo": no worktree registered for
# openspec/demo, no local branch, no remote-tracking ref, no proposal artifact
# source, and the change directory moved under openspec/changes/archive/.
#
# The archived copy is present on purpose. The registry row says the change
# directory is MOVED to the archive and never deleted, so a guard that demanded
# an empty openspec/changes/ tree would report LEFTOVER on every correctly
# finished change.
new_fixture() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/cleanup-complete-repo.XXXXXX")"
  STATE="$(mktemp -d "${TMPDIR:-/tmp}/cleanup-complete-state.XXXXXX")"
  SANDBOXES+=("$REPO" "$STATE")
  git -C "$REPO" init -q -b main
  git -C "$REPO" config user.email test@example.invalid
  git -C "$REPO" config user.name "Test"
  mkdir -p "$REPO/openspec/changes/archive/2026-01-01-demo"
  printf 'base\n' > "$REPO/f.txt"
  git -C "$REPO" add -A
  git -C "$REPO" commit -qm base
}

# add_worktree <branch> [leaf] -> sets WT_PATH to a linked worktree of $REPO
# checked out on that branch. The parent directory is a sandbox of its own so
# the worktree is removed with everything else on exit. The leaf name is a
# parameter only so one case can put a space in it.
add_worktree() {
  local holder
  holder="$(mktemp -d "${TMPDIR:-/tmp}/cleanup-complete-wt.XXXXXX")"
  SANDBOXES+=("$holder")
  WT_PATH="$holder/${2:-tree}"
  git -C "$REPO" worktree add -q -b "$1" "$WT_PATH" >/dev/null
}

# 1. The whole point of the COMPLETE verdict: a finished cleanup is confirmed,
#    and the archived change directory is not mistaken for a leftover.
new_fixture
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a fully cleaned repository is COMPLETE"

# 2. Registry row — the worktree. Found by branch, exactly as the contract says
#    worktrees are found when the state file's map is absent.
new_fixture
add_worktree openspec/demo
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "a surviving worktree is LEFTOVER"
assert_reason "worktree" "a surviving worktree names its row"

# 2b. A worktree whose directory was deleted without `git worktree prune` is
#     still REGISTERED, and the registration is what run 2 is required to
#     remove. Case 2 passes for a guard that only tested whether the directory
#     exists, so without this case a half-finished cleanup reads as COMPLETE.
new_fixture
add_worktree openspec/demo
rm -rf "$WT_PATH"
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "a registered but pruned-away worktree is LEFTOVER"
assert_reason "worktree" "a registered but pruned-away worktree names its row"

# 2c. The reported path survives a space in it. `worktree list --porcelain`
#     emits the path raw, so a guard that reads it as a whitespace-delimited
#     field truncates it and hands the operator a path that does not exist —
#     while still saying LEFTOVER, which is why cases 2 and 2b pass over that
#     defect. Worktrees live under TMPDIR or a home directory, both of which
#     can contain a space.
#     The expected path is resolved with `pwd -P` first: git records a
#     worktree by its physical path, so under a TMPDIR reached through a
#     symlink — /var -> /private/var on macOS — the fixture's own path is not
#     the one git reports, and asserting the unresolved one would fail against
#     a correct guard.
new_fixture
add_worktree openspec/demo "tree with space"
WT_REAL="$(cd "$WT_PATH" && pwd -P)"
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "a worktree whose path contains a space is LEFTOVER"
assert_reason "$WT_REAL" "the breakdown names the worktree's whole path"

# 3. Registry row — the local branch.
new_fixture
git -C "$REPO" branch openspec/demo
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "a surviving local branch is LEFTOVER"
assert_reason "local branch" "a surviving local branch names its row"
assert_no_reason "remote-tracking ref" "a surviving local branch is not read as a remote one"

# 4. Registry row — the remote branch, seen locally as its tracking ref. This
#    row is new in this change: run 2 never deleted the remote branch before,
#    so a guard that omitted it would confirm a cleanup that left one behind.
new_fixture
git -C "$REPO" update-ref refs/remotes/origin/openspec/demo HEAD
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "a surviving remote-tracking ref is LEFTOVER"
assert_reason "remote-tracking ref" "a surviving remote-tracking ref names its row"
#     A ref under refs/remotes/ is not a local branch. Without this assertion a
#     guard that resolved the branch by pattern rather than by full ref name
#     reports both rows from one surviving ref, and the operator is sent after
#     a local branch that was deleted correctly.
assert_no_reason "local branch" "a surviving remote-tracking ref is not read as a local branch"

# 5. Registry row — the change directory, which must have been MOVED into the
#    archive. One still sitting at openspec/changes/<name>/ means the archive
#    step did not happen.
new_fixture
mkdir -p "$REPO/openspec/changes/demo"
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "an unarchived change directory is LEFTOVER"
assert_reason "openspec/changes/demo" "an unarchived change directory names its row"

# 6. Registry row — the proposal artifact source in the state directory.
new_fixture
: > "$STATE/demo-proposal-artifact.html"
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "a surviving proposal artifact source is LEFTOVER"
assert_reason "proposal artifact source" "a surviving artifact source names its row"

# 7. Every row is matched by its FULL name, never by prefix. A neighbouring
#    change called demo-other owns a worktree, a branch, a tracking ref, a
#    change directory and an artifact source of its own; none of them belongs
#    to demo. Without this case a guard grepping for `openspec/demo` reports
#    LEFTOVER over another change's live work and sends the operator hunting
#    for something that is not there.
new_fixture
add_worktree openspec/demo-other
git -C "$REPO" update-ref refs/remotes/origin/openspec/demo-other HEAD
mkdir -p "$REPO/openspec/changes/demo-other"
: > "$STATE/demo-other-proposal-artifact.html"
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "another change's artifacts are not this change's leftovers"

# 8. Several leftovers at once are reported together, on the one line the
#    contract allows, so the operator sees everything that remains in one
#    report rather than one item per finish run.
new_fixture
git -C "$REPO" branch openspec/demo
mkdir -p "$REPO/openspec/changes/demo"
: > "$STATE/demo-proposal-artifact.html"
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "several leftovers still produce exactly one line"
assert_reason "local branch" "the combined breakdown names the branch"
assert_reason "openspec/changes/demo" "the combined breakdown names the change directory"
assert_reason "proposal artifact source" "the combined breakdown names the artifact source"

# 9. A repository that cannot be read is not a verdict: exit non-zero, nothing
#    on stdout, and the reason on stderr — so a caller grepping stdout for
#    COMPLETE cannot read a confirmed cleanup out of a failure.
new_fixture
run_guard "/nonexistent/repo" demo "$STATE"
assert_no_verdict "an unreadable repository"

# 9b. A directory that exists but is not a git repository. Case 9 passes for a
#     guard that only tests -d, and such a guard would answer COMPLETE for
#     every one of the three rows it can only learn from git.
new_fixture
run_guard "$WORK" demo "$STATE"
assert_no_verdict "a directory that is not a git repository"

# 10. A state directory that is not there cannot answer whether the artifact
#     source is gone. Reading "absent" out of a path that was never readable is
#     the false COMPLETE this guard exists to prevent — a mistyped state
#     directory would otherwise confirm a cleanup nobody checked.
new_fixture
run_guard "$REPO" demo "/nonexistent/state"
assert_no_verdict "an unreadable state directory"

# 11. A missing argument is programmer error, reported rather than guessed at.
run_guard "" "" ""
[ "$RC" -eq 2 ] && pass "missing arguments -> exit 2" \
  || fail "missing arguments: expected exit 2, got rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "missing arguments: emits no verdict line" \
  || fail "missing arguments: emitted a verdict line: $OUT"

# 12. The change name is PR-controlled — it reaches here from a state file
#     anyone able to open a pull request can edit — and it is concatenated into
#     a path, a ref name and a `-d` test. `../../nonexistent-decoy` made every
#     row answer "already gone" and the guard reported COMPLETE while the real
#     change directory and artifact source were still there. The allowlist is
#     preserve-session-records.sh's Protection 1; these are the same shapes
#     test-check-unfinished-work.sh rejects, asserted here too so the two copies
#     of the rule cannot drift apart in silence.
for bad_name in "../../nonexistent-decoy" "demo*" "demo/../demo" ".hidden" "demo?x"; do
  new_fixture
  run_guard "$REPO" "$bad_name" "$STATE"
  [ "$RC" -eq 2 ] && pass "a change name outside the allowlist ($bad_name) -> exit 2" \
    || fail "change name $bad_name: expected exit 2, got rc=$RC out=$OUT"
  [ -z "$OUT" ] && pass "a change name outside the allowlist ($bad_name) emits no verdict line" \
    || fail "change name $bad_name: emitted a verdict line: $OUT"
  case "$ERR" in
    *"check-cleanup-complete: change name"*) pass "a rejected change name ($bad_name) names the failure" ;;
    *) fail "change name $bad_name: no named message on stderr: $ERR" ;;
  esac
done

# 12b. The decoy, demonstrated end to end rather than asserted from the message:
#      a real change directory and a real artifact source survive, and no name
#      may produce COMPLETE over them.
new_fixture
mkdir -p "$REPO/openspec/changes/demo"
: > "$STATE/demo-proposal-artifact.html"
run_guard "$REPO" "../../nonexistent-decoy" "$STATE"
[ -z "$OUT" ] && pass "a traversal-shaped name cannot report COMPLETE over a live change" \
  || fail "traversal: the guard produced a verdict for a name outside the repository: $OUT"

# 13. git's own failure is not read as an empty list. `|| true` on the worktree
#     listing would turn an unreadable repository into COMPLETE — the
#     reassuring verdict — which is the same silence this guard was added to
#     break. Real git could not be made to fail that call from a fixture
#     (a permission-stripped .git/worktrees and a bare repository were both
#     tried and both still answered), so the failure is injected: a stub named
#     `git` earlier on PATH fails `worktree list` and delegates everything else
#     to the real binary, which keeps the -d and rev-parse checks above honest.
REAL_GIT="$(command -v git)"
STUB_DIR="$WORK/gitstub"
mkdir -p "$STUB_DIR"
cat > "$STUB_DIR/git" <<STUB
#!/usr/bin/env bash
for arg in "\$@"; do
  if [ "\$arg" = "worktree" ]; then
    echo "fatal: injected failure — cannot list worktrees" >&2
    exit 128
  fi
done
exec "$REAL_GIT" "\$@"
STUB
chmod +x "$STUB_DIR/git"
new_fixture
OLD_PATH="$PATH"
PATH="$STUB_DIR:$PATH"
run_guard "$REPO" demo "$STATE"
PATH="$OLD_PATH"
assert_no_verdict "a failed worktree listing"

# 14. The five rows this guard checks are DERIVED from the registry in
#     skills/myflow-contracts/pipeline.md, and a derivation nothing checks is a
#     copy waiting to go stale: a registry row added later whose lifetime ends
#     at run 2 would go unchecked and this guard would confirm a cleanup that
#     left it behind. The guard therefore declares, in `registry-row-checked:`
#     and `registry-row-not-checked:` markers, which registry row each of its
#     decisions covers, and this case reads the real registry and asserts the
#     two sets match IN BOTH DIRECTIONS.
#
#     This is the one case that reads the repository tree rather than a
#     sandbox — read-only, and unavoidably so: the coupling is to the real
#     registry, and a fixture copy of it would be a third place for the same
#     list to drift.
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REGISTRY_FILE="$REPO_ROOT/skills/myflow-contracts/pipeline.md"
if [ ! -f "$REGISTRY_FILE" ]; then
  fail "registry coupling: $REGISTRY_FILE not found — the guard's rows have nothing to be derived from"
else
  awk '
    /^## Temporary artifacts registry/ { in_reg = 1; next }
    in_reg && /^## / { in_reg = 0 }
    in_reg && /^\|/ {
      # Registry row names carry no pipe, so a plain split is enough here; the
      # findings-table parser in check-unfinished-work.sh needs the escaping
      # rules because a finding note quotes shell.
      split($0, f, "|")
      name = f[2]
      gsub(/^[[:space:]]+/, "", name); gsub(/[[:space:]]+$/, "", name)
      if (name == "" || name == "Artifact" || name ~ /^:?-+:?$/) next
      print name
    }
  ' "$REGISTRY_FILE" | sort > "$WORK/registry-rows"

  awk '
    /^# registry-row-checked: / {
      s = $0; sub(/^# registry-row-checked: /, "", s); print s; next
    }
    /^# registry-row-not-checked: / {
      s = $0; sub(/^# registry-row-not-checked: /, "", s)
      # The reason follows the row name after " — "; index() on a literal keeps
      # this off multibyte regex behaviour.
      p = index(s, " — ")
      if (p > 0) s = substr(s, 1, p - 1)
      print s; next
    }
  ' "$GUARD" | sort > "$WORK/declared-rows"

  if [ ! -s "$WORK/registry-rows" ]; then
    fail "registry coupling: no rows parsed out of the registry in $REGISTRY_FILE"
  elif [ ! -s "$WORK/declared-rows" ]; then
    fail "registry coupling: the guard declares no registry rows at all"
  else
    MISSING_DECL="$(comm -23 "$WORK/registry-rows" "$WORK/declared-rows")"
    STALE_DECL="$(comm -13 "$WORK/registry-rows" "$WORK/declared-rows")"
    [ -z "$MISSING_DECL" ] \
      && pass "every registry row is declared checked or not-checked by the guard" \
      || fail "registry coupling: registry row(s) the guard declares nothing about: $(printf '%s' "$MISSING_DECL" | tr '\n' ';')"
    [ -z "$STALE_DECL" ] \
      && pass "the guard declares no row the registry no longer has" \
      || fail "registry coupling: the guard declares row(s) absent from the registry: $(printf '%s' "$STALE_DECL" | tr '\n' ';')"
  fi
fi

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-cleanup-complete: all cases pass\n'
