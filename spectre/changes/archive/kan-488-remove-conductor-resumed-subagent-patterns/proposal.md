# kan-488-remove-conductor-resumed-subagent-patterns

## Why

`/flow`'s conductor subagent, and the planner/researcher subagents behind it, are all
work-resume-work agents: dispatched once, then resumed across turns as the operator answers
questions or a stage's report arrives. Every subagent in Claude Code (unlike the main session,
which gets a 1-hour prompt-cache TTL on a subscription) sits on a 5-minute TTL, so any gap longer
than that TTL — a report wait, an operator answer that takes a while, a child-completion wakeup —
cold-starts the resumed agent and rewrites its whole context as a full-price cache write.
<!-- measured: code.claude.com/docs/en/prompt-caching, read and confirmed by the operator during the research session -->

Measured on real transcripts: kan-472's conductor wrote 19.3M cache-creation / 358M cache-read
tokens across a 7.5h run (~$120), with 37 full-context rewrites of 130k–920k tokens each — most
following report waits or resumes past the 5-minute TTL.
<!-- measured: usage-row analysis of ~/.claude/projects/-Users-tweety53-Projects-agents/3c753c6b…/subagents/agent-a7561c34….jsonl during the research session; the transcript is machine-local and this figure cannot be re-run, only re-read from the (now-deleted) staging note's own citation -->
The planner and `/flow-research`'s
researcher hit the same mechanism on a smaller scale (~118k tokens per cold resume). The parent
session, on the 1-hour TTL, rewrote only twice in the same 9.5h window.

## What changes

- **The conductor is deleted.** The parent session orchestrates `/flow`'s implement stage directly
  (`implement.md` §1/2/4, `review-panel.md`, `verify-and-handoff.md`) — its own Bash/Read work,
  `flow record`/`flow stage` marks, worktree add/remove, report reads, diff walks,
  `AskUserQuestion` calls. The `## Question`/`## Stage`/`## Handoff` relay contract, the
  conductor's model handshake and its `-role conductor` dispatch record all go.
- **The planner and `/flow-research`'s researcher are deleted too.** `/flow`'s brainstorming
  (`brainstorm-planner.md` sections B, C, D) and `/flow-research` (any argument shape) both run
  entirely in the current session, on the session's own model — no dispatched subagent, no relay,
  no handshake, no `opus` fallback. Questions become direct, batched `AskUserQuestion` calls (up to
  four independent questions per call) from the session itself.
- **Implementer, panel-bundle, panel-fix and verifier dispatch are unchanged** — they stay
  one-shot: dispatched fresh per task or round, write a report, done. Only the resumed middle
  layers (conductor, planner, researcher) are cut.
- **A strict, deterministic research-artifact path.** `/flow-research` writes to exactly one of
  three destinations — `<jira-key>.md` (no change exists yet), the existing change's `design.md`
  (change exists), or `<topic-slug>.md` (no key, bare session) — and `/flow KAN-XXX` answers "does
  a seed exist" with one `test -f`, never a glob or an inference. The `<key>-*.md` wildcard
  fallback and its disambiguation prompt are removed.
- **Startup visibility.** `/flow` prints whether it found a seed (and, if so, the resolved
  planning/toggle/model choices) right at kickoff instead of only after writing-plans completes.
- **`planningModel` / `## planning model` are removed end to end** — settings store field and
  migration, `/flow-settings` output, `project-configuration.md`, `check-model-keys.sh`,
  `/flow-research`'s own resolve line. Nothing dispatches on a `PLANNING_MODEL` once planning is
  inline.
- **The inline-parent context ceiling (250k/400k/6-bundle) is removed.** A parent on the 1-hour
  TTL pays per-call reads, not rewrites, so it runs as large as the model's window allows; the
  `## Context ceiling — clear and resume` handoff block goes.
- **Read-discipline rules bind the parent** wherever it now does the conductor's own work: never
  `cat` a report (verdict section via `sed -n`), never read `final-review.diff`/a dispatch
  bundle/a whole fix diff, test/lint output through `tail`, each phase file read once per run then
  by section, change artifacts read once at load-context.
- **Housekeeping bundled in**: reword `pipeline.md`'s Progress visibility section (currently
  worded around the conductor's `## Stage` relay), lower `check-contract-budget.sh`'s byte-size
  rows to the post-cut measured sizes, delete the stale
  `docs/superpowers/research/kan-326-myflow-rework.md` fixture.

**Explicitly rejected:** a global `subagentPromptCacheTtl: "1h"` — not scoped to flow, it raises
the cache-write multiplier (1.25x → 2.0x) for every subagent in every project on the machine,
including one-shot subagents that never idle past 5 minutes.
<!-- measured: code.claude.com/docs/en/prompt-caching, read and confirmed by the operator during the research session -->
