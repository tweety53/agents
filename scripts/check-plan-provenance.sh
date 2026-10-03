#!/usr/bin/env bash
# check-plan-provenance.sh — thin wrapper.
#
# All classification logic now lives in stats/internal/guard/planprovenance.go,
# the Go port (KAN-778) of the Python 3 guard this file used to exec. This file exists only so that
# .flow/project.md's declared lint command, the flow harness
# (TestCheckPlanProvenance), and any operator's muscle memory
# invoking this exact filename keep working unchanged — the CLI contract
# (exit codes, CHECK_PLAN_PROVENANCE_ROOT, argv, stdout/stderr shape) is
# byte-identical to what this script implemented directly before the
# reshape.
#
# WHY THE REWRITE: this guard's fence classifier was originally ~150 lines
# of Bash ERE matching, in the style of this repository's other guards.
# Five review panel passes and seven fix waves later it had shipped and
# then fixed every defect class enumerated in the module docstring
# planprovenance.go carries — three of them found by
# pass 5 alone, on a settled tree, after four of seven reviewers had
# already called it clean (full history and the canonical enumeration:
# that module docstring, and the kan-14
# plan-provenance design's "Post-review reshape" section — this comment does
# not restate the count, since a copied number is exactly what let an
# earlier, wrong count survive six review passes). The operator's call:
# stop patching an ERE allowlist and replace it with a real
# block-structure parser — a
# container stack recorded at fence-open time, and a language with O(1)
# indexing rather than bash 3.2's O(N) sequential array access. Python's
# standard library was chosen over a Markdown/CommonMark library because
# none is installed anywhere on this machine and this repository has no
# dependency management; the block parser is hand-written for the same
# reason the original Bash version was, just in a language that can build
# real structure instead of an allowlist of recognised prefixes.

# Header corrected for the Go port (KAN-778): the python3 probes this file
# ran before its exec are gone with the interpreter. flow-guard is built
# from this checkout, never taken from PATH: scripts/lib/flow-guard.sh
# derives it, and exits 2 (this guard's environment code) with the cause
# when it cannot — a missing go, or one present that fails, is exit 2,
# never 1.
#
# The default root is resolved here. The Python guard took REPO_ROOT from
# its own file's location; flow-guard runs from a build cache, so this shim
# resolves the checkout from its own location instead and hands it over in
# CHECK_PLAN_PROVENANCE_ROOT, only when that is unset — an explicit value,
# empty included, reaches the guard untouched.
set -euo pipefail
root="$(cd -P -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)" || {
  echo "check-plan-provenance: cannot resolve the checkout beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
export CHECK_PLAN_PROVENANCE_ROOT="${CHECK_PLAN_PROVENANCE_ROOT-$root}"
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-plan-provenance: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-plan-provenance 2 "check-plan-provenance:" "$@"
