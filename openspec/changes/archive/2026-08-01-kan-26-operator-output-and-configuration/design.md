## Context

The full brainstorming record for this change is
`docs/superpowers/specs/2026-07-31-kan-26-operator-output-and-configuration-design.md`. This
document carries the technical decisions; that one carries the dialogue they came from.

KAN-26 is slice B of KAN-17, filed at the end of slice A's planning run. It takes the seven asks
deferred there — the ones that change what the operator sees and how the pipeline is configured —
plus one appended to the issue after filing. KAN-17's original ask numbers are used throughout, so
each traces back without a lookup.

Two of the eight surfaces have no capability in `openspec/specs/` today: the progress view and the
manual test guide's shape. That is the thread slice A also pulled — the In Review defect survived
four archived changes because no requirement existed for it to violate.

## Goals / Non-Goals

**Goals:**

- Let the operator recover the full next-step handoff at any time, without a stored copy to go
  stale.
- Make what hides under each command visible, without creating a second copy of any tuned value.
- Rename planning effort so a level says what it means, without a migration pass.
- Make the model policy's existing override durable and per-change rather than transcript-only.
- Give the progress view and the manual test guide the requirements neither has.
- Stop follow-up issues accumulating one per change.

**Non-Goals:**

- Any change to the three states, the panel roster, plan provenance, the staging and commit split,
  or the finish contract's verdict scripts.
- A new guard script. Nothing here is mechanically checkable that an existing guard does not already
  cover.
- Amending `myflow-planning-gate`, which slice A wrote in the immediately preceding change.
- A migration pass over existing state files.
- A hook written into the operator's `settings.json`.
- Renaming `myflow-review-panel-economics` to match its widened subject — the same non-goal slice A
  declared, for the same reason.

## Decisions

### The status report regenerates the handoff rather than replaying a stored one

**ID:** `status-regenerates-handoff`
**Status:** active
**Chosen:** `/myflow-status <name>` renders the state's handoff block from the state file and the
artifacts, using a per-state template defined once in `pipeline.md` — one definition, two renderers.
**Considered:** Store each command's emitted handoff text in the state file and replay it verbatim —
rejected: it adds a schema field and goes wrong silently the moment anything it names moves, and a
removed worktree or a republished artifact is exactly when the operator re-reads it. Widen the
table's existing "Next" cell with the absolute paths — rejected as not the shape the original run
printed, which is what the ask asked for.

### The progress view is the harness's widget, not a lookalike block

**ID:** `progress-uses-harness-widget`
**Status:** active
**Chosen:** Every pipeline command registers its steps with the harness's task-list mechanism and
keeps their statuses current, stated harness-neutrally with a printed block where a harness offers
none.
**Considered:** `/myflow-do` alone, on the grounds that only it has a real task list — rejected by
the operator, who wants the view throughout. A printed block in that format and nothing else —
rejected once the operator identified the widget as the thing being asked for; a lookalike would sit
in scrollback while the live view stayed empty.

### `tasks.md` gains no third checkbox marker

**ID:** `no-third-task-marker`
**Status:** active
**Chosen:** The in-progress count comes from the harness's task list alone. Any surface reading
`tasks.md` from disk reports no task in progress.
**Considered:** A `- [~]` marker written at dispatch and resolved at completion — rejected on two
counts: a crashed run leaves a permanently in-progress task behind, and `tasks.md` is parsed by
`check-unfinished-work.sh` and `check-plan-provenance.py`, so a new marker is a change to a grammar
two guards depend on. Dropping the in-progress count — rejected as not the line the ask asked for.

This is why the widget is specified as a **view, never a record**: `tasks.md` stays the single
source of completion state, because it is the one the finish gate can see.

### The guide becomes a capability checklist and keeps its preamble

