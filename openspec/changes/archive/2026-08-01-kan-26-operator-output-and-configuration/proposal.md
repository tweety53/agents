## Why

KAN-17 carried eighteen asks. Slice A took the eleven that change how the pipeline behaves, plus one
raised during its planning run; this is slice B, the remainder — the asks that change what the
operator sees and how the pipeline is configured, plus one appended to KAN-26 after filing.

Two of the eight surfaces they touch have no capability in `openspec/specs/` today: the progress view
and the manual test guide's shape. That is the same gap slice A found, where the In Review defect
survived four archived changes precisely because no requirement existed for it to violate. Both get
a spec here rather than a convention.

The full brainstorming record is
`docs/superpowers/specs/2026-07-31-kan-26-operator-output-and-configuration-design.md`.

## What Changes

- **`/myflow-status <name>` regenerates the state's full handoff block** rather than printing a
  short "next" cell. The block's shape is defined once, in `pipeline.md`, and rendered by both the
  emitting command and `/myflow-status`, so the two cannot drift and nothing is stored to go stale.
- **Every pipeline command prints `/rename <name>` and `/color cyan`** at the start of its run, for
  the operator to paste. The spec records why they are printed rather than invoked, so a later
  reader does not retry it: `SlashCommand` exposes only `type: "prompt"` commands, and the harness
  gives Bash no writable `/dev/tty`.
- **`pipeline.md` gains a two-level stage table** beneath its state diagram — one row per command,
  with eight stages expanded to show the substructure they hide, the review panel among them. It
  states structure and cites the owning file for tuned thresholds. **The README drops its diagram**
  and links instead.
- **Planning effort is renamed.** Levels become `low` / `default` / `detailed`; the concept is called
  "planning effort"; the state-file key becomes `planningEffort`. **BREAKING** for the state file's
  key name, mitigated by reading the old key as its equivalent level and rewriting it in place — no
  migration pass, and no self-heal warning. That mitigation was removed at the review gate and then
  **restored** when the measurement it was removed on proved wrong; the full history, and the one
  thing that did change (an unmapped value reads as *not recorded*, never as *unparseable*), are
  under `rename-reaches-capability` in `design.md`.
- **Every pipeline command drives the harness's task-list mechanism**, so the live task count and
  per-task list render throughout a run. Stated harness-neutrally, with a printed block where a
  harness offers none.
- **`/myflow-start`'s creating run asks and records three models** — implementation, review panel,
  and panel fixes — in a `models` object in the state file. Defaults are unchanged; the field makes
  an override recorded rather than transcript-only.
- **The manual test guide becomes a capability-scoped behaviour checklist** in the operator's
  register, dropping per-step transcripts and rationale. Its checkbox syntax and
  `## Known incomplete` section are unchanged, because two guards read them.
- **Follow-up issues are named `<KEY> follow-up` and joined** when one already sits in To Do, rather
  than filed alongside it.
- **The per-state handoff templates are reconciled with the blocks the commands actually print.**
  Added during implementation, on the operator's decision, after the review panel found that the
  templates and the three commands disagreed on both label style and field set — so the change would
  have violated the single-definition requirement it introduces. `IN_PROGRESS` gains a second
  template for `/myflow-finish` run 1, whose handoff waits on a merge rather than on a diff review,
  and `Jira description (pre-edit)` is named a run-only field no regenerated block can reproduce.

## Capabilities

### New Capabilities

- `myflow-progress-visibility`: every pipeline command registers its steps with the harness's
  task-list mechanism and keeps their statuses current, so the live progress view renders; the
  widget is a view and never a record.
- `myflow-manual-test-guide`: the register, scope and required sections of
  `docs/manual-test/<name>.md`.

### Modified Capabilities

- `myflow-handoff-output`: the handoff block's shape is defined once and rendered by two commands;
  `/myflow-status <name>` regenerates it; every pipeline command prints the tab commands at the start
  of its run.
- `myflow-contract-distribution`: the pipeline diagram and its two-level stage table live in
  `pipeline.md`, which no other file copies.
- `myflow-model-policy`: three model roles are asked once and recorded in the state file; the
  recorded value is the operator override this capability already permits; the ledger still records
  what each dispatch actually ran on.
- `myflow-review-panel-economics`: "every slot runs on Sonnet" becomes the default rather than an
  absolute, pointing at the recorded override.
