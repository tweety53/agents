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
#   visual-verify-pass.sh close <canonical-worktree> <name> <key> <motions named> \
#     --session-token mf-<t>
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
#   8. step 5 (decision parent-starts-the-stack): a short GET of each
#      survivor's --app URLs, any answer counting. A survivor nothing answers
#      for prints `NEXT: start-stack` (exit 0, no row begun): the parent starts
#      the stack and runs `prepare --continue --started`. Under --started it
#      ends the stage `stopped` with `STOP: stack — <urls> answers nothing
#      after the stack was started, for <worktree>`, exit 1.
#   9. step 6: check-dev-stack-fresh.sh <worktree> per survivor, in process.
#      Exit 1 prints `NEXT: restart-stack` (the parent restarts the stack and
#      runs `prepare --continue --started`); under --started it ends the stage
#      `stopped` with `STOP: fingerprint — <the guard's output>`, exit 1.
#      Exit 2 continues, the bundle reading `fingerprint: not declared —
#      <its stderr reason>`.
#  10. per row, the verifier's bundle
#      <worktree>/.superpowers/sdd/verify-prompt-<key>.md: the absolute
#      worktree; `changeRoot:`; the path of skills/flow/visual-verify-verifier.md
#      under FLOW_GUARD_REPO_ROOT; the `KEY=value` lines prepare-workspace.sh
#      prints; the resolved `setup`, `capture` and `specs` commands
#      (`<frontend-root>` the worktree, `<merge-base>` its merge base) and
#      `screenshots` root (under the regression checkout's worktree when one
#      is declared); `mockups:` and `mockup frame:` when declared; one `app:`
#      line per --app URL; the `## run` commands; `stack:` (`found running` or
#      `started by this stage`), `fingerprint:` and `verify: <key>@<short sha>`
#      — this session's newest green inline verify row whose HEAD resolves in
#      the worktree, read from `flow record dispatches`; `mockup sources:`,
#      every HTML or CSS path design.md cites that exists, or `none`, and a
#      `mockup decoded: <path> → <dir>` line per bundled page (one carrying
#      `<script type="__bundler/template">`), decoded into
#      <worktree>/.superpowers/sdd/verify-mockup-<key>/ (index.html plus each
#      manifest entry, gzip-decompressed when `compressed`); and the
#      paragraph of skills/flow/visual-verify.md opening `**The relay
#      contract.**`, its report path resolved to
#      <worktree>/.superpowers/sdd/verify-report-<key>.md — exit 2 when that
#      paragraph is missing. Under --continue a key containing `-rest` adds
#      `## Incomplete attempt`: the worktree's previous row's report path and
#      the `verdict <key> <line>` lines the stage file holds for it; a
#      `visual-verify-<n>` key adds `## Previous attempt`: every earlier
#      report path of the worktree (decision redispatch-bundle-by-path).
#  11. one `flow record dispatch begin -C <canonical-worktree> -change <name>
#      -role verifier -model opus -effort low -key <key> -session-token <t>
#      -diff-base <HEAD>[,<sibling HEAD>…] [-round-base <sha>]` per survivor
#      (decision verifier-stays-opus) — a sibling is a worktree whose diff
#      the survivor's `ui paths` matched under --repo; keys suffixed
#      `-<repository name>` when the set holds more than one worktree — each
#      appending `row <key> <worktree>` to the stage file and printing
#      `DISPATCH: <key> <bundle>`; then `NEXT: dispatch <n>`, exit 0. A
#      refused begin ends the stage `stopped` with `STOP: row — flow record
#      dispatch begin refused for <key>: <its output>`, exit 1.
#
# --continue (decision prepare-continue-flag) begins no stage and runs no
# step 1–3 check: it reads the survivors, their merge bases and siblings and
# the set's size from the stage file, and its own --app and --round-base
# flags, then runs steps 8–11. A key ending in a survivor's
# `-<repository name>` runs that worktree alone, its key unchanged.
#
# close reads the report <worktree>/.superpowers/sdd/verify-report-<key>.md,
# <worktree> the one the stage file's `row <key>` line names (a key with no
# row is exit 2), and acts on it once the verifier's turn has ended.
# <motions named> is the count of motions the dispatch prompt named, 0 for
# `motions: none`. In order:
#   1. no report file: the row ends `-outcome aborted`, the stage ends
#      `stopped`, and `STOP: no report — <path> was never written…`, exit 1.
#   2. check-verify-report.sh <report> <motions named> --design
#      <changeRoot>/design.md and, with `mockups` declared,
#      check-frame-list.sh <changeRoot>/design.md <mockups root> <report>, in
#      process (exit 2 from either is this call's exit 2). A
#      check-verify-report exit 1, or a check-frame-list exit 1 naming any gap
#      other than `no sidecar line`, ends the row `-outcome aborted`, appends
#      each guard verdict line to the stage file as `verdict <key> <line>`
#      (the remainder's bundle reads them), prints them and
#        NEXT: rest <key>-rest[-<n>][-<repository name>]       exit 0
#      — `-rest`, then `-rest-<n>` from 2, placed before any
#      `-<repository name>` suffix. No commit; the stage stays open.
#   3. the report's `## Verdict` block: `exits:` naming a non-zero exit is a
#      block, its `cause:` (environment when absent). The stage's first such
#      report of the worktree ends the row `blocked -cause <cause>` and prints
#        NEXT: redispatch visual-verify-<n>[-<repository name>]   (<n> from 2)
#      with no commit. From the second: cause `environment` ends the row the
#      same way and prints `NEXT: cure-environment <next key>`, no commit;
#      `test-failure` or `missing-fixture` ends the row `blocked`, commits
#      (step 12), ends the stage and prints `NEXT: loop`.
#   4. `departures:` above 0, or a `no sidecar line` frame-list gap alone:
#      prints the `## Verdict` block (and the gap line), commits, and prints
#      `NEXT: judge` — the row and the stage left open for **Blocking**.
#   5. otherwise the row ends `-outcome completed`, commits, ends the stage,
#      and prints `NEXT: done`.
# Every commit is step 12 (decision close-commits-final-reports), never made
# after `rest`, `redispatch` or `cure-environment`: the report's `- spec:`
# path, every PNG it names, the full app suite's spec and zip its `- full app
# suite` lines name, and <changeRoot>/visual-verification/ — each path that
# exists committed in the checkout it lives in (a regression checkout's
# worktree, or its main checkout, for the captures; the canonical worktree
# for visual-verification/) by `git add -- <paths>` then `git commit -m
# "test: record the <key> visual verification" -m "Flow-Key:
# flow.visual-verify" -- <paths>`, so a file staged beforehand stays staged
# and uncommitted. A checkout whose branch is not spectre/<name> prints
#   PUSH: git -C <checkout> push
# Nothing to commit makes no commit. Then, always:
#   ROUND-BASE: <worktree>=<git -C <worktree> rev-parse HEAD>
# The stage ends once every surviving worktree's newest row is closed —
# `stopped` when any is `blocked`, else `completed` — through `flow stage end
# -C <canonical-worktree> -command /flow -stage flow.visual-verify`, then
# `flow record render -change <name> -kind ledger -repo <canonical-worktree>`
# and `STAGE: <outcome>`. Rows end through `flow record dispatch end -C
# <canonical-worktree> -change <name> -key <key> -session-token <t>
# -outcome <outcome> [-cause <cause>]`; a failed end is exit 2. The parent
# reads the `## Verdict` alone (decision verdict-only-handback).
#
# The stage file, <canonical-worktree>/.superpowers/sdd/visual-verify-stage.txt,
# written once the stage has begun, is the only state a later call reads
# (decision stage-file-names-its-rows):
#   token <session token>
#   begin <RFC 3339 instant>
#   set <worktrees in the resolved set>
#   worktree <path> <repository name> <merge base> <round base or ->[ <sibling>,…]
#   row <key> <worktree>
#   verdict <key> <guard verdict line>       close, on a remainder
#   blocked <key>                            close, on a report with a non-zero exit
#   closed <key> <outcome>                   close, on each row end
#
# Exit 0 on a NEXT: line; 1 on a STOP: line; 2 a usage error (a missing or
# unknown subcommand or argument, a token not mf-<t>, no --harness, a
# --merge-base, --since or --round-base not <worktree>=<sha>, an --app not
# <worktree>:<app-root>=<url>, --started without --continue, a <motions
# named> that is not a count) or an unreadable input — the cause on stderr,
# nothing recorded for a usage error.
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
