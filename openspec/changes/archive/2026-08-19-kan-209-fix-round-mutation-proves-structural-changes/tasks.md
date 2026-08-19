> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** A panel fix round proves the structural change it made, not only the test cases it was
asked to add — and records the proof where the next round's reviewers already read.

**Architecture:** Two surfaces. `scripts/test-check-unfinished-work.sh` gains a case pinning the new
record lines as inert to the guard that reads the same file; `skills/myflow-do/SKILL.md` gains the
rule itself, in section 5, immediately after the paragraph that already establishes the parent
re-verifying the subagent's work. No new script, no new guard, no change to any existing check.

**Tech Stack:** Bash (the guard harness), Markdown (the skill text). Nothing is added to either.

**Spec:** `openspec/changes/kan-209-fix-round-mutation-proves-structural-changes/design.md` and
`specs/myflow-review-panel-economics/spec.md` in the same directory. The approved design is
`docs/superpowers/specs/2026-08-19-kan-209-fix-round-mutation-proves-structural-changes-design.md`.

## Global Constraints

- **Task 1 before task 2.** Task 1 pins the record format that task 2's prose then mandates. Writing
  the prose first would let the format be chosen without the constraint that decides it.
- **No existing check changes.** `scripts/check-unfinished-work.sh` is not modified by either task.
  A task that alters what any guard already enforces has exceeded its scope.
- **No new guard script**, and no existing guard is extended to parse `fix-mutation:` or
  `fix-mutations-total:`. This is decision `no-guard` in `design.md`, not an omission.
- **No task edits `openspec/` or `docs/superpowers/`.**
- **Every new line of skill text is prose.** No mechanism, no field in `tasks.md`, no slot in any
  preset.

## Baseline

All measured 2026-08-19 against `097046f`.

- `scripts/test-check-unfinished-work.sh` passes with 122 `ok:` assertions, its highest-numbered
  case being 11.
  <!-- measured: bash scripts/test-check-unfinished-work.sh | grep -c '^ok:' @ 097046f -->
- `skills/myflow-do/SKILL.md` is 57995 bytes against a `check-contract-budget.sh` row of 71317 —
  13322 bytes of headroom.
  <!-- measured: wc -c skills/myflow-do/SKILL.md and grep the budgets table in scripts/check-contract-budget.sh @ 097046f -->
- `skills/myflow-do/SKILL-rationale.md` is 14132 bytes against a row of 15721 — 1589 bytes of
  headroom, which is the tighter of the two and the one task 2 has to watch.
  <!-- measured: wc -c skills/myflow-do/SKILL-rationale.md and grep the budgets table in scripts/check-contract-budget.sh @ 097046f -->
- `scripts/check-contract-budget.sh` reports `BUDGET-OK: 27 contract file(s) within budget`.
  <!-- measured: scripts/check-contract-budget.sh @ 097046f -->

---

### 1 `scripts/test-check-unfinished-work.sh` — the new record lines are inert to the guard

**Build:** green

