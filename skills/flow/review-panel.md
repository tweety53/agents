# Review panel

Loaded by `skills/flow/SKILL.md` immediately after `skills/flow/implement.md`'s `flow.sdd-tdd`
stage closes, on every implementation run — creating, resumed, or fix. Dispatches the resolved
roster: this run's decision's `panel.roster` (`<abs-worktree>/.superpowers/sdd/decision.json`, per
the `## Decision` block of **Decide**, `skills/flow/brainstorm-planner.md`), or — when the decision's `panel` is the string `default`,
a `micro` class — `REVIEWERS`, which `skills/flow/SKILL.md`'s **Model resolution** resolves from the
settings store and is canonical for. This file owns dispatch: mapping each resolved id to
its slot and spawning it.

```bash
flow stage begin -command '/flow' -stage flow.review-panel -harness <harness> -session-token mf-<literal-token> <name>
```

**The pass log is store rows, rendered — never a hand-written file.** Every fact this file records
about a panel run the parent records as it arises with `flow record pass` or `flow record mutation`
(`-change <name> -round <n>`, the round `0` for the initial panel and `1..n` for a fix round).
`flow record render -kind panel` renders them into the panel record's pass-log section under
`<abs-worktree>/.superpowers/sdd/reviews/` in the canonical worktree, beside the findings.

## Check base movement first

Once per worktree in this run's resolved set (**Resolving a change's worktrees**,
`skills/flow-contracts/worktree-resolution.md`), on every panel run — creating, resumed, or fix:

```bash
BASE="$(resolve-base-branch.sh <worktree>)"
check-base-moved.sh <worktree> "origin/$BASE" <working-notes-merge-base>
```

`<working-notes-merge-base>` is the merge base `skills/flow/implement.md`'s isolate-workspace step
recorded in this run's working notes — never the state file's `worktrees` map, which a creating run
has not written yet. `check-base-moved.sh` performs no fetch of its own; `resolve-base-branch.sh` is
what fetches, so this order — resolve, then check — is load-bearing.

Report every worktree's verdict: `MOVED` with no overlap is confirmed conflict-free and rebases
that worktree automatically — `git -C <worktree> rebase origin/$BASE` runs at once, with no
prompt, and the run reports that it happened rather than asking whether it should; its outcome
takes the **Clean** and **Conflict** sub-bullets below exactly as an operator-chosen **Rebase**
would. `REFUSE`, an exit 2, or an empty resolved set stops and asks; an overlap from any worktree
means the rebase is not confirmed conflict-free, so the run never takes that risk on itself — one
prompt for the whole change, shape per Operator prompts
(`skills/flow-contracts/operator-prompts.md`), which that contract's **Auto-resolution** resolves on
its recommended **Stop**; **Rebase** and **Continue** run only on an explicit operator instruction:

> **The base branch has moved and touches paths this change also touched — how should the
> panel proceed?**
> - **Stop — I'll rebase or reorder first** *(recommended)*
> - **Rebase onto `<base>` now, then continue**
> - **Continue — review as is**

**Stop** closes `flow.review-panel` with `-outcome stopped` and leaves the change at its current
state with nothing committed by this stage. **Continue** carries the reported movement into the
handoff and proceeds to the citation pre-check below.

**Rebase** runs `git -C <worktree> rebase origin/$BASE` only in a worktree whose own verdict was
`MOVED` — never one whose verdict was `CLEAR`, even though the prompt above is asked once for the
whole change. The automatic no-overlap rebase above is the same mechanism with nobody asked, and
the two sub-bullets that follow govern both:

- **Clean** (exit 0): that worktree's working-notes merge base becomes `origin/$BASE`'s resolved
  tip at rebase time; every later `<merge-base>` this file and the `worktrees` map
  `skills/flow/verify-and-handoff.md` writes read the working notes, so nothing else needs
  plumbing. Re-run `check-base-moved.sh` once more against the new value; a fresh overlap re-offers
  the prompt above rather than looping silently. The rebase clears every slot's held last-reviewed
  sha, so a re-run after it reads the whole `final-review.diff` under **Panel re-runs**' existing
  no-held-sha rule below. No re-verification runs here — the panel reads the rebased tree, and
  `flow.verify` follows. End the clean-rebase report with the fixed literal:

  > If this change's verification compares against a recorded baseline, recapture it now — a
  > proof taken against the pre-rebase base is void.

- **Conflict** (non-zero exit): never auto-abort, and never resolve the conflict — by editing the
  conflicting files or otherwise. An automatic rebase lands here too — no reported overlap does
  not rule out two commits touching one file incompatibly outside this change's own touched paths
  — and is handled identically. Leave the worktree mid-rebase exactly as `git rebase` left it,
  report the conflicting file(s) from `git status`, and hand off `git -C <worktree> rebase
  --continue` (after the **operator** resolves it) or `git -C <worktree> rebase --abort` as the
  next manual step. State stays as it was; this stage stops here and closes the mark `stopped`.

**Everything from the citation pre-check to the throwaway worktrees and the `[PRINCIPLES_PATH]`
and `[STANDARDS_PATHS]` resolution below is one Bash call**, under the turn discipline of **4. Execute (SDD + TDD)**
(`skills/flow/implement.md`): every verdict printed, and only the dispatch depends on them — an
over-cap exit, a `REFUSE`, an absent principles file or an operator prompt is read off that call's
output and handled before any slot launches. The base-movement check above is the one exception:
its rebase branch changes what every later step reads, so it runs and is read first.

**Run the citation pre-check before rebuilding the dispatch context bundle below**:

```bash
check-panel-citation-trigger.sh <worktree> <merge-base>
```

On exit 0, run `project-get.sh <worktree> "review panel citation check"` (**Guard resolution**,
`skills/flow-contracts/pipeline.md`; the key itself is canonical in **Project configuration**,
`skills/flow-contracts/project-configuration.md`). If declared, run its command
from the apply worktree and capture combined stdout+stderr verbatim to
`<abs-worktree>/.superpowers/sdd/citation-check.md`. `project-get.sh` exit 1 is the "key absent —
skip silently" case just named; exit 2 is reported and the worktree skipped, the same way the
guard's own exit 2 below is. On exit 1, skip silently — no file written, no CITATION CHECK
paragraph added below. Exit 2 — the guard could not answer (usage error, `<worktree>` not a
directory or not a git repo, or the merge-base not resolving) — report its stderr and skip this
worktree the same way exit 1 does.

**Never blocks.** The configured command's exit code is read by nobody — a non-zero exit (findings
present) still just writes the file. The real gate stays `flow.verify`'s existing `## lint` run,
unchanged by this step.

**Rebuild the dispatch context bundle at the start of this stage too** — never reused from <!-- refs-guard:allow -->
`skills/flow/implement.md`'s run. Overwrite the same path:

```bash
mkdir -p <worktree>/.superpowers/sdd
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  <worktree>/.superpowers/sdd/dispatch-context.md "" <canonical-worktree> <shape>
```

Report the script's stderr line (`bundle unchanged — reusing …` or `bundle rebuilt — …`) as part
of this stage's own reporting.

**A failed build never dispatches silently**, shape per Operator prompts
(`skills/flow-contracts/operator-prompts.md`):

> **CONTEXT BUNDLE FAILURE:** the gather above exited non-zero, or the bundle file is still
> absent when the `test -f` on the rebuild's output path above says so — a failed
> build. Record the cause with `flow record pass -round <round> -note 'context bundle: build
> failed — <the script's stderr>'`, then resolve it, never asked, per **Auto-resolution**
> (`skills/flow-contracts/operator-prompts.md`):
> - **Stop — resolve the build, then re-run the panel** *(recommended; silence defaults here)* —
>   closes `flow.review-panel` with `-outcome stopped`
> - **Continue — dispatch every slot without the bundle** — recorded with `flow record pass
>   -round <round> -note 'context bundle: dispatched without it — operator override'`, every
>   slot's prompt dropping the CONTEXT BUNDLE paragraph and naming the bundle's absence in its
>   place, so the reviewers' reading the plan and decision directly is on the record, never a
>   silent reduction of their context

## The roster

Every resolved id maps to one slot, dispatched this run because the resolved roster (the decision's `panel.roster`,
or `REVIEWERS` on a `default` panel — see the opening paragraph above) carries it — the
spawn column names the `default`-panel spawn, which the decided-panel paragraph below
overrides:

| id | Slot | How to spawn |
|---|------|---------------|
| `primary` | **Primary** — plan alignment and senior code review | `flow-low` + `primary-reviewer-prompt.md`: `final-review.diff` against `proposal.md`, `design.md` and each task's `**Files:**`/`**Tests:**`/`**Commit:**` fields in `tasks.md`, plus code quality, architecture, testing and production readiness |
| `principles` | **Principles** | `flow-low` + `principles-reviewer-prompt.md`; all three principle groups always apply <!-- refs-guard:allow --> |
| `failure-modes` | **Failure-modes** — error return, timeout, partial write, concurrent re-entry | `flow-low` + `failure-modes-reviewer-prompt.md` |
| `mutation` | **Mutation** — sabotage-proofing | `flow-low` + the mutation-testing brief below, own throwaway worktree copy per repository (see **The throwaway worktree**, `skills/flow/review-panel-optional-slots.md`) |

**A subagent-facing file is passed by absolute path, never read into this context.** This repository's
`primary-reviewer-prompt.md` (Primary), `principles-reviewer-prompt.md` and
`engineering-principles.md` (Principles),
`failure-modes-reviewer-prompt.md` (Failure-modes), and
`<project>/.flow/project.md`'s standards files are inputs to the slot that reads them; the
dispatcher resolves their paths, confirms each exists, and names them in the prompt.

**Fill every template placeholder from this table — never by reading the template.** Each is
substituted in the dispatch prompt with the value named here:

| Placeholder | Template | The parent substitutes |
|---|---|---|
| `[DIFF_PATH]` | all three | the absolute path of the diff this dispatch reads, in `<abs-worktree>/.superpowers/sdd/` of the canonical worktree: `final-review.diff` in pass 1; `late-fix.diff` under the late-fix reduction; on a re-run, `slot-delta-<round>-<id>.diff` or the round's `fix-round-N.diff`, whichever the panel re-run rules below assign the slot |
| `[ARTIFACT_PATHS]` | primary | the absolute paths of the change's `proposal.md`, `design.md` and `tasks.md` in the plan directory the dispatch context bundle reads (`<changeRoot>`, or a satellite's canonical change directory) |
| `[CONTEXT_BUNDLE_PATHS]` | primary, failure-modes | the paths the CONTEXT BUNDLE paragraph names, `<abs-worktree>/.superpowers/sdd/dispatch-context.md` once per worktree in the resolved set; on the CONTEXT BUNDLE FAILURE continue path, the literal `none — the bundle was not built` |
| `[PRINCIPLES_PATH]` | principles | as the Principles section below resolves it |
| `[STANDARDS_PATHS]` | principles | as the Principles section below resolves it; empty when none resolve |
| `[GLOBAL_CONSTRAINTS]` | principles | the literal `the design.md section of your context bundle` — never a selection or paraphrase; on the CONTEXT BUNDLE FAILURE continue path, the absolute path of the change's `design.md`, or `none` when it has none |

