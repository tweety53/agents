#!/usr/bin/env bash
# check-cleanup-complete.sh — verify that everything the cleanup registry says
# should be gone after /flow's archive run actually is.
#
# Usage: check-cleanup-complete.sh <repo> <change-name> <state-dir>
#
# Prints ONE verdict line to stdout:
#   COMPLETE: <reason>    every registry row whose lifetime ends at run 2 is gone
#   LEFTOVER: <breakdown> one or more are still present
#
# EITHER VERDICT MAY CARRY NOTES, appended after ` — `, and a `SKIPPED:` note
# says a row was NOT verified. It rides a COMPLETE line because a skip does not
# block — see the asymmetry paragraph below — so a caller that reads the token
# and discards the rest reports "cleanup verified" over a row nothing looked at,
# which is this guard's whole argument inverted one layer up. The rule that the
# clause is relayed word for word therefore lives with the consumer, in step 6
# of **Run 2 — the branch is merged** (`skills/flow-contracts/pipeline.md`),
# and is named here so a future editor of this line knows where it is kept.
#
# Exit 0 whenever a verdict was reached; exit 2 when it cannot answer at all —
# an unreadable repository or state directory, or a change name outside the
# allowlist. The VERDICT carries the answer, not the exit status
# — see check-finish-preflight.sh's header for why this repository separates
# them. Finish runs this once per repository, so each verdict names the
# repository it judged; a bare COMPLETE would be unattributable across several.
#
# WHY THE VERDICT WORDS DIFFER FROM check-unfinished-work.sh's. That guard says
# CLEAR/OUTSTANDING, this one says COMPLETE/LEFTOVER, and the two vocabularies
# are deliberately not shared. They answer OPPOSITE questions about absence, as
# the paragraph below spells out: there a missing file counts AGAINST the
# change, here a missing artifact IS the answer. A reader who carried the
# meaning of CLEAR across to COMPLETE would carry that inversion with it.
# check-finish-preflight.sh's third vocabulary (RUN1/RUN2/REFUSE) differs for a
# further reason again: it selects a procedure rather than reporting a state.
#
# Reporting a leftover is the whole point: run 2 previously assumed its own
# removals succeeded.
#
# THE SIX ROWS, AND WHY THEY ARE THESE SIX. The registry in
# skills/flow-contracts/pipeline.md owns the list; the rows whose lifetime
# ends at run 2 are the worktree, the local branch, the remote branch, the
# change directory, the proposal artifact source and the workspace database and
# bucket. The per-task diffs, the panel record and the session ledger are not
# checked separately because the registry removes them WITH the worktree — a
# surviving one of those is a surviving worktree, already reported. The state
# file is not checked at all: its row says it is never removed, so its presence
# is correct and reporting it would train the operator to ignore this guard's
# output. The claimed cache index is not checked for a different reason again,
# and it is the only row whose reason is an inability rather than a decision:
# its row says nothing in this pipeline removes it and nothing can, because the
# index is claimed by probing rather than derived from the change name and is
# never recorded, so by the time this guard runs there is nothing that could say
# which index to look at. A row this guard invented a check for would be
# checking an index it guessed.
#
# THE SIXTH ROW IS ANSWERED BY ASKING, NOT BY LOOKING, and it is the only one
# that is. The other five live in this repository's own git bookkeeping and
# filesystem, which this guard can read directly. A workspace's database and
# bucket live inside services this guard knows nothing about, and it must stay
# project-agnostic — it can hold neither `psql -l` nor any project's
# object-store client. So the project answers for them: it declares a
# `survivors` command in its .flow/project.md, this guard runs it, and its
# output and exit code are the row's verdict. Both are specified under "What
# `survivors` prints, and what its exit code means" in
# skills/flow-contracts/project-configuration.md, which is canonical; the
# reasoning for a third verb beside `create` and `remove` is under "Creation and
# cleanup" in skills/flow-contracts/workspace-isolation.md.
#
# "RAN THE REMOVAL" IS NOT "VERIFIED GONE", which is the whole reason that row
# is not settled from the removal command's exit code. A removal that reported
# success against a stale connection, and a bucket a policy refused to delete,
# both leave the row's promise broken with nothing having failed. This guard
# therefore never runs `remove` and never reads its result — it asks.
#
# THE TWO NON-EMPTY ANSWERS ARE NOT SYMMETRIC. A reported survivor BLOCKS the
# terminal state; a `survivors` command that could not reach its service is
# reported by name and with its exit code, and the run continues. Blocking an
# already-merged change over a service that happens to be stopped trades a
# stranded change for a few megabytes of stale storage, which is the wrong
# trade — the argument is under "Creation and cleanup" in
# skills/flow-contracts/workspace-isolation.md, together with why one non-zero
# exit is enough where a reader might expect two. A project that declares no
# `survivors` command, one whose configuration this guard cannot read — absent
# permission, but also a path that is not a regular file, and any scan of it that
# FAILED rather than found nothing — and one that declares the
# `## workspace isolation` section more than once, are all reported as SKIPPED
# for the same reason and with the same effect: skipped is never passed, and a
# skip does not block. An input this guard cannot resolve is reported rather than
# resolved to one of its readings, and a command whose failure would otherwise be
# indistinguishable from its negative answer is read by its exit status rather
# than by its empty output. A project that
# declares no `## workspace isolation` section at all is silent here, because a
# step whose artifact is already absent is a success.
#
# THAT DERIVATION IS DECLARED BELOW RATHER THAN LEFT IN PROSE, because a prose
# restatement of a table in another file is exactly what goes stale: a registry
# row added later whose lifetime ends at run 2 would not be checked here, and
# this guard would then report COMPLETE over a real leftover — the failure it
# exists to prevent, one layer up. Every registry row is named in exactly one of
# the two lists below, and TestCheckCleanupComplete/14 (ccRegistryCoupling,
# stats/internal/guard/check_cleanup_complete_test.go) reads BOTH the
# registry and these markers and fails when they disagree in either direction:
# a registry row with no line here, or a line here naming a row the registry no
# longer has. Adding a registry row therefore fails this guard's suite until
# someone decides, in writing, which list it belongs in.
#
# registry-row-checked: Worktree
# registry-row-checked: Local branch
# registry-row-checked: Remote branch
# registry-row-checked: Change directory
# registry-row-checked: Proposal artifact source
# registry-row-checked: Workspace database and bucket
# registry-row-not-checked: Per-task and review diffs — removed with the worktree
# registry-row-not-checked: Panel slot verbatim reports — removed with the worktree
# registry-row-not-checked: Panel record — lives in the store; nothing removes it
# registry-row-not-checked: SDD ledger — lives in the store; nothing removes it
# registry-row-not-checked: Rendered ledger and panel record — committed and
#   archived with the change, so nothing removes them and there is nothing for
#   this guard to find gone
# registry-row-not-checked: Brainstorm design document — removed with the worktree
# registry-row-not-checked: Dispatch context bundle — removed with the worktree
# registry-row-not-checked: Bugbot's or Mutation's throwaway worktree copy — created and removed
#   entirely within the review panel stage, immediately after that slot's dispatch
#   closes; it never survives to run 2, so there is nothing here for this guard to
#   find gone
# registry-row-not-checked: Archive branch — nothing in this pipeline removes it;
#   run 2 is terminal and the pull request it opens outlives the run, so there is
#   no later run to delete the branch it was opened from (kan-239)
# registry-row-not-checked: State file — never removed; it is the terminal record
# registry-row-not-checked: Claimed cache index — this pipeline removes nothing and this guard checks nothing; the index is probed rather than derived, so run 2 has no derivation to repeat. A project that writes its claim where a probe can see it may release it in its own `remove` command and report it through `survivors`; that is the project's tooling and this marker does not claim it
# registry-row-not-checked: Self-review context bundle — written on `defer` and
#   committed on the archive branch, which run 2 does not touch again after
#   step 9; it is removed only by `/flow-self-review`, a separate command run
#   later, not by anything this guard's cleanup checks derive from
#
# ABSENCE IS THE ANSWER HERE, not a gap in the evidence. The run-1 gate treats
# a file it cannot find as outstanding, because a missing record proves nothing
# about the work. This guard asks the opposite question — "is it gone?" — so a
# row it cannot find is that row answered. What must never be inferred is
# absence from a path that was never readable, which is why an unreadable
# repository and an unreadable state directory both refuse to answer instead.
#
# ONE RESIDUAL EXCEPTION TO THAT SENTENCE, NAMED RATHER THAN CLAIMED AWAY. The
# local-branch and remote-tracking-ref rows are the only two answered out of
# git's ref store rather than out of the filesystem, and for them the sentence
# above holds for every shape but one. A ref whose LOOSE FILE is unreadable or
# corrupt is now told apart from a ref that does not exist, and reported SKIPPED
# — see refState in stats/internal/guard/cleanupcomplete.go for the two
# commands that takes and why one will not do.
# A ref whose ANCESTOR DIRECTORY is unreadable is not: git enumerates the refs
# it can reach and reports nothing at all about the ones it cannot, in
# `show-ref`, in `rev-parse`, in `for-each-ref` and in `ls-remote` alike, with
# and without --quiet, so `refs/heads/spectre/` at mode 000 is indistinguishable
# from an empty `refs/heads/spectre/`. There is no channel to read, so there is
# nothing this guard can do but say so here.
#
# WHAT AN OPERATOR SHOULD KNOW. In that one shape a COMPLETE verdict does not
# prove the two branch rows are gone; it proves nothing was found where this
# guard was able to look. It is silent — no skip note, because the condition is
# unobservable from here — so it is the one leftover this guard cannot promise
# to report. If a run 2 that verified COMPLETE is followed by `git branch` still
# listing `spectre/<name>`, check the modes on `.git/refs/heads/` and its
# subdirectories before suspecting anything else. Every OTHER unreadable input
# in this guard — the repository, the state directory, packed-refs, the project
# configuration, the survivors command — refuses or reports SKIPPED.
#
# THE SURVIVORS COMMAND IS RUN UNDER A BOUNDED WAIT, and a timeout is a third
# skip reason rather than a fourth verdict. Run 2 is unattended, so a command
# that never returns strands an already-merged change short of FINISHED with
# nobody watching — and closing stdin, which this guard also does, only covers a
# command blocking on an interactive prompt. An unreachable host with no connect
# timeout, a lock wait, a frozen container and an infinite loop are all untouched
# by it. The bound is the 60 seconds **Worktree cleanup**
# (`skills/flow-contracts/pipeline.md`) already gives the project-supplied
# `## stop` command, and the OUTCOME deliberately differs: there a timeout is a
# failed check, because an un-stopped stack is a reason not to remove a worktree
# and there is no other answer to fall back on; here the contract already defines
# one for "the command said nothing about survivors" — the reported skip — and
# blocking an already-merged change over a slow command is what
# **Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`)
# forbids. The skip names the timeout distinctly from a non-zero exit, because
# "your command is slow" and "your command is broken" have different remedies.
#
# THE ARCHIVED CHANGE DIRECTORY IS NOT A LEFTOVER. The registry says the change
# directory is MOVED into spectre/changes/archive/ and never deleted, so only
# one still sitting at spectre/changes/<name>/ counts.
#
# A <name>-fix-N SUB-CHANGE COUNTS TOO, and used not to. Under spectre a
# sub-change is a FLAT SIBLING of its parent under spectre/changes/, never a
# directory inside it -- `spectre new` refuses an id that is not a single flat
# directory name -- so `spectre archive <name>` cannot reach one and each
# sub-change needs its own call (run 2 step 3,
# skills/flow-contracts/finish-contract-run2.md). Before this row the guard
# reported COMPLETE with the parent archived and the child left behind, and
# nothing anywhere said so.
#
# HOW TO HAND-VERIFY A LEFTOVER VERDICT (KAN-446). Check each row the
# breakdown names against the filesystem yourself. A worktree kept
# legitimately — the state file retains an entry whose removal failed, per
# state-file.md's `worktrees` contract — is the known structural shape; any
# other row still on the list is really there.

# Header corrected for the Go port (KAN-760): the registry-marker check and
# ref_state now live in Go, as named above. The body's reasoning for every
# branch is in stats/internal/guard/cleanupcomplete.go, beside the code.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-cleanup-complete: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-cleanup-complete 2 "check-cleanup-complete:" "$@"
