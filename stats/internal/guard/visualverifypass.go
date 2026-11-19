package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// The contract below moved here from the header of the deleted
// scripts/visual-verify-pass.sh shim (KAN-1058).
//
// flow-guard visual-verify-pass — the parent's side of `flow.visual-verify`
// (skills/flow/verify-and-handoff.md's **Visual verification** and
// skills/flow/visual-verify.md): one call before each verifier dispatch,
// one after (decision thin-parent-prepare-close). This header is canonical
// for every argument, step and closing line of its subcommands.
//
// Usage:
//   flow-guard visual-verify-pass prepare <canonical-worktree> <name> <key> \
//     --harness <h> --session-token mf-<t> --merge-base <worktree>=<sha>... \
//     [--worktree <worktree>]... [--since <worktree>=<sha>]... \
//     [--round-base <worktree>=<sha>]... [--defects <worktree>=<frame>[,<frame>…]]... \
//     [--app <worktree>:<app-root>=<url>]... \
//     [--continue [--started]]
//   flow-guard visual-verify-pass close <canonical-worktree> <name> <key> <motions named> \
//     --session-token mf-<t>
//   flow-guard visual-verify-pass close --check-only <report> <motions named> [--mockups <dir>] [--design <design.md>] [--frames <frame>[,<frame>…]]
//   flow-guard visual-verify-pass measure <worktree> <name> <key> <spec path> [--prototype] [--frames <frame>[,<frame>…]]
//
// Arguments:
//   <canonical-worktree>  the change's canonical worktree, where its rows are
//                         recorded and its stage file lives
//   <name>                the plain change name
//   <key>                 the stage's dispatch key (`visual-verify`, or the
//                         one **The loop** gives a round's re-run)
//   --harness, --session-token  the stage's begin mark; the token is mf-<t>
//   --merge-base          each worktree's merge base — one per worktree of the
//                         set, required without --continue
//   --worktree            every other worktree of the run's resolved set
//   --since               a fix run's fix-start sha for a worktree: its
//                         diff for steps 1–2 runs from it instead
//   --round-base          a scoped round's base for a worktree
//   --defects             a scoped round's defects: frames for a worktree, given with its --round-base
//   --app                 an app the stage may probe and pre-flight, the
//                         `<app-root>=<url>` argument the in-process visual-preflight check (visual-verify-pass prepare)
//                         takes, prefixed with its worktree (decision
//                         parent-passes-apps); a worktree none names
//                         defaults to its project's `preflight apps` row,
//                         `<name>` and each `${VAR}` replaced from its
//                         prepare-workspace output, a pair whose root is
//                         absent dropped (kan-1086 self-review, F10)
//
// prepare runs, in order, stopping at the first closing line:
//   1. the kickoff check — a `regression checkout` `## apps` declares, or,
//      when the canonical worktree's branch records a flowBase, any `## apps`
//      repository, that the set holds no worktree in:
//        STOP: kickoff <repo>[ <repo>…]        exit 1, before any `flow` call
//      (not run when the decision records `visual.verify` `skipped`)
//   2. the stage file (below); no stage mark is made.
//   3. steps 1–2: each worktree's `git diff --name-only <since|merge-base>...HEAD`
//      piped through flow-guard check-visual-trigger <worktree>, in process; every
//      other worktree's diff also through `flow-guard check-visual-trigger --repo
//      <repository name> <canonical-worktree>`, whose exit 0 makes the
//      canonical worktree survive. <repository name> is a worktree's main
//      checkout's basename.
//   4. `<canonical-worktree>/.superpowers/sdd/decision.json` recording
//      `visual.verify` `skipped` ends the stage `skipped`, printing one
//      `Visual: skipped at Decide — <reason>` line per survivor, then
//        NEXT: skip — skipped at Decide — <reason>      exit 0
//      No survivor ends the stage `completed`:
//        NEXT: skip — no worktree's diff matched `ui paths`   exit 0
//   5. --round-base: every passed round diff (`<sha>..HEAD`) empty ends the
//      stage `completed` with `NEXT: skip — round diff empty`; every changed
//      path under its worktree's `mockups` root prints `NEXT: compose-only`,
//      the stage left open. Exit 0.
//   6. step 3, once per stage: the in-process verify-green check (visual-verify-pass close) <canonical-worktree>
//      <name> <token> <other worktrees…>; its exit 1 ends the stage `stopped`
//      with `STOP: verify-not-green — <its verdict line>`, exit 1. Then per
//      survivor flow-guard prepare-workspace piped into the in-process visual-preflight check (visual-verify-pass prepare)
//      with that worktree's --app pairs, or its declared `preflight apps`
//      (a survivor with neither is a failing check), and with `mockups` declared the in-process sidecar-reach check (visual-verify-pass prepare) <sidecar>
//      <mockups root> on each existing `<spec>.mockups` of the specs the
//      `specs` row prints (run from the capture root, `<frontend-root>` and
//      `<merge-base>` substituted; a non-zero exit is a failing `specs`
//      check), or on the whole <capture root> when no `specs` row is
//      declared (kan-1086 self-review, F11) — with `design repo` declared
//      too, flow-guard render-frames <worktree> <design.md> <those sidecars…>
//      first (kan-1096). Every failing check — an exit 1, or an exit 2
//      carrying its stderr, check-verify-green's exit 2 included — is
//      collected; then one line names them all (no stage mark is ended):
//        STOP: preflight — <guard>: <evidence>[; <guard>: <evidence>…]   exit 1
//      No verifier row is begun on any STOP.
//  6a. with no failing check, an unscoped pass (no --round-base) with one
//      surviving worktree — the set's other worktrees, a regression
//      checkout's among them, dispatch no verifier — and `mockups` declared
//      counts F, the distinct frame ids of the capture specs'
//      sidecar lines plus the declared frames no line maps (decision
//      shard-unscoped-single-worktree). Above 25 it writes one `spec <abs>`
//      line per capture spec and one `slice s<i>of<n> <frame,…>` line per
//      slice vvpSlices cuts (decisions four-shards-slices-grow,
//      slice-by-frame-id), and deletes every
//      <changeRoot>/visual-verification/record-s*.md. Otherwise it writes
//      nothing and the pass runs unsharded.
//   7. step 5 (decision parent-starts-the-stack): a short GET of each
//      survivor's --app URLs, any answer counting. A survivor nothing answers
//      for prints `NEXT: start-stack` (exit 0, no row begun): the parent starts
//      the stack and runs `prepare --continue --started`. Under --started it
//      ends the stage `stopped` with `STOP: stack — <urls> answers nothing
//      after the stack was started, for <worktree>`, exit 1.
//   8. step 6: flow-guard check-dev-stack-fresh <worktree> per survivor, in process.
//      Exit 1 prints `NEXT: restart-stack` (the parent restarts the stack and
//      runs `prepare --continue --started`); under --started it ends the stage
//      `stopped` with `STOP: fingerprint — <the guard's output>`, exit 1.
//      Exit 2 continues, the bundle reading `fingerprint: not declared —
//      <its stderr reason>`.
//  9a. on a stage with `slice` lines and a key with no shard segment
//      (decision prepare-captures-sharded): unless a `captured` line is there,
//      `setup` and one `capture <spec> --update-snapshots=all`, `<spec>`
//      every `spec` line's path a task's capture does not stand for
//      (rcTaskCaptured) space-joined (kan-1087: its verifier ran 4
//      specs, 70 tests, as one run of 8.5 minutes),
//      run from the capture root, then `captured` is appended; a non-zero
//      exit prints `SHARDS: none — <command> exit <n>`, drops the `slice`
//      lines and continues unsharded (decision
//      capture-failure-falls-back-serial).
//   9. per row, the verifier's bundle
//      <worktree>/.superpowers/sdd/verify-prompt-<key>.md: the absolute
//      worktree; `changeRoot:`; the path of skills/flow/visual-verify-verifier.md
//      under FLOW_GUARD_REPO_ROOT; the `KEY=value` lines flow-guard prepare-workspace
//      prints; the resolved `setup`, `capture` and `specs` commands
//      (`<frontend-root>` the worktree — or, when a sibling's diff made it
//      survive, that sibling — and `<merge-base>` its merge base; kan-1083
//      self-review, F3) and
//      `screenshots` root (under the regression checkout's worktree when one
//      is declared); on a row with no round base, `captured: <spec> <spec>…`
//      naming the specs a task's capture still stands for (rcTaskCaptured,
//      recapture.go, canonical) — the verifier's capture run leaves them
//      out and every view is still read; on a row with a round base, `captured: <spec> <spec>…`
//      when a flow-guard recapture receipt of every worktree given a
//      --round-base records a pass of those specs on exactly the source
//      files `<round base>..HEAD` holds there, each at HEAD's blob — the
//      verifier's capture run leaves them out; `mockups:` and `mockup frame:` when declared; one `app:`
//      line per --app URL; the `## run` commands; `stack:` (`found running` or
//      `started by this stage`), `fingerprint:` and `verify: <key>@<short sha>`
//      — this session's newest green inline verify row whose HEAD resolves in
//      the worktree, read from `flow record dispatches`; `mockup sources:`,
//      the absolute path of every HTML or CSS file design.md cites that
//      exists, or `none`, and a
//      `mockup decoded: <path> → <dir>` line per bundled page (one carrying
//      `<script type="__bundler/template">`), decoded into
//      <worktree>/.superpowers/sdd/verify-mockup-<key>/<page path, extension dropped, `/` as `-`>/ (index.html plus each
//      manifest entry, gzip-decompressed when `compressed`); and the
//      paragraph of skills/flow/visual-verify.md opening `**The relay
//      contract.**`, its report path resolved to
//      <worktree>/.superpowers/sdd/verify-report-<key>.md — exit 2 when that
//      paragraph is missing. Under --continue a key containing `-rest` adds
//      `## Incomplete attempt`: the worktree's previous row's report path and
//      the `verdict <key> <line>` lines the stage file holds for it; a
//      `visual-verify-<n>` key adds `## Previous attempt`: every earlier
//      report path of the worktree (decision redispatch-bundle-by-path) — of
//      its own shard index only, on a shard key. Every bundle carries a
//      `## Measure` section (decision bundle-names-measure-call): on a shard
//      key `slice: s<i>of<n> <slice>`, then
//        measure: flow-guard visual-verify-pass measure <worktree> <name> <key less any -rest[-<n>]> <spec path>[ --frames <slice>]
//      and a `## Report checks` section, the two guard calls close's step 2
//      runs on the report, for the verifier to run first (kan-1084 … kan-1087:
//      11 of 12 incomplete first reports carried a frame-list gap the
//      verifier's own check-verify-report run could not show it):
//        check: the in-process verify-report check (visual-verify-pass close) <report> <motions named> --design <changeRoot>/design.md[ --frames <slice>]
//        check: the in-process frame-list check (visual-verify-pass close) <changeRoot>/design.md <mockups root> <report>[ --frames <slice>]
//      the second only with `mockups` declared.
//  10. one `flow record dispatch begin -C <canonical-worktree> -change <name>
//      -role verifier -model <model> -effort <effort> -key <key> -session-token <t>
//      -diff-base <HEAD>[,<sibling HEAD>…] [-round-base <sha>]` per survivor,
//      the pair the first base pair of the `visual-verifier` allowed-pairs row
//      (a failed read or an unusable row
//      ends the stage `stopped` with `STOP: pair — <why>`, exit 1) — a sibling is a worktree whose diff
//      the survivor's `ui paths` matched under --repo; keys suffixed
//      `-<repository name>` when the set holds more than one worktree; on a
//      stage with `slice` lines one row per `slice` line instead, keyed
//      `<key>-s<i>of<n>`, the segment after any `-<repository name>`
//      (`visual-verify-gymie-s2of3`; decision shard-key-segment) — each
//      appending `row <key> <worktree>` to the stage file and printing
//      `DISPATCH: <key> <bundle> — description "<key>", subagent_type
//      flow-<effort>, model <model>, prompt opens "Role: visual-verifier"`, the
//      Agent call's four fields (kan-1086 self-review, F9); then
//      `NEXT: dispatch <n>`, exit 0. A
//      refused begin ends the stage `stopped` with `STOP: row — flow record
//      dispatch begin refused for <key>: <its output>`, exit 1. Before each
//      begin of a row with a round base and a key with no `-rest[-<n>]`
//      segment it writes <worktree>/.superpowers/sdd/visual-round-<key>.txt
//      holding `<round base> <RFC3339Nano instant>[ <frame>,…]`, the frames that worktree's --defects named, the instant read right
//      before that begin (decision round-file-per-row); a row begun with no
//      round base, or a `-rest` key, removes that file (decision
//      no-round-file-for-rest).
//
// --continue (decision prepare-continue-flag) begins no stage and runs no
// step 1–3 check: it reads the survivors, their merge bases and siblings and
// the set's size from the stage file, and its own --app and --round-base
// flags, then runs steps 8–11. A key ending in a survivor's
// `-<repository name>` runs that worktree alone, its key unchanged; a shard
// key, ending `-s<i>of<n>`, runs its one row. A
// survivor no --app pair names and no `preflight apps` row resolves is a
// usage error, exit 2, `no --app passed for <worktree>` (kan-1102
// self-review, F3).
//
// close reads the report <worktree>/.superpowers/sdd/verify-report-<key>.md,
// <worktree> the one the stage file's `row <key>` line names (a key with no
// row is exit 2), and acts on it once the verifier's turn has ended.
// <motions named> is the count of motions the dispatch prompt named, 0 for
// `motions: none`. In order:
//   1. no report file: the row ends `-outcome aborted`, no stage mark is ended,
//      and `STOP: no report — <path> was never written…`, exit 1.
//   2. the in-process verify-report check (visual-verify-pass close) <report> <motions named> --design
//      <changeRoot>/design.md and, with `mockups` declared,
//      the in-process frame-list check (visual-verify-pass close) <changeRoot>/design.md <mockups root> <report>, in
//      process (exit 2 from either is this call's exit 2); on a shard key both
//      are passed `--frames` with the stage file's `slice` of the key's shard
//      segment. A check-verify-report exit 1, or a check-frame-list exit 1 naming any gap
//      other than `no sidecar line`, ends the row `-outcome aborted`, appends
//      each guard verdict line to the stage file as `verdict <key> <line>`
//      (the remainder's bundle reads them), prints them, then one
//        NEXT: rest <key>-rest[-<n>][-<repository name>]       exit 0
//      — `-rest`, then `-rest-<n>` from 2, placed before any
//      `-<repository name>` suffix, the shard segment kept last
//      (`visual-verify-rest-s2of4`). No commit; the stage stays open.
//   3. the report's `## Verdict` block: `exits:` naming a non-zero exit is a
//      block, its `cause:` (environment when absent). The blocked count is the
//      stage's blocked rows of the worktree and the key's shard index. The
//      stage's first such report of the worktree and shard index ends the row
//      `blocked -cause <cause>` and prints
//        NEXT: redispatch visual-verify-<n>[-fix-<k>[-full]][-<repository name>][-s<i>of<n>]   (<n> from 2)
//      — the closing key's `-fix-<k>[-full]` segment, its shard segment and its
//      `-<repository name>` suffix kept (`visual-verify-2-fix-3`,
//      `visual-verify-2-s2of3`) — with no commit. From the second: cause `environment` ends the row the
//      same way and prints `NEXT: cure-environment <next key>`, no commit;
//      `test-failure` or `missing-fixture` ends the row `blocked`, commits
//      (step 12), ends the stage and prints `NEXT: loop`.
//   4. `departures:` above 0, or a `no sidecar line` frame-list gap alone:
//      prints the `## Verdict` block (and the gap line), commits, and prints
//      `NEXT: judge` — the row and the stage left open for **Blocking**.
//   On `NEXT: loop` and `NEXT: judge` a key carrying `-fix-<k>` prints, right
//   before its `NEXT:` line,
//        ESCALATE: <why> — the next round is a No-progress escalation
//   when <k> is 3 or more (`round <k> is the loop's third failed round or
//   later`), or when its `departures:` exceed those of the report the newest
//   visual-verify key opened a round with — the `**Opener:**` of the
//   `In-run fix <j>` tasks of <changeRoot>/tasks.md, <j> from <k> down, its
//   report in the row's worktree (`departures <opener's> → <this> since
//   <opener key>`).
//   5. otherwise the row ends `-outcome completed`, commits, ends the stage,
//      and prints `NEXT: done`. For a row with
//      <worktree>/.superpowers/sdd/visual-round-<key>.txt, the end carries
//      `-closes-loop` when every frame of the closing list — the frames
//      check-frame-list declares (with `mockups` declared), every frame of
//      <changeRoot>/visual-verification/prototype.mockups, and every frame of
//      visual-manifest-<key>.json's pairs — has a pair whose capture's mtime
//      is after the file's instant or whose spec step 10's `captured:` rule
//      names now, every pair whose capture differs from the
//      round base has its frame's own per-frame line in the report, every
//      frame the round file names under --defects has its own per-frame line
//      in the report, no pair's
//      `<spec>.mockups` differs from the round base and `git diff --name-only
//      <round base> -- <mockups root>` in <worktree> names nothing, and, when
//      <changeRoot>/visual-verification/prototype.mockups exists,
//      <worktree>/.superpowers/sdd/visual-prototype.txt equals the
//      `<sha-256 hex> <name>` lines of that map and each prototype/*.png now;
//      before `NEXT: done` it prints `CLOSES-LOOP: yes`, or
//      `CLOSES-LOOP: no — <the first failing frame or path and why>`; a row
//      with no such file prints neither. Every `completed` close, scoped or
//      not, then rewrites visual-prototype.txt with those lines when the map
//      exists.
// On the `loop`, `judge` and `done` paths (steps 3–5) a shard's close rewrites
// <changeRoot>/visual-verification.md as the
// `record-s<i>of<n>.md` parts that exist, joined in shard order by one blank
// line (decision record-parts-merged-by-close).
// Every commit is step 12 (decision close-commits-final-reports), never made
// after `rest`, `redispatch` or `cure-environment`: the report's `- spec:`
// path, every PNG it names, the full app suite's spec and zip its `- full app
// suite` lines name and the suite PNGs flow-guard resolve-visual-screenshots
// resolves for that spec — each path that
// exists committed in the checkout it lives in (a regression checkout's
// worktree, or its main checkout, for the captures) by `git add -- <paths>` then `git commit -m
// "test: record the <key> visual verification" -m "Flow-Key:
// flow.visual-verify" -- <paths>`, so a file staged beforehand stays staged
// and uncommitted. A checkout whose branch is not spectre/<name> prints
//   PUSH: git -C <checkout> push
// Nothing to commit makes no commit. Then, always:
//   ROUND-BASE: <worktree>=<git -C <worktree> rev-parse HEAD>
// The stage verdict is `stopped` when the newest row of any surviving worktree
// or shard index is `blocked`, else `completed`; no stage mark is made. Then
// `flow record render -change <name> -kind ledger -repo <canonical-worktree>`
// and `STAGE: <outcome>`. Rows end through `flow record dispatch end -C
// <canonical-worktree> -change <name> -key <key> -session-token <t>
// -outcome <outcome> [-cause <cause>]`; a failed end is exit 2. The parent
// reads the `## Verdict` alone (decision verdict-only-handback).
//
// measure (decision measure-one-call) is the verifier's step 9
// (skills/flow/visual-verify-verifier.md): its measuring chain in one call,
// in process, carrying on past every step's non-zero exit. <spec path> is the
// capture spec, resolved against <worktree>; <changeRoot> is
// <worktree>/<spec root>/changes/<name>. In order:
//   1. flow-guard resolve-visual-screenshots <worktree> <spec's basename>, its output
//      piped into flow-guard check-identical-baselines.
//   2. with the `<spec path>.mockups` sidecar and a declared `mockups` root —
//      or, under --prototype, <changeRoot>/visual-verification/prototype.mockups
//      and <changeRoot>/visual-verification/prototype/ — under --prototype
//      only the map's lines whose screenshot this spec resolves, copied to
//      <worktree>/.superpowers/sdd/visual-measure-<key>/prototype.mockups
//      (kan-1057 self-review, F2) — the resolver's output
//      piped into flow-guard compose-mockup-frames <sidecar> <mockups root>
//      <changeRoot>/visual-verification [<mockup frame>] (no fourth argument
//      under --prototype or without a declared `mockup frame`). Neither: the
//      manifest carries no pairs. With --frames, only the sidecar's lines
//      whose frame is in the slice — every `frame`, comment and blank line
//      kept — copied to
//      <worktree>/.superpowers/sdd/visual-measure-<key>/<spec basename>.mockups
//      and composed from there; a frame outside the slice is neither composed
//      nor measured, and its `.frame.png` is left alone (decision
//      manifest-per-shard).
//   3. per sidecar line (a pair), with the cropped `<stem>.frame.png` as the
//      frame when compose wrote one, else the mockup itself:
//      flow-guard measure-visual-properties <frame> <capture> --scale 1 --props
//      bands,seams into <worktree>/.superpowers/sdd/visual-measure-<key>/<frame id>.json,
//      and --props matrix into <changeRoot>/visual-verification/matrix/<frame id>.json.
//      A pair compose did not write is recorded with its non-zero compose exit
//      and not measured; every other pair still is.
//   4. the manifest <worktree>/.superpowers/sdd/visual-manifest-<key>.json:
//      `captures`, `identical` (the group lines), `steps` (resolve, identical
//      and compose, each {exit, stderr}) and `pairs[]` of {spec, frame, capture,
//      mockup, composite, diff, exits{resolve, identical, compose, bands,
//      matrix}, stderr, bands_file, matrix_file, bands, seams, over, cells} —
//      `bands` and `seams` the entries flagged by mvpOverLength and
//      mvpOverColour (measurevisualproperties.go) or missing/extra, `over`
//      the matrix's summary.over, each row carrying `figures`, `{<key>: {a,
//      b}}` per over key other than `unresolved`, read from the matrix row it
//      indexes (kan-1057 self-review, F20); `cells` step 10's matrix filled
//      from the matrix JSON (decision cells-in-measure), one `{row, class,
//      <column>…}` entry per matrix row, each of the 11 columns `n/a — <why>`
//      where vrNAColumns (verifyreport.go) lists it for the class,
//      `unresolved — <reason>` on an unresolved row, else `{a, b, delta}` —
//      per side each leaf the column reads — vvmColumnLeaves
//      (visualverifypass_measure.go) is canonical for which, per class — null for the side a missing or
//      extra row lacks, and per leaf the delta the matrix row carries
//      (vvmCells, visualverifypass_measure.go), null when the matrix did not
//      parse; an exit is null for a step not run. `spec` is the resolved spec path the pair was measured from. A
//      call merges into an existing manifest of its key, replacing only its
//      own spec's pairs (and any pair of a frame it measured): `captures` and
//      `identical` are unioned, existing lines first, and `steps` are this
//      call's; an absent or unparsable manifest is written fresh (kan-1102
//      self-review, F2).
//   5. with <worktree>/.superpowers/sdd/visual-round-<key>.txt present
//      (prepare's step 10), before `MANIFEST:` it prints
//        changed: <frame id> … — round base <sha>
//      — each frame of this call's pairs whose capture is absent, untracked
//      or differs from its blob at the round base (decision
//      changed-by-blob-hash) — or `changed: none — round base <sha>`; with no
//      such file, no `changed:` line. Then it prints
//        MANIFEST: <path>
//   Exit 0 when every step exited 0, 1 when any did not, 2 on a usage error
//   or an output directory or file it cannot write.
//
// The stage file, <canonical-worktree>/.superpowers/sdd/visual-verify-stage.txt,
// written once the stage has begun, is the only state a later call reads
// (decision stage-file-names-its-rows):
//   token <session token>
//   begin <RFC 3339 instant>
//   set <worktrees in the resolved set>
//   worktree <path> <repository name> <merge base> <round base or ->[ <sibling>=<its merge base>,…]
//   spec <absolute capture spec path>        prepare, on a sharded pass
//   slice s<i>of<n> <frame>[,<frame>…]       prepare, on a sharded pass
//   captured                                 prepare, once a sharded pass's capture ran
//   row <key> <worktree>
//   verdict <key> <guard verdict line>       close, on a remainder
//   blocked <key>                            close, on a report with a non-zero exit
//   closed <key> <outcome>                   close, on each row end
//
// Exit 0 on a NEXT: line; 1 on a STOP: line; 2 a usage error (a missing or
// unknown subcommand or argument, a token not mf-<t>, no --harness, a
// --merge-base, --since or --round-base not <worktree>=<sha>, an --app not
// <worktree>:<app-root>=<url>, a --defects not <worktree>=<frame>[,<frame>…]
// or with no --round-base for its worktree, --started without --continue, a <motions
// named> that is not a count) or an unreadable input — the cause on stderr,
// nothing recorded for a usage error.

