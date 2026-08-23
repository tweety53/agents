# Tasks — stop the panel's Code review slot hanging

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** delete the forking skill from panel slot 3, and give every panel slot a wall-clock ceiling
the dispatcher enforces itself.

**Architecture:** documentation only. Two files change — `skills/myflow-do/SKILL.md` and
`skills/myflow-do/SKILL-rationale.md` — plus the delta specs already written beside this file. No Go
source, no SPA source, no guard script, no budget row.

**Spec:** `design.md` beside this file, and the two delta specs under `specs/`.

## Global constraints

- **Every task ends with every guard in `.myflow/project.md`'s `## lint` exiting clean.** There is
  no auto-fix for `scripts/check-*`; a hit is fixed by editing the offending line, never by
  weakening the guard or adding a suppression marker.
- **`check-vocabulary.sh` bans the bare token `code-review`** outside the structural exclusions its
  own header lists. Task 2 removes this repository's last live prose use of it. Do not touch
  `scripts/check-vocabulary.sh` or `scripts/test-check-vocabulary.sh` — the guard's own fixtures
  keep exercising the exclusion, and narrowing the guard because its last real hit went away is the
  opposite of the repair.
- **A normative sentence is never reworded.** `check-normative-inventory.sh` reports the set of
  `SHALL` / `SHALL NOT` / `MUST` / `MUST NOT` sentences and returns no verdict; comparing two
  inventories is this plan's job, in task 4. A difference is resolved by restoring the sentence,
  never by accepting the new inventory — except for the sentences this change deliberately adds and
  deletes, which task 4 enumerates in advance.
- **No budget row moves.** Before the first edit, `skills/myflow-do/SKILL.md` was 85,556 bytes
  against a 103,578-byte budget and `skills/myflow-do/SKILL-rationale.md` 16,735 against 20,902.
  <!-- measured: wc -c on both files; scripts/check-contract-budget.sh:192-193 @ 441ed2b -->
  Task 4 confirmed both clear after the edits, at 87,388 and 18,793 — a growth of 3,890 bytes, not
  the "roughly 2 KB" this constraint first estimated, and still well inside both budgets.
  <!-- measured: task 4's verification run; see verification.md -->
  The measured figure replaces the estimate here rather than the estimate being defended.
- **The panel record's format does not change.** New rows are recorded in the format that already
  exists; no column is added.

---

### 1 Capture the normative-inventory baseline

**Build:** green

**Files:**
- Create: `openspec/changes/kan-302-panel-code-review-slot-hangs-on-fork/normative-baseline.txt`

**Interfaces:**
- Consumes: nothing.
- Produces: the baseline task 4 diffs against.

**Commit:** `chore(openspec): capture the normative-inventory baseline`

**Nothing may be edited before this task lands.** A baseline captured after the first edit proves
nothing. Run it against a clean working tree.

- [x] **Step 1: Capture the inventory**

```bash unverified:the redirect target does not exist until this step runs
scripts/check-normative-inventory.sh \
  > openspec/changes/kan-302-panel-code-review-slot-hangs-on-fork/normative-baseline.txt
```

Its contract is exit `0` the inventory printed, exit `2` it cannot answer — it has no violation
code. A non-zero exit here is an environment failure: stop and fix it rather than proceeding with a
partial baseline.

- [x] **Step 2: Confirm the file is non-empty and sorted**

```bash verified:run at task 1, where the bare form below was found wrong
wc -l openspec/changes/kan-302-panel-code-review-slot-hangs-on-fork/normative-baseline.txt
LC_ALL=C sort -c openspec/changes/kan-302-panel-code-review-slot-hangs-on-fork/normative-baseline.txt
```

**`LC_ALL=C` is not optional here.** `check-normative-inventory.sh` sorts under `LC_ALL=C`, per its
own source, so a bare `sort -c` in this machine's default `en_US.UTF-8` locale collates differently
and reports false disorder — it did, at line 8, on this task's first run. The mismatch is the
checker's locale, never the baseline's content.

A baseline of zero lines is the guard failing silently, not a repository with no normative
sentences.

---

