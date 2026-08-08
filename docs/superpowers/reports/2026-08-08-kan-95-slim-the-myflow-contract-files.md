# Implementer and fix-round reports — kan-95-slim-the-myflow-contract-files

Each task's implementer and each panel fix round wrote its own account: what it changed, what
it verified and how, the judgment calls it made, and the sweeps it ran. Preserved here because
`git worktree remove --force` destroys ignored files, and `.superpowers/` is ignored.

The per-move ledgers and the review panel record are preserved separately, under
`docs/superpowers/ledgers/` and `docs/superpowers/reviews/`.


---

## task-1-report

# Task 1 report — Remove `/myflow-info` and move the pipeline explanation to `README.md`

**Status:** DONE

**Verification summary:** All five required guards exit 0
(`check-references.sh`, `check-vocabulary.sh`, `check-contract-budget.sh`,
`test-check-contract-budget.sh`, `test-setup.sh`), and
`grep -rn 'myflow-info' . --exclude-dir=.git --exclude-dir=docs --exclude-dir=archive --exclude-dir=.superpowers`
returns matches only under `openspec/`.

## What was done

**Deleted:**
- `skills/myflow-info/SKILL.md` (and the now-empty `skills/myflow-info/` directory)
- `commands/myflow-info.md`
- `commands-claude/myflow-info.md`

**`skills/myflow-contracts/pipeline.md`:** deleted the entire `## Pipeline flow` section (heading
through the line before `## Command surface` — the mermaid diagram, `### Level 1`, `### Level 2`,
and all nine `####` stage subsections), 64,701 → 50,932 bytes (−13,769, matching the brief's
measured byte count exactly). Also updated: `## Command surface`'s "Three pipeline commands and two
read-only ones" → "Three pipeline commands, plus one read-only one"; the `## State transitions` and
`## Git boundaries` tables' `/myflow-status`, `/myflow-info` rows → `/myflow-status` alone; the
`## Progress visibility` and tab-commands sentences dropped their `/myflow-info` mentions.

**`skills/myflow-contracts/pipeline-rationale.md`:** deleted the mirrored subtree — `## Pipeline
flow` through the last `#### Self-review — /myflow-finish run 2` heading (interpreted as the full
subtree the core section owned, not just the top heading, since the mirror rule is about the
heading *tree* and an orphaned `### Level 1` under nothing would itself be a broken mirror).

