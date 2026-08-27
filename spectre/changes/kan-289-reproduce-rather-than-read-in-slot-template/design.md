# Design — kan-289

Canonical design for this change. `proposal.md` states what and why; this file states how, and
records every decision and open question.

## Part 1 — the dispatch template

### The two variants

One label, `**REPRODUCE, DON'T READ:**`, two texts.

**Reviewer variant** — carried verbatim into every review-panel slot dispatch and into the per-task
combined reviewer:

> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one check you make MUST exercise the real
> thing. A claim you did not run is worth less than one you did: a doc comment, a type signature
> and a passing test can each read plausibly and be false. Run it before you accept it, and run it
> before you reject it.

**Implementer variant** — carried verbatim into the implementer dispatch:

> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one test you write MUST exercise the real
> thing. A test backed by a fake or a hand-built value passes while the real integration is broken:
> the shape you construct by hand is not the shape the real producer emits. Build the value the way
> production builds it, or assert against the real boundary.

### The three sites

| File | Family | Variant |
|------|--------|---------|
| `skills/flow/review-panel.md` | every slot's dispatch prompt — all five slots | reviewer |
| `skills/flow/implement.md` | the per-task combined reviewer | reviewer |
| `skills/flow/implement.md` | the implementer dispatch | implementer |

In `review-panel.md` the reviewer variant joins the cluster of things that already travel with every
slot dispatch — the per-finding reproducer requirement and the CONTEXT BUNDLE paragraph. In
`implement.md` the implementer variant joins the `REQUIRED SUB-SKILL` / `REQUIRED READING` /
`CONTEXT BUNDLE` / `PLAN PROVENANCE` run, and the reviewer variant is stated with the per-task review
paragraph.

All five panel slots carry it, Bugbot and Security included — the same reach the existing per-finding
reproducer requirement has ("Carry this requirement on every slot's dispatch prompt").

### The guard

`scripts/check-reproduce-not-read.sh`, harness `scripts/test-check-reproduce-not-read.sh`, both
registered in the project's `## lint` and `## test` lists.

Per required site it asserts the label is present and that variant's load-bearing phrases are
present:

- `skills/flow/review-panel.md` — at least one block, carrying the reviewer phrases.
- `skills/flow/implement.md` — at least two blocks, one carrying the reviewer phrases and one the
  implementer phrases, because both families live in that one file.

Exit codes follow the house convention: `0` clean, `1` a required site missing its block or one of
its phrases, `2` cannot answer at all (a scoped file missing or unreadable). A renamed file is a `2`,
never a silent green.

### Two guards this part feeds

The contract-budget guard carries a budget row for each edited skill file. Both grow, so both rows
are raised deliberately — the correct response to a genuine addition, per that guard's own rule.

The normative-inventory guard will report the two new `MUST` sentences. This is an addition, not a
deletion: the inventory grows, and nothing is restored.

## Part 2 — the rename

### In scope

`skills/`, `scripts/`, `rules/`, `commands/`, `commands-claude/`, `stats/`, `spectre/specs/`,
`.myflow/`, and root `README.md`, `CLAUDE.md`, `AGENTS.md`, `setup.sh`.

Covering, layer by layer:

| Layer | From | To |
|-------|------|-----|
| CLI binary | `myflow`, `stats/cmd/myflow` | `flow`, `stats/cmd/flow` |
| daemon | `myflowd`, `stats/cmd/myflowd` | `flowd`, `stats/cmd/flowd` |
| container + volume | `myflow-postgres`, `myflow-pgdata` | `flow-postgres`, `flow-pgdata` |
| database + role | `myflow`, `myflow_uitest`, `myflow_<id>` | `flow`, `flow_uitest`, `flow_<id>` |
| environment | `MYFLOWD_*`, `MYFLOW_*` | `FLOWD_*`, `FLOW_*` |
| project config | `.myflow/project.md` | `.flow/project.md` |
| contract skill | `skills/myflow-contracts/` | `skills/flow-contracts/` |
| rule stub | `rules/myflow-manual-review.mdc` | `rules/flow-manual-review.mdc` |
| contract names | `myflow-task-commit-fields`, `myflow-fix` … | `flow-task-commit-fields`, `flow-fix` … |
| launchd agent | `com.tweety53.myflowd.plist` | `com.tweety53.flowd.plist` |
| state root | `~/Agents/myflow/state` | `~/Agents/flow/state` |
| Go/TS identifiers | package, type and symbol names carrying `myflow` | `flow` |