`ValidReviewers` in `<agents repo>/stats/internal/store/settings.go` is the id vocabulary this table exhausts.
`DEFAULT_MODEL` is `skills/flow/SKILL.md`'s **Model resolution** value
for this run. Every slot in this table
is dispatched on `flow-low` (`agents/flow-low.md`, on
a `default` panel) and carries the same model rule. The `flow-<effort>` definitions are
this repository's own, whose `tools:` allowlist omits `Agent` — **No forking** below is backed by a
capability the slot structurally does not have, not by the NO DELEGATION paragraph alone;
`general-purpose` is a harness-provided type whose tool set cannot be restricted.

**Check for an operator-named id at two points**: at the start of this stage (has the operator, in
this run's own argument or in the session before this stage, named an id the resolved list does not
carry?), and again at the start of every fix round below — an operator may ask mid-run, after seeing
pass 1's result, and that request adds the slot starting from the round it was made, never
retroactively to a pass already closed. It is never written back to the settings store. Record which
slots were added this way and why (the operator's own words) with `flow record pass -round <round>`,
and record explicitly when none were: "no addition this round — the resolved list ran alone."

**On a decided panel** (the decision's `panel` an object), model and effort belong to the dispatch, not the slot:
each entry of the decision's `panel.dispatches` carries its `slots` and its own `model` and
`effort`, and every slot in it runs on that pair in pass 1 — and in every fix-round re-run on the
decision's `panel.rerun_dispatch` pair instead (**Panel re-runs**) — the dispatch's `subagent_type` is
`flow-<effort>` — the effort comes from the definition, the model from the Agent tool's own
`model` parameter, passed explicitly on the dispatch — and both `model` and `-effort` are recorded. The roster carries no per-slot model. On harness `zcode` the pair given and recorded is `glm-5.3-flash` / `high` instead (**Harness mapping**, `skills/flow-contracts/model-policy.md`). A compact roster
(the decision's `panel.compact`) is recorded with `flow record pass -round 0 -note 'roster: compact — <rolled value>'`; a full
roster records `roster: full`.

**Load `skills/flow/review-panel-optional-slots.md`** before dispatching any round whose roster
carries `mutation` or an `exp-` slot — it carries **Experimental slot** and **The throwaway
worktree**. A round whose roster carries none of them never reads it.

**Before writing `final-review.diff`**, run

```bash
generate-relocation-comparison.sh <worktree> <changeRoot> <merge-base>
```

A non-zero exit (`2` — cannot answer) prints one line and the run continues without the comparison
file. This call is never a gate — generation never blocks the panel.

```bash
check-panel-diff-size.sh <worktree> <merge-base>
```

once per worktree in the resolved set, unchanged. **The gating count is the sum across
worktrees** — one slot now reads every section — and the over-cap report names the sum and each
worktree's own count.

Exit 0 proceeds. Exit 1 proceeds too, unasked: the panel dispatches reading the whole diff
regardless of its size, and the over-cap fact is reported, never put to the operator as a
question. Exit 2 — the guard could not measure, not an over-cap verdict — stops the run. Record
the measured count, the cap in force, and the automatic proceed decision where the cap was
exceeded with `flow record pass -round <round>` on **every** run, including exit-0 runs.

Then run

```bash
check-panel-docs-only.sh <worktree> <merge-base>
```

### The docs-only reduction

**Exit 0 from
every worktree in the resolved set — every path this branch touched, committed since the merge
base, staged or unstaged, ends `.md` or `.mdc` — reduces pass 1 to `primary` alone**, plus every
slot the operator's per-run instruction
named at this stage's start. Every other resolved slot is recorded with
`flow record pass -round 0 -note 'not dispatched — docs-only reduction: <slot>'`.
`primary` is the reduced roster even when the resolved list does not carry it — the
same shape **Model resolution** (`skills/flow/SKILL.md`) already defines for an empty store list.
This reduction applies to a decided roster unchanged: it still narrows to `primary` alone, on the model and
effort of the decided dispatch that carried `primary` — one dispatch, never bundled.

**Exit 1 runs the resolved roster unchanged**; the first non-documentation path any worktree's run
printed is recorded beside the verdict. An empty touched-path set is exit 1 too. One worktree at
exit 1 or 2 runs the resolved roster unchanged for the whole change.
**Exit 2 reports the guard's stderr and runs the resolved roster unchanged**: an unanswered
question never reduces a panel.

Record the verdict, the printed path where there is one, and the roster actually dispatched with
`flow record pass -round <round>` on **every** run, beside the diff-size
fields above.

**This and the late-fix reduction below are the only automatic reductions, and they only ever
remove.** No slot is ever added by diff
size, touched area, or any other automatic trigger: beyond the reduced or resolved roster,
anything reaches the panel only through an explicit per-run operator instruction, for that run
only.

Write `<abs-worktree>/.superpowers/sdd/final-review.diff` (the canonical worktree's) once per round from **every**
worktree in the change's resolved set (**Resolving a change's worktrees**,
`skills/flow-contracts/worktree-resolution.md`), in resolved order — one first line naming the
file's own semantics, then each worktree's section, opened by a header naming it and its own
working-notes merge base, followed by that worktree's `git diff <merge-base>` (staged and
unstaged):

```sh
: > <abs-worktree>/.superpowers/sdd/final-review.diff
printf '# final-review.diff — working tree vs merge-base, unstaged changes included\n' \
  >> <abs-worktree>/.superpowers/sdd/final-review.diff
# for each <worktree> in the resolved set, in order:
printf '# worktree: %s — merge base %s\n' "<worktree>" "<merge-base>" \
  >> <abs-worktree>/.superpowers/sdd/final-review.diff
git -C <worktree> diff <merge-base> >> <abs-worktree>/.superpowers/sdd/final-review.diff
```

A single-worktree change writes the same shape with one header. The first line is the file's
own answer to the reviewer who reads it as anything narrower: it is a plain working-tree diff
against the merge base, so work still unstaged or uncommitted is already in it. Then dispatch the round's
`panel.dispatches` in the canonical worktree, each reading the whole combined file; a role is
never dispatched once per worktree: one pass reads every worktree's section, so a seam between two repositories is in one pass's view. **Bundled dispatch** below states how the roster is
grouped into those dispatches.

### The late-fix reduction

**A fix run whose delta is small against an already-clean, already-verified stage reduces pass 1
to `primary` alone reading the delta only** — the narrow late-fix review path. An appended fix
that a full pass would bury under re-covered ground is reviewed by one targeted dispatch on what
the fix actually changed; the full-read coverage stays exactly where it earns its keep, on every
round this reduction does not fire for. Like the docs-only reduction this only ever removes, and
it fires only on a **fix run** — a `/flow` invocation whose argument is fix instructions — never
on a creating run, whose pass 1 is always the full resolved roster. Every condition must hold,
checked once when this stage opens its first round on the fix run:

1. **Already clean, already verified** — the change's store findings carry no `open` row
   (`check-panel-findings-closed.sh <worktree> <change>` exits 0), and this stage has closed
   clean before on this change, the pass log or the rendered panel record naming that close;
2. **the base has not moved since that close** — every worktree's `Check base movement first`
   verdict this round was `CLEAR`, so no rebase has invalidated the reviewed state;
3. **the delta is small** — per worktree, `git diff --numstat` from the sha that close reviewed
   there to the current tree, insertions and deletions summed across the resolved set, is at
   most **40 changed lines**; the `-diff-base` its clean dispatches recorded names the canonical
   worktree's sha, and a peer worktree's since-close sha comes from the panel record's
   per-worktree sha list;
4. **no scope growth** — the delta adds no task line to the change's plan `tasks.md`;
5. **the panel's own machinery is untouched** — the delta names no path under
   `skills/flow/review-panel*.md`, no `scripts/check-panel-*.sh` guard, and no reviewer-prompt or
   principles file the panel's own slots read as their instructions
   (`skills/flow/*-reviewer-prompt.md`, `skills/flow/engineering-principles.md`) — of the
   repository the change edits. A consuming project typically carries none of those paths and
   the condition holds vacuously there; a fix to anything the panel reads as its own brief is
   never its own reviewer.

**On trigger, pass 1 is one dispatch: `primary` alone**, plus every slot the operator named at
this stage's start, reading not the whole `final-review.diff` but only a
`<abs-worktree>/.superpowers/sdd/late-fix.diff` written from the since-close range — the same
per-worktree sectioned shape as `final-review.diff`, each `# worktree:` header naming the
since-close sha. On a decided panel the dispatch runs on the decision's `panel.rerun_dispatch`
pair under the fix-round re-run's 5-minute ceiling; on a `default` panel, which carries no
decision to read a pair from, on `DEFAULT_MODEL` at `low` effort under the ordinary 15-minute
ceiling. Mutation and Failure-modes are not dispatched, and each dropped slot is recorded
with `flow record pass -round <round> -note 'not dispatched — late-fix reduction: <slot>'`, the
docs-only reduction's own convention. Every entry check above still runs as any round's — base
movement, the diff-size cap, the docs-only guard — and where both reductions fire, the late-fix
one governs the read scope, the roster being `primary` alone either way. The reduction itself is
recorded with `flow record pass -round <round>`:
`late-fix reduction: <n> changed lines since <sha>`.

A finding the targeted dispatch raises feeds the ordinary fix-round loop unchanged, and Minors
defer under the standing rule — but **any Critical or Important it raises voids the reduction
for the rest of the run**: a delta that small producing a defect that severe means the narrow
read's context was not enough, so every later round this run opens takes the full path. When the
targeted dispatch reads the delta clean, or raises only Minors that defer under the standing
rule, the close sha moves to the round's HEAD and the stage closes under the existing rules —
the reduction is a read-scope decision, never a weaker close.

### Bundled dispatch

**At most two review dispatches per round, each carrying one to three roles**, on
decided and `default` panels alike and in both execution modes. A dispatch carrying one role covers that role alone; a roster the
two dispatches cannot hold shrinks to what they hold.

**Grouping.** On a decided panel, the decision's `panel.grouping` is `static` — the class's row in
the tree of **Decide** (`skills/flow/brainstorm-planner.md`), unchanged, no override — or `free` — the
planner's own grouping within the ≤2 × ≤3 cap, recorded as `panel.grouping_reason`. On a `default` panel,
the settings-store roster is grouped deterministically by the same static logic, no roll and no
planner: the floor roles (`primary`, `principles`) fill the first
dispatch, every other role the second, up to three, the mutating role (`mutation`) last; a list the two cannot hold is truncated in store order, and the truncation is
recorded with `flow record pass -round <round>`.

**One `dispatches` row per bundle** — the same `flow record dispatch begin`/`end` pair below, with
`-slot` the bundle's roles `+`-joined in roster order (`primary+principles+failure-modes`) and
`-model`/`-effort` the bundle's own, from the decision's `panel.dispatches` entry. Every finding still records its own single role in `-slot`, with the
bundle's `-dispatch-seq`.

An incident a round's slot caused takes the same course (the incident rule of **4. Execute
(SDD + TDD)**, `skills/flow/implement.md`): recorded with `flow record incident`, and the next dispatch to that role — a
re-run, or the panel-fix subagent — carries it verbatim.

**The bundle prompt** carries the shared paragraphs — CONTEXT BUNDLE, WORKTREES, TOOLS, FOREGROUND
BUILDS, REPRODUCE DON'T READ, CITATION CHECK, ENTRY CONTEXT, MODEL HANDSHAKE, the reproducer rule —
once, then one
**PASS `<id>`** section per role in roster order, each carrying exactly the brief that role's solo
dispatch carries above and its own REPORT FILE line naming `panel-report-<round>-<id>.md`. The mutating
role (`mutation`) is always the last pass of a bundle and still works in its
throwaway copy (**The throwaway worktree**, `skills/flow/review-panel-optional-slots.md`); the reading passes before them read the shared
`<worktree>`. The return message carries one findings summary per role under a heading naming the
role; the parent records each finding under that role.

Every bundle prompt also carries this paragraph verbatim:

> **INDEPENDENT PASSES:** each pass starts from `final-review.diff` and the code, never from an
> earlier pass's report or conclusions. Do not cite, defer to, or skip a defect because an earlier
> pass raised it — if it sits in this pass's angle, raise it again under this pass. Write each
> pass's report file before beginning the next pass.

**No de-duplication across roles**: the same defect raised by two passes is two `F<n>` rows.

**Re-runs are re-grouped by the same grouping**, carrying only the roles re-running this round — a
group whose other members are clean dispatches with its re-running members only. On
a decided panel a fix round's re-running roles are never bundled: one dispatch per
role, its `-slot` that role alone (**Panel re-runs**).

The rendered panel record's pass-log section and the `IN_PROGRESS` handoff's `Panel:`
line name the dispatches as `+`-joined groups (`primary+principles · failure-modes+mutation`).

**Every slot's dispatch is recorded**, the same pair section 4 of `skills/flow/implement.md`
records for an implementer:

```bash
flow record dispatch begin -change <name> -role reviewer -slot <slot|slot+slot+slot> -model <m> -effort <e> \
  -diff-base <sha> -key panel-<round>-<slot|slot+slot+slot> \
  -session-token mf-<literal-token>
flow record dispatch end -change <name> -key panel-<round>-<slot|slot+slot+slot> \
  -session-token mf-<literal-token> -outcome completed
```

Every dispatch of a round launches in one message; every `begin` is recorded in the next Bash
call, one call for all; the round's wait is one call whose condition is `test -s` on every
launched pass's report file; every `end` is recorded in one call once they all exist (the turn discipline of **4. Execute
(SDD + TDD)**, `skills/flow/implement.md`). A dispatch whose report never appears within its
ceiling takes the breach path under **No forking, and a wall-clock ceiling on every slot** below.

