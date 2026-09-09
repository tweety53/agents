# kan-257-assert-mutation-edit-landed

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

- [x] 1. Hold the mutation assert wording in the dispatch-paragraph guard and its harness

**Build:** red — the extended guard fails against the unedited `skills/flow/review-panel.md`
until Task 2 lands; this task's work folds into Task 2's commit.
**Squash-with:** Task 2
**Files:**
- Modify: `scripts/check-dispatch-paragraphs.sh`
- Modify: `scripts/test-check-dispatch-paragraphs.sh`
**Tests:** `case 35`, `case 36`, `case 37`
**Regression:** reverting this task's work removes the phrase-drop coverage for the three new
assert phrases: mutating the guard's `[mutation]` entry to drop "confirm the edit landed", "a
refusal, not a surviving mutant" or "never buys a test" would no longer fail any harness case.
**Baseline:** before=34 after=37
<!-- measured: grep -oE 'case [0-9]+' scripts/test-check-dispatch-paragraphs.sh | sort -u | wc -l @ branch spectre/kan-257-assert-mutation-edit-landed -->
<!-- predicted: same grep after this task: one new drop-case per new phrase -->
**Commit:** feat(flow): require every mutation to assert its edit landed

  - [x] **Step 1: Extend the guard's `[mutation]` entry from three shared phrases to six**

In `scripts/check-dispatch-paragraphs.sh`, replace the `[mutation]` row of `ENTRY_SHARED_PHRASES`:

```bash verified:authored in-tree for this change — phrase strings match Step 5's expected output and Task 2's paragraph verbatim
  [mutation]="mutation-proved before you end your turn${US}confirm an existing test fails, and restore${US}a surviving mutant${US}confirm the edit landed${US}a refusal, not a surviving mutant${US}never buys a test"
```

Update the guard's header comment block accordingly: the `MUTATION PROOF shared phrases` note
changes "must carry all three" to "must carry the full set of six" — worded to dodge
`check-vocabulary.sh`, which blacklists the literal retired-panel-roster phrase that "all six"
is — and enumerates the three new phrases.

  - [x] **Step 2: Update the harness's `MUTATION_BLOCK` and add three deficient variants**

In `scripts/test-check-dispatch-paragraphs.sh`, `MUTATION_BLOCK` (line ~243) becomes the new
paragraph text of Task 2 Step 3, blockquote-joined — i.e. the same lines the real
`skills/flow/review-panel.md` will carry. Then add three variants beside the existing
`MUTATION_BLOCK_NO_*` ones, each the full new block with one new phrase removed:

```bash verified:transcribed from scripts/test-check-dispatch-paragraphs.sh's existing MUTATION_BLOCK_NO_* shape @ branch spectre/kan-257-assert-mutation-edit-landed
MUTATION_BLOCK_NO_EDIT_LANDED='> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you
> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a
> scratch tree, or flip the single value it turns on — but before the tests run, the edit is
> asserted: the target changed where you intended, not nowhere and not somewhere else. An edit that
> never applied is a refusal, not a surviving mutant — redo it with a working mechanism; it never
> buys a test. Then confirm an existing test fails, and restore.
> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore
> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to
> mutate is your judgment, not the script'"'"'s. Each mutation alters one mechanism — where a single
> revert would also change state a second check reads, split it into surgical mutations, one per
> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before
> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a
> `fix-mutations-total:` count, in the shape the review-panel contract'"'"'s fenced block gives. Where
> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather
> than deciding it yourself.'
```

