# Every path citation in an installed file names its root

**Change:** `kan-102-citations-resolve-to-installed-paths`
**Jira:** KAN-102 — "Guard: a citation in an installed file must resolve to an installed path"
**Date:** 2026-08-20

## The problem

`setup.sh` installs this repository's `skills/*/`, `commands*/` and always-on `rules/*.mdc` into
other projects, and **copies** `CLAUDE.md` and `AGENTS.md` into their roots. It does not install
`README.md`, `openspec/`, `docs/`, `scripts/` or `stats/`.

A backticked path inside one of those installed files is read by an agent standing in the *target*
project. So `CLAUDE.md`'s `` `README.md` `` names the target project's README, not this
repository's — and the target project's README is a tracked, pull-request-editable file. The agent
reads content a contributor wrote, believing it to be canonical myflow documentation.

`scripts/check-references.sh` cannot see this. It resolves every citation against the agents
repository, where the target does exist — which is precisely the resolution a run never performs.

The corpus already carries half a convention: `<agents repo>/` marks a path in this checkout, used
at ten sites and given a resolution procedure under **Where the agents repository is**
(`skills/myflow-contracts/project-configuration.md`). Everything else is bare, and bare is
ambiguous.

## What is already fixed, and what is not

KAN-102 was filed as two halves. **Half 1 — executable citations — has already landed**, under
KAN-73. Guard scripts now ship inside `skills/*/scripts/`, a guard is named by basename and
resolves against the running command's own skill directory per **Guard resolution**
(`skills/myflow-contracts/pipeline.md`), and `scripts/check-guard-symlinks.sh` rules 2 and 3
enforce both halves of that: every guard a skill invokes has a symlink in its own `scripts/`
directory, and no repository-relative `scripts/<name>` path sits in an invoking position. The
arbitrary-code-execution route this ticket was raised to Highest for is closed.

**Half 2 — read-only citations — is live and unguarded.** Nothing checks that a path an installed
file merely *reads* resolves to something that was installed.

## The ambiguity that makes a guard non-trivial

"Does this path exist in the agents repository, and did `setup.sh` install it?" is not a usable
test, because two different kinds of path both answer the same way:

| Citation in an installed file | Intended meaning | Correct? |
|---|---|---|
| `README.md` | this repository's README | **no** — resolves to the target project's |
| `openspec/specs/myflow-model-policy/spec.md` | this repository's spec | **no** — the target project has no such file |
| `stats/cmd/myflow/stage.go` | this repository's source | **no** |
| `.myflow/project.md` | the target project's own configuration | **yes** |
| `openspec/changes/<name>/` | the target project's own change directory | **yes** |
| `docs/superpowers/` | the target project's own planning directory | **yes** |

Both kinds exist in this repository and neither is installed. Resolution alone cannot separate
them, because the difference is *intent*, and intent has to be written down.

## Decision: every citation names its root

Three roots, one of them implicit because it is already unambiguous:

| Form | Resolves against | Example |
|---|---|---|
| `skills/…`, `rules/…`, `commands/…`, `commands-claude/…`, `hooks/…` | the installed root, wherever the harness put it — `~/.claude/skills/`, `~/.cursor/skills/`, `~/.codex/skills/`, or a project's `.claude/skills/` | `skills/myflow-contracts/pipeline.md` |
| `<agents repo>/…` | this checkout, resolved by the two-step procedure already stated under **Where the agents repository is** (`skills/myflow-contracts/project-configuration.md`) | `<agents repo>/README.md` |
| `<project>/…` | the project the command is running against | `<project>/.myflow/project.md` |

**A path citation under none of the three is a violation.** The installed roots keep their bare
form for a concrete reason, not for convenience: they are the one class that already resolves
correctly in every harness, and rewriting the 554 of them would break
`scripts/check-references.sh`, whose entire coverage comes from resolving exactly those `.md`
paths against the repository root to confirm the cited section still exists.

`<project>/` is new vocabulary. It inverts the convention KAN-102's own text records — that an
unprefixed path is project-relative — deliberately: under the old convention a bare path was
correct by default and wrong by accident, with nothing able to tell the two apart.

## The guard

`scripts/check-installed-citations.sh` — argument-free and self-scoped from its own location, like
`check-references.sh` and `check-guard-symlinks.sh`, with `CHECK_INSTALLED_CITATIONS_ROOT` as an
opt-in override for the companion harness alone.

### Deriving the installed set

The set is derived by **running the installer**, never by re-implementing its globs:

```bash
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
HOME="$SANDBOX" ./setup.sh all "$SANDBOX/proj"
```

Then read back what appeared — every symlink resolved to its repository-relative source, plus the
copied `CLAUDE.md` and `AGENTS.md` — and take the first path segment of each as an installed root.

This is KAN-102's "derive the allowed-target set from `setup.sh` rather than hardcoding it, so the
guard cannot drift from what the installer does", taken literally. A glob re-implementation drifts
the first time an install path changes; a real run cannot. Measured cost: 0.6s per mode.

