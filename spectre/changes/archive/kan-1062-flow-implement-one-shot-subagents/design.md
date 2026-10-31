## Context

- Ticket KAN-1062. The design was settled in one brainstorming round (planner `plan-1`,
  `.superpowers/sdd/planner/round-1.md` in this worktree) and approved on 2026-10-31: the operator
  took the recommended option of every question (`answers-1.md` beside it). The cost evidence is
  `proposal.md`'s `## Why`; this section carries what the design rests on beyond it.
- **The reachability check reproduces on the base** (`5302fcb2`): a finished child's leftover work is
  "the parent's own inline fix" (`skills/flow/implement.md` lines 116–120), the gated fix is applied
  by the parent (`skills/flow/gated-review-fix.md` line 6), a clean pass's Minors and every
  full-suite failure are fixed by the parent (`implement.md` 965–971, 751), a pick conflict is "the
  parent's own to fix" (`skills/flow/sdd-dispatch.md` line 35), the wait is a `sleep 5` poll
  (`implement.md` 1094–1096), and `execution` is `inline` for every class but `big`
  (`stats/internal/guard/planclass.go` 114–116). Nothing to withdraw.
- **The ticket's 77% main share is not reproduced in the store.** The baseline this change is judged
  against is the board's own `flow.sdd-tdd` main share, 45.8%, and the stage's mean cost, $12.44 per
  run, over 2026-10-01..31 (`proposal.md`).
- **Where a `big` parent's turns go**, from three main transcripts in their `flow.sdd-tdd` windows
  (kan-1060, kan-1061, gymie kan-1023): 104, 69 and 117 main turns; main context 144K, 157K and 155K
  on entry, growing to 315K, 248K and 341K; main cache reads 24.8M, 14.0M and 28.9M. By what each
  turn's tool call did, `flow record`/`stage`/`tasks` calls took 39%, 26% and 21% of those cache
  reads; the parent's own fix and investigation work 20–35%; close-task and check guards 8%, 19% and
  11%; Agent launches 9%, 7% and 10%, the launches 22, 9 and 16 prompts of 2.4–9.3K characters each;
  poll waits 0%, 0% and 10%.
  <!-- measured: main transcripts e6caac14…, f4e6fadf… and e905b265… parsed by planner plan-1's scratch script, deduplicated by message.id @ machine-local transcripts read 2026-10-31, not re-runnable at a ref -->
- **The parent's prompt cache is written at the one-hour rate.** A main session's usage records
  `ephemeral_5m_input_tokens` 0 on every turn and non-zero `ephemeral_1h_input_tokens`, so a wake
  inside the hour re-reads the parent's context at the cache rate and a wake after it rewrites it.
  <!-- measured: grep -o '"ephemeral_(1h|5m)_input_tokens":[0-9]*' over main transcript 6c9341a8….jsonl — 92 turns at 5m 0, 1h non-zero @ machine-local transcript read 2026-10-31, not re-runnable at a ref -->
- **The contract text is planned, not written by the implementers.** `contract-edits/briefs-on-disk.patch`
  (the relocation, decision `briefs-on-disk`) and `contract-edits/one-shot-sdd.patch` (the
  behaviour), applied in that order to `5302fcb2`, are canonical for every contract sentence this
  change adds, moves or rewords, and `verbatim-moves.txt` is the acknowledgement list
  `check-verbatim-moves.sh` reads, harvested from them (decision `contract-as-patches`). Applied to
  the base in a scratch clone, both pass `check-references.sh`, `check-markdown-integrity.sh`,
  `check-guard-symlinks.sh`, `check-ask-silence.sh`, `check-stage-mark-calls.sh`,
  `check-installed-citations.sh` and, with the acknowledgement list, `check-verbatim-moves.sh`;
  `check-normative-inventory.sh` prints the base's inventory unchanged; and, with the `dpSites`
  rows Task 5 states, `check-dispatch-paragraphs.sh` validates 43 sites.
  <!-- measured: git apply of both patches onto 5302fcb2 in a scratchpad clone, then each guard named above, and check-dispatch-paragraphs.sh printing "43 site(s) validated" @ 5302fcb2 plus the two patches, 2026-10-31 -->
