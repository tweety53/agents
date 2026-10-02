# Visual verification — steps 3–13

Loaded by **Visual verification** (`skills/flow/verify-and-handoff.md`) only once at least one
worktree in this run's resolved set survives its step 2 — a changed path matched a declared `ui
paths` glob. That section carries this stage's `begin` mark and steps 1 and 2; everything from
step 3 on, and the stage's `end` mark, is here. A run whose diff touches no UI path never reads
this file.

## The verifier dispatch

**`VERIFY_MODEL` governs the one verifier dispatch** — `flow.visual-verify`'s (**Visual
verification**, `skills/flow/verify-and-handoff.md`); `flow.verify` runs inline in the parent
and dispatches no verifier. `VERIFY_MODEL` is the fixed literal `opus`, dispatched at effort
`low` through `subagent_type: flow-low`, read from neither the settings store nor
`<project>/.flow/project.md`; a plain-language session instruction does not override it; and it
never falls back, because it is never resolved — the point is a predictable model for mechanical
verification runs regardless of what the decision chose for any other dispatch.

`flow.visual-verify` dispatches this subagent, one verifier per worktree — the closed list's one
verifier row (**Dispatch sites — the parent's closed list**, `skills/flow/implement.md`); the
parent dispatches nothing else in this file but the tooling analyst of **A missed defect — the
tooling analysis** below. `subagent_type: flow-low` (`agents/flow-low.md`, effort `low`), the Agent tool's
`model` parameter set to `VERIFY_MODEL` (**Model resolution**, `skills/flow/SKILL.md`) — the
literal `opus`, never a decision pair and never a session override — mapped on harness `zcode` per
**Harness mapping** (`skills/flow-contracts/model-policy.md`), which the handshake below then
compares against. Its prompt carries, verbatim:

> Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
> Include this instruction verbatim in any prompt you write for another agent.

**The relay contract.** The verifier has no operator channel, fixes nothing, edits no source, and
runs every command in the foreground. Its turn ends with a single `## Report` block, and its last
act before that is writing the same block to
`<abs-worktree>/.superpowers/sdd/verify-report-<key>.md` — `<key>` this dispatch's own key from
**Recording** below — which is what the parent waits on (the turn discipline of **4. Execute
(SDD + TDD)**, `skills/flow/implement.md`). The first line of its first reply is `Model: <the model named in its
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

**Recording.** The parent records each dispatch, `-role verifier`, `-task` omitted, `-model
opus -effort low`, `-key visual-verify`, suffixed `-<worktree basename>` when this run's resolved set holds
more than one worktree — the pair's semantics are section 4 of `skills/flow/implement.md`, cited
here, not restated.

**A `## Report` carrying any non-zero exit is re-dispatched once.** The second dispatch is recorded
under `-key visual-verify-2` — the same `-<worktree basename>` suffix rule — with a prompt
identical to the first plus the first `## Report` verbatim under a `## Previous attempt` heading,
so the verifier can tell an environmental failure (a build the worktree lacked, a flaky harness)
from a defect in the branch. **The second report is final.** Another non-zero exit whose block
cause is `test-failure` or `missing-fixture` takes **The loop** (`skills/flow/verify-fix-loop.md`);
one whose cause is `environment` ends your turn with `## Question` naming the failing command and
its output, verbatim. Never run the failing command yourself to check it, and never dispatch a
third verifier on the same HEAD. The ledger render and this stage's `end` mark follow whichever
report was last.

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

A verifier that ends without a `## Report`, or whose agent dies, is closed `-outcome aborted`
and blocks this handoff exactly as a failed command would, naming the death.

## A missed defect — the tooling analysis

**A miss is a defect the fix instructions report in a view an earlier round of this stage
passed** — that round's report, at the relay contract's path above
(`<abs-worktree>/.superpowers/sdd/verify-report-<key>.md`), exists and names no departure for it.
The parent classifies each defect the fix instructions name, from those reports, before the
verifier dispatch. On a fix run with at least one miss, **the parent dispatches one tooling
analyst per worktree with a miss, after step 3's pre-flight passes and before that worktree's
verifier**, and that verifier runs the sweeps it writes. A worktree with no miss dispatches
none.