**Files:**
- Modify: `scripts/test-check-unfinished-work.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the checked constraint task 2's prose cites — that `fix-mutation:` and
  `fix-mutations-total:` lines belong in the pass log entry and never inside the marker block.

`.superpowers/sdd/final-review-panel.md` has two readers. A human reads the findings table and the
pass log; `scripts/check-unfinished-work.sh` reads the marker block and the table's identifiers.
KAN-209 adds lines for the first reader to a file the second one parses, and two of that guard's
rules constrain where they may go:

- Every `finding-status:` marker must occupy **one unbroken run of consecutive lines** (the guard's
  `M_SPAN` check). A `fix-mutation:` line between two markers would split that run and make a
  correct record fail.
- The findings table's identifiers come from lines matching
  `^\|?[[:space:]]*F[0-9]+[[:space:]]*\|` (the guard's `ROW_IDS`). A line shaped like a table row
  would enter that set and unbalance the row-versus-marker comparison.

Placed in the pass log, after the marker block, the new lines hit neither. That is true as designed
and is exactly the kind of claim that stops being true after the next edit, so this task makes it a
case rather than a comment.

- [x] **Step 1: Write the failing-by-construction case**

Append to `scripts/test-check-unfinished-work.sh`, immediately **before** the final
`if [ "$FAILURES" -ne 0 ]` block at the end of the file:

````bash unverified:confirm the helper names `new_fixture`, `write_panel`, `run_guard`, `assert_verdict` and `assert_reason` still read exactly this way at implementation time — they are stable in the harness today but this snippet is transcribed, not generated
# 12. The pass log's fix-mutation: lines are inert to this guard (KAN-209).
#     A fix round records what it mutated in the pass log entry, in the same
#     file this guard reads for a different purpose. The lines are placed
#     OUTSIDE the marker block on purpose: one between two markers would split
#     the unbroken run cases 4s-0 and 4s-0b assert on, and one shaped like a
#     table row would enter the row-identifier set the 4q/4r subgroup compares
#     against the markers. This case is what makes that placement a checked
#     property of the guard rather than a claim in a design document.
#
#     The CLEAR fixture carries TWO closed markers, not one: with only one
#     marker there is no "between two markers" for a misplaced line to occupy,
#     and the contiguity mechanism could never be exercised.
add_pass_log() {
  cat >> "$WT/.superpowers/sdd/final-review-panel.md" <<'PASSLOG'

### Pass 2 — targeted

fix-mutation: scripts/check-thing.sh — reverted the unit-separator delimiter to a tab — scripts/test-check-thing.sh case 4 failed
fix-mutation: skills/myflow-do/SKILL.md — none — prose only
fix-mutations-total: 2
PASSLOG
}

new_fixture
write_panel 2 fixed fixed
add_pass_log
run_guard "$WT" demo
assert_verdict "CLEAR:" "fix-mutation: lines in the pass log leave a clear record clear"

new_fixture
write_panel 2 open fixed
add_pass_log
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "fix-mutation: lines do not hide an open finding"
````

- [x] **Step 2: Run it and watch it pass, then mutate it and watch it fail**

```bash verified:authored in-tree for this change
bash scripts/test-check-unfinished-work.sh
```

Expected: `check-unfinished-work: all cases pass`, with three more `ok:` lines than before.

This case passes on first run — the guard already ignores these lines. A case that has never been
seen to fail proves nothing, so mutate it in **two isolated** ways, one mechanism each, restoring
between them:

1. **Move one `fix-mutation:` line into the marker block** — insert it between the CLEAR
   fixture's two `finding-status:` lines. Expect the `CLEAR:` assertion to fail, naming the markers
   as spread over more lines than there are markers. Restore.
2. **Reshape one `fix-mutation:` line as a table row** — write it as
   `| F9 | fix-mutation | scripts/check-thing.sh |`. Expect the same `CLEAR:` assertion to fail, on
   the table and the marker block naming different findings. Restore.

**Both mutations must be run against the fixtures the case actually commits.** Bumping a fixture to
make a mutation possible, then committing the unbumped version, proves nothing — that is the exact
defect class this whole change exists to name, and it is why the CLEAR fixture carries two markers
rather than one.

Two mutations rather than one combined edit, because a single line that is both inside the block
and shaped like a row would trip either check and could not tell you which mechanism you proved.
That is the same false-pass shape KAN-200's fix round 3 found and split, and it is why this plan
asks for the split explicitly.

- [x] **Step 3: Record both mutations, then commit**

Report each mutation and the assertion it broke. Then:

```bash verified:authored in-tree for this change
git add scripts/test-check-unfinished-work.sh
git commit -m "test(kan-209-fix-round-mutation-proves-structural-changes): pin fix-mutation: record lines as inert to check-unfinished-work.sh"
```

**Tests:** the assertions added to `scripts/test-check-unfinished-work.sh` as cases 12, 12a and 12b
— a clear record staying clear (the assertion both mutations flip), an open finding not hidden, and
the two negative cases pinning the guard's own sensitivity to a misplaced line, added by the panel's
first fix round.

**Regression:** Reverting this task removes the only check that the pass log's `fix-mutation:` lines
leave `check-unfinished-work.sh`'s verdict alone. A later edit that moved them into the marker block
— the natural place a reader would want them — would then fail a correct panel record with no test
naming why.

**Baseline:** before=122 after=128 `ok:` assertions in `scripts/test-check-unfinished-work.sh`.
<!-- measured: bash scripts/test-check-unfinished-work.sh | grep -c '^ok:' @ branch openspec/kan-209-fix-round-mutation-proves-structural-changes gave 128 after the panel's first fix round; 122 was measured @ 097046f before this task -->

**Commit:** `test(kan-209-fix-round-mutation-proves-structural-changes): pin fix-mutation: record lines as inert to check-unfinished-work.sh`

---

### 2 `skills/myflow-do/SKILL.md` — the rule, and its reasoning

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-do/SKILL-rationale.md`
- Modify: `scripts/check-contract-budget.sh` *(only if the additions outgrow either budget row — see
  step 4)*

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 1's checked constraint, cited by the new text rather than re-argued.
- Produces: nothing later tasks depend on — this is the last task.

