#!/usr/bin/env bash
# Assertion harness for check-cleanup-complete.sh. Builds throwaway git
# repositories and state directories under a sandboxed TMPDIR and asserts the
# guard's verdict and exit status. Never MODIFIES the real repository tree —
# with two deliberate reads, both for the same reason: the thing under test is
# the guard's agreement with a canonical file, and a fixture copy of that file
# would be a third place for the same rule to drift.
#   - the registry-coupling case reads skills/myflow-contracts/pipeline.md;
#   - the workspace-id case reads skills/myflow-contracts/workspace-isolation.md
#     and RUNS the derivation block it carries, asserting the guard's own id
#     against it byte for byte.
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

# skip <label> <reason> — a case this environment cannot decide, reported as
# neither a pass nor a failure. Every case that uses it turns on a file mode of
# 000 still being readable by root, so the unreadable-configuration case and the
# unreadable-ref case have nothing to assert when the suite runs as root; the
# ref case skips for a second environmental reason too, a git that stores the
# branch packed rather than loose. Printing "ok" there would be a green
# result for an assertion that never ran, which is the vacuous pass every guard
# in this repository is written against.
skip() { printf 'skip: %s (%s)\n' "$1" "$2"; }

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

# run_guard_bounded <seconds> <repo> <change-name> <state-dir> — run_guard with
# the guard's survivor-command bound shortened for the duration of the call.
#
# WHY THE SEAM EXISTS, AND WHY IT IS NOT A WEAKENED CONFIGURATION. The bound the
# guard ships is 60 seconds, and a case that exercised it literally would cost a
# minute of wall clock on every run of this suite — so the timeout cases below
# would be the slowest thing in the repository and would be the first thing
# someone deleted. The override is the same opt-in idiom check-references.sh
# already uses for CHECK_REFERENCES_ROOT, and it can only ever shorten: the
# guard validates it as a positive integer and falls back to its own default
# otherwise, so no value of it removes the bound or lets the guard report a
# workspace as verified that it did not verify. A shorter bound produces MORE
# skips, never fewer, and a skip is reported rather than passed.
#
# The assignment is written out rather than prefixed onto the call, because a
# `VAR=x func` prefix on a shell FUNCTION does not reliably unset itself
# afterwards across shells — and a bound that leaked into the next case would
# silently shorten it.
run_guard_bounded() {
  local secs="$1"; shift
  CHECK_CLEANUP_SURVIVORS_TIMEOUT="$secs"
  export CHECK_CLEANUP_SURVIVORS_TIMEOUT
  run_guard "$@"
  unset CHECK_CLEANUP_SURVIVORS_TIMEOUT
}

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

# run_guard_in_locale <locale> <repo> <change-name> <state-dir> — run_guard with
# LC_ALL and LANG set for the guard's own process and for NOTHING else.
#
# THE ASSIGNMENT IS A PREFIX ON AN EXTERNAL COMMAND, which is the one form of it
# that cannot leak: it scopes the variables to that command's environment and
# leaves this shell's untouched. The same prefix on a shell FUNCTION does not
# reliably unset itself afterwards, which is the hazard run_guard_bounded
# records — and exporting-then-unsetting around the call is worse again here,
# because `unset LANG` does not RESTORE this harness's ambient locale, it removes
# it. Every case after such a loop would then run under C without saying so, and
# the first casualty is case 21 below: it would read as an assertion about the
# guard while being an assertion about a locale the loop above it had silently
# imposed.
run_guard_in_locale() {
  local loc="$1"; shift
  set +e
  OUT="$(LC_ALL="$loc" LANG="$loc" "$GUARD" "$@" 2>"$ERRFILE")"
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

# assert_out <needle> <label> / assert_not_out <needle> <label> — the verdict
# line carries, or does not carry, this text anywhere in it.
#
# assert_reason above is the same test named for a LEFTOVER breakdown. The
# workspace row needs both namings because it is the one row whose "verified"
# and "skipped" notes land on a COMPLETE line, which is not a breakdown: a
# skipped verification is reported rather than passed, and a reader who found
# that note under the word "reason" would read it as a leftover.
assert_out() {
  case "$OUT" in
    *"$1"*) pass "$2" ;;
    *) fail "$2: the verdict line does not carry '$1': $OUT" ;;
  esac
}

assert_not_out() {
  case "$OUT" in
    *"$1"*) fail "$2: the verdict line carries '$1': $OUT" ;;
    *) pass "$2" ;;
  esac
}

# assert_absent <path> <label> — nothing created that file. Used to prove a
# command was NOT run, which several cases below turn on: this guard asks the
# project which resources survived, and a guard that ran the project's `remove`
# command, or inferred an answer from it, leaves a canary behind.
assert_absent() {
  [ ! -e "$1" ] && pass "$2" || fail "$2: $1 exists, so the command that creates it was run"
}

