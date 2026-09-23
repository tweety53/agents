#!/usr/bin/env bash
# Assertion harness for check-tree-markers.sh.
#
# Builds throwaway git repositories under TMPDIR and invokes the REAL
# scripts/check-tree-markers.sh against them — never a copy of its logic,
# and never a hand-written expected-output string asserted without having
# run the guard to produce it.
#
# The failure shapes the guard exists for, pinned case by case: a rewritten
# artifact that lost its marker (case 3), a deleted artifact (case 4), and
# never-committed content destroyed by a `git checkout` (case 5, verbatim
# the KAN-579 incident — the artifacts were never in git, so only the
# recorded markers can answer whether they survived). Case 6 pins the
# scoping decision: churn outside the marker list — a moving HEAD, new
# files, edits to unlisted files — is clean, because a working implementer
# legitimately moves everything it was dispatched to move.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-tree-markers.sh"
FAILED=0

WORKTREES=()
SNAPS=()
cleanup() {
  local p
  for p in "${WORKTREES[@]:-}"; do
    [ -n "$p" ] && rm -rf "$p"
  done
  for p in "${SNAPS[@]:-}"; do
    [ -n "$p" ] && rm -rf "$p"
  done
  # The EXIT trap's last status replaces the script's own on this bash —
  # end on success so a green suite exits 0, never the loop's final 1.
  true
}
trap cleanup EXIT

