## ADDED Requirements

### Requirement: A folded red task is checked against its partner's commit

A task tagged `**Build:** red` carries `**Squash-with:** Task <N>`, and its commit is folded into that
partner's commit before review, per **The build-green tag**
(`skills/myflow-contracts/build-green.md`). After the fold the red task has **no commit of its own**:
its declared `**Commit:**` subject no longer exists anywhere, and the surviving commit's diff spans
both tasks' files.

The runtime guard SHALL therefore resolve a red task's checks against the **partner's** commit rather
than against a commit of its own:

- `**Files:**` — the subset test SHALL be taken against the **union** of the red task's declared files
  and its partner's, since one commit now carries both.
- `**Commit:**` — the red task's own declared subject SHALL NOT be required to match. The surviving
  commit's subject SHALL be required to match the **partner's** declared `**Commit:**` field.
- Every other check SHALL apply unchanged.

The same resolution SHALL apply when the guard is invoked for the **partner**, so that checking either
task id against the folded commit gives the same verdict.

`**Squash-with:**` names **one or more** partners, per **The build-green tag**
(`skills/myflow-contracts/build-green.md`). The whole set folds into **one** commit, so:

- the union SHALL span every task in the fold — the red task and all its partners — whichever id the
  guard is invoked for, so a partner is never blamed for a **sibling** partner's declared file;
- every named partner SHALL declare the **same** `**Commit:**` subject. Two partners naming different
  subjects cannot both describe the one surviving commit, so the guard SHALL report that disagreement
  as a plan defect rather than take the first-listed partner's subject and report the commit as a
  mismatch;
- the partner set SHALL be re-validated when the guard is invoked for a partner, not only for the red
  task, so a fold the red task's id rejects cannot pass through a partner's id.

**Two red tasks MAY name the same partner, and their folds are then one unit.** Each red task's
commit folds into that shared partner's commit, so one commit carries every task in both folds. The
guard SHALL resolve the **whole** combined fold — every red task reachable through a shared partner,
and every partner of each — from **every** id in it, so that the union and the agreed subject do not
depend on which id the guard was invoked for. Resolving only the first red task that names the
invoked partner gives that partner a narrower union and, where the two folds declare different
subjects, a different verdict from the red ids — one commit with two verdicts, which the
same-verdict requirement above forbids. Two joined folds declaring **different** `**Commit:**`
subjects SHALL fail as one plan defect, reported naming the shared task that joined them and the red
tasks that disagree.

A partner declaring **no** `**Commit:**` field contributes **no** subject. It SHALL NOT be counted as
a distinct subject: an absent field disagrees with nothing, and counting it fails a fold in which
only one subject was ever declared. The one declared subject SHALL be what the surviving commit is
checked against from **every** id in the fold, the non-declaring partner's included — a partner whose
own field is absent takes the fold's agreed subject rather than checking against nothing, since one
verdict from either id is required above.

A fold in which **no** partner declares a `**Commit:**` subject at all SHALL fail the guard, naming
the red task. The surviving commit's subject would otherwise be checked against nothing and any
subject whatever would pass, which is not a commit the guard can vouch for. That failure SHALL be
reported only when the partners are otherwise valid, so a fold whose partner is missing or itself red
reports that one defect rather than two. This SHALL hold inside a fold only: `**Commit:**` remains
optional for an ordinary task, and tightening it there would reach every task in every plan rather
than the folded ones this requirement governs.

The `**Squash-with:**` grammar SHALL be defined in **exactly one place**, and both guards that read
the field SHALL import that definition rather than each carrying a copy of it. A copy per guard held
in step by a comment asserting parity is not one definition: such a comment has already been wrong
once about the very field it described, and the false claim survived four review rounds.

**Which line** in a task's body is that field SHALL be decided in that same single place, by one
function both guards call, and neither guard SHALL keep a selection loop of its own. Two loops are a
copy per guard by another name, and these two disagreed: one committed to a line only once its value
gated, the other took the first field-shaped line whatever its value, so the same body had a
different field in each guard. The comment asserting they read one field the same way was itself
wrong — the second such comment about this field to be.

