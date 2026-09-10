# Remove flow's resumed subagents — research notes

Source: none

Explored via `/flow-research` on 2026-09-10 against this repository's own `spectre/` tree, the
phase files under `skills/flow/`, `skills/flow-research/SKILL.md`, the archived designs of
kan-372, kan-374, kan-377, kan-431, kan-472 and kan-483, and the sidechain transcripts under
`~/.claude/projects/-Users-tweety53-Projects-agents/<session>/subagents/agent-<agentId>.jsonl`
(usage rows per API call: `input_tokens`, `cache_creation_input_tokens`,
`cache_read_input_tokens`). The topic the operator brought, verbatim: "conductor burns a fuck ton
of tokens when resumed uncached. I don't like it. It slows me down because it burns weekly quota
too fast. Need to get rid of it and other similar subagents with work-resume-work logic." The
metric the operator watches is the subscription weekly usage meter and the terminal's live token
counter; the concrete symptom is a session running for about ten seconds and burning about 700k
tokens. Three investigate-then-ask rounds were run on the conductor, one more on planning, then
the operator fixed the decisions recorded in sections 4–8. The prior notes `flow-speedup.md`
(2026-09-04) and `flow-gymie-implementation-speedup.md` (2026-09-06) measured the conductor's
turn count and context; neither measured the cache-rewrite mechanism this note is about.

**Process note — a decision about `/flow-research` itself, recorded here for the change to
carry into `skills/flow-research/SKILL.md`:** a research session never parks an unresolved item
under **Open / undesigned** for a future reader to discover. Anything a session surfaces and
cannot settle from the tree is asked of the operator before the session ends — batched into the
same `AskUserQuestion` calls as every other question — and recorded with the operator's own
answer or their explicit deferral, in their words. The section may then hold only items the
operator chose to defer, each carrying that choice; an empty section is the ordinary outcome.
This note's own first draft violated the rule; its Open list was closed by asking.

## 1. What actually burns: cache rewrites of a large context

Every API call re-sends the whole conversation. When the prompt-cache prefix is warm the
re-send is billed as cache read (0.1× the input rate, and the same 0.1× weight on the weekly
meter); when it is cold the whole context is re-written as cache creation (1.25× on a 5-minute
TTL). A "10-second, 700k-token" call is exactly one cold call on a 700k context.

**kan-472's conductor** (`3c753c6b…/subagents/agent-a7561c34….jsonl`, sonnet, 12:39–20:10, 705
API calls, context 130k → 920k):

| Figure | Value |
|---|---|
| cache-creation tokens | 19.3M (≈ $48 at the seeded $2.50/Mtok) |
| cache-read tokens | 358M (≈ $72 at $0.20/Mtok) |
| calls that rewrote > 50k | 37, each 130k–920k |
| rewrites after a single Bash call that ran 5.8–9.4 min | 30 |
| rewrites after a coordinator (`SendMessage`) resume | 7 — four with a gap > 5 min, three with a gap of 0.5–1.6 min that missed anyway |

The 30 long Bash calls were the report waits. The contract (`skills/flow/implement.md`, **Turn
discipline**) states the wait as `for i in $(seq 1 48); do test -s <report> && break; sleep 5;
done` — 240 s, "under the prompt-cache TTL rather than the Bash tool's ten-minute cap" — but the
conductor wrote `seq 1 110` (550 s) every time, and nothing mechanical pins the bound. The
observed cache boundary is sharp: hits at gaps ≤ 4.9 min, misses at gaps ≥ 5.6 min.

**kan-473's conductor** (`0c8cd734…/agent-a0eeb2d5…`, 119 calls, 1.4 h): 7 rewrites, 1.14M
tokens, every one after a gap ≥ 5.7 min; four of them `[SYSTEM NOTIFICATION]` child-completion
wakeups — the conductor had ended its turn with a child in flight, which the contract forbids,
and was woken cold.

**The planner** (`0c8cd734…/agent-a2786924…`, fable, 30 calls, 4 resumes): one cold resume after
an 11.8-minute operator answer, 118k tokens rewritten (≈ $1.5 at fable's $12.50/Mtok cache
write); the other three resumes, under a minute apart, were warm.

