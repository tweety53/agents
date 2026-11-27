# Visual verification

The pipeline's one visual contract. A core phase file names only the project's visual check list,
`flow-guard run-verify-list <worktree> visual`, and the `ui paths` trigger, `flow-guard
check-visual-trigger`; every other visual rule a parent follows is here, or in the sibling this file
names for it — `skills/flow/fix-visual.md` for a visual fix, `skills/flow/sync-onto-base-visual.md`
for a recapture after a sync, `skills/flow/visual-verify-verifier.md` for the verifier's own steps.
Three sites load it, each for its own sections: `skills/flow/implement.md` for **During
`flow.implement`**, `skills/flow-interactive/SKILL.md` for **A request's visual list**, and
**Visual verification** (`skills/flow/verify-and-handoff.md`) for **The verifier dispatch** and
**Steps 3–13**.

**Visual verification** loads it only once
`flow-guard visual-verify-pass prepare` closes on a line other than `NEXT: skip` — at least one worktree in
this run's resolved set survived its step 2. That section carries the `prepare` call, which runs this
`flow.verify` mark's steps 1 and 2 (**Steps 1–2**, below); everything from step 3 on is
here. A run whose diff touches no UI path never reads this file.

## During `flow.implement`

**The implementer's `Visual close:` line.** An implementer stub (**Every implementer dispatch's
prompt is a stub**, `skills/flow/implement.md`) whose group has a task naming a `ui paths` file in
its `**Files:**` carries, in its `Visual close:` place, this line, each `<…>` filled:

```text
Visual close: <skill-dir>/fix-visual.md, **The implementer measures before it hands back** and **A fix for a visual departure runs its capture spec before it closes**, binds each task here whose commit reaches a capture spec carrying a `.mockups` sidecar — key `task-<id>`, base the commit the task started from; verifier file <skill-dir>/visual-verify-verifier.md; capture: <the `capture` command>; mockups: <the resolved `mockups` root>, frame <the `mockup frame` value>
```

**A view is verified against its mockup in the task that draws it**, so a
departure is fixed by the implementer that drew it, and `flow.verify` confirms the change
instead of finding it.

**The capture rules ride on a task's prompt.** Every task of **One stack for the verification
tasks** (`skills/flow/implement.md`), and every task whose `**Files:**` add or change a capture spec, carries the capture rules of step 8 of `skills/flow/visual-verify-verifier.md` by absolute path: a capture is taken in the state it names — every interaction producing that state first, each asserted, and the pointer parked clear of every control after the last click.

**The test list's visual tail.** `flow-guard run-verify-list`'s `test` list (**The parent runs the
test list**, `skills/flow/implement.md`) ends with that section's `verify` run on the specs its `specs`
row maps for the diff — the smoke set alone when it maps nothing or no `specs` row is declared, the
list failing when it exits non-zero, and the whole suite, `<specs>` empty, only when the operator asks for it
— when its own worktree's diff from its base branch touches a `ui paths` glob and
`decision.json` does not record `visual.verify` `skipped`, its header canonical for the rest, so a
capture a shared-component change moved fails here rather than at `flow.verify`. When only the
`--repo` run step 2 of **Steps 1–2** (below) adds exits 0,
for another worktree of the resolved set, its merge base from this run's working notes, the parent
runs that `verify` after the list, against the canonical worktree's stack, under the stack
rules of **One stack for the verification tasks** (`skills/flow/implement.md`).
The section's `setup` row runs before the list when the toolchain is missing. A failure a product defect caused
is fixed in the test list's fix round; a capture the change moved on purpose — one whose difference
the diff makes as design.md or the task's `**Files:**` intend — is re-recorded in it: its
committed PNGs are deleted and `capture` runs with `<spec>` substituted for its spec, writing them
as PNGs that do not yet exist, and they are committed by the regression checkout's rules, step 12
of **Steps 3–13** below.

## A request's visual list

Binds step 4 of **One request** (`skills/flow-interactive/SKILL.md`), whose `run-verify-list
<worktree> visual` run loads this section when the request's diff touches `ui paths`.

