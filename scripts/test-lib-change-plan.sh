#!/usr/bin/env bash
# test-lib-change-plan.sh — assertion harness for scripts/lib/change-plan.sh,
# the shared resolver KAN-363 task 7 adds. Sources the library directly
# rather than through either caller (check-unfinished-work.sh,
# check-task-commit-fields.sh — tasks 8 and 9), per this repository's own
# convention (see test-lib-coverage.sh's header): the thing under test is
# the library's own contract, not any caller's use of it.
#
# Every fixture here is a real directory tree built with mktemp -d — a real
# tasks.md, a real link.md, a real peers file — and change_plan_path is run
# against it for real, per this change's own "reproduce, don't read"
# instruction: a hand-built string standing in for a tree would not exercise
# the filesystem boundary this function actually crosses (spec_root_leaf's
# own probes, the peers-file read, the `cd` that resolves a relative peer
# path).
#
# The six cases design.md's task 7 names: a plain change; a satellite with
# the canonical worktree passed; a satellite resolving through peers; a
# satellite whose peer is absent; a link.md with no `## Part of` (the
# canonical side's own link.md, carrying `## Parts` instead); and one
# rejected name per containment rule (change name, peer name, canonical
# change id — three sub-cases, since three distinct names are concatenated
# into a path per the task's own header comment).
#
# Bash 3.2 is the floor, as test-check-finish-preflight.sh's header records:
# indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB="$SCRIPT_DIR/lib/change-plan.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

if [ ! -r "$LIB" ]; then
  echo "test-lib-change-plan: cannot read $LIB" >&2
  exit 2
fi
# shellcheck source=lib/change-plan.sh
source "$LIB"

# assert_eq <label> <expected> <actual>
assert_eq() {
  if [ "$2" = "$3" ]; then
    pass "$1"
  else
    fail "$1 — expected [$2], got [$3]"
  fi
}

# assert_zero_rc <label> <rc>
assert_zero_rc() {
  if [ "$2" -eq 0 ]; then
    pass "$1"
  else
    fail "$1 — expected exit 0, got $2"
  fi
}

# assert_nonzero_rc <label> <rc>
assert_nonzero_rc() {
  if [ "$2" -ne 0 ]; then
    pass "$1"
  else
    fail "$1 — expected a non-zero exit, got 0"
  fi
}

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# make_change_dir <worktree> <name> — a bare spectre/changes/<name>/ dir.
make_tree() {
  mkdir -p "$1/spectre/changes"
}

# ---------------------------------------------------------------------------
# Case 1: a plain change — spectre/changes/<name>/tasks.md exists, no link.md
# involved at all.
# ---------------------------------------------------------------------------
PLAIN="$WORK/case1-worktree"
make_tree "$PLAIN"
mkdir -p "$PLAIN/spectre/changes/plain-change"
printf '# plain\n\n- [ ] 1. do a thing\n' > "$PLAIN/spectre/changes/plain-change/tasks.md"

set +e
OUT="$(change_plan_path "$PLAIN" "plain-change")"
RC=$?
set -e
assert_zero_rc "case 1: a plain change resolves at exit 0" "$RC"
assert_eq "case 1: it prints the plan's own tasks.md path" \
  "$PLAIN/spectre/changes/plain-change/tasks.md" "$OUT"

# change_plan_dir (KAN-363 task 8): the same resolution, printing the
# DIRECTORY the tasks.md lives in rather than the tasks.md path itself —
# check-unfinished-work.sh's fix-sub-change sweep needs this to know which
# changes/ tree and which id to sweep for a satellite.
set +e
OUT="$(change_plan_dir "$PLAIN" "plain-change")"
RC=$?
set -e
assert_zero_rc "case 1 dir: a plain change's directory resolves at exit 0" "$RC"
assert_eq "case 1 dir: it prints the change's own directory" \
  "$PLAIN/spectre/changes/plain-change" "$OUT"

