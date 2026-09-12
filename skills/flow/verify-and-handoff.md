# Verify, stage, and hand off

Loaded by `skills/flow/SKILL.md` once `skills/flow/review-panel.md` closes clean. Carries the stage
order design.md's `workspace-export-lint-merge` and `run-instructions-reorder` decisions produce:
**verify → visual-verify → stage-diff → run-instructions → write-in-progress** — the old
`do.run-instructions → do.workspace-export → do.lint-and-test → do.stage-diff → do.write-in-progress`
order, with the middle two merged into one `flow.verify` stage and `run-instructions` moved to
immediately before the state write. `flow.visual-verify` sits between `flow.verify` and
`flow.stage-diff` — a later insertion, not part of either decision above. `flow.verify` runs its
commands inline, in the parent's own Bash calls; `flow.visual-verify` alone dispatches a
`verifier` subagent (**The verifier dispatch**, below); the parent keeps every mark and every
block decision.

## Verify

**Load `skills/flow-contracts/worktree-resolution.md`** before resolving this run's worktree set,
below.

```bash
flow stage begin -command '/flow' \
  -stage flow.verify \
  -harness <harness> \
  -session-token mf-<literal-token> \
  <name>
```

This stage is design.md's `workspace-export-lint-merge`: the old `do.workspace-export` and
`do.lint-and-test` stages, unconditional automated pass/block checks with nothing interactive
between them, merged into one mark.

**First, validate the section and export what it declares — with the script, not by eye.** Run

```bash
prepare-workspace.sh <worktree>
```

once per worktree in this run's resolved set — the same set **2. Isolate the workspace**
(`skills/flow/implement.md`) resolved, non-empty by construction — never a raw read of the state
file's `worktrees` map. Per **Resolving a change's worktrees**
(`skills/flow-contracts/worktree-resolution.md`), report an empty resolved set and do not proceed.