// visualVerifyPass is `flow-guard visual-verify-pass`: the comment above is the contract for every argument, step and closing line of its subcommands.
// It runs the parent's side of `flow.visual-verify` (decision
// thin-parent-prepare-close) — the existing guards in process, the `flow`
// CLI from PATH — and closes on one `NEXT:` (exit 0) or `STOP:` (exit 1)
// line; exit 2 is a usage error or an unreadable input.
func init() { Registry["visual-verify-pass"] = visualVerifyPass }

const vvpUsage = "usage: flow-guard visual-verify-pass prepare <canonical-worktree> <name> <key> --harness <h> --session-token mf-<t> --merge-base <worktree>=<sha>... [--worktree <worktree>]... [--since <worktree>=<sha>]... [--round-base <worktree>=<sha>]... [--defects <worktree>=<frame>[,<frame>…]]... [--app <worktree>:<app-root>=<url>]... [--continue [--started]]\n" +
	"       flow-guard visual-verify-pass close <canonical-worktree> <name> <key> <motions named> --session-token mf-<t>\n" +
	"       flow-guard visual-verify-pass close --check-only <report> <motions named> [--mockups <dir>] [--design <design.md>] [--frames <frame>[,<frame>…]]\n" +
	"       flow-guard visual-verify-pass measure <worktree> <name> <key> <spec path> [--prototype] [--frames <frame>[,<frame>…]]"

