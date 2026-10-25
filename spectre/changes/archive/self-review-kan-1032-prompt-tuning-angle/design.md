# self-review-kan-1032-prompt-tuning-angle

## Context

This change fixes the seven non-`big` findings `proposal.md` names on the change's own branch,
`spectre/kan-1032-prompt-tuning-angle`, in its own worktree, per step 3 of
`skills/flow-self-review/SKILL.md`. F-A1 to F-A4's evidence is the pass file's sections 1, 5 and 6;
F-A5 came from the coordinator; F-A6 and F-A7 are the planner's own. The brainstorm record is
`.superpowers/sdd/2026-10-25-self-review-kan-1032-prompt-tuning-angle-design.md` in this worktree.
None is `big`: the widest, F-A3, touches 8 files. <!-- measured: the paths in Task 3's **Files:** in tasks.md → 8 -->

What planning measured beyond the pass:

- **F-A1.** `flow record dispatch begin` exits non-zero only on a real refusal: an unreachable
  store journals and exits 0 (`classifyRecordWrite`, `stats/cmd/flow/record.go`), and a refused
  write exits 1, a validation failure (the inline dirty-tree check included) 2. So a non-zero
  `begin` in `runverifylist.go`'s `record` is always a refusal, and stopping on it never turns an
  unreachable store into a failure. No line of the agents `## lint` or `## test` list writes a
  tracked file, so the second group's `begin` sees the tree the first saw.
- **F-A2.** `.superpowers/sdd/reproducers/r0-M1-1.sh`, after the hand widening, still filters
  (`-run TestLoopLanding`); it caught the killing test only because the new name shares the
  prefix.
- **F-A3.** The parent transcript
  (`~/.claude/projects/-Users-tweety53-Projects-agents/4d61a117-3338-4979-b6e4-7a7e49fcf619.jsonl`)
  shows three launches described `pipeline-fix-30-plan sync before implementation`: the first two
  prompts carry no `Key:` line (`LOOP_KEY_LINE` does not match), and both were denied by
  `begin_row_reason` (`hooks/enforce-agent-baseline.py`) — `ROW_KEY` admits a `pipeline-fix-`
  description word, and with no loop-key line the hook demands an open row under the session's
  change. The parent ran `flow record dispatch begin -change kan-1032-prompt-tuning-angle -key
  pipeline-fix-30-plan` at 18:29:00Z, the second launch was denied again, and the third, with
  `Key: pipeline-fix-30-plan` added, went through `enforce-loop-dispatch.py`. The parent then
  hand-began and ended every later loop step under kan-1032 although no check asked. A read-only
  query on the dev store finds exactly 5 agentless `pipeline-fix-` rows under a non-loop change
  with a same-key twin under `pipeline-fix-<that change>-<k>`, all kan-1032's; the other 13
  changes the pass named hold rows from before loops had their own change, with no twin, and are
  left alone. `flow record dispatch begin` exempts `pipeline-fix-` keys from the implementer-key
  check under any change (`stats/cmd/flow/record.go`); `check-loop-dispatch.sh` records every
  `pipeline-fix-` row under a `pipeline-fix-` change (`lpFromKey`, `stats/internal/guard/loop.go`),
  while an own-branch self-review records `self-review-` rows under the run's own change, so only
  the `pipeline-fix-` prefix can be refused. `findings.dispatch_id` is the one foreign key onto
  `dispatches` (`0010_run_records.sql`).
- **F-A4.** The store lists kan-1032's in-run fix as row 1020, `flow-fix`, `fixed`, ref `3936b37d…`.
  The self-review pass's own `fixed` rows are recorded by `land` (`recordFixedRows`,
  `stats/internal/guard/integratechange_land.go`), after `commit`, and each names a sha the
  committed self-review report carries. The repair line for kan-1032 is already appended to
  `spectre/changes/archive/kan-1032-prompt-tuning-angle/narrative.md` (uncommitted, the
  coordinator's), so this plan carries only the guard.
- **F-A5.** `prompttuning.Case.Source.Round` is a plain `int` (`stats/internal/prompttuning/prompttuning.go`),
  so an absent `round` and `round: 0` decode alike; `c.Source.Round < 1`
  (`stats/internal/guard/prompttuningcases.go`) refuses both. `skills/flow/review-panel.md` states
  "the round `0` for the initial panel".
- **F-A7.** The plan commit was denied by `hooks/protect-main-checkout.py` ("no `chore` subject")
  when one Bash call wrote a quoted heredoc holding an apostrophe and then ran `git commit -m
  "chore(spectre): …"`; a scratch run of the hook reproduces it, and allows the same command with
  no apostrophe or with the commit alone. `tokenize` falls back to raw words on unbalanced quotes,
  keeping the quote on the subject word.
- **F-A6.** This planner's first `Explore` dispatch was denied with `DESCRIPTION_REASON` (the
  description opened `self-review plan —`, no row key). `description_reason` and
  `begin_row_reason` check every non-fork dispatch in a flow session; no closed-list row is an
  `Explore` or `Plan` dispatch (`skills/flow/implement.md`), and neither type edits files.

