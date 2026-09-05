# flow: merge brainstorm gate, self-review key, and self-review filing prompt — design

**Change:** `kan-435-flow-merge-brainstorm-gate-self-review-key-and`
**Jira:** KAN-435
**Date:** 2026-09-05
**Source:** `docs/superpowers/research/flow-speedup.md` — attention merges 1, 2a, 2b (§7 round two,
decided in §8).

## Context

Three prompt reductions, prose only. Every flow edited already exists; no capability spec under
`spectre/specs/` covers these stages, so no spec edit is planned. No guard or test parses either
prompt; stage keys are unchanged. `check-self-review-report.sh` parses the report, whose shape is
unchanged. No new script: `project-get.sh` reads the new key as it reads every other.

## 1. Merged brainstorm gate

### `skills/flow/brainstorm-planner.md` — **Convergence**

The confirm prompt becomes:

> **That is everything I have settled. Anything still unclear before I move on?**
> - **Nothing unclear — approve the design and move on** *(recommended)*
> - **Another round — I have something** *(default — anything short of an explicit "approve and
>   move on" is treated as this)*
> - **Revise — I have a change to the design**

"End the stage only on an explicit choice of **move on**" becomes "…of **approve the design and
move on**". The safe-default/recommended split is unchanged — this remains the one prompt in the
file where they differ, and the `⚠ another round — no explicit answer` line still prints when the
default fires.

*Revise* is a round: it counts toward the round-3 offer exactly as *Another round* does, and its
difference is what the planner's next turn opens with — the changed design section(s) are
re-presented before the next confirm, in place of new questions.

The paragraph "The convergence loop's own exit — an explicit **move on** at the confirm above — is
what closes the checklist itself; the **design approval** the HARD GATE requires is a separate,
later act by the same operator and gets its own mark:" is replaced by: the explicit **approve the
design and move on** answer is at once the convergence exit that closes the checklist and the
design approval the HARD GATE requires. The parent still marks `flow.brainstorm` end, then
`flow.design-approval` begin and end, around that one relayed answer — the four-line block that
follows stays, its comment line reading "the operator's approve-and-move-on answer, relayed — this
is the HARD GATE above".

### `skills/flow/brainstorm-planner.md` — **The checklist**

The HARD GATE bullet keeps "do not run `spectre new` until the user approves the design" and adds:
approval is the merged confirm's first option under **Convergence**; no separate approval question
is asked.

### `skills/flow/brainstorm.md` — **The relay**

"Section B's convergence confirm, its third-round offer, and the HARD GATE design-approval question
are relayed the same way" becomes "Section B's merged convergence-and-approval confirm and its
third-round offer are relayed the same way". The mark block's comment line becomes "the operator's
approve-and-move-on answer, relayed through the planner's merged confirm — the HARD GATE".

### `skills/flow-contracts/pipeline.md` — **Stage exit**

One clause: the same explicit answer may both close the checklist and grant the design approval,
as **Convergence** (`skills/flow/brainstorm-planner.md`) defines. Substance unchanged.

### `skills/flow-contracts/pipeline-rationale.md` — **Stage exit**

One sentence with the measurement: `flow.design-approval`'s median wall clock was 0 s over 49 runs
in this repository's dev stats store on 2026-09-04 — the approval was a reflex seconds after the
confirm — which is why one answer now serves as both.

## 2. `## self review` project key

### `skills/flow-contracts/project-configuration.md`

Optional-keys table, a new row directly after `## default landing route`:

> `## self review` — Optional. One of the literal bodies `run` or `skip` — this section holds that
> value and nothing else, never free-form prose, matching `## default landing route`'s own
> single-line-literal shape. Resolved by run 2's step 9 (`skills/flow/archive.md`) before its skip
> prompt: `skip` skips self-review without asking, `run` runs it without asking, absent asks as
> today. A body matching neither literal exactly is reported by name and dropped, resolving as if
> the key were absent.

The matching paragraph after the table gains `## self review` beside `## planning model`: matched
exactly as `## default landing route`'s body is — two literals in place of three, same
report-by-name-and-drop. No change to `check-model-keys.sh`.

### `skills/flow/archive.md` step 9

Before "The skip prompt fires first": run `project-get.sh <main-checkout> "self review"` (exit 1:
absent) and match the body against `run` / `skip` byte-for-byte after trimming; a body matching
neither is reported by name and dropped. `skip` ends step 9 here with the handoff's `Self-review`
line reading `skipped — project default`. `run` proceeds to the reasoning pass with no prompt.
Absent: the skip prompt fires as today.

The handoff template line becomes `**Self-review:** <path> (rating: <n>/5) | skipped | skipped —
project default`. `skills/flow-contracts/handoff-blocks.md`'s one-line shape statement of that
field is updated to match.

The guardrail "Never ask the self-review skip prompt, a per-angle filing ask, or the rating
question before `FINISHED`" becomes "…the skip prompt or the filing-and-rating prompt…"; the
key's resolution happens inside step 9, after `FINISHED`, like the prompt it replaces.

### `skills/flow-contracts/finish-contract-run2.md` step 9

"It is skippable per run, with running it the default" gains: a project's `## self review` key
(**Project configuration**, `skills/flow-contracts/project-configuration.md`) decides without
asking when present and valid; the per-run prompt is the absent case.

### `.flow/project.md`

A new section directly after `## default landing route`:

```markdown verified:authored in-tree for this change
## self review

`skip`
```

with one sentence of reason: the report series ended at kan-380 — the six changes after it all
answered "No" to a prompt that fires after `FINISHED`, when the operator has walked away — and 30
reports yielded 9 Jira tickets; the key ratifies that and ends the series. `run` is the value to
set when the series is wanted back.

## 3. One filing-and-rating prompt

### `skills/flow-contracts/finish-contract-run2.md` step 9 (canonical)

"The filing ask is **one multi-select prompt per angle**…" becomes: the filing ask and the rating
are **one `AskUserQuestion` call**. Findings from every angle fill up to three multi-select
questions, each carrying at most three findings — every option prefixed with its angle's label —
plus **None — file nothing** as that question's `(default, recommended)` option; the rating is the
call's last question. More than nine findings roll the overflow into one further call of the same
shape, without the rating. Shape per **Operator prompts**
(`skills/flow-contracts/operator-prompts.md`). The label-on-filed-issue sentence, "Every finding is
explained in the message body before any prompt fires", and the report's per-angle none-marker are
unchanged. "the operator's rating and the per-angle filing prompts still run in that session"
becomes "the filing-and-rating prompt still runs in that session".

### `skills/flow/archive.md` step 9

The per-angle prompt block is replaced by the executing wording:

> **File any of these findings as Jira issues?** *(one question per three findings, each option
> `<label>: <finding>`)*
> - **`<label>`: <finding 1>**
> - **`<label>`: <finding 2>**
> - **`<label>`: <finding 3>**
> - **None — file nothing** *(default, recommended)*
>
> **Rate this flow run:**
> - **5 — excellent**
> - **4 — good**
> - **3 — fine**
> - **2 — rough** — a `1` is typed through the tool's free-text "Other"

all in one `AskUserQuestion` call; the "Then ask the operator to rate the run" sentence goes. The
reasoning-pass subagent's "does not ask the rating question, and does not run the filing prompts"
becomes "does not run the filing-and-rating prompt".

### `skills/flow-contracts/operator-prompts.md` — **The multi-select variant**

The last sentence, "The two existing multi-select call sites choose oppositely: the self-review
filing ask's safe default is "file nothing" (silence selects None), while
`skills/myflow-do/SKILL.md`'s optional review-slot ask defaults to "include all of them" (silence
selects every fired trigger)." becomes: "The one live multi-select call site, the self-review
filing ask, chooses the empty set: silence selects **None — file nothing**." The
`skills/myflow-do/SKILL.md` reference goes with it.

