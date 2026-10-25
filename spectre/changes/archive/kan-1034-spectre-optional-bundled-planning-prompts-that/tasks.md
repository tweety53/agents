# kan-1034-spectre-optional-bundled-planning-prompts-that

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** yes — the generic planning paragraphs move from the flow skills into `skills/spectre-planning/SKILL.md`
> **Design shape:** plan-mode
> **Tasks appended:** 1
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

- [x] 1. Skill: move the generic planning prompts into `spectre-planning`
  - [x] **Step 1: Create `skills/spectre-planning/SKILL.md` with frontmatter `name: spectre-planning` and a description naming both modes. Move verbatim, paragraph by paragraph, from `skills/flow/brainstorm-planner.md`: the generic checklist bullets (invoke `superpowers:brainstorming` 1–8, display-only formatting, reported-defect history check, HARD GATE, no "does this look right?", batched questions, approved design is the source, tracked-path Done-when); **Explore and design agents** without its `plan-agent-pair.sh` and auto-resolution paragraphs; **Stage exit**; **Convergence** without the `## decisions: recommended` lines; section C's stubs, three artifacts, four plan sections, spec-edit rule, **Decisions** and **Open questions**; section D's writing-plans invocation, task shape and every generic plan-writing rule and plan-guard expectation; the plan review gate's shape. Keep each moved heading's text unchanged so citations resolve by heading.**
  - [x] **Step 2: Move `skills/flow/brainstorm.md`'s Plan-mode paragraph into the skill as `### Plan mode`, generalised: on Claude Code call `EnterPlanMode` before exploring and take approval from `ExitPlanMode`; on Cursor switch to Plan mode before exploring, or ask the user to switch when the agent cannot. `brainstorm.md` keeps only its flow-only clause (after `flow.brainstorm` begins, `STARTED` and the worktree exist) and cites the skill's section.**
  - [x] **Step 3: In each source file replace every moved passage with a `Load` directive naming `skills/spectre-planning/SKILL.md`, or a citation of the moved heading in that file; flow-only lines stay. `skills/flow-plan/SKILL.md`'s plan-mode paragraph cites the new section. Re-point every citation of a moved heading found by `grep -rn 'brainstorm-planner.md' skills rules commands-claude README.md AGENTS.md CLAUDE.md`. Add the file to the planning load set in `stats/internal/guard/loadsets.go`, and a row for `skills/spectre-planning/` in the skill index of `CLAUDE.md` and `AGENTS.md`.**
  - [x] **Step 4: Verify: `scripts/check-verbatim-moves.sh && scripts/check-references.sh && scripts/check-installed-citations.sh && scripts/check-markdown-integrity.sh && scripts/check-stage-mark-calls.sh && scripts/check-guard-symlinks.sh && (cd stats && gofmt -l . && go vet ./internal/guard/...)`, then `git grep -n -F -e 'Plan mode on Claude Code' -- skills` prints nothing.**
**Build:** green
**Files:** `skills/spectre-planning/SKILL.md`, `skills/flow/brainstorm-planner.md`, `skills/flow/brainstorm.md`, `skills/flow-plan/SKILL.md`, `stats/internal/guard/loadsets.go`, `CLAUDE.md`, `AGENTS.md`, `skills/spectre-planning/scripts/check-plan-shape.sh`, `skills/spectre-planning/scripts/lib`, `stats/internal/guard/stagemarkcalls.go`, `stats/internal/guard/loadsets_test.go`
**Allowed-collateral:** `skills/**/*.md`, `rules/*.mdc`, `commands-claude/**/*.md`, `README.md` — any sentence `check-verbatim-moves.sh` flags as reworded is acknowledged in this change's own `verbatim-moves.txt`, a planning artifact, never in this commit
**Tests:** none — repairs existing tests: `TestLoadSets`
**Regression:** reverting the commit restores the prompts inside flow only; spectre has nothing to vendor.

Correction (2026-10-25): the plan declared seven files; the commit also adds the skill's
`scripts/check-plan-shape.sh` and `scripts/lib` symlinks, which `check-guard-symlinks.sh` rule 2
requires once the skill names `check-plan-shape.sh`, declares the skill expected-zero in
`stagemarkcalls.go`, since it carries no stage mark of its own, and adds the skill to
`loadsets_test.go`'s fixture, whose load set the task extends.
**Commit:** feat(spectre-planning): one source for the generic planning prompts
**After:** none

