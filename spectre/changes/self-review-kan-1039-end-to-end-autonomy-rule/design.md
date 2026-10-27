## Context

One spectre change for the nine non-`big` findings of the self-review pass of
`kan-1039-end-to-end-autonomy-rule` (evidence:
`/Users/tweety53/Projects/agents-worktrees/kan-1039-end-to-end-autonomy-rule/.superpowers/sdd/self-review-pass.md`).
The fixes ride the change's own branch `spectre/kan-1039-end-to-end-autonomy-rule` in this
worktree, from the HEAD it held when the pass started. Every finding is carried by exactly one task.

Facts settled at plan time:

- **1.2** reproduced: the pass's minimal command is denied by today's hook from the main
  checkout, and a quote-aware scan in `command_segments` admits it while every existing
  `TestProtectMainCheckout*` row still holds (the run is quoted in Task 2).
- **1.3**: the tests that read every shell script of the tree are
  `TestNoPipeIntoEarlyExitGrep` (`setup.sh`, `scripts/*.sh`, `scripts/lib/*.sh`) and
  `TestShimMissingLibCannotAnswer` (`scripts/*.sh`) in `stats/internal/guard/`, and case 5 of
  `TestGitConfigPins` (`scripts/`) in `stats/internal/repotest/`; `pfGoAudits` already names
  `TestGitConfigPins` but only for a touched Go file.
- **1.4**: every pass 1 dispatches `mutation+failure-modes` (**The always-on extra dispatches**,
  `skills/flow/review-panel.md`), so on pass 1 the reading dispatches keep the kan-920 overlap and
  the mutation bundle alone moves after the suite. `flow record pass` is append-only, so a second
  `panel-round-begins.sh` call for the round adds a second pass row, which no reader refuses
  (`stats/internal/records/render.go` reads only whether a round has one).
- **1.6**: `check-panel-findings-closed`'s tests answer every `flow` call they do not name with the
  findings JSON, whose rows carry `round`, so a mutation read added there proves every round in the
  existing tests unchanged.
- **2.1**: dispatch row id 8196 (`full-suite-fix-1`, seq 14) reads started
  `2026-10-27T16:44:08+02:00` and ended `2026-10-27T17:24:18+02:00`; the newest panel reviewer end before it is
  `panel-0-mutation+failure-modes`'s `2026-10-27T17:14:22+02:00` (`flow record dispatches`, read at
  plan time). `flowd retime-dispatch` and `flowd restamp-dispatch` exist in
  `stats/cmd/flowd/main.go`.
- **5.1**: `flow suite list`'s newest row is from 2026-10-09; `flow suite record -suite <name> --
  <command>` runs and times a child and never changes its exit code
  (`stats/cmd/flow/suite.go`).

## Decisions

### Begins recorded before the launch message

**ID:** begins-before-launch
**Status:** active
**Chosen:** every place that prescribes the order — `skills/flow/review-panel.md` (the round's
begins), `skills/flow/implement.md` (the `begin` paragraph and step 3 of **The next implementer
overlaps the guard**), `skills/flow/SKILL-rationale.md` and `scripts/panel-round-begins.sh`'s
header — states one Bash call recording every begin immediately before the one launch message,
and `review-panel.md` names the hook that enforces it.
**Considered:** relaxing `begin_row_reason` in `hooks/enforce-agent-baseline.py` to admit a launch
whose row opens in the next call — the hook's denial is what keeps every launch on a recorded
row, and the harvester already stamps the agent id from the later launch result.

### Single-quoted spans read as literal

**ID:** single-quotes-literal
**Status:** active
**Chosen:** `outside_single_quotes(text)` in `hooks/protect-main-checkout.py` blanks every `'…'`
and `$'…'` span (a `'` inside `"…"` or behind a backslash stays literal), and `command_segments`
tests `SUBST_RE` on that text; when a quote never closes the function answers None and the test
runs on the text as before, so no command reads more permissively than today.
**Considered:** dropping the backtick from `SUBST_RE` — admits a real backtick substitution
unscanned; a full shell tokenizer (`shlex`) — raises on the unbalanced input the hook must still
fail closed on, and changes every segment the hook reads.

### Script scanners in pipeline-fix verify

**ID:** pf-script-scanners
**Status:** active
**Chosen:** `pfChecks` adds, for a touched `setup.sh` or `scripts/**/*.sh`,
`cd stats && go test ./internal/guard/ -run '^Test(NoPipeIntoEarlyExitGrep|ShimMissingLibCannotAnswer)$' -count=1`
when `stats/internal/guard` exists in the worktree, and
`cd stats && go test ./internal/repotest/ -run '^TestGitConfigPins$' -count=1` when
`stats/internal/repotest` does.
**Considered:** the whole guard package for any touched script — minutes of tests for a one-line
script edit, where only these scan scripts; adding a script to `pfGoAudits`' trigger — it would
also run `TestFlowAddrDeclaration`, which reads only Go.