### Out of scope

- **`openspec/`** — frozen at the 2026-08-25 cutover and never written to again. Renaming it
  contradicts the freeze.
- **`docs/superpowers/`** and **`spectre/changes/archive/`** — historical records of runs that
  actually invoked `/myflow-do` under a binary called `myflow`. Rewriting them falsifies the record.
- **The retired command-name constants** in `stats/internal/stages/names.go` — `/myflow-start`,
  `/myflow-do`, `/myflow-finish`, `/myflow-fast`.
- **The `mf-` session-token prefix.**
- **Existing Jira issue labels.**

The excluded trees account for 486 files carrying the string, against 231 files in the live corpus.
<!-- measured: git ls-files -z 'openspec/*' 'docs/superpowers/*' 'spectre/changes/archive/*' | xargs -0 grep -ril myflow | wc -l; and grep -ril myflow skills scripts rules commands commands-claude stats .myflow README.md CLAUDE.md AGENTS.md setup.sh | wc -l @ 07acd2e (before this change) -->

### The four operator-run steps

Renaming the container, renaming the database, reinstalling the launchd agent, and installing the
renamed binary each require the dev workspace's daemon and storage to stop. No agent action stops
them. The handoff carries the four commands, in order — the database rename in particular requires
zero live connections, so the daemon must be down before it runs, and bringing it back up does not
repair a run that already fell through to the on-disk journal.

### The CLI and the daemon speak a renamed protocol and must move together

**The rename changes a wire value, not only names.** The daemon stamps every response with a trust
header so a look-alike server squatting the port cannot be mistaken for a genuine refusal, and that
header is renamed with everything else: `Myflow-Daemon: myflowd/1` becomes `Flow-Daemon: flowd/1`, in
`stats/internal/api/server.go` and `stats/internal/client/client.go` together.

Within the branch the two agree. Across the cutover they do not: **a renamed CLI cannot trust the
running daemon, and the old CLI cannot trust a renamed one.** Measured directly — a branch-built CLI
against the live daemon reports `response missing Flow-Daemon header -- not trusted as a store
answer`, and the live daemon really does send `Myflow-Daemon: myflowd/1`.
<!-- measured: curl -sI against the running daemon, and a branch-built CLI run against it @ branch spectre/kan-289-reproduce-rather-than-read-in-slot-template -->

**This degrades safely, and that is by design rather than luck.** An untrusted response is a store
failure, and every CLI path that touches the store falls back on any failure and exits 0: the write
goes to the on-disk journal, and the daemon replays it once it is reachable again. So a `/flow` run
during the window records to the journal instead of the store rather than failing. What it does
**not** do is read back what it just wrote, so a run spanning the window sees a stale record.

**Consequence for the runbook: steps 2 and 4 are one operation, not two.** Installing the renamed CLI
while the old daemon still runs opens the window; restarting the daemon from the same tree closes it.
Do them back to back, with no `/flow` run in between — the same instruction the container rename
already carries, for a second and independent reason.

### The cutover is ordered, and the order matters

The branch is internally consistent at every commit, but the machine is not consistent with it until
the operator finishes the sequence below. **These steps run in this order; none may be reordered or
merged.**

1. **The branch merges.** Nothing on the machine has changed yet: the installed skills are symlinks
   into the main checkout, so they flip to the renamed text the moment the merge lands.
2. **Build and install the renamed CLI** (`bin/flow`) into `~/.local/bin`. The skills now say `flow`,
   so this must happen before the next `/flow` run, not after it. The old `~/.local/bin/myflow`
   binary stays in place until step 5 — a `/flow` run mid-sequence must be able to fall back.
3. **Rename `.myflow/` to `.flow/` in every consuming project** — this repository,
   `~/Projects/gymie`, `~/Projects/spectre-e2e`. Until this is done, the renamed resolver reports its
   actionable error in those projects.
