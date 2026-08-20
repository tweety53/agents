## Context

`setup.sh` installs `skills/*/`, `commands*/` and always-on `rules/*.mdc` into other projects, and
copies `CLAUDE.md` and `AGENTS.md` into their roots. It installs neither `README.md`, `openspec/`,
`docs/`, `scripts/` nor `stats/`.

A backticked path inside an installed file is read by an agent standing in the target project. So
`` `README.md` `` in `CLAUDE.md` names that project's README, which a contributor can write, and not
this repository's. `scripts/check-references.sh` cannot see the defect: it resolves citations against
the agents repository, where the target does exist — the one resolution a run never performs.

KAN-102's executable half closed under KAN-73. Guards now ship inside `skills/*/scripts/`, resolve by
basename against the running command's own skill directory, and `scripts/check-guard-symlinks.sh`
rules 2 and 3 keep repository-relative paths out of every invoking position. What remains is the
read-only half: 200 citations across 31 installed files that name no root at all.

## Goals / Non-Goals

**Goals**

- Every path citation in an installed file states which root it resolves against.
- A guard enforces that, and derives what "installed" means from the installer itself.
- `scripts/check-references.sh` keeps every unit of coverage it has today.

**Non-Goals**

- Re-opening the executable half. KAN-73 closed it and this change does not revisit it.
- Covering opt-in rules, whose text reaches a project by inlining rather than by installation. See
  **Open questions**.
- Any runtime behaviour change. Every edit here is documentation text or a lint-time guard.

## Decisions

### Every citation names its root, rather than a list deciding which bare paths are safe

**ID:** `citation-names-its-root`
**Status:** active
**Chosen:** three recognised roots — the installed roots bare, `<agents repo>/`, `<project>/` — with
any citation under none of them a violation. Nothing about the rule is hand-maintained, so it cannot
fail open.
**Considered:**
- *A written list of project-relative roots* (`.myflow/`, `docs/superpowers/`, `openspec/changes/`,
  `.superpowers/`), with anything else that exists here and is not installed requiring the prefix.
  Nine sites to fix instead of 200 — ruled out because the list is maintained by hand and drifts the
  moment the pipeline writes into a new project directory. It fails open, which is what a guard must
  never do.
- *A written list of agents-repo-only trees* (`README.md`, `openspec/specs/`, `stats/`). Same nine
  fixes, easier to read — ruled out for the mirror-image reason: a new top-level directory here is
  silently uncovered until somebody remembers to add it.

### The installed roots keep their bare form

**ID:** `installed-roots-stay-bare`
**Status:** active
**Chosen:** `skills/`, `rules/`, `commands/`, `commands-claude/` and `hooks/` are recognised roots
and need no prefix. They are the one class already unambiguous in every harness, since that is
exactly where the harness puts them.
**Considered:** *prefixing all three roots*, so that nothing is implicit — 554 further citations
rewritten as `<skills>/myflow-contracts/pipeline.md`. Ruled out on two concrete costs, not on size
alone: `scripts/check-references.sh` derives its **entire** coverage from resolving those `.md`
paths against the repository root, so prefixing them deletes that guard and requires rebuilding it;
and every `skills/*/SKILL.md` and contract file would blow its `scripts/check-contract-budget.sh`
row at once, turning a ratchet meant to catch section growth into a formality raised in bulk.

### The installed set is derived by running the installer

**ID:** `derive-by-running-setup`
**Status:** active
**Chosen:** run `setup.sh` into a sandbox and read back what appeared — each symlink resolved to its
repository-relative source, plus each copied file. Measured at 0.6s per mode. This is KAN-102's
"derive the allowed-target set from `setup.sh` rather than hardcoding it" taken literally: a new
install path is covered the day it lands, with no edit to the guard.
**Considered:**
- *Re-implementing the installer's globs inside the guard* — cheaper per run, and wrong the first
  time `setup.sh` changes. The drift would be silent: the guard keeps passing while scanning the
  wrong set.
- *A written list of installed paths* — the same drift, arriving sooner.

A second property settles a real case rather than a hypothetical one.
`rules/kotlin-backend-development-standard.mdc` is opt-in, so `always_on_rules()` never selects it
and the installer never links it. It carries roughly forty Java package paths — `core/domain/`,
`infrastructure/persistence/`, `ports/` — that no classifier can distinguish from directory
citations. Deriving from a real run drops the file from scope automatically, with no exception
written anywhere for a maintainer to later wonder about.

### The classifier is Python, behind a shell wrapper

