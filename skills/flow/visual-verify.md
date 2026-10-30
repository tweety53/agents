# Visual verification — steps 3–13

Loaded by **Visual verification** (`skills/flow/verify-and-handoff.md`) only once
`visual-verify-pass.sh prepare` closes on a line other than `NEXT: skip` — at least one worktree in
this run's resolved set survived its step 2. That section carries the `prepare` call, which runs this
stage's `begin` mark and steps 1 and 2; everything from step 3 on, and the stage's `end` mark, is
here. A run whose diff touches no UI path never reads this file.

## The verifier dispatch

**`VERIFY_MODEL` governs the one verifier dispatch** — `flow.visual-verify`'s (**Visual
verification**, `skills/flow/verify-and-handoff.md`); `flow.verify` runs inline in the parent
and dispatches no verifier. `VERIFY_MODEL` is the fixed literal `opus`, dispatched on its
**Allowed pairs** row (`skills/flow-contracts/model-policy.md`), read from neither the settings store nor
`<project>/.flow/project.md`; a plain-language session instruction does not override it; and it
never falls back, because it is never resolved — the point is a predictable model for mechanical
verification runs regardless of what the decision chose for any other dispatch. **The verifier is
never dispatched on another model, whatever any other rule — a project's, an operator's global
instructions — says about browser-driving agents.**

