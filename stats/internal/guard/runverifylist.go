// runverifylist.go is run-verify-list.sh ported to Go (KAN-1020).
//
// run-verify-list.sh — run one of a project's verify lists the way **Inline
// verify** (skills/flow/verify-and-handoff.md) prescribes, and print its
// `## Report` lines, so no run hand-builds the runner (the
// self-review-backlog-flow-6 self-review pass found the parent writing one
// per run, its stdin detach patched in on a second call).
//
// Usage: run-verify-list.sh <worktree> lint|test|visual [<change> <key> <session-token> <model> <effort> [<canonical-worktree>]]
//        flow-guard run-verify-list -only <file> <worktree> visual [<canonical-worktree>]
//        flow-guard run-verify-list -full <worktree> test|visual [<change> <key> <session-token> <model> <effort> [<canonical-worktree>]]
//
// THE NARROWED RUN (decision narrow-via-run-verify-list-flags; kan-1091: the
// failing-run protocol was three runs -- plain 40 minutes, a flake detector
// 36, a five-repeat 49). `-only <file>` runs the `specs` row only to group
// its failures by the change's specs, and puts the file's entries -- the
// visual-failing.txt a failing run wrote -- in the `verify` row's `<specs>`,
// followed by `--retries=1`. A test that passes on its retry is flaky, never
// red: the run exits by the tests that failed twice, and prints
//   flaky (<n>): <path:line> … — passed on retry: fix each at its source
// whatever its exit. It takes the `visual` list only and no recording
// arguments: it records no dispatch row, never reads or writes the green
// cache, and so never stands as the `verify` gate. A missing or empty
// -only file, or a `verify` cell with no `<specs>` exits 2 before
// any command runs.
//
// RECORDING (kan-1033 self-review: the parent hand-recorded the verify row and
// back-filled a forgotten begin with a guessed start). Given the five trailing
// arguments, the list records its own `verifier` row through `flow record
// dispatch`: `begin` (`-agent-id inline`, `-diff-base` the HEAD of
// <worktree>, which check-verify-green judges the row by) before its first command, and `end`
// after its last: `-outcome completed` when the list exited 0, `-outcome
// blocked -cause test-failure` when a command failed, `-outcome skipped` when
// every command of a `test` list was skipped. No `flow`
// on PATH exits 2 before any command runs. The optional sixth recording
// argument, <canonical-worktree>, is the `-C` of both calls — the worktree
// holding spectre/changes/<change> when <worktree> is another checkout of a
// multi-repo change (kan-1014 self-review); it defaults to <worktree>. A `begin` the store refuses prints
// `run-verify-list: flow record dispatch begin refused — no command ran: <its
// output>` on stderr and exits 4 before any command (kan-1032 self-review: the
// list ran in full, its row unrecorded, and was re-run only to get the row); a
// failed `end` prints `run-verify-list: flow record dispatch end failed: <its
// output>` on stderr and never changes the exit code. A key folding to bare
// `verify` is refused, exit 2 before the `begin`, when `flow state get` of the
// change lists more than one worktree: each repository's row then carries its
// `-<repository name>` suffix (kan-1099 self-review, F5: `verify-fix-5` was
// recorded for one repository of a four-worktree change). Before its `begin`
// the list reads `flow record dispatches` from that `-C` and ends `-outcome
// aborted` each row of its session token, role `verifier`, agent `inline`,
// no outcome, a key other than its own folding by vgFixSuffix to its own
// key's group: the row a run killed before its end left open (kan-1099
// self-review, F11). An unreadable store skips this; a failed end prints
// `run-verify-list: flow record dispatch end of superseded row <key> failed:
// <its output>` on stderr and the list still runs.
//
// THE LIST is every non-blank, non-`#` line inside a ```bash fence of
// `project-get.sh <worktree> <list>`'s output, in the order printed; `lint`
// appends `flow-guard check-spec-reach <worktree>`. A `# inputs: <pathspec>...`
// line of the `test` list declares what the one command below it reads, and a
// `# suite: <name>` line names the one command below it as that suite: it runs
// under `flow suite record -C <worktree> -suite <name> --`, so its runtime
// reaches the store `flow suite list` reads, and runs bare when no `flow` is
// on PATH; every other `#` line is a comment. A line holding a `<word>`
// placeholder — `./gradlew verifyChange :<module>:test --tests '<selectors>'`
// — is a template, not a command: the shell reads `<module>` as a redirect,
// so the list is refused before anything runs (kan-983 self-review). Each command runs from
// <worktree> in a fresh `bash -c`, stdin detached (</dev/null) so none waits
// on a terminal, and the whole list always runs: a failure never stops it.
// A `test` list's commands run concurrently, each into its own output file,
// reported in list order once every one has exited (kan-920 self-review: the
// guard harnesses and `go test -race` ran one after the other, 6½ minutes; run
// together they took 307 s, each slowed by the other's CPU). A `lint` list
// runs in order, one command at a time.
// A `test` list then ends with the `## visual verification` `verify` row,
// run once in the foreground after every other command has exited, `<specs>`
// filled as THE SMOKE FALLBACK states, from the section's `regression checkout` when declared
// and <worktree> otherwise (verify-list-runs-visual, kan-1047: the run never
// ran the visual verify by hand) -- when the row is declared,
// `<worktree>/.superpowers/sdd/decision.json` does not record
// `visual.verify` `skipped`, HEAD is not the green sha, and `git diff
// --name-only <merge-base>..HEAD` touches a `ui paths` glob, the merge base
// the commit `flow.kickoff` branched <worktree> from -- its value in the
// worktrees map of `flow state get -C <canonical-worktree or worktree>
// <change>`, <change> the recording argument or the branch name less
// `spectre/` (decision selection-base-kickoff-commit) -- and only with no
// such value that of HEAD and resolve-base-branch's base (its `origin/` ref
// when fetched), printed as `- kickoff base: not recorded for <worktree> —
// selecting from merge-base HEAD <ref>`. The same base fills the `specs`
// row's `<merge-base>`. A base that does not resolve runs it. It reports and fails the
// list like any other command; a `lint` list never runs it.
// A `visual` list runs that `verify` row alone, as a `test` list's tail
// would, and never the `## lint` or `## test` commands; when the conditions
// above skip it, it prints `- visual verify — not run: …` and exits 0
// (kan-1052 self-review, F2: /flow-fast and /flow-interactive had no command
// that ran the row). Unlike a `test` list's tail, it never falls back to a
// `regression checkout`'s main checkout: one holding no worktree on
// <worktree>'s branch prints `- visual verify — not run: …` naming it and
// exits 1 (kan-1052 self-review, R1: the main checkout's build is not the
// change's).
// The list runs `flow-guard prepare-workspace <worktree>` itself before the
// `begin` and gives every command the KEY=value lines it prints, a variable
// the caller's environment already sets winning -- save in a worktree on a
// named branch that names no change (pwChangeName), /flow-fast's and
// /flow-interactive's, which carries no workspace; a prepare-workspace that
// exits non-zero prints its output and exits 2 before any command (kan-1099
// self-review, F16). This script retries
// nothing — the one inline re-run of a failing command stays the caller's,
// and a race-marked command takes **The loop** whatever its re-run exits.
//
// THE GREEN SHA (KAN-887's self-review: verify re-ran the whole list minutes
// after a green test-list run, the delta a test rename). A `test` run that
// exits 0 on a tree clean before and after it — `git status --porcelain
// --untracked-files=all` empty — writes HEAD to the worktree's own git dir
// (`git rev-parse --git-path flow-test-green`), never into the tree. A later
// `test` run on a clean tree skips every command when HEAD is that sha, and
// otherwise skips each command whose `# inputs:` pathspecs, plus
// `.flow/project.md` itself (an edited command must run), `git diff --quiet
// <sha> HEAD` finds untouched. A command declaring no inputs reads everything
// and always runs; a dirty tree, no git, no recorded sha or a failing diff
// runs the whole list. A skip passes and prints instead of output:
//   - `<command>` — skipped, no input changed since <sha> passed the list
// ponytail: the sha ignores the exported KEY=value lines; a run under other
// isolation values reuses the green, so key the file by them if they ever vary.
//
// THE GREEN COMMANDS (kan-1103 self-review, F10: a re-run after one failed
// command re-ran every command already green at the same HEAD). A `test` run
// on a tree clean before and after it writes `flow-test-green-cmds` to the
// worktree's own git dir: HEAD on line one, then each command that exited 0
// -- the visual verify's command as it ran included -- one per line, keeping
// the earlier lines when the file already named that HEAD. A later `test` run
// at the same clean HEAD skips each listed command, printing
//   - `<command>` — skipped, passed at <sha>
// A failing run still runs every other command, and the whole-list sha above
// and its skip are unchanged.
//
// The visual row's coverage also keys on every sibling worktree's HEAD in the
// change's state record, since its specs read the siblings' code (kan-1087
// self-review, F7): `flow-test-green` carries a second line `siblings
// <path>=<HEAD> …`, and the row is covered, or skipped as passed at HEAD, only
// while that line equals the current one.
//
// VERIFY FAST. A `test` run in a worktree whose
// `.superpowers/sdd/fix-run.json` -- the fix-run judge's verdict, written by
// flow-guard fix-run-floor, whose header is canonical for it -- says verify
// fast skips every command and the visual verify, printing
//   - `<command>` — skipped, fix-run verify fast: the change is prose
// writes neither green file and ends its row `skipped`. The verdict holds
// only while the change's class, re-derived from the tree, is still prose,
// and only while `flow record findings` holds no Critical or Important
// finding of the verdict's round or a later one; otherwise the whole list
// runs, a voided verdict saying so on a line of its own. A `lint` list is
// never shortened.
//
// A SKIPPED SUITE. A `test` command opening with `flow-guard unchanged
// <pathspecs> ||` -- an optional path before `flow-guard`, each pathspec a
// bare word of letters, digits and `_ . / : ! @ % + = , -` or one
// single-quoted string holding no quote -- has that guard run in process from
// <worktree> once neither green file skipped it. On the guard's exit 0 the
// command prints
//   - `<command>` — skipped, <the guard's line>
// runs no child, records no suite row, joins neither green file as passed,
// and counts as skipped for the row's outcome (decision
// skipped-suite-records-no-row: the skipped suite was stored as a passing run
// of a fraction of a second). Any other shape -- a command substitution among
// the pathspecs included -- and any other exit runs the command as before.
//
// THE FULL SUITE (decisions full-is-the-list-unskipped,
// full-runs-the-visual-suite-whole, full-rows-carry-a-suffix,
// full-failure-names-its-escape). `-full`, a leading flag beside `-only`,
// runs the `test` or `visual` list with every skip off, only
// when the operator asks for it. Every command's environment gains
// `FLOW_VERIFY_FULL=1` after every inherited entry, so `flow-guard unchanged`
// answers 1 and each suite runs. Neither green file skips a command and A
// SKIPPED SUITE's check is not run, while both files are written as on any
// run. A judged verify fast prints
//   - fix-run verify fast set aside — -full runs the whole list
// skips nothing and leaves fix-run.json as it was. A suite is recorded as
// `<name>-full` and the visual row as `visual-full`. The visual `verify` row
// runs whenever the cell is declared, `<specs>` emptied and the `specs` row
// not run, whatever the diff, the green sha and decision.json say, from the
// `regression checkout` resolved as on any run. A failing command the green
// sha covers at this HEAD, or the green-commands file lists, gains
//   escaped the default list — it passed at <sha>
// in its block, after its output lines. `-full` with `-only`, with the
// `lint` list, or with no <worktree> and list exits
// 2 before any command runs.
//
// THE OPERATOR'S SWITCH. `-full` exits 2 before any command unless
// `FLOW_OPERATOR_FULL_SUITE=1` is in its environment. Only the operator sets
// that variable; an agent never does, and hands the operator the command
// instead. A run without `-full` puts `FLOW_VERIFY_FULL=` after every
// inherited entry of each command's environment, so no caller's export turns
// a default run into the full suite.
//
// THE SMOKE FALLBACK. Outside `-full` the visual `verify` row never runs with
// `<specs>` empty. With no `specs` row, a `specs` row with no merge base, one
// that exits non-zero or one that prints nothing, `<specs>` is filled with
// `--grep @smoke` -- the tests the regression suite tags `@smoke` -- and a
// stderr line names which of the four it was; the missing merge base and the
// failing `specs` row still fail the list. A `verify` cell with no `<specs>`
// can only run whole: it is not run and fails the list.
//
// THE ROUND'S SELECTION (kan-1099: every verify re-run selected from the
// change's whole diff, which a shared theme file makes every capture, 35
// minutes a round). A re-run -- a recording <key> check-verify-green's
// vgFixSuffix folds, `verify-fix-<k>…` or `verify-<k>…` -- fills the `specs`
// row's `<merge-base>` with the round base of `<frontend-root>`: the
// `diffBase` of the newest `verifier` row of the change, a `verify…` key of
// another round, that resolves as a commit there -- the HEAD its previous
// verify ran at -- read from `flow record dispatches -change <change> -C
// <canonical-worktree or worktree>`, and printed as
//   - visual selection: the round's diff — <frontend-root> since <short sha>
// With no such row, or no `flow`, the kickoff base stands. A plain re-run
// adds to `<specs>` the spec file of each visual-failing.txt entry the
// selection does not already name, printed as
//   - visual selection: plus <n> spec file(s) of visual-failing.txt
// so its green row covers the round's specs, the earlier failures and the
// tail the `specs` row prints; green, it empties visual-failing.txt. `-only`, which carries no key, selects
// as the newest verify row of <worktree> did: from its round base when that
// row's key is a re-run's, from the kickoff base otherwise.
// ponytail: a re-run whose `specs` row maps nothing runs the smoke set alone,
// without the visual-failing.txt files -- `--grep` would intersect them; run
// the row twice if a project's `specs` row ever prints no tail of its own.
//
// A WIDE SELECTION (kan-1089, kan-1091, kan-1099: a change touching every
// screen selected 646-679 of 680 tests, and 17 such runs of 35-50 minutes
// each all failed). A `specs` output naming more than rvlWideSelection (40)
// spec files is narrowed, never refused -- the cap 5554c41a refused on
// stopped a run on a question: `<specs>` takes the tokens whose spec file
// the diff of the checkout the row runs from names -- since its round base on
// a re-run, its kickoff base otherwise, none when it has no base -- then the
// tokens the `specs` row prints with `<merge-base>` filled with `HEAD`, the
// tail it prints for an empty diff, and prints
//   - visual selection: <n> spec files reached, more than 40 — narrowed to the <m> the change edits plus the smoke tail; the rest runs only under -full, which the operator starts
// Narrowed to nothing, it runs the smoke set (THE SMOKE FALLBACK). The
// failures still group by the whole output, and a `-only` failing list runs
// whatever its size.
//
// THE LINT STAMP (kan-1081 self-review, F1: a request ended landed with no
// evidence its lint list ran green). A `lint` run — the appended
// check-spec-reach included — that exits 0 on a tree clean before and after
// it writes HEAD plus a newline to the worktree's own git dir (`git rev-parse
// --git-path flow-lint-green`), never into the tree; a failing or dirty run
// leaves an earlier stamp as it was. Nothing is skipped on it: `flow record
// request end -outcome landed` and `flow interactive land` read it, refusing
// a commit the stamp does not hold.
//
// OUTPUT, one block per command in list order, stdout and stderr together:
//   - `<command>` — exit <n>
//     <each output line, indented two spaces; past 20 lines,
//     `(truncated — last 20 of <N> lines)` and then the last 20>
//     `data race — never environmental` — when the command's whole output,
//     not only its tail, carries `WARNING: DATA RACE` or `race detected
//     during execution of test`
//     `full output: <path>` — on a non-zero exit inside a git worktree, the
//     command's whole output copied to `<git-path flow-verify-failed>/<list>-<n>.log`
//     (<n> its 0-based list position, `spec-reach` for the appended
//     check-spec-reach.sh), never into the tree; each `lint` or `test` run
//     first removes that list's earlier logs, so a printed path holds this
//     run's output
//     `result files: <dir>` — beside it, when the run wrote JUnit `TEST-*.xml`
//     files under the worktree: each copied into a per-run
//     `<git-path flow-verify-failed>/<list>-<n>-results-<ns>/` no later run
//     removes, so a flake's first failing run stays traceable
//   - the visual verify, when it ran, last, as `- `<verify row>` — exit <n>`,
//     its kept output `visual-<mode>-r<N>.log`, <mode> `plain` or `only`
//     and <N> one past the highest round already kept, with
//     `result files: <git-path flow-verify-failed>/visual-results-r<N>` --
//     the newest `test-results/` directory the run wrote, copied -- both kept
//     until the worktree is removed (decision
//     per-round-artifacts-round-counter); a failing one also adds
//     `failing list: <git-path flow-verify-failed>/visual-failing.txt` (or
//     `failing list: none parsed`, the file then emptied), the tests its
//     Playwright summary names as failed, interrupted or flaky, and
//     `failing in the change's specs (<n>): …` / `failing elsewhere (<n>): …`,
//     grouped by the `specs` row's output (rvlKeepFailures)
//
// Exit codes:
//   0  every command exited 0
//   1  at least one command exited non-zero, or a `visual` list refused a
//      `regression checkout` holding no worktree on the branch
//   2  usage (not `<worktree> lint|test|visual` plus none, all five, or all six recording
//      arguments, no flow binary on PATH for those, or <worktree> is not a
//      directory), a list line holding a `<word>` placeholder — its stderr
//      line names it — or project-get.sh exit 2 — its stderr line says why,
//      or a recording <key> that does not fold, by check-verify-green's
//      vgFixSuffix, to `verify` or `verify-<s>`, <s> the worktree's basename
//      or its repository's name (kan-1067 self-review, 1-C: `verify-fix-1b`
//      formed its own group and verify stayed NOT-GREEN), or `-full` with
//      `-only`, with the `lint` list, or with no <worktree> and list
//   3  the list is not declared (project-get.sh exit 1): the caller
//      auto-detects, as **Inline verify** says
//   4  the store refused the recording begin (stderr names why): no command ran
//
// project-get, check-spec-reach and run-capped run in this same binary
// (runSibling, selfExe). The test
// list runs on rgtRun, a job pool, with one job per command so
// every one starts at once (decision verify-list-on-rgtrun).

