# Design — kan-377: cut `/flow` run cost and time

Source: KAN-377. Brainstormed 2026-09-02 from the research note that issue was written from; the
numbers below name their source, and two of the note's claims are corrected here on measurements
taken during planning.

## Context

- 38 `/flow` changes since 2026-08-20 cost $2,578 in the store (~$68 per change): `flow.sdd-tdd`
  41%, `flow.review-panel` 32%, everything else ≤6% — as the store prices them.
- The store over-counts: `ParseAssistantRecords` (`stats/internal/harvest/transcript.go`) emits one
  usage record per JSONL line, and Claude Code writes one line per content block with identical
  `usage`. Measured on this planner's own transcript during planning: 56 assistant lines, 15
  `message.id`s, one message on 8 lines. `claude-fable-5-1` has no pricing row, so the parent
  session's cost reads $0 in every run since the session moved to fable.
- kan-374, deduplicated: the parent spent 66.5M context-tokens over 183 calls (41% of tokens) against
  95.0M for all subagents; panel stage 41 calls × 486K median context, sdd-tdd 57 × 348K, the
  integrate/archive tail 26 × 615K. Fable is $10/$50 per MTok — 5× sonnet, not the "≈ opus, 2.5×"
  the note assumed — so the parent's dollar share is above half, and every parent-side saving is
  worth more than the note projected.
- Panel: 77 rounds, 262 slot dispatches; 45 re-run rounds (130 dispatches), 94 dispatches whose only
  output was "nothing new"; 9 re-run rounds triggered by Minor findings alone; 55 of 62 fix-round
  findings came from Primary or a prior raiser, 7 from a slot running only because of Full
  escalation (0 Critical); a non-raiser Bugbot re-run yielded zero across 32 changes.
- Per-task review is strictly serial: 69 of 290 implementation minutes (24%) over 8 changes.

## 1. Conductor — the implementation half runs in one subagent

**Scope.** On a creating run, once the planner's `## Plan` returns, and on a fix run, once the
parent has run `flow.document-fix`, the parent dispatches one **conductor** subagent that runs
`flow.load-context` through `flow.write-in-progress` — `skills/flow/implement.md` §1, §2 and §4,
`skills/flow/review-panel.md`, `skills/flow/verify-and-handoff.md` — following those files as the
conductor. Every "you" in them already addresses it. The parent becomes: kickoff, planner relay,
one conductor dispatch and relay, integrate/archive.