- `myflow-state-machine`: the state file carries `planningEffort` and `models`; both have a live
  consumer, and both read as "not recorded" when absent.
- `myflow-jira-projection`: follow-up naming, the join search and its two To Do statuses, and the
  append-only join write.

### Renamed Capabilities

- `myflow-effort` → `myflow-planning-effort`: the capability, its levels and its state-file key are
  renamed together. The delta carries every requirement twice — ADDED under the new name, REMOVED
  under the old.

## Impact

- **Contracts:** `skills/myflow-contracts/pipeline.md` (stage table, handoff block template, model
  policy), `state-file.md` (`planningEffort`, `models`, the planning-effort table),
  `jira-integration.md` (follow-up naming and joining).
- **Skills:** `myflow-start` (three model prompts, renamed effort question, tab lines, task
  registration), `myflow-do` (task registration, guide register, recorded models on dispatch),
  `myflow-finish` (task registration, follow-up naming and joining), `myflow-status` (handoff
  regeneration), `myflow-info` (can show the diagram).
- **Commands:** both trees, wherever they describe the effort question or the states they accept.
- **Guards:** `check-vocabulary.sh` gains one retired literal, `"effort":`. No new guard script.
- **Docs:** `README.md` loses its diagram and links to `pipeline.md`; `CLAUDE.md` and `AGENTS.md`
  where they name the effort levels.
- **Existing state files:** state files on this machine do carry the retired key, several with a
  non-null value, and one of those is **this change's own open state file** — an earlier version of
  this bullet claimed the opposite and was disproved by measurement at the pass-3 review. The
  figures, how they were measured, and what they decided are in `design.md` under
  **The effort rename reaches the capability, not just the levels** (`rename-reaches-capability`);
  they are deliberately not restated here, since a second copy of a count is what let the wrong one
  survive three passes. What the rename needs of those files is the compatibility read, which is in
  force.

## Fix rounds

- **Round 4 (F40–F48).** The pass-4 review panel raised nine findings, two Critical, and the run
  stopped rather than opening the round — recording its reasoning for the operator, who overruled it.
  The round corrects two contract defects that survive the merge boundary or misstate their own
  guarantees (F40, F41), one security gap in the Jira append guard (F43), three internal
  contradictions or restatements in the self-heal and state-file contracts (F42, F44, F45), one stale
  worked example (F48), and — the class this round exists to end — two false claims left in the
  planning artifacts themselves (F46, F47), which archive verbatim and would otherwise ship as
  durable wrong text. Planned as Task 13; no capability is added or removed.
- **Round 5 (F49–F57).** Pass 5 ran the full roster against the round-4 tree and confirmed all nine
  round-4 findings closed, three slots verifying independently. It raised nine of its own, none
  Critical against the diff: two contract claims stated more broadly than their own tables support
  (F49, F56), two hardenings of the Jira join the round-4 security fix left short — an unscoped
  project search and a fence-break that swallowed the very count round 4 added (F50, F51) — and five
  smaller gaps (F52, F54, F55, F57), plus one against this run's own bookkeeping (F53). Planned as
  Task 14; no capability is added or removed.
- **Round 6 (F61–F70).** Pass 6 ran the full roster and confirmed all twelve round-5 findings closed,
  raising **no Critical against the diff**. Its ten findings are four Important — a command tree
  contradicting its skill, a JQL clause built from an unconstrained value, a self-heal write path
  missing the retired-key rewrite, and `/myflow-info` hardcoding a diagram its own guardrail forbids —
  and six Minor. The escalation ladder's five-round limit was **overridden by the operator at the
  handback**, who chose to fix all ten and re-run the full panel rather than withdraw the Minor set.
  Planned as Task 15; no capability is added or removed.
- **Round 7 (F71–F77).** Pass 7 raised seven findings, zero Critical, with three of seven slots clean
  and the primary reporting ready for the human gate. Four Important: an unqualified "kept identical"
  claim between the trigger digest and its canonical source, a project-key check whose safety argument
  needs a whole-value match it never states, and two on the vocabulary guard. The operator overrode
  the ladder a second time and directed that the guard findings be closed by **narrowing the guard's
  claim** rather than widening its pattern a third time — the entry had produced a finding in three
  consecutive passes because it was being asked to prove what no literal list can. Planned as
  Task 16; no capability is added or removed.