package guard

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// The contract below moved here from the header of the deleted
// scripts/run-verify-list.sh shim (KAN-1058).
//
// flow-guard run-verify-list — run one of a project's verify lists the way **Inline
// verify** (skills/flow/verify-and-handoff.md) prescribes, and print its
// `## Report` lines.
//
// Usage: flow-guard run-verify-list <worktree> lint|test|visual [<change> <key> <session-token> <model> <effort> [<canonical-worktree>]]
//        flow-guard run-verify-list -only <file> <worktree> visual [<canonical-worktree>]
//        flow-guard run-verify-list -full <worktree> test|visual [<change> <key> <session-token> <model> <effort> [<canonical-worktree>]]
//
// Exit codes:
//   0  every command exited 0
//   1  at least one command exited non-zero, or a `visual` list refused a
//      `regression checkout` holding no worktree on the branch
//   2  usage (not `<worktree> lint|test|visual` plus none, all five, or all six recording
//      arguments, no flow binary on PATH for those, or <worktree> is not a
//      directory), a list line holding a `<word>` placeholder — its stderr
//      line names it — or flow-guard project-get exit 2 — its stderr line says why
//      -- or `-only` with another list, recording arguments, a missing or
//      empty file, or a `verify` cell with no `<specs>` (its stderr line
//      names the cause) -- or `-full` with `-only`, with the `lint` list, or
//      with no <worktree> and list
//   3  the list is not declared (flow-guard project-get exit 1): the caller
//      auto-detects, as **Inline verify** says
//   4  the store refused the recording begin (stderr names why): no command ran
//
// With the recording arguments the list ends the row it began with one of
// three outcomes: blocked with cause test-failure when a command failed,
// skipped when every command of a test list was skipped, and completed
// otherwise — a lint list always runs flow-guard check-spec-reach, so it never ends
// skipped. The begin carries `-diff-base <sha>`, the HEAD of <worktree> (the
// verified worktree, never <canonical-worktree>), which flow-guard check-verify-green
// judges the row by; a <worktree> git cannot resolve a HEAD in records none,
// and that row is never green (decision verify-row-records-head). Lint and
// test share the row: each group writes a `flow-verify-running-<key>-<token>-<list>`
// mark under `git rev-parse --git-path` once its begin is accepted and removes
// it at its end, and a group that would end completed or skipped while a
// sibling's mark stands records no end — the last group to finish ends the
// row, and a sibling killed before its end leaves it open for the parent to
// end by hand (kan-1086 self-review, F3).
//
// A `# suite: <name>` line in the `test` list names the one command below it
// as that suite: it runs under `flow suite record -C <worktree> -suite <name>
// --`, so its runtime reaches the store `flow suite list` reads; with no flow
// on PATH it runs bare. A plain run of the visual `verify` row runs under
// `flow suite record -C <worktree> -suite visual --` the same way; a
// narrowed run never does.
//
// A `test` list ends with the `## visual verification` `verify` row, run once
// after every other command, when the worktree's diff from its base branch
// touches a `ui paths` glob and decision.json does not record `visual.verify`
// `skipped`; a base that does not resolve runs it, and it fails the list like
// any other command (kan-1047). A declared `specs` row runs first, its paths
// put in `<specs>`: a non-zero exit or an unresolved base fails the list,
// and the whole suite never runs in its place, with no `specs` row either
// (THE SMOKE FALLBACK). A sibling
// worktree in the change's state record whose diff touches the `ui paths`
// as `<its repository name>/<path>` runs the row too, its path and kickoff
// base filling `<frontend-root>` and `<merge-base>` (siblingFrontend).
//
// A `visual` list runs that `verify` row alone, as a `test` list's tail would,
// and never the `## lint` or `## test` commands; when the conditions above
// skip it, it prints `- visual verify — not run: …` and exits 0 (kan-1052
// self-review, F2).
//
// A `test` list under a fix-run judge's verify fast (flow-guard fix-run-floor)
// skips every command and ends its row skipped.
//
// THE FULL SUITE (decisions full-is-the-list-unskipped,
// full-runs-the-visual-suite-whole, full-rows-carry-a-suffix,
// full-failure-names-its-escape). `-full`, a leading flag beside `-only`,
// runs the `test` or `visual` list with every skip off, only
// when the operator asks for it. Every command's environment gains
// `FLOW_VERIFY_FULL=1` after every inherited entry, so `flow-guard unchanged`
// answers 1 and each suite runs. Neither green file skips a command and A
// SKIPPED SUITE's check is not run, while both files are written as on any
// run. A judged verify fast prints
//   - fix-run verify fast set aside — -full runs the whole list
// skips nothing and leaves fix-run.json as it was. A suite is recorded as
// `<name>-full` and the visual row as `visual-full`. The visual `verify` row
// runs whenever the cell is declared, `<specs>` emptied and the `specs` row
// not run, whatever the diff, the green sha and decision.json say, from the
// `regression checkout` resolved as on any run. A failing command the green
// sha covers at this HEAD, or the green-commands file lists, gains
//   escaped the default list — it passed at <sha>
// in its block, after its output lines. `-full` with `-only`, with the
// `lint` list, or with no <worktree> and list exits
// 2 before any command runs.
//
// THE OPERATOR'S SWITCH. `-full` exits 2 before any command unless
// `FLOW_OPERATOR_FULL_SUITE=1` is in its environment. Only the operator sets
// that variable; an agent never does, and hands the operator the command
// instead. A run without `-full` puts `FLOW_VERIFY_FULL=` after every
// inherited entry of each command's environment, so no caller's export turns
// a default run into the full suite.
//
// THE SMOKE FALLBACK. Outside `-full` the visual `verify` row never runs with
// `<specs>` empty. With no `specs` row, a `specs` row with no merge base, one
// that exits non-zero or one that prints nothing, `<specs>` is filled with
// `--grep @smoke` -- the tests the regression suite tags `@smoke` -- and a
// stderr line names which of the four it was; the missing merge base and the
// failing `specs` row still fail the list. A `verify` cell with no `<specs>`
// can only run whole: it is not run and fails the list.
//
// A SKIPPED SUITE. A `test` command opening with `flow-guard unchanged
// <pathspecs> ||` (an optional path before `flow-guard`; each pathspec a bare
// word of letters, digits and `_ . / : ! @ % + = , -` or one single-quoted
// string holding no quote) has that guard run in process; on its exit 0 the
// command prints `- <command in backticks> — skipped, <the guard's line>`,
// runs no child, records no suite row, joins neither green file as passed and
// counts as skipped for the row's outcome. Any other shape, and any other
// exit, runs as before.
//
// The mechanism, and its full contract — the list, the green sha, the output
// — is stats/internal/guard/runverifylist.go; its test is
// stats/internal/guard/run_verify_list_test.go.