**ID:** `guide-capability-checklist`
**Status:** active
**Chosen:** One tickable line per user-visible behaviour, grouped by capability, scoped to the blast
radius; a short how-to-run preamble and `## Known incomplete` retained; transcripts,
expected-output blocks and rationale dropped.
**Considered:** The checklist with no preamble at all — rejected: this repository's own guides are
script invocations, and with no `## apps` section to carry them there is nowhere else for the
commands to live. Keeping the present style under a length cap — rejected as not changing the
register the ask objected to.

The checkbox syntax and the `## Known incomplete` section are explicitly unchanged. The register is
prose; those two shapes are machine-read, and changing them would break a guard while appearing only
to shorten a document.

### The effort rename reaches the capability, not just the levels

**ID:** `rename-reaches-capability`
**Status:** active
**Chosen:** `myflow-effort` → `myflow-planning-effort`, levels `low` / `default` / `detailed`,
state-file key `effort` → `planningEffort`.
**Considered:** Renaming the levels and the operator-facing label while leaving the capability and
the key as `effort` — rejected by the operator: the stored key would no longer match the label it
stores. Renaming the concept and the key but not the capability — offered as the recommendation and
rejected in favour of going further.

The cost is accepted and stated: an OpenSpec capability rename carries every requirement twice in
the delta, ADDED under the new name and REMOVED under the old.

**The compatibility read this decision chose was reversed at the review gate and then restored**, and
all three steps are recorded here rather than tidied down to whichever one is current. Where a
decision ended up matters less than how it moved, and this change's own contracts require a reversal
to be recorded rather than deleted — which applies to the reversal of a reversal too.

1. **Chosen, originally.** The key rename would be mitigated by reading `effort` as its equivalent
   level and rewriting it in place, with no migration pass and no self-heal announcement.
2. **Reversed at the review gate.** A pass-2 finding argued that the compatibility apparatus — the
   fallback read itself, a precedence rule for a file carrying both keys, a bounded mapping table,
   and a defined outcome for a value outside that table — was restated across six files and had
   itself opened every hole four separate findings had reported against it (F11, F12 and F14 in
   pass 1; F23 in pass 2), where the retired `stage` field's mechanism answers the identical
   situation in one sentence, already implemented and already tested. Put to the operator with a
   summary of what the reversal would cost, they chose it: a file carrying `effort` would be
   unparseable, and self-heal would rebuild it from artifacts and announce the correction.
3. **Re-reversed, because that summary was false.** The operator had been told the cost was one
   field affecting only a revision round's reasoning depth, on two state files already `FINISHED`
   and never read again. Measured against the real state directory, none of that held — the figures
   are below. The corrected facts were put back to the operator, who chose to **restore the
   compatibility read**. That is what is in force.

**What the corrected measurement says.** Of the state files on this machine, five carry the retired
key and four hold a non-null value under it — three `medium` and one `high`, every one inside the
mapped three. One of those four is **this change's own state file, at `STARTED`**, carrying a live
`artifactUrl` and `jiraIssue: KAN-26`; the other three are `FINISHED`. So the affected set was twice
what the summary reported, and it included an open change rather than only closed ones.

