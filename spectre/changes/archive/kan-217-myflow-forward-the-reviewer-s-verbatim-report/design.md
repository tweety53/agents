# Design — forward the reviewer's verbatim report to the fix agent

## Context

A panel slot's report reaches the dispatcher and goes no further: the dispatcher distills it into a
one-line `flow record finding -note` and discards the rest. At fix time the fix subagent receives
that distillation alone. `proposal.md` is canonical for the problem and its measured cost; this
file is canonical for the mechanism, the decisions, and the guard's resulting shape.

## Capturing the report

The dispatcher writes each dispatched slot's returned report, byte for byte, to

```text verified:authored in-tree for this change
<abs-worktree>/.superpowers/sdd/panel-report-<round>-<id>.md
```

at the point that slot's `flow record dispatch end` is recorded — the moment the report is in hand.

- `<round>` is the value that round's findings carry on `flow record finding -round`: `0` for the
  initial panel, `1..n` for a fix round.
- `<id>` is the resolved reviewer id — `primary`, `principles`, `code-review-low`, `bugbot`,
  `security` — never the slot's display name.

Every dispatched slot writes a file, including a slot that raised nothing. A slot substituted per
**An unspawnable id is substituted, not skipped** writes under its own id, so the file set matches
the resolved roster regardless of how a slot was spawned.

Where nothing verbatim is available — an agent that returned no report, or a harness that hands the
dispatcher only its own summary of one — the file is still written and carries exactly one line:

```text verified:authored in-tree for this change
no verbatim report captured — <reason>
```

## Forwarding it to the fix subagent

The fix-subagent dispatch prompt gains one labelled blockquote, and the per-finding structured
block gains its slot's report path as a field alongside `F<n>`, slot, severity, `file:line`, theme,
reproducer text and bounces.

The paragraph states four things: the file is the reviewer's own report, unedited; read it before
acting on the finding; the structured block is the dispatcher's summary of it, direction and never
a source of fact; and where the two disagree the report wins, while a block-only assertion is
unchecked until the fix agent establishes it.

Only the slots behind this round's surviving findings have their reports forwarded. Every slot's
report is written; not every slot's is read.

## Decisions

### Where the verbatim report is persisted

**ID:** report-as-worktree-file
**Status:** active
**Chosen:** a file per slot per round under `<abs-worktree>/.superpowers/sdd/` — it sits beside
`final-review.diff`, `fix-round-N.diff` and `slot-delta-<round>-<slot>.diff`, is removed with the
worktree, and needs no Go, schema, migration or render change, matching the ticket's "cost: near
zero".
**Considered:** a store column on the dispatch row (`flow record dispatch end -report`) — survives
the archive and is queryable, but costs a migration, CLI changes, render changes and their tests,
far past what the ticket asks for; a file now plus a store column filed as a follow-up — doubles
the surface for the same benefit at fix time.

### Which dispatches carry the report

**ID:** panel-fix-round-only
**Status:** active
**Chosen:** `skills/flow/review-panel.md`'s fix subagent alone — the exact path the ticket names
and the one KAN-189's F20/F23 came through.
**Considered:** also the per-task review handback in `skills/flow/implement.md` — that review hands
back to the same implementer, which already holds the context, so the re-derivation cost this
change targets does not arise there; also the bounce path back to a raising slot — that slot
authored the report it would be handed.

### Filenames key on the reviewer id, not the slot name

**ID:** filename-keys-on-id
**Status:** active
**Chosen:** the resolved id, a closed vocabulary (`ValidReviewers` in
`<agents repo>/stats/internal/store/settings.go`), so no filename needs ad-hoc slugging.
**Considered:** the slot display name — `Code review (low)` carries a space and parentheses, which
would need a slugging rule invented here and kept in step with the roster table.

### A report is written for every slot, forwarded for some

**ID:** write-all-forward-some
**Status:** active
**Chosen:** write on every dispatch, forward only the slots behind surviving findings — a clean
slot's report is the evidence a `findings-total: 0` render rests on, while forwarding it would put
context the fix agent has no finding for into its window.
**Considered:** writing only for slots that raised findings — cheaper by nothing measurable, and it
loses the clean-slot evidence.

### An uncapturable report writes the file anyway