// keepResults copies every JUnit result file (`TEST-*.xml`) under wt written
// since the run started into dir, at its path relative to wt, and returns dir
// -- or "" when there was none. A failing suite's result files are what a
// flake is traced from, and the next run overwrites them (kan-1017's
// self-review: an 8-test failure left nothing to trace). dir is per run and
// never removed by a later one. .git and node_modules are not walked.
func keepResults(wt, dir string, since time.Time) string {
	kept := ""
	filepath.WalkDir(wt, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if n := d.Name(); n == ".git" || n == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasPrefix(d.Name(), "TEST-") || !strings.HasSuffix(d.Name(), ".xml") {
			return nil
		}
		if fi, err := d.Info(); err != nil || fi.ModTime().Before(since) {
			return nil
		}
		rel, _ := filepath.Rel(wt, p)
		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		dst := filepath.Join(dir, rel)
		if os.MkdirAll(filepath.Dir(dst), 0o777) == nil && os.WriteFile(dst, b, 0o666) == nil {
			kept = dir
		}
		return nil
	})
	return kept
}

func init() { Registry["run-verify-list"] = runVerifyList }

var (
	rvlFenceOpen = regexp.MustCompile("^```bash[[:space:]]*$")
	rvlInputs    = regexp.MustCompile(`^[[:space:]]*# inputs:(.*)$`)
	rvlSuite     = regexp.MustCompile(`^[[:space:]]*# suite:[[:space:]]*([A-Za-z0-9._-]+)[[:space:]]*$`)
	rvlTemplate  = regexp.MustCompile(`<[A-Za-z][A-Za-z0-9_-]*>`)
	// rvlGuarded is a `## test` line's leading `flow-guard unchanged
	// <pathspecs> ||` (A SKIPPED SUITE); rvlPathspec one of its pathspecs.
	rvlGuarded  = regexp.MustCompile(`^\s*(?:[A-Za-z0-9_./-]*/)?flow-guard\s+unchanged((?:\s+(?:[A-Za-z0-9_./:!@%+=,-]+|'[^']*'))+)\s*\|\|`)
	rvlPathspec = regexp.MustCompile(`[A-Za-z0-9_./:!@%+=,-]+|'[^']*'`)
)

// rvlOneCommand reports whether rest, what follows a guarded line's `||`, is
// one command the guard's exit 0 skips whole: a brace group closing at the
// line's end, or a command with no list operator. `a || b && c` still runs c
// on a's exit 0, so such a line runs as before.
func rvlOneCommand(rest string) bool {
	rest = strings.TrimSpace(rest)
	if strings.HasPrefix(rest, "{") {
		depth := 0
		for i, r := range rest {
			switch r {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					return i == len(rest)-1
				}
			}
		}
		return false
	}
	return !strings.ContainsAny(rest, ";&|\n")
}

// rvlGuardSkip runs a command's leading `flow-guard unchanged <pathspecs> ||`
// in process, from wt, and answers the guard's skip line when it exits 0 --
// "" for a command of any other shape or any other exit, which runs as
// before (decision skipped-suite-records-no-row).
func rvlGuardSkip(env Env, wt, cmd string) string {
	m := rvlGuarded.FindStringSubmatch(cmd)
	if m == nil || !rvlOneCommand(cmd[len(m[0]):]) {
		return ""
	}
	specs := rvlPathspec.FindAllString(m[1], -1)
	for i, p := range specs {
		specs[i] = strings.Trim(p, "'")
	}
	e := quietReasons(env)
	e.Dir = wt
	var out bytes.Buffer
	if Registry["unchanged"](specs, e, &out, io.Discard) != 0 {
		return ""
	}
	return strings.TrimSpace(out.String())
}

const (
	// rvlOperatorFull is THE OPERATOR'S SWITCH: the variable only the
	// operator sets. -full is refused without it.
	rvlOperatorFull = "FLOW_OPERATOR_FULL_SUITE"
	// rvlSmokeSet fills `<specs>` when no selection can: Playwright's own
	// filter for the tests tagged `@smoke` (THE SMOKE FALLBACK).
	rvlSmokeSet = "--grep @smoke"
	// rvlWideSelection is the most spec files a selection names and still
	// runs as it is (A WIDE SELECTION).
	rvlWideSelection = 40
)

// rvlNarrowing is what A WIDE SELECTION narrows with: tail, the `specs` row
// over an empty diff, and ownBase, the commit the diff of the checkout the
// row runs from is read from -- "" when no base of it is known.
type rvlNarrowing struct{ tail, ownBase string }

// rvlSpecFiles is the distinct spec files tokens name, in order.
func rvlSpecFiles(tokens []string) []string {
	var files []string
	for _, tok := range tokens {
		if p, _, _ := strings.Cut(tok, ":"); !slices.Contains(files, p) {
			files = append(files, p)
		}
	}
	return files
}

// rvlVisualVerify is the `## visual verification` `verify` row a `test` list
// ends with, the directory it runs from, and -- when the section declares a
// `specs` row -- that row with `<frontend-root>` and `<merge-base>`
// substituted, whose printed paths the caller puts in `<specs>`; with no
// `specs` row, specs is "" and the caller fills `<specs>` with the smoke set
// (THE SMOKE FALLBACK). ok false when the
// row is not declared, decision.json records `visual.verify` `skipped`, the
// green sha covers the tree, or the diff from the base branch touches no
// `ui paths` glob (verify-list-runs-visual). A base that does not resolve
// runs it: running more is the safe side -- but a `specs` row then has no
// merge base to scope to, and specs comes back "" with ok true and cmd ""
// so the caller fails the list and runs the smoke set, never the whole suite
// (verify-list-runs-targeted). fallback is the `regression checkout` dir
// fell back to when that checkout holds no worktree on wt's branch, "" otherwise.
func rvlVisualVerify(env Env, wt string, covered bool) (cmd, specs, dir, fallback string, ok bool) {
	cmd, specs, dir, fallback, _, ok = rvlVisualVerifyFor(env, wt, covered, false, "", "", func(string) string { return "" }, io.Discard)
	return
}