# assert_present <path> <label> — that file was created, i.e. the command was
# run. Without this, every "verified clean" case below also passes for a guard
# that executes nothing at all.
assert_present() {
  [ -f "$1" ] && pass "$2" || fail "$2: $1 was never created, so the command was not run"
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

# declare_isolation <survivors-cell> -> writes $REPO/.myflow/project.md with a
# `## workspace isolation` section. The argument is the `Runs` cell of the
# `survivors` row, written exactly as a project would write it — backticks
# included, because the guard has to strip them. An EMPTY argument declares
# `create` and `remove` and no `survivors` row, which is its own case.
#
# The `create` and `remove` rows each leave a canary file. Nothing in this
# guard's contract may run either of them: it verifies by ASKING, so a guard
# that ran `remove` itself, or read an answer out of it, is caught by the canary
# rather than by a verdict that would look identical.
declare_isolation() {
  mkdir -p "$REPO/.myflow"
  {
    printf '# fixture project configuration\n\n'
    printf '## workspace isolation\n\n'
    printf '| Resource | Variable | Default | In a workspace |\n'
    printf '|----------|----------|---------|----------------|\n'
    printf '| `database` | `DB_URL` | `appdb` | `appdb_<id_underscored>` |\n'
    printf '\n'
    printf '| Command | Runs |\n'
    printf '|---------|------|\n'
    printf '| `create` | `touch %s` |\n' "'$REPO/create-ran'"
    printf '| `remove` | `touch %s` |\n' "'$REPO/remove-ran'"
    if [ -n "${1:-}" ]; then
      printf '| `survivors` | %s |\n' "$1"
    fi
  } > "$REPO/.myflow/project.md"
}

# write_script <name> <body> -> an executable script in the repository root.
# Cases declare it as `./<name>`, which asserts one thing beyond the body: a
# declared command runs with the repository as its working directory, exactly as
# the `./scripts/workspace survivors` of the contract's worked example needs.
write_script() {
  printf '#!/usr/bin/env bash\n%s\n' "$2" > "$REPO/$1"
  chmod +x "$REPO/$1"
}

# 1. The whole point of the COMPLETE verdict: a finished cleanup is confirmed,
#    and the archived change directory is not mistaken for a leftover.
new_fixture
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a fully cleaned repository is COMPLETE"

# 2. Registry row — the worktree. Found by branch, exactly as the contract says
#    worktrees are found when the state file's map is absent or empty.
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

# 4b-4d. A REF GIT CANNOT READ IS NOT A REF THAT DOES NOT EXIST. `show-ref
#    --verify --quiet` answers 1 for both, so a ref whose loose file is corrupt
#    or unreadable — a half-written file, a transient repack, a permission
#    change — read as "that branch is gone" and fed a COMPLETE that nothing had
#    checked. This is the guard's own "absence from a path that was never
#    readable must never be inferred" applied to the two rows that live in git's
#    ref store rather than on the filesystem.
#
#    THE EXPECTED ANSWER IS A SKIP, NOT A LEFTOVER. The row is unanswered, not
#    answered "still there": a LEFTOVER would send the operator to delete a
#    branch that may already be gone, and would block an already-merged change
#    over a condition it cannot fix from that session. SKIPPED is the answer this
#    guard already gives every input it cannot resolve, and it rides a COMPLETE
#    line because a skip does not block.
#
#    THE NEEDLE IS THE REF NAME, not the row's English name: the COMPLETE
#    sentence itself contains the words "local branch" and "remote-tracking
#    ref", so asserting on those would pass over the very fail-open this case
#    exists to catch.

# 4b. A corrupt loose ref — content git cannot resolve. This variant needs no
#     permission trickery, so it runs in every environment this suite runs in.
new_fixture
git -C "$REPO" branch openspec/demo
printf 'garbage\n' > "$REPO/.git/refs/heads/openspec/demo"
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a corrupt local branch ref does not block the terminal state"
assert_out "SKIPPED" "a corrupt local branch ref is reported as skipped"
assert_out "refs/heads/openspec/demo" "the skip names the ref it could not read"

# 4c. The same class on the remote-tracking row, which is read the same way and
#     would otherwise carry the same fail-open unnoticed.
new_fixture
git -C "$REPO" update-ref refs/remotes/origin/openspec/demo HEAD
printf 'garbage\n' > "$REPO/.git/refs/remotes/origin/openspec/demo"
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a corrupt remote-tracking ref does not block the terminal state"
assert_out "SKIPPED" "a corrupt remote-tracking ref is reported as skipped"
assert_out "refs/remotes/origin/openspec/demo" "the skip names the remote-tracking ref it could not read"

# 4d. An UNREADABLE loose ref — the permission half of the same class, which is
#     the shape a real repack or a tightened umask produces. Skipped rather than
#     asserted when this process can read a mode-000 file, for the reason the
#     `skip` helper's comment gives.
new_fixture
git -C "$REPO" branch openspec/demo
REF_FILE="$REPO/.git/refs/heads/openspec/demo"
if [ ! -f "$REF_FILE" ]; then
  skip "an unreadable local branch ref is reported as skipped" \
    "this git stores the ref packed rather than loose"
else
  chmod 000 "$REF_FILE"
  if [ -r "$REF_FILE" ]; then
    chmod 644 "$REF_FILE"
    skip "an unreadable local branch ref is reported as skipped" \
      "this process can read a mode-000 file"
  else
    run_guard "$REPO" demo "$STATE"
    chmod 644 "$REF_FILE"
    assert_verdict "COMPLETE:" "an unreadable local branch ref does not block the terminal state"
    assert_out "SKIPPED" "an unreadable local branch ref is reported as skipped"
    assert_out "refs/heads/openspec/demo" "the skip names the unreadable ref"
  fi
fi

# 4e. BACKWARDS COMPATIBILITY OF THE SAME CODE PATH, proved rather than argued:
#     a repository whose refs are all readable answers exactly as it did before,
#     with no skip note anywhere. The disambiguation runs on every invocation, so
#     a version of it that mistook a healthy repository for an unreadable one
#     would turn every clean run into an unverified one.
new_fixture
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a repository with no branch at all is still COMPLETE"
assert_not_out "SKIPPED" "a readable ref store produces no skip note"

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
#     records.Destination's Protection 1 (stats/internal/records/render.go);
#     these are the same shapes test-check-unfinished-work.sh rejects, asserted
#     here too so the two copies of the rule cannot drift apart in silence.
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

# ---------------------------------------------------------------------------
# 15-27. The workspace database and bucket row.
#
# This is the one registry row whose verdict is not read out of git or the
# filesystem but ASKED of the project, by running the `survivors` command its
# .myflow/project.md declares. skills/myflow-contracts/project-configuration.md
# is normative for that command's output and its exit code — "What `survivors`
# prints, and what its exit code means" — and the cases below are one per branch
# of it, plus the branches its neighbouring rules imply.
#
# THE ASYMMETRY THESE CASES EXIST TO PIN DOWN. A reported survivor BLOCKS the
# terminal state; an unreachable service does NOT. Both are non-empty outcomes
# of the same command and a guard that collapsed them would either strand an
# already-merged change over a stopped database (case 18) or confirm a cleanup
# it never verified (case 17). The exit code is what separates them, and
# nothing else is.
# ---------------------------------------------------------------------------

# 15. The overwhelmingly common case, including this repository: a project with
#     no configuration file at all. It must behave exactly as it did before this
#     row existed — same verdict, and no workspace note of any kind on it.
new_fixture
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a project with no .myflow/project.md is COMPLETE"
assert_not_out "workspace" "a project with no configuration says nothing about a workspace"

# 15b. A configuration that declares no isolation. The decoy command table sits
#      under a DIFFERENT heading and would leave a canary if it ran, so this
#      case fails a guard that scans the file for a `survivors` row instead of
#      scanning the `## workspace isolation` section for one — which is the
#      difference between reading a declaration and finding a string.
new_fixture
mkdir -p "$REPO/.myflow"
{
  printf '# fixture project configuration\n\n'
  printf '## test\n\n'
  printf '| Command | Runs |\n'
  printf '|---------|------|\n'
  printf '| `survivors` | `touch %s` |\n' "'$REPO/decoy-ran'"
} > "$REPO/.myflow/project.md"
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a project declaring no isolation is COMPLETE"
assert_not_out "workspace" "a project declaring no isolation says nothing about a workspace"
assert_absent "$REPO/decoy-ran" "a command outside the isolation section is never run"

# 16. Exit 0 with empty output — the ONLY result that verifies the removal. The
#     canary assertions are what stop this case from also passing for a guard
#     that runs nothing: an empty report and no report at all produce the same
#     verdict and must not produce the same evidence.
new_fixture
write_script survivors.sh "touch '$REPO/survivors-ran'"
declare_isolation '`./survivors.sh`'
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "an empty survivor report is COMPLETE"
assert_present "$REPO/survivors-ran" "the declared survivors command is actually run"
assert_absent "$REPO/remove-ran" "the guard never runs the project's remove command"
assert_absent "$REPO/create-ran" "the guard never runs the project's create command"
assert_out "survivor report" "an empty survivor report is reported as verified"

# 16b. A report that is empty apart from blank lines is still empty. A trailing
#      newline is what every well-behaved command emits, and a guard that
#      counted blank lines would report a leftover named nothing at all — a
#      blocked terminal state with no resource for the operator to remove.
new_fixture
write_script survivors.sh "printf '\n\n'"
declare_isolation '`./survivors.sh`'
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a survivor report of blank lines is COMPLETE"

# 17. Exit 0 WITH output — every line is a survivor, and a survivor blocks the
#     terminal state. Two lines, because one line cannot tell a guard that
#     reports the whole report from one that reports each line.
new_fixture
write_script survivors.sh "printf 'db-survivor-one\ndb-survivor-two\n'"
declare_isolation '`./survivors.sh`'
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "a reported survivor is LEFTOVER"
assert_reason "db-survivor-one" "the breakdown names the first survivor"
assert_reason "db-survivor-two" "the breakdown names the second survivor"
#     Run under the SHIPPED bound, with the harness's seam deliberately unset:
#     an ordinary report is not a timeout, so the shipped default is not a hair
#     trigger. Every other case in this block asserts the same thing for free.
assert_not_out "timed out" "an ordinary survivor report is not reported as a timeout"

# 18. Any non-zero exit — the service could not be reached, so the check says
#     nothing about survivors. Reported by name AND with the exit code, and the
#     run continues: blocking an already-merged change over a stopped database
#     is what the contract forbids.
new_fixture
write_script survivors.sh "exit 7"
declare_isolation '`./survivors.sh`'
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "an unreachable service does not block the terminal state"
assert_out "SKIPPED" "an unreachable service is reported as skipped"
assert_out "./survivors.sh" "the skip names the command"
#     The needle carries the word as well as the number: a bare "7" is
#     satisfied by the sandbox path in the verdict line, so it passed while the
#     guard reported no exit code at all.
assert_out "exited 7" "the skip carries the exit code"
#     An ordinary non-zero exit is NOT a timeout, and the operator has to be
#     able to tell "your command is broken" from "your command is slow" — the
#     two need different remedies and only one of them is worth waiting for.
#     Cases 28 and 28b assert the other direction.
assert_not_out "timed out" "an ordinary non-zero exit is not reported as a timeout"

# 18b. A non-zero exit says NOTHING about survivors, so whatever it printed is
#      not a survivor list. Without this case a guard that reads stdout before
#      the exit code turns a stopped service into a blocked terminal state —
#      the exact inversion of the rule case 18 asserts.
new_fixture
write_script survivors.sh "printf 'noise-not-a-survivor\n'; exit 3"
declare_isolation '`./survivors.sh`'
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "output from a failed survivor report is not a leftover"
assert_not_out "noise-not-a-survivor" "output from a failed survivor report is not read as a survivor"
assert_out "SKIPPED" "a failed survivor report that printed something is still a skip"

# 19. A project that declares `create` and `remove` but no `survivors`. The
#     verification is reported as SKIPPED, never as passed — and the canaries
#     prove the guard did not fall back to running `remove` and reading its exit
#     code, which is the inference the requirement forbids by name.
new_fixture
declare_isolation ''
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "no survivors command does not block the terminal state"
assert_out "SKIPPED" "no survivors command is reported as skipped, not as passed"
assert_absent "$REPO/remove-ran" "no survivors command: the removal is not run in its place"
assert_absent "$REPO/create-ran" "no survivors command: nothing else in the table is run either"

# 19b. A `survivors` row whose command cell is empty declares no command. It is
#      reported as skipped exactly as a missing row is — never repaired, never
#      guessed at, and never turned into an empty string handed to a shell.
new_fixture
mkdir -p "$REPO/.myflow"
{
  printf '## workspace isolation\n\n'
  printf '| Command | Runs |\n'
  printf '|---------|------|\n'
  printf '| `remove` | `touch %s` |\n' "'$REPO/remove-ran'"
  printf '| `survivors` |  |\n'
} > "$REPO/.myflow/project.md"
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "an empty survivors cell does not block the terminal state"
assert_out "SKIPPED" "an empty survivors cell is reported as skipped"
assert_absent "$REPO/remove-ran" "an empty survivors cell does not fall back to the removal"

# 19c. The isolation section ENDS at the next heading. A `survivors` row in a
#      later section is that section's, and running it would be this guard
#      executing a command the project never declared for this purpose — while
#      reporting a verified workspace on the strength of it.
new_fixture
mkdir -p "$REPO/.myflow"
{
  printf '## workspace isolation\n\n'
  printf '| Command | Runs |\n'
  printf '|---------|------|\n'
  printf '| `create` | `true` |\n'
  printf '\n## test\n\n'
  printf '| Command | Runs |\n'
  printf '|---------|------|\n'
  printf '| `survivors` | `touch %s` |\n' "'$REPO/later-section-ran'"
} > "$REPO/.myflow/project.md"
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a survivors row in a later section does not block the terminal state"
assert_out "SKIPPED" "a survivors row in a later section leaves the verification skipped"
assert_absent "$REPO/later-section-ran" "a survivors row in a later section is never run"

# 19d. TWO `## workspace isolation` HEADINGS ARE AN AMBIGUOUS DECLARATION, and
#      this guard answers ambiguity the same way everywhere: it reports and
#      skips. The file is hand-written, so a heading duplicated by a bad merge
#      or a copied block is a realistic mistake, and the two sections can carry
#      two different `survivors` commands naming two different services.
#      Resolving that silently to whichever comes first runs a command the
#      author may not have meant and then reports the row VERIFIED on its
#      answer — the same false COMPLETE case 26 refuses to infer from an
#      unreadable file, reached from a readable one. Skipping is the outcome
#      cases 19b, 22 and 26 already give an input this guard cannot resolve:
#      reported by name, never passed, and it does not block a merged change.
#
#      Neither command runs. A skip that had already executed the first section's
#      command would be a skip in the verdict line only.
new_fixture
mkdir -p "$REPO/.myflow"
{
  printf '## workspace isolation\n\n'
  printf '| Command | Runs |\n'
  printf '|---------|------|\n'
  printf '| `survivors` | `touch %s` |\n' "'$REPO/first-section-ran'"
  printf '\n## workspace isolation\n\n'
  printf '| Command | Runs |\n'
  printf '|---------|------|\n'
  printf '| `survivors` | `touch %s` |\n' "'$REPO/second-section-ran'"
} > "$REPO/.myflow/project.md"
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a duplicated isolation heading does not block the terminal state"
assert_out "SKIPPED" "a duplicated isolation heading is reported as skipped, not resolved silently"
assert_out "workspace isolation" "the skip names the heading that is duplicated"
assert_out "declares 2" "the skip says how many of them it found"
assert_absent "$REPO/first-section-ran" "a duplicated heading does not run the first section's command"
assert_absent "$REPO/second-section-ran" "a duplicated heading does not run the second section's command"

# 20. The workspace id, asserted against the CANONICAL derivation rather than
#     against a copy of it — for a SET of change names, one per step of that
#     derivation, and under every locale this machine has. The block is
#     extracted from skills/myflow-contracts/workspace-isolation.md, DRIVEN with
#     each name in turn, and the guard's own id compared with what it produces,
#     byte for byte. A second implementation that disagrees by one byte yields a
#     different database name for the same change — the failure that file exists
#     to prevent — and it cannot be caught by asserting the guard against a
#     literal written here, since that literal would be a third implementation.
#
#     WHY THE BLOCK IS DRIVEN RATHER THAN RUN AS WRITTEN, and why that is the
#     whole of this case's strength. The block hard-codes one example name, and
#     that name is already pure [a-z0-9-] — so it passes through normalisation
#     unchanged, and a guard that normalises in the wrong ORDER, truncates the
#     wrong string, folds case through the ambient locale and trims one trailing
#     `-` instead of all of them still reproduces it exactly. That is not
#     hypothetical: it is what this guard did while this case reported green.
#     Pinning one ASCII-clean name pins nothing, and pinning a SHAPE — the
#     `demo-x-<4 hex>` this case used to assert for its second name — pins less
#     than nothing, because it survives every one of those four errors too.
#
#     THE REWRITE IS ASSERTED, NEVER ASSUMED. The block's single `name=`
#     assignment is rewritten to take an argument; a block that no longer
#     carries exactly one such assignment FAILS this case rather than silently
#     falling back to the one name it hard-codes, which would be a vacuous pass
#     wearing this comment.
CANON_SRC="$REPO_ROOT/skills/myflow-contracts/workspace-isolation.md"
if [ ! -f "$CANON_SRC" ]; then
  fail "workspace id: $CANON_SRC not found — the guard's derivation has nothing to be checked against"
else
  CANON_RAW="$WORK/canonical-derivation.raw.sh"
  CANON_SH="$WORK/canonical-derivation.sh"
  # Every fenced block in the canonical file that ASSIGNS `id` or
  # `id_underscored`, in file order, concatenated. Selecting by what a block
  # assigns rather than by its position keeps this working when a block is
  # added or reordered; the port-offset and lsof blocks assign neither and are
  # not extracted.
  awk '
    /^```/ {
      if (inb) { if (want) printf "%s", buf; inb = 0; buf = ""; want = 0 }
      else     { inb = 1; buf = ""; want = 0 }
      next
    }
    inb {
      buf = buf $0 "\n"
      if ($0 ~ /^id=/ || $0 ~ /^id_underscored=/) want = 1
    }
  ' "$CANON_SRC" > "$CANON_RAW"

  CANON_ASSIGNMENTS="$(grep -c '^name=' "$CANON_RAW" || true)"
  if [ "$CANON_ASSIGNMENTS" != "1" ]; then
    fail "workspace id: the block extracted from $CANON_SRC carries $CANON_ASSIGNMENTS top-level 'name=' assignments, not 1, so it cannot be driven with an arbitrary change name — pin the guard against the file's published worked values instead of letting this case shrink back to the single name the block hard-codes"
  else
    sed 's/^name=.*/name="$1"/' "$CANON_RAW" > "$CANON_SH"
    printf 'printf "%%s %%s\\n" "$id" "$id_underscored"\n' >> "$CANON_SH"

    # canon_ids <locale> <name> -> the canonical `<id> <id_underscored>` for that
    # name, with the block run under that locale. Empty when the block could not
    # be run at all, which every comparison below reports as a mismatch rather
    # than skipping over.
    #
    # THE LOCALE IS PASSED TO THE CANONICAL SIDE TOO, not only to the guard. The
    # block pins `LC_ALL=C` on both its `tr` calls, so its answer is supposed to
    # be the same under every locale — running it under one and the guard under
    # another would compare two things that differ for a reason neither side is
    # being tested for, and would quietly excuse the block from the invariance it
    # claims.
    canon_ids() { LC_ALL="$1" LANG="$1" bash "$CANON_SH" "$2" 2>/dev/null || printf ''; }

    # SELF-CHECK OF THE DRIVER, before it is trusted with anything. The block
    # publishes its worked values as `# <name> -> <id>` comments, and driving it
    # with each of those names must reproduce the id written beside it. Without
    # this, a driver whose rewrite silently failed — or a `name=` line the sed
    # matched but did not replace — would hand every comparison below the same
    # id for every name, and each would then be compared against a guard result
    # for a name the canonical side never saw.
    CANON_PUBLISHED="$WORK/canonical-published"
    awk -F' -> ' '
      /^[[:space:]]*#[[:space:]]*[^[:space:]]+[[:space:]]+->[[:space:]]+[a-z0-9-]+[[:space:]]*$/ {
        n = $1; sub(/^[[:space:]]*#[[:space:]]*/, "", n); sub(/[[:space:]]+$/, "", n)
        v = $2; sub(/[[:space:]]+$/, "", v)
        print n "\t" v
      }
    ' "$CANON_SRC" > "$CANON_PUBLISHED"

    CANON_PUBLISHED_N="$(wc -l < "$CANON_PUBLISHED" | tr -d ' ')"
    if [ "$CANON_PUBLISHED_N" -lt 3 ]; then
      fail "workspace id: $CANON_SRC publishes $CANON_PUBLISHED_N worked '<name> -> <id>' values, fewer than the 3 this self-check needs to establish that the canonical block is being driven by its argument at all"
    else
      while IFS=$'\t' read -r pub_name pub_id; do
        CANON_GOT="$(canon_ids C "$pub_name")"
        [ "${CANON_GOT%% *}" = "$pub_id" ] \
          && pass "the canonical block driven with '$pub_name' yields its published id $pub_id" \
          || fail "workspace id: the canonical block driven with '$pub_name' yields '${CANON_GOT%% *}', but $CANON_SRC publishes $pub_id"
      done < "$CANON_PUBLISHED"
    fi

    # THE COMPARISON SET, and why each name is in it. Every one exercises a step
    # of the derivation that no other one does, and each was chosen by reading
    # that derivation rather than by watching the guard:
    #   kan-15-parallel-…  pure [a-z0-9-] and long — segmenting and the 12-char
    #                      budget. The name the block already hard-codes, kept
    #                      so the strengthened case still covers what it did.
    #   Demo_X             upper case and `_` — normalisation, at a length the
    #                      truncation never reaches.
    #   KAN-99-Fix.Thing   `.` as a SEGMENT BOUNDARY, which it is only when the
    #                      normalisation runs BEFORE the segmenting. This is the
    #                      name that separates the two orders: normalise-first
    #                      keeps `kan-99-fix`, segment-first keeps `kan-99`.
    #   Trailing.__        a prefix whose normalised form ends in THREE `-` —
    #                      the name that separates "strip every trailing `-`"
    #                      from "strip one".
    #   abcdefgh-ijkl-x    a segment boundary at EXACTLY 13 characters, which is
    #                      the only shape that separates a 12-character budget
    #                      from a 13-character one. Every other name here lands
    #                      nowhere near the boundary and so pins the number not
    #                      at all.
    #   Averyverylong…     a FIRST segment longer than the budget, so there is no
    #                      boundary to fall back to and step 3's other half — the
    #                      bare 12-character truncation — is what answers.
    #   Fix-I-Bug          a bare capital `I`, the one ASCII letter whose lower
    #                      case is not `i` in every locale: a Turkish locale
    #                      folds it to `ı` (U+0131, two UTF-8 bytes, and so two
    #                      `-` after the complement). This is the name that bites
    #                      a derivation reaching for `[:upper:]`/`[:lower:]`
    #                      without `LC_ALL=C`. It is INERT where the platform's
    #                      `tr` ignores the locale for those classes — stock
    #                      macOS is one such platform — so it pins nothing here
    #                      and everything on a GNU userland. It is in the set for
    #                      the machine the guard gets copied to, not for this one.
    # A non-ASCII name is deliberately absent and case 21 records why: the
    # guard's own change-name allowlist refuses one before the workspace row is
    # ever reached, so there is no guard-side id to compare.
    CANON_NAMES=(
      "kan-15-parallel-myflow-do-task-lanes"
      "Demo_X"
      "KAN-99-Fix.Thing"
      "Trailing.__"
      "abcdefgh-ijkl-x"
      "Averyverylongfirstsegment"
      "Fix-I-Bug"
    )

    # THE LOCALES, and why the comparison is run once per locale rather than
    # once. Step 2 of the derivation is byte-wise by contract, and the canonical
    # block spends two `LC_ALL=C` prefixes saying so — which is only worth
    # asserting against a locale whose case table is NOT ASCII's. A locale this
    # machine lacks is skipped rather than forced: LC_ALL naming a missing
    # locale silently falls back to C, and the case would then pass by not
    # running. `locale charmap` is the probe because it reports the codeset the
    # locale actually resolved to, so a missing UTF-8 locale answers US-ASCII
    # and is caught.
    CANON_LOCALES=(C)
    for canon_loc in en_US.UTF-8 tr_TR.UTF-8; do
      if [ "$(LC_ALL="$canon_loc" locale charmap 2>/dev/null)" = "UTF-8" ]; then
        CANON_LOCALES+=("$canon_loc")
      else
        skip "workspace id under LC_ALL=$canon_loc" "this machine has no such locale"
      fi
    done

    new_fixture
    cat > "$REPO/record.sh" <<'SH'
#!/usr/bin/env bash
printf '%s %s\n' "$1" "$2" > "$(dirname "$0")/derived-id"
SH
    chmod +x "$REPO/record.sh"
    declare_isolation '`./record.sh "<id>" "<id_underscored>"`'

    # One fixture for the whole matrix. The guard only READS the repository, and
    # record.sh overwrites its own output, so a fresh git init per name would
    # buy nothing but wall clock — and this loop is the largest thing in the
    # suite by run count.
    for canon_loc in "${CANON_LOCALES[@]}"; do
      for canon_name in "${CANON_NAMES[@]}"; do
        rm -f "$REPO/derived-id"
        run_guard_in_locale "$canon_loc" "$REPO" "$canon_name" "$STATE"
        assert_verdict "COMPLETE:" "'$canon_name' reaches a verdict under LC_ALL=$canon_loc"
        if [ ! -f "$REPO/derived-id" ]; then
          fail "workspace id (LC_ALL=$canon_loc): the guard never substituted the tokens for '$canon_name' — no record was written"
        else
          GUARD_IDS="$(cat "$REPO/derived-id")"
          CANON_EXPECTED="$(canon_ids "$canon_loc" "$canon_name")"
          [ "$GUARD_IDS" = "$CANON_EXPECTED" ] \
            && pass "'$canon_name' derives '$GUARD_IDS', matching the canonical block (LC_ALL=$canon_loc)" \
            || fail "workspace id (LC_ALL=$canon_loc): the guard derived '$GUARD_IDS' for '$canon_name', the canonical block derives '$CANON_EXPECTED'"
        fi
      done
    done

    # 21. A NON-ASCII change name is refused under EVERY locale, and this case
    #     exists because it once was not. It lives inside case 20's block because
    #     it uses the same per-locale driver, and it is asserted per locale
    #     rather than left to whichever one the suite happens to inherit.
    #
    #     WHAT THIS CASE USED TO RECORD, and why the change is a decision rather
    #     than a fix to the case. The allowlist was spelled with bracket RANGES,
    #     and a bracket range is a COLLATING range: under `LC_ALL=C` `İ` fell
    #     outside `A-Za-z` and the name was refused before the workspace row was
    #     reached, while under `en_US.UTF-8` it collated INSIDE and the name was
    #     admitted. This case asserted both halves and endorsed neither, on the
    #     grounds that the allowlist was one rule character for character across
    #     every guard carrying it, and forking it here was a decision for whoever
    #     owned them all. That decision was taken: each of them now enumerates the
    #     allowed characters instead of ranging over them, a literal list has
    #     no endpoints for a collation order to reorder, and the accepted set is
    #     what the range form accepted under `LC_ALL=C` — so the split is closed
    #     rather than merely visible. check-cleanup-complete.sh's own Protection 1
    #     comment carries the measurement; it took that role from the
    #     record-copying script this repository has since retired.
    #
    #     THE ASSERTION IS THAT THE ANSWER NO LONGER DEPENDS ON THE ENVIRONMENT,
    #     which is why it is a loop over the same locales case 20 uses and not a
    #     single refusal under C. A case that ran under one locale would read as a
    #     statement about the guard while being a statement about the environment.
    #
    #     THE CANONICAL FILE'S FOURTH PUBLISHED WORKED VALUE, `--stanbul-6b8a`, IS
    #     NOT ORPHANED BY THIS. It is pinned by the self-check loop above, which
    #     drives the canonical block with every `# <name> -> <id>` comment the
    #     file publishes. What is gone is the guard-side comparison for that name,
    #     and it is gone because there is now no guard-side id to compare: a name
    #     the allowlist refuses never reaches the derivation, under any locale.
    new_fixture
    cat > "$REPO/record.sh" <<'SH'
#!/usr/bin/env bash
printf '%s %s\n' "$1" "$2" > "$(dirname "$0")/derived-id"
SH
    chmod +x "$REPO/record.sh"
    declare_isolation '`./record.sh "<id>" "<id_underscored>"`'

    for canon_loc in "${CANON_LOCALES[@]}"; do
      rm -f "$REPO/derived-id"
      run_guard_in_locale "$canon_loc" "$REPO" "İstanbul-test" "$STATE"
      assert_no_verdict "a non-ASCII change name under LC_ALL=$canon_loc"
      assert_absent "$REPO/derived-id" "a change name refused under LC_ALL=$canon_loc never reaches the derivation"
    done

    # 21b. The characters that MOVED when the ranges went, named one at a time.
    #      Case 21's `İstanbul-test` is one name, and one name cannot tell a
    #      pattern that is locale-independent from one that merely happens to
    #      agree about `İ`. Every probe below was ADMITTED by the range form
    #      under `en_US.UTF-8` and refused by it under `LC_ALL=C`, measured on
    #      this machine — a Latin letter with a diacritic, a ligature, a Roman
    #      numeral and a fullwidth capital, so the set is not four spellings of
    #      one hazard. `½x` and `a‐b` are deliberately absent: the range form
    #      refused those under every locale, so they would pass for a guard that
    #      changed nothing.
    for split_name in "écho" "ﬀoo" "ⅰx" "Ａbc"; do
      for canon_loc in "${CANON_LOCALES[@]}"; do
        rm -f "$REPO/derived-id"
        run_guard_in_locale "$canon_loc" "$REPO" "$split_name" "$STATE"
        assert_no_verdict "the change name '$split_name' under LC_ALL=$canon_loc"
        assert_absent "$REPO/derived-id" "'$split_name' under LC_ALL=$canon_loc never reaches the derivation"
      done
    done
  fi
fi

# 22. A command naming a token the contract does not name is reported and
#     DROPPED, never handed to the shell: `<container>` reaching sh is a
#     redirection, which can truncate a file the operator never named.
new_fixture
write_script survivors.sh "touch '$REPO/survivors-ran'"
declare_isolation '`./survivors.sh <container>`'
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a command carrying an unknown token does not block the terminal state"
assert_out "SKIPPED" "a command carrying an unknown token is reported as skipped"
assert_out "<container>" "the skip names the token it did not recognise"
assert_absent "$REPO/survivors-ran" "a command carrying an unknown token is never executed"

# 23. A `|` inside the command is part of the command, not a cell boundary. A
#     guard splitting the row on every pipe runs `./survivors.sh` alone and
#     reports BOTH lines as survivors, so the operator is sent after a resource
#     the project's own filter had already excluded.
new_fixture
write_script survivors.sh "printf 'alpha\nbeta\n'"
declare_isolation '`./survivors.sh | grep beta`'
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "a command containing a pipe still runs"
assert_reason "beta" "a command containing a pipe runs in full"
assert_no_reason "alpha" "a command containing a pipe is not truncated at the pipe"

# 23b. A FAILING STAGE INSIDE A PIPELINE IS AN ANSWER OF NOTHING, exactly as a
#      failing simple command is. A shell reports a pipeline's status from its
#      LAST stage only, so `./tool | cat` where the tool is missing or broken
#      exits 0 carrying an empty stdout — and the guard reads that as the one
#      result that VERIFIES the removal. That is the false COMPLETE this whole
#      guard exists to prevent, reached with nothing having failed visibly, and
#      case 23 above cannot catch it because both of its stages succeed.
#
#      The SILENT failing stage is the one that reaches the false COMPLETE, and
#      it is the likeliest shape there is: a missing interpreter, a build tool
#      not on PATH, a client that writes its diagnosis to stderr. The exit code
#      asserted is the FAILING STAGE'S, not the pipeline's last — a skip
#      reporting 0 would be a skip whose number contradicts it, and the operator
#      needs the number the project's own tooling chose.
new_fixture
write_script broken.sh "exit 9"
declare_isolation '`./broken.sh | cat`'
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a failing stage inside a pipeline does not block the terminal state"
assert_out "SKIPPED" "a failing stage inside a pipeline is reported as skipped"
assert_out "exited 9" "the skip carries the failing stage's exit code"
assert_not_out "is empty" "a failing pipeline is never reported as a verified empty report"

# 23b2. …and case 18b's rule holds for a pipeline too: a command that failed
#       said nothing about survivors, so whatever an earlier stage printed
#       before failing is not read as a survivor list. Without this, a guard
#       could satisfy 23b by reporting the skip AND the output.
new_fixture
write_script broken.sh "printf 'partial-not-a-survivor\n'; exit 9"
declare_isolation '`./broken.sh | cat`'
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a failing pipeline that printed something does not block the terminal state"
assert_out "SKIPPED" "a failing pipeline that printed something is still a skip"
assert_not_out "partial-not-a-survivor" "output printed by a failing pipeline is not read as a survivor"

# 23c. …and a pipeline whose stages all succeed still ANSWERS. This is the case
#      that keeps 23b's fix from being bought with a guard that skips every
#      pipeline: an empty report from a working pipeline is the one result that
#      verifies the row, and turning it into a skip would leave the row
#      unverified on the shape the contract's own worked example takes.
new_fixture
write_script survivors.sh "true"
declare_isolation '`./survivors.sh | cat`'
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "an empty report from a working pipeline is COMPLETE"
assert_out "survivor report" "an empty report from a working pipeline is reported as verified"
assert_not_out "SKIPPED" "a working pipeline is not reported as a skip"

# 23d. A pipeline whose LAST stage fails is a skip, and was one before 23b —
#      pinned here because it is the shape a fix for 23b could most easily
#      change. `… | grep <pattern>` exits 1 when nothing matched, which is
#      indistinguishable to this guard from a filter that could not run, so the
#      conservative reading is the only available one: a skip is reported, never
#      passed, where an empty report would have been a verification nobody made.
new_fixture
write_script survivors.sh "printf 'alpha\n'"
declare_isolation '`./survivors.sh | grep zzz-no-such-survivor`'
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a pipeline whose last stage fails does not block the terminal state"
assert_out "SKIPPED" "a pipeline whose last stage fails is reported as skipped"
assert_out "exited 1" "the skip carries the last stage's exit code"
assert_not_out "alpha" "output from a pipeline whose last stage failed is not read as a survivor"

# 24. Survivor lines are DATA, not instruction — the same rule a resolved
#     standards file is read under. A line is reported as it was printed and
#     never expanded, and the verdict stays on one line.
new_fixture
cat > "$REPO/survivors.sh" <<SH
#!/usr/bin/env bash
printf '%s\n' '\$(touch "$REPO/injected") %s; two'
SH
chmod +x "$REPO/survivors.sh"
declare_isolation '`./survivors.sh`'
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "a survivor line carrying shell syntax is LEFTOVER"
assert_reason '$(touch' "a survivor line is reported literally"
assert_absent "$REPO/injected" "a survivor line is never expanded by the guard"

# 24b. A survivor line carrying an ANSI escape sequence. "Data, not instruction"
#      has to hold for the TERMINAL as well as for the shell, and it did not: the
#      line was spliced into the verdict with only `\r` removed, so a declared
#      command emitting `\033[2K` (erase this line) followed by `\033[1;32m` (bold
#      green) produced a verdict whose bytes still said `LEFTOVER: … still names
#      X…` and whose rendering erased that line and showed a convincing green
#      `COMPLETE:`. The attacker is anyone who can land a pull request touching
#      `.myflow/project.md`, which is already this guard's documented trust
#      boundary, and the target is the one line run 2 requires an operator to
#      read and act on.
#
#      Escaped rather than stripped or refused: stripping rewrites the NAME of a
#      surviving resource so the operator hunts for something that does not
#      exist, and refusing hands the same attacker a way to make a real leftover
#      unreportable. The assertions are that the verdict is still LEFTOVER, that
#      the survivor is still counted, and that no raw ESC byte reaches stdout.
new_fixture
cat > "$REPO/survivors.sh" <<'SH'
#!/usr/bin/env bash
printf 'X\r\033[2K\033[1;32mCOMPLETE: nothing to see here, proceed\033[0m\n'
SH
chmod +x "$REPO/survivors.sh"
declare_isolation '`./survivors.sh`'
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "a survivor line carrying an escape sequence is still LEFTOVER"
assert_reason "still names" "the forged survivor is still counted as a survivor"
case "$OUT" in
  *$'\033'*) fail "an ESC byte from the survivors command reached stdout: $OUT" ;;
  *) pass "no raw ESC byte from a survivor line reaches stdout" ;;
