# Tasks — kan-36-more-brainstorming-rounds-if-needed

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Repository:** `/Users/tweety53/Projects/agents` — a documentation-and-guards repository. Every task
below edits Markdown; there is no compiled artifact and no application to run. Verification is the
three lint guards named in `.myflow/project.md` plus targeted `grep` assertions on the edited text.

**Order:** slice C first (smallest and independent), then D, then A, then B. C touches only
`jira-integration.md`; D, A and B all touch `skills/myflow-start/SKILL.md`, and D and B both touch
`skills/myflow-contracts/pipeline.md`, so doing them in this order keeps each task's diff readable
rather than interleaved.

**Global constraints, carried into every dispatch:**

- **No suppression markers, ever.** Fix the offending line; never silence a guard.
- **Citation over copy.** Where a value or rule is needed in a second place, cite the single
  statement with a bold token plus its backticked path so `scripts/check-references.sh` checks it.
- Every path written into a contract, skill or handoff is **absolute** where it names something the
  operator opens.

---

## Slice C — `TO DO URGENT` is a `To Do` synonym

### C1. State the To Do name set once, and rewrite the order as positions

- [x] Edit `skills/myflow-contracts/jira-integration.md`, section **Transitions** (the paragraph
      beginning "**Transitions are forward-only.**", currently at line 106).
- [x] Replace the four-ordered-**names** wording with four ordered **positions**, each identified by
      the names mapped onto it, and state the To Do name set exactly once here. The set is `To Do`
      and `TO DO URGENT`.
- [x] Keep every existing rule in that paragraph intact: resolve by name and never by identifier,
      case-insensitive matching, the usual spellings, forward-only, and no call when the issue is
      already at or past the target.
- [x] Add one sentence stating that mapping is by **enumerated name and never by inference**, naming
      `statusCategory` as the inference that is forbidden and why — it groups `TO DO URGENT` with
      `In Progress` under `indeterminate`.

Target shape for the set statement (wording may differ; the single-statement property may not):

```markdown unverified:confirm the surrounding paragraph's existing wording is preserved verbatim around this insertion
**The To Do position carries two names — `To Do` and `TO DO URGENT`.** This is the one statement of
that set; every other site cites it rather than enumerating it again.
```

**Verify:**

```bash verified:both commands run against this tree during planning; check-vocabulary.sh exited 0
grep -n "TO DO URGENT" skills/myflow-contracts/jira-integration.md
scripts/check-vocabulary.sh
```

Expect the set stated once in the **Transitions** section, and the **Unrecognised statuses** and
join-search occurrences to be citations or explanatory prose rather than a second enumeration.

### C2. Narrow the unrecognised-status ask

- [x] Edit the **Unrecognised statuses** section of `skills/myflow-contracts/jira-integration.md`
      (currently beginning line 111).
- [x] Change its trigger from "a status outside the four ordered names" to "a status matching no name
      mapped onto any position, including the mapped synonyms".
- [x] Leave the carve-out bound untouched: the ask is still one of exactly two, still asked once per
      run, still transitions only on an explicit yes, and the sentence bounding the set at two is
      **not** edited.
- [x] Add one sentence recording that this narrows which statuses reach the ask and does **not** add
      an interactive question.

**Verify:**

```bash verified:grep run against this tree during planning; the bounding sentence is present at line 140
grep -n "exactly two carve-outs\|Two are the whole set" skills/myflow-contracts/jira-integration.md
```

The bounding sentence must still be present and unmodified.

### C3. Make the join search cite the set instead of enumerating it

- [x] Edit the paragraph in `skills/myflow-contracts/jira-integration.md` beginning
      `**"A To Do status" means exactly two names`  (currently line 482).
- [x] Replace the enumeration with a citation of the single statement written in C1, using the bold
      token plus backticked path form so `scripts/check-references.sh` checks it.