// rvlRound is a re-run key's round, `verify-fix-<k>` or `verify-<k>`, "" for
// a first verify's key.
func rvlRound(key string) string {
	m := vgFixSuffix.FindStringSubmatch(key)
	if m == nil {
		return ""
	}
	return strings.TrimSuffix(m[0], m[2])
}

// rvlRoundBase answers each root's round base for the run keyed key (THE
// ROUND'S SELECTION): "" for a first verify, or with no earlier row whose
// diffBase resolves in root. With no key -- a narrowed run -- it answers as
// for the newest verify row of wt.
func rvlRoundBase(env Env, wt, change, canon, key string) func(root string) string {
	none := func(string) string { return "" }
	if change == "" {
		return none
	}
	if canon == "" {
		canon = wt
	}
	raw, err := pfdRead(env, env.Dispatches, "dispatches", change, canon)
	var rows []struct{ Key, Role, DiffBase string }
	if err != nil || json.Unmarshal(raw, &rows) != nil {
		return none
	}
	rows = slices.DeleteFunc(rows, func(r struct{ Key, Role, DiffBase string }) bool {
		return r.Role != "verifier" || (r.Key != "verify" && !strings.HasPrefix(r.Key, "verify-"))
	})
	if key == "" {
		i := len(rows) - 1
		for ; i >= 0 && !verifyKeyOK(env, wt, rows[i].Key, "", ""); i-- {
		}
		if i < 0 {
			return none
		}
		key, rows = rows[i].Key, rows[:i]
	}
	tag := rvlRound(key)
	if tag == "" {
		return none
	}
	git := envGit(env)
	return func(root string) string {
		// Rows of this round -- a sibling's, begun beside this one at its
		// own new HEAD -- are never the previous verify.
		for i := len(rows) - 1; i >= 0; i-- {
			if r := rows[i]; r.DiffBase != "" && rvlRound(r.Key) != tag && git("-C", root, "cat-file", "-e", r.DiffBase+"^{commit}").Run() == nil {
				return r.DiffBase
			}
		}
		return ""
	}
}

// rvlKickoffBases is the worktrees map of `flow state get -C <canon>
// <change>`: each worktree's absolute path, in record order, and the commit
// `flow.kickoff` branched it from (decision selection-base-kickoff-commit)
// -- none when flow is not on PATH or fails.
func rvlKickoffBases(env Env, change, canon string) (worktrees []string, bases map[string]string) {
	bases = map[string]string{}
	flow, ok := lookPath(env, "flow")
	if !ok || change == "" {
		return nil, bases
	}
	get := exec.Command(flow, "state", "get", "-C", canon, change)
	get.Dir, get.Env = env.Dir, icChildEnv(env)
	out, err := get.Output()
	if err != nil {
		return nil, bases
	}
	st := icParseState(string(out))
	for i, k := range st.keys {
		if st.values[i] != "-" {
			worktrees, bases[k] = append(worktrees, k), st.values[i]
			if r, err := filepath.EvalSymlinks(k); err == nil && bases[r] == "" {
				bases[r] = st.values[i]
			}
		}
	}
	return worktrees, bases
}

// rvlSiblingStamp is `<path>=<HEAD>` of each worktree rvlKickoffBases returns
// for change (canon defaulting to wt), less wt itself, sorted and space-joined:
// "" when none, or when wt's project declares no visual `verify` row.
func rvlSiblingStamp(env Env, wt, change, canon string) string {
	if b, err := os.ReadFile(filepath.Join(wt, ".flow", "project.md")); err != nil || vvCells(string(b))["verify"] == "" {
		return ""
	}
	if canon == "" {
		canon = wt
	}
	worktrees, _ := rvlKickoffBases(env, change, canon)
	self, _ := filepath.EvalSymlinks(wt)
	var stamp []string
	for _, p := range worktrees {
		if r, _ := filepath.EvalSymlinks(p); p == wt || p == self || r == wt || r == self {
			continue
		}
		head, _ := capture(envGit(env)("-C", p, "rev-parse", "HEAD"))
		stamp = append(stamp, p+"="+head)
	}
	slices.Sort(stamp)
	return strings.Join(stamp, " ")
}

// rvlVisualVerifyFor is rvlVisualVerify with the change's name and its
// canonical worktree (wt when ""), whose state record gives the merge base:
// the kickoff base for wt, and only with none -- the fallback printed on
// stdout -- the merge base of HEAD and the remote base branch. The record's
// other worktrees are wt's siblings: the row also runs when a sibling's diff
// from its own base matches wt's `ui paths` under `--repo <its repository
// name>`, and the `specs` row then reads the first such sibling and its base
// (siblingFrontend, the rule `visual-verify-pass prepare` follows). full
// (THE FULL SUITE) runs the row whenever it is declared, `<specs>` emptied and
// no `specs` row, whatever covered, decision.json and the diff say. round
// answers a root's round base, which replaces the kickoff base in the `specs`
// row (THE ROUND'S SELECTION). nar is set beside a substituted `specs` row.
func rvlVisualVerifyFor(env Env, wt string, covered, full bool, change, canon string, round func(root string) string, stdout io.Writer) (cmd, specs, dir, fallback string, nar rvlNarrowing, ok bool) {
	if canon == "" {
		canon = wt
	}
	b, err := os.ReadFile(filepath.Join(wt, ".flow", "project.md"))
	if err != nil {
		return "", "", "", "", nar, false
	}
	cells := vvCells(string(b))
	if cells["verify"] == "" || (covered && !full) {
		return "", "", "", "", nar, false
	}
	// checkout resolves the `regression checkout` the row runs from.
	checkout := func() (dir, fallback string) {
		dir = wt
		if c := cells["regression checkout"]; c != "" {
			dir = c
			if !filepath.IsAbs(c) {
				dir = filepath.Join(wt, c)
			}
			// The cell names a main checkout; the change's worktree of it, on
			// wt's branch, is where the change's code and captures live.
			git := envGit(env)
			fallback = dir
			if branch, _ := capture(git("-C", wt, "branch", "--show-current")); branch != "" {
				if w := rvsWorktreeOn(git("-C", dir, "worktree", "list", "--porcelain"), branch); w != "" {
					dir, fallback = w, ""
				}
			}
		}
		return dir, fallback
	}
	if full {
		dir, fallback = checkout()
		return strings.ReplaceAll(cells["verify"], "<specs>", ""), "", dir, fallback, nar, true
	}
	var d struct {
		Visual struct {
			Verify string `json:"verify"`
		} `json:"visual"`
	}
	if raw, err := os.ReadFile(filepath.Join(wt, ".superpowers", "sdd", "decision.json")); err == nil && json.Unmarshal(raw, &d) == nil && d.Visual.Verify == "skipped" {
		return "", "", "", "", nar, false
	}
	git := envGit(env)
	worktrees, bases := rvlKickoffBases(env, change, canon)
	mb := ""
	changed := func() ([]string, error) {
		// The record may name wt by its symlink-resolved form (/private/var
		// for /var on macOS), or wt be that form.
		// rvlKickoffBases keys every record also by its resolved form.
		r := wt
		if e, err := filepath.EvalSymlinks(wt); err == nil {
			r = e
		}
		mb = bases[r]
		if mb == "" {
			var base bytes.Buffer
			if Registry["resolve-base-branch"]([]string{wt}, quietReasons(env), &base, io.Discard) != 0 {
				return nil, fmt.Errorf("no base branch")
			}
			ref := resolveRemoteBase(git, wt, strings.TrimSpace(base.String()))
			fmt.Fprintf(stdout, "- kickoff base: not recorded for %s — selecting from merge-base HEAD %s\n", wt, ref)
			m, ok := capture(git("-C", wt, "merge-base", "HEAD", ref))
			if !ok {
				return nil, fmt.Errorf("no merge base")
			}
			mb = strings.TrimSpace(m)
		}
		names, ok := capture(git("-C", wt, "diff", "--no-renames", "--name-only", mb+"..HEAD"))
		if !ok {
			return nil, fmt.Errorf("no diff")
		}
		return strings.Split(names, "\n"), nil
	}
	own := visualTrigger(quietReasons(env), wt, changed, io.Discard, io.Discard) != 1
	bases[wt] = mb
	siblings := slices.DeleteFunc(worktrees, func(w string) bool { return w == wt })
	root, base := siblingFrontend(wt, siblings, bases, func(s string) bool {
		names, ok := capture(git("-C", s, "diff", "--no-renames", "--name-only", bases[s]+"..HEAD"))
		common, ok2 := capture(git("-C", s, "rev-parse", "--path-format=absolute", "--git-common-dir"))
		if !ok || !ok2 {
			return false
		}
		e := quietReasons(env)
		e.Stdin = strings.NewReader(names + "\n")
		return Registry["check-visual-trigger"]([]string{"--repo", filepath.Base(filepath.Dir(common)), wt}, e, io.Discard, io.Discard) == 0
	})
	if !own && root == wt {
		return "", "", "", "", nar, false
	}
	dir, fallback = checkout()
	if s := cells["specs"]; s != "" {
		rb := round(root)
		if rb != "" {
			fmt.Fprintf(stdout, "- visual selection: the round's diff — %s since %s\n", root, vgShort(rb))
			base = rb
		}
		if base == "" {
			return "", "", dir, fallback, nar, true
		}
		// The row's own checkout is read from its round base on a re-run,
		// from its kickoff base otherwise.
		nar = rvlNarrowing{tail: strings.NewReplacer("<frontend-root>", root, "<merge-base>", "HEAD").Replace(s), ownBase: bases[dir]}
		if dir == root {
			nar.ownBase = base
		} else if rb != "" {
			if r := round(dir); r != "" {
				nar.ownBase = r
			}
		}
		return cells["verify"], strings.NewReplacer("<frontend-root>", root, "<merge-base>", base).Replace(s), dir, fallback, nar, true
	}
	return cells["verify"], "", dir, fallback, nar, true
}