The same plan-tree discipline rides every round, re-run rounds included: before the launches the
parent makes the reviewer-dispatch planning commit (**Planning commits**,
`skills/flow-contracts/git-boundaries.md`) and runs
`check-plan-unchanged.sh snapshot <worktree> <name> <snapshot-file>`, and once every report file
exists it runs `check-plan-unchanged.sh verify <worktree> <name> <snapshot-file>` before any
report is read or finding recorded — the slots read those artifacts, and a flight that changed them has
invalidated the reviews that flew. Exit 1 or 2 stops the round the same way the per-task
reviewer's stop works (the plan-tree rule of **4. Execute
(SDD + TDD)**, `skills/flow/implement.md`); the slots' own read-only briefs are the first line of defense, this
verify is the assertion that a breach cannot slide past as a clean report. The round brackets
itself with content markers beside that guard (the content-marker rule of **4. Execute
(SDD + TDD)**, `skills/flow/implement.md`): the marker list names the plan
artifacts and working notes the round reads, `check-tree-markers.sh snapshot <worktree>
<markers-file> <snapshot-file>` runs with the plan-tree guard's own snapshot, and
`check-tree-markers.sh verify <worktree> <markers-file> <snapshot-file>` runs before any report
is read or finding recorded — a slot's clean report is never the answer to what happened to the tree.

`-slot` names the dispatch's roles from **The roster** table above. `-role` is
`reviewer` for every one; `-task` is omitted. `-diff-base <sha>` is passed on a dispatch whose
roles are all reading against a delta and on no other; it takes one
sha, so it carries the **canonical worktree's** held last-reviewed sha, and the
panel record names every worktree's sha beside the delta path. `-model` is `DEFAULT_MODEL` (or this run's override) on
a `default` panel and the dispatch's own model from the decision's
`panel.dispatches` on a decided panel — bundled or one-role alike, no exception — or, on a
fix-round re-run on a decided panel, `panel.rerun_dispatch`'s model. `-effort` likewise:
`low` on a `default` panel, and the dispatch's own effort on a decided panel,
`panel.rerun_dispatch`'s `low` on a re-run.

**`-agent-id` is never typed on a slot's row, never invented** — the daemon captures the launch identifier,
pairing each launch with the begin nearest it in time;
record the round's `begin`s in the one Bash call after its launch message, as above, and never reuse a `-key`.

**Record a slot's dispatch before recording that slot's findings**, and carry the seq the command
printed — `recorded: dispatch <seq>` — into each of that slot's `flow record finding` calls as
`-dispatch-seq <seq>`.

**Every slot must supply, per finding, a reproducer**: a runnable command that demonstrates the
defect, or the literal exemption form `none — <reason>`. **The exit-code convention is fixed and
one-directional: a reproducer exits non-zero when the defect it demonstrates is present, and exits
0 only when that defect is not** — `run-reproducer.sh` reads a non-zero exit as *defect
demonstrated* and a 0 exit as *defect not demonstrated*, so a command that exits 0 because its own
diagnostic succeeded reads as the defect already gone. The one exception is the
mutation-reproducer convention, which the mutation-testing brief below defines and a reproducer
declares with the exact `# mutation-reproducer` line: such a reproducer is read inverted, its
exit 0 being the defect present. Author it to fail pre-fix
and pass post-fix. **A runnable reproducer also declares what it demonstrates and where**: one
`# demonstrates: <path>:<line>:<content>` line per cited location, within the script's first 10
lines — the same window the `# mutation-reproducer` declaration reads — in the grep-output shape
the author pastes straight off the defect-present tree. The declaration names the file
and line the instrument reads and the content expected there, and
`check-panel-reproducer-exit-contract.sh` resolves every citation against the tree under review
before the reproducer runs: the path must stay inside the worktree, the file and line must exist,
and the content must appear on that line. A citation that does not resolve joins the inverted
class in that guard's exit 1 and is bounced like it — once, back to the raising slot. The
mutation-reproducer convention is exempt from the declaration: what such a reproducer demonstrates
is the mutated tree it builds at run time, and the `--reproducer-sha` pin below is its instrument
audit. **A runnable reproducer also declares what its checks read**: one
`# premise: <path>:<line>:<content>` line per file, test/class name or `tasks.md` task id its
checks read, in the same first-10-lines window — at least one for every runnable reproducer;
mutation-declared ones are exempt, the sha pin being their instrument audit. A premise must
resolve in every tree the script runs in, unlike the demonstrates citation, which names where
the defect was and must resolve only against the defect-present tree. The script's body asserts
every declared premise before its real checks run: a missing premise is a loud failure naming
it on stderr, exiting non-zero — never exit 0, the vacuous pass a rename must never produce
(KAN-839). Carry the premise rule on every slot's dispatch prompt. **The exemption form is available to Minor
findings only: an Important-severity finding must carry a runnable command** — one that
`check-panel-reproducers.sh` accepts and the parent can run — and the guard rejects the exemption
at Important. A demonstrating command needing a pipe, a
quote, a glob or any other shell metacharacter is written as a script rather than abandoned — the
guards refuse a metacharacter in the recorded line, never one inside a script. The slot writes it
to `<abs-worktree>/.superpowers/sdd/reproducers/<round>-<id>-<n>.sh` — `<round>` this round's
number, `<id>` the slot's own resolved reviewer id, `<n>` that slot's own 1-based finding index —
gives it a shebang and `chmod +x`, and records that same path, the path relative to the worktree —
not the `<abs-worktree>/`-prefixed form above; `run-reproducer.sh` refuses an absolute token.
Mutation writes its own into the canonical worktree, never its
`<worktree>-<slot>-<round>` copy, which is removed the moment its dispatch closes. The parent
records the path the slot supplied verbatim — there is no rename step. **The cwd contract**: the
reproducer always executes with its working directory set to the worktree the parent passes
`run-reproducer.sh` — per finding, the worktree whose copy of the recorded relative path resolves:
the canonical worktree first, else the one other entry of the state record's `worktrees` map
carrying the path, the resolution `check-panel-reproducer-exit-contract.sh` performs before its
audit and the parent's own per-finding runs repeat (KAN-658) — never the worktree the finding's own
`file:line` prefix names. A reproducer whose script resolves in one tree while the defect it
demonstrates lives in another resolves that worktree inside the
script by its absolute path and never assumes it inherited that worktree's cwd: a script authored
on that assumption fails with a wrong-project error before it demonstrates anything, and a
reproducer's first failure must be the defect, not the directory. Carry this requirement on
every slot's dispatch prompt.

**A reproducer is authored, or repaired, only with both recorded exits.** Before the finding's
reproducer line is recorded, its author has executed the script in both directions and recorded
both exits: once against a scratch worktree at the pre-fix commit, where it must read defect
demonstrated, and once against the fix, where it must read defect not demonstrated.
`prove-reproducer.sh <worktree> <pre-fix-ref> <reproducer-path>` runs both legs and prints both
exits for the record — the detached scratch worktree it materializes at the pre-fix commit, with
the script copied to the same worktree-relative path inside it, is what keeps the pre-fix leg off
the already-fixed live tree, the failure mode where a pre-fix check silently reads a fixed
worktree and proves nothing. At pass-1 authoring, where no fix exists yet, the demonstrated leg
runs against the tree the finding was raised on and its exit is recorded in the slot's report
beside the finding; at repair the pair is recorded in the fix round's report. The parent's
guards below stay as the second layer, unchanged.

**A reproducer for a finding whose defect may be target-specific is authored and run on the target
the defect manifests on — the target `flow.visual-verify`'s verifier drives, never the test
target alone.** A reported behaviour that will not reproduce on the JVM/desktop test target after
honest attempts is not thereby nonexistent: the same defect can live only in the browser/wasmJs
build the verifier drives. Author the reproducer to drive that target the way the finding
describes — real input events (`page.mouse.wheel()`), a fresh account, data seeded past the
resting state — and a finding closes on that target's evidence, never on the test target's clean
exit alone.