`flow.visual-verify` dispatches this subagent, one verifier per worktree — the closed list's one
verifier row (**Dispatch sites — the parent's closed list**, `skills/flow/implement.md`); the
parent dispatches nothing else in this file but the tooling analyst of **A missed defect — the
tooling analysis** below. `subagent_type: flow-low` (`agents/flow-low.md`, effort `low`), the Agent tool's
`model` parameter set to `VERIFY_MODEL` (**Model resolution**, `skills/flow/SKILL.md`) — the
literal `opus`, never a decision pair and never a session override — mapped on harnesses `zcode`
and `muse` per
**Harness mapping** (`skills/flow-contracts/model-policy.md`), which the handshake below then
compares against. Its prompt opens with the line `Role: visual-verifier` and carries, verbatim:

> Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
> Include this instruction verbatim in any prompt you write for another agent.

**The relay contract.** The verifier has no operator channel, fixes nothing, edits no source, and
runs every command in the foreground. Its turn ends with its report's `## Verdict` block alone, and
its last act before that is writing that `## Verdict`, with the full `## Report` under it, to
`<abs-worktree>/.superpowers/sdd/verify-report-<key>.md` — `<key>` this dispatch's own key from
**Recording** below — which is what the parent waits on (the turn discipline of **4. Execute
(SDD + TDD)**, `skills/flow/implement.md`); the parent reads the `## Verdict` alone. The first line of its first reply is `Model: <the model named in its
own system prompt>`.

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

**The prompt also carries the MODEL HANDSHAKE paragraph**:

> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
> system prompt>` and nothing else on that line. Answer it before any tool call.

**The prompt also carries the REPORT FORMAT paragraph**, `<agents repo>` resolved:

> **REPORT FORMAT:** Before writing `## Report`, read the header of
> `<agents repo>/scripts/check-verify-report.sh`, and run it on your report before handing back:
> an exit 1 names the steps you still owe — do them in this dispatch and run it again. Hand back
> an incomplete report only for a blocker you cannot cure, named; time, context and a departure
> found are none. Report every departure you see in this one pass.

kan-1009: the paragraph priced a rejected report as one more dispatch, and a verifier handed back
with step 10 undone on that reading; its remainder ran 9m34s.

**Recording.** `visual-verify-pass.sh prepare` records each dispatch's `begin` and `close` its `end`,
`-role verifier`, `-task` omitted, `-model
opus -effort low`, `-key visual-verify`, suffixed `-<repository name>` when this run's resolved set holds
more than one worktree — the worktree's main checkout's basename, the name
`check-visual-trigger.sh --repo` takes, since every worktree's own basename is the change name — the pair's semantics are section 4 of `skills/flow/implement.md`, cited
here, not restated.
Its `begin` carries `-diff-base` with the sha `git -C <worktree> rev-parse HEAD` prints before the dispatch, the HEAD it verifies: `check-visual-verify-dispatched.sh` counts a verifier row only for that HEAD, and only while no UI path the change touches has changed since.
A `-diff-base` found wrong after the begin is corrected on that row's `end`, which takes `-diff-base` too, never by back-filling a second row (kan-1008: an agentless `-heads` row carried the correction).
A verifier whose worktree's `ui paths` also match a sibling worktree's diff — a `<repo>/**` glob naming the sibling's repository — verifies that sibling's HEAD as well, so its `-diff-base` names each such HEAD too, comma-joined after its own: the guard checks each worktree against the one sha it resolves, and no row is recorded for a sibling no verifier was dispatched to (kan-983's `visual-verify-2-gymie-frontend`).

**`close` checks every report for completeness before anything else reads it**, ends its row and
prints the next step — the parent's own Bash call, once the report file exists:

```bash
visual-verify-pass.sh close <canonical-worktree> <name> <key> <motions named> --session-token mf-<literal-token>
```

`<motions named>` is the count of motions this dispatch's prompt named, `0` for `motions: none`.
`close` runs `check-verify-report.sh` with `--design <changeRoot>/design.md` and, when `mockups` is
declared, `check-frame-list.sh`; their headers are canonical for the report's `steps:` line, its
admissible statuses and the frame list, and `close`'s header for its outcomes. Act on its closing line:

- `NEXT: rest <key>` — **the report is incomplete and is never accepted, whatever its prose says**,
  and no fix task is planned from it, its departures included, until its remainder completes it
  (KAN-1000). `close` closed the dispatch `-outcome aborted`; run `prepare --continue` for the key it
  names — `-rest`, then `-rest-<n>` from 2, before any `-<repository name>` suffix — and dispatch the
  remainder at once, its bundle carrying the incomplete report's path and the guards' verdict lines
  under `## Incomplete attempt`. A remainder completes its dispatch rather than adding one, so it
  counts toward no re-dispatch and no fix round. No count ends these either; a remainder that leaves
  the same steps undone on the same evidence takes **Fewest operator actions**
  (`skills/flow-contracts/pipeline.md`), and one whose blocker only the operator can cure ends your
  turn with `## Question` naming that blocker and the undone steps.
- `NEXT: redispatch <key>` — **a complete report carrying any non-zero exit is re-dispatched, and no
  count ends the re-dispatches.** `close` closed the stage's first such report `-outcome blocked`;
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
  `missing-fixture` takes **The loop** (`skills/flow/verify-fix-loop.md`); `close` closed the row
  `blocked`.
- `NEXT: judge` — a complete report naming departures, or a `no sidecar line` frame-list gap:
  `close` printed the report's `## Verdict` and left the row and the stage open for **Blocking**
  below. What blocks takes **The loop**, whose step 1 closes both; when nothing does, end the row
  with `flow record dispatch end … -outcome completed` and the stage, for every worktree of the set, with the `end` mark below, then render the ledger with `flow record render -change <name> -kind ledger`.
- `NEXT: done` — every exit zero and no departure: `close` closed the row `-outcome completed`.

**Except after `rest`, `redispatch` and `cure-environment`, `close` has run step 12** and printed
`ROUND-BASE:`; and once every surviving worktree's newest row of the stage is closed it has ended the
stage — `stopped` when any is `blocked`, else `completed` — and rendered the ledger, printing
`STAGE:`. An exit 2 is a usage error or an unreadable input, corrected and `close` run again.

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

**Handshake.** Compare the `Model:` line against `opus` (never a decision pair or a session
override) and apply **The handshake** (`skills/flow/implement.md`, **The parent orchestrates directly**),
unchanged — its first- and second-mismatch course and its single-model-harness case alike, `opus`
the requested model it names.

A verifier that ends without writing its report file, or whose agent dies, is closed by `close`
`-outcome aborted` with `STOP: no report`, and blocks this handoff exactly as a failed command would,
naming the death.

## A missed defect — the tooling analysis

**A miss is a defect the fix instructions report in a view an earlier round of this stage
passed** — that round's report, at the relay contract's path above
(`<abs-worktree>/.superpowers/sdd/verify-report-<key>.md`), exists and names no departure for it.
The parent classifies each defect the fix instructions name, from those reports, before the
verifier dispatch. On a fix run with at least one miss, **the parent dispatches one tooling
analyst per worktree with a miss, after step 3's pre-flight passes and before that worktree's
verifier**, and that verifier runs the sweeps it writes: `prepare --miss` stops after step 3 with
`NEXT: analyst`, and `prepare --continue` follows the analyst. A worktree with no miss dispatches
none. A scoped round (**The loop**, `skills/flow/verify-fix-loop.md`) passed only the motions
its prompt named.

**Load `skills/flow/visual-verify-tooling-analysis.md`** only when a worktree has at least one
miss — it carries the analyst's dispatch, its prompt and recording, the verifier's re-run with
the sweeps it writes, and the analyst's abort.

## Steps 3–13

Steps 1, 2, 3, 5, 6, 12 and 13 are the parent's — `visual-verify-pass.sh prepare` and `close`
run them, and the parent starts and stops the stack, all in the parent's own Bash calls, never a
subagent's. Steps 4 and 8–11 and the motion step are run by one verifier per worktree
surviving steps 1–3, dispatched per **The verifier dispatch** above under the key `prepare` began; the
parent applies **Blocking** to its report. `prepare` prints one `DISPATCH: <key> <bundle>` line per
verifier, then `NEXT: dispatch`; its bundle,
`<abs-worktree>/.superpowers/sdd/verify-prompt-<key>.md`, states the absolute worktree path; the
`KEY=value` lines `prepare-workspace.sh` printed for it; this section's resolved `setup`, `capture` and
`specs` commands and `screenshots` root, its resolved `mockups` root when
declared, and its `mockup frame` value when declared; the worktree-resolved URL of each app `ui paths`
matched; the project's `## run` commands; steps 5 and 6's results, as the report's `stack:` and
`fingerprint:` lines state them — what step 5 found or started, and step 6's exit or its
`not declared` reason; the green inline verify row the report's `verify:` line names; `<changeRoot>`;
the relay contract above, with its report path resolved for this dispatch's `<key>`; the
`visual-verify-pass.sh measure` call of the verifier's step 9 with its arguments, its manifest key
this dispatch's key with any `-rest[-<n>]` segment removed, since a remainder completes the pass it
continues; and
to read the absolute path of
`visual-verify-verifier.md` beside this file and run its steps 4 and 8–11 and its motion step as written against the stack steps 5 and 6 left serving the
worktree's build, starting, stopping and restarting nothing but the merge-base stack step 10's
attribution starts, committing and pushing nothing. The parent's prompt carries the bundle's path and
the lines no script can resolve: the views touched; `sweeps-<n>.md`'s path
when **A missed defect — the tooling analysis** above completed this round, with the re-run rule
its loaded file states, as written; the change's motions, named as the paragraph below states; and the
`moved:` frames when **The loop** (`skills/flow/verify-fix-loop.md`) names them.

**`prepare` names the mockup sources in the bundle** — `mockup sources:`, then every HTML or CSS
path `design.md` cites that exists in the tree, or `mockup sources: none` — and decodes a bundled
page, one carrying `<script type="__bundler/template">`, into <!-- citations-guard:allow -->
`<abs-worktree>/.superpowers/sdd/verify-mockup-<key>/`, naming that directory beside it.

The verifier runs `check-verify-report.sh` with `--design <changeRoot>/design.md`, which refuses a
report, or the `visual-verification.md` beside `design.md`, reading `no mockup source` while
`design.md` cites a source that exists (kan-1017: every verifier left 7 of 11 matrix columns n/a).
It also refuses step 10 reported n/a while `design.md` cites an HTML source that exists: an
approved prototype is rendered as each capture's frame (step 10,
`skills/flow/visual-verify-verifier.md`; kan-1038).

**The parent names the change's motions in the prompt** — `motions:`, then one `<motion id>:
<trigger> → <end state>` line per motion, or `motions: none`. A motion is every animation,
transition, entrance, exit, slide or gesture `design.md` or `proposal.md` names, and every one
the diff adds, changes or removes — a removed entrance included, since its frames are what prove
it is gone. A scoped re-run names the subset **The loop** (`skills/flow/verify-fix-loop.md`)
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
   entry's recipe — the `kickoff-worktree.sh --from` call, then the regression
   checkout's `setup` (`skills/flow/cross-repo-worktrees.md`) — adds the worktree to the set, and runs
   `prepare` again: kan-983's first verifier was dispatched without it and blocked on the
   missing checkout, ~4.45M tokens spent for nothing. The recipe's merge-order record runs with
   it, its planning commit included (**Isolate the workspace — additional worktrees**,
   `skills/flow/cross-repo-worktrees.md`): a repository the canonical `link.md`'s `## Merge order`
   omits is one run 1 never sequences and the self-review bundle never reads (kan-998). **A change
   whose worktree records a flowBase gets a worktree in every `## apps` repository the set holds
   none in, the same way** — the same `STOP: kickoff <repo>`, the `kickoff-worktree.sh --from` call
   and its merge-order record — so the stack serves `<base>`'s build, never a main checkout on the
   default branch: kan-1022's stack served the `develop` frontend to a change based on
   feature/templates, and 28 specs failed on it. `check-visual-preflight.sh` fails a
   worktree whose branch records a flowBase while an `## apps` repository holds no
   `<repo>-worktrees/<name>` worktree. <!-- citations-guard:allow -->

   `prepare` then runs the four checks — ports held only by their own app, the base-URL
   configuration resolving to the workspace's values, origin allowed, one Playwright checkout per
   workspace — against each surviving worktree, as
   `prepare-workspace.sh <worktree> | check-visual-preflight.sh <worktree> <app-root>=<resolved-url> ...`
   with that worktree's `--app` arguments.
   Each `<app-root>=<resolved-url>` is an app `ui paths` matched: its package root relative
   to the worktree and its worktree-resolved URL. A root in another repository — a cross-repo
   change's sibling worktree — is passed relative to the worktree too
   (`../../<sibling-worktrees>/<name>`), never absolute, which the guard refuses as not a
   directory under it; the guard checks it as that repository's own worktree. The guard's header is canonical for
   each check's rule.

   Before them, once per stage, it runs `check-verify-green.sh <canonical-worktree> <name> mf-<literal-token>`
   with the set's other worktrees after it; its header is canonical for the verdict. Its exit 1 →
   no verifier is dispatched: **Verify** has not closed green on this HEAD, so `prepare` ends the
   stage `stopped` with `STOP: verify-not-green`, the run returns to **Verify**, and visual
   verification starts again only once it has.

   When `mockups` is declared, it also runs `check-sidecar-reach.sh <capture
   root> <mockups root>`, `<capture root>` the `regression checkout`'s worktree when one is
   declared, else this worktree, and `<mockups root>` the resolved `mockups` root
   (`<project>/<mockups>`); its header is canonical for the rule. Its exit 1 names each
   sidecar line whose screenshot no spec beside it takes, or whose frame the mockups root does
   not hold, a failing check: kan-998's stale lines surfaced only at compose, after a 13-minute
   capture. A guard's exit 2 — it cannot answer — is a failing check, carrying its stderr.

   **Any failing check ends the stage here**: no verifier is dispatched, `prepare` ends the stage
   `-outcome stopped`, and its `STOP:` line names every failing check with the evidence above.
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
6. **Fingerprint the served bundle, if `fingerprint` is declared.** A screenshot is evidence only
   of what the app was serving when it was taken, and a stack step 5 found already running may be
   serving a build older than the worktree. `prepare` runs `check-dev-stack-fresh.sh <worktree>` — it reads the row and
   runs it, so the stage never parses the table by hand; its header is canonical for its three
   exits. Exit 0 → the served bundle is the worktree's build; continue. Exit 1 → `prepare` prints
   `NEXT: restart-stack`: stop the stack, start it from `start` when declared, else `## run` — never
   a restart command that drops the isolation flags `start` carries — record that this stage started
   it (step 13 stops it), and run `prepare --continue --started`. An exit 1 under `--started` blocks,
   carrying the guard's output, and ends the stage as a step-3 failure does: no verifier is
   dispatched.
   Exit 2 → the report reads `fingerprint: not declared — <the guard's stderr reason>` and the stage
   continues; the report makes the gap visible in every handoff, but this stage cannot prove what it
   was never told how to check.
12. **Commit the spec and its PNGs, and stop there** — `close`'s commit. A declared `regression checkout` receives
    them; with none declared, it commits to the change's own branch instead. **The commit is
    pathspec-scoped** — `git add --` the spec, its PNGs and `<changeRoot>/visual-verification/`
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
    checkout is expected to be, and `check-visual-verification.sh` reports a mismatch against
    its real `origin`, but that is an identity assertion, not an authorisation. A commit in a
    resolved worktree is on the change's own `spectre/<name>`, and integrate's push under **Branch
    backup** (`skills/flow-contracts/git-boundaries.md`) carries it to the remote — no hand push is printed for it (kan-1026: the operator was told to push a
    worktree integrate had already pushed). When the commit landed in the main checkout, `close`
    prints the push command as `PUSH:` for the handoff to give the operator, naming that checkout's
    current branch:

    ```bash
    git -C <regression checkout> push
    ```

    `<changeRoot>/visual-verification/` is committed with the change root.
    **The full app suite rides the same commit**: `full-app-suite.spec.ts`, its PNGs and
    `full-app-suite.zip` (step 8) join the pathspec above, in the same checkout the capture spec
    lands in. Last, `close` prints `git -C <abs-worktree> rev-parse HEAD` as
    `ROUND-BASE: <worktree>=<sha>` — a later fix round's base (**The loop**,
    `skills/flow/verify-fix-loop.md`).
