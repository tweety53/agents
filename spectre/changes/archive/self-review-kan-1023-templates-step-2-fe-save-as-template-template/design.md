## Context

One self-review fix change for gymie kan-1023, planned from the pass file named in `proposal.md`.
Every finding there that is not `big` is planned here; each was reproduced or traced before its fix
was designed. Brainstorm working copy:
`.superpowers/sdd/2026-10-25-self-review-kan-1023-templates-step-2-fe-save-as-template-template-design.md`
(gitignored).

**F1, reproduced.** A scratch Go run of `uwSpecRE` (`stats/internal/guard/unfinishedwork.go:1070`)
over ``- failed (kan-341-route-b.spec.ts:224:5) and `fidelity/a.spec.ts:3` [x/b.spec.ts:9]
'c.spec.ts:1' "d.spec.ts:2"`` captured `(kan-341-route-b.spec.ts`, `` `fidelity/a.spec.ts``,
`[x/b.spec.ts`, `'c.spec.ts` and `"d.spec.ts`. The citation match at line 568 and the reason at
line 599 both use the captured spec, so the demanded citation carries the stray character.

**F2, confirmed.** Step 2 of **The loop** (`skills/flow/verify-fix-loop.md`) requires a permanent
probe per fixed departure and nothing for a branch the fix adds beyond it; the gymie run's Tasks
13–18 each declared `**Regression:** each probe fails with its fix reverted`, and panel rounds 3,
5, 7 and 9 raised Importants on untested branches those fixes introduced.

**F3, disproved.** `flow record decisions -change kan-1023-templates-step-2-fe-save-as-template-template
-C /Users/tweety53/Projects/gymie-worktrees/kan-1023-templates-step-2-fe-save-as-template-template`
reads `"class":"big"` on both rows (317 and 318). `flow record dispatches` holds `task-13-reviewer`
through `task-18-reviewer`, so every appended task's gate fired. `stats/internal/guard/closetask.go`
line 168 leaves a fired task unticked on `big`, and **The gated per-task reviewer**
(`skills/flow/implement.md`) has the parent tick each clean task in the call that closes the
reviewer's row — the parent's tick the narrative recorded is that step, not a workaround.

**F4, traced** in the gymie run's parent transcript
`/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/e905b265-37ae-40b2-9776-72e831e83ebc.jsonl`
against `stats/internal/harvest/watcher.go` and `attribute.go`. Three causes:

- **Row 7560** (`task-13-implementer`, 1h31m). Its launch (transcript line 1868, agent
  `afdfbfaf17eebe758`) was described `Task 13/13 (visual-verify fixes)`, the shape **Every
  dispatch's Agent-tool `description`** (`skills/flow/implement.md`) prescribes for task
  dispatches. `AgentLaunch.Key` is the description's first word (`transcript.go:590`), so the keyed
  stamp looked up the key `Task` and matched nothing; the time pairing then retired the row's begin
  (02:21:33Z) on the two denied Agent calls at 02:22:05Z and 02:22:32Z, each nearer than the real
  launch (02:22:58Z). No row holds `afdfbfaf17eebe758`.
- **Rows 7562, 7567, 7577, 7613, 7621** (`task-13-implementer-fix-1`, `panel-fix-3`,
  `task-14-implementer-fix-2`, `panel-fix-11`, `panel-fix-12`). Each was the parent's own inline
  work, begun without the `-agent-id inline` the skills prescribe; no Agent launch falls between
  its begin and its end. `DispatchAttributor.Attribute` gives main-thread records to inline windows
  only, so these rows can never be measured.
- **Rows 7616, 7624** (`verify-fix-5-gymie`, `verify-fix-6-gymie`, inline). Each window is 8 s
  inside one Bash call that began and ended it; no parent record lies inside it — the issuing
  message precedes `started_at` and sits in the overlapping inline row 7615's (7623's) window. The
  share is zero, but nothing ever writes a figure, so the ledger renders "not measured".

## Decisions

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-low — the operator's cap: opus at low effort for every flow dispatch until
2026-10-28.
**Considered:** flow-medium, the default first review when none is recorded — over the cap.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, Tasks 1–4 in plan order, one fixer dispatch — four small fixes in one
repository; Tasks 3 and 4 share `watcher.go` and `records.go` and must run in series anyway.
**Considered:** one group per finding — parallel fixers on Tasks 1 and 2 save minutes but add
dispatches against the operator's low-effort cap, and Tasks 3–4 cannot split.

### Spec regex excludes opening punctuation

