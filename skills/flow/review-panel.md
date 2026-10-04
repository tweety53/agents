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
sync-panel-base.sh [--rebase] <worktree> <working-notes-merge-base>
```

`<working-notes-merge-base>` is the merge base `skills/flow/implement.md`'s isolate-workspace step
recorded in this run's working notes — never the state file's `worktrees` map, which a creating run
has not written yet. The script fetches, prints `BASE: <base>` and every `check-base-moved.sh`
verdict, and rebases a worktree whose verdict is `MOVED` with no overlap automatically, with no
prompt — the run reports that it happened rather than asking whether it should. Its header
(`<agents repo>/scripts/sync-panel-base.sh`) is the full contract. Exit 0 is settled; 3 is an overlap: the rebase
is not confirmed conflict-free, so the run never takes that risk on itself; 1 is `CONFLICT`; 2 — a
`REFUSE`, or anything it cannot answer — and an empty resolved set stop and ask. An exit 3 from any
worktree is one prompt for the whole change, shape per Operator prompts
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

**Rebase** re-runs the script with `--rebase` only for a worktree that exited 3 — never one whose
verdict was `CLEAR`, even though the prompt above is asked once for the whole change. The automatic
no-overlap rebase above is the same mechanism with nobody asked, and the two outcomes that follow
govern both:

- **Clean** (`REBASED:`): that worktree's working-notes merge base becomes the sha the last
  `REBASED:` line names; every later `<merge-base>` this file and the `worktrees` map
  `skills/flow/verify-and-handoff.md` writes read the working notes, so nothing else needs
  plumbing. A fresh overlap on the script's own re-check is exit 3 again and re-offers the prompt
  above rather than looping silently. The rebase clears every slot's held last-reviewed
  sha, so a re-run after it reads the whole `final-review.diff` under **Panel re-runs**' existing
  no-held-sha rule (`skills/flow/review-panel-fix-round.md`). No re-verification runs here — the panel reads the rebased tree, and
  `flow.verify` follows. End the clean-rebase report with the fixed literal:

  > If this change's verification compares against a recorded baseline, recapture it now — a
  > proof taken against the pre-rebase base is void.

- **Conflict** (exit 1): never auto-abort, and never resolve the conflict — by editing the
  conflicting files or otherwise. An automatic rebase lands here too — no reported overlap does
  not rule out two commits touching one file incompatibly outside this change's own touched paths
  — and is handled identically. The worktree is left mid-rebase exactly as `git rebase` left it;
  report the conflicting file(s) the `CONFLICT:` line names, and hand off `git -C <worktree> rebase
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
`primary-reviewer-prompt.md` and `reviewer-calibration.md` (Primary), `principles-reviewer-prompt.md` and
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
| `[CALIBRATION_PATH]` | primary | the **absolute** path of `reviewer-calibration.md` beside this file, resolved as `[PRINCIPLES_PATH]` is |
| `[CONTEXT_BUNDLE_PATHS]` | primary, failure-modes | the paths the CONTEXT BUNDLE paragraph names, `<abs-worktree>/.superpowers/sdd/dispatch-context.md` once per worktree in the resolved set; on the CONTEXT BUNDLE FAILURE continue path, the literal `none — the bundle was not built` |
| `[PRINCIPLES_PATH]` | principles | as the Principles section below resolves it |
| `[STANDARDS_PATHS]` | principles | as the Principles section below resolves it; empty when none resolve |
| `[GLOBAL_CONSTRAINTS]` | principles | the literal `the design.md section of your context bundle` — never a selection or paraphrase; on the CONTEXT BUNDLE FAILURE continue path, the absolute path of the change's `design.md`, or `none` when it has none |

`ValidReviewers` in `<agents repo>/stats/internal/store/settings.go` is the id vocabulary this table exhausts.
A `default` panel, which records no pair, runs every slot on the literal `opus`
(**Model and effort**, `skills/flow/brainstorm-planner.md`). Every slot in this table
is dispatched on `flow-low` (`agents/flow-low.md`, on
a `default` panel) and carries the same model rule.

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
carries `mutation` — it carries the mutation slot's prompt and **The throwaway worktree**.
**Load `skills/flow/review-panel-experimental-slot.md`** before dispatching any round whose roster
carries an `exp-` slot — it carries **Experimental slot**. A round whose roster carries none of
them never reads it.

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