**Every** structural element both guards read out of a `tasks.md` SHALL be defined in that same
single place, pattern **and** selection: which lines open a task, which task an id names, where a
task's body ends, which line is its `**Build:**` tag, and which line is its `**Squash-with:**` field.
Sharing a pattern while each guard keeps its own loop over the body is a copy per guard by another
name — the failure this requirement was first written about — and it recurred for `**Build:**` after
`**Squash-with:**` had been shared.

A `**Build:**` tag SHALL be **line-scoped**, and a task's tag SHALL be the **first** non-fenced line
that is `**Build:** green` or `**Build:** red` in full, in both guards. Reading it as a field whose
value joins following continuation lines, or letting a later `**Build:**` line overwrite an earlier
one, gave the two guards different tags for one body: `**Build:** red` followed by `**Build:**
green` was red in the plan check and green in the commit check, and `**Build:** red` followed by an
unblanked prose line was red in the plan check and no tag at all in the commit check. Either way one
guard resolved the fold while the other put the task on the ordinary single-commit path and checked
it against a `**Commit:**` subject the fold deleted.

A task heading, `**Build:**` tag or `**Squash-with:**` field inside a fenced code block SHALL open no
task and be read as no field, in **both** guards. A worked example in a plan's prose is
documentation; read as structure by one guard only, an example fold defective on purpose failed
every task in the plan.

That SHALL hold for **every** field a guard reads out of a task body — `**Files:**`, `**Commit:**`
and the rest — and **what counts as a fence SHALL be the shared grammar module's one definition**:
CommonMark's three-or-more backticks or tildes, preceded by up to three columns of indent. A guard
that gates its field loop on a narrower spelling sees a `~~~` or an indented example block as
ordinary prose, and the example `**Files:**` inside it then overwrites the task's real declaration —
so the declared file is reported as undeclared collateral and the example's file passes.

A fence a task body **opens and never closes** SHALL be reported as a plan defect of its own, naming
the task and the line the fence opened on, and SHALL **replace** — not accompany — every check whose
input that fence swallowed. An unclosed fence runs to the end of the containing block, so the fields
below it genuinely are code: swallowing them is correct, and it is the diagnosis that was wrong.
Reporting only the consequence blames the commit for a defect in the plan — the file the task really
did declare is called undeclared collateral, and the tag it really did carry is called missing. The
detection SHALL live in the shared grammar module, so both guards name the same defect. An unclosed
fence in a task body is always a plan defect, so this raises no false positive.

A task id carrying more than one heading SHALL resolve to its **first** heading, in both guards. The
duplicate is itself a plan defect the plan check reports; until it is fixed, the two guards still
have to read one task out of one id, and they read opposite ones.

A guard wrapper whose Python half imports that shared module SHALL check the module is present and
exit 2 naming its path, rather than letting a missing module surface as an import traceback and exit
1. This SHALL hold for **both** wrappers: an import is invisible to the sibling-derivation the
symlink guard performs by grepping a guard's source, so a wrapper without the check also ships a
sibling dependency nothing can see.

A `**Squash-with:**` field SHALL be **line-scoped**. Its value is the text of its own
`**Squash-with:**` line and nothing else — a following line is never part of it, even in a guard
that joins continuation lines for its other fields. The value's grammar is closed and short, so it
has no reason to wrap; and under a joined reading an unrelated prose line placed beneath the field
with no blank line between silently changes what the field names, which is a non-local dependency
invisible where the field is read. A field wrapped across two lines is therefore a plan defect, not
a longer field: the ids on its second line name nobody, and the guards fail it — the commit check on
the narrower file union that partner list implies, the plan check by validating only the partners
actually named on the line. Where a body carries more than one `**Squash-with:**` line, the task's
field is the **first line whose value gates**; a non-gating line ahead of it is skipped rather than
read as the field, in both guards.

A `**Squash-with:**` field's whole value SHALL be gated as `Task <id>[, <id>...]` — ids being digits,
dots, commas and whitespace only — **before** any partner id is extracted from it, exactly as
`check-task-build-green.py` gates the same field. A value that does not gate SHALL fail the guard,
naming the value. It SHALL NOT be treated as naming no partner: the two guards read one field,
`check-task-build-green.py` already fails a red task whose `**Squash-with:**` it cannot parse, and a
red task demoted to the ordinary single-commit path would be checked against a `**Commit:**` subject
the fold deleted. Extracting ids from an ungated value lets any digit run in free text become a
partner id, which silently widens the fold's `**Files:**` union to an unrelated task's declared files
and passes a commit carrying a file no member of the fold declared.

