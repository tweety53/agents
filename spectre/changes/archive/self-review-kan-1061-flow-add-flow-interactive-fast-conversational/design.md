## Context

The self-review fix loop of kan-1061, planned on the change's own branch
`spectre/kan-1061-flow-add-flow-interactive-fast-conversational` in
`/Users/tweety53/Projects/agents-worktrees/kan-1061-flow-add-flow-interactive-fast-conversational`;
its fix commits ride that run's own route. The pass is `.superpowers/sdd/self-review-pass.md` in
that worktree; the brainstorm trail is
`.superpowers/sdd/2026-10-30-self-review-kan-1061-flow-add-flow-interactive-fast-conversational-design.md`.
Every question was auto-resolved to its recommended option.

**What `begin` checks today.** `runRecordDispatchBegin` (`stats/cmd/flow/record.go`) refuses a
`panel-` reviewer key with an empty `-slot` and nothing else about the slot; the store
(`ApplyDispatchRecord`, `stats/internal/api/records.go`) checks only that each `+`-joined part is a
panel slot. A back-fill calls `runRecordDispatchBegin`, so one check there covers both. Its success
line `recorded: dispatch <seq>` is read by `skills/flow/review-panel.md` and two guard tests as the
first stdout line.

**kan-1061's five rows, as stored.** Dev store, read-only `psql` in `flow-postgres`:
<!-- measured: docker exec flow-postgres psql -U flow -d flow -At -c "select d.id, d.seq, d.dispatch_key, d.slot from dispatches d join changes c on c.id=d.change_id where c.name='kan-1061-flow-add-flow-interactive-fast-conversational' and d.role='reviewer' and d.dispatch_key like 'panel-%' order by seq" @ dev store, 2026-10-30 (live data, not re-runnable at a ref) -->
row id 8830 (`panel-3-test-sanity+call-efficiency`, slot `test-sanity`), 8831
(`panel-3-mutation+failure-modes`, `mutation`), 8834 (`panel-4-mutation+failure-modes`,
`mutation`), 8847 (`panel-5-test-sanity+call-efficiency`, `test-sanity`) and 8848
(`panel-5-mutation+failure-modes`, `mutation`). Rounds 0 and 2 carry the joined slot.

**What the slot rule would have refused.** Over the last three weeks of reviewer panel rows, the
keys whose remainder after `panel-<r>-` is neither the slot nor the slot followed by `-` are,
besides kan-1061's five, only keys off the **Key shape** column of the Dispatch sites table
(`skills/flow/implement.md`): `panel-1-rerun`, `panel-3-extra`, `panel-2-verify`,
`panel-1-reauthor`, `panel-3-primary-principles`, and two `-cite` rows recording one role of a
joined key. Suffixed contract keys — `panel-9-primary-kan-924-…`, `panel-6-mutation+failure-modes-rerun`,
`panel-0-principles-repro` — pass.
<!-- measured: docker exec flow-postgres psql -U flow -d flow -At -c "select dispatch_key, slot from dispatches where role='reviewer' and dispatch_key ~ '^panel-[0-9]+-' and started_at > now() - interval '21 days' and regexp_replace(regexp_replace(dispatch_key,'^panel-[0-9]+-',''),'-(retry|cite)$','') <> coalesce(slot,'')" @ dev store, 2026-10-30 (live data) -->

**The Role line is not always begin's to know.** The Dispatch sites table gives a panel reviewer
row one of `panel-reviewer`, `panel-extra`, `panel-rerun` or `late-fix` by which pass the dispatch
is; `begin`'s flags carry no pass. Every other row's id follows from `-role` and the key.

**Hook tests run the hooks on the real PATH.** `runHook`
(`stats/internal/repotest/helpers_test.go`) passes `os.Environ()`, so a hook that shells `flow`
on every deny would write the dev store from the test suite. `enforce-agent-baseline.py` already
takes its pair checker from `DISPATCH_PAIR_CHECK` "for the hook's own harness only".

## Decisions

### A panel row's slot is its key's slot part

**ID:** sr1-slot-matches-key
**Status:** active
**Chosen:** `runRecordDispatchBegin` refuses (exit 2) a `reviewer` row whose key is a panel key (`records.PanelDispatchRound` ok) when the key's remainder after `panel-<r>-` is neither `-slot` nor begins with `-slot` followed by `-`, naming the key, the slot and `skills/flow/review-panel.md`. The predicate is `records.PanelKeyNamesSlot(key, slot string) bool` beside `PanelDispatchRound` in `stats/internal/records/render.go`, called by `begin` and by `reslot-dispatch` (sr1-reslot-dispatch).
**Considered:** deriving `-slot` from the key — a key's remainder may carry a worktree or attempt suffix (`-kan-924-…`, `-rerun`, `-repro`) whose split from the slot needs the slot itself, so the parent's `-slot` stays the input and the key checks it; checking only `+`-joined keys — a single-slot key under a joined `-slot` (`panel-3-extra`) is the same mis-credit; the store-side check in `ApplyDispatchRecord` — a replayed journal row of an older run would then fail, where `begin` sees only new rows.