- [x] Keep the surrounding rules unchanged: not derived from `statusCategory`, an issue at any other
      status is not a candidate, no question is asked about it, and the paragraph explaining that
      this does not reopen the unrecognised-status rule.

**Verify:**

```bash verified:script run against this tree during planning; it exited 0
scripts/check-references.sh
```

Then confirm by reading that the phrase "exactly two names" no longer introduces a second list.

---

## Slice D — `/myflow-start` stages its planning artifacts and never commits them (reverted at the panel handback)

The commit exception described in earlier revisions of this section was rewritten three times across
two review passes and produced every Critical finding either pass raised; its final form left
`openspec/<name>` checked out in the main checkout and broke every following `/myflow-do`. Presented
with that outcome, the operator chose to revert rather than attempt a further rewrite — see decision
`prohibit-rather-than-legalise` in `design.md`. The tasks below record the revert actually performed,
not the commit exception it undoes.

### D1. Revert the three files slice D touched exclusively

- [x] Restore `skills/myflow-do/SKILL.md`, `commands/myflow-start.md` and
      `commands-claude/myflow-start.md` to their pre-slice-D content: none of the three carries any
      slice A/B/C content, so `git checkout 99b751e -- <path>` is a full, correct revert for each.

**Verify:**

```bash verified:run against this tree during the revert; the diff was empty
git diff 99b751e -- skills/myflow-do/SKILL.md commands/ commands-claude/
```

### D2. Remove the commit and branch language from the two shared files, keeping slices A and B

- [x] Edit `skills/myflow-contracts/pipeline.md`: restore the Git boundaries row's `/myflow-start`
      cell to "stages planning artifacts and never commits"; remove the "Why `/myflow-start`
      commits" and "`/myflow-start` creates exactly one branch" paragraphs; restore
      "`/myflow-finish` is what commits them"; remove the skip-vs-fail paragraph written for
      `/myflow-start`'s two commits; remove the `STARTED` template's `Git:` line and the paragraph
      explaining it; restore the Level 1 stage table's `/myflow-start` row to name no commit or
      branch stage.
- [x] Edit `skills/myflow-start/SKILL.md`: remove section **A**'s branch-resolution paragraph and
      the paragraph tying both commits to that branch; make section **B** stage, not commit, the
      design document; make section **F** read "Stage the planning artifacts" again, with no commit
      instruction and no message convention; restore the guardrail `**Never** write code, create a
      worktree, or create a branch.` and add one forbidding commits.
- [x] Keep every slice A/B passage in both files untouched: the convergence loop and its two
      prompts, `### Convergence`, `### Open questions`, the `**Open questions:**` handoff line,
      section **A**'s revision-round scoping, and the convergence guardrails.

**Verify:**

```bash verified:all three run against this tree after each file, during the revert; each exited 0
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
```

### D3. Delete the `myflow-command-surface` delta spec

- [x] Remove `specs/myflow-command-surface/` from this change. The stage-only resolution complies
      with the requirement already in `openspec/specs/myflow-command-surface/spec.md`, so a delta
      restating an unchanged requirement would be noise in the permanent spec tree.

**Verify:**

```bash unverified:confirm no reference to specs/myflow-command-surface/ remains outside decision history in design.md
ls specs/
```

---

## Slice A — brainstorming converges

### A1. Give the Brainstorm expansion the loop's structure

- [x] Edit the `#### Brainstorm — /myflow-start` level-2 expansion in
      `skills/myflow-contracts/pipeline.md` (currently line 90).
- [x] Add the **structure** only: the stage iterates, applying one convergence test after every
      planning-stage exchange; it ends by confirming with the operator rather than on the command's
      own judgment; later rounds are offered rather than taken.
- [x] Cite `skills/myflow-start/SKILL.md` for the tuned mechanics — the threshold and the two prompt
      shapes — following the "state the structure, cite the file that owns the tuned values" rule
      stated at the top of the level-2 section.
- [x] Do **not** write the threshold number into `pipeline.md`.

