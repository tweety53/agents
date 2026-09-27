#!/usr/bin/env bash
# check-installed-citations.sh — fail when a path citation in an installed
# file names no root.
#
# A file `setup.sh` installs or copies is read by an agent standing in
# whatever tree the harness placed it — never necessarily this checkout. A
# bare citation like `README.md` in such a file is read against THAT tree,
# which is exactly how a bug ships. Which roots count (canonical definition:
# the kan-239 citation-roots spec — not restated here), how the installed set
# is derived by running the real setup.sh in a throwaway sandbox, and every
# classifier exclusion are in stats/internal/guard/installedcitations.go.
#
# GUARD CONTRACT:
#   0  clean — every citation in every installed file names a root, and
#      every scanned member's coverage is either non-zero or declared.
#   1  violations found — one or more citations name no root, or a member
#      reported zero coverage without a declared reason. Fix the named
#      file:line by prefixing the citation — or, for a line whose backticked
#      content is a non-path marker or illustration, declare it with the
#      `citations-guard:allow` line exemption; never silence a real hit.
#   2  environment — the guard cannot determine anything: the root override
#      is set but empty or not a directory, the sandboxed installer run
#      failed, or a file it needed to read could not be read or decoded. A
#      refusal writes its reason to stderr and NOTHING to stdout, so no
#      caller can read an absent report as a clean one.
#
# Takes no arguments. CHECK_INSTALLED_CITATIONS_ROOT, when set, replaces the
# scanned root (this checkout) — for tests only, never a normal invocation.
#
# Ported to Go (KAN-778), replacing this wrapper and
# the Python guard it ran: classification and the coverage decision
# (scripts/lib/coverage.sh's Go twin) now run in one process, so the
# sentinel-prefix hand-off between the two files is gone. The binary does
# not live in this checkout, so this shim exports FLOW_GUARD_REPO_ROOT —
# this script's physical parent directory, the root the Python derived from
# its own resolved location.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-installed-citations: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(cd -P "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-installed-citations 2 "check-installed-citations:" "$@"