// vvpExit unwinds a step to visualVerifyPass with its exit code.
type vvpExit int

// vvp is one call's state: its arguments and whether it began the stage.
type vvp struct {
	env            Env
	stdout, stderr io.Writer

	canon, name, key, harness, token string
	bases                            map[string]string // round base per worktree root, read by loadBases
	motions                          int
	worktrees                        []string          // canonical first, then each --worktree
	mergeBase, since, roundBase      map[string]string // by worktree
	apps                             map[string][]string
	defects                          map[string][]string // by worktree, --defects
	cont, started                    bool
	siblings                         map[string][]string // a worktree's siblings whose diff its `ui paths` matched
	setSize                          int
}

func visualVerifyPass(args []string, env Env, stdout, stderr io.Writer) (code int) {
	v := &vvp{env: env, stdout: stdout, stderr: stderr,
		mergeBase: map[string]string{}, since: map[string]string{}, roundBase: map[string]string{}, defects: map[string][]string{}, apps: map[string][]string{}}
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(vvpExit)
			if !ok {
				panic(r)
			}
			code = int(e)
		}
	}()
	if len(args) == 0 {
		v.usage("no subcommand")
	}
	switch args[0] {
	case "prepare":
		v.parse(args[1:], true)
		v.prepare()
	case "close":
		if len(args) > 1 && args[1] == "--check-only" {
			v.checkOnly(args[2:])
			return 0
		}
		v.parse(args[1:], false)
		v.close()
	case "measure":
		v.measure(args[1:])
	default:
		v.usage("unknown subcommand '" + args[0] + "'")
	}
	return 0
}