**`README.md`:** new `## How the pipeline works` section added after the existing pipeline overview,
carrying the mermaid diagram verbatim, a Level 1 table (four commands now, not five — `/myflow-info`
dropped, `self-review` un-marked since it has no Level 2 expansion per the delta spec's 8-row list),
and all eight required Level 2 expansions, rewritten for a human reader with every threshold-owning
citation preserved. Also removed: the `myflow-info/` tree line, the `/myflow-info` mention in the
skills sentence, the `/myflow-info` row in the commands-reference table, and repointed both the
`Pipeline flow` and `Level 1 — the stages of each command` citations to the new section (self-cited
via `README.md`).

**Repointed citations** (every live citation to the deleted `Pipeline flow` / `Level 1` headings,
beyond the brief's own file list, since check-references.sh scans the whole tree): `CLAUDE.md`,
`AGENTS.md` (skill-index row, commands-table row, digest paragraph — the "Expect both renderings...
`/myflow-info` reads the canonical block live" sentence is now false and was rewritten, not just
repointed — and the three `Level 1` citations in the commands table), `skills/README.md` (command
map row, tree line, both citations), `skills/myflow-start/SKILL.md` (2 citations),
`skills/myflow-start/SKILL-rationale.md` (1 citation).

**Beyond the brief's file list, found necessary by the verification gates and fixed:**
- `skills/myflow-contracts/finish-contract.md` and `skills/myflow-finish/SKILL.md` each cited
  `**Self-review — /myflow-finish run 2** (pipeline.md)`, a heading that's now gone (not one of the
  8 required Level-2 stages). Repointed both to
  `**Requirement: Self-review runs only after FINISHED is written** (openspec/specs/myflow-self-review/spec.md)`
  — the fuller, normative source the old subsection already deferred to.
- `skills/myflow-contracts/project-configuration.md` (2 literal `myflow-info` mentions — an example
  path list and a "this is the shape myflow-info's read falls back to" sentence) and
  `skills/myflow-status/SKILL.md` (a `Pipeline reference | /myflow-info` row) — caught by the
  step-5 grep, not in the brief's Modify list.

**`scripts/check-contract-budget.sh`:** removed the `skills/myflow-info/SKILL.md 6297` row.

**`scripts/test-check-contract-budget.sh`:** the "two skills with different budgets" fixture used
`skills/myflow-info` (budget 6297) as the small-budget skill against `skills/myflow-do` (47356) with
a 20000-byte test file (20000 > 6297, < 47356). Retargeted to `skills/openspec-explore` (real budget
14285): 20000 > 14285 and < 47356 still holds, so the property under test is preserved.

**Staged** via `git add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/'
':(exclude)docs/superpowers/' ':(exclude).serena/'` — added one more exclude than the brief's literal
command, for `.serena/`, Serena MCP's own untracked project metadata directory, not part of this
task's deliverable (the outer repo already carries the same untracked directory unstaged).

## Incident: wrong-repo edit, caught and reverted

Early in the task, before activating Serena on this worktree, a `replace_content` call landed on
`/Users/tweety53/Projects/agents/skills/myflow-contracts/pipeline.md` (the outer main-checkout repo,
Serena's default-active project) instead of this worktree — Serena's active project had not been
switched yet. Caught immediately via `git diff --stat` in that repo (202 lines removed, matching the
`## Pipeline flow` deletion), reverted with `git checkout -- skills/myflow-contracts/pipeline.md`,
confirmed clean, then `activate_project`'d this worktree explicitly and re-verified the target with
a `dry_run` before repeating the edit here. Re-confirmed at the end of the task that the outer repo
carries no modified tracked files.

## Concerns

1. **`openspec/specs/myflow-handoff-output/spec.md` still names `/myflow-info`** (lines 328, 345) and
   is outside the four `openspec/specs/myflow-{command-surface,contract-distribution,progress-visibility}/`
   files tasks.md's Step 5 names as expected-to-still-match. The brief's own context explicitly says
   "Do not edit anything under `openspec/`", which I followed, so I left this file untouched and it
   shows up as a fifth non-`openspec/changes/kan-95-*/` match in the sweep. Likely just an omission
   in Step 5's expected-list (that capability isn't part of kan-95's delta specs), not a defect in
   this task's work — flagging so it isn't mistaken for something Task 1 should have caught.
2. **The `Self-review` ledger row's destination reads `— (rewritten for README.md)` per the brief's
   blanket instruction for this task, but it's the weakest fit of the twelve rows** — that content
   isn't actually reproduced in README (the delta spec's required 8-row Level-2 list excludes it);
   only the bare `self-review` stage name survives in the Level 1 sequence, unmarked, with a citation
   to the OpenSpec capability. Noted inline in the ledger itself.
3. **Interpreted "delete its mirrored `## Pipeline flow` heading" in `pipeline-rationale.md` as "delete
   the mirrored subtree,"** not literally one heading line — leaving just the top heading gone would
   have orphaned `### Level 1`, `### Level 2`, and nine `####` headings under nothing, which breaks
   the very mirror the instruction was protecting rather than repairing it.

## Ledger

See `.superpowers/sdd/task-1-ledger.md` (also reproduced below).

| Removed passage | Source heading | Destination | Pointer left |
|---|---|---|---|
| ` ```mermaid stateDiagram-v2 [*] --> STARTED: /myflow-start STARTED --> ` | Pipeline flow | — (rewritten for README.md) | none |
| **One row per command** — the five this | Level 1 — the stages of each command | — (rewritten for README.md) | none |
| Each expansion states the **structure** — the shape | Level 2 — the stages that hide substructure | — (rewritten for README.md) | none |
| superpowers:brainstorming runs its checklist in full and ends | Brainstorm — `/myflow-start` | — (rewritten for README.md) | none |
| superpowers:writing-plans enriches `tasks.md` from a checkbox scaffold into | Writing-plans — `/myflow-start` | — (rewritten for README.md) | none |
| One implementer dispatch per checkbox in `tasks.md`, or | SDD + TDD per task — `/myflow-do` | — (rewritten for README.md) | none |
| **Three required slots and four conditional ones.** Primary | The review panel — `/myflow-do` | — (rewritten for README.md) | none |
| `scripts/check-finish-preflight.sh` decides which run happens, from three | The preflight verdict — `/myflow-finish` | — (rewritten for README.md) | none |
| Runs **before** the landing question and before any | The unfinished-work gate — `/myflow-finish` run 1 | — (rewritten for README.md) | none |
| The operator is asked once, before any git | The landing routes — `/myflow-finish` run 1 | — (rewritten for README.md) | none |
| Every removal is *remove-or-move if present*, which is | Cleanup — `/myflow-finish` run 2 | — (rewritten for README.md) | none |
| Self-review runs only after `FINISHED` is written — | Self-review — `/myflow-finish` run 2 | — (rewritten for README.md) | none |

---

## task-10-report

# Task 10 report — re-anchor every budget row and record the result

## Status: DONE

## What was done

1. Measured every file covered by `scripts/check-contract-budget.sh`'s guard: all `*.md` under
   `skills/myflow-contracts/`, plus every `skills/*/SKILL.md` and its `SKILL-rationale.md`
   sibling — 23 files total.
2. Recomputed every row in `budgets()` as `floor(actual_bytes × 1.25)`, using the file's current
   size (the size it has at the end of this change). Full arithmetic and old/new values are in
   `.superpowers/sdd/task-10-ledger.md`.
3. Confirmed the two files deleted during this change (`skills/myflow-contracts/state-self-heal.md`,
   `skills/myflow-info/SKILL.md`) already carried no row and no fixture reference — an earlier
   task in this change had already removed them, so no removal was needed here.
4. Confirmed coverage is exact: 23 covered files on disk, 23 rows in `budgets()`, one-to-one, no
   file without a row and no row naming a nonexistent file. The three files created during this
   change (`handoff-blocks.md`, `project-configuration-rationale.md`,
   `workspace-isolation-rationale.md`) each already had a row from their own task; all three were
   re-anchored to their current size like every other row.
5. Measured the real per-run load for `/myflow-do`.
6. Ran every declared lint and test command; all exit 0.

## Files changed

- `scripts/check-contract-budget.sh` — every value in the `budgets()` heredoc updated to
  `floor(actual × 1.25)` against the file's current byte count. No structural change to the
  script's logic.
- `scripts/test-check-contract-budget.sh` — one fixture depends on production budget values:
  the "a SKILL-rationale.md over budget fails" case sizes a fixture file to a fixed byte count and
  asserts it exceeds `skills/myflow-do/SKILL-rationale.md`'s real budget row. That row moved from
  8,370 to 9,665, so the fixture's `head -c 9000` (previously over 8,370, now under 9,665) no
  longer exercised the over-budget path. Changed to `head -c 10000`, safely above the new 9,665
  budget. No other fixture needed a change — the other production-coupled cases
  (`skill-over`, `skill-distinct`) still hold under the new numbers.

## Verification summary

Guard before edit: `BUDGET-OK: 23 contract file(s) within budget` (exit 0) — passed even before
re-anchoring, because every stale budget was more generous than the re-anchored value, never
tighter; the harness's own test suite is what caught the fixture coupling issue after the edit.
Guard after edit: `BUDGET-OK: 23 contract file(s) within budget` (exit 0). Guard's own test harness:
failed once after the row change (`FAIL a SKILL-rationale.md over budget fails: exit 0, wanted 1`),
fixed by widening the fixture's fixed size past the new budget, then `all checks passed` (exit 0).
All 6 lint commands and all 12 test commands listed in the brief exit 0 — see full list below.

## Per-run byte figure for /myflow-do

```
   35825 skills/myflow-do/SKILL.md
   31151 skills/myflow-contracts/pipeline.md
   38275 skills/myflow-contracts/project-configuration.md
   25317 skills/myflow-contracts/workspace-isolation.md
   15362 skills/myflow-contracts/state-file.md
   15591 skills/myflow-contracts/jira-integration.md
  161521 total
```

**New total: 161,521 bytes.** Starting figure: 210,481 bytes. Delta: −48,960 bytes, a 23.27%
reduction.

The change's proposal projected roughly 128,500 bytes. The measured figure, 161,521 bytes, misses
that projection by +33,021 bytes (25.7% above the projected number). This is the honest measured
result, not an adjusted one: three of the four evicted files classified more conservatively than
the proposal assumed, each time deliberately and each time confirmed by review during earlier
tasks in this change. The projection was a projection, not a commitment, and it did not land.

## Lint and test results (all exit 0)

Lint:
- `scripts/check-vocabulary.sh` — `✓ Stage-vocabulary guard: clean` / `✓ Panel-vocabulary guard: clean`
- `scripts/check-references.sh` — `check-references: all referenced sections resolve`
- `scripts/check-plan-provenance.sh` — `check-plan-provenance: 3 file(s) scanned, all provenance stated`
- `scripts/check-task-build-green.sh` — clean, no output
- `scripts/check-workspace-isolation.sh` — `ISOLATION-OK: … declares no ## workspace isolation section`
- `scripts/check-contract-budget.sh` — `BUDGET-OK: 23 contract file(s) within budget`

Test:
- `scripts/test-setup.sh` — `PASS — 192 assertions, 0 failures`
- `scripts/test-check-references.sh` — `All check-references assertions passed`
- `scripts/test-check-plan-provenance.sh` — `All check-plan-provenance assertions passed`
- `scripts/test-check-finish-preflight.sh` — `check-finish-preflight: all cases pass`
- `scripts/test-preserve-session-records.sh` — `preserve-session-records: all cases pass`
- `scripts/test-check-unfinished-work.sh` — `check-unfinished-work: all cases pass`
- `scripts/test-check-cleanup-complete.sh` — `check-cleanup-complete: all cases pass`
- `scripts/test-gather-self-review-context.sh` — `gather-self-review-context: all cases pass`
- `scripts/test-uncommitted-review-package.sh` — `uncommitted-review-package: all cases pass`
- `scripts/test-check-task-build-green.sh` — `all cases passed`
- `scripts/test-check-workspace-isolation.sh` — `check-workspace-isolation: all cases pass`
  (one sub-case skipped: gymie's real declaration, unreadable from this worktree — pre-existing,
  unrelated to this task)
- `scripts/test-check-contract-budget.sh` — `all checks passed`

## Ledger

`ledger: no passages removed` — this task moves no prose, only budget numbers. Full per-row table
in `.superpowers/sdd/task-10-ledger.md`.

## Staged

`scripts/check-contract-budget.sh` and `scripts/test-check-contract-budget.sh` staged via
`git add`. `.serena/` was not staged (never touched by `git add -A`, only these two named files
were added). Confirmed `git -C /Users/tweety53/Projects/agents status --short` shows no tracked
`M` entries.

## Concerns

None. Every covered file has exactly one row; no row lacks a file; the guard and its harness both
pass; the per-run figure is the honest measured number, reported as-is even though it misses the
proposal's projection.

---

## task-2-report

# Task 2 report — Move the handoff block templates to `handoff-blocks.md`

**Status: DONE**

## What was done

1. Created `skills/myflow-contracts/handoff-blocks.md` with the header shape copied from
   `skills/myflow-contracts/finish-contract.md` (verbatim, per the brief), followed by a
   `## Handoff blocks` heading and the moved `### The block each state renders` section — moved
   byte-for-byte from `skills/myflow-contracts/pipeline.md` (lines 265–465 of the file as it stood at
   commit `4be855f25a6e097e8fadab9a6754f220310984f4`), with exactly one permitted edit applied: the
   trailing position word `below` was deleted from the final sentence, `**IntelliJ commands**
   (\`pipeline.md\`) below.` → `**IntelliJ commands** (\`pipeline.md\`).`, because the citing text now
   lives in a different file from the `## IntelliJ commands` section it points at. The citation's
   path itself needed no change: `pipeline.md` (bare) already resolves correctly from
   `skills/myflow-contracts/` via `check-references.sh`'s referring-file-directory fallback, and the
   repo already uses this bare-filename form for same-directory cross-file citations (e.g.
   `finish-contract.md` citing `pipeline.md`).

2. Left the stub in `pipeline.md` under `## Handoff output`, exactly as specified in the brief
   (verbatim). The section boundary I moved was `### The block each state renders` up to (not
   including) the sibling `### The tab commands, printed at the start of a run` heading — that
   subsection is a separate concern (tab naming/colour) not part of "the block each state renders"
   and stays in `pipeline.md` untouched, immediately followed by `## IntelliJ commands` as before.

3. Repointed every citation of the form `**The block each state renders** (\`skills/myflow-contracts/pipeline.md\`)`
   to `skills/myflow-contracts/handoff-blocks.md`, in:
   `skills/myflow-start/SKILL-rationale.md`, `skills/myflow-do/SKILL-rationale.md`,
   `skills/myflow-finish/SKILL-rationale.md`, `skills/myflow-start/SKILL.md`,
   `skills/myflow-do/SKILL.md`, `skills/myflow-finish/SKILL.md`, `skills/myflow-status/SKILL.md`
   (4 occurrences), and `openspec/specs/myflow-handoff-output/spec.md` (edited but deliberately left
   **unstaged** — see Scope decision below).

   I also fixed one additional stale reference in `skills/myflow-status/SKILL.md` step 4, which
   pointed at `**Handoff output** (pipeline.md)` for "the per-state template" — a citation the grep
   for the literal string didn't catch (different bold token) but which became factually wrong by the
   same move, since the per-state templates are no longer under `## Handoff output` in `pipeline.md`.
   Rewrote that sentence to explicitly load `handoff-blocks.md` there instead, which doubles as the
   brief's required "load line at step 4's detail view".

   `skills/myflow-contracts/pipeline-rationale.md:32` (a same-named rationale heading) and
   `openspec/changes/kan-95-slim-the-myflow-contract-files/{tasks.md,proposal.md}` also match the
   grep but carry no citation-with-path to repoint (plain prose mentioning the term, or — for
   pipeline-rationale.md — Task 6's job, not this task's). Left untouched.

4. Wired the loaders: added the load/repoint line in `skills/myflow-status/SKILL.md` step 4 (above),
   and added the `handoff-blocks.md` row to the **Index** table in `skills/myflow-contracts/SKILL.md`,
   stating it is loaded by `/myflow-status` and no other command.

5. Added the budget row: `skills/myflow-contracts/handoff-blocks.md 16765` (actual size 13,412 bytes
   × 1.25 = 16,765, exact, floored) to `budgets()` in `scripts/check-contract-budget.sh`, in
   alphabetical position between `finish-contract.md` and `jira-followups.md`.

6. Staged everything my task touched, excluding `openspec/`, `docs/manual-test/` and
   `docs/superpowers/`, per the required `git add` pathspec.

## Scope decision: the live OpenSpec spec was edited but left unstaged

`openspec/specs/myflow-handoff-output/spec.md` (the **live**, pre-change capability spec — distinct
from this change's delta spec, which already had the correct citation) contained a real citation to
`**The block each state renders** (\`skills/myflow-contracts/pipeline.md\`)`. I repointed it for
corpus accuracy, but did **not** stage it: `openspec/` specs are synced from a change's delta at
archive time by `/myflow-finish`, not by `/myflow-do`/implementation tasks, and the task's own
"MYFLOW — NO COMMITS" boundary explicitly excludes `openspec/` from staging. Neither
`check-references.sh` nor `check-vocabulary.sh` scan `openspec/` (their `DEFAULT_TARGETS` are
`rules skills commands commands-claude README.md AGENTS.md CLAUDE.md` / `+ scripts`), so this edit
was not required for any guard to pass — it's a courtesy correction, left in the working tree
unstaged, consistent with the Step 7 exclusion.

`openspec/changes/kan-95-slim-the-myflow-contract-files/tasks.md` and `proposal.md` were **not**
edited: they mention "The block each state renders" only in plain descriptive prose about this task
itself (no backticked-path citation to repoint), and they are the change's own planning artifacts,
not files any `/myflow-*` command loads.

## Move-fidelity check (the point of the task)

```
git show 4be855f25a6e097e8fadab9a6754f220310984f4:skills/myflow-contracts/pipeline.md > /tmp/kan95-t2-before.md
```

Extracted lines 265–465 of that before-image (`### The block each state renders` through the line
immediately preceding `### The tab commands, printed at the start of a run`) and diffed against the
same span as written into `handoff-blocks.md` (lines 12–212, i.e. everything from that file's
`### The block each state renders` heading to EOF). Verified first that the before-image extraction
is byte-identical to what I actually sourced the move from in the worktree (it is — `diff` empty).

**Full diff, line by line:**

```
201c201
< **IntelliJ commands** (`pipeline.md`) below.
---
> **IntelliJ commands** (`pipeline.md`).
```

That is the **only** line that differs, and it is accounted for by the permitted position-word
deletion described in item 1 above (`below` deleted; the citation's `pipeline.md` path itself is
unchanged because the target — `## IntelliJ commands` — did not move). No other line differs by so
much as one byte.

## Ledger

See `.superpowers/sdd/task-2-ledger.md` (one row — the whole section moved as a single unit).

| Removed passage | Source heading | Destination | Pointer left |
|---|---|---|---|
| The block a state hands off is defined | `### The block each state renders` (under `## Handoff output`) | `skills/myflow-contracts/handoff-blocks.md` § The block each state renders | "The block a state hands off is defined in **The block each state renders** (`skills/myflow-contracts/handoff-blocks.md`), which is canonical for the three per-state templates, the run-only rule and the rendering-selection table. `/myflow-status` loads it; a producing command carries only the block it prints." |

## Verification

```
scripts/check-references.sh    → check-references: all referenced sections resolve   (exit 0)
scripts/check-vocabulary.sh    → ✓ Stage-vocabulary guard: clean / ✓ Panel-vocabulary guard: clean (exit 0)
scripts/check-contract-budget.sh → BUDGET-OK: 22 contract file(s) within budget       (exit 0)
```

Plus the move-fidelity diff above (one line, accounted for).

## Worktree/isolation check

`git -C /Users/tweety53/Projects/agents status --short` was run before finishing. It shows no `M`
(modified-tracked) entries at all — only pre-existing staged `A` entries from earlier work in this
change (`openspec/changes/...`, `docs/superpowers/specs/...`) and an untracked `.serena/`, none of
which are files this task touched or created. No edits were made to the main checkout by this task.

## Files touched

- Created: `/Users/tweety53/Projects/agents-worktrees/openspec-kan-95-slim-the-myflow-contract-files/skills/myflow-contracts/handoff-blocks.md`
- Modified (staged): `skills/myflow-contracts/pipeline.md`, `skills/myflow-status/SKILL.md`,
  `skills/myflow-contracts/SKILL.md`, `scripts/check-contract-budget.sh`,
  `skills/myflow-start/SKILL-rationale.md`, `skills/myflow-do/SKILL-rationale.md`,
  `skills/myflow-finish/SKILL-rationale.md`, `skills/myflow-start/SKILL.md`,
  `skills/myflow-do/SKILL.md`, `skills/myflow-finish/SKILL.md`
- Modified (unstaged, deliberately): `openspec/specs/myflow-handoff-output/spec.md`
- Ledger: `.superpowers/sdd/task-2-ledger.md`

---

## task-3-report

# Task 3 report — Compress the handoff blocks

**Status: DONE**

## What was done

1. **`skills/myflow-contracts/handoff-blocks.md` — folded the `STARTED` template.** Replaced the
   four separate `Decisions recorded`, `Open questions`, `Planning effort` and `Models` lines with
   the single folded line the brief specified verbatim:
   `**Recorded:** <N> decisions · <N> open questions · effort <level, or "not recorded — planned at
   default"> · models <implementation>/<reviewPanel>/<panelFix>`. `Jira` and `Jira description
   (pre-edit)` were left unfolded — both are run-only, and folding them into the on-disk `Recorded`
   line is exactly what the new mixing rule forbids.

   Updated the explanatory paragraph directly under the template (previously "Why the
   open-questions line is not run-only") to match the new geometry: it now says the decisions and
   open-questions counts sit inside the same `Recorded` line rather than being two separate lines,
   and explicitly states both still read `none` when zero — the wording that used to live inline in
   the two placeholders now lives in this paragraph. This is what keeps the fold a pure layout
   change: no rendering behavior for a zero count was altered, only where its description sits.

2. **Folded the `/myflow-do` template.** `Panel` stayed on its own line (run-only). `Progress` and
   `Git` (both on-disk) folded into `**Staged:** <completed>/<total> tasks · <git state>`. Reworded
   every subsequent reference from "the `Git` line" to "the git state on the `Staged` line" /
   "git state" (prose paragraph, table header, and the review-command instruction inside the
   template) — the Git-state ↔ review-command pairing table itself is otherwise unchanged.

3. **Restated the label rule as a folded-line rule.** "The label set and the field set are
   identical" is now "the same folded lines, the same fields within each folded line, and the same
   order," and I added the new paragraph requiring a folded line to group values of one kind
   (on-disk or run-only) and never both, with the `Jira` line's own-line status as the concrete
   example — matching the delta spec's normative text almost verbatim.

4. **Mirrored the fold into the producing skills.**
   - `skills/myflow-start/SKILL.md`: folded its own `Decisions recorded`/`Open
     questions`/`Planning effort`/`Models` lines into one `Recorded` line, keeping its existing
     literal enumerations (`<N> decisions | none`, `<N> open questions | none`, the three-way
     effort enumeration including "reused from the creating run", and the per-model
     "not recorded" enumeration) — all inside the canonical placeholder's value space.
   - `skills/myflow-do/SKILL.md`: folded its `Progress`/`Git` lines into `**Staged:** N/N tasks ·
     staged and uncommitted | committed as two commits and pushed to the PR branch`, and reworded
     the "Print one review command, the one that matches the `Git` line" sentence to reference the
     `Staged` line's git state instead.
   - `skills/myflow-finish/SKILL.md`: **not edited.** Its two blocks (`IN_PROGRESS` after run 1,
     and the run-2 terminal block) have no two adjacent fields of the same kind — `Change`
     (on-disk), `Route` (run-only), `PR` (on-disk), `Outstanding` (run-only) alternate kind field by
     field — so the fold rule finds nothing to fold there, and the brief's Step 1/Step 2 templates
     don't touch these blocks either. Confirmed by grep that no stray "Git line" / "Progress:" /
     "Decisions recorded" reference exists in this file that my renames elsewhere would have made
     stale.

5. Staged everything with the required exclusion pathspec.

## Value-preservation check (required deliverable)

Full before/after value lists for every block touched are in
`.superpowers/sdd/task-3-ledger.md`. Summary: both folds are set-preserving. The only thing that
moved besides line layout is the "reads `none`/not recorded when absent" wording for the two
`STARTED` counts, which shifted from inline placeholder text to the paragraph immediately below the
canonical template — called out explicitly in that paragraph so it isn't silently lost.

## Ledger

`.superpowers/sdd/task-3-ledger.md`: `ledger: no passages removed` — this task reshapes lines in
place, no prose was moved or evicted.

## Verification

```
scripts/check-references.sh      → check-references: all referenced sections resolve   (exit 0)
scripts/check-vocabulary.sh      → ✓ Stage-vocabulary guard: clean / ✓ Panel-vocabulary guard: clean (exit 0)
scripts/check-contract-budget.sh → BUDGET-OK: 22 contract file(s) within budget          (exit 0)
```

Also ran (not required by the brief's Step 5, but listed under Global Constraints):
`check-plan-provenance.sh` (3 files scanned, all provenance stated), `check-task-build-green.sh`
(exit 0), `check-workspace-isolation.sh` (exit 0). Ran all three required scripts once before
editing (RED baseline: all three already passed, since this task changes layout only and adds no
new stale references) and again after editing (GREEN), per TDD discipline — no guard was weakened
or suppressed.

## Worktree/isolation check

`git -C /Users/tweety53/Projects/agents status --short` shows no `M` (modified-tracked) entries —
only pre-existing staged `A` entries from earlier tasks in this change and an untracked `.serena/`.
No edits were made to the main checkout by this task. I did not need Serena tools for this task
(pure Markdown editing via Read/Edit), so no `activate_project` call was made or needed.

## Files touched

- Modified (staged): `skills/myflow-contracts/handoff-blocks.md`, `skills/myflow-start/SKILL.md`,
  `skills/myflow-do/SKILL.md`
- Not modified: `skills/myflow-finish/SKILL.md` (listed in the brief's Files section, but its two
  blocks have no fold to apply — see item 4 above)
- Ledger: `.superpowers/sdd/task-3-ledger.md`

---

## task-4-report

# Task 4 report — Slim `/myflow-status` — columns and data sources

## Status: DONE

## What changed

File touched: `skills/myflow-status/SKILL.md` (the only file this task is scoped to). No other file
in the worktree was modified by this task.

### Step 1 — dropped the worktree column

In `### 3. Render the table`, removed the `Worktree / branch` column, its separator cell, and its
two example values from the rendered `## myflow status` table. The table now reads `Change | Jira |
State | PR | Next | Updated`, matching the brief's example verbatim. Replaced the old
"Worktree paths are absolute..." sentence with the brief's required replacement: "The absolute
worktree path is given in the detail view, taken from the `worktrees` keys."

### Step 2 — removed the `gh pr list` probe

Deleted item 5 of `### 2. Resolve each change's state` (the `gh pr list --head <branch> ...` line)
entirely. Rewrote the PR column's description (also in step 3) to: the number is parsed from
`prUrl`, `—` when it is `null`, and it never reports open/merged/closed because that needs a network
call this command no longer makes. Updated the table's example PR cell from `#42 open` to `#42` to
match (the brief's own step-1 example already showed the post-step-2 value, so both steps land on
the same cell).

### Step 3 — merge status untouched

Item 4 of step 2 (the three-step merge-base resolution, the resolve-before-compare rule, the
inconclusive cases, and the multi-repo combination rule) was not touched — verified by diff review,
no lines in that block changed.

### Step 4/5 — verify and stage

All four required checks pass (see Verification below). Staged only this file with `git add --
skills/myflow-status/SKILL.md`; `openspec/`, `docs/manual-test/`, `docs/superpowers/` were never
touched or staged.

## The ordering subtlety — the permitted `prUrl` correction

Per your explicit instruction, I did **not** delete the self-heal bullet "Do apply the one permitted
correction" (in `### 2`, under "Self-heal is monotonic..."). Its old wording tied the correction's
evidence directly to "this command's own `gh pr list` probe (step 2, item 5)" — that citation now
points at nothing, since item 5 is gone. I reworded the bullet in place (still ~8 lines, same
position) to say plainly that the correction's precondition — a conclusive PR-non-existence probe run
by this command — can no longer be met, so the correction never fires from `/myflow-status` any
longer, while the correction's definition itself (what it would do, if it could fire) stays. The
sentence explicitly says removing it is a later task's job, not this one's, per your instruction that
Task 5 owns that text. I also updated the two other "PR state" promises that referenced the same
retired capability so nothing in the file still implies this command checks GitHub: the detail-view
bullet ("PR number, state, and URL when one exists" → "PR number and URL when one exists — not
whether it is open, merged or closed... check the forge for that") and one Guardrails bullet (the old
"Report `gh` being unavailable... as PR state unknown" → a direct "Never call `gh`... " guardrail
reflecting the new no-network reality). I also dropped `Bash(gh:*)` from the `allowed-tools`
frontmatter key, since no step in the file invokes `gh` any more and leaving the grant in would be a
stale permission (Principle of Least Privilege, from `engineering-principles.md`).

I did **not** touch `skills/myflow-contracts/state-self-heal.md` — the contract there is written
generically ("a usable PR CLI for the host answered"), is not scoped to this task, and is not made
false by this change; it still correctly defines what the permitted correction *would* do if some
command could establish it conclusively.

## Ledger

`.superpowers/sdd/task-4-ledger.md` — six removed passages (the gh-probe step-2 item, the worktree
table column + its intro sentence + its trailing description sentence, the stale PR-cell example
value, the detail-view PR bullet, the Guardrails gh bullet, and the `Bash(gh:*)` grant), each mapped
`— (deleted, capability removed)` with the replacement sentence named in "Pointer left" where one
exists. A note below the table explains why the self-heal correction paragraph carries no row (edited
in place, not removed — Task 5's deletion, not this task's).

## Verification (all exit 0 / no matches, run after the edits)

```
grep -n 'gh pr list' skills/myflow-status/SKILL.md   → no matches (grep exit 1)
scripts/check-references.sh                          → "all referenced sections resolve", exit 0
scripts/check-vocabulary.sh                           → both guards clean, exit 0
scripts/check-contract-budget.sh                       → "BUDGET-OK: 22 contract file(s) within budget", exit 0
```

RED confirmation: `git diff --cached -- skills/myflow-status/SKILL.md` shows 2 removed lines
matching `gh pr list` (the deleted step-2 item, and the old self-heal-bullet citation), and 0 added
lines matching it — the grep guard's before/after state.

Read the full file end to end after editing (273 lines). No surviving sentence still promises
gh-probe-supplied information (open/merged/closed PR state, or a "PR state unknown" branch for a
probe that no longer runs). The one deliberately-inert paragraph (the self-heal correction) says so
of itself.

## Worktree hygiene

Serena was activated on this worktree path before any edit
(`/Users/tweety53/Projects/agents-worktrees/openspec-kan-95-slim-the-myflow-contract-files`).
`git -C /Users/tweety53/Projects/agents status --short` shows no tracked `M` entries (only `A` rows
for the new OpenSpec change proposal and one `??` for `.serena/`) — no edits leaked into the main
checkout.

## Concerns

None blocking. One judgment call worth flagging explicitly: I removed `Bash(gh:*)` from the
frontmatter `allowed-tools` and reworded three sentences beyond the two the brief names verbatim
(the detail-view PR bullet, one Guardrails bullet, and the frontmatter tool grant). All three were
strictly downstream of "remove the `gh pr list` probe" — each was a promise or permission that only
existed because of that probe — and none of them touch the worktree-column removal or the merge-status
step. If a reviewer would rather these three land as a separate follow-up rather than folded into
Task 4, they're easy to isolate in the diff (they're the frontmatter line, the detail-view bullet,
and the last Guardrails bullet before "Never guess a state...").

## Fix round — review Findings 1 and 2

Review came back with the spec check passing and two Important findings, both the same stale-promise
defect class this task was cut to remove:

**Finding 1 — frontmatter `description` (line 3).** It still listed "worktree" among the fields the
report shows. That was accurate before this task (the table had a `Worktree / branch` column); it is
false now for the no-argument table, since the worktree path survives only in the detail view.
**Fix:** dropped "worktree, " from the comma list. New text: "Show every open myflow change with its
pipeline state, PR, next command, and last update. Read-only. Use for /myflow-status." I did not add
a "(detail view only)" qualifier — the reviewer offered either fix, and the description is already a
short field list rather than a field-by-field explainer, so a bare drop keeps it accurate without
adding a footnote genre that doesn't exist anywhere else in the description.

**Finding 2 — Guardrails bullet (lines 254–257 pre-fix).** "Clearing a `prUrl` that a conclusive probe
disproved is the one permitted correction..." carried the identical dead precondition as the step-2
self-heal bullet I'd already caveated at lines 107–114 on the first pass, but this one sat in
Guardrails, uncaveated, right next to the bullet I *did* fix (the `gh`-guardrail at what was then
259–260) — exactly the juxtaposition that makes an uncaveated neighbor read as live guidance. **Fix:**
appended one sentence: "**This command can no longer apply that correction** — the precondition and
why are stated once under **2. Resolve each change's state** above; do not restate it here." Pointed
at the existing explanation rather than duplicating it a second time (Single Source of Truth) — the
brief's other option, restating the caveat verbatim, would have created exactly the kind of second
independently-maintained copy this whole change is trying to eliminate elsewhere in the corpus.

Neither fix touched anything else; no restructuring, no other sentence moved.

### Third-pass end-to-end re-read, hunting for a third stale promise

Read the file in full again after both fixes (274 lines) specifically looking for a sentence that was
true while the `gh pr list` probe existed and is false now. Checked every remaining mention of `PR`,
`prUrl`, `open`/`merged`/`closed`, `unknown`, `conclusive`, `item 5`, and the two already-caveated
paragraphs for drift. **Found no third instance.** The two closest look-alikes I checked and ruled
out:
- Line 224 ("`prUrl` is consulted only where **that probe** was inconclusive") and lines 241–244 (the
  one-way `prUrl` test) — both refer to the **merge-status** probe (step 2 item 4, local git,
  untouched by this task) and to `prUrl`-as-a-recorded-field, not to the removed network probe. Still
  true.
- The Guardrails bullet "Never guess a state when artifacts are ambiguous — show `?`..." — this is
  about pipeline-state ambiguity generally (self-heal), not PR state specifically; the PR-specific `?`
  case was already removed with the PR-column rewrite in the first pass. Still true.

### Re-verification (after the fix, from the worktree root)

```
scripts/check-references.sh        → "check-references: all referenced sections resolve", exit 0
scripts/check-vocabulary.sh         → both guards clean, exit 0
scripts/check-contract-budget.sh     → "BUDGET-OK: 22 contract file(s) within budget", exit 0
grep -n 'gh ' skills/myflow-status/SKILL.md   → no matches (grep exit 1)
```

### Ledger update

Added one row to `.superpowers/sdd/task-4-ledger.md` for the `worktree, ` phrase dropped from the
frontmatter `description` (Finding 1). Finding 2 added no row — like the step-2 self-heal bullet, it
was caveated in place, not removed; the ledger's explanatory section (now covering both paragraphs)
was updated accordingly, and still records that both paragraphs' actual deletion is Task 5's.

### Worktree hygiene (re-confirmed)

`git -C /Users/tweety53/Projects/agents status --short` shows no tracked `M` entries. Staged only
`skills/myflow-status/SKILL.md` in the worktree via `git add`; no commit made.

---

## task-5-report

# Task 5 report — Delete state self-heal

## Status: DONE_WITH_CONCERNS

## Summary

Deleted `skills/myflow-contracts/state-self-heal.md` (`git rm`) and stripped every citation of it
from the corpus. Re-homed the three rules `state-file.md` depended on (closed-schema rule,
`planningEffort`/`models` absent-key exception, the general "unparseable" definition they both rest
on) into `state-file.md` itself, using the brief's given definition as the floor and extending it
with the "parseable-but-incomplete is unparseable in full" rule I found on my own read. Removed
self-heal from `/myflow-status` (mechanism, `⚠` legend, guardrails), rewrote the "Resolving a
change's worktrees" reason in `pipeline.md` and `finish-contract.md`, and stripped the remaining
citations from `pipeline-rationale.md`, `skills/myflow-contracts/SKILL.md`,
`rules/myflow-manual-review.mdc`, `CLAUDE.md`, `AGENTS.md`, `skills/README.md`, and the budget row in
`scripts/check-contract-budget.sh`. `scripts/test-check-contract-budget.sh` had no fixture case
naming the path — nothing to remove there.

Full per-move ledger (24 rows for `state-self-heal.md`'s own content, plus a mechanical-repairs
section for the rest of the diff): `.superpowers/sdd/task-5-ledger.md`.

## Beyond the brief's named file list

A repo-wide case-insensitive sweep for "self-heal" (run after finishing the named files, specifically
to catch what a literal `state-self-heal\|State self-heal` grep would miss on case alone) turned up
three more citations not in the brief's Files list, all now fixed:

- `commands/myflow-status.md` and `commands-claude/myflow-status.md` — "Its one write is state
  self-heal: correcting a stale cache..." → rewritten to state the command is entirely read-only.
- `skills/myflow-start/SKILL-rationale.md` — a clause explaining why `artifactUrl` may be `null` on a
  state file blamed it on "self-heal rebuilds... and self-heal names it among unrecovered fields."
  Rewritten to the actual, simpler, self-heal-independent reason: the field is nullable and a
  hand-edited or incompletely-written file can carry it that way.

Without these three, the required final grep would have failed (they don't match the case-sensitive
`state-self-heal\|State self-heal` pattern at all, so even a narrower literal-only pass would have
missed them — I found them only by broadening to case-insensitive).

## Judgment calls

**The unparseable definition's placement and scope.** Placed the brief's given paragraph right after
`state-file.md`'s JSON schema block, before the field-by-field list, so both the `models`/
`planningEffort` exception (mid-list) and the five-retired-fields paragraph (near the end) can cite
it as "the closed-schema rule stated above" without a second cross-file citation. Extended it with
one sentence beyond the brief's floor: "JSON that parses but is missing one or more of the fields
this contract requires is unparseable in full on that account alone, not partially recovered" — this
was self-heal's own separate rule (its "JSON that parses but is missing one or more of the fields...
unparseable in full, not partially recovered" paragraph), not covered by the brief's example sentence
literally, and losing it would have quietly weakened the closed-schema rule to "missing means null"
rather than "missing means unparseable."

**The null-merge-base citation (`state-file.md`, `worktrees` field bullet).** Not one of the brief's
three named citations, but it cited self-heal for *how* a `null` merge base arises (self-heal's
rebuild-from-`git worktree list` path, which reports no merge base). Since nothing performs that
rebuild any more, I reworded the explanation to the reason that still holds — a hand-edited or
out-of-band-modified file — and kept the rule itself (null is legal, every "missing" rule applies to
it unchanged) exactly as it was; `finish-contract.md`'s preflight, `handoff-blocks.md`, and
`myflow-status/SKILL.md`'s merge-status check all still treat "no recorded merge base" as a live,
expected case independent of self-heal, so the rule was never actually self-heal-dependent — only its
stated *reason* was.

**The monotonicity carve-out (`state-file.md`, "State writes are monotonic" paragraph).** Not named
by the brief either. Since the dispatch's own context note says the `prUrl` correction is deleted
*outright* (not re-homed) from `/myflow-status`, and nothing else in the corpus ever performed it, the
carve-out sentence in `state-file.md` ("No command may write a state earlier than the one it found.
The single carve-out is described under State self-heal...") now describes a mechanism that exists
nowhere. Deleted the carve-out sentence, kept the monotonicity rule itself.

**The self-referential rationale paragraph (`state-file.md`, "This paragraph is the only statement of
that reasoning... State self-heal applies the rule and cites this one...").** This existed only to
explain *why* self-heal cites `state-file.md` instead of re-arguing the point. With self-heal gone,
there's no second party left to cite it, so the paragraph's entire reason for existing is gone.
Deleted it; kept the substantive conclusion one paragraph up ("unmapped value reads as *not
recorded*, never unparseable") unchanged, and simplified its own internal reasoning to stop
referencing "routes the file to self-heal" as a live possibility (rewrote to "no command infers or
rebuilds a state file's contents").

**The retired `stage`-field paragraphs (self-heal's "no legacy-value migration for `stage`" and its
follow-up contrasting `stage` with `effort`).** Checked whether any surviving file depends on this
specific claim. It doesn't: the general closed-schema rule (any undocumented key is unparseable)
already implies the same outcome for `stage` as for any other retired key with no exception, and
nothing in `skills/`, `rules/`, `commands*/`, `CLAUDE.md` or `AGENTS.md` cites the `stage` paragraph
by name. The only other mention of `stage` as a field name in the whole repo is
`openspec/specs/myflow-state-machine/spec.md:210` — a **pre-existing baseline spec this change's own
delta specs do not touch**, and `openspec/` is off-limits to this task. Classified as deleted with
the mechanism; no rule lost. Detailed in the ledger's closing note.

**The stale-`prUrl` gap section.** Nothing cites it; its point (a stale `null` `prUrl` while a PR
exists is undetected) is now a special case of the general cost the `myflow-state-integrity` spec's
Migration note already states explicitly ("never corrected by anything, and never flagged"). Deleted
with the mechanism rather than relocating the operator-workaround advice, since nothing in the brief
or the corpus asked for a new home for it and inventing one would be authoring fresh guidance outside
this task's scope.

**`skills/myflow-status/SKILL.md`'s section-2 item numbering.** Deleting items 1-3 (worktree
resolution, `tasks.md` check, manual-test check — all self-heal-validation-only, each scoped "for the
rest of *this step*" in the original text) left item 4 (merge status) as the sole item. Converted it
from a numbered list entry to a named "**Merge status**" paragraph rather than renumbering it to "1",
and retargeted every cross-reference ("step 2 item 4", "that item names") to name the check instead
of a number — renumbering to "1" would have made "item 1" ambiguous with the *nested* three-step list
inside the merge-status check itself.

## Verification (all from the worktree root, after the edits)

```
scripts/check-references.sh          → "check-references: all referenced sections resolve", exit 0
scripts/check-vocabulary.sh          → both guards clean, exit 0
scripts/check-contract-budget.sh     → "BUDGET-OK: 21 contract file(s) within budget", exit 0
scripts/test-check-contract-budget.sh → "all checks passed" (21 sub-checks), exit 0
scripts/test-setup.sh                → "✓ PASS — 192 assertions, 0 failures", exit 0
```

Also ran, not required but covered by this repo's full lint list per `.myflow/project.md`:
`scripts/check-plan-provenance.sh` (clean), `scripts/check-task-build-green.sh` (clean),
`scripts/check-workspace-isolation.sh` (clean — no `## workspace isolation` declared).

RED confirmation: before any edit, `scripts/check-references.sh`/`check-vocabulary.sh`/
`check-contract-budget.sh` were run and were already green (nothing was broken yet); the RED signal
for this task is the required grep itself, which found the file's citations before deletion and finds
only the permitted exception after (see below).

### The required grep

```
grep -rn 'state-self-heal\|State self-heal' . --exclude-dir=.git --exclude-dir=docs --exclude-dir=archive --exclude-dir=.superpowers
```

Result: every match is under `openspec/changes/kan-95-slim-the-myflow-contract-files/` or
`openspec/specs/myflow-state-integrity/spec.md` (both permitted), **except one**:

```
openspec/specs/myflow-contract-distribution/spec.md:184:SHALL the four contract sections (State file, State self-heal, Project configuration, Jira
```

This is the reason for **DONE_WITH_CONCERNS** rather than **DONE**.

## The one open concern

`openspec/specs/myflow-contract-distribution/spec.md` is the **live baseline spec** (pre-dates this
change; not one of `kan-95`'s own delta files). Its "Always-on rule layer carries only the trigger
and the pointers" requirement enumerates "the four contract sections (State file, State self-heal,
Project configuration, Jira integration)" as the sections `rules/myflow-manual-review.mdc` must not
inline. I checked `openspec/changes/kan-95-slim-the-myflow-contract-files/specs/myflow-contract-distribution/spec.md`
(this change's own delta for that capability) and confirmed it does **not** touch that requirement —
it only modifies two unrelated requirements ("The pipeline diagram..." and "Extracted contracts live
in a dedicated on-demand skill"). So this baseline line is untouched by this change's planned deltas
and will still read "four contract sections... State self-heal" even after `/myflow-finish` syncs
`kan-95`'s deltas into the baseline.

I did **not** edit it. The dispatch says, without qualification, "do not edit anything under
`openspec/`," and that instruction is more specific and more absolute than the verification grep's
stated allow-list. Between a hard "never touch this directory" rule and a verification command that
turns out to assume a slightly narrower scope than the actual corpus, I followed the rule.

This looks like a gap in the change's own planning (the delta for `myflow-contract-distribution`
should probably also update that one requirement's four-item list to three), not something Task 5 can
fix without violating its own boundary. Flagging it explicitly rather than silently accepting a
failing grep or silently editing `openspec/` to make it pass.

## Worktree hygiene

Serena was activated on this worktree
(`/Users/tweety53/Projects/agents-worktrees/openspec-kan-95-slim-the-myflow-contract-files`) before
any edit. `git -C /Users/tweety53/Projects/agents status --short` shows the same pre-existing `A`/`??`
rows from earlier tasks (`openspec/changes/kan-95-.../`, `docs/superpowers/specs/...`,
`.serena/`) and **no `M` entries** — nothing leaked into the main checkout from this task.

## Staging

```
git add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
```

No commit made. `openspec/`, `docs/manual-test/` and `docs/superpowers/` were never staged by this
task (confirmed via `git status --short` after staging — only the pre-existing untracked entries
under those paths remain, carried from earlier tasks, unstaged by me).

## Fix round — review finding: the worktree-scan reason was false

Review (and the coordinator, independently) found that my Part-B rewrite of the
`Resolving a change's worktrees` placement reason — "`/myflow-do` and `/myflow-finish` both need it,
and only `/myflow-finish` loads `finish-contract.md`" — was false. `/myflow-do` never cites that
section and never runs the `git worktree list --porcelain` scan: it creates its worktree through
`superpowers:using-git-worktrees` and reads paths from the state file's `worktrees` map, same as
`/myflow-status`. With self-heal gone, `/myflow-finish` is the section's **only** consumer, and it
already loads `finish-contract.md` — which is exactly the shape
**A section reachable from only one command lives in its own file**
(`openspec/specs/myflow-contract-economy/spec.md`) governs: such a section moves whole, rules and
reasoning together, into the file of its one reader, rather than staying in the shared core with a
reworded justification.

**What I did:**

1. Moved `## Resolving a change's worktrees` out of `skills/myflow-contracts/pipeline.md` in full —
   the description sentence, the `worktrees`-map/absent-map scan and its bash snippet, the
   "never guess a path" paragraph, and the `substr`-vs-`$2` paragraph — verbatim, into
   `skills/myflow-contracts/finish-contract.md` as a new `### Resolving a change's worktrees`
   subsection placed immediately before `### Worktree cleanup` (both under `## Finish contract`).
2. **Deleted the false reason sentence rather than rewriting it again.** Per the coordinator's
   explicit instruction, no replacement reason was invented — in its new home the section needs no
   placement justification, since it now sits in the contract of the only command that reads it.
3. Left no stub in `pipeline.md` — the heading and all its content are gone, nothing points into
   `pipeline.md` for this section any more (same shape as Task 1's `Pipeline flow` removal, not Task
   2's stub pattern).
4. Repointed `### Worktree cleanup`'s citation from a cross-file one
   (`**Resolving a change's worktrees** (\`skills/myflow-contracts/pipeline.md\`) — hoisted into the
   core because...`) to a same-file reference (`**Resolving a change's worktrees** above`), dropping
   the reason clause entirely since the two sections now share a file and need no cross-file
   justification at all.
5. Updated `skills/myflow-contracts/SKILL.md`'s index table: removed "and resolving a change's
   worktrees" from the `pipeline.md` row's description, added "resolving a change's worktrees, and"
   to the `finish-contract.md` row's description — the index must describe what each file actually
   holds after the move.
6. Added a Part C to `.superpowers/sdd/task-5-ledger.md` with the moved passages, the deleted false
   reason, and the two permitted edits (bold-token+path repointing on the `Worktree cleanup`
   citation; no stale position word existed to delete in the moved passage itself).

**What I did not do:** touch `openspec/` (the coordinator's own `myflow-state-machine` delta and the
`myflow-contract-distribution` gap it found are both theirs, confirmed present in the worktree as `A`
rows, not edited by me).

### Re-verification (from the worktree root, after the move)

```
scripts/check-references.sh          → "check-references: all referenced sections resolve", exit 0
scripts/check-vocabulary.sh          → both guards clean, exit 0
scripts/check-contract-budget.sh     → "BUDGET-OK: 21 contract file(s) within budget", exit 0
scripts/test-check-contract-budget.sh → "all checks passed" (21 sub-checks), exit 0
scripts/test-setup.sh                → "✓ PASS — 192 assertions, 0 failures", exit 0
```

```
grep -rn "Resolving a change's worktrees" . --exclude-dir=.git --exclude-dir=docs --exclude-dir=archive --exclude-dir=.superpowers
```

Result: only `skills/myflow-contracts/finish-contract.md` (the new heading, plus the one same-file
reference from `Worktree cleanup`) outside `openspec/` — no citation left pointing into `pipeline.md`.

**`scripts/check-cleanup-complete.sh` agreement, re-confirmed.** Its own worktree-discovery logic
(`git -C "$REPO" worktree list --porcelain`, then
`awk -v b="refs/heads/openspec/$NAME" '/^worktree / { w = substr($0, 10) }'`) still matches the moved
snippet exactly — same `worktree list --porcelain` invocation, same `substr($0, 10)` offset for the
path, same reliance on `$2` only for the branch line. The move relocated the snippet without touching
its content, so nothing to reconcile.

**Byte budgets after the move:** `pipeline.md` 36,395 bytes (budget 80,876), `finish-contract.md`
26,786 bytes (budget 32,037) — both comfortably inside their ratchets; `check-contract-budget.sh`
confirms `BUDGET-OK`.

**Worktree hygiene, re-confirmed.** `git -C /Users/tweety53/Projects/agents status --short` still
shows no `M` entries — only pre-existing `A`/`??` rows, now including the coordinator's own
`openspec/changes/kan-95-slim-the-myflow-contract-files/specs/myflow-state-machine/spec.md` (their
work, not mine). Restaged with the same exclusion pathspec; `openspec/`, `docs/manual-test/` and
`docs/superpowers/` remain untouched by this task.

### Status: DONE

The one open concern from the first pass (the `myflow-contract-distribution/spec.md:184` baseline
gap) is now the coordinator's own finding, already addressed on their side via the new
`myflow-state-machine` delta — nothing left for this task to flag. The fix-round finding (the false
worktree-scan reason) is resolved by the move, not a reword. All required verification, including the
two `grep`s and the `check-cleanup-complete.sh` agreement check, is clean.

---

## task-6-report

# Task 6 report — Evict rationale from `pipeline.md`

**Status: DONE**

## Summary

Read `pipeline.md` end to end (36,395 bytes) after confirming Tasks 1–5 had already removed
`## Pipeline flow`, the block templates, `## State self-heal`, and `## Resolving a change's
worktrees` (moved to `finish-contract.md` by Task 5 — its own ledger records this, and its "argument
for `substr` over `$2`" content is therefore out of scope here, exactly as the brief's own
context note anticipated).

Classified every remaining section (States, Command surface, State transitions, Wrong state for
this command, Git boundaries, Preserving the session records, Progress visibility, Handoff output,
IntelliJ commands, Temporary artifacts registry, State file, Project configuration, Jira
integration, Model policy, Change name resolution) paragraph by paragraph against the core/rationale
test. Nine sections (States, Command surface, State transitions, Wrong state for this command,
IntelliJ commands, Finish contract, State file, Project configuration, Jira integration) came back
wholly core — every paragraph in them either states a rule directly or is too short/tightly bound to
its rule to separate without losing the rule. Verified this rather than assuming it, including
checking whether phrases quoted verbatim elsewhere (`skills/README.md`, `rules/myflow-manual-
review.mdc`, `CLAUDE.md`, `skills/myflow-do/SKILL.md`) depended on the exact wording staying put.

The other six sections (Git boundaries, Preserving the session records, Progress visibility, Handoff
output, Temporary artifacts registry, Model policy — the two the brief flagged as richest) each
carried genuine mixed passages: a rule fused with justification, alternatives-considered, or history.
Applied rule extraction 22 times: moved each original passage to `pipeline-rationale.md` under its
mirrored heading, byte-for-byte, and left one new core sentence stating the rule and citing the
appendix heading. One passage (a "why the second source" justification inside Change name
resolution) was pure argument with no rule of its own — its flanking rule-sentences were already
present and untouched — so it moved as a plain wholly-rationale passage with a bare pointer, not a
rule-extraction row. Full 22-row ledger, with every new core sentence quoted, is at
`.superpowers/sdd/task-6-ledger.md`.

Also verified and named the classification-candidate list in the brief itself rather than trusting
it blind: confirmed the `&&`-chain-vs-`set -e` argument, the "do not harmonise the two orderings"
paragraph, and the no-third-checkbox-marker paragraph as real extraction candidates (all three
became ledger rows), and found several more the brief's list didn't name (the Model-policy override
paragraphs, the Temporary-artifacts-registry cleanup/cache-index paragraphs, the Handoff-output
next-command and never-stages-planning-paths bullets, the tab-command colour/timing paragraph).

## The two "Wrong state for this command" headings

Pre-existing, not a defect this task introduced or should fix. `pipeline.md` has exactly one real
`## Wrong state for this command` section (with its own rule prose). Immediately after it sits a
fenced code block that is the literal handoff-block *template* this section documents — and that
template's first line is itself the text `## Wrong state for this command`, because the block it
shows is what the command prints, heading and all. `grep -n "^#"` matches both because grep is not
fence-aware; `scripts/check-references.sh`'s own `strip_fenced_lines` helper (and any real markdown
renderer) is, and correctly sees one heading. Not a split-by-formatting-accident case and not two
sections saying different things — there is nothing to merge, and touching either occurrence would
either delete the real heading or corrupt the example template. Left both alone; documented in the
ledger's Notes section per the brief's explicit instruction not to resolve this silently.

## Verification

All four required commands run before and after, all exit 0 after:

```
scripts/check-references.sh     → check-references: all referenced sections resolve
scripts/check-vocabulary.sh     → Stage-vocabulary guard: clean; Panel-vocabulary guard: clean
scripts/check-contract-budget.sh → BUDGET-OK: 21 contract file(s) within budget
scripts/test-setup.sh           → PASS — 192 assertions, 0 failures
```

`scripts/test-check-contract-budget.sh` also re-run after the budget-row edit below (22 checks, all
pass) since that script exercises `check-contract-budget.sh`'s own logic.

**One guard failure hit and fixed, not weakened.** `check-contract-budget.sh` failed once, mid-task:
`pipeline-rationale.md` grew past its existing budget row (20,731 bytes — set before this change, per
that guard's own doc comment, as "the size the file actually had... plus 25%"). The guard's own error
message names the sanctioned fix for deliberate growth: raise the row. `pipeline-rationale.md`
growing is this task's entire mechanism, not a regression, and Task 10 ("Re-anchor every budget row")
is explicitly the later task that recomputes every row precisely across the whole plan — so I raised
just this one row to `floor(27615 × 1.25) = 34518`, the same formula the guard's header and Task 10's
own brief both specify, and left every other row untouched for Task 10 to finish. No rule was
weakened, no suppression added, no scope narrowed.

**Two stale `above`/`below` position words found and deleted (never substituted), the one edit the
partition rule allows beyond citation-repointing.** Both were introduced by my own moves, not
pre-existing: (1) a Model-policy passage said `unknown (agent-defined)` "stays exactly as written
above," where "above" pointed at two mentions that stayed in the core, not in the appendix with it —
deleted "above". (2) a "which file to change first" passage said a requirement "anchors the defaults
below," where "the defaults" (the Opus/Sonnet table) stayed in the core — deleted "below". Checked one
more `above`-bearing moved passage (the Temporary-artifacts-registry "one paragraph above" sentence)
and confirmed it was already loose/rhetorical in the *original* file — three paragraphs, not one,
separated the two spots in the source — so the move didn't break anything a rule requires fixing;
left it untouched rather than "fixing" a pre-existing looseness the partition rule doesn't cover.

**End-to-end read of the new `pipeline.md`, core-only.** Read the whole 31,151-byte file straight
through pretending the appendix does not exist. Every section still states its own rule directly —
tables, code blocks, and prohibitions are all unchanged in place, and every eviction left exactly one
pointer sentence naming the rule already stated beside it and citing where the argument for it now
lives. No section required a trip to `pipeline-rationale.md` to know what to do; the appendix is
needed only to know *why*.

## Byte counts

`pipeline.md`: **36,395 → 31,151 bytes** (−5,244, ~14%).
`pipeline-rationale.md`: grew to 27,591 bytes final, as intended — a `/myflow-*` run never loads it,
so its growth is not part of the per-run budget this task exists to cut.

This lands well above the brief's ~23,000 projection (and above the 28,000 example the parent task
named as still acceptable). I did not chase the number — Model policy (8,548 → 6,470B, −24%),
Temporary artifacts registry (5,571 → 4,666B, −16%), Git boundaries (3,990 → 3,176B, −20%), and
Preserving the session records (2,184 → 1,378B, −37%) all took real, verified extractions in the
sections the brief itself named "where most of the win is." The remaining bulk sits in nine sections
that I read closely and found to be honestly core: they carry rule tables, ordered steps, and
prohibitions with no separable argument, several of them phrases other files (`skills/README.md`,
`rules/myflow-manual-review.mdc`, `CLAUDE.md`, `skills/myflow-do/SKILL.md`) quote or depend on
verbatim. The brief is explicit that this is the correct outcome when classification is honest
("If honest classification lands at 28,000, that is the right answer and you report it") — this
result is 3,000 bytes past that named example, for the reason above.

## Concerns for the reviewing panel

1. **Budget row raised mid-plan** (`scripts/check-contract-budget.sh`, `pipeline-rationale.md` row:
   20,731 → 34,518). Sanctioned by the guard's own message and consistent with Task 10's own
   documented formula, but it is a change outside this task's two named files
   (`pipeline.md`, `pipeline-rationale.md`) — flagging explicitly rather than burying it, per Task
   5's own precedent of surfacing out-of-brief edits for the panel.
2. **Final size (31,151B) is well above the ~23,000 projection.** Documented above with the specific
   sections and percentages; every extraction is in the ledger for the panel to check against the
   diff in both directions, per the ledger requirement's own review-panel instruction.
3. `git add -A` (per the prescribed staging command) also staged `.serena/.gitignore` and
   `.serena/project.yml` — Serena MCP project-registration files created when I activated the
   project in this worktree per the caller's instruction. These are not part of this task's
   intended diff; flagging so the panel/finish step can decide whether to keep or drop them (main
   checkout shows the same `.serena/` as untracked and pre-existing, so this is not new drift I
   introduced into the corpus, just a side effect of the prescribed `add -A`).

## Files touched

- `/Users/tweety53/Projects/agents-worktrees/openspec-kan-95-slim-the-myflow-contract-files/skills/myflow-contracts/pipeline.md`
- `/Users/tweety53/Projects/agents-worktrees/openspec-kan-95-slim-the-myflow-contract-files/skills/myflow-contracts/pipeline-rationale.md`
- `/Users/tweety53/Projects/agents-worktrees/openspec-kan-95-slim-the-myflow-contract-files/scripts/check-contract-budget.sh` (one row raised; see Concerns)
- Ledger: `/Users/tweety53/Projects/agents-worktrees/openspec-kan-95-slim-the-myflow-contract-files/.superpowers/sdd/task-6-ledger.md`

Verified `git -C /Users/tweety53/Projects/agents status --short` shows no tracked `M` entries — the
prior task's stray main-checkout edits are not present; only new/untracked `openspec/` and
`docs/superpowers/` planning artifacts, which are normal and unrelated to this task.

Staged with the prescribed command:
`git -C "$WORKTREE" add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'`

---

## task-7-report

# Task 7 report — Evict rationale from `skills/myflow-do/SKILL.md`

**Status: DONE**

## Summary

Activated Serena on this worktree first (per the caller's instruction, since an earlier task had
written into the main checkout by mistake) and confirmed `git -C /Users/tweety53/Projects/agents
status --short` shows no tracked `M` entries after this task's edits — nothing leaked outside the
worktree.

Read `skills/myflow-do/SKILL.md` (37,900 bytes) and `skills/myflow-do/SKILL-rationale.md` (3,333
bytes) end to end. Measured the three largest sections with an `awk` byte tally, matching the
brief's figures closely (`## 5. The review panel` 11,674B, `## 7. Verify, stage, and hand off`
9,666B, `## 6. Write the manual test guide` 4,999B — the small deltas from the brief's numbers are
Task 3's fold, as the brief anticipated).

Classified every section paragraph-by-paragraph, bullet-by-bullet against the core/rationale test.
This file is markedly denser than `pipeline.md` (Task 6's target): almost every paragraph is either
wholly a rule (dispatch instructions, exit-code tables, marker-format rules, staging commands — all
named in the brief as must-stay) or a rule with its argument welded into the same sentence, with very
little standalone argument-only prose. Found and applied ten rule extractions, all in `## 2`, `## 5`,
`## 6` and `## 7` — the sections with any separable argument at all. Read `openspec/changes/kan-95-
slim-the-myflow-contract-files/specs/myflow-contract-economy/spec.md`'s amended "verbatim partition"
requirement before starting and followed its exact mechanics: the original passage (rule sentence(s)
plus argument) moves to the appendix byte-for-byte, unchanged; the core keeps its rule sentence(s) —
reusing the original wording verbatim where possible, since the rule itself was never the part being
removed — plus one new pointer sentence citing the appendix heading, and states no argument. This
matches the pattern already established in `pipeline.md`/`pipeline-rationale.md` (checked as a
precedent before writing the first edit): the bolded rule-lead sentence legitimately appears in both
files, once as the terse core statement, once as the lead-in to the fuller appendix passage.

Full 10-row ledger, with every new core sentence quoted and two rejected/considered-and-declined
candidates recorded, is at `.superpowers/sdd/task-7-ledger.md`.

**One transcription mistake caught and fixed before finishing.** My first draft of two appendix
insertions (the `[PRINCIPLES_PATH]` "why absolute" passage and the `create`-command consequence
passage) added connective words ("to `engineering-principles.md`", "— which is why this step does
not call the project's `create` command") that are not in the original `SKILL.md` text — a violation
of the byte-for-byte-unchanged rule. Caught this on review before running the guards, and replaced
both with the exact original wording, no rewording.

## What stayed core, and why the yield is small

Sections read and found **wholly core** (listed in full, with reasons, in the ledger): the
frontmatter/intro, `## State gate`, `## Superpowers Basic Workflow`, `## 1`, `## 3`, most of `## 4`
(the four required implementer dispatch blocks in particular — explicitly named in the brief as
must-stay, and each is a blockquote of pure instruction with no argument sentence anywhere in any of
the four), the slot roster table and most of `## 5`, the behaviour-checklist rules and most of `## 6`,
and the guard-invocation table, staging commands, exclusion-pathspec note, one-commit-exception
procedure, and handback prompt in `## 7`. `## Guardrails` is a bare prohibition list with nothing to
extract.

Two candidates were **considered and rejected**, not silently skipped — both recorded in the ledger's
Notes section: a one-line disambiguation in `## 4` ("the two rules differ on purpose") where
extracting it would have required duplicating a large surrounding paragraph in the appendix for a
~70-byte gain, and the three-paragraph "Every path is absolute" / "One guide can carry both" bullet in
`## 6`, which is genuinely dense with interleaved rules and citations with no clean paragraph-level
rule/argument boundary that wouldn't risk losing an operative clause (e.g. the "resolve from the
workspace id rather than a later-exported variable" ordering rule, which reads as argument but is
actually load-bearing procedural sequencing).

## Verification

All four required commands run before and after this task's edits; all exit 0 after:

```
scripts/check-references.sh      → check-references: all referenced sections resolve
scripts/check-vocabulary.sh      → Stage-vocabulary guard: clean; Panel-vocabulary guard: clean
scripts/check-contract-budget.sh → BUDGET-OK: 21 contract file(s) within budget
scripts/test-setup.sh            → PASS — 192 assertions, 0 failures
```

**One guard failure hit and fixed, not weakened.** `check-contract-budget.sh` failed once, mid-task:
`skills/myflow-do/SKILL-rationale.md` grew past its existing budget row (4,158 bytes) once the ten
extractions landed. Per the guard's own sanctioned fix (raise the row using `size × 1.25`) and per
the brief's explicit sanction for this exact situation, raised the row to
`floor(6,696 × 1.25) = 8,370`. `skills/myflow-do/SKILL.md`'s own row (47,356) was never at risk — the
file shrank, it did not grow. No rule was weakened, no suppression added, no scope narrowed.

**End-to-end read of the new `SKILL.md`, core-only.** Read the whole 36,411-byte file straight
through, pretending the appendix does not exist, as if I were the agent about to run `/myflow-do`.
Every one of the ten eviction points still states a complete, actionable rule at its location — the
"why" is what moved, never the "what to do." Confirmed I could execute every step correctly without
opening `SKILL-rationale.md`: compute-the-workspace-id-once is still a complete instruction, the
review panel's model-selection and marker-format rules are all still fully stated (only the
justification and two incident anecdotes moved), the manual-test-guide instruction to write/refresh
the file is unchanged, and the `create`-command and export-every-row rules in section 7 are both
still fully actionable. Found no place where a rule left with its argument.

## Byte counts

`skills/myflow-do/SKILL.md`: **37,900 → 36,411 bytes** (−1,489, ~3.9%).
`skills/myflow-do/SKILL-rationale.md`: 3,333 → 6,696 bytes, as intended — a `/myflow-*` run never
loads it, so its growth is not part of the per-run budget this task exists to cut.

This lands far short of the brief's ~26,000-byte projection, and I want to be direct about why rather
than let the number speak for itself: this file simply does not carry much separable argument. Unlike
`pipeline.md` (Task 6, 36,395 → 31,151B, ~14%), which had six sections dense with alternatives-
considered and incident history, `skills/myflow-do/SKILL.md` is almost entirely dispatch
instructions, exit-code tables, marker-format enforcement rules, and staging commands — the brief's
own list of "what is core in this file and must not move" covers most of its bulk by name. I went
through every section paragraph-by-paragraph (documented in the ledger) rather than stopping once the
obvious extractions were done, and found ten real rule-extraction candidates, all in the four
sections that have any argument at all (`## 2`, `## 5`, `## 6`, `## 7`); the other seven sections
came back genuinely clean. I did not stop early to protect the number, and I did not force marginal
extractions to chase it either — the two rejected candidates above are exactly that judgment call
made explicit.

## Staging

Staged per the brief's Step 4 pathspec:

```bash
git -C "$WORKTREE" add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
```

`git status --short` after staging showed no `.serena/` entries this time (Serena's project registry
for this worktree already existed from an earlier task's activation, so this run's
`activate_project` call created nothing new to stage) — confirmed, no `git reset` needed.

## Concerns for the reviewing panel

1. **Yield is small and I'm reporting it honestly rather than reaching for more.** 3.9%, far under
   the ~26,000-byte projection. Every one of the ten extractions is in the ledger for the panel to
   check against the diff in both directions; the "wholly core, unchanged" list in the ledger names
   every section I read and decided against touching, with a one-line reason each.
2. **`scripts/check-contract-budget.sh`'s row for `skills/myflow-do/SKILL-rationale.md` was raised**
   (4,158 → 8,370), sanctioned by the guard's own message and the brief's explicit permission, but
   flagging explicitly since it's a change outside the two files this task's brief names directly.
3. One transcription slip (documented above, in "Summary") added un-sourced words to two appendix
   passages on the first pass; caught and corrected before running any guard, but flagging so the
   panel checks those two rows (`[PRINCIPLES_PATH]` / `create`-command) particularly closely against
   the diff for byte-for-byte fidelity.

## Round 2 — review findings addressed

The panel's review of this task found one Critical and three Important findings, plus named a fourth
(unpressed) candidate. All four are now addressed. Verification re-run in full (all four commands
below), all exit 0.

**Critical — ledger row 2's appendix copy had drifted wrapping, not drifted words.**
`skills/myflow-do/SKILL-rationale.md:35-37` (the "There is no parent-model inheritance…" passage)
carried the same words as the original but re-wrapped at different line breaks: the original (at
`git show 4dd7953201d0f48bcd7086cf9b637f388f0ef51d:skills/myflow-do/SKILL.md`, lines 161-164) breaks
after "no", "the" and "rather"; my copy broke after "which" and "one". This happened because I typed
the paragraph out by hand from what I'd read, rather than copying the exact substring — words
survived, line breaks didn't. No citation was repointed in that passage and no `above`/`below` was
deleted, so neither of the two permitted edits covers a rewrap here; **The split is a verbatim
partition, with citation repointing as its only edit** only allows a rewrap when a citation's length
change would otherwise leave an over-long line, confined to that one paragraph — not what happened.
Fixed by regenerating the passage directly from `/tmp/t7-before.md` (the pre-task image at that
commit) rather than hand-rewrapping, so the fix itself carries no hand-typed risk either.

**Independent re-check of the other nine Round-1 rows, done rather than trusted.** Wrote a script
(`/tmp/verify_rows2.py`, then folded into `/tmp/verify_all.py` below) that extracts the exact
byte range for every row from `/tmp/t7-before.md` and diffs it character-for-character against the
corresponding span in `skills/myflow-do/SKILL-rationale.md`. All nine other Round-1 rows matched
byte-for-byte on the first check — no further wrapping drift. The other nine rows were whole-paragraph
or whole-bullet copy-pastes done via the Edit tool's exact-match mechanics (old_string/new_string
against text already read verbatim from the file), which is presumably why only the one row typed
from memory (rather than copied) drifted.

**Three Important findings — argument left in the core — evicted, same rules as Round 1:**

1. `## 5. The review panel § Optional slot selection`: "A reviewer too many costs tokens; one too
   few costs a defect." (cost-benefit justification for the "Borderline → ask, include as default"
   rule stated in the sentence before it) moved to the mirrored `### Optional slot selection` heading
   in the appendix; core keeps the rule plus a new pointer.
2. `## 7. Verify, stage, and hand off`: "The two reasons are one case, deliberately: both end with
   nothing having run the guard, and a session that recognised only 'the repository does not carry
   it' would answer a failed resolution by doing nothing at all and saying nothing about it." (why
   the two script-not-found cases are merged; not itself actionable) moved to the appendix. The core
   keeps the operative sentence before it ("say in the handoff that the validation was performed
   manually and why the script was not run") and the operative sentence after it ("It is never
   skipped for want of the script…"), with the pointer inserted between them, replacing exactly the
   argument sentence and nothing else.
3. `## 7. Verify, stage, and hand off` (blockquote): "The exclusion is what keeps them out of the
   diff, rather than a filter applied when the diff is displayed… so nothing is lost by leaving them
   unstaged here." (design-choice defence for "Those three paths are never staged," already stated as
   the block's first sentence) moved to the appendix, preserving the `> ` blockquote marker on every
   line that carried one in the original — the first line of the moved span did not carry `> ` in the
   original (it followed directly after the bolded lead sentence on the same physical line), so it
   does not carry one in the appendix either; only lines 2-5, which were fresh physical lines in the
   source, keep their `> `.

**Fourth eviction — the "worktree ports" sentence, taken as named:** inside the "Every path is
absolute" bullet in `## 6`, "A worktree's applications bind their own ports, so the documented URL an
operator opens out of habit reaches whichever workspace holds the default port — a different change's
application, answering plausibly and about the wrong work." moved to the appendix, including the
original's 2-space list-continuation indent on its wrapped lines (preserved rather than stripped, to
stay strictly byte-for-byte — this differs from how I'd reasoned about indentation in Round 1's
rejected candidates, and I'm treating the stricter reading as correct given the Critical finding
above). The load-bearing ordering rule around it — "Resolve each URL from this worktree's workspace
id, the way section 2 computed it" — was left untouched in the core, exactly as instructed.

**All four evictions verified byte-for-byte** against `/tmp/t7-before.md` with the same
`/tmp/verify_all.py` script extended to cover them (checked twice — the first pass caught two
self-introduced slicing errors in the verification script itself, not in the content: a missing
2-space indent on the ports sentence, and the blockquote's first line wrongly carrying a `> ` prefix
in my first attempt at that insertion, caught by the same script before this report was written).
Final run: all 14 rows (10 from Round 1, 4 from Round 2) MATCH.

**Byte counts after Round 2:** `skills/myflow-do/SKILL.md` **37,900 → 35,825** (−2,075, ~5.5%, up
from Round 1's ~3.9%). `skills/myflow-do/SKILL-rationale.md`: 3,333 → 7,732 bytes — still within the
budget row raised in Round 1 (8,370), so no further budget-row change was needed this round.

**Final core-only read-through, repeated.** Read all four new eviction points in place: the
"Borderline → ask" rule, the "Every path is absolute" / "Every URL is the one this worktree
resolved" / "Resolve each URL from this worktree's workspace id" sequence (now with the pointer
inserted between "declared base." and "Resolve each URL", reads grammatically and loses no
instruction), the script-not-found paragraph (reads as a complete instruction with the pointer
sitting between two still-operative sentences), and "Those three paths are never staged" followed
immediately by the literal git commands above it, which were always the actual operative content.
Confirmed all four are still fully actionable with no trip to the appendix required.

**Verification re-run, all exit 0:**
```
scripts/check-references.sh      → check-references: all referenced sections resolve
scripts/check-vocabulary.sh      → Stage-vocabulary guard: clean; Panel-vocabulary guard: clean
scripts/check-contract-budget.sh → BUDGET-OK: 21 contract file(s) within budget
scripts/test-setup.sh            → PASS — 192 assertions, 0 failures
```

---

## task-8-report

# Task 8 report — split `project-configuration.md` into a core and a new appendix

## Status

Done. `skills/myflow-contracts/project-configuration-rationale.md` created; core file split with 18
evicted passages, all citation-repointed and pointer-linked; budget row and SKILL.md index row
added; all required verification scripts exit 0.

## Verification summary

All six required commands exit 0 from the worktree root: `scripts/check-references.sh`,
`scripts/check-vocabulary.sh`, `scripts/check-workspace-isolation.sh`,
`scripts/check-contract-budget.sh`, `scripts/test-check-workspace-isolation.sh` (192 assertions, 0
failures), `scripts/test-setup.sh` (192 assertions, 0 failures). Also ran (not required by this
task's list, but named in the brief's Global Constraints) `scripts/check-plan-provenance.sh` and
`scripts/check-task-build-green.sh` — both exit 0.

Every one of the 18 removed passages was byte-verified by script against the pre-task image
(`git show 47f508711197037c917a88f60f5bfa22944bafab:skills/myflow-contracts/project-configuration.md`,
45,284 bytes — matches the brief's stated starting size). Each was extracted from that image with
`sed -n '<range>p'` and diffed against the text placed in the appendix; all 18 diffs are empty. The
two permitted `above`/`below` deletions (in the `.mdc`-extension passage and the "dropped `url` row"
passage) were each verified to be the *only* difference between the pre-task line range and the
appendix copy — confirmed by a second, separate `diff`. No passage was retyped from memory; every
move was `sed`-extracted programmatically, per the Task 7 lesson in the brief.

Confirmed `git -C /Users/tweety53/Projects/agents status --short` shows no tracked `M` entries — no
edits leaked into the main checkout.

Read the resulting core end to end. An agent could resolve a project's configuration correctly with
only that file: every `## <key>` is still defined, the `## standards` resolution/containment rules
are untouched, the `<agents repo>` derivation's five operative lines are intact, both
`## workspace isolation` schema tables and their enforcement/review-duty lists are untouched, and
every evicted passage left either an unchanged rule sentence (mixed) or a bare pointer (wholly
rationale) naming `project-configuration-rationale.md § Where the agents repository is`.

## Byte counts

| | Before | After |
|---|---|---|
| `skills/myflow-contracts/project-configuration.md` | 45,284 | 38,275 |
| `skills/myflow-contracts/project-configuration-rationale.md` (new) | 0 | 13,225 |

Core shrank by 6,999 bytes even though the diff also *adds* ~6,200 bytes of new pointer/citation
text (each of the 18 evictions leaves a "See **Where the agents repository is** (`...rationale.md`)
for..." sentence in the core) — the appendix's 13,225 bytes is the sum of the 18 moved passages
(byte-identical to source) plus a 4-line title/note preamble.

This lands well above the plan's ~28,000-byte projection. Per the brief's own "do not chase a
number" instruction, the honest reason: two large zones the brief explicitly protects — "How a
`## standards` entry resolves to a file" / the containment block (treat every clause as core), and
the two `## workspace isolation` schema tables plus their cell-form, validation, and
`scripts/check-workspace-isolation.sh` enforcement-list prose (the brief's stated "hazard" in this
file) — together account for roughly 20,000 of the remaining 38,275 bytes and were deliberately left
whole rather than extracted, including their attached one- or two-sentence justificatory clauses,
because none of them separates into a rule/argument pair cleanly enough to extract without either
duplicating the argument or losing an operative fact the guard depends on. What did move is the
genuinely discursive material: measured-layout examples, a worked-example walkthrough, container/
timeout mechanism narration, and "why this differs from that" comparisons.

## Concerns / provenance discrepancy

**The brief's Step 1 "measured" claim does not hold, and I did not follow it as read.** It states the
core has three headings — `# Project configuration`, `## Where the agents repository is`, `##
workspace isolation` — tagged `measured: grep -n '^#\+ ' ... @ f763481`. Re-running that exact grep
at that exact commit does return three matches, but the third (line 379) sits inside a fenced
` ```markdown ``` ` block: it's the heading of the file's own *worked example* of a project's
`.myflow/project.md`, not a real heading of this contract file (confirmed with
`awk '/^```/{print NR}'`, fence 378–397). The document genuinely has only two real headings; every
line from 89 to EOF — including all `## workspace isolation` schema content — sits under the single
`## Where the agents repository is` H2. This is the same shape Task 6's ledger flagged for
`pipeline.md`'s "Wrong state for this command" fenced example. No new heading was fabricated to match
the brief's claim: the appendix mirrors the core's actual two-heading tree. `check-references.sh`
would not have caught either choice (a heading with that text does exist, just inside a fence), which
is exactly the class of gap the ledger — not the reference guard — exists to catch. Full detail in
`.superpowers/sdd/task-8-ledger.md`'s Notes section.

No other concerns. The two attacker-influenced-input protected zones named in the brief were left
fully intact and unextracted, as instructed.

## Ledger

Full 18-row per-move ledger with byte-verification detail: `.superpowers/sdd/task-8-ledger.md`.

---

## task-9-report

# Task 9 report — Split `workspace-isolation.md` into a core and a new appendix

## Status

Done. `skills/myflow-contracts/workspace-isolation-rationale.md` created; core file split with 32
evicted/extracted passages, all citation-repointed and pointer-linked; budget row and `SKILL.md`
index row added; all six required verification scripts exit 0.

## Verification summary

All six required commands exit 0 from the worktree root: `scripts/check-references.sh`,
`scripts/check-vocabulary.sh`, `scripts/check-workspace-isolation.sh`,
`scripts/check-contract-budget.sh` (23 contract files, was 22), `scripts/test-check-workspace-
isolation.sh` (all cases pass), `scripts/test-setup.sh` (192 assertions, 0 failures). Also ran (named
in the brief's Global Constraints, not in this task's required list) `scripts/check-plan-
provenance.sh` and `scripts/check-task-build-green.sh` — both exit 0.

**Every extraction was built and verified programmatically, never retyped from memory** (the Task 7
lesson). A Python script (`str.find`/`str.count`/`str.replace`) located each of the 32 moved/split
passages in the pre-task image
(`git show 541e61453e766b08192176fbd89b95a8c2bf506b:skills/myflow-contracts/workspace-isolation.md`,
31,079 bytes — matches the brief's stated starting size and the fence-aware five-heading scan given
in the prompt, both independently re-confirmed with `sed -n` byte tallies before starting: 8290,
5500, 5362, 5419, 4909 for the five sections, 1599 for the preamble, summing to 31,079). Three
programmatic checks were run (the Task-7 lesson, "verify by script not by eye"), all passing:

1. Every `moved_text` chunk placed in the appendix is a **unique, exact substring of the pre-task
   image** (`text.count(chunk) == 1`), confirmed for all 32 chunks.
2. Every `moved_text` chunk appears **verbatim in the assembled appendix file**.
3. **Full reconciliation**: reversing all 32 core replacements (substituting each new core paragraph
   back for the exact original text it replaced) reproduces the 31,079-byte pre-task image
   **byte-for-byte**, confirmed by direct string equality, newlines included. This is the strongest
   available check that nothing outside the intended 32 spans was touched and that every retyped
   citation sentence sits exactly where the original text used to be.

Read the resulting 25,317-byte core end to end, pretending the appendix does not exist. An agent
could derive a workspace id (full formula, both `verified:` code blocks, and every non-ASCII edge-
case bullet's operative clause, all untouched), determine what the id derives (three-values list,
`<id_underscored>` construction and its code block, port-offset formula and its `lsof` free-check,
all untouched), claim a cache index (the probe-and-atomic-claim procedure, index-0 exclusion, and
exhaustion handling, all untouched), resolve the empty-id case (the rule itself, the main-checkout/
apply-worktree asymmetry, and both refusal bullets, all untouched), and run creation and cleanup (on-
demand creation, the migration-not-copy rule, the removal-command pointer, the full survivor-report
contract verbatim, and the skip-not-fail-on-unreachable-service rule) — every section still states
its own rule directly, and every eviction left a pointer sentence naming the rule already stated
beside it plus a citation to where the argument now lives.

Confirmed `git -C /Users/tweety53/Projects/agents status --short` shows no tracked `M` entries — no
edits leaked into the main checkout (only the pre-existing staged `A` entries from earlier tasks in
this change, and untracked `.serena/`, which was staged by `git add -A` and explicitly `git reset -q
-- .serena/`'d back out before finishing).

## One deliberate deviation from the brief, reported rather than silently applied

The brief lists "the argument for probing rather than deriving the cache index" as rationale. The
paragraph carrying that argument — "**The reason is the size of the space.** A cache offers
**sixteen** indices..." — was kept **wholly core, unchanged** instead. The very next paragraph in the
source opens "**That argument holds only while a probe can see the previous claim...**", referring
straight back to it by position. Per the amended spec's own rule ("a passage that refers to its
neighbour by position stays with that neighbour"), and since the only two edits a moved-or-retained
passage may carry are repointing a citation and deleting a stale `above`/`below` — neither of which
covers rewording "That argument" into something appendix-safe — moving the size-of-the-space
paragraph away would have left a dangling reference with no sanctioned fix. Kept both together. No
fact is lost from the appendix by this: the "sixteen indices" / "six percent" figures this paragraph
gives are independently restated in a definitely-core paragraph later in the same section ("Fifteen
claimable indices is a real ceiling..."). Full reasoning in the ledger's header note.

## Byte counts

| | Before | After |
|---|---|---|
| `skills/myflow-contracts/workspace-isolation.md` | 31,079 | 25,317 |
| `skills/myflow-contracts/workspace-isolation-rationale.md` (new) | 0 | 11,037 |

Core shrank by 5,762 bytes (−18.5%) even though the diff also *adds* roughly 32 new pointer/citation
sentences (each extraction leaves a "See **X** (`...rationale.md`) for Y" sentence in the core, one
per removed passage) — the appendix's 11,037 bytes is the sum of the 32 moved passages (byte-
identical to source) plus a 4-line title/note preamble and five section headings.

This lands well under the plan's ~20,000-byte projection (25,317 vs. ~20,000) — the honest reason,
matching the brief's own framing that "core here is dense": the five explicitly-protected zones (every
derivation formula, the `<id>`/`<id_underscored>` construction, the cache-index probe procedure, the
empty-id rule and its resolution, and the survivor-report contract) account for the bulk of the file
and were left whole or nearly whole, plus the one deliberate deviation above (~660 bytes kept core
that the brief's general hint would have moved). I did not chase the ~20,000 number — every remaining
core sentence was checked individually against the core/rationale test rather than trimmed to hit a
target, per the plan's own "do not chase a number" instruction and Task 6/7/8's precedent (all three
landed materially above or below their own projections for the same reason).

## Formatting note

Every new citation sentence was rewrapped, together with the original bold rule sentence(s) it sits
beside, to the file's established ~100-column hard-wrap (matching the existing wrapped style already
visible in e.g. `project-configuration.md`'s own rule-extraction citations) — this re-flows only
*newly authored or retained-in-place* text; every passage that *moved* to the appendix keeps its
original line breaks byte-for-byte, verified per the reconciliation check above. Two list bullets
(the "Upper case is folded" / "Only trailing `-`" bullets under **The workspace id**) needed a
bullet-aware wrap (2-space continuation indent) to keep their markdown list structure intact.

## Files touched

- `/Users/tweety53/Projects/agents-worktrees/openspec-kan-95-slim-the-myflow-contract-files/skills/myflow-contracts/workspace-isolation.md`
- `/Users/tweety53/Projects/agents-worktrees/openspec-kan-95-slim-the-myflow-contract-files/skills/myflow-contracts/workspace-isolation-rationale.md` (new)
- `/Users/tweety53/Projects/agents-worktrees/openspec-kan-95-slim-the-myflow-contract-files/scripts/check-contract-budget.sh` (new row: `workspace-isolation-rationale.md 13796` = floor(11037 × 1.25); the existing `workspace-isolation.md 38848` row left untouched per Task 6/8 precedent — a shrunk core file doesn't need its already-generous budget tightened; Task 10 re-anchors every row)
- `/Users/tweety53/Projects/agents-worktrees/openspec-kan-95-slim-the-myflow-contract-files/skills/myflow-contracts/SKILL.md` (one new row in the **Rationale appendices** table, not the Index table)
- Ledger: `/Users/tweety53/Projects/agents-worktrees/openspec-kan-95-slim-the-myflow-contract-files/.superpowers/sdd/task-9-ledger.md`

Staged with the prescribed command:
`git -C "$WORKTREE" add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'`,
then `git reset -q -- .serena/` to drop the Serena project-registration files `git add -A` picked up.

---

## panel-fix-round-1

# Panel fix round 1 — F5, F7, F6, F2/F8, F1

Status: all five assigned findings fixed. All nine required verification commands exit 0.

Verification summary: ran `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
`scripts/check-plan-provenance.sh`, `scripts/check-task-build-green.sh`,
`scripts/check-workspace-isolation.sh`, `scripts/check-contract-budget.sh`,
`scripts/test-setup.sh`, `scripts/test-check-contract-budget.sh`, and
`scripts/test-check-workspace-isolation.sh` from the worktree root — all nine exited 0. All
edits are staged (`git add`), excluding `openspec/`, `docs/manual-test/`, `docs/superpowers/`
and `.serena/`.

## F5 — `README.md` citations from installed files

Added a new normative section, **Stage exit — never the command's own judgment**, to
`skills/myflow-contracts/pipeline.md` (an installed contract every command loads), stating the
rule that a looping stage — most concretely `/myflow-start`'s brainstorm convergence loop — never
closes on the command's own judgment, only on an explicit operator answer, with the one bounded
exception for a session that cannot ask at all. Repointed both citation sites in
`skills/myflow-start/SKILL.md` (former lines ~235 and ~448) from `(`README.md`)` to
`**Stage exit — never the command's own judgment** (`skills/myflow-contracts/pipeline.md`)`.

Swept `skills/`, `rules/`, `commands/`, `commands-claude/` for the same `(`README.md`)` citation
shape. Three remaining hits, all left untouched as intentional: `skills/README.md` (both hits)
and `skills/myflow-start/SKILL-rationale.md` are themselves never installed/never loaded by a
run — they sit beside `README.md` in the source repo by design, the same pattern the current
`CLAUDE.md` documents for `skills/README.md`. Root `CLAUDE.md`/`AGENTS.md`/`README.md` were left
untouched — they are outside the sweep's named scope and their `Level 1`/`Level 2` structure is a
separate, deliberate part of this branch's own restructuring, not part of the five assigned
findings.

## F7 — empty `worktrees` map fall-through

`skills/myflow-contracts/finish-contract.md`: changed both occurrences of the fall-through
condition from "absent" to "absent or empty" — the canonical statement under
**Resolving a change's worktrees**, and its restatement under **Worktree cleanup** — so an empty
`worktrees: {}` map now also triggers the per-repository `git worktree list --porcelain` scan
instead of silently examining zero worktrees.

## F6 — restore the "never search the filesystem" prohibition to core

Appended "That is never a licence to skip the step, and never a reason to search the filesystem
for a checkout that might be one." to the core sentence in
`skills/myflow-contracts/project-configuration.md` ending "...treat a miss as an absence rather
than a nearer guess." The original clause in
`skills/myflow-contracts/project-configuration-rationale.md` was left in place, unchanged.

## F2/F8 — sub-headings in `project-configuration-rationale.md`

Split the appendix's single `## Where the agents repository is` section into 13 `###`
sub-headings, one per topic, moving no prose between files (only restructuring within the
appendix): The `.mdc` routing rule; The per-skill link, not the `skills/` directory;
Project-local installs need no link; Confirm the derived root before using it; Prose beside the
tables is for the reader; The accepted cost of `url`-row duplication; A `url` row with no token;
The dropped `url` row is not exempt; Working directory for `survivors`, `remove`, and `create`;
The one-minute bound; The three-`url`-rows worked example; No path is isolated inside a command;
Where enforcement happens. Repointed all 18 citation sites in
`skills/myflow-contracts/project-configuration.md` at their specific sub-heading, in the
named-section form `**Heading** (`path`)`. `scripts/check-references.sh` confirms every citation
resolves.

## F1 — over-budget fixture too close to the budget

`scripts/test-check-contract-budget.sh`: raised the over-budget `SKILL-rationale.md` fixture from
10,000 to 20,000 bytes, well clear of the 9,665-byte budget (which was not touched).

Manual verification: copied `check-contract-budget.sh` and `test-check-contract-budget.sh` to a
scratch directory under `/tmp`, changed the guard's over-budget comparison from
`[ "$actual" -gt "$max" ]` to `[ "$actual" -lt 0 ]` (a comparison that can never be true for a
byte count), and re-ran the harness against the broken guard. Result: 3 failures — "a file over
budget fails", "a skills/<x>/SKILL.md over budget fails", and "a SKILL-rationale.md over budget
fails" (the fixture raised in this fix) — all reporting `exit 0, wanted 1`, confirming the harness
still catches a broken over-budget check after the fixture was raised. The scratch copy was
deleted afterward; the real guard was never modified.

---

## panel-fix-round-2

# Panel fix round 2 — F9/F4, F10, F11

Status: all four findings fixed (F9 and F4 done together, as instructed). All nine required
verification commands exit 0.

Verification summary: ran `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
`scripts/check-plan-provenance.sh`, `scripts/check-task-build-green.sh`,
`scripts/check-workspace-isolation.sh`, `scripts/check-contract-budget.sh`,
`scripts/test-setup.sh`, `scripts/test-check-contract-budget.sh`, and
`scripts/test-check-workspace-isolation.sh` from the worktree root — all nine exited 0. Also
read the new `skills/myflow-contracts/handoff-blocks.md` end to end as `/myflow-status`
rendering a detail view with only that file open, for every state (`STARTED`, both
`IN_PROGRESS` renderings, `FINISHED`): every template, every folded-line/run-only/
missing-vs-dropped rule, the Git-line review-command table, and the `IN_PROGRESS`
rendering-selection table (including which row a merge-status/`prUrl` combination lands on)
are answerable from the core file alone; only the "why" behind each of those decisions routes
to the appendix. All edits staged (`git add`), excluding `openspec/`, `docs/manual-test/`,
`docs/superpowers/` and `.serena/`.

## F9 + F4 — split `handoff-blocks.md`; give F4's orphan its home

Read the amended requirement first (`A section reachable from only one command lives in its own
file`, `openspec/changes/kan-95-slim-the-myflow-contract-files/specs/myflow-contract-economy/spec.md`)
and the split mechanics (`The split is a verbatim partition, with citation repointing as its
only edit`, same file) before touching anything.

**New byte counts:** `skills/myflow-contracts/handoff-blocks.md` — **12,240 bytes** (was 14,063).
`skills/myflow-contracts/handoff-blocks-rationale.md` — **10,800 bytes** (new file). Budget rows
added/re-anchored at `floor(actual × 1.25)`: `handoff-blocks.md` → 15,300 (was 17,578),
`handoff-blocks-rationale.md` → 13,500 (new row). `scripts/check-contract-budget.sh` confirms
both pass (`BUDGET-OK: 24 contract file(s) within budget`).

**What moved, and how.** Extracted every moved passage with `sed`/`cat` (never retyped) and
`diff`-verified each against its source before writing the assembled file, so the new appendix
is byte-for-byte identical to the four passages pulled from `handoff-blocks.md` plus the orphan
section pulled from `pipeline-rationale.md`. No position words (`above`/`below`) needed deleting
in any of them — none were present.

- Four mixed passages used **rule extraction** (whole passage moves to the appendix
  byte-for-byte; core gains a freshly authored short rule statement + citation, never a restated
  argument): the regeneration-vs-storage passage (core now states "no command stores the emitted
  text... always regenerates" and cites the appendix for why), the open-questions passage (core
  now states the operative "renders `none` when zero" rule directly and cites the appendix for
  why it's not run-only), the pre-check-ordering passage, and the recorded-merge-base
  three-conditions passage (both of the latter two: core keeps the short factual anchor already
  fully covered by the rendering-selection table, cites the appendix for the mechanism/argument).
- The `### The block each state renders` orphan in `pipeline-rationale.md` (lines 95–181, ~86
  lines: why `Jira` is run-only, why `IN_PROGRESS` has two renderings, why `Route`/`Outstanding`/
  `Panel` are run-only, and the rest of the merge-status/`prUrl` reasoning) moved whole into
  `handoff-blocks-rationale.md` under the same heading name (it mirrors both the passages above
  and this orphan, since both concern the identical core section) and was deleted from
  `pipeline-rationale.md` — no stub left behind, per the operator's explicit instruction, since
  `pipeline.md`'s own stub for this heading already cites `handoff-blocks.md`, not
  `pipeline-rationale.md`.
- Confirmed `handoff-blocks.md` now cites the reasoning where it belongs: added short
  `(run-only)` pointer sentences next to `Panel` (after the `IN_PROGRESS`-after-`/myflow-do`
  template) and next to `Route`/`Outstanding` (after the run-1 template), plus a one-clause note
  on why `IN_PROGRESS` needs two renderings — all citing
  `**The block each state renders** (`skills/myflow-contracts/handoff-blocks-rationale.md`)`.
- `skills/myflow-contracts/SKILL.md`'s **Rationale appendices** table (not the Index table)
  gained a row: `handoff-blocks-rationale.md` → `handoff-blocks.md`.

**Concern:** `pipeline-rationale.md`'s own budget row (34,488) is now well above its actual size
(20,995 bytes, after the 87-line deletion) — not a check failure (ceiling, not a floor), and the
operator's housekeeping instruction named only the new file and `handoff-blocks.md`'s row, so I
left it untouched. Worth a deliberate re-anchor in a later pass if the ratchet should stay tight.

## F10 — stale "five-command" count

Changed `skills/myflow-contracts/pipeline-rationale.md:83` from "five-command" to "four-command".

**Stale-count sweep** (`skills/`, `rules/`, `commands/`, `commands-claude/`, `README.md`,
`CLAUDE.md`, `AGENTS.md`, plus `scripts/`): grepped for `\bfive\b`, `\bfour\b` near "command",
"the five"/"five commands"/"5-command", every `<number>-command`/`<number> command(s)` shape, and
leftover `myflow-info` references. Every other "five" hit is unrelated to the pipeline's command
count (five review-panel passes, five retired state-file fields, five workspace-isolation
`Resource` words, five engineering principles under SOLID, etc.) — false positives, not stale
pipeline tallies. `README.md:85` ("the four this pipeline has, three of them pipeline commands
and one read-only") and `pipeline.md:62` ("Three pipeline commands, plus one read-only one")
are already correct. No `myflow-info` references remain anywhere outside the archived
change history. No third stale copy found — confirmed by grepping for the literal phrase
"last-line convention", which now appears exactly once (the fixed copy).

## F11 — orphaned mid-sentence passages in `skills/myflow-do/SKILL-rationale.md`

Added five short lead-in sentences (each its own paragraph, never touching the moved text) at
every genuinely dangling passage found — the three named plus two more located by reading the
whole file for unintroduced pronouns/references ("Then", "There is no", "The subagent's working
directory is", "The two reasons", "The exclusion"):

- `## 2. Isolate the workspace` — before "Then compute this worktree's workspace id...".
- `## 5. The review panel` — before "There is no parent-model inheritance..." and before "The
  subagent's working directory is...".
- `## 7. Verify, stage, and hand off` — before "The two reasons are one case..." and before "The
  exclusion is what keeps them out of the diff...".

`git diff` on that file confirms the change is pure insertion — every pre-existing line is
byte-identical; only new lines were added.

## No concerns beyond the one noted under F9/F4.

## Addendum — full budget re-anchor (post round-2 review)

Re-measured all 24 covered files and re-anchored every row in `scripts/check-contract-budget.sh`
to `floor(actual × 1.25)`. 8 rows changed (`SKILL.md`, `finish-contract.md`,
`pipeline-rationale.md`, `pipeline.md`, `project-configuration-rationale.md`,
`project-configuration.md`, `myflow-do/SKILL-rationale.md`, `myflow-start/SKILL.md`); the other
16 were already correct (untouched by either fix round, or already re-anchored when written).
Coverage confirmed 1:1 — 24 files, 24 rows, identical sets, no orphan row, no uncovered file.

Per-run loaded total (`skills/myflow-do/SKILL.md` + `pipeline.md` + `project-configuration.md` +
`workspace-isolation.md` + `state-file.md` + `jira-integration.md`): **162,751 bytes**, vs the
210,481-byte starting point — a **47,730-byte reduction, 22.68%**. This is 1,230 bytes *above*
Task 10's own reported 161,521, because both fix rounds added rules back to loaded files (F5's
restored gate rule in `pipeline.md`, F6's restored prohibition in `project-configuration.md`) —
stated plainly, not adjusted to look better.

All nine guards/harnesses re-run after the re-anchor: all exit 0, including
`test-check-contract-budget.sh` — no fixture decoupled (the `SKILL-rationale.md` over-budget
fixture raised to 20,000 bytes under F1 stays far clear of every real-budget number touched here,
so the re-anchor didn't threaten that test's assertion).

---

## panel-fix-round-3

# Panel fix round 3 — F12, F22, F13, F21, F14, F15, F19, F16

Status: all eight assigned findings fixed. All nine required verification commands exit 0.

Verification summary: ran `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
`scripts/check-plan-provenance.sh`, `scripts/check-task-build-green.sh`,
`scripts/check-workspace-isolation.sh`, `scripts/check-contract-budget.sh`, `scripts/test-setup.sh`,
`scripts/test-check-contract-budget.sh`, and `scripts/test-check-cleanup-complete.sh` from the
worktree root — all nine exited 0 (no byte-budget row required re-anchoring). `git -C
/Users/tweety53/Projects/agents status --short` shows no tracked `M` entries. All edits are staged
(`git add`) in this worktree, excluding `openspec/`, `docs/manual-test/`, `docs/superpowers/` and
`.serena/` (unstaged after a stray `git add -A` picked it up).

## F12 — `README.md` citations in installed `CLAUDE.md`/`AGENTS.md`

Repointed all eight bare `` (`README.md`) `` citations (`CLAUDE.md:86,122,124,126`,
`AGENTS.md:132,168,170,172`) to installed, loaded files: the "state diagram and per-command stage
table" sentence now cites **States** (`skills/myflow-contracts/pipeline.md`) for the diagram and
each command's own `SKILL.md` for its stage sequence, and says plainly that `README.md`'s "How the
pipeline works" is not copied by `setup.sh` into any project and so is not cited here as a source of
record; the three "Level 1 — the stages of each command" citations (start/do/finish rows) now each
cite that command's own `skills/<name>/SKILL.md` directly (no heading invented — plain file
citations, matching the pattern already used elsewhere in the same table for
`skills/myflow-contracts/pipeline.md`/`state-file.md`).

**Sweep of every file `setup.sh` installs or copies** (`skills/*/` — whole directories, so every
`SKILL.md`/`SKILL-rationale.md`/`myflow-contracts/*.md` inside; `commands/*.md`;
`commands-claude/*.md`; `rules/*.mdc`, whose always-on bodies are also inlined into every project's
`CLAUDE.md`/`AGENTS.md` managed block; `CLAUDE.md`; `AGENTS.md`) for citations naming a file not
installed:

- `skills/myflow-start/SKILL-rationale.md:25` had the same bare `` (`README.md`) `` shape — fixed as
  F14 below.
- `rules/kotlin-backend-development-standard.mdc:153` — "Reference: `README.md`,
  `src/app/auth/README.md`" — **not a defect**: these are example doc pointers for a Kotlin
  *target* project's own README, unrelated to myflow pipeline citation; left untouched.
- `skills/myflow-contracts/pipeline-rationale.md` and `state-file.md` cite
  `openspec/specs/myflow-model-policy/spec.md` / `myflow-planning-effort/spec.md`. `openspec/` is
  not installed by `setup.sh`, but these paths are the target project's **own** synced OpenSpec
  state (populated by `/myflow-finish` run 2's delta-sync, not arbitrary editable content, and not a
  filename any project is likely to have collided with) — a materially different risk shape from
  `README.md`, and pre-existing (confirmed via `git log`, predates this branch). Left untouched.
- `skills/README.md` (root of `skills/`, a **file**, not a directory) is **not installed** —
  `install_skills()` only globs `"$SKILLS_SRC"/*/`, which matches directories only. Its own two
  bare `` (`README.md`) `` citations (lines 9–10, 67) are the same shape as F12 but carry no runtime
  risk since nothing ever loads this file in a target project. Left untouched, per the fix's
  "already fixed by round 1's F5 sweep, intentionally out of scope" precedent recorded in
  `panel-fix-round-1.md`.
- No other bare `` (`README.md`) `` citations found in any installed file.

## F22 — restore `README.md:119-127`'s duplicate of the "Stage exit" rule

Replaced the verbatim restatement in `README.md` (the "never on the command's own judgment … one
bounded exception … An operator who is present but silent …" paragraph) with a short paraphrase plus
a citation to **Stage exit — never the command's own judgment**
(`skills/myflow-contracts/pipeline.md`), matching the pattern the very next paragraph
(`README.md:129-130`, the threshold/prompts pointer) already used. This is the eviction round 1's F5
should have performed when it added the rule to `pipeline.md` — recorded in
`.superpowers/sdd/panel-fix-ledger.md`'s round-3 table.

## F13 — worktree resolution: four sites bypassed the empty-map fallback

Read `openspec/changes/kan-95-slim-the-myflow-contract-files/specs/myflow-finish-cleanup/spec.md`'s
amended requirement first. Added an explicit rule to **Resolving a change's worktrees**
(`skills/myflow-contracts/finish-contract.md`) that every step needing "the worktrees" — the
preflight verdict, the unfinished-work gate, and run 2's removal — resolves the set through that
procedure and never reads the state-file `worktrees` map raw, plus the requirement's
resolved-set-still-empty clause: that state stops the run (exactly like `REFUSE`/asking the operator)
rather than passing vacuously. Updated all four sites the finding named to resolve through that
section instead of the raw map, and to state the empty-resolved-set stop condition explicitly:
`finish-contract.md:46` (preflight, multi-repo loop), `finish-contract.md:60` (unfinished-work gate),
`skills/myflow-finish/SKILL.md:56` (preflight verdict list — added a fourth bullet for the
resolved-empty case), `skills/myflow-finish/SKILL.md:81` (unfinished-work gate — added a matching
bullet). The forward citation from the two earlier sites to the later-defined
**Resolving a change's worktrees** section is intentional, per the finding's own note.

## F21 — stale "absent" comments in the cleanup-verifier scripts

`scripts/check-cleanup-complete.sh:302-303` and `scripts/test-check-cleanup-complete.sh:317-318`:
reworded "the state file's map is absent" to "the state file's map is absent or empty" in both
comments, matching the contract's current wording. No functional change — both scripts scan
unconditionally regardless of the map's state.

## F14 — bare `README.md` citation in an installed rationale appendix

`skills/myflow-start/SKILL-rationale.md:25` cited `` **How the pipeline works** (`README.md`) —
specifically its brainstorm expansion `` for the convergence-test structure. Repointed to
**Convergence** (`skills/myflow-start/SKILL.md`) — the actual canonical, installed source for that
structure (confirmed by reading the `### Convergence` section, which covers exactly "what counts as
a planning-stage exchange, the convergence test itself, and why it is one test rather than a rule
per gate," and which `pipeline.md`'s own "Stage exit" section already cites the same way).

## F15 — undefined "three conditions" in `handoff-blocks.md`

Moved the three conditions (recorded-and-resolving / absent / recorded-but-unresolvable) and the
"resolve, then compare, never compare alone" imperative from
`handoff-blocks-rationale.md` into the core `handoff-blocks.md` paragraph that referenced them, so
the normative table no longer has an undefined term. Trimmed the rationale paragraph to keep only
the "why" — why the recorded-but-unresolvable condition is the dangerous one (compared as a bare
string it reads as "not equal to `HEAD`," which falls through to the bare ancestor test) — and
retitled it accordingly. Recorded in the ledger as a genuine cross-file move.

## F19 — coarse citations in `handoff-blocks.md`

Sub-divided `handoff-blocks-rationale.md`'s single `### The block each state renders` heading (which
covered the entire ~14-topic appendix) into 14 precise `### Why …` sub-headings, matching the
granularity `project-configuration-rationale.md` already uses (13 sub-headings under one `##`).
Repointed all seven citation sites in `handoff-blocks.md` at the specific sub-heading each actually
needs: regeneration-beats-storage; open-questions-on-disk (paired in the same sentence with
Jira-run-only, since that one pointer names two distinct facts — kept as two separate named
citations rather than inventing a combined heading); pre-check-ordering; recorded-but-unresolvable
(F15's target); `IN_PROGRESS`-two-renderings; `Route`/`Outstanding`-run-only; `Panel`-run-only. No
prose moved between files — headings and citation targets only, so this is not a ledger row, matching
round 1's F2/F8 precedent for the same shape of change. `scripts/check-references.sh` confirms every
citation resolves.

## F16 — per-move ledger for rounds 1–3

Wrote `.superpowers/sdd/panel-fix-ledger.md` (gitignored, like the rest of `.superpowers/sdd/` — not
staged) covering every rule extraction and section move performed across all three fix rounds, in
the required four-column shape. Reconstructed rounds 1 and 2 from `panel-fix-round-1.md` and
`panel-fix-round-2.md`'s own descriptions plus the current file contents (`git diff` was not
available against a clean baseline for those rounds, since both are already committed-in-worktree
history at this point; the reports' own before/after quotes were sufficient to identify exact
passages). Flagged one drift risk found while reconstructing round 1: F6's fix
(`project-configuration.md`) left a verbatim duplicate of the appended clause in
`project-configuration-rationale.md` — the same shape F22 fixed for F5/`README.md` — but it was not
raised as a panel finding this round, so it was recorded in the ledger and left unfixed rather than
fixed unassigned.

## Concerns

- The F6/`project-configuration.md` duplicate noted above (in F16's section) is a real,
  un-flagged instance of the same drift class F22 fixes. Worth a future finding.
- `skills/README.md`'s own two `README.md` citations (noted under F12's sweep) are the same shape
  as F12 but carry no install-time risk since the file is never installed. Left untouched,
  consistent with round 1's treatment of the same file.

---

## panel-fix-round-4

# Panel fix round 4 — status report

**Status:** All six findings fixed. All nine required verification commands exit 0. No byte-budget
row required re-anchoring.

**Verification summary:** `check-vocabulary.sh`, `check-references.sh`, `check-plan-provenance.sh`,
`check-task-build-green.sh`, `check-workspace-isolation.sh`, `check-contract-budget.sh`,
`test-setup.sh` (192 assertions), `test-check-contract-budget.sh`, and
`test-check-cleanup-complete.sh` all exit 0 on the final state of the worktree.

## Findings

- **S1 (Critical).** `finish-contract.md:269` and `myflow-finish/SKILL.md:318` both cited
  `openspec/specs/myflow-self-review/spec.md` as canonical for run 2 step 8's *procedure* — a path
  `setup.sh` never installs, so every installed project either had no procedure for that step
  (benign) or, worse, ran a contributor-editable OpenSpec file as the source of truth for the one
  step that fires right after `FINISHED` is written (hostile). Fixed by restoring the procedure text
  (recovered from base commit `f7634817738fcf451a673f81170f328d04c15fe9`'s deleted
  `#### Self-review — /myflow-finish run 2` section) into `finish-contract.md`'s own step 8 — an
  installed file `/myflow-finish` actually loads — and repointing both citation sites at it. The
  openspec `Requirement:` heading is kept only in its legitimate role: "the file to change first,"
  never "the runtime source of the procedure." Sweep found a third, uncited copy of the same defect
  in `README.md:101-104` (not in the original panel brief) and fixed it the same way.

- **S2 (Important).** The empty-worktree-set-is-never-a-vacuous-pass rule lived only in
  `finish-contract.md`, which only `/myflow-finish` loads — so `/myflow-do`'s workspace-isolation
  gate and `/myflow-status`'s merge-status report, both of which iterate a change's worktrees, had
  no rule to cite and both looped over the raw state-file `worktrees` map directly. An empty map
  made `/myflow-do`'s real safety guard run zero times and report success. Fixed by hoisting the
  rule into a new **Resolving a change's worktrees** section in `pipeline.md` (loaded by every
  command); `finish-contract.md` now cites it instead of restating it, keeping only what's specific
  to `/myflow-finish` (the git-scan fallback, the `substr`-not-`$2` parsing note). `myflow-do/
  SKILL.md:387` and `myflow-status/SKILL.md:58` now resolve through the new section. Sweep found a
  second, uncited copy of the same "answered once per key" / "every recorded worktree" pattern at
  `myflow-status/SKILL.md:83-87`, ~30 lines past the cited line — fixed the same way.

- **A2 (Important).** Five sites still described the worktree gates as running "once per **recorded**
  worktree," contradicting the scan-fallback resolved-set behaviour: `README.md` (three sites),
  `commands/myflow-finish.md`, `commands-claude/myflow-finish.md`. Reworded each to describe the
  resolved set, citing **Resolving a change's worktrees** (`pipeline.md`) rather than restating the
  mechanism, per README's own stated preference. `CLAUDE.md`/`AGENTS.md`'s `/myflow-finish` row
  carried the same wording and was fixed in the same edit as F24 (same line).

- **F24 (Minor).** Eight citations in `CLAUDE.md`/`AGENTS.md` read "spelled out in its own
  `SKILL.md`" with no section name, so `check-references.sh` — which only verifies named-section
  citations — checked none of them. Six of the eight (the three per-command table rows in each file,
  which name a specific path) now carry a real, guard-verified heading: `/myflow-start` → **A.
  Resolve the change**; `/myflow-do` → **1. Load context and validate the plan**; `/myflow-finish` →
  **Run 1 — integrate**. The remaining two (`CLAUDE.md:86`, `AGENTS.md:132`) cite the bare literal
  `SKILL.md` describing the general per-command pattern, not one specific file — no path resolves
  for them so the guard could never check them either before or after this round; left as-is and
  noted rather than silently skipped.

- **S3 (Minor).** `myflow-start/SKILL-rationale.md:23-25` cited **Convergence**
  (`skills/myflow-start/SKILL.md`) for the convergence-test definition, the planning-stage-exchange
  definition, and the one-test-not-a-rule-per-gate argument — circular, since the rationale file is
  the appendix *to* that section, and the section itself doesn't carry that reasoning. Repointed to
  **Stage exit — never the command's own judgment** (`skills/myflow-contracts/pipeline.md`), verified
  to actually carry the convergence-test description and the single-universal-rule framing the
  sentence claims.

- **A1 (Minor).** `myflow-finish/SKILL.md:48` said "the four preflight checks"; round 3 added a
  fifth outcome bullet (the resolved-set-empty case) to the list two lines below. Dropped the
  number.

## Sweeps run after each fix

- Self-review "canonical for the procedure" / openspec-spec citation: 1 site found beyond the
  panel's two (`README.md:101-104`), fixed.
- "recorded worktree" / raw `worktrees` map iteration: found and fixed a second copy in
  `myflow-status/SKILL.md` (lines 83-87) beyond the cited line 58.
- "four preflight checks" / stale counts: none found beyond A1's own site.
- No other file under `skills/`, `commands/`, `commands-claude/`, `README.md`, `CLAUDE.md`,
  `AGENTS.md` still names `openspec/specs/myflow-self-review/spec.md` as the procedure source, or
  still reads "once per recorded worktree" / "worktrees\` map" as a raw iteration target — confirmed
  by a final grep sweep with zero remaining hits outside `openspec/` and `docs/`.

## Not touched (per brief)

`openspec/`, `docs/superpowers/specs/`, the `project-configuration.md` /
`project-configuration-rationale.md` contested duplicate, and the bare `scripts/*.sh` citations
were left exactly as found.

---

## panel-fix-round-5

# Panel fix round 5 — status report

**Status:** All four findings fixed. All nine required verification commands exit 0. No byte-budget
row required re-anchoring.

**Verification summary:** `check-vocabulary.sh`, `check-references.sh`, `check-plan-provenance.sh`,
`check-task-build-green.sh`, `check-workspace-isolation.sh`, `check-contract-budget.sh`,
`test-setup.sh` (192 assertions), `test-check-contract-budget.sh`, and
`test-check-cleanup-complete.sh` all exit 0 on the final state of the worktree.

## Findings

- **1 (Critical).** Round 4's own fix for the raw-map-iteration defect (S2 above) made
  `/myflow-do`'s workspace-isolation guard in section 7 resolve "the set" per **Resolving a change's
  worktrees** (`pipeline.md`) and stop on an empty resolved set — correct in general, but
  `/myflow-do` never stated *how it itself* resolves that set, and `pipeline.md` explicitly leaves
  that to each command (naming `finish-contract.md` as canonical only for `/myflow-finish`'s own
  procedure). The state file's `worktrees` map is `{}` at `/myflow-start` and stays empty until
  `/myflow-do` writes it at the very end of section 7 — after the guard runs. So on every first run
  the only resolution available was the empty on-disk map, and the new rule ordered a stop before
  lint, test, staging or the state write ever ran. Fixed by giving `/myflow-do` its own resolution in
  section 2: the resolved set is the worktree this run created or resumed there (plus any additional
  worktree the change affects), recorded in the run's own working notes — not the state file — the
  moment the worktree exists, well before section 7's guard or the section 7 state write. Section 7
  now cites that resolution instead of only the generic pipeline rule, so the empty-set stop fires
  only if section 2 itself failed to produce a worktree, never on the ordinary shape of a first run.
  Traced a first run end to end after the fix: the guard now has a non-empty set on a change whose
  on-disk state file still reads `worktrees: {}`.

- **2 (Bugbot, red).** `CLAUDE.md:128` and `AGENTS.md:174` said "the stages of both runs, in order,
  begin at **Run 1 — integrate**" — true of run 1, false of run 2, whose stages begin at a separate
  `# Run 2 — archive and clean up` heading in the same file. Fixed by naming both headings: "the
  stages of run 1... begin at **Run 1 — integrate**, and the stages of run 2 begin at **Run 2 —
  archive and clean up** (both `skills/myflow-finish/SKILL.md`)."

- **3 (Bugbot, yellow).** `finish-contract.md:312` said "Per the pipeline-wide rule above" — the rule
  lives in `pipeline.md`, cited two paragraphs earlier by name, not "above" in this file. Fixed by
  replacing "above" with the citation: "Per **Resolving a change's worktrees**
  (`skills/myflow-contracts/pipeline.md`)". Left `finish-contract.md:330`'s own "**Resolving a
  change's worktrees** above" untouched — that one points at this same file's own `### Resolving a
  change's worktrees` heading a few paragraphs up, a legitimate same-file reference, not the
  cross-file case the finding named.

- **4 (Minor).** `state-file.md:147` and `myflow-do/SKILL.md:80` both called the state file's
  `worktrees` map "the authoritative list of affected worktrees," which read as contradicting
  `pipeline.md`'s "never loops over the state file's `worktrees` map directly." Both are true of
  different things — the map is the authoritative *record*, the resolved set is what a step
  *iterates* — so qualified both: `state-file.md:147` now says "authoritative **recorded** list" and
  adds a sentence that a step resolves the set first per **Resolving a change's worktrees**
  (`pipeline.md`) rather than looping over this map directly; `myflow-do/SKILL.md`'s section 2 (now
  folded into the finding-1 fix) says the map becomes "the authoritative recorded list of affected
  worktrees" at the section 7 write, the same qualifier.

## Sweeps run after each fix

- "authoritative list of affected worktrees" (finding 4's phrase): 2 sites total, both in the panel
  brief (`state-file.md:147`, `myflow-do/SKILL.md:80`); both fixed, none found elsewhere.
- "begin at **Run 1 — integrate**" / stage-citation-covers-both-runs pattern (finding 2's phrase):
  2 sites total, both in the panel brief (`CLAUDE.md:128`, `AGENTS.md:174`); both fixed, none found
  in `README.md`, `commands/`, `commands-claude/`, or any `skills/` file.
- "rule above" / "above" as a cross-file pointer (finding 3's pattern): checked every "above" hit in
  `skills/`, `README.md`, `CLAUDE.md`, `AGENTS.md` — all remaining ones are legitimate same-file
  references (e.g. `finish-contract.md:330`'s own section pointer, `pipeline.md`'s implementer-rule
  callbacks, `handoff-blocks.md`'s internal rule references); none is a stale cross-file "above" like
  finding 3's.
- "once per worktree in the set resolved" / raw-map-iteration pattern (finding 1's shape): checked
  `myflow-status/SKILL.md:58` (same wording as the pre-fix `myflow-do` guard) — not a bug there:
  `/myflow-status` is a read-only report, and `pipeline.md`'s own rule distinguishes a *gate* (stops)
  from a *read-only report* (says so in its output) on an empty resolved set; status already takes
  the report branch ("merge status: unknown, no worktree recorded") rather than halting, so it needed
  no fix. Checked `myflow-finish/SKILL.md:63-93` and `README.md:189-196`: both already cite
  `finish-contract.md`'s own resolution procedure (the git-scan fallback), not the bare
  `pipeline.md` rule alone — no defect. Checked `skills/myflow-do/SKILL-rationale.md` for a stale
  restatement of the old section 2 wording: none found (its section 2 entry covers only the
  workspace-id derivation, untouched by this fix).

## Not touched (per brief)

`openspec/`, `docs/superpowers/specs/`, the `project-configuration.md` /
`project-configuration-rationale.md` contested duplicate (confirmed still being edited concurrently
by another process during this round — left exactly alone), and the bare `scripts/*.sh` citations
were left exactly as found.