`MUTATION_BLOCK_NO_REFUSAL` drops "a refusal, not a surviving mutant" (that sentence reads "An
edit that never applied is an unapplied edit — redo it …"); `MUTATION_BLOCK_NO_NEVER_BUYS` drops
"never buys a test" (the sentence ends "… redo it with a working mechanism."). Each keeps the
other five phrases byte-identical.

  - [x] **Step 3: Add cases 35–37, one per new phrase, mirroring cases 28–30**

After case 30's block, three new cases in the exact existing shape — fixture, run, exit-code
assertion, phrase-named-in-output assertion. Case 35:

```bash verified:transcribed from case 28's body, scripts/test-check-dispatch-paragraphs.sh @ branch spectre/kan-257-assert-mutation-edit-landed
# ===========================================================================
# Case 35: a MUTATION PROOF block is present but missing "confirm the edit
# landed" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK

$MUTATION_BLOCK_NO_EDIT_LANDED"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK

$FOREGROUND_BLOCK

$FOREGROUND_BLOCK

$TARGETED_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 35: exits 1" || fail "case 35: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"confirm the edit landed"*) pass "case 35: names the missing phrase" ;;
  *) fail "case 35: expected 'confirm the edit landed' named in output, got: $OUT" ;;
esac
```

Case 36 is the same body with `$MUTATION_BLOCK_NO_REFUSAL`, `case 36` labels, and the expected
phrase `"a refusal, not a surviving mutant"` in both assertion branches. Case 37 is the same body
with `$MUTATION_BLOCK_NO_NEVER_BUYS`, `case 37` labels, and the expected phrase `"never buys a
test"`. Update the harness header comment: "Cases 27-30" gains "Cases 35-37 cover the assert
phrases KAN-257 added, one case per phrase, each dropped in turn."

  - [x] **Step 4: Run the harness — 37 cases, all green**

Run: `scripts/test-check-dispatch-paragraphs.sh`
Expected: every case passes, including 27–30 (their deficient fixtures keep the new phrases —
only one phrase is dropped per fixture) and the new 35–37.

  - [x] **Step 5: Confirm the red state the tag claims**

Run: `scripts/check-dispatch-paragraphs.sh`
Expected: exit 1, `DISPATCH-PARAGRAPHS-INVALID`, one violation per missing new phrase — "confirm
the edit landed", "a refusal, not a surviving mutant", "never buys a test" — each named against
`skills/flow/review-panel.md`. This is the red state; Task 2 turns it green. Do not commit here;
the work folds into Task 2's commit.

- [x] 2. Put the assert obligation into the review panel contract's two mutation passages

**Build:** green
**Files:**
- Modify: `skills/flow/review-panel.md`
**Tests:** **none**
**Regression:** no test coverage changes; reverting this task's commit re-opens the red state
Step 5 of Task 1 pins (the guard fails against the reverted file) and removes the assert
obligation from the two mutation passages.
**Baseline:** before=0 after=0
<!-- predicted: no test-count change — this task edits contract prose only; Task 1's harness count is unchanged -->
**Commit:** feat(flow): require every mutation to assert its edit landed

  - [x] **Step 1: Capture the normative inventory before editing**

Run: `scripts/check-normative-inventory.sh > "${TMPDIR}/kan257-inventory-before.txt"`
Expected: exit 0 (exit 2 = cannot answer — stop and report). Only `skills/flow/review-panel.md`
changes in this task, so the later diff isolates this task's sentences.

  - [x] **Step 2: Add the assert sentences to § The mutation-testing brief**

In `skills/flow/review-panel.md`, after the brief's "…and establish whether an existing test
fails." sentence, insert:

```markdown verified:authored in-tree for this change — approved at the design gate (normative-only scope, refusal semantics)
A mutation only counts once its edit landed: every mutation **must** confirm its edit landed
before the tests run — the target changed where the slot intended, not nowhere and not somewhere
else. An edit that never applied **must** be redone with a working mechanism: it is a refusal,
never a **surviving mutant**, and it never buys a test.
```

  - [x] **Step 3: Rewrite the MUTATION PROOF paragraph — assert obligation in, basename citation in**

Replace the fix-round dispatch blockquote at `skills/flow/review-panel.md` (the paragraph
`check-dispatch-paragraphs.sh` guards, currently carrying
`<agents repo>/scripts/mutate-and-verify.sh`) with exactly the block Task 1 Step 2 encoded as
`MUTATION_BLOCK`:

```markdown verified:authored in-tree for this change — byte-identical to Task 1's MUTATION_BLOCK; keeps all three pre-existing guard phrases intact
> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you
> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a
> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit
> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that
> never applied is a refusal, not a surviving mutant — redo it with a working mechanism; it never
> buys a test. Then confirm an existing test fails, and restore.
> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore
> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to
> mutate is your judgment, not the script's. Each mutation alters one mechanism — where a single
> revert would also change state a second check reads, split it into surgical mutations, one per
> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before
> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a
> `fix-mutations-total:` count, in the shape the review-panel contract's fenced block gives. Where
> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather
> than deciding it yourself.
```

The citation drops the `<agents repo>/scripts/` prefix per Guard resolution
(`skills/flow-contracts/pipeline.md`): shipped guards are named by basename, the same way
`run-reproducer.sh` and `check-panel-reproducers.sh` are cited in this file today.

  - [x] **Step 4: Resolve the inventory diff**

Run: `scripts/check-normative-inventory.sh > "${TMPDIR}/kan257-inventory-after.txt" && diff "${TMPDIR}/kan257-inventory-before.txt" "${TMPDIR}/kan257-inventory-after.txt"`
Expected: additions only (Step 2's MUST sentences, and any sentence of the reworded paragraph
carrying a modal verb). Every removed or changed line is resolved by intent — an unexplained
removal restores the original sentence. No budget-relevant number is stated here; sizes are
Step 5's business.

  - [x] **Step 5: Run this task's lint and confirm the pair's green**

Run, in order:

```bash verified:commands copied from .flow/project.md's ## lint list, restricted to the files this plan touches
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-contract-budget.sh
scripts/check-dispatch-paragraphs.sh
```

Expected: all exit 0 — `check-dispatch-paragraphs.sh` going green here is what retires Task 1's
red tag. If `check-contract-budget.sh` reports `review-panel.md` over budget, raise its
`budgets()` row to the new size plus 25% and re-run (slack measured at plan time: the file is
51198 bytes against a 58732 budget, and the two edits add well under 1000).
<!-- measured: wc -c skills/flow/review-panel.md && sed -n '/budgets()/,/^}/p' scripts/check-contract-budget.sh @ branch spectre/kan-257-assert-mutation-edit-landed -->

  - [x] **Step 6: Re-run the harness, then commit both tasks' work**

Run: `scripts/test-check-dispatch-paragraphs.sh`
Expected: 37 cases green.

```bash verified:authored in-tree for this change
git add scripts/check-dispatch-paragraphs.sh scripts/test-check-dispatch-paragraphs.sh skills/flow/review-panel.md
git commit -m "feat(flow): require every mutation to assert its edit landed"
```