**Decision:** source-in-agents-vendored
**Decision:** cursor-plan-mode-by-instruction
**Decision:** skill-name-spectre-planning

Correction (2026-10-25): Step 1 moved every listed paragraph verbatim, which left the skill
citing flow files and flow guards; the review panel's fix round moved each flow-dependent
paragraph back to `skills/flow/brainstorm-planner.md` and generalised the `/flow`-scoped
wording, per decision `planning-skill-flow-free`.

- [x] 2. spectre: vendor the skill and document the two modes
  - [x] **Step 1: In the spectre worktree `/Users/tweety53/Projects/spectre-worktrees/kan-1034-spectre-optional-bundled-planning-prompts-that` (branch `spectre/kan-1034-spectre-optional-bundled-planning-prompts-that`), copy `skills/spectre-planning/SKILL.md` from the agents worktree to `.claude/skills/spectre-planning/SKILL.md` byte for byte, and add `.claude/skills/spectre-planning/UPSTREAM.md` shaped like `.claude/skills/brainstorming/UPSTREAM.md`, naming the agents repo, the task 1 commit sha and "Unmodified".**
  - [x] **Step 2: Add `.cursor/rules/spectre-planning.mdc` (frontmatter `description`, `alwaysApply: false`) telling Cursor to read `.claude/skills/spectre-planning/SKILL.md` when planning a spectre change, Plan mode first.**
  - [x] **Step 3: README: a `## Two modes` section beside `## Bundled: brainstorming and writing-plans` — tool only (`go install`) vs tool + planning prompts (plugin install, or copy `.claude/skills/spectre-planning/`; on Cursor also `.cursor/rules/`) — and add `spectre-planning` to the bundled list.**
  - [x] **Step 4: Add `TestPluginShipsSpectrePlanning` to `plugin_manifest_test.go`: `.claude/skills/spectre-planning/SKILL.md` exists, its frontmatter `name` is `spectre-planning`, and `UPSTREAM.md` sits beside it. Verify: `gofmt -l . && go vet ./... && go test -run 'TestPluginShipsSpectrePlanning' -count=1 .` in the spectre worktree, and `diff` of the agents `skills/spectre-planning/SKILL.md` against the vendored copy prints nothing.**
**Build:** green
**Files:** `.claude/skills/spectre-planning/SKILL.md`, `.claude/skills/spectre-planning/UPSTREAM.md`, `.cursor/rules/spectre-planning.mdc`, `README.md`, `plugin_manifest_test.go`
**Tests:** `TestPluginShipsSpectrePlanning`
**Regression:** reverting the commit drops the vendored skill; `TestPluginShipsSpectrePlanning` no longer exists and the plugin ships no planning prompts.
**Baseline:** before=1 after=2 <!-- measured: grep -c '^func Test' plugin_manifest_test.go @ e093bad -->
**Commit:** feat(skills): vendor spectre-planning and document the two modes
**After:** Task 1

**Decision:** mode-by-install

- [x] 3. spectre: the manifest test's comment names what UPSTREAM.md now holds
  - [x] **Step 1: In the spectre worktree `/Users/tweety53/Projects/spectre-worktrees/kan-1034-spectre-optional-bundled-planning-prompts-that`, reword the comment above `TestPluginShipsSpectrePlanning` in `plugin_manifest_test.go`: "with the UPSTREAM.md that names its source commit" becomes "with the UPSTREAM.md that names its source path on agents main".**
  - [x] **Step 2: Verify: `gofmt -l . && go vet ./... && go test -run 'TestPluginShipsSpectrePlanning' -count=1 .` in the spectre worktree, then `git grep -n -F -e 'names its source commit' -- plugin_manifest_test.go` prints nothing.**
**Build:** green
**Files:** `plugin_manifest_test.go`
**Tests:** none — repairs existing tests: `TestPluginShipsSpectrePlanning`
**Regression:** reverting the commit restores a comment that misstates what UPSTREAM.md names.
**Commit:** fix(skills): say UPSTREAM.md names a source path
**After:** Task 2
