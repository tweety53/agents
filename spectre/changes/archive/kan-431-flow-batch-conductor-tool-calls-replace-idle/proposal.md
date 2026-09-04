# kan-431-flow-batch-conductor-tool-calls-replace-idle

**Jira:** KAN-431

## Why

kan-404's conductor transcript made 191 API calls at about $0.047 each, 29 of them idle polls —
`echo "still waiting"` every one to two seconds until a child's completion notification landed —
and most of the rest one guard, one record or one stage mark per turn
(`docs/superpowers/research/flow-speedup.md`, section 6 round three; cost lever 1). The batched
shape is about 62 calls: $5.5–6.5 and 9–11 minutes saved per change.
<!-- measured: unique message ids and usage rows in ~/.claude/projects/-Users-tweety53-Projects-agents/5d204fa9-d2bd-4e22-8bf8-25b8710cc811/subagents/agent-a0a0c76b220754d6f.jsonl, per docs/superpowers/research/flow-speedup.md section 6 round three @ 29281b9 -->

The polls have a contract cause. **4. Execute** (`skills/flow/implement.md`) says "never end a
turn with a child in flight", and **Dispatch the conductor** closes a turn that ends without one
of its three blocks as `aborted` — so the conductor cannot yield to the harness's notification and
fills the gap with tool calls. No sentence forbids batching; one assumes it does not happen
(`skills/flow/review-panel.md`: "Record it per finding, as its verdict is reached, not batched at
the round's end").

## What changes

- `skills/flow/implement.md` **4. Execute** gains a **Turn discipline** paragraph beside "Never end
  a turn with a child in flight": independent calls share one Bash call and independent launches
  one message; every wait on a child is one foreground bounded loop on `test -s <report>`,
  re-issued while false; `flow stage begin` rides a stage's first command and `flow stage end` its
  last. The task boundary is restated as its batched shape, and the implementer dispatch gains a
  REPORT FILE paragraph naming `<abs-worktree>/.superpowers/sdd/implementer-report-<k>.md`.
- `skills/flow/review-panel.md` cites that paragraph and states its own batches — pre-work in one
  call, launches in one message, records in one, findings in one, reproducers and worktree removal
  in one, `status fixed` in one after the diff walk — and gives the fix subagent a REPORT FILE
  paragraph (`panel-fix-report-<round>.md`). The per-finding sentence is reworded so an aborted
  round still leaves every verdict already reached recorded.
- `skills/flow/verify-and-handoff.md` cites it, batches the verifier's launch, wait, record and
  ledger render, and gives the verifier a REPORT FILE paragraph (`verify-report-<key>.md`).
- `scripts/check-contract-budget.sh`'s row for `skills/flow/review-panel.md` is raised.
- Unchanged: `scripts/gather-dispatch-context.sh`, `scripts/check-task-commit-fields.sh`,
  `scripts/run-reproducer.sh` — the batched calls chain them as they are; the `## Stage` returns
  (the research note's cost lever 4); `scripts/check-dispatch-paragraphs.sh` and
  `scripts/check-stage-mark-calls.sh`.