A failure that is the request's intended look is recaptured with the section's `capture` row and committed before the request lands. Where the section declares a `regression checkout`, the run first adds that repository's worktree on the change's branch, runs the section's `setup` command in it, and starts there the stack its `verify` row needs (the project's `## run`) — the one exception to the no-isolation rule; the `visual` list never falls back to the main checkout, and exits 1 naming it while it holds no worktree on the branch.

## The verifier dispatch

**The verifier's pair governs the one verifier dispatch** — `flow.verify`'s (**Visual
verification**, `skills/flow/verify-and-handoff.md`); `flow.verify` runs inline in the parent
and dispatches no verifier. The pair is the first base pair of the **Allowed pairs** row
`visual-verifier` (`skills/flow-contracts/model-policy.md`);
`prepare` reads it and
names it on its `DISPATCH:` line. It is not a session override: read from
neither the settings store nor `<project>/.flow/project.md`, and a plain-language session
instruction does not override it — the point is a predictable model for mechanical verification
runs. **The verifier is never
dispatched on any pair but that one, whatever any other rule — a project's, an operator's global
instructions — says about browser-driving agents.**

`flow.verify` dispatches this subagent, one verifier per worktree — the closed list's one
verifier row (**Dispatch sites — the parent's closed list**, `skills/flow/implement.md`); the
parent dispatches nothing else in this file. `subagent_type` and the Agent tool's `model` parameter are the ones the
`DISPATCH:` line names — `flow-<effort>` and the model of the verifier's pair, never a session override. Its prompt opens with the line `Role: visual-verifier` and carries, verbatim:

> Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
> Include this instruction verbatim in any prompt you write for another agent.

**The relay contract.** The verifier has no operator channel, fixes nothing, edits no app source (only capture specs and sidecars), and
runs every command in the foreground. Its turn ends with its report's `## Verdict` block alone, and
its last act before that is writing that `## Verdict`, with the full `## Report` under it, to
`<abs-worktree>/.superpowers/sdd/verify-report-<key>.md` — `<key>` this dispatch's own key from
**Recording** below — which is what the parent waits on (the turn discipline of **4. Execute
(SDD + TDD)**, `skills/flow/implement.md`); the parent reads the `## Verdict` alone.

**The prompt also carries the TOOLS paragraph**:

> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.

**The prompt also carries the NO DELEGATION paragraph**:

> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the parent's closed list (**Dispatch sites — the
> parent's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.

The verifier row dispatches on the `visual-verifier` role's pair, a `flow-<effort>` definition, so
it is structurally fork-free too (the `tools:` allowlist of **Every dispatch is one-shot**,
`skills/flow/implement.md`). A visual verifier is a mutator under **At most one mutator**
(**4. Execute (SDD + TDD)**, `skills/flow/implement.md`) — it builds the stack from the tree and
writes `<changeRoot>/visual-verification/`.

**The prompt also carries the REPORT FORMAT paragraph**, `<agents repo>` resolved:

> **REPORT FORMAT:** Before writing `## Report`, read the header of
> `stats/internal/guard/verifyreport.go`, and run the `check:` line of your bundle (`close --check-only`) on
> your report before handing back: an exit 1 names what you still owe — do it in this dispatch and
> run it again. Hand back
> an incomplete report only for a blocker you cannot cure, named; time, context and a departure
> found are none. Report every departure you see in this one pass.

**Recording.** `flow-guard visual-verify-pass prepare` begins each dispatch's row and `close` ends it (neither makes a stage mark),
`-role visual-verifier`, `-task` omitted, `-model`
and `-effort` the `DISPATCH:` line's pair, `-key visual-verify`, suffixed `-<repository name>` when this run's resolved set holds
more than one worktree — the worktree's main checkout's basename, the name
`flow-guard check-visual-trigger --repo` takes, since every worktree's own basename is the change name — the pair's semantics are section 4 of `skills/flow/implement.md`, cited
here, not restated.
Its `begin` carries `-diff-base` with the sha `git -C <worktree> rev-parse HEAD` prints before the dispatch, the HEAD it verifies.
A `-diff-base` found wrong after the begin is corrected on that row's `end`, which takes `-diff-base` too, never by back-filling a second row.
A verifier whose worktree's `ui paths` also match a sibling worktree's diff — a `<repo>/**` glob naming the sibling's repository — verifies that sibling's HEAD as well, so its `-diff-base` names each such HEAD too, comma-joined after its own: the guard checks each worktree against the one sha it resolves, and no row is recorded for a sibling no verifier was dispatched to.

**`close` checks every report for completeness before anything else reads it**, ends its row and
prints the next step — the parent's own Bash call, once the report file exists:

```bash
flow-guard visual-verify-pass close <canonical-worktree> <name> <key> <motions named>
```

`<motions named>` is the count of motions this dispatch's prompt named, `0` for `motions: none`.
`close` runs the report check with `--design <changeRoot>/design.md` and, when `mockups` is
declared, the frame-list check (`close --check-only` is those two alone); the headers of
`stats/internal/guard/verifyreport.go` and `framelist.go` are canonical for the report's `steps:` line, its
admissible statuses and the frame list, and `close`'s header for its outcomes. Act on its closing line:

- `NEXT: rest <key>` — **the report is incomplete and is never accepted, whatever its prose says**,
  and no fix task is planned from it, its departures included, until its remainder completes it. `close` closed the dispatch `-outcome aborted`; run `prepare --continue` for the key it
  names — `-rest`, then `-rest-<n>` from 2, before any `-<repository name>` suffix — and dispatch the
  remainder at once, its bundle carrying the incomplete report's path and the guards' verdict lines
  under `## Incomplete attempt`. A remainder completes its dispatch rather than adding one, so it
  counts toward no re-dispatch and no fix round. No count ends these either; a remainder that leaves
  the same steps undone on the same evidence takes **Fewest operator actions**
  (`skills/flow-contracts/pipeline.md`), as do three remainders of one dispatch that each close
  `NEXT: rest` (**Three failed rounds**, there), and one whose blocker only the operator can cure ends your
  turn with `## Question` naming that blocker and the undone steps.
- `NEXT: redispatch <key>` — **a complete report carrying any non-zero exit is re-dispatched, and no
  count ends the re-dispatches.** Three re-dispatches that each close `NEXT: redispatch` take
  **Three failed rounds** (`skills/flow-contracts/pipeline.md`). `close` closed the stage's first such report `-outcome blocked`;
  run `prepare --continue` for the key it names — `visual-verify-<n>`, `<n>` counting from 2, under
  the same `-<repository name>` suffix rule — and dispatch again, its bundle carrying every earlier
  report's path under `## Previous attempt`, so the verifier can tell an environmental failure (a
  build the worktree lacked, a flaky harness) from a defect in the branch.