### Repair the five rows with a reslot command

**ID:** sr1-reslot-dispatch
**Status:** active
**Chosen:** `func (s *Store) SetDispatchSlot(ctx context.Context, id int64, slot string) error` in `stats/internal/store/records.go` (`UPDATE dispatches SET slot = $2 WHERE id = $1 AND role = 'reviewer'`, `ErrDispatchNotFound` when no row matched) and `flowd reslot-dispatch <dispatch-id> <slot>` in `stats/cmd/flowd/main.go`, beside `reoutcome-dispatch`, which reads the row's key first and refuses a slot `records.PanelKeyNamesSlot` rejects. The fixer runs it on ids 8830, 8831, 8834, 8847 and 8848 once the task's commit exists, from the worktree's build against the dev store, and reads the rows back.
**Considered:** `flowd restamp-dispatch` — it rewrites agent and figures, never the slot; leaving the rows — kan-995's lesson, a forward-only fix leaves the store wrong.

### Carry the capture shape into the planner's capture-task paragraph

**ID:** sr2-capture-shape-at-plan
**Status:** active
**Chosen:** one sentence appended to the **Write a capture task that baselines against a mockup to measure before it commits** paragraph of `skills/flow/brainstorm-planner.md`, citing steps 8 and 9 of `skills/flow/visual-verify-verifier.md`: every capture is the full page or viewport, never an element or clipped region, and each captures a state no sibling capture shows, so no two come back byte-identical.
**Considered:** restating the verifier's whole step-8 paragraph — a second copy of a rule the verifier owns; a guard on `toHaveScreenshot` receivers — a spec may clip for its own assertion, which step 8 allows.

### Begin prints the launch lines

**ID:** sr3-begin-prints-launch-lines
**Status:** active
**Chosen:** after `recorded: dispatch <seq>`, a live `begin` (no `-started-at`, `-agent-id` not `inline`) that exits 0 prints `description: <opening>` and, where the row determines one, `Role: <id>`. The opening is `Task <x>/<n>` or `Tasks <a+b>/<n>` for a `task-<ids>-implementer…` or `task-<ids>-reviewer…` key — `<n>` the plan's column-0 task count (`countPlanLines(plan, taskLinePattern)`, `stats/cmd/flow/tasks.go`), the line omitted when the plan cannot be read — and the key itself for every other row. The Role line is `implementer` for a `task-*-implementer…` key, `gated-reviewer` for `task-*-reviewer`, `gated-re-reviewer` for `task-*-reviewer-fix-<k>`, `panel-merger` for role `merger`, `panel-fix` for role `panel-fix`, `visual-verifier` for a `visual-verify*` verifier, `tooling-analyst` for a `tooling-analysis-*` planner, and none for a panel reviewer row, a loop key (`self-review-*`, `pipeline-fix-*`, whose `Key:` line names its role) or any other row. `skills/flow/implement.md` tells the parent to paste both lines and, where no Role line is printed, to take the id from the Role line column.
**Considered:** guessing a panel reviewer's id from its round and slot — the first pass is the round `flow record next-round` prints, not round 0, and a later round's extras may be `panel-extra`, `panel-rerun` or `late-fix`, so a guess would put a wrong id the hook refuses into a pasted line; a new `-pass` flag — configuration nobody asked for, on a call already carrying the role.

### Each hook denial writes an incident row

**ID:** sr4-hook-denial-incident
**Status:** active
**Chosen:** in `hooks/enforce-agent-baseline.py` and `hooks/enforce-loop-dispatch.py`, `deny` writes its JSON first, then runs `flow record incident -C <payload cwd> -guard <enforce-agent-baseline|enforce-loop-dispatch> -symptom <the reason's first line, at most 300 characters> -recovery "re-sent per the deny reason" -minutes-lost 0`, plus `-change <name>` in the agent-baseline hook when `session_change` resolves one, with output captured, a bounded timeout, and every failure swallowed — the deny stands either way. The command is `FLOW_HOOK_INCIDENT_CMD` when set, else `flow`, an override for the hooks' own harness; `runHook` sets it to `true` by default, and the new tests point it at an argv-logging stub.
**Considered:** writing only in a flow session — a deny outside one is still a guard refusal whose count the pass wants; a fake `flow` first on PATH in every hook test — the agent-baseline hook already runs `flow record dispatches`, and a blanket fake would change what those tests exercise; the daemon logging denials — the hooks are the only party that sees them.

### The prompt-tuning cases are their own task

**ID:** sr5-cases-task
**Status:** active
**Chosen:** one task writes `docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-1/` and `-2/` with the files of step 2a's table (`skills/flow-self-review/SKILL.md`), the `case.json` bodies verbatim from the pass file's angle-8 section, renders each with `flow prompt-tuning render` and passes `scripts/check-prompt-tuning-cases.sh`, committed as `docs(prompt-tuning): kan-1061-flow-add-flow-interactive-fast-conversational cases`.
**Considered:** the self-review session writing them before step 3, as step 2a states — the dispatcher placed them in this plan, and one fixer writing them keeps this session changing nothing itself.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 1 changes what every panel `begin` admits and writes store rows, and Task 4 changes two hooks every dispatch passes through.
**Considered:** `flow-low` — the findings are not all mechanical.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, key `self-review-kan-1061-flow-add-flow-interactive-fast-conversational-fix`, Tasks 1–5 in plan order — Tasks 1 and 3 share `stats/cmd/flow/record.go` and run in order, and the five are small.
**Considered:** a separate docs group for Tasks 2 and 5 — the fix row's one pair is `opus/flow-low` either way, so a second dispatch buys nothing.