### 2 Delete the `code-review` skill from slot 3

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-do/SKILL-rationale.md`

**Interfaces:**
- Consumes: task 1's baseline exists.
- Produces: a slot 3 with one unconditional shape and no skill invocation.

**Commit:** `fix(myflow-do): drop the forking code-review skill from panel slot 3`

This is the defect's instance. Task 3 states the class.

- [x] **Step 1: Rewrite the slot table's row 3**

In `skills/myflow-do/SKILL.md`, the roster table's row 3 currently reads:

```markdown verified:skills/myflow-do/SKILL.md:460 @ 441ed2b
| 3 | **Code review (low)** | `light` | `models.reviewPanel` | general-purpose + the harness's `code-review` skill at effort `low`, against `final-review.diff` |
```

Its *How to spawn* cell becomes a `general-purpose` reviewer briefed for high-confidence defects
only, against `final-review.diff`. The `#`, *Slot*, *Required?* and *Model* cells are unchanged —
the slot keeps its number, its name, its preset and its model, per decision `slot3-keep-name` in
`design.md`.

- [x] **Step 2: Cut the substitution clause from the `-model` paragraph**

The paragraph at `skills/myflow-do/SKILL.md:490-494` ends:

```markdown verified:skills/myflow-do/SKILL.md:493-494 @ 441ed2b
Slot 3 names its real model, and the fallback reviewer that replaces it where the harness
offers no `code-review` skill does too.
```

There is no longer a fallback and no harness condition, so the trailing clause goes. What remains
must still say that slot 3 records its real model rather than `unknown (agent-defined)` — that fact
is what the paragraph is for, and it is not a fact this change reverses.

- [x] **Step 3: Rewrite the `### Code review (low)` subsection**

`skills/myflow-do/SKILL.md:528-543`. Everything about the skill goes: the invocation, the effort
level, the "return its findings in its report back rather than leaving them wherever the skill
displays them" instruction — there is no longer a host surface to displace findings from — and the
whole `**Where the harness offers no code-review skill**` paragraph.

What the rewritten subsection states, and nothing more:

- The slot is a `general-purpose` subagent on `models.reviewPanel`, briefed to report
  high-confidence defects only against `<abs-worktree>/.superpowers/sdd/final-review.diff`.
- It invokes no skill, so there is no harness-availability condition and nothing for the panel
  record to name as a substitution.
- Its findings are ordinary `F<n>` rows and marker lines, exactly as before.
- Because the dispatcher names its model, its dispatch row records that model and never
  `unknown (agent-defined)`.

- [x] **Step 4: Record the reversal in the rationale, without deleting the old reason**

`skills/myflow-do/SKILL-rationale.md:94-107`, the `### Code review (low)` section. The existing
paragraph explaining why the skill was chosen over a narrowed Superpowers reviewer **stays** — per
decision `rationale-records-reversal` in `design.md`, and because **Cut, never paraphrase**
(`rules/be-brief.mdc`) names a recorded reason a rejected alternative was rejected among the things
never cut. Recast it as the reason that *held until KAN-295*, and add beside it:

- what the skill's fork does to the panel's contract — the parent never observes the inner agent's
  completion;
- the three KAN-295 observations, with their numbers. **Carry them from `proposal.md`**, which
  records each with its provenance comment, rather than restating them from memory — a number
  copied without its comment is exactly what `check-plan-provenance.sh` catches next run.

- why the concern that motivated the skill — that a narrowed Superpowers reviewer duplicates slot
  0's reading — did not survive: a duplicated reading that returns beats a distinct lens that does
  not.

The `**Where an unavailable skill substitutes rather than drops**` paragraph in that same section
describes a condition that no longer exists. Delete it; the requirement it explained is removed in
this change's `myflow-review-panel-roster` delta, and its surviving half — that the panel never
falls back to two required slots — is restated in the replacement requirement and in task 3's
forking rule.

- [x] **Step 5: Confirm no live prose use of the token remains**

```bash verified:this is the same grep shape .myflow/project.md documents for guard corpora
grep -rn 'code-review' skills/ rules/ commands/ commands-claude/ openspec/specs/
```

Two classes of expected survivor, and nothing else:

- `superpowers:requesting-code-review` in `skills/myflow-do/SKILL.md` and `skills/README.md` — a
  real external Superpowers skill that `check-vocabulary.sh` excludes structurally and that this
  change does not touch.