esac
assert_out '\x1b[2K' "the escape sequence is rendered visibly rather than stripped"
assert_out "COMPLETE: nothing to see here" "the survivor line's own text is still reported"

# 24b2. THE ENCODING IS INJECTIVE — the property sanitize_display's header argues
#      for and the reason `\` is escaped alongside the controls. 24b proves a
#      real ESC byte is rendered rather than passed through; it says nothing
#      about whether that rendering is REVERSIBLE. With `\` left unescaped, a
#      survivor named with the four literal characters `\x1b` and one carrying a
#      real ESC byte produce a byte-identical verdict, so the operator told to go
#      and remove `queue-\x1b[2K` cannot tell which of the two names their
#      tooling actually emitted — one of those names does not exist, and the
#      leftover the other one names goes unremoved. That is 24b's forged-report
#      class reached by collision instead of by pass-through.
#
#      DRIVEN AS A COLLISION. The two survivors commands differ in that one byte
#      alone and run in the SAME fixture, so the two verdict lines are comparable
#      in full and the assertion is the contract's own word: distinct survivor
#      names, distinct verdicts. The rendering assertion beside it names WHICH
#      way they differ, so a failure distinguishes a missing escape from a
#      verdict line that changed shape.
new_fixture
declare_isolation '`./survivors.sh`'