The rule lands in section 5, immediately **after** the paragraph beginning *"Once the fix subagent
reports, re-run every dispatched finding's reproducer"* and its closing
`See **Panel re-runs** (skills/myflow-do/SKILL-rationale.md) …` sentence, and **before** the
paragraph beginning *"**A bounce is a guard-class failure, not a review finding**"*.

That position is load-bearing. The rule reads as an extension of a mechanism the reader has just
met — the parent re-verifying rather than trusting the subagent's report — instead of as an
unrelated obligation bolted onto the end of the section. Do not move it next to
`### Bugbot's mutation-testing brief`: that subsection briefs a reviewer about a diff and applies
only where a preset dispatches Bugbot, and adjacency would invite a later reader to make this rule
inherit that condition.

- [x] **Step 1: Add the rule to `skills/myflow-do/SKILL.md`**

`````markdown verified:authored in-tree for this change
### The fix round mutation-proves what it changed

**Every executable behaviour the fix changed is mutation-proved, not only the test cases the round
adds.** When the fix subagent reports, it names the executable behaviours its fix changed — a guard
script, Go, TypeScript, shell, anything a test could fail on. **You** then mutate each one — revert
it in a scratch tree, or flip the single value it turns on — confirm an existing test fails, and
restore. The subagent does not certify its own mutations, for the same reason it does not re-run its
own reproducers one paragraph above.

**Each mutation alters one mechanism.** Where a single revert would also change state a second check
reads, split it into surgical mutations, one per mechanism. A mutation touching shared state can
pass by cross-contamination — the case goes red because the *other* check broke, not because the
mutated mechanism works — and that reports coverage which does not exist, which is worse than
recording no mutation at all.

**A surviving mutation is repaired in this round.** Where no existing test fails, add the test that
catches it before the round closes. It is not raised as an `F<n>` finding and costs no extra pass:
the round has the behaviour in hand, and a finding would spend a full round recovering context it
has not lost. This is deliberately a different disposition from a surviving mutant Bugbot reports,
which is a reviewer's reading of a diff someone else wrote.

**Record each one in this pass's log entry**, beside the mode, the slots that ran and the diff path:

```text
fix-mutation: <path> — <what was mutated> — <the test that failed>
fix-mutation: <path> — none — <reason>
fix-mutations-total: <n>
```

One line per changed behaviour, using the same `none — <reason>` exemption form the record already
uses for `finding-reproducer:`. A round that changed only prose records the exemption rather than
recording nothing.

**These lines go in the pass log entry and never inside the marker block.**
`check-unfinished-work.sh` requires every `finding-status:` marker to occupy one unbroken run of
consecutive lines, and reads the findings table's identifiers off lines matching
`^\|?[[:space:]]*F[0-9]+[[:space:]]*\|`. A line that split that run, or that looked like a row,
changes that guard's verdict on a record this rule does not otherwise touch — which is why
`scripts/test-check-unfinished-work.sh` carries a case for it.

**No guard reads these lines, and none is added.** What holds the rule instead is this: **the fix
round does not close, and the run does not reach the handoff, while an executable behaviour it
changed carries neither a line nor an exemption.** See **The fix round mutation-proves what it
changed** (`skills/myflow-do/SKILL-rationale.md`) for why the parent runs the mutations and why no
guard reads the record.

This binds the fix round under every roster, `light` included — the obligation is the round's, not a
slot's, so a preset that dispatches no Bugbot is exactly where the round's own proof is the only
mutation reasoning that happens at all. It adds no slot to any preset.
`````