**ID:** spec-regex-excludes-opening-punctuation
**Status:** active
**Chosen:** exclude `(`, `[`, backtick, `'` and `"` from `uwSpecRE`'s spec character class — a
Playwright spec path never contains them, and the capture then is the spec alone.
**Considered:** trimming the captured spec after the match — a second step for what the class
states once; normalising the narrative citation instead — the reason line would still print the
stray character.

### Visual fix tests every new branch

**ID:** visual-fix-tests-every-new-branch
**Status:** active
**Chosen:** one sentence in step 2 of **The loop** (`skills/flow/verify-fix-loop.md`): every
branch a visual fix adds beyond its departure carries a test that fails with it removed, named in
the task's `**Regression:**` beside the probe.
**Considered:** relying on the MUTATION PROOF paragraph already in the dispatch — it runs after the
fix and found these defects only at the panel; a guard parsing `**Regression:**` for branch
names — no mechanical way to know what a branch is.

### close-task waits on big by design

**ID:** close-task-big-waits-by-design
**Status:** active
**Chosen:** no change to `close-task.sh` — F3 is disproved by the evidence in `## Context`.
**Considered:** ticking fired tasks on `big` in close-task — it would tick a task before its
gated reviewer passed, which **The guard's pass ticks an ungated task.** (`skills/flow/implement.md`)
forbids.

### Task description stamps by prefix

**ID:** task-description-stamps-by-prefix
**Status:** active
**Chosen:** a launch described `Task <ids>/<n>` or `Tasks <ids>/<n>` carries the key prefix
`task-<ids>-`; `StampSessionDispatchAgent` treats a key ending in `-` as a prefix and stamps the
session's newest agentless row under it; begins under a used prefix are dropped from time pairing
— the same row `hooks/enforce-agent-baseline.py` already resolves that description to.
**Considered:** changing the prescribed description to open with the key — the `flow-task-list`
mod and the operator's status-line shape read the `Task x/n` prefix; loosening the denial
retirement — it guards a panel round's denied slot and would mis-stamp there.

### Settle closed rows

**ID:** settle-closed-rows
**Status:** active
**Chosen:** a watcher settle pass at each committed batch: a begun row the store reports ended,
harvested past its end, is settled once — agentless with no Agent launch or denial between its
begin and its end → stamped `inline`; every inline row so settled is re-measured over the
transcript up to the committed offset, a row with no share written as zero tokens.
**Considered:** a mandatory `-agent-id` on `dispatch begin` — 29 test begins and five skill sites
change, and every refusal costs the parent a turn; treating a row still agentless at its end as
inline — races the harvester on a foreground agent whose launch line lands seconds before the end;
opening an inline window at its issuing message — double-counts a record an overlapping inline
row already holds (rows 7615/7616).

### Repair the eight rows

**ID:** k1023-rows-repair
**Status:** active
**Chosen:** `flowd restamp-dispatch` widened so a parent transcript also repairs an agentless row
(marked inline first) and writes a zero share; then row 7560 restamped from agent
`afdfbfaf17eebe758`'s transcript and rows 7562, 7567, 7577, 7613, 7616, 7621, 7624 from the parent
transcript — a forward fix alone leaves the eight rows wrong (**A fix repairs what its defect
already wrote wrong**, `skills/flow-self-review/SKILL.md`).
**Considered:** hand-written SQL against the live store — no test pins it and the next such repair
writes it again.

## Open questions

## Live check

From this worktree with every task committed (`cd stats`, `make -s web-dist` first), before
landing; the protected `flowd` on `127.0.0.1:4173` is never restarted, and no test runs against
the live database:

1. **Rows repaired.** `docker exec flow-postgres psql -U flow -d flow -tc "select id,
   coalesce(agent_id,''), metrics ? 'tokens' from dispatches where id in
   (7560,7562,7567,7577,7613,7616,7621,7624) order by id"` — before: every agent id empty but
   7616/7624 (`inline`), every `tokens` false; after: 7560 `afdfbfaf17eebe758`, the other seven
   `inline`, every `tokens` true. Not working looks like any `f`.
2. **Ledger.** With `G=/Users/tweety53/Projects/gymie-worktrees/kan-1023-templates-step-2-fe-save-as-template-template`,
   `flow record render -change kan-1023-templates-step-2-fe-save-as-template-template -kind ledger
   -repo $G -C $G`, then `grep -c 'not measured' $G/.superpowers/sdd/ledgers/kan-1023-templates-step-2-fe-save-as-template-template.md`
   — before: 8 (measured 2026-10-25); after: 0.
3. **Settle pass.** Stand-in for the live daemon: the Task 4 watcher tests run against a scratch
   transcript and fake deps (`go test ./internal/harvest -run 'TestWatcherSettles' -count=1`);
   not working looks like an agentless closed row left empty or an inline zero-share row left with
   no `tokens`.