4. **Stop the daemon, rename the container and database, restart.** This is the only step that stops
   the dev workspace's service and storage, and it is the operator's to run.

   **This step is what makes `scripts/workspace.sh` correct again, and until it runs, worktree
   creation is broken.** That script addresses the container by name (`CONTAINER`) and connects as a
   superuser role (`SUPERUSER`) in order to create and drop a change's own `flow_<id>` database. The
   branch ends with both naming `flow-postgres` and `flow`, because a branch that ends naming the old
   container would need a second change to finish the rename. The consequence is a real window: from
   the merge until this step completes, `workspace.sh` names a container that does not exist yet, so
   **no new `/flow` run can isolate a workspace.** Discovered when task 7 correctly declined to
   rename those two values against a container it was forbidden to touch.

   **So do not start a `/flow` run between step 1 and step 4.** That is the whole mitigation, and it
   is a sequencing instruction rather than a code change: making the container name configurable, or
   dual-reading both names, would be configurability nobody asked for to paper over a window measured
   in minutes.

   `ALTER DATABASE myflow
   RENAME TO flow` requires zero live connections, so the daemon must be down first. Bringing it back
   up does not repair a run that already fell through to the on-disk journal, so run this when no
   `/flow` run is in flight.
5. **Reinstall the launchd agent** under its new label and remove the old one, then remove the
   superseded `myflow` binary.

**During implementation the worktree and the main checkout disagree, and that is expected.** The work
happens in an apply worktree; the installed skills resolve through the main checkout, so this run and
any concurrent one keep reading the pre-rename text. Guards run in the worktree read the worktree's
own renamed `.flow/project.md`. Nothing is broken by the disagreement — it closes at step 1.

### `.myflow/` reaches other projects

`~/Projects/gymie/.myflow` and `~/Projects/spectre-e2e/.myflow` exist alongside this repository's.
Every reader performs a hard cutover: finding `.myflow/` and no `.flow/`, it reports an actionable
error naming the project root and the exact `git mv` to run, and exits non-zero. It never falls back
to the old path and never reads both.

**There are three readers, all of them shell, and there is deliberately no Go one.** They are
`scripts/check-workspace-isolation.sh`, `scripts/check-cleanup-complete.sh` and `setup.sh`'s
`install_project_standards`. Task 11 was dispatched to "update the Go configuration resolver" and
measurement found none existed — no Go code read `<project>/.flow/project.md` at all, at the merge
base or since. A Go resolver was written anyway, to the letter of the task, and the tasks 10-11
review then found it had **no production caller** and had already diverged from the three shell
readers on an edge case (an empty `.flow/` directory beside a populated `.myflow/`, which the Go side
treated as "nothing declared" — the exact silent outcome this decision forbids).

It was deleted rather than repaired. Keeping it would mean maintaining an unused API that must stay
byte-consistent with three shell guards forever, and it had already failed to; the divergence was not
a bug to fix but the predictable cost of a fourth implementation nobody calls. **No abstraction until
a second caller exists** — here there was not even a first.

## Decisions

### How far the dispatch paragraph reaches

**ID:** dispatch-families-three
**Status:** active
**Chosen:** all three dispatch-prompt families — panel slots, per-task reviewer, implementer —
because the ticket's own evidence is a defect class the implementer creates, the per-task reviewer
misses, and the panel catches last.
**Considered:** panel slots alone (ticket-faithful and smallest, but catches the defect one review
too late and says nothing to the party that introduces it); panel slots plus per-task reviewer
(catches it at the task commit, still silent toward the implementer).

### Whether the template gets a guard

**ID:** guard-the-template
**Status:** active
**Chosen:** a presence guard plus its harness, scoped to this new block. The ticket's root cause is
that the instruction lived in no template and depended on the dispatcher remembering; a guard is
what stops a later prose edit from reproducing that failure one level up.
**Considered:** prose only, matching the `kan-201` precedent that added CONTEXT BUNDLE to three
families with no guard — rejected because nothing then asserts the paragraph is there, and every
task's `**Tests:**` field would have to lean on guards that do not check it; a wider guard covering
every "carried verbatim into a dispatch" block in the corpus — rejected as scope not asked for, and
addable later against a guard that will already exist.

### One text or two

**ID:** two-variants
**Status:** active
**Chosen:** two variants under one shared label. A reviewer judges someone else's claim; an
implementer writes the test that will later be judged. The obligation differs — "at least one check
you make" against "at least one test you write" — and so does the failure each is warned about. The
shared label keeps the guard simple: one label pattern, variants distinguished by their own phrases.
**Considered:** one byte-identical text at all three sites, as CONTEXT BUNDLE does — rejected
because it would have to be vague enough to fit both jobs, and vagueness is what the ticket is
repairing; one shared text plus an implementer-only extra sentence — rejected as the same text with
a seam, carrying both the vagueness and the divergence.