func (v *vvp) usage(why string) {
	fmt.Fprintf(v.stderr, "visual-verify-pass: %s\n%s\n", why, vvpUsage)
	Reason(v.env, "usage")
	panic(vvpExit(2))
}

// die prints the error and exits 2.
func (v *vvp) die(id, format string, a ...any) {
	fmt.Fprintf(v.stderr, "visual-verify-pass: "+format+"\n", a...)
	Reason(v.env, id)
	panic(vvpExit(2))
}

func (v *vvp) say(format string, a ...any) { fmt.Fprintf(v.stdout, format+"\n", a...) }

// next prints the closing `NEXT:` line and exits 0.
func (v *vvp) next(line string) {
	v.say("NEXT: %s", line)
	panic(vvpExit(0))
}

// stop prints the closing `STOP:` line and exits 1.
func (v *vvp) stop(id, outcome, line string) {
	v.say("STOP: %s", line)
	Reason(v.env, id)
	panic(vvpExit(1))
}

// abs resolves p against env.Dir, as every guard here does.
func (v *vvp) abs(p string) string { return pcAbs(v.env, p) }

func (v *vvp) parse(args []string, prepare bool) {
	pos := 3
	if !prepare {
		pos = 4
	}
	if len(args) < pos {
		v.usage("missing arguments")
	}
	v.canon, v.name, v.key = v.abs(args[0]), args[1], args[2]
	if !prepare {
		n, err := fmt.Sscan(args[3], &v.motions)
		if n != 1 || err != nil || v.motions < 0 {
			v.usage("<motions named> '" + args[3] + "' is not a count")
		}
	}
	if !plainChangeName(v.name) || v.key == "" {
		v.usage("change name '" + v.name + "' or key '" + v.key + "' is not plain")
	}
	v.worktrees = []string{v.canon}
	rest := args[pos:]
	val := func(i int) string {
		if i+1 >= len(rest) {
			v.usage(rest[i] + " takes a value")
		}
		return rest[i+1]
	}
	pair := func(flag, s string) (string, string) {
		w, sha, ok := strings.Cut(s, "=")
		if !ok || w == "" || sha == "" {
			v.usage(flag + " '" + s + "' is not <worktree>=<sha>")
		}
		return v.abs(w), sha
	}
	for i := 0; i < len(rest); i++ {
		switch f := rest[i]; {
		case f == "--session-token":
			v.token = val(i)
			i++
		case prepare && f == "--harness":
			v.harness = val(i)
			i++
		case prepare && f == "--worktree":
			v.worktrees = append(v.worktrees, v.abs(val(i)))
			i++
		case prepare && (f == "--merge-base" || f == "--since" || f == "--round-base"):
			w, sha := pair(f, val(i))
			map[string]map[string]string{"--merge-base": v.mergeBase, "--since": v.since, "--round-base": v.roundBase}[f][w] = sha
			i++
		case prepare && f == "--defects":
			d := val(i)
			w, frames, _ := strings.Cut(d, "=")
			if w == "" || frames == "" {
				v.usage("--defects '" + d + "' is not <worktree>=<frame>[,<frame>…]")
			}
			for _, fr := range strings.Split(frames, ",") {
				if fr == "" || strings.ContainsAny(fr, " \t\n") {
					v.usage("--defects '" + d + "' is not <worktree>=<frame>[,<frame>…]")
				}
				v.defects[v.abs(w)] = append(v.defects[v.abs(w)], fr)
			}
			i++
		case prepare && f == "--app":
			a := val(i)
			w, app, ok := strings.Cut(a, ":")
			root, url, ok2 := strings.Cut(app, "=")
			if !ok || !ok2 || w == "" || root == "" || url == "" {
				v.usage("--app '" + a + "' is not <worktree>:<app-root>=<url> — under zsh write \"${W}:<app-root>=<url>\"; \"$W:s…\" applies zsh's :s modifier")
			}
			v.apps[v.abs(w)] = append(v.apps[v.abs(w)], root+"="+url)
			i++
		case prepare && f == "--continue":
			v.cont = true
		case prepare && f == "--started":
			v.started = true
		default:
			v.usage("unknown argument '" + f + "'")
		}
	}
	for w := range v.defects {
		if v.roundBase[w] == "" {
			v.usage("--defects for " + w + " without --round-base")
		}
	}
	if !strings.HasPrefix(v.token, "mf-") || len(v.token) == 3 {
		v.usage("--session-token mf-<literal-token> is required")
	}
	if prepare && v.harness == "" {
		v.usage("--harness is required")
	}
	if v.started && !v.cont {
		v.usage("--started is given with --continue only")
	}
	if prepare && !v.cont {
		for _, w := range v.worktrees {
			if v.mergeBase[w] == "" {
				v.usage("no --merge-base for " + w)
			}
		}
	}
	for _, w := range v.worktrees {
		if !isDir(w) {
			v.die("not-a-directory", "%s is not a directory", w)
		}
	}
}

