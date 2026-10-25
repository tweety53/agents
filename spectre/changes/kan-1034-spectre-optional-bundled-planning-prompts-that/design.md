## Context

- The planning prompts sit in `skills/flow/brainstorm-planner.md` B–D, `skills/flow/brainstorm.md` (**Plan mode on Claude Code**) and `skills/flow-plan/SKILL.md`, mixed with flow-only lines paragraph by paragraph.
- spectre (`/Users/tweety53/Projects/spectre`, `github.com/tweety53/spectre`) is a separate repo. It already ships a Claude Code plugin (`.claude-plugin/plugin.json`, skills from `.claude/skills/`) vendoring superpowers `brainstorming` and `writing-plans` with an `UPSTREAM.md` each.
- `check-verbatim-moves.sh` and `check-references.sh` scan only this repo's `skills/`, `rules/`, `commands-claude/`, so the single source must live here for them to prove the move.
- Cursor support was removed from this repo in kan-843; Cursor's Plan mode is a UI mode an agent may not be able to enter itself.

## Decisions

### Single source in the agents repo, vendored by spectre

**ID:** source-in-agents-vendored
**Status:** active
**Chosen:** `skills/spectre-planning/SKILL.md` here; flow loads it; spectre copies it into `.claude/skills/spectre-planning/` with an `UPSTREAM.md` naming the agents commit — the pattern spectre already uses for superpowers.
**Considered:** source in the spectre repo with flow citing the installed copy — the verbatim-move and reference guards cannot resolve across repos, so every moved sentence fails and every citation dangles.

### Mode is chosen by what is installed

**ID:** mode-by-install
**Status:** active
**Chosen:** the binary alone is tool-only; installing the plugin (or copying its skills) is the prompts mode — no CLI change.
**Considered:** `spectre init --with-prompts` writing the skill files — new Go code and tests for what the plugin install already does.

### Cursor plan mode by instruction

**ID:** cursor-plan-mode-by-instruction
**Status:** active
**Chosen:** the skill tells an agent on Cursor to switch to Plan mode before exploring, or to ask the user to switch when it cannot; spectre ships `.cursor/rules/spectre-planning.mdc` pointing Cursor at the skill. No Cursor harness in `setup.sh`.
**Considered:** reviving a Cursor harness in `setup.sh` so flow runs on Cursor — out of the ticket's scope, which is spectre-only use.

### Skill name `spectre-planning`

**ID:** skill-name-spectre-planning
**Status:** active
**Chosen:** `spectre-planning` — names the tool it serves and the job.
**Considered:** `planning-prompts` — collides in meaning with `flow-plan` and does not say which tool it belongs to.

## Open questions

## Recommended approach

Move the generic paragraphs verbatim into `skills/spectre-planning/SKILL.md` (decision `source-in-agents-vendored`), generalising only the plan-mode paragraph to name Claude Code and Cursor (decision `cursor-plan-mode-by-instruction`). Each flow file keeps its flow-only lines and replaces each moved passage with a `Load` directive or a `(**Heading**, `skills/spectre-planning/SKILL.md`)` citation; citations of moved headings elsewhere are re-pointed. spectre vendors the file, documents the two modes (decision `mode-by-install`) and asserts in its manifest test that the skill ships. One approach followed from the operator's answers, so no competing Plan agents ran.

## Critical files

- `skills/spectre-planning/SKILL.md` — new; the single source of the generic planning prompts.
- `skills/flow/brainstorm-planner.md` — loses the generic paragraphs of B–D, keeps flow-only ones, loads the new skill.
- `skills/flow/brainstorm.md` — its plan-mode paragraph moves; keeps the flow-only clause.
- `skills/flow-plan/SKILL.md` — its plan-mode paragraph cites the new skill.
- `skills/**/*.md`, `rules/*.mdc`, `README.md` — citations of moved headings re-pointed.
- `stats/internal/guard/loadsets.go` — planning load set gains the new file.
- `CLAUDE.md`, `AGENTS.md` — skill index row.
- spectre: `.claude/skills/spectre-planning/SKILL.md`, `.claude/skills/spectre-planning/UPSTREAM.md`, `.cursor/rules/spectre-planning.mdc`, `README.md`, `plugin_manifest_test.go`.

## Existing code to reuse

- spectre's vendoring pattern — `.claude/skills/brainstorming/UPSTREAM.md`.
- `scripts/check-verbatim-moves.sh` — proves the text moved rather than changed.
- `scripts/check-references.sh` — proves every re-pointed citation resolves.
- `stats/internal/guard/planshape.go` rule F19 — unchanged; the skill names its expectation.
- `setup.sh` `install_skills` — installs every `skills/*` directory, so the new skill needs no installer change.

## Verification

- agents: every `## lint` line of `.flow/project.md`, `scripts/check-verbatim-moves.sh`, and `cd stats && make -s web-dist && go vet ./... && gofmt -l . && go test ./internal/guard/...`.
- spectre: `go test ./...`; `diff` of the vendored `SKILL.md` against the agents source prints nothing.
- Scenario: in a scratch spectre tree, `spectre new demo`, fill `design.md` per the skill with a `tasks.md` carrying `> **Design shape:** plan-mode`; `scripts/check-plan-shape.sh` passes, and drops to a `missing section` hit when one plan section is removed.

## Live check

None — markdown skills and a test; no running service or persistent state changes.