A `**Squash-with:**` field naming a task that does not exist in the same plan SHALL fail the guard —
that is a plan defect, and resolving against a missing partner would silently widen the file set to
nothing. A partner that is itself tagged `red` SHALL fail for the same reason
`check-task-build-green.py` already fails it.

A value that does not gate SHALL be reported from **every** task id in the plan, not only from the
red task carrying it. The guard cannot know which commit that task folded into — that is the whole
content of the defect — so it cannot know which folds the plan really has, and no verdict it reaches
for any task in that plan is one it can vouch for. Reported from the red task alone, the defect is
invisible from the very task the broken field's free text names, which is instead told that the
folded commit's other file is undeclared collateral: true of the commit, and about the wrong field.

An **unresolvable partner reference SHALL join nothing.** Where a red task names a partner that does
not exist in the plan or is itself `red`, that reference SHALL be reported against the red task
carrying it, and the named task — together with any valid fold that task belongs to — SHALL stay out
of the red task's fold membership. Growing a fold's membership through a reference the guard has
already determined to be illegal makes one red task's broken field fail a separate and entirely
correct folded commit, from every id in it.

This SHALL NOT be read as contradicting the plan-wide report above. There the guard knows both the
named partner and the reason the reference is illegal, so it knows the fold does not exist and the
defect is local to one field; with an ungated value it knows neither.

**This is a defect being repaired, not a new capability.** The guard reads `**Squash-with:**` today
only so the field terminates a preceding field's continuation, and acts on its value nowhere — so a
`Build: red` task can never pass it. The pipeline has produced folded commits for some time; commit
`377b7dd` in this repository is one, carrying two `Task-Id:` trailers.

#### Scenario: A folded red task passes on its partner's subject

- **WHEN** the guard is invoked for a `Build: red` task whose commit was folded into its partner's
- **THEN** it checks the surviving commit's subject against the **partner's** declared `Commit:` field
- **AND** does not fail merely because the red task's own declared subject is absent

#### Scenario: The file set is the union of both tasks

- **WHEN** a folded commit touches files declared by the red task and files declared by its partner
- **THEN** neither set is reported as undeclared collateral

#### Scenario: Either task id gives the same verdict

- **WHEN** the guard is invoked for the red task and then for its partner, against the same folded
  commit
- **THEN** both invocations reach the same verdict

#### Scenario: Partners declaring different subjects fail

- **WHEN** a red task's `Squash-with:` names two partners whose declared `Commit:` subjects differ
- **THEN** the guard fails naming the disagreement, rather than reporting a subject mismatch

#### Scenario: Two folds joined by a shared partner reach one verdict

- **WHEN** two `Build: red` tasks name the same green partner in their `Squash-with:` fields, and one
  commit carries every task in both folds
- **THEN** every id in the combined fold sees the union of all of their declared files
- **AND** where the two folds declare different `Commit:` subjects, every id fails naming the shared
  task, rather than the shared task's id passing a commit the red ids reject

#### Scenario: A partner declaring no subject is not a disagreement

- **WHEN** a fold's partners are one declaring a `Commit:` subject and one declaring no `Commit:`
  field at all
- **THEN** the guard checks the surviving commit against the one declared subject
- **AND** does not report a disagreement, from any id in the fold

#### Scenario: A fold declaring no subject anywhere fails

- **WHEN** no partner named anywhere in a fold — the combined group where a shared partner joined
  two folds into one commit, not one red task's own `Squash-with:` list — declares a `Commit:`
  subject
- **THEN** the guard fails, naming the shared task that joined the folds where there is one and the
  red task otherwise, rather than vouching for a commit whose subject it checked against nothing
- **AND** where any red in that combined group has a partner declaring a subject, that is the fold's
  subject and no violation is reported, from any id in the fold

#### Scenario: An unresolved partner set reports only the partner it could not resolve

- **WHEN** a red task's `Squash-with:` names a partner absent from the plan or itself red, alongside
  partners that resolve
- **THEN** the guard reports that partner and nothing else — neither a subject disagreement between
  the partners that did resolve, nor a missing fold subject, both of which would name one broken
  field twice

#### Scenario: An ordinary task's Commit: field stays optional