**ID:** uncapturable-writes-the-reason
**Status:** active
**Chosen:** write the file with `no verbatim report captured — <reason>`, so the fix agent is told
the fact is missing.
**Considered:** omitting the file — the fix agent then silently falls back to the block alone,
which is the exact failure this change exists to stop, and nothing in the record would say it
happened.

### `-note` carries the reviewer's own sentence

**ID:** note-quotes-the-reviewer
**Status:** active
**Chosen:** require it in **Recording findings** — the paraphrase is created at that call, so this
closes the hole at its source rather than only downstream, and it costs prose alone.
**Considered:** leaving `-note` as a dispatcher summary and relying on the forwarded report to
correct it downstream — narrower, but it leaves the rendered panel table, which outlives the
worktree, carrying words no reviewer wrote.

### The guard is generalized and renamed rather than duplicated

**ID:** generalize-and-rename-the-guard
**Status:** active
**Chosen:** collapse `check-reproduce-not-read.sh`'s single-label constants into a table of
required dispatch paragraphs and rename it `check-dispatch-paragraphs.sh` — its block-extraction
and phrase-matching machinery already does exactly this job, and the rename keeps the name
describing what the guard covers.
**Considered:** a new `scripts/check-verbatim-report.sh` — a second copy of the same machinery, one
more `## lint` entry and one more harness; extending in place without renaming — zero rename
ripple, but the script's name would then describe half of what it does, and this repository's
guards are read by name.

**Rename cost, measured:** two file renames plus five citations —
`.flow/project.md` lines 106, 158 and 247, `scripts/test-check-plan-shape.sh:20`, and
`spectre/changes/kan-121-run-the-guards-own-parsers-over-tasks-md-at-plan/tasks.md:77`.
`scripts/check-references.sh` catches any missed among the live ones.
<!-- measured: grep -rn "check-reproduce-not-read" . --exclude-dir=.git @ 2026-08-28, branch main -->

### A finished change's plan is left frozen

**ID:** leave-finished-plans-frozen
**Status:** active
**Chosen:** repair only the live citation, `scripts/test-check-plan-shape.sh:20`, and leave
`spectre/changes/kan-121-run-the-guards-own-parsers-over-tasks-md-at-plan/tasks.md:77` alone — that
plan records what was true when it was written, `scripts/check-references.sh` does not scan
`spectre/changes/`, so nothing breaks, and kan-121's own run set the precedent by dropping its
repair to kan-359's plan once that change was out of reach.
**Considered:** rewriting the finished plan's line to the new name — it would make a landed
change's record describe a guard that did not exist while that change ran, trading a harmless
stale name for a false one.

### Forwarding a report does not breach "Inline no source excerpt"

**ID:** path-not-paste
**Status:** active
**Chosen:** hand the report over as an absolute path, never pasted into the prompt — nothing is
inlined, so that rule keeps binding the dispatcher's own block unchanged.
**Considered:** pasting the report into the fix prompt — it would breach the rule outright and put
a stale copy in the prompt where a path reads the live file.

## Open questions

None.

## The guard's shape after the change

`scripts/check-dispatch-paragraphs.sh` holds one table of required paragraphs, each entry carrying
its label, its shared phrases, its variants' own phrases, and the sites that require it with each
site's minimum block count and required variants.

| Paragraph | Site | Min blocks | Variants |
|---|---|---|---|
| `**REPRODUCE, DON'T READ:**` | `skills/flow/review-panel.md` | 1 | reviewer |
| `**REPRODUCE, DON'T READ:**` | `skills/flow/implement.md` | 2 | reviewer **and** implementer |
| `**VERBATIM REPORT — THE FACT:**` | `skills/flow/review-panel.md` | 1 | *(none)* |

The existing block extraction, the `file:line` reporting shape, and the exit-code contract — `0`
clean, `1` a required site missing its block or a phrase, `2` cannot answer at all — are unchanged.
`CHECK_REPRODUCE_NOT_READ_ROOT` becomes `CHECK_DISPATCH_PARAGRAPHS_ROOT`, keeping its
set-but-empty refusal.

The VERBATIM REPORT entry's phrases are held as short literals, per the label-plus-phrases approach
the guard already uses, so a paragraph reworded around them still passes:

- `the reviewer's own report`
- `never a source of fact`
- `the report wins`

**What a green run does not prove**, unchanged from the guard's own header: only that each label
and its phrases are present at each required site — never that a dispatcher actually wrote the
report file, and never that a fix agent read it.