// flow runs one `flow` call from PATH, its combined output and exit code.
func (v *vvp) flow(args ...string) (string, int) {
	out, _, rc := pcFlow(v.env, true, args...)
	return strings.TrimSpace(string(out)), rc
}

// foldedGuards are the visual-verify checks run in process by prepare and
// close alone: none is a flow-guard subcommand (decision guard-checks-in-process).
var foldedGuards = map[string]Func{
	"check-frame-list":          checkFrameList,
	"check-verify-report":       checkVerifyReport,
	"check-verify-green":        checkVerifyGreen,
	"check-visual-preflight":    checkVisualPreflight,
	"check-visual-verification": checkVisualVerification,
	"check-sidecar-reach":       checkSidecarReach,
}

// guard runs a registered guard in process: its exit, stdout and stderr.
func (v *vvp) guard(name string, stdin io.Reader, args ...string) (int, string, string) {
	env := quietReasons(v.env)
	env.Stdin = stdin
	var out, errb bytes.Buffer
	fn := foldedGuards[name]
	if fn == nil {
		fn = Registry[name]
	}
	rc := fn(args, env, &out, &errb)
	return rc, strings.TrimSpace(out.String()), strings.TrimSpace(errb.String())
}

func (v *vvp) git(args ...string) (string, bool) { return capture(envGit(v.env)(args...)) }

