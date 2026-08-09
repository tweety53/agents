# myflow-do — rationale

This file is the reasoning behind `skills/myflow-do/SKILL.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

## State gate

## Superpowers Basic Workflow

## 1. Load context and validate the plan

## 2. Isolate the workspace (first run only)

Why the workspace id is computed fresh in this step, rather than carried over from an earlier one:

**Then compute this worktree's workspace id from the change name.** The derivation is stated once
under **The workspace id** (`skills/myflow-contracts/workspace-isolation.md`), which is canonical
for it — do not restate it here, and do not re-derive it by hand. Compute it once per run, on a fix
run exactly as on the first: the derivation is deterministic, so a later run reproduces the same id
rather than reading one back, which is why nothing about it is written to the state file. Two later
steps consume that one value — section 6 resolves the run instructions' URLs from it, and section 7
resolves the project's declared isolation rows against it — so an id derived twice in one run is two
chances to disagree.

## 3. Documenting a fix, before implementing it

Appending is recommended because most fixes are corrections within the change's existing scope, and
a sub-change per fix round buys a directory tree the operator has to read back. A nested sub-change
is never archived alone — it goes with its parent.

## 4. Execute (SDD + TDD)

## 5. The review panel

Free prose is not a record of a finding's state: a state that cannot be counted cannot be enforced.

Why panel dispatches never inherit a model from the parent session:

There is no
parent-model inheritance and no economy tier — the panel's cost must not depend on which model the
operator happens to be running, and a recorded value is a deliberate decision for one change rather
than an inheritance path.

Why a reviewer given a repo-relative path can end up with no principle list at all:

The subagent's working directory is
the project worktree, which has no `skills/` tree, so a repo-relative path opens nothing and the
reviewer runs with no principle list.

`scripts/check-unfinished-work.sh` reads **only the marker block**. It never parses the table: not
its header, not its column order, not its cell boundaries, not where it starts or stops. So an
unescaped `|` inside a cell is just text, a reordered header changes nothing, and a row that lost a
boundary pipe still counts. That is the point of the split — the previous shape asked a hand-rolled
table parser to recover one fact from a grammar defined in prose, and it failed **open** six
distinct ways across three review passes before it was replaced.

- Each `F<n>` names **one** finding. A reused identifier is reported on each side separately, so two
  distinct findings labelled `F1` in both the table and the marker block cannot cancel out — that
  shape hid an open Critical, with the word `open` never appearing in a marker at all.
- The marker lines sit on **consecutive lines**, one unbroken block. This is what stops a marker
  quoted elsewhere — inside a fenced example, say — standing in for a marker that was never written,
  which is the one route that still under-counted when the redesign was attacked.

**The table carries no status column, on purpose.** A finding's state is written once, on its
marker line. A status cell beside the marker is a second surface that can silently disagree with the
line that governs: the machine's direction is protected — a marker reading `open` blocks whatever a
cell says — but nothing protects a reader who sees `fixed` in the table and believes it. State the
fact once. To read a finding's state, look up its `F<n>` in the marker block.

### Code review (low)

Why the light preset's third slot is the harness's `code-review` skill rather than a narrowed
Superpowers reviewer: a narrowed Superpowers reviewer is the same reviewer slot 0 already runs, so
the panel would carry two readings from one reviewer rather than a genuinely different check. The
harness's `code-review` skill is a distinct lens the panel does not otherwise get, which is what a
lighter roster is trading Bugbot's defect hunt for.

Why an unavailable skill substitutes rather than drops: dropping the slot would make a missing
harness feature a silent way to weaken review — the panel's required-slot count would depend on
what the operator's harness happens to ship, rather than on the roster the operator chose. Falling
back to Bugbot instead of a substitute reviewer would silently convert the operator's `light` choice
into `standard`, which is not what the operator recorded.

### Bugbot's mutation-testing brief

Why the brief rides on the dispatch prompt rather than on Bugbot's own definition: Bugbot carries
its own agent definition, which the dispatcher does not edit, so the prompt is the only lever this
command has over what Bugbot is told to do.

Why a surviving mutant is an ordinary finding rather than an advisory note: an advisory class would
be a second kind of finding that nothing enforces, sitting beside a bar — zero open findings at any
severity — that enforces every other one. A finding class the guard does not check is a finding
class an operator can silently ignore.

### Optional slot selection

**Borderline → ask**, with **include** as the default. A reviewer too many costs tokens; one too
few costs a defect.

### Panel re-runs

## 6. Resolve the run instructions

In the same run, resolve the run instructions for the handoff. This is why reviewing and testing
are one gate: the diff and the way to run the change are produced together and can never drift
apart.

Nothing reads the run instructions back — no later command, no guard, no fix round consults a
recorded copy of them — so writing them to a file would be a record with no consumer. Printing them
into the handoff instead means they are only ever as stale as the run that just produced them, and
there is no second copy to fall out of sync with the worktree it describes.

A worktree's applications bind their own ports, so a URL taken from the project's declared base
instead of this worktree's resolved one reaches whichever workspace holds the default port — a
different change's application, answering plausibly and about the wrong work.

## 7. Verify, stage, and hand off

**This is the only place a project's declaration is validated, and that is why it happens here.**
The rules are mechanical, so re-deriving them by reading rows is the failure the guard exists to
remove; and the guard ships in the agents repository while the projects it judges do not, so a lint
list reaches one repository and this command reaches all of them. `/myflow-finish` does not repeat
it: run 2 reads the `survivors` row alone, and every input it cannot resolve is already a reported
skip under **Run 2 — the branch is merged** (`skills/myflow-contracts/finish-contract.md`). Adding a
blocking validation there would strand an already-merged change over text nothing in that session
can correct — the trade
**Creation and cleanup** (`skills/myflow-contracts/workspace-isolation.md`) rejects when it weighs
a change stranded short of its terminal state against stale storage.

Two distinct triggers land in the same outcome, and here is why that is deliberate rather than an
oversight:

The two
reasons are one case, deliberately: both end with nothing having run the guard, and a session that
recognised only "the repository does not carry it" would answer a failed resolution by doing
nothing at all and saying nothing about it.

- **Every declared row, not a subset.** A value this run resolved and did not export is a value
  nothing reads, so the workspace is isolated on paper while the applications and their checks still
  reach the project's shared resource — the same silent wrong answer as having derived nothing.

Creating a workspace database and bucket for a run that
only ever linted would leave behind resources nobody asked for and only `/myflow-finish` run 2
removes.

Why the two planning paths are left unstaged rather than filtered out of a display:

The exclusion is what keeps them out of the diff, rather
> than a filter applied when the diff is displayed: a filtered display leaves them in the staging
> area, where the IDE's staged-changes pane and `git status` show them again. The list is fixed —
> the pipeline chooses these paths itself, so no project can differ. `/myflow-finish` stages and
> commits them separately, so nothing is lost by leaving them unstaged here.

**The `reset` is what enforces the rule; without it the `add` only assumes it** — the reason is
stated once under **Git boundaries** (`skills/myflow-contracts/pipeline.md`) and is not re-derived
here. What is specific to this command is *whose* staging it retracts (an implementer subagent's own
`git add`, or a worktree resumed with a dirty index) and why `git reset -- <paths>` is the tool:
it touches the index only, restores a tracked path to its `HEAD` entry instead of staging a deletion
the way `git rm --cached` would, and succeeds when a path is absent — which `docs/superpowers/` is
on every run that has not preserved records yet, and where `git restore --staged` would refuse the
whole command and unstage nothing.

The block below is **not** a second definition of the handoff. It is this command's rendering of the
`IN_PROGRESS`-after-`/myflow-do` template, which is defined once under
**The block each state renders** (`skills/myflow-contracts/handoff-blocks.md`) and is canonical for the
labels, the field set and their order. What this block adds is the enumeration of the literal
alternatives `/myflow-do` writes. **Change the template first and bring this block with it** — a
field added here and not there is drift the moment `/myflow-status <name>` regenerates the same
state.

## Guardrails