13. **Stop the stack only if step 5 or step 6 started it** — the parent, once the verifier's report is in. A stack the operator already had running is left
    alone.

The report's `frames:` line counts against the change's frame list (step 10) and names every declared frame absent from the
per-frame lines, with its reason. The matrix rides `visual-verification.md` (step 11) under the
frame's entry, and the report's `matrix:` and `over-tolerance:` lines count its rows, its n/a cells
and the rows `measure` flagged. **The parent reads no image.** `check-verify-report.sh`, which
`close` runs, refuses a composed frame with no `bands:`, `seams:`, `matrix:` or `over-tolerance:`
line, an empty matrix cell, an n/a cell in a column that applies to its row's class and not reading
`no mockup source`, and a sweep its line reads neither `done` nor n/a — `<why it does not apply>`:
**each makes the report incomplete — never a departure, and never a gap the handoff leaves for the
operator** — and its remainder is dispatched as `NEXT: rest` above states.

**The frame list is a guard too**, which `close` runs whenever `mockups` is declared —
`check-frame-list.sh <changeRoot>/design.md <mockups root> <report>`. Its header is canonical for the
declared list it derives and the gaps it names. A `no sidecar line` gap is a frame with no sidecar
line, which blocks under **Blocking** below (`NEXT: judge`), and every other gap joins
`check-verify-report.sh`'s verdict lines the remainder's bundle carries.