It also settles a case a glob would get wrong. `rules/kotlin-backend-development-standard.mdc` is
opt-in — `always_on_rules()` selects only `alwaysApply: true`, so the installer never links it —
and it carries roughly forty Java package paths (`core/domain/`, `infrastructure/persistence/`)
that a classifier cannot distinguish from directory citations. Deriving from a real run drops the
file from scope automatically, with no exception written anywhere.

The sandbox discipline is `scripts/test-setup.sh`'s, reused rather than reinvented: refuse to run
`setup.sh` unless both `HOME` and the project directory are inside the sandbox this run created.

### Classifying citations

Scope: each installed `.md`/`.mdc` file. For each backticked token, decide whether it is a path
citation, then whether it names a root.

A naive "contains a slash" test yields 723 hits on this corpus, roughly three-quarters of them
noise — absolute paths, `~/.claude/…`, URLs, `$SCRIPT_DIR/…`, `origin/main`, `refs/heads/…`,
regex fragments such as `[A-Za-z0-9._`, and glob shapes. Fenced `bash`/`sh`/`zsh` command lines
must come out as well: a path in `git -C <abs-worktree> reset -- openspec/` is a shell argument,
not a citation, and cannot carry a prefix.

That is block structure plus token shape — the same problem that moved `check-plan-provenance`
from a hand-rolled Bash ERE to a real parser. **The classifier is Python**, stdlib only, behind a
thin `.sh` wrapper that `exec`s it, matching the precedent `.myflow/project.md` already records
for that guard.

Slash-less tokens need one further rule, so `README.md` is caught without flagging every generic
mention of `SKILL.md` or `tasks.md`: **a bare filename is a citation only when it resolves to a
real file at the agents-repository root.** Verified against the corpus, that catches exactly
`README.md` and `setup.sh` and nothing else.

## Two existing guards need touching

- **`scripts/check-references.sh`** resolves backticked `.md`/`.mdc` paths against the repository
  root to verify the cited section still exists. After the rewrite,
  `<agents repo>/openspec/specs/…/spec.md` stops resolving and that coverage vanishes silently —
  the worst failure mode a guard has. It must strip a leading `<agents repo>/` before resolving.
  `<project>/…` paths cannot be resolved at all and fall through its existing
  unresolvable-path path.
- **`scripts/check-contract-budget.sh`** was checked and needs nothing. Roughly twelve bytes per
  site across 200 sites in 31 files is under 1% growth on the largest file, well inside the 25%
  headroom every budget row carries.

## The rewrite

200 sites across 31 files. By first segment: `scripts/` 49, `openspec/` 44, `.myflow/` 28,
`docs/` 22, `.superpowers/` 18, `specs/` 8, `setup.sh` 4, `agents/` 4, `.cursor/` 4, `stats/` 3,
`README.md` 3, and a tail of single sites. Heaviest files: `pipeline.md` 35,
`myflow-do/SKILL.md` 32, `finish-contract.md` 13, `AGENTS.md` 13, `CLAUDE.md` 11.

**`pipeline.md`'s Guard resolution section loses its prose exemption.** It currently states that a
`scripts/<name>` citation outside an invoking position is "prose describing a guard rather than
running it" and keeps its repository-relative path. Under the new convention there is no such
carve-out: prose about `<agents repo>/scripts/check-vocabulary.sh` says which repository it means,
and the classifier no longer has to distinguish invoking from describing. That deletion is the one
contract edit this change makes beyond mechanical prefixing.

## Testing

`scripts/test-check-installed-citations.sh`, fixture-tree driven like every sibling harness: one
fixture per classifier decision — a bare violation, each of the three roots, a fenced command
line, a URL, a `$VAR` path, a regex fragment, a slash-less root file, a slash-less generic
filename — plus the refusal cases: an unreadable root, and a sandbox escape attempt. Registered in
`.myflow/project.md`'s `## lint` and `## test` lists.

## Alternatives considered

- **A written list of project-relative roots** (`.myflow/`, `docs/superpowers/`,
  `openspec/changes/`, `.superpowers/`). Nine sites to fix instead of 200. Rejected: the list is
  hand-maintained, and it drifts the moment the pipeline writes into a new project directory —
  failing open, which is what a guard must never do.
- **A written list of agents-repo-only trees** (`README.md`, `openspec/specs/`, `stats/`). The
  same nine fixes, simpler to read. Rejected for the mirror-image reason: a new top-level
  directory in this repository is silently uncovered until somebody remembers the list.
- **Prefixing all three roots, including the 554 `skills/…` citations.** Maximum consistency,
  nothing implicit. Rejected: a ~700-site rewrite that breaks `check-references.sh`'s entire
  resolution mechanism and blows every `check-contract-budget.sh` row, buying nothing over an
  implicit root that is already unambiguous in every harness.

## Open questions

**Opt-in rules reach a project by a second route this guard does not cover.**
`install_project_standards` inlines an opt-in rule's *text* into the project's own managed block
rather than linking the file, so bare paths inside
`rules/kotlin-backend-development-standard.mdc` still land in a target project. Scoping the guard
to what `setup.sh` installs or copies is KAN-102's own wording and is what keeps the derivation
drift-proof; widening it to inlined text is a separate decision, recorded here for a follow-up
rather than made now.
