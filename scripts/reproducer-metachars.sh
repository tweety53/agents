#!/usr/bin/env bash
# reproducer-metachars.sh
#
# The single source of truth for the shell-metacharacter set a reproducer
# command line must never carry. check-panel-reproducers.sh and
# run-reproducer.sh each ban this same set — one lexically, at
# record-validation time, with no worktree to resolve against; the other at
# run time, against a real worktree and a real file on disk — and the two
# copies have already drifted twice inside this change: the backslash was
# added for finding F33 and the two quote characters for F45, each time in
# one file first and the other catching up later. There is no existing
# precedent under scripts/ for a sourced helper, so this is the smallest
# thing that removes the duplication: one variable definition, sourced by
# both callers, so a character added to the ban list is added once. The
# CHECKS built from this set are deliberately NOT moved here and stay
# duplicated in both scripts — that duplication is what lets each script
# apply the set in its own way (a pure lexical scan here, a per-token scan
# against a resolved file there); only the DATA the checks scan for has one
# home.
#
# Usage: `source "<dir of this file>/reproducer-metachars.sh"` binds
# REPRODUCER_METACHARS in the caller's shell. Not executable on its own.
REPRODUCER_METACHARS='|;&$`<>(){}~*?[]#\'\''"'