Write `<abs-worktree>/.superpowers/sdd/final-review.diff` (the canonical worktree's) and its
`[TOUCHED_FILES]` list once per round from **every** worktree in the change's resolved set
(**Resolving a change's worktrees**, `skills/flow-contracts/worktree-resolution.md`), in resolved
order, with one call:

```bash
write-panel-diff.sh final <abs-worktree> <worktree> <merge-base> [<worktree> <merge-base>…]
```

Its header (`<agents repo>/scripts/write-panel-diff.sh`) is canonical for both files' shape. Exit 0
prints the two paths; exit 2 — it cannot answer, and nothing was written — stops the run with its
stderr. Then dispatch the round's
`panel.dispatches` in the canonical worktree, each reading the whole combined file; a role is
never dispatched once per worktree: one pass reads every worktree's section, so a seam between two repositories is in one pass's view. **Bundled dispatch** below states how the roster is
grouped into those dispatches.

### The late-fix reduction

**Load `skills/flow/review-panel-late-fix.md` only when** this run is a fix run — a `/flow`
invocation whose argument is fix instructions, or a plain message at `IN_PROGRESS`
(`skills/flow/SKILL.md`) — before this stage opens its first round; it carries
the reduction's conditions, its single dispatch and what voids it.

### Bundled dispatch

Before dispatching any panel round,
re-check the decision's `panel.dispatches`/`panel.grouping`.

**One bundled review dispatch per round is the starting shape: every role the round dispatches
rides one dispatch**, one rendered **PASS `<id>`** section per role in roster order, on decided and
`default` panels alike and in both execution modes — findings still recorded per role. A roster
past the three-role cap spills its overflow — the mutating role (`mutation`) last — into a second
dispatch, and a roster the two dispatches cannot hold shrinks to what they hold. **The full-roster
shape — one dispatch per role — is the explicit fallback, taken when a bundled dispatch's findings
need separation** — a finding the bundle cannot attribute to its role, or a bundled read that must
separate to hold review quality — and recorded with `flow record pass -round <round>` naming why.

**Grouping.** On a decided panel, the decision's `panel.grouping` is `static` — the class's row in
the tree of **Decide** (`skills/flow/brainstorm-planner.md`), unchanged, no override — or `free` — the
planner's own grouping within the ≤2 × ≤3 cap, recorded as `panel.grouping_reason`. On a `default` panel,
the settings-store roster is grouped deterministically by the same static logic, no roll and no
planner: the floor roles (`primary`, `principles`) open the first dispatch and it fills to its
three-role cap in roster order, the mutating role (`mutation`) last, and only the overflow opens
the second; a list the two cannot hold is truncated in store order, and the truncation is
recorded with `flow record pass -round <round>`.

**One `dispatches` row per bundle** — the same `flow record dispatch begin`/`end` pair below, with
`-slot` the bundle's roles `+`-joined in roster order (`primary+principles+failure-modes`) and
`-model`/`-effort` the bundle's own, from the decision's `panel.dispatches` entry. Every finding still records its own single role in `-slot`, with the
bundle's `-dispatch-seq`.

An incident a round's slot caused takes the same course (the incident rule of **4. Execute
(SDD + TDD)**, `skills/flow/implement.md`): recorded with `flow record incident`, and the next dispatch to that role — a
re-run, or the panel-fix subagent — carries it verbatim.

**Every dispatch's brief is rendered, never hand-assembled** — one call per dispatch, a one-role
dispatch included, once the round's diff is written:

```bash
render-slot-prompt.sh <skill-dir> <round> <abs-worktree> <plan-dir> <slot>[+<slot>…] \
  -diff final|late-fix|delta|fix-round [-no-bundle] [-standard <path>…] [-fix-report <path>…] -- <worktree> …
```

`<skill-dir>` is the directory this file was read from and `<plan-dir>` the plan directory the
context bundle reads; `-standard` carries **Principles**' resolved `[STANDARDS_PATHS]`,
`-fix-report` a re-run's `<fix report>`, and `-no-bundle` the CONTEXT BUNDLE FAILURE continue
path. Its header (`<agents repo>/scripts/render-slot-prompt.sh`) is canonical for the file it
writes: the shared paragraphs below, then one **PASS `<id>`** section per role in roster order,
each with its own REPORT FILE line. Exit 0 prints the rendered path; exit 1 names a placeholder it
could not fill, and exit 2 means it cannot answer — either stops the round before its launches.
`mutation`'s pass is never rendered: its brief and its throwaway copy stay typed by the parent,
beside the rendered path where a dispatch bundles it — the render still carries every other
role's pass and the INDEPENDENT PASSES paragraph. The Agent call's prompt carries only the baseline
pointer, MODEL HANDSHAKE, CONTEXT BUNDLE, the relocation-comparison pointer where one exists, the
reproducer rule, and `read <rendered path> in full first — it is your brief`. The return message carries one findings summary per role under a heading naming the
role; the parent records each finding under that role.

Every bundle prompt also carries this paragraph verbatim:

> **INDEPENDENT PASSES:** each pass starts from `final-review.diff` and the code, never from an
> earlier pass's report or conclusions. Do not cite, defer to, or skip a defect because an earlier
> pass raised it — if it sits in this pass's angle, raise it again under this pass. Write each
> pass's report file before beginning the next pass.

**No de-duplication across roles**: the same defect raised by two passes is two `F<n>` rows.

**Re-runs are re-grouped by the same grouping**, carrying only the roles re-running this round — a
group whose other members are clean dispatches with its re-running members only — on decided and
`default` panels alike; the full-roster fallback above separates them one per role when the
bundled findings need separation (**Panel re-runs**).

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
report is read or finding recorded. Exit 1 or 2 stops the round the same way the per-task
reviewer's stop works (the plan-tree rule of **4. Execute
(SDD + TDD)**, `skills/flow/implement.md`). The round brackets
itself with content markers beside that guard (the content-marker rule of **4. Execute
(SDD + TDD)**, `skills/flow/implement.md`): the marker list names the plan
artifacts and working notes the round reads, `check-tree-markers.sh snapshot <worktree>
<markers-file> <snapshot-file>` runs with the plan-tree guard's own snapshot, and
`check-tree-markers.sh verify <worktree> <markers-file> <snapshot-file>` runs before any report
is read or finding recorded.

