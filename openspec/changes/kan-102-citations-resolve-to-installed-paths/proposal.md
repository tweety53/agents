## Why

`setup.sh` installs this repository's skills, commands and always-on rules into other projects, and
copies `CLAUDE.md` and `AGENTS.md` into their roots. A backticked path inside one of those installed
files is read by an agent standing in the **target** project, so `CLAUDE.md`'s `` `README.md` ``
names that project's README — a tracked, pull-request-editable file — and not this repository's. The
agent reads content a contributor wrote as though it were canonical myflow documentation.

`scripts/check-references.sh` cannot see this: it resolves every citation against the agents
repository, where the target does exist. That is precisely the resolution a run never performs.

KAN-102's executable half already closed under KAN-73 — guard scripts ship inside `skills/*/scripts/`
and `scripts/check-guard-symlinks.sh` rules 2 and 3 keep repository-relative paths out of every
invoking position. The read-only half is live and unguarded, and 200 citations across 31 installed
files resolve to whatever repository the agent happens to be standing in.

## What Changes

- **BREAKING (authoring convention).** Every path citation in a file `setup.sh` installs or copies
  SHALL name its root. Three roots: the installed roots (`skills/`, `rules/`, `commands/`,
  `commands-claude/`, `hooks/`) keep their bare form because they already resolve correctly in every
  harness; `<agents repo>/` names this checkout; `<project>/` — new vocabulary — names the project
  the command is running against. A citation under none of the three is a violation.
- **New guard** `scripts/check-installed-citations.sh`, argument-free and self-scoped, which derives
  the installed set by **running the installer into a sandbox** rather than re-implementing its
  globs, then reports every citation in an installed file that names no root.
- **New harness** `scripts/test-check-installed-citations.sh`, fixture-driven like its siblings, both
  registered in `.myflow/project.md`'s `## lint` and `## test` lists.
- **`scripts/check-references.sh` strips a leading `<agents repo>/`** before resolving a cited path.
  Without this, the rewrite silently deletes that guard's coverage of every prefixed `.md` citation —
  a guard that stops checking without failing.
- **`skills/myflow-contracts/pipeline.md`'s Guard resolution section loses its prose exemption.** The
  carve-out letting a `scripts/<name>` mention outside an invoking position keep its
  repository-relative path is replaced by the `<agents repo>/` prefix, so the classifier no longer
  has to distinguish describing a guard from running one.
- **200 citations rewritten across 31 installed files**, heaviest in `pipeline.md` (35),
  `myflow-do/SKILL.md` (32), `finish-contract.md` (13), `AGENTS.md` (13) and `CLAUDE.md` (11).

## Capabilities

### New Capabilities

- `myflow-citation-roots`: every path citation in an installed file names its root, and a guard
  enforces it by deriving the installed set from the installer itself.

### Modified Capabilities

- `myflow-contract-distribution`: **A named guard resolves against the running command's own skill
  directory** currently exempts prose about this repository's own guards from the basename rule and
  lets it keep a repository-relative path. That exemption is replaced by the `<agents repo>/` prefix.
- `agents-repo-verification`: **A guard fails when a cross-referenced section no longer exists** must
  resolve a `<agents repo>/`-prefixed citation, and this repository's declared guard set gains the
  new guard and its harness.

## Impact

- `scripts/check-installed-citations.sh` and `scripts/check-installed-citations.py` (new),
  `scripts/test-check-installed-citations.sh` (new).
- `scripts/check-references.sh` — prefix stripping before resolution.
- `.myflow/project.md` — `## lint` and `## test` entries.
- 31 installed files under `skills/`, `rules/`, `commands*/`, plus `CLAUDE.md` and `AGENTS.md`.
- `scripts/check-contract-budget.sh` — measured. No row is raised, but `skills/myflow-do/SKILL.md`
  is tight: about 355 added bytes against 549 of headroom. Every other row is comfortable.
- No runtime behaviour changes. Every edit is to documentation text or to a lint-time guard.
