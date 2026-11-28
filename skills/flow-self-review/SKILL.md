---
name: flow-self-review
description: Run a change's self-review pass, a one-shot dispatch — inside `/flow`'s integrate and `/flow-interactive`'s stop, or standalone from the context bundle flowd saved at cleanup; fix and land every flow-fix finding that is not big with one fixer, suggest the impactful big ones and other angles' findings, record each in the flow store. Standalone, not a pipeline stage.
user-invocable: false
allowed-tools: Bash(git:*), Bash(flow:*), Bash(go:*)
license: MIT
compatibility: Standalone, requires the change's default branch, or its `--base` branch, to be checked out and a context bundle saved in flowd.
---

Run a change's self-review reasoning pass — the only one the pipeline has — over the bundle
`/flow`'s run 1 (**Run the self-review pass**,
`skills/flow-contracts/finish-contract-run1.md`, canonical for that bundle's shape) or **Explicit stop**
(`skills/flow-interactive/SKILL.md`) assembles: inside that run, or standalone, over the one
run 2's cleanup saved in flowd (`flow change save`). This file is canonical for the seven angles, what is fixed, what is suggested and the store record. **Neither the pass nor its fixes run
in this session**: step 2 dispatches the pass, and step 3 fixes, each a one-shot
dispatch on the pair this file names — never the model this session runs on.

**This is not a pipeline stage.** It takes one change name, writes no
per-change state file, and marks no `flow stage` call.

**No flags.** The only argument is the change name.

**Announce at start:** "Using flow-self-review."

## Workflow

### 1. Resolve and refuse

Read the bundle with `flow self-review bundle -change <name>`, run from the main checkout — outside
the change's worktree it prints the bundle flowd saved —
writing its stdout to `<sr-dir>/bundle.md`, never printing it. Its exit 1 — flowd holding none,
or flowd unreachable, its stderr saying which: print `no saved bundle for <name>` or
`flowd unreachable` with that line, and stop.

