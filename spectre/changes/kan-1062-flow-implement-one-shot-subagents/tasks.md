# kan-1062-flow-implement-one-shot-subagents

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** yes — the implementer and gated reviewer dispatch paragraphs move verbatim out of `skills/flow/implement.md` into `skills/flow/implementer-brief.md` and `skills/flow/gated-reviewer-brief.md`
> **Design shape:** plan-mode

**Goal:** Take `flow.sdd-tdd`'s implementation and fix work off the main session (KAN-1062): every class
but `micro` dispatches implementers, every `sdd` fix is a fresh one-shot, a group boundary is one Bash
call, prompts are stubs over briefs on disk, a child's summary is its final message, and the parent
waits on completion notifications.

**Architecture:** `plan-class`'s tree prints `execution sdd` below `micro`; `flow record dispatch
begin` refuses the parent's own task or full-suite row under `sdd`; `flow record finding -from`
numbers a ref-less line; the dispatch hook admits a `full-suite-fix-<n>` key; the contract text lands
as two planned patches — the brief relocation with its `dpSites` rows, then the behaviour.

**Tech Stack:** Go (`stats/internal/guard`, `stats/cmd/flow`, `stats/internal/prompttuning`,
`stats/internal/repotest`), Python 3 standard library (`hooks/`), Bash shim headers, Markdown.

**Spec:** `design.md` beside this file; the contract text in `contract-edits/briefs-on-disk.patch`
and then `contract-edits/one-shot-sdd.patch` beside this file — canonical for every contract
sentence, applied in that order, read-only to every task; `verbatim-moves.txt` beside this file —
the acknowledgement list `check-verbatim-moves.sh` reads, harvested from the two patches and
committed with the plan.

## Global constraints

- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1062-flow-implement-one-shot-subagents`
  (`<wt>`), branch `spectre/kan-1062-flow-implement-one-shot-subagents`; `**Files:**` paths are
  relative to `<wt>`. Never touch the main checkout, never create a worktree, never write
  `spectre/changes/`.
- Go: `gofmt -w` every Go file changed; `cd stats && make -s web-dist && go vet ./... && gofmt -l .`
  exits clean. Guard tests build fixture trees under `t.TempDir()`; `stats/cmd/flow` tests use the
  package's fake daemons (`genuineDaemon`, `beginUnderDecision`, `recordRepo`) — never the live store.
- Never stop, restart or drop `flowd` on 4173, the `flow-postgres` container or the `flow` database.
- Every Go test run sets a Bash `timeout` of 600000.
- Mutation proofs restore with `scripts/break-and-prove.sh` — never `git checkout --` over
  uncommitted edits.
- A changed script's or hook's header is its contract and changes in the same task as its code.
- `hooks/*.py` stay Python 3 standard library only.
- A contract patch that does not apply cleanly is applied with `git apply --3way` and resolved to the
  patch's text; its sentences are never reworded — `check-verbatim-moves.sh` holds them to
  `verbatim-moves.txt` character for character.

## Review Focus

- Task 2: the refusal binds only a live begin with `-agent-id inline`, an `implementer` or
  `panel-fix` role, a `task-` or `full-suite-fix-` key and a newest decision of `execution: sdd`;
  `self-review-` and `pipeline-fix-` keys, a back-fill and a failed decisions read refuse nothing.
- Task 3: a ref-less line's ref is above both the store's highest and every ref the file names; the
  refs are written into the file before anything is recorded, so a re-run rewrites and never adds; a
  failed store read records nothing and leaves the file's bytes unchanged.
- Task 5: every moved paragraph lands byte for byte (the relocation comparison), and the dispatch
  paragraph guard's parity cases still compare the port against the historical bash on equal input.
- Task 6: the one-call boundary block, the fixer table and the gated stage-end path agree with
  `close-task.sh`'s exit codes, `check-task-reviewer-single-dispatch`'s key shapes and the one-mutator
  rule.