### What the guard asserts

**ID:** label-plus-phrases
**Status:** active
**Chosen:** the label plus a small set of load-bearing phrases per variant. Catches a deletion and
catches a block edited down to nothing, without the guard holding a verbatim copy of the text.
**Considered:** label presence alone — rejected because a block gutted to its label still passes,
and the label is not the instruction; the full literal text held in the guard — rejected because it
makes the guard a further copy of prompt text, so every legitimate rewording becomes a two-file edit
and the guard drifts from intent the first time one is skipped.

Panel finding F33 on `kan-201` is the precedent that verbatim copies of prompt text are correct where
they are: the operator withdrew "CONTEXT BUNDLE is repeated three times" on the grounds that a
pointer cannot resolve inside a subagent prompt. That licenses the copies at the three dispatch
sites; it does not license another copy inside a guard that nothing dispatches.

### Scope of the install check

**ID:** repo-source-only
**Status:** active
**Chosen:** the guard checks repository source only. `~/.claude/skills/flow` is a symlink to
`skills/flow/` in this checkout, so the installed copy *is* the source and there is no second copy to
drift.
**Considered:** a sandboxed installer sweep in the style of the installed-citations guard — rejected:
that guard exists because citations resolve differently once installed, which is not true of a file
the installer symlinks.

### No capability spec

**ID:** no-spec
**Status:** active
**Chosen:** no capability spec is written. `spectre/specs/` is empty and the three most recently
archived changes wrote none, so the practice in the spectre tree is no spec; `openspec/specs/` is
frozen and never written to again.
**Considered:** authoring the first `spectre/specs/` entry — rejected as a migration decision larger
than this change, which would set the tree's convention as a side effect of a prose addition.

### How far the rename reaches

**ID:** rename-everything-identifiers-included
**Status:** active
**Chosen:** the full rename, identifiers included — binary, daemon, container, database, environment
variables, `.myflow/`, contract skill directory, Go and TypeScript identifiers. The operator asked
for it explicitly after being shown the three narrower boundaries and the dev-stack risk.
**Considered:** prose and contract names only (no running service touched, but leaves the corpus
reading half-renamed); prose, contracts and `.myflow/` (still leaves the daemon, database and
environment disagreeing with the product name); prose only (smallest, and the least coherent
outcome).

### The frozen and historical trees are excluded

**ID:** rename-excludes-frozen-and-historical
**Status:** active
**Chosen:** `openspec/`, `docs/superpowers/` and `spectre/changes/archive/` are not renamed. The
first is frozen by the project's own configuration; the other two are records of runs that really did
invoke `/myflow-do` under a binary really called `myflow`.
**Considered:** renaming them for corpus-wide consistency — rejected because a record edited to say
something that did not happen is worse than an inconsistent one, and because the freeze is a stated
rule this change has no mandate to break.

### Retired command names are data, not identifiers

**ID:** retired-command-constants-are-data
**Status:** active
**Chosen:** every `/myflow-*` **string literal** in `stats/` stays exactly as it is. They are values
already written into `stage_runs` rows in the live database; renaming them would make the code
disagree with stored history and silently break every lookup of an existing row.
**Considered:** renaming them with a data migration over `stage_runs` — rejected: it rewrites history
to no benefit, and the rows record which command actually ran.

**Corrected during implementation, after task 5.** This decision originally said the names were
constants in `stats/internal/stages/names.go`. That was wrong, and measuring it is what showed so:
`names.go` declares exactly one `Command`, `/flow`, and carries no retired name at all — the
`kan-326` rework had already reduced it. The retired names live instead as bare string literals
spread across 39 files.
<!-- measured: grep -rlo '/myflow-[a-z]*' stats/ --include='*.go' | wc -l, and grep -rn 'Command = "/myflow' stats/ --include='*.go' @ branch spectre/kan-289-reproduce-rather-than-read-in-slot-template -->