- `NEXT: cure-environment <key>` — from the second report on, a non-zero exit whose block cause is
  `environment` is cured by your own Bash calls — the start, install or build its output names —
  before `prepare --continue` for the key it names and the next re-dispatch; end your turn with
  `## Question` naming the failing command and its output, verbatim, only when the cure needs a
  decision only the operator can make. Never run the failing command yourself to check it. A
  re-dispatch whose report repeats the previous one's failure on the same evidence takes **Fewest
  operator actions** (`skills/flow-contracts/pipeline.md`).
- `NEXT: loop` — from the second report on, a non-zero exit whose block cause is `test-failure` or
  `missing-fixture` takes the fix procedure (`skills/flow/fix.md`); `close` closed the row
  `blocked`.
- `NEXT: judge` — a complete report naming departures, or a `no sidecar line` frame-list gap:
  `close` printed the report's `## Verdict` and left the row and the stage open for **Blocking**
  below. What blocks takes the fix procedure, whose step 1 closes both; when nothing does, end the row
  with `flow record dispatch end … -outcome completed`, for every worktree of the set, then render the ledger with `flow record render -change <name> -kind ledger`.
- `NEXT: done` — every exit zero and no departure: `close` closed the row `-outcome completed`.

**An `ESCALATE:` line above `NEXT: loop` or `NEXT: judge` makes the next round a No-progress
escalation** (**Three failed rounds**, `skills/flow-contracts/pipeline.md`).