**Verify:**

```bash verified:script run against this tree during planning; it exited 0
scripts/check-references.sh
grep -n "converg\|another round" skills/myflow-contracts/pipeline.md
```

### A2. Write the convergence test and the two prompts into section B

- [x] Edit `skills/myflow-start/SKILL.md` section **B. Basic Workflow #1 — Brainstorming**
      (currently line 128).
- [x] Add the convergence test: after every planning-stage exchange — a question round, a
      design-section approval, the operator's spec review — determine whether a question is now held
      that the inputs do not answer, and open another round while one is.
- [x] Add the **confirm** prompt, fired when the test comes back empty, with named options and
      *move on* marked as the recommendation.
- [x] Add the **soft-bound offer**, fired from the third round onward in place of silently opening
      one, showing what is still open and what would be asked, with *another round* marked as the
      recommendation.
- [x] State the threshold — the third round — here, and state plainly that the two prompts recommend
      opposite courses and that each is honest because of its own trigger, so a later editor does not
      harmonise them.
- [x] State that there is no hard cap: no round count ends the stage on the command's own authority.

Prompt shapes to follow (the planning-gate capability requires named options with a marked
recommendation, so open prose is not an option here):

```markdown unverified:confirm the AskUserQuestion option labels match the repo's existing prompt style in myflow-finish/SKILL.md
> **That is everything I hold open. Anything still unclear before I move on?**
> - **Nothing unclear — move on** *(recommended)*
> - **Another round — I have something**

> **Round <n> would ask about <what is open>. Open it?**
> - **Yes — run another round** *(recommended)*
> - **No — record what is open and move on**
```

**Verify:**

```bash verified:grep run against this tree during planning; section B begins at line 128
grep -n "third round\|convergence\|recommended" skills/myflow-start/SKILL.md | head -20
```

### A3. Scope the loop on a revision round

- [x] Edit the **Revising an existing proposal** paragraph in `skills/myflow-start/SKILL.md`
      section A (currently line 67).
- [x] State that a revision round applies the convergence test to what the feedback reopened and to
      whatever that opens in turn, and does not re-brainstorm settled parts.
- [x] Keep the existing rules: the design gate is skipped unless the feedback reopens an
      architectural question, artifacts are revised in place, and the artifact republishes to the
      same source path.

**Verify:**

```bash verified:grep run against this tree during planning; the paragraph is present at line 67
grep -n "Revising an existing proposal" skills/myflow-start/SKILL.md
```

### A4. Add the loop to the guardrails

- [x] Edit the **Guardrails** list in `skills/myflow-start/SKILL.md`.
- [x] Add a guardrail forbidding leaving brainstorming while holding an unanswered question, and one
      forbidding a planning effort level from ending the loop early.
- [x] Keep the existing guardrail that a level may group questions into fewer rounds but never turn a
      question into an assumption — the new one sits beside it rather than replacing it.

**Verify:**

```bash verified:script run against this tree during planning; it exited 0
scripts/check-vocabulary.sh
```

---

## Slice B — the open-questions record and its count

### B1. Add `## Open questions` to the design artifact

- [x] Edit `skills/myflow-start/SKILL.md` section **C. Create the change and its artifacts**
      (currently line 141), where `design.md` is described as carrying `## Decisions`.
- [x] Add `## Open questions` to that description, and add a subsection beneath the existing
      `### Decisions` subsection giving the entry shape, the immutable-ID rule, the
      `answered by <decision-id>` transition, and the never-delete rule.
- [x] State that a stage which left nothing open records none and leaves the section empty, mirroring
      the rule `## Decisions` already carries.

Entry shape to write into the skill:

```markdown verified:authored in-tree for this change; the same shape is used in this change's own design.md
### <the question>

**ID:** <kebab-case-slug>
**Status:** open
**Why it is open:** <deferred by the operator, blocked on something external, …>
**What it affects:** <what would change depending on the answer>
```

