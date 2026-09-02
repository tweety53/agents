# kan-377-flow-cut-run-cost-and-time-conductor-subagent

Jira: KAN-377.

## Why

- 38 `/flow` changes since 2026-08-20 cost $2,578 in the store, ~$68 per change: `flow.sdd-tdd`
  41%, `flow.review-panel` 32%, everything else ≤6% — as the store prices them, and it prices
  wrong twice over: `ParseAssistantRecords` emits one usage record per JSONL line while Claude
  Code writes one line per content block with identical `usage` (this change's own planner
  transcript: 56 assistant lines, 15 `message.id`s, one message on 8 lines), and
  <!-- measured: grep -c '"type":"assistant"' and grep -o '"id":"msg_[^"]*"' | sort -u | wc -l on this planner's own transcript under ~/.claude/projects, 2026-09-02 -->
  `claude-fable-5-1` has no pricing row, so the parent session has cost $0 in every run since the
  session moved to fable.
- The parent session's own context is the largest single share. kan-374, deduplicated: parent
  66.5M context-tokens over 183 calls (41% of tokens) against 95.0M for every subagent combined;
  panel stage 41 calls × 486K median context, sdd-tdd 57 × 348K, the integrate/archive tail 26 ×
  615K. Of the 409K context at panel start, ~104K is the parent's own emitted tool inputs (18
  byte-for-byte `panel-report` heredocs and finding notes), ~50K skill text read whole, ~22K
  re-reading persisted tool results. Fable is $10/$50 per MTok — 5× sonnet — so the parent is above
  half the dollars.
- Re-runs are half the panel and mostly confirmation: 77 rounds, 262 slot dispatches, 45 re-run
  rounds (130 dispatches), 94 dispatches whose only output was "nothing new". Of 62 fix-round
  findings, 55 came from Primary or a prior raiser; 7 from a slot running solely because of Full
  escalation (0 Critical). Every panel record trips an auto-escalation trigger — "altered a guard's
  behaviour" is always true here. 9 re-run rounds (25 slot dispatches) were triggered by Minor
  findings alone; kan-88 ran three rounds on one stale `Baseline:`. A non-raiser Bugbot re-run has
  zero yield across 32 changes at 2–4M tokens each, and the fix round already mutation-proves every
  behaviour a fix changed.
- Per-task review is strictly serial: 69 of 290 implementation minutes (24%) across 8 changes, for
  4 visible fix-after-review dispatches over 110 reviewed tasks. The reviewer reads an immutable
  commit range; the serial rule only protects writers.

## What changes

- **Conductor subagent.** One subagent, dispatched by the parent on `DEFAULT_MODEL` with the
  planner pattern's mechanics (`## Question` relay, literal session token, `Model:` handshake, a
  `conductor` role on `flow record dispatch`), runs `flow.load-context` through
  `flow.write-in-progress` on creating **and** fix runs, following `implement.md`,
  `review-panel.md` and `verify-and-handoff.md` as the conductor. It returns `## Stage` per stage,
  relays every operator prompt, and ends with `## Handoff`. The parent becomes kickoff, planner
  relay, one conductor dispatch and integrate/archive; on a fix run it runs `flow.document-fix`
  first. Expected: parent 66.5M → ~15M context-tokens per run.
- **Panel rounds.** A Minor finding blocks and is fixed, but no slot re-runs while nothing above
  Minor was raised. The Targeted/Full ladder, its five triggers and its re-run cap-check section are
  deleted; every diff-reading slot re-runs on its delta since last read, Primary included; Bugbot and
  Security re-run only as a prior raiser or on a new Critical; the cap check is one call per distinct
  delta sha. "Stale clean result" is redefined as a required re-run that did not happen. The fix
  subagent maintains the `Baseline:`/`Files:` fields its fix invalidates. Expected ≈ −21% panel
  tokens on kan-374.
- **Per-task review overlapped.** Bundle N's reviewers dispatch together with bundle N+1's
  implementer; a review fix waits for N+1's commit, resumes N's implementer, folds by fixup +
  autosquash, then N+2 dispatches. Reviewer rows record `clean`/`fix`. Expected −24%
  implementation wall time.
- **Orchestrator trims.** Slots write their own report files; the conductor records `F<n>` rows
  only; subagent-facing files are passed by path; the dispatch bundle is never read back.
- **Store.** The harvester deduplicates usage by `message.id`; a `claude-fable-5-1` rate row;
  `conductor` in the role list; `-agent-id` on `dispatch begin` for every dispatch;
  `bestDispatchWindow` narrows an id-ambiguous window set by interval.
- **Not changed:** the roster and Bugbot's brief; no settings field; no new stage key; the
  nested-transcript session binding the issue lists — measured unnecessary during planning
  (`design.md`, `nested-binding-not-needed`).

## Acceptance

- A creating or fix `/flow` run in Claude Code performs load-context through write-in-progress
  inside one conductor dispatch recorded under role `conductor` with its model; the parent records
  no implementer or slot dispatch itself.
- A fix round whose findings were all Minor dispatches no panel slot; a non-Minor round dispatches
  every diff-reading slot on its delta, Bugbot only when it raised or a new Critical appeared.
  `grep -n "Targeted\|Full re-run" skills/flow/review-panel.md` returns nothing.
- Bundle N+1's implementer dispatch begins before bundle N's reviewer dispatch ends, visible in the
  dispatch records.
- Harvested usage for a known transcript equals its message-id count, not its line count; a
  fable-parent run shows non-zero USD.
- `cd stats && go vet ./... && gofmt -l .`, `cd stats/web && npx tsc -b` and every
  `scripts/check-*.sh` guard exit clean.