var (
	rvlANSI         = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	rvlFailureBlock = regexp.MustCompile(`^\s+\d+ (failed|interrupted|flaky)$`)
	rvlFailureLine  = regexp.MustCompile(`^\s+\[[^\]]+\] › (\S+):(\d+):\d+ › `)
)

// rvlSummaryBlocks is every test a Playwright list-reporter log names in the
// summary blocks of the given kinds (generateSummaryMessage, Playwright
// 1.62.1): a `  N failed|interrupted|flaky` line opens a block, and each
// following `    [project] › path:line:col › titles` line adds `path:line`
// until a line that does not match. ANSI colour codes are stripped, and an
// entry is kept once, at its first appearance.
func rvlSummaryBlocks(log []byte, kinds ...string) []string {
	var out []string
	open := false
	for _, l := range strings.Split(rvlANSI.ReplaceAllString(string(log), ""), "\n") {
		if m := rvlFailureBlock.FindStringSubmatch(l); m != nil {
			open = slices.Contains(kinds, m[1])
			continue
		}
		m := rvlFailureLine.FindStringSubmatch(l)
		if !open || m == nil {
			open = false
			continue
		}
		if e := m[1] + ":" + m[2]; !slices.Contains(out, e) {
			out = append(out, e)
		}
	}
	return out
}

// rvlVisualFailures is the failed, interrupted and flaky tests of a log: the
// list `-only` re-runs (decision narrow-via-run-verify-list-flags).
func rvlVisualFailures(log []byte) []string {
	return rvlSummaryBlocks(log, "failed", "interrupted", "flaky")
}

// rvlKeepFailures writes the failures a failing visual run's log names to
// <keep>/visual-failing.txt, one per line -- an empty file when none parse,
// so `-only` never re-runs an earlier round's list -- and prints the file and
// the failures grouped as in the change's specs (their path is the path part
// of a specs token; every failure when specs is nil) or elsewhere (decision
// untouched-spec-outside-specs-output). A file it cannot write is printed as
// not written, with the error, and nothing else.
func rvlKeepFailures(keep string, log []byte, specs []string, stdout io.Writer) {
	file := filepath.Join(keep, "visual-failing.txt")
	failures := rvlVisualFailures(log)
	body := ""
	for _, f := range failures {
		body += f + "\n"
	}
	if err := os.WriteFile(file, []byte(body), 0o666); err != nil {
		fmt.Fprintf(stdout, "  failing list: not written — %v\n", err)
		return
	}
	if len(failures) == 0 {
		fmt.Fprintln(stdout, "  failing list: none parsed")
		return
	}
	fmt.Fprintf(stdout, "  failing list: %s\n", file)
	var mine, elsewhere []string
	for _, f := range failures {
		path, _, _ := strings.Cut(f, ":")
		in := specs == nil
		for _, s := range specs {
			if p, _, _ := strings.Cut(s, ":"); p == path {
				in = true
			}
		}
		if in {
			mine = append(mine, f)
		} else {
			elsewhere = append(elsewhere, f)
		}
	}
	group := func(name string, g []string) {
		fmt.Fprintf(stdout, "  failing %s (%d):%s\n", name, len(g), strings.TrimRight(" "+strings.Join(g, " "), " "))
	}
	group("in the change's specs", mine)
	group("elsewhere", elsewhere)
}

var rvlRoundFile = regexp.MustCompile(`^visual-(?:.+-r(\d+)\.log|results-r(\d+))$`)

// rvlNextRound claims the round one past the highest <N> among keep's
// `visual-*-r<N>.log` and `visual-results-r<N>` -- 1 when there is none -- by
// creating `visual-results-r<N>`, taking the next round when a concurrent run
// created it first.
func rvlNextRound(keep string) int {
	n := 0
	entries, _ := os.ReadDir(keep)
	for _, e := range entries {
		if m := rvlRoundFile.FindStringSubmatch(e.Name()); m != nil {
			r, _ := strconv.Atoi(m[1] + m[2])
			n = max(n, r)
		}
	}
	for n++; errors.Is(os.Mkdir(filepath.Join(keep, fmt.Sprintf("visual-results-r%d", n)), 0o777), fs.ErrExist); n++ {
	}
	return n
}

// rvlKeepPlaywrightResults copies the newest directory under
// <dir>/test-results/ modified since the run started -- the failing run's
// Playwright output, diffs and traces -- to dst, and returns dst, or "" when
// there was none. dst is the round's own and no later run removes it.
func rvlKeepPlaywrightResults(dir, dst string, since time.Time) string {
	root := filepath.Join(dir, "test-results")
	entries, _ := os.ReadDir(root)
	newest, at := "", since
	for _, e := range entries {
		if fi, err := e.Info(); err == nil && e.IsDir() && !fi.ModTime().Before(at) {
			newest, at = e.Name(), fi.ModTime()
		}
	}
	if newest == "" || os.CopyFS(dst, os.DirFS(filepath.Join(root, newest))) != nil {
		return ""
	}
	return dst
}

// rvlParse reads a verify list out of `project-get <worktree> <list>`'s
// output: every non-blank, non-`#` line inside a ```bash fence, in order.
// inputs[i] is the pathspec list the `# inputs:` line above cmds[i] named,
// suites[i] the `# suite:` line's name. fix-run-floor reads the `test`
// list's inputs through it too.
func rvlParse(body string) (cmds, inputs, suites []string) {
	pending, suite, fenced := "", "", false
	for _, line := range strings.Split(strings.TrimRight(body, "\n"), "\n") {
		switch {
		case rvlFenceOpen.MatchString(line):
			fenced = true
			continue
		case strings.HasPrefix(line, "```"):
			fenced = false
			continue
		}
		fields := strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '\t' })
		if !fenced || len(fields) == 0 {
			continue
		}
		if m := rvlInputs.FindStringSubmatch(line); m != nil {
			pending = m[1]
			continue
		}
		if m := rvlSuite.FindStringSubmatch(line); m != nil {
			suite = m[1]
			continue
		}
		if strings.HasPrefix(fields[0], "#") {
			continue
		}
		cmds = append(cmds, line)
		inputs = append(inputs, pending)
		suites = append(suites, suite)
		pending, suite = "", ""
	}
	return cmds, inputs, suites
}