// repoName is a worktree's repository name: its main checkout's basename,
// the name `check-visual-trigger.sh --repo` takes.
func (v *vvp) repoName(w string) string {
	if common, ok := v.git("-C", w, "rev-parse", "--path-format=absolute", "--git-common-dir"); ok {
		return filepath.Base(filepath.Dir(common))
	}
	return filepath.Base(w)
}

// The stage file, <canonical-worktree>/.superpowers/sdd/visual-verify-stage.txt,
// is the only state a later call reads (decision stage-file-names-its-rows):
//
//	token <session token>
//	begin <RFC 3339 instant>
//	set <worktrees in the resolved set>
//	worktree <path> <repository name> <merge base> <round base or ->[ <sibling>=<its merge base>,…]
//	spec <absolute capture spec path>
//	slice s<i>of<n> <frame>[,<frame>…]
//	captured
//	row <key> <worktree>
//	verdict <key> <guard verdict line>
//	blocked <key>
//	closed <key> <outcome>
func (v *vvp) stagePath() string {
	return filepath.Join(v.canon, ".superpowers", "sdd", "visual-verify-stage.txt")
}

func (v *vvp) writeStage(survivors []string) {
	var b strings.Builder
	fmt.Fprintf(&b, "token %s\nbegin %s\nset %d\n", v.token, time.Now().UTC().Format(time.RFC3339), len(v.worktrees))
	for _, w := range survivors {
		rb := v.roundBase[w]
		if rb == "" {
			rb = "-"
		}
		fmt.Fprintf(&b, "worktree %s %s %s %s", w, v.repoName(w), v.mergeBase[w], rb)
		if sib := v.siblings[w]; len(sib) > 0 {
			// each sibling with its merge base, which a --continue bundle's specs row needs
			pairs := make([]string, len(sib))
			for i, s := range sib {
				pairs[i] = s + "=" + v.mergeBase[s]
			}
			fmt.Fprintf(&b, " %s", strings.Join(pairs, ","))
		}
		b.WriteString("\n")
	}
	if err := os.MkdirAll(filepath.Dir(v.stagePath()), 0o755); err != nil {
		v.die("cannot-write-the-stage-file", "cannot write the stage file: %v", err)
	}
	if err := os.WriteFile(v.stagePath(), []byte(b.String()), 0o644); err != nil {
		v.die("cannot-write-the-stage-file", "cannot write the stage file: %v", err)
	}
}

