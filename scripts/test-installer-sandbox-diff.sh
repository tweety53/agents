#!/usr/bin/env bash
# Assertion harness for installer-sandbox-diff.sh (KAN-589). Builds throwaway
# two-tree fixture pairs under a sandboxed TMPDIR, each tree carrying a
# minimal stand-in installer whose output is deterministic and derived from
# that tree's own data files, and asserts the diff tool's exit codes, its
# normalized-manifest output and its setup.sh-at-root guard against them.
# Never touches the real repository tree or the real HOME. Same shape as
# test-break-and-prove.sh: an indexed TREES array (bash 3.2 has no
# associative arrays), removed by an EXIT trap.
#
# Case 1 does the load-bearing work twice over: two fixtures built
# independently install symlinks pointing at two different absolute tree
# paths, so a clean diff proves both the fresh-HOME-per-side contract (a
# reused HOME would leave one sandbox empty and every line would differ) and
# the symlink-target normalization (raw targets would differ on every link).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIFF="$SCRIPT_DIR/installer-sandbox-diff.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

assert_rc() { # assert_rc <expected> <what>
	if [ "$DIFF_RC" -eq "$1" ]; then
		pass "$2"
	else
		fail "$2 — expected exit $1, got $DIFF_RC"
	fi
}

assert_contains() { # assert_contains <needle> <what>
	if grep -qF -- "$1" "$OUT_FILE"; then
		pass "$2"
	else
		fail "$2 — output has no line matching: $1"
	fi
}

TREES=()
cleanup() {
	[ "${#TREES[@]}" -eq 0 ] && return 0
	for p in "${TREES[@]}"; do
		rm -rf "$p" "${p}.out" 2>/dev/null || true
	done
}
trap cleanup EXIT

OUT_FILE=""

# new_tree [variant] — a fixture tree with a stand-in installer at its root.
# Variants: base (the plain tree), content (installed block text differs),
# mode (installed file's exec bit differs), nosetup (no installer at all).
new_tree() {
	local variant="${1:-base}"
	TREE="$(mktemp -d "${TMPDIR:-/tmp}/installer-sandbox-diff-test.XXXXXX")"
	TREES+=("$TREE")
	OUT_FILE="${TREE}.out"
	cat >"$TREE/setup.sh" <<'EOF'
#!/usr/bin/env bash
# Stand-in for setup.sh: deterministic output derived from THIS tree's
# data files, installed under $HOME the way `setup.sh global` installs.
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
mkdir -p "$HOME/.cfg/bin"
ln -s "$DIR/data/rule-a.mdc" "$HOME/.cfg/rule-a.mdc"
{ printf 'begin\n'; cat "$DIR/data/block.txt"; printf 'end\n'; } > "$HOME/AGENTS.md"
printf '#!/bin/sh\necho skill\n' > "$HOME/.cfg/bin/skill.sh"
chmod 755 "$HOME/.cfg/bin/skill.sh"
EOF
	mkdir -p "$TREE/data"
	printf 'rule a\n' >"$TREE/data/rule-a.mdc"
	printf 'managed block v1\n' >"$TREE/data/block.txt"
	case "$variant" in
	content) printf 'managed block v2\n' >"$TREE/data/block.txt" ;;
	mode) printf 'chmod 700 "$HOME/.cfg/bin/skill.sh"\n' >>"$TREE/setup.sh" ;;
	nosetup) rm "$TREE/setup.sh" ;;
	esac
}

run_diff() { # run_diff <old-tree> [new-tree] — output captured outside the trees;
	# a one-argument call is itself case 9's subject, so forward verbatim
	set +e
	"$DIFF" "$@" >"$OUT_FILE" 2>&1
	DIFF_RC=$?
	set -e
}

# Case 1 — identical trees: exit 0, both symlink targets normalized.
new_tree base
OLD="$TREE"
new_tree base
run_diff "$OLD" "$TREE"
assert_rc 0 "case 1: identical trees exit 0"
assert_contains "installer output identical" "case 1: identical verdict printed"

# Case 2 — a managed-block content change is found and named.
new_tree base
OLD="$TREE"
new_tree content
run_diff "$OLD" "$TREE"
assert_rc 1 "case 2: content difference exits 1"
assert_contains "AGENTS.md" "case 2: diff names the changed managed file"

# Case 3 — an exec-bit change on an installed file is found.
new_tree base
OLD="$TREE"
new_tree mode
run_diff "$OLD" "$TREE"
assert_rc 1 "case 3: mode difference exits 1"
assert_contains ".cfg/bin/skill.sh" "case 3: diff names the file whose mode changed"

# Case 4/5 — the kan-536 pitfall: no setup.sh at a tree root is refused
# before anything runs, and the error names the tree.
new_tree base
GOOD="$TREE"
new_tree nosetup
run_diff "$GOOD" "$TREE"
assert_rc 2 "case 4: setup.sh missing from the new tree exits 2"
assert_contains "has no setup.sh at its root" "case 4: error names the missing installer"
run_diff "$TREE" "$GOOD"
assert_rc 2 "case 5: setup.sh missing from the old tree exits 2"

# Case 6 — the normalization reads the tree path as a literal, never a glob
# pattern: identical trees under a path carrying a metacharacter diff clean.
new_tree base
OLD="$TREE"
new_tree base
GLOBROOT="$(mktemp -d "${TMPDIR:-/tmp}/installer-sandbox-diff-test.glob[1].XXXXXX")"
TREES+=("$GLOBROOT")
mv "$OLD" "$GLOBROOT/old"
mv "$TREE" "$GLOBROOT/new"
run_diff "$GLOBROOT/old" "$GLOBROOT/new"
assert_rc 0 "case 6: identical trees under a glob-metacharacter path exit 0"

# Case 7 — a TMPDIR that cannot hold the sandbox is the contract's exit 2
# (environment error), never the 1 reserved for "differ".
new_tree base
OLD="$TREE"
new_tree base
set +e
TMPDIR="$OLD/no-such-tmpdir" "$DIFF" "$OLD" "$TREE" >"$OUT_FILE" 2>&1
DIFF_RC=$?
set -e
assert_rc 2 "case 7: unusable TMPDIR exits 2"

# Case 8 — the retention half of the contract: a differing run keeps the
# sandbox, prints its path, and the printed directory exists.
new_tree base
OLD="$TREE"
new_tree content
run_diff "$OLD" "$TREE"
assert_rc 1 "case 8: differing run exits 1"
assert_contains "kept for inspection" "case 8: differing run names the kept sandbox"
KEPT="$(sed -n 's/.*kept for inspection: //p' "$OUT_FILE" | tail -n 1)"
if [ -n "$KEPT" ] && [ -d "$KEPT" ]; then
	pass "case 8: the kept sandbox path exists"
else
	fail "case 8: the kept sandbox path is missing or not a directory: $KEPT"
fi

# Case 9 — the bad-argument-count guard exits 2.
new_tree base
run_diff "$TREE"
assert_rc 2 "case 9: one-argument invocation exits 2"

if [ "$FAILURES" -eq 0 ]; then
	printf 'installer-sandbox-diff: all cases pass\n'
	exit 0
fi
printf 'installer-sandbox-diff: %d failure(s)\n' "$FAILURES" >&2
exit 1