- The hits in `openspec/specs/myflow-review-panel-roster/spec.md` — the requirement this change
  removes. **A delta spec does not edit `openspec/specs/`; finish run 2 syncs it.** They are
  therefore expected to survive task 2 and to survive task 4, and their disappearance is not
  something this change can verify. Do not edit `openspec/specs/` to make this grep quieter.

A hit anywhere else naming the harness's skill means a site was missed.

---

### 3 Add the forking rule and the wall-clock ceiling

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-do/SKILL-rationale.md`

**Interfaces:**
- Consumes: task 2, complete — the rules are written against a slot 3 that already has its new
  shape.
- Produces: the two normative rules the delta specs in
  `specs/myflow-review-panel-economics/spec.md` require.

**Commit:** `feat(myflow-do): bound every panel slot with a wall-clock ceiling`

- [x] **Step 1: Add the subsection**

A new `###` subsection in `skills/myflow-do/SKILL.md`'s section 5, placed after the dispatch-record
paragraphs and before `### Code review (low)` — it governs every slot, so it belongs above the
per-slot subsections rather than among them.

It carries both rules, in this order:

**No panel slot is dispatched onto a skill or agent that forks its own background agent.** State
why in the same breath: the parent never observes a forked agent's completion, so the slot cannot
report through the panel's contract — its findings land on a surface the panel does not read, or
never arrive, and the slot neither returns a result nor reliably ends. State the repair: dispatch
the slot on a shape that reports back directly. **Never drop the slot** — *No preset moves the
handoff bar*, earlier in this same section, already forbids it.

**Every panel slot carries a 15-minute wall-clock ceiling from its dispatch.** The dispatcher does
not block indefinitely on a completion notification; it tracks each in-flight slot's elapsed time
and enforces the ceiling itself, so a slot emitting output while making no progress is still
bounded. State the rule against the mechanism, not against one harness's tool — the same form
**Progress visibility** (`skills/myflow-contracts/pipeline.md`) uses — and say explicitly that a
dispatcher waiting on a notification alone has dropped the requirement rather than adapted it.

The breach ladder, in order, as a numbered or clearly sequenced list, because the order matters:

1. Stop the slot.
2. Close its dispatch row: `myflow record dispatch end … -outcome timed-out`.
3. Record the breach in the panel record, naming the slot and its elapsed time.
4. Re-dispatch that one slot once. Other slots are unaffected.

A **second** breach of the same slot stops and asks the operator, shaped per **Operator prompts**
(`skills/myflow-contracts/operator-prompts.md`), with three answers: re-dispatch again, proceed
without the slot, or stop the run.

**Render the prompt, do not merely cite the contract.** That contract requires the question with
named options, exactly one marked `*(default, recommended)*`, and that option named as what silence
selects. **Stop the run is that option** — re-dispatching on silence would loop, and proceeding
without the slot on silence would weaken review. Match the blockquote rendering the existing prompts
in this same file already use; do not invent a layout. What must survive alongside the default is
that a second breach is never resolved *silently*.

An earlier draft of this step said "the dispatcher makes no choice of its own here," which
contradicts the contract it cites. That sentence is deleted rather than reworded, per decision
`breach-prompt-default-stop` in `design.md`.

And the three facts that keep a timed-out slot from reading as a clean one: it raises no finding, it
consumes no fix round, and it is not a clean result for the final pass. Where the operator chooses
to proceed without it, the panel record names that slot as **not run** — distinct from a slot the
operator declined and from a slot whose trigger never fired, both of which that section already
distinguishes.

- [x] **Step 2: Add the rationale**

A matching `###` section in `skills/myflow-do/SKILL-rationale.md`, mirroring the core's heading, per
this repository's core/appendix convention. It carries the reasoning the core file states in one
line each:

- why 15 minutes — 2.5 minutes observed good, 600s watchdog that did not fire, 4 hours hung;
<!-- measured: KAN-295 panel rounds 3-4 and the re-dispatch after round 4; see proposal.md -->
- why one automatic retry before asking — the ticket's own evidence is a re-dispatch returning in
  2.5 minutes, so a retry usually clears it, and an operator interrupt for that is a worse trade;
<!-- measured: KAN-295, the re-dispatch that followed round 4 -->

- why the ceiling is not a guard script — a script has no handle on an in-flight subagent and could
  only read a timestamp the agent itself writes, which is the agent enforcing it with extra steps;