<!-- measured: find + jq over /Users/tweety53/Agents/myflow/state/*/*.json on 2026-07-31 (UTC); machine-local, re-measured at pass 3 -->

**And a rebuild does not lose one field.** `state-self-heal.md` infers `state` — plus, as of this
pass, the `worktrees` **keys**, recovered from `git worktree list` — and nothing else. Everything
artifacts cannot reproduce is lost: `branch`, `artifactUrl`, `jiraIssue`, `planningEffort`,
`models`, `prUrl`, and each recovered worktree's merge base. The planning level belongs *in* that
list rather than in place of it, and that is the whole of the correction: the summary priced the
reversal at that one field, where a rebuild takes it **and** the operator-set data and pipeline
bookkeeping beside it — for the open file above, the published proposal link and the tracker key,
announced as unrecovered but gone all the same. `planningEffort` is named in that announcement too,
per the rule this round added to `state-self-heal.md`; being exempt from it only where a successful
read showed there was nothing to recover is what keeps the announcement's own promise honest. And because the reversed rule made the file
unparseable on the **key**, whatever value it held, every one of the five would have been rebuilt
rather than the two the summary named.

**One thing did not go back, deliberately.** An **unmapped** value under the retired key now reads
as *not recorded* rather than making the file unparseable. A pass-3 finding established that
*unparseable* is a verdict only self-heal acts on, and that nothing routes a file to self-heal on
account of an unrecognised key: `/myflow-do` and `/myflow-finish` never invoke it, and
`/myflow-status` reads the file through a literal `jq` projection that silently ignores keys it does
not name. The rule promised an announcement no command emits, which is a guarantee in prose and
nothing in practice. *Not recorded* needs no detection to be true and discards only a value that
already mapped to no level.

**The apparatus's cost is unchanged and still accepted rather than hidden:** four sub-rules across
six files is real maintenance surface, and it was an honest case for the reversal. What decided it
in the end was not that cost but the measurement — the mitigation turned out to be protecting live
operator data on an open change, which is precisely the case it was designed for.

### The panel-fix model keeps the strongest-model default

**ID:** `panel-fix-default-unchanged`
**Status:** active
**Chosen:** "Panel fixes" names the subagent that *applies* a fix, which is an implementer, so its
default remains the strongest available model. The new field makes the value overridable and
recorded, and changes no default.
**Considered:** Taking KAN-17's wording literally and defaulting fix waves to Sonnet — rejected as a
policy change contradicting two live requirements at once: implementer subagents run on the
strongest available model, and fix rounds escalate the panel's breadth rather than its model
*because* implementers already sit at the ceiling. Reading "review fixes" as the reviewer re-reading
after a fix — put to the operator and not the meaning intended.

### The three model choices are three separate prompts

**ID:** `three-separate-model-prompts`
**Status:** active
**Chosen:** One question per role on the creating run, each naming its default and marking it as the
recommendation.
**Considered:** A single question offering the three defaults with an option to customise — offered
as the recommendation and rejected by the operator. Recording the defaults silently and overriding
only by instruction — rejected: it leaves the field discoverable only by reading the state file.

### The tab commands are printed, because they cannot be invoked

**ID:** `tab-commands-printed`
**Status:** active
**Chosen:** Each pipeline command prints `/rename <change-name>` and `/color cyan` at the start of
its run, for the operator to paste, and the contract records why they are printed rather than
invoked.
**Considered:** Installing a `UserPromptSubmit` hook via `setup.sh` — rejected: it would need a
spike to learn whether a hook process can reach the terminal at all, it is Claude Code-only, and it
writes into a file myflow does not own. Dropping the ask — rejected, as the copy-paste pair delivers
most of the value for no machinery.

Both commands are real and both are unreachable from inside a run. That was established rather than
assumed, and the evidence is recorded in **Implementation notes** below so the next reader does not
repeat the investigation.

### The diagram stays small, and a stage table carries the detail

**ID:** `diagram-table-in-pipeline`
**Status:** active
**Chosen:** `pipeline.md` carries the three-state diagram plus a stage table beneath it; `README.md`
links rather than copies.
**Considered:** One flowchart with a subgraph per command — rejected: everything is in one picture,
but the loops and branches flatten. One flowchart per command plus the state diagram as an index —
rejected as three diagrams to keep in step. Three layouts were compared visually before the choice.

`pipeline.md` rather than `README.md` is what lets `/myflow-info` show it: that command reads
`pipeline.md` at invocation time and is forbidden from answering from memory.

### The stage table has two levels, with eight expansions

**ID:** `stage-table-two-levels`
**Status:** active
**Chosen:** Level 1 is one row per command; level 2 expands every stage that hides substructure —
brainstorm, writing-plans, SDD + TDD per task, the review panel, the preflight verdict, the
unfinished-work gate, the landing routes, and run 2's cleanup.
**Considered:** Expanding only the four that hide a machine — the panel, per-task TDD, the
unfinished-work gate and cleanup. Expanding the review panel alone, which is the stage the operator
named. Both rejected in favour of the complete set.

A single-level table hid the pipeline's largest machine behind the cell "review panel → 0 findings".

### Level 2 states structure and cites tuned thresholds

**ID:** `structure-stated-thresholds-cited`
**Status:** active
**Chosen:** An expansion states what is stable — which slots are required and which conditional,
that every slot runs on the panel's model (Sonnet by default) except those dispatched by
`subagent_type`, that no handoff occurs while any finding is open at any severity, that escalation
widens breadth rather than model, that a `REFUSE` stops the run — and cites the owning file for the
tuned values: the changed-line counts, the per-slot trigger lists, the conditions forcing a full
re-run.
**Considered:** Reproducing `myflow-do`'s roster and trigger tables in full — rejected: that is the
second copy `readme-links-not-copies` rejects one level up, and accepting it here would make the
rule inconsistent with itself. Citing everything and stating nothing — rejected as less detail than
the ask asked for.

Citations use the named-section form `check-references.sh` verifies, so a cited section that is
renamed fails the guard. The *absence* of a copied threshold is not guarded and cannot be; stating
the rule beside the table is the only mitigation, which is the same protection the cleanup registry
relies on.

### The README links rather than copies

**ID:** `readme-links-not-copies`
**Status:** active
**Chosen:** `README.md` carries neither the diagram nor the stage table, and points at
`pipeline.md`.
**Considered:** Keeping the small three-state diagram in the README as an at-a-glance shape while
the stage table lives only in `pipeline.md` — offered explicitly and rejected: it is one duplicated
diagram that no guard covers.
**Trade-off accepted:** a reader browsing the repository on a forge loses the at-a-glance shape.

### Nine capabilities, one change

**ID:** `nine-capabilities-one-change`
**Status:** active
**Chosen:** Each ask lands in the capability whose subject it actually is; the two unspecced
surfaces get their own. Two new, six amended, one renamed.
**Considered:** Folding the progress view and the guide into `myflow-handoff-output` — rejected:
that spec's subject would widen from "handoff output" to "every operator-facing surface", after
which "which spec do I amend?" stops having one answer. Splitting into two further changes —
rejected: KAN-26 is already the deferred half of KAN-17, and a third planning run buys smaller
review panels at the cost of another issue to file, track and finish.

`myflow-review-panel-economics` is the ninth, found while designing rather than while mapping: its
"every slot runs on Sonnet" is absolute today and would contradict a recorded panel model.

### The join search enumerates exactly two To Do statuses

**ID:** `join-search-two-todo-statuses`
**Status:** active
**Chosen:** A join candidate sits at `To Do` or `TO DO URGENT`, enumerated by name. Any other status
is not a candidate and provokes no question — the run files a new follow-up.
**Considered:** Joining only follow-ups filed for the same issue — rejected as the reading the ask
explicitly ruled out. Asking the operator before each join — rejected as another prompt during
finish; **that rejection was reversed at the review gate, and the reversal is recorded as
`join-confirmed-before-write` below.** Deriving the set from Jira's `statusCategory` — rejected, and
this is the important rejection: it is the same inference slice A's `jira-unknown-status-ask-bounded`
forbids for transitions, and `TO DO URGENT` shares the `indeterminate` category with `In Progress`.

Including `TO DO URGENT` here does not reopen that rule. It governs transitions, where inferring a
position freezes the board for a whole change; a search filter performs no transition and can only
miss a candidate.

### A join is confirmed with the operator before any write

**ID:** `join-confirmed-before-write`
**Status:** active
**Chosen:** On a match, show the candidate's key, title and current status and ask; only an explicit
yes joins, and anything else files a new follow-up instead. A failed search is a third outcome that
files nothing and says so, rather than being read as "no match".
**Considered:** Joining automatically on a match — this was the earlier call, recorded above under
`join-search-two-todo-statuses` as "rejected as another prompt during finish", and **it is the
alternative this decision reverses.** The review gate weighed it again with the search's shape in
view and came out the other way: the search selects a **write target** by label, title and status,
all three of which any project member can set, over a deliberately unnarrowed project-wide scan —
so the pipeline appends to, retitles and relabels an issue chosen entirely by attacker-settable
fields, with the confirmation the only check on it. Against that, the cost is one bounded question,
asked only when a candidate was actually found, on a course the operator has already chosen; and
declining is not a failure but the other correct outcome, since a duplicate follow-up is visible and
mergeable while a write to the wrong issue is neither. Also considered: reading a failed search as
"no match" — rejected, because it files a new follow-up on every transient tracker failure and so
reintroduces exactly the duplicate proliferation the join exists to prevent, silently, since a
created issue looks like a success.

**The earlier rejection is left standing where it was written**, annotated rather than deleted. What
changed is not that the prompt became cheaper but that its subject became clear: the first call
weighed it as a prompt about *bookkeeping*, and the review gate weighed it as a prompt about a
**write to an issue this pipeline did not choose**. Recording both is the point — a decision log
that quietly loses its reversals cannot show why the pipeline asks.

**"Project-wide" above describes the scan as it stood when this was weighed**, and it was the wider
thing that phrase makes it sound: the query carried no `project` clause at all, so the population
was every project the session's connection could reach. `join-search-project-scoped` below is the
entry that closed that, and it does not weaken this decision — the confirmation is still the only
check on a target chosen by attacker-settable fields; what narrowed is who can set them.

### The follow-up search carries a mandatory project clause

**ID:** `join-search-project-scoped`
**Status:** active
**Chosen:** The join search's JQL carries a `project` clause, built from the key(s) `## jira` names
and otherwise from the linked issue's prefix. The clause's assembly, the key shape the configured
value must match, and what becomes of a value that does not are stated once under
**Follow-up issues** (`skills/myflow-contracts/jira-integration.md`) and are not restated here.
**Considered:** Leaving the query unscoped — which is what it was until this was recorded, and the
reason `join-confirmed-before-write` above described the scan as project-wide when it was in fact
site-wide. Rejected: "any member of that project" was then not the honest description of who could
plant a candidate. Interpolating the configured value as written — rejected once the clause existed:
`.myflow/project.md` is editable in any pull request, so an unvalidated value can carry JQL syntax
and widen the scoping the clause was added to establish.

`join-confirmed-before-write` above is the entry whose reading this changes, and it is annotated
rather than rewritten, on the same rule that kept the earlier rejection standing. What the clause
does **not** touch is the deliberate width *within* the project: a follow-up filed for another
change is still a match, exactly as `join-search-two-todo-statuses` above leaves it.

### A joined issue is retitled generically

**ID:** `join-retitles-generic`
**Status:** active
**Chosen:** On the first join the title becomes `myflow follow-up`; the items are appended under a
dated `## From <KEY>` heading; the joined issue's labels are unioned with the incoming ones.
**Considered:** Keeping the original `<KEY> follow-up` title and appending only — offered as the
recommendation and rejected by the operator. Retitling to list every source — rejected as growing
without bound.

The append is governed by the pre-write assertion the description sync already carries, so a join
that cannot reproduce the existing description verbatim writes nothing.

### The vocabulary guard gains the JSON key form and nothing else

**ID:** `vocabulary-json-key-only`
**Status:** active
**Chosen:** `"effort":` is added to `check-vocabulary.sh` as a retired literal. `medium` and `high`
are not.
**Considered:** Adding the level words as well — offered and rejected: they are ordinary English
throughout this repository, `Medium` is also a Jira priority name, and the guard's own header states
that it proves a fixed list of literals is absent rather than that a rename is complete. The false
positives could not honestly be fixed, and this repository's lint policy forbids the suppression
that would follow.

## Implementation notes

### Why the tab commands are printed rather than invoked

Both commands exist in the harness, and neither is reachable from inside a run. Recorded here so the
next reader does not repeat the investigation or treat the printing as an oversight:

| Fact | Evidence |
|------|----------|
| `/color` is *"Set the prompt bar color for this session"*, `type: "local"` / `"local-jsx"` | read from the Claude Code binary |
| `/rename` is *"Rename the current conversation"*, same types, and a setting controls whether it also updates the terminal tab title | same |
| The `SlashCommand` tool exposes only commands of `type: "prompt"` | same — the gate is `p.type !== "prompt"` |
| No writable `/dev/tty` is available to a command | attempted; failed with `device not configured` |
| Accepted colours: `red`, `blue`, `green`, `yellow`, `purple`, `orange`, `pink`, `cyan`, plus `default` | read from the binary |

<!-- measured: strings over /Users/tweety53/.local/share/claude/versions/2.1.220 plus a /dev/tty
     write attempt, both on this machine on 2026-07-31; machine-local and version-specific, so no
     repository ref applies -->

`cyan` is chosen over `red`, `yellow` and `orange` because those already read as error and warning
states elsewhere in the pipeline's output.

### The state file after this change

```json unverified:the field set is this change's design; confirm against state-file.md once task group 4 lands
{
  "state": "IN_PROGRESS",
  "branch": "openspec/<name>",
  "worktrees": { "/absolute/path": "<merge-base sha>" },
  "artifactUrl": null,
  "jiraIssue": null,
  "planningEffort": null,
  "models": {
    "implementation": null,
    "reviewPanel": null,
    "panelFix": null
  },
  "prUrl": null,
  "updatedAt": "2026-07-31T00:00:00Z",
  "updatedBy": "/myflow-do"
}
```

`planningEffort` and `models` both read as "not recorded" when **absent**, which is distinct from
present-and-null. That exception already exists for `effort`; this change extends it rather than
inventing a second rule.

### The old key is read, never migrated

No pass is run over existing state files. A file carrying `effort` is read as its equivalent level
and rewritten under `planningEffort` on the next write it receives anyway. Where both keys are
present the current one wins; a value outside the mapped three reads as *not recorded* rather than
making the file unparseable, for the reason recorded under `rename-reaches-capability` above.

**A change in flight is affected, and that is the point of the read.** Of the state files on this
machine, four hold a non-null value under the retired key, and one of them is this change's own file
at `STARTED`, holding a live `artifactUrl` and `jiraIssue`. An earlier draft of this section said
nothing in flight was affected and counted two `FINISHED` files; that count was wrong, and correcting
it is what restored the compatibility read.

<!-- measured: find + jq over /Users/tweety53/Agents/myflow/state/*/*.json on 2026-07-31 (UTC); machine-local, re-measured at pass 3 -->