`-slot` names the dispatch's roles from **The roster** table above. `-role` is
`reviewer` for every one; `-task` is omitted. `-diff-base <sha>` is passed on a dispatch whose
roles are all reading against a delta and on no other; it takes one
sha, so it carries the **canonical worktree's** held last-reviewed sha, and the
panel record names every worktree's sha beside the delta path. `-model` is `opus` (or this run's override) on
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
exits for the record. At pass-1 authoring, where no fix exists yet, the demonstrated leg
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

**Resolve `[TOUCHED_FILES]` before dispatching any slot**: it is the `.touched` file the
round's `write-panel-diff.sh` call wrote beside the diff the slot reads, over that diff's own
range, so a diff-reading re-run's named context narrows with the read. The paths under
`skills/flow-contracts/` in that list are the change's **named contracts**.

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
of `engineering-principles.md` **beside this file** — `skills/flow/`, always. Under the global install that is
`~/.claude/skills/flow/engineering-principles.md`; under a
project-local install it is `<project>/.claude/skills/…` or `<project>/.zcode/skills/…`.
Resolve it from where this file was actually read — never hardcode a repo-relative
`skills/…` path: the subagent's working directory is the project worktree, which has no
`skills/` tree, so a relative path fails to open and the reviewer loses its principle
list. Confirm the file
exists before spawning; if it does not, stop and report rather than dispatching a blind reviewer.

**Resolve `[STANDARDS_PATHS]` before dispatching the principles slot**, from the entries
`project-get.sh <worktree> standards` prints (exit 1: none declared), resolved per the entry-form
and containment rules of **Project configuration — standards** (`skills/flow-contracts/project-configuration-standards.md`),
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

A slot that supplies nothing for a finding has not supplied a legal exemption: record that omission
as its own open finding.

**Render the record when the panel closes** — every slot's result clean, no finding open — into
the canonical worktree, never into whichever worktree the closing pass runs in, so a
multi-worktree run leaves one copy:

```bash
flow record render -change <name> -kind panel -repo <canonical-worktree>
```

**A panel that raised nothing still renders**, declaring `findings-total: 0`. There is no
skip-the-render shortcut.

**A `withdrawn` marker's reason is checked for being there at all.**

## Panel re-runs

