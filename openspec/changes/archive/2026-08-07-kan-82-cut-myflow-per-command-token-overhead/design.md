# Design — cut myflow per-command token overhead

Source: `docs/superpowers/specs/2026-08-07-kan-82-cut-myflow-per-command-token-overhead-design.md`,
approved at the design gate on 2026-08-07. Adapted here, not duplicated: where the two differ, the
delta specs under `specs/` are normative and this file is the how.

## Context

| File | Lines | Bytes | Loaded by |
|------|-------|-------|-----------|
| `skills/myflow-contracts/pipeline.md` | 1540 | 103326 | every `/myflow-*` command |
| `skills/myflow-contracts/jira-integration.md` | 732 | 52783 | `/myflow-start`, `/myflow-do`, `/myflow-finish` |
| `skills/myflow-contracts/project-configuration.md` | 542 | 45236 | `/myflow-do`, `/myflow-finish` |
| `skills/myflow-contracts/state-file.md` | 236 | 15951 | any state read or write |

<!-- measured:wc -l -c skills/myflow-contracts/*.md @ adedf66 -->

Structural content — table rows, fenced blocks, headings, list items, block quotes — against
everything else:

| File | structural | prose | prose % |
|------|-----------|-------|---------|
| `pipeline.md` | 21702 | 81624 | 78% |
| `jira-integration.md` | 4985 | 47798 | 90% |

<!-- measured:awk classifier over each file @ adedf66 -->

`### Follow-up issues` in `jira-integration.md` is 457 lines / 35467 bytes.

<!-- measured:awk '/^### Follow-up issues/{f=1} f{n++; b+=length($0)+1}' @ adedf66 -->

The corpus carries 132 backticked references to `pipeline.md` and 61 to `jira-integration.md`,
inside 258 bold-token cross-references overall.

<!-- measured:grep -roh over *.md and *.mdc, archive excluded @ adedf66 -->

## Architecture

Five files where there were two.

```text verified:authored in-tree for this change
skills/myflow-contracts/
  pipeline.md                    core       every command
  pipeline-rationale.md          appendix   editors only
  jira-integration.md            core       start, do, finish
  jira-integration-rationale.md  appendix   editors only
  jira-followups.md              whole      finish run 1 only
```

The partition test, the verbatim rule, the heading-mirroring rule, the runtime rule and the budget
guard's contract are all stated once in
`specs/myflow-contract-economy/spec.md` and are not restated here.

### Order of work

The order matters, because each step is verifiable only once the one before it has landed:

1. **Move `Follow-up issues` out** of `jira-integration.md` into `jira-followups.md`, whole. This
   is a pure file move of a contiguous 457-line block plus the citations that point into it. It is
   independently valuable and independently verifiable, and it removes 67% of that file before any
   partition judgment is needed.
2. **Partition `jira-integration.md`** — the smaller and simpler of the two partitions, so the
   partition mechanics and the line-multiset check are proved on it first.
3. **Partition `pipeline.md`** — the large one, with the mechanics already settled.
4. **Repoint citations** after each of the three, not once at the end: `check-references.sh` is the
   proof, and running it against one moved file at a time is what localises a failure.
5. **Write the budget guard** last, because its numbers are the sizes the first three steps
   produced.

### Why the appendix is a sibling file rather than a marked region

A marked region inside one file — core above, rationale below a runtime-stop marker — needs no
repointing and no new files, and saves nothing: a `Read` of the file still pulls the whole thing
into the context window. The cost this change exists to remove is paid at load, so the split has to
be a file boundary.

### Why `jira-followups.md` is not itself split

It is loaded by exactly one command. Splitting it would save `/myflow-finish` run 1 some of its
load while costing that same command a second file to open, and would leave the reasoning for the
most adversarial rules in this corpus — the JQL scoping argument, the title-sanitisation argument —
one file away from the rules they justify, for no benefit to any other command.

## Verification

| What must hold | How it is proved |
|----------------|------------------|
| No line was lost or altered | Multiset of non-blank lines in core+appendix equals that of `git show <base>:<path>`, excluding lines differing only by a repointed citation path |
| No citation went stale | `scripts/check-references.sh` exits 0 |
| The cores are and stay small | `scripts/check-contract-budget.sh` exits 0 |
| The new files ship | `scripts/test-setup.sh`, plus the added scenario over `install_skills` |
| Nothing else regressed | Every command in `.myflow/project.md`'s `## lint` and `## test` |

The line-multiset check is the load-bearing one, and it is a **line** multiset rather than a
sentence multiset deliberately: a line is what `sort` and `comm` compare without a parser, and the
partition moves whole lines by construction.

## Decisions

### Scope is the two contract splits; the system-prompt cut is not in it

**ID:** scope-contract-splits-only
**Status:** active
**Chosen:** Split `pipeline.md` and `jira-integration.md` only — pruning the harness system prompt
is an operator action against an installed environment, not a change to this repository, and needs
its own issue.
**Considered:** All three cuts, with something in-repo standing in for the third (trimming this
repo's own skill `description:` frontmatter) — rejected because the frontmatter is a small fraction
of the listing and the real lever is uninstalling plugins, which no repo change reaches. `pipeline.md`
alone — rejected because the Follow-up issues move is a pure file move with a large payoff and no
new mechanics. Every contract file at once, including `project-configuration.md` — rejected as a
much larger partition surface for a file two commands load rather than five.

### The rationale goes to a sibling appendix, one per contract

**ID:** sibling-appendix-per-contract
**Status:** active
**Chosen:** `<name>-rationale.md` beside each core — the core cites it by section name, so
`check-references.sh` verifies the link, and an editor of one contract loads only that contract's
reasoning.
**Considered:** One shared `rationale.md` for all contracts — rejected because an editor of one
contract would load the reasoning of all of them, which is the same audience mismatch one level up.
Same file with a runtime-stop marker — rejected because a `Read` still pulls the whole file, so it
saves nothing that matters. Deleting the rationale and relying on git history — rejected because
several sections of this corpus exist precisely because an editor re-derived a decision wrongly, and
git history is not where an editor looks.

### The core is a strict partition of today's text, not a rewrite

**ID:** strict-verbatim-partition
**Status:** active
**Chosen:** Every sentence lands in exactly one file, verbatim; the only permitted edit is
repointing a citation whose target moved. The issue's "no two independently authored statements"
constraint then holds by construction, and the diff reviews as a move.
**Considered:** Partition then tighten the core for flow — rejected because the diff stops being a
pure move and every reworded line becomes something the panel must check by hand. Rewrite the core
as a fresh summary — rejected as exactly the two-statements-of-one-rule failure the issue forbids.

### A stale position word may be deleted, and a referenced passage stays with its referrer

**ID:** position-words-deletable
**Status:** active
**Chosen:** A move may delete an `above` or `below` the move made false — a deletion, never a
substitution — and a passage another passage depends on by reference is classified with that
passage rather than moved away from it.
**Considered:** Keeping the path-repointing-only rule and returning every referenced passage to the
core instead — rejected because it shrinks the appendix further on a saving that is already modest,
and because it cannot fix the ten stale `above`s inside `jira-followups.md`, whose referents stayed
in `jira-integration.md` by design. Permitting the word to be *reworded* into a filename — rejected
because a substitution is authored text, and the diff would stop being reviewable as a move.

**This reverses `strict-verbatim-partition`'s round-3 refinement, and the review panel is what
forced it.** Three slots independently found the same class of defect: a split moves passages across
a file boundary that did not exist before, so a position word that was true inside one file becomes
false across two. Roughly a dozen instances survived, and one — the term "the pre-check" in
`pipeline.md` — left a **core** reader with an undefined term whose definition sits in a file no run
loads. That is the failure mode a contract split exists to avoid, reached by obeying two rules that
were each individually sound.

A deleted word still appears as a `<`/`>` pair in the line-multiset check, so it is reviewed exactly
as a repointed citation is; the check's power is unchanged.

### The unit that moves is a whole paragraph, not a sentence

**ID:** paragraph-granularity-partition
**Status:** active
**Chosen:** Only whole blank-line-delimited paragraphs, bullets and tables move. A paragraph mixing
a rule with its justification stays in the core entire, and the lines that remain around any moved
passage are never re-wrapped.
**Considered:** Splitting at the sentence boundary, which is what this design originally instructed
— found **unexecutable** while implementing task 2.1. These files are hard-wrapped at roughly 100
characters, so a sentence boundary almost never lands on a line break; splitting mid-paragraph means
re-wrapping the lines that stay, and a re-wrapped line is indistinguishable from a lost one in the
line-multiset check. Switching to a sentence-multiset check with a wrapping normaliser would have
restored sentence granularity, and was rejected: it gives up the pure-move property that makes the
diff reviewable, and split sentences strand their anaphora — "That narrows the window" — in the
other file.

**What it costs, measured rather than estimated.** `jira-integration.md` is unusually rule-dense and
yielded only 2472 bytes to its appendix under this rule, a cut of 11% on that file. `pipeline.md`
carries 66336 of its 102580 bytes as whole pure-prose paragraphs — 64% — so the ceiling there is far
higher and the large win survives.

<!-- measured: python3 fence-aware paragraph classifier over pipeline.md @ branch openspec/kan-82-cut-myflow-per-command-token-overhead -->

The linked issue's target of roughly 8k tokens for `pipeline.md` is **not** reachable under this
rule. `budget-achieved-plus-25` is what makes that a smaller win rather than a failed change, and it
was chosen before this constraint was known.

### `jira-integration.md` splits on both axes, into three files

**ID:** jira-three-file-split
**Status:** active
**Chosen:** Core, rationale appendix, and `jira-followups.md` — the last carrying its rules and its
reasoning inline, since only `/myflow-finish` run 1 loads it.
**Considered:** Core/appendix only — rejected because Follow-up issues' normative half would stay
in the core and `/myflow-start` would keep paying for it. By-command only, leaving rationale in
place — rejected because it leaves the largest remaining file unpartitioned. A fourth file,
`jira-followups-rationale.md` — rejected because the only command that would benefit from it is the
only command that loads the file at all, and it would separate this corpus's most adversarial
reasoning from the rules it justifies.

### The appendix mirrors the core's heading tree

**ID:** appendix-mirrors-headings
**Status:** active
**Chosen:** Same `##`/`###` headings, same order, in both files; a wholly normative section leaves
its appendix heading present with no body.
**Considered:** Free-form headings grouped by topic in the appendix — rejected because a citation
must then name the right heading and the two structures drift. One flat list keyed by core heading —
rejected as long headings with no nesting, in files that nest three deep today.

### Budgets are measured in bytes and declared inside the guard

**ID:** budget-bytes-in-guard
**Status:** active
**Chosen:** A path-to-max-bytes table inside `scripts/check-contract-budget.sh`. Deterministic, no
tokenizer dependency, one place to read, and raising a budget is a visible diff in the guard.
**Considered:** Lines — rejected because a long-line file cheats it. A budget in each file's own
header — rejected because the file can then raise its own limit. A budget in `.myflow/project.md` —
rejected because that file configures a consuming project, and this is a rule about this
repository's own sources.

### Each budget is the achieved size plus 25%

**ID:** budget-achieved-plus-25
**Status:** active
**Chosen:** Split first, then set each budget from the resulting byte count plus 25%. The guard is
a ratchet against regrowth and can never fail the split itself.
**Considered:** A target declared up front that the split must hit — rejected because it creates
pressure to push normative text into an appendix to make a number, which is the one failure mode
that would make this change harmful rather than merely disappointing. Achieved size with no
headroom — rejected as forcing a guard edit for every ordinary sentence added.

### The no-appendix-at-runtime rule is stated, not enforced

**ID:** runtime-rule-unenforced
**Status:** active
**Chosen:** State in `SKILL.md`, in each core's header and in `rules/myflow-manual-review.mdc` that
a `/myflow-*` run never loads an appendix; enforce nothing.
**Considered:** A guard that fails when a skill file names a `-rationale.md` path in a load
instruction — rejected as machinery around a rule no observed failure has broken, in a repository
that already carries eight guards.

### The budget guard covers every contract file, not only the split ones

**ID:** budget-covers-all-contracts
**Status:** active
**Chosen:** Every `skills/myflow-contracts/*.md` gets a budget, split or not, and a file with no
entry is itself a failure — so a contract added later cannot silently escape the ratchet.
**Considered:** Only the files this change splits — rejected because `project-configuration.md`
(45236 bytes) is the next file to have this problem and would grow unwatched. Every contract file
*and* every `SKILL.md` in the repository, including `skills/myflow-do/SKILL.md` at 40529 bytes —
rejected as widening this change's surface; those files are ratchetable later by adding rows.

## Open questions

*(none — every question raised during planning was answered before the design gate)*