**Placement.** `skills/flow/implement.md` opens with a **Dispatch the conductor** section (the twin
of brainstorm.md's *Dispatch the planner*, cited for the shared mechanics, not copied). Its §3
`flow.document-fix` moves above that section: on a fix run the parent resolves the worktree from the
state file's `worktrees` map, runs the planner dispatch and the Jira description sync, and only then
dispatches the conductor. The fix-run stage order becomes document-fix → load-context → isolate
(resume) → sdd-tdd → …, which validates the appended plan after the fix's edit rather than before.
`skills/flow/SKILL.md`'s router sends both run kinds through that section.

**The prompt** carries, verbatim where the text is fixed: the agent-baseline pointer; `<name>`;
the project root; `<changeRoot>`; the literal session token; the harness; `DEFAULT_MODEL`; the
resolved `REVIEWERS` list; the guard-presence result; the run kind and, on a fix run, the operator's
fix instructions; the instruction to read the three phase files as the conductor; and the relay
contract below. The conductor runs every stage mark in its range itself — a depth-2 transcript
carries the top session's `sessionId` (measured, below), so its marks bind exactly as the parent's.

**Model.** `DEFAULT_MODEL`, recorded on the dispatch row; a plain-language session instruction
overrides it for that run only, recorded with the dispatch, never written back. No settings field.

**Record and handshake.** The launch is asynchronous and returns the agent id, so:

```bash
flow record dispatch begin -change <name> -role conductor -model <DEFAULT_MODEL> \
  -key conductor -agent-id <id> -session-token mf-<literal-token> -started-at <ts>
```

immediately after the launch returns, before anything else. `conductor` joins `recordRoles` in
`stats/cmd/flow/record.go`. The conductor's first line is `Model: <its own model>`; a mismatch
against `DEFAULT_MODEL` (or the override) records `end -key conductor -outcome fallback` and
`begin -key conductor-<model> -model <model>` and **continues on the running agent — no
re-dispatch**. The planner's opus re-dispatch exists because fable may be unavailable;
`DEFAULT_MODEL` has no such case, and a re-dispatch to opus would raise the cost the conductor exists
to cut.

**Returns — the relay contract.** The conductor ends a turn only with one of:

- `## Question` — the question plus named options; the parent asks it verbatim through the
  harness's question tool and resumes the conductor with the answer. Every operator prompt inside
  the covered stages goes this way: the over-cap choice, a second wall-clock breach, the
  non-converging-finding handback, BLOCKED, an empty resolved-worktree set, a plan-quality repair
  that needs the operator.
- `## Stage flow.<key>` plus one line of outcome — at every `flow stage end` it runs. The parent
  updates the harness task list and resumes with `continue`. **Progress visibility**
  (`skills/flow-contracts/pipeline.md`) gains one sentence: on the implementation branch the task
  list's granularity is the stage, because the conductor has no task-list tool — the accepted
  visibility trade.
- `## Handoff` carrying the `IN_PROGRESS` handoff block verbatim. The parent prints it unchanged,
  records `end -key <the key open> -outcome completed -ended-at <ts>`, and the run is at
  `IN_PROGRESS`.

**The conductor never ends a turn with a child in flight.** It waits for every implementer,
reviewer, slot and fix subagent it launched before a `## Stage` or `## Question` — a notification
arriving to an ended turn is the failure this rule prevents. Every prompt above is reached with no
child running: the over-cap prompt precedes the slot dispatch, a second breach follows the slot's
stop, BLOCKED and the handback follow a return.

**Failure.** A conductor that ends without one of the three blocks, or whose agent dies, is closed
`-outcome aborted`; the parent reports it and prints `/flow <name>` for the operator to re-run —
the existing re-entry rules (checkbox state, the state file's worktrees, findings in the store)
resume from wherever it stopped. No automatic retry: a re-run is the same act, and the operator
should see the death.

**At depth**, the Agent tool offers no `bugbot` or `security-review` type. `review-panel.md`'s
existing "an unspawnable id is substituted, not skipped" rule applies; the file says this is now
the norm rather than the exception (4 `bugbot` dispatches ever against 1,659 `general-purpose`).

**Dropped: nested-transcript session binding.** The note proposed binding a `<X>/subagents/*.jsonl`
transcript to the session holding `subagents/agent-<X>.jsonl`. Measured during planning (Claude
Code 2.1.257) with a haiku probe dispatched from this depth-1 planner: the depth-2 transcript
landed at `~/.claude/projects/<project>/<top-session>/subagents/agent-<id>.jsonl` — the same
directory as every depth-1 transcript — with `sessionId` equal to the top session, `isSidechain:
true`, and a `.meta.json` carrying `parentAgentId` and `spawnDepth: 2`. `resolveSessionTokens`,
`discoverTranscripts` and per-dispatch `-agent-id` attribution need no change. The note's contrary
probe result is superseded by this measurement.

**Expected (kan-374 as the reference, from the note):** parent 66.5M → ~15M context-tokens per run;
the tail's 26 calls 16.3M → ~4M because the parent enters integrate at ~150K instead of 531K; ~150K
output tokens move from fable to `DEFAULT_MODEL`.

## 2. Panel rounds

**A Minor finding blocks and is fixed, but triggers no re-run.** Every finding a round raises goes to
the fix subagent as today, and each is closed by the existing mechanised check — reproducer
re-run exits 0 **and** the fix diff (`fix-round-N.diff`, still written from `FIX_BASE`) touches a
path the finding named with a non-comment, non-whitespace change. When every finding the round
raised was Minor, no slot re-runs: the stage proceeds to `check-panel-findings-closed.sh` and its
close. A finding failing verification takes the handback loop as today; that loop re-runs no slot
either.

**One re-run rule replaces the Targeted/Full ladder.** Deleted from `review-panel.md`: the mode
table, the five auto-escalation triggers, "Targeting is a cost optimization…", and the whole *The
diff-size cap check on a re-run* section. In their place, when a round raised anything above
Minor:

- every diff-reading slot in the resolved roster (plus any operator-added slot already in play)
  re-runs on its delta — `git diff <the HEAD sha that slot last read> HEAD`, written to
  `slot-delta-<round>-<slot>.diff`; Primary reads a delta like the rest (all 18 of its fix-round
  findings sat in a file the fix touched). A slot whose delta is empty is not dispatched and the
  record says `not re-run — nothing new since its last read`;
- Bugbot and Security, which read no diff, re-run only when that slot raised a finding in the
  previous round or the previous round raised a new Critical;
- the cap check is `check-panel-diff-size.sh <worktree> <sha>` once per distinct last-read sha among
  the diff-reading slots dispatched this round; the gating count is the largest, and an exit 1 puts
  the existing over-cap choice to the operator. Both counts are recorded in
  `final-review-panel.md`.

**Stale, redefined.** `skills/flow/SKILL.md`'s and `skills/flow/verify-and-handoff.md`'s "never
hand off with … a stale clean result", and review-panel's closing "no stale result", read: **a
slot's clean result is stale when the re-run rule above required that slot to re-run and it has
not.** A non-Minor fix Bugbot did not raise leaves Bugbot's result current: the round's own
mutation-proof covers what the fix changed.

**The fix subagent maintains the plan fields its fix invalidates** — `Baseline:` counts and
`Files:` paths in the worktree's `tasks.md`, an uncommitted edit to a planning path — in the same
pass, so a fix that adds test cases no longer raises a stale-Baseline Minor on the next round
(kan-88 ran three rounds on exactly that).

**Slots write their own reports.** Each slot's prompt names
`<abs-worktree>/.superpowers/sdd/panel-report-<round>-<id>.md` and requires the slot to write its
full report there before ending its turn; its return message is the findings summary. After the
slot's `dispatch end`, the conductor confirms the file is non-empty and otherwise writes the single
line `no verbatim report captured — <reason>`. The conductor records `F<n>` rows and never re-emits
a report.

**Context trims.** Subagent-facing files — superpowers' `code-reviewer.md`,
`principles-reviewer-prompt.md`, the subagent-driven-development skill — are passed by absolute
path and never read into the orchestrating context; the dispatch bundle is written and never read
back by the orchestrator.

**Expected (from the note):** ≈ −21% panel tokens on kan-374 (round 2 removed, round 1 11.4M →
~6.7M), more on long tails.

## 3. Per-task review overlapped with the next implementer

The unit is the bundle `plan-dispatch-bundles.sh` emits; "task N's reviewer" means one reviewer per
task in bundle N. `implement.md` §4's loop becomes, at each boundary and in this order:

1. Bundle N+1's implementer commits.
2. If bundle N's review raised a fix: resume bundle N's implementer (`SendMessage`, key
   `task-<n>-implementer-fix-<k>`, `-agent-id` the implementer's own id); it commits `git commit
   --fixup=<task-sha>` and runs `git rebase --autosquash`. A conflict there is between two of the
   branch's own commits — the implementer resolves it; `integrate.md`'s never-auto-resolve rule
   concerns the operator's base branch.
3. `check-task-commit-fields.sh` runs for bundle N+1's tasks and for every task whose sha the
   rebase rewrote. The guard reads git objects and `tasks.md` only — safe while the tree changes
   — and a failure goes back to the same implementer before anything below.
4. Dispatch together: bundle N's re-reviewer if step 2 ran (the task's full range,
   `<task-base>..<new-task-sha>`), bundle N+1's reviewers (`git diff <task-base>..<task-sha>`, an
   immutable range) and bundle N+2's implementer.

A clean review ticks the task's checkbox (`flow tasks tick`) — the guard has already passed. The
last bundle's reviewers run alone; the conductor may overlap them with the panel's pre-work only
(citation check, bundle rebuild, relocation comparison, diff-size check); `final-review.diff` is
written and the slots dispatched only once the last review is clean and any fix folded. "At most
one implementer in flight per worktree" is unchanged: a reviewer reads a committed range.

Every reviewer's `dispatch end` carries `-outcome clean` or `-outcome fix`, so per-task review yield
becomes measurable — `-outcome` is unvalidated text today; no Go change.

**Expected (from the note):** −24% implementation wall time, tokens unchanged.

## 4. Records and the store

- **`-agent-id` on `begin` for every dispatch** the conductor makes — implementer, reviewer,
  panel-fix, red-partner — since every launch is asynchronous and returns the id at launch;
  `begin` is recorded immediately after the launch returns, before any other action, and `end` may
  repeat the id. Two dispatches starting at one instant are told apart only by id; a resumed fix
  dispatch shares the implementer's id — kan-374's three unattributed `task-N-implementer-fix-1`
  rows.
- **`bestDispatchWindow`** (`stats/internal/harvest/attribute.go`): when the id pass finds more than
  one window, filter those by `contains(ts)`; exactly one → attributed; otherwise the narrowed set
  is the ambiguity reported.
- **Harvester dedupe.** `rawMessage` gains `ID` (`json:"id"`); `ParseAssistantRecords` keeps the
  first record per non-empty `message.id` within one call. Lines of one message share every usage
  field, so the first suffices. Ceiling, marked in code: a message whose lines straddle two reads
  counts twice; the lines of one message are written in one burst, so this is rare and bounded by
  one message per read boundary.
- **Pricing row** for `claude-fable-5-1`, read from the published pricing page on 2026-09-02:
  input 10, 5m cache write 12.50, 1h cache write 20, cache read 0.25 (0.025× — the one model whose
  read is not 0.1×), output 50, no fast-mode rate (fast mode is Opus 5/4.8 only).
- **`conductor`** in `recordRoles`.

## 5. Deliberately unchanged

The roster (`primary`, `principles`, `code-review-low`, `bugbot`) and Bugbot's brief; no settings
field; no new stage key; no per-task finding rows; `-also-raised-by` on `flow record finding`
not added (overlap is unmeasurable today and no roster change is on the table); the ten
uniquely-id'd unattributed kan-374 slots stay uninvestigated.

## Files

`skills/flow/SKILL.md`, `skills/flow/implement.md`, `skills/flow/review-panel.md`,
`skills/flow/verify-and-handoff.md`, `skills/flow-contracts/pipeline.md`,
`skills/flow-contracts/model-policy.md` (the role list line), `commands/flow.md` and
`commands-claude/flow.md` (one line each), `scripts/check-contract-budget.sh` (rows for files
that grow), `stats/cmd/flow/record.go`, `stats/internal/harvest/transcript.go`,
`stats/internal/harvest/attribute.go`, `stats/internal/store/pricing_seed.go`, each with its test
file.

## Testing and guards

Go, RED before GREEN: `ParseAssistantRecords` on a fixture where two lines share an id yields one
record and a line with no id is kept; `SeedPricingRates` prices a fable-only stage run; `record
dispatch begin -role conductor` is accepted and `-role foreman` refused; `bestDispatchWindow` with
two same-id windows of which one contains `ts` attributes to it, and with neither containing it
reports both. Skills: `check-dispatch-paragraphs.sh` sites and minimum counts preserved through the
rewrite; `check-stage-mark-calls.sh`; `check-references.sh`; `check-contract-budget.sh` rows raised
in the same commit as the growth; `check-normative-inventory.sh` captured before the first prose
edit and after the last, every difference justified in the panel record; the acceptance grep
`grep -n "Targeted\|Full re-run" skills/flow/review-panel.md` empty; `cd stats && go vet ./... &&
gofmt -l .` and `cd stats/web && npx tsc -b` clean.

## Decisions

### The implementation half runs in one conductor subagent, on both run kinds

**ID:** conductor-runs-implementation-half
**Status:** active
**Chosen:** one subagent covering `flow.load-context` through `flow.write-in-progress` on creating
and fix runs alike — removes 42M parent context-tokens and shrinks the tail with one relay
mechanism, and keeps one orchestration path in the phase files.
**Considered:** panel-only conductor — removes 19.6M and leaves the parent at 348K through sdd-tdd
and 615K through the tail; one conductor per phase file — fresh context each but two more
dispatch/relay cycles for a need nothing measured; creating runs only — two "you"s in one text.

### The conductor runs on `DEFAULT_MODEL`

**ID:** conductor-model-default
**Status:** active
**Chosen:** `DEFAULT_MODEL`, session-instruction override, recorded on the row — the orchestrator
already ran on sonnet for 16 of 29 panel changes with no worse outcome, its judgement is fenced by
guards, and the rate delta is ~$5 per run.
**Considered:** `PLANNING_MODEL` — keeps today's judgement, forgoes the saving, no evidence it is
needed; a `conductorModel` settings field — a kan-370-sized change adding configurability nobody
asked for.

### A handshake mismatch continues on the running agent

**ID:** conductor-handshake-no-redispatch
**Status:** active
**Chosen:** record `fallback` on `conductor`, open `conductor-<model>`, continue — the planner's opus
re-dispatch exists for fable's availability, which `DEFAULT_MODEL` does not share.
**Considered:** the planner's opus re-dispatch verbatim — raises cost, addresses no known failure.

### The conductor returns per stage

**ID:** conductor-stage-returns
**Status:** active
**Chosen:** `## Stage` at every stage end, `## Question` for prompts, `## Handoff` at the end; the
parent updates the task list at stage granularity.
**Considered:** question-and-handoff only — no progress visibility for an hour-long run; per-task
returns — relay traffic for a granularity nobody asked for.

### A fix run's `flow.document-fix` runs in the parent, before the conductor

**ID:** fix-run-document-fix-first
**Status:** active
**Chosen:** parent resolves the worktree from the state file, runs document-fix, then dispatches the
conductor — the planner relay and the Jira sync are already parent work, and the appended plan is
validated after the edit.
**Considered:** the conductor runs document-fix — a planner relayed through a relay; the parent runs
§1–§3 then dispatches — the parent reads the plan and worktree, the context it is shedding.

### A dead conductor is reported, not retried

**ID:** conductor-no-auto-retry
**Status:** active
**Chosen:** close `-outcome aborted`, report, print `/flow <name>` — the re-run is the same act and
the operator should see the death.
**Considered:** one automatic re-dispatch as `conductor-2` — hides the failure and duplicates the
re-entry the command already has.

### Nested-transcript binding is not needed

**ID:** nested-binding-not-needed
**Status:** active
**Chosen:** no harvester change — measured on Claude Code 2.1.257: a depth-2 transcript sits in the
top session's `subagents/` directory with the top session's `sessionId`.
**Considered:** the note's parent-directory binding rule — built on a probe result this planning
session could not reproduce.

### A Minor finding blocks and is fixed, and triggers no re-run

**ID:** minor-blocks-no-rerun
**Status:** active
**Chosen:** close by reproducer flip plus diff touch; no slot re-runs when the round raised nothing
above Minor — removes the 9 Minor-only rounds and kan-88's Baseline loop.
**Considered:** Minor does not block — 79 Minors become the operator's work; today's rule — a Minor
re-runs like a Critical.

### Delta re-runs replace the Targeted/Full ladder

**ID:** delta-rerun-replaces-ladder
**Status:** active
**Chosen:** every diff-reading slot re-runs on its delta, Primary included; Bugbot and Security only
as prior raiser or on a new Critical; one cap-check call per distinct delta sha.
**Considered:** prune the always-true triggers and keep the ladder — buys little, since deltas
already removed the whole-diff re-reads the ladder's cost came from; keep the ladder unchanged.

### A clean result is stale only when a required re-run is missing

**ID:** stale-means-required-rerun-missing
**Status:** active
**Chosen:** stale = the re-run rule required the slot and it did not run — reconciles the issue's
"Minor fixes do not stale a result" with Bugbot's prior-raiser rule.
**Considered:** stale = any non-Minor fix after the slot's last read — would block every handoff in
which Bugbot correctly did not re-run.

### The fix subagent maintains `Baseline:` and `Files:`

**ID:** fix-subagent-maintains-plan-fields
**Status:** active
**Chosen:** the fix pass updates the fields its fix invalidates — the loop was plan bookkeeping.
**Considered:** leave it to the next round — that is the three-round loop.

### Slots write their own reports; the orchestrator passes paths

**ID:** orchestrator-context-trims
**Status:** active
**Chosen:** slot writes `panel-report-<round>-<id>.md`, conductor records rows only; subagent-facing
files by path; bundle never re-read — ~20K output and ~47K context tokens off the orchestrator.
**Considered:** keep the heredoc re-emission — the report is the slot's text, not the dispatcher's.

### Per-task review is kept and overlapped with the next implementer

**ID:** per-task-review-overlap
**Status:** active
**Chosen:** reviewers of bundle N run with bundle N+1's implementer; a fix waits for N+1's commit,
resumes N's implementer, folds by fixup + autosquash, then N+2 dispatches.
**Considered:** drop per-task review as a trial — −15% sidechain, −24% wall, unmeasurable risk to
pass 1; batch every per-task fix before the panel — fixes land late and later tasks build on them;
fix in a throwaway worktree and transplant — a second working copy per fix; keep it serial.

### `-agent-id` moves to `begin`; id-ambiguous windows are narrowed by interval

**ID:** agent-id-on-begin
**Status:** active
**Chosen:** every async launch records its id on `begin`; `bestDispatchWindow` filters an
id-ambiguous set by `contains(ts)` — attributes kan-374's resumed-fix rows.
**Considered:** keep `-agent-id` on `end` — loses the id for a dispatch that never returns and
cannot separate two dispatches sharing a start instant.

### The harvester deduplicates by `message.id` within a read

**ID:** harvest-dedupe-by-message-id
**Status:** active
**Chosen:** first record per id per call — every figure in the store drops to its true value; the
straddle case counts one message twice and is marked as the ceiling.
**Considered:** persisting the last-seen id per transcript beside the offset — a schema change for
a case bounded to one message per read boundary.

### A pricing row for `claude-fable-5-1`

**ID:** fable-pricing-row
**Status:** active
**Chosen:** the published rates, read on 2026-09-02, cache read 0.25.
**Considered:** deriving from opus — wrong on every column.

### Reviewer dispatches record `clean` or `fix`

**ID:** reviewer-outcome-word
**Status:** active
**Chosen:** the outcome word on `dispatch end` — per-task yield measurable with no Go change.
**Considered:** per-task finding rows — a second findings vocabulary for a question one word answers.

### The roster and Bugbot's brief stay as they are

**ID:** roster-unchanged
**Status:** active
**Chosen:** four slots, briefs unchanged — every slot raised Criticals no other brief covers; overlap
is unrecorded, so a merge or drop is a guess.
**Considered:** make Bugbot's pre-existing-defect findings non-blocking — most of their cost goes
with the Minor rule anyway; merge Primary and Code-review-low — forbidden by review-panel and
unmeasurable; drop a slot — the data supports none.

## Open questions

None recorded.