- why not the harness's own stall watchdog — round 4 is the evidence it does not fire on a slot
  emitting output, and myflow does not configure three harnesses' watchdogs;
- why panel dispatches only — implementers are serialised per worktree and report a commit sha, so a
  hang there is visible by a different route, and widening the rule to them would be speculative.

- [x] **Step 3: Check the file against its budget**

```bash verified:scripts/check-contract-budget.sh is a lint step in .myflow/project.md
scripts/check-contract-budget.sh
```

Exits clean, per the global constraint above. If it does not, the correct response is to raise the
row for the file that grew — never to narrow the guard's scope or delete a row.

---

### 4 Verify: guards clean, inventory reconciled

**Build:** green

**Files:**
- Create: `openspec/changes/kan-302-panel-code-review-slot-hangs-on-fork/verification.md`

**Interfaces:**
- Consumes: tasks 1-3, complete.
- Produces: the evidence the review panel reads alongside the diff.

**Commit:** `chore(openspec): record this change's verification evidence`

- [x] **Step 1: Run the full lint set**

Every command in `.myflow/project.md`'s `## lint` section. Run `cd stats && gofmt -w .` first per
the Lint Fix Priority rule's auto-fix step, even though this change touches no Go source — a dirty
tree from another source would otherwise surface as this change's failure.

`check-normative-inventory.sh` prints roughly a thousand lines and no verdict; that is its
documented behaviour, not a failure. Every other guard must exit clean.

- [x] **Step 2: Diff the normative inventory against task 1's baseline**

```bash unverified:depends on task 1's baseline and tasks 2-3's edits
scripts/check-normative-inventory.sh > /tmp/kan-302-normative-after.txt
diff openspec/changes/kan-302-panel-code-review-slot-hangs-on-fork/normative-baseline.txt \
     /tmp/kan-302-normative-after.txt
```

**Every difference must be one this change deliberately made.** The permitted set, and nothing
else:

- **Removed** — the `SHALL` sentences of the roster spec's *The light preset's third slot invokes
  the harness's code-review skill*, which this change removes. These leave the inventory only at
  finish run 2, when the delta syncs into `openspec/specs/`; until then they are still present and
  the diff shows no removal from that file. Say which of the two states the run observed rather
  than assuming either.
- **Not added, and never will be: the delta specs' own sentences.** An earlier draft of this bullet
  claimed `openspec/changes/` is inside the guard's owned corpus and that the delta specs' `SHALL`
  sentences therefore appear as soon as the files exist. **That premise is false.**
  `scripts/lib/owned-corpus.sh` scopes the corpus to `skills/`, `rules/`, `openspec/specs/`,
  `commands/`, `commands-claude/`, `.myflow/` and the `.md`/`.mdc` files sitting *directly* at the
  repository root, and its own header records that `openspec/changes/` is excluded **deliberately**,
  precisely so a during-a-change before/after comparison stays valid. So a delta spec's sentences are
  never inventoried, at any point in a change's life — not before the sync, not after. They enter the
  inventory only at finish run 2, as part of `openspec/specs/`.
  <!-- verified: scripts/lib/owned-corpus.sh:94 OWNED_CORPUS_SCOPE_DIRS, and that file's own header -->
  Found by task 4's per-task reviewer, as finding F1.
- **Added, in `skills/myflow-do/SKILL.md`** — the normative sentences tasks 2 and 3 write into the
  skill itself, task 3's forking rule and ceiling rule above all. `skills/` **is** a corpus scope
  root, so these are the only additions this diff can show, and showing them is correct. This bullet
  exists because an earlier draft of the set named only the delta-spec files and would have read as
  forbidding them.

Any other difference is a normative sentence this change reworded by accident. **Restore it.** Do
not accept the new inventory.

- [x] **Step 3: Re-run `openspec validate`**

```bash verified:this is the invocation that passed when the delta specs were written
openspec validate --changes kan-302-panel-code-review-slot-hangs-on-fork
```

- [x] **Step 4: Write `verification.md`**

One section per check: the lint run's result, the inventory diff with each difference classified
against the permitted set above, the budget headroom measured after the edits, and the `grep` from
task 2 step 5 with its surviving hits listed. Numbers measured in this run replace the numbers
recorded in `design.md` and `proposal.md` where they differ — in those files too, not only here.