func (v *vvp) appendStage(line string) {
	f, err := os.OpenFile(v.stagePath(), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		v.die("cannot-append-stage-file", "cannot append to the stage file: %v", err)
	}
	defer f.Close()
	fmt.Fprintln(f, line)
}

// readStage is the stage file's lines, split into their leading word and
// the rest; exit 2 when it cannot be read.
func (v *vvp) readStage() [][2]string {
	b, err := os.ReadFile(v.stagePath())
	if err != nil {
		v.die("cannot-read-the-stage-file", "cannot read the stage file %s — run prepare first: %v", v.stagePath(), err)
	}
	var out [][2]string
	for _, l := range strings.Split(strings.TrimRight(string(b), "\n"), "\n") {
		k, rest, _ := strings.Cut(l, " ")
		out = append(out, [2]string{k, rest})
	}
	return out
}

// roundFile is <sdd>/visual-round-<key>.txt, `<round base> <instant>`: prepare
// writes it for a scoped row right before its begin, and measure and close
// read it (decision round-file-per-row).
func roundFile(sdd, key string) string {
	return filepath.Join(sdd, "visual-round-"+key+".txt")
}

// readRound is key's round file: its round base, instant and defects:
// frames, ok false when the file is absent or not
// `<sha> <RFC3339Nano instant>[ <frame>,…]`.
func readRound(sdd, key string) (base string, at time.Time, defects []string, ok bool) {
	b, err := os.ReadFile(roundFile(sdd, key))
	if err != nil {
		return "", time.Time{}, nil, false
	}
	f := strings.Fields(string(b))
	if len(f) != 2 && len(f) != 3 {
		return "", time.Time{}, nil, false
	}
	if len(f) == 3 {
		defects = strings.Split(f[2], ",")
	}
	at, err = time.Parse(time.RFC3339Nano, f[1])
	return f[0], at, defects, err == nil
}