- **WHEN** a task with no `Squash-with:`, named by no red task, declares no `Commit:` field
- **THEN** the guard passes it, as it did before folds were resolved

#### Scenario: A Squash-with naming a missing partner fails

- **WHEN** a red task's `Squash-with:` names a task id absent from the same plan
- **THEN** the guard fails, rather than resolving against nothing

#### Scenario: A prose line under a Squash-with field is not part of its value

- **WHEN** a well-formed `Squash-with: Task 2` is followed, with no blank line between, by a line of
  prose
- **THEN** the field names Task 2, and both guards reach that same partner list
- **AND** neither guard reports the field as malformed

#### Scenario: A partner list wrapped onto a second line names one partner

- **WHEN** a `Squash-with:` field's ids continue onto a following line
- **THEN** only the ids on the field's own line are partners
- **AND** the commit check fails on the file the unnamed task declared, rather than silently widening
  the fold to it

#### Scenario: An invalid reference does not contaminate an unrelated valid fold

- **WHEN** a red task names a partner that is itself red, and that partner has its own valid fold
  with a green task, carried by its own commit
- **THEN** the guard fails the red task carrying the invalid reference, naming it
- **AND** every id in the separate valid fold passes, rather than being failed for the first red
  task's broken field

#### Scenario: A malformed Squash-with is discoverable from the task its free text names

- **WHEN** the guard is invoked for the green task a red task's ungated `Squash-with:` value mentions
  in free text
- **THEN** it reports the malformed field, naming the red task that carries it
- **AND** it does not instead report the folded commit's other file as undeclared collateral

#### Scenario: A non-gating Squash-with line ahead of a gating one is not the field

- **WHEN** a red task's body carries `**Squash-with:** Task 2 (see note below)` on one line and
  `**Squash-with:** Task 2` on the next
- **THEN** the gating line is the task's field, so the fold resolves against Task 2
- **AND** the commit check and the plan check reach the same verdict on that body

#### Scenario: A body carrying two Build lines has one tag

- **WHEN** a task's body carries `**Build:** red` on one line and `**Build:** green` on the next
- **THEN** the first line is the task's tag, so the task is red in both guards
- **AND** the commit check resolves its fold rather than checking it as an ordinary single-commit task

#### Scenario: A prose line under a Build tag is not part of it

- **WHEN** `**Build:** red` is followed, with no blank line between, by a line of ordinary prose
- **THEN** the task is still red in both guards, and the prose is not joined onto the tag's value

#### Scenario: A fenced example task heading opens no task

- **WHEN** a task's body shows a fenced example carrying a `### <id>` heading, a `**Build:** red` tag
  and a deliberately malformed `**Squash-with:**` value
- **THEN** neither guard opens a task for it, so the example is reported as no violation at all

#### Scenario: A tilde-fenced or indented example field is not a field

- **WHEN** a task body declares `**Files:**` and then shows a second `**Files:**` inside a `~~~`
  block, or inside a fence indented by up to three columns
- **THEN** the guard reads the real declaration, not the example's
- **AND** a commit touching only the example's file is reported as undeclared collateral

#### Scenario: An unclosed fence is named as the defect instead of the commit

- **WHEN** a task body opens a `~~~` fence it never closes and declares its `**Build:**`,
  `**Files:**` and `**Commit:**` fields below it
- **THEN** both guards report the unclosed fence, naming the task and the line it opened on
- **AND** neither reports the commit's file as undeclared collateral, nor the task as untagged

#### Scenario: A duplicated task id resolves to its first heading

- **WHEN** two headings in one plan carry the same task id
- **THEN** both guards resolve that id to the first heading — its tag, its `**Files:**` and its
  `**Commit:**` subject — while the plan check still reports the duplicate itself

#### Scenario: A guard copy without the shared grammar module exits 2

- **WHEN** either guard's wrapper is run from a directory holding no `lib/plan_grammar.py`
- **THEN** it exits 2 naming the missing module path, rather than exiting 1 with an import traceback

#### Scenario: Free text in a Squash-with field is a malformed field, not a partner list

- **WHEN** a red task's `Squash-with:` value carries anything beyond `Task <id>[, <id>...]` — say
  `Task 3 (see step 2)`
- **THEN** the guard fails, naming the value as malformed
- **AND** no digit in that free text is read as a partner id, so the fold's file union is not widened
  to a task the prose merely mentioned