### Mutation bundle launched after the pass-1 suite

**ID:** mutation-after-suite
**Status:** active
**Chosen:** on pass 1 the parent records the reading dispatches' begins and launches them, runs
the full suite in the foreground, and once it has exited records the `mutation+failure-modes`
begin with a second `panel-round-begins.sh` call and launches it, then waits on every report;
stated in `skills/flow/implement.md` (**The pass-1 panel reads while the suite runs.**) and
`skills/flow/review-panel.md` (**On pass 1 the full suite runs between the begins and the
wait.**).
**Considered:** raising the suite's `-timeout` — widens a tolerance over a load the run creates
itself (`rules/fix-determinism-at-the-source.mdc`); running the suite before every pass-1 slot —
gives back the kan-920 overlap the reading slots still have; capping the mutation slot's `go
test -p` — the slot still competes for the same CPUs for its whole run.

### In-run fix line derived from the store row

**ID:** derive-fix-line
**Status:** active
**Chosen:** `requireFixLines` (`stats/internal/guard/integratechange_commit.go`) appends
`- In-run pipeline fix: <ref, first 7> — <note> (blast radius <N> files)` to the narrative
`commit` reads, for each fixed `flow-fix` row no line names whose note is one non-empty line and
whose blast radius is set, printing `narrative: appended <the line>`; a row lacking either still
stops `narrative` as today.
**Considered:** appending at **Write `IN_PROGRESS`** — a planning-session fix can land after the
creating run wrote it, and `commit` is the step that already reads the store row; writing the line
in the planning session — `narrative.md` does not exist at `STARTED`.

### Fix-mutation rows checked at panel close

**ID:** mutation-row-at-panel-close
**Status:** active
**Chosen:** the fixed-round predicate of `check-unfinished-work` moves into
`uwUnprovedFixRounds` (`stats/internal/guard/unfinishedwork.go`), and
`check-panel-findings-closed` calls it on the findings it already read, raising a violation that
names the rounds; a mutation read that fails is exit 2, cannot answer, as in the landing gate.
**Considered:** a check in `flow record finding -status fixed` — the mutation row is recorded
after the finding is marked fixed, so it would refuse the ordinary order.

### Full-suite fix floored at the panel's reviewer ends

**ID:** full-suite-fix-floor
**Status:** active
**Chosen:** `panelFixFloor` (`stats/cmd/flow/record.go`) also floors a `-role panel-fix` row keyed
`full-suite-fix-<n>` at the newest end among panel reviewer rows that ended no later than the fix
commit; this change's row 8196 is repaired with `flowd retime-dispatch` to that floor, then
`flowd restamp-dispatch` over the parent transcript to re-attribute its tokens.
**Considered:** flooring at round 0 only — a full-suite fix after a later round's panel would
keep claiming that panel's wait; leaving row 8196 as recorded — a fix repairs what its defect
already wrote (`skills/flow-self-review/SKILL.md`, step 3).

### Suite runtimes recorded by run-verify-list

**ID:** suite-line-in-test-list
**Status:** active
**Chosen:** a `# suite: <name>` line in a `## test` fence names the one command below it as that
suite; `run-verify-list.sh` runs such a command as `flow suite record -C <worktree> -suite <name>
-- bash -c <command>` when `flow` is on PATH, so the existing store path records it and its exit
code passes through; `.flow/project.md` declares `guard-tests`, `stats-go` and `stats-spa`, and
`skills/flow-contracts/project-configuration.md` documents the line.
**Considered:** mapping command text to suite names inside `run-verify-list` — names this
repository's suites in a guard every project runs; a new `-duration` flag on `flow suite record` —
a second store path for what the wrapper already records.

### Mutation runs bypass the test cache

**ID:** mutation-no-cache
**Status:** active
**Chosen:** one sentence at the end of the reproducer paragraph of **The mutation-testing brief**
(`skills/flow/review-panel.md`): every test run of a mutant passes the runner's no-cache flag,
`-count=1` for `go test`.
**Considered:** a guard on the reproducer's command — `check-mutation-reproducer-pin.sh` reads
the reproducer, not the slot's own sweep runs, where the cache bit.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — a security hook's parsing, integrate's `commit` and the panel-close
gate change behaviour on paths every run takes.
**Considered:** `flow-low` — the prose tasks alone would qualify, the gates do not.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** three groups, in series — `self-review-kan-1039-end-to-end-autonomy-rule-fix-g1`:
Tasks 1, 2, 4, 9 (diffs verified at plan time: the prose and the hook);
`-fix-g2`: Tasks 3, 5, 6, 8 (the guard package, one test-count chain);
`-fix-g3`: Task 7 (`flow record` and the store repair). Each on `flow-low`/`opus`.
**Considered:** one fixer for all nine — the guard package's Go work and the store repair in one
context; one per task — nine dispatches where three file-coherent groups serve.

## Open questions

none