## Open questions

## Recommended approach

Mechanical refusals and outputs at the points a run already passes through: `begin` checks a panel
row's slot against its key (sr1-slot-matches-key) and prints the launch lines it already knows
(sr3-begin-prints-launch-lines); a `flowd` direct-database subcommand repairs the five rows the
defect wrote (sr1-reslot-dispatch); one planner sentence carries the capture shape
(sr2-capture-shape-at-plan); the two dispatch hooks record each deny as an incident
(sr4-hook-denial-incident); one docs task writes the two prompt-tuning cases (sr5-cases-task).
Deriving the slot from the key, guessing a panel Role line and a flow-session-only incident were
each rejected in their decisions.

## Critical files

- `stats/cmd/flow/record.go` — `runRecordDispatchBegin`: the slot check and the launch lines.
- `stats/cmd/flow/record_test.go` — begin's tests.
- `stats/internal/records/render.go` — `PanelKeyNamesSlot` beside `PanelDispatchRound`.
- `stats/internal/store/records.go`, `stats/internal/store/records_test.go` — `SetDispatchSlot`.
- `stats/cmd/flowd/main.go` — `flowd reslot-dispatch`.
- `skills/flow/brainstorm-planner.md` — the capture-task paragraph.
- `skills/flow/implement.md` — the paste instruction under the description rule.
- `hooks/enforce-agent-baseline.py`, `hooks/enforce-loop-dispatch.py` — the incident write.
- `stats/internal/repotest/helpers_test.go`, `stats/internal/repotest/enforce_agent_baseline_hook_test.go`, `stats/internal/repotest/enforce_loop_dispatch_hook_test.go` — the hooks' harness.
- `docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-1/`, `-2/` — the cases.

## Existing code to reuse

- `records.PanelDispatchRound` (`stats/internal/records/render.go`) — the panel key's round parse.
- `countPlanLines`, `taskLinePattern` (`stats/cmd/flow/tasks.go`), `planPath` — the plan's task count.
- `reoutcomeDispatch` / `Store.SetDispatchOutcome` (`stats/cmd/flowd/main.go`, `stats/internal/store/records.go`) — the direct-database repair shape `reslot-dispatch` copies.
- `session_change`, `bound` (`hooks/enforce-agent-baseline.py`) — the change lookup and the wall-clock bound; `DISPATCH_PAIR_CHECK` — the harness-override pattern.
- `runHook`, `openRowFixture` (`stats/internal/repotest/`) — the hook harness and its argv-logging fake.
- `flow prompt-tuning render`, `scripts/check-prompt-tuning-cases.sh` — the case render and guard.

## Verification

- Each task's targeted tests and lint, then `scripts/pipeline-fix.sh verify <worktree>` on the branch.
- `flow record dispatches -change kan-1061-flow-add-flow-interactive-fast-conversational` shows rows 8830, 8831, 8834, 8847 and 8848 with the joined slots after Task 1.
- A scratch `begin` of a new panel key with a single slot under its joined key exits 2; with the joined slot it prints `recorded: dispatch <seq>` and `description: <key>`.
- `scripts/check-prompt-tuning-cases.sh` exits 0 on the two new folders.

## Live check

Run at `flow.verify` on this worktree's build, never restarting the dev `flowd`:

- **The repaired rows (Task 1).** Before: the five rows above carry `test-sanity` or `mutation`. After: `flow record dispatches -change kan-1061-flow-add-flow-interactive-fast-conversational` shows `test-sanity+call-efficiency` on 8830 and 8847 and `mutation+failure-modes` on 8831, 8834 and 8848. Not working looks like any single slot still on those keys.
- **Begin's refusal and launch lines (Tasks 1, 3) against a scratch store.** `scripts/live-store-copy.sh up sr-kan-1061`, then, with the printed `FLOW_RECORDS_ADDR`, a `begin` of a new panel key under kan-1061 with a single slot under a joined key exits 2 naming the key, and one with its joined slot prints `recorded: dispatch <seq>` then `description: <key>`; a `task-1-implementer` begin prints `description: Task 1/<n>` and `Role: implementer`; `scripts/live-store-copy.sh down sr-kan-1061`. Not working looks like exit 0 on the single slot, or no `description:` line.
- **A hook denial's incident (Task 4).** A payload with no baseline pointer piped to `hooks/enforce-agent-baseline.py` with `FLOW_RECORDS_ADDR` on the scratch copy prints the deny and leaves one new incident with guard `enforce-agent-baseline` in that store; not working looks like no new incident row.