**Blocking.** This stage blocks the `IN_PROGRESS` handoff on: **a workspace pre-flight that failed
(step 3), which ends the stage before any dispatch**, a failed `setup`,
a genuine `capture` failure — the full app suite's included — **a frame this change added or
updated with no capture or no sidecar line, a full app suite or a `full-app-suite.zip` step 8 did
not write**, a stack that could not be started, **a `fingerprint` that still exits
non-zero after step 6's restart**, a `check-spec-reach.sh` exit 1 or 2,
an unreadable PNG, **a `compose-mockup-frames.sh` exit 1 or 2 and a departure from the mockup the
verifier reports in a composite**, **a composed frame with a `missing`, `extra` or over-tolerance
band whose line names no cause — a pre-existing attribution is not one (**the sweep**,
`skills/flow-contracts/known-bugs.md`) — and a composed frame with a `missing`,
`extra` or over-tolerance seam or an off-centre line whose line names no cause**, **a captured view
no frame was composed for whose report carries no capture-scope `sweeps:` line, or one whose line
leaves a sweep neither `done` nor n/a — `<why it does not apply>`**, and **a defect the
verifier reports in a captured screenshot — even when every assertion passed.** **A named motion
with no strip under `<changeRoot>/visual-verification/motion/` or no entry in
`visual-verification.md`, and a frame its strip shows breaking the motion step's rule, block as a
departure would.**