// captureChanged reports whether a file differs from its blob at base
// (decision changed-by-blob-hash): an empty path, a file git cannot hash, one
// absent at base, or a checkout that cannot resolve base is changed.
func (v *vvp) captureChanged(path, base string) bool {
	if path == "" {
		return true
	}
	dir := filepath.Dir(path)
	base = v.baseIn(dir, base)
	now, ok := v.git("-C", dir, "hash-object", path)
	if !ok {
		return true
	}
	then, ok := v.git("-C", dir, "rev-parse", "--verify", "--quiet", base+":./"+filepath.Base(path))
	return !ok || now != then
}

// cells is a worktree's `## visual verification` two-cell rows.
func (v *vvp) cells(w string) map[string]string {
	b, _ := os.ReadFile(filepath.Join(w, ".flow", "project.md"))
	return vvCells(string(b))
}

// cellPath is a cell's path value, backticks dropped, resolved against w.
func cellPath(w, cell string) string {
	cell = strings.Trim(cell, "`")
	if cell == "" || filepath.IsAbs(cell) {
		return cell
	}
	return filepath.Join(w, cell)
}

// basesFile is <sdd>/visual-round-<key>.bases, one `<worktree root> <sha>` line
// per --round-base prepare was given (decision round-base-per-checkout): the
// round file holds the canonical worktree's sha alone, which resolves in no
// other repository.
func basesFile(sdd, key string) string {
	return filepath.Join(sdd, "visual-round-"+key+".bases")
}

// loadBases reads key's bases file into v.bases; absent, it leaves them empty.
func (v *vvp) loadBases(sdd, key string) {
	v.bases = map[string]string{}
	b, err := os.ReadFile(basesFile(sdd, key))
	if err != nil {
		return
	}
	for _, l := range strings.Split(string(b), "\n") {
		if w, sha, ok := strings.Cut(l, " "); ok {
			v.bases[w] = strings.TrimSpace(sha)
		}
	}
}

// baseIn is the round base to read in dir's checkout: base itself when that
// checkout resolves it, else the sha recorded for the checkout's own root.
func (v *vvp) baseIn(dir, base string) string {
	if _, ok := v.git("-C", dir, "rev-parse", "--verify", "--quiet", base+"^{commit}"); ok {
		return base
	}
	root, ok := v.git("-C", dir, "rev-parse", "--show-toplevel")
	if !ok {
		return base
	}
	if real, err := filepath.EvalSymlinks(root); err == nil {
		root = real
	}
	for w, sha := range v.bases {
		if real, err := filepath.EvalSymlinks(w); err == nil && real == root {
			return sha
		}
	}
	return base
}
