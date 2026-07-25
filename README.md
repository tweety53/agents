# agents-data

Portable agent configuration for the **Gymie** project.
Contains mandatory coding rules, an index of project-specific skills, and installation
instructions for every supported AI harness.

---

## What's in here

```
agents-data/
├── README.md                          ← this file
├── CLAUDE.md                          ← drop into project root for Claude Code
├── AGENTS.md                          ← drop into project root for Codex / OpenAI CLI
├── setup.sh                           ← helper: symlinks skills + rules + commands into harness dirs
├── rules/
│   └── myflow-manual-review.mdc       ← myflow stage boundaries (always-on)
├── commands/                          ← Cursor slash commands (myflow + opsx)
├── commands-claude/                   ← Claude Code slash commands (myflow only)
└── skills/                            ← project-specific OpenSpec / /myflow skills
    ├── README.md                      ← myflow command map
    ├── openspec-propose/
    ├── openspec-propose-superpowers/
    ├── openspec-apply-change/
    ├── openspec-apply-superpowers/   ← includes strict-panel reviewer prompts
    ├── openspec-apply-fix-superpowers/  ← fixes for Gate B/C findings, reuses the panel prompts above
    ├── openspec-manual-test-superpowers/
    ├── openspec-code-review-superpowers/  ← coverage check, tests/linters, commit + push + open PR (never merges)
    ├── openspec-archive-change/
    ├── openspec-archive-superpowers/  ← the "finish" stage: verify PR merged, sync specs, archive
    ├── openspec-full-cycle-superpowers/
    ├── openspec-explore/
    ├── openspec-sync-specs/
    ├── openspec-update-change/
    ├── myflow-status/                 ← read-only stage report for open changes
    └── myflow-info/                   ← reads the rule file and explains the pipeline
```

**Rules** (always-on, mandatory):
- Lint Fix Priority — never suppress/bypass ktlint or detekt
- Kotlin Backend Development Standard — module layout, dependency rules, checklist

**Skills** (loaded on demand):
- `/myflow-start`, `/myflow-do`, manual review (Gate B), `/myflow-do-fix` (Gate B/C/D fixes), `/myflow-manual-test` (Gate C, always asks whether to skip), `/myflow-code-review`, `/myflow-finish`, `/myflow-full`, `/myflow-status`, `/myflow-info` — OpenSpec + Superpowers with **manual review + manual test before commit**
- `/opsx:*` — lighter OpenSpec-only variants
- `/opsx:explore` — thinking-partner mode (no code)

**myflow pipeline (default):**

```text
start → do (#2–#6, staged + uncommitted) → manual review (Gate B, optional do-fix×N) → manual test (Gate C, optional do-fix×N) → code review (commit + push + open PR) → PR review (Gate D, human-merged) → finish (archive)
```

See `rules/myflow-manual-review.mdc` and `skills/README.md`.

---

## Quick installation per harness

### Cursor

Cursor reads rules from `.cursor/rules/` and skills from `.cursor/skills/`.
The gymie project already has these in place at their canonical locations.
Nothing to do — this repo is a reference/export copy only.

If you want to use these in a **different** Cursor project, run:

```bash
cd /path/to/other-project
./path/to/agents-data/setup.sh cursor
```

This symlinks `agents-data/skills/` into `.cursor/skills/`, `agents-data/rules/` into `.cursor/rules/`, and `agents-data/commands/` into `.cursor/commands/`.

---

### Claude Code

Claude Code reads `CLAUDE.md` from the project root and discovers skills from
`.claude/skills/` (when Superpowers is installed).

**Step 1 — Install Superpowers** (general workflow skills: brainstorming, TDD, etc.)

In a Claude Code session inside the project:
```
/plugin install prime-radiant-inc/superpowers
```

**Step 2 — Add project instructions**

```bash
cp /path/to/agents-data/CLAUDE.md /path/to/gymie/CLAUDE.md
```

**Step 3 — Add project-specific skills and slash commands**

```bash
cd /path/to/gymie
./path/to/agents-data/setup.sh claude-code
# or manually:
mkdir -p .claude/skills .claude/commands
for d in /path/to/agents-data/skills/*/; do
  ln -sf "$d" .claude/skills/
done
for f in /path/to/agents-data/commands-claude/*.md; do
  ln -sf "$f" .claude/commands/
done
```