# ---------------------------------------------------------------------------
# Case 2: a satellite with the canonical worktree passed. The satellite
# worktree carries only link.md; the canonical worktree carries the real
# tasks.md under a differently-named change id.
# ---------------------------------------------------------------------------
SAT2="$WORK/case2-satellite"
CANON2="$WORK/case2-canonical"
make_tree "$SAT2"
make_tree "$CANON2"
mkdir -p "$SAT2/spectre/changes/sat-change"
cat > "$SAT2/spectre/changes/sat-change/link.md" <<'EOF'
## Part of

`peerx:canon-change`
EOF
mkdir -p "$CANON2/spectre/changes/canon-change"
printf '# canonical\n\n- [x] 1. done\n' > "$CANON2/spectre/changes/canon-change/tasks.md"

set +e
OUT="$(change_plan_path "$SAT2" "sat-change" "$CANON2")"
RC=$?
set -e
assert_zero_rc "case 2: a satellite with the canonical worktree passed resolves at exit 0" "$RC"
assert_eq "case 2: it prints the canonical worktree's own tasks.md path" \
  "$CANON2/spectre/changes/canon-change/tasks.md" "$OUT"

# change_plan_dir on the same satellite fixture: the CANONICAL change's own
# directory, under CANON2, not SAT2 — the whole reason task 8's sweep needed
# this function rather than a plain `dirname` on change_plan_path's answer.
set +e
OUT="$(change_plan_dir "$SAT2" "sat-change" "$CANON2")"
RC=$?
set -e
assert_zero_rc "case 2 dir: a satellite's directory resolves at exit 0" "$RC"
assert_eq "case 2 dir: it prints the canonical worktree's own change directory" \
  "$CANON2/spectre/changes/canon-change" "$OUT"

# ---------------------------------------------------------------------------
# Case 3: a satellite resolving through peers — no canonical worktree
# argument, so the peer name is read from spectre/peers and the plan is
# found on the other side of a real `cd`.
# ---------------------------------------------------------------------------
PARENT3="$WORK/case3-parent"
mkdir -p "$PARENT3"
SAT3="$PARENT3/sat-tree"
PEER3="$PARENT3/peer-tree"
make_tree "$SAT3"
make_tree "$PEER3"
mkdir -p "$SAT3/spectre/changes/sat-change"
cat > "$SAT3/spectre/changes/sat-change/link.md" <<'EOF'
## Part of

`peery:canon-change`
EOF
printf 'peery ../peer-tree\n' > "$SAT3/spectre/peers"
mkdir -p "$PEER3/spectre/changes/canon-change"
printf '# canonical\n\n- [x] 1. done\n' > "$PEER3/spectre/changes/canon-change/tasks.md"

set +e
OUT="$(change_plan_path "$SAT3" "sat-change")"
RC=$?
set -e
assert_zero_rc "case 3: a satellite resolving through peers resolves at exit 0" "$RC"
assert_eq "case 3: it prints the peer tree's own tasks.md path" \
  "$PEER3/spectre/changes/canon-change/tasks.md" "$OUT"

# ---------------------------------------------------------------------------
# Case 2b: a satellite with the canonical worktree passed, but the plan is
# not there — must NOT fall back to peers even though a peers file exists
# and would otherwise resolve. design.md's guards-take-the-canonical-
# worktree-path: a wrong answer from flow's own resolved worktree set is a
# fact worth failing loudly on, not a cue to keep searching.
# ---------------------------------------------------------------------------
PARENT2B="$WORK/case2b-parent"
mkdir -p "$PARENT2B"
SAT2B="$PARENT2B/sat-tree"
CANON2B="$WORK/case2b-canonical"
PEER2B="$PARENT2B/peer-tree"
make_tree "$SAT2B"
make_tree "$CANON2B"
make_tree "$PEER2B"
mkdir -p "$SAT2B/spectre/changes/sat-change"
cat > "$SAT2B/spectre/changes/sat-change/link.md" <<'EOF'
## Part of

`peerw:canon-change`
EOF
printf 'peerw ../peer-tree\n' > "$SAT2B/spectre/peers"
mkdir -p "$PEER2B/spectre/changes/canon-change"
printf '# canonical via peers\n\n- [x] 1. done\n' > "$PEER2B/spectre/changes/canon-change/tasks.md"
# CANON2B carries no canon-change/tasks.md at all — the canonical-worktree
# argument's own resolution comes up empty.