**A visual round whose report names more departures than the visual report that opened its round
is a round without progress too.** The visual stage's `close` prints `ESCALATE:` for it and for
every failed round from the third on (`flow-guard visual-verify-pass`'s header canonical), and the
next round is the escalation of **Fewest operator actions** (`skills/flow-contracts/pipeline.md`), never another dispatch on the base pair.

**Except after `rest`, `redispatch` and `cure-environment`, `close` has run step 12** and printed
`ROUND-BASE:`; and once the newest row of every surviving worktree is closed it has rendered the
ledger and printed `STAGE: stopped` when any is `blocked`, else `STAGE: completed`. It ends no stage mark: `flow.verify` is ended once, by **Verify** (`skills/flow/verify-and-handoff.md`). An exit 2 is a usage error or an unreadable input, corrected and `close` run again.

**A pass is one verifier per worktree, whatever its frame count.** Within a scoped pass a frame whose capture is byte-identical to one a report judged on a line of its own is carried forward, `flow-guard visual-verify-pass`'s `measure` step 5 canonical.

**An environment cure that changes what the stack serves — a served root switched, a worktree added
to the set — voids every earlier pass**, so the next re-dispatch runs unscoped: its `prepare` is
passed no `--round-base`.

**A final report that is a block closes its dispatch `-outcome blocked -cause <cause>`, never
`completed`.** The cause is one of the closed set the CLI validates, named from the report's own
facts: `environment` — a stack or tool the environment would not run (a stack that could not be
started, a build prerequisite the worktree lacked); `test-failure` — a lint/test command that
failed; `missing-fixture` — a fixture or baseline the verify needed and the worktree did not
carry. A report with every exit zero still
closes `-outcome completed`.


**A failing capture from a spec the change did not touch is first re-run on the base.** Failing there too, it is stale baseline drift, re-baselined in the same round and never a departure of this change.

A verifier that ends without writing its report file, or whose agent dies, is closed by `close`
`-outcome aborted` with `STOP: no report`, and blocks this handoff exactly as a failed command would,
naming the death.

## Steps 1–2

What `flow-guard visual-verify-pass prepare` runs before any other step of **Visual verification**
(`skills/flow/verify-and-handoff.md`), its header canonical:

1. **Resolve the section — with the guard, not by eye.** `prepare` pipes each worktree's diff into
   `flow-guard check-visual-trigger <worktree>`; exit 2 with a `VISUAL-TRIGGER-NOT-CONFIGURED:` line on stderr
   skips that worktree for the rest of this stage.
2. **Match the diff** — the same call's verdict. Exit 0 → at least one changed path matched a declared
   `ui paths` glob; continue. Exit 1, or exit 2 with a `VISUAL-TRIGGER-CANNOT-ANSWER:` line, skips the
   worktree.
   `flow-guard check-visual-trigger` owns the glob semantics (`**` spanning directories, a leading
   dot-slash prefix, an absolute glob, a glob with a space); nothing here restates them.

   In a run whose resolved set holds more than one worktree, every other worktree's diff is also
   matched against the canonical worktree's `ui paths` through `flow-guard check-visual-trigger --repo
   <repo-name>`, each path prefixed with that worktree's repository name — its main checkout's
   basename, the name a glob such as `gymie-frontend/**` spells another repository's files with — and
   its exit 0 makes the canonical worktree survive this step as the canonical worktree's own exit 0
   does.

   Each worktree's diff runs from its `--merge-base` — never `origin/<base>` or another ref picked by
   hand, which matches every path the base gained since the worktree was cut as this change's own.

**A `skipped` decision ends the stage at step 2.** When this run's decision
(`<abs-worktree>/.superpowers/sdd/decision.json`) records `visual.verify` `skipped` (step 6 of
**Plan check**, `skills/flow/brainstorm.md`), `prepare` closes the stage `skipped` and prints
`NEXT: skip — skipped at plan check — <reason>`, and every worktree surviving step 2 prints
`Visual: skipped at plan check — <reason>`.
`required`, `not configured` or no `visual` at all takes the load of **Visual verification** (`skills/flow/verify-and-handoff.md`), unchanged.

## Steps 3–13

Steps 1, 2, 3, 5, 6, 12 and 13 are the parent's — `flow-guard visual-verify-pass prepare` and `close`
run them, and the parent starts and stops the stack, all in the parent's own Bash calls, never a
subagent's. Steps 4 and 8–11 and the motion step are run by one verifier per worktree
surviving steps 1–3, dispatched per **The verifier dispatch** above under the key `prepare` began; the
parent applies **Blocking** to its report. `prepare` prints one `DISPATCH: <key> <bundle> — <description, subagent_type, model and Role: line>` line per
verifier, the Agent call taking those four as printed, then `NEXT: dispatch`; its bundle,
`<abs-worktree>/.superpowers/sdd/verify-prompt-<key>.md`, states the absolute worktree path; the
`KEY=value` lines `flow-guard prepare-workspace` printed for it; this section's resolved `setup`, `capture` and
`specs` commands and `screenshots` root, its resolved `mockups` root when
declared, and its `mockup frame` value when declared; the worktree-resolved URL of each app `ui paths`
matched; the project's `## run` commands; step 5's result, as the report's `stack:`
line states it — what step 5 found or started; the green inline verify row the report's `verify:` line names; `<changeRoot>`;
the relay contract above, with its report path resolved for this dispatch's `<key>`; the
`flow-guard visual-verify-pass measure` call of the verifier's step 9 with its arguments, its manifest key
this dispatch's key with any `-rest[-<n>]` segment removed, since a remainder completes the pass it
continues; the `check:` lines, the two guard calls `close` runs on the report; and
to read the absolute path of
`visual-verify-verifier.md` beside this file and run its steps 4 and 8–11 and its motion step as written against the stack step 5 left serving the
worktree's build, starting, stopping and restarting nothing, committing and pushing nothing. The parent's prompt carries the bundle's path and
the lines no script can resolve: the views touched; the change's motions, named as the paragraph below states; and the
`moved:` frames and, in a scoped fix round, the `defects:` frames — every frame this round's defects name — when the fix procedure (`skills/flow/fix.md`) names them. The parent passes the same `defects:` frames to `prepare` as `--defects`, beside `--round-base`. With `--defects`, `prepare`'s preflight renders and reach-checks only the sidecars holding those frames.

**`prepare` names the mockup sources in the bundle** — `mockup sources:`, then every HTML or CSS
path `design.md` cites that exists in the tree, or `mockup sources: none` — and decodes a bundled
page, one carrying `<script type="__bundler/template">`, into
`<abs-worktree>/.superpowers/sdd/verify-mockup-<key>/`, naming that directory beside it.

The verifier's `close --check-only` runs the report check with `--design <changeRoot>/design.md`, which refuses a
report, or the `visual-verification.md` beside `design.md`, reading `no mockup source` while
`design.md` cites a source that exists.
It also refuses step 10 reported n/a while `design.md` cites an HTML source that exists: an
approved prototype is rendered as each capture's frame (step 10,
`skills/flow/visual-verify-verifier.md`).

**The parent names the change's motions in the prompt** — `motions:`, then one `<motion id>:
<trigger> → <end state>` line per motion, or `motions: none`. A motion is every animation,
transition, entrance, exit, slide or gesture `design.md` or `proposal.md` names, and every one
the diff adds, changes or removes — a removed entrance included, since its frames are what prove
it is gone. A scoped re-run names the subset the fix procedure (`skills/flow/fix.md`)
states.

**A flagged diff that draws nothing is named in the prompt, with the wire contract to check.** When
every path step 2 matched is a web-layer file that draws nothing on screen — a DTO, a converter, a
filter, no component, style or layout — the parent adds `wire contract:` and one `<method> <path>:
<the payload fields and wire names the diff adds or changes>; <the change's stated acceptance
behaviour>` line per endpoint the diff serves, read from `proposal.md`, `design.md` and the diff.
The `ui paths` flag stands; the verification follows the diff's substance. A departure the
verifier reports on a `wire contract` line blocks as a defect in a captured screenshot does
(**Blocking** below).

3. **Pre-flight the workspace, before anything is dispatched** — `prepare`'s checks, run before it
   begins any verifier row. The verifier's one re-dispatch
   cannot repair an environment that cannot pass. **A `regression checkout` this run's resolved
   worktree set holds no worktree for gets one first.** When `## apps` declares the repository
   `## visual verification`'s `regression checkout` names and the set holds no worktree in it,
   `prepare` prints `STOP: kickoff <repo>` before it begins anything, and the parent runs that
   entry's recipe — the `flow-guard kickoff-worktree --from` call, then the regression
   checkout's `setup` (`skills/flow/cross-repo-worktrees.md`) — adds the worktree to the set, and runs
   `prepare` again. The recipe's merge-order record runs with
   it (**Isolate the workspace — additional worktrees**,
   `skills/flow/cross-repo-worktrees.md`): a repository the canonical `link.md`'s `## Merge order`
   omits is one run 1 never sequences and the self-review bundle never reads. **A change
   whose worktree records a flowBase gets a worktree in every `## apps` repository the set holds
   none in, the same way** — the same `STOP: kickoff <repo>`, the `flow-guard kickoff-worktree --from` call
   and its merge-order record — so the stack serves `<base>`'s build, never a main checkout on the
   default branch. `prepare`'s pre-flight check fails a
   worktree whose branch records a flowBase while an `## apps` repository holds no
   `<repo>-worktrees/<name>` worktree.

   `prepare` then runs the four checks — ports held only by their own app (an answering server running from another worktree of the repository fails), the base-URL
   configuration resolving to the workspace's values, origin allowed, one Playwright checkout per
   workspace — against each surviving worktree, given
   the pre-flight check's arguments: <worktree> and `<app-root>=<resolved-url> ...`
   with that worktree's `--app` arguments. A worktree whose `<project>/.flow/project.md` declares `preflight apps` needs no `--app`: `prepare` reads its pairs from that row.
   Each `<app-root>=<resolved-url>` is an app `ui paths` matched: its package root relative
   to the worktree and its worktree-resolved URL. That URL is read from the stack that is serving the build (the project's `## run` output, or the ports files it writes), never from `prepare-workspace`'s derived values: those name the ports the workspace is assigned, which are not always the ports the stack binds, and a verifier bundle naming an unreachable URL fails every probe. A root in another repository — a cross-repo
   change's sibling worktree — is passed relative to the worktree too
   (`../../<sibling-worktrees>/<name>`), never absolute, which the guard refuses as not a
   directory under it; the guard checks it as that repository's own worktree. The guard's header is canonical for
   each check's rule.

   Before them, once per stage, it runs the verify-green check on `<canonical-worktree> <name> <the stage's token>`
   with the set's other worktrees after it; `stats/internal/guard/verifygreen.go`'s header is canonical for the verdict. Its exit 1 →
   no verifier is dispatched: **Verify** has not closed green on this HEAD, so `prepare` ends the
   stage `stopped` with `STOP: verify-not-green`, the run returns to **Verify**, and visual
   verification starts again only once it has.

   When `mockups` is declared, it also runs the sidecar-reach check on `<capture
   root> <mockups root>`, `<capture root>` the `regression checkout`'s worktree when one is
   declared, else this worktree, and `<mockups root>` the resolved `mockups` root
   (`<project>/<mockups>`); its header is canonical for the rule. With `design repo` declared too,
   `flow-guard render-frames <worktree> <changeRoot>/design.md <sidecars>` runs first, so the tagged frames
   they cite exist before this check, compose and the frame-list check read them; its exit 1 or 2 is a
   failing check. With `specs` declared it runs on each existing `<spec>.mockups` of the specs that row prints, run from `<capture root>`, instead of the whole root, and a failing `specs` command is a failing check. Its exit 1 names each
   sidecar line whose screenshot no spec beside it takes, or whose frame the mockups root does
   not hold, a failing check. A guard's exit 2 — it cannot answer — is a failing check, carrying its stderr.
   **Of the sidecars reach-checked, a pass judges and captures only the change's own**: a sidecar
   counts when it or its spec changed in `<capture root>`'s diff from its `--round-base`, else its
   `--since`, else its `--merge-base`; a `--defects` pass keeps every sidecar holding a defect frame.
   Another change's sidecar-carrying spec the `specs` row reaches is neither captured nor judged — its
   committed baseline guards it through `verify` — and with no sidecar of the change's own, nothing is captured. The verifier's own capture run (step 8 of `skills/flow/visual-verify-verifier.md`) takes the bundle's `own specs:` line, the same sidecars' specs.

   **Any failing check ends the stage here**: no verifier is dispatched, `prepare` stops the pass —
   no stage mark is made — and its `STOP:` line names every failing check with the evidence above.
   The run proceeds to the `IN_PROGRESS` handoff, which names the environment cause — the operator
   fixes the environment and re-runs. This is a handoff, never a `## Question` and never a
   re-dispatch: the re-dispatch below is for a verifier report, not for an environment this stage
   has proven cannot pass.

5. **Probe before starting anything** — `prepare`, after step 3 and before the
   dispatch, as step 6 is. It probes the `--app` URL of each app `ui paths` matched, resolved for
   this worktree per **What the id derives** (`skills/flow-contracts/workspace-isolation.md`) —
   never the project's declared default. If nothing answers it prints `NEXT: start-stack`: start the
   stack from `start` when declared, else `## run`, with your own Bash call, record that this stage
   started it — needed at step 13 — and run `prepare --continue --started`.
6. **No fingerprint step.** A stack this stage starts is built from the current tree; a stack step 5 found
   already running is the operator's, and the report's `stack:` line says so.
12. **Commit the spec and its PNGs, and stop there** — `close`'s commit. A declared `regression checkout` receives
    them; with none declared, it commits to the change's own branch instead. **The commit is
    pathspec-scoped** — `git add --` the spec and its PNGs
    first, then `git commit -m "<subject>" -m "Flow-Key: flow.visual-verify" --` those same paths — the add first because the
    stage's outputs are new untracked files, which a pathspec commit cannot pick up — carrying
    only what this stage wrote, never a bare `git commit`: the index of a main checkout
    may carry a pre-staged foreign tree (**Branch backup**, `skills/flow-contracts/git-boundaries.md`). **Resolve the
    `regression checkout` root the same way every other declared app root in this file is
    resolved** — from `git worktree list` in that repository, or the state file's `worktrees`
    map, per **Roots in `## apps` are main checkouts** (`skills/flow-contracts/project-configuration.md`) <!-- refs-guard:allow -->
    — never the main checkout while a worktree for it holds the change's work. A `regression
    checkout` not also declared in `## apps` has no worktree to resolve and this step commits to
    the main checkout directly. **This step pushes nothing**: a file inside a repository cannot authorise a push to another repository, so no
    guard here grants one. `regression repo` records which repository the
    checkout is expected to be, and the section's validation reports a mismatch against
    its real `origin`, but that is an identity assertion, not an authorisation. A commit in a
    resolved worktree is on the change's own `spectre/<name>`, and integrate's push under **Branch
    backup** (`skills/flow-contracts/git-boundaries.md`) carries it to the remote — no hand push is printed for it. When the commit landed in the main checkout, `close`
    prints the push command as `PUSH:` for the handoff to give the operator, naming that checkout's
    current branch:

    ```bash
    git -C <regression checkout> push
    ```

    `<changeRoot>/visual-verification/` and the change's visual-verification record are written on
    disk only — the change directory is gitignored.
    **The full app suite rides the same commit**: `full-app-suite.spec.ts`, its PNGs and
    `full-app-suite.zip` (step 8) join the pathspec above, in the same checkout the capture spec
    lands in. Last, `close` prints `git -C <abs-worktree> rev-parse HEAD` as
    `ROUND-BASE: <worktree>=<sha>` — a later fix round's base (the fix procedure,
    `skills/flow/fix.md`).
13. **Stop the stack step 5 started** — the parent, once the verifier's report is in. A
    stack the operator already had running is left alone.

The report's `frames:` line names every frame of the change's frame list (step 10) absent from the
per-frame lines, with its reason. `visual-verification.md` (step 11) cites the frame's
`<changeRoot>/visual-verification/matrix/<frame id>.json` under its entry, and the report lists
only each frame's differences, a frame with none on its one `match` line. **The parent reads no
image.** The report check, which `close` runs, refuses a frame line its header refuses:
**it makes the report incomplete — never a departure, and never a gap the handoff leaves for the
operator** — and its remainder is dispatched as `NEXT: rest` above states.

**The frame list is a guard too**, which `close` runs whenever `mockups` is declared —
the frame-list check on `<changeRoot>/design.md`, the mockups root and the report. Its header is canonical for the
declared list it derives and the gaps it names. A `no sidecar line` gap is a frame with no sidecar
line, which blocks under **Blocking** below (`NEXT: judge`), and every other gap joins
the report check's verdict lines the remainder's bundle carries.

**Blocking.** This stage blocks the `IN_PROGRESS` handoff on: **a workspace pre-flight that failed
(step 3), which ends the stage before any dispatch**, a failed `setup`,
a genuine `capture` failure — the full app suite's included — **a frame this change added or
updated with no capture or no sidecar line, a full app suite or a `full-app-suite.zip` step 8 did
not write**, a stack that could not be started,
an unreadable PNG, **a `flow-guard compose-mockup-frames` exit 1 or 2 and a departure from the mockup the
verifier reports in a composite**, **a composed frame's `missing`, `extra` or over-tolerance band
the verifier judges a departure whose line names no cause — a pre-existing attribution is not one
(**the sweep**, `skills/flow-contracts/known-bugs.md`) — and a composed frame's `missing`,
`extra` or over-tolerance seam or off-centre line it judges a departure whose line names no cause**, and **a defect the
verifier reports in a captured screenshot — even when every assertion passed.** **A named motion
with no strip under `<changeRoot>/visual-verification/motion/` or no entry in
`visual-verification.md`, and a frame its strip shows breaking the motion step's rule, block as a
departure would.**

**A verified frame is frozen**. Once a verifier reports a frame clean, a later fix round may change it only inside the
regions its `defects:` name for that frame; a `changed:` frame whose capture differs from its
round base's — pixels or DOM — outside those regions is a regression, reported as a departure
naming the region and the fix whose diff moved it, even where the new capture still matches the
mockup.

**One decision never excuses several large frames.** `check-verify-report` refuses a report where two or more frames whose manifest `diff` is above 0.03 cite the same `(decision <ID>)` on their frame line; each such frame needs a reason of its own.

**A departure caused by missing account data is fixed by adding the data, never explained away.** When a
frame shows an element the render lacks because the seeded account lacks the data that shows it —
a grant, a template, a scheduled workout — the fix round adds that data where the captured account
comes from — the project's dev seed for a seeded account, the capture spec's own setup for one the
spec registers fresh — then re-captures. "Seed difference" is a cause only for
content the frame draws with different values (names, counts, dates), never for an element the
seed hides.

**A position departure of at most 1dp — 2px at 2x — on a band or an element may stand, and only
when all of these hold:** a fix round has already tried to close it at its source; the run records
why no easy fix exists, such as a font's glyph metrics; `visual-verification.md` carries it beside
the frame with its measured value and that reason; and the handoff's `**Decisions:**` lines name it.
`measure` applies the `tolerance` row itself (`skills/flow-contracts/project-configuration-visual.md`); `visual-verification.md` cites `<changeRoot>/visual-verification/within-tolerance.md` as its appendix.
It is never the first answer, and colour, a missing or extra element, and copy get no tolerance (the operator states this approximation as the requirement, the one widening `rules/fix-determinism-at-the-source.mdc` admits).
**`measure` folds a uniform offset before any model reads the `over` rows**: a frame whose paired rows all sit one dy from the mockup is one row, with the pair's `shift` carrying `uniform: true` — every other row is judged relative to that offset. Folding widens no tolerance. The offset row blocks like any other unless it is identical on every frame of the spec and `visual-verification.md` records its own cause beside it.

**Every blocking item but the pre-flight and a stack that could not be started
is the branch's own defect, never a `## Question`:** it takes the fix procedure
(`skills/flow/fix.md`). Those two are the environment's and end `flow.verify` as their
steps state.

**A step-3 pre-flight failure ends `flow.verify`
`-outcome stopped` instead of `completed`**, per those steps.