`git branch --show-current` must be `<landing-branch>`: `<default-branch>` (the project's own default branch), or a `--base` change's `<base>` — a branch other than the change's own (`spectre/<name>`, `<name>` on `/flow-interactive`). A
checkout on any other branch: print the branch and stop.

**Both refusals bind a standalone pass only.** A pass run inside `/flow`'s integrate or
`/flow-interactive`'s stop has its bundle at `<sr-dir>/bundle.md`, runs on the change branch, and starts at step 2.

### 2. Read the bundle and run the seven angles

**This session passes file paths, never file contents.** Every handoff between this file's steps
is a file in `<sr-dir>`, `<state-dir>/self-review/<name>/`, `<state-dir>` being the path `flow
state dir` prints — outside every checkout, never committed. The one exception is the pass's own output: the harness refuses a subagent's Write under `<sr-dir>`, so the pass hands back `findings.md`'s full content and this session writes it. This session reads each dispatch's
hand-back, never the files it names, save where a step below says it reads one; every judgement —
`big`, filing, landing, and when a step runs again — stays this session's.
Before the pass is dispatched, this session empties `<sr-dir>` of every file but `bundle.md`, so a re-run never fills an earlier run's draft. A pass already run is never run again: while the change holds a completed `self-review-<name>-pass` row (`flow record dispatches -change <name>`) and the loop branch `self-review-<name>` exists in `<agents repo>`, this session keeps `<sr-dir>` whole, dispatches no pass and resumes at step 3 from `<sr-dir>/findings.md`.

Dispatch the pass, in-run and standalone alike: key `self-review-<name>-pass`, on the `self-review-pass` row
of **Allowed pairs** (`skills/flow-contracts/model-policy.md`), its prompt naming the bundle's path — `<bundle>`
standalone, `<sr-dir>/bundle.md` in-run — and `<sr-dir>`, and carrying this step's rules and step 3's classification, the
dispatch paragraphs step 3 names and READ-ONLY REVIEW. It hands back the full content of `<sr-dir>/findings.md`: every finding under an `F<n>` id, with its angle, its
evidence, its severity, its theme, `uncategorized` when none fits (**The theme list**, `skills/flow-self-review/themes.md`, which the pass reads) and its classification — `noted` or fixable, in or out of scope, and for a fixable one its
counted blast radius and whether it is `big`.
Beside that content it hands back its per-angle counts, the non-`big` finding ids, one line per `big` finding, and whether it met a product defect. Before step 3, this session writes the returned file to its target path, byte for byte. This session keeps the filing and the
store rows. The pass's dispatch row is opened under `<name>` by the hook;
when the pass hands back, this session closes it, `flow record dispatch end -change <name> -key
self-review-<name>-pass -session-token self-review-<name> -outcome completed`.

**Every suspected defect is traced to a verdict before the pass returns it.** Reproduced, it is
fixable, `big` or not; disproved, it is `noted`, the evidence that disproved it in its line. "Not
traced", "unverified" or a cause left unknown is never a classification: such a finding goes to
step 3's fixer as a non-`big` finding whose first task is the trace, never into the report or
into the run's final report as an open item. An observed test failure is fixed as the fix procedure (`skills/flow/fix.md`) states
for one that does not reproduce within its bounded runs and goes straight to code, and a fix that closes it unreproduced is sent back.

The pass runs the seven-angle table over the bundle, every angle explicit —
present-but-empty for an angle with no findings, never omitted:

| # | Angle | Label |
|---|-------|-------|
| 1 | Problems encountered, and what pipeline change would avoid them | `flow-fix` |
| 2 | What went well, and how to reproduce it | `flow-improvement` |
| 3 | Instruction gaps that cost review-panel or visual-verify fix rounds, and the generic implementer or reviewer prompt update that would have avoided them | `flow-prompt-tuning` |
| 4 | Token, time and wall-clock cost the run spent working or waiting, and what pipeline change would reduce it without quality loss | `flow-cost` |
| 5 | A guard, check, hook, refusal, lint rule, script or other automation that would have avoided a problem or removed a manual step | `flow-automation` |
| 6 | An instruction, guard or skill section that can be made smaller — merged, shortened or relaxed — without losing what it protects | `flow-simplify` |
| 7 | An instruction, guard or skill section that can be deleted outright: it never fires, never loads, or costs more friction than it saves | `flow-cut` |

A silent angle and a skipped angle are indistinguishable to a reader, which is why an empty angle says so.

Angles 6 and 7 read `<sr-dir>/prune-candidates.md`, the stdout of `flow-guard prune-candidates`, which this session writes after emptying `<sr-dir>` and names in the pass prompt beside the bundle: the registered guards that never ran and those that ran and never refused over the last 30 days, and each guard or hook with two or more incidents in that window — its count, minutes lost and latest symptom — and each `.md` under `skills/` no session loaded in that window, by the loads the `flow-skill-load` hook records per file, never per section, annotated with the condition of a conditional load set that names it. Its `never-loaded window:` line opens at the earliest recorded load, and the never-loaded lines say nothing about any time before it. Its exit 2 leaves its stderr line in that file instead, and the two angles then read the bundle alone. A line is a candidate, never a finding: the pass reads the guard's or section's text and call sites and reports a finding only where removing or shrinking it loses nothing the pipeline needs — a guard that never refuses because the mistake it catches stopped happening is a `flow-cut` finding; one that never refuses because it runs before every caller already got it right is not. Every finding of these two angles quotes the prune-candidates line it rests on in its `## Examples`.

Angles 1 and 5 read the bundle's `failed tool calls per dispatch` section — each dispatch's failed
tool calls, their trimmed error and the call that fixed them (absent when none failed). A failure
the agent caused and corrected is a finding when it recurs or its cause is a pipeline gap; it takes
the `agent-command-mistake` theme. No credential, `.env`, `~/.claude` or `~/.config` content is in it.

Angle 3 reads the run's review-panel findings and its visual-verify findings, each judged alone on its own severity — no severity
needs another alongside it. A Critical finding qualifies; an Important finding qualifies; a Minor
finding qualifies only when the pass judges it genuinely useful, meaning it is in a class the
prompts should handle and would recur. The verifier grades no severity: a visual-verify departure blocks the handoff
(`skills/flow/visual-verify.md`, **Steps 3–13**), so it is judged as an Important. A finding qualifies only when its root cause is a generic
gap in instructions:

- **`implementer`** — a rule in the implementer instructions (`skills/flow/implementer-brief.md`,
  `skills/flow/implement.md`, `skills/flow/fixer-brief.md`, `skills/flow/engineering-principles.md`) would have prevented the
  bad code.
- **`reviewer`** — either a **miss**, a defect raised only in a later round or by a later slot that
  a first-round reviewer instruction (`skills/flow/*-reviewer-prompt.md`) would have caught, or a **false positive**, a finding later
  disproved or withdrawn that cost a fix round for nothing. A visual-verify finding's reviewer
  is the verifier: its miss or false positive targets `skills/flow/visual-verify-verifier.md` or
  `skills/flow/visual-verify.md`.

A **late-round miss** — a finding a slot first raises in round 1 or later on a hunk that same slot
already reviewed, unchanged, in an earlier round — is traced to why the earlier rounds missed it,
whatever its severity: the pass compares that slot's earlier dispatches with the round that raised
it — the slot's reviewer prompt file at the change's merge base, the diff each was given, the
findings it returned. A generic instruction gap is a `reviewer` miss finding naming the rounds that missed it; any other cause — a re-run shown only the delta, or reviewer variance the
bundle cannot explain — is an angle-3 finding naming that cause.

An angle-3 finding's note opens with its trace, `prompt: <file> | slot: <slot> round <r> | fix: <sha>` — the prompt file the update belongs in, the review slot and round that raised the defect, and the commit that fixed it in the change — then states the defect class in one sentence with no names from the task, quotes the
current prompt text that shows no rule covers the class, names the prompt file the update belongs
in, and says why the update generalises rather than patching one incident. A finding that fails any
of those is not reported. **Zero findings is a normal result** — the angle then carries the
none-marker like any empty angle.

The visual-verify step is analysed on both of its halves, because its fix rounds are where a run
most often spends repeated rounds: its prompts, by the findings above, and its tooling — the capture,
the mockup comparison, the scripts and browser setup `skills/flow/visual-verify.md` drives. A
tooling gap that cost a visual-verify fix round — a capture that missed the state the defect was in,
a comparison that could not see it, a setup step that failed and was redone — is an angle-3 finding.

An `In-run pipeline fix:` line in the bundle (**Pipeline defects found mid-run**,
`skills/flow-contracts/pipeline.md`) that names a sha is reported under angle 1 as fixed and
is never offered for filing; one that reads `deferred` is an angle-1 finding like any other.

**A finding is fixed or suggested only from the seven angles.** A `flow-fix` finding is
fixed in step 3 unless it is `big`, and suggested under its angle when it is. A `flow-improvement`,
`flow-prompt-tuning`, `flow-cost`, `flow-automation`, `flow-simplify` or `flow-cut` finding is never fixed in the run: it is suggested, in step 4. An observation with nothing to fix is reported and recorded `noted`, never
filed. **An observation that names
an action the run can take is a fixable finding, never `noted`**, unless the action is `big`, reaches production
(`rules/never-touch-production.mdc`) or deletes what cannot be restored. **The pass covers
the pipeline and the project's own dev tooling — build scripts, dev-stack and test-harness
config, guards — and nothing else.** A finding about the project's own product code is out of
scope: it is never filed, and goes to a fix run instead — a pass inside a run stops
here, at the end of this step (**the sweep**, `skills/flow-contracts/known-bugs.md`). A landing HEAD
whose UI no visual verifier saw is such a finding, never a risk line. The draft carries no
section beyond the seven angles.

**One combined pass** — never three separate reads. The pass covers what the bundle holds and
nothing beyond it.

### 3. Fix every `flow-fix` finding that is not `big`

**Every finding is classified before any work**, by the pass, per the blast-radius rule of **Pipeline defects found
mid-run** (`skills/flow-contracts/pipeline.md`): `big` when its fix touches 60 or more files — a
design choice never is, and neither is a
reach outside `<agents repo>`, whose fix goes where that section places it. An
`In-run pipeline fix:` line naming a sha is already fixed and is not classified again. **A `big` finding is not fixed here**: step 4 files it, and its fix is a separate flow run.

**A finding whose fix would add or extend a guard, check, hook, refusal, lint rule, script or other
automation is never built by the pass.** It is a `flow-automation` finding, suggested in step 4 like any other, its
detail naming the proposal.

**Every other `flow-fix` finding that is not `big` is fixed and landed without asking**, on one branch for the
whole pass. A pass with none creates no worktree and dispatches nothing. No scope or cost question
with a recommended option is put to the operator anywhere in this pass, as a prompt or as prose in
a message (**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`).
**A fix repairs what its defect already wrote wrong** — a store row, a record or a file the defect
left corrupted is repaired in the same commit, a small command added where none exists, never left
as a forward-only fix with the repair an operator's option.
A finding is left unfixed only when it is `big`, or disproved — the pass shows, with evidence, that
the defect it names does not exist.

**This session fixes nothing itself** save **The base sync** below: the fix is a fresh one-shot dispatch on the pair its row names, never a fork of this session (**Dispatch sites — the parent's closed list**, `skills/flow/implement.md`), its prompt carrying the TOOLS, FOREGROUND BUILDS and
NO DELEGATION paragraphs of `skills/flow/implementer-brief.md` verbatim. Every dispatch is one-shot: no `SendMessage` to it once
it returns.

**The pairs are the `self-review-pass` and `loop-*` rows** of **Allowed pairs**
(`skills/flow-contracts/model-policy.md`): `self-review-pass` for the pass and `loop-fix` for the fixer, the latter
on **the fixer pair** per **Model and effort** (`skills/flow-contracts/model-policy.md`).

**Every dispatch is recorded mechanically.**
`flow-guard hook enforce-loop-dispatch` hands each Agent call whose prompt carries a
`Key: self-review-<name>-…` line to
`<agents repo>/scripts/flow-guard check-loop-dispatch`, which records it in flowd. The parent closes each dispatch
once it returns with `<agents repo>/scripts/flow-guard pipeline-fix end <key> <outcome>`: `completed`, or
`aborted` for one that did not finish or never ran — a dispatch denied after it was recorded is
closed `aborted` and may be dispatched again under the same key. A refusal or failure of any of these tools on a step this file prescribes is a
pipeline defect, fixed in-run per **Pipeline defects found mid-run**
(`skills/flow-contracts/pipeline.md`) and the run resumed once the fix lands — never a stop or a
question to the operator.

`<agents-base>` is the one **Pipeline defects found mid-run** binds. `<fix-branch>` is always
`self-review-<name>`, based at `origin/<agents-base>`, landed by `flow-guard pipeline-fix land`; fixes never go onto the change's own branch.

1. **Fix** — this session creates the worktree with
   `<agents repo>/scripts/flow-guard pipeline-fix start <name>` — never the main checkout — then
   dispatches one fixer, key `self-review-<name>-fix`, its prompt naming `<sr-dir>/findings.md` and
   the non-`big` finding ids, whose evidence and counted blast radius that file carries. It fixes
   them all in that worktree, one commit per finding with a conventional subject scoped to the
   module it moved and no `Task-Id:` trailer, adding one test that fails without the fix wherever the fix changes behaviour,
   and running the `<agents repo>/.flow/project.md` `## lint` lines its files need. It never writes
   `<agents repo>/spectre/changes/`, and never merges or pushes. It reports each finding's commit, or a finding it found `big` once under way, left uncommitted.
2. **Review once, primary only.** The fixer's diff against `origin/<agents-base>` is reviewed by one
   fresh dispatch on the `loop-review-1` row, key `self-review-<name>-review-1`, per
   `skills/flow/primary-reviewer-prompt.md`, writing its
   report to `<sr-dir>/review-1.md`; the parent closes it with `flow-guard pipeline-fix end <key> <report>`,
   the report carrying its `Critical:` and `Important:` lines. Neither line ⇒ item 3. Otherwise one fresh fixer dispatch,
   key `self-review-<name>-fix-r<r>`, corrects them in the same worktree, then a fresh re-review on the
   `loop-review-r` row, key `self-review-<name>-review-r<r>`, `<r>` counting from 1. The loop ends at a review with neither line, and
   three reviews in a row that still carry one take **Three failed rounds** (`skills/flow-contracts/pipeline.md`). Minors
   are never fixed here and never block.
3. **Land once.** Land by `<agents repo>`'s `## default landing route`
   (`flow-guard project-get <agents repo> 'default landing route'`), without asking — merge and push as
   step 4 of **Pipeline defects found mid-run** states it, with `self-review-<name>` as the
   branch. `<agents repo>/scripts/flow-guard pipeline-fix land <fix worktree> > <sr-dir>/verify-<v>.log 2>&1` runs the `## lint` lines and the tests on the rebased HEAD before it pushes, `<v>` counting land attempts from 1; this session reads only its exit code and its `red:` lines. A red line is fixed by a fresh fixer dispatch, key `self-review-<name>-fix-v<v>`, its prompt naming the failing command and `<sr-dir>/verify-<v>.log`, then `land` again. **One retry rule**: a red branch never lands and a red line is never waived, and three red lands in a row take **Three failed rounds** (`skills/flow-contracts/pipeline.md`): the next fixer dispatch carries every `verify-<v>.log` of those rounds. The pass is not done until every fixed finding's commit is on `origin/<agents-base>`. Each fixed
   finding's sha is read off `<agents-base>` after the landing, never from
   the branch before its rebase; on a route that opens a pull request, once that pull request
   merges.

**The base sync.** A finding that a `--base` change's `<base>` (`skills/flow/SKILL.md`) lacks a
commit the project's `<default-branch>` carries is no fixer task: in each repository of the
change whose `<base>` lacks it, this session adds a detached worktree of `origin/<base>`, named
`base-sync-<name>`, in that repository's worktrees directory after a `git fetch origin`, runs
`git merge --no-edit origin/<default-branch>` in it, then the `## lint` and `## test` commands
`<project>/.flow/project.md` declares for that repository — a repository with no declared `## lint`
and `## test` command has no green run; on a clean merge and a green run it
pushes the merge to `<base>` on `origin` without asking, removes the worktree, and records the
finding `fixed` with the merge commit's sha. A merge that conflicts, a red or absent run, or a
rejected push pushes nothing,
removes the worktree, and is filed in step 4 as a `big` finding is.

### 4. Suggest the findings

Nothing is asked and no Jira issue is created: the operator reads the suggestions on the Self-review
dashboard's **To review** tab, approves or declines each, and creates and plans any task themselves.
Only an impactful finding is suggested: a Critical or Important one (**Calibration**, `skills/flow/primary-reviewer-prompt.md`). A Minor
finding is never recorded as a suggestion — `flow self-review finding` refuses `suggested` with `-severity minor` — and is
left out of the report; this holds for suggestions only, and fixes are recorded as they always were. Every impactful
`big` finding, every impactful `flow-fix` finding step 3 or **The base sync** left unfixed, and every impactful
`flow-improvement`, `flow-prompt-tuning`, `flow-cost`, `flow-automation`, `flow-simplify` and `flow-cut` finding is recorded in step 5 as `suggested` (after its similarity check), one row each,
carrying its full `-detail`, written for the operator who decides on it and has not seen the run: one markdown body
with these headings, in this order — `## Proposal` (the one change asked for, one or two sentences), `## What happened`
(the situation in plain words: what the run was doing, what went wrong or well, and why that matters), `## Examples`
(the concrete incidents, one bullet each: `<where it happened> — <what happened>`), `## Current text` (the exact
sentence of the current prompt or contract that shows the gap, quoted; omit for a finding with no such text),
`## Proposed change` (each file and a fenced `diff` of the edit; omit when the change is not an edit),
`## Effort and risk` (three lines, `Effort: S|M|L`, `Touches: <files>`, `Risk: <one line>`, and a fourth, `Guard: new — <the guard's name> | extends <the existing flow-guard it extends>`, on a finding about a guard — one that adds a check, a refusal or a verdict — so the operator sees at once whether it asks for a new guard or more from one that exists, by searching `guard.Registry` first), `## Benefit` (what stops
happening or gets cheaper), `## Done when` (a checklist, one `- [ ]` item each) and `## Trace` (the prompt file, the
review slot and round that raised the defect, and the change's fix commit sha for a `flow-prompt-tuning` finding;
the bundle section or log line the finding came from for any other), so a later change can be tied back to
what prompted it.

**The detail is written to be read cold.** Say what happened as a short story in everyday words — the
situation, the thing that went wrong, what it cost — before any term of this pipeline; each pipeline term
(a slot, a round, a gate, a shard, a theme tag) is glossed in a few words the first time it appears, and a task
number, finding id or sha is never the only way a thing is named. Every example says what a person would have
seen, not only which step failed. State cause and effect in one sentence each, so the reader never has to join two
sections to see why the proposal follows from the evidence; never a finding-id-only reference to another
finding. A reader who knows the project but not this run must be able to approve or decline from the body alone.

A `fixed` row carries a `-detail` too: what was wrong, why, what the fix changed and the
evidence it holds.

### 5. Record every finding in the flow store

One call per finding, every disposition alike, before step 6 — a fixed finding's `-ref` read
off `<agents-base>` after step 3's landing:

```bash
flow self-review finding -change <name> -angle <label> -disposition fixed|suggested|declined|noted [-ref <sha>] [-blast-radius <N>] [-severity critical|important|minor] -note '<the finding, one line>' [-detail '<markdown>']
```

`-note` opens with the finding's theme prefix, `uncategorized` for a finding no tag fits (**The theme list**, `skills/flow-self-review/themes.md`).
Before recording each `suggested` finding, and before an `uncategorized` one, the pass runs `flow self-review similar '<the finding in a few words>'` and reads the earlier findings it prints; it ranks the project's recorded findings by their wording, and a close match shows which theme the earlier one took. Each hit carries its `disposition` and, on a suggestion, its `reviewStatus`. A close match to an earlier suggestion with `reviewStatus` **declined** is recorded `noted`, never `suggested`, its note naming the match. A close match to one still `to-review` is still recorded `suggested`, its `## Examples` naming the earlier finding.

When a landed commit implements a `suggested` finding, `flow self-review finding-implement <id> -commit <sha>` marks it: it leaves `to-review` as `approved` with the reason `implemented in <sha>`, repeating it changes nothing, and an unknown id exits 1. Never record a second finding for the same thing.
`-ref` is the landed sha for `fixed`, and omitted for `suggested`, `declined` and `noted`; `-detail` is required for `suggested`;
`-blast-radius` is step 3's count; `-severity` is the grade the pass gave the finding, required for `fixed` and `suggested`, which `flow self-review finding` refuses without it, and omitted for `declined` or `noted` when the pass gave none. An `In-run pipeline fix:`
line naming a sha gets no second row when `flow self-review findings -change <name>` already
lists that sha (**Pipeline defects found mid-run**, `skills/flow-contracts/pipeline.md`). A row this step records for such a line, which the pass never grades, takes `-severity` by the rule of **Pipeline defects found mid-run** (`skills/flow-contracts/pipeline.md`). A store failure is one
warning line and the pass continues to the listing below.
A row is stated as recorded only once `flow self-review findings -change <name>` lists it: read that listing before step 7, record every row it lacks again, and fix a store that still refuses per **Pipeline defects found mid-run** (`skills/flow-contracts/pipeline.md`).

### 6. Delete the scratch directory

No report is written or committed: the findings live in flowd. Delete `<sr-dir>` once step 5's
listing holds every row and item 2 of step 3 recorded its `fixed` rows.

### 7. Report

One line per finding — its angle, what it was, and its landed sha or `suggested` or `noted` — then
the shas landed, the suggestions recorded (the count), each theme miss the pass handed back, and each product defect the pass met (**the sweep**,
`skills/flow-contracts/known-bugs.md`).

## Guardrails

- **Never** push directly to a protected branch under any name but the project's own default
  branch, or the `<landing-branch>` step 1 accepted, and never force-push.
