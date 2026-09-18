#!/usr/bin/env bash
# installer-sandbox-diff.sh — run `setup.sh global` from two trees under two
# fresh HOME sandboxes and diff the installed result.
#
# Usage: installer-sandbox-diff.sh <old-tree> <new-tree>
#
#   scripts/installer-sandbox-diff.sh ../agents.main .worktrees/kan-536-…   # refactored installer vs main
#
# The installed result is compared as a normalized manifest of everything the
# installer created under each sandbox HOME, one line per entry:
#
#   l <rel-path> -> <readlink target, with the source tree replaced by <TREE>>
#   d <stat mode> <rel-path>
#   f <stat mode> <cksum crc size> <rel-path>
#
# Symlink targets are normalized against their own tree because every install
# link points back into the tree it was installed FROM — compared raw, an
# identical installer would diff on every single link. Regular files carry a
# cksum, so the managed blocks (CLAUDE.md, AGENTS.md, .zshrc, the generated
# agent-baseline.md) are compared by content; modes catch an exec bit lost or
# gained. The manifests are compared on one machine with one stat, so the
# BSD/GNU stat format difference never matters.
#
# The three per-attempt pitfalls this replaces (kan-536's hand-rolled diff
# needed three tries) are closed by construction:
#   - a fixture without setup.sh at its root — both trees are checked for
#     setup.sh before anything runs, and the error names the tree;
#   - a lost exec bit — the installer is invoked as `bash setup.sh`, which
#     needs no exec bit, while installed modes are still compared;
#   - a reused HOME that turned run two into an idempotent refresh — each
#     side gets its own fresh `mktemp -d` HOME, never either tree and never
#     each other. Identical trees therefore diff clean only if both runs
#     installed into two separate, empty sandboxes.
#
# MANAGED_BLOCK_POSTPROCESS is unset for both runs: it is the installer's own
# extension point, and a value exported by the caller would rewrite one side's
# zcode block and read as a difference.
#
# Exit codes:
#   0  the installed results are byte-identical; the sandbox is removed
#   1  they differ — the unified manifest diff goes to stdout and the whole
#      sandbox (both HOMEs, both installer logs, both manifests) is KEPT and
#      its path printed, so a difference can be walked in place
#   2  usage or environment error — bad argument count, an argument that is
#      not a directory, setup.sh missing from a tree root (named), or an
#      installer run that failed (log tail printed)

set -euo pipefail

die() { echo "ERROR: $*" >&2; exit 2; }

[[ $# -eq 2 ]] || die "Usage: $0 <old-tree> <new-tree>"

resolve_tree() { # resolve_tree <arg> — absolute path, must hold setup.sh
	local tree
	tree="$(cd "$1" 2>/dev/null && pwd)" || die "$1 is not a directory"
	[[ -f "$tree/setup.sh" ]] || die "$tree has no setup.sh at its root — the installer cannot be run from it"
	printf '%s\n' "$tree"
}

OLD_TREE="$(resolve_tree "$1")"
NEW_TREE="$(resolve_tree "$2")"

WORK="$(mktemp -d "${TMPDIR:-/tmp}/installer-sandbox-diff.XXXXXX")"
cleanup() {
	local rc=$?
	if (( rc == 0 )); then
		rm -rf "$WORK"
	else
		echo "sandboxes, logs and manifests kept for inspection: $WORK" >&2
	fi
}
trap cleanup EXIT

# BSD stat prints the full st_mode in octal; GNU stat -c '%a' the permission
# bits. Both are deterministic, and both manifests are produced (and compared)
# on this machine, so the two formats never meet.
case "$(uname -s)" in
Darwin) entry_mode() { stat -f '%p' "$1"; } ;;
*) entry_mode() { stat -c '%a' "$1"; } ;;
esac

# manifest <home> <tree> — the normalized manifest of one sandbox HOME
manifest() {
	local home="$1" tree="$2" rel path target
	while IFS= read -r rel; do
		path="$home/${rel#./}"
		if [[ -L "$path" ]]; then
			target="$(readlink "$path")"
			target="${target//$tree/<TREE>}"
			printf 'l %s -> %s\n' "${rel#./}" "$target"
		elif [[ -d "$path" ]]; then
			printf 'd %s %s\n' "$(entry_mode "$path")" "${rel#./}"
		else
			printf 'f %s %s %s\n' "$(entry_mode "$path")" "$(cksum <"$path")" "${rel#./}"
		fi
	done < <(cd "$home" && find . -mindepth 1 | LC_ALL=C sort)
}

run_install() { # run_install <tree> <home> <log>
	local tree="$1" home="$2" log="$3"
	(
		cd "$tree" &&
			export HOME="$home" &&
			unset MANAGED_BLOCK_POSTPROCESS &&
			bash ./setup.sh global
	) >"$log" 2>&1 || {
		echo "ERROR: setup.sh failed in $tree — last 15 lines of $log:" >&2
		tail -n 15 "$log" >&2 || true
		exit 2
	}
}

mkdir -p "$WORK/old-home" "$WORK/new-home"
run_install "$OLD_TREE" "$WORK/old-home" "$WORK/old-install.log"
run_install "$NEW_TREE" "$WORK/new-home" "$WORK/new-install.log"

manifest "$WORK/old-home" "$OLD_TREE" >"$WORK/old.manifest"
manifest "$WORK/new-home" "$NEW_TREE" >"$WORK/new.manifest"

if diff -u "$WORK/old.manifest" "$WORK/new.manifest" >"$WORK/manifest.diff"; then
	echo "installer output identical: $OLD_TREE vs $NEW_TREE"
else
	cat "$WORK/manifest.diff"
	echo "installer output differs: $OLD_TREE vs $NEW_TREE" >&2
	exit 1
fi
