# Quiet progress

Loaded by the load directive of **Quiet progress** (`skills/flow-contracts/pipeline.md`), only when
the project's `## progress` key turns the mode on.

**With the mode on, the run writes no status prose of its own on a routine event** — dispatching a
subagent, a subagent returning clean, sending a subagent a decision, waiting on a background agent,
a guard passing, a commit landing. This suspends the per-unit and what-I-am-on lines of the
"Keep the user posted" paragraph of `rules/be-brief.mdc` for the run, and nothing else in that
rule: a turn still ends only where it says.

**It likewise writes nothing for** a bookkeeping note about a plan-field correction ("Q2
re-captured unchanged, so its baseline is dropped from task 14's Files.") or an end-of-turn
progress summary ("Task 13 passed review and is ticked … Task 14 … is running now."), a narrated
step ("Now commit the frontend link and run close-task for tasks 1–6.", "Run RED first.", "Now the
fix."), or a stage-level status line ("🔍 Review panel (primary + principles) — in review"). When
nothing needs the operator — no stop, no auto-resolution line, no handoff — the turn ends with no
text at all.

**What still prints:**

- a stop for the operator — a question, a `## Question` handback, a blocked step;
- one line per auto-resolution taken, naming the question and the option (**Auto-resolution**,
  `skills/flow-contracts/operator-prompts.md`);
- the handoff block;
- everything on the **Never compress** list of `rules/be-brief.mdc`, in full.

**Progress visibility** (`skills/flow-contracts/pipeline.md`) is unaffected: the flow-task-list mod's view, and the equivalent block
a harness without it prints, are not the run's status prose.
