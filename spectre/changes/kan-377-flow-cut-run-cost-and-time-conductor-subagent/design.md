# Design — kan-377-flow-cut-run-cost-and-time-conductor-subagent

## Context

Architectural change to this repository's own `/flow` skills and its stats store. `proposal.md` carries why and the measurements; this file carries how. Approved 2026-09-02 as
`docs/superpowers/specs/2026-09-02-kan-377-flow-cut-run-cost-and-time-conductor-subagent-design.md`.

## 1. Conductor

**Scope.** On a creating run, once the planner's `## Plan` returns, and on a fix run, once the
parent has run `flow.document-fix`, the parent dispatches one **conductor** subagent that runs
`flow.load-context` through `flow.write-in-progress` — `skills/flow/implement.md` §1, §2 and §4,
`skills/flow/review-panel.md`, `skills/flow/verify-and-handoff.md` — following those files as the
conductor; every "you" in them already addresses it.

**Placement.** `implement.md` opens with a **Dispatch the conductor** section (the twin of
brainstorm.md's *Dispatch the planner*, cited for the shared mechanics, never copied). Its §3
`flow.document-fix` moves above that section: on a fix run the parent resolves the worktree from
the state file's `worktrees` map, runs the planner dispatch and the Jira description sync, then
dispatches the conductor. Fix-run stage order: document-fix → load-context → isolate (resume) →
sdd-tdd → …, so the appended plan is validated after the edit. `SKILL.md`'s router sends both run
kinds through that section.

**Prompt.** The agent-baseline pointer; `<name>`; project root; `<changeRoot>`; the literal session
token; the harness; `DEFAULT_MODEL`; the resolved `REVIEWERS` list; the guard-presence result; the
run kind and, on a fix run, the fix instructions; the instruction to read the three phase files as
the conductor; the relay contract. The conductor runs every stage mark in its range itself — a
depth-2 transcript carries the top session's `sessionId`, so its marks bind exactly as the parent's.

**Model.** `DEFAULT_MODEL`, recorded on the row; a plain-language session instruction overrides it
for that run only, recorded with the dispatch, never written back. No settings field.

**Record and handshake.** The launch is asynchronous and returns the agent id:

```bash verified:flag set from stats/cmd/flow/record.go's usage lines 84-88 at main (d4d6cbe); -role conductor is the value task 1 adds
flow record dispatch begin -change <name> -role conductor -model <DEFAULT_MODEL> \
  -key conductor -agent-id <id> -session-token mf-<literal-token> -started-at <ts>
```

recorded immediately after the launch returns, before anything else. The first line of the
conductor's first reply is `Model: <its own model>`; a mismatch against `DEFAULT_MODEL` (or the
override) records `end -key conductor -outcome fallback` and `begin -key conductor-<model> -model
<model>` and continues on the running agent — no re-dispatch.

**Returns.** The conductor ends a turn only with one of:

- `## Question` — question plus named options; the parent asks it verbatim and resumes the
  conductor with the answer. Every operator prompt inside the covered stages goes this way: the
  over-cap choice, a second wall-clock breach, the non-converging-finding handback, BLOCKED, an
  empty resolved-worktree set, a plan-quality repair needing the operator.
- `## Stage flow.<key>` plus one line of outcome, at every `flow stage end` it runs; the parent
  updates the harness task list and resumes with `continue`. **Progress visibility**
  (`skills/flow-contracts/pipeline.md`) gains one sentence: on the implementation branch the task
  list's granularity is the stage.
- `## Handoff` carrying the `IN_PROGRESS` handoff block verbatim; the parent prints it unchanged
  and records `end -key <the key open> -outcome completed -ended-at <ts>`.

**The conductor never ends a turn with a child in flight.** It waits for every implementer,
reviewer, slot and fix subagent before a `## Stage` or `## Question`. Every prompt above is reached
with no child running: the over-cap prompt precedes the slot dispatch, a second breach follows the
slot's stop, BLOCKED and the handback follow a return.

