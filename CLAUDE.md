# Gymie — Agent Instructions (Claude Code)

This file is the active instruction set for Claude Code sessions on the Gymie project.
It contains mandatory rules and an index of project-specific skills.

---

## Mandatory Rules

### Lint Fix Priority

Fixing lint violations is **mandatory** and takes **priority** over finishing unrelated work.
Do not bypass checks to ship faster.

1. **Fix, don't bypass** — Refactor or reformat code to satisfy ktlint, detekt, and other project lint rules.
2. **No new suppressions** — Do not add `@Suppress`, `@SuppressWarnings`, `// detekt:ignore`, or ktlint disable comments unless the user explicitly asks for that approach.
3. **No config weakening** — Do not disable rules or lower thresholds in `detekt.yml`, `.editorconfig`, ESLint config, or Gradle lint settings to make checks pass.
4. **Verify before completion** — Run `./gradlew ktlintCheck detekt` (or the repo-equivalent lint command) before claiming work is done.
5. **Auto-fix first** — Use `./gradlew ktlintFormat` for formatting issues before manual fixes.

Pre-approved suppressions and config deviations live in `CONTRIBUTING.md`. Do not expand that list without user approval.

```kotlin
// BAD — suppressing instead of fixing
@Suppress("LongMethod")
fun processOrder() { /* 200 lines */ }

// GOOD — extract helpers until detekt passes
fun processOrder() {
    validateOrder()
    chargePayment()
    notifyCustomer()
}
```

When lint fails on your changes, stop and fix it first.

---

### Kotlin Backend Development Standard

Applies to all `src/**/*.kt` and `**/*.kts` files.