set +e
OUT="$(change_plan_path "$SAT2B" "sat-change" "$CANON2B")"
RC=$?
set -e
assert_nonzero_rc "case 2b: a canonical worktree passed with no plan there does not fall back to peers" "$RC"
assert_eq "case 2b: it prints nothing to stdout" "" "$OUT"

# ---------------------------------------------------------------------------
# Case 4: a satellite whose peer is absent — declared in peers, but nothing
# is checked out at the resolved path. peer-absence-is-not-a-finding: the
# caller decides whether that is a refusal or a verdict, this function just
# cannot reach the plan.
# ---------------------------------------------------------------------------
PARENT4="$WORK/case4-parent"
mkdir -p "$PARENT4"
SAT4="$PARENT4/sat-tree"
make_tree "$SAT4"
mkdir -p "$SAT4/spectre/changes/sat-change"
cat > "$SAT4/spectre/changes/sat-change/link.md" <<'EOF'
## Part of

`ghost:canon-change`
EOF
printf 'ghost ../not-checked-out\n' > "$SAT4/spectre/peers"

set +e
OUT="$(change_plan_path "$SAT4" "sat-change")"
RC=$?
set -e
assert_nonzero_rc "case 4: a satellite whose peer is absent cannot resolve" "$RC"
assert_eq "case 4: it prints nothing to stdout" "" "$OUT"

# ---------------------------------------------------------------------------
# Case 5: link.md with no `## Part of` — the canonical side's own link.md,
# carrying `## Parts` instead. Neither branch of the resolution order
# applies, so this returns 1 exactly like a change with no link.md at all.
# ---------------------------------------------------------------------------
CANON5="$WORK/case5-canonical"
make_tree "$CANON5"
mkdir -p "$CANON5/spectre/changes/canon-change"
printf '# canonical\n\n- [x] 1. done\n' > "$CANON5/spectre/changes/canon-change/tasks.md"
mkdir -p "$CANON5/spectre/changes/canon-only-satellite-dir"
cat > "$CANON5/spectre/changes/canon-only-satellite-dir/link.md" <<'EOF'
## Parts

`peerz:some-part`

## Merge order

1. `.`
2. `peerz`
EOF

set +e
OUT="$(change_plan_path "$CANON5" "canon-only-satellite-dir")"
RC=$?
set -e
assert_nonzero_rc "case 5: a link.md with no ## Part of cannot resolve" "$RC"
assert_eq "case 5: it prints nothing to stdout" "" "$OUT"

# ---------------------------------------------------------------------------
# Case 5b: an empty `## Part of` section immediately followed by a `##
# Parts` heading that does carry a code span. _change_plan_link_part_of
# must stop scanning at the next `## ` heading rather than reading past it
# into an unrelated section — planted so a missing "stop at the next
# heading" break would return the LATER section's code span
# ("peerz:some-part") as if it were `## Part of`'s own content, and that
# bogus ref resolves to a real plan below: without the break, this
# malformed link.md would wrongly resolve at exit 0.
# ---------------------------------------------------------------------------
SAT5B="$WORK/case5b-worktree"
PEER5B="$WORK/case5b-peer"
make_tree "$SAT5B"
make_tree "$PEER5B"
mkdir -p "$SAT5B/spectre/changes/sat-change"
cat > "$SAT5B/spectre/changes/sat-change/link.md" <<'EOF'
## Part of

## Parts

`peerz:some-part`
EOF
printf 'peerz ../case5b-peer\n' > "$SAT5B/spectre/peers"
mkdir -p "$PEER5B/spectre/changes/some-part"
printf '# escaped plan that must never be reached\n' > "$PEER5B/spectre/changes/some-part/tasks.md"

set +e
OUT="$(change_plan_path "$SAT5B" "sat-change" 2>/dev/null)"
RC=$?
set -e
assert_nonzero_rc "case 5b: an empty ## Part of does not read past the next heading" "$RC"
assert_eq "case 5b: it prints nothing to stdout" "" "$OUT"

