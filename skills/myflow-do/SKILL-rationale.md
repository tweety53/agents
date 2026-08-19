# myflow-do — rationale

This file is the reasoning behind `skills/myflow-do/SKILL.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

**Load `skills/myflow-contracts/pipeline.md` first** — it is canonical for the states, the
command→state transition table, git boundaries, and the handoff output shape.

## State gate

## Superpowers Basic Workflow

## 1. Load context and validate the plan

## 2. Isolate the workspace (first run only)

Why the merge base and path go into working notes rather than the state file at this point: they are
what section 7's workspace-isolation guard needs, and this run knows them from here on regardless of
what the on-disk state file says until section 7's own write.

Why the resolved worktree set is non-empty by construction on every ordinary run, first or fix
alike: the worktree is known the moment it is created or resumed above, well before section 7's
empty-set stop could ever fire on the ordinary shape of a first run.

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

It names the preset in force for this run, per
**State file** (`skills/myflow-contracts/state-file.md`), which is canonical for the field's shape,
its values and the absent-key rule.

Why no preset moves the handoff bar: a preset able to lower it would be a way to skip review, not a
way to size the panel's reading.

Why the containment rule on `[STANDARDS_PATHS]` is non-negotiable: it stands between an
attacker-editable list in a tracked file (`.myflow/project.md`'s `## standards` entries) and an
arbitrary file read landing in a committed review record.

Free prose is not a record of a finding's state: a state that cannot be counted cannot be enforced.

Why a `withdrawn` marker's reason is checked for being there at all: a fix subagent rewriting
`open` to `withdrawn` would be closing its own gate.

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

- `<status>` is **exactly** `open`, `fixed` or `withdrawn`, compared byte for byte. `Open`,
  `WITHDRAWN`, `open (needs discussion)`, an empty value and a value carrying an invisible character
  are none of the three, and none of them reads as closed.
- `withdrawn` **carries its reason on the same line** — the reason is part of the state, not a note
  about it.
- Each `F<n>` names **one** finding. A reused identifier is reported on each side separately, so two
  distinct findings labelled `F1` in both the table and the marker block cannot cancel out — that
  shape hid an open Critical, with the word `open` never appearing in a marker at all.
- The marker lines sit on **consecutive lines**, one unbroken block. This is what stops a marker
  quoted elsewhere — inside a fenced example, say — standing in for a marker that was never written,
  which is the one route that still under-counted when the redesign was attacked.
- `findings-total: <n>` appears **exactly once** and equals the number of marker lines. A record
  with no total line is outstanding however clean it reads: zero findings is not something to infer
  from silence. A panel that raised nothing says `findings-total: 0` and carries no markers.

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

Why the not-re-run scoping reaches conditional slots alone: a conditional slot's region is exactly
its trigger's subject, while a required slot has no bounded region to check staleness against.

Why checking the path alone was not enough: an argument such as `../../../etc/passwd` was free to
escape the worktree while the path token itself stayed clean.

Why an embedded newline needs no separate ban: the anchored `finding-reproducer:` marker line this
guard reads is one line by construction, so it could never carry one.

Why the 20-second bound is not `check-cleanup-complete.sh`'s 60-second `SURVIVORS_TIMEOUT`: a
reproducer is one short-lived check under test, not a wait for a whole cleanup pass to finish. Why a
killed reproducer's exit status is never read as pass or fail: this mirrors `run_survivors`'s own
`WS_TIMED_OUT` flag, kept apart from the exit code so a killed process is never mistaken for an
answer.

Why the post-fix re-run needs the materiality condition on top of the symmetric exit-0 check: a
symmetric re-run alone cannot tell "the production code changed" from "the reproducer's own target
changed," and membership alone cannot tell a real fix from a comment or whitespace edit on a named
path.

### The fix round mutation-proves what it changed

**Why the parent runs the mutations.** KAN-200's fix round 2 was not careless. It mutation-proved
eight added harness cases, eight for eight, reported individually. It still shipped its own largest
structural change — a record protocol's delimiter — unproved, and reverting that protocol left all
eighteen harness cases green, because the harness contained no literal tab byte and no `\037` byte
anywhere. The gap was found in round 3, by a reviewer who thought to ask whether the previous round
had held itself to its own standard. A rule that asks the same actor to widen its own scope is
asking for the judgment that already failed. `/myflow-do` settled a related question one paragraph
earlier, where the parent re-runs each reproducer rather than accepting the subagent's account of
its own success — related, not identical: a reproducer is a re-runnable command sitting behind two
guards (`check-panel-reproducers.sh`'s shape check and the parent's own re-run), while a
`fix-mutation:` line is prose the parent itself writes and no guard reads, so it could be
fabricated undetectably in a way a reproducer's flip cannot.

**Why no guard reads the record.** To be more than a nag, a guard would have to decide from a diff
whether a round changed executable behaviour or edited a comment. A script cannot make that
classification, and in a repository that is mostly prose it fails in the direction that matters: it
fires on every prose fix round until someone narrows it into vacuity. KAN-197 examined and rejected
two mechanisms of exactly this shape.