**The parent session in the same runs**: 433 calls, 162M cache-read tokens at a context of
250k–650k, only 2 rewrites (gaps of 66 and 73 min). The parent survives idle gaps its subagents
do not — section 2 says why.

**The context size itself** is the second half. At 920k every call costs ~92k input-equivalents
even when warm. The conductor's context was 1.04M chars of tool results (Bash 616k — `cat` of
implementer reports, panel-fix reports, one 21k-char group bundle read back despite "never read
the bundle back"; Read 350k — the three phase files at 125k chars, the change artifacts at 79k,
`verify-and-handoff.md` read twice more) plus 671k chars of its own tool inputs. Six panel-fix
rounds in one run drove most of the Bash half.

## 2. Cache TTL facts (from `code.claude.com/docs/en/prompt-caching`, verified by the operator)

- Claude Code has two TTLs, 5 minutes and 1 hour, chosen per request bucket: the **main
  conversation** is on 1 hour by default on a subscription within included usage; the
  **"everything else" bucket** — every subagent, workflow agent, teammate, fork, compaction call
  and session-title request — is on 5 minutes. That is the split section 1 shows: the parent
  rewrote twice in 9.5 h, the conductor 37 times, the planner once.
- `subagentPromptCacheTtl: "1h"` / `CLAUDE_CODE_SUBAGENT_PROMPT_CACHE_TTL=1h` moves the whole
  "everything else" bucket to 1 hour at a 2.0× cache-write multiplier instead of 1.25×; reads
  stay 0.1×. The subagent-definition frontmatter field `experimental.cacheTtl` exists as a
  narrower, per-definition alternative. **Both rejected** — section 5.
- Weekly-meter weights are the same multipliers as billing (reads 0.1×, writes 1.25×/2.0×), so
  the meter gain of any fix has the same shape as its dollar gain.
- A `SendMessage` resume "can keep reading the prompt cache the original run warmed"
  (`sub-agents.md`). Whether a `<system-reminder>`/task-notification injection changes the cached
  prefix is not documented; the three short-gap misses in section 1 stay unexplained.
- Not documented: any way for an agent to wait on a child in-turn without either holding a Bash
  call open or ending its turn.

The operator chose not to tune any TTL and to make structural changes instead: no flow agent is
resumed across turns any more (sections 4 and 8).

## 3. Inventory: which flow subagents resume, which do not

| Role | Dispatched by | Resumed? | Observed cold cost | Decision |
|---|---|---|---|---|
| conductor (`implement.md` **Dispatch the conductor**) | parent | yes — every `## Question`, every `## Stage` relay (`continue`), every wakeup | 19.3M tokens on kan-472, 1.14M on kan-473 | removed (section 4) |
| planner (`brainstorm.md` **Dispatch the planner**; `planner-fix-<n>` on fix runs) | parent | yes — 3–5 resumes per run, one per operator answer or return | ~118k on fable per cold resume | removed (section 8) |
| `/flow-research` researcher (`skills/flow-research/SKILL.md`) | parent | yes — every operator answer | same shape as the planner (this session included) | removed (section 8) |
| implementer, panel-fix, panel bundle, visual verifier, self-review | conductor (parent inline) | no — one-shot, report file, done | one cache write at dispatch (50–100k, the duplicated baseline KAN-372 targets) | unchanged |

After both changes, every remaining flow subagent is one-shot: dispatched fresh, writes a report,
ends. The only long-lived context in a run is the main session's, on the 1-hour TTL.

## 4. Decision: the parent orchestrates — no conductor, the four dispatch rows unchanged

**Hybrid.** The conductor role is deleted; the parent runs `skills/flow/implement.md` sections
1, 2 and 4, `skills/flow/review-panel.md` and `skills/flow/verify-and-handoff.md` itself — every
guard, gather, `flow record`, `flow stage`, worktree add/remove, report read and diff walk that
today is "the conductor's own Bash and Read work" — and dispatches the four rows of **Dispatch
sites — the conductor's closed list** itself: implementer per group, panel bundle, panel-fix per
round, visual verifier per worktree. Those four stay exactly as they are — one-shot, fresh per
task or round, never resumed, on the models and paragraphs they carry today. The operator's words:
"Implementer, fixer, review subagents should stay on SDD. Just the conductor should go. They don't
do that work-resume-work thing and get spawned fresh for their task."

