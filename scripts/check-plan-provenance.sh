#!/usr/bin/env bash
# check-plan-provenance.sh — thin wrapper.
#
# All classification logic now lives in check-plan-provenance.py (Python 3,
# standard library only). This file exists only so that
# .myflow/project.md's declared lint command, the myflow harness
# (test-check-plan-provenance.sh), and any operator's muscle memory
# invoking this exact filename keep working unchanged — the CLI contract
# (exit codes, CHECK_PLAN_PROVENANCE_ROOT, argv, stdout/stderr shape) is
# byte-identical to what this script implemented directly before the
# reshape.
#
# WHY THE REWRITE: this guard's fence classifier was originally ~150 lines
# of Bash ERE matching, in the style of this repository's other guards.
# Five review panel passes and seven fix waves later it had shipped and
# then fixed every defect class enumerated in check-plan-provenance.py's
# module docstring — three of them found by
# pass 5 alone, on a settled tree, after four of seven reviewers had
# already called it clean (full history and the canonical enumeration:
# check-plan-provenance.py's own module docstring, and
# openspec/changes/kan-14-plan-provenance/design.md's "Post-review
# reshape" section — this comment does not restate the count, since a
# copied number is exactly what let an earlier, wrong count survive six
# review passes). The operator's call: stop patching an ERE
# allowlist and replace it with a real block-structure parser — a
# container stack recorded at fence-open time, and a language with O(1)
# indexing rather than bash 3.2's O(N) sequential array access. Python's
# standard library was chosen over a Markdown/CommonMark library because
# none is installed anywhere on this machine and this repository has no
# dependency management; the block parser is hand-written for the same
# reason the original Bash version was, just in a language that can build
# real structure instead of an allowlist of recognised prefixes.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
command -v python3 >/dev/null 2>&1 || {
  echo "check-plan-provenance.sh: python3 not found on PATH — cannot run the guard" >&2
  exit 2
}

# Critical 4 (pass-14 fix wave): `command -v` proves the FILE exists; it
# does not prove it RUNS. The realistic case is macOS without the Command
# Line Tools installed, where /usr/bin/python3 is a stub that exits 1 with
# "xcrun: error: invalid active developer path" and never reaches the
# guard at all. Exit 1 is this contract's "violations found" — so a CI
# gate on a machine with no toolchain reported violations that did not
# exist, printed no findings to explain them, and there was no way for the
# caller to tell that apart from a genuinely dirty plan.
#
# Probing with a trivial program is what distinguishes "python3 is a name
# on PATH" from "python3 is a working interpreter". Failure is exit 2,
# "cannot determine anything" — the same code an absent python3 already
# produced, which is the honest answer in both cases. stderr from the
# probe is passed through, because the operator's actual fix (`xcode-select
# --install`) is in that message and nothing else here can name it.
if ! python3 -c 'import sys; sys.exit(0)'; then
  echo "check-plan-provenance.sh: python3 is present but failed to run a trivial program (see above) — cannot run the guard" >&2
  exit 2
fi

exec python3 "$SCRIPT_DIR/check-plan-provenance.py" "$@"