**Every slot's dispatch prompt also carries the CONTEXT BUNDLE paragraph** — the same shape <!-- refs-guard:allow -->
`skills/flow/implement.md`'s implementer dispatch carries; for every worktree in this run's
resolved set, naming that worktree's own whole-plan bundle
`<abs-worktree>/.superpowers/sdd/dispatch-context.md`, one path each. **For every worktree whose
`<abs-worktree>/.superpowers/sdd/relocation-comparison.md` exists, every slot's dispatch prompt also
names its absolute path**, framed as a review input to audit against the diff — never a substitute
for reading `final-review.diff` itself.

**Every slot's dispatch prompt also carries the FOREGROUND BUILDS paragraph**:

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.

**Every slot's dispatch prompt also carries the TOOLS paragraph**:

> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.

**Every slot's dispatch prompt also carries the NO DELEGATION paragraph**:

> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the parent's closed list (**Dispatch sites — the
> parent's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.

**Every slot carries the MODEL HANDSHAKE paragraph** — no exception:

> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
> system prompt>` and nothing else on that line. Answer it before any tool call.

The dispatcher compares that line against the model this slot was given and applies **The
handshake** (`skills/flow/implement.md`, **The parent orchestrates directly**), unchanged: a first mismatch
is a fallback plus one retry under `panel-<round>-<slot|slot+slot+slot>-retry`; a second is a fallback plus
the **AskUserQuestion** it states.

**Every slot's dispatch prompt also carries the REPRODUCE, DON'T READ paragraph**:

> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one check you make MUST exercise the real
> thing. A claim you did not run is worth less than one you did: a doc comment, a type signature
> and a passing test can each read plausibly and be false. Run it before you accept it, and run it
> before you reject it.

**Every slot's dispatch prompt also carries the REPORT FILE paragraph**, with the round and the
slot's resolved id substituted:

> **REPORT FILE:** write your full report to
> `<abs-worktree>/.superpowers/sdd/panel-report-<round>-<id>.md` before you end your turn — every
> finding with its `file:line`, severity, the sentence naming the defect, and its reproducer. Your
> return message is the findings summary; the file is the record.

**Every slot's dispatch prompt also carries the WORKTREES paragraph**, listing the resolved set
and, when it holds more than one worktree, the qualification rule:

> **WORKTREES:** this change spans `<abs-worktree-1>`, `<abs-worktree-2>`, …; `final-review.diff`
> is sectioned by worktree, each section headed `# worktree: <path> — merge base <sha>`. When more
> than one is listed, prefix every finding's `file:line` with that worktree's basename
> (`gymie-frontend:src/Foo.tsx:42`), and author that finding's reproducer per the cwd contract the
> reproducer rule states: it executes from the canonical worktree the parent passes the runner,
> never from the worktree its own prefix names, and resolves that worktree by absolute path inside
> itself.

**Every slot's dispatch prompt also carries the CITATION CHECK paragraph, present for every
worktree whose citation pre-check above wrote `<abs-worktree>/.superpowers/sdd/citation-check.md`,
one path each**, naming its absolute path alongside `final-review.diff`:

> **CITATION CHECK:** `<abs-worktree>/.superpowers/sdd/citation-check.md` carries this project's
> pre-panel citation scan, captured before your dispatch. It is informational — its own exit code
> was not gating — but a stale citation it reports is worth raising as your own finding if it sits
> in this diff's blast radius.

**Every slot's dispatch prompt also carries the ENTRY CONTEXT paragraph**, the round's
`[TOUCHED_FILES]` list inlined directly beneath it:

> **ENTRY CONTEXT:** `final-review.diff` is your entry artifact, and the touched-files list in
> this prompt is your named entry context — begin from those two, not from a whole-tree
> exploration. The list names every path this branch touched, per worktree, with the named
> contracts among them; read a listed file only as far as the diff and a finding require, and step
> outside the list only when a finding cannot otherwise be established. Name the coverage trade in
> your report: which touched files you read in full, which you read only at the diff's hunks, and
> which you deliberately did not read.

**Resolve `[TOUCHED_FILES]` before dispatching any slot**, once per round beside the
`final-review.diff` write above: per worktree in the resolved set, that worktree's
`git diff --name-status <merge-base>` under a `# worktree: <path> — merge base <sha>` header of
its own — the same sectioning the diff file uses. The paths under `skills/flow-contracts/` in
that list are the change's **named contracts**. A diff-reading re-run computes the list from the
delta's own range instead — `git diff --name-status <held-sha> HEAD` per worktree — so the named context
narrows with the read.

**The mutation slot's dispatch prompt also carries the MUTATION ENTRY CONTEXT paragraph**, every
test the plan's `**Tests:**` fields name and every test file in the touched-files list inlined
directly beneath it:

> **MUTATION ENTRY CONTEXT:** `final-review.diff`, at the merge-base sha each `# worktree:`
> header carries, is your diff base; the tests beneath this paragraph are the tests your
> mutations target. Begin the brief's search for a behaviour's covering test from those two —
> never a whole-tree sweep. Read a test outside the list only when the behaviour you are mutating
> cannot otherwise be judged caught or survived, and name that read in your report.

### No forking, and a wall-clock ceiling on every slot

No panel slot is dispatched onto a skill or agent that forks its own background agent — a forked
agent reports to nobody the dispatcher is tracking. Repair it by dispatching the slot on a shape
that reports back directly; never drop the slot.

Every panel slot carries a 15-minute wall-clock ceiling from its dispatch — 5 minutes for a
fix-round re-run dispatch on a decided panel, which reads a delta at `low` effort
and has no business running longer. The dispatcher
tracks each in-flight slot's elapsed time itself rather than blocking indefinitely on a completion
notification. **The ceiling is a record-and-review bound, not a stop**: it cannot halt
an in-flight dispatch — a slot the harness runs as one blocking call pays the overrun in full
before the breach can even be recorded — so a run is never promised a stop the harness cannot
deliver.

On a breach, in order: stop the slot where the harness offers a handle on the in-flight dispatch,
and where it does not, let the dispatch return and record the breach then; close its dispatch row
(`-outcome timed-out`); record the breach in the panel record, naming the slot and its elapsed
time; re-dispatch that one slot once.

A second breach of the same slot is a prompt, shape per Operator prompts
(`skills/flow-contracts/operator-prompts.md`), resolved per that contract's **Auto-resolution** on
its recommended **Stop the run**:

> **Slot `<slot>` breached the wall-clock ceiling a second time. How should this proceed?**
> - **Re-dispatch it again**
> - **Proceed without the slot**
> - **Stop the run** *(default, recommended)* — ends at `IN_PROGRESS` with the implementation
>   committed on the branch

A timed-out slot raises no finding, consumes no fix round, and is not a clean result for the final
pass.

### The mutation-testing brief

Wherever the panel dispatches Mutation, the dispatch prompt carries a mutation-testing
brief: for each behaviour the diff changes, mutate it — flip a condition, drop a guard, move a
boundary, remove a branch, move an interaction off its target, overlay an earlier commit's tree and
run the tests — and establish whether an existing test fails. A mutation only counts once its edit
landed: every mutation must confirm its edit landed before the tests run — the target changed
where the slot intended, not nowhere and not somewhere else. An edit that never applied must be
redone with a working mechanism: it is a refusal, never a **surviving mutant**, and it never buys
a test. A mutation no test catches is a
**surviving mutant**, an ordinary finding that blocks the handoff exactly as any other, unless it is
withdrawn with a reason under the fix round's handback. A surviving mutant's reproducer carries the exact line
`# mutation-reproducer` within its first 10 lines — the declaration
`run-reproducer.sh` reads as the mutation convention, since the build succeeding
with the mutation landed is the bug present, the reverse of the generic exit-code contract.

### Principles

Principles, when dispatched, is the panel's judgment check on *how* the code is built. It reads
`engineering-principles.md` — never a pasted copy — and owns the project's **hard invariants** from
its standards files.

**Resolve `[PRINCIPLES_PATH]` before dispatching the principles slot.** It is the **absolute** path
of `engineering-principles.md` **beside this file** — `skills/flow/`, always. Confirm the file
exists before spawning; if it does not, stop and report rather than dispatching a blind reviewer.

**Resolve `[STANDARDS_PATHS]` before dispatching the principles slot**, from the entries
`project-get.sh <worktree> standards` prints (exit 1: none declared), resolved per the entry-form
and containment rules of **Project configuration** (`skills/flow-contracts/project-configuration.md`),
never by reading the template. Pass an **empty** value when none resolve.
Record which standards files were passed, or that none resolved.

## Recording findings, and the record's format

**The slot writes its own report.** Each slot writes
`<abs-worktree>/.superpowers/sdd/panel-report-<round>-<id>.md` itself, per the REPORT FILE paragraph
its prompt carries — `<round>` the same value that round's findings carry on `-round` (`0` initial,
`1..n` fix rounds), `<id>` the resolved reviewer id, never the slot display name. As each slot's
`flow record dispatch end` is recorded, run **The throwaway worktree**'s (`skills/flow/review-panel-optional-slots.md`) fold-back for that slot's
copies first, then confirm the file exists and is non-empty (`test -s`); when it is not, write it
yourself carrying the single line `no verbatim report captured — <reason>`.
Every dispatched slot ends up with one, a slot that raised nothing included. **Never re-emit a
slot's report from this context** — record its `F<n>` rows and cite the file, per **Read
discipline**'s never-`cat`-a-report rule (`skills/flow/implement.md`).

**Every finding is a row in the store. The panel record is rendered from those rows.** The parent
records every finding itself. Every
finding a round raised is recorded in one Bash call, one `flow record finding` per finding:

```bash
flow record finding -change <name> -ref F<n> -round <r> -slot <slot> -severity <sev> \
  -location <file:line> -status open -reproducer <command | none — reason> \
  -dispatch-seq <seq> -note <the finding>
```

`-round` is `0` for the initial panel and `1..n` for a fix round. `-ref` is unique within the change;
the store's own constraint enforces it. **`-note` carries the reviewer's own sentence naming the
defect, not a dispatcher restatement.** Where the reviewer's wording runs long, quote the sentence
that names the defect and leave the rest to the report file. **A fix round updates the finding it
resolved rather than appending a second row:**

```bash
flow record status -change <name> -ref F<n> -status fixed
```

**`-status` carries the whole status text the marker line shows** — a withdrawal passes its reason
with it: `-status 'withdrawn <reason>'`.

**Render the record when the panel closes** — every slot's result clean, no finding open — into
the canonical worktree, never into whichever worktree the closing pass runs in, so a
multi-worktree run leaves one copy:

```bash
flow record render -change <name> -kind panel -repo <canonical-worktree>
```

**A panel that raised nothing still renders**, declaring `findings-total: 0`. There is no
skip-the-render shortcut.

The record carries a findings table, one row per finding:

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Mutation | Minor | `src/Foo.kt:42` | replaced the silent catch |

and, below it, the marker block — one line per row, plus the count:

```
findings-total: 1
finding-status: F1 fixed
```

The marker format is never quoted inside the record itself. The renderer neutralises any marker
label a finding's note or location happens to carry, on the way out only.

**The reproducer each finding's slot supplied gets a marker block of its own**, separate from the
`finding-status:` block above:

```
reproducers-total: 1
finding-reproducer: F1 scripts/test-check-plan-unchanged.sh
```

A finding recorded with no reproducer renders the `none — <reason>` exemption form.

**The table carries no status column, on purpose.** To read a finding's state, look up its `F<n>`
in the marker block.

**A `withdrawn` marker's reason is checked for being there at all.** A finding is recorded `fixed`
by the parent at the fix round's verification step below, never by the fix subagent.

## Panel re-runs

**Every round this stage dispatches after pass 1 — each fix-round re-run — opens by running **Check base movement first** again: the same
per-worktree `resolve-base-branch.sh` then `check-base-moved.sh` pair, with that section's
verdicts, automatic no-overlap rebase, overlap prompt and conflict handling unchanged — a `MOVED`
with no overlap rebases unasked at a round boundary just as at entry.** **Continue** at a round
boundary proceeds into the round's own remaining steps — the citation pre-check that follows the
entry check is an entry step, and the round does not re-run it. The entry check ran once,
before pass 1; a base that moves while earlier rounds ran would otherwise reach the final round
— and then integrate — unchallenged, its conflict surfacing only after review has closed. A
conflict found here surfaces while the panel is still active and the operator is already
engaged.

**Pass 1 runs the roster **The docs-only reduction** or **The late-fix reduction** chose — the
resolved roster, `primary` alone on a docs-only branch, or `primary` alone on the late-fix delta
where the reduction's trigger holds — plus every slot the operator named at this stage's start
that it did not already carry.** Only re-runs after a fix are scoped. Record
`FIX_BASE` — the branch tip the fix round starts from — commit the fix, then write
`<abs-worktree>/.superpowers/sdd/fix-round-N.diff` from `git diff "$FIX_BASE"..HEAD`.

**A branch the remote already holds takes the fix as one new commit on top, never a rewrite.**
This is the normal case: the branch is pushed with every commit (**Branch backup**,
`skills/flow-contracts/git-boundaries.md`), so the fix stages the changed paths
(`git add -- <the changed paths>` — a pathspec commit reads tracked paths only) and is a plain
`git commit -m ... -- <the changed paths>` at the tip — the pathspec-scoped default (**Branch backup**, `skills/flow-contracts/git-boundaries.md`),
so it carries only the paths the finding named, whatever else the index holds — pushed
plain like any other commit, and every downstream commit keeps its sha.
**Rewrite-based folding is for unpushed history only**: the fixup — stage first
(`git add -- <the changed paths>`), then `git commit --fixup=<task-sha> -- <the changed paths>`,
scoped by the same default — targets the **original** task
commit (`<task-sha>`), and the autosquash folds it in immediately, before anything pushes. That
rebase's upstream is the parent of the task commit it targets (`<task-sha>^`), never the base
branch re-resolved — replaying only the branch's own commits after `<task-sha>^` is what keeps
commits that landed on the base mid-panel out of the fold, where an `origin/$BASE` upstream would
carry them straight into the round's delta. The route's own diff is
`git diff "$FIX_BASE"..HEAD`, read once the fold has landed — the same held pre-fix sha to HEAD
as the plain route; the fold rewrites `<task-sha>` in place, so no diff endpoint is ever
re-resolved, and a movable base can never leak into the delta. The fold is bracketed by the
ancestry guard: `guard-autosquash.sh targets <worktree> <task-sha>` before the autosquash,
`guard-autosquash.sh after <worktree> <task-sha>^ <changeRoot>/tasks.md` once it lands — the
post-check's base is the fold's own upstream, never `FIX_BASE`, which the fold deliberately holds
stale as its diff endpoint, and a refusal from either stops the round before anything builds on
the rewritten history (`<agents repo>/scripts/guard-autosquash.sh`). A conflict there is between
two of the branch's own commits, resolved by hand, keeping both sides — the resolve-in-place rule
of a base-branch rebase (its Conflict case, under **Sync the branch onto the base**, `skills/flow-contracts/finish-contract-run1.md`) concerns the operator's base, never this one. The
fold never crosses the run's own uncommitted planning edits —
`aside-planning-artifacts.sh <aside|restore> <worktree>` around the rebase: set aside before it,
restored once it has finished or aborted, never mid-way; restore refuses while the rebase is
still unresolved, and the paths it sets aside are the spec tree's changes directory — the leaf
`<agents repo>/scripts/lib/spec-root.sh` resolves — and `<project>/docs/superpowers/` only, never
implementation WIP.

**A clean `git rebase --autosquash` is not evidence the fix survived it.** Where the fixup and the
commit it folds into touch nearby lines, git's 3-way auto-merge can resolve in favour of the
pre-fix side — it exits 0, prints no conflict marker, and leaves no `fixup!` commit behind. The
reproducer rerun and diff check below (**Once the fix subagent reports…**) are what catch this;
they must run against the post-rebase file content, never be satisfied by the fixup commit's
presence or the rebase's own exit code.

**A fixup whose fold empties its target commit is dropped, never kept as a no-op commit.** When the
fold leaves the task commit it targeted with an empty diff against its own parent, the fix exactly
undid what it folded into: `git reset --hard <the folded commit's parent>` takes the pair off the
unpushed branch, and the plan's task entry stays — the record of the mistake the branch no longer
carries. `git diff "$FIX_BASE"..HEAD` then reads the restoration itself, and the round's reproducer
re-runs and diff check close on it unchanged.

**Which slots re-run, and on what, follows from the severities the round raised — never from a
mode table, a trigger list, or a round count.**

**A Critical is confirmed against a primary source before any fix round rewrites working code on
it.** The confirmation is the parent's, at acceptance: where the finding's premise is what something
outside this change does — a tool's flag semantics, a library's behaviour, a published format — the
parent checks that premise against the thing's own published documentation or source, never a
secondary summary, and records the confirmation with the round's pass log. A premise the primary
source does not support withdraws the finding with that reason through the round's auto-decisions
below, and the working code stands.

**Every Critical and Important goes to the fix; whether a Minor does follows from the rest of the
round, and a Minor never causes a fix round of its own.** Every Critical and Important the round
raised goes to the fix subagent below, closed by the verification that follows it — the reproducer
re-run exits 0 *and* the fix diff touches a path the finding named, or one of the path condition's
accepted alternatives below closes the finding. **A round that raised a
Critical or Important sends every Minor it raised to that same fix**, closed by the same
verification. **A round that raised no Critical and no Important defers every Minor** — a fix round
costs more time than its Minors are worth — with
`flow record status -change <name> -ref F<n> -status 'deferred <reason>' -category <doc-only|pre-existing|cosmetic|coverage-gap|out-of-scope|other>`,
the category naming the mechanism the reason clause states, so the deferred-Minor rate is a query
rather than a hand-read. Nothing in that round is fixed, inline or otherwise, and no slot re-runs:
proceed to **Deferred findings go to KNOWN-BUGS.md at round close**, below, and then to
`check-panel-findings-closed.sh` and the stage close. A fixed finding that fails
verification takes the handback below, and that loop re-runs no slot either.

**A deferral's reason is one clause naming the mechanism — never a rationale essay, in the store
row or in the round's output.**

**A rule that changes mid-run governs from the round it lands in, and each round's pass log names
the rule it decided under.** When the wording of a rule this file states — the Minor-deferral
default above included — changes while a panel is in flight, by an operator instruction or by an
edit to this file, the new wording governs from the first round decided after it lands: a round
the old rule already closed stands as closed, never re-decided retroactively. Each round decided
under wording that changed during the run records that wording beside its decisions, with the
round's own pass-log rows — `flow record pass -change <name> -round <round> -note 'decided under:
<the rule as this round applied it>'` — so the rendered pass log shows which rule governed what.

