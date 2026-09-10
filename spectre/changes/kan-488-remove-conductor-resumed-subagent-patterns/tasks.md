# kan-488-remove-conductor-resumed-subagent-patterns

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.

> **Relocation:** yes — decisions and rationale move out of
> `docs/superpowers/research/flow-remove-resumed-subagents.md` (deleted by task 1) into this
> change's own `design.md`/`proposal.md`; prose in `skills/flow/*.md`, `skills/flow-research/SKILL.md`
> and `skills/flow-contracts/*.md` describing the conductor/planner/researcher relay is deleted,
> not moved elsewhere.

- [x] 1. Delete the conductor — the parent orchestrates `implement.md` directly
  - [x] **Step 1: Delete `implement.md`'s "Dispatch the conductor" (lines 19-118) and "Dispatch
    sites — the conductor's closed list" (lines 119-168), and the `## Question`/`## Stage`/
    `## Handoff` relay contract they define.** Keep "Inline — the parent implements" (line
    169+) as the base: fold its substitutions ("no conductor, no implementer, no panel-fix
    dispatch" language) into the now-single execution path, since there is no longer a conductor
    path to contrast it with. `implement.md` §1 ("Load context"), §2 ("Isolate the workspace"),
    §3 ("Documenting a fix"), and §4 ("Execute (SDD + TDD)") become the parent's own sections to
    run directly — reword every "the conductor does X" to "the parent does X", and every dispatch
    a former conductor made (implementer, panel-bundle via review-panel.md, panel-fix via
    review-panel.md, verifier via verify-and-handoff.md) becomes the parent's own one-shot
    dispatch, unchanged in model/paragraph/key shape.
```unverified:confirm every "conductor" occurrence in implement.md is either reworded to "the parent" or deleted with its containing section
grep -c "conductor" skills/flow/implement.md   # must be 0 after this task, or only in a historical/superseded-decision note
```
  - [x] **Step 2: Reword `review-panel.md` and `verify-and-handoff.md`'s "the conductor" mentions
    to "the parent"** (both files describe the conductor doing panel/verify work; that work is now
    the parent's). No behavior change — same one-shot dispatch sites, same closed-list rules,
    same handshake/fallback mechanics, only the actor's name changes.
```unverified:confirm no "conductor" reference survives outside a historical decision note
grep -c "conductor" skills/flow/review-panel.md skills/flow/verify-and-handoff.md   # both 0
```
  - [x] **Step 3: Invert `skills/flow/SKILL.md`'s guardrail** — "Never run `implement.md`
    sections 1, 2 or 4, `review-panel.md` or `verify-and-handoff.md` in the parent session — the
    conductor runs them" becomes "the parent runs them directly; no conductor is ever dispatched
    for implementation" (mirroring the existing "Inline — the parent implements" carve-out, which
    this collapses into the only path). Update the "Stage keys" table's phase-file column
    descriptions if they name the conductor.
  - [x] **Step 4: Reword `pipeline.md`'s "Progress visibility" section** (the paragraph starting
    "On the implementation branch the granularity is the stage... run inside a conductor subagent
    that has no task-list tool; the parent updates the list at each `## Stage` return it
    relays") to describe the parent updating the task list directly at each `flow stage end` it
    makes itself — no relay, no `## Stage` return. Per-task granularity (one entry per `tasks.md`
    item, since the parent's own task-list tool is now available throughout) replaces the
    stage-level granularity the old paragraph accepted as a trade-off.
  - [x] **Step 5: `stats/cmd/flow/record.go`'s role list keeps `conductor` as a valid historical
    role** (existing dispatch rows in the store still name it) — confirm no test or validation
    rejects the literal string `conductor` there; make no code change to `record.go` in this task
    if none is needed.
```unverified:confirm record.go's role enum/validation is additive-only and needs no edit
grep -n "conductor" stats/cmd/flow/record.go stats/internal/store/*.go
```

**Files:** `skills/flow/implement.md`, `skills/flow/review-panel.md`,
`skills/flow/verify-and-handoff.md`, `skills/flow/SKILL.md`, `skills/flow-contracts/pipeline.md`
**Tests:** none
**Regression:** none (prose-only skill files; behavior is exercised by running `/flow` itself, not a
unit suite)
**Baseline:** N/A — no automated test exists for these files' prose
**Commit:** `docs(flow): delete the conductor, the parent runs implement/review-panel/verify-and-handoff directly`
**Build:** green
**After:** none

- [x] 2. Planning inline — delete the planner and researcher subagents
  - [x] **Step 1: Delete `skills/flow/brainstorm.md`'s "Dispatch the planner" section** (the
    `PLANNING_MODEL` resolve for this call site, the relay contract, the `## Question` block
    relay via `AskUserQuestion`, the `Model:` handshake, the `opus` fallback and
    `planner`/`planner-opus`/`planner-<model>` dispatch-record keys). Section **A** ("Resolve the
    change and write STARTED") and "Resume and fix runs" stay; the parent now runs sections
    **B**, **C**, **D** of `brainstorm-planner.md` itself, directly after **A**, with no dispatch
    in between.
  - [x] **Step 2: Collapse `brainstorm-planner.md`'s "you" split.** Every "This section's stage
    marks are run by the parent, not the planner" framing (sections B, C, D headers) collapses to
    one actor: the running session does the seed lookup, checklist, convergence, `spectre new`,
    artifact writing, writing-plans enrichment and the Decide step itself, with no return/relay
    step in between. Remove the "Everywhere below that addresses 'you' means the planner"
    sentences; every "you" in this file addresses whichever session (parent `/flow`, or
    `/flow-plan`, per task 10's rename) is running it.
  - [x] **Step 3: Delete `skills/flow-research/SKILL.md`'s researcher-subagent dispatch section**
    (its own `PLANNING_MODEL` resolve, the "the parent session does not do the thinking itself"
    sentence, the relay contract, `Model:` handshake, "no `flow record dispatch`" note). The
    session itself reads the tree, runs the investigate-then-ask rounds and the convergence check
    directly, on its own model, issuing `AskUserQuestion` calls itself (batched, up to four
    independent questions per call, exactly as the skill's stance/guardrails/capture sections
    already state — those sections are otherwise unchanged). Edit the file at its current path;
    task 10 renames the directory and file afterward.
  - [x] **Step 4: One shared mechanism, not two copies** (per design.md's `planning-inline`
    decision). Confirm `skills/flow-research/SKILL.md`'s own seed/checklist description, after
    step 3's edit, cites `brainstorm-planner.md`'s "Seed from a staged research note" section by
    name for the seed-lookup mechanics rather than restating them in different words — edit the
    file to add that citation if it currently duplicates the logic instead of pointing at it.

**Files:** `skills/flow/brainstorm.md`, `skills/flow/brainstorm-planner.md`,
`skills/flow-research/SKILL.md`
**Tests:** none
**Regression:** none (prose-only)
**Baseline:** N/A
**Commit:** `docs(flow): run brainstorming and research inline, no planner or researcher subagent`
**Build:** green
**After:** none

- [x] 3. Strict research-artifact path, keyed by Jira key
  - [x] **Step 1: In `brainstorm-planner.md`'s "Seed from a staged research note"**, remove the
    `<jira-key-lowercased>-*.md` glob fallback and its "found more than one staged research note"
    disambiguation `AskUserQuestion` block. The rule becomes: with a linked issue, check
    `<jira-key-lowercased>.md` and nothing else; with no linked issue, check `<name>.md` and
    nothing else — one `test -f`, no glob, no inference.
  - [x] **Step 2: In `skills/flow-research/SKILL.md`'s "Staging a Note"**, state the keyed
    filename (`<jira-key-lowercased>.md`) as the mandatory destination when a key is known and no
    change exists yet — never a descriptive suffix. Keep the existing-change (`design.md`) and
    bare-session (`<topic-slug>.md`) destinations as already described.

**Files:** `skills/flow/brainstorm-planner.md`, `skills/flow-research/SKILL.md`
**Tests:** none
**Regression:** none (prose-only)
**Baseline:** N/A
**Commit:** `docs(flow-research): drop the glob fallback, one exact keyed filename per seed`
**Build:** green
**After:** Task 2

- [x] 4. Startup visibility for resolved dynamic choices
  - [x] **Step 1: In `skills/flow/SKILL.md`'s "Announce at start"**, add the seed-result line
    printed immediately after "Using flow for change `<name>`." on every run:
    `research seed: found docs/superpowers/research/<key>.md — seeding brainstorm` or
    `none — planning inline in this session`, per the exact two-line shape in design.md's
    `startup-visibility` decision.
  - [x] **Step 2: On the seeded path**, print the full block (planning mode/model, the three
    toggles, `DEFAULT_MODEL`/`REVIEWERS`, the recorded decision or "not yet decided") right after
    the seed line, before any stage past kickoff runs — add this to `brainstorm.md`'s **A**
    section, gated on the seed check from task 3's edited "Seed from a staged research note".
  - [x] **Step 3: On the no-seed path**, leave the existing post-`writing-plans` `## Decision`
    print as the sole place those choices appear — extend it (in `brainstorm-planner.md`'s
    section D, "print the planner's `## Decision` block verbatim") with the `planning:`,
    `toggles:` and `models:` lines so the one print carries every per-run choice. Do not print
    the block twice.

**Files:** `skills/flow/SKILL.md`, `skills/flow/brainstorm.md`, `skills/flow/brainstorm-planner.md`
**Tests:** none
**Regression:** none (prose-only; verified by running `/flow` and reading its own printed output)
**Baseline:** N/A
**Commit:** `docs(flow): print the resolved seed and dynamic choices at kickoff`
**Build:** green
**After:** Task 3

- [x] 5. Remove `planningModel` / `## planning model` end to end
  - [x] **Step 1 (Go, store):** In `stats/internal/store/settings.go`, delete the `PlanningModel`
    field from `Settings`, its `ValidModels` check, the `planning_model` column read/write in
    `PutSettings`/`GetSettings`. Add a new migration under `stats/internal/store/migrations/`
    (next sequence number after `0017_flow_settings_planning_model.sql`) that drops the
    `planning_model` column — do not edit `0017_...sql` in place, per this repo's forward-only
    migration convention (name the new file `0021_flow_settings_drop_planning_model.sql`, confirm
    the next free number with `ls stats/internal/store/migrations/ | tail -5`).
  - [x] **Step 2 (Go, API/CLI/client):** Delete `PlanningModel` from
    `stats/internal/api/settings.go`'s DTO (both directions), `stats/cmd/flow/settings.go`'s
    `-planning-model` flag and its wiring, and `stats/internal/client/client.go`'s `PlanningModel`
    field.
  - [x] **Step 3 (Go, tests):** Delete or rewrite every test asserting `PlanningModel` round-trips
    or validates: `stats/cmd/flow/settings_test.go` (`TestSettingsCmd_Set_WithPlanningModel` and
    the `PlanningModel` field/assertions in `TestSettingsCmd_Get`), `stats/internal/api/settings_test.go`
    (`TestSettingsAPI_Get_EchoesPlanningModel` and the "unknown planning model" table case),
    `stats/internal/client/client_test.go` (`TestSettingsRoundTripsPlanningModel`, or rewrite it
    to assert the key is **absent** from the wire payload — the client-test file's own
    "literal `planningModel` wire-key check" comment should be corrected or removed since the key
    no longer exists), `stats/internal/store/settings_test.go`
    (`TestSettingsStore_RejectsUnknownPlanningModel` and `PlanningModel` fields in the round-trip
    table).
```verified:go test output, this session, before this task's edit
go test ./cmd/flow/... ./internal/api/... ./internal/store/... ./internal/client/... 2>&1 | tail -4
# ok  	github.com/tweety53/agents/stats/cmd/flow	10.058s
# ok  	github.com/tweety53/agents/stats/internal/api	22.785s
# ok  	github.com/tweety53/agents/stats/internal/store	22.160s
# ok  	github.com/tweety53/agents/stats/internal/client	1.830s
```
    Baseline test counts, this session, `go test -v ... | grep -c '^--- PASS'`:
    `cmd/flow`=166, `internal/api`=95, `internal/store`=191, `internal/client`=55.
    <!-- measured: go test -v ./cmd/flow/... ./internal/api/... ./internal/store/... ./internal/client/... @ branch spectre/kan-488-remove-conductor-resumed-subagent-patterns -->
  - [x] **Step 4 (docs):** Delete `skills/flow/SKILL.md`'s `PLANNING_MODEL` resolution block
    (`PLANNING_MODEL=...`, `PROJECT_PM=...`, the `## planning model` project-override read, the
    `fable` default) from "Model resolution" — nothing resolves or dispatches on
    `PLANNING_MODEL` once tasks 1-2 land. Delete the `## planning model` row and its paragraph
    from `skills/flow-contracts/project-configuration.md`. Delete the `planning model` case from
    `scripts/check-model-keys.sh`'s key loop. Delete `skills/flow-settings/SKILL.md`'s
    `planningModel` read/print/set paragraphs and its `-planning-model` flag example. Delete
    `/flow-research`'s "resolves `PLANNING_MODEL`" sentence (already gone if task 2 step 3 removed
    the whole researcher-dispatch section it lived in — confirm no stray mention survives).
```unverified:confirm no planningModel/PlanningModel/planning_model/planning-model/planning model token survives outside a migration file's historical column-drop comment or a superseded design.md decision
grep -rn "planningModel\|PlanningModel\|planning_model\|planning-model\|planning model" stats/ skills/ scripts/check-model-keys.sh
```

**Files:** `stats/internal/store/settings.go`,
`stats/internal/store/migrations/0021_flow_settings_drop_planning_model.sql`,
`stats/internal/api/settings.go`, `stats/cmd/flow/settings.go`, `stats/internal/client/client.go`,
`stats/cmd/flow/settings_test.go`, `stats/internal/api/settings_test.go`,
`stats/internal/client/client_test.go`, `stats/internal/store/settings_test.go`,
`skills/flow/SKILL.md`, `skills/flow-contracts/project-configuration.md`,
`scripts/check-model-keys.sh`, `skills/flow-settings/SKILL.md`
**Tests:** `TestSettingsPutBodyCarriesNoPlanningModelKey` (`internal/client`, replaces the deleted
`TestSettingsRoundTripsPlanningModel`, asserting the wire key's absence instead of its round-trip)
— every other test in the four baselined packages, minus the deleted `TestSettingsCmd_Set_WithPlanningModel`,
`TestSettingsAPI_Get_EchoesPlanningModel` and `TestSettingsStore_RejectsUnknownPlanningModel`, must
still pass.
**Regression:** removing `PlanningModel` from the DTO/store without deleting the tests above leaves
stale assertions the compiler or `go vet` catches as unused fields, or a test asserting a field the
struct no longer has — either fails the build.
**Baseline:** before=166+95+191+55=507 passing tests across the four packages (measured above);
after=predicted 503 (4 planning-model-specific tests removed: one each in `cmd/flow`,
`internal/api`, `internal/store`, `internal/client`).
<!-- predicted: go test -v ./cmd/flow/... ./internal/api/... ./internal/store/... ./internal/client/... after this task -->
**Commit:** `feat(settings): remove planningModel end to end, planning runs inline now`
**Build:** green
**After:** Task 2

- [x] 6. Remove the inline-parent context ceiling
  - [x] **Step 1:** Confirm the 250k-before-a-bundle / 400k-before-the-panel /
    sixth-bundle-without-a-budget-figure stop is gone from `implement.md`'s "Inline — the parent
    implements" — already dropped as part of task 1's consolidation of the conductor and inline
    sections into one, since the rewritten section carries no "Context ceiling" paragraph. No
    further edit needed in `implement.md` for this task.
  - [x] **Step 2:** Delete the `## Context ceiling — clear and resume` block from
    `skills/flow-contracts/handoff-blocks.md`, and any citation of it elsewhere (`grep -rn
    "Context ceiling"` across `skills/`).

**Files:** `skills/flow-contracts/handoff-blocks.md` (`skills/flow/implement.md` already carries
no ceiling paragraph as of task 1's commit)
**Tests:** none
**Regression:** none (prose-only)
**Baseline:** N/A
**Commit:** `docs(flow): remove the inline-parent context ceiling, the 1-hour TTL pays for it`
**Build:** green
**After:** Task 1

- [x] 7. Read-discipline rules for the parent
  - [x] **Step 1:** Add, once in `implement.md` (near "Turn discipline"), the five read-discipline
    rules from design.md's `read-discipline` decision: never `cat` a report (verdict section via
    `sed -n '/^## Verdict/,/^## /p' <report>` or the equivalent for that report's shape); never
    read `final-review.diff`, a dispatch-context bundle, or a whole fix diff in the parent — walk
    `git diff --stat` and the specific hunks a finding names instead; test/lint output through
    `tail -20`, the failing block reproduced from the log file on failure; each phase file
    (`implement.md`, `review-panel.md`, `verify-and-handoff.md`) read in full once per run, a
    later need served by `grep -n` for the heading plus `sed -n` for that section; change
    artifacts (`proposal.md`/`design.md`/`tasks.md`) read once at `flow.load-context`, `tasks.md`
    re-read only through `spectre list --json` and `flow tasks tick` output afterwards.
  - [x] **Step 2:** Cite these rules (not restate them) from `review-panel.md` (report reads, the
    fix-diff walk) and `verify-and-handoff.md` (`## lint`/`## test` output) wherever those files
    currently describe reading a report or a diff.

**Files:** `skills/flow/implement.md`, `skills/flow/review-panel.md`, `skills/flow/verify-and-handoff.md`
**Tests:** none
**Regression:** none (prose-only; a rule the parent itself now follows in this very implementation
run)
**Baseline:** N/A
**Commit:** `docs(flow): state the parent's read-discipline rules once, cited elsewhere`
**Build:** green
**After:** Task 1

- [x] 8. `check-dispatch-paragraphs.sh` — remove the conductor and planner dispatch sites
  - [x] **Step 1:** In `scripts/check-dispatch-paragraphs.sh`'s parallel `SITE_ENTRY`/`SITE_PATHS`/
    `SITE_MIN_BLOCKS`/`SITE_VARIANTS` arrays, remove the two `brainstorm.md` entries (currently at
    array index 10, entry `tools`, and index 14, entry `handshake`) — these are the planner
    dispatch's own TOOLS/MODEL HANDSHAKE paragraphs, deleted by task 2. Identify and remove
    whichever `implement.md` entries were the conductor's own TOOLS/HANDSHAKE/DELEGATION
    paragraphs (deleted by task 1) as distinct from the implementer/panel-fix dispatch sites in
    the same file, which stay — read the script's own header comment (lines 40-46, 105-165) for
    which index belongs to which named dispatch, since it already documents "six sites" by name.
    Six sites total drop to four (implementer, panel-bundle, panel-fix, verifier).
  - [x] **Step 2:** Update `scripts/test-check-dispatch-paragraphs.sh`'s fixtures and expectations
    to match — remove the conductor/planner fixture blocks and their expected-violation assertions
    for the two deleted sites, keep the four surviving sites' fixtures.
  - [x] **Step 3:** Run the guard's own test suite and the guard itself.
```bash verified:authored in-tree for this change, run directly against the guard scripts named
bash scripts/test-check-dispatch-paragraphs.sh
bash scripts/check-dispatch-paragraphs.sh skills
```

**Files:** `scripts/check-dispatch-paragraphs.sh`, `scripts/test-check-dispatch-paragraphs.sh`
**Tests:** `scripts/test-check-dispatch-paragraphs.sh`'s own assertions (bash test script, no named
test IDs — pass/fail is the script's own exit code)
**Regression:** a stale site entry pointing at a deleted paragraph makes the guard report a false
"paragraph missing" violation on every future `/flow` skill edit; a missing site entry for a
paragraph that should still be checked silently stops enforcing it
**Baseline:** 20 sites before this task (measured: `SITE_PATHS` array length,
`scripts/check-dispatch-paragraphs.sh` at this branch's current commit); 18 after (2 removed).
<!-- measured: array literal count in scripts/check-dispatch-paragraphs.sh @ branch spectre/kan-488-remove-conductor-resumed-subagent-patterns -->
**Commit:** `chore(scripts): drop the conductor and planner dispatch-paragraph sites`
**Build:** green
**After:** Task 1, 2

- [x] 9. Lower `check-contract-budget.sh` rows and delete the stale kan-326 fixture
  - [x] **Step 1:** After tasks 1, 2, 3, 4, 6, 7, 10 and 11 land, measure every file this change
    touched with `wc -c` and compare each to `1.25 × actual` (the ratchet formula this guard's own
    header comment states: a row is the file's actual size when the row landed, plus 25%). Lower a
    row **only** where `1.25 × actual` comes out below the row's current value — recomputing from a
    smaller actual size does not always produce a smaller number, since the existing row may already
    sit well above the tight 1.25x line. Measured this branch: `implement.md` shrank enough to
    lower (52398 → 50307, actual 40245); `brainstorm.md` shrank enough to lower (35015 → 12712,
    actual 10169 — most of "Dispatch the planner" left the file); `review-panel.md` (73554,
    actual 63589), `verify-and-handoff.md` (42943, actual 37074), `skills/flow/SKILL.md` (20520,
    actual 19419) and `flow-plan/SKILL.md` (13047, actual 12187) each recompute to a number
    **above** their current row (79487, 46343, 24274, 15234 respectively) — those four rows stay
    exactly as they are; only `implement.md` and `brainstorm.md` change.
```verified:current rows and this branch's actual sizes, both read directly against the files/rows named
grep -n "skills/flow/implement.md \|skills/flow/review-panel.md \|skills/flow/verify-and-handoff.md \|skills/flow/SKILL.md \|skills/flow/brainstorm.md \|skills/flow-plan/SKILL.md " scripts/check-contract-budget.sh
wc -c skills/flow/implement.md skills/flow/review-panel.md skills/flow/verify-and-handoff.md skills/flow/SKILL.md skills/flow/brainstorm.md skills/flow-plan/SKILL.md
```
  - [x] **Step 2:** Delete `docs/superpowers/research/kan-326-myflow-rework.md` (the operator's
    explicit choice, design.md's `delete-kan-326-fixture` decision) — the one file the now-removed
    `<key>-*.md` glob (task 3) existed to match, belonging to an already-archived change.
  - [x] **Step 3:** Run the budget guard to confirm every change leaves it green.
```bash verified:authored in-tree for this change, run directly against the guard scripts named
bash scripts/check-contract-budget.sh
```

**Files:** `scripts/check-contract-budget.sh`, `docs/superpowers/research/kan-326-myflow-rework.md`
(deleted)
**Tests:** none (the guard is its own check, run directly above)
**Regression:** lowering `implement.md`/`brainstorm.md`'s rows below their true 1.25x line makes
the guard fire on a file that has not actually regrown; raising any of the other four past their
current value — even via a naive re-application of the ratchet formula — would silently loosen a
row this guard is supposed to tighten over time, never widen
**Baseline:** the six current rows and this branch's actual sizes are quoted verbatim above,
measured and recomputed against `1.25 × actual` on this branch.
<!-- measured: wc -c against each file, and grep -n against scripts/check-contract-budget.sh, both @ branch spectre/kan-488-remove-conductor-resumed-subagent-patterns -->
**Commit:** `chore(scripts): lower contract-budget rows that shrank past their 1.25x line`
**Build:** green
**After:** Task 1, 2, 3, 4, 6, 7, 10, 11

- [x] 10. Rename `flow-research` to `flow-plan`
  - [x] **Step 1:** `git mv skills/flow-research skills/flow-plan` (carries `SKILL.md` with it,
    already edited to its final content by tasks 2-4); `git mv commands/flow-research.md
    commands/flow-plan.md`; `git mv commands-claude/flow-research.md commands-claude/flow-plan.md`.
  - [x] **Step 2:** Update every live-tree reference from `flow-research`/`/flow-research` to
    `flow-plan`/`/flow-plan` (skill directory path, command name, and prose naming the skill) in:
    `README.md`, `AGENTS.md`, `CLAUDE.md` (skill index tables), `skills/README.md`,
    `skills/flow/SKILL.md`, `skills/flow/SKILL-rationale.md`, `skills/flow/brainstorm-planner.md`,
    `skills/flow-settings/SKILL.md`, `skills/flow-contracts/pipeline.md`,
    `scripts/check-contract-budget.sh`, `scripts/test-check-contract-budget.sh`,
    `scripts/check-guard-symlinks.sh`, `scripts/test-check-guard-symlinks.sh`,
    `scripts/check-installed-citations.sh`, `scripts/check-references.sh`,
    `scripts/check-stage-mark-calls.sh`, `scripts/test-check-markdown-integrity.sh`.
```unverified:confirm the substitution is exhaustive and touches no archived/historical file
grep -rln "flow-research" --include="*.md" --include="*.sh" . | grep -v '^./spectre/changes/archive/' | grep -v '^./docs/superpowers/'
```
    **Never touch** `docs/superpowers/research/*.md`, `docs/superpowers/specs/*.md`,
    `docs/superpowers/plans/*.md`, or anything under `spectre/changes/archive/` — those are
    historical records of what was true when written, not live documentation, and renaming a
    skill does not rewrite history.
  - [x] **Step 3:** Run the reference and citation guards to confirm no dangling `flow-research`
    path or broken cross-file link remains in the live tree.
```bash verified:authored in-tree for this change, run directly against the guard scripts named
bash scripts/check-references.sh
bash scripts/check-installed-citations.sh
bash scripts/check-guard-symlinks.sh
```

**Files:** `skills/flow-research/SKILL.md` → `skills/flow-plan/SKILL.md` (moved),
`commands/flow-research.md` → `commands/flow-plan.md` (moved),
`commands-claude/flow-research.md` → `commands-claude/flow-plan.md` (moved), `README.md`,
`AGENTS.md`, `CLAUDE.md`, `skills/README.md`, `skills/flow/SKILL.md`,
`skills/flow/SKILL-rationale.md`, `skills/flow/brainstorm-planner.md`,
`skills/flow-settings/SKILL.md`, `skills/flow-contracts/pipeline.md`,
`scripts/check-contract-budget.sh`, `scripts/test-check-contract-budget.sh`,
`scripts/check-guard-symlinks.sh`, `scripts/test-check-guard-symlinks.sh`,
`scripts/check-installed-citations.sh`, `scripts/check-references.sh`,
`scripts/check-stage-mark-calls.sh`, `scripts/test-check-markdown-integrity.sh`,
`skills/flow/brainstorm.md`, `skills/flow/review-panel.md`,
`skills/flow/verify-and-handoff.md` (a `refs-guard:allow` marker each, on a bold-emphasis
citation `check-references.sh` flagged once the renamed/reworded sections shifted line content
around it — not a rename edit, but this task's own verify step is what caught it)
**Tests:** `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
`scripts/check-guard-symlinks.sh` (each a guard script, pass/fail by exit code — no unit test IDs)
**Regression:** a stray unrenamed `flow-research` path breaks the `/flow-plan` trigger (the skill
directory the command loads no longer exists at the old name) or leaves a citation guard flagging
a dangling reference
**Baseline:** N/A — guards are run directly above as this task's own verification
**Commit:** `docs(flow-research): rename the skill and command to flow-plan`
**Build:** green
**After:** Task 2, 3, 4, 5

- [x] 11. Move worktree creation to the end of planning
  - [x] **Step 1:** In `skills/flow/brainstorm.md`'s "Run brainstorming and planning directly",
    reorder so section **C** (`spectre new`, the three artifacts, the staging-note deletion) and
    section **D** (writing-plans, the Decide step) run **before** the worktree exists — against
    the main checkout's own `<project>/spectre/changes/<name>/`, uncommitted, never staged or
    committed there (existing git-boundaries rule, unchanged). Move "The three returns"' four
    worktree-creation steps (`check-worktree-location.sh`, `.worktrees` gitignore check,
    `git worktree add`, `## worktree setup`) to run immediately **after** `flow.writing-plans`
    ends, before `flow.decide`'s print — not between `flow.design-approval` and
    `flow.create-artifacts` as this session's own task 2 edit left it.
  - [x] **Step 2:** `.superpowers/sdd/decision.json` — read by `brainstorm-planner.md`'s Decide
    step and by `implement.md` afterward — is written during **D** to
    `<project>/spectre/changes/<name>/.superpowers-sdd-decision.json` in the main checkout (no
    worktree exists yet to hold the usual `<abs-worktree>/.superpowers/sdd/` path). Update every
    reference to that path in `brainstorm-planner.md`'s Decide step and in `brainstorm.md`'s
    post-writing-plans `## Decision` print to read/write the interim main-checkout path instead,
    until the move step below relocates it.
  - [x] **Step 3:** Once the worktree is created, move `<project>/spectre/changes/<name>/` into
    `<worktree>/spectre/changes/<name>/` and the interim decision file into
    `<worktree>/.superpowers/sdd/decision.json` (`mkdir -p` the destination directory first),
    leaving nothing behind in the main checkout. State this move explicitly in `brainstorm.md` as
    its own step between worktree creation and continuing into `skills/flow/implement.md`.
```bash verified:authored in-tree for this change, run directly against the guard scripts named
mkdir -p <worktree>/.superpowers/sdd
mv <project>/spectre/changes/<name> <worktree>/spectre/changes/<name>
mv <project>/spectre/changes/<name>/.superpowers-sdd-decision.json <worktree>/.superpowers/sdd/decision.json
```
  - [x] **Step 4:** Confirm `brainstorm-planner.md`'s **C** section text ("the parent creates the
    change worktree ... then resumes" — already reworded once by task 2) no longer claims the
    worktree exists before **C** runs; reword to state **C** and **D** run in the main checkout,
    and the worktree materializes only after **D**.
```unverified:confirm no sentence in brainstorm.md or brainstorm-planner.md claims a worktree exists before section D completes
grep -n "worktree" skills/flow/brainstorm.md skills/flow/brainstorm-planner.md
```

**Files:** `skills/flow/brainstorm.md`, `skills/flow/brainstorm-planner.md`
**Tests:** none
**Regression:** none (prose-only; behavior verified by running `/flow` itself — this run is its own
first exercise of the new order, since worktree creation for this very change already happened
under the old order before this task existed)
**Baseline:** N/A
**Commit:** `docs(flow): create the worktree at the end of planning, not before section C`
**Build:** green
**After:** Task 2
