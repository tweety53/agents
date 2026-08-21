# Apply the be-brief standard to this repository's Markdown

## Why

This repository requires brevity of every agent it dispatches and applies none to its own text.

The owned corpus — every `.md` / `.mdc` under `skills/`, `rules/`, `openspec/specs/`, `commands/`,
`commands-claude/`, `.myflow/` and the repository root (non-recursive), with `node_modules/`,
`.superpowers/`, `openspec/changes/archive/` and `docs/superpowers/` excluded — is
**92 files, 1.22 MB**. It is defined by those roots, not by subtracting exclusions from the whole
tree: `docs/self-review/` and `stats/` fall outside it.
<!-- measured: find . \( -name "*.md" -o -name "*.mdc" \) excluding node_modules/.superpowers/archive/docs, xargs wc -c @ branch main -->
`scripts/check-contract-budget.sh` ratchets **13** of them; the other **79**, totalling **1.16 MB**, have nothing holding them back.
<!-- measured: sed -n '/budgets()/,/^EOF/p' scripts/check-contract-budget.sh | grep -c '\.md ' @ branch main -->

`skills/myflow-do/SKILL.md` alone is **71 KB** of text loaded on every `/myflow-do` and
`/myflow-fast` run.
<!-- measured: wc -c skills/myflow-do/SKILL.md @ branch main -->

Regrowth is the failure mode, not any single bloated file: one unremarkable paragraph at a time,
with no guard that notices.

## What changes

- **A standard**, as a clause in `rules/be-brief.mdc`, naming two subjects: this repository's own
  Markdown, and the artifacts a `/myflow-*` run generates in any project. The file's existing
  "docs and specs stay full" sentence is reworded so completeness and non-repetition stop reading
  as contradictory.
- **A widened guard.** `scripts/check-contract-budget.sh` ratchets every owned file rather than 13,
  with its scan root extended past `skills/`.
- **A proof.** A script that inventories every SHALL / SHALL NOT / MUST / MUST NOT sentence in the
  corpus, so the trim can be shown to have dropped none of them.
- **The trim itself**, cuts only — restatement, duplication, hedging, meta-prose. No paraphrase.

## Impact

- Affected specs: `myflow-contract-economy`, `agents-repo-verification`
- Affected code: `rules/be-brief.mdc`, `scripts/check-contract-budget.sh`, a new inventory script
  and its harness, `.myflow/project.md`, and every owned Markdown file
- `setup.sh global` must be re-run: the reworded rule renders into the managed block of
  `~/.claude/CLAUDE.md` and `~/.codex/AGENTS.md`

## Non-goals

- Rewriting for density — a requirement re-said differently is a requirement changed
- New core/rationale partitions
- Other projects' existing documentation
- Any reduction target
