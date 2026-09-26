#!/usr/bin/env bash
# check-panel-reproducer-exit-contract.sh <worktree> <change-name>
#
# THE MECHANICAL EXIT-CODE CONTRACT CHECK FOR PANEL REPRODUCERS (KAN-554).
# kan-468's review panel supplied a reproducer for its Important finding
# whose exit-code condition was inverted: it read "defect not demonstrated"
# on the actually-buggy code, and the inversion was caught only by a
# human-grade deferred self-review pass — after the panel record had been
# rendered and relied on. Until this guard existed, nothing between the
# finding being recorded and the fix being dispatched executed the
# reproducer and compared its answer to the claim the finding makes. The
# lexical sibling, check-panel-reproducers.sh, validates the recorded
# command's SHAPE and never runs it; the prose in skills/flow/review-panel.md
# reads exit codes only at fix-dispatch time, as a bounce loop handled by
# attention. This guard is the check that runs instead of attention:
# before the panel relies on a reproducer, the reproducer must demonstrate.
#
# THE CLAIM, AND THE CONTRACT. An OPEN finding claims its defect is present
# in the tree under review, and its reproducer's claim is the same claim:
# run against this worktree, the command must produce the runner's "defect
# demonstrated" verdict under whichever exit-code convention it declares —
# a generic one demonstrates with a non-zero exit, a declared
# mutation-reproducer with exit 0 (the build succeeds with the mutation
# landed, KAN-568). A reproducer whose verdict here is "defect not
# demonstrated" contradicts the claim it is recorded under, whichever way
# its author meant the condition: that inverted reading is the defect class
# this guard exists to name, mechanically, before anything downstream is
# built on the reproducer. Findings whose status is not exactly `open` —
# fixed, deferred, withdrawn — claim nothing about the current tree, so
# their reproducers are skipped: a fixed finding's reproducer is SUPPOSED
# to read "not demonstrated" now, and demanding the opposite verdict of it
# would invert this guard into nonsense. The `none — <reason>` exemption
# and a bare `none` claim nothing runnable and are skipped too: the
# exemption's own rules (legal everywhere but Important) and the bare-none
# violation are check-panel-reproducers.sh's subjects, and that guard runs
# first in the pipeline that calls this one.
#
# THE RUNNER IS THE VERDICT, NEVER A RE-IMPLEMENTATION. Each runnable
# reproducer is executed by run-reproducer's own implementation (the Go
# port calls runReproducer in-process — the same code run-reproducer.sh
# execs, which the per-finding dispatch decisions run) — because the
# exit-code vocabulary (demonstrated / not demonstrated / refused /
# unverifiable / cannot answer), the argv-exec barrier, the resolved
# containment and the process-group kill all live there, tested by their
# own tests (stats/internal/guard/runreproducer_test.go). This
# guard adds exactly one thing: the comparison of the runner's verdict to
# the claim, as a gate with its own exit code. The runner is invoked BARE —
# worktree and command line, no --pre-fix-verdict — because at dispatch time
# the expected pre-fix verdict is always "demonstrated": passing the flag
# would turn every first verdict into the ambiguity refusal and make the
# gate unanswerable.
#
# THE INSTRUMENT AUDIT (KAN-606). The exit-code contract above reads only
# the verdict, and a verdict is only as good as the instrument behind it:
# kan-552's deferred self-review caught a reproducer whose greps had
# captured mismatched text and cited the wrong test's assertion — the
# script flipped green exactly as recorded, but what it demonstrated was
# not the defect — and only the round-1 re-run reviewers had audited the
# instrument at all. So before this guard invokes the runner, every
# runnable reproducer's own text must declare what it demonstrates and
# where: one `# demonstrates: <path>:<line>:<content>` line per cited
# location, in the grep-output shape the author pastes straight off the
# defect-present tree, inside the same first-10-lines window the
# mutation-reproducer convention (KAN-568) reads its own declaration from.
# Each citation is resolved against the finding's own tree — the tree the
# per-finding resolution in stats/internal/guard/panelexitcontract.go
# names, never unconditionally the guard's
# worktree argument: the declared
# path must be relative and stay inside that worktree — lexically (the
# sibling lexical guard's own shape classes) and then physically, realpath-
# resolved and required to remain under that worktree exactly as
# run-reproducer.sh resolves the reproducer's own path token —, the file
# must exist, the line must exist, and
# the declared content must appear on that line. Any citation that does not
# resolve is a violation and the reproducer is NOT run — the runner's
# verdict would answer a question this audit has already settled, and a
# "demonstrated" spent on an unresolvable instrument is exactly the green
# flip this guard exists to deny. A mutation-declared reproducer is exempt
# from the audit: the content it demonstrates is the mutated tree it builds
# at run time, not a location on this tree, and its instrument is audited
# by the KAN-568 sha-pin machinery — demanding resolution here would invert
# the audit into nonsense for exactly the convention KAN-568 added.
#
# Exit codes:
#   0  every open finding with a runnable reproducer demonstrated the
#      defect; findings claiming nothing about the current tree are skipped
#   1  violations found: at least one open finding's reproducer read "defect
#      not demonstrated" on this tree — the inverted class —, was refused
#      by the runner as unusable (a shape the lexical guard's earlier pass
#      did not see: the record changed after it ran, or the command resolves
#      outside the worktree through a symlink), or failed the instrument
#      audit — no `# demonstrates:` declaration within the first 10 lines,
#      a malformed one, a citation outside the worktree, a file, line or
#      content the tree does not carry, or a script that cannot be read to
#      audit at all; each named on stderr
#   2  cannot answer at all — usage, a worktree or change name that fails
#      containment, the store unreachable, the change's state record absent
#      or unreadable (a cross-repo change's guards answer only through its
#      canonical worktree, so a peer tree's store read is a cannot-answer
#      and never a clean verdict — KAN-658), jq failing, an open finding
#      carrying no reproducer field at all, a reproducer path token
#      resolving in several of the change's recorded worktrees, or any
#      reproducer the runner could not verdict (timeout, surviving process,
#      plumbing failure).
#      Cannot-answer outranks exit 1: a read that could not be completed is
#      never reported as a verdict, the same precedence its sibling guard
#      gives a null reproducer over a clean answer.

# Header corrected for the Go port (KAN-760): the runner is called in-process
# rather than as "$SCRIPT_DIR/run-reproducer.sh". The body's reasoning for
# every branch is in stats/internal/guard/panelexitcontract.go.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-panel-reproducer-exit-contract: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-panel-reproducer-exit-contract 2 "check-panel-reproducer-exit-contract:" "$@"