# ---------------------------------------------------------------------------
# Case 6a: a rejected change name (arg 2) — containment rule 1 of 3. Built
# so the escape actually lands on a real file if the containment check is
# ever removed: BAD6/spectre/changes/../../secret/tasks.md resolves (the
# kernel, not bash, walks `..`) to a real tasks.md one level above
# BAD6/spectre/ — a genuine reproduction of the "planted/clear" hazard
# check-unfinished-work.sh's own header records, not an escape that merely
# happens to land nowhere.
# ---------------------------------------------------------------------------
BAD6="$WORK/case6-worktree"
make_tree "$BAD6"
mkdir -p "$BAD6/secret"
printf '# secret plan that must never be reached\n' > "$BAD6/secret/tasks.md"

set +e
OUT="$(change_plan_path "$BAD6" "../../secret" 2>/dev/null)"
RC=$?
set -e
assert_nonzero_rc "case 6a: a change name outside the allowlist is rejected" "$RC"
assert_eq "case 6a: it prints nothing to stdout" "" "$OUT"

# ---------------------------------------------------------------------------
# Case 6b: a rejected peer name, embedded in link.md's ## Part of —
# containment rule 2 of 3. Built, like case 6a, so the escape actually
# lands on a real file if the containment check is ever removed: the
# peers file declares a literal "../escape" entry pointing at a real
# sibling tree, "$WORK/elsewhere", that carries a genuine canon-change
# plan. Without the check, peer_name reaches _change_plan_peer_root
# unfiltered, matches that entry by plain string equality, and resolves
# a real tasks.md — the case the earlier fixture's absent "elsewhere"
# directory let pass either way.
# ---------------------------------------------------------------------------
SAT6B="$WORK/case6b-worktree"
make_tree "$SAT6B"
mkdir -p "$SAT6B/spectre/changes/sat-change"
cat > "$SAT6B/spectre/changes/sat-change/link.md" <<'EOF'
## Part of

`../escape:canon-change`
EOF
printf '../escape ../elsewhere\n' > "$SAT6B/spectre/peers"
mkdir -p "$WORK/elsewhere/spectre/changes/canon-change"
printf '# escaped plan that must never be reached\n' > "$WORK/elsewhere/spectre/changes/canon-change/tasks.md"

set +e
OUT="$(change_plan_path "$SAT6B" "sat-change" 2>/dev/null)"
RC=$?
set -e
assert_nonzero_rc "case 6b: a peer name outside the allowlist is rejected" "$RC"
assert_eq "case 6b: it prints nothing to stdout" "" "$OUT"

# ---------------------------------------------------------------------------
# Case 6c: a rejected canonical change id, embedded in link.md's
# ## Part of — containment rule 3 of 3. Built, like case 6a, so the escape
# actually lands on a real file if the containment check is ever removed:
# CANON6C/spectre/changes/../escape resolves (the kernel, not bash, walks
# `..`) to a real tasks.md at CANON6C/spectre/escape/tasks.md, one level
# above changes/.
# ---------------------------------------------------------------------------
SAT6C="$WORK/case6c-worktree"
CANON6C="$WORK/case6c-canonical"
make_tree "$SAT6C"
make_tree "$CANON6C"
mkdir -p "$SAT6C/spectre/changes/sat-change"
cat > "$SAT6C/spectre/changes/sat-change/link.md" <<'EOF'
## Part of

`peerx:../escape`
EOF
mkdir -p "$CANON6C/spectre/escape"
printf '# escaped plan that must never be reached\n' > "$CANON6C/spectre/escape/tasks.md"

set +e
OUT="$(change_plan_path "$SAT6C" "sat-change" "$CANON6C" 2>/dev/null)"
RC=$?
set -e
assert_nonzero_rc "case 6c: a canonical change id outside the allowlist is rejected" "$RC"
assert_eq "case 6c: it prints nothing to stdout" "" "$OUT"

# ---------------------------------------------------------------------------
if [ "$FAILURES" -eq 0 ]; then
  printf '\n✓ PASS\n'
  exit 0
fi
printf '\n✗ FAIL — %s failure(s)\n' "$FAILURES" >&2
exit 1