**Load `skills/flow/visual-verify-tooling-analysis.md`** only when a worktree has at least one
miss — it carries the analyst's dispatch, its prompt and recording, the verifier's re-run with
the sweeps it writes, and the analyst's abort.

## Steps 3–13

Steps 1, 2, 3, 5, 6, 12 and 13 are the parent's — those steps, `prepare-workspace.sh` and the ledger render
are the parent's own Bash calls, never a subagent's. Steps 4 and 7–11 are run by one verifier per worktree
surviving steps 1–3, dispatched per **The verifier dispatch** above with `-key visual-verify`; the
parent applies **Blocking** to its report. Its prompt states: the absolute worktree path; the
`KEY=value` lines **Verify** (`skills/flow/verify-and-handoff.md`) exported for it; this section's resolved `setup`, `verify`, `capture` and
`specs` commands and `screenshots` root, its resolved `mockups` root when
declared, and its `mockup frame` value when declared; the worktree-resolved URL of each app `ui paths`
matched; the project's `## run` commands; the views touched; `<changeRoot>`; the relay contract
above, with its report path resolved for this dispatch's `<key>`; `sweeps-<n>.md`'s path
when **A missed defect — the tooling analysis** above completed this round, with the re-run rule
its loaded file states, as written; and to read the absolute path of
`visual-verify-verifier.md` beside this file and run its steps 4 and 7–11 as written against the stack steps 5 and 6 left serving the
worktree's build, starting, stopping and restarting nothing, committing and pushing nothing.

3. **Pre-flight the workspace, before anything is dispatched.** The verifier's one re-dispatch
   cannot repair an environment that cannot pass. With the parent's own Bash call, before the
   dispatch below, run the four checks — ports held only by their own app, the base-URL
   configuration resolving to the workspace's values, origin allowed, one Playwright checkout per
   workspace — against this worktree:

   ```bash
   prepare-workspace.sh <worktree> | check-visual-preflight.sh <worktree> <app-root>=<resolved-url> ...
   ```

   One `<app-root>=<resolved-url>` argument per app `ui paths` matched: its package root relative
   to the worktree and its worktree-resolved URL. A root in another repository — a cross-repo
   change's sibling worktree — is passed as is; the guard checks it as that repository's own
   worktree. The guard's header is canonical for
   each check's rule. Exit 0 → every check passed. Exit 1 → each `FAIL:` line names a failing check
   and its evidence. Exit 2 → it cannot answer; treat it as a failing check, carrying its stderr.

   **Any failing check ends the stage here**: no verifier is dispatched, the stage's `end` mark
   carries `-outcome stopped`, and the report names every failing check with the evidence above.
   The run proceeds to the `IN_PROGRESS` handoff, which names the environment cause — the operator
   fixes the environment and re-runs. This is a handoff, never a `## Question` and never a
   re-dispatch: the re-dispatch below is for a verifier report, not for an environment this stage
   has proven cannot pass.

5. **Probe before starting anything** — with the parent's own Bash calls, after step 3 and before the
   dispatch, as step 6 is. Probe the URL of each app `ui paths` matched, resolved for
   this worktree per **What the id derives** (`skills/flow-contracts/workspace-isolation.md`) —
   never the project's declared default. If nothing answers, start the stack from `start` when
   declared, else `## run`, and record that this stage started it — needed at step 13.
6. **Fingerprint the served bundle, if `fingerprint` is declared.** A screenshot is evidence only
   of what the app was serving when it was taken, and a stack step 5 found already running may be
   serving a build older than the worktree. Run `check-dev-stack-fresh.sh <worktree>` — it reads the row and
   runs it, so the stage never parses the table by hand; its header is canonical for its three
   exits. Exit 0 → the served bundle is the worktree's build; continue. Exit 1 → stop the stack,
   start it from `start` when declared, else `## run` — never a restart command that drops the
   isolation flags `start` carries — record that this stage started it (step 13 stops it), and run
   the guard once more. A second exit 1 blocks, carrying the guard's output, and ends the stage as
   a step-3 failure does: no verifier is dispatched.
   Exit 2 → report `fingerprint: not declared — <the guard's stderr reason>` and continue; the
   report makes the gap visible in every handoff, but this stage cannot prove what it was never
   told how to check.