- **`implement.md` shrinks less than the round's estimate.** The relocation takes it from 88,669 to
  75,421 bytes; the behaviour patch adds the stub, the fixer table and the boundary block back, for
  81,283. The two briefs are 15,612 and 3,910 bytes, read by the children, never by the parent.
  <!-- measured: wc -c over skills/flow/implement.md, implementer-brief.md and gated-reviewer-brief.md @ 5302fcb2, then with briefs-on-disk.patch, then with both patches, in a scratchpad clone 2026-10-31 -->
- **Constraints the design keeps.** At most one mutator per worktree (**4. Execute (SDD + TDD)**,
  `skills/flow/implement.md`). Only the parent dispatches and waits; no child dispatches, waits on or
  resumes another, and a `flow-<effort>` agent has no `Agent` tool. Implementers and fixers run on
  `opus` by default and never `high` (**Model and effort**, `skills/flow/brainstorm-planner.md`); the
  gated reviewer rows stay on their unchanged **Allowed pairs** rows. No new role, no new Allowed-pairs
  row and no new decision field: every fixer is role `implementer`, and `flow record dispatch begin`
  already refuses a row without its model and effort.
- **Expected effect (modelled, not measured).** On `big`, main cache reads fall by about half: the
  one-call boundary, fix work leaving main, the stub prompts and notification waits, compounding
  through slower context growth. On `regular` and `small`, main share falls from 80–93% to about
  25–30%, total cost per run flat to −30%. The board's overall `flow.sdd-tdd` main share falls from
  about 46% to 20–25%, total stage cost flat to −10%.
  <!-- predicted: the stage leaderboard over the window after about five /flow runs land on this change (## Live check) -->

## Decisions

### Every class but micro runs sdd

**ID:** sdd-below-micro
**Status:** active
**Chosen:** `plan-class.sh`'s tree prints `execution sdd · implementer chosen` for `small`, `regular`
and `big`; `micro` alone stays `inline`. Decide's steps 2 and 5 — the implementer and fixer pairs,
the groups — then run for those classes too, unchanged. The gated per-task reviewer stays `big`-only,
so `small` and `regular` keep the panel's GATE FOCUS. `/flow-fast` reads the same tree and follows
(decision `flow-fast-follows-tree`). An inline parent does every TDD turn at 150–340K context, where a
fresh implementer does them at roughly 30–80K for a fixed price of its bundle's cache write and about
three parent turns per group boundary.
**Considered:** `regular` and `big` only — `small` would keep 93% of its stage on main for a saving of
$70 in 30 days, but the same mechanism costs nothing extra to apply to it. `big` only, as today —
80% of `regular`'s stage cost stays on main and the board's overall share barely moves.
**Supersedes:** KAN-472's `inline-parent-implements-panel-dispatches`, for its implementer half, on
every class but `micro`.

### A group boundary is one Bash call and one launch message

**ID:** one-call-boundary
**Status:** active
**Chosen:** the wake carries the child's summary head; the parent transcribes any legitimate
correction, then one Bash call runs the markers verify, `close-task.sh` and — on its exit 0 or 3 —
the next group's gather, snapshot and begin and, on `big` when a gate fired, the reviewer bundle's
diffs, planning commit, snapshots and begin; the next message launches. `implement.md` carries the
call as one literal fenced block in place of the three numbered boundary steps. A returning gated
bundle's checks, findings, record end and ticks are one Bash call too, riding the boundary call when
both children return in one wake.
**Considered:** a new Go `sdd-boundary` command that runs the call and prints the launch lines — the
existing scripts already do every link, so chaining them in one documented call buys the same turn
saving with no new program; it stays the follow-up if the board shows boundaries still split.

### Briefs live on disk and every prompt is a stub

**ID:** briefs-on-disk
**Status:** active
**Chosen:** the implementer dispatch paragraphs (FLOW — COMMIT-PER-TASK through REPORT, DON'T DECIDE,
the `Build: red` bundle note included) move verbatim from `implement.md` into
`skills/flow/implementer-brief.md`, and the gated reviewer bundle's shared paragraphs and per-pass
REPORT FILE into `skills/flow/gated-reviewer-brief.md`. Each prompt is a literal stub `implement.md`
states — its Role line, the two baseline sentences, the report-tersely line, `Key:`, MODEL HANDSHAKE,
the line naming its brief, the per-dispatch values, an incident carried verbatim and its REPORT FILE
path, so a child that skips its brief still owes its file (kan-918). The inline (`micro`) parent loads
`implementer-brief.md` and is bound by it in the same words. `dpSites` moves each moved paragraph's
rows to the brief that now carries it. Static files need no renderer: nothing in them is filled per
dispatch. The implementer brief joins the `impl` area of the prompt-tuning map, as
`sdd-dispatch.md` is.
**Considered:** keeping the paragraphs in `implement.md` and the prompts — every launch is kept in the
parent's context for the rest of the run, 100.6K characters of prompts on kan-1060; a
`render-slot-prompt.sh`-style renderer — the panel's renderer exists because its slot briefs fill
placeholders, which these do not.

### A child's summary is its own final message — no merger

**ID:** summary-head-no-merger
**Status:** active
**Chosen:** an implementer or fixer opens its report with a fixed summary head — `Model:`, `status:`,
`commit:`, `correction:`, `undeclared:`, `fixed:`/`not-fixed:`, `fix-mutation:`, `question:`,
`report:` — and returns the head alone as its final message, which the completion notification
carries; its `Model:` line is what **The handshake** compares. A gated reviewer bundle returns one
line per pass and writes the bundle's findings file itself. The parent opens a report body only where
a judgement needs it.
**Considered:** a sonnet/low merger per group boundary folding the implementer report, the close-task
output and the reviewer verdicts into one file — the ticket's literal pattern, but each dispatch yields
one report, so a merger folds nothing and adds a dispatch and about 0.75M main cache reads per
boundary, the cost KAN-1060's design rejected a summariser for.

### The wait is the completion notification

**ID:** notification-wait
**Status:** active
**Chosen:** every child the parent launches goes out with `run_in_background: true` and the parent
ends its turn; `hooks/flow-stop-guard.py` already lets a turn end while a launched agent has not
notified. While a child is out and no shorter timer is armed, one background `sleep 3000` wakes the
parent before its one-hour prompt cache lapses, and a wake with a child still out re-arms it. "Never
end a turn with a child in flight" becomes never report a stage boundary or ask the operator with one.
A child that dies notifies too and takes its no-report path.
**Considered:** keeping the bounded `sleep 5` / `test -s` loop (KAN-857) — every poll turn re-reads the
parent's context at the cache rate for nothing, which the panel stopped doing in 983be8de.
**Supersedes:** the bounded poll loop of KAN-857's **Turn discipline**.

### Under sdd every fix is a fresh one-shot

**ID:** fixes-to-fresh-one-shots
**Status:** active
**Chosen:** under `sdd` the parent edits no code. A `close-task.sh` exit 1 a re-commit repairs
(`task-<n>-implementer-recommit-<k>`, the `fixer` pair), a wave pick conflict
(`task-<ids>-implementer-pick-<k>`, the group's pair, the conflicted pick aborted first), the gated
fix (`task-<ids>-implementer-fix-<k>`, the `fixer` pair, decision `gated-fix-at-stage-end`) and every
full-suite failure (`full-suite-fix-<n>`, the `fixer` pair) go to a fresh one-shot, role
`implementer`, recorded with its pair at `begin`. The full-suite fix is recorded as `implementer`, not
`panel-fix`: a dispatched `panel-fix` row under that key would count as a panel fix round in
`check-panel-fix-single-dispatch`. Every key but `full-suite-fix-<n>` already matches the dispatch
hook's `ROW_KEY`; that shape is added. The parent still runs the full suite itself, one Bash call
through `tail`, and records the green sha verify relies on. Under `inline` (`micro`) the parent keeps
fixing itself, as today.
**Considered:** the parent fixing inline, as today — the measured inline parent pays 80–93% of its
stage at 150–340K context; a new `fixer` role with its own Allowed-pairs row and decision field — the
`implementer` role and the decision's existing `fixer` pair already carry every case.
**Supersedes:** KAN-862's `parent-fixes-minors-inline`, for the gated reviewer's Minors under `sdd`.

### The store refuses the parent's own fix row under sdd

**ID:** store-refuses-inline-under-sdd
**Status:** active
**Chosen:** `flow record dispatch begin` gains the mirror of its refusal of a subagent row under an
`inline` decision: under a newest decision of `execution: sdd`, an `implementer` or `panel-fix` row
begun `-agent-id inline` on a `task-*` or `full-suite-fix-*` key is refused with exit 2, naming the
decision. `self-review-` and `pipeline-fix-` keys stay exempt, and, like the refusal it mirrors, it
binds a live begin only — a back-fill (`-started-at`) records work already done. It makes
`fixes-to-fresh-one-shots` mechanical on the record, the board's only source for main's share.
**Considered:** a hook blocking the parent's edits outright — a hook cannot reliably tell the main
session's tool calls from a child's; a prompt-size cap in the dispatch hook — it limits a symptom the
stub prompts already remove.

### The parent keeps every judgement

**ID:** main-keeps-judgement
**Status:** active
**Chosen:** the parent alone decides whether a correction or an undeclared path is legitimate,
whether a finding stands or is withdrawn with its clause, a BLOCKED report or a question per
**Auto-resolution**, a pivot and its re-decision, and whether a fixer or a re-review goes out; it
makes every `flow record finding`, `status` and `incident` call. An undeclared path is transcribed
with `flow tasks files`, a plan edit, never a fix.
**Considered:** letting a fixer close its own findings — a fixer judging its own fix is the self-review
the panel exists to avoid.

### The gated reviewers' fixes run once, at the stage's end

**ID:** gated-fix-at-stage-end
**Status:** active
**Chosen:** once the last group's guard has passed and every gated bundle has returned, one fixer
carries every `fix` verdict and every Minor a gated pass left open, then one re-review bundle carries
every task whose pass was `fix`, keyed `task-<ids>-reviewer-fix-<k>` — a key
`check-task-reviewer-single-dispatch` already exempts from its group-split rule. No implementer is in
flight then, so the one-mutator rule holds, and every task commit is pushed, so each fix is a commit on
top per task carrying that task's `Task-Id:`. A Minor-only task ticks on its clean verdict and its
Minors stay open findings until the fixer closes them; a `fix` task stays unticked until its re-review
closes clean. Later groups build on the flagged code meanwhile, as they already do: the next group is
in flight when a verdict arrives.
**Considered:** fixing at the next boundary in place of the next group's launch — the flaw is fixed
sooner, at one fixer dispatch per `fix` verdict and the pipeline serialised behind each fix.

### The stub carries MODEL HANDSHAKE

**ID:** stub-carries-handshake
**Status:** active
**Chosen:** the implementer and gated reviewer stubs carry the MODEL HANDSHAKE paragraph verbatim on
one line, as the panel's stub does, and the paragraph stays in `implement.md`, so **The handshake**'s
"every dispatch prompt in this pipeline carries that paragraph verbatim" holds unchanged and
`check-dispatch-paragraphs` keeps its two `implement.md` handshake blocks.
**Considered:** the handshake in the brief alone — a child that answers without reading its brief
would return no `Model:` line for the parent to compare.

### A fixer's work is one paragraph of the implementer brief

**ID:** fix-dispatch-in-implementer-brief
**Status:** active
**Chosen:** every fixer reads `implementer-brief.md`; its FIX DISPATCH paragraph says what a
`FIX INPUT:` line makes it fix and how each kind commits — a finding's fix one commit per task on top
carrying `Task-Id:` and `Flow-Key:`, a re-commit rewriting only the refused unpushed commit, a pick
re-applying the member's commits, a full-suite fix with `Flow-Key: full-suite-fix-<n>` and no
`Task-Id:` — with MUTATION PROOF by pointer when the line names it, and one `fixed:` or `not-fixed:`
head line per finding.
**Considered:** a third `fixer-brief.md` — a second file for one paragraph, with every other paragraph
of the implementer brief still binding; the child reading `gated-review-fix.md` — that file addresses
the parent.

### flow record finding -from numbers a ref-less line

**ID:** from-assigns-refs
**Status:** active
**Chosen:** a gated reviewer writes its findings file with no `ref`; `flow record finding -from` gives
each ref-less line the next free ref — above both the change's highest recorded ref and every ref the
file names — writes the refs into the file before recording, so a re-run rewrites the same rows, and
takes `-slot`, `-round`, `-status`, `-reproducer` and `-dispatch-seq` from its command line. When the
store cannot be read it records nothing and leaves the file as it was.
**Considered:** the reviewer numbering from a first ref its stub names — two gated bundles can be out
at once with neither file loaded, and `-from` rewrites by ref, so the second would overwrite the
first's findings; the parent rewriting refs with a `jq` pass before loading — parent work this change
removes, and a hand pipeline per load.

### The gated reviewer reads its group's bundle

**ID:** gated-context-is-group-bundle
**Status:** active
**Chosen:** a gated pass's `[CONTEXT_BUNDLE_PATHS]` is `dispatch-context-group-<g>.md`, the bundle
already gathered for the task's own group's implementer: the plan header and that group's tasks
beside the proposal and design. The primary prompt's per-task scoping paragraph keeps plan alignment
to the task.
**Considered:** the per-task `task-<n>-context.md` the parent composes (07d7a9fc) — a Read and a Write
per gated task on the parent's 150–340K context, the boundary work this change removes; a Bash
extraction of the task entry and its cited design sections in the boundary call — a hand `awk` over
`tasks.md` and `design.md` per task, the quoting footgun **Process lessons** warns of.
**Supersedes:** 07d7a9fc's per-task context file.

### One marker list per run

**ID:** markers-written-once
**Status:** active
**Chosen:** the parent writes the content-marker list once, before the stage's first dispatch, and
every later dispatch reuses it with a snapshot of its own; the strings it pins — decision IDs, task
titles, section headings — outlive every dispatch of the run.
**Considered:** a fresh list per dispatch, as the rule read before — a heredoc per boundary for the
same lines.

### The contract text lands as planned patches

**ID:** contract-as-patches
**Status:** active
**Chosen:** the planner writes the contract edits as two patches beside this file, applied by Tasks 5
and 6 and read-only to every task, and harvests `verbatim-moves.txt` from them — kan-1060's pattern.
The panel reviews the exact sentences, and the acknowledgement list is committed with the plan, since
a `/flow` implementer never writes `spectre/changes/`.
**Considered:** implementers writing the prose from this design — the wording would be decided at
implementation, and the acknowledgement list could only be transcribed afterwards from each task's
FAIL lines.

### The scope is flow.sdd-tdd's own fixes

**ID:** sdd-tdd-scope
**Status:** active
**Chosen:** the four fix kinds of `fixes-to-fresh-one-shots` are the change. The review panel's own
inline Minor-only commit (**Panel re-runs**, `skills/flow/review-panel.md`) and `flow.verify`'s fix
loop are unchanged; the full-suite fix, which a pass-1 panel runs, is in.
**Considered:** moving the panel's Minor-only commit to a fixer as well — outside the measured stage,
where a real fix round already dispatches the panel-fix subagent.

### /flow-fast follows the tree

**ID:** flow-fast-follows-tree
**Status:** active
**Chosen:** `/flow-fast` reads the same tree, so its `small` and `regular` plans now dispatch an
implementer per group under its existing `sdd` path; its own exception — no `close-task.sh`, a group
closed on its dispatch end and a push — stays, and its only edit is the re-pointed PIN BEFORE
REFACTOR citation. Its whole `flow.sdd-tdd` spend was $6 in 30 days.
**Considered:** reworking `/flow-fast`'s `sdd` path onto the one-call boundary — the boundary is built
on `close-task.sh`, which `/flow-fast` does not run.

## Open questions

None.

## Recommended approach

Make `sdd` the execution of every class but `micro` (`sdd-below-micro`) and make the `sdd` parent a
coordinator that waits and judges: each child is a one-shot launched in the background on a stub
prompt naming a brief on disk (`briefs-on-disk`, `stub-carries-handshake`), returns its summary as its
final message (`summary-head-no-merger`), and wakes the parent through its completion notification
(`notification-wait`); a group boundary is one Bash call (`one-call-boundary`); every fix the stage
needs is a fresh fixer (`fixes-to-fresh-one-shots`, `fix-dispatch-in-implementer-brief`), the gated
ones batched at the stage's end (`gated-fix-at-stage-end`); the store refuses the parent's own fix row
under `sdd` (`store-refuses-inline-under-sdd`); and the parent keeps every judgement
(`main-keeps-judgement`). The contract text is planned as two patches (`contract-as-patches`).

Rejected: a merger per boundary (`summary-head-no-merger`), a boundary program
(`one-call-boundary`), a fixer at every boundary (`gated-fix-at-stage-end`), a new fixer role
(`fixes-to-fresh-one-shots`), a hook on the parent's edits (`store-refuses-inline-under-sdd`), and
keeping `small` or `regular` inline (`sdd-below-micro`).

## Critical files

- `stats/internal/guard/planclass.go` — `pcTree`, the execution line of the tree.
- `stats/internal/guard/plan_class_test.go` — the tree's expectations per class.
- `scripts/plan-class.sh` — the tree's contract header.
- `stats/cmd/flow/record.go` — the `dispatch begin` decision refusals; `runRecordFindingFrom`.
- `stats/cmd/flow/record_test.go` — `beginUnderDecision` and the refusal tables; the `-from` tests.
- `hooks/enforce-agent-baseline.py` — `ROW_KEY`, the closed list of description keys.
- `stats/internal/repotest/enforce_agent_baseline_hook_test.go` — the hook's key cases.
- `stats/internal/guard/dispatchparagraphs.go` — `dpSites`, which file carries each dispatch paragraph.
- `stats/internal/guard/check_dispatch_paragraphs_test.go` — the guard's fixture trees and parity cases.
- `scripts/check-dispatch-paragraphs.sh` — the site table its header documents.
- `stats/internal/prompttuning/prompttuning.go`, `stats/internal/prompttuning/prompttuning_test.go` —
  the implementer brief joins the `impl` prompt area.
- `stats/internal/guard/stagetaskfiles.go`, `scripts/stage-task-files.sh` — header citations of FLOW —
  COMMIT-PER-TASK.
- `skills/flow/implement.md` — the stubs, the boundary, the fixer table, the waits.
- `skills/flow/implementer-brief.md`, `skills/flow/gated-reviewer-brief.md` — new, the briefs.
- `skills/flow/gated-review-fix.md` — the stage-end fixer and re-review.
- `skills/flow/sdd-dispatch.md` — the pick conflict's fixer.
- `skills/flow/brainstorm-planner.md` — Decide step 1.
- `skills/flow/SKILL-rationale.md` — the reasoning of the changed rules.
- `skills/flow/verify-fix-loop.md`, `skills/flow/review-panel.md`, `skills/flow-fast/SKILL.md`,
  `skills/flow-self-review/SKILL.md`, `skills/flow-contracts/plan-amendment.md` — citations of the
  moved paragraphs.
- `contract-edits/briefs-on-disk.patch`, `contract-edits/one-shot-sdd.patch`, `verbatim-moves.txt` —
  beside this file, the planned contract text and its acknowledgement list.

## Existing code to reuse

- `close-task.sh`, `check-tree-markers.sh`, `check-plan-unchanged.sh`, `gather-dispatch-context.sh` —
  every link of the one-call boundary; no new program.
- `hooks/flow-stop-guard.py`'s `agents_pending` — already lets a turn end while a launched agent has
  not notified.
- The review panel's ceiling timer (**No forking, and a wall-clock ceiling on every slot**,
  `skills/flow/review-panel.md`) — the background `sleep` mechanism the keep-warm timer reuses.
- `render-slot-prompt.sh`'s pointer-to-a-brief prompt shape (**Bundled dispatch**,
  `skills/flow/review-panel.md`) — the shape the stubs follow, without a renderer.