**When the round raised anything above Minor, re-run on deltas.** A slot's last-reviewed sha is
held **per slot per worktree**: each dispatch sets that slot's sha in every worktree to the HEAD it
was dispatched against, and a slot not dispatched in a round keeps the shas it had. A delta is
`<abs-worktree>/.superpowers/sdd/slot-delta-<round>-<slot>.diff` (the canonical worktree's), combined with the same
per-worktree sections as `final-review.diff`, without its semantics first line — one `# worktree:`
header per worktree, followed by that worktree's `git
diff <held-sha> HEAD`; a worktree in which the slot holds no sha contributes its whole `git diff
<merge-base>` section. Every slot's dispatch prompt names the path it was given and, for a delta,
each worktree's starting sha. **Check base movement first** above clears every slot's held sha in
the rebased worktree on a clean rebase, whether taken at panel entry or at a round boundary, so
that worktree's section falls under the no-held-sha rule in the next round. Then:

- **a slot re-runs only when it raised a Critical or Important in the previous round, or the
  previous round raised a new Critical** — every slot in the resolved roster, Primary included,
  and every operator-added slot already dispatched in an earlier pass of this run, on that one
  rule. A Minor, fixed or deferred, re-runs no slot: a fixed Minor closes on the verification
  below alone. A slot that raised nothing keeps the result it has;
  the round's own mutation-proof (below) covers what the fix changed;
- **a diff-reading slot that re-runs reads its delta**; Mutation reads no diff
  file and re-runs in its pass-1 shape, throwaway worktree included. **A diff-reading slot whose
  delta is empty in every worktree is not dispatched** — record `not re-run —
  nothing new since its last read` with `flow record pass -round <round>`;
- **a slot the operator has not named for this run is never added here** — that addition happens
  only through the explicit-request check **The roster** states, at the start of any round;
- **on a decided panel, each re-running role runs alone, in its own dispatch, on
  the decision's `panel.rerun_dispatch` pair**, under the 5-minute ceiling — never bundled with
  another role. **The re-run is targeted at what that role raised and nothing else**: in place of
  its held-sha delta it reads the round's `fix-round-N.diff` plus the sites of its own open
  findings, each opened at its recorded `file:line` in the current tree, and its prompt lists
  those `F<n>` rows verbatim, states that it is re-reviewing their fix, and names that diff path.
  Its verdict is per listed finding — fixed, or not fixed with the reproducer output — plus any
  defect the fix diff itself introduces at those sites.
  A role that raised nothing in the previous round does not re-run under the new-Critical clause
  above either — with no finding of its own to target it has nothing to re-review. Each
  `-model`/`-effort` recorded is the rerun pair (**Bundled dispatch**).

**A fix-round re-run reviews the fix, never the branch** — and no pass after pass 1 re-reads the
whole branch; the paragraph below is the one statement of a re-run's read scope. Every re-running slot's dispatch prompt —
on every panel, Mutation included — carries the FIX-ROUND SCOPE paragraph,
with `<fix report>` the round's `panel-fix-report-<round>.md` (every chunk's file on a chunked
round):

> **FIX-ROUND SCOPE:** this is a fix-round re-review. Its scope is the fix diff you were given
> and the sites of the findings it fixes, plus at most the code neighbouring those hunks — the
> enclosing function or section — and nothing else on the branch. Run only the tests and
> specs that diff touches, never the module, repository or live-spec suite. The fix report at
> `<fix report>` carries each fixed finding's proof: the test run before the fix and after it,
> and the failure with the fix reverted. Check that proof against the diff; reproduce it
> yourself only where it is missing, does not match the diff, or does not show the failure it
> claims — and say which of those it was.

**From a change's third fix round on, a fix round is scoped.** A re-running diff-reading slot
reads the round's `fix-round-N.diff` plus the sites of every finding an earlier round raised —
each site opened at its recorded `file:line` in the current tree — in place of its held-sha
delta; the re-run rule above is unchanged, and so is everything the round's own mutation-proof
covers. The scoping exists because a round that re-reads a growing fix diff regress-checks by
volume, not by site.

**The cap check on a re-run** is `check-panel-diff-size.sh <worktree> <sha> <cap>` once per
worktree per **distinct** held sha among the diff-reading slots dispatched this round (two slots
sharing a sha in a worktree need one call there, not two); a slot with no held sha in a worktree
counts from that worktree's merge base. **The gating count is the largest per-slot sum across
worktrees** — the largest single combined read any one slot this round faces — and an exit-1
result from a call contributing to it proceeds unasked exactly as the pass-1 check does (**The
roster**, above), reporting the gating sum, its per-worktree counts and, when it differs, the
full-branch sum. Record both with `flow record pass -round <round>` for this round,
alongside the agents-ran/why/diff-path lines the fix pass records.

**The docs-only guard runs again beside that cap check**, `check-panel-docs-only.sh <worktree>
<merge-base>`. A branch that stays docs-only keeps the
reduced roster, and `primary` re-runs on its delta as above. A branch the fix round made no
longer docs-only — exit 1 or 2 where pass 1 saw exit 0 — dispatches, in this round, every
resolved slot not yet dispatched this run, each reading the whole `final-review.diff` under the
no-last-reviewed-sha rule above; Mutation among them takes its pass-1 shape, throwaway
worktree included. Record which slots joined this way and the path the guard printed.

Handoff still requires **zero open findings at any severity** from every agent that has run, and
no stale result — where **a slot's clean result is stale when the rule above required that slot to
re-run and it has not, or when any commit or working-tree change to source landed after that slot's
last read, from any stage — `flow.verify` included**. An unrecorded edit after the panel closes is
stale by definition, not only one a fix round produced. A fix against which a slot raised no
finding leaves that slot's result current: the round's own mutation-proof (below) covers what the
fix changed. The same holds for a delta the late-fix reduction's targeted dispatch read — clean,
or with every finding it raised a deferred Minor under the standing rule:
`primary`'s read on the since-close range leaves every slot it did not dispatch current —
the reduction's own conditions, already clean and already verified with the machinery untouched,
being what the full roster's coverage rests on for that delta.

Union all **open** findings, dedupe by **defect identity — file:line + theme.** *File:line* is the
finding's own recorded location, taken verbatim from the findings table. *Theme* is the finding's
one-sentence Note column, reduced to its own defect noun phrase — the shortest phrase naming what is
wrong, severity words and slot names stripped out.

**Before dispatching the fix subagent**, the parent runs:

```bash
check-panel-reproducers.sh <worktree> <change>
```

Exit 0 proceeds. Exit 1 covers two classes: a **missing or malformed field** is added before
dispatch; a **rejected reproducer shape** (a shell metacharacter, an absolute path, a `..` segment,
a leading `-`, a URL, a NUL byte) is a **refusal** — the line is recorded **unverifiable** and put to
the operator, never silently rewritten. Exit 2 stops the run. Both guards first read the change's
state record through the worktree argument, and a store-reached-but-absent record is exit 2 — on a
cross-repo change the record resolves only through the canonical worktree's project key, so that
shape is what invoking a guard on a peer tree reads as, and the empty findings array a peer tree's
store read would otherwise answer is never a clean verdict (KAN-658).

**The exit-code contract is checked mechanically before any dispatch decision reads a reproducer by
hand**:

```bash
check-panel-reproducer-exit-contract.sh <worktree> <change>
```

The guard runs every **open** finding's runnable reproducer through `run-reproducer.sh` against the
finding's own worktree — per finding, the worktree whose copy of the reproducer's recorded relative
path resolves: the canonical worktree first, else the one other entry of the state record's
`worktrees` map carrying the path; a path resolving in several recorded worktrees is cannot-answer,
its tree genuinely unchoosable (KAN-658) — bare, and requires the verdict *defect demonstrated* —
the exit-code behaviour an open
finding's reproducer claims on the tree under review. Before anything runs, the guard audits the
instrument itself: each runnable reproducer carries the `# demonstrates:
<path>:<line>:<content>` declaration its authoring rule above requires, within the script's first
10 lines, and the guard resolves every citation against the finding's resolved worktree — the path
stays inside the
tree, the file exists, the line exists, the content appears on that line. A reproducer whose
citation does not resolve, or whose script cannot be read to audit, is never run — a verdict spent
on an unresolvable instrument is the green flip the audit exists to deny. The audit also resolves
every `# premise:` declaration the same way, tolerantly: absence of premise lines violates
nothing, and a declared-but-unresolvable premise joins exit 1's violation classes;
mutation-declared reproducers skip it as they skip the demonstrates audit (KAN-839). Findings at any other
status claim nothing about the current tree and are skipped, as are the exemption and bare-`none`
forms the lexical guard above owns, and a mutation-declared reproducer skips the audit: what it
demonstrates is the mutated tree it builds at run time, and the `--reproducer-sha` pin below is
its instrument audit. Exit 0 proceeds to the per-finding runs below. Exit 1 names every finding
whose reproducer contradicted its claim, one disposition per class: a reproducer that read *not
demonstrated* — the inverted class —, one whose demonstrates citation did not resolve, and one
whose script cannot be read to audit at all, are
bounced exactly as the per-finding run's own answer 1 below,
once, back to the raising slot; a reproducer the runner refused as unusable is recorded
**unverifiable** and put to the operator, exactly as the per-finding run's own answer 2 below, since
a refused reproducer never ran and so carries no passing output a bounce could carry. Exit 2 — any
reproducer the runner could not verdict: a timeout, a surviving process, a plumbing failure, the
state record absent or unreadable, an ambiguous worktree — stops
the run, the same as the lexical guard's exit 2.

**For each open finding whose record carries a runnable `finding-reproducer:` command**, the
parent runs it — every finding's run in one Bash call (each throwaway copy is already
gone, removed as its slot's dispatch closed), each run followed by `; echo "F<n>: exit $?"` so every exit code stays
readable:

```bash
run-reproducer.sh <worktree> "<the finding's finding-reproducer: text>"
```

`<worktree>` is the cwd contract's per-finding resolution — the worktree whose copy of the
recorded relative path resolves, canonical first, else the one recorded worktree carrying it
(KAN-658) — the same tree the exit-contract guard above already ran this reproducer in.

Read its exit code: **0** dispatches the finding; **1** bounces it once, back to the raising slot,
carrying the reproducer's passing output; **2** is a refusal — recorded **unverifiable**, put to the
operator; **3** is a timeout or a detached survivor — recorded **unverifiable**, put to the
operator, with a surviving pid named when the script names one; **4** stops this finding's dispatch
decision entirely — a finding recorded `none — <reason>` is dispatched without a run.

A slot that supplies nothing for a finding has not supplied a legal exemption: record that omission
as its own open finding.

**Once the fix subagent reports, re-run every dispatched finding's reproducer** under the same
constraints, carrying `--pre-fix-verdict <that finding's dispatch-time verdict>` — `demonstrated`
or `not-demonstrated`, the verdict that finding's dispatch-time run printed (its recorded exit
code names it: 0 means `demonstrated`, 1 `not-demonstrated`) — **and `--reproducer-sha <the sha
that run printed>`**: the verdict comparison is valid only between two runs of the same file, so the
runner pins the re-run to the dispatch-time reproducer and refuses a mismatch (exit 2, never
executed). A refused re-run means the reproducer was re-authored — its mutation-convention
declaration included — and it is re-run against the defect-present code — the pre-fix leg
`prove-reproducer.sh` runs, against a scratch worktree at the pre-fix commit — for a fresh
verdict and sha before the round continues: a run that answered `2`, `3` or `4` refused, went
unverifiable,
or stopped before any dispatch, so nothing re-runs for it — and
require the reproducer now to exit **0**, which the script answers **1**. The flag makes the
script refuse (exit 2) a reproducer whose verdict here is identical to its pre-fix verdict —
ambiguous under either exit-code convention, the expected one named in the script's message — so
an ambiguous reproducer is recorded unverifiable and put to the operator rather than read as an
unfinished fix. **The parent runs these re-runs itself, in its own
Bash calls** (**Dispatch sites — the
parent's closed list**, `skills/flow/implement.md`). **The flip alone does not close a finding — the fix's
diff must also touch at least one path the finding named, with a non-comment, non-whitespace
change.** A fix that meets neither that match nor one of the three shapes below is not a fix: the
finding stays open and goes through the handback below. **Beside that match, three shapes close on their own evidence, each judged from the same
fix diff or repaired history the match reads, never from the fix subagent's report alone:**
(a) **a test-only fix** — every non-context change in the diff is in a test file, a file the
project's `## test` commands run or that only test code imports — closes on the
flip the re-run already required, when the finding's pinned reproducer flipped demonstrated → not
demonstrated: the defect was a missing or weakened test, and the diff that adds or strengthens it
is the fix; (b) **a comment-only diff closes a comment-only finding** — the finding's defect is a
stale or wrong comment and the diff's change to the named path is comment-only, a repair that can
never carry the non-comment change the match demands; (c) **a history repair closes a finding
located in a commit record** — the defect lives in a commit rather than in a file, and the fix
amends, fixup-folds or rewrites the named commit under the rewrite rules above, closing on the
commit as it now stands, read post-repair: no diff-path evidence exists for a defect in history.

A re-run whose script fails a premise assertion — the loud non-zero the authoring rule requires —
reads *demonstrated*, identical to the dispatch-time verdict, and is refused as ambiguous by
design (KAN-839): the renamed premise voids the reproducer, and the refusal routes it to
re-authoring. That refusal is the fix working, never a runner bug.

A finding meeting both conditions is recorded closed there and then — a Minor recorded
`none — <reason>` has no reproducer to flip and closes on the path condition alone, or on its
comment-only alternative when the finding's defect is a comment and the diff's comment-only
repair touches that named path:

```bash
flow record status -change <name> -ref F<n> -status fixed
```

**The parent records it, never the fix subagent.** **Record every verdict this turn reached in one call,
never deferred to the round's end** — the reproducer re-runs and the fix diff are read in one
call, each finding judged, then every `status fixed` recorded together, so an aborted round
still leaves every already-verified finding closed. **A finding is recorded `fixed` only after
the re-run that verifies it — never in the same Bash call as that re-run.** The call that runs
the re-runs prints each exit and is read; a separate later call records `status fixed` for
exactly the findings it verified, and no others. A `fixed` recorded in the same call as its
re-run stands verified before that exit is known — the ordering KAN-582's F1 rode in on, luck
rather than discipline — which is the defect the close guard's fixed-without-clean-rerun class
below exists to catch. **A finding failing
the match and all three shapes is left untouched** on `open`, for the handback below. This walk follows
**Read discipline** (`skills/flow/implement.md`): the specific hunks a finding names, never the <!-- refs-guard:allow -->
whole fix diff.

### The fix round mutation-proves what it changed

**Every executable behaviour the fix changed is mutation-proved, not only the test cases the round
adds.** The fix subagent performs the proof and reports it, per the MUTATION PROOF paragraph its
dispatch carries; the parent runs no build of its own here. A survivor the fix subagent cannot
judge real or equivalent goes through the same handback the section already names.

**A fix that adds or strengthens a test is proved by a flip of the fixed line itself, one flip per
fixed finding.** The mutation flips the line the fix changed — the code the added or strengthened
test exists to catch — and the test confirmed to fail under the flip is that added or strengthened
test, not merely any test that happens to trip: the flip is what shows the fix's own test tests
what it claims. The flip is recorded with the round's other `flow record mutation` rows before the
round closes.

**Record each one with one `flow record mutation -change <name> -round <round> -path <path>
-mutated <what> -test <test>` call per line, transcribed from the fix subagent's report — the
exemption form records `-mutated none -test <reason>`; the lines the calls write are:**

```text
fix-mutation: <path> — <what was mutated> — <the test that failed>
fix-mutation: <path> — none — <reason>
fix-mutations-total: <n>
```

**The rendered record carries these lines in its pass-log section, one round's count after that
round's lines, and never inside the marker block.**

**A fix to a guard script is proved by the guard's own observable, measured on both sides of the
fix.** A guard script is a check that exists to fail on a defect class — this project's
`scripts/check-*` family, or wherever its `## lint` section names its guards. The mutation probe
is the guard run itself against the defect state: on the pre-fix code the guard reports the
defect — the probe fails — and on the post-fix code it does not — the probe passes. The
`fix-mutation:` line for that behaviour carries the measured pre/post observable as its third
field, `<pre>→<post> <what the observable counts>`
(`2→0 orphaned temp lists`), never a bare test name.

**The parent checks the reported list against the fix diff before the round can close, reading the
`fix-mutation:` lines and walking the diff itself.** Walk every
hunk of the fix diff with a non-comment, non-whitespace change: each one is either covered by a
reported line, or is not an executable behaviour at all. A fixed finding whose fix added or
strengthened a test closes only with its own flip in the pass log — the per-finding obligation
above is checked beside this walk, and a finding short its flip stays open for the handback below,
never closed on another finding's line. A hunk that removes or weakens a test or an
assertion states in the record what it used to cover and names what still covers that same behaviour
now — checked by running the named covering test against the **pre-fix** code and confirming it
fails. A hunk whose path is draw/geometry code is held to the **PIXEL PROBE** paragraph
beside the mutation lines: the fix's report names the probe assertion it landed against the
actual rendered pixels or geometry, and a fix commit for this class of bug that landed no such
probe is rejected at the fix step — the round does not close and the finding stays open, for the
handback below — rather than the regression being discovered next round. **The same
walk holds the fix subagent to its PLAN FIELDS obligation:** a hunk that adds a
test case, adds a file, or changes what a task's `**Baseline:**` counts, whose task's `**Tests:**`,
`**Baseline:**` or `**Files:**` field in `<changeRoot>/tasks.md` does not reflect it, does not close
the round; it goes to the handback.

Beside the reproducer re-runs above, the round close re-runs the task-field guard mechanically:
`check-task-commit-fields.sh <worktree> <task-id> <task-sha> "" <canonical-worktree>
<name>` for every task a fixup folded into — `<task-id>` from that commit's `Task-Id:` trailer,
`<task-sha>` the folded commit as it now stands, the remaining arguments resolved the way
`skills/flow/implement.md`'s task-close step resolves them. Exit 1 does not close the round; it
goes to the handback. Exit 2 — the guard's not-a-verdict close (the same course
`skills/flow/implement.md`'s task-close step gives it) — stops the run: a handback cannot repair
an inability. This catches an undeclared file the fixup added and a declared
test it removed or renamed, read post-autosquash. The stale-field classes are the guard's verdicts: a task declaring `**Baseline:** before=N
after=M` fails when the changed files' `@Test` delta at the commit does not measure it — skipped
where the counted set carries no `@Test` at either revision, per the skip-not-fail rule — and a
`**Tests:**` name, backticked or bare camelCase, that the tree's content at the commit no longer
contains fails with it. A test added to the commit with no `**Baseline:**` declared and no
`**Tests:**` naming it stays the walk's judgment.

**The round close runs the project's configured build-green guard too, when the project declares
one (**The guard's scope**, `skills/flow-contracts/build-green.md`), over the plan's `tasks.md` —
exit 1 does not close the round; it goes to the handback, and exit 2 — the same not-a-verdict
close — stops the run.** This is the gate that holds a
plan appended to mid-run to the tags it published under: tasks a fix round appends carrying
`**Build:** pending` keep the round open until every tag reads `green` or `red`.

This binds the fix round every run — the obligation is the round's, not a slot's, so a run where
Mutation is not in the resolved roster or added this run is exactly where the round's own proof
is the only mutation reasoning that happens at all.

**Rebuild the dispatch context bundle before dispatching the fix subagent**, same as above,
overwriting the same path. Report the script's stderr line (`bundle unchanged — reusing …` or
`bundle rebuilt — …`) as part of this round's own reporting.

**Carry each surviving finding to the fix subagent as a structured block**, not a bare restatement
of its prose: its `F<n>`, the slot that raised it, its severity, its `file:line`, its theme, the
text of its `finding-reproducer:` line, its slot's report path, and any bounce already recorded
against its defect identity. **Inline no source excerpt.**

**Every fix subagent's dispatch prompt also carries the VERBATIM REPORT — THE FACT paragraph**:

> **VERBATIM REPORT — THE FACT:** each finding below names the file its slot's report was written
> to. That file is the reviewer's own report, unedited — read it before you act on the finding.
> The structured block is the dispatcher's summary of it: direction on what to work on, and
> never a source of fact. Where the two disagree the report wins. Where the block asserts
> something the report does not, treat it as unchecked and establish it yourself before building
> on it.

**Every fix subagent's dispatch prompt also carries the FINDINGS ARE INPUT paragraph**:

> **FINDINGS ARE INPUT:** a finding names a defect and suggests a route to fixing it — findings
> are input, not orders. Your obligation is to resolve the defect the finding names; the
> suggested route is the raising slot's proposal, never a binding instruction. Where the literal
> route would break something the code already requires — a required ordering, a declared
> contract, a stated invariant — take the route that resolves the defect without the damage, and
> your report records the deviation and justifies it: what the literal route would have broken,
> and how yours resolves the defect. A silent deviation is an unfixed finding, and so is a
> literal compliance that leaves the defect standing.

**Every fix subagent's dispatch prompt also carries the PLAN FIELDS paragraph**:

> **PLAN FIELDS:** your fix owns its plan record, as part of the fix itself: when it adds a test
> case or changes what tests a task names, update that task's `**Tests:**` field; when it changes
> what a task's `**Baseline:**` counts or `**Files:**` paths declare — a test case added, a file
> created — update those fields too. All of it lands in the worktree's
> `<project>/spectre/changes/<name>/tasks.md` in this same pass — never left for a reviewer to
> catch next round. Edit them; do not stage or commit them — the plan record is
> a planning path and is committed later by the pipeline, never in a fixup.

**Every fix subagent's dispatch prompt also carries the ROUND SCOPE paragraph**:

> **ROUND SCOPE:** your round's output is the diff and the report, nothing else. Do not rewrite
> `proposal.md` or `design.md`: unaffected sections are never restated, and a decision your fix
> genuinely overturns is superseded by one appended entry under `## Decisions`, never a per-round
> rewrite of the file. Deferral rationale is written nowhere — a Minor is deferred with its
> one-clause reason in the store. Your report names what you fixed, the behaviours you changed,
> and the task commits your fixups folded into, then stops.

**Every fix subagent's dispatch prompt also carries the FOREGROUND BUILDS paragraph**:

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.

**Every fix subagent's dispatch prompt also carries the TOOLS paragraph**:

> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.

**Every fix subagent's dispatch prompt also carries the NO DELEGATION paragraph**:

> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the parent's closed list (**Dispatch sites — the
> parent's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.

**Every fix subagent's dispatch prompt also carries the MODEL HANDSHAKE paragraph**:

> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
> system prompt>` and nothing else on that line. Answer it before any tool call.

Dispatched on `DEFAULT_MODEL` (below); the dispatcher compares that line against it and applies
**The handshake** (`skills/flow/implement.md`, **The parent orchestrates directly**), unchanged: a first
mismatch is a fallback plus one retry under `panel-fix-<round>[-<chunk>]-retry` — the only
retry key `check-panel-fix-single-dispatch.sh` accepts; a second is a fallback plus
the **AskUserQuestion** it states.

**Every fix subagent's dispatch prompt also carries the TARGETED TESTS paragraph**:

> **TARGETED TESTS:** Run only the tests the `**Tests:**` fields of the tasks your fix folds into
> name, and the tests your fix adds, through the build
> tool's own selector — `--tests '<class>'` for Gradle, `-run '<name>'` for `go test`, `-t
> '<name>'` for vitest — once for RED, once for GREEN, and again only after a source edit. Never
> run the module or repository suite mid-task: the full `## test` list runs again in
> `flow.verify`. Pipe a test run's output through `tail` so a green
> run costs lines of context, not a build log.

**Every fix subagent's dispatch prompt also carries the OUTPUT BUDGET paragraph**:

> **OUTPUT BUDGET:** Every tool result stays in your context, and every later turn re-reads your
> whole context — a large output is paid for again on every turn after it. Read a file over 200
> lines by line range — `grep -n` for the symbol, then `sed -n '<a>,<b>p'` or Read with
> `offset`/`limit` — and never re-read a file already in your context unless you have edited it
> since. Cap every search (`| head -40`) and every build, lint or install run (`| tail -30`), and
> reproduce a failing block from its log rather than printing the whole log. Never print a
> generated file — a lockfile, a snapshot, a bundle, a build artifact.

**Every fix subagent's dispatch prompt also carries the MUTATION PROOF paragraph**:

> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you
> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a
> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit
> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that
> never applied is a refusal, not a surviving mutant — redo it with a working mechanism; it never
> buys a test. Then confirm an existing test fails, and restore. Where your fix adds or
> strengthens a test, the mechanism to mutate is the fixed line itself — flip the line your fix
> changed, the code the added or strengthened test exists to catch — and the test confirmed to
> fail under the flip is that added or strengthened test: one flip per fixed finding, the proof
> that your fix's own test tests what it claims.
> `mutate-and-verify.sh <patch-file> <harness>` mechanizes backup, apply, run, report and restore
> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to
> mutate is your judgment, not the script's. Each mutation alters one mechanism — where a single
> revert would also change state a second check reads, split it into surgical mutations, one per
> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before
> your turn ends. Where your fix changed a guard script — a check that exists to fail on a defect
> class — the mechanism to mutate is the fix itself: run the guard against the defect state with
> the fix reverted, where it must report the defect, and with the fix applied, where it must not;
> the line records the measured pre/post observable, not a bare test name. Record one
> `fix-mutation:` line per behaviour in your report, plus a
> `fix-mutations-total:` count, in exactly this shape: `fix-mutation: <path> — <what was mutated> —
> <the test that failed>`; `fix-mutation: <path> — none — <reason>` for a behaviour you exempt; for
> a guard-script fix the third field is the measured `<pre>→<post> <what the observable counts>`
> (`2→0 orphaned temp lists`); and one `fix-mutations-total: <n>` line after them. Where
> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather
> than deciding it yourself.

**Every fix subagent's dispatch prompt also carries the PIXEL PROBE paragraph**:

> **PIXEL PROBE:** every fix you land to draw/geometry code — code that computes what is
> drawn: extents, bounds, gridlines, offsets, paths, positions, sizes — carries a probe assertion
> against the actual rendered pixels or geometry, in the style the reviewers' reproducers
> already use: drive the real draw path and assert on the value it renders or computes, never on
> a hand-derived copy of it. A fix commit for this class of bug that lands no such probe is
> rejected at the fix step — the finding stays open and the round does not close — so the
> regression is caught this round, not discovered by the next one. Where pixels cannot be
> rendered in this environment, the probe asserts on the geometry the draw path actually
> computes — the same numbers the renderer will paint — and your report names why the pixel
> output itself was not asserted.

**Every fix subagent's dispatch prompt also carries the REPORT FILE paragraph**:

> **REPORT FILE:** write your report to
> `<abs-worktree>/.superpowers/sdd/panel-fix-report-<round>.md` — a chunked round's chunk `<n>`
> appends its own suffix, `panel-fix-report-<round>-<n>.md` — as your **last** act — after
> the rebase and your final test run — naming each finding you addressed, the executable
> behaviours your fix changed, the `fix-mutation:` and `fix-mutations-total:` lines this round's
> contract requires, and the task commit each fixup folded into. Each fixed finding carries its
> proof, each run's command and the tail of its output: the test run before the fix, the same
> run after it, and the failure that run shows with the fix reverted — the re-running slot checks
> this proof rather than re-deriving it. The dispatcher waits on that file's presence.

Give the surviving findings to fix subagents in **chunks of at most 10 findings**. The round's
findings are split into sequential chunks of at most 10 — the cap that keeps one dispatch from
re-reading the whole branch across every finding — each chunk one panel-fix dispatch carrying its chunk of
the combined list. Never one dispatch per reviewer, per slot, or per finding — that split
fragments one diff into competing fixups against the same worktree. Chunks run in order, each
dispatch awaited in the foreground before the next chunk begins and before the round's reproducer
re-runs begin, and no fix subagent is left in flight when the turn ends. The round's first
chunk's `-key` is exactly `panel-fix-<round>`; each further chunk appends `-<n>`
(`panel-fix-<round>-2`, `-3`, …), contiguous from 2; the handshake retry suffixes `-retry` onto
its own chunk's key (`panel-fix-<round>[|-<n>]-retry`), and any other key shape is a violation
whatever the dispatch count. A round may not chunk freely: its chunk count is bounded by
`ceil(findings raised in earlier rounds / 10)`. Before recording each `dispatch begin`, confirm that key has
no panel-fix begin already recorded — a second begin under a fresh key is over-dispatching even
when every key is well-formed. `check-panel-fix-single-dispatch.sh` holds every run's panel close
to exactly this shape (**Before closing the stage**, below). Inline
(`skills/flow/implement.md`'s **Inline — the parent implements**), the parent applies the fix
itself under the same paragraphs, dispatching no subagent, and records the pass with `-role
panel-fix -model <parent model> -effort <parent effort> -agent-id inline`, `begin` before its first
edit and `end` once the fix commit lands, as that section's **Records** bullet states. Where a finding is
confirmed as a real defect, the fix subagent invokes **superpowers:systematic-debugging** before
writing its fix. **Dispatch it on `DEFAULT_MODEL`**. **On an `sdd` decision, dispatch it instead on the decision's
`fixer` object** — its own model and effort, chosen apart from the implementer's, `subagent_type:
flow-<effort>` with that `model` passed as the Agent tool's own `model` parameter, and
`-model`/`-effort` below carry that pair. On harness `zcode` the pair given and recorded is `glm-5.3-flash` / `high` instead (**Harness mapping**, `skills/flow-contracts/model-policy.md`). Record
every pass with `flow record pass -round <round>`: which agents ran, why,
the diff path they read, and — when this pass bounced any finding — each bounced finding's defect
identity together with the reproducer output it carried back.

**The fix subagent's own dispatch is recorded too, with `-role panel-fix`:**

```bash
flow record dispatch begin -change <name> -role panel-fix -model <m> -effort <e> \
  -key panel-fix-<round>[-<chunk>] -session-token mf-<literal-token>
flow record dispatch end -change <name> -key panel-fix-<round>[-<chunk>] \
  -session-token mf-<literal-token> -commit <partner-task-sha> -outcome completed
```

`-commit` is the task commit the fixup was folded into.

A Minor either fixed or deferred blocks nothing; a Minor left `open` blocks exactly as a Critical
does. When fix rounds do not converge, the run decides each unresolved finding itself, one finding
at a time, and does not ask:

- **A defect a code or document change can resolve takes another fix round** — at Critical or
  Important, and even where the fix reaches past the change's original scope. A Minor that reached
  this loop joins that round only beside a Critical or Important taking one; with none, it is
  deferred under the round's Minor-deferral default above.
- **A finding no change to the tree can resolve is withdrawn** — a verification-only ask, a proof
  that needs an environment the run does not have, a defect something already covers — recorded
  `-status 'withdrawn <reason>'`, the reason one clause naming that mechanism. That reason stands
  where the operator's would.
- **Only a genuine inability reaches the operator:** a finding the run cannot judge either way, a
  fix that needs an irreversible or outward-facing action, or a defect identity still open after
  two automatic rounds on it. Only then is the prompt below raised, shape per Operator prompts
  (`skills/flow-contracts/operator-prompts.md`), and resolved per that contract's **Auto-resolution**
  and **What still stops**: its recommended option is taken unasked, and asked only on a repeat for
  the same finding or where the fix is irreversible or outward-facing:

> **`<location>` — <the finding, in one line>. The fix round did not resolve it.**
> - **Take another round on it** *(default, recommended)*
> - **Withdraw it — I'll give the reason** — the reason is recorded on the finding's marker line
> - **Stop the run and hand it back to me**

**Every automatic decision is recorded in the pass log** with `flow record pass -round <round>
-note 'auto-decided F<n>: another round — <the defect>'` or `-note 'auto-decided F<n>: withdrawn —
<reason>'`, so the handoff shows it. Only this loop records `withdrawn`: automatically with the
run's reason, or on the operator's answer with theirs.

**A fix round closes on a clean re-run, never on its own verification.** The reproducer re-runs
and the fix-diff path check above close the findings that fix addressed; they never close the
panel. After a round that dispatched a `panel-fix` chunk, the delta re-run the rules above trigger
runs before the close guards below do, and the stage may close only on a re-run that raised no new
finding above Minor; an empty-delta re-run whose slots were recorded `not re-run — nothing new
since its last read` counts as clean, and a re-run that re-raises a defect withdrawn
under the handback above is recorded `withdrawn` with its original reason, never
`open`, and does not stand in the way of that clean round. The last round before the stage close
is therefore one of: a pass 1 that raised nothing, a re-run that raised nothing, a re-run whose
only raise was a defect recorded `withdrawn` under the carve-out above, or a round every one of
whose findings was Minor, which re-runs no slot and closes beside them. A
re-run that finds a fix incomplete opens the next fix round under the rules above, and the cycle repeats until a
re-run comes back clean.

### Deferred findings go to KNOWN-BUGS.md at round close

**A round close that leaves findings newly recorded `deferred` records them in
`<project>/KNOWN-BUGS.md`, unasked, before the close guards below run** — **Deferred review
findings** (`skills/flow-contracts/known-bugs.md`) is canonical for the entry and its commit. No
Jira issue is filed and no prompt fires at this close.

**Before closing the stage**, the parent runs both close guards:

```bash
check-panel-findings-closed.sh <worktree> <change>
```

Exit 0 proceeds to the stage close below. Exit 1 means a finding still reads `open` in the store,
or a Minor reads `deferred` in a round that raised a Critical or Important not recorded
`withdrawn` — the Minor-deferral default of **Panel re-runs** above, violated; the line names the refs and the round. It also
fires on the class **Panel re-runs**' ordering above added: a finding recorded `fixed`
whose slot has no clean re-run dispatch of it in any later round. Either way, return to the
handback loop above for them. Exit 2 stops the run.

Beside it, run

```bash
check-panel-fix-single-dispatch.sh <worktree> <change> <session-token>
```

— the token this run stamped on its own dispatches. Exit 0 proceeds to the stage close below.
Exit 1 names every violation of the chunked fix-dispatch contract above — an over-bound chunk
count, non-contiguous chunks, a chunked round with no bare key, a per-chunk count or retry
violation, a key outside the canonical shape — and is a prompt, resolved on its recommended
**Continue** per **Auto-resolution** (`skills/flow-contracts/operator-prompts.md`):

> **The fix round(s) broke the chunked fix-dispatch shape:** <the guard's violation lines>
> - **Continue — the violation stays recorded in this run's output** *(default, recommended)* —
>   the findings may already be verified closed, and the round's work is real
> - **Stop the run**

Exit 2 stops the run — a close the guard cannot answer for is not a clean close.

Beside them, run the project's configured build-green guard, when the project declares one
(**The guard's scope**, `skills/flow-contracts/build-green.md`), over the plan's `tasks.md` — the
same gate the round close above runs, so a pass that raised nothing and closed without a fix
round cannot close the stage over tags it never checked. Exit 0 proceeds to the stage close
below. Exit 1 means the plan still carries a build-green violation: fix the tags as a
planning-path edit, re-run the guard to exit 0, and only then close — a stage never closes on a
plan the guard fails. Exit 2 stops the run.

```bash
flow stage end -command '/flow' -stage flow.review-panel -outcome completed <name>
```

Once this stage ends clean — zero open findings at any severity, no stale result as **Panel
re-runs** defines it — continue into
`skills/flow/verify-and-handoff.md`.