- [x] 1. Print `execution sdd` for every plan class but `micro`
  - [x] **Step 1: Write the test.** `stats/internal/guard/plan_class_test.go`:
    `TestPlanClassTreeSddBelowMicro` — `pcTree` for `small`, `regular` and `big`, at a compact,
    bundle and experimental roll on each side of its threshold, opens
    `tree: class <class> · execution sdd · implementer chosen`, and for `micro`
    `tree: class micro · execution inline · implementer skipped — inline`. Bring
    `TestPlanClassTree`'s `small` and `regular` rows and its two `regular` whole-guard expectations
    (the `pinTree` helper and the `tree-192` case) to `execution sdd · implementer chosen`. Run
    `cd stats && go test ./internal/guard -run 'TestPlanClassTree' -count=1` — fails.
    <!-- measured: scripts/plan-class.sh spectre/changes/kan-1062-flow-implement-one-shot-subagents/tasks.md 1 prints "class: regular" and "tree: class regular · execution inline · implementer skipped — inline" @ branch spectre/kan-1062-flow-implement-one-shot-subagents, before Task 1 -->
  - [x] **Step 2: Implement.** `stats/internal/guard/planclass.go`, `pcTree`: every class `micro`'s
    early return leaves gets `execution sdd · implementer chosen`; its doc comment says `micro` alone
    is `inline`. `scripts/plan-class.sh`'s header: one sentence after the tree line's description —
    the execution is `sdd`, implementer `chosen`, on every class but `micro`, which is `inline`
    (decision `sdd-below-micro`).
  - [x] **Step 3: Prove.** `scripts/break-and-prove.sh` restoring the `big`-only condition in
    `pcTree` — `TestPlanClassTreeSddBelowMicro` fails; restored, it passes.
  - [x] **Step 4: Verify.** `cd stats && go test ./internal/guard -run 'TestPlanClass' -count=1 | tail -20`
    passes; `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && make -s web-dist && go vet ./internal/guard`
    exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/planclass.go`, `stats/internal/guard/plan_class_test.go`, `scripts/plan-class.sh`
**Tests:** `TestPlanClassTreeSddBelowMicro`
**Regression:** `TestPlanClassTreeSddBelowMicro` fails — `small` and `regular` plans are decided `inline` and the parent implements them on its own context.
**Commit:** feat(guard): print execution sdd for every plan class but micro
**After:** none

**Decision:** sdd-below-micro

- [x] 2. Refuse the parent's own task or full-suite row under an `sdd` decision
  - [x] **Step 1: Write the test.** `stats/cmd/flow/record_test.go`:
    `TestRecordDispatchBeginRefusesInlineUnderSdd`, a table through `beginUnderDecision` with
    ``decisionsBody(`{"execution":"sdd"}`)``, beside `TestRecordDispatchBeginRefusesSubagentUnderInline`:
    `-agent-id inline -role implementer -key task-1-implementer` exits 2; `-agent-id inline -role
    implementer -key full-suite-fix-1` exits 2; `-agent-id inline -role panel-fix -key
    full-suite-fix-1` exits 2; `-agent-id inline -role implementer -key self-review-x-fix` exits 0;
    `-agent-id inline -role panel-fix -key panel-fix-1` exits 0; `-agent-id none -role implementer
    -key task-1-implementer` exits 0; a back-fill's begin, `-agent-id inline -role implementer -key
    task-1-implementer -started-at 2026-01-02T03:04:05Z`, exits 0; under a decisions read that fails,
    `-agent-id inline -role implementer -key task-1-implementer` exits 0. Each refusal's stderr names `execution: sdd`. Every
    row carries `-model opus -effort low`. Run
    `cd stats && go test ./cmd/flow -run 'TestRecordDispatchBeginRefuses' -count=1` — fails.
    <!-- measured: a scratch test calling beginUnderDecision with decisionsBody(`{"execution":"sdd"}`) — inline implementer task-1-implementer with and without -started-at, and inline panel-fix full-suite-fix-1 — exits 0 for all three with empty stderr @ 5302fcb2 in a scratchpad clone, go test ./cmd/flow -run TestZZSpike, 2026-10-31 -->
  - [x] **Step 2: Implement.** `stats/cmd/flow/record.go`, the `latestDecision` read under
    `if *startedAt == ""` that refuses a subagent row under an `inline` decision: for the same roles
    and the same `self-review-`/`pipeline-fix-` exemption, an `-agent-id inline` row whose key opens
    `task-` or `full-suite-fix-` under `execution: sdd` prints
    `flow: -role <role> -key <key> -agent-id inline records the parent's own work, but the change's recorded decision is execution: sdd — under sdd every task and every fix is a fresh one-shot dispatch (skills/flow/implement.md, Every fix is a fresh one-shot)`
    to stderr and exits 2. The comment above the block states both directions.
  - [x] **Step 3: Prove.** `scripts/break-and-prove.sh` dropping the new `execution == "sdd"` arm —
    `TestRecordDispatchBeginRefusesInlineUnderSdd` fails; restored, it passes.
  - [x] **Step 4: Verify.** `cd stats && go test ./cmd/flow -run 'TestRecordDispatchBegin' -count=1 | tail -20`
    passes; `cd stats && gofmt -l cmd/flow` prints nothing; `cd stats && make -s web-dist && go vet ./cmd/flow`
    exits 0. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchBeginRefusesInlineUnderSdd`
**Regression:** `TestRecordDispatchBeginRefusesInlineUnderSdd` fails — an `sdd` parent's own fix records as an inline row and the board counts it as dispatched work.
**Commit:** feat(stats): refuse an inline task or full-suite row under an sdd decision
**After:** none

**Decision:** store-refuses-inline-under-sdd

Correction (2026-10-31): the plan declared only the sdd arm in `record.go`. At the base the `-role implementer` key-shape check (`implementerKey`) refuses `full-suite-fix-<n>`, `task-<n>-implementer-recommit-<k>` and `task-<ids>-implementer-pick-<k>` before any decision is read, so the refusal could never name `execution: sdd` and every fixer key Task 6's contract dispatches would be refused. Shipped instead: a `fixerKey` regexp that widens only the shape check; `implementerTask`/`taskWorkKey` still read `implementerKey` alone.

- [x] 3. Number a findings file's ref-less lines from the change's next free ref
  - [x] **Step 1: Write the test.** `stats/cmd/flow/record_test.go`:
    `TestRecordFindingFromAssignsRefs`, on `TestRecordFindingFromFile`'s fixture with a fake daemon
    whose run record for the change holds findings `F1`–`F3`. (a) A file of two lines carrying
    `task`, `severity`, `location` and `note` and no `ref`, loaded with
    `-slot gated-reviewer -round 0 -status open -reproducer none -dispatch-seq 4 -from <file>`,
    posts `F4` then `F5`, each with slot `gated-reviewer`, round 0, dispatch-seq 4 and its own line's
    fields, and the file now carries `"ref":"F4"` and `"ref":"F5"` on those lines. (b) Loading the
    rewritten file again posts `F4` and `F5` again and no `F6`. (c) A file whose first line names
    `"ref":"F9"` and whose second names none posts the second as `F10`. (d) With the run-record read
    answering 500, a ref-less file exits 1, posts nothing, and its bytes are unchanged.
    `TestRecordFindingFromFile` still passes: a file whose every line names its ref reads no run
    record. Run `cd stats && go test ./cmd/flow -run 'TestRecordFindingFrom' -count=1` — fails.
  - [x] **Step 2: Implement.** `stats/cmd/flow/record.go`, `runRecordFindingFrom`: when any line
    carries no `ref`, read the change's run record once, as `runRecordNextRef` does through
    `readNextRecord`, with the command line's `-addr`, `-timeout`, `-C` and `-change`; number every
    ref-less line in file order from one above the higher of the store's highest `F<n>` and every
    `F<n>` the file names; rewrite those lines of the file with their refs before the first is
    recorded, leaving every other line's bytes as they were; then record as today. A failed read
    exits 1 naming the file, before anything is posted or written. Its comment and the `-from` usage
    line state the numbering (decision `from-assigns-refs`).
  - [x] **Step 3: Prove.** `scripts/break-and-prove.sh` dropping the write-back — case (b) of
    `TestRecordFindingFromAssignsRefs` fails on an `F6`; restored, it passes.
  - [x] **Step 4: Verify.** `cd stats && go test ./cmd/flow -run 'TestRecordFinding|TestRunRecordNextRef' -count=1 | tail -20`
    passes; `cd stats && gofmt -l cmd/flow` prints nothing; `cd stats && make -s web-dist && go vet ./cmd/flow`
    exits 0. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordFindingFromAssignsRefs`
**Regression:** `TestRecordFindingFromAssignsRefs` fails — a gated reviewer's findings file is refused for its missing refs, or a re-run records every finding a second time under new refs.
**Commit:** feat(stats): number a findings file's ref-less lines from the change's next free ref
**After:** Task 2

**Decision:** from-assigns-refs

- [x] 4. Admit a `full-suite-fix-<n>` key as a dispatch description's row
  - [x] **Step 1: Write the test.** `stats/internal/repotest/enforce_agent_baseline_hook_test.go`:
    `TestEnforceAgentBaselineHookAdmitsFullSuiteFixKey`, on `TestEnforceAgentBaselineHookAdmitsRestKeys`'s
    shape — with an open row keyed `full-suite-fix-1`, a flow dispatch described
    `full-suite-fix-1 fix the failing suite` passes with no output; with an open row keyed
    `full-suite-fix-1`, one described `full-suite-fixes fix the failing suite` is denied, its reason
    naming `opens with the row it records`. Run
    `cd stats && go test ./internal/repotest -run 'TestEnforceAgentBaselineHookAdmitsFullSuiteFixKey' -count=1` — fails.
    <!-- measured: python3 -I loading hooks/enforce-agent-baseline.py: ROW_DESCRIPTION.match("full-suite-fix-1 fix the failing suite") is None, while "task-3+4-implementer-fix-1 x", "task-2-implementer-recommit-1 x" and "task-5+6-implementer-pick-1 x" match @ branch spectre/kan-1062-flow-implement-one-shot-subagents, before Task 4 -->
  - [x] **Step 2: Implement.** `hooks/enforce-agent-baseline.py`: `ROW_KEY` gains the alternative
    `full-suite-fix-\d+`, and the comment above it names `full-suite-fix-2` among its examples.
  - [x] **Step 3: Prove.** `scripts/break-and-prove.sh` dropping the alternative — the admit case of
    `TestEnforceAgentBaselineHookAdmitsFullSuiteFixKey` fails; restored, it passes.
  - [x] **Step 4: Verify.** `cd stats && go test ./internal/repotest -run 'TestEnforceAgentBaselineHook' -count=1 | tail -20`
    passes; `scripts/check-python-suppressions.sh` exits 0. Commit.

**Build:** green
**Files:** `hooks/enforce-agent-baseline.py`, `stats/internal/repotest/enforce_agent_baseline_hook_test.go`
**Tests:** `TestEnforceAgentBaselineHookAdmitsFullSuiteFixKey`
**Regression:** `TestEnforceAgentBaselineHookAdmitsFullSuiteFixKey` fails — the full-suite fixer's launch is denied in every flow run.
**Commit:** feat(hooks): admit a full-suite-fix key as a dispatch description's row
**After:** none

**Decision:** fixes-to-fresh-one-shots

- [x] 5. Move the implementer and gated reviewer briefs out of `implement.md` behind stub prompts
  - [x] **Step 1: Save the inventory.** `scripts/check-normative-inventory.sh` into a file in the
    session scratchpad.
  - [x] **Step 2: Write the tests.** `stats/internal/guard/check_dispatch_paragraphs_test.go`:
    `TestCheckDispatchParagraphsBriefSites` — a fixture tree whose `skills/flow/implement.md`
    carries one FOREGROUND BUILDS block, one REPRODUCE, DON'T READ block of the reviewer variant and
    two MODEL HANDSHAKE blocks, whose `skills/flow/implementer-brief.md` carries every implementer
    paragraph the new rows name and whose `skills/flow/gated-reviewer-brief.md` carries the
    reviewer's, beside the other sites' clean files, exits 0 printing `43 site(s) validated`; the
    same tree with the TARGETED TESTS block moved from `implementer-brief.md` into `implement.md`
    exits 1 naming
    `skills/flow/implementer-brief.md:0: requires at least 1 block(s) carrying the label "**TARGETED TESTS:**", found 0`;
    with the READ-ONLY REVIEW block removed from `gated-reviewer-brief.md` it exits 1 naming that
    file. Bring the existing fixtures to the new table: `dpCases`' clean tree and every case
    expecting a violation in `implement.md` for a paragraph that moved, and the clean run's
    `37 site(s) validated` to `43`. The parity case "port: an unquoted label line is a one-line
    block" keeps its `implement.md` as the historical bash at d71a2327 reads it and adds the two
    brief files the port now requires, so port and bash still print the same lines.
    `stats/internal/prompttuning/prompttuning_test.go`, `TestAreaFor`:
    `skills/flow/implementer-brief.md` maps to `{"impl", ""}` and `MappedTargets()` holds 10 paths.
    Run `cd stats && go test ./internal/guard -run 'TestCheckDispatchParagraphs' -count=1` and
    `cd stats && go test ./internal/prompttuning -run 'TestAreaFor' -count=1` — both fail.
  - [x] **Step 3: Apply.** `git -C <wt> apply spectre/changes/kan-1062-flow-implement-one-shot-subagents/contract-edits/briefs-on-disk.patch`.
  - [x] **Step 4: Implement.** `stats/internal/guard/dispatchparagraphs.go`, `dpSites`: each row
    below takes the place of the old `skills/flow/implement.md` row of its entry, in table order;
    the `handshake` row for `implement.md` stays at 2 and every other row is unchanged.

```go verified:check-dispatch-paragraphs.sh printed "DISPATCH-PARAGRAPHS-OK … 43 site(s) validated" over 5302fcb2 with both contract patches applied and these rows, in a scratchpad clone 2026-10-31
{"reproduce", "skills/flow/implement.md", 1, []string{"reviewer"}},
{"reproduce", "skills/flow/implementer-brief.md", 1, []string{"implementer"}},
{"reproduce", "skills/flow/gated-reviewer-brief.md", 1, []string{"reviewer"}},
{"foreground", "skills/flow/implement.md", 1, nil},
{"foreground", "skills/flow/implementer-brief.md", 1, nil},
{"foreground", "skills/flow/gated-reviewer-brief.md", 1, nil},
{"targeted", "skills/flow/implementer-brief.md", 1, nil},
{"tools", "skills/flow/implementer-brief.md", 1, nil},
{"tools", "skills/flow/gated-reviewer-brief.md", 1, nil},
{"delegation", "skills/flow/implementer-brief.md", 1, nil},
{"delegation", "skills/flow/gated-reviewer-brief.md", 1, nil},
{"prove", "skills/flow/implementer-brief.md", 1, nil},
{"decide", "skills/flow/implementer-brief.md", 1, nil},
{"budget", "skills/flow/implementer-brief.md", 1, nil},
{"readonly", "skills/flow/gated-reviewer-brief.md", 1, nil},
{"proofruns", "skills/flow/implementer-brief.md", 1, nil},
{"subskill", "skills/flow/implementer-brief.md", 2, []string{"tdd", "debugging"}},
```

  - [x] **Step 5: Implement the rest.** `scripts/check-dispatch-paragraphs.sh`'s header: its site table to these rows, and every
    per-paragraph note naming `implement.md`'s block count or its implementer and gated reviewer
    dispatches brought to the file that now carries them. `stats/internal/prompttuning/prompttuning.go`:
    `promptAreas` maps `skills/flow/implementer-brief.md` to `{"impl", ""}` (decision
    `briefs-on-disk`).
  - [x] **Step 6: Verify.** `cd stats && go test ./internal/guard -run 'TestCheckDispatchParagraphs' -count=1 | tail -20`
    and `cd stats && go test ./internal/prompttuning ./cmd/flow -run 'TestAreaFor|TestPromptTuning' -count=1 | tail -20`
    pass; `scripts/check-dispatch-paragraphs.sh` prints `43 site(s) validated`;
    `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.sh`, `scripts/check-guard-symlinks.sh`,
    `scripts/check-installed-citations.sh` and `scripts/check-ask-silence.sh` exit 0;
    `scripts/check-normative-inventory.sh` prints the inventory saved in Step 1;
    `scripts/run-verify-list.sh <wt> lint` exits 0. Commit.

**Build:** green
**Files:** `skills/flow/implement.md`, `skills/flow/implementer-brief.md`, `skills/flow/gated-reviewer-brief.md`, `skills/flow/verify-fix-loop.md`, `skills/flow/review-panel.md`, `skills/flow/SKILL-rationale.md`, `skills/flow-fast/SKILL.md`, `skills/flow-self-review/SKILL.md`, `skills/flow-contracts/plan-amendment.md`, `scripts/stage-task-files.sh`, `stats/internal/guard/stagetaskfiles.go`, `stats/internal/guard/dispatchparagraphs.go`, `stats/internal/guard/check_dispatch_paragraphs_test.go`, `scripts/check-dispatch-paragraphs.sh`, `stats/internal/prompttuning/prompttuning.go`, `stats/internal/prompttuning/prompttuning_test.go`
**Tests:** `TestCheckDispatchParagraphsBriefSites`
**Regression:** `TestCheckDispatchParagraphsBriefSites` fails — a brief that lost a dispatch paragraph passes the guard, and an implementer works without it.
**Commit:** feat(flow): move the implementer and gated reviewer briefs out of implement.md behind stub prompts
**After:** none

**Decision:** briefs-on-disk
**Decision:** stub-carries-handshake
**Decision:** contract-as-patches

- [x] 6. Run every `sdd` fix as a fresh one-shot and wait on completion notifications
  - [x] **Step 1: Apply.** Save `scripts/check-normative-inventory.sh`'s output to a file in the
    session scratchpad; then
    `git -C <wt> apply spectre/changes/kan-1062-flow-implement-one-shot-subagents/contract-edits/one-shot-sdd.patch`.
  - [x] **Step 2: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.sh`, `scripts/check-guard-symlinks.sh`,
    `scripts/check-dispatch-paragraphs.sh`, `scripts/check-installed-citations.sh`,
    `scripts/check-ask-silence.sh` and `scripts/check-stage-mark-calls.sh` exit 0;
    `scripts/check-normative-inventory.sh` prints the inventory saved in Step 1;
    `scripts/run-verify-list.sh <wt> lint` exits 0. Commit.

**Build:** green
**Files:** `skills/flow/implement.md`, `skills/flow/implementer-brief.md`, `skills/flow/gated-reviewer-brief.md`, `skills/flow/gated-review-fix.md`, `skills/flow/sdd-dispatch.md`, `skills/flow/brainstorm-planner.md`, `skills/flow/SKILL-rationale.md`
**Tests:** **none** — contract prose; Step 2's guards are the check.
**Regression:** none — prose.
**Commit:** feat(flow): run every sdd fix as a fresh one-shot and wait on completion notifications
**After:** Task 1, 2, 3, 4, 5

**Decision:** sdd-below-micro
**Decision:** one-call-boundary
**Decision:** summary-head-no-merger
**Decision:** notification-wait
**Decision:** fixes-to-fresh-one-shots
**Decision:** main-keeps-judgement
**Decision:** gated-fix-at-stage-end
**Decision:** fix-dispatch-in-implementer-brief
**Decision:** gated-context-is-group-bundle
**Decision:** markers-written-once
**Decision:** sdd-tdd-scope
**Decision:** flow-fast-follows-tree