# make_repo -> a git repository with one committed artifact holding two
# known strings, and a notes file whose uncommitted edit carries a string
# of its own. Prints the repo path.
make_repo() {
  local wt
  wt="$(mktemp -d "${TMPDIR:-/tmp}/check-tree-markers-test.XXXXXX")" || {
    printf 'make_repo: mktemp failed -- aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  WORKTREES+=("$wt")
  git -C "$wt" init -q
  git -C "$wt" config user.email test@example.com
  git -C "$wt" config user.name test
  mkdir -p "$wt/spectre/changes/demo"
  printf '## Decisions\n\n**ID:** the-decision\n' > "$wt/spectre/changes/demo/design.md"
  printf 'working notes: committed baseline\n' > "$wt/notes.md"
  git -C "$wt" add -A
  git -C "$wt" commit -q -m original
  printf 'working notes: uncommitted\n' >> "$wt/notes.md"
  printf '%s' "$wt"
}

# make_markers -> a markers file for the artifacts above: two spec lines
# (path<TAB>ERE) on the committed artifact, the notes marker that exists
# only in the uncommitted edit, and one line whose marker matches nothing.
# Prints the path.
make_markers() {
  local d f
  d="$(mktemp -d "${TMPDIR:-/tmp}/check-tree-markers-markers.XXXXXX")" || {
    printf 'make_markers: mktemp failed -- aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  SNAPS+=("$d")
  f="$d/markers"
  printf 'spectre/changes/demo/design.md\tthe-decision\n' > "$f"
  printf 'spectre/changes/demo/design.md\t## Decisions\n' >> "$f"
  printf 'notes.md\tuncommitted\n' >> "$f"
  printf 'notes.md\tno-such-marker\n' >> "$f"
  printf '%s' "$f"
}

# new_snap -> a scratch path for a snapshot file. Prints the path.
new_snap() {
  local d
  d="$(mktemp -d "${TMPDIR:-/tmp}/check-tree-markers-snap.XXXXXX")" || {
    printf 'new_snap: mktemp failed -- aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  SNAPS+=("$d")
  printf '%s' "$d/snapshot"
}

run_guard() {
  local mode="$1" wt="$2" markers="$3" snap="$4"
  "$GUARD" "$mode" "$wt" "$markers" "$snap"
}

expect_exit() {
  local label="$1" want="$2"; shift 2
  local out got
  set +e
  out="$("$@" 2>&1)"; got=$?
  set -e
  if [[ "$got" != "$want" ]]; then
    printf 'FAIL %s: expected exit %s, got %s\n%s\n' "$label" "$want" "$got" "$out"
    FAILED=1
  else
    printf 'ok: %s\n' "$label"
  fi
  LAST_OUT="$out"
}

expect_names() {
  local label="$1" needle="$2"
  case "${LAST_OUT:-}" in
    *"$needle"*) printf 'ok: %s\n' "$label" ;;
    *) printf 'FAIL %s: expected %s named in output, got:\n%s\n' "$label" "$needle" "${LAST_OUT:-<empty>}"
       FAILED=1 ;;
  esac
}

expect_file_line() {
  local label="$1" file="$2" needle="$3"
  if grep -qF -- "$needle" "$file"; then
    printf 'ok: %s\n' "$label"
  else
    printf 'FAIL %s: expected %s in %s, got:\n' "$label" "$needle" "$file"
    sed 's/^/  /' "$file"
    FAILED=1
  fi
}

# ===========================================================================
# Case 1: snapshot records counts — exit 0, and the snapshot file carries
# one path/marker/status line per spec line, `missing` never appearing
# while the files exist and a 0 recorded for the marker that matches
# nothing.
# ===========================================================================
wt="$(make_repo)"; markers="$(make_markers)"; snap="$(new_snap)"
expect_exit 'case 1a: snapshot exits 0' 0 run_guard snapshot "$wt" "$markers" "$snap"
expect_file_line 'case 1b: records the decision-id marker count' "$snap" 'spectre/changes/demo/design.md	the-decision	1'
expect_file_line 'case 1c: records the heading marker count' "$snap" 'spectre/changes/demo/design.md	## Decisions	1'
expect_file_line 'case 1d: records a zero count for an absent marker' "$snap" 'notes.md	no-such-marker	0'

# ===========================================================================
# Case 2: verify on an untouched tree exits 0.
# ===========================================================================
expect_exit 'case 2: verify clean on an untouched tree exits 0' 0 run_guard verify "$wt" "$markers" "$snap"

# ===========================================================================
# Case 3: an artifact rewritten without its known string — verify exits 1
# and names the artifact.
# ===========================================================================
wt="$(make_repo)"; markers="$(make_markers)"; snap="$(new_snap)"
expect_exit 'case 3a: snapshot exits 0' 0 run_guard snapshot "$wt" "$markers" "$snap"
printf '## Decisions\n\n**ID:** another-decision\n' > "$wt/spectre/changes/demo/design.md"
expect_exit 'case 3b: verify after a marker-losing rewrite exits 1' 1 run_guard verify "$wt" "$markers" "$snap"
expect_names 'case 3c: names the rewritten artifact' 'design.md'
expect_names 'case 3d: says the clean-state claim is not evidence' 'not evidence'

# ===========================================================================
# Case 4: an artifact deleted after the snapshot — verify exits 1, even
# though every surviving marker line would still agree.
# ===========================================================================
wt="$(make_repo)"; markers="$(make_markers)"; snap="$(new_snap)"
expect_exit 'case 4a: snapshot exits 0' 0 run_guard snapshot "$wt" "$markers" "$snap"
rm "$wt/spectre/changes/demo/design.md"
expect_exit 'case 4b: verify after a deletion exits 1' 1 run_guard verify "$wt" "$markers" "$snap"
expect_names 'case 4c: names the deleted artifact' 'design.md'

# ===========================================================================
# Case 5: THE KAN-579 SHAPE — never-committed content destroyed by a
# `git checkout` after the snapshot. The marker exists only in the
# uncommitted edit, so git's own view of the tree cannot answer whether
# that content survived; the recorded markers can.
# ===========================================================================
wt="$(make_repo)"; markers="$(make_markers)"; snap="$(new_snap)"
expect_exit 'case 5a: snapshot exits 0' 0 run_guard snapshot "$wt" "$markers" "$snap"
git -C "$wt" checkout -- .
expect_exit 'case 5b: verify after the checkout destroys the uncommitted edit exits 1' 1 run_guard verify "$wt" "$markers" "$snap"
expect_names 'case 5c: names the destroyed notes file' 'notes.md'

# ===========================================================================
# Case 6: churn outside the marker list is clean — a moving HEAD, new
# files, edits to unlisted files. A working implementer legitimately moves
# everything it was dispatched to move; the markers pin only what must
# survive.
# ===========================================================================
wt="$(make_repo)"; markers="$(make_markers)"; snap="$(new_snap)"
expect_exit 'case 6a: snapshot exits 0' 0 run_guard snapshot "$wt" "$markers" "$snap"
printf 'source: edited\n' > "$wt/src.txt"
printf 'planted\n' > "$wt/planted.md"
git -C "$wt" add -A
git -C "$wt" commit -q -m implementer-work
expect_exit 'case 6: verify ignores churn outside the marker list exits 0' 0 run_guard verify "$wt" "$markers" "$snap"

# ===========================================================================
# Case 7: exit 2 refusals — bad invocation, unreadable markers file, and
# every spec line that could point grep outside the worktree or at
# nothing: an absolute path, a `..` component, a line without a tab, an
# empty marker, and a snapshot path whose directory does not exist.
# ===========================================================================
wt="$(make_repo)"; snap="$(new_snap)"
expect_exit 'case 7a: no arguments exits 2' 2 "$GUARD"
expect_exit 'case 7b: a bad mode exits 2' 2 run_guard check "$wt" "$snap" "$snap"
expect_exit 'case 7c: a non-directory worktree exits 2' 2 run_guard snapshot "$wt/no-such-dir" "$snap" "$snap"
expect_exit 'case 7d: an unreadable markers file exits 2' 2 run_guard snapshot "$wt" "$snap/no-such-markers" "$snap"
bad="$(mktemp -d "${TMPDIR:-/tmp}/check-tree-markers-bad.XXXXXX")"
SNAPS+=("$bad")
printf '/etc/passwd\troot\n' > "$bad/absolute"
expect_exit 'case 7e: an absolute path in the markers file exits 2' 2 run_guard snapshot "$wt" "$bad/absolute" "$snap"
printf '../escape\troot\n' > "$bad/dotdot"
expect_exit 'case 7f: a .. path in the markers file exits 2' 2 run_guard snapshot "$wt" "$bad/dotdot" "$snap"
printf 'no-tab-line\n' > "$bad/notab"
expect_exit 'case 7g: a line without a tab exits 2' 2 run_guard snapshot "$wt" "$bad/notab" "$snap"
printf 'notes.md	\n' > "$bad/emptymarker"
expect_exit 'case 7h: an empty marker exits 2' 2 run_guard snapshot "$wt" "$bad/emptymarker" "$snap"
expect_exit 'case 7i: snapshot to an unwritable path exits 2' 2 run_guard snapshot "$wt" "$markers" "$snap/no-dir/snapshot"

# ===========================================================================
# Case 8: verify with no readable snapshot file — exit 2, cannot answer.
# A snapshot runs before the dispatch, never after.
# ===========================================================================
wt="$(make_repo)"; markers="$(make_markers)"
snap="$(new_snap)/absent"
expect_exit 'case 8: verify without a snapshot file exits 2' 2 run_guard verify "$wt" "$markers" "$snap"

# ===========================================================================
# Case 9: F1 (panel round 1) — an existing DIRECTORY as the snapshot path.
# mv would file the snapshot inside it, the promised path would hold
# nothing, and exit 0 would claim a snapshot that is not there. Refused at
# exit 2 before anything is written.
# ===========================================================================
wt="$(make_repo)"; markers="$(make_markers)"
snapdir="$(mktemp -d "${TMPDIR:-/tmp}/check-tree-markers-snapdir.XXXXXX")"
SNAPS+=("$snapdir")
expect_exit 'case 9: snapshot to an existing directory exits 2' 2 run_guard snapshot "$wt" "$markers" "$snapdir"

# ===========================================================================
# Case 10: F2 (panel round 1) — a markers file with zero valid lines.
# Two empty line-sets compare equal, so a destroyed tree would verify
# clean over an empty pin set. Refused at exit 2.
# ===========================================================================
wt="$(make_repo)"
empty="$(mktemp -d "${TMPDIR:-/tmp}/check-tree-markers-empty.XXXXXX")"
SNAPS+=("$empty")
: > "$empty/markers"
snap="$(new_snap)"
expect_exit 'case 10a: snapshot over an empty markers file exits 2' 2 run_guard snapshot "$wt" "$empty/markers" "$snap"
expect_exit 'case 10b: verify over an empty markers file exits 2' 2 run_guard verify "$wt" "$empty/markers" "$snap"

if [ "$FAILED" -ne 0 ]; then
  exit 1
fi
exit 0