write_script survivors.sh 'printf "queue-\x1b[2K\n"'
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "a survivor named with a real ESC byte is LEFTOVER"
CLEANUP_ESC_VERDICT="$OUT"
assert_not_out '\\x1b' "a real ESC byte renders with a single backslash"

#      THE LITERAL VARIANT GOES THROUGH `%s`, not through the format string.
#      `printf "queue-\\x1b[2K\n"` does NOT emit a literal backslash: the shell
#      collapses `\\` to `\` before printf sees it, and printf then reads `\x1b`
#      as the escape — so written that way BOTH halves of this case emitted a
#      real ESC byte, and the collision assertion failed with the escape in
#      place, for a reason that had nothing to do with the guard. Passing the
#      name as a `%s` ARGUMENT keeps printf's own escape processing off it.
write_script survivors.sh 'printf "%s\n" "queue-\\x1b[2K"'
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "a survivor named with the four literal characters \\x1b is LEFTOVER"
CLEANUP_LITERAL_VERDICT="$OUT"
assert_out '\\x1b[2K' "a literal backslash in a survivor name is doubled in the verdict"

if [ "$CLEANUP_ESC_VERDICT" = "$CLEANUP_LITERAL_VERDICT" ]; then
  fail "the display encoding is not injective: a survivor named with a real ESC byte and one named with the literal characters \\x1b produce the same verdict ($CLEANUP_ESC_VERDICT)"