## Budgets

`check-contract-budget.sh` rows, current size / budget on `main`: `archive.md` 16420/18748,
`brainstorm-planner.md` 18619/23248, `brainstorm.md` 10006/35015, `operator-prompts.md` 2021/2432,
`pipeline.md` 24595/36155, `pipeline-rationale.md` 14706/20935, `project-configuration.md`
46575/48175, `finish-contract-run2.md` 30129/36019, `.flow/project.md` 25224/26450. Every edit
above is net-neutral or a few hundred bytes; a trip raises that row by the overage and says so in
the commit body.

## Decisions

### Chunk findings across up to three questions in one call

**ID:** chunk-findings-one-call
**Status:** active
**Chosen:** up to three multi-select questions of 3 findings + None each plus the rating as the
fourth question, overflow past 9 into one further call — the tool caps a call at 4 questions and a
question at 4 options, so "one question listing every finding" cannot hold more than 3.
**Considered:** one question capped at 3 findings, the rest recorded as declined-not-offered —
silently drops findings from the filing ask.

### Rating options 5/4/3/2, a 1 via free text

**ID:** rating-four-options
**Status:** active
**Chosen:** `5 — excellent` / `4 — good` / `3 — fine` / `2 — rough`, a `1` typed through the
tool's "Other" — the 1–5 integer scale the report and handoff carry survives intact.
**Considered:** `1–2 — rough` recorded as `2` — a 1 becomes inexpressible from the options.

### Drop the stale second multi-select example

**ID:** drop-myflow-do-example
**Status:** active
**Chosen:** name the self-review filing ask as the one live multi-select call site — no optional
review-slot multi-select ask exists anywhere under `skills/flow/` today.
**Considered:** re-point the reference at a current file — none carries that ask.

### No guard case for `## self review`

**ID:** no-self-review-guard
**Status:** active
**Chosen:** the phase file validates the body the way `## default landing route` is validated
today; a bad body is reported and dropped.
**Considered:** a `check-model-keys.sh` case — a guard for a two-literal key that already fails
safe by dropping.

### Revise is a round that re-presents the changed section

**ID:** revise-is-a-round
**Status:** active
**Chosen:** *Revise* counts toward the round-3 offer like *Another round*; the next turn opens by
re-presenting the changed design section(s) before the confirm instead of asking new questions.
**Considered:** *Revise* outside the round count — an unbounded loop the round-3 offer could not
see.

### Handoff line names the project default

**ID:** skipped-project-default-line
**Status:** active
**Chosen:** `**Self-review:** skipped — project default` on the key, `handoff-blocks.md`'s shape
statement matched — the reader can tell a configured skip from a declined prompt, as `Route: … —
from this project's configured default, not asked` already does.
**Considered:** plain `skipped` — indistinguishable from a decline.

## Open questions

None.