The existing `inline` execution mode (kan-472, `inline-parent-implements-panel-dispatches`) —
parent implements small/regular-class plans itself, dispatching only panel bundles and the
verifier — is untouched; `sdd` now means "the parent dispatches implementers" rather than "a
conductor does". Waves, groups, the `Task-Id:` guard, the panel's bundled dispatch and the
panel-fix single-dispatch rule all survive with "the conductor" read as "the parent".

Why this and not a smaller fix: the parent is on the 1-hour TTL already (section 2), so every
wait, operator answer and child wakeup that cold-started the conductor is warm in the parent for
up to an hour; the nine `## Stage` relay turns and the `continue` resumes disappear; there is one
large context per run instead of two (conductor 920k beside a parent at 650k on kan-472); the
dispatch prompt's own ~46k fixed load and the `Model:` handshake for the conductor go. kan-377's
reason for the conductor — moving 42M context-tokens out of the parent — was measured before
kan-431 batched the turns and before the rewrite mechanism was understood; what the parent takes
back is the batched version.

What is given up: `DEFAULT_MODEL` for orchestration independent of the session model
(kan-377 `conductor-model-default`) — the parent's own model now prices every orchestration call;
the closed-list / NO DELEGATION machinery of KAN-449, kan-483 and kan-484 shrinks to the parent's
four rows and the leaves' NO DELEGATION paragraphs.

**Mechanics that change with the role:**

- `implement.md` **Dispatch the conductor** and **The parent backstop** go; **Inline — the parent
  implements** stays; the `## Question` / `## Stage` / `## Handoff` relay contract goes — the
  parent asks through `AskUserQuestion` directly, updates the harness task list at each `flow stage
  end` directly (`pipeline.md` **Progress visibility**, "run inside a conductor subagent" reworded),
  prints the `IN_PROGRESS` handoff directly.
- The conductor's `flow record dispatch begin/end -role conductor` pair goes; `conductor` stays in
  `stats/cmd/flow/record.go`'s role list for history.
- **Turn discipline** (`implement.md`) binds the parent unchanged: one Bash call per wait, bounded
  under the TTL — now the hour, so `seq 1 48` × 5 s stays as the loop and the ten-minute Bash cap
  is the only ceiling that matters.
- "Never end a turn with a child in flight" binds the parent unchanged; a parent turn that ends is
  the operator's turn, and the parent's cache survives it.
- `skills/flow/SKILL.md`'s guardrail "Never run `implement.md` sections 1, 2 or 4,
  `review-panel.md` or `verify-and-handoff.md` in the parent session — the conductor runs them" is
  inverted: the parent runs them.
- `scripts/check-dispatch-paragraphs.sh` pins six sites; the conductor and planner sites are
  removed (section 8), four remain. `scripts/test-check-dispatch-paragraphs.sh` follows.
  `check-panel-fix-single-dispatch.sh` is unchanged (one panel-fix per round, whoever dispatches).
  `check-task-commit-fields.py`, `prepare-archive-branch.sh`, `lib/post-mutation-check.sh` mention
  the conductor in prose only.
- `skills/flow-contracts/model-policy.md` and `project-configuration.md` each carry one conductor
  mention to reword; `brainstorm.md` one.
- `scripts/check-contract-budget.sh` rows for `implement.md`, `review-panel.md`,
  `verify-and-handoff.md` and `SKILL.md` are lowered to the measured post-cut sizes, never raised.

## 5. Rejected: a cache-TTL setting for the planner and researcher

Offered during the session as the fix for the planner's and researcher's cold resumes, then
rejected by the operator after reading the docs: `subagentPromptCacheTtl` is not per role — it
moves every request in the "everything else" bucket, in every project on the machine, to the
2.0× write multiplier, including one-shot agents that never idle past five minutes (Explore
searches, panel bundles, verifiers, one-off general-purpose lookups) — a pure cost increase for
them with no gap to survive. The per-definition `experimental.cacheTtl` frontmatter field would
narrow it, but the operator declined TTL tuning altogether in favour of removing the resumed
roles (section 8). Recorded here so the next reader does not re-propose it.

## 6. Decision: remove the context ceiling