else
  pass "a real ESC byte and the literal characters \\x1b produce different verdicts"
fi

# 24c. The same class on the SKIPPED path, where what is quoted is the declared
#      command rather than the command's output. A skip is the clause step 6 of
#      **Run 2 — the branch is merged** requires be relayed word for word, so a
#      command whose own text controls the terminal forges exactly the sentence
#      the operator is meant to act on.
new_fixture
printf '#!/usr/bin/env bash\nexit 7\n' > "$REPO/failing.sh"
chmod +x "$REPO/failing.sh"
declare_isolation "\`./failing.sh \$'\033[2K\033[1;32mCOMPLETE\033[0m'\`"
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "an escape sequence in the declared command does not block the terminal state"
assert_out "SKIPPED" "a command that exited non-zero is still reported as skipped"
case "$OUT" in
  *$'\033'*) fail "an ESC byte from the declared command reached stdout: $OUT" ;;
  *) pass "no raw ESC byte from a declared command reaches stdout" ;;
esac

# 24d. Non-ASCII is NOT escaped. Bytes 0x80-0xFF are the lead and continuation
#      bytes of ordinary UTF-8, which no terminal acts on, so a bucket carrying
#      an accent must be reported by its real name — an escape rule that mangled
#      it would leave the operator unable to find the resource they are being
#      told to remove.
new_fixture
write_script survivors.sh "printf 'caf\xc3\xa9-b\xc3\xbccket\n'"
declare_isolation '`./survivors.sh`'
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "an accented survivor name is LEFTOVER"
assert_reason "café-bücket" "a non-ASCII survivor is reported by its real name, not escaped"