Follow [JetBrains Kotlin coding conventions](https://kotlinlang.org/docs/coding-conventions.html).
`./gradlew ktlintCheck detekt` must pass with zero violations.

#### Gradle module layout

| Path | Module | Role |
|------|--------|------|
| `src/app/` | `:app` | Composition root, Flyway migrations, Boot entry point |
| `src/gateway/` | `:gateway` | Public edge — routing, JWT validation, CORS |
| `src/app/<domain>/` | `:auth`, … | Domain library — `core` + `infrastructure` |
| `src/app/common/` | `:common` | Shared cross-cutting types and filters |

#### Domain module layout (`core` + `infrastructure`)

```
com/gymie/auth/
├── core/                                    # business logic ONLY
│   ├── domain/                              # domain models
│   │   └── <domainName>/
│   │       └── specification/
│   ├── exceptions/                          # business exceptions
│   ├── ports/                               # outbound interfaces
│   │   ├── persistence/<domainName>/
│   │   ├── jwt/
│   │   ├── token/
│   │   └── security/
│   └── services/
│       └── <serviceName>/
│           ├── <Service>.kt                 # interface
│           └── impl/<Service>Impl.kt
└── infrastructure/                          # everything non-business
    ├── configurations/
    ├── integrations/
    │   ├── redis/
    │   └── jwt/
    ├── persistence/
    │   ├── dao/
    │   ├── entities/
    │   └── converters/
    ├── presentation/web/
    │   ├── dto/
    │   ├── exceptions/
    │   ├── validators/
    │   ├── converters/
    │   └── security/
    └── scheduler/
```

#### Dependency rules

- **`core`** operates on `core.domain` models and **`core.ports`** interfaces only.
- **`core.services`** depend only on `core` packages.
- **`infrastructure`** implements `ports`, hosts HTTP, DB, caches, and Spring wiring.
- **`infrastructure.presentation`** converts DTOs ↔ domain before/after calling `core.services`.
- **Core must not import Spring, JPA, Redis, HTTP, or JWT types.**

```
presentation/web → core.services → core.ports ← infrastructure.persistence / integrations
                      ↓
                 core.domain
```

Do **not** use top-level `application/`, `api/`, `dto/`, `config/`, or legacy `service/` packages in domain modules.

#### Agent checklist

1. Business logic? → `core/services/<name>/impl/`
2. Domain model? → `core/domain/`
3. Business error? → `core/exceptions/`
4. Outbound contract? → `core/ports/`
5. DB / cache / JWT / HTTP / config? → matching `infrastructure/` subtree
6. Run `./gradlew ktlintCheck detekt :auth:test :app:test`

---

## Project Skills (OpenSpec / /myflow workflow)

These skills live in `skills/` next to this file (or in `.claude/skills/` if installed there).
To invoke a skill: **read its `SKILL.md` file** then follow the instructions within.

All skills require the `openspec` CLI to be installed.

### Skill index

| Skill directory | Trigger | Purpose |
|-----------------|---------|---------|
| `skills/openspec-propose/` | `/opsx:propose` | Propose a change — create all planning artifacts in one step |
| `skills/openspec-propose-superpowers/` | `/myflow-start` | Propose with Superpowers Basic Workflow #1 (brainstorming) + #3 (writing-plans) |
| `skills/openspec-apply-change/` | `/opsx:apply` | Implement tasks from an OpenSpec change |
| `skills/openspec-apply-superpowers/` | `/myflow-do` | Implement with Superpowers #2–#6 (**stage with `git add`**; no commits; Gate B then Gate C next) |
| `skills/openspec-apply-fix-superpowers/` | `/myflow-do-fix` | Fix a Gate B / C / D finding — document in proposal (append or nested) → Superpowers #4–#6 in the **existing** worktree. Stages only at Gate B and C (no commits); **commits and pushes to the PR branch at Gate D** |
| `skills/openspec-manual-test-superpowers/` | `/myflow-manual-test` | Gate C — write `docs/manual-test/<name>.md`, reply with link only |
| `skills/openspec-archive-change/` | `/opsx:archive` | Archive a completed change (delta sync + move to archive) |
| `skills/openspec-review-superpowers/` | `/myflow-review` | Verify Gate C → coverage check → tests/linters → **commit + push + open PR** (never merges) |
| `skills/openspec-archive-superpowers/` | `/myflow-finish` | Verify the PR merged → delta sync → archive (also archives nested `<name>-fix-N` sub-changes) |
| `skills/openspec-full-cycle-superpowers/` | `/myflow-full` | start → do → Gate B review → Gate C manual test → review (commit+push+PR) → Gate D, stop |
| `skills/openspec-explore/` | `/opsx:explore` | Thinking-partner mode — explore ideas, investigate, no implementation |
| `skills/openspec-sync-specs/` | `/opsx:sync-specs` | Sync delta specs from a change to main specs |
| `skills/openspec-update-change/` | `/opsx:update` | Revise existing planning artifacts; keep them coherent |
| `skills/myflow-status/` | `/myflow-status` | Read-only stage report for open changes |
| `skills/myflow-info/` | `/myflow-info` | Reads the rule file and explains the pipeline |

### /myflow commands summary

**Pipeline:** `start → do (#2–#6) → manual review (Gate B, optional do-fix×N) → manual test (Gate C, optional do-fix×N) → review (commit + push + open PR) → (human) PR review + merge (Gate D) → finish (archive)`

Also follow `rules/myflow-manual-review.mdc` (always-on stage boundaries).

`<name>` is **optional** on every command below — if omitted, the sole active (non-archived) change relevant to that stage is used automatically; if there are multiple, you're asked which.

**Model:** `/myflow-start` → **Opus** (brainstorming/design benefits from stronger reasoning). Every other command → **Sonnet**. In Claude Code this is enforced via `model:` frontmatter on each command; Cursor and Codex don't support per-command model selection yet, so switch manually — see "Model policy" in `rules/myflow-manual-review.mdc`. Note `/myflow-full` runs as one session on Sonnet throughout, including Phase A — for brainstorming-heavy new work, run `/myflow-start` standalone first.

| Command | What it does |
|---------|-------------|
| `/myflow-start <name>` | Brainstorm → design approval gate → OpenSpec artifacts → writing-plans enriched tasks |
| `/myflow-do <name>` | git worktree → validate plan → SDD + TDD → **strict review panel** (primary + Bugbot + Security + Adversarial + Senior + Conventions) → **`git add -A` (staged + uncommitted; no #7)** |
| *(Gate B)* | **You** open the worktree in IDE and review staged changes (`git diff --cached`) |
| `/myflow-do-fix <name>` | Fix something found at Gate B (manual review), Gate C (manual test), or Gate D (PR review) — documents it in `proposal.md`/`tasks.md` (or a linked nested `<name>-fix-N` sub-change, your choice) → resumes the **same** worktree → SDD + TDD → strict review panel (targeted re-run by default; always full for Gate D origins or with `full-panel`). **Gate B and C: staged, no commits. Gate D: commits and pushes to the PR branch** (the one place this command commits; it still never merges). Loop as many rounds as needed at any of the three gates. |
| `/myflow-manual-test <name>` | Write `docs/manual-test/<name>.md` (run apps + checklist); always asks whether to skip Gate C (default No); reply with **link only** |
| *(Gate C)* | **You** run the apps and check off items in the guide |
| `/myflow-review <name>` | Verifies every Gate C box is checked (or the guide is marked `SKIPPED`) — checks test coverage against delta specs (routes gaps to `/myflow-do-fix`) — then tests/linters → **commit + push + open PR** (never merges) |
| *(Gate D)* | **You** review the PR and merge it |
| `/myflow-finish <name>` | Verifies the PR merged → delta sync → archive (also archives any nested `<name>-fix-N` sub-changes together) |
| `/myflow-full <name>` | Full cycle with Gate A (proposal) + Gate B (review) + Gate C (manual test) + review, ending at Gate D (PR open, stop) |
| `/myflow-status <name>` | Read-only stage report for open changes |
| `/myflow-info` | Reads the rule file and explains the pipeline |

**Flags:** `skip-propose`, `propose-only`, `skip-review`, `skip-manual-test` (now pre-answers the Gate C skip prompt with Yes and must announce it), `full-panel` (forces all six review agents over the whole-branch diff on every re-run, instead of the default targeted re-run), `commit-during-apply` (legacy)

### How to invoke a skill

Read the skill file, then follow it:

```
Read file: skills/openspec-propose-superpowers/SKILL.md
(then follow the instructions in that file)
```

### Superpowers general skills

The Superpowers plugin provides general-purpose workflow skills (brainstorming, TDD,
subagent-driven-development, etc.). These are referenced by the `/myflow-*` skills above.

Install Superpowers in Claude Code:
```
/plugin install prime-radiant-inc/superpowers
```

After install, general skills auto-trigger from their descriptions. Project-specific `/myflow-*`
skills are loaded on demand by reading their `SKILL.md` as described above.