## Decisions

### A refused begin stops the verify list

**ID:** verify-list-stops-on-refused-begin
**Status:** active
**Chosen:** a non-zero `begin` prints `run-verify-list: flow record dispatch begin refused — no
command ran: <its output>` and exits 4 before any command; a failed `end` stays a warning that
never changes the exit code — the refusal is knowable before the list runs and a re-run is the
only way to get the row, so running first wastes the whole list.
**Considered:** keep running and only warn louder — the row is still lost and the re-run still
happens; pre-checking the dirty tree inside `run-verify-list` — duplicates the CLI's own check,
and misses every other refusal cause.

### A mutation reproducer runs the package's whole suite

**ID:** mutation-reproducer-runs-whole-package
**Status:** active
**Chosen:** one sentence in **The mutation-testing brief** (`skills/flow/review-panel.md`) — the
brief is what the mutation slot reads when it writes the reproducer.
**Considered:** `run-reproducer.sh` refusing a test-name filter — a filter's syntax differs per
build tool (`-run`, `--tests`, `-t`), so the guard would be a per-language parser for a rule the
slot can follow from one sentence.

### The hook asks for the Key line, never a run-change row

**ID:** loop-key-needs-its-key-line
**Status:** active
**Chosen:** in a flow session, a dispatch whose description's first word opens `pipeline-fix-` or
`self-review-` and whose prompt has no `LOOP_KEY_LINE` is denied by a new reason naming the
missing `Key: <word>` line and saying `hooks/enforce-loop-dispatch.py` records the row under the
loop's change — checked before `begin_row_reason`, with no store read; `begin_row_reason` never
reaches such a word.
**Considered:** dropping `pipeline-fix-` from `ROW_KEY` — the description check would then deny a
loop launch with the generic "open with its row" text, which still never names the `Key:` line.

### begin refuses a loop key under a run's change

**ID:** begin-refuses-loop-key-off-loop-change
**Status:** active
**Chosen:** `flow record dispatch begin` exits 2 on a `-key` opening `pipeline-fix-` whose change
does not open `pipeline-fix-`, naming `hooks/enforce-loop-dispatch.py` — the one mechanical stop
for the hand-recording habit the transcript shows on every later loop step.
**Considered:** documentation alone — the parent hand-recorded `-fix`, `-review-1`, `-fix-1` and
`-review-2` with nothing asking; refusing `self-review-` keys too — an own-branch self-review
records them under the run's own change by design.

### Migration 0047 deletes the stored twins

**ID:** migration-drops-loop-twins
**Status:** active
**Chosen:** `stats/internal/store/migrations/0047_drop_loop_dispatch_twins.sql` deletes every
agentless dispatch whose key opens `pipeline-fix-`, under a change whose name does not, that has a
same-key row under a change `pipeline-fix-<that change's name>-<k>` of the same project — nulling
any `findings.dispatch_id` onto it first. Idempotent; it runs on flowd's next start.
**Considered:** a `flowd` repair subcommand — an operator step for a one-time cleanup a migration
does unattended; deleting every `pipeline-fix-` row under a non-loop change — 13 older changes'
rows are the only record of their loops.

### commit requires the line for every fixed in-run row

**ID:** commit-requires-in-run-fix-line
**Status:** active
**Chosen:** `integrate-change commit`, after the format check, reads `flow self-review findings
-change <name>` and stops `narrative` naming each `flow-fix` `fixed` row whose `ref` no
`In-run pipeline fix:` line's sha prefixes, skipping a row whose sha the self-review report
(`c.report()`) already carries; a read that fails or does not parse is one warning line and the
check is skipped.
**Considered:** checking against the narrative's prose — no shape to match; failing closed on an
unreadable store — the store being down is no defect of the narrative.

### An absent round is not round 0

**ID:** round-zero-is-a-round
**Status:** active
**Chosen:** `Source.Round` becomes `*int`; the guard refuses `nil` (`source.round is not set`) and
a negative value, and accepts 0.
**Considered:** dropping the check — an absent `round` would then pass silently; `< 0` alone on the
`int` — the same, since an absent key reads 0.

### Explore and Plan pass the row checks

**ID:** planning-agents-pass-row-checks
**Status:** active
**Chosen:** `description_reason`, `begin_row_reason` and the new loop-key reason return `None`
for `subagent_type` `Explore` or `Plan` — the read-only planning agents section B dispatches,
which no ledger row records.
**Considered:** requiring a row key on their description — no closed-list row exists for them, so
the begun-row check would then deny them as well.

### A split message is not judged