func runVerifyList(args []string, env Env, stdout, stderr io.Writer) int {
	// A closed reader (`| head`) must not kill the list before its row's
	// `end`: writes return EPIPE instead (decision f6-ignore-sigpipe, kan-1037).
	// Notify, not Ignore: an ignored SIGPIPE is inherited by every list
	// command, where Go resets a handled one to the default in each child.
	sigpipe := make(chan os.Signal, 1)
	signal.Notify(sigpipe, syscall.SIGPIPE)
	defer signal.Stop(sigpipe)
	abs := func(p string) string {
		if env.Dir != "" && !filepath.IsAbs(p) {
			return filepath.Join(env.Dir, p)
		}
		return p
	}
	const usage = "usage: flow-guard run-verify-list <worktree> lint|test|visual [<change> <key> <session-token> <model> <effort> [<canonical-worktree>]]"
	const narrowUsage = "usage: flow-guard run-verify-list -only <file> <worktree> visual [<canonical-worktree>]"
	const fullUsage = "usage: flow-guard run-verify-list -full <worktree> test|visual [<change> <key> <session-token> <model> <effort> [<canonical-worktree>]]"
	// A narrowing flag re-runs the visual `verify` row alone, narrowed, and
	// records nothing: no row it ends can stand as the gate (decision
	// narrow-via-run-verify-list-flags).
	mode, only, full := "plain", []string(nil), false
	for len(args) > 0 && (args[0] == "-only" || args[0] == "-full") {
		if args[0] == "-full" {
			full, args = true, args[1:]
			continue
		}
		mode = "only"
		if len(args) < 2 {
			fmt.Fprintln(stderr, narrowUsage)
			Reason(env, "usage")
			return 2
		}
		b, err := os.ReadFile(abs(args[1]))
		if err != nil {
			fmt.Fprintf(stderr, "run-verify-list: -only: %v\n", err)
			Reason(env, "only-list-unreadable")
			return 2
		}
		if only = strings.Fields(string(b)); len(only) == 0 {
			fmt.Fprintf(stderr, "run-verify-list: -only %s is empty — nothing to re-run\n", args[1])
			Reason(env, "only-is-empty")
			return 2
		}
		args = args[2:]
	}
	if full && mode != "plain" {
		fmt.Fprintln(stderr, "run-verify-list: -full runs the whole list — not with -only")
		Reason(env, "full-with-only-refused")
		return 2
	}
	if full && (len(args) < 2 || (args[1] != "test" && args[1] != "visual")) {
		if len(args) >= 2 && args[1] == "lint" {
			fmt.Fprintln(stderr, "run-verify-list: -full takes the test or visual list, never lint")
		} else {
			fmt.Fprintln(stderr, fullUsage)
		}
		Reason(env, "full-needs-test-or-visual")
		return 2
	}
	// THE OPERATOR'S SWITCH: only the operator sets it, never an agent.
	operator := env.Getenv(rvlOperatorFull) == "1"
	if full && !operator {
		fmt.Fprintf(stderr, "run-verify-list: -full runs the whole suite, which only the operator starts — no command ran. The operator runs it with %s=1 set; an agent never sets that variable, it hands the operator the command instead\n", rvlOperatorFull)
		Reason(env, "full-needs-the-operator")
		return 2
	}
	// change and canon name the state record the kickoff base is read from.
	change, canon := "", ""
	if mode != "plain" {
		if (len(args) != 2 && len(args) != 3) || args[1] != "visual" {
			fmt.Fprintln(stderr, narrowUsage)
			Reason(env, "usage")
			return 2
		}
		if len(args) == 3 {
			canon = abs(args[2])
		}
		args = args[:2]
	}
	if (len(args) != 2 && len(args) != 7 && len(args) != 8) || (args[1] != "lint" && args[1] != "test" && args[1] != "visual") {
		fmt.Fprintln(stderr, usage)
		Reason(env, "usage")
		return 2
	}
	if fi, err := os.Stat(abs(args[0])); err != nil || !fi.IsDir() {
		fmt.Fprintln(stderr, usage)
		Reason(env, "usage")
		return 2
	}
	record := func(string, ...string) bool { return true }
	abortSuperseded := func() {}
	blockedMark := ""
	runKey := "" // the recording key, "" on an unrecorded run
	if len(args) >= 7 {
		flow, ok := lookPath(env, "flow")
		if !ok {
			fmt.Fprintln(stderr, "run-verify-list: no flow binary on PATH")
			Reason(env, "no-flow-binary-on-path")
			return 2
		}
		wt, verified := args[0], args[0]
		change = args[2]
		if len(args) == 8 {
			wt = args[7]
			canon = abs(wt)
		}
		key, token, model, effort := args[3], args[4], args[5], args[6]
		runKey = key
		if !verifyKeyOK(env, abs(args[0]), key, change, abs(wt)) {
			fmt.Fprintf(stderr, "run-verify-list: key %q does not fold to verify-<repository name>, or to bare verify when the change has one worktree (skills/flow/verify-and-handoff.md, **Recording.**): a change spanning several worktrees requires the -<repository name> suffix — no command ran\n", key)
			Reason(env, "key-does-not-fold")
			return 2
		}
		blockedMark = "flow-verify-blocked-" + key + "-" + token
		record = func(verb string, flags ...string) bool {
			argv := []string{"record", "dispatch", verb, "-C", wt, "-change", change}
			if verb == "begin" {
				argv = append(argv, "-role", "verifier", "-model", model, "-effort", effort, "-agent-id", "inline")
				// The verified worktree's HEAD, which check-verify-green judges
				// the row by (decision verify-row-records-head); a worktree git
				// cannot resolve records none, and that row is never green.
				if head, ok := capture(envGit(env)("-C", abs(verified), "rev-parse", "HEAD")); ok && head != "" {
					flags = append(flags, "-diff-base", head)
				}
			}
			argv = append(argv, "-key", key, "-session-token", token)
			argv = append(argv, flags...)
			rc, out := kwRun(env, "", flow, argv...)
			if rc == 0 {
				return true
			}
			if verb == "begin" {
				fmt.Fprintf(stderr, "run-verify-list: flow record dispatch begin refused — no command ran: %s\n", strings.TrimSpace(out))
			} else {
				fmt.Fprintf(stderr, "run-verify-list: flow record dispatch %s failed: %s\n", verb, strings.TrimSpace(out))
			}
			return false
		}
		// The open inline row of this token and group a killed run left
		// is ended aborted before this begin (f11-abort-superseded-row).
		abortSuperseded = func() {
			raw, err := pfdRead(env, env.Dispatches, "dispatches", change, wt)
			var rows []struct {
				Key          string `json:"key"`
				Role         string `json:"role"`
				AgentID      string `json:"agentId"`
				SessionToken string `json:"sessionToken"`
				Outcome      string `json:"outcome"`
			}
			if err != nil || json.Unmarshal(raw, &rows) != nil {
				return
			}
			group := vgFixSuffix.ReplaceAllString(key, "verify$2")
			for _, r := range rows {
				if r.Role != "verifier" || r.AgentID != "inline" || r.SessionToken != token || r.Outcome != "" ||
					r.Key == key || vgFixSuffix.ReplaceAllString(r.Key, "verify$2") != group {
					continue
				}
				if rc, out := kwRun(env, "", flow, "record", "dispatch", "end", "-C", wt, "-change", change, "-key", r.Key, "-session-token", token, "-outcome", "aborted"); rc != 0 {
					fmt.Fprintf(stderr, "run-verify-list: flow record dispatch end of superseded row %s failed: %s\n", r.Key, strings.TrimSpace(out))
				}
			}
		}
		args = args[:2]
	}
	wt, list := args[0], args[1]
	childEnv := icChildEnv(env)
	// The workspace's KEY=value lines, a variable the caller sets winning
	// (f16-run-prepares-workspace). A worktree on a branch that names no
	// change -- /flow-fast's and /flow-interactive's `<name>` -- carries no
	// workspace, and prepare-workspace would refuse it.
	branch, _ := capture(envGit(env)("-C", abs(wt), "symbolic-ref", "--quiet", "--short", "HEAD"))
	if _, change := pwChangeName(strings.TrimSpace(branch)); change || strings.TrimSpace(branch) == "" {
		prep := runSibling(env, "prepare-workspace", abs(wt))
		prep.Env, prep.Stderr = childEnv, stderr
		prepOut, prepErr := prep.Output()
		if prepErr != nil {
			fmt.Fprintf(stderr, "run-verify-list: flow-guard prepare-workspace failed on %s — no command ran:\n%s", wt, prepOut)
			Reason(env, "prepare-workspace-failed")
			return 2
		}
		for _, kv := range icLines(string(prepOut)) {
			if k, _, ok := strings.Cut(kv, "="); ok && icExportable.MatchString(kv) && env.Getenv(k) == "" {
				childEnv = append(childEnv, kv)
			}
		}
	}
	// rawVerify is the `verify` cell, its `<specs>` unfilled; a narrowed run
	// appends its flag's arguments after the paths it puts there.
	rawVerify, suffix := "", map[string]string{"only": " --retries=1"}[mode]
	if b, err := os.ReadFile(filepath.Join(abs(wt), ".flow", "project.md")); err == nil {
		rawVerify = vvCells(string(b))["verify"]
	}
	if mode != "plain" && rawVerify != "" && !strings.Contains(rawVerify, "<specs>") {
		fmt.Fprintf(stderr, "run-verify-list: the `verify` cell has no <specs> placeholder to narrow — no command ran: %s\n", rawVerify)
		Reason(env, "verify-cell-unnarrowable")
		return 2
	}
	fill := func(tokens []string) string {
		return strings.ReplaceAll(rawVerify, "<specs>", strings.TrimSpace(strings.Join(tokens, " ")+suffix))
	}

	// A visual list runs the `verify` row alone: no list to read.
	var body []byte
	if list != "visual" {
		get := runSibling(env, "project-get", wt, list)
		get.Env, get.Stderr = childEnv, stderr
		var err error
		body, err = get.Output()
		switch exitCode(err) {
		case 0:
		case 1:
			Reason(env, "no-list")
			return 3
		default:
			Reason(env, "project-get-failed")
			return 2
		}
	}

	cmds, inputs, suites := rvlParse(string(body))
	for _, c := range cmds {
		if rvlTemplate.MatchString(c) {
			fmt.Fprintf(stderr, "run-verify-list: %s line is a template, not a command — move it out of the fence: %s\n", list, c)
			Reason(env, "line-is-a-template")
			return 2
		}
	}

	git := envGit(env)
	gitOut := func(args ...string) (string, bool) {
		return capture(git(append([]string{"-C", wt}, args...)...))
	}
	clean := func() bool {
		st, ok := gitOut("status", "--porcelain", "--untracked-files=all")
		return ok && st == ""
	}
	headSHA, since, greenFile, sibRecorded := "", "", "", ""
	cmdsFile, greenAt := "", []string{} // THE GREEN COMMANDS: those passed at headSHA
	if list == "test" {
		if h, ok := gitOut("rev-parse", "-q", "--verify", "HEAD"); ok && clean() {
			if g, ok := gitOut("rev-parse", "--path-format=absolute", "--git-path", "flow-test-green"); ok {
				headSHA, greenFile = h, g
				b, _ := os.ReadFile(g)
				l := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
				since = l[0]
				if len(l) > 1 {
					sibRecorded = strings.TrimPrefix(l[1], "siblings ")
				}
			}
			if g, ok := gitOut("rev-parse", "--path-format=absolute", "--git-path", "flow-test-green-cmds"); ok && headSHA != "" {
				cmdsFile = g
				b, _ := os.ReadFile(g)
				if l := strings.Split(strings.TrimRight(string(b), "\n"), "\n"); l[0] == headSHA {
					greenAt = l[1:]
				}
			}
		}
	}
	passed := []string{} // commands that exited 0 this run, written to cmdsFile
	lintSHA, lintFile := "", ""
	if list == "lint" {
		if h, ok := gitOut("rev-parse", "-q", "--verify", "HEAD"); ok && clean() {
			if g, ok := gitOut("rev-parse", "--path-format=absolute", "--git-path", "flow-lint-green"); ok {
				lintSHA, lintFile = h, g
			}
		}
	}
	// unchanged reports whether a recorded green sha covers a command.
	unchanged := func(in string) bool {
		if since == "" || full {
			return false
		}
		if since == headSHA {
			return true
		}
		if strings.ReplaceAll(in, " ", "") == "" {
			return false
		}
		specs := strings.FieldsFunc(in, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' })
		diff := append([]string{"diff", "--quiet", since, headSHA, "--"}, specs...)
		return git(append([]string{"-C", wt}, append(diff, ".flow/project.md")...)...).Run() == nil
	}

	tmp := env.Getenv("TMPDIR")
	if tmp == "" {
		tmp = "/tmp"
	}
	outDir, err := os.MkdirTemp(tmp, "run-verify-list.")
	if err != nil {
		fmt.Fprintf(stderr, "run-verify-list: cannot create a temporary directory: %v\n", err)
		Reason(env, "cannot-create-a-temporary-directory")
		return 2
	}
	defer os.RemoveAll(outDir)
	keep, _ := gitOut("rev-parse", "--path-format=absolute", "--git-path", "flow-verify-failed")
	if keep != "" {
		os.MkdirAll(keep, 0o777)
		// A visual round's log is kept until the worktree goes (decision
		// per-round-artifacts-round-counter); only lint and test logs are
		// this run's alone.
		if list != "visual" {
			old, _ := filepath.Glob(filepath.Join(keep, list+"-*.log"))
			for _, f := range old {
				os.Remove(f)
			}
		}
	}

	started := time.Now().Add(-time.Second)
	failed := 0
	// report prints one command's block; name is its kept log's basename.
	report := func(shown string, rc int, file, name string) {
		if rc != 0 {
			failed = 1
		}
		fmt.Fprintf(stdout, "- `%s` — exit %d\n", shown, rc)
		out, _ := os.ReadFile(file)
		var lines []string
		if len(out) > 0 {
			lines = strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
		}
		if n := len(lines); n > 20 {
			fmt.Fprintf(stdout, "  (truncated — last 20 of %d lines)\n", n)
			lines = lines[n-20:]
		}
		for _, l := range lines {
			fmt.Fprintf(stdout, "  %s\n", l)
		}
		if bytes.Contains(out, []byte("WARNING: DATA RACE")) || bytes.Contains(out, []byte("race detected during execution of test")) {
			fmt.Fprintln(stdout, "  data race — never environmental")
		}
		if rc != 0 && keep != "" {
			log := filepath.Join(keep, name+".log")
			os.WriteFile(log, out, 0o666)
			fmt.Fprintf(stdout, "  full output: %s\n", log)
			if dir := keepResults(abs(wt), filepath.Join(keep, fmt.Sprintf("%s-results-%d", name, time.Now().UnixNano())), started); dir != "" {
				fmt.Fprintf(stdout, "  result files: %s\n", dir)
			}
		}
	}
	// Each command runs as `(cd "$WT" && bash -c …)` did: from the worktree,
	// PWD its logical path, so `pwd` prints the path the caller named.
	cmdEnv := append(childEnv[:len(childEnv):len(childEnv)], "PWD="+abs(wt))
	if full {
		cmdEnv = append(cmdEnv, "FLOW_VERIFY_FULL=1")
	} else {
		// A default run is the smoke set whatever its caller exported.
		cmdEnv = append(cmdEnv, "FLOW_VERIFY_FULL=")
	}
	// Every command runs under run-capped, which kills its process tree
	// past the project's `## resource cap` (kan-1044).
	self, err := selfExe()
	if err != nil {
		fmt.Fprintf(stderr, "run-verify-list: cannot resolve flow-guard's own executable: %v\n", err)
		Reason(env, "cannot-resolve-flow-guard")
		return 2
	}
	capped := []string{self, "run-capped", "--", "bash", "-c"}
	// run runs one command in the foreground.
	run := func(shown, cmd, key string) {
		results := rgtRun(context.Background(), [][]string{append(capped[:len(capped):len(capped)], cmd)}, 1, outDir, abs(wt), cmdEnv)
		report(shown, results[0].rc, results[0].out, list+"-"+key)
	}

	abortSuperseded()
	if !record("begin") {
		Reason(env, "record-failed")
		return 4
	}
	// This group's running mark, written once its begin is accepted and
	// removed at its end, so a sibling ending green leaves the row to the
	// last group running (kan-1086 self-review, F3).
	running := ""
	if blockedMark != "" {
		if p, ok := gitOut("rev-parse", "--path-format=absolute", "--git-path", "flow-verify-running-"+strings.TrimPrefix(blockedMark, "flow-verify-blocked-")); ok {
			running = p
			os.WriteFile(running+"-"+list, nil, 0o666)
		}
	}
	// VERIFY FAST: a fix run's judged verdict (fix-run-floor) skips every
	// `test` command while the change is still prose and the store holds no
	// Critical or Important finding of the verdict's round or later.
	fast, fastWhy := false, ""
	if list == "test" {
		if v, ok := fixRunVerdict(env, abs(wt)); ok && v.Verify == "fast" && full {
			fmt.Fprintln(stdout, "- fix-run verify fast set aside — -full runs the whole list")
		} else if ok && v.Verify == "fast" {
			if change == "" {
				branch, _ := gitOut("branch", "--show-current")
				change = strings.TrimPrefix(branch, "spectre/")
			}
			store := abs(wt)
			if canon != "" {
				store = canon
			}
			if fastWhy = frSevereSince(env, store, change, v.Round); fastWhy == "" {
				fast = true
			} else {
				fmt.Fprintf(stdout, "- fix-run verify fast voided — %s: the whole list runs\n", fastWhy)
			}
		}
	}
	suffixFull := "" // THE FULL SUITE: its rows are `<name>-full`
	if full {
		suffixFull = "-full"
	}
	// escaped answers the sha a command passed the default list at, "" when none.
	escaped := func(c string) string {
		switch {
		case !full:
			return ""
		case since != "" && since == headSHA:
			return since
		case slices.Contains(greenAt, c):
			return headSHA
		}
		return ""
	}
	skip := make([]bool, len(cmds))
	passedAt := make([]bool, len(cmds))
	guardSkip := make([]string, len(cmds)) // A SKIPPED SUITE: the guard's skip line
	var argvs [][]string
	for i, c := range cmds {
		switch {
		case fast:
			skip[i] = true
		case unchanged(inputs[i]):
			skip[i] = true
		case list == "test" && !full && slices.Contains(greenAt, c):
			skip[i], passedAt[i] = true, true
		case list == "test" && !full && func() bool { guardSkip[i] = rvlGuardSkip(env, abs(wt), c); return guardSkip[i] != "" }():
			skip[i] = true
		case list == "test":
			if flow, ok := lookPath(env, "flow"); ok && suites[i] != "" {
				argvs = append(argvs, append([]string{flow, "suite", "record", "-C", abs(wt), "-suite", suites[i] + suffixFull, "--"}, append(capped[:len(capped):len(capped)], c)...))
			} else {
				argvs = append(argvs, append(capped[:len(capped):len(capped)], c))
			}
		default:
			run(c, c, fmt.Sprint(i))
		}
	}
	if list == "test" {
		results := rgtRun(context.Background(), argvs, max(len(argvs), 1), outDir, abs(wt), cmdEnv)
		for i, c := range cmds {
			if fast {
				fmt.Fprintf(stdout, "- `%s` — skipped, fix-run verify fast: the change is prose\n", c)
				continue
			}
			if passedAt[i] {
				fmt.Fprintf(stdout, "- `%s` — skipped, passed at %s\n", c, headSHA)
				continue
			}
			if guardSkip[i] != "" {
				fmt.Fprintf(stdout, "- `%s` — skipped, %s\n", c, guardSkip[i])
				continue
			}
			if skip[i] {
				fmt.Fprintf(stdout, "- `%s` — skipped, no input changed since %s passed the list\n", c, since)
				continue
			}
			r := results[0]
			results = results[1:]
			report(c, r.rc, r.out, list+"-"+fmt.Sprint(i))
			if r.rc == 0 {
				passed = append(passed, c)
			} else if sha := escaped(c); sha != "" {
				fmt.Fprintf(stdout, "  escaped the default list — it passed at %s\n", sha)
			}
		}
	}
	visualRan, sibStamp := false, ""
	if list != "lint" && !fast {
		if change == "" {
			branch, _ := gitOut("branch", "--show-current")
			change = strings.TrimPrefix(branch, "spectre/")
		}
		sibStamp = rvlSiblingStamp(env, abs(wt), change, canon)
		sibSame := sibRecorded == sibStamp
		// rounded is set once a round base scopes the `specs` row (THE
		// ROUND'S SELECTION).
		rounded := false
		round := func(root string) string {
			rb := rvlRoundBase(env, abs(wt), change, canon, runKey)(root)
			rounded = rounded || rb != ""
			return rb
		}
		failingFile := filepath.Join(keep, "visual-failing.txt")
		if cmd, specs, dir, fallback, nar, ok := rvlVisualVerifyFor(env, abs(wt), unchanged("") && sibSame, full, change, canon, round, stdout); !ok && list == "visual" {
			fmt.Fprintln(stdout, "- visual verify — not run: no `verify` row, `visual.verify` skipped, or no `ui paths` change")
		} else if ok && list == "visual" && fallback != "" {
			// /flow-fast and /flow-interactive set up no regression-checkout
			// worktree; the main checkout holds a build that is not the
			// change's (kan-1052 self-review, R1).
			failed = 1
			fmt.Fprintf(stdout, "- visual verify — not run: `regression checkout` %s holds no worktree on this branch; create it, and start the stack the `verify` row needs from it, then re-run\n", fallback)
		} else if ok {
			visualRan = true
			var specTokens []string // the `specs` row's output, nil with no row
			// tokens fill `<specs>`; smokeWhy, when set, is why the smoke set
			// alone fills it. Outside -full the row never runs with `<specs>`
			// empty: the whole suite runs only when the operator asks
			// (verify-list-runs-targeted, THE SMOKE FALLBACK).
			var tokens []string
			smokeWhy := ""
			switch {
			case full:
			case !strings.Contains(rawVerify, "<specs>"):
				failed = 1
				cmd = ""
				fmt.Fprintf(stdout, "- visual verify — not run: the `verify` cell has no <specs> placeholder, so it can only run whole, and the whole suite runs only when the operator asks: %s\n", rawVerify)
			case len(only) > 0:
				// The `specs` row runs only to group the failures; the
				// list alone picks what runs.
				if cmd != "" && specs != "" {
					r := rgtRun(context.Background(), [][]string{{"bash", "-c", rvlStdoutOnly(specs, outDir)}}, 1, outDir, dir, cmdEnv)[0]
					if out, err := os.ReadFile(r.out); err == nil && r.rc == 0 {
						specTokens = append([]string{}, strings.Fields(string(out))...)
					}
				}
				tokens = only
			case cmd == "":
				failed = 1
				fmt.Fprintln(stdout, "- `specs` — not run, no merge base to scope the visual verify to; the whole suite runs only when the operator asks")
				smokeWhy = "the `specs` row has no merge base to scope to"
			case specs == "":
				smokeWhy = "no `specs` row is declared"
			default:
				r := rgtRun(context.Background(), [][]string{{"bash", "-c", rvlStdoutOnly(specs, outDir)}}, 1, outDir, dir, cmdEnv)[0]
				out, _ := os.ReadFile(r.out)
				paths := strings.Fields(string(out))
				switch {
				case r.rc != 0:
					if b, err := os.ReadFile(filepath.Join(outDir, "specs.stderr")); err == nil {
						if f, err := os.OpenFile(r.out, os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
							_, _ = f.Write(b)
							_ = f.Close()
						}
					}
					report(specs, r.rc, r.out, list+"-specs")
					smokeWhy = "the `specs` row failed"
				case len(paths) == 0:
					fmt.Fprintf(stdout, "- `%s` — exit 0, mapped no spec\n", specs)
					smokeWhy = "the `specs` row mapped no spec"
				default:
					tokens, specTokens = paths, paths
					if n := len(rvlSpecFiles(paths)); n > rvlWideSelection {
						// A WIDE SELECTION: the specs the change edits, then
						// the `specs` row's tail.
						var kept []string
						if names, ok := capture(envGit(env)("-C", dir, "diff", "--no-renames", "--relative", "--name-only", nar.ownBase+"..HEAD")); ok && nar.ownBase != "" {
							edited := strings.Split(names, "\n")
							for _, tok := range paths {
								if p, _, _ := strings.Cut(tok, ":"); slices.Contains(edited, p) {
									kept = append(kept, tok)
								}
							}
						}
						edited := len(rvlSpecFiles(kept))
						r := rgtRun(context.Background(), [][]string{{"bash", "-c", rvlStdoutOnly(nar.tail, outDir)}}, 1, outDir, dir, cmdEnv)[0]
						if out, err := os.ReadFile(r.out); err == nil && r.rc == 0 {
							for _, tok := range strings.Fields(string(out)) {
								if !slices.Contains(kept, tok) {
									kept = append(kept, tok)
								}
							}
						}
						fmt.Fprintf(stdout, "- visual selection: %d spec files reached, more than %d — narrowed to the %d the change edits plus the smoke tail; the rest runs only under -full, which the operator starts\n", n, rvlWideSelection, edited)
						if tokens = kept; len(kept) == 0 {
							smokeWhy = "the wide selection narrowed to no spec"
						}
					}
					if b, _ := os.ReadFile(failingFile); smokeWhy == "" && rounded && mode == "plain" && keep != "" {
						n := 0
						for _, f := range strings.Fields(string(b)) {
							if p, _, _ := strings.Cut(f, ":"); !slices.Contains(tokens, p) {
								tokens, n = append(tokens[:len(tokens):len(tokens)], p), n+1
							}
						}
						if n > 0 {
							fmt.Fprintf(stdout, "- visual selection: plus %d spec file(s) of visual-failing.txt\n", n)
						}
					}
				}
			}
			if smokeWhy != "" {
				tokens = strings.Fields(rvlSmokeSet)
				fmt.Fprintf(stderr, "run-verify-list: the visual verify runs the smoke set alone (`%s`) — %s; the whole suite runs only when the operator asks\n", rvlSmokeSet, smokeWhy)
			}
			switch {
			case full:
				cmd = fill(nil)
			case rawVerify != "" && !strings.Contains(rawVerify, "<specs>"):
			default:
				cmd = fill(tokens)
			}
			switch {
			case cmd == "":
			case list == "test" && !full && sibSame && slices.Contains(greenAt, cmd):
				fmt.Fprintf(stdout, "- `%s` — skipped, passed at %s\n", cmd, headSHA)
			default:
				// A plain run's runtime reaches the store as suite `visual`; a
				// narrowed one is not that suite (decision suite-records-scope).
				argv := []string{"bash", "-c", cmd}
				if flow, ok := lookPath(env, "flow"); ok && mode == "plain" {
					argv = append([]string{flow, "suite", "record", "-C", abs(wt), "-suite", "visual" + suffixFull, "--"}, argv...)
				}
				results := rgtRun(context.Background(), [][]string{argv}, 1, outDir, dir, cmdEnv)
				name, round := list+"-visual", 0
				if results[0].rc != 0 && keep != "" {
					round = rvlNextRound(keep)
					name = fmt.Sprintf("visual-%s-r%d", mode, round)
				}
				report(cmd, results[0].rc, results[0].out, name)
				if results[0].rc == 0 {
					passed = append(passed, cmd)
					if _, err := os.Stat(failingFile); err == nil && rounded && mode == "plain" && keep != "" {
						os.WriteFile(failingFile, nil, 0o666) // the run covered its entries
					}
				} else if keep != "" {
					dst := filepath.Join(keep, fmt.Sprintf("visual-results-r%d", round))
					if kept := rvlKeepPlaywrightResults(dir, dst, started); kept != "" {
						fmt.Fprintf(stdout, "  result files: %s\n", kept)
					} else {
						os.Remove(dst) // the empty claim
					}
					failLog, _ := os.ReadFile(results[0].out)
					rvlKeepFailures(keep, failLog, specTokens, stdout)
				}
				// A retry-pass is flaky, never red (THE NARROWED RUN).
				log, _ := os.ReadFile(results[0].out)
				if flaky := rvlSummaryBlocks(log, "flaky"); len(flaky) > 0 {
					fmt.Fprintf(stdout, "  flaky (%d): %s — passed on retry: fix each at its source\n", len(flaky), strings.Join(flaky, " "))
				}
			}
		}
	}
	// A fast run ran no command, so it stamps no green a later run could skip on.
	if headSHA != "" && failed == 0 && !fast && clean() {
		green := headSHA + "\n"
		if sibStamp != "" {
			green += "siblings " + sibStamp + "\n"
		}
		os.WriteFile(greenFile, []byte(green), 0o666)
	}
	if cmdsFile != "" && !fast && clean() {
		var lines []string
		for _, c := range append(greenAt, passed...) {
			if !slices.Contains(lines, c) {
				lines = append(lines, c)
			}
		}
		os.WriteFile(cmdsFile, []byte(strings.Join(append([]string{headSHA}, lines...), "\n")+"\n"), 0o666)
	}
	if list == "lint" {
		self, _ := selfExe()
		run("flow-guard check-spec-reach "+wt, `"`+self+`" check-spec-reach "`+wt+`"`, "spec-reach")
		if lintSHA != "" && failed == 0 && clean() {
			os.WriteFile(lintFile, []byte(lintSHA+"\n"), 0o666)
		}
	}
	// Lint and test share one dispatch key, run at once, and the later
	// group's end overwrites the row (skills/flow/verify-and-handoff.md,
	// Recording): each group marks its own failure, clears its own mark when
	// green, and ends blocked while any group's mark stands. The ends run
	// under one lock, held through the record, so whichever group ends last
	// sees the other's mark — a blocked group is never masked. A group that
	// would end the row completed or skipped while a sibling's running mark
	// stands records no end: the last group to finish ends it, and a sibling
	// killed before its end leaves the row open for the parent to end by
	// hand (kan-1086 self-review, F3).
	anyBlocked, siblingRunning := false, false
	if blockedMark != "" {
		if mark, ok := gitOut("rev-parse", "--path-format=absolute", "--git-path", blockedMark); ok {
			if lock, err := os.OpenFile(mark+".lock", os.O_CREATE|os.O_RDWR, 0o666); err == nil {
				defer lock.Close()
				syscall.Flock(int(lock.Fd()), syscall.LOCK_EX)
			}
			if failed != 0 {
				os.WriteFile(mark+"-"+list, nil, 0o666)
			} else {
				os.Remove(mark + "-" + list)
			}
			for _, l := range []string{"lint", "test", "visual"} {
				if _, err := os.Stat(mark + "-" + l); err == nil {
					anyBlocked = true
				}
			}
			if running != "" {
				os.Remove(running + "-" + list)
				for _, l := range []string{"lint", "test", "visual"} {
					if _, err := os.Stat(running + "-" + l); err == nil {
						siblingRunning = true
					}
				}
			}
		}
	}
	switch {
	case failed != 0 || anyBlocked:
		record("end", "-outcome", "blocked", "-cause", "test-failure")
	case siblingRunning:
		// A sibling group still runs: it ends the row.
	case list != "lint" && !visualRan && len(cmds) > 0 && !slices.Contains(skip, false):
		// A lint list always ran check-spec-reach.sh, so it never ends skipped.
		record("end", "-outcome", "skipped")
	default:
		record("end", "-outcome", "completed")
	}
	return failed
}

// verifyKeyOK: key folds, by check-verify-green's vgFixSuffix, to `verify` or
// `verify-<s>`, <s> wt's basename or its repository's name (the directory
// holding its git common dir); a git failure admits only the basename. Bare
// `verify` is refused when change's state record, read from canon, holds
// more than one worktree (f5-verify-key-suffix); change "" skips that read.
func verifyKeyOK(env Env, wt, key, change, canon string) bool {
	g := vgFixSuffix.ReplaceAllString(key, "verify$2")
	if g == "verify" {
		worktrees, _ := rvlKickoffBases(env, change, canon)
		return len(worktrees) <= 1
	}
	if g == "verify-"+filepath.Base(wt) {
		return true
	}
	out, ok := capture(envGit(env)("-C", wt, "rev-parse", "--path-format=absolute", "--git-common-dir"))
	return ok && g == "verify-"+filepath.Base(filepath.Dir(out))
}

// rgtJob is one harness's run: its exit code, its wall time and the file
// its merged stdout+stderr was captured into.
type rgtJob struct {
	rc, elapsed int
	out         string
}

// rgtRun runs each argv in dir — `bash
// -c <command>` for run-verify-list's test list — at most jobs at a time,
// starting them in list order, stdin /dev/null. Results are indexed by list position,
// so every later walk is in list order by construction. stdout and stderr
// share one capture file -- one descriptor, so the harness's own write order
// is kept -- and a file rather than a pipe, so a background process a
// harness leaves holding its output never delays the run.
func rgtRun(ctx context.Context, argvs [][]string, jobs int, tmpDir, dir string, childEnv []string) []rgtJob {
	results := make([]rgtJob, len(argvs))
	sem := make(chan struct{}, jobs)
	var wg sync.WaitGroup
	for i, argv := range argvs {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() { <-sem; wg.Done() }()
			r := &results[i]
			r.out = filepath.Join(tmpDir, strconv.Itoa(i)+".out")
			f, err := os.Create(r.out)
			if err != nil {
				r.rc = 126
				return
			}
			defer f.Close()
			cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
			cmd.Dir, cmd.Env, cmd.Stdout, cmd.Stderr = dir, childEnv, f, f
			// Per-harness wall time is whole Unix seconds end minus start,
			// what the script's `$(date +%s)` arithmetic computed -- never
			// a truncated time.Since, which differs whenever a run
			// straddles a second boundary.
			s := time.Now().Unix()
			r.rc = exitCode(cmd.Run())
			r.elapsed = int(time.Now().Unix() - s)
		}()
	}
	wg.Wait()
	return results
}

// rvlStdoutOnly wraps a `specs` command so its stderr -- the notes naming the
// files that select nothing -- goes to a file beside the run's output, never
// into the spec list it prints on stdout (kan-1099: rgtRun merges both streams,
// and a note became a Playwright filter, "No tests found").
func rvlStdoutOnly(specs, outDir string) string {
	q := "'" + strings.ReplaceAll(filepath.Join(outDir, "specs.stderr"), "'", `'\''`) + "'"
	return "{ " + specs + "\n} 2>" + q
}