**Load `skills/flow/review-panel-fix-round.md` only when** a round has recorded a Critical or
Important finding, before that round's fix opens, or a close guard below sends the run back to the
handback loop, or a section of it is cited at the point of use (the full-suite fix, the gated
reviewer's fix, the stage-diff check) — it carries every fix round's procedure: the
round-boundary base re-check, the fix commit routes, which slots re-run and on what, the
reproducer guards and re-runs, the fix's mutation proof and round close, the fix chunks and the
non-convergence loop, while the fix subagent's dispatch paragraphs and dispatch record stay in
this file.

**Every Critical and Important goes to the fix; every Minor is fixed too, and a Minor never
causes a fix round of its own.** Every Critical and Important the round raised goes to the fix
subagent below, closed by the verification that follows it — the reproducer re-run exits 0 *and*
the fix diff touches a path the finding named, or one of the path condition's accepted
alternatives below closes the finding. **A round that raised a Critical or Important sends every
Minor it raised to that same fix**, closed by the same verification. **A round that raised no
Critical and no Important has its Minors fixed by the parent, inline, at the round's close** — no
fix subagent, no reproducer run, no mutation proof, no slot re-run and no dispatch record. The
parent edits each Minor's named location, stages the changed paths and makes one commit per
worktree at the branch tip —
`git -C <worktree> commit -m "fix(<module>): review Minors" -- <the changed paths>`, the scope
naming the module the edits moved — pushed per **Branch backup**
(`skills/flow-contracts/git-boundaries.md`), then records each finding
`flow record status -change <name> -ref F<n> -status fixed`. `flow.verify`'s lint and tests,
which run after this stage, are that fix's verification. **A Minor is never deferred**: one no
change to the tree can resolve, or that names no defect, is withdrawn with
`-status 'withdrawn <reason>'`, the reason one clause naming the mechanism — out of scope,
pre-existing and cosmetic are not reasons. Then proceed to `check-panel-findings-closed.sh` and
the stage close. A fixed finding that fails verification takes the handback below, and that loop
re-runs no slot either.

**A rule that changes mid-run governs from the round it lands in, and each round's pass log names
the rule it decided under.** When the wording of a rule this file states — the Minor
rule above included — changes while a panel is in flight, by an operator instruction or by an
edit to this file, the new wording governs from the first round decided after it lands: a round
the old rule already closed stands as closed, never re-decided retroactively. Each round decided
under wording that changed during the run records that wording beside its decisions, with the
round's own pass-log rows — `flow record pass -change <name> -round <round> -note 'decided under:
<the rule as this round applied it>'` — so the rendered pass log shows which rule governed what.

Handoff still requires **zero open findings at any severity** from every agent that has run, and
no stale result — where **a slot's clean result is stale when the rule above required that slot to
re-run and it has not, or when any commit or working-tree change to source landed after that slot's
last read, from any stage — `flow.verify` included**. An unrecorded edit after the panel closes is
stale by definition, not only one a fix round produced. A fix against which a slot raised no
finding leaves that slot's result current: the round's own mutation-proof (below) covers what the
fix changed. **The parent's inline Minor commit leaves every slot's result current** — a source change
after a slot's last read that does not make that result stale. A delta the late-fix
reduction read has its own carve-out (`skills/flow/review-panel-late-fix.md`).

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
> `<changeRoot>/tasks.md` in this same pass — never left for a reviewer to
> catch next round. Edit them; do not stage or commit them — the plan record is
> a planning path and is committed later by the pipeline, never in a fixup.

**Every fix subagent's dispatch prompt also carries the ROUND SCOPE paragraph**:

> **ROUND SCOPE:** your round's output is the diff and the report, nothing else. Do not rewrite
> `proposal.md` or `design.md`: unaffected sections are never restated, and a decision your fix
> genuinely overturns is superseded by one appended entry under `## Decisions`, never a per-round
> rewrite of the file. Your report names what you fixed, the behaviours you changed,
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

Dispatched on the fixer pair's model (below); the dispatcher compares that line against it and applies
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

**The fix subagent's own dispatch is recorded too, with `-role panel-fix`:**

```bash
flow record dispatch begin -change <name> -role panel-fix -model <m> -effort <e> \
  -key panel-fix-<round>[-<chunk>] -session-token mf-<literal-token>
flow record dispatch end -change <name> -key panel-fix-<round>[-<chunk>] \
  -session-token mf-<literal-token> -commit <partner-task-sha> -outcome completed
```

`-commit` is the task commit the fixup was folded into.

**Before closing the stage**, the parent runs both close guards:

```bash
check-panel-findings-closed.sh <worktree> <change>
```

Exit 0 proceeds to the stage close below. Exit 1 means a finding still reads `open` or `deferred`
in the store — nothing is deferred (**Panel re-runs** above); the line names the refs. It also
fires on the class **Panel re-runs**' ordering above added: a Critical or Important recorded
`fixed` whose slot has no clean re-run dispatch of it in any later round — a fixed Minor needs
none. Either way, return to the
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