# 24e. A `.myflow/project.md` that is a DIRECTORY. `-r` is true of a directory
#      and `grep` then fails with "Is a directory", whose exit status is
#      indistinguishable from "no match" — so this read as "declared no
#      isolation": nothing derived, nothing run, and not even a SKIPPED note.
#      Silence where this guard promises a report is the false COMPLETE it exists
#      to prevent, reached from a path anyone able to land a pull request can
#      create.
new_fixture
mkdir -p "$REPO/.myflow/project.md"
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a project.md that is a directory does not block the terminal state"
assert_out "SKIPPED" "a project.md that is a directory is reported as skipped, not as absent"

# 24f. A `.myflow/project.md` that is a symlink to nowhere. `-e` follows the
#      link, so the absent test answered true and the guard stayed silent — an
#      absence manufactured out of a path deliberately pointed at nothing.
new_fixture
mkdir -p "$REPO/.myflow"
ln -s "$REPO/.myflow/no-such-configuration-target" "$REPO/.myflow/project.md"
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a dangling project.md symlink does not block the terminal state"
assert_out "SKIPPED" "a dangling project.md symlink is reported as skipped, not as absent"

# 24g. A `.myflow/project.md` that is a symlink to a REAL file is still read, so
#      the two cases above exclude a file type rather than an indirection.
new_fixture
mkdir -p "$REPO/.myflow"
write_script survivors.sh "printf 'linked-survivor\n'"
declare_isolation '`./survivors.sh`'
mv "$REPO/.myflow/project.md" "$REPO/.myflow/real-project.md"
ln -s "$REPO/.myflow/real-project.md" "$REPO/.myflow/project.md"
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "a project.md that is a symlink to a real file is still read"
assert_reason "linked-survivor" "the linked configuration's survivors command ran"