**Verify:**

```bash verified:grep run against this tree during planning; the Decisions subsection begins at line 158
grep -n "## Decisions\|Open questions" skills/myflow-start/SKILL.md
```

### B2. Carry the section into the published artifact

- [x] Edit `skills/myflow-start/SKILL.md` section **E. Publish the proposal artifact**
      (currently line 202), whose sentence lists what the page must carry.
- [x] Add the open questions to that list, beside the design and its `## Decisions`.

**Verify:**

```bash verified:grep run against this tree during planning; section E begins at line 202
grep -n "Decisions`, the delta specs" skills/myflow-start/SKILL.md
```

### B3. Add the count to the `STARTED` template

- [x] Edit the `**STARTED**` template under **The block each state renders** in
      `skills/myflow-contracts/pipeline.md` (currently line 465).
- [x] Add `**Open questions:** <count, or "none">` immediately after `**Decisions recorded:**`.
- [x] Do **not** mark it `(run-only)` — it is derived from `design.md` on disk, so
      `/myflow-status <name>` renders it. Add one sentence saying so, beside the existing paragraph
      explaining why the `Jira` line *is* run-only, so the contrast is stated rather than inferred.

**Verify:**

```bash verified:grep run against this tree during planning; the template begins at line 465
grep -n "Decisions recorded" skills/myflow-contracts/pipeline.md
```

### B4. Add the same line to the skill's rendering

- [x] Edit the handoff block in `skills/myflow-start/SKILL.md` section F (currently line 258).
- [x] Add `**Open questions:** <N> | none` immediately after `**Decisions recorded:** <N> | none`,
      keeping the skill's enumerate-the-literal-alternatives style.
- [x] Confirm the label set and order now match the template exactly — that is the property
      `/myflow-status` reproduces.

**Verify:**

```bash verified:grep run against this tree during planning; both blocks are present at the cited lines
diff <(grep -o '^\*\*[A-Za-z ]*:\*\*' skills/myflow-start/SKILL.md | head -20) /dev/null || true
grep -n "Decisions recorded" skills/myflow-start/SKILL.md skills/myflow-contracts/pipeline.md
```

Read both blocks side by side and confirm the labels and their order agree.

---

## Final verification

### V1. Run every configured guard

- [x] Run all three lint commands declared in `.myflow/project.md`, from the repository root.

```bash verified:all three run against this tree during planning; check-vocabulary.sh exited 0 and the other two are the project's declared lint set
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
```

All three must exit 0. `check-plan-provenance.sh` must report that all provenance is stated for this
change's `tasks.md`, `design.md` and `proposal.md`.

<!-- predicted: scripts/check-plan-provenance.sh after every task above -->

### V2. Run the guards' own test harnesses

- [x] Run the test commands declared in `.myflow/project.md`, since `check-references.sh`'s behaviour
      is load-bearing for slice C's citation.

```bash verified:the harness list is copied from .myflow/project.md's ## test section
scripts/test-check-references.sh
scripts/test-check-plan-provenance.sh
```

Both must exit 0. The other five harnesses in that section cover `/myflow-finish` helpers this change
does not touch; run them if the diff turns out to reach one.

### V3. Confirm the contract has no remaining contradiction

- [x] Read `skills/myflow-contracts/pipeline.md`'s Git boundaries row for `/myflow-start`,
      `skills/myflow-start/SKILL.md` sections B and F, and the **live** requirement at
      `openspec/specs/myflow-command-surface/spec.md` side by side. This change carries no delta for
      that capability — the stage-only resolution complies with the requirement already there, which
      is the point of the resolution.
- [x] Confirm all three now say the same thing: `/myflow-start` stages its planning artifacts and
      commits nothing. That is the three-way contradiction removed by making the existing
      prohibition true rather than by carving an exception into it.
- [x] Confirm the To Do name set appears as a statement in exactly one place, with every other
      occurrence a citation or explanatory prose.