## Recommended approach

Fix each defect where it lives: the begin-order prose (`begins-before-launch`), the hook's
substitution test (`single-quotes-literal`), `pfChecks` (`pf-script-scanners`), the pass-1 launch
order (`mutation-after-suite`), integrate's narrative step (`derive-fix-line`), the panel-close
guard (`mutation-row-at-panel-close`), the back-fill floor and its repair (`full-suite-fix-floor`),
the verify list's suite recording (`suite-line-in-test-list`) and the mutation brief
(`mutation-no-cache`). Every rejected alternative is under its decision.

## Critical files

- `skills/flow/review-panel.md` — the round's begins, the pass-1 suite paragraph, the mutation brief.
- `skills/flow/implement.md` — the `begin` paragraph, step 3 of the overlap, the pass-1 overlap.
- `skills/flow/SKILL-rationale.md` — the rationale copy of the `begin` paragraph.
- `scripts/panel-round-begins.sh` — its header's order.
- `hooks/enforce-agent-baseline.py` — `begin_row_reason`, the denial the prose now states (read, unchanged).
- `hooks/protect-main-checkout.py`, `stats/internal/repotest/protect_main_checkout_test.go` — `command_segments` and its tests.
- `stats/internal/guard/pipelinefix.go`, `stats/internal/guard/pipeline_fix_test.go`, `scripts/pipeline-fix.sh` — `pfChecks` and the verify contract.
- `stats/internal/guard/integratechange_commit.go`, `stats/internal/guard/integratechange.go`, `stats/internal/guard/integrate_change_prepare_test.go`, `skills/flow/integrate.md` — the narrative step, its header and its stop text.
- `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/panelfindingsclosed.go`, `stats/internal/guard/check_panel_findings_closed_test.go`, `scripts/check-panel-findings-closed.sh`, `skills/flow/review-panel-fix-round.md` — the predicate, the panel-close guard and its contract.
- `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go` — `panelFixFloor`.
- `stats/cmd/flowd/main.go` — `retime-dispatch`, `restamp-dispatch` (read, unchanged).
- `stats/internal/guard/runverifylist.go`, `stats/internal/guard/run_verify_list_test.go`, `scripts/run-verify-list.sh`, `.flow/project.md`, `skills/flow-contracts/project-configuration.md` — the verify list and its `# suite:` line.
- `spectre/changes/self-review-kan-1039-end-to-end-autonomy-rule/verbatim-moves.txt` — the acknowledged reworded and new sentences.

## Existing code to reuse

- `SUBST_RE` and `command_segments` (`hooks/protect-main-checkout.py`); the `pmFixture` of `stats/internal/repotest/protect_main_checkout_test.go`.
- `pfGoAudits`, `isDir` and the `add` closure of `pfChecks` (`stats/internal/guard/pipelinefix.go`); `TestPipelineFixChecksGoAudits`' fixture.
- `icFixFormat` and `c.say` (`stats/internal/guard/integratechange_commit.go`); `icRunCases` and the `sr-findings.json` fake of `integrate_change_prepare_test.go`.
- The fixed-round block of `stats/internal/guard/unfinishedwork.go`; `pcFlow` (`flowcli.go`); `cfcBinDispatches` and `cfcRun` (`check_panel_findings_closed_test.go`).
- `records.PanelDispatchRound` and `TestRecordDispatchBackfillFloorsAPanelFixAtItsRoundsReviewers`' fixture (`stats/cmd/flow/record_test.go`).
- `rvlInputs`' parsing loop and `lookPath` (`stats/internal/guard/runverifylist.go`); `flow suite record` (`stats/cmd/flow/suite.go`).
- `flowd retime-dispatch`, `flowd restamp-dispatch` (`stats/cmd/flowd/main.go`).

## Verification

- `cd stats && make -s web-dist && go test ./internal/guard/ ./internal/repotest/ ./cmd/flow/ -count=1` and `scripts/run-guard-tests.sh` green; each new test fails with its fix reverted.
- `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` pass with the sentences acknowledged.
- The 1.2 reproducer, run through the hook with the main checkout as cwd, is allowed.
- `flow record dispatches -change kan-1039-end-to-end-autonomy-rule` shows row 8196 starting at the mutation bundle's end.

## Live check

- **Suite rows (5.1).** Before: `flow suite list | head -3` — newest row dated 2026-10-09. Exercise: verify's `run-verify-list.sh <worktree> test` on this branch. After: `flow suite list` lists a `guard-tests`, a `stats-go` and a `stats-spa` row dated the day verify ran. Not working looks like: the newest row still dated 2026-10-09.
- **Row 8196 (2.1).** Before: started `2026-10-27T16:44:08+02:00`, cache read 4158740 (`flow record dispatches -change kan-1039-end-to-end-autonomy-rule`). After Task 7's repair: started `2026-10-27T17:14:22+02:00`, cache read below the before figure. Not working looks like: the start unchanged.