`implement.md` **Inline — the parent implements** stops an inline run under 250k tokens before a
bundle or 400k before the panel (or after the sixth bundle without a budget figure) and prints
`## Context ceiling — clear and resume` (`skills/flow-contracts/handoff-blocks.md`). The operator
removes it: a parent on the 1-hour TTL pays per-call reads, not rewrites, and is let run as large
as the model's window allows. kan-472's decision `ceiling-250k-400k-6-bundles` is superseded; the
handoff block, its `handoff-blocks.md` section and the `<total_tokens>`-reading paragraph go. The
run's only size controls are then the read-discipline rules of section 7 and the model's window.

## 7. Decision: read-discipline rules for the parent, in the same change

The conductor's 920k context was mostly reads it did not need whole. The parent, doing the same
work, follows rules stated once in `implement.md` (the section **Turn discipline** already binds
the reader) and cited from the panel and verify files:

- **Never `cat` a report.** An implementer, panel-fix, panel or verifier report is read for its
  verdict section only — `sed -n '/^## Verdict/,/^## /p' <report>` or the equivalent for that
  report's shape — never the whole file. The report file's existence (`test -s`) is the wait
  condition; its body is read once, narrowly.
- **Never read `final-review.diff`, a dispatch-context bundle, or a panel-fix diff in the parent.**
  Slots read the diff; the parent walks a fix's hunks through `git diff --stat` and the specific
  hunks a finding names, not the whole diff. "Never read the bundle back" (already in §4) is
  extended to every generated file the parent produces for a child.
- **Targeted test output through `tail`** — already the rule for implementers (TARGETED TESTS);
  it now binds the parent's own `## lint`/`## test` runs in `flow.verify` and the full-suite run
  after a shared wave: `| tail -20`, the failing block reproduced from the log file on a failure.
- **Phase files read once per run.** `implement.md`, `review-panel.md`, `verify-and-handoff.md`
  are each read once, in full, at the start of the stage that needs them; a later need is served
  by `grep -n` for the heading plus `sed -n` for that section, never a second full Read. kan-472's
  conductor read `verify-and-handoff.md` three times (55k chars).
- **Change artifacts read once**, `proposal.md`/`design.md`/`tasks.md` at `flow.load-context`;
  `tasks.md` re-read only through `spectre list --json` and `flow tasks tick` output afterwards.

## 8. Decision: planning is done by the running session — no planner, no researcher subagent

Two entry points, neither dispatching a planning subagent; whichever session runs does the
thinking on its own model.

1. **`/flow-research`, any argument shape, entirely in the current session.** With a Jira key,
   with a change name, or bare — including the invocation that produced this note — the session
   itself reads the tree and the code, runs the investigate-then-ask rounds and the convergence
   check exactly as `skills/flow-research/SKILL.md` states, on its own model, dispatching no
   subagent, ever. The operator's words: "No subagents at all!!!" Only the dispatch mechanic
   changes: `skills/flow-research/SKILL.md`'s **The research subagent** section goes — with its
   "the parent session does not do the thinking itself" sentence, the `PLANNING_MODEL` resolve,
   the relay-contract and `Model:` handshake citations, and the "no `flow record dispatch`"
   note — and every "you" in the skill addresses the session; the stance, guardrails and
   capture sections are otherwise unchanged. **Questions are the session's own direct
   `AskUserQuestion` calls, batched as the skill already states** — every pending question whose
   wording does not depend on another pending answer in one call, up to four per call, a
   dependent question waiting for the next turn — executed directly instead of ending a turn with
   `## Question` blocks for a parent to re-ask; the relay's "prose before the question is relayed
   too" rule is moot, the session prints its own summary before the call. **Capturing the
   Outcome** is unchanged — an existing
   change in scope (a key or name that resolves to `<project>/spectre/changes/<id>/`, or a
   conversation already about one) gets the offer to write into that change's `design.md` in the
   fixed structure; no change yet gets the staging-note fallback at
   `<project>/docs/superpowers/research/<topic-slug>.md`, which `/flow`'s seed rule already reads.
   The operator's intended use: run it ahead of time on the ticket, then `/flow KAN-XXX`
   separately for implementation, seeded from what it wrote.
