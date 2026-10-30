#!/usr/bin/env bash
# visual-verify-pass.sh — the parent's side of `flow.visual-verify`
# (skills/flow/verify-and-handoff.md's **Visual verification** and
# skills/flow/visual-verify.md): one call before each verifier dispatch,
# one after (decision thin-parent-prepare-close). This header is canonical
# for every argument, step and closing line of its subcommands.
#
# Usage:
#   visual-verify-pass.sh prepare <canonical-worktree> <name> <key> \
#     --harness <h> --session-token mf-<t> --merge-base <worktree>=<sha>... \
#     [--worktree <worktree>]... [--since <worktree>=<sha>]... \
#     [--round-base <worktree>=<sha>]... [--app <worktree>:<app-root>=<url>]... \
#     [--miss] [--continue [--started]]
#
# Arguments:
#   <canonical-worktree>  the change's canonical worktree, where its rows are
#                         recorded and its stage file lives
#   <name>                the plain change name
#   <key>                 the stage's dispatch key (`visual-verify`, or the
#                         one **The loop** gives a round's re-run)
#   --harness, --session-token  the stage's begin mark; the token is mf-<t>
#   --merge-base          each worktree's merge base — one per worktree of the
#                         set, required without --continue
#   --worktree            every other worktree of the run's resolved set
#   --since               a fix run's since-close sha for a worktree: its
#                         diff for steps 1–2 runs from it instead
#   --round-base          a scoped round's base for a worktree
#   --app                 an app the stage may probe and pre-flight, the
#                         `<app-root>=<url>` argument check-visual-preflight.sh
#                         takes, prefixed with its worktree (decision
#                         parent-passes-apps)
#   --miss                a fix run with a missed defect: stop for the tooling
#                         analyst once step 3 passes (decision
#                         analyst-after-preflight)
#
# prepare runs, in order, stopping at the first closing line:
#   1. the kickoff check — a `regression checkout` `## apps` declares, or,
#      when the canonical worktree's branch records a flowBase, any `## apps`
#      repository, that the set holds no worktree in:
#        STOP: kickoff <repo>[ <repo>…]        exit 1, before any `flow` call
#   2. `flow stage begin -C <canonical-worktree> -command /flow -stage
#      flow.visual-verify -harness <h> -session-token <t> <name>`, then the
#      stage file (below).
#   3. steps 1–2: each worktree's `git diff --name-only <since|merge-base>...HEAD`
#      piped through check-visual-trigger.sh <worktree>, in process; every
#      other worktree's diff also through `check-visual-trigger.sh --repo
#      <repository name> <canonical-worktree>`, whose exit 0 makes the
#      canonical worktree survive. <repository name> is a worktree's main
#      checkout's basename.
#   4. `<canonical-worktree>/.superpowers/sdd/decision.json` recording
#      `visual.verify` `skipped` ends the stage `skipped`, printing one
#      `Visual: skipped at Decide — <reason>` line per survivor, then
#        NEXT: skip — skipped at Decide — <reason>      exit 0
#      No survivor ends the stage `completed`:
#        NEXT: skip — no worktree's diff matched `ui paths`   exit 0
#   5. --round-base: every passed round diff (`<sha>..HEAD`) empty ends the
#      stage `completed` with `NEXT: skip — round diff empty`; every changed
#      path under its worktree's `mockups` root prints `NEXT: compose-only`,
#      the stage left open. Exit 0.
#   6. step 3, once per stage: check-verify-green.sh <canonical-worktree>
#      <name> <token> <other worktrees…>; its exit 1 ends the stage `stopped`
#      with `STOP: verify-not-green — <its verdict line>`, exit 1. Then per
#      survivor prepare-workspace.sh piped into check-visual-preflight.sh
#      with that worktree's --app pairs (a survivor with none is a failing
#      check), and with `mockups` declared check-sidecar-reach.sh <capture
#      root> <mockups root>. Every failing check — an exit 1, or an exit 2
#      carrying its stderr, check-verify-green's exit 2 included — is
#      collected; then the stage ends `stopped` and one line names them all:
#        STOP: preflight — <guard>: <evidence>[; <guard>: <evidence>…]   exit 1
#      No verifier row is begun on any STOP.
#   7. --miss: NEXT: analyst          exit 0, no row begun
#
# The stage file, <canonical-worktree>/.superpowers/sdd/visual-verify-stage.txt,
# written once the stage has begun, is the only state a later call reads
# (decision stage-file-names-its-rows):
#   token <session token>
#   begin <RFC 3339 instant>
#   set <worktrees in the resolved set>
#   worktree <path> <repository name> <merge base> <round base or ->
#
# Exit 0 on a NEXT: line; 1 on a STOP: line; 2 a usage error (a missing or
# unknown subcommand or argument, a token not mf-<t>, no --harness, a
# --merge-base, --since or --round-base not <worktree>=<sha>, an --app not
# <worktree>:<app-root>=<url>, --started without --continue) or an unreadable
# input — the cause on stderr, nothing recorded for a usage error.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 with the cause when it
# cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "visual-verify-pass: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(flow_guard_root)" || exit 2
export FLOW_GUARD_REPO_ROOT
FLOW_GUARD_WORKSPACE_ISOLATION="$SCRIPT_DIR/check-workspace-isolation.sh"
export FLOW_GUARD_WORKSPACE_ISOLATION
flow_guard_exec visual-verify-pass 2 "visual-verify-pass:" "$@"
