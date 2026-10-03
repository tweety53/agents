#!/usr/bin/env bash
# check-panel-reproducers.sh <worktree> <change-name>
#
# THE CHANGE NAME IS A SECOND ARGUMENT because a change's findings are rows
# in the store, keyed by change name, not a file at a fixed path. `flow
# record findings -change <name> -C <worktree>` answers with the JSON array
# of that change's findings — one object per finding, carrying at least
# `ref`, `status` and `reproducer` — and this guard reads that array, never
# a rendered Markdown file.
#
# Every finding in the store must declare how it was reproduced, so that
# /flow's implement phase can run that command and require it to FAIL before dispatching a
# fix instruction built on it. A finding with no runnable check declares the
# exemption form `none — <reason>` instead — but the exemption form is
# available to MINOR findings only: a finding recorded at Important severity
# must carry a runnable command, and this guard rejects the exemption at
# Important (KAN-503). The exemption stays legal at Minor and at Critical,
# which already goes to the fix unconditionally.
#
# Reads ONLY the decoded JSON array `flow record findings` prints. It
# never parses a findings table, a Markdown file, or a marker block — for
# the same reason check-unfinished-work.sh does not: a hand-rolled parser
# has repeatedly hidden an open finding behind a shape it was not written to
# recognise, while a store-decoded field has no cells to split, no escaping
# rule and no table boundary to track.
#
# Exit codes:
#   0  every finding the store returned has exactly one well-formed
#      reproducer (a runnable command, or the `none — <reason>` exemption
#      on a finding whose severity is not Important)
#   1  violations found; each is reported on stderr
#   2  cannot answer at all — no worktree, no change name, a change name
#      outside the allowlist, the change's state record absent or
#      unreadable (a cross-repo change's guards answer only through its
#      canonical worktree, so a peer tree's store read is a cannot-answer
#      and never a clean verdict — KAN-658), or the store unreachable
#
# Ported to Go (KAN-778): the banned-character set is compiled in rather than
# sourced from reproducer-metachars.sh beside this script. The body's
# reasoning for every branch is in stats/internal/guard/panelreproducers.go.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-panel-reproducers: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-panel-reproducers 2 "check-panel-reproducers:" "$@"