2. **`/flow KAN-XXX` finding no prior `/flow-research` output for that change** runs
   `brainstorm-planner.md`'s sections B, C and D itself — the checklist, convergence, `spectre
   new`, the three artifacts, writing-plans and the Decide step — in the current session, then
   continues into implementation as today. `brainstorm.md` **Dispatch the planner**, its relay
   contract, `Model:` handshake, `opus` fallback and `planner`/`planner-opus`/`planner-<model>`
   record keys go; `flow.document-fix`'s `planner-fix-<n>` dispatch on a fix run is the same
   section run by the parent. `brainstorm-planner.md`'s "you" is the session; its "This section's
   stage marks are run by the parent, not the planner" split collapses — one actor.

What this supersedes: kan-374's `planner-subagent` decision ("one planner subagent on
`PLANNING_MODEL`, parent relays — the only mechanism that sets a model from inside a run") and
with it the reason `planningModel` (settings store, `/flow-settings`) and the `## planning model`
project key (`project-configuration.md`) exist — planning runs on the session's model, so both
are removed in this change (breakdown item **Remove `planningModel` and `## planning model`**). kan-474's planner batching
(up to four `## Question` blocks per turn) survives as the session's own AskUserQuestion batching,
which the research skill already states. The `planner` role stays in `record.go` for history.

Why: the planner and the researcher are the two remaining work-resume-work agents — resumed once
per operator answer, cold whenever the answer takes over five minutes — and after section 4 they
would be the only ones. The session that runs them is on the 1-hour TTL and, on a brainstorm,
already holds the operator's conversation; a relay through a second context bought only a model
choice the operator no longer wants.

### The strict artifact path — decided

`/flow` must answer "do research notes exist for this ticket" with one file test and no
inference. The decision, keyed off what `/flow` already resolves first (the Jira key, per
`brainstorm.md` **A**):

| `/flow-research` situation | Writes to | `/flow` reads it how |
|---|---|---|
| A Jira key is named (or resolved from the conversation) and **no change exists yet** | `<project>/docs/superpowers/research/<jira-key-lowercased>.md` in the main checkout — e.g. `docs/superpowers/research/kan-490.md` | `test -f` on that exact path at kickoff; seeds **B**, and **C** folds it into `design.md` and deletes it in the artifact commit, so it ends inside the change's own artifacts in the worktree |
| A change already exists (`<project>/spectre/changes/<id>/`, in the worktree once one exists) — named, or the conversation is about it | that change's `design.md`, in the fixed section structure | `/flow` reads `design.md` at `flow.load-context` as today; a `STARTED` resume reads it before **B** |
| No key and no change — a bare thinking-partner session | `<project>/docs/superpowers/research/<topic-slug>.md` | seeded only when `<name>.md` equals the slug exactly, as today — best effort, no guarantee; this is the one path the operator does not run `/flow` against by key |

What this changes in the tree: `brainstorm-planner.md` **Seed from a staged research note**
loses the `<jira-key-lowercased>-*.md` glob fallback and its "more than one match" prompt — the
exact filename is the whole rule, keyed or slug — and `skills/flow-research/SKILL.md`
**Staging a Note** states the keyed filename as mandatory when a key is known, never a
descriptive suffix. The one file the glob existed for, `docs/superpowers/research/kan-326-myflow-rework.md`,
is a pre-mechanization fixture of an archived change and is renamed or left un-seeded. No
worktree is created early: the note's content reaches the worktree through **C**'s adoption, so
`/flow-research` needs neither `spectre new` nor a branch — the guardrails stand.

### Announce the resolved dynamic choices at kickoff — decided

Two display points, chosen by the seed check — never both, nothing moved from where `/flow`
already shows a decision today.

**The seed line always prints at start**, as the second line after `skills/flow/SKILL.md`'s
"Using flow for change `<name>`." — it is what selects the path:

```
Using flow for change `<name>`.
  research seed: found docs/superpowers/research/<key>.md — seeding brainstorm
               | none — planning inline in this session
```

**Seeded path (a prior `/flow-research` produced the keyed note):** the full block follows the
seed line, before any stage past kickoff runs —