This rejects only the diff-classifying guard. The cheaper variant `check-panel-reproducers.sh`
already precedents — a shape-and-count check that `fix-mutations-total:` matches the number of
`fix-mutation:` lines and that each is well formed — was considered and is not adopted either, for
a reason specific to this record: a reproducer's shape check works because the shape it counts is
itself evidence — a well-formed `finding-reproducer:` line names a command the parent then runs, so
a malformed one is caught before it can hide anything. A `fix-mutation:` line names no command;
counting well-formed ones only proves the parent typed a well-formed sentence, never that a mutation
ran or that a test failed. The shape check would certify the report's grammar, not its truth, which
is the thing that matters here.

A third guard shape is different from both of the above: a grep-based lint flagging a reserved
marker label — `finding-status:`, `findings-total:`, `finding-reproducer:` — appearing outside its
own marker position, the family `scripts/check-unfinished-work.sh` and
`scripts/check-panel-reproducers.sh` already form. The evidence for it is real, and is stated
plainly rather than minimised: this exact defect class was independently rediscovered three times
in this change's own review — F2, F9 and F11, in three different scopes across three rounds — and
each time the fix was more prose rather than a check. That recurrence is a genuine argument for
mechanising it. It is nonetheless out of scope here: it is a guard over the record's
well-formedness, not over mutation-proof, which is the guard the operator declined above, and it
belongs beside the record-format guards it would extend rather than inside a rule about fix-round
behaviour. It is carried forward as a follow-up candidate rather than dropped.

What the record buys instead is that the question becomes askable at all — a recorded field turns a
lucky unprompted question into a line the next round's reviewers read.

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
can correct — the trade that **Creation and cleanup**
(`skills/myflow-contracts/workspace-isolation.md`) rejects when it weighs a change stranded short of
its terminal state against stale storage.

The script-absent case and the guard-exit-2 case are two distinct triggers that land in the same
outcome, and here is why that is deliberate rather than an oversight: both end with nothing having
run the guard, and a session that recognised only "the repository does not carry it" would answer a
failed resolution by doing nothing at all and saying nothing about it.

- **Every declared row, not a subset.** A value this run resolved and did not export is a value
  nothing reads, so the workspace is isolated on paper while the applications and their checks still
  reach the project's shared resource — the same silent wrong answer as having derived nothing.

Creating a workspace database and bucket for a run that
only ever linted would leave behind resources nobody asked for and only `/myflow-finish` run 2
removes.

Why the two planning paths are left unstaged rather than filtered out of a display:

The exclusion is what keeps them out of the diff, rather than a filter applied when the diff is
displayed: a filtered display leaves them in the staging area, where the IDE's staged-changes pane
and `git status` show them again. The list is fixed — the pipeline chooses these paths itself, so no
project can differ. `/myflow-finish` stages and commits them separately, so nothing is lost by
leaving them unstaged here.

**The `reset` is what enforces the rule; without it the `add` only assumes it** — the reason is
stated once under **Git boundaries** (`skills/myflow-contracts/pipeline.md`) and is not re-derived
here. What is specific to this command is *whose* staging it retracts (an implementer subagent's own
`git add`, or a worktree resumed with a dirty index) and why `git reset -- <paths>` is the tool:
it touches the index only, restores a tracked path to its `HEAD` entry instead of staging a deletion
the way `git rm --cached` would, and succeeds when a path is absent — which `docs/superpowers/` is
on every run that has not preserved records yet, and where `git restore --staged` would refuse the
whole command and unstage nothing.

The outcome table under
**Preserving the session records** (`skills/myflow-contracts/pipeline.md`) is canonical for all
three outcomes.

Carry `artifactUrl`, `jiraIssue`, `planningEffort`, `models`, `prUrl` and `reviewPanelRoster` forward
verbatim, per the carry-forward rule in **State file** (`skills/myflow-contracts/state-file.md`),
which is canonical for what a write must re-emit.

That same carry-forward rule, which is canonical and is not restated here, is what governs the
planning-effort field's mapped-level carry-forward.

The template's third git state — committed and pushed with no PR — is one `/myflow-do` never emits
and `/myflow-status` does; the pairing is canonical under **The block each state renders**
(`skills/myflow-contracts/handoff-blocks.md`).

The block below is **not** a second definition of the handoff. It is this command's rendering of the
`IN_PROGRESS`-after-`/myflow-do` template, which is defined once under
**The block each state renders** (`skills/myflow-contracts/handoff-blocks.md`) and is canonical for the
labels, the field set and their order. What this block adds is the enumeration of the literal
alternatives `/myflow-do` writes. **Change the template first and bring this block with it** — a
field added here and not there is drift the moment `/myflow-status <name>` regenerates the same
state.

That ordering is what
makes a fix round raised after a PR is open refresh the preserved records rather than leave them a
round stale.

## Guardrails
