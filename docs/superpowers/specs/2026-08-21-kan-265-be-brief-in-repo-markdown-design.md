# KAN-265 — be-brief in repo Markdown

**Jira:** KAN-265 — Apply the be-brief rule to every Markdown file in the agents repo
**Change:** `kan-265-be-brief-in-repo-markdown`
**Date:** 2026-08-21

## Problem

This repository asks brevity of every agent it dispatches and applies none to its own text. The
owned corpus — excluding `node_modules/`, `.superpowers/`, `openspec/changes/archive/` and
`docs/superpowers/` — is **120 files, 1.46 MB**:

| Area | Bytes | Ratcheted today |
|------|-------|-----------------|
| `skills/**` | 604 KB | 13 files, 219 KB |
| `openspec/specs/**` | 462 KB | none |
| Root `.md` | 71 KB | none |
| `rules/*` | 45 KB | none |
| `commands*/` | 28 KB | none |
| `.myflow/project.md` | 15 KB | none |

`scripts/check-contract-budget.sh` ratchets 13 files. The remaining **107 files, 1.24 MB** have
nothing holding them back, and `skills/myflow-do/SKILL.md` — 71 KB, loaded on every `/myflow-do` and
`/myflow-fast` run — is the single largest owned file.

`openspec/specs/myflow-contract-economy/spec.md` already governs the core/rationale split and the
byte budget, but only for `skills/myflow-contracts/`. Nothing covers the rest.

## The tension this change has to resolve

`rules/be-brief.mdc` is written about what an agent **says**, and exempts written files explicitly:

> **Prose only** — code, commits, docs and specs stay full.

Read literally, applying that rule to documentation inverts its own carve-out. The resolution is
that "full" and "brief" are answering different questions:

- **full** = completeness. Every requirement, scenario, exit code and worked example is present.
- **brief** = non-repetition. Nothing is said that a reader already has from a canonical statement
  elsewhere.

A file can be both, and every file here is required to be. The clause added below states this, and
the existing sentence is reworded so the two stop reading as contradictory.

## Decisions

### D1 — Scope is the whole owned corpus

All 120 owned `.md` / `.mdc` files, `README.md` and `CONTRIBUTING.md` included.

Excluded, and excluded structurally rather than by a list that goes stale: `node_modules/` (not
ours), `.superpowers/` (SDD run snapshots, historical records of what a file said at a moment),
`openspec/changes/archive/` (archived changes are a record, and editing one rewrites history), and
`docs/superpowers/` (session records, same reason).

### D2 — The standard is a clause in `rules/be-brief.mdc`, scoped to two subjects

Not a new spec, and not a widening of `myflow-contract-economy`. The clause names exactly two
subjects:

1. **This repository's own Markdown.** Named explicitly. `be-brief.mdc` installs globally into
   every project's `CLAUDE.md`, and a general rule about "a repository's documentation" would
   silently govern projects nobody evaluated it against.
2. **The artifacts a `/myflow-*` run generates, in any project** — `proposal.md`, `design.md`,
   `tasks.md`, delta specs, panel records, self-review reports. These are written by this
   repository's own pipeline wherever they land, so the standard travels with the writer rather
   than with the repository.

Other projects' pre-existing documentation is untouched.

### D3 — Cuts only, never paraphrase

A passage may be **deleted**. It may not be **reworded to say the same thing in fewer words**.

Paraphrase is the failure mode that matters: these files are the pipeline's runtime source, and a
requirement re-said slightly differently is a requirement changed, with nothing to detect it. A
deletion is visible in a diff and provable against the inventory in D5.

In scope for deletion:

- a passage restating what another file states canonically, where a citation already points there
- a rule stated twice within one file
- hedging that adds no constraint
- meta-prose about the document rather than its subject

Never deleted:

- any normative sentence (SHALL / SHALL NOT / MUST / MUST NOT)
- an exit-code contract, an ordering constraint, a scenario
- a worked example
- a recorded reason a rejected alternative was rejected — that is what stops it being re-proposed

### D4 — The guard is the existing ratchet, widened

`scripts/check-contract-budget.sh` grows its per-file table from 34 rows to one row per owned file,
and its scan root widens past `skills/` to the areas in D1.

Hand-maintained, as today. A generator would make "just regenerate it" the path of least resistance,
which is a ratchet that does not ratchet. An undeclared file remains a violation, so a file added
later cannot slip in unbudgeted.

A byte budget cannot see restatement. It is not trying to: the failure mode that produced 1.24 MB is
**regrowth**, one unremarkable paragraph at a time, and a ratchet makes each one a deliberate,
reviewable edit. Detecting duplication was considered and rejected for this change — these files
legitimately quote one another's requirement titles and citation paths, so a detector needs an
allowlist, and the allowlist becomes the thing that rots.

Generated artifacts in other repositories get **no** guard. They land at a new path in a different
repository every change; there is nothing stable to ratchet. For them D2's clause stands alone,
applied by the writing-plans and brainstorming stages that produce them.

### D5 — The proof is a normative-sentence inventory

A new script extracts every sentence containing SHALL, SHALL NOT, MUST or MUST NOT from the owned
corpus, normalises whitespace, and prints them sorted. Today's corpus carries 1340 such occurrences.

The trim SHALL leave that set **byte-identical**. It is captured before the first edit, again after
the last, and the two are diffed.

This exists because none of the guards already in `## lint` would notice a deleted requirement.
`check-references.sh` and `check-installed-citations.sh` catch a citation that stops resolving;
`check-markdown-integrity.py` catches structural damage; the 30 harnesses test script behaviour. A
`SHALL` sentence can be deleted with all of them green.

### D6 — No reduction target

A percentage target pressures an implementer to keep cutting once the genuine restatement runs out,
and the next thing to cut is a normative sentence. The ratchet locks in whatever was honestly cut,
which is the durable half.

## Execution

The trim runs **file-group by file-group**, largest-first, so the runtime-loaded files are edited
while attention is freshest. Each group is one task carrying its own inventory check and its own
full lint run — not one check at the end, where a regression cannot be attributed to a file.

Order:

1. `skills/myflow-do/`, `skills/myflow-contracts/` — the largest runtime-loaded text
2. the remaining `skills/`
3. `rules/`, `commands*/`, `.myflow/`
4. `openspec/specs/`
5. root `.md`

`rules/be-brief.mdc` is edited **first**, before any trimming, so every later group is cut under a
standard that is already written down.

`setup.sh global` is re-run at the end: the reworded rule renders into the managed block of
`~/.claude/CLAUDE.md` and `~/.codex/AGENTS.md`, and `scripts/check-installed-rules.sh` — landing
separately on `fix/stale-install-guard` — would otherwise report the install stale.

## Risks

- **A normative sentence lost by deletion.** Mitigated by D5's inventory, which is the only check
  that would see it.
- **A cut that creates a gap no file covers.** Several contracts are canonical for exactly one thing
  and deliberately restate nothing; deleting a passage as "restatement" when it is the canonical
  statement leaves nothing behind it. Mitigated by requiring a cut to name the file that still
  carries the statement.
- **Budget rows encoding a bad trim.** A ratchet locks in whatever it is given. The row is written
  after the group's inventory check passes, never before.

## Out of scope

- Rewriting for density
- New core/rationale partitions
- Other projects' existing documentation
- Any reduction target