`prepare-workspace.sh` runs `check-workspace-isolation.sh` against the worktree first, then — only
if that passes — derives and exports the variables the project's `## workspace isolation` section
declares, resolved against the workspace id, and prints one `KEY=value` line per exported variable
to stdout. Exit 0 means the printed lines are what to carry forward into `## lint` and `## test`
below (nothing printed means the project declares no `## workspace isolation` section). A non-zero
exit is the dropped-row case (exit 1, relay the script's own lines verbatim and stop) or the
cannot-answer case (exit 2, stop the same way) — stop **before** `## lint` and `## test`, without
writing the state file.

**Load `skills/flow-contracts/workspace-isolation.md` only when `prepare-workspace.sh` exited
non-zero, or exited 0 with stderr naming a `cache index` row** — the procedures for both live
there; the ordinary exit-0 run loads nothing.

**A declared `cache index` row is never among the printed `KEY=value` lines** — the script reports
it by name on stderr instead. On an exit-0 run whose stderr names a `cache index` row, probe the
project's cache here, claim a free index atomically, and record that claim in the cache itself under
an entry naming this workspace, per **The cache index**
(`skills/flow-contracts/workspace-isolation.md`).

**When the script cannot be located**, apply the same rules by hand from **Project configuration** (`skills/flow-contracts/project-configuration.md`)
and **Workspace isolation** (`skills/flow-contracts/workspace-isolation.md`), and say in the handoff that the validation and
export were performed manually and why.

**This step does not call the project's `create` command.** `create` is called by whatever starts
the project's applications, per **Project configuration**
(`skills/flow-contracts/project-configuration.md`), and this step starts none of them — it
exports, lints, tests, and hands off.

**After the panel closes, the parent edits no source.** Any source change from here on makes
every slot's result stale (**Panel re-runs**, `skills/flow/review-panel.md`), and the only path
that changes source is a fix run the operator starts. `## lint`, `## test` and
`check-spec-reach.sh <worktree>` **are the parent's own Bash calls, run inline per worktree,
never through a subagent** — its work in this stage is `prepare-workspace.sh`, those commands, the
visual-verify dispatch below and the ledger render. A failing check is never "just re-run to see":
the run below gives it exactly one inline re-run, per **Inline verify — a failing command** below.

### Inline verify

Resolve the commands `project-get.sh <worktree> lint` and `project-get.sh <worktree> test` print
(auto-detect on exit 1). **The parent itself runs them, per worktree — never a subagent.**
Export the `KEY=value` lines `prepare-workspace.sh` printed for that worktree, then run the lint
commands, then the test commands, in the order printed, then `check-spec-reach.sh <worktree>` —
one more command in the same list, whose exit 0 line `Spec reach: not configured` is the ordinary
case for a project with no `regression checkout` (its header is canonical for its exit codes). Run
every command in order and do not stop at the first failure. **Nothing runs them later** —
`/flow`'s integrate phase has no verification gate — so a non-zero exit blocks this handoff. Per
**Read discipline**'s test/lint-through-`tail` rule (`skills/flow/implement.md`), pipe each
command's own run through `tail` before reading it; the `## Report` block below still carries the
full truncated-tail text for the operator, which is a different concern from what the parent reads
mid-run.

```text verified:design.md section 2 of this change
## Report
- `<command>` — exit <n>
  <the command's output, verbatim, or its last 40 lines when longer, stated as truncated>
```

The parent writes this `## Report` itself and shows it as this stage's output.
`prepare-workspace.sh`, both `project-get.sh` calls and this stage's `begin` mark are one Bash
call; the lint and test run, this run's `flow record dispatch begin`, and the ledger render below
are one more.

**Inline verify — a failing command.** A non-zero exit from any command in the list earns **one**
inline re-run of that command — the environmental-flake case. A second non-zero exit from the same
command ends the turn with `## Question` naming the command and its output, verbatim; the operator
resolves it through a fix run. Never treat a passing re-run as license to skip the rest of the
list — every remaining command in the order above still runs.

**Recording.** One `dispatches` row per worktree, `-role verifier -key verify -model <parent
model> -effort <parent effort> -agent-id inline`, suffixed `-<worktree basename>` when this
run's resolved set holds more than one worktree — the same convention **Inline — the parent
implements** (`skills/flow/implement.md`) uses for implementer and panel-fix rows. `begin` is
recorded before the first command in the list; `end` after the `## Report` is written, carrying
`-outcome completed` (or `-outcome stopped` on the `## Question` handback above).

### The verifier dispatch

`flow.visual-verify` dispatches this subagent, one verifier per worktree — the closed list's one
verifier row (**Dispatch sites — the parent's closed list**, `skills/flow/implement.md`); the
parent dispatches nothing else in this file. `subagent_type: general-purpose`, the Agent tool's
`model` parameter set to `VERIFY_MODEL` (**Model resolution**, `skills/flow/SKILL.md`) — the
literal `sonnet`, never `DEFAULT_MODEL` and never a session override. Its prompt carries, verbatim:

> Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
> Include this instruction verbatim in any prompt you write for another agent.

**The relay contract.** The verifier has no operator channel, fixes nothing, edits no source, and
runs every command in the foreground. Its turn ends with a single `## Report` block, and its last
act before that is writing the same block to
`<abs-worktree>/.superpowers/sdd/verify-report-<key>.md` — `<key>` this dispatch's own key from
**Recording** below — which is what the parent waits on (**Turn discipline**,
`skills/flow/implement.md`). The first line of its first reply is `Model: <the model named in its
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
sonnet`, `-key visual-verify`, suffixed `-<worktree basename>` when this run's resolved set holds
more than one worktree — the pair's semantics are section 4 of `skills/flow/implement.md`, cited
here, not restated.

**A `## Report` carrying any non-zero exit is re-dispatched once.** The second dispatch is recorded
under `-key visual-verify-2` — the same `-<worktree basename>` suffix rule — with a prompt
identical to the first plus the first `## Report` verbatim under a `## Previous attempt` heading,
so the verifier can tell an environmental failure (a build the worktree lacked, a flaky harness)
from a defect in the branch. **The second report is final.** Another non-zero exit ends your turn
with `## Question` naming the failing command and its output, verbatim; the operator resolves it
through a fix run. Never run the failing command yourself to check it, and never dispatch a third
verifier. The ledger render and this stage's `end` mark follow whichever report was last.

**Handshake.** Compare the `Model:` line against `sonnet` (never `DEFAULT_MODEL` or a session
override) and apply **The handshake** (`skills/flow/implement.md`, **The parent orchestrates directly**),
unchanged: a first mismatch closes `<key>` `-outcome fallback` and re-dispatches once under
`<key>-retry`; a second mismatch closes `<key>-retry` `-outcome fallback` too and ends the turn
with `## Question` naming `sonnet` and both models that answered, options **Continue on `<the
model the second handshake named>`** or **Stop the run**.

**A mark or a record never blocks — proceed regardless of whether it reached the store.** A
verifier that ends without a `## Report`, or whose agent dies, is closed `-outcome aborted` and
blocks this handoff exactly as a failed command would, naming the death.

**Load `skills/flow-contracts/session-records.md`** before reading the render outcome below.

**Confirm this run recorded a ledger:**

```bash
flow record render -change <name> -kind ledger -repo <abs-worktree>
```

Read the outcome word, not the exit code. `rendered: <dest>` is ordinary. **`MISSING: ledger — no
rows for <name>` means this run recorded no dispatch at all**, reported plainly here rather than
discovered later. `journalled: ledger` and a non-zero exit are reported the same way. None of these
gates or stops the run — unlike the lint and test exits above. The outcome words are the table under
**Rendering the session records** (`skills/flow-contracts/session-records.md`).

```bash
flow stage end -command '/flow' -stage flow.verify -outcome completed <name>
```

## Visual verification

```bash
flow stage begin -command '/flow' \
  -stage flow.visual-verify \
  -harness <harness> \
  -session-token mf-<literal-token> \
  <name>
```

Reads the `## visual verification` section, canonical in
`skills/flow-contracts/project-configuration.md`. This stage owns its procedure — nothing else in
this pipeline restates it. Resolve once per worktree in this run's resolved set, the same set
**Verify** above resolved:

Steps 1, 2 and 11 are the parent's — those steps, `prepare-workspace.sh` and the ledger render
are the parent's own Bash calls, never a subagent's. Steps 3–10 and 12 are run by one verifier per worktree
surviving steps 1–2, dispatched per **The verifier dispatch** above with `-key visual-verify`; the
parent applies **Blocking** to its report. Its prompt states: the absolute worktree path; the
`KEY=value` lines **Verify** exported for it; this section's resolved `setup`, `verify`, `capture`,
`fingerprint` and `start` commands and `screenshots` root, its resolved `mockups` root when
declared, and its `mockup frame` value when declared; the worktree-resolved URL of each app `ui paths`
matched; the project's `## run` commands; the views touched; `<changeRoot>`; and to run steps 3–10
and 12 below as written, committing and pushing nothing.

1. **Resolve the section** — read that worktree's own `<project>/.flow/project.md` directly, by
   its own shape and closed vocabulary. A project declaring no section → this worktree prints
   `Visual: not configured` and is skipped for the rest of this stage.
2. **Match the diff — with the guard, not by eye.** Run

   ```bash
   git -C <worktree> diff --name-only <merge-base>..HEAD | check-visual-trigger.sh <worktree>
   ```

   Exit 0 → at least one changed path matched a declared `ui paths` glob; continue. Exit 1 → this
   worktree prints `Visual: no UI paths touched` and is skipped for the rest of this stage. Exit 2 →
   the guard could not answer (this should not happen here, since step 1 already confirmed the
   section resolves) — report its stderr and skip this worktree the same way exit 1 does.
   `check-visual-trigger.sh` owns the glob semantics (`**` spanning directories, a leading
   dot-slash prefix, an absolute glob, a glob with a space); nothing here restates them.
3. **Run `setup`, if declared.** A non-zero exit blocks, printing the command verbatim.
4. **Probe before starting anything.** Probe the URL of each app `ui paths` matched, resolved for
   this worktree per **What the id derives** (`skills/flow-contracts/workspace-isolation.md`) —
   never the project's declared default. If nothing answers, start the stack from `start` when
   declared, else `## run`, and record that this stage started it — needed at step 11.
5. **Fingerprint the served bundle, if `fingerprint` is declared.** A screenshot is evidence only
   of what the app was serving when it was taken, and a stack step 4 found already running may be
   serving a build older than the worktree — KAN-29's last fix round captured, and nearly accepted,
   the bug the fix had removed. Run `fingerprint`. Exit 0 → the served bundle is the worktree's
   build; continue. Non-zero → stop the stack, start it from `start` when declared, else `## run`,
   record that this stage started it (step 11 stops it), and run `fingerprint` once more. A second
   non-zero exit blocks, carrying the command's output. No row declared → report
   `fingerprint: not declared` and continue; the report makes the gap visible in every handoff, but
   this stage cannot prove what it was never told how to check.
6. **Run `verify`.** A non-zero exit blocks.
7. **Capture** — author a spec covering the views this change touched, then run `capture` with
   `<spec>` substituted for the spec's path. `screenshots`'s root-not-leaf shape is canonical in
   `skills/flow-contracts/project-configuration.md`; nothing here restates it. **Every screenshot
   this spec takes is the full page or viewport, never a clipped region.** A clip is the right tool
   for an implementer's own targeted assertion (a fixed piece of text, an icon), but this stage's
   own job — page-wide styling (background, shadow, border, font, spacing) matching the mockup — is
   exactly what a clip is built to hide; step 9 below cannot compose a clip against a full mockup
   frame and call the result a fidelity check. **`capture` creates this change's baseline**: writing
   a PNG that does not yet exist is its success path, not a failure — `verify` is the regression gate
   over an already-committed baseline, `capture` is not, and only a `capture` failure for some other
   reason blocks (see **Blocking** below). Then run `check-spec-reach.sh <worktree>` — the spec
   `capture` just wrote must be reached by a `package.json` script of the `regression checkout`;
   exit 1 (an orphan, named) or 2 (cannot answer) blocks.
8. **Read every captured PNG — resolve their paths with the guard, not by eye.** Run

   ```bash
   resolve-visual-screenshots.sh <worktree> <spec's basename>
   ```

   Exit 0 prints one absolute PNG path per line — that is what selects this run's fresh output from
   the committed baseline PNGs already sitting under the same `screenshots` root, rather than joining
   `screenshots` with a guessed filename. Exit 1 (zero matches) blocks, indistinguishable from a view
   that never rendered; exit 2 (cannot answer) blocks the same way. **Read every printed path** — no
   script can do that — and state, per view, what was seen. An unreadable PNG is reported and blocks
   too.
9. **Compose captured frames against their mockups, if `mockups` is declared.** With a
   `<spec>.mockups` sidecar beside the capture spec, run

   ```bash
   resolve-visual-screenshots.sh <worktree> <spec's basename> | compose-mockup-frames.sh <spec>.mockups <worktree>/<mockups> <changeRoot>/visual-verification '<mockup frame value>'
   ```

   The fourth argument is this project's resolved `mockup frame` value, quoted; omit it entirely
   when the row is absent.

   Exit 0 prints one `<composite path> diff=<ratio>` line per pair — **read every one** and state,
   per pair, what the difference panel's white regions are, the ratio, and where the capture
   departs from the frame. The ratio is information, never a threshold: a pair blocks on a
   departure you can name, not on a number. Exit 1 (a broken map — a screenshot no capture
   matched, a frame file absent, a malformed line, a capture whose size differs from the cropped
   frame) or 2 (cannot answer — a malformed `mockup frame` value, Pillow absent, included) blocks.

   **A clean composite read, or a structural match, is not the same as a verified match at the
   control level — treat it as necessary, never sufficient.** Full-page comparison catches wrong
   text, wrong regions, wrong overall layout; it does not catch a border style, an icon's glyph, or
   a colour step, all of which are invisible at full-page scale (KAN-30 fix round 6: a field's
   underline-only focus border, drawn against a mockup showing a full outline, read as a match at
   composite scale and was found only once the two were cropped and zoomed side by side). Before
   accepting any field, button, icon, or toggle as matching its mockup:
   1. **Crop and zoom (2–3x) both the mockup region and the corresponding capture, side by side.**
      Do this for every interactive control the view carries, not only ones that already look
      suspicious at full scale.
   2. **Exercise every interactive state the mockup draws for that control, not only its resting
      one** — a focused field, a field with a value typed (so a clear icon or a validation state
      actually renders), a pressed button, an enabled toggle. A control checked only at rest has
      not been checked at all if the mockup draws it focused or active; capture that state
      separately rather than inferring it from the resting frame.

   **Enumerate every mockup frame the change touches into an explicit checklist before starting
   the sweep, and account for each one by name at handoff — done, or why not.** "Did a
   screen-by-screen sweep" is unfalsifiable without a fixed list: a partial pass reads exactly like
   a complete one if nothing forces naming what was skipped (KAN-30 fix round 7: two prior rounds
   each reported partial coverage as if it were the whole set, and the gap was caught only when the
   operator asked "so is X fully compared and fixed?" of a specific frame). Build the list from
   every frame id `design.md`/`tasks.md` cites for this change, or from the mockups directory
   itself when no such citation exists; a frame reached only by inference from another (already
   covered by the composite diff, "shipped" and pre-existing, blocked by a real environment limit)
   still gets named, with the reason.

   **When exercising a flow by scripted navigation (not manual clicks), screenshot after every
   single action and confirm the resulting screen against an expected marker — a heading, a test
   tag, a distinctive label — before issuing the next action.** Never chain two or more blind
   actions and inspect only the final screenshot. A coordinate that assumed a fixed layout silently
   steers the whole remaining sequence onto the wrong screen the moment real content shifts it — an
   extra suggestion card, a longer note, a wrapped title — and the resulting screenshot can still
   look plausible enough to accept at a glance (KAN-30 fix round 7: a blind click landed on a
   day's "Repeat that workout" suggestion instead of "Create a group session" because an extra card
   existed on that day only, and the sweep almost recorded the wrong screen as verified). This is
   the scripted-navigation analogue of the crop-and-zoom rule above: a plausible end state is
   necessary, never sufficient, evidence that every step along the way went where it was meant to.

   **Check every control against itself, not only against its mockup: crop and diff each
   appearance of a named state (selected, active, pressed) and of a shared component (a button
   variant, a sibling card) across every place it shows in the same flow.** Two captures can each
   match their own frame and still disagree with each other — the mockup draws a state once, so only
   the captures can show the drift (KAN-30 fix round 8: one dialog's "selected" fill differed
   between two sub-states; two sibling cards diverged in shadow; "Add user" carried a border on one
   code path and none on another, because the empty and populated roster states routed to two
   different shared button components). When a screen renders the same logical control from
   different code branches depending on state — empty versus populated, first versus subsequent —
   capture it in every reachable branch and diff the branches against each other, never only the
   branch the walkthrough reached first.

   **Drive every ranged control through its whole range, and every dynamic list or picker into
   its empty state.** A wheel, slider, drag handle or multi-step selector is exercised in every
   direction and past where it wraps or clamps, with a screenshot along the way — the resting
   capture and one direction prove nothing about the other (KAN-30 fix round 8: a time wheel
   scrolled correctly one way only, a state-derivation bug no resting capture shows). A list or
   picker is captured with zero items as well as populated: an empty state sits outside any normal
   walkthrough, which is how one shipped in its pre-restyle appearance beneath a restyled populated
   sibling (same round).

   **For every picker or selector, enumerate its options and ask whether any is predictable to
   fail on submit; one that is gets reported as a defect however good the rejection reads.** No
   mockup draws the invalid-selection case, so no composite can see it — submit the options the
   rules already forbid and name every one the list should have excluded instead of offered
   (KAN-30 fix round 8: a picker offered an option submit always rejected, surfaced only as a
   post-hoc toast). `rules/design-mockups-are-specs.mdc` puts the same question to the
   implementer; this is the verifier's side of it.

   **No sidecar is never a silent skip.** A mockups directory sitting unused is what let kan-30's
   own screens ship four fix rounds deep with their real, drawn frames never once diffed against
   the app — `mockups: no map` was reported and accepted every round, because nothing required
   the sidecar that triggers the compose step to exist. Before reporting `mockups: no map`, list
   `<abs-worktree>/<mockups>` and check it for a frame plausibly matching any view this change
   touched — by filename, by a frame id `design.md` or `tasks.md` cites for this change, or by the
   screen family the touched `ui paths` name. A plausible match exists → **author the
   `<spec>.mockups` sidecar yourself**, one `<screenshot name> <frame id>` line per captured view
   with a real frame, then run the compose command above — composing against a sidecar this stage
   just wrote is not a special case, and it is committed at step 11 along with everything else this
   stage writes. No plausible match anywhere in the directory → report `mockups: no map for
   <spec's basename> — searched <mockups dir>, no frame for <views>`, naming what was searched, and
   continue. Not declared → report `mockups: not declared` and continue. The sidecar's shape is
   canonical in **visual verification** (`skills/flow-contracts/project-configuration.md`).
10. **Write `<changeRoot>/visual-verification.md`** — one entry per view: its absolute screenshot
    path, resolved by the same recursive search step 8 used, and what was seen; and, per composed
    pair, the composite's absolute path, the frame id, its `diff=` ratio, and what was seen.
11. **Commit the spec and its PNGs, and stop there.** A declared `regression checkout` receives
    them; with none declared, commit to the change's own branch instead. **Resolve the
    `regression checkout` root the same way every other declared app root in this file is
    resolved** — from `git worktree list` in that repository, or the state file's `worktrees`
    map, per **Roots in `## apps` are main checkouts** (`skills/flow-contracts/project-configuration.md`) <!-- refs-guard:allow -->
    — never the main checkout while a worktree for it holds the change's work. A `regression
    checkout` not also declared in `## apps` has no worktree to resolve and this step commits to
    the main checkout directly, exactly as before. **Never push** — see `no-automatic-push`
    (design.md): a file inside a repository cannot authorise a push to another repository, so no
    guard here grants one, worktree or not. `regression repo` still records which repository the
    checkout is expected to be, and `check-visual-verification.sh` still reports a mismatch against
    its real `origin`, but that is an identity assertion, not an authorisation. When a commit landed
    in a `regression checkout`, print the push command for the operator to run by hand, naming
    whichever branch actually received the commit — the change's own `spectre/<name>` when a
    worktree resolved, the main checkout's current branch otherwise:

    ```bash
    git -C <regression checkout> push
    ```

    `<changeRoot>/visual-verification/` is committed with the change root.
12. **Stop the stack only if step 4 or step 5 started it.** A stack the operator already had running is left
    alone.

```text verified:design.md section 3 of this change
## Report
- setup: <not declared | exit <n>>
- stack: <already running | started and stopped | could not be started — <output>>
- fingerprint: <not declared | exit 0 | mismatch → restarted → exit <n>>
- verify: exit <n>
  <output, verbatim or last 40 lines>
- capture: exit <n>
  <output, verbatim or last 40 lines>
- spec reach: exit <n>
- spec: <absolute spec path>
- <view>: <absolute PNG path> — <what was seen, including any defect>
- mockups: <not declared | no map for <spec> | exit <n>>
- <frame id>: <absolute composite path> diff=<ratio> — <match, or the departure seen>
- visual-verification.md: written | not written — <reason>
```

Every non-zero exit, unreadable PNG, `resolve-visual-screenshots.sh` exit 1 or 2 and named defect
is carried in the report; the `Visual:` handoff line is built from its view entries.

**Blocking.** This stage blocks the `IN_PROGRESS` handoff on: a failed `setup`, a failed `verify`,
a genuine `capture` failure — **never a first-run snapshot write, which is `capture`'s own success
path per step 7 above** — a stack that could not be started, **a `fingerprint` that still exits
non-zero after step 5's restart**, a `check-spec-reach.sh` exit 1 or 2,
an unreadable PNG, **a `compose-mockup-frames.sh` exit 1 or 2 — including a capture whose size
differs from the cropped frame — and a departure from the mockup the
verifier reports in a composite**, and **a defect the
verifier reports in a captured screenshot — even when every assertion passed.** That last one is the whole
point of this stage: three defects have shipped invisible to a diff, a five-pass review panel and
both test suites, and obvious the moment the page was opened.

```bash
flow stage end -command '/flow' -stage flow.visual-verify -outcome completed <name>
```

## Stage, excluding the planning paths

```bash
flow stage begin -command '/flow' -stage flow.stage-diff -harness <harness> -session-token mf-<literal-token> <name>
```

Confirm every intended task checkbox is `[x]`, and that `git log <merge-base>..HEAD` shows one
commit per completed task, with every fix-round and red-task-partner fixup already folded in via
`git rebase --autosquash` — no stray `fixup!` commit should remain unsquashed, unless a PR already
exists (below). From here to the handoff, each stage's `begin` mark rides its first command and
its `end` mark its last (**Turn discipline**, `skills/flow/implement.md`) <!-- refs-guard:allow -->.

In **every** affected worktree:

```bash
git -C <worktree> status
git -C <worktree> log <merge-base>..HEAD --oneline
```

> **`<project>/spectre/changes/` and `<project>/docs/superpowers/` are never part of a task
> commit.** `<project>/spectre/specs/` is not one of them — a capability spec belongs in the task
> commit that implements its requirement. This step only confirms nothing slipped in.

**Load `skills/flow-contracts/git-boundaries.md`** before committing below.

**The one planning-commit exception.** Every task and fixup commit already sits on the branch,
pushed as it landed (**Branch backup**, `skills/flow-contracts/git-boundaries.md`). If the state
file records a `prUrl`, a PR is already open, so this run also commits
`<project>/spectre/changes/` and `<project>/docs/superpowers/` and pushes everything to the PR
branch; otherwise this step commits nothing more. On that path only — and in this order — run
`flow record render -change <name> -kind all -repo <worktree>`; then `commit-split.sh <worktree>
<name> "<impl-msg>" "chore(spectre): plan and session records"`; then push the branch
`--force-with-lease`, since the split reshaped it. `<impl-msg>`
covers working-tree edits the operator made at the human gate without staging them — derive it the
same way a fixup commit's subject is derived — `fix(<module>): <what changed since the last task
commit>`.

The render overwrites in place. `MISSING: <kind>` means the store holds no rows of that kind and
nothing was written — report it. **A non-zero exit means a destination was refused or could not be
written** — report it, and continue committing the fix.

```bash
flow stage end -command '/flow' -stage flow.stage-diff -outcome completed <name>
```

## Resolve the run instructions

```bash
flow stage begin -command '/flow' -stage flow.run-instructions -harness <harness> -session-token mf-<literal-token> <name>
```

Resolve the run instructions for the handoff's `Running:` section. It writes no file.

- **Every app root is absolute**, resolved from `git worktree list` or the state file's `worktrees`
  keys. Never a relative sibling path, and never a main-checkout path while a worktree holds the
  work.
- **Every start command comes from `<project>/.flow/project.md`'s `## run`**, with every path
  made absolute.
- **Every URL is the one this worktree resolved**, never the project's declared base. Resolve each
  URL from this worktree's workspace id. A project that declares no isolation resolves nothing.
  An application whose port is fixed outside that project's own repository keeps its default,
  named with a short note.
- Apps in scope come from `## apps` in `<project>/.flow/project.md`, or auto-detection.
- **Where the project declares no runnable application**, resolve the `## lint` and `## test`
  commands instead.
- **Before this stage ends, start the stack.** This runs on every run — first and fix alike, not
  only a fix run. A fix run's motivation is the sharpest example: it hands the operator a diff and
  the run instructions this stage resolves, and if the applications those instructions name are
  still serving pre-fix code, the operator reviews one thing and runs another — measured, not
  hypothetical, in this repository's own `<project>/stats/internal/web/embed.go`, whose `//go:embed
  all:dist` makes a running daemon blind to an SPA source change until it is rebuilt. But the same
  applies on a first run: a handoff is meant to hand off a running stack, not a command that starts
  one, so before this stage ends, rebuild and restart every application the run instructions above
  name — the ones just resolved, and no others — from the project's `## run` commands.

  Resolve the start from the two keys the project already declares, in order: its `## stop`, when
  it declares a command, then its `## run`. **A `## stop` that deliberately declares no command
  means there is nothing to stop, not that this rule is skipped** — the start is then whatever
  `## run` does to bring the application up fresh. No new project-configuration key is added for
  this, and none is needed.

  **Never the flow dev stack.** `flowd` on `127.0.0.1:4173`, its `flow-postgres` container and the
  `flow` database inside it are never stopped, restarted or dropped by any run —
  `<project>/CLAUDE.md` states that prohibition and this rule does not weaken it. Where a project's
  own `## run` names that service, the prohibition wins over this start rule, never the reverse.
  This is separate from **Visual verification**'s own start/stop rule above (step 11): that stage
  stops only the stack it started for its own probe, and that rule is not restated here. This rule
  starts whatever the run instructions name, on every run, regardless of whether that stage ran
  or started anything.

  **Where every application `## apps` names is one this prohibition covers, the start is
  nothing — stated, not silently skipped.** This repository is that case: its `## apps` names
  exactly one URL-bearing application, the flow stats daemon on `127.0.0.1:4173`, and that is the
  protected daemon itself. A run against this repository therefore starts nothing before this
  stage ends, and the handoff states which application was skipped and why:

  ```
  Not started: flow stats daemon (http://127.0.0.1:4173) — protected, see
  <project>/CLAUDE.md's "Never stop the dev workspace's stats service or its storage".
  ```

  `flow.visual-verify`'s own `make ui-test-up`/`make ui-test-down` pair (steps 4–5 and step 11 above)
  is a different mechanism entirely — it starts and stops the disposable UI-test stack on
  `127.0.0.1:4174` for that stage's own probe, and `4174` is not an application `## apps` names at
  all, so it is never this rule's start target.

  **A start that fails blocks this stage**, naming the application and what the command printed —
  handing over run instructions that cannot be followed is the failure this rule exists to prevent.
  Where the project declares no runnable application, there is nothing to start and the rule is
  satisfied by saying so, not by silently skipping it.

  Two worked examples, both real, and they resolve differently:

  ```bash verified:read from /Users/tweety53/Projects/gymie/.flow/project.md and this repository's own .flow/project.md
  # gymie — `## stop` declares a command, so the start is stop-then-run:
  ./gradlew devStop
  docker compose up -d && ./gradlew devStart -PfrontendRoot=<abs> -PadminFrontendRoot=<abs>

  # this repository — `## apps` names only the protected daemon, so the start is nothing at all;
  # see the paragraph above for the handoff line this produces instead of a command.
  ```

- **The stack behind the URLs is checked, not trusted.** When the `Running:` block below would
  carry URL lines, run `check-dev-stack-fresh.sh <worktree>` first: the project's declared
  `fingerprint` row (the `## visual verification` section) proves what the stack serves is the
  worktree's own build — the check KAN-334's handoff lacked when it printed a URL backed by a
  bundle eleven commits stale. The guard's own header is canonical for what its three exits
  cover. Exit 0 adds nothing. Exit 1 adds one line beside the URLs, naming the application, its
  resolved URL and the guard's own stderr reason, restart-shaped and never a restatement of the
  verdict: `Stale: <app> (<url>) — <the guard's reason>; restart the stack before testing.`
  Exit 2 adds `Freshness: unverified — <the guard's stderr reason>` instead, the same
  visible-gap rule **Visual verification**'s own step 5 runs on a missing row. A refused start
  above already ends the run, so this check only ever runs on a start that succeeded or was
  skipped by the protected-service rule.

- **The `Running:` block is the start command's own output.** Once the start above succeeds, its
  `Running:` lines are every line of that command's output containing `http://` or `https://`,
  verbatim, followed by the `## stop` command (or, where `## stop` declares none, the same `## run`
  command that started it). `devStart`-style commands already announce every resolved URL on
  success; that announcement — not the isolation table, which cannot know the pinned-port handoff
  stack's URLs once a port has moved — is the source.

- **A refused start is relayed, never resolved.** When the start command above exits non-zero
  because a port it needs is already held, this stage does not retry, does not pick another port,
  and does not stop the holder. It prints that command's output verbatim, then the holder's own
  `devStop`-style stop command when the output names the holder's worktree (or, when it does not,
  a note to run `check-worktree-processes.sh`-style diagnosis and stop the holder's stack by hand).
  This run then ends at `IN_PROGRESS` with the diff still staged, and the handoff names `/flow
  <name>` as the next command — the same change, re-run once the holder is stopped. It never stops
  a stack this run did not start.

```bash
flow stage end -command '/flow' -stage flow.run-instructions -outcome completed <name>
```

## Write `IN_PROGRESS`

```bash
flow stage begin -command '/flow' -stage flow.write-in-progress -harness <harness> -session-token mf-<literal-token> <name>
```

**Append this run's own narrative first.** Append to
`<abs-worktree>/spectre/changes/<name>/narrative.md` (create it with the title `# <name> —
session narrative` when absent) one section `## <YYYY-MM-DD> — <creating run | fix run>` holding
this session's own prose account of the run — problems hit, workarounds, time sinks, environment
gaps, operator decisions taken mid-run — and nothing the ledger or panel record already holds. It
rides the next planning commit through `commit-split.sh`'s existing `<project>/spectre/changes/` pathspec;
nothing else stages it.

Write the state file: `IN_PROGRESS` from `STARTED`, otherwise **the state exactly as read**.
Populate `worktrees` with one absolute-path key per affected worktree and its merge base. Carry
`artifactUrl` (always `null` under `/flow`), `jiraIssue`, `planningEffort` (always `null`),
`models.default` (always `null` — `/flow` resolves models from the settings store per run, never
records a value into the per-change state) and `prUrl` forward verbatim. The state file lives
outside the repo — never `git add` it.

```bash
flow stage end -command '/flow' -stage flow.write-in-progress -outcome completed <name>
```

**Produce the handoff's `Records:` count**, one call per affected worktree:

```bash
flow record journal-count -change <name> -C <abs-worktree>
```

**Produce the handoff's `Costs:` line the same way**, one call per affected worktree:

```bash
flow record cost-status -change <name>
```

It exits 0 always — `unknown` included. Render exactly what it printed.

**Produce the handoff's `Deferred:` count and `### Deferred minors` list the same way**, one call
per affected worktree:

```bash
flow record findings -change <name> -C <abs-worktree>
```

Filter the result on a `status` that starts with `deferred`. `Deferred:` is the count of matches;
`### Deferred minors` lists one row per match, `F<n> <location> — <note> — <reason>` (the reason is
the text following `deferred ` in that finding's status), and reads `none` when the count is `0`.

```
## Implementation staged — review and test | Implementation committed — review and test

**Change:** <name>
**Panel:** clean — roster: <the slot list this run dispatched>; reduced: <"docs-only — " followed by the resolved slot(s) not dispatched, or "no">; <default|dynamic — class, compact?, rerun policy, dispatches: <group> · <group>>; added this run: <slot(s) an explicit operator instruction added beyond the resolved list, or "none — resolved list ran alone">
**Visual:** not configured | no UI paths touched | <view>: <absolute screenshot path>[, <view>: <absolute screenshot path> …][ — push with: git -C <regression checkout> push]
**Staged:** N/N tasks staged and uncommitted | N/N tasks committed on branch | committed, plus one planning-artifacts commit, and pushed to the PR branch
**Records:** all writes reached the store | N write(s) journalled — the store was unreachable | unknown — the journal could not be counted
**Deferred:** <count of deferred Minors>
**Costs:** <the line `flow record cost-status` printed>
**Guards:** all present | N missing — those checks were performed by hand (see the guard presence check above)
**Jira description (pre-edit):** <the text as it stood before the write, verbatim in a fenced block, inside <details> when long> | omitted — this run wrote no description

Worktree:   <absolute worktree path>

Running:
  <url line>  # <from the start command's output>
  <stop command>

Review the diff, then run it:
  git -C <absolute worktree path> diff <merge base>..HEAD
  open -na "IntelliJ IDEA" --args "<absolute worktree path>"

### Deferred minors
<one row per deferred Minor, `F<n> <location> — <note> — <reason>`, or `none` when there are none>

Re-run this command to fix anything you find, or bare to move on to integrating it.

Next:
/flow <name>
```

**Heading and `Staged:` line select from whether the worktree carries any commits yet** — a
creating run's very first `IN_PROGRESS` write, reached before any task committed, would be an
anomaly (implementation always commits per task), so in practice this always reads "committed" —
the "staged and uncommitted" alternative is carried only for symmetry with the phrasing an operator
resuming mid-panel might see, and should not occur in an ordinary run. **The `Panel:` line is
`/flow`'s own** — it states what **Review panel** (`skills/flow/review-panel.md`) actually
dispatched this run: the resolved roster or its docs-only reduction to `primary` (**The docs-only
reduction**, `skills/flow/review-panel.md`), any slot an explicit operator instruction added
beyond the resolved list; and, per **Bundled dispatch** and **The `## Decision` block**
(`skills/flow/review-panel.md`, `skills/flow/brainstorm.md`), whether `REVIEW_PANEL_TOGGLE` was
`default` or `dynamic`, the run's class, whether the roster was `compact` or `full`, its rerun
policy (`delta` or `full`), and the dispatch groups as `+`-joined roles — the same fields and
shape `skills/flow-contracts/handoff-blocks.md`'s `Panel:` line carries for `/flow-status`'s
regenerated view of the same state.

**The `Records` line is printed on every run of this branch, journalled or not.** **The `Costs:`
line is printed the same way — always, `unknown` included.**

**The `Visual:` line reports `flow.visual-verify`'s own outcome.** Every screenshot path in it is
absolute, per **Handoff output** (`skills/flow-contracts/pipeline.md`)'s every-path-is-absolute
rule — the operator must be able to open the PNG. **Its push clause appears only when step 10
committed to a `regression checkout`** — the stage never pushes itself, per `no-automatic-push`, so
this is the command the operator runs by hand to land that commit.

The pre-edit description line is present only on a fix run that synced the description in **3.
Documenting a fix** (`skills/flow/implement.md`), and reproduces that text without summarising or
reflowing it.

**The parent prints this block directly, as this stage's own output** — no return, no relay: the
running session assembles it here and shows it to the operator in the same turn (**The parent
orchestrates directly**, `skills/flow/implement.md`).

## Guardrails

- **Commit per task and per fixup** — never `<project>/spectre/changes/` or
  `<project>/docs/superpowers/` in a task or fixup commit. **Never push, merge, or open a PR** —
  except the `prUrl` exception above.
- **Never** run `finishing-a-development-branch`.
- **Never** create a second worktree for the same change.
- **Never** advance the state from `IN_PROGRESS`; write back what you read.
- **Never** hand off with an open finding of any severity, or a stale clean result — stale as
  **Panel re-runs** (`skills/flow/review-panel.md`) defines it.
- **Never** mark a task's checkbox before that task's review passes.
- **Never** edit source after the panel closes — a fix run is the only path that changes source.
