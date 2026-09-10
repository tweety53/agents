## Context

Dispatch usage reaches the store only through the harvester's second attribution pass
(`DispatchAttributor`): identity by `agentId` first, then the interval fallback, and — since
kan-212 — ambiguity between candidates is refused, not guessed. The pass reads sidechain records
parsed out of transcript files the watcher already walks, and `discoverTranscripts`
(watcher.go) has always walked `subagents/agent-*.jsonl` alongside top-level session files. Each
agent file is one subagent's whole conversation — every line carries that file's own `agentId`,
`isSidechain: true`, and a per-message `usage` object, so the existing
`ParseAssistantRecords`/`ReadNewRecords` machinery reads it unchanged.

`Watcher.attributeDispatches` is called per file batch and already holds the batch's `path`, so
a batch's source file is decidable at the exact point where the inference pass runs today.

The failure this change fixes is measured, not hypothetical: in the kan-377 run, every dispatch
row whose `agent_id` belongs to a resumed implementer (`a2ad390ad5a6c1507`, `aa576359cc4457756`,
`a9bda1ce0c922e083` — two rows each) is stamped unattributed with reason
`matched more than one dispatch`, because the identity pass matches two rows and the timestamp
narrowing finds no containing window (row windows are hand-typed; kan-212 F3 measured them
rounded to the minute). The same file whose records fail window attribution carries the whole
answer: it is that agent's usage, and only the row-grain split among its own resumes is ever in
question.

## Decisions

### Agent-file batches credit rows directly by the file's agentId

**ID:** agent-file-direct-credit
**Status:** active
**Chosen:** a batch read from a path under a `subagents/` directory never enters
`DispatchAttributor`; its records are summed per target row resolved by the file's agentId, and
merged through the existing `MergeDispatchMetrics` — the source names its dispatch, so no
window is consulted.
**Considered:** narrowing the inference pass among same-id rows by containment or nearest
window — rejected, hand-typed row windows are the measured cause of the failure, and a narrowed
tie between two resumes of one agent reintroduces it one layer down; fixing the skills to record
distinct agentIds per resume — rejected, it cannot repair already-recorded rows and adds a rule
every dispatch site must remember forever.

### A resumed agent's rows split the file's usage by row start order

**ID:** resumed-split-by-started-at
**Status:** active
**Chosen:** with rows ordered by `started_at`, a record joins the latest row whose `started_at`
is at or before the record's timestamp; a record earlier than the first row's start joins the
first row. Every record lands in exactly one row, no row is skipped, and the inference is
bounded to choosing among resumes of the *same* agent — it can never move spend across agents,
which is the error the refuse-to-guess rule exists to prevent.
**Considered:** crediting the agent-level total to the first row only — rejected, it silently
moves later tasks' implementer cost onto an earlier task's row; segmenting the file at resume
boundaries in the transcript content — rejected, user-role lines also carry tool results, so a
content-based boundary is a heuristic pretending to be data.

### Top-level transcript batches keep the two-pass inference

**ID:** inference-stays-for-embedded-sidechains
**Status:** active
**Chosen:** records arriving inside a top-level session transcript (embedded sidechains, the
shape Cursor/Codex and pre-agent-file Claude transcripts produce) continue through
`DispatchAttributor` unchanged, ambiguity stamp included.
**Considered:** routing every batch through the new direct credit — impossible, an embedded
sidechain line is not a file per agent, so there is no per-dispatch source to name.

### Zero-row agent files are skipped silently

**ID:** unrecorded-agents-skipped
**Status:** active
**Chosen:** an agent file whose agentId matches no dispatch row contributes nothing at the
dispatch grain and reports nothing — the same silence `Attributor.Attribute` already applies to
a record matching no window; the stage grain still counts its tokens.
**Considered:** stamping or logging each unrecorded agent — rejected for this change, there is
no row to stamp and no caller watching that log; a fork's subagents are today outside the
dispatch protocol entirely.

### No historical backfill

**ID:** no-historical-backfill
**Status:** active
**Chosen:** rows already stamped unattributed stay as they are; this change corrects future
runs only, and a backfill (locate each stale row's agent file, re-read it from zero, re-credit)
is a follow-up candidate with its own file-discovery surface.
**Considered:** shipping the backfill here — rejected for scope: harvest offsets for every
historical agent file already sit at EOF, so a repair pass is new machinery (per-file discovery
across project transcript roots) serving no green test this plan carries.

## Open questions

### How zcode subagent sessions could bind to their dispatch rows

**ID:** zcode-subagent-binding
**Status:** open
**Why it is open:** zcode rollout files are one per subagent *session*
(`model-io-sess_<uuid>.jsonl`), carry no agentId and no sidechain flag, and no recorded data
connects a subagent's session id to the dispatch row its launch created — the parent never
sees it.
**What it affects:** dispatch-grain cost for zcode-dispatched subagents stays unattributed;
closing this needs either the launch result or the dispatch marks to carry the subagent's
session id, which is a skills-and-CLI change, not a harvester one.