# 25. A survivor and a git-derived leftover are reported together, on the one
#     line the contract allows — the workspace row joins the existing breakdown
#     rather than replacing or preempting it.
new_fixture
git -C "$REPO" branch openspec/demo
write_script survivors.sh "printf 'db-survivor-one\n'"
declare_isolation '`./survivors.sh`'
run_guard "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "a survivor and a branch are one verdict"
assert_reason "local branch" "the combined breakdown still names the branch"
assert_reason "db-survivor-one" "the combined breakdown names the survivor"

# 26. A configuration that exists but cannot be read is not "declares no
#     isolation". Inferring absence from a path that was never readable is the
#     false COMPLETE this whole guard exists to prevent, so the unreadable file
#     is reported as a skip.
new_fixture
declare_isolation '`./survivors.sh`'
chmod 000 "$REPO/.myflow/project.md"
if [ -r "$REPO/.myflow/project.md" ]; then
  skip "an unreadable project configuration is reported" "this user can read a mode-000 file"
else
  run_guard "$REPO" demo "$STATE"
  assert_verdict "COMPLETE:" "an unreadable project configuration does not block the terminal state"
  assert_out "SKIPPED" "an unreadable project configuration is reported as skipped, not as absent"
fi
chmod 644 "$REPO/.myflow/project.md"

# 27. The workspace row is answered per repository, from the repository it was
#     asked about. The guard runs once per repository, so a survivors command
#     declared in one must not be found from another — a neighbouring checkout's
#     leftovers are not this one's.
new_fixture
write_script survivors.sh "printf 'other-repo-survivor\n'"
declare_isolation '`./survivors.sh`'
new_fixture
run_guard "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "another repository's isolation is not this repository's"
assert_not_out "other-repo-survivor" "another repository's survivors are not reported here"
assert_not_out "workspace" "a repository declaring no isolation stays silent about the workspace"

# ---------------------------------------------------------------------------
# 28-29b. The bound on the survivors command.
#
# Run 2 is unattended. `</dev/null` covers only a command blocking on an
# interactive prompt; it does nothing for one that hangs for any other reason —
# an unreachable host with no connect timeout, a lock wait, a frozen container,
# an infinite loop — and a hang there strands an already-merged change short of
# FINISHED with nobody watching. The project-supplied `## stop` command is
# already given a bounded wait for exactly this reason; see **Worktree cleanup**
# (`skills/myflow-contracts/pipeline.md`).
#
# WHY A TIMEOUT IS A SKIP HERE AND A FAILED CHECK THERE. A `## stop` timeout has
# no fallback: an un-stopped stack is a reason not to remove a worktree, and
# blocking is the safe direction. A survivors timeout has one — the reported
# skip the contract already defines for "the command said nothing about
# survivors" — and blocking an already-merged change over a slow command is what
# **Creation and cleanup** (`skills/myflow-contracts/workspace-isolation.md`)
# forbids. Same bound, opposite outcome, and the difference is which direction is
# safe at that point in the run.
# ---------------------------------------------------------------------------

# 28. A survivors command that exceeds the bound. It is reported as a skip and
#     run 2 continues — and the skip says TIMED OUT rather than carrying an exit
#     code, because "your command is slow" and "your command is broken" need
#     different remedies. Whatever it printed before it was killed is not a
#     survivor list, for the same reason case 18b asserts of a failed one: a
#     command that never answered said nothing about survivors.
new_fixture
write_script survivors.sh "printf 'noise-before-the-hang\n'; sleep 300"
declare_isolation '`./survivors.sh`'
run_guard_bounded 2 "$REPO" demo "$STATE"
assert_verdict "COMPLETE:" "a survivors command that hangs does not block the terminal state"
assert_out "SKIPPED" "a hung survivors command is reported as skipped"
assert_out "./survivors.sh" "the timeout skip names the command"
assert_out "timed out" "the timeout skip says it timed out"
assert_not_out "exited" "a timeout is not reported as an ordinary non-zero exit"
assert_not_out "noise-before-the-hang" "output printed before the kill is not read as a survivor"

# 28b. The killed command leaves NOTHING running. `bash -c 'cmd'` execs into the
#      command, and a pipeline forks further, so killing the one pid the guard
#      holds leaves the real work alive — a `sleep` still running long after the
#      guard reported, on a machine nobody is watching. The declared command is
#      a pipeline precisely because that is the shape a single-pid kill misses,
#      and the sleep's duration is the fingerprint: it is unusual enough that
#      pgrep matching it cannot match anything else on the machine.
if ! command -v pgrep >/dev/null 2>&1; then
  skip "a timed-out survivors command leaves no process behind" "this machine has no pgrep"
else
  new_fixture
  write_script slow.sh "sleep 2718"
  declare_isolation '`./slow.sh | cat`'
  run_guard_bounded 2 "$REPO" demo "$STATE"
  assert_verdict "COMPLETE:" "a hung survivors PIPELINE does not block the terminal state"
  assert_out "timed out" "a hung survivors pipeline is reported as a timeout"
  if pgrep -f 'sleep 2718' >/dev/null 2>&1; then
    fail "a timed-out survivors command leaks its children: 'sleep 2718' is still running"
    pkill -f 'sleep 2718' >/dev/null 2>&1 || true
  else
    pass "a timed-out survivors command leaves no process behind"
  fi
fi

# 28d. THE SHAPE THE GROUP KILL CANNOT REACH, driven rather than argued. Case
#      28b proves the kill reaps a pipeline, and every process in that case is
#      inside the group the guard created — so it says nothing about the shape a
#      project actually declares. A `survivors` command that reaches its service
#      through `docker exec` signals a PROXY: the query runs in another PID
#      namespace, which no host-side signal can name, so the guard's kill ends
#      its own wait and leaves the query running inside a container other
#      concurrent workspaces share. That cannot be started here — this suite
#      builds throwaway git repositories and nothing else — so the STRUCTURAL
#      case is driven instead: a child that claims a process group of its own
#      before the kill arrives is outside the group for the same reason and by
#      the same mechanism, and `set -m` is how a POSIX shell claims one.
#
#      THE ASSERTIONS ARE THE THREE PROMISES THAT SURVIVE THE ESCAPE, because
#      those are what run 2 depends on and none of them is obvious: the guard
#      RETURNS (`wait` names the direct child alone, so a surviving descendant
#      cannot strand it), the row is reported as a SKIP rather than as a
#      verification, and what the escapee printed is not read as a survivor.
#      The in-group half of the same command is asserted reaped, which is what
#      keeps this from passing for a guard that stopped killing the group at all.
#
#      THE LAST ASSERTION IS TWO-SIDED ON PURPOSE. The escapee surviving is a
#      documented limitation — check-cleanup-complete.sh's run_survivors header
#      and skills/myflow-contracts/project-configuration.md both state it — so
#      finding it gone is not a quiet improvement, it is those two paragraphs
#      becoming false. This case therefore fails in BOTH directions, exactly as
#      the registry-coupling case 14 does, and names what to edit.
if ! command -v pgrep >/dev/null 2>&1; then
  skip "a process that leaves the guard's process group outlives the bound" "this machine has no pgrep"