### The state-machine field list had already drifted

`myflow-state-machine`'s field list enumerates the state file's fields and never gained `effort`
when that field was added, while `myflow-effort` required it — so the closed-schema rule and the
field list disagreed. The delta for that capability lists `planningEffort` and `models` explicitly
and names `effort` among the keys the file must not contain, which closes the gap rather than
reproducing it one field wider.

## Risks / Trade-offs

- **The progress widget is harness-specific.** The printed fallback is the portable path, and on
  Cursor and Codex it will be the only one exercised until someone tests there.
- **The capability rename produces a bulky delta** — every requirement written twice. Mechanical,
  but large to review.
- **A creating run now asks four questions before brainstorming**: planning effort plus three
  models. Bounded by a revision round asking none.
- **The guide's register is prose**, so nothing mechanical enforces it; the panel's principles slot
  is the only check. The two machine-read shapes are unchanged, which is what keeps the guard
  working.
- **A forge reader loses the at-a-glance diagram from the README**, accepted for having no second
  copy to drift.
- **The eight level-2 expansions add maintenance surface to `pipeline.md`**, and the
  structure-here/thresholds-cited line is a judgment a later editor could blur by pasting a trigger
  list in "to make it complete". No guard can catch that.
- **The follow-up search reads externally-authored titles and descriptions.** Slice A's
  data-never-instructions clause governs that text and is cited rather than restated.

## Open Questions

None. Every question raised during brainstorming was put to the operator and answered, and each is
recorded as a decision above.