Without `.claude/commands/`, `/myflow-*` typed in the CLI will fail with "Unknown command" —
Claude Code only auto-discovers skills by their `SKILL.md` `name:` (e.g. `openspec-apply-superpowers`),
not by the `/myflow-*` alias. The `commands-claude/*.md` files are thin wrappers that map the
`/myflow-*` name to the underlying skill.

**Verify**: In a new Claude Code session, ask: *"What project skills do you have?"*
The agent should be able to list and describe the `/myflow-*` skills, and typing `/myflow-do`
should resolve without an "Unknown command" error.

---

### Codex (OpenAI)

Codex reads `AGENTS.md` from the project root and discovers skills natively
when the Superpowers Codex plugin is installed.

**Step 1 — Install Superpowers for Codex**

The Superpowers Codex plugin is distributed from a separate fork repo.
Install it per the Superpowers README (look for the Codex install section).

Enable multi-agent support:
```toml
# ~/.codex/config.toml
[features]
multi_agent = true
```

**Step 2 — Add project instructions**

```bash
cp /path/to/agents-data/AGENTS.md /path/to/gymie/AGENTS.md
```

**Step 3 — Add project-specific skills**

```bash
cd /path/to/gymie
./path/to/agents-data/setup.sh codex
# or manually:
mkdir -p .codex/skills
for d in /path/to/agents-data/skills/*/; do
  ln -sf "$d" .codex/skills/
done
```

**Model note:** Codex has no per-skill/per-command model override mechanism — model is a session or profile-level setting (`~/.codex/config.toml`). Switch to a stronger model manually before invoking `openspec-propose-superpowers` (the `/myflow-start` equivalent); the rest of the pipeline is fine on your default.

---

### Gemini CLI

Gemini reads `GEMINI.md` from the project root and discovers skills via its extension
manifest's `contextFileName`.

**Step 1 — Install Superpowers for Gemini**

```bash
gemini extensions install prime-radiant-inc/superpowers
```

**Step 2 — Create GEMINI.md** with the rules inlined (same content as CLAUDE.md).

**Step 3 — Skills**: Gemini's Superpowers extension auto-discovers skills in its bundle.
For project-specific skills, symlink into `.gemini/skills/` if that path is recognized,
or inline the skill content into `GEMINI.md`.

---

## How skills work (no Superpowers installed)

If Superpowers is **not** installed, the agent can still use the project skills
by reading the `SKILL.md` file directly:

```
Read file: .claude/skills/openspec-propose-superpowers/SKILL.md
(follow the instructions in that file)
```

The `/myflow-*` skills internally reference Superpowers skills (brainstorming, TDD, etc.).
Without Superpowers those general skills won't auto-trigger, so the overall workflow is
degraded but the OpenSpec-specific steps still work.

---

## /myflow commands reference

`<name>` is **optional** on every command below — if omitted, the sole active (non-archived) change relevant to that stage is used automatically; if there are multiple, you're asked which.