12. **Commit the spec and its PNGs, and stop there.** A declared `regression checkout` receives
    them; with none declared, commit to the change's own branch instead. **The commit is
    pathspec-scoped** — `git add --` the spec, its PNGs and `<changeRoot>/visual-verification/`
    first, then `git commit -m "<subject>" --` those same paths — the add first because the
    stage's outputs are new untracked files, which a pathspec commit cannot pick up — carrying
    only what this stage wrote, never a bare `git commit`: the index of a main checkout
    may carry a pre-staged foreign tree (**Branch backup**, `skills/flow-contracts/git-boundaries.md`). **Resolve the
    `regression checkout` root the same way every other declared app root in this file is
    resolved** — from `git worktree list` in that repository, or the state file's `worktrees`
    map, per **Roots in `## apps` are main checkouts** (`skills/flow-contracts/project-configuration.md`) <!-- refs-guard:allow -->
    — never the main checkout while a worktree for it holds the change's work. A `regression
    checkout` not also declared in `## apps` has no worktree to resolve and this step commits to
    the main checkout directly. **Never push**: a file inside a repository cannot authorise a push to another repository, so no
    guard here grants one, worktree or not. `regression repo` records which repository the
    checkout is expected to be, and `check-visual-verification.sh` reports a mismatch against
    its real `origin`, but that is an identity assertion, not an authorisation. When a commit landed
    in a `regression checkout`, print the push command for the operator to run by hand, naming
    whichever branch actually received the commit — the change's own `spectre/<name>` when a
    worktree resolved, the main checkout's current branch otherwise:

    ```bash
    git -C <regression checkout> push
    ```

    `<changeRoot>/visual-verification/` is committed with the change root.
    **The full app suite rides the same commit**: `full-app-suite.spec.ts`, its PNGs and
    `full-app-suite.zip` (step 8) join the pathspec above, in the same checkout the capture spec
    lands in.
13. **Stop the stack only if step 5 or step 6 started it** — the parent, once the verifier's report is in. A stack the operator already had running is left
    alone.

The report's `frames:` line counts against the change's frame list (step 10) and names every declared frame absent from the
per-frame lines, with its reason; the parent reconciles it against `design.md`'s own frame
list before applying **Blocking**, and a declared frame with no line blocks as a departure
would. The matrix rides `visual-verification.md` (step 11) under the frame's entry, and the
report's `matrix:` line counts its rows and its n/a cells; the parent opens the frame beside the
matrix and a visible element with no row, or an empty cell, blocks as a departure would.

**Blocking.** This stage blocks the `IN_PROGRESS` handoff on: **a workspace pre-flight that failed
(step 3), which ends the stage before any dispatch**, a failed `setup`, a failed `verify`,
a genuine `capture` failure — the full app suite's included — **a frame this change added or
updated with no capture or no sidecar line, a full app suite or a `full-app-suite.zip` step 8 did
not write**, a stack that could not be started, **a `fingerprint` that still exits
non-zero after step 6's restart**, a `check-spec-reach.sh` exit 1 or 2,
an unreadable PNG, **a `compose-mockup-frames.sh` exit 1 or 2 and a departure from the mockup the
verifier reports in a composite**, **a frame on the change's declared list with no line in the
report, a composed frame with no `bands:` line or with a `missing`, `extra` or over-tolerance
band whose line names no cause, a composed frame with no `seams:` line or with a `missing`,
`extra` or over-tolerance seam or an off-centre line whose line names no cause, and a composed frame with no `matrix:` line, a matrix row missing for an element the
frame visibly draws, or a cell that is neither the script's numbers nor an n/a with its reason
— the parent's own reconciliation, step 10**, and **a defect the
verifier reports in a captured screenshot — even when every assertion passed.**

**Every blocking item but the pre-flight, a stack that could not be started and the fingerprint
is the branch's own defect, never a `## Question`:** it takes **The loop**
(`skills/flow/verify-fix-loop.md`). Those three are the environment's and end the stage as their
steps state.

```bash
flow stage end -command '/flow' -stage flow.visual-verify -outcome completed <name>
```

**A step-3 pre-flight failure, a step-6 fingerprint still failing after its restart, or the
in-run fix loop's cap closes this mark `-outcome stopped` instead of `completed`**, per those
steps and **The cap** (`skills/flow/verify-fix-loop.md`) — the one outcome variant this
stage's end mark carries.