- [x] **Step 2: Add the reasoning to `skills/myflow-do/SKILL-rationale.md`**

Append as a new section. Keep it to roughly the length below: this file has the tighter budget of
the two.

````markdown verified:authored in-tree for this change
### The fix round mutation-proves what it changed

**Why the parent runs the mutations.** KAN-200's fix round 2 was not careless. It mutation-proved
eight added harness cases, eight for eight, reported individually. It still shipped its own largest
structural change — a record protocol's delimiter — unproved, and reverting that protocol left all
eighteen harness cases green, because the harness contained no literal tab byte and no `\037` byte
anywhere. The gap was found in round 3, by a reviewer who thought to ask whether the previous round
had held itself to its own standard. A rule that asks the same actor to widen its own scope is
asking for the judgment that already failed. `/myflow-do` settled the identical question one
paragraph earlier, where the parent re-runs each reproducer rather than accepting the subagent's
account of its own success.

**Why no guard reads the record.** To be more than a nag, a guard would have to decide from a diff
whether a round changed executable behaviour or edited a comment. A script cannot make that
classification, and in a repository that is mostly prose it fails in the direction that matters: it
fires on every prose fix round until someone narrows it into vacuity. KAN-197 examined and rejected
two mechanisms of exactly this shape. What the record buys instead is that the question becomes
askable at all — a recorded field turns a lucky unprompted question into a line the next round's
reviewers read.
````

- [x] **Step 3: Run the guards that read these two files**

```bash verified:authored in-tree for this change
scripts/check-references.sh
scripts/check-vocabulary.sh
scripts/check-markdown-integrity.py
scripts/check-contract-budget.sh
```

All four must exit clean. `check-references.sh` is the one most likely to hit: the new text cites
**The fix round mutation-proves what it changed** in `SKILL-rationale.md`, so that heading must
exist with exactly that wording before the citation resolves. Write step 2's section before running
it, or expect a hit naming that section.

- [x] **Step 4: If a budget row trips, raise it**

`skills/myflow-do/SKILL-rationale.md` has 1589 bytes of headroom and step 2 spends most of it. When
`check-contract-budget.sh` reports either file over budget, edit that file's row in the guard's
`budgets()` table to the file's new size plus 25%, per the guard's own documented policy. Raising a
row for a genuine addition is the correct response; narrowing the guard's scope or deleting a row is
not.

- [x] **Step 5: Run the rest of the project's lint list, then commit**

```bash verified:authored in-tree for this change
scripts/check-plan-provenance.sh
scripts/check-task-build-green.sh
scripts/check-stage-mark-calls.sh
scripts/check-guard-symlinks.sh
scripts/check-self-review-report.sh
```

Then:

```bash verified:authored in-tree for this change
git add skills/myflow-do/SKILL.md skills/myflow-do/SKILL-rationale.md scripts/check-contract-budget.sh
git commit -m "docs(kan-209-fix-round-mutation-proves-structural-changes): a fix round mutation-proves every behaviour it changes"
```

`scripts/check-contract-budget.sh` is included in the `git add` only if step 4 actually edited it.

**Tests:** *(none — this task changes skill documentation and adds no test; the record format it
mandates is covered by task 1's case 12)*

**Regression:** Reverting this task leaves the fix round with no obligation to prove the structural
change it makes alongside the cases it adds — the exact state in which KAN-200's fix round 2 shipped
its delimiter change unproved and all eighteen harness cases stayed green.

**Baseline:** before=0 after=0 — this task adds no test.

**Commit:** `docs(kan-209-fix-round-mutation-proves-structural-changes): a fix round mutation-proves every behaviour it changes`