**Model:** `/myflow-start` → Opus (enforced via frontmatter in Claude Code; switch manually in Cursor/Codex, which don't support per-command model selection yet). Every other command → Sonnet. `/myflow-full` runs as one session on Sonnet throughout, including Phase A — for brainstorming-heavy new work, run `/myflow-start` standalone first.

| Command | Skill file | What it does |
|---------|-----------|-------------|
| `/myflow-start <name>` | `openspec-propose-superpowers` | Brainstorm → design gate → artifacts → plan |
| `/myflow-do <name>` | `openspec-apply-superpowers` | Worktree → validate plan → SDD+TDD → **strict review panel** (primary+Bugbot+Security+Adversarial+Senior+Economic Senior) → **`git add` (staged; no commits)** |
| *(manual review)* | User (Gate B) | Inspect **staged** diff in worktree IDE |
| `/myflow-do-fix <name>` | `openspec-apply-fix-superpowers` | Fix a Gate B/C finding — document in `proposal.md`/`tasks.md` (append) or a linked nested `<name>-fix-N` sub-change (your choice) → resume the **same** worktree → SDD+TDD → full strict review panel re-run → staged; no commits. Loop at either gate as many rounds as needed. |
| `/myflow-manual-test <name>` | `openspec-manual-test-superpowers` | Gate C — write `docs/manual-test/<name>.md` (run apps + checklist); always asks whether to skip (default No); reply with link only |
| `/myflow-code-review <name>` | `openspec-code-review-superpowers` | Verify Gate C (or `SKIPPED`) → test coverage check (routes gaps to `/myflow-do-fix`) → tests/linters → **commit + push + open PR** (never merges) |
| *(PR review)* | User (Gate D) | Review and merge the PR on the forge — nothing in myflow merges |
| `/myflow-finish <name>` | `openspec-archive-superpowers` | Verify the PR merged → delta sync → archive (also archives nested `<name>-fix-N` sub-changes together) |
| `/myflow-full <name>` | `openspec-full-cycle-superpowers` | Full cycle: Gate A + Gate B + Gate C + code review, ending at Gate D (PR open, stop); `/myflow-finish` is always a separate human-initiated step |
| `/myflow-status <name>` | — (read-only) | Stage report for open changes |
| `/myflow-info` | — (read-only) | Reads the rule file and explains the pipeline |
| `/opsx:propose <name>` | `openspec-propose` | Lightweight: artifacts only, no Superpowers steps |
| `/opsx:apply <name>` | `openspec-apply-change` | Lightweight: implement tasks only |
| `/opsx:archive <name>` | `openspec-archive-change` | Lightweight: archive only |
| `/opsx:explore` | `openspec-explore` | Thinking-partner mode — no implementation |
| `/opsx:sync-specs <name>` | `openspec-sync-specs` | Sync delta specs to main specs |
| `/opsx:update <name>` | `openspec-update-change` | Revise planning artifacts, keep coherent |

**Flags:** `skip-propose`, `propose-only`, `skip-review`, `skip-manual-test` (pre-answers the Gate C skip prompt with Yes and must announce it), `commit-during-apply` (legacy per-task commits during apply)

All skills require the `openspec` CLI (`npm install -g openspec` or check project README).

---

## Keeping in sync

**Default (gymie):** Cursor auto-exports after agent file edits.

Gymie’s `.cursor/hooks.json` runs `.cursor/hooks/sync-agents-data.sh` on `afterFileEdit` when the edited path is under:

- `.cursor/skills/**`
- `.cursor/rules/myflow-*.mdc`
- `.cursor/commands/myflow-*.md` or `opsx-*.md`

Destination defaults to the sibling repo `../agents-data` (override with `AGENTS_DATA`). Sync is fail-open (never blocks the edit). Logs: `gymie/.cursor/hooks/logs/sync-agents-data.log`.

Manual full sync from gymie:

```bash
cd /Users/tweety53/Projects/gymie
.cursor/hooks/sync-agents-data.sh --force
# or: AGENTS_DATA=/path/to/agents-data .cursor/hooks/sync-agents-data.sh --force
```

`AGENTS.md` / `CLAUDE.md` myflow summary tables are **not** auto-rewritten — update those when the pipeline description changes.

Legacy one-shot rsync (same effect as `--force`):

```bash
cd /Users/tweety53/Projects/gymie
AGENTS=/Users/tweety53/Projects/agents-data

# Skills (full dirs — includes templates / reviewer prompts)
for skill_dir in .cursor/skills/*/; do
  name=$(basename "$skill_dir")
  mkdir -p "$AGENTS/skills/$name"
  rsync -a --delete "$skill_dir" "$AGENTS/skills/$name/"
done
cp .cursor/skills/README.md "$AGENTS/skills/README.md"

# Rules
mkdir -p "$AGENTS/rules"
cp .cursor/rules/myflow-manual-review.mdc "$AGENTS/rules/myflow-manual-review.mdc"

# Commands
mkdir -p "$AGENTS/commands"
cp .cursor/commands/myflow-*.md .cursor/commands/opsx-*.md "$AGENTS/commands/"

# Portable rule reference in agents-data skills
for f in openspec-apply-superpowers openspec-apply-fix-superpowers openspec-archive-superpowers openspec-code-review-superpowers openspec-full-cycle-superpowers openspec-manual-test-superpowers openspec-propose-superpowers; do
  sed -i '' 's|Also follow `.cursor/rules/myflow-manual-review.mdc`.|Also follow **rules/myflow-manual-review.mdc** (Cursor: `.cursor/rules/myflow-manual-review.mdc`).|g' \
    "$AGENTS/skills/$f/SKILL.md"
done

echo "Skills + rules + commands synced"
```