else
  new_fixture
  # 4271 escapes, 4272 stays in the group. Both durations are unusual enough
  # that a pgrep matching them cannot match anything else on this machine.
  write_script escaper.sh "set -m
sleep 4271 &
set +m
printf 'noise-from-the-escaped-shape\n'
sleep 4272"
  declare_isolation '`./escaper.sh`'
  ESC_START="$SECONDS"
  run_guard_bounded 2 "$REPO" demo "$STATE"
  ESC_ELAPSED="$((SECONDS - ESC_START))"
  assert_verdict "COMPLETE:" "an escaping survivors command does not block the terminal state"
  assert_out "timed out" "an escaping survivors command is reported as a timeout skip"
  assert_not_out "noise-from-the-escaped-shape" "the escaped shape's output is not read as a survivor"
  [ "$ESC_ELAPSED" -lt 30 ] \
    && pass "a surviving descendant does not strand the guard (${ESC_ELAPSED}s)" \
    || fail "the guard took ${ESC_ELAPSED}s while a descendant outlived its bound"
  pgrep -f 'sleep 4272' >/dev/null 2>&1 \
    && fail "the in-group half of an escaping command was not reaped: 'sleep 4272' is still running" \
    || pass "the in-group half of an escaping command is still reaped"
  if pgrep -f 'sleep 4271' >/dev/null 2>&1; then
    pass "a process that leaves the guard's process group outlives the bound, as documented"
  else
    fail "a process that left the guard's process group did NOT outlive the bound — check-cleanup-complete.sh's run_survivors header and skills/myflow-contracts/project-configuration.md both document that it does, and are now wrong"
  fi
  # Scoped to these two durations, never a bare pattern: this suite must not
  # reach a process it did not start.
  pkill -f 'sleep 4271' >/dev/null 2>&1 || true
  pkill -f 'sleep 4272' >/dev/null 2>&1 || true
fi

# 28c. The bounded wait needs somewhere to capture the command's stdout, because
#      a command substitution cannot be bounded — it blocks until EOF. When that
#      scratch file cannot be created the guard reports a skip and still reaches
#      a verdict: under `set -e` an unhandled mktemp failure would abort the run
#      with no verdict line at all, which is a caller grepping stdout for
#      COMPLETE finding nothing and a change stopped for the wrong reason. The
#      guard reads TMPDIR for that file and for nothing else, so pointing it at a
#      path that does not exist is enough to force the branch.
new_fixture
write_script survivors.sh "touch '$REPO/survivors-ran'"
declare_isolation '`./survivors.sh`'
OLD_TMPDIR="${TMPDIR:-}"
TMPDIR="/nonexistent/no-such-temporary-directory"
export TMPDIR
run_guard "$REPO" demo "$STATE"
if [ -n "$OLD_TMPDIR" ]; then TMPDIR="$OLD_TMPDIR"; export TMPDIR; else unset TMPDIR; fi
assert_verdict "COMPLETE:" "an uncreatable scratch file still reaches a verdict"
assert_out "SKIPPED" "an uncreatable scratch file is reported as skipped"
assert_absent "$REPO/survivors-ran" "an uncreatable scratch file means the command is never run"
assert_not_out "timed out" "an uncreatable scratch file is not reported as a timeout"

# 29. A survivors command that returns JUST INSIDE the bound is answered
#     normally — its exit code and its output are read exactly as they always
#     were. Without this case a guard that treated the bound as the deadline for
#     STARTING rather than for finishing, or that killed on any elapsed poll,
#     would turn every slow-but-working report into a skip, and a skip is never
#     a verification.
#
#     The margin is 3 seconds rather than a hair, and deliberately so: the
#     guard's deadline is counted in whole seconds, so the bound fires anywhere
#     in [bound-1, bound]. A case pressed right up against it would fail on a
#     loaded machine, and a flaky case that everyone learns to re-run is worth
#     less than a slightly generous one that never lies.
new_fixture
write_script survivors.sh "sleep 2; printf 'db-survivor-slow\n'"
declare_isolation '`./survivors.sh`'
run_guard_bounded 5 "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "a survivors command returning inside the bound is answered normally"
assert_reason "db-survivor-slow" "a slow-but-finished report still names its survivor"
assert_not_out "timed out" "a survivors command inside the bound is not reported as a timeout"

# 29b. The bound does not become a WAIT. A command that returns at once must
#      return at once with the shipped 60-second bound in force — the seam
#      deliberately left unset here, so this is also the case that proves the
#      default is a bound rather than a sleep. A guard that waited out its bound
#      before reading the result would pass every case above and make run 2 a
#      minute slower per repository.
new_fixture
write_script survivors.sh "true"
declare_isolation '`./survivors.sh`'
FAST_START="$SECONDS"
run_guard "$REPO" demo "$STATE"
FAST_ELAPSED="$((SECONDS - FAST_START))"
assert_verdict "COMPLETE:" "a survivors command that returns at once is COMPLETE"
[ "$FAST_ELAPSED" -lt 10 ] \
  && pass "the default bound is a bound, not a wait (${FAST_ELAPSED}s)" \
  || fail "the guard took ${FAST_ELAPSED}s for a survivors command that returns immediately"

# 29c. A junk override falls back to the shipped bound rather than to nothing —
#      and specifically does not become a bound that has ALREADY EXPIRED. The
#      value here is the shape that does it: a seconds count too large for the
#      shell's integer wraps negative when the deadline is computed, a deadline
#      in the past fires the moment the command starts, and every survivors
#      command in the pipeline would then be reported as a timeout it never had.
#      A two-second command is what tells that apart from a working fallback; an
#      instant one is answered either way.
new_fixture
write_script survivors.sh "sleep 2; printf 'db-survivor-junk-bound\n'"
declare_isolation '`./survivors.sh`'
run_guard_bounded 99999999999999999999 "$REPO" demo "$STATE"
assert_verdict "LEFTOVER:" "an out-of-range bound falls back to the shipped one"
assert_reason "db-survivor-junk-bound" "an out-of-range bound still lets the report be read"
assert_not_out "timed out" "an out-of-range bound does not fire the timeout immediately"

# 29e. Backwards compatibility, proved rather than argued: a project that
#      declares no isolation is not touched by any of this. Case 15 asserts the
#      verdict; this asserts that nothing was WAITED on either, which is the
#      failure a bound placed outside the isolation branch would introduce — it
#      would cost every repository in the pipeline its bound on every run 2,
#      including this one, which declares no isolation at all.
new_fixture
NOISO_START="$SECONDS"
run_guard "$REPO" demo "$STATE"
NOISO_ELAPSED="$((SECONDS - NOISO_START))"
assert_verdict "COMPLETE:" "a project declaring no isolation is unchanged by the bound"
assert_not_out "timed out" "a project declaring no isolation says nothing about a timeout"
assert_not_out "SKIPPED" "a project declaring no isolation reports no skip"
[ "$NOISO_ELAPSED" -lt 10 ] \
  && pass "a project declaring no isolation waits on nothing (${NOISO_ELAPSED}s)" \
  || fail "a project declaring no isolation took ${NOISO_ELAPSED}s"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-cleanup-complete: all cases pass\n'