**A departure caused by missing account data is fixed by adding the data, never explained away.** When a
frame shows an element the render lacks because the seeded account lacks the data that shows it —
a grant, a template, a scheduled workout — the fix round adds that data where the captured account
comes from — the project's dev seed for a seeded account, the capture spec's own setup for one the
spec registers fresh — then re-captures. "Seed difference" is a cause only for
content the frame draws with different values (names, counts, dates), never for an element the
seed hides (gymie KAN-1015: A1's Group session row, hidden because the seed had no grant).

**A position departure of at most 1dp — 2px at 2x — on a band or an element may stand, and only
when all of these hold:** a fix round has already tried to close it at its source; the run records
why no easy fix exists, such as a font's glyph metrics; `visual-verification.md` carries it beside
the frame with its measured value and that reason; and the handoff's `**Decisions:**` lines name it.
It is never the first answer, and colour, a missing or extra element, and copy get no tolerance
(KAN-1000: the operator states this approximation as the requirement, the one widening
`rules/fix-determinism-at-the-source.mdc` admits).

**The parent runs the `end` mark below itself only on `NEXT: judge` and `NEXT: compose-only`**;
on every other path `prepare` or `close` has already ended the stage, and a second end is never run.

**Every blocking item but the pre-flight, a stack that could not be started and the fingerprint
is the branch's own defect, never a `## Question`:** it takes **The loop**
(`skills/flow/verify-fix-loop.md`). Those three are the environment's and end the stage as their
steps state.

```bash
flow stage end -command '/flow' -stage flow.visual-verify -outcome completed <name>
```

**A step-3 pre-flight failure or a step-6 fingerprint still failing after its restart closes this
mark `-outcome stopped` instead of `completed`**, per those steps — the one outcome variant this
stage's end mark carries.