**ID:** unread-message-is-not-judged
**Status:** active
**Chosen:** `commit_lacks_trailer` (`hooks/protect-main-checkout.py`) returns False for a message
value opening with a quote — the shape only `tokenize`'s unbalanced-quote fallback leaves — as it
already does for any message the command text does not state.
**Considered:** stripping heredoc bodies before lexing — a parser for a case the docstring's own
"never judged" rule already covers in one line.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — a migration that deletes store rows and two hook deny paths are not
mechanical.
**Considered:** flow-low — fits docs-and-literal fixes, not a data-deleting migration.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** two groups, in series — group 1 Tasks 3 and 6 (the baseline hook, the record CLI, the
migration, and the docs naming the loop's recording; Task 6 edits the hook Task 3 edits); group 2
Tasks 1, 2, 4, 5 and 7 (the verify-list, mutation-brief, integrate-change, prompt-tuning and
commit-hook fixes, each confined to its own files).
**Considered:** one fixer for all seven — one context over about 20 files <!-- predicted: the union of every task's **Files:** in tasks.md -->; one fixer per task — seven
dispatches for tasks of a few lines each.

## Open questions

None.

## Recommended approach

Fix each finding where it starts: `runverifylist.go`'s `record` for F-A1, the brief for F-A2
(`mutation-reproducer-runs-whole-package` rejects a runner guard), the baseline hook plus a
`begin` refusal plus a migration for F-A3 (`loop-key-needs-its-key-line`,
`begin-refuses-loop-key-off-loop-change`, `migration-drops-loop-twins`), `integrate-change
commit` for F-A4 (`commit-requires-in-run-fix-line`), the case type for F-A5
(`round-zero-is-a-round`) and the hook's two row checks for F-A6
(`planning-agents-pass-row-checks`), and `commit_lacks_trailer` for F-A7
(`unread-message-is-not-judged`). Every prose sentence the tasks add is in
`patches/task-<n>.diff` and listed in `verbatim-moves.txt`.

## Critical files

- `stats/internal/guard/runverifylist.go` — the verify-list runner and its `record` (F-A1).
- `stats/internal/guard/run_verify_list_test.go` — its tests (F-A1).
- `scripts/run-verify-list.sh` — the shim's exit-code header (F-A1).
- `skills/flow/verify-and-handoff.md` — **Recording** under **Inline verify** (F-A1).
- `skills/flow/review-panel.md` — **The mutation-testing brief** (F-A2).
- `hooks/enforce-agent-baseline.py` — the row checks (F-A3, F-A6).
- `stats/internal/repotest/enforce_agent_baseline_hook_test.go` — the hook's tests (F-A3, F-A6).
- `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go` — `flow record dispatch begin` (F-A3).
- `stats/internal/store/migrations/0047_drop_loop_dispatch_twins.sql`,
  `stats/internal/store/legacyshapes_test.go` — the repair and its test (F-A3).
- `skills/flow/implement.md`, `skills/flow-contracts/pipeline.md` — the closed list and
  **Pipeline defects found mid-run** (F-A3).
- `stats/internal/guard/integratechange_commit.go`, `stats/internal/guard/integratechange.go`,
  `stats/internal/guard/integrate_change_prepare_test.go`,
  `stats/internal/guard/integrate_change_test.go`, `skills/flow/integrate.md` — `commit` and its
  harness (F-A4).
- `stats/internal/prompttuning/prompttuning.go`, `stats/internal/prompttuning/prompttuning_test.go`,
  `stats/internal/guard/prompttuningcases.go`, `stats/internal/guard/check_prompt_tuning_cases_test.go`
  — the case type and its guard (F-A5).
- `hooks/protect-main-checkout.py`, `stats/internal/repotest/protect_main_checkout_test.go` — the
  change-branch commit check (F-A7).

## Existing code to reuse

- `kwRun` and the `record` closure in `runverifylist.go` — the begin call whose rc F-A1 reads.
- `LOOP_KEY_LINE`, `ROW_DESCRIPTION`, `in_flow_session` and `openRowPayload`/`openRowFixture`
  (`hooks/enforce-agent-baseline.py`, its test) — the loop-key and session tests F-A3 and F-A6
  reuse.
- `TestMigration0031RewritesLegacyRows` and `newTestDatabase` (`stats/internal/store`) — the
  migration-test pattern.
- `icFixLine`/`icFixFormat` (`integratechange_commit.go`), the `self-review findings` read in
  `recordFixedRows` and `c.report()` (`integratechange_land.go`), `icRunCases` and the fake `flow`
  in `integrate_change_test.go`.
- `pmFixture` and `TestProtectMainCheckoutDeniesAnUntraileredChangeBranchCommit`'s setup
  (`stats/internal/repotest/protect_main_checkout_test.go`) for F-A7's test.

## Verification

- Per task, the steps' targeted tests and guards.
- After the last task: `cd stats && make -s web-dist && go vet ./... && gofmt -l .` prints nothing,
  `go test ./...` from `stats/` passes, and every `scripts/check-*.sh` line of `.flow/project.md`'s
  `## lint` exits 0 (`check-verbatim-moves.sh` against this change's `verbatim-moves.txt`).

## Live check

None at `flow.verify` — a self-review loop runs no `flow.verify`. Migration 0047 is exercised on a
scratch database by its test; the dev store's before figure is the 5 twins the read-only query in
**Context** found, and after flowd's next start the same query reads 0 — a non-zero count is the
migration not applied.
