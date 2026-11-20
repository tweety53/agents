---
name: flow-self-review
description: Run a change's self-review pass, a one-shot dispatch — inside `/flow`'s integrate and `/flow-fast`'s verify, or standalone from the context bundle flowd saved at cleanup; plan in full, fix and land every flow-fix finding that is not big, file the big ones and each other angle's findings, record each in the flow store. Standalone, not a pipeline stage. Use for /flow-self-review.
allowed-tools: Bash(git:*), Bash(flow:*), Bash(go:*)
license: MIT
compatibility: Standalone, requires the change's default branch, or its `--base` branch, to be checked out and a context bundle saved in flowd.
---

Run a change's self-review reasoning pass — the only one the pipeline has — over the bundle
`/flow`'s run 1 (**Run the self-review pass**,
`skills/flow-contracts/finish-contract-run1.md`, canonical for that bundle's shape) or **5. Verify**
(`skills/flow-fast/SKILL.md`) assembles: inside that run, or, as this standalone command, from the
bundle run 2's cleanup saved in flowd (`flow change save`). This file is canonical for the four angles, what is fixed, what is filed and the store record. **Neither the pass nor its fixes run
in this session**: step 2 dispatches the pass, and step 3 plans, fixes and reviews, each a one-shot
dispatch on the pair this file names — never the model this session runs on.

**This is a standalone command, not a pipeline stage.** It takes one change name, writes no
per-change state file, and marks no `flow stage` call.

**No flags.** The only argument is the change name.

**Announce at start:** "Using flow-self-review."

## Workflow

### 1. Resolve and refuse

Read the bundle with `flow self-review bundle -change <name>`, run from the main checkout — outside
the change's worktree it prints the bundle flowd saved (decision self-review-bundle-to-flowd) —
writing its stdout to `<sr-dir>/bundle.md`, never printing it. Its exit 1 — flowd holding none,
or flowd unreachable, its stderr saying which: print `no saved bundle for <name>` or
`flowd unreachable` with that line, and stop.