- `stats/cmd/flow/record.go`'s refusal of a subagent row under an `inline` decision and its table test
  `TestRecordDispatchBeginRefusesSubagentUnderInline` with `beginUnderDecision` — mirrored for `sdd`.
- `runRecordFindingFrom` and `runRecordNextRef` — the loader gains ref numbering from the next-ref
  read it already has.
- `TestEnforceAgentBaselineHookAdmitsRestKeys`'s open-row fixture — the shape of the new key's test.
- The decision's existing `fixer` pair, the `implementer` role's **Allowed pairs** rows and the
  `task-<ids>-implementer-fix-<k>` key family `check-task-reviewer-single-dispatch` and the hook
  already accept.

## Verification

- Each task's own verify step: its targeted Go tests (`go test -run`), `gofmt -l`, `go vet`, and for
  Tasks 5 and 6 `run-verify-list.sh <worktree> lint`, which carries `check-dispatch-paragraphs.sh`,
  `check-verbatim-moves.sh`, `check-references.sh` and the rest, with the normative inventory compared
  before and after.
- The full `## test` list after the last group: `guard-tests` (`run-guard-tests.sh`, which runs the
  Go guard tests), `stats-go` (the `record.go` and repotest hook tests).
- `scripts/plan-dispatch-bundles.sh` and `scripts/plan-class.sh` over this change's own `tasks.md`.
- The `## Live check` below at `flow.verify`, and the board read after landing it names.