**ID:** `classifier-in-python`
**Status:** active
**Chosen:** `scripts/check-installed-citations.sh` execs `scripts/check-installed-citations.py`,
stdlib only, matching the precedent `scripts/check-plan-provenance.py` set and `.myflow/project.md`
records.
**Considered:** *a Bash implementation*, consistent with most guards here. Ruled out on measurement:
a naive "contains a `/`" test yields 723 candidates on this corpus and roughly three quarters are not
citations — absolute paths, `~/.claude/…`, URLs, `$SCRIPT_DIR/…`, `origin/main`, `refs/heads/…`,
regex fragments such as `[A-Za-z0-9._`, glob shapes. Separating them needs fenced-block structure as
well as token shape, which is the same problem that moved `check-plan-provenance` out of a
hand-rolled ERE after five review passes and seven fix waves.

### A slash-less token is a citation only when it resolves at the repository root

**ID:** `bare-filename-must-resolve`
**Status:** active
**Chosen:** a backticked token with no `/` counts as a citation only if a real file of that name sits
at the agents-repository root. Verified against the corpus: this catches exactly `README.md` and
`setup.sh`, and nothing else.
**Considered:** *ignoring slash-less tokens entirely* — simpler, and it misses `README.md`, which is
the ticket's own headline example and its most severe one, since `CLAUDE.md` auto-loads into every
session of every project it is copied into. *Flagging every slash-less `.md` token* — flags every
generic mention of `SKILL.md`, `tasks.md` and `proposal.md`, which are shapes, not files.

### `check-references.sh` strips the prefix rather than being left to lose coverage

**ID:** `check-references-strips-prefix`
**Status:** active
**Chosen:** strip a leading `<agents repo>/` before resolving, so a citation that names this checkout
is checked exactly as it was before it carried a root.
**Considered:** *leaving it alone*, on the grounds that the guard would simply skip the prefixed
paths. Ruled out because that is the specific shape of failure a guard must never have: it would
keep exiting `0` while checking strictly less than before, and nothing would report the loss.

## Risks / Trade-offs

- **A classifier that is too eager turns the guard into noise, and one too lax leaves the defect.**
  Mitigated by the fixture harness carrying one fixture per decision — each excluded shape and each
  included shape — so a change to the classifier that moves either boundary fails a named test rather
  than silently changing the corpus verdict.
- **`<project>/` is new vocabulary, and it inverts the convention KAN-102's own text records** — that
  an unprefixed path is project-relative. That is deliberate: under the old convention a bare path
  was correct by default and wrong by accident, with nothing able to tell the two apart. The cost is
  that anyone reading an older archived change sees the earlier convention; archived changes are not
  rewritten.
- **The guard executes `setup.sh` on every lint run.** Bounded at 0.6s per mode and confined by
  `scripts/test-setup.sh`'s existing refusal — both `HOME` and the project directory must lie inside
  the sandbox that run created, or the guard refuses rather than proceeding.
- **`scripts/check-contract-budget.sh` was measured, and one row is tight.** The added bytes are
  15 per `<agents repo>/` site and 10 per `<project>/` site. `skills/myflow-do/SKILL.md` carries 32
  sites, worth about 355 bytes, against **549 bytes** of remaining headroom — it fits, with roughly
  190 bytes to spare, and any prose added to that file in the same change trips the ratchet. Every
  other row is comfortable: `pipeline.md` adds about 435 bytes against 9512 of headroom,
  `finish-contract.md` 140 against 4232. `CLAUDE.md` and `AGENTS.md` carry no budget row at all.
  Should the `myflow-do` row trip, the correct response is to raise that one row — a genuine
  addition is exactly what `.myflow/project.md` says a raise is for — and never to narrow the guard
  or delete a row.

## Migration Plan

The rewrite is mechanical and lands with the guard that enforces it, so the repository never sits in
a state where the convention is declared and unchecked. Order: the guard and its harness first,
red against the current corpus; then the rewrite, which turns it green; then
`scripts/check-references.sh`'s prefix stripping, whose own harness proves the coverage survived.

## Open questions

### Opt-in rules reach a project as inlined text, not as installed files

**ID:** `opt-in-rules-inlined`
**Status:** open
**Why it is open:** `install_project_standards` inlines an opt-in rule's text into a project's own
managed block rather than linking the file, so bare paths inside
`rules/kotlin-backend-development-standard.mdc` still reach a target project by that second route.
Scoping the guard to what `setup.sh` installs or copies is KAN-102's own wording, and it is what
keeps the derivation drift-proof; widening it to inlined text is a separate decision, deferred by the
operator rather than settled here.
**What it affects:** whether the guard's scan set is "files the installer ships" or "text the
installer causes to appear in a project". The second is a larger set, needs a different derivation,
and would pull the roughly forty Java package paths in that rule back into scope along with the
classifier problem they pose.