**Failure.** A conductor that ends without one of the three blocks, or whose agent dies, is closed
`-outcome aborted`; the parent reports it and prints `/flow <name>` — the existing re-entry rules
(checkbox state, the state file's worktrees, findings in the store) resume from where it stopped.

**At depth** the Agent tool offers no `bugbot` or `security-review` type; `review-panel.md`'s
substitution rule applies and the file says this is now the norm.

## 2. Panel rounds

**Minor.** Every finding a round raises goes to the fix subagent; each is closed by the existing
check — reproducer re-run exits 0 and `fix-round-N.diff` (still written from `FIX_BASE`) touches
a path the finding named with a non-comment, non-whitespace change. When every finding the round
raised was Minor, no slot re-runs: proceed to `check-panel-findings-closed.sh` and the close. A
finding failing verification takes the handback loop; that loop re-runs no slot either.

**Re-run rule** — replaces the Targeted/Full mode table, the five auto-escalation triggers,
"Targeting is a cost optimization…" and the *The diff-size cap check on a re-run* section. When a
round raised anything above Minor:

- every diff-reading slot in the resolved roster, plus any operator-added slot already in play,
  re-runs on `git diff <the HEAD sha it last read> HEAD`, written to
  `slot-delta-<round>-<slot>.diff`; Primary reads a delta like the rest. An empty delta is not
  dispatched: `not re-run — nothing new since its last read`;
- Bugbot and Security, which read no diff, re-run only when that slot raised a finding in the
  previous round or the previous round raised a new Critical;
- the cap check is `check-panel-diff-size.sh <worktree> <sha>` once per distinct last-read sha
  among the diff-reading slots dispatched; the gating count is the largest; exit 1 puts the existing
  over-cap choice to the operator; both counts go in `final-review-panel.md`.

**Stale.** `SKILL.md`'s and `verify-and-handoff.md`'s "stale clean result" guardrails and
review-panel's closing "no stale result" read: a slot's clean result is stale when the re-run rule
required that slot to re-run and it has not.

**Plan fields.** The fix subagent updates the `Baseline:` counts and `Files:` paths its fix
invalidates in the worktree's `tasks.md` — an uncommitted edit to a planning path — in the same
pass.

**Reports.** Each slot's prompt names `<abs-worktree>/.superpowers/sdd/panel-report-<round>-<id>.md`
and requires the slot to write its full report there before ending its turn; its return message is
the findings summary. After the slot's `dispatch end` the conductor confirms the file is non-empty,
otherwise writes `no verbatim report captured — <reason>`. The conductor records `F<n>` rows and
never re-emits a report.

**Trims.** superpowers' `code-reviewer.md`, `principles-reviewer-prompt.md` and the
subagent-driven-development skill are passed by absolute path, never read into the orchestrating
context; the dispatch bundle is written and never read back by the orchestrator.

## 3. Per-task review overlapped

The unit is the bundle `plan-dispatch-bundles.sh` emits; "bundle N's reviewers" is one reviewer per
task in it. `implement.md` §4's loop at each boundary, in order:

1. Bundle N+1's implementer commits.
2. If bundle N's review raised a fix: resume bundle N's implementer (`SendMessage`, key
   `task-<n>-implementer-fix-<k>`, `-agent-id` the implementer's own id); it commits `git commit
   --fixup=<task-sha>` and runs `git rebase --autosquash`; a conflict is between two of the
   branch's own commits and the implementer resolves it — `integrate.md`'s never-auto-resolve rule
   concerns the operator's base branch.
3. `check-task-commit-fields.sh` runs for bundle N+1's tasks and for every task whose sha the
   rebase rewrote — it reads git objects and `tasks.md` only; a failure goes back to the same
   implementer before anything below.
4. Dispatch together: bundle N's re-reviewer if step 2 ran (`<task-base>..<new-task-sha>`), bundle
   N+1's reviewers (`git diff <task-base>..<task-sha>`), and bundle N+2's implementer.

A clean review ticks the task's checkbox. The last bundle's reviewers run alone; the conductor may
overlap them with the panel's pre-work only (citation check, bundle rebuild, relocation comparison,
diff-size check); `final-review.diff` is written and the slots dispatched once the last review is
clean and any fix folded. "At most one implementer in flight per worktree" is unchanged. Every
reviewer's `dispatch end` carries `-outcome clean` or `-outcome fix`.

## 4. Records and the store

- `-agent-id` on `begin` for every dispatch the conductor makes — implementer, reviewer,
  panel-fix, red-partner — recorded immediately after the launch returns; `end` may repeat it.
- `bestDispatchWindow` (`stats/internal/harvest/attribute.go`): when the id pass finds more than one
  window, filter those by `contains(ts)`; exactly one → attributed; otherwise the narrowed set is
  the ambiguity reported.
- `rawMessage` gains `ID` (`json:"id"`); `ParseAssistantRecords` keeps the first record per
  non-empty `message.id` within one call. Ceiling, marked in code: a message whose lines straddle
  two reads counts twice — lines of one message are written in one burst, so this is bounded by one
  message per read boundary.
- Pricing row `claude-fable-5-1`, from the published pricing page on 2026-09-02: input 10, 5m
  cache write 12.50, 1h cache write 20, cache read 0.25 (0.025× — the one model whose read is not
  0.1×), output 50, no fast-mode rate.
- `conductor` in `recordRoles` (`stats/cmd/flow/record.go`).

## Files

`skills/flow/SKILL.md`, `skills/flow/implement.md`, `skills/flow/review-panel.md`,
`skills/flow/verify-and-handoff.md`, `skills/flow-contracts/pipeline.md`,
`skills/flow-contracts/model-policy.md` (the role list line), `commands/flow.md` and
`commands-claude/flow.md` (one line each), `scripts/check-contract-budget.sh` (rows for files that
grow), `stats/cmd/flow/record.go`, `stats/internal/harvest/transcript.go`,
`stats/internal/harvest/attribute.go`, `stats/internal/store/pricing_seed.go`, each with its test.

## Testing and guards

Go, RED before GREEN: two lines sharing an id yield one record and a line with no id is kept;
`SeedPricingRates` prices a fable-only stage run; `-role conductor` accepted and `-role foreman`
refused; `bestDispatchWindow` with two same-id windows of which one contains `ts` attributes to it,
and with neither containing it reports both. Skills: `check-dispatch-paragraphs.sh` sites and
minimums preserved; `check-stage-mark-calls.sh`; `check-references.sh`; `check-contract-budget.sh`
rows raised in the same commit as the growth; `check-normative-inventory.sh` captured before and
after each prose task with every difference named in the commit body; the acceptance grep empty;
`cd stats && go vet ./... && gofmt -l .` and `cd stats/web && npx tsc -b` clean.

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
**Chosen:** no harvester change — measured on Claude Code 2.1.257 with a haiku probe dispatched from
the depth-1 planner: the depth-2 transcript landed at
`~/.claude/projects/<project>/<top-session>/subagents/agent-<id>.jsonl`, the same directory as every
depth-1 transcript, with `sessionId` equal to the top session, `isSidechain: true`, and a
`.meta.json` carrying `parentAgentId` and `spawnDepth: 2`.
**Considered:** the research note's parent-directory binding rule — built on a probe result this
planning session could not reproduce.

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
**Chosen:** every diff-reading slot re-runs on its delta, Primary included (all 18 of its fix-round
findings sat in a file the fix touched); Bugbot and Security only as prior raiser or on a new
Critical; one cap-check call per distinct delta sha.
**Considered:** prune the always-true triggers and keep the ladder — buys little, since deltas
already removed the whole-diff re-reads the ladder's cost came from; keep the ladder unchanged.

### A clean result is stale only when a required re-run is missing

**ID:** stale-means-required-rerun-missing
**Status:** active
**Chosen:** stale = the re-run rule required the slot and it did not run — reconciles the issue's
"Minor fixes do not stale a result" with Bugbot's prior-raiser rule; a non-Minor fix Bugbot did not
raise is covered by the round's own mutation-proof.
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
id-ambiguous set by `contains(ts)` — attributes kan-374's three resumed-fix rows.
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