## Live check

Every step runs at `flow.verify`, against this worktree's own build and a scratch copy of the dev
store: `scripts/live-store-copy.sh up kan-1062`, its printed `FLOW_ADDR`/`FLOW_RECORDS_ADDR` exported,
and `scripts/live-store-copy.sh down kan-1062` at the end. The `flow` CLI is the worktree's own build
(`make -C stats build`, then `stats/bin/flow`); "before" is the installed CLI and the main checkout's
`plan-class.sh`, which still carry the base.

1. **The tree.** `scripts/plan-class.sh spectre/changes/kan-1062-flow-implement-one-shot-subagents/tasks.md 1`
   prints a `tree:` line reading `execution sdd · implementer chosen`, and so does the same call on an
   archived plan whose class is `regular` or `small` (`spectre/changes/archive/*/tasks.md`, the first
   whose `class:` line reads either). Before, from the main checkout's script: `execution inline`.
   Not working: `inline` still prints for any class but `micro`.
2. **The refusal.** On the scratch store, record for a scratch change `kan-1062-live` a copy of this
   change's `.superpowers/sdd/decision.json` with `execution` set to `sdd`, through
   `stats/bin/flow record decision -change kan-1062-live -session-token mf-kan1062live-<UTC start time> -file <the copy>`,
   then
   `stats/bin/flow record dispatch begin -change kan-1062-live -role implementer -agent-id inline -key task-1-implementer -model opus -effort low -session-token mf-kan1062live-<UTC start time>`
   exits 2 naming `execution: sdd`; the same with `-key full-suite-fix-1` exits 2; with
   `-key self-review-x-fix` it begins (exit 0). Before, with the installed CLI: the first call exits 0.
   Not working: any of the first two exits 0.
3. **Ref numbering.** On the same scratch change, a file of two ref-less lines
   (`{"task":1,"severity":"Minor","location":"a.go:1","note":"x"}`) loaded with
   `stats/bin/flow record finding -change kan-1062-live -slot gated-reviewer -round 0 -status open -reproducer none -dispatch-seq 1 -from <file>`
   records two findings whose refs follow the change's highest, and the file now names those refs;
   loading it again prints `updated:` for both and adds no ref. Before, with the installed CLI: the
   load is refused for the missing `-ref`. Not working: a third ref appears, or the second load
   prints `recorded:`.
4. **The hook.** The Go test `TestEnforceAgentBaselineHookAdmitsFullSuiteFixKey` run against the
   worktree's hook admits a `full-suite-fix-1 …` description with its row open and denies
   `full-suite-fixes …`; before, the first is denied too.

The acceptance itself — main share down on the board — needs real runs after landing, so verify
cannot measure it. After about five `/flow` runs, read
`http://127.0.0.1:4173/api/v1/stats/stage-leaderboard` over the post-landing window: `flow.sdd-tdd`'s
`mainSharePct` beside its mean cost per run. The before figures are 45.8% and $12.44. A main share that
stays near 45%, or a mean cost per run that rises, is the change not working.