```
  planning:  inline, this session (<model>)
  toggles:   execution mode <default|dynamic> · implementer model <…> · review panel <…>
  models:    default <DEFAULT_MODEL> · reviewers <REVIEWERS>
  decision:  <class / execution / implementer / panel / groups from the recorded decision>
             | not yet decided — the `## Decision` block follows writing-plans
```

The `decision:` line reads `flow record decisions -change <name>`'s newest row on a resumed
`STARTED` run or a fix run; on a seeded creating run **D**'s Decide step has not run yet, so the
line says so and the block is completed by the existing `## Decision` print the moment **D**
records it.

**No-seed path (inline planning fallback):** nothing beyond the seed line at start. The
choices appear where `/flow` already shows them today — the verbatim `## Decision` block printed
after `flow.writing-plans`, once **D** returns (`brainstorm.md`, "print the planner's `##
Decision` block verbatim", now the session's own) — extended with the `planning:`, `toggles:`
and `models:` lines above so that one print carries every per-run choice. That point is neither
moved nor duplicated.

Both prints are the parent's own output, never a stage mark.

## 9. Closed by asking — the operator's answers on the remaining items

- **Short-gap `SendMessage` cache misses** (kan-472 conductor 14:01, 15:29, 18:48, gaps of
  0.5–1.6 min, cache read 25k = system prompt only): operator — "Moot, drop it." Not
  root-caused; no flow agent is resumed once this change lands.
- **The session model prices orchestration and planning** (kan-377 had the conductor on
  `DEFAULT_MODEL`, kan-374 had planning on fable; no in-run switch exists for the main session):
  operator — "Accepted, no option preserved." No hint line, no mechanism.
- **Size bound for the inline-implementing parent** (small/regular class, writes the code itself,
  gymie's 6–19-task plans): operator — "Truly unlimited, same answer as the orchestrating
  parent; no distinction." The model's window is the only bound on either shape; section 6's
  ceiling removal covers both.
- **`docs/superpowers/research/kan-326-myflow-rework.md`** (the one file the removed glob
  fallback matched; its change is archived): operator — "Delete it." Deleted in this change, in
  the same task that removes the glob from **Seed from a staged research note**.
- **`pipeline.md` **Progress visibility**** (its implementation-branch paragraph is worded around
  the conductor's `## Stage` relay): operator — "reword it in this change." Same task as the
  conductor removal; the mechanism (one task-list entry per stage, marked at each `flow stage
  end`) is unchanged.
- **`scripts/check-contract-budget.sh` rows** (byte-size budgets per contract file, not cost
  tracking): operator — "lower them in this change." The task that cuts each of `implement.md`,
  `review-panel.md`, `verify-and-handoff.md`, `SKILL.md`, `brainstorm.md` and
  `skills/flow-research/SKILL.md` measures the result with `wc -c` and lowers that file's row in
  the same commit, so the budget cannot silently re-grow.

## Step-by-step breakdown

### Conductor removal

**What:** Delete the conductor role — `implement.md` **Dispatch the conductor**, **The parent
backstop**, the `## Question`/`## Stage`/`## Handoff` relay contract, the conductor handshake
and its `-role conductor` dispatch record — so no agent in the implementation half is resumed
across turns.
**Why:** The conductor is the one flow agent whose resumes are frequent and on a context of
hundreds of thousands of tokens; on the 5-minute subagent TTL every wait over five minutes,
operator answer and child wakeup rewrote 130k–920k tokens (19.3M on kan-472, 37 rewrites).
**Uses:** `skills/flow/implement.md`, `skills/flow/SKILL.md` (guardrail inversion),
`skills/flow-contracts/pipeline.md` **Progress visibility**, `skills/flow/brainstorm.md`,
`skills/flow-contracts/model-policy.md`, `skills/flow-contracts/project-configuration.md`,
`scripts/check-dispatch-paragraphs.sh`, `scripts/test-check-dispatch-paragraphs.sh`,
`scripts/check-contract-budget.sh`, `stats/cmd/flow/record.go` (role kept for history).

### The parent takes the conductor's own Bash and Read work

**What:** The parent runs `flow.load-context` through `flow.write-in-progress` itself — every
guard, gather, record, stage mark, worktree add/remove, wait loop, report read and diff walk —
asking the operator through `AskUserQuestion` directly and updating the task list per stage.
**Why:** The parent is on the 1-hour cache TTL on a subscription (2 rewrites in 9.5 h against
the conductor's 37), so the same waits and answers return warm; the relay turns and one of two
large contexts disappear; kan-431's batched turn shape is what the parent takes back, not
kan-377's pre-batch measurement.
**Uses:** `skills/flow/implement.md` §1, §2, §4 and **Turn discipline** (the 240 s wait loop
unchanged), `skills/flow/review-panel.md`, `skills/flow/verify-and-handoff.md`,
`scripts/check-task-commit-fields.sh`, `scripts/gather-dispatch-context.sh`,
`scripts/plan-dispatch-bundles.sh`, `flow record`, `flow stage`, `flow tasks tick`.

### Implementer, panel-fix, panel-bundle and verifier dispatch preserved

**What:** The four rows of **Dispatch sites — the conductor's closed list** become the parent's
own dispatch sites, unchanged in key shape, model, paragraphs (FLOW, TDD sub-skills, TARGETED
TESTS, REPORT FILE, NO DELEGATION, TOOLS, MODEL HANDSHAKE) and one-per-worktree / one-per-round
rules; waves and groups stay.
**Why:** These agents are one-shot — dispatched fresh per task or round, report to a file, never
resumed — so they carry none of the work-resume-work cost; the operator confirmed they stay on
the SDD dispatch pattern and only the middle layer goes.
**Uses:** `skills/flow/implement.md` §4 (implementer dispatch, waves), `skills/flow/review-panel.md`
**Bundled dispatch** and the fix step, `skills/flow/verify-and-handoff.md` **The verifier
dispatch**, `scripts/check-dispatch-paragraphs.sh` (four sites),
`scripts/check-panel-fix-single-dispatch.sh` (unchanged).

### Planning inline: `/flow-research` and `/flow`'s brainstorm run in the session itself

**What:** `/flow-research`, with any argument or none, does its research in the current session
and captures as the skill already states — an existing change's `design.md`, else the staging
note under `docs/superpowers/research/`; `/flow KAN-XXX` with no seed runs
`brainstorm-planner.md` B, C and D itself. No planner, `planner-fix-<n>` or researcher subagent
in any path; no relay, handshake or `opus` fallback for planning. Questions are direct, batched
`AskUserQuestion` calls (up to four independent questions per call) from the session itself.
**Why:** The planner and researcher were the last two work-resume-work agents — cold on every
operator answer over five minutes (118k on fable per miss) — and their only reason to be
subagents was choosing `PLANNING_MODEL`, which the operator no longer wants; the session is on the
1-hour TTL and already holds the conversation. A global cache-TTL setting was rejected for its
blast radius (section 5).
**Uses:** `skills/flow-research/SKILL.md` (**The research subagent** removed; **Capturing the
Outcome** unchanged), `skills/flow/brainstorm.md` **Dispatch the planner** (removed),
`skills/flow/brainstorm-planner.md` B/C/D (the session's own), `skills/flow/implement.md` §3
(`flow.document-fix` run by the parent), `skills/flow-settings/SKILL.md` and
`skills/flow-contracts/project-configuration.md` (`planningModel` / `## planning model`, dead),
`scripts/check-dispatch-paragraphs.sh` (planner site removed), `stats/cmd/flow/record.go`
(`planner` role kept for history).

### Strict research-artifact path keyed by Jira key

**What:** `/flow-research` writes to exactly one of three deterministic destinations —
`<project>/docs/superpowers/research/<jira-key-lowercased>.md` when a key is known and no change
exists, the existing change's `design.md` when one exists, `<topic-slug>.md` only for a bare
keyless session; `/flow` checks the keyed path with one `test -f`; the `<key>-*.md` glob and its
disambiguation prompt are removed.
**Why:** The operator requires `/flow KAN-XXX` to know whether notes exist without inference; the
keyed exact-filename rule already exists in the seed logic and needs no worktree before `/flow`
creates one — **C**'s adoption moves the content into the change's artifacts and deletes the note.
**Uses:** `skills/flow-research/SKILL.md` **Staging a Note** / **Capturing the Outcome**,
`skills/flow/brainstorm-planner.md` **Seed from a staged research note** and **C** ("Delete the
adopted staging note"), `skills/flow-contracts/jira-integration.md` **Change naming**,
`docs/superpowers/research/kan-326-myflow-rework.md` (fixture to rename or leave).

### Announce resolved dynamic choices at kickoff

**What:** The seed result (keyed note found → seeding, or none → inline planning) prints right
after "Using flow for change `<name>`." on every run. On the seeded path the full block —
planning mode and model, the three toggles, `DEFAULT_MODEL`/`REVIEWERS`, the recorded decision
or "not yet decided" — prints there too, before any stage past kickoff. On the no-seed path the
same lines join the existing post-`writing-plans` `## Decision` print, which is neither moved
nor duplicated.
**Why:** The operator wants every per-run choice this redesign introduces visible, not read out
of a stage mark: up front when research already happened, and at the point `/flow` already
shows its decision when planning had to run first.
**Uses:** `skills/flow/SKILL.md` (**Announce at start**, **Model resolution**),
`skills/flow/brainstorm.md` **A** and the `## Decision` print, `flow record decisions -change <name>`,
`test -f docs/superpowers/research/<key>.md`.

### Remove `planningModel` and `## planning model`

**What:** Delete the planning-model setting end to end: `skills/flow/SKILL.md`'s `PLANNING_MODEL`
resolution block (the `planningModel` read, the `## planning model` project override, the
`fable` default); `flow settings set -planning-model`, the `PlanningModel` field of the settings
row and its `ValidModels` check, and the `planning model:` line in `/flow-settings`'s output; the
`## planning model` row and paragraph in `project-configuration.md`; the key's case in
`scripts/check-model-keys.sh`; `/flow-research`'s "resolves `PLANNING_MODEL`" sentence. The
store's `planning_model` column is dropped by a migration or left nullable and unread —
implementer's call at the schema. `flow settings get` no longer emits `planningModel`.
**Why:** Nothing dispatches on `PLANNING_MODEL` once planning is inline (section 8); the setting
existed only to pick the planner subagent's model (kan-374 `planning-model-setting`), and a
setting no run reads is the kind of dead configurability the operator wants removed now, not left
as later cleanup.
**Uses:** `skills/flow/SKILL.md` **Model resolution**, `skills/flow-settings/SKILL.md`,
`skills/flow-contracts/project-configuration.md`, `skills/flow-research/SKILL.md`,
`stats/cmd/flow/settings.go`, `stats/internal/store/settings.go` (+ a migration under
`stats/internal/store/migrations/`), `stats/internal/api` settings DTO, `scripts/check-model-keys.sh`
(`self review model` case stays), `.flow/project.md` (no `## planning model` body here today).

### Context ceiling removal

**What:** Delete the 250k/400k/six-bundle stop from **Inline — the parent implements** and the
`## Context ceiling — clear and resume` handoff block.
**Why:** The ceiling guarded a parent against rewriting a huge context; on the 1-hour TTL a
large parent context costs 0.1× per call and no rewrite, so the operator lets a run use the
model's window instead of paying a `/clear` and a re-read of every phase file mid-change.
**Uses:** `skills/flow/implement.md` **Inline — the parent implements**,
`skills/flow-contracts/handoff-blocks.md` **Context ceiling — clear and resume**, kan-472's
`design.md` decision `ceiling-250k-400k-6-bundles` (superseded, never deleted).

### Read-discipline rules for the parent

**What:** Five rules, stated once in `implement.md` and cited elsewhere: never `cat` a report
(verdict section via `sed -n`); never read `final-review.diff`, a dispatch bundle or a whole fix
diff in the parent; test and lint output through `tail`; each phase file read in full once per
run, then by section; change artifacts read once at load-context.
**Why:** kan-472's conductor reached 920k tokens on 1.04M chars of tool results — reports and
bundles `cat`-ed whole, phase files re-read — and at that size even a warm call costs ~92k
input-equivalents; with the ceiling gone these rules are the run's only size control.
**Uses:** `skills/flow/implement.md` (**Turn discipline**, the TARGETED TESTS paragraph's `tail`
rule), `skills/flow/review-panel.md` (report reads, the fix-diff walk),
`skills/flow/verify-and-handoff.md` (`## lint`/`## test` output), `sed`, `tail`, `git diff --stat`.