`git branch --show-current` must be `<landing-branch>`: `<default-branch>` (the project's own default branch), or a `--base` change's `<base>` — a branch other than the change's own (`spectre/<name>`, `<name>` on `/flow-fast`). A
checkout on any other branch: print the branch and stop.

**Both refusals bind the standalone command only.** A pass run inside `/flow`'s integrate or
`/flow-fast`'s verify has its bundle at `<sr-dir>/bundle.md`, runs on the change branch, and starts at step 2.

### 2. Read the bundle and run the four angles

**This session passes file paths, never file contents.** Every handoff between this file's steps
is a file in `<sr-dir>`, `<state-dir>/self-review/<name>/`, `<state-dir>` being the path `flow
state dir` prints — outside every checkout, never committed. The one exception is the pass's own output: the harness refuses a subagent's Write under `<sr-dir>`, so the pass hands back its files' full contents and this session writes them. This session reads each dispatch's
hand-back, never the files it names, save where a step below says it reads one; every judgement —
`big`, filing, landing, and when a step runs again — stays this session's.
Before the pass is dispatched, this session empties `<sr-dir>` of every file but `bundle.md`, so a re-run never fills an earlier run's draft or copies its cases. A pass already run is never run again: while the change holds a completed `self-review-<name>-pass` row (`flow record dispatches -change <name>`) and the loop branch `self-review-<name>` exists in `<agents repo>`, this session keeps `<sr-dir>` whole, dispatches no pass and resumes at step 3 from `<sr-dir>/findings.md`.

**Cost and speed observations are counted across changes by theme.** Before the pass is dispatched, this session runs `flow self-review themes` and carries its lines in the pass's prompt; a failed call carries none.
Every `flow-cost` observation the pass returns names the one listed theme it is about, `other` when none fits.
An observation whose theme the listing marks `due` is a finding, never `noted`, however its own change reads: its line states the listing's count.

Dispatch the pass, in-run and standalone alike: key `self-review-<name>-pass`, on the `pass` row
of **Allowed pairs** (`skills/flow-contracts/model-policy.md`), its prompt naming the bundle's path — `<bundle>`
standalone, `<sr-dir>/bundle.md` in-run — and `<sr-dir>`, and carrying this step's rules and step 3's classification, the
dispatch paragraphs step 3 names and READ-ONLY REVIEW. It hands back the full content of `<sr-dir>/findings.md`: every finding under an `F<n>` id, with its angle, its
evidence, its severity and its classification — `noted` or fixable, in or out of scope, and for a fixable one its
counted blast radius and whether it is `big` — and a `## Offered` section laying out every `big` finding and every `flow-cost`, `flow-improvement` and `flow-prompt-tuning` finding in step 4's explanation layout.
It hands back the full content of `<sr-dir>/report-draft.md`, step 6's input, each disposition step 3 decides written as a placeholder: `fixed: <F-id>` for a non-`big` finding, `offered: <F-id>` for a `big` one or one of those three angles'.
It hands back every case file under `<sr-dir>/cases/<case-id>/` the same way, each under its target path, and writes no file itself.
Beside those contents it hands back its per-angle counts, the non-`big` finding ids, one line per `big` finding, its case ids, and whether it met a product defect. Before step 2a, this session writes each returned file to its target path, byte for byte. This session keeps the filing and the
store rows. The pass's dispatch row is opened under `<name>` by the hook;
when the pass hands back, this session closes it, `flow record dispatch end -change <name> -key
self-review-<name>-pass -session-token self-review-<name> -outcome completed`.

**Every suspected defect is traced to a verdict before the pass returns it.** Reproduced, it is
fixable, `big` or not; disproved, it is `noted`, the evidence that disproved it in its line. "Not
traced", "unverified" or a cause left unknown is never a classification: such a finding goes to
step 3's planner as a non-`big` finding whose plan opens on the trace, never into the report or
into the run's final report as an open item. An observed test failure is fixed as the fix procedure (`skills/flow/fix.md`) states
for one that does not reproduce within its bounded runs and goes straight to code, and a plan that allows closing it unreproduced is sent back.

The pass runs the four-angle table over the bundle, every angle explicit —
present-but-empty for an angle with no findings, never omitted:

| # | Angle | Label |
|---|-------|-------|
| 1 | Problems encountered, and what pipeline change would avoid them | `flow-fix` |
| 2 | Token/time cost and wall-clock time the run spent waiting or working, and what pipeline change would reduce it without quality loss | `flow-cost` |
| 3 | What went well, and how to reproduce it | `flow-improvement` |
| 4 | Instruction gaps that cost review-panel or visual-verify fix rounds, and the generic implementer or reviewer prompt update that would have avoided them | `flow-prompt-tuning` |

Angle 2 covers both what a step spends and elapsed time: serial steps that could run in parallel, slow guards, builds or test runs, waits, and redundant re-runs. A silent angle and a
skipped angle are indistinguishable to a reader, which is why an empty angle says so.

Angle 4 reads the run's review-panel findings and its visual-verify findings, each judged alone on its own severity — no severity
needs another alongside it. A Critical finding qualifies; an Important finding qualifies; a Minor
finding qualifies only when the pass judges it genuinely useful, meaning it is in a class the
prompts should handle and would recur. The verifier grades no severity: a visual-verify departure blocks the handoff
(`skills/flow/visual-verify.md`, **Steps 3–13**), so it is judged as an Important, and its case records
`source.slot` `visual-verify` and `source.severity` `Important`. A finding is a case only when its root cause is a generic
gap in instructions:

- **`implementer`** — a rule in the implementer instructions (`skills/flow/implementer-brief.md`,
  `skills/flow/implement.md`, `skills/flow/fixer-brief.md`, `skills/flow/engineering-principles.md`) would have prevented the
  bad code.
- **`reviewer`** — either a **miss**, a defect raised only in a later round or by a later slot that
  a first-round reviewer instruction (`skills/flow/*-reviewer-prompt.md`,
  `skills/flow/reviewer-calibration.md`) would have caught, or a **false positive**, a finding later
  disproved or withdrawn that cost a fix round for nothing. A visual-verify finding's reviewer
  is the verifier: its miss or false positive targets `skills/flow/visual-verify-verifier.md` or
  `skills/flow/visual-verify.md`.

A **late-round miss** — a finding a slot first raises in round 1 or later on a hunk that same slot
already reviewed, unchanged, in an earlier round — is traced to why the earlier rounds missed it,
whatever its severity: the pass compares that slot's earlier dispatches with the round that raised
it — the slot's reviewer prompt file at the change's merge base, the diff each was given, the
findings it returned. A generic instruction gap is a `reviewer` miss case whose `why_picked` names
the rounds that missed it; any other cause — a re-run shown only the delta, or reviewer variance the
bundle cannot explain — is an angle-4 finding naming that cause, classified by step 3's
classification.

The visual-verify step is analysed on both of its halves, because its fix rounds are where a run
most often spends repeated rounds: its prompts, by the cases above, and its tooling — the capture,
the mockup comparison, the scripts and browser setup `skills/flow/visual-verify.md` drives. A
tooling gap that cost a visual-verify fix round — a capture that missed the state the defect was in,
a comparison that could not see it, a setup step that failed and was redone — is reported under
angle 4 as a finding, never as a case, and is filed with its angle's task in step 5.

Every case carries a `scope` along two independent axes, language and layer (`frontend` or
`backend`), each pinned or left open:

| `scope` | The rule holds in | Target |
|---|---|---|
| `generic` | every language, frontend and backend | the shared prompt under `skills/` |
| `language:<lang>` | one language, frontend and backend | `rules/lang-<lang>.mdc` |
| `layer:<frontend\|backend>` | one layer, every language | `rules/layer-<frontend\|backend>.mdc` |
| `language:<lang>/layer:<frontend\|backend>` | one language on one layer | `rules/lang-<lang>-<frontend\|backend>.mdc` |

`<lang>` is the lowercase language name (`kotlin`, `go`, `typescript`). An axis is pinned only when
the rule would be wrong or meaningless without it: a rule true for both layers of a language leaves
the layer open, and a rule true in every language on one layer leaves the language open. The test,
per axis: would the rule still be true, word for word, in a project on the other layer, or in
another language? If yes, that axis stays open. The pass picks the widest scope that is still true.
Every `rules/` target is an opt-in rule file a project receives only by naming it in its
`<project>/.flow/project.md` `## standards`; a scoped rule never goes into a shared prompt.

A case is picked only when its `why_picked` states all four:

1. The defect class in one sentence, with no names from the task.
2. The current prompt text quoted, showing that no rule covers the class — the case's `target` and
   its `before` are that quote.
3. Why the update generalises: other code it would apply to, or other runs that show the class.
   Being one vivid incident is never enough.
4. Why the update is not patch accretion: it states a principle, not a special case.

**Zero cases is a normal result** — the section then carries the none-marker like any empty angle.
The pass never lowers the bar to produce a case; a finding that fails any of the four is not
picked and not mentioned. Every case is recorded `noted` — never fixed in step 3 — its report
line naming its folder, and step 5's `flow-prompt-tuning` task names each case folder:

```
- **[flow-prompt-tuning]** <defect class> — <type> — docs/prompt-tuning/<case-id>/ — noted
```

The pass hands back each case's files, step 2a's table save `index.html`, for `<sr-dir>/cases/<case-id>/`, and this session writes them there.

An `In-run pipeline fix:` line in the bundle (**Pipeline defects found mid-run**,
`skills/flow-contracts/pipeline.md`) that names a sha is reported under angle 1 as fixed and
is never offered for filing; one that reads `deferred` is an angle-1 finding like any other.

**A finding is fixed or filed only from the four angles.** A `flow-fix` finding is
fixed in step 3 unless it is `big`, and offered under its angle when it is. A `flow-cost`,
`flow-improvement` or `flow-prompt-tuning` finding is never fixed in the run: it is offered, for its
angle's task in step 5. An observation with nothing to fix is reported and recorded `noted`, never
offered. **An observation that names
an action the run can take is a fixable finding, never `noted`**, unless the action is `big`, reaches production
(`rules/never-touch-production.mdc`) or deletes what cannot be restored. **The pass covers
the pipeline and the project's own dev tooling — build scripts, dev-stack and test-harness
config, guards — and nothing else.** A finding about the project's own product code is out of
scope: it is never offered or filed, and goes to a fix run instead — a pass inside a run stops
here, at the end of this step (**the sweep**, `skills/flow-contracts/known-bugs.md`). A landing HEAD
whose UI no visual verifier saw is such a finding, never a risk line. The draft carries no
section beyond the four angles.

**One combined pass** — never four separate reads. The pass covers what the bundle holds and
nothing beyond it.

### 2a. Write the prompt-tuning cases

**Every case the pass returns is a deliverable, landed on `<agents repo>` `main` before step 8
reports it.**

A pass with no angle-4 case writes nothing here. Otherwise this session, after the pass and before
step 3, first resolves the worktree of `<fix-branch>` — the branch named at the end of this step —
creating it with `<agents repo>/scripts/flow-guard pipeline-fix start <name>` when it is
`self-review-<name>` and absent; it never writes the agents main checkout. It copies each `<sr-dir>/cases/<case-id>/` folder the pass wrote into that
worktree — `<agents repo>-worktrees/<name>/docs/prompt-tuning/<case-id>/` on the change's own branch,
`<agents repo>-worktrees/self-review-<name>/docs/prompt-tuning/<case-id>/` otherwise, `<case-id>` being
`<name>-<n>`, `<n>` counting from 1 within the change, each holding the files the pass wrote:

| File | Content |
|---|---|
| `case.json` | `id`, `type` (`implementer`/`reviewer`), `reviewer_kind` (`miss`/`false-positive`, reviewer only), `scope` (one of the four forms of angle 4's scope table), `source` {`project`, `change`, `jira`, `repo`, `fix_sha`, `finding_ref`, `round`, `slot`, `severity`}, `target` (agents-repo-relative prompt path), `target_sha` (the agents commit the `before` text was read at), `defect_class`, `why_picked`, `before`, `after` |
| `finding.md` | the reviewer's own sentence, severity, location |
| `bad.diff` | the code as the implementer first wrote it: the hunk(s) the finding names, from the commit before the fix |
| `good.diff` | the fix-round commit's diff for those hunks (`git show <fix_sha>`) |
| `task.md` | the task text as `tasks.md` recorded it (implementer cases; `none` for reviewer cases) |
| `index.html` | rendered by `flow prompt-tuning render <case-dir>` from the files above, never written by hand |

`bad.diff`, `good.diff` and `task.md` are extracted with `git -C <repo> show <sha>:<path>`. A `reviewer` case whose `target` is `skills/flow/visual-verify-verifier.md` or
`skills/flow/visual-verify.md` also carries `capture.png`, the capture the verifier judged, and
`mockup.png`, the mockup frame it judged that capture against, both copied into the case folder.
Once a case's folder is copied, run `flow prompt-tuning render <case-dir>`.

The cases are committed as `docs(prompt-tuning): <name> cases` on the branch step 3 names as
`<fix-branch>`, always `self-review-<name>`. When step 3 has no
worktree — no non-`big` finding — the worktree this step created carries the cases alone: this session commits them there and lands that
cases-only branch by step 3 item 5's route.

### 3. Fix every `flow-fix` finding that is not `big`

**Every finding is classified before any work**, by the pass, per the blast-radius rule of **Pipeline defects found
mid-run** (`skills/flow-contracts/pipeline.md`): `big` when its fix touches 60 or more files — a
design choice never is, since the planner takes the recommended option, and neither is a
reach outside `<agents repo>`, whose fix goes where that section places it. An
`In-run pipeline fix:` line naming a sha is already fixed and is not classified again.

**A finding whose fix adds a guard or an automation is never built by the pass.** Any finding
whose fix would add or extend a guard, check, hook, refusal, lint rule, script or other automation
is filed instead, all of them together as
**one** Jira task per pass: its summary `Self-review <name>: guard and automation proposals`, its
description one `## Why` / `## How` / `## DoD` block per finding (**Labels and body of issues the
pipeline creates**, `skills/flow-contracts/jira-integration-finish.md`), each finding a `###`
heading inside each section. The operator reviews it by hand and decides; the pass records each
such finding `filed` under that issue's key and never asks about it.

**Every other `flow-fix` finding that is not `big` is fixed and landed without asking**, on one branch for the
whole pass. A pass with none creates no worktree and dispatches nothing. No scope or cost question
with a recommended option is put to the operator anywhere in this pass, as a prompt or as prose in
a message (**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`).
**The planner never drops a
non-`big` finding because its first design would misfire** — a guard that would flag the
pipeline's own commits, a check that would refuse legitimate history. That objection is a design
constraint: the planner designs within it, measures the result against the history it named, and
plans the fix. A finding leaves step 3 unfixed only when it is `big`, or disproved — the planner
shows, with evidence, that the defect it names does not exist. A planner report that drops a
finding on any other ground is sent back to a fresh planner dispatch carrying the objection as a
constraint, never filed in step 5.
**A fix repairs what its defect already wrote wrong** — a store row, a record or a file the defect
left corrupted is repaired in the same plan, a small command added where none exists, never left
as a forward-only fix with the repair an operator's option.
A finding a re-plan
did not solve — the same finding back on the same evidence — never leaves the loop: the next re-plan, key
`self-review-<name>-plan-<n>`, stays on the planner pair and changes approach under **Fewest operator actions**
(`skills/flow-contracts/pipeline.md`) — never a stop and never an ask. Three failed rounds of the
review loop or of the verify loop are answered the same way, by the next re-plan (**Three failed
rounds**, `skills/flow-contracts/pipeline.md`), its prompt carrying every `review-<r>.md` and
`verify-<v>.log` of those rounds. Every planning dispatch runs on
the `self-review-plan` role's pair (`pipeline-fix-plan` in a pipeline-fix loop), never an escalation. **This session plans,
changes and reviews nothing itself** (`design.md`, `one-shot-subagent-fix-loop`), save item 4's
ticks and **The base sync** below: each numbered step below that does is a fresh dispatch on the pair it names, never a fork of this session (**Dispatch sites — the parent's closed list**, `skills/flow/implement.md`), its prompt carrying the TOOLS, FOREGROUND BUILDS and
NO DELEGATION paragraphs of `skills/flow/implementer-brief.md` verbatim — a review's READ-ONLY REVIEW
(`skills/flow/gated-reviewer-brief.md`) too, and the plan's
PLANNING DELEGATION below in place of NO DELEGATION. Every dispatch is one-shot: no `SendMessage` to it once
it returns. The `self-review-fix` row carries no no-progress escalation: a self-review fix that makes no progress is answered by the fresh re-plan above. The `pipeline-fix-fix` row keeps it: a pipeline-fix round that made no progress is dispatched again on the `no-progress-fixer` role (**Fewest operator actions**, `skills/flow-contracts/pipeline.md`).

**The pairs are the `self-review-*` and `pipeline-fix-*` rows** of **Allowed pairs**
(`skills/flow-contracts/model-policy.md`), a key's suffix naming its row — `pass`, `plan`, `fix`, `review-1` and `review-r` for
every re-review.

**Each loop dispatch runs on its row's role's pair**, per **Model and effort**
(`skills/flow/brainstorm-planner.md`) — **the planner pair** on the `plan` row, the
`flow-planner-<effort>` definitions being the `flow-<effort>` ones plus `Agent` (decision
`planner-carries-agent`); **the fixer pair** on the `fix` row; and **the review pair** on the two
`review` rows.

**Every loop dispatch is recorded mechanically.**
`<agents repo>/hooks/enforce-loop-dispatch.py` hands each Agent call whose prompt carries a
`Key: pipeline-fix-<k>-…` or `Key: self-review-<name>-…` line to
`<agents repo>/scripts/flow-guard check-loop-dispatch`, which records it in flowd. The parent closes each dispatch
once it returns with `<agents repo>/scripts/flow-guard pipeline-fix end <key> <outcome>`: `completed`, or
`aborted` for one that did not finish or never ran — a dispatch denied after it was recorded is
closed `aborted` and may be dispatched again under the same key. A review is closed with
`<agents repo>/scripts/flow-guard pipeline-fix end <key> <report-file>`, its reviewer's report, whose `Critical: <n>` and
`Important: <n>` lines decide the outcome: `clean` when both are 0, `fix` otherwise. For an own-branch loop in another project, `flow-guard pipeline-fix end`
runs from that project's working directory, never from `<agents repo>`. A refusal or failure of any of these tools on a step this file prescribes is a
pipeline defect, fixed in-run per **Pipeline defects found mid-run**
(`skills/flow-contracts/pipeline.md`) and the loop resumed once the fix lands — never a stop or a
question to the operator.

`<agents-base>` is the one **Pipeline defects found mid-run** binds. `<fix-branch>` is always
`self-review-<name>`, based at `origin/<agents-base>`, landed by `flow-guard pipeline-fix land`; fixes never go onto the change's own branch.

1. **Plan, then fix** — this session creates the worktree with
   `<agents repo>/scripts/flow-guard pipeline-fix start <name>` — never the main checkout. It then dispatches key `self-review-<name>-plan` on the planner pair, its
   prompt naming `<sr-dir>/findings.md` and the non-`big` finding ids, whose evidence and counted blast radius that file carries. The planner
   plans them all in full as one spectre change, `self-review-<name>`, in that worktree, per
   `skills/flow/brainstorm-planner.md`: brainstorming (section B), the artifacts (section C —
   `proposal.md`, and `design.md` with `## Context`, `## Decisions` and `## Open questions`) and
   writing plans (section D — a `tasks.md` in which each finding is carried by exactly one task). The `flow-planner-<effort>` agent types, like the
   `flow-<effort>` definitions they copy, carry no Skill tool, so the prompt names the path of the installed superpowers plugin's
   `skills/brainstorming/SKILL.md` and `skills/writing-plans/SKILL.md` — on Claude Code the
   newest `~/.claude/plugins/cache/*/superpowers/*/skills/<skill>/SKILL.md` — and the planner reads and
   follows each file where those sections invoke its skill. **Every question it meets is
   auto-resolved to its recommended option** per **Auto-resolution**
   (`skills/flow-contracts/operator-prompts-auto-resolution.md`) and recorded as a decision: it
   never asks the operator. Section B's design file goes under
   `<agents repo>-worktrees/self-review-<name>/.superpowers/sdd/`, or
   `<agents repo>-worktrees/<name>/.superpowers/sdd/` on the change's own branch. When it writes
   or changes a prototype, its report names the file's absolute path, and this session publishes it
   and prints its URL in its final report (**The checklist**, `skills/flow/brainstorm-planner.md`). Section D's
   `spectre validate "self-review-<name>"` passes before it hands back. It runs none of what records or gates a pipeline run: no `flow stage` call, no **Plan check gate** — no `decision.json`, no `## Decision` block — and no **Plan
   review gate**. `design.md`'s decisions record instead the first review's pair and the grouping
   of tasks into fixer dispatches. The pair is the entry `**ID:** first-review-effort`,
   its `**Chosen:**` line the `self-review-review-1` role's pair (`pipeline-fix-review-1` in a
   pipeline-fix loop), written `` `flow-<effort>` on `<model>` ``. It commits
   none of the artifacts — `<agents repo>/spectre/changes/` is gitignored — and never merges or pushes.
   A run-1 re-run never folds the loop's fix commits: the newest commit a `fixed` row names is its
   reshape base (`skills/flow-contracts/finish-contract-run1.md`).

   The plan dispatch carries, in place of NO DELEGATION:

   > **PLANNING DELEGATION:** You may dispatch fresh `Explore` and `Plan` subagents through the
   > `Agent` tool while you plan, per **B. Basic Workflow #1 — Brainstorming** (`skills/flow/brainstorm-planner.md`) — never a
   > `fork`, and never an agent that edits files or commits. Each dispatch runs on its own Role's pair
   > (**Model and effort**, `skills/flow/brainstorm-planner.md`), and every dispatch prompt carries the line
   > `Role: explorer` (an Explore agent) or `Role: plan-designer` (a Plan agent) and the two sentences of **Propagate this**
   > (`~/.claude/rules/agent-baseline.md`) verbatim. The agents investigate and propose; you
   > decide, and every choice is auto-resolved to its recommended option. Fixing, implementing and
   > reviewing are never delegated.

   Then one fixer per recorded group, in series — key `self-review-<name>-fix`, `-fix-g<g>` when
   the planner recorded more than one group — on the fixer pair, its prompt naming its tasks. It
   implements them in that worktree, one commit per task with its declared `**Commit:**` subject
   and no `Task-Id:` trailer, never writing `<agents repo>/spectre/changes/` — the ticks are item
   4's — adding one test that fails without the fix wherever the fix changes behaviour, and running
   the `<agents repo>/.flow/project.md` `## lint` lines its files need. It reports each task's
   commit, or a finding it found `big` once under way, left uncommitted. It never merges or
   pushes.
2. **Review** — key `self-review-<name>-review-<r>` on the review pair, over
   `git diff <fix-base>...<fix-branch>`, its prompt naming each finding beside its task's
   commit, judging the code against `tasks.md` and `design.md`'s decisions, and asking each review
   finding graded Critical, Important or Minor, and its report to open with one `Critical: <n>`
   and one `Important: <n>` line counting them.
   Its prompt names `<sr-dir>/review-<r>.md` as its report file: it writes its report there and hands back only those two lines, and this session closes it with `<agents repo>/scripts/flow-guard pipeline-fix end <key> <sr-dir>/review-<r>.md`.
3. **Fix the review's findings** — key `self-review-<name>-fix-<r>` on the fixer pair, its prompt naming
   `<sr-dir>/review-<r>.md`. Each fix is folded into the commit of the task it fixes
   (`git commit --fixup <that commit>`, then `git rebase --autosquash <fix-base>`), so
   every task stays one commit; a commit the review judges not to fix its finding, or to make
   things worse, is dropped from the branch. **A review whose every finding is Minor is closed by
   that one dispatch, with no re-review** — the no-re-review half of **Panel re-runs**
   (`skills/flow/review-panel.md`) — and step 4 is its verification. After a review that raised a
   Critical or Important, step 2 runs again, `<r>` plus one, until a review comes back clean or
   raises only Minors, under **Fewest operator actions** (`skills/flow-contracts/pipeline.md`). A
   dropped commit's finding is filed in step 5 as a `big` one is; so is a fix found `big` once
   under way.
4. **Verify** — once a review comes back clean, or a Minor-only review's fix is committed, this session ticks each task whose commit is on `<fix-branch>`
   (`flow tasks tick -C <fix worktree> self-review-<name> <n>`), then runs
   `<agents repo>/scripts/flow-guard pipeline-fix verify <fix worktree> > <sr-dir>/verify-<v>.log 2>&1`, `<v>` counting verify runs from 1, and reads only its exit code and its `red:` lines.
   On a `self-review-<name>` branch whose route is merge and push, that separate verify is not run:
   step 5's `<agents repo>/scripts/flow-guard pipeline-fix land <fix worktree> > <sr-dir>/verify-<v>.log 2>&1` runs the same lint
   lines and tests on the rebased HEAD before it pushes, and its exit code and `red:` lines are read
   as this verify's.
   A red line is fixed by a fresh
   dispatch of step 3, key
   `self-review-<name>-fix-v<v>`, its prompt naming the failing command and `<sr-dir>/verify-<v>.log`; then step 2, then this step's lint lines and tests again, `<v>` plus one. A red branch never lands, and a
   red line is never waived — **Fewest operator actions** governs this loop as it does the review.
5. **Land once.** Land by `<agents repo>`'s `## default landing route`
   (`flow-guard project-get <agents repo> 'default landing route'`), without asking — merge and push as
   step 4 of **Pipeline defects found mid-run** states it, with `self-review-<name>` as the
   branch. The pass is not done until every fixed finding's commit is on `origin/<agents-base>`. Each fixed
   finding's sha is read off `<agents-base>` after the landing, never from
   the branch before its rebase; on a route that opens a pull request, once that pull request
   merges.

**The base sync.** A finding that a `--base` change's `<base>` (`skills/flow/SKILL.md`) lacks a
commit the project's `<default-branch>` carries is no task of the loop: in each repository of the
change whose `<base>` lacks it, this session adds a detached worktree of `origin/<base>`, named
`base-sync-<name>`, in that repository's worktrees directory after a `git fetch origin`, runs
`git merge --no-edit origin/<default-branch>` in it, then the `## lint` and `## test` commands
`<project>/.flow/project.md` declares for that repository — a repository with no declared `## lint`
and `## test` command has no green run; on a clean merge and a green run it
pushes the merge to `<base>` on `origin` without asking, removes the worktree, and records the
finding `fixed` with the merge commit's sha. A merge that conflicts, a red or absent run, or a
rejected push pushes nothing,
removes the worktree, and is filed in step 5 as a `big` finding is.

### 4. Explain

Every finding is explained in the message body first, before anything is filed — what was
observed, what breaks, and what the fix would be. Every fixed finding is one line in the same
body, naming its landed sha; it is not offered.
This session first fills each `fixed: <F-id>` placeholder of `<sr-dir>/report-draft.md` with the sha `git log --format='%h %s'` shows for its task's commit subject, rewriting one whose finding step 3 left unfixed as `offered: <F-id>`.
It then prints that body from `<sr-dir>/findings.md`'s `## Offered` section and the draft's `fixed:` lines, never composing either.

The explanation message renders in Markdown, in one table, every finding a row: `#`, observed, what breaks, fix (sha).

### 5. File the offered findings

Nothing is asked: every offered finding is filed. All of the pass's `big` findings, and every
`flow-fix` finding step 3 or **The base sync** left unfixed, are filed together as **one** Jira issue, its summary `Self-review <name>: big findings`. Each of
`flow-cost`, `flow-improvement` and `flow-prompt-tuning` with a finding or, for
`flow-prompt-tuning`, a case is filed as **one** Jira task of its own, its summary
`Self-review <name>: <label> findings`, the `flow-prompt-tuning` task naming each case folder
(`docs/prompt-tuning/<case-id>/`). Each issue's description is one `## Why` / `## How` / `## DoD`
block per **Labels and body of issues the pipeline creates**
(`skills/flow-contracts/jira-integration-finish.md`), each finding a `###` heading inside each
section, and it carries each of its findings' angle labels on top of that label set. Each issue,
once created, is transitioned to the status named **Self-review**, the transition resolved by that
name per **Transitions** (`skills/flow-contracts/jira-integration.md`); a transition that fails or
is not offered is one `⚠ Jira: skipped — <reason>` line and the issue stays filed.
`## jira` absent or `none` in `<project>/.flow/project.md`, or no Atlassian tooling available in
this session: print `⚠ Jira: skipped — <reason>` and record those findings `declined` instead.

### 6. Record every finding in the flow store

First this session fills each `offered: <F-id>` placeholder of `<sr-dir>/report-draft.md` with `filed: <KEY>` from step 5 or `declined`.
It reads the filled draft once, here, and records the rows from its lines.

One call per finding, every disposition alike, before step 7 — save a `fixed`
finding:

```bash
flow self-review finding -change <name> -angle <label> -disposition fixed|filed|declined|noted [-ref <sha|KEY>] [-blast-radius <N>] [-severity critical|important|minor] [-theme <id>] -note '<the finding, one line>'
```

`-ref` is the landed sha for `fixed`, the issue key for `filed`, and omitted for `declined` and `noted`;
`-blast-radius` is step 3's count; `-severity` is the grade the pass gave the finding, required for `fixed` and `filed`, which `flow self-review finding` refuses without it, and omitted for `declined` or `noted` when the pass gave none; `-theme` is the theme the pass named, on every `flow-cost` row whatever its disposition — `flow self-review finding` refuses a `noted` one without it, and a `fixed`, `filed` or `declined` row carrying a theme settles it for 30 days. An `In-run pipeline fix:`
line naming a sha gets no second row when `flow self-review findings -change <name>` already
lists that sha (**Pipeline defects found mid-run**, `skills/flow-contracts/pipeline.md`). A row this step records for such a line, which the pass never grades, takes `-severity` by the rule of **Pipeline defects found mid-run** (`skills/flow-contracts/pipeline.md`). A store failure is one
warning line and the pass continues to the listing below.
A row is stated as recorded only once `flow self-review findings -change <name>` lists it: read that listing before step 8, record every row it lacks again, and fix a store that still refuses per **Pipeline defects found mid-run** (`skills/flow-contracts/pipeline.md`).

### 7. Delete the scratch directory

No report is written or committed: the findings live in flowd. Delete `<sr-dir>` once step 6's
listing holds every row and item 5 of step 3 recorded its `fixed` rows.

### 8. Report

End naming the shas landed, the Jira keys filed, each prompt-tuning case written in step 2a by the absolute path of its `index.html` in `<agents repo>`'s main checkout once landed (`<agents repo>/docs/prompt-tuning/<case-id>/index.html`) (each `none` when empty), and each product defect the pass met (**the sweep**, `skills/flow-contracts/known-bugs.md`).

## Guardrails

- **Never** push directly to a protected branch under any name but the project's own default
  branch, or the `<landing-branch>` step 1 accepted, and never force-push.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Run a deferred self-review pass for `<name>` | `/flow-self-review <name>` |