**Not all of them are test fixtures**, which is the part that matters for the sweep: production code
carries them too, in `stats/internal/api/stats.go`, `stats/internal/fallback/statefile.go`,
`stats/internal/client/client.go`, `stats/internal/records/render.go`,
`stats/internal/store/changes.go` and `stats/cmd/flow/state.go`.

**The exclusion is therefore per-occurrence, not per-file, and task 9 classifies each one** rather
than skipping a file wholesale. Three kinds occur: a literal that is **stored data** (a fixture or a
query asserting against a `stage_runs` row that really carries that command) stays; a literal that is
**live behaviour** (a filter or a route selecting on a command name) stays, because the rows it
selects still carry the old name; a **doc comment** describing what a retired command used to do
stays, because rewriting it makes a true sentence false. Nothing in this set is renamed — but the
reason differs per occurrence, and a sweep that cannot tell them apart is a sweep that will rename
the wrong one.

### The session-token prefix is unchanged

**ID:** session-token-prefix-unchanged
**Status:** active
**Chosen:** `mf-` stays. It is an opaque correlator prefix, validated by both the CLI and the daemon,
and it prefixes every token already in the store and in every unreplayed journal entry.
**Considered:** `fl-` for new tokens — rejected: it splits the correlator space for a two-character
cosmetic gain, and the prefix is not the word `myflow`.

### `.myflow/` is a hard cutover with an actionable error

**ID:** dotmyflow-hard-cutover
**Status:** active
**Chosen:** the resolver reports an actionable error naming the directory to rename when it finds
`.myflow/` and no `.flow/`, and exits.
**Considered:** dual-reading with `.flow/` preferred — rejected as configurability nobody asked for,
and as the shape that lets two consuming projects sit in different states indefinitely without anyone
noticing; a silent hard cutover — rejected because two other projects on this machine carry
`.myflow/`, and a bare "not found" would send the operator debugging the wrong layer.

### The stateful steps are the operator's

**ID:** operator-runs-the-stateful-steps
**Status:** active
**Chosen:** renaming the container, renaming the database, reinstalling the launchd agent and
installing the renamed binary are written out for the operator and never run by the pipeline. Each
requires the dev workspace's daemon and storage to stop.
**Considered:** running them as part of the change — rejected outright: no agent action stops that
service or that storage.

### The vocabulary guard carries the rename

**ID:** vocabulary-guard-carries-the-rename
**Status:** active
**Chosen:** `myflow` and its identifier shapes are added to the vocabulary guard's retired-literal
list, scoped to the live corpus. This is the mechanism that keeps the rename from regrowing, and it
mirrors part 1's own guard decision.
**Considered:** relying on review alone — rejected for the reason that guard's own header already
states: a rename is not proven complete by a green run, but an unguarded one regrows silently.

### Jira labels change in the documentation only

**ID:** jira-labels-docs-only
**Status:** active
**Chosen:** the contract text switches to `flow-improvement`, `flow-fix`, `flow-cost`,
`flow-automation` and `flow`, so newly filed issues carry the new labels. Existing issues keep the
labels they have.
**Considered:** leaving the documented taxonomy alone (consistent with the board, inconsistent with
everything else this change renames); relabelling existing issues through the API (an outward-facing
bulk write over a board this repository does not own).

**Consequence, recorded rather than hidden:** the board carries both taxonomies indefinitely, and any
label-based search has to match either form.

### Two concerns, one branch, separate commits

**ID:** two-concerns-separate-commits
**Status:** active
**Chosen:** the dispatch-template work and the rename land as separate commits with separate scopes
on the one branch, at the operator's explicit request to carry both in this task.
**Considered:** a second change for the rename — rejected because the operator asked for it here;
recorded so the size of the resulting diff is not mistaken for scope creep.

## Open questions

### Does `flow` collide with another binary on a machine that installs this?

**ID:** flow-binary-name-collision
**Status:** open
**Why it is open:** neither `flow` nor `flowd` is on this machine's PATH today, so nothing is broken
here — but `flow` is also the name of a widely installed JavaScript type checker, and this pipeline
installs into `~/.local/bin`, which precedes `/opt/homebrew/bin` on this PATH. A machine that later
installs that tool would get whichever the PATH order favours, silently.
**What it affects:** whether the installed CLI keeps the bare name `flow` or takes a distinguishing
one. Deciding it later costs another rename of the binary name alone, not of the corpus.
